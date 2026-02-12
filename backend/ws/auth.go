package ws

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ClientType represents the type of WebSocket client
type ClientType string

const (
	ClientTypePublic  ClientType = "public"  // Unauthenticated - market data only
	ClientTypePrivate ClientType = "private" // Authenticated - full access
)

// extractTokenFromRequest extracts JWT token from query parameter or header
func extractTokenFromRequest(r *http.Request) string {
	// Try query parameter first (ws://localhost/ws?token=xyz)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	// Fall back to Authorization header (Authorization: Bearer <token>)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	return ""
}

// validateClientAuth validates authentication and returns (userID, clientType, error)
func validateClientAuth(hub *Hub, r *http.Request, requireAuth bool) (string, ClientType, error) {
	token := extractTokenFromRequest(r)

	// If no token provided
	if token == "" {
		if requireAuth {
			return "", "", fmt.Errorf("authentication required")
		}
		// Public access allowed
		return "", ClientTypePublic, nil
	}

	// Token provided - validate it
	if hub.authService == nil {
		return "", "", fmt.Errorf("auth service not configured")
	}

	claims, err := hub.authService.ValidateToken(token)
	if err != nil {
		if requireAuth {
			return "", "", fmt.Errorf("invalid token: %v", err)
		}
		// Invalid token but public access allowed
		log.Printf("[WS-Auth] Invalid token provided, downgrading to public access: %v", err)
		return "", ClientTypePublic, nil
	}

	// Valid token
	userID := claims.UserID
	log.Printf("[WS-Auth] Authenticated user: %s", userID)
	return userID, ClientTypePrivate, nil
}
