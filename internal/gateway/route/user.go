package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type UserRoute struct {
	middleware  *middleware.Middleware
	userHandler *handler.UserHandler
}

func (u *UserRoute) UserRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v1/users/profile", u.middleware.WrapAuth(u.userHandler.GetUserProfile))
	mux.Handle("PATCH /v1/users/profile", u.middleware.WrapAuth(u.userHandler.UpdateUserProfile))
}

func NewUserRoute(middleware *middleware.Middleware, userHandler *handler.UserHandler) *UserRoute {
	return &UserRoute{
		middleware:  middleware,
		userHandler: userHandler,
	}
}
