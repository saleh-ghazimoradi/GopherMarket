package middleware

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"net/http"
)

func (m *Middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if !m.isOriginAllowed(origin) {
			helper.ForbiddenResponse(w, "Not allowed by CORS")
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) isOriginAllowed(origin string) bool {
	for _, allowed := range m.allowedOrigins {
		if allowed == origin {
			return true
		}
	}
	return false
}
