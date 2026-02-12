package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/bbook"
	"github.com/epic1st/rtx/backend/internal/core"
)

// OrderBookLevel represents a single price level in the order book
type OrderBookLevel struct {
	Price            float64 `json:"price"`
	Volume           float64 `json:"volume"`
	OrderCount       int     `json:"orderCount"`
	CumulativeVolume float64 `json:"cumulativeVolume"`
}

// OrderBookSnapshot represents the full order book for a symbol
type OrderBookSnapshot struct {
	Symbol    string           `json:"symbol"`
	Timestamp time.Time        `json:"timestamp"`
	Bids      []OrderBookLevel `json:"bids"` // Descending order (highest first)
	Asks      []OrderBookLevel `json:"asks"` // Ascending order (lowest first)
	Spread    float64          `json:"spread"`
	MidPrice  float64          `json:"midPrice"`
}

// HubBroadcaster interface for WebSocket broadcasting
type HubBroadcaster interface {
	BroadcastMarketData(message []byte)
}

// OrderBookService generates synthetic order book data
type OrderBookService struct {
	mu            sync.RWMutex
	engine        *core.Engine
	priceCallback func(symbol string) (bid, ask float64, ok bool)
	snapshots     map[string]*OrderBookSnapshot
	levels        int           // Number of levels per side
	refreshRate   time.Duration // How often to refresh volumes
	stopChan      chan bool
	running       bool

	// WebSocket broadcasting
	hub               HubBroadcaster // WebSocket hub for broadcasting
	wsStopChan        chan bool
	wsRunning         bool
	subscribedSymbols []string // Symbols to broadcast order book for
}

// NewOrderBookService creates a new order book service
func NewOrderBookService(engine *core.Engine, priceCallback func(symbol string) (bid, ask float64, ok bool), levels int, refreshMS int) *OrderBookService {
	if levels <= 0 {
		levels = 20 // Default
	}
	if refreshMS <= 0 {
		refreshMS = 500 // Default
	}

	return &OrderBookService{
		engine:            engine,
		priceCallback:     priceCallback,
		snapshots:         make(map[string]*OrderBookSnapshot),
		levels:            levels,
		refreshRate:       time.Duration(refreshMS) * time.Millisecond,
		stopChan:          make(chan bool),
		running:           false,
		wsStopChan:        make(chan bool),
		wsRunning:         false,
		subscribedSymbols: []string{},
	}
}

// GetSnapshot returns the current order book snapshot for a symbol
func (obs *OrderBookService) GetSnapshot(symbol string) (*OrderBookSnapshot, error) {
	symbol = strings.ToUpper(symbol)

	// Get current price
	if obs.priceCallback == nil {
		return nil, fmt.Errorf("price feed not available")
	}

	bid, ask, ok := obs.priceCallback(symbol)
	if !ok {
		return nil, fmt.Errorf("no price available for %s", symbol)
	}

	// Get symbol spec for pip size and digits
	symbols := obs.engine.GetSymbols()
	var spec *bbook.SymbolSpec
	for _, s := range symbols {
		if s.Symbol == symbol {
			specCopy := bbook.SymbolSpec{
				Symbol:         s.Symbol,
				ContractSize:   s.ContractSize,
				PipSize:        s.PipSize,
				PipValue:       s.PipValue,
				MarginPercent:  s.MarginPercent,
				MinVolume:      s.MinVolume,
				MaxVolume:      s.MaxVolume,
				VolumeStep:     s.VolumeStep,
				SwapLong:       s.SwapLong,
				SwapShort:      s.SwapShort,
			}
			spec = &specCopy
			break
		}
	}

	if spec == nil {
		// Auto-generate spec if not found
		spec = bbook.GenerateSymbolSpec(symbol)
	}

	// Calculate digits from pip size
	digits := calculateDigits(spec.PipSize)

	// Generate order book levels
	bids := obs.generateLevels(bid, "bid", obs.levels, digits, spec.PipSize)
	asks := obs.generateLevels(ask, "ask", obs.levels, digits, spec.PipSize)

	spread := roundToDigits(ask-bid, digits)
	midPrice := roundToDigits((bid+ask)/2, digits)

	snapshot := &OrderBookSnapshot{
		Symbol:    symbol,
		Timestamp: time.Now(),
		Bids:      bids,
		Asks:      asks,
		Spread:    spread,
		MidPrice:  midPrice,
	}

	// Cache the snapshot
	obs.mu.Lock()
	obs.snapshots[symbol] = snapshot
	obs.mu.Unlock()

	return snapshot, nil
}

// generateLevels generates synthetic order book levels
func (obs *OrderBookService) generateLevels(basePrice float64, side string, levels, digits int, pipSize float64) []OrderBookLevel {
	result := make([]OrderBookLevel, levels)
	cumulativeVol := 0.0

	for i := 0; i < levels; i++ {
		// Calculate price level (spacing = 1 pip)
		var price float64
		if side == "bid" {
			// Bids go downward from base price
			price = basePrice - float64(i)*pipSize
		} else {
			// Asks go upward from base price
			price = basePrice + float64(i)*pipSize
		}
		price = roundToDigits(price, digits)

		// Generate realistic volume with exponential decay
		// Higher volume near market price, decreases as we move away
		baseVolume := 5.0 + rand.Float64()*10.0 // 5-15 lots base
		decayFactor := math.Exp(-float64(i) * 0.15)
		volume := baseVolume * decayFactor
		volume = math.Round(volume*100) / 100 // Round to 2 decimals

		// Random variation (±20%)
		variation := 0.8 + rand.Float64()*0.4
		volume *= variation

		// Order count (1-5 orders per level)
		orderCount := 1 + rand.Intn(5)

		cumulativeVol += volume

		result[i] = OrderBookLevel{
			Price:            price,
			Volume:           math.Round(volume*100) / 100,
			OrderCount:       orderCount,
			CumulativeVolume: math.Round(cumulativeVol*100) / 100,
		}
	}

	return result
}

// StartUpdater starts the background updater that refreshes volumes
func (obs *OrderBookService) StartUpdater() {
	obs.mu.Lock()
	if obs.running {
		obs.mu.Unlock()
		return
	}
	obs.running = true
	obs.mu.Unlock()

	go func() {
		ticker := time.NewTicker(obs.refreshRate)
		defer ticker.Stop()

		log.Printf("[OrderBook] Updater started (refresh every %v)", obs.refreshRate)

		for {
			select {
			case <-ticker.C:
				obs.refreshVolumes()

			case <-obs.stopChan:
				log.Println("[OrderBook] Updater stopped")
				obs.mu.Lock()
				obs.running = false
				obs.mu.Unlock()
				return
			}
		}
	}()
}

// StopUpdater stops the background updater
func (obs *OrderBookService) StopUpdater() {
	obs.mu.Lock()
	if !obs.running {
		obs.mu.Unlock()
		return
	}
	obs.mu.Unlock()

	close(obs.stopChan)
}

// refreshVolumes slightly varies the volumes in cached snapshots
func (obs *OrderBookService) refreshVolumes() {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	for symbol, snapshot := range obs.snapshots {
		// Apply small random changes to volumes (±10%)
		for i := range snapshot.Bids {
			variation := 0.9 + rand.Float64()*0.2
			snapshot.Bids[i].Volume *= variation
			snapshot.Bids[i].Volume = math.Round(snapshot.Bids[i].Volume*100) / 100
		}

		for i := range snapshot.Asks {
			variation := 0.9 + rand.Float64()*0.2
			snapshot.Asks[i].Volume *= variation
			snapshot.Asks[i].Volume = math.Round(snapshot.Asks[i].Volume*100) / 100
		}

		// Recalculate cumulative volumes
		cumBid := 0.0
		for i := range snapshot.Bids {
			cumBid += snapshot.Bids[i].Volume
			snapshot.Bids[i].CumulativeVolume = math.Round(cumBid*100) / 100
		}

		cumAsk := 0.0
		for i := range snapshot.Asks {
			cumAsk += snapshot.Asks[i].Volume
			snapshot.Asks[i].CumulativeVolume = math.Round(cumAsk*100) / 100
		}

		snapshot.Timestamp = time.Now()

		log.Printf("[OrderBook] Refreshed %s (Bid: %.2f lots, Ask: %.2f lots)", symbol, cumBid, cumAsk)
	}
}

// GetCachedSnapshot returns a cached snapshot (used by WebSocket broadcaster)
func (obs *OrderBookService) GetCachedSnapshot(symbol string) *OrderBookSnapshot {
	obs.mu.RLock()
	defer obs.mu.RUnlock()

	snapshot, ok := obs.snapshots[symbol]
	if !ok {
		return nil
	}

	return snapshot
}

// Helper: calculate digits from pip size
func calculateDigits(pipSize float64) int {
	if pipSize >= 1.0 {
		return 0 // Crypto BTC, indices (integer pricing)
	} else if pipSize >= 0.1 {
		return 1 // Indices with decimal
	} else if pipSize >= 0.01 {
		return 2 // JPY pairs, gold
	} else if pipSize >= 0.001 {
		return 3 // Silver
	} else if pipSize >= 0.0001 {
		return 4 // Forex majors (quote to 5 digits, but 4 after decimal)
	}
	return 5 // Default
}

// Helper: round to specific number of digits
func roundToDigits(value float64, digits int) float64 {
	multiplier := math.Pow(10, float64(digits))
	return math.Round(value*multiplier) / multiplier
}

// =========================================
// HTTP HANDLERS
// =========================================

// OrderBookHandler provides HTTP handlers for order book API
type OrderBookHandler struct {
	service *OrderBookService
}

// NewOrderBookHandler creates a new order book handler
func NewOrderBookHandler(service *OrderBookService) *OrderBookHandler {
	return &OrderBookHandler{
		service: service,
	}
}

// GetOrderBook handles GET /api/orderbook/:symbol
func (obh *OrderBookHandler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from path: /api/orderbook/:symbol
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	var symbol string
	for i, part := range parts {
		if part == "orderbook" && i+1 < len(parts) {
			symbol = parts[i+1]
			// Skip if it's "summary"
			if symbol == "summary" {
				continue
			}
			break
		}
	}

	if symbol == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	snapshot, err := obh.service.GetSnapshot(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"snapshot": snapshot,
	})
}

// GetOrderBookSummary handles GET /api/orderbook/:symbol/summary
func (obh *OrderBookHandler) GetOrderBookSummary(w http.ResponseWriter, r *http.Request) {
	// CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract symbol from path: /api/orderbook/:symbol/summary
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	var symbol string
	for i, part := range parts {
		if part == "orderbook" && i+1 < len(parts) {
			symbol = parts[i+1]
			break
		}
	}

	if symbol == "" {
		http.Error(w, "Symbol required", http.StatusBadRequest)
		return
	}

	snapshot, err := obh.service.GetSnapshot(symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get top 5 levels
	topBids := snapshot.Bids
	if len(topBids) > 5 {
		topBids = topBids[:5]
	}

	topAsks := snapshot.Asks
	if len(topAsks) > 5 {
		topAsks = topAsks[:5]
	}

	// Calculate total volumes
	totalBidVolume := 0.0
	for _, bid := range snapshot.Bids {
		totalBidVolume += bid.Volume
	}

	totalAskVolume := 0.0
	for _, ask := range snapshot.Asks {
		totalAskVolume += ask.Volume
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"symbol":         snapshot.Symbol,
		"timestamp":      snapshot.Timestamp,
		"topBids":        topBids,
		"topAsks":        topAsks,
		"spread":         snapshot.Spread,
		"midPrice":       snapshot.MidPrice,
		"totalBidVolume": math.Round(totalBidVolume*100) / 100,
		"totalAskVolume": math.Round(totalAskVolume*100) / 100,
	})
}

// SetWebSocketHub sets the WebSocket hub for broadcasting
func (obs *OrderBookService) SetWebSocketHub(hub HubBroadcaster, symbols []string) {
	obs.mu.Lock()
	defer obs.mu.Unlock()

	obs.hub = hub
	obs.subscribedSymbols = symbols
}

// StartWebSocketBroadcaster starts broadcasting order book updates via WebSocket
func (obs *OrderBookService) StartWebSocketBroadcaster() {
	obs.mu.Lock()
	if obs.wsRunning || obs.hub == nil {
		obs.mu.Unlock()
		return
	}
	obs.wsRunning = true
	obs.mu.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		log.Println("[OrderBook] WebSocket broadcaster started (1s interval)")

		for {
			select {
			case <-ticker.C:
				obs.broadcastOrderBooks()

			case <-obs.wsStopChan:
				log.Println("[OrderBook] WebSocket broadcaster stopped")
				obs.mu.Lock()
				obs.wsRunning = false
				obs.mu.Unlock()
				return
			}
		}
	}()
}

// StopWebSocketBroadcaster stops the WebSocket broadcaster
func (obs *OrderBookService) StopWebSocketBroadcaster() {
	obs.mu.Lock()
	if !obs.wsRunning {
		obs.mu.Unlock()
		return
	}
	obs.mu.Unlock()

	close(obs.wsStopChan)
}

// broadcastOrderBooks broadcasts order book updates for subscribed symbols
func (obs *OrderBookService) broadcastOrderBooks() {
	obs.mu.RLock()
	symbols := obs.subscribedSymbols
	hub := obs.hub
	obs.mu.RUnlock()

	if hub == nil || len(symbols) == 0 {
		return
	}

	for _, symbol := range symbols {
		// Get fresh snapshot (will generate if not cached)
		snapshot, err := obs.GetSnapshot(symbol)
		if err != nil {
			continue
		}

		// Get top 5 levels
		topBids := snapshot.Bids
		if len(topBids) > 5 {
			topBids = topBids[:5]
		}

		topAsks := snapshot.Asks
		if len(topAsks) > 5 {
			topAsks = topAsks[:5]
		}

		// Create WebSocket message
		message := map[string]interface{}{
			"type":      "orderbook_update",
			"symbol":    snapshot.Symbol,
			"timestamp": snapshot.Timestamp,
			"bids":      topBids,
			"asks":      topAsks,
			"spread":    snapshot.Spread,
			"midPrice":  snapshot.MidPrice,
		}

		// Marshal and broadcast via hub
		data, err := json.Marshal(message)
		if err != nil {
			log.Printf("[OrderBook] Failed to marshal %s: %v", symbol, err)
			continue
		}

		// Broadcast to all WebSocket clients
		hub.BroadcastMarketData(data)
	}
}

// GetOrderBookServiceFromEnv creates OrderBookService from environment variables
func GetOrderBookServiceFromEnv(engine *core.Engine, priceCallback func(symbol string) (bid, ask float64, ok bool)) *OrderBookService {
	levels := 20 // Default
	if envLevels := os.Getenv("ORDER_BOOK_LEVELS"); envLevels != "" {
		if parsed, err := strconv.Atoi(envLevels); err == nil && parsed > 0 {
			levels = parsed
		}
	}

	refreshMS := 500 // Default
	if envRefresh := os.Getenv("ORDER_BOOK_REFRESH_MS"); envRefresh != "" {
		if parsed, err := strconv.Atoi(envRefresh); err == nil && parsed > 0 {
			refreshMS = parsed
		}
	}

	return NewOrderBookService(engine, priceCallback, levels, refreshMS)
}
