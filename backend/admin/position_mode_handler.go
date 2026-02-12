package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/epic1st/rtx/backend/internal/core"
)

// PositionModeHandler handles position mode API endpoints
type PositionModeHandler struct {
	engine  *core.Engine
	authSvc *AuthService
}

// NewPositionModeHandler creates a new position mode handler
func NewPositionModeHandler(engine *core.Engine, authSvc *AuthService) *PositionModeHandler {
	return &PositionModeHandler{
		engine:  engine,
		authSvc: authSvc,
	}
}

// GetAccountPositionMode handles GET /api/accounts/:id/position-mode
func (pmh *PositionModeHandler) GetAccountPositionMode(w http.ResponseWriter, r *http.Request) {
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
	admin, err := pmh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract account ID
	accountID, err := pmh.extractAccountID(r)
	if err != nil {
		respondError(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	// Get account
	account, ok := pmh.engine.GetAccount(accountID)
	if !ok {
		respondError(w, "Account not found", http.StatusNotFound)
		return
	}

	log.Printf("[PositionModeHandler] Admin %s retrieved position mode for account %d", admin.Username, accountID)

	respondJSON(w, map[string]interface{}{
		"success":      true,
		"accountId":    accountID,
		"positionMode": account.PositionMode,
	})
}

// UpdateAccountPositionMode handles PUT /api/accounts/:id/position-mode
func (pmh *PositionModeHandler) UpdateAccountPositionMode(w http.ResponseWriter, r *http.Request) {
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
	admin, err := pmh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract account ID
	accountID, err := pmh.extractAccountID(r)
	if err != nil {
		respondError(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	var req struct {
		PositionMode string `json:"positionMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate position mode
	if err := core.ValidatePositionMode(req.PositionMode); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if account has open positions (cannot change mode with open positions)
	nettingEngine := core.NewNettingEngine(pmh.engine)
	if err := nettingEngine.CanChangePositionMode(accountID); err != nil {
		respondError(w, err.Error(), http.StatusConflict)
		return
	}

	// Update position mode
	if err := pmh.engine.UpdatePositionMode(accountID, req.PositionMode); err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[PositionModeHandler] Admin %s updated position mode for account %d to %s",
		admin.Username, accountID, req.PositionMode)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Position mode updated successfully",
	})
}

// ListPositionModes handles GET /api/position-modes
func (pmh *PositionModeHandler) ListPositionModes(w http.ResponseWriter, r *http.Request) {
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
	admin, err := pmh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[PositionModeHandler] Admin %s listed position modes", admin.Username)

	modes := []map[string]interface{}{
		{
			"value":       string(core.PositionModeHedging),
			"label":       "Hedging",
			"description": "Allows multiple positions per symbol (MT5-style)",
		},
		{
			"value":       string(core.PositionModeNetting),
			"label":       "Netting",
			"description": "Only one aggregated position per symbol (opposite orders net out)",
		},
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"modes":   modes,
	})
}

// Helper methods

func (pmh *PositionModeHandler) authenticate(r *http.Request) (*Admin, error) {
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

	admin, err := pmh.authSvc.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (pmh *PositionModeHandler) extractAccountID(r *http.Request) (int64, error) {
	// Extract account ID from URL path
	// Expected path: /api/accounts/:id/position-mode
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find "accounts" and get the next part
	for i, part := range parts {
		if part == "accounts" && i+1 < len(parts) {
			idStr := parts[i+1]
			// Skip if it's a subresource
			if idStr == "position-mode" {
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
