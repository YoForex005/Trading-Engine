# Architecture

**Analysis Date:** 2026-01-18

## Pattern Overview

**Overall:** Event-Driven Microkernel with Hybrid Execution Model (B-Book/A-Book)

**Key Characteristics:**
- Real-time market data distribution through WebSocket hub
- Dual execution paths: Internal B-Book engine and external LP routing via FIX protocol
- In-memory state management with file-based persistence
- Pub/sub architecture for price feeds and account updates

## Layers

**Market Data Layer:**
- Purpose: Aggregates quotes from multiple liquidity providers and distributes to clients
- Location: `backend/lpmanager/`, `backend/ws/`
- Contains: LP adapters (Binance, OANDA), WebSocket hub, quote aggregation
- Depends on: External LP APIs (REST/WebSocket), FIX gateway for institutional LPs
- Used by: B-Book engine (for execution prices), frontend clients (for charts/quotes)

**Execution Layer:**
- Purpose: Processes trading orders through either internal B-Book engine or external LP routing
- Location: `backend/internal/core/`, `backend/fix/`, `backend/router/`
- Contains: B-Book engine (positions, orders, trades), FIX gateway (LP connectivity), smart router (A/B-Book decision)
- Depends on: Market data layer (for current prices), ledger (for balance updates)
- Used by: API handlers (order placement), risk engine (margin checks)

**Account Management Layer:**
- Purpose: Manages user accounts, balances, and transaction history
- Location: `backend/internal/core/ledger.go`, `backend/auth/`
- Contains: Ledger system (deposits/withdrawals/PnL), account state, authentication service
- Depends on: Nothing (core domain layer)
- Used by: Execution layer (balance checks), API layer (account queries)

**Persistence Layer:**
- Purpose: Stores tick data, OHLC bars, and session state
- Location: `backend/tickstore/`, `backend/fixstore/`
- Contains: Daily tick files, OHLC cache, FIX sequence number persistence
- Depends on: File system
- Used by: Market data layer (tick storage), API layer (historical data queries), FIX gateway (session recovery)

**API Layer:**
- Purpose: Exposes HTTP REST endpoints and WebSocket streams for clients
- Location: `backend/api/`, `backend/internal/api/handlers/`
- Contains: HTTP handlers, CORS middleware, route definitions
- Depends on: All layers (orchestrates system)
- Used by: Frontend clients (desktop, admin dashboards)

**Client Layer:**
- Purpose: User interfaces for trading and administration
- Location: `clients/desktop/`, `admin/broker-admin/`, `admin/super-admin/`
- Contains: React/Next.js applications with real-time WebSocket connections
- Depends on: API layer
- Used by: End users (traders, admins)

## Data Flow

**Market Data Flow:**

1. LP adapters connect to external providers (Binance WebSocket, OANDA REST, YOFX FIX)
2. Quotes arrive in LP-specific format, converted to unified `Quote` struct
3. LPManager aggregates quotes into single channel `quotesChan`
4. Main goroutine pipes quotes to WebSocket Hub as `MarketTick`
5. Hub broadcasts to all connected WebSocket clients (non-blocking with buffered channels)
6. Hub notifies B-Book engine of price updates (triggers SL/TP checks)
7. Hub persists tick to DailyStore and updates OHLC cache

**Order Execution Flow (B-Book):**

1. Client sends POST to `/api/orders/market` with symbol, side, volume, SL/TP
2. API handler extracts accountID from JWT token
3. Engine validates: account status, symbol exists, volume within limits
4. Engine fetches current price from Hub via `priceCallback`
5. Engine calculates required margin and checks free margin
6. Engine creates Order (FILLED status) and Position (OPEN status)
7. Engine deducts commission from account balance
8. Ledger records commission entry
9. Engine stores Trade record for audit trail
10. API returns Position to client
11. Hub sends position update via WebSocket

**Order Execution Flow (A-Book - FIX):**

1. Client sends order to API
2. Router decides A-Book based on rules (user group, volume, symbol)
3. FIX Gateway constructs NewOrderSingle (35=D) message
4. Gateway sends to LP via TCP/TLS connection with sequence number tracking
5. LP responds with ExecutionReport (35=8)
6. Gateway parses execution, extracts fill price/status
7. OMS updates internal position state
8. Position broadcast to client via WebSocket

**Account Balance Flow:**

1. Admin initiates deposit via `/admin/deposit` with accountID, amount, method
2. Ledger validates amount > 0
3. Ledger locks mutex, updates balance cache, creates LedgerEntry
4. Account balance increased, LedgerEntry persisted
5. API returns updated balance
6. Frontend polls or receives WebSocket update

**State Management:**
- In-memory maps for accounts, positions, orders (locked with `sync.RWMutex`)
- No database - state resets on server restart (demo mode)
- Tick data persisted to JSON files in `data/ticks/{symbol}/{date}.json`
- FIX sequence numbers persisted to `fixstore/{sessionID}/` for session recovery

## Key Abstractions

**Engine (B-Book):**
- Purpose: Core trading engine for internal execution
- Examples: `backend/internal/core/engine.go`
- Pattern: In-memory state machine with mutex-protected maps

**Hub (WebSocket):**
- Purpose: Real-time event distribution to connected clients
- Examples: `backend/ws/hub.go`
- Pattern: Goroutine-based pub/sub with buffered channels (2048 buffer to prevent blocking)

**LPManager:**
- Purpose: Multi-source liquidity aggregation
- Examples: `backend/lpmanager/manager.go`, `backend/lpmanager/adapters/`
- Pattern: Registry pattern with adapter interface

**FIXGateway:**
- Purpose: FIX protocol session management and order routing
- Examples: `backend/fix/gateway.go`
- Pattern: Stateful session with sequence number tracking, heartbeat monitoring

**TickStore:**
- Purpose: Time-series data persistence with daily rotation
- Examples: `backend/tickstore/service.go`, `backend/tickstore/daily_store.go`
- Pattern: Write-optimized append-only log with in-memory OHLC cache

**Ledger:**
- Purpose: Double-entry accounting for all balance changes
- Examples: `backend/internal/core/ledger.go`
- Pattern: Append-only transaction log with balance cache

## Entry Points

**HTTP Server:**
- Location: `backend/cmd/server/main.go`
- Triggers: Application startup
- Responsibilities: Wire all dependencies, register routes, start WebSocket hub, auto-connect FIX sessions

**WebSocket Hub Goroutine:**
- Location: `backend/ws/hub.go` (Run method)
- Triggers: `go hub.Run()` in main
- Responsibilities: Manage client connections, broadcast price updates, handle subscribe/unsubscribe

**LP Quote Aggregator:**
- Location: `backend/lpmanager/manager.go` (StartQuoteAggregation)
- Triggers: Called from main after initialization
- Responsibilities: Start all enabled LP adapters, aggregate quotes into unified channel

**Quote Pipe Goroutine:**
- Location: `backend/cmd/server/main.go` (line 125)
- Triggers: Launched in main to bridge LPManager → Hub
- Responsibilities: Read from `lpMgr.GetQuotesChan()`, convert to MarketTick, call `hub.BroadcastTick()`

**FIX Session Auto-Connect:**
- Location: `backend/cmd/server/main.go` (line 475)
- Triggers: 3 seconds after server start
- Responsibilities: Connect YOFX1 FIX session for institutional LP

## Error Handling

**Strategy:** Fail-fast with error returns, log-and-continue for non-critical paths

**Patterns:**
- API handlers return HTTP error codes (400, 404, 500) with JSON error messages
- Engine methods return `(result, error)` tuples - caller decides whether to abort
- WebSocket send failures use non-blocking select with default case (drop message rather than block)
- FIX protocol errors logged but session continues (resend requests, sequence resets)
- LP disconnections trigger auto-reconnect with exponential backoff

## Cross-Cutting Concerns

**Logging:** Standard library `log` package with structured prefixes `[B-Book]`, `[Hub]`, `[LPManager]`, `[FIX]`

**Validation:** Input validation at API boundary (handlers), business rules in engine layer

**Authentication:** JWT tokens generated in `backend/auth/service.go`, validated in middleware (not shown but referenced in handlers)

**Concurrency:**
- All shared state protected by `sync.RWMutex` (read-heavy workloads use RLock)
- WebSocket clients have individual buffered send channels (1024 messages)
- Hub broadcast channel buffered (2048) to prevent price feed blocking on slow consumers
- FIX gateway uses mutex for sequence number access (critical for protocol compliance)

**Session Management:**
- HTTP: Stateless with JWT tokens
- WebSocket: Stateful connections managed by Hub (register/unregister channels)
- FIX: Stateful with sequence number persistence in `fixstore/` directory

---

*Architecture analysis: 2026-01-18*
