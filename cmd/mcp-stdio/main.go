// Command mcp-stdio runs the svarog MCP server over stdio (for Cursor / local agents).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	// ServeStdio blocks until the client disconnects or ctx is cancelled.
	done := make(chan error, 1)
	go func() { done <- mcpserver.ServeStdio(s) }()

	select {
	case <-ctx.Done():
		return nil
	case err := <-done:
		return err
	}
}
