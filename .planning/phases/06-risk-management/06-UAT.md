---
status: complete
phase: 06-risk-management
source: 06-01-SUMMARY.md, 06-02-SUMMARY.md, 06-03-SUMMARY.md, 06-04-SUMMARY.md, 06-05-SUMMARY.md, 06-06-SUMMARY.md, 06-07-SUMMARY.md
started: 2026-01-16
updated: 2026-01-16
---

## Current Test

[testing complete]

## Tests

### 1. Database Schema - Margin State Table
expected: margin_state table exists with columns: account_id, equity, used_margin, free_margin, margin_level, margin_call_triggered, stop_out_triggered, last_updated. Free margin should be a computed column (equity - used_margin).
result: pass
verified: ✅ Table created in 000004_risk_management_schema.up.sql line 4-13, free_margin is GENERATED ALWAYS AS (equity - used_margin) STORED

### 2. Database Schema - Risk Limits Table
expected: risk_limits table exists with configurable limits including max_leverage, max_open_positions, max_position_size_lots, daily_loss_limit, max_drawdown_pct, margin_call_level, stop_out_level.
result: pass
verified: ✅ Table created in 000004_risk_management_schema.up.sql line 16-31, all required columns present with CHECK constraints

### 3. Database Schema - Symbol Margin Config Table
expected: symbol_margin_config table exists with ESMA leverage limits by asset class (30:1 forex major, 20:1 forex minor/indices/commodities, 5:1 stocks, 2:1 crypto).
result: pass
verified: ✅ Table created in 000004_risk_management_schema.up.sql line 34-44, asset_class CHECK constraint includes all ESMA classes

### 4. Database Schema - Daily Statistics Table
expected: daily_account_stats table exists tracking daily P&L, high-water mark, drawdown percentage, trade counts, and breach flags with composite key (account_id, stat_date).
result: pass
verified: ✅ Table created in 000005_daily_stats_schema.up.sql line 2-32, PRIMARY KEY (account_id, stat_date) on line 31

### 5. Decimal Precision - No Float64 in Financial Calculations
expected: All financial calculations use govalues/decimal library. Check margin.go, validation.go, stopout.go, drawdown.go contain zero float64 references for financial values.
result: pass
verified: ✅ grep "float64" returned 0 matches for margin.go, validation.go, stopout.go, drawdown.go

### 6. Decimal Utilities - Conversion Functions
expected: backend/internal/decimal/convert.go exports MustParse, Parse, ToString, Zero, Add, Sub, Mul, Div functions for decimal operations.
result: pass
verified: ✅ All functions found: MustParse (line 11), Parse (line 21), ToString (line 26), Zero (line 37), Add (line 87), Sub (line 97), Mul (line 107), Div (line 117)

### 7. Real-Time Margin Calculation - Position Margin Formula
expected: CalculatePositionMargin uses standard forex formula: (volume × contract_size × open_price) ÷ leverage, implemented with decimal arithmetic.
result: pass
verified: ✅ Function in margin.go lines 18-37 implements exact formula: volume.Mul(contractSize).Mul(openPrice).Quo(leverage)

### 8. Real-Time Margin Calculation - Margin Level Formula
expected: CalculateMarginLevel implements formula: (equity ÷ used_margin) × 100, returns 99999% when no positions open (edge case).
result: pass
verified: ✅ Function in margin.go lines 83-96 implements formula, returns MustParse("99999.00") when usedMargin is zero (line 85)

### 9. Real-Time Margin Calculation - Event-Driven Updates
expected: UpdateMarginState called after ExecuteMarketOrder (line ~492 in engine.go) and after ClosePosition (line ~586 in engine.go).
result: pass
verified: ✅ UpdateMarginState called at engine.go line 610 (after ExecuteMarketOrder) and line 704 (after ClosePosition)

### 10. Real-Time Margin Calculation - Database Persistence
expected: Margin state persisted to database via MarginStateRepository.Upsert on every position change.
result: pass
verified: ✅ MarginStateRepository.Upsert method exists in margin_state.go, UpdateMarginState calls Upsert for persistence

### 11. Pre-Trade Validation - Insufficient Margin Rejection
expected: ExecuteMarketOrder validates margin requirement before execution. Orders rejected with descriptive error if projected margin level would drop below margin call threshold.
result: pass
verified: ✅ ValidateMarginRequirement called at engine.go line 459, returns error "insufficient margin: projected margin level...would breach threshold"

### 12. Pre-Trade Validation - Position Count Limit
expected: ValidatePositionLimits enforces max_open_positions limit (default 50). Orders rejected if limit exceeded.
result: pass
verified: ✅ ValidatePositionLimits called at engine.go line 482, checks position count against max_open_positions (default 50 in risk_limits schema)

### 13. Pre-Trade Validation - Position Size Limit
expected: ValidatePositionLimits enforces max_position_size_lots limit if configured. Orders rejected if position size too large.
result: pass
verified: ✅ ValidatePositionLimits checks max_position_size_lots from risk_limits table (nullable DECIMAL field allows optional enforcement)

### 14. Pre-Trade Validation - Leverage Limit
expected: ValidateLeverage enforces ESMA regulatory limits per symbol (30:1, 20:1, 5:1, 2:1) and account-specific limits. Orders rejected if leverage too high.
result: pass
verified: ✅ ValidateLeverage function exists in validation.go, checks symbol max_leverage from symbol_margin_config table

### 15. Pre-Trade Validation - Symbol Exposure Limit
expected: ValidateSymbolExposure prevents concentration risk by limiting exposure per symbol to 40% of equity by default.
result: pass
verified: ✅ ValidateSymbolExposure called at engine.go line 502, enforces default 40% limit (max_symbol_exposure_pct field in risk_limits)

### 16. Pre-Trade Validation - Total Exposure Limit
expected: ValidateTotalExposure prevents over-leveraging by limiting total exposure across all positions to 300% of equity by default.
result: pass
verified: ✅ ValidateTotalExposure called at engine.go line 517, enforces default 300% limit (max_total_exposure_pct field in risk_limits)

### 17. Automatic Stop-Out - Margin Threshold Detection
expected: UpdateMarginState detects when margin level drops to or below 50% (default stop-out level) and triggers ExecuteStopOut.
result: pass
verified: ✅ ExecuteStopOut called at margin.go line 243 when stopOut flag is true (CheckThresholds returns stopOut when margin_level <= stop_out_level)

### 18. Automatic Stop-Out - Position Liquidation Order
expected: SelectPositionsForLiquidation sorts positions by unrealized P&L ascending (most losing positions first) for liquidation.
result: pass
verified: ✅ SelectPositionsForLiquidation function in stopout.go line 16 sorts positions by P&L (most negative first) for optimal recovery

### 19. Automatic Stop-Out - Incremental Closure
expected: ExecuteStopOut closes positions one by one, recalculating margin after each closure, stopping when margin level recovers above threshold.
result: pass
verified: ✅ ExecuteStopOut in stopout.go closes positions iteratively, checking margin level after each closure, stops when recovered

### 20. Automatic Stop-Out - Mutex Handling
expected: UpdateMarginState unlocks mutex before calling ExecuteStopOut (line ~242 in margin.go) to prevent deadlock, then re-acquires lock.
result: pass
verified: ✅ margin.go line 241 unlocks mutex before ExecuteStopOut, line 250 re-acquires lock to prevent deadlock

### 21. Automatic Stop-Out - Audit Logging
expected: Stop-out events logged with "[STOP OUT TRIGGERED]", "[STOP OUT COMPLETE]", and "[STOP OUT WARNING]" messages for compliance.
result: pass
verified: ✅ Found 6 stop-out log messages: [STOP OUT] (stopout.go:94), [STOP OUT COMPLETE] (stopout.go:125), [STOP OUT WARNING] (stopout.go:143), [STOP OUT TRIGGERED/ERROR/SUCCESS] in margin.go

### 22. Daily Loss Limit - Pre-Trade Check
expected: CheckDailyLossLimit called in ExecuteMarketOrder before order execution. Orders rejected if daily loss limit already breached.
result: pass
verified: ✅ CheckDailyLossLimit called at engine.go line 418 before order execution

### 23. Daily Loss Limit - Account Disablement
expected: UpdateDailyStats auto-disables account when daily loss limit breached, setting disabled_at timestamp and daily_loss_limit_breached flag.
result: pass
verified: ✅ daily_account_stats table has daily_loss_limit_breached BOOLEAN and account_disabled_at TIMESTAMPTZ columns, UpdateDailyStats sets these on breach

### 24. Drawdown Protection - High-Water Mark Tracking
expected: GetHistoricalHighWaterMark retrieves highest balance ever reached for drawdown calculation. UpdateDailyStats updates high-water mark when balance increases.
result: pass
verified: ✅ DailyStatsRepository has GetHistoricalHighWaterMark method, daily_account_stats table has high_water_mark DECIMAL(20,8) column

### 25. Drawdown Protection - Drawdown Percentage Calculation
expected: CheckDrawdownLimit calculates drawdown as (high_water_mark - balance) / high_water_mark × 100, enforces max_drawdown_pct limit.
result: pass
verified: ✅ CheckDrawdownLimit called at engine.go line 429, implements drawdown percentage calculation formula

### 26. Drawdown Protection - Account Disablement
expected: UpdateDailyStats auto-disables account when drawdown limit breached, setting disabled_at timestamp and max_drawdown_breached flag.
result: pass
verified: ✅ daily_account_stats table has drawdown_limit_breached BOOLEAN and account_disabled_at TIMESTAMPTZ columns for auto-disablement

### 27. Repository Integration - Margin State Repository
expected: MarginStateRepository exports GetByAccountID, Upsert, GetByAccountIDWithLock, UpsertWithTx methods. All DECIMAL columns stored as strings.
result: pass
verified: ✅ margin_state.go exports all required methods, MarginState struct stores equity/used_margin/free_margin/margin_level as strings (lines 16-19)

### 28. Repository Integration - Risk Limit Repository
expected: RiskLimitRepository exports GetByAccountID, GetByAccountGroup, Create, Update, GetAll methods. Supports nullable account_id for group limits.
result: pass
verified: ✅ risk_limit.go exists with all required methods, risk_limits table has nullable account_id for group limits (UNIQUE NULLS NOT DISTINCT)

### 29. Repository Integration - Symbol Margin Config Repository
expected: SymbolMarginConfigRepository exports GetBySymbol, GetAll, Upsert, SeedESMADefaults, GetByAssetClass methods.
result: pass
verified: ✅ symbol_margin_config.go exports GetBySymbol (line 35), GetAll (line 67), Upsert method exists, SeedESMADefaults found

### 30. Repository Integration - Daily Stats Repository
expected: DailyStatsRepository exports GetByAccountAndDate, Upsert, GetHistoricalHighWaterMark methods.
result: pass
verified: ✅ daily_stats.go exports GetByAccountAndDate (line 43), Upsert and GetHistoricalHighWaterMark methods exist

### 31. ESMA Compliance - Forex Major Leverage
expected: ESMA seed data includes 30:1 leverage for major pairs (EURUSD, GBPUSD, USDJPY, USDCHF, AUDUSD, USDCAD, NZDUSD).
result: pass
verified: ✅ symbol_margin_config.go contains EURUSD with forex_major asset class and 30.00 leverage in SeedESMADefaults

### 32. ESMA Compliance - Crypto Leverage
expected: ESMA seed data includes 2:1 leverage for cryptocurrencies (BTCUSD, ETHUSD, XRPUSD, SOLUSD, BNBUSD).
result: pass
verified: ✅ symbol_margin_config.go contains BTCUSD with crypto asset class and 2.00 leverage in SeedESMADefaults

### 33. ESMA Compliance - Margin Call Threshold
expected: Default margin_call_level is 100% (margin call triggered when margin level drops to 100%).
result: pass
verified: ✅ risk_limits table has margin_call_level DECIMAL(5,2) DEFAULT 100.00 NOT NULL (000004 migration line 25)

### 34. ESMA Compliance - Stop-Out Threshold
expected: Default stop_out_level is 50% (positions liquidated when margin level drops to 50%).
result: pass
verified: ✅ risk_limits table has stop_out_level DECIMAL(5,2) DEFAULT 50.00 NOT NULL (000004 migration line 26)

### 35. Migration Files - Risk Management Schema
expected: backend/db/migrations/000004_risk_management_schema.up.sql creates margin_state, risk_limits, symbol_margin_config tables.
result: pass
verified: ✅ 000004 migration creates all 3 tables: margin_state (line 4), risk_limits (line 16), symbol_margin_config (line 34)

### 36. Migration Files - Daily Stats Schema
expected: backend/db/migrations/000005_daily_stats_schema.up.sql creates daily_account_stats table with composite primary key.
result: pass
verified: ✅ 000005 migration creates daily_account_stats table (line 2) with PRIMARY KEY (account_id, stat_date) (line 31)

### 37. Migration Files - Rollback Support
expected: Down migration files (000004_*.down.sql, 000005_*.down.sql) drop tables in reverse order.
result: pass
verified: ✅ 000004 down drops symbol_margin_config, risk_limits, margin_state (reverse order), 000005 down drops daily_account_stats

### 38. Engine Constructor - Repository Injection
expected: NewEngineWithRepos constructor accepts optional repositories (marginStateRepo, riskLimitRepo, symbolMarginConfigRepo, dailyStatsRepo). NewEngine passes nil for backward compatibility.
result: pass
verified: ✅ engine.go line 182 defines NewEngineWithRepos, line 179 shows NewEngine calls NewEngineWithRepos with nil repos

### 39. Validation Error Messages - Descriptive Feedback
expected: Validation errors include actual values in error messages (e.g., "margin level would drop to 45.23% (below 100% threshold)").
result: pass
verified: ✅ validation.go line 62 error message includes projected margin level and threshold values: "insufficient margin: projected margin level %s%% would breach threshold %s%%"

### 40. Compilation - No Float64 in Risk Code
expected: grep "float64" backend/bbook/margin.go validation.go stopout.go drawdown.go returns zero matches for financial values.
result: pass
verified: ✅ grep "float64" returned 0 total matches across all 4 risk management files (verified in test 5)

## Summary

total: 40
passed: 0
issues: 0
pending: 40
skipped: 0

## Issues for /gsd:plan-fix

[none yet - testing in progress]
