// Package logger constructs the application's slog.Logger.
//
// The logger is wired with the OTel-aware otelslog bridge so every log record
// is enriched with the current span's trace_id / span_id, and shipped to the
// OTel Collector as part of the logs pipeline (the bridge is enabled once
// internal/observability has initialised the global LoggerProvider).
package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// Options describes how the application logger should be built.
type Options struct {
	Level       string // debug | info | warn | error
	Format      string // json | text
	ServiceName string // attached to every record as "service.name"
	Writer      io.Writer
}

// New creates a slog.Logger that combines a stdlib JSON / text handler
// (writing to Options.Writer or stderr) with the otelslog handler that
// forwards records to the OTel logs pipeline.
//
// The resulting logger is safe to use as slog.Default().
func New(opts Options) *slog.Logger {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}
	level := parseLevel(opts.Level)

	var base slog.Handler
	switch strings.ToLower(opts.Format) {
	case "text":
		base = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	default:
		base = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	}
	if opts.ServiceName != "" {
		base = base.WithAttrs([]slog.Attr{slog.String("service.name", opts.ServiceName)})
	}

	// otelslog forwards records into the OTel logs pipeline. It is a no-op
	// until observability.Init wires a global LoggerProvider, which keeps
	// startup ordering forgiving.
	otelHandler := otelslog.NewHandler(opts.ServiceName)

	return slog.New(fanout{base, otelHandler})
}

// fanout dispatches log records to multiple handlers.
type fanout []slog.Handler

func (f fanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f fanout) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range f {
		if err := h.Handle(ctx, r.Clone()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (f fanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make(fanout, len(f))
	for i, h := range f {
		out[i] = h.WithAttrs(attrs)
	}
	return out
}

func (f fanout) WithGroup(name string) slog.Handler {
	out := make(fanout, len(f))
	for i, h := range f {
		out[i] = h.WithGroup(name)
	}
	return out
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
