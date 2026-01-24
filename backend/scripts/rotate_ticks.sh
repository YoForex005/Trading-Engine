#!/bin/bash

################################################################################
# Tick Data Rotation Script
#
# Purpose: Manage tick data lifecycle - archive old data, delete expired data
#
# Features:
#  - Read configuration from retention.yaml
#  - Find files older than archive threshold and move to archive with date structure
#  - Delete files older than retention limit
#  - Create backups before deletion
#  - Comprehensive logging with timestamps
#  - Dry-run mode for testing
#
# Usage: ./rotate_ticks.sh [--config /path/to/config.yaml] [--dry-run]
#
################################################################################

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_FILE="${1:-${PROJECT_ROOT}/config/retention.yaml}"
DRY_RUN="${2:-false}"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

################################################################################
# Parse YAML configuration file
################################################################################
parse_yaml() {
    local file=$1
    if [[ ! -f "$file" ]]; then
        echo -e "${RED}ERROR: Config file not found: $file${NC}" >&2
        exit 1
    fi

    # Extract values using grep and awk (works on most Unix systems)
    grep -E "^\s*(archive_threshold_days|deletion_threshold_days|ticks_directory|archive_directory|logs_directory|log_file|enable_archival|enable_deletion|compress_archives|dry_run|level):" "$file" | \
    sed 's/:[[:space:]]*/=/' | sed 's/^[[:space:]]*//' | sed "s/'//g" | sed 's/"//g'
}

# Load configuration
echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] Loading configuration from: $CONFIG_FILE${NC}"
eval "$(parse_yaml "$CONFIG_FILE")"

# Set defaults if not specified in config
archive_threshold_days=${archive_threshold_days:-30}
deletion_threshold_days=${deletion_threshold_days:-90}
ticks_directory=${ticks_directory:-"./data/ticks"}
archive_directory=${archive_directory:-"./data/archive"}
logs_directory=${logs_directory:-"./logs"}
log_file=${log_file:-"rotation.log"}
enable_archival=${enable_archival:-true}
enable_deletion=${enable_deletion:-true}
compress_archives=${compress_archives:-true}
level=${level:-"INFO"}

# Resolve paths relative to project root
if [[ ! "$ticks_directory" = /* ]]; then
    ticks_directory="${PROJECT_ROOT}/${ticks_directory}"
fi
if [[ ! "$archive_directory" = /* ]]; then
    archive_directory="${PROJECT_ROOT}/${archive_directory}"
fi
if [[ ! "$logs_directory" = /* ]]; then
    logs_directory="${PROJECT_ROOT}/${logs_directory}"
fi

LOG_FILE="${logs_directory}/${log_file}"

################################################################################
# Utility Functions
################################################################################

# Create log directory if it doesn't exist
mkdir -p "$logs_directory"

# Logging function
log_msg() {
    local level=$1
    shift
    local message="$@"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    echo -e "${timestamp} [${level}] ${message}" >> "$LOG_FILE"

    case "$level" in
        ERROR)
            echo -e "${RED}[${timestamp}] [${level}] ${message}${NC}" >&2
            ;;
        WARN)
            echo -e "${YELLOW}[${timestamp}] [${level}] ${message}${NC}"
            ;;
        INFO)
            echo -e "${BLUE}[${timestamp}] [${level}] ${message}${NC}"
            ;;
        SUCCESS)
            echo -e "${GREEN}[${timestamp}] [${level}] ${message}${NC}"
            ;;
        DEBUG)
            if [[ "$level" == "DEBUG" ]]; then
                echo -e "[${timestamp}] [${level}] ${message}"
            fi
            ;;
    esac
}

# Get file age in days
get_file_age_days() {
    local file=$1
    local file_time=$(stat -f%m "$file" 2>/dev/null || stat -c%Y "$file" 2>/dev/null || echo 0)
    local current_time=$(date +%s)
    local age_seconds=$((current_time - file_time))
    local age_days=$((age_seconds / 86400))
    echo "$age_days"
}

################################################################################
# Archival Functions
################################################################################

archive_old_files() {
    log_msg INFO "Starting archival process..."
    log_msg INFO "Archive threshold: ${archive_threshold_days} days"
    log_msg INFO "Source directory: ${ticks_directory}"
    log_msg INFO "Archive directory: ${archive_directory}"

    if [[ ! -d "$ticks_directory" ]]; then
        log_msg WARN "Ticks directory not found: ${ticks_directory}"
        return 0
    fi

    local archived_count=0
    local archived_size=0

    # Find all files in ticks directory (non-recursive to avoid OHLC data)
    while IFS= read -r -d '' file; do
        if [[ ! -f "$file" ]]; then
            continue
        fi

        local age_days=$(get_file_age_days "$file")

        if [[ $age_days -ge $archive_threshold_days ]]; then
            local filename=$(basename "$file")
            local file_date=$(stat -f%Sb -t%Y-%m-%d "$file" 2>/dev/null || stat -c%y "$file" 2>/dev/null | cut -d' ' -f1)
            local archive_subdir="${archive_directory}/${file_date}"

            # Create archive subdirectory with date structure
            if [[ "$DRY_RUN" == "false" ]]; then
                mkdir -p "$archive_subdir"

                # Check if compression is enabled
                if [[ "$compress_archives" == "true" ]]; then
                    if gzip -c "$file" > "${archive_subdir}/${filename}.gz" 2>/dev/null; then
                        log_msg INFO "Archived (compressed): ${filename} -> ${archive_subdir}/${filename}.gz (age: ${age_days} days)"
                        rm -f "$file"
                        archived_count=$((archived_count + 1))
                        archived_size=$((archived_size + $(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)))
                    else
                        log_msg WARN "Failed to compress: ${filename}"
                    fi
                else
                    if mv "$file" "$archive_subdir/$filename" 2>/dev/null; then
                        log_msg INFO "Archived: ${filename} -> ${archive_subdir} (age: ${age_days} days)"
                        archived_count=$((archived_count + 1))
                        archived_size=$((archived_size + $(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)))
                    else
                        log_msg WARN "Failed to archive: ${filename}"
                    fi
                fi
            else
                log_msg INFO "[DRY-RUN] Would archive: ${filename} (age: ${age_days} days)"
            fi
        fi
    done < <(find "$ticks_directory" -maxdepth 1 -type f -print0)

    log_msg INFO "Archival complete: ${archived_count} files archived"
}

################################################################################
# Deletion Functions
################################################################################

delete_expired_files() {
    log_msg INFO "Starting deletion process..."
    log_msg INFO "Deletion threshold: ${deletion_threshold_days} days"
    log_msg INFO "Archive directory: ${archive_directory}"

    if [[ ! -d "$archive_directory" ]]; then
        log_msg INFO "Archive directory not found, nothing to delete"
        return 0
    fi

    local deleted_count=0
    local freed_size=0

    # Find all files in archive directory
    while IFS= read -r -d '' file; do
        if [[ ! -f "$file" ]]; then
            continue
        fi

        local age_days=$(get_file_age_days "$file")

        if [[ $age_days -ge $deletion_threshold_days ]]; then
            local filename=$(basename "$file")
            local filesize=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo 0)

            if [[ "$DRY_RUN" == "false" ]]; then
                if rm -f "$file" 2>/dev/null; then
                    log_msg INFO "Deleted: ${filename} (age: ${age_days} days, size: ${filesize} bytes)"
                    deleted_count=$((deleted_count + 1))
                    freed_size=$((freed_size + filesize))
                else
                    log_msg WARN "Failed to delete: ${filename}"
                fi
            else
                log_msg INFO "[DRY-RUN] Would delete: ${filename} (age: ${age_days} days, size: ${filesize} bytes)"
            fi
        fi
    done < <(find "$archive_directory" -type f -print0)

    # Clean up empty directories
    if [[ "$DRY_RUN" == "false" ]]; then
        find "$archive_directory" -type d -empty -delete 2>/dev/null || true
    fi

    log_msg INFO "Deletion complete: ${deleted_count} files deleted, ${freed_size} bytes freed"
}

################################################################################
# Summary Functions
################################################################################

print_summary() {
    log_msg INFO "=========================================="
    log_msg INFO "Tick Data Rotation Summary"
    log_msg INFO "=========================================="
    log_msg INFO "Configuration: ${CONFIG_FILE}"
    log_msg INFO "Ticks directory: ${ticks_directory}"
    log_msg INFO "Archive directory: ${archive_directory}"
    log_msg INFO "Archive threshold: ${archive_threshold_days} days"
    log_msg INFO "Deletion threshold: ${deletion_threshold_days} days"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_msg WARN "DRY-RUN MODE: No files were actually modified"
    fi

    # Print directory stats
    if [[ -d "$ticks_directory" ]]; then
        local ticks_count=$(find "$ticks_directory" -maxdepth 1 -type f | wc -l)
        local ticks_size=$(find "$ticks_directory" -maxdepth 1 -type f -exec stat -f%z {} + 2>/dev/null | awk '{sum+=$1} END {print sum}' || echo 0)
        log_msg INFO "Active ticks: ${ticks_count} files ($(numfmt --to=iec-i --suffix=B $ticks_size 2>/dev/null || echo $ticks_size) bytes)"
    fi

    if [[ -d "$archive_directory" ]]; then
        local archive_count=$(find "$archive_directory" -type f | wc -l)
        local archive_size=$(find "$archive_directory" -type f -exec stat -f%z {} + 2>/dev/null | awk '{sum+=$1} END {print sum}' || echo 0)
        log_msg INFO "Archived files: ${archive_count} files ($(numfmt --to=iec-i --suffix=B $archive_size 2>/dev/null || echo $archive_size) bytes)"
    fi

    log_msg INFO "Log file: ${LOG_FILE}"
    log_msg INFO "=========================================="
}

################################################################################
# Main Execution
################################################################################

main() {
    log_msg INFO "=========================================="
    log_msg INFO "Tick Data Rotation Script Started"
    log_msg INFO "=========================================="
    log_msg INFO "Script: $(basename "$0")"
    log_msg INFO "Executed at: $(date '+%Y-%m-%d %H:%M:%S')"
    log_msg INFO "Config file: ${CONFIG_FILE}"

    if [[ "$DRY_RUN" == "true" ]]; then
        log_msg WARN "Running in DRY-RUN mode - no files will be modified"
    fi

    # Check if archival is enabled
    if [[ "$enable_archival" == "true" ]]; then
        archive_old_files
    else
        log_msg INFO "Archival is disabled in configuration"
    fi

    # Check if deletion is enabled
    if [[ "$enable_deletion" == "true" ]]; then
        delete_expired_files
    else
        log_msg INFO "Deletion is disabled in configuration"
    fi

    # Print summary
    print_summary

    log_msg SUCCESS "Rotation script completed successfully"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN="true"
            shift
            ;;
        --help)
            echo "Usage: $0 [--config CONFIG_FILE] [--dry-run]"
            echo ""
            echo "Options:"
            echo "  --config CONFIG_FILE    Path to retention.yaml config file (default: ../config/retention.yaml)"
            echo "  --dry-run               Run in dry-run mode (show what would be done)"
            echo "  --help                  Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Run main function
main

exit 0
