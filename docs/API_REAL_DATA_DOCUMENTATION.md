# RTX Trading Engine - API Documentation: Real Data Edition

**Date**: January 30, 2026
**Version**: 2.0 (Real Data)
**Status**: ✅ Production Ready

---

## Overview

This API documentation describes the RTX Trading Engine REST API and WebSocket interface with **100% real YOFX market data**. All data is sourced directly from YOFX FIX feeds with no simulation or mock data.

---

## Table of Contents

1. [Market Data Endpoints](#market-data-endpoints)
2. [WebSocket Market Data API](#websocket-market-data-api)
3. [Real Data Verification](#real-data-verification)
4. [Data Quality Guarantees](#data-quality-guarantees)
5. [Admin Endpoints](#admin-endpoints)

---

## Market Data Endpoints

### GET /api/market/ticks/{symbol}

Get latest market tick for a specific symbol.

**Request**:
```http
GET /api/market/ticks/EURUSD HTTP/1.1
Host: api.example.com
Authorization: Bearer {token}
```

**Response** (200 OK):
```json
{
  "symbol": "EURUSD",
  "bid": 1.1045156847,
  "ask": 1.1046156847,
  "spread": 0.0001,
  "high24h": 1.1051234567,
  "low24h": 1.1032456789,
  "dailyChange": 0.0234,
  "timestamp": 1769772266771,
  "lp": "YOFX",
  "source": "real",
  "verified": true
}
```

**Response Fields**:
| Field | Type | Description |
|-------|------|-------------|
| `symbol` | string | Currency pair (e.g., EURUSD) |
| `bid` | float64 | Current bid price (real YOFX) |
| `ask` | float64 | Current ask price (real YOFX) |
| `spread` | float64 | Bid-ask spread in pips |
| `high24h` | float64 | 24-hour high (real YOFX) |
| `low24h` | float64 | 24-hour low (real YOFX) |
| `dailyChange` | float64 | Daily percentage change |
| `timestamp` | int64 | Unix milliseconds timestamp |
| `lp` | string | Liquidity provider ("YOFX" only) |
| `source` | string | "real" (no simulation) |
| `verified` | boolean | Data verified as real |

**Data Guarantee**:
- ✅ All prices are from YOFX FIX feed
- ✅ No simulation or mock data
- ✅ High/low tracked from real data
- ✅ LP tag is always "YOFX"

---

### GET /api/market/ohlc/{symbol}/{timeframe}

Get OHLC (Open, High, Low, Close) data for a symbol.

**Request**:
```http
GET /api/market/ohlc/EURUSD/M5 HTTP/1.1
Host: api.example.com
Authorization: Bearer {token}
```

**Parameters**:
- `symbol`: Currency pair (EURUSD, GBPUSD, etc.)
- `timeframe`: M1, M5, M15, H1, H4, D1

**Response** (200 OK):
```json
[
  {
    "timestamp": 1769771700000,
    "open": 1.1044,
    "high": 1.1048,
    "low": 1.1040,
    "close": 1.1046,
    "volume": 12500,
    "lp": "YOFX",
    "verified": true
  },
  {
    "timestamp": 1769772000000,
    "open": 1.1046,
    "high": 1.1051,
    "low": 1.1045,
    "close": 1.1049,
    "volume": 15300,
    "lp": "YOFX",
    "verified": true
  }
]
```

**Data Guarantee**:
- ✅ Candles built from real YOFX ticks
- ✅ No synthetic data generation
- ✅ High/low verified from tick data
- ✅ All data from YOFX source

---

### GET /api/symbols/all

Get list of all available symbols.

**Request**:
```http
GET /api/symbols/all HTTP/1.1
Host: api.example.com
Authorization: Bearer {token}
```

**Response** (200 OK):
```json
[
  {
    "symbol": "EURUSD",
    "description": "EUR/USD",
    "available": true,
    "lp": "YOFX",
    "baseAsset": "EUR",
    "quoteAsset": "USD",
    "minVolume": 1000,
    "maxVolume": 1000000,
    "pipValue": 0.0001
  },
  {
    "symbol": "GBPUSD",
    "description": "GBP/USD",
    "available": true,
    "lp": "YOFX",
    "baseAsset": "GBP",
    "quoteAsset": "USD",
    "minVolume": 1000,
    "maxVolume": 1000000,
    "pipValue": 0.0001
  }
]
```

**Guarantee**:
- ✅ Only YOFX-supported symbols shown
- ✅ All available through real FIX feed
- ✅ No simulation symbols

---

### GET /api/diagnostics/market-data

Real-time market data diagnostics and verification.

**Request**:
```http
GET /api/diagnostics/market-data HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
```

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2026-01-30T12:34:56Z",
  "fixSessions": {
    "YOFX2": {
      "status": "LOGGED_IN",
      "inboundSeqNum": 123456,
      "outboundSeqNum": 67890,
      "lastMessageTime": "2026-01-30T12:34:55.234Z",
      "messagesPerSecond": 245
    },
    "YOFX1": {
      "status": "DISCONNECTED",
      "inboundSeqNum": 0,
      "outboundSeqNum": 0,
      "lastMessageTime": null,
      "messagesPerSecond": 0
    }
  },
  "subscribedSymbols": [
    "EURUSD",
    "GBPUSD",
    "USDJPY",
    "EURGBP",
    "AUDUSD",
    "USDCAD",
    "NZDUSD",
    "USDCHF",
    "EURJPY",
    "GBPJPY"
  ],
  "subscribedCount": 29,
  "ticksPerSymbol": {
    "EURUSD": 523641,
    "GBPUSD": 512345,
    "USDJPY": 498765
  },
  "totalTickCount": 14856234,
  "latestTicks": {
    "EURUSD": {
      "bid": 1.1045156847,
      "ask": 1.1046156847,
      "high24h": 1.1051234567,
      "low24h": 1.1032456789,
      "timestamp": 1769772266771,
      "lp": "YOFX",
      "ageSeconds": 0.025
    },
    "GBPUSD": {
      "bid": 1.2678234567,
      "ask": 1.2679234567,
      "high24h": 1.2695678901,
      "low24h": 1.2634567890,
      "timestamp": 1769772265234,
      "lp": "YOFX",
      "ageSeconds": 1.562
    }
  },
  "averageLatencyMs": 8.34,
  "maxLatencyMs": 45.23,
  "p99LatencyMs": 32.15,
  "dataQuality": {
    "realDataPercentage": 100.0,
    "simulationDataPercentage": 0.0,
    "mockDataPercentage": 0.0,
    "dataVerification": "passed"
  },
  "last24hStats": {
    "totalTicks": 14856234,
    "averageTickRate": 172.12,
    "peakTickRate": 1245,
    "minTickRate": 0,
    "uptime": 99.98
  }
}
```

**Key Fields**:
- `dataQuality.realDataPercentage`: Should be 100.0
- `dataQuality.simulationDataPercentage`: Should be 0.0
- `dataQuality.mockDataPercentage`: Should be 0.0
- `dataQuality.dataVerification`: Should be "passed"
- All ticks should show `lp: "YOFX"`
- High/low should be populated

---

## WebSocket Market Data API

### Connection

**URL**: `ws://api.example.com/ws`

**Authentication**: Bearer token passed via query or header

**Connection Example** (JavaScript):
```javascript
const token = 'your_jwt_token';
const ws = new WebSocket(`ws://api.example.com/ws?token=${token}`);

ws.onopen = () => {
  console.log('Connected to real market data feed');

  // Subscribe to symbols
  ws.send(JSON.stringify({
    type: 'subscribe',
    symbols: ['EURUSD', 'GBPUSD', 'USDJPY']
  }));
};

ws.onmessage = (event) => {
  const tick = JSON.parse(event.data);
  console.log(`${tick.symbol}: ${tick.bid}-${tick.ask} (LP: ${tick.lp})`);
};

ws.onerror = (error) => console.error('WebSocket error:', error);
ws.onclose = () => console.log('Disconnected from market data');
```

### Subscribe to Symbols

**Message**:
```json
{
  "type": "subscribe",
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"]
}
```

**Response**:
```json
{
  "type": "subscription_confirmed",
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"],
  "status": "subscribed"
}
```

### Receive Market Ticks

**Real-time Tick** (received every market data update):
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.1045156847,
  "ask": 1.1046156847,
  "spread": 0.0001,
  "high24h": 1.1051234567,
  "low24h": 1.1032456789,
  "dailyChange": 0.0234,
  "timestamp": 1769772266771,
  "lp": "YOFX",
  "source": "real"
}
```

**Guarantees**:
- ✅ Every tick is from YOFX FIX feed
- ✅ No simulated or mock ticks
- ✅ High/low are real YOFX values
- ✅ LP tag is always "YOFX"
- ✅ Timestamps are accurate to millisecond

### Unsubscribe from Symbols

**Message**:
```json
{
  "type": "unsubscribe",
  "symbols": ["EURUSD"]
}
```

---

## Real Data Verification

### Data Verification Endpoints

#### Verify Data Source

**Request**:
```http
GET /api/verify/data-source HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
```

**Response**:
```json
{
  "verification": "passed",
  "dataSource": "YOFX_FIX_FEED",
  "lastVerified": "2026-01-30T12:34:56Z",
  "results": {
    "allTicksFromYOFX": true,
    "noSimulationDetected": true,
    "noMockDataDetected": true,
    "highLowVerified": true,
    "timestampsValid": true,
    "spreadsReasonable": true
  },
  "sampleSize": 10000,
  "verificationDuration": "2.345s",
  "confidence": 99.97
}
```

#### Check for Simulation

**Request**:
```http
GET /api/verify/no-simulation HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
```

**Response**:
```json
{
  "hasSimulation": false,
  "hasArtificialData": false,
  "hasMockFallback": false,
  "details": {
    "simulationCodeFound": false,
    "mockDataDetected": false,
    "fallbackTriggered": false,
    "artifactsPrecision": "none"
  },
  "verification": "PASSED",
  "timestamp": "2026-01-30T12:34:56Z"
}
```

#### LP Tag Verification

**Request**:
```http
GET /api/verify/lp-tags HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
```

**Response**:
```json
{
  "verification": "passed",
  "totalTicks": 1000000,
  "lpDistribution": {
    "YOFX": 1000000,
    "MOCK": 0,
    "SIMULATION": 0
  },
  "allTicksFromYOFX": true,
  "noOtherSources": true,
  "timestamp": "2026-01-30T12:34:56Z"
}
```

---

## Data Quality Guarantees

### Assurances

| Guarantee | Status | Verification |
|-----------|--------|--------------|
| 100% real market data | ✅ | All ticks from YOFX FIX |
| No simulation fallback | ✅ | Code review & monitoring |
| No mock data | ✅ | LP tag always "YOFX" |
| High/low from real data | ✅ | YOFX FIX feed |
| Accurate timestamps | ✅ | <1ms precision |
| Correct spreads | ✅ | Real bid-ask pairs |
| No data gaps | ⚠️ | Monitored continuously |

### Performance SLA

- **Uptime**: 99.8% (streaming)
- **Latency**: <50ms from YOFX to client
- **Data Freshness**: <100ms (typically <10ms)
- **Availability**: 24/5 (excluding weekends)

### Monitoring

System continuously monitors:
- ✅ YOFX FIX connection status
- ✅ Tick rate and freshness
- ✅ High/low accuracy
- ✅ LP tag consistency
- ✅ Data anomalies

---

## Admin Endpoints

### GET /admin/fix/status

Get current FIX session status.

**Request**:
```http
GET /admin/fix/status HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
```

**Response**:
```json
{
  "sessions": {
    "YOFX2": {
      "status": "LOGGED_IN",
      "inboundSeqNum": 123456,
      "outboundSeqNum": 67890,
      "lastMessageTime": "2026-01-30T12:34:55.234Z"
    },
    "YOFX1": {
      "status": "DISCONNECTED",
      "inboundSeqNum": 0,
      "outboundSeqNum": 0
    }
  },
  "primarySession": "YOFX2",
  "backupSession": "YOFX1",
  "overallStatus": "HEALTHY"
}
```

### POST /admin/fix/restart

Restart FIX sessions.

**Request**:
```http
POST /admin/fix/restart HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
Content-Type: application/json
```

**Response**:
```json
{
  "status": "restart_initiated",
  "message": "FIX sessions restarting...",
  "expectedRestartTime": "5s"
}
```

### POST /admin/symbol/enable

Enable a symbol for market data.

**Request**:
```http
POST /admin/symbol/enable HTTP/1.1
Host: api.example.com
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "symbol": "EURUSD"
}
```

**Response**:
```json
{
  "symbol": "EURUSD",
  "status": "enabled",
  "subscribedAt": "2026-01-30T12:34:56Z"
}
```

---

## Error Handling

### Market Data Unavailable

If YOFX connection is lost, the system will:
1. Log error immediately
2. Attempt automatic reconnection
3. Alert monitoring team
4. **NOT** fall back to simulation

**Error Response**:
```json
{
  "error": "market_data_unavailable",
  "message": "YOFX connection lost - no simulation fallback",
  "timestamp": "2026-01-30T12:34:56Z",
  "recommendation": "Contact support - check YOFX status"
}
```

### No Mock Data Fallback

The system will **never** return mock or simulated data. If real data is unavailable:
- WebSocket connection may close
- API returns error (not mock data)
- Client must handle gracefully

---

## Migration Guide

### For Applications Using Old Mock Data

If your application was built with mock data support, update to expect:

**Before** (could receive mock):
```json
{
  "symbol": "EURUSD",
  "bid": 1.1045,
  "ask": 1.1046,
  "lp": "MOCK"  // Could be MOCK
}
```

**After** (only real):
```json
{
  "symbol": "EURUSD",
  "bid": 1.1045,
  "ask": 1.1046,
  "lp": "YOFX",  // Always YOFX
  "high24h": 1.1051,
  "low24h": 1.1032,
  "dailyChange": 0.0234
}
```

### Required Updates

1. ✅ Update LP tag checks: `if tick.lp == "YOFX"` (not "MOCK")
2. ✅ Handle new fields: `high24h`, `low24h`, `dailyChange`
3. ✅ Remove simulation fallback code
4. ✅ Update tests to use real data samples
5. ✅ Update documentation

---

## Testing Real Data

### Test Endpoints (Development)

```bash
# Get latest tick
curl http://localhost:7999/api/market/ticks/EURUSD \
  -H "Authorization: Bearer $TOKEN"

# Get diagnostics
curl http://localhost:7999/api/diagnostics/market-data \
  -H "Authorization: Bearer $TOKEN"

# Verify data source
curl http://localhost:7999/api/verify/data-source \
  -H "Authorization: Bearer $TOKEN"

# Check for simulation
curl http://localhost:7999/api/verify/no-simulation \
  -H "Authorization: Bearer $TOKEN"
```

### Test Script

```bash
#!/bin/bash
echo "Testing Real Data API..."

# Test 1: Verify data source
echo "1. Verifying data source..."
curl -s http://localhost:7999/api/verify/data-source \
  -H "Authorization: Bearer $TOKEN" | jq '.results'

# Test 2: Check LP tags
echo "2. Checking LP tags..."
curl -s http://localhost:7999/api/verify/lp-tags \
  -H "Authorization: Bearer $TOKEN" | jq '.lpDistribution'

# Test 3: Get market diagnostics
echo "3. Getting market diagnostics..."
curl -s http://localhost:7999/api/diagnostics/market-data \
  -H "Authorization: Bearer $TOKEN" | jq '.dataQuality'

# Test 4: Verify no simulation
echo "4. Verifying no simulation..."
curl -s http://localhost:7999/api/verify/no-simulation \
  -H "Authorization: Bearer $TOKEN" | jq '.verification'

echo "All tests completed!"
```

---

## Document Information

- **Version**: 2.0 (Real Data Edition)
- **Date**: January 30, 2026
- **Status**: ✅ Production Ready
- **Previous Version**: 1.0 (Mock/Hybrid)

---

## Key Takeaways

1. ✅ **100% Real Data**: All ticks come from YOFX FIX feed
2. ✅ **No Simulation**: No fallback to mock data
3. ✅ **Enhanced Data**: High/low 24-hour tracking added
4. ✅ **Verifiable**: Multiple verification endpoints available
5. ✅ **Production Ready**: Ready for live trading

**Assurance**: Every market data tick you receive is authenticated real data from YOFX.
