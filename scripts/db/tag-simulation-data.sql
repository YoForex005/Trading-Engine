-- ============================================================================
-- Tag Simulation Data Script
-- ============================================================================
-- Purpose: Tag simulation data with flag for identification/audit
-- DOES NOT delete data - only marks it for tracking
-- Usage: psql trading_engine -f tag-simulation-data.sql
-- ============================================================================

\echo '════════════════════════════════════════════════════════════'
\echo '  Tag Simulation Data for Audit Trail'
\echo '════════════════════════════════════════════════════════════'
\echo ''

-- Start transaction for safety
BEGIN;

\echo '1. Counting simulation ticks to be tagged...'

SELECT
    COUNT(*) as ticks_to_tag
FROM ticks
WHERE (lp_source = 'SIM' OR lp_source = 'SIMULATION')
  AND (flags & 1) = 0;  -- Not already flagged

\echo ''
\echo '2. Preview symbols affected...'

SELECT
    symbol,
    COUNT(*) as tick_count
FROM ticks
WHERE (lp_source = 'SIM' OR lp_source = 'SIMULATION')
  AND (flags & 1) = 0
GROUP BY symbol
ORDER BY COUNT(*) DESC
LIMIT 10;

\echo ''
\echo '3. Tagging simulation data (setting bit 0 in flags)...'

-- Tag simulation data by setting bit 0
UPDATE ticks
SET flags = flags | 1  -- Set bit 0 to indicate simulation
WHERE (lp_source = 'SIM' OR lp_source = 'SIMULATION')
  AND (flags & 1) = 0;  -- Only if not already flagged

\echo ''
\echo '4. Verification...'

-- Verify tagging
SELECT
    COUNT(*) as tagged_simulation_ticks,
    'All simulation ticks are now tagged with flags |= 1' as status
FROM ticks
WHERE (lp_source = 'SIM' OR lp_source = 'SIMULATION')
  AND (flags & 1) = 1;

\echo ''
\echo 'Commit changes? (yes/no)'
\echo 'NOTE: This does NOT delete data, only marks it.'
\echo ''

-- Uncomment the next line to auto-commit, or run \gexec manually
-- COMMIT;

\echo 'Transaction is still open. Run COMMIT to save changes or ROLLBACK to cancel.'
\echo ''
