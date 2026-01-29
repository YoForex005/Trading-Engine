# Trading Engine Admin Dashboard

A comprehensive production-grade admin dashboard for real-time monitoring and control of the trading engine.

## Features

### Real-time Monitoring Views
- **System Overview**: Live CPU, memory, connections, and throughput metrics with historical trends
- **Order Flow Monitor**: Real-time order tracking with filtering, sorting, and export
- **LP Health Monitor**: Liquidity provider status, latency, uptime, and fill rates
- **User Activity Monitor**: Active sessions, user orders, and P&L tracking
- **Error Dashboard**: Error logs with severity filtering and real-time alerts

### Control Panels
- **Routing Control**: Configure A-Book/B-Book/C-Book routing rules
- **LP Management**: Manage liquidity provider connections and settings
- **Symbol Management**: Configure trading symbols, margins, and trading hours
- **User Management**: User accounts, roles, and permissions
- **Risk Control**: Risk limits, position controls, and automated actions

### Analytics Views
- **Performance Analytics**: System performance metrics and latency analysis
- **Trading Analytics**: Volume, P&L, fill rates, and top symbols
- **User Analytics**: User profitability, win rates, and trading patterns
- **Audit Trail**: Complete audit log of system actions

### Additional Features
- Real-time WebSocket updates with auto-reconnection
- Dark mode support
- Mobile responsive design
- Export to CSV/PDF
- Role-based access control
- Comprehensive error handling and loading states
- Interactive charts (Recharts)
- Filtering and sorting on all tables
- Session persistence

## Tech Stack

- **React 18** with TypeScript
- **Vite** for fast builds
- **Tailwind CSS** for styling
- **Recharts** for data visualization
- **React Router** for navigation
- **date-fns** for date formatting
- **Lucide React** for icons
- **WebSocket** for real-time updates

## Prerequisites

- Node.js 20+
- npm 9+ or bun

## Installation

```bash
# Install dependencies
npm install
# or
bun install
```

## Development

```bash
# Start development server (http://localhost:3001)
npm run dev
# or
bun run dev
```

The development server includes:
- Hot module replacement (HMR)
- Proxy to backend API (http://localhost:8080)
- WebSocket proxy for real-time updates

## Build

```bash
# Type check
npm run typecheck

# Build for production
npm run build

# Preview production build
npm run preview
```

## Project Structure

```
src/
├── components/
│   ├── monitoring/          # Real-time monitoring views
│   │   ├── SystemOverview.tsx
│   │   ├── OrderFlowMonitor.tsx
│   │   ├── LPHealthMonitor.tsx
│   │   ├── UserActivityMonitor.tsx
│   │   └── ErrorDashboard.tsx
│   ├── controls/            # Control panels
│   │   ├── RoutingControl.tsx
│   │   ├── LPManagement.tsx
│   │   ├── SymbolManagement.tsx
│   │   ├── UserManagement.tsx
│   │   └── RiskControl.tsx
│   ├── analytics/           # Analytics views
│   │   ├── PerformanceAnalytics.tsx
│   │   ├── TradingAnalytics.tsx
│   │   ├── UserAnalytics.tsx
│   │   └── AuditTrail.tsx
│   ├── shared/              # Reusable UI components
│   │   ├── Card.tsx
│   │   ├── Badge.tsx
│   │   ├── Button.tsx
│   │   └── Table.tsx
│   ├── Layout.tsx           # Main layout with sidebar
│   └── Login.tsx            # Authentication
├── services/
│   ├── api.ts               # REST API client
│   └── websocket.ts         # WebSocket service
├── types/
│   └── index.ts             # TypeScript type definitions
├── styles/
│   └── index.css            # Global styles
├── App.tsx                  # Main app component
└── main.tsx                 # Entry point
```

## API Integration

The dashboard expects the following backend API endpoints:

### Authentication
- `POST /api/auth/login` - Login with username/password
- `POST /api/auth/logout` - Logout

### Monitoring
- `GET /api/metrics/system` - System metrics
- `GET /api/orders` - Order list with filters
- `GET /api/lps` - Liquidity providers
- `GET /api/sessions` - Active user sessions
- `GET /api/errors` - Error logs

### Controls
- `GET /api/routing/rules` - Routing rules
- `POST /api/routing/rules` - Create routing rule
- `PUT /api/routing/rules/:id` - Update routing rule
- `DELETE /api/routing/rules/:id` - Delete routing rule
- `GET /api/symbols` - Trading symbols
- `GET /api/users` - User list
- `GET /api/risk/limits` - Risk limits

### Analytics
- `GET /api/analytics/trading` - Trading analytics
- `GET /api/analytics/users` - User analytics
- `GET /api/analytics/performance` - Performance metrics
- `GET /api/audit` - Audit logs

### Export
- `GET /api/export` - Export data to CSV/PDF

### WebSocket Messages

The dashboard listens for these WebSocket message types:

```typescript
{
  type: 'system_metrics' | 'order_update' | 'lp_status' | 'user_activity' | 'error_log' | 'alert',
  timestamp: number,
  data: T
}
```

## Configuration

Edit `vite.config.ts` to change the backend API and WebSocket URLs:

```typescript
server: {
  port: 3001,
  proxy: {
    '/api': {
      target: 'http://localhost:8080',  // Change this
      changeOrigin: true,
    },
    '/ws': {
      target: 'ws://localhost:8080',    // Change this
      ws: true,
    },
  },
}
```

## Authentication

The dashboard uses JWT tokens stored in localStorage:
- Token is sent in `Authorization: Bearer <token>` header
- Login returns `{ token, user }` object
- Token is cleared on logout

## Dark Mode

Dark mode is automatically applied based on user preference and persisted in localStorage. Toggle with the moon/sun icon in the header.

## Development Tips

1. **Hot Reload**: Changes to React components trigger instant updates
2. **Type Safety**: TypeScript catches errors at compile time
3. **API Mocking**: Use the browser's network tab to intercept API calls for testing
4. **WebSocket Testing**: Use tools like wscat to test WebSocket connections

## Production Deployment

1. Build the production bundle:
   ```bash
   npm run build
   ```

2. Serve the `dist` folder with a static file server or integrate with your backend

3. Ensure backend API and WebSocket endpoints are configured correctly

4. Set up proper authentication and authorization

5. Configure CORS if API is on a different domain

## Browser Support

- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)

## Performance

- Code splitting for optimal load times
- Lazy loading of routes
- Optimized re-renders with React hooks
- Debounced filters and search
- Virtualized lists for large datasets (can be added if needed)

## Security

- JWT authentication
- Role-based access control (RBAC)
- Input validation
- XSS protection via React
- CSRF protection (implement on backend)
- Secure WebSocket connections (wss://)

## License

Proprietary - All rights reserved
