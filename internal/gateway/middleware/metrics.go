package middleware

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"net/http"
	"time"
)

func (m *Middleware) Metrics(next http.Handler) http.Handler {
	meter := otel.Meter("gophermarket-http")

	requestCounter, _ := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	requestDuration, _ := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithExplicitBucketBoundaries(
			0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
		),
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			duration := time.Since(start).Seconds()
			attrs := metric.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
				attribute.Int("http.status_code", rec.status),
			)
			requestDuration.Record(r.Context(), duration, attrs)
			requestCounter.Add(r.Context(), 1, attrs)
		}()
		next.ServeHTTP(rec, r)
	})
}
