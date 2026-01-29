# Codebase Concerns

**Analysis Date:** 2026-01-18

## Security Vulnerabilities

**Hardcoded API Credentials:**
- Issue: Production API keys exposed in source code
- Files: `backend/cmd/server/main.go` lines 23-24
- Impact: OANDA API key `977e1a77e25bac3a688011d6b0e845dd-8e3ab3a7682d9351af4c33be65e89b70` and account ID hardcoded in main.go
- Fix approach: Move to environment variables, add validation that credentials are not committed
- Security Risk: **CRITICAL** - Live trading credentials in version control

**Weak Default JWT Secret:**
- Issue: Fallback JWT secret key is predictable
- Files: `backend/auth/token.go` line 18
- Impact: If JWT_SECRET env var missing, uses `"super_secret_dev_key_do_not_use_in_prod"` allowing token forgery
- Trigger: Production deployment without JWT_SECRET set
- Fix approach: Require JWT_SECRET in production, panic if missing in non-dev mode

**Hardcoded Admin Password:**
- Issue: Admin password hash generated from literal "password" string
- Files: `backend/auth/service.go` lines 29-31
- Impact: Every instance has same admin password "password" - easily brute-forced
- Fix approach: Require admin password via secure config/environment variable on first startup

**FIX Protocol Credentials in Code:**
- Issue: FIX trading credentials hardcoded in test files
- Files: `backend/fix/test_yofx_443.go` line 31, `backend/fix/test_all_features.go` line 20, `backend/fix/test_connection.go` line 60
- Impact: Password `"Brand#143"` exposed in multiple test files, risk of accidental production use
- Fix approach: Move to environment variables or secure credential store

**Plaintext Password Fallback:**
- Issue: Legacy plaintext password support in login flow
- Files: `backend/auth/service.go` lines 84-88
- Impact: Accounts with plaintext passwords accepted, auto-upgraded but leaves window for plaintext storage
- Fix approach: Force bcrypt hashing on account creation, remove plaintext fallback

**CORS Wide Open:**
- Issue: WebSocket accepts all origins
- Files: `backend/ws/hub.go` line 17
- Impact: `CheckOrigin: func(r *http.Request) bool { return true }` allows any domain to connect
- Fix approach: Implement proper origin whitelist or validate against allowed domains

**No TLS Certificate Validation Configuration:**
- Issue: TLS config uses system defaults without pinning
- Files: `backend/fix/gateway.go` line 399
- Impact: Potential MITM attacks on FIX connections if CA compromised
- Fix approach: Add certificate pinning for known LP servers

## Performance Bottlenecks

**Unbounded Tick Data Growth:**
- Problem: 143 tick data files totaling 138MB tracked in Git LFS, growing continuously
- Files: `backend/data/ticks/` directory
- Cause: Daily rotation stores forever without cleanup by default
- Impact: Disk usage grows ~30-50MB per day across 130+ symbols, Git LFS bandwidth costs
- Improvement path: `backend/tickstore/daily_store.go` has cleanup (line 257) but maxDaysKeep must be set non-zero

**Synchronous File I/O in Hot Path:**
- Problem: Tick storage writes to disk during broadcast
- Files: `backend/ws/hub.go` lines 110-112, `backend/tickstore/service.go`
- Cause: StoreTick called on every market tick broadcast (could be 100+ ticks/second)
- Impact: File I/O can block market data distribution under high load
- Improvement path: Already has 30-second persist timer (daily_store.go:162), but direct calls bypass batching

**Global Tick Counter Without Reset:**
- Problem: Counter will overflow after 9.2 quintillion ticks
- Files: `backend/ws/hub.go` line 80
- Cause: `var tickCounter int64 = 0` increments forever
- Impact: Unrealistic but technically unbounded growth, logs every 1000 ticks
- Improvement path: Use atomic operations, reset daily, or remove after testing

**Large File Sizes:**
- Problem: Multiple files exceed 500 lines indicating high complexity
- Files:
  - `backend/fix/gateway.go` - 2527 lines
  - `backend/internal/core/engine.go` - 821 lines
  - `backend/bbook/api.go` - 746 lines
  - `backend/bbook/engine.go` - 739 lines
  - `backend/api/server.go` - 691 lines
- Cause: Monolithic file organization, all functionality in single files
- Impact: Difficult to navigate, test, and maintain; higher risk of merge conflicts
- Improvement path: Split by feature/concern (e.g., gateway.go → connection.go, messages.go, parser.go)

**Time.Sleep in Production Code:**
- Problem: Blocking sleep calls in LP adapters
- Files: `backend/lpmanager/adapters/binance.go` line 353, `backend/binance/client.go` line 193, `backend/cmd/server/main.go` lines 476, 502
- Cause: Reconnection delay and startup sequencing using sleep
- Impact: Blocks goroutines, delays recovery, prevents graceful shutdown
- Improvement path: Use context.WithTimeout, time.After with select, or exponential backoff with context

**No Connection Pooling:**
- Problem: HTTP clients created per-request in LP adapters
- Files: `backend/oanda/client.go`, `backend/binance/client.go`
- Cause: No shared http.Client with connection pooling
- Impact: TCP connection overhead on every API call, slow LP communication
- Improvement path: Create singleton http.Client with configured MaxIdleConns

## Concurrency Risks

**Potential Deadlock in LP Manager:**
- Issue: startLPLocked acquires lock while already holding lock
- Files: `backend/lpmanager/manager.go` lines 282-299
- Why fragile: `StartQuoteAggregation` holds lock (line 267), launches goroutine calling `startLPLocked` which tries to re-lock (line 294)
- Impact: Comment at line 283 acknowledges the issue: "ISSUE: We need separate public/private methods"
- Workaround: Currently releases lock before goroutine, but ToggleLP (line 163) calls startLPLocked while holding lock
- Fix approach: Refactor to separate lockable and non-lockable methods

**Race Condition in Hub Client Map:**
- Issue: Client map access pattern vulnerable to races
- Files: `backend/ws/hub.go` lines 148-198
- Why fragile: Register/unregister use Lock, broadcast uses RLock, but client send loops can fail without proper cleanup
- Impact: Client send channel closed while hub still references it (line 174 closes, but broadcast at line 191 may write)
- Test coverage: No concurrent access tests
- Fix approach: Add happens-before guarantee, test with -race flag

**FIX Gateway Sequence Number Races:**
- Issue: Sequence numbers read/written without consistent locking
- Files: `backend/fix/gateway.go` lines 78-82
- Why fragile: msgStore has RWMutex but OutSeqNum/InSeqNum accessed without always holding lock
- Impact: Sequence number mismatches cause FIX protocol violations, session disconnects
- Fix approach: Audit all seq number access, ensure atomic increments or mutex protection

**No Panic Recovery in Goroutines:**
- Problem: Background goroutines lack panic recovery
- Files: `backend/ws/hub.go` Run() method, `backend/lpmanager/manager.go` aggregateQuotes, `backend/tickstore/daily_store.go` persistPeriodically
- Cause: No defer/recover wrappers on long-running goroutines
- Impact: Single panic in hub, LP aggregator, or tick persister crashes entire server
- Fix approach: Wrap goroutine bodies with `defer func() { if r := recover(); r != nil { log error } }()`

## Data Integrity Issues

**No Persistence Layer:**
- Problem: All trading data exists only in memory
- Files: `backend/internal/core/engine.go` - Account, Position, Order, Trade all in maps
- Impact: Server restart loses all accounts, positions, orders, and trade history
- Blocks: Production deployment, regulatory compliance, disaster recovery
- Fix approach: Add database layer (PostgreSQL recommended), persist to disk on state changes

**No Transaction Safety:**
- Problem: Multi-step operations not atomic
- Files: `backend/internal/core/engine.go` ExecuteMarketOrder (lines 400+)
- Impact: Order execution updates account, creates position, creates trade - if any step fails mid-way, inconsistent state
- Fix approach: Implement transaction pattern, rollback mechanism, or event sourcing

**Tick Data Deduplication Only on Merge:**
- Problem: Real-time ticks not deduplicated
- Files: `backend/tickstore/daily_store.go` line 61
- Impact: Duplicate timestamps possible if LP sends same tick twice, inflates storage
- Fix approach: Add timestamp deduplication to StoreTick method

**Account State Not Persisted:**
- Problem: CreateAccount creates in-memory only
- Files: `backend/internal/core/engine.go`, `backend/bbook/engine.go`
- Impact: Account balances, positions reset on every restart
- Fix approach: Add database writes for account CRUD operations

## Incomplete Implementations

**TODO: Account-Specific WebSocket:**
- Issue: WebSocket broadcasts to all clients regardless of account
- Files: `backend/cmd/server/main.go` line 515
- Impact: Privacy leak - all connected clients see all account updates
- Fix approach: Add account filtering in hub, authenticate WebSocket connections

**TODO: OANDA Trade Modification:**
- Issue: LP trade modification not implemented
- Files: `backend/api/server.go` line 388
- Impact: Cannot modify open positions on OANDA LP
- Fix approach: Implement OrderReplaceRequest FIX message or OANDA REST API call

**Missing Error Handling in Daily Store:**
- Issue: JSON marshal/unmarshal errors ignored
- Files: `backend/tickstore/daily_store.go` lines 239, 363
- Impact: Silent data loss if tick data can't be serialized
- Fix approach: Log errors, retry with exponential backoff, alert on persistent failures

**No Graceful Shutdown:**
- Problem: Server stops immediately on interrupt
- Files: `backend/cmd/server/main.go`
- Impact: In-flight orders lost, WebSocket clients disconnected without cleanup, tick data not flushed
- Fix approach: Add signal handling, context cancellation, drain channels before exit

## Testing Gaps

**Zero Test Files:**
- What's not tested: Entire codebase
- Files: No `*_test.go` files found (find command returned 0)
- Risk: No verification that any functionality works, regressions undetected, refactoring unsafe
- Priority: **HIGH**
- Fix approach: Start with critical path tests (order execution, position management, authentication)

**No Integration Tests:**
- What's not tested: FIX gateway connectivity, LP adapter interactions, WebSocket message flow
- Risk: Cannot verify end-to-end flows without manual testing
- Fix approach: Add test harness for FIX messages, mock LP responses, WebSocket client simulator

**No Concurrency Tests:**
- What's not tested: Race conditions in hub, LP manager, engine
- Risk: Data races and deadlocks only manifest under production load
- Fix approach: Add tests with `go test -race`, concurrent client simulators

**No Load Tests:**
- What's not tested: Performance under high tick volume, many concurrent clients
- Risk: Unknown capacity limits, potential DoS from legitimate traffic
- Fix approach: Benchmark tests for tick ingestion, WebSocket broadcast, order execution throughput

## Configuration Management

**Mixed Config Sources:**
- Problem: Configuration scattered across hardcoded constants, JSON files, environment variables
- Files: `backend/cmd/server/main.go` lines 38-46 (BrokerConfig struct), `backend/data/lp_config.json`, env vars
- Impact: Difficult to understand what's configurable, risk of config drift between environments
- Fix approach: Centralize to single config file format, environment variable overrides

**No Config Validation:**
- Problem: Invalid configurations accepted silently
- Files: `backend/lpmanager/manager.go` LoadConfig
- Impact: Server starts with broken LP configs, fails at runtime
- Fix approach: Add validation on startup, fail fast if config invalid

**Environment Variable Fallbacks Unsafe:**
- Problem: getEnvOrDefault pattern hides missing required config
- Files: `backend/fix/gateway.go` line 337
- Impact: Production deployment may use development defaults unknowingly
- Fix approach: Separate required vs. optional config, panic on missing required vars in prod

## Operational Concerns

**No Metrics/Monitoring:**
- Problem: No Prometheus/StatsD metrics exported
- Files: Entire codebase
- Impact: Cannot monitor system health, tick rates, order latency, error rates
- Fix approach: Add metrics library, expose /metrics endpoint, instrument critical paths

**Log Level Hardcoded:**
- Problem: Debug logs always on or always off
- Files: `backend/cmd/server/main.go` line 370 checks DEBUG env var, but inconsistently applied
- Impact: Cannot adjust verbosity without code changes
- Fix approach: Structured logging with configurable levels (info, warn, error, debug)

**Single log.Fatal in Production:**
- Problem: Main goroutine can crash entire server
- Files: `backend/cmd/server/main.go` line 546
- Trigger: HTTP server startup failure
- Impact: No recovery, requires manual restart
- Fix approach: Supervisor process, health check endpoint, auto-restart on failure

**No Circuit Breakers:**
- Problem: LP failures cause indefinite reconnection attempts
- Files: `backend/lpmanager/adapters/binance.go`, `backend/lpmanager/adapters/oanda.go`
- Impact: Server resources exhausted retrying dead LPs
- Fix approach: Implement circuit breaker pattern, exponential backoff with max retries

**FIX Sequence Numbers Persisted to Filesystem:**
- Problem: Sequence numbers stored in `./fixstore` directory
- Files: `backend/fix/gateway.go` line 51, storeDir usage throughout
- Impact: If filesystem fails or directory deleted, sequence number reset causes FIX protocol errors
- Fix approach: Use database for sequence number persistence, atomic file writes, backup/restore

## Dependency Risks

**Minimal Dependencies (Good but Fragile):**
- Issue: Only 4 dependencies in go.mod
- Files: `backend/go.mod`
- Impact: Low supply chain risk but missing critical libraries (no DB driver, no metrics, no structured logging)
- Migration plan: Add production-ready libraries (zap/logrus for logging, sqlx for DB, prometheus for metrics)

**No Dependency Pinning:**
- Risk: go.mod uses version ranges, potential for breaking changes
- Impact: Build reproducibility not guaranteed
- Migration plan: Use `go mod vendor` or strict version pinning

**Gorilla WebSocket End of Life:**
- Risk: gorilla/websocket in maintenance mode, will be archived
- Impact: Security patches may stop
- Migration plan: Evaluate migration to nhooyr.io/websocket or golang.org/x/net/websocket

## Scalability Limits

**In-Memory State Bounded by RAM:**
- Current capacity: Limited to RAM size for accounts, positions, ticks
- Limit: ~10K accounts × 100 positions = 1M positions × 200 bytes = 200MB minimum
- Files: `backend/internal/core/engine.go`
- Scaling path: Add database persistence, shard by account ID, separate read/write paths

**Single-Process Architecture:**
- Problem: Cannot horizontally scale
- Impact: One server handles all accounts, all LPs, all WebSocket clients
- Scaling path: Microservices architecture - separate market data, execution, risk services

**No Rate Limiting:**
- Problem: API endpoints unprotected from abuse
- Files: `backend/api/server.go`
- Impact: Single client can overwhelm server with requests
- Scaling path: Add rate limiting middleware (golang.org/x/time/rate)

**WebSocket Broadcast to All Clients:**
- Problem: Hub broadcasts every tick to every client
- Files: `backend/ws/hub.go` lines 179-197
- Limit: Network bandwidth saturates at ~10K clients × 100 ticks/sec × 200 bytes = 200MB/s
- Scaling path: Client-side symbol subscriptions, only broadcast relevant ticks per client

## Code Quality Issues

**Inconsistent Error Handling:**
- Problem: Some errors logged and returned, some swallowed, some panic
- Files: `backend/lpmanager/adapters/binance.go` vs. `backend/fix/test_all_features.go` (log.Fatalf)
- Impact: Unpredictable failure modes, difficult to debug
- Fix approach: Establish error handling conventions, wrap errors with context

**Magic Numbers:**
- Problem: Unmarked constants throughout codebase
- Examples: 100000 (contract size), 2048 (buffer sizes), 10 (timeouts)
- Files: `backend/risk/calculator.go` line 168, `backend/ws/hub.go` line 57
- Impact: Difficult to tune, unclear why specific values chosen
- Fix approach: Extract to named constants with comments explaining rationale

**Dead Code:**
- Problem: Unused services created but not wired
- Files: `backend/api/server.go` creates omsService, riskEngine, smartRouter, fixGateway but never calls them
- Impact: Wasted memory, misleading about system capabilities
- Fix approach: Remove unused components or document intended future use

**No Input Validation:**
- Problem: API handlers accept arbitrary input
- Files: `backend/api/server.go`, `backend/bbook/api.go`
- Impact: Negative volumes, zero prices, invalid symbols accepted
- Fix approach: Add validation layer, return 400 Bad Request for invalid input

---

*Concerns audit: 2026-01-18*
