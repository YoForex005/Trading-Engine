# Quick Start Guide - Trading Engine Deployment

## 🚀 Get Started in 5 Minutes

### Local Development (Docker Compose)

```bash
# 1. Clone and setup
git clone https://github.com/epic1st/trading-engine.git
cd trading-engine

# 2. Configure environment
cp .env.example .env
# Edit .env with your API keys

# 3. Start all services
docker-compose -f deployments/docker/docker-compose.yml up -d

# 4. Verify
curl http://localhost:7999/health
# Expected: "OK"

# 5. Access services
# API: http://localhost:7999
# WebSocket: ws://localhost:7999/ws
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9091
```

### Production Deployment (Kubernetes)

```bash
# 1. Setup infrastructure (Terraform)
cd deployments/terraform
terraform init
terraform apply -var="environment=production"

# 2. Configure kubectl
aws eks update-kubeconfig --region us-east-1 --name production-trading-engine

# 3. Create secrets
kubectl create namespace production
kubectl create secret generic trading-engine-secrets \
  --from-literal=database-password='YOUR_PASSWORD' \
  --from-literal=redis-password='YOUR_PASSWORD' \
  --from-literal=jwt-secret='YOUR_SECRET' \
  --from-literal=oanda-api-key='YOUR_KEY' \
  -n production

# 4. Deploy with Helm
helm install trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set image.tag=v1.0.0

# 5. Verify deployment
kubectl get pods -n production
kubectl get svc -n production
kubectl get ingress -n production

# 6. Access
# https://api.trading-engine.com
```

## 📁 Project Structure

```
trading-engine/
├── .github/
│   └── workflows/
│       ├── ci-cd.yml              # Main CI/CD pipeline
│       └── security-scan.yml      # Security scanning
├── backend/
│   ├── cmd/server/main.go         # Main application
│   └── ...
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile.api         # API server image
│   │   ├── Dockerfile.websocket   # WebSocket server
│   │   ├── Dockerfile.fix-gateway # FIX gateway
│   │   ├── Dockerfile.worker      # Background workers
│   │   ├── docker-compose.yml     # Local development
│   │   └── prometheus.yml         # Metrics config
│   ├── helm/
│   │   └── trading-engine/
│   │       ├── Chart.yaml         # Helm chart
│   │       ├── values.yaml        # Default values
│   │       └── templates/         # K8s manifests
│   ├── terraform/
│   │   ├── main.tf                # Infrastructure
│   │   ├── variables.tf           # Configuration
│   │   └── outputs.tf             # Outputs
│   └── monitoring/
│       └── alert-rules.yml        # Prometheus alerts
├── tests/
│   └── load/
│       └── load-test.js           # k6 load tests
└── docs/
    └── deployment/
        ├── README.md              # Full deployment guide
        ├── RUNBOOK.md             # Operations manual
        ├── CICD-OVERVIEW.md       # Pipeline details
        ├── INFRASTRUCTURE.md      # Architecture
        └── QUICK-START.md         # This file
```

## 🔧 Common Tasks

### Build Docker Images

```bash
cd backend

# API Server
docker build -f ../deployments/docker/Dockerfile.api -t trading-engine-api .

# WebSocket
docker build -f ../deployments/docker/Dockerfile.websocket -t trading-engine-ws .

# FIX Gateway
docker build -f ../deployments/docker/Dockerfile.fix-gateway -t trading-engine-fix .

# Workers
docker build -f ../deployments/docker/Dockerfile.worker -t trading-engine-worker .
```

### Deploy to Staging

```bash
# Commit to develop branch
git checkout develop
git add .
git commit -m "feat: new feature"
git push origin develop

# GitHub Actions automatically:
# 1. Runs tests
# 2. Builds images
# 3. Deploys to staging
# 4. Runs smoke tests
# 5. Runs load tests
```

### Deploy to Production

```bash
# Create release tag
git checkout main
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0

# GitHub Actions:
# 1. Runs full test suite
# 2. Builds images
# 3. Deploys to blue environment
# 4. Runs smoke tests
# 5. Switches traffic (blue-green)
# 6. Monitors for issues
```

### Scale Services

```bash
# Manual scaling
kubectl scale deployment trading-engine-api -n production --replicas=10

# Update HPA limits
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set api.autoscaling.maxReplicas=20
```

### View Logs

```bash
# Recent logs
kubectl logs -n production deployment/trading-engine-api --tail=100

# Follow logs
kubectl logs -f -n production deployment/trading-engine-api

# Logs from specific pod
kubectl logs -n production <pod-name>

# All containers in pod
kubectl logs -n production <pod-name> --all-containers=true
```

### Rollback Deployment

```bash
# List releases
helm history trading-engine -n production

# Rollback to previous
helm rollback trading-engine -n production

# Rollback to specific version
helm rollback trading-engine 5 -n production
```

## 🔍 Monitoring

### Access Dashboards

```bash
# Grafana (port-forward)
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# http://localhost:3000 (admin/prom-operator)

# Prometheus
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090
# http://localhost:9090

# Jaeger
kubectl port-forward -n monitoring svc/jaeger-query 16686:16686
# http://localhost:16686
```

### Key Metrics

```bash
# Pod status
kubectl get pods -n production

# Resource usage
kubectl top pods -n production
kubectl top nodes

# HPA status
kubectl get hpa -n production

# Service endpoints
kubectl get svc -n production

# Ingress status
kubectl get ingress -n production
```

## 🧪 Testing

### Run Unit Tests

```bash
cd backend
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Integration Tests

```bash
# Start dependencies
docker-compose -f deployments/docker/docker-compose.yml up -d postgres redis

# Run tests
cd backend
DATABASE_URL=postgresql://trading:trading_pass@localhost:5432/trading_engine \
REDIS_URL=redis://localhost:6379 \
go test -v -tags=integration ./...
```

### Run Load Tests

```bash
# Install k6
brew install k6  # macOS
# or sudo apt install k6  # Linux

# Run test
k6 run tests/load/load-test.js \
  --env BASE_URL=https://staging.trading-engine.com \
  --env WS_URL=wss://staging.trading-engine.com/ws

# Stress test
k6 run tests/load/load-test.js --vus 5000 --duration 10m
```

## 🔐 Security

### Scan for Vulnerabilities

```bash
# Go dependencies
cd backend
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Container images
docker build -t trading-engine-api:test -f deployments/docker/Dockerfile.api .
trivy image trading-engine-api:test

# Infrastructure
cd deployments/terraform
tfsec .
```

### Rotate Secrets

```bash
# Generate new secrets
NEW_JWT_SECRET=$(openssl rand -base64 32)
NEW_DB_PASSWORD=$(openssl rand -base64 24)

# Update in Kubernetes
kubectl create secret generic trading-engine-secrets \
  --from-literal=jwt-secret="$NEW_JWT_SECRET" \
  --from-literal=database-password="$NEW_DB_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods to pick up new secrets
kubectl rollout restart deployment/trading-engine-api -n production
```

## 🆘 Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n production

# Describe pod for events
kubectl describe pod <pod-name> -n production

# Check logs
kubectl logs <pod-name> -n production
```

### Database Connection Issues

```bash
# Test connectivity
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -- \
  psql -h trading-engine-postgresql -U trading -d trading_engine

# Check database logs
kubectl logs -n production statefulset/trading-engine-postgresql
```

### High Latency

```bash
# Check pod resources
kubectl top pods -n production

# Check node resources
kubectl top nodes

# Check database performance
kubectl exec -it -n production statefulset/trading-engine-postgresql-0 -- \
  psql -U trading -d trading_engine -c "SELECT * FROM pg_stat_activity;"
```

### Service Unreachable

```bash
# Check service
kubectl get svc -n production

# Check endpoints
kubectl get endpoints -n production

# Check ingress
kubectl describe ingress trading-engine-ingress -n production

# Test from within cluster
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://trading-engine-api:7999/health
```

## 📚 Documentation

- **Full Deployment Guide**: [README.md](./README.md)
- **Operations Runbook**: [RUNBOOK.md](./RUNBOOK.md)
- **CI/CD Overview**: [CICD-OVERVIEW.md](./CICD-OVERVIEW.md)
- **Infrastructure**: [INFRASTRUCTURE.md](./INFRASTRUCTURE.md)

## 🔗 Useful Links

### Production URLs
- API: https://api.trading-engine.com
- WebSocket: wss://ws.trading-engine.com
- Admin: https://admin.trading-engine.com

### Monitoring
- Grafana: https://grafana.trading-engine.com
- Prometheus: https://prometheus.trading-engine.com
- Jaeger: https://jaeger.trading-engine.com

### Source Code
- GitHub: https://github.com/epic1st/trading-engine
- Container Registry: https://ghcr.io/epic1st/trading-engine

## 🤝 Support

- **Slack**: #trading-engine
- **Email**: devops@trading-engine.com
- **On-Call**: PagerDuty
- **Issues**: GitHub Issues

---

**Last Updated**: 2026-01-18
**Version**: 1.0
