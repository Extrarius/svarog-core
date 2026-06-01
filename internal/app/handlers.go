package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
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

// Register creates a new user account. It rejects an email that already
// exists with ErrEmailTaken and stores only the bcrypt hash of the password.
func (h *Handlers) Register(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return RegisterOutput{}, ErrInvalidCredentials
	}

	if _, err := h.users.FindByEmail(ctx, email); err == nil {
		return RegisterOutput{}, ErrEmailTaken
	} else if !errors.Is(err, ErrUserNotFound) {
		return RegisterOutput{}, fmt.Errorf("register: lookup email: %w", err)
	}

	hash, err := h.hasher.Hash(in.Password)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("register: hash password: %w", err)
	}

	user, err := h.users.Create(ctx, email, hash)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("register: create user: %w", err)
	}

	h.log.InfoContext(ctx, "user registered", slog.String("user_id", user.ID))
	return RegisterOutput{UserID: user.ID}, nil
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

// Login verifies credentials and, on success, creates a new session. The
// returned Token is the raw opaque token to be placed in the session cookie;
// only its hash is persisted. Invalid email or password both map to
// ErrInvalidCredentials (no user-enumeration).
func (h *Handlers) Login(ctx context.Context, in LoginInput) (LoginOutput, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return LoginOutput{}, ErrInvalidCredentials
	}

	user, err := h.users.FindByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		return LoginOutput{}, ErrInvalidCredentials
	} else if err != nil {
		return LoginOutput{}, fmt.Errorf("login: lookup user: %w", err)
	}

	if err := h.hasher.Verify(user.PasswordHash, in.Password); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return LoginOutput{}, ErrInvalidCredentials
		}
		return LoginOutput{}, fmt.Errorf("login: verify password: %w", err)
	}

	token, hash, err := h.tokens.New()
	if err != nil {
		return LoginOutput{}, fmt.Errorf("login: new token: %w", err)
	}

	now := h.clock.Now()
	expiresAt := now.Add(h.sessionTTL)
	if _, err := h.sessions.Create(ctx, Session{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
		UserAgent: in.UserAgent,
		IP:        in.IP,
	}); err != nil {
		return LoginOutput{}, fmt.Errorf("login: create session: %w", err)
	}

	h.log.InfoContext(ctx, "user logged in", slog.String("user_id", user.ID))
	return LoginOutput{UserID: user.ID, Token: token, ExpiresAt: expiresAt}, nil
}

// LogoutInput identifies the session to revoke.
type LogoutInput struct {
	SessionID string
}

// Logout revokes the supplied session. Revoking an unknown or already-revoked
// session is treated as success so that the operation is idempotent.
func (h *Handlers) Logout(ctx context.Context, in LogoutInput) error {
	if in.SessionID == "" {
		return nil
	}
	err := h.sessions.Revoke(ctx, in.SessionID, h.clock.Now())
	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("logout: revoke session: %w", err)
	}
	h.log.InfoContext(ctx, "session revoked", slog.String("session_id", in.SessionID))
	return nil
}

// MeInput carries the raw session token presented by the caller.
type MeInput struct {
	Token string
}

// MeOutput describes the currently authenticated user together with the id of
// the session that authenticated them (used by Logout).
type MeOutput struct {
	UserID    string
	Email     string
	SessionID string
}

// Me validates a raw session token and resolves the user behind it. It is the
// single authentication entry point reused by the transport interceptor to
// guard protected RPCs. Expired or unknown tokens surface ErrSessionNotFound /
// ErrSessionExpired which the transport maps to codes.Unauthenticated.
func (h *Handlers) Me(ctx context.Context, in MeInput) (MeOutput, error) {
	if in.Token == "" {
		return MeOutput{}, ErrSessionNotFound
	}

	hash := h.tokens.Hash(in.Token)
	now := h.clock.Now()

	session, err := h.sessions.FindActiveByTokenHash(ctx, hash, now)
	if errors.Is(err, ErrSessionNotFound) || errors.Is(err, ErrSessionExpired) {
		return MeOutput{}, err
	} else if err != nil {
		return MeOutput{}, fmt.Errorf("me: lookup session: %w", err)
	}

	if err := h.sessions.TouchLastSeen(ctx, session.ID, now); err != nil {
		// Best-effort: a failure to record activity must not deny access.
		h.log.WarnContext(ctx, "touch last_seen failed", slog.String("session_id", session.ID), slog.Any("error", err))
	}

	user, err := h.users.FindByID(ctx, session.UserID)
	if errors.Is(err, ErrUserNotFound) {
		return MeOutput{}, ErrSessionNotFound
	} else if err != nil {
		return MeOutput{}, fmt.Errorf("me: lookup user: %w", err)
	}

	return MeOutput{UserID: user.ID, Email: user.Email, SessionID: session.ID}, nil
}

// normalizeEmail lower-cases and trims surrounding whitespace so that lookups
// and uniqueness checks are case-insensitive regardless of client input.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
