-- RTX Trading Engine - Core Database Schema
-- Database: PostgreSQL (TimescaleDB extension recommended for ticks/history)

-- 1. USERS & ACCOUNTS
-- Users: The physical entities (Traders, Admins)
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(50) NOT NULL DEFAULT 'TRADER', -- TRADER, ADMIN, SUPER_ADMIN
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Accounts: The trading buckets. One user can have multiple accounts (e.g. USD, EUR, Demo, Crypto)
CREATE TABLE accounts (
    account_id BIGSERIAL PRIMARY KEY, -- Integer ID often easier for Trading Terminals like MT5
    user_id UUID REFERENCES users(user_id),
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    balance DECIMAL(18, 8) DEFAULT 0.0,
    equity DECIMAL(18, 8) DEFAULT 0.0, -- Snapshot, real-time is in memory/Redis
    leverage INT DEFAULT 100,
    account_type VARCHAR(50) DEFAULT 'RETAIL', -- RETAIL, VIP, INSTITUTIONAL
    group_id VARCHAR(50) DEFAULT 'demo-forex', -- For routing rules (A-Book vs B-Book)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. INSTRUMENTS
-- Symbols: Tradable assets (EURUSD, BTCUSD, etc.)
CREATE TABLE symbols (
    symbol_id SERIAL PRIMARY KEY,
    name VARCHAR(20) UNIQUE NOT NULL, -- e.g. "EURUSD"
    base_currency VARCHAR(10) NOT NULL,
    quote_currency VARCHAR(10) NOT NULL,
    contract_size DECIMAL(18, 8) DEFAULT 100000,
    min_volume DECIMAL(18, 8) DEFAULT 0.01,
    max_volume DECIMAL(18, 8) DEFAULT 100.00,
    digits INT DEFAULT 5, -- Pip precision
    category VARCHAR(50), -- FOREX, CRYPTO, INDICES
    is_tradable BOOLEAN DEFAULT TRUE
);

-- 3. TRADING ACTIVITY
-- Orders: Instructions to Buy/Sell
CREATE TABLE orders (
    order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id BIGINT REFERENCES accounts(account_id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    type VARCHAR(20) NOT NULL, -- MARKET, LIMIT, STOP
    volume DECIMAL(18, 8) NOT NULL,
    price_requested DECIMAL(18, 8), -- For Limit/Stop
    price_executed DECIMAL(18, 8), -- For actual fills
    sl DECIMAL(18, 8), -- Stop Loss
    tp DECIMAL(18, 8), -- Take Profit
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, FILLED, REJECTED, CANCELED
    routing_type VARCHAR(20), -- A_BOOK, B_BOOK
    lp_order_id VARCHAR(100), -- ID from LMAX/LP if A-Booked
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    filled_at TIMESTAMP WITH TIME ZONE
);

-- Trades: History of executions (Account entries)
CREATE TABLE trades (
    trade_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(order_id),
    account_id BIGINT REFERENCES accounts(account_id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    volume DECIMAL(18, 8) NOT NULL,
    entry_price DECIMAL(18, 8) NOT NULL,
    exit_price DECIMAL(18, 8),
    profit DECIMAL(18, 8),
    commission DECIMAL(18, 8) DEFAULT 0.0,
    swap DECIMAL(18, 8) DEFAULT 0.0,
    open_time TIMESTAMP WITH TIME ZONE NOT NULL,
    close_time TIMESTAMP WITH TIME ZONE
);

-- 4. CONFIGURATION
-- Routing Rules: Decides where orders go
CREATE TABLE routing_rules (
    rule_id SERIAL PRIMARY KEY,
    group_pattern VARCHAR(100), -- e.g. "VIP-%", "demo-%"
    symbol_pattern VARCHAR(100), -- e.g. "EURUSD", "*"
    min_volume DECIMAL(18, 8) DEFAULT 0,
    action VARCHAR(20) NOT NULL, -- A_BOOK, B_BOOK, REJECT
    target_lp VARCHAR(50) -- e.g. "LMAX_PROD"
);

-- 5. TICK HISTORY (Broker-Level Market Data Storage)
-- Stores all incoming ticks for historical chart data
CREATE TABLE tick_history (
    id BIGSERIAL,
    broker_id VARCHAR(50) NOT NULL DEFAULT 'default',
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(18, 8) NOT NULL,
    ask DECIMAL(18, 8) NOT NULL,
    spread DECIMAL(18, 8),
    timestamp TIMESTAMPTZ NOT NULL,
    lp VARCHAR(50),
    PRIMARY KEY (id, timestamp)
);

-- Create hypertable for time-series optimization (TimescaleDB)
-- SELECT create_hypertable('tick_history', 'timestamp', if_not_exists => TRUE);

-- Index for fast symbol lookups
CREATE INDEX idx_tick_symbol_time ON tick_history (symbol, timestamp DESC);
CREATE INDEX idx_tick_broker ON tick_history (broker_id, symbol);

-- 6. ADVANCED ORDER TYPES
-- Extended order fields for pending orders, trailing stops, etc.
ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_subtype VARCHAR(20); -- BUY_STOP, SELL_LIMIT, etc.
ALTER TABLE orders ADD COLUMN IF NOT EXISTS trigger_price DECIMAL(18, 8); -- For stop orders
ALTER TABLE orders ADD COLUMN IF NOT EXISTS limit_price DECIMAL(18, 8); -- For stop-limit orders
ALTER TABLE orders ADD COLUMN IF NOT EXISTS expiry TIMESTAMPTZ; -- For pending orders (GTC, GTD)
ALTER TABLE orders ADD COLUMN IF NOT EXISTS oco_group_id UUID; -- For OCO linking
ALTER TABLE orders ADD COLUMN IF NOT EXISTS trailing_stop_distance DECIMAL(18, 8);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS trailing_stop_type VARCHAR(20); -- FIXED, STEP, ATR
ALTER TABLE orders ADD COLUMN IF NOT EXISTS max_slippage DECIMAL(18, 8);

-- 7. TP LADDER (Multiple Take-Profits with Partial Closes)
CREATE TABLE tp_ladder (
    id SERIAL PRIMARY KEY,
    trade_id VARCHAR(100) NOT NULL, -- OANDA trade ID
    order_id UUID REFERENCES orders(order_id),
    level INT NOT NULL, -- 1, 2, 3
    price DECIMAL(18, 8) NOT NULL,
    close_percent DECIMAL(5, 2) NOT NULL, -- e.g. 50.00
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, TRIGGERED, CANCELLED
    triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tp_ladder_trade ON tp_ladder (trade_id, status);

-- 8. PENDING ORDERS TABLE (for internal order book)
CREATE TABLE pending_orders (
    id SERIAL PRIMARY KEY,
    order_id UUID NOT NULL UNIQUE,
    account_id BIGINT,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    order_type VARCHAR(20) NOT NULL, -- LIMIT, STOP, STOP_LIMIT
    volume DECIMAL(18, 8) NOT NULL,
    entry_price DECIMAL(18, 8), -- For limit orders
    trigger_price DECIMAL(18, 8), -- For stop orders
    sl DECIMAL(18, 8),
    tp DECIMAL(18, 8),
    oco_pair_id UUID, -- Links to another pending order
    expiry TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, TRIGGERED, CANCELLED, EXPIRED
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    triggered_at TIMESTAMPTZ
);

CREATE INDEX idx_pending_orders_status ON pending_orders (status, symbol);

-- ============================================
-- 9. RTX B-BOOK INTERNAL TRADING SYSTEM
-- ============================================
-- All balance/equity/margin computed internally, NOT from LP

-- RTX Internal Accounts (B-Book source of truth)
CREATE TABLE rtx_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    account_number VARCHAR(50) UNIQUE, -- e.g. RTX-100001
    currency VARCHAR(10) DEFAULT 'USD',
    balance DECIMAL(18, 8) DEFAULT 0, -- Updated on realized P/L
    leverage INT DEFAULT 100, -- 1:100
    margin_mode VARCHAR(20) DEFAULT 'HEDGING', -- HEDGING/NETTING
    status VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE/SUSPENDED/CLOSED
    is_demo BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rtx_accounts_user ON rtx_accounts(user_id);

-- RTX Open Positions
CREATE TABLE rtx_positions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES rtx_accounts(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY/SELL
    volume DECIMAL(18, 8) NOT NULL, -- Lots
    open_price DECIMAL(18, 8) NOT NULL,
    current_price DECIMAL(18, 8), -- Cached for P/L
    open_time TIMESTAMPTZ DEFAULT NOW(),
    sl DECIMAL(18, 8),
    tp DECIMAL(18, 8),
    swap DECIMAL(18, 8) DEFAULT 0,
    commission DECIMAL(18, 8) DEFAULT 0,
    unrealized_pnl DECIMAL(18, 8) DEFAULT 0, -- Cached
    status VARCHAR(20) DEFAULT 'OPEN', -- OPEN/CLOSED
    close_price DECIMAL(18, 8),
    close_time TIMESTAMPTZ,
    close_reason VARCHAR(50) -- MANUAL/SL/TP/MARGIN_CALL
);

CREATE INDEX idx_rtx_positions_account ON rtx_positions(account_id, status);
CREATE INDEX idx_rtx_positions_symbol ON rtx_positions(symbol, status);

-- RTX Orders (Pending + History)
CREATE TABLE rtx_orders (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES rtx_accounts(id),
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL, -- MARKET/LIMIT/STOP/STOP_LIMIT
    side VARCHAR(10) NOT NULL, -- BUY/SELL
    volume DECIMAL(18, 8) NOT NULL,
    price DECIMAL(18, 8), -- For limit orders
    trigger_price DECIMAL(18, 8), -- For stop orders
    sl DECIMAL(18, 8),
    tp DECIMAL(18, 8),
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING/FILLED/CANCELLED/REJECTED/EXPIRED
    reject_reason TEXT,
    filled_price DECIMAL(18, 8),
    filled_at TIMESTAMPTZ,
    position_id BIGINT REFERENCES rtx_positions(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_rtx_orders_account ON rtx_orders(account_id, status);

-- RTX Trades (Execution Records)
CREATE TABLE rtx_trades (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT REFERENCES rtx_orders(id),
    position_id BIGINT REFERENCES rtx_positions(id),
    account_id BIGINT REFERENCES rtx_accounts(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY/SELL/CLOSE_BUY/CLOSE_SELL
    volume DECIMAL(18, 8) NOT NULL,
    price DECIMAL(18, 8) NOT NULL,
    realized_pnl DECIMAL(18, 8) DEFAULT 0, -- Set on closing trades
    commission DECIMAL(18, 8) DEFAULT 0,
    swap DECIMAL(18, 8) DEFAULT 0,
    executed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rtx_trades_account ON rtx_trades(account_id, executed_at DESC);

-- RTX Ledger (Double-Entry Accounting)
CREATE TABLE rtx_ledger (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT REFERENCES rtx_accounts(id),
    type VARCHAR(30) NOT NULL,
    -- DEPOSIT, WITHDRAW, REALIZED_PNL, COMMISSION, SWAP, ADJUSTMENT, BONUS, FEE
    amount DECIMAL(18, 8) NOT NULL, -- Positive=credit, Negative=debit
    balance_after DECIMAL(18, 8), -- Account balance after this entry
    currency VARCHAR(10) DEFAULT 'USD',
    description TEXT,
    ref_type VARCHAR(30), -- TRADE/POSITION/ADMIN/SYSTEM
    ref_id BIGINT, -- ID of trade/position/etc
    admin_id UUID, -- Admin who made adjustment
    payment_method VARCHAR(50), -- BANK/CRYPTO/CARD/MANUAL/BONUS
    payment_ref VARCHAR(100), -- External reference (bank txn, crypto hash)
    status VARCHAR(20) DEFAULT 'COMPLETED', -- PENDING/COMPLETED/FAILED/REVERSED
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_rtx_ledger_account ON rtx_ledger(account_id, created_at DESC);
CREATE INDEX idx_rtx_ledger_type ON rtx_ledger(type, status);

-- Symbol Specifications (for margin/pip calculation)
CREATE TABLE rtx_symbols (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    base_currency VARCHAR(10),
    quote_currency VARCHAR(10),
    contract_size DECIMAL(18, 8) DEFAULT 100000, -- Forex=100k, Gold=100, Crypto=1
    pip_size DECIMAL(18, 8) DEFAULT 0.0001, -- 0.01 for JPY pairs
    pip_value DECIMAL(18, 8) DEFAULT 10, -- Per lot in account currency
    min_volume DECIMAL(18, 8) DEFAULT 0.01,
    max_volume DECIMAL(18, 8) DEFAULT 100,
    volume_step DECIMAL(18, 8) DEFAULT 0.01,
    margin_percent DECIMAL(5, 2) DEFAULT 1.00, -- 1% = 100:1 leverage
    swap_long DECIMAL(18, 8) DEFAULT 0,
    swap_short DECIMAL(18, 8) DEFAULT 0,
    commission_per_lot DECIMAL(18, 8) DEFAULT 0,
    spread_markup DECIMAL(18, 8) DEFAULT 0, -- Additional spread
    is_enabled BOOLEAN DEFAULT TRUE
);

-- Insert default symbols
INSERT INTO rtx_symbols (symbol, base_currency, quote_currency, contract_size, pip_size, pip_value) VALUES
('EURUSD', 'EUR', 'USD', 100000, 0.0001, 10),
('GBPUSD', 'GBP', 'USD', 100000, 0.0001, 10),
('USDJPY', 'USD', 'JPY', 100000, 0.01, 9.09),
('XAUUSD', 'XAU', 'USD', 100, 0.1, 10),
('BTCUSD', 'BTC', 'USD', 1, 1, 1),
('ETHUSD', 'ETH', 'USD', 1, 0.1, 0.1)
ON CONFLICT (symbol) DO NOTHING;
