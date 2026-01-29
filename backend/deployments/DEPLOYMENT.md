# Trading Engine - Deployment Guide

Complete enterprise-grade deployment automation with CI/CD pipelines, containerization, and Kubernetes orchestration.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Local Development](#local-development)
- [CI/CD Pipelines](#cicd-pipelines)
- [Deployment Strategies](#deployment-strategies)
- [Kubernetes Setup](#kubernetes-setup)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Overview

This deployment system provides:

- **Zero-downtime deployments** with rolling updates
- **Automated rollback** on health check failure
- **Blue-green deployment** support
- **Canary deployment** with progressive traffic shifting
- **Comprehensive health checks** and smoke tests
- **Enterprise security** scanning and compliance
- **Performance testing** automation
- **Multi-environment** support (staging/production)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      GitHub Actions                          │
│  ┌──────────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐ │
│  │    CI    │  │    CD    │  │  Security  │  │   Perf   │ │
│  └──────────┘  └──────────┘  └────────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │    Ingress   │→ │    Service   │→ │  Deployment  │      │
│  │   (NGINX)    │  │ (LoadBalancer)│  │  (3 Pods)    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Postgres   │  │    Redis     │  │  Prometheus  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

### Required Tools

- **Docker** (20.10+)
- **kubectl** (1.28+)
- **AWS CLI** (2.0+)
- **Helm** (3.0+) (optional)
- **k6** (for load testing)

### AWS Resources

- **EKS Cluster** for Kubernetes
- **ECR** for Docker images
- **RDS** for PostgreSQL (optional)
- **ElastiCache** for Redis (optional)
- **Route53** for DNS
- **ACM** for SSL certificates

## Local Development

### Using Docker Compose

```bash
# Start all services
docker-compose -f deployments/docker-compose.yml up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

### Build Docker Image

```bash
# Build backend
docker build -f deployments/Dockerfile.backend -t trading-engine:latest .

# Build with specific version
docker build -f deployments/Dockerfile.backend \
  --build-arg VERSION=v1.0.0 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t trading-engine:v1.0.0 .
```

### Local Testing

```bash
# Run health checks
./scripts/deploy/health-check.sh local

# Test locally
curl http://localhost:8080/health
curl http://localhost:8080/metrics
```

## CI/CD Pipelines

### 1. Continuous Integration (ci.yml)

Runs on every push and PR:

- **Unit tests** with race detection
- **Integration tests** with real database
- **Linting** (golangci-lint, staticcheck)
- **Code coverage** with Codecov
- **Docker build** and push to registry
- **Security scanning** (Trivy, Gosec)

### 2. Continuous Deployment (cd.yml)

Deploys to staging/production:

- **Staging**: Automatic on main branch
- **Production**: Manual or on version tags
- **Blue-green** or **canary** deployment
- **Database migrations**
- **Health checks** and smoke tests
- **Automated rollback** on failure

### 3. Security Scanning (security-scan.yml)

Daily security scans:

- **Dependency vulnerabilities** (Snyk, Nancy)
- **Code analysis** (CodeQL, Semgrep)
- **Container scanning** (Trivy, Grype)
- **Secret detection** (Gitleaks, TruffleHog)
- **License compliance** (FOSSA)

### 4. Performance Testing (performance-test.yml)

Weekly performance tests:

- **Load testing** with k6
- **Stress testing** with gradual ramp-up
- **Go benchmarks** with historical tracking
- **Database performance** profiling
- **Memory profiling**

## Deployment Strategies

### Rolling Update (Default)

Zero-downtime rolling update:

```bash
./scripts/deploy/deploy-staging.sh
```

### Blue-Green Deployment

Complete environment switch:

```bash
export DEPLOYMENT_STRATEGY=blue-green
./scripts/deploy/deploy-production.sh
```

Process:
1. Deploy "green" environment
2. Run smoke tests
3. Switch traffic to green
4. Delete old "blue" environment

### Canary Deployment

Progressive traffic shifting:

```bash
export DEPLOYMENT_STRATEGY=canary
./scripts/deploy/deploy-production.sh
```

Process:
1. Deploy canary with 10% traffic
2. Monitor metrics for 5 minutes
3. Gradually increase: 10% → 25% → 50% → 100%
4. Automatic rollback if metrics fail

## Kubernetes Setup

### 1. Create Namespace

```bash
kubectl create namespace production
kubectl create namespace staging
```

### 2. Configure Secrets

```bash
# Database secret
kubectl create secret generic trading-engine-secrets \
  --from-literal=database-url="postgres://user:pass@host:5432/db?sslmode=require" \
  --from-literal=redis-url="redis://:pass@host:6379/0" \
  --from-literal=jwt-secret="your-super-secret-key" \
  -n production

# Docker registry secret
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=your-username \
  --docker-password=your-token \
  -n production

# TLS certificate
kubectl create secret tls trading-engine-tls \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  -n production
```

### 3. Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f deployments/kubernetes/ -n production

# Or deploy with script
export ECR_REGISTRY=your-registry
export IMAGE_TAG=v1.0.0
./scripts/deploy/deploy-production.sh
```

### 4. Verify Deployment

```bash
# Check pods
kubectl get pods -n production

# Check deployment
kubectl get deployment trading-engine -n production

# Check service
kubectl get service trading-engine -n production

# Check ingress
kubectl get ingress -n production

# View logs
kubectl logs -f deployment/trading-engine -n production
```

## Database Migrations

### Run Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://user:pass@host:5432/db?sslmode=require"

# Run migrations
./scripts/deploy/migrate.sh

# Or manually with golang-migrate
migrate -path ./migrations \
  -database "$DATABASE_URL" \
  up
```

### Rollback Migrations

```bash
migrate -path ./migrations \
  -database "$DATABASE_URL" \
  down 1
```

## Health Checks

### Automated Health Checks

```bash
# Staging
./scripts/deploy/health-check.sh staging

# Production
./scripts/deploy/health-check.sh production
```

Checks performed:
- Basic health endpoint (/)
- Database connectivity
- Redis connectivity
- API smoke tests
- Response time validation
- Kubernetes pod status

### Manual Health Checks

```bash
# Health endpoint
curl https://api.trading-engine.example.com/health

# Database health
curl https://api.trading-engine.example.com/health/db

# Redis health
curl https://api.trading-engine.example.com/health/redis

# Metrics
curl https://api.trading-engine.example.com/metrics
```

## Rollback

### Automatic Rollback

Triggered automatically if:
- Health checks fail after deployment
- Canary metrics exceed thresholds
- Deployment timeout

### Manual Rollback

```bash
# Rollback staging
./scripts/deploy/rollback.sh staging

# Rollback production
./scripts/deploy/rollback.sh production
```

## Monitoring

### Prometheus Metrics

Available at `/metrics`:

- HTTP request rate, latency, errors
- Database connection pool stats
- Redis operations
- Go runtime metrics (goroutines, memory, GC)
- Custom business metrics

### Grafana Dashboards

Access at `http://localhost:3000`:

- **Application Dashboard**: Request rates, errors, latency
- **Database Dashboard**: Connections, query performance
- **Infrastructure Dashboard**: CPU, memory, network
- **Business Metrics**: Trading-specific KPIs

### Alerts

Configure in Prometheus:

```yaml
# Example alert
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: High error rate detected
```

## Environment Variables

### Required

- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - JWT signing secret

### Optional

- `ENV` - Environment name (development/staging/production)
- `PORT` - HTTP server port (default: 8080)
- `METRICS_PORT` - Metrics server port (default: 9090)
- `LOG_LEVEL` - Logging level (debug/info/warn/error)
- `SENTRY_DSN` - Sentry error tracking DSN
- `DATADOG_API_KEY` - DataDog monitoring API key

## Scaling

### Horizontal Pod Autoscaling

```bash
# Already configured in hpa.yaml
kubectl get hpa -n production

# Manual scaling
kubectl scale deployment trading-engine --replicas=5 -n production
```

### Vertical Scaling

Edit resource limits in `deployment.yaml`:

```yaml
resources:
  limits:
    memory: "4Gi"
    cpu: "2000m"
  requests:
    memory: "2Gi"
    cpu: "1000m"
```

## Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl describe pod <pod-name> -n production

# Check logs
kubectl logs <pod-name> -n production

# Check events
kubectl get events -n production --sort-by='.lastTimestamp'
```

### Deployment Stuck

```bash
# Check rollout status
kubectl rollout status deployment/trading-engine -n production

# View rollout history
kubectl rollout history deployment/trading-engine -n production

# Rollback
./scripts/deploy/rollback.sh production
```

### Database Connection Issues

```bash
# Test from pod
kubectl exec -it <pod-name> -n production -- sh
wget -qO- http://localhost:8080/health/db

# Check secrets
kubectl get secret trading-engine-secrets -n production -o yaml
```

### High Memory Usage

```bash
# Check pod metrics
kubectl top pod -n production

# Generate memory profile
go tool pprof http://localhost:9090/debug/pprof/heap
```

## Security Best Practices

1. **Never commit secrets** to git
2. **Use Kubernetes secrets** or external secret managers (AWS Secrets Manager, Vault)
3. **Enable RBAC** for service accounts
4. **Use network policies** to restrict pod communication
5. **Scan images** regularly for vulnerabilities
6. **Enable pod security policies**
7. **Use TLS** for all external communication
8. **Rotate credentials** regularly

## Cost Optimization

1. **Right-size resources** based on actual usage
2. **Use spot instances** for non-critical workloads
3. **Enable cluster autoscaling**
4. **Set pod disruption budgets** appropriately
5. **Clean up unused resources** (old images, PVs)
6. **Use resource quotas** per namespace

## Support

For issues or questions:

- Check the [Troubleshooting](#troubleshooting) section
- Review pod logs and events
- Contact DevOps team
- Create an issue in the repository

## References

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Prometheus](https://prometheus.io/docs/)
- [Grafana](https://grafana.com/docs/)
