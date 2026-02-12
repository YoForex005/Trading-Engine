package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// SwapRateConfig represents swap/rollover rate configuration for a symbol
type SwapRateConfig struct {
	Symbol           string    `json:"symbol"`
	SwapLong         float64   `json:"swap_long"`          // Swap rate for long positions (pips)
	SwapShort        float64   `json:"swap_short"`         // Swap rate for short positions (pips)
	TripleSwapDay    string    `json:"triple_swap_day"`    // Day of week for triple swap (Monday-Sunday)
	SwapFreeEligible bool      `json:"swap_free_eligible"` // If true, swap-free accounts get 0 swap
	ContractSize     float64   `json:"contract_size"`      // Contract size for swap calculation
	Enabled          bool      `json:"enabled"`            // If false, no swap is charged
	UpdatedAt        time.Time `json:"updated_at"`
	UpdatedBy        string    `json:"updated_by"`
}

// SwapPreviewRequest represents a request to preview swap calculation
type SwapPreviewRequest struct {
	Symbol    string  `json:"symbol"`
	Direction string  `json:"direction"` // BUY or SELL
	Volume    float64 `json:"volume"`
}

// SwapPreviewResponse represents swap calculation preview
type SwapPreviewResponse struct {
	Symbol           string  `json:"symbol"`
	Direction        string  `json:"direction"`
	Volume           float64 `json:"volume"`
	SwapRate         float64 `json:"swap_rate"`
	DailySwap        float64 `json:"daily_swap"`
	TripleSwap       float64 `json:"triple_swap"`
	WeeklySwap       float64 `json:"weekly_swap"`        // 5 days + 1 triple swap
	MonthlySwap      float64 `json:"monthly_swap"`       // Approx 30 days (26 normal + 4 triple)
	ContractSize     float64 `json:"contract_size"`
	SwapFreeEligible bool    `json:"swap_free_eligible"`
}

// SwapRateStore manages swap rate configurations
type SwapRateStore struct {
	mu      sync.RWMutex
	configs map[string]*SwapRateConfig
}

func NewSwapRateStore() *SwapRateStore {
	store := &SwapRateStore{
		configs: make(map[string]*SwapRateConfig),
	}

	// Initialize with 30+ realistic swap rates
	store.initializeDefaultRates()

	log.Println("[SwapRates] Swap rate configuration system initialized (30+ symbols with realistic rates)")

	return store
}

func (s *SwapRateStore) initializeDefaultRates() {
	now := time.Now()

	// Forex pairs - typical swap rates in points
	forexPairs := map[string][2]float64{
		// Major pairs
		"EURUSD": {-0.65, 0.15},   // Negative swap long, positive short
		"GBPUSD": {-0.80, 0.20},
		"USDJPY": {0.10, -0.75},   // Positive long (buying USD = earning), negative short
		"USDCHF": {0.15, -0.70},
		"AUDUSD": {-0.50, 0.10},
		"USDCAD": {0.05, -0.60},
		"NZDUSD": {-0.55, 0.12},

		// Minor/cross pairs
		"EURGBP": {-0.45, -0.20},  // Both negative
		"EURJPY": {-0.55, 0.25},
		"GBPJPY": {-0.90, 0.30},
		"AUDJPY": {-0.40, 0.18},
		"EURAUD": {-0.70, 0.22},
		"EURCHF": {-0.50, -0.15},
		"GBPCHF": {-0.85, 0.28},
		"CADCHF": {-0.35, -0.10},
		"CHFJPY": {-0.30, -0.05},
		"AUDCAD": {-0.45, 0.08},
		"AUDNZD": {-0.40, -0.12},
		"EURCZK": {-2.50, 1.80},   // Exotic pairs have higher swaps
		"EURTRY": {-15.0, 8.50},   // Turkish lira has very high swap
		"USDMXN": {-8.50, 4.20},
	}

	// Metals - typically negative on both sides due to storage costs
	metals := map[string][2]float64{
		"XAUUSD": {-2.50, -1.80}, // Gold
		"XAGUSD": {-0.80, -0.60}, // Silver
		"XCUUSD": {-1.20, -0.90}, // Copper
		"XPTUSD": {-3.00, -2.20}, // Platinum
		"XPDUSD": {-3.50, -2.50}, // Palladium
	}

	// Crypto - very high swaps both ways
	crypto := map[string][2]float64{
		"BTCUSD": {-25.0, -20.0}, // Bitcoin
		"ETHUSD": {-15.0, -12.0}, // Ethereum
		"XRPUSD": {-8.00, -6.50}, // Ripple
		"BNBUSD": {-10.0, -8.00}, // Binance Coin
		"SOLUSD": {-12.0, -9.50}, // Solana
	}

	// Indices - typically negative on both sides
	indices := map[string][2]float64{
		"SPX500USD": {-0.50, -0.40},
		"NAS100USD": {-0.60, -0.45},
		"US30USD":   {-0.55, -0.42},
		"UK100GBP":  {-0.48, -0.38},
		"DE30EUR":   {-0.52, -0.40},
		"JP225USD":  {-0.45, -0.35},
	}

	// Add all forex pairs
	for symbol, rates := range forexPairs {
		s.configs[symbol] = &SwapRateConfig{
			Symbol:           symbol,
			SwapLong:         rates[0],
			SwapShort:        rates[1],
			TripleSwapDay:    "Wednesday",
			SwapFreeEligible: true, // Most forex is swap-free eligible
			ContractSize:     100000,
			Enabled:          true,
			UpdatedAt:        now,
			UpdatedBy:        "system",
		}
	}

	// Add metals
	for symbol, rates := range metals {
		s.configs[symbol] = &SwapRateConfig{
			Symbol:           symbol,
			SwapLong:         rates[0],
			SwapShort:        rates[1],
			TripleSwapDay:    "Wednesday",
			SwapFreeEligible: false, // Metals typically not swap-free
			ContractSize:     100,
			Enabled:          true,
			UpdatedAt:        now,
			UpdatedBy:        "system",
		}
	}

	// Add crypto
	for symbol, rates := range crypto {
		s.configs[symbol] = &SwapRateConfig{
			Symbol:           symbol,
			SwapLong:         rates[0],
			SwapShort:        rates[1],
			TripleSwapDay:    "Friday", // Crypto triple swap on Friday
			SwapFreeEligible: false,     // Crypto never swap-free
			ContractSize:     1,
			Enabled:          true,
			UpdatedAt:        now,
			UpdatedBy:        "system",
		}
	}

	// Add indices
	for symbol, rates := range indices {
		s.configs[symbol] = &SwapRateConfig{
			Symbol:           symbol,
			SwapLong:         rates[0],
			SwapShort:        rates[1],
			TripleSwapDay:    "Wednesday",
			SwapFreeEligible: false, // Indices typically not swap-free
			ContractSize:     10,
			Enabled:          true,
			UpdatedAt:        now,
			UpdatedBy:        "system",
		}
	}
}

// GetAll returns all swap rate configurations
func (s *SwapRateStore) GetAll() []*SwapRateConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := make([]*SwapRateConfig, 0, len(s.configs))
	for _, config := range s.configs {
		configCopy := *config
		configs = append(configs, &configCopy)
	}

	return configs
}

// GetBySymbol returns swap rate config for a specific symbol
func (s *SwapRateStore) GetBySymbol(symbol string) (*SwapRateConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[symbol]
	if !exists {
		return nil, false
	}

	configCopy := *config
	return &configCopy, true
}

// Update updates swap rate configuration for a symbol
func (s *SwapRateStore) Update(symbol string, updates *SwapRateConfig, updatedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.configs[symbol]
	if !exists {
		// Create new config if doesn't exist
		config = &SwapRateConfig{
			Symbol:       symbol,
			ContractSize: 100000, // Default forex contract size
		}
		s.configs[symbol] = config
	}

	// Update fields
	if updates.SwapLong != 0 || updates.SwapLong == config.SwapLong {
		config.SwapLong = updates.SwapLong
	}
	if updates.SwapShort != 0 || updates.SwapShort == config.SwapShort {
		config.SwapShort = updates.SwapShort
	}
	if updates.TripleSwapDay != "" {
		config.TripleSwapDay = updates.TripleSwapDay
	}
	if updates.ContractSize > 0 {
		config.ContractSize = updates.ContractSize
	}

	config.SwapFreeEligible = updates.SwapFreeEligible
	config.Enabled = updates.Enabled
	config.UpdatedAt = time.Now()
	config.UpdatedBy = updatedBy

	log.Printf("[SwapRates] Updated swap rates for %s: long=%.2f, short=%.2f", symbol, config.SwapLong, config.SwapShort)
	return nil
}

// BulkUpdate updates multiple symbols at once
func (s *SwapRateStore) BulkUpdate(updates []*SwapRateConfig, updatedBy string) (int, error) {
	successCount := 0

	for _, update := range updates {
		if err := s.Update(update.Symbol, update, updatedBy); err != nil {
			log.Printf("[SwapRates] Failed to update %s: %v", update.Symbol, err)
			continue
		}
		successCount++
	}

	log.Printf("[SwapRates] Bulk update completed: %d/%d symbols updated", successCount, len(updates))
	return successCount, nil
}

// CalculateSwapPreview calculates swap amounts for preview
func (s *SwapRateStore) CalculateSwapPreview(symbol, direction string, volume float64) (*SwapPreviewResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[symbol]
	if !exists {
		return nil, fmt.Errorf("swap rate config not found for symbol: %s", symbol)
	}

	if !config.Enabled {
		return &SwapPreviewResponse{
			Symbol:           symbol,
			Direction:        direction,
			Volume:           volume,
			SwapRate:         0,
			DailySwap:        0,
			TripleSwap:       0,
			WeeklySwap:       0,
			MonthlySwap:      0,
			ContractSize:     config.ContractSize,
			SwapFreeEligible: config.SwapFreeEligible,
		}, nil
	}

	// Get swap rate based on direction
	var swapRate float64
	directionUpper := strings.ToUpper(direction)
	if directionUpper == "BUY" || directionUpper == "LONG" {
		swapRate = config.SwapLong
	} else if directionUpper == "SELL" || directionUpper == "SHORT" {
		swapRate = config.SwapShort
	} else {
		return nil, fmt.Errorf("invalid direction: %s (must be BUY or SELL)", direction)
	}

	// Calculate swap amounts
	// Formula: swap = rate * volume * contractSize / 10
	dailySwap := swapRate * volume * config.ContractSize / 10
	tripleSwap := dailySwap * 3

	// Weekly: 5 regular days + 1 triple swap day = 5 + 3 = 8 days worth
	weeklySwap := (dailySwap * 5) + tripleSwap

	// Monthly: Approximately 26 regular days + 4 triple swap days = 26 + 12 = 38 days worth
	monthlySwap := (dailySwap * 26) + (tripleSwap * 4)

	return &SwapPreviewResponse{
		Symbol:           symbol,
		Direction:        direction,
		Volume:           volume,
		SwapRate:         swapRate,
		DailySwap:        dailySwap,
		TripleSwap:       tripleSwap,
		WeeklySwap:       weeklySwap,
		MonthlySwap:      monthlySwap,
		ContractSize:     config.ContractSize,
		SwapFreeEligible: config.SwapFreeEligible,
	}, nil
}

// SwapRateHandler handles HTTP requests for swap rate configuration
type SwapRateHandler struct {
	store       *SwapRateStore
	authService *auth.Service
}

func NewSwapRateHandler(store *SwapRateStore, authService *auth.Service) *SwapRateHandler {
	return &SwapRateHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListAll returns all swap rate configurations
// GET /admin/swap-rates
func (h *SwapRateHandler) HandleListAll(w http.ResponseWriter, r *http.Request) {
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

	configs := h.store.GetAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"swap_rates": configs,
		"count":      len(configs),
	})
}

// HandleGetBySymbol returns swap rate config for a specific symbol
// GET /admin/swap-rates/:symbol
func (h *SwapRateHandler) HandleGetBySymbol(w http.ResponseWriter, r *http.Request) {
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

	// Extract symbol from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	symbol := pathParts[3]

	config, exists := h.store.GetBySymbol(symbol)
	if !exists {
		http.Error(w, fmt.Sprintf("Swap rate config not found for symbol: %s", symbol), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// HandleUpdate updates swap rate configuration for a symbol
// PUT /admin/swap-rates/:symbol
func (h *SwapRateHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract symbol from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	symbol := pathParts[3]

	var updates SwapRateConfig
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Extract admin username from JWT token
	updatedBy := "admin"

	if err := h.store.Update(symbol, &updates, updatedBy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get updated config
	config, _ := h.store.GetBySymbol(symbol)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// HandleBulkUpdate updates multiple symbols at once
// POST /admin/swap-rates/bulk-update
func (h *SwapRateHandler) HandleBulkUpdate(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		SwapRates []*SwapRateConfig `json:"swap_rates"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.SwapRates) == 0 {
		http.Error(w, "No swap rates provided", http.StatusBadRequest)
		return
	}

	// TODO: Extract admin username from JWT token
	updatedBy := "admin"

	successCount, err := h.store.BulkUpdate(request.SwapRates, updatedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"updated_count": successCount,
		"total_count":   len(request.SwapRates),
	})
}

// HandlePreview calculates swap preview for a symbol
// GET /admin/swap-rates/preview/:symbol?direction=BUY&volume=1.0
func (h *SwapRateHandler) HandlePreview(w http.ResponseWriter, r *http.Request) {
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

	// Extract symbol from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	symbol := pathParts[4]

	// Get query parameters
	direction := r.URL.Query().Get("direction")
	volumeStr := r.URL.Query().Get("volume")

	if direction == "" || volumeStr == "" {
		http.Error(w, "Missing required parameters: direction, volume", http.StatusBadRequest)
		return
	}

	volume := 0.0
	if _, err := fmt.Sscanf(volumeStr, "%f", &volume); err != nil || volume <= 0 {
		http.Error(w, "Invalid volume", http.StatusBadRequest)
		return
	}

	preview, err := h.store.CalculateSwapPreview(symbol, direction, volume)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}
