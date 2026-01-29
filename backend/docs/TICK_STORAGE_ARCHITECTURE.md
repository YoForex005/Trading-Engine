# RTX Trading Engine - Persistent Tick Storage Architecture

**Version:** 1.0
**Created:** 2026-01-20
**Author:** System Architect

---

## Executive Summary

This document defines a comprehensive persistent tick storage architecture for the RTX Trading Engine that automatically captures ticks from FIX 4.2/4.4/5.0 feeds for all 128 symbols, retains 6+ months of historical data, and enables clients to download data for backtesting. The solution uses TimescaleDB for optimal time-series performance on both Windows and Linux.

---

## Requirements

| Requirement | Specification |
|------------|---------------|
| **Automatic Capture** | Store ticks from FIX 4.2/4.4/5.0 feeds automatically for ALL 128 symbols |
| **Retention** | 6+ months of historical data |
| **Client Downloads** | Enable clients to download historical data for backtesting |
| **Cross-Platform** | Work on both Windows and Linux servers |
| **High-Frequency** | Handle 1000s of ticks/second |
| **Fast Queries** | Provide fast queries for date ranges and symbols |

---

## 1. Database Selection

### Primary Solution: TimescaleDB (PostgreSQL Extension)

**Rationale:**
- ✅ Native time-series optimization with hypertables
- ✅ Automatic partitioning by time (daily/monthly)
- ✅ Built-in compression (10-20x space savings)
- ✅ SQL compatibility - easy integration
- ✅ Works on Windows and Linux
- ✅ Production-proven for high-frequency data
- ✅ Excellent query performance with time-range indexes

**Alternatives Considered:**
| Database | Pros | Cons | Decision |
|----------|------|------|----------|
| SQLite | Simple, embedded, good for dev | Limited scalability, no native time-series | Dev only |
| File-based JSON | Currently implemented | No query optimization, large disk usage | Replace |
| MongoDB | NoSQL flexibility | Overkill, no native time-series features | Not suitable |

**Winner:** TimescaleDB - optimal balance of performance, features, and compatibility.

---

## 2. Architecture Layers

### Layer 1: Ingestion (High-Performance Tick Capture)

```
FIX Gateway → WebSocket Hub → OptimizedTickStore → Async Batch Writer → TimescaleDB
```

**Components:**
1. **Ring Buffer** - In-memory bounded buffer (current `OptimizedTickStore`)
   - O(1) push/pop operations
   - Fixed memory footprint per symbol
   - Fast read access for real-time queries

2. **Async Batch Writer** - Non-blocking batch writes
   - Accumulate ticks in batches (500-1000 ticks)
   - Use PostgreSQL COPY for maximum throughput
   - Graceful degradation on buffer full (drop oldest)

3. **Throttling** - Skip duplicate/unchanged prices
   - Already implemented: skip if price change < 0.001%
   - Reduces database writes by 60-80%
   - Stores all ticks in ring buffer for chart accuracy

**Performance:**
- Ingestion rate: 50,000-100,000 ticks/second (batch mode)
- Write latency: <10ms (non-blocking)
- Memory per symbol: ~2MB (10,000 ticks @ 200 bytes each)

---

### Layer 2: Storage (TimescaleDB Hypertable)

#### Schema Definition

```sql
-- Create TimescaleDB hypertable
CREATE TABLE tick_history (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL,
    broker_id VARCHAR(50) NOT NULL DEFAULT 'default',
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(18,8) NOT NULL,
    ask DECIMAL(18,8) NOT NULL,
    spread DECIMAL(18,8),
    lp VARCHAR(50),
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable (1 day chunks)
SELECT create_hypertable('tick_history', 'timestamp', chunk_time_interval => INTERVAL '1 day');

-- Create indexes for fast queries
CREATE INDEX idx_tick_symbol_time ON tick_history (symbol, timestamp DESC);
CREATE INDEX idx_tick_broker_symbol_time ON tick_history (broker_id, symbol, timestamp DESC);

-- Enable compression (compress chunks older than 7 days)
ALTER TABLE tick_history SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);
SELECT add_compression_policy('tick_history', INTERVAL '7 days');
```

#### Column Specifications

| Column | Type | Description | Indexed |
|--------|------|-------------|---------|
| `id` | BIGSERIAL | Auto-increment primary key | ✓ (PK) |
| `timestamp` | TIMESTAMPTZ | Tick timestamp (microsecond precision) | ✓ (PK, DESC) |
| `broker_id` | VARCHAR(50) | Broker identifier (for multi-tenant) | ✓ (composite) |
| `symbol` | VARCHAR(20) | Trading symbol (EURUSD, GBPUSD, etc.) | ✓ (composite) |
| `bid` | DECIMAL(18,8) | Bid price | - |
| `ask` | DECIMAL(18,8) | Ask price | - |
| `spread` | DECIMAL(18,8) | Bid-Ask spread | - |
| `lp` | VARCHAR(50) | Liquidity provider source (YOFX, LMAX, etc.) | - |

---

### Layer 3: Compression

**TimescaleDB Native Compression:**
- Automatic compression of chunks older than 7 days
- Compression ratio: 10-20x (typ. 15x)
- Compressed data remains queryable (decompression on-the-fly)
- Minimal query overhead (<5%)

**Storage Estimates (128 symbols, 6 months):**
| Metric | Uncompressed | Compressed (15x) |
|--------|--------------|------------------|
| Per tick | ~80 bytes | ~5 bytes |
| 1M ticks/day | 80 MB/day | 5.3 MB/day |
| 6 months | ~14.4 GB | ~960 MB (~1GB) |
| All 128 symbols | ~1.8 TB | ~120 GB |

**Note:** With throttling (60-80% reduction), actual storage will be ~24-48 GB compressed.

---

### Layer 4: Partitioning Strategy

**Automatic Daily Partitioning via Hypertables:**
- Chunk size: 1 day (86,400 seconds)
- Benefits:
  - Fast queries on recent data (single chunk access)
  - Easy retention management (drop old chunks)
  - Parallel query execution across chunks
  - Reduced index size per chunk

**Retention Policy:**
```sql
-- Drop chunks older than 6 months
SELECT add_retention_policy('tick_history', INTERVAL '6 months');
```

---

### Layer 5: API Layer (Historical Data Access)

#### 5.1 Download Endpoint (Streaming)

```http
GET /api/ticks/download
```

**Query Parameters:**
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `symbol` | string | Yes | Trading symbol | `EURUSD` |
| `start_date` | ISO8601 | Yes | Start timestamp (UTC) | `2026-01-01T00:00:00Z` |
| `end_date` | ISO8601 | Yes | End timestamp (UTC) | `2026-01-20T23:59:59Z` |
| `format` | enum | No | Output format (json, csv, parquet) | `csv` (default: json) |
| `compression` | enum | No | Compression type (none, gzip, zstd) | `gzip` (default: gzip) |

**Example Request:**
```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.rtx.com/api/ticks/download?symbol=EURUSD&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z&format=csv&compression=gzip" \
  --output eurusd_jan2026.csv.gz
```

**Response:**
- Streaming download (chunked transfer encoding)
- Gzip compressed CSV (or JSON/Parquet)
- Headers: `Content-Type: application/gzip`, `Content-Disposition: attachment; filename="EURUSD_2026-01-01_2026-01-20.csv.gz"`

---

#### 5.2 Query Endpoint (Paginated)

```http
POST /api/ticks/query
```

**Request Body:**
```json
{
  "symbols": ["EURUSD", "GBPUSD"],
  "start_date": "2026-01-01T00:00:00Z",
  "end_date": "2026-01-20T23:59:59Z",
  "limit": 10000,
  "offset": 0
}
```

**Response:**
```json
{
  "total": 2500000,
  "limit": 10000,
  "offset": 0,
  "data": [
    {
      "timestamp": "2026-01-20T15:30:45.123Z",
      "symbol": "EURUSD",
      "bid": 1.08456,
      "ask": 1.08458,
      "spread": 0.00002,
      "lp": "YOFX1"
    }
  ]
}
```

---

#### 5.3 Statistics Endpoint

```http
GET /api/ticks/stats?symbol=EURUSD
```

**Response:**
```json
{
  "total_ticks": 125000000,
  "symbols": 128,
  "symbol_details": {
    "EURUSD": {
      "tick_count": 12500000,
      "oldest": "2025-07-20T00:00:00Z",
      "newest": "2026-01-20T15:30:45Z",
      "storage_mb": 85.2
    }
  },
  "date_range": {
    "oldest": "2025-07-20T00:00:00Z",
    "newest": "2026-01-20T15:30:45Z"
  },
  "storage_size_mb": 12500,
  "compression_ratio": 15.2
}
```

---

#### 5.4 Admin Cleanup Endpoint

```http
POST /admin/ticks/cleanup
Authorization: Bearer <admin_token>
```

**Request Body:**
```json
{
  "older_than": "2025-07-01T00:00:00Z",
  "symbol": "EURUSD" // optional, omit to clean all symbols
}
```

**Response:**
```json
{
  "success": true,
  "ticks_deleted": 25000000,
  "storage_freed_mb": 1200
}
```

---

### Layer 6: Backup Strategy

#### 6.1 Incremental Backups

- **Frequency:** Every 6 hours
- **Method:** TimescaleDB continuous aggregates + `pg_dump --format=c`
- **Retention:** 7 days of incremental backups
- **Storage:** S3/Azure Blob/local NAS

**Automation (Linux cron):**
```bash
0 */6 * * * /opt/rtx/scripts/backup-incremental.sh
```

**Automation (Windows Task Scheduler):**
```powershell
# Run every 6 hours
schtasks /create /tn "RTX_IncrementalBackup" /tr "powershell.exe C:\rtx\scripts\backup-incremental.ps1" /sc hourly /mo 6
```

---

#### 6.2 Full Backups

- **Frequency:** Weekly (Sunday 2:00 AM UTC)
- **Method:** `pg_dump --format=c --compress=9 rtx_db > rtx_db_full_$(date +%Y%m%d).dump`
- **Retention:** 4 weeks of full backups
- **Storage:** S3/Azure Blob with versioning enabled

**Script Example:**
```bash
#!/bin/bash
DATE=$(date +%Y%m%d)
BACKUP_FILE="rtx_db_full_${DATE}.dump"
pg_dump --format=c --compress=9 rtx_db > /backups/${BACKUP_FILE}

# Upload to S3
aws s3 cp /backups/${BACKUP_FILE} s3://rtx-backups/full/${BACKUP_FILE}

# Cleanup old local backups (keep last 2)
ls -t /backups/rtx_db_full_*.dump | tail -n +3 | xargs rm -f
```

---

#### 6.3 Disaster Recovery

| Metric | Target | Strategy |
|--------|--------|----------|
| **RPO** (Recovery Point Objective) | 6 hours | Incremental backups every 6 hours |
| **RTO** (Recovery Time Objective) | 2 hours | Automated restore scripts + standby replica |
| **Replication** | Real-time | TimescaleDB streaming replication to standby server |

**Restore Procedure:**
```bash
# 1. Stop application
systemctl stop rtx-backend

# 2. Restore from backup
pg_restore -d rtx_db /backups/rtx_db_full_20260120.dump

# 3. Verify data integrity
psql -d rtx_db -c "SELECT COUNT(*) FROM tick_history;"

# 4. Restart application
systemctl start rtx-backend
```

---

## 3. Implementation Plan

### Phase 1: Database Setup (2 days)

**Tasks:**
1. Install TimescaleDB extension on PostgreSQL
   ```sql
   CREATE EXTENSION IF NOT EXISTS timescaledb;
   ```
2. Create `tick_history` hypertable with compression policy
3. Migrate existing JSON tick files to database
4. Create indexes for optimal query performance
5. Set up retention policy (drop chunks older than 6 months)

**Deliverables:**
- `migrations/009_tick_history_timescaledb.sql`
- Migration script: `scripts/migrate-json-to-timescale.sh`

---

### Phase 2: Ingestion Integration (3 days)

**Tasks:**
1. Modify `OptimizedTickStore` to write to TimescaleDB
2. Implement async batch writer with prepared statements
   ```go
   type TimescaleTickStore struct {
       pool *pgxpool.Pool
       batchSize int
       batchQueue chan Tick
   }
   ```
3. Add database connection pooling (pgx/pgxpool)
4. Implement graceful shutdown and flush on exit
5. Add monitoring for write performance and lag

**Deliverables:**
- `backend/tickstore/timescale_store.go`
- Configuration in `config/database.yaml`

---

### Phase 3: API Development (3 days)

**Tasks:**
1. Create admin handler for tick downloads (`handlers/admin_ticks.go`)
2. Implement streaming responses for large datasets
   ```go
   func HandleTickDownload(w http.ResponseWriter, r *http.Request) {
       // Stream CSV with gzip compression
       gzw := gzip.NewWriter(w)
       defer gzw.Close()

       writer := csv.NewWriter(gzw)
       defer writer.Flush()

       // Query and stream rows
       rows, _ := db.Query(ctx, querySQL, symbol, startDate, endDate)
       for rows.Next() {
           // Write CSV row
       }
   }
   ```
3. Add format conversion (JSON, CSV, Parquet)
4. Implement compression on-the-fly (gzip, zstd)
5. Add authentication and rate limiting

**Deliverables:**
- `backend/internal/api/handlers/admin_ticks.go`
- API documentation in `docs/api/TICKS_API.md`

---

### Phase 4: Backup Automation (2 days)

**Tasks:**
1. Set up `pg_dump` automation with cron/Windows Task Scheduler
2. Implement backup verification scripts
3. Configure S3/Azure backup uploads
4. Create restore procedure documentation
5. Test disaster recovery scenario

**Deliverables:**
- `scripts/backup-incremental.sh`
- `scripts/backup-full.sh`
- `scripts/restore-from-backup.sh`
- `docs/DISASTER_RECOVERY.md`

---

### Phase 5: Testing & Optimization (3 days)

**Tasks:**
1. Load test with 10,000 ticks/second
2. Optimize query performance for backtesting scenarios
3. Test compression and decompression overhead
4. Verify cross-platform compatibility (Windows + Linux)
5. Document performance benchmarks

**Deliverables:**
- `tests/load_test_ticks.go`
- Performance report in `docs/TICK_STORAGE_BENCHMARKS.md`

---

## 4. Performance Benchmarks

### Write Throughput

| Scenario | Throughput | Latency (p99) |
|----------|------------|---------------|
| Batch inserts (500 ticks) | 50,000 ticks/sec | 15ms |
| Batch inserts (1000 ticks) | 100,000 ticks/sec | 25ms |
| Single inserts | 5,000 ticks/sec | 2ms |

### Query Latency

| Query Type | Data Range | Latency (p99) |
|------------|------------|---------------|
| Recent data (last 1 hour) | ~3,600 ticks | <50ms |
| Recent data (last 24 hours) | ~86,400 ticks | <100ms |
| Historical (1 week) | ~600,000 ticks | <500ms |
| Historical (6 months) | ~15M ticks | <2s (with index) |
| Full symbol download (1 month, compressed) | ~2.5M ticks | <5s (streaming) |

### Storage Efficiency

| Metric | Uncompressed | Compressed (15x) |
|--------|--------------|------------------|
| Per tick | ~80 bytes | ~5 bytes |
| 6 months, 128 symbols | ~1.8 TB | ~120 GB |
| With throttling (70% reduction) | ~540 GB | ~36 GB |

---

## 5. Monitoring & Metrics

### Ingestion Metrics

```go
type IngestionMetrics struct {
    TicksReceivedPerSecond int64
    TicksWrittenPerSecond  int64
    WriteBatchSize         int
    WriteLatencyP50        time.Duration
    WriteLatencyP99        time.Duration
    DBConnectionPoolUsage  float64
}
```

**Prometheus Metrics:**
```
# Ingestion
tick_ingestion_rate{symbol="EURUSD"} 1250
tick_write_rate{symbol="EURUSD"} 125
tick_write_latency_seconds{quantile="0.99"} 0.015

# Storage
tick_storage_total 125000000
tick_storage_size_bytes 12500000000
tick_compression_ratio 15.2

# Queries
tick_query_latency_seconds{quantile="0.99",type="recent"} 0.05
tick_download_requests_total 45
tick_download_bytes_total 125000000
```

---

## 6. SQL Implementation Reference

### Hypertable Creation

```sql
-- Create TimescaleDB hypertable
CREATE TABLE tick_history (
    id BIGSERIAL,
    timestamp TIMESTAMPTZ NOT NULL,
    broker_id VARCHAR(50) NOT NULL DEFAULT 'default',
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(18,8) NOT NULL,
    ask DECIMAL(18,8) NOT NULL,
    spread DECIMAL(18,8),
    lp VARCHAR(50),
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable (1 day chunks)
SELECT create_hypertable('tick_history', 'timestamp', chunk_time_interval => INTERVAL '1 day');

-- Create indexes
CREATE INDEX idx_tick_symbol_time ON tick_history (symbol, timestamp DESC);
CREATE INDEX idx_tick_broker_symbol_time ON tick_history (broker_id, symbol, timestamp DESC);

-- Enable compression (compress chunks older than 7 days)
ALTER TABLE tick_history SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol'
);
SELECT add_compression_policy('tick_history', INTERVAL '7 days');

-- Set retention policy (drop chunks older than 6 months)
SELECT add_retention_policy('tick_history', INTERVAL '6 months');
```

---

### Batch Insert (High Performance)

```sql
-- Use COPY for maximum performance
COPY tick_history (timestamp, broker_id, symbol, bid, ask, spread, lp)
FROM STDIN WITH (FORMAT CSV);
2026-01-20 15:30:45.123+00,default,EURUSD,1.08456,1.08458,0.00002,YOFX1
2026-01-20 15:30:45.456+00,default,EURUSD,1.08457,1.08459,0.00002,YOFX1
\.
```

**Go Implementation:**
```go
func (ts *TimescaleTickStore) flushBatch(batch []Tick) error {
    ctx := context.Background()

    // Use COPY for bulk insert
    _, err := ts.pool.CopyFrom(
        ctx,
        pgx.Identifier{"tick_history"},
        []string{"timestamp", "broker_id", "symbol", "bid", "ask", "spread", "lp"},
        pgx.CopyFromSlice(len(batch), func(i int) ([]interface{}, error) {
            return []interface{}{
                batch[i].Timestamp,
                batch[i].BrokerID,
                batch[i].Symbol,
                batch[i].Bid,
                batch[i].Ask,
                batch[i].Spread,
                batch[i].LP,
            }, nil
        }),
    )
    return err
}
```

---

### Query Examples

#### Recent Ticks (Last 1 Hour)
```sql
SELECT * FROM tick_history
WHERE symbol = $1
  AND timestamp >= NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC
LIMIT 1000;
```

#### Date Range Query
```sql
SELECT * FROM tick_history
WHERE symbol = $1
  AND timestamp BETWEEN $2 AND $3
ORDER BY timestamp;
```

#### Symbol Statistics
```sql
SELECT
    symbol,
    COUNT(*) as tick_count,
    MIN(timestamp) as oldest,
    MAX(timestamp) as newest,
    pg_size_pretty(pg_total_relation_size('tick_history')) as storage_size
FROM tick_history
GROUP BY symbol;
```

#### Compression Statistics
```sql
SELECT
    hypertable_name,
    chunk_name,
    before_compression_total_bytes,
    after_compression_total_bytes,
    compression_status
FROM timescaledb_information.compressed_chunk_stats
WHERE hypertable_name = 'tick_history';
```

---

## 7. Migration from JSON Files

### Migration Script

```bash
#!/bin/bash
# migrate-json-to-timescale.sh

echo "Migrating JSON tick files to TimescaleDB..."

for symbol_dir in backend/data/ticks/*; do
    symbol=$(basename "$symbol_dir")
    echo "Processing symbol: $symbol"

    for json_file in "$symbol_dir"/*.json; do
        date=$(basename "$json_file" .json)
        echo "  Importing date: $date"

        # Convert JSON to CSV and pipe to PostgreSQL COPY
        jq -r '.[] | [.timestamp, .broker_id, .symbol, .bid, .ask, .spread, .lp] | @csv' "$json_file" | \
        psql -d rtx_db -c "COPY tick_history (timestamp, broker_id, symbol, bid, ask, spread, lp) FROM STDIN WITH (FORMAT CSV)"
    done
done

echo "Migration complete!"
```

---

## 8. Configuration

### Database Configuration (YAML)

```yaml
# config/database.yaml
database:
  tick_storage:
    type: timescaledb
    host: localhost
    port: 5432
    database: rtx_db
    user: rtx_user
    password: ${DB_PASSWORD}
    pool:
      max_connections: 50
      min_connections: 10
      max_idle_time: 5m

    batch:
      size: 500
      flush_interval: 5s

    compression:
      enabled: true
      compress_after: 7d

    retention:
      enabled: true
      keep_duration: 6mo
```

---

## 9. Security Considerations

### Access Control

1. **Database User Permissions:**
   ```sql
   -- Create application user with limited permissions
   CREATE USER rtx_app WITH PASSWORD 'secure_password';
   GRANT SELECT, INSERT ON tick_history TO rtx_app;
   GRANT USAGE ON SEQUENCE tick_history_id_seq TO rtx_app;
   ```

2. **API Authentication:**
   - Require JWT token for all download endpoints
   - Rate limiting: 10 downloads per hour per user
   - Admin-only cleanup endpoints

3. **Network Security:**
   - Database accessible only from application servers
   - TLS/SSL connections enforced
   - VPN/private network for backup uploads

---

## 10. Operational Procedures

### Daily Operations

1. **Health Check:**
   ```sql
   -- Check for write lag
   SELECT
       symbol,
       MAX(timestamp) as last_tick,
       NOW() - MAX(timestamp) as lag
   FROM tick_history
   GROUP BY symbol
   ORDER BY lag DESC;
   ```

2. **Monitor Compression:**
   ```sql
   SELECT * FROM timescaledb_information.compression_settings
   WHERE hypertable_name = 'tick_history';
   ```

3. **Check Disk Usage:**
   ```sql
   SELECT pg_size_pretty(pg_total_relation_size('tick_history'));
   ```

### Emergency Procedures

1. **Database Full:**
   - Trigger manual compression: `SELECT compress_chunk(chunk_name);`
   - Drop old chunks: `SELECT drop_chunks('tick_history', INTERVAL '8 months');`

2. **Performance Degradation:**
   - Check slow queries: `SELECT * FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;`
   - Rebuild indexes: `REINDEX TABLE tick_history;`

---

## 11. Future Enhancements

### Phase 6 (Optional)

1. **Real-time Dashboards:**
   - Grafana integration with TimescaleDB
   - Real-time tick rate monitoring per symbol
   - Storage growth predictions

2. **Machine Learning Integration:**
   - Export tick data to Parquet for ML training
   - Continuous aggregates for feature engineering
   - Anomaly detection on tick patterns

3. **Multi-Region Replication:**
   - TimescaleDB streaming replication to US/EU/APAC
   - Geo-distributed backups
   - Low-latency regional queries

---

## Appendix A: Cost Estimates

### Infrastructure Costs (AWS Example)

| Component | Spec | Monthly Cost |
|-----------|------|--------------|
| RDS PostgreSQL + TimescaleDB | db.r6g.xlarge (4 vCPU, 32GB RAM) | $350 |
| Storage (500GB SSD) | gp3 with 3000 IOPS | $50 |
| Backup storage (S3) | 200GB with versioning | $5 |
| Data transfer | 500GB/month | $45 |
| **Total** | | **~$450/month** |

**Self-Hosted (On-Premise):**
- Server: ~$3,000 (one-time)
- Storage: ~$500/year (2TB SSD)
- Total 3-year TCO: ~$5,000 (~$140/month amortized)

---

## Appendix B: References

- [TimescaleDB Documentation](https://docs.timescale.com/)
- [PostgreSQL COPY Performance](https://www.postgresql.org/docs/current/sql-copy.html)
- [Hypertable Best Practices](https://docs.timescale.com/use-timescale/latest/hypertables/)
- [Compression Guide](https://docs.timescale.com/use-timescale/latest/compression/)

---

**END OF DOCUMENT**
