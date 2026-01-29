# Trading Engine - Deployment System

Enterprise-grade deployment automation with CI/CD pipelines, containerization, and Kubernetes orchestration.

## 📁 Structure

```
deployments/
├── .github/workflows/          # GitHub Actions CI/CD
│   ├── ci.yml                  # Continuous Integration
│   ├── cd.yml                  # Continuous Deployment
│   ├── security-scan.yml       # Daily security scanning
│   └── performance-test.yml    # Weekly performance testing
│
├── kubernetes/                 # Kubernetes manifests
│   ├── deployment.yaml         # Main + canary deployments
│   ├── service.yaml            # Services (LB, headless, canary)
│   ├── ingress.yaml            # NGINX ingress with TLS
│   ├── configmap.yaml          # Configuration
│   ├── secrets.yaml            # Secrets (template)
│   ├── hpa.yaml                # Horizontal Pod Autoscaler
│   ├── rbac.yaml               # RBAC configuration
│   └── pvc.yaml                # Persistent volume claims
│
├── scripts/deploy/             # Deployment scripts
│   ├── deploy-staging.sh       # Deploy to staging
│   ├── deploy-production.sh    # Deploy to production
│   ├── rollback.sh             # Automated rollback
│   ├── health-check.sh         # Post-deployment verification
│   ├── canary-metrics.sh       # Canary metrics validation
│   └── migrate.sh              # Database migrations
│
├── docker-compose.yml          # Local development
├── docker-compose.production.yml # Production-like setup
├── Dockerfile.backend          # Backend multi-stage build
├── Dockerfile.frontend         # Frontend with NGINX
├── nginx/                      # NGINX configuration
└── monitoring/                 # Prometheus/Grafana config
```

## 🚀 Features

### CI/CD Pipelines

- **Continuous Integration** - Automated testing, linting, security scanning
- **Continuous Deployment** - Automated staging/production deployment
- **Security Scanning** - Daily vulnerability and compliance checks
- **Performance Testing** - Weekly load and stress testing

### Deployment Strategies

- **Rolling Update** - Zero-downtime incremental updates
- **Blue-Green** - Complete environment switching
- **Canary** - Progressive traffic shifting (10% → 25% → 50% → 100%)

### Infrastructure

- **Kubernetes** - Container orchestration with auto-scaling
- **Docker** - Multi-stage optimized builds
- **NGINX** - Reverse proxy with SSL/TLS
- **Prometheus** - Metrics collection
- **Grafana** - Visualization and dashboards

### Automation

- **Automated rollback** on health check failure
- **Database migrations** as part of deployment
- **Health checks** and smoke tests
- **Secret management** with Kubernetes secrets
- **Horizontal Pod Autoscaling** based on CPU/memory/custom metrics

## 📋 Quick Start

See [QUICKSTART.md](QUICKSTART.md) for rapid deployment guide.

See [DEPLOYMENT.md](DEPLOYMENT.md) for comprehensive documentation.

### Local Development

```bash
docker-compose up -d
```

### Deploy to Staging

```bash
export ECR_REGISTRY=your-registry
export IMAGE_TAG=$(git rev-parse --short HEAD)
./scripts/deploy/deploy-staging.sh
```

### Deploy to Production

```bash
export ECR_REGISTRY=your-registry
export IMAGE_TAG=v1.0.0
export DEPLOYMENT_STRATEGY=blue-green
./scripts/deploy/deploy-production.sh
```

## 🔧 Configuration

### Required Secrets

```bash
kubectl create secret generic trading-engine-secrets \
  --from-literal=database-url="..." \
  --from-literal=redis-url="..." \
  --from-literal=jwt-secret="..." \
  -n production
```

### GitHub Secrets

- `DOCKER_USERNAME`, `DOCKER_PASSWORD`
- `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`
- `STAGING_DATABASE_URL`, `PRODUCTION_DATABASE_URL`
- `SLACK_WEBHOOK` (optional)

## 📊 Monitoring

- **Prometheus** - http://localhost:9091 (via port-forward)
- **Grafana** - http://localhost:3000 (admin/admin)
- **Jaeger** - http://localhost:16686 (tracing)

## 🐛 Troubleshooting

```bash
# Check pods
kubectl get pods -n production

# View logs
kubectl logs -f deployment/trading-engine -n production

# Rollback
./scripts/deploy/rollback.sh production

# Run health checks
./scripts/deploy/health-check.sh production
```

## 🔐 Security

- Daily vulnerability scanning (Trivy, Grype, Snyk)
- Secret detection (Gitleaks, TruffleHog)
- Code analysis (CodeQL, Semgrep)
- License compliance (FOSSA)
- Container security (non-root user, minimal image)

## 📈 Performance

- Horizontal Pod Autoscaling (3-10 replicas)
- Resource limits and requests
- Connection pooling and caching
- Load balancing with NGINX
- Weekly performance testing

## 📚 Documentation

- [DEPLOYMENT.md](DEPLOYMENT.md) - Complete deployment guide
- [QUICKSTART.md](QUICKSTART.md) - Quick start guide
- [kubernetes/](kubernetes/) - Kubernetes manifest documentation

## 🎯 Deployment Checklist

- [ ] Code reviewed and merged
- [ ] Tests passing in CI
- [ ] Secrets configured
- [ ] Database migrations ready
- [ ] Team notified
- [ ] Rollback plan prepared
- [ ] Monitoring configured
- [ ] Health checks verified

## 📞 Support

For issues or questions, see the troubleshooting section in [DEPLOYMENT.md](DEPLOYMENT.md).
