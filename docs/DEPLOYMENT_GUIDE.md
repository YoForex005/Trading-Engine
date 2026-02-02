# Production Deployment Guide

**Date**: January 30, 2026
**Version**: 1.0
**Status**: Production Ready

---

## Table of Contents

1. [Pre-Deployment Checklist](#pre-deployment-checklist)
2. [Environment Setup](#environment-setup)
3. [Security Configuration](#security-configuration)
4. [Production Deployment](#production-deployment)
5. [Post-Deployment Verification](#post-deployment-verification)
6. [Monitoring & Alerting](#monitoring--alerting)
7. [Rollback Procedures](#rollback-procedures)
8. [Troubleshooting](#troubleshooting)

---

## Pre-Deployment Checklist

### Phase 1: Planning (Week -1)

- [ ] Review architecture documentation
- [ ] Identify all required environment variables
- [ ] Plan database migration strategy
- [ ] Coordinate with YOFX for production credentials
- [ ] Set up monitoring infrastructure
- [ ] Create incident response procedures
- [ ] Schedule deployment window
- [ ] Notify stakeholders of deployment

### Phase 2: Security (Day -1)

- [ ] Generate all production credentials
- [ ] Create bcrypt hash for admin password
- [ ] Generate JWT secret (32+ bytes)
- [ ] Generate encryption key (32 bytes base64)
- [ ] Create CSRF secret
- [ ] Store all credentials in secure vault
- [ ] Document credential rotation procedure
- [ ] Set up SSL/TLS certificates

### Phase 3: Infrastructure (Day -1)

- [ ] Provision production database (PostgreSQL)
- [ ] Provision Redis instance
- [ ] Configure network security groups
- [ ] Enable encryption at rest
- [ ] Set up backup procedures
- [ ] Configure firewall rules
- [ ] Enable VPN/proxy for admin access
- [ ] Set up load balancer (if multi-instance)

### Phase 4: Code (Day 0)

- [ ] Review all changes in SIMULATION_REMOVAL_CHANGELOG.md
- [ ] Run all tests locally
- [ ] Build production Docker image (if using containers)
- [ ] Verify dependencies are pinned
- [ ] Check for any hardcoded values
- [ ] Verify .gitignore includes .env files
- [ ] Tag release in git
- [ ] Prepare rollback strategy

---

## Environment Setup

### 1. Production Database Setup

**PostgreSQL Configuration**:

```bash
# Create production database
createdb -U postgres -E UTF8 trading_prod

# Create trading user with limited permissions
psql -U postgres -d trading_prod << EOF
CREATE USER trading_user WITH PASSWORD 'secure_db_password_min_32_chars';
GRANT CONNECT ON DATABASE trading_prod TO trading_user;
GRANT USAGE ON SCHEMA public TO trading_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO trading_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO trading_user;
EOF

# Verify connection
psql -U trading_user -d trading_prod -h localhost -c "SELECT 1"
```

**Backup Configuration**:

```bash
# Create backup directory
mkdir -p /var/backups/trading-engine
chmod 700 /var/backups/trading-engine

# Set up daily backup script
cat > /usr/local/bin/backup-trading-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/var/backups/trading-engine"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/trading_prod_$DATE.sql.gz"

pg_dump -U trading_user -d trading_prod | gzip > "$BACKUP_FILE"

# Keep only last 30 days
find "$BACKUP_DIR" -name "trading_prod_*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE"
EOF

chmod +x /usr/local/bin/backup-trading-db.sh

# Add to crontab for daily 2 AM backup
# 0 2 * * * /usr/local/bin/backup-trading-db.sh >> /var/log/trading-backup.log 2>&1
```

### 2. Redis Setup

**Production Redis Configuration**:

```bash
# Install Redis
apt-get install redis-server redis-tools

# Configuration file: /etc/redis/redis.conf
# Apply these settings:
# - requirepass your_redis_password
# - maxmemory 4gb  (adjust based on machine)
# - maxmemory-policy allkeys-lru  (evict least recently used)
# - appendonly yes  (persistence)
# - appendfsync everysec  (balance between safety and performance)

# Start Redis
systemctl start redis-server
systemctl enable redis-server  # Auto-start on reboot

# Verify connection
redis-cli -a your_redis_password ping
# Expected: PONG
```

### 3. Application Directory Setup

```bash
# Create application directory
mkdir -p /opt/trading-engine/backend
mkdir -p /var/log/trading-engine
mkdir -p /var/run/trading-engine

# Set permissions
chown -R trading:trading /opt/trading-engine
chown -R trading:trading /var/log/trading-engine
chown -R trading:trading /var/run/trading-engine

chmod 750 /opt/trading-engine
chmod 750 /var/log/trading-engine
chmod 750 /var/run/trading-engine
```

---

## Security Configuration

### 1. Generate Secure Credentials

**Admin Password Hash** (bcrypt):
```bash
# Generate hash for admin password
echo -n "YourSecurePassword123!@#" | htpasswd -niB -C 10 admin | cut -d ":" -f 2

# Output example: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/eIK
# Copy this value to ADMIN_PASSWORD_HASH environment variable
```

**JWT Secret** (random 32+ bytes base64):
```bash
# Generate JWT secret
openssl rand -base64 32

# Output example: rK3L+fQ7mNzXaYpQ1wSbVz9eJtGhP2fXxK8R4vL5pM0=
# Copy this value to JWT_SECRET environment variable
```

**Encryption Key** (random 32 bytes base64):
```bash
# Generate master encryption key
openssl rand -base64 32

# Output example: xM2N+gR8oOaYpQ1wSbVz9eJtGhP2fXxK8R4vL5pM0=
# Copy this value to MASTER_ENCRYPTION_KEY environment variable
```

**CSRF Secret** (random 32+ bytes base64):
```bash
# Generate CSRF secret
openssl rand -base64 32

# Copy this value to CSRF_SECRET environment variable
```

### 2. Create Production .env File

**Location**: `/opt/trading-engine/backend/.env.production`

**Permissions**: `600` (readable only by trading user)

**Content**:
```bash
# ============================================
# ENVIRONMENT CONFIGURATION
# ============================================
ENVIRONMENT=production
PORT=7999

# ============================================
# SECURITY CONFIGURATION (from above)
# ============================================
ADMIN_PASSWORD_HASH=$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/eIK
ADMIN_EMAIL=admin@yourbroker.com
ADMIN_IP_WHITELIST=203.0.113.0/24,198.51.100.0/24  # Your office IPs only

JWT_SECRET=rK3L+fQ7mNzXaYpQ1wSbVz9eJtGhP2fXxK8R4vL5pM0=
JWT_EXPIRY=24h

MASTER_ENCRYPTION_KEY=xM2N+gR8oOaYpQ1wSbVz9eJtGhP2fXxK8R4vL5pM0=
CSRF_SECRET=aBcDeF1gHiJkLmNoPqRsT2uVwXyZaBcDeF1gHiJkLmNo

# ============================================
# YOFX PRODUCTION CREDENTIALS
# ============================================
YOFX_PROXY_HOST=prod-yofx.broker.com
YOFX_PROXY_USER=prod_yofx_username
YOFX_PROXY_PASS=prod_yofx_password_from_vault

# ============================================
# BROKER CONFIGURATION
# ============================================
BROKER_NAME=RTX Trading
BROKER_DISPLAY_NAME=RTX Trading Corporation
PRICE_FEED_LP=YOFX
PRICE_FEED_NAME=YOFX Professional
EXECUTION_MODE=BBOOK
MARGIN_MODE=HEDGING
DEFAULT_ACCOUNT_LEVERAGE=100
DEFAULT_ACCOUNT_BALANCE=10000.0
DEFAULT_ACCOUNT_CURRENCY=USD
MAX_TICKS_PER_SYMBOL=50000

# ============================================
# DATABASE CONFIGURATION
# ============================================
DB_HOST=trading-db.internal
DB_PORT=5432
DB_NAME=trading_prod
DB_USER=trading_user
DB_PASSWORD=secure_db_password_from_vault
DB_SSL_MODE=require

# ============================================
# REDIS CONFIGURATION
# ============================================
REDIS_HOST=trading-cache.internal
REDIS_PORT=6379
REDIS_PASSWORD=redis_password_from_vault

# ============================================
# CORS CONFIGURATION
# ============================================
CORS_ALLOWED_ORIGINS=https://app.yourbroker.com,https://admin.yourbroker.com,https://api.yourbroker.com
API_RATE_LIMIT=1000
API_RATE_LIMIT_BURST=100

# ============================================
# MONITORING (Optional)
# ============================================
LOG_LEVEL=info
SENTRY_DSN=https://your_sentry_dsn@sentry.io/project_id
DATADOG_API_KEY=your_datadog_api_key
```

**Security Notes**:
- Store all secrets in a secure vault (HashiCorp Vault, AWS Secrets Manager, etc.)
- Never commit .env.production to git
- Rotate credentials quarterly
- Use different credentials per environment

### 3. Setup Systemd Service

**File**: `/etc/systemd/system/trading-engine.service`

```ini
[Unit]
Description=RTX Trading Engine Backend
After=network.target postgresql.service redis.service
Wants=network-online.target

[Service]
Type=simple
User=trading
Group=trading
WorkingDirectory=/opt/trading-engine/backend

# Load environment from production .env
EnvironmentFile=/opt/trading-engine/backend/.env.production

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=yes
PrivateTmp=yes
PrivateDevices=yes

# Resource limits
LimitNOFILE=65535
LimitNPROC=32768

# Start command
ExecStart=/opt/trading-engine/backend/server

# Restart policy
Restart=always
RestartSec=10

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=trading-engine

[Install]
WantedBy=multi-user.target
```

**Enable Service**:
```bash
# Enable auto-start on reboot
systemctl daemon-reload
systemctl enable trading-engine

# Start the service
systemctl start trading-engine

# Check status
systemctl status trading-engine

# View logs
journalctl -u trading-engine -f
```

---

## Production Deployment

### 1. Pre-Deployment Validation

```bash
# Run this checklist before deploying

# 1. Verify all environment variables are set
echo "Checking required environment variables..."
for var in ADMIN_PASSWORD_HASH JWT_SECRET MASTER_ENCRYPTION_KEY \
           YOFX_PROXY_HOST DB_HOST REDIS_HOST; do
    if [ -z "${!var}" ]; then
        echo "ERROR: $var is not set!"
        exit 1
    fi
done
echo "✓ All required variables are set"

# 2. Test database connection
echo "Testing database connection..."
PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -d "$DB_NAME" -h "$DB_HOST" -c "SELECT 1" || exit 1
echo "✓ Database connection successful"

# 3. Test Redis connection
echo "Testing Redis connection..."
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping || exit 1
echo "✓ Redis connection successful"

# 4. Verify binary exists and is executable
if [ ! -x /opt/trading-engine/backend/server ]; then
    echo "ERROR: Binary not found or not executable"
    exit 1
fi
echo "✓ Binary is present and executable"

echo "All pre-deployment checks passed!"
```

### 2. Database Migration

```bash
# If upgrading from a previous version, run migrations
cd /opt/trading-engine/backend

# Backup database first
PGPASSWORD="$DB_PASSWORD" pg_dump -U "$DB_USER" -d "$DB_NAME" -h "$DB_HOST" | \
    gzip > /var/backups/trading-engine/pre-upgrade-$(date +%s).sql.gz

# Run migrations (if any)
# go run migrations/main.go  # Example
```

### 3. Deploy Backend

```bash
# Copy binary to production
cp backend/server /opt/trading-engine/backend/server
chmod 750 /opt/trading-engine/backend/server

# Verify permissions
ls -la /opt/trading-engine/backend/server

# Set correct ownership
chown trading:trading /opt/trading-engine/backend/server

# Deploy .env file securely
cp .env.production /opt/trading-engine/backend/.env.production
chmod 600 /opt/trading-engine/backend/.env.production
chown trading:trading /opt/trading-engine/backend/.env.production
```

### 4. Start Application

```bash
# Start the systemd service
systemctl start trading-engine

# Wait for startup
sleep 5

# Check status
systemctl status trading-engine

# Verify it's listening
netstat -tlnp | grep 7999

# Check logs for startup messages
journalctl -u trading-engine -n 50
```

---

## Post-Deployment Verification

### 1. Health Checks

```bash
# Test HTTP health endpoint
curl -X GET http://localhost:7999/health

# Expected response:
# {
#   "status": "ok",
#   "timestamp": "2026-01-30T12:34:56Z",
#   "checks": {
#     "database": "ok",
#     "redis": "ok",
#     "fix_yofx2": "connected",
#     "websocket": "ok"
#   }
# }
```

### 2. FIX Connection Verification

```bash
# Check YOFX connection status
curl -X GET http://localhost:7999/admin/fix/status \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Expected response:
# {
#   "sessions": {
#     "YOFX2": {
#       "status": "LOGGED_IN",
#       "inboundSeqNum": 12345,
#       "outboundSeqNum": 67890
#     }
#   },
#   "subscribedSymbols": 29,
#   "totalTicks": 15234,
#   "latencyMs": 8.5
# }
```

### 3. Market Data Verification

```bash
# Check if market data is being received
curl -X GET http://localhost:7999/api/diagnostics/market-data \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Expected response should show:
# - Multiple symbols with real YOFX data
# - LP field is always "YOFX"
# - high24h and low24h fields are populated
# - Timestamps are recent
```

### 4. WebSocket Connection Test

```bash
# Create a test file: ws-test.html
cat > ws-test.html << 'EOF'
<!DOCTYPE html>
<html>
<head><title>WebSocket Test</title></head>
<body>
  <h1>WebSocket Market Data Test</h1>
  <div id="messages"></div>
  <script>
    const ws = new WebSocket('ws://localhost:7999/ws');
    ws.onopen = () => console.log('Connected');
    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      console.log('Received:', msg);
      document.getElementById('messages').innerHTML +=
        `<p>${msg.symbol}: ${msg.bid}-${msg.ask} (LP: ${msg.lp})</p>`;
    };
    ws.onerror = (e) => console.error('Error:', e);
  </script>
</body>
</html>
EOF

# Open in browser and verify real YOFX ticks are flowing
```

### 5. Admin Dashboard Access

```bash
# Test admin login
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"YourSecurePassword123!@#"}'

# Expected response:
# {
#   "token": "eyJhbGc...",
#   "expiresIn": 86400
# }
```

---

## Monitoring & Alerting

### 1. Log Monitoring

**File**: `/etc/rsyslog.d/trading-engine.conf`

```bash
# Send trading engine logs to separate file
:programname, isequal, "trading-engine" /var/log/trading-engine/app.log
:programname, isequal, "trading-engine" stop

# Configure logrotate
cat > /etc/logrotate.d/trading-engine << 'EOF'
/var/log/trading-engine/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0644 trading trading
    sharedscripts
    postrotate
        systemctl reload rsyslog > /dev/null 2>&1 || true
    endscript
}
EOF
```

### 2. Key Metrics to Monitor

**Database**:
- Connection pool usage
- Query latency (p50, p95, p99)
- Disk usage growth
- Backup success/failure

**YOFX FIX**:
- Connection status (LOGGED_IN / DISCONNECTED)
- Sequence number gaps (indicates missing messages)
- Latency to YOFX (should be <50ms)
- Tick rate (expected 100-500 ticks/sec)

**API**:
- Request rate (per minute/hour)
- Response time (p50, p95, p99)
- Error rate (5xx, 4xx)
- WebSocket connection count

**System**:
- CPU usage
- Memory usage
- Disk I/O
- Network throughput

### 3. Alerting Rules

```bash
# Alert on high CPU usage (>80% for 5 minutes)
- alert: HighCPUUsage
  expr: cpu_usage > 80
  for: 5m

# Alert on FIX disconnection
- alert: YOFXDisconnected
  expr: fix_session_status == 0
  for: 1m

# Alert on market data lag (no new ticks in 30 seconds)
- alert: NoMarketData
  expr: time() - last_tick_timestamp > 30
  for: 1m

# Alert on database connection errors
- alert: DatabaseConnectionError
  expr: db_connection_errors > 5
  for: 2m

# Alert on high latency to YOFX (>200ms)
- alert: HighLatencyToYOFX
  expr: yofx_latency_ms > 200
  for: 5m
```

---

## Rollback Procedures

### Emergency Rollback (If Critical Issues)

```bash
# 1. Stop the current deployment
systemctl stop trading-engine

# 2. Restore database from backup
LATEST_BACKUP=$(ls -t /var/backups/trading-engine/*.sql.gz | head -1)
PGPASSWORD="$DB_PASSWORD" pg_restore -U "$DB_USER" -d "$DB_NAME" -h "$DB_HOST" \
    --clean --if-exists < <(gunzip -c "$LATEST_BACKUP")

# 3. Restore previous binary
cp /opt/trading-engine/backend/server.backup /opt/trading-engine/backend/server
chmod 750 /opt/trading-engine/backend/server

# 4. Start with previous .env
cp /opt/trading-engine/backend/.env.production.backup \
   /opt/trading-engine/backend/.env.production

# 5. Start service
systemctl start trading-engine

# 6. Verify
systemctl status trading-engine
curl http://localhost:7999/health
```

### Planned Rollback

```bash
# For planned rollback during deployment window:

# 1. Create database checkpoint
PGPASSWORD="$DB_PASSWORD" pg_dump -U "$DB_USER" -d "$DB_NAME" -h "$DB_HOST" | \
    gzip > /var/backups/trading-engine/pre-deployment-$(date +%s).sql.gz

# 2. Keep previous binary
cp /opt/trading-engine/backend/server /opt/trading-engine/backend/server.v1

# 3. Keep previous .env
cp /opt/trading-engine/backend/.env.production \
   /opt/trading-engine/backend/.env.production.v1

# 4. Deploy new version
# ... deployment steps ...

# 5. If needed, restore to v1
cp /opt/trading-engine/backend/server.v1 /opt/trading-engine/backend/server
systemctl restart trading-engine
```

---

## Troubleshooting

### Issue: "YOFX session: DISCONNECTED"

**Symptoms**: FIX gateway not connected to YOFX

**Diagnosis**:
```bash
# 1. Check configuration
grep YOFX /opt/trading-engine/backend/.env.production

# 2. Test connectivity
telnet $YOFX_PROXY_HOST 2605

# 3. Check firewall
ufw status | grep 2605

# 4. Review logs
journalctl -u trading-engine -f | grep -i "yofx\|fix"
```

**Solution**:
```bash
# 1. Verify YOFX credentials with provider
# 2. Check firewall allows outbound on port 2605
# 3. Restart FIX session
curl -X POST http://localhost:7999/admin/fix/restart \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# 4. If still disconnected, restart service
systemctl restart trading-engine
```

### Issue: "No market data received"

**Symptoms**: YOFX connected but no ticks flowing

**Diagnosis**:
```bash
# 1. Check YOFX connection
curl http://localhost:7999/admin/fix/status

# 2. Check symbol subscriptions
curl http://localhost:7999/api/symbols/subscribed

# 3. Monitor FIX sequence numbers (should be increasing)
watch -n 1 'curl -s http://localhost:7999/admin/fix/status | jq ".sessions"'
```

**Solution**:
```bash
# 1. Verify symbols are enabled
curl -X POST http://localhost:7999/admin/symbol/enable \
  -H "Content-Type: application/json" \
  -d '{"symbol":"EURUSD"}'

# 2. Refresh symbol list
curl -X POST http://localhost:7999/admin/fix/refresh-symbols

# 3. Check YOFX server status with provider
# 4. Restart YOFX session if needed
systemctl restart trading-engine
```

### Issue: "Database connection failed"

**Symptoms**: PostgreSQL cannot be reached

**Diagnosis**:
```bash
# 1. Test connection
PGPASSWORD="$DB_PASSWORD" psql -U "$DB_USER" -d "$DB_NAME" \
    -h "$DB_HOST" -c "SELECT 1"

# 2. Check network connectivity
telnet $DB_HOST 5432

# 3. Verify credentials
grep DB_ /opt/trading-engine/backend/.env.production
```

**Solution**:
```bash
# 1. Verify database is running
systemctl status postgresql

# 2. Start if stopped
systemctl start postgresql

# 3. Check credentials in .env
# 4. Test connection again
# 5. Restart application
systemctl restart trading-engine
```

### Issue: "Memory usage continuously increasing"

**Symptoms**: Out of memory crashes

**Diagnosis**:
```bash
# 1. Monitor memory usage
watch -n 1 'free -h && ps aux | grep server'

# 2. Check garbage collection settings
echo $GOGC  # Should be 50
echo $GOMEMLIMIT  # Should be 2GiB

# 3. Check for memory leaks in logs
journalctl -u trading-engine | grep -i "memory\|gc\|alloc"
```

**Solution**:
```bash
# 1. Increase GOMEMLIMIT in .env
GOMEMLIMIT=4GiB  # Increase if you have more RAM

# 2. Reduce tick storage
MAX_TICKS_PER_SYMBOL=25000  # Reduce from 50000

# 3. Restart service
systemctl restart trading-engine

# 4. Contact development team if issue persists
```

---

## Document Information

- **Version**: 1.0
- **Date**: January 30, 2026
- **Status**: ✅ Production Ready
- **Next Review**: 30 days post-deployment

---

## Support & Escalation

**For issues during deployment**:
1. Check logs: `journalctl -u trading-engine -f`
2. Run health check: `curl http://localhost:7999/health`
3. Review troubleshooting section above
4. Contact development team with logs attached
5. Escalate to on-call engineer if critical

