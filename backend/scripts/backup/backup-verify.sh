#!/bin/bash
# Backup verification script
# Verifies backup integrity by restoring to test database

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"
LOG_DIR="${BACKUP_LOG_DIR:-/var/log/trading-engine/backup}"
VERIFY_ID="$(date +%Y%m%d-%H%M%S)"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    # shellcheck source=/dev/null
    source "$CONFIG_FILE"
fi

TEST_DB_NAME="${TEST_DB_NAME:-trading_engine_verify_${VERIFY_ID}}"
TEST_DB_HOST="${DB_HOST:-localhost}"
TEST_DB_PORT="${DB_PORT:-5432}"
TEST_DB_USER="${DB_USER:-postgres}"
GPG_RECIPIENT="${GPG_RECIPIENT:-backup@trading-engine.local}"

BACKUP_FILE="${1:-}"

if [[ -z "$BACKUP_FILE" ]]; then
    echo "Usage: $0 <backup-file>"
    exit 1
fi

if [[ ! -f "$BACKUP_FILE" ]]; then
    echo "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

mkdir -p "$LOG_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_DIR/backup-verify.log"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_DIR/backup-verify.log" >&2
}

# Cleanup function
cleanup() {
    local exit_code=$?

    log "Cleaning up verification environment..."

    # Drop test database
    if [[ -n "$TEST_DB_NAME" ]]; then
        PGPASSWORD="${DB_PASSWORD}" psql -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USER" \
            -c "DROP DATABASE IF EXISTS $TEST_DB_NAME;" 2>/dev/null || true
    fi

    # Remove temp files
    rm -rf "$TEMP_DIR"

    if [[ $exit_code -eq 0 ]]; then
        log "Verification completed successfully"
    else
        error "Verification failed with exit code $exit_code"
    fi

    exit $exit_code
}
trap cleanup EXIT INT TERM

log "=========================================="
log "Starting backup verification: $VERIFY_ID"
log "Backup file: $BACKUP_FILE"
log "=========================================="

START_TIME=$(date +%s)
TEMP_DIR="/tmp/verify-${VERIFY_ID}"
mkdir -p "$TEMP_DIR"

# 1. Verify file integrity
log "Checking file integrity..."
if [[ ! -s "$BACKUP_FILE" ]]; then
    error "Backup file is empty"
    exit 1
fi

FILE_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
log "Backup file size: $FILE_SIZE"

# 2. Decrypt if encrypted
if [[ "$BACKUP_FILE" == *.gpg ]]; then
    log "Decrypting backup..."
    gpg --decrypt --output "$TEMP_DIR/backup.tar.gz" "$BACKUP_FILE"

    if [[ $? -ne 0 ]]; then
        error "Decryption failed"
        exit 1
    fi

    BACKUP_ARCHIVE="$TEMP_DIR/backup.tar.gz"
else
    BACKUP_ARCHIVE="$BACKUP_FILE"
fi

# 3. Extract archive
log "Extracting backup archive..."
tar -xzf "$BACKUP_ARCHIVE" -C "$TEMP_DIR"

if [[ $? -ne 0 ]]; then
    error "Archive extraction failed"
    exit 1
fi

# 4. Verify metadata
log "Verifying backup metadata..."
METADATA_FILE=$(find "$TEMP_DIR" -name "backup-metadata.json" | head -1)

if [[ ! -f "$METADATA_FILE" ]]; then
    error "Metadata file not found"
    exit 1
fi

log "Metadata contents:"
cat "$METADATA_FILE" | tee -a "$LOG_DIR/backup-verify.log"

# Extract backup type from metadata
BACKUP_TYPE=$(grep -o '"backup_type": *"[^"]*"' "$METADATA_FILE" | cut -d'"' -f4)
log "Backup type: $BACKUP_TYPE"

# 5. Verify PostgreSQL dump
log "Verifying PostgreSQL dump..."
POSTGRES_DUMP=$(find "$TEMP_DIR" -name "postgres-*.dump" | head -1)

if [[ ! -f "$POSTGRES_DUMP" ]]; then
    error "PostgreSQL dump file not found"
    exit 1
fi

# Check dump file integrity
pg_restore --list "$POSTGRES_DUMP" > /dev/null 2>&1

if [[ $? -ne 0 ]]; then
    error "PostgreSQL dump file is corrupted"
    exit 1
fi

log "PostgreSQL dump integrity check passed"

# 6. Test restore to temporary database
log "Creating test database: $TEST_DB_NAME"
PGPASSWORD="${DB_PASSWORD}" createdb \
    -h "$TEST_DB_HOST" \
    -p "$TEST_DB_PORT" \
    -U "$TEST_DB_USER" \
    "$TEST_DB_NAME"

if [[ $? -ne 0 ]]; then
    error "Failed to create test database"
    exit 1
fi

log "Restoring backup to test database..."
PGPASSWORD="${DB_PASSWORD}" pg_restore \
    -h "$TEST_DB_HOST" \
    -p "$TEST_DB_PORT" \
    -U "$TEST_DB_USER" \
    -d "$TEST_DB_NAME" \
    --verbose \
    "$POSTGRES_DUMP" 2>&1 | tee -a "$LOG_DIR/backup-verify.log"

if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
    error "Database restore failed"
    exit 1
fi

log "Database restore completed successfully"

# 7. Verify database content
log "Verifying database content..."

# Check table count
TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USER" -d "$TEST_DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | xargs)
log "Tables restored: $TABLE_COUNT"

if [[ $TABLE_COUNT -eq 0 ]]; then
    error "No tables found in restored database"
    exit 1
fi

# Check for critical tables
CRITICAL_TABLES=("users" "accounts" "orders" "positions" "transactions")
for table in "${CRITICAL_TABLES[@]}"; do
    EXISTS=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USER" -d "$TEST_DB_NAME" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" | xargs)

    if [[ "$EXISTS" != "t" ]]; then
        error "Critical table missing: $table"
        exit 1
    fi

    log "✓ Table verified: $table"
done

# 8. Verify Redis backup
log "Verifying Redis backup..."
REDIS_DUMP=$(find "$TEMP_DIR" -name "redis-*.rdb" | head -1)

if [[ -f "$REDIS_DUMP" ]]; then
    # Check RDB file format
    if head -c 5 "$REDIS_DUMP" | grep -q "REDIS"; then
        log "✓ Redis backup verified"
    else
        error "Redis backup file format invalid"
        exit 1
    fi
else
    log "No Redis backup found (may be expected)"
fi

# 9. Verify configuration files
log "Verifying configuration backups..."
CONFIG_BACKUP=$(find "$TEMP_DIR" -type d -name "config" | head -1)

if [[ -d "$CONFIG_BACKUP" ]]; then
    CONFIG_COUNT=$(find "$CONFIG_BACKUP" -type f | wc -l)
    log "Configuration files backed up: $CONFIG_COUNT"
else
    log "No configuration backup found"
fi

# 10. Calculate checksums
log "Calculating backup checksums..."
CHECKSUM_FILE="$TEMP_DIR/checksums.txt"
find "$TEMP_DIR" -type f -exec sha256sum {} \; > "$CHECKSUM_FILE"
log "Checksums calculated: $(wc -l < "$CHECKSUM_FILE") files"

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Final summary
log "=========================================="
log "Verification completed successfully!"
log "=========================================="
log "Backup file: $BACKUP_FILE"
log "Backup type: $BACKUP_TYPE"
log "File size: $FILE_SIZE"
log "Tables restored: $TABLE_COUNT"
log "Verification duration: ${DURATION}s"
log "Test database: $TEST_DB_NAME (will be dropped)"
log "=========================================="

exit 0
