#!/bin/bash

# ============================================================================
# REAL DATA ONLY - Backend Testing Script
# ============================================================================
# Purpose: Validate that Trading Engine uses ONLY real YOFX data
# NO simulation code should be active
# ============================================================================

set -e  # Exit on error

# Configuration
API_BASE="http://localhost:7999"
LOGFILE="test_real_data_$(date +%Y%m%d_%H%M%S).log"
PASS_COUNT=0
FAIL_COUNT=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOGFILE"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1" | tee -a "$LOGFILE"
    ((PASS_COUNT++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1" | tee -a "$LOGFILE"
    ((FAIL_COUNT++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOGFILE"
}

# ============================================================================
# TEST 1: FIX Connection Establishment
# ============================================================================
test_fix_connections() {
    log "========================================="
    log "TEST 1: FIX Connection Establishment"
    log "========================================="

    # Check YOFX1 (Trading Account) connection
    log "Checking YOFX1 connection status..."
    YOFX1_STATUS=$(curl -s "$API_BASE/admin/fix/status" | jq -r '.sessions.YOFX1 // "UNKNOWN"')

    if [ "$YOFX1_STATUS" = "LOGGED_IN" ]; then
        log_pass "YOFX1 session is LOGGED_IN"
    else
        log_fail "YOFX1 session status: $YOFX1_STATUS (expected: LOGGED_IN)"
    fi

    # Check YOFX2 (Market Data Feed) connection
    log "Checking YOFX2 connection status..."
    YOFX2_STATUS=$(curl -s "$API_BASE/admin/fix/status" | jq -r '.sessions.YOFX2 // "UNKNOWN"')

    if [ "$YOFX2_STATUS" = "LOGGED_IN" ]; then
        log_pass "YOFX2 session is LOGGED_IN"
    else
        log_fail "YOFX2 session status: $YOFX2_STATUS (expected: LOGGED_IN)"
    fi

    # Verify no other sessions are active (should only have YOFX1 and YOFX2)
    SESSION_COUNT=$(curl -s "$API_BASE/admin/fix/status" | jq '.sessions | length')
    if [ "$SESSION_COUNT" -eq 2 ]; then
        log_pass "Only 2 FIX sessions active (YOFX1, YOFX2)"
    else
        log_warn "Found $SESSION_COUNT FIX sessions (expected: 2)"
    fi
}

# ============================================================================
# TEST 2: Market Data Subscriptions
# ============================================================================
test_market_data_subscriptions() {
    log "========================================="
    log "TEST 2: Market Data Subscriptions"
    log "========================================="

    # Get list of subscribed symbols
    log "Fetching subscribed symbols..."
    SUBSCRIBED=$(curl -s "$API_BASE/api/symbols/subscribed")
    SYMBOL_COUNT=$(echo "$SUBSCRIBED" | jq 'length')

    log "Found $SYMBOL_COUNT subscribed symbols"

    if [ "$SYMBOL_COUNT" -gt 0 ]; then
        log_pass "Market data subscriptions active ($SYMBOL_COUNT symbols)"

        # Verify major pairs are subscribed
        for SYMBOL in "EURUSD" "GBPUSD" "USDJPY" "XAUUSD"; do
            IS_SUBSCRIBED=$(echo "$SUBSCRIBED" | jq -r "any(. == \"$SYMBOL\")")
            if [ "$IS_SUBSCRIBED" = "true" ]; then
                log_pass "  - $SYMBOL is subscribed"
            else
                log_warn "  - $SYMBOL is NOT subscribed"
            fi
        done
    else
        log_fail "No market data subscriptions found"
    fi

    # Test subscribing to a new symbol
    log "Testing dynamic subscription for EURUSD..."
    SUBSCRIBE_RESULT=$(curl -s -X POST "$API_BASE/api/symbols/subscribe" \
        -H "Content-Type: application/json" \
        -d '{"symbol":"EURUSD"}')

    SUCCESS=$(echo "$SUBSCRIBE_RESULT" | jq -r '.success')
    if [ "$SUCCESS" = "true" ]; then
        log_pass "Successfully subscribed to EURUSD"
    else
        ERROR_MSG=$(echo "$SUBSCRIBE_RESULT" | jq -r '.error // "Unknown error"')
        log_fail "Failed to subscribe to EURUSD: $ERROR_MSG"
    fi
}

# ============================================================================
# TEST 3: Tick Broadcasting with LP Tag Verification
# ============================================================================
test_tick_broadcasting() {
    log "========================================="
    log "TEST 3: Tick Broadcasting & LP Tags"
    log "========================================="

    # Get market data diagnostics
    log "Checking market data diagnostics..."
    DIAGNOSTICS=$(curl -s "$API_BASE/api/diagnostics/market-data")

    TOTAL_TICKS=$(echo "$DIAGNOSTICS" | jq -r '.totalTicksReceived // 0')
    ACTIVE_STREAMS=$(echo "$DIAGNOSTICS" | jq -r '.activeStreams // 0')

    log "Total ticks received: $TOTAL_TICKS"
    log "Active streams: $ACTIVE_STREAMS"

    if [ "$TOTAL_TICKS" -gt 0 ]; then
        log_pass "Ticks are being received ($TOTAL_TICKS total)"
    else
        log_fail "No ticks received yet"
    fi

    # Verify LP tags are "YOFX" (not "SIM" or "Simulated")
    log "Verifying LP tags in latest ticks..."
    LP_TAGS=$(echo "$DIAGNOSTICS" | jq -r '.latestTicks | to_entries[] | .value.LP // "UNKNOWN"' | sort -u)

    HAS_SIM=false
    while IFS= read -r LP; do
        log "Found LP tag: $LP"
        if [[ "$LP" == "SIM"* ]] || [[ "$LP" == "Simulated"* ]]; then
            HAS_SIM=true
            log_fail "  - SIMULATION LP TAG FOUND: $LP (NOT ALLOWED IN PRODUCTION)"
        elif [ "$LP" = "YOFX" ]; then
            log_pass "  - Real data LP tag: $LP"
        else
            log_warn "  - Unknown LP tag: $LP"
        fi
    done <<< "$LP_TAGS"

    if [ "$HAS_SIM" = true ]; then
        log_fail "CRITICAL: Simulation LP tags detected in market data"
    else
        log_pass "No simulation LP tags found - using real data only"
    fi

    # Check specific symbol ticks
    log "Verifying specific symbol ticks..."
    for SYMBOL in "EURUSD" "GBPUSD" "XAUUSD"; do
        TICK=$(echo "$DIAGNOSTICS" | jq -r ".latestTicks.\"$SYMBOL\"")
        if [ "$TICK" != "null" ]; then
            BID=$(echo "$TICK" | jq -r '.bid')
            ASK=$(echo "$TICK" | jq -r '.ask')
            LP=$(echo "$TICK" | jq -r '.LP')
            log_pass "  - $SYMBOL: Bid=$BID Ask=$ASK LP=$LP"

            # Verify LP is YOFX
            if [ "$LP" = "YOFX" ]; then
                log_pass "    ✓ LP tag is YOFX (real data)"
            else
                log_fail "    ✗ LP tag is '$LP' (expected: YOFX)"
            fi
        else
            log_warn "  - $SYMBOL: No tick data available"
        fi
    done
}

# ============================================================================
# TEST 4: Database Storage Validation
# ============================================================================
test_database_storage() {
    log "========================================="
    log "TEST 4: Database Storage Validation"
    log "========================================="

    # Check SQLite database for SIM entries
    DB_PATH="data/ticks/ticks_BROKER-001_$(date +%Y%m%d).db"

    if [ -f "$DB_PATH" ]; then
        log "Checking database: $DB_PATH"

        # Count total tick entries
        TOTAL_TICKS=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks;" 2>/dev/null || echo "0")
        log "Total ticks in database: $TOTAL_TICKS"

        if [ "$TOTAL_TICKS" -gt 0 ]; then
            log_pass "Database contains $TOTAL_TICKS ticks"

            # Check for SIM LP tags
            SIM_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks WHERE lp LIKE '%SIM%' OR lp LIKE '%Simulated%';" 2>/dev/null || echo "0")

            if [ "$SIM_COUNT" -eq 0 ]; then
                log_pass "No simulation entries found in database"
            else
                log_fail "CRITICAL: Found $SIM_COUNT simulation entries in database"

                # Show sample SIM entries
                log "Sample simulation entries:"
                sqlite3 "$DB_PATH" "SELECT symbol, lp, timestamp FROM ticks WHERE lp LIKE '%SIM%' LIMIT 5;" 2>/dev/null || true
            fi

            # Verify YOFX entries exist
            YOFX_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks WHERE lp = 'YOFX';" 2>/dev/null || echo "0")
            log "YOFX entries: $YOFX_COUNT"

            if [ "$YOFX_COUNT" -gt 0 ]; then
                log_pass "Database contains $YOFX_COUNT YOFX entries"
            else
                log_warn "No YOFX entries found in database"
            fi

        else
            log_warn "Database is empty or not accessible"
        fi
    else
        log_warn "Database file not found: $DB_PATH"
    fi
}

# ============================================================================
# TEST 5: WebSocket Broadcasting Verification
# ============================================================================
test_websocket_broadcasting() {
    log "========================================="
    log "TEST 5: WebSocket Broadcasting"
    log "========================================="

    # Test WebSocket endpoint availability
    log "Testing WebSocket endpoint..."
    WS_TEST=$(curl -s -I "$API_BASE/ws" | head -n 1)

    if echo "$WS_TEST" | grep -q "101"; then
        log_pass "WebSocket endpoint is available (Switching Protocols)"
    elif echo "$WS_TEST" | grep -q "400"; then
        log_pass "WebSocket endpoint responds (400 = needs upgrade header, normal)"
    else
        log_warn "WebSocket endpoint response: $WS_TEST"
    fi

    # Check real-time tick endpoint
    log "Checking real-time tick data via HTTP..."
    TICK_DATA=$(curl -s "$API_BASE/admin/fix/ticks")

    TOTAL_TICK_COUNT=$(echo "$TICK_DATA" | jq -r '.totalTickCount // 0')
    SYMBOL_COUNT=$(echo "$TICK_DATA" | jq -r '.symbolCount // 0')

    log "WebSocket hub stats: $TOTAL_TICK_COUNT ticks, $SYMBOL_COUNT symbols"

    if [ "$TOTAL_TICK_COUNT" -gt 0 ]; then
        log_pass "WebSocket hub is broadcasting ticks ($TOTAL_TICK_COUNT total)"
    else
        log_fail "WebSocket hub has no tick data"
    fi

    # Verify latest ticks have YOFX LP tags
    log "Verifying LP tags in WebSocket data..."
    WS_LP_TAGS=$(echo "$TICK_DATA" | jq -r '.latestTicks | to_entries[] | .value.LP // "UNKNOWN"' | sort -u)

    while IFS= read -r LP; do
        if [[ "$LP" == "SIM"* ]]; then
            log_fail "  - Found SIM LP tag in WebSocket: $LP"
        elif [ "$LP" = "YOFX" ]; then
            log_pass "  - WebSocket LP tag: $LP"
        fi
    done <<< "$WS_LP_TAGS"
}

# ============================================================================
# TEST 6: Code Audit for Simulation References
# ============================================================================
test_code_audit() {
    log "========================================="
    log "TEST 6: Code Audit - Simulation References"
    log "========================================="

    # Search for simulation-related code in backend
    log "Searching for simulation code patterns..."

    # Check main.go for simulation logic
    if grep -q "Simulated\|SIM\|simulation" backend/cmd/server/main.go 2>/dev/null; then
        SIM_LINES=$(grep -n "Simulated\|SIM\|simulation" backend/cmd/server/main.go | head -5)
        log_warn "Found simulation references in main.go:"
        echo "$SIM_LINES" | while read -r line; do
            log "  $line"
        done
    else
        log_pass "No simulation references in main.go"
    fi

    # Check for SIM LP assignments in code
    log "Checking for hardcoded 'SIM' LP assignments..."
    SIM_HARDCODE=$(grep -r "LP.*=.*\"SIM" backend/ --include="*.go" 2>/dev/null || true)

    if [ -z "$SIM_HARDCODE" ]; then
        log_pass "No hardcoded 'SIM' LP assignments found"
    else
        log_fail "Found hardcoded SIM LP assignments:"
        echo "$SIM_HARDCODE"
    fi
}

# ============================================================================
# TEST 7: Performance & Latency Validation
# ============================================================================
test_performance() {
    log "========================================="
    log "TEST 7: Performance & Latency"
    log "========================================="

    # Test API response time
    log "Testing API response times..."

    START_TIME=$(date +%s%N)
    curl -s "$API_BASE/api/diagnostics/market-data" > /dev/null
    END_TIME=$(date +%s%N)
    LATENCY_MS=$(( (END_TIME - START_TIME) / 1000000 ))

    log "API latency: ${LATENCY_MS}ms"

    if [ "$LATENCY_MS" -lt 100 ]; then
        log_pass "API latency is excellent (<100ms): ${LATENCY_MS}ms"
    elif [ "$LATENCY_MS" -lt 500 ]; then
        log_pass "API latency is acceptable (<500ms): ${LATENCY_MS}ms"
    else
        log_warn "API latency is high (>500ms): ${LATENCY_MS}ms"
    fi

    # Check tick update frequency
    log "Measuring tick update frequency..."
    TICK_COUNT_1=$(curl -s "$API_BASE/admin/fix/ticks" | jq -r '.totalTickCount // 0')
    sleep 5
    TICK_COUNT_2=$(curl -s "$API_BASE/admin/fix/ticks" | jq -r '.totalTickCount // 0')

    TICKS_PER_SEC=$(( (TICK_COUNT_2 - TICK_COUNT_1) / 5 ))
    log "Tick rate: $TICKS_PER_SEC ticks/second"

    if [ "$TICKS_PER_SEC" -gt 0 ]; then
        log_pass "Market data is flowing ($TICKS_PER_SEC ticks/sec)"
    else
        log_fail "No new ticks received in 5 seconds"
    fi
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================
main() {
    log "============================================================================"
    log "REAL DATA ONLY - Backend Testing Suite"
    log "Testing Trading Engine for YOFX real data compliance"
    log "Start Time: $(date)"
    log "============================================================================"

    # Check dependencies
    command -v curl >/dev/null 2>&1 || { log_fail "curl is required but not installed. Aborting."; exit 1; }
    command -v jq >/dev/null 2>&1 || { log_fail "jq is required but not installed. Aborting."; exit 1; }

    # Wait for server to be ready
    log "Checking if server is running..."
    if curl -s "$API_BASE/health" > /dev/null 2>&1; then
        log_pass "Server is running at $API_BASE"
    else
        log_fail "Server is not accessible at $API_BASE"
        exit 1
    fi

    # Run all tests
    test_fix_connections
    test_market_data_subscriptions
    test_tick_broadcasting
    test_database_storage
    test_websocket_broadcasting
    test_code_audit
    test_performance

    # Summary
    log "============================================================================"
    log "TEST SUMMARY"
    log "============================================================================"
    log "Total Tests Passed: $PASS_COUNT"
    log "Total Tests Failed: $FAIL_COUNT"

    if [ "$FAIL_COUNT" -eq 0 ]; then
        log_pass "ALL TESTS PASSED - Ready for production deployment"
        exit 0
    else
        log_fail "$FAIL_COUNT test(s) failed - Review issues before deployment"
        exit 1
    fi
}

# Run main function
main
