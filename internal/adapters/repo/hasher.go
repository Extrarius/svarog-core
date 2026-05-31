package repo

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Extrarius/svarog-core/internal/app"
)

// BcryptHasher hashes and verifies passwords with bcrypt.
//
// It implements app.Hasher.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher returns a BcryptHasher using the default bcrypt cost.
func NewBcryptHasher() BcryptHasher {
	return BcryptHasher{cost: bcrypt.DefaultCost}
}

// Hash returns the bcrypt hash of the supplied password.
func (h BcryptHasher) Hash(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("repo: bcrypt hash: %w", err)
	}
	return string(out), nil
}

// Verify checks the supplied password against an existing bcrypt hash.
// A failed comparison is reported as app.ErrInvalidCredentials so the
// transport layer can map it to codes.Unauthenticated.
func (h BcryptHasher) Verify(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return app.ErrInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("repo: bcrypt verify: %w", err)
	}
	return nil
}
