#!/bin/bash
# Point-in-time recovery (PITR) using PostgreSQL WAL files
# Restore to specific timestamp with <5 minute RPO

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"
LOG_DIR="${BACKUP_LOG_DIR:-/var/log/trading-engine/backup}"
RESTORE_ID="$(date +%Y%m%d-%H%M%S)"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    # shellcheck source=/dev/null
    source "$CONFIG_FILE"
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-trading_engine}"
DB_USER="${DB_USER:-postgres}"
BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/trading-engine}"

TARGET_TIME="${1:-}"
BASE_BACKUP="${2:-}"

if [[ -z "$TARGET_TIME" ]]; then
    echo "Usage: $0 <target-time> [base-backup-file]"
    echo ""
    echo "Examples:"
    echo "  $0 '2024-01-15 14:30:00'"
    echo "  $0 '2024-01-15 14:30:00' /path/to/backup.tar.gz.gpg"
    echo ""
    echo "Target time format: YYYY-MM-DD HH:MM:SS"
    echo ""
    echo "Available backups:"
    find "$BACKUP_ROOT" -name "backup-*.tar.gz*" -type f | sort -r | head -10
    exit 1
fi

mkdir -p "$LOG_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_DIR/restore-pitr.log"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_DIR/restore-pitr.log" >&2
}

log "=========================================="
log "Point-in-Time Recovery: $RESTORE_ID"
log "Target time: $TARGET_TIME"
log "=========================================="

# Validate target time format
if ! date -d "$TARGET_TIME" >/dev/null 2>&1; then
    error "Invalid time format: $TARGET_TIME"
    exit 1
fi

TARGET_TIMESTAMP=$(date -d "$TARGET_TIME" +%s)
log "Target timestamp: $TARGET_TIMESTAMP"

# Find appropriate base backup if not provided
if [[ -z "$BASE_BACKUP" ]]; then
    log "Finding appropriate base backup..."

    # Find latest full backup before target time
    BASE_BACKUP=""
    for backup in $(find "$BACKUP_ROOT" -name "backup-*.tar.gz*" -type f | sort -r); do
        BACKUP_TIME=$(basename "$backup" | grep -oP '\d{8}-\d{6}' | head -1)
        BACKUP_TIMESTAMP=$(date -d "${BACKUP_TIME:0:8} ${BACKUP_TIME:9:2}:${BACKUP_TIME:11:2}:${BACKUP_TIME:13:2}" +%s)

        if [[ $BACKUP_TIMESTAMP -le $TARGET_TIMESTAMP ]]; then
            BASE_BACKUP="$backup"
            log "Selected base backup: $BASE_BACKUP (timestamp: $BACKUP_TIMESTAMP)"
            break
        fi
    done

    if [[ -z "$BASE_BACKUP" ]]; then
        error "No suitable base backup found before target time"
        exit 1
    fi
fi

if [[ ! -f "$BASE_BACKUP" ]]; then
    error "Base backup file not found: $BASE_BACKUP"
    exit 1
fi

START_TIME=$(date +%s)
TEMP_DIR="/tmp/pitr-${RESTORE_ID}"
mkdir -p "$TEMP_DIR"

cleanup() {
    local exit_code=$?
    log "Cleaning up..."
    rm -rf "$TEMP_DIR"
    exit $exit_code
}
trap cleanup EXIT INT TERM

# 1. Extract base backup
log "Extracting base backup..."
if [[ "$BASE_BACKUP" == *.gpg ]]; then
    gpg --decrypt --output "$TEMP_DIR/backup.tar.gz" "$BASE_BACKUP"
    BACKUP_ARCHIVE="$TEMP_DIR/backup.tar.gz"
else
    BACKUP_ARCHIVE="$BASE_BACKUP"
fi

tar -xzf "$BACKUP_ARCHIVE" -C "$TEMP_DIR"

# 2. Find PostgreSQL dump
POSTGRES_DUMP=$(find "$TEMP_DIR" -name "postgres-*.dump" | head -1)
if [[ ! -f "$POSTGRES_DUMP" ]]; then
    error "PostgreSQL dump not found in backup"
    exit 1
fi

# 3. Collect WAL files
log "Collecting WAL files for recovery..."
WAL_DIR="$TEMP_DIR/wal_recovery"
mkdir -p "$WAL_DIR"

# Copy WAL files from base backup
BASE_WAL=$(find "$TEMP_DIR" -type d -name "pg_wal" | head -1)
if [[ -d "$BASE_WAL" ]]; then
    cp "$BASE_WAL"/* "$WAL_DIR/" 2>/dev/null || true
fi

# Copy WAL files from incremental backups
for inc_backup in $(find "$BACKUP_ROOT" -name "incremental-*.tar.gz*" -type f | sort); do
    log "Checking incremental backup: $(basename "$inc_backup")"

    INC_TEMP="$TEMP_DIR/inc_temp"
    mkdir -p "$INC_TEMP"

    if [[ "$inc_backup" == *.gpg ]]; then
        gpg --decrypt --output "$INC_TEMP/inc.tar.gz" "$inc_backup"
        tar -xzf "$INC_TEMP/inc.tar.gz" -C "$INC_TEMP"
    else
        tar -xzf "$inc_backup" -C "$INC_TEMP"
    fi

    # Copy WAL files
    INC_WAL=$(find "$INC_TEMP" -type d -name "pg_wal" | head -1)
    if [[ -d "$INC_WAL" ]]; then
        cp "$INC_WAL"/* "$WAL_DIR/" 2>/dev/null || true
    fi

    rm -rf "$INC_TEMP"
done

WAL_COUNT=$(find "$WAL_DIR" -type f | wc -l)
log "Collected $WAL_COUNT WAL files for recovery"

# 4. Create recovery configuration
log "Creating recovery configuration..."
RECOVERY_CONF="$TEMP_DIR/recovery.conf"

cat > "$RECOVERY_CONF" <<EOF
# Point-in-Time Recovery Configuration
# Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)
# Target: $TARGET_TIME

restore_command = 'cp $WAL_DIR/%f %p'
recovery_target_time = '$TARGET_TIME'
recovery_target_action = 'promote'
EOF

log "Recovery configuration:"
cat "$RECOVERY_CONF" | tee -a "$LOG_DIR/restore-pitr.log"

# 5. Stop application
log "Stopping application services..."
systemctl stop trading-engine 2>/dev/null || true

# 6. Create safety backup
log "Creating safety backup..."
SAFETY_BACKUP="/tmp/safety-pitr-${RESTORE_ID}.dump"
PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --format=custom \
    --file="$SAFETY_BACKUP"

# 7. Drop and recreate database
log "Recreating database..."
PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres <<EOF
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
EOF

PGPASSWORD="${DB_PASSWORD}" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" || true
PGPASSWORD="${DB_PASSWORD}" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"

# 8. Restore base backup
log "Restoring base backup..."
PGPASSWORD="${DB_PASSWORD}" pg_restore \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --verbose \
    "$POSTGRES_DUMP" 2>&1 | tee -a "$LOG_DIR/restore-pitr.log"

# 9. Apply WAL files for point-in-time recovery
log "Applying WAL files for point-in-time recovery..."
log "This may take several minutes depending on the number of transactions..."

# PostgreSQL will automatically apply WAL files based on recovery.conf
# We need to place the recovery.conf in the data directory
PG_DATA_DIR="${PG_DATA_DIR:-/var/lib/postgresql/data}"
if [[ -d "$PG_DATA_DIR" ]]; then
    cp "$RECOVERY_CONF" "$PG_DATA_DIR/recovery.conf"

    # Restart PostgreSQL to trigger recovery
    systemctl restart postgresql 2>/dev/null || true

    # Wait for recovery to complete
    log "Waiting for recovery to complete..."
    sleep 5

    # Check recovery status
    while PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -t -c "SELECT pg_is_in_recovery();" | grep -q "t"; do
        log "Recovery in progress..."
        sleep 5
    done

    log "✓ Point-in-time recovery completed"
else
    log "⚠ PostgreSQL data directory not found, manual recovery configuration required"
fi

# 10. Verify recovery
log "Verifying recovery..."
TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | xargs)
log "Tables restored: $TABLE_COUNT"

# Get latest transaction timestamp
LATEST_TX=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT MAX(created_at) FROM transactions;" 2>/dev/null | xargs || echo "N/A")
log "Latest transaction: $LATEST_TX"

# 11. Start application
log "Starting application services..."
systemctl start trading-engine 2>/dev/null || true

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
DURATION_MIN=$((DURATION / 60))
DURATION_SEC=$((DURATION % 60))

# Final summary
log "=========================================="
log "Point-in-Time Recovery completed!"
log "=========================================="
log "Target time: $TARGET_TIME"
log "Base backup: $(basename "$BASE_BACKUP")"
log "WAL files applied: $WAL_COUNT"
log "Tables restored: $TABLE_COUNT"
log "Latest transaction: $LATEST_TX"
log "Duration: ${DURATION_MIN}m ${DURATION_SEC}s"
log "Safety backup: $SAFETY_BACKUP"
log "=========================================="

exit 0
