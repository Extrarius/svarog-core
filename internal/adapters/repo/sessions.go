package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Extrarius/svarog-core/internal/app"
)

// Sessions is a pgx-backed implementation of app.SessionRepo.
type Sessions struct {
	pool *pgxpool.Pool
}

// NewSessions constructs a Sessions repository.
func NewSessions(pool *pgxpool.Pool) *Sessions {
	return &Sessions{pool: pool}
}

const sessionColumns = "id, user_id, token_hash, expires_at, created_at, last_seen_at, revoked_at, user_agent, ip"

// Create inserts a new session row.
func (r *Sessions) Create(ctx context.Context, s app.Session) (app.Session, error) {
	const q = `
		INSERT INTO sessions (user_id, token_hash, expires_at, user_agent, ip)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::inet)
		RETURNING ` + sessionColumns

	row := r.pool.QueryRow(ctx, q,
		s.UserID,
		s.TokenHash,
		s.ExpiresAt,
		s.UserAgent,
		s.IP,
	)
	return scanSession(row)
}

// FindActiveByTokenHash returns the active session matching the supplied
// token hash. Sessions that are revoked or whose expires_at is in the past
// are treated as missing.
func (r *Sessions) FindActiveByTokenHash(ctx context.Context, tokenHash []byte, now time.Time) (app.Session, error) {
	const q = `
		SELECT ` + sessionColumns + `
		FROM sessions
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2`

	row := r.pool.QueryRow(ctx, q, tokenHash, now)
	s, err := scanSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.Session{}, app.ErrSessionNotFound
	}
	return s, err
}

// TouchLastSeen updates the last_seen_at column for sliding-renewal logic.
func (r *Sessions) TouchLastSeen(ctx context.Context, sessionID string, at time.Time) error {
	const q = `UPDATE sessions SET last_seen_at = $2 WHERE id = $1`
	_, err := r.pool.Exec(ctx, q, sessionID, at)
	if err != nil {
		return fmt.Errorf("repo: touch session: %w", err)
	}
	return nil
}

// Revoke marks the session as revoked.
func (r *Sessions) Revoke(ctx context.Context, sessionID string, at time.Time) error {
	const q = `UPDATE sessions SET revoked_at = $2 WHERE id = $1 AND revoked_at IS NULL`
	_, err := r.pool.Exec(ctx, q, sessionID, at)
	if err != nil {
		return fmt.Errorf("repo: revoke session: %w", err)
	}
	return nil
}

func scanSession(row pgx.Row) (app.Session, error) {
	var (
		s         app.Session
		revokedAt *time.Time
		ip        *string
	)
	if err := row.Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.ExpiresAt,
		&s.CreatedAt,
		&s.LastSeenAt,
		&revokedAt,
		&s.UserAgent,
		&ip,
	); err != nil {
		return app.Session{}, fmt.Errorf("repo: scan session: %w", err)
	}
	s.RevokedAt = revokedAt
	if ip != nil {
		s.IP = *ip
	}
	return s, nil
}
