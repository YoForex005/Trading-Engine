-- Migration: 008_add_compliance_reporting
-- Description: Enhanced compliance reporting tables for MiFID II and SEC Rule 606
-- Author: Trading Engine Team
-- Date: 2026-01-19

-- ============================================================================
-- UP Migration
-- ============================================================================

-- ============================================================================
-- Best Execution Reports (MiFID II RTS 27/28)
-- ============================================================================

CREATE TABLE IF NOT EXISTS best_execution_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id VARCHAR(100) UNIQUE NOT NULL,
    report_period_start TIMESTAMPTZ NOT NULL,
    report_period_end TIMESTAMPTZ NOT NULL,
    generated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Summary metrics
    total_orders BIGINT DEFAULT 0,
    total_volume DECIMAL(20, 8) DEFAULT 0,
    average_spread DECIMAL(20, 8),
    average_slippage DECIMAL(20, 8),
    fill_rate DECIMAL(5, 2),
    average_latency_ms DECIMAL(10, 2),

    -- Report data (JSON for flexibility)
    venue_breakdown JSONB,
    instrument_stats JSONB,
    metadata JSONB,

    -- Audit trail
    generated_by UUID REFERENCES users(id),
    exported_at TIMESTAMPTZ,
    export_format VARCHAR(10),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_best_execution_reports_period ON best_execution_reports(report_period_start, report_period_end);
CREATE INDEX idx_best_execution_reports_generated_at ON best_execution_reports(generated_at DESC);

-- ============================================================================
-- Venue Execution Metrics (Detailed breakdown)
-- ============================================================================

CREATE TABLE IF NOT EXISTS venue_execution_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id UUID REFERENCES best_execution_reports(id) ON DELETE CASCADE,

    venue_name VARCHAR(100) NOT NULL,
    venue_type VARCHAR(50) CHECK (venue_type IN ('LP', 'Exchange', 'B-Book', 'C-Book')),

    -- Order metrics
    order_count BIGINT DEFAULT 0,
    volume_executed DECIMAL(20, 8) DEFAULT 0,

    -- Quality metrics
    average_spread DECIMAL(20, 8),
    average_slippage DECIMAL(20, 8),
    fill_rate DECIMAL(5, 2),
    reject_rate DECIMAL(5, 2),
    average_latency_ms DECIMAL(10, 2),

    -- Timing
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_venue_metrics_report ON venue_execution_metrics(report_id);
CREATE INDEX idx_venue_metrics_venue ON venue_execution_metrics(venue_name);

-- ============================================================================
-- Order Routing Reports (SEC Rule 606)
-- ============================================================================

CREATE TABLE IF NOT EXISTS order_routing_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id VARCHAR(100) UNIQUE NOT NULL,
    quarter VARCHAR(2) NOT NULL CHECK (quarter IN ('Q1', 'Q2', 'Q3', 'Q4')),
    year INTEGER NOT NULL,
    generated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Summary
    total_orders_routed BIGINT DEFAULT 0,
    total_payment_received DECIMAL(20, 8) DEFAULT 0,
    total_payment_paid DECIMAL(20, 8) DEFAULT 0,
    net_payment DECIMAL(20, 8) DEFAULT 0,

    -- Detailed data
    routing_data JSONB,
    payment_analysis JSONB,
    metadata JSONB,

    -- Audit
    generated_by UUID REFERENCES users(id),
    filed_with_regulator BOOLEAN DEFAULT false,
    filing_date TIMESTAMPTZ,
    filing_reference VARCHAR(100),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_order_routing_reports_quarter ON order_routing_reports(year, quarter);
CREATE INDEX idx_order_routing_reports_generated_at ON order_routing_reports(generated_at DESC);

-- ============================================================================
-- Venue Routing Statistics (Detailed breakdown)
-- ============================================================================

CREATE TABLE IF NOT EXISTS venue_routing_stats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id UUID REFERENCES order_routing_reports(id) ON DELETE CASCADE,

    venue_name VARCHAR(100) NOT NULL,

    -- Order counts
    orders_routed BIGINT DEFAULT 0,
    orders_routed_pct DECIMAL(5, 2),
    non_directed_orders BIGINT DEFAULT 0,
    market_orders BIGINT DEFAULT 0,
    marketable_limit BIGINT DEFAULT 0,
    non_marketable_limit BIGINT DEFAULT 0,
    other_orders BIGINT DEFAULT 0,

    -- Payment for order flow
    average_fee_per_order DECIMAL(20, 8),
    average_rebate_per_order DECIMAL(20, 8),
    net_payment_received DECIMAL(20, 8),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_venue_routing_stats_report ON venue_routing_stats(report_id);
CREATE INDEX idx_venue_routing_stats_venue ON venue_routing_stats(venue_name);

-- ============================================================================
-- Execution Quality Snapshots (Real-time tracking)
-- ============================================================================

CREATE TABLE IF NOT EXISTS execution_quality_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    snapshot_time TIMESTAMPTZ NOT NULL,

    venue_name VARCHAR(100) NOT NULL,
    symbol VARCHAR(20) NOT NULL,

    -- Quality metrics for the snapshot period (e.g., last 1 hour)
    order_count INTEGER DEFAULT 0,
    fill_rate DECIMAL(5, 2),
    average_spread DECIMAL(20, 8),
    average_slippage DECIMAL(20, 8),
    average_latency_ms DECIMAL(10, 2),
    reject_rate DECIMAL(5, 2),

    -- Volume metrics
    volume_executed DECIMAL(20, 8),

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_execution_quality_venue_symbol ON execution_quality_snapshots(venue_name, symbol);
CREATE INDEX idx_execution_quality_snapshot_time ON execution_quality_snapshots(snapshot_time DESC);

-- ============================================================================
-- Audit Trail Exports (Track all exports for compliance)
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_trail_exports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    export_id VARCHAR(100) UNIQUE NOT NULL,

    -- Export criteria
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    entity_type_filter VARCHAR(50),

    -- Export details
    total_records BIGINT DEFAULT 0,
    export_format VARCHAR(10) CHECK (export_format IN ('json', 'csv', 'pdf')),
    file_path TEXT,
    file_size_bytes BIGINT,

    -- Audit
    requested_by UUID REFERENCES users(id),
    purpose TEXT,
    requested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMPTZ,

    -- Integrity
    checksum VARCHAR(128),
    encryption_enabled BOOLEAN DEFAULT false,

    metadata JSONB
);

CREATE INDEX idx_audit_exports_requested_by ON audit_trail_exports(requested_by);
CREATE INDEX idx_audit_exports_requested_at ON audit_trail_exports(requested_at DESC);
CREATE INDEX idx_audit_exports_period ON audit_trail_exports(period_start, period_end);

-- ============================================================================
-- Enhanced Audit Log with Tamper-Proof Features
-- ============================================================================

-- Add hash column to existing audit_log for tamper detection
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS content_hash VARCHAR(128);
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS previous_hash VARCHAR(128);
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS chain_verified BOOLEAN DEFAULT true;

-- Create function to generate audit hash
CREATE OR REPLACE FUNCTION generate_audit_hash(
    p_timestamp TIMESTAMPTZ,
    p_user_id UUID,
    p_table_name VARCHAR,
    p_record_id UUID,
    p_action VARCHAR,
    p_previous_hash VARCHAR
)
RETURNS VARCHAR AS $$
DECLARE
    v_hash_input TEXT;
    v_hash VARCHAR(128);
BEGIN
    -- Concatenate all fields for hashing
    v_hash_input := CONCAT(
        EXTRACT(EPOCH FROM p_timestamp)::TEXT,
        '|', COALESCE(p_user_id::TEXT, ''),
        '|', p_table_name,
        '|', p_record_id::TEXT,
        '|', p_action,
        '|', COALESCE(p_previous_hash, '')
    );

    -- Generate SHA-256 hash (simplified - use pgcrypto extension in production)
    v_hash := encode(digest(v_hash_input, 'sha256'), 'hex');

    RETURN v_hash;
END;
$$ LANGUAGE plpgsql;

-- Update audit trigger to include hash generation
CREATE OR REPLACE FUNCTION audit_trigger_function_with_hash()
RETURNS TRIGGER AS $$
DECLARE
    v_previous_hash VARCHAR(128);
    v_new_hash VARCHAR(128);
BEGIN
    -- Get the hash of the most recent audit entry for this table
    SELECT content_hash INTO v_previous_hash
    FROM audit_log
    WHERE table_name = TG_TABLE_NAME
    ORDER BY created_at DESC
    LIMIT 1;

    IF TG_OP = 'DELETE' THEN
        v_new_hash := generate_audit_hash(
            CURRENT_TIMESTAMP,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            TG_TABLE_NAME,
            OLD.id,
            TG_OP,
            v_previous_hash
        );

        INSERT INTO audit_log (table_name, record_id, action, old_data, content_hash, previous_hash)
        VALUES (TG_TABLE_NAME, OLD.id, TG_OP, row_to_json(OLD), v_new_hash, v_previous_hash);
        RETURN OLD;

    ELSIF TG_OP = 'UPDATE' THEN
        v_new_hash := generate_audit_hash(
            CURRENT_TIMESTAMP,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            TG_TABLE_NAME,
            NEW.id,
            TG_OP,
            v_previous_hash
        );

        INSERT INTO audit_log (table_name, record_id, action, old_data, new_data, content_hash, previous_hash)
        VALUES (TG_TABLE_NAME, NEW.id, TG_OP, row_to_json(OLD), row_to_json(NEW), v_new_hash, v_previous_hash);
        RETURN NEW;

    ELSIF TG_OP = 'INSERT' THEN
        v_new_hash := generate_audit_hash(
            CURRENT_TIMESTAMP,
            NULLIF(current_setting('app.current_user_id', true), '')::UUID,
            TG_TABLE_NAME,
            NEW.id,
            TG_OP,
            v_previous_hash
        );

        INSERT INTO audit_log (table_name, record_id, action, new_data, content_hash, previous_hash)
        VALUES (TG_TABLE_NAME, NEW.id, TG_OP, row_to_json(NEW), v_new_hash, v_previous_hash);
        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- Data Retention Policies
-- ============================================================================

-- Retention metadata table
CREATE TABLE IF NOT EXISTS data_retention_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(100) NOT NULL,
    retention_years INTEGER NOT NULL,
    archive_enabled BOOLEAN DEFAULT true,
    archive_path TEXT,
    last_cleanup_at TIMESTAMPTZ,
    next_cleanup_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Insert default policies
INSERT INTO data_retention_policies (table_name, retention_years, archive_enabled) VALUES
    ('audit_log', 7, true),
    ('order_audit', 7, true),
    ('position_audit', 7, true),
    ('account_audit', 7, true),
    ('compliance_events', 7, true),
    ('api_access_log', 2, true),
    ('user_activity_log', 3, true),
    ('risk_events', 5, true),
    ('system_events', 1, false)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Compliance Alert Rules
-- ============================================================================

CREATE TABLE IF NOT EXISTS compliance_alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_name VARCHAR(200) NOT NULL,
    rule_type VARCHAR(50) NOT NULL CHECK (rule_type IN (
        'execution_quality', 'routing_concentration', 'payment_threshold',
        'audit_anomaly', 'regulatory_threshold'
    )),
    enabled BOOLEAN DEFAULT true,

    -- Thresholds
    threshold_config JSONB NOT NULL,

    -- Actions
    alert_severity VARCHAR(20) CHECK (alert_severity IN ('info', 'warning', 'critical')),
    notification_emails TEXT[],
    auto_report BOOLEAN DEFAULT false,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Insert default alert rules
INSERT INTO compliance_alert_rules (rule_name, rule_type, alert_severity, threshold_config) VALUES
    ('Poor Execution Quality', 'execution_quality', 'warning',
     '{"fill_rate_threshold": 95.0, "reject_rate_threshold": 5.0}'),
    ('Routing Concentration Risk', 'routing_concentration', 'warning',
     '{"max_venue_percentage": 80.0}'),
    ('High Payment for Order Flow', 'payment_threshold', 'info',
     '{"max_payment_percentage": 10.0}'),
    ('Audit Chain Break', 'audit_anomaly', 'critical',
     '{"verify_chain": true}')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Compliance Alerts (Triggered alerts)
-- ============================================================================

CREATE TABLE IF NOT EXISTS compliance_alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_id UUID REFERENCES compliance_alert_rules(id),

    severity VARCHAR(20) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,

    -- Context
    triggered_by_entity VARCHAR(50),
    triggered_by_id UUID,
    threshold_value DECIMAL(20, 8),
    actual_value DECIMAL(20, 8),

    -- Status
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'acknowledged', 'resolved', 'false_positive')),
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    resolution_notes TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE INDEX idx_compliance_alerts_status ON compliance_alerts(status);
CREATE INDEX idx_compliance_alerts_severity ON compliance_alerts(severity);
CREATE INDEX idx_compliance_alerts_created_at ON compliance_alerts(created_at DESC);

-- ============================================================================
-- Regulatory Filing Tracker
-- ============================================================================

CREATE TABLE IF NOT EXISTS regulatory_filings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filing_type VARCHAR(100) NOT NULL,
    filing_period VARCHAR(20) NOT NULL,
    regulator VARCHAR(100) NOT NULL CHECK (regulator IN ('SEC', 'ESMA', 'FCA', 'FINRA', 'CFTC', 'Other')),

    -- Filing details
    report_id UUID,
    filing_reference VARCHAR(100),
    filing_deadline TIMESTAMPTZ NOT NULL,
    filing_status VARCHAR(20) DEFAULT 'pending' CHECK (filing_status IN ('pending', 'submitted', 'accepted', 'rejected', 'amended')),

    -- Submission
    submitted_at TIMESTAMPTZ,
    submitted_by UUID REFERENCES users(id),
    confirmation_number VARCHAR(100),

    -- Acceptance/Rejection
    regulator_response TEXT,
    response_received_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

CREATE INDEX idx_regulatory_filings_status ON regulatory_filings(filing_status);
CREATE INDEX idx_regulatory_filings_deadline ON regulatory_filings(filing_deadline);
CREATE INDEX idx_regulatory_filings_type ON regulatory_filings(filing_type, regulator);

-- ============================================================================
-- Views for Compliance Dashboard
-- ============================================================================

-- Recent Best Execution Metrics
CREATE OR REPLACE VIEW v_recent_execution_quality AS
SELECT
    venue_name,
    symbol,
    AVG(fill_rate) as avg_fill_rate,
    AVG(average_spread) as avg_spread,
    AVG(average_slippage) as avg_slippage,
    AVG(average_latency_ms) as avg_latency_ms,
    SUM(order_count) as total_orders
FROM execution_quality_snapshots
WHERE snapshot_time > NOW() - INTERVAL '24 hours'
GROUP BY venue_name, symbol;

-- Open Compliance Alerts Summary
CREATE OR REPLACE VIEW v_open_compliance_alerts AS
SELECT
    severity,
    alert_type,
    COUNT(*) as alert_count,
    MIN(created_at) as oldest_alert,
    MAX(created_at) as newest_alert
FROM compliance_alerts
WHERE status = 'open'
GROUP BY severity, alert_type
ORDER BY
    CASE severity
        WHEN 'critical' THEN 1
        WHEN 'warning' THEN 2
        ELSE 3
    END;

-- ============================================================================
-- Functions for Compliance Checks
-- ============================================================================

-- Function to verify audit chain integrity
CREATE OR REPLACE FUNCTION verify_audit_chain(p_table_name VARCHAR, p_limit INTEGER DEFAULT 1000)
RETURNS TABLE (
    audit_id UUID,
    is_valid BOOLEAN,
    error_message TEXT
) AS $$
DECLARE
    v_record RECORD;
    v_computed_hash VARCHAR(128);
BEGIN
    FOR v_record IN
        SELECT * FROM audit_log
        WHERE table_name = p_table_name
        ORDER BY created_at DESC
        LIMIT p_limit
    LOOP
        -- Verify hash
        v_computed_hash := generate_audit_hash(
            v_record.created_at,
            v_record.user_id,
            v_record.table_name,
            v_record.record_id,
            v_record.action,
            v_record.previous_hash
        );

        IF v_computed_hash = v_record.content_hash THEN
            audit_id := v_record.id;
            is_valid := true;
            error_message := NULL;
            RETURN NEXT;
        ELSE
            audit_id := v_record.id;
            is_valid := false;
            error_message := 'Hash mismatch - possible tampering';
            RETURN NEXT;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- DOWN Migration
-- ============================================================================

/*
-- To rollback:

DROP VIEW IF EXISTS v_open_compliance_alerts;
DROP VIEW IF EXISTS v_recent_execution_quality;

DROP TABLE IF EXISTS regulatory_filings CASCADE;
DROP TABLE IF EXISTS compliance_alerts CASCADE;
DROP TABLE IF EXISTS compliance_alert_rules CASCADE;
DROP TABLE IF EXISTS data_retention_policies CASCADE;
DROP TABLE IF EXISTS execution_quality_snapshots CASCADE;
DROP TABLE IF EXISTS venue_routing_stats CASCADE;
DROP TABLE IF EXISTS order_routing_reports CASCADE;
DROP TABLE IF EXISTS venue_execution_metrics CASCADE;
DROP TABLE IF EXISTS best_execution_reports CASCADE;
DROP TABLE IF EXISTS audit_trail_exports CASCADE;

DROP FUNCTION IF EXISTS verify_audit_chain(VARCHAR, INTEGER);
DROP FUNCTION IF EXISTS audit_trigger_function_with_hash();
DROP FUNCTION IF EXISTS generate_audit_hash(TIMESTAMPTZ, UUID, VARCHAR, UUID, VARCHAR, VARCHAR);

ALTER TABLE audit_log DROP COLUMN IF EXISTS content_hash;
ALTER TABLE audit_log DROP COLUMN IF EXISTS previous_hash;
ALTER TABLE audit_log DROP COLUMN IF EXISTS chain_verified;
*/
