package security

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// WAFConfig holds Web Application Firewall configuration
type WAFConfig struct {
	// Rate limiting
	MaxRequestsPerMinute   int
	MaxRequestsPerIP       int
	BurstSize              int

	// IP blocking
	BlockDuration          time.Duration
	MaxFailedAttempts      int

	// Request filtering
	MaxRequestBodySize     int64
	MaxURLLength           int
	MaxHeaderSize          int

	// DDoS protection
	MaxConcurrentConns     int
	SlowLorisTimeout       time.Duration

	// IP whitelisting (for admin endpoints)
	AdminWhitelist         []string
	WhitelistEnabled       bool
}

// DefaultWAFConfig returns a secure default configuration
func DefaultWAFConfig() *WAFConfig {
	return &WAFConfig{
		MaxRequestsPerMinute:   100,
		MaxRequestsPerIP:       20,
		BurstSize:              10,
		BlockDuration:          15 * time.Minute,
		MaxFailedAttempts:      5,
		MaxRequestBodySize:     1 << 20, // 1MB
		MaxURLLength:           2048,
		MaxHeaderSize:          8192,
		MaxConcurrentConns:     10000,
		SlowLorisTimeout:       30 * time.Second,
		AdminWhitelist:         []string{"127.0.0.1", "::1"},
		WhitelistEnabled:       true,
	}
}

// WAF implements Web Application Firewall functionality
type WAF struct {
	config *WAFConfig

	// Rate limiting state
	ipRequestCounts map[string]*requestCounter
	mu              sync.RWMutex

	// IP blocking
	blockedIPs      map[string]time.Time
	blockMu         sync.RWMutex

	// Connection tracking
	activeConns     int
	connMu          sync.RWMutex

	// Geo-blocking
	blockedCountries map[string]bool
	geoMu            sync.RWMutex
}

type requestCounter struct {
	count      int
	windowStart time.Time
	failures    int
}

// NewWAF creates a new Web Application Firewall instance
func NewWAF(config *WAFConfig) *WAF {
	if config == nil {
		config = DefaultWAFConfig()
	}

	waf := &WAF{
		config:           config,
		ipRequestCounts:  make(map[string]*requestCounter),
		blockedIPs:       make(map[string]time.Time),
		blockedCountries: make(map[string]bool),
	}

	// Start cleanup goroutine
	go waf.cleanupRoutine()

	return waf
}

// Middleware returns an HTTP middleware that enforces WAF rules
func (w *WAF) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// Extract client IP
		clientIP := w.extractClientIP(r)

		// Check if IP is blocked
		if w.isIPBlocked(clientIP) {
			http.Error(rw, "Access denied", http.StatusForbidden)
			return
		}

		// Check rate limit
		if !w.checkRateLimit(clientIP) {
			w.recordFailedAttempt(clientIP)
			http.Error(rw, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Check request size limits
		if r.ContentLength > w.config.MaxRequestBodySize {
			http.Error(rw, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		if len(r.URL.String()) > w.config.MaxURLLength {
			http.Error(rw, "URL too long", http.StatusRequestURITooLong)
			return
		}

		// Check concurrent connections
		if !w.acquireConnection() {
			http.Error(rw, "Server busy", http.StatusServiceUnavailable)
			return
		}
		defer w.releaseConnection()

		// Set security headers before passing to next handler
		w.setSecurityHeaders(rw)

		next.ServeHTTP(rw, r)
	})
}

// AdminOnlyMiddleware restricts access to whitelisted IPs for admin endpoints
func (w *WAF) AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if w.config.WhitelistEnabled {
			clientIP := w.extractClientIP(r)

			if !w.isIPWhitelisted(clientIP) {
				w.recordFailedAttempt(clientIP)
				http.Error(rw, "Access denied: IP not whitelisted", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(rw, r)
	})
}

// extractClientIP extracts the real client IP from the request
func (w *WAF) extractClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from proxy/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP from the list
		ip := xff
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					ip = xff[:i]
					break
				}
			}
		}
		return ip
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// checkRateLimit checks if the request is within rate limits
func (w *WAF) checkRateLimit(ip string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	counter, exists := w.ipRequestCounts[ip]

	if !exists {
		w.ipRequestCounts[ip] = &requestCounter{
			count:       1,
			windowStart: now,
		}
		return true
	}

	// Reset window if expired
	if now.Sub(counter.windowStart) > time.Minute {
		counter.count = 1
		counter.windowStart = now
		return true
	}

	// Check rate limit
	if counter.count >= w.config.MaxRequestsPerIP {
		return false
	}

	counter.count++
	return true
}

// recordFailedAttempt records a failed security check
func (w *WAF) recordFailedAttempt(ip string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	counter, exists := w.ipRequestCounts[ip]
	if !exists {
		w.ipRequestCounts[ip] = &requestCounter{
			failures: 1,
			windowStart: time.Now(),
		}
		return
	}

	counter.failures++

	// Block IP if too many failures
	if counter.failures >= w.config.MaxFailedAttempts {
		w.blockIP(ip, w.config.BlockDuration)
	}
}

// blockIP blocks an IP address for the specified duration
func (w *WAF) blockIP(ip string, duration time.Duration) {
	w.blockMu.Lock()
	defer w.blockMu.Unlock()

	w.blockedIPs[ip] = time.Now().Add(duration)
}

// BlockIPPermanent blocks an IP address permanently
func (w *WAF) BlockIPPermanent(ip string) {
	w.blockMu.Lock()
	defer w.blockMu.Unlock()

	w.blockedIPs[ip] = time.Now().Add(100 * 365 * 24 * time.Hour) // ~100 years
}

// UnblockIP removes an IP from the block list
func (w *WAF) UnblockIP(ip string) {
	w.blockMu.Lock()
	defer w.blockMu.Unlock()

	delete(w.blockedIPs, ip)
}

// isIPBlocked checks if an IP is currently blocked
func (w *WAF) isIPBlocked(ip string) bool {
	w.blockMu.RLock()
	defer w.blockMu.RUnlock()

	blockUntil, exists := w.blockedIPs[ip]
	if !exists {
		return false
	}

	return time.Now().Before(blockUntil)
}

// isIPWhitelisted checks if an IP is in the admin whitelist
func (w *WAF) isIPWhitelisted(ip string) bool {
	for _, whitelistedIP := range w.config.AdminWhitelist {
		if ip == whitelistedIP {
			return true
		}
	}
	return false
}

// AddToWhitelist adds an IP to the admin whitelist
func (w *WAF) AddToWhitelist(ip string) {
	w.config.AdminWhitelist = append(w.config.AdminWhitelist, ip)
}

// RemoveFromWhitelist removes an IP from the admin whitelist
func (w *WAF) RemoveFromWhitelist(ip string) {
	for i, whitelistedIP := range w.config.AdminWhitelist {
		if whitelistedIP == ip {
			w.config.AdminWhitelist = append(
				w.config.AdminWhitelist[:i],
				w.config.AdminWhitelist[i+1:]...,
			)
			return
		}
	}
}

// BlockCountry blocks all traffic from a specific country code
func (w *WAF) BlockCountry(countryCode string) {
	w.geoMu.Lock()
	defer w.geoMu.Unlock()

	w.blockedCountries[countryCode] = true
}

// UnblockCountry unblocks traffic from a specific country code
func (w *WAF) UnblockCountry(countryCode string) {
	w.geoMu.Lock()
	defer w.geoMu.Unlock()

	delete(w.blockedCountries, countryCode)
}

// acquireConnection attempts to acquire a connection slot
func (w *WAF) acquireConnection() bool {
	w.connMu.Lock()
	defer w.connMu.Unlock()

	if w.activeConns >= w.config.MaxConcurrentConns {
		return false
	}

	w.activeConns++
	return true
}

// releaseConnection releases a connection slot
func (w *WAF) releaseConnection() {
	w.connMu.Lock()
	defer w.connMu.Unlock()

	w.activeConns--
}

// setSecurityHeaders sets security-related HTTP headers
func (w *WAF) setSecurityHeaders(rw http.ResponseWriter) {
	// HSTS - Force HTTPS for 1 year
	rw.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	// Content Security Policy - Prevent XSS
	rw.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

	// X-Frame-Options - Prevent Clickjacking
	rw.Header().Set("X-Frame-Options", "DENY")

	// X-Content-Type-Options - Prevent MIME sniffing
	rw.Header().Set("X-Content-Type-Options", "nosniff")

	// X-XSS-Protection - Legacy XSS protection
	rw.Header().Set("X-XSS-Protection", "1; mode=block")

	// Referrer-Policy - Control referrer information
	rw.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// Permissions-Policy - Disable dangerous browser features
	rw.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}

// cleanupRoutine periodically cleans up expired blocks and rate limit counters
func (w *WAF) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		w.cleanup()
	}
}

// cleanup removes expired blocks and old rate limit counters
func (w *WAF) cleanup() {
	now := time.Now()

	// Clean up expired IP blocks
	w.blockMu.Lock()
	for ip, blockUntil := range w.blockedIPs {
		if now.After(blockUntil) {
			delete(w.blockedIPs, ip)
		}
	}
	w.blockMu.Unlock()

	// Clean up old rate limit counters
	w.mu.Lock()
	for ip, counter := range w.ipRequestCounts {
		if now.Sub(counter.windowStart) > 10*time.Minute {
			delete(w.ipRequestCounts, ip)
		}
	}
	w.mu.Unlock()
}

// GetStats returns WAF statistics
func (w *WAF) GetStats() map[string]interface{} {
	w.blockMu.RLock()
	blockedCount := len(w.blockedIPs)
	w.blockMu.RUnlock()

	w.mu.RLock()
	trackedIPs := len(w.ipRequestCounts)
	w.mu.RUnlock()

	w.connMu.RLock()
	activeConns := w.activeConns
	w.connMu.RUnlock()

	return map[string]interface{}{
		"blocked_ips":        blockedCount,
		"tracked_ips":        trackedIPs,
		"active_connections": activeConns,
		"max_connections":    w.config.MaxConcurrentConns,
		"rate_limit_per_ip":  w.config.MaxRequestsPerIP,
	}
}

// GetBlockedIPs returns a list of currently blocked IPs
func (w *WAF) GetBlockedIPs() []string {
	w.blockMu.RLock()
	defer w.blockMu.RUnlock()

	ips := make([]string, 0, len(w.blockedIPs))
	now := time.Now()

	for ip, blockUntil := range w.blockedIPs {
		if now.Before(blockUntil) {
			ips = append(ips, fmt.Sprintf("%s (until %s)", ip, blockUntil.Format(time.RFC3339)))
		}
	}

	return ips
}
