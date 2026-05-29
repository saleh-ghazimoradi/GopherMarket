package cmd

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/consumer"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/tracing"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
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

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sLogger := logger.NewSlogLogger()

		cfg, err := config.GetConfigInstance()
		if err != nil {
			sLogger.Error("failed to get config", "error", err)
			return
		}

		tracerShutdown, err := tracing.Setup(ctx, sLogger, cfg, "gophermarket-notifier")
		if err != nil {
			sLogger.Error("failed to setup tracing", "err", err)
			return
		}
		defer func() {
			if err := tracerShutdown(context.Background()); err != nil {
				sLogger.Error("tracer shutdown error", "err", err)
			}
		}()

		emailNotifier := service.NewEmailNotifier(cfg)
		eventHandler := handler.NewEventNotifierHandler(emailNotifier, sLogger)

		cons, err := consumer.NewWatermillConsumer(ctx, cfg, sLogger)
		if err != nil {
			sLogger.Error("failed to create consumer", "error", err)
			return
		}
		defer cons.Close()

		cons.RegisterHandler(cfg.Event.UserLoggedIn, eventHandler.HandleUserLoggedIn)
		cons.RegisterHandler(cfg.Event.PasswordResetRequested, eventHandler.HandlePasswordResetRequested)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			sLogger.Info("Received shutdown signal, stopping consumer...")
			cancel()
			cons.Close()
		}()

		sLogger.Info("Starting notification service...")
		if err := cons.Start(ctx); err != nil {
			sLogger.Error("consumer stopped with error", "error", err)
		}
		sLogger.Info("Notification service gracefully stopped.")
	},
}

func init() {
	rootCmd.AddCommand(notifierCmd)
}
