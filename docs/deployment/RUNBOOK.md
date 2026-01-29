# Operations Runbook - Trading Engine

## Emergency Contacts

| Role | Contact | On-Call Schedule |
|------|---------|------------------|
| DevOps Lead | devops@trading-engine.com | 24/7 |
| Backend Lead | backend@trading-engine.com | Business hours |
| DBA | dba@trading-engine.com | 24/7 |
| Security | security@trading-engine.com | 24/7 |

## Quick Reference

### Service URLs (Production)

- API: https://api.trading-engine.com
- WebSocket: wss://ws.trading-engine.com
- Admin Panel: https://admin.trading-engine.com
- Grafana: https://grafana.trading-engine.com
- Prometheus: https://prometheus.trading-engine.com (Internal)
- Jaeger: https://jaeger.trading-engine.com (Internal)

### Critical Metrics Dashboard

```bash
# Access Grafana
https://grafana.trading-engine.com/d/trading-engine-overview

# Key metrics to watch:
- Order execution latency (p95 < 300ms)
- Error rate (< 0.1%)
- Active WebSocket connections
- Database connection pool utilization (< 80%)
- CPU/Memory usage per pod
- FIX session status
```

## Incident Response Procedures

### Severity Levels

- **P0 (Critical)**: Complete service outage, data loss, security breach
- **P1 (High)**: Significant functionality impaired, affects > 50% users
- **P2 (Medium)**: Partial functionality impaired, affects < 50% users
- **P3 (Low)**: Minor issues, no user impact

### P0: Complete Service Outage

**Symptoms:**
- Health check failing
- All API requests returning errors
- Database unreachable
- Zero WebSocket connections

**Immediate Actions:**

```bash
# 1. Check service status
kubectl get pods -n production
kubectl get services -n production
kubectl get ingress -n production

# 2. Check recent deployments
helm history trading-engine -n production

# 3. Rollback if recent deployment
helm rollback trading-engine -n production

# 4. Check infrastructure
kubectl get nodes
kubectl top nodes

# 5. Check database
kubectl logs -n production statefulset/trading-engine-postgresql --tail=100

# 6. Emergency notification
# Post to #incidents Slack channel
# Page on-call engineer
```

**Root Cause Investigation:**

```bash
# Application logs
kubectl logs -n production deployment/trading-engine-api --tail=500

# Database logs
kubectl logs -n production statefulset/trading-engine-postgresql --tail=500

# Ingress logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller --tail=500

# System events
kubectl get events -n production --sort-by='.lastTimestamp'
```

### P1: High Error Rate

**Symptoms:**
- Error rate > 5%
- Slow response times (p95 > 1s)
- Increased timeout errors

**Diagnostic Steps:**

```bash
# 1. Check error logs
kubectl logs -n production deployment/trading-engine-api \
  --tail=1000 | grep -i error

# 2. Check metrics
curl https://api.trading-engine.com/metrics | grep http_requests

# 3. Database performance
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT * FROM pg_stat_activity WHERE state = 'active';"

# 4. Check slow queries
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

**Mitigation:**

```bash
# Scale up API pods
kubectl scale deployment trading-engine-api -n production --replicas=10

# Restart problematic pods
kubectl rollout restart deployment/trading-engine-api -n production

# Clear Redis cache if stale
kubectl exec -it -n production statefulset/trading-engine-redis-0 -- redis-cli FLUSHALL
```

### P1: Database Performance Issues

**Symptoms:**
- Slow query execution
- Connection pool exhausted
- High CPU on database

**Diagnostic Steps:**

```bash
# Check active connections
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT count(*) FROM pg_stat_activity;"

# Check locks
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT * FROM pg_locks WHERE granted = false;"

# Check slow queries
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC;"
```

**Mitigation:**

```bash
# Kill long-running queries (carefully!)
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '5 minutes';"

# Vacuum analyze
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "VACUUM ANALYZE;"

# Increase connection pool (temporary)
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set postgresql.primary.resources.limits.cpu=4000m
```

### P1: Memory Leak

**Symptoms:**
- Increasing memory usage over time
- OOMKilled pods
- Pod restarts

**Diagnostic Steps:**

```bash
# Check memory usage
kubectl top pods -n production

# Check pod restarts
kubectl get pods -n production -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.containerStatuses[0].restartCount}{"\n"}{end}'

# Get heap dump (Go)
kubectl exec -it -n production <pod-name> -- curl http://localhost:6060/debug/pprof/heap > heap.prof
```

**Mitigation:**

```bash
# Increase memory limits
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set api.resources.limits.memory=4Gi

# Enable memory profiling
kubectl set env deployment/trading-engine-api -n production \
  GODEBUG=gctrace=1

# Restart affected pods
kubectl rollout restart deployment/trading-engine-api -n production
```

### P0: Security Breach

**Symptoms:**
- Unauthorized access detected
- Unusual API activity
- Data exfiltration alerts

**Immediate Actions:**

```bash
# 1. ISOLATE - Block external traffic
kubectl patch ingress trading-engine-ingress -n production \
  -p '{"metadata":{"annotations":{"nginx.ingress.kubernetes.io/whitelist-source-range":"10.0.0.0/8"}}}'

# 2. ASSESS - Check logs
kubectl logs -n production deployment/trading-engine-api \
  --since=1h | grep -E "401|403|suspicious"

# 3. PRESERVE - Export logs
kubectl logs -n production deployment/trading-engine-api --all-containers=true \
  > incident-logs-$(date +%Y%m%d-%H%M%S).log

# 4. NOTIFY
# - Security team
# - Management
# - Legal (if data breach)

# 5. ROTATE SECRETS
kubectl delete secret trading-engine-secrets -n production
kubectl create secret generic trading-engine-secrets \
  --from-literal=jwt-secret='NEW_SECRET' \
  ... # (all secrets)

# 6. Force pod restart
kubectl rollout restart deployment/trading-engine-api -n production
```

## Routine Operations

### Daily Checks

```bash
# Health check
curl https://api.trading-engine.com/health

# Check pod health
kubectl get pods -n production

# Check HPA status
kubectl get hpa -n production

# Review Grafana dashboards
# - System overview
# - Trading metrics
# - Error rates

# Check alerts
# - Prometheus alerts
# - PagerDuty incidents
```

### Weekly Tasks

```bash
# Review slow queries
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT query, calls, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 20;"

# Database maintenance
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "VACUUM ANALYZE;"

# Check certificate expiry
kubectl get certificates -n production

# Review resource usage trends
# Grafana: Resource Utilization dashboard

# Backup verification
aws s3 ls s3://trading-engine-backups/
```

### Monthly Tasks

```bash
# Security updates
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set image.tag=v1.x.x

# Review and update HPA settings
# Based on traffic patterns

# Database optimization
# Review indexes, update statistics

# Cost optimization review
# Check resource utilization vs allocation

# Disaster recovery drill
# Test backup restoration

# Security audit
# Review access logs, update RBAC
```

## Scaling Operations

### Scale Up (Anticipated Load)

```bash
# Before major event/announcement

# 1. Increase API replicas
kubectl scale deployment trading-engine-api -n production --replicas=20

# 2. Increase WebSocket replicas
kubectl scale deployment trading-engine-websocket -n production --replicas=30

# 3. Increase database resources
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set postgresql.primary.resources.limits.cpu=8000m \
  --set postgresql.primary.resources.limits.memory=16Gi

# 4. Pre-warm cache
# Populate Redis with frequently accessed data

# 5. Monitor closely
watch kubectl get hpa -n production
```

### Scale Down (Off-Peak Hours)

```bash
# Reduce to minimum replicas

kubectl scale deployment trading-engine-api -n production --replicas=3
kubectl scale deployment trading-engine-websocket -n production --replicas=5

# Note: HPA will override manual scaling
# Better to adjust HPA min replicas:
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set api.autoscaling.minReplicas=3
```

## Deployment Procedures

### Standard Deployment (Zero Downtime)

```bash
# 1. Pre-deployment checks
kubectl get pods -n production  # All healthy
helm test trading-engine -n production  # Pass tests

# 2. Deploy new version
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set image.tag=v1.2.0 \
  --wait \
  --timeout 10m

# 3. Monitor rollout
kubectl rollout status deployment/trading-engine-api -n production

# 4. Smoke test
curl https://api.trading-engine.com/health
# Run load test at low volume

# 5. Monitor metrics for 30 minutes
# Check error rates, latency, throughput

# 6. If issues: rollback
helm rollback trading-engine -n production
```

### Emergency Hotfix

```bash
# 1. Build and push hotfix image
docker build -t ghcr.io/epic1st/trading-engine-api:v1.2.1-hotfix .
docker push ghcr.io/epic1st/trading-engine-api:v1.2.1-hotfix

# 2. Deploy immediately
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set image.tag=v1.2.1-hotfix \
  --wait

# 3. Monitor
kubectl logs -f -n production deployment/trading-engine-api

# 4. Verify fix
# Test affected functionality
```

## Database Operations

### Manual Backup

```bash
# Full backup
kubectl exec -n production statefulset/trading-engine-postgresql-0 -- \
  pg_dump -U trading -Fc trading_engine > backup-$(date +%Y%m%d-%H%M%S).dump

# Upload to S3
aws s3 cp backup-*.dump s3://trading-engine-backups/manual/
```

### Restore from Backup

```bash
# Download backup
aws s3 cp s3://trading-engine-backups/backup-20260118.dump .

# Scale down application
kubectl scale deployment trading-engine-api -n production --replicas=0

# Restore
kubectl exec -i -n production statefulset/trading-engine-postgresql-0 -- \
  pg_restore -U trading -d trading_engine -c < backup-20260118.dump

# Scale up application
kubectl scale deployment trading-engine-api -n production --replicas=3
```

### Database Migration

```bash
# Backup before migration
# Run backup procedure above

# Apply migration
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -f /migrations/001_new_schema.sql

# Verify
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "\dt"
```

## Monitoring & Alerting

### Key Metrics to Watch

| Metric | Threshold | Action |
|--------|-----------|--------|
| Error rate | > 1% | Investigate logs |
| p95 latency | > 500ms | Check database, scale up |
| CPU usage | > 80% | Scale horizontally |
| Memory usage | > 85% | Check for leaks, scale vertically |
| Database connections | > 80% of max | Investigate connection leaks |
| Disk usage | > 85% | Clean logs, expand storage |
| WebSocket connections | > 50K | Scale WebSocket pods |

### Alert Response Times

- **P0**: Immediate (< 5 minutes)
- **P1**: 15 minutes
- **P2**: 1 hour
- **P3**: Next business day

## Post-Incident Review

### Template

```markdown
# Incident Report: [Brief Description]

**Date:** 2026-01-18
**Duration:** X hours
**Severity:** PX
**Impact:** X users affected

## Timeline

- **HH:MM** - Initial detection
- **HH:MM** - Investigation started
- **HH:MM** - Root cause identified
- **HH:MM** - Fix applied
- **HH:MM** - Service restored
- **HH:MM** - Incident closed

## Root Cause

[Technical explanation]

## Impact

- Users affected: X
- Downtime: X minutes
- Lost transactions: X
- Revenue impact: $X

## Resolution

[What was done to fix]

## Prevention

- [ ] Action item 1
- [ ] Action item 2
- [ ] Code changes required
- [ ] Documentation updates

## Lessons Learned

- What went well
- What could be improved
- Process improvements
```

## Contact Information

### External Services

- AWS Support: https://console.aws.amazon.com/support
- GitHub Support: support@github.com
- OANDA Support: api@oanda.com
- Binance Support: support@binance.com

### Internal Documentation

- Architecture: `/docs/architecture/`
- API Documentation: `/docs/api/`
- Database Schema: `/docs/database/`
- Security: `/docs/security/`
