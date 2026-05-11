// Package gateway hosts the HTTP transport adapter built on grpc-gateway.
//
// The Server type owns an *http.Server that fronts the gRPC server via a
// runtime.ServeMux. Once `make proto-gen` has produced
// api/gen/go/auth/v1/auth.pb.gw.go, the Auth handler should be registered as
// documented in the TODO inside New.
package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server bundles the HTTP listener and lifecycle helpers.
type Server struct {
	addr   string
	log    *slog.Logger
	server *http.Server
}

// Options describes how the gateway server should be constructed.
type Options struct {
	Addr       string
	Logger     *slog.Logger
	GRPCTarget string // host:port the gateway forwards to (typically Options.Addr of grpc.Server)
}

// New constructs a Server.
func New(opts Options) (*Server, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("gateway: Logger is required")
	}

	mux := runtime.NewServeMux()

	// TODO: after `make proto-gen` add:
	//
	//   import (
	//       authv1 "github.com/Extrarius/svarog-core/api/gen/go/auth/v1"
	//       "google.golang.org/grpc"
	//       "google.golang.org/grpc/credentials/insecure"
	//   )
	//
	//   conn, err := grpc.NewClient(opts.GRPCTarget,
	//       grpc.WithTransportCredentials(insecure.NewCredentials()))
	//   if err != nil { return nil, err }
	//   if err := authv1.RegisterAuthServiceHandler(context.Background(), mux, conn); err != nil { return nil, err }
	_ = opts.GRPCTarget

	handler := otelhttp.NewHandler(mux, "gateway")

	srv := &http.Server{
		Addr:              opts.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &Server{
		addr:   opts.Addr,
		log:    opts.Logger,
		server: srv,
	}, nil
}

// Run serves HTTP until ctx is cancelled. It uses Shutdown on cancellation
// for a graceful drain.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("HTTP gateway started", slog.String("addr", s.addr))
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.log.Info("HTTP gateway stopping")
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("gateway: serve: %w", err)
		}
		return nil
	}
}
