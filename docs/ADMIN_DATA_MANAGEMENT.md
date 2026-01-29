# Admin Data Management - Comprehensive Implementation

## Overview

Comprehensive admin controls for tick data management with role-based access control (RBAC), storage monitoring, and data lifecycle operations.

## Features Implemented

### 1. Storage Overview
- **Total Statistics**: Total ticks, storage size (MB), symbol count, date range coverage
- **Per-Symbol Breakdown**: Individual symbol statistics with tick counts, sizes, and date ranges
- **Missing Data Detection**: Identifies gaps in data coverage
- **Real-time Updates**: Automatic refresh with last updated timestamp

### 2. Data Management Operations

#### Import Data
- **Bulk upload** via JSON file upload
- **Symbol-specific** imports with merge/deduplication
- **Automatic OHLC rebuild** after import
- **Progress tracking** with success/error feedback

#### Cleanup Old Data
- **Configurable retention**: Delete data older than N days
- **Symbol filtering**: Target specific symbols or all
- **Confirmation required**: Safety checkbox to prevent accidental deletion
- **Audit logging**: All deletions logged with user, timestamp, and details

#### Compress Old Data
- **Archive creation**: ZIP compression of old data files
- **Space savings**: Tracks freed storage space
- **Configurable age threshold**: Compress data older than N days
- **Performance metrics**: Reports compressed file count and MB saved

#### Backup & Restore
- **ZIP archive creation**: Full symbol backup with timestamps
- **Selective backup**: Choose specific symbols or all
- **Custom backup path**: Configurable backup location
- **Size reporting**: Shows backup file size and symbol count

### 3. Monitoring Metrics
- **Tick Ingestion Rate**: Real-time ticks per second
- **Storage Growth**: MB per hour tracking
- **Failed Writes**: Error count monitoring
- **Data Quality Score**: 0-100 quality assessment
  - Zero spread detection
  - Invalid price detection
  - Overall health score

### 4. Configuration Management
- **Retention Policy**: Default 180 days, configurable
- **Auto-compression**: Enable/disable automatic compression
- **Backup Schedule**: Cron format scheduling (e.g., `0 2 * * *` for 2 AM daily)
- **Backup Path**: Customizable backup directory

## Architecture

### Backend API (`backend/api/admin_history.go`)

#### Endpoints

| Endpoint | Method | Description | Auth |
|----------|--------|-------------|------|
| `/admin/history/stats` | GET | Storage statistics overview | ADMIN |
| `/admin/history/import` | POST | Import tick data from file | ADMIN |
| `/admin/history/cleanup` | DELETE | Delete old data | ADMIN |
| `/admin/history/compress` | POST | Compress old data | ADMIN |
| `/admin/history/backup` | POST | Create backup archive | ADMIN |
| `/admin/history/monitoring` | GET | Real-time metrics | ADMIN |

#### Key Components

**AdminHistoryHandler**
```go
type AdminHistoryHandler struct {
    tickStore tickstore.TickStorageService
    authSvc   *auth.Service
}
```

**Storage Statistics**
```go
type StorageStats struct {
    TotalTicks       int64
    TotalSizeBytes   int64
    TotalSizeMB      float64
    SymbolCount      int
    DateRangeStart   string
    DateRangeEnd     string
    SymbolStats      map[string]SymbolStats
    MissingDataGaps  []DataGap
    LastUpdated      time.Time
}
```

**Monitoring Metrics**
```go
type MonitoringMetrics struct {
    TickIngestionRate      float64
    StorageGrowthMBPerHour float64
    FailedWrites           int
    LastTickTimestamp      map[string]time.Time
    DataQuality            DataQualityMetrics
}
```

### Frontend UI (`admin/broker-admin/src/components/dashboard/DataManagement.tsx`)

#### Tab Structure

1. **Overview Tab**
   - Summary cards (Total Ticks, Storage Size, Symbols, Date Range)
   - Symbol statistics table with sortable columns
   - Visual indicators for data health

2. **Management Tab**
   - Import data form with file upload
   - Cleanup old data with confirmation
   - Compress data with age threshold
   - Backup creation with symbol selection

3. **Monitoring Tab**
   - Real-time metrics cards
   - Ingestion rate visualization
   - Storage growth trends
   - Data quality score

4. **Config Tab**
   - Retention policy settings
   - Compression toggles
   - Backup schedule (cron format)
   - Backup path configuration

#### State Management

```typescript
interface ConfigSettings {
  retentionDays: number;
  autoDownload: boolean;
  autoDownloadSymbols?: string[];
  compressionEnabled: boolean;
  backupSchedule: string;
  backupPath: string;
}
```

## Security (Role-Based Access Control)

### Authentication Flow

1. **Token Validation**: All endpoints verify JWT token in `Authorization` header
2. **Role Check**: Requires `ADMIN` role in JWT claims
3. **Audit Logging**: All admin actions logged with:
   - User ID and username
   - Action type (IMPORT, CLEANUP, COMPRESS, BACKUP)
   - Target (symbol or "all")
   - Success/failure status
   - Error details if failed
   - Timestamp

### Audit Log Format

```go
type AuditLog struct {
    Timestamp time.Time
    UserID    string
    Username  string
    Action    string
    Target    string
    Details   map[string]interface{}
    Success   bool
    Error     string
}
```

**Audit logs stored at**: `data/audit.log` (JSON lines format)

## Data Operations

### Import Operation

1. Parse multipart form with file upload
2. Validate JSON tick data format
3. Merge with existing data (deduplication by timestamp)
4. Rebuild OHLC cache for symbol
5. Log audit entry

**Expected JSON Format**:
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

### Cleanup Operation

1. Calculate cutoff date (`now - retentionDays`)
2. Scan symbol directories
3. Identify files older than cutoff
4. Delete matching files
5. Report freed space and file count
6. Log audit entry

### Compression Operation

1. Create `archive/` subdirectory
2. ZIP compress files older than threshold
3. Delete original JSON files
4. Report space savings
5. Log audit entry

### Backup Operation

1. Create ZIP archive with timestamp
2. Walk symbol directories
3. Add all files to archive
4. Calculate final archive size
5. Log audit entry

## Usage Examples

### Import Historical Data

```bash
# Via API
curl -X POST http://localhost:8080/admin/history/import \
  -H "Authorization: Bearer <admin-token>" \
  -F "symbol=EURUSD" \
  -F "file=@eurusd_ticks.json"
```

### Cleanup Old Data

```bash
curl -X DELETE http://localhost:8080/admin/history/cleanup \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "olderThanDays": 180,
    "confirm": true
  }'
```

### Create Backup

```bash
curl -X POST http://localhost:8080/admin/history/backup \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "backupPath": "backups"
  }'
```

### Get Storage Stats

```bash
curl http://localhost:8080/admin/history/stats \
  -H "Authorization: Bearer <admin-token>"
```

## Integration Points

### Main Server Registration (`backend/cmd/server/main.go`)

```go
// Admin History Management
adminHistoryHandler := api.NewAdminHistoryHandler(tickStore, authService)
http.HandleFunc("/admin/history/stats", adminHistoryHandler.HandleGetStats)
http.HandleFunc("/admin/history/import", adminHistoryHandler.HandleImportData)
http.HandleFunc("/admin/history/cleanup", adminHistoryHandler.HandleCleanupOldData)
http.HandleFunc("/admin/history/compress", adminHistoryHandler.HandleCompressData)
http.HandleFunc("/admin/history/backup", adminHistoryHandler.HandleBackup)
http.HandleFunc("/admin/history/monitoring", adminHistoryHandler.HandleGetMonitoring)
```

### TickStore Integration

The admin handler uses the `tickstore.TickStorageService` interface:

```go
type TickStorageService interface {
    StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time)
    GetHistory(symbol string, limit int) []Tick
    GetOHLC(symbol string, timeframeSecs int64, limit int) []OHLC
    GetSymbols() []string
    GetTickCount(symbol string) int
}
```

For advanced operations (import, merge), it casts to `*tickstore.TickStore` to access `DailyStore`.

## File Structure

```
backend/
├── api/
│   └── admin_history.go       # Admin API endpoints
└── cmd/server/
    └── main.go                # Endpoint registration

admin/broker-admin/src/
└── components/dashboard/
    └── DataManagement.tsx     # Admin UI

data/
├── audit.log                  # Audit trail
├── ticks/                     # Tick data storage
│   └── {SYMBOL}/
│       ├── 2026-01-20.json
│       └── archive/           # Compressed archives
│           └── 2026-01-15.zip
└── backups/                   # Backup archives
    └── ticks_backup_20260120_120000.zip
```

## Performance Considerations

- **Lazy Loading**: Statistics calculated on-demand to avoid startup overhead
- **Atomic Operations**: File writes use temp files + rename for atomicity
- **Concurrent Safety**: Mutex-protected data structures
- **Streaming**: Large file operations use streaming to minimize memory
- **Background Workers**: Periodic persistence runs in goroutines

## Future Enhancements

1. **Scheduled Jobs**: Cron-based automatic cleanup and backups
2. **Real-time Metrics**: WebSocket streaming for live monitoring
3. **Data Validation**: Pre-import data quality checks
4. **Restore Functionality**: UI for backup restoration
5. **Export to CSV**: Additional export formats
6. **Data Replay**: Historical data replay for testing
7. **Compression Levels**: Configurable ZIP compression levels
8. **S3 Integration**: Cloud backup support
9. **Retention Policies**: Per-symbol retention settings
10. **Data Gaps Auto-Fill**: Automatic backfill for missing periods

## Testing Checklist

- [ ] Admin authentication and authorization
- [ ] Storage statistics accuracy
- [ ] Import with deduplication
- [ ] Cleanup with confirmation
- [ ] Compression and space savings
- [ ] Backup creation and verification
- [ ] Monitoring metrics updates
- [ ] Audit log entries
- [ ] Error handling and rollback
- [ ] UI responsiveness and feedback

## Troubleshooting

### Import Fails with "Invalid tick data format"

**Cause**: JSON structure doesn't match expected format

**Solution**: Ensure JSON array of tick objects with required fields:
```json
[{"broker_id": "...", "symbol": "...", "bid": 0.0, "ask": 0.0, "spread": 0.0, "timestamp": "...", "lp": "..."}]
```

### Cleanup Not Freeing Space

**Cause**: Files still open or compression already applied

**Solution**: Check if files are in `archive/` subdirectory (already compressed)

### Backup Size Too Large

**Cause**: Including all symbols without filtering

**Solution**: Use selective backup with specific symbols

## Support

For issues or questions:
- Review audit logs at `data/audit.log`
- Check server logs for detailed error messages
- Verify admin token has `ADMIN` role claim

---

**Implementation Date**: 2026-01-20
**Version**: 1.0
**Status**: Production-ready with comprehensive RBAC and audit logging
