# Deployment Guide - Trading Engine

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Local Development](#local-development)
4. [Docker Deployment](#docker-deployment)
5. [Kubernetes Deployment](#kubernetes-deployment)
6. [Cloud Deployment (AWS)](#cloud-deployment-aws)
7. [Monitoring Setup](#monitoring-setup)
8. [Backup & Recovery](#backup--recovery)
9. [Security Hardening](#security-hardening)
10. [Troubleshooting](#troubleshooting)

## Overview

The Trading Engine platform consists of:

- **API Server** (Port 7999): REST API for trading operations
- **WebSocket Server** (Port 8080): Real-time market data streaming
- **FIX Gateway** (Port 9090): FIX 4.4 protocol connectivity
- **Background Workers**: Order processing, price aggregation
- **PostgreSQL**: Primary database
- **Redis**: Caching and pub/sub
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards

## Prerequisites

### Required Software

- Docker 24.0+
- Docker Compose 2.20+
- Kubernetes 1.28+
- Helm 3.12+
- Terraform 1.6+
- kubectl 1.28+
- AWS CLI 2.0+ (for AWS deployment)

### Required Access

- GitHub Container Registry credentials
- AWS account (for cloud deployment)
- Domain names and SSL certificates
- API keys for liquidity providers (OANDA, Binance)

## Local Development

### Using Docker Compose

```bash
# Clone repository
git clone https://github.com/epic1st/trading-engine.git
cd trading-engine

# Set environment variables
cp .env.example .env
# Edit .env with your configuration

# Start all services
docker-compose -f deployments/docker/docker-compose.yml up -d

# View logs
docker-compose -f deployments/docker/docker-compose.yml logs -f

# Stop services
docker-compose -f deployments/docker/docker-compose.yml down
```

### Service URLs (Local)

- API: http://localhost:7999
- WebSocket: ws://localhost:7999/ws
- Prometheus: http://localhost:9091
- Grafana: http://localhost:3000 (admin/admin)
- Jaeger: http://localhost:16686

## Docker Deployment

### Build Images

```bash
# Build all service images
cd backend

# API Server
docker build -f ../deployments/docker/Dockerfile.api -t trading-engine-api:latest .

# WebSocket Server
docker build -f ../deployments/docker/Dockerfile.websocket -t trading-engine-websocket:latest .

# FIX Gateway
docker build -f ../deployments/docker/Dockerfile.fix-gateway -t trading-engine-fix:latest .

# Workers
docker build -f ../deployments/docker/Dockerfile.worker -t trading-engine-worker:latest .
```

### Multi-Architecture Build

```bash
# Setup buildx
docker buildx create --name multiarch --driver docker-container --use
docker buildx inspect --bootstrap

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f deployments/docker/Dockerfile.api \
  -t ghcr.io/epic1st/trading-engine-api:latest \
  --push \
  ./backend
```

## Kubernetes Deployment

### Prerequisites

```bash
# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Add Helm repositories
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### Deploy to Kubernetes

```bash
# Create namespace
kubectl create namespace trading-engine

# Create secrets
kubectl create secret generic trading-engine-secrets \
  --from-literal=database-password='YOUR_DB_PASSWORD' \
  --from-literal=redis-password='YOUR_REDIS_PASSWORD' \
  --from-literal=jwt-secret='YOUR_JWT_SECRET' \
  --from-literal=oanda-api-key='YOUR_OANDA_KEY' \
  --from-literal=oanda-account-id='YOUR_OANDA_ACCOUNT' \
  -n trading-engine

# Install with Helm
helm install trading-engine ./deployments/helm/trading-engine \
  --namespace trading-engine \
  --values deployments/helm/trading-engine/values-production.yaml \
  --set image.tag=v1.0.0

# Verify deployment
kubectl get pods -n trading-engine
kubectl get services -n trading-engine
```

### Upgrade Deployment

```bash
# Upgrade to new version
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace trading-engine \
  --set image.tag=v1.1.0 \
  --wait

# Rollback if needed
helm rollback trading-engine -n trading-engine
```

### Blue-Green Deployment

```bash
# Deploy to blue environment
helm install trading-engine-blue ./deployments/helm/trading-engine \
  --namespace production \
  --set environment=blue \
  --set image.tag=v1.1.0

# Test blue environment
kubectl port-forward svc/trading-engine-blue 8000:7999 -n production

# Switch traffic
kubectl patch service trading-engine-prod -n production \
  -p '{"spec":{"selector":{"environment":"blue"}}}'

# Scale down green environment
kubectl scale deployment trading-engine-green -n production --replicas=1
```

## Cloud Deployment (AWS)

### Infrastructure Setup

```bash
cd deployments/terraform

# Initialize Terraform
terraform init

# Plan infrastructure
terraform plan \
  -var="environment=production" \
  -var="aws_region=us-east-1" \
  -out=tfplan

# Apply infrastructure
terraform apply tfplan

# Get EKS credentials
aws eks update-kubeconfig \
  --region us-east-1 \
  --name production-trading-engine
```

### Deploy to EKS

```bash
# Install metrics server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Install ingress controller
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace

# Install cert-manager for SSL
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Deploy application
helm install trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --create-namespace \
  --values deployments/helm/trading-engine/values-aws-production.yaml
```

### Auto-Scaling Configuration

```bash
# Install Cluster Autoscaler
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml

# Install Metrics Server (if not already installed)
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Verify HPA
kubectl get hpa -n production
```

## Monitoring Setup

### Prometheus

```bash
# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.retention=30d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=50Gi
```

### Grafana Dashboards

```bash
# Access Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80

# Default credentials: admin / prom-operator

# Import dashboards from deployments/monitoring/dashboards/
```

### Alerting

```bash
# Configure Alertmanager
kubectl create secret generic alertmanager-config \
  --from-file=alertmanager.yaml=deployments/monitoring/alertmanager-config.yaml \
  -n monitoring

# Apply alert rules
kubectl apply -f deployments/monitoring/alert-rules.yaml -n monitoring
```

## Backup & Recovery

### Database Backup

```bash
# Automated daily backups (configured in Terraform)
# Manual backup
kubectl exec -n production deployment/postgres -- \
  pg_dump -U trading trading_engine > backup-$(date +%Y%m%d).sql

# Restore from backup
kubectl exec -i -n production deployment/postgres -- \
  psql -U trading trading_engine < backup-20260118.sql
```

### Velero for Kubernetes Backups

```bash
# Install Velero
velero install \
  --provider aws \
  --plugins velero/velero-plugin-for-aws:v1.8.0 \
  --bucket trading-engine-backups \
  --backup-location-config region=us-east-1 \
  --snapshot-location-config region=us-east-1

# Create backup
velero backup create trading-engine-backup --include-namespaces production

# Restore
velero restore create --from-backup trading-engine-backup
```

## Security Hardening

### Network Policies

```bash
# Network policies are deployed automatically via Helm
# Verify
kubectl get networkpolicies -n production
```

### Pod Security Standards

```bash
# Enforce pod security
kubectl label namespace production \
  pod-security.kubernetes.io/enforce=restricted \
  pod-security.kubernetes.io/audit=restricted \
  pod-security.kubernetes.io/warn=restricted
```

### Secrets Management with Vault

```bash
# Install Vault
helm install vault hashicorp/vault \
  --namespace vault \
  --create-namespace

# Configure Vault CSI provider
# See: docs/security/vault-setup.md
```

### SSL/TLS Configuration

```bash
# Create ClusterIssuer for Let's Encrypt
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@trading-engine.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

## Troubleshooting

### Common Issues

#### Pods not starting

```bash
# Check pod status
kubectl get pods -n production

# View pod logs
kubectl logs -n production deployment/trading-engine-api --tail=100

# Describe pod for events
kubectl describe pod -n production <pod-name>
```

#### Database connection issues

```bash
# Test database connectivity
kubectl run -it --rm debug --image=postgres:16-alpine --restart=Never -- \
  psql -h trading-engine-postgresql -U trading -d trading_engine

# Check database logs
kubectl logs -n production statefulset/trading-engine-postgresql
```

#### High latency

```bash
# Check metrics
kubectl top pods -n production
kubectl top nodes

# View Prometheus metrics
kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090

# Check HPA status
kubectl get hpa -n production
```

### Performance Tuning

```bash
# Increase pod resources
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set api.resources.limits.cpu=4000m \
  --set api.resources.limits.memory=4Gi

# Adjust HPA thresholds
helm upgrade trading-engine ./deployments/helm/trading-engine \
  --namespace production \
  --set api.autoscaling.targetCPUUtilizationPercentage=60
```

### Disaster Recovery

```bash
# Restore from last backup
velero restore create --from-backup trading-engine-backup-latest

# Database point-in-time recovery
# See AWS RDS documentation for PITR

# Verify restoration
kubectl get all -n production
```

## CI/CD Pipeline

### GitHub Actions Workflows

All deployments are automated via GitHub Actions:

- `.github/workflows/ci-cd.yml` - Main CI/CD pipeline
- `.github/workflows/security-scan.yml` - Security scanning

### Manual Deployment Trigger

```bash
# Trigger deployment via GitHub CLI
gh workflow run ci-cd.yml \
  -f environment=production \
  -f version=v1.0.0
```

## Load Testing

```bash
# Install k6
brew install k6  # macOS
# or
sudo apt install k6  # Linux

# Run load test
k6 run tests/load/load-test.js \
  --env BASE_URL=https://api.trading-engine.com \
  --env WS_URL=wss://ws.trading-engine.com/ws

# Run stress test
k6 run tests/load/stress-test.js --vus 5000 --duration 10m
```

## Support

- Documentation: `/docs`
- Issues: GitHub Issues
- Monitoring: Grafana dashboards
- Logs: Kibana / CloudWatch
