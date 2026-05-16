package config

import (
	"github.com/caarlos0/env/v11"
	"sync"
	"time"
)

var (
	instance *Config
	initErr  error
	once     sync.Once
)

type Config struct {
	Application Application
	Redis       Redis
	Postgresql  Postgresql
	Server      Server
	JWT         JWT
	AWS         AWS
	Upload      Upload
	SMTP        SMTP
	Event       Event
	RateLimiter RateLimiter
}

type Event struct {
	UserLoggedIn string `env:"USER_LOGGED_IN"`
}

type RateLimiter struct {
	RPS     float64 `env:"RPS"`
	Burst   int     `env:"BURST"`
	Enabled bool    `env:"ENABLED"`
}

type Application struct {
	Version     string `env:"APP_VERSION"`
	Environment string `env:"APP_ENVIRONMENT"`
}

type Redis struct {
	Host         string        `env:"REDIS_HOST"`
	Port         string        `env:"REDIS_PORT"`
	Password     string        `env:"REDIS_PASSWORD"`
	DB           int           `env:"REDIS_DB"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT"`
	PoolSize     int           `env:"REDIS_POOL_SIZE"`
	PoolTimeout  time.Duration `env:"REDIS_POOL_TIMEOUT"`
	RPM          int           `env:"REDIS_RPM"`
}

type Postgresql struct {
	Host        string        `env:"POSTGRES_HOST"`
	Port        string        `env:"POSTGRES_PORT"`
	User        string        `env:"POSTGRES_USER"`
	Password    string        `env:"POSTGRES_PASSWORD"`
	Name        string        `env:"POSTGRES_NAME"`
	MaxOpenConn int           `env:"POSTGRES_MAX_OPEN_CONN"`
	MaxIdleConn int           `env:"POSTGRES_MAX_IDLE_CONN"`
	MaxIdleTime time.Duration `env:"POSTGRES_MAX_IDLE_TIME"`
	SSLMode     string        `env:"POSTGRES_SSL_MODE"`
	Timeout     time.Duration `env:"POSTGRES_TIMEOUT"`
}

type Server struct {
	Host         string        `env:"SERVER_HOST"`
	Port         string        `env:"SERVER_PORT"`
	IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT"`
}

type JWT struct {
	Secret              string        `env:"JWT_SECRET"`
	ExpiresIn           time.Duration `env:"JWT_EXPIRES_IN"`
	RefreshTokenExpires time.Duration `env:"JWT_REFRESH_TOKEN_EXPIRES"`
}

type Upload struct {
	Path            string `env:"UPLOAD_PATH"`
	MaxFileSize     int64  `env:"UPLOAD_MAX_FILE_SIZE"`
	UploadProviders string `env:"UPLOAD_PROVIDERS"`
}

type SMTP struct {
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
	From     string `env:"SMTP_FROM"`
}

type AWS struct {
	Region          string `env:"AWS_REGION"`
	AccessKeyId     string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	S3Bucket        string `env:"AWS_S3_BUCKET"`
	S3Endpoint      string `env:"AWS_S3_ENDPOINT"`
	EventQueueName  string `env:"AWS_EVENT_QUEUE_NAME"`
}

func GetConfigInstance() (*Config, error) {
	once.Do(func() {
		instance = &Config{}
		initErr = env.Parse(instance)
		if initErr != nil {
			instance = nil
		}
	})
	return instance, initErr
}
