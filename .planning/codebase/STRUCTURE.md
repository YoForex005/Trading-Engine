# Codebase Structure

**Analysis Date:** 2026-01-18

## Directory Layout

```
trading-engine/
├── backend/                  # Go backend services (44 .go files)
│   ├── cmd/                  # Entry points
│   ├── api/                  # HTTP server
│   ├── internal/             # Private packages (core, handlers)
│   ├── fix/                  # FIX protocol gateway
│   ├── ws/                   # WebSocket hub
│   ├── lpmanager/            # LP adapters and aggregation
│   ├── tickstore/            # Tick persistence
│   ├── auth/                 # Authentication
│   ├── bbook/                # Legacy B-Book (superseded by internal/core)
│   ├── oms/                  # Order Management System
│   ├── risk/                 # Risk engine
│   ├── router/               # Smart order router
│   ├── orders/               # Pending orders, trailing stops
│   ├── oanda/                # OANDA API client
│   ├── binance/              # Binance API client
│   ├── flexymarkets/         # FlexyMarkets client
│   ├── data/                 # JSON data files (ticks, OHLC, config)
│   ├── fixstore/             # FIX sequence number persistence
│   └── server                # Compiled binary
├── clients/
│   └── desktop/              # React + Vite trading terminal
│       ├── src/
│       │   ├── components/   # UI components
│       │   ├── App.tsx       # Main app component
│       │   └── main.tsx      # Entry point
│       └── package.json
├── admin/
│   ├── broker-admin/         # Next.js broker dashboard
│   │   ├── src/
│   │   │   ├── app/          # Next.js 14 app router
│   │   │   ├── components/   # Admin UI components
│   │   │   └── types/        # TypeScript types
│   │   └── package.json
│   └── super-admin/          # Next.js super admin platform
├── scripts/                  # Build/deployment scripts
├── .planning/                # GSD planning documents
│   └── codebase/             # Codebase analysis (this file)
├── README.md                 # Project documentation
└── TECHNICAL_COMMITTEE_SPEC.md
```

## Directory Purposes

**backend/cmd/server/:**
- Purpose: Application entry point and dependency wiring
- Contains: `main.go` (553 lines) - initializes all services, registers routes, starts goroutines
- Key files: `main.go`

**backend/api/:**
- Purpose: HTTP server and legacy route handlers
- Contains: Server struct, OANDA passthrough handlers, risk calculators
- Key files: `server.go` (691 lines)

**backend/internal/:**
- Purpose: Private packages not exported for external use
- Contains: Core B-Book engine, API handlers, domain models
- Key files:
  - `internal/core/engine.go` (821 lines) - trading engine
  - `internal/core/ledger.go` - accounting system
  - `internal/core/pnl.go` - P/L calculations
  - `internal/api/handlers/*.go` - HTTP handlers for B-Book API

**backend/fix/:**
- Purpose: FIX 4.4 protocol implementation for LP connectivity
- Contains: Session management, message parsing, order routing
- Key files:
  - `gateway.go` (2527 lines) - main FIX gateway
  - `sessions.json` - LP session configurations
  - `test_*.go` - integration tests

**backend/ws/:**
- Purpose: WebSocket hub for real-time market data and updates
- Contains: Client management, broadcast logic, subscription handling
- Key files: `hub.go` (247 lines)

**backend/lpmanager/:**
- Purpose: Multi-LP quote aggregation
- Contains: Manager, registry, adapter interface
- Key files:
  - `manager.go` - orchestrates all LPs
  - `adapters/binance.go` - Binance WebSocket adapter
  - `adapters/oanda.go` - OANDA REST adapter
  - `lp.go` - interfaces and types
  - `registry.go` - adapter registry

**backend/tickstore/:**
- Purpose: Time-series tick data storage with OHLC caching
- Contains: Daily file rotation, in-memory buffers, OHLC aggregation
- Key files:
  - `service.go` - main TickStore service
  - `daily_store.go` - file-per-day persistence
  - `ohlc_cache.go` - real-time OHLC bars (M1, M5, M15, H1, H4, D1)

**backend/auth/:**
- Purpose: Authentication and session management
- Contains: JWT token generation/validation, user service
- Key files: `service.go`, `token.go`

**backend/data/:**
- Purpose: JSON file storage for configuration and market data
- Contains: LP configs, tick files (organized by symbol/date), OHLC files (by symbol/timeframe)
- Key files:
  - `lp_config.json` - LP connection settings
  - `ticks/{SYMBOL}/{YYYY-MM-DD}.json` - daily tick files
  - `ohlc/{SYMBOL}/{TIMEFRAME}.json` - OHLC bars

**backend/fixstore/:**
- Purpose: FIX session state persistence
- Contains: Sequence numbers, session metadata
- Key files: `{sessionID}/seqnums.txt`

**backend/bbook/:**
- Purpose: Legacy B-Book implementation (deprecated in favor of internal/core)
- Contains: Old engine, ledger, PnL (kept for backward compatibility)
- Key files: `engine.go`, `ledger.go`, `pnl.go`, `api.go`

**backend/oms/:**
- Purpose: Order Management System (legacy, not actively used)
- Contains: Order state management
- Key files: `service.go`

**backend/risk/:**
- Purpose: Risk calculations and margin checks
- Contains: Risk engine, lot size calculator
- Key files: `engine.go`, `calculator.go`

**backend/router/:**
- Purpose: Smart order router for A-Book/B-Book decision
- Contains: Routing logic based on user groups, volumes
- Key files: `smart_router.go`

**backend/orders/:**
- Purpose: Advanced order types (pending, trailing stops)
- Contains: Pending order management, position tracking
- Key files: `pending.go`, `position.go`, `trailing.go`

**clients/desktop/:**
- Purpose: Trading terminal UI
- Contains: React components, TradingView charts, WebSocket client
- Key files:
  - `src/App.tsx` (20,636 lines) - main application
  - `src/components/` - UI components

**admin/broker-admin/:**
- Purpose: Broker administrator dashboard
- Contains: Account management, symbol config, LP management
- Key files:
  - `src/app/page.tsx` - main dashboard
  - `src/components/dashboard/` - admin components

**admin/super-admin/:**
- Purpose: Platform-level administration (broker management)
- Contains: Super admin controls
- Key files: `src/app/page.tsx`

## Key File Locations

**Entry Points:**
- `backend/cmd/server/main.go`: HTTP server startup, dependency injection
- `clients/desktop/src/main.tsx`: React app entry
- `admin/broker-admin/src/app/page.tsx`: Admin dashboard entry

**Configuration:**
- `backend/go.mod`: Go dependencies (jwt, websocket, crypto, uuid)
- `backend/data/lp_config.json`: LP connection configs
- `backend/fix/sessions.json`: FIX session configurations
- `clients/desktop/package.json`: Frontend dependencies (React 19, Vite, TailwindCSS)
- `admin/broker-admin/package.json`: Admin dependencies (Next.js 16)

**Core Logic:**
- `backend/internal/core/engine.go`: B-Book trading engine (accounts, positions, orders)
- `backend/fix/gateway.go`: FIX protocol implementation (session management, order routing)
- `backend/ws/hub.go`: WebSocket broadcast hub
- `backend/lpmanager/manager.go`: LP aggregation orchestrator

**Testing:**
- `backend/fix/test_*.go`: FIX integration tests

## Naming Conventions

**Files:**
- Go: `snake_case.go` (e.g., `smart_router.go`, `daily_store.go`)
- TypeScript: `PascalCase.tsx` for components, `kebab-case.ts` for utilities
- Config: `lowercase.json` or `UPPERCASE.md`

**Directories:**
- Go packages: `lowercase` (single word) or `camelCase` (e.g., `lpmanager`, `tickstore`)
- Frontend: `kebab-case` (e.g., `broker-admin`)

**Go Types:**
- Structs: `PascalCase` (e.g., `Engine`, `MarketTick`, `LPSession`)
- Interfaces: `PascalCase` with `I` prefix or descriptive noun (e.g., `LPAdapter`)
- Constants: `MixedCase` with prefix (e.g., `MsgTypeLogon`, `DefaultStoreDir`)

## Where to Add New Code

**New LP Adapter:**
- Primary code: `backend/lpmanager/adapters/{provider}.go`
- Tests: `backend/lpmanager/adapters/{provider}_test.go`
- Register: In `main.go` via `lpMgr.RegisterAdapter()`

**New API Endpoint (B-Book):**
- Implementation: `backend/internal/api/handlers/{domain}.go`
- Wire route: `backend/cmd/server/main.go` (add `http.HandleFunc()`)

**New Frontend Component:**
- Implementation: `clients/desktop/src/components/{ComponentName}.tsx`
- Import: In `src/App.tsx`

**New Admin Feature:**
- Implementation: `admin/broker-admin/src/components/dashboard/{FeatureName}.tsx`
- Add to dashboard: `admin/broker-admin/src/app/page.tsx`

**New Symbol/Market:**
- Add to: `backend/internal/core/engine.go` (initDefaultSymbols method)
- Or: Dynamic via `/admin/symbols` API

**Utilities:**
- Backend: `backend/{domain}/` (create new package if cross-cutting)
- Frontend: `clients/desktop/src/lib/` or `clients/desktop/src/utils/`

## Special Directories

**backend/data/:**
- Purpose: Runtime data storage (ticks, OHLC, configs)
- Generated: Yes (tick files created on-the-fly)
- Committed: Partial (configs committed, tick data in .gitattributes for LFS)

**backend/fixstore/:**
- Purpose: FIX session state persistence
- Generated: Yes (sequence numbers created during FIX session)
- Committed: No (excluded in .gitignore)

**backend/server (binary):**
- Purpose: Compiled Go executable
- Generated: Yes (via `go build`)
- Committed: No

**clients/desktop/node_modules/:**
- Purpose: npm dependencies
- Generated: Yes (via `npm install`)
- Committed: No

**admin/*/node_modules/:**
- Purpose: npm dependencies
- Generated: Yes
- Committed: No

**.planning/:**
- Purpose: GSD planning and codebase analysis
- Generated: Yes (by GSD commands)
- Committed: Yes (tracked for project knowledge)

**Migration Note:**
The codebase has legacy components (`backend/bbook/`, `backend/oms/`) that are superseded by newer implementations (`backend/internal/core/`). The main entry point (`cmd/server/main.go`) uses the new internal/core engine. Legacy code is retained for backward compatibility but should not be extended.

---

*Structure analysis: 2026-01-18*
