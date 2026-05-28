package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type AuthRoute struct {
	authHandler *handler.AuthHandler
	middleware  *middleware.Middleware
}

func (a *AuthRoute) AuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/auth/register", a.authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", a.authHandler.Login)
	mux.HandleFunc("POST /v1/auth/google", a.authHandler.GoogleLogin)
	mux.HandleFunc("POST /v1/auth/forgot-password", a.authHandler.ForgotPassword)
	mux.HandleFunc("POST /v1/auth/reset-password", a.authHandler.ResetPassword)
	mux.HandleFunc("POST /v1/auth/refresh", a.authHandler.RefreshToken)
	mux.HandleFunc("POST /v1/auth/logout", a.authHandler.Logout)
	mux.Handle("PATCH /v1/auth/change-password", a.middleware.WrapAuth(a.authHandler.ChangePassword))
}

func NewAuthRoute(authHandler *handler.AuthHandler, middleware *middleware.Middleware) *AuthRoute {
	return &AuthRoute{
		authHandler: authHandler,
		middleware:  middleware,
	}
}
