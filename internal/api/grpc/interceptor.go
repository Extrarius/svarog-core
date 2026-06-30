package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Extrarius/svarog-core/internal/app"
)

// Metadata keys bridge the HTTP gateway and the gRPC server. The gateway reads
// the session cookie and forwards the raw token as MetaSessionToken; the
// AuthService answers with MetaSetSession / MetaClearSession so the gateway can
// emit the appropriate Set-Cookie header.
const (
	MetaSessionToken  = "x-session-token"
	MetaSetSession    = "x-set-session"
	MetaSetSessionExp = "x-set-session-exp"
	MetaClearSession  = "x-clear-session"
	MetaUserAgent     = "grpcgateway-user-agent"
	MetaForwardedFor  = "x-forwarded-for"
)

type identityKey struct{}

func withIdentity(ctx context.Context, id app.MeOutput) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// IdentityFromContext returns the identity injected by the auth interceptor for
// protected methods. ok is false on public methods.
func IdentityFromContext(ctx context.Context) (app.MeOutput, bool) {
	id, ok := ctx.Value(identityKey{}).(app.MeOutput)
	return id, ok
}

// protectedMethods lists fully-qualified gRPC methods that require a valid
// session. Register and Login are intentionally public.
var protectedMethods = map[string]bool{
	"/auth.v1.AuthService/Logout":        true,
	"/auth.v1.AuthService/Me":            true,
	"/auth.v1.AuthService/ListSessions":             true,
	"/auth.v1.AuthService/RevokeSession":            true,
	"/auth.v1.AuthService/RevokeAllOtherSessions":   true,
}

// authUnaryInterceptor validates the session token for protected methods and
// injects the resolved identity into the downstream context.
func authUnaryInterceptor(handlers *app.Handlers) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !protectedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		id, err := handlers.Me(ctx, app.MeInput{Token: metaValue(ctx, MetaSessionToken)})
		switch {
		case errors.Is(err, app.ErrSessionNotFound), errors.Is(err, app.ErrSessionExpired):
			return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
		case err != nil:
			return nil, status.Error(codes.Internal, "authentication failed")
		}

		return handler(withIdentity(ctx, id), req)
	}
}

// metaValue returns the first value of an incoming metadata key, or "".
func metaValue(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
