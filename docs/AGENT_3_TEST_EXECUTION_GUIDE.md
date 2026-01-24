# Agent 3 - Test Execution Guide

## Quick Start

### Run Automated Tests

```bash
cd clients/desktop
npm test src/test/buildOHLCFromTicks.test.ts
```

### Manual Browser Testing

1. Start backend server:
```bash
cd backend/cmd/server
go run main.go
```

2. Start frontend:
```bash
cd clients/desktop
npm run dev
```

3. Open browser console (F12) and run:
```javascript
// Test API endpoint
const response = await fetch('http://localhost:7999/api/history/ticks?symbol=EURUSD&date=2026-01-20&limit=10');
const data = await response.json();
console.log('API Response:', data);

// Verify tick format
console.log('First tick:', data.ticks[0]);
console.log('Timestamp (ms):', data.ticks[0].timestamp);
```

## Test Coverage Summary

### Automated Tests (buildOHLCFromTicks.test.ts)

| Category | Tests | Status |
|----------|-------|--------|
| Edge Cases | 3 | Ready |
| M1 Timeframe | 2 | Ready |
| M5 Timeframe | 2 | Ready |
| M15 Timeframe | 1 | Ready |
| Time Boundaries | 2 | Ready |
| OHLC Calculations | 5 | Ready |
| Sorting | 2 | Ready |
| Performance | 1 | Ready |
| Time Conversion | 3 | Ready |
| **TOTAL** | **21** | **Ready** |

### Expected Results

All tests should PASS except for potential issues with invalid input data (which currently has no validation).

## Data Flow Verification Checklist

- [ ] Backend server running on port 7999
- [ ] API endpoint `/api/history/ticks` accessible
- [ ] API returns tick data with millisecond timestamps
- [ ] Frontend fetches tick data successfully
- [ ] `buildOHLCFromTicks` converts ticks to candles
- [ ] Candles display correctly on chart
- [ ] Time buckets align to timeframe boundaries
- [ ] OHLC values calculated correctly
- [ ] Volume counts ticks per candle

## Known Issues

### 1. No Input Validation
**Severity:** Medium
**Impact:** Invalid timestamps create NaN candles
**Location:** `TradingChart.tsx` line 746
**Fix:** Add validation before processing

### 2. Type Coercion Without Checks
**Severity:** Low
**Impact:** Runtime type errors possible
**Location:** `TradingChart.tsx` line 756
**Fix:** Add runtime type checks

## Performance Benchmarks

Expected performance for 5000 ticks:
- Processing time: < 100ms
- Memory usage: < 5MB
- Candles created: ~83 (for M1)

## Test Results Format

```
PASS  src/test/buildOHLCFromTicks.test.ts
  buildOHLCFromTicks - Edge Cases
    ✓ should return empty array for empty input
    ✓ should return empty array for null input
    ✓ should handle single tick
  buildOHLCFromTicks - M1 Timeframe
    ✓ should aggregate ticks within same minute
    ✓ should create separate candles for different minutes
  ...

Test Suites: 1 passed, 1 total
Tests:       21 passed, 21 total
```

## Debugging Tips

### Backend Issues

```bash
# Check if backend is running
curl http://localhost:7999/api/history/available

# Test specific symbol
curl "http://localhost:7999/api/history/ticks?symbol=EURUSD&date=2026-01-20&limit=5"

# Verify timestamp format
curl "http://localhost:7999/api/history/ticks?symbol=EURUSD&date=2026-01-20&limit=1" | jq '.ticks[0].timestamp'
```

### Frontend Issues

```javascript
// Check if ticks are received
console.log('Ticks length:', data.ticks.length);
console.log('First tick:', data.ticks[0]);

// Test buildOHLC manually
const testTicks = [
  { timestamp: 1737375600000, bid: 1.04532, ask: 1.04535 }
];
const candles = buildOHLCFromTicks(testTicks, '1m');
console.log('Candles:', candles);

// Verify time conversion
const tickTime = 1737375600000;
const seconds = Math.floor(tickTime / 1000);
const candleTime = Math.floor(seconds / 60) * 60;
console.log('Tick (ms):', tickTime);
console.log('Seconds:', seconds);
console.log('Candle time:', candleTime);
```

## Files Created by Agent 3

1. **TEST_HISTORICAL_DATA_FLOW.md**
   - Complete test documentation
   - Data flow trace
   - Test cases with expected results

2. **buildOHLCFromTicks.test.ts**
   - Automated unit tests
   - 21 test cases
   - Performance benchmarks

3. **AGENT_3_TEST_EXECUTION_GUIDE.md**
   - This file
   - Quick reference for running tests

## Next Steps

1. Run automated tests to verify functionality
2. Fix input validation issues
3. Add error handling to TradingChart component
4. Monitor performance with large datasets
5. Document any edge cases discovered during testing

## Success Metrics

- [ ] All 21 automated tests pass
- [ ] No console errors during chart rendering
- [ ] Candles display at correct time positions
- [ ] OHLC values match expected calculations
- [ ] Performance under 100ms for 5000 ticks

---

**Agent 3 Status:** Mission Complete
**Test Suite:** Ready for execution
**Issues Found:** 2 (documented in TEST_HISTORICAL_DATA_FLOW.md)
**Recommendations:** Implement input validation
