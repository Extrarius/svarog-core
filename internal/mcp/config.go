// Package mcp implements the svarog MCP server (tools, resources, prompts).
package mcp

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds runtime settings for MCP binaries.
type Config struct {
	HTTPAddr string `envconfig:"MCP_HTTP_ADDR" default:":8000"`

	DBHost     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	DBPort     int    `envconfig:"POSTGRES_PORT" default:"5432"`
	DBUser     string `envconfig:"POSTGRES_USER" default:"svarog"`
	DBPassword string `envconfig:"POSTGRES_PASSWORD" default:"svarog"`
	DBName     string `envconfig:"POSTGRES_DB" default:"svarog"`
	DBSSLMode  string `envconfig:"POSTGRES_SSLMODE" default:"disable"`

	// SvarogHTTPBase is the public HTTP gateway base URL (no trailing slash),
	// e.g. http://localhost:8080 or http://app:8080 inside compose.
	SvarogHTTPBase string `envconfig:"SVAROG_HTTP_BASE" default:"http://localhost:8080"`

	AppName    string `envconfig:"APP_NAME" default:"svarog-core"`
	AppVersion string `envconfig:"APP_VERSION" default:"0.0.1"`
}

// DSN returns a libpq connection string for pgx.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// Load reads MCP configuration from the environment.
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("mcp config: %w", err)
	}
	return cfg, nil
}
