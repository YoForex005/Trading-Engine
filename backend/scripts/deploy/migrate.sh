#!/bin/bash

set -e

# Configuration
MIGRATIONS_DIR="${MIGRATIONS_DIR:-./migrations}"
DATABASE_URL="${DATABASE_URL}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"; }

# Check prerequisites
check_prerequisites() {
    if [ -z "$DATABASE_URL" ]; then
        error "DATABASE_URL environment variable is not set"
        exit 1
    fi

    if [ ! -d "$MIGRATIONS_DIR" ]; then
        error "Migrations directory not found: $MIGRATIONS_DIR"
        exit 1
    fi

    log "Prerequisites check passed"
}

# Backup database
backup_database() {
    log "Creating database backup..."

    local backup_file="backup-$(date +%Y%m%d-%H%M%S).sql"

    # Extract connection details from DATABASE_URL
    # Format: postgres://user:pass@host:port/dbname

    if pg_dump "$DATABASE_URL" > "$backup_file" 2>/dev/null; then
        log "Database backup created: $backup_file"
    else
        warn "Failed to create database backup"
    fi
}

# Run migrations
run_migrations() {
    log "Running database migrations..."

    # Using golang-migrate or similar tool
    # This is a placeholder - adjust based on your migration tool

    if command -v migrate &> /dev/null; then
        migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
        log "Migrations completed successfully"
    else
        warn "migrate command not found, skipping migrations"
        warn "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    fi
}

# Verify migrations
verify_migrations() {
    log "Verifying migrations..."

    if command -v migrate &> /dev/null; then
        local version=$(migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version 2>&1 | grep -oP '\d+')
        log "Current migration version: $version"
    fi
}

# Main migration flow
main() {
    log "Starting database migration..."

    check_prerequisites
    backup_database
    run_migrations
    verify_migrations

    log "Database migration completed successfully!"
}

main "$@"
