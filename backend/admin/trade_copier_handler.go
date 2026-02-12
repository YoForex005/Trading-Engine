package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// TradeCopierHandler provides HTTP handlers for trade copier management
type TradeCopierHandler struct {
	copierSvc *TradeCopierService
	authSvc   *AuthService
}

// NewTradeCopierHandler creates a new trade copier handler
func NewTradeCopierHandler(copierSvc *TradeCopierService, authSvc *AuthService) *TradeCopierHandler {
	return &TradeCopierHandler{
		copierSvc: copierSvc,
		authSvc:   authSvc,
	}
}

// ListRelations handles GET /admin/trade-copier/relations - returns all copy relations
func (tch *TradeCopierHandler) ListRelations(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[TradeCopierHandler] Admin %s listing copy relations", admin.Username)

	relations := tch.copierSvc.ListRelations()
	respondJSON(w, map[string]interface{}{
		"success":   true,
		"relations": relations,
		"count":     len(relations),
	})
}

// GetRelation handles GET /admin/trade-copier/relations/:id - returns a specific relation
func (tch *TradeCopierHandler) GetRelation(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract relation ID
	relationID, err := tch.extractRelationID(r)
	if err != nil {
		respondError(w, "Invalid relation ID", http.StatusBadRequest)
		return
	}

	relation, err := tch.copierSvc.GetRelation(relationID)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[TradeCopierHandler] Admin %s retrieved relation %d", admin.Username, relationID)

	respondJSON(w, map[string]interface{}{
		"success":  true,
		"relation": relation,
	})
}

// CreateRelation handles POST /admin/trade-copier/relations - creates a new copy relation
func (tch *TradeCopierHandler) CreateRelation(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		MasterAccountID   int64    `json:"masterAccountId"`
		FollowerAccountID int64    `json:"followerAccountId"`
		CopyMode          CopyMode `json:"copyMode"`          // proportional, fixed, inverse
		FixedLotSize      float64  `json:"fixedLotSize"`      // For fixed mode
		LotMultiplier     float64  `json:"lotMultiplier"`     // For proportional mode
		MaxLots           float64  `json:"maxLots"`           // Maximum lots per trade
		InverseCopy       bool     `json:"inverseCopy"`       // Copy in opposite direction
		SkipSymbols       []string `json:"skipSymbols"`       // Symbols to skip
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.MasterAccountID == 0 {
		respondError(w, "masterAccountId is required", http.StatusBadRequest)
		return
	}
	if req.FollowerAccountID == 0 {
		respondError(w, "followerAccountId is required", http.StatusBadRequest)
		return
	}
	if req.CopyMode == "" {
		req.CopyMode = CopyModeProportional // Default
	}

	// Set defaults
	if req.MaxLots == 0 {
		req.MaxLots = 100.0
	}
	if req.CopyMode == CopyModeProportional && req.LotMultiplier == 0 {
		req.LotMultiplier = 1.0
	}

	relation, err := tch.copierSvc.CreateRelation(
		req.MasterAccountID,
		req.FollowerAccountID,
		req.CopyMode,
		req.FixedLotSize,
		req.LotMultiplier,
		req.MaxLots,
		req.InverseCopy,
		req.SkipSymbols,
	)

	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[TradeCopierHandler] Admin %s created copy relation %d (Master=%d, Follower=%d)",
		admin.Username, relation.ID, req.MasterAccountID, req.FollowerAccountID)

	respondJSON(w, map[string]interface{}{
		"success":  true,
		"relation": relation,
		"message":  "Copy relation created successfully",
	})
}

// UpdateRelation handles PUT /admin/trade-copier/relations/:id - updates a copy relation
func (tch *TradeCopierHandler) UpdateRelation(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract relation ID
	relationID, err := tch.extractRelationID(r)
	if err != nil {
		respondError(w, "Invalid relation ID", http.StatusBadRequest)
		return
	}

	var req struct {
		CopyMode      *CopyMode `json:"copyMode,omitempty"`
		FixedLotSize  *float64  `json:"fixedLotSize,omitempty"`
		LotMultiplier *float64  `json:"lotMultiplier,omitempty"`
		MaxLots       *float64  `json:"maxLots,omitempty"`
		InverseCopy   *bool     `json:"inverseCopy,omitempty"`
		SkipSymbols   []string  `json:"skipSymbols,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = tch.copierSvc.UpdateRelation(
		relationID,
		req.CopyMode,
		req.FixedLotSize,
		req.LotMultiplier,
		req.MaxLots,
		req.InverseCopy,
		req.SkipSymbols,
	)

	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch updated relation
	relation, _ := tch.copierSvc.GetRelation(relationID)

	log.Printf("[TradeCopierHandler] Admin %s updated copy relation %d", admin.Username, relationID)

	respondJSON(w, map[string]interface{}{
		"success":  true,
		"relation": relation,
		"message":  "Copy relation updated successfully",
	})
}

// DeleteRelation handles DELETE /admin/trade-copier/relations/:id - deletes a copy relation
func (tch *TradeCopierHandler) DeleteRelation(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract relation ID
	relationID, err := tch.extractRelationID(r)
	if err != nil {
		respondError(w, "Invalid relation ID", http.StatusBadRequest)
		return
	}

	err = tch.copierSvc.DeleteRelation(relationID)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[TradeCopierHandler] Admin %s deleted copy relation %d", admin.Username, relationID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Copy relation deleted successfully",
	})
}

// PauseRelation handles PUT /admin/trade-copier/relations/:id/pause - pauses a copy relation
func (tch *TradeCopierHandler) PauseRelation(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract relation ID
	relationID, err := tch.extractRelationID(r)
	if err != nil {
		respondError(w, "Invalid relation ID", http.StatusBadRequest)
		return
	}

	err = tch.copierSvc.PauseRelation(relationID)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[TradeCopierHandler] Admin %s paused copy relation %d", admin.Username, relationID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Copy relation paused",
	})
}

// ResumeRelation handles PUT /admin/trade-copier/relations/:id/resume - resumes a copy relation
func (tch *TradeCopierHandler) ResumeRelation(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract relation ID
	relationID, err := tch.extractRelationID(r)
	if err != nil {
		respondError(w, "Invalid relation ID", http.StatusBadRequest)
		return
	}

	err = tch.copierSvc.ResumeRelation(relationID)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[TradeCopierHandler] Admin %s resumed copy relation %d", admin.Username, relationID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Copy relation resumed",
	})
}

// GetRelationLog handles GET /admin/trade-copier/relations/:id/log - returns copy log
func (tch *TradeCopierHandler) GetRelationLog(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract relation ID
	relationID, err := tch.extractRelationID(r)
	if err != nil {
		respondError(w, "Invalid relation ID", http.StatusBadRequest)
		return
	}

	// Verify relation exists
	_, err = tch.copierSvc.GetRelation(relationID)
	if err != nil {
		respondError(w, "Copy relation not found", http.StatusNotFound)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logs := tch.copierSvc.GetRelationLogs(relationID, limit)

	log.Printf("[TradeCopierHandler] Admin %s retrieved %d logs for relation %d",
		admin.Username, len(logs), relationID)

	respondJSON(w, map[string]interface{}{
		"success":    true,
		"relationId": relationID,
		"logs":       logs,
		"count":      len(logs),
	})
}

// GetStats handles GET /admin/trade-copier/stats - returns trade copier statistics
func (tch *TradeCopierHandler) GetStats(w http.ResponseWriter, r *http.Request) {
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
	admin, err := tch.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := tch.copierSvc.GetStats()

	log.Printf("[TradeCopierHandler] Admin %s retrieved trade copier stats", admin.Username)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// Helper methods

func (tch *TradeCopierHandler) authenticate(r *http.Request) (*Admin, error) {
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

	admin, err := tch.authSvc.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (tch *TradeCopierHandler) extractRelationID(r *http.Request) (int64, error) {
	// Extract relation ID from URL path
	// Expected paths:
	// - /admin/trade-copier/relations/:id
	// - /admin/trade-copier/relations/:id/pause
	// - /admin/trade-copier/relations/:id/resume
	// - /admin/trade-copier/relations/:id/log
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find "relations" and get the next part
	for i, part := range parts {
		if part == "relations" && i+1 < len(parts) {
			idStr := parts[i+1]
			// Skip if it's a subresource
			if idStr == "log" || idStr == "pause" || idStr == "resume" || idStr == "stats" {
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
