# Spread Calculation Logic Analysis - Frontend Issue Report

## Executive Summary

The spread calculation is working correctly in the backend but there are **issues in the frontend** that prevent proper spread display in the Market Watch table. The issue involves:

1. **Missing `spread` field** in the `MarketWatchItem` type definition
2. **Spread not being calculated** when converting tick data to `MarketWatchItem` format
3. **Spread column renders "!" symbol** because the tick object doesn't have the field populated

---

## Root Cause Analysis

### Backend: CORRECT ✓

**Spread is calculated correctly in backend (3 locations):**

#### 1. FlexyMarkets Quote Conversion (`backend/cmd/server/main.go:341-350`)
```go
tick := &ws.MarketTick{
    Type:      "tick",
    Symbol:    quote.Symbol,
    Bid:       quote.Bid,
    Ask:       quote.Ask,
    Spread:    quote.Ask - quote.Bid,  // ← Correct calculation
    Timestamp: quote.Timestamp,
    LP:        quote.LP,
}
hub.BroadcastTick(tick)
```

#### 2. FIX Protocol Market Data (`backend/cmd/server/main.go:1360-1373`)
```go
tick := &ws.MarketTick{
    Type:      "tick",
    Symbol:    md.Symbol,
    Bid:       md.Bid,
    Ask:       md.Ask,
    Spread:    md.Ask - md.Bid,  // ← Correct calculation
    Timestamp: md.Timestamp.Unix(),
    LP:        "YOFX",
}
hub.BroadcastTick(tick)
```

#### 3. Historical Data Simulation (`backend/cmd/server/main.go:186-194`)
```go
return &ws.MarketTick{
    Type:      "tick",
    Symbol:    cache.Symbol,
    Bid:       bid,
    Ask:       ask,
    Spread:    ask - bid,  // ← Correct calculation
    Timestamp: time.Now().Unix(),
    LP:        "OANDA-HISTORICAL",
}
```

**Backend Summary:**
- ✓ Spread = (Ask - Bid) in absolute pips
- ✓ Not divided by pip size
- ✓ Correctly broadcast via WebSocket to clients

---

### Frontend: ISSUES ✗

#### Issue 1: Missing `spread` Field in Type Definition

**File:** `clients/desktop/src/types/trading.ts` (lines 33-46)

```typescript
export type MarketWatchItem = {
  symbol: string;
  description?: string;
  bid: number;
  ask: number;
  last: number;
  change: number;
  changePercent: number;
  volume: number;
  high24h: number;
  low24h: number;
  timestamp: number;
  direction?: 'up' | 'down' | 'neutral';
  // ❌ MISSING: spread field
};
```

**Impact:** TypeScript compiler doesn't recognize spread as a valid field.

---

#### Issue 2: Spread Not Calculated in Professional MarketWatch

**File:** `clients/desktop/src/components/professional/MarketWatch.tsx` (lines 34-54)

```typescript
const marketItems = useMemo(() => {
  return Object.entries(ticks).map(([symbol, tick]): MarketWatchItem => {
    const mid = (tick.bid + tick.ask) / 2;
    const change = tick.prevBid ? tick.bid - tick.prevBid : 0;
    const changePercent = tick.prevBid ? (change / tick.prevBid) * 100 : 0;

    return {
      symbol,
      bid: tick.bid,
      ask: tick.ask,
      last: mid,
      change,
      changePercent,
      volume: 0,
      high24h: tick.bid,
      low24h: tick.bid,
      timestamp: tick.timestamp,
      direction: change > 0 ? 'up' : change < 0 ? 'down' : 'neutral',
      // ❌ MISSING: No spread calculation
    };
  });
}, [ticks]);
```

**Impact:** Even if the type included spread, it's never calculated when converting ticks to MarketWatchItem.

---

#### Issue 3: Spread Fallback Doesn't Work in Layout Component

**File:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx` (lines 889-891)

```typescript
case 'spread':
    content = tick.spread?.toFixed(0) || '!';
    cellClass = 'text-zinc-400 text-[10px]';
    break;
```

**Current Behavior:**
- Tries to access `tick.spread` from the Tick interface
- Falls back to '!' when undefined
- Since `tick.spread` is often `undefined`, displays '!' instead of calculated spread

**The Tick Interface does have spread:**
```typescript
interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;  // ← Present here
    prevBid?: number;
    dailyChange?: number;
    high?: number;
    low?: number;
    volume?: number;
    last?: number;
    open?: number;
    close?: number;
    tickHistory?: number[];
}
```

**Why this works in MarketWatchPanel but not Professional MarketWatch:**
- MarketWatchPanel reads from `ticks` object which comes directly from WebSocket (which includes spread)
- Professional MarketWatch converts to `MarketWatchItem` type (which doesn't have spread field)

---

## Data Flow Comparison

### ✓ Working Component: MarketWatchPanel (Layout)
```
WebSocket Tick (with spread)
    ↓
Tick interface (spread?: number)
    ↓
Direct rendering in MarketWatchRow
    ↓
Spread displays correctly (shows pip value)
```

### ✗ Broken Component: Professional MarketWatch
```
WebSocket Tick (with spread)
    ↓
Tick interface (spread?: number)
    ↓
Convert to MarketWatchItem (NO spread field) ← Issue here
    ↓
Can't access spread property
    ↓
Spread column never populated
```

---

## The Export CSV Bug Connection

**File:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx` (lines 338-354)

```typescript
const handleExport = useCallback(() => {
    const headers = ['Symbol', 'Bid', 'Ask', 'Spread', 'Daily Change %'];
    const rows = Object.keys(ticks).map(sym => {
        const t = ticks[sym];
        return [sym, t.bid, t.ask, t.spread || (t.ask - t.bid), t.dailyChange || 0].join(',');
        //                           ↑ Fallback calculation here
    });
    // ...
}, [ticks]);
```

**Why CSV export works:**
- It accesses `ticks` directly (which has spread from backend)
- Has fallback: `t.spread || (t.ask - t.bid)`
- This fallback shouldn't be needed if spread was always populated

**Note:** Even though CSV export works, the fallback calculation suggests the code author knew spread might be missing.

---

## Summary of Issues

| Component | File | Issue | Current Behavior |
|-----------|------|-------|-------------------|
| **Tick Type** | `types/trading.ts` | Missing spread field | N/A (interface level) |
| **MarketWatchItem Type** | `types/trading.ts` | No spread field | Can't store spread |
| **Professional MarketWatch** | `professional/MarketWatch.tsx` | Spread not calculated when converting ticks | Spread never passed to MarketWatchItem |
| **MarketWatchPanel** | `layout/MarketWatchPanel.tsx` | Relies on tick.spread directly | Works when spread comes from backend |
| **CSV Export** | `layout/MarketWatchPanel.tsx` | Has fallback calculation | Works, but shouldn't need fallback |

---

## Fixes Required

### Fix 1: Add `spread` to MarketWatchItem Type
**File:** `clients/desktop/src/types/trading.ts`

```typescript
export type MarketWatchItem = {
  symbol: string;
  description?: string;
  bid: number;
  ask: number;
  last: number;
  change: number;
  changePercent: number;
  volume: number;
  high24h: number;
  low24h: number;
  timestamp: number;
  direction?: 'up' | 'down' | 'neutral';
  spread?: number;  // ← ADD THIS
};
```

### Fix 2: Calculate Spread in Professional MarketWatch
**File:** `clients/desktop/src/components/professional/MarketWatch.tsx` (lines 40-52)

```typescript
return {
  symbol,
  bid: tick.bid,
  ask: tick.ask,
  last: mid,
  change,
  changePercent,
  volume: 0,
  high24h: tick.bid,
  low24h: tick.bid,
  timestamp: tick.timestamp,
  direction: change > 0 ? 'up' : change < 0 ? 'down' : 'neutral',
  spread: tick.ask - tick.bid,  // ← ADD THIS
};
```

### Fix 3: Clean Up CSV Export Fallback
**File:** `clients/desktop/src/components/layout/MarketWatchPanel.tsx` (line 343)

```typescript
// After Fix 1 & 2 are done, this can be simplified:
return [sym, t.bid, t.ask, t.spread, t.dailyChange || 0].join(',');
// Instead of:
// return [sym, t.bid, t.ask, t.spread || (t.ask - t.bid), t.dailyChange || 0].join(',');
```

---

## Verification Checklist

After applying fixes:

- [ ] Spread column displays numerical values (not "!") in MarketWatchPanel
- [ ] Spread column displays numerical values in Professional MarketWatch
- [ ] Spread = Ask - Bid (in absolute pips, not divided by pip size)
- [ ] CSV export includes spread values without fallback
- [ ] TypeScript compilation succeeds without type errors
- [ ] No console errors related to undefined spread

---

## Backend Format Verification

Backend sends spread in absolute pips format:

```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.08456,
  "ask": 1.08470,
  "spread": 0.00014,
  "timestamp": 1705704532,
  "lp": "YOFX"
}
```

**Spread interpretation:**
- Value: `0.00014` (absolute pips)
- For 5-decimal pairs (EUR/USD): = 1.4 pips
- Not normalized by pip size (pip size is 0.0001, so 0.00014 / 0.0001 = 1.4 pips)
- Frontend should display the raw value from backend

---

## Conclusion

The spread calculation logic is **correct in the backend** and broadcasts properly. The issue is entirely on the **frontend** where:

1. The TypeScript type doesn't include a spread field
2. When converting ticks to MarketWatchItem, spread is never extracted
3. This causes the spread column to fallback to "!" since the value is undefined

All three fixes above are required for complete resolution of the spread display issue.
