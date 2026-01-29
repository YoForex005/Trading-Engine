# Component Architecture Diagram

## Client Trading Interface Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         App.tsx (Main Shell)                             │
│                                                                          │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │                     WebSocket Service                             │  │
│  │  • Auto-reconnect with exponential backoff                        │  │
│  │  • Tick buffering (100ms throttle)                                │  │
│  │  • Subscription management                                        │  │
│  │  • Connection state tracking                                      │  │
│  └─────────────────────────┬─────────────────────────────────────────┘  │
│                            │ publishes to                               │
│  ┌─────────────────────────▼─────────────────────────────────────────┐  │
│  │                  Zustand Global Store                             │  │
│  │  • ticks: Record<symbol, Tick>                                    │  │
│  │  • positions: Position[]                                          │  │
│  │  • orders: Order[]                                                │  │
│  │  • account: Account                                               │  │
│  │  • selectedSymbol: string                                         │  │
│  │  • UI state (chart type, timeframe, etc.)                         │  │
│  └───────────────────┬───────────────────────┬───────────────────────┘  │
│                      │ consumed by           │ consumed by             │
│  ┌───────────────────▼──────────┐  ┌─────────▼────────────────────┐   │
│  │  Trading Chart Component     │  │  Order Entry Panel           │   │
│  │  ┌────────────────────────┐  │  │  ┌────────────────────────┐ │   │
│  │  │ Lightweight Charts     │  │  │  │ Order Type Selector    │ │   │
│  │  │ • Candlestick          │  │  │  │ • Market               │ │   │
│  │  │ • Line                 │  │  │  │ • Limit                │ │   │
│  │  │ • Area                 │  │  │  │ • Stop                 │ │   │
│  │  │ • Real-time updates    │  │  │  │ • Stop-Limit           │ │   │
│  │  └────────────────────────┘  │  │  └────────────────────────┘ │   │
│  │  ┌────────────────────────┐  │  │  ┌────────────────────────┐ │   │
│  │  │ Chart Controls         │  │  │  │ Volume Input           │ │   │
│  │  │ • Timeframe selector   │  │  │  │ • Quick select         │ │   │
│  │  │ • Chart type toggle    │  │  │  │ • Price inputs         │ │   │
│  │  │ • Maximize button      │  │  │  │ • SL/TP inputs         │ │   │
│  │  └────────────────────────┘  │  │  └────────────────────────┘ │   │
│  └──────────────────────────────┘  │  ┌────────────────────────┐ │   │
│                                     │  │ BUY/SELL Buttons       │ │   │
│  ┌──────────────────────────────┐  │  │ • Live pricing         │ │   │
│  │  Position List Component     │  │  │ • Loading states       │ │   │
│  │  ┌────────────────────────┐  │  │  └────────────────────────┘ │   │
│  │  │ Position Table         │  │  └──────────────────────────────┘   │
│  │  │ • Symbol               │  │                                      │
│  │  │ • Type (BUY/SELL)      │  │  ┌──────────────────────────────┐   │
│  │  │ • Volume               │  │  │  Account Info Dashboard      │   │
│  │  │ • Open Price           │  │  │  ┌────────────────────────┐ │   │
│  │  │ • Current Price        │  │  │  │ Health Indicator       │ │   │
│  │  │ • SL/TP                │  │  │  │ • Margin level         │ │   │
│  │  │ • Real-time P&L        │  │  │  │ • Color coded          │ │   │
│  │  │ • Actions              │  │  │  │ • Utilization bar      │ │   │
│  │  └────────────────────────┘  │  │  └────────────────────────┘ │   │
│  │  ┌────────────────────────┐  │  │  ┌────────────────────────┐ │   │
│  │  │ Total P&L Summary      │  │  │  │ Metric Cards           │ │   │
│  │  │ Close All Button       │  │  │  │ • Balance              │ │   │
│  │  └────────────────────────┘  │  │  │ • Equity               │ │   │
│  │  ┌────────────────────────┐  │  │  │ • Used Margin          │ │   │
│  │  │ Modify SL/TP Modal     │  │  │  │ • Free Margin          │ │   │
│  │  │ (conditional)          │  │  │  └────────────────────────┘ │   │
│  │  └────────────────────────┘  │  │  ┌────────────────────────┐ │   │
│  └──────────────────────────────┘  │  │ Unrealized P&L         │ │   │
│                                     │  │ Position Summary       │ │   │
│  ┌──────────────────────────────┐  │  │ Risk Warning           │ │   │
│  │  Bottom Dock                 │  │  │ (conditional)          │ │   │
│  │  • Positions Tab             │  │  └────────────────────────┘ │   │
│  │  • Orders Tab                │  └──────────────────────────────┘   │
│  │  • History Tab               │                                      │
│  │  • Ledger Tab                │  ┌──────────────────────────────┐   │
│  │  • Resizable height          │  │  Symbol Selector             │   │
│  └──────────────────────────────┘  │  • Dropdown list             │   │
│                                     │  • Search functionality      │   │
│  ┌──────────────────────────────┐  │  • Real-time spread display  │   │
│  │  Header/Navigation           │  └──────────────────────────────┘   │
│  │  • Broker name               │                                      │
│  │  • Connection status         │  ┌──────────────────────────────┐   │
│  │  • Account selector          │  │  Error Boundary              │   │
│  │  • Logout button             │  │  • Wraps all components      │   │
│  └──────────────────────────────┘  │  • Graceful error handling   │   │
│                                     └──────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

## Admin Panel Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    page.tsx (Admin Dashboard)                           │
│                                                                          │
│  ┌───────────────┐  ┌──────────────────────────────────────────────┐   │
│  │   Sidebar     │  │           Main Content Area                  │   │
│  │               │  │                                              │   │
│  │ Navigation:   │  │  ┌────────────────────────────────────────┐ │   │
│  │               │  │  │    Header                              │ │   │
│  │ • Accounts    │  │  │    • Page title                        │ │   │
│  │ • Ledger      │  │  │    • Refresh button                    │ │   │
│  │ • Routing     │  │  └────────────────────────────────────────┘ │   │
│  │ • LP Status   │  │                                              │   │
│  │ • LP Manage   │  │  ┌────────────────────────────────────────┐ │   │
│  │ • Feeds       │  │  │    Dynamic Content (Tab-Based)         │ │   │
│  │ • Risk        │  │  │                                        │ │   │
│  │ • Settings    │  │  │  ┌──────────────────────────────────┐ │ │   │
│  │               │  │  │  │  Accounts View                   │ │ │   │
│  └───────────────┘  │  │  │  • Account table                 │ │ │   │
│                     │  │  │  • Search/filter                 │ │ │   │
│  ┌───────────────┐  │  │  │  • Action buttons                │ │ │   │
│  │   Footer      │  │  │  │  • Selected account panel        │ │ │   │
│  │               │  │  │  └──────────────────────────────────┘ │ │   │
│  │ • Status      │  │  │                                        │ │   │
│  │ • B-Book      │  │  │  ┌──────────────────────────────────┐ │ │   │
│  │   Online      │  │  │  │  Ledger View                     │ │ │   │
│  └───────────────┘  │  │  │  • Transaction table             │ │ │   │
│                     │  │  │  • Filters (date, type, account) │ │ │   │
│                     │  │  │  • Export functionality          │ │ │   │
│                     │  │  │  • Pagination                    │ │ │   │
│                     │  │  └──────────────────────────────────┘ │ │   │
│                     │  │                                        │ │   │
│                     │  │  ┌──────────────────────────────────┐ │ │   │
│                     │  │  │  LP Status View                  │ │ │   │
│                     │  │  │  • LP list with status           │ │ │   │
│                     │  │  │  • Connection indicators         │ │ │   │
│                     │  │  │  • Quote aggregation stats       │ │ │   │
│                     │  │  │  • Tick rates                    │ │ │   │
│                     │  │  └──────────────────────────────────┘ │ │   │
│                     │  │                                        │ │   │
│                     │  │  ┌──────────────────────────────────┐ │ │   │
│                     │  │  │  LP Management View              │ │ │   │
│                     │  │  │  • Add LP form                   │ │ │   │
│                     │  │  │  • LP configuration              │ │ │   │
│                     │  │  │  • Symbol mapping                │ │ │   │
│                     │  │  │  • Enable/disable toggles        │ │ │   │
│                     │  │  └──────────────────────────────────┘ │ │   │
│                     │  │                                        │ │   │
│                     │  │  ┌──────────────────────────────────┐ │ │   │
│                     │  │  │  Symbols View                    │ │ │   │
│                     │  │  │  • Symbol list                   │ │ │   │
│                     │  │  │  • Enable/disable toggles        │ │ │   │
│                     │  │  │  • Spread configuration          │ │ │   │
│                     │  │  │  • Commission settings           │ │ │   │
│                     │  │  └──────────────────────────────────┘ │ │   │
│                     │  │                                        │ │   │
│                     │  │  ┌──────────────────────────────────┐ │ │   │
│                     │  │  │  Risk View                       │ │ │   │
│                     │  │  │  • Exposure dashboard            │ │ │   │
│                     │  │  │  • Position heatmap              │ │ │   │
│                     │  │  │  • Margin alerts                 │ │ │   │
│                     │  │  │  • P&L tracking                  │ │ │   │
│                     │  │  └──────────────────────────────────┘ │ │   │
│                     │  │                                        │ │   │
│                     │  │  ┌──────────────────────────────────┐ │ │   │
│                     │  │  │  Settings View                   │ │ │   │
│                     │  │  │  • Execution mode toggle         │ │ │   │
│                     │  │  │  • LP selection                  │ │ │   │
│                     │  │  │  • Broker configuration          │ │ │   │
│                     │  │  │  • System settings               │ │ │   │
│                     │  │  │  • Restart backend button        │ │ │   │
│                     │  │  └──────────────────────────────────┘ │ │   │
│                     │  │                                        │ │   │
│                     │  └────────────────────────────────────────┘ │   │
│                     └──────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │                      Modals (Conditional)                         │  │
│  │                                                                   │  │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────────┐ │  │
│  │  │ Deposit Modal  │  │ Withdraw Modal │  │ Adjust Modal       │ │  │
│  │  │ • Amount input │  │ • Amount input │  │ • Amount input     │ │  │
│  │  │ • Method       │  │ • Method       │  │ • Reason           │ │  │
│  │  │ • Reference    │  │ • Reference    │  │ • Category         │ │  │
│  │  │ • Submit btn   │  │ • Submit btn   │  │ • Submit btn       │ │  │
│  │  └────────────────┘  └────────────────┘  └────────────────────┘ │  │
│  │                                                                   │  │
│  │  ┌────────────────┐  ┌────────────────┐                         │  │
│  │  │ Bonus Modal    │  │ Password Modal │                         │  │
│  │  │ • Amount input │  │ • New password │                         │  │
│  │  │ • Type         │  │ • Confirm      │                         │  │
│  │  │ • Description  │  │ • Submit btn   │                         │  │
│  │  │ • Submit btn   │  └────────────────┘                         │  │
│  │  └────────────────┘                                              │  │
│  └──────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Backend (Go Server)                            │
│                          Port: 7999                                     │
└───────────┬─────────────────────────────────────────────────┬──────────┘
            │                                                 │
            │ WebSocket (/ws)                                 │ HTTP REST
            │ Real-time ticks                                 │ API calls
            │                                                 │
    ┌───────▼──────────┐                             ┌────────▼─────────┐
    │                  │                             │                  │
    │  WebSocket       │                             │   API Service    │
    │  Service         │                             │   Layer          │
    │                  │                             │                  │
    │  • connect()     │                             │  • fetch()       │
    │  • subscribe()   │                             │  • POST/GET      │
    │  • disconnect()  │                             │  • error handle  │
    │  • onState()     │                             │  • retry logic   │
    │                  │                             │                  │
    └───────┬──────────┘                             └────────┬─────────┘
            │                                                 │
            │ publishes events                                │ returns data
            │                                                 │
    ┌───────▼─────────────────────────────────────────────────▼─────────┐
    │                                                                    │
    │                     Zustand Global Store                          │
    │                                                                    │
    │  • setTick(symbol, tick) ──────────► ticks: Record<symbol, Tick> │
    │  • setPositions(positions) ────────► positions: Position[]       │
    │  • setOrders(orders) ───────────────► orders: Order[]            │
    │  • setAccount(account) ─────────────► account: Account           │
    │  • setSelectedSymbol(symbol) ───────► selectedSymbol: string     │
    │  • setWsConnected(bool) ────────────► wsConnected: boolean       │
    │  • setLoadingStates({...}) ─────────► loading states             │
    │                                                                    │
    └───────────────┬──────────────────────────────────────────────────┘
                    │
                    │ consumed by components via useAppStore() hook
                    │
    ┌───────────────▼──────────────────────────────────────────────────┐
    │                                                                    │
    │                    React Components                                │
    │                                                                    │
    │  ┌──────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
    │  │ OrderEntryPanel  │  │ PositionList     │  │ AccountInfo     │ │
    │  │                  │  │                  │  │ Dashboard       │ │
    │  │ const {          │  │ const {          │  │ const {         │ │
    │  │   ticks,         │  │   positions,     │  │   account,      │ │
    │  │   accountId,     │  │   ticks,         │  │   positions,    │ │
    │  │   orderVolume,   │  │   accountId      │  │ } = useAppStore │ │
    │  │   setLoading     │  │ } = useAppStore  │  │                 │ │
    │  │ } = useAppStore  │  │                  │  │                 │ │
    │  │                  │  │                  │  │                 │ │
    │  │ // Use data      │  │ // Calc live P&L │  │ // Show metrics │ │
    │  │ // Place orders  │  │ // Close pos     │  │ // Health calc  │ │
    │  └──────────────────┘  └──────────────────┘  └─────────────────┘ │
    │                                                                    │
    │  ┌──────────────────┐  ┌──────────────────┐  ┌─────────────────┐ │
    │  │ TradingChart     │  │ SymbolSelector   │  │ BottomDock      │ │
    │  │                  │  │                  │  │                 │ │
    │  │ const {          │  │ const {          │  │ const {         │ │
    │  │   ticks,         │  │   selected,      │  │   positions,    │ │
    │  │   selectedSymbol,│  │   setSelected    │  │   orders,       │ │
    │  │   timeframe,     │  │ } = useAppStore  │  │   trades        │ │
    │  │   chartType      │  │                  │  │ } = useAppStore │ │
    │  │ } = useAppStore  │  │                  │  │                 │ │
    │  │                  │  │                  │  │                 │ │
    │  │ // Render chart  │  │ // Show list     │  │ // Tab panels   │ │
    │  │ // Update live   │  │ // Handle change │  │ // Resize       │ │
    │  └──────────────────┘  └──────────────────┘  └─────────────────┘ │
    │                                                                    │
    └────────────────────────────────────────────────────────────────────┘
```

## Component Lifecycle

### Client Trading Interface Startup

```
1. App.tsx mounts
   │
   ├─► Initialize Zustand store
   │   └─► Load persisted preferences from localStorage
   │
   ├─► Initialize WebSocket service
   │   ├─► Connect to ws://localhost:7999/ws
   │   ├─► Register state change listener
   │   ├─► Subscribe to '*' (all symbols)
   │   └─► Start ping interval (30s)
   │
   ├─► User Login Component
   │   ├─► Enter account ID
   │   └─► Call setAuthenticated(true, accountId)
   │
   └─► After Authentication
       │
       ├─► Fetch Account Data (interval: 1s)
       │   ├─► GET /api/account/summary?accountId={id}
       │   ├─► GET /api/positions?accountId={id}
       │   ├─► GET /api/trades?accountId={id}
       │   └─► GET /api/ledger?accountId={id}
       │
       ├─► WebSocket receives ticks
       │   ├─► Buffer in Map<symbol, tick>
       │   ├─► Flush every 100ms
       │   └─► Update Zustand store → setTick(symbol, tick)
       │
       └─► Components render with live data
           ├─► OrderEntryPanel shows current bid/ask
           ├─► PositionList calculates live P&L
           ├─► AccountInfoDashboard updates metrics
           └─► TradingChart updates in real-time
```

### Order Placement Flow

```
1. User clicks BUY/SELL in OrderEntryPanel
   │
   ├─► Validate inputs (volume, prices, SL/TP)
   │
   ├─► Set loading state: setLoadingStates({ isPlacingOrder: true })
   │
   ├─► Build request body
   │   ├─► accountId
   │   ├─► symbol
   │   ├─► side (BUY/SELL)
   │   ├─► volume
   │   ├─► price (for limit/stop orders)
   │   ├─► sl (optional)
   │   └─► tp (optional)
   │
   ├─► POST to appropriate endpoint
   │   ├─► Market: POST /api/orders/market
   │   ├─► Limit: POST /order/limit
   │   ├─► Stop: POST /order/stop
   │   └─► Stop-Limit: POST /order/stop-limit
   │
   ├─► Backend processes order
   │   ├─► Validates account balance
   │   ├─► Checks margin requirements
   │   ├─► Executes trade (B-Book) or routes to LP (A-Book)
   │   └─► Updates position in database
   │
   ├─► Response received
   │   ├─► Success: Show success alert, reset form
   │   └─► Error: Show error alert with message
   │
   ├─► Set loading state: setLoadingStates({ isPlacingOrder: false })
   │
   └─► Next polling cycle (1s) fetches updated positions
       └─► PositionList re-renders with new position
```

### Real-Time P&L Update Flow

```
1. WebSocket receives new tick
   │
   ├─► Tick buffered in websocket.ts
   │
   ├─► Flush interval (100ms) triggers
   │   └─► setTick(symbol, tick) in Zustand store
   │
   ├─► PositionList.tsx subscribes to store changes
   │   └─► useAppStore() hook detects tick update
   │
   ├─► useMemo recalculates enrichedPositions
   │   ├─► For each position:
   │   │   ├─► Get current tick for position.symbol
   │   │   ├─► Use bid for LONG, ask for SHORT
   │   │   ├─► Calculate price difference
   │   │   ├─► Calculate P&L = diff * volume * contract size
   │   │   └─► Subtract commission and swap
   │   └─► Store as { ...position, livePrice, livePnL }
   │
   ├─► Component re-renders
   │   ├─► Table rows update with new current prices
   │   ├─► P&L cells update with new values
   │   └─► Colors update (green/red based on profit/loss)
   │
   └─► Total P&L summary updates
       └─► Aggregates all position P&L values
```

## File Structure

```
trading-engine/
├── backend/
│   ├── api/
│   │   └── server.go              # HTTP handlers
│   ├── ws/
│   │   └── hub.go                 # WebSocket hub
│   └── cmd/
│       └── server/
│           └── main.go            # Entry point
│
├── clients/
│   └── desktop/
│       ├── src/
│       │   ├── services/
│       │   │   └── websocket.ts   # WebSocket service (NEW)
│       │   ├── store/
│       │   │   └── useAppStore.ts # Zustand store (NEW)
│       │   ├── components/
│       │   │   ├── OrderEntryPanel.tsx         # NEW
│       │   │   ├── PositionList.tsx            # NEW
│       │   │   ├── AccountInfoDashboard.tsx    # NEW
│       │   │   ├── TradingChart.tsx            # Existing
│       │   │   ├── BottomDock.tsx              # Existing
│       │   │   ├── Login.tsx                   # Existing
│       │   │   └── ErrorBoundary.tsx           # Existing
│       │   ├── App.tsx                         # Main app
│       │   └── main.tsx                        # Entry point
│       ├── package.json
│       └── vite.config.ts
│
├── admin/
│   └── broker-admin/
│       ├── src/
│       │   ├── app/
│       │   │   ├── page.tsx                    # Main dashboard
│       │   │   └── layout.tsx
│       │   ├── components/
│       │   │   ├── dashboard/
│       │   │   │   ├── AccountsView.tsx
│       │   │   │   ├── LedgerView.tsx
│       │   │   │   ├── LPStatusView.tsx
│       │   │   │   ├── LPManagementView.tsx
│       │   │   │   ├── RiskView.tsx
│       │   │   │   ├── RoutingView.tsx
│       │   │   │   ├── SettingsView.tsx
│       │   │   │   └── SymbolsView.tsx
│       │   │   └── ui/
│       │   │       ├── Modal.tsx
│       │   │       └── NavItem.tsx
│       │   └── types/
│       │       └── index.ts
│       ├── package.json
│       └── next.config.ts
│
└── docs/
    ├── FRONTEND_IMPLEMENTATION.md      # Complete guide
    ├── QUICK_START.md                  # Setup instructions
    ├── IMPLEMENTATION_SUMMARY.md       # What was built
    └── COMPONENT_ARCHITECTURE.md       # This file
```

## Technology Choices Explained

### Why Zustand over Redux?
- **Simpler API**: Less boilerplate, easier to learn
- **Better TypeScript**: First-class TypeScript support
- **Smaller Bundle**: ~1KB vs 5KB+ for Redux
- **No Provider**: Direct imports, no context wrapping
- **DevTools**: Still has Redux DevTools integration

### Why Lightweight Charts over TradingView?
- **Free & Open Source**: No licensing fees
- **Lightweight**: Smaller bundle size
- **Mobile Friendly**: Better touch support
- **Customizable**: Full control over appearance
- **Real-Time Optimized**: Designed for live data

### Why Native WebSocket over Socket.io?
- **Lighter**: No additional library needed
- **Control**: Full control over reconnection logic
- **Backend Simplicity**: Go's native WebSocket support
- **Performance**: Lower overhead, faster

### Why Vite over Create React App?
- **Faster**: Lightning-fast dev server
- **Modern**: ESM-based, optimized for modern browsers
- **Smaller**: Tree-shaking, better bundle optimization
- **HMR**: Near-instant hot module replacement

## Summary

This architecture provides:

✅ **Separation of Concerns**: Services, state, and components are clearly separated
✅ **Scalability**: Easy to add new components and features
✅ **Maintainability**: Clear data flow, single source of truth
✅ **Performance**: Optimized with throttling, memoization, and efficient state updates
✅ **Developer Experience**: TypeScript, hot reload, DevTools integration
✅ **Production Ready**: Error boundaries, loading states, reconnection logic
