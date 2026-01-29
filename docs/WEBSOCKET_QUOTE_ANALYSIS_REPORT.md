# WebSocket Quote Analysis Report
**Date:** 2026-01-19
**Analyst:** Claude Code
**WebSocket Endpoint:** ws://localhost:7999/ws
**Admin Endpoint:** http://localhost:7999/admin/fix/ticks

## Executive Summary

**VERDICT: SIMULATED DATA** 🎭

The Trading Engine is currently broadcasting **simulated market data** for testing purposes. The quotes are NOT coming from real market sources.

---

## Evidence

### 1. LP Field Analysis (PRIMARY EVIDENCE)

**Finding:** All ticks have `"lp": "SIMULATED"` field

```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 0.3872000000000768,
  "ask": 0.3873500000000768,
  "spread": 0.00015000000000000001,
  "timestamp": 1768823017,
  "lp": "SIMULATED"    ← CLEAR SIMULATION MARKER
}
```

**Conclusion:** ❌ The `lp` (Liquidity Provider) field explicitly shows `"SIMULATED"`, which is a definitive marker that these are NOT real market quotes.

---

### 2. Source Code Evidence

From `D:\Tading engine\Trading-Engine\backend\cmd\server\main.go`:

**Lines 926-961: Real FIX Gateway Data Pipeline**
```go
// Pipe FIX market data to WebSocket hub
go func() {
    fixGateway := server.GetFIXGateway()
    // ...
    for md := range fixGateway.GetMarketData() {
        tick := &ws.MarketTick{
            Type:      "tick",
            Symbol:    md.Symbol,
            Bid:       md.Bid,
            Ask:       md.Ask,
            Spread:    md.Ask - md.Bid,
            Timestamp: md.Timestamp.Unix(),
            LP:        "YOFX", // ← Real FIX LP source
        }
        hub.BroadcastTick(tick)
    }
}()
```

**Lines 963-1042: Simulated Fallback Mechanism**
```go
// Simulated market data fallback - enables testing when LP data unavailable
go func() {
    // Wait 30 seconds to see if real market data arrives
    time.Sleep(30 * time.Second)

    tickMutex.RLock()
    hasRealData := totalTickCount > 0
    tickMutex.RUnlock()

    if hasRealData {
        log.Println("[SIM-MD] Real market data detected, simulation not needed")
        return
    }

    log.Println("[SIM-MD] No real market data after 30s - starting simulated prices for testing")
    log.Println("[SIM-MD] NOTE: These are SIMULATED prices for UI testing only!")

    // Generate simulated ticks every 500ms
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    for range ticker.C {
        for symbol, basePrice := range currentPrices {
            // Small random walk: -2 to +2 pips
            pip := pipSizes[symbol]
            change := (float64(time.Now().UnixNano()%5) - 2) * pip
            currentPrices[symbol] = basePrice + change

            tick := &ws.MarketTick{
                Type:      "tick",
                Symbol:    symbol,
                Bid:       bid,
                Ask:       ask,
                Spread:    spread,
                Timestamp: time.Now().Unix(),
                LP:        "SIMULATED", // ← Clearly marked as simulated
            }
            hub.BroadcastTick(tick)
        }
    }
}()
```

**Conclusion:** The server has two data pipelines:
1. **Real data** from FIX gateway (lines 926-961) marked with `LP: "YOFX"`
2. **Simulated data** fallback (lines 963-1042) marked with `LP: "SIMULATED"`

The simulation kicks in after 30 seconds if no real market data arrives.

---

### 3. Price Characteristics

**Simulated Price Patterns:**

| Symbol | Current Bid | Issues |
|--------|-------------|--------|
| EURUSD | 0.38720 | ❌ Unrealistic (should be ~1.0850) |
| GBPUSD | 0.56720 | ❌ Unrealistic (should be ~1.2650) |
| USDJPY | 86.71999 | ❌ Unrealistic (should be ~156.50) |
| AUDUSD | -0.07279 | ❌ **NEGATIVE PRICE** (impossible in real markets) |
| AUDCHF | -0.13179 | ❌ **NEGATIVE PRICE** (impossible in real markets) |
| NZDUSD | -0.13279 | ❌ **NEGATIVE PRICE** (impossible in real markets) |

**Key Issues:**
1. **Negative prices:** Several currency pairs show negative bid prices, which is impossible in real forex markets
2. **Price drift:** The simulation uses a simple random walk that causes prices to drift far from realistic levels
3. **Unrealistic values:** Major pairs like EURUSD showing 0.387 instead of ~1.085

**Code Analysis (lines 1015-1019):**
```go
// Small random walk: -2 to +2 pips
pip := pipSizes[symbol]
change := (float64(time.Now().UnixNano()%5) - 2) * pip
currentPrices[symbol] = basePrice + change  // ← Accumulates changes, causing drift
```

The simulation accumulates price changes without mean reversion, causing unrealistic drift.

---

### 4. Timestamp Patterns

**Expected for Simulation:**
- Regular 500ms intervals (from `time.NewTicker(500 * time.Millisecond)`)
- All symbols update simultaneously
- Perfectly consistent timing

**Evidence:**
All symbols in the latest snapshot have **identical timestamp: 1768823017** (2026-01-19), confirming batch simulation.

---

### 5. Spread Analysis

**Finding:** Perfectly consistent spreads

| Symbol Type | Expected Spread | Actual Spread | Consistency |
|-------------|-----------------|---------------|-------------|
| Major Pairs | 0.00015 | 0.00015 | 100% |
| JPY Pairs | 0.015 | 0.015 | 100% |

**Code Evidence (lines 1022-1023):**
```go
spread := pip * 1.5 // 1.5 pip spread
ask := bid + spread
```

Real markets have **variable spreads** that change based on:
- Market volatility
- Time of day
- Liquidity conditions
- News events

The simulation uses a **fixed 1.5 pip spread**, which is unrealistic.

---

## How to Identify Real vs. Simulated Data

### Real Market Data Indicators (LP: "YOFX")
✅ LP field shows `"YOFX"`
✅ Realistic price levels (e.g., EURUSD ~1.08-1.09)
✅ Variable spreads based on market conditions
✅ Irregular timestamp intervals
✅ All prices are positive
✅ Natural price movement with market volatility

### Simulated Data Indicators (LP: "SIMULATED")
❌ LP field shows `"SIMULATED"`
❌ Unrealistic price levels
❌ Negative prices (impossible)
❌ Perfect spread consistency (1.5 pips always)
❌ Regular 500ms timestamp intervals
❌ Batch updates (all symbols same timestamp)
❌ Simple random walk without mean reversion

---

## Why Is Simulation Running?

Based on the code analysis, the simulation activates because:

1. **FIX Gateway Not Connected:** The real FIX gateway (YOFX) is not providing market data
2. **30-Second Timeout:** After 30 seconds of no real data, simulation starts
3. **Fallback for Testing:** This is intentional to enable UI testing without live market connection

**From server logs (expected):**
```
[SIM-MD] No real market data after 30s - starting simulated prices for testing
[SIM-MD] NOTE: These are SIMULATED prices for UI testing only!
```

---

## Connection to Real Market Data

To receive **real market data** instead of simulated:

### Current Status
The server code shows two FIX sessions should auto-connect:

**Lines 889-923: Auto-Connect FIX Sessions**
```go
go func() {
    time.Sleep(3 * time.Second)

    // Connect YOFX1 (Trading)
    log.Println("[FIX] Auto-connecting YOFX1 session (Trading)...")
    if err := server.ConnectToLP("YOFX1"); err != nil {
        log.Printf("[FIX] Failed to auto-connect YOFX1: %v", err)
    }

    // Connect YOFX2 (Market Data)
    time.Sleep(2 * time.Second)
    log.Println("[FIX] Auto-connecting YOFX2 session (Market Data)...")
    if err := server.ConnectToLP("YOFX2"); err != nil {
        log.Printf("[FIX] Failed to auto-connect YOFX2: %v", err)
    } else {
        // Auto-subscribe to forex symbols
        time.Sleep(2 * time.Second)
        fixGateway := server.GetFIXGateway()
        if fixGateway != nil {
            forexSymbols := []string{"EURUSD", "GBPUSD", "USDJPY"}
            for _, symbol := range forexSymbols {
                fixGateway.SubscribeMarketData("YOFX2", symbol)
            }
        }
    }
}()
```

### Required Steps for Real Data

1. **Configure FIX Sessions:**
   - Edit `D:\Tading engine\Trading-Engine\backend\fix\config\sessions.json`
   - Set correct FIX credentials for YOFX1 and YOFX2

2. **Check FIX Connection Status:**
   ```bash
   curl http://localhost:7999/admin/fix/status
   ```

3. **Verify Market Data Subscription:**
   ```bash
   curl http://localhost:7999/admin/fix/ticks
   ```
   - Check `totalTickCount > 0` for real data
   - Check `lp: "YOFX"` instead of `"SIMULATED"`

4. **Monitor Server Logs:**
   - Look for `[FIX-WS] Piping FIX tick` messages
   - Real data stops simulation: `[SIM-MD] Real market data now available - stopping simulation`

---

## Conclusion

The Trading Engine WebSocket stream at `ws://localhost:7999/ws` is currently broadcasting **SIMULATED DATA** as evidenced by:

1. ✅ **Primary Evidence:** LP field explicitly shows `"SIMULATED"`
2. ✅ **Price Evidence:** Negative prices and unrealistic values
3. ✅ **Code Evidence:** Simulation fallback mechanism is active
4. ✅ **Pattern Evidence:** Perfect spread consistency and regular timestamps

**For real market data:** Configure and connect FIX gateway sessions (YOFX1/YOFX2) to receive `LP: "YOFX"` quotes instead.

---

**Report Generated:** 2026-01-19
**Tool Used:** Direct API inspection + Source code analysis
**Confidence Level:** 100% (Definitive markers present)
