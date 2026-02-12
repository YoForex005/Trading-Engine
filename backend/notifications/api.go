package notifications

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/epic1st/rtx/backend/auth"
)

// APIHandler handles HTTP requests for notifications
type APIHandler struct {
	manager     *NotificationManager
	store       *Store
	authService *auth.Service
}

// NewAPIHandler creates a new notification API handler
func NewAPIHandler(manager *NotificationManager, store *Store, authService *auth.Service) *APIHandler {
	return &APIHandler{
		manager:     manager,
		store:       store,
		authService: authService,
	}
}

// HandleListNotifications - GET /api/notifications
// Query params: ?unread=true (optional), ?limit=50 (default: 50), ?offset=0 (default: 0)
func (h *APIHandler) HandleListNotifications(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	unreadOnly := r.URL.Query().Get("unread") == "true"
	limit := parseIntParam(r.URL.Query().Get("limit"), 50)
	offset := parseIntParam(r.URL.Query().Get("offset"), 0)

	var notifications []*Notification

	if unreadOnly {
		notifications = h.store.GetUnread(userID)
	} else {
		notifications = h.store.GetHistory(userID, limit, offset)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
}

// HandleGetUnreadCount - GET /api/notifications/unread-count
func (h *APIHandler) HandleGetUnreadCount(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count := h.store.GetUnreadCount(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": count,
	})
}

// HandleMarkAsRead - PUT /api/notifications/{id}/read
func (h *APIHandler) HandleMarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract notification ID from URL path
	// Expected format: /api/notifications/{id}/read
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	notificationID := parts[3]

	err = h.store.MarkAsRead(userID, notificationID)
	if err != nil {
		if err == ErrNotificationNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorized {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notification marked as read",
	})
}

// HandleMarkAllRead - PUT /api/notifications/read-all
func (h *APIHandler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.store.MarkAllRead(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "All notifications marked as read",
	})
}

// HandleDeleteNotification - DELETE /api/notifications/{id}
func (h *APIHandler) HandleDeleteNotification(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token
	userID, err := h.getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract notification ID from URL path
	// Expected format: /api/notifications/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	notificationID := parts[3]

	err = h.store.Delete(userID, notificationID)
	if err != nil {
		if err == ErrNotificationNotFound {
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		}
		if err == ErrUnauthorized {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Notification deleted",
	})
}

// getUserIDFromToken extracts user ID from JWT token in Authorization header
func (h *APIHandler) getUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]

	// Validate token using auth service
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		log.Printf("[Notifications API] Token validation failed: %v", err)
		return "", fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

// parseIntParam parses an integer query parameter with a default value
func parseIntParam(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
