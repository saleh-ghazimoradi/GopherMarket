package route

import (
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

type RegisterRoute struct {
	middleware       *middleware.Middleware
	healthCheckRoute *HealthCheckRoute
	authRoute        *AuthRoute
	userRoute        *UserRoute
	categoryRoute    *CategoryRoute
	productRoute     *ProductRoute
	cartRoute        *CartRoute
	orderRoute       *OrderRoute
	discountRoute    *DiscountRoute
	GraphQLRoute     *GraphQLRoute
}

type Options func(*RegisterRoute)

func WithHealthCheckRoute(healthCheckRoute *HealthCheckRoute) Options {
	return func(r *RegisterRoute) {
		r.healthCheckRoute = healthCheckRoute
	}
}

func WithMiddleware(middleware *middleware.Middleware) Options {
	return func(r *RegisterRoute) {
		r.middleware = middleware
	}
}

func WithAuthRoute(authRoute *AuthRoute) Options {
	return func(r *RegisterRoute) {
		r.authRoute = authRoute
	}
}

func WithUserRoute(userRoute *UserRoute) Options {
	return func(r *RegisterRoute) {
		r.userRoute = userRoute
	}
}

func WithCategoryRoute(categoryRoute *CategoryRoute) Options {
	return func(r *RegisterRoute) {
		r.categoryRoute = categoryRoute
	}
}

func WithProductRoute(productRoute *ProductRoute) Options {
	return func(r *RegisterRoute) {
		r.productRoute = productRoute
	}
}

func WithCartRoute(cartRoute *CartRoute) Options {
	return func(r *RegisterRoute) {
		r.cartRoute = cartRoute
	}
}

func WithOrderRoute(orderRoute *OrderRoute) Options {
	return func(r *RegisterRoute) {
		r.orderRoute = orderRoute
	}
}

func WithDiscountRoute(discountRoute *DiscountRoute) Options {
	return func(r *RegisterRoute) {
		r.discountRoute = discountRoute
	}
}

func WithGraphQLRoute(graphQLRoute *GraphQLRoute) Options {
	return func(r *RegisterRoute) {
		r.GraphQLRoute = graphQLRoute
	}
}

func (r *RegisterRoute) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /swagger/*any", httpSwagger.Handler(httpSwagger.URL("/docs/swagger.json")))
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))))
	mux.Handle("GET /api-docs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/rapidoc.html")
	}))

	r.healthCheckRoute.HealthCheckRoutes(mux)
	r.authRoute.AuthRoutes(mux)
	r.userRoute.UserRoutes(mux)
	r.categoryRoute.CategoryRoutes(mux)
	r.productRoute.ProductRoutes(mux)
	r.cartRoute.CartRoutes(mux)
	r.orderRoute.OrderRoute(mux)
	r.discountRoute.DiscountRoutes(mux)
	r.GraphQLRoute.GraphQLRoutes(mux)

	//hppOpts := middleware.HPPOptions{
	//	CheckQuery:                  true,
	//	CheckBody:                   false,
	//	CheckBodyOnlyForContentType: "",
	//	Whitelist:                   []string{"email", "password", "name"},
	//}

	return middleware.Chain(mux,
		r.middleware.Recover,         // 1st Layer: Catches panics at the absolute boundary
		r.middleware.Compression,     // 2nd Layer: Compresses outbound stream right before the wire
		r.middleware.Tracing,         // 3rd Layer: Captures panics *inside* the trace span to log errors
		r.middleware.Metrics,         // 4th Layer: Accurately counts uncompressed metrics & statuses
		r.middleware.Logging,         // 5th Layer: Logs clean, uncompressed request/response data
		r.middleware.SecurityHeaders, // 6th Layer: Evaluates secure browser headers cleanly
		r.middleware.CORS,            // 7th Layer: Validates origins early
		r.middleware.Limiter,         // 8th Layer: Drops bad/flood traffic before hitting the app engine
	)
}

func NewRegisterRoute(opts ...Options) *RegisterRoute {
	r := &RegisterRoute{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}
