package middleware

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/epic1st/rtx/backend/admin"
)

// IPFilterMiddleware wraps HTTP handlers to check IP whitelist/blacklist
type IPFilterMiddleware struct {
	store *admin.IPFilterStore
}

// NewIPFilterMiddleware creates a new IP filter middleware
func NewIPFilterMiddleware(store *admin.IPFilterStore) *IPFilterMiddleware {
	return &IPFilterMiddleware{
		store: store,
	}
}

// Wrap returns a middleware function that checks IP filtering
func (m *IPFilterMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract real IP address
		ip := m.extractRealIP(r)

		// Check if IP is allowed
		allowed, reason := m.store.CheckIP(ip)

		if !allowed {
			// Log blocked attempt
			log.Printf("[IPFilter] BLOCKED request from %s to %s %s (reason: %s)",
				ip, r.Method, r.URL.Path, reason)

			// Return 403 Forbidden
			http.Error(w, "Access denied: "+reason, http.StatusForbidden)
			return
		}

		// IP is allowed, continue to next handler
		next.ServeHTTP(w, r)
	})
}

// WrapFunc is a convenience method for wrapping http.HandlerFunc
func (m *IPFilterMiddleware) WrapFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract real IP address
		ip := m.extractRealIP(r)

		// Check if IP is allowed
		allowed, reason := m.store.CheckIP(ip)

		if !allowed {
			// Log blocked attempt
			log.Printf("[IPFilter] BLOCKED request from %s to %s %s (reason: %s)",
				ip, r.Method, r.URL.Path, reason)

			// Return 403 Forbidden
			http.Error(w, "Access denied: "+reason, http.StatusForbidden)
			return
		}

		// IP is allowed, continue to next handler
		next(w, r)
	}
}

// extractRealIP extracts the real client IP address from request headers
// Priority: X-Forwarded-For > X-Real-IP > RemoteAddr
func (m *IPFilterMiddleware) extractRealIP(r *http.Request) string {
	// Try X-Forwarded-For header (most common behind proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first one (original client)
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if m.isValidIP(ip) {
				return ip
			}
		}
	}

	// Try X-Real-IP header (common with nginx)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		ip := strings.TrimSpace(xri)
		if m.isValidIP(ip) {
			return ip
		}
	}

	// Fallback to RemoteAddr
	// RemoteAddr format: "IP:port" or "[IPv6]:port"
	remoteAddr := r.RemoteAddr

	// Strip port if present
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}

	// If no port, return as-is
	return remoteAddr
}

// isValidIP checks if a string is a valid IP address
func (m *IPFilterMiddleware) isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
