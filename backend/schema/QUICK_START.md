# Tick Storage Database - Quick Start Guide

## 5-Minute Setup

### Prerequisites

1. **SQLite Installation**

   **Windows:**
   ```powershell
   # Download SQLite from https://www.sqlite.org/download.html
   # Or use Chocolatey:
   choco install sqlite

   # Verify installation
   sqlite3 --version
   ```

   **Linux:**
   ```bash
   # Ubuntu/Debian
   sudo apt-get update
   sudo apt-get install sqlite3

   # CentOS/RHEL
   sudo yum install sqlite

   # Verify installation
   sqlite3 --version
   ```

2. **Go (for migration tool)**
   ```bash
   # Verify Go installation
   go version
   # If not installed: https://go.dev/doc/install
   ```

### Step 1: Create Your First Database (2 minutes)

**Option A: Command Line**

```bash
# Create directory structure
mkdir -p data/ticks/db/2026/01

# Create database and apply schema
cd data/ticks/db/2026/01
sqlite3 ticks_2026-01-20.db < ../../../../backend/schema/ticks.sql

# Verify schema created
sqlite3 ticks_2026-01-20.db "SELECT name FROM sqlite_master WHERE type='table';"
```

**Option B: PowerShell (Windows)**

```powershell
# Create directory structure
New-Item -ItemType Directory -Path "data\ticks\db\2026\01" -Force

# Create database and apply schema
Get-Content backend\schema\ticks.sql | sqlite3 data\ticks\db\2026\01\ticks_2026-01-20.db

# Verify schema
sqlite3 data\ticks\db\2026\01\ticks_2026-01-20.db "SELECT name FROM sqlite_master WHERE type='table';"
```

**Expected Output:**
```
ticks
symbols
lp_sources
tick_partitions
tick_quality_metrics
migration_log
```

### Step 2: Test Insert and Query (1 minute)

```bash
# Open database
sqlite3 data/ticks/db/2026/01/ticks_2026-01-20.db

# Insert test tick
INSERT INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
VALUES ('EURUSD', 1737388800000, 1.0850, 1.0852, 0.0002, 'YOFX');

# Query back
SELECT
    symbol,
    datetime(timestamp / 1000, 'unixepoch') as time,
    bid,
    ask,
    spread,
    lp_source
FROM ticks
LIMIT 10;

# Exit
.quit
```

**Expected Output:**
```
EURUSD|2026-01-20 12:00:00|1.085|1.0852|0.0002|YOFX
```

### Step 3: Migrate Existing JSON Data (2 minutes)

**Install Go dependencies:**
```bash
cd backend/schema
go mod init tick-migration
go get github.com/mattn/go-sqlite3
```

**Run migration:**
```bash
# Migrate all JSON files to SQLite
go run migrate_to_sqlite.go \
    --input-dir "../../data/ticks" \
    --output-dir "../../data/ticks/db" \
    --batch-size 1000 \
    --verbose

# Expected output:
# [INFO] Processing symbol: EURUSD
# [INFO]   Migrating data/ticks/EURUSD/2026-01-19.json: 84523 ticks
# [INFO]   ✓ Migrated 84523 ticks in 2.3s
# ...
# [INFO] === Migration Complete ===
# [INFO] Duration: 45.2s
# [INFO] Symbols processed: 100
# [INFO] Files processed: 300
# [INFO] Ticks migrated: 25,347,891
# [INFO] ✅ Migration successful!
```

## Daily Operations

### Automated Rotation (Run at Midnight)

**Linux (Cron):**
```bash
# Edit crontab
crontab -e

# Add this line (run at midnight UTC)
0 0 * * * /path/to/backend/schema/rotate_tick_db.sh rotate >> /var/log/tick-rotation.log 2>&1
```

**Windows (Task Scheduler):**
```powershell
# Create scheduled task
$trigger = New-ScheduledTaskTrigger -Daily -At "00:00"
$action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-File C:\Trading-Engine\backend\schema\rotate_tick_db.ps1 -Action rotate"
Register-ScheduledTask -TaskName "TickDBRotation" -Trigger $trigger -Action $action -User "SYSTEM"
```

### Automated Compression (Run Daily)

**Linux (Cron):**
```bash
# Add to crontab (run at 2 AM)
0 2 * * * /path/to/backend/schema/compress_old_dbs.sh compress >> /var/log/tick-compression.log 2>&1
```

**Windows (Task Scheduler):**
```powershell
# Install zstd first: https://github.com/facebook/zstd/releases
# Then create scheduled task similar to rotation
```

## Common Queries

### Get Recent Ticks
```sql
-- Last 100 ticks for EURUSD
SELECT
    datetime(timestamp / 1000, 'unixepoch') as time,
    bid,
    ask,
    spread
FROM ticks
WHERE symbol = 'EURUSD'
ORDER BY timestamp DESC
LIMIT 100;
```

### Get Ticks for Time Range
```sql
-- EURUSD ticks for last hour
SELECT COUNT(*) as tick_count
FROM ticks
WHERE symbol = 'EURUSD'
  AND timestamp > (strftime('%s', 'now') - 3600) * 1000;
```

### Get Statistics
```sql
-- Daily statistics by symbol
SELECT
    symbol,
    COUNT(*) as ticks,
    ROUND(AVG(spread), 6) as avg_spread,
    ROUND(MIN(bid), 5) as min_bid,
    ROUND(MAX(ask), 5) as max_ask
FROM ticks
GROUP BY symbol
ORDER BY ticks DESC;
```

### Find Gaps
```sql
-- Find gaps > 60 seconds
WITH tick_gaps AS (
    SELECT
        symbol,
        timestamp,
        LEAD(timestamp) OVER (PARTITION BY symbol ORDER BY timestamp) as next_timestamp
    FROM ticks
    WHERE symbol = 'EURUSD'
)
SELECT
    datetime(timestamp / 1000, 'unixepoch') as gap_start,
    datetime(next_timestamp / 1000, 'unixepoch') as gap_end,
    ROUND((next_timestamp - timestamp) / 1000.0, 2) as gap_seconds
FROM tick_gaps
WHERE (next_timestamp - timestamp) / 1000.0 > 60
ORDER BY gap_seconds DESC
LIMIT 20;
```

## Integration with Go Code

### Basic Integration Example

```go
package tickstore

import (
    "database/sql"
    "time"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteTickStore struct {
    db *sql.DB
}

// NewSQLiteTickStore creates a new SQLite tick store
func NewSQLiteTickStore(dbPath string) (*SQLiteTickStore, error) {
    // Open with performance optimizations
    dsn := dbPath + "?_journal_mode=WAL&_synchronous=NORMAL&cache=shared&_busy_timeout=5000"
    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, err
    }

    // Test connection
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &SQLiteTickStore{db: db}, nil
}

// StoreTick inserts a tick (use batch for better performance)
func (s *SQLiteTickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) error {
    _, err := s.db.Exec(`
        INSERT OR IGNORE INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
        VALUES (?, ?, ?, ?, ?, ?)
    `, symbol, timestamp.UnixMilli(), bid, ask, spread, lp)
    return err
}

// StoreTicks batch insert (much faster)
func (s *SQLiteTickStore) StoreTicks(ticks []Tick) error {
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT OR IGNORE INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
        VALUES (?, ?, ?, ?, ?, ?)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, tick := range ticks {
        _, err := stmt.Exec(tick.Symbol, tick.Timestamp.UnixMilli(),
            tick.Bid, tick.Ask, tick.Spread, tick.LP)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

// GetRecentTicks retrieves recent ticks for a symbol
func (s *SQLiteTickStore) GetRecentTicks(symbol string, limit int) ([]Tick, error) {
    rows, err := s.db.Query(`
        SELECT timestamp, bid, ask, spread, lp_source
        FROM ticks
        WHERE symbol = ?
        ORDER BY timestamp DESC
        LIMIT ?
    `, symbol, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var ticks []Tick
    for rows.Next() {
        var t Tick
        var ts int64
        err := rows.Scan(&ts, &t.Bid, &t.Ask, &t.Spread, &t.LP)
        if err != nil {
            return nil, err
        }
        t.Timestamp = time.UnixMilli(ts)
        t.Symbol = symbol
        ticks = append(ticks, t)
    }

    return ticks, rows.Err()
}

// Close closes the database connection
func (s *SQLiteTickStore) Close() error {
    return s.db.Close()
}
```

### Usage in Main Application

```go
// Initialize store
store, err := NewSQLiteTickStore("data/ticks/db/2026/01/ticks_2026-01-20.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Store single tick
err = store.StoreTick("EURUSD", 1.0850, 1.0852, 0.0002, "YOFX", time.Now())
if err != nil {
    log.Printf("Failed to store tick: %v", err)
}

// Batch insert (preferred for performance)
var batch []Tick
// ... collect ticks ...
if len(batch) >= 500 {
    if err := store.StoreTicks(batch); err != nil {
        log.Printf("Failed to store batch: %v", err)
    }
    batch = batch[:0] // Clear batch
}

// Query recent ticks
ticks, err := store.GetRecentTicks("EURUSD", 100)
if err != nil {
    log.Printf("Failed to query ticks: %v", err)
}

for _, tick := range ticks {
    fmt.Printf("%s: %.5f/%.5f (spread: %.5f)\n",
        tick.Timestamp.Format("15:04:05"),
        tick.Bid, tick.Ask, tick.Spread)
}
```

## Troubleshooting

### Database Locked Error

**Problem:** `database is locked`

**Solutions:**
```bash
# 1. Check for other processes
lsof data/ticks/db/2026/01/ticks_2026-01-20.db  # Linux
# Or use Process Explorer on Windows

# 2. Ensure WAL mode is enabled
sqlite3 ticks_2026-01-20.db "PRAGMA journal_mode;"
# Should return: wal

# 3. Increase busy timeout in connection string
# DSN: dbPath + "?_busy_timeout=5000"
```

### Slow Queries

**Problem:** Queries taking too long

**Solutions:**
```sql
-- 1. Check if indexes exist
SELECT name FROM sqlite_master WHERE type='index';

-- 2. Analyze tables
ANALYZE;

-- 3. Check query plan
EXPLAIN QUERY PLAN
SELECT * FROM ticks WHERE symbol = 'EURUSD' LIMIT 100;

-- 4. Increase cache size
PRAGMA cache_size = -128000;  -- 128MB
```

### Migration Fails

**Problem:** Migration tool errors

**Solutions:**
```bash
# 1. Check Go installation
go version

# 2. Install SQLite driver
go get github.com/mattn/go-sqlite3

# 3. Run with verbose flag
go run migrate_to_sqlite.go --verbose --dry-run

# 4. Check JSON file format
head -20 data/ticks/EURUSD/2026-01-19.json
```

## Performance Tips

1. **Batch Inserts**: Use transactions with 500-1000 rows per batch
2. **WAL Mode**: Always enable (`PRAGMA journal_mode = WAL`)
3. **Cache Size**: Increase for better performance (`PRAGMA cache_size = -64000`)
4. **Synchronous**: Use NORMAL for balance (`PRAGMA synchronous = NORMAL`)
5. **Memory Mapping**: Enable for large files (`PRAGMA mmap_size = 268435456`)

## Monitoring

### Check Database Size
```bash
# Linux
du -sh data/ticks/db/2026/01/*.db

# Windows
Get-ChildItem data\ticks\db\2026\01\*.db | Select-Object Name, @{N='Size (MB)';E={[math]::Round($_.Length / 1MB, 2)}}
```

### Check Tick Counts
```sql
-- Total ticks
SELECT COUNT(*) FROM ticks;

-- Ticks per symbol
SELECT symbol, COUNT(*) FROM ticks GROUP BY symbol;

-- Ticks per hour (today)
SELECT
    strftime('%H:00', timestamp / 1000, 'unixepoch') as hour,
    COUNT(*) as ticks
FROM ticks
WHERE DATE(timestamp / 1000, 'unixepoch') = DATE('now')
GROUP BY hour
ORDER BY hour;
```

### Health Check Script
```bash
#!/bin/bash
# Quick health check

DB_PATH="data/ticks/db/2026/01/ticks_2026-01-20.db"

echo "Database Health Check"
echo "===================="

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "❌ Database not found: $DB_PATH"
    exit 1
fi
echo "✅ Database exists"

# Check integrity
INTEGRITY=$(sqlite3 "$DB_PATH" "PRAGMA integrity_check;")
if [ "$INTEGRITY" = "ok" ]; then
    echo "✅ Integrity check passed"
else
    echo "❌ Integrity check failed: $INTEGRITY"
    exit 1
fi

# Check tick count
TICK_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks;")
echo "✅ Total ticks: $TICK_COUNT"

# Check recent activity (last hour)
RECENT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks WHERE timestamp > (strftime('%s', 'now') - 3600) * 1000;")
echo "✅ Recent ticks (last hour): $RECENT"

# Check database size
SIZE=$(du -h "$DB_PATH" | cut -f1)
echo "✅ Database size: $SIZE"

echo ""
echo "✅ All checks passed!"
```

## Next Steps

1. **Review Schema**: Read `ticks.sql` for detailed schema documentation
2. **Migration**: Run `migrate_to_sqlite.go` to convert existing JSON data
3. **Automation**: Set up daily rotation and compression scripts
4. **Integration**: Modify Go code to use SQLite instead of JSON
5. **Monitoring**: Set up alerts for disk space and data quality
6. **Testing**: Verify performance meets requirements

## Resources

- **Full Documentation**: `backend/schema/README.md`
- **Decision Rationale**: `backend/schema/DECISION_RATIONALE.md`
- **Maintenance Queries**: `backend/schema/maintenance.sql`
- **Migration Tool**: `backend/schema/migrate_to_sqlite.go`
- **Rotation Scripts**: `backend/schema/rotate_tick_db.sh` (Linux) or `.ps1` (Windows)
- **Compression Script**: `backend/schema/compress_old_dbs.sh`

## Support

For questions or issues:
1. Check this Quick Start guide
2. Review the comprehensive README
3. Run health check script
4. Check logs for errors
5. Open GitHub issue with details
