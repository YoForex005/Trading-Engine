package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/models"
)

// UserPreferencesHandler handles user preference-related HTTP requests
type UserPreferencesHandler struct {
	db *sql.DB
}

// NewUserPreferencesHandler creates a new user preferences handler
func NewUserPreferencesHandler(db *sql.DB) *UserPreferencesHandler {
	return &UserPreferencesHandler{db: db}
}

// HandleSavePrintPreferences handles POST /api/user/print-preferences
func (h *UserPreferencesHandler) HandleSavePrintPreferences(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from JWT token
	userID := getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.PrintPreferences
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[UserPreferences] Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	validPageSizes := map[string]bool{"A4": true, "Letter": true, "Legal": true}
	if !validPageSizes[req.PageSize] {
		req.PageSize = "A4" // Default to A4
	}

	validOrientations := map[string]bool{"portrait": true, "landscape": true}
	if !validOrientations[req.Orientation] {
		req.Orientation = "portrait" // Default to portrait
	}

	// Set defaults for margins if not provided
	if req.MarginTop == 0 {
		req.MarginTop = 0.5
	}
	if req.MarginBottom == 0 {
		req.MarginBottom = 0.5
	}
	if req.MarginLeft == 0 {
		req.MarginLeft = 0.5
	}
	if req.MarginRight == 0 {
		req.MarginRight = 0.5
	}

	// Validate margins (must be positive and reasonable)
	if req.MarginTop < 0 || req.MarginTop > 5 ||
		req.MarginBottom < 0 || req.MarginBottom > 5 ||
		req.MarginLeft < 0 || req.MarginLeft > 5 ||
		req.MarginRight < 0 || req.MarginRight > 5 {
		http.Error(w, "Invalid margin values (must be between 0 and 5 inches)", http.StatusBadRequest)
		return
	}

	// Upsert print preferences
	_, err := h.db.Exec(`
		INSERT INTO print_preferences (
			user_id, page_size, orientation, margin_top, margin_bottom,
			margin_left, margin_right, include_header, include_footer,
			header_text, footer_text, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET
			page_size = EXCLUDED.page_size,
			orientation = EXCLUDED.orientation,
			margin_top = EXCLUDED.margin_top,
			margin_bottom = EXCLUDED.margin_bottom,
			margin_left = EXCLUDED.margin_left,
			margin_right = EXCLUDED.margin_right,
			include_header = EXCLUDED.include_header,
			include_footer = EXCLUDED.include_footer,
			header_text = EXCLUDED.header_text,
			footer_text = EXCLUDED.footer_text,
			updated_at = CURRENT_TIMESTAMP
	`, userID, req.PageSize, req.Orientation, req.MarginTop, req.MarginBottom,
		req.MarginLeft, req.MarginRight, req.IncludeHeader, req.IncludeFooter,
		req.HeaderText, req.FooterText)

	if err != nil {
		log.Printf("[UserPreferences] Failed to save print preferences: %v", err)
		http.Error(w, "Failed to save print preferences", http.StatusInternalServerError)
		return
	}

	log.Printf("[UserPreferences] Saved print preferences for user %s", userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Print preferences saved successfully",
	})
}

// HandleGetPrintPreferences handles GET /api/user/print-preferences
func (h *UserPreferencesHandler) HandleGetPrintPreferences(w http.ResponseWriter, r *http.Request) {
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

	var prefs models.PrintPreferences
	err := h.db.QueryRow(`
		SELECT user_id, page_size, orientation, margin_top, margin_bottom,
		       margin_left, margin_right, include_header, include_footer,
		       header_text, footer_text, created_at, updated_at
		FROM print_preferences
		WHERE user_id = $1
	`, userID).Scan(
		&prefs.UserID,
		&prefs.PageSize,
		&prefs.Orientation,
		&prefs.MarginTop,
		&prefs.MarginBottom,
		&prefs.MarginLeft,
		&prefs.MarginRight,
		&prefs.IncludeHeader,
		&prefs.IncludeFooter,
		&prefs.HeaderText,
		&prefs.FooterText,
		&prefs.CreatedAt,
		&prefs.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Return default preferences
		prefs = models.PrintPreferences{
			UserID:        userID,
			PageSize:      "A4",
			Orientation:   "portrait",
			MarginTop:     0.5,
			MarginBottom:  0.5,
			MarginLeft:    0.5,
			MarginRight:   0.5,
			IncludeHeader: true,
			IncludeFooter: true,
		}
	} else if err != nil {
		log.Printf("[UserPreferences] Failed to get print preferences: %v", err)
		http.Error(w, "Failed to get print preferences", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// HandleGetUserDataFolder handles GET /api/user/data-folder
func (h *UserPreferencesHandler) HandleGetUserDataFolder(w http.ResponseWriter, r *http.Request) {
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

	// Determine if running in web or desktop mode
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		platform = "web" // Default to web
	}

	if platform == "desktop" {
		// For desktop, return OS-specific path
		dataPath := getDesktopDataPath(userID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"platform": "desktop",
			"path":     dataPath,
			"os":       runtime.GOOS,
		})
		return
	}

	// For web, return virtual file system structure
	virtualFS := buildVirtualFileSystem(userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"platform": "web",
		"root":     virtualFS,
	})
}

// HandleCheckUnsavedChanges handles GET /api/session/unsaved-changes
func (h *UserPreferencesHandler) HandleCheckUnsavedChanges(w http.ResponseWriter, r *http.Request) {
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

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	var hasUnsavedData bool
	var unsavedWorkspaceJSON []byte

	err := h.db.QueryRow(`
		SELECT has_unsaved_data, unsaved_workspace
		FROM user_sessions
		WHERE id = $1 AND user_id = $2
	`, sessionID, userID).Scan(&hasUnsavedData, &unsavedWorkspaceJSON)

	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hasUnsavedData": false,
		})
		return
	}

	if err != nil {
		log.Printf("[Session] Failed to check unsaved changes: %v", err)
		http.Error(w, "Failed to check unsaved changes", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"hasUnsavedData": hasUnsavedData,
	}

	if hasUnsavedData && unsavedWorkspaceJSON != nil {
		var workspace models.Workspace
		if err := json.Unmarshal(unsavedWorkspaceJSON, &workspace); err == nil {
			response["workspace"] = workspace
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleTerminateSession handles DELETE /api/session
func (h *UserPreferencesHandler) HandleTerminateSession(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		SessionID string `json:"sessionId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check for unsaved changes
	var hasUnsavedData bool
	var workspaceName sql.NullString
	var unsavedWorkspaceJSON []byte

	err := h.db.QueryRow(`
		SELECT has_unsaved_data, unsaved_workspace
		FROM user_sessions
		WHERE id = $1 AND user_id = $2
	`, req.SessionID, userID).Scan(&hasUnsavedData, &unsavedWorkspaceJSON)

	if err != nil && err != sql.ErrNoRows {
		log.Printf("[Session] Failed to check session: %v", err)
		http.Error(w, "Failed to check session", http.StatusInternalServerError)
		return
	}

	unsavedWorkspaces := []string{}
	if hasUnsavedData && unsavedWorkspaceJSON != nil {
		var workspace models.Workspace
		if err := json.Unmarshal(unsavedWorkspaceJSON, &workspace); err == nil {
			workspaceName.String = workspace.Name
			workspaceName.Valid = true
			unsavedWorkspaces = append(unsavedWorkspaces, workspace.Name)
		}
	}

	// Delete session
	_, err = h.db.Exec(`
		DELETE FROM user_sessions
		WHERE id = $1 AND user_id = $2
	`, req.SessionID, userID)

	if err != nil {
		log.Printf("[Session] Failed to delete session: %v", err)
		http.Error(w, "Failed to terminate session", http.StatusInternalServerError)
		return
	}

	log.Printf("[Session] Terminated session %s for user %s", req.SessionID, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":           true,
		"message":           "Session terminated successfully",
		"hadUnsavedData":    hasUnsavedData,
		"unsavedWorkspaces": unsavedWorkspaces,
	})
}

// Helper functions

func getDesktopDataPath(userID string) string {
	// SECURITY: Sanitize userID to prevent path traversal attacks
	sanitized, err := sanitizeUserID(userID)
	if err != nil {
		log.Printf("[SECURITY] Path traversal attempt detected: userID=%s", userID)
		return ""
	}

	var basePath string

	switch runtime.GOOS {
	case "windows":
		basePath = filepath.Join(os.Getenv("APPDATA"), "TradingEngine", "Users", sanitized)
	case "darwin":
		basePath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "TradingEngine", "Users", sanitized)
	case "linux":
		basePath = filepath.Join(os.Getenv("HOME"), ".config", "TradingEngine", "Users", sanitized)
	default:
		basePath = filepath.Join(".", "data", "users", sanitized)
	}

	return basePath
}

// sanitizeUserID prevents path traversal attacks
func sanitizeUserID(userID string) (string, error) {
	// Remove leading/trailing whitespace
	userID = strings.TrimSpace(userID)

	// Reject empty userID
	if userID == "" {
		return "", errors.New("empty userID")
	}

	// Clean the path (resolves . and ..)
	cleaned := filepath.Clean(userID)

	// Reject any path that still contains ".."
	if strings.Contains(cleaned, "..") {
		return "", errors.New("path traversal detected: contains '..'")
	}

	// Reject absolute paths (starting with / or drive letter)
	if filepath.IsAbs(cleaned) {
		return "", errors.New("absolute paths not allowed")
	}

	// Reject paths with path separators (only allow simple identifiers)
	if strings.Contains(cleaned, string(filepath.Separator)) {
		return "", errors.New("path separators not allowed in userID")
	}

	// Limit length to prevent buffer overflow attacks
	if len(cleaned) > 255 {
		return "", errors.New("userID too long (max 255 characters)")
	}

	// Only allow alphanumeric, underscore, hyphen
	for _, c := range cleaned {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return "", errors.New("userID contains invalid characters (only alphanumeric, underscore, hyphen allowed)")
		}
	}

	return cleaned, nil
}

func buildVirtualFileSystem(userID string) models.UserDataFolder {
	now := time.Now()

	return models.UserDataFolder{
		Path:     "/",
		Type:     "directory",
		Modified: now,
		Children: []models.UserDataFolder{
			{
				Path:     "/workspaces",
				Type:     "directory",
				Modified: now,
				Children: []models.UserDataFolder{},
			},
			{
				Path:     "/templates",
				Type:     "directory",
				Modified: now,
				Children: []models.UserDataFolder{},
			},
			{
				Path:     "/exports",
				Type:     "directory",
				Modified: now,
				Children: []models.UserDataFolder{},
			},
			{
				Path:     "/drawings",
				Type:     "directory",
				Modified: now,
				Children: []models.UserDataFolder{},
			},
			{
				Path:     "/indicators",
				Type:     "directory",
				Modified: now,
				Children: []models.UserDataFolder{},
			},
		},
	}
}

// RegisterRoutes registers all user preferences routes
func (h *UserPreferencesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/user/print-preferences", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			h.HandleSavePrintPreferences(w, r)
		} else if r.Method == "GET" {
			h.HandleGetPrintPreferences(w, r)
		} else if r.Method == "OPTIONS" {
			setCORSHeaders(w)
			w.WriteHeader(http.StatusOK)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/user/data-folder", h.HandleGetUserDataFolder)
	mux.HandleFunc("/api/session/unsaved-changes", h.HandleCheckUnsavedChanges)
	mux.HandleFunc("/api/session", h.HandleTerminateSession)
}
