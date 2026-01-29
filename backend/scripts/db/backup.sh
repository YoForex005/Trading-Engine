#!/bin/bash
# Database backup script for trading engine
# Usage: ./backup.sh [backup_name]

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-trading_engine}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"

# Parse arguments
BACKUP_NAME="${1:-}"
if [ -z "$BACKUP_NAME" ]; then
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    BACKUP_NAME="backup_${TIMESTAMP}"
fi

# Create backup directory
mkdir -p "$BACKUP_DIR"

BACKUP_FILE="${BACKUP_DIR}/${BACKUP_NAME}.sql"
BACKUP_FILE_COMPRESSED="${BACKUP_FILE}.gz"

echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              Trading Engine - Database Backup                     ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${YELLOW}Configuration:${NC}"
echo "  Database: $DB_NAME"
echo "  Host: $DB_HOST:$DB_PORT"
echo "  Backup file: $BACKUP_FILE_COMPRESSED"
echo ""

# Check if pg_dump is available
if ! command -v pg_dump &> /dev/null; then
    echo -e "${RED}Error: pg_dump is not installed${NC}"
    exit 1
fi

# Perform backup
echo -e "${YELLOW}→ Creating database backup...${NC}"
PGPASSWORD=$DB_PASSWORD pg_dump \
    -h $DB_HOST \
    -p $DB_PORT \
    -U $DB_USER \
    -d $DB_NAME \
    --format=plain \
    --no-owner \
    --no-acl \
    --verbose \
    --file="$BACKUP_FILE" 2>&1 | grep -v "^pg_dump:"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}  ✓ Backup created successfully${NC}"
else
    echo -e "${RED}  ✗ Backup failed${NC}"
    exit 1
fi

# Compress backup
echo -e "${YELLOW}→ Compressing backup...${NC}"
gzip -f "$BACKUP_FILE"
echo -e "${GREEN}  ✓ Backup compressed${NC}"

# Get file size
FILE_SIZE=$(du -h "$BACKUP_FILE_COMPRESSED" | cut -f1)

# Create backup metadata
METADATA_FILE="${BACKUP_DIR}/${BACKUP_NAME}.metadata.json"
cat > "$METADATA_FILE" <<EOF
{
  "backup_name": "$BACKUP_NAME",
  "database": "$DB_NAME",
  "host": "$DB_HOST",
  "port": "$DB_PORT",
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "file_size": "$FILE_SIZE",
  "file_path": "$BACKUP_FILE_COMPRESSED",
  "format": "sql.gz",
  "pg_version": "$(pg_dump --version | head -n1)"
}
EOF

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Backup Completed Successfully!                 ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Backup details:${NC}"
echo "  File: $BACKUP_FILE_COMPRESSED"
echo "  Size: $FILE_SIZE"
echo "  Metadata: $METADATA_FILE"
echo ""

# List recent backups
echo -e "${YELLOW}Recent backups:${NC}"
ls -lht "$BACKUP_DIR" | grep ".sql.gz" | head -5

# Cleanup old backups (keep last 30 days)
echo ""
read -p "Clean up backups older than 30 days? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}→ Cleaning up old backups...${NC}"
    find "$BACKUP_DIR" -name "*.sql.gz" -type f -mtime +30 -delete
    find "$BACKUP_DIR" -name "*.metadata.json" -type f -mtime +30 -delete
    echo -e "${GREEN}  ✓ Old backups cleaned up${NC}"
fi

echo ""
echo -e "${YELLOW}To restore this backup, run:${NC}"
echo "  ./restore.sh $BACKUP_NAME"
echo ""
