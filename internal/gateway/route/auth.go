package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"net/http"
)

type AuthRoute struct {
	authHandler *handler.AuthHandler
}

func (a *AuthRoute) AuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/register", a.authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", a.authHandler.Login)
	mux.HandleFunc("POST /v1/auth/refresh", a.authHandler.RefreshToken)
	mux.HandleFunc("POST /v1/auth/logout", a.authHandler.Logout)
}

func NewAuthRoute(authHandler *handler.AuthHandler) *AuthRoute {
	return &AuthRoute{
		authHandler: authHandler,
	}
}
