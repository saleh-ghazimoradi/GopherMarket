package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/domain"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/dto"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/awsCfg"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// notifierCmd represents the notifier command
var notifierCmd = &cobra.Command{
	Use:   "notifier",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("notifier called")

		ctx := context.Background()

		sLogger := logger.NewSlogLogger()

		cfg, err := config.GetConfigInstance()
		if err != nil {
			sLogger.Error("failed to get config", "error", err)
			return
		}

		sLogger.Info("Starting notification service...")

		emailNotifier := service.NewEmailNotifier(cfg)

		awsConfig, err := awsCfg.CreateAWSConfig(ctx, cfg.AWS.S3Endpoint, cfg.AWS.Region)
		if err != nil {
			sLogger.Error("failed to create aws config", "error", err)
			return
		}

		logger := watermill.NewStdLogger(false, false)

		subscriber, err := sqs.NewSubscriber(sqs.SubscriberConfig{
			AWSConfig: awsConfig,
		}, logger)

		if err != nil {
			sLogger.Error("failed to create sqs subscriber", "error", err)
			return
		}

		messages, err := subscriber.Subscribe(ctx, cfg.AWS.EventQueueName)
		if err != nil {
			_ = subscriber.Close()
			sLogger.Error("failed to subscribe to events", "error", err)
			return
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		sLogger.Info("Notification service started. Waiting for messages...")

		for {
			select {
			case msg := <-messages:
				if err := processMessage(cfg, sLogger, msg, emailNotifier); err != nil {
					sLogger.Error("failed to process message", "error", err)
					msg.Nack()
				} else {
					msg.Ack()
				}
			case <-sigChan:
				sLogger.Info("Notification service shutting down")
				_ = subscriber.Close()
				return
			}
		}
	},
}

func processMessage(cfg *config.Config, logger *slog.Logger, msg *message.Message, emailNotifier service.Notifier) error {
	eventType := msg.Metadata.Get("event_type")

	switch eventType {
	case cfg.Event.UserLoggedIn:
		return handleUserLoggedIn(logger, msg, emailNotifier)
	case cfg.Event.PasswordResetRequested:
		return handlePasswordResetRequested(logger, msg, emailNotifier)
	default:
		logger.Error("Unknown event type", "type", eventType)
		return nil
	}
}

func handleUserLoggedIn(logger *slog.Logger, msg *message.Message, emailNotifier service.Notifier) error {
	var user domain.User

	if err := json.Unmarshal(msg.Payload, &user); err != nil {
		return err
	}

	userName := user.FirstName + " " + user.LastName
	if userName == " " {
		userName = "User"
	}

	logger.Info("Sending login notification for", "email", user.Email)

	return emailNotifier.SendLoginNotification(user.Email, userName)
}

func handlePasswordResetRequested(logger *slog.Logger, msg *message.Message, emailNotifier service.Notifier) error {
	var event dto.PasswordResetEmailEvent
	if err := json.Unmarshal(msg.Payload, &event); err != nil {
		return err
	}

	email := &dto.Email{
		To:      event.Email,
		Subject: "Password Reset Request",
		Body:    fmt.Sprintf("Click the link to reset your password: %s", event.ResetLink),
	}

	logger.Info("Sending email notification for password reset")

	return emailNotifier.Send(email)
}

func init() {
	rootCmd.AddCommand(notifierCmd)
}
