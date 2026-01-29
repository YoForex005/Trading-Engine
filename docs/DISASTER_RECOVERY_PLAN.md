# Disaster Recovery & Business Continuity Plan
## RTX Trading Engine v3.0

**Document Version:** 1.0
**Last Updated:** 2026-01-18
**Owner:** DR/BC Team
**Review Cycle:** Quarterly

---

## Executive Summary

This document outlines the comprehensive Disaster Recovery (DR) and Business Continuity (BC) strategy for the RTX Trading Engine. The plan ensures minimal downtime and data loss in the event of system failures, natural disasters, or security incidents.

**Key Metrics:**
- **RTO (Recovery Time Objective):** 15 minutes for critical trading systems
- **RPO (Recovery Point Objective):** 1 minute for transaction data
- **Target Availability:** 99.95% (26.3 minutes downtime/month)

---

## 1. Business Impact Analysis

### 1.1 Critical System Classification

| System Component | Criticality | RTO | RPO | Financial Impact/Hour |
|-----------------|-------------|-----|-----|----------------------|
| Trading Engine Core | **CRITICAL** | 5 min | 30 sec | $50,000+ |
| Market Data Feed (OANDA/Binance) | **CRITICAL** | 10 min | 1 min | $30,000 |
| Order Routing & Execution | **CRITICAL** | 5 min | 30 sec | $50,000+ |
| WebSocket Server (Real-time) | **HIGH** | 10 min | 5 min | $20,000 |
| PostgreSQL Database | **CRITICAL** | 15 min | 1 min | $40,000 |
| FIX Gateway (YoFx) | **HIGH** | 15 min | 5 min | $15,000 |
| Admin Panel | **MEDIUM** | 1 hour | 15 min | $5,000 |
| Desktop Client | **MEDIUM** | 2 hours | N/A | $3,000 |
| Historical Data (OHLC/Ticks) | **LOW** | 24 hours | 1 hour | $1,000 |

### 1.2 Recovery Priority Order

1. **Phase 1 (0-5 min):** PostgreSQL database + Trading Engine Core
2. **Phase 2 (5-15 min):** Market data feeds + Order routing + WebSocket
3. **Phase 3 (15-60 min):** FIX Gateway + Admin panel
4. **Phase 4 (1-24 hours):** Historical data restoration + Desktop clients

---

## 2. Backup Strategy

### 2.1 Database Backups (PostgreSQL + TimescaleDB)

#### 2.1.1 Backup Schedule

```bash
# Full Backup: Daily at 02:00 UTC
0 2 * * * /opt/rtx/scripts/backup-full.sh

# Incremental Backup: Every 6 hours
0 */6 * * * /opt/rtx/scripts/backup-incremental.sh

# Transaction Log Backup: Every 15 minutes
*/15 * * * * /opt/rtx/scripts/backup-wal.sh

# Continuous WAL Archiving: Real-time
archive_mode = on
archive_command = '/opt/rtx/scripts/archive-wal.sh %p %f'
```

#### 2.1.2 Backup Retention Policy

| Backup Type | Retention (Hot) | Retention (Cold) | Storage Location |
|-------------|----------------|------------------|------------------|
| Full Backup | 7 days | 1 year | S3 Standard / GCS Standard |
| Incremental | 7 days | 30 days | S3 Standard |
| WAL Logs | 7 days | 30 days | S3 Standard |
| Point-in-Time Snapshots | 24 hours | 30 days | S3 Intelligent-Tiering |

#### 2.1.3 Backup Verification

```bash
# Automated restore test: Daily
0 3 * * * /opt/rtx/scripts/verify-backup.sh

# Checksum validation: Every backup
sha256sum /backups/latest.tar.gz > /backups/latest.sha256

# Data integrity check
pg_verifybackup /backups/latest
```

### 2.2 Application Data Backups

#### 2.2.1 Configuration Files

```bash
# Backup locations
/opt/rtx/config/           -> S3://rtx-backups/config/
/opt/rtx/backend/data/     -> S3://rtx-backups/data/
/etc/nginx/                -> S3://rtx-backups/nginx/
/etc/systemd/system/rtx*   -> S3://rtx-backups/systemd/
```

#### 2.2.2 Market Data (Ticks & OHLC)

```bash
# File-based tick storage
/opt/rtx/backend/data/ticks/ -> S3://rtx-ticks/
/opt/rtx/backend/data/ohlc/  -> S3://rtx-ohlc/

# Sync every hour
0 * * * * aws s3 sync /opt/rtx/backend/data/ticks/ s3://rtx-ticks/ --delete
```

### 2.3 Secret & Credential Backup

```bash
# Encrypted backup of secrets
/opt/rtx/.env -> Encrypted -> S3://rtx-secrets/
API Keys -> AWS Secrets Manager / HashiCorp Vault
FIX Session Configs -> Encrypted S3
```

---

## 3. High Availability Architecture

### 3.1 Database HA (PostgreSQL Streaming Replication)

```
┌─────────────────┐
│   Primary DB    │
│  (RW Instance)  │ ──────┐
└─────────────────┘       │
         │                │ WAL Stream
         │ Sync Repl      │
         ▼                ▼
┌─────────────────┐  ┌─────────────────┐
│  Standby DB 1   │  │  Standby DB 2   │
│  (Sync Replica) │  │ (Async Replica) │
│   us-east-1a    │  │   us-west-2a    │
└─────────────────┘  └─────────────────┘
```

**Configuration:**

```sql
-- postgresql.conf (Primary)
wal_level = replica
max_wal_senders = 10
synchronous_commit = on
synchronous_standby_names = 'standby1'

-- recovery.conf (Standby)
standby_mode = on
primary_conninfo = 'host=primary-db port=5432 user=replicator password=***'
restore_command = 'cp /var/lib/postgresql/archive/%f %p'
```

**Automatic Failover with Patroni:**

```yaml
# patroni.yml
scope: rtx-cluster
namespace: /service/
name: rtx-primary

restapi:
  listen: 0.0.0.0:8008
  connect_address: rtx-primary:8008

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 1048576
    postgresql:
      parameters:
        max_connections: 500
        shared_buffers: 2GB
        effective_cache_size: 6GB
```

### 3.2 Application HA (Horizontal Scaling)

```
                    ┌──────────────┐
                    │ Load Balancer│
                    │   (HAProxy)  │
                    └──────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  RTX Node 1  │  │  RTX Node 2  │  │  RTX Node 3  │
│  (Active)    │  │  (Active)    │  │  (Active)    │
│ us-east-1a   │  │ us-east-1b   │  │ us-east-1c   │
└──────────────┘  └──────────────┘  └──────────────┘
```

**HAProxy Configuration:**

```haproxy
frontend rtx_api
    bind *:7999
    mode http
    option httplog
    default_backend rtx_servers

backend rtx_servers
    mode http
    balance leastconn
    option httpchk GET /health
    http-check expect status 200
    server rtx1 10.0.1.10:7999 check inter 5s fall 3 rise 2
    server rtx2 10.0.1.11:7999 check inter 5s fall 3 rise 2
    server rtx3 10.0.1.12:7999 check inter 5s fall 3 rise 2
```

### 3.3 WebSocket HA (Session Persistence)

```
┌──────────────┐
│   Client     │
└──────────────┘
        │
        ▼
┌──────────────┐
│   ALB/NLB    │  ← Sticky sessions enabled
│ (Target IP)  │
└──────────────┘
        │
        ▼
┌──────────────┐
│ WebSocket    │  ← Redis Pub/Sub for broadcast
│   Server     │
└──────────────┘
```

**Redis Pub/Sub for Multi-Instance WS:**

```go
// Broadcast to all WS instances via Redis
func (h *Hub) BroadcastTick(tick *MarketTick) {
    h.redis.Publish("market-ticks", tick)
}

// Subscribe on each instance
func (h *Hub) SubscribeToRedis() {
    pubsub := h.redis.Subscribe("market-ticks")
    for msg := range pubsub.Channel() {
        h.localBroadcast(msg.Payload)
    }
}
```

### 3.4 Cache Layer (Redis Sentinel)

```
┌─────────────────┐
│  Redis Master   │
│  (Port 6379)    │
└─────────────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌─────────┐ ┌─────────┐
│ Slave 1 │ │ Slave 2 │
└─────────┘ └─────────┘
         │
         ▼
┌─────────────────┐
│  Sentinel Quorum│
│   (3 instances) │
└─────────────────┘
```

**Redis Sentinel Config:**

```
sentinel monitor rtx-redis 10.0.1.20 6379 2
sentinel down-after-milliseconds rtx-redis 5000
sentinel failover-timeout rtx-redis 10000
sentinel parallel-syncs rtx-redis 1
```

---

## 4. Geographic Redundancy

### 4.1 Multi-Region Deployment

```
┌──────────────────────────────────────────────────────────┐
│                     Route 53 (DNS)                       │
│         Health Checks + Failover Routing Policy          │
└──────────────────────────────────────────────────────────┘
                        │
        ┌───────────────┴───────────────┐
        │                               │
        ▼                               ▼
┌──────────────────┐          ┌──────────────────┐
│  PRIMARY REGION  │          │ SECONDARY REGION │
│    us-east-1     │          │    us-west-2     │
├──────────────────┤          ├──────────────────┤
│ - RTX Cluster    │          │ - RTX Cluster    │
│ - PostgreSQL     │  WAL     │ - PostgreSQL     │
│ - Redis          │◄────────►│ - Redis          │
│ - Load Balancer  │  Async   │ - Load Balancer  │
└──────────────────┘  Repl    └──────────────────┘
```

### 4.2 Cross-Region Data Replication

**PostgreSQL Logical Replication:**

```sql
-- Primary (us-east-1)
CREATE PUBLICATION rtx_pub FOR ALL TABLES;

-- Secondary (us-west-2)
CREATE SUBSCRIPTION rtx_sub
    CONNECTION 'host=primary-db port=5432 dbname=rtx user=replicator'
    PUBLICATION rtx_pub
    WITH (copy_data = true);
```

**Redis Cross-Region Replication:**

```bash
# Use Redis Enterprise Active-Active or
# AWS ElastiCache Global Datastore
```

### 4.3 DNS Failover Configuration

```yaml
# Route53 Health Check
HealthCheck:
  Type: HTTPS
  ResourcePath: /health
  FullyQualifiedDomainName: api.rtx-trading.com
  Port: 443
  RequestInterval: 30
  FailureThreshold: 3

# Failover Record
RecordSet:
  Name: api.rtx-trading.com
  Type: A
  SetIdentifier: Primary
  Failover: PRIMARY
  HealthCheckId: !Ref HealthCheck
  AliasTarget:
    DNSName: primary-lb.us-east-1.amazonaws.com

RecordSet:
  Name: api.rtx-trading.com
  Type: A
  SetIdentifier: Secondary
  Failover: SECONDARY
  AliasTarget:
    DNSName: secondary-lb.us-west-2.amazonaws.com
```

---

## 5. Disaster Recovery Procedures

### 5.1 Database Recovery Procedures

#### 5.1.1 Full Database Restore

```bash
#!/bin/bash
# /opt/rtx/scripts/restore-full.sh

BACKUP_FILE=$1
RESTORE_DIR=/var/lib/postgresql/restore

# 1. Stop PostgreSQL
sudo systemctl stop postgresql

# 2. Clear existing data
sudo rm -rf /var/lib/postgresql/14/main/*

# 3. Extract backup
tar -xzf "$BACKUP_FILE" -C "$RESTORE_DIR"

# 4. Copy data
sudo cp -r "$RESTORE_DIR"/* /var/lib/postgresql/14/main/

# 5. Set permissions
sudo chown -R postgres:postgres /var/lib/postgresql/14/main

# 6. Start PostgreSQL
sudo systemctl start postgresql

# 7. Verify
psql -U postgres -d rtx -c "SELECT COUNT(*) FROM rtx_accounts;"
```

#### 5.1.2 Point-in-Time Recovery (PITR)

```bash
#!/bin/bash
# Restore to specific timestamp: 2026-01-18 10:30:00 UTC

TARGET_TIME="2026-01-18 10:30:00 UTC"

# 1. Stop PostgreSQL
sudo systemctl stop postgresql

# 2. Restore latest base backup
tar -xzf /backups/base-backup-latest.tar.gz -C /var/lib/postgresql/14/main/

# 3. Create recovery.conf
cat > /var/lib/postgresql/14/main/recovery.conf <<EOF
restore_command = 'cp /var/lib/postgresql/archive/%f %p'
recovery_target_time = '$TARGET_TIME'
recovery_target_action = promote
EOF

# 4. Start PostgreSQL (auto-recovery)
sudo systemctl start postgresql

# 5. Monitor recovery
tail -f /var/log/postgresql/postgresql-14-main.log
```

#### 5.1.3 Failover to Replica

```bash
#!/bin/bash
# /opt/rtx/scripts/failover-replica.sh

# Using Patroni for automatic failover
patronictl -c /etc/patroni/patroni.yml failover rtx-cluster

# Manual failover (if Patroni unavailable)
# On Standby Server:
sudo -u postgres psql -c "SELECT pg_promote();"

# Update application connection string
sed -i 's/primary-db/standby1-db/g' /opt/rtx/.env

# Restart application
sudo systemctl restart rtx-backend
```

### 5.2 Application Recovery

#### 5.2.1 Deploy New Instance from AMI

```bash
#!/bin/bash
# Launch replacement instance from latest AMI

LATEST_AMI=$(aws ec2 describe-images \
    --owners self \
    --filters "Name=name,Values=rtx-backend-*" \
    --query 'sort_by(Images, &CreationDate)[-1].ImageId' \
    --output text)

aws ec2 run-instances \
    --image-id $LATEST_AMI \
    --instance-type c5.2xlarge \
    --key-name rtx-prod-key \
    --security-group-ids sg-xxxxx \
    --subnet-id subnet-xxxxx \
    --iam-instance-profile Name=rtx-backend-role \
    --user-data file://user-data.sh \
    --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=rtx-backend-dr}]'
```

#### 5.2.2 Restore Configuration from Backup

```bash
#!/bin/bash
# /opt/rtx/scripts/restore-config.sh

S3_BUCKET="rtx-backups"
RESTORE_DATE=$(date -u +%Y-%m-%d)

# Download latest config
aws s3 cp s3://$S3_BUCKET/config/$RESTORE_DATE/ /opt/rtx/config/ --recursive

# Download secrets
aws secretsmanager get-secret-value \
    --secret-id rtx/prod/env \
    --query SecretString \
    --output text > /opt/rtx/.env

# Restore systemd services
aws s3 cp s3://$S3_BUCKET/systemd/ /etc/systemd/system/ --recursive --exclude "*" --include "rtx*"

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart rtx-backend
```

### 5.3 Region Failover Procedure

```bash
#!/bin/bash
# /opt/rtx/scripts/region-failover.sh

# 1. Update DNS to point to secondary region
aws route53 change-resource-record-sets \
    --hosted-zone-id Z1234567890ABC \
    --change-batch file://failover-dns.json

# 2. Promote secondary database to primary
ssh admin@us-west-2-db "sudo -u postgres psql -c 'SELECT pg_promote();'"

# 3. Update application config in secondary region
ssh admin@us-west-2-app "sed -i 's/standby/primary/g' /opt/rtx/.env"

# 4. Restart applications in secondary region
ssh admin@us-west-2-app "sudo systemctl restart rtx-backend"

# 5. Verify services
curl -f https://api.rtx-trading.com/health || echo "FAILOVER FAILED"

# 6. Send notification
aws sns publish \
    --topic-arn arn:aws:sns:us-east-1:123456789:rtx-alerts \
    --message "Region failover completed: us-east-1 -> us-west-2"
```

---

## 6. Monitoring & Alerting

### 6.1 Health Checks

```yaml
# /opt/rtx/monitoring/health-checks.yml
checks:
  - name: database_connection
    type: postgresql
    interval: 10s
    timeout: 5s
    query: "SELECT 1"

  - name: api_health
    type: http
    url: http://localhost:7999/health
    interval: 10s
    expected_status: 200

  - name: websocket_health
    type: websocket
    url: ws://localhost:7999/ws
    interval: 30s

  - name: disk_space
    type: disk
    path: /var/lib/postgresql
    threshold: 85%

  - name: replication_lag
    type: postgresql
    query: "SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()));"
    threshold: 60  # seconds
```

### 6.2 Alert Configuration

```yaml
# alertmanager.yml
route:
  receiver: 'rtx-oncall'
  group_by: ['alertname', 'severity']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

  routes:
    - match:
        severity: critical
      receiver: 'pagerduty-critical'
      continue: true

    - match:
        severity: warning
      receiver: 'slack-warnings'

receivers:
  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: <key>
        description: '{{ .CommonAnnotations.description }}'

  - name: 'slack-warnings'
    slack_configs:
      - api_url: <webhook>
        channel: '#rtx-alerts'

  - name: 'rtx-oncall'
    email_configs:
      - to: 'oncall@rtx-trading.com'
```

### 6.3 Critical Metrics

```prometheus
# Prometheus rules
groups:
  - name: rtx_database
    interval: 10s
    rules:
      - alert: DatabaseDown
        expr: up{job="postgresql"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          description: "PostgreSQL is down for {{ $value }}s"

      - alert: ReplicationLagHigh
        expr: pg_replication_lag_seconds > 60
        for: 5m
        labels:
          severity: warning

      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.15
        for: 10m
        labels:
          severity: warning

  - name: rtx_application
    interval: 10s
    rules:
      - alert: APIHighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical

      - alert: WebSocketConnectionsDrop
        expr: rate(websocket_connections_total[5m]) < -100
        for: 2m
        labels:
          severity: warning
```

---

## 7. Runbooks

### 7.1 Database Failover Runbook

**Scenario:** Primary PostgreSQL database failure

**Detection:**
- Alert: `DatabaseDown` fired
- Replication lag increases to infinity
- Application logs show connection errors

**Response Steps:**

1. **Verify primary is down (1 min)**
   ```bash
   pg_isready -h primary-db.rtx-trading.com
   ssh admin@primary-db "sudo systemctl status postgresql"
   ```

2. **Check standby health (1 min)**
   ```bash
   patronictl -c /etc/patroni/patroni.yml list
   ```

3. **Initiate failover (2 min)**
   ```bash
   # Automatic with Patroni
   patronictl -c /etc/patroni/patroni.yml failover rtx-cluster --force

   # Manual if needed
   ssh admin@standby1 "sudo -u postgres psql -c 'SELECT pg_promote();'"
   ```

4. **Update application DNS/config (3 min)**
   ```bash
   # Patroni updates automatically, or:
   aws route53 change-resource-record-sets --hosted-zone-id Z123 --change-batch file://promote-standby.json
   ```

5. **Verify application connectivity (2 min)**
   ```bash
   curl -f https://api.rtx-trading.com/health
   psql -h standby1-db -c "SELECT COUNT(*) FROM rtx_trades WHERE created_at > NOW() - INTERVAL '1 hour';"
   ```

6. **Monitor for 15 minutes**
   - Check error logs
   - Verify trade execution
   - Monitor replication lag on new standby

7. **Post-incident**
   - Investigate root cause
   - Rebuild failed primary as new standby
   - Update documentation

**Estimated Total RTO:** 9-15 minutes

### 7.2 Application Deployment Runbook

**Scenario:** Deploy new backend version with rollback capability

```bash
#!/bin/bash
# /opt/rtx/scripts/deploy-backend.sh

VERSION=$1
ROLLBACK_VERSION=$(cat /opt/rtx/current-version.txt)

# 1. Download new version
aws s3 cp s3://rtx-releases/rtx-backend-${VERSION}.tar.gz /tmp/

# 2. Extract to staging
tar -xzf /tmp/rtx-backend-${VERSION}.tar.gz -C /opt/rtx/staging/

# 3. Run smoke tests
/opt/rtx/staging/server --version
/opt/rtx/staging/server --validate-config

# 4. Backup current version
cp -r /opt/rtx/backend /opt/rtx/backend-rollback-${ROLLBACK_VERSION}

# 5. Deploy (blue-green)
sudo systemctl stop rtx-backend
cp -r /opt/rtx/staging/* /opt/rtx/backend/
sudo systemctl start rtx-backend

# 6. Health check (30s timeout)
for i in {1..30}; do
    if curl -f http://localhost:7999/health; then
        echo "Deployment successful"
        echo $VERSION > /opt/rtx/current-version.txt
        exit 0
    fi
    sleep 1
done

# 7. Rollback on failure
echo "Deployment failed, rolling back to $ROLLBACK_VERSION"
sudo systemctl stop rtx-backend
cp -r /opt/rtx/backend-rollback-${ROLLBACK_VERSION}/* /opt/rtx/backend/
sudo systemctl start rtx-backend
exit 1
```

### 7.3 Data Recovery Runbook

**Scenario:** Accidental data deletion or corruption

```bash
#!/bin/bash
# /opt/rtx/scripts/recover-data.sh

TARGET_TIME="2026-01-18 14:30:00 UTC"
AFFECTED_TABLE="rtx_positions"

# 1. Create recovery database
sudo -u postgres psql -c "CREATE DATABASE rtx_recovery;"

# 2. Restore to recovery database
pg_restore -d rtx_recovery /backups/rtx-full-$(date +%Y%m%d).dump

# 3. Apply WAL logs up to target time
sudo -u postgres pg_waldump /var/lib/postgresql/archive/ > /tmp/wal-analysis.log

# 4. Extract missing data
psql -d rtx_recovery -c "
    COPY (
        SELECT * FROM $AFFECTED_TABLE
        WHERE open_time >= '$TARGET_TIME'::timestamptz - INTERVAL '1 hour'
        AND open_time <= '$TARGET_TIME'::timestamptz
    ) TO '/tmp/recovered-data.csv' CSV HEADER;
"

# 5. Review and import to production
# MANUAL REVIEW REQUIRED
cat /tmp/recovered-data.csv

# After review:
psql -d rtx -c "COPY $AFFECTED_TABLE FROM '/tmp/recovered-data.csv' CSV HEADER;"
```

---

## 8. Testing & Validation

### 8.1 DR Test Schedule

| Test Type | Frequency | Last Executed | Next Scheduled |
|-----------|-----------|---------------|----------------|
| Backup Restore Test | Weekly | 2026-01-15 | 2026-01-22 |
| Database Failover Drill | Monthly | 2026-01-10 | 2026-02-10 |
| Full DR Simulation | Quarterly | 2025-12-15 | 2026-03-15 |
| Region Failover Test | Semi-Annual | 2025-10-01 | 2026-04-01 |
| Chaos Engineering | Weekly | 2026-01-17 | 2026-01-24 |

### 8.2 Automated DR Tests

```bash
#!/bin/bash
# /opt/rtx/scripts/test-dr-weekly.sh

DATE=$(date +%Y-%m-%d)
LOG_FILE="/var/log/rtx/dr-test-${DATE}.log"

echo "=== DR Test Started: $DATE ===" | tee -a $LOG_FILE

# Test 1: Backup Restore
echo "[1/5] Testing backup restore..." | tee -a $LOG_FILE
/opt/rtx/scripts/test-restore.sh | tee -a $LOG_FILE

# Test 2: Database connectivity failover
echo "[2/5] Testing DB failover..." | tee -a $LOG_FILE
/opt/rtx/scripts/test-db-failover.sh | tee -a $LOG_FILE

# Test 3: Application instance replacement
echo "[3/5] Testing instance replacement..." | tee -a $LOG_FILE
/opt/rtx/scripts/test-instance-replace.sh | tee -a $LOG_FILE

# Test 4: Configuration restore
echo "[4/5] Testing config restore..." | tee -a $LOG_FILE
/opt/rtx/scripts/test-config-restore.sh | tee -a $LOG_FILE

# Test 5: DNS failover
echo "[5/5] Testing DNS failover..." | tee -a $LOG_FILE
/opt/rtx/scripts/test-dns-failover.sh | tee -a $LOG_FILE

# Generate report
echo "=== DR Test Completed ===" | tee -a $LOG_FILE
cat $LOG_FILE | mail -s "DR Test Report - $DATE" ops@rtx-trading.com
```

### 8.3 Chaos Engineering Tests

```yaml
# chaos-experiments.yml
experiments:
  - name: kill_random_backend_instance
    type: ec2_termination
    schedule: "0 */6 * * *"  # Every 6 hours
    filters:
      tag: "Environment=production,Service=rtx-backend"
    count: 1

  - name: inject_network_latency
    type: network_latency
    schedule: "0 9 * * 1"  # Every Monday 9am
    target: rtx-database
    latency: 500ms
    duration: 10m

  - name: fill_disk_space
    type: disk_fill
    schedule: "0 10 * * 3"  # Every Wednesday 10am
    target: rtx-backend
    path: /var/lib/postgresql
    size: 80%
    duration: 5m
```

---

## 9. Security During DR

### 9.1 Backup Encryption

```bash
# Encrypt backups using AWS KMS
aws s3 cp /backups/rtx-full.dump \
    s3://rtx-backups/encrypted/ \
    --sse aws:kms \
    --sse-kms-key-id arn:aws:kms:us-east-1:123456789:key/abc-123

# Encrypt local backups with GPG
gpg --encrypt --recipient ops@rtx-trading.com /backups/rtx-full.dump
```

### 9.2 Access Control

```yaml
# IAM Policy for DR operations
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::rtx-backups/*"
      ],
      "Condition": {
        "StringEquals": {
          "aws:PrincipalTag/Role": "DR-Team"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:RunInstances",
        "ec2:TerminateInstances"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "aws:RequestedRegion": ["us-east-1", "us-west-2"]
        }
      }
    }
  ]
}
```

### 9.3 Audit Logging

```bash
# CloudTrail for all DR activities
aws cloudtrail create-trail \
    --name rtx-dr-trail \
    --s3-bucket-name rtx-audit-logs \
    --include-global-service-events \
    --is-multi-region-trail \
    --enable-log-file-validation

# Log all DR script executions
exec > >(tee -a /var/log/rtx/dr-operations.log)
exec 2>&1
echo "$(date -u +%Y-%m-%dT%H:%M:%SZ) - $USER - $0 $@"
```

---

## 10. Business Continuity

### 10.1 Communication Plan

| Stakeholder | Contact Method | Update Frequency |
|-------------|---------------|------------------|
| C-Level Executives | Email + SMS | Every 30 min |
| Operations Team | Slack #incidents | Real-time |
| Clients (Traders) | Status Page + Email | Every 15 min |
| Regulators | Email | Within 1 hour |
| Liquidity Providers | FIX Message + Email | Within 15 min |

### 10.2 Status Page Templates

```markdown
# INCIDENT: Trading System Degraded Performance

**Status:** Investigating
**Impact:** High - Order execution delayed by 30-60 seconds
**Start Time:** 2026-01-18 14:32 UTC
**Affected Services:** Trading Engine, Order Routing

## Updates

**14:45 UTC** - We are investigating reports of delayed order execution. Market data feeds are operational.

**14:52 UTC** - Root cause identified: Primary database failover in progress. ETA for resolution: 5 minutes.

**14:58 UTC** - Failover completed. System returning to normal operation. Monitoring for stability.

**15:05 UTC** - All systems operational. Incident resolved.

## Resolution
Automatic failover to standby database completed successfully. No data loss occurred. All pending orders processed.
```

### 10.3 Work-From-Home Readiness

```yaml
# Remote access setup
vpn:
  provider: OpenVPN
  locations:
    - us-east-1
    - us-west-2
  mfa_required: true

remote_tools:
  - name: Jump Host
    access: SSH with MFA

  - name: AWS Console
    access: SSO with MFA

  - name: Database Access
    access: pgBouncer via VPN

  - name: Monitoring
    access: Grafana (read-only for non-ops)
```

---

## 11. Contact Information

### 11.1 Escalation Path

```
Level 1: On-Call Engineer (responds within 5 min)
   ↓
Level 2: Infrastructure Lead (responds within 15 min)
   ↓
Level 3: CTO (responds within 30 min)
   ↓
Level 4: CEO (for regulatory/legal issues)
```

### 11.2 Emergency Contacts

| Role | Name | Phone | Email | Backup |
|------|------|-------|-------|--------|
| On-Call Engineer | [Name] | +1-555-0100 | oncall@rtx.com | PagerDuty |
| Infra Lead | [Name] | +1-555-0101 | infra@rtx.com | Slack DM |
| Database Admin | [Name] | +1-555-0102 | dba@rtx.com | Phone |
| Security Lead | [Name] | +1-555-0103 | security@rtx.com | Phone |
| CTO | [Name] | +1-555-0104 | cto@rtx.com | Phone |

### 11.3 Vendor Contacts

| Vendor | Service | Support Number | Account ID |
|--------|---------|----------------|------------|
| AWS | Cloud Infrastructure | +1-866-788-0188 | 123456789 |
| OANDA | FX Liquidity | +1-800-xxx-xxxx | RTX-001 |
| YoFx | FIX Gateway | +44-xxx-xxxx | Client-123 |
| PagerDuty | Incident Management | support@pagerduty.com | - |

---

## 12. Compliance & Regulatory

### 12.1 Regulatory Reporting

```bash
# Generate incident report for regulators
/opt/rtx/scripts/generate-incident-report.sh \
    --incident-id INC-2026-001 \
    --start-time "2026-01-18 14:32:00 UTC" \
    --end-time "2026-01-18 15:05:00 UTC" \
    --impact "Order execution delayed 30-60s" \
    --output /tmp/regulatory-report.pdf

# Auto-send to regulatory bodies within 1 hour
mail -s "Incident Report - RTX Trading" \
    -a /tmp/regulatory-report.pdf \
    compliance@sec.gov < /tmp/report-body.txt
```

### 12.2 Data Retention for DR

| Data Type | Retention Period | Justification |
|-----------|------------------|---------------|
| Transaction Logs | 7 years | Regulatory requirement |
| Backups (Full) | 1 year | Business continuity |
| Audit Logs | 5 years | Compliance |
| Incident Reports | 10 years | Legal |

---

## Document Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-18 | DR Team | Initial comprehensive DR/BC plan |

---

## Approval Signatures

**CTO:** ______________________ Date: __________
**CFO:** ______________________ Date: __________
**Compliance Officer:** ________ Date: __________

---

**Next Review Date:** 2026-04-18
