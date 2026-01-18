package cbook

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// APIHandlers provides HTTP handlers for the C-Book routing system
type APIHandlers struct {
	engine *CBookEngine
}

// NewAPIHandlers creates API handlers
func NewAPIHandlers(engine *CBookEngine) *APIHandlers {
	return &APIHandlers{
		engine: engine,
	}
}

// RegisterRoutes registers all API routes
func (h *APIHandlers) RegisterRoutes(mux *http.ServeMux) {
	// Client profiling
	mux.HandleFunc("/api/cbook/profiles", h.handleGetProfiles)
	mux.HandleFunc("/api/cbook/profile/", h.handleGetProfile)

	// Routing
	mux.HandleFunc("/api/cbook/route", h.handleRoute)
	mux.HandleFunc("/api/cbook/routing/rules", h.handleRoutingRules)
	mux.HandleFunc("/api/cbook/routing/stats", h.handleRoutingStats)

	// Exposure management
	mux.HandleFunc("/api/cbook/exposure", h.handleExposure)
	mux.HandleFunc("/api/cbook/exposure/limits", h.handleExposureLimits)

	// Analytics
	mux.HandleFunc("/api/cbook/analytics/performance", h.handlePerformance)
	mux.HandleFunc("/api/cbook/analytics/pnl", h.handlePnLReport)
	mux.HandleFunc("/api/cbook/analytics/effectiveness", h.handleEffectiveness)
	mux.HandleFunc("/api/cbook/analytics/recommendations", h.handleRecommendations)

	// ML
	mux.HandleFunc("/api/cbook/ml/stats", h.handleMLStats)
	mux.HandleFunc("/api/cbook/ml/predict", h.handleMLPredict)
	mux.HandleFunc("/api/cbook/ml/export", h.handleMLExport)
	mux.HandleFunc("/api/cbook/ml/import", h.handleMLImport)

	// Compliance
	mux.HandleFunc("/api/cbook/compliance/audit", h.handleAuditLogs)
	mux.HandleFunc("/api/cbook/compliance/alerts", h.handleAlerts)
	mux.HandleFunc("/api/cbook/compliance/best-execution", h.handleBestExecution)
	mux.HandleFunc("/api/cbook/compliance/export", h.handleAuditExport)

	// Dashboard
	mux.HandleFunc("/api/cbook/dashboard", h.handleDashboard)

	// Admin
	mux.HandleFunc("/api/cbook/admin/config", h.handleAdminConfig)
}

// handleGetProfiles returns all client profiles
func (h *APIHandlers) handleGetProfiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	classification := r.URL.Query().Get("classification")

	var profiles []*ClientProfile
	if classification != "" {
		profiles = h.engine.GetProfilesByClassification(ClientClassification(classification))
	} else {
		profiles = h.engine.GetAllProfiles()
	}

	respondJSON(w, profiles)
}

// handleGetProfile returns a specific client profile
func (h *APIHandlers) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountIDStr := r.URL.Query().Get("accountId")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	profile, exists := h.engine.GetClientProfile(accountID)
	if !exists {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	respondJSON(w, profile)
}

// handleRoute makes a routing decision
func (h *APIHandlers) handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AccountID  int64   `json:"accountId"`
		UserID     string  `json:"userId"`
		Username   string  `json:"username"`
		Symbol     string  `json:"symbol"`
		Side       string  `json:"side"`
		Volume     float64 `json:"volume"`
		Volatility float64 `json:"volatility"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	decision, err := h.engine.RouteOrder(
		req.AccountID, req.UserID, req.Username,
		req.Symbol, req.Side, req.Volume, req.Volatility,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, decision)
}

// handleRoutingRules manages routing rules
func (h *APIHandlers) handleRoutingRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rules := h.engine.GetRoutingRules()
		respondJSON(w, rules)

	case http.MethodPost:
		var rule RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		h.engine.AddRoutingRule(&rule)
		respondJSON(w, map[string]string{"status": "created"})

	case http.MethodPut:
		var rule RoutingRule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := h.engine.UpdateRoutingRule(rule.ID, &rule); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		respondJSON(w, map[string]string{"status": "updated"})

	case http.MethodDelete:
		ruleID := r.URL.Query().Get("id")
		if err := h.engine.DeleteRoutingRule(ruleID); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		respondJSON(w, map[string]string{"status": "deleted"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRoutingStats returns routing statistics
func (h *APIHandlers) handleRoutingStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.engine.GetRoutingStats()
	respondJSON(w, stats)
}

// handleExposure returns current exposure
func (h *APIHandlers) handleExposure(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	symbol := r.URL.Query().Get("symbol")

	if symbol != "" {
		exposure := h.engine.GetExposure(symbol)
		respondJSON(w, exposure)
	} else {
		exposures := h.engine.GetAllExposures()
		respondJSON(w, exposures)
	}
}

// handleExposureLimits manages exposure limits
func (h *APIHandlers) handleExposureLimits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var limit ExposureLimit
	if err := json.NewDecoder(r.Body).Decode(&limit); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	h.engine.SetExposureLimit(limit.Symbol, &limit)
	respondJSON(w, map[string]string{"status": "updated"})
}

// handlePerformance returns client performance metrics
func (h *APIHandlers) handlePerformance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountIDStr := r.URL.Query().Get("accountId")
	limitStr := r.URL.Query().Get("limit")

	if accountIDStr != "" {
		accountID, _ := strconv.ParseInt(accountIDStr, 10, 64)
		perf, exists := h.engine.GetClientPerformance(accountID)
		if !exists {
			http.Error(w, "Performance data not found", http.StatusNotFound)
			return
		}
		respondJSON(w, perf)
	} else {
		limit := 10
		if limitStr != "" {
			limit, _ = strconv.Atoi(limitStr)
		}
		performers := h.engine.GetTopPerformers(limit)
		respondJSON(w, performers)
	}
}

// handlePnLReport generates P&L report
func (h *APIHandlers) handlePnLReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	report := h.engine.GeneratePnLReport()
	respondJSON(w, report)
}

// handleEffectiveness returns routing effectiveness metrics
func (h *APIHandlers) handleEffectiveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "1W"
	}

	effectiveness := h.engine.GetRoutingEffectiveness(period)
	respondJSON(w, effectiveness)
}

// handleRecommendations returns optimization recommendations
func (h *APIHandlers) handleRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recommendations := h.engine.GetRecommendedAdjustments()
	respondJSON(w, recommendations)
}

// handleMLStats returns ML model statistics
func (h *APIHandlers) handleMLStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.engine.GetMLStats()
	respondJSON(w, stats)
}

// handleMLPredict generates ML prediction for a client
func (h *APIHandlers) handleMLPredict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountIDStr := r.URL.Query().Get("accountId")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	profile, exists := h.engine.GetClientProfile(accountID)
	if !exists {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	// Access ML predictor through engine
	prediction, err := h.engine.mlPredictor.Predict(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, prediction)
}

// handleMLExport exports ML model
func (h *APIHandlers) handleMLExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := h.engine.ExportMLModel()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=ml_model.json")
	w.Write(data)
}

// handleMLImport imports ML model
func (h *APIHandlers) handleMLImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.engine.ImportMLModel(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]string{"status": "imported"})
}

// handleAuditLogs returns audit logs
func (h *APIHandlers) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountID, _ := strconv.ParseInt(r.URL.Query().Get("accountId"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	logs := h.engine.GetAuditLogs(accountID, time.Time{}, time.Time{}, limit)
	respondJSON(w, logs)
}

// handleAlerts returns compliance alerts
func (h *APIHandlers) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	severity := r.URL.Query().Get("severity")
	resolved := r.URL.Query().Get("resolved") == "true"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 100
	}

	alerts := h.engine.GetComplianceAlerts(severity, resolved, limit)
	respondJSON(w, alerts)
}

// handleBestExecution generates best execution report
func (h *APIHandlers) handleBestExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse time range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var startTime, endTime time.Time
	if startStr != "" {
		startTime, _ = time.Parse(time.RFC3339, startStr)
	} else {
		startTime = time.Now().AddDate(0, 0, -7) // Last 7 days
	}

	if endStr != "" {
		endTime, _ = time.Parse(time.RFC3339, endStr)
	} else {
		endTime = time.Now()
	}

	report := h.engine.GenerateBestExecutionReport(startTime, endTime)
	respondJSON(w, report)
}

// handleAuditExport exports audit trail
func (h *APIHandlers) handleAuditExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startTime := time.Now().AddDate(0, -1, 0) // Last month
	endTime := time.Now()

	data, err := h.engine.ExportAuditTrail(startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=audit_trail.json")
	w.Write(data)
}

// handleDashboard returns comprehensive dashboard data
func (h *APIHandlers) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dashboard := h.engine.GetDashboardData()
	respondJSON(w, dashboard)
}

// handleAdminConfig manages admin configuration
func (h *APIHandlers) handleAdminConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		config := map[string]interface{}{
			"ml_enabled":         h.engine.mlEnabled,
			"auto_learn":         h.engine.autoLearn,
			"strict_compliance":  h.engine.strictCompliance,
		}
		respondJSON(w, config)

	case http.MethodPost:
		var config struct {
			MLEnabled        *bool `json:"mlEnabled"`
			AutoLearn        *bool `json:"autoLearn"`
			StrictCompliance *bool `json:"strictCompliance"`
		}

		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if config.MLEnabled != nil {
			h.engine.EnableML(*config.MLEnabled)
		}
		if config.AutoLearn != nil {
			h.engine.EnableAutoLearning(*config.AutoLearn)
		}
		if config.StrictCompliance != nil {
			h.engine.EnableStrictCompliance(*config.StrictCompliance)
		}

		respondJSON(w, map[string]string{"status": "updated"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Helper function to respond with JSON
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
