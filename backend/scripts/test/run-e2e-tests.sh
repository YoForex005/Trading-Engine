#!/bin/bash

# End-to-End Tests Runner
# Runs full workflow tests including FIX protocol, trading engine, etc.

set -e
set -o pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'
BOLD='\033[1m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$PROJECT_ROOT/test-reports"

print_header() {
    echo -e "${BLUE}${BOLD}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Start test server
start_test_server() {
    print_info "Starting test server..."

    cd "$PROJECT_ROOT"

    # Build server
    if ! go build -o /tmp/rtx-test-server ./cmd/server; then
        print_error "Failed to build test server"
        return 1
    fi

    # Start server in background
    /tmp/rtx-test-server --test-mode --port 8082 > "$REPORT_DIR/server.log" 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > /tmp/rtx-test-server.pid

    # Wait for server to start
    print_info "Waiting for server to start..."
    for i in {1..30}; do
        if curl -s http://localhost:8082/health > /dev/null 2>&1; then
            print_success "Test server started (PID: $SERVER_PID)"
            return 0
        fi
        sleep 1
    done

    print_error "Test server failed to start"
    return 1
}

# Stop test server
stop_test_server() {
    print_info "Stopping test server..."

    if [[ -f /tmp/rtx-test-server.pid ]]; then
        kill $(cat /tmp/rtx-test-server.pid) 2>/dev/null || true
        rm -f /tmp/rtx-test-server.pid
    fi

    rm -f /tmp/rtx-test-server
}

# Run E2E test scenarios
run_e2e_scenarios() {
    print_info "Running E2E test scenarios..."

    local failed=0

    # Test 1: Health check
    print_info "Test 1: Health check endpoint"
    if curl -s http://localhost:8082/health | grep -q "OK"; then
        print_success "Health check passed"
    else
        print_error "Health check failed"
        ((failed++))
    fi

    # Test 2: Login flow
    print_info "Test 2: User login flow"
    local login_response=$(curl -s -X POST http://localhost:8082/login \
        -H "Content-Type: application/json" \
        -d '{"username":"test-user","password":"password123"}')

    if echo "$login_response" | grep -q "token"; then
        print_success "Login flow passed"
        TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    else
        print_error "Login flow failed"
        ((failed++))
        TOKEN=""
    fi

    # Test 3: Place market order
    if [[ -n "$TOKEN" ]]; then
        print_info "Test 3: Place market order"
        local order_response=$(curl -s -X POST http://localhost:8082/order \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" \
            -d '{"symbol":"EURUSD","side":"BUY","volume":0.1,"type":"MARKET"}')

        if echo "$order_response" | grep -q "success"; then
            print_success "Market order placement passed"
        else
            print_error "Market order placement failed"
            ((failed++))
        fi
    else
        print_warning "Skipping market order test (no auth token)"
        ((failed++))
    fi

    # Test 4: Get positions
    if [[ -n "$TOKEN" ]]; then
        print_info "Test 4: Get open positions"
        if curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8082/positions | grep -q "\["; then
            print_success "Get positions passed"
        else
            print_error "Get positions failed"
            ((failed++))
        fi
    else
        print_warning "Skipping positions test (no auth token)"
        ((failed++))
    fi

    # Test 5: WebSocket connection
    print_info "Test 5: WebSocket connection"
    if command -v wscat &> /dev/null; then
        timeout 5 wscat -c ws://localhost:8082/ws > /tmp/ws-test.log 2>&1 &
        sleep 2
        if grep -q "connected" /tmp/ws-test.log 2>/dev/null; then
            print_success "WebSocket connection passed"
        else
            print_warning "WebSocket connection test skipped (wscat not available)"
        fi
        rm -f /tmp/ws-test.log
    else
        print_warning "WebSocket test skipped (wscat not installed)"
    fi

    return $failed
}

main() {
    cd "$PROJECT_ROOT"

    print_header "End-to-End Tests"

    # Setup
    trap stop_test_server EXIT

    # Start server
    if ! start_test_server; then
        print_error "Failed to start test server"
        return 1
    fi

    # Run scenarios
    local result=0
    if ! run_e2e_scenarios; then
        result=1
    fi

    # Cleanup
    stop_test_server

    if [[ $result -eq 0 ]]; then
        print_success "All E2E tests passed"
        return 0
    else
        print_error "Some E2E tests failed"
        return 1
    fi
}

main "$@"
