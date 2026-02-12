//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/google/uuid"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
	OrderTypeStop   OrderType = "stop"
)

// ExecutionRecord represents a single trade execution record
type ExecutionRecord struct {
	ID             string    `json:"id"`
	OrderID        string    `json:"order_id"`
	Symbol         string    `json:"symbol"`
	Side           string    `json:"side"` // buy/sell
	Volume         float64   `json:"volume"`
	RequestedPrice float64   `json:"requested_price"`
	ExecutedPrice  float64   `json:"executed_price"`
	Slippage       float64   `json:"slippage"`        // In pips
	ExecutionTimeMs int      `json:"execution_time_ms"` // Execution latency in milliseconds
	LP             string    `json:"lp"`              // Liquidity Provider
	OrderType      OrderType `json:"order_type"`
	Timestamp      time.Time `json:"timestamp"`
}

// ExecutionStats represents aggregated execution statistics
type ExecutionStats struct {
	TotalExecutions    int     `json:"total_executions"`
	AvgExecutionTimeMs float64 `json:"avg_execution_time_ms"`
	AvgSlippage        float64 `json:"avg_slippage"`
	FillRate           float64 `json:"fill_rate"`    // Percentage of orders filled
	RequoteRate        float64 `json:"requote_rate"` // Percentage of orders requoted
	TotalVolume        float64 `json:"total_volume"`
	PositiveSlippage   int     `json:"positive_slippage_count"`
	NegativeSlippage   int     `json:"negative_slippage_count"`
	ZeroSlippage       int     `json:"zero_slippage_count"`
}

// LPQualityMetrics represents execution quality per LP
type LPQualityMetrics struct {
	LP                 string  `json:"lp"`
	TotalExecutions    int     `json:"total_executions"`
	AvgExecutionTimeMs float64 `json:"avg_execution_time_ms"`
	AvgSlippage        float64 `json:"avg_slippage"`
	FillRate           float64 `json:"fill_rate"`
	RequoteRate        float64 `json:"requote_rate"`
	TotalVolume        float64 `json:"total_volume"`
}

// SymbolQualityMetrics represents execution quality per symbol
type SymbolQualityMetrics struct {
	Symbol             string  `json:"symbol"`
	TotalExecutions    int     `json:"total_executions"`
	AvgExecutionTimeMs float64 `json:"avg_execution_time_ms"`
	AvgSlippage        float64 `json:"avg_slippage"`
	TotalVolume        float64 `json:"total_volume"`
}

// HourlyExecutionData represents execution quality for a specific hour
type HourlyExecutionData struct {
	Hour               int     `json:"hour"` // 0-23
	TotalExecutions    int     `json:"total_executions"`
	AvgExecutionTimeMs float64 `json:"avg_execution_time_ms"`
	AvgSlippage        float64 `json:"avg_slippage"`
	TotalVolume        float64 `json:"total_volume"`
}

// ExecutionAnalyticsStore manages execution records and analytics
type ExecutionAnalyticsStore struct {
	mu      sync.RWMutex
	records []*ExecutionRecord
}

func NewExecutionAnalyticsStore() *ExecutionAnalyticsStore {
	store := &ExecutionAnalyticsStore{
		records: make([]*ExecutionRecord, 0),
	}

	// Initialize with 500+ mock execution records
	store.initializeMockData()

	log.Println("[ExecutionAnalytics] Execution analytics system initialized (500+ mock execution records)")

	return store
}

func (s *ExecutionAnalyticsStore) initializeMockData() {
	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
		"EURGBP", "EURJPY", "GBPJPY", "XAUUSD", "BTCUSD",
	}

	lps := []string{
		"YOFX", "OANDA", "IC Markets", "LMAX", "Saxo Bank",
	}

	orderTypes := []OrderType{
		OrderTypeMarket, OrderTypeLimit, OrderTypeStop,
	}

	sides := []string{"buy", "sell"}

	now := time.Now()

	// Generate 500 execution records over the last 7 days
	for i := 0; i < 500; i++ {
		symbol := symbols[rand.Intn(len(symbols))]
		lp := lps[rand.Intn(len(lps))]
		orderType := orderTypes[rand.Intn(len(orderTypes))]
		side := sides[rand.Intn(len(sides))]

		// Random timestamp in last 7 days
		daysAgo := rand.Intn(7)
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		timestamp := now.Add(-time.Duration(daysAgo)*24*time.Hour - time.Duration(hoursAgo)*time.Hour - time.Duration(minutesAgo)*time.Minute)

		// Generate realistic prices and slippage based on symbol
		var requestedPrice, executedPrice, slippage float64
		var executionTimeMs int

		switch symbol {
		case "EURUSD", "GBPUSD", "AUDUSD", "USDCAD":
			requestedPrice = 1.0 + rand.Float64()*0.5
			// Slippage: -2 to +2 pips (0.0002)
			slippage = (rand.Float64()*4.0 - 2.0)
			executedPrice = requestedPrice + (slippage * 0.0001)

		case "USDJPY":
			requestedPrice = 100.0 + rand.Float64()*50.0
			slippage = (rand.Float64()*4.0 - 2.0)
			executedPrice = requestedPrice + (slippage * 0.01)

		case "EURGBP":
			requestedPrice = 0.8 + rand.Float64()*0.1
			slippage = (rand.Float64()*4.0 - 2.0)
			executedPrice = requestedPrice + (slippage * 0.0001)

		case "EURJPY", "GBPJPY":
			requestedPrice = 120.0 + rand.Float64()*40.0
			slippage = (rand.Float64()*4.0 - 2.0)
			executedPrice = requestedPrice + (slippage * 0.01)

		case "XAUUSD":
			requestedPrice = 1800.0 + rand.Float64()*200.0
			slippage = (rand.Float64()*8.0 - 4.0) // Gold has higher slippage
			executedPrice = requestedPrice + (slippage * 0.1)

		case "BTCUSD":
			requestedPrice = 40000.0 + rand.Float64()*10000.0
			slippage = (rand.Float64()*20.0 - 10.0) // Crypto has much higher slippage
			executedPrice = requestedPrice + (slippage * 1.0)
		}

		// Execution time varies by LP quality
		switch lp {
		case "YOFX", "LMAX":
			executionTimeMs = 10 + rand.Intn(40) // Fast: 10-50ms
		case "IC Markets":
			executionTimeMs = 20 + rand.Intn(60) // Medium: 20-80ms
		case "OANDA":
			executionTimeMs = 30 + rand.Intn(80) // Slower: 30-110ms
		case "Saxo Bank":
			executionTimeMs = 40 + rand.Intn(100) // Slowest: 40-140ms
		}

		// Market orders execute faster
		if orderType == OrderTypeMarket {
			executionTimeMs = int(float64(executionTimeMs) * 0.7)
		}

		volume := 0.01 + rand.Float64()*9.99 // 0.01 to 10.0 lots

		record := &ExecutionRecord{
			ID:             uuid.New().String(),
			OrderID:        fmt.Sprintf("ORD-%d", 10000+i),
			Symbol:         symbol,
			Side:           side,
			Volume:         math.Round(volume*100) / 100,
			RequestedPrice: math.Round(requestedPrice*100000) / 100000,
			ExecutedPrice:  math.Round(executedPrice*100000) / 100000,
			Slippage:       math.Round(slippage*10) / 10,
			ExecutionTimeMs: executionTimeMs,
			LP:             lp,
			OrderType:      orderType,
			Timestamp:      timestamp,
		}

		s.records = append(s.records, record)
	}

	// Sort by timestamp (oldest first)
	sort.Slice(s.records, func(i, j int) bool {
		return s.records[i].Timestamp.Before(s.records[j].Timestamp)
	})
}

// GetOverallStats calculates overall execution statistics
func (s *ExecutionAnalyticsStore) GetOverallStats() *ExecutionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.records) == 0 {
		return &ExecutionStats{}
	}

	var totalExecutionTime int
	var totalSlippage float64
	var totalVolume float64
	var positiveSlippage, negativeSlippage, zeroSlippage int

	for _, record := range s.records {
		totalExecutionTime += record.ExecutionTimeMs
		totalSlippage += record.Slippage
		totalVolume += record.Volume

		if record.Slippage > 0 {
			positiveSlippage++
		} else if record.Slippage < 0 {
			negativeSlippage++
		} else {
			zeroSlippage++
		}
	}

	count := len(s.records)
	avgExecutionTime := float64(totalExecutionTime) / float64(count)
	avgSlippage := totalSlippage / float64(count)

	// Fill rate: assume 95% fill rate (5% rejections not in this dataset)
	fillRate := 95.0

	// Requote rate: assume 3% requote rate
	requoteRate := 3.0

	return &ExecutionStats{
		TotalExecutions:    count,
		AvgExecutionTimeMs: math.Round(avgExecutionTime*100) / 100,
		AvgSlippage:        math.Round(avgSlippage*100) / 100,
		FillRate:           fillRate,
		RequoteRate:        requoteRate,
		TotalVolume:        math.Round(totalVolume*100) / 100,
		PositiveSlippage:   positiveSlippage,
		NegativeSlippage:   negativeSlippage,
		ZeroSlippage:       zeroSlippage,
	}
}

// GetRecords returns paginated records with filters
func (s *ExecutionAnalyticsStore) GetRecords(symbol, lp, orderType string, startDate, endDate time.Time, page, pageSize int) ([]*ExecutionRecord, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter records
	filtered := make([]*ExecutionRecord, 0)
	for _, record := range s.records {
		// Apply filters
		if symbol != "" && record.Symbol != symbol {
			continue
		}
		if lp != "" && record.LP != lp {
			continue
		}
		if orderType != "" && string(record.OrderType) != orderType {
			continue
		}
		if !startDate.IsZero() && record.Timestamp.Before(startDate) {
			continue
		}
		if !endDate.IsZero() && record.Timestamp.After(endDate) {
			continue
		}

		filtered = append(filtered, record)
	}

	totalCount := len(filtered)

	// Apply pagination
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= totalCount {
		return []*ExecutionRecord{}, totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	// Return copies to avoid race conditions
	result := make([]*ExecutionRecord, end-start)
	for i := start; i < end; i++ {
		recordCopy := *filtered[i]
		result[i-start] = &recordCopy
	}

	return result, totalCount
}

// GetQualityByLP calculates execution quality metrics per LP
func (s *ExecutionAnalyticsStore) GetQualityByLP() []*LPQualityMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lpMap := make(map[string]*LPQualityMetrics)

	for _, record := range s.records {
		if _, exists := lpMap[record.LP]; !exists {
			lpMap[record.LP] = &LPQualityMetrics{
				LP:              record.LP,
				TotalExecutions: 0,
				FillRate:        95.0,
				RequoteRate:     3.0,
			}
		}

		metrics := lpMap[record.LP]
		metrics.TotalExecutions++
		metrics.AvgExecutionTimeMs += float64(record.ExecutionTimeMs)
		metrics.AvgSlippage += record.Slippage
		metrics.TotalVolume += record.Volume
	}

	// Calculate averages
	result := make([]*LPQualityMetrics, 0, len(lpMap))
	for _, metrics := range lpMap {
		if metrics.TotalExecutions > 0 {
			metrics.AvgExecutionTimeMs = math.Round(metrics.AvgExecutionTimeMs/float64(metrics.TotalExecutions)*100) / 100
			metrics.AvgSlippage = math.Round(metrics.AvgSlippage/float64(metrics.TotalExecutions)*100) / 100
			metrics.TotalVolume = math.Round(metrics.TotalVolume*100) / 100
		}
		result = append(result, metrics)
	}

	// Sort by total executions (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalExecutions > result[j].TotalExecutions
	})

	return result
}

// GetQualityBySymbol calculates execution quality metrics per symbol
func (s *ExecutionAnalyticsStore) GetQualityBySymbol() []*SymbolQualityMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	symbolMap := make(map[string]*SymbolQualityMetrics)

	for _, record := range s.records {
		if _, exists := symbolMap[record.Symbol]; !exists {
			symbolMap[record.Symbol] = &SymbolQualityMetrics{
				Symbol:          record.Symbol,
				TotalExecutions: 0,
			}
		}

		metrics := symbolMap[record.Symbol]
		metrics.TotalExecutions++
		metrics.AvgExecutionTimeMs += float64(record.ExecutionTimeMs)
		metrics.AvgSlippage += record.Slippage
		metrics.TotalVolume += record.Volume
	}

	// Calculate averages
	result := make([]*SymbolQualityMetrics, 0, len(symbolMap))
	for _, metrics := range symbolMap {
		if metrics.TotalExecutions > 0 {
			metrics.AvgExecutionTimeMs = math.Round(metrics.AvgExecutionTimeMs/float64(metrics.TotalExecutions)*100) / 100
			metrics.AvgSlippage = math.Round(metrics.AvgSlippage/float64(metrics.TotalExecutions)*100) / 100
			metrics.TotalVolume = math.Round(metrics.TotalVolume*100) / 100
		}
		result = append(result, metrics)
	}

	// Sort by total executions (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalExecutions > result[j].TotalExecutions
	})

	return result
}

// GetHourlyHeatmap returns execution quality data bucketed by hour of day (0-23)
func (s *ExecutionAnalyticsStore) GetHourlyHeatmap() []*HourlyExecutionData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialize 24 hourly buckets
	hourlyData := make(map[int]*HourlyExecutionData)
	for i := 0; i < 24; i++ {
		hourlyData[i] = &HourlyExecutionData{
			Hour:            i,
			TotalExecutions: 0,
		}
	}

	// Aggregate data by hour
	for _, record := range s.records {
		hour := record.Timestamp.Hour()
		data := hourlyData[hour]

		data.TotalExecutions++
		data.AvgExecutionTimeMs += float64(record.ExecutionTimeMs)
		data.AvgSlippage += record.Slippage
		data.TotalVolume += record.Volume
	}

	// Calculate averages
	result := make([]*HourlyExecutionData, 24)
	for i := 0; i < 24; i++ {
		data := hourlyData[i]
		if data.TotalExecutions > 0 {
			data.AvgExecutionTimeMs = math.Round(data.AvgExecutionTimeMs/float64(data.TotalExecutions)*100) / 100
			data.AvgSlippage = math.Round(data.AvgSlippage/float64(data.TotalExecutions)*100) / 100
			data.TotalVolume = math.Round(data.TotalVolume*100) / 100
		}
		result[i] = data
	}

	return result
}

// ExecutionAnalyticsHandler handles HTTP requests for execution analytics
type ExecutionAnalyticsHandler struct {
	store       *ExecutionAnalyticsStore
	authService *auth.AuthService
}

func NewExecutionAnalyticsHandler(store *ExecutionAnalyticsStore, authService *auth.AuthService) *ExecutionAnalyticsHandler {
	return &ExecutionAnalyticsHandler{
		store:       store,
		authService: authService,
	}
}

// HandleGetStats returns overall execution statistics
// GET /admin/execution/stats
func (h *ExecutionAnalyticsHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.store.GetOverallStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetRecords returns paginated execution records with filters
// GET /admin/execution/records?symbol=EURUSD&lp=YOFX&order_type=market&page=1&page_size=50
func (h *ExecutionAnalyticsHandler) HandleGetRecords(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	lp := r.URL.Query().Get("lp")
	orderType := r.URL.Query().Get("order_type")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	// Parse dates
	var startDate, endDate time.Time
	if startDateStr != "" {
		startDate, _ = time.Parse(time.RFC3339, startDateStr)
	}
	if endDateStr != "" {
		endDate, _ = time.Parse(time.RFC3339, endDateStr)
	}

	// Parse pagination
	page := 1
	pageSize := 50
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
			pageSize = ps
		}
	}

	records, totalCount := h.store.GetRecords(symbol, lp, orderType, startDate, endDate, page, pageSize)

	response := map[string]interface{}{
		"records":     records,
		"total_count": totalCount,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (totalCount + pageSize - 1) / pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetQualityByLP returns execution quality metrics per LP
// GET /admin/execution/by-lp
func (h *ExecutionAnalyticsHandler) HandleGetQualityByLP(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	metrics := h.store.GetQualityByLP()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleGetQualityBySymbol returns execution quality metrics per symbol
// GET /admin/execution/by-symbol
func (h *ExecutionAnalyticsHandler) HandleGetQualityBySymbol(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	metrics := h.store.GetQualityBySymbol()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleGetHourlyHeatmap returns hourly execution quality data
// GET /admin/execution/heatmap
func (h *ExecutionAnalyticsHandler) HandleGetHourlyHeatmap(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	heatmap := h.store.GetHourlyHeatmap()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(heatmap)
}
