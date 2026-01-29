# Database Migration Plan

## Executive Summary

This document outlines a phased, zero-downtime migration from in-memory storage to a production PostgreSQL/TimescaleDB stack.

## Current State Analysis

### Existing Data Structures

```go
// In-Memory Storage
- Accounts: map[int64]*Account
- Positions: map[int64]*Position
- Orders: map[int64]*Order
- Trades: []Trade
- Ledger: map[int64][]LedgerEntry
- Ticks: JSON files (~81MB)
- OHLC: JSON files per symbol/timeframe
```

### Data Volume Estimates

| Data Type | Current Size | Growth Rate | 30-Day Projection |
|-----------|--------------|-------------|-------------------|
| Accounts | ~100 records | 10/day | 400 records |
| Positions | ~50 active | 20/day | 650 records |
| Orders | ~200/day | 200/day | 6,000 records |
| Trades | ~150/day | 150/day | 4,500 records |
| Tick Data | 81MB | ~3MB/day | 170MB |
| OHLC Data | ~50MB | ~1MB/day | 80MB |

## Migration Phases

### Phase 1: Infrastructure Setup (Week 1)

#### Objectives
- Set up PostgreSQL and TimescaleDB instances
- Configure Redis cache
- Set up message queue (NATS/RabbitMQ)
- Configure monitoring and backups

#### Tasks

1. **Database Provisioning**
```bash
# PostgreSQL (Primary)
docker run -d \
  --name trading-postgres \
  -e POSTGRES_DB=trading_engine \
  -e POSTGRES_USER=trading_app \
  -e POSTGRES_PASSWORD=secure_password \
  -p 5432:5432 \
  -v /data/postgres:/var/lib/postgresql/data \
  postgres:15-alpine

# TimescaleDB (Time-Series)
docker run -d \
  --name trading-timescale \
  -e POSTGRES_DB=market_data \
  -e POSTGRES_USER=trading_app \
  -e POSTGRES_PASSWORD=secure_password \
  -p 5433:5432 \
  -v /data/timescale:/var/lib/postgresql/data \
  timescale/timescaledb:latest-pg15

# Redis (Cache)
docker run -d \
  --name trading-redis \
  -p 6379:6379 \
  -v /data/redis:/data \
  redis:7-alpine redis-server --appendonly yes
```

2. **Schema Creation**
```bash
# Apply PostgreSQL schema
psql -h localhost -p 5432 -U trading_app -d trading_engine \
  -f docs/database/02-ddl-schema.sql

# Apply TimescaleDB schema
psql -h localhost -p 5433 -U trading_app -d market_data \
  -f docs/database/03-timescaledb-schema.sql
```

3. **Connection Pooling Setup**
```bash
# Install PgBouncer
apt-get install pgbouncer

# Configure pgbouncer.ini
[databases]
trading_engine = host=localhost port=5432 dbname=trading_engine
market_data = host=localhost port=5433 dbname=market_data

[pgbouncer]
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
reserve_pool_size = 5
reserve_pool_timeout = 3
```

4. **Monitoring Setup**
```bash
# Prometheus PostgreSQL Exporter
docker run -d \
  --name postgres-exporter \
  -p 9187:9187 \
  -e DATA_SOURCE_NAME="postgresql://trading_app:password@localhost:5432/trading_engine?sslmode=disable" \
  prometheuscommunity/postgres-exporter

# Grafana
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -v /data/grafana:/var/lib/grafana \
  grafana/grafana
```

#### Deliverables
- [ ] PostgreSQL instance running
- [ ] TimescaleDB instance running
- [ ] Redis instance running
- [ ] PgBouncer configured
- [ ] Monitoring dashboards configured
- [ ] Backup scripts automated

### Phase 2: Data Migration (Week 2)

#### Objectives
- Migrate existing in-memory data to PostgreSQL
- ETL historical tick data to TimescaleDB
- Validate data integrity
- Benchmark query performance

#### Migration Scripts

**Script 1: Migrate Accounts**

```go
// backend/migration/migrate_accounts.go
package migration

import (
    "database/sql"
    "log"
    "github.com/epic1st/rtx/backend/internal/core"
)

func MigrateAccounts(engine *core.Engine, db *sql.DB) error {
    stmt, err := db.Prepare(`
        INSERT INTO accounts (
            id, account_number, user_id, account_type, currency,
            balance, leverage, margin_mode, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
        ON CONFLICT (id) DO UPDATE SET
            balance = EXCLUDED.balance,
            equity = EXCLUDED.equity,
            updated_at = NOW()
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    // Get all accounts from in-memory engine
    accounts := engine.GetAllAccounts()

    for _, acc := range accounts {
        accountType := "LIVE"
        if acc.IsDemo {
            accountType = "DEMO"
        }

        _, err = stmt.Exec(
            acc.ID,
            acc.AccountNumber,
            acc.UserID,
            accountType,
            acc.Currency,
            acc.Balance,
            acc.Leverage,
            acc.MarginMode,
            acc.Status,
        )
        if err != nil {
            log.Printf("Failed to migrate account %s: %v", acc.AccountNumber, err)
            continue
        }
    }

    log.Printf("Migrated %d accounts", len(accounts))
    return nil
}
```

**Script 2: Migrate Positions**

```go
// backend/migration/migrate_positions.go
package migration

import (
    "database/sql"
    "log"
    "github.com/epic1st/rtx/backend/internal/core"
)

func MigratePositions(engine *core.Engine, db *sql.DB) error {
    stmt, err := db.Prepare(`
        INSERT INTO positions (
            id, account_id, symbol, side, volume, open_price,
            current_price, stop_loss, take_profit, commission,
            swap, unrealized_pnl, status, open_time
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        ON CONFLICT (id) DO UPDATE SET
            current_price = EXCLUDED.current_price,
            unrealized_pnl = EXCLUDED.unrealized_pnl,
            status = EXCLUDED.status
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    positions := engine.GetAllPositions()

    for _, pos := range positions {
        _, err = stmt.Exec(
            pos.ID,
            pos.AccountID,
            pos.Symbol,
            pos.Side,
            pos.Volume,
            pos.OpenPrice,
            pos.CurrentPrice,
            pos.SL,
            pos.TP,
            pos.Commission,
            pos.Swap,
            pos.UnrealizedPnL,
            pos.Status,
            pos.OpenTime,
        )
        if err != nil {
            log.Printf("Failed to migrate position %d: %v", pos.ID, err)
            continue
        }
    }

    log.Printf("Migrated %d positions", len(positions))
    return nil
}
```

**Script 3: Migrate Ledger**

```go
// backend/migration/migrate_ledger.go
package migration

import (
    "database/sql"
    "log"
    "github.com/epic1st/rtx/backend/bbook"
)

func MigrateLedger(ledger *bbook.Ledger, db *sql.DB) error {
    stmt, err := db.Prepare(`
        INSERT INTO ledger_entries (
            id, account_id, transaction_type, amount,
            balance_before, balance_after, currency, description,
            reference_type, reference_id, payment_method,
            payment_reference, status, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        ON CONFLICT (id) DO NOTHING
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    entries := ledger.GetAllEntries(0) // 0 = no limit

    for _, entry := range entries {
        balanceBefore := entry.BalanceAfter - entry.Amount

        _, err = stmt.Exec(
            entry.ID,
            entry.AccountID,
            entry.Type,
            entry.Amount,
            balanceBefore,
            entry.BalanceAfter,
            entry.Currency,
            entry.Description,
            entry.RefType,
            entry.RefID,
            entry.PaymentMethod,
            entry.PaymentRef,
            entry.Status,
            entry.CreatedAt,
        )
        if err != nil {
            log.Printf("Failed to migrate ledger entry %d: %v", entry.ID, err)
            continue
        }
    }

    log.Printf("Migrated %d ledger entries", len(entries))
    return nil
}
```

**Script 4: Migrate Tick Data**

```go
// backend/migration/migrate_ticks.go
package migration

import (
    "database/sql"
    "encoding/json"
    "io/ioutil"
    "log"
    "path/filepath"
    "time"
)

type Tick struct {
    Timestamp int64   `json:"timestamp"`
    Bid       float64 `json:"bid"`
    Ask       float64 `json:"ask"`
}

func MigrateTickData(ticksDir string, db *sql.DB) error {
    stmt, err := db.Prepare(`
        INSERT INTO ticks (time, symbol, bid, ask, liquidity_provider)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (time, symbol) DO NOTHING
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    // Process each symbol directory
    symbols := []string{"BTCUSD", "ETHUSD", "BNBUSD", "SOLUSD", "XRPUSD", "EURUSD"}

    for _, symbol := range symbols {
        symbolDir := filepath.Join(ticksDir, symbol)
        files, err := ioutil.ReadDir(symbolDir)
        if err != nil {
            log.Printf("Failed to read directory %s: %v", symbolDir, err)
            continue
        }

        for _, file := range files {
            if filepath.Ext(file.Name()) != ".json" {
                continue
            }

            filePath := filepath.Join(symbolDir, file.Name())
            data, err := ioutil.ReadFile(filePath)
            if err != nil {
                log.Printf("Failed to read file %s: %v", filePath, err)
                continue
            }

            var ticks []Tick
            if err := json.Unmarshal(data, &ticks); err != nil {
                log.Printf("Failed to parse JSON %s: %v", filePath, err)
                continue
            }

            // Batch insert for performance
            tx, err := db.Begin()
            if err != nil {
                return err
            }

            txStmt := tx.Stmt(stmt)
            for _, tick := range ticks {
                timestamp := time.Unix(tick.Timestamp/1000, (tick.Timestamp%1000)*1000000)
                _, err = txStmt.Exec(timestamp, symbol, tick.Bid, tick.Ask, "HISTORICAL")
                if err != nil {
                    log.Printf("Failed to insert tick: %v", err)
                }
            }

            if err := tx.Commit(); err != nil {
                log.Printf("Failed to commit transaction: %v", err)
            }

            log.Printf("Migrated %d ticks for %s from %s", len(ticks), symbol, file.Name())
        }
    }

    return nil
}
```

**Migration Runner**

```go
// backend/cmd/migrate/main.go
package main

import (
    "database/sql"
    "log"

    _ "github.com/lib/pq"
    "github.com/epic1st/rtx/backend/internal/core"
    "github.com/epic1st/rtx/backend/migration"
)

func main() {
    // Connect to PostgreSQL
    pgDB, err := sql.Open("postgres",
        "host=localhost port=6432 user=trading_app password=secure_password dbname=trading_engine sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer pgDB.Close()

    // Connect to TimescaleDB
    tsDB, err := sql.Open("postgres",
        "host=localhost port=5433 user=trading_app password=secure_password dbname=market_data sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer tsDB.Close()

    // Initialize in-memory engine (load current state)
    engine := core.NewEngine()
    // ... load data into engine

    // Run migrations
    log.Println("Starting migration...")

    if err := migration.MigrateAccounts(engine, pgDB); err != nil {
        log.Fatalf("Account migration failed: %v", err)
    }

    if err := migration.MigratePositions(engine, pgDB); err != nil {
        log.Fatalf("Position migration failed: %v", err)
    }

    if err := migration.MigrateLedger(engine.GetLedger(), pgDB); err != nil {
        log.Fatalf("Ledger migration failed: %v", err)
    }

    if err := migration.MigrateTickData("./data/ticks", tsDB); err != nil {
        log.Fatalf("Tick data migration failed: %v", err)
    }

    log.Println("Migration completed successfully")
}
```

#### Data Validation

```go
// backend/migration/validate.go
package migration

import (
    "database/sql"
    "fmt"
    "log"
    "github.com/epic1st/rtx/backend/internal/core"
)

func ValidateMigration(engine *core.Engine, db *sql.DB) error {
    // Validate account count
    var dbAccountCount int
    err := db.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&dbAccountCount)
    if err != nil {
        return err
    }

    memAccountCount := len(engine.GetAllAccounts())
    if dbAccountCount != memAccountCount {
        return fmt.Errorf("account count mismatch: DB=%d, Memory=%d",
            dbAccountCount, memAccountCount)
    }

    // Validate balances
    accounts := engine.GetAllAccounts()
    for _, acc := range accounts {
        var dbBalance float64
        err := db.QueryRow("SELECT balance FROM accounts WHERE id = $1", acc.ID).
            Scan(&dbBalance)
        if err != nil {
            return fmt.Errorf("failed to query account %d: %v", acc.ID, err)
        }

        if abs(dbBalance - acc.Balance) > 0.01 {
            return fmt.Errorf("balance mismatch for account %d: DB=%.2f, Memory=%.2f",
                acc.ID, dbBalance, acc.Balance)
        }
    }

    log.Println("✓ Validation passed: All data matches")
    return nil
}

func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
```

#### Deliverables
- [ ] All accounts migrated
- [ ] All positions migrated
- [ ] All orders migrated
- [ ] All ledger entries migrated
- [ ] All tick data migrated
- [ ] Data validation passed
- [ ] Performance benchmarks completed

### Phase 3: Dual-Write Implementation (Week 3)

#### Objectives
- Implement dual-write to both in-memory and database
- Compare results for consistency
- Monitor performance impact
- Fix any discrepancies

#### Implementation

```go
// backend/internal/core/dual_write_engine.go
package core

import (
    "database/sql"
    "log"
    "sync"
)

type DualWriteEngine struct {
    memoryEngine *Engine
    db           *sql.DB
    enabled      bool
    mu           sync.RWMutex
}

func NewDualWriteEngine(memEngine *Engine, db *sql.DB) *DualWriteEngine {
    return &DualWriteEngine{
        memoryEngine: memEngine,
        db:           db,
        enabled:      true,
    }
}

func (e *DualWriteEngine) ExecuteMarketOrder(accountID int64, symbol, side string, volume, sl, tp float64) (*Position, error) {
    // Execute in memory first
    pos, err := e.memoryEngine.ExecuteMarketOrder(accountID, symbol, side, volume, sl, tp)
    if err != nil {
        return nil, err
    }

    // Write to database asynchronously
    if e.enabled {
        go e.writePositionToDB(pos)
    }

    return pos, nil
}

func (e *DualWriteEngine) writePositionToDB(pos *Position) {
    stmt, err := e.db.Prepare(`
        INSERT INTO positions (
            id, account_id, symbol, side, volume, open_price,
            current_price, stop_loss, take_profit, commission,
            swap, unrealized_pnl, status, open_time
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
    `)
    if err != nil {
        log.Printf("ERROR: Failed to prepare statement: %v", err)
        return
    }
    defer stmt.Close()

    _, err = stmt.Exec(
        pos.ID, pos.AccountID, pos.Symbol, pos.Side, pos.Volume,
        pos.OpenPrice, pos.CurrentPrice, pos.SL, pos.TP,
        pos.Commission, pos.Swap, pos.UnrealizedPnL,
        pos.Status, pos.OpenTime,
    )
    if err != nil {
        log.Printf("ERROR: Failed to write position to DB: %v", err)
    }
}
```

#### Monitoring

```go
// backend/internal/monitoring/dual_write_monitor.go
package monitoring

import (
    "database/sql"
    "log"
    "time"
)

type DualWriteMonitor struct {
    db          *sql.DB
    memEngine   *core.Engine
    checkInterval time.Duration
}

func NewDualWriteMonitor(db *sql.DB, engine *core.Engine) *DualWriteMonitor {
    return &DualWriteMonitor{
        db:          db,
        memEngine:   engine,
        checkInterval: 1 * time.Minute,
    }
}

func (m *DualWriteMonitor) Start() {
    ticker := time.NewTicker(m.checkInterval)
    for range ticker.C {
        m.checkConsistency()
    }
}

func (m *DualWriteMonitor) checkConsistency() {
    // Compare position counts
    var dbPositionCount int
    err := m.db.QueryRow("SELECT COUNT(*) FROM positions WHERE status = 'OPEN'").
        Scan(&dbPositionCount)
    if err != nil {
        log.Printf("ERROR: Failed to query position count: %v", err)
        return
    }

    memPositionCount := len(m.memEngine.GetAllPositions())

    if dbPositionCount != memPositionCount {
        log.Printf("WARNING: Position count mismatch: DB=%d, Memory=%d",
            dbPositionCount, memPositionCount)
    }

    // Compare account balances
    accounts := m.memEngine.GetAllAccounts()
    for _, acc := range accounts {
        var dbBalance float64
        err := m.db.QueryRow("SELECT balance FROM accounts WHERE id = $1", acc.ID).
            Scan(&dbBalance)
        if err != nil {
            log.Printf("ERROR: Failed to query balance for account %d: %v", acc.ID, err)
            continue
        }

        if abs(dbBalance - acc.Balance) > 0.01 {
            log.Printf("WARNING: Balance mismatch for account %d: DB=%.2f, Memory=%.2f",
                acc.ID, dbBalance, acc.Balance)
        }
    }
}
```

#### Deliverables
- [ ] Dual-write implementation complete
- [ ] Consistency monitoring active
- [ ] Performance impact measured (<5% latency increase)
- [ ] All discrepancies fixed

### Phase 4: Read Cutover (Week 4)

#### Objectives
- Gradually shift reads from memory to database
- Monitor latency and errors
- Implement Redis caching for hot data
- Complete cutover with feature flag

#### Implementation

```go
// backend/internal/core/read_strategy.go
package core

import (
    "database/sql"
    "context"
    "github.com/go-redis/redis/v8"
)

type ReadStrategy interface {
    GetAccount(id int64) (*Account, error)
    GetPositions(accountID int64) ([]*Position, error)
}

// MemoryReadStrategy reads from in-memory
type MemoryReadStrategy struct {
    engine *Engine
}

func (s *MemoryReadStrategy) GetAccount(id int64) (*Account, error) {
    acc, ok := s.engine.GetAccount(id)
    if !ok {
        return nil, ErrAccountNotFound
    }
    return acc, nil
}

// DatabaseReadStrategy reads from PostgreSQL with Redis cache
type DatabaseReadStrategy struct {
    db    *sql.DB
    redis *redis.Client
    ctx   context.Context
}

func (s *DatabaseReadStrategy) GetAccount(id int64) (*Account, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("account:%d", id)
    cached, err := s.redis.Get(s.ctx, cacheKey).Result()
    if err == nil {
        var acc Account
        json.Unmarshal([]byte(cached), &acc)
        return &acc, nil
    }

    // Cache miss - query database
    var acc Account
    err = s.db.QueryRow(`
        SELECT id, account_number, user_id, account_type, currency,
               balance, leverage, margin_mode, status
        FROM accounts WHERE id = $1
    `, id).Scan(
        &acc.ID, &acc.AccountNumber, &acc.UserID, &acc.AccountType,
        &acc.Currency, &acc.Balance, &acc.Leverage, &acc.MarginMode,
        &acc.Status,
    )
    if err != nil {
        return nil, err
    }

    // Cache result
    data, _ := json.Marshal(acc)
    s.redis.Set(s.ctx, cacheKey, data, 60*time.Second)

    return &acc, nil
}

// HybridReadStrategy with feature flag
type HybridReadStrategy struct {
    memStrategy *MemoryReadStrategy
    dbStrategy  *DatabaseReadStrategy
    dbReadPercent int // 0-100
}

func (s *HybridReadStrategy) GetAccount(id int64) (*Account, error) {
    // Gradually increase DB read percentage
    if rand.Intn(100) < s.dbReadPercent {
        return s.dbStrategy.GetAccount(id)
    }
    return s.memStrategy.GetAccount(id)
}
```

#### Redis Caching

```go
// backend/internal/cache/redis_cache.go
package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
)

type RedisCache struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisCache(addr string) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "",
        DB:       0,
        PoolSize: 100,
    })

    return &RedisCache{
        client: rdb,
        ctx:    context.Background(),
    }
}

func (c *RedisCache) SetAccount(acc *Account, ttl time.Duration) error {
    data, err := json.Marshal(acc)
    if err != nil {
        return err
    }

    key := fmt.Sprintf("account:%d", acc.ID)
    return c.client.Set(c.ctx, key, data, ttl).Err()
}

func (c *RedisCache) GetAccount(id int64) (*Account, error) {
    key := fmt.Sprintf("account:%d", id)
    data, err := c.client.Get(c.ctx, key).Result()
    if err != nil {
        return nil, err
    }

    var acc Account
    if err := json.Unmarshal([]byte(data), &acc); err != nil {
        return nil, err
    }

    return &acc, nil
}

func (c *RedisCache) InvalidateAccount(id int64) error {
    key := fmt.Sprintf("account:%d", id)
    return c.client.Del(c.ctx, key).Err()
}

func (c *RedisCache) SetLatestPrice(symbol string, bid, ask float64, ttl time.Duration) error {
    data := map[string]float64{"bid": bid, "ask": ask}
    jsonData, _ := json.Marshal(data)

    key := fmt.Sprintf("price:%s", symbol)
    return c.client.Set(c.ctx, key, jsonData, ttl).Err()
}
```

#### Deliverables
- [ ] Read strategy implemented
- [ ] Redis cache integrated
- [ ] Gradual rollout: 0% → 25% → 50% → 75% → 100%
- [ ] Performance validated (<10ms p99 latency)
- [ ] Error rate <0.01%

### Phase 5: Cleanup & Optimization (Week 5)

#### Objectives
- Remove in-memory storage code
- Optimize database indexes
- Tune PostgreSQL parameters
- Document lessons learned

#### Database Optimization

```sql
-- Analyze tables for statistics
ANALYZE accounts;
ANALYZE positions;
ANALYZE orders;
ANALYZE ledger_entries;

-- Identify slow queries
SELECT
    queryid,
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Add missing indexes based on query patterns
CREATE INDEX CONCURRENTLY idx_positions_account_symbol_status
    ON positions(account_id, symbol, status);

CREATE INDEX CONCURRENTLY idx_ledger_account_type_created
    ON ledger_entries(account_id, transaction_type, created_at DESC);

-- Vacuum and reindex
VACUUM ANALYZE;
REINDEX DATABASE trading_engine;
```

#### PostgreSQL Tuning

```ini
# /etc/postgresql/15/main/postgresql.conf

# Memory Settings
shared_buffers = 16GB                   # 25% of RAM
effective_cache_size = 48GB             # 75% of RAM
maintenance_work_mem = 2GB
work_mem = 64MB

# Checkpoint Settings
checkpoint_completion_target = 0.9
wal_buffers = 16MB
max_wal_size = 4GB
min_wal_size = 1GB

# Planner Settings
random_page_cost = 1.1                  # For SSD
effective_io_concurrency = 200          # For SSD

# Connection Settings
max_connections = 500
superuser_reserved_connections = 10

# Logging
log_min_duration_statement = 1000       # Log queries >1s
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on

# Autovacuum
autovacuum = on
autovacuum_max_workers = 4
autovacuum_naptime = 30s
```

#### Deliverables
- [ ] In-memory code removed
- [ ] Database optimized
- [ ] Performance benchmarks passed
- [ ] Documentation complete
- [ ] Production deployment plan ready

## Rollback Plan

### Immediate Rollback (<24 hours)

If critical issues are detected:

1. **Stop writes to database**
```go
dualWriteEngine.enabled = false
```

2. **Switch reads back to memory**
```go
hybridReadStrategy.dbReadPercent = 0
```

3. **Investigate issue**
- Check error logs
- Compare data consistency
- Identify root cause

4. **Fix and retry**
- Deploy fix
- Gradually increase db_read_percent
- Monitor closely

### Complete Rollback (>24 hours)

If fundamental issues discovered:

1. **Stop database writes**
2. **Restore from backup**
3. **Re-sync in-memory state**
4. **Postpone migration**
5. **Conduct post-mortem**

## Success Criteria

- [ ] Zero data loss
- [ ] <5% performance degradation
- [ ] <0.01% error rate
- [ ] All balances reconcile
- [ ] All positions match
- [ ] Monitoring dashboards green

## Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Data loss | Low | Critical | Frequent backups, dual-write validation |
| Performance degradation | Medium | High | Gradual rollout, caching, optimization |
| Inconsistency | Medium | High | Continuous monitoring, validation checks |
| Downtime | Low | Critical | Zero-downtime approach, rollback plan |
| Bug in migration script | Medium | Medium | Extensive testing, data validation |

## Timeline

| Week | Phase | Key Activities | Go/No-Go Decision |
|------|-------|----------------|-------------------|
| 1 | Infrastructure | Setup databases, monitoring | ✓ All systems operational |
| 2 | Migration | ETL data, validate | ✓ Data integrity verified |
| 3 | Dual-Write | Implement, monitor | ✓ Consistency validated |
| 4 | Read Cutover | Gradual rollout | ✓ Performance acceptable |
| 5 | Cleanup | Optimize, document | ✓ Production ready |

---

**Status**: Ready for Phase 1
**Next Review**: 2026-01-25
**Responsible**: Database Team
