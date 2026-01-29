# Database Migrations

This directory contains all database migrations for the RTX Trading Engine.

## Overview

The migration system provides versioned database schema management with:
- **Up/Down migrations** - Forward and backward migration support
- **Transaction safety** - All migrations run in transactions
- **Version tracking** - Automatic tracking of applied migrations
- **PostgreSQL support** - Optimized for PostgreSQL with TimescaleDB

## Migration Tool

### Installation

```bash
# Build the migration tool
go build -o bin/migrate ./cmd/migrate

# Or use the helper script
./scripts/db/migrate.sh rebuild
```

### Usage

```bash
# Initialize migrations table
./bin/migrate -init

# Run all pending migrations
./bin/migrate -up

# Rollback last migration
./bin/migrate -down

# Show migration status
./bin/migrate -status

# Migrate to specific version
./bin/migrate -version=1
```

### Helper Script

The `scripts/db/migrate.sh` helper provides convenient commands:

```bash
# Initialize
./scripts/db/migrate.sh init

# Run all migrations
./scripts/db/migrate.sh up

# Rollback one migration
./scripts/db/migrate.sh down

# Show status
./scripts/db/migrate.sh status

# Fresh start (drop all and re-run)
./scripts/db/migrate.sh fresh
```

## Database Schema

### Core Tables

**Users & Accounts:**
- `users` - User authentication and profile data
- `accounts` - Trading accounts with balance, equity, margin tracking
- `trading_groups` - Groups with custom settings (markup, commission, leverage)

**Trading:**
- `orders` - Order lifecycle tracking
- `positions` - Open and closed positions
- `trades` - Order fills and executions
- `order_history` - Analytics data for order patterns

**Risk Management:**
- `risk_alerts` - Real-time risk alerts
- `liquidation_events` - Historical liquidation records
- `symbol_correlations` - Inter-symbol correlation matrix

**Admin:**
- `admin_users` - Admin authentication with RBAC
- `admin_audit_log` - Complete audit trail (JSONB details)
- `group_symbol_settings` - Per-group, per-symbol configuration

**FIX API:**
- `fix_credentials` - Encrypted FIX credentials
- `fix_sessions` - Active FIX session tracking

### Indexes

All tables have appropriate indexes for:
- Primary key lookups
- Foreign key joins
- Common query patterns (account_id, symbol, created_at, status)
- Audit log queries (admin_id, action, resource_type, created_at)

## Creating New Migrations

### 1. Create Migration File

Create a new file in `db/migrations/`:

```go
// 002_add_feature.go
package migrations

import "database/sql"

func init() {
    RegisterMigration(&Migration{
        Version: 2,
        Name:    "add_feature",
        Up:      addFeatureUp,
        Down:    addFeatureDown,
    })
}

func addFeatureUp(tx *sql.Tx) error {
    _, err := tx.Exec(`
        ALTER TABLE users ADD COLUMN new_field VARCHAR(255);
        CREATE INDEX idx_users_new_field ON users(new_field);
    `)
    return err
}

func addFeatureDown(tx *sql.Tx) error {
    _, err := tx.Exec(`
        DROP INDEX IF EXISTS idx_users_new_field;
        ALTER TABLE users DROP COLUMN new_field;
    `)
    return err
}
```

### 2. Version Numbering

Use sequential integers for version numbers:
- 001 - Initial schema
- 002 - Add feature X
- 003 - Modify table Y
- etc.

### 3. Best Practices

✅ **DO:**
- Keep migrations small and focused
- Always provide down migrations
- Use transactions (automatic)
- Add indexes for new columns
- Use IF EXISTS/IF NOT EXISTS for safety
- Test both up and down migrations

❌ **DON'T:**
- Mix schema changes with data migrations in one file
- Modify existing migrations after deployment
- Use DROP without IF EXISTS
- Forget to add indexes
- Make migrations that can't be rolled back

## Environment Configuration

Set these environment variables or use `.env` file:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSL_MODE=disable
```

## Production Deployment

### First Time Setup

```bash
# 1. Create database
createdb trading_engine

# 2. Run migrations
./bin/migrate -up
```

### Updating Existing Database

```bash
# 1. Check current status
./bin/migrate -status

# 2. Run pending migrations
./bin/migrate -up
```

### Rollback Procedure

```bash
# 1. Rollback last migration
./bin/migrate -down

# 2. Verify status
./bin/migrate -status
```

## Testing Migrations

```bash
# Test up migration
./bin/migrate -up

# Verify schema
psql trading_engine -c "\dt"

# Test down migration
./bin/migrate -down

# Verify rollback worked
psql trading_engine -c "\dt"

# Re-apply
./bin/migrate -up
```

## Troubleshooting

### Migration Failed

If a migration fails mid-execution:
1. The transaction will be rolled back automatically
2. Fix the migration SQL
3. Re-run `./bin/migrate -up`

### Manual Fix Required

If you need to manually fix the database:

```bash
# 1. Check current version
SELECT * FROM schema_migrations;

# 2. Manually apply fix
psql trading_engine -c "YOUR FIX SQL"

# 3. Update migration record if needed
# (Only in emergency situations)
```

### Reset Development Database

```bash
# Complete fresh start
./scripts/db/migrate.sh fresh
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run database migrations
  run: |
    go build -o bin/migrate ./cmd/migrate
    ./bin/migrate -up
  env:
    DB_HOST: localhost
    DB_NAME: trading_engine_test
    DB_USER: postgres
    DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
```

### Docker

```dockerfile
# In your Dockerfile
COPY db/migrations /app/db/migrations
COPY cmd/migrate /app/cmd/migrate
RUN go build -o /app/bin/migrate /app/cmd/migrate

# Run migrations on startup
CMD ["/app/bin/migrate", "-up"]
```

## Schema Versioning

The `schema_migrations` table tracks all applied migrations:

```sql
SELECT version, name, applied_at 
FROM schema_migrations 
ORDER BY version;
```

This allows the system to know exactly which migrations have been applied and which are pending.

## Support

For migration issues:
1. Check migration logs
2. Verify database connectivity
3. Review migration SQL syntax
4. Check for conflicting schema changes
5. Consult `PRODUCTION_STATUS.md` for known issues
