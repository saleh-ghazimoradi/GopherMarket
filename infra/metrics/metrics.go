package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	promclient "github.com/prometheus/client_golang/prometheus"

	"github.com/saleh-ghazimoradi/GopherMarket/config"
)

func Setup(cfg *config.Metrics, serviceName string) (func(context.Context) error, http.Handler, error) {
	if !cfg.Enabled {
		slog.Info("metrics disabled")
		noop := func(context.Context) error { return nil }
		return noop, nil, nil
	}

	exporter, err := prometheus.New(
		prometheus.WithRegisterer(promclient.DefaultRegisterer), // use the original DefaultRegisterer
	)
	if err != nil {
		return nil, nil, fmt.Errorf("prometheus exporter: %w", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("resource: %w", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(provider)

	slog.Info("metrics initialised", "port", cfg.Port)

	shutdown := func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}

	return shutdown, promhttp.Handler(), nil
}

func Serve(ctx context.Context, handler http.Handler, port string) error {
	if handler == nil {
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	slog.Info("metrics server listening", "port", port)
	return srv.ListenAndServe()
}
