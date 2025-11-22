package interceptor

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type (
	authRateLimitInterceptorSuite struct {
		suite.Suite
		*require.Assertions

		logger         log.Logger
		metricsHandler metrics.Handler
	}
)

func TestAuthRateLimitInterceptorSuite(t *testing.T) {
	suite.Run(t, &authRateLimitInterceptorSuite{})
}

func (s *authRateLimitInterceptorSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.logger = log.NewTestLogger()
	s.metricsHandler = metrics.NoopMetricsHandler
}

func (s *authRateLimitInterceptorSuite) TestDisabledRateLimiting() {
	cfg := config.AuthRateLimit{
		Enabled: false,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	// Should always allow requests when disabled
	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	ctx := s.createContextWithIP("192.168.1.1")
	resp, err := interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	s.NoError(err)
	s.Equal("response", resp)
}

func (s *authRateLimitInterceptorSuite) TestAuthFailureTracking() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 5,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	// Handler that returns auth error
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewPermissionDenied("auth failed", "")
	}

	ctx := s.createContextWithIP("192.168.1.1")

	// Make 3 failed attempts - should not be locked out yet
	for i := 0; i < 3; i++ {
		resp, err := interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
		s.Nil(resp)
		s.Error(err)
		s.IsType(&serviceerror.PermissionDenied{}, err)
	}

	// Verify the IP is not locked out yet
	s.False(interceptor.isLockedOut("192.168.1.1"))
	s.Equal(3, interceptor.trackers["192.168.1.1"].failures)
}

func (s *authRateLimitInterceptorSuite) TestLockoutAfterThresholdExceeded() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 3,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	// Handler that returns auth error
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewPermissionDenied("auth failed", "")
	}

	ctx := s.createContextWithIP("192.168.1.1")

	// Make 3 failed attempts to trigger lockout
	for i := 0; i < 3; i++ {
		_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	}

	// Verify IP is locked out
	s.True(interceptor.isLockedOut("192.168.1.1"))

	// Next request should be blocked before reaching handler
	handlerCalled := false
	successHandler := func(ctx context.Context, req any) (any, error) {
		handlerCalled = true
		return "success", nil
	}

	resp, err := interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, successHandler)
	s.Nil(resp)
	s.Error(err)
	s.Equal(errAuthRateLimited, err)
	s.False(handlerCalled, "Handler should not be called when IP is locked out")
}

func (s *authRateLimitInterceptorSuite) TestClearFailuresOnSuccess() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 5,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	ctx := s.createContextWithIP("192.168.1.1")

	// Make 2 failed attempts
	failHandler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewPermissionDenied("auth failed", "")
	}
	for i := 0; i < 2; i++ {
		_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, failHandler)
	}

	// Verify failures are tracked
	s.Equal(2, interceptor.trackers["192.168.1.1"].failures)

	// Make successful request
	successHandler := func(ctx context.Context, req any) (any, error) {
		return "success", nil
	}
	resp, err := interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, successHandler)
	s.NoError(err)
	s.Equal("success", resp)

	// Verify failures are cleared
	_, exists := interceptor.trackers["192.168.1.1"]
	s.False(exists, "Tracker should be removed after successful auth")
}

func (s *authRateLimitInterceptorSuite) TestSlidingWindow() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 3,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	ctx := s.createContextWithIP("192.168.1.1")
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewPermissionDenied("auth failed", "")
	}

	// Make 2 failed attempts
	for i := 0; i < 2; i++ {
		_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	}

	// Manually set the first failure time to >1 minute ago to trigger window reset
	interceptor.mu.Lock()
	tracker := interceptor.trackers["192.168.1.1"]
	tracker.firstFailure = time.Now().Add(-65 * time.Second)
	interceptor.mu.Unlock()

	// Next failure should reset the window
	_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	// Verify window was reset (failures should be 1, not 3)
	s.Equal(1, interceptor.trackers["192.168.1.1"].failures)
	s.False(interceptor.isLockedOut("192.168.1.1"))
}

func (s *authRateLimitInterceptorSuite) TestMaxTrackedIPs() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 5,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewPermissionDenied("auth failed", "")
	}

	// Fill up to max tracked IPs (default is 10000)
	// For testing purposes, we'll manually set a lower limit
	interceptor.mu.Lock()
	interceptor.config.maxTrackedIPs = 5
	interceptor.mu.Unlock()

	// Track 5 IPs
	for i := 1; i <= 5; i++ {
		ctx := s.createContextWithIP(s.makeIP(i))
		_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	}

	// Verify we have 5 tracked IPs
	s.Equal(5, len(interceptor.trackers))

	// Try to track a 6th IP - should be rejected
	ctx := s.createContextWithIP(s.makeIP(6))
	_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	// Should still have only 5 tracked IPs
	s.Equal(5, len(interceptor.trackers))
	_, exists := interceptor.trackers[s.makeIP(6)]
	s.False(exists, "Should not track new IP when limit reached")
}

func (s *authRateLimitInterceptorSuite) TestCleanup() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 5,
		LockoutDuration:      1 * time.Second,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewPermissionDenied("auth failed", "")
	}

	// Create a failed attempt
	ctx := s.createContextWithIP("192.168.1.1")
	_, _ = interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)

	// Verify tracker exists
	s.Equal(1, len(interceptor.trackers))

	// Manually set last failure to old time
	interceptor.mu.Lock()
	tracker := interceptor.trackers["192.168.1.1"]
	tracker.lastFailure = time.Now().Add(-20 * time.Minute)
	tracker.lockedUntil = time.Now().Add(-10 * time.Minute)
	interceptor.mu.Unlock()

	// Run cleanup
	interceptor.cleanup()

	// Verify tracker was removed
	s.Equal(0, len(interceptor.trackers))
}

func (s *authRateLimitInterceptorSuite) TestNoIPInContext() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 3,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	// Handler should be called even without IP (fail-open behavior)
	handlerCalled := false
	handler := func(ctx context.Context, req any) (any, error) {
		handlerCalled = true
		return "response", nil
	}

	// Create context without peer info
	ctx := context.Background()
	resp, err := interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	s.NoError(err)
	s.Equal("response", resp)
	s.True(handlerCalled, "Handler should be called when IP cannot be extracted (fail-open)")
}

func (s *authRateLimitInterceptorSuite) TestNonAuthErrors() {
	cfg := config.AuthRateLimit{
		Enabled:              true,
		MaxFailuresPerMinute: 3,
		LockoutDuration:      5 * time.Minute,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	ctx := s.createContextWithIP("192.168.1.1")

	// Handler that returns non-auth error
	handler := func(ctx context.Context, req any) (any, error) {
		return nil, serviceerror.NewInternal("internal error")
	}

	resp, err := interceptor.Intercept(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	s.Nil(resp)
	s.Error(err)

	// Non-auth errors should not be tracked
	_, exists := interceptor.trackers["192.168.1.1"]
	s.False(exists, "Non-auth errors should not be tracked")
}

func (s *authRateLimitInterceptorSuite) TestDefaultConfiguration() {
	// Test with minimal config (should apply defaults)
	cfg := config.AuthRateLimit{
		Enabled: true,
	}
	interceptor := NewAuthRateLimitInterceptor(cfg, s.logger, s.metricsHandler)
	defer interceptor.Stop()

	// Verify defaults were applied
	s.Equal(defaultMaxFailuresPerMinute, interceptor.config.maxFailuresPerMinute)
	s.Equal(defaultLockoutDuration, interceptor.config.lockoutDuration)
	s.Equal(defaultCleanupInterval, interceptor.config.cleanupInterval)
	s.Equal(defaultMaxTrackedIPs, interceptor.config.maxTrackedIPs)
}

// Helper functions

func (s *authRateLimitInterceptorSuite) createContextWithIP(ip string) context.Context {
	addr := &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: 12345,
	}
	return peer.NewContext(context.Background(), &peer.Peer{
		Addr: addr,
	})
}

func (s *authRateLimitInterceptorSuite) makeIP(n int) string {
	return "192.168.1." + string(rune('0'+n))
}
