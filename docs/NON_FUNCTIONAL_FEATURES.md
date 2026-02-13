# Trading Engine Desktop Client - Non-Functional Features Report

> **Generated:** 2026-02-12 | **Total Issues:** 74 Non-Functional Features + 220+ Code Quality Issues
> **Scanned:** 256 TypeScript/TSX files in `clients/desktop/src/`
> **Backend Endpoints:** 428 total | 100+ unused by frontend

---

## Table of Contents

- [Category 1: Mock Data Components (29)](#category-1-mock-data-components-29-features)
- [Category 2: Missing Backend Endpoints (9)](#category-2-missing-backend-endpoints-9-features)
- [Category 3: Broken API Endpoint Paths (6)](#category-3-broken-api-endpoint-paths-6-features)
- [Category 4: Broken String Literal Bug (2)](#category-4-broken-string-literal-bug-2-features)
- [Category 5: Placeholder/Stub UI Features (12)](#category-5-placeholderstub-ui-features-12-features)
- [Category 6: Empty Settings Tabs (5)](#category-6-empty-settings-tabs-5-features)
- [Category 7: Bottom Dock Placeholder Tabs (4)](#category-7-bottom-dock-placeholder-tabs-4-features)
- [Category 8: Disabled Menu Items (1)](#category-8-disabled-menu-items-1-feature)
- [Category 9: WebSocket Integration Gaps (5)](#category-9-websocket-integration-gaps-5-features)
- [Category 10: LocalStorage-Only Persistence (6)](#category-10-localstorage-only-persistence-6-features)
- [Code Quality Issues (220+)](#code-quality-issues-220)
- [Unused Backend Endpoints (100+)](#unused-backend-endpoints-100)
- [Fix Priority Roadmap](#fix-priority-roadmap)

---

## Category 1: Mock Data Components (29 Features)

These components exist in the UI but display **100% fake/randomly generated data** with zero API integration.

### 1.1 Advanced Order Submissions (5 stubs)

| # | Feature | File | Lines | Issue |
|---|---------|------|-------|-------|
| 1 | OCO Order Submit | `store/useAdvancedOrdersStore.ts` | 200-206 | `console.log()` + fake `setTimeout(500)` + `alert()`. Comment: `// TODO: Replace with actual API call` |
| 2 | Trailing Stop Submit | `store/useAdvancedOrdersStore.ts` | 208-213 | Same stub pattern - no API call |
| 3 | Time-Based Order Submit | `store/useAdvancedOrdersStore.ts` | 215-220 | Same stub pattern - no API call |
| 4 | Bracket Order Submit | `store/useAdvancedOrdersStore.ts` | 222-227 | Same stub pattern - no API call |
| 5 | Scaled Order Submit | `store/useAdvancedOrdersStore.ts` | 229-234 | Same stub pattern - no API call |

**Current behavior:**
```typescript
submitOCOOrder: async () => {
    console.log('[AdvancedOrders] Submitting OCO Order:', ocoOrder);
    // TODO: Replace with actual API call
    await new Promise((resolve) => setTimeout(resolve, 500));
    alert('OCO Order submitted successfully!');
},
```

**Expected behavior:** Should POST to `/api/orders/oco` (or equivalent) with proper error handling and response parsing.

---

### 1.2 Store-Level Mock Data (5 stores)

| # | Feature | File | Lines | Mock Function | Data Generated |
|---|---------|------|-------|---------------|----------------|
| 6 | News Feed | `store/useNewsStore.ts` | 36-212 | `generateMockNews()` | 17 hardcoded news items, `refreshNews()` just shuffles timestamps |
| 7 | Sentiment Analysis | `store/useSentimentStore.ts` | 54-186 | `generateSymbolSentiments()`, `generateNewsItems()`, `generateSentimentIndicators()` | Random bullish %, fake fear/greed index, fake VIX |
| 8 | Trading Signals | `store/useSignalStore.ts` | 69-237 | `generateMockSignals()`, `generateHistoricalSignals()` | 15 fake active + 20 fake historical signals |
| 9 | Symbol Screener | `store/useScreenerStore.ts` | 85-152 | `generateMockSymbols()` | 40+ symbols with fabricated prices, RSI, MAs, volumes |
| 10 | Backtest Engine | `store/useBacktestStore.ts` | 52-75 | N/A | `startBacktest()` only sets `isRunning: true`, no API call |

---

### 1.3 Component-Level Mock Data (19 components)

| # | Feature | File | Lines | Mock Function | Impact |
|---|---------|------|-------|---------------|--------|
| 11 | Copy Trading Leaderboard | `components/CopyTradingLeaderboard.tsx` | 51-112 | `generateMockTraders()` | 25 fake traders with random performance stats |
| 12 | Economic Calendar | `components/EconomicCalendar.tsx` | 63-147 | `generateMockEvents()` | Hardcoded economic events, not from real API |
| 13 | News Impact Analyzer | `components/NewsImpactAnalyzer.tsx` | 57-130 | `generateUpcomingEvents()`, `generateEventTypeData()` | All impact data is fabricated |
| 14 | Market Replay | `components/MarketReplay.tsx` | 48-88 | `generateMockTicks()` | Random walk algorithm, no real historical data |
| 15 | Order Book | `components/OrderBook.tsx` | 52-57, 268 | `generateMockOrderBook()` | Simulated bid/ask depth data |
| 16 | Order Flow Visualizer | `components/OrderFlowVisualizer.tsx` | 34-56, 368 | `generateMockTrades()` | 200 fake trades for visualization |
| 17 | Correlation Matrix | `components/CorrelationMatrix.tsx` | 28-59, 135 | `generateMockReturns()` | Artificial correlations based on symbol names, not market data |
| 18 | Multi-Account Manager | `components/MultiAccountManager.tsx` | 30, 164+, 434 | `generateMockAccounts()` | 5 simulated accounts with random equity updates |
| 19 | Strategy Tester | `components/StrategyTester.tsx` | 22-161 | `runBacktest()` | Fake trades with random win probability, Sharpe = `Math.random()` |
| 20 | Trading Simulator | `components/TradingSimulator.tsx` | 7-49 | N/A | Hardcoded base prices, movement = `(Math.random() - 0.5) * 0.001` |
| 21 | Performance Analytics | `components/PerformanceAnalytics.tsx` | 37, 347 | `generateMockTrades()` | Fake trade history for metrics |
| 22 | Equity Curve Tracker | `components/EquityCurveTracker.tsx` | 35, 177 | `generateMockData()` | Simulated daily equity snapshots |
| 23 | Trade Analytics | `components/TradeAnalytics.tsx` | 36, 98 | `generateMockTrades()` | Fake trade statistics |
| 24 | Risk Heatmap | `components/RiskHeatmap.tsx` | 21, 79 | `generateMockPositions()` | Fake position data for heatmap |
| 25 | Risk Exposure Panel | `components/RiskExposurePanel.tsx` | 30, 585 | `generateMockPositions()` | Fabricated risk exposure data |
| 26 | Depth of Market | `components/professional/DepthOfMarket.tsx` | 37, 329 | `generateMockDepth()` | Simulated market depth |
| 27 | Exposure Heatmap | `components/ExposureHeatmap.tsx` | 252-277 | N/A | `// TODO: Replace with actual API endpoint`, falls back to mock |
| 28 | Trade Journal | `components/TradeJournal.tsx` | 61, 481 | `generateMockTrades()` | Fake journal entries |
| 29 | Trade Journal V2 | `components/TradeJournalV2.tsx` | 59, 1129 | `generateMockTrades()` | Fake journal entries (v2) |

---

## Category 2: Missing Backend Endpoints (9 Features)

The frontend calls these endpoints but they **do not exist** in the backend - all will return 404.

| # | Endpoint | Method | Frontend Caller | File | Impact |
|---|----------|--------|----------------|------|--------|
| 30 | `/risk/calculate-lot` | GET | PositionSizeCalculator, OrderEntry | `services/api.ts:457` | Cannot calculate position sizes |
| 31 | `/risk/margin-preview` | POST | OrderEntryPanel, TradePanel | `services/api.ts:469` | Cannot preview margin before trading |
| 32 | `/admin/execution-mode` | GET | AdminPanel | `services/api.ts:616` | Cannot view ABOOK/BBOOK mode |
| 33 | `/admin/execution-mode` | POST | AdminPanel | `services/api.ts:622` | Cannot switch execution modes |
| 34 | `/api/analytics/export/pdf` | GET | ExportDialog | `services/api.ts:707` | PDF export broken |
| 35 | `/api/analytics/export/csv` | GET | ExportDialog | `services/api.ts:725` | CSV export broken |
| 36 | `/api/performance` | GET | PerformanceAnalytics | `services/api.ts:742` | Performance report broken |
| 37 | `/api/alerts/rules/{id}/test` | POST | AlertRulesManager | `services/api.ts:694` | Cannot test alert rules |
| 38 | `/api/history/info` | GET | historyClient | `api/historyClient.ts:47` | Cannot validate symbol data ranges |

---

## Category 3: Broken API Endpoint Paths (6 Features)

These endpoints use **wrong URL patterns** (missing `/api/` prefix) and will hit wrong routes or 404.

| # | Function | Current Path (Wrong) | Expected Path | File | Line |
|---|----------|---------------------|---------------|------|------|
| 39 | `placeLimitOrder()` | `/order/limit` | `/api/orders/limit` | `services/api.ts` | 353 |
| 40 | `placeStopOrder()` | `/order/stop` | `/api/orders/stop` | `services/api.ts` | 369 |
| 41 | `placeStopLimitOrder()` | `/order/stop-limit` | `/api/orders/stop-limit` | `services/api.ts` | 386 |
| 42 | `getTicks()` | `/ticks` | `/api/market/ticks` | `services/api.ts` | 425 |
| 43 | `getOHLC()` | `/ohlc` | `/api/market/ohlc` | `services/api.ts` | 438 |
| 44 | `getPendingOrders()` | `/orders/pending` | `/api/orders/pending` | `services/api.ts` | 395 |

---

## Category 4: Broken String Literal Bug (2 Features)

`AlertRulesManager.tsx` uses **single quotes** instead of **backticks** for template strings. Variables will NOT interpolate - API calls fetch a literal string `"${API_ENDPOINTS.alertRules}"` instead of the actual URL.

| # | Feature | File | Line | Broken Code |
|---|---------|------|------|-------------|
| 45 | Alert Rules - Fetch All | `components/AlertRulesManager.tsx` | 37 | `fetch('${API_ENDPOINTS.alertRules}')` should be `` fetch(`${API_ENDPOINTS.alertRules}`) `` |
| 46 | Alert Rules - Create/Update | `components/AlertRulesManager.tsx` | 51 | Same bug - all CRUD operations silently fail |

---

## Category 5: Placeholder/Stub UI Features (12 Features)

These features show UI elements but trigger `alert()` dialogs or `console.log()` instead of real functionality.

| # | Feature | File | Line | Current Behavior |
|---|---------|------|------|-----------------|
| 47 | Order Entry from MarketWatch | `components/MarketWatch.tsx` | 82 | `alert('Order entry for ${symbol} - not yet implemented')` |
| 48 | Tick Chart from MarketWatch | `components/MarketWatch.tsx` | 91 | `alert('Tick chart for ${symbol} - not yet implemented')` |
| 49 | Chart Properties Dialog | `components/TradingChart.tsx` | 1178 | `alert('Chart properties dialog not yet implemented')` |
| 50 | Strategy Optimization | `components/StrategyTester.tsx` | 345 | Checkbox disabled with `opacity-50`, label: "Optimization (Coming Soon)" |
| 51 | Undo (Ctrl+Z) | `services/keyboardShortcuts.ts` | 58 | Shortcut registered, dispatches UNDO action, but no handler exists |
| 52 | Redo (Ctrl+Y) | `services/keyboardShortcuts.ts` | 68 | Shortcut registered, dispatches REDO action, but no handler exists |
| 53 | Drawing Line Style Submenu | `components/DrawingContextMenu.tsx` | 115 | `onClick={() => { }}` - empty handler, does nothing |
| 54 | Screener Open Chart Click | `components/SymbolScreener.tsx` | 460 | `onClick` only does `console.log('Open chart for', symbol)`, doesn't open chart |
| 55 | Account Panel Mock Info | `components/AccountPanel.tsx` | 34 | Hardcoded: `accountNumber: 'RTX-000123'`, comment: `// Mock account info` |
| 56 | Account Panel Transactions | `components/AccountPanel.tsx` | 53, 582 | Falls back to `generateMockTransactions()` on API failure |
| 57 | OrderEntry Hardcoded Values | `components/OrderEntry.tsx` | 138 | `leverage: 100`, `currency: 'USD'`, `contractSize: 100000` - comment: "Mock values" |
| 58 | Calendar Tab Mock Events | `components/layout/CalendarTab.tsx` | 34, 65 | `MOCK_EVENTS` constant array |

---

## Category 6: Empty Settings Tabs (5 Features)

These settings tabs render a `PlaceholderTab` component showing only an icon, title, and description text - no actual functionality.

**File:** `components/settings/OptionsDialog.tsx`

| # | Tab | Line | Description Shown | What Should Exist |
|---|-----|------|-------------------|-------------------|
| 59 | Events | 117 | "Configure sound notifications and system events" | Sound/notification event configuration form |
| 60 | Email | 119 | "SMTP server configuration for alerts" | SMTP settings form with test button |
| 61 | FTP | 120 | "Automated report publishing settings" | FTP server config for report upload |
| 62 | Community | 121 | "Login to the proprietary RTX5 ecosystem" | RTX5 Community login/registration |
| 63 | Signals | 122 | "Subscribe to institutional signal providers" | Signal provider subscription management |

---

## Category 7: Bottom Dock Placeholder Tabs (4 Features)

These bottom dock tabs render a generic `PlaceholderTab` with "No active items to display".

**File:** `components/BottomDock.tsx` (Lines 45-49, 157-159)

| # | Tab | Line | What Should Exist |
|---|-----|------|-------------------|
| 64 | News | 45 | Real-time financial news feed |
| 65 | Mailbox | 46 | Internal messaging / broker communications |
| 66 | Company | 48 | Company information and announcements |
| 67 | Alerts (Bottom Dock) | 49 | Alert notifications panel (AlertRulesManager exists but isn't used here) |

**PlaceholderTab code (lines 524-531):**
```tsx
function PlaceholderTab({ title }: { title: string }) {
    return (
        <div className="flex-1 flex flex-col items-center justify-center text-zinc-600">
            <Briefcase size={32} className="mb-2 opacity-20" />
            <div className="text-xs font-medium">RTX5 {title}</div>
            <div className="text-[10px]">No active items to display</div>
        </div>
    );
}
```

---

## Category 8: Disabled Menu Items (1 Feature)

| # | Feature | File | Line | Status |
|---|---------|------|------|--------|
| 68 | "Open Deleted" menu item | `components/layout/MenuBar.tsx` | 208 | `disabled: true` - grayed out, no handler |

---

## Category 9: WebSocket Integration Gaps (5 Features)

| # | Feature | File | Issue |
|---|---------|------|-------|
| 69 | Position Update Stream | `hooks/useWebSocket.ts` | No handler for real-time position P&L updates via WebSocket |
| 70 | Order Fill Notifications | `hooks/useWebSocket.ts` | No handler for order execution/fill events |
| 71 | Margin Call Alerts | `hooks/useWebSocket.ts` | No handler for margin level warnings |
| 72 | Heartbeat/Ping | `services/websocket-enhanced.ts` | No keep-alive ping mechanism implemented |
| 73 | Reconnect Re-subscribe | `services/websocket-enhanced.ts` | After reconnect, does NOT re-subscribe to previous channels - users silently lose data |

---

## Category 10: LocalStorage-Only Persistence (6 Features)

These stores use Zustand `persist()` middleware but only save to `localStorage`. Users lose all data on logout, cache clear, or device change.

| # | Store | File | Data Lost |
|---|-------|------|-----------|
| 74 | Trade Journal | `store/useTradeJournalStore.ts` | All journal entries |
| 75 | Currency Preferences | `store/useCurrencyStore.ts` | Exchange rate preferences |
| 76 | Chart Templates | `store/useTemplateStore.ts` | All saved chart templates |
| 77 | Chart Drawings | `store/useDrawingStore.ts` | All chart drawings and annotations |
| 78 | Session Map | `store/useSessionMapStore.ts` | Session state and layout |
| 79 | Position Size History | `store/usePositionSizeStore.ts` | Calculation history |

---

## Code Quality Issues (220+)

These don't break features but degrade UX and maintainability.

### Alert() Calls Instead of Toast Notifications (66+ instances)

| Area | Files | Count |
|------|-------|-------|
| Order Placement | `App.tsx`, `AppOrderEntryPanel.tsx` | 12 |
| Position Management | `PositionList.tsx`, `PositionsPanel.tsx` | 14 |
| Trade Panel | `TradePanel.tsx` | 4 |
| Market Watch | `MarketWatchPanel.tsx` | 6 |
| Workspace Manager | `WorkspaceManager.tsx` | 13 |
| Chart Templates | `ChartTemplateDialog.tsx`, `ChartTemplateManager.tsx` | 5 |
| Other Components | Various | 12+ |

### Console.log Debug Statements (150+ instances)

| File | Count | Pattern |
|------|-------|---------|
| `App.tsx` | 15 | `console.log('[App]...', value)` |
| `MarketWatchPanel.tsx` | 15 | Trading operation logging |
| `MenuBar.tsx` | 8 | Action logging |
| `TradingChart.tsx` | 10+ | Chart operation logging |
| Other files | 100+ | Various debug logs |

### Empty Catch Blocks (6 instances)

| File | Line | Code |
|------|------|------|
| `App.tsx` | 448 | `.catch(() => {})` |
| `AdvancedOrderPanel.tsx` | 68 | `.catch(() => { })` |
| `TradingChart.tsx` | 341, 346, 402, 964 | `try { ... } catch (e) { }` |

### Weak ID Generation (15+ instances)

Using `Math.random()` instead of UUID library for generating IDs:

| File | Lines |
|------|-------|
| `hooks/useKeyboardShortcut.ts` | 37, 71 |
| `store/useChartTemplateStore.ts` | 54, 98, 128 |
| `store/useNotificationStore.ts` | 49 |
| `store/useDrawingStore.ts` | 117, 155, 208, 228 |
| `store/useAlertStore.ts` | 57 |
| `store/useTradeJournalStore.ts` | 86 |

---

## Unused Backend Endpoints (100+)

The backend (`backend/cmd/server/main.go`) has **428 registered endpoints**. Over 100 have **zero frontend implementation**:

| Feature Group | Endpoint Count | Example Endpoints | Frontend Status |
|---------------|---------------|-------------------|-----------------|
| Broker P&L Analytics | 7 | `/admin/broker-pnl/summary`, `/admin/broker-pnl/top-symbols` | Not Implemented |
| Margin Management | 10 | `/admin/margin/calls`, `/admin/margin/liquidations` | Not Implemented |
| LP Performance | 8 | `/admin/lp-performance/slippage`, `/admin/lp-performance/execution-quality` | Not Implemented |
| Compliance/Surveillance | 15 | `/admin/compliance/breaches`, `/admin/surveillance/*` | Not Implemented |
| Spread Monitoring | 4 | `/admin/spreads/live`, `/admin/spreads/alerts` | Not Implemented |
| MAM/PAMM | 3 | `/admin/mam`, `/admin/mam/stats` | Not Implemented |
| System Logs | 5 | `/admin/logs`, `/admin/logs/error-rate` | Not Implemented |
| Transaction Management | 4 | `/admin/transactions`, `/admin/transactions/stats` | Not Implemented |
| Commission Tiers | 7 | `/admin/commissions/tiers`, `/admin/commissions/clients` | Not Implemented |
| Automated Trading | 3 | `/admin/automated-trading/eas` | Not Implemented |
| Social Trading | 5 | `/admin/social-trading/copies`, `/admin/social-trading/providers` | Not Implemented |
| Audit Trail | 4 | `/admin/audit-trail`, `/admin/audit-trail/export` | Not Implemented |
| Document Management | 8+ | `/admin/documents/*`, `/admin/client-notes/*` | Not Implemented |
| Data Management | 8 | `/admin/history/compress`, `/admin/backups/*` | Not Implemented |
| Market Data/News | 10 | `/admin/market-news/*`, `/admin/market-data/*` | Not Implemented |

> **Note:** These unused endpoints are primarily admin-panel features. They should be implemented in `admin/broker-admin` and `admin/super-admin` rather than the desktop client.

---

## Fix Priority Roadmap

### Phase 1: Critical (Trading Functionality Broken)

| Priority | Task | Features Fixed | Effort |
|----------|------|---------------|--------|
| P0 | Fix AlertRulesManager string literal bug (single quotes → backticks) | #45, #46 | 5 min |
| P0 | Fix 6 broken API endpoint paths in `services/api.ts` | #39-#44 | 15 min |
| P0 | Implement 5 advanced order submit functions with real API calls | #1-#5 | 2-3 hrs |
| P0 | Implement `/risk/calculate-lot` backend endpoint | #30 | 2-3 hrs |
| P0 | Implement `/risk/margin-preview` backend endpoint | #31 | 2-3 hrs |

### Phase 2: High (Core Features Using Mock Data)

| Priority | Task | Features Fixed | Effort |
|----------|------|---------------|--------|
| P1 | Connect OrderBook to real WebSocket data | #15 | 2-3 hrs |
| P1 | Connect Symbol Screener to real market data API | #9 | 3-4 hrs |
| P1 | Connect Trading Signals to real signal provider API | #8 | 3-4 hrs |
| P1 | Implement WebSocket reconnect re-subscribe | #73 | 1-2 hrs |
| P1 | Add WebSocket handlers for position/order updates | #69, #70, #71 | 3-4 hrs |
| P1 | Implement export endpoints (PDF, CSV) | #34, #35 | 4-6 hrs |

### Phase 3: Medium (Enhanced Features)

| Priority | Task | Features Fixed | Effort |
|----------|------|---------------|--------|
| P2 | Connect News Feed to real news API | #6, #12 | 4-6 hrs |
| P2 | Connect Sentiment Analysis to real data source | #7 | 4-6 hrs |
| P2 | Connect Economic Calendar to real API | #12 | 3-4 hrs |
| P2 | Replace mock data in analytics components | #21-#29 | 8-12 hrs |
| P2 | Implement Settings tabs (Events, Email, FTP) | #59-#61 | 6-8 hrs |
| P2 | Implement Bottom Dock tabs | #64-#67 | 4-6 hrs |
| P2 | Add server-side persistence for stores | #74-#79 | 4-6 hrs |

### Phase 4: Low (Polish & UX)

| Priority | Task | Features Fixed | Effort |
|----------|------|---------------|--------|
| P3 | Replace all `alert()` calls with toast notifications | 66+ instances | 4-6 hrs |
| P3 | Remove `console.log` debug statements | 150+ instances | 2-3 hrs |
| P3 | Fix empty catch blocks with proper error handling | 6 instances | 1-2 hrs |
| P3 | Replace `Math.random()` IDs with UUID | 15+ instances | 1-2 hrs |
| P3 | Implement Undo/Redo handlers | #51, #52 | 4-6 hrs |
| P3 | Implement placeholder features (Order Entry, Tick Chart, Chart Props) | #47-#49 | 6-8 hrs |
| P3 | Implement Community and Signals Hub tabs | #62, #63 | 8-12 hrs |

---

## Summary Statistics

| Metric | Value |
|--------|-------|
| **Total Non-Functional Features** | 74 |
| **Mock Data Components** | 29 |
| **Missing/Broken API Endpoints** | 17 |
| **Placeholder UI Elements** | 22 |
| **WebSocket Gaps** | 5 |
| **Persistence Issues** | 6 |
| **Code Quality Issues** | 220+ |
| **Unused Backend Endpoints** | 100+ |
| **Estimated Total Fix Effort** | 80-120 hours |
| **Files Needing Changes** | 51+ |

---

*This report was generated by scanning all 256 TypeScript/TSX source files in the desktop client using 5 parallel analysis agents.*
