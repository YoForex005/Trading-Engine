# Database Migrations

This directory contains PostgreSQL database migrations for the Trading Engine.

## Overview

The migration system provides:
- **Versioned migrations** - Sequential, numbered migration files
- **Up/Down migrations** - Full rollback support
- **Atomic execution** - Each migration runs in a transaction
- **Dry-run mode** - Preview changes without applying
- **Migration tracking** - `schema_migrations` table tracks applied migrations
- **Production-safe** - Idempotent operations, concurrent index creation

## Migration Files

| Version | File | Description |
|---------|------|-------------|
| 001 | `001_init_schema.sql` | Core tables (users, accounts, orders, positions, trades, instruments) |
| 002 | `002_add_indexes.sql` | Performance indexes (B-tree, GIN, BRIN, text search) |
| 003 | `003_add_audit_tables.sql` | Audit trail tables with automatic triggers |
| 004 | `004_add_lp_tables.sql` | Liquidity provider and routing tables |
| 005 | `005_add_risk_tables.sql` | Risk management and monitoring tables |

## Running Migrations

### Using the Migration Tool

```bash
# Initialize migration tracking table
./bin/migrate -cmd init

# Run all pending migrations
./bin/migrate -cmd up

# Rollback last migration
./bin/migrate -cmd down

# Check migration status
./bin/migrate -cmd status

# Dry run (preview changes)
./bin/migrate -cmd up -dry-run

# Verbose mode
./bin/migrate -cmd up -verbose
```

### Building the Migration Tool

```bash
# Build the migration tool
go build -o bin/migrate ./cmd/migrate

# Or run directly
go run ./cmd/migrate -cmd status
```

### Environment Variables

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=trading_engine
export DB_SSLMODE=disable
```

### Custom Connection String

```bash
./bin/migrate -cmd up -conn "postgresql://user:pass@host:5432/dbname?sslmode=disable"
```

## Database Setup Scripts

### Setup Development Database

```bash
# Setup local PostgreSQL database
./scripts/db/setup-dev-db.sh
```

This script:
1. Creates `trading_engine` and `trading_engine_test` databases
2. Enables required PostgreSQL extensions
3. Runs all migrations
4. Displays connection details

### Seed Test Data

```bash
# Populate database with test data
./scripts/db/seed-test-data.sh
```

Seeds:
- 4 test users (admin, trader1, trader2, demo)
- 15 instruments (Forex, Crypto, Commodities, Indices)
- 4 test accounts
- 4 liquidity providers
- Risk limits and circuit breakers

### Backup Database

```bash
# Create backup with timestamp
./scripts/db/backup.sh

# Create backup with custom name
./scripts/db/backup.sh my_backup_name
```

Backups are stored in `./backups/` directory as compressed `.sql.gz` files.

### Restore Database

```bash
# List available backups
./scripts/db/restore.sh

# Restore specific backup
./scripts/db/restore.sh backup_20260118_120000
```

The restore script:
1. Creates a pre-restore backup
2. Drops and recreates the database
3. Restores from backup
4. Optionally runs migrations
5. Verifies restoration

## Migration File Format

Each migration file follows this structure:

```sql
-- Migration: 001_init_schema
-- Description: Initial database schema
-- Author: Your Name
-- Date: 2026-01-18

-- ============================================================================
-- UP Migration
-- ============================================================================

CREATE TABLE IF NOT EXISTS my_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- DOWN Migration
-- ============================================================================

-- To rollback, uncomment and run the following:
/*
DROP TABLE IF EXISTS my_table CASCADE;
*/
```

## Best Practices

### Idempotency

All migrations should be idempotent:

```sql
-- Good
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);

-- Bad
CREATE TABLE users (...);  -- Fails if table exists
```

### Concurrent Index Creation

Use `CONCURRENTLY` for production safety:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_name ON table(column);
```

### Transactions

Each migration runs in a transaction automatically. Avoid manual transactions:

```sql
-- Good (automatic transaction)
CREATE TABLE users (...);
ALTER TABLE users ADD COLUMN email VARCHAR(255);

-- Bad (nested transaction fails)
BEGIN;
CREATE TABLE users (...);
COMMIT;
```

### Large Data Migrations

For large data migrations, use batching:

```sql
-- Process in batches
DO $$
DECLARE
    batch_size INT := 10000;
    offset_val INT := 0;
BEGIN
    LOOP
        UPDATE large_table SET new_column = old_column
        WHERE id IN (
            SELECT id FROM large_table
            ORDER BY id LIMIT batch_size OFFSET offset_val
        );

        EXIT WHEN NOT FOUND;
        offset_val := offset_val + batch_size;
        COMMIT;
    END LOOP;
END $$;
```

### Rollback Strategy

Always provide DOWN migrations:

```sql
-- UP
CREATE TABLE orders (...);
CREATE INDEX idx_orders_symbol ON orders(symbol);

-- DOWN
DROP INDEX IF EXISTS idx_orders_symbol;
DROP TABLE IF EXISTS orders CASCADE;
```

## Schema Migrations Table

The `schema_migrations` table tracks applied migrations:

```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INTEGER,
    checksum VARCHAR(64)
);
```

Query applied migrations:

```sql
SELECT * FROM schema_migrations ORDER BY version DESC;
```

## Troubleshooting

### Migration Failed Mid-Way

Migrations are atomic (run in transactions). If a migration fails, it's automatically rolled back.

```bash
# Check status
./bin/migrate -cmd status

# Fix the migration file
# Re-run
./bin/migrate -cmd up
```

### Schema Out of Sync

```bash
# Check what migrations are pending
./bin/migrate -cmd status

# Apply pending migrations
./bin/migrate -cmd up
```

### Rollback to Specific Version

```bash
# Rollback one migration at a time
./bin/migrate -cmd down
./bin/migrate -cmd down
./bin/migrate -cmd down

# Verify
./bin/migrate -cmd status
```

### Restore from Backup

```bash
# List available backups
ls -lht backups/

# Restore
./scripts/db/restore.sh backup_name
```

## Production Deployment

### Pre-Deployment

1. **Backup database**
   ```bash
   ./scripts/db/backup.sh pre_deploy_$(date +%Y%m%d_%H%M%S)
   ```

2. **Dry run migrations**
   ```bash
   ./bin/migrate -cmd up -dry-run
   ```

3. **Test on staging**
   ```bash
   DB_NAME=staging_db ./bin/migrate -cmd up
   ```

### Deployment

1. **Set maintenance mode** (optional)
2. **Run migrations**
   ```bash
   ./bin/migrate -cmd up
   ```
3. **Verify**
   ```bash
   ./bin/migrate -cmd status
   ```
4. **Exit maintenance mode**

### Rollback

If issues occur:

```bash
# Rollback migration
./bin/migrate -cmd down

# Or restore from backup
./scripts/db/restore.sh pre_deploy_backup
```

## Auto-Migration on Server Start

Add to server startup code:

```go
package main

import (
    "log"
    "github.com/epic1st/rtx/backend/database"
)

func main() {
    // Connect to database
    db, err := database.Connect(database.GetConnectionString())
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Optional: Auto-migrate on startup
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

Set environment variable:

```bash
export AUTO_MIGRATE=true
```

## Contributing

When adding a new migration:

1. **Create migration file** with next version number
2. **Follow naming convention** `00X_description.sql`
3. **Include UP and DOWN** migrations
4. **Test locally**
   ```bash
   ./bin/migrate -cmd up
   ./bin/migrate -cmd down
   ```
5. **Test idempotency** (run twice)
6. **Document** in this README

## Support

For issues or questions:
- Check migration status: `./bin/migrate -cmd status`
- View logs with verbose mode: `./bin/migrate -cmd up -verbose`
- Create backup before experimenting: `./scripts/db/backup.sh`
