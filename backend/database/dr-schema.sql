-- ============================================
-- RTX Trading Engine - Disaster Recovery Schema Extensions
-- Purpose: Track backup, restore, and failover operations
-- ============================================

-- Backup History Table
CREATE TABLE IF NOT EXISTS backup_history (
    id BIGSERIAL PRIMARY KEY,
    backup_type VARCHAR(20) NOT NULL, -- 'full', 'incremental', 'wal'
    backup_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    size_bytes BIGINT,
    duration_seconds INT,
    s3_path TEXT,
    checksum VARCHAR(64),
    status VARCHAR(20) DEFAULT 'completed', -- 'started', 'completed', 'failed'
    error_message TEXT,
    postgres_version TEXT,
    retention_days INT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_backup_history_date ON backup_history(backup_date DESC);
CREATE INDEX idx_backup_history_type ON backup_history(backup_type, status);

-- Restore History Table
CREATE TABLE IF NOT EXISTS restore_history (
    id BIGSERIAL PRIMARY KEY,
    restore_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    backup_file TEXT NOT NULL,
    restore_duration_seconds INT,
    accounts_count INT,
    positions_count INT,
    trades_count INT,
    status VARCHAR(20) DEFAULT 'completed',
    error_message TEXT,
    restored_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_restore_history_date ON restore_history(restore_date DESC);

-- Failover History Table
CREATE TABLE IF NOT EXISTS failover_history (
    id BIGSERIAL PRIMARY KEY,
    failover_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    from_host VARCHAR(100),
    to_host VARCHAR(100),
    replication_lag_seconds NUMERIC(10, 2),
    services_restarted INT,
    status VARCHAR(20) DEFAULT 'completed',
    error_message TEXT,
    failover_reason TEXT,
    initiated_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_failover_history_date ON failover_history(failover_date DESC);

-- Health Check History Table
CREATE TABLE IF NOT EXISTS health_check_history (
    id BIGSERIAL PRIMARY KEY,
    check_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_name VARCHAR(50) NOT NULL, -- 'database', 'api', 'websocket', etc.
    status VARCHAR(20) NOT NULL, -- 'pass', 'fail', 'warn'
    duration_ms INT,
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partition by day for efficient querying
SELECT create_hypertable('health_check_history', 'check_date', if_not_exists => TRUE);

CREATE INDEX idx_health_check_date ON health_check_history(check_date DESC);
CREATE INDEX idx_health_check_name_status ON health_check_history(check_name, status);

-- DR Metrics Summary View
CREATE OR REPLACE VIEW dr_metrics_summary AS
SELECT
    'backups' AS metric_type,
    COUNT(*) AS total_count,
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS success_count,
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failure_count,
    AVG(duration_seconds) AS avg_duration_seconds,
    MAX(backup_date) AS last_occurrence
FROM backup_history
WHERE backup_date > NOW() - INTERVAL '30 days'
UNION ALL
SELECT
    'restores' AS metric_type,
    COUNT(*),
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END),
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END),
    AVG(restore_duration_seconds),
    MAX(restore_date)
FROM restore_history
WHERE restore_date > NOW() - INTERVAL '30 days'
UNION ALL
SELECT
    'failovers' AS metric_type,
    COUNT(*),
    SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END),
    SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END),
    NULL,
    MAX(failover_date)
FROM failover_history
WHERE failover_date > NOW() - INTERVAL '30 days';

-- Health Check Latest Status View
CREATE OR REPLACE VIEW health_check_latest AS
SELECT DISTINCT ON (check_name)
    check_name,
    status,
    duration_ms,
    details,
    check_date
FROM health_check_history
ORDER BY check_name, check_date DESC;

-- Backup Schedule Compliance Check
CREATE OR REPLACE FUNCTION check_backup_compliance()
RETURNS TABLE (
    backup_type VARCHAR(20),
    expected_interval INTERVAL,
    last_backup TIMESTAMPTZ,
    hours_since_last NUMERIC,
    is_compliant BOOLEAN
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        bt.backup_type,
        bt.expected_interval,
        bh.last_backup,
        EXTRACT(EPOCH FROM (NOW() - bh.last_backup)) / 3600 AS hours_since_last,
        (NOW() - bh.last_backup) < bt.expected_interval AS is_compliant
    FROM (
        VALUES
            ('full'::VARCHAR(20), INTERVAL '1 day'),
            ('incremental', INTERVAL '6 hours'),
            ('wal', INTERVAL '15 minutes')
    ) AS bt(backup_type, expected_interval)
    LEFT JOIN LATERAL (
        SELECT MAX(backup_date) AS last_backup
        FROM backup_history
        WHERE backup_history.backup_type = bt.backup_type
          AND status = 'completed'
    ) bh ON true;
END;
$$ LANGUAGE plpgsql;

-- Alert on backup compliance issues
CREATE OR REPLACE FUNCTION alert_backup_compliance()
RETURNS TEXT AS $$
DECLARE
    non_compliant_count INT;
    alert_message TEXT;
BEGIN
    SELECT COUNT(*)
    INTO non_compliant_count
    FROM check_backup_compliance()
    WHERE NOT is_compliant;

    IF non_compliant_count > 0 THEN
        SELECT string_agg(
            format('%s backup overdue by %.1f hours',
                   backup_type,
                   hours_since_last - EXTRACT(EPOCH FROM expected_interval) / 3600),
            ', '
        )
        INTO alert_message
        FROM check_backup_compliance()
        WHERE NOT is_compliant;

        RETURN format('ALERT: %s', alert_message);
    ELSE
        RETURN 'All backups compliant';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Cleanup old health check records (keep 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_health_checks()
RETURNS INT AS $$
DECLARE
    deleted_count INT;
BEGIN
    DELETE FROM health_check_history
    WHERE check_date < NOW() - INTERVAL '30 days';

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Insert sample data for testing
INSERT INTO backup_history (backup_type, backup_date, size_bytes, duration_seconds, s3_path, checksum, status)
VALUES
    ('full', NOW() - INTERVAL '1 day', 5368709120, 300, 's3://rtx-backups/backups/full/20260118/rtx-full-20260118.dump', 'abc123...', 'completed'),
    ('incremental', NOW() - INTERVAL '6 hours', 1073741824, 60, 's3://rtx-backups/wal-archive/', 'def456...', 'completed')
ON CONFLICT DO NOTHING;

-- Grant permissions
GRANT SELECT ON backup_history TO rtx_readonly;
GRANT SELECT ON restore_history TO rtx_readonly;
GRANT SELECT ON failover_history TO rtx_readonly;
GRANT SELECT ON health_check_history TO rtx_readonly;
GRANT SELECT ON dr_metrics_summary TO rtx_readonly;
GRANT SELECT ON health_check_latest TO rtx_readonly;

-- Comments for documentation
COMMENT ON TABLE backup_history IS 'Track all backup operations for compliance and auditing';
COMMENT ON TABLE restore_history IS 'Record all database restore operations';
COMMENT ON TABLE failover_history IS 'Log all database failover events';
COMMENT ON TABLE health_check_history IS 'Time-series health check results';
COMMENT ON FUNCTION check_backup_compliance() IS 'Verify backups are running on schedule';
COMMENT ON FUNCTION alert_backup_compliance() IS 'Generate alert message for overdue backups';
