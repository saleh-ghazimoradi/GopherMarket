package httpserver

import (
	"context"
	"crypto/tls"
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

type HTTPServer struct {
	host         string
	port         string
	handler      http.Handler
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
	errLog       *log.Logger
	logger       *slog.Logger
	certFile     string
	keyFile      string
}

type Option func(*HTTPServer)

func WithHost(host string) Option {
	return func(s *HTTPServer) {
		s.host = host
	}
}

func WithPort(port string) Option {
	return func(s *HTTPServer) {
		s.port = port
	}
}

func WithHandler(handler http.Handler) Option {
	return func(s *HTTPServer) {
		s.handler = handler
	}
}

func WithReadTimeout(readTimeout time.Duration) Option {
	return func(s *HTTPServer) {
		s.readTimeout = readTimeout
	}
}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(s *HTTPServer) {
		s.writeTimeout = writeTimeout
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(s *HTTPServer) {
		s.idleTimeout = idleTimeout
	}
}

func WithErrorLog(errLogger *log.Logger) Option {
	return func(s *HTTPServer) {
		s.errLog = errLogger
	}
}

func WithLogger(slogLogger *slog.Logger) Option {
	return func(s *HTTPServer) {
		s.logger = slogLogger
	}
}

func WithCert(certFile string) Option {
	return func(s *HTTPServer) {
		s.certFile = certFile
	}
}

func WithKey(keyFile string) Option {
	return func(s *HTTPServer) {
		s.keyFile = keyFile
	}
}

func (s *HTTPServer) addr() string {
	return fmt.Sprintf("%s:%s", s.host, s.port)
}

func (s *HTTPServer) Connect() error {
	server := &http.Server{
		Addr:         s.addr(),
		Handler:      s.handler,
		IdleTimeout:  s.idleTimeout,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		ErrorLog:     s.errLog,
	}

	if s.certFile != "" && s.keyFile != "" {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			//ClientAuth: tls.RequireAndVerifyClientCert, // enforce mTLS. These two, clientAuth, ClientCAs are used when you are working with desktop or mobile. Not the browser. Mostly used in microservices. mTLs: mutual TLS.
			//ClientCAs:  loadClientCAs(),
		}
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

	s.logger.Info("starting server", "addr", server.Addr, "tls", s.certFile != "")

	if s.certFile != "" && s.keyFile != "" {
		if err := server.ListenAndServeTLS(s.certFile, s.keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	s.logger.Info("stopped server", "addr", server.Addr)
	return nil
}

func NewHTTPServer(opts ...Option) *HTTPServer {
	s := &HTTPServer{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

//func loadClientCAs() *x509.CertPool {
//	clientCAs := x509.NewCertPool()
//	caCert, err := os.ReadFile("cert.pem")
//	if err != nil {
//		log.Fatalln("Could not read client certificate:", err)
//	}
//	if ok := clientCAs.AppendCertsFromPEM(caCert); !ok {
//		log.Fatalln("Could not append client certificate")
//	}
//	return clientCAs
//}
