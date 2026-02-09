-- Migration: Create workspaces table
-- Description: Store trading workspace configurations with complete state
-- Created: 2024-01-01

-- Create workspaces table
CREATE TABLE IF NOT EXISTS workspaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    user_id TEXT NOT NULL,
    account_id TEXT,

    -- JSON columns for workspace state
    charts TEXT NOT NULL,           -- JSON array of chart configurations
    layout TEXT NOT NULL,            -- JSON object for layout configuration
    market_watch TEXT NOT NULL,      -- JSON object for market watch config
    order_panel TEXT NOT NULL,       -- JSON object for order panel settings

    -- Metadata
    version TEXT NOT NULL DEFAULT '1.0.0',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    thumbnail TEXT,                  -- Base64 encoded screenshot (optional)
    is_default INTEGER NOT NULL DEFAULT 0,
    tags TEXT,                       -- JSON array of tags

    -- Constraints
    CHECK(json_valid(charts)),
    CHECK(json_valid(layout)),
    CHECK(json_valid(market_watch)),
    CHECK(json_valid(order_panel)),
    CHECK(tags IS NULL OR json_valid(tags))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_workspaces_user_id ON workspaces(user_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_account_id ON workspaces(account_id);
CREATE INDEX IF NOT EXISTS idx_workspaces_updated_at ON workspaces(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_workspaces_user_updated ON workspaces(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_workspaces_is_default ON workspaces(user_id, is_default);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_workspaces_timestamp
AFTER UPDATE ON workspaces
BEGIN
    UPDATE workspaces SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert sample workspace for testing
INSERT OR IGNORE INTO workspaces (
    id, name, description, user_id,
    charts, layout, market_watch, order_panel,
    version, is_default
) VALUES (
    'ws_default_001',
    'Default Workspace',
    'Default trading workspace with basic configuration',
    'default',
    '[{"id":"chart1","symbol":"EURUSD","timeframe":"1m","chartType":"candlestick","indicators":[],"drawings":[],"settings":{"showGrid":true,"showVolume":true,"showCrosshair":true,"priceScale":"AUTO","backgroundColor":"#1a1a1a","gridColor":"#2a2a2a","textColor":"#d1d4dc","candleUpColor":"#26a69a","candleDownColor":"#ef5350","wickColor":"#737375","showLegend":true,"showTimescale":true,"timezone":"UTC"}}]',
    '{"version":"1.0.0","theme":"dark","panels":{"marketWatch":{"visible":true,"collapsed":false,"width":300},"orderBook":{"visible":true,"collapsed":false,"width":300},"timeSales":{"visible":true,"collapsed":false,"width":300},"orderEntry":{"visible":true,"collapsed":false,"height":250},"positions":{"visible":true,"collapsed":false,"height":200},"alerts":{"visible":true,"collapsed":false,"height":150},"navigator":{"visible":true,"collapsed":false,"width":250},"toolbars":{"visible":true,"collapsed":false}},"windows":{},"splitRatios":{"leftSidebar":0.2,"rightSidebar":0.2,"bottomPanel":0.3}}',
    '{"symbols":["EURUSD","GBPUSD","USDJPY"],"columns":[],"sortBy":"symbol","sortDirection":"asc","favorites":[],"showSparklines":true,"updateInterval":1000}',
    '{"defaultVolume":0.01,"defaultSlPips":20,"defaultTpPips":40,"confirmOrders":true,"oneClickTrading":false,"showCalculators":true,"riskPercent":1}',
    '1.0.0',
    1
);
