package route

import (
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"net/http"
)

type GraphQLRoute struct {
	gqlHandler http.Handler
	middleware *middleware.Middleware
}

func (g *GraphQLRoute) GraphQLRoutes(mux *http.ServeMux) {
	mux.Handle("/graphql", g.middleware.GraphQLAuth(g.gqlHandler))
	mux.Handle("/graphql/playground", playground.Handler("GraphQL Playground", "/graphql"))
}

func NewGraphQLRoute(gqlHandler http.Handler, middleware *middleware.Middleware) *GraphQLRoute {
	return &GraphQLRoute{
		gqlHandler: gqlHandler,
		middleware: middleware,
	}
}
