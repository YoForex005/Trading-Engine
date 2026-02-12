package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GroupsHandler provides HTTP handlers for trading group operations
type GroupsHandler struct {
	groupMgmt  *GroupManagementService
	userMgmt   *UserManagementService
	authSvc    *AuthService
}

// NewGroupsHandler creates a new groups handler
func NewGroupsHandler(groupMgmt *GroupManagementService, userMgmt *UserManagementService, authSvc *AuthService) *GroupsHandler {
	return &GroupsHandler{
		groupMgmt: groupMgmt,
		userMgmt:  userMgmt,
		authSvc:   authSvc,
	}
}

// ListGroups handles GET /admin/groups - returns all trading groups
func (gh *GroupsHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[GroupsHandler] Admin %s listing groups", admin.Username)

	groups := gh.groupMgmt.ListGroups()
	respondJSON(w, map[string]interface{}{
		"success": true,
		"groups":  groups,
		"count":   len(groups),
	})
}

// GetGroup handles GET /admin/groups/:id - returns a specific group
func (gh *GroupsHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from path
	groupID, err := gh.extractGroupID(r)
	if err != nil {
		respondError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	group, err := gh.groupMgmt.GetGroup(groupID)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[GroupsHandler] Admin %s retrieved group %d", admin.Username, groupID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"group":   group,
	})
}

// CreateGroup handles POST /admin/groups - creates a new trading group
func (gh *GroupsHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		ExecutionMode  string   `json:"executionMode"`  // BBOOK, ABOOK, HYBRID
		Markup         float64  `json:"markup"`         // Spread markup in pips
		Commission     float64  `json:"commission"`     // Commission per lot
		MaxLeverage    float64  `json:"maxLeverage"`
		DefaultBalance float64  `json:"defaultBalance"`
		EnabledSymbols []string `json:"enabledSymbols"`
		MarginMode     string   `json:"marginMode"` // HEDGING, NETTING
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, "Group name is required", http.StatusBadRequest)
		return
	}

	if req.ExecutionMode == "" {
		req.ExecutionMode = "BBOOK"
	}

	if req.MarginMode == "" {
		req.MarginMode = "HEDGING"
	}

	if req.MaxLeverage == 0 {
		req.MaxLeverage = 100.0
	}

	if req.DefaultBalance == 0 {
		req.DefaultBalance = 5000.0
	}

	if len(req.EnabledSymbols) == 0 {
		req.EnabledSymbols = []string{"EURUSD", "GBPUSD", "USDJPY"}
	}

	ipAddress := getIPAddress(r)

	group, err := gh.groupMgmt.CreateGroup(
		req.Name,
		req.Description,
		req.ExecutionMode,
		req.Markup,
		req.Commission,
		req.MaxLeverage,
		req.DefaultBalance,
		req.EnabledSymbols,
		req.MarginMode,
		admin,
		ipAddress,
	)

	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[GroupsHandler] Admin %s created group %s (ID: %d)", admin.Username, group.Name, group.ID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"group":   group,
		"message": "Group created successfully",
	})
}

// UpdateGroup handles PUT /admin/groups/:id - updates an existing group
func (gh *GroupsHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from path
	groupID, err := gh.extractGroupID(r)
	if err != nil {
		respondError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name           *string   `json:"name,omitempty"`
		Description    *string   `json:"description,omitempty"`
		ExecutionMode  *string   `json:"executionMode,omitempty"`
		Markup         *float64  `json:"markup,omitempty"`
		Commission     *float64  `json:"commission,omitempty"`
		MaxLeverage    *float64  `json:"maxLeverage,omitempty"`
		DefaultBalance *float64  `json:"defaultBalance,omitempty"`
		EnabledSymbols []string  `json:"enabledSymbols,omitempty"`
		MarginMode     *string   `json:"marginMode,omitempty"`
		Reason         string    `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		req.Reason = "Admin update via API"
	}

	ipAddress := getIPAddress(r)

	err = gh.groupMgmt.UpdateGroup(
		groupID,
		req.Name,
		req.Description,
		req.ExecutionMode,
		req.Markup,
		req.Commission,
		req.MaxLeverage,
		req.DefaultBalance,
		req.EnabledSymbols,
		req.MarginMode,
		admin,
		req.Reason,
		ipAddress,
	)

	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch updated group
	group, _ := gh.groupMgmt.GetGroup(groupID)

	log.Printf("[GroupsHandler] Admin %s updated group %d", admin.Username, groupID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"group":   group,
		"message": "Group updated successfully",
	})
}

// DeleteGroup handles DELETE /admin/groups/:id - deletes a group
func (gh *GroupsHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from path
	groupID, err := gh.extractGroupID(r)
	if err != nil {
		respondError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	// Try to parse body for reason (optional)
	json.NewDecoder(r.Body).Decode(&req)

	if req.Reason == "" {
		req.Reason = "Deleted via API"
	}

	ipAddress := getIPAddress(r)

	err = gh.groupMgmt.DeleteGroup(groupID, admin, req.Reason, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[GroupsHandler] Admin %s deleted group %d", admin.Username, groupID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Group deleted successfully",
	})
}

// ListGroupClients handles GET /admin/groups/:id/clients - lists all clients in a group
func (gh *GroupsHandler) ListGroupClients(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from path
	groupID, err := gh.extractGroupID(r)
	if err != nil {
		respondError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Verify group exists
	_, err = gh.groupMgmt.GetGroup(groupID)
	if err != nil {
		respondError(w, "Group not found", http.StatusNotFound)
		return
	}

	// Get all users and filter by group ID
	allUsers, _ := gh.userMgmt.GetAllUsers()
	groupClients := make([]*UserAccountInfo, 0)

	for _, user := range allUsers {
		if user.GroupID == groupID {
			groupClients = append(groupClients, user)
		}
	}

	log.Printf("[GroupsHandler] Admin %s listed %d clients in group %d", admin.Username, len(groupClients), groupID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"groupId": groupID,
		"clients": groupClients,
		"count":   len(groupClients),
	})
}

// AssignClientToGroup handles POST /admin/groups/:id/clients - assigns a client to a group
func (gh *GroupsHandler) AssignClientToGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from path
	groupID, err := gh.extractGroupID(r)
	if err != nil {
		respondError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Verify group exists
	group, err := gh.groupMgmt.GetGroup(groupID)
	if err != nil {
		respondError(w, "Group not found", http.StatusNotFound)
		return
	}

	var req struct {
		AccountID int64  `json:"accountId"`
		UserID    string `json:"userId"`
		Reason    string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AccountID == 0 && req.UserID == "" {
		respondError(w, "Either accountId or userId is required", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		req.Reason = "Assigned to group via API"
	}

	ipAddress := getIPAddress(r)

	// Use UserManagementService to update the user's group
	// This will be implemented by calling the user management service
	err = gh.userMgmt.AssignUserToGroup(req.AccountID, groupID, admin, req.Reason, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[GroupsHandler] Admin %s assigned account %d to group %s (ID: %d)",
		admin.Username, req.AccountID, group.Name, groupID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Client assigned to group successfully",
		"groupId": groupID,
		"accountId": req.AccountID,
	})
}

// RemoveClientFromGroup handles DELETE /admin/groups/:id/clients/:accountId - removes a client from a group
func (gh *GroupsHandler) RemoveClientFromGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := gh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID and account ID from path
	groupID, err := gh.extractGroupID(r)
	if err != nil {
		respondError(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	accountID, err := gh.extractAccountID(r)
	if err != nil {
		respondError(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	// Try to parse body for reason (optional)
	json.NewDecoder(r.Body).Decode(&req)

	if req.Reason == "" {
		req.Reason = "Removed from group via API"
	}

	ipAddress := getIPAddress(r)

	// Assign to default group (ID: 1 - Standard)
	err = gh.userMgmt.AssignUserToGroup(accountID, 1, admin, req.Reason, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[GroupsHandler] Admin %s removed account %d from group %d",
		admin.Username, accountID, groupID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Client removed from group successfully",
	})
}

// Helper methods

func (gh *GroupsHandler) authenticate(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, http.ErrNoCookie
	}

	sessionID := parts[1]
	ipAddress := getIPAddress(r)

	admin, err := gh.authSvc.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (gh *GroupsHandler) extractGroupID(r *http.Request) (int64, error) {
	// Extract group ID from URL path
	// Expected path: /admin/groups/:id or /admin/groups/:id/clients
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find the index of "groups" and get the next part
	for i, part := range parts {
		if part == "groups" && i+1 < len(parts) {
			idStr := parts[i+1]
			// Skip if it's "clients" (for list endpoint)
			if idStr == "clients" {
				continue
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

	return 0, http.ErrNoCookie
}

func (gh *GroupsHandler) extractAccountID(r *http.Request) (int64, error) {
	// Extract account ID from URL path
	// Expected path: /admin/groups/:groupId/clients/:accountId
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find "clients" and get the next part
	for i, part := range parts {
		if part == "clients" && i+1 < len(parts) {
			idStr := parts[i+1]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

	return 0, http.ErrNoCookie
}
