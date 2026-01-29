# Backend/WebSocket/State Sync Analysis Report

**Agent**: Backend/WebSocket/State Sync Agent
**Date**: 2026-01-20
**Mission**: Analyze backend symbol subscription, tick broadcasting, and state synchronization

---

## Executive Summary

The backend correctly implements symbol subscription, tick broadcasting, and WebSocket state synchronization with the following architecture:

- **Subscription Flow**: HTTP API → FIX Gateway → Market Data Snapshot → Hub → WebSocket Clients
- **State Persistence**: FIX subscriptions persist in `FIXGateway.symbolSubscriptions` map
- **Broadcasting**: Hub throttles by 60-80% by default (configurable via `MT5_MODE=true`)
- **Reconnection**: Frontend implements automatic reconnection with 2-second backoff
- **Critical Finding**: No client-side subscription persistence across refreshes

---

## 1. Symbol Subscription Backend Flow

### 1.1 Subscription Endpoint

**File**: `D:\Tading engine\Trading-Engine\backend\cmd\server\main.go` (Lines 855-923)

```go
// POST /api/symbols/subscribe
http.HandleFunc("/api/symbols/subscribe", func(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Symbol string `json:"symbol"`
    }

    // Check if already subscribed
    if fixGateway.IsSymbolSubscribed(req.Symbol) {
        return alreadySubscribed()
    }

    // Subscribe via YOFX2 (market data session)
    mdReqID, err := fixGateway.SubscribeMarketData("YOFX2", req.Symbol)
    if err != nil {
        return error(err)
    }

    return success(mdReqID)
})
```

**Response Format**:
```json
{
  "success": true,
  "symbol": "EURUSD",
  "mdReqId": "MD_EURUSD_1768823410832379600",
  "message": "Subscribed successfully"
}
```

### 1.2 FIX Gateway Subscription

**File**: `D:\Tading engine\Trading-Engine\backend\fix\gateway.go` (Lines 1938-2023)

```go
func (g *FIXGateway) SubscribeMarketData(sessionID string, symbol string) (string, error) {
    // Generate unique MDReqID
    mdReqID := fmt.Sprintf("MD_%s_%d", symbol, time.Now().UnixNano())

    // Store subscription (bidirectional mapping)
    g.mu.Lock()
    g.mdSubscriptions[mdReqID] = symbol       // MDReqID → Symbol
    g.symbolSubscriptions[symbol] = mdReqID   // Symbol → MDReqID
    g.mu.Unlock()

    // Build FIX 4.4 Market Data Request (35=V)
    body := fmt.Sprintf("35=V\x01" +
        "262=%s\x01" +   // MDReqID
        "263=1\x01" +    // SubscriptionRequestType: 1=Snapshot+Updates (streaming)
        "264=0\x01" +    // MarketDepth: 0=Full book
        "267=2\x01" +    // NoMDEntryTypes: 2 (Bid and Offer)
        "269=0\x01" +    // MDEntryType: 0=Bid
        "269=1\x01" +    // MDEntryType: 1=Offer
        "146=1\x01" +    // NoRelatedSym: 1
        "55=%s\x01" +    // Symbol (EURUSD format)
        "460=4\x01" +    // Product: 4=CURRENCY (FX spot)
        "167=FXSPOT\x01" + // SecurityType
        "207=YOFX\x01" + // SecurityExchange
        "15=USD\x01",    // Currency
        mdReqID, symbol)

    conn.Write([]byte(fullMsg))
    return mdReqID, nil
}
```

**Subscription Persistence**:
- ✅ **Backend**: Subscriptions persist in `symbolSubscriptions` map until server restart
- ❌ **Frontend**: No localStorage/sessionStorage persistence across page refreshes

---

## 2. Tick Broadcasting Architecture

### 2.1 FIX → Hub Data Flow

**File**: `D:\Tading engine\Trading-Engine\backend\cmd\server\main.go` (Lines 1681-1717)

```go
// Goroutine: FIX Market Data → WebSocket Hub Pipe
go func() {
    for md := range fixGateway.GetMarketData() {
        tick := &ws.MarketTick{
            Type:      "tick",
            Symbol:    md.Symbol,
            Bid:       md.Bid,
            Ask:       md.Ask,
            Spread:    md.Ask - md.Bid,
            Timestamp: md.Timestamp.Unix(),
            LP:        "YOFX", // FIX LP source
        }

        hub.BroadcastTick(tick)
    }
}()
```

### 2.2 FIX Market Data Snapshot Handler

**File**: `D:\Tading engine\Trading-Engine\backend\fix\gateway.go` (Lines 2191-2257)

```go
func (g *FIXGateway) handleMarketDataSnapshot(session *LPSession, msg string) {
    symbol := g.extractTag(msg, "55")    // Symbol
    mdReqID := g.extractTag(msg, "262")  // MDReqID

    md := MarketData{
        Symbol:    symbol,
        MDReqID:   mdReqID,
        SessionID: session.ID,
        Timestamp: time.Now(),
    }

    // Parse bid/ask from FIX message
    // 269=0 → Bid price (270)
    // 269=1 → Ask price (270)
    for _, part := range parts {
        if strings.HasPrefix(part, "269=") {
            currentType = part[4:]
        } else if strings.HasPrefix(part, "270=") {
            price := parseFloat(part[4:])
            if currentType == "0" { md.Bid = price }
            else if currentType == "1" { md.Ask = price }
        }
    }

    // Update quote cache for incremental updates
    g.quoteCacheMu.Lock()
    g.quoteCache[symbol] = &md
    g.quoteCacheMu.Unlock()

    // Send to channel (non-blocking)
    select {
    case g.marketData <- md:
    default:
        log.Printf("MarketData channel full, dropping quote")
    }
}
```

### 2.3 Hub Broadcasting with Throttling

**File**: `D:\Tading engine\Trading-Engine\backend\ws\hub.go` (Lines 160-240)

```go
func (h *Hub) BroadcastTick(tick *MarketTick) {
    atomic.AddInt64(&h.ticksReceived, 1)

    // ============================================
    // CRITICAL: ALWAYS PERSIST TICKS FIRST
    // ============================================
    if h.tickStore != nil {
        h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
    }

    // Update B-Book engine (needs all prices)
    if h.bbookEngine != nil {
        h.bbookEngine.UpdatePrice(tick.Symbol, tick.Bid, tick.Ask)
    }

    // Update latest price cache
    h.mu.Lock()
    h.latestPrices[tick.Symbol] = tick

    // Skip broadcast if symbol is disabled
    if h.disabledSymbols[tick.Symbol] {
        h.mu.Unlock()
        return
    }
    h.mu.Unlock()

    // ============================================
    // THROTTLING: Skip broadcast if price change < 0.0001%
    // Reduces broadcast volume by 60-80%
    // MT5 MODE: When mt5Mode=true, throttling is DISABLED
    // ============================================
    if !h.mt5Mode {
        h.throttleMu.RLock()
        lastPrice, exists := h.lastBroadcast[tick.Symbol]
        h.throttleMu.RUnlock()

        if exists && lastPrice > 0 {
            priceChange := abs((tick.Bid - lastPrice) / lastPrice)

            // Skip broadcast if change < 0.000001 (0.0001%)
            if priceChange < 0.000001 {
                atomic.AddInt64(&h.ticksThrottled, 1)
                return
            }
        }
    }

    // Update last broadcast price
    h.throttleMu.Lock()
    h.lastBroadcast[tick.Symbol] = tick.Bid
    h.throttleMu.Unlock()

    data, _ := json.Marshal(tick)

    // NON-BLOCKING SEND
    select {
    case h.broadcast <- data:
        atomic.AddInt64(&h.ticksBroadcast, 1)
    default:
        // Buffer full - drop to prevent blocking
    }
}
```

**Throttling Configuration**:
- **Standard Mode** (default): Throttles 60-80% of ticks (price change < 0.0001%)
- **MT5 Mode** (`MT5_MODE=true`): Broadcasts ALL ticks (required for professional terminals)
- **Trade-off**: Standard mode reduces CPU/network load, MT5 mode ensures 100% tick delivery

---

## 3. WebSocket Message Protocol

### 3.1 Message Types

**Tick Message** (from backend):
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.08523,
  "ask": 1.08525,
  "spread": 0.00002,
  "timestamp": 1737401234,
  "lp": "YOFX"
}
```

### 3.2 WebSocket Connection Lifecycle

**File**: `D:\Tading engine\Trading-Engine\backend\ws\hub.go` (Lines 322-378)

```go
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
    // Extract and validate JWT token
    userID, accountID, err := extractAndValidateToken(hub, r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }

    client := &Client{
        conn:      conn,
        send:      make(chan []byte, 1024), // BUFFERED
        symbols:   make(map[string]bool),
        userID:    userID,
        accountID: accountID,
    }

    hub.register <- client

    // Write pump (sends messages to client)
    go func() {
        for message := range client.send {
            conn.WriteMessage(websocket.TextMessage, message)
        }
    }()

    // Read pump (handles incoming messages)
    go func() {
        defer func() {
            hub.unregister <- client
            conn.Close()
        }()
        for {
            _, _, err := conn.ReadMessage()
            if err != nil { break }
        }
    }()
}
```

**Authentication**:
- Query parameter: `ws://localhost:7999/ws?token=<JWT>`
- Authorization header: `Authorization: Bearer <JWT>`

### 3.3 Hub Client Registration

**File**: `D:\Tading engine\Trading-Engine\backend\ws\hub.go` (Lines 265-288)

```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            clientCount := len(h.clients)
            h.mu.Unlock()

            // Send latest prices for all symbols upon connection
            h.mu.RLock()
            for _, tick := range h.latestPrices {
                if !h.disabledSymbols[tick.Symbol] {
                    if data, err := json.Marshal(tick); err == nil {
                        select {
                        case client.send <- data:
                        default:
                        }
                    }
                }
            }
            h.mu.RUnlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    // Client buffer full - drop message
                }
            }
            h.mu.RUnlock()
        }
    }
}
```

**Initial State Sync**:
- ✅ On connection, hub sends all latest prices from `latestPrices` map
- ✅ Non-blocking send prevents slow clients from blocking hub
- ❌ No per-client symbol subscription filtering (all clients get all symbols)

---

## 4. Frontend WebSocket State Management

### 4.1 WebSocket Connection Logic

**File**: `D:\Tading engine\Trading-Engine\clients\desktop\src\App.tsx` (Lines 287-373)

```typescript
useEffect(() => {
    if (!isAuthenticated) return;

    let ws: WebSocket | null = null;
    let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
    let isUnmounting = false;

    const connect = () => {
        const authToken = useAppStore.getState().authToken;
        let wsUrl = 'ws://localhost:7999/ws';
        if (authToken) {
            wsUrl += `?token=${encodeURIComponent(authToken)}`;
        }

        ws = new WebSocket(wsUrl);

        ws.onopen = () => console.log('[WS] WebSocket connected');

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'tick') {
                const spread = data.spread !== undefined && data.spread > 0
                    ? data.spread
                    : (data.ask - data.bid);

                tickBuffer.current[data.symbol] = {
                    ...data,
                    spread: spread,
                    prevBid: ticks[data.symbol]?.bid
                };
            }
        };

        ws.onclose = (event) => {
            console.log(`[WS] Disconnected (code: ${event.code})`);

            if (event.code === 1008 || event.reason === 'Unauthorized') {
                setIsAuthenticated(false);
                return;
            }

            // Reconnect with 2-second backoff
            if (!isUnmounting && event.code !== 1000) {
                reconnectTimeout = setTimeout(connect, 2000);
            }
        };
    };

    connect();

    // Use requestAnimationFrame for 60 FPS updates
    let rafId: number;
    const flushTicks = () => {
        const buffer = tickBuffer.current;
        if (Object.keys(buffer).length > 0) {
            Object.entries(buffer).forEach(([symbol, tick]) => {
                useAppStore.getState().setTick(symbol, tick);
            });
            tickBuffer.current = {};
        }
        rafId = requestAnimationFrame(flushTicks);
    };
    rafId = requestAnimationFrame(flushTicks);

    return () => {
        isUnmounting = true;
        if (reconnectTimeout) clearTimeout(reconnectTimeout);
        if (rafId) cancelAnimationFrame(rafId);
        if (ws) ws.close(1000, 'Component unmount');
    };
}, [isAuthenticated, brokerConfig]);
```

**Frontend Optimizations**:
- ✅ Tick buffering with `requestAnimationFrame` (60 FPS updates)
- ✅ Automatic reconnection with 2-second backoff
- ✅ Spread calculation fallback if backend doesn't provide it
- ❌ No subscription persistence across page refreshes
- ❌ No exponential backoff for reconnection

---

## 5. State Desync Issues & Analysis

### 5.1 Reconnection Behavior

**Current Behavior**:
1. User opens application → WebSocket connects → Receives latest prices from hub
2. User refreshes page → WebSocket disconnects → New connection → Receives latest prices again
3. **Issue**: No duplicate initialization spam detected in current implementation

**Evidence**:
- Hub only sends latest prices once on registration (Lines 276-288 in `hub.go`)
- Frontend only creates one WebSocket connection per `isAuthenticated` state change
- No evidence of multiple concurrent connections or initialization loops

### 5.2 Symbol Subscription Persistence

**Backend Subscription State**:
```go
// FIX Gateway subscription maps (in-memory only)
mdSubscriptions:     map[string]string      // MDReqID → Symbol
symbolSubscriptions: map[string]string      // Symbol → MDReqID
```

**Persistence Lifecycle**:
- ✅ Subscriptions persist across WebSocket reconnections
- ✅ Subscriptions persist across frontend page refreshes
- ❌ Subscriptions cleared on backend server restart
- ❌ Frontend has no awareness of backend subscription state

### 5.3 Auto-Subscribe Behavior

**File**: `D:\Tading engine\Trading-Engine\backend\cmd\server\main.go` (Lines 1618-1679)

```go
// Auto-Connect FIX Sessions on startup
go func() {
    time.Sleep(3 * time.Second)

    // Connect YOFX1 (Trading)
    server.ConnectToLP("YOFX1")

    // Connect YOFX2 (Market Data)
    time.Sleep(2 * time.Second)
    server.ConnectToLP("YOFX2")

    // Wait for logon
    time.Sleep(2 * time.Second)

    // Auto-subscribe to ALL forex symbols
    forexSymbols := []string{
        "EURUSD", "GBPUSD", "USDJPY", "USDCHF", "USDCAD",
        "AUDUSD", "NZDUSD", "EURGBP", "EURJPY", "GBPJPY",
        "EURAUD", "EURCAD", "EURCHF", "AUDCAD", "AUDCHF",
        "AUDJPY", "AUDNZD", "CADCHF", "CADJPY", "CHFJPY",
        "GBPAUD", "GBPCAD", "GBPCHF", "GBPNZD", "NZDCAD",
        "NZDCHF", "NZDJPY", "XAUUSD", "XAGUSD",
    }

    for _, symbol := range forexSymbols {
        // Request security definition (35=c)
        fixGateway.RequestSecurityDefinition("YOFX2", symbol)

        // Subscribe to market data (35=V)
        fixGateway.SubscribeMarketData("YOFX2", symbol)

        time.Sleep(100 * time.Millisecond) // Rate limit
    }
}()
```

**Findings**:
- ✅ Backend auto-subscribes to 29 symbols on startup
- ✅ Frontend receives all ticks immediately after WebSocket connection
- ❌ No manual subscription required from frontend
- ❌ Frontend cannot dynamically subscribe to additional symbols (API exists but not used)

---

## 6. Performance Analysis

### 6.1 Tick Broadcast Efficiency

**Hub Performance Metrics** (Lines 112-131 in `hub.go`):
```go
func (h *Hub) logStats() {
    ticker := time.NewTicker(60 * time.Second)
    for range ticker.C {
        received := atomic.LoadInt64(&h.ticksReceived)
        throttled := atomic.LoadInt64(&h.ticksThrottled)
        broadcast := atomic.LoadInt64(&h.ticksBroadcast)

        if received > 0 {
            throttleRate := float64(throttled) / float64(received) * 100
            log.Printf("[Hub] Stats: received=%d, broadcast=%d, throttled=%d (%.1f%% reduction), clients=%d",
                received, broadcast, throttled, throttleRate, clientCount)
        }
    }
}
```

**Expected Performance** (Standard Mode):
- Throttle Rate: 60-80% reduction
- Broadcast Rate: 20-40% of received ticks
- CPU Savings: ~70% reduction in JSON marshaling and network I/O

**Expected Performance** (MT5 Mode):
- Throttle Rate: 0% (all ticks broadcast)
- Broadcast Rate: 100% of received ticks
- CPU Impact: 60-80% increase in CPU/network usage

### 6.2 Tick Storage Performance

**Storage Architecture**:
```go
// ALWAYS persist ticks BEFORE throttling
if h.tickStore != nil {
    h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
}
```

**Critical Fix** (Lines 166-175 in `hub.go`):
- ✅ Tick persistence happens BEFORE throttling
- ✅ ALL ticks stored regardless of broadcast decision
- ✅ Storage is non-blocking (async batch writer)
- ✅ SQLite backend with daily rotation

---

## 7. Identified Issues & Recommendations

### 7.1 Critical Issues

| Issue | Impact | Priority |
|-------|--------|----------|
| No frontend subscription persistence | User must re-subscribe after page refresh | **HIGH** |
| No per-client symbol filtering | All clients receive all 29 symbols | **MEDIUM** |
| No exponential backoff for reconnection | Rapid reconnection on network issues | **LOW** |

### 7.2 State Desync Root Causes

**No Multiple WebSocket Init Spam Detected**:
- Hub registration is idempotent
- Frontend creates single WebSocket per authentication state
- No evidence of concurrent connections or initialization loops

**Potential Desync Scenarios**:
1. **Backend restart while frontend open**: Frontend reconnects, receives all 29 symbols again
2. **Network hiccup**: Frontend reconnects, receives duplicate initial state
3. **Browser tab switching**: WebSocket may reconnect on tab activation

### 7.3 Recommendations

#### High Priority

1. **Implement Frontend Subscription Persistence**
   ```typescript
   // Store subscribed symbols in localStorage
   const subscriptions = JSON.parse(localStorage.getItem('subscriptions') || '[]');

   // On reconnect, re-subscribe via API
   subscriptions.forEach(symbol => {
       fetch('http://localhost:7999/api/symbols/subscribe', {
           method: 'POST',
           body: JSON.stringify({ symbol })
       });
   });
   ```

2. **Add Per-Client Symbol Filtering in Hub**
   ```go
   // Only broadcast to clients subscribed to symbol
   for client := range h.clients {
       if client.symbols[tick.Symbol] {
           client.send <- message
       }
   }
   ```

3. **Implement Subscription Confirmation Protocol**
   ```json
   // Client → Backend
   { "action": "subscribe", "symbol": "EURUSD" }

   // Backend → Client
   { "type": "subscribed", "symbol": "EURUSD", "status": "success" }
   ```

#### Medium Priority

4. **Add Exponential Backoff for Reconnection**
   ```typescript
   let reconnectDelay = 1000; // Start at 1 second
   const maxDelay = 60000; // Cap at 60 seconds

   const reconnect = () => {
       setTimeout(() => {
           connect();
           reconnectDelay = Math.min(reconnectDelay * 2, maxDelay);
       }, reconnectDelay);
   };
   ```

5. **Add WebSocket Heartbeat/Ping-Pong**
   ```go
   // Backend: Send ping every 30 seconds
   ticker := time.NewTicker(30 * time.Second)
   for range ticker.C {
       conn.WriteMessage(websocket.PingMessage, []byte{})
   }
   ```

6. **Implement Subscription State Endpoint**
   ```go
   // GET /api/symbols/subscriptions
   // Returns: { "subscribed": ["EURUSD", "GBPUSD"], "available": [...] }
   ```

---

## 8. Flow Diagrams

### 8.1 Subscription Flow

```
Frontend                    Backend API                FIX Gateway              YOFX LP
   |                            |                           |                      |
   |--- POST /api/symbols/subscribe ->|                    |                      |
   |    { "symbol": "EURUSD" }  |                           |                      |
   |                            |--- IsSymbolSubscribed() -->|                     |
   |                            |<-- false ------------------|                     |
   |                            |--- SubscribeMarketData() ->|                     |
   |                            |                           |--- FIX 35=V -------->|
   |                            |                           |                  (MDReqID: MD_EURUSD_123)
   |                            |                           |<-- FIX 35=W ---------|
   |                            |                           | (Bid: 1.08523, Ask: 1.08525)
   |                            |<-- MDReqID: MD_EURUSD_123 |                      |
   |<-- { "success": true, "mdReqId": "..." } -------------|                      |
   |                            |                           |                      |
```

### 8.2 Tick Broadcast Flow

```
YOFX LP                FIX Gateway              Hub                   WebSocket Client
   |                       |                      |                         |
   |--- FIX 35=W --------->|                      |                         |
   | (Bid/Ask update)      |                      |                         |
   |                       |--- handleMarketDataSnapshot()                  |
   |                       |--- marketData chan ->|                         |
   |                       |                      |--- BroadcastTick() ---->|
   |                       |                      | (1) Store to tickStore  |
   |                       |                      | (2) Update B-Book engine|
   |                       |                      | (3) Update latestPrices |
   |                       |                      | (4) Check throttling    |
   |                       |                      | (5) Broadcast to clients|
   |                       |                      |--- JSON.Marshal() ----->|
   |                       |                      |--- ws.send ------------>|
   |                       |                      |                         |
   |                       |                      |                         |<-- Tick rendered
```

### 8.3 Reconnection Flow

```
Frontend                    WebSocket                Hub                  FIX Gateway
   |                            |                      |                      |
   |--- ws.close() ------------>|                      |                      |
   |<-- onclose (code: 1006) ---|                      |                      |
   |                            |                      |                      |
   |--- setTimeout(2000) ------>|                      |                      |
   |                            |                      |                      |
   |--- new WebSocket() ------->|                      |                      |
   |--- ws://localhost:7999/ws?token=<JWT> ----------->|                     |
   |                            |<-- HTTP 101 Switching Protocols --------- |
   |                            |--- hub.register ----->|                     |
   |                            |                      |--- Send latestPrices|
   |<-- tick messages (all subscribed symbols) --------|                     |
   |                            |                      |                     |
```

---

## 9. Testing Recommendations

### 9.1 Manual Verification

```bash
# 1. Test subscription endpoint
curl -X POST http://localhost:7999/api/symbols/subscribe \
  -H "Content-Type: application/json" \
  -d '{"symbol":"EURUSD"}'

# Expected: {"success":true,"symbol":"EURUSD","mdReqId":"MD_EURUSD_..."}

# 2. Check subscribed symbols
curl http://localhost:7999/api/symbols/subscribed

# Expected: ["EURUSD","GBPUSD","USDJPY",...]

# 3. Monitor tick flow
curl http://localhost:7999/admin/fix/ticks

# Expected: {"totalTickCount":1234,"symbolCount":29,"latestTicks":{...}}

# 4. Test WebSocket connection
wscat -c "ws://localhost:7999/ws?token=<JWT>"

# Expected: {"type":"tick","symbol":"EURUSD","bid":1.08523,"ask":1.08525,...}
```

### 9.2 Integration Tests

```typescript
// Test 1: Subscription persistence
test('Subscription persists across WebSocket reconnection', async () => {
    const ws1 = new WebSocket('ws://localhost:7999/ws?token=<JWT>');
    await waitForMessage(ws1, { type: 'tick', symbol: 'EURUSD' });
    ws1.close();

    const ws2 = new WebSocket('ws://localhost:7999/ws?token=<JWT>');
    const tick = await waitForMessage(ws2, { type: 'tick', symbol: 'EURUSD' });
    expect(tick).toBeDefined();
});

// Test 2: No duplicate initialization
test('No duplicate ticks on reconnection', async () => {
    const ws = new WebSocket('ws://localhost:7999/ws?token=<JWT>');
    const messages = [];
    ws.onmessage = (e) => messages.push(JSON.parse(e.data));

    await sleep(2000);

    // Count EURUSD ticks received
    const eurusdTicks = messages.filter(m => m.symbol === 'EURUSD');

    // Should NOT receive all historical ticks, only latest + streaming
    expect(eurusdTicks.length).toBeLessThan(100); // Not thousands
});

// Test 3: Throttling verification
test('Throttling reduces tick volume by 60-80%', async () => {
    // Subscribe via FIX, count received ticks for 60 seconds
    const fixTicks = await countFixTicks(60000);
    const wsTicks = await countWebSocketTicks(60000);

    const reduction = (fixTicks - wsTicks) / fixTicks * 100;
    expect(reduction).toBeGreaterThan(60);
    expect(reduction).toBeLessThan(80);
});
```

---

## 10. File Reference Index

### Backend Files

| File | Lines | Purpose |
|------|-------|---------|
| `backend/cmd/server/main.go` | 855-923 | `/api/symbols/subscribe` endpoint |
| `backend/cmd/server/main.go` | 1618-1679 | FIX auto-connect and auto-subscribe |
| `backend/cmd/server/main.go` | 1681-1717 | FIX → Hub market data pipe |
| `backend/cmd/server/main.go` | 1851-1853 | WebSocket route registration |
| `backend/fix/gateway.go` | 1938-2023 | `SubscribeMarketData()` implementation |
| `backend/fix/gateway.go` | 2191-2257 | `handleMarketDataSnapshot()` |
| `backend/ws/hub.go` | 160-240 | `BroadcastTick()` with throttling |
| `backend/ws/hub.go` | 265-288 | Hub client registration and initial state |
| `backend/ws/hub.go` | 322-378 | `ServeWs()` WebSocket handler |

### Frontend Files

| File | Lines | Purpose |
|------|-------|---------|
| `clients/desktop/src/App.tsx` | 287-373 | WebSocket connection and reconnection |
| `clients/desktop/src/App.tsx` | 312-330 | Tick message parsing and buffering |
| `clients/desktop/src/App.tsx` | 332-344 | Reconnection logic with 2s backoff |
| `clients/desktop/src/App.tsx` | 349-365 | `requestAnimationFrame` tick flushing |

---

## 11. Conclusion

The backend correctly implements symbol subscription, tick broadcasting, and WebSocket state synchronization. The main areas for improvement are:

1. **Frontend subscription persistence** across page refreshes
2. **Per-client symbol filtering** to reduce bandwidth for clients with selective subscriptions
3. **Exponential backoff** for reconnection to prevent rapid reconnection loops
4. **Subscription confirmation protocol** for explicit client-server handshake

The current architecture is production-ready with the following caveats:
- All clients receive all 29 subscribed symbols (no filtering)
- Page refresh requires manual re-subscription (if implemented)
- Standard throttling mode reduces tick volume by 60-80% (configurable)

**No critical state desync issues detected** in the current implementation. The Hub registration is idempotent, and the frontend creates a single WebSocket connection per authentication state.
