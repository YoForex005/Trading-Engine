# Database Setup Guide

Complete guide for setting up and managing the PostgreSQL database for the Trading Engine.

## Quick Start

```bash
# 1. Setup database and run migrations
./scripts/db/setup-dev-db.sh

# 2. Seed test data
./scripts/db/seed-test-data.sh

# 3. Verify setup
./bin/migrate -cmd status
```

## Architecture

### Database Schema

The database consists of 5 migration sets:

1. **001_init_schema.sql** - Core Trading Tables
   - Users & Authentication (users, user_roles)
   - Accounts & Balances (accounts, account_transactions)
   - Orders (orders, order_fills)
   - Positions & Trades
   - Instruments & Market Data

2. **002_add_indexes.sql** - Performance Optimization
   - 80+ strategic indexes (B-tree, GIN, BRIN)
   - Text search indexes (pg_trgm)
   - JSONB indexes for metadata
   - Time-series optimizations

3. **003_add_audit_tables.sql** - Compliance & Audit
   - Generic audit log with automatic triggers
   - User activity tracking
   - Order, position, account audit trails
   - Risk events and compliance monitoring
   - System events and API access logs

4. **004_add_lp_tables.sql** - Liquidity Providers
   - LP configuration and instruments
   - Real-time quotes and pricing
   - Order routing decisions
   - Performance metrics
   - Client profiling for smart routing
   - SOR (Smart Order Routing) rules

5. **005_add_risk_tables.sql** - Risk Management
   - Risk limits and breaches
   - Margin requirements and calls
   - Position and portfolio risk metrics
   - Circuit breakers
   - Stress testing scenarios

### Key Features

- **Atomic Migrations** - Each migration runs in a transaction
- **Idempotent Operations** - Safe to run multiple times
- **Rollback Support** - Every migration has UP and DOWN
- **Concurrent Indexes** - Non-blocking index creation
- **Audit Trails** - Automatic audit logging via triggers
- **Time-Series Optimization** - BRIN indexes for large tables
- **JSONB Support** - Flexible metadata storage

## Prerequisites

### PostgreSQL Installation

**macOS:**
```bash
brew install postgresql@16
brew services start postgresql@16
```

**Ubuntu:**
```bash
sudo apt update
sudo apt install postgresql-16 postgresql-client-16
sudo systemctl start postgresql
```

**Docker:**
```bash
docker run --name trading-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=trading_engine \
  -p 5432:5432 \
  -d postgres:16-alpine
```

### Required Extensions

The setup script automatically enables:
- `uuid-ossp` - UUID generation
- `pg_trgm` - Text search
- `btree_gist` - Advanced indexing

## Environment Configuration

Create a `.env` file in the backend directory:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=trading_engine
DB_SSLMODE=disable

# Optional: Auto-migrate on server start
AUTO_MIGRATE=false

# Optional: Backup directory
BACKUP_DIR=./backups
```

Load environment:
```bash
source .env
```

## Migration Management

### Migration Tool

Build the migration tool:
```bash
go build -o bin/migrate ./cmd/migrate
```

### Commands

```bash
# Initialize migration tracking
./bin/migrate -cmd init

# Run all pending migrations
./bin/migrate -cmd up

# Rollback last migration
./bin/migrate -cmd down

# Check migration status
./bin/migrate -cmd status

# Dry run (preview changes)
./bin/migrate -cmd up -dry-run

# Verbose output
./bin/migrate -cmd up -verbose
```

### Custom Connection

```bash
./bin/migrate -cmd up -conn "postgresql://user:pass@host:5432/dbname"
```

## Development Workflow

### 1. Initial Setup

```bash
# Setup database
./scripts/db/setup-dev-db.sh

# This will:
# - Create trading_engine and trading_engine_test databases
# - Enable required extensions
# - Run all migrations
# - Display connection details
```

### 2. Seed Test Data

```bash
./scripts/db/seed-test-data.sh

# Seeds:
# - 4 users (admin, trader1, trader2, demo)
# - 15 instruments (EURUSD, BTCUSD, XAUUSD, etc.)
# - 4 accounts (live, demo, margin)
# - 4 liquidity providers
# - Risk limits and circuit breakers
```

### 3. Verify Setup

```bash
# Check migration status
./bin/migrate -cmd status

# Connect to database
psql -h localhost -U postgres -d trading_engine

# List tables
\dt

# Check table counts
SELECT
    schemaname,
    tablename,
    (xpath('/row/cnt/text()', xml_count))[1]::text::int as row_count
FROM (
    SELECT
        schemaname,
        tablename,
        query_to_xml(format('select count(*) as cnt from %I.%I', schemaname, tablename), false, true, '') as xml_count
    FROM pg_tables
    WHERE schemaname = 'public'
) t
ORDER BY row_count DESC;
```

## Backup & Restore

### Create Backup

```bash
# Automatic timestamp
./scripts/db/backup.sh

# Custom name
./scripts/db/backup.sh production_backup_v1
```

Backups are stored in `./backups/` as compressed `.sql.gz` files with metadata.

### Restore Backup

```bash
# List available backups
./scripts/db/restore.sh

# Restore specific backup
./scripts/db/restore.sh backup_20260118_120000
```

The restore process:
1. Creates a pre-restore backup automatically
2. Terminates active connections
3. Drops and recreates database
4. Restores from backup
5. Optionally runs migrations
6. Verifies restoration

### Automated Backup Schedule

Add to crontab:
```bash
# Daily backup at 2 AM
0 2 * * * /path/to/trading-engine/backend/scripts/db/backup.sh daily_$(date +\%Y\%m\%d)

# Weekly backup on Sunday at 3 AM
0 3 * * 0 /path/to/trading-engine/backend/scripts/db/backup.sh weekly_$(date +\%Y\%m\%d)
```

## Production Deployment

### Pre-Deployment Checklist

- [ ] Create backup
- [ ] Test migrations on staging
- [ ] Review migration changes
- [ ] Plan rollback strategy
- [ ] Schedule maintenance window

### Deployment Steps

```bash
# 1. Create pre-deployment backup
./scripts/db/backup.sh pre_deploy_$(date +%Y%m%d_%H%M%S)

# 2. Test migrations (dry-run)
./bin/migrate -cmd up -dry-run

# 3. Set maintenance mode (optional)
# Update load balancer / set app to maintenance

# 4. Run migrations
./bin/migrate -cmd up

# 5. Verify
./bin/migrate -cmd status
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM users;"

# 6. Exit maintenance mode
```

### Rollback Procedure

If issues occur:

```bash
# Option 1: Rollback migration
./bin/migrate -cmd down

# Option 2: Restore from backup
./scripts/db/restore.sh pre_deploy_20260118_120000

# Verify
./bin/migrate -cmd status
```

## Advanced Usage

### Auto-Migration on Server Start

Add to `cmd/server/main.go`:

```go
import (
    "os"
    "github.com/epic1st/rtx/backend/database"
)

func main() {
    // Connect to database
    db, err := database.Connect(database.GetConnectionString())
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Auto-migrate if enabled
    if os.Getenv("AUTO_MIGRATE") == "true" {
        migrator := database.NewMigrator(db)
        if err := migrator.Initialize(); err != nil {
            log.Fatal("Migration init failed:", err)
        }
        if err := migrator.Up(); err != nil {
            log.Fatal("Migration failed:", err)
        }
        log.Println("✓ Migrations completed")
    }

    // Start server...
}
```

Enable with:
```bash
export AUTO_MIGRATE=true
```

### Creating New Migrations

1. **Create migration file** with next version number:
   ```bash
   touch migrations/006_add_my_feature.sql
   ```

2. **Follow template:**
   ```sql
   -- Migration: 006_add_my_feature
   -- Description: Add new feature
   -- Author: Your Name
   -- Date: 2026-01-18

   -- UP Migration
   CREATE TABLE IF NOT EXISTS my_table (...);

   -- DOWN Migration
   /*
   DROP TABLE IF EXISTS my_table CASCADE;
   */
   ```

3. **Test migration:**
   ```bash
   ./bin/migrate -cmd up
   ./bin/migrate -cmd down
   ./bin/migrate -cmd up  # Test idempotency
   ```

### Database Maintenance

```bash
# Vacuum and analyze
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "VACUUM ANALYZE;"

# Reindex
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "REINDEX DATABASE trading_engine;"

# Check database size
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT
    pg_size_pretty(pg_database_size('trading_engine')) as db_size,
    pg_size_pretty(pg_total_relation_size('orders')) as orders_size,
    pg_size_pretty(pg_total_relation_size('trades')) as trades_size;
"

# Check index usage
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC
LIMIT 20;
"
```

## Monitoring

### Connection Pooling

The migration system configures connection pooling:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Query Performance

```sql
-- Enable query logging (postgresql.conf)
log_statement = 'all'
log_duration = on
log_min_duration_statement = 1000  -- Log queries > 1s

-- Check slow queries
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
WHERE mean_time > 1000
ORDER BY mean_time DESC
LIMIT 10;
```

## Troubleshooting

### Connection Issues

```bash
# Test connection
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME

# Check PostgreSQL is running
# macOS
brew services list | grep postgresql

# Ubuntu
sudo systemctl status postgresql

# Docker
docker ps | grep postgres
```

### Migration Failures

```bash
# Check status
./bin/migrate -cmd status

# View detailed logs
./bin/migrate -cmd up -verbose

# Rollback and retry
./bin/migrate -cmd down
# Fix migration file
./bin/migrate -cmd up
```

### Permission Issues

```sql
-- Grant all permissions
GRANT ALL PRIVILEGES ON DATABASE trading_engine TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;
```

### Disk Space

```bash
# Check disk space
df -h

# Check database size
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT pg_size_pretty(pg_database_size('trading_engine'));
"

# Clean up old backups
find ./backups -name "*.sql.gz" -mtime +30 -delete
```

## Security Best Practices

1. **Strong passwords** - Use complex passwords for database users
2. **SSL/TLS** - Enable SSL for production: `DB_SSLMODE=require`
3. **Network isolation** - Use firewall rules to restrict access
4. **Regular backups** - Automated daily backups
5. **Audit logging** - Monitor database access via audit tables
6. **Encryption at rest** - Enable PostgreSQL encryption
7. **Role-based access** - Create separate users for different services

## Performance Tuning

### PostgreSQL Configuration

Edit `postgresql.conf`:

```conf
# Memory settings
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 16MB
maintenance_work_mem = 128MB

# Write-ahead log
wal_buffers = 16MB
checkpoint_completion_target = 0.9
max_wal_size = 2GB

# Query planner
random_page_cost = 1.1  # For SSD
effective_io_concurrency = 200

# Autovacuum
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 1min
```

Restart PostgreSQL after changes.

## Support

For issues or questions:
1. Check migration status: `./bin/migrate -cmd status`
2. Review logs with verbose mode
3. Consult the [migrations README](../migrations/README.md)
4. Create a backup before experimenting

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Migration System README](../migrations/README.md)
- [Database Schema](../database/migrate.go)
