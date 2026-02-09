# Workspace Schema Quick Reference

## Table Summary

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `workspaces` | Main workspace container | Multi-workspace, single default per user |
| `workspace_charts` | Individual chart configs | Indicators, drawings, position-based |
| `user_print_preferences` | Print/export settings | Page setup, margins, watermark |
| `workspace_templates` | Reusable templates | Public/private, usage tracking, versioning |
| `workspace_snapshots` | Workspace backups | Auto/manual, expiration-based retention |
| `workspace_sharing` | Share workspaces | Granular permissions, expiration support |

## Common Queries

### Get User's Default Workspace with Charts
```sql
SELECT
    w.*,
    json_agg(
        json_build_object(
            'id', wc.id,
            'symbol', wc.symbol,
            'timeframe', wc.timeframe,
            'chartType', wc.chart_type,
            'indicators', wc.indicators,
            'drawings', wc.drawings,
            'settings', wc.settings
        ) ORDER BY wc.chart_index
    ) AS charts
FROM workspaces w
LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
WHERE w.user_id = $1 AND w.is_default = TRUE
GROUP BY w.id;
```

### Create Workspace
```sql
INSERT INTO workspaces (user_id, name, description, is_default, settings)
VALUES ($1, $2, $3, $4, $5::jsonb)
RETURNING id, created_at;
```

### Add Chart to Workspace
```sql
INSERT INTO workspace_charts (
    workspace_id, chart_index, symbol, timeframe, chart_type,
    indicators, drawings, settings
)
VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7::jsonb, $8::jsonb)
RETURNING id;
```

### Update Chart Indicators
```sql
UPDATE workspace_charts
SET indicators = $1::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE workspace_id = $2 AND chart_index = $3;
```

### Save Print Preferences
```sql
INSERT INTO user_print_preferences (
    user_id, page_size, orientation, margins, color_mode,
    image_quality, image_dpi
)
VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7)
ON CONFLICT (user_id) DO UPDATE SET
    page_size = EXCLUDED.page_size,
    orientation = EXCLUDED.orientation,
    margins = EXCLUDED.margins,
    color_mode = EXCLUDED.color_mode,
    image_quality = EXCLUDED.image_quality,
    image_dpi = EXCLUDED.image_dpi,
    updated_at = CURRENT_TIMESTAMP;
```

### Create Workspace Snapshot
```sql
INSERT INTO workspace_snapshots (
    workspace_id, snapshot_name, snapshot_type,
    workspace_data, charts_data, created_by
)
SELECT
    w.id,
    $2, -- snapshot_name
    'manual', -- snapshot_type
    row_to_json(w.*),
    json_agg(wc.*),
    $3 -- user_id
FROM workspaces w
LEFT JOIN workspace_charts wc ON w.id = wc.workspace_id
WHERE w.id = $1
GROUP BY w.id
RETURNING id, created_at;
```

### Share Workspace
```sql
INSERT INTO workspace_sharing (
    workspace_id, shared_by, shared_with,
    can_view, can_edit, can_delete, message
)
VALUES ($1, $2, $3, TRUE, $4, FALSE, $5)
ON CONFLICT (workspace_id, shared_with) DO UPDATE SET
    can_view = TRUE,
    can_edit = EXCLUDED.can_edit,
    message = EXCLUDED.message;
```

### Get Shared Workspaces
```sql
SELECT
    w.*,
    ws.can_view,
    ws.can_edit,
    ws.can_delete,
    ws.shared_by,
    ws.message,
    u.username AS shared_by_username
FROM workspaces w
JOIN workspace_sharing ws ON w.id = ws.workspace_id
JOIN users u ON ws.shared_by = u.id
WHERE ws.shared_with = $1
    AND (ws.expires_at IS NULL OR ws.expires_at > CURRENT_TIMESTAMP);
```

### Search Public Templates
```sql
SELECT *
FROM workspace_templates
WHERE is_public = TRUE
    AND template_type = $1
    AND tags && $2::text[]
ORDER BY usage_count DESC, created_at DESC
LIMIT 20;
```

### Get Workspace Statistics
```sql
SELECT * FROM workspace_statistics WHERE user_id = $1;

-- Or refresh first
REFRESH MATERIALIZED VIEW CONCURRENTLY workspace_statistics;
SELECT * FROM workspace_statistics WHERE user_id = $1;
```

## JSON Structure Examples

### Workspace Settings
```json
{
  "theme": "dark",
  "layoutMode": "grid",
  "autoSave": true,
  "refreshRate": 1000,
  "timezone": "America/New_York",
  "gridRows": 2,
  "gridColumns": 2
}
```

### Chart Indicators
```json
[
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
]
```

### Chart Drawings
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
  },
  {
    "type": "rectangle",
    "points": [
      {"time": "2024-01-05T00:00:00Z", "price": 1.0820},
      {"time": "2024-01-10T00:00:00Z", "price": 1.0840}
    ],
    "color": "#FFFF00",
    "fillColor": "#FFFF0033",
    "style": "dashed",
    "text": "Supply Zone"
  }
]
```

### Chart Settings
```json
{
  "showGrid": true,
  "showVolume": true,
  "showCrosshair": true,
  "backgroundColor": "#1E1E1E",
  "gridColor": "#2C2C2C",
  "textColor": "#FFFFFF",
  "upColor": "#26A69A",
  "downColor": "#EF5350",
  "scaleMode": "logarithmic",
  "timeScale": {
    "rightOffset": 12,
    "barSpacing": 6,
    "minBarSpacing": 2
  }
}
```

### Print Margins
```json
{
  "top": 10,
  "right": 10,
  "bottom": 10,
  "left": 10
}
```

### Template Configuration (Workspace)
```json
{
  "name": "Forex Scalping Setup",
  "description": "Multi-timeframe analysis for EUR/USD scalping",
  "settings": {
    "theme": "dark",
    "layoutMode": "grid",
    "gridRows": 2,
    "gridColumns": 2
  },
  "charts": [
    {
      "symbol": "EURUSD",
      "timeframe": "1m",
      "chartType": "candlestick",
      "indicators": [...]
    }
  ]
}
```

## Maintenance Commands

### Refresh Statistics
```sql
REFRESH MATERIALIZED VIEW CONCURRENTLY workspace_statistics;
```

### Clean Expired Data
```sql
-- Clean expired snapshots
SELECT clean_expired_snapshots();

-- Clean expired shares
SELECT clean_expired_shares();

-- Manual cleanup
DELETE FROM workspace_snapshots
WHERE snapshot_type = 'auto'
    AND created_at < NOW() - INTERVAL '30 days';
```

### Analyze Table Statistics
```sql
ANALYZE workspaces;
ANALYZE workspace_charts;
ANALYZE workspace_templates;
```

### Check Index Usage
```sql
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
    AND tablename LIKE 'workspace%'
ORDER BY idx_scan DESC;
```

### Table Sizes
```sql
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) -
                   pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
FROM pg_tables
WHERE schemaname = 'public'
    AND tablename LIKE 'workspace%'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Performance Tips

1. **Use Indexes:** Queries on user_id, workspace_id always use indexes
2. **JSON Queries:** Use `@>` operator for containment checks
3. **Partial Indexes:** Leverage WHERE clauses that match partial indexes
4. **Batch Operations:** Use transactions for multiple chart updates
5. **Materialized Views:** Refresh workspace_statistics periodically, not on every request
6. **Connection Pooling:** Use prepared statements for repeated queries

## Security Best Practices

1. **Row-Level Security:** Enable RLS for multi-tenant security
2. **Validate User IDs:** Always verify user_id matches authenticated user
3. **Sanitize JSON:** Validate JSON structure before insert
4. **Share Expiration:** Always set expires_at for temporary shares
5. **Audit Logging:** Track workspace modifications via updated_at

## Backup Strategy

1. **pg_dump:** Regular full database backups
2. **WAL Archiving:** Point-in-time recovery capability
3. **Snapshot Feature:** Users can create manual snapshots
4. **Export API:** Allow users to export workspaces as JSON

## Migration Commands

```bash
# Apply migration
go run ./cmd/migrate up

# Check status
go run ./cmd/migrate status

# Rollback
go run ./cmd/migrate down
```

## API Endpoint Mapping

| Endpoint | SQL Operation | Table(s) |
|----------|--------------|----------|
| `GET /api/workspaces` | SELECT all workspaces | workspaces |
| `GET /api/workspaces/:id` | SELECT with JOIN | workspaces, workspace_charts |
| `POST /api/workspaces` | INSERT | workspaces |
| `PUT /api/workspaces/:id` | UPDATE | workspaces |
| `DELETE /api/workspaces/:id` | DELETE (CASCADE) | workspaces, workspace_charts |
| `POST /api/workspaces/:id/charts` | INSERT | workspace_charts |
| `PUT /api/charts/:id` | UPDATE | workspace_charts |
| `GET /api/print-preferences` | SELECT | user_print_preferences |
| `PUT /api/print-preferences` | INSERT ON CONFLICT UPDATE | user_print_preferences |
| `GET /api/templates` | SELECT with filters | workspace_templates |
| `POST /api/workspaces/:id/snapshot` | INSERT with subquery | workspace_snapshots |
| `POST /api/workspaces/:id/share` | INSERT | workspace_sharing |

## Troubleshooting

### Issue: Multiple default workspaces
**Solution:**
```sql
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

### Issue: Orphaned charts
**Solution:**
```sql
DELETE FROM workspace_charts
WHERE workspace_id NOT IN (SELECT id FROM workspaces);
```

### Issue: Slow workspace loading
**Solution:**
- Use the pre-built query with json_agg
- Consider caching frequently accessed workspaces
- Check EXPLAIN ANALYZE for query plan

### Issue: Large snapshot storage
**Solution:**
- Implement retention policy
- Compress old snapshots
- Archive to external storage

## References

- Full Documentation: `README_WORKSPACE_SCHEMA.md`
- Migration File: `009_create_workspace_tables.go`
- PostgreSQL JSON: https://www.postgresql.org/docs/current/functions-json.html
