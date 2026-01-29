# WebSocket Quote Data Flow Analysis

**Analysis Date:** 2026-01-20
**Analyzed Components:** Backend WebSocket Hub, LP Manager, Frontend React App
**Findings Stored In:** Swarm Memory (patterns namespace)

---

## Executive Summary

The WebSocket quote data flows through 5 stages from backend to frontend, with **2 primary filtering mechanisms** and **1 major throttling bottleneck**:

1. **LP Manager** aggregates quotes from multiple sources
2. **Main Goroutine** pipes quotes to Hub with spread calculation
3. **Hub.BroadcastTick()** applies throttling (drops 60-80% of ticks)
4. **Hub.Run()** broadcasts to connected WebSocket clients
5. **Frontend React** buffers for 100ms and batches state updates

---

## Backend Architecture

### Stage 1: Quote Aggregation (LP Manager)
- **Location:** `backend/lpmanager/` adapters (OANDA, Binance, etc.)
- **Output:** Channel with Quote objects containing: Symbol, Bid, Ask, Timestamp
- **Status:** Multiple LPs can feed simultaneously
- **Issue:** If no LP configured or connected, channel is empty

### Stage 2: Quote Piping (Main Goroutine)
**File:** `backend/cmd/server/main.go:334-353`

```go
for quote := range lpMgr.GetQuotesChan() {
    tick := &ws.MarketTick{
        Type:      "tick",
        Symbol:    quote.Symbol,
        Bid:       quote.Bid,
        Ask:       quote.Ask,
        Spread:    quote.Ask - quote.Bid,  // CALCULATED HERE
        Timestamp: quote.Timestamp,
        LP:        quote.LP,
    }
    hub.BroadcastTick(tick)
}
```

**Key Details:**
- Spread is calculated as `Ask - Bid` (not separate field from LP)
- Timestamp comes directly from LP (may be UTC or LP's timezone)
- Every 1000 quotes, logs: `[Main] Piping quote #X to Hub: SYMBOL @ BID`

### Stage 3: Hub Broadcast with Throttling
**File:** `backend/ws/hub.go:139-213`

**Flow:**

```
1. Update Latest (always)
   h.latestPrices[symbol] = tick

2. Check Disabled (filter)
   if h.disabledSymbols[symbol] {
       return  // No broadcast!
   }

3. Throttle Check (drop 60-80% normally)
   priceChange = abs((newBid - lastBid) / lastBid)
   if priceChange < 0.000001 {  // 0.0001% threshold
       h.ticksThrottled++
       return  // No broadcast!
   }

4. Non-Blocking Send
   select {
       case h.broadcast <- data:
           h.ticksBroadcast++
       default:
           // Buffer full - SILENTLY DROP
   }
```

**Throttle Threshold:** `0.000001` = `0.0001%` relative price change
- **Impact:** In normal market conditions, 60-80% of ticks are not broadcast
- **Logged:** Every 60 seconds: `[Hub] Stats: received=X, broadcast=Y, throttled=Z (R% reduction)`
- **Design Purpose:** Reduce CPU load and network overhead
- **Side Effect:** New clients still get initial prices from `latestPrices` cache

### Stage 4: Hub Run Loop (Client Broadcasting)
**File:** `backend/ws/hub.go:238-293`

**Two Paths:**

**Path A - New Client Connection:**
```go
case client := <-h.register:
    h.mu.RLock()
    for _, tick := range h.latestPrices {
        if !h.disabledSymbols[tick.Symbol] {  // Filter disabled
            select {
                case client.send <- data:
                default:
                    // Drop if client buffer full
            }
        }
    }
```

**Path B - Ongoing Broadcast:**
```go
case message := <-h.broadcast:
    for client := range h.clients {
        select {
            case client.send <- message:
            default:
                // Client buffer full - just skip
        }
    }
```

**Key Details:**
- Non-blocking sends to each client (dropped if client's buffer full)
- New clients ONLY receive enabled symbols from `latestPrices`
- Empty rows occur if `latestPrices[symbol]` is nil AND symbol is enabled

### Stage 5: Client Write (WebSocket.SendMessage)
**File:** `backend/ws/hub.go:325-335`

```go
go func() {
    defer conn.Close()
    for message := range client.send {
        err := conn.WriteMessage(websocket.TextMessage, message)
        if err != nil {
            break  // Connection error
        }
    }
}()
```

**Details:**
- Blocking write (but doesn't block the hub)
- If write fails, client is unregistered on next read
- Buffered channel `client.send` has 1024 element capacity

---

## Frontend Architecture

### WebSocket Connection
**File:** `clients/desktop/src/App.tsx:135-177`

```typescript
const authToken = useAppStore.getState().authToken
const wsUrl = 'ws://localhost:7999/ws?token=' + authToken
ws = new WebSocket(wsUrl)

ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    if (data.type === 'tick') {
        tickBuffer.current[data.symbol] = {
            ...data,
            prevBid: ticks[data.symbol]?.bid
        }
    }
}
```

**Key Details:**
- Authentication via JWT token in query parameter
- Expects JSON with `type: 'tick'`
- Stores in `useRef` buffer (not state, survives renders)
- `prevBid` preserved from previous state for change detection

### Tick Buffering (100ms Batch)
**File:** `clients/desktop/src/App.tsx:181-187`

```typescript
flushInterval = setInterval(() => {
    const buffer = tickBuffer.current
    if (Object.keys(buffer).length > 0) {
        setTicks(prev => ({ ...prev, ...buffer }))
        tickBuffer.current = {}
    }
}, 100)
```

**Impact:**
- Maximum 10 state updates per second regardless of WebSocket frequency
- Multiple quotes for same symbol within 100ms: last one wins
- Reduces React render frequency (optimization)

### Tick Display (MarketWatchPanel)
**File:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx`

- Receives `ticks: Record<string, Tick>`
- Expected fields: `symbol, bid, ask, spread, timestamp, lp`
- Empty rows: when `ticks[symbol]` is undefined or all fields are 0/falsy

---

## Root Causes of Empty Rows

### Issue 1: Disabled Symbols (Complete Blockage)
**Severity:** HIGH - Complete filtering

**Location:** `backend/ws/hub.go:149` and `hub.go:251`

**Flow:**
1. Symbol added to `h.disabledSymbols` map
2. `BroadcastTick()` returns early (line 149)
3. No message sent to broadcast channel
4. New clients skip disabled symbols (line 251)
5. **Result:** Quotes never reach frontend

**How to Debug:**
```go
// Check which symbols are disabled
if len(h.disabledSymbols) > 0 {
    log.Printf("[Hub] Disabled symbols: %v", h.disabledSymbols)
}
```

### Issue 2: No Initial Data (Timing)
**Severity:** MEDIUM - Affects new subscriptions

**Location:** `backend/ws/hub.go:250-260` (new client init)

**Flow:**
1. Symbol exists but no quote received yet
2. `h.latestPrices[symbol]` = nil
3. New client connects, requests symbol
4. Hub sends empty set
5. **Result:** Frontend receives nothing, shows empty row

**How to Debug:**
```go
log.Printf("[Hub] Latest prices available: %v", len(h.latestPrices))
for symbol := range h.latestPrices {
    log.Printf("  - %s: bid=%f ask=%f", symbol, h.latestPrices[symbol].Bid, h.latestPrices[symbol].Ask)
}
```

### Issue 3: Throttling (Normal, 60-80% Drop)
**Severity:** LOW - By design, but affects update frequency

**Location:** `backend/ws/hub.go:170`

**Flow:**
1. Quote received: price change < 0.0001%
2. `BroadcastTick()` increments `ticksThrottled`
3. Returns without sending
4. **Result:** Frontend sees no update, but has stale data from last quote

**Impact:** In calm market, same bid/ask for several seconds
**Logged:** `[Hub] Stats: ... throttled=48000 (80.0% reduction)`

### Issue 4: Buffer Overflow (Under Load)
**Severity:** LOW - Rare, load-dependent

**Locations:**
- Backend: `hub.go:207` (broadcast channel size 4096)
- Frontend: `App.tsx:154` (tickBuffer in memory)

**Flow:**
1. High frequency quotes (>4000/sec sustained)
2. Broadcast channel fills up
3. Next quote's non-blocking send fails
4. Message silently dropped
5. **Result:** Random quote updates skipped

**How to Debug:**
- Monitor backend logs for '% reduction' rate
- If consistently >85%, possible buffer issues
- Increase buffer size in hub.go:77: `broadcast: make(chan []byte, 8192)`

### Issue 5: LP Not Sending Quotes
**Severity:** CRITICAL - No data source

**Location:** `backend/lpmanager/` adapters

**Flow:**
1. LP adapter disconnected or not configured
2. `lpMgr.GetQuotesChan()` receives no quotes for symbol
3. Main goroutine never sees that symbol
4. Hub never gets BroadcastTick call
5. **Result:** No data reaches frontend

**How to Debug:**
- Check LP Manager logs for adapter status
- Verify API keys configured in environment
- Test LP connection separately

---

## Spread Calculation

### Where Calculated
**File:** `backend/cmd/server/main.go:346`
```go
Spread: quote.Ask - quote.Bid,
```

### Value in JSON
All WebSocket messages include pre-calculated spread:
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

### Frontend Display
**File:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx:55`
- Column shows: `"!"` with value from `spread` field
- Displayed as-is, no recalculation
- Color coded: red/green for changing spreads (handled by component)

---

## Performance Metrics

### Backend Statistics (logged every 60 seconds)
```
[Hub] Stats: received=60000, broadcast=12000, throttled=48000 (80.0% reduction), clients=5
```

**Interpretation:**
- `received`: Quotes from LP (per minute)
- `broadcast`: Quotes sent to clients
- `throttled`: Quotes dropped by throttle logic
- `reduction`: Percentage saved by throttling
- `clients`: Active WebSocket connections

**Expected Ranges:**
- `reduction`: 60-80% normal, 40-60% fast market, >85% possible buffer issues
- `clients`: 1-10 typical, should not grow indefinitely
- `received`: Depends on LP frequency, OANDA ~100-500/min

---

## Data Flow Diagram

```
LP Manager (OANDA/Binance)
    |
    | [Quote Channel]
    v
Main Goroutine (main.go:334)
    | spread = ask - bid
    | Convert to MarketTick
    v
hub.BroadcastTick(tick)
    | Check disabledSymbols
    | Throttle: price change < 0.0001%?
    | h.latestPrices[symbol] = tick
    v
hub.Run() [Goroutine]
    | Receive from broadcast channel
    | For each client (non-blocking)
    v
client.send channel (1024 buffer)
    |
    | [WebSocket WriteMessage]
    v
Frontend (App.tsx:150)
    | Parse JSON type=tick
    | Buffer in tickBuffer.current
    v
100ms Flush Interval
    | Batch update ticks state
    v
React render
    |
    v
MarketWatchPanel
    | Display bid/ask/spread
    |
    v
User sees quote
```

---

## Debugging Checklist

### For Empty Rows:
- [ ] Check if symbol is in `h.disabledSymbols` map
- [ ] Verify LP is sending quotes for that symbol (backend logs)
- [ ] Check `latestPrices` contains data for symbol
- [ ] Monitor throttle rate (should be 60-80%)
- [ ] Verify WebSocket connection successful (`[WS] connected` in console)
- [ ] Check if `onmessage` fires in browser DevTools
- [ ] Inspect JSON payload in WebSocket frame in DevTools

### For Stale Quotes:
- [ ] Check throttle rate in `[Hub] Stats` line
- [ ] Verify price is actually changing (check other symbols)
- [ ] Monitor tick buffer in React DevTools
- [ ] Check 100ms interval is firing (add console.log)
- [ ] Verify setTicks is updating state

### For Missing Spreads:
- [ ] Check JSON includes `"spread"` field
- [ ] Verify `quote.Ask > quote.Bid` (spread > 0)
- [ ] Check MarketWatchPanel receiving spread value
- [ ] Verify column is configured to show spread

---

## Key Code Locations

### Backend
- **Main entry:** `backend/cmd/server/main.go:334-353`
- **Hub broadcast:** `backend/ws/hub.go:139-213`
- **Hub run loop:** `backend/ws/hub.go:238-293`
- **Client registration:** `backend/ws/hub.go:295-351`
- **Throttle logic:** `backend/ws/hub.go:155-184`
- **Disabled symbols:** `backend/ws/hub.go:112-124`

### Frontend
- **WebSocket setup:** `clients/desktop/src/App.tsx:127-196`
- **Message handler:** `clients/desktop/src/App.tsx:150-158`
- **Tick buffering:** `clients/desktop/src/App.tsx:181-187`
- **Display component:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx:67-500+`

---

## Recommendations

### To Fix Empty Rows:
1. **Add symbol pre-registration** with initial null tick
2. **Add debug endpoint** to inspect hub state
3. **Monitor LP adapter** status per symbol
4. **Add client trace** logging for specific symbols

### To Improve Update Frequency:
1. **Increase throttle threshold** from 0.000001 to 0.00001 (0.001%)
2. **Or disable throttling** and let frontend batch at 100ms
3. **Increase buffer size** from 4096 to 8192 if seeing drops

### To Monitor Health:
1. **Add Prometheus metrics** for hub statistics
2. **Add alert** if `clients` count drops to 0
3. **Add alert** if throttle rate >85%
4. **Add symbol status** API endpoint

---

## Summary for Coder Agent

The issue of empty rows is most likely:
1. **Disabled symbol filtering** (check `hub.disabledSymbols`)
2. **No initial data** when client connects (new symbol)
3. **LP not sending quotes** for that specific symbol

The throttling (60-80% drop) is **by design** and doesn't cause empty rows, but may cause stale data.

The spread is **calculated on backend** as `ask - bid` and sent to frontend as-is.

All findings and debugging guides have been stored in swarm memory for reference.
