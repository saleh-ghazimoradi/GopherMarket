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
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/route"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/server"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/oauth"
	"github.com/saleh-ghazimoradi/GopherMarket/pkg/uploadStrategy"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"log/slog"
	"os"
	"time"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
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

		// Start runtime metrics (goroutines, memory, etc.)
		if cfg.Metrics.Enabled {
			if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(5 * time.Second)); err != nil {
				sLogger.Error("failed to start runtime metrics", "err", err)
			}
		}

		// Start metrics HTTP server in background
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

		redis, err := redisCfg.Connect(ctx)
		if err != nil {
			sLogger.Error("failed to connect to redis", "err", err)
			return
		}

		defer func() {
			if err := redis.Close(); err != nil {
				sLogger.Error("failed to close redis", "err", err)
				return
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

		/*----------Dependencies----------*/
		middlewares := middleware.NewMiddleware(sLogger, cfg)
		watermillPublisher, err := publisher.NewWatermillPublisher(ctx, cfg)
		if err != nil {
			sLogger.Error("failed to create publisher", "err", err)
			return
		}
		googleOAuth := oauth.NewGoogleOAuth(cfg.GoogleOAuth.ClientId)

		/*----------Upload Strategy----------*/
		var uploadStrategies uploadStrategy.UploadStrategy
		if cfg.Upload.UploadProviders == "s3" {
			uploadStrategies = uploadStrategy.NewS3Strategy(cfg)
		} else {
			uploadStrategies = uploadStrategy.NewLocalStrategy(cfg.Upload.Path)
		}

		/*----------Repositories----------*/
		cacheRepository := repository.NewRedisCache(redis)
		userRepository := repository.NewUserRepository(db, db)
		tokenRepository := repository.NewTokenRepository(db, db)
		cartRepository := repository.NewCartRepository(db, db)
		cartItemRepository := repository.NewCartItemRepository(db, db)
		categoryRepository := repository.NewCategoryRepository(db, db)
		productRepository := repository.NewProductRepository(db, db)
		orderRepository := repository.NewOrderRepository(db, db)
		resetTokenRepository := repository.NewResetTokenRepository(redis)

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

		/*----------Routes----------*/
		healthRoute := route.NewHealthCheckRoute(healthHandler)
		authRoute := route.NewAuthRoute(authHandler)
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

		/*----------HTTP Server----------*/
		serverOpts := []server.Option{
			server.WithHost(cfg.Server.Host),
			server.WithPort(cfg.Server.Port),
			server.WithHandler(routes.RegisterRoutes()),
			server.WithReadTimeout(cfg.Server.ReadTimeout),
			server.WithWriteTimeout(cfg.Server.WriteTimeout),
			server.WithIdleTimeout(cfg.Server.IdleTimeout),
			server.WithErrorLog(slog.NewLogLogger(sLogger.Handler(), slog.LevelError)),
			server.WithLogger(sLogger),
		}
		if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
			serverOpts = append(serverOpts, server.WithCert(cfg.Server.CertFile), server.WithKey(cfg.Server.KeyFile))
		}

		httpServer := server.NewServer(serverOpts...)

		sLogger.Info("starting server",
			"addr", cfg.Server.Host+":"+cfg.Server.Port,
			"env", cfg.Application.Environment,
			"tls", cfg.Server.CertFile != "",
		)

		if err := httpServer.Connect(); err != nil {
			sLogger.Error("failed to connect to http server", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
