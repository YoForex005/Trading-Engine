package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// SessionsHandler handles session management endpoints
type SessionsHandler struct {
	sessionStore *SessionStore
	authService  *AuthService
}

// NewSessionsHandler creates a new sessions handler
func NewSessionsHandler(sessionStore *SessionStore, authService *AuthService) *SessionsHandler {
	return &SessionsHandler{
		sessionStore: sessionStore,
		authService:  authService,
	}
}

// HandleGetMySessions handles GET /admin/sessions - list caller's active sessions
func (h *SessionsHandler) HandleGetMySessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from auth context
	// TODO: Get from JWT token / auth middleware
	// For now, use a placeholder - should be extracted from authenticated session
	userID := r.Header.Get("X-User-ID") // Temporary - should come from JWT
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessions := h.sessionStore.GetActiveUserSessions(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// HandleGetUserSessions handles GET /admin/sessions/user/:id - list user's sessions (SUPER_ADMIN only)
func (h *SessionsHandler) HandleGetUserSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract target user ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/sessions/user/")
	targetUserID := path

	if targetUserID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	// TODO: Check if caller is SUPER_ADMIN
	// For now, allow all authenticated admins
	// Should check: if callerAdmin.Role != RoleSuperAdmin { return 403 }

	sessions := h.sessionStore.GetUserSessions(targetUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":   targetUserID,
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// HandleRevokeSession handles DELETE /admin/sessions/:id - revoke specific session
func (h *SessionsHandler) HandleRevokeSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/sessions/")
	sessionID := path

	if sessionID == "" || sessionID == "revoke-all" || strings.HasPrefix(sessionID, "user/") {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	// Get caller ID
	// TODO: Extract from JWT token
	callerID := r.Header.Get("X-User-ID")
	if callerID == "" {
		callerID = "admin-user" // Placeholder
	}

	// Verify session exists and caller owns it (or is SUPER_ADMIN)
	session, err := h.sessionStore.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Check ownership
	// TODO: Allow SUPER_ADMIN to revoke any session
	if session.UserID != callerID {
		http.Error(w, "Cannot revoke another user's session", http.StatusForbidden)
		return
	}

	// Revoke the session
	if err := h.sessionStore.RevokeSession(sessionID, callerID); err != nil {
		log.Printf("[SessionsHandler] Failed to revoke session: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[SessionsHandler] Session %s revoked by %s", sessionID, callerID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Session revoked successfully",
		"sessionId": sessionID,
	})
}

// HandleRevokeAllMySessions handles POST /admin/sessions/revoke-all - revoke all caller's sessions
func (h *SessionsHandler) HandleRevokeAllMySessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get caller ID
	// TODO: Extract from JWT token
	callerID := r.Header.Get("X-User-ID")
	if callerID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Revoke all sessions
	count := h.sessionStore.RevokeAllUserSessions(callerID, callerID)

	log.Printf("[SessionsHandler] Revoked %d sessions for user %s", count, callerID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All sessions revoked successfully",
		"count":   count,
	})
}

// HandleRevokeUserSessions handles POST /admin/sessions/revoke-user/:id - revoke all for user (SUPER_ADMIN only)
func (h *SessionsHandler) HandleRevokeUserSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract target user ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/sessions/revoke-user/")
	targetUserID := path

	if targetUserID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	// Get caller ID
	// TODO: Extract from JWT token and check SUPER_ADMIN role
	callerID := r.Header.Get("X-User-ID")
	if callerID == "" {
		callerID = "admin-user" // Placeholder
	}

	// TODO: Check if caller is SUPER_ADMIN
	// if callerRole != RoleSuperAdmin { return 403 }

	// Revoke all sessions for target user
	count := h.sessionStore.RevokeAllUserSessions(targetUserID, callerID)

	log.Printf("[SessionsHandler] Admin %s revoked %d sessions for user %s", callerID, count, targetUserID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All user sessions revoked successfully",
		"userId":  targetUserID,
		"count":   count,
	})
}

// HandleGetSessionStats handles GET /admin/sessions/stats - get session statistics (SUPER_ADMIN only)
func (h *SessionsHandler) HandleGetSessionStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Check if caller is SUPER_ADMIN

	stats := h.sessionStore.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// RegisterRoutes registers all session management routes
func (h *SessionsHandler) RegisterRoutes(mux *http.ServeMux, corsMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// GET /admin/sessions/stats - get session statistics
	mux.HandleFunc("/admin/sessions/stats", corsMiddleware(h.HandleGetSessionStats))

	// POST /admin/sessions/revoke-all - revoke all caller's sessions
	mux.HandleFunc("/admin/sessions/revoke-all", corsMiddleware(h.HandleRevokeAllMySessions))

	// GET /admin/sessions/user/:id - list user's sessions (SUPER_ADMIN only)
	// POST /admin/sessions/revoke-user/:id - revoke all for user (SUPER_ADMIN only)
	mux.HandleFunc("/admin/sessions/user/", corsMiddleware(h.HandleGetUserSessions))
	mux.HandleFunc("/admin/sessions/revoke-user/", corsMiddleware(h.HandleRevokeUserSessions))

	// GET /admin/sessions - list caller's sessions
	// DELETE /admin/sessions/:id - revoke specific session
	mux.HandleFunc("/admin/sessions/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// This handles both /admin/sessions and /admin/sessions/:id
		if r.URL.Path == "/admin/sessions" {
			h.HandleGetMySessions(w, r)
		} else {
			h.HandleRevokeSession(w, r)
		}
	}))

	log.Println("[SessionsHandler] Session management routes registered (5 endpoints)")
}
