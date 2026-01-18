#!/bin/bash
# Full system backup script with encryption and multi-destination support
# RTO < 15 minutes, RPO < 5 minutes

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"
LOG_DIR="${BACKUP_LOG_DIR:-/var/log/trading-engine/backup}"
BACKUP_ID="$(date +%Y%m%d-%H%M%S)"
TEMP_DIR="/tmp/backup-${BACKUP_ID}"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    # shellcheck source=/dev/null
    source "$CONFIG_FILE"
else
    echo "ERROR: Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Default configuration
BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/trading-engine}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-trading_engine}"
DB_USER="${DB_USER:-postgres}"
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
GPG_RECIPIENT="${GPG_RECIPIENT:-backup@trading-engine.local}"
S3_BUCKET="${S3_BUCKET:-}"
OFFSITE_HOST="${OFFSITE_HOST:-}"
ENABLE_ENCRYPTION="${ENABLE_ENCRYPTION:-true}"
ENABLE_S3="${ENABLE_S3:-false}"
ENABLE_OFFSITE="${ENABLE_OFFSITE:-false}"
ENABLE_VERIFICATION="${ENABLE_VERIFICATION:-true}"
ALERT_EMAIL="${ALERT_EMAIL:-ops@trading-engine.local}"

# Create directories
mkdir -p "$LOG_DIR" "$BACKUP_ROOT" "$TEMP_DIR"
BACKUP_DIR="$BACKUP_ROOT/full-${BACKUP_ID}"
mkdir -p "$BACKUP_DIR"

# Logging functions
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_DIR/backup-full.log"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_DIR/backup-full.log" >&2
}

send_alert() {
    local subject="$1"
    local message="$2"
    if command -v mail >/dev/null 2>&1; then
        echo "$message" | mail -s "$subject" "$ALERT_EMAIL"
    fi
    # Also log to syslog
    logger -t trading-engine-backup -p user.alert "$subject: $message"
}

# Cleanup on exit
cleanup() {
    local exit_code=$?
    log "Cleaning up temporary files..."
    rm -rf "$TEMP_DIR"

    if [[ $exit_code -ne 0 ]]; then
        error "Backup failed with exit code $exit_code"
        send_alert "Backup Failed" "Full backup $BACKUP_ID failed. Check logs at $LOG_DIR/backup-full.log"
    fi

    exit $exit_code
}
trap cleanup EXIT INT TERM

# Start backup
log "=========================================="
log "Starting full backup: $BACKUP_ID"
log "=========================================="

START_TIME=$(date +%s)

# 1. Backup PostgreSQL database
log "Backing up PostgreSQL database..."
PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "$DB_HOST" \
    -p "$DB_PORT" \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --format=custom \
    --compress=9 \
    --verbose \
    --file="$TEMP_DIR/postgres-${BACKUP_ID}.dump" 2>&1 | tee -a "$LOG_DIR/backup-full.log"

if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
    log "PostgreSQL backup completed: $(du -h "$TEMP_DIR/postgres-${BACKUP_ID}.dump" | cut -f1)"
else
    error "PostgreSQL backup failed"
    exit 1
fi

# 2. Backup PostgreSQL WAL files for point-in-time recovery
log "Backing up PostgreSQL WAL files..."
if [[ -d "${PG_WAL_DIR:-/var/lib/postgresql/data/pg_wal}" ]]; then
    mkdir -p "$TEMP_DIR/pg_wal"
    cp -r "${PG_WAL_DIR:-/var/lib/postgresql/data/pg_wal}"/* "$TEMP_DIR/pg_wal/" || true
    log "WAL files backed up: $(du -sh "$TEMP_DIR/pg_wal" | cut -f1)"
fi

# 3. Backup Redis data
log "Backing up Redis data..."
if command -v redis-cli >/dev/null 2>&1; then
    # Trigger Redis BGSAVE
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" BGSAVE

    # Wait for BGSAVE to complete
    sleep 2
    while redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LASTSAVE | grep -q "$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" LASTSAVE)"; do
        sleep 1
    done

    # Copy RDB file
    REDIS_RDB="${REDIS_RDB_PATH:-/var/lib/redis/dump.rdb}"
    if [[ -f "$REDIS_RDB" ]]; then
        cp "$REDIS_RDB" "$TEMP_DIR/redis-${BACKUP_ID}.rdb"
        log "Redis backup completed: $(du -h "$TEMP_DIR/redis-${BACKUP_ID}.rdb" | cut -f1)"
    fi
else
    log "Redis CLI not found, skipping Redis backup"
fi

# 4. Backup configuration files
log "Backing up configuration files..."
CONFIG_BACKUP="$TEMP_DIR/config"
mkdir -p "$CONFIG_BACKUP"

# Backup application config
if [[ -d "${APP_CONFIG_DIR:-/etc/trading-engine}" ]]; then
    cp -r "${APP_CONFIG_DIR:-/etc/trading-engine}" "$CONFIG_BACKUP/" || true
fi

# Backup environment files
if [[ -f "${APP_ROOT:-.}/.env" ]]; then
    cp "${APP_ROOT:-.}/.env" "$CONFIG_BACKUP/" || true
fi

# Backup nginx/proxy configs
if [[ -d "/etc/nginx/sites-enabled" ]]; then
    mkdir -p "$CONFIG_BACKUP/nginx"
    cp /etc/nginx/sites-enabled/* "$CONFIG_BACKUP/nginx/" || true
fi

log "Configuration backup completed: $(du -sh "$CONFIG_BACKUP" | cut -f1)"

# 5. Backup application state
log "Backing up application state..."
STATE_BACKUP="$TEMP_DIR/state"
mkdir -p "$STATE_BACKUP"

# Backup data directories
if [[ -d "${APP_DATA_DIR:-./data}" ]]; then
    cp -r "${APP_DATA_DIR:-./data}" "$STATE_BACKUP/" || true
    log "Application state backup completed: $(du -sh "$STATE_BACKUP" | cut -f1)"
fi

# 6. Backup log files
log "Backing up recent log files..."
LOG_BACKUP="$TEMP_DIR/logs"
mkdir -p "$LOG_BACKUP"

# Last 7 days of logs
if [[ -d "${APP_LOG_DIR:-/var/log/trading-engine}" ]]; then
    find "${APP_LOG_DIR:-/var/log/trading-engine}" -type f -mtime -7 -exec cp {} "$LOG_BACKUP/" \; || true
    log "Log files backup completed: $(du -sh "$LOG_BACKUP" | cut -f1)"
fi

# 7. Create metadata file
log "Creating backup metadata..."
cat > "$TEMP_DIR/backup-metadata.json" <<EOF
{
  "backup_id": "$BACKUP_ID",
  "backup_type": "full",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "hostname": "$(hostname)",
  "database": {
    "host": "$DB_HOST",
    "port": $DB_PORT,
    "name": "$DB_NAME",
    "version": "$(PGPASSWORD="${DB_PASSWORD}" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c 'SELECT version();' | head -1 | xargs)"
  },
  "redis": {
    "host": "$REDIS_HOST",
    "port": $REDIS_PORT,
    "version": "$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" INFO server | grep redis_version | cut -d: -f2 | tr -d '\r')"
  },
  "components": [
    "postgresql",
    "redis",
    "configuration",
    "application_state",
    "logs"
  ],
  "encryption": $ENABLE_ENCRYPTION,
  "compression": true
}
EOF

# 8. Compress backup
log "Compressing backup..."
tar -czf "$TEMP_DIR/backup-${BACKUP_ID}.tar.gz" -C "$TEMP_DIR" \
    --exclude="backup-${BACKUP_ID}.tar.gz" \
    --exclude="backup-${BACKUP_ID}.tar.gz.gpg" \
    .

COMPRESSED_SIZE=$(du -h "$TEMP_DIR/backup-${BACKUP_ID}.tar.gz" | cut -f1)
log "Compression completed: $COMPRESSED_SIZE"

# 9. Encrypt backup
if [[ "$ENABLE_ENCRYPTION" == "true" ]]; then
    log "Encrypting backup with GPG..."
    gpg --encrypt \
        --recipient "$GPG_RECIPIENT" \
        --output "$BACKUP_DIR/backup-${BACKUP_ID}.tar.gz.gpg" \
        "$TEMP_DIR/backup-${BACKUP_ID}.tar.gz"

    ENCRYPTED_SIZE=$(du -h "$BACKUP_DIR/backup-${BACKUP_ID}.tar.gz.gpg" | cut -f1)
    log "Encryption completed: $ENCRYPTED_SIZE"
    FINAL_FILE="$BACKUP_DIR/backup-${BACKUP_ID}.tar.gz.gpg"
else
    mv "$TEMP_DIR/backup-${BACKUP_ID}.tar.gz" "$BACKUP_DIR/"
    FINAL_FILE="$BACKUP_DIR/backup-${BACKUP_ID}.tar.gz"
fi

# Copy metadata
cp "$TEMP_DIR/backup-metadata.json" "$BACKUP_DIR/"

# 10. Upload to S3
if [[ "$ENABLE_S3" == "true" ]] && [[ -n "$S3_BUCKET" ]]; then
    log "Uploading to S3: s3://$S3_BUCKET/backups/full/"
    if command -v aws >/dev/null 2>&1; then
        aws s3 cp "$FINAL_FILE" "s3://$S3_BUCKET/backups/full/backup-${BACKUP_ID}.tar.gz.gpg" \
            --storage-class STANDARD_IA \
            --metadata "backup-type=full,backup-id=$BACKUP_ID,hostname=$(hostname)"

        aws s3 cp "$BACKUP_DIR/backup-metadata.json" "s3://$S3_BUCKET/backups/full/backup-${BACKUP_ID}-metadata.json"

        log "S3 upload completed"
    else
        error "AWS CLI not found, skipping S3 upload"
    fi
fi

# 11. Upload to offsite location
if [[ "$ENABLE_OFFSITE" == "true" ]] && [[ -n "$OFFSITE_HOST" ]]; then
    log "Uploading to offsite location: $OFFSITE_HOST"
    if command -v rsync >/dev/null 2>&1; then
        rsync -avz --progress \
            "$FINAL_FILE" \
            "$BACKUP_DIR/backup-metadata.json" \
            "$OFFSITE_HOST:/backups/trading-engine/full/" 2>&1 | tee -a "$LOG_DIR/backup-full.log"

        log "Offsite upload completed"
    else
        error "rsync not found, skipping offsite upload"
    fi
fi

# 12. Verify backup
if [[ "$ENABLE_VERIFICATION" == "true" ]]; then
    log "Verifying backup integrity..."
    "$SCRIPT_DIR/backup-verify.sh" "$FINAL_FILE" 2>&1 | tee -a "$LOG_DIR/backup-full.log"

    if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
        log "Backup verification passed"
    else
        error "Backup verification failed"
        exit 1
    fi
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
DURATION_MIN=$((DURATION / 60))
DURATION_SEC=$((DURATION % 60))

# Final summary
log "=========================================="
log "Backup completed successfully!"
log "=========================================="
log "Backup ID: $BACKUP_ID"
log "Location: $BACKUP_DIR"
log "Final size: $(du -h "$FINAL_FILE" | cut -f1)"
log "Duration: ${DURATION_MIN}m ${DURATION_SEC}s"
log "Components: PostgreSQL, Redis, Config, State, Logs"
log "Encryption: $ENABLE_ENCRYPTION"
log "S3 upload: $ENABLE_S3"
log "Offsite upload: $ENABLE_OFFSITE"
log "=========================================="

# Send success notification
send_alert "Backup Successful" "Full backup $BACKUP_ID completed in ${DURATION_MIN}m ${DURATION_SEC}s. Size: $(du -h "$FINAL_FILE" | cut -f1)"

# Update last backup timestamp
echo "$BACKUP_ID" > "$BACKUP_ROOT/last-full-backup.txt"
date +%s > "$BACKUP_ROOT/last-full-backup-timestamp.txt"

exit 0
