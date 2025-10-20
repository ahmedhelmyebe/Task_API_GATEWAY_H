package util // Small helpers for request context values

import (
	"context"
)

// Keys stored in context to avoid magic strings
const (
	CtxKeyRequestID = "req.id"
	CtxKeyUserID    = "auth.sub"
	CtxKeyUserRole  = "auth.role"
)

// With sets k=v in a context and returns child context.
func With(ctx context.Context, k, v any) context.Context { return context.WithValue(ctx, k, v) }

// Get pulls a value from context.
func Get[T any](ctx context.Context, k any) (T, bool) { // generic getter
	v, ok := ctx.Value(k).(T)
	return v, ok
}