// Package tracer propagates a correlation/trace ID through context so that all
// logs emitted during a single unit of work can be grouped together.
//
// Unlike an HTTP service there is no request middleware here: this bot's units of
// work are Discord interactions and scheduled sync runs, so callers start a trace
// at each entry point with NewContext (or WithTraceID when an ID already exists).
package tracer

import (
	"context"

	"github.com/google/uuid"
)

// contextKey is an unexported type for context keys so they can't collide with
// keys defined in other packages.
type contextKey string

const (
	// ContextTraceIDKey holds the trace/correlation ID.
	ContextTraceIDKey contextKey = "trace_id"
	// ContextUserIDKey holds the acting user's ID (e.g. the Discord user who ran
	// a command), when known.
	ContextUserIDKey contextKey = "user_id"
)

// NewContext returns a child context carrying a freshly generated trace ID. Call
// it once per unit of work (per interaction, per sync run).
func NewContext(ctx context.Context) context.Context {
	return WithTraceID(ctx, uuid.NewString())
}

// WithTraceID returns a child context carrying the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextTraceIDKey, traceID)
}

// WithUserID returns a child context carrying the acting user's ID.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextUserIDKey, userID)
}

// TraceIDFromContext returns the trace ID stored in ctx, or "" if absent.
func TraceIDFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(ContextTraceIDKey).(string); ok {
		return s
	}
	return ""
}

// UserIDFromContext returns the user ID stored in ctx, or "" if absent.
func UserIDFromContext(ctx context.Context) string {
	if s, ok := ctx.Value(ContextUserIDKey).(string); ok {
		return s
	}
	return ""
}
