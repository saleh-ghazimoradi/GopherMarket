package handler

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/graph"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/graph/resolver"
	"github.com/vektah/gqlparser/v2/ast"
	"net/http"
)

type GraphQLHandler struct {
	handler http.Handler
}

func (g *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.handler.ServeHTTP(w, r)
}

func NewGraphQLHandler(r *resolver.Resolver) *GraphQLHandler {
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: r})
	srv := handler.New(schema)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return &GraphQLHandler{handler: srv}
}
