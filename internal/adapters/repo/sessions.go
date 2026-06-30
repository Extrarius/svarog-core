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

// sessionSelectCols projects inet ip as text so pgx can scan into Go string.
const sessionSelectCols = "id, user_id, token_hash, expires_at, created_at, last_seen_at, revoked_at, user_agent, COALESCE(host(ip), '') AS ip"

// Create inserts a new session row.
func (r *Sessions) Create(ctx context.Context, s app.Session) (app.Session, error) {
	const q = `
		INSERT INTO sessions (user_id, token_hash, expires_at, user_agent, ip)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::inet)
		RETURNING ` + sessionSelectCols

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
		SELECT ` + sessionSelectCols + `
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

// ListActiveByUserID returns active sessions for a user ordered by last_seen_at desc.
func (r *Sessions) ListActiveByUserID(ctx context.Context, userID string, now time.Time) ([]app.SessionSummary, error) {
	const q = `
		SELECT id, user_agent, COALESCE(host(ip), '') AS ip, created_at, last_seen_at
		FROM sessions
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
		ORDER BY last_seen_at DESC`

	rows, err := r.pool.Query(ctx, q, userID, now)
	if err != nil {
		return nil, fmt.Errorf("repo: list sessions: %w", err)
	}
	defer rows.Close()

	var out []app.SessionSummary
	for rows.Next() {
		var s app.SessionSummary
		if err := rows.Scan(&s.ID, &s.UserAgent, &s.IP, &s.CreatedAt, &s.LastSeenAt); err != nil {
			return nil, fmt.Errorf("repo: scan session summary: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: list sessions rows: %w", err)
	}
	return out, nil
}

// RevokeOwned marks a session revoked only when it belongs to the given user.
func (r *Sessions) RevokeOwned(ctx context.Context, sessionID, userID string, at time.Time) error {
	const q = `
		UPDATE sessions
		SET revoked_at = $3
		WHERE id = $1
		  AND user_id = $2
		  AND revoked_at IS NULL`

	tag, err := r.pool.Exec(ctx, q, sessionID, userID, at)
	if err != nil {
		return fmt.Errorf("repo: revoke owned session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return app.ErrSessionNotFound
	}
	return nil
}

// RevokeAllExcept revokes all active sessions for a user except the given one.
func (r *Sessions) RevokeAllExcept(ctx context.Context, userID, exceptSessionID string, at time.Time) (int, error) {
	const q = `
		UPDATE sessions
		SET revoked_at = $3
		WHERE user_id = $1
		  AND id != $2
		  AND revoked_at IS NULL`

	tag, err := r.pool.Exec(ctx, q, userID, exceptSessionID, at)
	if err != nil {
		return 0, fmt.Errorf("repo: revoke all except: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func scanSession(row pgx.Row) (app.Session, error) {
	var (
		s         app.Session
		revokedAt *time.Time
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
		&s.IP,
	); err != nil {
		return app.Session{}, fmt.Errorf("repo: scan session: %w", err)
	}
	s.RevokedAt = revokedAt
	return s, nil
}
