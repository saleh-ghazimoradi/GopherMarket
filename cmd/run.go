package cmd

import (
	"context"
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/postgresql"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/redis"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/handler"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/middleware"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/gateway/route"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/repository"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/server"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/service"
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
		middlewares := middleware.NewMiddleware(sLogger)

		/*----------Repositories----------*/
		userRepository := repository.NewUserRepository(db, db)
		cartRepository := repository.NewCartRepository(db, db)
		tokenRepository := repository.NewTokenRepository(db, db)

		/*----------Services----------*/
		authService := service.NewAuthService(userRepository, cartRepository, tokenRepository, cfg)

		/*----------Handlers----------*/
		healthHandler := handler.NewHealthCheckHandler(cfg)
		authHandler := handler.NewAuthHandler(authService)

		/*----------Routes----------*/
		healthRoute := route.NewHealthCheckRoute(healthHandler)
		authRoute := route.NewAuthRoute(authHandler)

		/*----------Route Registry----------*/
		routes := route.NewRegisterRoute(
			route.WithMiddleware(middlewares),
			route.WithHealthCheckRoute(healthRoute),
			route.WithAuthRoute(authRoute),
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
