package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// RequestsPerSecond is the number of requests allowed per second
	RequestsPerSecond float64
	// RequestsPerMinute is the number of requests allowed per minute
	RequestsPerMinute float64
	// BurstSize is the maximum number of requests allowed in a burst
	BurstSize int
	// CleanupInterval is how often to clean up inactive clients
	CleanupInterval time.Duration
	// ClientTimeout is how long to keep a client's limiter after last request
	ClientTimeout time.Duration
}

// DefaultRateLimitConfig returns a sensible default configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 10,      // 10 requests per second
		RequestsPerMinute: 500,     // 500 requests per minute
		BurstSize:         20,      // Allow bursts of 20 requests
		CleanupInterval:   5 * time.Minute,
		ClientTimeout:     10 * time.Minute,
	}
}

// RateLimiter implements per-IP rate limiting
type RateLimiter struct {
	config    RateLimitConfig
	limiters  map[string]*clientLimiter
	mu        sync.RWMutex
	stopCh    chan struct{}
	once      sync.Once
}

// clientLimiter holds rate limiter info for a single client
type clientLimiter struct {
	limiter   *rate.Limiter
	lastSeen  time.Time
	requestID string
}

// NewRateLimiter creates a new rate limiter with the given configuration
func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:   config,
		limiters: make(map[string]*clientLimiter),
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanupInactiveClients()

	return rl
}

// Middleware returns an HTTP middleware function for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := rl.getClientIP(r)
		allowed, remaining, resetTime := rl.Allow(clientIP)

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", rl.config.RequestsPerSecond))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		if !allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(int64(resetTime.Sub(time.Now()).Seconds()), 10))
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareWithExclusions returns a rate limiting middleware that excludes certain endpoints
func (rl *RateLimiter) MiddlewareWithExclusions(exclusions []string) func(http.Handler) http.Handler {
	excludeMap := make(map[string]bool)
	for _, path := range exclusions {
		excludeMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for excluded paths
			if excludeMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			clientIP := rl.getClientIP(r)
			allowed, remaining, resetTime := rl.Allow(clientIP)

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", rl.config.RequestsPerSecond))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(resetTime.Sub(time.Now()).Seconds()), 10))
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Allow checks if a request from the given IP is allowed
// Returns: allowed (bool), remaining (int64), resetTime (time.Time)
func (rl *RateLimiter) Allow(clientIP string) (bool, int64, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cl, exists := rl.limiters[clientIP]
	if !exists {
		// Create a new limiter for this client
		// Use per-second limit with burst size
		limit := rate.Limit(rl.config.RequestsPerSecond)
		cl = &clientLimiter{
			limiter:  rate.NewLimiter(limit, rl.config.BurstSize),
			lastSeen: time.Now(),
		}
		rl.limiters[clientIP] = cl
	}

	cl.lastSeen = time.Now()

	// Check if request is allowed
	allowed := cl.limiter.Allow()

	// Calculate remaining requests and reset time
	tokens := cl.limiter.Tokens()
	remaining := int64(tokens)
	if remaining < 0 {
		remaining = 0
	}

	// Reset time is 1 second from now (token refill interval)
	resetTime := time.Now().Add(time.Second)

	return allowed, remaining, resetTime
}

// AllowN checks if n requests from the given IP are allowed
func (rl *RateLimiter) AllowN(clientIP string, n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cl, exists := rl.limiters[clientIP]
	if !exists {
		limit := rate.Limit(rl.config.RequestsPerSecond)
		cl = &clientLimiter{
			limiter:  rate.NewLimiter(limit, rl.config.BurstSize),
			lastSeen: time.Now(),
		}
		rl.limiters[clientIP] = cl
	}

	cl.lastSeen = time.Now()
	return cl.limiter.AllowN(time.Now(), n)
}

// ResetClient removes the rate limiter for the given client IP
func (rl *RateLimiter) ResetClient(clientIP string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limiters, clientIP)
}

// cleanupInactiveClients periodically removes inactive client limiters
func (rl *RateLimiter) cleanupInactiveClients() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.performCleanup()
		}
	}
}

// performCleanup removes client limiters that haven't been used recently
func (rl *RateLimiter) performCleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timeout := rl.config.ClientTimeout

	for ip, cl := range rl.limiters {
		if now.Sub(cl.lastSeen) > timeout {
			delete(rl.limiters, ip)
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.once.Do(func() {
		close(rl.stopCh)
	})
}

// GetStats returns statistics about the rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return map[string]interface{}{
		"active_clients":           len(rl.limiters),
		"requests_per_second":      rl.config.RequestsPerSecond,
		"requests_per_minute":      rl.config.RequestsPerMinute,
		"burst_size":               rl.config.BurstSize,
		"cleanup_interval_seconds": int(rl.config.CleanupInterval.Seconds()),
		"client_timeout_seconds":   int(rl.config.ClientTimeout.Seconds()),
	}
}

// getClientIP extracts the client IP from the request
// Checks X-Forwarded-For, X-Real-IP headers first for proxy support
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ip, _, err := net.SplitHostPort(xff); err == nil {
			return ip
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

// KeyBasedRateLimiter allows rate limiting by custom keys (user ID, API key, etc.)
type KeyBasedRateLimiter struct {
	config    RateLimitConfig
	limiters  map[string]*clientLimiter
	mu        sync.RWMutex
	stopCh    chan struct{}
	once      sync.Once
}

// NewKeyBasedRateLimiter creates a new key-based rate limiter
func NewKeyBasedRateLimiter(config RateLimitConfig) *KeyBasedRateLimiter {
	krl := &KeyBasedRateLimiter{
		config:   config,
		limiters: make(map[string]*clientLimiter),
		stopCh:   make(chan struct{}),
	}

	go krl.cleanupInactiveClients()
	return krl
}

// Allow checks if a request with the given key is allowed
func (krl *KeyBasedRateLimiter) Allow(key string) (bool, int64, time.Time) {
	krl.mu.Lock()
	defer krl.mu.Unlock()

	cl, exists := krl.limiters[key]
	if !exists {
		limit := rate.Limit(krl.config.RequestsPerSecond)
		cl = &clientLimiter{
			limiter:  rate.NewLimiter(limit, krl.config.BurstSize),
			lastSeen: time.Now(),
		}
		krl.limiters[key] = cl
	}

	cl.lastSeen = time.Now()

	allowed := cl.limiter.Allow()
	tokens := cl.limiter.Tokens()
	remaining := int64(tokens)
	if remaining < 0 {
		remaining = 0
	}
	resetTime := time.Now().Add(time.Second)

	return allowed, remaining, resetTime
}

// Reset removes the rate limiter for the given key
func (krl *KeyBasedRateLimiter) Reset(key string) {
	krl.mu.Lock()
	defer krl.mu.Unlock()
	delete(krl.limiters, key)
}

// Stop stops the cleanup goroutine
func (krl *KeyBasedRateLimiter) Stop() {
	krl.once.Do(func() {
		close(krl.stopCh)
	})
}

// cleanupInactiveClients periodically removes inactive client limiters
func (krl *KeyBasedRateLimiter) cleanupInactiveClients() {
	ticker := time.NewTicker(krl.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-krl.stopCh:
			return
		case <-ticker.C:
			krl.performCleanup()
		}
	}
}

// performCleanup removes client limiters that haven't been used recently
func (krl *KeyBasedRateLimiter) performCleanup() {
	krl.mu.Lock()
	defer krl.mu.Unlock()

	now := time.Now()
	timeout := krl.config.ClientTimeout

	for key, cl := range krl.limiters {
		if now.Sub(cl.lastSeen) > timeout {
			delete(krl.limiters, key)
		}
	}
}
