#!/bin/bash

set -e

# Configuration
ENVIRONMENT="${1:-staging}"
MAX_RETRIES=5
RETRY_DELAY=10

# URLs based on environment
if [ "$ENVIRONMENT" = "production" ]; then
    BASE_URL="${PRODUCTION_URL:-https://api.trading-engine.example.com}"
elif [ "$ENVIRONMENT" = "staging" ]; then
    BASE_URL="${STAGING_URL:-https://staging.trading-engine.example.com}"
else
    BASE_URL="http://localhost:8080"
fi

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"; }

# Check health endpoint
check_health() {
    local url="$1"
    local retries=0

    while [ $retries -lt $MAX_RETRIES ]; do
        if curl -f -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200"; then
            return 0
        fi

        retries=$((retries + 1))
        if [ $retries -lt $MAX_RETRIES ]; then
            warn "Health check failed, retrying in ${RETRY_DELAY}s... (${retries}/${MAX_RETRIES})"
            sleep $RETRY_DELAY
        fi
    done

    return 1
}

# Validate response time
check_response_time() {
    local url="$1"
    local max_time=2  # 2 seconds

    local response_time=$(curl -o /dev/null -s -w '%{time_total}' "$url")
    local response_time_ms=$(echo "$response_time * 1000" | bc)

    log "Response time: ${response_time_ms}ms"

    if (( $(echo "$response_time > $max_time" | bc -l) )); then
        warn "Response time exceeds threshold (${response_time}s > ${max_time}s)"
        return 1
    fi

    return 0
}

# Check database connectivity
check_database() {
    local health_url="${BASE_URL}/health/db"

    log "Checking database connectivity..."
    if check_health "$health_url"; then
        log "Database check passed"
        return 0
    else
        error "Database check failed"
        return 1
    fi
}

# Check Redis connectivity
check_redis() {
    local health_url="${BASE_URL}/health/redis"

    log "Checking Redis connectivity..."
    if check_health "$health_url"; then
        log "Redis check passed"
        return 0
    else
        error "Redis check failed"
        return 1
    fi
}

# API smoke tests
run_smoke_tests() {
    log "Running smoke tests..."

    # Test 1: Get server info
    log "Test 1: Server info..."
    if ! curl -f -s "${BASE_URL}/api/v1/info" | jq -e '.version' > /dev/null; then
        error "Server info test failed"
        return 1
    fi
    log "✓ Server info test passed"

    # Test 2: Authentication endpoint
    log "Test 2: Authentication..."
    if ! curl -f -s -o /dev/null "${BASE_URL}/api/v1/auth/health"; then
        error "Authentication test failed"
        return 1
    fi
    log "✓ Authentication test passed"

    # Test 3: Metrics endpoint
    log "Test 3: Metrics..."
    if ! curl -f -s -o /dev/null "${BASE_URL}/metrics"; then
        error "Metrics test failed"
        return 1
    fi
    log "✓ Metrics test passed"

    log "All smoke tests passed"
    return 0
}

# Check critical metrics
check_metrics() {
    log "Checking critical metrics..."

    local metrics=$(curl -s "${BASE_URL}/metrics")

    # Check error rate
    local error_rate=$(echo "$metrics" | grep 'http_requests_total{.*status="5.."}' | awk '{sum+=$2} END {print sum}')
    if [ -n "$error_rate" ] && [ "$error_rate" -gt 10 ]; then
        warn "High error rate detected: $error_rate"
    fi

    # Check response time p95
    local p95=$(echo "$metrics" | grep 'http_request_duration_seconds{quantile="0.95"}' | awk '{print $2}')
    if [ -n "$p95" ]; then
        log "P95 response time: ${p95}s"
    fi

    log "Metrics check completed"
}

# Check deployment status
check_deployment_status() {
    if [ "$ENVIRONMENT" != "local" ]; then
        log "Checking Kubernetes deployment status..."

        local namespace="${ENVIRONMENT}"
        local ready_pods=$(kubectl get deployment trading-engine -n ${namespace} -o jsonpath='{.status.readyReplicas}')
        local desired_pods=$(kubectl get deployment trading-engine -n ${namespace} -o jsonpath='{.spec.replicas}')

        if [ "$ready_pods" = "$desired_pods" ]; then
            log "All pods are ready (${ready_pods}/${desired_pods})"
        else
            error "Not all pods are ready (${ready_pods}/${desired_pods})"
            return 1
        fi
    fi
}

# Main health check flow
main() {
    log "Starting health checks for ${ENVIRONMENT}..."
    log "Target URL: ${BASE_URL}"

    local failed=0

    # Basic health check
    log "Checking basic health endpoint..."
    if check_health "${BASE_URL}/health"; then
        log "✓ Basic health check passed"
    else
        error "✗ Basic health check failed"
        failed=$((failed + 1))
    fi

    # Response time check
    log "Checking response time..."
    if check_response_time "${BASE_URL}/health"; then
        log "✓ Response time check passed"
    else
        warn "✗ Response time check failed"
    fi

    # Database check
    if check_database; then
        log "✓ Database check passed"
    else
        error "✗ Database check failed"
        failed=$((failed + 1))
    fi

    # Redis check
    if check_redis; then
        log "✓ Redis check passed"
    else
        error "✗ Redis check failed"
        failed=$((failed + 1))
    fi

    # Smoke tests
    if run_smoke_tests; then
        log "✓ Smoke tests passed"
    else
        error "✗ Smoke tests failed"
        failed=$((failed + 1))
    fi

    # Metrics check
    check_metrics

    # Deployment status
    if check_deployment_status; then
        log "✓ Deployment status check passed"
    else
        error "✗ Deployment status check failed"
        failed=$((failed + 1))
    fi

    # Summary
    echo ""
    if [ $failed -eq 0 ]; then
        log "========================================="
        log "All health checks passed! ✓"
        log "========================================="
        exit 0
    else
        error "========================================="
        error "${failed} health check(s) failed! ✗"
        error "========================================="
        exit 1
    fi
}

main "$@"
