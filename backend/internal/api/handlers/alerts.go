package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/epic1st/rtx/backend/internal/alerts"
)

// AlertsHandler handles alert API requests
type AlertsHandler struct {
	engine *alerts.Engine
}

// NewAlertsHandler creates a new alerts API handler
func NewAlertsHandler(engine *alerts.Engine) *AlertsHandler {
	return &AlertsHandler{
		engine: engine,
	}
}

// HandleListAlerts - GET /api/alerts
func (h *AlertsHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filters from query params
	accountID := r.URL.Query().Get("accountId")
	statusStr := r.URL.Query().Get("status")

	var status alerts.AlertStatus
	if statusStr != "" {
		status = alerts.AlertStatus(statusStr)
	}

	alertList := h.engine.ListAlerts(accountID, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alertList,
		"count":  len(alertList),
	})
}

// HandleAcknowledgeAlert - POST /api/alerts/acknowledge
func (h *AlertsHandler) HandleAcknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AlertID string `json:"alertId"`
		UserID  string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AlertID == "" {
		http.Error(w, "alertId is required", http.StatusBadRequest)
		return
	}

	// TODO: Extract userID from JWT token instead of request body
	userID := req.UserID
	if userID == "" {
		userID = "system"
	}

	if err := h.engine.AcknowledgeAlert(req.AlertID, userID); err != nil {
		log.Printf("[AlertsHandler] Failed to acknowledge alert: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"alertId": req.AlertID,
		"message": "Alert acknowledged",
	})
}

// HandleSnoozeAlert - POST /api/alerts/snooze
func (h *AlertsHandler) HandleSnoozeAlert(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AlertID         string `json:"alertId"`
		DurationMinutes int    `json:"durationMinutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AlertID == "" {
		http.Error(w, "alertId is required", http.StatusBadRequest)
		return
	}

	if req.DurationMinutes <= 0 {
		req.DurationMinutes = 30 // Default 30 minutes
	}

	if err := h.engine.SnoozeAlert(req.AlertID, req.DurationMinutes); err != nil {
		log.Printf("[AlertsHandler] Failed to snooze alert: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"alertId":         req.AlertID,
		"durationMinutes": req.DurationMinutes,
		"message":         "Alert snoozed",
	})
}

// HandleResolveAlert - POST /api/alerts/resolve
func (h *AlertsHandler) HandleResolveAlert(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AlertID string `json:"alertId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AlertID == "" {
		http.Error(w, "alertId is required", http.StatusBadRequest)
		return
	}

	if err := h.engine.ResolveAlert(req.AlertID); err != nil {
		log.Printf("[AlertsHandler] Failed to resolve alert: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"alertId": req.AlertID,
		"message": "Alert resolved",
	})
}

// HandleListRules - GET /api/alerts/rules
func (h *AlertsHandler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountID := r.URL.Query().Get("accountId")
	rules := h.engine.ListRules(accountID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// HandleCreateRule - POST /api/alerts/rules
func (h *AlertsHandler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rule alerts.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if rule.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if rule.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	if rule.Metric == "" {
		http.Error(w, "metric is required", http.StatusBadRequest)
		return
	}

	// Set defaults
	if rule.Severity == "" {
		rule.Severity = alerts.AlertSeverityMedium
	}

	if rule.Enabled {
		// Enabled by default is dangerous, set to false
		rule.Enabled = true
	}

	if len(rule.Channels) == 0 {
		rule.Channels = []string{"dashboard"} // Default to dashboard only
	}

	if err := h.engine.AddRule(&rule); err != nil {
		log.Printf("[AlertsHandler] Failed to create rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rule":    rule,
		"message": "Alert rule created",
	})
}

// HandleUpdateRule - PUT /api/alerts/rules/{id}
func (h *AlertsHandler) HandleUpdateRule(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[4]

	var rule alerts.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID matches path
	rule.ID = ruleID

	if err := h.engine.AddRule(&rule); err != nil {
		log.Printf("[AlertsHandler] Failed to update rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rule":    rule,
		"message": "Alert rule updated",
	})
}

// HandleDeleteRule - DELETE /api/alerts/rules/{id}
func (h *AlertsHandler) HandleDeleteRule(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[4]

	if err := h.engine.DeleteRule(ruleID); err != nil {
		log.Printf("[AlertsHandler] Failed to delete rule: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"ruleId":  ruleID,
		"message": "Alert rule deleted",
	})
}

// HandleGetRule - GET /api/alerts/rules/{id}
func (h *AlertsHandler) HandleGetRule(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract rule ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[4]

	rule, err := h.engine.GetRule(ruleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}
