# Tick Data System - Implementation Recommendations

## Priority-Based Roadmap

**Current Status:** ✅ Operational (Grade B+)
**Target Status:** ⭐ Production-Ready (Grade A)

---

## Priority 1: Critical (Week 1)

### 1.1 SQLite Database Migration

**Current:** JSON files only
**Target:** SQLite with proper indexing
**Effort:** 2-3 days
**Impact:** High (performance, scalability)

**Implementation Plan:**

```sql
-- Schema design
CREATE TABLE ticks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    broker_id TEXT NOT NULL DEFAULT 'default',
    symbol TEXT NOT NULL,
    bid REAL NOT NULL,
    ask REAL NOT NULL,
    spread REAL NOT NULL,
    lp TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Indexes for fast queries
CREATE INDEX idx_symbol_timestamp ON ticks(symbol, timestamp DESC);
CREATE INDEX idx_timestamp ON ticks(timestamp DESC);
CREATE INDEX idx_symbol ON ticks(symbol);

-- OHLC aggregation table
CREATE TABLE ohlc (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT NOT NULL,
    timeframe INTEGER NOT NULL, -- in seconds (60, 300, 900, 3600, etc.)
    open REAL NOT NULL,
    high REAL NOT NULL,
    low REAL NOT NULL,
    close REAL NOT NULL,
    volume INTEGER NOT NULL,
    timestamp INTEGER NOT NULL,
    UNIQUE(symbol, timeframe, timestamp)
);

CREATE INDEX idx_ohlc_symbol_tf ON ohlc(symbol, timeframe, timestamp DESC);
```

**Code Changes:**

```go
// backend/tickstore/sqlite_store.go (NEW FILE)
package tickstore

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteTickStore struct {
    db          *sql.DB
    ringBuffers map[string]*TickRingBuffer // Keep for fast recent queries
    mu          sync.RWMutex
}

func NewSQLiteTickStore(dbPath string, maxTicksPerSymbol int) (*SQLiteTickStore, error) {
    db, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=rwc")
    if err != nil {
        return nil, err
    }

    // Enable WAL mode for better concurrency
    db.Exec("PRAGMA journal_mode=WAL")
    db.Exec("PRAGMA synchronous=NORMAL")

    // Create schema
    if err := createSchema(db); err != nil {
        return nil, err
    }

    return &SQLiteTickStore{
        db:          db,
        ringBuffers: make(map[string]*TickRingBuffer),
    }, nil
}

func (s *SQLiteTickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) {
    // Store in ring buffer for fast recent access
    s.storeInRingBuffer(symbol, bid, ask, spread, lp, timestamp)

    // Async insert to SQLite
    go s.insertToDatabase(symbol, bid, ask, spread, lp, timestamp)
}

func (s *SQLiteTickStore) GetHistory(symbol string, limit int) []Tick {
    // First try ring buffer (fastest)
    ticks := s.getRingBufferTicks(symbol, limit)
    if len(ticks) >= limit {
        return ticks
    }

    // Fallback to database for historical data
    return s.getFromDatabase(symbol, limit)
}
```

**Migration Script:**

```bash
#!/bin/bash
# backend/scripts/migrate_json_to_sqlite.sh

echo "Migrating tick data from JSON to SQLite..."

TICK_DIR="data/ticks"
SQLITE_DB="data/ticks.db"

# Create SQLite database
sqlite3 "$SQLITE_DB" < backend/tickstore/schema.sql

# Migrate each symbol
for symbol_dir in "$TICK_DIR"/*; do
    symbol=$(basename "$symbol_dir")
    echo "Migrating $symbol..."

    # Process all JSON files for this symbol
    for json_file in "$symbol_dir"/*.json; do
        echo "  Processing $(basename $json_file)..."

        # Use jq to extract ticks and insert into SQLite
        jq -r '.[] | [.broker_id, .symbol, .bid, .ask, .spread, .lp, (.timestamp | fromdateiso8601)] | @csv' \
            "$json_file" | \
        while IFS=, read -r broker_id symbol bid ask spread lp timestamp; do
            sqlite3 "$SQLITE_DB" \
                "INSERT INTO ticks (broker_id, symbol, bid, ask, spread, lp, timestamp) \
                 VALUES ($broker_id, $symbol, $bid, $ask, $spread, $lp, $timestamp)"
        done
    done
done

echo "Migration complete!"

# Verify
echo "Total ticks migrated:"
sqlite3 "$SQLITE_DB" "SELECT COUNT(*) FROM ticks"
```

**Testing:**

```bash
# Test query performance
sqlite3 data/ticks.db "EXPLAIN QUERY PLAN SELECT * FROM ticks WHERE symbol='EURUSD' ORDER BY timestamp DESC LIMIT 100"

# Benchmark
time sqlite3 data/ticks.db "SELECT * FROM ticks WHERE symbol='EURUSD' AND timestamp > strftime('%s', 'now', '-1 hour') LIMIT 1000"
```

---

### 1.2 Automated Retention Policy

**Current:** Manual cleanup only
**Target:** Automated 6-month retention
**Effort:** 1 day
**Impact:** High (prevents unbounded growth)

**Implementation:**

```go
// backend/tickstore/retention.go (NEW FILE)
package tickstore

import (
    "database/sql"
    "log"
    "time"
)

type RetentionPolicy struct {
    db              *sql.DB
    retentionPeriod time.Duration
    cleanupInterval time.Duration
    stopChan        chan struct{}
}

func NewRetentionPolicy(db *sql.DB, retentionMonths int) *RetentionPolicy {
    return &RetentionPolicy{
        db:              db,
        retentionPeriod: time.Duration(retentionMonths) * 30 * 24 * time.Hour,
        cleanupInterval: 24 * time.Hour, // Run daily
        stopChan:        make(chan struct{}),
    }
}

func (rp *RetentionPolicy) Start() {
    go rp.cleanupLoop()
}

func (rp *RetentionPolicy) Stop() {
    close(rp.stopChan)
}

func (rp *RetentionPolicy) cleanupLoop() {
    ticker := time.NewTicker(rp.cleanupInterval)
    defer ticker.Stop()

    // Run immediately on start
    rp.cleanup()

    for {
        select {
        case <-rp.stopChan:
            return
        case <-ticker.C:
            rp.cleanup()
        }
    }
}

func (rp *RetentionPolicy) cleanup() {
    cutoffTime := time.Now().Add(-rp.retentionPeriod).Unix()

    log.Printf("[Retention] Starting cleanup for data older than %v", rp.retentionPeriod)

    // Delete old ticks
    result, err := rp.db.Exec("DELETE FROM ticks WHERE timestamp < ?", cutoffTime)
    if err != nil {
        log.Printf("[Retention] Failed to delete old ticks: %v", err)
        return
    }

    rowsAffected, _ := result.RowsAffected()
    log.Printf("[Retention] Deleted %d old ticks", rowsAffected)

    // Delete old OHLC data
    result, err = rp.db.Exec("DELETE FROM ohlc WHERE timestamp < ?", cutoffTime)
    if err != nil {
        log.Printf("[Retention] Failed to delete old OHLC: %v", err)
        return
    }

    rowsAffected, _ = result.RowsAffected()
    log.Printf("[Retention] Deleted %d old OHLC bars", rowsAffected)

    // Vacuum database to reclaim space
    log.Printf("[Retention] Running VACUUM to reclaim disk space...")
    _, err = rp.db.Exec("VACUUM")
    if err != nil {
        log.Printf("[Retention] VACUUM failed: %v", err)
    } else {
        log.Printf("[Retention] VACUUM completed successfully")
    }
}
```

**Integration in main.go:**

```go
// backend/cmd/server/main.go

// Initialize tick store with SQLite
tickStore, err := tickstore.NewSQLiteTickStore("data/ticks.db", brokerConfig.MaxTicksPerSymbol)
if err != nil {
    log.Fatalf("Failed to initialize tick store: %v", err)
}

// Start retention policy (6 months)
retentionPolicy := tickstore.NewRetentionPolicy(tickStore.GetDB(), 6)
retentionPolicy.Start()
defer retentionPolicy.Stop()

log.Println("[Retention] 6-month retention policy started (runs daily at 2 AM)")
```

---

### 1.3 API Rate Limiting

**Current:** No rate limiting
**Target:** 10 req/sec per IP, burst 20
**Effort:** 0.5 day
**Impact:** Medium (security, stability)

**Implementation:**

```go
// backend/api/middleware/rate_limit.go (NEW FILE)
package middleware

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

type IPRateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }

    return limiter
}

func (rl *IPRateLimiter) Limit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr

        limiter := rl.getLimiter(ip)
        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

**Integration:**

```go
// backend/cmd/server/main.go

// Create rate limiter (10 req/sec, burst 20)
rateLimiter := middleware.NewIPRateLimiter(rate.Limit(10), 20)

// Apply to tick data endpoints
http.Handle("/ticks", rateLimiter.Limit(http.HandlerFunc(server.HandleGetTicks)))
http.Handle("/ohlc", rateLimiter.Limit(http.HandlerFunc(server.HandleGetOHLC)))
http.Handle("/api/history/ticks/", rateLimiter.Limit(http.HandlerFunc(historyHandler.HandleGetTicks)))

log.Println("[RateLimit] API rate limiting enabled (10 req/sec per IP)")
```

---

## Priority 2: Important (Week 2)

### 2.1 File Compression

**Current:** All files uncompressed
**Target:** Gzip files older than 7 days
**Effort:** 1 day
**Impact:** Medium (3-5x storage savings)

**Implementation:**

```go
// backend/tickstore/compression.go (NEW FILE)
package tickstore

import (
    "compress/gzip"
    "io"
    "log"
    "os"
    "path/filepath"
    "time"
)

type CompressionService struct {
    dataDir         string
    compressionAge  time.Duration
    compressionInterval time.Duration
    stopChan        chan struct{}
}

func NewCompressionService(dataDir string, compressionAgeDays int) *CompressionService {
    return &CompressionService{
        dataDir:         dataDir,
        compressionAge:  time.Duration(compressionAgeDays) * 24 * time.Hour,
        compressionInterval: 24 * time.Hour,
        stopChan:        make(chan struct{}),
    }
}

func (cs *CompressionService) Start() {
    go cs.compressionLoop()
}

func (cs *CompressionService) Stop() {
    close(cs.stopChan)
}

func (cs *CompressionService) compressionLoop() {
    ticker := time.NewTicker(cs.compressionInterval)
    defer ticker.Stop()

    // Run immediately
    cs.compress()

    for {
        select {
        case <-cs.stopChan:
            return
        case <-ticker.C:
            cs.compress()
        }
    }
}

func (cs *CompressionService) compress() {
    log.Println("[Compression] Starting compression of old tick files...")

    cutoffTime := time.Now().Add(-cs.compressionAge)
    compressed := 0

    filepath.Walk(cs.dataDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // Only compress .json files (not already compressed)
        if !info.IsDir() && filepath.Ext(path) == ".json" {
            if info.ModTime().Before(cutoffTime) {
                if err := cs.compressFile(path); err != nil {
                    log.Printf("[Compression] Failed to compress %s: %v", path, err)
                } else {
                    compressed++
                }
            }
        }

        return nil
    })

    log.Printf("[Compression] Compressed %d files", compressed)
}

func (cs *CompressionService) compressFile(srcPath string) error {
    // Open source file
    srcFile, err := os.Open(srcPath)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    // Create gzip file
    dstPath := srcPath + ".gz"
    dstFile, err := os.Create(dstPath)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    // Create gzip writer
    gzipWriter := gzip.NewWriter(dstFile)
    defer gzipWriter.Close()

    // Copy and compress
    _, err = io.Copy(gzipWriter, srcFile)
    if err != nil {
        return err
    }

    // Delete original file after successful compression
    return os.Remove(srcPath)
}
```

---

### 2.2 Admin Authentication

**Current:** No auth on `/admin/*`
**Target:** JWT with admin role
**Effort:** 1 day
**Impact:** Medium (security)

**Implementation:**

```go
// backend/api/middleware/admin_auth.go (NEW FILE)
package middleware

import (
    "net/http"
    "strings"
    "github.com/epic1st/rtx/backend/auth"
)

func RequireAdmin(authService *auth.Service) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            token := strings.TrimPrefix(authHeader, "Bearer ")

            // Validate token and check admin role
            user, err := authService.ValidateToken(token)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            if !user.IsAdmin {
                http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
                return
            }

            // User is admin, proceed
            next.ServeHTTP(w, r)
        })
    }
}
```

**Apply to admin endpoints:**

```go
// Protect admin endpoints
adminAuth := middleware.RequireAdmin(authService)

http.Handle("/admin/history/stats", adminAuth(http.HandlerFunc(adminHistoryHandler.HandleGetStats)))
http.Handle("/admin/history/cleanup", adminAuth(http.HandlerFunc(adminHistoryHandler.HandleCleanupOldData)))
http.Handle("/admin/history/compress", adminAuth(http.HandlerFunc(adminHistoryHandler.HandleCompressData)))
http.Handle("/admin/fix/status", adminAuth(http.HandlerFunc(...)))
```

---

### 2.3 Monitoring & Alerts

**Effort:** 1 day
**Impact:** Medium (proactive issue detection)

**Metrics to Monitor:**

1. **Disk Space**
   - Alert if <10GB free
   - Alert if growth >20GB/day

2. **Write Failures**
   - Alert on failed batch writes
   - Alert on database errors

3. **Performance**
   - Alert if API latency >1s
   - Alert if tick rate <10/sec (during market hours)

4. **Data Quality**
   - Alert if no new data for >5 minutes (market hours)
   - Alert if duplicate tick rate >95%

**Implementation:** Integrate with existing alert system (backend/internal/alerts)

---

## Priority 3: Nice to Have (Weeks 3-4)

### 3.1 Time-Series Database Migration

**Consider:** InfluxDB or TimescaleDB
**Effort:** 1 week
**Impact:** High (long-term performance)

**Benefits:**
- Optimized for time-series queries
- Better compression (10-100x)
- Built-in downsampling
- Continuous aggregates

**Decision Point:** Evaluate after SQLite migration

---

### 3.2 Analytics Dashboard

**Features:**
- Tick volume charts
- Symbol activity heatmap
- Storage usage trends
- API usage metrics

**Effort:** 3 days
**Impact:** Low (nice to have)

---

### 3.3 Cold Storage Archive

**Implementation:**
- Move data >3 months to S3/Azure Blob
- Keep recent data in hot storage
- On-demand retrieval for old data

**Effort:** 2 days
**Impact:** Low (cost optimization)

---

## Implementation Timeline

### Week 1 (Priority 1)

| Day | Task | Deliverable |
|-----|------|-------------|
| Mon | SQLite schema design | schema.sql |
| Tue | SQLite migration script | sqlite_store.go |
| Wed | Test SQLite migration | Migrated data |
| Thu | Retention policy impl | retention.go |
| Fri | Rate limiting | middleware/rate_limit.go |

### Week 2 (Priority 2)

| Day | Task | Deliverable |
|-----|------|-------------|
| Mon | File compression | compression.go |
| Tue | Admin auth | middleware/admin_auth.go |
| Wed | Monitoring alerts | alert integration |
| Thu | Testing & QA | Test report |
| Fri | Documentation | Updated docs |

### Week 3-4 (Priority 3 - Optional)

| Task | Status |
|------|--------|
| Time-Series DB evaluation | Research phase |
| Analytics dashboard | Design phase |
| Cold storage | Planning phase |

---

## Success Metrics

### After Priority 1 Implementation

- ✅ Query latency <100ms (vs current 50-200ms)
- ✅ Storage growth controlled (<5GB/month)
- ✅ Zero rate limit abuse incidents
- ✅ Automated retention running daily

### After Priority 2 Implementation

- ✅ Storage reduced by 60-70% (compression)
- ✅ Zero unauthorized admin access
- ✅ Proactive issue detection (alerts)
- ✅ All production requirements met

---

## Risk Mitigation

### Backup Strategy

```bash
# Daily automated backup
#!/bin/bash
# scripts/backup_tick_data.sh

BACKUP_DIR="/backups/tick_data"
DATE=$(date +%Y%m%d)

# Backup SQLite database
sqlite3 data/ticks.db ".backup $BACKUP_DIR/ticks_$DATE.db"

# Compress backup
gzip "$BACKUP_DIR/ticks_$DATE.db"

# Retain last 30 days
find "$BACKUP_DIR" -name "ticks_*.db.gz" -mtime +30 -delete
```

### Rollback Plan

1. Keep JSON files for 30 days after SQLite migration
2. Tag SQLite migration commit in git
3. Document rollback procedure
4. Test rollback in staging environment

---

## Cost-Benefit Analysis

### Current State (JSON)

- Storage: 165MB today, 10-15GB in 6 months
- Query latency: 50-200ms (file I/O)
- Maintenance: Manual cleanup required
- Risk: Unbounded growth, no rate limiting

### Future State (SQLite + Optimizations)

- Storage: 50-100MB today, 2-3GB in 6 months (with compression)
- Query latency: <100ms (indexed queries)
- Maintenance: Fully automated
- Risk: Mitigated

**Savings:**
- 60-70% storage reduction
- 2-3x faster queries
- Zero manual intervention
- Production-ready stability

---

## Conclusion

**Recommended Action:** Implement Priority 1 items immediately (Week 1)

**Justification:**
- System is operational but not production-ready
- SQLite migration is critical for scalability
- Retention policy prevents unbounded growth
- Rate limiting protects against abuse

**Total Effort:** 1-2 weeks for production-ready system

---

**Last Updated:** January 20, 2026
**Status:** Ready for Implementation
