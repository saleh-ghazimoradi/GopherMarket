package logger

import (
	"log/slog"
	"os"

	slogotel "github.com/remychantenay/slog-otel"
)

func NewSlogLogger() *slog.Logger {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02T15:04:05"))
			}
			return a
		},
	})

	// slog-otel injects trace_id, span_id from the context automatically
	otelHandler := slogotel.OtelHandler{
		Next: jsonHandler,
	}

	return slog.New(otelHandler)
}
