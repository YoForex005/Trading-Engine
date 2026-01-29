#!/bin/bash
###############################################################################
# RTX Trading Engine - Incremental Backup Script
# Purpose: Backup WAL files for point-in-time recovery
# Schedule: Every 6 hours via cron
###############################################################################

set -euo pipefail

# Configuration
WAL_ARCHIVE_DIR="/var/lib/postgresql/14/archive"
BACKUP_DIR="/var/backups/rtx/incremental"
S3_BUCKET="${RTX_BACKUP_BUCKET:-rtx-backups}"
DATE=$(date +%Y%m%d_%H%M%S)
LOG_FILE="/var/log/rtx/backup-incremental-${DATE}.log"

mkdir -p "$BACKUP_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

log() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $1" | tee -a "$LOG_FILE"
}

log "=== Starting Incremental Backup ==="

# Sync WAL files to S3
aws s3 sync \
    "$WAL_ARCHIVE_DIR" \
    "s3://${S3_BUCKET}/wal-archive/" \
    --sse aws:kms \
    --sse-kms-key-id "${RTX_KMS_KEY}" \
    --exclude "*" \
    --include "*.gz" \
    2>&1 | tee -a "$LOG_FILE"

# Cleanup old WAL files (keep 7 days)
find "$WAL_ARCHIVE_DIR" -name "*.gz" -mtime +7 -delete

log "=== Incremental Backup Completed ==="
