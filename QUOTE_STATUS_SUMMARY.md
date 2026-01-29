# Quote Status Summary

## Quick Answer: SIMULATED 🎭

The WebSocket stream at `ws://localhost:7999/ws` is broadcasting **SIMULATED** market data.

## Evidence (3 Definitive Markers)

### 1. LP Field ❌
```json
{
  "lp": "SIMULATED"  // All quotes show this explicit marker
}
```

### 2. Negative Prices ❌
Several currency pairs have impossible negative prices:
- AUDUSD: -0.07279
- AUDCHF: -0.13179
- NZDUSD: -0.13279

Real forex prices cannot be negative.

### 3. Unrealistic Price Levels ❌
| Pair | Current | Should Be |
|------|---------|-----------|
| EURUSD | 0.387 | ~1.085 |
| GBPUSD | 0.567 | ~1.265 |
| USDJPY | 86.71 | ~156.50 |

## Why?

The FIX gateway (YOFX) is not connected. After 30 seconds of no real data, the server automatically starts simulated quotes for UI testing.

## How to Get Real Data

Real quotes will show `"lp": "YOFX"` instead of `"SIMULATED"`.

To enable real data:
1. Configure FIX credentials in `backend/fix/config/sessions.json`
2. Restart server
3. Check `/admin/fix/status` to verify connection
4. Real quotes will have realistic prices and `"lp": "YOFX"`

## Verification Commands

```bash
# Check quote source
curl http://localhost:7999/admin/fix/ticks

# Look for:
# "lp": "SIMULATED"  = Fake data
# "lp": "YOFX"       = Real data
```

---

**Full Analysis:** See `docs/WEBSOCKET_QUOTE_ANALYSIS_REPORT.md`
