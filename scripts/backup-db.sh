#!/bin/bash
set -e  # Exit on error

# Configuration
BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/trading_engine_$TIMESTAMP.sql.gz"

# Create backup directory if not exists
mkdir -p "$BACKUP_DIR"

# Database connection from environment or docker-compose
DB_HOST="${DB_HOST:-db}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-trading_engine}"
DB_USER="${DB_USER:-postgres}"

# Run pg_dump with compression
echo "Starting backup at $TIMESTAMP..."
PGPASSWORD="$DB_PASSWORD" pg_dump \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  --format=custom \
  --compress=9 \
  --verbose \
  | gzip > "$BACKUP_FILE"

# Verify backup file created
if [ -f "$BACKUP_FILE" ]; then
  SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
  echo "Backup completed: $BACKUP_FILE ($SIZE)"
else
  echo "ERROR: Backup file not created"
  exit 1
fi

# Cleanup old backups (keep last 7 days)
find "$BACKUP_DIR" -name "trading_engine_*.sql.gz" -mtime +7 -delete

echo "Backup successful"
