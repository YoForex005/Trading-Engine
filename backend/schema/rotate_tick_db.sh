#!/bin/bash
# ============================================================================
# ROTATE TICK DATABASE DAILY
# ============================================================================
# Purpose: Create new daily database at midnight and close previous day's DB
# Schedule: Run at 00:00 UTC via cron or task scheduler
# ============================================================================

set -euo pipefail

# Configuration
DB_DIR="${DB_DIR:-data/ticks/db}"
SCHEMA_FILE="${SCHEMA_FILE:-backend/schema/ticks.sql}"
BACKUP_DIR="${BACKUP_DIR:-data/ticks/backup}"
ENABLE_BACKUP="${ENABLE_BACKUP:-true}"
VERBOSE="${VERBOSE:-true}"
DRY_RUN="${DRY_RUN:-false}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Logging
log() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
    fi
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check dependencies
check_dependencies() {
    if ! command -v sqlite3 &> /dev/null; then
        error "sqlite3 not found. Please install SQLite3."
        exit 1
    fi

    if [ ! -f "$SCHEMA_FILE" ]; then
        error "Schema file not found: $SCHEMA_FILE"
        exit 1
    fi
}

# Get database path for a specific date
get_db_path() {
    local date="$1"
    local year=$(echo "$date" | cut -d'-' -f1)
    local month=$(echo "$date" | cut -d'-' -f2)
    echo "$DB_DIR/$year/$month/ticks_$date.db"
}

# Create directory structure
create_directories() {
    local date="$1"
    local year=$(echo "$date" | cut -d'-' -f1)
    local month=$(echo "$date" | cut -d'-' -f2)
    local dir="$DB_DIR/$year/$month"

    if [ "$DRY_RUN" = true ]; then
        log "[DRY RUN] Would create directory: $dir"
        return 0
    fi

    mkdir -p "$dir"
    log "Created directory: $dir"

    if [ "$ENABLE_BACKUP" = true ]; then
        mkdir -p "$BACKUP_DIR/$year/$month"
        log "Created backup directory: $BACKUP_DIR/$year/$month"
    fi
}

# Create new database with schema
create_database() {
    local db_path="$1"

    if [ -f "$db_path" ]; then
        warn "Database already exists: $db_path"
        return 0
    fi

    if [ "$DRY_RUN" = true ]; then
        log "[DRY RUN] Would create database: $db_path"
        return 0
    fi

    log "Creating database: $db_path"

    # Create database and apply schema
    if sqlite3 "$db_path" < "$SCHEMA_FILE"; then
        log "  ✓ Schema applied successfully"
    else
        error "  ✗ Failed to apply schema"
        return 1
    fi

    # Set performance pragmas
    sqlite3 "$db_path" << EOF
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = -64000;
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 268435456;
EOF

    if [ $? -eq 0 ]; then
        log "  ✓ Performance pragmas configured"
    else
        warn "  ⚠️ Failed to configure some pragmas"
    fi

    # Verify database
    if sqlite3 "$db_path" "PRAGMA integrity_check;" > /dev/null 2>&1; then
        log "  ✓ Database integrity verified"
    else
        error "  ✗ Database integrity check failed"
        return 1
    fi

    # Get file size
    local size=$(du -h "$db_path" | cut -f1)
    log "  Database created successfully (size: $size)"

    return 0
}

# Close previous database (checkpoint WAL)
close_database() {
    local db_path="$1"

    if [ ! -f "$db_path" ]; then
        warn "Database not found: $db_path"
        return 0
    fi

    if [ "$DRY_RUN" = true ]; then
        log "[DRY RUN] Would close database: $db_path"
        return 0
    fi

    log "Closing database: $db_path"

    # Checkpoint WAL to merge into main database
    if sqlite3 "$db_path" "PRAGMA wal_checkpoint(TRUNCATE);" > /dev/null 2>&1; then
        log "  ✓ WAL checkpoint completed"
    else
        warn "  ⚠️ WAL checkpoint failed (database may still be in use)"
    fi

    # Verify integrity
    if sqlite3 "$db_path" "PRAGMA integrity_check;" > /dev/null 2>&1; then
        log "  ✓ Final integrity check passed"
    else
        error "  ✗ Final integrity check failed"
        return 1
    fi

    # Get statistics
    local tick_count=$(sqlite3 "$db_path" "SELECT COUNT(*) FROM ticks;" 2>/dev/null || echo "unknown")
    local symbol_count=$(sqlite3 "$db_path" "SELECT COUNT(DISTINCT symbol) FROM ticks;" 2>/dev/null || echo "unknown")
    local size=$(du -h "$db_path" | cut -f1)

    log "  Database statistics:"
    log "    - Total ticks: $tick_count"
    log "    - Symbols: $symbol_count"
    log "    - Size: $size"

    return 0
}

# Backup database
backup_database() {
    local db_path="$1"
    local date="$2"

    if [ "$ENABLE_BACKUP" = false ]; then
        log "Backup disabled, skipping"
        return 0
    fi

    if [ ! -f "$db_path" ]; then
        warn "Database not found for backup: $db_path"
        return 0
    fi

    local year=$(echo "$date" | cut -d'-' -f1)
    local month=$(echo "$date" | cut -d'-' -f2)
    local backup_path="$BACKUP_DIR/$year/$month/ticks_$date.db"

    if [ "$DRY_RUN" = true ]; then
        log "[DRY RUN] Would backup to: $backup_path"
        return 0
    fi

    log "Backing up database to: $backup_path"

    # Use SQLite backup command for safe copy
    sqlite3 "$db_path" ".backup '$backup_path'" 2>/dev/null

    if [ $? -eq 0 ] && [ -f "$backup_path" ]; then
        local size=$(du -h "$backup_path" | cut -f1)
        log "  ✓ Backup completed (size: $size)"
        return 0
    else
        error "  ✗ Backup failed"
        return 1
    fi
}

# Update rotation metadata
update_metadata() {
    local today="$1"
    local yesterday="$2"

    local metadata_file="$DB_DIR/rotation_metadata.json"

    if [ "$DRY_RUN" = true ]; then
        log "[DRY RUN] Would update metadata: $metadata_file"
        return 0
    fi

    cat > "$metadata_file" << EOF
{
    "last_rotation": "$(date -u '+%Y-%m-%d %H:%M:%S UTC')",
    "current_date": "$today",
    "previous_date": "$yesterday",
    "current_db": "$(get_db_path "$today")",
    "previous_db": "$(get_db_path "$yesterday")"
}
EOF

    log "Metadata updated: $metadata_file"
}

# Main rotation process
perform_rotation() {
    local today=$(date '+%Y-%m-%d')
    local yesterday=$(date -d '1 day ago' '+%Y-%m-%d' 2>/dev/null || date -v-1d '+%Y-%m-%d')

    log "=== Daily Database Rotation ==="
    log "Today: $today"
    log "Yesterday: $yesterday"
    log ""

    # Step 1: Create directories
    log "Step 1: Creating directory structure"
    create_directories "$today"
    echo ""

    # Step 2: Close yesterday's database
    log "Step 2: Closing yesterday's database"
    local yesterday_db=$(get_db_path "$yesterday")
    close_database "$yesterday_db"
    echo ""

    # Step 3: Backup yesterday's database
    log "Step 3: Backing up yesterday's database"
    backup_database "$yesterday_db" "$yesterday"
    echo ""

    # Step 4: Create today's database
    log "Step 4: Creating today's database"
    local today_db=$(get_db_path "$today")
    if ! create_database "$today_db"; then
        error "Failed to create today's database"
        return 1
    fi
    echo ""

    # Step 5: Update metadata
    log "Step 5: Updating rotation metadata"
    update_metadata "$today" "$yesterday"
    echo ""

    # Step 6: Pre-create tomorrow's database (optional, for faster rotation)
    local tomorrow=$(date -d '1 day' '+%Y-%m-%d' 2>/dev/null || date -v+1d '+%Y-%m-%d')
    log "Step 6: Pre-creating tomorrow's database ($tomorrow)"
    create_directories "$tomorrow"
    local tomorrow_db=$(get_db_path "$tomorrow")
    create_database "$tomorrow_db"
    echo ""

    log "=== Rotation Complete ==="
    log "✅ Successfully rotated to: $today_db"

    return 0
}

# Status check
check_status() {
    local today=$(date '+%Y-%m-%d')
    local today_db=$(get_db_path "$today")

    log "=== Database Rotation Status ==="
    log "Current date: $today"
    log "Expected database: $today_db"

    if [ -f "$today_db" ]; then
        log "✅ Today's database exists"

        # Check if database is accessible
        if sqlite3 "$today_db" "SELECT 1;" > /dev/null 2>&1; then
            log "✅ Database is accessible"

            # Get statistics
            local tick_count=$(sqlite3 "$today_db" "SELECT COUNT(*) FROM ticks;" 2>/dev/null || echo "error")
            local symbol_count=$(sqlite3 "$today_db" "SELECT COUNT(DISTINCT symbol) FROM ticks;" 2>/dev/null || echo "error")

            log "Statistics:"
            log "  - Ticks: $tick_count"
            log "  - Symbols: $symbol_count"
        else
            error "✗ Database is not accessible"
            return 1
        fi
    else
        error "✗ Today's database does not exist"
        warn "Run rotation: $0 rotate"
        return 1
    fi

    # Check metadata
    local metadata_file="$DB_DIR/rotation_metadata.json"
    if [ -f "$metadata_file" ]; then
        log ""
        log "Last rotation metadata:"
        cat "$metadata_file"
    fi

    return 0
}

# Main entry point
main() {
    local action="${1:-rotate}"

    case "$action" in
        rotate)
            check_dependencies
            perform_rotation
            ;;
        status)
            check_dependencies
            check_status
            ;;
        help|--help|-h)
            cat << EOF
Tick Database Rotation Tool

USAGE:
    $0 [ACTION]

ACTIONS:
    rotate       Perform daily database rotation (default)
    status       Check rotation status
    help         Show this help message

ENVIRONMENT VARIABLES:
    DB_DIR          Database directory (default: data/ticks/db)
    SCHEMA_FILE     Schema SQL file (default: backend/schema/ticks.sql)
    BACKUP_DIR      Backup directory (default: data/ticks/backup)
    ENABLE_BACKUP   Enable backups (default: true)
    VERBOSE         Verbose output (default: true)
    DRY_RUN         Dry run mode (default: false)

EXAMPLES:
    # Perform rotation
    $0 rotate

    # Check status
    $0 status

    # Dry run (see what would happen)
    DRY_RUN=true $0 rotate

CRON SCHEDULE:
    # Run daily at midnight UTC
    0 0 * * * /path/to/rotate_tick_db.sh rotate >> /var/log/tick-rotation.log 2>&1

    # Run status check every hour
    0 * * * * /path/to/rotate_tick_db.sh status >> /var/log/tick-status.log 2>&1

EOF
            ;;
        *)
            error "Unknown action: $action"
            error "Run '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
