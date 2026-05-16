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
	mux.HandleFunc("POST /v1/auth/google", a.authHandler.GoogleLogin)
	mux.HandleFunc("POST /v1/auth/forgot-password", a.authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/auth/reset-password", a.authHandler.ResetPassword)
	mux.HandleFunc("POST /v1/auth/refresh", a.authHandler.RefreshToken)
	mux.HandleFunc("POST /v1/auth/logout", a.authHandler.Logout)
}

func NewAuthRoute(authHandler *handler.AuthHandler) *AuthRoute {
	return &AuthRoute{
		authHandler: authHandler,
	}
}
