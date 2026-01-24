#!/bin/bash
# ============================================================================
# LINUX/macOS CRON SETUP FOR DATABASE ROTATION & COMPRESSION
# ============================================================================
# Purpose: Automatically create cron jobs for daily DB rotation and
#          weekly compression following the 6-month retention policy
# Usage: ./setup_linux_cron.sh [install|uninstall|status]
# ============================================================================

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
ROTATE_SCRIPT="$SCRIPT_DIR/rotate_tick_db.sh"
COMPRESS_SCRIPT="$SCRIPT_DIR/compress_old_dbs.sh"
LOG_DIR="/var/log/trading-engine"
CRONTAB_BACKUP="/tmp/trading-engine-crontab.backup.$(date +%Y%m%d_%H%M%S)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Check if running as root
check_privileges() {
    if [[ $EUID -ne 0 ]]; then
        error "This script must be run as root (or with sudo)"
        error "Please run: sudo $0"
        exit 1
    fi
    log "Running as root"
}

# Check script paths
check_scripts() {
    if [ ! -f "$ROTATE_SCRIPT" ]; then
        error "Rotation script not found: $ROTATE_SCRIPT"
        exit 1
    fi
    success "Found rotation script: $ROTATE_SCRIPT"

    if [ ! -f "$COMPRESS_SCRIPT" ]; then
        error "Compression script not found: $COMPRESS_SCRIPT"
        exit 1
    fi
    success "Found compression script: $COMPRESS_SCRIPT"
}

# Make scripts executable
make_executable() {
    log "Making scripts executable..."

    if chmod +x "$ROTATE_SCRIPT"; then
        success "Made rotation script executable"
    else
        error "Failed to make rotation script executable"
        return 1
    fi

    if chmod +x "$COMPRESS_SCRIPT"; then
        success "Made compression script executable"
    else
        error "Failed to make compression script executable"
        return 1
    fi

    return 0
}

# Create log directory
create_log_dir() {
    log "Creating log directory: $LOG_DIR"

    if mkdir -p "$LOG_DIR"; then
        chmod 755 "$LOG_DIR"
        success "Created log directory"
    else
        warn "Could not create log directory (may not have permission)"
    fi
}

# Install cron jobs
install_cron() {
    log "Installing cron jobs..."

    # Backup existing crontab
    if crontab -l > "$CRONTAB_BACKUP" 2>/dev/null; then
        log "Backed up existing crontab to: $CRONTAB_BACKUP"
    fi

    # Create temporary crontab file
    local temp_crontab="/tmp/trading-engine-crontab.tmp"

    # Add header
    cat > "$temp_crontab" << 'EOF'
# ============================================================================
# Trading Engine Database Automation Cron Jobs
# ============================================================================
# Automatically rotate and compress tick databases for 6-month retention policy
# ============================================================================

EOF

    # Add existing crontab entries (excluding our entries)
    if crontab -l 2>/dev/null | grep -v "Trading Engine" | grep -v "rotate_tick_db.sh" | grep -v "compress_old_dbs.sh"; then
        crontab -l 2>/dev/null | grep -v "Trading Engine" | grep -v "rotate_tick_db.sh" | grep -v "compress_old_dbs.sh" >> "$temp_crontab"
    fi

    # Add rotation job - Daily at midnight UTC
    cat >> "$temp_crontab" << EOF

# Daily database rotation at midnight UTC
0 0 * * * $ROTATE_SCRIPT rotate >> $LOG_DIR/rotation.log 2>&1

EOF

    # Add status check job - Every hour
    cat >> "$temp_crontab" << EOF

# Hourly database status check
0 * * * * $ROTATE_SCRIPT status >> $LOG_DIR/status.log 2>&1

EOF

    # Add compression job - Weekly on Sunday at 2 AM UTC
    cat >> "$temp_crontab" << EOF

# Weekly compression on Sunday at 02:00 UTC
0 2 * * 0 $COMPRESS_SCRIPT compress >> $LOG_DIR/compression.log 2>&1

EOF

    # Install new crontab
    if crontab "$temp_crontab"; then
        success "Installed cron jobs"
        success "Cron schedule:"
        echo ""
        echo "  Daily rotation:        0 0 * * * (midnight UTC)"
        echo "  Hourly status check:   0 * * * * (every hour)"
        echo "  Weekly compression:    0 2 * * 0 (Sunday 2 AM UTC)"
        echo ""
        success "Log files: $LOG_DIR/rotation.log, status.log, compression.log"
    else
        error "Failed to install crontab"
        rm -f "$temp_crontab"
        return 1
    fi

    rm -f "$temp_crontab"
    return 0
}

# Uninstall cron jobs
uninstall_cron() {
    log "Removing cron jobs..."

    # Backup existing crontab
    if crontab -l > "$CRONTAB_BACKUP" 2>/dev/null; then
        log "Backed up crontab to: $CRONTAB_BACKUP"
    fi

    # Create temporary crontab file without our entries
    local temp_crontab="/tmp/trading-engine-crontab.tmp"

    # Copy all entries except ours
    if crontab -l 2>/dev/null | grep -v "rotate_tick_db.sh" | grep -v "compress_old_dbs.sh" > "$temp_crontab"; then
        crontab "$temp_crontab"
        success "Removed cron jobs"
    else
        warn "No existing crontab to modify"
    fi

    rm -f "$temp_crontab"
    return 0
}

# Show current cron status
show_status() {
    log "Current cron jobs for Trading Engine:"
    echo ""

    if crontab -l 2>/dev/null | grep -E "rotate_tick_db.sh|compress_old_dbs.sh"; then
        success "Cron jobs are installed"
    else
        warn "No cron jobs found"
    fi

    echo ""
    log "Full crontab:"
    crontab -l 2>/dev/null || warn "No crontab installed"

    echo ""
    if [ -d "$LOG_DIR" ]; then
        log "Recent log entries:"
        for log_file in "$LOG_DIR"/*.log; do
            if [ -f "$log_file" ]; then
                echo ""
                echo "=== $(basename "$log_file") ==="
                tail -n 10 "$log_file"
            fi
        done
    else
        warn "Log directory does not exist yet: $LOG_DIR"
    fi
}

# Test scripts (dry run)
test_scripts() {
    log "Testing scripts in dry-run mode..."
    echo ""

    log "Testing rotation script..."
    if DRY_RUN=true "$ROTATE_SCRIPT" rotate; then
        success "Rotation script test passed"
    else
        error "Rotation script test failed"
        return 1
    fi

    echo ""

    log "Testing compression script..."
    if DRY_RUN=true "$COMPRESS_SCRIPT" compress; then
        success "Compression script test passed"
    else
        error "Compression script test failed"
        return 1
    fi

    echo ""
    success "All script tests passed"
    return 0
}

# Main function
main() {
    local action="${1:-status}"

    echo ""
    echo "============================================================"
    echo "  Linux/macOS Cron Setup for Database Automation"
    echo "============================================================"
    echo ""

    case "$action" in
        install)
            check_privileges
            check_scripts
            make_executable
            create_log_dir
            test_scripts
            install_cron
            show_status
            ;;
        uninstall)
            check_privileges
            uninstall_cron
            log "Cron jobs removed"
            log "Backup available at: $CRONTAB_BACKUP"
            ;;
        status)
            show_status
            ;;
        test)
            check_scripts
            make_executable
            test_scripts
            ;;
        help|--help|-h)
            cat << EOF
Linux/macOS Cron Setup for Trading Engine Database Automation

USAGE:
    sudo $0 [ACTION]

ACTIONS:
    install    Install cron jobs (default)
    uninstall  Remove cron jobs
    status     Show current cron status
    test       Test scripts in dry-run mode
    help       Show this help message

SCHEDULE:
    - Daily rotation at midnight UTC (0 0 * * *)
    - Hourly status check (0 * * * *)
    - Weekly compression on Sunday 2 AM UTC (0 2 * * 0)

LOG FILES:
    - $LOG_DIR/rotation.log
    - $LOG_DIR/status.log
    - $LOG_DIR/compression.log

EXAMPLES:
    # Install cron jobs
    sudo $0 install

    # Check status
    sudo $0 status

    # Test scripts before installing
    sudo $0 test

    # Remove cron jobs
    sudo $0 uninstall

NOTES:
    - Root privileges required
    - Backup crontab is created before installation
    - All paths are absolute for cron compatibility
    - Logs are written to $LOG_DIR/

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
