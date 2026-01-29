#!/bin/bash
# Cleanup old backups according to retention policy
# Retention: 7 daily, 4 weekly, 12 monthly

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
RETENTION_DAILY="${RETENTION_DAILY:-7}"
RETENTION_WEEKLY="${RETENTION_WEEKLY:-4}"
RETENTION_MONTHLY="${RETENTION_MONTHLY:-12}"

mkdir -p "$LOG_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_DIR/cleanup.log"
}

log "=========================================="
log "Starting backup cleanup"
log "Retention policy: ${RETENTION_DAILY}d / ${RETENTION_WEEKLY}w / ${RETENTION_MONTHLY}m"
log "=========================================="

DELETED_COUNT=0
DELETED_SIZE=0

# Function to get file size in bytes
get_size() {
    if [[ -f "$1" ]]; then
        stat -f %z "$1" 2>/dev/null || stat -c %s "$1" 2>/dev/null || echo 0
    else
        echo 0
    fi
}

# 1. Cleanup daily backups (keep last N days)
log "Cleaning up daily backups (keep last ${RETENTION_DAILY} days)..."

find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f -mtime "+${RETENTION_DAILY}" | while read -r file; do
    size=$(get_size "$file")
    log "Deleting daily backup: $(basename "$file") ($(numfmt --to=iec "$size" 2>/dev/null || echo "$size bytes"))"
    rm -f "$file"
    ((DELETED_COUNT++)) || true
    DELETED_SIZE=$((DELETED_SIZE + size))
done

# 2. Cleanup incremental backups (keep last N days)
log "Cleaning up incremental backups (keep last ${RETENTION_DAILY} days)..."

find "$BACKUP_ROOT" -name "incremental-*.tar.gz*" -type f -mtime "+${RETENTION_DAILY}" | while read -r file; do
    size=$(get_size "$file")
    log "Deleting incremental backup: $(basename "$file") ($(numfmt --to=iec "$size" 2>/dev/null || echo "$size bytes"))"
    rm -f "$file"
    ((DELETED_COUNT++)) || true
    DELETED_SIZE=$((DELETED_SIZE + size))
done

# 3. Keep weekly backups (first backup of each week)
log "Managing weekly backups (keep last ${RETENTION_WEEKLY} weeks)..."

# Get unique weeks
WEEKS=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f -mtime "-$((RETENTION_DAILY * 4))" -printf "%TY-%U\n" | sort -u | tail -n "$RETENTION_WEEKLY")

# Tag weekly backups
for week in $WEEKS; do
    first_backup=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f -newermt "$week-Mon" ! -newermt "$week-Sun" | sort | head -1)
    if [[ -n "$first_backup" ]] && [[ -f "$first_backup" ]]; then
        weekly_link="$BACKUP_ROOT/weekly-$week.tar.gz.gpg"
        ln -sf "$first_backup" "$weekly_link" 2>/dev/null || true
        log "Tagged weekly backup: $week -> $(basename "$first_backup")"
    fi
done

# Cleanup old weekly backups
find "$BACKUP_ROOT" -name "weekly-*.tar.gz*" -type l -mtime "+$((RETENTION_WEEKLY * 7))" | while read -r file; do
    log "Removing old weekly link: $(basename "$file")"
    rm -f "$file"
done

# 4. Keep monthly backups (first backup of each month)
log "Managing monthly backups (keep last ${RETENTION_MONTHLY} months)..."

# Get unique months
MONTHS=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f -mtime "-$((RETENTION_WEEKLY * 7))" -printf "%TY-%Tm\n" | sort -u | tail -n "$RETENTION_MONTHLY")

# Tag monthly backups
for month in $MONTHS; do
    first_backup=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f -newermt "$month-01" ! -newermt "$month-31" | sort | head -1)
    if [[ -n "$first_backup" ]] && [[ -f "$first_backup" ]]; then
        monthly_link="$BACKUP_ROOT/monthly-$month.tar.gz.gpg"
        ln -sf "$first_backup" "$monthly_link" 2>/dev/null || true
        log "Tagged monthly backup: $month -> $(basename "$first_backup")"
    fi
done

# Cleanup old monthly backups
find "$BACKUP_ROOT" -name "monthly-*.tar.gz*" -type l -mtime "+$((RETENTION_MONTHLY * 30))" | while read -r file; do
    log "Removing old monthly link: $(basename "$file")"
    rm -f "$file"
done

# 5. Cleanup orphaned directories
log "Cleaning up orphaned directories..."

find "$BACKUP_ROOT" -type d -empty -delete 2>/dev/null || true

# 6. Cleanup S3 old backups
if [[ "${ENABLE_S3:-false}" == "true" ]] && [[ -n "${S3_BUCKET:-}" ]] && command -v aws >/dev/null 2>&1; then
    log "Cleaning up S3 backups..."

    # Daily backups
    aws s3 ls "s3://$S3_BUCKET/backups/full/" | awk '{print $4}' | while read -r file; do
        backup_date=$(echo "$file" | grep -oP '\d{8}')
        if [[ -n "$backup_date" ]]; then
            days_old=$(( ($(date +%s) - $(date -d "$backup_date" +%s)) / 86400 ))
            if [[ $days_old -gt $RETENTION_DAILY ]]; then
                log "Deleting S3 backup: $file (${days_old} days old)"
                aws s3 rm "s3://$S3_BUCKET/backups/full/$file"
            fi
        fi
    done
fi

# Summary
log "=========================================="
log "Cleanup completed"
log "Files deleted: $DELETED_COUNT"
if [[ $DELETED_SIZE -gt 0 ]]; then
    log "Space freed: $(numfmt --to=iec "$DELETED_SIZE" 2>/dev/null || echo "$DELETED_SIZE bytes")"
fi

# Disk usage report
CURRENT_USAGE=$(du -sh "$BACKUP_ROOT" 2>/dev/null | cut -f1 || echo "unknown")
log "Current backup storage usage: $CURRENT_USAGE"

# Count current backups
FULL_COUNT=$(find "$BACKUP_ROOT" -name "full-*.tar.gz*" -type f | wc -l)
INC_COUNT=$(find "$BACKUP_ROOT" -name "incremental-*.tar.gz*" -type f | wc -l)
WEEKLY_COUNT=$(find "$BACKUP_ROOT" -name "weekly-*.tar.gz*" -type l | wc -l)
MONTHLY_COUNT=$(find "$BACKUP_ROOT" -name "monthly-*.tar.gz*" -type l | wc -l)

log "Backups retained:"
log "  - Full:        $FULL_COUNT"
log "  - Incremental: $INC_COUNT"
log "  - Weekly:      $WEEKLY_COUNT"
log "  - Monthly:     $MONTHLY_COUNT"
log "=========================================="

exit 0
