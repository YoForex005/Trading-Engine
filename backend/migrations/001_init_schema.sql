-- Migration: 001_init_schema
-- Description: Initial database schema for trading engine
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For text search
CREATE EXTENSION IF NOT EXISTS "btree_gist"; -- For advanced indexing

-- ============================================================================
-- Users and Authentication
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    country_code CHAR(2),
    kyc_status VARCHAR(20) DEFAULT 'pending' CHECK (kyc_status IN ('pending', 'approved', 'rejected', 'under_review')),
    kyc_level INTEGER DEFAULT 0 CHECK (kyc_level BETWEEN 0 AND 3),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    email_verified_at TIMESTAMPTZ,
    two_factor_enabled BOOLEAN DEFAULT false,
    two_factor_secret VARCHAR(255),
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'trader', 'admin', 'super_admin', 'risk_manager', 'compliance_officer')),
    granted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ,
    UNIQUE(user_id, role)
);

-- ============================================================================
-- Accounts and Balances
-- ============================================================================

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    account_type VARCHAR(20) NOT NULL CHECK (account_type IN ('demo', 'live', 'margin', 'islamic')),
    currency CHAR(3) DEFAULT 'USD',
    balance DECIMAL(20, 8) DEFAULT 0.00 CHECK (balance >= 0),
    equity DECIMAL(20, 8) DEFAULT 0.00,
    margin_used DECIMAL(20, 8) DEFAULT 0.00,
    margin_available DECIMAL(20, 8) DEFAULT 0.00,
    margin_level DECIMAL(10, 2),
    unrealized_pnl DECIMAL(20, 8) DEFAULT 0.00,
    realized_pnl DECIMAL(20, 8) DEFAULT 0.00,
    leverage INTEGER DEFAULT 1 CHECK (leverage > 0 AND leverage <= 500),
    max_leverage INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    is_locked BOOLEAN DEFAULT false,
    locked_reason TEXT,
    credit DECIMAL(20, 8) DEFAULT 0.00,
    broker_id VARCHAR(100),
    group_name VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS account_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    transaction_type VARCHAR(30) NOT NULL CHECK (transaction_type IN ('deposit', 'withdrawal', 'credit', 'debit', 'transfer', 'commission', 'swap', 'bonus', 'adjustment')),
    amount DECIMAL(20, 8) NOT NULL,
    currency CHAR(3) NOT NULL,
    balance_before DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    reference_id UUID,
    reference_type VARCHAR(50),
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMPTZ,
    metadata JSONB
);

-- ============================================================================
-- Orders
-- ============================================================================

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_order_id VARCHAR(100),
    symbol VARCHAR(20) NOT NULL,
    order_type VARCHAR(20) NOT NULL CHECK (order_type IN ('market', 'limit', 'stop', 'stop_limit', 'trailing_stop', 'take_profit')),
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(20, 8) NOT NULL CHECK (quantity > 0),
    filled_quantity DECIMAL(20, 8) DEFAULT 0.00,
    remaining_quantity DECIMAL(20, 8),
    price DECIMAL(20, 8),
    stop_price DECIMAL(20, 8),
    limit_price DECIMAL(20, 8),
    trailing_distance DECIMAL(20, 8),
    average_fill_price DECIMAL(20, 8),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'pending_new', 'new', 'partially_filled', 'filled', 'cancelled', 'rejected', 'expired')),
    time_in_force VARCHAR(10) DEFAULT 'GTC' CHECK (time_in_force IN ('GTC', 'IOC', 'FOK', 'DAY', 'GTD')),
    good_till_date TIMESTAMPTZ,
    execution_type VARCHAR(20) CHECK (execution_type IN ('market', 'abook', 'bbook', 'hybrid')),
    routing_decision VARCHAR(20),
    take_profit DECIMAL(20, 8),
    stop_loss DECIMAL(20, 8),
    is_reduce_only BOOLEAN DEFAULT false,
    is_post_only BOOLEAN DEFAULT false,
    commission DECIMAL(20, 8) DEFAULT 0.00,
    commission_currency CHAR(3),
    slippage DECIMAL(20, 8),
    reject_reason TEXT,
    lp_provider VARCHAR(100),
    lp_order_id VARCHAR(100),
    parent_order_id UUID REFERENCES orders(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    submitted_at TIMESTAMPTZ,
    filled_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS order_fills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    fill_price DECIMAL(20, 8) NOT NULL,
    fill_quantity DECIMAL(20, 8) NOT NULL,
    fill_commission DECIMAL(20, 8) DEFAULT 0.00,
    liquidity_type VARCHAR(10) CHECK (liquidity_type IN ('maker', 'taker')),
    venue VARCHAR(100),
    execution_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Positions
-- ============================================================================

CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('long', 'short')),
    quantity DECIMAL(20, 8) NOT NULL CHECK (quantity >= 0),
    entry_price DECIMAL(20, 8) NOT NULL,
    current_price DECIMAL(20, 8),
    mark_price DECIMAL(20, 8),
    average_price DECIMAL(20, 8),
    liquidation_price DECIMAL(20, 8),
    margin_required DECIMAL(20, 8),
    unrealized_pnl DECIMAL(20, 8) DEFAULT 0.00,
    realized_pnl DECIMAL(20, 8) DEFAULT 0.00,
    commission_paid DECIMAL(20, 8) DEFAULT 0.00,
    swap_charges DECIMAL(20, 8) DEFAULT 0.00,
    stop_loss DECIMAL(20, 8),
    take_profit DECIMAL(20, 8),
    trailing_stop DECIMAL(20, 8),
    is_open BOOLEAN DEFAULT true,
    open_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    close_time TIMESTAMPTZ,
    last_updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    UNIQUE(account_id, symbol, side, is_open) WHERE is_open = true
);

-- ============================================================================
-- Trades (Execution History)
-- ============================================================================

CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position_id UUID REFERENCES positions(id),
    order_id UUID REFERENCES orders(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL CHECK (side IN ('buy', 'sell')),
    quantity DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    value DECIMAL(20, 8) NOT NULL,
    commission DECIMAL(20, 8) DEFAULT 0.00,
    swap DECIMAL(20, 8) DEFAULT 0.00,
    pnl DECIMAL(20, 8),
    execution_type VARCHAR(20) CHECK (execution_type IN ('market', 'abook', 'bbook', 'hybrid')),
    liquidity_type VARCHAR(10) CHECK (liquidity_type IN ('maker', 'taker')),
    lp_provider VARCHAR(100),
    venue VARCHAR(100),
    execution_id VARCHAR(100) UNIQUE,
    is_opening BOOLEAN DEFAULT true,
    executed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Instruments and Market Data
-- ============================================================================

CREATE TABLE IF NOT EXISTS instruments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(20) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    instrument_type VARCHAR(20) CHECK (instrument_type IN ('forex', 'crypto', 'cfd', 'commodity', 'index', 'stock')),
    base_currency CHAR(3),
    quote_currency CHAR(3),
    tick_size DECIMAL(20, 8),
    min_quantity DECIMAL(20, 8),
    max_quantity DECIMAL(20, 8),
    quantity_step DECIMAL(20, 8),
    contract_size DECIMAL(20, 8) DEFAULT 1.00,
    leverage_max INTEGER DEFAULT 100,
    margin_rate DECIMAL(10, 6),
    commission_rate DECIMAL(10, 6),
    swap_long DECIMAL(20, 8),
    swap_short DECIMAL(20, 8),
    is_tradable BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    trading_hours JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Triggers for updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_accounts_updated_at BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON positions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_instruments_updated_at BEFORE UPDATE ON instruments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_account_transactions_updated_at BEFORE UPDATE ON account_transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
DROP TRIGGER IF EXISTS update_account_transactions_updated_at ON account_transactions;
DROP TRIGGER IF EXISTS update_instruments_updated_at ON instruments;
DROP TRIGGER IF EXISTS update_positions_updated_at ON positions;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS trades CASCADE;
DROP TABLE IF EXISTS positions CASCADE;
DROP TABLE IF EXISTS order_fills CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS account_transactions CASCADE;
DROP TABLE IF EXISTS accounts CASCADE;
DROP TABLE IF EXISTS instruments CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;
DROP TABLE IF EXISTS users CASCADE;

DROP EXTENSION IF EXISTS btree_gist;
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS "uuid-ossp";
*/
