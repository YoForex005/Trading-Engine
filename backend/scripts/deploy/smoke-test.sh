#!/bin/bash

set -euo pipefail

# RTX Trading Engine - Smoke Test Script
# This script runs basic smoke tests against a deployed environment

URL="${1:-http://localhost:8080}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

test_health_endpoint() {
    log_info "Testing health endpoint..."
    if curl -sf "$URL/health" > /dev/null; then
        log_pass "Health endpoint responding"
    else
        log_fail "Health endpoint not responding"
    fi
}

test_ready_endpoint() {
    log_info "Testing ready endpoint..."
    if curl -sf "$URL/ready" > /dev/null; then
        log_pass "Ready endpoint responding"
    else
        log_fail "Ready endpoint not responding"
    fi
}

test_metrics_endpoint() {
    log_info "Testing metrics endpoint..."
    if curl -sf "$URL:9090/metrics" > /dev/null; then
        log_pass "Metrics endpoint responding"
    else
        log_fail "Metrics endpoint not responding"
    fi
}

test_api_endpoints() {
    log_info "Testing API endpoints..."

    # Test root endpoint
    if curl -sf "$URL/api/v1" > /dev/null; then
        log_pass "API root endpoint responding"
    else
        log_fail "API root endpoint not responding"
    fi
}

test_response_time() {
    log_info "Testing response time..."
    local response_time=$(curl -o /dev/null -s -w '%{time_total}' "$URL/health")
    local threshold=1.0

    if (( $(echo "$response_time < $threshold" | bc -l) )); then
        log_pass "Response time acceptable: ${response_time}s"
    else
        log_fail "Response time too slow: ${response_time}s (threshold: ${threshold}s)"
    fi
}

test_ssl() {
    if [[ $URL == https* ]]; then
        log_info "Testing SSL certificate..."
        if curl -sf --max-time 5 "$URL/health" > /dev/null; then
            log_pass "SSL certificate valid"
        else
            log_fail "SSL certificate invalid or expired"
        fi
    else
        log_info "Skipping SSL test (HTTP URL)"
    fi
}

print_summary() {
    echo ""
    echo "==================================="
    echo "Smoke Test Summary"
    echo "==================================="
    echo -e "Passed: ${GREEN}$PASSED${NC}"
    echo -e "Failed: ${RED}$FAILED${NC}"
    echo "==================================="

    if [ $FAILED -gt 0 ]; then
        exit 1
    fi
}

main() {
    log_info "Running smoke tests against: $URL"
    echo ""

    test_health_endpoint
    test_ready_endpoint
    test_metrics_endpoint
    test_api_endpoints
    test_response_time
    test_ssl

    print_summary
}

main "$@"
