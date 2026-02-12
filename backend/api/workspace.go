package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/epic1st/rtx/backend/models"
)

// WorkspaceHandler handles workspace-related HTTP requests
type WorkspaceHandler struct {
	db *sql.DB
}

// NewWorkspaceHandler creates a new workspace handler
func NewWorkspaceHandler(db *sql.DB) *WorkspaceHandler {
	return &WorkspaceHandler{db: db}
}

// HandleSaveWorkspace handles POST /api/workspace/save
func (h *WorkspaceHandler) HandleSaveWorkspace(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from JWT token (you'll need to implement this)
	userID := getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description,omitempty"`
		Config      models.WorkspaceConfig `json:"config"`
		IsDefault   bool                   `json:"isDefault"`
		Overwrite   bool                   `json:"overwrite"` // If true, overwrite existing
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Workspace] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.Name == "" {
		http.Error(w, "Workspace name is required", http.StatusBadRequest)
		return
	}

	if len(req.Name) > 255 {
		http.Error(w, "Workspace name too long (max 255 characters)", http.StatusBadRequest)
		return
	}

	// Set version if not provided
	if req.Config.Version == "" {
		req.Config.Version = "1.0"
	}

	// Check if workspace with this name already exists
	var existingID int64
	err := h.db.QueryRow(`
		SELECT id FROM workspaces
		WHERE user_id = $1 AND name = $2
	`, userID, req.Name).Scan(&existingID)

	if err == nil {
		// Workspace exists
		if !req.Overwrite {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":      "workspace_exists",
				"message":    "Workspace with this name already exists",
				"workspaceId": existingID,
				"requiresConfirmation": true,
			})
			return
		}

		// Update existing workspace
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			log.Printf("[Workspace] Failed to marshal config: %v", err)
			http.Error(w, "Failed to save workspace", http.StatusInternalServerError)
			return
		}

		_, err = h.db.Exec(`
			UPDATE workspaces
			SET description = $1, config = $2, is_default = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $4 AND user_id = $5
		`, req.Description, configJSON, req.IsDefault, existingID, userID)

		if err != nil {
			log.Printf("[Workspace] Failed to update workspace: %v", err)
			http.Error(w, "Failed to update workspace", http.StatusInternalServerError)
			return
		}

		log.Printf("[Workspace] Updated workspace %d for user %s", existingID, userID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     true,
			"workspaceId": existingID,
			"message":     "Workspace updated successfully",
			"action":      "updated",
		})
		return
	}

	// Create new workspace
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		log.Printf("[Workspace] Failed to marshal config: %v", err)
		http.Error(w, "Failed to save workspace", http.StatusInternalServerError)
		return
	}

	// If this is set as default, unset all other defaults for this user
	if req.IsDefault {
		_, err = h.db.Exec(`
			UPDATE workspaces SET is_default = FALSE WHERE user_id = $1
		`, userID)
		if err != nil {
			log.Printf("[Workspace] Failed to unset default workspaces: %v", err)
		}
	}

	var workspaceID int64
	err = h.db.QueryRow(`
		INSERT INTO workspaces (user_id, name, description, config, is_default)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, req.Name, req.Description, configJSON, req.IsDefault).Scan(&workspaceID)

	if err != nil {
		log.Printf("[Workspace] Failed to create workspace: %v", err)
		http.Error(w, "Failed to create workspace", http.StatusInternalServerError)
		return
	}

	log.Printf("[Workspace] Created workspace %d for user %s", workspaceID, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"workspaceId": workspaceID,
		"message":     "Workspace saved successfully",
		"action":      "created",
	})
}

// HandleLoadWorkspace handles GET /api/workspace/:id
func (h *WorkspaceHandler) HandleLoadWorkspace(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from JWT token
	userID := getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract workspace ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	workspaceIDStr := pathParts[3]
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	var workspace models.Workspace
	var configJSON []byte

	err = h.db.QueryRow(`
		SELECT id, user_id, name, description, config, is_default, created_at, updated_at
		FROM workspaces
		WHERE id = $1 AND user_id = $2
	`, workspaceID, userID).Scan(
		&workspace.ID,
		&workspace.UserID,
		&workspace.Name,
		&workspace.Description,
		&configJSON,
		&workspace.IsDefault,
		&workspace.CreatedAt,
		&workspace.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("[Workspace] Failed to load workspace: %v", err)
		http.Error(w, "Failed to load workspace", http.StatusInternalServerError)
		return
	}

	// Unmarshal config
	if err := json.Unmarshal(configJSON, &workspace.Config); err != nil {
		log.Printf("[Workspace] Failed to unmarshal config: %v", err)
		http.Error(w, "Failed to parse workspace configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspace)
}

// HandleListWorkspaces handles GET /api/workspaces
func (h *WorkspaceHandler) HandleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from JWT token
	userID := getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.db.Query(`
		SELECT id, name, description, is_default, created_at, updated_at
		FROM workspaces
		WHERE user_id = $1
		ORDER BY is_default DESC, updated_at DESC
	`, userID)

	if err != nil {
		log.Printf("[Workspace] Failed to list workspaces: %v", err)
		http.Error(w, "Failed to list workspaces", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var workspaces []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, description string
		var isDefault bool
		var createdAt, updatedAt string

		if err := rows.Scan(&id, &name, &description, &isDefault, &createdAt, &updatedAt); err != nil {
			log.Printf("[Workspace] Failed to scan workspace: %v", err)
			continue
		}

		workspaces = append(workspaces, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"isDefault":   isDefault,
			"createdAt":   createdAt,
			"updatedAt":   updatedAt,
		})
	}

	if workspaces == nil {
		workspaces = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspaces)
}

// HandleDeleteWorkspace handles DELETE /api/workspace/:id
func (h *WorkspaceHandler) HandleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from JWT token
	userID := getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract workspace ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	workspaceIDStr := pathParts[3]
	workspaceID, err := strconv.ParseInt(workspaceIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid workspace ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(`
		DELETE FROM workspaces
		WHERE id = $1 AND user_id = $2
	`, workspaceID, userID)

	if err != nil {
		log.Printf("[Workspace] Failed to delete workspace: %v", err)
		http.Error(w, "Failed to delete workspace", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Workspace not found", http.StatusNotFound)
		return
	}

	log.Printf("[Workspace] Deleted workspace %d for user %s", workspaceID, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Workspace deleted successfully",
	})
}

// Helper function to extract user ID from request (implement based on your auth system)
func getUserIDFromRequest(r *http.Request) string {
	// Extract from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Parse JWT token and extract user ID
	// This is a placeholder - implement based on your auth.Service.ValidateToken
	// For now, return a mock user ID for testing
	// TODO: Integrate with auth.Service.ValidateToken()

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return ""
	}

	// Mock implementation - replace with real JWT validation
	// user, err := authService.ValidateToken(tokenString)
	// if err != nil {
	// 	return ""
	// }
	// return user.ID

	return "demo-user" // Placeholder
}

// Helper function to set CORS headers
// DEPRECATED: Use middleware.NewCORSMiddleware instead
// This function is kept for backward compatibility but should not be used with wildcard in production
func setCORSHeaders(w http.ResponseWriter) {
	// SECURITY WARNING: Wildcard CORS is insecure
	// In production, use config.CORS.AllowedOrigins with specific domains
	origin := os.Getenv("ALLOWED_ORIGINS")
	if origin == "" {
		origin = "http://localhost:3000" // Safe default for development
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// RegisterRoutes registers all workspace routes
func (h *WorkspaceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/workspace/save", h.HandleSaveWorkspace)
	mux.HandleFunc("/api/workspace/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			h.HandleDeleteWorkspace(w, r)
		} else {
			h.HandleLoadWorkspace(w, r)
		}
	})
	mux.HandleFunc("/api/workspaces", h.HandleListWorkspaces)
}
