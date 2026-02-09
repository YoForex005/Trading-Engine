# Trading Engine History APIs - Complete Reference

Your trading engine has **7 main APIs** for accessing candle/pricing history data:

## 📊 OHLC (Candle) APIs

### 1. **Get OHLC Candles** ⭐ (Main Candle API)
```
GET /api/history/ohlc?symbol={SYMBOL}&timeframe={TF}&limit={N}
```

**Parameters:**
- `symbol` - Trading pair (e.g., `XRPUSD`, `EURUSD`)
- `timeframe` - Candle timeframe: `1m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`
- `limit` - Number of candles (default: 1000, max: 50000)

**Example:**
```bash
curl "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit=100"
```

**Response:**
```json
[
  {
    "time": 1706365200,
    "open": 0.52450,
    "high": 0.52480,
    "low": 0.52440,
    "close": 0.52470,
    "volume": 12500
  },
  ...
]
```

---

## 🎯 Tick (Raw Price) APIs

### 2. **Get Tick Data by Symbol Path**
```
GET /api/history/ticks/{SYMBOL}?from={DATE}&to={DATE}&page={N}&page_size={N}
```

**Parameters:**
- `{SYMBOL}` - Path parameter (e.g., `/api/history/ticks/EURUSD`)
- `from` - Start date (RFC3339 format, default: 7 days ago)
- `to` - End date (RFC3339 format, default: now)
- `page` - Page number (default: 1)
- `page_size` - Items per page (default: 1000, max: 10000)
- `format` - Response format: `json`, `csv`, `binary`

**Example:**
```bash
curl "http://localhost:7999/api/history/ticks/EURUSD?page=1&page_size=1000"
```

### 3. **Get Tick Data by Query Parameters**
```
GET /api/history/ticks?symbol={SYMBOL}&date={DATE}&offset={N}&limit={N}
```

**Parameters:**
- `symbol` - Trading pair
- `date` - Specific date (format: `YYYY-MM-DD`, default: today)
- `offset` - Starting position (default: 0)
- `limit` - Max ticks (default: 5000, max: 50000)

**Example:**
```bash
curl "http://localhost:7999/api/history/ticks?symbol=XRPUSD&date=2026-01-27&limit=5000"
```

**Response:**
```json
{
  "symbol": "XRPUSD",
  "date": "2026-01-27",
  "ticks": [
    {
      "timestamp": 1706365200000,
      "bid": 0.52450,
      "ask": 0.52452,
      "spread": 0.00002
    },
    ...
  ],
  "total": 125000,
  "offset": 0,
  "limit": 5000
}
```

### 4. **Bulk Tick Download (Multiple Symbols)**
```
POST /api/history/ticks/bulk
```

**Request Body:**
```json
{
  "symbols": ["EURUSD", "GBPUSD", "XRPUSD"],
  "from": "2026-01-20T00:00:00Z",
  "to": "2026-01-27T23:59:59Z",
  "format": "json"
}
```

**Response:** Gzipped JSON with all ticks for all symbols
- Max 50 symbols per request
- Automatically compressed with gzip

---

## ℹ️ Metadata APIs

### 5. **Get Available Symbols**
```
GET /api/history/symbols
```

Returns list of all symbols with historical data available.

**Response:**
```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "display_name": "EURUSD",
      "category": "forex",
      "available": true,
      "tick_count": 125000,
      "last_updated": "2026-01-27T14:00:00Z"
    },
    ...
  ],
  "total": 35
}
```

### 6. **Get Symbol Info**
```
GET /api/history/info?symbol={SYMBOL}
```

Returns metadata about a specific symbol's historical data.

**Example:**
```bash
curl "http://localhost:7999/api/history/info?symbol=XRPUSD"
```

**Response:**
```json
{
  "symbol": "XRPUSD",
  "earliestDate": "2026-01-20",
  "latestDate": "2026-01-27",
  "tickCount": 125000,
  "availableDays": 7,
  "lastUpdated": "2026-01-27T14:00:00Z"
}
```

### 7. **Get Available Data Overview**
```
GET /api/history/available
```

Returns overview of all symbols with available historical data.

**Response:**
```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "earliest_tick": "2026-01-20T00:00:00Z",
      "latest_tick": "2026-01-27T14:00:00Z",
      "total_ticks": 125000,
      "available_days": 7,
      "last_updated": "2026-01-27T14:00:00Z"
    },
    ...
  ],
  "total": 35
}
```

---

## 🔧 Admin API

### **Backfill Historical Data**
```
POST /admin/history/backfill
```

**Request Body:**
```json
{
  "symbol": "EURUSD",
  "ticks": [
    {
      "timestamp": "2026-01-20T00:00:00Z",
      "symbol": "EURUSD",
      "bid": 1.08450,
      "ask": 1.08452,
      "spread": 0.00002,
      "lp": "YOFX"
    },
    ...
  ],
  "source": "external_provider"
}
```

---

## 📝 Quick Examples

### Get Last 100 1-minute Candles for XRPUSD
```bash
curl "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit=100"
```

### Get Last 500 1-hour Candles for EURUSD
```bash
curl "http://localhost:7999/api/history/ohlc?symbol=EURUSD&timeframe=1h&limit=500"
```

### Get Today's Tick Data for BTCUSD
```bash
curl "http://localhost:7999/api/history/ticks?symbol=BTCUSD&date=2026-01-27&limit=10000"
```

### Get All Available Symbols
```bash
curl "http://localhost:7999/api/history/symbols"
```

---

## 💡 Key Features

✅ **7 History APIs** total (1 OHLC + 3 Tick + 3 Metadata + 1 Admin)  
✅ **OHLC/Candle Data** - Multiple timeframes (1m, 5m, 15m, 30m, 1h, 4h, 1d)  
✅ **Raw Tick Data** - Bid/Ask/Spread with millisecond precision  
✅ **Pagination** - Efficient data retrieval for large datasets  
✅ **Multiple Formats** - JSON, CSV, Binary  
✅ **Compression** - Gzip support for large  downloads  
✅ **Rate Limiting** - Protects server from overload  
✅ **CORS Enabled** - Works with frontend applications  

---

## 🚀 Currently Running

Based on your terminal, you already have an OHLC request running:
```bash
curl "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit..."
```

This is the main API you should use for chart candles! 📈
