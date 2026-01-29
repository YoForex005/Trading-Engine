#!/bin/bash
# Database restore script for trading engine
# Usage: ./restore.sh <backup_name>

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
    echo -e "${RED}Error: Backup name is required${NC}"
    echo "Usage: $0 <backup_name>"
    echo ""
    echo "Available backups:"
    ls -1t "$BACKUP_DIR" | grep ".sql.gz$" | sed 's/.sql.gz$//' | head -10
    exit 1
fi

BACKUP_FILE="${BACKUP_DIR}/${BACKUP_NAME}.sql.gz"
METADATA_FILE="${BACKUP_DIR}/${BACKUP_NAME}.metadata.json"

echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              Trading Engine - Database Restore                    ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}Error: Backup file not found: $BACKUP_FILE${NC}"
    exit 1
fi

# Show backup metadata if available
if [ -f "$METADATA_FILE" ]; then
    echo -e "${YELLOW}Backup metadata:${NC}"
    cat "$METADATA_FILE" | grep -v "^{" | grep -v "^}" | sed 's/^  /  /'
    echo ""
fi

echo -e "${YELLOW}Configuration:${NC}"
echo "  Database: $DB_NAME"
echo "  Host: $DB_HOST:$DB_PORT"
echo "  Backup file: $BACKUP_FILE"
echo ""

# Confirmation
echo -e "${RED}WARNING: This will replace all data in the '$DB_NAME' database!${NC}"
read -p "Are you sure you want to continue? (yes/N): " -r
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Restore cancelled"
    exit 0
fi

# Create a pre-restore backup
echo ""
echo -e "${YELLOW}→ Creating pre-restore backup...${NC}"
PRE_RESTORE_BACKUP="pre_restore_$(date +"%Y%m%d_%H%M%S")"
PGPASSWORD=$DB_PASSWORD pg_dump \
    -h $DB_HOST \
    -p $DB_PORT \
    -U $DB_USER \
    -d $DB_NAME \
    --format=plain \
    --no-owner \
    --no-acl \
    --file="${BACKUP_DIR}/${PRE_RESTORE_BACKUP}.sql" 2>/dev/null

if [ $? -eq 0 ]; then
    gzip -f "${BACKUP_DIR}/${PRE_RESTORE_BACKUP}.sql"
    echo -e "${GREEN}  ✓ Pre-restore backup created: ${PRE_RESTORE_BACKUP}.sql.gz${NC}"
else
    echo -e "${YELLOW}  ⚠ Could not create pre-restore backup (database might be empty)${NC}"
fi

# Drop existing connections
echo ""
echo -e "${YELLOW}→ Terminating active connections...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres <<EOF
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();
EOF

echo -e "${GREEN}  ✓ Connections terminated${NC}"

# Drop and recreate database
echo ""
echo -e "${YELLOW}→ Recreating database...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres <<EOF
DROP DATABASE IF EXISTS $DB_NAME;
CREATE DATABASE $DB_NAME;
EOF

echo -e "${GREEN}  ✓ Database recreated${NC}"

# Restore backup
echo ""
echo -e "${YELLOW}→ Restoring backup...${NC}"
gunzip -c "$BACKUP_FILE" | PGPASSWORD=$DB_PASSWORD psql \
    -h $DB_HOST \
    -p $DB_PORT \
    -U $DB_USER \
    -d $DB_NAME \
    --quiet 2>&1 | grep -v "^NOTICE:" | grep -v "^SET$" || true

if [ ${PIPESTATUS[1]} -eq 0 ]; then
    echo -e "${GREEN}  ✓ Backup restored successfully${NC}"
else
    echo -e "${RED}  ✗ Restore failed${NC}"
    echo ""
    echo -e "${YELLOW}Attempting to restore pre-restore backup...${NC}"
    gunzip -c "${BACKUP_DIR}/${PRE_RESTORE_BACKUP}.sql.gz" | PGPASSWORD=$DB_PASSWORD psql \
        -h $DB_HOST \
        -p $DB_PORT \
        -U $DB_USER \
        -d $DB_NAME \
        --quiet
    echo -e "${GREEN}  ✓ Rolled back to pre-restore state${NC}"
    exit 1
fi

# Verify restore
echo ""
echo -e "${YELLOW}→ Verifying restore...${NC}"
TABLE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo -e "${GREEN}  ✓ Tables restored: $TABLE_COUNT${NC}"

# Run migrations to ensure schema is up to date
echo ""
read -p "Run migrations to ensure schema is up to date? (Y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo -e "${YELLOW}→ Running migrations...${NC}"
    cd "$(dirname "$0")/../.."
    if [ -f "bin/migrate" ]; then
        ./bin/migrate -cmd up
    else
        echo -e "${YELLOW}  ⚠ Migration tool not found, skipping migrations${NC}"
    fi
fi

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                   Restore Completed Successfully!                 ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Restore details:${NC}"
echo "  Database: $DB_NAME"
echo "  Tables restored: $TABLE_COUNT"
echo "  Backup file: $BACKUP_FILE"
echo ""
echo -e "${YELLOW}Pre-restore backup:${NC}"
echo "  File: ${BACKUP_DIR}/${PRE_RESTORE_BACKUP}.sql.gz"
echo "  To rollback: ./restore.sh $PRE_RESTORE_BACKUP"
echo ""
