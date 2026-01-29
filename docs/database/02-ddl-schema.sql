-- ============================================================================
-- RTX Trading Engine - PostgreSQL Schema DDL
-- Version: 1.0
-- Description: Production-grade schema for trading platform
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- ============================================================================
-- SCHEMA: users
-- Purpose: Authentication, authorization, and user management
-- ============================================================================

-- Users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt hash
    role VARCHAR(20) NOT NULL DEFAULT 'TRADER', -- ADMIN, TRADER, READONLY
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED, CLOSED
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    two_factor_secret VARCHAR(255),
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(50),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

COMMENT ON TABLE users IS 'System users with authentication credentials';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash with cost factor 12';
COMMENT ON COLUMN users.metadata IS 'Additional user metadata in JSON format';

-- Roles table
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description, permissions) VALUES
('ADMIN', 'Full system access', '["*"]'::jsonb),
('TRADER', 'Trading operations', '["trade:execute", "account:view", "position:view"]'::jsonb),
('READONLY', 'Read-only access', '["account:view", "position:view"]'::jsonb);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ============================================================================
-- SCHEMA: accounts
-- Purpose: Trading accounts, balances, and account configuration
-- ============================================================================

-- Accounts table
CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    account_number VARCHAR(20) NOT NULL UNIQUE, -- RTX-000001
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    account_type VARCHAR(20) NOT NULL, -- DEMO, LIVE
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    balance DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    equity DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    margin DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    free_margin DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    margin_level DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    leverage INT NOT NULL DEFAULT 100,
    margin_mode VARCHAR(20) NOT NULL DEFAULT 'HEDGING', -- HEDGING, NETTING
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED, CLOSED
    stop_out_level DECIMAL(5, 2) NOT NULL DEFAULT 50.00,
    margin_call_level DECIMAL(5, 2) NOT NULL DEFAULT 100.00,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT chk_account_type CHECK (account_type IN ('DEMO', 'LIVE')),
    CONSTRAINT chk_margin_mode CHECK (margin_mode IN ('HEDGING', 'NETTING')),
    CONSTRAINT chk_status CHECK (status IN ('ACTIVE', 'SUSPENDED', 'CLOSED')),
    CONSTRAINT chk_positive_balance CHECK (balance >= 0)
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_account_number ON accounts(account_number);
CREATE INDEX idx_accounts_status ON accounts(status);
CREATE INDEX idx_accounts_account_type ON accounts(account_type);

COMMENT ON TABLE accounts IS 'Trading accounts with margin and equity tracking';
COMMENT ON COLUMN accounts.margin_mode IS 'HEDGING allows multiple positions per symbol, NETTING allows only one';

-- Account history table (for audit)
CREATE TABLE account_history (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- CREATED, MODIFIED, SUSPENDED, CLOSED
    old_values JSONB,
    new_values JSONB,
    changed_by VARCHAR(50),
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_account_history_account_id ON account_history(account_id);
CREATE INDEX idx_account_history_created_at ON account_history(created_at);

-- ============================================================================
-- SCHEMA: instruments
-- Purpose: Trading symbols, specifications, and trading hours
-- ============================================================================

-- Instruments table
CREATE TABLE instruments (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL UNIQUE,
    base_currency VARCHAR(10) NOT NULL,
    quote_currency VARCHAR(10) NOT NULL,
    instrument_type VARCHAR(20) NOT NULL, -- FOREX, CRYPTO, CFD, COMMODITY
    contract_size DECIMAL(18, 4) NOT NULL DEFAULT 100000,
    pip_size DECIMAL(10, 8) NOT NULL DEFAULT 0.0001,
    pip_value DECIMAL(10, 4) NOT NULL DEFAULT 10.00,
    min_volume DECIMAL(10, 4) NOT NULL DEFAULT 0.01,
    max_volume DECIMAL(10, 4) NOT NULL DEFAULT 100.00,
    volume_step DECIMAL(10, 4) NOT NULL DEFAULT 0.01,
    min_spread DECIMAL(10, 8) NOT NULL DEFAULT 0.00001,
    margin_percent DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    commission_per_lot DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    swap_long DECIMAL(10, 4) NOT NULL DEFAULT 0.00,
    swap_short DECIMAL(10, 4) NOT NULL DEFAULT 0.00,
    swap_type VARCHAR(20) NOT NULL DEFAULT 'POINTS', -- POINTS, CURRENCY, PERCENTAGE
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    trading_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    quote_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_instruments_symbol ON instruments(symbol);
CREATE INDEX idx_instruments_type ON instruments(instrument_type);
CREATE INDEX idx_instruments_enabled ON instruments(enabled);

COMMENT ON TABLE instruments IS 'Trading instrument specifications and contract details';

-- Trading hours table
CREATE TABLE trading_hours (
    id SERIAL PRIMARY KEY,
    instrument_id INT NOT NULL REFERENCES instruments(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL, -- 0=Sunday, 6=Saturday
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_trading_hours_instrument_id ON trading_hours(instrument_id);

-- ============================================================================
-- SCHEMA: orders
-- Purpose: Order lifecycle, pending orders, and order history
-- ============================================================================

-- Orders table
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    symbol VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL, -- MARKET, LIMIT, STOP, STOP_LIMIT
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    volume DECIMAL(18, 4) NOT NULL,
    price DECIMAL(18, 8),
    trigger_price DECIMAL(18, 8),
    limit_price DECIMAL(18, 8),
    stop_loss DECIMAL(18, 8),
    take_profit DECIMAL(18, 8),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, TRIGGERED, FILLED, PARTIALLY_FILLED, CANCELLED, REJECTED, EXPIRED
    filled_volume DECIMAL(18, 4) NOT NULL DEFAULT 0,
    filled_price DECIMAL(18, 8),
    filled_at TIMESTAMP WITH TIME ZONE,
    position_id BIGINT,
    expiry_time TIMESTAMP WITH TIME ZONE,
    max_slippage DECIMAL(10, 8),
    time_in_force VARCHAR(20) NOT NULL DEFAULT 'GTC', -- GTC, IOC, FOK, DAY
    oco_pair_id BIGINT,
    reject_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT chk_order_type CHECK (order_type IN ('MARKET', 'LIMIT', 'STOP', 'STOP_LIMIT')),
    CONSTRAINT chk_side CHECK (side IN ('BUY', 'SELL')),
    CONSTRAINT chk_status CHECK (status IN ('PENDING', 'TRIGGERED', 'FILLED', 'PARTIALLY_FILLED', 'CANCELLED', 'REJECTED', 'EXPIRED')),
    CONSTRAINT chk_time_in_force CHECK (time_in_force IN ('GTC', 'IOC', 'FOK', 'DAY')),
    CONSTRAINT chk_positive_volume CHECK (volume > 0)
);

CREATE INDEX idx_orders_account_id ON orders(account_id);
CREATE INDEX idx_orders_symbol ON orders(symbol);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX idx_orders_account_symbol_status ON orders(account_id, symbol, status);

COMMENT ON TABLE orders IS 'All orders including pending, filled, and cancelled';

-- Order fills table (for partial fills)
CREATE TABLE order_fills (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    fill_volume DECIMAL(18, 4) NOT NULL,
    fill_price DECIMAL(18, 8) NOT NULL,
    commission DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    liquidity_provider VARCHAR(50),
    execution_venue VARCHAR(50),
    filled_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_fills_order_id ON order_fills(order_id);
CREATE INDEX idx_order_fills_filled_at ON order_fills(filled_at DESC);

-- ============================================================================
-- SCHEMA: positions
-- Purpose: Open and closed positions with P&L tracking
-- ============================================================================

-- Positions table
CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    volume DECIMAL(18, 4) NOT NULL,
    open_price DECIMAL(18, 8) NOT NULL,
    current_price DECIMAL(18, 8) NOT NULL,
    stop_loss DECIMAL(18, 8),
    take_profit DECIMAL(18, 8),
    commission DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    swap DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    unrealized_pnl DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    realized_pnl DECIMAL(18, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN', -- OPEN, CLOSED
    open_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    close_time TIMESTAMP WITH TIME ZONE,
    close_price DECIMAL(18, 8),
    close_reason VARCHAR(50), -- MANUAL, STOP_LOSS, TAKE_PROFIT, MARGIN_CALL, SYSTEM
    magic_number INT,
    comment TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT chk_position_side CHECK (side IN ('BUY', 'SELL')),
    CONSTRAINT chk_position_status CHECK (status IN ('OPEN', 'CLOSED')),
    CONSTRAINT chk_positive_volume CHECK (volume > 0)
);

CREATE INDEX idx_positions_account_id ON positions(account_id);
CREATE INDEX idx_positions_symbol ON positions(symbol);
CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_account_status ON positions(account_id, status);
CREATE INDEX idx_positions_open_time ON positions(open_time DESC);

COMMENT ON TABLE positions IS 'Trading positions with P&L tracking';

-- Position history table (for partial closes)
CREATE TABLE position_history (
    id BIGSERIAL PRIMARY KEY,
    position_id BIGINT NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- OPENED, MODIFIED, PARTIAL_CLOSE, CLOSED
    volume_change DECIMAL(18, 4),
    price DECIMAL(18, 8),
    realized_pnl DECIMAL(18, 2),
    old_stop_loss DECIMAL(18, 8),
    new_stop_loss DECIMAL(18, 8),
    old_take_profit DECIMAL(18, 8),
    new_take_profit DECIMAL(18, 8),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_position_history_position_id ON position_history(position_id);
CREATE INDEX idx_position_history_created_at ON position_history(created_at DESC);

-- ============================================================================
-- SCHEMA: transactions
-- Purpose: Financial ledger with double-entry accounting
-- ============================================================================

-- Ledger table
CREATE TABLE ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    transaction_type VARCHAR(50) NOT NULL, -- DEPOSIT, WITHDRAWAL, REALIZED_PNL, COMMISSION, SWAP, ADJUSTMENT, BONUS, TRANSFER
    amount DECIMAL(18, 2) NOT NULL,
    balance_before DECIMAL(18, 2) NOT NULL,
    balance_after DECIMAL(18, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    description TEXT,
    reference_type VARCHAR(50), -- TRADE, POSITION, ORDER, ADMIN, SYSTEM
    reference_id BIGINT,
    payment_method VARCHAR(50), -- BANK_TRANSFER, CREDIT_CARD, CRYPTO, MANUAL, BONUS
    payment_reference VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED', -- PENDING, COMPLETED, FAILED, REVERSED
    processed_by VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT chk_transaction_type CHECK (transaction_type IN ('DEPOSIT', 'WITHDRAWAL', 'REALIZED_PNL', 'COMMISSION', 'SWAP', 'ADJUSTMENT', 'BONUS', 'TRANSFER')),
    CONSTRAINT chk_ledger_status CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED'))
);

CREATE INDEX idx_ledger_account_id ON ledger_entries(account_id);
CREATE INDEX idx_ledger_transaction_type ON ledger_entries(transaction_type);
CREATE INDEX idx_ledger_created_at ON ledger_entries(created_at DESC);
CREATE INDEX idx_ledger_account_created ON ledger_entries(account_id, created_at DESC);
CREATE INDEX idx_ledger_reference ON ledger_entries(reference_type, reference_id);

COMMENT ON TABLE ledger_entries IS 'Financial transaction ledger with audit trail';

-- Deposits/Withdrawals table
CREATE TABLE payment_transactions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    ledger_entry_id BIGINT REFERENCES ledger_entries(id),
    transaction_type VARCHAR(20) NOT NULL, -- DEPOSIT, WITHDRAWAL
    amount DECIMAL(18, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,
    payment_provider VARCHAR(50),
    external_transaction_id VARCHAR(255),
    bank_account_last4 VARCHAR(4),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, PROCESSING, COMPLETED, FAILED, CANCELLED
    failure_reason TEXT,
    approved_by VARCHAR(50),
    approved_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb,
    CONSTRAINT chk_payment_type CHECK (transaction_type IN ('DEPOSIT', 'WITHDRAWAL')),
    CONSTRAINT chk_payment_status CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED'))
);

CREATE INDEX idx_payment_transactions_account_id ON payment_transactions(account_id);
CREATE INDEX idx_payment_transactions_status ON payment_transactions(status);
CREATE INDEX idx_payment_transactions_created_at ON payment_transactions(created_at DESC);

-- ============================================================================
-- SCHEMA: risk
-- Purpose: Risk management, exposure tracking, and margin calls
-- ============================================================================

-- Risk limits table
CREATE TABLE risk_limits (
    id SERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES accounts(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    limit_type VARCHAR(50) NOT NULL, -- MAX_POSITION_SIZE, MAX_LEVERAGE, MAX_DRAWDOWN, MAX_DAILY_LOSS
    limit_value DECIMAL(18, 4) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_one_target CHECK (
        (account_id IS NOT NULL AND user_id IS NULL) OR
        (account_id IS NULL AND user_id IS NOT NULL)
    )
);

CREATE INDEX idx_risk_limits_account_id ON risk_limits(account_id);
CREATE INDEX idx_risk_limits_user_id ON risk_limits(user_id);

-- Margin calls table
CREATE TABLE margin_calls (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    margin_level DECIMAL(10, 2) NOT NULL,
    equity DECIMAL(18, 2) NOT NULL,
    margin DECIMAL(18, 2) NOT NULL,
    call_type VARCHAR(20) NOT NULL, -- WARNING, MARGIN_CALL, STOP_OUT
    notified_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_type VARCHAR(20), -- DEPOSIT, POSITION_CLOSED, LIQUIDATED
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_margin_calls_account_id ON margin_calls(account_id);
CREATE INDEX idx_margin_calls_notified_at ON margin_calls(notified_at DESC);

-- Exposure tracking table
CREATE TABLE exposure_snapshots (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    net_volume DECIMAL(18, 4) NOT NULL,
    buy_volume DECIMAL(18, 4) NOT NULL,
    sell_volume DECIMAL(18, 4) NOT NULL,
    net_exposure_usd DECIMAL(18, 2) NOT NULL,
    position_count INT NOT NULL,
    account_count INT NOT NULL,
    snapshot_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exposure_symbol_time ON exposure_snapshots(symbol, snapshot_time DESC);

-- ============================================================================
-- SCHEMA: audit
-- Purpose: Compliance, regulatory reporting, and admin actions
-- ============================================================================

-- Admin actions audit log
CREATE TABLE admin_audit_log (
    id BIGSERIAL PRIMARY KEY,
    admin_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    admin_username VARCHAR(50) NOT NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50), -- USER, ACCOUNT, ORDER, POSITION, SYMBOL
    target_id VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    ip_address INET NOT NULL,
    user_agent TEXT,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_admin_user_id ON admin_audit_log(admin_user_id);
CREATE INDEX idx_admin_audit_created_at ON admin_audit_log(created_at DESC);
CREATE INDEX idx_admin_audit_action ON admin_audit_log(action);
CREATE INDEX idx_admin_audit_target ON admin_audit_log(target_type, target_id);

COMMENT ON TABLE admin_audit_log IS 'Immutable audit log of all admin actions';

-- Trade execution audit log
CREATE TABLE trade_execution_log (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    order_id BIGINT,
    position_id BIGINT,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    volume DECIMAL(18, 4) NOT NULL,
    price DECIMAL(18, 8) NOT NULL,
    execution_venue VARCHAR(50), -- B_BOOK, OANDA, BINANCE, etc.
    latency_ms INT,
    slippage DECIMAL(10, 8),
    executed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_trade_execution_account_id ON trade_execution_log(account_id);
CREATE INDEX idx_trade_execution_executed_at ON trade_execution_log(executed_at DESC);
CREATE INDEX idx_trade_execution_symbol ON trade_execution_log(symbol);

COMMENT ON TABLE trade_execution_log IS 'Detailed execution log for regulatory compliance';

-- System events log
CREATE TABLE system_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL, -- STARTUP, SHUTDOWN, ERROR, WARNING, INFO
    severity VARCHAR(20) NOT NULL DEFAULT 'INFO', -- DEBUG, INFO, WARNING, ERROR, CRITICAL
    component VARCHAR(50) NOT NULL, -- API, EXECUTION, RISK, LP_GATEWAY, etc.
    message TEXT NOT NULL,
    stack_trace TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_system_events_event_type ON system_events(event_type);
CREATE INDEX idx_system_events_severity ON system_events(severity);
CREATE INDEX idx_system_events_created_at ON system_events(created_at DESC);
CREATE INDEX idx_system_events_component ON system_events(component);

-- ============================================================================
-- TRIGGERS & FUNCTIONS
-- ============================================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_instruments_updated_at BEFORE UPDATE ON instruments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically set account balance from ledger
CREATE OR REPLACE FUNCTION update_account_balance_from_ledger()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE accounts
    SET balance = NEW.balance_after,
        updated_at = NOW()
    WHERE id = NEW.account_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_account_balance
    AFTER INSERT ON ledger_entries
    FOR EACH ROW
    WHEN (NEW.status = 'COMPLETED')
    EXECUTE FUNCTION update_account_balance_from_ledger();

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- Account summary view
CREATE VIEW v_account_summary AS
SELECT
    a.id,
    a.account_number,
    a.user_id,
    u.username,
    a.account_type,
    a.currency,
    a.balance,
    COALESCE(SUM(p.unrealized_pnl), 0) as total_unrealized_pnl,
    a.balance + COALESCE(SUM(p.unrealized_pnl), 0) as equity,
    COALESCE(SUM(
        CASE WHEN p.status = 'OPEN' THEN
            (p.volume * i.contract_size * p.open_price) / a.leverage
        ELSE 0 END
    ), 0) as used_margin,
    a.balance + COALESCE(SUM(p.unrealized_pnl), 0) -
    COALESCE(SUM(
        CASE WHEN p.status = 'OPEN' THEN
            (p.volume * i.contract_size * p.open_price) / a.leverage
        ELSE 0 END
    ), 0) as free_margin,
    COUNT(CASE WHEN p.status = 'OPEN' THEN 1 END) as open_positions
FROM accounts a
LEFT JOIN users u ON a.user_id = u.id
LEFT JOIN positions p ON a.id = p.account_id
LEFT JOIN instruments i ON p.symbol = i.symbol
WHERE a.status = 'ACTIVE'
GROUP BY a.id, u.username;

-- Open positions view
CREATE VIEW v_open_positions AS
SELECT
    p.id,
    p.account_id,
    a.account_number,
    p.symbol,
    p.side,
    p.volume,
    p.open_price,
    p.current_price,
    p.stop_loss,
    p.take_profit,
    p.commission,
    p.swap,
    p.unrealized_pnl,
    p.open_time,
    i.pip_size,
    i.pip_value
FROM positions p
JOIN accounts a ON p.account_id = a.id
JOIN instruments i ON p.symbol = i.symbol
WHERE p.status = 'OPEN';

-- Pending orders view
CREATE VIEW v_pending_orders AS
SELECT
    o.id,
    o.account_id,
    a.account_number,
    o.symbol,
    o.order_type,
    o.side,
    o.volume,
    o.price,
    o.trigger_price,
    o.stop_loss,
    o.take_profit,
    o.status,
    o.time_in_force,
    o.created_at
FROM orders o
JOIN accounts a ON o.account_id = a.id
WHERE o.status IN ('PENDING', 'TRIGGERED', 'PARTIALLY_FILLED');

-- ============================================================================
-- GRANT PERMISSIONS
-- ============================================================================

-- Create application roles
CREATE ROLE trading_app WITH LOGIN PASSWORD 'CHANGE_ME_IN_PRODUCTION';
CREATE ROLE readonly_app WITH LOGIN PASSWORD 'CHANGE_ME_IN_PRODUCTION';
CREATE ROLE admin_app WITH LOGIN PASSWORD 'CHANGE_ME_IN_PRODUCTION';

-- Grant permissions to trading_app (read/write)
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO trading_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO trading_app;

-- Grant permissions to readonly_app (read only)
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_app;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO readonly_app;

-- Grant all permissions to admin_app
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO admin_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO admin_app;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
