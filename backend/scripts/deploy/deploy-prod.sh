#!/bin/bash

set -euo pipefail

# RTX Trading Engine - Production Deployment Script
# This script implements Blue-Green deployment strategy

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
NAMESPACE="rtx-prod"
IMAGE_TAG="${IMAGE_TAG:-latest}"
MIGRATION_REQUIRED="${MIGRATION_REQUIRED:-false}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

log_blue() {
    echo -e "${BLUE}[BLUE]${NC} $1"
}

# Get current active deployment (blue or green)
get_active_deployment() {
    kubectl get service rtx-backend -n "$NAMESPACE" -o jsonpath='{.spec.selector.version}' 2>/dev/null || echo "blue"
}

# Get inactive deployment
get_inactive_deployment() {
    local active=$(get_active_deployment)
    if [ "$active" == "blue" ]; then
        echo "green"
    else
        echo "blue"
    fi
}

pre_deployment_checks() {
    log_info "Running pre-deployment checks..."

    # Check kubectl connection
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
    fi

    # Verify image exists
    log_info "Verifying Docker image exists..."
    if ! docker pull "ghcr.io/OWNER/rtx-backend:$IMAGE_TAG" &> /dev/null; then
        log_error "Docker image ghcr.io/OWNER/rtx-backend:$IMAGE_TAG not found"
    fi

    # Check namespace exists
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_error "Namespace $NAMESPACE does not exist"
    fi

    log_info "Pre-deployment checks passed"
}

backup_deployment() {
    log_info "Backing up current deployments..."
    mkdir -p "$PROJECT_ROOT/backups/production"

    local timestamp=$(date +%Y%m%d-%H%M%S)

    kubectl get deployment rtx-backend-blue -n "$NAMESPACE" -o yaml > "$PROJECT_ROOT/backups/production/blue-$timestamp.yaml" || true
    kubectl get deployment rtx-backend-green -n "$NAMESPACE" -o yaml > "$PROJECT_ROOT/backups/production/green-$timestamp.yaml" || true
    kubectl get service rtx-backend -n "$NAMESPACE" -o yaml > "$PROJECT_ROOT/backups/production/service-$timestamp.yaml" || true

    log_info "Backups saved to $PROJECT_ROOT/backups/production/"
}

run_database_migrations() {
    if [ "$MIGRATION_REQUIRED" != "true" ]; then
        log_info "Skipping database migrations (not required)"
        return
    fi

    log_info "Running database migrations..."
    "$SCRIPT_DIR/run-migrations.sh" production

    if [ $? -ne 0 ]; then
        log_error "Database migration failed"
    fi

    log_info "Database migrations completed successfully"
}

deploy_to_inactive() {
    local inactive=$(get_inactive_deployment)
    local active=$(get_active_deployment)

    log_info "Current active deployment: $active"
    log_info "Deploying to inactive deployment: $inactive"

    # Update inactive deployment with new image
    kubectl set image deployment/rtx-backend-$inactive \
        rtx-backend=ghcr.io/OWNER/rtx-backend:$IMAGE_TAG \
        -n "$NAMESPACE"

    log_info "Waiting for $inactive deployment to be ready..."
    kubectl rollout status deployment/rtx-backend-$inactive -n "$NAMESPACE" --timeout=900s

    if [ $? -ne 0 ]; then
        log_error "Deployment rollout failed"
    fi

    log_info "$inactive deployment is ready"
}

health_check() {
    local deployment=$1
    log_info "Running health check on $deployment deployment..."

    # Get pod IP
    local pod=$(kubectl get pod -n "$NAMESPACE" -l app=rtx-backend,version=$deployment -o jsonpath='{.items[0].metadata.name}')

    if [ -z "$pod" ]; then
        log_error "No pods found for $deployment deployment"
    fi

    # Check health endpoint
    if ! kubectl exec -n "$NAMESPACE" "$pod" -- curl -sf http://localhost:8080/health > /dev/null; then
        log_error "Health check failed for $deployment deployment"
    fi

    log_info "Health check passed for $deployment"
}

run_smoke_tests() {
    local deployment=$1
    log_info "Running smoke tests on $deployment..."

    # Run smoke tests
    if ! "$SCRIPT_DIR/smoke-test.sh" "https://${deployment}.rtx-trading.com"; then
        log_error "Smoke tests failed for $deployment"
    fi

    log_info "Smoke tests passed for $deployment"
}

switch_traffic() {
    local new_active=$1
    local old_active=$2

    log_info "Switching traffic from $old_active to $new_active..."

    # Update service selector to point to new deployment
    kubectl patch service rtx-backend -n "$NAMESPACE" \
        -p "{\"spec\":{\"selector\":{\"version\":\"$new_active\"}}}"

    log_info "Traffic switched to $new_active"

    # Wait for 30 seconds to allow connections to drain
    log_info "Waiting for connections to drain..."
    sleep 30
}

monitor_metrics() {
    log_info "Monitoring metrics for 5 minutes..."

    # Monitor for 5 minutes
    for i in {1..30}; do
        echo -n "."
        sleep 10
    done
    echo ""

    log_info "Metrics monitoring complete"
}

rollback() {
    local target_deployment=$1
    log_error "Rolling back to $target_deployment..."

    kubectl patch service rtx-backend -n "$NAMESPACE" \
        -p "{\"spec\":{\"selector\":{\"version\":\"$target_deployment\"}}}"

    log_info "Rollback complete. Service pointing to $target_deployment"
}

update_old_deployment() {
    local deployment=$1
    log_info "Updating $deployment to match current active deployment..."

    kubectl set image deployment/rtx-backend-$deployment \
        rtx-backend=ghcr.io/OWNER/rtx-backend:$IMAGE_TAG \
        -n "$NAMESPACE"

    kubectl rollout status deployment/rtx-backend-$deployment -n "$NAMESPACE" --timeout=900s

    log_info "$deployment updated successfully"
}

verify_deployment() {
    log_info "Verifying deployment..."

    echo ""
    log_info "Deployments:"
    kubectl get deployments -n "$NAMESPACE" -l app=rtx-backend

    echo ""
    log_info "Pods:"
    kubectl get pods -n "$NAMESPACE" -l app=rtx-backend

    echo ""
    log_info "Services:"
    kubectl get services -n "$NAMESPACE"

    echo ""
    log_info "Current active deployment: $(get_active_deployment)"
}

main() {
    log_info "=== RTX Trading Engine - Production Deployment ==="
    log_info "Image tag: $IMAGE_TAG"
    log_info "Migration required: $MIGRATION_REQUIRED"

    # Get current state
    local active_deployment=$(get_active_deployment)
    local inactive_deployment=$(get_inactive_deployment)

    log_info "Current active: $active_deployment"
    log_info "Deploying to: $inactive_deployment"

    # Deployment steps
    pre_deployment_checks
    backup_deployment
    run_database_migrations
    deploy_to_inactive
    health_check "$inactive_deployment"
    run_smoke_tests "$inactive_deployment"

    # Confirm before switching traffic
    read -p "Switch traffic to $inactive_deployment? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        log_warn "Deployment cancelled by user"
        exit 0
    fi

    switch_traffic "$inactive_deployment" "$active_deployment"
    monitor_metrics

    # Check if everything is working
    if ! health_check "$inactive_deployment"; then
        rollback "$active_deployment"
        log_error "Deployment failed. Rolled back to $active_deployment"
    fi

    # Update old deployment
    update_old_deployment "$active_deployment"

    verify_deployment

    log_info "=== Production deployment completed successfully! ==="
    log_info "Active deployment: $inactive_deployment"
    log_info "Image: ghcr.io/OWNER/rtx-backend:$IMAGE_TAG"
}

main "$@"
