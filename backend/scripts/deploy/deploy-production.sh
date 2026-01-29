#!/bin/bash

set -e

# Configuration
NAMESPACE="production"
DEPLOYMENT_NAME="trading-engine"
IMAGE_TAG="${IMAGE_TAG:-latest}"
ECR_REGISTRY="${ECR_REGISTRY}"
DEPLOYMENT_STRATEGY="${DEPLOYMENT_STRATEGY:-blue-green}"
MAX_WAIT_TIME=600

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO:${NC} $1"; }

# Safety checks
safety_checks() {
    log "Running safety checks..."

    # Check if in production namespace
    CURRENT_NS=$(kubectl config view --minify --output 'jsonpath={..namespace}')
    if [ "$CURRENT_NS" != "$NAMESPACE" ]; then
        warn "Not in production namespace. Switching to ${NAMESPACE}"
        kubectl config set-context --current --namespace=${NAMESPACE}
    fi

    # Verify production exists
    if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
        error "Production namespace does not exist!"
        exit 1
    fi

    # Check cluster connectivity
    if ! kubectl cluster-info &> /dev/null; then
        error "Cannot connect to Kubernetes cluster"
        exit 1
    fi

    # Require confirmation for production
    read -p "Are you sure you want to deploy to PRODUCTION? (yes/no): " CONFIRM
    if [ "$CONFIRM" != "yes" ]; then
        error "Deployment cancelled by user"
        exit 1
    fi

    log "Safety checks passed"
}

# Backup current deployment
backup_deployment() {
    log "Backing up current deployment..."

    BACKUP_FILE="backup-${DEPLOYMENT_NAME}-$(date +%Y%m%d-%H%M%S).yaml"
    kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE} -o yaml > ${BACKUP_FILE}

    log "Backup saved to ${BACKUP_FILE}"
}

# Blue-Green Deployment
blue_green_deploy() {
    log "Starting Blue-Green deployment..."

    # Create green deployment
    info "Creating green deployment..."
    export ECR_REGISTRY IMAGE_TAG
    cat deployments/kubernetes/deployment.yaml | \
        sed 's/trading-engine/trading-engine-green/g' | \
        envsubst | \
        kubectl apply -f - -n ${NAMESPACE}

    # Wait for green to be ready
    info "Waiting for green deployment to be ready..."
    kubectl rollout status deployment/trading-engine-green -n ${NAMESPACE} --timeout=${MAX_WAIT_TIME}s

    # Run smoke tests on green
    info "Running smoke tests on green deployment..."
    if ! ./scripts/deploy/health-check.sh green; then
        error "Smoke tests failed on green deployment"
        kubectl delete deployment trading-engine-green -n ${NAMESPACE}
        exit 1
    fi

    # Switch traffic to green
    info "Switching traffic to green deployment..."
    kubectl patch service ${DEPLOYMENT_NAME} -n ${NAMESPACE} \
        -p '{"spec":{"selector":{"app":"trading-engine-green"}}}'

    # Wait and monitor
    sleep 60

    # If successful, delete blue
    info "Deleting blue deployment..."
    kubectl delete deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE}

    # Rename green to blue
    kubectl patch deployment trading-engine-green -n ${NAMESPACE} \
        --type=json \
        -p='[{"op": "replace", "path": "/metadata/name", "value":"trading-engine"}]'

    log "Blue-Green deployment completed"
}

# Canary Deployment
canary_deploy() {
    log "Starting Canary deployment..."

    # Deploy canary with 10% traffic
    info "Deploying canary version..."
    export ECR_REGISTRY IMAGE_TAG
    envsubst < deployments/kubernetes/deployment.yaml | kubectl apply -f - -n ${NAMESPACE}

    # Scale canary to 1 replica
    kubectl scale deployment trading-engine-canary --replicas=1 -n ${NAMESPACE}

    # Wait for canary to be ready
    kubectl rollout status deployment/trading-engine-canary -n ${NAMESPACE} --timeout=${MAX_WAIT_TIME}s

    # Set canary ingress weight to 10%
    kubectl patch ingress trading-engine-ingress-canary -n ${NAMESPACE} \
        --type=json \
        -p='[{"op": "replace", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight", "value":"10"}]'

    info "Canary deployed with 10% traffic. Monitoring for 5 minutes..."
    sleep 300

    # Check canary metrics
    if ! ./scripts/deploy/canary-metrics.sh; then
        error "Canary metrics check failed"
        rollback_canary
        exit 1
    fi

    # Gradually increase traffic: 10% -> 25% -> 50% -> 100%
    for weight in 25 50 100; do
        info "Increasing canary traffic to ${weight}%..."
        kubectl patch ingress trading-engine-ingress-canary -n ${NAMESPACE} \
            --type=json \
            -p="[{\"op\": \"replace\", \"path\": \"/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight\", \"value\":\"${weight}\"}]"

        sleep 120

        if ! ./scripts/deploy/canary-metrics.sh; then
            error "Canary metrics check failed at ${weight}%"
            rollback_canary
            exit 1
        fi
    done

    # Promote canary to production
    info "Promoting canary to production..."
    kubectl set image deployment/${DEPLOYMENT_NAME} \
        app=${ECR_REGISTRY}/trading-engine:${IMAGE_TAG} \
        -n ${NAMESPACE}

    kubectl rollout status deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE}

    # Scale down canary
    kubectl scale deployment trading-engine-canary --replicas=0 -n ${NAMESPACE}

    log "Canary deployment completed"
}

# Rollback canary
rollback_canary() {
    warn "Rolling back canary deployment..."
    kubectl scale deployment trading-engine-canary --replicas=0 -n ${NAMESPACE}
    kubectl patch ingress trading-engine-ingress-canary -n ${NAMESPACE} \
        --type=json \
        -p='[{"op": "replace", "path": "/metadata/annotations/nginx.ingress.kubernetes.io~1canary-weight", "value":"0"}]'
}

# Rolling Update Deployment
rolling_deploy() {
    log "Starting Rolling Update deployment..."

    export ECR_REGISTRY IMAGE_TAG
    envsubst < deployments/kubernetes/deployment.yaml | kubectl apply -f - -n ${NAMESPACE}

    kubectl rollout status deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE} --timeout=${MAX_WAIT_TIME}s

    log "Rolling update completed"
}

# Main deployment flow
main() {
    log "Starting production deployment..."
    log "Strategy: ${DEPLOYMENT_STRATEGY}"
    log "Image: ${ECR_REGISTRY}/trading-engine:${IMAGE_TAG}"

    safety_checks
    backup_deployment

    case "$DEPLOYMENT_STRATEGY" in
        blue-green)
            blue_green_deploy
            ;;
        canary)
            canary_deploy
            ;;
        rolling)
            rolling_deploy
            ;;
        *)
            error "Unknown deployment strategy: $DEPLOYMENT_STRATEGY"
            exit 1
            ;;
    esac

    # Run final health checks
    log "Running post-deployment health checks..."
    if ! ./scripts/deploy/health-check.sh production; then
        error "Post-deployment health check failed!"
        ./scripts/deploy/rollback.sh production
        exit 1
    fi

    log "Production deployment completed successfully!"
    log "Deployed version: ${IMAGE_TAG}"
}

main "$@"
