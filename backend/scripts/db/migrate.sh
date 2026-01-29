#!/bin/bash

# RTX Trading Engine - Database Migration Script
# This script provides convenient wrappers around the migrate CLI tool

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$SCRIPT_DIR/../.."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Build the migrate tool if not exists
build_migrate() {
    echo -e "${YELLOW}Building migration tool...${NC}"
    cd "$PROJECT_ROOT"
    go build -o bin/migrate ./cmd/migrate
    echo -e "${GREEN}✅ Migration tool built${NC}"
}

# Check if migrate binary exists
if [ ! -f "$PROJECT_ROOT/bin/migrate" ]; then
    build_migrate
fi

MIGRATE="$PROJECT_ROOT/bin/migrate"

# Function to show usage
usage() {
    echo "RTX Trading Engine - Database Migration Helper"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  init      - Initialize migrations table"
    echo "  up        - Run all pending migrations"
    echo "  down      - Rollback last migration"
    echo "  status    - Show migration status"
    echo "  rebuild   - Rebuild migration tool"
    echo "  fresh     - Drop all tables and re-run migrations (DANGEROUS)"
    echo ""
    exit 1
}

# Handle commands
case "${1:-}" in
    init)
        echo -e "${YELLOW}Initializing migrations table...${NC}"
        $MIGRATE -init
        ;;
    up)
        echo -e "${YELLOW}Running all pending migrations...${NC}"
        $MIGRATE -up
        ;;
    down)
        echo -e "${YELLOW}Rolling back last migration...${NC}"
        $MIGRATE -down
        ;;
    status)
        $MIGRATE -status
        ;;
    rebuild)
        build_migrate
        ;;
    fresh)
        echo -e "${RED}⚠️  WARNING: This will drop all tables and re-run migrations!${NC}"
        read -p "Are you sure? Type 'YES' to confirm: " -r
        echo
        if [[ $REPLY == "YES" ]]; then
            echo -e "${YELLOW}Dropping all tables...${NC}"
            # Run down migrations until none left
            while $MIGRATE -down 2>/dev/null; do
                echo "Rolled back one migration"
            done
            echo -e "${YELLOW}Re-running all migrations...${NC}"
            $MIGRATE -up
            echo -e "${GREEN}✅ Database reset complete${NC}"
        else
            echo "Cancelled"
            exit 1
        fi
        ;;
    *)
        usage
        ;;
esac
