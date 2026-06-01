package mcp

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Deps groups collaborators injected into MCP tool handlers.
type Deps struct {
	Pool       *pgxpool.Pool
	HTTPClient *http.Client
	Config     Config
}

// Close releases resources held by Deps.
func (d *Deps) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

// NewDeps connects to Postgres and prepares an HTTP client for svarog API calls.
func NewDeps(ctx context.Context, cfg Config) (*Deps, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Deps{
		Pool:       pool,
		HTTPClient: http.DefaultClient,
		Config:     cfg,
	}, nil
}
