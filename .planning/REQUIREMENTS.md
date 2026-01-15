# Requirements: MT5 Alternative Broker Platform

**Defined:** 2026-01-15
**Core Value:** Brokers can launch and operate a complete trading business rivaling MT5 in capability, with professional client trading tools and comprehensive broker management systems.

## v1 Requirements

Requirements for v1.0 - Production-ready foundation with core trading and admin capabilities.

### Security (SECURITY)

- [ ] **SECURITY-01**: Rotate hardcoded OANDA API credentials to environment variables
- [ ] **SECURITY-02**: Replace weak JWT secret with cryptographically random key in environment variable
- [ ] **SECURITY-03**: Implement CORS validation with origin whitelist for WebSocket connections
- [ ] **SECURITY-04**: Remove plaintext password fallback, enforce bcrypt for all passwords
- [ ] **SECURITY-05**: Implement environment variable configuration system (.env support)

### Testing Infrastructure (TEST)

- [ ] **TEST-01**: Add Go testing framework and test structure
- [ ] **TEST-02**: Core engine unit tests (account, position, order execution)
- [ ] **TEST-03**: LP manager integration tests (adapter lifecycle, quote aggregation)
- [ ] **TEST-04**: WebSocket hub tests (connection management, broadcast)
- [ ] **TEST-05**: Frontend component tests (TradingChart, indicator system)
- [ ] **TEST-06**: End-to-end integration tests (order flow, position management)
- [ ] **TEST-07**: Load testing infrastructure (concurrent users, high-frequency ticks)

### Scalability (SCALE)

- [ ] **SCALE-01**: Design and implement database schema (PostgreSQL or similar)
- [ ] **SCALE-02**: Migrate account data from file storage to database
- [ ] **SCALE-03**: Migrate position data from file storage to database
- [ ] **SCALE-04**: Migrate trade history from file storage to database
- [x] **SCALE-05**: Implement caching layer for tick data (Redis or similar)
- [x] **SCALE-06**: Implement caching layer for OHLC data
- [x] **SCALE-07**: Optimize LP manager for O(1) lookups (map-based instead of linear search)

### Deployment & Operations (DEPLOY)

- [x] **DEPLOY-01**: Create Docker container for Go backend
- [x] **DEPLOY-02**: Create Docker container for React frontend
- [x] **DEPLOY-03**: Docker Compose setup for local development
- [x] **DEPLOY-04**: Implement CI/CD pipeline (GitHub Actions or similar)
- [x] **DEPLOY-05**: Add structured logging with log levels
- [x] **DEPLOY-06**: Implement monitoring (Prometheus metrics or similar)
- [x] **DEPLOY-07**: Add health check endpoints
- [x] **DEPLOY-08**: Implement automated backups for database
- [x] **DEPLOY-09**: Create deployment documentation

### Advanced Order Types (ORDER)

- [ ] **ORDER-01**: Stop-loss order implementation (backend and frontend)
- [ ] **ORDER-02**: Take-profit order implementation (backend and frontend)
- [ ] **ORDER-03**: Trailing stop orders (dynamic stop-loss following price)
- [ ] **ORDER-04**: Pending orders: Buy Limit
- [ ] **ORDER-05**: Pending orders: Sell Limit
- [ ] **ORDER-06**: Pending orders: Buy Stop
- [ ] **ORDER-07**: Pending orders: Sell Stop
- [ ] **ORDER-08**: OCO (One-Cancels-Other) order linking
- [ ] **ORDER-09**: Order modification (price, quantity, SL/TP)
- [ ] **ORDER-10**: Order expiration (time-based order cancellation)

### Risk Management (RISK)

- [ ] **RISK-01**: Margin requirement calculation engine
- [ ] **RISK-02**: Real-time margin level monitoring
- [ ] **RISK-03**: Margin call alerts (configurable thresholds)
- [ ] **RISK-04**: Stop-out mechanism (automatic position liquidation)
- [ ] **RISK-05**: Position size limits (per symbol, per account)
- [ ] **RISK-06**: Maximum open positions limit
- [ ] **RISK-07**: Equity monitoring with alerts
- [ ] **RISK-08**: Maximum drawdown protection
- [ ] **RISK-09**: Daily loss limits
- [ ] **RISK-10**: Leverage controls (per symbol, per account group)

### Multiple Asset Classes (ASSET)

- [ ] **ASSET-01**: Forex pairs - extend current support with more pairs
- [ ] **ASSET-02**: Cryptocurrencies - extend Binance integration
- [ ] **ASSET-03**: Stocks/Equities - add LP integration (Interactive Brokers or similar)
- [ ] **ASSET-04**: Commodities (Gold, Silver, Oil, Gas, etc.)
- [ ] **ASSET-05**: Indices (S&P 500, NASDAQ, Dow Jones, DAX, etc.)
- [ ] **ASSET-06**: CFDs (Contracts for Difference) support
- [ ] **ASSET-07**: Symbol metadata (tick size, contract size, margin requirements)
- [ ] **ASSET-08**: Asset-specific trading rules (market hours, trading sessions)

### Client Trading Terminal (TERMINAL)

- [ ] **TERMINAL-01**: Refactor oversized TradingChart component (extract smaller components)
- [ ] **TERMINAL-02**: Package desktop application as executable (Electron or Tauri)
- [ ] **TERMINAL-03**: Auto-update mechanism for desktop client
- [ ] **TERMINAL-04**: Multi-account support (switch between accounts)
- [ ] **TERMINAL-05**: One-click trading interface
- [ ] **TERMINAL-06**: Order templates (save order configurations)
- [ ] **TERMINAL-07**: Trading history panel
- [ ] **TERMINAL-08**: Economic calendar integration
- [ ] **TERMINAL-09**: News feed integration
- [ ] **TERMINAL-10**: Alert system (price alerts, indicator alerts)

### User & Account Management (USER)

- [ ] **USER-01**: Create and manage user accounts (broker manager)
- [ ] **USER-02**: KYC document upload and storage
- [ ] **USER-03**: KYC verification workflow (pending, approved, rejected)
- [ ] **USER-04**: Manual deposit processing
- [ ] **USER-05**: Automated deposit processing (payment gateway integration)
- [ ] **USER-06**: Manual withdrawal processing
- [ ] **USER-07**: Automated withdrawal processing
- [ ] **USER-08**: Balance adjustments and corrections (with audit trail)
- [ ] **USER-09**: Account suspension (temporary block)
- [ ] **USER-10**: Account termination (permanent closure)
- [ ] **USER-11**: Account groups (different leverage, spreads per group)
- [ ] **USER-12**: User search and filtering

### Platform Monitoring (MONITOR)

- [ ] **MONITOR-01**: System health dashboard (CPU, memory, disk, network)
- [ ] **MONITOR-02**: LP connection status monitoring
- [ ] **MONITOR-03**: LP failover and redundancy management
- [ ] **MONITOR-04**: Real-time trade volume monitoring
- [ ] **MONITOR-05**: Platform-wide risk exposure dashboard
- [ ] **MONITOR-06**: Active positions overview (all accounts)
- [ ] **MONITOR-07**: Active orders overview (all accounts)
- [ ] **MONITOR-08**: WebSocket connection monitoring (active clients)
- [ ] **MONITOR-09**: Error rate and alert system
- [ ] **MONITOR-10**: Historical metrics and trends

### Market Configuration (MARKET)

- [ ] **MARKET-01**: Add new tradable symbols (symbol creation interface)
- [ ] **MARKET-02**: Remove symbols (with validation if positions exist)
- [ ] **MARKET-03**: Configure spreads per symbol (fixed or percentage)
- [ ] **MARKET-04**: Configure markup/commission per symbol
- [ ] **MARKET-05**: Set leverage limits per symbol
- [ ] **MARKET-06**: Set leverage limits per account group
- [ ] **MARKET-07**: Define trading hours (market open/close times)
- [ ] **MARKET-08**: Define session breaks (lunch breaks, etc.)
- [ ] **MARKET-09**: Market depth configuration
- [ ] **MARKET-10**: Liquidity pool settings per symbol
- [ ] **MARKET-11**: Symbol groups and categories

### Reporting & Compliance (REPORT)

- [ ] **REPORT-01**: Trade history reports (filterable by user, date, symbol)
- [ ] **REPORT-02**: Export trade reports (CSV, Excel, PDF)
- [ ] **REPORT-03**: Audit logs for all admin actions
- [ ] **REPORT-04**: Regulatory compliance report generation
- [ ] **REPORT-05**: P&L reports per user
- [ ] **REPORT-06**: P&L reports per symbol
- [ ] **REPORT-07**: Platform-wide P&L aggregation
- [ ] **REPORT-08**: Commission and fee reports
- [ ] **REPORT-09**: Deposit/withdrawal reports
- [ ] **REPORT-10**: Risk exposure reports
- [ ] **REPORT-11**: Scheduled report generation (daily, weekly, monthly)

### Broker Manager Application (MANAGER)

- [ ] **MANAGER-01**: Package broker manager as desktop executable
- [ ] **MANAGER-02**: Role-based access control (super admin, broker admin, support)
- [ ] **MANAGER-03**: User session management for broker manager
- [ ] **MANAGER-04**: Dashboard with key metrics (users, volume, P&L, exposure)
- [ ] **MANAGER-05**: Real-time updates via WebSocket
- [ ] **MANAGER-06**: Multi-window support (detachable panels)
- [ ] **MANAGER-07**: Notification system (alerts, warnings, errors)
- [ ] **MANAGER-08**: Bulk operations (batch user updates, mass emails)
- [ ] **MANAGER-09**: Data export capabilities
- [ ] **MANAGER-10**: Auto-update mechanism for broker manager

### Client API Documentation (API-CLIENT)

- [ ] **API-CLIENT-01**: Swagger/OpenAPI specification for client API
- [ ] **API-CLIENT-02**: Authentication endpoints documentation (login, signup, password reset)
- [ ] **API-CLIENT-03**: Account endpoints documentation (summary, balance, margin)
- [ ] **API-CLIENT-04**: Order endpoints documentation (place, modify, cancel)
- [ ] **API-CLIENT-05**: Position endpoints documentation (list, close, modify)
- [ ] **API-CLIENT-06**: Market data endpoints documentation (symbols, quotes, OHLC)
- [ ] **API-CLIENT-07**: Trade history endpoints documentation
- [ ] **API-CLIENT-08**: WebSocket API documentation (subscription, message formats)
- [ ] **API-CLIENT-09**: Error codes and response formats documentation
- [ ] **API-CLIENT-10**: Rate limiting documentation
- [ ] **API-CLIENT-11**: Interactive API testing interface (Swagger UI)

### Admin/Broker API Documentation (API-ADMIN)

- [ ] **API-ADMIN-01**: Swagger/OpenAPI specification for admin API
- [ ] **API-ADMIN-02**: User management endpoints documentation
- [ ] **API-ADMIN-03**: Financial operations endpoints documentation (deposits, withdrawals)
- [ ] **API-ADMIN-04**: Platform monitoring endpoints documentation
- [ ] **API-ADMIN-05**: Market configuration endpoints documentation
- [ ] **API-ADMIN-06**: Reporting endpoints documentation
- [ ] **API-ADMIN-07**: Audit log endpoints documentation
- [ ] **API-ADMIN-08**: Role and permission management endpoints
- [ ] **API-ADMIN-09**: Error codes and response formats documentation
- [ ] **API-ADMIN-10**: Rate limiting documentation
- [ ] **API-ADMIN-11**: Interactive API testing interface (Swagger UI)

## v2 Requirements

Deferred to future releases after v1.0 foundation.

### Algorithmic Trading (ALGO)

- **ALGO-01**: Expert Advisor (EA) framework for automated trading
- **ALGO-02**: Strategy backtesting engine
- **ALGO-03**: Strategy optimization tools
- **ALGO-04**: MQL-like scripting language support
- **ALGO-05**: EA marketplace/repository

### Mobile Applications (MOBILE)

- **MOBILE-01**: iOS native trading application
- **MOBILE-02**: Android native trading application
- **MOBILE-03**: Mobile push notifications
- **MOBILE-04**: Mobile biometric authentication

### Social Trading (SOCIAL)

- **SOCIAL-01**: Copy trading system (follow traders)
- **SOCIAL-02**: Strategy provider leaderboard
- **SOCIAL-03**: Social trading analytics
- **SOCIAL-04**: Revenue sharing for strategy providers

### Advanced Features (ADVANCED)

- **ADVANCED-01**: Multi-language support (i18n)
- **ADVANCED-02**: White-label customization (branding, colors, logo)
- **ADVANCED-03**: Plugin/extension system
- **ADVANCED-04**: Advanced charting features (custom studies, drawing tools)
- **ADVANCED-05**: Market depth visualization (Level II quotes)

## Out of Scope

| Feature | Reason |
|---------|--------|
| Cryptocurrency wallet integration | Focus on CFD trading model, not crypto custody which requires different regulatory compliance |
| Blockchain/DeFi integration | Not core to traditional broker operations, different business model |
| Peer-to-peer trading | Platform operates as broker, not marketplace |
| Margin lending to other users | Complex regulatory implications, focus on broker-to-client relationship |
| Proprietary trading (prop trading) | Platform for retail brokers, not prop firms |
| Options trading | Highly complex, defer to future milestone after core platform stable |
| Futures trading | Complex margin requirements and rollover, defer to v2.0+ |

## Traceability

Which phases cover which requirements. Updated by create-roadmap workflow.

| Requirement | Phase | Status |
|-------------|-------|--------|
| SECURITY-01 | Phase 1 | Pending |
| SECURITY-02 | Phase 1 | Pending |
| SECURITY-03 | Phase 1 | Pending |
| SECURITY-04 | Phase 1 | Pending |
| SECURITY-05 | Phase 1 | Pending |
| SCALE-01 | Phase 2 | Pending |
| SCALE-02 | Phase 2 | Pending |
| SCALE-03 | Phase 2 | Pending |
| SCALE-04 | Phase 2 | Pending |
| TEST-01 | Phase 3 | Pending |
| TEST-02 | Phase 3 | Pending |
| TEST-03 | Phase 3 | Pending |
| TEST-04 | Phase 3 | Pending |
| TEST-05 | Phase 3 | Pending |
| TEST-06 | Phase 3 | Pending |
| TEST-07 | Phase 3 | Pending |
| DEPLOY-01 | Phase 4 | Pending |
| DEPLOY-02 | Phase 4 | Pending |
| DEPLOY-03 | Phase 4 | Pending |
| DEPLOY-04 | Phase 4 | Pending |
| DEPLOY-05 | Phase 4 | Pending |
| DEPLOY-06 | Phase 4 | Pending |
| DEPLOY-07 | Phase 4 | Pending |
| DEPLOY-08 | Phase 4 | Pending |
| DEPLOY-09 | Phase 4 | Pending |
| SCALE-05 | Phase 4 | Pending |
| SCALE-06 | Phase 4 | Pending |
| SCALE-07 | Phase 4 | Pending |
| ORDER-01 | Phase 5 | Pending |
| ORDER-02 | Phase 5 | Pending |
| ORDER-03 | Phase 5 | Pending |
| ORDER-04 | Phase 5 | Pending |
| ORDER-05 | Phase 5 | Pending |
| ORDER-06 | Phase 5 | Pending |
| ORDER-07 | Phase 5 | Pending |
| ORDER-08 | Phase 5 | Pending |
| ORDER-09 | Phase 5 | Pending |
| ORDER-10 | Phase 5 | Pending |
| RISK-01 | Phase 6 | Pending |
| RISK-02 | Phase 6 | Pending |
| RISK-03 | Phase 6 | Pending |
| RISK-04 | Phase 6 | Pending |
| RISK-05 | Phase 6 | Pending |
| RISK-06 | Phase 6 | Pending |
| RISK-07 | Phase 6 | Pending |
| RISK-08 | Phase 6 | Pending |
| RISK-09 | Phase 6 | Pending |
| RISK-10 | Phase 6 | Pending |
| ASSET-01 | Phase 7 | Pending |
| ASSET-02 | Phase 7 | Pending |
| ASSET-03 | Phase 7 | Pending |
| ASSET-04 | Phase 7 | Pending |
| ASSET-05 | Phase 7 | Pending |
| ASSET-06 | Phase 7 | Pending |
| ASSET-07 | Phase 7 | Pending |
| ASSET-08 | Phase 7 | Pending |
| TERMINAL-01 | Phase 8 | Pending |
| TERMINAL-02 | Phase 8 | Pending |
| TERMINAL-03 | Phase 8 | Pending |
| TERMINAL-04 | Phase 8 | Pending |
| TERMINAL-05 | Phase 8 | Pending |
| TERMINAL-06 | Phase 8 | Pending |
| TERMINAL-07 | Phase 8 | Pending |
| TERMINAL-08 | Phase 8 | Pending |
| TERMINAL-09 | Phase 8 | Pending |
| TERMINAL-10 | Phase 8 | Pending |
| USER-01 | Phase 9 | Pending |
| USER-02 | Phase 9 | Pending |
| USER-03 | Phase 9 | Pending |
| USER-04 | Phase 9 | Pending |
| USER-05 | Phase 9 | Pending |
| USER-06 | Phase 9 | Pending |
| USER-07 | Phase 9 | Pending |
| USER-08 | Phase 9 | Pending |
| USER-09 | Phase 9 | Pending |
| USER-10 | Phase 9 | Pending |
| USER-11 | Phase 9 | Pending |
| USER-12 | Phase 9 | Pending |
| MONITOR-01 | Phase 10 | Pending |
| MONITOR-02 | Phase 10 | Pending |
| MONITOR-03 | Phase 10 | Pending |
| MONITOR-04 | Phase 10 | Pending |
| MONITOR-05 | Phase 10 | Pending |
| MONITOR-06 | Phase 10 | Pending |
| MONITOR-07 | Phase 10 | Pending |
| MONITOR-08 | Phase 10 | Pending |
| MONITOR-09 | Phase 10 | Pending |
| MONITOR-10 | Phase 10 | Pending |
| MARKET-01 | Phase 11 | Pending |
| MARKET-02 | Phase 11 | Pending |
| MARKET-03 | Phase 11 | Pending |
| MARKET-04 | Phase 11 | Pending |
| MARKET-05 | Phase 11 | Pending |
| MARKET-06 | Phase 11 | Pending |
| MARKET-07 | Phase 11 | Pending |
| MARKET-08 | Phase 11 | Pending |
| MARKET-09 | Phase 11 | Pending |
| MARKET-10 | Phase 11 | Pending |
| MARKET-11 | Phase 11 | Pending |
| REPORT-01 | Phase 12 | Pending |
| REPORT-02 | Phase 12 | Pending |
| REPORT-03 | Phase 12 | Pending |
| REPORT-04 | Phase 12 | Pending |
| REPORT-05 | Phase 12 | Pending |
| REPORT-06 | Phase 12 | Pending |
| REPORT-07 | Phase 12 | Pending |
| REPORT-08 | Phase 12 | Pending |
| REPORT-09 | Phase 12 | Pending |
| REPORT-10 | Phase 12 | Pending |
| REPORT-11 | Phase 12 | Pending |
| MANAGER-01 | Phase 13 | Pending |
| MANAGER-02 | Phase 13 | Pending |
| MANAGER-03 | Phase 13 | Pending |
| MANAGER-04 | Phase 13 | Pending |
| MANAGER-05 | Phase 13 | Pending |
| MANAGER-06 | Phase 13 | Pending |
| MANAGER-07 | Phase 13 | Pending |
| MANAGER-08 | Phase 13 | Pending |
| MANAGER-09 | Phase 13 | Pending |
| MANAGER-10 | Phase 13 | Pending |
| API-CLIENT-01 | Phase 14 | Pending |
| API-CLIENT-02 | Phase 14 | Pending |
| API-CLIENT-03 | Phase 14 | Pending |
| API-CLIENT-04 | Phase 14 | Pending |
| API-CLIENT-05 | Phase 14 | Pending |
| API-CLIENT-06 | Phase 14 | Pending |
| API-CLIENT-07 | Phase 14 | Pending |
| API-CLIENT-08 | Phase 14 | Pending |
| API-CLIENT-09 | Phase 14 | Pending |
| API-CLIENT-10 | Phase 14 | Pending |
| API-CLIENT-11 | Phase 14 | Pending |
| API-ADMIN-01 | Phase 15 | Pending |
| API-ADMIN-02 | Phase 15 | Pending |
| API-ADMIN-03 | Phase 15 | Pending |
| API-ADMIN-04 | Phase 15 | Pending |
| API-ADMIN-05 | Phase 15 | Pending |
| API-ADMIN-06 | Phase 15 | Pending |
| API-ADMIN-07 | Phase 15 | Pending |
| API-ADMIN-08 | Phase 15 | Pending |
| API-ADMIN-09 | Phase 15 | Pending |
| API-ADMIN-10 | Phase 15 | Pending |
| API-ADMIN-11 | Phase 15 | Pending |

**Coverage:**
- v1 requirements: 116 total
- Mapped to phases: 116
- Unmapped: 0 ✓

---
*Requirements defined: 2026-01-15*
*Last updated: 2026-01-15 after roadmap creation*
