# RTX Trading Engine - Tick Storage Implementation Summary

**Date:** 2026-01-20
**Status:** Ready for Implementation
**Priority:** High

---

## Quick Reference

| Component | Status | Location |
|-----------|--------|----------|
| Architecture Document | ✅ Complete | `backend/docs/TICK_STORAGE_ARCHITECTURE.md` |
| TimescaleDB Implementation | ✅ Complete | `backend/tickstore/timescale_store.go` |
| Admin API Handler | ✅ Complete | `backend/internal/api/handlers/admin_ticks.go` |
| Database Migration | ✅ Complete | `backend/migrations/009_tick_history_timescaledb.sql` |
| Configuration Example | ✅ Complete | `backend/config/database.yaml` |
| Migration Script | ✅ Complete | `backend/scripts/migrate-json-to-timescale.sh` |
| API Documentation | ✅ Complete | `backend/docs/api/TICKS_API.md` |

---

## Implementation Checklist

### Phase 1: Database Setup (2 days)

- [ ] **1.1** Install TimescaleDB extension on PostgreSQL
  ```sql
  CREATE EXTENSION IF NOT EXISTS timescaledb;
  ```

- [ ] **1.2** Run migration `009_tick_history_timescaledb.sql`
  ```bash
  psql -h localhost -U rtx_user -d rtx_db -f backend/migrations/009_tick_history_timescaledb.sql
  ```

- [ ] **1.3** Verify hypertable creation
  ```sql
  SELECT * FROM timescaledb_information.hypertables WHERE hypertable_name = 'tick_history';
  ```

- [ ] **1.4** Configure compression policy
  ```sql
  SELECT * FROM timescaledb_information.jobs WHERE hypertable_name = 'tick_history';
  ```

- [ ] **1.5** Migrate existing JSON data to TimescaleDB
  ```bash
  bash backend/scripts/migrate-json-to-timescale.sh
  ```

---

### Phase 2: Backend Integration (3 days)

- [ ] **2.1** Install Go dependencies
  ```bash
  go get github.com/jackc/pgx/v5
  go get github.com/jackc/pgx/v5/pgxpool
  ```

- [ ] **2.2** Update `backend/cmd/server/main.go` to use `TimescaleTickStore`
  ```go
  // Replace:
  // tickStore := tickstore.NewOptimizedTickStore(brokerID, 10000)

  // With:
  cfg := tickstore.TimescaleConfig{
      Host:            "localhost",
      Port:            5432,
      Database:        "rtx_db",
      User:            "rtx_app",
      Password:        os.Getenv("TICK_DB_PASSWORD"),
      MaxConnections:  50,
      MinConnections:  10,
      BrokerID:        brokerID,
      BatchSize:       500,
      FlushInterval:   5 * time.Second,
      MaxTicksPerSymbol: 10000,
  }
  tickStore, err := tickstore.NewTimescaleTickStore(cfg)
  if err != nil {
      log.Fatalf("Failed to create tick store: %v", err)
  }
  ```

- [ ] **2.3** Wire up Admin Tick Handler
  ```go
  adminTickHandler := handlers.NewAdminTickHandler(tickStore)
  http.HandleFunc("/api/ticks/download", adminTickHandler.HandleTickDownload)
  http.HandleFunc("/api/ticks/query", adminTickHandler.HandleTickQuery)
  http.HandleFunc("/api/ticks/stats", adminTickHandler.HandleTickStats)
  http.HandleFunc("/api/ticks/symbols", adminTickHandler.HandleTickSymbols)
  http.HandleFunc("/api/ticks/export-ohlc", adminTickHandler.HandleTickExportOHLC)
  http.HandleFunc("/admin/ticks/cleanup", adminTickHandler.HandleTickCleanup)
  ```

- [ ] **2.4** Add graceful shutdown
  ```go
  // In main.go
  defer tickStore.Stop() // Flush remaining ticks on shutdown
  ```

- [ ] **2.5** Test batch writes
  ```bash
  go test -v backend/tickstore/timescale_store_test.go
  ```

---

### Phase 3: API Development (3 days)

- [ ] **3.1** Test download endpoint
  ```bash
  curl -H "Authorization: Bearer <token>" \
    "http://localhost:8080/api/ticks/download?symbol=EURUSD&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z&format=csv&compression=gzip" \
    --output test.csv.gz
  ```

- [ ] **3.2** Test query endpoint
  ```bash
  curl -X POST -H "Authorization: Bearer <token>" \
    -H "Content-Type: application/json" \
    -d '{"symbols": ["EURUSD"], "start_date": "2026-01-20T00:00:00Z", "end_date": "2026-01-20T23:59:59Z", "limit": 1000, "offset": 0}' \
    http://localhost:8080/api/ticks/query
  ```

- [ ] **3.3** Test stats endpoint
  ```bash
  curl -H "Authorization: Bearer <token>" \
    http://localhost:8080/api/ticks/stats
  ```

- [ ] **3.4** Add rate limiting (10 downloads per hour)
  ```go
  // Use golang.org/x/time/rate
  limiter := rate.NewLimiter(rate.Every(6*time.Minute), 1)
  ```

- [ ] **3.5** Document API in Swagger/OpenAPI

---

### Phase 4: Backup Automation (2 days)

- [ ] **4.1** Set up incremental backups (Linux cron)
  ```bash
  # Edit crontab
  crontab -e

  # Add entry (every 6 hours)
  0 */6 * * * /opt/rtx/scripts/backup-incremental.sh
  ```

- [ ] **4.2** Set up full backups (Sunday 2 AM)
  ```bash
  # Add to crontab
  0 2 * * 0 /opt/rtx/scripts/backup-full.sh
  ```

- [ ] **4.3** Configure S3 backup uploads
  ```bash
  # Install AWS CLI
  apt-get install awscli

  # Configure credentials
  aws configure
  ```

- [ ] **4.4** Test backup restore
  ```bash
  bash backend/scripts/restore-from-backup.sh rtx_db_full_20260120.dump
  ```

- [ ] **4.5** Document disaster recovery procedures

---

### Phase 5: Testing & Optimization (3 days)

- [ ] **5.1** Load test with 10,000 ticks/second
  ```bash
  go test -v -bench=. backend/tickstore/timescale_store_test.go
  ```

- [ ] **5.2** Benchmark query performance
  ```sql
  EXPLAIN ANALYZE SELECT * FROM tick_history WHERE symbol = 'EURUSD' AND timestamp >= NOW() - INTERVAL '1 hour';
  ```

- [ ] **5.3** Test compression (chunks older than 7 days)
  ```sql
  SELECT * FROM get_tick_compression_stats();
  ```

- [ ] **5.4** Test retention policy (drop chunks older than 6 months)
  ```sql
  SELECT drop_chunks('tick_history', INTERVAL '7 months');
  ```

- [ ] **5.5** Verify cross-platform compatibility (Windows + Linux)

---

## Environment Variables

Add to `.env` file:

```bash
# TimescaleDB Connection
DB_HOST=localhost
DB_PORT=5432
DB_NAME=rtx_db
DB_USER=rtx_user
DB_PASSWORD=your_secure_password
TICK_DB_PASSWORD=your_tick_db_password

# AWS S3 Backups (optional)
AWS_ACCESS_KEY=your_access_key
AWS_SECRET_KEY=your_secret_key
AWS_REGION=us-east-1
AWS_BUCKET=rtx-backups
```

---

## Quick Start Guide

### 1. Install TimescaleDB

**Ubuntu/Debian:**
```bash
sudo apt-get install postgresql-14 postgresql-14-timescaledb
sudo systemctl restart postgresql
```

**Windows:**
```powershell
# Download from https://www.timescale.com/download
# Run installer and select TimescaleDB extension
```

### 2. Run Migration

```bash
# Set environment variables
export DB_PASSWORD="your_password"

# Run migration
psql -h localhost -U rtx_user -d rtx_db -f backend/migrations/009_tick_history_timescaledb.sql
```

### 3. Migrate Existing Data

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=rtx_db
export DB_USER=rtx_app
export DB_PASSWORD=your_password
export TICK_DATA_DIR=./backend/data/ticks

# Run migration script
bash backend/scripts/migrate-json-to-timescale.sh
```

### 4. Update Application Code

```go
// In backend/cmd/server/main.go

cfg := tickstore.TimescaleConfig{
    Host:            os.Getenv("DB_HOST"),
    Port:            5432,
    Database:        os.Getenv("DB_NAME"),
    User:            os.Getenv("DB_USER"),
    Password:        os.Getenv("TICK_DB_PASSWORD"),
    MaxConnections:  50,
    MinConnections:  10,
    BrokerID:        "default",
    BatchSize:       500,
    FlushInterval:   5 * time.Second,
    MaxTicksPerSymbol: 10000,
}

tickStore, err := tickstore.NewTimescaleTickStore(cfg)
if err != nil {
    log.Fatalf("Failed to create tick store: %v", err)
}
defer tickStore.Stop()

// Wire up API handlers
adminTickHandler := handlers.NewAdminTickHandler(tickStore)
http.HandleFunc("/api/ticks/download", adminTickHandler.HandleTickDownload)
http.HandleFunc("/api/ticks/query", adminTickHandler.HandleTickQuery)
http.HandleFunc("/api/ticks/stats", adminTickHandler.HandleTickStats)
```

### 5. Test API

```bash
# Get token
TOKEN=$(curl -X POST -H "Content-Type: application/json" \
  -d '{"username": "admin@example.com", "password": "your_password"}' \
  http://localhost:8080/api/login | jq -r .token)

# Download ticks
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/ticks/download?symbol=EURUSD&start_date=2026-01-01T00:00:00Z&end_date=2026-01-20T23:59:59Z&format=csv&compression=gzip" \
  --output eurusd.csv.gz
```

---

## Performance Expectations

| Metric | Target | Actual (After Implementation) |
|--------|--------|-------------------------------|
| Write throughput | 50,000 ticks/sec | _TBD_ |
| Query latency (24h) | <100ms | _TBD_ |
| Query latency (6mo) | <2s | _TBD_ |
| Compression ratio | 15x | _TBD_ |
| Storage (6mo, 128 symbols) | ~50GB | _TBD_ |

---

## Monitoring Queries

### Check Ingestion Rate

```sql
SELECT
    symbol,
    COUNT(*) FILTER (WHERE timestamp >= NOW() - INTERVAL '1 minute') AS ticks_last_minute,
    COUNT(*) FILTER (WHERE timestamp >= NOW() - INTERVAL '1 hour') AS ticks_last_hour
FROM tick_history
GROUP BY symbol
ORDER BY ticks_last_minute DESC;
```

### Check Storage Size

```sql
SELECT
    pg_size_pretty(pg_total_relation_size('tick_history')) AS total_size,
    pg_size_pretty(pg_total_relation_size('tick_history') - pg_total_relation_size('tick_history_compressed')) AS uncompressed_size,
    pg_size_pretty(pg_total_relation_size('tick_history_compressed')) AS compressed_size;
```

### Check Compression Status

```sql
SELECT * FROM get_tick_compression_stats();
```

### Check Symbol Statistics

```sql
SELECT * FROM get_tick_symbol_stats();
```

---

## Troubleshooting

### Issue: Slow Writes

**Symptoms:** Write latency > 50ms, queue building up

**Solutions:**
1. Increase batch size: `cfg.BatchSize = 1000`
2. Increase connection pool: `cfg.MaxConnections = 100`
3. Check database CPU/memory usage
4. Ensure compression is not running during peak hours

### Issue: High Memory Usage

**Symptoms:** Application consuming > 2GB RAM

**Solutions:**
1. Reduce ring buffer size: `cfg.MaxTicksPerSymbol = 5000`
2. Reduce connection pool: `cfg.MaxConnections = 25`
3. Enable throttling (already implemented)

### Issue: Query Timeouts

**Symptoms:** Queries taking > 30s

**Solutions:**
1. Ensure indexes are created
2. Check query plan: `EXPLAIN ANALYZE SELECT ...`
3. Run `ANALYZE tick_history;` to update statistics
4. Consider compressing old chunks manually

---

## Next Steps

1. **Week 1:** Complete Phase 1 (Database Setup)
2. **Week 2:** Complete Phase 2 (Backend Integration) and Phase 3 (API Development)
3. **Week 3:** Complete Phase 4 (Backup Automation) and Phase 5 (Testing)
4. **Week 4:** Deploy to production, monitor performance

---

## References

- **Architecture:** `backend/docs/TICK_STORAGE_ARCHITECTURE.md`
- **API Docs:** `backend/docs/api/TICKS_API.md`
- **TimescaleDB Docs:** https://docs.timescale.com/
- **Migration Script:** `backend/scripts/migrate-json-to-timescale.sh`

---

## Support Contacts

| Component | Contact |
|-----------|---------|
| Database (TimescaleDB) | DBA Team |
| Backend (Go) | Backend Team |
| API | API Team |
| Backups | DevOps Team |

---

**END OF SUMMARY**
