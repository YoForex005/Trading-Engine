# Codebase Structure

**Analysis Date:** 2026-01-15

## Directory Layout

```
Trading-Engine/
├── backend/               # Go backend service
│   ├── cmd/server/       # Main entry point
│   ├── internal/         # Private application code
│   │   ├── core/        # Core trading engine
│   │   └── api/handlers/# HTTP request handlers
│   ├── lpmanager/        # Liquidity provider management
│   ├── auth/             # Authentication service
│   ├── ws/               # WebSocket hub
│   ├── tickstore/        # Tick history storage
│   ├── binance/          # Binance API client
│   ├── oanda/            # OANDA API client
│   ├── flexymarkets/     # FlexyMarkets API client
│   ├── risk/             # Risk calculations
│   ├── orders/           # Order management
│   ├── router/           # Smart router (A/B-Book)
│   └── data/             # Configuration and tick data
├── clients/
│   └── desktop/          # React desktop trading app
│       ├── src/
│       │   ├── components/  # UI components
│       │   ├── services/    # Data fetching services
│       │   ├── hooks/       # Custom React hooks
│       │   └── indicators/  # Technical indicator engine
├── admin/
│   ├── broker-admin/     # Next.js broker dashboard
│   └── super-admin/      # Next.js super admin panel
└── scripts/              # Utility scripts
```

## Directory Purposes

**backend/cmd/server/**
- Purpose: Application entry point
- Files: `main.go` (server initialization, dependency injection, route registration)
- Key file: `main.go` - 500+ lines of server setup

**backend/internal/core/**
- Purpose: Core trading engine logic
- Files: `engine.go` (B-Book engine), `ledger.go` (transactions), `pnl.go` (P/L calculations)
- Key files:
  - `engine.go` - Account and position management
  - `ledger.go` - Transaction history
  - `pnl.go` - Real-time P/L updates

**backend/internal/api/handlers/**
- Purpose: HTTP request handlers for B-Book API
- Files: `api.go`, `account.go`, `positions.go`, `orders.go`, `admin.go`, `admin_symbols.go`, `ledger.go`, `market.go`, `lp.go`
- Pattern: One handler per domain area

**backend/lpmanager/**
- Purpose: Liquidity provider orchestration
- Files:
  - `manager.go` - LP lifecycle management
  - `lp.go` - Interfaces and types
  - `registry.go` - LP adapter registry
  - `adapters/*.go` - LP implementations

**backend/ws/**
- Purpose: WebSocket communication hub
- Files: `hub.go` - Client management, broadcast channel, price cache

**clients/desktop/src/components/**
- Purpose: React UI components
- Files:
  - `TradingChart.tsx` - Main chart component (952 lines - complex)
  - `Login.tsx` - Authentication UI
  - `MarketWatch/` - Symbol list with prices
  - `BottomDock.tsx` - Positions/orders panel
  - `IndicatorManager.tsx` - Indicator configuration

**clients/desktop/src/services/**
- Purpose: Frontend data layer
- Files:
  - `DataCache.ts` - IndexedDB cache for candles
  - `ExternalDataService.ts` - Backend API calls
  - `IndicatorStorage.ts` - localStorage for indicator preferences
  - `DataSyncService.ts` - Data synchronization

**clients/desktop/src/indicators/core/**
- Purpose: Technical analysis engine
- Files:
  - `IndicatorEngine.ts` - Indicator calculations (573 lines)
  - `__tests__/IndicatorEngine.test.ts` - Unit tests

## Key File Locations

**Entry Points:**
- Backend: `backend/cmd/server/main.go`
- Desktop: `clients/desktop/src/main.tsx`
- Broker Admin: `admin/broker-admin/src/app/page.tsx`

**Configuration:**
- Go deps: `backend/go.mod`, `backend/go.sum`
- Frontend deps: `clients/desktop/package.json`, `clients/desktop/bun.lock`
- LP config: `backend/data/lp_config.json`
- TypeScript: `clients/desktop/tsconfig.json`
- Vite: `clients/desktop/vite.config.ts`
- Tests: `clients/desktop/vitest.config.ts`

**Core Logic:**
- Trading engine: `backend/internal/core/engine.go`
- LP management: `backend/lpmanager/manager.go`
- WebSocket: `backend/ws/hub.go`
- Chart rendering: `clients/desktop/src/components/TradingChart.tsx`

**Testing:**
- Frontend tests: `clients/desktop/src/**/__tests__/*.test.ts`
- Test setup: `clients/desktop/src/test/setup.ts`
- No backend tests found

## Naming Conventions

**Files (Go):**
- lowercase: `manager.go`, `engine.go`, `hub.go`
- Test files: `*_test.go` (none found in backend)

**Files (TypeScript):**
- PascalCase components: `TradingChart.tsx`, `IndicatorManager.tsx`
- camelCase hooks: `useIndicators.ts`, `useChartData.ts`
- PascalCase services: `DataCache.ts`, `IndicatorStorage.ts`
- Test files: `*.test.ts` co-located with source

**Directories:**
- Go: lowercase with underscores rarely
- Frontend: PascalCase for component directories, lowercase for utilities

## Where to Add New Code

**New Backend Feature:**
- Handler: `backend/internal/api/handlers/{feature}.go`
- Business logic: `backend/internal/core/` or new package
- Tests: `backend/{package}/{file}_test.go` (currently missing)

**New Frontend Component:**
- Component: `clients/desktop/src/components/{Feature}.tsx`
- Service: `clients/desktop/src/services/{Feature}Service.ts`
- Hook: `clients/desktop/src/hooks/use{Feature}.ts`
- Tests: `clients/desktop/src/**/__tests__/{Feature}.test.ts`

**New LP Integration:**
- Client: `backend/{lpname}/client.go`
- Adapter: `backend/lpmanager/adapters/{lpname}.go`
- Config: Update `backend/data/lp_config.json`

**New Indicator:**
- Logic: `clients/desktop/src/indicators/core/IndicatorEngine.ts`
- Tests: `clients/desktop/src/indicators/core/__tests__/IndicatorEngine.test.ts`

## Special Directories

**backend/data/**
- Purpose: Persistent storage (JSON files)
- Files: `lp_config.json` (LP configuration), `ticks/` (tick history), `ohlc/` (candlestick cache)
- Committed: lp_config.json yes, tick data managed via Git LFS

**clients/desktop/node_modules/**
- Purpose: Frontend dependencies
- Committed: No (in .gitignore)

---

*Structure analysis: 2026-01-15*
*Update when directory structure changes*
