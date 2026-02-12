package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ExpertAdvisor represents an automated trading strategy
type ExpertAdvisor struct {
	ID            int64     `json:"id"`
	ClientID      int64     `json:"client_id"`
	ClientName    string    `json:"client_name"`
	Name          string    `json:"name"`
	Strategy      string    `json:"strategy"` // trend_following, mean_reversion, scalping, grid, martingale, breakout, arbitrage, news_trading
	Status        string    `json:"status"`   // running, paused, stopped, error, backtesting
	Symbol        string    `json:"symbol"`
	Timeframe     string    `json:"timeframe"`
	CreatedAt     time.Time `json:"created_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	LastTradeAt   *time.Time `json:"last_trade_at,omitempty"`
	TotalTrades   int       `json:"total_trades"`
	TotalPnL      float64   `json:"total_pnl"`
	CurrentDD     float64   `json:"current_drawdown"`
	MaxDD         float64   `json:"max_drawdown"`
	WinRate       float64   `json:"win_rate"`
	UptimeHours   float64   `json:"uptime_hours"`
	Version       string    `json:"version"`
}

// EAConfig represents EA configuration parameters
type EAConfig struct {
	EAID              int64              `json:"ea_id"`
	Parameters        map[string]string  `json:"parameters"`
	RiskLimits        EARiskLimits       `json:"risk_limits"`
	TradingHours      string             `json:"trading_hours"`
	MaxConcurrentTrades int              `json:"max_concurrent_trades"`
	LotSize           float64            `json:"lot_size"`
	StopLoss          float64            `json:"stop_loss"`
	TakeProfit        float64            `json:"take_profit"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// EARiskLimits defines risk management parameters
type EARiskLimits struct {
	MaxDailyLoss      float64 `json:"max_daily_loss"`
	MaxTotalDrawdown  float64 `json:"max_total_drawdown"`
	MaxPositionSize   float64 `json:"max_position_size"`
	MaxTradesPerDay   int     `json:"max_trades_per_day"`
}

// EAPerformance tracks EA performance metrics
type EAPerformance struct {
	EAID              int64     `json:"ea_id"`
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`
	TotalPnL          float64   `json:"total_pnl"`
	AvgWin            float64   `json:"avg_win"`
	AvgLoss           float64   `json:"avg_loss"`
	ProfitFactor      float64   `json:"profit_factor"`
	CurrentDrawdown   float64   `json:"current_drawdown"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	UptimeHours       float64   `json:"uptime_hours"`
	LastUpdated       time.Time `json:"last_updated"`
}

// EALog represents an EA execution log entry
type EALog struct {
	ID        int64     `json:"id"`
	EAID      int64     `json:"ea_id"`
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"event_type"` // start, stop, trade, error, warning, config_change
	Message   string    `json:"message"`
	TradeID   *int64    `json:"trade_id,omitempty"`
	Severity  string    `json:"severity"` // info, warning, error
}

// EAResourceUsage tracks server resource consumption
type EAResourceUsage struct {
	EAID       int64     `json:"ea_id"`
	CPUPercent float64   `json:"cpu_percent"`
	MemoryMB   float64   `json:"memory_mb"`
	ThreadCount int      `json:"thread_count"`
	Timestamp  time.Time `json:"timestamp"`
}

// EAStats represents platform-wide EA statistics
type EAStats struct {
	TotalEAs           int     `json:"total_eas"`
	RunningEAs         int     `json:"running_eas"`
	PausedEAs          int     `json:"paused_eas"`
	StoppedEAs         int     `json:"stopped_eas"`
	ErrorEAs           int     `json:"error_eas"`
	BacktestingEAs     int     `json:"backtesting_eas"`
	TotalTradesToday   int     `json:"total_trades_today"`
	TotalPnLToday      float64 `json:"total_pnl_today"`
	AvgWinRate         float64 `json:"avg_win_rate"`
	TopStrategy        string  `json:"top_strategy"`
	TotalCPUUsage      float64 `json:"total_cpu_usage"`
	TotalMemoryUsageMB float64 `json:"total_memory_usage_mb"`
}

// AutomatedTradingService manages Expert Advisors
type AutomatedTradingService struct {
	mu             sync.RWMutex
	eas            map[int64]*ExpertAdvisor
	configs        map[int64]*EAConfig
	logs           map[int64][]*EALog
	resourceUsage  map[int64]*EAResourceUsage
	nextEAID       int64
	nextLogID      int64
}

// NewAutomatedTradingService creates a new service with mock data
func NewAutomatedTradingService() *AutomatedTradingService {
	s := &AutomatedTradingService{
		eas:           make(map[int64]*ExpertAdvisor),
		configs:       make(map[int64]*EAConfig),
		logs:          make(map[int64][]*EALog),
		resourceUsage: make(map[int64]*EAResourceUsage),
		nextEAID:      1,
		nextLogID:     1,
	}
	s.generateMockData()
	return s
}

func (s *AutomatedTradingService) generateMockData() {
	strategies := []string{"trend_following", "mean_reversion", "scalping", "grid", "martingale", "breakout", "arbitrage", "news_trading"}
	statuses := []string{"running", "paused", "stopped", "error", "backtesting"}
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCHF", "BTCUSD", "ETHUSD", "XAUUSD"}
	timeframes := []string{"M1", "M5", "M15", "H1", "H4", "D1"}
	clientNames := []string{"Alice Johnson", "Bob Smith", "Charlie Davis", "Diana Wilson", "Edward Brown", "Fiona Clark", "George Lee", "Hannah Martin", "Ian White", "Julia Harris"}

	// Generate 20 EAs across 10 clients
	for i := 0; i < 20; i++ {
		clientID := int64(i%len(clientNames) + 1)
		eaID := s.nextEAID
		s.nextEAID++

		status := statuses[i%len(statuses)]
		strategy := strategies[i%len(strategies)]

		createdAt := time.Now().AddDate(0, 0, -rand.Intn(180))
		var startedAt *time.Time
		var lastTradeAt *time.Time

		if status == "running" || status == "paused" {
			t := createdAt.Add(time.Duration(rand.Intn(24)) * time.Hour)
			startedAt = &t

			if status == "running" {
				lt := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second)
				lastTradeAt = &lt
			}
		}

		totalTrades := 0
		totalPnL := 0.0
		winRate := 0.0
		uptimeHours := 0.0

		if status == "running" || status == "paused" {
			totalTrades = rand.Intn(500) + 50
			totalPnL = (rand.Float64()*20000 - 5000)
			winRate = 0.35 + rand.Float64()*0.4 // 35-75%
			uptimeHours = rand.Float64() * 720 // up to 30 days
		}

		ea := &ExpertAdvisor{
			ID:          eaID,
			ClientID:    clientID,
			ClientName:  clientNames[clientID-1],
			Name:        fmt.Sprintf("%s_%s_EA_%d", strategy, symbols[i%len(symbols)], i+1),
			Strategy:    strategy,
			Status:      status,
			Symbol:      symbols[i%len(symbols)],
			Timeframe:   timeframes[i%len(timeframes)],
			CreatedAt:   createdAt,
			StartedAt:   startedAt,
			LastTradeAt: lastTradeAt,
			TotalTrades: totalTrades,
			TotalPnL:    totalPnL,
			CurrentDD:   rand.Float64() * 1500,
			MaxDD:       rand.Float64() * 3000,
			WinRate:     winRate,
			UptimeHours: uptimeHours,
			Version:     fmt.Sprintf("v%d.%d.%d", rand.Intn(3)+1, rand.Intn(10), rand.Intn(20)),
		}
		s.eas[eaID] = ea

		// Generate config
		params := make(map[string]string)
		params["ma_period"] = fmt.Sprintf("%d", rand.Intn(100)+20)
		params["rsi_period"] = fmt.Sprintf("%d", rand.Intn(20)+10)
		params["atr_multiplier"] = fmt.Sprintf("%.1f", 1.5+rand.Float64()*1.5)

		config := &EAConfig{
			EAID: eaID,
			Parameters: params,
			RiskLimits: EARiskLimits{
				MaxDailyLoss:     500 + rand.Float64()*1500,
				MaxTotalDrawdown: 2000 + rand.Float64()*3000,
				MaxPositionSize:  0.5 + rand.Float64()*2.0,
				MaxTradesPerDay:  10 + rand.Intn(40),
			},
			TradingHours:        "00:00-23:59",
			MaxConcurrentTrades: rand.Intn(5) + 1,
			LotSize:            0.01 + rand.Float64()*0.99,
			StopLoss:           20 + rand.Float64()*80,
			TakeProfit:         30 + rand.Float64()*120,
			UpdatedAt:          time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour),
		}
		s.configs[eaID] = config

		// Generate 30 log entries per EA
		eventTypes := []string{"start", "stop", "trade", "error", "warning", "config_change"}

		s.logs[eaID] = make([]*EALog, 0, 30)
		for j := 0; j < 30; j++ {
			eventType := eventTypes[rand.Intn(len(eventTypes))]
			severity := "info"
			if eventType == "error" {
				severity = "error"
			} else if eventType == "warning" {
				severity = "warning"
			}

			var message string
			var tradeID *int64

			switch eventType {
			case "start":
				message = "EA started successfully"
			case "stop":
				message = "EA stopped by user"
			case "trade":
				message = fmt.Sprintf("Trade executed: %s %s %.2f lots",
					[]string{"BUY", "SELL"}[rand.Intn(2)], ea.Symbol, config.LotSize)
				tid := int64(rand.Intn(100000) + 1000)
				tradeID = &tid
			case "error":
				message = []string{
					"Connection timeout to broker server",
					"Insufficient margin for trade execution",
					"Invalid order parameters",
					"Market closed - cannot execute trade",
				}[rand.Intn(4)]
			case "warning":
				message = []string{
					"High spread detected - delaying trade",
					"Approaching daily loss limit",
					"Low liquidity warning",
				}[rand.Intn(3)]
			case "config_change":
				message = "EA configuration updated"
			}

			logEntry := &EALog{
				ID:        s.nextLogID,
				EAID:      eaID,
				Timestamp: time.Now().Add(-time.Duration(rand.Intn(2160)) * time.Hour), // last 90 days
				EventType: eventType,
				Message:   message,
				TradeID:   tradeID,
				Severity:  severity,
			}
			s.nextLogID++
			s.logs[eaID] = append(s.logs[eaID], logEntry)
		}

		// Generate resource usage
		cpuUsage := 0.5 + rand.Float64()*4.5 // 0.5-5%
		memoryUsage := 50 + rand.Float64()*200 // 50-250 MB

		s.resourceUsage[eaID] = &EAResourceUsage{
			EAID:       eaID,
			CPUPercent: cpuUsage,
			MemoryMB:   memoryUsage,
			ThreadCount: rand.Intn(4) + 2,
			Timestamp:  time.Now(),
		}
	}

	log.Printf("[AutomatedTrading] Generated mock data: %d EAs, %d configs, %d total logs",
		len(s.eas), len(s.configs), len(s.logs)*30)
}

// AutomatedTradingHandler handles HTTP requests for EA management
type AutomatedTradingHandler struct {
	service     *AutomatedTradingService
	authService *AuthService
}

// NewAutomatedTradingHandler creates a new handler
func NewAutomatedTradingHandler(service *AutomatedTradingService, authService *AuthService) *AutomatedTradingHandler {
	return &AutomatedTradingHandler{
		service:     service,
		authService: authService,
	}
}

// HandleListEAs handles GET /admin/automated-trading/eas
func (h *AutomatedTradingHandler) HandleListEAs(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	// Get filter parameters
	statusFilter := r.URL.Query().Get("status")
	strategyFilter := r.URL.Query().Get("strategy")
	clientIDStr := r.URL.Query().Get("client_id")

	var clientIDFilter int64
	if clientIDStr != "" {
		clientIDFilter, _ = strconv.ParseInt(clientIDStr, 10, 64)
	}

	eas := make([]*ExpertAdvisor, 0)
	for _, ea := range h.service.eas {
		// Apply filters
		if statusFilter != "" && ea.Status != statusFilter {
			continue
		}
		if strategyFilter != "" && ea.Strategy != strategyFilter {
			continue
		}
		if clientIDFilter > 0 && ea.ClientID != clientIDFilter {
			continue
		}
		eas = append(eas, ea)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    eas,
		"total":   len(eas),
	})
}

// HandleGetEA handles GET /admin/automated-trading/eas/:id
func (h *AutomatedTradingHandler) HandleGetEA(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract EA ID from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	eaID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	ea, exists := h.service.eas[eaID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA not found", http.StatusNotFound)
		return
	}

	config := h.service.configs[eaID]
	resourceUsage := h.service.resourceUsage[eaID]

	// Get recent logs (last 10)
	logs := h.service.logs[eaID]
	recentLogs := make([]*EALog, 0)
	if len(logs) > 10 {
		recentLogs = logs[len(logs)-10:]
	} else {
		recentLogs = logs
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"data":           ea,
		"config":         config,
		"resource_usage": resourceUsage,
		"recent_logs":    recentLogs,
	})
}

// HandleUpdateEA handles PUT /admin/automated-trading/eas/:id
func (h *AutomatedTradingHandler) HandleUpdateEA(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	eaID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Parameters          map[string]string `json:"parameters"`
		RiskLimits          *EARiskLimits     `json:"risk_limits"`
		TradingHours        string            `json:"trading_hours"`
		MaxConcurrentTrades int               `json:"max_concurrent_trades"`
		LotSize             float64           `json:"lot_size"`
		StopLoss            float64           `json:"stop_loss"`
		TakeProfit          float64           `json:"take_profit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.service.mu.Lock()
	defer h.service.mu.Unlock()

	config, exists := h.service.configs[eaID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA not found", http.StatusNotFound)
		return
	}

	// Update config
	if updateData.Parameters != nil {
		config.Parameters = updateData.Parameters
	}
	if updateData.RiskLimits != nil {
		config.RiskLimits = *updateData.RiskLimits
	}
	if updateData.TradingHours != "" {
		config.TradingHours = updateData.TradingHours
	}
	if updateData.MaxConcurrentTrades > 0 {
		config.MaxConcurrentTrades = updateData.MaxConcurrentTrades
	}
	if updateData.LotSize > 0 {
		config.LotSize = updateData.LotSize
	}
	if updateData.StopLoss > 0 {
		config.StopLoss = updateData.StopLoss
	}
	if updateData.TakeProfit > 0 {
		config.TakeProfit = updateData.TakeProfit
	}
	config.UpdatedAt = time.Now()

	// Add log entry
	logEntry := &EALog{
		ID:        h.service.nextLogID,
		EAID:      eaID,
		Timestamp: time.Now(),
		EventType: "config_change",
		Message:   "EA configuration updated via admin panel",
		Severity:  "info",
	}
	h.service.nextLogID++
	h.service.logs[eaID] = append(h.service.logs[eaID], logEntry)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "EA configuration updated successfully",
		"data":    config,
	})
}

// HandleStartEA handles POST /admin/automated-trading/eas/:id/start
func (h *AutomatedTradingHandler) HandleStartEA(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	eaID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	h.service.mu.Lock()
	defer h.service.mu.Unlock()

	ea, exists := h.service.eas[eaID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA not found", http.StatusNotFound)
		return
	}

	if ea.Status == "running" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA is already running", http.StatusBadRequest)
		return
	}

	// Start EA
	ea.Status = "running"
	now := time.Now()
	ea.StartedAt = &now

	// Add log entry
	logEntry := &EALog{
		ID:        h.service.nextLogID,
		EAID:      eaID,
		Timestamp: time.Now(),
		EventType: "start",
		Message:   "EA started by admin",
		Severity:  "info",
	}
	h.service.nextLogID++
	h.service.logs[eaID] = append(h.service.logs[eaID], logEntry)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "EA started successfully",
		"data":    ea,
	})
}

// HandleStopEA handles POST /admin/automated-trading/eas/:id/stop
func (h *AutomatedTradingHandler) HandleStopEA(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	eaID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	h.service.mu.Lock()
	defer h.service.mu.Unlock()

	ea, exists := h.service.eas[eaID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA not found", http.StatusNotFound)
		return
	}

	if ea.Status == "stopped" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA is already stopped", http.StatusBadRequest)
		return
	}

	// Stop EA
	ea.Status = "stopped"

	// Add log entry
	logEntry := &EALog{
		ID:        h.service.nextLogID,
		EAID:      eaID,
		Timestamp: time.Now(),
		EventType: "stop",
		Message:   "EA stopped by admin",
		Severity:  "info",
	}
	h.service.nextLogID++
	h.service.logs[eaID] = append(h.service.logs[eaID], logEntry)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "EA stopped successfully",
		"data":    ea,
	})
}

// HandleGetEALogs handles GET /admin/automated-trading/eas/:id/logs
func (h *AutomatedTradingHandler) HandleGetEALogs(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	eaID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	logs, exists := h.service.logs[eaID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA not found", http.StatusNotFound)
		return
	}

	// Filter by event type if specified
	eventTypeFilter := r.URL.Query().Get("event_type")
	severityFilter := r.URL.Query().Get("severity")

	filteredLogs := make([]*EALog, 0)
	for _, log := range logs {
		if eventTypeFilter != "" && log.EventType != eventTypeFilter {
			continue
		}
		if severityFilter != "" && log.Severity != severityFilter {
			continue
		}
		filteredLogs = append(filteredLogs, log)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    filteredLogs,
		"total":   len(filteredLogs),
	})
}

// HandleGetEAPerformance handles GET /admin/automated-trading/eas/:id/performance
func (h *AutomatedTradingHandler) HandleGetEAPerformance(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	eaID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid EA ID", http.StatusBadRequest)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	ea, exists := h.service.eas[eaID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "EA not found", http.StatusNotFound)
		return
	}

	// Calculate performance metrics
	winningTrades := int(float64(ea.TotalTrades) * ea.WinRate)
	losingTrades := ea.TotalTrades - winningTrades

	avgWin := 0.0
	avgLoss := 0.0
	profitFactor := 0.0

	if winningTrades > 0 {
		avgWin = (ea.TotalPnL + math.Abs(ea.TotalPnL*0.3)) / float64(winningTrades)
	}
	if losingTrades > 0 {
		avgLoss = math.Abs(ea.TotalPnL * 0.3) / float64(losingTrades)
	}
	if avgLoss > 0 {
		profitFactor = avgWin / avgLoss
	}

	sharpeRatio := 0.0
	if ea.TotalTrades > 10 {
		sharpeRatio = (ea.TotalPnL / float64(ea.TotalTrades)) / (ea.MaxDD / 10)
	}

	performance := &EAPerformance{
		EAID:            eaID,
		TotalTrades:     ea.TotalTrades,
		WinningTrades:   winningTrades,
		LosingTrades:    losingTrades,
		WinRate:         ea.WinRate,
		TotalPnL:        ea.TotalPnL,
		AvgWin:          avgWin,
		AvgLoss:         avgLoss,
		ProfitFactor:    profitFactor,
		CurrentDrawdown: ea.CurrentDD,
		MaxDrawdown:     ea.MaxDD,
		SharpeRatio:     sharpeRatio,
		UptimeHours:     ea.UptimeHours,
		LastUpdated:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    performance,
	})
}

// HandleGetStats handles GET /admin/automated-trading/stats
func (h *AutomatedTradingHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	stats := &EAStats{
		TotalEAs: len(h.service.eas),
	}

	strategyTrades := make(map[string]int)
	totalWinRate := 0.0
	eaCount := 0

	for _, ea := range h.service.eas {
		switch ea.Status {
		case "running":
			stats.RunningEAs++
		case "paused":
			stats.PausedEAs++
		case "stopped":
			stats.StoppedEAs++
		case "error":
			stats.ErrorEAs++
		case "backtesting":
			stats.BacktestingEAs++
		}

		// Count trades today (simulated)
		if ea.Status == "running" && ea.LastTradeAt != nil && time.Since(*ea.LastTradeAt) < 24*time.Hour {
			todayTrades := rand.Intn(20) + 1
			stats.TotalTradesToday += todayTrades
			stats.TotalPnLToday += (rand.Float64()*1000 - 300)
		}

		strategyTrades[ea.Strategy] += ea.TotalTrades
		totalWinRate += ea.WinRate
		eaCount++
	}

	// Find top strategy
	maxTrades := 0
	for strategy, trades := range strategyTrades {
		if trades > maxTrades {
			maxTrades = trades
			stats.TopStrategy = strategy
		}
	}

	if eaCount > 0 {
		stats.AvgWinRate = totalWinRate / float64(eaCount)
	}

	// Calculate total resource usage
	for _, usage := range h.service.resourceUsage {
		stats.TotalCPUUsage += usage.CPUPercent
		stats.TotalMemoryUsageMB += usage.MemoryMB
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}
