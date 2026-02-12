package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================
// Data Structures
// ============================================

// SurveillanceAlert represents a market abuse alert
type SurveillanceAlert struct {
	ID          int64           `json:"id"`
	Type        string          `json:"type"` // "wash_trading", "spoofing", "layering", "front_running", "insider_trading", "manipulation", "excessive_cancellation"
	Severity    string          `json:"severity"` // "critical", "high", "medium", "low"
	ClientID    int64           `json:"client_id"`
	ClientName  string          `json:"client_name"`
	Symbol      string          `json:"symbol"`
	Description string          `json:"description"`
	DetectedAt  time.Time       `json:"detected_at"`
	Status      string          `json:"status"` // "open", "investigating", "escalated", "resolved", "false_positive"
	AssignedTo  string          `json:"assigned_to,omitempty"`
	Evidence    []TradeEvidence `json:"evidence"`
	Resolution  string          `json:"resolution,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// SurveillanceRule represents a surveillance detection rule
type SurveillanceRule struct {
	ID            int64                  `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	Enabled       bool                   `json:"enabled"`
	Parameters    map[string]interface{} `json:"parameters"` // threshold, timeWindowMin, minTradeCount, etc.
	Severity      string                 `json:"severity"`
	CreatedAt     time.Time              `json:"created_at"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
	TriggerCount  int                    `json:"trigger_count"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// TradeEvidence represents evidence of suspicious trading activity
type TradeEvidence struct {
	TradeID    int64     `json:"trade_id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"` // "buy", "sell"
	Volume     float64   `json:"volume"`
	Price      float64   `json:"price"`
	Timestamp  time.Time `json:"timestamp"`
	FlagReason string    `json:"flag_reason"`
}

// SurveillanceStats represents surveillance dashboard statistics
type SurveillanceStats struct {
	TotalAlerts             int                    `json:"total_alerts"`
	OpenAlerts              int                    `json:"open_alerts"`
	ResolvedToday           int                    `json:"resolved_today"`
	FalsePositiveRate       float64                `json:"false_positive_rate"` // percentage
	AvgResolutionTimeHours  float64                `json:"avg_resolution_time_hours"`
	AlertsByType            map[string]int         `json:"alerts_by_type"`
	AlertsBySeverity        map[string]int         `json:"alerts_by_severity"`
	TopFlaggedClients       []ClientAlertCount     `json:"top_flagged_clients"`
	AlertTrend30Days        []DailyAlertCount      `json:"alert_trend_30_days"`
}

// ClientAlertCount represents alert count per client
type ClientAlertCount struct {
	ClientID   int64  `json:"client_id"`
	ClientName string `json:"client_name"`
	AlertCount int    `json:"alert_count"`
	OpenCount  int    `json:"open_count"`
}

// DailyAlertCount represents alert count per day
type DailyAlertCount struct {
	Date       string `json:"date"`
	AlertCount int    `json:"alert_count"`
}

// ClientSurveillanceProfile represents surveillance profile for a client
type ClientSurveillanceProfile struct {
	ClientID      int64               `json:"client_id"`
	ClientName    string              `json:"client_name"`
	RiskLevel     string              `json:"risk_level"` // "low", "medium", "high", "critical"
	TotalAlerts   int                 `json:"total_alerts"`
	OpenAlerts    int                 `json:"open_alerts"`
	RecentAlerts  []SurveillanceAlert `json:"recent_alerts"`
	AlertsByType  map[string]int      `json:"alerts_by_type"`
	WatchlistStatus string            `json:"watchlist_status"` // "none", "monitoring", "restricted"
	Notes         string              `json:"notes,omitempty"`
}

// UpdateAlertRequest represents request to update an alert
type UpdateAlertRequest struct {
	Status     string `json:"status,omitempty"`
	AssignedTo string `json:"assigned_to,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// UpdateRuleRequest represents request to update a rule
type UpdateRuleRequest struct {
	Enabled    *bool                  `json:"enabled,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Severity   string                 `json:"severity,omitempty"`
}

// ManualScanRequest represents request for manual surveillance scan
type ManualScanRequest struct {
	Symbol   string `json:"symbol,omitempty"`
	ClientID int64  `json:"client_id,omitempty"`
}

// ============================================
// Service
// ============================================

// TradeSurveillanceService manages trade surveillance and market abuse detection
type TradeSurveillanceService struct {
	mu          sync.RWMutex
	alerts      map[int64]*SurveillanceAlert
	rules       map[int64]*SurveillanceRule
	nextAlertID int64
	nextRuleID  int64
}

// NewTradeSurveillanceService creates a new trade surveillance service
func NewTradeSurveillanceService() *TradeSurveillanceService {
	s := &TradeSurveillanceService{
		alerts:      make(map[int64]*SurveillanceAlert),
		rules:       make(map[int64]*SurveillanceRule),
		nextAlertID: 101,
		nextRuleID:  9,
	}

	// Initialize 8 surveillance rules
	s.initRules()

	// Generate 100 alerts across 30 clients with 500 trade evidence records
	s.generateAlerts(100, 30)

	log.Printf("[TradeSurveillanceService] Initialized with %d rules, %d alerts",
		len(s.rules), len(s.alerts))

	return s
}

func (s *TradeSurveillanceService) initRules() {
	now := time.Now()

	rules := []*SurveillanceRule{
		{
			ID:          1,
			Name:        "Wash Trading Detector",
			Type:        "wash_trading",
			Description: "Detects wash trades (simultaneous buy and sell orders to inflate volume)",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"time_window_seconds": 60,
				"price_tolerance_pct": 0.1,
				"min_trade_pairs":     3,
			},
			Severity:     "high",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 45,
			UpdatedAt:    now.Add(-7 * 24 * time.Hour),
		},
		{
			ID:          2,
			Name:        "Spoofing Detector",
			Type:        "spoofing",
			Description: "Detects spoofing (placing fake orders to manipulate price then canceling)",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"cancel_threshold_pct": 80,
				"time_window_minutes":  10,
				"min_order_count":      5,
			},
			Severity:     "critical",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 28,
			UpdatedAt:    now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:          3,
			Name:        "Layering Detector",
			Type:        "layering",
			Description: "Detects layering (multiple orders at different prices to manipulate market)",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"min_layer_count":     5,
				"price_spread_pips":   10,
				"time_window_minutes": 15,
			},
			Severity:     "high",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 22,
			UpdatedAt:    now.Add(-10 * 24 * time.Hour),
		},
		{
			ID:          4,
			Name:        "Front Running Detector",
			Type:        "front_running",
			Description: "Detects front running (trading ahead of large client orders using inside information)",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"large_order_threshold": 10.0, // lots
				"time_window_seconds":   30,
				"price_impact_pips":     5,
			},
			Severity:     "critical",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 12,
			UpdatedAt:    now.Add(-15 * 24 * time.Hour),
		},
		{
			ID:          5,
			Name:        "Insider Trading Detector",
			Type:        "insider_trading",
			Description: "Detects potential insider trading (unusual trades before news events)",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"news_lookback_hours": 24,
				"volume_spike_ratio":  3.0,
				"profit_threshold_pct": 5.0,
			},
			Severity:     "critical",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 8,
			UpdatedAt:    now.Add(-20 * 24 * time.Hour),
		},
		{
			ID:          6,
			Name:        "Market Manipulation Detector",
			Type:        "manipulation",
			Description: "Detects general market manipulation patterns",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"price_swing_pct":     2.0,
				"time_window_minutes": 30,
				"volume_threshold":    100.0, // lots
			},
			Severity:     "high",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 35,
			UpdatedAt:    now.Add(-12 * 24 * time.Hour),
		},
		{
			ID:          7,
			Name:        "Excessive Order Cancellation Detector",
			Type:        "excessive_cancellation",
			Description: "Detects excessive order cancellations (order-to-trade ratio abuse)",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"cancel_ratio_threshold": 10.0, // 10:1 cancel-to-fill ratio
				"time_window_hours":      1,
				"min_order_count":        20,
			},
			Severity:     "medium",
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
			TriggerCount: 67,
			UpdatedAt:    now.Add(-3 * 24 * time.Hour),
		},
		{
			ID:          8,
			Name:        "Cross-Market Manipulation Detector",
			Type:        "manipulation",
			Description: "Detects manipulation across correlated instruments",
			Enabled:     false, // disabled by default (advanced rule)
			Parameters: map[string]interface{}{
				"correlation_threshold": 0.8,
				"price_divergence_pct":  1.5,
				"time_window_minutes":   20,
			},
			Severity:     "high",
			CreatedAt:    now.Add(-90 * 24 * time.Hour),
			TriggerCount: 5,
			UpdatedAt:    now.Add(-30 * 24 * time.Hour),
		},
	}

	for _, rule := range rules {
		lastTriggered := now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour)
		rule.LastTriggered = &lastTriggered
		s.rules[rule.ID] = rule
	}
}

func (s *TradeSurveillanceService) generateAlerts(count int, numClients int) {
	now := time.Now()
	alertTypes := []string{"wash_trading", "spoofing", "layering", "front_running", "insider_trading", "manipulation", "excessive_cancellation"}
	_ = []string{"critical", "high", "medium", "low"}
	statuses := []string{"open", "investigating", "escalated", "resolved", "false_positive"}
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "XAUUSD", "BTCUSD", "ETHUSD", "SPX500"}
	analysts := []string{"compliance-officer-1", "compliance-officer-2", "risk-manager", "senior-analyst"}

	for i := 0; i < count; i++ {
		clientID := int64(10001 + rand.Intn(numClients))
		clientName := "Client-" + strconv.FormatInt(clientID, 10)
		alertType := alertTypes[rand.Intn(len(alertTypes))]
		symbol := symbols[rand.Intn(len(symbols))]

		// Determine severity based on type
		severity := "medium"
		switch alertType {
		case "spoofing", "front_running", "insider_trading":
			severity = "critical"
		case "wash_trading", "layering", "manipulation":
			severity = "high"
		case "excessive_cancellation":
			severity = "medium"
		}

		// Generate status (60% resolved/false_positive, 40% open/investigating/escalated)
		status := statuses[rand.Intn(len(statuses))]
		if rand.Float64() < 0.6 {
			status = statuses[3+rand.Intn(2)] // resolved or false_positive
		}

		detectedAt := now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour)

		// Generate trade evidence (3-7 trades per alert)
		evidenceCount := 3 + rand.Intn(5)
		evidence := make([]TradeEvidence, evidenceCount)
		for j := 0; j < evidenceCount; j++ {
			side := "buy"
			if j%2 == 1 {
				side = "sell"
			}

			basePrice := 1.08500
			if symbol == "GBPUSD" {
				basePrice = 1.26500
			} else if symbol == "USDJPY" {
				basePrice = 149.500
			} else if symbol == "XAUUSD" {
				basePrice = 2050.00
			} else if symbol == "BTCUSD" {
				basePrice = 43500.0
			}

			evidence[j] = TradeEvidence{
				TradeID:    int64(100000 + i*10 + j),
				Symbol:     symbol,
				Side:       side,
				Volume:     0.1 + rand.Float64()*9.9, // 0.1-10 lots
				Price:      basePrice + (rand.Float64()-0.5)*basePrice*0.001,
				Timestamp:  detectedAt.Add(time.Duration(j) * 5 * time.Minute),
				FlagReason: s.getFlagReason(alertType),
			}
		}

		assignedTo := ""
		resolution := ""
		if status != "open" {
			assignedTo = analysts[rand.Intn(len(analysts))]
		}
		if status == "resolved" {
			resolution = "Confirmed market abuse. Client warned and account restricted."
		} else if status == "false_positive" {
			resolution = "False positive. Normal trading activity within market conditions."
		}

		alert := &SurveillanceAlert{
			ID:          int64(i + 1),
			Type:        alertType,
			Severity:    severity,
			ClientID:    clientID,
			ClientName:  clientName,
			Symbol:      symbol,
			Description: s.getAlertDescription(alertType, symbol, clientName),
			DetectedAt:  detectedAt,
			Status:      status,
			AssignedTo:  assignedTo,
			Evidence:    evidence,
			Resolution:  resolution,
			UpdatedAt:   detectedAt.Add(time.Duration(rand.Intn(24)) * time.Hour),
		}

		s.alerts[alert.ID] = alert
	}
}

func (s *TradeSurveillanceService) getFlagReason(alertType string) string {
	reasons := map[string]string{
		"wash_trading":           "Simultaneous offsetting trades detected",
		"spoofing":               "Large order placed and canceled without execution",
		"layering":               "Multiple orders at different price levels, rapid cancellation",
		"front_running":          "Trade executed immediately before large client order",
		"insider_trading":        "Unusual trade volume before news event",
		"manipulation":           "Coordinated trades causing abnormal price movement",
		"excessive_cancellation": "Cancel-to-fill ratio exceeds threshold",
	}
	return reasons[alertType]
}

func (s *TradeSurveillanceService) getAlertDescription(alertType string, symbol string, clientName string) string {
	descriptions := map[string]string{
		"wash_trading":           clientName + " executed offsetting buy/sell orders on " + symbol + " within 60 seconds, inflating volume without position change",
		"spoofing":               clientName + " placed large sell orders on " + symbol + " then canceled 95% before execution, manipulating bid-ask spread",
		"layering":               clientName + " placed 8 sequential buy orders on " + symbol + " at different prices, canceled all after price moved up 3 pips",
		"front_running":          clientName + " bought " + symbol + " 15 seconds before large institutional buy order, profited from price impact",
		"insider_trading":        clientName + " placed unusually large " + symbol + " position 6 hours before major economic announcement",
		"manipulation":           clientName + " executed coordinated trades on " + symbol + " causing 2.5% price swing in 15 minutes",
		"excessive_cancellation": clientName + " canceled 87 of 90 orders on " + symbol + " over 1 hour (96.7% cancellation rate)",
	}
	return descriptions[alertType]
}

// GetAlerts returns alerts with optional filters
func (s *TradeSurveillanceService) GetAlerts(alertType string, severity string, status string, clientID int64) []*SurveillanceAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*SurveillanceAlert, 0)
	for _, alert := range s.alerts {
		// Apply filters
		if alertType != "" && alert.Type != alertType {
			continue
		}
		if severity != "" && alert.Severity != severity {
			continue
		}
		if status != "" && alert.Status != status {
			continue
		}
		if clientID > 0 && alert.ClientID != clientID {
			continue
		}

		alerts = append(alerts, alert)
	}

	return alerts
}

// GetAlertByID returns an alert by ID
func (s *TradeSurveillanceService) GetAlertByID(id int64) *SurveillanceAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alerts[id]
}

// UpdateAlert updates an alert
func (s *TradeSurveillanceService) UpdateAlert(id int64, req UpdateAlertRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert := s.alerts[id]
	if alert == nil {
		return nil
	}

	if req.Status != "" {
		alert.Status = req.Status
	}
	if req.AssignedTo != "" {
		alert.AssignedTo = req.AssignedTo
	}
	if req.Resolution != "" {
		alert.Resolution = req.Resolution
	}

	alert.UpdatedAt = time.Now()
	return nil
}

// GetRules returns all surveillance rules
func (s *TradeSurveillanceService) GetRules() []*SurveillanceRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]*SurveillanceRule, 0, len(s.rules))
	for _, rule := range s.rules {
		rules = append(rules, rule)
	}
	return rules
}

// UpdateRule updates a surveillance rule
func (s *TradeSurveillanceService) UpdateRule(id int64, req UpdateRuleRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule := s.rules[id]
	if rule == nil {
		return nil
	}

	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Parameters != nil {
		for k, v := range req.Parameters {
			rule.Parameters[k] = v
		}
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}

	rule.UpdatedAt = time.Now()
	return nil
}

// GetClientProfile returns surveillance profile for a client
func (s *TradeSurveillanceService) GetClientProfile(clientID int64) *ClientSurveillanceProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientName := "Client-" + strconv.FormatInt(clientID, 10)
	totalAlerts := 0
	openAlerts := 0
	recentAlerts := make([]SurveillanceAlert, 0)
	alertsByType := make(map[string]int)

	for _, alert := range s.alerts {
		if alert.ClientID == clientID {
			totalAlerts++
			if alert.Status == "open" || alert.Status == "investigating" {
				openAlerts++
			}
			if len(recentAlerts) < 10 {
				recentAlerts = append(recentAlerts, *alert)
			}
			alertsByType[alert.Type]++
		}
	}

	// Determine risk level based on alerts
	riskLevel := "low"
	if totalAlerts >= 10 {
		riskLevel = "critical"
	} else if totalAlerts >= 5 {
		riskLevel = "high"
	} else if totalAlerts >= 2 {
		riskLevel = "medium"
	}

	watchlistStatus := "none"
	if openAlerts >= 3 {
		watchlistStatus = "restricted"
	} else if totalAlerts >= 5 {
		watchlistStatus = "monitoring"
	}

	return &ClientSurveillanceProfile{
		ClientID:        clientID,
		ClientName:      clientName,
		RiskLevel:       riskLevel,
		TotalAlerts:     totalAlerts,
		OpenAlerts:      openAlerts,
		RecentAlerts:    recentAlerts,
		AlertsByType:    alertsByType,
		WatchlistStatus: watchlistStatus,
		Notes:           "",
	}
}

// TriggerScan triggers a manual surveillance scan
func (s *TradeSurveillanceService) TriggerScan(req ManualScanRequest) map[string]interface{} {
	// Simulate scan (in real implementation, would run actual detection algorithms)
	s.mu.Lock()
	defer s.mu.Unlock()

	scanResult := map[string]interface{}{
		"success":        true,
		"scan_type":      "manual",
		"timestamp":      time.Now(),
		"new_alerts":     rand.Intn(3), // 0-2 new alerts found
		"scanned_trades": 1000 + rand.Intn(4000),
	}

	if req.Symbol != "" {
		scanResult["symbol"] = req.Symbol
	}
	if req.ClientID > 0 {
		scanResult["client_id"] = req.ClientID
	}

	return scanResult
}

// GetStats returns surveillance statistics
func (s *TradeSurveillanceService) GetStats() SurveillanceStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	openAlerts := 0
	resolvedToday := 0
	falsePositives := 0
	totalResolved := 0
	totalResolutionTime := 0.0
	alertsByType := make(map[string]int)
	alertsBySeverity := make(map[string]int)
	clientAlertCount := make(map[int64]int)
	clientOpenCount := make(map[int64]int)

	now := time.Now()
	today := now.Format("2006-01-02")

	for _, alert := range s.alerts {
		if alert.Status == "open" || alert.Status == "investigating" || alert.Status == "escalated" {
			openAlerts++
		}

		if (alert.Status == "resolved" || alert.Status == "false_positive") && alert.UpdatedAt.Format("2006-01-02") == today {
			resolvedToday++
		}

		if alert.Status == "false_positive" {
			falsePositives++
		}

		if alert.Status == "resolved" || alert.Status == "false_positive" {
			totalResolved++
			resolutionTime := alert.UpdatedAt.Sub(alert.DetectedAt).Hours()
			totalResolutionTime += resolutionTime
		}

		alertsByType[alert.Type]++
		alertsBySeverity[alert.Severity]++
		clientAlertCount[alert.ClientID]++
		if alert.Status == "open" || alert.Status == "investigating" {
			clientOpenCount[alert.ClientID]++
		}
	}

	// Calculate false positive rate
	falsePositiveRate := 0.0
	if totalResolved > 0 {
		falsePositiveRate = float64(falsePositives) / float64(totalResolved) * 100
	}

	// Calculate average resolution time
	avgResolutionTime := 0.0
	if totalResolved > 0 {
		avgResolutionTime = totalResolutionTime / float64(totalResolved)
	}

	// Top flagged clients
	topClients := make([]ClientAlertCount, 0)
	for clientID, count := range clientAlertCount {
		if len(topClients) < 10 {
			clientName := "Client-" + strconv.FormatInt(clientID, 10)
			topClients = append(topClients, ClientAlertCount{
				ClientID:   clientID,
				ClientName: clientName,
				AlertCount: count,
				OpenCount:  clientOpenCount[clientID],
			})
		}
	}

	// 30-day alert trend
	alertTrend := make([]DailyAlertCount, 30)
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -29+i)
		alertTrend[i] = DailyAlertCount{
			Date:       date.Format("2006-01-02"),
			AlertCount: 2 + rand.Intn(6), // 2-7 alerts per day
		}
	}

	return SurveillanceStats{
		TotalAlerts:            len(s.alerts),
		OpenAlerts:             openAlerts,
		ResolvedToday:          resolvedToday,
		FalsePositiveRate:      falsePositiveRate,
		AvgResolutionTimeHours: avgResolutionTime,
		AlertsByType:           alertsByType,
		AlertsBySeverity:       alertsBySeverity,
		TopFlaggedClients:      topClients,
		AlertTrend30Days:       alertTrend,
	}
}

// ============================================
// HTTP Handlers
// ============================================

// TradeSurveillanceHandler handles trade surveillance HTTP requests
type TradeSurveillanceHandler struct {
	service     *TradeSurveillanceService
	authService interface {
		ValidateAdminToken(r *http.Request) (int64, error)
	}
}

// NewTradeSurveillanceHandler creates a new trade surveillance handler
func NewTradeSurveillanceHandler(service *TradeSurveillanceService, authService interface {
	ValidateAdminToken(r *http.Request) (int64, error)
}) *TradeSurveillanceHandler {
	return &TradeSurveillanceHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetAlerts handles GET /admin/surveillance/alerts
func (h *TradeSurveillanceHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	alertType := r.URL.Query().Get("type")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")
	clientIDStr := r.URL.Query().Get("client_id")

	var clientID int64
	if clientIDStr != "" {
		clientID, _ = strconv.ParseInt(clientIDStr, 10, 64)
	}

	alerts := h.service.GetAlerts(alertType, severity, status, clientID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(alerts)
}

// HandleGetAlertByID handles GET /admin/surveillance/alerts/:id
func (h *TradeSurveillanceHandler) HandleGetAlertByID(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/surveillance/alerts/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	alert := h.service.GetAlertByID(id)
	if alert == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(alert)
}

// HandleUpdateAlert handles PUT /admin/surveillance/alerts/:id
func (h *TradeSurveillanceHandler) HandleUpdateAlert(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/surveillance/alerts/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	var req UpdateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAlert(id, req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Alert updated successfully",
	})
}

// HandleGetRules handles GET /admin/surveillance/rules
func (h *TradeSurveillanceHandler) HandleGetRules(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rules := h.service.GetRules()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(rules)
}

// HandleUpdateRule handles PUT /admin/surveillance/rules/:id
func (h *TradeSurveillanceHandler) HandleUpdateRule(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/surveillance/rules/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	var req UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateRule(id, req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Rule updated successfully",
	})
}

// HandleGetClientProfile handles GET /admin/surveillance/clients/:id
func (h *TradeSurveillanceHandler) HandleGetClientProfile(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/surveillance/clients/")
	clientID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	profile := h.service.GetClientProfile(clientID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(profile)
}

// HandleTriggerScan handles POST /admin/surveillance/scan
func (h *TradeSurveillanceHandler) HandleTriggerScan(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ManualScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result := h.service.TriggerScan(req)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(result)
}

// HandleGetStats handles GET /admin/surveillance/stats
func (h *TradeSurveillanceHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
