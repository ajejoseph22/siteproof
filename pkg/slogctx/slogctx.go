// Package slogctx provides context-aware structured logging with request tracing.
package slogctx

import (
	"context"
	"log/slog"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const loggerKey contextKey = "slog_logger"

// WithLogger returns a new context with the logger attached.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the logger from context, or the default logger if not present.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// With returns a new logger with the given attributes attached.
// This is a convenience wrapper around slog.Logger.With.
func With(ctx context.Context, args ...any) *slog.Logger {
	return FromContext(ctx).With(args...)
}

// Debug logs at debug level using the logger from context.
func Debug(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).DebugContext(ctx, msg, args...)
}

// Info logs at info level using the logger from context.
func Info(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).InfoContext(ctx, msg, args...)
}

// Warn logs at warn level using the logger from context.
func Warn(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).WarnContext(ctx, msg, args...)
}

// Error logs at error level using the logger from context.
func Error(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).ErrorContext(ctx, msg, args...)
}
