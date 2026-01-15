# Phase 6, Plan 1: Risk Management Schema & Repositories - Summary

**Completed:** 2026-01-16
**Plan:** `.planning/phases/06-risk-management/06-01-PLAN.md`
**Status:** ✅ Complete

## Objective

Create PostgreSQL schema and repositories for risk management data structures, providing the foundation for margin monitoring, risk limits, and regulatory compliance (ESMA leverage limits).

## What Was Built

### 1. Database Schema Migration (000004_risk_management_schema)

Created comprehensive risk management schema with three core tables:

**Files Created:**
- `backend/db/migrations/000004_risk_management_schema.up.sql`
- `backend/db/migrations/000004_risk_management_schema.down.sql`

**Tables:**

1. **margin_state** - Real-time margin tracking per account
   - account_id (PK, FK to accounts)
   - equity, used_margin (DECIMAL(20,8))
   - free_margin (computed column: equity - used_margin)
   - margin_level (DECIMAL(10,2))
   - margin_call_triggered, stop_out_triggered (BOOLEAN)
   - last_updated (TIMESTAMPTZ)

2. **risk_limits** - Configurable risk limits per account or group
   - id (PK), account_id (nullable FK), account_group (nullable)
   - max_leverage (DECIMAL(5,2), default 30.00)
   - max_open_positions (INT, default 50)
   - max_position_size_lots, daily_loss_limit, max_drawdown_pct (optional DECIMAL)
   - margin_call_level (default 100.00), stop_out_level (default 50.00)
   - Constraint: margin_call_level > stop_out_level

3. **symbol_margin_config** - ESMA regulatory leverage limits by symbol
   - symbol (PK)
   - asset_class (forex_major, forex_minor, stock, crypto, commodity, index)
   - max_leverage (DECIMAL(5,2) - ESMA limits: 30:1, 20:1, 5:1, 2:1)
   - margin_percentage (DECIMAL(5,4) - computed as 1/leverage * 100)
   - contract_size, tick_size, tick_value (DECIMAL for precision)

**Key Design Decisions:**
- All financial values use DECIMAL (not float64) to prevent precision errors
- Indexes on frequently queried columns (account_id, account_group, asset_class, last_updated)
- CHECK constraints enforce business rules (positive values, margin_call > stop_out)
- UNIQUE constraint with NULLS NOT DISTINCT for risk_limits (account_id, account_group)
- Generated column for free_margin ensures consistency

### 2. Repository Layer Implementation

Created three repositories following existing pattern from Phase 2:

**Files Created:**
- `backend/internal/database/repository/margin_state.go`
- `backend/internal/database/repository/risk_limit.go`
- `backend/internal/database/repository/symbol_margin_config.go`

**MarginStateRepository:**
- `GetByAccountID(ctx, accountID)` - Retrieve margin state
- `Upsert(ctx, state)` - Insert/update for real-time updates
- `GetByAccountIDWithLock(ctx, tx, accountID)` - SELECT FOR UPDATE for concurrent safety
- `UpsertWithTx(ctx, tx, state)` - Transaction-aware upsert

**RiskLimitRepository:**
- `GetByAccountID(ctx, accountID)` - Account-specific limits
- `GetByAccountGroup(ctx, group)` - Group-level limits
- `Create(ctx, limit)` - Insert new limit
- `Update(ctx, limit)` - Modify existing limit
- `GetAll(ctx)` - Admin/reporting query

**SymbolMarginConfigRepository:**
- `GetBySymbol(ctx, symbol)` - Symbol-specific config
- `GetAll(ctx)` - All symbol configurations
- `Upsert(ctx, config)` - Insert/update config
- `SeedESMADefaults(ctx)` - Initialize with ESMA-compliant leverage limits
- `GetByAssetClass(ctx, assetClass)` - Filter by asset class

**CRITICAL Implementation Details:**
- All DECIMAL database columns stored as `string` in Go structs (no float64)
- Nullable fields use pointers (*int64, *string)
- Parameterized queries ($1, $2) prevent SQL injection
- Constructor pattern: `New*Repository(pool *pgxpool.Pool)`
- Context-aware methods for cancellation support
- Transaction support for concurrent-safe operations

### 3. ESMA Default Configuration

`SeedESMADefaults()` provides regulatory-compliant starting data:

**Leverage Limits by Asset Class:**
- **Forex Major Pairs** (30:1): EURUSD, GBPUSD, USDJPY, USDCHF, AUDUSD, USDCAD, NZDUSD
- **Forex Minor Pairs** (20:1): EURGBP, EURJPY, GBPJPY, EURCHF
- **Cryptocurrencies** (2:1): BTCUSD, ETHUSD, XRPUSD, SOLUSD, BNBUSD
- **Indices** (20:1): US30, US500, NAS100
- **Commodities** (20:1): XAUUSD, XAGUSD, WTIUSD

Each symbol includes:
- Contract size, tick size, tick value
- Margin percentage (1/leverage * 100)
- Asset class classification

## Verification

✅ **All Success Criteria Met:**

1. Migration files created with correct naming (000004_*)
2. All 3 repository files compile without errors
   ```bash
   go build ./internal/database/repository
   # Success - no errors
   ```
3. No float64 types for financial data (verified via grep)
   ```bash
   grep "float64" margin_state.go risk_limit.go symbol_margin_config.go
   # No matches - all DECIMAL fields as strings
   ```
4. Foreign keys reference accounts table correctly (ON DELETE CASCADE)
5. CHECK constraints enforce business rules:
   - Positive equity/margin values
   - margin_call_level > stop_out_level
   - Valid asset_class enum values
   - Positive leverage/size limits

## Must-Haves Verification

✅ **All must-haves satisfied:**

**Truths:**
- ✅ Margin state persists per account with equity, used_margin, margin_level
- ✅ Risk limits configurable per account or account group
- ✅ Symbol margin configuration supports ESMA leverage limits (30:1, 20:1, 5:1, 2:1)

**Artifacts:**
- ✅ `backend/db/migrations/000004_risk_management_schema.up.sql` - Contains "CREATE TABLE margin_state"
- ✅ `backend/internal/database/repository/margin_state.go` - Exports MarginStateRepository
- ✅ `backend/internal/database/repository/risk_limit.go` - Exports RiskLimitRepository
- ✅ `backend/internal/database/repository/symbol_margin_config.go` - Exports SymbolMarginConfigRepository

**Key Links:**
- ✅ margin_state.account_id → accounts.id (FOREIGN KEY with ON DELETE CASCADE)
- ✅ risk_limits.account_id → accounts.id (nullable FK for group limits)
- ✅ All repositories use pgxpool.Pool via dependency injection (New*Repository pattern)

## Technical Decisions

1. **Migration Numbering:** Renamed to 000004 (000002 and 000003 already exist for audit trail and advanced order types)

2. **DECIMAL vs Float64:** All financial values stored as DECIMAL in database, represented as strings in Go. Conversion to decimal.Decimal happens in business logic layer (prevents precision errors that have caused production incidents - LSE halt, €12M German bank fine)

3. **Computed Column for free_margin:** PostgreSQL GENERATED ALWAYS AS ensures free_margin always equals equity - used_margin without application-level calculation risk

4. **UNIQUE NULLS NOT DISTINCT:** PostgreSQL 15+ feature allows risk_limits to have either account_id OR account_group set, but not both, preventing duplicate entries

5. **Transaction-Aware Methods:** Both Upsert and UpsertWithTx provided - regular Upsert for standalone updates, UpsertWithTx for multi-operation transactions (needed for margin calculation in future plans)

## Files Modified/Created

**Created (5 files):**
- `backend/db/migrations/000004_risk_management_schema.up.sql` (2,831 bytes)
- `backend/db/migrations/000004_risk_management_schema.down.sql` (153 bytes)
- `backend/internal/database/repository/margin_state.go` (5,847 bytes)
- `backend/internal/database/repository/risk_limit.go` (6,912 bytes)
- `backend/internal/database/repository/symbol_margin_config.go` (9,120 bytes)

**Total:** 5 files, ~24.8 KB of new code

## Next Steps

Plan 06-02 will implement:
- Real-time margin calculation engine
- Position margin calculation using symbol configs
- Margin monitoring with WebSocket updates
- Integration with existing Engine for position updates

## Key Learnings

1. **Migration Conflicts:** Always check existing migrations before creating new ones. This project already had 000002 (audit trail) and 000003 (advanced order types)

2. **PostgreSQL Features:** Generated columns (STORED) are cleaner than triggers for computed values. NULLS NOT DISTINCT simplifies nullable unique constraints.

3. **Repository Pattern Consistency:** Following Phase 2 pattern (pgxpool.Pool, context-aware methods, parameterized queries) ensures codebase maintainability

4. **Regulatory Compliance:** ESMA leverage limits are not suggestions - they're legal requirements for EU/UK/AU brokers. Seed data provides compliance-ready starting point

5. **Decimal Precision:** Storing as strings in Go and using DECIMAL in PostgreSQL is the ONLY acceptable approach for financial data. Float64 causes rounding errors that accumulate over millions of transactions.

## Dependencies

**Upstream (Required by this plan):**
- Phase 2 (Database Migration) - PostgreSQL schema with accounts table
- pgx v5 connection pool

**Downstream (Plans that depend on this):**
- Plan 06-02 - Margin calculation engine (uses all 3 repositories)
- Plan 06-03 - Risk monitoring service (uses MarginStateRepository)
- Plan 06-04 - Stop-out liquidation engine (uses all repositories)

---

**Duration:** ~35 minutes
**Complexity:** Medium (database schema + 3 repositories)
**Confidence:** HIGH - All verification passed, follows established patterns
