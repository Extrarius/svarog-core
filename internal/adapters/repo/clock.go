package repo

import "time"

// SystemClock implements app.Clock backed by the system wall clock.
type SystemClock struct{}

// NewSystemClock returns the canonical wall-clock implementation.
func NewSystemClock() SystemClock { return SystemClock{} }

// Now returns time.Now().UTC().
func (SystemClock) Now() time.Time { return time.Now().UTC() }
