-- Migration: 004_add_lp_tables
-- Description: Liquidity provider and routing tables
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- ============================================================================
-- Liquidity Providers
-- ============================================================================

CREATE TABLE IF NOT EXISTS liquidity_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    provider_type VARCHAR(20) NOT NULL CHECK (provider_type IN ('prime_broker', 'exchange', 'market_maker', 'aggregator')),
    protocol VARCHAR(20) CHECK (protocol IN ('FIX', 'REST', 'WEBSOCKET', 'PROPRIETARY')),
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100,
    max_order_size DECIMAL(20, 8),
    min_order_size DECIMAL(20, 8),
    supported_instruments TEXT[],
    connection_string TEXT,
    api_key_encrypted TEXT,
    api_secret_encrypted TEXT,
    connection_status VARCHAR(20) DEFAULT 'disconnected' CHECK (connection_status IN ('connected', 'disconnected', 'error', 'maintenance')),
    last_heartbeat TIMESTAMPTZ,
    latency_ms INTEGER,
    uptime_percentage DECIMAL(5, 2),
    total_volume_24h DECIMAL(20, 8) DEFAULT 0.00,
    total_trades_24h INTEGER DEFAULT 0,
    config JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- LP Instrument Configuration
-- ============================================================================

CREATE TABLE IF NOT EXISTS lp_instruments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lp_id UUID NOT NULL REFERENCES liquidity_providers(id) ON DELETE CASCADE,
    instrument_id UUID NOT NULL REFERENCES instruments(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    lp_symbol VARCHAR(50),
    is_enabled BOOLEAN DEFAULT true,
    min_quantity DECIMAL(20, 8),
    max_quantity DECIMAL(20, 8),
    tick_size DECIMAL(20, 8),
    commission_rate DECIMAL(10, 6),
    markup_bps INTEGER DEFAULT 0,
    spread_markup_bps INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 100,
    max_slippage_bps INTEGER,
    trading_hours JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(lp_id, instrument_id)
);

-- ============================================================================
-- LP Quotes and Pricing
-- ============================================================================

CREATE TABLE IF NOT EXISTS lp_quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lp_id UUID NOT NULL REFERENCES liquidity_providers(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(20, 8) NOT NULL,
    ask DECIMAL(20, 8) NOT NULL,
    bid_size DECIMAL(20, 8),
    ask_size DECIMAL(20, 8),
    spread_bps INTEGER,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    sequence_number BIGINT,
    is_stale BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Hypertable for time-series optimization (if using TimescaleDB)
-- SELECT create_hypertable('lp_quotes', 'timestamp', if_not_exists => TRUE);

-- ============================================================================
-- LP Order Routing
-- ============================================================================

CREATE TABLE IF NOT EXISTS lp_order_routing (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    lp_id UUID NOT NULL REFERENCES liquidity_providers(id),
    routing_decision VARCHAR(20) NOT NULL CHECK (routing_decision IN ('abook', 'bbook', 'hybrid', 'rejected')),
    routing_score DECIMAL(10, 4),
    routing_reason TEXT,
    expected_fill_price DECIMAL(20, 8),
    expected_slippage DECIMAL(20, 8),
    expected_commission DECIMAL(20, 8),
    actual_fill_price DECIMAL(20, 8),
    actual_slippage DECIMAL(20, 8),
    actual_commission DECIMAL(20, 8),
    routing_latency_ms INTEGER,
    execution_latency_ms INTEGER,
    success BOOLEAN,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    routed_at TIMESTAMPTZ,
    executed_at TIMESTAMPTZ,
    metadata JSONB
);

-- ============================================================================
-- LP Performance Metrics
-- ============================================================================

CREATE TABLE IF NOT EXISTS lp_performance_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lp_id UUID NOT NULL REFERENCES liquidity_providers(id) ON DELETE CASCADE,
    symbol VARCHAR(20),
    time_bucket TIMESTAMPTZ NOT NULL,
    bucket_interval VARCHAR(10) DEFAULT '1h' CHECK (bucket_interval IN ('1m', '5m', '15m', '1h', '1d')),
    total_orders INTEGER DEFAULT 0,
    successful_orders INTEGER DEFAULT 0,
    failed_orders INTEGER DEFAULT 0,
    rejected_orders INTEGER DEFAULT 0,
    total_volume DECIMAL(20, 8) DEFAULT 0.00,
    avg_fill_time_ms INTEGER,
    avg_slippage_bps INTEGER,
    avg_spread_bps INTEGER,
    best_bid DECIMAL(20, 8),
    best_ask DECIMAL(20, 8),
    quote_count INTEGER DEFAULT 0,
    uptime_seconds INTEGER DEFAULT 0,
    downtime_seconds INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(lp_id, symbol, time_bucket, bucket_interval)
);

-- ============================================================================
-- LP Balance and Exposure
-- ============================================================================

CREATE TABLE IF NOT EXISTS lp_exposure (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lp_id UUID NOT NULL REFERENCES liquidity_providers(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    net_position DECIMAL(20, 8) DEFAULT 0.00,
    long_positions DECIMAL(20, 8) DEFAULT 0.00,
    short_positions DECIMAL(20, 8) DEFAULT 0.00,
    unrealized_pnl DECIMAL(20, 8) DEFAULT 0.00,
    realized_pnl_24h DECIMAL(20, 8) DEFAULT 0.00,
    total_volume_24h DECIMAL(20, 8) DEFAULT 0.00,
    exposure_limit DECIMAL(20, 8),
    risk_utilization DECIMAL(5, 2),
    last_updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    UNIQUE(lp_id, symbol)
);

-- ============================================================================
-- Smart Order Routing (SOR) Rules
-- ============================================================================

CREATE TABLE IF NOT EXISTS sor_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_name VARCHAR(100) UNIQUE NOT NULL,
    rule_type VARCHAR(50) NOT NULL CHECK (rule_type IN ('abook', 'bbook', 'hybrid', 'conditional')),
    priority INTEGER DEFAULT 100,
    is_active BOOLEAN DEFAULT true,
    conditions JSONB NOT NULL,
    actions JSONB NOT NULL,
    fallback_action VARCHAR(20) CHECK (fallback_action IN ('abook', 'bbook', 'reject')),
    valid_from TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMPTZ,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Client Profiling (for intelligent routing)
-- ============================================================================

CREATE TABLE IF NOT EXISTS client_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    profile_type VARCHAR(20) DEFAULT 'neutral' CHECK (profile_type IN ('toxic', 'neutral', 'profitable', 'vip')),
    toxicity_score DECIMAL(5, 2) DEFAULT 50.00 CHECK (toxicity_score BETWEEN 0 AND 100),
    profitability_score DECIMAL(5, 2) DEFAULT 50.00,
    win_rate DECIMAL(5, 2),
    avg_trade_duration_seconds INTEGER,
    avg_trade_size DECIMAL(20, 8),
    preferred_instruments TEXT[],
    trading_style VARCHAR(20) CHECK (trading_style IN ('scalper', 'day_trader', 'swing_trader', 'position_trader')),
    total_trades INTEGER DEFAULT 0,
    total_volume DECIMAL(20, 8) DEFAULT 0.00,
    total_pnl DECIMAL(20, 8) DEFAULT 0.00,
    total_commission_paid DECIMAL(20, 8) DEFAULT 0.00,
    abook_percentage INTEGER DEFAULT 50,
    bbook_percentage INTEGER DEFAULT 50,
    risk_category VARCHAR(20) DEFAULT 'medium' CHECK (risk_category IN ('low', 'medium', 'high', 'extreme')),
    auto_route BOOLEAN DEFAULT true,
    manual_override_route VARCHAR(10) CHECK (manual_override_route IN ('abook', 'bbook', 'auto')),
    last_analyzed TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    UNIQUE(account_id)
);

-- ============================================================================
-- Indexes for LP Tables
-- ============================================================================

-- liquidity_providers indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_name ON liquidity_providers(name) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_is_active ON liquidity_providers(is_active);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_connection_status ON liquidity_providers(connection_status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_priority ON liquidity_providers(priority DESC) WHERE is_active = true;

-- lp_instruments indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_instruments_lp_id ON lp_instruments(lp_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_instruments_instrument_id ON lp_instruments(instrument_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_instruments_symbol ON lp_instruments(symbol) WHERE is_enabled = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_instruments_enabled ON lp_instruments(lp_id, is_enabled);

-- lp_quotes indexes (critical for performance)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_quotes_lp_symbol ON lp_quotes(lp_id, symbol, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_quotes_symbol_timestamp ON lp_quotes(symbol, timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_quotes_timestamp ON lp_quotes(timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_quotes_not_stale ON lp_quotes(symbol, timestamp DESC) WHERE is_stale = false;

-- lp_order_routing indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_order_routing_order_id ON lp_order_routing(order_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_order_routing_lp_id ON lp_order_routing(lp_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_order_routing_decision ON lp_order_routing(routing_decision);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_order_routing_created_at ON lp_order_routing(created_at DESC);

-- lp_performance_metrics indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_perf_lp_symbol_time ON lp_performance_metrics(lp_id, symbol, time_bucket DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_perf_time_bucket ON lp_performance_metrics(time_bucket DESC);

-- lp_exposure indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_exposure_lp_id ON lp_exposure(lp_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_exposure_symbol ON lp_exposure(symbol);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_exposure_risk ON lp_exposure(risk_utilization DESC) WHERE risk_utilization > 80;

-- sor_rules indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sor_rules_active ON sor_rules(priority DESC) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sor_rules_type ON sor_rules(rule_type) WHERE is_active = true;

-- client_profiles indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_client_profiles_user_id ON client_profiles(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_client_profiles_account_id ON client_profiles(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_client_profiles_type ON client_profiles(profile_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_client_profiles_toxicity ON client_profiles(toxicity_score DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_client_profiles_profitability ON client_profiles(profitability_score DESC);

-- ============================================================================
-- Triggers
-- ============================================================================

CREATE TRIGGER update_lp_updated_at BEFORE UPDATE ON liquidity_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_lp_instruments_updated_at BEFORE UPDATE ON lp_instruments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sor_rules_updated_at BEFORE UPDATE ON sor_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_client_profiles_updated_at BEFORE UPDATE ON client_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
-- Drop triggers
DROP TRIGGER IF EXISTS update_client_profiles_updated_at ON client_profiles;
DROP TRIGGER IF EXISTS update_sor_rules_updated_at ON sor_rules;
DROP TRIGGER IF EXISTS update_lp_instruments_updated_at ON lp_instruments;
DROP TRIGGER IF EXISTS update_lp_updated_at ON liquidity_providers;

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_client_profiles_profitability;
DROP INDEX CONCURRENTLY IF EXISTS idx_client_profiles_toxicity;
DROP INDEX CONCURRENTLY IF EXISTS idx_client_profiles_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_client_profiles_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_client_profiles_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_sor_rules_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_sor_rules_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_exposure_risk;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_exposure_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_exposure_lp_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_perf_time_bucket;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_perf_lp_symbol_time;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_order_routing_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_order_routing_decision;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_order_routing_lp_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_order_routing_order_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_quotes_not_stale;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_quotes_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_quotes_symbol_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_quotes_lp_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_instruments_enabled;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_instruments_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_instruments_instrument_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_instruments_lp_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_priority;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_connection_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_is_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_name;

-- Drop tables
DROP TABLE IF EXISTS client_profiles CASCADE;
DROP TABLE IF EXISTS sor_rules CASCADE;
DROP TABLE IF EXISTS lp_exposure CASCADE;
DROP TABLE IF EXISTS lp_performance_metrics CASCADE;
DROP TABLE IF EXISTS lp_order_routing CASCADE;
DROP TABLE IF EXISTS lp_quotes CASCADE;
DROP TABLE IF EXISTS lp_instruments CASCADE;
DROP TABLE IF EXISTS liquidity_providers CASCADE;
*/
