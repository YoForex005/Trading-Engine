# RTX Trading Engine - Disaster Recovery Scripts

This directory contains all automation scripts for backup, restore, failover, and monitoring operations.

---

## Directory Structure

```
scripts/dr/
├── backup-full.sh           # Full PostgreSQL backup with S3 upload
├── backup-incremental.sh    # WAL file sync to S3
├── archive-wal.sh           # PostgreSQL archive_command handler
├── restore-full.sh          # Database restore from backup
├── failover-database.sh     # Automatic database failover
├── health-check.sh          # System health monitoring
├── verify-backup.sh         # Backup integrity verification
├── test-restore.sh          # Weekly restore testing
├── generate-weekly-report.sh # DR metrics reporting
└── README.md                # This file
```

---

## Scripts Overview

### 1. backup-full.sh

**Purpose:** Create full database backup with encryption and S3 upload

**Usage:**
```bash
/opt/rtx/scripts/dr/backup-full.sh
```

**Schedule:** Daily at 02:00 UTC (via cron)

**What it does:**
1. Validates database connectivity
2. Creates pg_dump backup (custom format)
3. Generates SHA256 checksum
4. Uploads to S3 with KMS encryption
5. Verifies upload integrity
6. Records in `backup_history` table
7. Sends SNS notification

**Environment Variables:**
- `RTX_DB_NAME` - Database name (default: rtx)
- `RTX_DB_USER` - Database user (default: postgres)
- `RTX_DB_HOST` - Database host (default: localhost)
- `RTX_BACKUP_BUCKET` - S3 bucket (default: rtx-backups)
- `RTX_KMS_KEY` - KMS key for encryption
- `RTX_SNS_TOPIC` - SNS topic for alerts

**Outputs:**
- Backup file: `/var/backups/rtx/rtx-full-YYYYMMDD_HHMMSS.dump`
- Checksum: `/var/backups/rtx/rtx-full-YYYYMMDD_HHMMSS.dump.sha256`
- Metadata: `/var/backups/rtx/rtx-full-YYYYMMDD_HHMMSS.dump.meta`
- Log: `/var/log/rtx/backup-full-YYYYMMDD_HHMMSS.log`

**Exit Codes:**
- 0: Success
- 1: Failure (sends SNS alert)

---

### 2. backup-incremental.sh

**Purpose:** Sync WAL files to S3 for point-in-time recovery

**Usage:**
```bash
/opt/rtx/scripts/dr/backup-incremental.sh
```

**Schedule:** Every 6 hours (via cron)

**What it does:**
1. Syncs `/var/lib/postgresql/14/archive` to S3
2. Uploads with KMS encryption
3. Deletes local files older than 7 days

**Recovery Point Objective (RPO):** 6 hours

---

### 3. archive-wal.sh

**Purpose:** PostgreSQL archive_command to stream WAL files to S3

**Usage:** Called automatically by PostgreSQL
```bash
# In postgresql.conf:
archive_command = '/opt/rtx/scripts/dr/archive-wal.sh %p %f'
```

**What it does:**
1. Compresses WAL file with gzip
2. Immediately uploads to S3
3. Returns 0 to PostgreSQL on success

**Real-time Archiving:** Yes (WAL streamed as generated)

---

### 4. restore-full.sh

**Purpose:** Restore database from backup

**Usage:**
```bash
# Restore latest backup
/opt/rtx/scripts/dr/restore-full.sh latest

# Restore specific date
/opt/rtx/scripts/dr/restore-full.sh 20260118

# Verify-only mode (no changes)
/opt/rtx/scripts/dr/restore-full.sh latest --verify-only
```

**What it does:**
1. Downloads backup from S3
2. Verifies SHA256 checksum
3. Validates pg_restore can read file
4. **WARNING:** Drops and recreates database
5. Restores using parallel workers (4 jobs)
6. Reindexes and analyzes tables
7. Restarts RTX services
8. Runs health check

**Recovery Time Objective (RTO):** 15 minutes

**CRITICAL:** This script will DELETE all data in the database. Use `--verify-only` first!

---

### 5. failover-database.sh

**Purpose:** Promote standby to primary during disaster

**Usage:**
```bash
# Automatic failover (checks primary is down)
/opt/rtx/scripts/dr/failover-database.sh

# Force failover (skip checks)
/opt/rtx/scripts/dr/failover-database.sh --force
```

**What it does:**
1. Verifies primary database is down
2. Checks standby replication lag
3. Stops all RTX services
4. Promotes standby to primary
5. Updates DNS to new primary
6. Updates application config
7. Restarts services
8. Runs health checks

**Estimated Failover Time:** 9-15 minutes

**Prerequisites:**
- PostgreSQL streaming replication configured
- Standby server accessible via SSH
- DNS or load balancer for routing

---

### 6. health-check.sh

**Purpose:** Comprehensive system health monitoring

**Usage:**
```bash
/opt/rtx/scripts/dr/health-check.sh
```

**Schedule:** Every 60 seconds (via cron)

**What it checks:**
1. Database connectivity and query performance
2. API health endpoint (HTTP 200)
3. WebSocket server connectivity
4. Redis cache availability
5. Disk space usage (>85% alert)
6. Replication lag (>60s alert)
7. CPU load (>80% warn)
8. Memory usage (>90% warn)

**Outputs:**
- JSON metrics: `/var/lib/rtx/metrics/health.json`
- Log: `/var/log/rtx/health-check.log`
- CloudWatch metrics: `RTX/Health/FailedChecks`

**Alerting:**
- **Critical:** Database down, API down
- **Warning:** High lag, disk space, redis down

**Exit Codes:**
- 0: All checks passed
- 1: One or more checks failed

---

### 7. verify-backup.sh

**Purpose:** Verify backup integrity

**Usage:**
```bash
/opt/rtx/scripts/dr/verify-backup.sh
```

**Schedule:** Daily at 03:00 UTC (after full backup)

**What it does:**
1. Downloads latest backup from S3
2. Verifies SHA256 checksum
3. Tests pg_restore can read file
4. Validates table count matches production
5. Sends report via SNS

---

### 8. test-restore.sh

**Purpose:** Weekly restore drill to staging database

**Usage:**
```bash
/opt/rtx/scripts/dr/test-restore.sh
```

**Schedule:** Weekly on Sunday at 04:00 UTC

**What it does:**
1. Creates staging database `rtx_test`
2. Restores latest backup
3. Runs data integrity checks
4. Compares record counts with production
5. Drops staging database
6. Sends report

**Why:** Ensures backups are restorable (compliance requirement)

---

### 9. generate-weekly-report.sh

**Purpose:** Generate DR metrics report

**Usage:**
```bash
/opt/rtx/scripts/dr/generate-weekly-report.sh
```

**Schedule:** Every Monday at 09:00 UTC

**Report includes:**
- Backup success/failure rate
- Average backup duration
- Restore test results
- Replication lag trends
- Disk space trends
- Alert summary

**Output:** HTML email sent to ops team

---

## Environment Setup

All scripts expect these environment variables (set in `/opt/rtx/.env`):

```bash
# Database
export RTX_DB_HOST=localhost
export RTX_DB_NAME=rtx
export RTX_DB_USER=postgres
export RTX_DB_PORT=5432

# Standby
export RTX_STANDBY_HOST=standby1-db
export RTX_PRIMARY_HOST=primary-db

# AWS
export RTX_BACKUP_BUCKET=rtx-backups
export RTX_KMS_KEY=arn:aws:kms:us-east-1:123456789:key/abc-123
export RTX_SNS_TOPIC=arn:aws:sns:us-east-1:123456789:rtx-alerts
export RTX_HOSTED_ZONE_ID=Z1234567890ABC

# API
export RTX_API_URL=http://localhost:7999
export RTX_WS_URL=ws://localhost:7999/ws
export RTX_REDIS_HOST=localhost
```

Load with:
```bash
source /opt/rtx/.env
```

---

## Installation

```bash
# Copy scripts
sudo cp scripts/dr/*.sh /opt/rtx/scripts/dr/

# Make executable
sudo chmod +x /opt/rtx/scripts/dr/*.sh

# Set ownership
sudo chown -R rtx:rtx /opt/rtx/scripts/dr/

# Load environment
source /opt/rtx/.env

# Test
/opt/rtx/scripts/dr/health-check.sh
```

---

## Cron Configuration

Install cron jobs:

```bash
sudo crontab -u rtx config/cron/rtx-dr.cron
```

Or manually:

```cron
# Full backup - Daily 02:00 UTC
0 2 * * * /opt/rtx/scripts/dr/backup-full.sh

# Incremental - Every 6 hours
0 */6 * * * /opt/rtx/scripts/dr/backup-incremental.sh

# Health check - Every minute
* * * * * /opt/rtx/scripts/dr/health-check.sh

# Verify backup - Daily 03:00 UTC
0 3 * * * /opt/rtx/scripts/dr/verify-backup.sh

# Test restore - Sunday 04:00 UTC
0 4 * * 0 /opt/rtx/scripts/dr/test-restore.sh

# Weekly report - Monday 09:00 UTC
0 9 * * 1 /opt/rtx/scripts/dr/generate-weekly-report.sh
```

---

## Logging

All scripts log to `/var/log/rtx/`:

```bash
# View backup logs
tail -f /var/log/rtx/backup-full-*.log

# View health check logs
tail -f /var/log/rtx/health-check.log

# View failover logs
tail -f /var/log/rtx/failover-*.log

# View all DR logs
tail -f /var/log/rtx/*.log
```

---

## Monitoring

Check script execution in database:

```sql
-- Backup history
SELECT * FROM backup_history ORDER BY backup_date DESC LIMIT 10;

-- Restore history
SELECT * FROM restore_history ORDER BY restore_date DESC LIMIT 10;

-- Failover history
SELECT * FROM failover_history ORDER BY failover_date DESC LIMIT 10;

-- Health checks (last hour)
SELECT check_name, status, COUNT(*)
FROM health_check_history
WHERE check_date > NOW() - INTERVAL '1 hour'
GROUP BY check_name, status;

-- Backup compliance
SELECT * FROM check_backup_compliance();

-- DR summary
SELECT * FROM dr_metrics_summary;
```

---

## Troubleshooting

### Backup fails with "Insufficient disk space"

```bash
# Check disk usage
df -h /var/backups/rtx

# Delete old backups
find /var/backups/rtx -name "*.dump" -mtime +7 -delete

# Increase disk or mount larger volume
```

### Restore fails with "Checksum mismatch"

```bash
# Re-download backup
aws s3 cp s3://rtx-backups/backups/full/YYYYMMDD/rtx-full-*.dump /tmp/

# Verify checksum
sha256sum /tmp/rtx-full-*.dump

# Check S3 object integrity
aws s3api head-object --bucket rtx-backups --key backups/full/YYYYMMDD/rtx-full-*.dump
```

### Health check always fails

```bash
# Check PostgreSQL
pg_isready -h localhost

# Check API
curl http://localhost:7999/health

# Check logs
tail -f /var/log/rtx/health-check.log

# Manually run check
/opt/rtx/scripts/dr/health-check.sh
```

### Failover script can't SSH to standby

```bash
# Test SSH
ssh admin@standby1-db "echo test"

# Check SSH keys
cat ~/.ssh/id_rsa.pub
ssh admin@standby1-db "cat ~/.ssh/authorized_keys"

# Add key if missing
ssh-copy-id admin@standby1-db
```

---

## Security

- All backups encrypted with AWS KMS
- Scripts use environment variables (no hardcoded credentials)
- Database passwords in PostgreSQL `.pgpass` file
- SSH keys for remote access (no passwords)
- S3 bucket has versioning and public access blocked
- Audit logging to CloudTrail

---

## Compliance

These scripts satisfy:
- **SOC 2 Type II:** Backup testing, monitoring, alerting
- **ISO 27001:** Disaster recovery procedures
- **PCI DSS:** Data backup and recovery
- **MiFID II:** Transaction data retention
- **GDPR:** Data protection and availability

---

## Support

- **Documentation:** `/opt/rtx/docs/DISASTER_RECOVERY_PLAN.md`
- **Quick Setup:** `/opt/rtx/docs/DR_QUICK_SETUP.md`
- **Logs:** `/var/log/rtx/`
- **Metrics:** `/var/lib/rtx/metrics/`
- **Emergency:** `/opt/rtx/EMERGENCY.txt`

---

**Last Updated:** 2026-01-18
**Maintainer:** DR/BC Team
**Review Cycle:** Quarterly
