# Quick Start: Historical Data Simulation

## What This Does

Instead of showing random prices while waiting for YOFX connection, the backend now uses **real OANDA historical tick data** to simulate realistic market prices.

## How to Test

### 1. Start the Server

```bash
cd backend/cmd/server
./server.exe
```

### 2. Wait 30 Seconds

The server waits 30 seconds for real YOFX market data. If none arrives, you'll see:

```
[SIM-MD] No real market data after 30s - starting OANDA historical data simulation
[SIM-MD] Using real OANDA tick data with small variations for realistic prices
[SIM-MD] Loading historical data from: D:\Tading engine\Trading-Engine\backend
[HISTORICAL] Loaded 4575 ticks for EURUSD from 2026-01-19.json (avg spread: 0.00015, pip: 0.00010)
[HISTORICAL] Loaded 4575 ticks for GBPUSD from 2026-01-19.json (avg spread: 0.00015, pip: 0.00010)
...
[SIM-MD] Successfully loaded historical data for 16 symbols
```

### 3. Check the Frontend

Open your trading frontend and you should see:
- ✅ Realistic price levels (e.g., EURUSD around 1.17, not random numbers)
- ✅ Prices updating smoothly every 500ms
- ✅ LP label shows "OANDA-HISTORICAL"

### 4. Verify via API

```bash
# Get current ticks
curl http://localhost:7999/admin/fix/ticks

# Should show:
# - "symbolCount": 16
# - "lp": "OANDA-HISTORICAL"
# - Realistic bid/ask prices
```

## What You'll See

### Before (Old Random Simulation)
```json
{
  "EURUSD": {
    "bid": 1.0850,
    "ask": 1.08515,
    "lp": "SIMULATED"
  }
}
```
- Fixed starting price
- Random walk
- Generic label

### After (New Historical Data)
```json
{
  "EURUSD": {
    "bid": 1.17191,
    "ask": 1.17207,
    "spread": 0.00016,
    "lp": "OANDA-HISTORICAL"
  }
}
```
- Real OANDA price
- Real spread
- Clear data source

## When YOFX Connects

The simulation automatically stops when real YOFX data arrives:

```
[SIM-MD] Real market data now available - stopping historical simulation
```

Then you'll see:
```json
{
  "EURUSD": {
    "bid": 1.0855,
    "ask": 1.0856,
    "lp": "YOFX"
  }
}
```

## Supported Symbols

16 major forex pairs:
- **4-decimal**: EURUSD, GBPUSD, AUDUSD, USDCAD, USDCHF, NZDUSD, EURGBP, AUDCAD, AUDCHF, AUDNZD, AUDSGD
- **2-decimal**: USDJPY, EURJPY, GBPJPY, AUDJPY, AUDHKD

## Data Source

Historical data location: `backend/data/ticks/{SYMBOL}/`

Example: `backend/data/ticks/EURUSD/2026-01-19.json`

## Troubleshooting

### No Historical Data Loading?

Check the logs:
```
[SIM-MD] Failed to load historical data for EURUSD: ...
```

Verify files exist:
```bash
ls backend/data/ticks/EURUSD/
# Should show: 2026-01-03.json, 2026-01-19.json, etc.
```

### Server Won't Start?

Check port 7999 is not in use:
```bash
# Windows
netstat -ano | findstr :7999

# If in use, kill the process
taskkill /F /PID <PID>
```

### Want to Force Historical Mode?

The simulation only starts if NO real market data arrives in 30 seconds. To test:
1. Make sure YOFX is disconnected
2. Start server
3. Wait 30 seconds
4. Historical simulation kicks in

## Quick Verification Script

Run the verification script:
```bash
cd backend/cmd/server
bash verify_historical_data.sh
```

Expected output:
```
✓ Server is running
✓ Symbols loaded: 16
✓ LP Label: OANDA-HISTORICAL
Sample Prices:
EURUSD: Bid=1.17191 Ask=1.17207
GBPUSD: Bid=1.25601 Ask=1.25616
...
✓ Verification Complete!
```

## Benefits

1. **Better Frontend Testing**: See realistic prices during development
2. **Realistic Spreads**: Match actual market conditions
3. **Smooth Transition**: Seamlessly switch to live YOFX data when available
4. **No Changes Needed**: Frontend code works the same way
5. **Clear Labeling**: Always know if you're seeing historical or live data

## Next Steps

When YOFX market data is working:
- Historical simulation automatically stops
- Live YOFX ticks take over
- LP label changes from "OANDA-HISTORICAL" to "YOFX"
- Everything else works identically

No code changes needed!
