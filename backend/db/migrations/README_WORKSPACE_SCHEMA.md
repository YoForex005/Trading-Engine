# Workspace Persistence Database Schema

## Overview

Migration 009 provides comprehensive workspace persistence for the trading platform, enabling users to save chart configurations, layouts, technical indicators, drawing tools, and print preferences.

## Database Schema Design

### 1. Core Tables

#### `workspaces`
Main container for user workspace configurations.

**Key Features:**
- Multi-workspace support per user
- Single default workspace enforcement (via unique partial index)
- JSON-based flexible settings storage
- Automatic timestamp management

**Columns:**
- `id` - Primary key
- `user_id` - Foreign key to users table (CASCADE delete)
- `name` - Workspace name (unique per user)
- `description` - Optional description
- `is_default` - Default workspace flag (only one per user)
- `settings` - JSONB configuration (theme, layout mode, auto-save, refresh rate, timezone)
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

**Constraints:**
- `unique_workspace_name_per_user` - Ensures unique names per user
- `valid_workspace_settings` - Validates JSONB object structure
- Unique partial index ensures only one default workspace per user

**Indexes:**
- `idx_workspaces_user_id` - Fast user lookup
- `idx_workspaces_user_default` - Partial index for default workspaces
- `idx_workspaces_created_at` - Temporal queries
- `idx_workspaces_updated_at` - Recently modified workspaces
- `idx_workspaces_user_single_default` - Unique default enforcement

#### `workspace_charts`
Individual chart configurations within workspaces.

**Key Features:**
- Normalized chart storage (one row per chart)
- Position-based ordering via chart_index
- Separate storage for indicators, drawings, and settings
- Multi-window layout support

**Columns:**
- `id` - Primary key
- `workspace_id` - Foreign key to workspaces (CASCADE delete)
- `chart_index` - Position in workspace grid (0-based)
- `symbol` - Trading symbol (e.g., EURUSD, BTCUSD)
- `timeframe` - Chart timeframe (1m, 5m, 15m, 1h, 4h, 1d, etc.)
- `chart_type` - Visualization type (candlestick, line, bar, heikin-ashi, etc.)
- `indicators` - JSONB array of technical indicators
- `drawings` - JSONB array of drawing tools
- `settings` - JSONB chart-specific settings
- `window_position` - JSONB window geometry (x, y, width, height, zIndex)
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

**Constraints:**
- `unique_chart_index_per_workspace` - Unique position per workspace
- `valid_indicators` - Validates array structure
- `valid_drawings` - Validates array structure
- `valid_chart_settings` - Validates object structure
- `valid_window_position` - Validates object structure
- `positive_chart_index` - Ensures non-negative index
- `valid_chart_type` - Enum-style validation

**Indexes:**
- `idx_workspace_charts_workspace_id` - Fast workspace lookup
- `idx_workspace_charts_symbol` - Symbol-based filtering
- `idx_workspace_charts_timeframe` - Timeframe filtering
- `idx_workspace_charts_workspace_index` - Composite index for ordering
- `idx_workspace_charts_updated_at` - Recently modified charts
- GIN indexes on `indicators`, `drawings`, `settings` for JSON search

**Indicator JSON Structure:**
```json
[
  {
    "type": "SMA",
    "period": 20,
    "color": "#FF0000",
    "settings": {
      "source": "close",
      "offset": 0
    },
    "visible": true
  }
]
```

**Drawing JSON Structure:**
```json
[
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
]
```

#### `user_print_preferences`
User-specific print and export preferences.

**Key Features:**
- One preference record per user
- Comprehensive print options
- PDF and image export settings
- Watermark support

**Columns:**
- `user_id` - Primary key, foreign key to users (CASCADE delete)
- `page_size` - Page size (A4, A3, Letter, Legal, Tabloid)
- `orientation` - Page orientation (landscape, portrait)
- `margins` - JSONB margins in millimeters
- `include_header` - Show header
- `include_footer` - Show footer
- `header_template` - Custom header template
- `footer_template` - Custom footer template
- `color_mode` - Print color mode (color, grayscale, black-white)
- `include_grid` - Print grid lines
- `include_crosshair` - Print crosshair
- `include_volume` - Print volume bars
- `include_indicators` - Print indicators
- `include_drawings` - Print drawing tools
- `show_legend` - Show chart legend
- `show_title` - Show chart title
- `show_timeframe` - Show timeframe
- `show_symbol` - Show symbol
- `image_quality` - JPEG quality (1-100)
- `image_dpi` - Print DPI (72-600)
- `pdf_compression` - Enable PDF compression
- `watermark_enabled` - Enable watermark
- `watermark_text` - Watermark text
- `watermark_opacity` - Watermark opacity (0-1)
- `updated_at` - Last update timestamp

**Constraints:**
- Multiple CHECK constraints for valid values
- `valid_margins` - Validates JSONB structure

**Margins JSON Structure:**
```json
{
  "top": 10,
  "right": 10,
  "bottom": 10,
  "left": 10
}
```

#### `workspace_templates`
Reusable workspace and chart templates.

**Key Features:**
- Multiple template types (workspace, chart, indicator_set, etc.)
- Public/private sharing
- Usage tracking
- Template versioning

**Columns:**
- `id` - Primary key
- `user_id` - Foreign key to users (CASCADE delete)
- `name` - Template name (unique per user and type)
- `description` - Template description
- `template_type` - Template type (workspace, chart, indicator_set, etc.)
- `configuration` - JSONB complete template configuration
- `is_public` - Public visibility flag
- `usage_count` - Usage counter
- `last_used_at` - Last usage timestamp
- `tags` - Array of tags for categorization
- `preview_image` - Preview image (base64 or URL)
- `version` - Template version number
- `parent_template_id` - Parent template for versioning
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp

**Constraints:**
- `unique_template_name_per_user` - Unique names per user and type
- `valid_template_type` - Validates template type
- `valid_configuration` - Validates JSONB structure
- `positive_version` - Ensures positive version numbers

**Template Types:**
- `workspace` - Complete workspace configuration
- `chart` - Single chart configuration
- `indicator_set` - Collection of indicators
- `drawing_set` - Collection of drawing templates
- `color_scheme` - Color theme
- `layout` - Window layout configuration

**Indexes:**
- Standard B-tree indexes for filtering
- GIN indexes for tags and configuration search
- Partial index for public templates

#### `workspace_snapshots`
Automatic and manual workspace snapshots for recovery.

**Key Features:**
- Automatic snapshot creation
- Manual snapshot with notes
- Expiration-based retention
- Complete workspace state capture

**Columns:**
- `id` - Primary key
- `workspace_id` - Foreign key to workspaces (CASCADE delete)
- `snapshot_name` - Optional snapshot name
- `snapshot_type` - Type (auto, manual, system)
- `workspace_data` - JSONB complete workspace configuration
- `charts_data` - JSONB array of all chart configurations
- `created_by` - User who created snapshot
- `notes` - Optional notes
- `expires_at` - Expiration timestamp for auto-cleanup
- `created_at` - Creation timestamp

**Constraints:**
- `valid_snapshot_type` - Validates snapshot type
- `valid_workspace_data` - Validates JSONB structure
- `valid_charts_data` - Validates JSONB array structure

**Indexes:**
- `idx_workspace_snapshots_workspace_id` - Fast workspace lookup
- `idx_workspace_snapshots_created_at` - Temporal queries
- `idx_workspace_snapshots_expires_at` - Cleanup queries
- `idx_workspace_snapshots_type` - Type-based filtering

#### `workspace_sharing`
Share workspaces between users.

**Key Features:**
- Granular permissions (view, edit, delete)
- Expiration support
- Self-sharing prevention
- Duplicate prevention

**Columns:**
- `id` - Primary key
- `workspace_id` - Foreign key to workspaces (CASCADE delete)
- `shared_by` - User who shared (CASCADE delete)
- `shared_with` - Target user (CASCADE delete)
- `can_view` - View permission
- `can_edit` - Edit permission
- `can_delete` - Delete permission
- `message` - Optional sharing message
- `created_at` - Creation timestamp
- `expires_at` - Optional expiration timestamp

**Constraints:**
- `unique_workspace_share` - Prevents duplicate shares
- `no_self_share` - Prevents self-sharing

**Indexes:**
- Multiple indexes for efficient permission checks
- Partial index for expiring shares

### 2. Database Functions

#### `update_workspace_timestamp()`
Automatically updates the `updated_at` column on row modification.

**Usage:** Triggered on UPDATE for all tables with `updated_at` column.

#### `enforce_single_default_workspace()`
Ensures only one workspace per user can be marked as default.

**Logic:**
- When a workspace is set as default
- Unsets all other default workspaces for that user
- Triggered on INSERT or UPDATE

#### `increment_template_usage()`
Auto-increments usage count for templates.

**Logic:**
- Increments `usage_count`
- Updates `last_used_at` timestamp

#### `clean_expired_snapshots()`
Removes expired snapshots.

**Usage:** Call periodically via cron job or pg_cron.

#### `clean_expired_shares()`
Removes expired workspace shares.

**Usage:** Call periodically via cron job or pg_cron.

### 3. Triggers

| Trigger Name | Table | Event | Function |
|-------------|-------|-------|----------|
| `update_workspaces_timestamp` | workspaces | UPDATE | `update_workspace_timestamp()` |
| `update_workspace_charts_timestamp` | workspace_charts | UPDATE | `update_workspace_timestamp()` |
| `update_workspace_templates_timestamp` | workspace_templates | UPDATE | `update_workspace_timestamp()` |
| `update_user_print_preferences_timestamp` | user_print_preferences | UPDATE | `update_workspace_timestamp()` |
| `enforce_default_workspace` | workspaces | INSERT, UPDATE | `enforce_single_default_workspace()` |

### 4. Materialized Views

#### `workspace_statistics`
Aggregated workspace statistics per user.

**Columns:**
- `user_id` - User ID
- `total_workspaces` - Total workspace count
- `total_charts` - Total chart count
- `unique_symbols` - Unique symbol count
- `last_modified` - Last modification timestamp
- `total_snapshots` - Total snapshot count
- `total_templates` - Total template count

**Refresh Strategy:**
- Manual refresh via `REFRESH MATERIALIZED VIEW workspace_statistics`
- Can be scheduled via cron job or pg_cron
- Use `CONCURRENTLY` option for non-blocking refresh

**Usage:**
```sql
-- Refresh materialized view
REFRESH MATERIALIZED VIEW CONCURRENTLY workspace_statistics;

-- Query user statistics
SELECT * FROM workspace_statistics WHERE user_id = 123;
```

## Index Strategy

### B-tree Indexes
Used for:
- Primary key lookups
- Foreign key relationships
- Equality and range queries
- Sorting operations

### GIN Indexes
Used for:
- JSONB column searches
- Array column searches (tags)
- Full-text search capabilities

### Partial Indexes
Used for:
- Default workspace lookup (WHERE is_default = TRUE)
- Public templates (WHERE is_public = TRUE)
- Non-null parent templates
- Expired snapshots and shares

## Performance Considerations

### Query Optimization

1. **Workspace Loading:**
```sql
-- Efficient workspace loading with charts
SELECT
    w.*,
    json_agg(wc.*) AS charts
FROM workspaces w
LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
WHERE w.user_id = $1 AND w.is_default = TRUE
GROUP BY w.id;
```

2. **Chart Updates:**
```sql
-- Update specific chart (indexed by workspace_id and chart_index)
UPDATE workspace_charts
SET indicators = $1, updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = $2 AND chart_index = $3;
```

3. **Template Search:**
```sql
-- Search public templates by tag
SELECT *
FROM workspace_templates
WHERE is_public = TRUE
    AND tags && ARRAY['forex', 'technical-analysis']
ORDER BY usage_count DESC
LIMIT 20;
```

### JSON Query Optimization

```sql
-- Search for specific indicator type
SELECT *
FROM workspace_charts
WHERE indicators @> '[{"type": "SMA"}]';

-- Search for drawing type
SELECT *
FROM workspace_charts
WHERE drawings @> '[{"type": "trendline"}]';
```

## Data Retention Policies

### Recommended Retention Periods

| Data Type | Retention Period | Implementation |
|-----------|-----------------|----------------|
| Auto snapshots | 30 days | Periodic DELETE with expires_at check |
| Manual snapshots | User-controlled | Never auto-delete unless expired |
| Unused templates | 1 year (then archive) | Update is_public to FALSE |
| Expired shares | Immediate cleanup | Call `clean_expired_shares()` daily |
| Workspace history | 90 days | Optional feature, periodic cleanup |

### Cleanup SQL

```sql
-- Clean expired snapshots
SELECT clean_expired_snapshots();

-- Clean expired shares
SELECT clean_expired_shares();

-- Archive unused templates
UPDATE workspace_templates
SET is_public = FALSE
WHERE last_used_at < NOW() - INTERVAL '1 year'
    AND is_public = TRUE;

-- Delete old auto snapshots
DELETE FROM workspace_snapshots
WHERE snapshot_type = 'auto'
    AND created_at < NOW() - INTERVAL '30 days';
```

## Migration Usage

### Apply Migration

```bash
# Using migration tool
go run ./cmd/migrate up

# Or directly in SQL
psql -d trading_engine -f 009_create_workspace_tables.sql
```

### Rollback Migration

```bash
# Using migration tool
go run ./cmd/migrate down 009

# Or directly in SQL
-- Execute the workspaceTablesDown function
```

## API Integration Examples

### Create Workspace

```go
workspace := &Workspace{
    UserID:      userID,
    Name:        "My Trading Setup",
    Description: "EUR/USD scalping workspace",
    IsDefault:   true,
    Settings: map[string]interface{}{
        "theme": "dark",
        "layoutMode": "grid",
        "autoSave": true,
        "refreshRate": 1000,
        "timezone": "America/New_York",
    },
}

err := db.CreateWorkspace(workspace)
```

### Add Chart to Workspace

```go
chart := &WorkspaceChart{
    WorkspaceID: workspaceID,
    ChartIndex:  0,
    Symbol:      "EURUSD",
    Timeframe:   "5m",
    ChartType:   "candlestick",
    Indicators: []Indicator{
        {
            Type:     "SMA",
            Period:   20,
            Color:    "#FF0000",
            Settings: map[string]interface{}{"source": "close"},
            Visible:  true,
        },
    },
    Drawings: []Drawing{},
    Settings: map[string]interface{}{
        "showGrid": true,
        "showVolume": true,
    },
}

err := db.CreateChart(chart)
```

### Save Print Preferences

```go
prefs := &PrintPreferences{
    UserID:      userID,
    PageSize:    "A4",
    Orientation: "landscape",
    Margins: map[string]int{
        "top": 10, "right": 10, "bottom": 10, "left": 10,
    },
    ColorMode:         "color",
    ImageQuality:      95,
    ImageDPI:          300,
    WatermarkEnabled:  true,
    WatermarkText:     "Trading Engine",
    WatermarkOpacity:  0.3,
}

err := db.SavePrintPreferences(prefs)
```

## Security Considerations

1. **Row-Level Security (RLS):**
```sql
-- Enable RLS on workspaces
ALTER TABLE workspaces ENABLE ROW LEVEL SECURITY;

-- Users can only see their own workspaces
CREATE POLICY user_workspace_policy ON workspaces
    FOR ALL
    USING (user_id = current_user_id());
```

2. **Shared Workspace Access:**
```sql
-- Policy for shared workspaces
CREATE POLICY shared_workspace_policy ON workspaces
    FOR SELECT
    USING (
        user_id = current_user_id() OR
        id IN (
            SELECT workspace_id FROM workspace_sharing
            WHERE shared_with = current_user_id()
                AND can_view = TRUE
                AND (expires_at IS NULL OR expires_at > NOW())
        )
    );
```

## Monitoring and Maintenance

### Key Metrics to Monitor

1. **Workspace Count per User:**
```sql
SELECT user_id, COUNT(*) as workspace_count
FROM workspaces
GROUP BY user_id
ORDER BY workspace_count DESC;
```

2. **Average Charts per Workspace:**
```sql
SELECT AVG(chart_count) as avg_charts
FROM (
    SELECT workspace_id, COUNT(*) as chart_count
    FROM workspace_charts
    GROUP BY workspace_id
) sub;
```

3. **Storage Usage:**
```sql
SELECT
    pg_size_pretty(pg_total_relation_size('workspaces')) as workspaces_size,
    pg_size_pretty(pg_total_relation_size('workspace_charts')) as charts_size,
    pg_size_pretty(pg_total_relation_size('workspace_snapshots')) as snapshots_size;
```

4. **Snapshot Growth:**
```sql
SELECT
    DATE(created_at) as date,
    snapshot_type,
    COUNT(*) as count
FROM workspace_snapshots
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY date, snapshot_type
ORDER BY date DESC;
```

## Backup Recommendations

1. **Regular PostgreSQL backups** (pg_dump)
2. **Point-in-time recovery** (WAL archiving)
3. **Export workspace configurations** to JSON for user-initiated backups
4. **Replicate materialized views** after restore

## Future Enhancements

1. **Workspace Collaboration:**
   - Real-time multi-user editing
   - Comment system
   - Change tracking

2. **Advanced Template Features:**
   - Template marketplace
   - Rating and review system
   - Template categories

3. **Performance Optimizations:**
   - Partitioning large tables by user_id
   - Caching frequently accessed workspaces
   - Compression for old snapshots

4. **Audit Trail:**
   - Detailed change logging
   - Compliance reporting
   - User activity tracking

## Troubleshooting

### Common Issues

1. **Multiple Default Workspaces:**
   - Unique partial index prevents this
   - If data corruption occurs:
   ```sql
   -- Fix: Keep only the most recent default
   WITH ranked AS (
       SELECT id, user_id,
              ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY updated_at DESC) as rn
       FROM workspaces
       WHERE is_default = TRUE
   )
   UPDATE workspaces
   SET is_default = FALSE
   WHERE id IN (SELECT id FROM ranked WHERE rn > 1);
   ```

2. **Orphaned Charts:**
   - CASCADE delete should prevent this
   - Manual cleanup:
   ```sql
   DELETE FROM workspace_charts
   WHERE workspace_id NOT IN (SELECT id FROM workspaces);
   ```

3. **Large JSON Storage:**
   - Consider splitting very large configurations
   - Archive old snapshots
   - Compress preview images

## References

- PostgreSQL JSON Functions: https://www.postgresql.org/docs/current/functions-json.html
- PostgreSQL GIN Indexes: https://www.postgresql.org/docs/current/gin.html
- Materialized Views: https://www.postgresql.org/docs/current/rules-materializedviews.html
