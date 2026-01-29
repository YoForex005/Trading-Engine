# Tick Storage Database Schema

## Overview

This directory contains database schema designs for persistent tick storage in the Trading Engine. The recommended approach is **SQLite with daily partitions** for optimal cross-platform compatibility, simplicity, and performance.

## Files

- `ticks.sql` - Complete SQLite schema with three implementation options
- `migrate_to_sqlite.go` - Migration tool to convert existing JSON files to SQLite
- `compression.sh` - Daily compression script for old databases
- `maintenance.sql` - Maintenance queries and procedures

## Quick Start

### 1. Create a Database

```bash
# Create database directory
mkdir -p data/ticks/db/2026/01

# Initialize database with schema
sqlite3 data/ticks/db/2026/01/ticks_2026-01-20.db < backend/schema/ticks.sql

# Verify schema
sqlite3 data/ticks/db/2026/01/ticks_2026-01-20.db "SELECT name FROM sqlite_master WHERE type='table';"
```

### 2. Configure SQLite for Performance

```sql
-- Execute these PRAGMA statements when opening the database
PRAGMA journal_mode = WAL;              -- Write-Ahead Logging
PRAGMA synchronous = NORMAL;            -- Balance safety vs speed
PRAGMA cache_size = -64000;             -- 64MB cache
PRAGMA temp_store = MEMORY;             -- In-memory temp tables
PRAGMA mmap_size = 268435456;           -- 256MB memory mapping
```

### 3. Insert Ticks

```sql
-- Single insert
INSERT INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
VALUES ('EURUSD', 1737388800000, 1.0850, 1.0852, 0.0002, 'YOFX');

-- Batch insert (recommended for performance)
BEGIN TRANSACTION;
INSERT INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
VALUES
    ('EURUSD', 1737388800000, 1.0850, 1.0852, 0.0002, 'YOFX'),
    ('EURUSD', 1737388801000, 1.0851, 1.0853, 0.0002, 'YOFX'),
    ('GBPUSD', 1737388800000, 1.2650, 1.2652, 0.0002, 'YOFX');
COMMIT;
```

### 4. Query Ticks

```sql
-- Get recent ticks for a symbol
SELECT * FROM ticks
WHERE symbol = 'EURUSD'
ORDER BY timestamp DESC
LIMIT 100;

-- Get ticks for a time range
SELECT * FROM ticks
WHERE symbol = 'EURUSD'
  AND timestamp BETWEEN 1737388800000 AND 1737475200000
ORDER BY timestamp;

-- Get statistics
SELECT
    COUNT(*) as tick_count,
    AVG(spread) as avg_spread,
    MIN(bid) as min_bid,
    MAX(ask) as max_ask
FROM ticks
WHERE symbol = 'EURUSD';
```

## Recommended Architecture

### Option A: SQLite with Daily Partitions (RECOMMENDED)

**Best for:**
- Windows/Linux broker deployments
- Embedded systems (no external database server)
- Small to medium scale (1-100 symbols, 10-50 ticks/sec/symbol)

**Architecture:**
```
data/
└── ticks/
    └── db/
        ├── 2026/
        │   ├── 01/
        │   │   ├── ticks_2026-01-19.db         # 200-500MB uncompressed
        │   │   ├── ticks_2026-01-20.db         # Active (current day)
        │   │   └── ticks_2026-01-21.db         # Pre-created for rotation
        │   └── 02/
        │       └── ...
        └── archive/
            └── compressed/
                └── 2025/
                    └── 12/
                        └── ticks_2025-12-01.db.zst  # 50-100MB compressed
```

**Benefits:**
- ✅ Cross-platform (Windows, Linux, macOS)
- ✅ No external dependencies (embedded)
- ✅ Simple backup (copy files)
- ✅ Fast queries (50K+ inserts/sec with WAL)
- ✅ Natural time-based partitioning
- ✅ Easy to archive/delete old data
- ✅ Crash-resilient (per-day isolation)

**Limitations:**
- Single writer per database (mitigated by daily partitioning)
- Not suitable for distributed systems
- Manual sharding for very large scale

### Option B: PostgreSQL + TimescaleDB (Production-Grade)

**Best for:**
- Large-scale deployments (100+ symbols, 100+ ticks/sec/symbol)
- Multi-server setups with read replicas
- Need for advanced analytics and continuous aggregates

**Benefits:**
- ✅ Automatic partitioning (hypertables)
- ✅ Built-in compression policies
- ✅ Continuous aggregates (materialized views)
- ✅ Horizontal scaling with distributed hypertables
- ✅ Advanced monitoring and replication

**Requirements:**
- PostgreSQL 12+ with TimescaleDB extension
- Database server setup and maintenance
- More complex deployment

### Option C: Apache Parquet (Analytics-Optimized)

**Best for:**
- Data science workflows
- Historical analysis and backtesting
- Integration with Python (Pandas, Polars)

**Benefits:**
- ✅ 10-100x compression vs JSON
- ✅ Columnar storage (fast analytical queries)
- ✅ Direct cloud storage integration (S3, Azure)
- ✅ Schema evolution support

**Limitations:**
- Not suitable for real-time inserts
- Requires specialized libraries
- Best as export format, not primary storage

## Implementation Strategy

### Phase 1: Dual Storage (Migration Period)

1. Keep existing JSON file storage
2. Add SQLite storage in parallel
3. Write to both systems
4. Validate data consistency
5. Switch read queries to SQLite

### Phase 2: Full Migration

1. Migrate historical JSON data to SQLite
2. Disable JSON writes
3. Archive old JSON files
4. Monitor performance and stability

### Phase 3: Optimization

1. Implement compression for old databases
2. Set up archival to cold storage
3. Fine-tune query performance
4. Add monitoring and alerting

## Daily Rotation Process

```
Midnight UTC:
1. Close current database (flush WAL)
2. Create new database for new day
3. Copy schema to new database
4. Switch active database pointer
5. Queue previous database for compression (if > 7 days old)
```

## Compression Strategy

**Timeline:**
- **0-7 days**: Uncompressed (fast queries)
- **7-30 days**: Compressed with zstd (4-5x reduction)
- **30-180 days**: Archived to secondary storage
- **180+ days**: Cold storage (S3, Azure Blob)

**Compression command:**
```bash
# Compress database older than 7 days
zstd -19 --rm ticks_2026-01-10.db -o ticks_2026-01-10.db.zst

# Decompress for queries
zstd -d ticks_2026-01-10.db.zst
```

**Expected compression ratios:**
- zstd level 19: 4-5x (recommended)
- lz4: 2-3x (faster, less compression)
- gzip: 3-4x (slower, compatible)

## Performance Benchmarks

### SQLite (WAL mode, mid-range SSD)

| Operation | Performance | Notes |
|-----------|-------------|-------|
| Single INSERT | 5,000/sec | Without batching |
| Batch INSERT (500 rows) | 30,000-50,000/sec | With transaction |
| Recent ticks (100 rows) | <1ms | Hot query |
| Time range (1 hour) | 5-20ms | ~10K-50K ticks |
| Full day scan (1M ticks) | 100-300ms | Sequential scan |
| Aggregate stats | 50-200ms | With indexes |

### File Sizes (100 symbols, 10 ticks/sec/symbol)

| Period | Tick Count | Uncompressed | Compressed (zstd) |
|--------|------------|--------------|-------------------|
| 1 hour | 3.6M | 50MB | 12MB |
| 1 day | 86.4M | 1.2GB | 250MB |
| 1 week | 605M | 8.4GB | 1.7GB |
| 1 month | 2.6B | 36GB | 7.2GB |

## Data Integrity

### Preventing Duplicates

```sql
-- Using unique index
CREATE UNIQUE INDEX idx_ticks_unique ON ticks(symbol, timestamp);

-- Application-level (Go)
_, err := db.Exec(`
    INSERT OR IGNORE INTO ticks (symbol, timestamp, bid, ask, spread, lp_source)
    VALUES (?, ?, ?, ?, ?, ?)
`, symbol, timestamp, bid, ask, spread, lp)
```

### Detecting Gaps

```sql
-- Find gaps > 60 seconds
SELECT
    symbol,
    datetime(timestamp / 1000, 'unixepoch') as gap_start,
    datetime(lead_timestamp / 1000, 'unixepoch') as gap_end,
    (lead_timestamp - timestamp) / 1000.0 as gap_seconds
FROM (
    SELECT
        symbol,
        timestamp,
        LEAD(timestamp) OVER (PARTITION BY symbol ORDER BY timestamp) as lead_timestamp
    FROM ticks
    WHERE symbol = 'EURUSD'
)
WHERE (lead_timestamp - timestamp) / 1000.0 > 60;
```

### Validating Consistency

```sql
-- Find invalid ticks
SELECT * FROM ticks
WHERE spread < 0                     -- Negative spread
   OR spread > (bid * 0.1)          -- Spread > 10% (likely error)
   OR bid <= 0 OR ask <= 0          -- Invalid prices
   OR ask <= bid                     -- Ask must be > bid
   OR ABS(spread - (ask - bid)) > 0.00001;  -- Spread mismatch
```

## Monitoring

### Key Metrics

1. **Ingestion Rate**: ticks/second per symbol
2. **Storage Growth**: MB/day, GB/month
3. **Query Latency**: p50, p95, p99
4. **Gap Rate**: gaps/day per symbol
5. **Duplicate Rate**: duplicates/total ticks
6. **Disk Usage**: % free space

### Alert Thresholds

- ⚠️ Disk space < 20%
- ⚠️ Ingestion rate drops > 50% for > 5 minutes
- ⚠️ Gap > 5 minutes for active symbol
- ⚠️ Query p95 latency > 100ms
- ⚠️ Database rotation failed at midnight
- ⚠️ Compression job failed

## Backup Strategy

### Continuous Backup

```bash
# Every 5 minutes (WAL files)
rsync -av data/ticks/db/2026/01/*.db-wal /backup/ticks/2026/01/

# Daily (full database after rotation)
rsync -av data/ticks/db/2026/01/ticks_2026-01-19.db /backup/ticks/2026/01/
```

### Point-in-Time Recovery

```bash
# Restore database + replay WAL
cp /backup/ticks/2026/01/ticks_2026-01-19.db data/ticks/db/2026/01/
cp /backup/ticks/2026/01/ticks_2026-01-19.db-wal data/ticks/db/2026/01/
sqlite3 data/ticks/db/2026/01/ticks_2026-01-19.db "PRAGMA wal_checkpoint(TRUNCATE);"
```

### 3-2-1 Backup Rule

- **3 copies**: Local, backup server, cloud
- **2 different media**: SSD + HDD (or cloud)
- **1 offsite**: Cloud storage (S3, Azure Blob)

## Migration from JSON

### Automated Migration Tool

See `migrate_to_sqlite.go` for a complete migration script that:
1. Scans `data/ticks/` directory for JSON files
2. Parses each file and extracts ticks
3. Creates appropriate daily SQLite databases
4. Batch inserts ticks with deduplication
5. Validates tick counts match source
6. Generates migration report

### Manual Migration

```bash
# Run migration tool
go run backend/schema/migrate_to_sqlite.go \
    --input-dir "data/ticks" \
    --output-dir "data/ticks/db" \
    --batch-size 1000 \
    --verbose

# Verify migration
sqlite3 data/ticks/db/2026/01/ticks_2026-01-19.db \
    "SELECT symbol, COUNT(*) FROM ticks GROUP BY symbol;"
```

## Maintenance

### Daily Tasks (Automated)

```bash
# Rotate at midnight
./scripts/rotate_tick_db.sh

# Compress databases > 7 days old
./scripts/compress_old_dbs.sh

# Update statistics
sqlite3 ticks_2026-01-19.db "ANALYZE;"
```

### Weekly Tasks

```bash
# Vacuum to reclaim space
sqlite3 ticks_2026-01-19.db "VACUUM;"

# Verify integrity
sqlite3 ticks_2026-01-19.db "PRAGMA integrity_check;"

# Archive databases > 30 days old
./scripts/archive_to_cold_storage.sh
```

### Monthly Tasks

```bash
# Generate performance report
./scripts/generate_tick_storage_report.sh

# Review storage trends
du -sh data/ticks/db/*/*

# Test backup restore
./scripts/test_backup_restore.sh
```

## Troubleshooting

### Database Locked Error

**Cause**: Another process has exclusive lock on database

**Solutions:**
1. Enable WAL mode: `PRAGMA journal_mode = WAL;`
2. Check for zombie processes: `lsof | grep ticks.db`
3. Increase busy timeout: `PRAGMA busy_timeout = 5000;`

### Slow Queries

**Diagnosis:**
```sql
-- Enable query profiling
EXPLAIN QUERY PLAN SELECT * FROM ticks WHERE symbol = 'EURUSD' LIMIT 100;
```

**Solutions:**
1. Ensure indexes exist: `CREATE INDEX idx_ticks_symbol_timestamp`
2. Analyze tables: `ANALYZE;`
3. Increase cache size: `PRAGMA cache_size = -128000;` (128MB)

### Disk Space Issues

**Check usage:**
```bash
du -sh data/ticks/db/*/*
df -h /path/to/data
```

**Solutions:**
1. Compress old databases: `zstd -19 old_db.db`
2. Archive to cold storage
3. Delete very old databases (after backup)
4. Reduce retention period

## References

- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [SQLite WAL Mode](https://www.sqlite.org/wal.html)
- [TimescaleDB Documentation](https://docs.timescale.com/)
- [Apache Parquet](https://parquet.apache.org/)
- [zstd Compression](https://facebook.github.io/zstd/)

## Support

For questions or issues:
1. Check this README
2. Review `ticks.sql` comments
3. Open GitHub issue with details
