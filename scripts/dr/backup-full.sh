#!/bin/bash
###############################################################################
# RTX Trading Engine - Full Database Backup Script
# Purpose: Create full PostgreSQL backup with encryption and upload to S3
# Schedule: Daily at 02:00 UTC via cron
# Retention: 7 days hot, 1 year cold storage
###############################################################################

set -euo pipefail

# Configuration
BACKUP_DIR="/var/backups/rtx"
DB_NAME="${RTX_DB_NAME:-rtx}"
DB_USER="${RTX_DB_USER:-postgres}"
DB_HOST="${RTX_DB_HOST:-localhost}"
DB_PORT="${RTX_DB_PORT:-5432}"
S3_BUCKET="${RTX_BACKUP_BUCKET:-rtx-backups}"
KMS_KEY_ID="${RTX_KMS_KEY:-arn:aws:kms:us-east-1:123456789:key/abc-123}"
RETENTION_DAYS=7
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="rtx-full-${DATE}.dump"
LOG_FILE="/var/log/rtx/backup-full-${DATE}.log"

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Logging function
log() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $1" | tee -a "$LOG_FILE"
}

# Error handler
error_exit() {
    log "ERROR: $1"
    aws sns publish \
        --topic-arn "${RTX_SNS_TOPIC}" \
        --subject "RTX Backup FAILED - Full Backup" \
        --message "Full backup failed: $1. Check $LOG_FILE for details." \
        2>&1 | tee -a "$LOG_FILE"
    exit 1
}

trap 'error_exit "Script interrupted"' INT TERM

log "=== Starting Full Backup ==="
log "Database: $DB_NAME"
log "Backup file: $BACKUP_FILE"

# Step 1: Pre-backup validation
log "Step 1/8: Pre-backup validation"
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -q; then
    error_exit "Database is not ready"
fi

# Check disk space (require 20GB free)
AVAILABLE_SPACE=$(df -BG "$BACKUP_DIR" | awk 'NR==2 {print $4}' | sed 's/G//')
if [ "$AVAILABLE_SPACE" -lt 20 ]; then
    error_exit "Insufficient disk space: ${AVAILABLE_SPACE}GB available, 20GB required"
fi

log "Disk space OK: ${AVAILABLE_SPACE}GB available"

# Step 2: Create backup using pg_dump (custom format for parallel restore)
log "Step 2/8: Creating database backup"
START_TIME=$(date +%s)

pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    -F custom \
    -Z 6 \
    -v \
    -f "${BACKUP_DIR}/${BACKUP_FILE}" \
    2>&1 | tee -a "$LOG_FILE" || error_exit "pg_dump failed"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
BACKUP_SIZE=$(du -h "${BACKUP_DIR}/${BACKUP_FILE}" | cut -f1)

log "Backup completed in ${DURATION}s, size: ${BACKUP_SIZE}"

# Step 3: Generate checksum
log "Step 3/8: Generating SHA256 checksum"
sha256sum "${BACKUP_DIR}/${BACKUP_FILE}" > "${BACKUP_DIR}/${BACKUP_FILE}.sha256"
CHECKSUM=$(cat "${BACKUP_DIR}/${BACKUP_FILE}.sha256" | cut -d' ' -f1)
log "Checksum: $CHECKSUM"

# Step 4: Create metadata file
log "Step 4/8: Creating metadata"
cat > "${BACKUP_DIR}/${BACKUP_FILE}.meta" <<EOF
{
  "backup_type": "full",
  "database": "$DB_NAME",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "size_bytes": $(stat -f%z "${BACKUP_DIR}/${BACKUP_FILE}" 2>/dev/null || stat -c%s "${BACKUP_DIR}/${BACKUP_FILE}"),
  "size_human": "$BACKUP_SIZE",
  "checksum": "$CHECKSUM",
  "duration_seconds": $DURATION,
  "postgres_version": "$(psql -h $DB_HOST -U $DB_USER -t -c 'SELECT version();' | head -1 | xargs)",
  "retention_days": $RETENTION_DAYS
}
EOF

# Step 5: Upload to S3 with encryption
log "Step 5/8: Uploading to S3"
aws s3 cp \
    "${BACKUP_DIR}/${BACKUP_FILE}" \
    "s3://${S3_BUCKET}/backups/full/${DATE}/${BACKUP_FILE}" \
    --sse aws:kms \
    --sse-kms-key-id "$KMS_KEY_ID" \
    --storage-class STANDARD \
    --metadata "backup-type=full,database=$DB_NAME,checksum=$CHECKSUM" \
    2>&1 | tee -a "$LOG_FILE" || error_exit "S3 upload failed"

# Upload checksum and metadata
aws s3 cp "${BACKUP_DIR}/${BACKUP_FILE}.sha256" "s3://${S3_BUCKET}/backups/full/${DATE}/"
aws s3 cp "${BACKUP_DIR}/${BACKUP_FILE}.meta" "s3://${S3_BUCKET}/backups/full/${DATE}/"

log "Upload completed successfully"

# Step 6: Verify upload
log "Step 6/8: Verifying S3 upload"
REMOTE_SIZE=$(aws s3 ls "s3://${S3_BUCKET}/backups/full/${DATE}/${BACKUP_FILE}" | awk '{print $3}')
LOCAL_SIZE=$(stat -f%z "${BACKUP_DIR}/${BACKUP_FILE}" 2>/dev/null || stat -c%s "${BACKUP_DIR}/${BACKUP_FILE}")

if [ "$REMOTE_SIZE" != "$LOCAL_SIZE" ]; then
    error_exit "Upload verification failed: size mismatch (local: $LOCAL_SIZE, remote: $REMOTE_SIZE)"
fi

log "Verification successful: sizes match ($LOCAL_SIZE bytes)"

# Step 7: Cleanup old local backups
log "Step 7/8: Cleaning up old local backups"
find "$BACKUP_DIR" -name "rtx-full-*.dump" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "rtx-full-*.sha256" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "rtx-full-*.meta" -mtime +$RETENTION_DAYS -delete
log "Local cleanup completed"

# Step 8: S3 lifecycle (move to Glacier after 30 days)
log "Step 8/8: Configuring S3 lifecycle"
aws s3api put-bucket-lifecycle-configuration \
    --bucket "$S3_BUCKET" \
    --lifecycle-configuration file://- <<EOF
{
  "Rules": [
    {
      "Id": "ArchiveFullBackups",
      "Status": "Enabled",
      "Prefix": "backups/full/",
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER"
        }
      ],
      "Expiration": {
        "Days": 365
      }
    }
  ]
}
EOF

# Step 9: Update latest symlink
log "Creating latest symlink"
ln -sf "${BACKUP_FILE}" "${BACKUP_DIR}/rtx-full-latest.dump"
ln -sf "${BACKUP_FILE}.sha256" "${BACKUP_DIR}/rtx-full-latest.dump.sha256"

# Step 10: Record backup in database
log "Recording backup in database"
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "
INSERT INTO backup_history (
    backup_type, backup_date, size_bytes, duration_seconds, s3_path, checksum, status
) VALUES (
    'full',
    NOW(),
    $LOCAL_SIZE,
    $DURATION,
    's3://${S3_BUCKET}/backups/full/${DATE}/${BACKUP_FILE}',
    '$CHECKSUM',
    'completed'
);
" 2>&1 | tee -a "$LOG_FILE" || log "WARNING: Failed to record backup in database"

# Success notification
log "=== Full Backup Completed Successfully ==="
log "Duration: ${DURATION}s"
log "Size: ${BACKUP_SIZE}"
log "Location: s3://${S3_BUCKET}/backups/full/${DATE}/"

aws sns publish \
    --topic-arn "${RTX_SNS_TOPIC}" \
    --subject "RTX Backup SUCCESS - Full Backup" \
    --message "Full backup completed successfully. Size: $BACKUP_SIZE, Duration: ${DURATION}s, Checksum: $CHECKSUM" \
    2>&1 | tee -a "$LOG_FILE"

# Calculate statistics for monitoring
TOTAL_BACKUPS=$(aws s3 ls "s3://${S3_BUCKET}/backups/full/" | wc -l)
log "Total backups in S3: $TOTAL_BACKUPS"

# Send metrics to CloudWatch
aws cloudwatch put-metric-data \
    --namespace RTX/Backups \
    --metric-name BackupDuration \
    --value "$DURATION" \
    --unit Seconds \
    --dimensions BackupType=Full

aws cloudwatch put-metric-data \
    --namespace RTX/Backups \
    --metric-name BackupSize \
    --value "$LOCAL_SIZE" \
    --unit Bytes \
    --dimensions BackupType=Full

exit 0
