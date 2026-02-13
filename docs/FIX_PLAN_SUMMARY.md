# Trading Engine - Fix Plan Executive Summary

**Date:** 2026-02-12
**Status:** Analysis Complete - Ready to Implement
**Project:** RTX Trading Engine Desktop Client

---

## 📊 Current State Overview

### What's Working ✅
- **Authentication:** JWT login/logout, token management, auto-redirect on 401
- **Market Data:** Real-time WebSocket ticks, symbol list, bid/ask updates
- **Basic Trading:** Market orders, limit orders, stop orders, position management
- **Account Info:** Balance, equity, margin display via `/account/summary`
- **Backend:** Fully functional Go server with 428+ endpoints on port 7999

### What's Broken ❌
- **74 non-functional features** identified
- **29 components using mock data** (fake prices, indicators, signals)
- **9 missing backend endpoints** (risk calculator, exports, advanced orders)
- **6 broken API paths** (wrong URLs in frontend)
- **5 WebSocket gaps** (missing real-time updates for positions/orders)
- **220+ code quality issues** (79 alert() calls, 150+ console.logs)

---

## 🎯 The Fix - 4 Phases

### Phase 1: Critical Path (P0) - 1-2 Days ⚠️

**Must fix immediately - blocks core trading:**

1. **Fix 6 Broken API Paths** (2 hours)
   - File: `clients/desktop/src/services/api.ts`
   - Issue: Wrong URLs like `/order/limit` instead of `/api/orders/limit`
   - Impact: Limit/stop orders fail silently
   - Fix: Update URL strings in 6 functions

2. **Implement Risk Calculation Endpoints** (6-8 hours)
   - Files: Create `backend/api/handlers/risk_calculator.go`
   - Missing: `/risk/calculate-lot`, `/risk/margin-preview`
   - Impact: Position sizing and margin preview broken
   - Fix: Add 2 new backend handlers + register routes

3. **Add WebSocket Position/Order Updates** (4-6 hours)
   - File: `clients/desktop/src/services/websocket-enhanced.ts`
   - Missing: Real-time P&L, order fills, margin warnings
   - Impact: Users don't see changes until page refresh
   - Fix: Add 3 message handlers + backend broadcasts

4. **Implement WebSocket Re-subscription** (2-3 hours)
   - File: `clients/desktop/src/services/websocket-enhanced.ts`
   - Missing: Auto re-subscribe after reconnect
   - Impact: Users silently lose market data after disconnect
   - Fix: Track subscriptions, restore on reconnect

**Result:** Core trading functionality fully operational

---

### Phase 2: Core Features (P1) - 3-5 Days 🔥

**High-impact features users expect:**

1. **Connect Symbol Screener** (1 day)
   - Remove: `generateMockSymbols()` function
   - Add: Fetch real ticks, calculate real RSI/MA indicators
   - Impact: Screener shows accurate market data

2. **Connect Trading Signals** (1 day)
   - Remove: `generateMockSignals()` function
   - Add: Fetch from backend `/admin/signals/active`
   - Impact: Signal panel shows real opportunities

3. **Connect News Feed** (1 day)
   - Remove: 17 hardcoded fake articles
   - Add: Fetch from backend `/admin/market-news/articles`
   - Impact: Users see real financial news

4. **Implement Export Endpoints** (1 day)
   - Add: PDF/CSV export handlers in backend
   - Files: Create `backend/api/handlers/export.go`
   - Impact: Users can download trade reports

5. **Connect Order Book** (1 day)
   - Remove: `generateMockOrderBook()` function
   - Add: Subscribe to real order book WebSocket
   - Impact: Shows actual market depth

**Result:** All user-facing features use real data

---

### Phase 3: Advanced Features (P2) - 3-5 Days 📈

**Nice-to-have enhancements:**

1. **Advanced Order Types** (2 days)
   - Add: OCO, Trailing Stop, Bracket, Scaled orders
   - Files: Create `backend/api/handlers/advanced_orders.go`
   - Frontend already has UI - just needs backend

2. **Analytics Components** (2 days)
   - Replace mock data in 14 analytics components
   - Use real trade history from `/api/trades/history`
   - Calculate real performance metrics

3. **Server-Side Persistence** (2 days)
   - Move chart templates/drawings from localStorage to backend
   - Add database tables + API endpoints
   - Users keep data across devices

4. **Replace alert() Calls** (1 day)
   - Replace 79 browser alerts with toast notifications
   - Much better UX

**Result:** Professional-grade analytics and persistence

---

### Phase 4: Polish (P3) - 2-3 Days 🎨

**Quality improvements:**

1. Remove 150+ console.log statements (1 day)
2. Implement undo/redo handlers (1 day)
3. Replace Math.random() IDs with UUID (1 hour)
4. Implement placeholder features (1-2 days)

**Result:** Production-ready code quality

---

## 📁 Critical Files Reference

### Frontend (Must Modify)
```
clients/desktop/src/
├── services/api.ts                    ⚠️ Fix 6 broken endpoints (P0)
├── services/websocket-enhanced.ts     ⚠️ Add 4 missing handlers (P0)
├── store/useAdvancedOrdersStore.ts    ✅ Frontend ready, needs backend
├── store/useNewsStore.ts              ❌ Remove generateMockNews() (P1)
├── store/useSignalStore.ts            ❌ Remove generateMockSignals() (P1)
├── store/useScreenerStore.ts          ❌ Remove generateMockSymbols() (P1)
└── components/
    ├── OrderBook.tsx                  ❌ Remove generateMockOrderBook() (P1)
    ├── EconomicCalendar.tsx           ❌ Remove generateMockEvents() (P2)
    └── ... 17 more mock components
```

### Backend (Must Create)
```
backend/
├── api/handlers/
│   ├── risk_calculator.go            ❌ CREATE (P0 - 6-8 hours)
│   ├── advanced_orders.go            ❌ CREATE (P2 - 2 days)
│   └── export.go                     ❌ CREATE (P1 - 1 day)
└── cmd/server/main.go                ⚠️ Register 10+ new routes
```

---

## 🚀 Quick Start - What to Do Right Now

### Step 1: Fix Critical Bugs (Today)

Open `clients/desktop/src/services/api.ts` and fix these 6 lines:

```typescript
// Line 353
- return fetchWithTimeout(`${API_ENDPOINTS.order}/limit`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/limit`, ...)

// Line 369
- return fetchWithTimeout(`${API_ENDPOINTS.order}/stop`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/stop`, ...)

// Line 386
- return fetchWithTimeout(`${API_ENDPOINTS.order}/stop-limit`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/stop-limit`, ...)

// Line 438
- return fetchWithTimeout(`${API_ENDPOINTS.ohlc}?...`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.candles}?...`, ...)

// Lines 425, 395 - Verify these match backend routes
```

**Test:**
```bash
cd clients/desktop
npm run dev
# Try placing limit/stop orders - should work now
```

---

### Step 2: Add Risk Calculator (Tomorrow)

Create `backend/api/handlers/risk_calculator.go`:

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
)

func (h *APIHandler) HandleCalculateLot(w http.ResponseWriter, r *http.Request) {
    // ... see full implementation in IMPLEMENTATION_CHECKLIST.md
}

func (h *APIHandler) HandleMarginPreview(w http.ResponseWriter, r *http.Request) {
    // ... see full implementation in IMPLEMENTATION_CHECKLIST.md
}
```

Register in `backend/cmd/server/main.go`:
```go
http.HandleFunc("/risk/calculate-lot", apiHandler.HandleCalculateLot)
http.HandleFunc("/risk/margin-preview", apiHandler.HandleMarginPreview)
```

**Test:**
```bash
cd backend
go build -o server.exe ./cmd/server/main.go
./server.exe

# In another terminal
curl "http://localhost:7999/risk/calculate-lot?symbol=EURUSD&riskPercent=1&slPips=20&balance=10000"
```

---

### Step 3: Add WebSocket Updates (Day 3)

Update `clients/desktop/src/services/websocket-enhanced.ts`:

1. Add message type interfaces (3 types)
2. Update `handleMessage()` switch statement
3. Add 3 handler methods
4. Update backend to broadcast position/order events

**Test:**
```bash
# Place an order and watch for real-time fill notification
# Open position and watch P&L update without refresh
```

---

## 📋 Testing Checklist

### After Phase 1 (Critical)
- [ ] Login works
- [ ] Market/limit/stop orders place successfully
- [ ] Positions show real-time P&L
- [ ] Order fills show toast notification
- [ ] Margin preview calculates correctly
- [ ] WebSocket reconnects and restores subscriptions

### After Phase 2 (Core)
- [ ] Symbol Screener shows real RSI/MA
- [ ] News panel displays backend articles
- [ ] Trading Signals load from API
- [ ] Can export trades to PDF/CSV
- [ ] Order Book shows real depth

### After Phase 3 (Advanced)
- [ ] OCO orders work
- [ ] Performance Analytics uses real data
- [ ] Chart templates save to server
- [ ] No alert() dialogs (replaced with toasts)

### After Phase 4 (Polish)
- [ ] No console.log in production build
- [ ] Undo/Redo works for chart actions
- [ ] All IDs use UUID (no collisions)

---

## 📊 Effort Breakdown

| Phase | Days | Priority | Status |
|-------|------|----------|--------|
| Phase 1: Critical Path | 1-2 | **P0** | ⚠️ Start immediately |
| Phase 2: Core Features | 3-5 | **P1** | 🔥 High impact |
| Phase 3: Advanced | 3-5 | **P2** | 📈 Nice to have |
| Phase 4: Polish | 2-3 | **P3** | 🎨 Quality of life |
| **Total** | **9-15 days** | | |

---

## 🎯 Success Criteria

### Minimum Viable Fix (Phase 1 Only)
✅ All core trading functions work without errors
✅ Real-time updates for positions and orders
✅ Risk calculations available to traders
✅ WebSocket remains connected during sessions

**Timeline:** 1-2 days
**Effort:** ~20 hours
**Priority:** Must complete first

### Full Feature Parity (Phases 1-2)
✅ All user-facing features use real data
✅ No mock data in any component
✅ Exports and analytics work
✅ News, signals, screener functional

**Timeline:** 4-7 days
**Effort:** ~40 hours
**Priority:** Recommended for production

### Production Ready (Phases 1-4)
✅ Advanced order types available
✅ Professional code quality (no alerts, logs)
✅ Server-side persistence
✅ Full test coverage

**Timeline:** 9-15 days
**Effort:** ~80 hours
**Priority:** Ideal for commercial deployment

---

## 📞 Quick Reference

### Documentation
- **Full Fix Plan:** `docs/MISSING_FUNCTIONALITY_FIX_PLAN.md` (comprehensive)
- **Implementation Guide:** `docs/IMPLEMENTATION_CHECKLIST.md` (step-by-step)
- **This Summary:** `docs/FIX_PLAN_SUMMARY.md` (overview)

### Previous Reports
- `DESKTOP_CLIENT_FIX_VERIFICATION_REPORT.md` - What was already fixed
- `docs/DESKTOP_CLIENT_FIXES.md` - Original fixes documentation
- `docs/NON_FUNCTIONAL_FEATURES.md` - Detailed analysis of all 74 issues

### Server Status
- Backend: http://localhost:7999
- Desktop: http://localhost:5173
- Credentials: admin / Admin@123

---

## 🚦 Decision Matrix

### Should I fix everything?
**No.** Start with Phase 1 (P0) to unblock trading.

### Which phase is most important?
**Phase 1 (Critical Path)** - Without this, core trading is broken.

### Can I skip Phase 3-4?
**Yes**, for MVP. Phases 3-4 are polish and advanced features.

### What if I only have 1 week?
**Do Phase 1 + Phase 2 only.** This gives you a fully functional platform with real data.

### What's the bare minimum?
**Phase 1 Task 1.1-1.2** (fix API paths + risk calculator) - Takes ~10 hours, unblocks 80% of issues.

---

## 🎬 Next Steps

1. **Read:** `IMPLEMENTATION_CHECKLIST.md` for detailed code examples
2. **Fix:** 6 broken API paths in `services/api.ts` (2 hours)
3. **Create:** `risk_calculator.go` backend handler (6 hours)
4. **Test:** Place orders and verify they execute
5. **Continue:** Follow checklist for remaining Phase 1 tasks

---

**Document Version:** 1.0
**Generated:** 2026-02-12
**Total Issues Identified:** 74 non-functional features
**Total Code Quality Issues:** 220+
**Estimated Fix Time:** 9-15 days for complete implementation
**Minimum Fix Time:** 1-2 days for critical path only

**Status:** 🟢 Ready to implement - All issues documented with solutions
