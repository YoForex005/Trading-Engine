-- ========================================
-- WORKSPACE SCHEMA TEST DATA
-- ========================================
-- Purpose: Sample data for testing workspace persistence features
-- Usage: psql -d trading_engine -f workspace_schema_test_data.sql
-- Note: This script uses user_id = 1. Adjust as needed.
-- ========================================

BEGIN;

\echo 'Inserting test data for workspace persistence...'
\echo ''

-- ========================================
-- 1. CREATE TEST WORKSPACES
-- ========================================
\echo '=== Creating Test Workspaces ==='

-- Default workspace for user 1
INSERT INTO workspaces (user_id, name, description, is_default, settings)
VALUES (
    1,
    'Default Trading Setup',
    'My main trading workspace with multiple timeframes',
    TRUE,
    '{
        "theme": "dark",
        "layoutMode": "grid",
        "autoSave": true,
        "refreshRate": 1000,
        "timezone": "America/New_York",
        "gridRows": 2,
        "gridColumns": 2
    }'::jsonb
)
RETURNING id, name, is_default;

-- Additional workspace
INSERT INTO workspaces (user_id, name, description, is_default, settings)
VALUES (
    1,
    'Forex Scalping',
    'High-frequency EUR/USD scalping setup',
    FALSE,
    '{
        "theme": "dark",
        "layoutMode": "tabs",
        "autoSave": true,
        "refreshRate": 500,
        "timezone": "Europe/London"
    }'::jsonb
)
RETURNING id, name, is_default;

-- Swing trading workspace
INSERT INTO workspaces (user_id, name, description, is_default, settings)
VALUES (
    1,
    'Swing Trading',
    'Multi-day position analysis',
    FALSE,
    '{
        "theme": "light",
        "layoutMode": "grid",
        "autoSave": false,
        "refreshRate": 5000,
        "timezone": "UTC"
    }'::jsonb
)
RETURNING id, name, is_default;

\echo 'Workspaces created successfully'
\echo ''

-- ========================================
-- 2. ADD CHARTS TO WORKSPACES
-- ========================================
\echo '=== Adding Charts to Workspaces ==='

-- Get the first workspace ID (assuming it's the default one)
DO $$
DECLARE
    v_workspace_id BIGINT;
BEGIN
    SELECT id INTO v_workspace_id FROM workspaces WHERE user_id = 1 AND is_default = TRUE LIMIT 1;

    -- Chart 1: EUR/USD 5-minute with SMA and RSI
    INSERT INTO workspace_charts (
        workspace_id, chart_index, symbol, timeframe, chart_type,
        indicators, drawings, settings
    )
    VALUES (
        v_workspace_id,
        0,
        'EURUSD',
        '5m',
        'candlestick',
        '[
            {
                "type": "SMA",
                "period": 20,
                "color": "#FF0000",
                "settings": {
                    "source": "close",
                    "offset": 0,
                    "lineWidth": 2
                },
                "visible": true
            },
            {
                "type": "RSI",
                "period": 14,
                "color": "#00FF00",
                "settings": {
                    "overbought": 70,
                    "oversold": 30,
                    "levels": [30, 50, 70]
                },
                "visible": true
            }
        ]'::jsonb,
        '[
            {
                "type": "trendline",
                "points": [
                    {"time": "2024-01-01T00:00:00Z", "price": 1.0800},
                    {"time": "2024-01-15T00:00:00Z", "price": 1.0850}
                ],
                "color": "#0000FF",
                "style": "solid",
                "width": 2,
                "text": "Support Line"
            }
        ]'::jsonb,
        '{
            "showGrid": true,
            "showVolume": true,
            "showCrosshair": true,
            "backgroundColor": "#1E1E1E",
            "gridColor": "#2C2C2C",
            "textColor": "#FFFFFF",
            "upColor": "#26A69A",
            "downColor": "#EF5350"
        }'::jsonb
    );

    -- Chart 2: EUR/USD 1-hour
    INSERT INTO workspace_charts (
        workspace_id, chart_index, symbol, timeframe, chart_type,
        indicators, drawings, settings
    )
    VALUES (
        v_workspace_id,
        1,
        'EURUSD',
        '1h',
        'candlestick',
        '[
            {
                "type": "EMA",
                "period": 50,
                "color": "#FFAA00",
                "settings": {"source": "close"},
                "visible": true
            },
            {
                "type": "MACD",
                "settings": {
                    "fastPeriod": 12,
                    "slowPeriod": 26,
                    "signalPeriod": 9
                },
                "visible": true
            }
        ]'::jsonb,
        '[]'::jsonb,
        '{
            "showGrid": true,
            "showVolume": true,
            "backgroundColor": "#1E1E1E"
        }'::jsonb
    );

    -- Chart 3: GBP/USD 15-minute
    INSERT INTO workspace_charts (
        workspace_id, chart_index, symbol, timeframe, chart_type,
        indicators, drawings, settings
    )
    VALUES (
        v_workspace_id,
        2,
        'GBPUSD',
        '15m',
        'candlestick',
        '[
            {
                "type": "BollingerBands",
                "period": 20,
                "stdDev": 2,
                "color": "#9C27B0",
                "visible": true
            }
        ]'::jsonb,
        '[
            {
                "type": "rectangle",
                "points": [
                    {"time": "2024-01-05T00:00:00Z", "price": 1.2700},
                    {"time": "2024-01-10T00:00:00Z", "price": 1.2750}
                ],
                "color": "#FFFF00",
                "fillColor": "#FFFF0033",
                "style": "dashed",
                "text": "Supply Zone"
            }
        ]'::jsonb,
        '{
            "showGrid": true,
            "showVolume": false,
            "backgroundColor": "#1E1E1E"
        }'::jsonb
    );

    -- Chart 4: BTC/USD Daily
    INSERT INTO workspace_charts (
        workspace_id, chart_index, symbol, timeframe, chart_type,
        indicators, drawings, settings
    )
    VALUES (
        v_workspace_id,
        3,
        'BTCUSD',
        '1d',
        'candlestick',
        '[
            {
                "type": "SMA",
                "period": 200,
                "color": "#FF0000",
                "settings": {"source": "close"},
                "visible": true
            },
            {
                "type": "Volume",
                "color": "#4CAF50",
                "visible": true
            }
        ]'::jsonb,
        '[]'::jsonb,
        '{
            "showGrid": true,
            "showVolume": true,
            "scaleMode": "logarithmic",
            "backgroundColor": "#1E1E1E"
        }'::jsonb
    );

    RAISE NOTICE 'Charts added to workspace ID: %', v_workspace_id;
END $$;

\echo 'Charts created successfully'
\echo ''

-- ========================================
-- 3. SET PRINT PREFERENCES
-- ========================================
\echo '=== Setting Print Preferences ==='

INSERT INTO user_print_preferences (
    user_id,
    page_size,
    orientation,
    margins,
    include_header,
    include_footer,
    header_template,
    footer_template,
    color_mode,
    include_grid,
    include_crosshair,
    include_volume,
    include_indicators,
    include_drawings,
    show_legend,
    show_title,
    show_timeframe,
    show_symbol,
    image_quality,
    image_dpi,
    pdf_compression,
    watermark_enabled,
    watermark_text,
    watermark_opacity
)
VALUES (
    1,
    'A4',
    'landscape',
    '{"top": 10, "right": 10, "bottom": 10, "left": 10}'::jsonb,
    TRUE,
    TRUE,
    'Trading Analysis Report - {{date}}',
    'Page {{page}} of {{totalPages}} - Confidential',
    'color',
    TRUE,
    FALSE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    95,
    300,
    TRUE,
    TRUE,
    'Trading Engine Pro',
    0.3
)
ON CONFLICT (user_id) DO UPDATE SET
    page_size = EXCLUDED.page_size,
    orientation = EXCLUDED.orientation,
    margins = EXCLUDED.margins;

\echo 'Print preferences set successfully'
\echo ''

-- ========================================
-- 4. CREATE WORKSPACE TEMPLATES
-- ========================================
\echo '=== Creating Workspace Templates ==='

-- Scalping template
INSERT INTO workspace_templates (
    user_id,
    name,
    description,
    template_type,
    configuration,
    is_public,
    tags
)
VALUES (
    1,
    'Forex Scalping Setup',
    'Multi-timeframe setup optimized for EUR/USD scalping',
    'workspace',
    '{
        "name": "Scalping Template",
        "settings": {
            "theme": "dark",
            "layoutMode": "grid",
            "gridRows": 2,
            "gridColumns": 2,
            "refreshRate": 500
        },
        "charts": [
            {
                "symbol": "EURUSD",
                "timeframe": "1m",
                "chartType": "candlestick",
                "indicators": [
                    {"type": "EMA", "period": 9},
                    {"type": "EMA", "period": 21}
                ]
            }
        ]
    }'::jsonb,
    TRUE,
    ARRAY['forex', 'scalping', 'eur-usd', 'short-term']
);

-- Indicator set template
INSERT INTO workspace_templates (
    user_id,
    name,
    description,
    template_type,
    configuration,
    is_public,
    tags
)
VALUES (
    1,
    'Trend Following Indicators',
    'Collection of trend-following indicators',
    'indicator_set',
    '{
        "indicators": [
            {
                "type": "SMA",
                "period": 50,
                "color": "#FF0000",
                "settings": {"source": "close"}
            },
            {
                "type": "SMA",
                "period": 200,
                "color": "#0000FF",
                "settings": {"source": "close"}
            },
            {
                "type": "ADX",
                "period": 14,
                "color": "#00FF00"
            }
        ]
    }'::jsonb,
    TRUE,
    ARRAY['indicators', 'trend-following', 'technical-analysis']
);

-- Color scheme template
INSERT INTO workspace_templates (
    user_id,
    name,
    description,
    template_type,
    configuration,
    is_public,
    tags
)
VALUES (
    1,
    'Dark Blue Theme',
    'Professional dark blue color scheme',
    'color_scheme',
    '{
        "backgroundColor": "#0A1929",
        "gridColor": "#1E3A5F",
        "textColor": "#E3F2FD",
        "upColor": "#26A69A",
        "downColor": "#EF5350",
        "volumeUpColor": "#26A69A80",
        "volumeDownColor": "#EF535080",
        "crosshairColor": "#90CAF9"
    }'::jsonb,
    TRUE,
    ARRAY['theme', 'dark', 'blue', 'professional']
);

\echo 'Templates created successfully'
\echo ''

-- ========================================
-- 5. CREATE WORKSPACE SNAPSHOTS
-- ========================================
\echo '=== Creating Workspace Snapshots ==='

-- Manual snapshot
DO $$
DECLARE
    v_workspace_id BIGINT;
    v_snapshot_id BIGINT;
BEGIN
    SELECT id INTO v_workspace_id FROM workspaces WHERE user_id = 1 AND is_default = TRUE LIMIT 1;

    INSERT INTO workspace_snapshots (
        workspace_id,
        snapshot_name,
        snapshot_type,
        workspace_data,
        charts_data,
        created_by,
        notes
    )
    SELECT
        w.id,
        'Before Major Changes',
        'manual',
        row_to_json(w.*)::jsonb,
        COALESCE(json_agg(wc.* ORDER BY wc.chart_index)::jsonb, '[]'::jsonb),
        1,
        'Snapshot taken before implementing new trading strategy'
    FROM workspaces w
    LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
    WHERE w.id = v_workspace_id
    GROUP BY w.id
    RETURNING id INTO v_snapshot_id;

    RAISE NOTICE 'Snapshot created with ID: %', v_snapshot_id;
END $$;

-- Auto snapshot (with expiration)
DO $$
DECLARE
    v_workspace_id BIGINT;
BEGIN
    SELECT id INTO v_workspace_id FROM workspaces WHERE user_id = 1 AND is_default = TRUE LIMIT 1;

    INSERT INTO workspace_snapshots (
        workspace_id,
        snapshot_type,
        workspace_data,
        charts_data,
        created_by,
        expires_at
    )
    SELECT
        w.id,
        'auto',
        row_to_json(w.*)::jsonb,
        COALESCE(json_agg(wc.* ORDER BY wc.chart_index)::jsonb, '[]'::jsonb),
        1,
        NOW() + INTERVAL '30 days'
    FROM workspaces w
    LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
    WHERE w.id = v_workspace_id
    GROUP BY w.id;

    RAISE NOTICE 'Auto-snapshot created (expires in 30 days)';
END $$;

\echo 'Snapshots created successfully'
\echo ''

-- ========================================
-- 6. CREATE WORKSPACE SHARING
-- ========================================
\echo '=== Creating Workspace Shares ==='

-- Note: This assumes user ID 2 exists
-- Comment out if user 2 doesn't exist in your test environment

/*
DO $$
DECLARE
    v_workspace_id BIGINT;
BEGIN
    SELECT id INTO v_workspace_id FROM workspaces WHERE user_id = 1 AND name = 'Forex Scalping' LIMIT 1;

    -- Share workspace with user 2 (view-only)
    INSERT INTO workspace_sharing (
        workspace_id,
        shared_by,
        shared_with,
        can_view,
        can_edit,
        can_delete,
        message
    )
    VALUES (
        v_workspace_id,
        1,
        2,
        TRUE,
        FALSE,
        FALSE,
        'Check out my scalping setup! Feel free to learn from it.'
    );

    RAISE NOTICE 'Workspace shared with user 2';
END $$;
*/

\echo 'Workspace sharing configured (commented out - adjust user IDs as needed)'
\echo ''

-- ========================================
-- 7. VERIFY DATA
-- ========================================
\echo '=== Verifying Inserted Data ==='

\echo 'Workspaces:'
SELECT id, name, is_default, jsonb_pretty(settings) AS settings
FROM workspaces
WHERE user_id = 1
ORDER BY is_default DESC, created_at;

\echo ''
\echo 'Charts per workspace:'
SELECT
    w.name AS workspace,
    COUNT(wc.id) AS chart_count,
    string_agg(DISTINCT wc.symbol, ', ') AS symbols
FROM workspaces w
LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
WHERE w.user_id = 1
GROUP BY w.id, w.name
ORDER BY w.name;

\echo ''
\echo 'Templates:'
SELECT id, name, template_type, is_public, tags
FROM workspace_templates
WHERE user_id = 1
ORDER BY template_type, name;

\echo ''
\echo 'Snapshots:'
SELECT
    ws.id,
    w.name AS workspace,
    ws.snapshot_name,
    ws.snapshot_type,
    ws.created_at,
    ws.expires_at
FROM workspace_snapshots ws
JOIN workspaces w ON ws.workspace_id = w.id
WHERE w.user_id = 1
ORDER BY ws.created_at DESC;

\echo ''
\echo 'Print preferences:'
SELECT
    page_size,
    orientation,
    color_mode,
    image_quality,
    image_dpi,
    watermark_enabled
FROM user_print_preferences
WHERE user_id = 1;

\echo ''

-- ========================================
-- 8. SAMPLE QUERIES
-- ========================================
\echo '=== Sample Query Results ==='

\echo 'Default workspace with all charts:'
SELECT
    w.name,
    json_agg(
        json_build_object(
            'symbol', wc.symbol,
            'timeframe', wc.timeframe,
            'chartType', wc.chart_type,
            'indicatorCount', jsonb_array_length(wc.indicators),
            'drawingCount', jsonb_array_length(wc.drawings)
        ) ORDER BY wc.chart_index
    ) AS charts
FROM workspaces w
LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
WHERE w.user_id = 1 AND w.is_default = TRUE
GROUP BY w.id, w.name;

\echo ''
\echo 'Charts using SMA indicator:'
SELECT
    w.name AS workspace,
    wc.symbol,
    wc.timeframe,
    wc.indicators
FROM workspace_charts wc
JOIN workspaces w ON wc.workspace_id = w.id
WHERE w.user_id = 1
    AND wc.indicators @> '[{"type": "SMA"}]'
ORDER BY w.name, wc.chart_index;

\echo ''

-- ========================================
-- COMMIT OR ROLLBACK
-- ========================================
\echo '=== Test Data Insertion Complete ==='
\echo ''
\echo 'Review the data above. Type:'
\echo '  COMMIT;   - to keep the test data'
\echo '  ROLLBACK; - to discard the test data'
\echo ''

-- Uncomment one of the following:
-- COMMIT;
-- ROLLBACK;
