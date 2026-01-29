# RTX Trading Engine - DR Quick Setup Guide

This guide will help you implement the Disaster Recovery & Business Continuity plan in **under 2 hours**.

---

## Prerequisites

- Ubuntu 20.04+ or Amazon Linux 2
- PostgreSQL 14+ with TimescaleDB
- AWS CLI configured with proper credentials
- S3 bucket for backups (`rtx-backups`)
- SNS topic for alerts (`rtx-alerts`)
- Root or sudo access

---

## Step 1: Install Dependencies (10 minutes)

```bash
# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install PostgreSQL client tools
sudo apt-get install -y postgresql-client-14

# Install AWS CLI v2
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install

# Verify installations
psql --version
aws --version
```

---

## Step 2: Setup Directory Structure (5 minutes)

```bash
# Create RTX directories
sudo mkdir -p /opt/rtx/{scripts/dr,config,data}
sudo mkdir -p /var/backups/rtx/{full,incremental,restore}
sudo mkdir -p /var/log/rtx
sudo mkdir -p /var/lib/rtx/metrics

# Set ownership
sudo chown -R rtx:rtx /opt/rtx
sudo chown -R postgres:postgres /var/backups/rtx
sudo chown -R rtx:rtx /var/log/rtx

# Set permissions
chmod 750 /opt/rtx/scripts/dr
chmod 640 /opt/rtx/.env
```

---

## Step 3: Copy DR Scripts (10 minutes)

```bash
# Copy scripts from this repo
sudo cp scripts/dr/*.sh /opt/rtx/scripts/dr/
sudo chmod +x /opt/rtx/scripts/dr/*.sh

# Verify scripts
ls -lh /opt/rtx/scripts/dr/
```

Scripts included:
- `backup-full.sh` - Full database backup
- `backup-incremental.sh` - WAL file backup
- `archive-wal.sh` - PostgreSQL archive command
- `restore-full.sh` - Database restore
- `failover-database.sh` - Automatic failover
- `health-check.sh` - System health monitoring

---

## Step 4: Configure Environment Variables (10 minutes)

Create `/opt/rtx/.env`:

```bash
# RTX Trading Engine - Environment Configuration

# Database
export RTX_DB_HOST=localhost
export RTX_DB_NAME=rtx
export RTX_DB_USER=postgres
export RTX_DB_PORT=5432

# Standby Database (for failover)
export RTX_STANDBY_HOST=standby1-db.rtx-trading.com
export RTX_PRIMARY_HOST=primary-db.rtx-trading.com

# AWS Configuration
export RTX_BACKUP_BUCKET=rtx-backups
export RTX_KMS_KEY=arn:aws:kms:us-east-1:123456789:key/abc-123
export RTX_SNS_TOPIC=arn:aws:sns:us-east-1:123456789:rtx-alerts
export RTX_HOSTED_ZONE_ID=Z1234567890ABC

# API Configuration
export RTX_API_URL=http://localhost:7999
export RTX_WS_URL=ws://localhost:7999/ws
export RTX_REDIS_HOST=localhost

# Notification
export RTX_ALERT_EMAIL=ops@rtx-trading.com
```

Load environment:

```bash
source /opt/rtx/.env
# Add to .bashrc for persistence
echo "source /opt/rtx/.env" >> ~/.bashrc
```

---

## Step 5: Setup PostgreSQL for Continuous Archiving (15 minutes)

Edit `/etc/postgresql/14/main/postgresql.conf`:

```ini
# Enable WAL archiving for PITR
wal_level = replica
archive_mode = on
archive_command = '/opt/rtx/scripts/dr/archive-wal.sh %p %f'
archive_timeout = 300  # Force WAL rotation every 5 minutes

# Replication settings
max_wal_senders = 10
wal_keep_size = 1GB
synchronous_commit = on
synchronous_standby_names = 'standby1'

# Performance tuning
shared_buffers = 2GB
effective_cache_size = 6GB
maintenance_work_mem = 512MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1  # For SSD
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
sudo systemctl status postgresql
```

Create WAL archive directory:

```bash
sudo mkdir -p /var/lib/postgresql/14/archive
sudo chown postgres:postgres /var/lib/postgresql/14/archive
sudo chmod 700 /var/lib/postgresql/14/archive
```

---

## Step 6: Setup Database Schema for DR Tracking (10 minutes)

```bash
# Load DR schema
psql -h localhost -U postgres -d rtx -f backend/database/dr-schema.sql

# Verify tables created
psql -h localhost -U postgres -d rtx -c "\dt backup_history restore_history failover_history health_check_history"

# Check backup compliance function
psql -h localhost -U postgres -d rtx -c "SELECT * FROM check_backup_compliance();"
```

---

## Step 7: Configure S3 Bucket (10 minutes)

```bash
# Create S3 bucket (if not exists)
aws s3 mb s3://rtx-backups --region us-east-1

# Enable versioning
aws s3api put-bucket-versioning \
    --bucket rtx-backups \
    --versioning-configuration Status=Enabled

# Enable encryption
aws s3api put-bucket-encryption \
    --bucket rtx-backups \
    --server-side-encryption-configuration '{
        "Rules": [{
            "ApplyServerSideEncryptionByDefault": {
                "SSEAlgorithm": "aws:kms",
                "KMSMasterKeyID": "arn:aws:kms:us-east-1:123456789:key/abc-123"
            }
        }]
    }'

# Set lifecycle policy
aws s3api put-bucket-lifecycle-configuration \
    --bucket rtx-backups \
    --lifecycle-configuration file://config/s3-lifecycle.json

# Block public access
aws s3api put-public-access-block \
    --bucket rtx-backups \
    --public-access-block-configuration \
        BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
```

Create `config/s3-lifecycle.json`:

```json
{
  "Rules": [
    {
      "Id": "ArchiveFullBackups",
      "Status": "Enabled",
      "Prefix": "backups/full/",
      "Transitions": [
        { "Days": 30, "StorageClass": "GLACIER" }
      ],
      "Expiration": { "Days": 365 }
    },
    {
      "Id": "DeleteOldWALFiles",
      "Status": "Enabled",
      "Prefix": "wal-archive/",
      "Expiration": { "Days": 30 }
    }
  ]
}
```

---

## Step 8: Setup Cron Jobs (10 minutes)

```bash
# Edit crontab as rtx user
sudo crontab -u rtx -e

# Paste contents from config/cron/rtx-dr.cron
# Or install directly:
sudo crontab -u rtx config/cron/rtx-dr.cron

# Verify cron jobs
sudo crontab -u rtx -l
```

Key cron jobs:
- Full backup: Daily at 02:00 UTC
- Incremental backup: Every 6 hours
- Health check: Every minute
- Backup verification: Daily at 03:00 UTC
- Restore test: Weekly on Sunday

---

## Step 9: Setup Monitoring with Prometheus (20 minutes)

```bash
# Install Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.40.0/prometheus-2.40.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
sudo mv prometheus-2.40.0.linux-amd64 /opt/prometheus
cd /opt/prometheus

# Copy alerting rules
sudo cp config/monitoring/prometheus-rules.yml /opt/prometheus/rules.yml

# Create prometheus.yml
sudo tee /opt/prometheus/prometheus.yml > /dev/null <<EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - 'rules.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']

scrape_configs:
  - job_name: 'rtx-backend'
    static_configs:
      - targets: ['localhost:7999']

  - job_name: 'postgresql'
    static_configs:
      - targets: ['localhost:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['localhost:9121']

  - job_name: 'node_exporter'
    static_configs:
      - targets: ['localhost:9100']
EOF

# Create systemd service
sudo tee /etc/systemd/system/prometheus.service > /dev/null <<EOF
[Unit]
Description=Prometheus
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/opt/prometheus/prometheus \\
  --config.file=/opt/prometheus/prometheus.yml \\
  --storage.tsdb.path=/var/lib/prometheus/ \\
  --web.console.templates=/opt/prometheus/consoles \\
  --web.console.libraries=/opt/prometheus/console_libraries

[Install]
WantedBy=multi-user.target
EOF

# Start Prometheus
sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl start prometheus
```

Install exporters:

```bash
# PostgreSQL exporter
docker run -d --name postgres_exporter \
  -e DATA_SOURCE_NAME="postgresql://postgres@localhost:5432/rtx?sslmode=disable" \
  -p 9187:9187 \
  prometheuscommunity/postgres-exporter

# Redis exporter
docker run -d --name redis_exporter \
  -p 9121:9121 \
  oliver006/redis_exporter

# Node exporter
wget https://github.com/prometheus/node_exporter/releases/download/v1.5.0/node_exporter-1.5.0.linux-amd64.tar.gz
tar xvfz node_exporter-*.tar.gz
sudo mv node_exporter-1.5.0.linux-amd64/node_exporter /usr/local/bin/
sudo systemctl enable node_exporter
sudo systemctl start node_exporter
```

---

## Step 10: Test Backups (20 minutes)

```bash
# Run manual full backup
/opt/rtx/scripts/dr/backup-full.sh

# Verify backup in S3
aws s3 ls s3://rtx-backups/backups/full/ --recursive

# Verify backup in database
psql -h localhost -U postgres -d rtx -c "SELECT * FROM backup_history ORDER BY backup_date DESC LIMIT 5;"

# Test restore (dry-run)
/opt/rtx/scripts/dr/restore-full.sh latest --verify-only

# Run health check
/opt/rtx/scripts/dr/health-check.sh

# Check health metrics
cat /var/lib/rtx/metrics/health.json | jq .
```

---

## Step 11: Setup Alerting with SNS (10 minutes)

```bash
# Create SNS topic
aws sns create-topic --name rtx-alerts

# Subscribe email
aws sns subscribe \
    --topic-arn arn:aws:sns:us-east-1:123456789:rtx-alerts \
    --protocol email \
    --notification-endpoint ops@rtx-trading.com

# Confirm subscription in email

# Test alert
aws sns publish \
    --topic-arn arn:aws:sns:us-east-1:123456789:rtx-alerts \
    --subject "RTX Test Alert" \
    --message "DR setup completed successfully"
```

---

## Step 12: Document & Train (10 minutes)

```bash
# Create runbook shortcuts
sudo ln -s /opt/rtx/scripts/dr/failover-database.sh /usr/local/bin/rtx-failover
sudo ln -s /opt/rtx/scripts/dr/restore-full.sh /usr/local/bin/rtx-restore
sudo ln -s /opt/rtx/scripts/dr/health-check.sh /usr/local/bin/rtx-health

# Create emergency contact card
cat > /opt/rtx/EMERGENCY.txt <<EOF
RTX TRADING ENGINE - EMERGENCY CONTACTS

DATABASE FAILOVER:
  Command: sudo rtx-failover
  Runbook: /opt/rtx/docs/DISASTER_RECOVERY_PLAN.md

DATABASE RESTORE:
  Command: sudo rtx-restore [date]
  Example: sudo rtx-restore 20260118

HEALTH CHECK:
  Command: sudo rtx-health
  Dashboard: http://prometheus:9090

CONTACTS:
  On-Call Engineer: +1-555-0100
  Database Admin: +1-555-0102
  CTO: +1-555-0104

AWS CONSOLE:
  Backups: https://s3.console.aws.amazon.com/s3/buckets/rtx-backups
  Alerts: https://console.aws.amazon.com/sns/v3/home#/topics

PROMETHEUS:
  http://localhost:9090
EOF

cat /opt/rtx/EMERGENCY.txt
```

---

## Verification Checklist

After setup, verify:

- [ ] PostgreSQL WAL archiving active
- [ ] S3 bucket receiving WAL files
- [ ] Full backup completes successfully
- [ ] Restore test passes (verify-only)
- [ ] Health checks running every minute
- [ ] Prometheus collecting metrics
- [ ] SNS alerts working
- [ ] Cron jobs scheduled correctly
- [ ] DR schema tables created
- [ ] Emergency runbooks accessible

---

## Quick Reference Commands

```bash
# Backup
rtx-backup-now       # Manual full backup
rtx-backup-status    # Check last backup

# Restore
rtx-restore latest --verify-only  # Dry-run test
rtx-restore 20260118              # Actual restore
rtx-restore-pitr "2026-01-18 10:30:00 UTC"  # Point-in-time

# Failover
rtx-failover         # Automatic failover
rtx-failover --force # Force failover (skip checks)

# Monitoring
rtx-health           # Run health check
rtx-metrics          # View DR metrics
rtx-alerts           # List active alerts

# Maintenance
rtx-cleanup          # Delete old backups
rtx-verify-backups   # Verify all backups
rtx-test-restore     # Weekly restore test
```

---

## Next Steps

1. **Schedule DR Drill:** Calendar quarterly DR simulation
2. **Update Contacts:** Fill in actual phone numbers and emails
3. **Setup Secondary Region:** Implement multi-region redundancy
4. **Enable Chaos Engineering:** Test resilience with controlled failures
5. **Review Runbooks:** Familiarize team with failover procedures

---

## Support

- **Documentation:** `/opt/rtx/docs/DISASTER_RECOVERY_PLAN.md`
- **Runbooks:** `/opt/rtx/docs/`
- **Logs:** `/var/log/rtx/`
- **Metrics:** `/var/lib/rtx/metrics/`

---

**Setup Time:** ~2 hours
**Status:** Ready for production
**Last Updated:** 2026-01-18
