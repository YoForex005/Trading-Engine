# MT5 Charting System Analysis & Implementation Plan

**Agent**: Charting Agent
**Date**: 2026-01-20
**Mission**: Achieve MT5 parity for charting system

---

## Executive Summary

Current implementation uses **lightweight-charts v5.1.0** with basic candlestick rendering. The system has OHLC data with volume fields from backend but **volume is NOT rendered**. Multiple critical features missing for MT5 parity.

**Status**: 🔴 Significant gaps identified
**Complexity**: Medium-High (requires volume series, trade overlays, color fixes)
**Estimated Effort**: 4-6 hours implementation + 2 hours testing

---

## Current Architecture

### Frontend Components
```
clients/desktop/src/components/
├── TradingChart.tsx          # Main chart component (uses lightweight-charts)
├── ChartWithHistory.tsx      # Wrapper with historical data integration
└── ChartControls.tsx         # Timeframe/chart type selector
```

### Backend APIs
```
GET /ohlc?symbol=XAUUSD&timeframe=1m&limit=500
Response: [
  {
    "time": 1737364800,
    "open": 2720.50,
    "high": 2720.80,
    "low": 2720.30,
    "close": 2720.60,
    "volume": 125  ← EXISTS BUT NOT USED IN FRONTEND
  }
]
```

### Data Flow
```
FIX Gateway → TickStore → OHLC Aggregation → API → Frontend Chart
                   ↓
            Volume counted per bar (tick_count)
```

---

## Gap Analysis: Current vs MT5

| Feature | MT5 | Current | Status |
|---------|-----|---------|--------|
| **Candlesticks** | Teal/Red | Emerald/Red | ❌ Wrong color |
| **Volume Bars** | Cyan histogram | Not rendered | ❌ Missing |
| **Grid System** | Dotted lines | Default solid | ⚠️ Needs customization |
| **Trade Levels** | BUY/SELL lines | Entry lines only | ⚠️ Incomplete |
| **Price Labels** | Floating current | Static scale | ⚠️ Needs enhancement |
| **SL/TP Lines** | Dashed lines | Dashed (correct) | ✅ Working |
| **Timeframes** | M1-MN support | M1-D1 only | ⚠️ Missing W1, MN |

---

## Critical Issues

### 1. Volume Histogram Not Rendered
**Problem**: Backend provides `volume` field in OHLC data, but frontend TradingChart.tsx doesn't render it.

**Root Cause**: No `addHistogramSeries()` call in TradingChart component.

**Impact**: Users cannot see volume analysis (essential for MT5 traders).

**Fix Required**:
```typescript
// In TradingChart.tsx - Add volume series
const volumeSeries = chart.addHistogramSeries({
  color: '#06b6d4', // Cyan
  priceFormat: { type: 'volume' },
  priceScaleId: '', // Separate scale
  scaleMargins: { top: 0.8, bottom: 0 } // Bottom 20% of chart
});

// Update volume data when OHLC changes
const volumeData = data.map(d => ({
  time: d.time,
  value: d.volume,
  color: d.close >= d.open ? '#14b8a6' : '#ef4444' // Teal/Red
}));
volumeSeries.setData(volumeData);
```

### 2. Candlestick Colors Wrong
**Problem**: Current uses `#10b981` (emerald) for bullish candles. MT5 uses `#14b8a6` (teal).

**Current** (lines 130-133):
```typescript
upColor: '#10b981', downColor: '#ef4444',
```

**Should be**:
```typescript
upColor: '#14b8a6', downColor: '#ef4444',
```

### 3. Grid System Not Customized
**Current** (lines 64-67):
```typescript
grid: {
  vertLines: { color: '#27272a' },
  horzLines: { color: '#27272a' },
}
```

**MT5 Style**:
```typescript
grid: {
  vertLines: {
    color: '#27272a',
    style: 1, // Dotted
    visible: true
  },
  horzLines: {
    color: '#27272a',
    style: 1, // Dotted
    visible: true
  }
}
```

### 4. Trade Level Lines Incomplete
**Current**: Shows entry line, SL, TP for open positions (lines 367-411).

**Missing**:
- Pending order lines (different color, e.g., yellow/orange)
- Order volume/price labels always visible (not just on hover)
- BUY/SELL text labels (e.g., "BUY 0.05@4607.33")

**MT5 Format**:
```
─────────── BUY 0.05 @ 4607.33 ───────────  (Blue line)
─ ─ ─ ─ ─ ─ SL @ 4600.00 ─ ─ ─ ─ ─ ─ ─ ─  (Red dashed)
─ ─ ─ ─ ─ ─ TP @ 4650.00 ─ ─ ─ ─ ─ ─ ─ ─  (Green dashed)
```

### 5. Price Scale Enhancement
**Current**: Static price labels on right side.

**MT5 Feature**: Floating current price line that tracks bid/ask dynamically.

**Implementation**:
```typescript
// Add price line for current bid
series.createPriceLine({
  price: currentPrice.bid,
  color: '#ef4444',
  lineWidth: 2,
  lineStyle: 0, // Solid
  axisLabelVisible: true,
  title: 'BID'
});

// Add price line for current ask
series.createPriceLine({
  price: currentPrice.ask,
  color: '#14b8a6',
  lineWidth: 2,
  lineStyle: 0,
  axisLabelVisible: true,
  title: 'ASK'
});
```

### 6. Missing Timeframes
**Current**: M1, M5, M15, H1, H4, D1 (lines 467-471)

**Missing**: W1 (weekly), MN (monthly)

**Backend Support**: `ohlc_cache.go` only defines M1-D1. Need to add W1, MN.

---

## Implementation Plan

### Phase 1: Volume Histogram (Priority: CRITICAL)
**File**: `clients/desktop/src/components/TradingChart.tsx`

1. Add volume series state:
```typescript
const volumeSeriesRef = useRef<ISeriesApi<'Histogram'> | null>(null);
```

2. Create volume series after main series (after line 153):
```typescript
const volumeSeries = chart.addHistogramSeries({
  color: '#06b6d4',
  priceFormat: { type: 'volume' },
  priceScaleId: '',
  scaleMargins: { top: 0.8, bottom: 0 }
});
volumeSeriesRef.current = volumeSeries;
```

3. Update volume data when OHLC changes (in `fetchHistory` effect):
```typescript
if (data.length > 0 && volumeSeriesRef.current) {
  const volumeData = data.map(d => ({
    time: d.time,
    value: d.volume || 0,
    color: d.close >= d.open ? '#14b8a680' : '#ef444480' // Semi-transparent
  }));
  volumeSeriesRef.current.setData(volumeData);
}
```

### Phase 2: Fix Candlestick Colors (Priority: HIGH)
**File**: `clients/desktop/src/components/TradingChart.tsx`

**Change lines 130-132**:
```diff
- upColor: '#10b981', downColor: '#ef4444',
- borderUpColor: '#10b981', borderDownColor: '#ef4444',
- wickUpColor: '#10b981', wickDownColor: '#ef4444',
+ upColor: '#14b8a6', downColor: '#ef4444',
+ borderUpColor: '#14b8a6', borderDownColor: '#ef4444',
+ wickUpColor: '#14b8a6', wickDownColor: '#ef4444',
```

### Phase 3: Customize Grid System (Priority: MEDIUM)
**File**: `clients/desktop/src/components/TradingChart.tsx`

**Change lines 64-67**:
```diff
grid: {
-  vertLines: { color: '#27272a' },
-  horzLines: { color: '#27272a' },
+  vertLines: { color: '#27272a', style: 1, visible: true },
+  horzLines: { color: '#27272a', style: 1, visible: true },
}
```

Add grid toggle:
```typescript
const [gridVisible, setGridVisible] = useState(true);

// In chart options:
grid: {
  vertLines: { color: '#27272a', style: 1, visible: gridVisible },
  horzLines: { color: '#27272a', style: 1, visible: gridVisible }
}
```

### Phase 4: Enhanced Trade Levels (Priority: MEDIUM)
**File**: `clients/desktop/src/components/TradingChart.tsx`

1. Modify `PositionOverlay` component (line 361):
```typescript
// Show label always (not just on hover)
<div className="bg-blue-500 text-white text-[10px] px-1 rounded-sm -mt-5">
  <span>{pos.side} {pos.volume} @ {pos.openPrice.toFixed(5)}</span>
  <X size={12} className="cursor-pointer hover:text-red-300" onClick={onClose} />
</div>
```

2. Add pending order support:
```typescript
interface PendingOrder {
  id: number;
  type: 'BUY_LIMIT' | 'SELL_LIMIT' | 'BUY_STOP' | 'SELL_STOP';
  price: number;
  volume: number;
  sl?: number;
  tp?: number;
}

// Render pending orders with different color
const pendingOrderColor = '#f59e0b'; // Orange/yellow
```

### Phase 5: Current Price Line (Priority: MEDIUM)
**File**: `clients/desktop/src/components/TradingChart.tsx`

Add price lines after series creation:
```typescript
const bidLineRef = useRef<IPriceLine | null>(null);
const askLineRef = useRef<IPriceLine | null>(null);

// Update when currentPrice changes
useEffect(() => {
  if (!seriesRef.current || !currentPrice) return;

  // Remove old lines
  if (bidLineRef.current) seriesRef.current.removePriceLine(bidLineRef.current);
  if (askLineRef.current) seriesRef.current.removePriceLine(askLineRef.current);

  // Create new lines
  bidLineRef.current = seriesRef.current.createPriceLine({
    price: currentPrice.bid,
    color: '#ef4444',
    lineWidth: 1,
    lineStyle: 2, // Dashed
    axisLabelVisible: true,
    title: `${currentPrice.bid.toFixed(5)}`
  });

  askLineRef.current = seriesRef.current.createPriceLine({
    price: currentPrice.ask,
    color: '#14b8a6',
    lineWidth: 1,
    lineStyle: 2,
    axisLabelVisible: true,
    title: `${currentPrice.ask.toFixed(5)}`
  });
}, [currentPrice]);
```

### Phase 6: Add Missing Timeframes (Priority: LOW)
**Files**:
- `backend/tickstore/ohlc_cache.go`
- `backend/tickstore/service.go`
- `clients/desktop/src/components/TradingChart.tsx`

**Backend** (`ohlc_cache.go` after line 22):
```go
const (
  // ... existing
  TF_W1  Timeframe = "W1"  // Weekly
  TF_MN  Timeframe = "MN"  // Monthly
)

func TimeframeSeconds(tf Timeframe) int64 {
  // ... existing cases
  case TF_W1:
    return 604800  // 7 days
  case TF_MN:
    return 2592000 // 30 days (approximation)
}
```

**Frontend** (`TradingChart.tsx` line 467):
```typescript
const timeframes: { value: Timeframe; label: string }[] = [
  { value: '1m', label: 'M1' },
  { value: '5m', label: 'M5' },
  { value: '15m', label: 'M15' },
  { value: '1h', label: 'H1' },
  { value: '4h', label: 'H4' },
  { value: '1d', label: 'D1' },
  { value: '1w', label: 'W1' },  // NEW
  { value: '1M', label: 'MN' }   // NEW
];

// Update Timeframe type
export type Timeframe = '1m' | '5m' | '15m' | '1h' | '4h' | '1d' | '1w' | '1M';
```

---

## Testing Checklist

### Volume Histogram
- [ ] Volume bars render at bottom 20% of chart
- [ ] Volume colors match candle colors (teal/red)
- [ ] Volume scales correctly with max volume
- [ ] Volume updates in real-time with new ticks
- [ ] Volume persists across timeframe changes

### Candlestick Colors
- [ ] Bullish candles are teal (#14b8a6)
- [ ] Bearish candles are red (#ef4444)
- [ ] Colors consistent across all chart types
- [ ] Colors match MT5 screenshot

### Grid System
- [ ] Horizontal lines are dotted
- [ ] Vertical lines are dotted
- [ ] Grid toggle works
- [ ] Grid spacing reasonable

### Trade Levels
- [ ] Entry lines show "BUY/SELL volume @ price"
- [ ] Labels always visible (not just hover)
- [ ] SL/TP lines are dashed
- [ ] Pending orders render with different color
- [ ] Drag-to-modify still works

### Price Scale
- [ ] Current bid line tracks in real-time
- [ ] Current ask line tracks in real-time
- [ ] Price labels visible on right axis
- [ ] Auto-scale works with visible candles

### Timeframes
- [ ] M1, M5, M15, M30, H1, H4, D1 all work
- [ ] W1 (weekly) renders correctly
- [ ] MN (monthly) renders correctly
- [ ] Timeframe switching smooth
- [ ] Data loads for all timeframes

---

## Performance Considerations

### Volume Data
- Volume data adds ~20% overhead to OHLC payload
- Current limit=500 bars → 500 volume points
- Minimal impact (volume is single integer per bar)

### Price Lines
- Bid/Ask lines recreated on every price update
- Use refs to avoid memory leaks
- Remove old lines before creating new ones

### Overlay Complexity
- Position overlays update at 10fps (line 253)
- Consider reducing to 5fps if volume causes lag
- Use React.memo for PositionOverlay component

---

## Dependencies

### NPM Packages
- `lightweight-charts@5.1.0` ✅ Already installed
- No additional packages required

### Backend Changes
- Add W1, MN timeframe support (optional)
- Volume field already exists in OHLC struct

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Volume series performance | Low | Medium | Limit bars to 500, use efficient updates |
| Color change breaks themes | Low | Low | Only affects chart, other components unaffected |
| Price lines flicker | Medium | Low | Debounce updates, use refs properly |
| Grid style not supported | Very Low | Low | lightweight-charts supports dotted lines |
| Timeframe backend changes | Low | Medium | Test with existing timeframes first |

---

## Success Criteria

1. ✅ Volume histogram renders with cyan bars at bottom of chart
2. ✅ Candlestick colors match MT5 (teal/red)
3. ✅ Grid lines are dotted (horizontal and vertical)
4. ✅ Trade levels show "BUY/SELL volume @ price" labels
5. ✅ Current bid/ask price lines track in real-time
6. ✅ All timeframes M1-MN supported
7. ✅ No performance degradation (<100ms render time)
8. ✅ Chart matches MT5 screenshot visually

---

## Files Modified Summary

### Frontend
```
clients/desktop/src/components/TradingChart.tsx
  - Add volume histogram series
  - Fix candlestick colors
  - Customize grid dotted lines
  - Enhance trade level labels
  - Add bid/ask price lines
  - Add W1, MN timeframe support
```

### Backend (Optional)
```
backend/tickstore/ohlc_cache.go
  - Add TF_W1, TF_MN constants
  - Update TimeframeSeconds()

backend/tickstore/service.go
  - Add W1, MN to default timeframes
```

---

## Next Steps

1. **Immediate**: Fix candlestick colors (5 min fix)
2. **High Priority**: Implement volume histogram (1-2 hours)
3. **Medium Priority**: Enhance trade level labels (1 hour)
4. **Medium Priority**: Add bid/ask price lines (1 hour)
5. **Low Priority**: Add W1, MN timeframes (30 min backend + 15 min frontend)

---

## Memory Storage

All findings stored in namespace: `mt5-parity-charting`

**Keys**:
- `architecture-analysis`: Current implementation overview
- `color-scheme`: MT5 color requirements
- `api-endpoints`: OHLC API details
- `missing-features`: List of gaps
- `implementation-plan`: This document

---

## Conclusion

The charting system is **70% MT5-compliant**. Main gaps are:
1. Volume histogram rendering (CRITICAL)
2. Color accuracy (HIGH)
3. Trade level enhancements (MEDIUM)
4. Price line tracking (MEDIUM)

**Estimated total effort**: 6-8 hours for full MT5 parity.

**Recommended approach**: Implement in phases, starting with volume histogram and colors (highest visual impact).
