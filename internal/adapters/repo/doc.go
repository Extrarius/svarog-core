// Package repo contains pgx-based implementations of the persistence ports
// declared in internal/app (UserRepo, SessionRepo) as well as auxiliary
// adapters (bcrypt Hasher, system Clock) used from the composition root.
//
// All SQL strings live in this package; nothing outside repo should embed
// SQL literals.
package repo
