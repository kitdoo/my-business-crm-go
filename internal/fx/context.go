package fx

import (
	"context"
	"time"
)

// defaultInitTimeout caps initialization work so a hung dependency does not
// block startup forever.
const defaultInitTimeout = 30 * time.Second

// initContext creates a context for initialization operations. The caller
// MUST defer the returned cancel.
func initContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultInitTimeout)
}

// lifecycleContext returns a context safe for use inside Fx StartHook /
// StopHook callbacks. If the provided context is nil or already canceled
// it returns a fresh context with a graceful timeout.
func lifecycleContext(ctx context.Context) context.Context {
	if ctx == nil {
		newCtx, cancel := context.WithTimeout(context.Background(), defaultInitTimeout)
		_ = cancel
		return newCtx
	}

	select {
	case <-ctx.Done():
		const gracefulTimeout = 10 * time.Second
		newCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		_ = cancel
		return newCtx
	default:
		return ctx
	}
}
