# Trading Components Setup Guide

## Quick Start

The React frontend components are now complete and ready to use. Here's how to integrate them into your application.

## Components Available

All components are located in `/clients/desktop/src/components/`:

1. **TradingChart.tsx** - Real-time charting with TradingView-like features
2. **OrderEntry.tsx** - Advanced order placement with risk management
3. **PositionList.tsx** - Live positions table with P&L tracking
4. **AccountInfoDashboard.tsx** - Account metrics and health indicators
5. **OrderBook.tsx** - Live order book depth visualization
6. **TradeHistory.tsx** - Closed trades with analytics and export
7. **AdminPanel.tsx** - Broker configuration and LP management

## Installation

All dependencies are already configured in `package.json`:

```bash
cd clients/desktop
bun install
```

## Running the Application

```bash
# Development mode
bun run dev

# Production build
bun run build

# Preview production build
bun run preview
```

## Integration Examples

### Basic Trading View

```tsx
import { TradingChart, OrderEntry, AccountInfoDashboard } from './components';
import { useAppStore } from './store/useAppStore';

function TradingView() {
  const { selectedSymbol, ticks, account, accountId } = useAppStore();
  const currentTick = ticks[selectedSymbol];

  return (
    <div className="flex gap-4 h-screen p-4">
      {/* Chart */}
      <div className="flex-1">
        <TradingChart
          symbol={selectedSymbol}
          currentPrice={currentTick}
        />
      </div>

      {/* Sidebar */}
      <div className="w-80 flex flex-col gap-4">
        <AccountInfoDashboard />
        <OrderEntry
          symbol={selectedSymbol}
          currentBid={currentTick?.bid}
          currentAsk={currentTick?.ask}
          accountId={parseInt(accountId || '1')}
          balance={account?.balance || 0}
        />
      </div>
    </div>
  );
}
```

### Complete Dashboard

See `src/examples/TradingDashboard.tsx` for a full implementation.

```tsx
import { TradingDashboard } from './examples/TradingDashboard';

function App() {
  return <TradingDashboard />;
}
```

## Backend Integration

The components expect these endpoints to be available:

### WebSocket Connection
```
ws://localhost:8080/ws
```

Messages format:
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.10550,
  "ask": 1.10552,
  "spread": 0.00002,
  "timestamp": 1705680000
}
```

### REST API Endpoints

**Trading:**
- `POST /api/orders/market` - Market order
- `POST /api/orders/limit` - Limit order
- `POST /api/orders/stop` - Stop order
- `POST /api/positions/close` - Close position
- `POST /api/positions/modify` - Modify SL/TP

**Data:**
- `GET /api/account/summary?accountId={id}`
- `GET /api/positions?accountId={id}`
- `GET /api/trades/history?accountId={id}`
- `GET /ohlc?symbol={symbol}&timeframe={tf}&limit={n}`
- `GET /api/orderbook?symbol={symbol}&depth={n}`

**Admin:**
- `GET /api/admin/config`
- `POST /api/admin/config`
- `GET /api/admin/liquidity-providers`
- `GET /api/admin/symbols`

## State Management

Components use Zustand store located at `src/store/useAppStore.ts`.

The store manages:
- Market data (ticks, selected symbol)
- Trading data (positions, orders, trades)
- Account information
- UI state

### Adding Data

```typescript
const { setTick, setPositions, setAccount } = useAppStore();

// Update tick
setTick('EURUSD', {
  symbol: 'EURUSD',
  bid: 1.10550,
  ask: 1.10552,
  spread: 0.00002,
  timestamp: Date.now()
});

// Update positions
setPositions([
  {
    id: 1,
    symbol: 'EURUSD',
    side: 'BUY',
    volume: 0.1,
    openPrice: 1.10500,
    currentPrice: 1.10550,
    sl: 1.10400,
    tp: 1.10700,
    unrealizedPnL: 50.00
  }
]);

// Update account
setAccount({
  balance: 10000,
  equity: 10050,
  margin: 100,
  freeMargin: 9950,
  marginLevel: 10050,
  unrealizedPL: 50
});
```

## Styling Customization

Components use Tailwind CSS. Customize in `tailwind.config.js`:

```javascript
export default {
  theme: {
    extend: {
      colors: {
        // Custom brand colors
        brand: {
          primary: '#10b981', // emerald-500
          danger: '#ef4444',  // red-500
        }
      }
    }
  }
}
```

## Performance Tips

1. **Tick Throttling:** Already implemented (10 FPS updates)
2. **Large Position Lists:** Consider adding virtualization
3. **Chart Performance:** Limit visible candles to 500-1000
4. **WebSocket:** Auto-reconnect is built-in

## Testing Components

Each component can be tested individually:

```tsx
import { render, screen } from '@testing-library/react';
import { OrderEntry } from './components/OrderEntry';

test('renders order entry form', () => {
  render(
    <OrderEntry
      symbol="EURUSD"
      currentBid={1.10550}
      currentAsk={1.10552}
      accountId={1}
      balance={10000}
    />
  );

  expect(screen.getByText('Order Entry')).toBeInTheDocument();
});
```

## Common Issues

### WebSocket Not Connecting
```typescript
// Check WebSocket URL in App.tsx or WebSocket service
const ws = new WebSocket('ws://localhost:8080/ws');

// Add error handling
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};
```

### Positions Not Updating
```typescript
// Ensure tick data is flowing
const { ticks } = useAppStore();
console.log('Current ticks:', ticks);

// Check if positions array is populated
const { positions } = useAppStore();
console.log('Positions:', positions);
```

### Chart Not Rendering
```typescript
// Verify lightweight-charts is installed
import { createChart } from 'lightweight-charts';

// Check if OHLC endpoint is responding
fetch('http://localhost:8080/ohlc?symbol=EURUSD&timeframe=1m&limit=100')
  .then(res => res.json())
  .then(data => console.log('OHLC data:', data));
```

## Production Checklist

- [ ] Environment variables configured
- [ ] WebSocket URL updated for production
- [ ] API endpoints use production URLs
- [ ] Error boundaries implemented
- [ ] Analytics tracking added
- [ ] Performance monitoring enabled
- [ ] Build optimized (`bun run build`)
- [ ] Types checked (`tsc --noEmit`)
- [ ] Components tested

## Next Steps

1. **Run the app:** `bun run dev`
2. **Open browser:** http://localhost:5173
3. **Login** with test account
4. **Test components** individually
5. **Customize styling** as needed
6. **Connect to backend** API
7. **Deploy** when ready

## Support

For component-specific documentation, see:
- `/src/components/README.md` - Detailed component docs
- `/src/examples/TradingDashboard.tsx` - Complete example

## Architecture

```
clients/desktop/
├── src/
│   ├── components/          # All UI components
│   │   ├── TradingChart.tsx
│   │   ├── OrderEntry.tsx
│   │   ├── PositionList.tsx
│   │   ├── AccountInfoDashboard.tsx
│   │   ├── OrderBook.tsx
│   │   ├── TradeHistory.tsx
│   │   ├── AdminPanel.tsx
│   │   ├── index.ts         # Component exports
│   │   └── README.md        # Component docs
│   ├── store/
│   │   └── useAppStore.ts   # Zustand store
│   ├── services/
│   │   ├── api.ts           # API client
│   │   ├── websocket.ts     # WebSocket service
│   │   └── notifications.ts # Toast notifications
│   ├── examples/
│   │   └── TradingDashboard.tsx  # Complete example
│   ├── App.tsx              # Main app
│   └── main.tsx             # Entry point
└── package.json
```

## License

Proprietary - Trading Engine Platform
