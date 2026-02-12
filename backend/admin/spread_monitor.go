package admin

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SpreadStatus string

const (
	SpreadStatusNormal SpreadStatus = "normal"
	SpreadStatusWide   SpreadStatus = "wide"
	SpreadStatusAlert  SpreadStatus = "alert"
)

type SymbolGroup string

const (
	GroupForexMajors  SymbolGroup = "forex_majors"
	GroupForexMinors  SymbolGroup = "forex_minors"
	GroupForexExotics SymbolGroup = "forex_exotics"
	GroupMetals       SymbolGroup = "metals"
	GroupCrypto       SymbolGroup = "crypto"
	GroupIndices      SymbolGroup = "indices"
)

type SpreadData struct {
	Symbol        string       `json:"symbol"`
	Group         SymbolGroup  `json:"group"`
	Bid           float64      `json:"bid"`
	Ask           float64      `json:"ask"`
	CurrentSpread float64      `json:"currentSpread"` // in pips
	AvgSpread     float64      `json:"avgSpread"`
	MinSpread     float64      `json:"minSpread"`
	MaxSpread     float64      `json:"maxSpread"`
	Markup        float64      `json:"markup"`
	Status        SpreadStatus `json:"status"`
	AlertActive   bool         `json:"alertActive"`
}

type SpreadHistoryPoint struct {
	Timestamp string  `json:"timestamp"`
	Spread    float64 `json:"spread"`
}

type HourlyHeatmapData struct {
	Hour   int     `json:"hour"`
	Spread float64 `json:"spread"`
}

type LPSpreadData struct {
	LPName string  `json:"lpName"`
	Spread float64 `json:"spread"`
	Color  string  `json:"color"`
}

type SpreadAlert struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Threshold float64   `json:"threshold"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SpreadStats struct {
	AvgSpread     float64     `json:"avgSpread"`
	WidestSpread  *SpreadData `json:"widestSpread"`
	NarrowestSpread *SpreadData `json:"narrowestSpread"`
	AlertsActive  int         `json:"alertsActive"`
}

type SpreadMonitorStore struct {
	mu              sync.RWMutex
	spreads         map[string]*SpreadData
	history         map[string][]SpreadHistoryPoint
	heatmapData     map[string][]HourlyHeatmapData
	lpComparison    map[string][]LPSpreadData
	alerts          map[string]*SpreadAlert
	symbolAlertMap  map[string]string // symbol -> alert ID
}

var globalSpreadStore *SpreadMonitorStore

func init() {
	globalSpreadStore = NewSpreadMonitorStore()
	go globalSpreadStore.startSpreadSimulation()
}

func NewSpreadMonitorStore() *SpreadMonitorStore {
	store := &SpreadMonitorStore{
		spreads:        make(map[string]*SpreadData),
		history:        make(map[string][]SpreadHistoryPoint),
		heatmapData:    make(map[string][]HourlyHeatmapData),
		lpComparison:   make(map[string][]LPSpreadData),
		alerts:         make(map[string]*SpreadAlert),
		symbolAlertMap: make(map[string]string),
	}

	store.initializeSymbols()
	return store
}

func (s *SpreadMonitorStore) initializeSymbols() {
	symbols := []struct {
		symbol    string
		group     SymbolGroup
		bid       float64
		ask       float64
		avgSpread float64
		minSpread float64
		maxSpread float64
		markup    float64
	}{
		// Forex Majors
		{"EURUSD", GroupForexMajors, 1.08245, 1.08255, 1.2, 0.8, 2.5, 0.3},
		{"GBPUSD", GroupForexMajors, 1.26432, 1.26447, 1.5, 1.0, 3.0, 0.4},
		{"USDJPY", GroupForexMajors, 148.235, 148.250, 1.3, 1.0, 2.8, 0.3},
		{"USDCHF", GroupForexMajors, 0.87634, 0.87646, 1.4, 0.9, 2.6, 0.3},
		{"AUDUSD", GroupForexMajors, 0.65234, 0.65249, 1.6, 1.1, 3.2, 0.4},
		{"USDCAD", GroupForexMajors, 1.38456, 1.38471, 1.5, 1.0, 2.9, 0.4},
		{"NZDUSD", GroupForexMajors, 0.59123, 0.59141, 1.9, 1.3, 3.5, 0.5},

		// Forex Minors
		{"EURGBP", GroupForexMinors, 0.85632, 0.85650, 2.0, 1.4, 3.8, 0.5},
		{"EURJPY", GroupForexMinors, 160.456, 160.478, 2.1, 1.6, 4.2, 0.6},
		{"GBPJPY", GroupForexMinors, 187.234, 187.262, 2.5, 1.9, 5.0, 0.7},
		{"EURCHF", GroupForexMinors, 0.94567, 0.94587, 2.2, 1.5, 4.0, 0.6},
		{"EURAUD", GroupForexMinors, 1.65934, 1.65962, 2.9, 2.0, 5.5, 0.8},
		{"GBPAUD", GroupForexMinors, 1.93745, 1.93780, 3.4, 2.5, 6.5, 1.0},
		{"AUDCAD", GroupForexMinors, 0.90234, 0.90262, 2.7, 2.0, 5.2, 0.8},
		{"AUDJPY", GroupForexMinors, 96.734, 96.760, 2.5, 1.9, 4.8, 0.7},

		// Forex Exotics
		{"USDTRY", GroupForexExotics, 33.4567, 33.5234, 65.0, 45.0, 120.0, 15.0},
		{"USDZAR", GroupForexExotics, 18.2345, 18.2845, 48.0, 35.0, 85.0, 12.0},
		{"USDMXN", GroupForexExotics, 19.8734, 19.9234, 52.0, 38.0, 95.0, 12.0},
		{"USDSEK", GroupForexExotics, 10.6234, 10.6534, 28.0, 20.0, 55.0, 8.0},
		{"USDNOK", GroupForexExotics, 10.8456, 10.8756, 29.0, 21.0, 58.0, 8.0},

		// Metals
		{"XAUUSD", GroupMetals, 2034.56, 2034.86, 28.0, 20.0, 50.0, 8.0},
		{"XAGUSD", GroupMetals, 23.456, 23.486, 32.0, 22.0, 60.0, 8.0},
		{"XPTUSD", GroupMetals, 912.34, 913.34, 95.0, 70.0, 150.0, 25.0},
		{"XPDUSD", GroupMetals, 956.78, 957.78, 98.0, 75.0, 145.0, 25.0},

		// Crypto
		{"BTCUSD", GroupCrypto, 43256.78, 43276.45, 18.5, 12.0, 35.0, 5.0},
		{"ETHUSD", GroupCrypto, 2234.56, 2236.23, 15.2, 10.0, 28.0, 4.0},
		{"XRPUSD", GroupCrypto, 0.5234, 0.5256, 20.0, 15.0, 40.0, 6.0},
		{"SOLUSD", GroupCrypto, 98.456, 98.678, 21.0, 16.0, 38.0, 6.0},

		// Indices
		{"SPX500", GroupIndices, 4823.45, 4824.25, 75.0, 50.0, 120.0, 20.0},
		{"NAS100", GroupIndices, 16745.67, 16747.23, 150.0, 100.0, 250.0, 40.0},
		{"US30", GroupIndices, 37456.78, 37459.34, 245.0, 180.0, 380.0, 60.0},
		{"GER40", GroupIndices, 16934.56, 16936.23, 160.0, 110.0, 270.0, 45.0},
		{"UK100", GroupIndices, 7645.23, 7646.45, 118.0, 80.0, 200.0, 35.0},
		{"JP225", GroupIndices, 36234.56, 36237.89, 320.0, 220.0, 500.0, 80.0},
	}

	for _, sym := range symbols {
		currentSpread := sym.avgSpread * (0.9 + rand.Float64()*0.2)
		status := SpreadStatusNormal
		if currentSpread > sym.avgSpread*1.2 {
			status = SpreadStatusWide
		}

		s.spreads[sym.symbol] = &SpreadData{
			Symbol:        sym.symbol,
			Group:         sym.group,
			Bid:           sym.bid,
			Ask:           sym.ask,
			CurrentSpread: currentSpread,
			AvgSpread:     sym.avgSpread,
			MinSpread:     sym.minSpread,
			MaxSpread:     sym.maxSpread,
			Markup:        sym.markup,
			Status:        status,
			AlertActive:   false,
		}

		// Initialize 24h history (144 points, 10-minute intervals)
		s.history[sym.symbol] = s.generate24HHistory(sym.avgSpread, sym.minSpread)

		// Initialize hourly heatmap
		s.heatmapData[sym.symbol] = s.generateHourlyHeatmap(sym.avgSpread)

		// Initialize LP comparison
		rawSpread := currentSpread - sym.markup
		s.lpComparison[sym.symbol] = []LPSpreadData{
			{LPName: "YOFX", Spread: rawSpread, Color: "#3B82F6"},
			{LPName: "LP-PRIME", Spread: rawSpread + 0.1, Color: "#10B981"},
			{LPName: "LMAX", Spread: rawSpread + 0.2, Color: "#F59E0B"},
			{LPName: "B2Broker", Spread: rawSpread - 0.1, Color: "#8B5CF6"},
		}
	}

	// Initialize some sample alerts
	alert1 := &SpreadAlert{
		ID:        uuid.New().String(),
		Symbol:    "EURUSD",
		Threshold: 2.0,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.alerts[alert1.ID] = alert1
	s.symbolAlertMap["EURUSD"] = alert1.ID
	s.spreads["EURUSD"].AlertActive = true

	alert2 := &SpreadAlert{
		ID:        uuid.New().String(),
		Symbol:    "XAUUSD",
		Threshold: 32.0,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.alerts[alert2.ID] = alert2
	s.symbolAlertMap["XAUUSD"] = alert2.ID
	s.spreads["XAUUSD"].AlertActive = true

	alert3 := &SpreadAlert{
		ID:        uuid.New().String(),
		Symbol:    "GBPJPY",
		Threshold: 2.6,
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.alerts[alert3.ID] = alert3
	s.symbolAlertMap["GBPJPY"] = alert3.ID
	s.spreads["GBPJPY"].AlertActive = true
	s.spreads["GBPJPY"].Status = SpreadStatusAlert
}

func (s *SpreadMonitorStore) generate24HHistory(avgSpread, minSpread float64) []SpreadHistoryPoint {
	points := make([]SpreadHistoryPoint, 144)
	now := time.Now()

	for i := 143; i >= 0; i-- {
		timestamp := now.Add(time.Duration(-i*10) * time.Minute)
		hour := timestamp.Hour()

		// Wider spreads during low liquidity hours (10 PM - 6 AM)
		isLowLiquidity := hour >= 22 || hour <= 6
		volatility := 0.2
		if isLowLiquidity {
			volatility = 0.4
		}

		spread := math.Max(minSpread, avgSpread+(rand.Float64()-0.5)*avgSpread*volatility)
		points[143-i] = SpreadHistoryPoint{
			Timestamp: timestamp.Format(time.RFC3339),
			Spread:    spread,
		}
	}

	return points
}

func (s *SpreadMonitorStore) generateHourlyHeatmap(avgSpread float64) []HourlyHeatmapData {
	heatmap := make([]HourlyHeatmapData, 24)

	for hour := 0; hour < 24; hour++ {
		isLowLiquidity := hour >= 22 || hour <= 6
		multiplier := 1.0
		if isLowLiquidity {
			multiplier = 1.3
		}

		spread := avgSpread * multiplier * (0.9 + rand.Float64()*0.2)
		heatmap[hour] = HourlyHeatmapData{
			Hour:   hour,
			Spread: spread,
		}
	}

	return heatmap
}

func (s *SpreadMonitorStore) startSpreadSimulation() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.updateSpreads()
	}
}

func (s *SpreadMonitorStore) updateSpreads() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	hour := now.Hour()
	isLowLiquidity := hour >= 22 || hour <= 6

	for symbol, spread := range s.spreads {
		// Update current spread with slight variation
		volatility := 0.1
		if isLowLiquidity {
			volatility = 0.2
		}

		newSpread := spread.AvgSpread * (0.9 + rand.Float64()*0.2) * (1.0 + (rand.Float64()-0.5)*volatility)
		newSpread = math.Max(spread.MinSpread, math.Min(spread.MaxSpread, newSpread))

		spread.CurrentSpread = newSpread

		// Update status
		if alertID, hasAlert := s.symbolAlertMap[symbol]; hasAlert {
			if alert, exists := s.alerts[alertID]; exists && alert.Enabled {
				if newSpread >= alert.Threshold {
					spread.Status = SpreadStatusAlert
				} else {
					spread.Status = SpreadStatusNormal
				}
			}
		} else if newSpread > spread.AvgSpread*1.2 {
			spread.Status = SpreadStatusWide
		} else {
			spread.Status = SpreadStatusNormal
		}

		// Update bid/ask
		pipValue := 0.0001
		if symbol == "USDJPY" || symbol == "EURJPY" || symbol == "GBPJPY" || symbol == "AUDJPY" {
			pipValue = 0.01
		}
		halfSpreadPips := newSpread / 2.0
		spread.Ask = spread.Bid + (halfSpreadPips * pipValue)

		// Add to history (shift array)
		history := s.history[symbol]
		if len(history) >= 144 {
			history = history[1:]
		}
		history = append(history, SpreadHistoryPoint{
			Timestamp: now.Format(time.RFC3339),
			Spread:    newSpread,
		})
		s.history[symbol] = history
	}
}

// HTTP Handlers

func HandleGetLiveSpreads(w http.ResponseWriter, r *http.Request) {
	globalSpreadStore.mu.RLock()
	defer globalSpreadStore.mu.RUnlock()

	spreads := make([]*SpreadData, 0, len(globalSpreadStore.spreads))
	for _, spread := range globalSpreadStore.spreads {
		spreads = append(spreads, spread)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(spreads)
}

func HandleGetSpreadHistory(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path: /admin/spreads/{symbol}/history
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := parts[3]

	globalSpreadStore.mu.RLock()
	history, exists := globalSpreadStore.history[symbol]
	globalSpreadStore.mu.RUnlock()

	if !exists {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(history)
}

func HandleGetSpreadStats(w http.ResponseWriter, r *http.Request) {
	globalSpreadStore.mu.RLock()
	defer globalSpreadStore.mu.RUnlock()

	var totalSpread float64
	var widest, narrowest *SpreadData
	alertsActive := 0

	for _, spread := range globalSpreadStore.spreads {
		totalSpread += spread.CurrentSpread

		if widest == nil || spread.CurrentSpread > widest.CurrentSpread {
			widest = spread
		}
		if narrowest == nil || spread.CurrentSpread < narrowest.CurrentSpread {
			narrowest = spread
		}

		if spread.AlertActive {
			alertsActive++
		}
	}

	avgSpread := totalSpread / float64(len(globalSpreadStore.spreads))

	stats := SpreadStats{
		AvgSpread:       avgSpread,
		WidestSpread:    widest,
		NarrowestSpread: narrowest,
		AlertsActive:    alertsActive,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(stats)
}

func HandleGetSpreadHeatmap(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path: /admin/spreads/{symbol}/heatmap
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := parts[3]

	globalSpreadStore.mu.RLock()
	heatmap, exists := globalSpreadStore.heatmapData[symbol]
	globalSpreadStore.mu.RUnlock()

	if !exists {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(heatmap)
}

func HandleGetLPComparison(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path: /admin/spreads/{symbol}/lp-comparison
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	symbol := parts[3]

	globalSpreadStore.mu.RLock()
	lpData, exists := globalSpreadStore.lpComparison[symbol]
	globalSpreadStore.mu.RUnlock()

	if !exists {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(lpData)
}

func HandleGetSpreadAlerts(w http.ResponseWriter, r *http.Request) {
	globalSpreadStore.mu.RLock()
	defer globalSpreadStore.mu.RUnlock()

	alerts := make([]*SpreadAlert, 0, len(globalSpreadStore.alerts))
	for _, alert := range globalSpreadStore.alerts {
		alerts = append(alerts, alert)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(alerts)
}

func HandleCreateSpreadAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol    string  `json:"symbol"`
		Threshold float64 `json:"threshold"`
		Enabled   bool    `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	globalSpreadStore.mu.Lock()
	defer globalSpreadStore.mu.Unlock()

	// Check if symbol exists
	if _, exists := globalSpreadStore.spreads[req.Symbol]; !exists {
		http.Error(w, "Symbol not found", http.StatusNotFound)
		return
	}

	alert := &SpreadAlert{
		ID:        uuid.New().String(),
		Symbol:    req.Symbol,
		Threshold: req.Threshold,
		Enabled:   req.Enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	globalSpreadStore.alerts[alert.ID] = alert
	globalSpreadStore.symbolAlertMap[req.Symbol] = alert.ID
	globalSpreadStore.spreads[req.Symbol].AlertActive = req.Enabled

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

func HandleUpdateSpreadAlert(w http.ResponseWriter, r *http.Request) {
	// Extract alert ID from URL path: /admin/spreads/alerts/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	alertID := parts[4]

	var req struct {
		Threshold float64 `json:"threshold"`
		Enabled   bool    `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	globalSpreadStore.mu.Lock()
	defer globalSpreadStore.mu.Unlock()

	alert, exists := globalSpreadStore.alerts[alertID]
	if !exists {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	alert.Threshold = req.Threshold
	alert.Enabled = req.Enabled
	alert.UpdatedAt = time.Now()

	if spread, ok := globalSpreadStore.spreads[alert.Symbol]; ok {
		spread.AlertActive = req.Enabled
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(alert)
}

func HandleDeleteSpreadAlert(w http.ResponseWriter, r *http.Request) {
	// Extract alert ID from URL path: /admin/spreads/alerts/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}
	alertID := parts[4]

	globalSpreadStore.mu.Lock()
	defer globalSpreadStore.mu.Unlock()

	alert, exists := globalSpreadStore.alerts[alertID]
	if !exists {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	// Remove alert and update symbol
	symbol := alert.Symbol
	delete(globalSpreadStore.alerts, alertID)
	delete(globalSpreadStore.symbolAlertMap, symbol)

	if spread, ok := globalSpreadStore.spreads[symbol]; ok {
		spread.AlertActive = false
		spread.Status = SpreadStatusNormal
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	w.WriteHeader(http.StatusNoContent)
}
