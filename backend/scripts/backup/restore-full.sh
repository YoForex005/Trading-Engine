#!/bin/bash
# Full system restore script
# RTO < 15 minutes

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
else
    echo "ERROR: Configuration file not found: $CONFIG_FILE"
    exit 1
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-trading_engine}"
DB_USER="${DB_USER:-postgres}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/trading-engine}"

BACKUP_FILE="${1:-}"
FORCE="${2:-no}"

if [[ -z "$BACKUP_FILE" ]]; then
    echo "Usage: $0 <backup-file> [force]"
    echo ""
    echo "Available backups:"
    find "$BACKUP_ROOT" -name "backup-*.tar.gz*" -type f | sort -r | head -10
    exit 1
fi

if [[ ! -f "$BACKUP_FILE" ]]; then
    echo "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

mkdir -p "$LOG_DIR"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_DIR/restore-full.log"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_DIR/restore-full.log" >&2
}

# Safety check
if [[ "$FORCE" != "force" ]]; then
    echo "=========================================="
    echo "⚠️  WARNING: DESTRUCTIVE OPERATION"
    echo "=========================================="
    echo "This will REPLACE the current database and application state with the backup."
    echo ""
    echo "Backup file: $BACKUP_FILE"
    echo "Target database: $DB_NAME on $DB_HOST:$DB_PORT"
    echo ""
    echo "Are you absolutely sure? Type 'yes' to continue:"
    read -r confirmation

    if [[ "$confirmation" != "yes" ]]; then
        echo "Restore cancelled."
        exit 0
    fi
fi

log "=========================================="
log "Starting full system restore: $RESTORE_ID"
log "Backup file: $BACKUP_FILE"
log "=========================================="

START_TIME=$(date +%s)
TEMP_DIR="/tmp/restore-${RESTORE_ID}"
mkdir -p "$TEMP_DIR"

# Cleanup function
cleanup() {
    local exit_code=$?
    log "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"

    if [[ $exit_code -ne 0 ]]; then
        error "Restore failed with exit code $exit_code"
    fi

    exit $exit_code
}
trap cleanup EXIT INT TERM

# 1. Stop application services
log "Stopping application services..."
if systemctl is-active --quiet trading-engine 2>/dev/null; then
    systemctl stop trading-engine
    log "Application service stopped"
fi

# 2. Decrypt backup
if [[ "$BACKUP_FILE" == *.gpg ]]; then
    log "Decrypting backup..."
    gpg --decrypt --output "$TEMP_DIR/backup.tar.gz" "$BACKUP_FILE"
    BACKUP_ARCHIVE="$TEMP_DIR/backup.tar.gz"
else
    BACKUP_ARCHIVE="$BACKUP_FILE"
fi

# 3. Extract backup
log "Extracting backup archive..."
tar -xzf "$BACKUP_ARCHIVE" -C "$TEMP_DIR"

# 4. Read metadata
METADATA_FILE=$(find "$TEMP_DIR" -name "backup-metadata.json" | head -1)
if [[ -f "$METADATA_FILE" ]]; then
    log "Backup metadata:"
    cat "$METADATA_FILE" | tee -a "$LOG_DIR/restore-full.log"
fi

# 5. Create database backup before restore
log "Creating safety backup of current database..."
SAFETY_BACKUP="/tmp/safety-backup-${RESTORE_ID}.dump"
PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --format=custom \
    --file="$SAFETY_BACKUP"
log "Safety backup created: $SAFETY_BACKUP"

# 6. Drop existing database connections
log "Terminating existing database connections..."
PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres <<EOF
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = '$DB_NAME'
  AND pid <> pg_backend_pid();
EOF

# 7. Drop and recreate database
log "Dropping and recreating database: $DB_NAME"
PGPASSWORD="${DB_PASSWORD}" dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" || true
PGPASSWORD="${DB_PASSWORD}" createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"

# 8. Restore PostgreSQL database
log "Restoring PostgreSQL database..."
POSTGRES_DUMP=$(find "$TEMP_DIR" -name "postgres-*.dump" | head -1)

if [[ ! -f "$POSTGRES_DUMP" ]]; then
    error "PostgreSQL dump file not found in backup"
    exit 1
fi

PGPASSWORD="${DB_PASSWORD}" pg_restore \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --verbose \
    --no-owner \
    --no-acl \
    "$POSTGRES_DUMP" 2>&1 | tee -a "$LOG_DIR/restore-full.log"

if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
    log "✓ PostgreSQL restore completed"
else
    error "PostgreSQL restore failed"
    log "Attempting to restore safety backup..."
    PGPASSWORD="${DB_PASSWORD}" pg_restore \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        "$SAFETY_BACKUP"
    exit 1
fi

# 9. Restore Redis data
log "Restoring Redis data..."
REDIS_DUMP=$(find "$TEMP_DIR" -name "redis-*.rdb" | head -1)

if [[ -f "$REDIS_DUMP" ]]; then
    # Stop Redis
    if systemctl is-active --quiet redis 2>/dev/null; then
        systemctl stop redis
    fi

    # Replace RDB file
    REDIS_RDB="${REDIS_RDB_PATH:-/var/lib/redis/dump.rdb}"
    cp "$REDIS_RDB" "$REDIS_RDB.backup-${RESTORE_ID}" || true
    cp "$REDIS_DUMP" "$REDIS_RDB"
    chown redis:redis "$REDIS_RDB" 2>/dev/null || true

    # Start Redis
    systemctl start redis 2>/dev/null || true
    sleep 2

    log "✓ Redis restore completed"
else
    log "No Redis backup found, skipping"
fi

# 10. Restore configuration files
log "Restoring configuration files..."
CONFIG_BACKUP=$(find "$TEMP_DIR" -type d -name "config" | head -1)

if [[ -d "$CONFIG_BACKUP" ]]; then
    # Backup current config
    if [[ -d "${APP_CONFIG_DIR:-/etc/trading-engine}" ]]; then
        cp -r "${APP_CONFIG_DIR:-/etc/trading-engine}" "${APP_CONFIG_DIR:-/etc/trading-engine}.backup-${RESTORE_ID}"
    fi

    # Restore config
    cp -r "$CONFIG_BACKUP"/* "${APP_CONFIG_DIR:-/etc/trading-engine}/" || true
    log "✓ Configuration restore completed"
fi

# 11. Restore application state
log "Restoring application state..."
STATE_BACKUP=$(find "$TEMP_DIR" -type d -name "state" | head -1)

if [[ -d "$STATE_BACKUP" ]]; then
    # Backup current state
    if [[ -d "${APP_DATA_DIR:-./data}" ]]; then
        mv "${APP_DATA_DIR:-./data}" "${APP_DATA_DIR:-./data}.backup-${RESTORE_ID}"
    fi

    # Restore state
    cp -r "$STATE_BACKUP"/* "${APP_DATA_DIR:-./data}/" || true
    log "✓ Application state restore completed"
fi

# 12. Verify restore
log "Verifying restore..."

# Check database
TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | xargs)
log "Tables restored: $TABLE_COUNT"

if [[ $TABLE_COUNT -eq 0 ]]; then
    error "No tables found after restore"
    exit 1
fi

# 13. Start application services
log "Starting application services..."
systemctl start trading-engine 2>/dev/null || true
sleep 5

if systemctl is-active --quiet trading-engine 2>/dev/null; then
    log "✓ Application service started"
else
    log "⚠ Application service failed to start, check logs"
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
DURATION_MIN=$((DURATION / 60))
DURATION_SEC=$((DURATION % 60))

# Final summary
log "=========================================="
log "Restore completed successfully!"
log "=========================================="
log "Backup file: $BACKUP_FILE"
log "Tables restored: $TABLE_COUNT"
log "Duration: ${DURATION_MIN}m ${DURATION_SEC}s"
log "Safety backup: $SAFETY_BACKUP (keep for 24h)"
log "=========================================="
log ""
log "⚠️  IMPORTANT: Verify application functionality before removing safety backup"
log "To remove safety backup: rm $SAFETY_BACKUP"

exit 0
