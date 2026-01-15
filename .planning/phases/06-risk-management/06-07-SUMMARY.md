# Plan 06-07 Summary: Daily Loss Limits and Drawdown Protection

**Phase:** 06-risk-management
**Plan:** 07
**Status:** ✅ Complete
**Completed:** 2026-01-16

## Objective

Implement daily loss limits and maximum drawdown protection to prevent catastrophic account losses through single-day blowouts or prolonged losing streaks.

## What Was Built

### 1. Daily Statistics Database Schema

**Files Created:**
- `/backend/db/migrations/000005_daily_stats_schema.up.sql` (1,677 bytes)
- `/backend/db/migrations/000005_daily_stats_schema.down.sql` (192 bytes)
- `/backend/internal/database/repository/daily_stats.go` (3,731 bytes)

**Database Schema:**
- `daily_account_stats` table with composite primary key (account_id, stat_date)
- Tracks daily P&L (realized, unrealized, total)
- High-water mark and drawdown percentage tracking
- Trading statistics (trades opened/closed, winning/losing)
- Limit breach flags and account disablement timestamp
- Indexes on `stat_date` and breach conditions

**Repository Features:**
- `GetByAccountAndDate()` - Retrieve stats for specific account and date
- `Upsert()` - Insert or update daily statistics
- `GetHistoricalHighWaterMark()` - Get highest balance ever reached
- All DECIMAL columns stored as strings (no float64 precision issues)

### 2. Daily Loss and Drawdown Monitoring

**File Created:**
- `/backend/bbook/drawdown.go` (227 lines)

**Functions Implemented:**

1. **CheckDailyLossLimit()**
   - Validates if current daily loss exceeds configured limit
   - Calculates daily P&L from starting balance
   - Returns error if limit breached
   - Checks if account already disabled for the day

2. **CheckDrawdownLimit()**
   - Validates if drawdown from high-water mark exceeds limit
   - Calculates drawdown percentage: `(high_water_mark - balance) / high_water_mark * 100`
   - Updates high-water mark if balance increased
   - Returns error if limit breached

3. **UpdateDailyStats()**
   - Updates or creates daily statistics after position close
   - Tracks realized P&L, unrealized P&L, total P&L
   - Updates high-water mark and drawdown calculations
   - Increments trade counts (winning/losing)
   - Checks limits and auto-disables account on breach
   - Logs breach events for compliance audit

**Key Features:**
- All decimal arithmetic (zero float64 usage)
- Graceful handling of first trade of the day
- Auto-disablement on limit breach with timestamp
- Separate breach tracking for daily loss vs drawdown

### 3. Engine Integration

**File Modified:**
- `/backend/bbook/engine.go`

**Changes Made:**

1. **Engine Struct:**
   - Added `dailyStatsRepo *repository.DailyStatsRepository` field

2. **NewEngineWithRepos() Constructor:**
   - Added `dailyStatsRepo` parameter
   - Updated `NewEngine()` to pass `nil` for backward compatibility

3. **ExecuteMarketOrder() - Pre-trade Validation:**
   - Added daily loss limit check BEFORE order execution
   - Added drawdown limit check BEFORE order execution
   - Returns error if either limit breached
   - Only runs if repositories initialized

4. **ClosePosition() - Post-trade Statistics:**
   - Added daily statistics update AFTER position close
   - Passes realized P&L, current balance, trade outcome
   - Logs failures but doesn't block position closure
   - Updates breach flags if limits crossed

## Verification Results

All verification criteria met:

- ✅ Migration files exist (`000005_*.sql`)
- ✅ DailyStatsRepository compiles without errors
- ✅ drawdown.go compiles without errors
- ✅ drawdown.go has ZERO float64 usage (227 lines, all decimal)
- ✅ CheckDailyLossLimit called in ExecuteMarketOrder
- ✅ UpdateDailyStats called in ClosePosition
- ✅ Account disabled on limit breach (via UpdateDailyStats)

**Key Exports Verified:**
```go
func CheckDailyLossLimit(...)
func CheckDrawdownLimit(...)
func UpdateDailyStats(...)
```

**Database Schema Verified:**
- `CREATE TABLE daily_account_stats` present
- Composite primary key on (account_id, stat_date)
- All required columns present (P&L, drawdown, breach flags)

## Must-Haves Status

### Truths ✅
- ✅ Daily loss limit enforced per account (configurable USD amount)
- ✅ Drawdown percentage tracked from account high-water mark
- ✅ Account auto-disabled when limits breached (requires manual re-enable)

### Artifacts ✅
- ✅ `backend/bbook/drawdown.go` (227 lines, exports CheckDailyLossLimit, CheckDrawdownLimit, UpdateDailyStats)
- ✅ `backend/internal/database/repository/daily_stats.go` (exports DailyStatsRepository)
- ✅ `backend/db/migrations/000005_daily_stats_schema.up.sql` (contains CREATE TABLE daily_account_stats)

### Key Links ✅
- ✅ Engine.ExecuteMarketOrder → CheckDailyLossLimit (pre-trade validation)
- ✅ Engine.ClosePosition → UpdateDailyStats (daily P&L tracking after position close)
- ✅ UpdateDailyStats → DailyStatsRepository.Upsert (persistence of daily statistics)

## Technical Decisions

| Decision | Rationale |
|----------|-----------|
| Daily stats keyed by (account_id, stat_date) | Single row per account per day for efficient tracking |
| High-water mark tracking across all time | Enables accurate drawdown calculation from peak equity |
| Check limits BEFORE order execution | Prevents trading when account already disabled for the day |
| Update stats AFTER position close | Ensures P&L is realized before updating daily statistics |
| Auto-disable on breach with timestamp | Requires manual re-enable for safety, provides audit trail |
| Separate breach flags (daily loss vs drawdown) | Different recovery procedures for each limit type |
| Migration number 000005 | Follows sequential migration numbering convention |

## Risk Management Benefits

1. **Single-Day Protection:**
   - Prevents account blowouts from one bad trading day
   - Configurable per-account daily loss limits
   - Automatic circuit breaker functionality

2. **Long-Term Drawdown Protection:**
   - Tracks performance from historical peak (high-water mark)
   - Prevents prolonged losing streaks from wiping accounts
   - Aligns with industry best practices (prop firms, hedge funds)

3. **Regulatory Compliance:**
   - Provides investor protection mechanisms
   - Audit trail of breach events and disablements
   - Supports negative balance protection requirements

4. **Operational Safety:**
   - Manual re-enable requirement prevents accidental resumption
   - Separate tracking for different breach types
   - Logged events for compliance review

## Integration Points

**Upstream Dependencies:**
- Risk limits configuration (Plan 06-01) - provides daily_loss_limit and max_drawdown_pct
- Decimal precision utilities (Plan 06-02) - all financial calculations use decimal
- Engine structure (existing) - ExecuteMarketOrder and ClosePosition methods

**Downstream Impact:**
- Server initialization must pass DailyStatsRepository to Engine constructor
- Database migration 000005 must run before production deployment
- Manual account re-enable process needed after breach

## Files Modified

1. `/backend/db/migrations/000005_daily_stats_schema.up.sql` (new)
2. `/backend/db/migrations/000005_daily_stats_schema.down.sql` (new)
3. `/backend/internal/database/repository/daily_stats.go` (new)
4. `/backend/bbook/drawdown.go` (new, 227 lines)
5. `/backend/bbook/engine.go` (modified - added dailyStatsRepo field and limit checks)

## Success Criteria Met

- ✅ All tasks completed
- ✅ Daily loss limit enforced per account
- ✅ Drawdown percentage tracked from high-water mark
- ✅ Account auto-disabled when limits breached
- ✅ Daily statistics persisted to database
- ✅ Decimal arithmetic throughout (no float64)
- ✅ All code compiles without errors
- ✅ All must-have artifacts present and verified

## Next Steps

1. Run migration 000005 on database: `migrate up`
2. Update server initialization to pass DailyStatsRepository to Engine
3. Configure daily loss limits and drawdown limits per account in risk_limits table
4. Implement admin UI for manual account re-enable after breach
5. Add monitoring/alerts for daily loss and drawdown breaches

---

**Plan Status:** ✅ COMPLETE
**Phase 6 Progress:** 7/7 plans complete (100%)
