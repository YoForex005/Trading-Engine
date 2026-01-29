package migrations

import (
	"database/sql"
)

func init() {
	RegisterMigration(&Migration{
		Version: 7,
		Name:    "analytics_dashboard_schema",
		Up:      analyticsDashboardSchemaUp,
		Down:    analyticsDashboardSchemaDown,
	})
}

func analyticsDashboardSchemaUp(tx *sql.Tx) error {
	schema := `
	-- ========================================
	-- ANALYTICS DASHBOARD SCHEMA
	-- Migration 007: Real-time analytics tables
	-- ========================================
	-- Based on: REALTIME_ANALYTICS_ARCHITECTURE.md
	--           ANALYTICS_DASHBOARD_MASTER_PLAN.md
	-- Purpose: Track routing decisions, LP performance,
	--          exposure management, and rule effectiveness
	-- ========================================

	-- 1. ROUTING METRICS TABLE
	-- Tracks real-time routing decisions (A-Book/B-Book/C-Book)
	-- Partition strategy: Daily time-based partitioning for performance
	CREATE TABLE IF NOT EXISTS routing_metrics (
		id BIGSERIAL,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		symbol VARCHAR(50) NOT NULL,
		account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

		-- Routing decision details
		routing_decision VARCHAR(20) NOT NULL CHECK (routing_decision IN ('ABOOK', 'BBOOK', 'CBOOK')),
		volume DECIMAL(20, 5) NOT NULL,

		-- Exposure tracking
		exposure_before DECIMAL(20, 5) NOT NULL DEFAULT 0,
		exposure_after DECIMAL(20, 5) NOT NULL DEFAULT 0,

		-- Decision metadata
		reason TEXT,
		confidence_score DECIMAL(5, 4) CHECK (confidence_score >= 0 AND confidence_score <= 1),

		-- LP selection (for A-Book/C-Book)
		selected_lp VARCHAR(100),

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (id, timestamp)
	);

	-- Indexes for fast querying
	CREATE INDEX idx_routing_metrics_timestamp ON routing_metrics(timestamp DESC);
	CREATE INDEX idx_routing_metrics_symbol_timestamp ON routing_metrics(symbol, timestamp DESC);
	CREATE INDEX idx_routing_metrics_account_timestamp ON routing_metrics(account_id, timestamp DESC);
	CREATE INDEX idx_routing_metrics_routing_decision ON routing_metrics(routing_decision, timestamp DESC);
	CREATE INDEX idx_routing_metrics_selected_lp ON routing_metrics(selected_lp, timestamp DESC) WHERE selected_lp IS NOT NULL;

	-- Comment for documentation
	COMMENT ON TABLE routing_metrics IS 'Real-time routing decision tracking for A-Book/B-Book/C-Book analytics';
	COMMENT ON COLUMN routing_metrics.routing_decision IS 'ABOOK=LP routing, BBOOK=internalization, CBOOK=hybrid';
	COMMENT ON COLUMN routing_metrics.confidence_score IS 'ML confidence score for routing decision (0-1)';

	-- 2. LP PERFORMANCE TABLE
	-- Tracks liquidity provider metrics and performance
	CREATE TABLE IF NOT EXISTS lp_performance (
		id BIGSERIAL,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		lp_name VARCHAR(100) NOT NULL,
		symbol VARCHAR(50) NOT NULL,

		-- Performance metrics
		avg_latency_ms DECIMAL(10, 2) NOT NULL CHECK (avg_latency_ms >= 0),
		fill_rate_pct DECIMAL(5, 2) NOT NULL CHECK (fill_rate_pct >= 0 AND fill_rate_pct <= 100),
		slippage_bps DECIMAL(10, 2) NOT NULL DEFAULT 0,

		-- Volume metrics
		volume_24h DECIMAL(20, 5) NOT NULL DEFAULT 0,
		trade_count_24h BIGINT NOT NULL DEFAULT 0,

		-- Availability metrics
		uptime_pct DECIMAL(5, 2) NOT NULL CHECK (uptime_pct >= 0 AND uptime_pct <= 100),
		reject_rate_pct DECIMAL(5, 2) NOT NULL DEFAULT 0 CHECK (reject_rate_pct >= 0 AND reject_rate_pct <= 100),

		-- Quality metrics
		best_execution_pct DECIMAL(5, 2) DEFAULT 0 CHECK (best_execution_pct >= 0 AND best_execution_pct <= 100),
		avg_spread_bps DECIMAL(10, 2) DEFAULT 0,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (id, timestamp)
	);

	-- Indexes for LP performance queries
	CREATE INDEX idx_lp_performance_timestamp ON lp_performance(timestamp DESC);
	CREATE INDEX idx_lp_performance_lp_timestamp ON lp_performance(lp_name, timestamp DESC);
	CREATE INDEX idx_lp_performance_symbol_timestamp ON lp_performance(symbol, timestamp DESC);
	CREATE INDEX idx_lp_performance_lp_symbol ON lp_performance(lp_name, symbol, timestamp DESC);

	-- Comments
	COMMENT ON TABLE lp_performance IS 'Liquidity provider performance metrics for dashboard analytics';
	COMMENT ON COLUMN lp_performance.avg_latency_ms IS 'Average execution latency in milliseconds';
	COMMENT ON COLUMN lp_performance.fill_rate_pct IS 'Percentage of orders successfully filled';
	COMMENT ON COLUMN lp_performance.slippage_bps IS 'Average slippage in basis points';
	COMMENT ON COLUMN lp_performance.uptime_pct IS 'LP availability percentage over measurement period';

	-- 3. EXPOSURE SNAPSHOTS TABLE
	-- Tracks symbol exposure over time for heatmap visualization
	CREATE TABLE IF NOT EXISTS exposure_snapshots (
		id BIGSERIAL,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		symbol VARCHAR(50) NOT NULL,

		-- Exposure details
		net_exposure DECIMAL(20, 5) NOT NULL DEFAULT 0,
		long_exposure DECIMAL(20, 5) NOT NULL DEFAULT 0,
		short_exposure DECIMAL(20, 5) NOT NULL DEFAULT 0,

		-- Risk metrics
		utilization_pct DECIMAL(5, 2) NOT NULL DEFAULT 0 CHECK (utilization_pct >= 0 AND utilization_pct <= 100),
		limit_value DECIMAL(20, 5) NOT NULL,

		-- Position count
		open_positions_count INT NOT NULL DEFAULT 0,

		-- Margin impact
		margin_used DECIMAL(20, 5) NOT NULL DEFAULT 0,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY (id, timestamp)
	);

	-- Indexes for exposure queries
	CREATE INDEX idx_exposure_snapshots_timestamp ON exposure_snapshots(timestamp DESC);
	CREATE INDEX idx_exposure_snapshots_symbol_timestamp ON exposure_snapshots(symbol, timestamp DESC);
	CREATE INDEX idx_exposure_snapshots_utilization ON exposure_snapshots(utilization_pct DESC, timestamp DESC);

	-- Comments
	COMMENT ON TABLE exposure_snapshots IS 'Time-series snapshots of symbol exposure for risk management and heatmap visualization';
	COMMENT ON COLUMN exposure_snapshots.net_exposure IS 'Net exposure (long - short) for the symbol';
	COMMENT ON COLUMN exposure_snapshots.utilization_pct IS 'Percentage of exposure limit utilized';

	-- 4. RULE EFFECTIVENESS TABLE
	-- Tracks routing rule performance and scoring
	CREATE TABLE IF NOT EXISTS rule_effectiveness (
		id BIGSERIAL PRIMARY KEY,
		rule_id VARCHAR(100) NOT NULL,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Performance metrics
		sharpe_ratio DECIMAL(10, 4) DEFAULT 0,
		profit_factor DECIMAL(10, 4) DEFAULT 0,
		max_drawdown DECIMAL(10, 4) DEFAULT 0,

		-- P&L metrics
		total_pnl DECIMAL(20, 5) NOT NULL DEFAULT 0,
		gross_profit DECIMAL(20, 5) NOT NULL DEFAULT 0,
		gross_loss DECIMAL(20, 5) NOT NULL DEFAULT 0,

		-- Trade statistics
		trade_count BIGINT NOT NULL DEFAULT 0,
		win_count BIGINT NOT NULL DEFAULT 0,
		loss_count BIGINT NOT NULL DEFAULT 0,
		win_rate DECIMAL(5, 2) DEFAULT 0 CHECK (win_rate >= 0 AND win_rate <= 100),

		-- Execution quality
		avg_trade_duration INT DEFAULT 0,
		avg_slippage_bps DECIMAL(10, 2) DEFAULT 0,

		-- Rule metadata
		rule_name VARCHAR(255),
		rule_description TEXT,
		is_active BOOLEAN DEFAULT TRUE,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes for rule effectiveness queries
	CREATE INDEX idx_rule_effectiveness_rule_id ON rule_effectiveness(rule_id, timestamp DESC);
	CREATE INDEX idx_rule_effectiveness_timestamp ON rule_effectiveness(timestamp DESC);
	CREATE INDEX idx_rule_effectiveness_sharpe_ratio ON rule_effectiveness(sharpe_ratio DESC, timestamp DESC);
	CREATE INDEX idx_rule_effectiveness_profit_factor ON rule_effectiveness(profit_factor DESC, timestamp DESC);
	CREATE INDEX idx_rule_effectiveness_active ON rule_effectiveness(is_active, timestamp DESC);

	-- Comments
	COMMENT ON TABLE rule_effectiveness IS 'Routing rule performance scoring and effectiveness tracking';
	COMMENT ON COLUMN rule_effectiveness.sharpe_ratio IS 'Risk-adjusted return metric (return - risk-free rate) / standard deviation';
	COMMENT ON COLUMN rule_effectiveness.profit_factor IS 'Ratio of gross profit to gross loss (>1.0 = profitable)';
	COMMENT ON COLUMN rule_effectiveness.max_drawdown IS 'Maximum peak-to-trough decline percentage';

	-- 5. ALERTS TABLE
	-- Comprehensive alerting system for analytics dashboard
	CREATE TABLE IF NOT EXISTS alerts (
		id VARCHAR(255) PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Alert classification
		severity VARCHAR(20) NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
		type VARCHAR(100) NOT NULL,

		-- Context
		symbol VARCHAR(50),
		account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
		lp_name VARCHAR(100),

		-- Alert details
		message TEXT NOT NULL,
		details JSONB,

		-- Threshold context
		threshold_value DECIMAL(20, 5),
		current_value DECIMAL(20, 5),

		-- Acknowledgment tracking
		acknowledged BOOLEAN DEFAULT FALSE,
		acknowledged_by VARCHAR(255),
		acknowledged_at TIMESTAMPTZ,

		-- Resolution tracking
		resolved BOOLEAN DEFAULT FALSE,
		resolved_by VARCHAR(255),
		resolved_at TIMESTAMPTZ,
		resolution_notes TEXT,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes for alert queries
	CREATE INDEX idx_alerts_timestamp ON alerts(timestamp DESC);
	CREATE INDEX idx_alerts_severity_timestamp ON alerts(severity, timestamp DESC);
	CREATE INDEX idx_alerts_type_timestamp ON alerts(type, timestamp DESC);
	CREATE INDEX idx_alerts_acknowledged ON alerts(acknowledged, timestamp DESC);
	CREATE INDEX idx_alerts_resolved ON alerts(resolved, timestamp DESC);
	CREATE INDEX idx_alerts_account_id ON alerts(account_id, timestamp DESC) WHERE account_id IS NOT NULL;
	CREATE INDEX idx_alerts_symbol ON alerts(symbol, timestamp DESC) WHERE symbol IS NOT NULL;

	-- Comments
	COMMENT ON TABLE alerts IS 'Analytics dashboard alerting system for exposure, performance, and risk alerts';
	COMMENT ON COLUMN alerts.severity IS 'Alert severity level: LOW, MEDIUM, HIGH, or CRITICAL';
	COMMENT ON COLUMN alerts.details IS 'Additional context and metadata in JSON format';

	-- 6. ANALYTICS EXPORT AUDIT TABLE
	-- Tracks all data exports for compliance (GDPR, MiFID II, SEC Rule 606)
	CREATE TABLE IF NOT EXISTS analytics_export_audit (
		id BIGSERIAL PRIMARY KEY,
		export_id VARCHAR(255) UNIQUE NOT NULL,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- User context
		user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
		username VARCHAR(255) NOT NULL,
		ip_address VARCHAR(50),

		-- Export details
		export_type VARCHAR(50) NOT NULL CHECK (export_type IN ('CSV', 'EXCEL', 'PDF', 'JSON', 'PARQUET')),
		data_scope VARCHAR(100) NOT NULL,

		-- Date range
		date_from TIMESTAMPTZ,
		date_to TIMESTAMPTZ,

		-- Filters applied
		filters JSONB,

		-- File details
		file_size_bytes BIGINT,
		row_count BIGINT,
		file_hash VARCHAR(64),

		-- Storage location
		storage_path TEXT,
		s3_bucket VARCHAR(255),
		s3_key TEXT,

		-- Delivery
		email_sent BOOLEAN DEFAULT FALSE,
		email_recipients TEXT[],

		-- Status
		status VARCHAR(50) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')),
		error_message TEXT,

		-- Retention policy
		expires_at TIMESTAMPTZ,
		deleted_at TIMESTAMPTZ,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMPTZ
	);

	-- Indexes for export audit queries
	CREATE INDEX idx_export_audit_timestamp ON analytics_export_audit(timestamp DESC);
	CREATE INDEX idx_export_audit_user_id ON analytics_export_audit(user_id, timestamp DESC);
	CREATE INDEX idx_export_audit_export_type ON analytics_export_audit(export_type, timestamp DESC);
	CREATE INDEX idx_export_audit_status ON analytics_export_audit(status, timestamp DESC);
	CREATE INDEX idx_export_audit_expires_at ON analytics_export_audit(expires_at) WHERE deleted_at IS NULL;

	-- Comments
	COMMENT ON TABLE analytics_export_audit IS 'Audit log for all analytics data exports (GDPR, MiFID II, SEC Rule 606 compliance)';
	COMMENT ON COLUMN analytics_export_audit.file_hash IS 'SHA-256 hash for tamper detection and data integrity';
	COMMENT ON COLUMN analytics_export_audit.expires_at IS 'Auto-deletion timestamp per retention policy';

	-- 7. USER PREFERENCES TABLE
	-- Dashboard customization and user settings
	CREATE TABLE IF NOT EXISTS analytics_user_preferences (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

		-- Dashboard layout
		default_view VARCHAR(50) DEFAULT 'overview',
		dashboard_layout JSONB,

		-- Chart preferences
		chart_theme VARCHAR(50) DEFAULT 'light',
		chart_interval VARCHAR(20) DEFAULT '1m',

		-- Filter defaults
		default_symbols TEXT[],
		default_timerange VARCHAR(50) DEFAULT '24h',

		-- Alert preferences
		alert_channels TEXT[] DEFAULT ARRAY['in-dashboard', 'email'],
		email_alerts BOOLEAN DEFAULT TRUE,
		push_notifications BOOLEAN DEFAULT FALSE,

		-- Display preferences
		timezone VARCHAR(100) DEFAULT 'UTC',
		currency VARCHAR(10) DEFAULT 'USD',
		number_format VARCHAR(50) DEFAULT 'en-US',

		-- Privacy settings
		share_analytics BOOLEAN DEFAULT FALSE,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	-- Index
	CREATE INDEX idx_analytics_user_preferences_user_id ON analytics_user_preferences(user_id);

	-- Comments
	COMMENT ON TABLE analytics_user_preferences IS 'User-specific dashboard customization and preferences';

	-- ========================================
	-- MATERIALIZED VIEWS FOR PERFORMANCE
	-- Pre-computed aggregations for fast queries
	-- ========================================

	-- Materialized view: Hourly LP performance summary
	CREATE MATERIALIZED VIEW IF NOT EXISTS lp_performance_hourly AS
	SELECT
		date_trunc('hour', timestamp) AS hour,
		lp_name,
		symbol,
		AVG(avg_latency_ms) AS avg_latency_ms,
		AVG(fill_rate_pct) AS avg_fill_rate_pct,
		AVG(slippage_bps) AS avg_slippage_bps,
		SUM(volume_24h) AS total_volume,
		AVG(uptime_pct) AS avg_uptime_pct,
		AVG(reject_rate_pct) AS avg_reject_rate_pct,
		COUNT(*) AS sample_count
	FROM lp_performance
	GROUP BY hour, lp_name, symbol;

	-- Index on materialized view
	CREATE UNIQUE INDEX idx_lp_performance_hourly_unique ON lp_performance_hourly(hour, lp_name, symbol);
	CREATE INDEX idx_lp_performance_hourly_hour ON lp_performance_hourly(hour DESC);

	-- Materialized view: Daily routing decision summary
	CREATE MATERIALIZED VIEW IF NOT EXISTS routing_metrics_daily AS
	SELECT
		date_trunc('day', timestamp) AS day,
		routing_decision,
		symbol,
		COUNT(*) AS decision_count,
		SUM(volume) AS total_volume,
		AVG(confidence_score) AS avg_confidence_score
	FROM routing_metrics
	GROUP BY day, routing_decision, symbol;

	-- Index on materialized view
	CREATE UNIQUE INDEX idx_routing_metrics_daily_unique ON routing_metrics_daily(day, routing_decision, symbol);
	CREATE INDEX idx_routing_metrics_daily_day ON routing_metrics_daily(day DESC);

	-- ========================================
	-- TRIGGER FUNCTIONS FOR AUTO-UPDATES
	-- ========================================

	-- Function: Update timestamp on row modification
	CREATE OR REPLACE FUNCTION update_analytics_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Triggers for auto-updating timestamps
	CREATE TRIGGER update_rule_effectiveness_timestamp
		BEFORE UPDATE ON rule_effectiveness
		FOR EACH ROW
		EXECUTE FUNCTION update_analytics_timestamp();

	CREATE TRIGGER update_analytics_user_preferences_timestamp
		BEFORE UPDATE ON analytics_user_preferences
		FOR EACH ROW
		EXECUTE FUNCTION update_analytics_timestamp();

	-- ========================================
	-- DATA RETENTION POLICIES
	-- ========================================

	-- Note: Retention policies implemented via cron jobs or pg_cron
	-- Example:
	-- DELETE FROM routing_metrics WHERE timestamp < NOW() - INTERVAL '30 days';
	-- DELETE FROM lp_performance WHERE timestamp < NOW() - INTERVAL '90 days';
	-- DELETE FROM exposure_snapshots WHERE timestamp < NOW() - INTERVAL '90 days';
	-- DELETE FROM alerts WHERE timestamp < NOW() - INTERVAL '1 year' AND resolved = TRUE;
	-- DELETE FROM analytics_export_audit WHERE expires_at < NOW() AND deleted_at IS NULL;

	`

	_, err := tx.Exec(schema)
	return err
}

func analyticsDashboardSchemaDown(tx *sql.Tx) error {
	dropSchema := `
	-- Drop triggers
	DROP TRIGGER IF EXISTS update_rule_effectiveness_timestamp ON rule_effectiveness;
	DROP TRIGGER IF EXISTS update_analytics_user_preferences_timestamp ON analytics_user_preferences;

	-- Drop trigger function
	DROP FUNCTION IF EXISTS update_analytics_timestamp();

	-- Drop materialized views
	DROP MATERIALIZED VIEW IF EXISTS routing_metrics_daily;
	DROP MATERIALIZED VIEW IF EXISTS lp_performance_hourly;

	-- Drop tables (reverse order of creation to handle dependencies)
	DROP TABLE IF EXISTS analytics_user_preferences;
	DROP TABLE IF EXISTS analytics_export_audit;
	DROP TABLE IF EXISTS alerts;
	DROP TABLE IF EXISTS rule_effectiveness;
	DROP TABLE IF EXISTS exposure_snapshots;
	DROP TABLE IF EXISTS lp_performance;
	DROP TABLE IF EXISTS routing_metrics;
	`

	_, err := tx.Exec(dropSchema)
	return err
}
