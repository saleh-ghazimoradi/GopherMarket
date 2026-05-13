package service

import (
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"net"
	"net/smtp"
)

type Notifier interface {
	Send(email *dto.Email) error
	SendLoginNotification(email, username string) error
}

type emailNotifier struct {
	cfg *config.Config
}

func (e *emailNotifier) Send(email *dto.Email) error {
	addr := net.JoinHostPort(e.cfg.SMTP.Host, fmt.Sprintf("%d", e.cfg.SMTP.Port))

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	defer conn.Close()

	client, err := smtp.NewClient(conn, e.cfg.SMTP.Host)
	if err != nil {
		return err
	}

	defer func() {
		_ = client.Quit()
	}()

	if e.cfg.SMTP.Username != "" || e.cfg.SMTP.Password != "" {
		auth := smtp.PlainAuth("", e.cfg.SMTP.Username, e.cfg.SMTP.Password, e.cfg.SMTP.Host)
		if err = client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(e.cfg.SMTP.From); err != nil {
		return err
	}

	if err := client.Rcpt(email.To); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", e.cfg.SMTP.From, email.To, email.Subject, email.Body)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	return w.Close()
}

func (e *emailNotifier) SendLoginNotification(userEmail, username string) error {
	email := &dto.Email{
		To:      userEmail,
		Subject: "Login Notification",
		Body: fmt.Sprintf(`Hello %s
	
You have successfully logged into your account.

If this was not you, please contact support immediately.

Yours,
The GopherMarket Team,`, username),
	}
	return e.Send(email)
}

func NewEmailNotifier(cfg *config.Config) Notifier {
	return &emailNotifier{
		cfg: cfg,
	}
}
