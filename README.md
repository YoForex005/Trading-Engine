# RTX Trading Engine

RTX is a high-performance, multi-asset trading platform designed for microsecond-latency execution, A-Book/B-Book hybrid routing, and robust multi-terminal support (Desktop, Web, Mobile).

## Project Structure

```
rtx-trading-engine/
├── backend/                  # Go Backend Services
│   ├── main.go              # Entry point
│   ├── api/                 # HTTP/REST handlers
│   ├── auth/                # Authentication service
│   ├── oms/                 # Order Management System
│   ├── risk/                # Risk Engine (margin, equity)
│   ├── router/              # Smart Order Router (A/B-Book)
│   ├── fix/                 # FIX Gateway (LP connectivity)
│   ├── ws/                  # WebSocket Hub & Market Data
│   └── database/            # SQL schemas
├── clients/
│   └── desktop/             # React/TypeScript Desktop Terminal
├── admin/
│   ├── broker-admin/        # Next.js Broker Admin Dashboard
│   └── super-admin/         # Next.js Super Admin Platform Control
└── README.md
```

## Prerequisites

- **Node.js** v18+ (for frontends)
- **Go** v1.22+ (for backend) - `brew install go`
- **npm** (comes with Node.js)

## Quick Start

### 1. Start the Backend

```bash
cd backend
go mod tidy  # Download dependencies (first time only)
go run cmd/server/main.go
```

The backend will start on `http://localhost:8080` with:
- REST API endpoints
- WebSocket market data on `ws://localhost:8080/ws`

### 2. Start the Desktop Terminal

```bash
cd clients/desktop
npm install  # First time only
npm run dev
```

Open `http://localhost:5173` in your browser.

**Login credentials:**
- Username: `admin` / Password: `password`
- Username: `trader` / Password: `password`

### 3. Start the Broker Admin Dashboard

```bash
cd admin/broker-admin
npm install
npm run dev
```

Open `http://localhost:3000` (or the port shown in terminal).

### 4. Start the Super Admin Dashboard

```bash
cd admin/super-admin
npm install
npm run dev
```

Open `http://localhost:3001` (adjust port if needed).

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/login` | POST | Authenticate user |
| `/health` | GET | Health check |
| `/order` | POST | Place a new order |
| `/account` | GET | Get account info |
| `/admin/routes` | GET | Get routing rules |
| `/admin/lp-status` | GET | Get LP connection status |
| `/ws` | WS | Real-time market data stream |

## Architecture Highlights

- **Desktop Terminal**: React + TypeScript with WebSocket for real-time ticks
- **Backend**: Go with modular services (OMS, Risk, Router, FIX Gateway)
- **Smart Routing**: Configurable A-Book/B-Book rules based on volume and user group
- **FIX Gateway**: Simulated LMAX connection (ready for real implementation)
- **Admin Dashboards**: Next.js with Tailwind CSS

## Development Roadmap

- [x] Phase 0: Architecture & Design
- [x] Phase 1: Foundation (Desktop Shell + Backend Auth)
- [x] Phase 2: Core Trading (OMS, Risk Engine)
- [x] Phase 3: Liquidity (Smart Router, FIX Gateway)
- [x] Phase 4: Admin Terminals (Broker + Super Admin)
- [ ] Phase 5: Advanced Features (Algo Trading, Mobile App)

## License

Proprietary - All Rights Reserved

## Whole Project run
1.Backend: cd backend && go run cmd/server/main.go
2.Desktop: cd clients/desktop && npm run dev
3.Broker Admin: cd admin/broker-admin && npm run dev
4.Super Admin: cd admin/super-admin && npm run dev

## Market Data Features

**As of January 30, 2026**: Trading Engine uses **100% real YOFX market data** with the following features:

- ✅ **Real-time Ticks**: Direct FIX gateway feed from YOFX broker
- ✅ **24-Hour High/Low Tracking**: Enhanced market data with daily ranges
- ✅ **Zero Simulation**: Removed all mock data fallbacks
- ✅ **Secure Configuration**: All credentials externalized via environment variables
- ✅ **WebSocket API**: Real-time price streaming to connected clients
- ✅ **Production Ready**: Fully configured for multi-environment deployment

For configuration details, see [docs/MOCK_DATA_REMOVAL.md](docs/MOCK_DATA_REMOVAL.md)
For WebSocket API docs, see [docs/WEBSOCKET_MARKET_DATA_API.md](docs/WEBSOCKET_MARKET_DATA_API.md)