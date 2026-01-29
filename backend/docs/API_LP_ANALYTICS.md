# LP Analytics API Documentation

## Overview

The LP Analytics API provides performance comparison and analysis endpoints for liquidity providers. All endpoints use real-time database queries with no hardcoded values.

**Base URL**: `http://localhost:7999/api/analytics/lp`

---

## Authentication

All endpoints use the same authentication mechanism as other API endpoints. Include the `Authorization` header if required.

---

## Endpoints

### 1. LP Comparison

**GET** `/api/analytics/lp/comparison`

Compare all liquidity providers by performance metrics.

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `start_time` | string (RFC3339) | No | 24h ago | Start of time range |
| `end_time` | string (RFC3339) | No | Now | End of time range |
| `symbol` | string | No | All | Filter by specific symbol (e.g., "EURUSD") |
| `metric` | string | No | `latency` | Ranking metric: `latency`, `fill_rate`, or `slippage` |

#### Response

```json
{
  "lps": [
    {
      "name": "OANDA",
      "avg_latency_ms": 45.23,
      "fill_rate_pct": 98.5,
      "slippage_bps": 2,
      "volume_24h": 1250000.50,
      "rank": 1
    },
    {
      "name": "Binance",
      "avg_latency_ms": 52.10,
      "fill_rate_pct": 97.8,
      "slippage_bps": 3,
      "volume_24h": 980000.00,
      "rank": 2
    }
  ]
}
```

#### Field Descriptions

- `name`: LP identifier
- `avg_latency_ms`: Average execution latency in milliseconds
- `fill_rate_pct`: Percentage of successful order fills (0-100)
- `slippage_bps`: Average slippage in basis points
- `volume_24h`: Total volume in last 24 hours
- `rank`: Ranking based on selected metric (1 = best)

#### cURL Examples

```bash
# Compare all LPs (default: last 24h, latency metric)
curl -X GET "http://localhost:7999/api/analytics/lp/comparison"

# Compare by fill rate for EURUSD
curl -X GET "http://localhost:7999/api/analytics/lp/comparison?metric=fill_rate&symbol=EURUSD"

# Custom time range
curl -X GET "http://localhost:7999/api/analytics/lp/comparison?start_time=2026-01-01T00:00:00Z&end_time=2026-01-19T23:59:59Z"

# Compare by slippage
curl -X GET "http://localhost:7999/api/analytics/lp/comparison?metric=slippage"
```

#### Response Codes

- `200 OK`: Success
- `400 Bad Request`: Invalid metric parameter
- `500 Internal Server Error`: Database query failed

---

### 2. LP Performance Detail

**GET** `/api/analytics/lp/performance/{lp_name}`

Get detailed performance metrics and timeline for a specific LP.

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `lp_name` | string | Yes | LP identifier (e.g., "OANDA", "Binance") |

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `start_time` | string (RFC3339) | No | 24h ago | Start of time range |
| `end_time` | string (RFC3339) | No | Now | End of time range |
| `symbol` | string | No | All | Filter by specific symbol |

#### Response

```json
{
  "lp_name": "OANDA",
  "metrics": {
    "latency_p50": 42.5,
    "latency_p95": 78.2,
    "latency_p99": 125.8,
    "fill_rate": 98.5,
    "avg_slippage": 2,
    "uptime_pct": 99.95
  },
  "timeline": [
    {
      "timestamp": "2026-01-19T10:00:00Z",
      "avg_latency": 45.2,
      "fill_rate": 98.3,
      "volume": 125000.50,
      "order_count": 342
    },
    {
      "timestamp": "2026-01-19T11:00:00Z",
      "avg_latency": 43.8,
      "fill_rate": 98.7,
      "volume": 132000.00,
      "order_count": 378
    }
  ]
}
```

#### Field Descriptions

**Metrics:**
- `latency_p50`: 50th percentile latency (median) in ms
- `latency_p95`: 95th percentile latency in ms
- `latency_p99`: 99th percentile latency in ms
- `fill_rate`: Percentage of successful fills
- `avg_slippage`: Average slippage in basis points
- `uptime_pct`: Uptime percentage (0-100)

**Timeline:**
- `timestamp`: Time bucket timestamp
- `avg_latency`: Average latency for this time bucket
- `fill_rate`: Fill rate for this time bucket
- `volume`: Total volume for this time bucket
- `order_count`: Number of orders in this time bucket

#### cURL Examples

```bash
# Get OANDA performance (last 24h)
curl -X GET "http://localhost:7999/api/analytics/lp/performance/OANDA"

# Get Binance performance for EURUSD
curl -X GET "http://localhost:7999/api/analytics/lp/performance/Binance?symbol=EURUSD"

# Custom time range
curl -X GET "http://localhost:7999/api/analytics/lp/performance/OANDA?start_time=2026-01-18T00:00:00Z&end_time=2026-01-19T00:00:00Z"
```

#### Response Codes

- `200 OK`: Success
- `400 Bad Request`: LP name missing
- `404 Not Found`: LP not found or no data available
- `500 Internal Server Error`: Database query failed

---

### 3. LP Ranking

**GET** `/api/analytics/lp/ranking`

Get LP rankings by a specific metric with percentile scores.

#### Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `start_time` | string (RFC3339) | No | 24h ago | Start of time range |
| `end_time` | string (RFC3339) | No | Now | End of time range |
| `metric` | string | No | `latency` | Metric: `latency`, `fill_rate`, `slippage`, or `volume` |
| `limit` | integer | No | 10 | Maximum number of results |

#### Response

```json
{
  "rankings": [
    {
      "rank": 1,
      "lp_name": "OANDA",
      "value": 42.5,
      "percentile": 95.5
    },
    {
      "rank": 2,
      "lp_name": "Binance",
      "value": 48.3,
      "percentile": 87.2
    },
    {
      "rank": 3,
      "lp_name": "LMAX",
      "value": 52.1,
      "percentile": 78.8
    }
  ]
}
```

#### Field Descriptions

- `rank`: Ranking position (1 = best)
- `lp_name`: LP identifier
- `value`: Metric value (units depend on metric type)
- `percentile`: Performance percentile (0-100, higher is better)

#### Metric Value Units

- `latency`: milliseconds (lower is better)
- `fill_rate`: percentage (higher is better)
- `slippage`: basis points (lower is better)
- `volume`: total volume (higher indicates more liquidity)

#### cURL Examples

```bash
# Top 10 by latency (default)
curl -X GET "http://localhost:7999/api/analytics/lp/ranking"

# Top 5 by fill rate
curl -X GET "http://localhost:7999/api/analytics/lp/ranking?metric=fill_rate&limit=5"

# Top 3 by volume
curl -X GET "http://localhost:7999/api/analytics/lp/ranking?metric=volume&limit=3"

# Top 10 by slippage in custom time range
curl -X GET "http://localhost:7999/api/analytics/lp/ranking?metric=slippage&start_time=2026-01-01T00:00:00Z&end_time=2026-01-19T23:59:59Z"
```

#### Response Codes

- `200 OK`: Success
- `400 Bad Request`: Invalid metric parameter
- `500 Internal Server Error`: Database query failed

---

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "descriptive error message"
}
```

### Common Errors

**400 Bad Request**
```json
{
  "error": "invalid metric. Must be one of: latency, fill_rate, slippage"
}
```

**404 Not Found**
```json
{
  "error": "LP not found or no data available"
}
```

**500 Internal Server Error**
```json
{
  "error": "database query failed: <details>"
}
```

---

## Database Schema

The API queries the following tables:

### Primary Tables

1. **liquidity_providers**
   - LP configuration and metadata
   - Fields: `id`, `name`, `is_active`, `priority`

2. **lp_order_routing**
   - Order routing decisions and execution metrics
   - Fields: `lp_id`, `order_id`, `execution_latency_ms`, `actual_slippage`, `success`

3. **lp_performance_metrics**
   - Pre-aggregated performance metrics
   - Fields: `lp_id`, `time_bucket`, `avg_fill_time_ms`, `total_volume`, `uptime_seconds`

4. **orders**
   - Order details
   - Fields: `id`, `symbol`, `volume`, `status`

### Indexes Used

- `idx_lp_order_routing_created_at` - Time-based queries
- `idx_lp_perf_lp_symbol_time` - Performance metrics by LP and time
- `idx_lp_order_routing_lp_id` - LP filtering

---

## Performance Considerations

### Query Optimization

1. **Time Range**: Queries use indexed `created_at` and `time_bucket` columns
2. **Percentiles**: Calculated using PostgreSQL `PERCENTILE_CONT` function
3. **Connection Pooling**: Max 25 connections, 5 idle
4. **Default Limits**: 24-hour window default to prevent large scans

### Recommended Usage

- Use specific time ranges when possible
- Apply symbol filters for focused analysis
- Set reasonable `limit` values for ranking queries (default: 10)
- Consider caching results for dashboards

### Expected Response Times

- LP Comparison: < 200ms (typical)
- LP Performance: < 300ms (includes timeline)
- LP Ranking: < 150ms (with limit=10)

---

## Configuration

### Environment Variables

Required database configuration:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=trading_engine
DB_SSLMODE=disable  # or 'require' for production
```

### Database Connection

The handler establishes a connection pool on initialization:
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

---

## Testing

Run the test suite:

```bash
# Run all analytics LP tests
cd backend
go test -v ./internal/api/handlers -run TestHandleLP

# Run specific test
go test -v ./internal/api/handlers -run TestHandleLPComparison

# Run with database integration
DB_HOST=localhost DB_USER=postgres DB_PASSWORD=postgres go test -v ./internal/api/handlers
```

---

## Integration Examples

### JavaScript/TypeScript

```typescript
interface LPComparison {
  lps: Array<{
    name: string;
    avg_latency_ms: number;
    fill_rate_pct: number;
    slippage_bps: number;
    volume_24h: number;
    rank: number;
  }>;
}

async function getLPComparison(): Promise<LPComparison> {
  const response = await fetch(
    'http://localhost:7999/api/analytics/lp/comparison?metric=latency'
  );
  return response.json();
}
```

### Python

```python
import requests
from datetime import datetime, timedelta

def get_lp_ranking(metric='latency', limit=10):
    """Get LP ranking by metric"""
    url = 'http://localhost:7999/api/analytics/lp/ranking'
    params = {
        'metric': metric,
        'limit': limit,
        'start_time': (datetime.now() - timedelta(days=7)).isoformat() + 'Z',
        'end_time': datetime.now().isoformat() + 'Z'
    }
    response = requests.get(url, params=params)
    response.raise_for_status()
    return response.json()
```

### Go

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"
)

type LPPerformance struct {
    LPName   string `json:"lp_name"`
    Metrics  struct {
        LatencyP50  float64 `json:"latency_p50"`
        LatencyP95  float64 `json:"latency_p95"`
        LatencyP99  float64 `json:"latency_p99"`
        FillRate    float64 `json:"fill_rate"`
        AvgSlippage int     `json:"avg_slippage"`
        UptimePct   float64 `json:"uptime_pct"`
    } `json:"metrics"`
}

func getLPPerformance(lpName string) (*LPPerformance, error) {
    url := fmt.Sprintf("http://localhost:7999/api/analytics/lp/performance/%s", lpName)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var perf LPPerformance
    if err := json.NewDecoder(resp.Body).Decode(&perf); err != nil {
        return nil, err
    }
    return &perf, nil
}
```

---

## Changelog

### v1.0.0 (2026-01-19)
- Initial release
- Three endpoints: comparison, performance, ranking
- Real-time database queries
- Percentile calculations (p50, p95, p99)
- Symbol filtering support
- Dynamic time range queries
- Production error handling
- Comprehensive test coverage

---

## Support

For issues or questions:
- Check database connectivity: `psql -h localhost -U postgres -d trading_engine`
- Review logs: Server logs show `[Analytics]` prefix
- Verify tables exist: `\dt *lp*` in psql
- Check indexes: `\di *lp*` in psql
