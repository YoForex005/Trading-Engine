-- Migration 009: Tick History with TimescaleDB
-- Description: Creates TimescaleDB hypertable for persistent tick storage
-- Author: System Architect
-- Date: 2026-01-20

-- Prerequisites:
-- 1. TimescaleDB extension must be installed:
--    CREATE EXTENSION IF NOT EXISTS timescaledb;
-- 2. PostgreSQL 12+ recommended

-- ============================================================================
-- 1. CREATE TICK HISTORY TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS tick_history (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL,
    broker_id VARCHAR(50) NOT NULL DEFAULT 'default',
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(18,8) NOT NULL,
    ask DECIMAL(18,8) NOT NULL,
    spread DECIMAL(18,8),
    lp VARCHAR(50),
    PRIMARY KEY (id, timestamp)
);

-- Add comments
COMMENT ON TABLE tick_history IS 'Persistent tick storage for all 128 trading symbols';
COMMENT ON COLUMN tick_history.timestamp IS 'Tick timestamp (microsecond precision, UTC)';
COMMENT ON COLUMN tick_history.broker_id IS 'Broker identifier (for multi-tenant setups)';
COMMENT ON COLUMN tick_history.symbol IS 'Trading symbol (EURUSD, GBPUSD, etc.)';
COMMENT ON COLUMN tick_history.bid IS 'Bid price';
COMMENT ON COLUMN tick_history.ask IS 'Ask price';
COMMENT ON COLUMN tick_history.spread IS 'Bid-Ask spread';
COMMENT ON COLUMN tick_history.lp IS 'Liquidity provider source (YOFX, LMAX, etc.)';

-- ============================================================================
-- 2. CONVERT TO TIMESCALEDB HYPERTABLE
-- ============================================================================

-- Create hypertable with 1-day chunks
SELECT create_hypertable(
    'tick_history',
    'timestamp',
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- ============================================================================
-- 3. CREATE INDEXES FOR FAST QUERIES
-- ============================================================================

-- Index for symbol + timestamp queries (most common)
CREATE INDEX IF NOT EXISTS idx_tick_symbol_time
    ON tick_history (symbol, timestamp DESC);

-- Index for broker + symbol + timestamp queries (multi-tenant)
CREATE INDEX IF NOT EXISTS idx_tick_broker_symbol_time
    ON tick_history (broker_id, symbol, timestamp DESC);

-- Index for LP analysis
CREATE INDEX IF NOT EXISTS idx_tick_lp_time
    ON tick_history (lp, timestamp DESC);

-- ============================================================================
-- 4. ENABLE COMPRESSION
-- ============================================================================

-- Configure compression settings
ALTER TABLE tick_history SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol',
    timescaledb.compress_orderby = 'timestamp DESC'
);

-- Add compression policy: compress chunks older than 7 days
SELECT add_compression_policy(
    'tick_history',
    INTERVAL '7 days',
    if_not_exists => TRUE
);

-- ============================================================================
-- 5. SET RETENTION POLICY
-- ============================================================================

-- Automatically drop chunks older than 6 months
SELECT add_retention_policy(
    'tick_history',
    INTERVAL '6 months',
    if_not_exists => TRUE
);

-- ============================================================================
-- 6. CREATE CONTINUOUS AGGREGATES (OPTIONAL - FOR OHLC)
-- ============================================================================

-- 1-minute OHLC continuous aggregate
CREATE MATERIALIZED VIEW IF NOT EXISTS tick_ohlc_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', timestamp) AS bucket,
    symbol,
    FIRST(bid, timestamp) AS open,
    MAX(bid) AS high,
    MIN(bid) AS low,
    LAST(bid, timestamp) AS close,
    COUNT(*) AS volume
FROM tick_history
GROUP BY bucket, symbol
WITH NO DATA;

-- Add refresh policy for 1-minute OHLC (refresh every 1 minute)
SELECT add_continuous_aggregate_policy(
    'tick_ohlc_1m',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => TRUE
);

-- 5-minute OHLC continuous aggregate
CREATE MATERIALIZED VIEW IF NOT EXISTS tick_ohlc_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', timestamp) AS bucket,
    symbol,
    FIRST(bid, timestamp) AS open,
    MAX(bid) AS high,
    MIN(bid) AS low,
    LAST(bid, timestamp) AS close,
    COUNT(*) AS volume
FROM tick_history
GROUP BY bucket, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'tick_ohlc_5m',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes',
    if_not_exists => TRUE
);

-- 1-hour OHLC continuous aggregate
CREATE MATERIALIZED VIEW IF NOT EXISTS tick_ohlc_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', timestamp) AS bucket,
    symbol,
    FIRST(bid, timestamp) AS open,
    MAX(bid) AS high,
    MIN(bid) AS low,
    LAST(bid, timestamp) AS close,
    COUNT(*) AS volume
FROM tick_history
GROUP BY bucket, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'tick_ohlc_1h',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- ============================================================================
-- 7. CREATE HELPER FUNCTIONS
-- ============================================================================

-- Function to get compression stats
CREATE OR REPLACE FUNCTION get_tick_compression_stats()
RETURNS TABLE (
    chunk_name TEXT,
    before_compression_mb NUMERIC,
    after_compression_mb NUMERIC,
    compression_ratio NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.chunk_name::TEXT,
        ROUND(c.before_compression_total_bytes / 1024.0 / 1024.0, 2) AS before_compression_mb,
        ROUND(c.after_compression_total_bytes / 1024.0 / 1024.0, 2) AS after_compression_mb,
        ROUND(c.before_compression_total_bytes::NUMERIC / NULLIF(c.after_compression_total_bytes, 0), 2) AS compression_ratio
    FROM timescaledb_information.compressed_chunk_stats c
    WHERE c.hypertable_name = 'tick_history'
    ORDER BY c.chunk_name DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to get storage stats per symbol
CREATE OR REPLACE FUNCTION get_tick_symbol_stats()
RETURNS TABLE (
    symbol VARCHAR,
    tick_count BIGINT,
    oldest_tick TIMESTAMPTZ,
    newest_tick TIMESTAMPTZ,
    date_range_days NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        th.symbol,
        COUNT(*) AS tick_count,
        MIN(th.timestamp) AS oldest_tick,
        MAX(th.timestamp) AS newest_tick,
        ROUND(EXTRACT(EPOCH FROM (MAX(th.timestamp) - MIN(th.timestamp))) / 86400.0, 2) AS date_range_days
    FROM tick_history th
    GROUP BY th.symbol
    ORDER BY tick_count DESC;
END;
$$ LANGUAGE plpgsql;

-- Function to manually compress a chunk
CREATE OR REPLACE FUNCTION compress_tick_chunk(chunk_name TEXT)
RETURNS VOID AS $$
BEGIN
    EXECUTE format('SELECT compress_chunk(%L)', chunk_name);
    RAISE NOTICE 'Compressed chunk: %', chunk_name;
END;
$$ LANGUAGE plpgsql;

-- Function to manually decompress a chunk
CREATE OR REPLACE FUNCTION decompress_tick_chunk(chunk_name TEXT)
RETURNS VOID AS $$
BEGIN
    EXECUTE format('SELECT decompress_chunk(%L)', chunk_name);
    RAISE NOTICE 'Decompressed chunk: %', chunk_name;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 8. GRANT PERMISSIONS
-- ============================================================================

-- Create application user if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = 'rtx_app') THEN
        CREATE USER rtx_app WITH PASSWORD 'CHANGE_ME_IN_PRODUCTION';
    END IF;
END
$$;

-- Grant permissions
GRANT SELECT, INSERT ON tick_history TO rtx_app;
GRANT USAGE ON SEQUENCE tick_history_id_seq TO rtx_app;
GRANT SELECT ON tick_ohlc_1m, tick_ohlc_5m, tick_ohlc_1h TO rtx_app;
GRANT EXECUTE ON FUNCTION get_tick_compression_stats() TO rtx_app;
GRANT EXECUTE ON FUNCTION get_tick_symbol_stats() TO rtx_app;

-- ============================================================================
-- 9. VERIFY SETUP
-- ============================================================================

-- Check hypertable configuration
SELECT * FROM timescaledb_information.hypertables
WHERE hypertable_name = 'tick_history';

-- Check compression policy
SELECT * FROM timescaledb_information.jobs
WHERE hypertable_name = 'tick_history'
  AND proc_name = 'policy_compression';

-- Check retention policy
SELECT * FROM timescaledb_information.jobs
WHERE hypertable_name = 'tick_history'
  AND proc_name = 'policy_retention';

-- ============================================================================
-- 10. SAMPLE QUERIES
-- ============================================================================

-- Example 1: Get recent ticks for EURUSD
-- SELECT * FROM tick_history
-- WHERE symbol = 'EURUSD'
--   AND timestamp >= NOW() - INTERVAL '1 hour'
-- ORDER BY timestamp DESC
-- LIMIT 1000;

-- Example 2: Get ticks for date range
-- SELECT * FROM tick_history
-- WHERE symbol = 'EURUSD'
--   AND timestamp BETWEEN '2026-01-01' AND '2026-01-20'
-- ORDER BY timestamp;

-- Example 3: Get compression stats
-- SELECT * FROM get_tick_compression_stats();

-- Example 4: Get symbol stats
-- SELECT * FROM get_tick_symbol_stats();

-- Example 5: Get 1-minute OHLC data
-- SELECT * FROM tick_ohlc_1m
-- WHERE symbol = 'EURUSD'
--   AND bucket >= NOW() - INTERVAL '24 hours'
-- ORDER BY bucket DESC;

-- ============================================================================
-- ROLLBACK (IF NEEDED)
-- ============================================================================

-- CAUTION: This will delete ALL tick data!
-- Uncomment only if you need to rollback the migration

/*
DROP MATERIALIZED VIEW IF EXISTS tick_ohlc_1h CASCADE;
DROP MATERIALIZED VIEW IF EXISTS tick_ohlc_5m CASCADE;
DROP MATERIALIZED VIEW IF EXISTS tick_ohlc_1m CASCADE;
DROP FUNCTION IF EXISTS get_tick_compression_stats();
DROP FUNCTION IF EXISTS get_tick_symbol_stats();
DROP FUNCTION IF EXISTS compress_tick_chunk(TEXT);
DROP FUNCTION IF EXISTS decompress_tick_chunk(TEXT);
DROP TABLE IF EXISTS tick_history CASCADE;
*/

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 009 completed: tick_history TimescaleDB hypertable created';
    RAISE NOTICE 'Compression enabled: chunks older than 7 days';
    RAISE NOTICE 'Retention policy: 6 months';
    RAISE NOTICE 'Continuous aggregates: 1m, 5m, 1h OHLC';
END
$$;
