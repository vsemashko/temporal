package interceptor

import (
	"context"
	"net"
	"sync"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	// Default configuration values
	defaultMaxFailuresPerMinute = 10
	defaultLockoutDuration      = 5 * time.Minute
	defaultCleanupInterval      = 10 * time.Minute
	defaultMaxTrackedIPs        = 10000
)

var (
	errAuthRateLimited = serviceerror.NewResourceExhausted(
		enumspb.RESOURCE_EXHAUSTED_CAUSE_RPS_LIMIT,
		"authentication rate limit exceeded - too many failed attempts",
	)
)

// authRateLimitConfig contains internal configuration with defaults applied
type authRateLimitConfig struct {
	enabled              bool
	maxFailuresPerMinute int
	lockoutDuration      time.Duration
	cleanupInterval      time.Duration
	maxTrackedIPs        int
}

// authFailureTracker tracks authentication failures for a single IP address
type authFailureTracker struct {
	failures     int
	firstFailure time.Time
	lockedUntil  time.Time
	lastFailure  time.Time
}

// AuthRateLimitInterceptor implements rate limiting for authentication failures
type AuthRateLimitInterceptor struct {
	config         authRateLimitConfig
	trackers       map[string]*authFailureTracker
	mu             sync.RWMutex
	logger         log.Logger
	metricsHandler metrics.Handler
	stopCh         chan struct{}
	doneCh         chan struct{}
}

// NewAuthRateLimitInterceptor creates a new authentication rate limit interceptor
func NewAuthRateLimitInterceptor(
	cfg config.AuthRateLimit,
	logger log.Logger,
	metricsHandler metrics.Handler,
) *AuthRateLimitInterceptor {
	// Build internal config with defaults
	internalCfg := authRateLimitConfig{
		enabled:              cfg.Enabled,
		maxFailuresPerMinute: cfg.MaxFailuresPerMinute,
		lockoutDuration:      cfg.LockoutDuration,
		cleanupInterval:      defaultCleanupInterval,
		maxTrackedIPs:        defaultMaxTrackedIPs,
	}

	// Set defaults
	if internalCfg.maxFailuresPerMinute == 0 {
		internalCfg.maxFailuresPerMinute = defaultMaxFailuresPerMinute
	}
	if internalCfg.lockoutDuration == 0 {
		internalCfg.lockoutDuration = defaultLockoutDuration
	}

	interceptor := &AuthRateLimitInterceptor{
		config:         internalCfg,
		trackers:       make(map[string]*authFailureTracker),
		logger:         logger,
		metricsHandler: metricsHandler,
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}

	// Start cleanup goroutine
	if internalCfg.enabled {
		go interceptor.cleanupLoop()
	}

	return interceptor
}

// Intercept implements gRPC unary interceptor for authentication rate limiting
func (i *AuthRateLimitInterceptor) Intercept(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if !i.config.enabled {
		return handler(ctx, req)
	}

	// Extract client IP address
	clientIP := i.getClientIP(ctx)
	if clientIP == "" {
		// If we can't get the IP, allow the request (fail open)
		i.logger.Warn("Unable to extract client IP for auth rate limiting")
		return handler(ctx, req)
	}

	// Check if IP is currently locked out
	if i.isLockedOut(clientIP) {
		i.logger.Warn("Authentication attempt blocked - IP locked out",
			tag.NewStringTag("client-ip", clientIP))
		metrics.AuthRateLimitedCounter.With(i.metricsHandler).Record(1)
		return nil, errAuthRateLimited
	}

	// Call the handler
	resp, err := handler(ctx, req)

	// Track authentication failures
	if i.isAuthenticationError(err) {
		i.recordFailure(clientIP)
		i.logger.Warn("Authentication failure recorded",
			tag.NewStringTag("client-ip", clientIP),
			tag.Error(err))
	} else if err == nil {
		// Clear failures on successful authentication
		i.clearFailures(clientIP)
	}

	return resp, err
}

// getClientIP extracts the client IP address from the context
func (i *AuthRateLimitInterceptor) getClientIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}

	addr := p.Addr.String()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If we can't split, return the whole address
		return addr
	}
	return host
}

// isAuthenticationError checks if an error is an authentication failure
func (i *AuthRateLimitInterceptor) isAuthenticationError(err error) bool {
	if err == nil {
		return false
	}
	// Check for permission denied errors which indicate auth failures
	_, isPermissionDenied := err.(*serviceerror.PermissionDenied)
	return isPermissionDenied
}

// isLockedOut checks if an IP is currently locked out
func (i *AuthRateLimitInterceptor) isLockedOut(ip string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	tracker, exists := i.trackers[ip]
	if !exists {
		return false
	}

	now := time.Now()
	if now.Before(tracker.lockedUntil) {
		return true
	}

	return false
}

// recordFailure records an authentication failure for an IP
func (i *AuthRateLimitInterceptor) recordFailure(ip string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	tracker, exists := i.trackers[ip]

	if !exists {
		// Create new tracker
		if len(i.trackers) >= i.config.maxTrackedIPs {
			// Too many IPs tracked, log warning and don't track this one
			i.logger.Warn("Maximum tracked IPs exceeded, not tracking new IP",
				tag.NewStringTag("client-ip", ip))
			metrics.AuthRateTrackerOverflow.With(i.metricsHandler).Record(1)
			return
		}

		tracker = &authFailureTracker{
			failures:     1,
			firstFailure: now,
			lastFailure:  now,
		}
		i.trackers[ip] = tracker
		metrics.AuthFailureCounter.With(i.metricsHandler).Record(1)
		return
	}

	// Update existing tracker
	tracker.failures++
	tracker.lastFailure = now

	// Check if we should reset the window
	if now.Sub(tracker.firstFailure) > time.Minute {
		// More than a minute since first failure, reset window
		tracker.failures = 1
		tracker.firstFailure = now
		tracker.lockedUntil = time.Time{} // Clear lockout
	}

	// Check if we should lock out this IP
	if tracker.failures >= i.config.maxFailuresPerMinute {
		tracker.lockedUntil = now.Add(i.config.lockoutDuration)
		i.logger.Warn("IP locked out due to excessive auth failures",
			tag.NewStringTag("client-ip", ip),
			tag.NewInt("failures", tracker.failures),
			tag.NewDurationTag("lockout-duration", i.config.lockoutDuration))
		metrics.AuthLockoutCounter.With(i.metricsHandler).Record(1)
	}

	metrics.AuthFailureCounter.With(i.metricsHandler).Record(1)
}

// clearFailures clears the failure count for an IP (called on successful auth)
func (i *AuthRateLimitInterceptor) clearFailures(ip string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	delete(i.trackers, ip)
}

// cleanupLoop periodically removes old tracking data
func (i *AuthRateLimitInterceptor) cleanupLoop() {
	ticker := time.NewTicker(i.config.cleanupInterval)
	defer ticker.Stop()
	defer close(i.doneCh)

	for {
		select {
		case <-ticker.C:
			i.cleanup()
		case <-i.stopCh:
			return
		}
	}
}

// cleanup removes old tracking data
func (i *AuthRateLimitInterceptor) cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()

	now := time.Now()
	for ip, tracker := range i.trackers {
		// Remove trackers that are no longer locked out and haven't had failures recently
		if now.After(tracker.lockedUntil) && now.Sub(tracker.lastFailure) > i.config.cleanupInterval {
			delete(i.trackers, ip)
		}
	}

	// Update metrics
	metrics.AuthTrackedIPsGauge.With(i.metricsHandler).Record(float64(len(i.trackers)))
}

// Stop stops the cleanup loop
func (i *AuthRateLimitInterceptor) Stop() {
	close(i.stopCh)
	<-i.doneCh
}
