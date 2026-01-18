package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/cbook"
)

// PaginatedRulesResponse wraps rules with pagination metadata
type PaginatedRulesResponse struct {
	Rules      []*cbook.RoutingRule `json:"rules"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	TotalPages int                  `json:"totalPages"`
}

// CreateRuleRequest represents a request to create a routing rule
type CreateRuleRequest struct {
	Priority        int                          `json:"priority"`
	AccountIDs      []int64                      `json:"accountIds,omitempty"`
	UserGroups      []string                     `json:"userGroups,omitempty"`
	Symbols         []string                     `json:"symbols,omitempty"`
	MinVolume       float64                      `json:"minVolume"`
	MaxVolume       float64                      `json:"maxVolume"`
	Classifications []cbook.ClientClassification `json:"classifications,omitempty"`
	MinToxicity     float64                      `json:"minToxicity"`
	MaxToxicity     float64                      `json:"maxToxicity"`
	Action          cbook.RoutingAction          `json:"action"`
	TargetLP        string                       `json:"targetLp,omitempty"`
	HedgePercent    float64                      `json:"hedgePercent"`
	Description     string                       `json:"description"`
}

// UpdateRuleRequest represents a request to update a routing rule
type UpdateRuleRequest struct {
	Priority        *int                         `json:"priority,omitempty"`
	AccountIDs      []int64                      `json:"accountIds,omitempty"`
	UserGroups      []string                     `json:"userGroups,omitempty"`
	Symbols         []string                     `json:"symbols,omitempty"`
	MinVolume       *float64                     `json:"minVolume,omitempty"`
	MaxVolume       *float64                     `json:"maxVolume,omitempty"`
	Classifications []cbook.ClientClassification `json:"classifications,omitempty"`
	MinToxicity     *float64                     `json:"minToxicity,omitempty"`
	MaxToxicity     *float64                     `json:"maxToxicity,omitempty"`
	Action          *cbook.RoutingAction         `json:"action,omitempty"`
	TargetLP        *string                      `json:"targetLp,omitempty"`
	HedgePercent    *float64                     `json:"hedgePercent,omitempty"`
	Enabled         *bool                        `json:"enabled,omitempty"`
	Description     *string                      `json:"description,omitempty"`
}

// ReorderRulesRequest represents a request to reorder rule priorities
type ReorderRulesRequest struct {
	Rules []struct {
		ID       string `json:"id"`
		Priority int    `json:"priority"`
	} `json:"rules"`
}

// RuleConflict represents a detected conflict between rules
type RuleConflict struct {
	RuleID1 string `json:"ruleId1"`
	RuleID2 string `json:"ruleId2"`
	Reason  string `json:"reason"`
}

// HandleListRoutingRules returns paginated list of all routing rules
func (h *APIHandler) HandleListRoutingRules(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Get routing engine
	routingEngine := h.getRoutingEngine()
	if routingEngine == nil {
		http.Error(w, "Routing engine not available", http.StatusInternalServerError)
		return
	}

	// Get all rules
	allRules := routingEngine.GetRules()

	// Calculate pagination
	total := len(allRules)
	totalPages := (total + pageSize - 1) / pageSize

	// Validate page
	if page > totalPages && total > 0 {
		page = totalPages
	}

	// Get paginated slice
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		start = 0
		end = 0
	}

	if end > total {
		end = total
	}

	var rules []*cbook.RoutingRule
	if total > 0 {
		rules = allRules[start:end]
	} else {
		rules = make([]*cbook.RoutingRule, 0)
	}

	response := PaginatedRulesResponse{
		Rules:      rules,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("[RoutingRulesAPI] Listed rules: page %d/%d (%d total)", page, totalPages, total)
}

// HandleCreateRoutingRule creates a new routing rule
func (h *APIHandler) HandleCreateRoutingRule(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if err := validateCreateRuleRequest(&req); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Generate rule ID
	ruleID := generateRuleID()

	// Create rule
	rule := &cbook.RoutingRule{
		ID:              ruleID,
		Priority:        req.Priority,
		AccountIDs:      req.AccountIDs,
		UserGroups:      req.UserGroups,
		Symbols:         req.Symbols,
		MinVolume:       req.MinVolume,
		MaxVolume:       req.MaxVolume,
		Classifications: req.Classifications,
		MinToxicity:     req.MinToxicity,
		MaxToxicity:     req.MaxToxicity,
		Action:          req.Action,
		TargetLP:        req.TargetLP,
		HedgePercent:    req.HedgePercent,
		Description:     req.Description,
		Enabled:         true,
	}

	// Get routing engine
	routingEngine := h.getRoutingEngine()
	if routingEngine == nil {
		http.Error(w, "Routing engine not available", http.StatusInternalServerError)
		return
	}

	// Check for conflicts
	conflicts := h.detectRuleConflicts(routingEngine, rule)
	if len(conflicts) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":     "Rule conflicts detected",
			"conflicts": conflicts,
			"rule":      rule,
		})
		return
	}

	// Add rule
	routingEngine.AddRule(rule)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"ruleId":  rule.ID,
		"rule":    rule,
	})

	log.Printf("[RoutingRulesAPI] Created rule: %s (priority: %d)", rule.ID, rule.Priority)
}

// HandleUpdateRoutingRule updates an existing routing rule
func (h *APIHandler) HandleUpdateRoutingRule(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rule ID from path: /api/routing/rules/{id}
	ruleID := extractIDFromPath(r.URL.Path, "/api/routing/rules/")
	if ruleID == "" {
		http.Error(w, "Rule ID not found in path", http.StatusBadRequest)
		return
	}

	var req UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get routing engine
	routingEngine := h.getRoutingEngine()
	if routingEngine == nil {
		http.Error(w, "Routing engine not available", http.StatusInternalServerError)
		return
	}

	// Get existing rule
	existingRule := findRuleByID(routingEngine.GetRules(), ruleID)
	if existingRule == nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	// Update rule with new values
	updated := *existingRule

	if req.Priority != nil {
		updated.Priority = *req.Priority
	}
	if len(req.AccountIDs) > 0 {
		updated.AccountIDs = req.AccountIDs
	}
	if len(req.UserGroups) > 0 {
		updated.UserGroups = req.UserGroups
	}
	if len(req.Symbols) > 0 {
		updated.Symbols = req.Symbols
	}
	if req.MinVolume != nil {
		updated.MinVolume = *req.MinVolume
	}
	if req.MaxVolume != nil {
		updated.MaxVolume = *req.MaxVolume
	}
	if len(req.Classifications) > 0 {
		updated.Classifications = req.Classifications
	}
	if req.MinToxicity != nil {
		updated.MinToxicity = *req.MinToxicity
	}
	if req.MaxToxicity != nil {
		updated.MaxToxicity = *req.MaxToxicity
	}
	if req.Action != nil {
		updated.Action = *req.Action
	}
	if req.TargetLP != nil {
		updated.TargetLP = *req.TargetLP
	}
	if req.HedgePercent != nil {
		updated.HedgePercent = *req.HedgePercent
	}
	if req.Enabled != nil {
		updated.Enabled = *req.Enabled
	}
	if req.Description != nil {
		updated.Description = *req.Description
	}

	// Check for conflicts with updated rule
	conflicts := h.detectRuleConflicts(routingEngine, &updated)
	if len(conflicts) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":     "Rule conflicts detected",
			"conflicts": conflicts,
			"rule":      &updated,
		})
		return
	}

	// Update rule
	if err := routingEngine.UpdateRule(ruleID, &updated); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update rule: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rule":    &updated,
	})

	log.Printf("[RoutingRulesAPI] Updated rule: %s", ruleID)
}

// HandleDeleteRoutingRule deletes a routing rule
func (h *APIHandler) HandleDeleteRoutingRule(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rule ID from path
	ruleID := extractIDFromPath(r.URL.Path, "/api/routing/rules/")
	if ruleID == "" {
		http.Error(w, "Rule ID not found in path", http.StatusBadRequest)
		return
	}

	// Get routing engine
	routingEngine := h.getRoutingEngine()
	if routingEngine == nil {
		http.Error(w, "Routing engine not available", http.StatusInternalServerError)
		return
	}

	// Delete rule
	if err := routingEngine.DeleteRule(ruleID); err != nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Rule %s deleted", ruleID),
	})

	log.Printf("[RoutingRulesAPI] Deleted rule: %s", ruleID)
}

// HandleReorderRoutingRules bulk updates rule priorities
func (h *APIHandler) HandleReorderRoutingRules(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify admin authentication
	if !h.isAdminUser(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ReorderRulesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Rules) == 0 {
		http.Error(w, "No rules provided", http.StatusBadRequest)
		return
	}

	// Get routing engine
	routingEngine := h.getRoutingEngine()
	if routingEngine == nil {
		http.Error(w, "Routing engine not available", http.StatusInternalServerError)
		return
	}

	// Update each rule's priority
	for _, item := range req.Rules {
		rule := findRuleByID(routingEngine.GetRules(), item.ID)
		if rule == nil {
			http.Error(w, fmt.Sprintf("Rule not found: %s", item.ID), http.StatusNotFound)
			return
		}

		updated := *rule
		updated.Priority = item.Priority

		if err := routingEngine.UpdateRule(item.ID, &updated); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update rule %s: %v", item.ID, err), http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Updated %d rule priorities", len(req.Rules)),
	})

	log.Printf("[RoutingRulesAPI] Reordered %d rules", len(req.Rules))
}

// Helper functions

// isAdminUser checks if the request contains valid admin credentials
func (h *APIHandler) isAdminUser(r *http.Request) bool {
	// Check Authorization header for admin token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// Extract bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	token := parts[1]

	// In production, validate JWT token
	// For now, accept any non-empty token
	// TODO: Implement proper JWT validation
	return token != ""
}

// getRoutingEngine returns the routing engine from the APIHandler
func (h *APIHandler) getRoutingEngine() *cbook.RoutingEngine {
	if h.cbookEngine == nil {
		return nil
	}
	return h.cbookEngine.GetRoutingEngine()
}

// detectRuleConflicts identifies potential conflicts between rules
func (h *APIHandler) detectRuleConflicts(routingEngine *cbook.RoutingEngine, newRule *cbook.RoutingRule) []RuleConflict {
	var conflicts []RuleConflict

	existingRules := routingEngine.GetRules()

	for _, existing := range existingRules {
		if existing.ID == newRule.ID {
			continue // Skip comparing rule with itself
		}

		// Check if rules have overlapping conditions
		if rulesOverlap(existing, newRule) {
			// Check if they have conflicting actions
			if rulesConflict(existing, newRule) {
				conflicts = append(conflicts, RuleConflict{
					RuleID1: existing.ID,
					RuleID2: newRule.ID,
					Reason:  fmt.Sprintf("Overlapping conditions with conflicting actions: %v vs %v", existing.Action, newRule.Action),
				})
			}
		}
	}

	return conflicts
}

// rulesOverlap checks if two rules have overlapping filter conditions
func rulesOverlap(rule1, rule2 *cbook.RoutingRule) bool {
	// Rules overlap if:
	// 1. No symbol filter, or symbols match
	// 2. No account filter, or accounts match
	// 3. Volume ranges overlap

	// Check symbol overlap
	if len(rule1.Symbols) > 0 && len(rule2.Symbols) > 0 {
		symbolsMatch := false
		for _, s1 := range rule1.Symbols {
			for _, s2 := range rule2.Symbols {
				if s1 == s2 || s1 == "*" || s2 == "*" {
					symbolsMatch = true
					break
				}
			}
			if symbolsMatch {
				break
			}
		}
		if !symbolsMatch {
			return false
		}
	}

	// Check account overlap
	if len(rule1.AccountIDs) > 0 && len(rule2.AccountIDs) > 0 {
		accountsMatch := false
		for _, a1 := range rule1.AccountIDs {
			for _, a2 := range rule2.AccountIDs {
				if a1 == a2 {
					accountsMatch = true
					break
				}
			}
			if accountsMatch {
				break
			}
		}
		if !accountsMatch {
			return false
		}
	}

	// Check volume range overlap
	minMax1 := rule1.MaxVolume > 0
	minMax2 := rule2.MaxVolume > 0

	if minMax1 && minMax2 {
		// Both have max volumes - check if ranges overlap
		if rule1.MaxVolume < rule2.MinVolume || rule2.MaxVolume < rule1.MinVolume {
			return false
		}
	}

	return true
}

// rulesConflict checks if two overlapping rules have conflicting actions
func rulesConflict(rule1, rule2 *cbook.RoutingRule) bool {
	// Rules conflict if they have different routing actions
	// Example: One rejects while another accepts for same order
	if rule1.Action != rule2.Action {
		return true
	}

	// Same action but different LPs could also be a conflict
	if rule1.Action == cbook.ActionABook && rule1.TargetLP != rule2.TargetLP {
		return true
	}

	return false
}

// generateRuleID generates a unique rule ID
func generateRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}

// extractIDFromPath extracts the ID from a URL path
func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

// findRuleByID finds a rule by its ID
func findRuleByID(rules []*cbook.RoutingRule, id string) *cbook.RoutingRule {
	for _, rule := range rules {
		if rule.ID == id {
			return rule
		}
	}
	return nil
}

// validateCreateRuleRequest validates the create rule request
func validateCreateRuleRequest(req *CreateRuleRequest) error {
	if req.Action == "" {
		return fmt.Errorf("action is required")
	}

	// Validate action is valid
	validActions := map[cbook.RoutingAction]bool{
		cbook.ActionABook:        true,
		cbook.ActionBBook:        true,
		cbook.ActionPartialHedge: true,
		cbook.ActionReject:       true,
	}

	if !validActions[req.Action] {
		return fmt.Errorf("invalid action: %s", req.Action)
	}

	// Validate hedge percent for partial hedge
	if req.Action == cbook.ActionPartialHedge {
		if req.HedgePercent < 0 || req.HedgePercent > 100 {
			return fmt.Errorf("hedgePercent must be between 0 and 100")
		}
	}

	// Validate volume ranges
	if req.MinVolume > 0 && req.MaxVolume > 0 && req.MinVolume > req.MaxVolume {
		return fmt.Errorf("minVolume cannot be greater than maxVolume")
	}

	// Validate toxicity ranges
	if req.MinToxicity > 0 && req.MaxToxicity > 0 && req.MinToxicity > req.MaxToxicity {
		return fmt.Errorf("minToxicity cannot be greater than maxToxicity")
	}

	return nil
}
