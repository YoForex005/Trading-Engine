# High/Low Tracker - Implementation Guide

## Quick Reference

**Primary File**: `backend/tickstore/highlow_tracker.go` (NEW - to be created)
**Integration Files**:
- `backend/tickstore/optimized_store.go` - Add tracker field and calls
- `backend/ws/hub.go` - Update MarketTick struct, populate high/low
- `backend/fix/gateway.go` - Parse FIX Tag 269 (optional)
- `backend/api/history.go` - Add GET /api/highlow endpoint

---

## Implementation Steps

### STEP 1: Create HighLowTracker (NEW FILE)

**File**: `backend/tickstore/highlow_tracker.go`

```go
package tickstore

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"
)

// SymbolHighLow tracks 24-hour high/low prices for a symbol
type SymbolHighLow struct {
    Symbol       string    `json:"symbol"`
    High24h      float64   `json:"high_24h"`
    Low24h       float64   `json:"low_24h"`
    HighTime     time.Time `json:"high_time"`
    LowTime      time.Time `json:"low_time"`
    WindowStart  time.Time `json:"window_start"`
    LastUpdate   time.Time `json:"last_update"`
    LP           string    `json:"lp"`
}

// HighLowTracker manages high/low tracking for all symbols
type HighLowTracker struct {
    mu           sync.RWMutex
    data         map[string]*SymbolHighLow
    windowHours  int
    redisClient  *redis.Client
    enableRedis  bool
    updateCount  map[string]int // For async Redis persistence
}

// NewHighLowTracker creates a new tracker
func NewHighLowTracker(windowHours int, enableRedis bool, redisClient *redis.Client) *HighLowTracker {
    t := &HighLowTracker{
        data:         make(map[string]*SymbolHighLow),
        windowHours:  windowHours,
        enableRedis:  enableRedis,
        redisClient:  redisClient,
        updateCount:  make(map[string]int),
    }

    // Load from Redis if enabled
    if enableRedis && redisClient != nil {
        t.loadFromRedis()
    }

    // Start cleanup goroutine
    go t.cleanupStale()

    return t
}

// UpdateTick updates high/low for a symbol based on new tick
func (t *HighLowTracker) UpdateTick(symbol string, bid float64, lp string, timestamp time.Time) {
    t.mu.Lock()
    defer t.mu.Unlock()

    data, exists := t.data[symbol]
    if !exists {
        // Initialize new symbol
        data = &SymbolHighLow{
            Symbol:      symbol,
            High24h:     bid,
            Low24h:      bid,
            HighTime:    timestamp,
            LowTime:     timestamp,
            WindowStart: timestamp,
            LastUpdate:  timestamp,
            LP:          lp,
        }
        t.data[symbol] = data
        return
    }

    // Check if window needs reset (24h elapsed)
    windowDuration := time.Duration(t.windowHours) * time.Hour
    if timestamp.Sub(data.WindowStart) > windowDuration {
        // Reset window
        data.High24h = bid
        data.Low24h = bid
        data.HighTime = timestamp
        data.LowTime = timestamp
        data.WindowStart = timestamp
    } else {
        // Update high
        if bid > data.High24h {
            data.High24h = bid
            data.HighTime = timestamp
        }

        // Update low
        if bid < data.Low24h {
            data.Low24h = bid
            data.LowTime = timestamp
        }
    }

    data.LastUpdate = timestamp
    data.LP = lp

    // Async persist to Redis every 100 updates
    if t.enableRedis && t.redisClient != nil {
        t.updateCount[symbol]++
        if t.updateCount[symbol]%100 == 0 {
            go t.persistToRedis(symbol, data)
        }
    }
}

// GetHighLow returns current high/low for a symbol
func (t *HighLowTracker) GetHighLow(symbol string) (high, low float64, ok bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()

    data, exists := t.data[symbol]
    if !exists {
        return 0, 0, false
    }

    // Validate window is still active
    windowDuration := time.Duration(t.windowHours) * time.Hour
    if time.Since(data.WindowStart) > windowDuration {
        return 0, 0, false // Stale data
    }

    return data.High24h, data.Low24h, true
}

// GetHighLowData returns full high/low data for a symbol
func (t *HighLowTracker) GetHighLowData(symbol string) (*SymbolHighLow, bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()

    data, exists := t.data[symbol]
    if !exists {
        return nil, false
    }

    // Return a copy to avoid race conditions
    dataCopy := *data
    return &dataCopy, true
}

// cleanupStale removes symbols inactive for >48h
func (t *HighLowTracker) cleanupStale() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        t.mu.Lock()
        now := time.Now()
        removed := 0

        for symbol, data := range t.data {
            // Remove if no update in 48h
            if now.Sub(data.LastUpdate) > 48*time.Hour {
                delete(t.data, symbol)
                delete(t.updateCount, symbol)
                removed++
            }
        }

        t.mu.Unlock()

        if removed > 0 {
            // Log cleanup activity
            fmt.Printf("[HighLowTracker] Cleaned %d stale symbols\n", removed)
        }
    }
}

// persistToRedis saves data to Redis (async)
func (t *HighLowTracker) persistToRedis(symbol string, data *SymbolHighLow) error {
    if !t.enableRedis || t.redisClient == nil {
        return nil
    }

    ctx := context.Background()
    key := fmt.Sprintf("highlow:%s", symbol)

    pipe := t.redisClient.Pipeline()
    pipe.HSet(ctx, key, "high_24h", data.High24h)
    pipe.HSet(ctx, key, "low_24h", data.Low24h)
    pipe.HSet(ctx, key, "high_time", data.HighTime.UnixMilli())
    pipe.HSet(ctx, key, "low_time", data.LowTime.UnixMilli())
    pipe.HSet(ctx, key, "window_start", data.WindowStart.UnixMilli())
    pipe.HSet(ctx, key, "last_update", data.LastUpdate.UnixMilli())
    pipe.HSet(ctx, key, "lp", data.LP)
    pipe.Expire(ctx, key, 48*time.Hour)

    _, err := pipe.Exec(ctx)
    return err
}

// loadFromRedis restores data from Redis on startup
func (t *HighLowTracker) loadFromRedis() error {
    if !t.enableRedis || t.redisClient == nil {
        return nil
    }

    ctx := context.Background()
    keys, err := t.redisClient.Keys(ctx, "highlow:*").Result()
    if err != nil {
        return err
    }

    loaded := 0
    for _, key := range keys {
        symbol := key[8:] // Remove "highlow:" prefix

        data, err := t.redisClient.HGetAll(ctx, key).Result()
        if err != nil {
            continue
        }

        // Parse float64 and int64 values
        high24h, _ := parseFloat64(data["high_24h"])
        low24h, _ := parseFloat64(data["low_24h"])
        highTime := parseTimeFromMillis(data["high_time"])
        lowTime := parseTimeFromMillis(data["low_time"])
        windowStart := parseTimeFromMillis(data["window_start"])
        lastUpdate := parseTimeFromMillis(data["last_update"])
        lp := data["lp"]

        t.data[symbol] = &SymbolHighLow{
            Symbol:      symbol,
            High24h:     high24h,
            Low24h:      low24h,
            HighTime:    highTime,
            LowTime:     lowTime,
            WindowStart: windowStart,
            LastUpdate:  lastUpdate,
            LP:          lp,
        }
        loaded++
    }

    if loaded > 0 {
        fmt.Printf("[HighLowTracker] Loaded %d symbols from Redis\n", loaded)
    }

    return nil
}

// Helper functions
func parseFloat64(s string) (float64, error) {
    var f float64
    _, err := fmt.Sscanf(s, "%f", &f)
    return f, err
}

func parseTimeFromMillis(s string) time.Time {
    var ms int64
    fmt.Sscanf(s, "%d", &ms)
    return time.UnixMilli(ms)
}
```

---

### STEP 2: Integrate with OptimizedTickStore

**File**: `backend/tickstore/optimized_store.go`

**Add field to struct** (around line 25):
```go
type OptimizedTickStore struct {
    // ... existing fields ...
    highLowTracker *HighLowTracker  // NEW
}
```

**Update NewOptimizedTickStoreWithConfig** (around line 144):
```go
func NewOptimizedTickStoreWithConfig(cfg TickStoreConfig) *OptimizedTickStore {
    // ... existing code ...

    ts := &OptimizedTickStore{
        // ... existing fields ...
        highLowTracker: NewHighLowTracker(24, false, nil), // NEW: 24h window, Redis disabled by default
    }

    // ... rest of function ...
}
```

**Update StoreTick** (around line 196):
```go
func (ts *OptimizedTickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) {
    // ... existing code ...

    // NEW: Update high/low tracker
    if ts.highLowTracker != nil {
        ts.highLowTracker.UpdateTick(symbol, bid, lp, timestamp)
    }
}
```

**Add new methods** (at end of file):
```go
// GetHighLow returns 24h high/low for a symbol
func (ts *OptimizedTickStore) GetHighLow(symbol string) (high, low float64, ok bool) {
    if ts.highLowTracker == nil {
        return 0, 0, false
    }
    return ts.highLowTracker.GetHighLow(symbol)
}

// GetHighLowData returns full high/low data for a symbol
func (ts *OptimizedTickStore) GetHighLowData(symbol string) (*SymbolHighLow, bool) {
    if ts.highLowTracker == nil {
        return nil, false
    }
    return ts.highLowTracker.GetHighLowData(symbol)
}
```

---

### STEP 3: Update WebSocket Hub

**File**: `backend/ws/hub.go`

**Update MarketTick struct** (around line 72):
```go
type MarketTick struct {
    Type        string  `json:"type"`
    Symbol      string  `json:"symbol"`
    Bid         float64 `json:"bid"`
    Ask         float64 `json:"ask"`
    Spread      float64 `json:"spread"`
    Timestamp   int64   `json:"timestamp"`
    LP          string  `json:"lp"`
    DailyChange float64 `json:"dailyChange"`
    High24h     float64 `json:"high24h,omitempty"`  // NEW
    Low24h      float64 `json:"low24h,omitempty"`   // NEW
}
```

**Update BroadcastTick** (around line 225, before json.Marshal):
```go
func (h *Hub) BroadcastTick(tick *MarketTick) {
    // ... existing code up to line 227 (h.lastBroadcast[tick.Symbol] = tick.Bid) ...

    // NEW: Populate high/low from tick store
    if h.tickStore != nil {
        if ts, ok := h.tickStore.(interface{ GetHighLow(string) (float64, float64, bool) }); ok {
            high, low, ok := ts.GetHighLow(tick.Symbol)
            if ok {
                tick.High24h = high
                tick.Low24h = low
            }
        }
    }

    data, err := json.Marshal(tick)
    // ... rest of function ...
}
```

---

### STEP 4: Add API Endpoint

**File**: `backend/api/history.go`

**Add new handler** (at end of file):
```go
// HandleGetHighLow returns 24h high/low for a symbol
func (h *APIHandler) HandleGetHighLow(w http.ResponseWriter, r *http.Request) {
    cors(w)
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    symbol := r.URL.Query().Get("symbol")
    if symbol == "" {
        http.Error(w, "Missing symbol parameter", http.StatusBadRequest)
        return
    }

    // Try to get high/low from tick store
    if ts, ok := h.tickStore.(interface{ GetHighLowData(string) (*tickstore.SymbolHighLow, bool) }); ok {
        data, ok := ts.GetHighLowData(symbol)
        if !ok {
            http.Error(w, "Symbol not found or no data", http.StatusNotFound)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(data)
        return
    }

    http.Error(w, "High/low tracking not available", http.StatusServiceUnavailable)
}
```

**Register route** (in server.go or wherever routes are defined):
```go
http.HandleFunc("/api/highlow", apiHandler.HandleGetHighLow)
```

---

### STEP 5: FIX Gateway Integration (OPTIONAL)

**File**: `backend/fix/gateway.go`

**Update parseMarketDataSnapshot** (find the function and add):
```go
func (g *FIXGateway) parseMarketDataSnapshot(msg quickfix.Message, sessionID string) {
    // ... existing code to extract symbol, bid, ask ...

    // NEW: Extract high/low from MDEntries if available
    var high24h, low24h float64
    var hasHigh, hasLow bool

    noMDEntries, err := msg.Body.GetInt(tag.NoMDEntries)
    if err == nil {
        for i := 1; i <= noMDEntries; i++ {
            group := msg.Body.GetGroup(i, tag.NoMDEntries)
            entryType, _ := group.GetString(tag.MDEntryType)
            price, priceErr := group.GetFloat(tag.MDEntryPx)

            if priceErr == nil {
                if entryType == "7" { // High
                    high24h = price
                    hasHigh = true
                } else if entryType == "8" { // Low
                    low24h = price
                    hasLow = true
                }
            }
        }
    }

    // If FIX provides high/low, use them
    if hasHigh && hasLow {
        data.High24h = high24h
        data.Low24h = low24h
    }
    // Otherwise, calculate from bid/ask (already handled by HighLowTracker)

    // ... rest of function ...
}
```

---

## Testing

### Unit Test: `backend/tickstore/highlow_tracker_test.go`

```go
package tickstore

import (
    "testing"
    "time"
)

func TestHighLowTracker_UpdateTick(t *testing.T) {
    tracker := NewHighLowTracker(24, false, nil)

    now := time.Now()
    tracker.UpdateTick("EURUSD", 1.0850, "YOFX1", now)

    high, low, ok := tracker.GetHighLow("EURUSD")
    if !ok {
        t.Fatal("Expected data for EURUSD")
    }

    if high != 1.0850 || low != 1.0850 {
        t.Errorf("Expected high=low=1.0850, got high=%f low=%f", high, low)
    }

    // Update with higher price
    tracker.UpdateTick("EURUSD", 1.0860, "YOFX1", now.Add(1*time.Minute))
    high, low, ok = tracker.GetHighLow("EURUSD")
    if high != 1.0860 || low != 1.0850 {
        t.Errorf("Expected high=1.0860 low=1.0850, got high=%f low=%f", high, low)
    }

    // Update with lower price
    tracker.UpdateTick("EURUSD", 1.0840, "YOFX1", now.Add(2*time.Minute))
    high, low, ok = tracker.GetHighLow("EURUSD")
    if high != 1.0860 || low != 1.0840 {
        t.Errorf("Expected high=1.0860 low=1.0840, got high=%f low=%f", high, low)
    }
}

func TestHighLowTracker_WindowReset(t *testing.T) {
    tracker := NewHighLowTracker(24, false, nil)

    now := time.Now()
    tracker.UpdateTick("EURUSD", 1.0850, "YOFX1", now)

    // Simulate 25 hours later (beyond 24h window)
    future := now.Add(25 * time.Hour)
    tracker.UpdateTick("EURUSD", 1.0900, "YOFX1", future)

    high, low, ok := tracker.GetHighLow("EURUSD")
    if !ok {
        t.Fatal("Expected data for EURUSD")
    }

    // Window should have reset
    if high != 1.0900 || low != 1.0900 {
        t.Errorf("Expected window reset to 1.0900, got high=%f low=%f", high, low)
    }
}
```

---

## API Usage Examples

### Get High/Low via HTTP
```bash
curl "http://localhost:8080/api/highlow?symbol=EURUSD"
```

Response:
```json
{
  "symbol": "EURUSD",
  "high_24h": 1.0860,
  "low_24h": 1.0820,
  "high_time": "2026-01-30T14:30:00Z",
  "low_time": "2026-01-30T08:15:00Z",
  "window_start": "2026-01-29T15:00:00Z",
  "last_update": "2026-01-30T15:00:00Z",
  "lp": "YOFX1"
}
```

### WebSocket Message with High/Low
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.0845,
  "ask": 1.0847,
  "spread": 0.0002,
  "timestamp": 1738252800000,
  "lp": "YOFX1",
  "dailyChange": 0.15,
  "high24h": 1.0860,
  "low24h": 1.0820
}
```

---

## Performance Characteristics

- **Memory**: 120 bytes per symbol (100 symbols = 12 KB)
- **Update latency**: O(1) - single map lookup + field update
- **Lookup latency**: O(1) - single map lookup with RLock
- **Cleanup overhead**: Hourly scan removes stale symbols
- **Redis persistence**: Async every 100 updates (no blocking)

---

## Summary

This implementation provides:
1. Real-time high/low tracking with minimal overhead
2. Rolling 24-hour window with automatic reset
3. Optional Redis persistence for crash recovery
4. WebSocket broadcast integration
5. HTTP API endpoint
6. FIX message parsing support (optional)
7. Thread-safe concurrent access
8. Automatic cleanup of stale data
