# Quick Start Guide

Get your trading engine deployed in minutes.

## 🚀 Quick Commands

### Local Development

```bash
# Start all services
docker-compose -f deployments/docker-compose.yml up -d

# View logs
docker-compose logs -f backend

# Run health check
./scripts/deploy/health-check.sh local
```

### Deploy to Staging

```bash
# Configure AWS
aws configure

# Update kubeconfig
aws eks update-kubeconfig --name trading-engine-staging --region us-east-1

# Deploy
export ECR_REGISTRY=123456789.dkr.ecr.us-east-1.amazonaws.com
export IMAGE_TAG=$(git rev-parse --short HEAD)
./scripts/deploy/deploy-staging.sh
```

### Deploy to Production

```bash
# Update kubeconfig
aws eks update-kubeconfig --name trading-engine-production --region us-east-1

# Deploy with blue-green strategy
export ECR_REGISTRY=123456789.dkr.ecr.us-east-1.amazonaws.com
export IMAGE_TAG=v1.0.0
export DEPLOYMENT_STRATEGY=blue-green
./scripts/deploy/deploy-production.sh
```

### Rollback

```bash
# Rollback staging
./scripts/deploy/rollback.sh staging

# Rollback production
./scripts/deploy/rollback.sh production
```

## 📋 Pre-Deployment Checklist

### One-Time Setup

- [ ] AWS account configured
- [ ] EKS cluster created
- [ ] ECR repository created
- [ ] DNS configured in Route53
- [ ] SSL certificate in ACM
- [ ] Secrets created in Kubernetes

### Before Each Deployment

- [ ] Code reviewed and approved
- [ ] Tests passing in CI
- [ ] Database migrations ready
- [ ] Secrets updated (if needed)
- [ ] Rollback plan prepared
- [ ] Team notified

## 🔧 Configuration

### 1. Create Kubernetes Secrets

```bash
# Production secrets
kubectl create secret generic trading-engine-secrets \
  --from-literal=database-url="postgres://user:pass@host:5432/trading?sslmode=require" \
  --from-literal=redis-url="redis://:pass@host:6379/0" \
  --from-literal=jwt-secret="change-this-in-production" \
  -n production

# Docker registry
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=${GITHUB_USER} \
  --docker-password=${GITHUB_TOKEN} \
  -n production

# TLS certificate (if using cert-manager, this is automatic)
kubectl create secret tls trading-engine-tls \
  --cert=/path/to/cert.pem \
  --key=/path/to/key.pem \
  -n production
```

### 2. Configure GitHub Secrets

Add these secrets to your GitHub repository:

**Required:**
- `DOCKER_USERNAME` - Docker Hub username
- `DOCKER_PASSWORD` - Docker Hub password/token
- `AWS_ACCESS_KEY_ID` - AWS access key
- `AWS_SECRET_ACCESS_KEY` - AWS secret key
- `STAGING_DATABASE_URL` - Staging database URL
- `PRODUCTION_DATABASE_URL` - Production database URL

**Optional:**
- `SLACK_WEBHOOK` - Slack notification webhook
- `SECURITY_SLACK_WEBHOOK` - Security alerts webhook
- `SNYK_TOKEN` - Snyk security scanning
- `FOSSA_API_KEY` - FOSSA license scanning

### 3. Update Configuration

Edit these files for your environment:

```bash
# Update ECR registry
sed -i 's/123456789.dkr.ecr.us-east-1.amazonaws.com/YOUR_ECR_REGISTRY/g' \
  deployments/kubernetes/deployment.yaml

# Update domain name
sed -i 's/api.trading-engine.example.com/YOUR_DOMAIN/g' \
  deployments/kubernetes/ingress.yaml \
  deployments/nginx/nginx.conf
```

## 🔍 Verification

### Check Deployment

```bash
# Pods running
kubectl get pods -n production

# Services available
kubectl get svc -n production

# Ingress configured
kubectl get ingress -n production

# HPA active
kubectl get hpa -n production
```

### Run Health Checks

```bash
# Automated checks
./scripts/deploy/health-check.sh production

# Manual checks
curl https://api.trading-engine.example.com/health
curl https://api.trading-engine.example.com/api/v1/info
```

## 📊 Monitoring

### Access Dashboards

```bash
# Grafana (port-forward for local access)
kubectl port-forward svc/grafana 3000:3000 -n production

# Prometheus
kubectl port-forward svc/prometheus 9090:9090 -n production

# Jaeger (distributed tracing)
kubectl port-forward svc/jaeger 16686:16686 -n production
```

Open in browser:
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090
- Jaeger: http://localhost:16686

### View Logs

```bash
# Stream logs
kubectl logs -f deployment/trading-engine -n production

# Last 100 lines
kubectl logs --tail=100 deployment/trading-engine -n production

# Specific pod
kubectl logs <pod-name> -n production
```

## 🐛 Troubleshooting

### Pod Not Starting

```bash
# Describe pod
kubectl describe pod <pod-name> -n production

# Check events
kubectl get events -n production --sort-by='.lastTimestamp' | tail -20

# Shell into pod
kubectl exec -it <pod-name> -n production -- /bin/sh
```

### Deployment Failed

```bash
# Check rollout status
kubectl rollout status deployment/trading-engine -n production

# View rollout history
kubectl rollout history deployment/trading-engine -n production

# Rollback immediately
./scripts/deploy/rollback.sh production
```

### Database Connection Issues

```bash
# Test from pod
kubectl exec -it <pod-name> -n production -- sh
wget -qO- http://localhost:8080/health/db

# Check secrets
kubectl get secret trading-engine-secrets -n production -o jsonpath='{.data.database-url}' | base64 -d
```

## 📈 Performance Testing

### Load Test

```bash
# Install k6
brew install k6  # macOS
# or
sudo apt install k6  # Ubuntu

# Run load test
k6 run --vus 100 --duration 5m tests/performance/load-test.js
```

### Stress Test

```bash
# Gradual ramp-up
k6 run --stage 2m:100 \
       --stage 5m:200 \
       --stage 2m:300 \
       --stage 5m:300 \
       --stage 10m:0 \
       tests/performance/stress-test.js
```

## 🔐 Security

### Scan for Vulnerabilities

```bash
# Scan dependencies
go list -json -m all | docker run --rm -i sonatypecommunity/nancy:latest sleuth

# Scan Docker image
trivy image trading-engine:latest

# Scan Kubernetes manifests
kubectl apply -f deployments/kubernetes/ --dry-run=server
```

### Update Secrets

```bash
# Update database password
kubectl create secret generic trading-engine-secrets \
  --from-literal=database-url="new-url" \
  --dry-run=client -o yaml | kubectl apply -f - -n production

# Restart pods to pick up new secrets
kubectl rollout restart deployment/trading-engine -n production
```

## 🔄 Common Operations

### Scale Deployment

```bash
# Manual scaling
kubectl scale deployment trading-engine --replicas=5 -n production

# Check HPA status
kubectl get hpa -n production
```

### Update Image

```bash
# Update to new version
kubectl set image deployment/trading-engine \
  app=ghcr.io/your-org/trading-engine:v1.1.0 \
  -n production

# Watch rollout
kubectl rollout status deployment/trading-engine -n production
```

### Database Maintenance

```bash
# Run migrations
export DATABASE_URL="postgres://user:pass@host:5432/trading"
./scripts/deploy/migrate.sh

# Backup database
kubectl exec -it postgres-0 -n production -- \
  pg_dump -U postgres trading > backup-$(date +%Y%m%d).sql
```

## 📞 Support

- **Documentation**: See [DEPLOYMENT.md](DEPLOYMENT.md)
- **Logs**: Check Grafana or `kubectl logs`
- **Alerts**: Configure Slack notifications
- **Issues**: Create GitHub issue

## 🎯 Next Steps

1. **Set up monitoring alerts** in Prometheus
2. **Configure Slack notifications** for deployments
3. **Set up automated backups** for database
4. **Configure auto-scaling** based on metrics
5. **Implement disaster recovery** plan
6. **Document runbooks** for common incidents

---

**Remember**: Always test in staging before production!
