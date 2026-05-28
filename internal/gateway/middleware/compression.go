package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	w io.Writer
}

func (grw gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.w.Write(b)
}

func (grw gzipResponseWriter) Flush() {
	if f, ok := grw.ResponseWriter.(http.Flusher); ok {
		if gzw, ok := grw.w.(*gzip.Writer); ok {
			gzw.Flush()
		}
		f.Flush()
	}
}

func (m *Middleware) Compression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		var rw http.ResponseWriter = gzipResponseWriter{ResponseWriter: w, w: gz}

		next.ServeHTTP(rw, r)
	})
}
