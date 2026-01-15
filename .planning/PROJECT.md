# Trading Engine - MT5 Alternative Broker Platform

## What This Is

A complete broker trading platform ecosystem rivaling MetaTrader 5, providing professional-grade tools for both traders and broker administrators. The platform includes a client trading terminal (desktop exe), broker management tools (MT5 Manager clone as desktop exe), comprehensive REST APIs with Swagger documentation, and production-ready infrastructure for launching and operating a trading business.

## Core Value

Brokers can launch and operate a complete trading business rivaling MT5 in capability, with professional client trading tools and comprehensive broker management systems.

## Requirements

### Validated

Existing capabilities from current codebase:

- ✓ B-Book execution engine with position management — existing
- ✓ Real-time price feeds from multiple LPs (Binance, OANDA, FlexyMarkets) — existing
- ✓ WebSocket streaming for live prices — existing
- ✓ React desktop client with advanced charting — existing
- ✓ Technical indicators system — existing
- ✓ Basic order execution (market orders) — existing
- ✓ Account and position tracking — existing
- ✓ LP manager with pluggable adapters — existing

### Active

v1.0 Foundation - Production-ready platform with core trading and admin capabilities:

#### Production Infrastructure
- [ ] Security hardening: Rotate hardcoded credentials (OANDA API key, JWT secret)
- [ ] Security hardening: Fix CORS validation on WebSocket
- [ ] Security hardening: Remove plaintext password fallback
- [ ] Testing infrastructure: Add backend test framework and core engine tests
- [ ] Testing infrastructure: Add LP manager integration tests
- [ ] Testing infrastructure: Add WebSocket hub tests
- [ ] Scalability: Implement database layer (replace file-based storage)
- [ ] Scalability: Add caching layer for tick data and OHLC
- [ ] Deployment: Create Docker containers for backend and frontend
- [ ] Deployment: Implement CI/CD pipeline
- [ ] Deployment: Add monitoring and logging infrastructure
- [ ] Configuration: Environment variable support (.env)

#### Client Trading Terminal (Desktop exe)
- [ ] Advanced order types: Stop-loss orders
- [ ] Advanced order types: Take-profit orders
- [ ] Advanced order types: Trailing stop orders
- [ ] Advanced order types: Pending orders (buy/sell limit, buy/sell stop)
- [ ] Advanced order types: OCO (One-Cancels-Other) orders
- [ ] Risk management: Margin requirement calculation
- [ ] Risk management: Position size limits
- [ ] Risk management: Equity and margin monitoring with alerts
- [ ] Risk management: Max drawdown protection
- [ ] Multiple asset classes: Forex pairs (extend current support)
- [ ] Multiple asset classes: Cryptocurrencies (extend current Binance integration)
- [ ] Multiple asset classes: Stocks/Equities
- [ ] Multiple asset classes: Commodities (Gold, Silver, Oil, etc.)
- [ ] Multiple asset classes: Indices (S&P 500, NASDAQ, etc.)
- [ ] Multiple asset classes: CFDs
- [ ] Desktop application: Package as standalone executable
- [ ] Desktop application: Auto-update mechanism

#### Broker Manager (Desktop exe - MT5 Manager Clone)
- [ ] User & account management: Create and manage user accounts
- [ ] User & account management: KYC document upload and verification
- [ ] User & account management: Account deposits (manual and automated)
- [ ] User & account management: Account withdrawals (manual and automated)
- [ ] User & account management: Balance adjustments and corrections
- [ ] User & account management: Account suspension and termination
- [ ] Platform monitoring: System health dashboard
- [ ] Platform monitoring: LP connection status and failover management
- [ ] Platform monitoring: Real-time trade volume monitoring
- [ ] Platform monitoring: Platform-wide risk exposure dashboard
- [ ] Platform monitoring: Active positions and orders overview
- [ ] Market configuration: Add and remove tradable symbols
- [ ] Market configuration: Configure spreads per symbol
- [ ] Market configuration: Set leverage limits per symbol and account group
- [ ] Market configuration: Define trading hours and session breaks
- [ ] Market configuration: Market depth and liquidity settings
- [ ] Reporting & compliance: Trade history reports (filterable by user, date, symbol)
- [ ] Reporting & compliance: Audit logs for all admin actions
- [ ] Reporting & compliance: Regulatory compliance reports
- [ ] Reporting & compliance: P&L reports (per user, per symbol, platform-wide)
- [ ] Desktop application: Package as standalone executable
- [ ] Desktop application: Role-based access control (super admin, broker admin, support)

#### API Documentation
- [ ] Client API: Swagger/OpenAPI documentation for all trading endpoints
- [ ] Client API: Authentication endpoints (login, signup, password reset)
- [ ] Client API: Account endpoints (summary, balance, margin)
- [ ] Client API: Order endpoints (place, modify, cancel)
- [ ] Client API: Position endpoints (list, close, modify)
- [ ] Client API: Market data endpoints (symbols, quotes, OHLC)
- [ ] Client API: Trade history endpoints
- [ ] Admin/Broker API: Swagger/OpenAPI documentation for all broker management endpoints
- [ ] Admin/Broker API: User management endpoints
- [ ] Admin/Broker API: Financial operations endpoints (deposits, withdrawals)
- [ ] Admin/Broker API: Platform monitoring endpoints
- [ ] Admin/Broker API: Market configuration endpoints
- [ ] Admin/Broker API: Reporting endpoints
- [ ] API: Rate limiting and throttling
- [ ] API: Comprehensive error handling and status codes
- [ ] API: API versioning strategy

### Out of Scope

- Expert Advisors / Algorithmic trading — Complex feature, defer to v2.0+
- Mobile applications (iOS/Android) — Desktop-first approach, mobile in future milestone
- Social trading / Copy trading — Not core to broker operations, defer to v2.0+
- Multi-language support — English-first for v1.0
- White-label customization — Single branded platform for v1.0
- Cryptocurrency wallet integration — Focus on CFD trading, not custody

## Context

**Existing Codebase:**
- Go 1.24.0 backend with modular service architecture
- React 19.2.0 + Vite 7.2.4 frontend with TypeScript 5.9.3
- WebSocket real-time streaming (gorilla/websocket)
- JWT authentication (golang-jwt/jwt v5.3.0)
- Three LP adapters: Binance (WebSocket), OANDA (REST/Streaming), FlexyMarkets (Socket.IO)
- Lightweight Charts 4.2.3 for advanced charting
- Technical indicator engine with 10+ indicators
- File-based storage (needs database migration)

**Critical Issues to Address:**
- CRITICAL: Hardcoded OANDA API credentials in source code
- CRITICAL: Weak JWT secret ("super_secret_dev_key_do_not_use_in_prod")
- HIGH: CORS disabled on WebSocket (security risk)
- HIGH: Zero backend tests (0 Go test files found)
- MEDIUM: Oversized components (TradingChart.tsx: 952 lines)
- MEDIUM: Hardcoded localhost URLs throughout frontend

**Reference Model:**
- MetaTrader 5 (MT5) Manager as primary reference for broker management tools
- MT5 client terminal as reference for trading interface features
- Industry-standard broker operations and compliance requirements

## Constraints

- **Tech Stack**: Go backend, React/TypeScript frontend — Existing architecture proven, maintain consistency
- **Desktop Platform**: Electron or similar for cross-platform desktop executables — Required for exe packaging
- **Incremental Releases**: Ship v1.0 foundation, iterate toward full MT5 parity — Pragmatic approach to deliver value incrementally
- **API-First Design**: All functionality accessible via documented REST APIs — Enables integration and future expansion
- **Security**: Production-grade security from day one — Handling real money requires zero compromise
- **Compatibility**: Support Windows, macOS, Linux for desktop applications — Broad market coverage

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Separate desktop apps for client and admin | MT5 model proven, clear separation of concerns | — Pending |
| Swagger/OpenAPI for API documentation | Industry standard, auto-generates client SDKs, interactive testing | — Pending |
| Database migration from files | File-based storage not production-ready, need ACID guarantees and scalability | — Pending |
| YOLO + Comprehensive + Parallel execution | Maximum throughput with thorough coverage for large-scale project | — Pending |
| MT5 Manager as reference model | Established industry standard, clear feature requirements, proven UX patterns | — Pending |

---
*Last updated: 2026-01-15 after initialization*
