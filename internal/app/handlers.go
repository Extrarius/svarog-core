package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Handlers groups the auth use cases.
//
// It is the only object that needs to be wired from the composition root
// (cmd/main.go); transport adapters depend on this struct, not on individual
// repositories or hashers.
type Handlers struct {
	users      UserRepo
	sessions   SessionRepo
	hasher     Hasher
	tokens     TokenSource
	clock      Clock
	log        *slog.Logger
	sessionTTL time.Duration
}

// Config holds the values required to construct Handlers.
type Config struct {
	Users      UserRepo
	Sessions   SessionRepo
	Hasher     Hasher
	Tokens     TokenSource
	Clock      Clock
	Logger     *slog.Logger
	SessionTTL time.Duration
}

// New constructs Handlers and validates that all required collaborators are
// wired in. It does not perform any I/O.
func New(cfg Config) (*Handlers, error) {
	if cfg.Users == nil {
		return nil, errors.New("app: UserRepo is required")
	}
	if cfg.Sessions == nil {
		return nil, errors.New("app: SessionRepo is required")
	}
	if cfg.Hasher == nil {
		return nil, errors.New("app: Hasher is required")
	}
	if cfg.Tokens == nil {
		return nil, errors.New("app: TokenSource is required")
	}
	if cfg.Clock == nil {
		return nil, errors.New("app: Clock is required")
	}
	if cfg.Logger == nil {
		return nil, errors.New("app: Logger is required")
	}
	if cfg.SessionTTL <= 0 {
		return nil, errors.New("app: SessionTTL must be positive")
	}
	return &Handlers{
		users:      cfg.Users,
		sessions:   cfg.Sessions,
		hasher:     cfg.Hasher,
		tokens:     cfg.Tokens,
		clock:      cfg.Clock,
		log:        cfg.Logger,
		sessionTTL: cfg.SessionTTL,
	}, nil
}

// RegisterInput holds inputs for the Register use case.
type RegisterInput struct {
	Email    string
	Password string
}

// RegisterOutput is the result of a successful registration.
type RegisterOutput struct {
	UserID string
}

// Register creates a new user account.
//
// Real implementation is pending — transport handlers can already call this
// method; for now it returns a sentinel error so that integration is testable.
func (h *Handlers) Register(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	return RegisterOutput{}, fmt.Errorf("Register: %w", errNotImplemented)
}

// LoginInput holds inputs for the Login use case.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IP        string
}

// LoginOutput is the result of a successful login.
type LoginOutput struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
}

// Login authenticates a user and creates a new session.
func (h *Handlers) Login(ctx context.Context, in LoginInput) (LoginOutput, error) {
	return LoginOutput{}, fmt.Errorf("Login: %w", errNotImplemented)
}

// LogoutInput identifies the session to revoke.
type LogoutInput struct {
	SessionID string
}

// Logout revokes the supplied session.
func (h *Handlers) Logout(ctx context.Context, in LogoutInput) error {
	return fmt.Errorf("Logout: %w", errNotImplemented)
}

// MeInput identifies the session to look up.
type MeInput struct {
	SessionID string
}

// MeOutput describes the currently authenticated user.
type MeOutput struct {
	UserID string
	Email  string
}

// Me resolves the user behind a session.
func (h *Handlers) Me(ctx context.Context, in MeInput) (MeOutput, error) {
	return MeOutput{}, fmt.Errorf("Me: %w", errNotImplemented)
}

// errNotImplemented marks placeholder use case bodies.
//
// Transport adapters can detect this with errors.Is and translate it to a
// codes.Unimplemented response while the skeleton is being fleshed out.
var errNotImplemented = errors.New("app: not implemented")

// ErrNotImplemented is the exported sentinel matching errNotImplemented.
var ErrNotImplemented = errNotImplemented
