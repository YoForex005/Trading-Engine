# RTX Trading Engine - Deployment Quick Start

**Version:** 1.0.0
**For:** DevOps Engineers & SREs
**Prerequisites:** Docker, kubectl, AWS CLI (or equivalent cloud CLI)

---

## Quick Start (3 Steps)

### 1. Local Development (Docker Compose)

```bash
# Clone repository
git clone <repository-url>
cd backend

# Copy environment file
cp .env.example .env
# Edit .env with your credentials

# Start services
docker compose up -d

# Check health
curl http://localhost:7999/health
# Expected: "OK"

# View logs
docker compose logs -f backend

# Stop services
docker compose down
```

### 2. Kubernetes Development Cluster

```bash
# Create namespace
kubectl create namespace rtx-trading-dev

# Create secrets
kubectl create secret generic rtx-backend-secrets \
  --from-literal=DB_PASSWORD=your_password \
  --from-literal=JWT_SECRET=your_jwt_secret \
  --from-literal=ADMIN_PASSWORD_HASH='$2a$10$...' \
  -n rtx-trading-dev

# Apply ConfigMap
kubectl apply -f k8s/configmap.yaml -n rtx-trading-dev

# Deploy PostgreSQL
kubectl apply -f k8s/postgres-statefulset.yaml -n rtx-trading-dev
kubectl apply -f k8s/postgres-service.yaml -n rtx-trading-dev

# Wait for database
kubectl wait --for=condition=ready pod/postgres-0 -n rtx-trading-dev --timeout=300s

# Deploy Redis
kubectl apply -f k8s/redis-deployment.yaml -n rtx-trading-dev
kubectl apply -f k8s/redis-service.yaml -n rtx-trading-dev

# Deploy Backend
kubectl apply -f k8s/backend-deployment.yaml -n rtx-trading-dev
kubectl apply -f k8s/backend-service.yaml -n rtx-trading-dev

# Check status
kubectl get pods -n rtx-trading-dev
kubectl get services -n rtx-trading-dev

# Port forward
kubectl port-forward svc/rtx-backend-service 7999:80 -n rtx-trading-dev

# Test
curl http://localhost:7999/health
```

### 3. Production Deployment (Blue-Green)

```bash
# Set context to production cluster
kubectl config use-context production

# Backup database
kubectl exec -n rtx-trading postgres-0 -- pg_dump -U postgres trading_engine > backup-$(date +%Y%m%d).sql

# Deploy green environment
sed 's|name: rtx-backend|name: rtx-backend-green|g' k8s/backend-deployment.yaml | kubectl apply -f - -n rtx-trading

# Wait for rollout
kubectl rollout status deployment/rtx-backend-green -n rtx-trading

# Smoke test green
GREEN_POD=$(kubectl get pods -n rtx-trading -l app=rtx-backend,version=green -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n rtx-trading $GREEN_POD -- wget --spider http://localhost:7999/health

# Switch traffic to green
kubectl patch service rtx-backend-service -n rtx-trading -p '{"spec":{"selector":{"version":"green"}}}'

# Monitor for 5 minutes
watch -n 10 "kubectl get pods -n rtx-trading; kubectl top pods -n rtx-trading"

# If all good, remove blue
kubectl delete deployment rtx-backend-blue -n rtx-trading

# If issues, rollback
kubectl patch service rtx-backend-service -n rtx-trading -p '{"spec":{"selector":{"version":"blue"}}}'
```

---

## Common Operations

### View Logs

```bash
# Backend logs
kubectl logs -f deployment/rtx-backend -n rtx-trading

# Tail last 100 lines
kubectl logs --tail=100 deployment/rtx-backend -n rtx-trading

# All containers in pod
kubectl logs -f rtx-backend-xxxxx --all-containers -n rtx-trading

# PostgreSQL logs
kubectl logs -f postgres-0 -n rtx-trading
```

### Database Operations

```bash
# Connect to database
kubectl exec -it postgres-0 -n rtx-trading -- psql -U postgres trading_engine

# Run migration
kubectl run migration-$(date +%s) \
  --image=rtx-backend:latest \
  --restart=Never \
  --env="DB_HOST=postgres-service" \
  -n rtx-trading \
  -- /root/rtx-backend migrate up

# Backup database
kubectl exec postgres-0 -n rtx-trading -- pg_dump -U postgres trading_engine | gzip > backup-$(date +%Y%m%d-%H%M%S).sql.gz

# Restore database
gunzip -c backup.sql.gz | kubectl exec -i postgres-0 -n rtx-trading -- psql -U postgres trading_engine
```

### Scaling

```bash
# Manual scale
kubectl scale deployment rtx-backend --replicas=5 -n rtx-trading

# View HPA status
kubectl get hpa -n rtx-trading

# Edit HPA
kubectl edit hpa rtx-backend-hpa -n rtx-trading
```

### Rollback

```bash
# View deployment history
kubectl rollout history deployment/rtx-backend -n rtx-trading

# Rollback to previous version
kubectl rollout undo deployment/rtx-backend -n rtx-trading

# Rollback to specific revision
kubectl rollout undo deployment/rtx-backend --to-revision=3 -n rtx-trading

# Check rollout status
kubectl rollout status deployment/rtx-backend -n rtx-trading
```

### Monitoring

```bash
# Port forward Prometheus
kubectl port-forward svc/prometheus 9090:9090 -n rtx-trading

# Port forward Grafana
kubectl port-forward svc/grafana 3000:3000 -n rtx-trading

# View metrics
kubectl top pods -n rtx-trading
kubectl top nodes
```

---

## Troubleshooting

### Pod Not Starting

```bash
# Describe pod
kubectl describe pod rtx-backend-xxxxx -n rtx-trading

# Check events
kubectl get events -n rtx-trading --sort-by='.lastTimestamp'

# Check logs
kubectl logs rtx-backend-xxxxx -n rtx-trading --previous
```

### Database Connection Issues

```bash
# Test connectivity
kubectl run -it --rm debug --image=busybox --restart=Never -n rtx-trading -- nc -zv postgres-service 5432

# Check database status
kubectl exec -it postgres-0 -n rtx-trading -- pg_isready -U postgres

# View database connections
kubectl exec -it postgres-0 -n rtx-trading -- psql -U postgres -c "SELECT * FROM pg_stat_activity;"
```

### High Memory Usage

```bash
# Check memory usage
kubectl top pods -n rtx-trading

# Check resource limits
kubectl describe deployment rtx-backend -n rtx-trading | grep -A 5 Limits

# Increase limits
kubectl set resources deployment rtx-backend --limits=memory=1Gi -n rtx-trading
```

### Slow API Response

```bash
# Check Prometheus metrics
curl http://localhost:9090/api/v1/query?query=histogram_quantile(0.99,rtx_order_latency_seconds)

# Check database slow queries
kubectl exec -it postgres-0 -n rtx-trading -- psql -U postgres -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Check Redis
kubectl exec -it redis-0 -n rtx-trading -- redis-cli INFO stats
```

---

## Health Checks

### Manual Health Checks

```bash
# Backend health
curl http://localhost:7999/health

# PostgreSQL health
kubectl exec postgres-0 -n rtx-trading -- pg_isready -U postgres

# Redis health
kubectl exec redis-0 -n rtx-trading -- redis-cli ping
```

### Automated Health Checks

```bash
# Check liveness probes
kubectl describe pod rtx-backend-xxxxx -n rtx-trading | grep -A 5 Liveness

# Check readiness probes
kubectl describe pod rtx-backend-xxxxx -n rtx-trading | grep -A 5 Readiness
```

---

## Security Operations

### Rotate Secrets

```bash
# Update secret
kubectl create secret generic rtx-backend-secrets \
  --from-literal=DB_PASSWORD=new_password \
  --dry-run=client -o yaml | kubectl apply -f - -n rtx-trading

# Restart pods to pick up new secret
kubectl rollout restart deployment/rtx-backend -n rtx-trading
```

### View Security Policies

```bash
# View NetworkPolicy
kubectl get networkpolicy -n rtx-trading

# View PodSecurityPolicy
kubectl get psp

# View RBAC
kubectl get rolebinding -n rtx-trading
```

---

## Cost Monitoring

### AWS Cost Breakdown

```bash
# View EKS cluster cost
aws ce get-cost-and-usage \
  --time-period Start=2026-01-01,End=2026-01-31 \
  --granularity MONTHLY \
  --metrics "UnblendedCost" \
  --group-by Type=SERVICE

# View resource usage
kubectl top nodes
kubectl top pods -n rtx-trading
```

---

## Disaster Recovery

### Full Backup

```bash
# 1. Backup database
kubectl exec postgres-0 -n rtx-trading -- pg_dump -U postgres trading_engine | gzip > db-backup-$(date +%Y%m%d).sql.gz

# 2. Upload to S3
aws s3 cp db-backup-$(date +%Y%m%d).sql.gz s3://your-backup-bucket/backups/

# 3. Backup Kubernetes manifests (already in Git)
git add k8s/
git commit -m "Update Kubernetes manifests"
git push
```

### Full Restore

```bash
# 1. Deploy infrastructure
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secrets.yaml -n rtx-trading
kubectl apply -f k8s/configmap.yaml -n rtx-trading

# 2. Deploy database
kubectl apply -f k8s/postgres-statefulset.yaml -n rtx-trading
kubectl wait --for=condition=ready pod/postgres-0 -n rtx-trading --timeout=300s

# 3. Restore database
aws s3 cp s3://your-backup-bucket/backups/db-backup-20260118.sql.gz - | \
  gunzip | \
  kubectl exec -i postgres-0 -n rtx-trading -- psql -U postgres trading_engine

# 4. Deploy application
kubectl apply -f k8s/ -n rtx-trading
```

---

## Performance Tuning

### Database Performance

```bash
# Check database performance
kubectl exec -it postgres-0 -n rtx-trading -- psql -U postgres -c "
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC
LIMIT 20;"

# Analyze query performance
kubectl exec -it postgres-0 -n rtx-trading -- psql -U postgres -c "EXPLAIN ANALYZE SELECT * FROM orders WHERE account_id = 123;"
```

### Application Performance

```bash
# Check Go runtime metrics
curl http://localhost:7999/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Check goroutines
curl http://localhost:7999/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

---

## Environment Variables Reference

### Critical Environment Variables

| Variable | Example | Description |
|----------|---------|-------------|
| `DB_HOST` | `postgres-service` | Database hostname |
| `DB_PASSWORD` | `secret123` | Database password |
| `JWT_SECRET` | `32-char-secret` | JWT signing key |
| `ENVIRONMENT` | `production` | Environment name |
| `LOG_LEVEL` | `info` | Log verbosity |

### Viewing Current Config

```bash
# View ConfigMap
kubectl get configmap rtx-backend-config -o yaml -n rtx-trading

# View Secrets (base64 encoded)
kubectl get secret rtx-backend-secrets -o yaml -n rtx-trading

# Decode secret
kubectl get secret rtx-backend-secrets -o jsonpath='{.data.DB_PASSWORD}' -n rtx-trading | base64 -d
```

---

## CI/CD GitHub Actions

### Trigger Manual Deployment

```bash
# Go to GitHub repository
# Click "Actions" tab
# Select "Deploy RTX Trading Engine" workflow
# Click "Run workflow"
# Select branch: main (production) / staging / develop
```

### View Deployment Status

```bash
# Via GitHub Actions UI
# Or via CLI
gh run list --workflow="deploy.yml"
gh run view <run-id>
```

---

## Support & Escalation

### Get Help

```bash
# Describe resources
kubectl describe pod rtx-backend-xxxxx -n rtx-trading
kubectl describe deployment rtx-backend -n rtx-trading

# Get events
kubectl get events -n rtx-trading --sort-by='.lastTimestamp' | tail -20

# Collect logs for support
kubectl logs --tail=1000 deployment/rtx-backend -n rtx-trading > backend-logs.txt
kubectl describe pod rtx-backend-xxxxx -n rtx-trading > pod-describe.txt
```

### Emergency Contacts

- **On-Call Engineer:** [PagerDuty]
- **Team Lead:** [Slack: #rtx-trading]
- **DevOps Team:** [Slack: #devops]

---

## Cheat Sheet

```bash
# Quick deploy
kubectl apply -f k8s/ -n rtx-trading && kubectl rollout status deployment/rtx-backend -n rtx-trading

# Quick rollback
kubectl rollout undo deployment/rtx-backend -n rtx-trading

# Quick scale
kubectl scale deployment rtx-backend --replicas=5 -n rtx-trading

# Quick logs
kubectl logs -f deployment/rtx-backend -n rtx-trading --tail=100

# Quick health check
curl http://localhost:7999/health && \
kubectl exec postgres-0 -n rtx-trading -- pg_isready && \
kubectl exec redis-0 -n rtx-trading -- redis-cli ping

# Quick status
kubectl get all -n rtx-trading
```

---

**Document Version:** 1.0.0
**Last Updated:** 2026-01-18
**Maintained By:** DevOps Team
