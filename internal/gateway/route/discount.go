package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type DiscountRoute struct {
	middleware      *middleware.Middleware
	discountHandler *handler.DiscountHandler
}

func (d *DiscountRoute) DiscountRoutes(mux *http.ServeMux) {
	mux.Handle("POST /v1/products/discounts", d.middleware.WrapAdmin(d.discountHandler.CreateDiscount))
	mux.Handle("DELETE /v1/products/{productId}/discounts/{id}", d.middleware.WrapAdmin(d.discountHandler.DeleteDiscount))
}

func NewDiscountRoute(middleware *middleware.Middleware, discountHandler *handler.DiscountHandler) *DiscountRoute {
	return &DiscountRoute{
		middleware:      middleware,
		discountHandler: discountHandler,
	}
}
