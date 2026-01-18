#!/bin/bash

set -e

# Configuration
NAMESPACE="staging"
DEPLOYMENT_NAME="trading-engine"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ECR_REGISTRY="${ECR_REGISTRY}"
MAX_WAIT_TIME=600  # 10 minutes

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    if ! command -v kubectl &> /dev/null; then
        error "kubectl not found. Please install kubectl."
        exit 1
    fi

    if ! command -v aws &> /dev/null; then
        error "aws cli not found. Please install AWS CLI."
        exit 1
    fi

    if [ -z "$ECR_REGISTRY" ]; then
        error "ECR_REGISTRY environment variable is not set"
        exit 1
    fi

    if [ -z "$IMAGE_TAG" ]; then
        error "IMAGE_TAG environment variable is not set"
        exit 1
    fi

    log "Prerequisites check passed"
}

# Create namespace if it doesn't exist
create_namespace() {
    log "Creating namespace ${NAMESPACE} if it doesn't exist..."
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
}

# Apply Kubernetes manifests
apply_manifests() {
    log "Applying Kubernetes manifests..."

    # Apply in order
    kubectl apply -f deployments/kubernetes/rbac.yaml -n ${NAMESPACE}
    kubectl apply -f deployments/kubernetes/configmap.yaml -n ${NAMESPACE}
    kubectl apply -f deployments/kubernetes/secrets.yaml -n ${NAMESPACE}
    kubectl apply -f deployments/kubernetes/pvc.yaml -n ${NAMESPACE}

    # Update deployment with new image
    export ECR_REGISTRY IMAGE_TAG
    envsubst < deployments/kubernetes/deployment.yaml | kubectl apply -f - -n ${NAMESPACE}

    kubectl apply -f deployments/kubernetes/service.yaml -n ${NAMESPACE}
    kubectl apply -f deployments/kubernetes/ingress.yaml -n ${NAMESPACE}
    kubectl apply -f deployments/kubernetes/hpa.yaml -n ${NAMESPACE}

    log "Manifests applied successfully"
}

# Wait for deployment rollout
wait_for_rollout() {
    log "Waiting for deployment rollout..."

    if kubectl rollout status deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE} --timeout=${MAX_WAIT_TIME}s; then
        log "Deployment rolled out successfully"
        return 0
    else
        error "Deployment rollout failed or timed out"
        return 1
    fi
}

# Get deployment status
get_deployment_status() {
    log "Deployment status:"
    kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE}

    log "Pod status:"
    kubectl get pods -l app=${DEPLOYMENT_NAME} -n ${NAMESPACE}

    log "Recent events:"
    kubectl get events -n ${NAMESPACE} --sort-by='.lastTimestamp' | tail -10
}

# Check pod health
check_pod_health() {
    log "Checking pod health..."

    READY_PODS=$(kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE} -o jsonpath='{.status.readyReplicas}')
    DESIRED_PODS=$(kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.replicas}')

    if [ "$READY_PODS" = "$DESIRED_PODS" ]; then
        log "All pods are ready (${READY_PODS}/${DESIRED_PODS})"
        return 0
    else
        warn "Not all pods are ready (${READY_PODS}/${DESIRED_PODS})"
        return 1
    fi
}

# Main deployment flow
main() {
    log "Starting deployment to ${NAMESPACE}..."
    log "Image: ${ECR_REGISTRY}/trading-engine:${IMAGE_TAG}"

    check_prerequisites
    create_namespace
    apply_manifests

    if ! wait_for_rollout; then
        error "Deployment failed!"
        get_deployment_status
        exit 1
    fi

    if ! check_pod_health; then
        warn "Pod health check failed, but deployment succeeded"
    fi

    get_deployment_status
    log "Deployment to ${NAMESPACE} completed successfully!"
}

main "$@"
