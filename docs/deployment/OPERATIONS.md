# Operations Runbook

This runbook provides procedures for operating, troubleshooting, and maintaining the Trading Engine platform in production.

## Table of Contents

- [Health Checks](#health-checks)
- [Common Issues](#common-issues)
- [Backup and Recovery](#backup-and-recovery)
- [Performance Optimization](#performance-optimization)
- [Scaling Considerations](#scaling-considerations)
- [Security Checklist](#security-checklist)
- [Maintenance Procedures](#maintenance-procedures)
- [Monitoring and Alerts](#monitoring-and-alerts)
- [Incident Response](#incident-response)

## Health Checks

### Liveness Probe

Checks if the backend process is alive:

```bash
curl http://localhost:8080/health/live
```

**Expected response:**
```json
{"status": "ok"}
```

**HTTP 200**: Process is running
**No response**: Process is down, restart required

**Use case:** Kubernetes liveness probe, process monitoring

### Readiness Probe

Checks if the backend is ready to accept traffic:

```bash
curl http://localhost:8080/health/ready
```

**Expected response:**
```json
{"status": "ready"}
```

**HTTP 200**: All dependencies healthy
**HTTP 503**: One or more dependencies unavailable

**Use case:** Kubernetes readiness probe, load balancer health checks

### Detailed Health Status

Get detailed component health:

```bash
curl http://localhost:8080/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-01-16T10:30:45Z",
  "checks": {
    "database": "ok",
    "redis": "ok",
    "oanda": "ok"
  },
  "uptime_seconds": 86400
}
```

**Component status meanings:**
- `ok`: Component is healthy
- `degraded`: Component has issues but is operational
- `down`: Component is unavailable

## Common Issues

### Issue: High Order Latency

**Symptoms:**
- Order processing takes >100ms (P95)
- Slow trade execution
- Timeout errors in frontend

**Check:**
```promql
histogram_quantile(0.95, rate(trading_order_duration_seconds_bucket[5m]))
```

**Possible causes:**
1. Database slow queries
2. Redis cache misses
3. External API (OANDA) latency
4. High CPU/memory usage

**Troubleshooting:**
```bash
# Check database slow queries
docker-compose exec db psql -U postgres -d trading_engine
SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;

# Check Redis hit rate
docker-compose exec redis redis-cli INFO stats | grep keyspace

# Check CPU and memory
docker stats

# View backend logs for slow operations
docker-compose logs backend | jq 'select(.duration_ms > 100)'
```

**Solutions:**
1. **Database**: Add indexes to frequently queried columns
2. **Redis**: Increase cache TTL, add more cache keys
3. **API**: Implement retry logic with exponential backoff
4. **Resources**: Scale vertically (more CPU/RAM) or horizontally (more instances)

### Issue: Database Connection Errors

**Symptoms:**
- `/health/ready` returns 503
- Backend logs show "connection refused" or "too many clients"
- Orders fail to persist

**Check:**
```bash
# Test database connectivity
docker-compose exec db pg_isready

# Check active connections
docker-compose exec db psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

# View connection pool status in logs
docker-compose logs backend | jq 'select(.msg | contains("database"))'
```

**Possible causes:**
1. PostgreSQL not running
2. Connection pool exhausted
3. Network issues between backend and database
4. Database credentials incorrect

**Solutions:**

**PostgreSQL not running:**
```bash
docker-compose up -d db
docker-compose logs db
```

**Connection pool exhausted:**
```bash
# Increase max_connections in PostgreSQL
docker-compose exec db psql -U postgres -c "ALTER SYSTEM SET max_connections = 100;"
docker-compose restart db

# Or increase connection pool in backend (backend/internal/database/pool.go)
# Recommended: (CPU cores * 2) + 1
```

**Network issues:**
```bash
# Verify services can communicate
docker-compose exec backend ping db
docker-compose exec backend nc -zv db 5432
```

**Credentials incorrect:**
```bash
# Verify DATABASE_URL in .env
cat .env | grep DATABASE_URL

# Test connection manually
docker-compose exec db psql $DATABASE_URL
```

### Issue: Redis Cache Not Working

**Symptoms:**
- High database load
- Slow OHLC data retrieval
- Redis connection errors in logs

**Check:**
```bash
# Test Redis connectivity
docker-compose exec redis redis-cli PING

# View Redis logs
docker-compose logs redis

# Check backend Redis errors
docker-compose logs backend | jq 'select(.error | contains("redis"))'

# Check cache keys
docker-compose exec redis redis-cli KEYS "*"
```

**Possible causes:**
1. Redis not running
2. Redis out of memory
3. Network connectivity issues
4. Wrong Redis URL in configuration

**Solutions:**

**Redis not running:**
```bash
docker-compose up -d redis
docker-compose logs redis
```

**Out of memory:**
```bash
# Check memory usage
docker-compose exec redis redis-cli INFO memory

# Increase maxmemory or enable eviction
docker-compose exec redis redis-cli CONFIG SET maxmemory 256mb
docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

**Network issues:**
```bash
# Test connectivity
docker-compose exec backend nc -zv redis 6379

# Verify REDIS_URL
cat .env | grep REDIS_URL
```

**Clear cache (if corrupted):**
```bash
docker-compose exec redis redis-cli FLUSHALL
```

### Issue: Frontend 502 Bad Gateway

**Symptoms:**
- Frontend loads but API calls fail
- nginx returns 502 error
- "Failed to fetch" errors in browser console

**Check:**
```bash
# Test backend health
curl http://localhost:8080/health

# Check nginx logs
docker-compose logs frontend

# Verify backend is running
docker-compose ps backend
```

**Possible causes:**
1. Backend not running
2. Backend crashed or restarting
3. nginx misconfiguration
4. Wrong API URL in frontend build

**Solutions:**

**Backend not running:**
```bash
docker-compose up -d backend
docker-compose logs backend
```

**Backend crashed:**
```bash
# Check crash logs
docker-compose logs backend | tail -100

# Look for panic/fatal errors
docker-compose logs backend | jq 'select(.level=="fatal" or .level=="panic")'

# Restart backend
docker-compose restart backend
```

**nginx misconfiguration:**
```bash
# Test nginx config
docker-compose exec frontend nginx -t

# Reload nginx
docker-compose exec frontend nginx -s reload
```

**Wrong API URL:**
```bash
# Rebuild frontend with correct URL
docker-compose build --build-arg VITE_API_URL=http://localhost:8080 frontend
docker-compose up -d frontend
```

### Issue: Order Validation Failures

**Symptoms:**
- High order rejection rate
- "Insufficient margin" errors
- "Invalid symbol" errors

**Check:**
```promql
# Order rejection rate
rate(trading_orders_processed_total{status="rejected"}[5m])
  / rate(trading_orders_processed_total[5m])
```

**Troubleshooting:**
```bash
# View rejected orders in logs
docker-compose logs backend | jq 'select(.status=="rejected")'

# Check validation rules
docker-compose logs backend | jq 'select(.msg | contains("validation"))'

# Verify account balances
docker-compose exec db psql -U postgres -d trading_engine -c "SELECT id, balance, margin FROM accounts;"
```

**Solutions:**
1. **Insufficient margin**: Increase account balance or reduce position size
2. **Invalid symbol**: Check symbol is configured in LP config
3. **Invalid quantity**: Verify quantity meets minimum/maximum requirements
4. **Market hours**: Ensure trading during active market hours

### Issue: Position Not Closing

**Symptoms:**
- Close order submitted but position remains open
- Position stuck in "closing" state
- Margin not released after close

**Check:**
```bash
# View open positions
docker-compose exec db psql -U postgres -d trading_engine -c "SELECT * FROM positions WHERE status = 'open';"

# Check close order status
docker-compose exec db psql -U postgres -d trading_engine -c "SELECT * FROM orders WHERE type = 'close' ORDER BY created_at DESC LIMIT 10;"

# View position closing logs
docker-compose logs backend | jq 'select(.msg | contains("close"))'
```

**Solutions:**
1. **Retry close**: Submit new close order
2. **Manual intervention**: Update position status in database (emergency only)
3. **Check external LP**: Verify OANDA API is accessible and responding

### Issue: Database Disk Full

**Symptoms:**
- Write operations fail
- "no space left on device" errors
- Database crashes

**Check:**
```bash
# Check disk usage
df -h

# Check PostgreSQL data directory
docker-compose exec db du -sh /var/lib/postgresql/data

# Check Docker volumes
docker system df -v
```

**Solutions:**

**Immediate (emergency):**
```bash
# Delete old backups
rm /backups/postgres/trading_engine_*.sql.gz

# Clean Docker system
docker system prune -a

# Truncate audit log (if very large)
docker-compose exec db psql -U postgres -d trading_engine -c "TRUNCATE audit_log;"
```

**Long-term:**
1. Implement log rotation for audit_log table
2. Increase disk size
3. Archive old trades to separate storage
4. Enable PostgreSQL autovacuum to reclaim space

### Issue: High Memory Usage

**Symptoms:**
- Backend OOM (out of memory) kills
- Slow performance
- Containers restarting frequently

**Check:**
```bash
# Check container memory usage
docker stats

# Check Go heap usage
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# View memory metrics
curl http://localhost:8080/metrics | grep go_memstats
```

**Solutions:**
1. **Memory leak**: Profile with pprof, identify leaking goroutines/objects
2. **Cache too large**: Reduce Redis maxmemory or cache TTL
3. **Too many goroutines**: Check for goroutine leaks with `/debug/pprof/goroutine`
4. **Increase resources**: Allocate more memory to container

```bash
# Increase memory limit in docker-compose.yml
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 1G
```

## Backup and Recovery

### Manual Backup

Create an on-demand backup:

```bash
# Run backup script
docker-compose exec backend bash /app/scripts/backup-db.sh

# Or trigger GitHub Actions workflow
gh workflow run backup.yml
```

**Backup location:**
- Local: `/backups/postgres/`
- GitHub: Actions artifacts (30-day retention)

### List Available Backups

```bash
# Local backups
ls -lh /backups/postgres/

# GitHub Actions artifacts
gh run list --workflow=backup.yml
gh run download RUN_ID
```

### Restore from Backup

**Prerequisites:**
- Backup file exists
- Database is running
- All applications disconnected from database

**Restore procedure:**

```bash
# 1. Stop backend to prevent new writes
docker-compose stop backend

# 2. Decompress and restore
gunzip < /backups/postgres/trading_engine_20260116_120000.sql.gz | \
  docker-compose exec -T db psql -U postgres -d trading_engine

# 3. Verify data
docker-compose exec db psql -U postgres -d trading_engine -c "SELECT COUNT(*) FROM accounts;"

# 4. Restart backend
docker-compose start backend
```

**Restore from GitHub artifact:**
```bash
# Download artifact
gh run download RUN_ID -n postgres-backup

# Extract and restore
tar -xzf postgres-backup.tar.gz
gunzip < trading_engine_*.sql.gz | \
  docker-compose exec -T db psql -U postgres -d trading_engine
```

### Point-in-Time Recovery

For production, enable PostgreSQL WAL archiving:

```sql
-- postgresql.conf
archive_mode = on
archive_command = 'cp %p /backups/postgres/archive/%f'
wal_level = replica
```

Then use `pg_basebackup` and WAL replay for PITR.

## Performance Optimization

### LP Manager Optimization

The LP (Liquidity Provider) manager uses map-based lookups for O(1) performance:

**Implementation** (from Plan 04-06):
- Direct map access: `lpConfigMap[symbol]`
- No iteration required
- Constant-time lookups regardless of LP count

**Verify optimization:**
```bash
# Should show map initialization
docker-compose logs backend | jq 'select(.msg | contains("LP config"))'
```

### Tick Data Caching

Tick data is cached in Redis with 60-second TTL:

**Check cache performance:**
```bash
# View cache hit rate
docker-compose exec redis redis-cli INFO stats | grep keyspace

# Expected hit rate: >90%
# hits / (hits + misses) > 0.9
```

**Optimize:**
```bash
# Increase TTL for less volatile pairs (modify in code)
# Reduce TTL for highly volatile pairs
# Increase Redis maxmemory if evictions are high
```

### OHLC Data Caching

OHLC data cached with variable TTL based on timeframe:

- M1: 1 hour TTL
- M5/M15: 4 hours TTL
- H1/H4: 12 hours TTL
- D1: 7 days TTL

**Monitor:**
```bash
# Check OHLC cache keys
docker-compose exec redis redis-cli KEYS "ohlc:*"

# View specific cached data
docker-compose exec redis redis-cli GET "ohlc:BTCUSD:H1"
```

### Database Indexes

Ensure critical queries use indexes:

```sql
-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Analyze slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Add index for common query
CREATE INDEX idx_positions_symbol_status ON positions(symbol, status);
CREATE INDEX idx_orders_account_created ON orders(account_id, created_at);
CREATE INDEX idx_trades_position_executed ON trades(position_id, executed_at);
```

### Connection Pool Tuning

**Current settings** (from Plan 02-01):
- Max connections: 20
- Min connections: 5
- Formula: (CPU cores * 2) + 1

**Adjust for your environment:**
```go
// backend/internal/database/pool.go
poolConfig.MaxConns = 20          // Increase for high load
poolConfig.MinConns = 5           // Increase for consistent load
poolConfig.MaxConnLifetime = 1h   // Recycle connections
poolConfig.MaxConnIdleTime = 30m  // Close idle connections
```

## Scaling Considerations

### Horizontal Scaling

Running multiple backend instances:

**Requirements:**
- Load balancer (nginx, HAProxy, Kubernetes Service)
- Shared PostgreSQL database
- Shared Redis instance

**Example with docker-compose:**
```bash
# Scale backend to 3 instances
docker-compose up -d --scale backend=3

# Add nginx load balancer
# Update docker-compose.yml with nginx upstream configuration
```

**Considerations:**
- WebSocket connections require sticky sessions
- Ensure database can handle connection pool * instances
- Redis connection pooling across instances

### Vertical Scaling

Increasing resources for single instance:

**CPU:**
```yaml
# docker-compose.yml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2.0'
        reservations:
          cpus: '1.0'
```

**Memory:**
```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G
```

**Database connection pool:**
Increase based on new CPU count: `(CPUs * 2) + 1`

### Database Scaling

**Read replicas:**
- PostgreSQL streaming replication
- Route read queries to replicas
- Write queries to primary

**Connection pooling:**
- PgBouncer for connection pooling
- Reduces connection overhead
- Enables more connections with fewer resources

**Partitioning:**
- Partition trades table by date
- Archive old partitions to separate storage
- Improves query performance on recent data

### Redis Scaling

**Redis Cluster:**
- Sharding for larger datasets
- Multiple master nodes
- Automatic failover

**Redis Sentinel:**
- High availability
- Automatic failover
- Master election

## Security Checklist

Pre-deployment security verification:

### Environment Variables
- [ ] Change default passwords in .env (PostgreSQL, Grafana)
- [ ] Generate new JWT_SECRET (minimum 32 bytes)
- [ ] Use secrets management system (not .env files)
- [ ] Rotate OANDA API keys regularly

### Database Security
- [ ] Enable TLS for PostgreSQL connections
- [ ] Restrict database access to backend only
- [ ] Use strong passwords (minimum 16 characters)
- [ ] Enable audit logging for sensitive operations
- [ ] Regular backup testing (monthly)

### Redis Security
- [ ] Enable Redis authentication (requirepass)
- [ ] Restrict Redis access to backend network only
- [ ] Disable dangerous commands (FLUSHALL, CONFIG, etc.)
- [ ] Use Redis ACLs for fine-grained permissions

### Container Security
- [ ] Run containers as non-root user (already configured)
- [ ] Keep base images updated (monthly rebuilds)
- [ ] Scan images for vulnerabilities (Trivy, Snyk)
- [ ] Use distroless images for minimal attack surface
- [ ] Remove development tools from production images

### Network Security
- [ ] Use TLS for all external communications
- [ ] Restrict port access (firewall rules)
- [ ] Use VPN or private network for database access
- [ ] Enable CORS only for trusted origins
- [ ] Rate limiting on API endpoints

### Application Security
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS protection (Content-Security-Policy headers)
- [ ] CSRF protection for state-changing operations
- [ ] JWT token expiration (reasonable TTL)

### Monitoring & Logging
- [ ] Enable security audit logging
- [ ] Monitor failed login attempts
- [ ] Alert on unusual trading activity
- [ ] Log all administrative actions
- [ ] Retain logs for compliance (90+ days)

## Maintenance Procedures

### Weekly Maintenance

**Monday morning routine:**
```bash
# Check system health
curl http://localhost:8080/health

# Review error logs from past week
docker-compose logs --since 168h backend | jq 'select(.level=="error")' | wc -l

# Check backup status
ls -lh /backups/postgres/ | tail -20

# Database vacuum (if autovacuum disabled)
docker-compose exec db psql -U postgres -d trading_engine -c "VACUUM ANALYZE;"

# Check disk usage
df -h
docker system df
```

### Monthly Maintenance

**First of month routine:**
```bash
# Update Docker images
docker-compose pull
docker-compose up -d

# Rebuild application images
docker-compose build
docker-compose up -d backend frontend

# Archive old logs (>30 days)
find /var/log/trading-engine -mtime +30 -delete

# Test backup restore procedure
# (Run in staging environment, not production!)

# Review performance metrics
# Check Grafana dashboards for trends
# Analyze slow query log
# Review Redis eviction rate

# Security updates
docker scan trading-engine-backend:latest
docker scan trading-engine-frontend:latest

# Rotate JWT secrets
# Generate new secret, update .env, restart backend
```

### Quarterly Maintenance

**Every 3 months:**
- Review and update dependencies (Go modules, npm packages)
- Conduct security audit
- Review and optimize database indexes
- Archive historical trade data (>1 year old)
- Disaster recovery drill (full restore test)
- Review and update runbook (this document!)
- Performance testing and benchmarking
- Capacity planning review

## Monitoring and Alerts

### Recommended Alert Rules

Configure these alerts in Prometheus:

**Critical Alerts (page immediately):**
```yaml
- alert: BackendDown
  expr: up{job="trading-engine-backend"} == 0
  for: 1m

- alert: DatabaseDown
  expr: up{job="postgres"} == 0
  for: 1m

- alert: HighOrderRejectionRate
  expr: rate(trading_orders_processed_total{status="rejected"}[5m]) / rate(trading_orders_processed_total[5m]) > 0.5
  for: 5m

- alert: DiskSpaceCritical
  expr: node_filesystem_avail_bytes / node_filesystem_size_bytes < 0.1
  for: 5m
```

**Warning Alerts (investigate soon):**
```yaml
- alert: HighOrderLatency
  expr: histogram_quantile(0.95, rate(trading_order_duration_seconds_bucket[5m])) > 0.1
  for: 10m

- alert: HighMemoryUsage
  expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.8
  for: 10m

- alert: BackupFailed
  expr: time() - last_backup_timestamp_seconds > 21600  # 6 hours
  for: 1m

- alert: RedisEvictionHigh
  expr: rate(redis_evicted_keys_total[5m]) > 10
  for: 10m
```

### Alert Destinations

Configure Alertmanager for:
- **Critical**: PagerDuty, phone call
- **Warning**: Slack, email
- **Info**: Email, Slack

## Incident Response

### Severity Levels

**SEV1 (Critical):**
- Platform completely down
- Data loss risk
- Security breach
- Response time: Immediate (page on-call)

**SEV2 (High):**
- Degraded performance
- Partial outage
- Failed backups
- Response time: <30 minutes

**SEV3 (Medium):**
- Minor issues
- Non-critical component down
- High error rate (not blocking)
- Response time: <4 hours

**SEV4 (Low):**
- Cosmetic issues
- Documentation updates
- Response time: Next business day

### Incident Response Template

**When incident occurs:**

1. **Assess**: Determine severity level
2. **Communicate**: Post in incident channel
3. **Investigate**: Follow runbook troubleshooting steps
4. **Mitigate**: Apply temporary fix if needed
5. **Resolve**: Apply permanent fix
6. **Document**: Create post-mortem
7. **Follow-up**: Implement preventive measures

**Post-mortem template:**
```markdown
## Incident: [Title]
Date: 2026-01-16
Severity: SEV2
Duration: 45 minutes

### Timeline
- 10:00 - Alert triggered: High order latency
- 10:05 - Investigation started
- 10:15 - Root cause identified: Database connection pool exhausted
- 10:20 - Mitigation: Restarted backend, increased connection pool
- 10:45 - Incident resolved

### Root Cause
Database connection pool size (20) was insufficient for current load (30 req/s).

### Resolution
- Increased connection pool to 50
- Added monitoring alert for connection pool usage
- Documented scaling procedure in runbook

### Action Items
- [ ] Load test to determine optimal pool size
- [ ] Implement connection pool metrics
- [ ] Add auto-scaling for backend instances
```

## Emergency Contacts

Maintain a list of contacts for production issues:

- **Platform Owner**: [Name, Phone, Email]
- **Database Admin**: [Name, Phone, Email]
- **Security Team**: [Name, Phone, Email]
- **OANDA Support**: [Support email, Phone]
- **Infrastructure Team**: [Name, Phone, Email]

## Related Documentation

- [Docker Deployment Guide](./DOCKER.md)
- [Local Development Guide](./LOCAL_DEV.md)
- [Monitoring Guide](./MONITORING.md)
- [CI/CD Pipeline](./CI_CD.md)
