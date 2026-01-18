#!/bin/bash

set -euo pipefail

# RTX Trading Engine - Database Migration Runner
# This script runs database migrations in Kubernetes

ENVIRONMENT="${1:-development}"
NAMESPACE=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

set_namespace() {
    case "$ENVIRONMENT" in
        development|dev)
            NAMESPACE="rtx-dev"
            ;;
        staging)
            NAMESPACE="rtx-staging"
            ;;
        production|prod)
            NAMESPACE="rtx-prod"
            ;;
        *)
            log_error "Unknown environment: $ENVIRONMENT"
            exit 1
            ;;
    esac
}

get_migration_status() {
    log_info "Checking current migration status..."

    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration-status-$(date +%s)
  namespace: $NAMESPACE
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: migration
        image: ghcr.io/OWNER/rtx-backend:latest
        command: ["/app/rtx-backend", "migrate", "status"]
        envFrom:
        - configMapRef:
            name: rtx-config
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: rtx-secrets
              key: database-url
  backoffLimit: 1
EOF

    local job_name=$(kubectl get job -n "$NAMESPACE" -l app=db-migration-status --sort-by=.metadata.creationTimestamp -o jsonpath='{.items[-1].metadata.name}')

    kubectl wait --for=condition=complete job "$job_name" -n "$NAMESPACE" --timeout=120s

    kubectl logs job/"$job_name" -n "$NAMESPACE"
}

run_migrations() {
    log_info "Running database migrations..."

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
        image: ghcr.io/OWNER/rtx-backend:latest
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
      initContainers:
      - name: wait-for-db
        image: busybox:1.36
        command:
        - sh
        - -c
        - |
          until nc -z postgres-service 5432; do
            echo "Waiting for database..."
            sleep 2
          done
  backoffLimit: 3
EOF

    local job_name=$(kubectl get job -n "$NAMESPACE" -l app=db-migration --sort-by=.metadata.creationTimestamp -o jsonpath='{.items[-1].metadata.name}')

    log_info "Waiting for migrations to complete..."
    kubectl wait --for=condition=complete job "$job_name" -n "$NAMESPACE" --timeout=300s

    if [ $? -eq 0 ]; then
        log_info "Migrations completed successfully"
        kubectl logs job/"$job_name" -n "$NAMESPACE"
    else
        log_error "Migrations failed"
        kubectl logs job/"$job_name" -n "$NAMESPACE"
        exit 1
    fi
}

cleanup_migration_jobs() {
    log_info "Cleaning up old migration jobs..."
    kubectl delete jobs -n "$NAMESPACE" -l app=db-migration --field-selector status.successful=1
}

main() {
    set_namespace

    log_info "Running migrations for $ENVIRONMENT environment"
    log_info "Namespace: $NAMESPACE"

    get_migration_status
    run_migrations
    cleanup_migration_jobs

    log_info "Migration process completed"
}

main "$@"
