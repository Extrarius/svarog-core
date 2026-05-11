package grpc

import (
	"log/slog"

	"github.com/Extrarius/svarog-core/internal/app"
)

// AuthService is the transport adapter that exposes app.Handlers over gRPC.
//
// Methods are intentionally not implemented yet: the generated AuthServiceServer
// interface lives in api/gen/go/auth/v1 and only appears after `make proto-gen`
// runs. Once it exists, this file should:
//
//  1. embed `authv1.UnimplementedAuthServiceServer`, and
//  2. provide concrete Register / Login / Logout / Me methods that translate
//     protobuf messages into app.Register/Login/Logout/Me calls and back.
//
// The struct already accepts the dependencies it needs so wiring in
// cmd/main.go does not have to change at that point.
type AuthService struct {
	handlers *app.Handlers
	log      *slog.Logger
}

// NewAuthService constructs the placeholder AuthService.
func NewAuthService(handlers *app.Handlers, log *slog.Logger) *AuthService {
	return &AuthService{handlers: handlers, log: log}
}
