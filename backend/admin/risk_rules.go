package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RiskRule represents an automated risk management rule
type RiskRule struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Priority      int            `json:"priority"` // Higher priority = evaluated first
	Condition     RuleCondition  `json:"condition"`
	Action        string         `json:"action"` // Action to take when condition matches
	Status        string         `json:"status"` // active, disabled
	CreatedAt     time.Time      `json:"createdAt"`
	LastTriggered *time.Time     `json:"lastTriggered,omitempty"`
	HitCount      int            `json:"hitCount"` // Number of times rule was triggered
}

// RuleCondition defines when a rule should trigger
type RuleCondition struct {
	Metric   string  `json:"metric"`   // account_equity, margin_level, daily_loss, etc.
	Operator string  `json:"operator"` // gt, lt, gte, lte, eq
	Value    float64 `json:"value"`    // Threshold value
}

// RuleExecution represents a logged rule execution
type RuleExecution struct {
	ID          int64     `json:"id"`
	RuleID      int64     `json:"ruleId"`
	RuleName    string    `json:"ruleName"`
	AccountID   int64     `json:"accountId"`
	Timestamp   time.Time `json:"timestamp"`
	MetricValue float64   `json:"metricValue"`
	ActionTaken string    `json:"actionTaken"`
	Success     bool      `json:"success"`
	Message     string    `json:"message,omitempty"`
}

// AccountMetrics contains risk metrics for evaluation
type AccountMetrics struct {
	AccountEquity  float64 `json:"account_equity"`
	MarginLevel    float64 `json:"margin_level"`
	DailyLoss      float64 `json:"daily_loss"`
	PositionCount  float64 `json:"position_count"`
	PositionSize   float64 `json:"position_size"`
	TotalExposure  float64 `json:"total_exposure"`
	DrawdownPct    float64 `json:"drawdown_pct"`
}

// RiskRulesService manages risk rules and executions
type RiskRulesService struct {
	mu             sync.RWMutex
	rules          map[int64]*RiskRule
	executions     []*RuleExecution
	nextRuleID     int64
	nextExecutionID int64
}

// NewRiskRulesService creates a new risk rules service with pre-configured rules
func NewRiskRulesService() *RiskRulesService {
	svc := &RiskRulesService{
		rules:          make(map[int64]*RiskRule),
		executions:     make([]*RuleExecution, 0),
		nextRuleID:     1,
		nextExecutionID: 1,
	}

	// Initialize with 10+ pre-configured rules
	svc.initializeMockRules()

	return svc
}

// initializeMockRules creates realistic pre-configured risk rules
func (rrs *RiskRulesService) initializeMockRules() {
	mockRules := []struct {
		name     string
		priority int
		metric   string
		operator string
		value    float64
		action   string
		status   string
	}{
		// Critical margin rules (highest priority)
		{"Critical Margin Call", 1, "margin_level", "lt", 50, "force_close", "active"},
		{"Warning Margin Level", 2, "margin_level", "lt", 100, "send_alert", "active"},

		// Equity protection rules
		{"Account Equity Below Minimum", 3, "account_equity", "lt", 1000, "restrict_trading", "active"},
		{"Daily Loss Limit Exceeded", 4, "daily_loss", "gt", 5000, "restrict_trading", "active"},
		{"Severe Daily Loss", 5, "daily_loss", "gt", 10000, "notify_risk_manager", "active"},

		// Position management rules
		{"Max Position Count", 6, "position_count", "gt", 10, "send_alert", "active"},
		{"Excessive Position Size", 7, "position_size", "gt", 100, "send_email", "active"},
		{"Total Exposure Limit", 8, "total_exposure", "gt", 50000, "reduce_leverage", "active"},

		// Drawdown protection
		{"High Drawdown Warning", 9, "drawdown_pct", "gt", 20, "send_alert", "active"},
		{"Critical Drawdown", 10, "drawdown_pct", "gt", 30, "notify_risk_manager", "active"},
		{"Extreme Drawdown", 11, "drawdown_pct", "gt", 40, "force_close", "active"},

		// Additional monitoring rules
		{"Low Margin Level", 12, "margin_level", "lt", 200, "send_email", "active"},
		{"Position Count Threshold", 13, "position_count", "gte", 8, "send_email", "disabled"},
		{"Moderate Daily Loss", 14, "daily_loss", "gt", 2500, "send_alert", "active"},
	}

	now := time.Now()
	for i, mock := range mockRules {
		rule := &RiskRule{
			ID:       int64(i + 1),
			Name:     mock.name,
			Priority: mock.priority,
			Condition: RuleCondition{
				Metric:   mock.metric,
				Operator: mock.operator,
				Value:    mock.value,
			},
			Action:    mock.action,
			Status:    mock.status,
			CreatedAt: now.AddDate(0, 0, -i), // Stagger creation dates
			HitCount:  0,
		}

		rrs.rules[rule.ID] = rule
	}

	rrs.nextRuleID = int64(len(mockRules) + 1)
	log.Printf("[RiskRulesService] Initialized with %d pre-configured risk rules", len(mockRules))
}

// EvaluateRules checks all active rules against provided metrics and executes actions
func (rrs *RiskRulesService) EvaluateRules(accountID int64, metrics AccountMetrics) []*RuleExecution {
	rrs.mu.Lock()
	defer rrs.mu.Unlock()

	executions := make([]*RuleExecution, 0)

	// Convert metrics to map for easy lookup
	metricsMap := map[string]float64{
		"account_equity":  metrics.AccountEquity,
		"margin_level":    metrics.MarginLevel,
		"daily_loss":      metrics.DailyLoss,
		"position_count":  metrics.PositionCount,
		"position_size":   metrics.PositionSize,
		"total_exposure":  metrics.TotalExposure,
		"drawdown_pct":    metrics.DrawdownPct,
	}

	// Get all active rules sorted by priority
	activeRules := rrs.getActiveRulesSortedByPriority()

	now := time.Now()
	for _, rule := range activeRules {
		// Get metric value
		metricValue, ok := metricsMap[rule.Condition.Metric]
		if !ok {
			log.Printf("[RiskRulesService] Unknown metric: %s", rule.Condition.Metric)
			continue
		}

		// Evaluate condition
		if rrs.evaluateCondition(rule.Condition, metricValue) {
			// Rule matched - execute action
			execution := &RuleExecution{
				ID:          rrs.nextExecutionID,
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				AccountID:   accountID,
				Timestamp:   now,
				MetricValue: metricValue,
				ActionTaken: rule.Action,
				Success:     true,
			}

			// Execute the action
			message := rrs.executeAction(rule.Action, accountID, rule.Name, metricValue)
			execution.Message = message

			// Update rule stats
			rule.LastTriggered = &now
			rule.HitCount++

			// Store execution
			rrs.executions = append(rrs.executions, execution)
			executions = append(executions, execution)
			rrs.nextExecutionID++

			log.Printf("[RiskRulesService] Rule #%d '%s' triggered for account %d: %s (metric=%.2f)",
				rule.ID, rule.Name, accountID, rule.Action, metricValue)
		}
	}

	return executions
}

// getActiveRulesSortedByPriority returns active rules sorted by priority (highest first)
func (rrs *RiskRulesService) getActiveRulesSortedByPriority() []*RiskRule {
	activeRules := make([]*RiskRule, 0)
	for _, rule := range rrs.rules {
		if rule.Status == "active" {
			activeRules = append(activeRules, rule)
		}
	}

	// Simple bubble sort by priority (descending)
	for i := 0; i < len(activeRules)-1; i++ {
		for j := 0; j < len(activeRules)-i-1; j++ {
			if activeRules[j].Priority > activeRules[j+1].Priority {
				activeRules[j], activeRules[j+1] = activeRules[j+1], activeRules[j]
			}
		}
	}

	return activeRules
}

// evaluateCondition checks if a condition matches the metric value
func (rrs *RiskRulesService) evaluateCondition(condition RuleCondition, metricValue float64) bool {
	switch condition.Operator {
	case "gt":
		return metricValue > condition.Value
	case "lt":
		return metricValue < condition.Value
	case "gte":
		return metricValue >= condition.Value
	case "lte":
		return metricValue <= condition.Value
	case "eq":
		return metricValue == condition.Value
	default:
		log.Printf("[RiskRulesService] Unknown operator: %s", condition.Operator)
		return false
	}
}

// executeAction performs the specified action
func (rrs *RiskRulesService) executeAction(action string, accountID int64, ruleName string, metricValue float64) string {
	switch action {
	case "send_alert":
		// In production, this would send a real-time alert to the admin dashboard
		return fmt.Sprintf("Alert sent: '%s' triggered for account %d (value: %.2f)", ruleName, accountID, metricValue)

	case "send_email":
		// In production, this would send an email notification
		return fmt.Sprintf("Email notification sent for rule '%s' (account: %d, value: %.2f)", ruleName, accountID, metricValue)

	case "restrict_trading":
		// In production, this would disable trading for the account
		return fmt.Sprintf("Trading restricted for account %d due to '%s'", accountID, ruleName)

	case "force_close":
		// In production, this would close all positions for the account
		return fmt.Sprintf("Force close initiated for account %d (rule: '%s')", accountID, ruleName)

	case "reduce_leverage":
		// In production, this would reduce the account's leverage
		return fmt.Sprintf("Leverage reduced for account %d (rule: '%s')", accountID, ruleName)

	case "notify_risk_manager":
		// In production, this would send notification to risk management team
		return fmt.Sprintf("Risk manager notified: '%s' for account %d (value: %.2f)", ruleName, accountID, metricValue)

	default:
		return fmt.Sprintf("Unknown action: %s", action)
	}
}

// ListRules returns all risk rules
func (rrs *RiskRulesService) ListRules() []*RiskRule {
	rrs.mu.RLock()
	defer rrs.mu.RUnlock()

	rules := make([]*RiskRule, 0, len(rrs.rules))
	for _, rule := range rrs.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetRule returns a single risk rule
func (rrs *RiskRulesService) GetRule(id int64) (*RiskRule, error) {
	rrs.mu.RLock()
	defer rrs.mu.RUnlock()

	rule, ok := rrs.rules[id]
	if !ok {
		return nil, fmt.Errorf("risk rule %d not found", id)
	}
	return rule, nil
}

// CreateRule creates a new risk rule
func (rrs *RiskRulesService) CreateRule(name string, priority int, condition RuleCondition, action, status string) (*RiskRule, error) {
	rrs.mu.Lock()
	defer rrs.mu.Unlock()

	if name == "" {
		return nil, fmt.Errorf("rule name is required")
	}

	// Validate condition
	if err := rrs.validateCondition(condition); err != nil {
		return nil, err
	}

	// Validate action
	if err := rrs.validateAction(action); err != nil {
		return nil, err
	}

	// Validate status
	if status != "active" && status != "disabled" {
		status = "active" // Default
	}

	rule := &RiskRule{
		ID:        rrs.nextRuleID,
		Name:      name,
		Priority:  priority,
		Condition: condition,
		Action:    action,
		Status:    status,
		CreatedAt: time.Now(),
		HitCount:  0,
	}

	rrs.rules[rrs.nextRuleID] = rule
	rrs.nextRuleID++

	log.Printf("[RiskRulesService] Created risk rule #%d: %s", rule.ID, name)
	return rule, nil
}

// UpdateRule updates an existing risk rule
func (rrs *RiskRulesService) UpdateRule(id int64, name *string, priority *int, condition *RuleCondition, action *string, status *string) error {
	rrs.mu.Lock()
	defer rrs.mu.Unlock()

	rule, ok := rrs.rules[id]
	if !ok {
		return fmt.Errorf("risk rule %d not found", id)
	}

	if name != nil {
		rule.Name = *name
	}

	if priority != nil {
		rule.Priority = *priority
	}

	if condition != nil {
		if err := rrs.validateCondition(*condition); err != nil {
			return err
		}
		rule.Condition = *condition
	}

	if action != nil {
		if err := rrs.validateAction(*action); err != nil {
			return err
		}
		rule.Action = *action
	}

	if status != nil {
		if *status != "active" && *status != "disabled" {
			return fmt.Errorf("invalid status: must be 'active' or 'disabled'")
		}
		rule.Status = *status
	}

	log.Printf("[RiskRulesService] Updated risk rule #%d", id)
	return nil
}

// DeleteRule deletes a risk rule
func (rrs *RiskRulesService) DeleteRule(id int64) error {
	rrs.mu.Lock()
	defer rrs.mu.Unlock()

	if _, ok := rrs.rules[id]; !ok {
		return fmt.Errorf("risk rule %d not found", id)
	}

	delete(rrs.rules, id)
	log.Printf("[RiskRulesService] Deleted risk rule #%d", id)
	return nil
}

// GetExecutionHistory returns rule execution history with optional filters
func (rrs *RiskRulesService) GetExecutionHistory(accountID *int64, ruleID *int64, limit int) []*RuleExecution {
	rrs.mu.RLock()
	defer rrs.mu.RUnlock()

	executions := make([]*RuleExecution, 0)

	for _, exec := range rrs.executions {
		// Apply filters
		if accountID != nil && exec.AccountID != *accountID {
			continue
		}
		if ruleID != nil && exec.RuleID != *ruleID {
			continue
		}

		executions = append(executions, exec)

		// Apply limit
		if limit > 0 && len(executions) >= limit {
			break
		}
	}

	return executions
}

// validateCondition validates a rule condition
func (rrs *RiskRulesService) validateCondition(condition RuleCondition) error {
	validMetrics := map[string]bool{
		"account_equity": true, "margin_level": true, "daily_loss": true,
		"position_count": true, "position_size": true, "total_exposure": true, "drawdown_pct": true,
	}

	if !validMetrics[condition.Metric] {
		return fmt.Errorf("invalid metric: %s", condition.Metric)
	}

	validOperators := map[string]bool{"gt": true, "lt": true, "gte": true, "lte": true, "eq": true}
	if !validOperators[condition.Operator] {
		return fmt.Errorf("invalid operator: %s", condition.Operator)
	}

	return nil
}

// validateAction validates a rule action
func (rrs *RiskRulesService) validateAction(action string) error {
	validActions := map[string]bool{
		"send_alert": true, "send_email": true, "restrict_trading": true,
		"force_close": true, "reduce_leverage": true, "notify_risk_manager": true,
	}

	if !validActions[action] {
		return fmt.Errorf("invalid action: %s", action)
	}

	return nil
}

// RiskRulesHandler handles HTTP requests for risk rules management
type RiskRulesHandler struct {
	svc     *RiskRulesService
	authSvc *AuthService
}

// NewRiskRulesHandler creates a new risk rules handler
func NewRiskRulesHandler(svc *RiskRulesService, authSvc *AuthService) *RiskRulesHandler {
	return &RiskRulesHandler{
		svc:     svc,
		authSvc: authSvc,
	}
}

// ListRules handles GET /admin/risk-rules
func (rrh *RiskRulesHandler) ListRules(w http.ResponseWriter, r *http.Request) {
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
	admin, err := rrh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rules := rrh.svc.ListRules()
	log.Printf("[RiskRulesHandler] Admin %s listed %d risk rules", admin.Username, len(rules))

	respondJSON(w, map[string]interface{}{
		"success": true,
		"rules":   rules,
		"count":   len(rules),
	})
}

// CreateRule handles POST /admin/risk-rules
func (rrh *RiskRulesHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
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
	admin, err := rrh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name      string        `json:"name"`
		Priority  int           `json:"priority"`
		Condition RuleCondition `json:"condition"`
		Action    string        `json:"action"`
		Status    string        `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rule, err := rrh.svc.CreateRule(req.Name, req.Priority, req.Condition, req.Action, req.Status)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[RiskRulesHandler] Admin %s created risk rule #%d: %s", admin.Username, rule.ID, req.Name)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"rule":    rule,
		"message": "Risk rule created successfully",
	})
}

// UpdateRule handles PUT /admin/risk-rules/:id
func (rrh *RiskRulesHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
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
	admin, err := rrh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rule ID
	ruleID, err := rrh.extractRuleID(r)
	if err != nil {
		respondError(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name      *string        `json:"name,omitempty"`
		Priority  *int           `json:"priority,omitempty"`
		Condition *RuleCondition `json:"condition,omitempty"`
		Action    *string        `json:"action,omitempty"`
		Status    *string        `json:"status,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := rrh.svc.UpdateRule(ruleID, req.Name, req.Priority, req.Condition, req.Action, req.Status); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch updated rule
	rule, _ := rrh.svc.GetRule(ruleID)

	log.Printf("[RiskRulesHandler] Admin %s updated risk rule #%d", admin.Username, ruleID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"rule":    rule,
		"message": "Risk rule updated successfully",
	})
}

// DeleteRule handles DELETE /admin/risk-rules/:id
func (rrh *RiskRulesHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
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
	admin, err := rrh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract rule ID
	ruleID, err := rrh.extractRuleID(r)
	if err != nil {
		respondError(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	if err := rrh.svc.DeleteRule(ruleID); err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[RiskRulesHandler] Admin %s deleted risk rule #%d", admin.Username, ruleID)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Risk rule deleted successfully",
	})
}

// EvaluateRules handles POST /admin/risk-rules/evaluate
func (rrh *RiskRulesHandler) EvaluateRules(w http.ResponseWriter, r *http.Request) {
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
	admin, err := rrh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		AccountID int64          `json:"accountId"`
		Metrics   AccountMetrics `json:"metrics"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AccountID == 0 {
		respondError(w, "accountId is required", http.StatusBadRequest)
		return
	}

	executions := rrh.svc.EvaluateRules(req.AccountID, req.Metrics)

	log.Printf("[RiskRulesHandler] Admin %s evaluated rules for account %d: %d rules triggered",
		admin.Username, req.AccountID, len(executions))

	respondJSON(w, map[string]interface{}{
		"success":        true,
		"accountId":      req.AccountID,
		"executions":     executions,
		"triggeredCount": len(executions),
	})
}

// GetExecutionHistory handles GET /admin/risk-rules/history
func (rrh *RiskRulesHandler) GetExecutionHistory(w http.ResponseWriter, r *http.Request) {
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
	admin, err := rrh.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var accountID *int64
	if accountIDStr := r.URL.Query().Get("accountId"); accountIDStr != "" {
		if id, err := strconv.ParseInt(accountIDStr, 10, 64); err == nil {
			accountID = &id
		}
	}

	var ruleID *int64
	if ruleIDStr := r.URL.Query().Get("ruleId"); ruleIDStr != "" {
		if id, err := strconv.ParseInt(ruleIDStr, 10, 64); err == nil {
			ruleID = &id
		}
	}

	limit := 100 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	executions := rrh.svc.GetExecutionHistory(accountID, ruleID, limit)

	log.Printf("[RiskRulesHandler] Admin %s retrieved %d rule executions", admin.Username, len(executions))

	respondJSON(w, map[string]interface{}{
		"success":    true,
		"executions": executions,
		"count":      len(executions),
	})
}

// Helper methods

func (rrh *RiskRulesHandler) authenticate(r *http.Request) (*Admin, error) {
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

	admin, err := rrh.authSvc.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (rrh *RiskRulesHandler) extractRuleID(r *http.Request) (int64, error) {
	// Extract rule ID from URL path
	// Expected path: /admin/risk-rules/:id
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find "risk-rules" and get the next part
	for i, part := range parts {
		if part == "risk-rules" && i+1 < len(parts) {
			idStr := parts[i+1]
			// Skip if it's a subresource
			if idStr == "evaluate" || idStr == "history" {
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
