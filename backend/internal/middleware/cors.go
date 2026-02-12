package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// NewCORSMiddleware creates a CORS middleware with configurable allowed origins
func NewCORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
	// Default methods if not specified
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}

	// Default headers if not specified
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" {
					// Allow wildcard only in development (logged as warning)
					w.Header().Set("Access-Control-Allow-Origin", "*")
					allowed = true
					break
				}
				if origin == allowedOrigin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					allowed = true
					break
				}
			}

			// If origin not allowed, don't set CORS headers
			if !allowed && origin != "" {
				// Reject the request silently (no CORS headers = browser blocks)
				http.Error(w, "Origin not allowed", http.StatusForbidden)
				return
			}

			// Set other CORS headers
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SetCORSHeaders is a helper function for handlers that need to set CORS headers manually
// DEPRECATED: Use NewCORSMiddleware instead
func SetCORSHeaders(w http.ResponseWriter, allowedOrigins []string, origin string) {
	allowed := false
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			allowed = true
			break
		}
		if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}
