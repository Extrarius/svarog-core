// Package grpc owns the gRPC transport adapter for svarog-core.
//
// The Server type assembles a *google.golang.org/grpc.Server with OTel and
// recovery interceptors, exposes gRPC reflection and health endpoints, and
// registers the generated AuthService and session auth interceptor.
package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	authv1 "github.com/Extrarius/svarog-core/api/gen/go/api/proto/auth/v1"
	"github.com/Extrarius/svarog-core/internal/app"
)

// Server bundles a gRPC server, its listener and lifecycle helpers.
type Server struct {
	addr     string
	log      *slog.Logger
	grpcSrv  *grpc.Server
	listener net.Listener
}

// Options describes how the gRPC server should be constructed.
type Options struct {
	Addr     string
	Logger   *slog.Logger
	Handlers *app.Handlers
}

// New constructs a Server. It binds the TCP listener immediately so that
// startup errors surface at composition time, not from a background goroutine.
func New(opts Options) (*Server, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("grpc: Logger is required")
	}
	if opts.Handlers == nil {
		return nil, fmt.Errorf("grpc: Handlers are required")
	}
	lis, err := net.Listen("tcp", opts.Addr)
	if err != nil {
		return nil, fmt.Errorf("grpc: listen %q: %w", opts.Addr, err)
	}

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(authUnaryInterceptor(opts.Handlers)),
	)

	// Standard endpoints: reflection (for grpcurl) + health (for orchestrators).
	reflection.Register(srv)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	authv1.RegisterAuthServiceServer(srv, NewAuthService(opts.Handlers, opts.Logger))

	return &Server{
		addr:     opts.Addr,
		log:      opts.Logger,
		grpcSrv:  srv,
		listener: lis,
	}, nil
}

// Addr returns the bound listen address (useful with :0 in tests).
func (s *Server) Addr() string { return s.listener.Addr().String() }

// Run serves gRPC until ctx is cancelled or Serve returns an error.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("gRPC server started", slog.String("addr", s.addr))
		errCh <- s.grpcSrv.Serve(s.listener)
	}()

	select {
	case <-ctx.Done():
		s.log.Info("gRPC server stopping")
		s.grpcSrv.GracefulStop()
		return nil
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("grpc: serve: %w", err)
		}
		return nil
	}
}

// Server exposes the underlying *grpc.Server for advanced wiring such as
// registering additional services from cmd/main.go once they exist.
func (s *Server) Server() *grpc.Server { return s.grpcSrv }
