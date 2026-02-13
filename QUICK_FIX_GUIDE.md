# Trading Engine - QUICK FIX GUIDE

> **⚡ Start Here** - 5-Minute Overview of What to Fix

---

## 🚨 Critical Issues (Fix First)

### Issue #1: Broken API Endpoint Paths
**File:** `clients/desktop/src/services/api.ts`
**Lines:** 353, 369, 386, 425, 438, 395
**Time:** 2 hours
**Impact:** Limit/stop orders fail silently

**Quick Fix:**
```typescript
// Line 353 - placeLimitOrder
- `${API_ENDPOINTS.order}/limit`
+ `${API_ENDPOINTS.orders}/limit`

// Line 369 - placeStopOrder
- `${API_ENDPOINTS.order}/stop`
+ `${API_ENDPOINTS.orders}/stop`

// Line 386 - placeStopLimitOrder
- `${API_ENDPOINTS.order}/stop-limit`
+ `${API_ENDPOINTS.orders}/stop-limit`

// Line 438 - getOHLC
- `${API_ENDPOINTS.ohlc}?...`
+ `${API_ENDPOINTS.candles}?...`
```

---

### Issue #2: Missing Risk Calculator
**Backend:** `/risk/calculate-lot` endpoint doesn't exist
**Impact:** Position sizing broken
**Time:** 6-8 hours

**Quick Fix:**
1. Create `backend/api/handlers/risk_calculator.go`
2. Add this code:
```go
func (h *APIHandler) HandleCalculateLot(w http.ResponseWriter, r *http.Request) {
    symbol := r.URL.Query().Get("symbol")
    riskPercent, _ := strconv.ParseFloat(r.URL.Query().Get("riskPercent"), 64)
    slPips, _ := strconv.ParseFloat(r.URL.Query().Get("slPips"), 64)
    balance, _ := strconv.ParseFloat(r.URL.Query().Get("balance"), 64)

    pipValue := 0.0001 * 100000
    riskAmount := balance * (riskPercent / 100.0)
    recommendedLot := riskAmount / (slPips * pipValue)

    json.NewEncoder(w).Encode(map[string]float64{
        "recommendedLot": recommendedLot,
        "riskAmount": riskAmount,
    })
}
```
3. Register in `main.go`: `http.HandleFunc("/risk/calculate-lot", apiHandler.HandleCalculateLot)`

---

### Issue #3: WebSocket Missing Real-Time Updates
**File:** `clients/desktop/src/services/websocket-enhanced.ts`
**Impact:** Position P&L doesn't update until page refresh
**Time:** 4-6 hours

**Quick Fix:**
Add to `handleMessage()` method:
```typescript
switch (message.type) {
  case 'position_update':
    const store = useAppStore.getState();
    store.updatePosition(message.data.positionId, {
      currentPrice: message.data.currentPrice,
      pnl: message.data.pnl
    });
    break;

  case 'order_fill':
    useNotificationStore.getState().addNotification({
      type: 'success',
      title: 'Order Filled',
      message: `${message.data.side} ${message.data.volume} ${message.data.symbol}`
    });
    break;
}
```

---

### Issue #4: WebSocket Doesn't Re-subscribe
**File:** `clients/desktop/src/services/websocket-enhanced.ts`
**Impact:** Users lose market data after reconnect
**Time:** 2-3 hours

**Quick Fix:**
Add to class:
```typescript
private activeSubscriptions: Set<string> = new Set();

subscribe(channel: string, params?: any) {
  this.activeSubscriptions.add(`${channel}:${JSON.stringify(params || {})}`);
  // ... send subscribe message
}

private onOpen = () => {
  // Re-subscribe to all channels
  this.activeSubscriptions.forEach(sub => {
    const [channel, paramsStr] = sub.split(':');
    this.ws?.send(JSON.stringify({
      action: 'subscribe',
      channel,
      params: JSON.parse(paramsStr)
    }));
  });
}
```

---

## 🔥 High Priority Issues (Do Next)

### Issue #5: Symbol Screener Uses Fake Data
**File:** `clients/desktop/src/store/useScreenerStore.ts`
**Impact:** Shows wrong RSI/MA values
**Time:** 1 day

**Quick Fix:** Remove `generateMockSymbols()`, replace with:
```typescript
fetchScreenerData: async () => {
  const symbols = await api.getSymbols();
  const ticks = await Promise.all(
    symbols.data.map(s => api.getTicks(s.symbol, 100))
  );
  // Calculate real RSI/MA from ticks...
}
```

---

### Issue #6: News Feed Shows 17 Fake Articles
**File:** `clients/desktop/src/store/useNewsStore.ts`
**Impact:** Users see outdated/fake news
**Time:** 1 day

**Quick Fix:** Remove `generateMockNews()`, replace with:
```typescript
fetchNews: async () => {
  const response = await fetch(`${API_ENDPOINTS.marketNews.articles}`);
  const data = await response.json();
  set({ news: data.articles || [] });
}
```

---

### Issue #7: Can't Export Reports
**Backend:** `/api/analytics/export/pdf` doesn't exist
**Impact:** Users can't download trade reports
**Time:** 1 day

**Quick Fix:**
1. `go get github.com/jung-kurt/gofpdf`
2. Create `backend/api/handlers/export.go`
3. Add PDF/CSV generation handlers

---

## 📊 Statistics

| Metric | Count |
|--------|-------|
| **Total Issues** | 74 |
| **Critical (P0)** | 4 issues = **1-2 days** |
| **High Priority (P1)** | 5 issues = **3-5 days** |
| **Medium Priority (P2)** | 20+ issues = **3-5 days** |
| **Low Priority (P3)** | 45+ issues = **2-3 days** |

---

## 🎯 Recommended Path

### Path A: Minimum Fix (1-2 days)
✅ Fix 4 critical issues
✅ Trading platform works
❌ Still has mock data in analytics

**Good for:** Emergency fixes, MVP testing

---

### Path B: Full Feature Fix (4-7 days)
✅ Fix all P0 + P1 issues (9 total)
✅ All user-facing features use real data
✅ Exports and analytics work
❌ Still has some code quality issues

**Good for:** Production deployment, real users

---

### Path C: Production Ready (9-15 days)
✅ Fix all 74 issues
✅ Advanced order types (OCO, trailing stop)
✅ Clean code (no alerts, console.logs)
✅ Server-side persistence

**Good for:** Commercial product, long-term maintenance

---

## 📁 Files to Fix (Sorted by Priority)

### Priority 0 (Must Fix)
1. ✅ `clients/desktop/src/services/api.ts` - 6 broken paths
2. ❌ `backend/api/handlers/risk_calculator.go` - CREATE THIS
3. ✅ `clients/desktop/src/services/websocket-enhanced.ts` - Add handlers

### Priority 1 (Should Fix)
4. ✅ `clients/desktop/src/store/useScreenerStore.ts` - Remove mock
5. ✅ `clients/desktop/src/store/useNewsStore.ts` - Remove mock
6. ✅ `clients/desktop/src/store/useSignalStore.ts` - Remove mock
7. ❌ `backend/api/handlers/export.go` - CREATE THIS
8. ✅ `clients/desktop/src/components/OrderBook.tsx` - Remove mock

### Priority 2 (Nice to Have)
9. ❌ `backend/api/handlers/advanced_orders.go` - CREATE THIS
10-28. Various analytics components using mock data

### Priority 3 (Polish)
29-74. Code quality issues (alerts, console.logs, etc.)

---

## 🚀 How to Start (Right Now)

### Step 1: Open Your Editor
```bash
code clients/desktop/src/services/api.ts
```

### Step 2: Fix Line 353
```typescript
// Find this line:
return fetchWithTimeout(`${API_ENDPOINTS.order}/limit`, ...

// Change to:
return fetchWithTimeout(`${API_ENDPOINTS.orders}/limit`, ...
```

### Step 3: Repeat for Lines 369, 386, 438
Same pattern - fix URL paths

### Step 4: Test
```bash
cd clients/desktop
npm run dev
# Try placing a limit order - should work now!
```

### Step 5: Celebrate 🎉
You just fixed 20% of the critical issues in 5 minutes!

---

## 📚 Full Documentation

- **This Guide** - 5-minute overview
- **docs/FIX_PLAN_SUMMARY.md** - Executive summary (10 min read)
- **docs/IMPLEMENTATION_CHECKLIST.md** - Step-by-step with code (1 hour)
- **docs/MISSING_FUNCTIONALITY_FIX_PLAN.md** - Complete reference (3 hours)

---

## 💡 Pro Tips

1. **Start small**: Fix the 6 API paths first (2 hours)
2. **Test often**: After each fix, test in browser
3. **One file at a time**: Don't try to fix everything at once
4. **Use the checklist**: Follow docs/IMPLEMENTATION_CHECKLIST.md
5. **Ask for help**: If stuck, check the full plan document

---

## ⚡ Commands You'll Need

### Frontend
```bash
cd clients/desktop
npm run dev          # Start dev server
npm run typecheck    # Check TypeScript errors
npm run build        # Build for production
```

### Backend
```bash
cd backend
go build -o server.exe ./cmd/server/main.go   # Compile
./server.exe                                   # Run server
curl http://localhost:7999/health              # Test if running
```

---

## 🎯 Success = 4 Tests Pass

After fixing P0 issues, verify these work:

1. **Login** → See account balance
2. **Place limit order** → Order appears in list
3. **Open position** → P&L updates in real-time
4. **Restart backend** → Frontend reconnects automatically

If all 4 work: **Critical issues are fixed! ✅**

---

**Last Updated:** 2026-02-12
**Status:** Ready to implement
**Next Action:** Open `services/api.ts` and fix line 353
