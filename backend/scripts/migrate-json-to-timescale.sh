#!/bin/bash
# migrate-json-to-timescale.sh
# Migrates JSON tick files to TimescaleDB
# Author: System Architect
# Date: 2026-01-20

set -e # Exit on error

# ============================================================================
# CONFIGURATION
# ============================================================================

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-rtx_db}"
DB_USER="${DB_USER:-rtx_app}"
DB_PASSWORD="${DB_PASSWORD:-your_password}"

TICK_DATA_DIR="${TICK_DATA_DIR:-./backend/data/ticks}"
TEMP_DIR="/tmp/rtx_migration"
BATCH_SIZE=10000
BROKER_ID="default"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ============================================================================
# FUNCTIONS
# ============================================================================

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        log_error "jq is not installed. Please install it: sudo apt-get install jq"
        exit 1
    fi

    # Check if psql is installed
    if ! command -v psql &> /dev/null; then
        log_error "psql is not installed. Please install PostgreSQL client."
        exit 1
    fi

    # Check if tick data directory exists
    if [ ! -d "$TICK_DATA_DIR" ]; then
        log_error "Tick data directory not found: $TICK_DATA_DIR"
        exit 1
    fi

    # Test database connection
    export PGPASSWORD="$DB_PASSWORD"
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
        log_error "Cannot connect to database: $DB_HOST:$DB_PORT/$DB_NAME"
        exit 1
    fi

    log_info "Prerequisites check passed"
}

create_temp_dir() {
    log_info "Creating temporary directory: $TEMP_DIR"
    mkdir -p "$TEMP_DIR"
}

cleanup_temp_dir() {
    log_info "Cleaning up temporary directory: $TEMP_DIR"
    rm -rf "$TEMP_DIR"
}

convert_json_to_csv() {
    local json_file="$1"
    local csv_file="$2"
    local symbol="$3"

    # Convert JSON to CSV using jq
    # Expected JSON format: [{"timestamp": "2026-01-20T15:30:45.123Z", "bid": 1.08456, "ask": 1.08458, ...}]
    jq -r --arg broker "$BROKER_ID" --arg sym "$symbol" \
        '.[] | [.timestamp, $broker, $sym, .bid, .ask, (.spread // 0), (.lp // "")] | @csv' \
        "$json_file" > "$csv_file"

    local line_count=$(wc -l < "$csv_file")
    echo "$line_count"
}

import_csv_to_database() {
    local csv_file="$1"
    local symbol="$2"

    export PGPASSWORD="$DB_PASSWORD"

    # Use COPY command for fast bulk insert
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
        "COPY tick_history (timestamp, broker_id, symbol, bid, ask, spread, lp) FROM STDIN WITH (FORMAT CSV);" \
        < "$csv_file"
}

migrate_symbol() {
    local symbol_dir="$1"
    local symbol=$(basename "$symbol_dir")

    # Validate symbol (only alphanumeric uppercase, prevent command injection)
    if ! [[ "$symbol" =~ ^[A-Z0-9]+$ ]]; then
        log_error "Invalid symbol directory '$symbol' (must be alphanumeric uppercase)"
        return 1
    fi

    log_info "Processing symbol: $symbol"

    local total_ticks=0
    local file_count=0

    # Process each JSON file in the symbol directory
    for json_file in "$symbol_dir"/*.json; do
        if [ ! -f "$json_file" ]; then
            continue
        fi

        local date=$(basename "$json_file" .json)
        log_info "  Importing date: $date"

        # Convert JSON to CSV
        local csv_file="$TEMP_DIR/${symbol}_${date}.csv"
        local tick_count=$(convert_json_to_csv "$json_file" "$csv_file" "$symbol")

        if [ "$tick_count" -eq 0 ]; then
            log_warn "  No ticks found in $json_file"
            continue
        fi

        # Import CSV to database
        if import_csv_to_database "$csv_file" "$symbol"; then
            log_info "  Imported $tick_count ticks from $date"
            total_ticks=$((total_ticks + tick_count))
            file_count=$((file_count + 1))
        else
            log_error "  Failed to import $csv_file"
        fi

        # Clean up CSV file
        rm -f "$csv_file"
    done

    log_info "Symbol $symbol complete: $total_ticks ticks from $file_count files"
}

verify_migration() {
    log_info "Verifying migration..."

    export PGPASSWORD="$DB_PASSWORD"

    # Get total tick count
    local tick_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT COUNT(*) FROM tick_history;")

    log_info "Total ticks in database: $(echo $tick_count | xargs)"

    # Get symbol count
    local symbol_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT COUNT(DISTINCT symbol) FROM tick_history;")

    log_info "Total symbols: $(echo $symbol_count | xargs)"

    # Get date range
    local date_range=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT MIN(timestamp), MAX(timestamp) FROM tick_history;")

    log_info "Date range: $date_range"

    # Get storage size
    local storage_size=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
        "SELECT pg_size_pretty(pg_total_relation_size('tick_history'));")

    log_info "Storage size: $(echo $storage_size | xargs)"
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

main() {
    log_info "Starting JSON to TimescaleDB migration..."
    log_info "Source directory: $TICK_DATA_DIR"
    log_info "Database: $DB_HOST:$DB_PORT/$DB_NAME"

    # Check prerequisites
    check_prerequisites

    # Create temporary directory
    create_temp_dir

    # Get start time
    start_time=$(date +%s)

    # Process each symbol directory
    symbol_count=0
    for symbol_dir in "$TICK_DATA_DIR"/*; do
        if [ ! -d "$symbol_dir" ]; then
            continue
        fi

        migrate_symbol "$symbol_dir"
        symbol_count=$((symbol_count + 1))
    done

    # Get end time
    end_time=$(date +%s)
    duration=$((end_time - start_time))

    log_info "Migration complete!"
    log_info "Processed $symbol_count symbols in $duration seconds"

    # Verify migration
    verify_migration

    # Cleanup
    cleanup_temp_dir

    log_info "Done! You can now use TimescaleDB for tick storage."
    log_info "Consider compressing old chunks: SELECT compress_chunk(chunk_name) FROM timescaledb_information.chunks WHERE hypertable_name = 'tick_history' AND range_end < NOW() - INTERVAL '7 days';"
}

# ============================================================================
# ENTRY POINT
# ============================================================================

# Trap errors
trap 'log_error "Migration failed! Check logs for details."; cleanup_temp_dir; exit 1' ERR

# Run main
main

exit 0
