package middleware

import (
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
	"strings"
)

func (m *Middleware) Tracing(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "gophermarket-http",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return r.URL.Path != "/healthcheck" && !strings.HasPrefix(r.URL.Path, "/swagger")
		}),
	)
}
