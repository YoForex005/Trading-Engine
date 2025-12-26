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

// TickStore provides in-memory tick storage with file persistence
type TickStore struct {
	mu       sync.RWMutex
	ticks    map[string][]Tick // symbol -> ticks
	maxTicks int               // Max ticks per symbol
	filePath string            // Persistence file path
	brokerID string            // This broker's ID
}

// NewTickStore creates a new tick store instance
func NewTickStore(brokerID string, maxTicksPerSymbol int) *TickStore {
	ts := &TickStore{
		ticks:    make(map[string][]Tick),
		maxTicks: maxTicksPerSymbol,
		filePath: "data/ticks.json",
		brokerID: brokerID,
	}

	// Ensure data directory exists
	os.MkdirAll("data", 0755)

	// Load existing ticks from file
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

	// Append tick
	ts.ticks[symbol] = append(ts.ticks[symbol], tick)

	// Trim if exceeds max
	if len(ts.ticks[symbol]) > ts.maxTicks {
		ts.ticks[symbol] = ts.ticks[symbol][len(ts.ticks[symbol])-ts.maxTicks:]
	}
}

// GetHistory retrieves historical ticks for a symbol
func (ts *TickStore) GetHistory(symbol string, limit int) []Tick {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	ticks, ok := ts.ticks[symbol]
	if !ok {
		return []Tick{}
	}

	if limit <= 0 || limit > len(ticks) {
		limit = len(ticks)
	}

	// Return most recent ticks
	start := len(ticks) - limit
	if start < 0 {
		start = 0
	}

	result := make([]Tick, limit)
	copy(result, ticks[start:])
	return result
}

// GetOHLC aggregates ticks into OHLC bars for a given timeframe
func (ts *TickStore) GetOHLC(symbol string, timeframeSecs int64, limit int) []OHLC {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	ticks, ok := ts.ticks[symbol]
	if !ok || len(ticks) == 0 {
		return []OHLC{}
	}

	// Group ticks by candle time
	candles := make(map[int64]*OHLC)
	var candleTimes []int64

	for _, tick := range ticks {
		ts := tick.Timestamp.Unix()
		candleTime := (ts / timeframeSecs) * timeframeSecs
		price := (tick.Bid + tick.Ask) / 2

		if candle, exists := candles[candleTime]; exists {
			if price > candle.High {
				candle.High = price
			}
			if price < candle.Low {
				candle.Low = price
			}
			candle.Close = price
			candle.Volume++
		} else {
			candles[candleTime] = &OHLC{
				Time:   candleTime,
				Open:   price,
				High:   price,
				Low:    price,
				Close:  price,
				Volume: 1,
			}
			candleTimes = append(candleTimes, candleTime)
		}
	}

	// Convert to slice and limit
	result := make([]OHLC, 0, len(candleTimes))
	for _, t := range candleTimes {
		result = append(result, *candles[t])
	}

	if limit > 0 && len(result) > limit {
		result = result[len(result)-limit:]
	}

	return result
}

// GetSymbols returns all symbols with stored ticks
func (ts *TickStore) GetSymbols() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	symbols := make([]string, 0, len(ts.ticks))
	for symbol := range ts.ticks {
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

// saveToFile persists all ticks to JSON file
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
