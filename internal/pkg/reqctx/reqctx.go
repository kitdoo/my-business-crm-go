// Package reqctx carries the authenticated caller's user ID through a
// request's context.Context. No interceptor populates it yet — auth token
// validation (see users.Service.Login) is not wired into the gRPC server,
// per the same deferral noted for internal/rbac — so UserIDFromContext
// currently always misses. The plumbing exists so handlers that need
// CreatedBy (InventoryMovement, Sale) don't need a second pass once the
// interceptor lands.
package reqctx

import "context"

type contextKey struct{}

var userIDKey = contextKey{}

// WithUserID returns a copy of ctx carrying userID.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext returns the caller's user ID, if the context carries one.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}
