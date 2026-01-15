-- Rollback: Remove exposure limit columns from risk_limits table

ALTER TABLE risk_limits
DROP COLUMN IF EXISTS max_symbol_exposure_pct,
DROP COLUMN IF EXISTS max_total_exposure_pct;
