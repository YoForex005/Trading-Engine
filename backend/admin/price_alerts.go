package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Types
// ============================================

type AlertCondition string

const (
	ConditionPriceAbove    AlertCondition = "price_above"
	ConditionPriceBelow    AlertCondition = "price_below"
	ConditionPriceCross    AlertCondition = "price_cross"
	ConditionPercentChange AlertCondition = "percent_change"
	ConditionSpreadAbove   AlertCondition = "spread_above"
)

type NotifyVia string

const (
	NotifyEmail    NotifyVia = "email"
	NotifyPush     NotifyVia = "push"
	NotifySMS      NotifyVia = "sms"
	NotifyPlatform NotifyVia = "platform"
)

// ============================================
// Data Structures
// ============================================

type PriceAlert struct {
	ID             string         `json:"id"`
	ClientID       string         `json:"clientId"`
	ClientName     string         `json:"clientName"`
	Symbol         string         `json:"symbol"`
	Condition      AlertCondition `json:"condition"`
	TargetPrice    float64        `json:"targetPrice"`
	CurrentPrice   float64        `json:"currentPrice"`
	NotifyVia      []NotifyVia    `json:"notifyVia"`
	IsActive       bool           `json:"isActive"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	TriggeredCount int            `json:"triggeredCount"`
	LastTriggered  *time.Time     `json:"lastTriggered,omitempty"`
	Note           string         `json:"note,omitempty"`
}

type TriggeredAlert struct {
	ID               string         `json:"id"`
	AlertID          string         `json:"alertId"`
	ClientID         string         `json:"clientId"`
	ClientName       string         `json:"clientName"`
	Symbol           string         `json:"symbol"`
	Condition        AlertCondition `json:"condition"`
	TargetPrice      float64        `json:"targetPrice"`
	TriggeredPrice   float64        `json:"triggeredPrice"`
	TriggeredAt      time.Time      `json:"triggeredAt"`
	NotifyVia        []NotifyVia    `json:"notifyVia"`
	NotificationSent bool           `json:"notificationSent"`
}

type AlertStats struct {
	TotalAlerts         int                       `json:"totalAlerts"`
	ActiveCount         int                       `json:"activeCount"`
	TriggeredToday      int                       `json:"triggeredToday"`
	MostAlertedSymbols  []SymbolAlertCount        `json:"mostAlertedSymbols"`
	AlertsByCondition   map[string]int            `json:"alertsByCondition"`
	AlertsByNotifyVia   map[string]int            `json:"alertsByNotifyVia"`
}

type SymbolAlertCount struct {
	Symbol string `json:"symbol"`
	Count  int    `json:"count"`
}

// ============================================
// Store
// ============================================

type PriceAlertStore struct {
	mu               sync.RWMutex
	alerts           map[string]*PriceAlert
	triggeredHistory map[string]*TriggeredAlert
}

func NewPriceAlertStore() *PriceAlertStore {
	store := &PriceAlertStore{
		alerts:           make(map[string]*PriceAlert),
		triggeredHistory: make(map[string]*TriggeredAlert),
	}

	// Generate mock data
	store.generateMockData()

	return store
}

func (s *PriceAlertStore) generateMockData() {
	now := time.Now()

	// 30 popular symbols
	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
		"XAUUSD", "XAGUSD", "BTCUSD", "ETHUSD", "WTICOUSD",
		"SPX500USD", "US30USD", "NAS100USD", "UK100GBP", "DE30EUR",
		"EURJPY", "GBPJPY", "EURGBP", "AUDNZD", "NZDUSD",
		"USDCHF", "USDZAR", "USDTRY", "EURCAD", "GBPCAD",
		"AUDCAD", "CADJPY", "CHFJPY", "XPTUSD", "NATGASUSD",
	}

	conditions := []AlertCondition{
		ConditionPriceAbove, ConditionPriceBelow, ConditionPriceCross,
		ConditionPercentChange, ConditionSpreadAbove,
	}

	notifyOptions := [][]NotifyVia{
		{NotifyEmail},
		{NotifyPush},
		{NotifySMS},
		{NotifyPlatform},
		{NotifyEmail, NotifyPush},
		{NotifyEmail, NotifySMS},
		{NotifyPush, NotifyPlatform},
		{NotifyEmail, NotifyPush, NotifyPlatform},
	}

	// 500 alerts across 100 clients
	for i := 1; i <= 500; i++ {
		clientID := fmt.Sprintf("client-%d", (i%100)+1)
		symbol := symbols[rand.Intn(len(symbols))]
		condition := conditions[rand.Intn(len(conditions))]

		// Generate realistic target prices based on symbol
		var targetPrice, currentPrice float64
		switch {
		case strings.Contains(symbol, "USD") && len(symbol) == 6: // Forex
			basePrice := 1.0 + rand.Float64()*0.5
			targetPrice = basePrice + (rand.Float64()-0.5)*0.1
			currentPrice = basePrice + (rand.Float64()-0.5)*0.05
		case strings.HasPrefix(symbol, "XAU"): // Gold
			basePrice := 1800 + rand.Float64()*200
			targetPrice = basePrice + (rand.Float64()-0.5)*50
			currentPrice = basePrice + (rand.Float64()-0.5)*20
		case strings.HasPrefix(symbol, "XAG"): // Silver
			basePrice := 22 + rand.Float64()*3
			targetPrice = basePrice + (rand.Float64()-0.5)*2
			currentPrice = basePrice + (rand.Float64()-0.5)*1
		case strings.HasPrefix(symbol, "BTC"): // Bitcoin
			basePrice := 40000 + rand.Float64()*20000
			targetPrice = basePrice + (rand.Float64()-0.5)*5000
			currentPrice = basePrice + (rand.Float64()-0.5)*2000
		case strings.HasPrefix(symbol, "ETH"): // Ethereum
			basePrice := 2200 + rand.Float64()*800
			targetPrice = basePrice + (rand.Float64()-0.5)*300
			currentPrice = basePrice + (rand.Float64()-0.5)*150
		case strings.Contains(symbol, "WTI") || strings.Contains(symbol, "NATGAS"): // Oil/Gas
			basePrice := 60 + rand.Float64()*30
			targetPrice = basePrice + (rand.Float64()-0.5)*10
			currentPrice = basePrice + (rand.Float64()-0.5)*5
		default: // Indices
			basePrice := 10000 + rand.Float64()*20000
			targetPrice = basePrice + (rand.Float64()-0.5)*1000
			currentPrice = basePrice + (rand.Float64()-0.5)*500
		}

		createdAt := now.Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour)
		isActive := rand.Float64() > 0.2 // 80% active
		triggeredCount := 0
		var lastTriggered *time.Time

		// Some alerts have been triggered
		if rand.Float64() > 0.6 { // 40% have been triggered at least once
			triggeredCount = 1 + rand.Intn(5)
			triggered := now.Add(-time.Duration(rand.Intn(48)) * time.Hour)
			lastTriggered = &triggered
		}

		alert := &PriceAlert{
			ID:             fmt.Sprintf("alert-%d", i),
			ClientID:       clientID,
			ClientName:     fmt.Sprintf("Client %d", (i%100)+1),
			Symbol:         symbol,
			Condition:      condition,
			TargetPrice:    targetPrice,
			CurrentPrice:   currentPrice,
			NotifyVia:      notifyOptions[rand.Intn(len(notifyOptions))],
			IsActive:       isActive,
			CreatedAt:      createdAt,
			UpdatedAt:      createdAt.Add(time.Duration(rand.Intn(30)) * 24 * time.Hour),
			TriggeredCount: triggeredCount,
			LastTriggered:  lastTriggered,
			Note:           "",
		}

		s.alerts[alert.ID] = alert
	}

	// 200 triggered history entries
	alertIDs := []string{}
	for id := range s.alerts {
		alertIDs = append(alertIDs, id)
	}

	for i := 1; i <= 200; i++ {
		if len(alertIDs) == 0 {
			break
		}

		alertID := alertIDs[rand.Intn(len(alertIDs))]
		alert := s.alerts[alertID]

		triggeredAt := now.Add(-time.Duration(rand.Intn(60)) * 24 * time.Hour)
		triggeredPrice := alert.TargetPrice + (rand.Float64()-0.5)*10

		triggered := &TriggeredAlert{
			ID:               fmt.Sprintf("triggered-%d", i),
			AlertID:          alertID,
			ClientID:         alert.ClientID,
			ClientName:       alert.ClientName,
			Symbol:           alert.Symbol,
			Condition:        alert.Condition,
			TargetPrice:      alert.TargetPrice,
			TriggeredPrice:   triggeredPrice,
			TriggeredAt:      triggeredAt,
			NotifyVia:        alert.NotifyVia,
			NotificationSent: rand.Float64() > 0.1, // 90% sent successfully
		}

		s.triggeredHistory[triggered.ID] = triggered
	}

	log.Printf("[PriceAlertStore] Mock data generated: %d alerts, %d triggered history entries",
		len(s.alerts), len(s.triggeredHistory))
}

// ============================================
// Handler
// ============================================

type PriceAlertHandler struct {
	store       *PriceAlertStore
	authService *auth.Service
}

func NewPriceAlertHandler(store *PriceAlertStore, authService *auth.Service) *PriceAlertHandler {
	return &PriceAlertHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// 1. GET /admin/alerts/price — List all price alerts
// ============================================

func (h *PriceAlertHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	// Query filters
	symbolFilter := r.URL.Query().Get("symbol")
	conditionFilter := r.URL.Query().Get("condition")
	activeFilter := r.URL.Query().Get("active")

	alerts := []*PriceAlert{}
	for _, alert := range h.store.alerts {
		if symbolFilter != "" && alert.Symbol != symbolFilter {
			continue
		}
		if conditionFilter != "" && string(alert.Condition) != conditionFilter {
			continue
		}
		if activeFilter == "true" && !alert.IsActive {
			continue
		}
		if activeFilter == "false" && alert.IsActive {
			continue
		}
		alerts = append(alerts, alert)
	}

	// Sort by created date (newest first)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].CreatedAt.After(alerts[j].CreatedAt)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// ============================================
// 2. GET /admin/alerts/price/:clientId — Alerts for specific client
// ============================================

func (h *PriceAlertHandler) HandleGetClientAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract client ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/alerts/price/"), "/")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "triggered" || parts[0] == "stats" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}
	clientID := parts[0]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	alerts := []*PriceAlert{}
	for _, alert := range h.store.alerts {
		if alert.ClientID == clientID {
			alerts = append(alerts, alert)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

// ============================================
// 3. POST /admin/alerts/price — Create alert
// ============================================

func (h *PriceAlertHandler) HandleCreateAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ClientID    string         `json:"clientId"`
		ClientName  string         `json:"clientName"`
		Symbol      string         `json:"symbol"`
		Condition   AlertCondition `json:"condition"`
		TargetPrice float64        `json:"targetPrice"`
		NotifyVia   []NotifyVia    `json:"notifyVia"`
		Note        string         `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	now := time.Now()
	alertID := fmt.Sprintf("alert-%d", len(h.store.alerts)+1)

	alert := &PriceAlert{
		ID:             alertID,
		ClientID:       req.ClientID,
		ClientName:     req.ClientName,
		Symbol:         req.Symbol,
		Condition:      req.Condition,
		TargetPrice:    req.TargetPrice,
		CurrentPrice:   req.TargetPrice, // Would be fetched from market data in production
		NotifyVia:      req.NotifyVia,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
		TriggeredCount: 0,
		Note:           req.Note,
	}

	h.store.alerts[alertID] = alert

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// ============================================
// 4. PUT /admin/alerts/price/:id — Update alert
// ============================================

func (h *PriceAlertHandler) HandleUpdateAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/alerts/price/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}
	alertID := parts[0]

	var req struct {
		TargetPrice *float64       `json:"targetPrice,omitempty"`
		NotifyVia   []NotifyVia    `json:"notifyVia,omitempty"`
		IsActive    *bool          `json:"isActive,omitempty"`
		Note        *string        `json:"note,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	alert, exists := h.store.alerts[alertID]
	if !exists {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.TargetPrice != nil {
		alert.TargetPrice = *req.TargetPrice
	}
	if len(req.NotifyVia) > 0 {
		alert.NotifyVia = req.NotifyVia
	}
	if req.IsActive != nil {
		alert.IsActive = *req.IsActive
	}
	if req.Note != nil {
		alert.Note = *req.Note
	}
	alert.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// ============================================
// 5. DELETE /admin/alerts/price/:id — Delete alert
// ============================================

func (h *PriceAlertHandler) HandleDeleteAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/alerts/price/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}
	alertID := parts[0]

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	if _, exists := h.store.alerts[alertID]; !exists {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	delete(h.store.alerts, alertID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Alert deleted successfully",
	})
}

// ============================================
// 6. POST /admin/alerts/price/:id/trigger — Manually trigger alert
// ============================================

func (h *PriceAlertHandler) HandleTriggerAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	var alertID string
	for i, part := range pathParts {
		if part == "price" && i+1 < len(pathParts) {
			alertID = pathParts[i+1]
			break
		}
	}

	if alertID == "" || alertID == "triggered" || alertID == "stats" {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	alert, exists := h.store.alerts[alertID]
	if !exists {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	now := time.Now()
	triggeredID := fmt.Sprintf("triggered-%d", len(h.store.triggeredHistory)+1)

	triggered := &TriggeredAlert{
		ID:               triggeredID,
		AlertID:          alertID,
		ClientID:         alert.ClientID,
		ClientName:       alert.ClientName,
		Symbol:           alert.Symbol,
		Condition:        alert.Condition,
		TargetPrice:      alert.TargetPrice,
		TriggeredPrice:   alert.CurrentPrice,
		TriggeredAt:      now,
		NotifyVia:        alert.NotifyVia,
		NotificationSent: true,
	}

	h.store.triggeredHistory[triggeredID] = triggered
	alert.TriggeredCount++
	alert.LastTriggered = &now

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggered)
}

// ============================================
// 7. GET /admin/alerts/price/triggered — Triggered alert history
// ============================================

func (h *PriceAlertHandler) HandleGetTriggeredHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	triggered := []*TriggeredAlert{}
	for _, t := range h.store.triggeredHistory {
		triggered = append(triggered, t)
	}

	// Sort by triggered date (newest first)
	sort.Slice(triggered, func(i, j int) bool {
		return triggered[i].TriggeredAt.After(triggered[j].TriggeredAt)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"triggered": triggered,
		"total":     len(triggered),
	})
}

// ============================================
// 8. GET /admin/alerts/price/stats — Alert statistics
// ============================================

func (h *PriceAlertHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	totalAlerts := len(h.store.alerts)
	activeCount := 0
	triggeredToday := 0

	symbolCounts := make(map[string]int)
	conditionCounts := make(map[string]int)
	notifyViaCounts := make(map[string]int)

	for _, alert := range h.store.alerts {
		if alert.IsActive {
			activeCount++
		}
		symbolCounts[alert.Symbol]++
		conditionCounts[string(alert.Condition)]++
		for _, via := range alert.NotifyVia {
			notifyViaCounts[string(via)]++
		}
	}

	for _, t := range h.store.triggeredHistory {
		if t.TriggeredAt.After(today) {
			triggeredToday++
		}
	}

	// Get top 10 most alerted symbols
	type symbolCount struct {
		symbol string
		count  int
	}
	symbolList := []symbolCount{}
	for symbol, count := range symbolCounts {
		symbolList = append(symbolList, symbolCount{symbol, count})
	}
	sort.Slice(symbolList, func(i, j int) bool {
		return symbolList[i].count > symbolList[j].count
	})

	mostAlerted := []SymbolAlertCount{}
	limit := 10
	if len(symbolList) < limit {
		limit = len(symbolList)
	}
	for i := 0; i < limit; i++ {
		mostAlerted = append(mostAlerted, SymbolAlertCount{
			Symbol: symbolList[i].symbol,
			Count:  symbolList[i].count,
		})
	}

	stats := AlertStats{
		TotalAlerts:        totalAlerts,
		ActiveCount:        activeCount,
		TriggeredToday:     triggeredToday,
		MostAlertedSymbols: mostAlerted,
		AlertsByCondition:  conditionCounts,
		AlertsByNotifyVia:  notifyViaCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
