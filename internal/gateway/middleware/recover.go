package middleware

import (
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"net/http"
)

func (m *Middleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				helper.InternalServerError(w, "panic recovery hit", fmt.Errorf("%v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
