# WebSocket Quote Flow - Exact Code Locations

## Backend Flow Entry Points

### 1. LP Manager Quote Source
**File:** `backend/lpmanager/manager.go`
- Function: `StartQuoteAggregation()`
- Output: Channel returned by `GetQuotesChan()`
- Type: `Quote` struct with `Symbol`, `Bid`, `Ask`, `Timestamp`, `LP`

### 2. Main Goroutine - Quote Piping
**File:** `backend/cmd/server/main.go`
**Lines:** 334-353

```go
// Pipe quotes from LP Manager to Hub
go func() {
    var quoteCount int64 = 0
    for quote := range lpMgr.GetQuotesChan() {
        quoteCount++
        if quoteCount%1000 == 1 {
            log.Printf("[Main] Piping quote #%d to Hub: %s @ %.5f", quoteCount, quote.Symbol, quote.Bid)
        }
        tick := &ws.MarketTick{
            Type:      "tick",
            Symbol:    quote.Symbol,
            Bid:       quote.Bid,
            Ask:       quote.Ask,
            Spread:    quote.Ask - quote.Bid,  // LINE 346 - SPREAD CALCULATION
            Timestamp: quote.Timestamp,
            LP:        quote.LP,
        }
        hub.BroadcastTick(tick)
    }
    log.Println("[Main] Quote pipe closed!")
}()
```

**Key Line:** 346 - `Spread: quote.Ask - quote.Bid`

### 3. Hub Structure Definition
**File:** `backend/ws/hub.go`
**Lines:** 39-61

```go
type Hub struct {
    clients     map[*Client]bool
    broadcast   chan []byte
    register    chan *Client
    unregister  chan *Client
    tickStore   TickStorer
    bbookEngine *core.Engine
    authService *auth.Service

    mu              sync.RWMutex
    latestPrices    map[string]*MarketTick  // Cache of latest prices
    disabledSymbols map[string]bool         // FILTERING MAP

    // Throttling: Track last broadcast price per symbol to reduce CPU load
    lastBroadcast map[string]float64
    throttleMu    sync.RWMutex

    // Stats for monitoring
    ticksReceived  int64
    ticksThrottled int64
    ticksBroadcast int64
}
```

### 4. BroadcastTick Main Logic
**File:** `backend/ws/hub.go`
**Lines:** 139-213

```go
// BroadcastTick broadcasts a market tick to all clients with THROTTLING
// Throttling reduces CPU load by 60-80% by skipping tiny price changes
func (h *Hub) BroadcastTick(tick *MarketTick) {
    atomic.AddInt64(&h.ticksReceived, 1)

    // Update latest price (always - needed for queries)
    h.mu.Lock()
    h.latestPrices[tick.Symbol] = tick  // LINE 146 - ALWAYS UPDATE

    // Skip broadcast if symbol is disabled
    if h.disabledSymbols[tick.Symbol] {  // LINE 149 - DISABLED CHECK
        h.mu.Unlock()
        return
    }
    h.mu.Unlock()

    // ============================================
    // THROTTLING: Skip broadcast if price change < 0.0001% (1/100th of a pip)
    // This reduces broadcast volume by 60-80% and prevents CPU overload
    // ============================================
    h.throttleMu.RLock()
    lastPrice, exists := h.lastBroadcast[tick.Symbol]
    h.throttleMu.RUnlock()

    if exists && lastPrice > 0 {
        priceChange := (tick.Bid - lastPrice) / lastPrice
        if priceChange < 0 {
            priceChange = -priceChange
        }

        // Skip if change < 0.0001% - still update engine & store but skip broadcast
        if priceChange < 0.000001 {  // LINE 170 - THROTTLE THRESHOLD
            atomic.AddInt64(&h.ticksThrottled, 1)

            // Still update B-Book engine (needs all prices for accurate execution)
            if h.bbookEngine != nil {
                h.bbookEngine.UpdatePrice(tick.Symbol, tick.Bid, tick.Ask)
            }

            // Still store tick (for complete chart history)
            if h.tickStore != nil {
                h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
            }
            return  // NO BROADCAST SENT
        }
    }

    // Update last broadcast price
    h.throttleMu.Lock()
    h.lastBroadcast[tick.Symbol] = tick.Bid
    h.throttleMu.Unlock()

    // Notify B-Book engine of new price (for order execution)
    if h.bbookEngine != nil {
        h.bbookEngine.UpdatePrice(tick.Symbol, tick.Bid, tick.Ask)
    }

    // Persist tick for chart history
    if h.tickStore != nil {
        h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
    }

    data, err := json.Marshal(tick)
    if err != nil {
        return
    }

    // NON-BLOCKING SEND: If buffer full, drop tick to keep engine running
    select {
    case h.broadcast <- data:  // LINE 208 - SEND TO CHANNEL
        atomic.AddInt64(&h.ticksBroadcast, 1)
    default:
        // Buffer full - drop to prevent blocking (data still stored for history)
    }
}
```

**Critical Lines:**
- 146: Update `latestPrices` (always)
- 149: Check `disabledSymbols` (early return if disabled)
- 170: Throttle threshold check
- 208: Non-blocking send

### 5. DisabledSymbols Management
**File:** `backend/ws/hub.go`
**Lines:** 112-124

```go
// UpdateDisabledSymbols updates the local filter list
func (h *Hub) UpdateDisabledSymbols(disabled map[string]bool) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.disabledSymbols = disabled
}

// ToggleSymbol updates a single symbol's status
func (h *Hub) ToggleSymbol(symbol string, disabled bool) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.disabledSymbols[symbol] = disabled
}
```

### 6. Hub Run Loop - Broadcasting
**File:** `backend/ws/hub.go`
**Lines:** 238-293

```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            clientCount := len(h.clients)
            h.mu.Unlock()
            log.Printf("[Hub] Client connected. Total clients: %d", clientCount)

            // Send latest prices for all symbols upon connection
            h.mu.RLock()
            for _, tick := range h.latestPrices {
                if !h.disabledSymbols[tick.Symbol] {  // LINE 251 - FILTER DISABLED
                    if data, err := json.Marshal(tick); err == nil {
                        // Try non-blocking send to client on init
                        select {
                        case client.send <- data:
                        default:
                        }
                    }
                }
            }
            h.mu.RUnlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                log.Printf("[Hub] Client disconnected. Total clients: %d", len(h.clients))
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            clientCount := len(h.clients)
            h.mu.RUnlock()

            if clientCount == 0 {
                continue // No clients, skip broadcasting
            }

            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:  // LINE 284 - NON-BLOCKING SEND
                default:
                    // Client buffer full - just drop the message instead of disconnecting
                    // The client will get the next update
                }
            }
            h.mu.RUnlock()
        }
    }
}
```

**Critical Lines:**
- 251: Filter disabled symbols on new client init
- 284: Non-blocking send to each client

### 7. Hub Statistics Logging
**File:** `backend/ws/hub.go`
**Lines:** 91-110

```go
// logStats logs hub performance metrics every 60 seconds
func (h *Hub) logStats() {
    ticker := time.NewTicker(60 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        received := atomic.LoadInt64(&h.ticksReceived)
        throttled := atomic.LoadInt64(&h.ticksThrottled)
        broadcast := atomic.LoadInt64(&h.ticksBroadcast)

        if received > 0 {
            throttleRate := float64(throttled) / float64(received) * 100
            h.mu.RLock()
            clientCount := len(h.clients)
            h.mu.RUnlock()
            log.Printf("[Hub] Stats: received=%d, broadcast=%d, throttled=%d (%.1f%% reduction), clients=%d",
                received, broadcast, throttled, throttleRate, clientCount)
        }
    }
}
```

**Log Output Format:**
```
[Hub] Stats: received=60000, broadcast=12000, throttled=48000 (80.0% reduction), clients=5
```

---

## Frontend Flow Entry Points

### 1. WebSocket Connection Setup
**File:** `clients/desktop/src/App.tsx`
**Lines:** 127-196

```typescript
useEffect(() => {
    if (!isAuthenticated) return;

    let ws: WebSocket | null = null;
    let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
    let flushInterval: ReturnType<typeof setInterval> | null = null;
    let isUnmounting = false;

    const connect = () => {
        if (isUnmounting) return;

        const authToken = useAppStore.getState().authToken;
        let wsUrl = 'ws://localhost:7999/ws';  // LINE 139
        if (authToken) {
            wsUrl += `?token=${encodeURIComponent(authToken)}`;
        }

        console.log('[WS] Attempting connection to ' + wsUrl);
        ws = new WebSocket(wsUrl);
        wsRef.current = ws;

        ws.onopen = () => console.log('[WS] WebSocket connected');

        ws.onmessage = (event) => {  // LINE 150 - MESSAGE HANDLER
            try {
                const data = JSON.parse(event.data);
                if (data.type === 'tick') {
                    tickBuffer.current[data.symbol] = {
                        ...data,
                        prevBid: ticks[data.symbol]?.bid
                    };
                }
            } catch (e) {
                console.error('[WS] Parse error:', e);
            }
        };

        ws.onclose = (event) => {
            console.log(`[WS] Disconnected (code: ${event.code})`);
            wsRef.current = null;

            if (event.code === 1008 || event.reason === 'Unauthorized') {
                setIsAuthenticated(false);
                return;
            }

            if (!isUnmounting && event.code !== 1000) {
                reconnectTimeout = setTimeout(connect, 2000);
            }
        };
    };

    connect();

    flushInterval = setInterval(() => {  // LINE 181 - FLUSH INTERVAL
        const buffer = tickBuffer.current;
        if (Object.keys(buffer).length > 0) {
            setTicks(prev => ({ ...prev, ...buffer }));  // LINE 184
            tickBuffer.current = {};
        }
    }, 100);  // 100ms interval

    return () => {
        isUnmounting = true;
        if (reconnectTimeout) clearTimeout(reconnectTimeout);
        if (flushInterval) clearInterval(flushInterval);
        if (ws) ws.close(1000, 'Component unmount');
    };

}, [isAuthenticated, brokerConfig]);
```

**Critical Lines:**
- 139: WebSocket URL
- 150: Message handler
- 154-157: Buffer update
- 181-187: 100ms flush interval
- 184: State merge

### 2. Tick Buffer Reference
**File:** `clients/desktop/src/App.tsx`
**Lines:** 124-125

```typescript
// Tick buffer for throttled updates
const tickBuffer = useRef<Record<string, Tick>>({});
```

### 3. Tick Display Component
**File:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx`
**Lines:** 16-30

```typescript
interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;
    prevBid?: number;
    dailyChange?: number;
    high?: number;
    low?: number;
    volume?: number;
    last?: number;
    open?: number;
    close?: number; // Previous close
    tickHistory?: number[]; // For tick chart
}
```

**Component Props (Line 32-38):**
```typescript
interface MarketWatchPanelProps {
    ticks: Record<string, Tick>;
    allSymbols: any[];
    selectedSymbol: string;
    onSymbolSelect: (symbol: string) => void;
    className?: string;
}
```

---

## MarketTick JSON Message Format

**Struct Definition:**
**File:** `backend/ws/hub.go`
**Lines:** 64-72

```go
// MarketTick represents a price update for clients
type MarketTick struct {
    Type      string  `json:"type"`
    Symbol    string  `json:"symbol"`
    Bid       float64 `json:"bid"`
    Ask       float64 `json:"ask"`
    Spread    float64 `json:"spread"`
    Timestamp int64   `json:"timestamp"`
    LP        string  `json:"lp"` // Liquidity Provider source
}
```

**Actual JSON Sent to WebSocket:**
```json
{
    "type": "tick",
    "symbol": "EURUSD",
    "bid": 1.09350,
    "ask": 1.09365,
    "spread": 0.00015,
    "timestamp": 1705768900,
    "lp": "OANDA"
}
```

---

## Configuration & Initialization

### Hub Initialization
**File:** `backend/cmd/server/main.go`
**Lines:** 270-279

```go
hub := ws.NewHub()

// Set tick store on hub for storing incoming ticks
hub.SetTickStore(tickStore)

// Set B-Book engine on hub for dynamic symbol registration
hub.SetBBookEngine(bbookEngine)

// Set auth service on hub for WebSocket authentication
hub.SetAuthService(authService)
```

### Hub Run Start
**File:** `backend/cmd/server/main.go`
**Line:** 326

```go
// Start WebSocket hub
go hub.Run()
```

### LP Manager Activation
**File:** `backend/cmd/server/main.go`
**Line:** 329

```go
// Start LP Manager Aggregation
lpMgr.StartQuoteAggregation()
```

---

## Debugging & Monitoring

### Backend Log Lines to Search
```
[Hub] Stats:                     # 60-second statistics
[Main] Piping quote             # Quote received from LP
[Hub] Client connected          # New WebSocket client
[Hub] Client disconnected       # Client closed connection
throttle                        # Throttling occurred
```

### Frontend Log Lines to Search
```
[WS] Attempting connection      # WebSocket connection attempt
[WS] WebSocket connected        # Connection successful
[WS] Disconnected              # Connection closed
[WS] Parse error               # JSON parsing failed
```

### Key Metrics to Monitor
```
Backend:
- Stats received/broadcast ratio (normal: 60-80% reduction)
- Client count (should be stable)
- Throttle rate (> 85% = buffer pressure)

Frontend:
- WebSocket readyState (1=open, 0=connecting, 2=closing, 3=closed)
- tickBuffer.current length (should be cleared every 100ms)
- setTicks call frequency (should be ~10 Hz)
```

---

## Summary Table

| Component | File | Lines | Function |
|-----------|------|-------|----------|
| Quote Source | backend/lpmanager/ | N/A | StartQuoteAggregation() |
| Pipe Goroutine | backend/cmd/server/main.go | 334-353 | Main loop |
| Spread Calc | backend/cmd/server/main.go | 346 | ask - bid |
| Hub Structure | backend/ws/hub.go | 39-61 | Hub struct |
| Broadcast Main | backend/ws/hub.go | 139-213 | BroadcastTick() |
| Disabled Check | backend/ws/hub.go | 149 | if disabled return |
| Throttle Check | backend/ws/hub.go | 170 | if change < 0.000001 |
| Broadcast Send | backend/ws/hub.go | 208 | Non-blocking send |
| Run Loop | backend/ws/hub.go | 238-293 | Run() |
| New Client Init | backend/ws/hub.go | 250-260 | Send initial prices |
| Client Send | backend/ws/hub.go | 284 | Per-client send |
| Stats Logging | backend/ws/hub.go | 91-110 | logStats() |
| WS Connection | clients/desktop/src/App.tsx | 135-177 | WebSocket setup |
| Message Handler | clients/desktop/src/App.tsx | 150-158 | onmessage |
| Tick Buffer | clients/desktop/src/App.tsx | 124-125 | tickBuffer useRef |
| Flush Interval | clients/desktop/src/App.tsx | 181-187 | 100ms flush |
| Display Component | clients/desktop/src/components/layout/MarketWatchPanel.tsx | 67+ | MarketWatchPanel |

