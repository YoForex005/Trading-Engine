BEGIN;

-- Margin state table (per-account real-time margin)
CREATE TABLE margin_state (
    account_id BIGINT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
    equity DECIMAL(20,8) NOT NULL CHECK (equity >= 0),
    used_margin DECIMAL(20,8) NOT NULL DEFAULT 0 CHECK (used_margin >= 0),
    free_margin DECIMAL(20,8) NOT NULL GENERATED ALWAYS AS (equity - used_margin) STORED,
    margin_level DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (margin_level >= 0),
    margin_call_triggered BOOLEAN DEFAULT FALSE,
    stop_out_triggered BOOLEAN DEFAULT FALSE,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Risk limits table (per-account or account group limits)
CREATE TABLE risk_limits (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE,
    account_group VARCHAR(50),
    max_leverage DECIMAL(5,2) NOT NULL CHECK (max_leverage > 0) DEFAULT 30.00,
    max_open_positions INT NOT NULL DEFAULT 50 CHECK (max_open_positions > 0),
    max_position_size_lots DECIMAL(10,2) CHECK (max_position_size_lots > 0),
    daily_loss_limit DECIMAL(20,8) CHECK (daily_loss_limit > 0),
    max_drawdown_pct DECIMAL(5,2) CHECK (max_drawdown_pct > 0 AND max_drawdown_pct <= 100),
    margin_call_level DECIMAL(5,2) DEFAULT 100.00 NOT NULL CHECK (margin_call_level > 0),
    stop_out_level DECIMAL(5,2) DEFAULT 50.00 NOT NULL CHECK (stop_out_level > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT margin_levels_valid CHECK (margin_call_level > stop_out_level),
    CONSTRAINT unique_account_or_group UNIQUE NULLS NOT DISTINCT (account_id, account_group)
);

-- Symbol margin configuration table (ESMA regulatory limits)
CREATE TABLE symbol_margin_config (
    symbol VARCHAR(20) PRIMARY KEY,
    asset_class VARCHAR(20) NOT NULL CHECK (asset_class IN ('forex_major', 'forex_minor', 'stock', 'crypto', 'commodity', 'index')),
    max_leverage DECIMAL(5,2) NOT NULL CHECK (max_leverage > 0),
    margin_percentage DECIMAL(5,4) NOT NULL CHECK (margin_percentage > 0 AND margin_percentage <= 100),
    contract_size DECIMAL(20,8) NOT NULL DEFAULT 100000 CHECK (contract_size > 0),
    tick_size DECIMAL(10,8) NOT NULL DEFAULT 0.0001 CHECK (tick_size > 0),
    tick_value DECIMAL(20,8) NOT NULL DEFAULT 10 CHECK (tick_value > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_margin_state_last_updated ON margin_state(last_updated);
CREATE INDEX idx_risk_limits_account ON risk_limits(account_id) WHERE account_id IS NOT NULL;
CREATE INDEX idx_risk_limits_group ON risk_limits(account_group) WHERE account_group IS NOT NULL;
CREATE INDEX idx_symbol_margin_config_asset_class ON symbol_margin_config(asset_class);

COMMIT;
