package cmd

import (
	"fmt"
	"github.com/saleh-ghazimoradi/GopherMarket/config"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/migration"
	"github.com/saleh-ghazimoradi/GopherMarket/infra/postgresql"
	"github.com/saleh-ghazimoradi/GopherMarket/internal/logger"

	"github.com/spf13/cobra"
)

// migrateUpCmd represents the migrateUp command
var migrateUpCmd = &cobra.Command{
	Use:   "migrateUp",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("migrateUp called")

		sLogger := logger.NewSlogLogger()

		cfg, err := config.GetConfigInstance()
		if err != nil {
			sLogger.Error("failed to load config", "error", err)
			return
		}

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

		sqlDB, err := db.DB()
		if err != nil {
			sLogger.Error("failed to get sql.DB from gorm", "error", err)
			return
		}

		defer func() {
			if err := sqlDB.Close(); err != nil {
				sLogger.Error("error closing database connection", "error", err)
			}
		}()

		migrator, err := migration.NewMigrator(sqlDB, cfg.Postgresql.Name)
		if err != nil {
			sLogger.Error("migration init failed", "error", err)
			return
		}
		defer func() {
			if err := migrator.Close(); err != nil {
				sLogger.Error("failed to close migrator", "error", err)
			}
		}()

		if err := migrator.Up(); err != nil {
			sLogger.Error("migration failed", "error", err)
			return
		}

		sLogger.Info("Migrations applied successfully")
	},
}

func init() {
	rootCmd.AddCommand(migrateUpCmd)
}
