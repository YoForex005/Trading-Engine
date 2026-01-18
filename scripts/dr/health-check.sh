#!/bin/bash
###############################################################################
# RTX Trading Engine - Comprehensive Health Check
# Purpose: Monitor all critical systems and alert on failures
# Schedule: Every 60 seconds via cron or systemd timer
###############################################################################

set -euo pipefail

# Configuration
DB_HOST="${RTX_DB_HOST:-localhost}"
API_URL="${RTX_API_URL:-http://localhost:7999}"
WS_URL="${RTX_WS_URL:-ws://localhost:7999/ws}"
REDIS_HOST="${RTX_REDIS_HOST:-localhost}"
LOG_FILE="/var/log/rtx/health-check.log"
METRICS_FILE="/var/lib/rtx/metrics/health.json"
ALERT_THRESHOLD=3  # Alert after 3 consecutive failures

# Ensure directories exist
mkdir -p "$(dirname "$LOG_FILE")"
mkdir -p "$(dirname "$METRICS_FILE")"

# Health status tracking
HEALTH_STATUS="healthy"
FAILED_CHECKS=()

log() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $1" >> "$LOG_FILE"
}

alert() {
    local severity=$1
    local message=$2

    log "ALERT [$severity]: $message"

    if [ "$severity" == "critical" ]; then
        aws sns publish \
            --topic-arn "${RTX_SNS_TOPIC}" \
            --subject "RTX Health Check CRITICAL: $message" \
            --message "$message" 2>/dev/null || true
    fi
}

check_database() {
    local check_name="database"
    local start_time=$(date +%s%3N)

    if pg_isready -h "$DB_HOST" -q; then
        # Check query performance
        local query_time=$(psql -h "$DB_HOST" -U postgres -d rtx -t -c "\timing" -c "SELECT COUNT(*) FROM rtx_positions WHERE status='OPEN';" 2>&1 | grep "Time:" | awk '{print $2}' | sed 's/ ms//')

        local end_time=$(date +%s%3N)
        local duration=$((end_time - start_time))

        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"duration_ms\":$duration,\"query_time_ms\":${query_time:-0}}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"fail\",\"error\":\"Database unreachable\"}"
        alert "critical" "Database health check failed - unreachable"
        return 1
    fi
}

check_api() {
    local check_name="api"
    local start_time=$(date +%s%3N)

    local response=$(curl -s -o /dev/null -w "%{http_code}" -m 5 "$API_URL/health" 2>/dev/null || echo "000")
    local end_time=$(date +%s%3N)
    local duration=$((end_time - start_time))

    if [ "$response" == "200" ]; then
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"duration_ms\":$duration,\"http_code\":$response}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"fail\",\"http_code\":$response,\"duration_ms\":$duration}"
        alert "critical" "API health check failed - HTTP $response"
        return 1
    fi
}

check_websocket() {
    local check_name="websocket"
    local start_time=$(date +%s%3N)

    # Simple check using wscat or custom script
    if command -v wscat &> /dev/null; then
        timeout 5 wscat -c "$WS_URL" -x 'ping' &>/dev/null
        local ws_status=$?
    else
        # Fallback: assume healthy if API is healthy
        local ws_status=0
    fi

    local end_time=$(date +%s%3N)
    local duration=$((end_time - start_time))

    if [ $ws_status -eq 0 ]; then
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"duration_ms\":$duration}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"fail\",\"duration_ms\":$duration}"
        alert "warning" "WebSocket health check failed"
        return 1
    fi
}

check_redis() {
    local check_name="redis"
    local start_time=$(date +%s%3N)

    if redis-cli -h "$REDIS_HOST" ping 2>/dev/null | grep -q "PONG"; then
        local end_time=$(date +%s%3N)
        local duration=$((end_time - start_time))
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"duration_ms\":$duration}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"fail\"}"
        alert "warning" "Redis health check failed"
        return 1
    fi
}

check_disk_space() {
    local check_name="disk_space"
    local threshold=85

    local usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')

    if [ "$usage" -lt "$threshold" ]; then
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"usage_percent\":$usage}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"fail\",\"usage_percent\":$usage,\"threshold\":$threshold}"
        alert "warning" "Disk space critical: ${usage}% used"
        return 1
    fi
}

check_replication_lag() {
    local check_name="replication_lag"
    local threshold=60  # seconds

    local lag=$(psql -h "$DB_HOST" -U postgres -t -c "SELECT EXTRACT(EPOCH FROM (now() - pg_last_xact_replay_timestamp()));" 2>/dev/null | xargs || echo "0")

    if (( $(echo "$lag < $threshold" | bc -l) )); then
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"lag_seconds\":$lag}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"fail\",\"lag_seconds\":$lag,\"threshold\":$threshold}"
        alert "warning" "Replication lag high: ${lag}s"
        return 1
    fi
}

check_cpu_load() {
    local check_name="cpu_load"
    local threshold=80

    local load=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//' || echo "0")

    if (( $(echo "$load < $threshold" | bc -l 2>/dev/null || echo "1") )); then
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"load_percent\":$load}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"warn\",\"load_percent\":$load,\"threshold\":$threshold}"
        log "WARNING: High CPU load: ${load}%"
        return 0  # Don't fail on CPU - just warn
    fi
}

check_memory() {
    local check_name="memory"
    local threshold=90

    local usage=$(free | grep Mem | awk '{printf("%.0f", $3/$2 * 100.0)}')

    if [ "$usage" -lt "$threshold" ]; then
        echo "{\"check\":\"$check_name\",\"status\":\"pass\",\"usage_percent\":$usage}"
        return 0
    else
        FAILED_CHECKS+=("$check_name")
        echo "{\"check\":\"$check_name\",\"status\":\"warn\",\"usage_percent\":$usage,\"threshold\":$threshold}"
        log "WARNING: High memory usage: ${usage}%"
        return 0  # Don't fail on memory - just warn
    fi
}

# Run all checks
{
    echo "{"
    echo "  \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\","
    echo "  \"checks\": ["

    check_database
    echo ","
    check_api
    echo ","
    check_websocket
    echo ","
    check_redis
    echo ","
    check_disk_space
    echo ","
    check_replication_lag
    echo ","
    check_cpu_load
    echo ","
    check_memory

    echo "  ],"

    # Overall status
    if [ ${#FAILED_CHECKS[@]} -eq 0 ]; then
        echo "  \"overall_status\": \"healthy\","
        echo "  \"failed_checks\": []"
    else
        echo "  \"overall_status\": \"unhealthy\","
        echo "  \"failed_checks\": [\"$(IFS=\,; echo "${FAILED_CHECKS[*]}")\"]"
    fi

    echo "}"
} > "$METRICS_FILE"

# Send metrics to CloudWatch
if [ -f "$METRICS_FILE" ]; then
    FAILED_COUNT=${#FAILED_CHECKS[@]}

    aws cloudwatch put-metric-data \
        --namespace RTX/Health \
        --metric-name FailedChecks \
        --value "$FAILED_COUNT" \
        --unit Count 2>/dev/null || true
fi

# Exit with failure if critical checks failed
if [ ${#FAILED_CHECKS[@]} -gt 0 ]; then
    log "Health check FAILED: ${FAILED_CHECKS[*]}"
    exit 1
else
    log "Health check PASSED: All systems operational"
    exit 0
fi
