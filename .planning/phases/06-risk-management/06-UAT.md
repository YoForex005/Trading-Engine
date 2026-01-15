---
status: in_progress
phase: 06-risk-management
source: 06-01-SUMMARY.md, 06-02-SUMMARY.md, 06-03-SUMMARY.md, 06-04-SUMMARY.md, 06-05-SUMMARY.md, 06-06-SUMMARY.md, 06-07-SUMMARY.md
started: 2026-01-16
updated: 2026-01-16
---

## Current Test

### 1. Database Schema - Margin State Table
expected: margin_state table exists with columns: account_id, equity, used_margin, free_margin, margin_level, margin_call_triggered, stop_out_triggered, last_updated. Free margin should be a computed column (equity - used_margin).

## Tests

### 1. Database Schema - Margin State Table
expected: margin_state table exists with columns: account_id, equity, used_margin, free_margin, margin_level, margin_call_triggered, stop_out_triggered, last_updated. Free margin should be a computed column (equity - used_margin).

### 2. Database Schema - Risk Limits Table
expected: risk_limits table exists with configurable limits including max_leverage, max_open_positions, max_position_size_lots, daily_loss_limit, max_drawdown_pct, margin_call_level, stop_out_level.

### 3. Database Schema - Symbol Margin Config Table
expected: symbol_margin_config table exists with ESMA leverage limits by asset class (30:1 forex major, 20:1 forex minor/indices/commodities, 5:1 stocks, 2:1 crypto).

### 4. Database Schema - Daily Statistics Table
expected: daily_account_stats table exists tracking daily P&L, high-water mark, drawdown percentage, trade counts, and breach flags with composite key (account_id, stat_date).

### 5. Decimal Precision - No Float64 in Financial Calculations
expected: All financial calculations use govalues/decimal library. Check margin.go, validation.go, stopout.go, drawdown.go contain zero float64 references for financial values.

### 6. Decimal Utilities - Conversion Functions
expected: backend/internal/decimal/convert.go exports MustParse, Parse, ToString, Zero, Add, Sub, Mul, Div functions for decimal operations.

### 7. Real-Time Margin Calculation - Position Margin Formula
expected: CalculatePositionMargin uses standard forex formula: (volume × contract_size × open_price) ÷ leverage, implemented with decimal arithmetic.

### 8. Real-Time Margin Calculation - Margin Level Formula
expected: CalculateMarginLevel implements formula: (equity ÷ used_margin) × 100, returns 99999% when no positions open (edge case).

### 9. Real-Time Margin Calculation - Event-Driven Updates
expected: UpdateMarginState called after ExecuteMarketOrder (line ~492 in engine.go) and after ClosePosition (line ~586 in engine.go).

### 10. Real-Time Margin Calculation - Database Persistence
expected: Margin state persisted to database via MarginStateRepository.Upsert on every position change.

### 11. Pre-Trade Validation - Insufficient Margin Rejection
expected: ExecuteMarketOrder validates margin requirement before execution. Orders rejected with descriptive error if projected margin level would drop below margin call threshold.

### 12. Pre-Trade Validation - Position Count Limit
expected: ValidatePositionLimits enforces max_open_positions limit (default 50). Orders rejected if limit exceeded.

### 13. Pre-Trade Validation - Position Size Limit
expected: ValidatePositionLimits enforces max_position_size_lots limit if configured. Orders rejected if position size too large.

### 14. Pre-Trade Validation - Leverage Limit
expected: ValidateLeverage enforces ESMA regulatory limits per symbol (30:1, 20:1, 5:1, 2:1) and account-specific limits. Orders rejected if leverage too high.

### 15. Pre-Trade Validation - Symbol Exposure Limit
expected: ValidateSymbolExposure prevents concentration risk by limiting exposure per symbol to 40% of equity by default.

### 16. Pre-Trade Validation - Total Exposure Limit
expected: ValidateTotalExposure prevents over-leveraging by limiting total exposure across all positions to 300% of equity by default.

### 17. Automatic Stop-Out - Margin Threshold Detection
expected: UpdateMarginState detects when margin level drops to or below 50% (default stop-out level) and triggers ExecuteStopOut.

### 18. Automatic Stop-Out - Position Liquidation Order
expected: SelectPositionsForLiquidation sorts positions by unrealized P&L ascending (most losing positions first) for liquidation.

### 19. Automatic Stop-Out - Incremental Closure
expected: ExecuteStopOut closes positions one by one, recalculating margin after each closure, stopping when margin level recovers above threshold.

### 20. Automatic Stop-Out - Mutex Handling
expected: UpdateMarginState unlocks mutex before calling ExecuteStopOut (line ~242 in margin.go) to prevent deadlock, then re-acquires lock.

### 21. Automatic Stop-Out - Audit Logging
expected: Stop-out events logged with "[STOP OUT TRIGGERED]", "[STOP OUT COMPLETE]", and "[STOP OUT WARNING]" messages for compliance.

### 22. Daily Loss Limit - Pre-Trade Check
expected: CheckDailyLossLimit called in ExecuteMarketOrder before order execution. Orders rejected if daily loss limit already breached.

### 23. Daily Loss Limit - Account Disablement
expected: UpdateDailyStats auto-disables account when daily loss limit breached, setting disabled_at timestamp and daily_loss_limit_breached flag.

### 24. Drawdown Protection - High-Water Mark Tracking
expected: GetHistoricalHighWaterMark retrieves highest balance ever reached for drawdown calculation. UpdateDailyStats updates high-water mark when balance increases.

### 25. Drawdown Protection - Drawdown Percentage Calculation
expected: CheckDrawdownLimit calculates drawdown as (high_water_mark - balance) / high_water_mark × 100, enforces max_drawdown_pct limit.

### 26. Drawdown Protection - Account Disablement
expected: UpdateDailyStats auto-disables account when drawdown limit breached, setting disabled_at timestamp and max_drawdown_breached flag.

### 27. Repository Integration - Margin State Repository
expected: MarginStateRepository exports GetByAccountID, Upsert, GetByAccountIDWithLock, UpsertWithTx methods. All DECIMAL columns stored as strings.

### 28. Repository Integration - Risk Limit Repository
expected: RiskLimitRepository exports GetByAccountID, GetByAccountGroup, Create, Update, GetAll methods. Supports nullable account_id for group limits.

### 29. Repository Integration - Symbol Margin Config Repository
expected: SymbolMarginConfigRepository exports GetBySymbol, GetAll, Upsert, SeedESMADefaults, GetByAssetClass methods.

### 30. Repository Integration - Daily Stats Repository
expected: DailyStatsRepository exports GetByAccountAndDate, Upsert, GetHistoricalHighWaterMark methods.

### 31. ESMA Compliance - Forex Major Leverage
expected: ESMA seed data includes 30:1 leverage for major pairs (EURUSD, GBPUSD, USDJPY, USDCHF, AUDUSD, USDCAD, NZDUSD).

### 32. ESMA Compliance - Crypto Leverage
expected: ESMA seed data includes 2:1 leverage for cryptocurrencies (BTCUSD, ETHUSD, XRPUSD, SOLUSD, BNBUSD).

### 33. ESMA Compliance - Margin Call Threshold
expected: Default margin_call_level is 100% (margin call triggered when margin level drops to 100%).

### 34. ESMA Compliance - Stop-Out Threshold
expected: Default stop_out_level is 50% (positions liquidated when margin level drops to 50%).

### 35. Migration Files - Risk Management Schema
expected: backend/db/migrations/000004_risk_management_schema.up.sql creates margin_state, risk_limits, symbol_margin_config tables.

### 36. Migration Files - Daily Stats Schema
expected: backend/db/migrations/000005_daily_stats_schema.up.sql creates daily_account_stats table with composite primary key.

### 37. Migration Files - Rollback Support
expected: Down migration files (000004_*.down.sql, 000005_*.down.sql) drop tables in reverse order.

### 38. Engine Constructor - Repository Injection
expected: NewEngineWithRepos constructor accepts optional repositories (marginStateRepo, riskLimitRepo, symbolMarginConfigRepo, dailyStatsRepo). NewEngine passes nil for backward compatibility.

### 39. Validation Error Messages - Descriptive Feedback
expected: Validation errors include actual values in error messages (e.g., "margin level would drop to 45.23% (below 100% threshold)").

### 40. Compilation - No Float64 in Risk Code
expected: grep "float64" backend/bbook/margin.go validation.go stopout.go drawdown.go returns zero matches for financial values.

## Summary

total: 40
passed: 0
issues: 0
pending: 40
skipped: 0

## Issues for /gsd:plan-fix

[none yet - testing in progress]
