-- Migration: 007_add_routing_rules
-- Description: Routing rules, exposure limits, and audit tables for order routing engine
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- ============================================================================
-- Routing Rules
-- ============================================================================

CREATE TABLE IF NOT EXISTS routing_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    priority INTEGER NOT NULL DEFAULT 100,
    rule_name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    filters JSONB NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('route_abook', 'route_bbook', 'route_hybrid', 'reject', 'hold', 'alert')),
    target_lp UUID REFERENCES liquidity_providers(id) ON DELETE SET NULL,
    hedge_percent DECIMAL(5, 2) DEFAULT 0.00 CHECK (hedge_percent BETWEEN 0 AND 100),
    is_active BOOLEAN DEFAULT true,
    condition_type VARCHAR(20) DEFAULT 'and' CHECK (condition_type IN ('and', 'or', 'complex')),
    min_order_size DECIMAL(20, 8),
    max_order_size DECIMAL(20, 8),
    symbols TEXT[],
    account_ids UUID[],
    user_ids UUID[],
    trading_hours JSONB,
    risk_limits JSONB,
    metadata JSONB,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_rules_active ON routing_rules(priority DESC) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_rules_target_lp ON routing_rules(target_lp) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_rules_action ON routing_rules(action) WHERE is_active = true;

-- ============================================================================
-- Exposure Limits
-- ============================================================================

CREATE TABLE IF NOT EXISTS exposure_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(20) NOT NULL,
    scope VARCHAR(20) NOT NULL CHECK (scope IN ('global', 'lp', 'account', 'user', 'instrument')),
    scope_id UUID,
    max_net_exposure DECIMAL(20, 8) NOT NULL,
    max_gross_exposure DECIMAL(20, 8) NOT NULL,
    auto_hedge_level DECIMAL(5, 2) CHECK (auto_hedge_level BETWEEN 0 AND 100),
    auto_hedge_enabled BOOLEAN DEFAULT false,
    warning_threshold DECIMAL(5, 2) DEFAULT 80.00 CHECK (warning_threshold BETWEEN 0 AND 100),
    action_on_breach VARCHAR(20) DEFAULT 'alert' CHECK (action_on_breach IN ('alert', 'warn', 'hedge', 'reduce', 'reject')),
    current_net_exposure DECIMAL(20, 8) DEFAULT 0.00,
    current_gross_exposure DECIMAL(20, 8) DEFAULT 0.00,
    utilization_percentage DECIMAL(5, 2) DEFAULT 0.00,
    breach_count INTEGER DEFAULT 0,
    last_breach_at TIMESTAMPTZ,
    last_hedge_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    UNIQUE(symbol, scope, scope_id)
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exposure_limits_symbol ON exposure_limits(symbol) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exposure_limits_scope ON exposure_limits(scope, scope_id) WHERE is_active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exposure_limits_breach ON exposure_limits(utilization_percentage DESC) WHERE utilization_percentage > 80;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_exposure_limits_active ON exposure_limits(is_active, symbol);

-- ============================================================================
-- Routing Audit Trail
-- ============================================================================

CREATE TABLE IF NOT EXISTS routing_audit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('create', 'update', 'delete', 'activate', 'deactivate', 'manual_override')),
    entity_type VARCHAR(50) NOT NULL CHECK (entity_type IN ('routing_rule', 'exposure_limit', 'exposure_limit_breach', 'hedge_action')),
    entity_id UUID NOT NULL,
    rule_id UUID REFERENCES routing_rules(id) ON DELETE SET NULL,
    exposure_limit_id UUID REFERENCES exposure_limits(id) ON DELETE SET NULL,
    old_values JSONB,
    new_values JSONB,
    reason TEXT,
    impact_summary TEXT,
    affected_orders INTEGER,
    affected_positions INTEGER,
    changed_by UUID REFERENCES users(id),
    change_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_audit_entity ON routing_audit(entity_type, entity_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_audit_rule_id ON routing_audit(rule_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_audit_exposure_id ON routing_audit(exposure_limit_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_audit_change_type ON routing_audit(change_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_audit_timestamp ON routing_audit(change_timestamp DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_audit_changed_by ON routing_audit(changed_by);

-- ============================================================================
-- Exposure Limit Breach Events
-- ============================================================================

CREATE TABLE IF NOT EXISTS exposure_limit_breaches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    exposure_limit_id UUID NOT NULL REFERENCES exposure_limits(id) ON DELETE CASCADE,
    breach_type VARCHAR(20) NOT NULL CHECK (breach_type IN ('net_exposure', 'gross_exposure')),
    limit_value DECIMAL(20, 8) NOT NULL,
    actual_value DECIMAL(20, 8) NOT NULL,
    breach_percentage DECIMAL(5, 2) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('warning', 'critical')),
    trigger_order_id UUID REFERENCES orders(id),
    auto_action_taken VARCHAR(50),
    manual_override BOOLEAN DEFAULT false,
    override_reason TEXT,
    override_by UUID REFERENCES users(id),
    resolved BOOLEAN DEFAULT false,
    resolution_action VARCHAR(50),
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_breaches_exposure_limit_id ON exposure_limit_breaches(exposure_limit_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_breaches_resolved ON exposure_limit_breaches(resolved) WHERE resolved = false;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_breaches_severity ON exposure_limit_breaches(severity, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_breaches_created_at ON exposure_limit_breaches(created_at DESC);

-- ============================================================================
-- Hedge Actions
-- ============================================================================

CREATE TABLE IF NOT EXISTS hedge_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    exposure_limit_id UUID NOT NULL REFERENCES exposure_limits(id) ON DELETE CASCADE,
    hedge_type VARCHAR(20) NOT NULL CHECK (hedge_type IN ('auto', 'manual', 'emergency')),
    trigger_event VARCHAR(100),
    target_hedge_level DECIMAL(5, 2),
    hedge_instrument VARCHAR(20),
    hedge_direction VARCHAR(10) CHECK (hedge_direction IN ('long', 'short')),
    target_quantity DECIMAL(20, 8),
    actual_quantity DECIMAL(20, 8),
    hedge_price DECIMAL(20, 8),
    total_cost DECIMAL(20, 8),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'partially_completed', 'failed', 'cancelled')),
    initiated_by UUID REFERENCES users(id),
    initiated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,
    cancellation_reason TEXT,
    cancelled_at TIMESTAMPTZ,
    cancelled_by UUID REFERENCES users(id),
    error_message TEXT,
    metadata JSONB
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_hedge_actions_exposure_limit_id ON hedge_actions(exposure_limit_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_hedge_actions_status ON hedge_actions(status) WHERE status NOT IN ('completed', 'failed', 'cancelled');
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_hedge_actions_initiated_at ON hedge_actions(initiated_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_hedge_actions_type ON hedge_actions(hedge_type);

-- ============================================================================
-- Routing Performance Metrics
-- ============================================================================

CREATE TABLE IF NOT EXISTS routing_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID REFERENCES routing_rules(id) ON DELETE SET NULL,
    exposure_limit_id UUID REFERENCES exposure_limits(id) ON DELETE SET NULL,
    time_bucket TIMESTAMPTZ NOT NULL,
    bucket_interval VARCHAR(10) DEFAULT '1h' CHECK (bucket_interval IN ('1m', '5m', '15m', '1h', '1d')),
    total_orders_routed INTEGER DEFAULT 0,
    successful_routes INTEGER DEFAULT 0,
    failed_routes INTEGER DEFAULT 0,
    rejected_routes INTEGER DEFAULT 0,
    avg_routing_latency_ms DECIMAL(10, 2),
    total_exposure_managed DECIMAL(20, 8) DEFAULT 0.00,
    hedge_actions_triggered INTEGER DEFAULT 0,
    breach_alerts_generated INTEGER DEFAULT 0,
    effectiveness_score DECIMAL(5, 2),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rule_id, exposure_limit_id, time_bucket, bucket_interval)
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_perf_rule_id ON routing_performance(rule_id, time_bucket DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_perf_exposure_id ON routing_performance(exposure_limit_id, time_bucket DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_perf_time_bucket ON routing_performance(time_bucket DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_routing_perf_effectiveness ON routing_performance(effectiveness_score DESC);

-- ============================================================================
-- Triggers
-- ============================================================================

CREATE TRIGGER update_routing_rules_updated_at BEFORE UPDATE ON routing_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_exposure_limits_updated_at BEFORE UPDATE ON exposure_limits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
-- Drop triggers
DROP TRIGGER IF EXISTS update_exposure_limits_updated_at ON exposure_limits;
DROP TRIGGER IF EXISTS update_routing_rules_updated_at ON routing_rules;

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_perf_effectiveness;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_perf_time_bucket;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_perf_exposure_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_perf_rule_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_hedge_actions_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_hedge_actions_initiated_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_hedge_actions_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_hedge_actions_exposure_limit_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_breaches_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_breaches_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_breaches_resolved;
DROP INDEX CONCURRENTLY IF EXISTS idx_breaches_exposure_limit_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_audit_changed_by;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_audit_timestamp;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_audit_change_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_audit_exposure_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_audit_rule_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_audit_entity;
DROP INDEX CONCURRENTLY IF EXISTS idx_exposure_limits_active;
DROP INDEX CONCURRENTLY IF EXISTS idx_exposure_limits_breach;
DROP INDEX CONCURRENTLY IF EXISTS idx_exposure_limits_scope;
DROP INDEX CONCURRENTLY IF EXISTS idx_exposure_limits_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_rules_action;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_rules_target_lp;
DROP INDEX CONCURRENTLY IF EXISTS idx_routing_rules_active;

-- Drop tables
DROP TABLE IF EXISTS routing_performance CASCADE;
DROP TABLE IF EXISTS hedge_actions CASCADE;
DROP TABLE IF EXISTS exposure_limit_breaches CASCADE;
DROP TABLE IF EXISTS routing_audit CASCADE;
DROP TABLE IF EXISTS exposure_limits CASCADE;
DROP TABLE IF EXISTS routing_rules CASCADE;
*/
