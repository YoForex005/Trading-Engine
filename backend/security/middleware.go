package security

import (
	"net/http"
	"strings"
)

// SecurityMiddleware combines all security protections
type SecurityMiddleware struct {
	waf            *WAF
	csrf           *CSRFProtection
	sessionManager *SessionManager
	auditLogger    *AuditLogger
}

// NewSecurityMiddleware creates a comprehensive security middleware
func NewSecurityMiddleware(waf *WAF, csrf *CSRFProtection, sessionManager *SessionManager, auditLogger *AuditLogger) *SecurityMiddleware {
	return &SecurityMiddleware{
		waf:            waf,
		csrf:           csrf,
		sessionManager: sessionManager,
		auditLogger:    auditLogger,
	}
}

// Protect wraps a handler with all security protections
func (sm *SecurityMiddleware) Protect(next http.Handler) http.Handler {
	return sm.waf.Middleware(
		sm.csrf.Middleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Session validation
				sessionID := sm.extractSessionID(r)
				if sessionID != "" {
					clientIP := sm.waf.extractClientIP(r)
					if _, err := sm.sessionManager.ValidateSession(sessionID, clientIP); err != nil {
						http.Error(w, "Invalid session", http.StatusUnauthorized)
						return
					}
				}

				next.ServeHTTP(w, r)
			}),
		),
	)
}

// AdminProtect protects admin endpoints with additional security
func (sm *SecurityMiddleware) AdminProtect(next http.Handler) http.Handler {
	return sm.waf.AdminOnlyMiddleware(
		sm.Protect(next),
	)
}

// extractSessionID extracts session ID from request
func (sm *SecurityMiddleware) extractSessionID(r *http.Request) string {
	// Try cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}

	// Try Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// InputSanitizationMiddleware sanitizes all request inputs
func InputSanitizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse form if present
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			r.ParseForm()

			// Sanitize form values
			for key := range r.Form {
				values := r.Form[key]
				for i, value := range values {
					r.Form[key][i] = SanitizeInput(value)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HSTS - Force HTTPS
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
			"script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"font-src 'self'; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self'")

		// X-Frame-Options - Prevent Clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options - Prevent MIME sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection - Legacy XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")

		// Remove Server header
		w.Header().Del("Server")

		// X-Permitted-Cross-Domain-Policies
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS with security
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// Don't set CORS headers for disallowed origins
				w.Header().Set("Access-Control-Allow-Origin", "null")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestValidationMiddleware validates common request properties
func RequestValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate Content-Type for POST/PUT
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" &&
			   !strings.Contains(contentType, "application/json") &&
			   !strings.Contains(contentType, "application/x-www-form-urlencoded") &&
			   !strings.Contains(contentType, "multipart/form-data") {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}
		}

		// Validate User-Agent (block empty or suspicious)
		userAgent := r.Header.Get("User-Agent")
		if userAgent == "" {
			http.Error(w, "Missing User-Agent", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
