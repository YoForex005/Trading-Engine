# External Integrations

**Analysis Date:** 2026-01-18

## APIs & External Services

**Liquidity Providers:**
- OANDA - Forex and CFD price streaming and execution
  - SDK/Client: Custom HTTP/streaming client in `backend/oanda/client.go`
  - Auth: Hardcoded API key (OANDA_API_KEY = "977e1a77e25bac3a688011d6b0e845dd-8e3ab3a7682d9351af4c33be65e89b70")
  - Account: OANDA_ACCOUNT_ID = "101-004-37008470-002"
  - Environment: fxPractice (practice.oanda.com)
  - Streaming URL: https://stream-fxpractice.oanda.com
  - REST URL: https://api-fxpractice.oanda.com

- Binance - Cryptocurrency price streaming
  - SDK/Client: WebSocket client in `backend/binance/client.go`
  - WebSocket URL: wss://stream.binance.com:9443/stream
  - Auth: None (public market data)
  - Streams: Book ticker for BTCUSDT, ETHUSDT

**FIX Protocol Gateway:**
- YOFX T4B Trading Server
  - Connection: FIX 4.4 protocol over TCP
  - Host: 23.106.238.138:12336
  - SenderCompID: YOFX1
  - TargetCompID: YOFX
  - Username: YOFX1
  - Password: Brand#143 (stored in `backend/fix/sessions.json`)
  - Trading Account: 50153
  - SSL: Disabled (plain TCP)
  - Heartbeat: 30 seconds
  - Implementation: Custom FIX engine in `backend/fix/gateway.go`
  - Capabilities: Session management, order routing (35=D), execution reports (35=8), market data subscription (35=V/W/Y), position requests (35=AN/AO/AP)

## Data Storage

**Databases:**
- PostgreSQL with TimescaleDB extension (recommended)
  - Connection: Not configured (schema exists in `backend/database/schema.sql`)
  - Client: Native Go SQL (no ORM detected)
  - Schema: Users, accounts, orders, trades, positions, tick_history hypertable
  - Current State: Schema defined but database not actively used (in-memory engine in use)

**File Storage:**
- Local filesystem for tick data
  - Location: `backend/data/ticks/{SYMBOL}/{YYYY-MM-DD}.json`
  - Format: JSON files per symbol per day
  - OHLC cache: `backend/data/ohlc/{SYMBOL}/{TIMEFRAME}.json`
  - FIX sequence numbers: `backend/fixstore/` (persisted message store)
  - Git LFS: Configured for `*.json` files in `backend/data/ticks/`

**Caching:**
- In-memory caching in Go
  - Latest prices: `ws.Hub.latestPrices` map
  - OHLC data: `tickstore.TickStore` with configurable max ticks per symbol (default 50000)
  - No Redis or external cache detected

## Authentication & Identity

**Auth Provider:**
- Custom JWT-based authentication
  - Implementation: `backend/auth/service.go`
  - Token generation: JWT v5 with HS256 signing
  - Password hashing: bcrypt (cost 10)
  - Session storage: In-memory (no persistent session store)
  - Admin credentials: username="admin", password="password" (bcrypt hashed on startup)
  - Client auth: Account-based (username/password from in-memory engine)

## Monitoring & Observability

**Error Tracking:**
- None - Standard Go log package only

**Logs:**
- Standard output logging via Go `log` package
  - Format: `[Component] Message` (e.g., "[B-Book]", "[Binance]", "[FIX]", "[LPManager]")
  - Destination: stdout/stderr
  - No structured logging framework
  - Log files: Development logs in `.next/dev/logs/` for admin frontends

## CI/CD & Deployment

**Hosting:**
- Not configured - Local development only

**CI Pipeline:**
- None - No GitHub Actions, GitLab CI, or other CI config detected

**Build Scripts:**
- `backend/dev_runner.sh` - Development server launcher
- `backend/force_restart.sh` - Server restart script
- No production build scripts

## Environment Configuration

**Required env vars:**
- None currently required (all configs hardcoded or in JSON files)

**Secrets location:**
- Hardcoded in source code:
  - `backend/cmd/server/main.go`: OANDA_API_KEY, OANDA_ACCOUNT_ID
  - `backend/fix/sessions.json`: FIX credentials
- Admin password: Generated at runtime (bcrypt of "password")

**Configuration files:**
- `backend/data/lp_config.json` - LP manager settings (OANDA enabled, Binance disabled)
- `backend/fix/sessions.json` - FIX session configurations
- No environment-specific configs (.env files not in use)

## Webhooks & Callbacks

**Incoming:**
- None detected

**Outgoing:**
- None detected

## Internal Services

**WebSocket Server:**
- Real-time market data broadcasting
  - Endpoint: `ws://localhost:8080/ws`
  - Protocol: Gorilla WebSocket with JSON messages
  - Message types: Market ticks, positions, account updates
  - Client throttling: 100ms frontend throttling (10 FPS)
  - Buffering: 2048-message channel buffer for non-blocking broadcasts

**HTTP REST API:**
- Trading operations and account management
  - Base URL: `http://localhost:8080`
  - CORS: Wildcard enabled (`Access-Control-Allow-Origin: *`)
  - Endpoints documented in `backend/swagger.yaml`
  - Key endpoints: `/login`, `/order`, `/account`, `/positions`, `/admin/*`

**Internal LP Manager:**
- Multi-LP aggregation system
  - Location: `backend/lpmanager/manager.go`
  - Adapters: `backend/lpmanager/adapters/binance.go`, `backend/lpmanager/adapters/oanda.go`
  - Registry: Dynamic LP registration with priority-based quote selection
  - Config persistence: JSON file at `backend/data/lp_config.json`
  - Quote channel: Buffered channel (1000 quotes) for non-blocking aggregation

---

*Integration audit: 2026-01-18*
