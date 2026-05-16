package publisher

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	_ "github.com/aws/smithy-go/endpoints"
	"github.com/goccy/go-json"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/awsCfg"
)

type watermillPublisher struct {
	publisher message.Publisher
	queueName string
}

func (w *watermillPublisher) Publish(ctx context.Context, eventType string, payload any, metadata map[string]string) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), data)
	msg.Metadata.Set("event_type", eventType)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return w.publisher.Publish(w.queueName, msg)
}

func (w *watermillPublisher) Close() error {
	return w.publisher.Close()
}

func NewWatermillPublisher(ctx context.Context, cfg *config.Config) (Publisher, error) {
	logger := watermill.NewStdLogger(false, false)
	awsConfig, err := awsCfg.CreateAWSConfig(ctx, cfg.AWS.S3Endpoint, cfg.AWS.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	publisherConfig := sqs.PublisherConfig{
		AWSConfig: awsConfig,
		Marshaler: nil,
	}

	publisher, err := sqs.NewPublisher(publisherConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	return &watermillPublisher{
		publisher: publisher,
		queueName: cfg.AWS.EventQueueName,
	}, nil
}
