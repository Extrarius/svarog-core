// Package app contains the business logic of svarog-core.
//
// IMPORTANT: this package MUST only import the Go standard library and
// other internal/* packages that are themselves stdlib-only. All external
// dependencies (database drivers, gRPC, OpenTelemetry, etc.) live in
// adapters and are injected through the interfaces declared in this file.
package app

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors surfaced by use cases. Transport adapters map them to
// canonical gRPC codes; the app layer never references gRPC directly.
var (
	ErrEmailTaken         = errors.New("app: email is already taken")
	ErrInvalidCredentials = errors.New("app: invalid credentials")
	ErrSessionNotFound    = errors.New("app: session not found")
	ErrSessionExpired     = errors.New("app: session expired")
	ErrUserNotFound       = errors.New("app: user not found")
)

// User is the canonical user representation inside the app layer.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session represents a persisted login session.
//
// The plain token never lives in the database; only TokenHash (SHA-256 of the
// raw token bytes) is stored. The raw token is returned to the caller only
// when a session is freshly created.
type Session struct {
	ID          string
	UserID      string
	TokenHash   []byte
	ExpiresAt   time.Time
	CreatedAt   time.Time
	LastSeenAt  time.Time
	RevokedAt   *time.Time
	UserAgent   string
	IP          string
}

// UserRepo abstracts persistence of user accounts.
type UserRepo interface {
	Create(ctx context.Context, email, passwordHash string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
}

// SessionRepo abstracts persistence of sessions.
type SessionRepo interface {
	Create(ctx context.Context, s Session) (Session, error)
	FindActiveByTokenHash(ctx context.Context, tokenHash []byte, now time.Time) (Session, error)
	TouchLastSeen(ctx context.Context, sessionID string, at time.Time) error
	Revoke(ctx context.Context, sessionID string, at time.Time) error
}

// Hasher hashes and verifies passwords (typically bcrypt in production).
type Hasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

// Clock abstracts time.Now to keep use cases deterministic in tests.
type Clock interface {
	Now() time.Time
}

// TokenSource generates opaque session tokens and their SHA-256 hashes.
type TokenSource interface {
	New() (token string, hash []byte, err error)
	Hash(token string) []byte
}
