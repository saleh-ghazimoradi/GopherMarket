package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	host         string
	port         string
	handler      http.Handler
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	errLog       *log.Logger
	logger       *slog.Logger
}

type Option func(*Server)

func WithHost(host string) Option {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port string) Option {
	return func(s *Server) {
		s.port = port
	}
}

func WithHandler(handler http.Handler) Option {
	return func(s *Server) {
		s.handler = handler
	}
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = writeTimeout
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(s *Server) {
		s.idleTimeout = idleTimeout
	}
}

func WithErrorLog(errLogger *log.Logger) Option {
	return func(s *Server) {
		s.errLog = errLogger
	}
}

func WithLogger(slogLogger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = slogLogger
	}
}

func (s *Server) addr() string {
	return fmt.Sprintf("%s:%s", s.host, s.port)
}

func (s *Server) Connect() error {
	server := &http.Server{
		Addr:         s.addr(),
		Handler:      s.handler,
		IdleTimeout:  s.idleTimeout,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		ErrorLog:     s.errLog,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		se := <-quit

		s.logger.Info("caught shutdown signal", "signal", se.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		s.logger.Info("completing background tasks", "addr", server.Addr)
		shutdownError <- nil
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	s.logger.Info("Stopped server", "addr", server.Addr)
	return nil
}

func NewServer(opts ...Option) *Server {
	s := &Server{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
