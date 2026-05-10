package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type CategoryRoute struct {
	middleware      *middleware.Middleware
	categoryHandler *handler.CategoryHandler
}

func (c *CategoryRoute) CategoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/categories", c.categoryHandler.GetCategories)
	mux.Handle("POST /v1/categories", c.middleware.WrapAdmin(c.categoryHandler.CreateCategory))
	mux.Handle("PUT /v1/categories/{id}", c.middleware.WrapAdmin(c.categoryHandler.UpdateCategory))
	mux.Handle("DELETE /v1/categories/{id}", c.middleware.WrapAdmin(c.categoryHandler.DeleteCategory))
}

func NewCategoryRoute(middleware *middleware.Middleware, categoryHandler *handler.CategoryHandler) *CategoryRoute {
	return &CategoryRoute{
		middleware:      middleware,
		categoryHandler: categoryHandler,
	}
}
