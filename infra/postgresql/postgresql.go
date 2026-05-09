package postgresql

import (
	"context"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log/slog"
	"time"
)

type Postgresql struct {
	host        string
	port        string
	user        string
	password    string
	name        string
	maxOpenConn int
	maxIdleConn int
	maxIdleTime time.Duration
	sslMode     string
	timeout     time.Duration
	logger      *slog.Logger
}

type Options func(*Postgresql)

// WithHost sets the database host.
func WithHost(host string) Options {
	return func(p *Postgresql) {
		p.host = host
	}
}

// WithPort sets the database port.
func WithPort(port string) Options {
	return func(p *Postgresql) {
		p.port = port
	}
}

// WithUser sets the database user.
func WithUser(user string) Options {
	return func(p *Postgresql) {
		p.user = user
	}
}

// WithPassword sets the database password.
func WithPassword(password string) Options {
	return func(p *Postgresql) {
		p.password = password
	}
}

// WithName sets the database name.
func WithName(name string) Options {
	return func(p *Postgresql) {
		p.name = name
	}
}

// WithMaxOpenConn sets the maximum number of open connections.
func WithMaxOpenConn(maxOpenConn int) Options {
	return func(p *Postgresql) {
		p.maxOpenConn = maxOpenConn
	}
}

// WithMaxIdleConn sets the maximum number of idle connections.
func WithMaxIdleConn(maxIdleConn int) Options {
	return func(p *Postgresql) {
		p.maxIdleConn = maxIdleConn
	}
}

// WithMaxIdleTime sets the maximum idle time for a connection.
func WithMaxIdleTime(maxIdleTime time.Duration) Options {
	return func(p *Postgresql) {
		p.maxIdleTime = maxIdleTime
	}
}

// WithSSLMode sets the SSL mode (e.g., "disable", "require").
func WithSSLMode(mode string) Options {
	return func(p *Postgresql) {
		p.sslMode = mode
	}
}

// WithTimeout sets the timeout for connection establishment and ping.
func WithTimeout(timeout time.Duration) Options {
	return func(p *Postgresql) {
		p.timeout = timeout
	}
}

// WithLogger sets the structured logger.
func WithLogger(logger *slog.Logger) Options {
	return func(p *Postgresql) {
		p.logger = logger
	}
}

func (p *Postgresql) uri() string {
	connectTimeoutSeconds := 10
	if p.timeout > 0 {
		connectTimeoutSeconds = int(p.timeout.Seconds())
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC connect_timeout=%d",
		p.host, p.user, p.password, p.name, p.port, p.sslMode, connectTimeoutSeconds,
	)
}

func (p *Postgresql) Connect() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(p.uri()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		p.logger.Error("failed to open database connection", "error", err)
		return nil, fmt.Errorf("gorm open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		p.logger.Error("failed to get underlying sql.DB", "error", err)
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(p.maxOpenConn)
	sqlDB.SetMaxIdleConns(p.maxIdleConn)
	sqlDB.SetConnMaxIdleTime(p.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		p.logger.Error("failed to ping database", "error", err)
		return nil, fmt.Errorf("ping database: %w", err)
	}

	p.logger.Info("successfully connected to PostgreSQL",
		"host", p.host,
		"port", p.port,
		"dbname", p.name,
	)
	return db, nil

}

func NewPostgresql(opts ...Options) *Postgresql {
	p := &Postgresql{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
