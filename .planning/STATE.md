# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-01-15)

**Core value:** Brokers can launch and operate a complete trading business rivaling MT5 in capability, with professional client trading tools and comprehensive broker management systems.
**Current focus:** Phase 5 — Advanced Order Types (In Progress)

## Current Position

Phase: 6 of 15 (Risk Management)
Plan: 4 of 6 in current phase (06-01, 06-02, 06-03, 06-04 complete)
Status: In Progress
Last activity: 2026-01-16 — Completed 06-04-PLAN.md (Pre-Trade Risk Validation)

Progress: ▓▓▓░░░░░░░ 28% (5/15 phases, 23/total plans)

## Performance Metrics

**Velocity:**
- Total plans completed: 22
- Average duration: ~35 min per plan
- Total execution time: 1 day (2026-01-16)

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Security & Configuration | 3/3 | 1 day | ~8 hours |
| 2. Database Migration | 4/4 | ~3 hours | ~45 min |
| 3. Testing Infrastructure | 3/7 | ~90 min | ~30 min |
| 4. Deployment Operations | 8/8 | ~2.3 hours | ~17 min |
| 5. Advanced Order Types | 2/4 | ~90 min | ~45 min |
| 6. Risk Management | 3/6 | ~115 min | ~38 min |
**Recent Trend:**
- Last 22 plans: Phase 4 complete with documentation, Phase 3 partial, Phase 5 partial, Phase 6 in progress (22/22 plans)
- Trend: Excellent execution velocity, Phase 4 fully documented (3,148 lines)
- Trend: Deployment documentation provides comprehensive operations runbook with troubleshooting
## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

| Phase | Decision | Rationale |
|-------|----------|-----------|
| 02-01 | Used pgx v5 directly instead of database/sql | 30-50% performance improvement for PostgreSQL-specific operations |
| 02-01 | Connection pool: 20 max, 5 min connections | Optimized for trading platform based on (CPU cores * 2) + 1 baseline |
| 02-01 | DECIMAL for financial values, TIMESTAMPTZ for timestamps | Production-grade financial data handling with timezone awareness |
| 02-01 | Singleton pattern for connection pool | Application-wide pool reuse, initialized once at startup |
| 02-02 | Use REPEATABLE READ isolation for financial operations | Ensures consistency for balance updates and position closing per PostgreSQL best practices |
| 02-02 | Repository pattern with dependency injection | Enables testability and separates data access from business logic |
| 02-02 | Trades are immutable (no update methods) | Execution records should never be modified after creation |
| 02-02 | Pagination for trade queries | Large trading histories need efficient querying with limit/offset |
| 02-03 | Dependency injection for repositories | Pass all 4 repositories to Engine constructor for clean testability and separation of concerns |
| 02-03 | In-memory caching strategy | Keep accounts/positions/orders in memory for performance, write through to database |
| 02-03 | Idempotent migration | Check for existing data before migrating, safe to run on every startup |
| 02-03 | Keep deprecated persistence.go | File retained for rollback safety during Phase 2, can be removed after database proven stable |
| 03-01 | Selected govalues/decimal over shopspring/decimal | Modern Go idioms with generics, cross-validated via fuzz testing, banker's rounding, zero dependencies |
| 03-01 | Test utility package in internal/ | Provides shared helpers for all backend tests with decimal assertions and test data builders |
| 03-01 | Test fixtures in testdata/ | Follows Go convention for test fixtures, all financial values as strings (not floats) |
| 03-02 | Mock WebSocket for testing | Prevents actual connections during tests, includes simulateMessage helper for controlled testing |
| 03-02 | Reusable mock data generators | createMock* functions provide consistent test data for OHLC, ticks, and accounts |
| 03-02 | Type-only imports for TypeScript types | Required for verbatimModuleSyntax compatibility |
| 03-03 | Used float64 for tests matching current codebase | Tests written for actual implementation, decimal migration deferred |
| 03-03 | Repository tests skip pending database setup | Full integration tests with actual database in Plan 03-05 |
| 03-03 | All tests use table-driven pattern with t.Parallel() | Go idiom for concurrent test execution and maintainability |
| 04-06 | Map-based LP lookups with lpConfigMap | O(1) direct access vs O(n) iteration for high-frequency LP operations in trading platform |
| 04-03 | Use dorny/paths-filter@v3 for monorepo change detection | Saves 50-70% CI time by only running affected service tests in GitHub Actions |
| 04-03 | BuildKit cache with type=gha for Docker builds | 30-50% faster builds via GitHub Actions cache, cache-from reads, cache-to writes with mode=max |
| 04-03 | Only publish Docker images on main branch | Conserves GitHub Container Registry resources, prevents PR pollution |
| 04-07 | Use pg_dump with custom format and gzip compression | Space-efficient backups with standard PostgreSQL tools for disaster recovery |
| 04-07 | Schedule backups every 6 hours | Balances recovery granularity with storage costs and resource usage |
| 06-01 | DECIMAL columns as strings in Go repositories | Prevents float64 precision errors that have caused production incidents (LSE halt, €12M German bank fine) |
| 06-01 | Generated column for free_margin calculation | PostgreSQL GENERATED ALWAYS AS ensures free_margin = equity - used_margin without app-level risk |
| 06-01 | ESMA leverage limits in seed data | Provides regulatory-compliant starting configuration (30:1 major pairs, 20:1 minors, 5:1 stocks, 2:1 crypto) |
| 06-01 | Transaction-aware repository methods | Both standalone and tx-aware upsert methods for margin calculation atomicity |
| 04-07 | Dual retention strategy (7-day local, 30-day artifacts) | Quick access for recent backups, long-term recovery capability via GitHub artifacts |
| 05-02 | Trailing stop triggers move with favorable price only | BUY: trigger = ask - delta (rises with price). SELL: trigger = bid + delta (falls with price). Prevents moving against trader |
| 05-02 | Update trailing stops before checking triggers | updateTrailingStops() called before checkPriceTriggers() ensures latest market price used for adjustment |
| 05-02 | Max trailing delta = 10% of position price | Validation prevents unreasonable trailing distances that could trigger prematurely |
| 04-07 | GitHub Actions artifacts for MVP backup storage | Simpler than S3/cloud storage for development, recommend migration to cloud for production |
| 06-02 | Integrated govalues/decimal v0.1.36 for financial calculations | Eliminates float64 precision errors that caused LSE halt and €12M German bank fine |
| 06-02 | Created wrapper utilities (MustParse, Parse, ToString) | Provides consistent decimal API across codebase with panic vs error handling patterns |
| 06-02 | NewFromFloat64 marked migration-only with WARNING | Float64 conversion still has precision issues - only for migrating existing data |
| 06-02 | AssertDecimalNear for epsilon-based comparison | Some calculations have acceptable rounding variance, need both exact and near-equality assertions |
| 05-03 | Unified /api/orders/pending endpoint for all pending types | Single endpoint handles BUY_LIMIT, SELL_LIMIT, BUY_STOP, SELL_STOP - simpler API surface |
| 05-03 | Validate trigger price at order creation time | Prevents traders from placing impossible orders (e.g., BUY_LIMIT above market) |
| 05-03 | Separate order types instead of type+side combination | Using BUY_LIMIT vs LIMIT+BUY makes validation clearer and reduces ambiguity |
| 06-03 | Panic on decimal arithmetic overflow | Financial calculations shouldn't silently fail - overflow is a programming error requiring immediate attention |
| 06-03 | Event-driven margin calculation (not periodic) | Calculate on every position change per ESMA requirements - prevents negative balance scenarios |
| 06-03 | Optional repository injection in bbook.Engine | Maintains backward compatibility with tests while enabling real-time margin updates in production |

### Pending Todos

None yet.

### Blockers/Concerns

None yet.

## Session Continuity

Last session: 2026-01-16
Stopped at: Completed 06-03-PLAN.md (Real-Time Margin Calculation Engine)
Resume file: None

## Phase 1 Completion Summary

**Phase 1: Security & Configuration** ✅ Complete (2026-01-16)

All 3 plans executed successfully:
1. ✅ Environment Configuration & Secret Management (01-environment-secrets)
2. ✅ WebSocket Security & CORS Validation (02-websocket-cors)
3. ✅ Password Security Hardening (03-password-security)

**Success Criteria Verification:**
- ✅ No hardcoded credentials exist in codebase
- ✅ JWT tokens use cryptographically secure secret (44-byte)
- ✅ WebSocket connections validate origin against whitelist
- ✅ All passwords stored as bcrypt hashes (no plaintext fallback)
- ✅ Platform starts successfully using .env configuration

**Key Achievements:**
- Eliminated all hardcoded credentials (OANDA API keys, JWT secrets)
- Implemented production-grade CORS validation with wildcard support
- Enforced bcrypt-only password authentication
- Created comprehensive environment configuration system
- Added security logging and fail-safe behaviors

**Ready for:** Phase 3 - Testing Infrastructure

## Phase 2 Completion Summary

**Phase 2: Database Migration** ✅ Complete (2026-01-16)

All 4 plans executed successfully:
1. ✅ PostgreSQL Foundation & Schema (02-01)
2. ✅ Repository Pattern Implementation (02-02)
3. ✅ Trading Engine Database Integration (02-03)
4. ✅ Audit Trail & Compliance Logging (02-04)

**Success Criteria Verification:**
- ✅ Database schema created and migrated
- ✅ Account data loads from database (not JSON files)
- ✅ Position data persists to database
- ✅ Trade history queryable from database
- ✅ Platform restarts without data loss

**Key Achievements:**
- PostgreSQL database with 4 core tables (accounts, positions, orders, trades)
- Repository pattern with CRUD operations for all trading entities
- Connection pool singleton using pgx v5 for optimal performance
- Engine integrated with database via dependency injection
- Idempotent data migration from JSON to PostgreSQL
- Comprehensive audit trail using PostgreSQL triggers
- ACID compliance with REPEATABLE READ isolation for financial operations

**Verification:** All must-haves verified in codebase (02-VERIFICATION.md)

## Phase 4 Completion Summary

**Phase 4: Deployment & Operations** ✅ Complete (2026-01-16)

All 8 plans executed successfully:
1. ✅ Production Docker Images (04-01)
2. ✅ Docker Compose Development Environment (04-02)
3. ✅ GitHub Actions CI/CD Workflows (04-03)
4. ✅ Prometheus Metrics Collection (04-04)
5. ✅ Redis Caching Layer (04-05)
6. ✅ LP Manager Performance Optimization (04-06)
7. ✅ Database Backup Strategy (04-07)
8. ✅ Deployment Documentation (04-08)

**Success Criteria Verification:**
- ✅ Production-ready Docker images with security hardening
- ✅ Health check endpoints for orchestration
- ✅ CI/CD pipelines with monorepo path filtering
- ✅ Prometheus metrics for observability
- ✅ Redis caching for market data performance
- ✅ LP manager optimized from O(n) to O(1) lookups
- ✅ Automated database backups with disaster recovery capability
- ✅ Comprehensive deployment documentation (3,148 lines)

**Key Achievements:**
- Distroless Go backend image (2-5MB) with multi-stage builds
- Nginx-based React frontend with production optimizations
- Health endpoints (/health, /health/live, /health/ready) for K8s
- GitHub Actions CI/CD with path filtering and BuildKit caching
- Automated testing and Docker publishing to GHCR
- Prometheus metrics for trading operations and system performance
- Redis integration with TTL-based caching for OHLC and tick data
- Map-based LP configuration lookups for high-frequency operations
- Automated PostgreSQL backups with 6-hour schedule and dual retention strategy
- Disaster recovery capability with compressed backups and 30-day retention
- Complete deployment documentation: DOCKER.md, LOCAL_DEV.md, MONITORING.md, OPERATIONS.md (941 lines), CI_CD.md
- Operations runbook with troubleshooting, security checklist (28 items), and incident response procedures

**Ready for:** Phase 5 - Advanced Order Types
