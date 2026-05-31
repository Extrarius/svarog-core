// Package auth provides primitives for opaque session tokens.
//
// The package is intentionally stdlib-only so that internal/app may depend on
// it without breaking the Clean Architecture rule that forbids non-stdlib
// imports inside the business layer.
//
// Tokens are 32 cryptographically-random bytes encoded with URL-safe base64.
// Only the SHA-256 hash of the raw token is persisted; the raw token lives
// only inside an HttpOnly cookie sent to the client.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

// TokenByteLen is the number of random bytes drawn for each session token.
const TokenByteLen = 32

// DefaultCookieName is the cookie name used by Cookie helpers when no other
// name is provided.
const DefaultCookieName = "sid"

// TokenSource generates session tokens and computes their hashes.
type TokenSource struct{}

// NewTokenSource returns a stateless TokenSource.
func NewTokenSource() TokenSource { return TokenSource{} }

// New generates a fresh random token along with its SHA-256 hash.
func (TokenSource) New() (token string, hash []byte, err error) {
	buf := make([]byte, TokenByteLen)
	if _, err = rand.Read(buf); err != nil {
		return "", nil, fmt.Errorf("auth: generate token: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	h := sha256.Sum256([]byte(token))
	return token, h[:], nil
}

// Hash returns the SHA-256 hash of an already-encoded token string.
func (TokenSource) Hash(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// CookieOptions captures the runtime configuration of the session cookie.
type CookieOptions struct {
	Name     string
	Domain   string
	Path     string
	Secure   bool
	SameSite http.SameSite
}

// NewCookie builds the Set-Cookie value used when a session is created.
// expiresAt is taken into account so that browsers expire the cookie in sync
// with the database record.
func (o CookieOptions) NewCookie(token string, expiresAt time.Time) *http.Cookie {
	name := o.Name
	if name == "" {
		name = DefaultCookieName
	}
	path := o.Path
	if path == "" {
		path = "/"
	}
	return &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     path,
		Domain:   o.Domain,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   o.Secure,
		SameSite: o.SameSite,
	}
}

// ClearCookie produces a Set-Cookie value that instructs the browser to drop
// the session cookie immediately (used on logout).
func (o CookieOptions) ClearCookie() *http.Cookie {
	name := o.Name
	if name == "" {
		name = DefaultCookieName
	}
	path := o.Path
	if path == "" {
		path = "/"
	}
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		Domain:   o.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   o.Secure,
		SameSite: o.SameSite,
	}
}

// FromRequest extracts the raw token from the named cookie on an http.Request.
// It returns an empty string when the cookie is missing.
func FromRequest(r *http.Request, cookieName string) string {
	if cookieName == "" {
		cookieName = DefaultCookieName
	}
	c, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
