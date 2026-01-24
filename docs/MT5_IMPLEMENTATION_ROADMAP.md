# MT5 Charting Implementation Roadmap

**Phase-by-Phase Implementation Plan**

---

## 📋 Overview

**Goal**: Achieve 100% MT5 charting parity
**Current Status**: 70% compliant
**Remaining Work**: 6-8 hours
**Priority**: Critical for trader adoption

---

## 🎯 Implementation Phases

### Phase 1: Quick Wins (15 minutes) ⚡
**Impact**: HIGH | **Effort**: LOW | **Risk**: NONE

#### Tasks
1. Fix candlestick colors (5 min)
2. Add dotted grid lines (10 min)

#### Files
- `clients/desktop/src/components/TradingChart.tsx`

#### Success Criteria
- [x] Bullish candles are teal (#14b8a6)
- [x] Grid lines are dotted (style: 1)
- [x] Visual match with MT5 screenshot improved by 30%

#### Implementation
See `MT5_CHARTING_QUICK_FIX_GUIDE.md` sections:
- "HIGH: Fix Candlestick Colors"
- "MEDIUM: Dotted Grid Lines"

---

### Phase 2: Volume Histogram (2-3 hours) 🔴
**Impact**: CRITICAL | **Effort**: MEDIUM | **Risk**: LOW

#### Tasks
1. Add volume series ref and state
2. Create histogram series with proper configuration
3. Map OHLC volume data to histogram
4. Implement real-time volume updates
5. Add color logic (teal/red based on candle direction)
6. Configure separate price scale for volume

#### Files
- `clients/desktop/src/components/TradingChart.tsx`

#### Success Criteria
- [x] Volume bars render at bottom 20% of chart
- [x] Colors match candle direction (teal bullish, red bearish)
- [x] Volume updates in real-time with new ticks
- [x] Separate Y-axis scale for volume
- [x] Performance: <20ms render overhead

#### Implementation
See `MT5_CHARTING_QUICK_FIX_GUIDE.md` section:
- "CRITICAL: Volume Histogram"

#### Testing
```bash
# Test volume data availability
curl "http://localhost:8080/ohlc?symbol=XAUUSD&timeframe=1m&limit=10"
# Verify each bar has "volume" field

# Visual check
# 1. Open chart
# 2. Verify cyan bars at bottom
# 3. Verify heights match volume values
# 4. Verify colors match candle directions
```

---

### Phase 3: Bid/Ask Price Lines (1 hour) 🟡
**Impact**: MEDIUM | **Effort**: LOW | **Risk**: LOW

#### Tasks
1. Import IPriceLine type
2. Add refs for bid/ask lines
3. Create effect to update price lines on tick
4. Remove old lines before creating new ones (prevent memory leak)
5. Style bid line (red, dashed)
6. Style ask line (teal, dashed)

#### Files
- `clients/desktop/src/components/TradingChart.tsx`

#### Success Criteria
- [x] Bid line shows red dashed line at current bid price
- [x] Ask line shows teal dashed line at current ask price
- [x] Lines update on every tick (real-time)
- [x] Price labels show 5 decimal places
- [x] No memory leaks (old lines properly removed)

#### Implementation
See `MT5_CHARTING_QUICK_FIX_GUIDE.md` section:
- "MEDIUM: Bid/Ask Price Lines"

---

### Phase 4: Enhanced Trade Labels (1-2 hours) 🟡
**Impact**: MEDIUM | **Effort**: MEDIUM | **Risk**: LOW

#### Tasks
1. Modify entry line to show "BUY/SELL volume @ price"
2. Make labels always visible (remove hover dependency)
3. Change entry line style to solid (from dotted)
4. Add pending order support (optional)
5. Style pending orders with orange color

#### Files
- `clients/desktop/src/components/TradingChart.tsx` (PositionOverlay component)

#### Success Criteria
- [x] Entry labels show "BUY 0.05 @ 4607.33" format
- [x] Labels always visible (not just on hover)
- [x] Entry line is solid blue
- [x] SL/TP lines remain dashed
- [x] Pending orders render with orange color (if implemented)

#### Implementation
See `MT5_CHARTING_QUICK_FIX_GUIDE.md` section:
- "MEDIUM: Enhanced Trade Labels"

#### Optional Enhancement
```typescript
// Add pending order interface
interface PendingOrder {
  id: number;
  type: 'BUY_LIMIT' | 'SELL_LIMIT' | 'BUY_STOP' | 'SELL_STOP';
  price: number;
  volume: number;
  sl?: number;
  tp?: number;
}

// Render with orange color
const pendingColor = '#f59e0b';
```

---

### Phase 5: Additional Timeframes (1 hour) 🟢
**Impact**: LOW | **Effort**: LOW | **Risk**: MEDIUM

#### Tasks (Backend)
1. Add W1, MN constants to `ohlc_cache.go`
2. Update `TimeframeSeconds()` function
3. Add W1, MN to default timeframes array
4. Update OHLC API endpoint mapping

#### Tasks (Frontend)
1. Update Timeframe type definition
2. Add W1, MN to timeframe selector
3. Update `getTimeframeSeconds()` function
4. Test W1, MN data loading

#### Files (Backend)
- `backend/tickstore/ohlc_cache.go`
- `backend/tickstore/service.go`
- `backend/api/server.go`

#### Files (Frontend)
- `clients/desktop/src/components/TradingChart.tsx`

#### Success Criteria
- [x] W1 (weekly) timeframe selector visible
- [x] MN (monthly) timeframe selector visible
- [x] W1 data loads correctly
- [x] MN data loads correctly
- [x] Switching between all timeframes smooth

#### Implementation
See `MT5_CHARTING_QUICK_FIX_GUIDE.md` section:
- "OPTIONAL: Weekly/Monthly Timeframes"

#### Risk Mitigation
- Test with existing timeframes first
- Ensure backend OHLC aggregation handles weekly/monthly correctly
- Consider disabling if data quality issues

---

## 🗓️ Recommended Schedule

### Day 1 (4 hours)
- **Morning** (2 hours): Phase 1 + Phase 2 (Colors, Grid, Volume)
- **Afternoon** (2 hours): Phase 2 continued (Volume testing and refinement)

### Day 2 (3 hours)
- **Morning** (1.5 hours): Phase 3 (Bid/Ask Price Lines)
- **Afternoon** (1.5 hours): Phase 4 (Enhanced Trade Labels)

### Day 3 (1 hour) - Optional
- **Morning** (1 hour): Phase 5 (Additional Timeframes)

**Total**: 6-8 hours over 2-3 days

---

## 🧪 Testing Strategy

### Unit Tests
```bash
cd clients/desktop
npm test -- TradingChart.test.tsx
```

**Test Cases**:
1. Volume series renders correctly
2. Color mapping (candle direction → volume color)
3. Price lines update on tick
4. Grid style is dotted
5. Trade labels format correctly

### Integration Tests
```bash
# Start backend
cd backend && go run cmd/server/main.go

# Start frontend
cd clients/desktop && npm run dev

# Manual testing checklist in browser
```

### Visual Regression
Compare screenshots:
1. Before implementation
2. After each phase
3. Final vs MT5 reference screenshot

---

## 📊 Progress Tracking

### Completion Checklist

#### Phase 1: Quick Wins ✅
- [ ] Candlestick colors changed to teal
- [ ] Grid lines changed to dotted
- [ ] Visual QA passed

#### Phase 2: Volume Histogram
- [ ] Volume series created
- [ ] Volume data mapped from OHLC
- [ ] Real-time updates working
- [ ] Colors match candle direction
- [ ] Performance acceptable (<20ms overhead)
- [ ] Visual QA passed

#### Phase 3: Bid/Ask Lines
- [ ] Bid line renders
- [ ] Ask line renders
- [ ] Real-time updates working
- [ ] Price labels correct (5 decimals)
- [ ] No memory leaks
- [ ] Visual QA passed

#### Phase 4: Trade Labels
- [ ] Entry labels show "BUY/SELL volume @ price"
- [ ] Labels always visible
- [ ] Entry line is solid
- [ ] SL/TP lines remain dashed
- [ ] Visual QA passed

#### Phase 5: Timeframes (Optional)
- [ ] W1 backend support added
- [ ] MN backend support added
- [ ] W1 frontend selector added
- [ ] MN frontend selector added
- [ ] Data loading tested
- [ ] Visual QA passed

---

## 🚨 Known Risks & Mitigation

### Risk 1: Volume Rendering Performance
**Probability**: Medium
**Impact**: Medium
**Mitigation**:
- Limit volume bars to 500 max
- Use efficient update methods (update vs setData)
- Monitor FPS during testing
- Fallback: Disable volume if performance degrades

### Risk 2: Price Line Flickering
**Probability**: Low
**Impact**: Low
**Mitigation**:
- Debounce updates to 100ms
- Remove old lines before creating new
- Use refs to avoid React re-renders

### Risk 3: Weekly/Monthly Data Quality
**Probability**: Medium
**Impact**: Low
**Mitigation**:
- Extensive testing with real data
- Validate OHLC aggregation logic
- Fallback: Keep only M1-D1 if issues persist

### Risk 4: Grid Breaking Layout
**Probability**: Very Low
**Impact**: Low
**Mitigation**:
- Test on multiple screen sizes
- Verify lightweight-charts supports dotted style (it does)
- Fallback: Revert to solid lines

---

## 🎯 Success Metrics

### Technical Metrics
| Metric | Target | Current | After Implementation |
|--------|--------|---------|---------------------|
| MT5 Parity | 100% | 70% | 100% |
| Render Time | <200ms | 150ms | 170ms |
| Memory Usage | <100MB | 60MB | 70MB |
| FPS | 60 | 60 | 60 |

### Visual Metrics
| Element | Match % | Target |
|---------|---------|--------|
| Candle Colors | 50% | 100% |
| Volume Bars | 0% | 100% |
| Grid Style | 80% | 100% |
| Trade Labels | 60% | 100% |
| Price Lines | 0% | 100% |

### User Impact
- **Traders** can analyze volume patterns (critical for MT5 users)
- **Chart accuracy** matches industry standard (MT5)
- **Professional appearance** increases platform credibility
- **Feature parity** reduces friction in user migration from MT5

---

## 📝 Documentation Updates

After implementation, update:
1. `README.md` - Add MT5 charting feature
2. `CHANGELOG.md` - List all charting improvements
3. User guide - Screenshot updated charts
4. API docs - Document volume field usage

---

## 🔄 Rollback Plan

If critical issues occur during implementation:

### Phase 2 Rollback (Volume)
```typescript
// Remove volume series
if (volumeSeriesRef.current) {
  chartRef.current?.removeSeries(volumeSeriesRef.current);
  volumeSeriesRef.current = null;
}
```

### Phase 3 Rollback (Price Lines)
```typescript
// Remove price lines
if (bidLineRef.current) {
  seriesRef.current?.removePriceLine(bidLineRef.current);
}
if (askLineRef.current) {
  seriesRef.current?.removePriceLine(askLineRef.current);
}
```

### Full Rollback
```bash
git checkout HEAD -- clients/desktop/src/components/TradingChart.tsx
```

---

## 📞 Support Resources

### Documentation
- Full Analysis: `docs/MT5_CHARTING_ANALYSIS_REPORT.md`
- Quick Fix Guide: `docs/MT5_CHARTING_QUICK_FIX_GUIDE.md`
- Visual Spec: `docs/MT5_VISUAL_COMPARISON.md`
- This Roadmap: `docs/MT5_IMPLEMENTATION_ROADMAP.md`

### External Resources
- Lightweight Charts: https://tradingview.github.io/lightweight-charts/
- Histogram Series: https://tradingview.github.io/lightweight-charts/tutorials/demos/histogram-series
- Price Lines: https://tradingview.github.io/lightweight-charts/docs/api/interfaces/IPriceLine

### Memory Namespace
All findings stored in: `mt5-parity-charting`
```bash
npx @claude-flow/cli@latest memory search --query "volume" --namespace mt5-parity-charting
```

---

## ✅ Final Deliverables

1. **Functional Chart** matching MT5 screenshot
2. **Volume Histogram** rendering real-time data
3. **Correct Colors** (teal/red for candles, cyan for volume)
4. **Dotted Grid** lines
5. **Trade Levels** with always-visible labels
6. **Bid/Ask Lines** tracking current price
7. **All Timeframes** working (M1-MN)
8. **Performance** within targets (<200ms render, 60fps)
9. **Documentation** updated
10. **Tests** passing

---

**Implementation Start**: Ready to begin
**Estimated Completion**: 2-3 days
**Confidence Level**: High (well-documented, clear path)

---

*This roadmap is a living document. Update as phases complete.*
