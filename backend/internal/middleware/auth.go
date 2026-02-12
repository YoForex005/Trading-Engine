package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/epic1st/rtx/backend/auth"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user info in request context
	UserContextKey ContextKey = "user"
)

// RequireAuth is a middleware that validates JWT tokens
func RequireAuth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Printf("[AUTH] Missing Authorization header from %s", r.RemoteAddr)
				http.Error(w, "Unauthorized: Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Expected format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("[AUTH] Invalid Authorization header format from %s", r.RemoteAddr)
				http.Error(w, "Unauthorized: Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				log.Printf("[AUTH] Invalid token from %s: %v", r.RemoteAddr, err)
				http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Create user from claims
			user := &auth.User{
				ID:       claims.UserID,
				Username: claims.Username,
				Role:     claims.Role,
			}

			// Add user to request context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin is a middleware that requires admin role
func RequireAdmin(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, require authentication
			RequireAuth(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get user from context
				user, ok := r.Context().Value(UserContextKey).(*auth.User)
				if !ok {
					log.Printf("[AUTH] Failed to get user from context")
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Check if user is admin
				if user.Role != "ADMIN" {
					log.Printf("[AUTH] Non-admin user %s attempted to access admin endpoint: %s", user.Username, r.URL.Path)
					http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
					return
				}

				next.ServeHTTP(w, r)
			})).ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts the authenticated user from the request context
func GetUserFromContext(r *http.Request) (*auth.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*auth.User)
	return user, ok
}

// OptionalAuth is a middleware that extracts user info if token is present, but doesn't require it
func OptionalAuth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token := parts[1]
					claims, err := authService.ValidateToken(token)
					if err == nil {
						user := &auth.User{
							ID:       claims.UserID,
							Username: claims.Username,
							Role:     claims.Role,
						}
						ctx := context.WithValue(r.Context(), UserContextKey, user)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
