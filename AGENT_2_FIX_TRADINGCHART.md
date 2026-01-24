# TradingChart.tsx Historical Data Fix

## Bug Identified (Line 228)

**CRITICAL BUG**: Wrong port and wrong endpoint
```typescript
// WRONG:
const res = await fetch(`http://localhost:8080/ohlc?symbol=${symbol}&timeframe=${timeframe}&limit=500`);
```

## Root Cause Analysis

1. **Wrong Port**: `8080` → Backend server runs on port **7999** (confirmed in `backend/cmd/server/main.go` line 1915)
2. **Wrong Endpoint**: `/ohlc` → This endpoint **DOES NOT EXIST** in the backend
3. **Missing Data Conversion**: Backend provides tick data via `/api/history/ticks`, not OHLC candles

## Backend Investigation

### Server Port (backend/cmd/server/main.go)
```go
// Line 1915
port := ":" + cfg.Port  // Port 7999 from config
if err := http.ListenAndServe(port, handler); err != nil {
    log.Fatal(err)
}
```

### Historical Data API (backend/api/history.go)
Available endpoints:
- `GET /api/history/ticks?symbol=XXX&date=YYYY-MM-DD&limit=5000` - Returns tick data
- `GET /api/history/available` - Lists available symbols
- `GET /api/history/symbols` - Symbol metadata

Response format from `/api/history/ticks`:
```json
{
  "symbol": "EURUSD",
  "date": "2026-01-20",
  "ticks": [
    {
      "timestamp": 1737388800000,
      "bid": 1.05123,
      "ask": 1.05125,
      "spread": 0.00002
    }
  ],
  "total": 1234,
  "offset": 0,
  "limit": 5000
}
```

## Complete Fix

### 1. Update Fetch Logic (Line 227-274)

```typescript
// Fetch historical OHLC data when symbol or timeframe changes
useEffect(() => {
    const fetchHistory = async () => {
        if (!seriesRef.current) return;

        try {
            // FIXED: Correct port (7999) and endpoint (/api/history/ticks)
            // Backend server runs on port 7999 (see backend/cmd/server/main.go:1915)
            // Historical tick data API at /api/history/ticks (see backend/api/history.go)
            const dateStr = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
            const res = await fetch(`http://localhost:7999/api/history/ticks?symbol=${symbol}&date=${dateStr}&limit=5000`);

            if (!res.ok) {
                console.warn(`No historical data for ${symbol}: ${res.status} ${res.statusText}`);
                candlesRef.current = [];
                formingCandleRef.current = null;
                seriesRef.current.setData([]);
                return;
            }

            const data = await res.json();

            // Convert tick data to OHLC candles
            if (data.ticks && Array.isArray(data.ticks) && data.ticks.length > 0) {
                const candles = buildOHLCFromTicks(data.ticks, timeframe);

                if (candles.length > 0) {
                    // Store only historical (closed) candles
                    candlesRef.current = candles;

                    // Reset forming candle when new historical data is loaded
                    formingCandleRef.current = null;

                    const formattedData = formatDataForSeries(candles, chartType);
                    seriesRef.current.setData(formattedData);

                    // Set volume data
                    if (volumeSeriesRef.current && candles.length > 0) {
                        const volumeData = candles.map(bar => ({
                            time: bar.time,
                            value: bar.volume || 0,
                            color: bar.close >= bar.open ? 'rgba(6, 182, 212, 0.5)' : 'rgba(239, 68, 68, 0.5)',
                        }));
                        volumeSeriesRef.current.setData(volumeData);
                    }
                } else {
                    console.warn(`No candles built from ${data.ticks.length} ticks for ${symbol}`);
                    candlesRef.current = [];
                    formingCandleRef.current = null;
                    seriesRef.current.setData([]);
                }
            } else {
                console.warn(`No tick data returned for ${symbol}`);
                candlesRef.current = [];
                formingCandleRef.current = null;
                seriesRef.current.setData([]);
                if (volumeSeriesRef.current) {
                    volumeSeriesRef.current.setData([]);
                }
            }
        } catch (err) {
            console.error(`Error fetching historical data for ${symbol}:`, err);
            candlesRef.current = [];
        }
    };

    fetchHistory();
}, [symbol, timeframe, chartType]);
```

### 2. Add Tick-to-OHLC Converter Function (After line 710)

```typescript
/**
 * Converts tick data from backend API to OHLC candles
 * Backend returns: { timestamp: number (unix ms), bid: number, ask: number, spread: number }
 */
function buildOHLCFromTicks(ticks: any[], timeframe: Timeframe): OHLC[] {
    if (!ticks || ticks.length === 0) return [];

    const tfSeconds = getTimeframeSeconds(timeframe);
    const candleMap = new Map<number, OHLC>();

    // Group ticks into candles by time bucket
    for (const tick of ticks) {
        const price = (tick.bid + tick.ask) / 2; // Mid price
        const timestamp = Math.floor(tick.timestamp / 1000); // Convert ms to seconds
        const candleTime = (Math.floor(timestamp / tfSeconds) * tfSeconds) as Time;

        if (!candleMap.has(candleTime as number)) {
            // Create new candle
            candleMap.set(candleTime as number, {
                time: candleTime,
                open: price,
                high: price,
                low: price,
                close: price,
                volume: 1,
            });
        } else {
            // Update existing candle
            const candle = candleMap.get(candleTime as number)!;
            candle.high = Math.max(candle.high, price);
            candle.low = Math.min(candle.low, price);
            candle.close = price;
            candle.volume = (candle.volume || 0) + 1; // Tick count as volume
        }
    }

    // Sort candles by time
    return Array.from(candleMap.values()).sort((a, b) => (a.time as number) - (b.time as number));
}
```

## Success Criteria

After applying this fix:
1. Chart should fetch historical ticks from correct endpoint (`http://localhost:7999/api/history/ticks`)
2. Ticks should be converted to OHLC candles matching the selected timeframe
3. Chart should display 500+ historical candles on symbol open
4. No more fetch errors in console
5. Volume bars should display (based on tick count)

## Testing

```bash
# 1. Start backend
cd backend/cmd/server
go run main.go

# 2. Verify endpoint works
curl "http://localhost:7999/api/history/ticks?symbol=EURUSD&date=2026-01-20&limit=100"

# 3. Start frontend
cd clients/desktop
npm run dev

# 4. Open EURUSD chart and verify candles load
```

## Related Files
- `backend/cmd/server/main.go` - Server configuration (port 7999)
- `backend/api/history.go` - Historical data API implementation
- `clients/desktop/src/components/TradingChart.tsx` - Chart component (contains bug)
