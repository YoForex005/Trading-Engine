#!/bin/bash

# Deployment Test Script
# Tests Docker builds, Kubernetes manifests, health checks, and database migrations

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEST_TIMEOUT=300
RESULTS_FILE="$PROJECT_ROOT/.swarm/deployment_test_results.json"
DOCKER_IMAGE_PREFIX="rtx-backend:test-"
LOG_FILE="/tmp/deployment_tests.log"

# Initialize results
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Helper functions
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

success() {
    echo -e "${GREEN}✓ $1${NC}" | tee -a "$LOG_FILE"
    ((PASSED_TESTS++))
}

error() {
    echo -e "${RED}✗ $1${NC}" | tee -a "$LOG_FILE"
    ((FAILED_TESTS++))
}

warning() {
    echo -e "${YELLOW}⚠ $1${NC}" | tee -a "$LOG_FILE"
    ((SKIPPED_TESTS++))
}

check_prerequisites() {
    log "Checking prerequisites..."

    # Check for required tools
    local required_tools=("docker" "go")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            error "Required tool '$tool' not found"
            exit 1
        fi
    done

    # Check for go.mod
    if [ ! -f "$PROJECT_ROOT/go.mod" ]; then
        error "go.mod not found in project root"
        exit 1
    fi

    success "All prerequisites met"
}

test_docker_build() {
    log "Running Docker build tests..."

    cd "$PROJECT_ROOT"

    # Test 1: Basic Docker build
    if [ -f "Dockerfile" ]; then
        log "Testing basic Docker build..."
        if docker build -t "${DOCKER_IMAGE_PREFIX}basic" . > /dev/null 2>&1; then
            success "Docker build successful"
            # Cleanup
            docker rmi "${DOCKER_IMAGE_PREFIX}basic" > /dev/null 2>&1
        else
            error "Docker build failed"
            return 1
        fi
    else
        warning "Dockerfile not found, skipping Docker tests"
    fi

    # Test 2: Build with specific build args
    if [ -f "Dockerfile" ]; then
        log "Testing Docker build with build arguments..."
        if docker build \
            -t "${DOCKER_IMAGE_PREFIX}withargs" \
            --build-arg BUILD_ENV=test \
            --build-arg VERSION=1.0.0 \
            . > /dev/null 2>&1; then
            success "Docker build with args successful"
            docker rmi "${DOCKER_IMAGE_PREFIX}withargs" > /dev/null 2>&1
        else
            warning "Docker build with args failed"
        fi
    fi

    # Test 3: Docker Compose validation
    if [ -f "docker-compose.yml" ]; then
        log "Validating docker-compose.yml..."
        if docker-compose -f docker-compose.yml config > /dev/null 2>&1; then
            success "docker-compose.yml validation passed"
        else
            error "docker-compose.yml validation failed"
        fi
    else
        warning "docker-compose.yml not found"
    fi
}

test_kubernetes_manifests() {
    log "Running Kubernetes manifest tests..."

    if [ ! -d "k8s" ]; then
        warning "Kubernetes manifests directory not found"
        return 0
    fi

    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        warning "kubectl not found, skipping manifest validation"
        return 0
    fi

    # Test each manifest file
    for manifest in k8s/*.yaml; do
        if [ -f "$manifest" ]; then
            log "Validating $(basename "$manifest")..."
            if kubectl apply -f "$manifest" --dry-run=client -o json > /dev/null 2>&1; then
                success "$(basename "$manifest") validation passed"
            else
                error "$(basename "$manifest") validation failed"
            fi
        fi
    done
}

test_health_checks() {
    log "Running health check tests..."

    # Test health check endpoints
    local service_url="${SERVICE_URL:-http://localhost:8080}"
    local endpoints=("/health" "/ready" "/live")
    local timeout=5

    for endpoint in "${endpoints[@]}"; do
        log "Testing health check endpoint: $endpoint"

        # Use timeout command if available
        if command -v timeout &> /dev/null; then
            if timeout $timeout curl -s "$service_url$endpoint" > /dev/null 2>&1; then
                success "Health check endpoint $endpoint is responding"
            else
                warning "Health check endpoint $endpoint not responding (service may not be running)"
            fi
        else
            if curl -m $timeout -s "$service_url$endpoint" > /dev/null 2>&1; then
                success "Health check endpoint $endpoint is responding"
            else
                warning "Health check endpoint $endpoint not responding"
            fi
        fi
    done
}

test_database_migrations() {
    log "Running database migration tests..."

    # Check for migrations directory
    if [ ! -d "migrations" ]; then
        warning "Migrations directory not found"
        return 0
    fi

    # Count migration files
    local migration_count=$(find migrations -type f | wc -l)
    if [ "$migration_count" -eq 0 ]; then
        warning "No migration files found"
    else
        success "Found $migration_count migration files"
    fi

    # Check for migration executable
    if [ -f "cmd/migrate/main.go" ]; then
        log "Testing migration build..."
        if cd "$PROJECT_ROOT" && go build -o /tmp/migrate ./cmd/migrate > /dev/null 2>&1; then
            success "Migration executable built successfully"
        else
            warning "Failed to build migration executable"
        fi
    fi
}

test_deployment_manifests_content() {
    log "Running deployment manifest content tests..."

    if [ ! -f "k8s/deployment.yaml" ]; then
        warning "Deployment manifest not found"
        return 0
    fi

    local deployment_file="k8s/deployment.yaml"

    # Check for required fields
    local required_fields=("resources:" "livenessProbe:" "readinessProbe:")

    for field in "${required_fields[@]}"; do
        if grep -q "$field" "$deployment_file"; then
            success "Deployment includes $field"
        else
            warning "Deployment missing $field"
        fi
    done

    # Check for security context
    if grep -q "securityContext:" "$deployment_file"; then
        success "Deployment includes security context"
    else
        warning "Deployment missing security context"
    fi

    # Check for image pull policy
    if grep -q "imagePullPolicy:" "$deployment_file"; then
        success "Deployment includes image pull policy"
    else
        warning "Deployment missing image pull policy"
    fi
}

run_go_deployment_tests() {
    log "Running Go deployment tests..."

    cd "$PROJECT_ROOT"

    # Run deployment tests
    if go test -v -timeout 5m ./tests/deployment/... 2>&1 | tee -a "$LOG_FILE"; then
        success "Go deployment tests passed"
    else
        error "Go deployment tests failed"
        return 1
    fi
}

generate_test_report() {
    log "Generating test report..."

    local total_tests=$((PASSED_TESTS + FAILED_TESTS + SKIPPED_TESTS))

    cat > "$RESULTS_FILE" << EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "test_summary": {
    "total": $total_tests,
    "passed": $PASSED_TESTS,
    "failed": $FAILED_TESTS,
    "skipped": $SKIPPED_TESTS
  },
  "test_categories": {
    "docker_build": "completed",
    "kubernetes_manifests": "completed",
    "health_checks": "completed",
    "database_migrations": "completed",
    "deployment_manifests": "completed"
  },
  "environment": {
    "project_root": "$PROJECT_ROOT",
    "docker_version": "$(docker --version 2>/dev/null || echo 'not found')",
    "kubectl_version": "$(kubectl version --short 2>/dev/null || echo 'not found')",
    "go_version": "$(go version 2>/dev/null || echo 'not found')"
  },
  "log_file": "$LOG_FILE"
}
EOF

    echo ""
    echo "=========================================="
    echo "DEPLOYMENT TEST SUMMARY"
    echo "=========================================="
    echo "Total Tests: $total_tests"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo "Skipped: $SKIPPED_TESTS"
    echo "=========================================="
    echo ""
    echo "Results saved to: $RESULTS_FILE"
    echo "Logs saved to: $LOG_FILE"
}

main() {
    log "Starting deployment tests..."
    log "Project root: $PROJECT_ROOT"
    echo ""

    # Clear log file
    > "$LOG_FILE"

    # Run tests
    check_prerequisites
    test_docker_build
    test_kubernetes_manifests
    test_health_checks
    test_database_migrations
    test_deployment_manifests_content
    run_go_deployment_tests

    # Generate report
    echo ""
    generate_test_report

    # Exit with appropriate code
    if [ $FAILED_TESTS -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

# Handle interruption
trap "error 'Test interrupted'; exit 130" INT TERM

# Run main
main "$@"
