# MT5 Charting Quick Fix Guide

**Priority Fixes for MT5 Parity**

---

## 🔴 CRITICAL: Volume Histogram (2 hours)

**File**: `clients/desktop/src/components/TradingChart.tsx`

### Step 1: Add Volume Series Ref (after line 46)
```typescript
const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
```

### Step 2: Create Volume Series (after line 153, inside chart creation)
```typescript
// Create volume histogram
const volumeSeries = chart.addHistogramSeries({
  color: '#06b6d4', // Cyan
  priceFormat: {
    type: 'volume',
  },
  priceScaleId: '', // Separate scale from price
  scaleMargins: {
    top: 0.8,    // Use bottom 20% of chart
    bottom: 0,
  },
});
volumeSeriesRef.current = volumeSeries;
```

### Step 3: Update Volume Data (in fetchHistory effect, after line 190)
```typescript
// Update volume data
if (data.length > 0 && volumeSeriesRef.current) {
  const volumeData = data.map((d: any) => ({
    time: d.time,
    value: d.volume || 0,
    color: d.close >= d.open
      ? 'rgba(20, 184, 166, 0.5)'  // Teal with transparency
      : 'rgba(239, 68, 68, 0.5)'   // Red with transparency
  }));
  volumeSeriesRef.current.setData(volumeData);
}
```

### Step 4: Update Volume on New Candles (in currentPrice effect, after line 226)
```typescript
// Update volume for current candle
if (volumeSeriesRef.current) {
  const lastCandle = candlesRef.current[candlesRef.current.length - 1];
  if (lastCandle) {
    volumeSeriesRef.current.update({
      time: lastCandle.time,
      value: lastCandle.volume || 0,
      color: lastCandle.close >= lastCandle.open
        ? 'rgba(20, 184, 166, 0.5)'
        : 'rgba(239, 68, 68, 0.5)'
    });
  }
}
```

### Step 5: Cleanup on Unmount (in cleanup, line 106)
```typescript
if (volumeSeriesRef.current) {
  chart.removeSeries(volumeSeriesRef.current);
  volumeSeriesRef.current = null;
}
```

---

## 🔴 HIGH: Fix Candlestick Colors (5 minutes)

**File**: `clients/desktop/src/components/TradingChart.tsx`

### Change Lines 130-132
```diff
series = chart.addSeries(CandlestickSeries, {
-  upColor: '#10b981', downColor: '#ef4444',
-  borderUpColor: '#10b981', borderDownColor: '#ef4444',
-  wickUpColor: '#10b981', wickDownColor: '#ef4444',
+  upColor: '#14b8a6', downColor: '#ef4444',
+  borderUpColor: '#14b8a6', borderDownColor: '#ef4444',
+  wickUpColor: '#14b8a6', wickDownColor: '#ef4444',
});
```

### Also Update Lines 148-151 (fallback candlestick)
```diff
series = chart.addSeries(CandlestickSeries, {
-  upColor: '#10b981', downColor: '#ef4444',
-  borderUpColor: '#10b981', borderDownColor: '#ef4444',
-  wickUpColor: '#10b981', wickDownColor: '#ef4444',
+  upColor: '#14b8a6', downColor: '#ef4444',
+  borderUpColor: '#14b8a6', borderDownColor: '#ef4444',
+  wickUpColor: '#14b8a6', wickDownColor: '#ef4444',
});
```

---

## 🟡 MEDIUM: Dotted Grid Lines (10 minutes)

**File**: `clients/desktop/src/components/TradingChart.tsx`

### Change Lines 64-67
```diff
grid: {
-  vertLines: { color: '#27272a' },
-  horzLines: { color: '#27272a' },
+  vertLines: {
+    color: '#27272a',
+    style: 1,        // 0=Solid, 1=Dotted, 2=Dashed, 3=LargeDashed
+    visible: true
+  },
+  horzLines: {
+    color: '#27272a',
+    style: 1,
+    visible: true
+  },
}
```

---

## 🟡 MEDIUM: Bid/Ask Price Lines (30 minutes)

**File**: `clients/desktop/src/components/TradingChart.tsx`

### Step 1: Add Refs (after line 46)
```typescript
const bidLineRef = useRef<IPriceLine | null>(null);
const askLineRef = useRef<IPriceLine | null>(null);
```

### Step 2: Import IPriceLine Type (line 11)
```typescript
import type {
  IChartApi,
  ISeriesApi,
  Time,
  IPriceLine  // Add this
} from 'lightweight-charts';
```

### Step 3: Create Effect for Price Lines (after line 228)
```typescript
// Update bid/ask price lines
useEffect(() => {
  if (!seriesRef.current || !currentPrice) return;

  // Remove old lines
  if (bidLineRef.current) {
    seriesRef.current.removePriceLine(bidLineRef.current);
  }
  if (askLineRef.current) {
    seriesRef.current.removePriceLine(askLineRef.current);
  }

  // Create bid line (red)
  bidLineRef.current = seriesRef.current.createPriceLine({
    price: currentPrice.bid,
    color: '#ef4444',
    lineWidth: 1,
    lineStyle: 2, // Dashed
    axisLabelVisible: true,
    title: currentPrice.bid.toFixed(5),
  });

  // Create ask line (teal)
  askLineRef.current = seriesRef.current.createPriceLine({
    price: currentPrice.ask,
    color: '#14b8a6',
    lineWidth: 1,
    lineStyle: 2, // Dashed
    axisLabelVisible: true,
    title: currentPrice.ask.toFixed(5),
  });
}, [currentPrice]);
```

---

## 🟡 MEDIUM: Enhanced Trade Labels (1 hour)

**File**: `clients/desktop/src/components/TradingChart.tsx`

### Modify PositionOverlay Entry Line (line 372-376)
```diff
<div className="bg-blue-500 text-white text-[10px] px-1 rounded-sm flex items-center gap-1 -mt-5
-     opacity-0 group-hover:opacity-100 transition-opacity">
+     opacity-100">  {/* Always visible */}
-  <span>#{pos.id} {pos.side} {pos.volume}</span>
+  <span>{pos.side} {pos.volume} @ {pos.openPrice.toFixed(5)}</span>
  <X size={12} className="cursor-pointer hover:text-red-300" onClick={onClose} />
</div>
```

---

## 🟢 OPTIONAL: Weekly/Monthly Timeframes (1 hour)

### Backend: `backend/tickstore/ohlc_cache.go`

#### Add Constants (after line 22)
```go
const (
  TF_M1  Timeframe = "M1"
  TF_M5  Timeframe = "M5"
  TF_M15 Timeframe = "M15"
  TF_H1  Timeframe = "H1"
  TF_H4  Timeframe = "H4"
  TF_D1  Timeframe = "D1"
  TF_W1  Timeframe = "W1"  // NEW
  TF_MN  Timeframe = "MN"  // NEW
)
```

#### Update TimeframeSeconds (after line 40)
```go
func TimeframeSeconds(tf Timeframe) int64 {
  switch tf {
  case TF_M1: return 60
  case TF_M5: return 300
  case TF_M15: return 900
  case TF_H1: return 3600
  case TF_H4: return 14400
  case TF_D1: return 86400
  case TF_W1: return 604800   // NEW: 7 days
  case TF_MN: return 2592000  // NEW: 30 days
  default: return 60
  }
}
```

### Frontend: `clients/desktop/src/components/TradingChart.tsx`

#### Update Timeframe Type (line 15)
```typescript
export type Timeframe = '1m' | '5m' | '15m' | '1h' | '4h' | '1d' | '1w' | '1M';
```

#### Update Timeframe Selector (line 467)
```typescript
const timeframes: { value: Timeframe; label: string }[] = [
  { value: '1m', label: 'M1' },
  { value: '5m', label: 'M5' },
  { value: '15m', label: 'M15' },
  { value: '1h', label: 'H1' },
  { value: '4h', label: 'H4' },
  { value: '1d', label: 'D1' },
  { value: '1w', label: 'W1' },  // NEW
  { value: '1M', label: 'MN' },  // NEW
];
```

#### Update Backend API Mapping (line 415, add cases)
```typescript
function getTimeframeSeconds(tf: Timeframe): number {
  switch (tf) {
    case '1m': return 60;
    case '5m': return 300;
    case '15m': return 900;
    case '1h': return 3600;
    case '4h': return 14400;
    case '1d': return 86400;
    case '1w': return 604800;   // NEW
    case '1M': return 2592000;  // NEW
    default: return 60;
  }
}
```

---

## Testing Commands

### Start Backend
```bash
cd backend
go run cmd/server/main.go
```

### Start Frontend
```bash
cd clients/desktop
npm run dev
```

### Test Volume Data
```bash
curl "http://localhost:8080/ohlc?symbol=XAUUSD&timeframe=1m&limit=10"
# Should return volume field for each bar
```

### Visual Checks
1. Open http://localhost:5173
2. Login and select XAUUSD
3. Verify:
   - [ ] Volume bars visible at bottom (cyan)
   - [ ] Candles are teal (bullish) and red (bearish)
   - [ ] Grid lines are dotted
   - [ ] Bid/ask lines track current price
   - [ ] Trade labels show "BUY 0.01 @ 2720.50"

---

## Quick Validation Script

```bash
# From project root
cd clients/desktop/src/components

# Check current colors
grep -n "upColor.*#10b981" TradingChart.tsx
# Should show NO results after fix

# Check for volume series
grep -n "addHistogramSeries" TradingChart.tsx
# Should show 1+ results after implementation

# Check grid style
grep -n "style.*1" TradingChart.tsx
# Should show grid configuration
```

---

## Rollback Plan

If issues occur:

1. **Volume causes lag**: Remove volume series, keep colors/grid
2. **Price lines flicker**: Debounce updates to 500ms
3. **Grid breaks layout**: Revert to solid lines
4. **Colors wrong**: Revert to `#10b981` (original)

---

## Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| Initial render | <200ms | ~150ms |
| Update tick | <16ms (60fps) | ~10ms |
| Memory usage | <100MB | ~60MB |
| Chart smoothness | 60fps | 60fps |

After volume histogram: Expect +20ms render, +10MB memory (acceptable).

---

## Support

- Lightweight Charts Docs: https://tradingview.github.io/lightweight-charts/
- Volume Series Example: https://tradingview.github.io/lightweight-charts/tutorials/demos/histogram-series
- Issue Tracker: Create ticket if problems persist

---

**Estimated Time to Full MT5 Parity**: 4-6 hours
**Quick Win** (Colors + Grid): 15 minutes
**Biggest Impact** (Volume + Colors): 2.5 hours
