package features

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// HTTP Handlers for Advanced Trading Features
// These handlers integrate all the features with the REST API

type FeatureHandlers struct {
	orderService     *AdvancedOrderService
	indicatorService *IndicatorService
	strategyService  *StrategyService
	alertService     *AlertService
	reportService    *ReportService
}

func NewFeatureHandlers(
	orderService *AdvancedOrderService,
	indicatorService *IndicatorService,
	strategyService *StrategyService,
	alertService *AlertService,
	reportService *ReportService,
) *FeatureHandlers {
	return &FeatureHandlers{
		orderService:     orderService,
		indicatorService: indicatorService,
		strategyService:  strategyService,
		alertService:     alertService,
		reportService:    reportService,
	}
}

// ===== Advanced Order Types =====

func (h *FeatureHandlers) HandlePlaceBracketOrder(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var req struct {
		Symbol      string  `json:"symbol"`
		Side        string  `json:"side"`
		Volume      float64 `json:"volume"`
		EntryPrice  float64 `json:"entryPrice"`
		StopLoss    float64 `json:"stopLoss"`
		TakeProfit  float64 `json:"takeProfit"`
		EntryType   string  `json:"entryType"`
		TimeInForce string  `json:"timeInForce"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := h.orderService.PlaceBracketOrder(
		req.Symbol,
		req.Side,
		req.Volume,
		req.EntryPrice,
		req.StopLoss,
		req.TakeProfit,
		req.EntryType,
		TimeInForce(req.TimeInForce),
		nil,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *FeatureHandlers) HandlePlaceTWAPOrder(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var req struct {
		Symbol      string  `json:"symbol"`
		Side        string  `json:"side"`
		TotalVolume float64 `json:"totalVolume"`
		DurationMin int     `json:"durationMinutes"`
		Interval    int     `json:"intervalSeconds"`
		MinPrice    float64 `json:"minPrice"`
		MaxPrice    float64 `json:"maxPrice"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(req.DurationMin) * time.Minute)

	order, err := h.orderService.PlaceTWAPOrder(
		req.Symbol,
		req.Side,
		req.TotalVolume,
		startTime,
		endTime,
		req.Interval,
		req.MinPrice,
		req.MaxPrice,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *FeatureHandlers) HandleGetBracketOrders(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	orders := h.orderService.GetBracketOrders()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// ===== Technical Indicators =====

func (h *FeatureHandlers) HandleCalculateIndicator(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	symbol := r.URL.Query().Get("symbol")
	indicator := r.URL.Query().Get("indicator")
	periodStr := r.URL.Query().Get("period")

	period := 14
	if periodStr != "" {
		if p, err := strconv.Atoi(periodStr); err == nil {
			period = p
		}
	}

	var result interface{}
	var err error

	switch indicator {
	case "sma":
		result, err = h.indicatorService.SMA(symbol, period)
	case "ema":
		result, err = h.indicatorService.EMA(symbol, period)
	case "rsi":
		result, err = h.indicatorService.RSI(symbol, period)
	case "macd":
		result, err = h.indicatorService.MACD(symbol, 12, 26, 9)
	case "bb":
		result, err = h.indicatorService.BollingerBands(symbol, period, 2.0)
	case "atr":
		result, err = h.indicatorService.ATR(symbol, period)
	case "adx":
		result, err = h.indicatorService.ADX(symbol, period)
	case "stochastic":
		result, err = h.indicatorService.Stochastic(symbol, period, 3)
	case "pivot":
		result, err = h.indicatorService.PivotPoints(symbol)
	default:
		http.Error(w, "unknown indicator", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ===== Strategy Automation =====

func (h *FeatureHandlers) HandleCreateStrategy(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var strategy Strategy
	if err := json.NewDecoder(r.Body).Decode(&strategy); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.strategyService.CreateStrategy(&strategy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

func (h *FeatureHandlers) HandleGetStrategies(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	strategies := h.strategyService.GetAllStrategies()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

func (h *FeatureHandlers) HandleRunBacktest(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var req struct {
		StrategyID     string  `json:"strategyId"`
		StartDate      string  `json:"startDate"`
		EndDate        string  `json:"endDate"`
		InitialBalance float64 `json:"initialBalance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startDate, _ := time.Parse(time.RFC3339, req.StartDate)
	endDate, _ := time.Parse(time.RFC3339, req.EndDate)

	result, err := h.strategyService.RunBacktest(
		req.StrategyID,
		startDate,
		endDate,
		req.InitialBalance,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ===== Alerts =====

func (h *FeatureHandlers) HandleCreateAlert(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	var alert Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.alertService.CreateAlert(&alert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(created)
}

func (h *FeatureHandlers) HandleGetUserAlerts(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId required", http.StatusBadRequest)
		return
	}

	alerts := h.alertService.GetUserAlerts(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func (h *FeatureHandlers) HandleGetAlertTriggers(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId required", http.StatusBadRequest)
		return
	}

	triggers := h.alertService.GetTriggers(userID, 50)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(triggers)
}

// ===== Reports =====

func (h *FeatureHandlers) HandleGenerateTaxReport(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	accountID := r.URL.Query().Get("accountId")
	yearStr := r.URL.Query().Get("year")

	if accountID == "" || yearStr == "" {
		http.Error(w, "accountId and year required", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "invalid year", http.StatusBadRequest)
		return
	}

	report, err := h.reportService.GenerateTaxReport(accountID, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *FeatureHandlers) HandleGeneratePerformanceReport(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	accountID := r.URL.Query().Get("accountId")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	if accountID == "" {
		http.Error(w, "accountId required", http.StatusBadRequest)
		return
	}

	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()

	if startDateStr != "" {
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	report, err := h.reportService.GeneratePerformanceReport(accountID, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *FeatureHandlers) HandleGenerateDrawdownAnalysis(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	accountID := r.URL.Query().Get("accountId")
	balanceStr := r.URL.Query().Get("initialBalance")

	if accountID == "" || balanceStr == "" {
		http.Error(w, "accountId and initialBalance required", http.StatusBadRequest)
		return
	}

	initialBalance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		http.Error(w, "invalid initialBalance", http.StatusBadRequest)
		return
	}

	startDate := time.Now().AddDate(0, -3, 0)
	endDate := time.Now()

	report, err := h.reportService.GenerateDrawdownAnalysis(accountID, startDate, endDate, initialBalance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// ===== Helper =====

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// RegisterRoutes registers all feature routes
func (h *FeatureHandlers) RegisterRoutes() {
	// Advanced Orders
	http.HandleFunc("/api/orders/bracket", h.HandlePlaceBracketOrder)
	http.HandleFunc("/api/orders/twap", h.HandlePlaceTWAPOrder)
	http.HandleFunc("/api/orders/bracket/list", h.HandleGetBracketOrders)

	// Indicators
	http.HandleFunc("/api/indicators/calculate", h.HandleCalculateIndicator)

	// Strategies
	http.HandleFunc("/api/strategies", h.HandleGetStrategies)
	http.HandleFunc("/api/strategies/create", h.HandleCreateStrategy)
	http.HandleFunc("/api/strategies/backtest", h.HandleRunBacktest)

	// Alerts
	http.HandleFunc("/api/alerts/create", h.HandleCreateAlert)
	http.HandleFunc("/api/alerts/list", h.HandleGetUserAlerts)
	http.HandleFunc("/api/alerts/triggers", h.HandleGetAlertTriggers)

	// Reports
	http.HandleFunc("/api/reports/tax", h.HandleGenerateTaxReport)
	http.HandleFunc("/api/reports/performance", h.HandleGeneratePerformanceReport)
	http.HandleFunc("/api/reports/drawdown", h.HandleGenerateDrawdownAnalysis)
}
