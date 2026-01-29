# Testing Trading Engine APIs in Postman

## Quick Start Guide

### Base URL
```
http://localhost:7999
```

## 📊 1. Test OHLC/Candle API (Main One)

### Request Configuration:
- **Method:** `GET`
- **URL:** `http://localhost:7999/api/history/ohlc`
- **Params Tab:**
  | KEY       | VALUE  | DESCRIPTION           |
  | --------- | ------ | --------------------- |
  | symbol    | XRPUSD | Trading pair          |
  | timeframe | 1m     | 1m, 5m, 15m, 1h, etc. |
  | limit     | 100    | Number of candles     |

### Example URLs:
```
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit=100
http://localhost:7999/api/history/ohlc?symbol=EURUSD&timeframe=1h&limit=500
http://localhost:7999/api/history/ohlc?symbol=BTCUSD&timeframe=5m&limit=200
```

---

## 🎯 2. Test Tick Data API

### Request Configuration:
- **Method:** `GET`
- **URL:** `http://localhost:7999/api/history/ticks`
- **Params Tab:**
  | KEY    | VALUE      |
  | ------ | ---------- |
  | symbol | XRPUSD     |
  | date   | 2026-01-27 |
  | limit  | 1000       |

### Example URL:
```
http://localhost:7999/api/history/ticks?symbol=XRPUSD&date=2026-01-27&limit=1000
```

---

## 📋 3. Test Get Symbols API

### Request Configuration:
- **Method:** `GET`
- **URL:** `http://localhost:7999/api/history/symbols`
- **Params:** None needed

### Example URL:
```
http://localhost:7999/api/history/symbols
```

---

## ℹ️ 4. Test Symbol Info API

### Request Configuration:
- **Method:** `GET`
- **URL:** `http://localhost:7999/api/history/info`
- **Params Tab:**
  | KEY    | VALUE  |
  | ------ | ------ |
  | symbol | XRPUSD |

### Example URL:
```
http://localhost:7999/api/history/info?symbol=XRPUSD
```

---

## 💼 5. Test Bulk Download (POST Request)

### Request Configuration:
- **Method:** `POST`
- **URL:** `http://localhost:7999/api/history/ticks/bulk`
- **Headers Tab:**
  | KEY          | VALUE            |
  | ------------ | ---------------- |
  | Content-Type | application/json |

- **Body Tab:** (Select "raw" and "JSON")
```json
{
  "symbols": ["EURUSD", "GBPUSD", "XRPUSD"],
  "from": "2026-01-20T00:00:00Z",
  "to": "2026-01-27T23:59:59Z",
  "format": "json"
}
```

---

## 🚀 Step-by-Step Instructions

### For GET Requests (Most Common):

1. **Open Postman**
2. **Create New Request**
   - Click "New" → "HTTP Request"
3. **Set Method to GET**
4. **Enter URL:** `http://localhost:7999/api/history/ohlc`
5. **Add Parameters:**
   - Click "Params" tab below the URL
   - Add: `symbol` = `XRPUSD`
   - Add: `timeframe` = `1m`
   - Add: `limit` = `100`
6. **Click "Send"**
7. **View Response** in the bottom panel

### For POST Requests (Bulk Download):

1. **Create New Request**
2. **Set Method to POST**
3. **Enter URL:** `http://localhost:7999/api/history/ticks/bulk`
4. **Go to "Headers" tab:**
   - Add: `Content-Type` = `application/json`
5. **Go to "Body" tab:**
   - Select "raw"
   - Select "JSON" from dropdown
   - Paste the JSON request body
6. **Click "Send"**
7. **View Response**

---

## 📦 Postman Collection (Ready to Import)

Save this as `trading-engine.postman_collection.json`:

```json
{
  "info": {
    "name": "Trading Engine API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Get OHLC Candles",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit=100",
          "protocol": "http",
          "host": ["localhost"],
          "port": "7999",
          "path": ["api", "history", "ohlc"],
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "1m"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "Get Tick Data",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "http://localhost:7999/api/history/ticks?symbol=XRPUSD&date=2026-01-27&limit=1000",
          "protocol": "http",
          "host": ["localhost"],
          "port": "7999",
          "path": ["api", "history", "ticks"],
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "date", "value": "2026-01-27"},
            {"key": "limit", "value": "1000"}
          ]
        }
      }
    },
    {
      "name": "Get All Symbols",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "http://localhost:7999/api/history/symbols",
          "protocol": "http",
          "host": ["localhost"],
          "port": "7999",
          "path": ["api", "history", "symbols"]
        }
      }
    },
    {
      "name": "Get Symbol Info",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "http://localhost:7999/api/history/info?symbol=XRPUSD",
          "protocol": "http",
          "host": ["localhost"],
          "port": "7999",
          "path": ["api", "history", "info"],
          "query": [
            {"key": "symbol", "value": "XRPUSD"}
          ]
        }
      }
    },
    {
      "name": "Bulk Download Ticks",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"symbols\": [\"EURUSD\", \"GBPUSD\", \"XRPUSD\"],\n  \"from\": \"2026-01-20T00:00:00Z\",\n  \"to\": \"2026-01-27T23:59:59Z\",\n  \"format\": \"json\"\n}"
        },
        "url": {
          "raw": "http://localhost:7999/api/history/ticks/bulk",
          "protocol": "http",
          "host": ["localhost"],
          "port": "7999",
          "path": ["api", "history", "ticks", "bulk"]
        }
      }
    }
  ]
}
```

---

## 💡 Quick Test Checklist

- [ ] **Test 1:** Get OHLC for XRPUSD (1-minute, 100 candles)
- [ ] **Test 2:** Get OHLC for EURUSD (1-hour, 500 candles)
- [ ] **Test 3:** Get all available symbols
- [ ] **Test 4:** Get symbol info for BTCUSD
- [ ] **Test 5:** Get tick data for today
- [ ] **Test 6:** Bulk download multiple symbols

---

## ✅ Expected Responses

### OHLC Response:
```json
[
  {
    "time": 1706365200,
    "open": 0.52450,
    "high": 0.52480,
    "low": 0.52440,
    "close": 0.52470,
    "volume": 12500
  }
]
```

### Symbols Response:
```json
{
  "symbols": [
    {
      "symbol": "EURUSD",
      "display_name": "EURUSD",
      "category": "forex",
      "available": true,
      "tick_count": 125000
    }
  ],
  "total": 35
}
```

---

## 🔧 Troubleshooting

**If you get "Could not get response":**
- ✅ Make sure your backend is running: `go run cmd/server/main.go`
- ✅ Check the URL is `http://localhost:7999` (not https)
- ✅ Verify the server is running on port 7999

**If you get empty data:**
- ✅ The symbol might not have data yet
- ✅ Try `/api/history/symbols` first to see available symbols
- ✅ Check if your FIX connection is streaming data

---

## 🎯 Pro Tips

1. **Save requests** in a collection for reuse
2. **Use variables** for base URL: `{{baseUrl}}/api/history/ohlc`
3. **Test different symbols** to see what data is available
4. **Start with `/api/history/symbols`** to see what's available
5. **Use "Beautify"** button in Postman to format JSON responses
