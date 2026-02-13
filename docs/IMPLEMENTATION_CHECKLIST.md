# Trading Engine - Implementation Checklist

**Generated:** 2026-02-12
**Quick Reference:** Step-by-step implementation guide

---

## Phase 1: Critical Path (P0) - Must Fix First ⚠️

### Task 1.1: Fix Broken API Endpoint Paths (2 hours)

**File:** `clients/desktop/src/services/api.ts`

**Line Changes:**
```typescript
// Line 353 - placeLimitOrder
- return fetchWithTimeout(`${API_ENDPOINTS.order}/limit`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/limit`, ...)

// Line 369 - placeStopOrder
- return fetchWithTimeout(`${API_ENDPOINTS.order}/stop`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/stop`, ...)

// Line 386 - placeStopLimitOrder
- return fetchWithTimeout(`${API_ENDPOINTS.order}/stop-limit`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/stop-limit`, ...)

// Line 425 - getTicks
- return fetchWithTimeout(`${API_ENDPOINTS.ticks}?...`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.ticks}?...`, ...) // Verify backend route first

// Line 438 - getOHLC
- return fetchWithTimeout(`${API_ENDPOINTS.ohlc}?...`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.candles}?...`, ...) // Use /candles endpoint

// Line 395 - getPendingOrders
- return fetchWithTimeout(`${API_ENDPOINTS.orders}/pending`, ...)
+ return fetchWithTimeout(`${API_ENDPOINTS.orders}/pending`, ...) // Already correct, verify backend
```

**Verification:**
```bash
# Test each endpoint
curl http://localhost:7999/api/orders/limit -X POST -d '{"symbol":"EURUSD",...}'
curl http://localhost:7999/api/orders/stop -X POST -d '{"symbol":"EURUSD",...}'
curl http://localhost:7999/ticks?symbol=EURUSD
curl http://localhost:7999/candles?symbol=EURUSD&timeframe=1H
```

**Checklist:**
- [ ] Update `placeLimitOrder()` URL
- [ ] Update `placeStopOrder()` URL
- [ ] Update `placeStopLimitOrder()` URL
- [ ] Verify `getTicks()` endpoint with backend
- [ ] Update `getOHLC()` to use `/candles`
- [ ] Verify `getPendingOrders()` endpoint
- [ ] Test all 6 endpoints with curl
- [ ] Run frontend and test order placement

---

### Task 1.2: Implement Risk Calculation Endpoints (6-8 hours)

**Backend File:** Create `backend/api/handlers/risk_calculator.go`

**Step 1: Create the file**
```bash
cd backend/api/handlers
touch risk_calculator.go
```

**Step 2: Implement handlers** (see full code in main fix plan)
- [ ] Implement `HandleCalculateLot()`
- [ ] Implement `HandleMarginPreview()`
- [ ] Add helper functions for pip value calculation
- [ ] Add contract size lookup per symbol

**Step 3: Register routes in `main.go`**
```go
// Add after other /risk/* routes (around line 1500)
http.HandleFunc("/risk/calculate-lot", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    apiHandler.HandleCalculateLot(w, r)
})

http.HandleFunc("/risk/margin-preview", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    if r.Method == "POST" {
        apiHandler.HandleMarginPreview(w, r)
        return
    }
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
})
```

**Step 4: Test endpoints**
```bash
# Test lot calculation
curl "http://localhost:7999/risk/calculate-lot?symbol=EURUSD&riskPercent=1&slPips=20&balance=10000"

# Test margin preview
curl -X POST http://localhost:7999/risk/margin-preview \
  -H "Content-Type: application/json" \
  -d '{"symbol":"EURUSD","volume":0.01,"leverage":100}'
```

**Checklist:**
- [ ] Create `risk_calculator.go`
- [ ] Implement `HandleCalculateLot()`
- [ ] Implement `HandleMarginPreview()`
- [ ] Register routes in `main.go`
- [ ] Compile backend: `go build ./cmd/server/main.go`
- [ ] Test lot calculation endpoint
- [ ] Test margin preview endpoint
- [ ] Verify frontend can call both endpoints

---

### Task 1.3: WebSocket Position/Order Updates (4-6 hours)

**Frontend File:** `clients/desktop/src/services/websocket-enhanced.ts`

**Step 1: Add message type interfaces** (line ~50)
```typescript
interface PositionUpdateMessage {
  type: 'position_update';
  data: {
    positionId: number;
    symbol: string;
    volume: number;
    openPrice: number;
    currentPrice: number;
    pnl: number;
    timestamp: string;
  };
}

interface OrderFillMessage {
  type: 'order_fill';
  data: {
    orderId: number;
    symbol: string;
    side: 'BUY' | 'SELL';
    volume: number;
    fillPrice: number;
    timestamp: string;
  };
}

interface MarginWarningMessage {
  type: 'margin_warning';
  data: {
    marginLevel: number;
    equity: number;
    margin: number;
    warningLevel: 'caution' | 'critical';
    timestamp: string;
  };
}
```

**Step 2: Update handleMessage()** (line ~200)
```typescript
private handleMessage(event: MessageEvent) {
  try {
    const message = JSON.parse(event.data);

    switch (message.type) {
      case 'tick':
        // ... existing tick handler
        break;
      case 'position_update':
        this.handlePositionUpdate(message.data);
        break;
      case 'order_fill':
        this.handleOrderFill(message.data);
        break;
      case 'margin_warning':
        this.handleMarginWarning(message.data);
        break;
      default:
        console.warn('[WS] Unknown message type:', message.type);
    }
  } catch (error) {
    console.error('[WS] Failed to parse message:', error);
  }
}
```

**Step 3: Implement handlers** (add after handleMessage)
```typescript
private handlePositionUpdate(data: PositionUpdateMessage['data']) {
  const store = useAppStore.getState();
  // Update position in store
  const positions = store.positions.map(p =>
    p.id === data.positionId
      ? { ...p, currentPrice: data.currentPrice, pnl: data.pnl }
      : p
  );
  store.setPositions(positions);
}

private handleOrderFill(data: OrderFillMessage['data']) {
  const notify = useNotificationStore.getState().addNotification;
  notify({
    type: 'success',
    title: 'Order Filled',
    message: `${data.side} ${data.volume} ${data.symbol} @ ${data.fillPrice}`,
    duration: 5000,
  });

  // Refresh orders and positions
  const store = useAppStore.getState();
  store.fetchOrders();
  store.fetchPositions();
}

private handleMarginWarning(data: MarginWarningMessage['data']) {
  const notify = useNotificationStore.getState().addNotification;
  notify({
    type: data.warningLevel === 'critical' ? 'error' : 'warning',
    title: 'Margin Warning',
    message: `Margin level: ${data.marginLevel.toFixed(2)}%`,
    duration: 10000,
  });
}
```

**Backend Changes:** `backend/internal/api/websocket/hub.go`
```go
// Add broadcast methods
func (h *Hub) BroadcastPositionUpdate(accountID int, position Position) {
    msg := Message{
        Type: "position_update",
        Data: position,
    }
    h.broadcastToAccount(accountID, msg)
}

func (h *Hub) BroadcastOrderFill(accountID int, order Order) {
    msg := Message{
        Type: "order_fill",
        Data: order,
    }
    h.broadcastToAccount(accountID, msg)
}

func (h *Hub) BroadcastMarginWarning(accountID int, warning MarginWarning) {
    msg := Message{
        Type: "margin_warning",
        Data: warning,
    }
    h.broadcastToAccount(accountID, msg)
}
```

**Checklist:**
- [ ] Add message type interfaces to frontend
- [ ] Update `handleMessage()` switch statement
- [ ] Implement `handlePositionUpdate()`
- [ ] Implement `handleOrderFill()`
- [ ] Implement `handleMarginWarning()`
- [ ] Add broadcast methods to backend hub
- [ ] Call broadcasts from engine on events
- [ ] Test position P&L updates in real-time
- [ ] Test order fill notifications
- [ ] Test margin warning alerts

---

### Task 1.4: WebSocket Re-subscription on Reconnect (2-3 hours)

**Frontend File:** `clients/desktop/src/services/websocket-enhanced.ts`

**Step 1: Add subscription tracking** (add to class properties)
```typescript
class EnhancedWebSocket {
  private activeSubscriptions: Set<string> = new Set();
  // ... existing properties
}
```

**Step 2: Update subscribe method** (line ~150)
```typescript
subscribe(channel: string, params?: any) {
  const subscriptionKey = `${channel}:${JSON.stringify(params || {})}`;
  this.activeSubscriptions.add(subscriptionKey);

  if (this.ws?.readyState === WebSocket.OPEN) {
    this.ws.send(JSON.stringify({
      action: 'subscribe',
      channel,
      params,
    }));
    console.log(`[WS] Subscribed to ${channel}`, params);
  } else {
    console.log(`[WS] Queued subscription to ${channel} (will subscribe on connect)`);
  }
}
```

**Step 3: Update onOpen to restore subscriptions** (line ~180)
```typescript
private onOpen = () => {
  console.log('[WS] Connected successfully');
  this.reconnectAttempts = 0;

  // Re-subscribe to all active channels
  if (this.activeSubscriptions.size > 0) {
    console.log(`[WS] Re-subscribing to ${this.activeSubscriptions.size} channels`);
    this.activeSubscriptions.forEach(sub => {
      const [channel, paramsStr] = sub.split(':');
      const params = paramsStr === '{}' ? undefined : JSON.parse(paramsStr);

      this.ws?.send(JSON.stringify({
        action: 'subscribe',
        channel,
        params,
      }));
    });
  }

  this.emit('open');
};
```

**Step 4: Add unsubscribe method**
```typescript
unsubscribe(channel: string, params?: any) {
  const subscriptionKey = `${channel}:${JSON.stringify(params || {})}`;
  this.activeSubscriptions.delete(subscriptionKey);

  if (this.ws?.readyState === WebSocket.OPEN) {
    this.ws.send(JSON.stringify({
      action: 'unsubscribe',
      channel,
      params,
    }));
    console.log(`[WS] Unsubscribed from ${channel}`);
  }
}
```

**Checklist:**
- [ ] Add `activeSubscriptions` Set property
- [ ] Update `subscribe()` to track subscriptions
- [ ] Update `onOpen()` to restore subscriptions
- [ ] Implement `unsubscribe()` method
- [ ] Test: Subscribe to market data → Restart backend → Verify auto re-subscribe
- [ ] Verify no duplicate subscriptions on reconnect
- [ ] Check browser console for re-subscription logs

---

## Phase 2: Core Features (P1) - High Impact 🔥

### Task 2.1: Connect Symbol Screener to Real Data (1 day)

**File:** `clients/desktop/src/store/useScreenerStore.ts`

**Remove:** `generateMockSymbols()` function (lines 85-152)

**Replace with:**
```typescript
export const useScreenerStore = create<ScreenerState>((set, get) => ({
  symbols: [],
  isLoading: false,

  fetchScreenerData: async () => {
    set({ isLoading: true });
    try {
      // Step 1: Fetch symbol list
      const symbolsResp = await api.getSymbols();
      if (!symbolsResp.success) throw new Error('Failed to fetch symbols');

      // Step 2: Fetch latest ticks for each symbol (parallel)
      const tickPromises = symbolsResp.data.map(sym =>
        api.getTicks(sym.symbol, 100).catch(() => ({ data: [] }))
      );
      const ticksResponses = await Promise.all(tickPromises);

      // Step 3: Calculate indicators from real tick data
      const enrichedSymbols = symbolsResp.data.map((sym, i) => {
        const ticks = ticksResponses[i].data || [];
        const prices = ticks.map(t => (t.bid + t.ask) / 2);

        return {
          symbol: sym.symbol,
          bid: ticks[ticks.length - 1]?.bid || 0,
          ask: ticks[ticks.length - 1]?.ask || 0,
          change: calculatePriceChange(prices),
          changePercent: calculateChangePercent(prices),
          rsi: calculateRSI(prices, 14),
          sma20: calculateSMA(prices, 20),
          sma50: calculateSMA(prices, 50),
          volume: ticks.length, // Use tick count as proxy
        };
      });

      set({ symbols: enrichedSymbols, isLoading: false });
    } catch (error) {
      console.error('[Screener] Failed to fetch data:', error);
      set({ isLoading: false });
    }
  },
}));

// Technical indicator helpers
function calculateRSI(prices: number[], period: number): number {
  if (prices.length < period + 1) return 50;

  let gains = 0, losses = 0;
  for (let i = prices.length - period; i < prices.length; i++) {
    const change = prices[i] - prices[i - 1];
    if (change > 0) gains += change;
    else losses += Math.abs(change);
  }

  const avgGain = gains / period;
  const avgLoss = losses / period;
  const rs = avgGain / avgLoss;
  return 100 - (100 / (1 + rs));
}

function calculateSMA(prices: number[], period: number): number {
  if (prices.length < period) return 0;
  const slice = prices.slice(-period);
  return slice.reduce((a, b) => a + b, 0) / period;
}

function calculatePriceChange(prices: number[]): number {
  if (prices.length < 2) return 0;
  return prices[prices.length - 1] - prices[0];
}

function calculateChangePercent(prices: number[]): number {
  if (prices.length < 2 || prices[0] === 0) return 0;
  const change = prices[prices.length - 1] - prices[0];
  return (change / prices[0]) * 100;
}
```

**Component Update:** `components/SymbolScreener.tsx`
```typescript
// Replace useEffect that calls generateMockSymbols()
useEffect(() => {
  fetchScreenerData(); // Fetch real data on mount
}, []);
```

**Checklist:**
- [ ] Remove `generateMockSymbols()` function
- [ ] Implement `fetchScreenerData()` method
- [ ] Implement `calculateRSI()` helper
- [ ] Implement `calculateSMA()` helper
- [ ] Update component to call `fetchScreenerData()`
- [ ] Test screener displays real symbol data
- [ ] Verify RSI/SMA calculations are reasonable
- [ ] Test with different timeframes

---

### Task 2.2: Connect News Feed to Backend (1 day)

**File:** `clients/desktop/src/store/useNewsStore.ts`

**Remove:** `generateMockNews()` function (lines 36-212)

**Replace with:**
```typescript
export const useNewsStore = create<NewsState>((set, get) => ({
  news: [],
  isLoading: false,
  selectedCategory: 'all',

  fetchNews: async () => {
    set({ isLoading: true });
    try {
      const response = await fetch(`${API_ENDPOINTS.marketNews.articles}`);
      if (!response.ok) throw new Error('Failed to fetch news');

      const data = await response.json();
      set({
        news: data.articles || data || [],
        isLoading: false
      });
    } catch (error) {
      console.error('[News] Failed to fetch:', error);
      set({ isLoading: false });
    }
  },

  refreshNews: async () => {
    await get().fetchNews();
  },
}));
```

**Backend Verification:**
```bash
# Check if backend returns news
curl http://localhost:7999/admin/market-news/articles

# If returns mock data, that's OK for now
# Real news integration can be Phase 3
```

**Checklist:**
- [ ] Remove `generateMockNews()` function
- [ ] Implement `fetchNews()` method
- [ ] Test backend endpoint returns data
- [ ] Update components to call `fetchNews()`
- [ ] Verify news panel displays backend data
- [ ] (Optional) Add news filtering by category
- [ ] (Optional) Add news search functionality

---

### Task 2.3: Implement Export Endpoints (1 day)

**Backend File:** Create `backend/api/handlers/export.go`

**Step 1: Install PDF library**
```bash
cd backend
go get github.com/jung-kurt/gofpdf
```

**Step 2: Create export handlers** (see full code in main plan)

**Step 3: Register routes in `main.go`**
```go
// Add after other /api/analytics/* routes
http.HandleFunc("/api/analytics/export/pdf", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    apiHandler.HandleExportPDF(w, r)
})

http.HandleFunc("/api/analytics/export/csv", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    apiHandler.HandleExportCSV(w, r)
})
```

**Step 4: Test endpoints**
```bash
# Test PDF export
curl "http://localhost:7999/api/analytics/export/pdf?type=trades" -o trades.pdf

# Test CSV export
curl "http://localhost:7999/api/analytics/export/csv?type=trades" -o trades.csv
```

**Checklist:**
- [ ] Install `gofpdf` library
- [ ] Create `export.go` file
- [ ] Implement `HandleExportPDF()`
- [ ] Implement `HandleExportCSV()`
- [ ] Register routes in `main.go`
- [ ] Test PDF generation
- [ ] Test CSV generation
- [ ] Verify frontend ExportDialog works
- [ ] Test downloading reports from UI

---

## Phase 3: Polish (P2/P3) - Quality of Life 🎨

### Task 3.1: Replace alert() with Toasts (1 day)

**Files to update:** 23 files, 79 instances

**Search pattern:**
```bash
cd clients/desktop/src
grep -r "alert(" --include="*.tsx" --include="*.ts"
```

**Replacement pattern:**
```typescript
// Before
alert('Order placed successfully!');

// After
import { useNotificationStore } from '../store/useNotificationStore';

const notify = useNotificationStore.getState().addNotification;
notify({
  type: 'success',
  message: 'Order placed successfully!',
  duration: 3000,
});
```

**Top files to update:**
1. `components/WorkspaceManager.tsx` - 13 alerts
2. `components/PositionsPanel.tsx` - 7 alerts
3. `components/PositionList.tsx` - 7 alerts
4. `components/OrderEntryPanel.tsx` - 8 alerts
5. `components/layout/MarketWatchPanel.tsx` - 6 alerts
6. ...17 more files

**Checklist:**
- [ ] Update WorkspaceManager alerts (13)
- [ ] Update PositionsPanel alerts (7)
- [ ] Update PositionList alerts (7)
- [ ] Update OrderEntryPanel alerts (8)
- [ ] Update MarketWatchPanel alerts (6)
- [ ] Update remaining 18 files (38 alerts)
- [ ] Test each replaced alert still triggers notification
- [ ] Remove any lingering alert() calls

---

### Task 3.2: Remove console.log Statements (1 day)

**Strategy:** Gate logs behind development mode flag

**Create logger utility:** `utils/logger.ts`
```typescript
export const logger = {
  log: (...args: any[]) => {
    if (import.meta.env.DEV) {
      console.log(...args);
    }
  },
  error: (...args: any[]) => {
    console.error(...args); // Always log errors
  },
  warn: (...args: any[]) => {
    if (import.meta.env.DEV) {
      console.warn(...args);
    }
  },
};
```

**Replace across files:**
```bash
# Find all console.log calls
grep -r "console\.log" clients/desktop/src --include="*.tsx" --include="*.ts" | wc -l
# Output: 150+

# Replace manually in high-volume files:
# - App.tsx (15 logs)
# - MarketWatchPanel.tsx (15 logs)
# - TradingChart.tsx (10 logs)
# - MenuBar.tsx (8 logs)
```

**Checklist:**
- [ ] Create `utils/logger.ts`
- [ ] Replace console.log in App.tsx
- [ ] Replace console.log in MarketWatchPanel.tsx
- [ ] Replace console.log in TradingChart.tsx
- [ ] Replace console.log in MenuBar.tsx
- [ ] Replace in remaining 100+ files
- [ ] Build production and verify no logs appear

---

### Task 3.3: Replace Math.random() IDs with UUID (1 hour)

**Install library:**
```bash
cd clients/desktop
npm install uuid
npm install -D @types/uuid
```

**Files to update:**
- `hooks/useKeyboardShortcut.ts` - 2 instances
- `store/useChartTemplateStore.ts` - 3 instances
- `store/useNotificationStore.ts` - 1 instance
- `store/useDrawingStore.ts` - 4 instances
- `store/useAlertStore.ts` - 1 instance
- `store/useTradeJournalStore.ts` - 1 instance

**Replacement:**
```typescript
// Before
const id = `template-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

// After
import { v4 as uuidv4 } from 'uuid';
const id = uuidv4();
```

**Checklist:**
- [ ] Install uuid library
- [ ] Replace IDs in useKeyboardShortcut.ts
- [ ] Replace IDs in useChartTemplateStore.ts
- [ ] Replace IDs in useNotificationStore.ts
- [ ] Replace IDs in useDrawingStore.ts
- [ ] Replace IDs in useAlertStore.ts
- [ ] Replace IDs in useTradeJournalStore.ts
- [ ] Test no collisions occur with new IDs

---

## Final Testing Checklist ✅

### Critical Path Tests
- [ ] Login with admin/Admin@123
- [ ] Place market order - verify success
- [ ] Place limit order - verify success
- [ ] Place stop order - verify success
- [ ] View open positions - verify real data
- [ ] Close position - verify execution
- [ ] Check position P&L updates in real-time
- [ ] Receive order fill notification
- [ ] Test margin preview calculation
- [ ] Test risk calculator (lot size)
- [ ] Restart backend - verify WebSocket reconnects
- [ ] Verify market data resumes after reconnect

### Core Features Tests
- [ ] Symbol Screener shows real RSI/MA values
- [ ] News panel displays articles from backend
- [ ] Export trades to PDF - verify download
- [ ] Export trades to CSV - verify download
- [ ] Order Book shows real market depth (if connected)

### UI/UX Tests
- [ ] No alert() dialogs appear (replaced with toasts)
- [ ] No console.log output in production build
- [ ] Toast notifications appear for all actions
- [ ] Error messages are user-friendly

---

## Quick Commands Reference

### Backend
```bash
# Compile
cd backend
go build -o server.exe ./cmd/server/main.go

# Run
./server.exe

# Test endpoint
curl http://localhost:7999/health
```

### Frontend
```bash
# Install deps
cd clients/desktop
npm install

# Dev mode
npm run dev

# Type check
npm run typecheck

# Build production
npm run build

# Test
npm run test
```

---

**Document Version:** 1.0
**Last Updated:** 2026-02-12
**Total Tasks:** 17 major tasks
**Estimated Time:** 9-15 days
