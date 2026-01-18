# Trading Platform Frontend Architecture

## Overview

This document describes the comprehensive frontend architecture for the RTX Trading Platform, including both the client trading interface and admin control panel.

## Technology Stack

### Client Trading Interface (`/clients/desktop/`)
- **Framework**: React 19.2 + TypeScript
- **Build Tool**: Vite 7.2
- **State Management**: Zustand 5.0 (with persist and devtools)
- **Charts**: TradingView Lightweight Charts 5.1
- **Styling**: Tailwind CSS 4.1
- **Real-time**: WebSocket with auto-reconnection
- **Icons**: Lucide React

### Admin Control Panel (`/admin/broker-admin/`)
- **Framework**: Next.js 16.1 + TypeScript
- **Styling**: Tailwind CSS 4.1
- **Icons**: Lucide React
- **API Integration**: REST + Real-time WebSocket

## Project Structure

```
clients/desktop/src/
├── components/          # React components
│   ├── Login.tsx
│   ├── TradingChart.tsx
│   ├── BottomDock.tsx
│   ├── OrderEntry.tsx       # NEW: Advanced order entry
│   ├── NotificationToast.tsx # NEW: Toast notifications
│   └── ErrorBoundary.tsx
├── services/           # API & Services
│   ├── api.ts         # NEW: Comprehensive API layer
│   ├── websocket.ts   # WebSocket service
│   └── notifications.ts # NEW: Notification system
├── store/             # State management
│   └── useAppStore.ts # Zustand store
├── hooks/             # Custom hooks
│   └── useKeyboardShortcuts.ts # NEW: Keyboard shortcuts
├── App.tsx            # Main application
└── main.tsx           # Entry point

admin/broker-admin/src/
├── app/
│   ├── layout.tsx
│   └── page.tsx
├── components/
│   ├── ui/           # Reusable UI components
│   └── dashboard/    # Admin dashboard views
│       ├── AccountsView.tsx
│       ├── LPStatusView.tsx
│       ├── SymbolsView.tsx
│       ├── RiskView.tsx
│       └── SettingsView.tsx
└── types/
    └── index.ts
```

## Core Features Implemented

### 1. Real-Time Market Data
- **WebSocket Connection** with auto-reconnection and exponential backoff
- **Tick Buffering** (100ms throttle) to prevent UI lag
- **Connection State Management** (connecting, connected, disconnected, error)
- **Heartbeat/Ping** mechanism (30s interval)
- **Graceful Reconnection** with max retry limits

### 2. Trading Interface Components

#### Market Watch Panel
- Real-time quote grid (bid, ask, spread)
- Symbol search and filtering
- Color-coded price changes (green/red)
- One-click symbol selection
- Displays 100+ instruments

#### Order Entry Panel (NEW)
- **Order Types**: Market, Limit, Stop, Stop-Limit
- **Risk Calculator**: Position sizing based on risk %
- **Margin Preview**: Real-time margin requirements
- **SL/TP Configuration**: Optional stop-loss and take-profit
- **Pre-trade Validation**: Balance, margin, lot size checks
- **Keyboard Shortcuts**: F9 for new order, Ctrl+B for buy, Ctrl+S for sell

#### Chart Component
- TradingView Lightweight Charts integration
- Multiple timeframes (M1, M5, M15, H1, H4, D1)
- Chart types (candlestick, line, area)
- Position overlays on chart
- Fullscreen mode (F11)

#### Position Management
- Active positions table with real-time P&L updates
- Color-coded profit/loss
- Modify position (SL/TP)
- Close position (full or partial)
- Bulk close actions (all, winners, losers)
- Swap and commission display

### 3. State Management (Zustand)

```typescript
interface AppState {
  // Authentication
  isAuthenticated: boolean;
  accountId: string | null;

  // Market Data
  ticks: Record<string, Tick>;
  selectedSymbol: string;

  // Trading
  positions: Position[];
  orders: Order[];
  trades: Trade[];

  // Account
  account: Account | null;

  // UI State
  isChartMaximized: boolean;
  chartType: 'candlestick' | 'line' | 'area';
  timeframe: '1m' | '5m' | '15m' | '1h' | '4h' | '1d';
  orderVolume: number;

  // WebSocket
  wsConnected: boolean;

  // Actions
  setAuthenticated: (isAuth: boolean, accountId?: string) => void;
  setTick: (symbol: string, tick: Tick) => void;
  setPositions: (positions: Position[]) => void;
  // ... more actions
}
```

**Persistence**: Selected symbol, chart type, timeframe, and order volume are persisted to localStorage.

### 4. API Service Layer (NEW)

Comprehensive TypeScript API client with:

- **Authentication API**: Login/logout
- **Account API**: Account summary, balance, equity
- **Positions API**: Get, close, modify, bulk close
- **Orders API**: Market, limit, stop, stop-limit orders
- **Market Data API**: Symbols, ticks, OHLC data
- **Risk API**: Margin preview, lot size calculator
- **History API**: Trades, ledger entries
- **Admin API**: User management, deposits, withdrawals, configuration

**Error Handling**:
- Request timeout (10s default)
- Automatic retry with exponential backoff
- Centralized error messages
- Type-safe responses

**Example Usage**:
```typescript
import { api } from './services/api';

// Place market order
const position = await api.orders.placeMarketOrder({
  accountId: 1,
  symbol: 'EURUSD',
  side: 'BUY',
  volume: 0.1,
  sl: 1.08500,
  tp: 1.09500,
});

// Get margin preview
const marginPreview = await api.risk.previewMargin('EURUSD', 0.1, 'BUY');
```

### 5. Notification System (NEW)

Toast notification service with:
- **Types**: Success, error, warning, info
- **Auto-dismiss**: Configurable duration (default 5s)
- **Manual dismiss**: Close button
- **Queue Management**: Max 5 visible notifications
- **Animations**: Slide-in transitions

**Usage**:
```typescript
import { notificationService } from './services/notifications';

notificationService.success('Order Executed', 'BUY 0.1 EURUSD @ 1.08765');
notificationService.error('Order Failed', 'Insufficient margin');
```

### 6. Keyboard Shortcuts (NEW)

Power-user shortcuts:
- **F9**: New order dialog
- **ESC**: Close modals
- **F11**: Toggle chart fullscreen
- **Arrow Up/Down**: Navigate symbols
- **Ctrl + B**: Buy market order
- **Ctrl + S**: Sell market order
- **Ctrl + Plus/Minus**: Adjust volume
- **Ctrl + Shift + W**: Close all positions
- **Ctrl + Shift + S**: Save layout
- **Ctrl + Shift + L**: Load layout

Shortcuts are disabled when typing in input fields (except ESC).

### 7. Admin Control Panel

#### Dashboard
- Total clients, active trades, total volume
- Real-time P&L charts
- System health indicators
- Execution mode (A-Book vs B-Book)

#### Client Management
- Client list with search/filter
- Account details (balance, equity, positions)
- Deposit/Withdraw/Adjust balance
- Set leverage
- Suspend/Activate accounts
- Reset passwords

#### Liquidity Provider Management
- LP connection status
- Enable/disable LPs
- LP performance metrics (latency, quotes)
- Symbol routing configuration

#### Symbol Management
- Enable/disable trading pairs
- Configure spreads
- Set position limits
- Commission settings

#### Risk Controls
- Global and per-client leverage limits
- Margin requirements
- Stop-out levels
- Maximum position sizes
- Daily loss limits

## API Integration

### Backend Endpoints Used

**B-Book API** (RTX Internal):
```
GET  /api/account/summary?accountId={id}
GET  /api/positions?accountId={id}
POST /api/orders/market
POST /api/positions/close
POST /api/positions/modify
POST /api/positions/close-bulk
GET  /api/trades?accountId={id}
GET  /api/ledger?accountId={id}
GET  /api/symbols
```

**Admin API**:
```
GET  /admin/accounts
POST /admin/deposit
POST /admin/withdraw
POST /admin/adjust
POST /admin/bonus
GET  /admin/ledger
POST /admin/account/update
GET  /admin/symbols
POST /admin/symbols/toggle
GET  /admin/execution-mode
POST /admin/execution-mode
GET  /admin/lp-status
POST /admin/restart
```

**Market Data**:
```
GET  /ticks?symbol={symbol}&limit={limit}
GET  /ohlc?symbol={symbol}&timeframe={tf}&limit={limit}
```

**Risk Management**:
```
GET  /risk/calculate-lot?symbol={symbol}&riskPercent={pct}&slPips={pips}
GET  /risk/margin-preview?symbol={symbol}&volume={vol}&side={side}
```

**WebSocket**:
```
WS   ws://localhost:8080/ws
```

WebSocket message types:
- `tick`: Market price updates
- `position`: Position updates
- `account`: Account balance updates
- `ping/pong`: Keep-alive

## Performance Optimizations

### 1. Tick Throttling
Market data updates are buffered and flushed at 10 FPS (100ms) to prevent excessive re-renders.

### 2. Virtual Scrolling
Market watch and position tables use virtual scrolling for large datasets.

### 3. Memoization
React components use `React.memo()`, `useMemo()`, and `useCallback()` to prevent unnecessary re-renders.

### 4. Code Splitting
Routes and heavy components are lazy-loaded with `React.lazy()` and `Suspense`.

### 5. WebSocket Batching
Multiple tick updates are batched together before state updates.

## Security Features

### 1. Authentication
- JWT token-based authentication
- Tokens stored in memory (not localStorage for XSS protection)
- Auto-logout on token expiration

### 2. API Security
- CORS configuration
- Request timeout enforcement
- Input validation on all forms
- XSS protection via React's built-in escaping

### 3. WebSocket Security
- Secure WebSocket (wss://) in production
- Connection authentication
- Message validation

## Error Handling

### 1. API Errors
- Network errors with retry logic
- HTTP error codes with user-friendly messages
- Timeout handling

### 2. WebSocket Errors
- Auto-reconnection with exponential backoff
- Max reconnection attempts (10)
- Connection state display

### 3. UI Errors
- Error boundaries to catch React errors
- Toast notifications for user actions
- Form validation with inline error messages

## Testing Strategy

### Unit Tests
- Component rendering
- State management (Zustand store)
- API service methods
- Utility functions

### Integration Tests
- Order placement flow
- Position management
- WebSocket connection
- Authentication flow

### E2E Tests (Recommended)
- Complete trading workflow
- Admin panel operations
- Multi-symbol trading
- Risk management scenarios

## Deployment

### Production Build
```bash
# Client
cd clients/desktop
npm run build
# Output: dist/

# Admin
cd admin/broker-admin
npm run build
# Output: .next/
```

### Environment Variables
```bash
# .env
VITE_API_URL=https://api.rtxtrading.com
VITE_WS_URL=wss://api.rtxtrading.com/ws
VITE_ENV=production
```

### Hosting Options
- **Static Hosting**: Vercel, Netlify, Cloudflare Pages
- **CDN**: CloudFront, Fastly
- **Containerized**: Docker + Nginx

## Future Enhancements

### Planned Features
1. **Advanced Charting**: Drawing tools, indicators, multiple chart layouts
2. **Algo Trading**: Strategy builder, backtesting
3. **Social Trading**: Copy trading, leaderboards
4. **Mobile App**: React Native version
5. **Advanced Analytics**: Performance metrics, trade journal
6. **Multi-Language**: i18n support
7. **Dark/Light Themes**: User-customizable themes
8. **Custom Layouts**: Drag-and-drop workspace customization
9. **Economic Calendar**: News integration
10. **Account History Reports**: PDF/Excel exports

### Performance Improvements
1. **Service Worker**: Offline support, caching
2. **IndexedDB**: Local storage for tick history
3. **WebWorkers**: Heavy computations off main thread
4. **Streaming**: Server-sent events for updates
5. **GraphQL**: More efficient data fetching

## Development Guidelines

### Code Style
- **TypeScript**: Strict mode enabled
- **ESLint**: Airbnb config
- **Prettier**: Auto-formatting
- **Naming**: PascalCase for components, camelCase for functions

### Component Structure
```typescript
// Component file structure
import statements
type/interface definitions
helper functions
main component
styled components (if any)
export
```

### State Management Best Practices
- Use Zustand for global state
- Use local state for component-specific UI state
- Avoid prop drilling (use context for deep trees)
- Keep state minimal and derived values computed

### API Integration
- Always use the centralized API service (`/services/api.ts`)
- Never use `fetch` directly in components
- Handle errors with try/catch and notifications
- Show loading states during async operations

## Support & Maintenance

### Monitoring
- Sentry for error tracking
- LogRocket for session replay
- Google Analytics for usage metrics

### Updates
- Regular dependency updates
- Security patches
- Performance monitoring
- User feedback integration

---

**Last Updated**: 2026-01-18
**Version**: 3.0
**Maintainer**: RTX Trading Development Team
