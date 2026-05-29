package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"log/slog"
)

type EventNotifierHandler struct {
	notifier service.Notifier
	logger   *slog.Logger
}

func (h *EventNotifierHandler) HandleUserLoggedIn(ctx context.Context, payload []byte) error {
	var user domain.User
	if err := json.Unmarshal(payload, &user); err != nil {
		return err
	}

	userName := user.FirstName + " " + user.LastName
	if userName == " " {
		userName = "User"
	}

	h.logger.InfoContext(ctx, "Sending login notification", "email", user.Email)
	return h.notifier.SendLoginNotification(ctx, user.Email, userName)
}

func (h *EventNotifierHandler) HandlePasswordResetRequested(ctx context.Context, payload []byte) error {
	var event dto.PasswordResetEmailEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	email := &dto.Email{
		To:      event.Email,
		Subject: "Password Reset Request",
		Body:    fmt.Sprintf("Click the link to reset your password: %s", event.ResetURL),
	}

	h.logger.InfoContext(ctx, "Sending password reset email")
	return h.notifier.Send(ctx, email)
}

func NewEventNotifierHandler(notifier service.Notifier, logger *slog.Logger) *EventNotifierHandler {
	return &EventNotifierHandler{
		notifier: notifier,
		logger:   logger,
	}
}
