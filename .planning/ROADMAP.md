# Roadmap: MT5 Alternative Broker Platform

## Overview

Transform the existing Trading Engine into a complete, production-ready broker platform rivaling MetaTrader 5. The journey spans 15 phases, from security hardening and infrastructure modernization through advanced trading features to complete broker management tools with comprehensive API documentation. Each phase delivers verifiable capabilities that build toward the goal of enabling brokers to launch and operate a complete trading business.

## Phases

**Phase Numbering:**
- Integer phases (1-15): Planned v1.0 work
- Decimal phases (X.1, X.2): Urgent insertions if needed (marked with INSERTED)

- [x] **Phase 1: Security & Configuration** - Harden platform security and environment configuration
- [ ] **Phase 2: Database Migration** - Replace file storage with production database
- [ ] **Phase 3: Testing Infrastructure** - Establish comprehensive test coverage
- [ ] **Phase 4: Deployment & Operations** - Production deployment with CI/CD and monitoring
- [ ] **Phase 5: Advanced Order Types** - Complete order management system
- [ ] **Phase 6: Risk Management** - Margin and risk controls
- [ ] **Phase 7: Multi-Asset Support** - Multiple asset classes beyond FX/crypto
- [ ] **Phase 8: Client Trading Terminal** - Desktop executable for traders
- [ ] **Phase 9: User & Account Management** - Complete user lifecycle operations
- [ ] **Phase 10: Platform Monitoring** - Real-time health and risk dashboards
- [ ] **Phase 11: Market Configuration** - Dynamic symbol and market management
- [ ] **Phase 12: Reporting & Compliance** - Complete reporting and audit capabilities
- [ ] **Phase 13: Broker Manager Application** - Desktop MT5 Manager clone
- [ ] **Phase 14: Client API Documentation** - Swagger docs for client API
- [ ] **Phase 15: Admin API Documentation** - Swagger docs for admin API

## Phase Details

### Phase 1: Security & Configuration
**Goal**: Platform secured with production-grade credential management and configuration system
**Depends on**: Nothing (first phase)
**Requirements**: SECURITY-01, SECURITY-02, SECURITY-03, SECURITY-04, SECURITY-05
**Success Criteria** (what must be TRUE):
  1. No hardcoded credentials exist in codebase (all via environment variables)
  2. JWT tokens use cryptographically secure secret
  3. WebSocket connections validate origin against whitelist
  4. All passwords stored as bcrypt hashes (no plaintext fallback)
  5. Platform starts successfully using .env configuration
**Research**: Unlikely (security best practices are well-established)
**Plans**: 3 plans ready for execution

Plans:
- [x] 01-environment-secrets-PLAN.md — Environment Configuration & Secret Management (Wave 1) ✅
- [x] 02-websocket-cors-PLAN.md — WebSocket Security & CORS Validation (Wave 1) ✅
- [x] 03-password-security-PLAN.md — Password Security Hardening (Wave 1) ✅

### Phase 2: Database Migration
**Goal**: All application data persisted in production database with ACID guarantees
**Depends on**: Phase 1
**Requirements**: SCALE-01, SCALE-02, SCALE-03, SCALE-04
**Success Criteria** (what must be TRUE):
  1. Database schema created and migrated
  2. Account data loads from database (not JSON files)
  3. Position data persists to database
  4. Trade history queryable from database
  5. Platform restarts without data loss
**Research**: Completed (02-RESEARCH.md)
**Research topics**: PostgreSQL vs MySQL vs SQLite for broker platform, migration without downtime, schema design for trading data
**Plans**: 4 plans ready for execution

Plans:
- [ ] 02-01-PLAN.md — PostgreSQL Foundation & Schema (Wave 1)
- [ ] 02-02-PLAN.md — Repository Pattern Implementation (Wave 1)
- [ ] 02-03-PLAN.md — Trading Engine Database Integration (Wave 2)
- [ ] 02-04-PLAN.md — Audit Trail & Compliance Logging (Wave 2)

### Phase 3: Testing Infrastructure
**Goal**: Comprehensive test coverage provides confidence for refactoring and new features
**Depends on**: Phase 2
**Requirements**: TEST-01, TEST-02, TEST-03, TEST-04, TEST-05, TEST-06, TEST-07
**Success Criteria** (what must be TRUE):
  1. Go test suite runs and passes (backend tests exist)
  2. Core engine behavior verified by tests (account, position, execution)
  3. LP manager integration tests pass
  4. WebSocket hub tests cover connection lifecycle
  5. Frontend tests cover critical components
  6. End-to-end tests verify order flow
  7. Load tests validate platform handles concurrent users
**Research**: Unlikely (testing patterns well-established for Go and React)
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 4: Deployment & Operations
**Goal**: Platform deployable to production with automated CI/CD, monitoring, and operational visibility
**Depends on**: Phase 3
**Requirements**: DEPLOY-01, DEPLOY-02, DEPLOY-03, DEPLOY-04, DEPLOY-05, DEPLOY-06, DEPLOY-07, DEPLOY-08, DEPLOY-09, SCALE-05, SCALE-06, SCALE-07
**Success Criteria** (what must be TRUE):
  1. Platform runs in Docker containers (backend and frontend)
  2. docker-compose starts full stack locally
  3. CI/CD pipeline builds and tests on every commit
  4. Structured logs searchable and filterable
  5. Metrics collected and queryable (Prometheus or equivalent)
  6. Health checks respond correctly
  7. Database backups run automatically
  8. Caching layer reduces database load for tick/OHLC data
  9. LP lookups use O(1) map access
**Research**: Likely (deployment architecture, monitoring stack, caching strategy)
**Research topics**: Docker best practices for Go+React, CI/CD for broker platform, monitoring stack (Prometheus+Grafana vs alternatives), Redis caching patterns
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 5: Advanced Order Types
**Goal**: Traders can use all standard order types (SL, TP, trailing stops, pending orders, OCO)
**Depends on**: Phase 4
**Requirements**: ORDER-01, ORDER-02, ORDER-03, ORDER-04, ORDER-05, ORDER-06, ORDER-07, ORDER-08, ORDER-09, ORDER-10
**Success Criteria** (what must be TRUE):
  1. Trader can place stop-loss order and it executes when price hit
  2. Trader can place take-profit order and it executes when price hit
  3. Trader can set trailing stop that follows price movement
  4. Trader can place pending orders (buy/sell limit, buy/sell stop)
  5. Trader can link orders with OCO (one cancels other)
  6. Trader can modify existing orders (price, SL, TP)
  7. Orders expire automatically when time limit reached
**Research**: Unlikely (order types are standard, implementation patterns clear)
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 6: Risk Management
**Goal**: Platform enforces margin requirements, position limits, and automatic risk controls
**Depends on**: Phase 5
**Requirements**: RISK-01, RISK-02, RISK-03, RISK-04, RISK-05, RISK-06, RISK-07, RISK-08, RISK-09, RISK-10
**Success Criteria** (what must be TRUE):
  1. Margin requirement calculated correctly for each symbol
  2. Trader sees real-time margin level
  3. Margin call alert triggers at configured threshold
  4. Stop-out automatically closes positions when margin critical
  5. Position size limits prevent over-leverage
  6. Maximum open positions enforced
  7. Equity alerts notify trader of account changes
  8. Maximum drawdown protection stops trading when limit hit
  9. Daily loss limits enforced
  10. Leverage controls applied per symbol and account group
**Research**: Likely (margin calculation formulas, industry standard thresholds)
**Research topics**: Margin calculation for different asset classes, stop-out best practices, risk management standards
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 7: Multi-Asset Support
**Goal**: Platform supports trading across all major asset classes (FX, crypto, stocks, commodities, indices, CFDs)
**Depends on**: Phase 6
**Requirements**: ASSET-01, ASSET-02, ASSET-03, ASSET-04, ASSET-05, ASSET-06, ASSET-07, ASSET-08
**Success Criteria** (what must be TRUE):
  1. Additional forex pairs tradable (beyond current support)
  2. Cryptocurrency CFDs tradable (extended Binance integration)
  3. Stocks/equities tradable (new LP integrated)
  4. Commodities tradable (Gold, Silver, Oil, etc.)
  5. Indices tradable (S&P 500, NASDAQ, etc.)
  6. CFD contracts properly configured
  7. Symbol metadata correct (tick size, contract size, margin)
  8. Asset-specific rules enforced (market hours, sessions)
**Research**: Likely (LP integrations for stocks/commodities, CFD contract specifications)
**Research topics**: Stock market LP options (Interactive Brokers API, etc.), commodity data feeds, CFD margin and rollover, trading session management
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 8: Client Trading Terminal
**Goal**: Professional desktop trading application (exe) with all features traders need
**Depends on**: Phase 7
**Requirements**: TERMINAL-01, TERMINAL-02, TERMINAL-03, TERMINAL-04, TERMINAL-05, TERMINAL-06, TERMINAL-07, TERMINAL-08, TERMINAL-09, TERMINAL-10
**Success Criteria** (what must be TRUE):
  1. TradingChart component refactored (no 900+ line files)
  2. Desktop application packaged as executable
  3. Auto-update downloads and installs new versions
  4. Trader can switch between multiple accounts
  5. One-click trading executes orders instantly
  6. Order templates save and restore configurations
  7. Trading history panel shows past trades
  8. Economic calendar displays scheduled events
  9. News feed shows real-time market news
  10. Price alerts trigger notifications
**Research**: Likely (desktop packaging, auto-update mechanism)
**Research topics**: Electron vs Tauri for desktop packaging, auto-update implementation (Squirrel, etc.), news feed integration
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 9: User & Account Management
**Goal**: Brokers can manage complete user lifecycle (create, KYC, fund, suspend, terminate)
**Depends on**: Phase 8
**Requirements**: USER-01, USER-02, USER-03, USER-04, USER-05, USER-06, USER-07, USER-08, USER-09, USER-10, USER-11, USER-12
**Success Criteria** (what must be TRUE):
  1. Broker admin can create new user accounts
  2. Users can upload KYC documents
  3. Broker admin can approve/reject KYC
  4. Broker can process manual deposits
  5. Automated deposits work via payment gateway
  6. Broker can process manual withdrawals
  7. Automated withdrawals work via payment gateway
  8. Balance adjustments recorded with audit trail
  9. Broker can suspend accounts (reversible)
  10. Broker can terminate accounts (permanent)
  11. Account groups assign different trading conditions
  12. Broker can search and filter users
**Research**: Likely (payment gateway integration, KYC compliance)
**Research topics**: Payment gateway options (Stripe, PayPal, crypto), KYC compliance requirements, document verification services
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 10: Platform Monitoring
**Goal**: Brokers have real-time visibility into platform health, LP status, and risk exposure
**Depends on**: Phase 9
**Requirements**: MONITOR-01, MONITOR-02, MONITOR-03, MONITOR-04, MONITOR-05, MONITOR-06, MONITOR-07, MONITOR-08, MONITOR-09, MONITOR-10
**Success Criteria** (what must be TRUE):
  1. System health dashboard shows CPU, memory, disk, network
  2. LP connection status visible for all liquidity providers
  3. LP failover triggers automatically when connection lost
  4. Trade volume monitored in real-time
  5. Platform-wide risk exposure dashboard shows aggregate positions
  6. Active positions viewable across all accounts
  7. Active orders viewable across all accounts
  8. WebSocket connection count tracked
  9. Error rates monitored with alerts
  10. Historical metrics show trends over time
**Research**: Unlikely (monitoring patterns well-established)
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 11: Market Configuration
**Goal**: Brokers can dynamically configure markets, symbols, spreads, and trading rules
**Depends on**: Phase 10
**Requirements**: MARKET-01, MARKET-02, MARKET-03, MARKET-04, MARKET-05, MARKET-06, MARKET-07, MARKET-08, MARKET-09, MARKET-10, MARKET-11
**Success Criteria** (what must be TRUE):
  1. Broker can add new symbols via admin interface
  2. Broker can remove symbols (with position validation)
  3. Spreads configurable per symbol
  4. Markup/commission configurable per symbol
  5. Leverage limits set per symbol
  6. Leverage limits set per account group
  7. Trading hours defined per symbol
  8. Session breaks configured (lunch, overnight)
  9. Market depth settings customizable
  10. Liquidity pool allocation configurable
  11. Symbols organized into groups/categories
**Research**: Unlikely (market configuration patterns standard)
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 12: Reporting & Compliance
**Goal**: Complete reporting suite for operations, compliance, and regulatory requirements
**Depends on**: Phase 11
**Requirements**: REPORT-01, REPORT-02, REPORT-03, REPORT-04, REPORT-05, REPORT-06, REPORT-07, REPORT-08, REPORT-09, REPORT-10, REPORT-11
**Success Criteria** (what must be TRUE):
  1. Trade history reports generated with filters
  2. Reports exported in CSV, Excel, PDF
  3. Audit logs capture all admin actions
  4. Regulatory compliance reports generated
  5. P&L reports available per user
  6. P&L reports available per symbol
  7. Platform-wide P&L calculated
  8. Commission and fee reports generated
  9. Deposit/withdrawal reports available
  10. Risk exposure reports generated
  11. Scheduled reports run automatically (daily, weekly, monthly)
**Research**: Likely (regulatory compliance requirements, report formats)
**Research topics**: Broker regulatory requirements, audit log standards, automated report generation libraries
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 13: Broker Manager Application
**Goal**: Professional desktop broker management tool (MT5 Manager clone) with all admin capabilities
**Depends on**: Phase 12
**Requirements**: MANAGER-01, MANAGER-02, MANAGER-03, MANAGER-04, MANAGER-05, MANAGER-06, MANAGER-07, MANAGER-08, MANAGER-09, MANAGER-10
**Success Criteria** (what must be TRUE):
  1. Broker manager packaged as desktop executable
  2. Role-based access controls users by permission level
  3. Admin sessions managed securely
  4. Dashboard shows key metrics (users, volume, P&L, exposure)
  5. Real-time updates via WebSocket
  6. Panels detachable into separate windows
  7. Notification system alerts on important events
  8. Bulk operations process multiple records
  9. Data export works for all major entities
  10. Auto-update keeps broker manager current
**Research**: Likely (desktop application architecture for complex admin tool)
**Research topics**: Multi-window desktop app patterns, role-based UI rendering, real-time dashboard updates
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 14: Client API Documentation
**Goal**: Complete Swagger documentation enables integration with client trading API
**Depends on**: Phase 13
**Requirements**: API-CLIENT-01, API-CLIENT-02, API-CLIENT-03, API-CLIENT-04, API-CLIENT-05, API-CLIENT-06, API-CLIENT-07, API-CLIENT-08, API-CLIENT-09, API-CLIENT-10, API-CLIENT-11
**Success Criteria** (what must be TRUE):
  1. Swagger/OpenAPI spec exists for client API
  2. Authentication endpoints documented
  3. Account endpoints documented
  4. Order endpoints documented
  5. Position endpoints documented
  6. Market data endpoints documented
  7. Trade history endpoints documented
  8. WebSocket API documented
  9. Error codes and formats documented
  10. Rate limiting rules documented
  11. Interactive Swagger UI available for testing
**Research**: Unlikely (Swagger/OpenAPI standards well-established)
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

### Phase 15: Admin API Documentation
**Goal**: Complete Swagger documentation enables integration with broker admin API
**Depends on**: Phase 14
**Requirements**: API-ADMIN-01, API-ADMIN-02, API-ADMIN-03, API-ADMIN-04, API-ADMIN-05, API-ADMIN-06, API-ADMIN-07, API-ADMIN-08, API-ADMIN-09, API-ADMIN-10, API-ADMIN-11
**Success Criteria** (what must be TRUE):
  1. Swagger/OpenAPI spec exists for admin API
  2. User management endpoints documented
  3. Financial operations endpoints documented
  4. Platform monitoring endpoints documented
  5. Market configuration endpoints documented
  6. Reporting endpoints documented
  7. Audit log endpoints documented
  8. Role/permission management endpoints documented
  9. Error codes and formats documented
  10. Rate limiting rules documented
  11. Interactive Swagger UI available for testing
**Research**: Unlikely (Swagger/OpenAPI standards well-established)
**Plans**: TBD

Plans:
- [ ] TBD (to be planned)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → ... → 15

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Security & Configuration | 3/3 | ✅ Complete | 2026-01-16 |
| 2. Database Migration | 0/TBD | Not started | - |
| 3. Testing Infrastructure | 0/TBD | Not started | - |
| 4. Deployment & Operations | 0/TBD | Not started | - |
| 5. Advanced Order Types | 0/TBD | Not started | - |
| 6. Risk Management | 0/TBD | Not started | - |
| 7. Multi-Asset Support | 0/TBD | Not started | - |
| 8. Client Trading Terminal | 0/TBD | Not started | - |
| 9. User & Account Management | 0/TBD | Not started | - |
| 10. Platform Monitoring | 0/TBD | Not started | - |
| 11. Market Configuration | 0/TBD | Not started | - |
| 12. Reporting & Compliance | 0/TBD | Not started | - |
| 13. Broker Manager Application | 0/TBD | Not started | - |
| 14. Client API Documentation | 0/TBD | Not started | - |
| 15. Admin API Documentation | 0/TBD | Not started | - |
