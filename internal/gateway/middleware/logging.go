package middleware

import (
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		m.logger.Info("Incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"protocol", r.Proto,
			"remote_addr", r.RemoteAddr,
			"trace_id", traceID,
			"span_id", spanID,
		)
		next.ServeHTTP(w, r)
	})
}
