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
	r.GraphQLRoute.GraphQLRoutes(mux)
	return r.middleware.Recover(r.middleware.Logging(r.middleware.CORS(mux)))
}

func NewRegisterRoute(opts ...Options) *RegisterRoute {
	r := &RegisterRoute{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}
