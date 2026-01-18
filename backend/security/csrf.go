package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"
)

// CSRFProtection handles Cross-Site Request Forgery protection
type CSRFProtection struct {
	secret     []byte
	tokenStore map[string]tokenData
	mu         sync.RWMutex
	maxAge     time.Duration
}

type tokenData struct {
	token      string
	createdAt  time.Time
	sessionID  string
}

// NewCSRFProtection creates a new CSRF protection instance
func NewCSRFProtection(secret string) *CSRFProtection {
	if secret == "" {
		// Generate random secret if not provided
		randomBytes := make([]byte, 32)
		rand.Read(randomBytes)
		secret = base64.StdEncoding.EncodeToString(randomBytes)
	}

	csrf := &CSRFProtection{
		secret:     []byte(secret),
		tokenStore: make(map[string]tokenData),
		maxAge:     24 * time.Hour, // 24 hour token validity
	}

	// Start cleanup routine
	go csrf.cleanupRoutine()

	return csrf
}

// GenerateToken generates a new CSRF token for a session
func (c *CSRFProtection) GenerateToken(sessionID string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Sign token with HMAC
	signature := c.signToken(token, sessionID)
	signedToken := token + "." + signature

	// Store token
	c.mu.Lock()
	c.tokenStore[sessionID] = tokenData{
		token:     signedToken,
		createdAt: time.Now(),
		sessionID: sessionID,
	}
	c.mu.Unlock()

	return signedToken, nil
}

// ValidateToken validates a CSRF token
func (c *CSRFProtection) ValidateToken(sessionID, token string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if token exists for session
	storedToken, exists := c.tokenStore[sessionID]
	if !exists {
		return errors.New("no CSRF token found for session")
	}

	// Check token age
	if time.Since(storedToken.createdAt) > c.maxAge {
		return errors.New("CSRF token expired")
	}

	// Verify token matches
	if token != storedToken.token {
		return errors.New("CSRF token mismatch")
	}

	// Verify signature
	parts := splitToken(token)
	if len(parts) != 2 {
		return errors.New("invalid CSRF token format")
	}

	expectedSignature := c.signToken(parts[0], sessionID)
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return errors.New("CSRF token signature invalid")
	}

	return nil
}

// signToken creates HMAC signature for a token
func (c *CSRFProtection) signToken(token, sessionID string) string {
	h := hmac.New(sha256.New, c.secret)
	h.Write([]byte(token))
	h.Write([]byte(sessionID))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// Middleware returns HTTP middleware that enforces CSRF protection
func (c *CSRFProtection) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get session ID from cookie or header
		sessionID := c.extractSessionID(r)
		if sessionID == "" {
			http.Error(w, "Missing session", http.StatusForbidden)
			return
		}

		// Get CSRF token from header or form
		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			token = r.FormValue("csrf_token")
		}

		if token == "" {
			http.Error(w, "Missing CSRF token", http.StatusForbidden)
			return
		}

		// Validate token
		if err := c.ValidateToken(sessionID, token); err != nil {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractSessionID extracts session ID from request
func (c *CSRFProtection) extractSessionID(r *http.Request) string {
	// Try to get from cookie
	cookie, err := r.Cookie("session_id")
	if err == nil {
		return cookie.Value
	}

	// Try to get from Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// Extract from Bearer token (this is simplified)
		return auth
	}

	return ""
}

// cleanupRoutine periodically removes expired tokens
func (c *CSRFProtection) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired tokens
func (c *CSRFProtection) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for sessionID, tokenData := range c.tokenStore {
		if now.Sub(tokenData.createdAt) > c.maxAge {
			delete(c.tokenStore, sessionID)
		}
	}
}

// Helper function to split token
func splitToken(token string) []string {
	var parts []string
	start := 0

	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}

	if start < len(token) {
		parts = append(parts, token[start:])
	}

	return parts
}
