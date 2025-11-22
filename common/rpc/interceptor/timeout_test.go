package interceptor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"google.golang.org/grpc"
)

func TestNewTimeoutInterceptor_DefaultValues(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	interceptor := NewTimeoutInterceptor(
		false,
		DefaultTimeout,
		nil,
		logger,
		metricsHandler,
	)

	assert.False(t, interceptor.IsEnabled())
	assert.Equal(t, DefaultTimeout, interceptor.GetDefaultTimeout())
}

func TestNewTimeoutInterceptor_NormalizesTimeouts(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	tests := []struct {
		name            string
		defaultTimeout  time.Duration
		expectedDefault time.Duration
	}{
		{
			name:            "too small timeout normalized to min",
			defaultTimeout:  100 * time.Millisecond,
			expectedDefault: MinTimeout,
		},
		{
			name:            "too large timeout normalized to max",
			defaultTimeout:  20 * time.Minute,
			expectedDefault: MaxTimeout,
		},
		{
			name:            "valid timeout unchanged",
			defaultTimeout:  30 * time.Second,
			expectedDefault: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewTimeoutInterceptor(
				true,
				tt.defaultTimeout,
				nil,
				logger,
				metricsHandler,
			)

			assert.Equal(t, tt.expectedDefault, interceptor.GetDefaultTimeout())
		})
	}
}

func TestNewTimeoutInterceptor_NormalizesMethodTimeouts(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	methodTimeouts := map[string]time.Duration{
		"/valid":    5 * time.Second,
		"/tooSmall": 100 * time.Millisecond,
		"/tooLarge": 20 * time.Minute,
	}

	interceptor := NewTimeoutInterceptor(
		true,
		DefaultTimeout,
		methodTimeouts,
		logger,
		metricsHandler,
	)

	timeout, ok := interceptor.GetMethodTimeout("/valid")
	assert.True(t, ok)
	assert.Equal(t, 5*time.Second, timeout)

	timeout, ok = interceptor.GetMethodTimeout("/tooSmall")
	assert.True(t, ok)
	assert.Equal(t, MinTimeout, timeout)

	timeout, ok = interceptor.GetMethodTimeout("/tooLarge")
	assert.True(t, ok)
	assert.Equal(t, MaxTimeout, timeout)

	_, ok = interceptor.GetMethodTimeout("/nonexistent")
	assert.False(t, ok)
}

func TestTimeoutInterceptor_DisabledPassesThrough(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	interceptor := NewTimeoutInterceptor(
		false, // disabled
		5*time.Second,
		nil,
		logger,
		metricsHandler,
	)

	handlerCalled := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		// Verify context doesn't have a deadline since interceptor is disabled
		_, hasDeadline := ctx.Deadline()
		assert.False(t, hasDeadline, "Context should not have deadline when interceptor disabled")
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	resp, err := interceptor.Intercept(context.Background(), "request", info, handler)

	assert.True(t, handlerCalled)
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTimeoutInterceptor_RespectsExistingDeadline(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	interceptor := NewTimeoutInterceptor(
		true,
		5*time.Second,
		nil,
		logger,
		metricsHandler,
	)

	// Create context with existing deadline
	existingDeadline := time.Now().Add(10 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), existingDeadline)
	defer cancel()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		deadline, hasDeadline := ctx.Deadline()
		assert.True(t, hasDeadline, "Context should have deadline")
		// Deadline should be the original one, not modified
		assert.Equal(t, existingDeadline, deadline)
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	resp, err := interceptor.Intercept(ctx, "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTimeoutInterceptor_AppliesDefaultTimeout(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	defaultTimeout := 5 * time.Second
	interceptor := NewTimeoutInterceptor(
		true,
		defaultTimeout,
		nil,
		logger,
		metricsHandler,
	)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		deadline, hasDeadline := ctx.Deadline()
		assert.True(t, hasDeadline, "Context should have deadline")
		// Deadline should be approximately now + defaultTimeout
		expectedDeadline := time.Now().Add(defaultTimeout)
		assert.WithinDuration(t, expectedDeadline, deadline, 100*time.Millisecond)
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	resp, err := interceptor.Intercept(context.Background(), "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTimeoutInterceptor_AppliesMethodSpecificTimeout(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	methodTimeout := 2 * time.Second
	methodTimeouts := map[string]time.Duration{
		"/test.Service/SpecialMethod": methodTimeout,
	}

	interceptor := NewTimeoutInterceptor(
		true,
		10*time.Second, // default is different
		methodTimeouts,
		logger,
		metricsHandler,
	)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		deadline, hasDeadline := ctx.Deadline()
		assert.True(t, hasDeadline)
		// Should use method-specific timeout, not default
		expectedDeadline := time.Now().Add(methodTimeout)
		assert.WithinDuration(t, expectedDeadline, deadline, 100*time.Millisecond)
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SpecialMethod",
	}

	resp, err := interceptor.Intercept(context.Background(), "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestTimeoutInterceptor_HandlesTimeout(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	// Very short timeout to trigger timeout
	shortTimeout := 10 * time.Millisecond
	interceptor := NewTimeoutInterceptor(
		true,
		shortTimeout,
		nil,
		logger,
		metricsHandler,
	)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Sleep longer than timeout
		time.Sleep(100 * time.Millisecond)
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/SlowMethod",
	}

	resp, err := interceptor.Intercept(context.Background(), "request", info, handler)

	// Should return timeout error
	assert.Error(t, err)
	assert.Equal(t, TimeoutExceeded, err)
	assert.Nil(t, resp)
}

func TestTimeoutInterceptor_PropagatesHandlerError(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	interceptor := NewTimeoutInterceptor(
		true,
		5*time.Second,
		nil,
		logger,
		metricsHandler,
	)

	expectedError := errors.New("handler error")
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedError
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	resp, err := interceptor.Intercept(context.Background(), "request", info, handler)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, resp)
}

func TestTimeoutInterceptor_SuccessfulRequest(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	interceptor := NewTimeoutInterceptor(
		true,
		5*time.Second,
		nil,
		logger,
		metricsHandler,
	)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Fast handler that completes before timeout
		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/FastMethod",
	}

	resp, err := interceptor.Intercept(context.Background(), "request", info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
}

func TestTimeoutInterceptor_GetTimeout(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	methodTimeouts := map[string]time.Duration{
		"/configured/method": 3 * time.Second,
	}

	interceptor := NewTimeoutInterceptor(
		true,
		DefaultTimeout,
		methodTimeouts,
		logger,
		metricsHandler,
	)

	// Method with specific timeout
	timeout := interceptor.getTimeout("/configured/method")
	assert.Equal(t, 3*time.Second, timeout)

	// Method without specific timeout uses default
	timeout = interceptor.getTimeout("/unconfigured/method")
	assert.Equal(t, DefaultTimeout, timeout)
}

func TestTimeoutInterceptor_CancellationPropagates(t *testing.T) {
	logger := log.NewNoopLogger()
	metricsHandler := metrics.NoopMetricsHandler

	interceptor := NewTimeoutInterceptor(
		true,
		5*time.Second,
		nil,
		logger,
		metricsHandler,
	)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	handlerStarted := make(chan struct{})
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		close(handlerStarted)
		// Wait for cancellation
		<-ctx.Done()
		return nil, ctx.Err()
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Start interceptor in goroutine
	done := make(chan struct{})
	var resp interface{}
	var err error
	go func() {
		resp, err = interceptor.Intercept(ctx, "request", info, handler)
		close(done)
	}()

	// Wait for handler to start, then cancel
	<-handlerStarted
	cancel()

	// Wait for completion
	<-done

	// Should propagate cancellation error
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, resp)
}

func TestTimeoutInterceptor_Constants(t *testing.T) {
	// Verify timeout constants have sensible values
	assert.Equal(t, 60*time.Second, DefaultTimeout)
	assert.Equal(t, 1*time.Second, MinTimeout)
	assert.Equal(t, 10*time.Minute, MaxTimeout)

	// Verify min < default < max
	assert.True(t, MinTimeout < DefaultTimeout)
	assert.True(t, DefaultTimeout < MaxTimeout)
}

func TestTimeoutExceeded_Error(t *testing.T) {
	// Verify TimeoutExceeded error has correct properties
	require.NotNil(t, TimeoutExceeded)
	assert.Contains(t, TimeoutExceeded.Error(), "timeout")
}
