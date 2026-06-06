package resolver

import (
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
)

type Resolver struct {
	authService     service.AuthService
	cartService     service.CartService
	categoryService service.CategoryService
	orderService    service.OrderService
	productService  service.ProductService
	discountService service.DiscountService
	userService     service.UserService
	uploadService   service.UploadService
	cfg             *config.Config
}

type Options func(*Resolver)

func WithAuthService(authService service.AuthService) Options {
	return func(r *Resolver) {
		r.authService = authService
	}
}

func WithCartService(cartService service.CartService) Options {
	return func(r *Resolver) {
		r.cartService = cartService
	}
}

func WithCategoryService(categoryService service.CategoryService) Options {
	return func(r *Resolver) {
		r.categoryService = categoryService
	}
}

func WithOrderService(orderService service.OrderService) Options {
	return func(r *Resolver) {
		r.orderService = orderService
	}
}

func WithProductService(productService service.ProductService) Options {
	return func(r *Resolver) {
		r.productService = productService
	}
}

func WithDiscountService(discountService service.DiscountService) Options {
	return func(r *Resolver) {
		r.discountService = discountService
	}
}

func WithUserService(userService service.UserService) Options {
	return func(r *Resolver) {
		r.userService = userService
	}
}

func WithUploadService(uploadService service.UploadService) Options {
	return func(r *Resolver) {
		r.uploadService = uploadService
	}
}

func WithConfig(cfg *config.Config) Options {
	return func(r *Resolver) {
		r.cfg = cfg
	}
}

func NewResolver(opts ...Options) *Resolver {
	resolver := &Resolver{}
	for _, opt := range opts {
		opt(resolver)
	}
	return resolver
}
