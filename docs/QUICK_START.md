# Quick Start Guide

## Prerequisites

- Node.js 20+ and npm 9+
- Go 1.21+ (for backend)
- Git

## Installation

### 1. Backend Setup

```bash
cd backend

# Install Go dependencies
go mod download

# Build the server
go build -o server cmd/server/main.go

# Run the backend
./server
# Backend starts on http://localhost:7999
```

### 2. Client Trading Interface Setup

```bash
cd clients/desktop

# Install dependencies
npm install

# Add Zustand if not already installed
npm install zustand@^5.0.2

# Start development server
npm run dev
# Client opens on http://localhost:5173
```

### 3. Admin Panel Setup

```bash
cd admin/broker-admin

# Install dependencies
npm install

# Start development server
npm run dev
# Admin opens on http://localhost:3000
```

## Quick Test

### 1. Open Client Interface
Navigate to http://localhost:5173

### 2. Login
- Account ID: `1` (demo account)
- Username: `demo-user`
- Password: `password`

### 3. Test Trading
1. Select a symbol (e.g., EURUSD)
2. Wait for WebSocket connection (green dot)
3. Set volume to 0.01
4. Click BUY or SELL
5. View your position in the Positions tab
6. Close the position

### 4. Test Admin Panel
Navigate to http://localhost:3000

#### View Accounts
- Go to "Accounts" tab
- See demo account with $5000 balance

#### Test Deposit
1. Select demo account
2. Click "Deposit"
3. Enter $1000
4. Method: "Bank Transfer"
5. Submit

#### Toggle Execution Mode
1. Go to "Settings" tab
2. Switch between B-Book (internal) and A-Book (LP routing)
3. Note: A-Book requires active LP connection

#### Monitor Real-Time Data
1. Go to "LP Status" tab
2. View connected liquidity providers
3. Check quote aggregation status
4. Monitor tick rates

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                   Frontend Layer                     │
│                                                      │
│  ┌────────────────┐         ┌──────────────────┐   │
│  │ Client Trading │         │  Admin Control   │   │
│  │   Interface    │         │      Panel       │   │
│  │ (Vite/React)   │         │  (Next.js)       │   │
│  │  Port: 5173    │         │  Port: 3000      │   │
│  └────────┬───────┘         └────────┬─────────┘   │
│           │                          │             │
│           └──────────┬───────────────┘             │
└──────────────────────┼─────────────────────────────┘
                       │ HTTP/REST + WebSocket
┌──────────────────────┼─────────────────────────────┐
│                      ▼                              │
│              Backend Server (Go)                    │
│                Port: 7999                           │
│                                                     │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │  B-Book     │  │  Smart       │  │  LP       │ │
│  │  Engine     │  │  Router      │  │  Manager  │ │
│  └─────────────┘  └──────────────┘  └───────────┘ │
│                                                     │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │ WebSocket   │  │  Tick Store  │  │  FIX      │ │
│  │ Hub         │  │  (OHLC)      │  │  Gateway  │ │
│  └─────────────┘  └──────────────┘  └───────────┘ │
└─────────────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────┐
│           Liquidity Providers (Optional)            │
│                                                     │
│  OANDA │ Binance │ FIX Brokers │ Custom LPs         │
└─────────────────────────────────────────────────────┘
```

## Default Configuration

### Backend (`backend/cmd/server/main.go`)
- **Port**: 7999
- **Execution Mode**: B-Book (internal)
- **Default LP**: OANDA (for price feed)
- **Demo Balance**: $5000
- **Leverage**: 1:100
- **Margin Mode**: Hedging

### Client
- **API URL**: http://localhost:7999
- **WebSocket**: ws://localhost:7999/ws
- **Default Symbol**: EURUSD
- **Chart Timeframe**: 1 minute
- **Auto-Reconnect**: Yes (exponential backoff)

### Admin
- **API URL**: http://localhost:7999
- **Refresh Interval**: 5 seconds
- **Default View**: Accounts

## Common Operations

### Add Funds to Demo Account
```bash
curl -X POST http://localhost:7999/admin/deposit \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "1",
    "amount": 1000,
    "method": "Bank Transfer",
    "reference": "TEST-001"
  }'
```

### Place Market Order (API)
```bash
curl -X POST http://localhost:7999/api/orders/market \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "1",
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.01
  }'
```

### Get Account Summary
```bash
curl http://localhost:7999/api/account/summary?accountId=1
```

### Get Open Positions
```bash
curl http://localhost:7999/api/positions?accountId=1
```

## Troubleshooting

### Backend won't start
```bash
# Check if port 7999 is already in use
lsof -i :7999

# Kill existing process
kill -9 <PID>

# Or use a different port
PORT=8080 ./server
```

### Frontend won't connect
1. Check backend is running: `curl http://localhost:7999/health`
2. Check WebSocket: Install wscat and run:
   ```bash
   npm install -g wscat
   wscat -c ws://localhost:7999/ws
   ```
3. Check browser console for errors

### No market data
1. Verify backend logs show "Pipeline check: X ticks received"
2. Check LP connection status in admin panel
3. Ensure at least one LP is enabled
4. Restart backend: `POST /admin/restart`

### Orders not executing
1. Check account balance: `GET /api/account/summary?accountId=1`
2. Verify symbol is enabled: `GET /admin/symbols`
3. Check execution mode: `GET /admin/execution-mode`
4. Review backend logs for errors

## Development Workflow

### Hot Reload Development

#### Terminal 1: Backend
```bash
cd backend

# Install air for hot reload
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air
```

#### Terminal 2: Client
```bash
cd clients/desktop
npm run dev
# Auto-reloads on file changes
```

#### Terminal 3: Admin
```bash
cd admin/broker-admin
npm run dev
# Auto-reloads on file changes
```

### Making Changes

#### Add New Trading Symbol
1. Edit `backend/data/lp_config.json`
2. Add symbol to LP configuration
3. Restart backend
4. Symbol appears in client symbol list

#### Modify Spreads
1. Admin Panel → Settings
2. Adjust spread multiplier
3. Changes apply immediately

#### Add New Order Type
1. Backend: Add handler in `backend/api/server.go`
2. Client: Add UI in `components/OrderEntryPanel.tsx`
3. Test end-to-end

## Production Deployment

### Backend
```bash
cd backend
go build -o rtx-server cmd/server/main.go

# Run with systemd or supervisor
./rtx-server
```

### Client (Static Build)
```bash
cd clients/desktop
npm run build

# Serve with nginx, apache, or CDN
# Output: dist/
```

### Admin (Next.js Production)
```bash
cd admin/broker-admin
npm run build
npm run start

# Or deploy to Vercel/Netlify
```

## Next Steps

1. **Configure Real LPs**: Add your OANDA/Binance API keys
2. **Enable HTTPS**: Use nginx with SSL certificates
3. **Set Up Authentication**: Implement JWT with proper user management
4. **Add Monitoring**: Set up logging and alerting
5. **Database Integration**: Replace in-memory storage with PostgreSQL
6. **Backup Strategy**: Implement regular backups
7. **Load Testing**: Test with multiple concurrent users
8. **Security Audit**: Review authentication, authorization, input validation

## Resources

- **Full Documentation**: `/docs/FRONTEND_IMPLEMENTATION.md`
- **Backend Architecture**: `/backend/README.md`
- **API Reference**: `http://localhost:7999/docs` (Swagger)
- **Component Storybook**: Coming soon

## Support

- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: `/docs/` folder
- **Logs**: Check `backend/logs/` and browser console

Happy Trading! 🚀
