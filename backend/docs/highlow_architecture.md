# High/Low Tracking System Architecture

## 1. DATA STRUCTURE DESIGN

### Primary Storage: In-Memory Map with Redis Backup
**Rationale**: Real-time performance (O(1) lookups) with persistence for crash recovery

**Go Struct Definition**:
```go
package tickstore

import (
    "sync"
    "time"
)

// SymbolHighLow tracks 24-hour high/low prices for a symbol
type SymbolHighLow struct {
    Symbol       string    `json:"symbol"`
    High24h      float64   `json:"high_24h"`      // Highest bid in 24h window
    Low24h       float64   `json:"low_24h"`       // Lowest bid in 24h window
    HighTime     time.Time `json:"high_time"`     // When high was recorded
    LowTime      time.Time `json:"low_time"`      // When low was recorded
    WindowStart  time.Time `json:"window_start"`  // 24h window start (rolling)
    LastUpdate   time.Time `json:"last_update"`   // Last tick timestamp
    LP           string    `json:"lp"`            // Liquidity provider source
}

// HighLowTracker manages high/low tracking for all symbols
type HighLowTracker struct {
    mu           sync.RWMutex
    data         map[string]*SymbolHighLow  // symbol -> high/low data
    windowHours  int                        // Default: 24 hours
    redisClient  *redis.Client              // Optional: Redis for persistence
    enableRedis  bool                       // Feature flag for Redis backup
}
```

**Memory Efficiency**:
- Per symbol: ~120 bytes (2 floats + 4 timestamps + 2 strings)
- 100 symbols: ~12 KB
- 1000 symbols: ~120 KB
- Negligible memory overhead

---

## 2. FIX INTEGRATION POINTS

### FIX Message Sources for High/Low

**Option A: Direct from FIX MarketDataSnapshot (Tag 269)**
- MDEntryType=7 (Tag 269): Trading Session High Price
- MDEntryType=8 (Tag 269): Trading Session Low Price
- **Location**: `backend/fix/gateway.go` MarketData struct already has High24h/Low24h fields (lines 112-113)
- **Current Implementation**: Fields exist but NOT populated from FIX messages

**Option B: Calculate from Bid/Ask Ticks (Fallback)**
- Track min/max from incoming bid prices
- More reliable as YOFX may not send Tag 269 entries
- **Recommended**: Use this as primary method, FIX high/low as validation

**Implementation in gateway.go**:
```go
// In parseMarketDataSnapshot() or parseMarketDataIncrementalRefresh()
func (g *FIXGateway) parseMarketDataSnapshot(msg quickfix.Message, sessionID string) {
    // ... existing code ...

    // NEW: Extract high/low from MDEntries if available
    var high24h, low24h float64
    noMDEntries, _ := msg.Body.GetInt(tag.NoMDEntries)

    for i := 1; i <= noMDEntries; i++ {
        group := msg.Body.GetGroup(i, tag.NoMDEntries)
        entryType, _ := group.GetString(tag.MDEntryType)
        price, _ := group.GetFloat(tag.MDEntryPx)

        if entryType == "7" { // High
            high24h = price
        } else if entryType == "8" { // Low
            low24h = price
        }
    }

    // If FIX provides high/low, use them; otherwise, calculate
    // ... update HighLowTracker ...
}
```

---

## 3. UPDATE STRATEGY

### Real-Time Tracking on Each Tick
**Where**: `backend/ws/hub.go` BroadcastTick() - line 175 (after StoreTick)

```go
// In Hub.BroadcastTick() - ADD AFTER line 175
if h.tickStore != nil {
    h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())

    // NEW: Update high/low tracker
    if h.highLowTracker != nil {
        h.highLowTracker.UpdateTick(tick.Symbol, tick.Bid, tick.LP, time.Now())
    }
}
```

### Sliding 24-Hour Window Algorithm
**Pseudo-code**:
```
function UpdateTick(symbol, bid, lp, timestamp):
    lock(symbol)

    data = getOrCreate(symbol)

    // Check if we need to reset the window (24h elapsed)
    if timestamp - data.WindowStart > 24 hours:
        // Reset window
        data.High24h = bid
        data.Low24h = bid
        data.HighTime = timestamp
        data.LowTime = timestamp
        data.WindowStart = timestamp
    else:
        // Update high
        if bid > data.High24h:
            data.High24h = bid
            data.HighTime = timestamp

        // Update low
        if bid < data.Low24h:
            data.Low24h = bid
            data.LowTime = timestamp

    data.LastUpdate = timestamp
    data.LP = lp

    // Optional: Async persist to Redis every N updates
    if enableRedis and updateCount % 100 == 0:
        asyncPersistToRedis(symbol, data)

    unlock(symbol)
```

### Handling Market Close/Open Boundaries
- **Option 1: True Rolling 24h** - Continuous, no reset at midnight
- **Option 2: Session-based** - Reset at market open (requires session schedule)
- **Recommendation**: Start with rolling 24h (simpler), add session support later

---

## 4. PERFORMANCE CONSIDERATIONS

### Memory Overhead
- **In-memory map**: O(1) read/write, ~120 bytes per symbol
- **No historical storage**: Only track current 24h window
- **Cleanup**: Remove symbols inactive for >48h (prevent memory leak)

### Fast Lookup
```go
func (t *HighLowTracker) GetHighLow(symbol string) (high, low float64, ok bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()

    data, exists := t.data[symbol]
    if !exists {
        return 0, 0, false
    }

    // Validate window is still active
    if time.Since(data.WindowStart) > 24*time.Hour {
        return 0, 0, false  // Stale data
    }

    return data.High24h, data.Low24h, true
}
```

### Efficient Cleanup (Background Goroutine)
```go
func (t *HighLowTracker) cleanupStale() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        t.mu.Lock()
        now := time.Now()
        for symbol, data := range t.data {
            // Remove if no update in 48h
            if now.Sub(data.LastUpdate) > 48*time.Hour {
                delete(t.data, symbol)
            }
        }
        t.mu.Unlock()
    }
}
```

---

## 5. REDIS SCHEMA (OPTIONAL)

**Use Case**: Crash recovery, cross-instance synchronization

**Key Pattern**: `highlow:{symbol}`

**Data Structure**: Hash
```
HSET highlow:EURUSD high_24h 1.0850
HSET highlow:EURUSD low_24h 1.0820
HSET highlow:EURUSD high_time 1738252800000
HSET highlow:EURUSD low_time 1738249200000
HSET highlow:EURUSD window_start 1738166400000
HSET highlow:EURUSD last_update 1738252800000
HSET highlow:EURUSD lp YOFX1
EXPIRE highlow:EURUSD 172800  // 48 hours TTL
```

**Go Redis Integration**:
```go
import "github.com/go-redis/redis/v8"

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
```

**Restore on Startup**:
```go
func (t *HighLowTracker) loadFromRedis() error {
    if !t.enableRedis || t.redisClient == nil {
        return nil
    }

    ctx := context.Background()
    keys, _ := t.redisClient.Keys(ctx, "highlow:*").Result()

    for _, key := range keys {
        symbol := strings.TrimPrefix(key, "highlow:")
        data := t.redisClient.HGetAll(ctx, key).Val()

        // Parse and populate t.data[symbol]
        // ... conversion logic ...
    }

    return nil
}
```

---

## 6. INTEGRATION CHECKLIST

### Step 1: Add HighLowTracker to TickStore
**File**: `backend/tickstore/highlow_tracker.go` (NEW)
- Implement HighLowTracker struct
- Implement UpdateTick(), GetHighLow(), cleanupStale()
- Optional: Redis persistence methods

### Step 2: Integrate with OptimizedTickStore
**File**: `backend/tickstore/optimized_store.go`
```go
type OptimizedTickStore struct {
    // ... existing fields ...
    highLowTracker *HighLowTracker  // NEW
}

func (ts *OptimizedTickStore) StoreTick(...) {
    // ... existing code ...

    // NEW: Update high/low on each tick
    if ts.highLowTracker != nil {
        ts.highLowTracker.UpdateTick(symbol, bid, lp, timestamp)
    }
}

func (ts *OptimizedTickStore) GetHighLow(symbol string) (float64, float64, bool) {
    if ts.highLowTracker == nil {
        return 0, 0, false
    }
    return ts.highLowTracker.GetHighLow(symbol)
}
```

### Step 3: Update FIX Gateway MarketData Struct
**File**: `backend/fix/gateway.go`
- Ensure High24h/Low24h are populated from FIX messages (Tag 269 type 7/8)
- Add fallback to calculate from bid/ask

### Step 4: Update WebSocket Hub
**File**: `backend/ws/hub.go`
- Add HighLowTracker reference
- Call UpdateTick() in BroadcastTick()

### Step 5: Expose via API
**File**: `backend/api/history.go`
```go
func (h *APIHandler) HandleGetHighLow(w http.ResponseWriter, r *http.Request) {
    symbol := r.URL.Query().Get("symbol")

    high, low, ok := h.tickStore.GetHighLow(symbol)
    if !ok {
        http.Error(w, "Symbol not found or no data", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "symbol": symbol,
        "high_24h": high,
        "low_24h": low,
    })
}
```

### Step 6: Update WebSocket MarketTick Message
**File**: `backend/ws/hub.go`
```go
type MarketTick struct {
    // ... existing fields ...
    High24h float64 `json:"high24h,omitempty"`  // NEW
    Low24h  float64 `json:"low24h,omitempty"`   // NEW
}

// In BroadcastTick(), populate high/low before marshaling
if h.highLowTracker != nil {
    high, low, ok := h.highLowTracker.GetHighLow(tick.Symbol)
    if ok {
        tick.High24h = high
        tick.Low24h = low
    }
}
```

---

## 7. TESTING STRATEGY

### Unit Tests
- Test UpdateTick() with increasing/decreasing prices
- Test 24h window reset
- Test concurrent updates (race conditions)
- Test Redis persistence/recovery

### Integration Tests
- Test with live FIX feed
- Verify high/low matches across restarts
- Verify WebSocket clients receive high/low

---

## SUMMARY

**Storage**: In-memory map (120 bytes/symbol) + optional Redis backup
**Update**: Real-time on each tick in BroadcastTick()
**Window**: Rolling 24h, reset when 24h elapsed since WindowStart
**Lookup**: O(1) GetHighLow(symbol)
**Persistence**: Optional Redis with 48h TTL
**Cleanup**: Hourly background job removes stale (>48h) symbols
**API**: New endpoint /api/highlow?symbol=EURUSD
**WebSocket**: Add high24h/low24h to MarketTick JSON message
