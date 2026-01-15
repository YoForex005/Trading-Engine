# Architecture

**Analysis Date:** 2026-01-15

## Pattern Overview

**Overall:** Hybrid Modular Monolith with Service-Oriented Design

**Key Characteristics:**
- Go backend with modular packages for specific concerns
- React SPA frontend with WebSocket real-time data
- Component-based admin panels (Next.js)
- In-memory state with file persistence
- Multi-LP adapter pattern for external integrations

## Layers

**API Layer:**
- Purpose: HTTP/REST and WebSocket transport
- Contains: HTTP handlers, WebSocket hub, CORS middleware
- Files: `backend/cmd/server/main.go`, `backend/internal/api/handlers/*.go`, `backend/ws/hub.go`
- Depends on: Core engine, services
- Used by: Frontend clients (desktop, admin)

**Core Business Logic:**
- Purpose: Trading engine, position management, P/L calculations
- Contains: B-Book execution engine, ledger, order management
- Files: `backend/internal/core/engine.go`, `backend/internal/core/ledger.go`, `backend/internal/core/pnl.go`
- Depends on: Nothing (pure business logic)
- Used by: API handlers

**Service Layer:**
- Purpose: Peripheral services (LP management, authentication, tick storage)
- Contains: LP Manager, TickStore, Auth Service
- Files: `backend/lpmanager/manager.go`, `backend/tickstore/service.go`, `backend/auth/service.go`
- Depends on: Core abstractions
- Used by: API layer, main server

**Integration Layer:**
- Purpose: External liquidity provider connections
- Contains: LP adapters (Binance, OANDA, FlexyMarkets)
- Files: `backend/lpmanager/adapters/*.go`, `backend/binance/client.go`, `backend/oanda/client.go`
- Depends on: LP Manager interface
- Used by: LP Manager

## Data Flow

**Real-Time Market Data Flow:**

1. LP Adapters connect to external APIs (Binance WSS, OANDA Streaming)
2. Adapters stream quotes → LP Manager aggregates (`backend/lpmanager/manager.go`)
3. Aggregated quotes → WebSocket Hub (`backend/ws/hub.go`)
4. Hub broadcasts `MarketTick` JSON to connected clients
5. Frontend receives via WebSocket, buffers, updates UI at 10 FPS (`clients/desktop/src/App.tsx`)
6. Ticks cached in TickStore for historical retrieval (`backend/tickstore/service.go`)

**Order Execution Flow:**

1. User clicks Buy/Sell → Frontend POST to `/api/orders/market`
2. API handler routes to Core Engine (`backend/internal/core/engine.go`)
3. Engine validates symbol, creates Position, updates Ledger
4. P/L Engine recalculates equity and margin (`backend/internal/core/pnl.go`)
5. Position details returned to client
6. WebSocket broadcasts account update to client

**State Management:**
- Backend: In-memory maps with RWMutex synchronization
- Frontend: React state + WebSocket subscriptions
- Persistence: File-based (JSON) for configuration and tick history

## Key Abstractions

**LPAdapter Interface:**
- Purpose: Pluggable liquidity provider connections
- Implementations: `BinanceAdapter`, `OANDAAdapter`, `FlexyAdapter`
- Files: `backend/lpmanager/lp.go`, `backend/lpmanager/adapters/*.go`
- Pattern: Interface segregation - each LP implements standard methods

**Core Engine:**
- Purpose: B-Book execution engine maintaining accounts, positions, orders
- File: `backend/internal/core/engine.go`
- Pattern: Thread-safe service with RWMutex, callback injection for prices
- Methods: CreateAccount, ExecuteMarketOrder, ClosePosition, GetPositions

**WebSocket Hub:**
- Purpose: Broadcast market data and account updates to connected clients
- File: `backend/ws/hub.go`
- Pattern: Pub-sub hub with client registry and broadcast channel
- Features: Latest price cache, disabled symbol filtering, LP priority handling

**Ledger:**
- Purpose: Transaction history for accounts (deposits, withdrawals, P/L)
- File: `backend/internal/core/ledger.go`
- Pattern: Append-only transaction log with running balance

## Entry Points

**Backend:**
- Main server: `backend/cmd/server/main.go` - Initializes all services, registers routes
- HTTP server starts on :8080
- WebSocket endpoint: `/ws`
- API routes: `/login`, `/api/account/*`, `/api/positions/*`, `/api/orders/*`, `/admin/*`

**Frontend:**
- Desktop client: `clients/desktop/src/main.tsx` → `clients/desktop/src/App.tsx`
- Broker admin: `admin/broker-admin/src/app/page.tsx`
- Super admin: `admin/super-admin/src/app/page.tsx`

## Error Handling

**Strategy:** Throw errors, catch at boundaries

**Patterns:**
- Backend: Return error values, log at handler level, send HTTP error responses
- Frontend: Try/catch in async functions, ErrorBoundary for React components
- Gap: Many ignored errors in `backend/oanda/client.go` (blank error assignments)

## Cross-Cutting Concerns

**Logging:**
- Go: `log.Printf()` statements throughout
- Format: `[Service] Message` pattern (e.g., `[LP Manager]`, `[Binance]`)

**Authentication:**
- JWT-based tokens (`backend/auth/token.go`)
- bcrypt password hashing (`backend/auth/service.go`)
- Token validation via middleware (not implemented globally)

**CORS:**
- Custom `cors()` function called in every handler
- Files: `backend/internal/api/handlers/api.go`, `backend/bbook/api.go`
- Pattern: Allow all origins (security concern)

**Synchronization:**
- RWMutex for read-heavy operations (price lookups, account queries)
- Channels for async communication (quotes, broadcast, stop signals)

---

*Architecture analysis: 2026-01-15*
*Update when major patterns change*
