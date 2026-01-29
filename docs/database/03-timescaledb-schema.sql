-- ============================================================================
-- RTX Trading Engine - TimescaleDB Schema for Market Data
-- Version: 1.0
-- Description: High-performance time-series storage for ticks and OHLC data
-- ============================================================================

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ============================================================================
-- TICK DATA
-- Purpose: Store raw tick data with microsecond precision
-- ============================================================================

-- Tick data hypertable
CREATE TABLE ticks (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(18, 8) NOT NULL,
    ask DECIMAL(18, 8) NOT NULL,
    bid_volume DECIMAL(18, 4),
    ask_volume DECIMAL(18, 4),
    liquidity_provider VARCHAR(50),
    CONSTRAINT ticks_pkey PRIMARY KEY (time, symbol)
);

-- Convert to hypertable with 1-day chunks
SELECT create_hypertable('ticks', 'time', chunk_time_interval => INTERVAL '1 day');

-- Create indexes for efficient queries
CREATE INDEX idx_ticks_symbol_time ON ticks (symbol, time DESC);
CREATE INDEX idx_ticks_lp ON ticks (liquidity_provider, time DESC) WHERE liquidity_provider IS NOT NULL;

-- Add compression policy (compress data older than 7 days)
ALTER TABLE ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('ticks', INTERVAL '7 days');

-- Add retention policy (keep tick data for 30 days)
SELECT add_retention_policy('ticks', INTERVAL '30 days');

COMMENT ON TABLE ticks IS 'Raw tick data from liquidity providers';

-- ============================================================================
-- OHLC DATA (Candlesticks)
-- Purpose: Aggregated OHLC data for various timeframes
-- ============================================================================

-- OHLC 1-minute candles (base timeframe)
CREATE TABLE ohlc_1m (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    open DECIMAL(18, 8) NOT NULL,
    high DECIMAL(18, 8) NOT NULL,
    low DECIMAL(18, 8) NOT NULL,
    close DECIMAL(18, 8) NOT NULL,
    volume DECIMAL(18, 4) NOT NULL DEFAULT 0,
    tick_count INT NOT NULL DEFAULT 0,
    CONSTRAINT ohlc_1m_pkey PRIMARY KEY (time, symbol)
);

-- Convert to hypertable
SELECT create_hypertable('ohlc_1m', 'time', chunk_time_interval => INTERVAL '7 days');

-- Create continuous aggregate from ticks to 1-minute OHLC
CREATE MATERIALIZED VIEW ohlc_1m_continuous
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS time,
    symbol,
    first(bid, time) AS open,
    max(bid) AS high,
    min(bid) AS low,
    last(bid, time) AS close,
    count(*) AS tick_count
FROM ticks
GROUP BY time_bucket('1 minute', time), symbol
WITH NO DATA;

-- Add refresh policy (refresh every 1 minute)
SELECT add_continuous_aggregate_policy('ohlc_1m_continuous',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- OHLC 5-minute candles
CREATE MATERIALIZED VIEW ohlc_5m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('5 minutes', time) AS time,
    symbol,
    first(open, time) AS open,
    max(high) AS high,
    min(low) AS low,
    last(close, time) AS close,
    sum(volume) AS volume,
    sum(tick_count) AS tick_count
FROM ohlc_1m
GROUP BY time_bucket('5 minutes', time), symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('ohlc_5m',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes');

-- OHLC 15-minute candles
CREATE MATERIALIZED VIEW ohlc_15m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('15 minutes', time) AS time,
    symbol,
    first(open, time) AS open,
    max(high) AS high,
    min(low) AS low,
    last(close, time) AS close,
    sum(volume) AS volume,
    sum(tick_count) AS tick_count
FROM ohlc_1m
GROUP BY time_bucket('15 minutes', time), symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('ohlc_15m',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '15 minutes',
    schedule_interval => INTERVAL '15 minutes');

-- OHLC 1-hour candles
CREATE MATERIALIZED VIEW ohlc_1h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS time,
    symbol,
    first(open, time) AS open,
    max(high) AS high,
    min(low) AS low,
    last(close, time) AS close,
    sum(volume) AS volume,
    sum(tick_count) AS tick_count
FROM ohlc_1m
GROUP BY time_bucket('1 hour', time), symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('ohlc_1h',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- OHLC 4-hour candles
CREATE MATERIALIZED VIEW ohlc_4h
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('4 hours', time) AS time,
    symbol,
    first(open, time) AS open,
    max(high) AS high,
    min(low) AS low,
    last(close, time) AS close,
    sum(volume) AS volume,
    sum(tick_count) AS tick_count
FROM ohlc_1m
GROUP BY time_bucket('4 hours', time), symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('ohlc_4h',
    start_offset => INTERVAL '14 days',
    end_offset => INTERVAL '4 hours',
    schedule_interval => INTERVAL '4 hours');

-- OHLC daily candles
CREATE MATERIALIZED VIEW ohlc_1d
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', time) AS time,
    symbol,
    first(open, time) AS open,
    max(high) AS high,
    min(low) AS low,
    last(close, time) AS close,
    sum(volume) AS volume,
    sum(tick_count) AS tick_count
FROM ohlc_1m
GROUP BY time_bucket('1 day', time), symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('ohlc_1d',
    start_offset => INTERVAL '30 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');

-- Add compression to all OHLC tables
ALTER MATERIALIZED VIEW ohlc_1m_continuous SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

ALTER MATERIALIZED VIEW ohlc_5m SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

ALTER MATERIALIZED VIEW ohlc_15m SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

ALTER MATERIALIZED VIEW ohlc_1h SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

ALTER MATERIALIZED VIEW ohlc_4h SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

ALTER MATERIALIZED VIEW ohlc_1d SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

-- Add compression policies
SELECT add_compression_policy('ohlc_1m_continuous', INTERVAL '30 days');
SELECT add_compression_policy('ohlc_5m', INTERVAL '60 days');
SELECT add_compression_policy('ohlc_15m', INTERVAL '90 days');
SELECT add_compression_policy('ohlc_1h', INTERVAL '180 days');
SELECT add_compression_policy('ohlc_4h', INTERVAL '365 days');
SELECT add_compression_policy('ohlc_1d', INTERVAL '730 days');

-- ============================================================================
-- MARKET DEPTH (Order Book)
-- Purpose: Store market depth snapshots for analysis
-- ============================================================================

CREATE TABLE market_depth (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    level INT NOT NULL, -- 1 = best bid/ask, 2 = second level, etc.
    bid_price DECIMAL(18, 8),
    bid_volume DECIMAL(18, 4),
    ask_price DECIMAL(18, 8),
    ask_volume DECIMAL(18, 4),
    liquidity_provider VARCHAR(50),
    CONSTRAINT market_depth_pkey PRIMARY KEY (time, symbol, level, liquidity_provider)
);

-- Convert to hypertable
SELECT create_hypertable('market_depth', 'time', chunk_time_interval => INTERVAL '1 day');

-- Add compression
ALTER TABLE market_depth SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, level'
);

SELECT add_compression_policy('market_depth', INTERVAL '7 days');
SELECT add_retention_policy('market_depth', INTERVAL '30 days');

COMMENT ON TABLE market_depth IS 'Market depth snapshots for order book analysis';

-- ============================================================================
-- SPREADS TRACKING
-- Purpose: Track bid-ask spreads over time
-- ============================================================================

CREATE TABLE spreads (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    spread DECIMAL(18, 8) NOT NULL,
    spread_percentage DECIMAL(10, 4) NOT NULL,
    mid_price DECIMAL(18, 8) NOT NULL,
    liquidity_provider VARCHAR(50)
);

-- Convert to hypertable
SELECT create_hypertable('spreads', 'time', chunk_time_interval => INTERVAL '1 day');

-- Create continuous aggregate for average spreads per minute
CREATE MATERIALIZED VIEW avg_spreads_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS time,
    symbol,
    avg(spread) AS avg_spread,
    min(spread) AS min_spread,
    max(spread) AS max_spread,
    avg(spread_percentage) AS avg_spread_percentage
FROM spreads
GROUP BY time_bucket('1 minute', time), symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('avg_spreads_1m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- Add compression and retention
ALTER TABLE spreads SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);

SELECT add_compression_policy('spreads', INTERVAL '7 days');
SELECT add_retention_policy('spreads', INTERVAL '30 days');

-- ============================================================================
-- USEFUL VIEWS
-- ============================================================================

-- Latest prices view
CREATE VIEW v_latest_prices AS
SELECT DISTINCT ON (symbol)
    symbol,
    time,
    bid,
    ask,
    (bid + ask) / 2 AS mid_price,
    ask - bid AS spread,
    ((ask - bid) / ((bid + ask) / 2)) * 100 AS spread_percentage
FROM ticks
ORDER BY symbol, time DESC;

-- Symbol statistics (last 24 hours)
CREATE VIEW v_symbol_stats_24h AS
SELECT
    symbol,
    count(*) AS tick_count,
    max(bid) AS high_24h,
    min(bid) AS low_24h,
    first(bid, time) AS open_24h,
    last(bid, time) AS close_24h,
    last(bid, time) - first(bid, time) AS change_24h,
    ((last(bid, time) - first(bid, time)) / first(bid, time)) * 100 AS change_percentage_24h,
    avg(ask - bid) AS avg_spread_24h
FROM ticks
WHERE time > NOW() - INTERVAL '24 hours'
GROUP BY symbol;

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Function to get latest price for a symbol
CREATE OR REPLACE FUNCTION get_latest_price(p_symbol VARCHAR(20))
RETURNS TABLE (
    symbol VARCHAR(20),
    time TIMESTAMPTZ,
    bid DECIMAL(18, 8),
    ask DECIMAL(18, 8),
    mid_price DECIMAL(18, 8),
    spread DECIMAL(18, 8)
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        t.symbol,
        t.time,
        t.bid,
        t.ask,
        (t.bid + t.ask) / 2 AS mid_price,
        t.ask - t.bid AS spread
    FROM ticks t
    WHERE t.symbol = p_symbol
    ORDER BY t.time DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- Function to get OHLC data for a symbol and timeframe
CREATE OR REPLACE FUNCTION get_ohlc_data(
    p_symbol VARCHAR(20),
    p_timeframe VARCHAR(10),
    p_limit INT DEFAULT 100
)
RETURNS TABLE (
    time TIMESTAMPTZ,
    open DECIMAL(18, 8),
    high DECIMAL(18, 8),
    low DECIMAL(18, 8),
    close DECIMAL(18, 8),
    volume DECIMAL(18, 4),
    tick_count INT
) AS $$
BEGIN
    RETURN QUERY EXECUTE format(
        'SELECT time, open, high, low, close, volume, tick_count
         FROM ohlc_%s
         WHERE symbol = $1
         ORDER BY time DESC
         LIMIT $2',
        p_timeframe
    ) USING p_symbol, p_limit;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- PERFORMANCE TUNING
-- ============================================================================

-- Adjust chunk size based on data volume
-- For high-frequency symbols (>10k ticks/day), use smaller chunks
-- For low-frequency symbols, larger chunks are fine

-- Example: Adjust chunk size for BTCUSD (high volume)
-- SELECT set_chunk_time_interval('ticks', INTERVAL '6 hours', 'BTCUSD');

-- ============================================================================
-- MONITORING QUERIES
-- ============================================================================

-- Check hypertable statistics
CREATE VIEW v_hypertable_stats AS
SELECT
    hypertable_name,
    num_chunks,
    num_dimensions,
    compression_enabled,
    total_bytes,
    pg_size_pretty(total_bytes) AS total_size,
    pg_size_pretty(total_bytes - COALESCE(compressed_total_bytes, 0)) AS uncompressed_size,
    pg_size_pretty(COALESCE(compressed_total_bytes, 0)) AS compressed_size,
    CASE
        WHEN total_bytes > 0 THEN
            ROUND((COALESCE(compressed_total_bytes, 0)::NUMERIC / total_bytes::NUMERIC) * 100, 2)
        ELSE 0
    END AS compression_ratio
FROM timescaledb_information.hypertables h
LEFT JOIN timescaledb_information.compressed_hypertable_stats c
    ON h.hypertable_name = c.hypertable_name;

-- Check compression policies
CREATE VIEW v_compression_policies AS
SELECT
    hypertable,
    older_than,
    schedule_interval
FROM timescaledb_information.compression_settings;

-- Check continuous aggregate policies
CREATE VIEW v_continuous_aggregate_policies AS
SELECT
    view_name,
    schedule_interval,
    refresh_lag,
    refresh_interval
FROM timescaledb_information.continuous_aggregate_stats;

-- ============================================================================
-- GRANT PERMISSIONS
-- ============================================================================

-- Grant permissions to application user
GRANT SELECT, INSERT ON ticks TO trading_app;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO trading_app;
GRANT SELECT ON ALL TABLES IN SCHEMA _timescaledb_internal TO trading_app;

-- Grant read-only access
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly_app;
GRANT SELECT ON ALL TABLES IN SCHEMA _timescaledb_internal TO readonly_app;

-- ============================================================================
-- END OF TIMESCALEDB SCHEMA
-- ============================================================================
