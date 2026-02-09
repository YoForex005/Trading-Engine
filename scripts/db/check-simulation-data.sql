-- ============================================================================
-- Database Simulation Data Detection Script
-- ============================================================================
-- Purpose: Identify and report on simulation data in production database
-- Usage: psql trading_engine -f check-simulation-data.sql
-- ============================================================================

\echo '════════════════════════════════════════════════════════════'
\echo '  Simulation Data Detection Report'
\echo '════════════════════════════════════════════════════════════'
\echo ''

-- Check if ticks table exists
\echo '1. Checking if ticks table exists...'
SELECT CASE
    WHEN EXISTS (
        SELECT FROM information_schema.tables
        WHERE table_name = 'ticks'
    )
    THEN '✓ Ticks table found'
    ELSE '✗ Ticks table NOT found'
END as status;

\echo ''
\echo '2. Counting simulation ticks...'

-- Count simulation ticks
SELECT
    COUNT(*) as simulation_tick_count,
    CASE
        WHEN COUNT(*) = 0 THEN '✓ No simulation data found (Production Safe)'
        WHEN COUNT(*) > 0 THEN '✗ ALERT: Simulation data detected!'
        ELSE 'Unknown'
    END as status
FROM ticks
WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION';

\echo ''
\echo '3. LP Source Distribution...'

-- Show tick count by LP source
SELECT
    lp_source as "LP Source",
    COUNT(*) as "Tick Count",
    ROUND(100.0 * COUNT(*) / SUM(COUNT(*)) OVER (), 2) as "Percentage",
    TO_CHAR(MIN(to_timestamp(timestamp / 1000)), 'YYYY-MM-DD HH24:MI:SS') as "First Tick",
    TO_CHAR(MAX(to_timestamp(timestamp / 1000)), 'YYYY-MM-DD HH24:MI:SS') as "Last Tick"
FROM ticks
GROUP BY lp_source
ORDER BY COUNT(*) DESC;

\echo ''
\echo '4. Simulation Ticks by Symbol (if any)...'

-- Show simulation ticks by symbol
SELECT
    symbol as "Symbol",
    COUNT(*) as "Sim Tick Count",
    TO_CHAR(MIN(to_timestamp(timestamp / 1000)), 'YYYY-MM-DD HH24:MI:SS') as "First Sim Tick",
    TO_CHAR(MAX(to_timestamp(timestamp / 1000)), 'YYYY-MM-DD HH24:MI:SS') as "Last Sim Tick"
FROM ticks
WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION'
GROUP BY symbol
ORDER BY COUNT(*) DESC
LIMIT 20;

\echo ''
\echo '5. Recent Ticks (Last 10)...'

-- Show recent ticks to verify LP sources
SELECT
    TO_CHAR(to_timestamp(timestamp / 1000), 'YYYY-MM-DD HH24:MI:SS') as "Timestamp",
    symbol as "Symbol",
    lp_source as "LP Source",
    bid as "Bid",
    ask as "Ask"
FROM ticks
ORDER BY timestamp DESC
LIMIT 10;

\echo ''
\echo '6. Data Quality Check...'

-- Check for flagged simulation data
SELECT
    COUNT(*) as flagged_simulation_count,
    CASE
        WHEN COUNT(*) > 0 THEN 'Simulation data has been flagged (flags & 1 = 1)'
        ELSE 'No flagged simulation data'
    END as status
FROM ticks
WHERE (flags & 1) = 1;

\echo ''
\echo '════════════════════════════════════════════════════════════'
\echo '  Summary'
\echo '════════════════════════════════════════════════════════════'

-- Final summary
WITH summary AS (
    SELECT
        COUNT(*) as total_ticks,
        COUNT(*) FILTER (WHERE lp_source = 'SIM' OR lp_source = 'SIMULATION') as sim_ticks,
        COUNT(*) FILTER (WHERE lp_source NOT IN ('SIM', 'SIMULATION') AND lp_source IS NOT NULL) as real_ticks,
        COUNT(DISTINCT symbol) as total_symbols,
        COUNT(DISTINCT lp_source) as lp_count
    FROM ticks
)
SELECT
    total_ticks as "Total Ticks",
    real_ticks as "Real LP Ticks",
    sim_ticks as "Simulation Ticks",
    total_symbols as "Symbols Tracked",
    lp_count as "LP Sources",
    CASE
        WHEN sim_ticks = 0 THEN '✓ PRODUCTION SAFE'
        WHEN sim_ticks > 0 THEN '✗ PRODUCTION ALERT'
        ELSE 'UNKNOWN'
    END as "Status"
FROM summary;

\echo ''
\echo 'If simulation data was found, run:'
\echo '  psql trading_engine -f tag-simulation-data.sql'
\echo ''
