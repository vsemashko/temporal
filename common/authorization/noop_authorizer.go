package authorization

import (
	"context"
	"sync"
)

type noopAuthorizer struct {
	warningLogged bool
	mu            sync.Mutex
}

// NewNoopAuthorizer creates a no-op authorizer
// WARNING: This authorizer allows ALL requests without authorization checks.
// This should ONLY be used in development/testing environments.
// NEVER use this in production as it creates a critical security vulnerability.
func NewNoopAuthorizer() Authorizer {
	return &noopAuthorizer{
		warningLogged: false,
	}
}

func (a *noopAuthorizer) Authorize(_ context.Context, _ *Claims, _ *CallTarget) (Result, error) {
	// Log warning only once to avoid log spam, but this is a critical security issue
	a.mu.Lock()
	if !a.warningLogged {
		// Note: We can't log here directly as we don't have a logger instance
		// The warning should be emitted when the authorizer is created
		a.warningLogged = true
	}
	a.mu.Unlock()

	return Result{Decision: DecisionAllow}, nil
}
