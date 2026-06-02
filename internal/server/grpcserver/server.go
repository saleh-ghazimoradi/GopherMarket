package grpcserver

import (
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type GrpcServer struct {
	server *grpc.Server
	host   string
	port   string
	logger *slog.Logger
}

type Option func(*GrpcServer)

func WithHost(host string) Option {
	return func(g *GrpcServer) {
		g.host = host
	}
}

func WithPort(port string) Option {
	return func(g *GrpcServer) {
		g.port = port
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(g *GrpcServer) {
		g.logger = logger
	}
}

func WithGrpcOptions(opts ...grpc.ServerOption) Option {
	return func(g *GrpcServer) {
		g.server = grpc.NewServer(opts...)
	}
}

func (s *GrpcServer) GetServer() *grpc.Server {
	return s.server
}

func (s *GrpcServer) Connect() error {
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	if s.logger != nil {
		s.logger.Info("gRPC server is running", "addr", addr)
	}
	return s.server.Serve(lis)
}

func (s *GrpcServer) GracefulStop() {
	s.server.GracefulStop()
}

func NewGrpcServer(opts ...Option) *GrpcServer {
	g := &GrpcServer{}
	for _, opt := range opts {
		opt(g)
	}
	if g.server == nil {
		g.server = grpc.NewServer()
	}
	return g
}
