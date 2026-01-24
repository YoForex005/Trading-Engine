package tickstore

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// StorageBackend defines the storage backend type
type StorageBackend string

const (
	BackendJSON   StorageBackend = "json"
	BackendSQLite StorageBackend = "sqlite"
	BackendDual   StorageBackend = "dual" // Both JSON + SQLite (migration mode)
)

// OptimizedTickStore provides high-performance tick storage
// with bounded memory, async writes, and quote throttling.
// This replaces the original TickStore to prevent crashes.
type OptimizedTickStore struct {
	mu         sync.RWMutex
	brokerID   string
	maxTicks   int

	// Ring buffers for bounded memory (per-symbol)
	rings      map[string]*TickRingBuffer

	// Last prices for throttling (skip unchanged quotes)
	lastPrices map[string]float64
	throttleMu sync.RWMutex

	// Async batch writer (for JSON backend)
	writeQueue chan *Tick
	writeBatch []Tick
	batchMu    sync.Mutex
	batchSize  int

	// Storage backend
	backend       StorageBackend
	useJSONLegacy bool
	sqliteStore   *SQLiteStore

	// OHLC cache
	ohlcCache  *OHLCCache

	// Stats
	ticksReceived   int64
	ticksThrottled  int64
	ticksStored     int64

	// Control
	stopChan   chan struct{}
}

// TickRingBuffer is a fixed-size circular buffer for ticks
type TickRingBuffer struct {
	buffer []Tick
	head   int
	tail   int
	size   int
	count  int
	mu     sync.RWMutex
}

// NewTickRingBuffer creates a new ring buffer
func NewTickRingBuffer(size int) *TickRingBuffer {
	return &TickRingBuffer{
		buffer: make([]Tick, size),
		size:   size,
	}
}

// Push adds a tick to the ring buffer (O(1), no allocation)
func (rb *TickRingBuffer) Push(tick Tick) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buffer[rb.tail] = tick
	rb.tail = (rb.tail + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	} else {
		// Buffer full, overwrite oldest
		rb.head = (rb.head + 1) % rb.size
	}
}

// GetRecent returns the N most recent ticks
func (rb *TickRingBuffer) GetRecent(n int) []Tick {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if n > rb.count {
		n = rb.count
	}
	if n <= 0 {
		return nil
	}

	result := make([]Tick, n)
	start := (rb.tail - n + rb.size) % rb.size

	for i := 0; i < n; i++ {
		idx := (start + i) % rb.size
		result[i] = rb.buffer[idx]
	}

	return result
}

// Count returns the number of ticks in the buffer
func (rb *TickRingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// TickStoreConfig holds configuration for OptimizedTickStore
type TickStoreConfig struct {
	BrokerID         string
	MaxTicksPerSymbol int
	Backend          StorageBackend
	SQLiteBasePath   string // Path for SQLite databases
	EnableJSONLegacy bool
}

// NewOptimizedTickStore creates a high-performance tick store
func NewOptimizedTickStore(brokerID string, maxTicksPerSymbol int) *OptimizedTickStore {
	return NewOptimizedTickStoreWithConfig(TickStoreConfig{
		BrokerID:         brokerID,
		MaxTicksPerSymbol: maxTicksPerSymbol,
		Backend:          BackendJSON, // Default to JSON for backwards compatibility
		EnableJSONLegacy: true,
	})
}

// NewOptimizedTickStoreWithConfig creates a tick store with custom configuration
func NewOptimizedTickStoreWithConfig(cfg TickStoreConfig) *OptimizedTickStore {
	// Set defaults
	if cfg.MaxTicksPerSymbol == 0 {
		cfg.MaxTicksPerSymbol = 10000
	}
	if cfg.Backend == "" {
		cfg.Backend = BackendJSON
	}

	ts := &OptimizedTickStore{
		brokerID:      cfg.BrokerID,
		maxTicks:      cfg.MaxTicksPerSymbol,
		rings:         make(map[string]*TickRingBuffer),
		lastPrices:    make(map[string]float64),
		backend:       cfg.Backend,
		useJSONLegacy: cfg.EnableJSONLegacy,
		writeQueue:    make(chan *Tick, 10000), // Buffered async queue
		writeBatch:    make([]Tick, 0, 1000),
		batchSize:     500, // Flush every 500 ticks
		ohlcCache:     NewOHLCCache([]Timeframe{TF_M1, TF_M5, TF_M15, TF_H1, TF_H4, TF_D1}),
		stopChan:      make(chan struct{}),
	}

	// Initialize SQLite store if needed
	if cfg.Backend == BackendSQLite || cfg.Backend == BackendDual {
		basePath := cfg.SQLiteBasePath
		if basePath == "" {
			basePath = "data/ticks/db" // Default path
		}
		sqliteStore, err := NewSQLiteStore(basePath)
		if err != nil {
			log.Printf("[OptimizedTickStore] WARN: Failed to initialize SQLite: %v", err)
			log.Printf("[OptimizedTickStore] Falling back to JSON-only storage")
		} else {
			ts.sqliteStore = sqliteStore
			log.Printf("[OptimizedTickStore] SQLite storage initialized at %s", basePath)
		}
	}

	// Start async batch writer for JSON
	if ts.useJSONLegacy {
		go ts.asyncBatchWriter()
		go ts.periodicFlush()
	}

	log.Printf("[OptimizedTickStore] Initialized with backend=%s, ring buffers (max %d ticks/symbol)",
		cfg.Backend, cfg.MaxTicksPerSymbol)
	return ts
}

// StoreTick stores a tick with throttling and async persistence
// This is the HOT PATH - optimized for minimal latency
func (ts *OptimizedTickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) {
	atomic.AddInt64(&ts.ticksReceived, 1)

	// THROTTLING: Skip if price hasn't changed significantly (< 0.00001 = 0.001%)
	ts.throttleMu.RLock()
	lastPrice, exists := ts.lastPrices[symbol]
	ts.throttleMu.RUnlock()

	if exists {
		priceChange := (bid - lastPrice) / lastPrice
		if priceChange < 0 {
			priceChange = -priceChange
		}
		if priceChange < 0.00001 { // Skip if < 0.001% change
			atomic.AddInt64(&ts.ticksThrottled, 1)
			return
		}
	}

	// Update last price
	ts.throttleMu.Lock()
	ts.lastPrices[symbol] = bid
	ts.throttleMu.Unlock()

	tick := Tick{
		BrokerID:  ts.brokerID,
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Spread:    spread,
		Timestamp: timestamp,
		LP:        lp,
	}

	// Store in ring buffer (bounded memory, O(1))
	ts.mu.Lock()
	ring, ok := ts.rings[symbol]
	if !ok {
		ring = NewTickRingBuffer(ts.maxTicks)
		ts.rings[symbol] = ring
	}
	ts.mu.Unlock()

	ring.Push(tick)
	atomic.AddInt64(&ts.ticksStored, 1)

	// Update OHLC cache (fast, in-memory)
	ts.ohlcCache.UpdateFromTick(symbol, bid, ask, timestamp)

	// Persist to storage backend (non-blocking)
	ts.persistTick(&tick)
}

// persistTick persists a tick to the configured storage backend
func (ts *OptimizedTickStore) persistTick(tick *Tick) {
	// SQLite storage (preferred)
	if ts.sqliteStore != nil {
		// Create a copy with Unix milliseconds timestamp for SQLite
		sqliteTick := Tick{
			BrokerID:  tick.BrokerID,
			Symbol:    tick.Symbol,
			Bid:       tick.Bid,
			Ask:       tick.Ask,
			Spread:    tick.Spread,
			Timestamp: tick.Timestamp, // Keep as time.Time
			LP:        tick.LP,
		}
		if err := ts.sqliteStore.StoreTick(sqliteTick); err != nil {
			// Log error but don't block (queue full is expected under high load)
			if err.Error() != "write queue full" {
				log.Printf("[OptimizedTickStore] SQLite write error: %v", err)
			}
		}
	}

	// JSON storage (if enabled)
	if ts.useJSONLegacy {
		ts.queueJSONWrite(tick)
	}
}

// queueJSONWrite queues a tick for JSON persistence (legacy)
func (ts *OptimizedTickStore) queueJSONWrite(tick *Tick) {
	select {
	case ts.writeQueue <- tick:
		// Queued successfully
	default:
		// Queue full - skip persistence to prevent blocking
		// Data is still in ring buffer for queries
	}
}

// asyncBatchWriter batches ticks and writes them to disk asynchronously
func (ts *OptimizedTickStore) asyncBatchWriter() {
	for {
		select {
		case <-ts.stopChan:
			ts.flushBatch() // Final flush
			return
		case tick := <-ts.writeQueue:
			ts.batchMu.Lock()
			ts.writeBatch = append(ts.writeBatch, *tick)
			shouldFlush := len(ts.writeBatch) >= ts.batchSize
			ts.batchMu.Unlock()

			if shouldFlush {
				ts.flushBatch()
			}
		}
	}
}

// periodicFlush flushes remaining ticks every 30 seconds
func (ts *OptimizedTickStore) periodicFlush() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ts.stopChan:
			return
		case <-ticker.C:
			ts.flushBatch()
			ts.logStats()
		}
	}
}

// flushBatch writes accumulated ticks to disk
func (ts *OptimizedTickStore) flushBatch() {
	ts.batchMu.Lock()
	if len(ts.writeBatch) == 0 {
		ts.batchMu.Unlock()
		return
	}

	// Take ownership of batch
	batch := ts.writeBatch
	ts.writeBatch = make([]Tick, 0, ts.batchSize)
	ts.batchMu.Unlock()

	// Group by symbol and date
	bySymbolDate := make(map[string][]Tick)
	today := time.Now().Format("2006-01-02")

	for _, tick := range batch {
		key := tick.Symbol + ":" + today
		bySymbolDate[key] = append(bySymbolDate[key], tick)
	}

	// Write each file (append mode)
	basePath := "data/ticks"
	os.MkdirAll(basePath, 0755)

	for key, ticks := range bySymbolDate {
		parts := splitKey(key)
		symbol, date := parts[0], parts[1]

		symbolDir := filepath.Join(basePath, symbol)
		os.MkdirAll(symbolDir, 0755)

		filePath := filepath.Join(symbolDir, date+".json")

		// Read existing, append, write
		var existing []Tick
		if data, err := os.ReadFile(filePath); err == nil {
			json.Unmarshal(data, &existing)
		}

		existing = append(existing, ticks...)

		// Limit file size (keep last 50k ticks per file)
		if len(existing) > 50000 {
			existing = existing[len(existing)-50000:]
		}

		if data, err := json.Marshal(existing); err == nil {
			os.WriteFile(filePath, data, 0644)
		}
	}

	log.Printf("[OptimizedTickStore] Flushed %d ticks to disk", len(batch))
}

func splitKey(key string) []string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == ':' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key, ""}
}

func (ts *OptimizedTickStore) logStats() {
	received := atomic.LoadInt64(&ts.ticksReceived)
	throttled := atomic.LoadInt64(&ts.ticksThrottled)
	stored := atomic.LoadInt64(&ts.ticksStored)

	if received > 0 {
		throttleRate := float64(throttled) / float64(received) * 100
		log.Printf("[OptimizedTickStore] Stats: received=%d, stored=%d, throttled=%d (%.1f%%)",
			received, stored, throttled, throttleRate)
	}
}

// GetHistory returns recent ticks for a symbol
func (ts *OptimizedTickStore) GetHistory(symbol string, limit int) []Tick {
	ts.mu.RLock()
	ring, ok := ts.rings[symbol]
	ts.mu.RUnlock()

	if !ok {
		return nil
	}

	return ring.GetRecent(limit)
}

// GetOHLC returns OHLC bars
func (ts *OptimizedTickStore) GetOHLC(symbol string, timeframeSecs int64, limit int) []OHLC {
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

// GetSymbols returns all symbols
func (ts *OptimizedTickStore) GetSymbols() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	symbols := make([]string, 0, len(ts.rings))
	for symbol := range ts.rings {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// GetTickCount returns tick count for a symbol
func (ts *OptimizedTickStore) GetTickCount(symbol string) int {
	ts.mu.RLock()
	ring, ok := ts.rings[symbol]
	ts.mu.RUnlock()

	if !ok {
		return 0
	}
	return ring.Count()
}

// Stop gracefully stops the store
func (ts *OptimizedTickStore) Stop() {
	log.Printf("[OptimizedTickStore] Stopping...")
	close(ts.stopChan)

	// Stop SQLite store if initialized
	if ts.sqliteStore != nil {
		if err := ts.sqliteStore.Stop(); err != nil {
			log.Printf("[OptimizedTickStore] Error stopping SQLite: %v", err)
		}
	}

	log.Printf("[OptimizedTickStore] Stopped")
}

// Flush forces a flush of pending writes
func (ts *OptimizedTickStore) Flush() error {
	ts.flushBatch()
	return nil
}

// GetStorageStats returns storage backend statistics
func (ts *OptimizedTickStore) GetStorageStats() map[string]interface{} {
	return map[string]interface{}{
		"backend":         string(ts.backend),
		"ticks_received":  ts.ticksReceived,
		"ticks_stored":    ts.ticksStored,
		"ticks_throttled": ts.ticksThrottled,
		"use_json_legacy": ts.useJSONLegacy,
	}
}

// GetOHLCCache returns the OHLC cache
func (ts *OptimizedTickStore) GetOHLCCache() *OHLCCache {
	return ts.ohlcCache
}

// ProductionConfig creates a production-ready configuration with SQLite backend
func ProductionConfig(brokerID string) TickStoreConfig {
	return TickStoreConfig{
		BrokerID:         brokerID,
		MaxTicksPerSymbol: 10000,
		Backend:          BackendSQLite,
		SQLiteBasePath:   "data/ticks/db",
		EnableJSONLegacy: false, // SQLite only for production
	}
}

// DualStorageConfig creates a configuration with both SQLite and JSON (migration mode)
func DualStorageConfig(brokerID string) TickStoreConfig {
	return TickStoreConfig{
		BrokerID:         brokerID,
		MaxTicksPerSymbol: 10000,
		Backend:          BackendDual,
		SQLiteBasePath:   "data/ticks/db",
		EnableJSONLegacy: true, // Keep JSON during migration
	}
}
