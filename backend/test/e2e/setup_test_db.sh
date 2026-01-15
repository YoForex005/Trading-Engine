#!/bin/bash
# Setup test database for E2E tests

set -e

DB_NAME="trading_engine_test"
DB_USER="${POSTGRES_USER:-postgres}"
DB_HOST="${POSTGRES_HOST:-localhost}"
DB_PORT="${POSTGRES_PORT:-5432}"

echo "=========================================="
echo "Setting up test database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo "=========================================="

# Check if PostgreSQL is running
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
    echo "ERROR: PostgreSQL is not running on $DB_HOST:$DB_PORT"
    echo "Please start PostgreSQL before running E2E tests"
    exit 1
fi

# Drop existing test database if exists
echo "Dropping existing test database (if exists)..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true

# Create fresh test database
echo "Creating fresh test database..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"

# Find migrations directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATIONS_DIR="$SCRIPT_DIR/../../db/migrations"

if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "ERROR: Migrations directory not found at $MIGRATIONS_DIR"
    exit 1
fi

# Run migrations using golang-migrate if available
if command -v migrate &> /dev/null; then
    echo "Running migrations using golang-migrate..."
    migrate -database "postgres://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" \
            -path "$MIGRATIONS_DIR" up
else
    echo "WARNING: golang-migrate not found, attempting to run migrations manually..."

    # Run migrations manually in order
    for migration in "$MIGRATIONS_DIR"/*.up.sql; do
        if [ -f "$migration" ]; then
            echo "Running migration: $(basename $migration)"
            psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration"
        fi
    done
fi

echo "=========================================="
echo "Test database ready: $DB_NAME"
echo "Connection string: postgres://$DB_HOST:$DB_PORT/$DB_NAME"
echo "=========================================="
echo ""
echo "To run E2E tests:"
echo "  cd backend"
echo "  go test ./test/e2e -v"
echo ""
echo "To clean up test database:"
echo "  psql -U $DB_USER -c \"DROP DATABASE $DB_NAME;\""
echo "=========================================="
