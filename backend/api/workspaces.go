package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Workspace represents a complete trading workspace configuration
type Workspace struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	UserID      string          `json:"userId"`
	AccountID   string          `json:"accountId,omitempty"`
	Charts      json.RawMessage `json:"charts"`
	Layout      json.RawMessage `json:"layout"`
	MarketWatch json.RawMessage `json:"marketWatch"`
	OrderPanel  json.RawMessage `json:"orderPanel"`
	Version     string          `json:"version"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Thumbnail   string          `json:"thumbnail,omitempty"`
	IsDefault   bool            `json:"isDefault"`
	Tags        string          `json:"tags,omitempty"` // JSON array stored as string
}

// WorkspaceMetadata represents workspace listing information
type WorkspaceMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UserID      string    `json:"userId"`
	AccountID   string    `json:"accountId,omitempty"`
	ChartCount  int       `json:"chartCount"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Thumbnail   string    `json:"thumbnail,omitempty"`
	IsDefault   bool      `json:"isDefault"`
	Tags        []string  `json:"tags"`
}

// SaveWorkspaceRequest represents the save workspace payload
type SaveWorkspaceRequest struct {
	Workspace Workspace `json:"workspace"`
	Overwrite bool      `json:"overwrite"`
}

// SaveWorkspaceResponse represents the save response
type SaveWorkspaceResponse struct {
	Success         bool       `json:"success"`
	WorkspaceID     string     `json:"workspaceId,omitempty"`
	Message         string     `json:"message,omitempty"`
	Conflict        bool       `json:"conflict,omitempty"`
	ConflictVersion *Workspace `json:"conflictVersion,omitempty"`
}

// LoadWorkspaceResponse represents the load response
type LoadWorkspaceResponse struct {
	Success   bool       `json:"success"`
	Workspace *Workspace `json:"workspace,omitempty"`
	Message   string     `json:"message,omitempty"`
}

// ListWorkspacesResponse represents the list response
type ListWorkspacesResponse struct {
	Success    bool                 `json:"success"`
	Workspaces []*WorkspaceMetadata `json:"workspaces"`
	Message    string               `json:"message,omitempty"`
}

// DeleteWorkspaceResponse represents the delete response
type DeleteWorkspaceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// HandleSaveWorkspace handles POST /api/workspaces
func (s *Server) HandleSaveWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SaveWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Workspaces] Failed to decode request: %v", err)
		writeJSON(w, http.StatusBadRequest, SaveWorkspaceResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate workspace
	if req.Workspace.Name == "" {
		writeJSON(w, http.StatusBadRequest, SaveWorkspaceResponse{
			Success: false,
			Message: "Workspace name is required",
		})
		return
	}

	if req.Workspace.UserID == "" {
		writeJSON(w, http.StatusBadRequest, SaveWorkspaceResponse{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	// Set timestamps
	now := time.Now()
	if req.Workspace.ID == "" {
		req.Workspace.ID = generateWorkspaceID()
		req.Workspace.CreatedAt = now
	}
	req.Workspace.UpdatedAt = now

	// Check for conflicts if not overwriting
	if !req.Overwrite && req.Workspace.ID != "" {
		existingWorkspace, err := s.getWorkspaceByID(req.Workspace.ID)
		if err == nil && existingWorkspace != nil {
			// Conflict detected
			if existingWorkspace.UpdatedAt.After(req.Workspace.UpdatedAt) {
				writeJSON(w, http.StatusConflict, SaveWorkspaceResponse{
					Success:         false,
					Conflict:        true,
					ConflictVersion: existingWorkspace,
					Message:         "Workspace has been modified by another session",
				})
				return
			}
		}
	}

	// Save workspace with transaction for atomicity
	db, err := s.getDB()
	if err != nil {
		log.Printf("[Workspaces] Database error: %v", err)
		writeJSON(w, http.StatusInternalServerError, SaveWorkspaceResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Printf("[Workspaces] Failed to begin transaction: %v", err)
		writeJSON(w, http.StatusInternalServerError, SaveWorkspaceResponse{
			Success: false,
			Message: "Failed to begin transaction",
		})
		return
	}
	defer tx.Rollback()

	// Upsert workspace
	query := `
		INSERT INTO workspaces (
			id, name, description, user_id, account_id,
			charts, layout, market_watch, order_panel,
			version, created_at, updated_at, thumbnail, is_default, tags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			account_id = excluded.account_id,
			charts = excluded.charts,
			layout = excluded.layout,
			market_watch = excluded.market_watch,
			order_panel = excluded.order_panel,
			version = excluded.version,
			updated_at = excluded.updated_at,
			thumbnail = excluded.thumbnail,
			is_default = excluded.is_default,
			tags = excluded.tags
	`

	_, err = tx.Exec(
		query,
		req.Workspace.ID,
		req.Workspace.Name,
		req.Workspace.Description,
		req.Workspace.UserID,
		req.Workspace.AccountID,
		req.Workspace.Charts,
		req.Workspace.Layout,
		req.Workspace.MarketWatch,
		req.Workspace.OrderPanel,
		req.Workspace.Version,
		req.Workspace.CreatedAt,
		req.Workspace.UpdatedAt,
		req.Workspace.Thumbnail,
		req.Workspace.IsDefault,
		req.Workspace.Tags,
	)

	if err != nil {
		log.Printf("[Workspaces] Failed to save workspace: %v", err)
		writeJSON(w, http.StatusInternalServerError, SaveWorkspaceResponse{
			Success: false,
			Message: "Failed to save workspace",
		})
		return
	}

	// If this workspace is set as default, unset other defaults for this user
	if req.Workspace.IsDefault {
		_, err = tx.Exec(
			"UPDATE workspaces SET is_default = 0 WHERE user_id = ? AND id != ?",
			req.Workspace.UserID,
			req.Workspace.ID,
		)
		if err != nil {
			log.Printf("[Workspaces] Failed to unset other defaults: %v", err)
			// Continue anyway, this is not critical
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[Workspaces] Failed to commit transaction: %v", err)
		writeJSON(w, http.StatusInternalServerError, SaveWorkspaceResponse{
			Success: false,
			Message: "Failed to commit transaction",
		})
		return
	}

	log.Printf("[Workspaces] Successfully saved workspace: %s (user: %s)", req.Workspace.ID, req.Workspace.UserID)

	writeJSON(w, http.StatusOK, SaveWorkspaceResponse{
		Success:     true,
		WorkspaceID: req.Workspace.ID,
		Message:     "Workspace saved successfully",
	})
}

// HandleLoadWorkspace handles GET /api/workspaces/:id
func (s *Server) HandleLoadWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract workspace ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, LoadWorkspaceResponse{
			Success: false,
			Message: "Workspace ID is required",
		})
		return
	}
	workspaceID := parts[3]

	workspace, err := s.getWorkspaceByID(workspaceID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, LoadWorkspaceResponse{
				Success: false,
				Message: "Workspace not found",
			})
			return
		}

		log.Printf("[Workspaces] Failed to load workspace: %v", err)
		writeJSON(w, http.StatusInternalServerError, LoadWorkspaceResponse{
			Success: false,
			Message: "Failed to load workspace",
		})
		return
	}

	log.Printf("[Workspaces] Successfully loaded workspace: %s", workspaceID)

	writeJSON(w, http.StatusOK, LoadWorkspaceResponse{
		Success:   true,
		Workspace: workspace,
	})
}

// HandleListWorkspaces handles GET /api/workspaces
func (s *Server) HandleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("userId")
	accountID := r.URL.Query().Get("accountId")

	db, err := s.getDB()
	if err != nil {
		log.Printf("[Workspaces] Database error: %v", err)
		writeJSON(w, http.StatusInternalServerError, ListWorkspacesResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	query := `
		SELECT id, name, description, user_id, account_id,
		       updated_at, thumbnail, is_default, tags,
		       json_array_length(charts) as chart_count
		FROM workspaces
		WHERE 1=1
	`
	args := []interface{}{}

	if userID != "" {
		query += " AND user_id = ?"
		args = append(args, userID)
	}

	if accountID != "" {
		query += " AND account_id = ?"
		args = append(args, accountID)
	}

	query += " ORDER BY updated_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("[Workspaces] Failed to list workspaces: %v", err)
		writeJSON(w, http.StatusInternalServerError, ListWorkspacesResponse{
			Success: false,
			Message: "Failed to list workspaces",
		})
		return
	}
	defer rows.Close()

	workspaces := []*WorkspaceMetadata{}
	for rows.Next() {
		var ws WorkspaceMetadata
		var tagsJSON sql.NullString

		err := rows.Scan(
			&ws.ID,
			&ws.Name,
			&ws.Description,
			&ws.UserID,
			&ws.AccountID,
			&ws.UpdatedAt,
			&ws.Thumbnail,
			&ws.IsDefault,
			&tagsJSON,
			&ws.ChartCount,
		)

		if err != nil {
			log.Printf("[Workspaces] Failed to scan workspace: %v", err)
			continue
		}

		// Parse tags JSON
		if tagsJSON.Valid {
			var tags []string
			if err := json.Unmarshal([]byte(tagsJSON.String), &tags); err == nil {
				ws.Tags = tags
			}
		}
		if ws.Tags == nil {
			ws.Tags = []string{}
		}

		workspaces = append(workspaces, &ws)
	}

	log.Printf("[Workspaces] Listed %d workspaces", len(workspaces))

	writeJSON(w, http.StatusOK, ListWorkspacesResponse{
		Success:    true,
		Workspaces: workspaces,
	})
}

// HandleDeleteWorkspace handles DELETE /api/workspaces/:id
func (s *Server) HandleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract workspace ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, DeleteWorkspaceResponse{
			Success: false,
			Message: "Workspace ID is required",
		})
		return
	}
	workspaceID := parts[3]

	db, err := s.getDB()
	if err != nil {
		log.Printf("[Workspaces] Database error: %v", err)
		writeJSON(w, http.StatusInternalServerError, DeleteWorkspaceResponse{
			Success: false,
			Message: "Database error",
		})
		return
	}

	result, err := db.Exec("DELETE FROM workspaces WHERE id = ?", workspaceID)
	if err != nil {
		log.Printf("[Workspaces] Failed to delete workspace: %v", err)
		writeJSON(w, http.StatusInternalServerError, DeleteWorkspaceResponse{
			Success: false,
			Message: "Failed to delete workspace",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, DeleteWorkspaceResponse{
			Success: false,
			Message: "Workspace not found",
		})
		return
	}

	log.Printf("[Workspaces] Successfully deleted workspace: %s", workspaceID)

	writeJSON(w, http.StatusOK, DeleteWorkspaceResponse{
		Success: true,
		Message: "Workspace deleted successfully",
	})
}

// Helper function to get workspace by ID
func (s *Server) getWorkspaceByID(id string) (*Workspace, error) {
	db, err := s.getDB()
	if err != nil {
		return nil, err
	}

	var ws Workspace
	query := `
		SELECT id, name, description, user_id, account_id,
		       charts, layout, market_watch, order_panel,
		       version, created_at, updated_at, thumbnail, is_default, tags
		FROM workspaces
		WHERE id = ?
	`

	err = db.QueryRow(query, id).Scan(
		&ws.ID,
		&ws.Name,
		&ws.Description,
		&ws.UserID,
		&ws.AccountID,
		&ws.Charts,
		&ws.Layout,
		&ws.MarketWatch,
		&ws.OrderPanel,
		&ws.Version,
		&ws.CreatedAt,
		&ws.UpdatedAt,
		&ws.Thumbnail,
		&ws.IsDefault,
		&ws.Tags,
	)

	if err != nil {
		return nil, err
	}

	return &ws, nil
}

// Global database connection for workspaces
var workspaceDB *sql.DB
var workspaceDBMutex sync.Mutex

// Helper function to get database connection
func (s *Server) getDB() (*sql.DB, error) {
	workspaceDBMutex.Lock()
	defer workspaceDBMutex.Unlock()

	if workspaceDB != nil {
		return workspaceDB, nil
	}

	// Initialize SQLite database for workspaces
	db, err := sql.Open("sqlite3", "./data/workspaces.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create workspaces table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS workspaces (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		user_id TEXT NOT NULL,
		account_id TEXT,
		charts TEXT,
		layout TEXT,
		market_watch TEXT,
		order_panel TEXT,
		version TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		thumbnail TEXT,
		is_default INTEGER DEFAULT 0,
		tags TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_workspaces_user_id ON workspaces(user_id);
	CREATE INDEX IF NOT EXISTS idx_workspaces_account_id ON workspaces(account_id);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	workspaceDB = db
	return workspaceDB, nil
}

// Generate unique workspace ID
func generateWorkspaceID() string {
	return fmt.Sprintf("ws_%d", time.Now().UnixNano())
}

// Helper to write JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
