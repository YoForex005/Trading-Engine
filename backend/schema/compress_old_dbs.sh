#!/bin/bash
# ============================================================================
# COMPRESS OLD TICK DATABASES
# ============================================================================
# Purpose: Compress SQLite databases older than 7 days using zstd
# Schedule: Run daily via cron or task scheduler
# Compression: 4-5x reduction with zstd level 19
# ============================================================================

set -euo pipefail

# Configuration
DB_DIR="${DB_DIR:-data/ticks/db}"
DAYS_BEFORE_COMPRESS="${DAYS_BEFORE_COMPRESS:-7}"
COMPRESSION_LEVEL="${COMPRESSION_LEVEL:-19}"
KEEP_ORIGINAL="${KEEP_ORIGINAL:-false}"
DRY_RUN="${DRY_RUN:-false}"
VERBOSE="${VERBOSE:-true}"

# Colors for output (defined early for security error messages)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ============================================================================
# SECURITY: Path Traversal Prevention
# ============================================================================
# Validate DB_DIR is within allowed base directory to prevent path traversal
# attacks like: DB_DIR="../../etc/passwd" ./compress_old_dbs.sh
# ============================================================================
ALLOWED_BASE="data/ticks"

# Canonicalize both paths to resolve symlinks and relative paths
REAL_DB_DIR=$(realpath "$DB_DIR" 2>/dev/null || echo "")
REAL_ALLOWED_BASE=$(realpath "$ALLOWED_BASE" 2>/dev/null || echo "")

# Validate that DB_DIR resolves successfully
if [[ -z "$REAL_DB_DIR" ]]; then
    echo -e "${RED}[ERROR]${NC} DB_DIR path is invalid or does not exist: $DB_DIR" >&2
    exit 1
fi

# Validate that DB_DIR is within ALLOWED_BASE (prevent directory traversal)
if [[ -z "$REAL_ALLOWED_BASE" ]]; then
    echo -e "${RED}[ERROR]${NC} Base directory does not exist: $ALLOWED_BASE" >&2
    exit 1
fi

if [[ "$REAL_DB_DIR" != "$REAL_ALLOWED_BASE"* ]]; then
    echo -e "${RED}[ERROR]${NC} Security violation: DB_DIR must be within $ALLOWED_BASE/" >&2
    echo -e "${RED}[ERROR]${NC} Attempted path: $DB_DIR" >&2
    echo -e "${RED}[ERROR]${NC} Resolved to: $REAL_DB_DIR" >&2
    exit 1
fi

# Validate compression level is numeric and within safe range (1-22 for zstd)
if ! [[ "$COMPRESSION_LEVEL" =~ ^[0-9]+$ ]] || [ "$COMPRESSION_LEVEL" -lt 1 ] || [ "$COMPRESSION_LEVEL" -gt 22 ]; then
    echo -e "${RED}[ERROR]${NC} Invalid COMPRESSION_LEVEL: $COMPRESSION_LEVEL (must be 1-22)" >&2
    exit 1
fi

# Logging functions
log() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${GREEN}[INFO]${NC} $1"
    fi
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check dependencies
check_dependencies() {
    if ! command -v zstd &> /dev/null; then
        error "zstd not found. Please install zstd compression tool."
        error "  Ubuntu/Debian: sudo apt-get install zstd"
        error "  macOS: brew install zstd"
        error "  Windows: Download from https://github.com/facebook/zstd/releases"
        exit 1
    fi

    if ! command -v sqlite3 &> /dev/null; then
        warn "sqlite3 not found. Database integrity checks will be skipped."
    fi
}

# Calculate cutoff date
get_cutoff_date() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        date -v-${DAYS_BEFORE_COMPRESS}d +%Y-%m-%d
    else
        # Linux
        date -d "${DAYS_BEFORE_COMPRESS} days ago" +%Y-%m-%d
    fi
}

# Compress a single database file
compress_database() {
    local db_file="$1"
    local compressed_file="${db_file}.zst"

    # ========================================================================
    # SECURITY: Filename Validation (Command Injection Prevention)
    # ========================================================================
    # Validate filename format to prevent command injection attacks like:
    # touch "file.db; rm -rf /"
    # Only allow alphanumeric, forward slash, underscore, hyphen, period
    # ========================================================================
    local filename=$(basename "$db_file")
    if ! [[ "$filename" =~ ^[a-zA-Z0-9_.-]+\.db$ ]]; then
        error "Security violation: Invalid filename format: $filename"
        error "Allowed pattern: [a-zA-Z0-9_.-]+.db"
        return 1
    fi

    # Check if already compressed
    if [ -f "$compressed_file" ]; then
        log "Already compressed: $compressed_file"
        return 0
    fi

    # Get file size before compression
    local size_before=$(du -h "$db_file" | cut -f1)

    log "Compressing: $db_file (size: $size_before)"

    # Verify database integrity before compression (if sqlite3 available)
    if command -v sqlite3 &> /dev/null; then
        # SECURITY: Use -- separator and quoted variables to prevent injection
        if ! sqlite3 -- "$db_file" "PRAGMA integrity_check;" > /dev/null 2>&1; then
            error "Database integrity check failed: $db_file"
            return 1
        fi
    fi

    if [ "$DRY_RUN" = true ]; then
        log "  [DRY RUN] Would compress: $db_file"
        return 0
    fi

    # ========================================================================
    # SECURITY: Properly quoted variables and -- separator
    # ========================================================================
    # Prevents command injection via malicious filenames
    # Using -- prevents filenames starting with - from being treated as flags
    # ========================================================================
    # Compress with zstd
    if zstd -"${COMPRESSION_LEVEL}" -q -- "$db_file" -o "$compressed_file"; then
        local size_after=$(du -h "$compressed_file" | cut -f1)
        local ratio=$(echo "scale=2; $(stat -f%z "$db_file") / $(stat -f%z "$compressed_file")" | bc 2>/dev/null || echo "N/A")

        log "  ✓ Compressed: $compressed_file (size: $size_after, ratio: ${ratio}x)"

        # Remove original if configured
        if [ "$KEEP_ORIGINAL" = false ]; then
            # SECURITY: Use -- separator and quoted variable
            rm -- "$db_file"
            log "  Removed original: $db_file"
        fi

        return 0
    else
        error "  ✗ Compression failed: $db_file"
        return 1
    fi
}

# Process all databases in directory
process_databases() {
    local cutoff_date=$(get_cutoff_date)
    local total=0
    local compressed=0
    local skipped=0
    local failed=0

    log "Compression threshold: $DAYS_BEFORE_COMPRESS days (cutoff: $cutoff_date)"
    log "Compression level: zstd-${COMPRESSION_LEVEL}"
    log "Keep original: $KEEP_ORIGINAL"
    log ""

    # Find all .db files older than cutoff date
    while IFS= read -r -d '' db_file; do
        total=$((total + 1))

        # Extract date from filename (ticks_YYYY-MM-DD.db)
        local filename=$(basename "$db_file")
        local db_date=$(echo "$filename" | grep -oP '\d{4}-\d{2}-\d{2}' || echo "")

        if [ -z "$db_date" ]; then
            warn "Could not extract date from: $filename"
            skipped=$((skipped + 1))
            continue
        fi

        # Compare dates
        if [[ "$db_date" < "$cutoff_date" ]]; then
            if compress_database "$db_file"; then
                compressed=$((compressed + 1))
            else
                failed=$((failed + 1))
            fi
        else
            log "Skipping recent file: $filename (date: $db_date)"
            skipped=$((skipped + 1))
        fi
    done < <(find "$DB_DIR" -type f -name "*.db" ! -name "*.db.zst" -print0)

    # Summary
    echo ""
    log "=== Compression Summary ==="
    log "Total databases found: $total"
    log "Compressed: $compressed"
    log "Skipped (recent): $skipped"
    log "Failed: $failed"

    if [ $failed -gt 0 ]; then
        error "⚠️  Compression completed with $failed errors"
        return 1
    else
        log "✅ Compression completed successfully"
        return 0
    fi
}

# Decompress a database (utility function)
decompress_database() {
    local compressed_file="$1"
    local db_file="${compressed_file%.zst}"

    # ========================================================================
    # SECURITY: Filename Validation (Command Injection Prevention)
    # ========================================================================
    # Validate compressed filename format
    # ========================================================================
    local filename=$(basename "$compressed_file")
    if ! [[ "$filename" =~ ^[a-zA-Z0-9_.-]+\.db\.zst$ ]]; then
        error "Security violation: Invalid compressed filename format: $filename"
        error "Allowed pattern: [a-zA-Z0-9_.-]+.db.zst"
        return 1
    fi

    # Validate decompressed filename format
    local db_filename=$(basename "$db_file")
    if ! [[ "$db_filename" =~ ^[a-zA-Z0-9_.-]+\.db$ ]]; then
        error "Security violation: Invalid database filename format: $db_filename"
        error "Allowed pattern: [a-zA-Z0-9_.-]+.db"
        return 1
    fi

    if [ ! -f "$compressed_file" ]; then
        error "Compressed file not found: $compressed_file"
        return 1
    fi

    if [ -f "$db_file" ]; then
        error "Database already exists: $db_file"
        return 1
    fi

    log "Decompressing: $compressed_file"

    # ========================================================================
    # SECURITY: Properly quoted variables and -- separator
    # ========================================================================
    if zstd -d -q -- "$compressed_file" -o "$db_file"; then
        log "  ✓ Decompressed: $db_file"

        # Verify integrity
        if command -v sqlite3 &> /dev/null; then
            # SECURITY: Use -- separator and quoted variables
            if sqlite3 -- "$db_file" "PRAGMA integrity_check;" > /dev/null 2>&1; then
                log "  ✓ Integrity check passed"
            else
                error "  ✗ Integrity check failed"
                return 1
            fi
        fi

        return 0
    else
        error "  ✗ Decompression failed: $compressed_file"
        return 1
    fi
}

# List compressed databases
list_compressed() {
    log "=== Compressed Databases ==="
    find "$DB_DIR" -type f -name "*.db.zst" -exec du -h {} \; | sort -k2
}

# Main entry point
main() {
    local action="${1:-compress}"

    case "$action" in
        compress)
            log "=== Tick Database Compression Tool ==="
            check_dependencies
            process_databases
            ;;
        decompress)
            if [ -z "${2:-}" ]; then
                error "Usage: $0 decompress <compressed_file>"
                exit 1
            fi
            check_dependencies
            decompress_database "$2"
            ;;
        list)
            list_compressed
            ;;
        help|--help|-h)
            cat << EOF
Tick Database Compression Tool

USAGE:
    $0 [ACTION] [OPTIONS]

ACTIONS:
    compress     Compress databases older than threshold (default)
    decompress   Decompress a specific database file
    list         List all compressed databases
    help         Show this help message

ENVIRONMENT VARIABLES:
    DB_DIR                  Directory containing databases (default: data/ticks/db)
    DAYS_BEFORE_COMPRESS    Days before compressing (default: 7)
    COMPRESSION_LEVEL       zstd compression level 1-22 (default: 19)
    KEEP_ORIGINAL          Keep original after compression (default: false)
    DRY_RUN                Dry run mode (default: false)
    VERBOSE                Verbose output (default: true)

EXAMPLES:
    # Compress databases older than 7 days
    $0 compress

    # Compress with custom threshold
    DAYS_BEFORE_COMPRESS=14 $0 compress

    # Dry run to see what would be compressed
    DRY_RUN=true $0 compress

    # Decompress a specific database
    $0 decompress data/ticks/db/2026/01/ticks_2026-01-10.db.zst

    # List all compressed databases
    $0 list

CRON SCHEDULE:
    # Run daily at 2 AM
    0 2 * * * /path/to/compress_old_dbs.sh compress >> /var/log/tick-compression.log 2>&1

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
