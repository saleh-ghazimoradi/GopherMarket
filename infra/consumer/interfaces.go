package consumer

import "context"

type Handler func(ctx context.Context, payload []byte) error

type Consumer interface {
	RegisterHandler(eventType string, handler Handler)
	Start(ctx context.Context) error
	Close() error
}
