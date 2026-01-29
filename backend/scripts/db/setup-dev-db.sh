#!/bin/bash
# Setup development database for trading engine
# Usage: ./setup-dev-db.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-trading_engine}"
DB_NAME_TEST="${DB_NAME_TEST:-trading_engine_test}"

echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        Trading Engine - Development Database Setup               ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: PostgreSQL client (psql) is not installed${NC}"
    echo "Install PostgreSQL:"
    echo "  macOS: brew install postgresql"
    echo "  Ubuntu: sudo apt-get install postgresql-client"
    exit 1
fi

echo -e "${YELLOW}Configuration:${NC}"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo "  Test Database: $DB_NAME_TEST"
echo ""

# Test PostgreSQL connection
echo -e "${YELLOW}→ Testing PostgreSQL connection...${NC}"
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c '\q' 2>/dev/null; then
    echo -e "${RED}Error: Cannot connect to PostgreSQL server${NC}"
    echo "Make sure PostgreSQL is running:"
    echo "  macOS: brew services start postgresql"
    echo "  Ubuntu: sudo systemctl start postgresql"
    exit 1
fi
echo -e "${GREEN}  ✓ PostgreSQL connection successful${NC}"

# Create databases
echo ""
echo -e "${YELLOW}→ Creating databases...${NC}"

# Drop databases if they exist (development only!)
read -p "Drop existing databases if they exist? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "  Dropping existing databases..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME_TEST;" 2>/dev/null || true
fi

# Create main database
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres <<EOF
SELECT 'CREATE DATABASE $DB_NAME'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec
EOF

echo -e "${GREEN}  ✓ Database '$DB_NAME' created${NC}"

# Create test database
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres <<EOF
SELECT 'CREATE DATABASE $DB_NAME_TEST'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME_TEST')\gexec
EOF

echo -e "${GREEN}  ✓ Database '$DB_NAME_TEST' created${NC}"

# Enable extensions
echo ""
echo -e "${YELLOW}→ Enabling PostgreSQL extensions...${NC}"

PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME <<EOF
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gist";
EOF

echo -e "${GREEN}  ✓ Extensions enabled${NC}"

# Run migrations
echo ""
echo -e "${YELLOW}→ Running database migrations...${NC}"

cd "$(dirname "$0")/../.."

# Build migration tool if not exists
if [ ! -f "bin/migrate" ]; then
    echo "  Building migration tool..."
    mkdir -p bin
    go build -o bin/migrate ./cmd/migrate
fi

# Run migrations
./bin/migrate -cmd init
./bin/migrate -cmd up

echo -e "${GREEN}  ✓ Migrations completed${NC}"

# Show database info
echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                    Setup Completed Successfully!                  ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Database connection details:${NC}"
echo "  Connection string: postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
echo ""
echo -e "${YELLOW}Environment variables (add to .env file):${NC}"
cat << EOF

export DB_HOST=$DB_HOST
export DB_PORT=$DB_PORT
export DB_USER=$DB_USER
export DB_PASSWORD=$DB_PASSWORD
export DB_NAME=$DB_NAME
export DB_SSLMODE=disable

EOF

echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Source environment variables: source .env"
echo "  2. Run seed data: ./scripts/db/seed-test-data.sh"
echo "  3. Start the server: go run cmd/server/main.go"
echo ""
