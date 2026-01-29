#!/bin/bash
# Backup health monitoring and alerting

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"
LOG_DIR="${BACKUP_LOG_DIR:-/var/log/trading-engine/backup}"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    # shellcheck source=/dev/null
    source "$CONFIG_FILE"
fi

BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/trading-engine}"
ALERT_EMAIL="${ALERT_EMAIL:-ops@trading-engine.local}"

mkdir -p "$LOG_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" >&2
}

send_alert() {
    local severity="$1"
    local message="$2"

    # Send email
    if command -v mail >/dev/null 2>&1; then
        echo "$message" | mail -s "[$severity] Backup Health Alert" "$ALERT_EMAIL"
    fi

    # Send to syslog
    logger -t trading-engine-backup -p "user.$severity" "$message"

    # Send to Slack if configured
    if [[ -n "${ALERT_SLACK_WEBHOOK:-}" ]]; then
        curl -X POST "$ALERT_SLACK_WEBHOOK" \
            -H 'Content-Type: application/json' \
            -d "{\"text\": \"[$severity] Backup Health Alert: $message\"}" 2>/dev/null || true
    fi
}

HEALTH_ISSUES=()
WARNINGS=()

log "Running backup health checks..."

# 1. Check last full backup age
log "Checking last full backup..."
if [[ -f "$BACKUP_ROOT/last-full-backup-timestamp.txt" ]]; then
    LAST_FULL_TS=$(cat "$BACKUP_ROOT/last-full-backup-timestamp.txt")
    CURRENT_TS=$(date +%s)
    AGE_HOURS=$(( (CURRENT_TS - LAST_FULL_TS) / 3600 ))

    if [[ $AGE_HOURS -gt 48 ]]; then
        HEALTH_ISSUES+=("Last full backup is $AGE_HOURS hours old (expected < 24h)")
    elif [[ $AGE_HOURS -gt 30 ]]; then
        WARNINGS+=("Last full backup is $AGE_HOURS hours old")
    else
        log "✓ Last full backup: $AGE_HOURS hours ago"
    fi
else
    HEALTH_ISSUES+=("No full backup found")
fi

# 2. Check backup count
log "Checking backup count..."
FULL_COUNT=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f | wc -l)
INC_COUNT=$(find "$BACKUP_ROOT" -name "incremental-*.tar.gz*" -type f | wc -l)

if [[ $FULL_COUNT -eq 0 ]]; then
    HEALTH_ISSUES+=("No full backups found")
else
    log "✓ Full backups: $FULL_COUNT"
fi

if [[ $INC_COUNT -eq 0 ]]; then
    WARNINGS+=("No incremental backups found")
else
    log "✓ Incremental backups: $INC_COUNT"
fi

# 3. Check backup size
log "Checking backup sizes..."
LATEST_FULL=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f | sort -r | head -1)

if [[ -n "$LATEST_FULL" ]] && [[ -f "$LATEST_FULL" ]]; then
    SIZE=$(stat -f %z "$LATEST_FULL" 2>/dev/null || stat -c %s "$LATEST_FULL" 2>/dev/null)
    SIZE_MB=$((SIZE / 1024 / 1024))

    if [[ $SIZE_MB -lt 10 ]]; then
        HEALTH_ISSUES+=("Latest backup suspiciously small: ${SIZE_MB}MB")
    elif [[ $SIZE_MB -gt 10000 ]]; then
        WARNINGS+=("Latest backup very large: ${SIZE_MB}MB")
    else
        log "✓ Latest backup size: ${SIZE_MB}MB"
    fi
fi

# 4. Check disk space
log "Checking disk space..."
BACKUP_DISK=$(df "$BACKUP_ROOT" | tail -1)
DISK_USAGE=$(echo "$BACKUP_DISK" | awk '{print $5}' | tr -d '%')

if [[ $DISK_USAGE -gt 90 ]]; then
    HEALTH_ISSUES+=("Backup disk usage critical: ${DISK_USAGE}%")
elif [[ $DISK_USAGE -gt 80 ]]; then
    WARNINGS+=("Backup disk usage high: ${DISK_USAGE}%")
else
    log "✓ Disk usage: ${DISK_USAGE}%"
fi

# 5. Check last verification
log "Checking last verification..."
if [[ -f "$LOG_DIR/backup-verify.log" ]]; then
    LAST_VERIFY=$(stat -f %m "$LOG_DIR/backup-verify.log" 2>/dev/null || stat -c %Y "$LOG_DIR/backup-verify.log" 2>/dev/null)
    CURRENT_TS=$(date +%s)
    VERIFY_AGE_HOURS=$(( (CURRENT_TS - LAST_VERIFY) / 3600 ))

    if [[ $VERIFY_AGE_HOURS -gt 48 ]]; then
        HEALTH_ISSUES+=("Last verification $VERIFY_AGE_HOURS hours ago (expected daily)")
    else
        log "✓ Last verification: $VERIFY_AGE_HOURS hours ago"
    fi

    # Check for verification failures
    if grep -q "ERROR" "$LOG_DIR/backup-verify.log" | tail -100; then
        HEALTH_ISSUES+=("Recent verification errors detected")
    fi
else
    WARNINGS+=("No verification log found")
fi

# 6. Check S3 sync
if [[ "${ENABLE_S3:-false}" == "true" ]] && command -v aws >/dev/null 2>&1; then
    log "Checking S3 sync..."

    S3_COUNT=$(aws s3 ls "s3://${S3_BUCKET}/backups/full/" 2>/dev/null | wc -l || echo 0)

    if [[ $S3_COUNT -eq 0 ]]; then
        HEALTH_ISSUES+=("No backups found in S3")
    else
        log "✓ S3 backups: $S3_COUNT"
    fi
fi

# 7. Check GPG key
if [[ "${ENABLE_ENCRYPTION:-true}" == "true" ]]; then
    log "Checking GPG encryption key..."

    if gpg --list-keys "${GPG_RECIPIENT}" >/dev/null 2>&1; then
        log "✓ GPG key available"
    else
        HEALTH_ISSUES+=("GPG encryption key not found: ${GPG_RECIPIENT}")
    fi
fi

# 8. Check log file sizes
log "Checking log file sizes..."
LOG_SIZE=$(du -sm "$LOG_DIR" 2>/dev/null | cut -f1 || echo 0)

if [[ $LOG_SIZE -gt 1000 ]]; then
    WARNINGS+=("Backup log directory large: ${LOG_SIZE}MB")
else
    log "✓ Log directory size: ${LOG_SIZE}MB"
fi

# 9. Check for failed backups
log "Checking for recent failures..."
if [[ -f "$LOG_DIR/backup-full.log" ]]; then
    RECENT_FAILURES=$(grep -c "ERROR" "$LOG_DIR/backup-full.log" | tail -100 || echo 0)

    if [[ $RECENT_FAILURES -gt 5 ]]; then
        HEALTH_ISSUES+=("$RECENT_FAILURES recent backup errors detected")
    elif [[ $RECENT_FAILURES -gt 0 ]]; then
        WARNINGS+=("$RECENT_FAILURES recent backup errors")
    else
        log "✓ No recent failures"
    fi
fi

# Generate health report
log "=========================================="
log "Backup Health Report"
log "=========================================="

if [[ ${#HEALTH_ISSUES[@]} -eq 0 ]] && [[ ${#WARNINGS[@]} -eq 0 ]]; then
    log "✓ All health checks passed"
    log "Status: HEALTHY"

    # Write metrics
    if [[ "${ENABLE_METRICS:-true}" == "true" ]]; then
        echo "backup_health_status 1" > "$LOG_DIR/metrics.txt"
    fi

    exit 0
else
    if [[ ${#HEALTH_ISSUES[@]} -gt 0 ]]; then
        log "❌ CRITICAL ISSUES:"
        for issue in "${HEALTH_ISSUES[@]}"; do
            log "  - $issue"
        done

        # Send alert
        MESSAGE="Backup health check FAILED:\n\n"
        for issue in "${HEALTH_ISSUES[@]}"; do
            MESSAGE="${MESSAGE}- $issue\n"
        done
        send_alert "critical" "$MESSAGE"

        # Write metrics
        if [[ "${ENABLE_METRICS:-true}" == "true" ]]; then
            echo "backup_health_status 0" > "$LOG_DIR/metrics.txt"
            echo "backup_health_issues ${#HEALTH_ISSUES[@]}" >> "$LOG_DIR/metrics.txt"
        fi

        exit 1
    fi

    if [[ ${#WARNINGS[@]} -gt 0 ]]; then
        log "⚠️  WARNINGS:"
        for warning in "${WARNINGS[@]}"; do
            log "  - $warning"
        done

        # Send warning alert
        MESSAGE="Backup health check warnings:\n\n"
        for warning in "${WARNINGS[@]}"; do
            MESSAGE="${MESSAGE}- $warning\n"
        done
        send_alert "warning" "$MESSAGE"

        # Write metrics
        if [[ "${ENABLE_METRICS:-true}" == "true" ]]; then
            echo "backup_health_status 0.5" > "$LOG_DIR/metrics.txt"
            echo "backup_health_warnings ${#WARNINGS[@]}" >> "$LOG_DIR/metrics.txt"
        fi
    fi
fi

log "=========================================="
exit 0
