package admin

import (
	"encoding/json"
	"log"
	"net/http"
)

// RBACMiddleware provides role-based access control middleware
type RBACMiddleware struct {
	authService *AuthService
}

// NewRBACMiddleware creates a new RBAC middleware instance
func NewRBACMiddleware(authService *AuthService) *RBACMiddleware {
	return &RBACMiddleware{
		authService: authService,
	}
}

// RequirePermission returns a middleware that checks for a specific permission
func (m *RBACMiddleware) RequirePermission(permission Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cors(w)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Authenticate the admin
			admin, err := m.authenticateRequest(r)
			if err != nil {
				m.respondError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if admin has the required permission
			if !HasPermission(admin.Role, permission) {
				log.Printf("[RBAC] Permission denied: %s (role: %s) attempted %s", admin.Username, admin.Role, permission)
				m.respondError(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Permission granted, proceed to handler
			next(w, r)
		}
	}
}

// RequireRole returns a middleware that checks for a specific role level
func (m *RBACMiddleware) RequireRole(requiredRole AdminRole) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cors(w)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Authenticate the admin
			admin, err := m.authenticateRequest(r)
			if err != nil {
				m.respondError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if admin's role meets the requirement
			if !HasRole(admin.Role, requiredRole) {
				log.Printf("[RBAC] Role denied: %s (role: %s) requires at least %s", admin.Username, admin.Role, requiredRole)
				m.respondError(w, "Insufficient role level", http.StatusForbidden)
				return
			}

			// Role requirement met, proceed to handler
			next(w, r)
		}
	}
}

// RequireAnyPermission returns a middleware that checks for any of the specified permissions
func (m *RBACMiddleware) RequireAnyPermission(permissions ...Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cors(w)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Authenticate the admin
			admin, err := m.authenticateRequest(r)
			if err != nil {
				m.respondError(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if admin has any of the required permissions
			hasAnyPermission := false
			for _, permission := range permissions {
				if HasPermission(admin.Role, permission) {
					hasAnyPermission = true
					break
				}
			}

			if !hasAnyPermission {
				log.Printf("[RBAC] Permission denied: %s (role: %s) requires one of %v", admin.Username, admin.Role, permissions)
				m.respondError(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Permission granted, proceed to handler
			next(w, r)
		}
	}
}

// authenticateRequest extracts and validates the session token from the request
func (m *RBACMiddleware) authenticateRequest(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	// Extract Bearer token
	parts := len(authHeader)
	if parts < 7 || authHeader[:7] != "Bearer " {
		return nil, http.ErrNoCookie
	}

	sessionID := authHeader[7:]
	ipAddress := getIPAddress(r)

	admin, err := m.authService.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// respondError sends a JSON error response
func (m *RBACMiddleware) respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
