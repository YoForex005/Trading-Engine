#!/bin/bash

# Master Test Runner
# Runs both integration tests and E2E tests

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     RTX Trading Engine - Complete Test Suite             ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}\n"

# Check if backend server is running
check_backend() {
    echo -e "${YELLOW}Checking backend server...${NC}"

    if curl -s http://localhost:7999/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Backend server is running${NC}\n"
        BACKEND_RUNNING=true
    else
        echo -e "${RED}✗ Backend server is not running${NC}"
        echo "Starting backend server..."

        cd backend
        go build -o server cmd/server/main.go

        if [ $? -eq 0 ]; then
            ./server > /dev/null 2>&1 &
            SERVER_PID=$!
            echo -e "${GREEN}✓ Backend server started (PID: $SERVER_PID)${NC}"

            # Wait for server to be ready
            for i in {1..10}; do
                if curl -s http://localhost:7999/health > /dev/null 2>&1; then
                    echo -e "${GREEN}✓ Backend server is ready${NC}\n"
                    BACKEND_RUNNING=true
                    break
                fi
                sleep 1
            done

            if [ "$BACKEND_RUNNING" != "true" ]; then
                echo -e "${RED}✗ Backend server failed to start${NC}\n"
                kill $SERVER_PID 2>/dev/null || true
                exit 1
            fi
        else
            echo -e "${RED}✗ Failed to build backend server${NC}\n"
            exit 1
        fi

        cd ..
    fi
}

# Cleanup function
cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        echo -e "\n${YELLOW}Stopping backend server (PID: $SERVER_PID)...${NC}"
        kill $SERVER_PID 2>/dev/null || true
        echo -e "${GREEN}✓ Backend server stopped${NC}"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Parse arguments
INTEGRATION_ONLY=false
E2E_ONLY=false
SKIP_E2E=false

while [ $# -gt 0 ]; do
    case $1 in
        --integration-only)
            INTEGRATION_ONLY=true
            shift
            ;;
        --e2e-only)
            E2E_ONLY=true
            shift
            ;;
        --skip-e2e)
            SKIP_E2E=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --integration-only   Run only integration tests"
            echo "  --e2e-only          Run only E2E tests"
            echo "  --skip-e2e          Skip E2E tests (run integration only)"
            echo "  --help, -h          Show this help message"
            echo ""
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check backend
check_backend

# Run Integration Tests
if [ "$E2E_ONLY" != "true" ]; then
    echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         Integration Tests (Go)                            ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}\n"

    cd backend/tests/integration

    if ./run_tests.sh all; then
        echo -e "${GREEN}✓ Integration tests passed${NC}\n"
    else
        echo -e "${RED}✗ Integration tests failed${NC}\n"
        exit 1
    fi

    cd ../../..
fi

# Run E2E Tests
if [ "$INTEGRATION_ONLY" != "true" ] && [ "$SKIP_E2E" != "true" ]; then
    echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         E2E Tests (Playwright)                            ║${NC}"
    echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}\n"

    cd tests/e2e

    # Check if dependencies are installed
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}Installing E2E test dependencies...${NC}"
        npm install
        npx playwright install
        echo ""
    fi

    if ./run_tests.sh all; then
        echo -e "${GREEN}✓ E2E tests passed${NC}\n"
    else
        echo -e "${RED}✗ E2E tests failed${NC}\n"
        exit 1
    fi

    cd ../..
fi

# Summary
echo -e "${BLUE}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Test Summary                                      ║${NC}"
echo -e "${BLUE}╚═══════════════════════════════════════════════════════════╝${NC}\n"

if [ "$E2E_ONLY" != "true" ]; then
    echo -e "${GREEN}✓ Integration Tests: PASSED${NC}"
fi

if [ "$INTEGRATION_ONLY" != "true" ] && [ "$SKIP_E2E" != "true" ]; then
    echo -e "${GREEN}✓ E2E Tests: PASSED${NC}"
fi

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║     All Tests Completed Successfully! 🎉                  ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}\n"
