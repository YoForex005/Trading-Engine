BEGIN;

-- Accounts table
CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,  -- Will be bcrypt hashed in Phase 1
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    equity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    margin DECIMAL(20, 8) NOT NULL DEFAULT 0,
    free_margin DECIMAL(20, 8) NOT NULL DEFAULT 0,
    margin_level DECIMAL(20, 8) NOT NULL DEFAULT 0,
    leverage DECIMAL(10, 2) NOT NULL DEFAULT 100,
    margin_mode VARCHAR(20) NOT NULL DEFAULT 'NETTING',  -- HEDGING or NETTING
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',  -- ACTIVE or DISABLED
    is_demo BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Positions table
CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    volume DECIMAL(20, 8) NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    current_price DECIMAL(20, 8) NOT NULL,
    open_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sl DECIMAL(20, 8),
    tp DECIMAL(20, 8),
    swap DECIMAL(20, 8) NOT NULL DEFAULT 0,
    commission DECIMAL(20, 8) NOT NULL DEFAULT 0,
    unrealized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    close_price DECIMAL(20, 8),
    close_time TIMESTAMPTZ,
    close_reason VARCHAR(50),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Orders table
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,  -- MARKET, LIMIT, STOP, STOP_LIMIT
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    volume DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8),
    trigger_price DECIMAL(20, 8),
    sl DECIMAL(20, 8),
    tp DECIMAL(20, 8),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    filled_price DECIMAL(20, 8),
    filled_at TIMESTAMPTZ,
    position_id BIGINT REFERENCES positions(id) ON DELETE SET NULL,
    reject_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Trades table (completed/closed positions)
CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES orders(id) ON DELETE SET NULL,
    position_id BIGINT,  -- No FK since position may be deleted
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(4) NOT NULL CHECK (side IN ('BUY', 'SELL')),
    volume DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    realized_pnl DECIMAL(20, 8) NOT NULL,
    commission DECIMAL(20, 8) NOT NULL DEFAULT 0,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_positions_account ON positions(account_id) WHERE status = 'OPEN';
CREATE INDEX idx_positions_symbol ON positions(symbol) WHERE status = 'OPEN';
CREATE INDEX idx_orders_account ON orders(account_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_trades_account_time ON trades(account_id, executed_at DESC);
CREATE INDEX idx_trades_symbol_time ON trades(symbol, executed_at DESC);

COMMIT;
