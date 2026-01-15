-- Migration: Add exposure limit columns to risk_limits table
-- Created: 2026-01-16
-- Purpose: Enable per-account configuration of symbol and total exposure limits

ALTER TABLE risk_limits
ADD COLUMN IF NOT EXISTS max_symbol_exposure_pct DECIMAL(5,2),
ADD COLUMN IF NOT EXISTS max_total_exposure_pct DECIMAL(6,2);

-- Set default values for existing rows
UPDATE risk_limits
SET max_symbol_exposure_pct = 40.00
WHERE max_symbol_exposure_pct IS NULL;

UPDATE risk_limits
SET max_total_exposure_pct = 300.00
WHERE max_total_exposure_pct IS NULL;
