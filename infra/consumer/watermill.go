package consumer

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/awsCfg"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
)

type WatermillConsumer struct {
	subscriber message.Subscriber
	queueName  string
	handlers   map[string]Handler
	logger     *slog.Logger
}

func (c *WatermillConsumer) RegisterHandler(eventType string, handler Handler) {
	c.handlers[eventType] = handler
}

func (c *WatermillConsumer) Start(ctx context.Context) error {
	messages, err := c.subscriber.Subscribe(ctx, c.queueName)
	if err != nil {
		return fmt.Errorf("failed to subscribe to queue: %w", err)
	}

	c.logger.Info("Consumer started, waiting for messages...")

	for msg := range messages {
		c.processMessage(msg)
	}

	return nil
}

func (c *WatermillConsumer) processMessage(msg *message.Message) {
	eventType := msg.Metadata.Get("event_type")

	// 1. Extract Trace Context cleanly
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(msg.Metadata))
	tracer := otel.Tracer("gophermarket-consumer")
	ctx, span := tracer.Start(ctx, "process_message", trace.WithAttributes(
		attribute.String("event.type", eventType),
		attribute.String("message.id", msg.UUID),
	))
	defer span.End()

	// 2. Route to appropriate handler
	handler, exists := c.handlers[eventType]
	if !exists {
		c.logger.ErrorContext(ctx, "Unknown event type dropped", "type", eventType)
		msg.Ack() // Ack to drop unknown messages so they don't clog the queue
		return
	}

	// 3. Execute and manage Ack/Nack
	if err := handler(ctx, msg.Payload); err != nil {
		c.logger.ErrorContext(ctx, "Failed to process message", "error", err, "type", eventType)
		msg.Nack()
	} else {
		msg.Ack()
	}
}

func (c *WatermillConsumer) Close() error {
	return c.subscriber.Close()
}

func NewWatermillConsumer(ctx context.Context, cfg *config.Config, logger *slog.Logger) (Consumer, error) {
	awsConfig, err := awsCfg.CreateAWSConfig(ctx, cfg.AWS.S3Endpoint, cfg.AWS.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	wLogger := watermill.NewStdLogger(false, false)
	subscriber, err := sqs.NewSubscriber(sqs.SubscriberConfig{
		AWSConfig: awsConfig,
	}, wLogger)

	if err != nil {
		return nil, fmt.Errorf("failed to create sqs subscriber: %w", err)
	}

	return &WatermillConsumer{
		subscriber: subscriber,
		queueName:  cfg.AWS.EventQueueName,
		handlers:   make(map[string]Handler),
		logger:     logger,
	}, nil
}
