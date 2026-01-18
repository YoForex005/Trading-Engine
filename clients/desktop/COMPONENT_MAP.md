# Component Architecture Map

## Component Hierarchy

```
TradingDashboard (Example)
├── Sidebar Navigation
│   ├── Trading View
│   ├── Positions View
│   ├── Order Book View
│   ├── History View
│   └── Admin View
│
└── Content Area
    ├── Trading View
    │   ├── TradingChart
    │   │   ├── ChartControls
    │   │   ├── Candlestick/Line/Area Series
    │   │   └── Position Overlays (Entry/SL/TP)
    │   ├── AccountInfoDashboard
    │   │   ├── Account Health Indicator
    │   │   ├── Balance/Equity Display
    │   │   ├── Margin Statistics
    │   │   └── Risk Warnings
    │   └── OrderEntry
    │       ├── Order Type Selector
    │       ├── Buy/Sell Buttons
    │       ├── Volume Input
    │       ├── SL/TP Inputs
    │       ├── Risk Calculator
    │       └── Margin Preview
    │
    ├── Positions View
    │   └── PositionList
    │       ├── Position Table
    │       ├── Live P&L Updates
    │       ├── Close/Modify Actions
    │       └── Bulk Close Controls
    │
    ├── Order Book View
    │   └── OrderBook
    │       ├── Depth Selector
    │       ├── Asks (Sell Orders)
    │       ├── Current Market Price
    │       ├── Bids (Buy Orders)
    │       └── Volume Statistics
    │
    ├── History View
    │   └── TradeHistory
    │       ├── Statistics Dashboard
    │       ├── Filter Controls
    │       ├── Trade Table
    │       └── CSV Export
    │
    └── Admin View
        └── AdminPanel
            ├── Config Tab
            │   └── Broker Settings
            ├── LP Tab
            │   └── Liquidity Providers
            └── Symbols Tab
                └── Symbol Configuration
```

## Data Flow

```
WebSocket (ws://localhost:8080/ws)
    │
    ├─► Tick Updates
    │       │
    │       └─► useAppStore.setTick()
    │               │
    │               ├─► TradingChart (price updates)
    │               ├─► OrderEntry (current prices)
    │               ├─► OrderBook (market price)
    │               └─► PositionList (P&L calculation)
    │
REST API (http://localhost:8080/api)
    │
    ├─► Account Data (/account/summary)
    │       │
    │       └─► useAppStore.setAccount()
    │               │
    │               ├─► AccountInfoDashboard
    │               └─► OrderEntry (margin preview)
    │
    ├─► Positions (/positions)
    │       │
    │       └─► useAppStore.setPositions()
    │               │
    │               ├─► PositionList
    │               └─► TradingChart (overlays)
    │
    ├─► Trade History (/trades/history)
    │       │
    │       └─► TradeHistory
    │
    ├─► OHLC Data (/ohlc)
    │       │
    │       └─► TradingChart (historical candles)
    │
    ├─► Order Book (/orderbook)
    │       │
    │       └─► OrderBook
    │
    └─► Admin APIs (/admin/*)
            │
            └─► AdminPanel
```

## State Management (Zustand)

```
useAppStore
│
├── Market Data
│   ├── ticks: Record<string, Tick>
│   └── selectedSymbol: string
│
├── Trading Data
│   ├── positions: Position[]
│   ├── orders: Order[]
│   └── trades: Trade[]
│
├── Account Data
│   ├── account: Account | null
│   └── accountId: string | null
│
├── UI State
│   ├── isChartMaximized: boolean
│   ├── chartType: ChartType
│   ├── timeframe: Timeframe
│   └── orderVolume: number
│
└── Actions
    ├── setTick()
    ├── setPositions()
    ├── setAccount()
    ├── setChartType()
    └── ... more
```

## Component Dependencies

```
TradingChart
├── lightweight-charts (charting library)
├── lucide-react (icons)
└── useAppStore (positions, ticks)

OrderEntry
├── lucide-react (icons)
├── api service (margin preview, lot calculation)
└── useAppStore (not directly used, props passed)

PositionList
├── lucide-react (icons)
├── useAppStore (positions, ticks, accountId)
└── API fetch (close, modify)

AccountInfoDashboard
├── lucide-react (icons)
└── useAppStore (account, positions)

OrderBook
├── lucide-react (icons)
└── API fetch (orderbook) + mock fallback

TradeHistory
├── lucide-react (icons)
├── useAppStore (accountId)
└── API fetch (trade history)

AdminPanel
├── lucide-react (icons)
└── API fetch (config, LPs, symbols)
```

## API Integration Points

### Real-time Updates (WebSocket)
```typescript
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  switch (data.type) {
    case 'tick':
      setTick(data.symbol, data);
      break;
    case 'position_update':
      setPositions(data.positions);
      break;
    case 'account_update':
      setAccount(data.account);
      break;
  }
};
```

### Data Fetching (REST)
```typescript
// Positions
fetch(`/api/positions?accountId=${id}`)
  .then(res => res.json())
  .then(positions => setPositions(positions));

// Account
fetch(`/api/account/summary?accountId=${id}`)
  .then(res => res.json())
  .then(account => setAccount(account));

// OHLC
fetch(`/ohlc?symbol=${symbol}&timeframe=${tf}&limit=500`)
  .then(res => res.json())
  .then(candles => updateChart(candles));
```

### Trading Actions (REST)
```typescript
// Place order
await api.orders.placeMarketOrder({
  accountId: 1,
  symbol: 'EURUSD',
  side: 'BUY',
  volume: 0.1,
  sl: 1.10400,
  tp: 1.10700
});

// Close position
await fetch('/api/positions/close', {
  method: 'POST',
  body: JSON.stringify({
    accountId: 1,
    positionId: 123
  })
});

// Modify position
await fetch('/api/positions/modify', {
  method: 'POST',
  body: JSON.stringify({
    accountId: 1,
    positionId: 123,
    sl: 1.10450,
    tp: 1.10750
  })
});
```

## Component Communication Patterns

### Parent → Child (Props)
```typescript
<TradingChart
  symbol={selectedSymbol}        // From parent
  currentPrice={currentTick}     // From parent
  positions={positions}          // From parent
/>
```

### Child → Parent (Callbacks)
```typescript
<OrderEntry
  onOrderPlaced={() => {         // Callback to parent
    refreshPositions();
    showNotification('Order placed');
  }}
/>
```

### Sibling Communication (Shared Store)
```typescript
// Component A updates store
const { setSelectedSymbol } = useAppStore();
setSelectedSymbol('GBPUSD');

// Component B reads from store
const { selectedSymbol } = useAppStore();
// selectedSymbol is now 'GBPUSD'
```

## Event Flow Example: Placing an Order

```
1. User clicks BUY button in OrderEntry
   │
   ├─► OrderEntry.handlePlaceOrder()
   │
   ├─► POST /api/orders/market
   │       │
   │       └─► Backend processes order
   │               │
   │               ├─► Creates position in database
   │               └─► Sends WebSocket update
   │
   ├─► WebSocket receives position_update
   │       │
   │       └─► useAppStore.setPositions()
   │               │
   │               ├─► PositionList re-renders (new position)
   │               └─► TradingChart updates (position overlay)
   │
   └─► OrderEntry.onOrderPlaced() callback
           │
           └─► Parent shows success notification
```

## Styling Architecture

```
Tailwind CSS Classes
│
├── Layout
│   ├── flex, grid (positioning)
│   ├── p-*, m-*, gap-* (spacing)
│   └── h-*, w-* (sizing)
│
├── Colors (Dark Theme)
│   ├── bg-zinc-900 (backgrounds)
│   ├── border-zinc-800 (borders)
│   ├── text-zinc-300 (text)
│   ├── text-emerald-400 (buy/profit)
│   └── text-red-400 (sell/loss)
│
├── Interactive States
│   ├── hover:bg-zinc-800
│   ├── focus:outline-none
│   ├── disabled:opacity-50
│   └── transition-colors
│
└── Responsive
    ├── md:grid-cols-2 (medium screens)
    └── lg:grid-cols-3 (large screens)
```

## Performance Optimizations

```
1. Tick Buffering (100ms interval)
   ├─► Prevents UI lag from high-frequency updates
   └─► Implemented in App.tsx WebSocket handler

2. Memoization
   ├─► useMemo for expensive calculations
   ├─► Used in PositionList (P&L calculation)
   └─► Used in TradeHistory (statistics)

3. Conditional Rendering
   ├─► Only render visible components
   └─► Tab-based views in AdminPanel

4. Debounced API Calls
   ├─► Volume input in OrderEntry
   └─► Search in TradeHistory

5. WebSocket Reconnection
   ├─► Auto-reconnect with backoff
   └─► Prevents connection loss
```

This architecture provides a scalable, maintainable foundation for the trading platform.
