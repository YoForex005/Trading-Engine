# Trading Interface Components

Production-ready React components for a professional trading platform built with React 19.2.0, TypeScript, Zustand, and Tailwind CSS.

## Components Overview

### 1. TradingChart.tsx
TradingView-style chart with real-time updates using lightweight-charts.

**Features:**
- Multiple chart types: Candlestick, Heikin Ashi, Bar, Line, Area
- Real-time price updates via WebSocket
- Multiple timeframes: M1, M5, M15, H1, H4, D1
- Interactive position overlays (Entry, SL, TP)
- Drag-and-drop SL/TP modification
- Historical OHLC data loading
- Responsive and performant

**Usage:**
```tsx
import { TradingChart } from './components/TradingChart';

<TradingChart
  symbol="EURUSD"
  currentPrice={{ bid: 1.10550, ask: 1.10552 }}
  chartType="candlestick"
  timeframe="1m"
  positions={positions}
  onClosePosition={(id) => console.log('Close', id)}
  onModifyPosition={(id, sl, tp) => console.log('Modify', id, sl, tp)}
/>
```

### 2. OrderEntry.tsx
Advanced order placement form with risk management.

**Features:**
- Order types: Market, Limit, Stop, Stop-Limit
- Real-time margin preview
- Built-in risk calculator (position sizing)
- SL/TP configuration
- Volume validation
- Live bid/ask prices

**Usage:**
```tsx
import { OrderEntry } from './components/OrderEntry';

<OrderEntry
  symbol="EURUSD"
  currentBid={1.10550}
  currentAsk={1.10552}
  accountId={1}
  balance={10000}
  onOrderPlaced={() => console.log('Order placed')}
/>
```

### 3. PositionList.tsx
Real-time positions table with P&L tracking.

**Features:**
- Live P&L updates from tick data
- Position modification (SL/TP)
- Single position close
- Bulk close operations (All, Winners, Losers)
- Commission and swap tracking
- Responsive table layout

**Usage:**
```tsx
import { PositionList } from './components/PositionList';

<PositionList />
```

### 4. AccountInfoDashboard.tsx
Account metrics with visual health indicators.

**Features:**
- Real-time balance, equity, margin display
- Margin level indicator with color coding
- Margin utilization progress bar
- Risk warnings for high margin usage
- Position summary
- Health status (Healthy, Good, Warning, Critical)

**Usage:**
```tsx
import { AccountInfoDashboard } from './components/AccountInfoDashboard';

<AccountInfoDashboard />
```

### 5. OrderBook.tsx
Live order book with bid/ask depth visualization.

**Features:**
- Real-time order book updates
- Configurable depth (10, 20, 50 levels)
- Volume visualization bars
- Spread display with pip calculation
- Total volume aggregation
- Mock data fallback for development

**Usage:**
```tsx
import { OrderBook } from './components/OrderBook';

<OrderBook
  symbol="EURUSD"
  currentBid={1.10550}
  currentAsk={1.10552}
/>
```

### 6. TradeHistory.tsx
Closed trades history with analytics.

**Features:**
- Trade history table with filtering
- Search by symbol
- Filter by side (BUY/SELL)
- Filter by result (Profit/Loss)
- Date range filtering
- Statistics dashboard (Win rate, P&L, etc.)
- CSV export functionality
- Commission and swap tracking

**Usage:**
```tsx
import { TradeHistory } from './components/TradeHistory';

<TradeHistory accountId="1" />
```

### 7. AdminPanel.tsx
Administrative controls for broker configuration.

**Features:**
- Broker configuration (Routing mode, Execution, Leverage)
- Liquidity provider management
- Symbol configuration (Enable/disable, Spreads, Commission)
- Real-time LP status monitoring
- Save/Load configuration
- Tabbed interface for organization

**Usage:**
```tsx
import { AdminPanel } from './components/AdminPanel';

<AdminPanel />
```

## State Management

All components use Zustand for state management via `useAppStore`:

```typescript
import { useAppStore } from '../store/useAppStore';

const {
  // Market Data
  ticks,
  selectedSymbol,

  // Trading
  positions,
  orders,
  trades,

  // Account
  account,
  accountId,

  // Actions
  setTick,
  setPositions,
  setAccount,
  // ... more actions
} = useAppStore();
```

## API Integration

Components expect these backend endpoints:

### Trading
- `POST /api/orders/market` - Place market order
- `POST /api/orders/limit` - Place limit order
- `POST /api/orders/stop` - Place stop order
- `POST /api/positions/close` - Close position
- `POST /api/positions/modify` - Modify SL/TP
- `POST /api/positions/close-bulk` - Bulk close

### Data
- `GET /api/account/summary?accountId={id}` - Account info
- `GET /api/positions?accountId={id}` - Open positions
- `GET /api/trades/history?accountId={id}` - Trade history
- `GET /ohlc?symbol={symbol}&timeframe={tf}&limit={n}` - OHLC data
- `GET /api/orderbook?symbol={symbol}&depth={n}` - Order book
- `WS /ws` - WebSocket for real-time ticks

### Admin
- `GET /api/admin/config` - Broker configuration
- `POST /api/admin/config` - Save configuration
- `GET /api/admin/liquidity-providers` - LP list
- `POST /api/admin/lp/{name}/toggle` - Toggle LP
- `GET /api/admin/symbols` - Symbol configs
- `PATCH /api/admin/symbols/{symbol}` - Update symbol

## Styling

Components use Tailwind CSS with a dark theme:

**Color Palette:**
- Background: `bg-[#09090b]`, `bg-zinc-900`
- Borders: `border-zinc-800`, `border-zinc-700`
- Text: `text-zinc-300`, `text-zinc-500`
- Buy/Long: `text-emerald-400`, `bg-emerald-500`
- Sell/Short: `text-red-400`, `bg-red-500`
- Info: `text-blue-400`

## WebSocket Integration

Real-time data via WebSocket:

```typescript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'tick') {
    setTick(data.symbol, data);
  }
};
```

## TypeScript Types

All components are fully typed. Key interfaces:

```typescript
interface Tick {
  symbol: string;
  bid: number;
  ask: number;
  spread: number;
  timestamp: number;
}

interface Position {
  id: number;
  symbol: string;
  side: 'BUY' | 'SELL';
  volume: number;
  openPrice: number;
  currentPrice: number;
  sl: number;
  tp: number;
  unrealizedPnL: number;
}

interface Account {
  balance: number;
  equity: number;
  margin: number;
  freeMargin: number;
  marginLevel: number;
  unrealizedPL: number;
}
```

## Performance Optimizations

1. **Tick Buffering:** Throttled updates at 10 FPS (100ms intervals)
2. **Memoization:** `useMemo` for expensive calculations
3. **Virtual Scrolling:** For large lists (consider react-virtual)
4. **Lazy Loading:** Code splitting for admin panel
5. **WebSocket Reconnection:** Automatic with exponential backoff

## Testing

Each component can be tested independently:

```bash
# Run tests
bun run test

# Test specific component
bun run test -- "TradingChart"
```

## Example Usage

See `src/examples/TradingDashboard.tsx` for a complete example integrating all components.

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Dependencies

- React 19.2.0
- TypeScript 5.9.3
- Zustand 5.0.2
- Tailwind CSS 4.1.18
- lightweight-charts 5.1.0
- lucide-react 0.562.0

## License

Proprietary - Trading Engine Platform
