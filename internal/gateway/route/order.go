package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type OrderRoute struct {
	middleware   *middleware.Middleware
	orderHandler *handler.OrderHandler
}

func (o *OrderRoute) OrderRoute(mux *http.ServeMux) {
	mux.Handle("POST /v1/orders", o.middleware.WrapAuth(o.orderHandler.CreateOrder))
	mux.Handle("GET /v1/orders", o.middleware.WrapAuth(o.orderHandler.GetUserOrders))
	mux.Handle("GET /v1/orders/{id}", o.middleware.WrapAuth(o.orderHandler.GetUserOrder))
}

func NewOrderRoute(middleware *middleware.Middleware, orderHandler *handler.OrderHandler) *OrderRoute {
	return &OrderRoute{
		middleware:   middleware,
		orderHandler: orderHandler,
	}
}
