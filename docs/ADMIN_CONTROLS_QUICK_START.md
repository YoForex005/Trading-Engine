# Admin Data Management - Quick Start Guide

## Access the Admin Dashboard

1. **Login as Admin**
   ```
   URL: http://localhost:3000
   Username: admin
   Password: [your admin password]
   ```

2. **Navigate to Data Management**
   - Look for "Data Management" tab in the admin dashboard
   - Or access directly at `/data-management`

## Quick Operations

### View Storage Overview

**What it shows:**
- Total ticks stored across all symbols
- Total storage size (MB)
- Number of symbols with data
- Date range coverage
- Per-symbol breakdown

**How to access:**
1. Select "Overview" tab
2. View summary cards at top
3. Scroll down for symbol-by-symbol table

### Import Historical Data

**Use case:** Add external tick data to the system

**Steps:**
1. Go to "Management" tab
2. Find "Import Data" section
3. Enter symbol (e.g., `EURUSD`)
4. Click "Choose File" and select JSON file
5. Click "Import Data"
6. Wait for confirmation message

**JSON Format:**
```json
[
  {
    "broker_id": "demo",
    "symbol": "EURUSD",
    "bid": 1.0850,
    "ask": 1.0852,
    "spread": 0.0002,
    "timestamp": "2026-01-20T10:00:00Z",
    "lp": "YOFX"
  }
]
```

### Clean Up Old Data

**Use case:** Free up disk space by deleting old ticks

**Steps:**
1. Go to "Management" tab
2. Find "Cleanup Old Data" section
3. Optionally enter a specific symbol (leave empty for all)
4. Set "Delete data older than" (default: 180 days)
5. **Check the confirmation box**
6. Click "Delete Old Data"

**Warning:** This permanently deletes data. Make a backup first!

### Compress Old Data

**Use case:** Save disk space without deleting data

**Steps:**
1. Go to "Management" tab
2. Find "Compress Old Data" section
3. Optionally enter symbol (empty = all)
4. Set "Compress data older than" (default: 90 days)
5. Click "Compress Data"

**Result:**
- Old JSON files moved to `data/ticks/{SYMBOL}/archive/`
- Compressed as ZIP files
- Typically saves 60-80% disk space

### Create Backup

**Use case:** Full backup of all tick data

**Steps:**
1. Go to "Management" tab
2. Find "Create Backup" section
3. Optionally modify backup path (default: `backups`)
4. Click "Create Backup"
5. Wait for completion (large datasets may take minutes)

**Output:**
- Backup file: `backups/ticks_backup_YYYYMMDD_HHMMSS.zip`
- Contains all symbols in ZIP format
- Includes metadata about size and symbol count

### Monitor Real-time Metrics

**Use case:** Track system health and performance

**Steps:**
1. Go to "Monitoring" tab
2. View real-time metrics:
   - Tick ingestion rate (ticks/second)
   - Storage growth (MB/hour)
   - Failed writes count
   - Data quality score (0-100)

**Auto-refresh:** Updates every 10 seconds

### Configure Retention Policy

**Use case:** Set automatic data lifecycle rules

**Steps:**
1. Go to "Config" tab
2. Set "Retention Policy" (days)
   - Data older than this will be automatically cleaned
3. Enable/disable "Automatic compression"
4. Set "Backup Schedule" (cron format)
   - Example: `0 2 * * *` = Daily at 2 AM
5. Set "Backup Path" (directory)

**Cron Examples:**
- `0 2 * * *` - Daily at 2 AM
- `0 2 * * 0` - Weekly on Sunday at 2 AM
- `0 2 1 * *` - Monthly on 1st at 2 AM

## Common Workflows

### Workflow 1: Weekly Maintenance

```
1. Check storage overview
2. Review data quality score
3. Compress data older than 90 days
4. Delete data older than 180 days (after confirming backup)
5. Create full backup
```

### Workflow 2: Import New Historical Data

```
1. Prepare JSON file with tick data
2. Go to Management > Import Data
3. Select symbol and file
4. Import and verify in Overview tab
5. Check symbol statistics updated
```

### Workflow 3: Free Up Disk Space

```
1. Check current storage size in Overview
2. Compress old data (90+ days)
3. Note space savings
4. If more space needed, cleanup very old data (180+ days)
5. Verify in Overview that size decreased
```

## Troubleshooting

### Import Not Working

**Check:**
- JSON format matches expected structure
- File size under 100MB (split large files)
- Symbol name is valid (e.g., `EURUSD`, not `EUR/USD`)
- You have admin role

**Fix:**
- Validate JSON with online validator
- Check browser console for error details
- Verify admin token not expired

### Cleanup Not Freeing Space

**Possible causes:**
1. Data already compressed (in `archive/` folder)
2. Retention days too high (no files match criteria)
3. Disk cache not updated

**Solutions:**
- Check `data/ticks/{SYMBOL}/archive/` for ZIP files
- Lower retention days threshold
- Restart server to clear cache
- Use OS tools to verify actual disk usage

### Backup Taking Too Long

**Optimization:**
- Select specific symbols instead of all
- Run backup during low-traffic hours
- Consider incremental backups
- Compress old data first (already ZIP format)

### Quality Score Low

**Investigate:**
- Check "Failed Writes" count
- Review ticks with zero spread
- Look for invalid prices (negative, NaN)
- Check LP connection status

**Fix:**
- Fix LP connectivity issues
- Review tick validation rules
- Re-import corrected data

## Security Notes

**Admin-Only Access:**
- All operations require admin role
- JWT token validated on every request
- Unauthorized users see 403 Forbidden

**Audit Logging:**
- All admin actions logged to `data/audit.log`
- Includes user, timestamp, action, success/failure
- Review for security compliance

**Audit Log Location:**
```
data/audit.log
```

**View Recent Audit Entries:**
```bash
tail -f data/audit.log | jq .
```

## API Endpoints Reference

For programmatic access:

```bash
# Get storage stats
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/admin/history/stats

# Import data
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -F "symbol=EURUSD" \
  -F "file=@ticks.json" \
  http://localhost:8080/admin/history/import

# Cleanup old data
curl -X DELETE \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"olderThanDays":180,"confirm":true}' \
  http://localhost:8080/admin/history/cleanup

# Compress old data
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"olderThanDays":90}' \
  http://localhost:8080/admin/history/compress

# Create backup
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"backupPath":"backups"}' \
  http://localhost:8080/admin/history/backup

# Get monitoring metrics
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/admin/history/monitoring
```

## Best Practices

1. **Regular Backups**
   - Schedule weekly backups
   - Store off-system (S3, network drive)
   - Test restore process quarterly

2. **Gradual Cleanup**
   - Start with 180+ day retention
   - Compress before deleting
   - Verify backups exist first

3. **Monitor Quality**
   - Keep quality score above 90%
   - Investigate failed writes immediately
   - Review zero-spread ticks

4. **Audit Review**
   - Check audit log weekly
   - Verify expected admin actions
   - Investigate unexpected changes

5. **Capacity Planning**
   - Monitor storage growth rate
   - Plan cleanup before 80% full
   - Consider archive storage for old data

## Support

**Documentation:**
- Full docs: `docs/ADMIN_DATA_MANAGEMENT.md`
- Architecture overview in docs

**Logs:**
- Server logs: Check console output
- Audit logs: `data/audit.log`
- Application logs: Browser console

**Contact:**
- Internal support channel
- Development team for bugs
- Security team for audit issues

---

**Last Updated**: 2026-01-20
**Version**: 1.0
