package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// RoleStore manages user role assignments
type RoleStore struct {
	mu        sync.RWMutex
	userRoles map[int64]AdminRole // userID -> role
}

// NewRoleStore creates a new role store
func NewRoleStore() *RoleStore {
	return &RoleStore{
		userRoles: make(map[int64]AdminRole),
	}
}

// SetRole assigns a role to a user
func (rs *RoleStore) SetRole(userID int64, role AdminRole) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.userRoles[userID] = role
}

// GetRole retrieves a user's role
func (rs *RoleStore) GetRole(userID int64) (AdminRole, bool) {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	role, exists := rs.userRoles[userID]
	return role, exists
}

// RemoveRole removes a user's role assignment
func (rs *RoleStore) RemoveRole(userID int64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	delete(rs.userRoles, userID)
}

// GetAllRoleAssignments returns all user role assignments
func (rs *RoleStore) GetAllRoleAssignments() map[int64]AdminRole {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	result := make(map[int64]AdminRole, len(rs.userRoles))
	for userID, role := range rs.userRoles {
		result[userID] = role
	}
	return result
}

// RolesHandler handles role management endpoints
type RolesHandler struct {
	authService *AuthService
	roleStore   *RoleStore
}

// NewRolesHandler creates a new roles handler
func NewRolesHandler(authService *AuthService) *RolesHandler {
	return &RolesHandler{
		authService: authService,
		roleStore:   NewRoleStore(),
	}
}

// HandleGetAllRoles returns all available roles and their permissions
// GET /admin/roles
func (h *RolesHandler) HandleGetAllRoles(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate - only SUPER_ADMIN can view roles structure
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if admin.Role != RoleSuperAdmin {
		respondError(w, "Insufficient permissions - SUPER_ADMIN required", http.StatusForbidden)
		return
	}

	// Build response with all roles and their permissions
	type RoleInfo struct {
		Role        AdminRole    `json:"role"`
		Level       int          `json:"level"`
		Permissions []Permission `json:"permissions"`
	}

	roles := []RoleInfo{}
	roleHierarchy := GetAllRoles()

	for role, level := range roleHierarchy {
		permissions := GetPermissions(role)
		roles = append(roles, RoleInfo{
			Role:        role,
			Level:       level,
			Permissions: permissions,
		})
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"roles":   roles,
	})
}

// HandleGetUserRole returns a specific user's role
// GET /admin/users/:id/role
func (h *RolesHandler) HandleGetUserRole(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check permission
	if !HasPermission(admin.Role, PermUsersRead) {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	// Parse user ID from query parameter
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		respondError(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	// Get role from store
	role, exists := h.roleStore.GetRole(userID)
	if !exists {
		// Default role if not set
		role = RoleAdmin
	}

	respondJSON(w, map[string]interface{}{
		"success":     true,
		"userId":      userID,
		"role":        role,
		"permissions": GetPermissions(role),
	})
}

// HandleUpdateUserRole updates a user's role
// PUT /admin/users/:id/role
func (h *RolesHandler) HandleUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate - only SUPER_ADMIN can change roles
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if admin.Role != RoleSuperAdmin {
		respondError(w, "Insufficient permissions - SUPER_ADMIN required", http.StatusForbidden)
		return
	}

	// Parse request body
	var req struct {
		UserID int64     `json:"userId"`
		Role   AdminRole `json:"role"`
		Reason string    `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role
	if _, exists := roleHierarchy[req.Role]; !exists {
		respondError(w, "Invalid role specified", http.StatusBadRequest)
		return
	}

	// Store the role assignment
	h.roleStore.SetRole(req.UserID, req.Role)

	log.Printf("[RBAC] Role updated: User %d -> %s by %s (reason: %s)",
		req.UserID, req.Role, admin.Username, req.Reason)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"userId":  req.UserID,
		"role":    req.Role,
	})
}

// authenticate is a helper to authenticate requests
func (h *RolesHandler) authenticate(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	// Extract Bearer token
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, http.ErrNoCookie
	}

	sessionID := authHeader[7:]
	ipAddress := getIPAddress(r)

	admin, err := h.authService.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// RegisterRoutes registers role management routes
func (h *RolesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/roles", h.HandleGetAllRoles)
	mux.HandleFunc("/admin/users/role", h.HandleGetUserRole)
	mux.HandleFunc("/admin/users/role/update", h.HandleUpdateUserRole)

	log.Println("[RBAC] Role management routes registered")
}
