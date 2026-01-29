# RTX Trading Engine - Tick Storage API

**Version:** 1.0
**Base URL:** `https://api.rtx.com`
**Authentication:** Bearer Token (JWT)

---

## Overview

The Tick Storage API provides access to historical tick data stored in TimescaleDB. It supports:
- Streaming downloads for backtesting
- Paginated queries for analysis
- OHLC data export
- Storage statistics and management

All endpoints require authentication via JWT token in the `Authorization` header.

---

## Endpoints

### 1. Download Tick Data (Streaming)

Download historical tick data for backtesting.

**Endpoint:** `GET /api/ticks/download`

**Query Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `symbol` | string | Yes | Trading symbol | `EURUSD` |
| `start_date` | ISO8601 | Yes | Start timestamp (UTC) | `2026-01-01T00:00:00Z` |
| `end_date` | ISO8601 | Yes | End timestamp (UTC) | `2026-01-20T23:59:59Z` |
| `format` | enum | No | Output format (`json`, `csv`) | `csv` (default: `json`) |
| `compression` | enum | No | Compression type (`none`, `gzip`, `zstd`) | `gzip` (default: `gzip`) |

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.rtx.com/api/ticks/download?symbol=EURUSD&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z&format=csv&compression=gzip" \
  --output eurusd_jan2026.csv.gz
```

**Response:**

- **Status:** `200 OK`
- **Content-Type:** `application/gzip` (or `application/json`, `text/csv` based on format)
- **Content-Disposition:** `attachment; filename="EURUSD_2026-01-01_2026-01-20.csv.gz"`
- **Body:** Streaming download of tick data

**CSV Format:**

```csv
timestamp,broker_id,symbol,bid,ask,spread,lp
2026-01-20T15:30:45.123Z,default,EURUSD,1.08456000,1.08458000,0.00002000,YOFX1
2026-01-20T15:30:45.456Z,default,EURUSD,1.08457000,1.08459000,0.00002000,YOFX1
```

**JSON Format:**

```json
[
  {
    "timestamp": "2026-01-20T15:30:45.123Z",
    "broker_id": "default",
    "symbol": "EURUSD",
    "bid": 1.08456,
    "ask": 1.08458,
    "spread": 0.00002,
    "lp": "YOFX1"
  }
]
```

**Error Responses:**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing or invalid parameters |
| `401 Unauthorized` | Invalid or missing JWT token |
| `404 Not Found` | Symbol not found |
| `500 Internal Server Error` | Database query failed |

---

### 2. Query Tick Data (Paginated)

Query tick data with pagination for analysis.

**Endpoint:** `POST /api/ticks/query`

**Request Body:**

```json
{
  "symbols": ["EURUSD", "GBPUSD"],
  "start_date": "2026-01-01T00:00:00Z",
  "end_date": "2026-01-20T23:59:59Z",
  "limit": 10000,
  "offset": 0
}
```

**Example Request:**

```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"symbols": ["EURUSD"], "start_date": "2026-01-20T00:00:00Z", "end_date": "2026-01-20T23:59:59Z", "limit": 1000, "offset": 0}' \
  https://api.rtx.com/api/ticks/query
```

**Response:**

```json
{
  "total": 2500000,
  "limit": 1000,
  "offset": 0,
  "data": [
    {
      "timestamp": "2026-01-20T15:30:45.123Z",
      "broker_id": "default",
      "symbol": "EURUSD",
      "bid": 1.08456,
      "ask": 1.08458,
      "spread": 0.00002,
      "lp": "YOFX1"
    }
  ]
}
```

**Error Responses:**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid request body |
| `401 Unauthorized` | Invalid or missing JWT token |
| `500 Internal Server Error` | Database query failed |

---

### 3. Get Tick Statistics

Get storage statistics for all symbols or a specific symbol.

**Endpoint:** `GET /api/ticks/stats`

**Query Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `symbol` | string | No | Filter by symbol (omit for all) | `EURUSD` |

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  https://api.rtx.com/api/ticks/stats?symbol=EURUSD
```

**Response:**

```json
{
  "total_ticks": 125000000,
  "total_symbols": 128,
  "oldest_tick": "2025-07-20T00:00:00Z",
  "newest_tick": "2026-01-20T15:30:45Z",
  "storage_size": "12500 MB",
  "ticks_received": 250000000,
  "ticks_written": 125000000,
  "ticks_dropped": 0
}
```

---

### 4. Get Available Symbols

Get list of all symbols with tick data.

**Endpoint:** `GET /api/ticks/symbols`

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  https://api.rtx.com/api/ticks/symbols
```

**Response:**

```json
{
  "symbols": [
    "EURUSD",
    "GBPUSD",
    "USDJPY",
    "AUDUSD",
    "USDCAD"
  ],
  "count": 128
}
```

---

### 5. Export OHLC Data

Export OHLC (candlestick) data for backtesting.

**Endpoint:** `GET /api/ticks/export-ohlc`

**Query Parameters:**

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `symbol` | string | Yes | Trading symbol | `EURUSD` |
| `timeframe` | string | Yes | Timeframe (`1m`, `5m`, `15m`, `1h`, `4h`, `1d`) | `1h` |
| `start_date` | ISO8601 | Yes | Start timestamp (UTC) | `2026-01-01T00:00:00Z` |
| `end_date` | ISO8601 | Yes | End timestamp (UTC) | `2026-01-20T23:59:59Z` |

**Example Request:**

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.rtx.com/api/ticks/export-ohlc?symbol=EURUSD&timeframe=1h&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z" \
  --output eurusd_1h_ohlc.csv
```

**Response (CSV):**

```csv
timestamp,open,high,low,close,volume
2026-01-20T15:00:00Z,1.08450000,1.08475000,1.08440000,1.08460000,1250
2026-01-20T16:00:00Z,1.08460000,1.08490000,1.08455000,1.08480000,1350
```

---

### 6. Cleanup Old Ticks (Admin Only)

Delete ticks older than a specified date.

**Endpoint:** `POST /admin/ticks/cleanup`

**Authentication:** Admin JWT token required

**Request Body:**

```json
{
  "older_than": "2025-07-01T00:00:00Z",
  "symbol": "EURUSD"  // optional, omit to clean all symbols
}
```

**Example Request:**

```bash
curl -X POST \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"older_than": "2025-07-01T00:00:00Z"}' \
  https://api.rtx.com/admin/ticks/cleanup
```

**Response:**

```json
{
  "success": true,
  "ticks_deleted": 25000000,
  "storage_freed_mb": 1200
}
```

**Error Responses:**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid request body |
| `401 Unauthorized` | Invalid or missing admin token |
| `403 Forbidden` | User does not have admin privileges |
| `500 Internal Server Error` | Cleanup failed |

---

## Authentication

All endpoints require a JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

**Obtain a token:**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username": "trader@example.com", "password": "your_password"}' \
  https://api.rtx.com/api/login
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user_123",
    "email": "trader@example.com",
    "role": "TRADER"
  }
}
```

---

## Rate Limiting

| Endpoint | Rate Limit |
|----------|------------|
| `/api/ticks/download` | 10 downloads per hour |
| `/api/ticks/query` | 100 requests per minute |
| `/api/ticks/stats` | 60 requests per minute |
| `/api/ticks/symbols` | 60 requests per minute |
| `/api/ticks/export-ohlc` | 30 downloads per hour |
| `/admin/ticks/cleanup` | 1 request per day |

**Rate Limit Headers:**

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 1642708800
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional details about the error"
}
```

**Common Error Codes:**

| Code | Description |
|------|-------------|
| `MISSING_PARAMETER` | Required query parameter missing |
| `INVALID_DATE_FORMAT` | Date not in RFC3339 format |
| `SYMBOL_NOT_FOUND` | Symbol does not exist in database |
| `UNAUTHORIZED` | Invalid or missing JWT token |
| `FORBIDDEN` | Insufficient privileges |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded |
| `DATABASE_ERROR` | Internal database error |

---

## Best Practices

### 1. Batch Downloads

For large date ranges, split into multiple smaller requests to avoid timeouts:

```bash
# Download 1 month at a time
for month in {01..06}; do
  curl -H "Authorization: Bearer <token>" \
    "https://api.rtx.com/api/ticks/download?symbol=EURUSD&start_date=2026-${month}-01T00:00:00Z&end_date=2026-${month}-31T23:59:59Z&format=csv&compression=gzip" \
    --output "eurusd_2026-${month}.csv.gz"
done
```

### 2. Use Compression

Always use `compression=gzip` for downloads to reduce bandwidth:

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.rtx.com/api/ticks/download?symbol=EURUSD&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z&format=csv&compression=gzip" \
  --output eurusd_jan2026.csv.gz

# Decompress locally
gunzip eurusd_jan2026.csv.gz
```

### 3. Use OHLC for Backtesting

For backtesting strategies, use OHLC data instead of raw ticks to reduce data size:

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.rtx.com/api/ticks/export-ohlc?symbol=EURUSD&timeframe=1h&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z" \
  --output eurusd_1h_ohlc.csv
```

### 4. Monitor Rate Limits

Check response headers to avoid hitting rate limits:

```bash
curl -I -H "Authorization: Bearer <token>" \
  "https://api.rtx.com/api/ticks/stats"

# Check headers:
# X-RateLimit-Remaining: 58
```

---

## SDK Examples

### Python

```python
import requests
import gzip
import csv
from datetime import datetime

def download_ticks(symbol, start_date, end_date, token):
    url = "https://api.rtx.com/api/ticks/download"
    params = {
        "symbol": symbol,
        "start_date": start_date.isoformat() + "Z",
        "end_date": end_date.isoformat() + "Z",
        "format": "csv",
        "compression": "gzip"
    }
    headers = {"Authorization": f"Bearer {token}"}

    response = requests.get(url, params=params, headers=headers, stream=True)
    response.raise_for_status()

    # Decompress and parse CSV
    with gzip.open(response.raw, 'rt') as f:
        reader = csv.DictReader(f)
        ticks = list(reader)

    return ticks

# Usage
token = "your_jwt_token"
ticks = download_ticks("EURUSD", datetime(2026, 1, 1), datetime(2026, 1, 20), token)
print(f"Downloaded {len(ticks)} ticks")
```

### JavaScript (Node.js)

```javascript
const axios = require('axios');
const zlib = require('zlib');
const csv = require('csv-parser');

async function downloadTicks(symbol, startDate, endDate, token) {
    const url = 'https://api.rtx.com/api/ticks/download';
    const params = {
        symbol: symbol,
        start_date: startDate.toISOString(),
        end_date: endDate.toISOString(),
        format: 'csv',
        compression: 'gzip'
    };

    const response = await axios.get(url, {
        params: params,
        headers: { 'Authorization': `Bearer ${token}` },
        responseType: 'stream'
    });

    const ticks = [];
    return new Promise((resolve, reject) => {
        response.data
            .pipe(zlib.createGunzip())
            .pipe(csv())
            .on('data', (tick) => ticks.push(tick))
            .on('end', () => resolve(ticks))
            .on('error', reject);
    });
}

// Usage
const token = 'your_jwt_token';
downloadTicks('EURUSD', new Date('2026-01-01'), new Date('2026-01-20'), token)
    .then(ticks => console.log(`Downloaded ${ticks.length} ticks`));
```

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-20 | Initial release with TimescaleDB backend |

---

## Support

For API support, contact: api-support@rtx.com

For technical documentation: https://docs.rtx.com/tick-storage
