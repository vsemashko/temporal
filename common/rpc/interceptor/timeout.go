package interceptor

import (
	"context"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/metrics"
	"google.golang.org/grpc"
)

const (
	// DefaultTimeout is the default timeout for RPC methods when no specific timeout is configured
	DefaultTimeout = 60 * time.Second

	// MinTimeout is the minimum allowed timeout to prevent overly aggressive timeouts
	MinTimeout = 1 * time.Second

	// MaxTimeout is the maximum allowed timeout to prevent resource exhaustion
	MaxTimeout = 10 * time.Minute
)

var (
	// TimeoutExceeded is the error returned when an operation exceeds its timeout
	TimeoutExceeded = &serviceerror.ResourceExhausted{
		Cause:   enumspb.RESOURCE_EXHAUSTED_CAUSE_SYSTEM_OVERLOADED,
		Scope:   enumspb.RESOURCE_EXHAUSTED_SCOPE_SYSTEM,
		Message: "operation timeout exceeded",
	}
)

type (
	// TimeoutInterceptor enforces context timeouts for RPC methods to prevent resource exhaustion
	TimeoutInterceptor struct {
		enabled        bool
		defaultTimeout time.Duration
		methodTimeouts map[string]time.Duration
		logger         log.Logger
		metricsHandler metrics.Handler
	}
)

var _ grpc.UnaryServerInterceptor = (*TimeoutInterceptor)(nil).Intercept

// NewTimeoutInterceptor creates a new timeout enforcement interceptor
//
// Parameters:
// - enabled: Whether timeout enforcement is active (disabled by default for safety)
// - defaultTimeout: Default timeout for methods without specific timeout configured
// - methodTimeouts: Map of method name to specific timeout duration
// - logger: Logger for timeout events
// - metricsHandler: Metrics handler for timeout metrics
//
// The interceptor is designed to be fail-safe:
// - Disabled by default (operators must explicitly enable)
// - Enforces min/max timeout bounds
// - Only applies timeout if context doesn't already have a deadline
// - Logs all timeout occurrences for monitoring
func NewTimeoutInterceptor(
	enabled bool,
	defaultTimeout time.Duration,
	methodTimeouts map[string]time.Duration,
	logger log.Logger,
	metricsHandler metrics.Handler,
) *TimeoutInterceptor {
	// Validate and normalize default timeout
	if defaultTimeout < MinTimeout {
		defaultTimeout = MinTimeout
	}
	if defaultTimeout > MaxTimeout {
		defaultTimeout = MaxTimeout
	}

	// Validate and normalize method timeouts
	normalizedMethodTimeouts := make(map[string]time.Duration)
	for method, timeout := range methodTimeouts {
		if timeout < MinTimeout {
			timeout = MinTimeout
		}
		if timeout > MaxTimeout {
			timeout = MaxTimeout
		}
		normalizedMethodTimeouts[method] = timeout
	}

	return &TimeoutInterceptor{
		enabled:        enabled,
		defaultTimeout: defaultTimeout,
		methodTimeouts: normalizedMethodTimeouts,
		logger:         logger,
		metricsHandler: metricsHandler,
	}
}

// Intercept wraps the RPC handler with timeout enforcement
func (i *TimeoutInterceptor) Intercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// If disabled, pass through without modification
	if !i.enabled {
		return handler(ctx, req)
	}

	// If context already has a deadline, respect it and don't override
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return handler(ctx, req)
	}

	// Determine timeout for this method
	timeout := i.getTimeout(info.FullMethod)

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Record that we're applying a timeout
	metrics.ServiceRequestTimeoutEnforced.With(i.metricsHandler).Record(
		1,
		metrics.MethodTag(info.FullMethod),
		metrics.TimeoutTag(timeout.String()),
	)

	// Call handler with timeout context
	resp, err := handler(timeoutCtx, req)

	// Check if we hit the timeout
	if err != nil && timeoutCtx.Err() == context.DeadlineExceeded {
		// Log timeout event
		i.logger.Warn("RPC method exceeded timeout",
			tag.WorkflowHandlerName(info.FullMethod),
			tag.NewDurationTag("timeout", timeout),
			tag.Error(err),
		)

		// Record timeout metric
		metrics.ServiceRequestTimeoutExceeded.With(i.metricsHandler).Record(
			1,
			metrics.MethodTag(info.FullMethod),
			metrics.TimeoutTag(timeout.String()),
		)

		return nil, TimeoutExceeded
	}

	return resp, err
}

// getTimeout returns the timeout duration for a specific method
func (i *TimeoutInterceptor) getTimeout(method string) time.Duration {
	// Check for method-specific timeout
	if timeout, ok := i.methodTimeouts[method]; ok {
		return timeout
	}

	// Fall back to default timeout
	return i.defaultTimeout
}

// IsEnabled returns whether timeout enforcement is enabled
func (i *TimeoutInterceptor) IsEnabled() bool {
	return i.enabled
}

// GetDefaultTimeout returns the default timeout duration
func (i *TimeoutInterceptor) GetDefaultTimeout() time.Duration {
	return i.defaultTimeout
}

// GetMethodTimeout returns the timeout for a specific method, if configured
func (i *TimeoutInterceptor) GetMethodTimeout(method string) (time.Duration, bool) {
	timeout, ok := i.methodTimeouts[method]
	return timeout, ok
}
