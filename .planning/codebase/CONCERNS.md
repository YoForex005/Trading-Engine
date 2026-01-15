# Codebase Concerns

**Analysis Date:** 2026-01-15

## Security Issues

**CRITICAL: Hardcoded API Credentials**
- Issue: OANDA API key and account ID hardcoded in source code
- Files:
  - `backend/cmd/server/main.go` lines 23-24
- Risk: Credentials exposed in version control, potential unauthorized access
- Recommendation: Use environment variables, rotate credentials immediately

**CRITICAL: Weak JWT Secret**
- Issue: JWT secret key hardcoded as `"super_secret_dev_key_do_not_use_in_prod"`
- Files: `backend/auth/token.go` line 18
- Risk: Predictable secret allows JWT forgery
- Recommendation: Generate cryptographically random secret, store in environment variable

**HIGH: CORS Validation Disabled**
- Issue: WebSocket CORS check returns true for all origins
- Files: `backend/ws/hub.go` lines 15-18
- Risk: Enables CSRF attacks from any origin
- Recommendation: Implement origin whitelist, validate request origin

**HIGH: Plaintext Password Fallback**
- Issue: Legacy mode accepts plaintext passwords without enforcing bcrypt
- Files: `backend/auth/service.go` lines 81-84
- Risk: Downgrade attack possible, weak authentication
- Recommendation: Remove plaintext fallback, enforce bcrypt for all passwords

**MEDIUM: Manual CORS Implementation**
- Issue: CORS headers added manually in every handler (40+ locations)
- Files: `backend/internal/api/handlers/*.go`, `backend/bbook/api.go`
- Risk: Easy to forget, inconsistent application, security gaps
- Recommendation: Implement CORS middleware

## Error Handling Gaps

**Ignored I/O Errors**
- Issue: Multiple blank error assignments for `io.ReadAll()` calls
- Files:
  - `backend/oanda/client.go` lines 88, 136, 236, 285, 343, 432
- Risk: Silent failures reading API responses, hard to debug
- Recommendation: Log or return all errors

**Ignored Hashing Errors**
- Issue: bcrypt errors silently ignored
- Files:
  - `backend/auth/service.go` lines 31, 86
- Risk: Password hashing could fail silently
- Recommendation: Handle errors, fail fast on crypto failures

**LP Startup Failures Silent**
- Issue: LP adapter startup errors logged but not propagated
- Files: `backend/lpmanager/manager.go` lines 320-322
- Risk: System appears healthy when LP connections fail
- Recommendation: Track LP health, expose in admin dashboard

## Technical Debt

**Missing Backend Tests**
- Issue: 0 Go test files found
- Impact: Critical business logic untested (engine, LP manager, WebSocket hub)
- Priority: HIGH
- Files needing tests:
  - `backend/internal/core/engine.go` (861 lines)
  - `backend/lpmanager/manager.go`
  - `backend/ws/hub.go`
  - `backend/internal/core/ledger.go`
- Recommendation: Add unit tests for core business logic

**Oversized Components**
- Issue: Large components with mixed concerns
- Files:
  - `clients/desktop/src/components/TradingChart.tsx` (952 lines)
  - `clients/desktop/src/App.tsx` (589 lines)
  - `clients/desktop/src/indicators/core/IndicatorEngine.ts` (573 lines)
- Impact: Hard to maintain, test, and understand
- Recommendation: Extract smaller components, separate concerns

**Hardcoded Configuration**
- Issue: No environment variable support, hardcoded localhost URLs
- Files:
  - `clients/desktop/src/App.tsx` (multiple localhost:8080 URLs)
  - `clients/desktop/src/services/DataSyncService.ts` line 8
  - `clients/desktop/src/components/TradingChart.tsx` lines 116-118, 598
  - All frontend components with API calls
- Impact: Cannot deploy to different environments
- Recommendation: Centralize API configuration, use environment variables

**Excessive `any` Types**
- Issue: TypeScript `any` type used throughout, defeats type safety
- Files:
  - `clients/desktop/src/App.tsx` lines 69-70, 246, 331
  - `clients/desktop/src/components/TradingChart.tsx` lines 32, 62, 157, 229, 288, 494, 672, 819
  - `clients/desktop/src/indicators/core/IndicatorEngine.ts` line 375
- Impact: Type safety lost, runtime errors not caught at compile time
- Recommendation: Define proper types for all data structures

## Performance Concerns

**Large JSON File Loading**
- Issue: 85MB ticks.json file loaded at startup
- Files: `backend/data/ticks.json`
- Impact: Slow startup, high memory usage
- Recommendation: Implement database for tick storage, paginate queries

**Inefficient Data Structures**
- Issue: Linear searches through LP arrays
- Files: `backend/lpmanager/manager.go` lines 94-104, 127-135, 142-144
- Impact: O(n) performance for LP lookups
- Recommendation: Use map for O(1) LP lookup by ID

**Unoptimized Logging**
- Issue: RLock/RUnlock every 1000 ticks for logging
- Files: `backend/ws/hub.go` lines 95-101
- Impact: Lock contention on high-frequency updates
- Recommendation: Use atomic counters or rate-limited logging

## Missing Critical Features

**No Environment Configuration**
- Issue: No .env file support or .env.example
- Impact: Cannot configure for different environments
- Priority: HIGH
- Recommendation: Implement dotenv support, create .env.example

**No Database Layer**
- Issue: All state in memory, file-based persistence only
- Files: `backend/database/` exists but not used
- Impact: Data loss on restart, scalability limits
- Priority: MEDIUM
- Recommendation: Implement PostgreSQL/SQLite for persistent storage

**Incomplete Account Management**
- Issue: Hardcoded account ID = 1 throughout frontend
- Files:
  - `clients/desktop/src/components/TradingChart.tsx` line 116
  - `clients/desktop/src/components/AdvancedOrderPanel.tsx`
  - `clients/desktop/src/components/FloatingAccountPanel.tsx`
- Impact: Cannot support multiple accounts
- Recommendation: Pass accountId via props/context

## Test Coverage Gaps

**Backend: Zero Test Coverage**
- Priority: CRITICAL
- Impact: No safety net for refactoring, bugs undetected
- Recommendation: Start with core engine tests, LP manager tests

**Frontend: Partial Coverage**
- Good: Indicator storage, indicator engine
- Missing: TradingChart component, App component, most services
- Priority: HIGH
- Recommendation: Add React Testing Library tests for components

## Documentation Gaps

**No API Documentation**
- Issue: REST API not documented
- Impact: Hard for new developers, frontend integration unclear
- Recommendation: Add OpenAPI/Swagger spec or API documentation

**No Architecture Documentation**
- Issue: No high-level design docs (beyond this codebase analysis)
- Impact: Onboarding difficult, design decisions unclear
- Recommendation: Document architectural decisions, data flow

## Fragile Areas

**WebSocket Hub**
- Why fragile: Complex concurrency, multiple goroutines, broadcast channel
- Files: `backend/ws/hub.go`
- Common failures: Connection leaks, goroutine leaks, race conditions
- Safe modification: Add tests first, use race detector (`go test -race`)

**LP Manager**
- Why fragile: Dynamic adapter lifecycle, concurrent quote aggregation
- Files: `backend/lpmanager/manager.go`
- Common failures: Adapter startup failures, quote channel blocking
- Safe modification: Add integration tests, monitor goroutine count

**Trading Chart Component**
- Why fragile: 952 lines, manages chart state, drawings, indicators, positions
- Files: `clients/desktop/src/components/TradingChart.tsx`
- Common failures: State synchronization bugs, memory leaks
- Safe modification: Extract smaller components first, add prop validation

## Incomplete TODOs

**Backend:**
- `backend/cmd/server/main.go` line 504: "TODO: Implement account-specific WebSocket"
- `backend/api/server.go` line 388: "TODO: Implement OANDA trade modification"

**Frontend:**
- `clients/desktop/src/components/TradingChart.tsx` line 116: "TODO: Pass accountId to TradingChart"
- `clients/desktop/src/components/TradingChart.tsx` line 585: "TODO: Implement clear all"
- `clients/desktop/src/components/TradingChart.tsx` line 598: "TODO: Prop"

---

*Concerns audit: 2026-01-15*
*Update as issues are fixed or new ones discovered*
