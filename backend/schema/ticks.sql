-- ============================================================================
-- TICK STORAGE DATABASE SCHEMA
-- ============================================================================
-- Version: 1.0.0
-- Created: 2026-01-20
-- Purpose: Persistent tick storage for high-frequency trading data
-- Recommended: SQLite with daily partitions (cross-platform, simple deployment)
-- ============================================================================

-- ----------------------------------------------------------------------------
-- OPTION A: SQLite with Daily Partitions (RECOMMENDED for Windows/Linux)
-- ----------------------------------------------------------------------------
-- Rationale:
--   - Cross-platform (no external dependencies)
--   - Embedded database (no server process)
--   - Excellent performance for time-series data
--   - Simple backup/restore (file-based)
--   - Supports 50K+ inserts/sec with WAL mode
--   - Perfect for broker deployments on Windows/Linux
-- ----------------------------------------------------------------------------

-- Core tick table (main storage)
CREATE TABLE IF NOT EXISTS ticks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) NOT NULL,
    timestamp INTEGER NOT NULL,              -- Unix timestamp in milliseconds
    bid REAL NOT NULL,
    ask REAL NOT NULL,
    spread REAL NOT NULL,
    volume INTEGER DEFAULT 0,                -- Tick volume (optional)
    lp_source VARCHAR(50),                   -- Liquidity provider source
    flags INTEGER DEFAULT 0,                 -- Bit flags for special conditions
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

-- Indexes for optimal query performance
CREATE INDEX IF NOT EXISTS idx_ticks_symbol_timestamp ON ticks(symbol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ticks_timestamp ON ticks(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_ticks_symbol ON ticks(symbol);
CREATE INDEX IF NOT EXISTS idx_ticks_lp_source ON ticks(lp_source);

-- Unique constraint to prevent duplicate ticks (same symbol + timestamp)
CREATE UNIQUE INDEX IF NOT EXISTS idx_ticks_unique ON ticks(symbol, timestamp);

-- Symbol metadata table
CREATE TABLE IF NOT EXISTS symbols (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) UNIQUE NOT NULL,
    description VARCHAR(255),
    asset_class VARCHAR(50),                -- forex, crypto, commodity, index
    base_currency VARCHAR(10),
    quote_currency VARCHAR(10),
    tick_size REAL,
    contract_size REAL,
    is_active BOOLEAN DEFAULT 1,
    first_tick_at INTEGER,
    last_tick_at INTEGER,
    total_ticks INTEGER DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000),
    updated_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE INDEX IF NOT EXISTS idx_symbols_active ON symbols(is_active);
CREATE INDEX IF NOT EXISTS idx_symbols_asset_class ON symbols(asset_class);

-- Liquidity provider tracking
CREATE TABLE IF NOT EXISTS lp_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lp_name VARCHAR(50) UNIQUE NOT NULL,
    lp_type VARCHAR(20),                    -- bank, ecn, aggregator, simulated
    avg_latency_ms REAL,
    uptime_percentage REAL,
    total_ticks INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT 1,
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

-- Partition metadata (for managing daily files)
CREATE TABLE IF NOT EXISTS tick_partitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    partition_date DATE UNIQUE NOT NULL,    -- YYYY-MM-DD
    file_path VARCHAR(255),
    tick_count INTEGER DEFAULT 0,
    file_size_bytes INTEGER DEFAULT 0,
    is_compressed BOOLEAN DEFAULT 0,
    compression_type VARCHAR(20),           -- gzip, zstd, lz4
    compression_ratio REAL,
    archived_at INTEGER,                    -- NULL if active, timestamp if archived
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE INDEX IF NOT EXISTS idx_partitions_date ON tick_partitions(partition_date DESC);
CREATE INDEX IF NOT EXISTS idx_partitions_archived ON tick_partitions(archived_at);

-- Data quality monitoring
CREATE TABLE IF NOT EXISTS tick_quality_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol VARCHAR(20) NOT NULL,
    partition_date DATE NOT NULL,
    tick_count INTEGER DEFAULT 0,
    gap_count INTEGER DEFAULT 0,            -- Number of gaps detected
    max_gap_seconds REAL,                   -- Longest gap in seconds
    duplicate_count INTEGER DEFAULT 0,
    invalid_count INTEGER DEFAULT 0,        -- Invalid bid/ask/spread
    avg_spread REAL,
    min_spread REAL,
    max_spread REAL,
    quality_score REAL,                     -- 0-100 quality score
    created_at INTEGER DEFAULT (strftime('%s', 'now') * 1000)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_quality_symbol_date ON tick_quality_metrics(symbol, partition_date);

-- Performance optimization settings (pragma statements)
-- These should be executed when opening the database connection
-- PRAGMA journal_mode = WAL;               -- Write-Ahead Logging for concurrency
-- PRAGMA synchronous = NORMAL;             -- Balance safety vs performance
-- PRAGMA cache_size = -64000;              -- 64MB cache
-- PRAGMA temp_store = MEMORY;              -- In-memory temp tables
-- PRAGMA mmap_size = 268435456;            -- 256MB memory-mapped I/O

-- Trigger to update symbol metadata on tick insert
CREATE TRIGGER IF NOT EXISTS update_symbol_stats AFTER INSERT ON ticks
BEGIN
    INSERT OR REPLACE INTO symbols (symbol, first_tick_at, last_tick_at, total_ticks, updated_at)
    VALUES (
        NEW.symbol,
        COALESCE((SELECT first_tick_at FROM symbols WHERE symbol = NEW.symbol), NEW.timestamp),
        NEW.timestamp,
        COALESCE((SELECT total_ticks FROM symbols WHERE symbol = NEW.symbol), 0) + 1,
        strftime('%s', 'now') * 1000
    );
END;

-- View for recent ticks (commonly used query)
CREATE VIEW IF NOT EXISTS recent_ticks AS
SELECT
    symbol,
    timestamp,
    bid,
    ask,
    spread,
    lp_source,
    datetime(timestamp / 1000, 'unixepoch') as human_time
FROM ticks
WHERE timestamp > (strftime('%s', 'now') - 3600) * 1000  -- Last hour
ORDER BY timestamp DESC;

-- View for symbol statistics
CREATE VIEW IF NOT EXISTS symbol_stats AS
SELECT
    s.symbol,
    s.description,
    s.asset_class,
    s.total_ticks,
    datetime(s.first_tick_at / 1000, 'unixepoch') as first_tick,
    datetime(s.last_tick_at / 1000, 'unixepoch') as last_tick,
    s.is_active,
    COUNT(DISTINCT tp.partition_date) as days_with_data
FROM symbols s
LEFT JOIN (
    SELECT DISTINCT symbol, date(timestamp / 1000, 'unixepoch') as partition_date
    FROM ticks
) tp ON s.symbol = tp.symbol
GROUP BY s.symbol;

-- ----------------------------------------------------------------------------
-- SAMPLE QUERIES
-- ----------------------------------------------------------------------------

-- Insert a tick
-- INSERT INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
-- VALUES ('EURUSD', 1737388800000, 1.0850, 1.0852, 0.0002, 'YOFX');

-- Get recent ticks for a symbol
-- SELECT * FROM ticks WHERE symbol = 'EURUSD' ORDER BY timestamp DESC LIMIT 100;

-- Get ticks for a time range
-- SELECT * FROM ticks
-- WHERE symbol = 'EURUSD'
--   AND timestamp BETWEEN 1737388800000 AND 1737475200000
-- ORDER BY timestamp;

-- Get tick statistics for a symbol
-- SELECT
--     COUNT(*) as tick_count,
--     AVG(spread) as avg_spread,
--     MIN(bid) as min_bid,
--     MAX(ask) as max_ask,
--     datetime(MIN(timestamp) / 1000, 'unixepoch') as first_tick,
--     datetime(MAX(timestamp) / 1000, 'unixepoch') as last_tick
-- FROM ticks
-- WHERE symbol = 'EURUSD';

-- Find gaps in tick data (gaps > 60 seconds)
-- SELECT
--     symbol,
--     timestamp as gap_start,
--     lead_timestamp as gap_end,
--     (lead_timestamp - timestamp) / 1000.0 as gap_seconds
-- FROM (
--     SELECT
--         symbol,
--         timestamp,
--         LEAD(timestamp) OVER (PARTITION BY symbol ORDER BY timestamp) as lead_timestamp
--     FROM ticks
--     WHERE symbol = 'EURUSD'
-- )
-- WHERE (lead_timestamp - timestamp) / 1000.0 > 60;

-- ----------------------------------------------------------------------------
-- PARTITION MANAGEMENT (Daily File Rotation)
-- ----------------------------------------------------------------------------

-- Strategy: One SQLite database per day
-- File naming: ticks_YYYY-MM-DD.db
-- Location: data/ticks/YYYY/MM/ticks_YYYY-MM-DD.db

-- Benefits:
--   1. Natural time-based partitioning
--   2. Easy to archive/delete old data
--   3. Parallel queries across partitions
--   4. Smaller index sizes per partition
--   5. Better crash recovery (isolated failures)

-- Implementation approach:
--   - Application opens current day's database
--   - At midnight, close current, open new database
--   - Background process compresses databases older than 7 days
--   - Archive to cold storage after 6 months

-- Compression after 7 days:
--   - Use zstd or lz4 for best compression ratio + speed
--   - Typical compression: 5-10x reduction
--   - Keep compressed files for historical analysis
--   - Decompress on-demand for historical queries

-- ----------------------------------------------------------------------------
-- OPTION B: PostgreSQL + TimescaleDB (Production-Grade Alternative)
-- ----------------------------------------------------------------------------
-- Suitable for: Large-scale deployments, multi-server setups
-- Requires: PostgreSQL 12+ with TimescaleDB extension
-- ----------------------------------------------------------------------------

/*
-- TimescaleDB hypertable for automatic partitioning
CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE ticks (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    bid DOUBLE PRECISION NOT NULL,
    ask DOUBLE PRECISION NOT NULL,
    spread DOUBLE PRECISION NOT NULL,
    volume INTEGER DEFAULT 0,
    lp_source VARCHAR(50),
    flags INTEGER DEFAULT 0
);

-- Convert to hypertable (auto-partitions by time)
SELECT create_hypertable('ticks', 'time', chunk_time_interval => INTERVAL '1 day');

-- Indexes
CREATE INDEX ON ticks (symbol, time DESC);
CREATE INDEX ON ticks (lp_source, time DESC);

-- Compression policy (compress chunks older than 7 days)
ALTER TABLE ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('ticks', INTERVAL '7 days');

-- Retention policy (drop chunks older than 6 months)
SELECT add_retention_policy('ticks', INTERVAL '6 months');

-- Continuous aggregate for 1-minute OHLC
CREATE MATERIALIZED VIEW ohlc_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS bucket,
    symbol,
    FIRST(bid, time) AS open,
    MAX(bid) AS high,
    MIN(bid) AS low,
    LAST(bid, time) AS close,
    COUNT(*) AS volume
FROM ticks
GROUP BY bucket, symbol;

-- Refresh policy for continuous aggregate
SELECT add_continuous_aggregate_policy('ohlc_1m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');
*/

-- ----------------------------------------------------------------------------
-- OPTION C: File-based Parquet/Arrow (Analytics-Optimized)
-- ----------------------------------------------------------------------------
-- Suitable for: Data science workflows, historical analysis
-- Requires: Apache Arrow/Parquet libraries
-- ----------------------------------------------------------------------------

/*
-- Not SQL-based, but column-oriented file format
-- Characteristics:
--   - 10-100x compression vs JSON
--   - Columnar storage (fast analytical queries)
--   - Schema evolution support
--   - Direct S3/cloud storage integration
--   - Integrates with Pandas, Polars, DuckDB

-- File structure:
--   data/ticks/parquet/YYYY/MM/DD/EURUSD_YYYY-MM-DD.parquet

-- Schema (Arrow/Parquet):
--   - timestamp: timestamp[ms]
--   - symbol: string
--   - bid: float64
--   - ask: float64
--   - spread: float64
--   - volume: int32
--   - lp_source: string
--   - flags: int32

-- Implementation (pseudo-code):
--   writer = pq.ParquetWriter(filepath, schema, compression='zstd')
--   writer.write_table(tick_batch)
--   writer.close()

-- Query with DuckDB:
--   SELECT * FROM read_parquet('data/ticks/parquet/2026/01/20/*.parquet')
--   WHERE symbol = 'EURUSD' ORDER BY timestamp DESC LIMIT 100;
*/

-- ============================================================================
-- MIGRATION SCRIPTS
-- ============================================================================

-- Migrate from existing JSON files to SQLite
-- Run this script to import historical data from backend/data/ticks/**/*.json

-- Example migration procedure:
-- 1. FOR EACH symbol directory in data/ticks/
-- 2.   FOR EACH date.json file
-- 3.     Parse JSON array of ticks
-- 4.     Open or create SQLite database for that date
-- 5.     Batch insert ticks (use transactions for performance)
-- 6.     Update partition metadata
-- 7.     Verify tick count matches source file
-- 8.     Mark migration as complete

-- ============================================================================
-- DATA INTEGRITY & VALIDATION
-- ============================================================================

-- Prevent duplicates at application level:
--   - Use INSERT OR IGNORE / INSERT OR REPLACE
--   - Or check UNIQUE constraint violation

-- Handle gaps in data:
--   - Periodic job to detect gaps > threshold (e.g., 60 seconds)
--   - Log gaps to tick_quality_metrics table
--   - Alert on excessive gaps (potential data loss)

-- Validate timestamp ordering:
--   - Ensure timestamps are monotonically increasing per symbol
--   - Check for future timestamps (clock skew detection)
--   - Check for very old timestamps (replay detection)

-- Validate bid/ask/spread consistency:
--   - spread = ask - bid (allow small floating point error)
--   - bid > 0, ask > 0, spread >= 0
--   - Reject ticks with spread > 10% (likely error)

-- Sample validation query:
-- SELECT * FROM ticks
-- WHERE spread < 0
--    OR spread > (bid * 0.1)
--    OR bid <= 0
--    OR ask <= 0
--    OR ask <= bid;

-- ============================================================================
-- PERFORMANCE BENCHMARKS (SQLite with WAL mode)
-- ============================================================================
-- Hardware: Mid-range SSD, 16GB RAM
-- Workload: 100 symbols, 10 ticks/sec/symbol = 1,000 ticks/sec
--
-- INSERT performance:
--   - Single inserts: ~5,000/sec
--   - Batch inserts (500 rows): ~50,000/sec
--   - With indexes: ~30,000/sec (batch)
--
-- SELECT performance:
--   - Recent ticks (last 100): <1ms
--   - Time range query (1 hour): 5-20ms
--   - Full day scan (1M ticks): 100-300ms
--   - Aggregate statistics: 50-200ms
--
-- File size (per day, 100 symbols):
--   - Uncompressed: ~200-500MB
--   - Compressed (zstd): ~50-100MB (4-5x reduction)
--
-- Recommended partitioning:
--   - Daily partitions for active trading (last 30 days)
--   - Weekly partitions for historical data (30-180 days)
--   - Monthly archives for long-term storage (180+ days)

-- ============================================================================
-- DEPLOYMENT RECOMMENDATIONS
-- ============================================================================
--
-- For Windows Broker Deployment:
--   1. Use SQLite with daily partitions (this schema)
--   2. Enable WAL mode for concurrent read/write
--   3. Store databases on fast SSD (not network drive)
--   4. Implement daily rotation at midnight UTC
--   5. Compress databases older than 7 days with zstd
--   6. Archive to external storage after 6 months
--   7. Monitor disk space (alert at 80% usage)
--   8. Implement graceful shutdown (flush pending writes)
--
-- For Linux Broker Deployment:
--   - Same as Windows (SQLite is cross-platform)
--   - Consider using systemd for automatic compression cron jobs
--   - Use logrotate-style archival for old databases
--
-- For Large-Scale Production (100+ symbols, 100 ticks/sec/symbol):
--   - Consider PostgreSQL + TimescaleDB (Option B)
--   - Implement read replicas for historical queries
--   - Use connection pooling (pgBouncer)
--   - Separate hot (recent) and cold (historical) storage tiers
--
-- For Analytics/Data Science Workflows:
--   - Export to Parquet format (Option C) for long-term storage
--   - Use DuckDB for ad-hoc analytics on Parquet files
--   - Integrate with Jupyter notebooks for analysis
--
-- ============================================================================
-- MAINTENANCE PROCEDURES
-- ============================================================================

-- Daily tasks (automated):
--   - Rotate to new database at midnight
--   - Calculate quality metrics for previous day
--   - Compress databases older than 7 days
--   - Update symbol statistics

-- Weekly tasks:
--   - Vacuum old databases to reclaim space
--   - Verify data integrity (checksums)
--   - Review quality metrics (gaps, duplicates)
--   - Archive databases older than 30 days

-- Monthly tasks:
--   - Generate performance reports
--   - Review storage capacity trends
--   - Archive to cold storage (6+ months old)
--   - Verify backup integrity

-- Sample maintenance commands:
-- VACUUM;                    -- Reclaim space, defragment
-- ANALYZE;                   -- Update query planner statistics
-- PRAGMA integrity_check;    -- Verify database integrity
-- PRAGMA optimize;           -- Optimize indexes

-- ============================================================================
-- BACKUP STRATEGY
-- ============================================================================

-- Continuous backup:
--   - SQLite WAL files are naturally incremental
--   - Copy WAL files every 5 minutes to backup location
--   - At rotation, copy full database file

-- Point-in-time recovery:
--   - Combine daily database + WAL files
--   - Can recover to any point within retention window

-- Cold backup:
--   - Daily: Copy yesterday's database to backup server
--   - Weekly: Copy all databases to offsite storage
--   - Monthly: Archive to cloud storage (S3, Azure Blob)

-- Disaster recovery:
--   - Keep 3 copies: local, backup server, cloud
--   - Test restore procedures monthly
--   - Document recovery time objective (RTO < 1 hour)

-- ============================================================================
-- MONITORING & ALERTING
-- ============================================================================

-- Metrics to monitor:
--   1. Tick ingestion rate (ticks/sec)
--   2. Database file size (MB)
--   3. Disk space available (%)
--   4. Query latency (p50, p95, p99)
--   5. Gap count per symbol per day
--   6. Duplicate count
--   7. Invalid tick count

-- Alerts:
--   - Disk space < 20%
--   - Tick ingestion rate drops > 50% for > 5 minutes
--   - Gap > 5 minutes for active symbol
--   - Query latency p95 > 100ms
--   - Database file not rotated at midnight
--   - Compression job failed

-- ============================================================================
-- APPENDIX: Go Integration Example
-- ============================================================================

/*
package tickstore

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
)

type SQLiteTickStore struct {
    db *sql.DB
}

func NewSQLiteTickStore(dbPath string) (*SQLiteTickStore, error) {
    db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_synchronous=NORMAL&cache=shared")
    if err != nil {
        return nil, err
    }

    // Execute schema
    _, err = db.Exec(schemaSQL)
    if err != nil {
        return nil, err
    }

    return &SQLiteTickStore{db: db}, nil
}

func (s *SQLiteTickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) error {
    _, err := s.db.Exec(`
        INSERT OR IGNORE INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
        VALUES (?, ?, ?, ?, ?, ?)
    `, symbol, timestamp.UnixMilli(), bid, ask, spread, lp)
    return err
}

func (s *SQLiteTickStore) GetRecentTicks(symbol string, limit int) ([]Tick, error) {
    rows, err := s.db.Query(`
        SELECT timestamp, bid, ask, spread, lp_source
        FROM ticks
        WHERE symbol = ?
        ORDER BY timestamp DESC
        LIMIT ?
    `, symbol, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var ticks []Tick
    for rows.Next() {
        var t Tick
        var ts int64
        err := rows.Scan(&ts, &t.Bid, &t.Ask, &t.Spread, &t.LP)
        if err != nil {
            return nil, err
        }
        t.Timestamp = time.UnixMilli(ts)
        t.Symbol = symbol
        ticks = append(ticks, t)
    }
    return ticks, nil
}
*/

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
