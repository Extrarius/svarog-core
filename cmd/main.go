// Command svarog is the entry point of the svarog-core monolith.
//
// main wires together:
//   - configuration (envconfig)
//   - logging (slog + OTel logs bridge)
//   - observability (OTel SDK: traces / metrics / logs)
//   - persistence (pgxpool)
//   - business handlers (internal/app)
//   - transport (gRPC + grpc-gateway)
//
// The auth use-cases (Register / Login / Logout / Me) are fully wired:
// requests flow HTTP gateway -> gRPC -> app -> Postgres, with opaque session
// cookies bridged between the gateway and the gRPC auth interceptor.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/Extrarius/svarog-core/internal/adapters/repo"
	"github.com/Extrarius/svarog-core/internal/api/gateway"
	grpcsrv "github.com/Extrarius/svarog-core/internal/api/grpc"
	"github.com/Extrarius/svarog-core/internal/app"
	"github.com/Extrarius/svarog-core/internal/auth"
	"github.com/Extrarius/svarog-core/internal/config"
	"github.com/Extrarius/svarog-core/internal/logger"
	"github.com/Extrarius/svarog-core/internal/observability"
)

func main() {
	fmt.Println("svarog-core: hello, world")

	if err := run(); err != nil {
		// Use stderr directly so we don't depend on slog being initialised.
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- 1. Configuration ------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// --- 2. Logger -------------------------------------------------------
	log := logger.New(logger.Options{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		ServiceName: cfg.OTel.ServiceName,
	})
	slog.SetDefault(log)
	log.Info("starting svarog-core",
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)

	// --- 3. Observability (OTel SDK) ------------------------------------
	shutdownOTel, err := observability.Init(ctx, observability.Options{
		Endpoint:       cfg.OTel.Endpoint,
		Insecure:       cfg.OTel.Insecure,
		ServiceName:    cfg.OTel.ServiceName,
		ServiceVersion: cfg.OTel.ServiceVersion,
		Environment:    cfg.Env,
	})
	if err != nil {
		return fmt.Errorf("init observability: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := shutdownOTel(shutdownCtx); err != nil {
			log.Error("observability shutdown", slog.Any("err", err))
		}
	}()

	// --- 4. Database -----------------------------------------------------
	pool, err := pgxpool.New(ctx, cfg.DB.DSN())
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Warn("postgres ping failed at startup", slog.Any("err", err))
	}

	// --- 5. Business handlers -------------------------------------------
	handlers, err := app.New(app.Config{
		Users:      repo.NewUsers(pool),
		Sessions:   repo.NewSessions(pool),
		Hasher:     repo.NewBcryptHasher(),
		Tokens:     auth.NewTokenSource(),
		Clock:      repo.NewSystemClock(),
		Logger:     log,
		SessionTTL: cfg.Session.TTL,
	})
	if err != nil {
		return fmt.Errorf("build handlers: %w", err)
	}

	// --- 6. Transport ---------------------------------------------------
	grpcServer, err := grpcsrv.New(grpcsrv.Options{
		Addr:     cfg.GRPCAddr,
		Logger:   log.With(slog.String("component", "grpc")),
		Handlers: handlers,
	})
	if err != nil {
		return fmt.Errorf("build grpc server: %w", err)
	}

	gwServer, err := gateway.New(gateway.Options{
		Addr:         cfg.HTTPAddr,
		Logger:       log.With(slog.String("component", "gateway")),
		GRPCTarget:   grpcServer.Addr(),
		CookieName:   cfg.Session.CookieName,
		CookieDomain: cfg.Session.CookieDomain,
		CookieSecure: cfg.Session.CookieSecure,
	})
	if err != nil {
		return fmt.Errorf("build gateway server: %w", err)
	}

	// --- 7. Run, wait for signal, graceful shutdown ---------------------
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error { return grpcServer.Run(gctx) })
	g.Go(func() error { return gwServer.Run(gctx) })

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server group: %w", err)
	}

	log.Info("svarog-core stopped")
	return nil
}
