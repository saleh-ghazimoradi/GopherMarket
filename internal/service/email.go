package service

import (
	"context"
	"fmt"
	"net"
	"net/smtp"

	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Notifier interface {
	Send(ctx context.Context, email *dto.Email) error
	SendLoginNotification(ctx context.Context, email, username string) error
}

type emailNotifier struct {
	cfg    *config.Config
	tracer trace.Tracer
}

func (e *emailNotifier) Send(ctx context.Context, email *dto.Email) error {
	ctx, span := e.tracer.Start(ctx, "send_email",
		trace.WithAttributes(
			attribute.String("email.recipient", email.To),
			attribute.String("email.subject", email.Subject),
		))
	defer span.End()

	addr := net.JoinHostPort(e.cfg.SMTP.Host, fmt.Sprintf("%d", e.cfg.SMTP.Port))

	var conn net.Conn
	if err := func() error {
		_, dialSpan := e.tracer.Start(ctx, "smtp_dial",
			trace.WithAttributes(attribute.String("smtp.addr", addr)))
		defer dialSpan.End()

		var err error
		conn, err = net.Dial("tcp", addr)
		return err
	}(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.cfg.SMTP.Host)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp client: %w", err)
	}
	defer func() { _ = client.Quit() }()

	if e.cfg.SMTP.Username != "" || e.cfg.SMTP.Password != "" {
		auth := smtp.PlainAuth("", e.cfg.SMTP.Username, e.cfg.SMTP.Password, e.cfg.SMTP.Host)
		if err = client.Auth(auth); err != nil {
			span.RecordError(err)
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(e.cfg.SMTP.From); err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp mail from: %w", err)
	}

	if err := client.Rcpt(email.To); err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp data: %w", err)
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.cfg.SMTP.From, email.To, email.Subject, email.Body,
	)

	if _, err := w.Write([]byte(msg)); err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp write: %w", err)
	}

	if err := w.Close(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("smtp close: %w", err)
	}

	return nil
}

func (e *emailNotifier) SendLoginNotification(ctx context.Context, userEmail, username string) error {
	// A child span for the login notification.
	ctx, span := e.tracer.Start(ctx, "send_login_notification",
		trace.WithAttributes(
			attribute.String("email.recipient", userEmail),
			attribute.String("email.username", username),
		))
	defer span.End()

	email := &dto.Email{
		To:      userEmail,
		Subject: "Login Notification",
		Body: fmt.Sprintf(`Hello %s

You have successfully logged into your account.

If this was not you, please contact support immediately.

Yours,
The GopherMarket Team,`, username),
	}
	return e.Send(ctx, email)
}

func NewEmailNotifier(cfg *config.Config) Notifier {
	return &emailNotifier{
		cfg:    cfg,
		tracer: otel.Tracer("gophermarket-email-notifier"),
	}
}
