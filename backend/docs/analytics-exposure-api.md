# Exposure Heatmap Analytics API

## Overview

The Exposure Heatmap API provides real-time and historical exposure analytics for the trading platform. All endpoints return dynamic data calculated from current positions - no hardcoded values.

## Endpoints

### 1. GET /api/analytics/exposure/heatmap

Returns exposure heatmap data aggregated over time intervals.

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| start_time | int64 | No | 24h ago | Unix timestamp for start of range |
| end_time | int64 | No | now | Unix timestamp for end of range |
| interval | string | No | 1h | Time interval: 15m, 1h, 4h, 1d |
| symbols | string | No | all | Comma-separated list of symbols to filter |

#### Response

```json
{
  "timestamps": [1642444800, 1642448400, ...],
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"],
  "data": [
    [110000.50, 55000.25, -30000.10],
    [115000.75, 60000.00, -28000.50],
    ...
  ],
  "max_exposure": 115000.75,
  "min_exposure": -30000.10,
  "interval": "1h"
}
```

#### Field Descriptions

- `timestamps`: Array of Unix timestamps for each time bucket
- `symbols`: Array of symbol names (sorted alphabetically)
- `data`: 2D array where data[i][j] = exposure for timestamp[i] and symbol[j]
- `max_exposure`: Maximum exposure value across all data points
- `min_exposure`: Minimum exposure value across all data points
- `interval`: Echo of the requested interval

#### Example Request

```bash
curl "http://localhost:7999/api/analytics/exposure/heatmap?start_time=1642444800&end_time=1642531200&interval=1h&symbols=EURUSD,GBPUSD"
```

---

### 2. GET /api/analytics/exposure/current

Returns current exposure breakdown by symbol.

#### Query Parameters

None

#### Response

```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "net_exposure": 110000.50,
      "long": 220000.75,
      "short": 110000.25,
      "utilization_pct": 11.0,
      "limit": 1000000.0,
      "status": "normal"
    },
    {
      "symbol": "GBPUSD",
      "net_exposure": -50000.25,
      "long": 75000.00,
      "short": 125000.25,
      "utilization_pct": 5.0,
      "limit": 1000000.0,
      "status": "normal"
    }
  ],
  "timestamp": 1642531200
}
```

#### Field Descriptions

- `symbol`: Trading symbol name
- `net_exposure`: Net exposure (long - short) in notional value
- `long`: Total long position exposure
- `short`: Total short position exposure
- `utilization_pct`: Percentage of limit used (abs(net_exposure) / limit * 100)
- `limit`: Configured exposure limit for the symbol
- `status`: Risk status: "normal" (<75%), "warning" (75-90%), "critical" (>90%)

#### Status Thresholds

- **normal**: Utilization < 75%
- **warning**: Utilization 75-90%
- **critical**: Utilization > 90%

#### Example Request

```bash
curl "http://localhost:7999/api/analytics/exposure/current"
```

---

### 3. GET /api/analytics/exposure/history/{symbol}

Returns exposure timeline for a specific symbol.

#### URL Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| symbol | string | Yes | Trading symbol (e.g., EURUSD) |

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| start_time | int64 | No | 7d ago | Unix timestamp for start of range |
| end_time | int64 | No | now | Unix timestamp for end of range |
| interval | string | No | 1h | Time interval: 15m, 1h, 4h, 1d |

#### Response

```json
{
  "symbol": "EURUSD",
  "timeline": [
    {
      "timestamp": 1642444800,
      "net_exposure": 110000.50,
      "long": 220000.75,
      "short": 110000.25,
      "utilization_pct": 11.0
    },
    {
      "timestamp": 1642448400,
      "net_exposure": 115000.75,
      "long": 230000.00,
      "short": 115000.25,
      "utilization_pct": 11.5
    }
  ]
}
```

#### Field Descriptions

- `symbol`: Trading symbol name
- `timeline`: Array of exposure snapshots over time
  - `timestamp`: Unix timestamp for this snapshot
  - `net_exposure`: Net exposure at this time
  - `long`: Total long exposure at this time
  - `short`: Total short exposure at this time
  - `utilization_pct`: Percentage of limit used

#### Example Request

```bash
curl "http://localhost:7999/api/analytics/exposure/history/EURUSD?start_time=1642444800&end_time=1642531200&interval=1h"
```

---

## Data Calculation

### Notional Value Calculation

For each position, notional value is calculated as:

```
Notional = Volume × Current_Price × Contract_Size
```

Where:
- `Volume`: Position size in lots
- `Current_Price`: Latest market price (or open price if unavailable)
- `Contract_Size`: From symbol specification (default: 100,000 for forex)

### Exposure Aggregation

- **Long Exposure**: Sum of all BUY position notional values
- **Short Exposure**: Sum of all SELL position notional values
- **Net Exposure**: Long Exposure - Short Exposure

### Timeline Generation

1. Divide time range into buckets based on interval
2. For each bucket, calculate exposure from positions open at that time
3. Aggregate by symbol
4. Return time series data

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200 OK`: Success
- `400 Bad Request`: Invalid parameters
- `500 Internal Server Error`: Server error

Error response format:

```json
{
  "error": "Error message description"
}
```

## Performance Considerations

- All data is calculated in real-time from current positions
- No database queries - uses in-memory B-Book engine
- Efficient aggregation with O(n*m) complexity where n=positions, m=time_buckets
- Recommended limits:
  - Time range: < 90 days
  - Interval: >= 15m for large datasets
  - Symbols filter: Use to reduce data size

## Integration Example

### JavaScript/TypeScript

```typescript
// Fetch current exposure
async function getCurrentExposure() {
  const response = await fetch('http://localhost:7999/api/analytics/exposure/current');
  const data = await response.json();

  // Find critical symbols
  const critical = data.symbols.filter(s => s.status === 'critical');
  console.log('Critical exposures:', critical);

  return data;
}

// Fetch heatmap data
async function getHeatmap(startTime: number, endTime: number, interval: string = '1h') {
  const params = new URLSearchParams({
    start_time: startTime.toString(),
    end_time: endTime.toString(),
    interval
  });

  const response = await fetch(`http://localhost:7999/api/analytics/exposure/heatmap?${params}`);
  return response.json();
}

// Fetch symbol history
async function getSymbolHistory(symbol: string, days: number = 7) {
  const endTime = Math.floor(Date.now() / 1000);
  const startTime = endTime - (days * 24 * 60 * 60);

  const params = new URLSearchParams({
    start_time: startTime.toString(),
    end_time: endTime.toString(),
    interval: '1h'
  });

  const response = await fetch(
    `http://localhost:7999/api/analytics/exposure/history/${symbol}?${params}`
  );
  return response.json();
}
```

### Python

```python
import requests
from datetime import datetime, timedelta

BASE_URL = "http://localhost:7999"

def get_current_exposure():
    """Get current exposure by symbol"""
    response = requests.get(f"{BASE_URL}/api/analytics/exposure/current")
    return response.json()

def get_heatmap(start_time, end_time, interval="1h", symbols=None):
    """Get exposure heatmap data"""
    params = {
        "start_time": int(start_time.timestamp()),
        "end_time": int(end_time.timestamp()),
        "interval": interval
    }

    if symbols:
        params["symbols"] = ",".join(symbols)

    response = requests.get(
        f"{BASE_URL}/api/analytics/exposure/heatmap",
        params=params
    )
    return response.json()

def get_symbol_history(symbol, days=7, interval="1h"):
    """Get exposure history for a symbol"""
    end_time = datetime.now()
    start_time = end_time - timedelta(days=days)

    params = {
        "start_time": int(start_time.timestamp()),
        "end_time": int(end_time.timestamp()),
        "interval": interval
    }

    response = requests.get(
        f"{BASE_URL}/api/analytics/exposure/history/{symbol}",
        params=params
    )
    return response.json()

# Example usage
if __name__ == "__main__":
    # Get current exposure
    current = get_current_exposure()
    print(f"Found {len(current['symbols'])} symbols with exposure")

    # Get last 24 hours heatmap
    end = datetime.now()
    start = end - timedelta(hours=24)
    heatmap = get_heatmap(start, end, interval="1h")
    print(f"Heatmap: {len(heatmap['timestamps'])} time points, {len(heatmap['symbols'])} symbols")

    # Get EURUSD history
    history = get_symbol_history("EURUSD", days=7)
    print(f"EURUSD timeline: {len(history['timeline'])} data points")
```

## Testing

Run the test suite:

```bash
cd backend
go test -v ./internal/api/handlers/analytics_exposure_test.go \
  ./internal/api/handlers/analytics_exposure.go \
  ./internal/api/handlers/api.go
```

All tests should pass with real data from the B-Book engine.
