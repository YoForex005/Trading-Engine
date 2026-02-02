package migrations

import (
	"database/sql"
)

func init() {
	RegisterMigration(&Migration{
		Version: 9,
		Name:    "create_workspace_tables",
		Up:      workspaceTablesUp,
		Down:    workspaceTablesDown,
	})
}

func workspaceTablesUp(tx *sql.Tx) error {
	schema := `
	-- ========================================
	-- WORKSPACE PERSISTENCE SCHEMA
	-- Migration 009: Chart workspace and layout persistence
	-- ========================================
	-- Purpose: Store user workspace configurations, chart layouts,
	--          technical indicators, drawing tools, and print preferences
	-- Features: Multi-workspace support, chart state persistence,
	--           template system, and user customization
	-- ========================================

	-- 1. WORKSPACES TABLE
	-- Main workspace container for organizing charts and layouts
	CREATE TABLE IF NOT EXISTS workspaces (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

		-- Workspace identification
		name VARCHAR(100) NOT NULL,
		description TEXT,

		-- Default workspace flag (one per user)
		is_default BOOLEAN NOT NULL DEFAULT FALSE,

		-- Workspace settings (JSON configuration)
		-- Contains: theme, layout mode, auto-save, refresh rate, etc.
		settings JSONB NOT NULL DEFAULT '{}',

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Ensure unique workspace names per user
		CONSTRAINT unique_workspace_name_per_user UNIQUE(user_id, name),

		-- Validate settings structure
		CONSTRAINT valid_workspace_settings CHECK (jsonb_typeof(settings) = 'object')
	);

	-- Indexes for workspace queries
	CREATE INDEX idx_workspaces_user_id ON workspaces(user_id);
	CREATE INDEX idx_workspaces_user_default ON workspaces(user_id, is_default) WHERE is_default = TRUE;
	CREATE INDEX idx_workspaces_created_at ON workspaces(created_at DESC);
	CREATE INDEX idx_workspaces_updated_at ON workspaces(updated_at DESC);

	-- Partial index for fast default lookup
	CREATE UNIQUE INDEX idx_workspaces_user_single_default
		ON workspaces(user_id)
		WHERE is_default = TRUE;

	-- Comments for documentation
	COMMENT ON TABLE workspaces IS 'User workspace containers for organizing multiple chart layouts';
	COMMENT ON COLUMN workspaces.is_default IS 'Only one workspace can be default per user (enforced by unique index)';
	COMMENT ON COLUMN workspaces.settings IS 'JSON configuration: {theme, layoutMode, autoSave, refreshRate, timezone, etc.}';

	-- 2. WORKSPACE_CHARTS TABLE
	-- Individual chart configurations within workspaces
	CREATE TABLE IF NOT EXISTS workspace_charts (
		id BIGSERIAL PRIMARY KEY,
		workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

		-- Chart position in grid/tab layout
		chart_index INT NOT NULL,

		-- Market data
		symbol VARCHAR(20) NOT NULL,
		timeframe VARCHAR(10) NOT NULL, -- 1m, 5m, 15m, 1h, 4h, 1d, etc.

		-- Chart visualization
		chart_type VARCHAR(20) NOT NULL DEFAULT 'candlestick', -- candlestick, line, bar, heikin-ashi, etc.

		-- Technical indicators (array of configurations)
		-- Each indicator: {type, period, color, settings, visible}
		indicators JSONB NOT NULL DEFAULT '[]',

		-- Drawing tools (array of drawings)
		-- Each drawing: {type, points, color, style, text}
		drawings JSONB NOT NULL DEFAULT '[]',

		-- Chart-specific settings
		-- Contains: colors, grid, crosshair, volume, scale settings, etc.
		settings JSONB NOT NULL DEFAULT '{}',

		-- Window position and size (for multi-window layouts)
		-- Contains: {x, y, width, height, zIndex, minimized, maximized}
		window_position JSONB,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Ensure unique chart index per workspace
		CONSTRAINT unique_chart_index_per_workspace UNIQUE(workspace_id, chart_index),

		-- Validate JSON structures
		CONSTRAINT valid_indicators CHECK (jsonb_typeof(indicators) = 'array'),
		CONSTRAINT valid_drawings CHECK (jsonb_typeof(drawings) = 'array'),
		CONSTRAINT valid_chart_settings CHECK (jsonb_typeof(settings) = 'object'),
		CONSTRAINT valid_window_position CHECK (window_position IS NULL OR jsonb_typeof(window_position) = 'object'),

		-- Validate chart_index is positive
		CONSTRAINT positive_chart_index CHECK (chart_index >= 0),

		-- Validate chart type
		CONSTRAINT valid_chart_type CHECK (chart_type IN (
			'candlestick', 'line', 'bar', 'area', 'heikin-ashi', 'renko', 'kagi', 'point-figure'
		))
	);

	-- Indexes for chart queries
	CREATE INDEX idx_workspace_charts_workspace_id ON workspace_charts(workspace_id);
	CREATE INDEX idx_workspace_charts_symbol ON workspace_charts(symbol);
	CREATE INDEX idx_workspace_charts_timeframe ON workspace_charts(timeframe);
	CREATE INDEX idx_workspace_charts_workspace_index ON workspace_charts(workspace_id, chart_index);
	CREATE INDEX idx_workspace_charts_updated_at ON workspace_charts(updated_at DESC);

	-- GIN indexes for JSON searching
	CREATE INDEX idx_workspace_charts_indicators ON workspace_charts USING GIN (indicators);
	CREATE INDEX idx_workspace_charts_drawings ON workspace_charts USING GIN (drawings);
	CREATE INDEX idx_workspace_charts_settings ON workspace_charts USING GIN (settings);

	-- Comments
	COMMENT ON TABLE workspace_charts IS 'Individual chart configurations with indicators, drawings, and settings';
	COMMENT ON COLUMN workspace_charts.chart_index IS 'Position in workspace grid (0-based index)';
	COMMENT ON COLUMN workspace_charts.indicators IS 'Array of technical indicators: [{type, period, color, settings, visible}, ...]';
	COMMENT ON COLUMN workspace_charts.drawings IS 'Array of drawing tools: [{type, points, color, style, text}, ...]';
	COMMENT ON COLUMN workspace_charts.settings IS 'Chart-specific settings: {colors, grid, crosshair, volume, scale, etc.}';
	COMMENT ON COLUMN workspace_charts.window_position IS 'Window geometry for multi-window layouts';

	-- 3. USER_PRINT_PREFERENCES TABLE
	-- User-specific print/export preferences
	CREATE TABLE IF NOT EXISTS user_print_preferences (
		user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

		-- Page setup
		page_size VARCHAR(20) NOT NULL DEFAULT 'A4', -- A4, Letter, Legal, Tabloid
		orientation VARCHAR(20) NOT NULL DEFAULT 'landscape', -- landscape, portrait

		-- Margins (in millimeters)
		margins JSONB NOT NULL DEFAULT '{"top": 10, "right": 10, "bottom": 10, "left": 10}',

		-- Header/Footer
		include_header BOOLEAN NOT NULL DEFAULT TRUE,
		include_footer BOOLEAN NOT NULL DEFAULT TRUE,
		header_template TEXT,
		footer_template TEXT,

		-- Print options
		color_mode VARCHAR(20) NOT NULL DEFAULT 'color', -- color, grayscale, black-white
		include_grid BOOLEAN NOT NULL DEFAULT TRUE,
		include_crosshair BOOLEAN NOT NULL DEFAULT FALSE,
		include_volume BOOLEAN NOT NULL DEFAULT TRUE,
		include_indicators BOOLEAN NOT NULL DEFAULT TRUE,
		include_drawings BOOLEAN NOT NULL DEFAULT TRUE,

		-- Chart elements
		show_legend BOOLEAN NOT NULL DEFAULT TRUE,
		show_title BOOLEAN NOT NULL DEFAULT TRUE,
		show_timeframe BOOLEAN NOT NULL DEFAULT TRUE,
		show_symbol BOOLEAN NOT NULL DEFAULT TRUE,

		-- Export settings
		image_quality INT NOT NULL DEFAULT 95, -- JPEG quality 1-100
		image_dpi INT NOT NULL DEFAULT 300, -- Print DPI

		-- PDF settings
		pdf_compression BOOLEAN NOT NULL DEFAULT TRUE,

		-- Watermark
		watermark_enabled BOOLEAN NOT NULL DEFAULT FALSE,
		watermark_text TEXT,
		watermark_opacity DECIMAL(3, 2) DEFAULT 0.3,

		-- Audit trail
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Validate constraints
		CONSTRAINT valid_page_size CHECK (page_size IN ('A4', 'A3', 'Letter', 'Legal', 'Tabloid')),
		CONSTRAINT valid_orientation CHECK (orientation IN ('landscape', 'portrait')),
		CONSTRAINT valid_color_mode CHECK (color_mode IN ('color', 'grayscale', 'black-white')),
		CONSTRAINT valid_margins CHECK (jsonb_typeof(margins) = 'object'),
		CONSTRAINT valid_image_quality CHECK (image_quality BETWEEN 1 AND 100),
		CONSTRAINT valid_image_dpi CHECK (image_dpi BETWEEN 72 AND 600),
		CONSTRAINT valid_watermark_opacity CHECK (watermark_opacity IS NULL OR (watermark_opacity BETWEEN 0 AND 1))
	);

	-- Index
	CREATE INDEX idx_user_print_preferences_updated_at ON user_print_preferences(updated_at DESC);

	-- Comments
	COMMENT ON TABLE user_print_preferences IS 'User-specific print and export preferences for charts';
	COMMENT ON COLUMN user_print_preferences.margins IS 'Page margins in millimeters: {top, right, bottom, left}';
	COMMENT ON COLUMN user_print_preferences.image_quality IS 'JPEG export quality (1-100, higher is better)';
	COMMENT ON COLUMN user_print_preferences.image_dpi IS 'Print resolution in DPI (dots per inch)';

	-- 4. WORKSPACE_TEMPLATES TABLE
	-- Reusable workspace and chart templates
	CREATE TABLE IF NOT EXISTS workspace_templates (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

		-- Template identification
		name VARCHAR(100) NOT NULL,
		description TEXT,

		-- Template type
		template_type VARCHAR(50) NOT NULL,

		-- Template configuration (complete workspace or chart config)
		configuration JSONB NOT NULL,

		-- Sharing settings
		is_public BOOLEAN NOT NULL DEFAULT FALSE,

		-- Usage statistics
		usage_count BIGINT NOT NULL DEFAULT 0,
		last_used_at TIMESTAMPTZ,

		-- Tags for categorization
		tags TEXT[] DEFAULT '{}',

		-- Preview image (base64 or URL)
		preview_image TEXT,

		-- Version control
		version INT NOT NULL DEFAULT 1,
		parent_template_id BIGINT REFERENCES workspace_templates(id) ON DELETE SET NULL,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Ensure unique template names per user and type
		CONSTRAINT unique_template_name_per_user UNIQUE(user_id, name, template_type),

		-- Validate template type
		CONSTRAINT valid_template_type CHECK (template_type IN (
			'workspace', 'chart', 'indicator_set', 'drawing_set', 'color_scheme', 'layout'
		)),

		-- Validate configuration
		CONSTRAINT valid_configuration CHECK (jsonb_typeof(configuration) = 'object'),

		-- Validate version
		CONSTRAINT positive_version CHECK (version > 0)
	);

	-- Indexes for template queries
	CREATE INDEX idx_workspace_templates_user_id ON workspace_templates(user_id);
	CREATE INDEX idx_workspace_templates_type ON workspace_templates(template_type);
	CREATE INDEX idx_workspace_templates_public ON workspace_templates(is_public) WHERE is_public = TRUE;
	CREATE INDEX idx_workspace_templates_usage ON workspace_templates(usage_count DESC);
	CREATE INDEX idx_workspace_templates_created_at ON workspace_templates(created_at DESC);
	CREATE INDEX idx_workspace_templates_tags ON workspace_templates USING GIN (tags);
	CREATE INDEX idx_workspace_templates_parent ON workspace_templates(parent_template_id) WHERE parent_template_id IS NOT NULL;

	-- GIN index for configuration searching
	CREATE INDEX idx_workspace_templates_configuration ON workspace_templates USING GIN (configuration);

	-- Comments
	COMMENT ON TABLE workspace_templates IS 'Reusable templates for workspaces, charts, indicators, and layouts';
	COMMENT ON COLUMN workspace_templates.template_type IS 'Type: workspace, chart, indicator_set, drawing_set, color_scheme, layout';
	COMMENT ON COLUMN workspace_templates.configuration IS 'Complete template configuration (structure depends on template_type)';
	COMMENT ON COLUMN workspace_templates.is_public IS 'Public templates are visible to all users';
	COMMENT ON COLUMN workspace_templates.parent_template_id IS 'Reference to parent template for versioning';

	-- 5. WORKSPACE_SNAPSHOTS TABLE
	-- Automatic snapshots for workspace recovery
	CREATE TABLE IF NOT EXISTS workspace_snapshots (
		id BIGSERIAL PRIMARY KEY,
		workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,

		-- Snapshot metadata
		snapshot_name VARCHAR(100),
		snapshot_type VARCHAR(50) NOT NULL DEFAULT 'auto', -- auto, manual

		-- Complete workspace state at snapshot time
		workspace_data JSONB NOT NULL,
		charts_data JSONB NOT NULL,

		-- Snapshot metadata
		created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
		notes TEXT,

		-- Retention
		expires_at TIMESTAMPTZ,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

		-- Validate snapshot type
		CONSTRAINT valid_snapshot_type CHECK (snapshot_type IN ('auto', 'manual', 'system')),

		-- Validate JSON structures
		CONSTRAINT valid_workspace_data CHECK (jsonb_typeof(workspace_data) = 'object'),
		CONSTRAINT valid_charts_data CHECK (jsonb_typeof(charts_data) = 'array')
	);

	-- Indexes for snapshot queries
	CREATE INDEX idx_workspace_snapshots_workspace_id ON workspace_snapshots(workspace_id);
	CREATE INDEX idx_workspace_snapshots_created_at ON workspace_snapshots(created_at DESC);
	CREATE INDEX idx_workspace_snapshots_expires_at ON workspace_snapshots(expires_at) WHERE expires_at IS NOT NULL;
	CREATE INDEX idx_workspace_snapshots_type ON workspace_snapshots(snapshot_type, created_at DESC);

	-- Comments
	COMMENT ON TABLE workspace_snapshots IS 'Automatic and manual workspace snapshots for recovery and versioning';
	COMMENT ON COLUMN workspace_snapshots.workspace_data IS 'Complete workspace configuration at snapshot time';
	COMMENT ON COLUMN workspace_snapshots.charts_data IS 'Array of all chart configurations at snapshot time';

	-- 6. WORKSPACE_SHARING TABLE
	-- Share workspaces between users
	CREATE TABLE IF NOT EXISTS workspace_sharing (
		id BIGSERIAL PRIMARY KEY,
		workspace_id BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
		shared_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		shared_with BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

		-- Permissions
		can_view BOOLEAN NOT NULL DEFAULT TRUE,
		can_edit BOOLEAN NOT NULL DEFAULT FALSE,
		can_delete BOOLEAN NOT NULL DEFAULT FALSE,

		-- Sharing metadata
		message TEXT,

		-- Audit trail
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMPTZ,

		-- Prevent duplicate shares
		CONSTRAINT unique_workspace_share UNIQUE(workspace_id, shared_with),

		-- Prevent self-sharing
		CONSTRAINT no_self_share CHECK (shared_by != shared_with)
	);

	-- Indexes for sharing queries
	CREATE INDEX idx_workspace_sharing_workspace_id ON workspace_sharing(workspace_id);
	CREATE INDEX idx_workspace_sharing_shared_with ON workspace_sharing(shared_with);
	CREATE INDEX idx_workspace_sharing_shared_by ON workspace_sharing(shared_by);
	CREATE INDEX idx_workspace_sharing_expires_at ON workspace_sharing(expires_at) WHERE expires_at IS NOT NULL;

	-- Comments
	COMMENT ON TABLE workspace_sharing IS 'Workspace sharing permissions between users';

	-- ========================================
	-- TRIGGER FUNCTIONS FOR AUTO-UPDATES
	-- ========================================

	-- Function: Update timestamp on row modification
	CREATE OR REPLACE FUNCTION update_workspace_timestamp()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Function: Ensure only one default workspace per user
	CREATE OR REPLACE FUNCTION enforce_single_default_workspace()
	RETURNS TRIGGER AS $$
	BEGIN
		IF NEW.is_default = TRUE THEN
			-- Unset other default workspaces for this user
			UPDATE workspaces
			SET is_default = FALSE
			WHERE user_id = NEW.user_id
				AND id != NEW.id
				AND is_default = TRUE;
		END IF;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Function: Auto-increment usage count for templates
	CREATE OR REPLACE FUNCTION increment_template_usage()
	RETURNS TRIGGER AS $$
	BEGIN
		UPDATE workspace_templates
		SET usage_count = usage_count + 1,
			last_used_at = CURRENT_TIMESTAMP
		WHERE id = NEW.id;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Function: Clean expired snapshots
	CREATE OR REPLACE FUNCTION clean_expired_snapshots()
	RETURNS void AS $$
	BEGIN
		DELETE FROM workspace_snapshots
		WHERE expires_at IS NOT NULL
			AND expires_at < CURRENT_TIMESTAMP;
	END;
	$$ LANGUAGE plpgsql;

	-- Function: Clean expired shares
	CREATE OR REPLACE FUNCTION clean_expired_shares()
	RETURNS void AS $$
	BEGIN
		DELETE FROM workspace_sharing
		WHERE expires_at IS NOT NULL
			AND expires_at < CURRENT_TIMESTAMP;
	END;
	$$ LANGUAGE plpgsql;

	-- ========================================
	-- TRIGGERS
	-- ========================================

	-- Auto-update timestamps
	CREATE TRIGGER update_workspaces_timestamp
		BEFORE UPDATE ON workspaces
		FOR EACH ROW
		EXECUTE FUNCTION update_workspace_timestamp();

	CREATE TRIGGER update_workspace_charts_timestamp
		BEFORE UPDATE ON workspace_charts
		FOR EACH ROW
		EXECUTE FUNCTION update_workspace_timestamp();

	CREATE TRIGGER update_workspace_templates_timestamp
		BEFORE UPDATE ON workspace_templates
		FOR EACH ROW
		EXECUTE FUNCTION update_workspace_timestamp();

	CREATE TRIGGER update_user_print_preferences_timestamp
		BEFORE UPDATE ON user_print_preferences
		FOR EACH ROW
		EXECUTE FUNCTION update_workspace_timestamp();

	-- Enforce single default workspace
	CREATE TRIGGER enforce_default_workspace
		BEFORE INSERT OR UPDATE ON workspaces
		FOR EACH ROW
		WHEN (NEW.is_default = TRUE)
		EXECUTE FUNCTION enforce_single_default_workspace();

	-- ========================================
	-- MATERIALIZED VIEWS FOR PERFORMANCE
	-- ========================================

	-- Materialized view: Workspace summary statistics
	CREATE MATERIALIZED VIEW IF NOT EXISTS workspace_statistics AS
	SELECT
		w.user_id,
		COUNT(DISTINCT w.id) AS total_workspaces,
		COUNT(DISTINCT wc.id) AS total_charts,
		COUNT(DISTINCT wc.symbol) AS unique_symbols,
		MAX(w.updated_at) AS last_modified,
		COUNT(DISTINCT ws.id) AS total_snapshots,
		COUNT(DISTINCT wt.id) AS total_templates
	FROM workspaces w
	LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
	LEFT JOIN workspace_snapshots ws ON w.id = ws.workspace_id
	LEFT JOIN workspace_templates wt ON w.user_id = wt.user_id
	GROUP BY w.user_id;

	-- Index on materialized view
	CREATE UNIQUE INDEX idx_workspace_statistics_user_id ON workspace_statistics(user_id);

	-- Comments
	COMMENT ON MATERIALIZED VIEW workspace_statistics IS 'Aggregated workspace statistics per user for dashboard display';

	-- ========================================
	-- DATA RETENTION POLICIES
	-- ========================================

	-- Note: Implement via cron jobs or pg_cron
	-- Example retention policies:
	--
	-- Auto snapshots: Keep last 30 days
	-- DELETE FROM workspace_snapshots
	-- WHERE snapshot_type = 'auto'
	--   AND created_at < NOW() - INTERVAL '30 days';
	--
	-- Expired shares: Clean up immediately
	-- SELECT clean_expired_shares();
	--
	-- Unused templates: Archive after 1 year of no usage
	-- UPDATE workspace_templates
	-- SET is_public = FALSE
	-- WHERE last_used_at < NOW() - INTERVAL '1 year';

	-- ========================================
	-- INITIAL DATA (Optional)
	-- ========================================

	-- Default print preferences can be inserted on user creation
	-- or via application logic when first accessed

	`

	_, err := tx.Exec(schema)
	return err
}

func workspaceTablesDown(tx *sql.Tx) error {
	dropSchema := `
	-- Drop materialized views
	DROP MATERIALIZED VIEW IF EXISTS workspace_statistics;

	-- Drop triggers (must drop before functions)
	DROP TRIGGER IF EXISTS enforce_default_workspace ON workspaces;
	DROP TRIGGER IF EXISTS update_user_print_preferences_timestamp ON user_print_preferences;
	DROP TRIGGER IF EXISTS update_workspace_templates_timestamp ON workspace_templates;
	DROP TRIGGER IF EXISTS update_workspace_charts_timestamp ON workspace_charts;
	DROP TRIGGER IF EXISTS update_workspaces_timestamp ON workspaces;

	-- Drop trigger functions
	DROP FUNCTION IF EXISTS clean_expired_shares();
	DROP FUNCTION IF EXISTS clean_expired_snapshots();
	DROP FUNCTION IF EXISTS increment_template_usage();
	DROP FUNCTION IF EXISTS enforce_single_default_workspace();
	DROP FUNCTION IF EXISTS update_workspace_timestamp();

	-- Drop tables (reverse order to handle foreign key dependencies)
	DROP TABLE IF EXISTS workspace_sharing;
	DROP TABLE IF EXISTS workspace_snapshots;
	DROP TABLE IF EXISTS workspace_templates;
	DROP TABLE IF EXISTS user_print_preferences;
	DROP TABLE IF EXISTS workspace_charts;
	DROP TABLE IF EXISTS workspaces;
	`

	_, err := tx.Exec(dropSchema)
	return err
}
