-- ============================================================================
-- TICK STORAGE MAINTENANCE QUERIES
-- ============================================================================
-- Purpose: Common maintenance, monitoring, and troubleshooting queries
-- Usage: Run as needed for database health checks and optimization
-- ============================================================================

-- ----------------------------------------------------------------------------
-- DAILY MAINTENANCE
-- ----------------------------------------------------------------------------

-- 1. Analyze tables to update query planner statistics
ANALYZE;

-- 2. Check database integrity
PRAGMA integrity_check;

-- 3. Get database statistics
SELECT
    'Database size (MB)' as metric,
    ROUND(page_count * page_size / 1024.0 / 1024.0, 2) as value
FROM pragma_page_count(), pragma_page_size()
UNION ALL
SELECT
    'Free pages',
    freelist_count
FROM pragma_freelist_count()
UNION ALL
SELECT
    'Total ticks',
    COUNT(*)
FROM ticks;

-- 4. Get per-symbol statistics
SELECT
    symbol,
    COUNT(*) as tick_count,
    ROUND(AVG(spread), 6) as avg_spread,
    ROUND(MIN(spread), 6) as min_spread,
    ROUND(MAX(spread), 6) as max_spread,
    datetime(MIN(timestamp) / 1000, 'unixepoch') as first_tick,
    datetime(MAX(timestamp) / 1000, 'unixepoch') as last_tick,
    COUNT(DISTINCT lp_source) as lp_sources
FROM ticks
GROUP BY symbol
ORDER BY tick_count DESC;

-- 5. Find symbols with recent activity (last hour)
SELECT
    symbol,
    COUNT(*) as recent_ticks,
    datetime(MAX(timestamp) / 1000, 'unixepoch') as last_tick
FROM ticks
WHERE timestamp > (strftime('%s', 'now') - 3600) * 1000
GROUP BY symbol
ORDER BY recent_ticks DESC;

-- ----------------------------------------------------------------------------
-- DATA QUALITY CHECKS
-- ----------------------------------------------------------------------------

-- 6. Find invalid ticks (negative spread, invalid prices)
SELECT
    symbol,
    timestamp,
    bid,
    ask,
    spread,
    CASE
        WHEN spread < 0 THEN 'Negative spread'
        WHEN bid <= 0 THEN 'Invalid bid'
        WHEN ask <= 0 THEN 'Invalid ask'
        WHEN ask <= bid THEN 'Ask <= Bid'
        WHEN ABS(spread - (ask - bid)) > 0.00001 THEN 'Spread mismatch'
        WHEN spread > (bid * 0.1) THEN 'Excessive spread (>10%)'
        ELSE 'Unknown'
    END as issue
FROM ticks
WHERE spread < 0
   OR bid <= 0
   OR ask <= 0
   OR ask <= bid
   OR ABS(spread - (ask - bid)) > 0.00001
   OR spread > (bid * 0.1)
ORDER BY timestamp DESC
LIMIT 100;

-- 7. Detect duplicate ticks (same symbol + timestamp)
SELECT
    symbol,
    timestamp,
    COUNT(*) as duplicate_count,
    GROUP_CONCAT(DISTINCT bid || '/' || ask) as prices
FROM ticks
GROUP BY symbol, timestamp
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC
LIMIT 100;

-- 8. Find gaps in tick data (gaps > 60 seconds)
WITH tick_gaps AS (
    SELECT
        symbol,
        timestamp,
        LEAD(timestamp) OVER (PARTITION BY symbol ORDER BY timestamp) as next_timestamp,
        (LEAD(timestamp) OVER (PARTITION BY symbol ORDER BY timestamp) - timestamp) / 1000.0 as gap_seconds
    FROM ticks
)
SELECT
    symbol,
    datetime(timestamp / 1000, 'unixepoch') as gap_start,
    datetime(next_timestamp / 1000, 'unixepoch') as gap_end,
    ROUND(gap_seconds, 2) as gap_seconds
FROM tick_gaps
WHERE gap_seconds > 60
ORDER BY gap_seconds DESC
LIMIT 100;

-- 9. Check for future timestamps (clock skew)
SELECT
    symbol,
    timestamp,
    datetime(timestamp / 1000, 'unixepoch') as tick_time,
    (timestamp / 1000) - strftime('%s', 'now') as seconds_in_future
FROM ticks
WHERE timestamp / 1000 > strftime('%s', 'now') + 60  -- More than 1 minute in future
ORDER BY timestamp DESC
LIMIT 100;

-- 10. Check for very old timestamps (potential replay)
SELECT
    symbol,
    timestamp,
    datetime(timestamp / 1000, 'unixepoch') as tick_time,
    strftime('%s', 'now') - (timestamp / 1000) as seconds_old
FROM ticks
WHERE timestamp / 1000 < strftime('%s', 'now') - 86400  -- Older than 24 hours
ORDER BY timestamp
LIMIT 100;

-- ----------------------------------------------------------------------------
-- PERFORMANCE MONITORING
-- ----------------------------------------------------------------------------

-- 11. Index usage statistics
SELECT name, tbl_name
FROM sqlite_master
WHERE type = 'index'
ORDER BY tbl_name, name;

-- 12. Table sizes
SELECT
    name,
    ROUND(SUM(pgsize) / 1024.0 / 1024.0, 2) as size_mb
FROM dbstat
GROUP BY name
ORDER BY size_mb DESC;

-- 13. Largest symbols by tick count
SELECT
    symbol,
    COUNT(*) as tick_count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM ticks), 2) as percentage
FROM ticks
GROUP BY symbol
ORDER BY tick_count DESC
LIMIT 20;

-- 14. Hourly tick distribution
SELECT
    strftime('%Y-%m-%d %H:00', timestamp / 1000, 'unixepoch') as hour,
    COUNT(*) as tick_count,
    COUNT(DISTINCT symbol) as symbol_count,
    ROUND(AVG(spread), 6) as avg_spread
FROM ticks
GROUP BY hour
ORDER BY hour DESC
LIMIT 24;

-- 15. LP source distribution
SELECT
    COALESCE(lp_source, 'unknown') as lp_source,
    COUNT(*) as tick_count,
    COUNT(DISTINCT symbol) as symbol_count,
    ROUND(AVG(spread), 6) as avg_spread
FROM ticks
GROUP BY lp_source
ORDER BY tick_count DESC;

-- ----------------------------------------------------------------------------
-- WEEKLY MAINTENANCE
-- ----------------------------------------------------------------------------

-- 16. Vacuum to reclaim space (run after deletions)
-- VACUUM;

-- 17. Optimize database
-- PRAGMA optimize;

-- 18. Rebuild indexes (if corruption suspected)
-- REINDEX;

-- 19. Export symbol statistics to CSV
.mode csv
.output symbol_stats.csv
SELECT
    symbol,
    COUNT(*) as tick_count,
    ROUND(AVG(spread), 6) as avg_spread,
    ROUND(MIN(bid), 5) as min_bid,
    ROUND(MAX(ask), 5) as max_ask,
    datetime(MIN(timestamp) / 1000, 'unixepoch') as first_tick,
    datetime(MAX(timestamp) / 1000, 'unixepoch') as last_tick,
    COUNT(DISTINCT lp_source) as lp_sources
FROM ticks
GROUP BY symbol
ORDER BY symbol;
.output stdout

-- 20. Generate daily summary report
.mode table
SELECT '=== Daily Tick Storage Report ===' as report;
SELECT datetime('now') as generated_at;
SELECT '' as separator;

SELECT 'Total Ticks: ' || COUNT(*) FROM ticks;
SELECT 'Total Symbols: ' || COUNT(DISTINCT symbol) FROM ticks;
SELECT 'LP Sources: ' || COUNT(DISTINCT lp_source) FROM ticks;

SELECT '' as separator;
SELECT '=== Top 10 Symbols by Volume ===' as section;
SELECT
    symbol,
    COUNT(*) as ticks,
    ROUND(AVG(spread), 6) as avg_spread
FROM ticks
GROUP BY symbol
ORDER BY ticks DESC
LIMIT 10;

-- ----------------------------------------------------------------------------
-- TROUBLESHOOTING
-- ----------------------------------------------------------------------------

-- 21. Find locked tables
PRAGMA lock_status;

-- 22. Check WAL status
PRAGMA journal_mode;
PRAGMA wal_autocheckpoint;
PRAGMA wal_checkpoint;

-- 23. Get configuration
SELECT
    'journal_mode' as setting,
    (SELECT * FROM pragma_journal_mode) as value
UNION ALL
SELECT 'synchronous', (SELECT * FROM pragma_synchronous)
UNION ALL
SELECT 'cache_size', (SELECT * FROM pragma_cache_size)
UNION ALL
SELECT 'temp_store', (SELECT * FROM pragma_temp_store)
UNION ALL
SELECT 'page_size', (SELECT * FROM pragma_page_size);

-- 24. Find slow queries (requires query plan)
EXPLAIN QUERY PLAN
SELECT * FROM ticks
WHERE symbol = 'EURUSD'
  AND timestamp BETWEEN 1737388800000 AND 1737475200000
ORDER BY timestamp DESC
LIMIT 100;

-- 25. Check foreign key integrity (if using FKs)
PRAGMA foreign_key_check;

-- ----------------------------------------------------------------------------
-- CLEANUP OPERATIONS
-- ----------------------------------------------------------------------------

-- 26. Delete ticks older than 30 days (with backup first!)
-- DELETE FROM ticks
-- WHERE timestamp < (strftime('%s', 'now') - 30*86400) * 1000;

-- 27. Delete specific symbol's ticks
-- DELETE FROM ticks WHERE symbol = 'OBSOLETE_SYMBOL';

-- 28. Compress old ticks by sampling (keep every Nth tick)
-- DELETE FROM ticks
-- WHERE timestamp < (strftime('%s', 'now') - 30*86400) * 1000
--   AND id % 10 != 0;  -- Keep 1 out of 10 ticks

-- 29. Remove duplicate ticks (keep first occurrence)
-- DELETE FROM ticks
-- WHERE id NOT IN (
--     SELECT MIN(id)
--     FROM ticks
--     GROUP BY symbol, timestamp
-- );

-- 30. Clear all ticks (DANGER!)
-- DELETE FROM ticks;
-- VACUUM;

-- ----------------------------------------------------------------------------
-- MIGRATION QUERIES
-- ----------------------------------------------------------------------------

-- 31. Check migration log
SELECT
    source_file,
    target_db,
    ticks_migrated,
    migration_time as migration_time_ms,
    datetime(migrated_at / 1000, 'unixepoch') as migrated_at
FROM migration_log
ORDER BY migrated_at DESC
LIMIT 50;

-- 32. Migration summary
SELECT
    DATE(migrated_at / 1000, 'unixepoch') as migration_date,
    COUNT(*) as files_migrated,
    SUM(ticks_migrated) as total_ticks,
    ROUND(AVG(migration_time), 2) as avg_time_ms
FROM migration_log
GROUP BY migration_date
ORDER BY migration_date DESC;

-- ----------------------------------------------------------------------------
-- ARCHIVAL QUERIES
-- ----------------------------------------------------------------------------

-- 33. Export ticks to JSON (for specific date range)
.mode json
.output export_ticks.json
SELECT
    symbol,
    timestamp,
    bid,
    ask,
    spread,
    lp_source,
    datetime(timestamp / 1000, 'unixepoch') as human_time
FROM ticks
WHERE timestamp BETWEEN 1737388800000 AND 1737475200000
ORDER BY timestamp;
.output stdout

-- 34. Export OHLC data (1-minute bars)
.mode csv
.output ohlc_1m.csv
WITH bars AS (
    SELECT
        symbol,
        (timestamp / 60000) * 60000 as bar_timestamp,
        bid
    FROM ticks
)
SELECT
    symbol,
    datetime(bar_timestamp / 1000, 'unixepoch') as time,
    FIRST_VALUE(bid) OVER (PARTITION BY symbol, bar_timestamp ORDER BY timestamp) as open,
    MAX(bid) OVER (PARTITION BY symbol, bar_timestamp) as high,
    MIN(bid) OVER (PARTITION BY symbol, bar_timestamp) as low,
    LAST_VALUE(bid) OVER (PARTITION BY symbol, bar_timestamp ORDER BY timestamp) as close,
    COUNT(*) OVER (PARTITION BY symbol, bar_timestamp) as volume
FROM bars
GROUP BY symbol, bar_timestamp
ORDER BY bar_timestamp DESC;
.output stdout

-- ============================================================================
-- END OF MAINTENANCE QUERIES
-- ============================================================================
