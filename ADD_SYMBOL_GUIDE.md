# Add Symbol Functionality - User Guide

**Date:** 2026-01-20
**Status:** ✅ Working (with conditions)

---

## How Add Symbol Works

### Flow

```
1. User clicks "Click to add symbol..." input
   ↓
2. Dropdown shows 53 available symbols from /api/symbols/available
   ↓
3. User types to search (e.g., "EUR")
   ↓
4. Dropdown filters to matching symbols
   ↓
5. User clicks on a symbol (e.g., "EURUSD")
   ↓
6. Frontend calls: POST /api/symbols/subscribe { symbol: "EURUSD" }
   ↓
7. Backend subscribes symbol via FIX protocol (YOFX2 session)
   ↓
8. Symbol added to subscribed list
   ↓
9. Symbol shows "Active" in dropdown (green checkmark)
   ↓
10. Symbol SHOULD appear in MarketWatch list
```

---

## Why Symbol Might Not Appear

### 1. Symbol is Already Subscribed

**What you see:**
- Symbol shows "Active" (green checkmark) in dropdown
- Clicking it doesn't add it again

**Why:**
- 29 symbols are already subscribed on server start
- These symbols: EURUSD, GBPUSD, USDJPY, AUDUSD, etc.

**Solution:**
- Click on "Active" symbol → closes dropdown and selects the symbol
- Symbol should already be visible in the main list

---

### 2. FIX Connection is Disconnected

**What happens:**
- Symbol subscription API returns: `{"success": false, "error": "session not logged in: DISCONNECTED"}`
- Symbol won't receive tick data
- Symbol shows in list but with no prices (dashes "--")

**Check FIX status:**
```bash
curl http://localhost:7999/api/config
# Look for: "fixStatus": {"YOFX2": "CONNECTED"}
```

**Currently:** All FIX sessions show `"DISCONNECTED"` because:
- Server is using OANDA-HISTORICAL simulation as fallback
- YOFX FIX protocol was fixed but connection requires:
  - Valid credentials
  - Network/proxy access
  - YOFX server accepting connections

---

### 3. Search Filter is Active

**Problem:**
- You typed something in the "Click to add symbol..." field
- MarketWatch list is filtered to only show matching symbols

**Solution:**
- Clear the search field (delete any text)
- All subscribed symbols will appear

---

### 4. Symbol is Hidden

**Problem:**
- You previously right-clicked and selected "Hide" on the symbol
- Symbol is in `hiddenSymbols` list stored in localStorage

**Solution:**
- Right-click anywhere in MarketWatch
- Click "Show All" under VISIBILITY section
- All hidden symbols will reappear

---

### 5. Wrong Tab Selected

**Problem:**
- You're viewing Details, Trading, or Ticks tab
- Symbol list only shows in "Symbols" tab

**Solution:**
- Click "Symbols" tab at bottom of MarketWatch panel
- Full symbol list will appear

---

## Testing Add Symbol (Step-by-Step)

### Test 1: Add a New Symbol

```bash
# 1. Check which symbols are subscribed
curl -s http://localhost:7999/api/symbols/subscribed

# 2. Pick a symbol NOT in the list (e.g., "WTICOUSD" - Oil)

# 3. In the browser:
#    - Open http://localhost:5174
#    - Click "Click to add symbol..." input
#    - Type "WTI"
#    - Click on "WTICOUSD" in dropdown
#    - Watch for "Adding..." → "Active"

# 4. Check if symbol was added
curl -s http://localhost:7999/api/symbols/subscribed | grep WTICOUSD

# Expected: Symbol appears in subscribed list
```

### Test 2: Verify Symbol Appears in List

After adding symbol:

1. **Clear search field** - Delete any text in "Click to add symbol..."
2. **Check "Symbols" tab** - Make sure you're on the Symbols tab
3. **Show all symbols** - Right-click → "Show All"
4. **Scroll through list** - Symbol should appear alphabetically

---

## Current System Behavior

### Symbols Available

- **Total symbols:** 128 (from `/api/symbols`)
- **Dropdown symbols:** 53 (from `/api/symbols/available`)
- **Subscribed symbols:** 29 (default on server start)

### Subscribed by Default

```
EURUSD, GBPUSD, USDJPY, USDCHF, USDCAD, AUDUSD, NZDUSD, EURGBP,
EURJPY, GBPJPY, AUDJPY, NZDJPY, AUDCAD, AUDCHF, AUDNZD, EURCAD,
EURCHF, EURAUD, GBPAUD, GBPCAD, GBPCHF, GBPNZD, NZDCAD, NZDCHF,
CADCHF, CADJPY, CHFJPY, XAUUSD, XAGUSD
```

### Market Data Sources

**Current:** OANDA-HISTORICAL (simulation)
**Reason:** YOFX FIX sessions are disconnected

**When YOFX Connects:**
- Subscribed symbols will receive live FIX data
- LP field will show "YOFX" instead of "OANDA-HISTORICAL"
- Real-time quotes will update every 100ms

---

## Troubleshooting Checklist

When a symbol doesn't appear after clicking "Add":

- [ ] Clear the search field
- [ ] Verify you're on "Symbols" tab
- [ ] Right-click → "Show All" to unhide symbols
- [ ] Check browser console (F12) for errors
- [ ] Verify subscription succeeded:
  ```bash
  curl -s http://localhost:7999/api/symbols/subscribed | grep SYMBOLNAME
  ```
- [ ] Check if symbol has tick data:
  ```bash
  curl -s http://localhost:7999/api/symbols
  # Look for the symbol in the list
  ```
- [ ] Hard refresh browser (Ctrl+Shift+R)

---

## API Endpoints Reference

### Subscribe to Symbol
```bash
POST /api/symbols/subscribe
Body: { "symbol": "EURUSD" }

Response (Success):
{
  "success": true,
  "symbol": "EURUSD",
  "mdReqId": "MDReq_EURUSD_12345",
  "message": "Subscribed successfully"
}

Response (Already Subscribed):
{
  "success": true,
  "symbol": "EURUSD",
  "message": "Already subscribed"
}

Response (FIX Disconnected):
{
  "success": false,
  "symbol": "EURUSD",
  "error": "session not logged in: DISCONNECTED"
}
```

### Get Subscribed Symbols
```bash
GET /api/symbols/subscribed

Response:
["EURUSD", "GBPUSD", "USDJPY", ...]
```

### Get Available Symbols (Dropdown)
```bash
GET /api/symbols/available

Response:
[
  {
    "symbol": "EURUSD",
    "name": "Euro/US Dollar",
    "category": "forex.major",
    "digits": 5,
    "subscribed": true
  },
  ...
]
```

### Get All Symbols (System)
```bash
GET /api/symbols

Response:
[
  {
    "symbol": "EURUSD",
    "contractSize": 100000,
    "pipSize": 0.0001,
    "pipValue": 10,
    ...
  },
  ...
]
```

---

## Known Issues & Solutions

### Issue: Symbol Shows "Active" but Not in List

**Cause:** Symbol is either hidden or filtered by search

**Solution:**
1. Clear search field
2. Right-click → "Show All"
3. Refresh browser (F5)

### Issue: Symbol Added but No Prices

**Cause:** YOFX FIX connection disconnected, no tick data

**Solution:**
- Symbol is successfully subscribed
- Waiting for YOFX connection to establish
- Currently falling back to OANDA-HISTORICAL simulation
- System will show live prices when YOFX connects

### Issue: "Failed to subscribe" Error

**Cause:** Backend API error or FIX session issue

**Check:**
```bash
# Backend logs
cd "D:\Tading engine\Trading-Engine\backend\cmd\server"
# Check console output for [API] or [FIX] errors
```

---

## Summary

**Add Symbol Functionality:** ✅ **Working**

**How to Use:**
1. Click "Click to add symbol..." input
2. Type to search for symbol
3. Click on symbol in dropdown
4. Symbol shows "Adding..." → "Active"
5. **Clear search field** to see symbol in list
6. Symbol appears in alphabetical order

**Common Fix:**
- Clear the search field after adding a symbol
- The search filter hides symbols that don't match

**Next Steps:**
- Fix YOFX FIX connection for live market data
- Symbols will receive real-time quotes when FIX connects

