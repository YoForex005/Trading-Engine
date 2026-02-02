# Workspace Persistence Migration - Deployment Checklist

## Migration Information

- **Migration Number:** 009
- **Migration Name:** create_workspace_tables
- **File Size:** 544 lines
- **Components:**
  - 6 Tables
  - 29 Indexes
  - 5 Triggers
  - 5 Functions
  - 1 Materialized View

## Pre-Deployment Checklist

### Environment Preparation

- [ ] **PostgreSQL Version Check**
  ```bash
  psql --version
  # Required: PostgreSQL 12.0 or higher
  ```

- [ ] **Database Connection Test**
  ```bash
  psql -d trading_engine -c "SELECT version();"
  ```

- [ ] **Current Schema Backup**
  ```bash
  pg_dump -d trading_engine -F c -f backup_before_migration_009_$(date +%Y%m%d_%H%M%S).dump
  ```

- [ ] **Check Existing Tables**
  ```bash
  psql -d trading_engine -c "SELECT tablename FROM pg_tables WHERE tablename LIKE 'workspace%';"
  # Should return 0 rows (or existing workspace tables if upgrading)
  ```

- [ ] **Verify Users Table Exists**
  ```bash
  psql -d trading_engine -c "\d users"
  # Must have 'id' column of type BIGINT or BIGSERIAL
  ```

- [ ] **Check Disk Space**
  ```bash
  df -h | grep postgres
  # Ensure at least 500MB free space
  ```

- [ ] **Check Migration Status**
  ```bash
  cd backend
  go run ./cmd/migrate status
  ```

### Documentation Review

- [ ] Read `README_WORKSPACE_SCHEMA.md`
- [ ] Review `WORKSPACE_SCHEMA_QUICK_REFERENCE.md`
- [ ] Review `WORKSPACE_MIGRATION_SUMMARY.md`
- [ ] Understand rollback procedure

### Testing Environment

- [ ] Migration tested in development environment
- [ ] Validation script passed
- [ ] Test data script executed successfully
- [ ] Sample queries verified

## Deployment Steps

### Step 1: Final Backup

```bash
# Full database backup
pg_dump -d trading_engine -F c -f backup_full_$(date +%Y%m%d_%H%M%S).dump

# Verify backup
pg_restore --list backup_full_*.dump | head -20
```

- [ ] Full backup completed
- [ ] Backup verified and stored securely
- [ ] Backup location documented: ________________

### Step 2: Run Migration

**Option A: Using Migration Tool (Recommended)**

```bash
cd backend
go run ./cmd/migrate up

# Check output for:
# [Migrator] Applying migration 9 (create_workspace_tables)...
# [Migrator] Migration 9 (create_workspace_tables) applied successfully
```

- [ ] Migration command executed
- [ ] No error messages displayed
- [ ] Success message confirmed

**Option B: Direct SQL (If migration tool unavailable)**

```bash
# Start transaction
psql -d trading_engine -c "BEGIN;"

# Apply migration
psql -d trading_engine -f backend/db/migrations/009_create_workspace_tables.go

# Check for errors, then commit
psql -d trading_engine -c "COMMIT;"
```

- [ ] Transaction started
- [ ] Migration executed
- [ ] No errors occurred
- [ ] Transaction committed

### Step 3: Validate Migration

```bash
# Run validation script
psql -d trading_engine -f backend/db/migrations/validate_workspace_schema.sql

# Expected output:
# ✓ PASS: All 6 tables exist
# ✓ PASS: workspaces has X columns
# ✓ PASS: Found X indexes
# ✓ PASS: Found 5 triggers
# ✓ PASS: Found 5 functions
```

- [ ] Validation script executed
- [ ] All table checks passed
- [ ] All index checks passed
- [ ] All trigger checks passed
- [ ] All function checks passed
- [ ] Materialized view exists

### Step 4: Verify Data Structures

```bash
# Check tables
psql -d trading_engine -c "\dt workspace*"

# Check indexes
psql -d trading_engine -c "\di workspace*"

# Check triggers
psql -d trading_engine -c "
SELECT tgname, tgrelid::regclass, proname
FROM pg_trigger
JOIN pg_proc ON tgfoid = oid
WHERE tgrelid::regclass::text LIKE 'workspace%'
AND tgname NOT LIKE 'RI_%';"

# Check functions
psql -d trading_engine -c "\df *workspace*"

# Check materialized views
psql -d trading_engine -c "\dm workspace*"
```

- [ ] All 6 tables visible
- [ ] All 29+ indexes visible
- [ ] All 5 triggers visible
- [ ] All 5 functions visible
- [ ] Materialized view visible

### Step 5: Test Basic Operations

```bash
# Test workspace creation
psql -d trading_engine -c "
INSERT INTO workspaces (user_id, name, description, is_default, settings)
VALUES (1, 'Test Workspace', 'Deployment test', TRUE, '{\"theme\": \"dark\"}'::jsonb)
RETURNING id;"

# Verify insertion
psql -d trading_engine -c "SELECT * FROM workspaces WHERE name = 'Test Workspace';"

# Clean up test data
psql -d trading_engine -c "DELETE FROM workspaces WHERE name = 'Test Workspace';"
```

- [ ] Test insert successful
- [ ] Test select successful
- [ ] Test delete successful
- [ ] No constraint violations

### Step 6: Performance Check

```bash
# Analyze tables
psql -d trading_engine -c "ANALYZE workspaces;"
psql -d trading_engine -c "ANALYZE workspace_charts;"
psql -d trading_engine -c "ANALYZE workspace_templates;"

# Check table sizes
psql -d trading_engine -c "
SELECT
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE tablename LIKE 'workspace%'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

- [ ] Tables analyzed
- [ ] Table sizes reasonable (should be minimal after fresh migration)

### Step 7: Update Migration Status

```bash
# Check migration registry
psql -d trading_engine -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# Should show version 9 with name 'create_workspace_tables'
```

- [ ] Migration 9 recorded in schema_migrations
- [ ] Timestamp correct
- [ ] Migration name correct

## Post-Deployment Checklist

### Immediate Verification (Within 1 hour)

- [ ] Application starts without errors
- [ ] No new errors in application logs
- [ ] No new errors in PostgreSQL logs
- [ ] API endpoints respond correctly
- [ ] UI can load/save workspaces

### Short-Term Monitoring (24 hours)

- [ ] Monitor database CPU usage
- [ ] Monitor database memory usage
- [ ] Monitor query performance
- [ ] Check for slow queries
- [ ] Monitor table growth

### Tasks to Schedule

- [ ] **Daily:** Run expired snapshot cleanup
  ```sql
  SELECT clean_expired_snapshots();
  ```

- [ ] **Daily:** Run expired share cleanup
  ```sql
  SELECT clean_expired_shares();
  ```

- [ ] **Weekly:** Refresh materialized view
  ```sql
  REFRESH MATERIALIZED VIEW CONCURRENTLY workspace_statistics;
  ```

- [ ] **Weekly:** Check table sizes
  ```bash
  psql -d trading_engine -f backend/db/migrations/validate_workspace_schema.sql
  ```

- [ ] **Monthly:** VACUUM ANALYZE
  ```sql
  VACUUM ANALYZE workspaces;
  VACUUM ANALYZE workspace_charts;
  VACUUM ANALYZE workspace_templates;
  VACUUM ANALYZE workspace_snapshots;
  ```

## Rollback Procedure (If Needed)

### When to Rollback

Rollback if:
- Migration fails halfway
- Critical bugs discovered
- Performance issues
- Data corruption detected

### Rollback Steps

1. **Stop Application**
   ```bash
   # Stop your application servers
   systemctl stop trading-engine
   ```

2. **Run Rollback**
   ```bash
   cd backend
   go run ./cmd/migrate down
   ```

3. **Verify Rollback**
   ```bash
   psql -d trading_engine -c "\dt workspace*"
   # Should return no rows
   ```

4. **Restore from Backup (if needed)**
   ```bash
   pg_restore -d trading_engine backup_before_migration_009_*.dump
   ```

5. **Verify Application**
   ```bash
   # Restart application
   systemctl start trading-engine

   # Check logs
   tail -f /var/log/trading-engine/app.log
   ```

- [ ] Application stopped
- [ ] Rollback executed successfully
- [ ] Tables removed
- [ ] Backup restored (if needed)
- [ ] Application restarted
- [ ] Application functioning normally

## Success Criteria

Migration is successful when ALL of the following are true:

### Database Level
- ✅ All 6 tables created
- ✅ All 29+ indexes created
- ✅ All 5 triggers active
- ✅ All 5 functions available
- ✅ Materialized view exists
- ✅ No errors in PostgreSQL logs

### Application Level
- ✅ Application starts successfully
- ✅ No migration-related errors
- ✅ API endpoints respond
- ✅ UI loads without errors

### Data Level
- ✅ Can create workspaces
- ✅ Can add charts to workspaces
- ✅ Can save print preferences
- ✅ Can create templates
- ✅ Can create snapshots
- ✅ Can share workspaces

### Performance Level
- ✅ Query response times acceptable (<100ms for standard queries)
- ✅ No significant increase in CPU usage
- ✅ No significant increase in memory usage
- ✅ No slow query warnings

## Troubleshooting

### Issue: "relation already exists"

**Cause:** Tables already exist from previous migration
**Solution:**
```bash
# Check existing tables
psql -d trading_engine -c "\dt workspace*"

# If old version exists, drop manually or run rollback first
go run ./cmd/migrate down
go run ./cmd/migrate up
```

### Issue: "foreign key constraint violation"

**Cause:** Users table doesn't exist or has wrong structure
**Solution:**
```bash
# Verify users table
psql -d trading_engine -c "\d users"

# If id column is wrong type, migration needs adjustment
```

### Issue: "out of memory"

**Cause:** Insufficient database memory for large JSONB operations
**Solution:**
```bash
# Increase work_mem temporarily
psql -d trading_engine -c "SET work_mem = '256MB';"

# Then retry migration
```

### Issue: Migration hangs

**Cause:** Lock contention with active connections
**Solution:**
```bash
# Check active connections
psql -d trading_engine -c "SELECT * FROM pg_stat_activity WHERE datname = 'trading_engine';"

# Terminate if needed
psql -d trading_engine -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'trading_engine' AND pid != pg_backend_pid();"

# Retry migration
```

## Contact Information

**Database Administrator:** ___________________________
**Phone:** ___________________________
**Email:** ___________________________

**Backup Location:** ___________________________
**Monitoring Dashboard:** ___________________________

## Sign-Off

### Pre-Deployment
- [ ] Reviewed by: _________________ Date: _________
- [ ] Approved by: _________________ Date: _________

### Post-Deployment
- [ ] Deployed by: _________________ Date: _________ Time: _________
- [ ] Verified by: _________________ Date: _________ Time: _________
- [ ] Status: ☐ Success  ☐ Failed  ☐ Rolled Back

### Notes
```
_____________________________________________________________
_____________________________________________________________
_____________________________________________________________
_____________________________________________________________
```

---

**Document Version:** 1.0
**Last Updated:** 2024-01-XX
**Migration:** 009_create_workspace_tables
