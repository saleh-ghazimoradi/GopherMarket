package middleware

import (
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"log/slog"
)

type Middleware struct {
	cfg            *config.Config
	logger         *slog.Logger
	allowedOrigins []string
}

func NewMiddleware(cfg *config.Config, logger *slog.Logger, allowedOrigin []string) *Middleware {
	return &Middleware{
		cfg:            cfg,
		logger:         logger,
		allowedOrigins: allowedOrigin,
	}
}
