#!/bin/bash
# Incremental backup script using WAL archiving and rsync
# Fast incremental backups with minimal storage overhead

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${BACKUP_CONFIG:-$SCRIPT_DIR/backup.config}"
LOG_DIR="${BACKUP_LOG_DIR:-/var/log/trading-engine/backup}"
BACKUP_ID="$(date +%Y%m%d-%H%M%S)"

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
ENABLE_ENCRYPTION="${ENABLE_ENCRYPTION:-true}"
ALERT_EMAIL="${ALERT_EMAIL:-ops@trading-engine.local}"

# Create directories
mkdir -p "$LOG_DIR" "$BACKUP_ROOT"
BACKUP_DIR="$BACKUP_ROOT/incremental-${BACKUP_ID}"
mkdir -p "$BACKUP_DIR"

# Logging functions
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_DIR/backup-incremental.log"
}

error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*" | tee -a "$LOG_DIR/backup-incremental.log" >&2
}

send_alert() {
    local subject="$1"
    local message="$2"
    if command -v mail >/dev/null 2>&1; then
        echo "$message" | mail -s "$subject" "$ALERT_EMAIL"
    fi
    logger -t trading-engine-backup -p user.alert "$subject: $message"
}

# Check for last full backup
LAST_FULL_BACKUP=""
if [[ -f "$BACKUP_ROOT/last-full-backup.txt" ]]; then
    LAST_FULL_BACKUP=$(cat "$BACKUP_ROOT/last-full-backup.txt")
    LAST_FULL_DIR="$BACKUP_ROOT/full-${LAST_FULL_BACKUP}"

    if [[ ! -d "$LAST_FULL_DIR" ]]; then
        error "Last full backup directory not found: $LAST_FULL_DIR"
        send_alert "Incremental Backup Failed" "No valid full backup found. Run full backup first."
        exit 1
    fi
else
    error "No full backup found. Run full backup first."
    send_alert "Incremental Backup Failed" "No full backup reference found."
    exit 1
fi

# Start backup
log "=========================================="
log "Starting incremental backup: $BACKUP_ID"
log "Base backup: $LAST_FULL_BACKUP"
log "=========================================="

START_TIME=$(date +%s)

# 1. Backup PostgreSQL WAL files since last backup
log "Backing up PostgreSQL WAL files..."
WAL_BACKUP_DIR="$BACKUP_DIR/pg_wal"
mkdir -p "$WAL_BACKUP_DIR"

# Get last backup timestamp
LAST_BACKUP_TS=0
if [[ -f "$BACKUP_ROOT/last-backup-timestamp.txt" ]]; then
    LAST_BACKUP_TS=$(cat "$BACKUP_ROOT/last-backup-timestamp.txt")
fi

# Copy WAL files modified since last backup
if [[ -d "${PG_WAL_DIR:-/var/lib/postgresql/data/pg_wal}" ]]; then
    find "${PG_WAL_DIR:-/var/lib/postgresql/data/pg_wal}" \
        -type f \
        -newermt "@$LAST_BACKUP_TS" \
        -exec cp {} "$WAL_BACKUP_DIR/" \;

    WAL_COUNT=$(find "$WAL_BACKUP_DIR" -type f | wc -l)
    log "Copied $WAL_COUNT WAL files: $(du -sh "$WAL_BACKUP_DIR" | cut -f1)"
fi

# 2. Incremental Redis backup (AOF diff)
log "Backing up Redis incremental data..."
REDIS_BACKUP_DIR="$BACKUP_DIR/redis"
mkdir -p "$REDIS_BACKUP_DIR"

if command -v redis-cli >/dev/null 2>&1; then
    # Copy AOF file if it exists
    REDIS_AOF="${REDIS_AOF_PATH:-/var/lib/redis/appendonly.aof}"
    if [[ -f "$REDIS_AOF" ]]; then
        # Only copy if modified since last backup
        if [[ $(stat -f %m "$REDIS_AOF" 2>/dev/null || stat -c %Y "$REDIS_AOF" 2>/dev/null) -gt $LAST_BACKUP_TS ]]; then
            cp "$REDIS_AOF" "$REDIS_BACKUP_DIR/appendonly-${BACKUP_ID}.aof"
            log "Redis AOF backup completed: $(du -h "$REDIS_BACKUP_DIR/appendonly-${BACKUP_ID}.aof" | cut -f1)"
        else
            log "Redis AOF unchanged, skipping"
        fi
    fi
fi

# 3. Incremental file backup using rsync
log "Backing up changed files..."
FILES_BACKUP_DIR="$BACKUP_DIR/files"
mkdir -p "$FILES_BACKUP_DIR"

# Backup changed application data
if [[ -d "${APP_DATA_DIR:-./data}" ]]; then
    rsync -av \
        --delete \
        --link-dest="$LAST_FULL_DIR" \
        "${APP_DATA_DIR:-./data}/" \
        "$FILES_BACKUP_DIR/data/" 2>&1 | tee -a "$LOG_DIR/backup-incremental.log"

    log "File backup completed: $(du -sh "$FILES_BACKUP_DIR" | cut -f1)"
fi

# 4. Backup changed configuration files
log "Backing up changed configuration..."
CONFIG_BACKUP_DIR="$BACKUP_DIR/config"
mkdir -p "$CONFIG_BACKUP_DIR"

if [[ -d "${APP_CONFIG_DIR:-/etc/trading-engine}" ]]; then
    find "${APP_CONFIG_DIR:-/etc/trading-engine}" \
        -type f \
        -newermt "@$LAST_BACKUP_TS" \
        -exec cp --parents {} "$CONFIG_BACKUP_DIR/" \;

    log "Configuration backup completed"
fi

# 5. Create metadata file
log "Creating backup metadata..."
cat > "$BACKUP_DIR/backup-metadata.json" <<EOF
{
  "backup_id": "$BACKUP_ID",
  "backup_type": "incremental",
  "base_backup_id": "$LAST_FULL_BACKUP",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "hostname": "$(hostname)",
  "components": [
    "postgresql_wal",
    "redis_aof",
    "changed_files",
    "changed_config"
  ],
  "wal_files_count": $WAL_COUNT,
  "encryption": $ENABLE_ENCRYPTION
}
EOF

# 6. Compress backup
log "Compressing incremental backup..."
tar -czf "$BACKUP_DIR.tar.gz" -C "$BACKUP_ROOT" "incremental-${BACKUP_ID}"

COMPRESSED_SIZE=$(du -h "$BACKUP_DIR.tar.gz" | cut -f1)
log "Compression completed: $COMPRESSED_SIZE"

# 7. Encrypt backup
if [[ "$ENABLE_ENCRYPTION" == "true" ]]; then
    log "Encrypting backup..."
    gpg --encrypt \
        --recipient "$GPG_RECIPIENT" \
        --output "$BACKUP_DIR.tar.gz.gpg" \
        "$BACKUP_DIR.tar.gz"

    rm "$BACKUP_DIR.tar.gz"
    FINAL_FILE="$BACKUP_DIR.tar.gz.gpg"
    log "Encryption completed"
else
    FINAL_FILE="$BACKUP_DIR.tar.gz"
fi

# 8. Upload to S3 if enabled
if [[ "${ENABLE_S3:-false}" == "true" ]] && [[ -n "${S3_BUCKET:-}" ]]; then
    log "Uploading to S3..."
    if command -v aws >/dev/null 2>&1; then
        aws s3 cp "$FINAL_FILE" "s3://$S3_BUCKET/backups/incremental/" \
            --metadata "backup-type=incremental,backup-id=$BACKUP_ID,base-backup=$LAST_FULL_BACKUP"
        log "S3 upload completed"
    fi
fi

# Calculate duration
END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

# Final summary
log "=========================================="
log "Incremental backup completed!"
log "=========================================="
log "Backup ID: $BACKUP_ID"
log "Base backup: $LAST_FULL_BACKUP"
log "Final size: $(du -h "$FINAL_FILE" | cut -f1)"
log "Duration: ${DURATION}s"
log "WAL files: $WAL_COUNT"
log "=========================================="

# Update last backup timestamp
date +%s > "$BACKUP_ROOT/last-backup-timestamp.txt"

send_alert "Incremental Backup Successful" "Incremental backup $BACKUP_ID completed in ${DURATION}s. WAL files: $WAL_COUNT"

exit 0
