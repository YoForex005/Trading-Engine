# Historical Data Simulation Enhancement

## Overview
Enhanced the market data simulation fallback to use **real OANDA historical tick data** instead of random price generation, providing realistic price levels while waiting for YOFX market data connection.

## Changes Made

### 1. Added Historical Data Loading (`backend/cmd/server/main.go`)

#### New Data Structures
```go
type HistoricalTick struct {
    BrokerID  string  `json:"broker_id"`
    Symbol    string  `json:"symbol"`
    Bid       float64 `json:"bid"`
    Ask       float64 `json:"ask"`
    Spread    float64 `json:"spread"`
    Timestamp string  `json:"timestamp"`
    LP        string  `json:"lp"`
}

type HistoricalDataCache struct {
    Ticks      []HistoricalTick
    LastIndex  int
    Symbol     string
    AvgSpread  float64
    PipSize    float64
    mu         sync.RWMutex
}
```

#### Key Functions

**`loadHistoricalTickData(symbol, dataDir)`**
- Loads OANDA historical tick data from `backend/data/ticks/{SYMBOL}/`
- Automatically finds the most recent data file
- Calculates average spread from real OANDA data
- Determines correct pip size based on symbol (0.0001 for most, 0.01 for JPY/HKD pairs)
- Caches loaded data for performance

**`getNextHistoricalTick()`**
- Cycles through historical tick data
- Uses real historical price as baseline
- Adds small random variation (-2 to +2 pips) to simulate market movement
- Preserves realistic spreads from OANDA data
- Labels ticks as `LP: "OANDA-HISTORICAL"`

### 2. Modified Simulation Fallback

**Before:**
- Used hardcoded base prices
- Generated random walk prices
- Fixed 1.5 pip spread
- Labeled as `"SIMULATED"`

**After:**
- Loads real OANDA historical data from files
- Uses actual historical prices as baseline
- Uses actual average spreads from OANDA data
- Adds small variations for live market feel
- Labeled as `"OANDA-HISTORICAL"`

### 3. Smart Data Directory Detection

The code automatically detects the working directory and finds the historical data:
```go
workDir, _ := os.Getwd()
// Handles: backend/cmd/server, backend/, or project root
```

## Results

### Successfully Tested
✅ **16 symbols loading historical data:**
- EURUSD, GBPUSD, USDJPY, AUDUSD
- USDCAD, USDCHF, NZDUSD, EURGBP
- EURJPY, GBPJPY, AUDJPY, AUDCAD
- AUDCHF, AUDNZD, AUDSGD, AUDHKD

✅ **Each symbol loads 4,575-4,576 historical ticks**

✅ **Correct LP labeling:** All ticks marked as `"OANDA-HISTORICAL"`

✅ **Realistic prices:** Example EURUSD @ 1.0654 (actual OANDA range)

✅ **Realistic spreads:** Average 0.00015 (1.5 pips) from actual OANDA data

### Sample Output
```
[SIM-MD] No real market data after 30s - starting OANDA historical data simulation
[SIM-MD] Using real OANDA tick data with small variations for realistic prices
[SIM-MD] Loading historical data from: D:\Tading engine\Trading-Engine\backend
[HISTORICAL] Loaded 4575 ticks for EURUSD from 2026-01-19.json (avg spread: 0.00015, pip: 0.00010)
[HISTORICAL] Loaded 4575 ticks for GBPUSD from 2026-01-19.json (avg spread: 0.00015, pip: 0.00010)
...
[SIM-MD] Successfully loaded historical data for 16 symbols
```

### API Response
```json
{
  "EURUSD": {
    "type": "tick",
    "symbol": "EURUSD",
    "bid": 1.0654,
    "ask": 1.0655,
    "spread": 0.00015,
    "timestamp": 1768823641,
    "lp": "OANDA-HISTORICAL"
  }
}
```

## Benefits

1. **Realistic Price Levels**: Uses actual OANDA market data instead of random numbers
2. **Realistic Spreads**: Preserves actual bid/ask spreads from historical data
3. **Better Testing**: Frontend developers see real market conditions
4. **Smooth Transition**: When YOFX connects, seamlessly switches to live data
5. **Clear Labeling**: `"OANDA-HISTORICAL"` vs `"YOFX"` makes data source obvious

## Next Steps

When YOFX market data connection is established:
1. The simulation automatically stops (already implemented)
2. Live YOFX ticks will be labeled as `LP: "YOFX"`
3. No code changes needed - automatic fallback behavior

## Technical Notes

- **Thread-safe caching**: Historical data cached per symbol with mutex protection
- **Memory efficient**: Data loaded once and cycled through
- **Auto-detection**: Works regardless of working directory (root, backend, or cmd/server)
- **Graceful fallback**: If historical data missing for a symbol, it's skipped (no crash)
- **Performance**: 500ms tick generation rate (same as before)
