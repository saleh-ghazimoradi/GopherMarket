package middleware

import "net/http"

type MiddlewareFunc func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
