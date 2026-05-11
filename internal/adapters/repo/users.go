package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Extrarius/svarog-core/internal/app"
)

// Users is a pgx-backed implementation of app.UserRepo.
type Users struct {
	pool *pgxpool.Pool
}

// NewUsers constructs a Users repository.
func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{pool: pool}
}

const userColumns = "id, email, password_hash, created_at, updated_at"

// Create inserts a new user.
func (r *Users) Create(ctx context.Context, email, passwordHash string) (app.User, error) {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING ` + userColumns

	row := r.pool.QueryRow(ctx, q, email, passwordHash)
	return scanUser(row)
}

// FindByEmail returns the user matching the supplied email, or
// app.ErrUserNotFound if no row exists.
func (r *Users) FindByEmail(ctx context.Context, email string) (app.User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE email = $1`
	row := r.pool.QueryRow(ctx, q, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.User{}, app.ErrUserNotFound
	}
	return u, err
}

// FindByID returns the user with the supplied id.
func (r *Users) FindByID(ctx context.Context, id string) (app.User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	row := r.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.User{}, app.ErrUserNotFound
	}
	return u, err
}

func scanUser(row pgx.Row) (app.User, error) {
	var u app.User
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return app.User{}, fmt.Errorf("repo: scan user: %w", err)
	}
	return u, nil
}
