-- ========================================
-- WORKSPACE SCHEMA VALIDATION SCRIPT
-- ========================================
-- Purpose: Validate workspace persistence schema after migration
-- Usage: psql -d trading_engine -f validate_workspace_schema.sql
-- ========================================

\echo 'Starting workspace schema validation...'
\echo ''

-- ========================================
-- 1. CHECK TABLE EXISTENCE
-- ========================================
\echo '=== Checking Table Existence ==='

SELECT
    CASE
        WHEN COUNT(*) = 6 THEN '✓ PASS: All 6 tables exist'
        ELSE '✗ FAIL: Expected 6 tables, found ' || COUNT(*)::text
    END AS table_check
FROM information_schema.tables
WHERE table_schema = 'public'
    AND table_name IN (
        'workspaces',
        'workspace_charts',
        'user_print_preferences',
        'workspace_templates',
        'workspace_snapshots',
        'workspace_sharing'
    );

\echo ''

-- ========================================
-- 2. CHECK TABLE STRUCTURES
-- ========================================
\echo '=== Checking Table Structures ==='

-- Workspaces columns
SELECT
    CASE
        WHEN COUNT(*) >= 7 THEN '✓ PASS: workspaces has ' || COUNT(*)::text || ' columns'
        ELSE '✗ FAIL: workspaces missing columns'
    END AS workspaces_columns
FROM information_schema.columns
WHERE table_name = 'workspaces';

-- Workspace charts columns
SELECT
    CASE
        WHEN COUNT(*) >= 11 THEN '✓ PASS: workspace_charts has ' || COUNT(*)::text || ' columns'
        ELSE '✗ FAIL: workspace_charts missing columns'
    END AS workspace_charts_columns
FROM information_schema.columns
WHERE table_name = 'workspace_charts';

-- Print preferences columns
SELECT
    CASE
        WHEN COUNT(*) >= 20 THEN '✓ PASS: user_print_preferences has ' || COUNT(*)::text || ' columns'
        ELSE '✗ FAIL: user_print_preferences missing columns'
    END AS print_preferences_columns
FROM information_schema.columns
WHERE table_name = 'user_print_preferences';

-- Templates columns
SELECT
    CASE
        WHEN COUNT(*) >= 12 THEN '✓ PASS: workspace_templates has ' || COUNT(*)::text || ' columns'
        ELSE '✗ FAIL: workspace_templates missing columns'
    END AS templates_columns
FROM information_schema.columns
WHERE table_name = 'workspace_templates';

-- Snapshots columns
SELECT
    CASE
        WHEN COUNT(*) >= 9 THEN '✓ PASS: workspace_snapshots has ' || COUNT(*)::text || ' columns'
        ELSE '✗ FAIL: workspace_snapshots missing columns'
    END AS snapshots_columns
FROM information_schema.columns
WHERE table_name = 'workspace_snapshots';

-- Sharing columns
SELECT
    CASE
        WHEN COUNT(*) >= 9 THEN '✓ PASS: workspace_sharing has ' || COUNT(*)::text || ' columns'
        ELSE '✗ FAIL: workspace_sharing missing columns'
    END AS sharing_columns
FROM information_schema.columns
WHERE table_name = 'workspace_sharing';

\echo ''

-- ========================================
-- 3. CHECK INDEXES
-- ========================================
\echo '=== Checking Indexes ==='

SELECT
    CASE
        WHEN COUNT(*) >= 30 THEN '✓ PASS: Found ' || COUNT(*)::text || ' indexes'
        ELSE '✗ WARN: Expected ~30+ indexes, found ' || COUNT(*)::text
    END AS index_check
FROM pg_indexes
WHERE schemaname = 'public'
    AND tablename LIKE 'workspace%';

-- Check critical indexes
SELECT
    indexname,
    tablename,
    '✓ EXISTS' AS status
FROM pg_indexes
WHERE schemaname = 'public'
    AND indexname IN (
        'idx_workspaces_user_id',
        'idx_workspaces_user_single_default',
        'idx_workspace_charts_workspace_id',
        'idx_workspace_charts_indicators',
        'idx_workspace_templates_public'
    )
ORDER BY indexname;

\echo ''

-- ========================================
-- 4. CHECK CONSTRAINTS
-- ========================================
\echo '=== Checking Constraints ==='

SELECT
    conname AS constraint_name,
    conrelid::regclass AS table_name,
    CASE contype
        WHEN 'c' THEN '✓ CHECK'
        WHEN 'f' THEN '✓ FOREIGN KEY'
        WHEN 'p' THEN '✓ PRIMARY KEY'
        WHEN 'u' THEN '✓ UNIQUE'
        ELSE contype::text
    END AS constraint_type
FROM pg_constraint
WHERE connamespace = 'public'::regnamespace
    AND conrelid::regclass::text LIKE 'workspace%'
ORDER BY table_name, constraint_type, constraint_name;

\echo ''

-- ========================================
-- 5. CHECK TRIGGERS
-- ========================================
\echo '=== Checking Triggers ==='

SELECT
    CASE
        WHEN COUNT(*) >= 5 THEN '✓ PASS: Found ' || COUNT(*)::text || ' triggers'
        ELSE '✗ FAIL: Expected 5+ triggers, found ' || COUNT(*)::text
    END AS trigger_check
FROM pg_trigger
WHERE tgrelid IN (
    'workspaces'::regclass,
    'workspace_charts'::regclass,
    'workspace_templates'::regclass,
    'user_print_preferences'::regclass
)
AND tgname NOT LIKE 'RI_%'; -- Exclude internal triggers

-- List triggers
SELECT
    tgname AS trigger_name,
    tgrelid::regclass AS table_name,
    proname AS function_name,
    '✓ ACTIVE' AS status
FROM pg_trigger
JOIN pg_proc ON pg_trigger.tgfoid = pg_proc.oid
WHERE tgrelid IN (
    'workspaces'::regclass,
    'workspace_charts'::regclass,
    'workspace_templates'::regclass,
    'user_print_preferences'::regclass
)
AND tgname NOT LIKE 'RI_%'
ORDER BY table_name, trigger_name;

\echo ''

-- ========================================
-- 6. CHECK FUNCTIONS
-- ========================================
\echo '=== Checking Functions ==='

SELECT
    CASE
        WHEN COUNT(*) >= 5 THEN '✓ PASS: Found ' || COUNT(*)::text || ' functions'
        ELSE '✗ FAIL: Expected 5 functions, found ' || COUNT(*)::text
    END AS function_check
FROM pg_proc
WHERE pronamespace = 'public'::regnamespace
    AND proname IN (
        'update_workspace_timestamp',
        'enforce_single_default_workspace',
        'increment_template_usage',
        'clean_expired_snapshots',
        'clean_expired_shares'
    );

-- List functions
SELECT
    proname AS function_name,
    pg_get_function_result(oid) AS return_type,
    '✓ EXISTS' AS status
FROM pg_proc
WHERE pronamespace = 'public'::regnamespace
    AND proname IN (
        'update_workspace_timestamp',
        'enforce_single_default_workspace',
        'increment_template_usage',
        'clean_expired_snapshots',
        'clean_expired_shares'
    )
ORDER BY proname;

\echo ''

-- ========================================
-- 7. CHECK MATERIALIZED VIEWS
-- ========================================
\echo '=== Checking Materialized Views ==='

SELECT
    CASE
        WHEN COUNT(*) = 1 THEN '✓ PASS: workspace_statistics view exists'
        ELSE '✗ FAIL: workspace_statistics view not found'
    END AS matview_check
FROM pg_matviews
WHERE schemaname = 'public'
    AND matviewname = 'workspace_statistics';

\echo ''

-- ========================================
-- 8. CHECK FOREIGN KEY RELATIONSHIPS
-- ========================================
\echo '=== Checking Foreign Key Relationships ==='

SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    rc.delete_rule,
    '✓ VALID' AS status
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
JOIN information_schema.referential_constraints AS rc
    ON rc.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_name LIKE 'workspace%'
ORDER BY tc.table_name, kcu.column_name;

\echo ''

-- ========================================
-- 9. CHECK JSONB COLUMNS
-- ========================================
\echo '=== Checking JSONB Columns ==='

SELECT
    table_name,
    column_name,
    data_type,
    '✓ JSONB' AS status
FROM information_schema.columns
WHERE table_schema = 'public'
    AND table_name LIKE 'workspace%'
    AND data_type = 'jsonb'
ORDER BY table_name, column_name;

\echo ''

-- ========================================
-- 10. CHECK GIN INDEXES (for JSONB)
-- ========================================
\echo '=== Checking GIN Indexes ==='

SELECT
    tablename,
    indexname,
    '✓ GIN INDEX' AS status
FROM pg_indexes
WHERE schemaname = 'public'
    AND tablename LIKE 'workspace%'
    AND indexdef LIKE '%USING gin%'
ORDER BY tablename, indexname;

\echo ''

-- ========================================
-- 11. TEST DATA INSERTION (Optional)
-- ========================================
\echo '=== Running Test Data Insertion ==='

-- This section is commented out by default
-- Uncomment to test actual data insertion

/*
BEGIN;

-- Test: Insert workspace
INSERT INTO workspaces (user_id, name, description, is_default, settings)
VALUES (1, 'Test Workspace', 'Validation test', TRUE, '{"theme": "dark"}'::jsonb)
RETURNING id AS workspace_id;

-- Test: Insert chart
-- INSERT INTO workspace_charts (workspace_id, chart_index, symbol, timeframe, chart_type)
-- VALUES (1, 0, 'EURUSD', '5m', 'candlestick');

-- Test: Insert print preferences
-- INSERT INTO user_print_preferences (user_id, page_size, orientation)
-- VALUES (1, 'A4', 'landscape')
-- ON CONFLICT (user_id) DO NOTHING;

ROLLBACK; -- Don't actually commit test data
\echo '✓ Test data insertion successful (rolled back)'
*/

\echo ''

-- ========================================
-- 12. CHECK COMMENTS
-- ========================================
\echo '=== Checking Table Comments ==='

SELECT
    c.relname AS table_name,
    CASE
        WHEN d.description IS NOT NULL THEN '✓ HAS COMMENT'
        ELSE '- NO COMMENT'
    END AS comment_status,
    LEFT(d.description, 60) AS comment_preview
FROM pg_class c
LEFT JOIN pg_description d ON c.oid = d.objoid AND d.objsubid = 0
WHERE c.relnamespace = 'public'::regnamespace
    AND c.relname LIKE 'workspace%'
    AND c.relkind = 'r' -- Regular tables only
ORDER BY c.relname;

\echo ''

-- ========================================
-- 13. STORAGE SIZE ANALYSIS
-- ========================================
\echo '=== Storage Size Analysis ==='

SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) -
                   pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
FROM pg_tables
WHERE schemaname = 'public'
    AND tablename LIKE 'workspace%'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

\echo ''

-- ========================================
-- 14. VALIDATION SUMMARY
-- ========================================
\echo '=== Validation Summary ==='

SELECT
    'Workspace Persistence Schema' AS component,
    'Migration 009' AS migration,
    (SELECT COUNT(*) FROM information_schema.tables
     WHERE table_schema = 'public' AND table_name LIKE 'workspace%') AS tables_created,
    (SELECT COUNT(*) FROM pg_indexes
     WHERE schemaname = 'public' AND tablename LIKE 'workspace%') AS indexes_created,
    (SELECT COUNT(*) FROM pg_trigger
     WHERE tgrelid IN (
         'workspaces'::regclass,
         'workspace_charts'::regclass,
         'workspace_templates'::regclass,
         'user_print_preferences'::regclass
     ) AND tgname NOT LIKE 'RI_%') AS triggers_created,
    (SELECT COUNT(*) FROM pg_proc
     WHERE pronamespace = 'public'::regnamespace
     AND proname LIKE '%workspace%') AS functions_created,
    (SELECT COUNT(*) FROM pg_matviews
     WHERE schemaname = 'public' AND matviewname = 'workspace_statistics') AS matviews_created,
    '✓ VALIDATION COMPLETE' AS status;

\echo ''
\echo '=== Validation Complete ==='
\echo ''
\echo 'Next steps:'
\echo '1. Review any FAIL or WARN messages above'
\echo '2. Test API endpoints with actual data'
\echo '3. Monitor performance with EXPLAIN ANALYZE'
\echo '4. Set up periodic maintenance jobs (snapshot cleanup, etc.)'
\echo ''
