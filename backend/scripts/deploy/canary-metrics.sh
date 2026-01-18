#!/bin/bash

set -e

# Configuration
NAMESPACE="production"
CANARY_DEPLOYMENT="trading-engine-canary"
STABLE_DEPLOYMENT="trading-engine"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://prometheus:9090}"
THRESHOLD_ERROR_RATE=0.01  # 1%
THRESHOLD_LATENCY_P95=2.0  # 2 seconds

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"; }

# Query Prometheus
query_prometheus() {
    local query="$1"
    local result=$(curl -s -G --data-urlencode "query=${query}" "${PROMETHEUS_URL}/api/v1/query" | jq -r '.data.result[0].value[1]')
    echo "$result"
}

# Compare error rates
check_error_rate() {
    log "Checking error rates..."

    # Canary error rate
    local canary_errors=$(query_prometheus "rate(http_requests_total{deployment=\"${CANARY_DEPLOYMENT}\",status=~\"5..\"}[5m])")
    local canary_total=$(query_prometheus "rate(http_requests_total{deployment=\"${CANARY_DEPLOYMENT}\"}[5m])")
    local canary_rate=$(echo "scale=4; $canary_errors / $canary_total" | bc)

    # Stable error rate
    local stable_errors=$(query_prometheus "rate(http_requests_total{deployment=\"${STABLE_DEPLOYMENT}\",status=~\"5..\"}[5m])")
    local stable_total=$(query_prometheus "rate(http_requests_total{deployment=\"${STABLE_DEPLOYMENT}\"}[5m])")
    local stable_rate=$(echo "scale=4; $stable_errors / $stable_total" | bc)

    log "Canary error rate: ${canary_rate}"
    log "Stable error rate: ${stable_rate}"

    # Compare rates
    if (( $(echo "$canary_rate > $THRESHOLD_ERROR_RATE" | bc -l) )); then
        error "Canary error rate exceeds threshold (${canary_rate} > ${THRESHOLD_ERROR_RATE})"
        return 1
    fi

    # Canary should not be significantly worse than stable
    local diff=$(echo "scale=4; $canary_rate - $stable_rate" | bc)
    if (( $(echo "$diff > 0.005" | bc -l) )); then
        warn "Canary error rate is significantly higher than stable (+${diff})"
        return 1
    fi

    log "✓ Error rate check passed"
    return 0
}

# Compare latencies
check_latency() {
    log "Checking latencies..."

    # Canary P95 latency
    local canary_p95=$(query_prometheus "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{deployment=\"${CANARY_DEPLOYMENT}\"}[5m]))")

    # Stable P95 latency
    local stable_p95=$(query_prometheus "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{deployment=\"${STABLE_DEPLOYMENT}\"}[5m]))")

    log "Canary P95 latency: ${canary_p95}s"
    log "Stable P95 latency: ${stable_p95}s"

    # Check threshold
    if (( $(echo "$canary_p95 > $THRESHOLD_LATENCY_P95" | bc -l) )); then
        error "Canary P95 latency exceeds threshold (${canary_p95}s > ${THRESHOLD_LATENCY_P95}s)"
        return 1
    fi

    # Canary should not be significantly slower than stable
    local diff=$(echo "scale=4; $canary_p95 - $stable_p95" | bc)
    if (( $(echo "$diff > 0.5" | bc -l) )); then
        warn "Canary is significantly slower than stable (+${diff}s)"
        return 1
    fi

    log "✓ Latency check passed"
    return 0
}

# Check success rate
check_success_rate() {
    log "Checking success rates..."

    # Canary success rate
    local canary_success=$(query_prometheus "sum(rate(http_requests_total{deployment=\"${CANARY_DEPLOYMENT}\",status=~\"2..\"}[5m])) / sum(rate(http_requests_total{deployment=\"${CANARY_DEPLOYMENT}\"}[5m]))")

    # Stable success rate
    local stable_success=$(query_prometheus "sum(rate(http_requests_total{deployment=\"${STABLE_DEPLOYMENT}\",status=~\"2..\"}[5m])) / sum(rate(http_requests_total{deployment=\"${STABLE_DEPLOYMENT}\"}[5m]))")

    log "Canary success rate: ${canary_success}"
    log "Stable success rate: ${stable_success}"

    # Success rate should be > 99%
    if (( $(echo "$canary_success < 0.99" | bc -l) )); then
        error "Canary success rate is too low (${canary_success} < 0.99)"
        return 1
    fi

    log "✓ Success rate check passed"
    return 0
}

# Check resource usage
check_resources() {
    log "Checking resource usage..."

    # CPU usage
    local canary_cpu=$(query_prometheus "avg(rate(container_cpu_usage_seconds_total{pod=~\"${CANARY_DEPLOYMENT}.*\"}[5m]))")
    local stable_cpu=$(query_prometheus "avg(rate(container_cpu_usage_seconds_total{pod=~\"${STABLE_DEPLOYMENT}.*\"}[5m]))")

    log "Canary CPU: ${canary_cpu}"
    log "Stable CPU: ${stable_cpu}"

    # Memory usage
    local canary_mem=$(query_prometheus "avg(container_memory_usage_bytes{pod=~\"${CANARY_DEPLOYMENT}.*\"})")
    local stable_mem=$(query_prometheus "avg(container_memory_usage_bytes{pod=~\"${STABLE_DEPLOYMENT}.*\"})")

    log "Canary Memory: ${canary_mem} bytes"
    log "Stable Memory: ${stable_mem} bytes"

    # Check if canary uses significantly more resources
    local cpu_ratio=$(echo "scale=2; $canary_cpu / $stable_cpu" | bc)
    if (( $(echo "$cpu_ratio > 1.5" | bc -l) )); then
        warn "Canary uses significantly more CPU than stable (${cpu_ratio}x)"
    fi

    log "✓ Resource check completed"
    return 0
}

# Main metrics check
main() {
    log "Starting canary metrics analysis..."

    local failed=0

    # Run checks
    if ! check_error_rate; then
        failed=$((failed + 1))
    fi

    if ! check_latency; then
        failed=$((failed + 1))
    fi

    if ! check_success_rate; then
        failed=$((failed + 1))
    fi

    check_resources

    # Summary
    echo ""
    if [ $failed -eq 0 ]; then
        log "========================================="
        log "Canary metrics validation passed! ✓"
        log "========================================="
        exit 0
    else
        error "========================================="
        error "Canary metrics validation failed! ✗"
        error "${failed} check(s) failed"
        error "========================================="
        exit 1
    fi
}

main "$@"
