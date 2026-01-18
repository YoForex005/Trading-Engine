-- Migration: 007_add_config_and_activity_tables
-- Description: Add lp_subscriptions, config_versions, and user_activity_log tables for configuration management and audit tracking
-- Author: Trading Engine Team
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

-- ============================================================================
-- LP Subscriptions
-- ============================================================================

CREATE TABLE IF NOT EXISTS lp_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    lp_id UUID NOT NULL REFERENCES liquidity_providers(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(lp_id, symbol)
);

-- ============================================================================
-- Configuration Versions
-- ============================================================================

CREATE TABLE IF NOT EXISTS config_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    config_type VARCHAR(100) NOT NULL CHECK (config_type IN ('system', 'trading', 'risk', 'routing', 'lp', 'user', 'instrument')),
    version INTEGER NOT NULL DEFAULT 1,
    data JSONB NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(config_type, version)
);

-- ============================================================================
-- User Activity Log
-- ============================================================================

CREATE TABLE IF NOT EXISTS user_activity_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL CHECK (action IN (
        'login', 'logout', 'create_order', 'cancel_order', 'modify_order',
        'place_trade', 'close_position', 'update_settings', 'update_password',
        'enable_2fa', 'disable_2fa', 'change_role', 'access_report', 'export_data',
        'config_change', 'permission_grant', 'permission_revoke', 'api_key_create',
        'api_key_delete', 'webhook_create', 'webhook_delete', 'other'
    )),
    details JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- Indexes for Configuration and Activity Tables
-- ============================================================================

-- lp_subscriptions indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_subscriptions_lp_id ON lp_subscriptions(lp_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_subscriptions_symbol ON lp_subscriptions(symbol);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_subscriptions_enabled ON lp_subscriptions(enabled) WHERE enabled = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_lp_subscriptions_created_at ON lp_subscriptions(created_at DESC);

-- config_versions indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_config_versions_type ON config_versions(config_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_config_versions_type_version ON config_versions(config_type, version DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_config_versions_created_by ON config_versions(created_by);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_config_versions_created_at ON config_versions(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_config_versions_type_latest ON config_versions(config_type DESC, created_at DESC);

-- user_activity_log indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_user_id ON user_activity_log(user_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_account_id ON user_activity_log(account_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_action ON user_activity_log(action);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_created_at ON user_activity_log(created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_user_created ON user_activity_log(user_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_account_created ON user_activity_log(account_id, created_at DESC);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_ip_address ON user_activity_log(ip_address);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activity_log_action_created ON user_activity_log(action, created_at DESC);

-- ============================================================================
-- Triggers
-- ============================================================================

CREATE TRIGGER update_lp_subscriptions_updated_at BEFORE UPDATE ON lp_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
-- Drop triggers
DROP TRIGGER IF EXISTS update_lp_subscriptions_updated_at ON lp_subscriptions;

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_action_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_ip_address;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_account_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_user_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_action;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_account_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_activity_log_user_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_config_versions_type_latest;
DROP INDEX CONCURRENTLY IF EXISTS idx_config_versions_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_config_versions_created_by;
DROP INDEX CONCURRENTLY IF EXISTS idx_config_versions_type_version;
DROP INDEX CONCURRENTLY IF EXISTS idx_config_versions_type;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_subscriptions_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_subscriptions_enabled;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_subscriptions_symbol;
DROP INDEX CONCURRENTLY IF EXISTS idx_lp_subscriptions_lp_id;

-- Drop tables
DROP TABLE IF EXISTS user_activity_log CASCADE;
DROP TABLE IF EXISTS config_versions CASCADE;
DROP TABLE IF EXISTS lp_subscriptions CASCADE;
*/
