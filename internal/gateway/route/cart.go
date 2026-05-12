package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type CartRoute struct {
	middleware  *middleware.Middleware
	cartHandler *handler.CartHandler
}

func (c *CartRoute) CartRoutes(mux *http.ServeMux) {
	mux.Handle("GET /v1/carts", c.middleware.WrapAuth(c.cartHandler.GetCart))
	mux.Handle("POST /v1/carts/items", c.middleware.WrapAuth(c.cartHandler.AddToCart))
	mux.Handle("PUT /v1/carts/items/{id}", c.middleware.WrapAuth(c.cartHandler.UpdateCart))
	mux.Handle("DELETE /v1/carts/items/{id}", c.middleware.WrapAuth(c.cartHandler.RemoveCart))
}

func NewCartRoute(middleware *middleware.Middleware, cartHandler *handler.CartHandler) *CartRoute {
	return &CartRoute{
		middleware:  middleware,
		cartHandler: cartHandler,
	}
}
