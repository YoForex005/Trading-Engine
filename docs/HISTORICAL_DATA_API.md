# Historical Tick Data API Documentation

## Overview

The Historical Data API provides comprehensive access to tick-level market data with support for bulk downloads, pagination, compression, and rate limiting. The API serves historical tick data stored in the daily file-based system with efficient querying capabilities.

## Base URL

```
http://localhost:7999
```

## Authentication

Currently, most endpoints are public. Admin endpoints (prefixed with `/admin/`) should be protected with authentication in production.

---

## Public API Endpoints

### 1. GET /api/history/ticks/{symbol}

Download historical tick data for a specific symbol with pagination and format options.

**URL Parameters:**
- `symbol` (required): Trading symbol (e.g., EURUSD, GBPUSD, XAUUSD)

**Query Parameters:**
- `from` (optional): Start date in RFC3339 format (default: 7 days ago)
- `to` (optional): End date in RFC3339 format (default: now)
- `format` (optional): Response format - `json` (default), `csv`, `binary`
- `page` (optional): Page number for pagination (default: 1)
- `page_size` (optional): Items per page (default: 1000, max: 10000)

**Response Headers:**
- `Content-Encoding: gzip` (if client accepts gzip)
- `Content-Type: application/json` or `text/csv`

**Response (JSON):**
```json
{
  "symbol": "EURUSD",
  "from": "2026-01-13T00:00:00Z",
  "to": "2026-01-20T00:00:00Z",
  "count": 1000,
  "total_count": 150000,
  "page": 1,
  "page_size": 1000,
  "has_more": true,
  "ticks": [
    {
      "broker_id": "default",
      "symbol": "EURUSD",
      "bid": 1.08245,
      "ask": 1.08255,
      "spread": 0.00010,
      "timestamp": "2026-01-13T00:00:01Z",
      "lp": "YOFX"
    }
  ],
  "format": "json"
}
```

**Example Requests:**

```bash
# Get latest 1000 ticks for EURUSD (JSON)
curl "http://localhost:7999/api/history/ticks/EURUSD"

# Get specific date range with CSV format
curl "http://localhost:7999/api/history/ticks/GBPUSD?from=2026-01-01T00:00:00Z&to=2026-01-10T00:00:00Z&format=csv"

# Get page 2 with 5000 ticks per page
curl "http://localhost:7999/api/history/ticks/XAUUSD?page=2&page_size=5000"

# With gzip compression
curl -H "Accept-Encoding: gzip" "http://localhost:7999/api/history/ticks/EURUSD" | gunzip
```

**Rate Limiting:**
- 100 tokens per IP address
- Refills at 10 tokens/second
- Returns `429 Too Many Requests` when exceeded

---

### 2. POST /api/history/ticks/bulk

Download tick data for multiple symbols at once with automatic gzip compression.

**Request Body:**
```json
{
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"],
  "from": "2026-01-01T00:00:00Z",
  "to": "2026-01-20T00:00:00Z",
  "format": "json"
}
```

**Response:**
```json
{
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"],
  "from": "2026-01-01T00:00:00Z",
  "to": "2026-01-20T00:00:00Z",
  "data": {
    "EURUSD": [ /* array of ticks */ ],
    "GBPUSD": [ /* array of ticks */ ],
    "USDJPY": [ /* array of ticks */ ]
  },
  "count": 450000
}
```

**Constraints:**
- Maximum 50 symbols per request
- Response is always gzipped
- Download is provided as attachment

**Example:**

```bash
curl -X POST http://localhost:7999/api/history/ticks/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "symbols": ["EURUSD", "GBPUSD"],
    "from": "2026-01-10T00:00:00Z",
    "to": "2026-01-20T00:00:00Z"
  }' \
  --output bulk_ticks.json.gz
```

---

### 3. GET /api/history/available

List all available symbols with their earliest/latest ticks and metadata.

**Response:**
```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "earliest_tick": "2025-07-01T00:00:00Z",
      "latest_tick": "2026-01-20T15:30:00Z",
      "total_ticks": 1500000,
      "available_days": 203,
      "last_updated": "2026-01-20T15:30:00Z"
    },
    {
      "symbol": "GBPUSD",
      "earliest_tick": "2025-07-01T00:00:00Z",
      "latest_tick": "2026-01-20T15:30:00Z",
      "total_ticks": 1450000,
      "available_days": 203,
      "last_updated": "2026-01-20T15:30:00Z"
    }
  ],
  "total": 32
}
```

**Example:**

```bash
curl http://localhost:7999/api/history/available
```

**Use Cases:**
- Discover what data is available before downloading
- Check data freshness and coverage
- Build dynamic symbol selectors in UIs

---

### 4. GET /api/history/symbols

Get comprehensive list of all tradeable symbols with metadata and availability status.

**Response:**
```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "display_name": "EURUSD",
      "category": "forex",
      "available": true,
      "tick_count": 1500000,
      "last_updated": "2026-01-20T15:30:00Z"
    },
    {
      "symbol": "XAUUSD",
      "display_name": "XAUUSD",
      "category": "metals",
      "available": true,
      "tick_count": 800000,
      "last_updated": "2026-01-20T15:30:00Z"
    }
  ],
  "total": 128
}
```

**Symbol Categories:**
- `forex` - Forex major/cross/exotic pairs
- `metals` - Gold, Silver, Platinum, Palladium
- `energy` - Oil, Natural Gas
- `indices` - Stock indices
- `crypto` - Cryptocurrencies
- `commodities` - Other commodities

**Example:**

```bash
curl http://localhost:7999/api/history/symbols
```

---

## Admin API Endpoints

### 5. POST /admin/history/backfill

Import external historical data for backfilling.

**Request:**
```json
{
  "symbol": "EURUSD",
  "source": "external_provider",
  "ticks": [
    {
      "broker_id": "default",
      "symbol": "EURUSD",
      "bid": 1.08245,
      "ask": 1.08255,
      "spread": 0.00010,
      "timestamp": "2025-01-01T00:00:00Z",
      "lp": "EXTERNAL"
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "symbol": "EURUSD",
  "count": 50000,
  "source": "external_provider",
  "message": "Successfully backfilled 50000 ticks for EURUSD"
}
```

**Example:**

```bash
curl -X POST http://localhost:7999/admin/history/backfill \
  -H "Content-Type: application/json" \
  -d @backfill_data.json
```

---

### 6. GET /admin/history/stats

Get comprehensive statistics about historical data storage.

**Response:**
```json
{
  "total_symbols": 32,
  "total_ticks": 45000000,
  "total_size_bytes": 4500000000,
  "total_size_mb": 4292.96,
  "oldest_tick": "2025-07-01T00:00:00Z",
  "newest_tick": "2026-01-20T15:30:00Z",
  "days_of_data": 203,
  "storage_health": "healthy",
  "symbol_stats": [
    {
      "symbol": "EURUSD",
      "tick_count": 1500000,
      "size_bytes": 150000000,
      "oldest_tick": "2025-07-01T00:00:00Z",
      "newest_tick": "2026-01-20T15:30:00Z",
      "avg_tick_rate": 7389.16
    }
  ]
}
```

**Storage Health Values:**
- `healthy` - < 10GB total
- `warning` - 10GB - 50GB
- `critical` - > 50GB

**Example:**

```bash
curl http://localhost:7999/admin/history/stats
```

---

### 7. POST /admin/history/cleanup

Clean up old historical data files.

**Request:**
```json
{
  "older_than_days": 90,
  "symbols": ["EURUSD", "GBPUSD"],
  "dry_run": true
}
```

**Response:**
```json
{
  "success": true,
  "files_deleted": 180,
  "size_freed_mb": 1250.45,
  "cutoff_date": "2025-10-22",
  "dry_run": true
}
```

**Parameters:**
- `older_than_days` (required): Delete files older than N days
- `symbols` (optional): Specific symbols (empty = all)
- `dry_run` (optional): Preview without deleting (default: false)

**Example:**

```bash
# Dry run first to see what would be deleted
curl -X POST http://localhost:7999/admin/history/cleanup \
  -H "Content-Type: application/json" \
  -d '{"older_than_days": 90, "dry_run": true}'

# Actually delete
curl -X POST http://localhost:7999/admin/history/cleanup \
  -H "Content-Type: application/json" \
  -d '{"older_than_days": 90, "dry_run": false}'
```

---

### 8. GET /admin/history/monitoring

Get real-time monitoring data for the historical data system.

**Response:**
```json
{
  "active_symbols": 32,
  "ticks_per_second": 45.2,
  "avg_latency_ms": 12.5,
  "memory_usage_mb": 256.0,
  "disk_usage_mb": 4292.96,
  "last_tick_received": "2026-01-20T15:30:00Z",
  "health": "healthy",
  "alerts": []
}
```

**Health Values:**
- `healthy` - All systems operational
- `warning` - Some issues detected (see alerts)
- `critical` - Immediate attention required

**Example:**

```bash
curl http://localhost:7999/admin/history/monitoring
```

---

## Response Formats

### JSON Format

Default format. Returns tick data as JSON array with full precision.

**Advantages:**
- Easy to parse
- Widely supported
- Preserves data types

### CSV Format

Comma-separated values for Excel/spreadsheet import.

**Format:**
```csv
Timestamp,Symbol,Bid,Ask,Spread,LP
2026-01-13T00:00:01Z,EURUSD,1.08245,1.08255,0.00010,YOFX
```

**Advantages:**
- Excel-compatible
- Smaller file size
- Human-readable

### Binary Format (Planned)

Custom binary protocol for maximum efficiency.

**Currently returns JSON** - Binary protocol implementation pending.

---

## Rate Limiting

### Token Bucket Algorithm

- **Bucket Capacity:** 100 tokens per IP
- **Refill Rate:** 10 tokens/second
- **Cost:** 1 token per request

### Rate Limit Headers (Future)

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1642683600
Retry-After: 10
```

### Exceeded Response

```
HTTP/1.1 429 Too Many Requests
Retry-After: 10

Rate limit exceeded. Please try again later.
```

---

## Error Handling

### Common Error Responses

**400 Bad Request:**
```json
{
  "error": "Symbol is required"
}
```

**404 Not Found:**
```json
{
  "error": "Symbol INVALID not found"
}
```

**429 Too Many Requests:**
```
Rate limit exceeded. Please try again later.
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to fetch historical data: database error"
}
```

---

## Compression Support

### Automatic gzip Compression

The API automatically compresses responses when client supports it:

```bash
curl -H "Accept-Encoding: gzip" http://localhost:7999/api/history/ticks/EURUSD | gunzip
```

**Benefits:**
- 75-90% size reduction
- Faster downloads
- Lower bandwidth usage

---

## Best Practices

### 1. Use Pagination for Large Datasets

```bash
# Don't do this (could timeout)
curl "http://localhost:7999/api/history/ticks/EURUSD?from=2025-01-01T00:00:00Z"

# Do this instead
for page in {1..10}; do
  curl "http://localhost:7999/api/history/ticks/EURUSD?page=$page&page_size=5000"
done
```

### 2. Enable Compression

```bash
curl -H "Accept-Encoding: gzip" "http://localhost:7999/api/history/ticks/EURUSD" | gunzip > eurusd.json
```

### 3. Use Bulk Downloads for Multiple Symbols

```bash
# Instead of multiple single-symbol requests
curl -X POST http://localhost:7999/api/history/ticks/bulk \
  -H "Content-Type: application/json" \
  -d '{"symbols": ["EURUSD", "GBPUSD", "USDJPY"]}'
```

### 4. Check Availability First

```bash
# Check what's available before downloading
curl http://localhost:7999/api/history/available
```

### 5. Respect Rate Limits

```bash
# Add delays between requests
for symbol in EURUSD GBPUSD USDJPY; do
  curl "http://localhost:7999/api/history/ticks/$symbol"
  sleep 1  # Respect rate limits
done
```

---

## Use Cases

### 1. Backtesting Trading Strategies

```python
import requests
import pandas as pd

# Download historical data
response = requests.get(
    "http://localhost:7999/api/history/ticks/EURUSD",
    params={
        "from": "2025-01-01T00:00:00Z",
        "to": "2026-01-01T00:00:00Z",
        "page_size": 10000
    },
    headers={"Accept-Encoding": "gzip"}
)

data = response.json()
df = pd.DataFrame(data['ticks'])

# Run backtest
# ...
```

### 2. Building OHLC Candles

```javascript
const axios = require('axios');

async function buildCandles(symbol, timeframe) {
  const response = await axios.get(
    `http://localhost:7999/api/history/ticks/${symbol}`,
    { params: { page_size: 10000 } }
  );

  const ticks = response.data.ticks;
  // Aggregate into OHLC candles
  // ...
}
```

### 3. Market Analysis

```bash
# Download multiple symbols for correlation analysis
curl -X POST http://localhost:7999/api/history/ticks/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "symbols": ["EURUSD", "GBPUSD", "USDJPY", "EURJPY"],
    "from": "2026-01-01T00:00:00Z",
    "to": "2026-01-20T00:00:00Z"
  }' | gunzip > forex_data.json
```

---

## Performance Optimization

### Storage Layer

- **Daily File Rotation:** Data organized by date for efficient querying
- **In-Memory Caching:** Recent ticks cached for fast access
- **Indexed Queries:** Symbol-based indexing for quick lookups

### API Layer

- **Rate Limiting:** Token bucket algorithm prevents abuse
- **Gzip Compression:** Automatic compression for bandwidth savings
- **Pagination:** Prevents large response timeouts
- **Concurrent Handling:** Non-blocking I/O for high throughput

### Future Enhancements

- [ ] WebSocket streaming for real-time updates
- [ ] Binary protocol for maximum efficiency
- [ ] Columnar storage (Parquet) for analytics
- [ ] Query result caching with Redis
- [ ] GraphQL interface for flexible querying
- [ ] Multi-region CDN distribution

---

## Support

For issues or questions:
- GitHub: https://github.com/epic1st/rtx/issues
- Email: support@rtx.trading (if applicable)

---

## Changelog

### v1.0.0 (2026-01-20)
- Initial release
- 5 public endpoints + 4 admin endpoints
- Pagination, compression, rate limiting
- JSON/CSV format support
- Bulk download capability
- Historical data backfilling
