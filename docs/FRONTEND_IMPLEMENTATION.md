# Frontend Implementation Guide

## Overview

This document describes the enhanced frontend architecture for the RTX Trading Engine, including both the client trading interface and the admin control panel.

## Architecture

### Technology Stack

#### Client Trading Interface
- **Framework**: React 19 + TypeScript
- **Build Tool**: Vite 7
- **State Management**: Zustand 5.x
- **UI Library**: Tailwind CSS 4
- **Icons**: Lucide React
- **Charting**: Lightweight Charts 5.x
- **WebSocket**: Native WebSocket with custom reconnection logic

#### Admin Panel
- **Framework**: Next.js 16 + React 19
- **UI Library**: Tailwind CSS 4
- **Icons**: Lucide React
- **State**: React Hooks + Server Components

## Client Trading Interface

### Features

#### 1. Real-Time Market Data
- **WebSocket Service** (`src/services/websocket.ts`):
  - Auto-reconnection with exponential backoff
  - Tick buffering and throttling (100ms)
  - Subscription-based architecture
  - Connection state monitoring
  - Ping/keepalive mechanism

#### 2. Order Entry Panel
- **Market Orders**: Instant execution at current bid/ask
- **Limit Orders**: Entry at specific price
- **Stop Orders**: Triggered when price reaches level
- **Stop-Limit Orders**: Combination of stop and limit
- **SL/TP**: Optional stop-loss and take-profit on all orders
- **Quick Volume Selection**: Predefined lot sizes (0.01, 0.1, 0.5, 1, 5)

#### 3. Position Management
- **Real-Time P&L**: Live updates based on current market prices
- **Position List**: Sortable table with all open positions
- **Modify SL/TP**: Edit stop-loss and take-profit on open positions
- **Close Individual**: Close single positions
- **Close All**: Bulk close all positions
- **Commission & Swap Display**: Full cost breakdown

#### 4. Account Dashboard
- **Balance & Equity**: Real-time account metrics
- **Margin Monitoring**: Used margin, free margin, margin level
- **Health Indicators**: Visual warnings for margin levels
- **Utilization Bar**: Graphical representation of margin usage
- **Risk Warnings**: Alerts when margin exceeds 80%

#### 5. Trading Charts
- **Candlestick Charts**: OHLC data visualization
- **Line Charts**: Simplified price tracking
- **Area Charts**: Filled price visualization
- **Multiple Timeframes**: 1m, 5m, 15m, 1h, 4h, 1d
- **Real-Time Updates**: Live price feeds from WebSocket
- **Indicators**: Built-in technical indicators

### State Management (Zustand)

The application uses Zustand for centralized state management:

```typescript
// Global state accessible from any component
const {
  ticks,           // Real-time market data
  positions,       // Open positions
  orders,          // Pending orders
  account,         // Account info
  selectedSymbol,  // Current trading pair
  // ... and more
} = useAppStore();
```

### WebSocket Integration

```typescript
import { getWebSocketService } from './services/websocket';

// Initialize WebSocket
const ws = getWebSocketService('ws://localhost:7999/ws');

// Connect
ws.connect();

// Subscribe to market data
const unsubscribe = ws.subscribe('EURUSD', (tick) => {
  console.log('New tick:', tick);
});

// Listen to connection state
ws.onStateChange((state) => {
  console.log('Connection state:', state);
});

// Cleanup
ws.disconnect();
```

### API Endpoints

#### Market Data
- `GET /api/symbols` - List available symbols
- `GET /ticks?symbol=EURUSD&limit=500` - Historical ticks
- `GET /ohlc?symbol=EURUSD&timeframe=1m&limit=500` - OHLC data

#### Trading
- `POST /api/orders/market` - Place market order
- `POST /order/limit` - Place limit order
- `POST /order/stop` - Place stop order
- `POST /order/stop-limit` - Place stop-limit order
- `GET /orders/pending` - Get pending orders
- `POST /order/cancel` - Cancel pending order

#### Positions
- `GET /api/positions?accountId={id}` - Get open positions
- `POST /api/positions/close` - Close position
- `POST /api/positions/close-bulk` - Close all positions
- `POST /api/positions/modify` - Modify SL/TP

#### Account
- `GET /api/account/summary?accountId={id}` - Account metrics
- `GET /api/trades?accountId={id}` - Trade history
- `GET /api/ledger?accountId={id}` - Transaction history

## Admin Control Panel

### Features

#### 1. Account Management
- **Account List**: View all client accounts
- **Account Details**: Balance, equity, margin, positions
- **Deposit/Withdraw**: Add or remove funds
- **Manual Adjustments**: Balance corrections
- **Bonus Management**: Add promotional bonuses
- **Password Reset**: Admin password management

#### 2. Execution Model Control
- **A-Book Mode**: Route orders to liquidity providers
- **B-Book Mode**: Internal execution against broker
- **Per-Client Toggle**: Different execution models per account
- **Real-Time Switching**: No restart required

#### 3. Liquidity Provider Management
- **LP List**: View all configured LPs (OANDA, Binance, etc.)
- **Add LP**: Configure new liquidity providers
- **Toggle LP**: Enable/disable individual LPs
- **Symbol Mapping**: Configure which symbols each LP provides
- **Connection Status**: Real-time LP connectivity monitoring
- **Quote Aggregation**: Multi-LP price feed settings

#### 4. Symbol Management
- **Symbol List**: All available trading symbols
- **Enable/Disable**: Control which symbols clients can trade
- **Spread Configuration**: Adjust bid-ask spreads
- **Commission Settings**: Set per-symbol commissions

#### 5. Risk Monitoring
- **Exposure Dashboard**: Total exposure by symbol
- **Position Heatmap**: Visualize client positions
- **Margin Alerts**: Accounts approaching margin calls
- **P&L Monitoring**: Real-time profit/loss tracking
- **Risk Limits**: Set max position sizes, leverage limits

#### 6. Routing Rules
- **Smart Router**: Conditional order routing
- **Rule Builder**: Create routing rules based on:
  - Order size
  - Client tier
  - Symbol
  - Time of day
  - Market volatility

#### 7. Transaction Ledger
- **Full History**: All account transactions
- **Filters**: By account, type, date range
- **Export**: CSV/PDF reports
- **Audit Trail**: Complete transaction log

#### 8. Settings
- **Broker Name**: White-label configuration
- **Default Leverage**: System-wide leverage setting
- **Margin Mode**: Hedging vs Netting
- **Default Balance**: Demo account starting balance
- **Max Ticks**: Historical data retention
- **Backend Restart**: Graceful server restart

### Admin API Endpoints

#### Account Administration
- `GET /admin/accounts` - List all accounts
- `POST /admin/deposit` - Add funds
- `POST /admin/withdraw` - Remove funds
- `POST /admin/adjust` - Manual balance adjustment
- `POST /admin/bonus` - Add bonus
- `POST /admin/reset-password` - Reset account password
- `POST /admin/account/update` - Update account details

#### System Configuration
- `GET /api/config` - Get broker configuration
- `POST /api/config` - Update broker configuration
- `GET /admin/execution-mode` - Get execution mode
- `POST /admin/execution-mode` - Set A-Book/B-Book mode

#### LP Management
- `GET /admin/lps` - List all LPs
- `POST /admin/lps` - Add new LP
- `PUT /admin/lps/{id}` - Update LP config
- `DELETE /admin/lps/{id}` - Remove LP
- `POST /admin/lps/{id}/toggle` - Enable/disable LP
- `GET /admin/lps/{id}/symbols` - Get LP symbols
- `GET /admin/lp-status` - LP connection status

#### Symbol Management
- `GET /admin/symbols` - List all symbols
- `POST /admin/symbols/toggle` - Enable/disable symbol

#### Monitoring
- `GET /admin/ledger` - Full transaction ledger
- `GET /admin/routes` - Routing rules
- `GET /admin/fix/status` - FIX session status
- `POST /admin/fix/connect` - Connect FIX session
- `POST /admin/fix/disconnect` - Disconnect FIX

#### System Control
- `POST /admin/restart` - Graceful backend restart

## Deployment

### Client Trading Interface

```bash
cd clients/desktop

# Install dependencies
npm install

# Development
npm run dev
# Opens on http://localhost:5173

# Production build
npm run build
# Output: dist/

# Preview production build
npm run preview
```

### Admin Panel

```bash
cd admin/broker-admin

# Install dependencies
npm install

# Development
npm run dev
# Opens on http://localhost:3000

# Production build
npm run build

# Start production server
npm run start
```

## Configuration

### Environment Variables

#### Client (.env)
```env
VITE_API_BASE_URL=http://localhost:7999
VITE_WS_URL=ws://localhost:7999/ws
```

#### Admin (.env.local)
```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:7999
```

### Backend CORS

Ensure the backend allows requests from frontend origins:

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
```

## Responsive Design

Both interfaces are fully responsive:

### Desktop (1920x1080+)
- Full dashboard layout
- Multi-panel trading view
- Extended data tables
- Chart maximization mode

### Tablet (768x1024)
- Stacked layout
- Collapsible panels
- Optimized table columns

### Mobile (375x667+)
- Single-column layout
- Bottom sheet order entry
- Swipeable position list
- Simplified charts

## Security Considerations

### Client-Side
1. **No API Keys in Frontend**: All sensitive data handled by backend
2. **JWT Authentication**: Secure token-based auth
3. **HTTPS Only**: Production must use SSL/TLS
4. **WebSocket Auth**: Token validation on WS connections
5. **Input Validation**: All user inputs sanitized
6. **CSP Headers**: Content Security Policy enforcement

### Admin Panel
1. **Role-Based Access**: Admin-only endpoints
2. **Audit Logging**: All actions logged
3. **IP Whitelisting**: Restrict admin access by IP
4. **Two-Factor Auth**: Optional 2FA for admin logins
5. **Session Timeout**: Auto-logout after inactivity

## Performance Optimization

### Client
1. **Tick Throttling**: 100ms update interval prevents UI lag
2. **Virtual Scrolling**: Large position lists use virtualization
3. **Memoization**: React.memo for expensive components
4. **Code Splitting**: Lazy load chart libraries
5. **Asset Optimization**: Minified bundles, tree-shaking

### WebSocket
1. **Buffered Broadcasts**: Non-blocking sends
2. **Subscription Filtering**: Only requested symbols sent
3. **Compression**: Enable WebSocket compression
4. **Reconnection Logic**: Exponential backoff prevents server overload

## Testing

### Unit Tests
```bash
npm run test
```

### E2E Tests (Playwright)
```bash
npm run test:e2e
```

### Manual Testing Checklist

#### Client
- [ ] Login with demo account
- [ ] WebSocket connects and receives ticks
- [ ] Chart displays and updates in real-time
- [ ] Place market order (buy/sell)
- [ ] Place limit order
- [ ] Modify position SL/TP
- [ ] Close individual position
- [ ] Close all positions
- [ ] Account metrics update correctly
- [ ] Responsive layout on mobile

#### Admin
- [ ] View all accounts
- [ ] Deposit funds to account
- [ ] Withdraw funds
- [ ] Toggle execution mode (A-Book/B-Book)
- [ ] Enable/disable symbol
- [ ] View transaction ledger
- [ ] Add/remove LP
- [ ] Monitor LP connection status
- [ ] Restart backend

## Troubleshooting

### WebSocket Not Connecting
1. Check backend is running on port 7999
2. Verify CORS headers on backend
3. Check browser console for errors
4. Test with `wscat -c ws://localhost:7999/ws`

### Orders Not Executing
1. Verify account has sufficient balance
2. Check execution mode (A-Book requires LP connection)
3. Review backend logs for errors
4. Ensure symbol is enabled

### Admin Panel Empty
1. Check API endpoint: `http://localhost:7999/admin/accounts`
2. Verify backend is running
3. Check browser network tab for failed requests
4. Review CORS configuration

## Future Enhancements

### Planned Features
1. **TradingView Integration**: Advanced charting
2. **Mobile Apps**: React Native iOS/Android
3. **Social Trading**: Copy trading functionality
4. **Automated Strategies**: Algo trading support
5. **Advanced Orders**: OCO, trailing stops, iceberg
6. **Multi-Account**: Trade multiple accounts simultaneously
7. **Dark/Light Themes**: User preferences
8. **Notifications**: Desktop/push notifications
9. **Export Tools**: Download trade history, statements
10. **Analytics Dashboard**: Performance metrics, statistics

## Support

For issues or questions:
- GitHub Issues: [trading-engine/issues](https://github.com/epic1st/trading-engine/issues)
- Documentation: This file
- Backend Logs: `backend/logs/`
- Frontend Console: Browser DevTools

## License

Proprietary - All rights reserved
