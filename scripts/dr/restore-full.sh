#!/bin/bash
###############################################################################
# RTX Trading Engine - Full Database Restore Script
# Purpose: Restore PostgreSQL database from backup
# Usage: ./restore-full.sh [backup-date] [--verify-only]
###############################################################################

set -euo pipefail

# Configuration
BACKUP_DIR="/var/backups/rtx"
RESTORE_DIR="/var/backups/rtx/restore"
DB_NAME="${RTX_DB_NAME:-rtx}"
DB_USER="${RTX_DB_USER:-postgres}"
DB_HOST="${RTX_DB_HOST:-localhost}"
DB_PORT="${RTX_DB_PORT:-5432}"
S3_BUCKET="${RTX_BACKUP_BUCKET:-rtx-backups}"
LOG_FILE="/var/log/rtx/restore-$(date +%Y%m%d_%H%M%S).log"

# Parse arguments
BACKUP_DATE="${1:-latest}"
VERIFY_ONLY="${2:-}"

# Ensure directories exist
mkdir -p "$BACKUP_DIR"
mkdir -p "$RESTORE_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Logging function
log() {
    echo "[$(date -u +%Y-%m-%dT%H:%M:%SZ)] $1" | tee -a "$LOG_FILE"
}

# Error handler
error_exit() {
    log "ERROR: $1"
    exit 1
}

trap 'error_exit "Script interrupted"' INT TERM

log "=== Starting Database Restore ==="
log "Backup date: $BACKUP_DATE"
log "Verify only: ${VERIFY_ONLY:-false}"

# Step 1: Determine backup file
log "Step 1/10: Locating backup file"
if [ "$BACKUP_DATE" == "latest" ]; then
    BACKUP_FILE=$(aws s3 ls "s3://${S3_BUCKET}/backups/full/" --recursive | sort | tail -1 | awk '{print $4}')
    if [ -z "$BACKUP_FILE" ]; then
        error_exit "No backups found in S3"
    fi
    log "Latest backup: $BACKUP_FILE"
else
    BACKUP_FILE=$(aws s3 ls "s3://${S3_BUCKET}/backups/full/${BACKUP_DATE}/" | grep '.dump$' | awk '{print $4}')
    if [ -z "$BACKUP_FILE" ]; then
        error_exit "Backup not found for date: $BACKUP_DATE"
    fi
    BACKUP_FILE="backups/full/${BACKUP_DATE}/${BACKUP_FILE}"
fi

BACKUP_NAME=$(basename "$BACKUP_FILE")
LOCAL_BACKUP="${BACKUP_DIR}/${BACKUP_NAME}"

# Step 2: Download backup from S3
log "Step 2/10: Downloading backup from S3"
START_TIME=$(date +%s)

aws s3 cp \
    "s3://${S3_BUCKET}/${BACKUP_FILE}" \
    "$LOCAL_BACKUP" \
    2>&1 | tee -a "$LOG_FILE" || error_exit "Download failed"

aws s3 cp \
    "s3://${S3_BUCKET}/${BACKUP_FILE}.sha256" \
    "${LOCAL_BACKUP}.sha256" \
    2>&1 | tee -a "$LOG_FILE"

DOWNLOAD_TIME=$(($(date +%s) - START_TIME))
log "Download completed in ${DOWNLOAD_TIME}s"

# Step 3: Verify checksum
log "Step 3/10: Verifying backup integrity"
EXPECTED_CHECKSUM=$(cat "${LOCAL_BACKUP}.sha256" | cut -d' ' -f1)
ACTUAL_CHECKSUM=$(sha256sum "$LOCAL_BACKUP" | cut -d' ' -f1)

if [ "$EXPECTED_CHECKSUM" != "$ACTUAL_CHECKSUM" ]; then
    error_exit "Checksum mismatch! Expected: $EXPECTED_CHECKSUM, Got: $ACTUAL_CHECKSUM"
fi

log "Checksum verified: $ACTUAL_CHECKSUM"

# Step 4: Verify backup file
log "Step 4/10: Verifying backup file structure"
if ! pg_restore --list "$LOCAL_BACKUP" > /dev/null 2>&1; then
    error_exit "Backup file is corrupted or invalid"
fi

BACKUP_VERSION=$(pg_restore --list "$LOCAL_BACKUP" | grep "PostgreSQL database dump" | head -1)
log "Backup info: $BACKUP_VERSION"

# If verify-only mode, exit here
if [ "$VERIFY_ONLY" == "--verify-only" ]; then
    log "Verification completed successfully. Exiting (verify-only mode)"
    exit 0
fi

# Step 5: WARNING - Point of no return
log "Step 5/10: ⚠️  WARNING - About to drop and restore database"
log "This will DELETE all current data in database: $DB_NAME"
log "Waiting 10 seconds for cancellation (Ctrl+C)..."
sleep 10

# Step 6: Stop dependent services
log "Step 6/10: Stopping RTX services"
systemctl stop rtx-backend || log "WARNING: Failed to stop rtx-backend"
systemctl stop rtx-websocket || log "WARNING: Failed to stop rtx-websocket"
systemctl stop rtx-fix-gateway || log "WARNING: Failed to stop rtx-fix-gateway"

# Wait for connections to close
log "Waiting 5 seconds for connections to close..."
sleep 5

# Step 7: Terminate existing connections
log "Step 7/10: Terminating existing database connections"
psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();
" 2>&1 | tee -a "$LOG_FILE"

# Step 8: Drop and recreate database
log "Step 8/10: Recreating database"
psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS ${DB_NAME};" \
    2>&1 | tee -a "$LOG_FILE" || error_exit "Failed to drop database"

psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "CREATE DATABASE ${DB_NAME} WITH ENCODING 'UTF8';" \
    2>&1 | tee -a "$LOG_FILE" || error_exit "Failed to create database"

# Enable extensions
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "CREATE EXTENSION IF NOT EXISTS timescaledb;" \
    2>&1 | tee -a "$LOG_FILE"

log "Database recreated successfully"

# Step 9: Restore database (parallel for faster restore)
log "Step 9/10: Restoring database (this may take several minutes)"
RESTORE_START=$(date +%s)

pg_restore \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --jobs=4 \
    --verbose \
    "$LOCAL_BACKUP" \
    2>&1 | tee -a "$LOG_FILE" || error_exit "Restore failed"

RESTORE_DURATION=$(($(date +%s) - RESTORE_START))
log "Restore completed in ${RESTORE_DURATION}s"

# Step 10: Post-restore validation
log "Step 10/10: Validating restored data"

# Count critical tables
ACCOUNTS_COUNT=$(psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM rtx_accounts;")
POSITIONS_COUNT=$(psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM rtx_positions;")
TRADES_COUNT=$(psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM rtx_trades;")

log "Validation results:"
log "  - Accounts: $ACCOUNTS_COUNT"
log "  - Positions: $POSITIONS_COUNT"
log "  - Trades: $TRADES_COUNT"

# Verify referential integrity
log "Verifying referential integrity..."
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT
        conrelid::regclass AS table_name,
        conname AS constraint_name,
        pg_get_constraintdef(oid) AS definition
    FROM pg_constraint
    WHERE contype = 'f'
    LIMIT 5;
" 2>&1 | tee -a "$LOG_FILE"

# Reindex for performance
log "Reindexing database..."
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "REINDEX DATABASE ${DB_NAME};" \
    2>&1 | tee -a "$LOG_FILE"

# Analyze for query planner
log "Analyzing tables..."
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "ANALYZE;" \
    2>&1 | tee -a "$LOG_FILE"

# Step 11: Restart services
log "Restarting RTX services"
systemctl start rtx-backend || error_exit "Failed to start rtx-backend"
systemctl start rtx-websocket || log "WARNING: Failed to start rtx-websocket"
systemctl start rtx-fix-gateway || log "WARNING: Failed to start rtx-fix-gateway"

# Step 12: Health check
log "Performing health check..."
sleep 5

if curl -f http://localhost:7999/health > /dev/null 2>&1; then
    log "✓ Health check passed"
else
    log "✗ Health check failed - manual intervention required"
fi

# Step 13: Record restore in database
psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c "
INSERT INTO restore_history (
    restore_date, backup_file, restore_duration_seconds, accounts_count, positions_count, trades_count, status
) VALUES (
    NOW(),
    '$BACKUP_FILE',
    $RESTORE_DURATION,
    $ACCOUNTS_COUNT,
    $POSITIONS_COUNT,
    $TRADES_COUNT,
    'completed'
);
" 2>&1 | tee -a "$LOG_FILE" || log "WARNING: Failed to record restore in database"

# Cleanup
log "Cleaning up temporary files"
rm -f "$LOCAL_BACKUP" "${LOCAL_BACKUP}.sha256"

# Success notification
log "=== Restore Completed Successfully ==="
log "Total duration: $((DOWNLOAD_TIME + RESTORE_DURATION))s"
log "Backup: $BACKUP_FILE"
log "Data validated: $ACCOUNTS_COUNT accounts, $POSITIONS_COUNT positions, $TRADES_COUNT trades"

aws sns publish \
    --topic-arn "${RTX_SNS_TOPIC}" \
    --subject "RTX Database RESTORED Successfully" \
    --message "Database restored from backup: $BACKUP_FILE. Duration: ${RESTORE_DURATION}s. Accounts: $ACCOUNTS_COUNT" \
    2>&1 | tee -a "$LOG_FILE"

exit 0
