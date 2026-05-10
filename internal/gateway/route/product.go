package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type ProductRoute struct {
	middleware     *middleware.Middleware
	productHandler *handler.ProductHandler
}

func (p *ProductRoute) ProductRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/products", p.productHandler.GetProducts)
	mux.HandleFunc("GET /v1/products/{id}", p.productHandler.GetProduct)
	mux.Handle("POST /v1/products", p.middleware.WrapAdmin(p.productHandler.CreateProduct))
	mux.Handle("PUT /v1/products/{id}", p.middleware.WrapAdmin(p.productHandler.UpdateProduct))
	mux.Handle("DELETE /v1/products/{id}", p.middleware.WrapAdmin(p.productHandler.DeleteProduct))
	mux.Handle("POST /v1/products/{id}/image", p.middleware.WrapAdmin(p.productHandler.UploadProductImage))
}

func NewProductRoute(middleware *middleware.Middleware, productHandler *handler.ProductHandler) *ProductRoute {
	return &ProductRoute{
		middleware:     middleware,
		productHandler: productHandler,
	}
}
