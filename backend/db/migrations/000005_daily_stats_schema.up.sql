-- Daily account statistics for loss limits and drawdown tracking
CREATE TABLE daily_account_stats (
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    stat_date DATE NOT NULL,

    -- Daily P&L tracking
    starting_balance DECIMAL(20,8) NOT NULL,
    ending_balance DECIMAL(20,8) NOT NULL,
    realized_pl DECIMAL(20,8) NOT NULL DEFAULT 0,
    unrealized_pl DECIMAL(20,8) NOT NULL DEFAULT 0,
    total_pl DECIMAL(20,8) NOT NULL DEFAULT 0,

    -- High-water mark for drawdown calculation
    high_water_mark DECIMAL(20,8) NOT NULL,
    current_drawdown_pct DECIMAL(5,2) NOT NULL DEFAULT 0,
    max_drawdown_pct DECIMAL(5,2) NOT NULL DEFAULT 0,

    -- Trading statistics
    trades_opened INT NOT NULL DEFAULT 0,
    trades_closed INT NOT NULL DEFAULT 0,
    winning_trades INT NOT NULL DEFAULT 0,
    losing_trades INT NOT NULL DEFAULT 0,

    -- Limit breach tracking
    daily_loss_limit_breached BOOLEAN DEFAULT FALSE,
    drawdown_limit_breached BOOLEAN DEFAULT FALSE,
    account_disabled_at TIMESTAMPTZ,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (account_id, stat_date)
);

CREATE INDEX idx_daily_stats_date ON daily_account_stats(stat_date);
CREATE INDEX idx_daily_stats_breached ON daily_account_stats(account_id) WHERE daily_loss_limit_breached OR drawdown_limit_breached;

-- Account status tracking (extends existing accounts table logic)
-- Note: This assumes accounts table has a status column; if not, this should be added
-- ALTER TABLE accounts ADD COLUMN status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'DISABLED', 'SUSPENDED'));
-- ALTER TABLE accounts ADD COLUMN disabled_reason TEXT;
