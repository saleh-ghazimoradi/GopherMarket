package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"net/http"
)

type HealthCheckRoute struct {
	healthCheckHandler *handler.HealthCheckHandler
}

func (h *HealthCheckRoute) HealthCheckRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/healthcheck", h.healthCheckHandler.HealthCheck)
}

func NewHealthCheckRoute(healthCheckHandler *handler.HealthCheckHandler) *HealthCheckRoute {
	return &HealthCheckRoute{
		healthCheckHandler: healthCheckHandler,
	}
}
