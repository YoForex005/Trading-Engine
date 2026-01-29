-- Migration: 005_add_risk_tables
-- Description: Risk management and monitoring tables
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- ============================================================================
-- Risk Limits
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    limit_type VARCHAR(50) NOT NULL CHECK (limit_type IN (
        'account_leverage', 'position_size', 'daily_loss', 'max_drawdown',
        'exposure_limit', 'concentration_limit', 'order_size', 'open_positions'
    )),
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('global', 'user', 'account', 'instrument', 'group')),
    entity_id UUID,
    limit_value DECIMAL(20, 8) NOT NULL,
    warning_threshold DECIMAL(20, 8),
    currency CHAR(3),
    is_active BOOLEAN DEFAULT true,
    breached_action VARCHAR(20) DEFAULT 'reject' CHECK (breached_action IN ('reject', 'warn', 'reduce', 'close')),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Risk Limit Breaches
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_limit_breaches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    risk_limit_id UUID NOT NULL REFERENCES risk_limits(id),
    account_id UUID REFERENCES accounts(id),
    user_id UUID REFERENCES users(id),
    order_id UUID REFERENCES orders(id),
    position_id UUID REFERENCES positions(id),
    breach_type VARCHAR(50) NOT NULL,
    limit_value DECIMAL(20, 8) NOT NULL,
    actual_value DECIMAL(20, 8) NOT NULL,
    breach_percentage DECIMAL(5, 2),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('warning', 'breach', 'critical')),
    action_taken VARCHAR(50),
    auto_resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Margin Requirements
-- ============================================================================

CREATE TABLE IF NOT EXISTS margin_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    instrument_id UUID NOT NULL REFERENCES instruments(id),
    symbol VARCHAR(20) NOT NULL,
    position_size DECIMAL(20, 8) NOT NULL,
    initial_margin DECIMAL(20, 8) NOT NULL,
    maintenance_margin DECIMAL(20, 8) NOT NULL,
    margin_call_level DECIMAL(20, 8),
    stop_out_level DECIMAL(20, 8),
    leverage INTEGER NOT NULL,
    currency CHAR(3) NOT NULL,
    calculated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Margin Calls
-- ============================================================================

CREATE TABLE IF NOT EXISTS margin_calls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    user_id UUID NOT NULL REFERENCES users(id),
    call_type VARCHAR(20) NOT NULL CHECK (call_type IN ('margin_call', 'stop_out', 'liquidation')),
    margin_level DECIMAL(10, 2) NOT NULL,
    required_margin DECIMAL(20, 8) NOT NULL,
    available_margin DECIMAL(20, 8) NOT NULL,
    shortfall DECIMAL(20, 8) NOT NULL,
    positions_affected UUID[],
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'liquidated', 'cancelled')),
    notified_at TIMESTAMPTZ,
    notification_method VARCHAR(20),
    resolved_at TIMESTAMPTZ,
    resolution_method VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Position Risk Metrics
-- ============================================================================

CREATE TABLE IF NOT EXISTS position_risk_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    symbol VARCHAR(20) NOT NULL,
    value_at_risk DECIMAL(20, 8),
    conditional_var DECIMAL(20, 8),
    expected_shortfall DECIMAL(20, 8),
    volatility DECIMAL(10, 4),
    beta DECIMAL(10, 4),
    delta DECIMAL(10, 4),
    gamma DECIMAL(10, 4),
    theta DECIMAL(10, 4),
    vega DECIMAL(10, 4),
    max_drawdown DECIMAL(20, 8),
    current_drawdown DECIMAL(20, 8),
    unrealized_pnl_percentage DECIMAL(10, 2),
    distance_to_liquidation DECIMAL(20, 8),
    risk_score DECIMAL(5, 2),
    calculated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Portfolio Risk
-- ============================================================================

CREATE TABLE IF NOT EXISTS portfolio_risk (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    time_bucket TIMESTAMPTZ NOT NULL,
    total_exposure DECIMAL(20, 8),
    net_exposure DECIMAL(20, 8),
    gross_exposure DECIMAL(20, 8),
    portfolio_var DECIMAL(20, 8),
    portfolio_cvar DECIMAL(20, 8),
    correlation_risk DECIMAL(5, 2),
    concentration_risk DECIMAL(5, 2),
    liquidity_risk DECIMAL(5, 2),
    leverage_ratio DECIMAL(10, 2),
    sharpe_ratio DECIMAL(10, 4),
    sortino_ratio DECIMAL(10, 4),
    max_drawdown DECIMAL(20, 8),
    current_drawdown DECIMAL(20, 8),
    open_positions_count INTEGER,
    largest_position_percentage DECIMAL(5, 2),
    risk_score DECIMAL(5, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    UNIQUE(account_id, time_bucket)
);

-- ============================================================================
-- Circuit Breakers
-- ============================================================================

CREATE TABLE IF NOT EXISTS circuit_breakers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    breaker_name VARCHAR(100) UNIQUE NOT NULL,
    breaker_type VARCHAR(50) NOT NULL CHECK (breaker_type IN (
        'price_movement', 'volatility', 'volume_surge', 'loss_limit',
        'error_rate', 'manual', 'system_health'
    )),
    scope VARCHAR(20) NOT NULL CHECK (scope IN ('global', 'symbol', 'account', 'lp')),
    scope_id VARCHAR(100),
    trigger_condition JSONB NOT NULL,
    current_value DECIMAL(20, 8),
    threshold_value DECIMAL(20, 8),
    status VARCHAR(20) DEFAULT 'armed' CHECK (status IN ('armed', 'triggered', 'cooling_down', 'disabled')),
    triggered_at TIMESTAMPTZ,
    reset_at TIMESTAMPTZ,
    cooldown_period_seconds INTEGER DEFAULT 300,
    auto_reset BOOLEAN DEFAULT true,
    action VARCHAR(50) CHECK (action IN ('halt_trading', 'reject_orders', 'reduce_positions', 'alert_only')),
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Circuit Breaker Events
-- ============================================================================

CREATE TABLE IF NOT EXISTS circuit_breaker_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    breaker_id UUID NOT NULL REFERENCES circuit_breakers(id),
    event_type VARCHAR(20) NOT NULL CHECK (event_type IN ('triggered', 'reset', 'manual_override')),
    trigger_value DECIMAL(20, 8),
    threshold_value DECIMAL(20, 8),
    affected_orders UUID[],
    affected_positions UUID[],
    action_taken VARCHAR(100),
    triggered_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Risk Alerts
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN (
        'high_leverage', 'large_position', 'rapid_trading', 'unusual_pnl',
        'correlation_spike', 'liquidity_issue', 'system_anomaly', 'custom'
    )),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    account_id UUID REFERENCES accounts(id),
    user_id UUID REFERENCES users(id),
    symbol VARCHAR(20),
    alert_message TEXT NOT NULL,
    trigger_value DECIMAL(20, 8),
    threshold_value DECIMAL(20, 8),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'acknowledged', 'resolved', 'false_positive')),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,
    notification_sent BOOLEAN DEFAULT false,
    notification_channels TEXT[],
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Stress Test Scenarios
-- ============================================================================

CREATE TABLE IF NOT EXISTS stress_test_scenarios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scenario_name VARCHAR(100) UNIQUE NOT NULL,
    scenario_type VARCHAR(50) NOT NULL CHECK (scenario_type IN (
        'market_crash', 'volatility_spike', 'liquidity_crisis',
        'correlation_breakdown', 'black_swan', 'historical', 'custom'
    )),
    description TEXT,
    market_shocks JSONB NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Stress Test Results
-- ============================================================================

CREATE TABLE IF NOT EXISTS stress_test_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scenario_id UUID NOT NULL REFERENCES stress_test_scenarios(id),
    account_id UUID REFERENCES accounts(id),
    test_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    portfolio_value_before DECIMAL(20, 8),
    portfolio_value_after DECIMAL(20, 8),
    loss_amount DECIMAL(20, 8),
    loss_percentage DECIMAL(10, 2),
    margin_call_triggered BOOLEAN DEFAULT false,
    liquidation_triggered BOOLEAN DEFAULT false,
    positions_closed INTEGER,
    max_drawdown DECIMAL(20, 8),
    recovery_time_hours INTEGER,
    passed_test BOOLEAN,
    risk_score DECIMAL(5, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Indexes for Risk Tables
-- ============================================================================

-- risk_limits indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_limits_entity ON risk_limits(entity_type, entity_id) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_limits_type ON risk_limits(limit_type) WHERE is_active = true;

-- risk_limit_breaches indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_breaches_limit_id ON risk_limit_breaches(risk_limit_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_breaches_account_id ON risk_limit_breaches(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_breaches_severity ON risk_limit_breaches(severity, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_breaches_unresolved ON risk_limit_breaches(created_at DESC) WHERE auto_resolved = false AND resolved_at IS NULL;

-- margin_requirements indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_req_account_id ON margin_requirements(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_req_symbol ON margin_requirements(symbol);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_req_calculated_at ON margin_requirements(calculated_at DESC);

-- margin_calls indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_calls_account_id ON margin_calls(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_calls_user_id ON margin_calls(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_calls_status ON margin_calls(status, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_margin_calls_active ON margin_calls(created_at DESC) WHERE status = 'active';

-- position_risk_metrics indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_risk_position_id ON position_risk_metrics(position_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_risk_account_id ON position_risk_metrics(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_risk_score ON position_risk_metrics(risk_score DESC);

-- portfolio_risk indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_portfolio_risk_account_id ON portfolio_risk(account_id, time_bucket DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_portfolio_risk_score ON portfolio_risk(risk_score DESC);

-- circuit_breakers indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_circuit_breakers_status ON circuit_breakers(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_circuit_breakers_scope ON circuit_breakers(scope, scope_id) WHERE status IN ('armed', 'triggered');

-- circuit_breaker_events indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cb_events_breaker_id ON circuit_breaker_events(breaker_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_cb_events_created_at ON circuit_breaker_events(created_at DESC);

-- risk_alerts indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_alerts_account_id ON risk_alerts(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_alerts_severity ON risk_alerts(severity, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_alerts_status ON risk_alerts(status);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_alerts_active ON risk_alerts(created_at DESC) WHERE status = 'active';

-- stress_test_scenarios indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stress_scenarios_type ON stress_test_scenarios(scenario_type) WHERE is_active = true;

-- stress_test_results indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stress_results_scenario_id ON stress_test_results(scenario_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stress_results_account_id ON stress_test_results(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_stress_results_test_date ON stress_test_results(test_date DESC);

-- ============================================================================
-- Triggers
-- ============================================================================

CREATE TRIGGER update_risk_limits_updated_at BEFORE UPDATE ON risk_limits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_circuit_breakers_updated_at BEFORE UPDATE ON circuit_breakers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_stress_scenarios_updated_at BEFORE UPDATE ON stress_test_scenarios
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
-- Drop triggers
DROP TRIGGER IF EXISTS update_stress_scenarios_updated_at ON stress_test_scenarios;
DROP TRIGGER IF EXISTS update_circuit_breakers_updated_at ON circuit_breakers;
DROP TRIGGER IF EXISTS update_risk_limits_updated_at ON risk_limits;

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_stress_results_test_date;
DROP INDEX CONCURRENTLY IF EXISTS idx_stress_results_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_stress_results_scenario_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_stress_scenarios_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_alerts_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_alerts_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_alerts_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_alerts_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_cb_events_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_cb_events_breaker_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_circuit_breakers_scope;
DROP INDEX CONCURRENTLY IF EXISTS idx_circuit_breakers_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_portfolio_risk_score;
DROP INDEX CONCURRENTLY IF EXISTS idx_portfolio_risk_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_risk_score;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_risk_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_risk_position_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_calls_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_calls_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_calls_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_calls_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_req_calculated_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_req_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_margin_req_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_breaches_unresolved;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_breaches_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_breaches_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_breaches_limit_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_limits_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_limits_entity;

-- Drop tables
DROP TABLE IF EXISTS stress_test_results CASCADE;
DROP TABLE IF EXISTS stress_test_scenarios CASCADE;
DROP TABLE IF EXISTS risk_alerts CASCADE;
DROP TABLE IF EXISTS circuit_breaker_events CASCADE;
DROP TABLE IF EXISTS circuit_breakers CASCADE;
DROP TABLE IF EXISTS portfolio_risk CASCADE;
DROP TABLE IF EXISTS position_risk_metrics CASCADE;
DROP TABLE IF EXISTS margin_calls CASCADE;
DROP TABLE IF EXISTS margin_requirements CASCADE;
DROP TABLE IF EXISTS risk_limit_breaches CASCADE;
DROP TABLE IF EXISTS risk_limits CASCADE;
*/
