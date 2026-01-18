#!/bin/bash

set -e

# Configuration
ENVIRONMENT="${1:-staging}"
NAMESPACE="${ENVIRONMENT}"
DEPLOYMENT_NAME="trading-engine"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"; }

# Verify environment
verify_environment() {
    log "Verifying rollback to ${ENVIRONMENT}..."

    if [ "$ENVIRONMENT" != "staging" ] && [ "$ENVIRONMENT" != "production" ]; then
        error "Invalid environment: $ENVIRONMENT"
        exit 1
    fi

    if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
        error "Namespace ${NAMESPACE} does not exist"
        exit 1
    fi
}

# Find previous revision
find_previous_revision() {
    log "Finding previous revision..."

    CURRENT_REVISION=$(kubectl rollout history deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE} | tail -2 | head -1 | awk '{print $1}')
    PREVIOUS_REVISION=$((CURRENT_REVISION - 1))

    if [ $PREVIOUS_REVISION -lt 1 ]; then
        error "No previous revision found"
        exit 1
    fi

    log "Current revision: ${CURRENT_REVISION}"
    log "Rolling back to revision: ${PREVIOUS_REVISION}"
}

# Rollback deployment
rollback_deployment() {
    log "Rolling back deployment..."

    # Undo last rollout
    if kubectl rollout undo deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE}; then
        log "Rollback initiated successfully"
    else
        error "Failed to initiate rollback"
        exit 1
    fi
}

# Wait for rollback to complete
wait_for_rollback() {
    log "Waiting for rollback to complete..."

    if kubectl rollout status deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE} --timeout=600s; then
        log "Rollback completed successfully"
    else
        error "Rollback failed or timed out"
        exit 1
    fi
}

# Verify rollback
verify_rollback() {
    log "Verifying rollback..."

    # Check pod status
    READY_PODS=$(kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE} -o jsonpath='{.status.readyReplicas}')
    DESIRED_PODS=$(kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.replicas}')

    if [ "$READY_PODS" = "$DESIRED_PODS" ]; then
        log "All pods are ready (${READY_PODS}/${DESIRED_PODS})"
    else
        warn "Not all pods are ready (${READY_PODS}/${DESIRED_PODS})"
    fi

    # Run health checks
    log "Running health checks..."
    if ./scripts/deploy/health-check.sh ${ENVIRONMENT}; then
        log "Health checks passed"
    else
        error "Health checks failed after rollback"
        exit 1
    fi
}

# Show rollback history
show_history() {
    log "Deployment history:"
    kubectl rollout history deployment/${DEPLOYMENT_NAME} -n ${NAMESPACE}

    log "Current deployment status:"
    kubectl get deployment ${DEPLOYMENT_NAME} -n ${NAMESPACE}

    log "Pod status:"
    kubectl get pods -l app=${DEPLOYMENT_NAME} -n ${NAMESPACE}
}

# Send notification
send_notification() {
    warn "ROLLBACK EXECUTED: ${ENVIRONMENT} environment rolled back to previous version"

    # Send Slack notification if webhook is configured
    if [ -n "$SLACK_WEBHOOK" ]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"🔴 ROLLBACK: ${ENVIRONMENT} deployment rolled back to previous version\"}" \
            $SLACK_WEBHOOK
    fi
}

# Main rollback flow
main() {
    log "Starting rollback for ${ENVIRONMENT}..."

    verify_environment
    find_previous_revision

    # Require confirmation for production
    if [ "$ENVIRONMENT" = "production" ]; then
        read -p "Are you sure you want to rollback PRODUCTION? (yes/no): " CONFIRM
        if [ "$CONFIRM" != "yes" ]; then
            error "Rollback cancelled by user"
            exit 1
        fi
    fi

    rollback_deployment
    wait_for_rollback
    verify_rollback
    show_history
    send_notification

    log "Rollback for ${ENVIRONMENT} completed successfully!"
}

main "$@"
