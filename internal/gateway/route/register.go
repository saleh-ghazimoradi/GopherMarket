package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type RegisterRoute struct {
	middleware       *middleware.Middleware
	healthCheckRoute *HealthCheckRoute
	authRoute        *AuthRoute
}

type Options func(*RegisterRoute)

func WithHealthCheckRoute(healthCheckRoute *HealthCheckRoute) Options {
	return func(r *RegisterRoute) {
		r.healthCheckRoute = healthCheckRoute
	}
}

func WithMiddleware(middleware *middleware.Middleware) Options {
	return func(r *RegisterRoute) {
		r.middleware = middleware
	}
}

func WithAuthRoute(authRoute *AuthRoute) Options {
	return func(r *RegisterRoute) {
		r.authRoute = authRoute
	}
}

func (r *RegisterRoute) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	r.healthCheckRoute.HealthCheckRoutes(mux)
	r.authRoute.AuthRoutes(mux)
	return r.middleware.Recover(r.middleware.Logging(r.middleware.CORS(mux)))
}

func NewRegisterRoute(opts ...Options) *RegisterRoute {
	r := &RegisterRoute{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}
