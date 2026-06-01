// Package gateway hosts the HTTP transport adapter built on grpc-gateway.
//
// The Server type owns an *http.Server that fronts the gRPC server via a
// runtime.ServeMux. It also bridges session cookies to gRPC metadata: the
// incoming session cookie is forwarded to the gRPC server, and Set-Cookie /
// clear-cookie instructions returned as response-header metadata are turned
// into real Set-Cookie headers.
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	authv1 "github.com/Extrarius/svarog-core/api/gen/go/api/proto/auth/v1"
	grpcsrv "github.com/Extrarius/svarog-core/internal/api/grpc"
	"github.com/Extrarius/svarog-core/internal/auth"
)

// Server bundles the HTTP listener, the upstream gRPC connection and lifecycle
// helpers.
type Server struct {
	addr   string
	log    *slog.Logger
	server *http.Server
	conn   *grpc.ClientConn
}

// Options describes how the gateway server should be constructed.
type Options struct {
	Addr       string
	Logger     *slog.Logger
	GRPCTarget string // host:port the gateway forwards to (typically the grpc.Server addr)

	CookieName   string
	CookieDomain string
	CookieSecure bool
}

// New constructs a Server, dialing the upstream gRPC server and registering the
// AuthService handler on the mux.
func New(opts Options) (*Server, error) {
	if opts.Logger == nil {
		return nil, fmt.Errorf("gateway: Logger is required")
	}
	if opts.GRPCTarget == "" {
		return nil, fmt.Errorf("gateway: GRPCTarget is required")
	}

	cookieOpts := auth.CookieOptions{
		Name:     opts.CookieName,
		Domain:   opts.CookieDomain,
		Path:     "/",
		Secure:   opts.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	}

	mux := runtime.NewServeMux(
		runtime.WithMetadata(sessionMetadata(cookieOpts.Name)),
		runtime.WithForwardResponseOption(cookieForwarder(cookieOpts)),
	)

	conn, err := grpc.NewClient(opts.GRPCTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("gateway: dial grpc %q: %w", opts.GRPCTarget, err)
	}
	if err := authv1.RegisterAuthServiceHandler(context.Background(), mux, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("gateway: register auth handler: %w", err)
	}

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
		conn:   conn,
	}, nil
}

// sessionMetadata copies the raw session token from the cookie (and the client
// IP) into outgoing gRPC metadata so the auth interceptor can validate it.
func sessionMetadata(cookieName string) func(context.Context, *http.Request) metadata.MD {
	return func(_ context.Context, r *http.Request) metadata.MD {
		pairs := make([]string, 0, 4)
		if token := auth.FromRequest(r, cookieName); token != "" {
			pairs = append(pairs, grpcsrv.MetaSessionToken, token)
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			pairs = append(pairs, grpcsrv.MetaForwardedFor, xff)
		}
		if len(pairs) == 0 {
			return metadata.MD{}
		}
		return metadata.Pairs(pairs...)
	}
}

// cookieForwarder turns the AuthService's response-header metadata into real
// Set-Cookie headers, then strips the bridge metadata so it does not leak as
// Grpc-Metadata-* response headers.
func cookieForwarder(opts auth.CookieOptions) func(context.Context, http.ResponseWriter, proto.Message) error {
	return func(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
		md, ok := runtime.ServerMetadataFromContext(ctx)
		if !ok {
			return nil
		}

		if tokens := md.HeaderMD.Get(grpcsrv.MetaSetSession); len(tokens) > 0 && tokens[0] != "" {
			var expiresAt time.Time
			if exps := md.HeaderMD.Get(grpcsrv.MetaSetSessionExp); len(exps) > 0 {
				if t, err := time.Parse(time.RFC3339, exps[0]); err == nil {
					expiresAt = t
				}
			}
			http.SetCookie(w, opts.NewCookie(tokens[0], expiresAt))
		}

		if clear := md.HeaderMD.Get(grpcsrv.MetaClearSession); len(clear) > 0 && clear[0] == "1" {
			http.SetCookie(w, opts.ClearCookie())
		}

		for _, k := range []string{grpcsrv.MetaSetSession, grpcsrv.MetaSetSessionExp, grpcsrv.MetaClearSession} {
			w.Header().Del("Grpc-Metadata-" + http.CanonicalHeaderKey(k))
		}
		return nil
	}
}

// Run serves HTTP until ctx is cancelled. It uses Shutdown on cancellation for
// a graceful drain and closes the upstream gRPC connection.
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
		err := s.server.Shutdown(shutdownCtx)
		_ = s.conn.Close()
		return err
	case err := <-errCh:
		_ = s.conn.Close()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("gateway: serve: %w", err)
		}
		return nil
	}
}
