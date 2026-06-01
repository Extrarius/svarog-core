// Command mcp-http runs the svarog MCP server over streamable HTTP (default /mcp).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mcpserver "github.com/mark3labs/mcp-go/server"

	svarogmcp "github.com/Extrarius/svarog-core/internal/mcp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := svarogmcp.Load()
	if err != nil {
		return err
	}

	deps, err := svarogmcp.NewDeps(ctx, cfg)
	if err != nil {
		return fmt.Errorf("mcp deps: %w", err)
	}
	defer deps.Close()

	s := svarogmcp.Build(deps)
	httpSrv := mcpserver.NewStreamableHTTPServer(s, mcpserver.WithEndpointPath("/mcp"))

	addr := cfg.HTTPAddr
	slog.Info("mcp-http listening", slog.String("addr", addr), slog.String("path", "/mcp"))

	errCh := make(chan error, 1)
	go func() { errCh <- httpSrv.Start(addr) }()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}
