package cmd

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/metrics"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/postgresql"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/publisher"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/redis"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/tracing"
	graphqlHandler "github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/graph/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/graph/resolver"
	grpcHandler "github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/grpc/protos"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/route"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/server/grpcserver"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/server/httpserver"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/oauth"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/uploadStrategy"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the GopherMarket API application servers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		/*----------Slog Logger----------*/
		sLogger := logger.NewSlogLogger()

		/*----------Config----------*/
		cfg, err := config.GetConfigInstance()
		if err != nil {
			sLogger.Error("Failed to get config instance", "err", err)
			return
		}

		/*---------- Tracing ----------*/
		tracerShutdown, err := tracing.Setup(ctx, sLogger, cfg, "gophermarket-api")
		if err != nil {
			sLogger.Error("failed to setup tracing", "err", err)
			os.Exit(1)
		}
		defer func() {
			if err := tracerShutdown(context.Background()); err != nil {
				sLogger.Error("tracer shutdown error", "err", err)
			}
		}()

		authTracer := otel.Tracer("service.auth")
		userTracer := otel.Tracer("service.user")
		categoryTracer := otel.Tracer("service.category")
		productTracer := otel.Tracer("service.product")
		cartTracer := otel.Tracer("service.cart")
		orderTracer := otel.Tracer("service.order")
		uploadTracer := otel.Tracer("service.upload")

		/*---------- Metrics ----------*/
		metricsShutdown, metricsHandler, err := metrics.Setup(&cfg.Metrics, sLogger, "gophermarket-api")
		if err != nil {
			sLogger.Error("failed to setup metrics", "err", err)
			os.Exit(1)
		}
		defer func() {
			if err := metricsShutdown(context.Background()); err != nil {
				sLogger.Error("metrics shutdown error", "err", err)
			}
		}()

		if cfg.Metrics.Enabled {
			if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(5 * time.Second)); err != nil {
				sLogger.Error("failed to start runtime metrics", "err", err)
			}
		}

		if cfg.Metrics.Enabled && metricsHandler != nil {
			go func() {
				if err := metrics.Serve(ctx, sLogger, metricsHandler, cfg.Metrics.Port); err != nil {
					sLogger.Error("metrics server error", "err", err)
				}
			}()
		}

		/*----------Redis----------*/
		redisCfg := redis.NewRedis(
			redis.WithHost(cfg.Redis.Host),
			redis.WithPort(cfg.Redis.Port),
			redis.WithDB(cfg.Redis.DB),
			redis.WithPassword(cfg.Redis.Password),
			redis.WithDialTimeout(cfg.Redis.DialTimeout),
			redis.WithReadTimeout(cfg.Redis.ReadTimeout),
			redis.WithWriteTimeout(cfg.Redis.WriteTimeout),
			redis.WithPoolSize(cfg.Redis.PoolSize),
			redis.WithPoolTimeout(cfg.Redis.PoolTimeout),
		)

		redisClient, err := redisCfg.Connect(ctx)
		if err != nil {
			sLogger.Error("failed to connect to redis", "err", err)
			return
		}
		defer func() {
			if err := redisClient.Close(); err != nil {
				sLogger.Error("failed to close redis", "err", err)
			}
		}()

		/*----------Postgresql----------*/
		pgCfg := postgresql.NewPostgresql(
			postgresql.WithHost(cfg.Postgresql.Host),
			postgresql.WithPort(cfg.Postgresql.Port),
			postgresql.WithUser(cfg.Postgresql.User),
			postgresql.WithPassword(cfg.Postgresql.Password),
			postgresql.WithName(cfg.Postgresql.Name),
			postgresql.WithMaxOpenConn(cfg.Postgresql.MaxOpenConn),
			postgresql.WithMaxIdleConn(cfg.Postgresql.MaxIdleConn),
			postgresql.WithMaxIdleTime(cfg.Postgresql.MaxIdleTime),
			postgresql.WithSSLMode(cfg.Postgresql.SSLMode),
			postgresql.WithTimeout(cfg.Postgresql.Timeout),
			postgresql.WithLogger(sLogger),
		)

		db, err := pgCfg.Connect()
		if err != nil {
			sLogger.Error("failed to connect database", "error", err)
			return
		}

		allowedOrigins := []string{
			"https://localhost:4000",
		}

		/*----------Dependencies----------*/
		middlewares := middleware.NewMiddleware(cfg, sLogger, allowedOrigins)
		watermillPublisher, err := publisher.NewWatermillPublisher(ctx, cfg)
		if err != nil {
			sLogger.Error("failed to create publisher", "err", err)
			return
		}
		googleOAuth := oauth.NewGoogleOAuth(cfg.GoogleOAuth.ClientID)

		/*----------Upload Strategy----------*/
		var uploadStrategies uploadStrategy.UploadStrategy
		if cfg.Upload.UploadProviders == "s3" {
			uploadStrategies = uploadStrategy.NewS3Strategy(cfg)
		} else {
			uploadStrategies = uploadStrategy.NewLocalStrategy(cfg.Upload.Path)
		}

		/*----------Repositories----------*/
		cacheRepository := repository.NewRedisCache(redisClient)
		userRepository := repository.NewUserRepository(db, db)
		tokenRepository := repository.NewTokenRepository(db, db)
		cartRepository := repository.NewCartRepository(db, db)
		cartItemRepository := repository.NewCartItemRepository(db, db)
		categoryRepository := repository.NewCategoryRepository(db, db)
		productRepository := repository.NewProductRepository(db, db)
		orderRepository := repository.NewOrderRepository(db, db)
		resetTokenRepository := repository.NewResetTokenRepository(redisClient)

		/*----------Services----------*/
		authService := service.NewAuthService(userRepository, cartRepository, tokenRepository, resetTokenRepository, googleOAuth, watermillPublisher, cfg, sLogger, authTracer)
		userService := service.NewUserService(userRepository, userTracer)
		categoryService := service.NewCategoryService(categoryRepository, categoryTracer)
		productService := service.NewProductService(productRepository, cacheRepository, productTracer)
		uploadService := service.NewUploadService(uploadStrategies, uploadTracer)
		cartService := service.NewCartService(cartRepository, cartItemRepository, productRepository, cartTracer)
		orderService := service.NewOrderService(orderRepository, cartRepository, cartItemRepository, productRepository, db, orderTracer)

		/*----------GraphQL----------*/
		graphQLResolver := resolver.NewResolver(
			resolver.WithAuthService(authService),
			resolver.WithUserService(userService),
			resolver.WithCategoryService(categoryService),
			resolver.WithProductService(productService),
			resolver.WithCartService(cartService),
			resolver.WithOrderService(orderService),
			resolver.WithUploadService(uploadService),
			resolver.WithConfig(cfg),
		)

		gqlHandler := graphqlHandler.NewGraphQLHandler(graphQLResolver)
		tracedGQLHandler := otelhttp.NewHandler(gqlHandler, "graphql")
		graphqlRoute := route.NewGraphQLRoute(tracedGQLHandler, middlewares)

		/*----------Handlers----------*/
		healthHandler := handler.NewHealthCheckHandler(cfg)
		authHandler := handler.NewAuthHandler(authService, cfg)
		userHandler := handler.NewUserHandler(userService)
		categoryHandler := handler.NewCategoryHandler(categoryService)
		productHandler := handler.NewProductHandler(productService, uploadService)
		cartHandler := handler.NewCartHandler(cartService)
		orderHandler := handler.NewOrderHandler(orderService)

		authGrpcHandler := grpcHandler.NewAuthGrpcHandler(authService)
		cartGrpcHandler := grpcHandler.NewCartGrpcHandler(cartService)
		categoryGrpcHandler := grpcHandler.NewCategoryGrpcHandler(categoryService)
		orderGrpcHandler := grpcHandler.NewOrderGrpcHandler(orderService)
		productGrpcHandler := grpcHandler.NewProductGrpcHandler(productService)
		userGrpcHandler := grpcHandler.NewUserGrpcHandler(userService)

		/*----------Routes----------*/
		healthRoute := route.NewHealthCheckRoute(healthHandler)
		authRoute := route.NewAuthRoute(authHandler, middlewares)
		userRoute := route.NewUserRoute(middlewares, userHandler)
		categoryRoute := route.NewCategoryRoute(middlewares, categoryHandler)
		productRoute := route.NewProductRoute(middlewares, productHandler)
		cartRoute := route.NewCartRoute(middlewares, cartHandler)
		orderRoute := route.NewOrderRoute(middlewares, orderHandler)

		/*----------Route Registry----------*/
		routes := route.NewRegisterRoute(
			route.WithMiddleware(middlewares),
			route.WithHealthCheckRoute(healthRoute),
			route.WithAuthRoute(authRoute),
			route.WithUserRoute(userRoute),
			route.WithCategoryRoute(categoryRoute),
			route.WithProductRoute(productRoute),
			route.WithCartRoute(cartRoute),
			route.WithOrderRoute(orderRoute),
			route.WithGraphQLRoute(graphqlRoute),
		)

		/*---------- gRPC Server Initialization ----------*/
		grpcServer := grpcserver.NewGrpcServer(
			grpcserver.WithHost(cfg.GrpcServer.Host),
			grpcserver.WithPort(cfg.GrpcServer.Port),
			grpcserver.WithLogger(sLogger),
			grpcserver.WithGrpcOptions(grpc.StatsHandler(otelgrpc.NewServerHandler())),
		)

		protos.RegisterAuthServiceServer(grpcServer.GetServer(), authGrpcHandler)
		protos.RegisterCartServiceServer(grpcServer.GetServer(), cartGrpcHandler)
		protos.RegisterCategoryServiceServer(grpcServer.GetServer(), categoryGrpcHandler)
		protos.RegisterOrderServiceServer(grpcServer.GetServer(), orderGrpcHandler)
		protos.RegisterProductServiceServer(grpcServer.GetServer(), productGrpcHandler)
		protos.RegisterUserServiceServer(grpcServer.GetServer(), userGrpcHandler)

		/*---------- HTTP Server Initialization ----------*/
		httpServerOpts := []httpserver.Option{
			httpserver.WithHost(cfg.HTTPServer.Host),
			httpserver.WithPort(cfg.HTTPServer.Port),
			httpserver.WithHandler(routes.RegisterRoutes()),
			httpserver.WithReadTimeout(cfg.HTTPServer.ReadTimeout),
			httpserver.WithWriteTimeout(cfg.HTTPServer.WriteTimeout),
			httpserver.WithIdleTimeout(cfg.HTTPServer.IdleTimeout),
			httpserver.WithErrorLog(slog.NewLogLogger(sLogger.Handler(), slog.LevelError)),
			httpserver.WithLogger(sLogger),
		}

		if cfg.HTTPServer.CertFile != "" && cfg.HTTPServer.KeyFile != "" {
			httpServerOpts = append(httpServerOpts, httpserver.WithCert(cfg.HTTPServer.CertFile), httpserver.WithKey(cfg.HTTPServer.KeyFile))
		}
		httpServer := httpserver.NewHTTPServer(httpServerOpts...)

		/*---------- Execution and Lifecycles ----------*/
		serverErrors := make(chan error, 2)

		// Launch gRPC Engine
		go func() {
			sLogger.Info("starting gRPC server", "addr", cfg.GrpcServer.Host+":"+cfg.GrpcServer.Port)
			if err := grpcServer.Connect(); err != nil {
				serverErrors <- fmt.Errorf("gRPC server run error: %w", err)
			}
		}()

		// Launch HTTP Engine
		go func() {
			if err := httpServer.Connect(); err != nil {
				serverErrors <- fmt.Errorf("HTTP server run error: %w", err)
			}
		}()

		shutdownChan := make(chan os.Signal, 1)
		signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-serverErrors:
			sLogger.Error("critical runtime server error", "err", err)
		case sig := <-shutdownChan:
			sLogger.Info("system signal intercepted, initializing termination layout", "signal", sig.String())

			cancel()

			sLogger.Info("halting gRPC listeners gracefully...")
			grpcServer.GracefulStop()
			sLogger.Info("gRPC server context dropped cleanly")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
