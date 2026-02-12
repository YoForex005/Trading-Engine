package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

type AggregatedPosition struct {
	Symbol         string  `json:"symbol"`
	NetVolume      float64 `json:"netVolume"`      // lots (positive = net long, negative = net short)
	LongVolume     float64 `json:"longVolume"`     // total long lots
	ShortVolume    float64 `json:"shortVolume"`    // total short lots
	AvgEntryPrice  float64 `json:"avgEntryPrice"`
	CurrentPrice   float64 `json:"currentPrice"`
	UnrealizedPnL  float64 `json:"unrealizedPnL"`  // USD
	PositionCount  int     `json:"positionCount"`
	LongCount      int     `json:"longCount"`
	ShortCount     int     `json:"shortCount"`
	LastUpdate     time.Time `json:"lastUpdate"`
}

type IndividualPosition struct {
	ID            int64     `json:"id"`
	ClientID      int64     `json:"clientId"`
	ClientName    string    `json:"clientName"`
	Symbol        string    `json:"symbol"`
	Direction     string    `json:"direction"` // "long" or "short"
	Volume        float64   `json:"volume"`    // lots
	EntryPrice    float64   `json:"entryPrice"`
	CurrentPrice  float64   `json:"currentPrice"`
	UnrealizedPnL float64   `json:"unrealizedPnL"` // USD
	OpenTime      time.Time `json:"openTime"`
	GroupID       int64     `json:"groupId"`
}

type SymbolBreakdown struct {
	Symbol         string               `json:"symbol"`
	NetVolume      float64              `json:"netVolume"`
	LongVolume     float64              `json:"longVolume"`
	ShortVolume    float64              `json:"shortVolume"`
	LongPositions  []IndividualPosition `json:"longPositions"`
	ShortPositions []IndividualPosition `json:"shortPositions"`
	AvgEntryPrice  float64              `json:"avgEntryPrice"`
	CurrentPrice   float64              `json:"currentPrice"`
	UnrealizedPnL  float64              `json:"unrealizedPnL"`
}

type NettingModeConfig struct {
	GroupID     int64     `json:"groupId"`
	GroupName   string    `json:"groupName"`
	NettingMode string    `json:"nettingMode"` // "hedging" or "netting"
	Description string    `json:"description"`
	ClientCount int       `json:"clientCount"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AggregatedExposureData struct {
	Symbol        string  `json:"symbol"`
	NetVolume     float64 `json:"netVolume"`     // lots
	NetExposureUSD float64 `json:"netExposureUsd"` // USD
	HedgingRatio  float64 `json:"hedgingRatio"`  // 0-100% (how much is hedged)
	ExposedAmount float64 `json:"exposedAmount"` // USD not hedged
	RiskLevel     string  `json:"riskLevel"`     // "low", "medium", "high"
}

type LargestPosition struct {
	ID            int64   `json:"id"`
	ClientID      int64   `json:"clientId"`
	ClientName    string  `json:"clientName"`
	Symbol        string  `json:"symbol"`
	Direction     string  `json:"direction"`
	Volume        float64 `json:"volume"` // lots
	VolumeUSD     float64 `json:"volumeUsd"`
	UnrealizedPnL float64 `json:"unrealizedPnL"`
}

type ConcentrationRisk struct {
	Symbol             string  `json:"symbol"`
	TotalVolume        float64 `json:"totalVolume"`        // lots
	PercentOfTotal     float64 `json:"percentOfTotal"`     // %
	ConcentrationLevel string  `json:"concentrationLevel"` // "low", "medium", "high", "critical"
	RiskFlag           bool    `json:"riskFlag"`           // true if >30%
}

type PositionStats struct {
	TotalOpenPositions int     `json:"totalOpenPositions"`
	TotalVolumeLots    float64 `json:"totalVolumeLots"`
	TotalVolumeUSD     float64 `json:"totalVolumeUsd"`
	NetLongVolume      float64 `json:"netLongVolume"`
	NetShortVolume     float64 `json:"netShortVolume"`
	LongBias           float64 `json:"longBias"`      // % of total that is long
	ShortBias          float64 `json:"shortBias"`     // % of total that is short
	SymbolsWithPositions int   `json:"symbolsWithPositions"`
	AvgPositionSize    float64 `json:"avgPositionSize"` // lots
	TotalUnrealizedPnL float64 `json:"totalUnrealizedPnL"` // USD
}

// ============================================
// Service
// ============================================

type PositionAggregationService struct {
	positions     map[int64]*IndividualPosition
	nettingConfig map[int64]*NettingModeConfig
	nextPositionID int64
	mu            sync.RWMutex
}

func NewPositionAggregationService() *PositionAggregationService {
	service := &PositionAggregationService{
		positions:     make(map[int64]*IndividualPosition),
		nettingConfig: make(map[int64]*NettingModeConfig),
		nextPositionID: 1,
	}
	service.initializeMockData()
	return service
}

func (s *PositionAggregationService) initializeMockData() {
	now := time.Now()

	// ============================================
	// Initialize 5 Trading Groups with Netting Modes
	// ============================================
	groups := []NettingModeConfig{
		{
			GroupID:     1,
			GroupName:   "Retail Traders",
			NettingMode: "netting",
			Description: "Retail clients with netting mode (one position per symbol)",
			ClientCount: 40,
			UpdatedAt:   now.AddDate(0, -3, 0),
		},
		{
			GroupID:     2,
			GroupName:   "Professional Traders",
			NettingMode: "hedging",
			Description: "Professional clients with hedging mode (multiple positions allowed)",
			ClientCount: 30,
			UpdatedAt:   now.AddDate(0, -2, 0),
		},
		{
			GroupID:     3,
			GroupName:   "Institutional",
			NettingMode: "hedging",
			Description: "Institutional clients with full hedging capabilities",
			ClientCount: 20,
			UpdatedAt:   now.AddDate(0, -1, 0),
		},
		{
			GroupID:     4,
			GroupName:   "VIP Clients",
			NettingMode: "hedging",
			Description: "VIP clients with premium account features",
			ClientCount: 8,
			UpdatedAt:   now.AddDate(0, 0, -15),
		},
		{
			GroupID:     5,
			GroupName:   "Demo Accounts",
			NettingMode: "netting",
			Description: "Demo trading accounts",
			ClientCount: 2,
			UpdatedAt:   now.AddDate(0, -6, 0),
		},
	}

	for i := range groups {
		s.nettingConfig[groups[i].GroupID] = &groups[i]
	}

	// ============================================
	// Initialize 500 Positions across 100 clients and 30 symbols
	// ============================================
	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "USDCHF", "NZDUSD", "EURGBP",
		"EURJPY", "GBPJPY", "AUDJPY", "EURAUD", "EURCHF", "AUDCAD", "GBPAUD", "GBPCAD",
		"GBPCHF", "AUDCHF", "NZDJPY", "EURCAD", "BTCUSD", "ETHUSD", "XAUUSD", "XAGUSD",
		"US30", "NAS100", "SPX500", "UK100", "DE40", "JP225",
	}

	directions := []string{"long", "short"}

	// Base prices for realistic mock data
	basePrices := map[string]float64{
		"EURUSD": 1.0850, "GBPUSD": 1.2650, "USDJPY": 148.50, "AUDUSD": 0.6550,
		"USDCAD": 1.3650, "USDCHF": 0.8850, "NZDUSD": 0.5950, "EURGBP": 0.8550,
		"EURJPY": 161.00, "GBPJPY": 187.80, "AUDJPY": 97.30, "EURAUD": 1.6550,
		"EURCHF": 0.9600, "AUDCAD": 0.8950, "GBPAUD": 1.9300, "GBPCAD": 1.7250,
		"GBPCHF": 1.1200, "AUDCHF": 0.5800, "NZDJPY": 88.35, "EURCAD": 1.4800,
		"BTCUSD": 42000.0, "ETHUSD": 2200.0, "XAUUSD": 2050.0, "XAGUSD": 24.50,
		"US30": 38500.0, "NAS100": 15800.0, "SPX500": 4850.0, "UK100": 7600.0,
		"DE40": 17200.0, "JP225": 33500.0,
	}

	positionID := int64(1)
	for i := 0; i < 500; i++ {
		clientID := int64((i % 100) + 1)
		symbol := symbols[i%len(symbols)]
		direction := directions[rand.Intn(2)]
		groupID := int64((i % 5) + 1)

		// Realistic position sizes
		volume := 0.01 + rand.Float64()*5.0 // 0.01 to 5.01 lots
		if symbol == "BTCUSD" || symbol == "ETHUSD" {
			volume = 0.01 + rand.Float64()*0.5 // Smaller sizes for crypto
		}

		basePrice := basePrices[symbol]
		entryPrice := basePrice * (1 + (rand.Float64()-0.5)*0.02) // ±1% from base
		currentPrice := basePrice * (1 + (rand.Float64()-0.5)*0.01) // Current price varies ±0.5%

		// Calculate unrealized P&L
		var pnl float64
		if direction == "long" {
			pnl = (currentPrice - entryPrice) * volume * 100000 // Assuming standard lot size
		} else {
			pnl = (entryPrice - currentPrice) * volume * 100000
		}

		// For indices and commodities, adjust lot calculation
		if symbol == "XAUUSD" || symbol == "XAGUSD" || strings.Contains(symbol, "US30") ||
		   strings.Contains(symbol, "NAS") || strings.Contains(symbol, "SPX") ||
		   strings.Contains(symbol, "UK100") || strings.Contains(symbol, "DE40") ||
		   strings.Contains(symbol, "JP225") {
			if direction == "long" {
				pnl = (currentPrice - entryPrice) * volume
			} else {
				pnl = (entryPrice - currentPrice) * volume
			}
		}

		// For crypto
		if symbol == "BTCUSD" || symbol == "ETHUSD" {
			if direction == "long" {
				pnl = (currentPrice - entryPrice) * volume
			} else {
				pnl = (entryPrice - currentPrice) * volume
			}
		}

		openTime := now.AddDate(0, 0, -rand.Intn(30)) // Opened within last 30 days

		position := &IndividualPosition{
			ID:            positionID,
			ClientID:      clientID,
			ClientName:    "Client" + strconv.FormatInt(clientID, 10),
			Symbol:        symbol,
			Direction:     direction,
			Volume:        volume,
			EntryPrice:    entryPrice,
			CurrentPrice:  currentPrice,
			UnrealizedPnL: pnl,
			OpenTime:      openTime,
			GroupID:       groupID,
		}

		s.positions[positionID] = position
		positionID++
	}

	s.nextPositionID = positionID

	log.Printf("[PositionAggregationService] Initialized with %d positions across 100 clients, 30 symbols, 5 trading groups",
		len(s.positions))
}

// ============================================
// Service Methods
// ============================================

func (s *PositionAggregationService) GetAggregatedPositions() []AggregatedPosition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Aggregate by symbol
	aggregated := make(map[string]*AggregatedPosition)

	for _, pos := range s.positions {
		if _, exists := aggregated[pos.Symbol]; !exists {
			aggregated[pos.Symbol] = &AggregatedPosition{
				Symbol:     pos.Symbol,
				LastUpdate: time.Now(),
			}
		}

		agg := aggregated[pos.Symbol]
		agg.PositionCount++
		agg.UnrealizedPnL += pos.UnrealizedPnL

		if pos.Direction == "long" {
			agg.LongVolume += pos.Volume
			agg.LongCount++
		} else {
			agg.ShortVolume += pos.Volume
			agg.ShortCount++
		}
	}

	// Calculate net volume and average entry price
	for symbol, agg := range aggregated {
		agg.NetVolume = agg.LongVolume - agg.ShortVolume

		// Calculate weighted average entry price
		totalWeightedPrice := 0.0
		totalVolume := 0.0
		for _, pos := range s.positions {
			if pos.Symbol == symbol {
				totalWeightedPrice += pos.EntryPrice * pos.Volume
				totalVolume += pos.Volume
				agg.CurrentPrice = pos.CurrentPrice // Use any position's current price
			}
		}
		if totalVolume > 0 {
			agg.AvgEntryPrice = totalWeightedPrice / totalVolume
		}
	}

	// Convert to slice
	result := make([]AggregatedPosition, 0, len(aggregated))
	for _, agg := range aggregated {
		result = append(result, *agg)
	}

	// Sort by absolute net volume descending
	sort.Slice(result, func(i, j int) bool {
		absI := result[i].NetVolume
		if absI < 0 {
			absI = -absI
		}
		absJ := result[j].NetVolume
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	return result
}

func (s *PositionAggregationService) GetSymbolBreakdown(symbol string) *SymbolBreakdown {
	s.mu.RLock()
	defer s.mu.RUnlock()

	breakdown := &SymbolBreakdown{
		Symbol:         symbol,
		LongPositions:  []IndividualPosition{},
		ShortPositions: []IndividualPosition{},
	}

	for _, pos := range s.positions {
		if pos.Symbol == symbol {
			breakdown.UnrealizedPnL += pos.UnrealizedPnL
			breakdown.CurrentPrice = pos.CurrentPrice

			if pos.Direction == "long" {
				breakdown.LongVolume += pos.Volume
				breakdown.LongPositions = append(breakdown.LongPositions, *pos)
			} else {
				breakdown.ShortVolume += pos.Volume
				breakdown.ShortPositions = append(breakdown.ShortPositions, *pos)
			}
		}
	}

	breakdown.NetVolume = breakdown.LongVolume - breakdown.ShortVolume

	// Calculate weighted average entry
	totalWeightedPrice := 0.0
	totalVolume := 0.0
	for _, pos := range breakdown.LongPositions {
		totalWeightedPrice += pos.EntryPrice * pos.Volume
		totalVolume += pos.Volume
	}
	for _, pos := range breakdown.ShortPositions {
		totalWeightedPrice += pos.EntryPrice * pos.Volume
		totalVolume += pos.Volume
	}
	if totalVolume > 0 {
		breakdown.AvgEntryPrice = totalWeightedPrice / totalVolume
	}

	return breakdown
}

func (s *PositionAggregationService) GetNettingModes() []NettingModeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	modes := make([]NettingModeConfig, 0, len(s.nettingConfig))
	for _, config := range s.nettingConfig {
		modes = append(modes, *config)
	}

	sort.Slice(modes, func(i, j int) bool {
		return modes[i].GroupID < modes[j].GroupID
	})

	return modes
}

func (s *PositionAggregationService) UpdateNettingMode(groupID int64, mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.nettingConfig[groupID]
	if !exists {
		return nil // Group not found
	}

	config.NettingMode = mode
	config.UpdatedAt = time.Now()

	return nil
}

func (s *PositionAggregationService) GetExposure() []AggregatedExposureData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Aggregate net exposure by symbol
	exposure := make(map[string]*AggregatedExposureData)

	for _, pos := range s.positions {
		if _, exists := exposure[pos.Symbol]; !exists {
			exposure[pos.Symbol] = &AggregatedExposureData{
				Symbol: pos.Symbol,
			}
		}

		exp := exposure[pos.Symbol]
		volumeUSD := pos.Volume * pos.CurrentPrice * 100000

		if pos.Direction == "long" {
			exp.NetVolume += pos.Volume
			exp.NetExposureUSD += volumeUSD
		} else {
			exp.NetVolume -= pos.Volume
			exp.NetExposureUSD -= volumeUSD
		}
	}

	// Calculate hedging ratio and risk level
	result := make([]AggregatedExposureData, 0, len(exposure))
	for _, exp := range exposure {
		// Simplified hedging ratio (assume 50% is hedged by LP)
		hedgingRatio := 50.0 + rand.Float64()*30 // 50-80%
		exp.HedgingRatio = hedgingRatio
		exp.ExposedAmount = exp.NetExposureUSD * (100 - hedgingRatio) / 100

		// Determine risk level based on exposed amount
		absExposed := exp.ExposedAmount
		if absExposed < 0 {
			absExposed = -absExposed
		}

		if absExposed < 50000 {
			exp.RiskLevel = "low"
		} else if absExposed < 200000 {
			exp.RiskLevel = "medium"
		} else {
			exp.RiskLevel = "high"
		}

		result = append(result, *exp)
	}

	// Sort by absolute exposed amount descending
	sort.Slice(result, func(i, j int) bool {
		absI := result[i].ExposedAmount
		if absI < 0 {
			absI = -absI
		}
		absJ := result[j].ExposedAmount
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	return result
}

func (s *PositionAggregationService) GetLargestPositions() []LargestPosition {
	s.mu.RLock()
	defer s.mu.RUnlock()

	positions := make([]LargestPosition, 0, len(s.positions))

	for _, pos := range s.positions {
		volumeUSD := pos.Volume * pos.CurrentPrice * 100000

		positions = append(positions, LargestPosition{
			ID:            pos.ID,
			ClientID:      pos.ClientID,
			ClientName:    pos.ClientName,
			Symbol:        pos.Symbol,
			Direction:     pos.Direction,
			Volume:        pos.Volume,
			VolumeUSD:     volumeUSD,
			UnrealizedPnL: pos.UnrealizedPnL,
		})
	}

	// Sort by volume USD descending
	sort.Slice(positions, func(i, j int) bool {
		absI := positions[i].VolumeUSD
		if absI < 0 {
			absI = -absI
		}
		absJ := positions[j].VolumeUSD
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	// Return top 20
	if len(positions) > 20 {
		positions = positions[:20]
	}

	return positions
}

func (s *PositionAggregationService) GetConcentrationRisk() []ConcentrationRisk {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate total volume across all positions
	totalVolume := 0.0
	symbolVolume := make(map[string]float64)

	for _, pos := range s.positions {
		totalVolume += pos.Volume
		symbolVolume[pos.Symbol] += pos.Volume
	}

	// Calculate concentration percentages
	risks := make([]ConcentrationRisk, 0, len(symbolVolume))

	for symbol, volume := range symbolVolume {
		percent := (volume / totalVolume) * 100

		var level string
		var flag bool

		if percent > 30 {
			level = "critical"
			flag = true
		} else if percent > 20 {
			level = "high"
			flag = true
		} else if percent > 10 {
			level = "medium"
			flag = false
		} else {
			level = "low"
			flag = false
		}

		risks = append(risks, ConcentrationRisk{
			Symbol:             symbol,
			TotalVolume:        volume,
			PercentOfTotal:     percent,
			ConcentrationLevel: level,
			RiskFlag:           flag,
		})
	}

	// Sort by percent descending
	sort.Slice(risks, func(i, j int) bool {
		return risks[i].PercentOfTotal > risks[j].PercentOfTotal
	})

	return risks
}

func (s *PositionAggregationService) GetStats() PositionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := PositionStats{
		TotalOpenPositions: len(s.positions),
	}

	symbolMap := make(map[string]bool)
	longVolume := 0.0
	shortVolume := 0.0

	for _, pos := range s.positions {
		stats.TotalVolumeLots += pos.Volume
		stats.TotalVolumeUSD += pos.Volume * pos.CurrentPrice * 100000
		stats.TotalUnrealizedPnL += pos.UnrealizedPnL
		symbolMap[pos.Symbol] = true

		if pos.Direction == "long" {
			longVolume += pos.Volume
			stats.NetLongVolume += pos.Volume
		} else {
			shortVolume += pos.Volume
			stats.NetShortVolume += pos.Volume
		}
	}

	stats.SymbolsWithPositions = len(symbolMap)

	if stats.TotalOpenPositions > 0 {
		stats.AvgPositionSize = stats.TotalVolumeLots / float64(stats.TotalOpenPositions)
	}

	totalVolume := longVolume + shortVolume
	if totalVolume > 0 {
		stats.LongBias = (longVolume / totalVolume) * 100
		stats.ShortBias = (shortVolume / totalVolume) * 100
	}

	return stats
}

// ============================================
// HTTP Handlers
// ============================================

type PositionAggregationHandler struct {
	service     *PositionAggregationService
	authService *auth.Service
}

func NewPositionAggregationHandler(service *PositionAggregationService, authService *auth.Service) *PositionAggregationHandler {
	return &PositionAggregationHandler{
		service:     service,
		authService: authService,
	}
}

// 1. GET /admin/positions/aggregated - Aggregated positions per symbol
func (h *PositionAggregationHandler) HandleGetAggregated(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	positions := h.service.GetAggregatedPositions()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions": positions,
		"count":     len(positions),
	})
}

// 2. GET /admin/positions/aggregated/:symbol - Symbol breakdown
func (h *PositionAggregationHandler) HandleGetSymbolBreakdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/admin/positions/aggregated/")
	if symbol == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	breakdown := h.service.GetSymbolBreakdown(symbol)
	json.NewEncoder(w).Encode(breakdown)
}

// 3. GET /admin/positions/netting-mode - Get netting configurations
func (h *PositionAggregationHandler) HandleGetNettingModes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	modes := h.service.GetNettingModes()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": modes,
		"count":  len(modes),
	})
}

// 4. PUT /admin/positions/netting-mode/:groupId - Update netting mode
func (h *PositionAggregationHandler) HandleUpdateNettingMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := strings.TrimPrefix(r.URL.Path, "/admin/positions/netting-mode/")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req struct {
		NettingMode string `json:"nettingMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateNettingMode(groupID, req.NettingMode); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Netting mode updated successfully",
		"groupId": groupID,
		"mode":    req.NettingMode,
	})
}

// 5. GET /admin/positions/exposure - Net exposure per symbol
func (h *PositionAggregationHandler) HandleGetExposure(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	exposure := h.service.GetExposure()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exposure": exposure,
		"count":    len(exposure),
	})
}

// 6. GET /admin/positions/largest - Top 20 largest positions
func (h *PositionAggregationHandler) HandleGetLargest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	positions := h.service.GetLargestPositions()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"positions": positions,
		"count":     len(positions),
	})
}

// 7. GET /admin/positions/concentration - Concentration risk
func (h *PositionAggregationHandler) HandleGetConcentration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	risks := h.service.GetConcentrationRisk()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"risks": risks,
		"count": len(risks),
	})
}

// 8. GET /admin/positions/stats - Aggregate statistics
func (h *PositionAggregationHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}
