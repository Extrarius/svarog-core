// Package config loads application configuration from environment variables.
//
// Values mirror the keys documented in .env.example. The package is intended
// to be called once from cmd/main.go; subsystems receive the already-parsed
// substructs (DB, OTel, Session, …) through their constructors.
package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// App captures the full runtime configuration of svarog-core.
type App struct {
	Env     string `envconfig:"APP_ENV" default:"local"`
	Name    string `envconfig:"APP_NAME" default:"svarog-core"`
	Version string `envconfig:"APP_VERSION" default:"0.0.1"`

	GRPCAddr string `envconfig:"GRPC_ADDR" default:":9090"`
	HTTPAddr string `envconfig:"HTTP_ADDR" default:":8080"`

	DB      DB
	Session Session
	OTel    OTel
	Log     Log
}

// DB groups Postgres connection settings.
type DB struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     int    `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" default:"svarog"`
	Password string `envconfig:"POSTGRES_PASSWORD" default:"svarog"`
	Database string `envconfig:"POSTGRES_DB" default:"svarog"`
	SSLMode  string `envconfig:"POSTGRES_SSLMODE" default:"disable"`
}

// DSN returns a libpq-style connection string suitable for pgxpool.
func (d DB) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

// Session captures cookie + lifetime settings for opaque session tokens.
type Session struct {
	CookieName   string        `envconfig:"SESSION_COOKIE_NAME" default:"sid"`
	TTL          time.Duration `envconfig:"SESSION_TTL" default:"24h"`
	CookieDomain string        `envconfig:"SESSION_COOKIE_DOMAIN" default:""`
	CookieSecure bool          `envconfig:"SESSION_COOKIE_SECURE" default:"false"`
}

// OTel groups OpenTelemetry exporter settings.
type OTel struct {
	Endpoint       string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
	Insecure       bool   `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"true"`
	ServiceName    string `envconfig:"OTEL_SERVICE_NAME" default:"svarog-core"`
	ServiceVersion string `envconfig:"OTEL_SERVICE_VERSION" default:"0.0.1"`
}

// Log captures slog handler settings.
type Log struct {
	Level  string `envconfig:"LOG_LEVEL" default:"info"`  // debug | info | warn | error
	Format string `envconfig:"LOG_FORMAT" default:"json"` // json | text
}

// Load reads configuration from the process environment.
func Load() (App, error) {
	var cfg App
	if err := envconfig.Process("", &cfg); err != nil {
		return App{}, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
