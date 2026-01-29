#!/bin/bash

# E2E Test Runner Script
# Runs Playwright E2E tests with various options

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== RTX Trading Engine E2E Tests ===${NC}\n"

# Parse command line arguments
MODE="${1:-all}"

# Check if backend is running
check_backend() {
    echo -e "${YELLOW}Checking backend server...${NC}"

    if curl -s http://localhost:7999/health > /dev/null; then
        echo -e "${GREEN}✓ Backend server is running${NC}\n"
        return 0
    else
        echo -e "${RED}✗ Backend server is not running${NC}"
        echo "Please start the backend server first:"
        echo "  cd ../../backend && go run cmd/server/main.go"
        echo ""
        exit 1
    fi
}

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}Installing dependencies...${NC}"
    npm install
    echo ""
fi

# Check if Playwright browsers are installed
if [ ! -d "node_modules/.cache/ms-playwright" ]; then
    echo -e "${YELLOW}Installing Playwright browsers...${NC}"
    npx playwright install
    echo ""
fi

# Check backend before running tests
check_backend

case $MODE in
    all)
        echo "Running all E2E tests..."
        npx playwright test
        ;;

    trading)
        echo "Running trading workflow tests..."
        npx playwright test trading_workflow_test.js
        ;;

    admin)
        echo "Running admin workflow tests..."
        npx playwright test admin_workflow_test.js
        ;;

    headed)
        echo "Running tests in headed mode..."
        npx playwright test --headed
        ;;

    debug)
        echo "Running tests in debug mode..."
        npx playwright test --debug
        ;;

    ui)
        echo "Opening Playwright UI..."
        npx playwright test --ui
        ;;

    report)
        echo "Opening test report..."
        npx playwright show-report
        ;;

    specific)
        if [ -z "$2" ]; then
            echo -e "${RED}Error: Please specify test name${NC}"
            echo "Usage: $0 specific \"test name\""
            exit 1
        fi
        echo "Running specific test: $2"
        npx playwright test -g "$2"
        ;;

    install)
        echo "Installing dependencies and browsers..."
        npm install
        npx playwright install
        ;;

    *)
        echo -e "${RED}Unknown test mode: $MODE${NC}"
        echo ""
        echo "Usage: $0 [MODE] [OPTIONS]"
        echo ""
        echo "Modes:"
        echo "  all       - Run all E2E tests (default)"
        echo "  trading   - Run trading workflow tests only"
        echo "  admin     - Run admin workflow tests only"
        echo "  headed    - Run with visible browser"
        echo "  debug     - Run in debug mode"
        echo "  ui        - Open Playwright UI"
        echo "  report    - Show test report"
        echo "  specific  - Run specific test (requires test name)"
        echo "  install   - Install dependencies and browsers"
        echo ""
        echo "Examples:"
        echo "  $0 all"
        echo "  $0 trading"
        echo "  $0 specific \"Complete Trading Workflow\""
        echo ""
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}=== Tests Completed ===${NC}"
