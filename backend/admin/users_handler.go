package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// AdminUserResponse is the sanitized admin response (no password hash)
type AdminUserResponse struct {
	ID               int64     `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Role             AdminRole `json:"role"`
	Status           string    `json:"status"`
	TwoFactorEnabled bool      `json:"twoFactorEnabled"`
	LastLogin        string    `json:"lastLogin"`
	CreatedAt        string    `json:"createdAt"`
	CreatedBy        string    `json:"createdBy"`
}

// CreateAdminRequest is the request body for creating an admin
type CreateAdminRequest struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     AdminRole `json:"role"`
	Status   string    `json:"status,omitempty"` // Optional, defaults to ACTIVE
}

// UpdateAdminRequest is the request body for updating an admin
type UpdateAdminRequest struct {
	Email    string    `json:"email,omitempty"`
	Role     AdminRole `json:"role,omitempty"`
	Status   string    `json:"status,omitempty"`
	Password string    `json:"password,omitempty"` // Optional password change
}

// HandleListAdmins returns all admin users
func (h *AdminHandler) HandleListAdmins(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check permission (only super admins and broker admins can list admins)
	if admin.Role != RoleSuperAdmin && admin.Role != RoleBrokerAdmin {
		http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
		return
	}

	// Get all admins
	admins := h.authService.ListAdmins()

	// Convert to sanitized response (remove password hashes)
	response := make([]AdminUserResponse, 0, len(admins))
	for _, a := range admins {
		response = append(response, AdminUserResponse{
			ID:               a.ID,
			Username:         a.Username,
			Email:            a.Email,
			Role:             a.Role,
			Status:           a.Status,
			TwoFactorEnabled: a.TwoFactorEnabled,
			LastLogin:        a.LastLogin.Format("2006-01-02 15:04:05"),
			CreatedAt:        a.CreatedAt.Format("2006-01-02 15:04:05"),
			CreatedBy:        a.CreatedBy,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[AdminAPI] Admin %s listed %d admins", admin.Username, len(admins))
}

// HandleCreateAdmin creates a new admin user
func (h *AdminHandler) HandleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check permission (only super admins can create admins)
	if admin.Role != RoleSuperAdmin {
		http.Error(w, "Forbidden: only SUPER_ADMIN can create admins", http.StatusForbidden)
		return
	}

	// Parse request
	var req CreateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Missing required fields: username, email, password", http.StatusBadRequest)
		return
	}

	// Validate role
	validRoles := map[AdminRole]bool{
		RoleSuperAdmin:  true,
		RoleBrokerAdmin: true,
		RoleRiskManager: true,
		RoleDealer:      true,
		RoleAccountant:  true,
		RoleSupport:     true,
		RoleViewer:      true,
	}
	if !validRoles[req.Role] {
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	// Default status to ACTIVE if not provided
	status := req.Status
	if status == "" {
		status = "ACTIVE"
	}

	// Create admin
	newAdmin, err := h.authService.CreateAdmin(req.Username, req.Email, req.Password, req.Role, nil, admin.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update status if not ACTIVE
	if status != "ACTIVE" {
		newAdmin.Status = status
	}

	// Return sanitized response
	response := AdminUserResponse{
		ID:               newAdmin.ID,
		Username:         newAdmin.Username,
		Email:            newAdmin.Email,
		Role:             newAdmin.Role,
		Status:           newAdmin.Status,
		TwoFactorEnabled: newAdmin.TwoFactorEnabled,
		LastLogin:        newAdmin.LastLogin.Format("2006-01-02 15:04:05"),
		CreatedAt:        newAdmin.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:        newAdmin.CreatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Printf("[AdminAPI] Admin %s created new admin: %s (role: %s)", admin.Username, newAdmin.Username, newAdmin.Role)
}

// HandleUpdateAdmin updates an existing admin
func (h *AdminHandler) HandleUpdateAdmin(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check permission (only super admins can update admins)
	if admin.Role != RoleSuperAdmin {
		http.Error(w, "Forbidden: only SUPER_ADMIN can update admins", http.StatusForbidden)
		return
	}

	// Extract admin ID from URL path
	// Expected: /admin/users/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	adminID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid admin ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var req UpdateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate role if provided
	if req.Role != "" {
		validRoles := map[AdminRole]bool{
			RoleSuperAdmin:  true,
			RoleBrokerAdmin: true,
			RoleRiskManager: true,
			RoleDealer:      true,
			RoleAccountant:  true,
			RoleSupport:     true,
			RoleViewer:      true,
		}
		if !validRoles[req.Role] {
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}
	}

	// Update admin
	updatedAdmin, err := h.authService.UpdateAdmin(adminID, req.Email, req.Role, req.Status, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return sanitized response
	response := AdminUserResponse{
		ID:               updatedAdmin.ID,
		Username:         updatedAdmin.Username,
		Email:            updatedAdmin.Email,
		Role:             updatedAdmin.Role,
		Status:           updatedAdmin.Status,
		TwoFactorEnabled: updatedAdmin.TwoFactorEnabled,
		LastLogin:        updatedAdmin.LastLogin.Format("2006-01-02 15:04:05"),
		CreatedAt:        updatedAdmin.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:        updatedAdmin.CreatedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[AdminAPI] Admin %s updated admin ID %d", admin.Username, adminID)
}

// HandleDisableAdmin disables an admin account
func (h *AdminHandler) HandleDisableAdmin(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check permission (only super admins can disable admins)
	if admin.Role != RoleSuperAdmin {
		http.Error(w, "Forbidden: only SUPER_ADMIN can disable admins", http.StatusForbidden)
		return
	}

	// Extract admin ID from URL path
	// Expected: /admin/users/{id}/disable
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	adminID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid admin ID", http.StatusBadRequest)
		return
	}

	// Prevent disabling self
	if adminID == admin.ID {
		http.Error(w, "Cannot disable your own account", http.StatusBadRequest)
		return
	}

	// Disable admin
	if err := h.authService.DisableAdmin(adminID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Admin disabled successfully",
	})

	log.Printf("[AdminAPI] Admin %s disabled admin ID %d", admin.Username, adminID)
}

// HandleEnableAdmin enables an admin account
func (h *AdminHandler) HandleEnableAdmin(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check permission (only super admins can enable admins)
	if admin.Role != RoleSuperAdmin {
		http.Error(w, "Forbidden: only SUPER_ADMIN can enable admins", http.StatusForbidden)
		return
	}

	// Extract admin ID from URL path
	// Expected: /admin/users/{id}/enable
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	adminID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid admin ID", http.StatusBadRequest)
		return
	}

	// Enable admin
	if err := h.authService.EnableAdmin(adminID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Admin enabled successfully",
	})

	log.Printf("[AdminAPI] Admin %s enabled admin ID %d", admin.Username, adminID)
}
