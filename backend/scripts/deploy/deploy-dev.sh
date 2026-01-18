#!/bin/bash

set -euo pipefail

# RTX Trading Engine - Development Deployment Script
# This script deploys the application to the development environment

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
NAMESPACE="rtx-dev"
IMAGE_TAG="${IMAGE_TAG:-dev-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl not found. Please install kubectl."
        exit 1
    fi

    if ! command -v docker &> /dev/null; then
        log_error "docker not found. Please install Docker."
        exit 1
    fi

    log_info "Prerequisites check passed"
}

create_namespace() {
    log_info "Creating namespace $NAMESPACE if it doesn't exist..."
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    kubectl label namespace "$NAMESPACE" environment=development --overwrite
}

apply_configmaps() {
    log_info "Applying ConfigMaps..."
    kubectl apply -f "$PROJECT_ROOT/k8s/configmap.yaml" -n "$NAMESPACE"
}

apply_secrets() {
    log_info "Applying Secrets..."

    # Check if secrets exist, create if not
    if ! kubectl get secret rtx-secrets -n "$NAMESPACE" &> /dev/null; then
        log_warn "Secrets not found. Creating from template..."
        kubectl apply -f "$PROJECT_ROOT/k8s/secret.yaml" -n "$NAMESPACE"
    else
        log_info "Secrets already exist, skipping creation"
    fi
}

deploy_database() {
    log_info "Deploying PostgreSQL StatefulSet..."
    kubectl apply -f "$PROJECT_ROOT/k8s/statefulset.yaml" -n "$NAMESPACE"

    log_info "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=ready pod -l app=postgres -n "$NAMESPACE" --timeout=300s || true
}

run_migrations() {
    log_info "Running database migrations..."

    # Create migration job
    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration-$(date +%s)
  namespace: $NAMESPACE
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: migration
        image: ghcr.io/OWNER/rtx-backend:$IMAGE_TAG
        command: ["/app/rtx-backend", "migrate", "up"]
        envFrom:
        - configMapRef:
            name: rtx-config
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: rtx-secrets
              key: database-url
  backoffLimit: 3
EOF

    log_info "Waiting for migrations to complete..."
    kubectl wait --for=condition=complete job -l app=db-migration -n "$NAMESPACE" --timeout=300s || log_warn "Migration job timeout"
}

deploy_backend() {
    log_info "Deploying backend application..."

    # Update deployment with new image tag
    kubectl set image deployment/rtx-backend rtx-backend=ghcr.io/OWNER/rtx-backend:$IMAGE_TAG -n "$NAMESPACE" || \
    kubectl apply -f "$PROJECT_ROOT/k8s/deployment.yaml" -n "$NAMESPACE"

    kubectl apply -f "$PROJECT_ROOT/k8s/service.yaml" -n "$NAMESPACE"
    kubectl apply -f "$PROJECT_ROOT/k8s/ingress.yaml" -n "$NAMESPACE"

    log_info "Waiting for deployment to be ready..."
    kubectl rollout status deployment/rtx-backend -n "$NAMESPACE" --timeout=600s
}

deploy_autoscaling() {
    log_info "Deploying HPA..."
    kubectl apply -f "$PROJECT_ROOT/k8s/hpa.yaml" -n "$NAMESPACE"
}

verify_deployment() {
    log_info "Verifying deployment..."

    echo ""
    log_info "Pods:"
    kubectl get pods -n "$NAMESPACE"

    echo ""
    log_info "Services:"
    kubectl get services -n "$NAMESPACE"

    echo ""
    log_info "Ingress:"
    kubectl get ingress -n "$NAMESPACE"

    echo ""
    log_info "HPA:"
    kubectl get hpa -n "$NAMESPACE"
}

run_smoke_tests() {
    log_info "Running smoke tests..."

    # Get service URL
    SERVICE_URL=$(kubectl get ingress rtx-backend-ingress -n "$NAMESPACE" -o jsonpath='{.spec.rules[0].host}')

    if [ -z "$SERVICE_URL" ]; then
        log_warn "Could not determine service URL, skipping smoke tests"
        return
    fi

    log_info "Testing health endpoint..."
    curl -f "https://$SERVICE_URL/health" || log_warn "Health check failed"

    log_info "Testing ready endpoint..."
    curl -f "https://$SERVICE_URL/ready" || log_warn "Ready check failed"
}

cleanup() {
    log_info "Cleaning up old resources..."
    kubectl delete jobs -l app=db-migration -n "$NAMESPACE" --field-selector status.successful=1
}

main() {
    log_info "Starting deployment to $NAMESPACE environment"
    log_info "Image tag: $IMAGE_TAG"

    check_prerequisites
    create_namespace
    apply_configmaps
    apply_secrets
    deploy_database
    run_migrations
    deploy_backend
    deploy_autoscaling
    verify_deployment
    run_smoke_tests
    cleanup

    log_info "Deployment completed successfully!"
}

# Run main function
main "$@"
