package grpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authv1 "github.com/Extrarius/svarog-core/api/gen/go/api/proto/auth/v1"
	"github.com/Extrarius/svarog-core/internal/app"
)

// AuthService is the transport adapter that exposes app.Handlers over gRPC.
// It translates protobuf messages into app use-case calls, maps domain errors
// to canonical gRPC status codes, and uses header metadata to drive the
// gateway's Set-Cookie behaviour (see interceptor.go for the metadata keys).
type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	handlers *app.Handlers
	log      *slog.Logger
}

// NewAuthService constructs the AuthService adapter.
func NewAuthService(handlers *app.Handlers, log *slog.Logger) *AuthService {
	return &AuthService{handlers: handlers, log: log}
}

// Register creates a new account. A duplicate email maps to AlreadyExists.
func (s *AuthService) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	out, err := s.handlers.Register(ctx, app.RegisterInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	switch {
	case errors.Is(err, app.ErrEmailTaken):
		return nil, status.Error(codes.AlreadyExists, "email already registered")
	case errors.Is(err, app.ErrInvalidCredentials):
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	case err != nil:
		return nil, status.Error(codes.Internal, "registration failed")
	}
	return &authv1.RegisterResponse{UserId: out.UserID}, nil
}

// Login validates credentials and, on success, asks the gateway to set the
// session cookie via response-header metadata.
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	out, err := s.handlers.Login(ctx, app.LoginInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: metaValue(ctx, MetaUserAgent),
		IP:        sanitizeClientIP(metaValue(ctx, MetaForwardedFor)),
	})
	switch {
	case errors.Is(err, app.ErrInvalidCredentials):
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	case err != nil:
		return nil, status.Error(codes.Internal, "login failed")
	}

	header := metadata.Pairs(
		MetaSetSession, out.Token,
		MetaSetSessionExp, out.ExpiresAt.UTC().Format(time.RFC3339),
	)
	if err := grpc.SetHeader(ctx, header); err != nil {
		s.log.WarnContext(ctx, "set login header failed", slog.Any("error", err))
	}
	return &authv1.LoginResponse{UserId: out.UserID}, nil
}

// Logout revokes the current session (resolved by the interceptor) and asks
// the gateway to clear the cookie.
func (s *AuthService) Logout(ctx context.Context, _ *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	id, _ := IdentityFromContext(ctx)
	if err := s.handlers.Logout(ctx, app.LogoutInput{SessionID: id.SessionID}); err != nil {
		return nil, status.Error(codes.Internal, "logout failed")
	}
	if err := grpc.SetHeader(ctx, metadata.Pairs(MetaClearSession, "1")); err != nil {
		s.log.WarnContext(ctx, "set logout header failed", slog.Any("error", err))
	}
	return &authv1.LogoutResponse{}, nil
}

// Me returns the identity resolved by the auth interceptor.
func (s *AuthService) Me(ctx context.Context, _ *authv1.MeRequest) (*authv1.MeResponse, error) {
	id, ok := IdentityFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no session")
	}
	return &authv1.MeResponse{UserId: id.UserID, Email: id.Email}, nil
}

// ListSessions returns active sessions for the authenticated user.
func (s *AuthService) ListSessions(ctx context.Context, _ *authv1.ListSessionsRequest) (*authv1.ListSessionsResponse, error) {
	id, ok := IdentityFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no session")
	}

	out, err := s.handlers.ListSessions(ctx, app.ListSessionsInput{
		UserID:           id.UserID,
		CurrentSessionID: id.SessionID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "list sessions failed")
	}

	sessions := make([]*authv1.SessionInfo, 0, len(out.Sessions))
	for _, item := range out.Sessions {
		sessions = append(sessions, &authv1.SessionInfo{
			Id:         item.ID,
			UserAgent:  item.UserAgent,
			Ip:         item.IP,
			CreatedAt:  item.CreatedAt.UTC().Format(time.RFC3339),
			LastSeenAt: item.LastSeenAt.UTC().Format(time.RFC3339),
			IsCurrent:  item.IsCurrent,
		})
	}
	return &authv1.ListSessionsResponse{Sessions: sessions}, nil
}

// RevokeSession revokes one session owned by the authenticated user.
func (s *AuthService) RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest) (*authv1.RevokeSessionResponse, error) {
	id, ok := IdentityFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no session")
	}

	out, err := s.handlers.RevokeSession(ctx, app.RevokeSessionInput{
		UserID:           id.UserID,
		CurrentSessionID: id.SessionID,
		SessionID:        req.GetSessionId(),
	})
	switch {
	case errors.Is(err, app.ErrSessionNotFound):
		return nil, status.Error(codes.NotFound, "session not found")
	case err != nil:
		return nil, status.Error(codes.Internal, "revoke session failed")
	}

	if out.RevokedCurrent {
		if err := grpc.SetHeader(ctx, metadata.Pairs(MetaClearSession, "1")); err != nil {
			s.log.WarnContext(ctx, "set revoke header failed", slog.Any("error", err))
		}
	}
	return &authv1.RevokeSessionResponse{}, nil
}

// RevokeAllOtherSessions revokes all sessions except the current one.
func (s *AuthService) RevokeAllOtherSessions(ctx context.Context, _ *authv1.RevokeAllOtherSessionsRequest) (*authv1.RevokeAllOtherSessionsResponse, error) {
	id, ok := IdentityFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no session")
	}

	out, err := s.handlers.RevokeAllOtherSessions(ctx, app.RevokeAllOtherSessionsInput{
		UserID:           id.UserID,
		CurrentSessionID: id.SessionID,
	})
	switch {
	case errors.Is(err, app.ErrSessionNotFound):
		return nil, status.Error(codes.Unauthenticated, "no session")
	case err != nil:
		return nil, status.Error(codes.Internal, "revoke other sessions failed")
	}

	return &authv1.RevokeAllOtherSessionsResponse{RevokedCount: int32(out.RevokedCount)}, nil
}

// sanitizeClientIP keeps the first valid IP from X-Forwarded-For style values.
func sanitizeClientIP(raw string) string {
	raw = strings.TrimSpace(strings.Split(raw, ",")[0])
	if raw == "" || net.ParseIP(raw) == nil {
		return ""
	}
	return raw
}
