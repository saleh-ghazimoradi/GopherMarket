package publisher

import "context"

type Publisher interface {
	Publish(ctx context.Context, eventType string, payload any, metadata map[string]string) error
	Close() error
}
