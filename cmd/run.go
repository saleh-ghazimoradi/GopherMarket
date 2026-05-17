package cmd

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/postgresql"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/publisher"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/redis"
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
	"log/slog"
	"os"
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
		authService := service.NewAuthService(userRepository, cartRepository, tokenRepository, resetTokenRepository, googleOAuth, watermillPublisher, cfg, sLogger)
		userService := service.NewUserService(userRepository)
		categoryService := service.NewCategoryService(categoryRepository)
		productService := service.NewProductService(productRepository, cacheRepository)
		uploadService := service.NewUploadService(uploadStrategies)
		cartService := service.NewCartService(cartRepository, cartItemRepository, productRepository)
		orderService := service.NewOrderService(orderRepository, cartRepository, cartItemRepository, productRepository, db)

		/*----------GraphQL----------*/
		graphQLResolver := resolver.NewResolver(
			resolver.WithAuthService(authService),
			resolver.WithUserService(userService),
			resolver.WithCategoryService(categoryService),
			resolver.WithProductService(productService),
			resolver.WithCartService(cartService),
			resolver.WithOrderService(orderService),
			resolver.WithUploadService(uploadService),
		)

		gqlHandler := graphqlHandler.NewGraphQLHandler(graphQLResolver)
		graphqlRoute := route.NewGraphQLRoute(gqlHandler, middlewares)

		/*----------Handlers----------*/
		healthHandler := handler.NewHealthCheckHandler(cfg)
		authHandler := handler.NewAuthHandler(authService)
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
		httpServer := server.NewServer(
			server.WithHost(cfg.Server.Host),
			server.WithPort(cfg.Server.Port),
			server.WithHandler(routes.RegisterRoutes()),
			server.WithReadTimeout(cfg.Server.ReadTimeout),
			server.WithWriteTimeout(cfg.Server.WriteTimeout),
			server.WithIdleTimeout(cfg.Server.IdleTimeout),
			server.WithErrorLog(slog.NewLogLogger(sLogger.Handler(), slog.LevelError)),
			server.WithLogger(sLogger),
		)

		sLogger.Info("starting server", "addr", cfg.Server.Host+":"+cfg.Server.Port, "env", cfg.Application.Environment)
		if err := httpServer.Connect(); err != nil {
			sLogger.Error("failed to connect to http server", "err", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
