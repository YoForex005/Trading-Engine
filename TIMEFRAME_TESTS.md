# Testing All Timeframes - Postman Quick Reference

## ✅ Complete Timeframe Support

Your trading engine now supports **ALL 9 timeframes** from your UI:

| UI Label | API Value      | Seconds   | Description |
| -------- | -------------- | --------- | ----------- |
| M1       | `1m` or `m1`   | 60        | 1 Minute    |
| M5       | `5m` or `m5`   | 300       | 5 Minutes   |
| M15      | `15m` or `m15` | 900       | 15 Minutes  |
| M30      | `30m` or `m30` | 1,800     | 30 Minutes  |
| H1       | `1h` or `h1`   | 3,600     | 1 Hour      |
| H4       | `4h` or `h4`   | 14,400    | 4 Hours     |
| D1       | `1d` or `d1`   | 86,400    | 1 Day       |
| W1       | `1w` or `w1`   | 604,800   | 1 Week      |
| MN       | `mn` or `1mo`  | 2,592,000 | 1 Month     |

---

## 🚀 Test All Timeframes in Postman

### Quick Test URLs (Copy & Paste):

```bash
# M1 - 1 Minute
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit=100

# M5 - 5 Minutes
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=5m&limit=100

# M15 - 15 Minutes
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=15m&limit=100

# M30 - 30 Minutes
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=30m&limit=100

# H1 - 1 Hour
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1h&limit=100

# H4 - 4 Hours
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=4h&limit=100

# D1 - 1 Day
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1d&limit=100

# W1 - 1 Week ⭐ NEW
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=w1&limit=100

# MN - Monthly ⭐ NEW
http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=mn&limit=100
```

---

## 📋 Complete Postman Collection (Import This)

```json
{
  "info": {
    "name": "All Timeframes Test",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "M1 - 1 Minute",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1m&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "1m"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "M5 - 5 Minutes",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=5m&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "5m"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "M15 - 15 Minutes",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=15m&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "15m"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "M30 - 30 Minutes",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=30m&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "30m"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "H1 - 1 Hour",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1h&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "1h"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "H4 - 4 Hours",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=4h&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "4h"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "D1 - 1 Day",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=1d&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "1d"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "W1 - 1 Week",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=w1&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "w1"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    },
    {
      "name": "MN - Monthly",
      "request": {
        "method": "GET",
        "url": {
          "raw": "http://localhost:7999/api/history/ohlc?symbol=XRPUSD&timeframe=mn&limit=100",
          "query": [
            {"key": "symbol", "value": "XRPUSD"},
            {"key": "timeframe", "value": "mn"},
            {"key": "limit", "value": "100"}
          ]
        }
      }
    }
  ]
}
```

---

## ✅ Test Checklist

Test each timeframe from your UI:

- [ ] **M1** (1 minute) - Should return 100 candles
- [ ] **M5** (5 minutes) - Should return 100 candles
- [ ] **M15** (15 minutes) - Should return 100 candles
- [ ] **M30** (30 minutes) - Should return 100 candles
- [ ] **H1** (1 hour) - Should return 100 candles
- [ ] **H4** (4 hours) - Should return 100 candles
- [ ] **D1** (1 day) - Should return 100 candles
- [ ] **W1** (1 week) - Should return 100 candles ⭐ NEW
- [ ] **MN** (monthly) - Should return 100 candles ⭐ NEW

---

## 🔄 Alternative Naming Supported

Both formats work (for compatibility):

```bash
# Standard format
timeframe=1m  timeframe=5m  timeframe=15m  timeframe=30m
timeframe=1h  timeframe=4h  timeframe=1d   timeframe=1w  timeframe=mn

# MT4/MT5 style format
timeframe=m1  timeframe=m5  timeframe=m15  timeframe=m30
timeframe=h1  timeframe=h4  timeframe=d1   timeframe=w1  timeframe=mn
```

Both will work! The backend now accepts both formats.

---

## 💡 Expected Response Format

All timeframes return the same format:

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
  {
    "time": 1706365260,
    "open": 0.52470,
    "high": 0.52490,
    "low": 0.52460,
    "close": 0.52480,
    "volume": 13200
  }
]
```

---

## 🎯 What Changed

✅ **Added W1 (Weekly)** support - 604,800 seconds  
✅ **Added MN (Monthly)** support - 2,592,000 seconds  
✅ **Added alternative naming** - Both `1m` and `m1` work now  
✅ **Backend now matches your UI** - All 9 timeframes supported

---

## 🚨 Important Notes

1. **Restart your backend** after the changes:
   ```bash
   # Stop current server (Ctrl+C)
   # Restart
   go run cmd/server/main.go
   ```

2. **Weekly/Monthly candles** may have less data depending on your history depth

3. **Monthly approximation**: Uses 30 days (2.592M seconds) as average month

---

## 🎨 Your Timeframe Selector

![Timeframe Selector](file:///C:/Users/ADMIN/.gemini/antigravity/brain/4dd69b2a-4ad7-45cd-a285-b57a75f42e42/uploaded_media_1769504895044.png)

All buttons (M1, M5, M15, M30, H1, H4, D1, W1, MN) will now work correctly with the API! ✅
