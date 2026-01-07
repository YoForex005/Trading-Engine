package tickstore

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

// Tick represents a single market price update
type Tick struct {
	BrokerID  string    `json:"broker_id"`
	Symbol    string    `json:"symbol"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Spread    float64   `json:"spread"`
	Timestamp time.Time `json:"timestamp"`
	LP        string    `json:"lp"`
}

// OHLC represents a candlestick bar
type OHLC struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int     `json:"volume"`
}

// TickStore provides unified tick storage with file persistence
type TickStore struct {
	mu         sync.RWMutex
	ticks      map[string][]Tick // In-memory tick buffer (current session)
	maxTicks   int               // Max ticks per symbol in memory
	filePath   string            // Legacy file path
	brokerID   string
	dailyStore *DailyStore // Daily file rotation
	ohlcCache  *OHLCCache  // Pre-computed OHLC
}

// NewTickStore creates a new tick store instance
func NewTickStore(brokerID string, maxTicksPerSymbol int) *TickStore {
	ts := &TickStore{
		ticks:      make(map[string][]Tick),
		maxTicks:   maxTicksPerSymbol,
		filePath:   "data/ticks.json",
		brokerID:   brokerID,
		dailyStore: NewDailyStore(brokerID, 30), // Keep 30 days
		ohlcCache:  NewOHLCCache([]Timeframe{TF_M1, TF_M5, TF_M15, TF_H1, TF_H4, TF_D1}),
	}

	// Ensure data directory exists
	os.MkdirAll("data", 0755)

	// Load legacy ticks (migration support)
	ts.loadFromFile()

	// Start periodic persistence
	go ts.persistPeriodically()

	log.Printf("[TickStore] Initialized for broker '%s' with max %d ticks per symbol", brokerID, maxTicksPerSymbol)
	return ts
}

// StoreTick stores a new tick for a symbol
func (ts *TickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tick := Tick{
		BrokerID:  ts.brokerID,
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Spread:    spread,
		Timestamp: timestamp,
		LP:        lp,
	}

	// Store in memory buffer
	ts.ticks[symbol] = append(ts.ticks[symbol], tick)

	// Trim if exceeds max
	if len(ts.ticks[symbol]) > ts.maxTicks {
		ts.ticks[symbol] = ts.ticks[symbol][len(ts.ticks[symbol])-ts.maxTicks:]
	}

	// Store in daily file system
	ts.dailyStore.StoreTick(symbol, tick)

	// Update OHLC cache
	ts.ohlcCache.UpdateFromTick(symbol, bid, ask, timestamp)
}

// GetHistory retrieves historical ticks for a symbol
func (ts *TickStore) GetHistory(symbol string, limit int) []Tick {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Try memory first for recent data
	ticks, ok := ts.ticks[symbol]
	if ok && len(ticks) >= limit {
		if limit <= 0 || limit > len(ticks) {
			limit = len(ticks)
		}
		start := len(ticks) - limit
		result := make([]Tick, limit)
		copy(result, ticks[start:])
		return result
	}

	// Fall back to daily store for more historical data
	return ts.dailyStore.GetHistory(symbol, limit, 7) // Look back 7 days
}

// GetOHLC retrieves OHLC bars for a symbol and timeframe
func (ts *TickStore) GetOHLC(symbol string, timeframeSecs int64, limit int) []OHLC {
	// Map timeframe seconds to enum
	var tf Timeframe
	switch timeframeSecs {
	case 60:
		tf = TF_M1
	case 300:
		tf = TF_M5
	case 900:
		tf = TF_M15
	case 3600:
		tf = TF_H1
	case 14400:
		tf = TF_H4
	case 86400:
		tf = TF_D1
	default:
		tf = TF_M1
	}

	return ts.ohlcCache.GetBars(symbol, tf, limit)
}

// GetSymbols returns all symbols with stored ticks
func (ts *TickStore) GetSymbols() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	symbolMap := make(map[string]bool)

	// From memory
	for symbol := range ts.ticks {
		symbolMap[symbol] = true
	}

	// From daily store
	for _, sym := range ts.dailyStore.GetSymbols() {
		symbolMap[sym] = true
	}

	symbols := make([]string, 0, len(symbolMap))
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// GetTickCount returns the number of ticks stored for a symbol
func (ts *TickStore) GetTickCount(symbol string) int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.ticks[symbol])
}

// persistPeriodically saves ticks to file every 30 seconds
func (ts *TickStore) persistPeriodically() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		ts.saveToFile()
	}
}

// saveToFile persists current session ticks to JSON file (legacy compat)
func (ts *TickStore) saveToFile() {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	data, err := json.Marshal(ts.ticks)
	if err != nil {
		log.Printf("[TickStore] Error marshaling ticks: %v", err)
		return
	}

	if err := os.WriteFile(ts.filePath, data, 0644); err != nil {
		log.Printf("[TickStore] Error saving ticks: %v", err)
		return
	}

	// Count total ticks
	total := 0
	for _, ticks := range ts.ticks {
		total += len(ticks)
	}
	log.Printf("[TickStore] Persisted %d ticks across %d symbols", total, len(ts.ticks))
}

// loadFromFile loads ticks from JSON file
func (ts *TickStore) loadFromFile() {
	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[TickStore] Error reading ticks file: %v", err)
		}
		return
	}

	if err := json.Unmarshal(data, &ts.ticks); err != nil {
		log.Printf("[TickStore] Error parsing ticks file: %v", err)
		return
	}

	total := 0
	for _, ticks := range ts.ticks {
		total += len(ticks)
	}
	log.Printf("[TickStore] Loaded %d ticks across %d symbols from file", total, len(ts.ticks))
}

// GetDailyStore returns the daily store for direct access
func (ts *TickStore) GetDailyStore() *DailyStore {
	return ts.dailyStore
}

// GetOHLCCache returns the OHLC cache for direct access
func (ts *TickStore) GetOHLCCache() *OHLCCache {
	return ts.ohlcCache
}

// RebuildOHLCForSymbol rebuilds OHLC cache from tick history
func (ts *TickStore) RebuildOHLCForSymbol(symbol string) {
	ticks := ts.dailyStore.GetHistory(symbol, 0, 30) // All ticks for 30 days
	if len(ticks) > 0 {
		ts.ohlcCache.RebuildFromTicks(symbol, ticks)
	}
}
