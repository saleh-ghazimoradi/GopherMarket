package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net"
	"time"
)

type Redis struct {
	host         string
	port         string
	db           int
	password     string
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	poolSize     int
	poolTimeout  time.Duration
}

type Option func(*Redis)

func WithHost(host string) Option {
	return func(r *Redis) {
		r.host = host
	}
}

func WithPort(port string) Option {
	return func(r *Redis) {
		r.port = port
	}
}

func WithDB(db int) Option {
	return func(r *Redis) {
		r.db = db
	}
}

func WithPassword(password string) Option {
	return func(r *Redis) {
		r.password = password
	}
}

func WithDialTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.dialTimeout = timeout
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.readTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.writeTimeout = timeout
	}
}

func WithPoolSize(poolSize int) Option {
	return func(r *Redis) {
		r.poolSize = poolSize
	}
}

func WithPoolTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.poolTimeout = timeout
	}
}

func (r *Redis) uri() string {
	return net.JoinHostPort(r.host, r.port)
}

func (r *Redis) Connect(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         r.uri(),
		DB:           r.db,
		Password:     r.password,
		DialTimeout:  r.dialTimeout,
		ReadTimeout:  r.readTimeout,
		WriteTimeout: r.writeTimeout,
		PoolSize:     r.poolSize,
		PoolTimeout:  r.poolTimeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping failed: %w (and failed to close: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return client, nil
}

func NewRedis(opts ...Option) *Redis {
	r := &Redis{}
	for _, o := range opts {
		o(r)
	}
	return r
}
