-- Migration: 003_add_audit_tables
-- Description: Audit trail and compliance tables
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- ============================================================================
-- Audit Log (Generic Audit Trail)
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('INSERT', 'UPDATE', 'DELETE', 'SELECT')),
    user_id UUID REFERENCES users(id),
    user_ip INET,
    user_agent TEXT,
    old_data JSONB,
    new_data JSONB,
    changed_fields TEXT[],
    query TEXT,
    session_id VARCHAR(100),
    request_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- Partition audit_log by month for performance
CREATE TABLE IF NOT EXISTS audit_log_2026_01 PARTITION OF audit_log
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- ============================================================================
-- User Activity Log
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL CHECK (activity_type IN (
        'login', 'logout', 'login_failed', 'password_change', 'password_reset',
        'email_change', '2fa_enabled', '2fa_disabled', 'kyc_submitted',
        'profile_update', 'settings_change', 'api_key_created', 'api_key_revoked'
    )),
    ip_address INET,
    user_agent TEXT,
    device_fingerprint VARCHAR(255),
    location_country CHAR(2),
    location_city VARCHAR(100),
    session_id VARCHAR(100),
    success BOOLEAN DEFAULT true,
    failure_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Order Audit Trail
-- ============================================================================

CREATE TABLE IF NOT EXISTS order_audit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'created', 'submitted', 'acknowledged', 'partially_filled', 'filled',
        'cancelled', 'rejected', 'expired', 'amended', 'triggered'
    )),
    previous_status VARCHAR(20),
    new_status VARCHAR(20),
    previous_data JSONB,
    new_data JSONB,
    triggered_by VARCHAR(50),
    reason TEXT,
    lp_provider VARCHAR(100),
    execution_venue VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Position Audit Trail
-- ============================================================================

CREATE TABLE IF NOT EXISTS position_audit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'opened', 'increased', 'decreased', 'closed', 'modified',
        'stop_loss_hit', 'take_profit_hit', 'liquidated', 'margin_call'
    )),
    previous_quantity DECIMAL(20, 8),
    new_quantity DECIMAL(20, 8),
    previous_price DECIMAL(20, 8),
    new_price DECIMAL(20, 8),
    pnl_change DECIMAL(20, 8),
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Account Audit Trail
-- ============================================================================

CREATE TABLE IF NOT EXISTS account_audit (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'created', 'activated', 'deactivated', 'locked', 'unlocked',
        'leverage_changed', 'margin_call', 'liquidation', 'balance_adjusted',
        'settings_changed', 'closed'
    )),
    previous_data JSONB,
    new_data JSONB,
    changed_by UUID REFERENCES users(id),
    reason TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Risk Events Log
-- ============================================================================

CREATE TABLE IF NOT EXISTS risk_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'margin_call', 'stop_out', 'liquidation', 'position_limit_breach',
        'daily_loss_limit', 'concentration_risk', 'leverage_breach',
        'exposure_limit', 'circuit_breaker_triggered', 'market_manipulation_alert'
    )),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    account_id UUID REFERENCES accounts(id),
    user_id UUID REFERENCES users(id),
    position_id UUID REFERENCES positions(id),
    order_id UUID REFERENCES orders(id),
    symbol VARCHAR(20),
    risk_metric VARCHAR(50),
    threshold_value DECIMAL(20, 8),
    actual_value DECIMAL(20, 8),
    action_taken VARCHAR(100),
    auto_resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- Compliance Events
-- ============================================================================

CREATE TABLE IF NOT EXISTS compliance_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'kyc_check', 'aml_screening', 'sanctions_check', 'pep_check',
        'transaction_monitoring', 'unusual_activity', 'regulatory_report',
        'best_execution_review', 'trade_reconstruction', 'mifid_report'
    )),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'alert', 'critical')),
    user_id UUID REFERENCES users(id),
    account_id UUID REFERENCES accounts(id),
    order_id UUID REFERENCES orders(id),
    trade_id UUID REFERENCES trades(id),
    check_result VARCHAR(20) CHECK (check_result IN ('pass', 'fail', 'review_required', 'pending')),
    risk_score DECIMAL(5, 2),
    flagged_reason TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMPTZ,
    review_notes TEXT,
    reported_to_regulator BOOLEAN DEFAULT false,
    report_reference VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- System Events and Errors
-- ============================================================================

CREATE TABLE IF NOT EXISTS system_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN (
        'server_start', 'server_stop', 'server_error', 'database_error',
        'lp_connection_lost', 'lp_connection_restored', 'api_error',
        'websocket_error', 'data_pipeline_error', 'backup_completed',
        'migration_completed', 'config_changed', 'security_alert'
    )),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('debug', 'info', 'warning', 'error', 'critical')),
    component VARCHAR(100),
    message TEXT NOT NULL,
    stack_trace TEXT,
    error_code VARCHAR(50),
    user_id UUID REFERENCES users(id),
    ip_address INET,
    request_id VARCHAR(100),
    session_id VARCHAR(100),
    resolved BOOLEAN DEFAULT false,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- ============================================================================
-- API Access Log
-- ============================================================================

CREATE TABLE IF NOT EXISTS api_access_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    api_key_id UUID,
    endpoint VARCHAR(255) NOT NULL,
    http_method VARCHAR(10) NOT NULL,
    request_body JSONB,
    response_status INTEGER,
    response_time_ms INTEGER,
    ip_address INET,
    user_agent TEXT,
    error_message TEXT,
    rate_limit_remaining INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- Indexes for Audit Tables
-- ============================================================================

-- audit_log indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_log_table_record ON audit_log(table_name, record_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_log_session_id ON audit_log(session_id) WHERE session_id IS NOT NULL;

-- user_activity_log indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activity_user_id ON user_activity_log(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activity_type ON user_activity_log(activity_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activity_created_at ON user_activity_log(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activity_ip ON user_activity_log(ip_address);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activity_session ON user_activity_log(session_id) WHERE session_id IS NOT NULL;

-- order_audit indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_audit_order_id ON order_audit(order_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_audit_account_id ON order_audit(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_audit_user_id ON order_audit(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_audit_event_type ON order_audit(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_audit_created_at ON order_audit(created_at DESC);

-- position_audit indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_audit_position_id ON position_audit(position_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_audit_account_id ON position_audit(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_audit_user_id ON position_audit(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_audit_event_type ON position_audit(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_audit_created_at ON position_audit(created_at DESC);

-- account_audit indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_audit_account_id ON account_audit(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_audit_user_id ON account_audit(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_audit_event_type ON account_audit(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_account_audit_created_at ON account_audit(created_at DESC);

-- risk_events indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_events_event_type ON risk_events(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_events_severity ON risk_events(severity);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_events_account_id ON risk_events(account_id) WHERE account_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_events_user_id ON risk_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_events_created_at ON risk_events(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_risk_events_unresolved ON risk_events(created_at DESC) WHERE auto_resolved = false AND resolved_at IS NULL;

-- compliance_events indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_compliance_events_event_type ON compliance_events(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_compliance_events_severity ON compliance_events(severity);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_compliance_events_user_id ON compliance_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_compliance_events_result ON compliance_events(check_result);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_compliance_events_created_at ON compliance_events(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_compliance_events_pending_review ON compliance_events(created_at DESC) WHERE check_result = 'review_required' AND reviewed_at IS NULL;

-- system_events indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_system_events_event_type ON system_events(event_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_system_events_severity ON system_events(severity);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_system_events_component ON system_events(component);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_system_events_created_at ON system_events(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_system_events_unresolved ON system_events(created_at DESC) WHERE resolved = false;

-- api_access_log indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_access_user_id ON api_access_log(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_access_endpoint ON api_access_log(endpoint);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_access_created_at ON api_access_log(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_access_ip ON api_access_log(ip_address);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_api_access_status ON api_access_log(response_status);

-- ============================================================================
-- Audit Triggers (Automatic Audit Trail)
-- ============================================================================

-- Generic audit trigger function
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (table_name, record_id, action, old_data)
        VALUES (TG_TABLE_NAME, OLD.id, TG_OP, row_to_json(OLD));
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (table_name, record_id, action, old_data, new_data)
        VALUES (TG_TABLE_NAME, NEW.id, TG_OP, row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (table_name, record_id, action, new_data)
        VALUES (TG_TABLE_NAME, NEW.id, TG_OP, row_to_json(NEW));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Apply audit triggers to critical tables
CREATE TRIGGER audit_users AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_accounts AFTER INSERT OR UPDATE OR DELETE ON accounts
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_orders AFTER INSERT OR UPDATE OR DELETE ON orders
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_positions AFTER INSERT OR UPDATE OR DELETE ON positions
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_trades AFTER INSERT OR UPDATE OR DELETE ON trades
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
-- Drop triggers
DROP TRIGGER IF EXISTS audit_trades ON trades;
DROP TRIGGER IF EXISTS audit_positions ON positions;
DROP TRIGGER IF EXISTS audit_orders ON orders;
DROP TRIGGER IF EXISTS audit_accounts ON accounts;
DROP TRIGGER IF EXISTS audit_users ON users;

-- Drop function
DROP FUNCTION IF EXISTS audit_trigger_function();

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_api_access_status;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_access_ip;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_access_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_access_endpoint;
DROP INDEX CONCURRENTLY IF EXISTS idx_api_access_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_system_events_unresolved;
DROP INDEX CONCURRENTLY IF EXISTS idx_system_events_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_system_events_component;
DROP INDEX CONCURRENTLY IF EXISTS idx_system_events_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_system_events_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_compliance_events_pending_review;
DROP INDEX CONCURRENTLY IF EXISTS idx_compliance_events_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_compliance_events_result;
DROP INDEX CONCURRENTLY IF EXISTS idx_compliance_events_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_compliance_events_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_compliance_events_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_events_unresolved;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_events_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_events_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_events_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_events_severity;
DROP INDEX CONCURRENTLY IF EXISTS idx_risk_events_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_audit_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_audit_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_audit_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_account_audit_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_audit_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_audit_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_audit_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_audit_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_position_audit_position_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_audit_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_audit_event_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_audit_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_audit_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_order_audit_order_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_activity_session;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_activity_ip;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_activity_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_activity_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_user_activity_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_log_session_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_log_action;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_log_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_log_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_audit_log_table_record;

-- Drop tables
DROP TABLE IF EXISTS api_access_log CASCADE;
DROP TABLE IF EXISTS system_events CASCADE;
DROP TABLE IF EXISTS compliance_events CASCADE;
DROP TABLE IF EXISTS risk_events CASCADE;
DROP TABLE IF EXISTS account_audit CASCADE;
DROP TABLE IF EXISTS position_audit CASCADE;
DROP TABLE IF EXISTS order_audit CASCADE;
DROP TABLE IF EXISTS user_activity_log CASCADE;
DROP TABLE IF EXISTS audit_log_2026_01 CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;
*/
