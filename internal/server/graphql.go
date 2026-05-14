package server

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/graph"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/graph/resolver"
	"github.com/vektah/gqlparser/v2/ast"
)

type GraphqlServer struct {
	resolver *resolver.Resolver
}

func (g *GraphqlServer) Connect() *handler.Server {
	schema := graph.NewExecutableSchema(graph.Config{Resolvers: g.resolver})
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

	return srv
}

func NewGraphqlServer(resolver *resolver.Resolver) *GraphqlServer {
	return &GraphqlServer{
		resolver: resolver,
	}
}
