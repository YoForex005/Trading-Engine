# Testing Tick-to-Candle Aggregation

## Quick Test Guide

### What to Look For

1. **Console Logs** - Watch for new candle messages:
   ```
   [TradingChart] New candle detected! Previous: 2026-01-20T14:30:00.000Z, New: 2026-01-20T14:31:00.000Z
   ```

2. **Chart Behavior** - For M1 timeframe:
   - New candle should appear every 60 seconds
   - Candle times should align to minute boundaries (e.g., 14:30:00, 14:31:00)
   - Forming candle should update in real-time with ticks
   - Closed candles should remain unchanged

3. **Volume Bars** - Should update when candles close

### Test Scenarios

#### Scenario 1: M1 Timeframe (1 minute candles)
```
Expected: New candle every 60 seconds
Time boundaries: 14:30:00, 14:31:00, 14:32:00, etc.
```

**Test Steps**:
1. Select M1 timeframe
2. Watch for at least 3 candle formations
3. Verify candles align to minute boundaries
4. Check that ticks update the forming candle's high/low/close

#### Scenario 2: M5 Timeframe (5 minute candles)
```
Expected: New candle every 300 seconds (5 minutes)
Time boundaries: 14:30:00, 14:35:00, 14:40:00, etc.
```

**Test Steps**:
1. Select M5 timeframe
2. Wait 5 minutes
3. Verify new candle appears at correct time boundary

#### Scenario 3: Timeframe Switch
```
Expected: Forming candle resets when timeframe changes
```

**Test Steps**:
1. Start with M1
2. Switch to M5
3. Verify new forming candle starts with correct time bucket
4. Switch back to M1
5. Verify M1 candle forms correctly

### Debug Checklist

- [ ] Console shows new candle messages every 60s for M1
- [ ] Candle times align to time boundaries (not random times)
- [ ] Forming candle updates with each tick
- [ ] Closed candles remain unchanged
- [ ] Volume bars update when candles close
- [ ] Historical data loads correctly
- [ ] Bid/Ask price lines visible
- [ ] No errors in browser console

### Common Issues

**Issue**: Candles not forming at regular intervals
- Check: Is WebSocket tick stream running?
- Check: Are ticks arriving frequently enough?
- Check: Is `currentTick` updating in Zustand store?

**Issue**: Multiple candles with same timestamp
- Check: Time-bucket calculation (should use `Math.floor`)
- Check: Comparison logic (`formingCandleRef.current.time !== candleTime`)

**Issue**: Candles forming too frequently
- Check: `getTimeframeSeconds()` returns correct values
- Check: Time is in seconds (not milliseconds)

### Performance Expectations

- **Tick processing**: < 1ms per tick
- **Chart update**: Smooth, no visible lag
- **Memory**: Stable, no leaks with long-running session
- **Volume calculation**: Increments correctly with each tick

### Integration Points

This fix works with:
- **Zustand Store** (`useCurrentTick` hook)
- **WebSocket** (backend tick stream)
- **TradingView Lightweight Charts** (chart rendering)
- **Volume Histogram** (separate series for volume bars)

---

**Test Duration**: 5-10 minutes per timeframe
**Critical Timeframes**: M1, M5 (most commonly used)
**Success Criteria**: New candles form at exact time boundaries
