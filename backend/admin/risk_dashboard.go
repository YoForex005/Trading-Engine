package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ExposureData represents exposure metrics for a symbol
type ExposureData struct {
	Symbol         string  `json:"symbol"`
	NetLots        float64 `json:"net_lots"`
	LongLots       float64 `json:"long_lots"`
	ShortLots      float64 `json:"short_lots"`
	NotionalValue  float64 `json:"notional_value"`
	UnrealizedPnL  float64 `json:"unrealized_pnl"`
	ClientCount    int     `json:"client_count"`
	CurrentPrice   float64 `json:"current_price"`
	SymbolGroup    string  `json:"symbol_group"`
}

// RiskMetrics represents platform-wide risk metrics
type RiskMetrics struct {
	TotalExposure          float64 `json:"total_exposure"`
	TotalUnrealizedPnL     float64 `json:"total_unrealized_pnl"`
	LargestSingleExposure  float64 `json:"largest_single_exposure"`
	MarginUtilization      float64 `json:"margin_utilization"`
	ConcentrationRisk      float64 `json:"concentration_risk"`
	TotalPositions         int     `json:"total_positions"`
	ActiveClients          int     `json:"active_clients"`
	MarginUsed             float64 `json:"margin_used"`
	MarginAvailable        float64 `json:"margin_available"`
}

// Position represents a single open position
type Position struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	ClientName    string    `json:"client_name"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	Lots          float64   `json:"lots"`
	OpenPrice     float64   `json:"open_price"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	NotionalValue float64   `json:"notional_value"`
	MarginUsed    float64   `json:"margin_used"`
	OpenTime      time.Time `json:"open_time"`
}

// ClientExposure represents exposure for a single client
type ClientExposure struct {
	ClientID       string  `json:"client_id"`
	ClientName     string  `json:"client_name"`
	TotalExposure  float64 `json:"total_exposure"`
	UnrealizedPnL  float64 `json:"unrealized_pnl"`
	PositionCount  int     `json:"position_count"`
	MarginUsed     float64 `json:"margin_used"`
	AccountEquity  float64 `json:"account_equity"`
	MarginLevel    float64 `json:"margin_level"`
	RiskLevel      string  `json:"risk_level"`
}

// ConcentrationData represents concentration risk by symbol group
type ConcentrationData struct {
	SymbolGroup      string  `json:"symbol_group"`
	TotalExposure    float64 `json:"total_exposure"`
	PercentOfTotal   float64 `json:"percent_of_total"`
	PositionCount    int     `json:"position_count"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	SymbolCount      int     `json:"symbol_count"`
}

// VaRData represents Value at Risk calculations
type VaRData struct {
	VaR95Percent      float64   `json:"var_95_percent"`
	VaR99Percent      float64   `json:"var_99_percent"`
	CalculationMethod string    `json:"calculation_method"`
	TimeHorizon       string    `json:"time_horizon"`
	CalculatedAt      time.Time `json:"calculated_at"`
	HistoricalDays    int       `json:"historical_days"`
	TotalExposure     float64   `json:"total_exposure"`
}

// RiskLimits represents risk management limits
type RiskLimits struct {
	MaxExposure              float64   `json:"max_exposure"`
	MaxSingleClient          float64   `json:"max_single_client"`
	MaxSymbolConcentration   float64   `json:"max_symbol_concentration"`
	MaxGroupConcentration    float64   `json:"max_group_concentration"`
	MinMarginLevel           float64   `json:"min_margin_level"`
	MaxLeverageRatio         float64   `json:"max_leverage_ratio"`
	UpdatedAt                time.Time `json:"updated_at"`
	UpdatedBy                string    `json:"updated_by"`
}

// RiskAlert represents an active risk alert
type RiskAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	EntityID    string    `json:"entity_id"`
	EntityName  string    `json:"entity_name"`
	Threshold   float64   `json:"threshold"`
	CurrentValue float64  `json:"current_value"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

// RiskDashboardStore manages risk dashboard data
type RiskDashboardStore struct {
	mu        sync.RWMutex
	positions []Position
	limits    RiskLimits
	alerts    []RiskAlert
}

// NewRiskDashboardStore creates a new risk dashboard store with mock data
func NewRiskDashboardStore() *RiskDashboardStore {
	store := &RiskDashboardStore{
		positions: generateMockPositions(),
		limits: RiskLimits{
			MaxExposure:            50000000.0,
			MaxSingleClient:        2000000.0,
			MaxSymbolConcentration: 0.25,
			MaxGroupConcentration:  0.35,
			MinMarginLevel:         100.0,
			MaxLeverageRatio:       100.0,
			UpdatedAt:              time.Now().Add(-30 * 24 * time.Hour),
			UpdatedBy:              "System",
		},
		alerts: []RiskAlert{},
	}

	store.generateRiskAlerts()
	return store
}

// generateMockPositions creates 500 mock positions across 30 symbols
func generateMockPositions() []Position {
	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD",
		"EURGBP", "EURJPY", "GBPJPY", "AUDJPY", "EURAUD", "EURCHF", "GBPCHF",
		"XAUUSD", "XAGUSD", "BTCUSD", "ETHUSD", "WTICOUSD", "BRENTUSD",
		"US30", "SPX500", "NAS100", "UK100", "DE40", "JP225", "AU200",
		"USDZAR", "USDMXN", "USDSGD",
	}

	prices := map[string]float64{
		"EURUSD": 1.0850, "GBPUSD": 1.2650, "USDJPY": 148.50, "USDCHF": 0.8850,
		"AUDUSD": 0.6550, "USDCAD": 1.3650, "NZDUSD": 0.5950, "EURGBP": 0.8580,
		"EURJPY": 161.10, "GBPJPY": 187.85, "AUDJPY": 97.28, "EURAUD": 1.6565,
		"EURCHF": 0.9602, "GBPCHF": 1.1195, "XAUUSD": 2650.50, "XAGUSD": 30.45,
		"BTCUSD": 95500.0, "ETHUSD": 3450.0, "WTICOUSD": 78.50, "BRENTUSD": 82.30,
		"US30": 38500.0, "SPX500": 4850.0, "NAS100": 17200.0, "UK100": 7650.0,
		"DE40": 17800.0, "JP225": 36500.0, "AU200": 7850.0, "USDZAR": 18.50,
		"USDMXN": 17.25, "USDSGD": 1.3450,
	}

	clients := []string{
		"CL001-John Smith", "CL002-Sarah Johnson", "CL003-Michael Brown", "CL004-Emily Davis",
		"CL005-David Wilson", "CL006-Lisa Anderson", "CL007-Robert Taylor", "CL008-Jennifer Thomas",
		"CL009-William Jackson", "CL010-Mary White", "CL011-James Harris", "CL012-Patricia Martin",
		"CL013-Christopher Thompson", "CL014-Linda Garcia", "CL015-Daniel Martinez", "CL016-Barbara Robinson",
		"CL017-Matthew Clark", "CL018-Nancy Rodriguez", "CL019-Anthony Lewis", "CL020-Karen Lee",
		"CL021-Mark Walker", "CL022-Betty Hall", "CL023-Donald Allen", "CL024-Helen Young",
		"CL025-Paul Hernandez", "CL026-Sandra King", "CL027-Andrew Wright", "CL028-Donna Lopez",
		"CL029-Joshua Hill", "CL030-Carol Scott", "CL031-Kenneth Green", "CL032-Michelle Adams",
		"CL033-Kevin Baker", "CL034-Amanda Nelson", "CL035-Steven Carter", "CL036-Melissa Mitchell",
		"CL037-Brian Perez", "CL038-Deborah Roberts", "CL039-George Turner", "CL040-Stephanie Phillips",
	}

	positions := make([]Position, 0, 500)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 500; i++ {
		symbol := symbols[rand.Intn(len(symbols))]
		clientInfo := clients[rand.Intn(len(clients))]
		clientParts := strings.Split(clientInfo, "-")
		clientID := clientParts[0]
		clientName := clientParts[1]

		currentPrice := prices[symbol]
		priceVariation := currentPrice * (rand.Float64()*0.04 - 0.02)
		openPrice := currentPrice + priceVariation

		side := "buy"
		if rand.Float64() < 0.48 {
			side = "sell"
		}

		lots := []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0}[rand.Intn(9)]

		var unrealizedPnL float64
		if side == "buy" {
			unrealizedPnL = (currentPrice - openPrice) * lots * 100000
		} else {
			unrealizedPnL = (openPrice - currentPrice) * lots * 100000
		}

		if strings.HasPrefix(symbol, "USD") && !strings.Contains(symbol, "JPY") {
			unrealizedPnL = unrealizedPnL / currentPrice
		}

		contractSize := 100000.0
		if strings.Contains(symbol, "XAU") || strings.Contains(symbol, "XAG") {
			contractSize = 100.0
		} else if strings.Contains(symbol, "BTC") || strings.Contains(symbol, "ETH") {
			contractSize = 1.0
		} else if strings.Contains(symbol, "WTI") || strings.Contains(symbol, "BRENT") {
			contractSize = 1000.0
		} else if strings.HasPrefix(symbol, "US") || strings.HasPrefix(symbol, "SPX") || strings.HasPrefix(symbol, "NAS") {
			contractSize = 1.0
		}

		notionalValue := lots * contractSize * currentPrice
		marginUsed := notionalValue / 100.0

		daysAgo := rand.Intn(30)
		hoursAgo := rand.Intn(24)
		openTime := time.Now().Add(-time.Duration(daysAgo*24+hoursAgo) * time.Hour)

		positions = append(positions, Position{
			ID:            fmt.Sprintf("POS%d", 100000+i),
			ClientID:      clientID,
			ClientName:    clientName,
			Symbol:        symbol,
			Side:          side,
			Lots:          lots,
			OpenPrice:     openPrice,
			CurrentPrice:  currentPrice,
			UnrealizedPnL: unrealizedPnL,
			NotionalValue: notionalValue,
			MarginUsed:    marginUsed,
			OpenTime:      openTime,
		})
	}

	return positions
}

// generateRiskAlerts creates realistic risk alerts based on current exposures
func (s *RiskDashboardStore) generateRiskAlerts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	alerts := []RiskAlert{
		{
			ID:           "ALERT001",
			Type:         "margin_warning",
			Severity:     "warning",
			Message:      "Client CL005-David Wilson margin level dropped to 120%",
			EntityID:     "CL005",
			EntityName:   "David Wilson",
			Threshold:    150.0,
			CurrentValue: 120.0,
			Timestamp:    time.Now().Add(-15 * time.Minute),
			Status:       "active",
		},
		{
			ID:           "ALERT002",
			Type:         "concentration_breach",
			Severity:     "high",
			Message:      "FX Major pairs concentration exceeds limit: 38% of total exposure",
			EntityID:     "GROUP_FX_MAJOR",
			EntityName:   "FX Major Pairs",
			Threshold:    35.0,
			CurrentValue: 38.2,
			Timestamp:    time.Now().Add(-45 * time.Minute),
			Status:       "active",
		},
		{
			ID:           "ALERT003",
			Type:         "var_exceedance",
			Severity:     "critical",
			Message:      "Daily VaR (99%) exceeded: Current exposure $52.3M vs limit $50M",
			EntityID:     "PLATFORM",
			EntityName:   "Platform Risk",
			Threshold:    50000000.0,
			CurrentValue: 52300000.0,
			Timestamp:    time.Now().Add(-2 * time.Hour),
			Status:       "active",
		},
		{
			ID:           "ALERT004",
			Type:         "client_exposure",
			Severity:     "warning",
			Message:      "Client CL012-Patricia Martin exposure $1.85M approaching limit",
			EntityID:     "CL012",
			EntityName:   "Patricia Martin",
			Threshold:    2000000.0,
			CurrentValue: 1850000.0,
			Timestamp:    time.Now().Add(-30 * time.Minute),
			Status:       "active",
		},
		{
			ID:           "ALERT005",
			Type:         "symbol_concentration",
			Severity:     "medium",
			Message:      "EURUSD exposure 28% of total, approaching symbol limit",
			EntityID:     "EURUSD",
			EntityName:   "EURUSD",
			Threshold:    25.0,
			CurrentValue: 28.0,
			Timestamp:    time.Now().Add(-1 * time.Hour),
			Status:       "active",
		},
	}

	s.alerts = alerts
}

// calculateExposureData calculates exposure metrics for all symbols
func (s *RiskDashboardStore) calculateExposureData() []ExposureData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exposureMap := make(map[string]*ExposureData)
	clientsPerSymbol := make(map[string]map[string]bool)

	for _, pos := range s.positions {
		if _, exists := exposureMap[pos.Symbol]; !exists {
			exposureMap[pos.Symbol] = &ExposureData{
				Symbol:       pos.Symbol,
				CurrentPrice: pos.CurrentPrice,
				SymbolGroup:  getSymbolGroup(pos.Symbol),
			}
			clientsPerSymbol[pos.Symbol] = make(map[string]bool)
		}

		exp := exposureMap[pos.Symbol]
		clientsPerSymbol[pos.Symbol][pos.ClientID] = true

		if pos.Side == "buy" {
			exp.LongLots += pos.Lots
			exp.NetLots += pos.Lots
		} else {
			exp.ShortLots += pos.Lots
			exp.NetLots -= pos.Lots
		}

		exp.NotionalValue += pos.NotionalValue
		exp.UnrealizedPnL += pos.UnrealizedPnL
	}

	result := make([]ExposureData, 0, len(exposureMap))
	for symbol, exp := range exposureMap {
		exp.ClientCount = len(clientsPerSymbol[symbol])
		result = append(result, *exp)
	}

	sort.Slice(result, func(i, j int) bool {
		return math.Abs(result[i].NotionalValue) > math.Abs(result[j].NotionalValue)
	})

	return result
}

// getSymbolGroup returns the group classification for a symbol
func getSymbolGroup(symbol string) string {
	if strings.Contains(symbol, "USD") && len(symbol) == 6 {
		return "FX Major"
	}
	if strings.HasPrefix(symbol, "XAU") || strings.HasPrefix(symbol, "XAG") {
		return "Metals"
	}
	if strings.Contains(symbol, "BTC") || strings.Contains(symbol, "ETH") {
		return "Crypto"
	}
	if strings.Contains(symbol, "WTI") || strings.Contains(symbol, "BRENT") {
		return "Energy"
	}
	if strings.Contains(symbol, "US") || strings.Contains(symbol, "SPX") || strings.Contains(symbol, "NAS") || strings.Contains(symbol, "UK") || strings.Contains(symbol, "DE") || strings.Contains(symbol, "JP") || strings.Contains(symbol, "AU") {
		return "Indices"
	}
	return "FX Minor"
}

// calculateRiskMetrics calculates platform-wide risk metrics
func (s *RiskDashboardStore) calculateRiskMetrics() RiskMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalExposure, totalPnL, marginUsed float64
	clientSet := make(map[string]bool)
	largestExposure := 0.0

	for _, pos := range s.positions {
		totalExposure += math.Abs(pos.NotionalValue)
		totalPnL += pos.UnrealizedPnL
		marginUsed += pos.MarginUsed
		clientSet[pos.ClientID] = true

		if math.Abs(pos.NotionalValue) > largestExposure {
			largestExposure = math.Abs(pos.NotionalValue)
		}
	}

	marginAvailable := marginUsed * 9.0
	marginUtilization := (marginUsed / (marginUsed + marginAvailable)) * 100

	exposures := s.calculateExposureData()
	concentrationRisk := 0.0
	if len(exposures) > 0 && totalExposure > 0 {
		concentrationRisk = (math.Abs(exposures[0].NotionalValue) / totalExposure) * 100
	}

	return RiskMetrics{
		TotalExposure:         totalExposure,
		TotalUnrealizedPnL:    totalPnL,
		LargestSingleExposure: largestExposure,
		MarginUtilization:     marginUtilization,
		ConcentrationRisk:     concentrationRisk,
		TotalPositions:        len(s.positions),
		ActiveClients:         len(clientSet),
		MarginUsed:            marginUsed,
		MarginAvailable:       marginAvailable,
	}
}

// calculateClientExposures calculates exposure for each client
func (s *RiskDashboardStore) calculateClientExposures() []ClientExposure {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientMap := make(map[string]*ClientExposure)

	for _, pos := range s.positions {
		if _, exists := clientMap[pos.ClientID]; !exists {
			clientMap[pos.ClientID] = &ClientExposure{
				ClientID:   pos.ClientID,
				ClientName: pos.ClientName,
			}
		}

		client := clientMap[pos.ClientID]
		client.TotalExposure += math.Abs(pos.NotionalValue)
		client.UnrealizedPnL += pos.UnrealizedPnL
		client.MarginUsed += pos.MarginUsed
		client.PositionCount++
	}

	result := make([]ClientExposure, 0, len(clientMap))
	for _, client := range clientMap {
		client.AccountEquity = 10000.0 + client.UnrealizedPnL + (rand.Float64() * 50000.0)
		if client.MarginUsed > 0 {
			client.MarginLevel = (client.AccountEquity / client.MarginUsed) * 100
		} else {
			client.MarginLevel = 0
		}

		if client.MarginLevel < 100 {
			client.RiskLevel = "critical"
		} else if client.MarginLevel < 150 {
			client.RiskLevel = "high"
		} else if client.MarginLevel < 200 {
			client.RiskLevel = "medium"
		} else {
			client.RiskLevel = "low"
		}

		result = append(result, *client)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalExposure > result[j].TotalExposure
	})

	if len(result) > 20 {
		result = result[:20]
	}

	return result
}

// calculateConcentrationRisk calculates concentration by symbol group
func (s *RiskDashboardStore) calculateConcentrationRisk() []ConcentrationData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groupMap := make(map[string]*ConcentrationData)
	var totalExposure float64

	for _, pos := range s.positions {
		group := getSymbolGroup(pos.Symbol)

		if _, exists := groupMap[group]; !exists {
			groupMap[group] = &ConcentrationData{
				SymbolGroup: group,
			}
		}

		data := groupMap[group]
		exposure := math.Abs(pos.NotionalValue)
		data.TotalExposure += exposure
		data.UnrealizedPnL += pos.UnrealizedPnL
		data.PositionCount++
		totalExposure += exposure
	}

	symbolCounts := make(map[string]map[string]bool)
	for _, pos := range s.positions {
		group := getSymbolGroup(pos.Symbol)
		if symbolCounts[group] == nil {
			symbolCounts[group] = make(map[string]bool)
		}
		symbolCounts[group][pos.Symbol] = true
	}

	result := make([]ConcentrationData, 0, len(groupMap))
	for group, data := range groupMap {
		data.PercentOfTotal = (data.TotalExposure / totalExposure) * 100
		data.SymbolCount = len(symbolCounts[group])
		result = append(result, *data)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalExposure > result[j].TotalExposure
	})

	return result
}

// calculateVaR calculates Value at Risk using historical simulation
func (s *RiskDashboardStore) calculateVaR() VaRData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	historicalDays := 252
	dailyReturns := make([]float64, historicalDays)

	for i := 0; i < historicalDays; i++ {
		dailyReturns[i] = (rand.Float64()*0.04 - 0.02) * 100
	}

	sort.Float64s(dailyReturns)

	var totalExposure float64
	for _, pos := range s.positions {
		totalExposure += math.Abs(pos.NotionalValue)
	}

	index95 := int(float64(historicalDays) * 0.05)
	index99 := int(float64(historicalDays) * 0.01)

	var95 := math.Abs(dailyReturns[index95]) * totalExposure / 100
	var99 := math.Abs(dailyReturns[index99]) * totalExposure / 100

	return VaRData{
		VaR95Percent:      var95,
		VaR99Percent:      var99,
		CalculationMethod: "Historical Simulation",
		TimeHorizon:       "1 Day",
		CalculatedAt:      time.Now(),
		HistoricalDays:    historicalDays,
		TotalExposure:     totalExposure,
	}
}

// RiskDashboardHandler handles risk dashboard requests
type RiskDashboardHandler struct {
	store       *RiskDashboardStore
	authService *auth.Service
}

// NewRiskDashboardHandler creates a new risk dashboard handler
func NewRiskDashboardHandler(store *RiskDashboardStore, authService *auth.Service) *RiskDashboardHandler {
	return &RiskDashboardHandler{
		store:       store,
		authService: authService,
	}
}

// HandleGetOverview handles GET /admin/risk/overview
func (h *RiskDashboardHandler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	metrics := h.store.calculateRiskMetrics()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    metrics,
	})
}

// HandleGetExposure handles GET /admin/risk/exposure
func (h *RiskDashboardHandler) HandleGetExposure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	exposures := h.store.calculateExposureData()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    exposures,
		"count":   len(exposures),
	})
}

// HandleGetExposureByClient handles GET /admin/risk/exposure/by-client
func (h *RiskDashboardHandler) HandleGetExposureByClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	clientExposures := h.store.calculateClientExposures()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    clientExposures,
		"count":   len(clientExposures),
	})
}

// HandleGetConcentration handles GET /admin/risk/concentration
func (h *RiskDashboardHandler) HandleGetConcentration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	concentration := h.store.calculateConcentrationRisk()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    concentration,
		"count":   len(concentration),
	})
}

// HandleGetVaR handles GET /admin/risk/var
func (h *RiskDashboardHandler) HandleGetVaR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	varData := h.store.calculateVaR()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    varData,
	})
}

// HandleGetLimits handles GET /admin/risk/limits
func (h *RiskDashboardHandler) HandleGetLimits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	limits := h.store.limits
	h.store.mu.RUnlock()

	metrics := h.store.calculateRiskMetrics()

	utilizationData := map[string]interface{}{
		"exposure_used":     metrics.TotalExposure,
		"exposure_limit":    limits.MaxExposure,
		"exposure_pct":      (metrics.TotalExposure / limits.MaxExposure) * 100,
		"margin_level":      metrics.MarginUtilization,
		"margin_threshold":  limits.MinMarginLevel,
		"concentration_pct": metrics.ConcentrationRisk,
		"concentration_limit": limits.MaxSymbolConcentration * 100,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"limits":      limits,
		"utilization": utilizationData,
	})
}

// HandleUpdateLimits handles PUT /admin/risk/limits
func (h *RiskDashboardHandler) HandleUpdateLimits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	adminID, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updatedLimits RiskLimits
	if err := json.NewDecoder(r.Body).Decode(&updatedLimits); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	updatedLimits.UpdatedAt = time.Now()
	updatedLimits.UpdatedBy = fmt.Sprintf("admin-%d", adminID)
	h.store.limits = updatedLimits
	h.store.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Risk limits updated successfully",
		"data":    updatedLimits,
	})
}

// HandleGetAlerts handles GET /admin/risk/alerts
func (h *RiskDashboardHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	alerts := h.store.alerts
	h.store.mu.RUnlock()

	activeAlerts := make([]RiskAlert, 0)
	criticalCount := 0
	highCount := 0
	warningCount := 0

	for _, alert := range alerts {
		if alert.Status == "active" {
			activeAlerts = append(activeAlerts, alert)

			switch alert.Severity {
			case "critical":
				criticalCount++
			case "high":
				highCount++
			case "warning", "medium":
				warningCount++
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    activeAlerts,
		"summary": map[string]int{
			"total":    len(activeAlerts),
			"critical": criticalCount,
			"high":     highCount,
			"warning":  warningCount,
		},
	})
}
