package handler

import (
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/helper"
	"net/http"
)

type HealthCheckHandler struct {
	cfg *config.Config
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Check if the server is running
// @Tags         healthcheck
// @Produce      json
// @Success      200 {string} string "I'm breathing!"
// @Router       /healthcheck [get]
func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"status": "available",
		"system_info": map[string]any{
			"environment": h.cfg.Application.Environment,
			"version":     h.cfg.Application.Version,
		},
	}
	helper.SuccessResponse(w, "I'm breathing!", data)
}

func NewHealthCheckHandler(cfg *config.Config) *HealthCheckHandler {
	return &HealthCheckHandler{
		cfg: cfg,
	}
}
