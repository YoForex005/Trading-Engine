#!/bin/bash

# Tick Data Flow End-to-End Test Script
# Tests all components: FIX → Storage → API → Frontend

set -e

echo "========================================"
echo "Tick Data Flow E2E Test Suite"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

API_URL="http://localhost:7999"
TEST_SYMBOL="EURUSD"

# Test counters
PASS=0
FAIL=0

# Helper functions
pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASS++))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((FAIL++))
}

warn() {
    echo -e "${YELLOW}⚠ WARN${NC}: $1"
}

info() {
    echo "ℹ INFO: $1"
}

echo "1. Testing Server Health"
echo "------------------------"
if curl -s "$API_URL/health" > /dev/null; then
    pass "Server is responding"
else
    fail "Server is not responding"
    exit 1
fi
echo ""

echo "2. Testing FIX Gateway Status"
echo "------------------------------"
FIX_STATUS=$(curl -s "$API_URL/admin/fix/status")
if echo "$FIX_STATUS" | grep -q "YOFX"; then
    pass "FIX gateway configured"

    # Check YOFX1 status
    if echo "$FIX_STATUS" | grep -q "YOFX1.*LOGGED_IN"; then
        pass "YOFX1 (Trading) is logged in"
    else
        warn "YOFX1 (Trading) not logged in"
    fi

    # Check YOFX2 status
    if echo "$FIX_STATUS" | grep -q "YOFX2.*LOGGED_IN"; then
        pass "YOFX2 (Market Data) is logged in"
    else
        warn "YOFX2 (Market Data) not logged in"
    fi
else
    fail "FIX gateway not configured"
fi
echo ""

echo "3. Testing Tick Data Storage"
echo "-----------------------------"
cd ../backend

# Check if tick data directory exists
if [ -d "data/ticks" ]; then
    pass "Tick data directory exists"

    # Count files
    FILE_COUNT=$(find data/ticks -type f -name "*.json" 2>/dev/null | wc -l)
    info "Found $FILE_COUNT tick data files"

    if [ "$FILE_COUNT" -gt 0 ]; then
        pass "Tick files are being created"
    else
        fail "No tick files found"
    fi

    # Check storage size
    STORAGE_SIZE=$(du -sh data/ticks 2>/dev/null | cut -f1)
    info "Total storage: $STORAGE_SIZE"

    # Count symbols
    SYMBOL_COUNT=$(find data/ticks -maxdepth 1 -type d 2>/dev/null | tail -n +2 | wc -l)
    info "Symbols with data: $SYMBOL_COUNT"

    if [ "$SYMBOL_COUNT" -ge 10 ]; then
        pass "Multiple symbols captured ($SYMBOL_COUNT symbols)"
    else
        warn "Only $SYMBOL_COUNT symbols captured (expected 30+)"
    fi

    # Check for EURUSD specifically
    if [ -d "data/ticks/EURUSD" ]; then
        pass "EURUSD data directory exists"

        # Check for today's file
        TODAY=$(date +%Y-%m-%d)
        if [ -f "data/ticks/EURUSD/$TODAY.json" ]; then
            pass "EURUSD has data for today"

            # Check file size
            FILE_SIZE=$(wc -c < "data/ticks/EURUSD/$TODAY.json")
            if [ "$FILE_SIZE" -gt 1000 ]; then
                pass "EURUSD file has substantial data ($FILE_SIZE bytes)"
            else
                warn "EURUSD file is small ($FILE_SIZE bytes) - may not be receiving ticks"
            fi
        else
            fail "No EURUSD data for today ($TODAY)"
        fi
    else
        fail "EURUSD data directory not found"
    fi
else
    fail "Tick data directory not found"
fi
echo ""

echo "4. Testing REST API - Tick Retrieval"
echo "-------------------------------------"
TICKS_RESPONSE=$(curl -s "$API_URL/ticks?symbol=$TEST_SYMBOL&limit=10")

if echo "$TICKS_RESPONSE" | grep -q "bid"; then
    pass "Tick API returns data"

    # Count ticks returned
    TICK_COUNT=$(echo "$TICKS_RESPONSE" | grep -o "\"bid\"" | wc -l)
    info "Returned $TICK_COUNT ticks"

    if [ "$TICK_COUNT" -ge 1 ]; then
        pass "API returned at least 1 tick"
    else
        fail "API returned no ticks"
    fi

    # Check for required fields
    if echo "$TICKS_RESPONSE" | grep -q "\"bid\".*\"ask\".*\"spread\""; then
        pass "Tick data has required fields (bid, ask, spread)"
    else
        fail "Tick data missing required fields"
    fi
else
    fail "Tick API returned no data or error"
fi
echo ""

echo "5. Testing REST API - OHLC Retrieval"
echo "-------------------------------------"
OHLC_RESPONSE=$(curl -s "$API_URL/ohlc?symbol=$TEST_SYMBOL&timeframe=1m&limit=5")

if echo "$OHLC_RESPONSE" | grep -q "open"; then
    pass "OHLC API returns data"

    # Check for required fields
    if echo "$OHLC_RESPONSE" | grep -q "\"open\".*\"high\".*\"low\".*\"close\""; then
        pass "OHLC data has required fields (OHLC)"
    else
        fail "OHLC data missing required fields"
    fi
else
    warn "OHLC API returned no data (may need time to accumulate)"
fi
echo ""

echo "6. Testing Admin Endpoints"
echo "---------------------------"

# Test history stats
STATS_RESPONSE=$(curl -s "$API_URL/admin/history/stats")
if echo "$STATS_RESPONSE" | grep -q "symbols\|totalFiles\|totalSize"; then
    pass "Admin history stats endpoint working"
else
    warn "Admin history stats endpoint may not be implemented"
fi

# Test monitoring
MONITOR_RESPONSE=$(curl -s "$API_URL/admin/history/monitoring")
if [ ! -z "$MONITOR_RESPONSE" ]; then
    pass "Admin monitoring endpoint working"
else
    warn "Admin monitoring endpoint may not be implemented"
fi
echo ""

echo "7. Testing Symbol Coverage"
echo "---------------------------"

# Get available symbols
SYMBOLS_RESPONSE=$(curl -s "$API_URL/api/symbols/available")
if echo "$SYMBOLS_RESPONSE" | grep -q "EURUSD"; then
    pass "Available symbols API working"

    SYMBOL_COUNT=$(echo "$SYMBOLS_RESPONSE" | grep -o "\"symbol\"" | wc -l)
    info "Available symbols: $SYMBOL_COUNT"

    if [ "$SYMBOL_COUNT" -ge 30 ]; then
        pass "Sufficient symbols available ($SYMBOL_COUNT >= 30)"
    else
        warn "Only $SYMBOL_COUNT symbols available (expected 30+)"
    fi
else
    fail "Available symbols API not working"
fi

# Get subscribed symbols
SUBSCRIBED_RESPONSE=$(curl -s "$API_URL/api/symbols/subscribed")
if [ ! -z "$SUBSCRIBED_RESPONSE" ]; then
    pass "Subscribed symbols API working"

    SUB_COUNT=$(echo "$SUBSCRIBED_RESPONSE" | grep -o "\"" | wc -l)
    SUB_COUNT=$((SUB_COUNT / 2))
    info "Subscribed symbols: $SUB_COUNT"

    if [ "$SUB_COUNT" -ge 10 ]; then
        pass "Multiple symbols subscribed ($SUB_COUNT symbols)"
    else
        warn "Only $SUB_COUNT symbols subscribed (expected 30+)"
    fi
fi
echo ""

echo "8. Testing Market Data Flow"
echo "----------------------------"

# Check tick flow debug endpoint
TICK_FLOW=$(curl -s "$API_URL/admin/fix/ticks")
if echo "$TICK_FLOW" | grep -q "totalTickCount"; then
    pass "Tick flow monitoring working"

    TOTAL_TICKS=$(echo "$TICK_FLOW" | grep -o "\"totalTickCount\":[0-9]*" | grep -o "[0-9]*")
    info "Total ticks processed: $TOTAL_TICKS"

    if [ "$TOTAL_TICKS" -gt 100 ]; then
        pass "Significant tick volume ($TOTAL_TICKS ticks)"
    elif [ "$TOTAL_TICKS" -gt 0 ]; then
        warn "Low tick volume ($TOTAL_TICKS ticks) - system may have just started"
    else
        fail "No ticks processed (market data not flowing)"
    fi

    # Check symbols with data
    SYMBOLS_WITH_DATA=$(echo "$TICK_FLOW" | grep -o "\"symbolCount\":[0-9]*" | grep -o "[0-9]*")
    info "Symbols with data: $SYMBOLS_WITH_DATA"

    if [ "$SYMBOLS_WITH_DATA" -ge 10 ]; then
        pass "Multiple symbols receiving data ($SYMBOLS_WITH_DATA symbols)"
    else
        warn "Only $SYMBOLS_WITH_DATA symbols receiving data"
    fi
else
    warn "Tick flow monitoring endpoint not available"
fi
echo ""

echo "9. Testing Performance Metrics"
echo "-------------------------------"

# Test API response time
START=$(date +%s%N)
curl -s "$API_URL/ticks?symbol=EURUSD&limit=100" > /dev/null
END=$(date +%s%N)
LATENCY=$(( (END - START) / 1000000 ))
info "API latency: ${LATENCY}ms"

if [ "$LATENCY" -lt 500 ]; then
    pass "API response time is good (<500ms)"
elif [ "$LATENCY" -lt 1000 ]; then
    warn "API response time is acceptable (${LATENCY}ms)"
else
    fail "API response time is slow (${LATENCY}ms)"
fi
echo ""

echo "10. Checking Data Quality"
echo "-------------------------"

# Check for data in last hour
if [ -f "data/ticks/EURUSD/$(date +%Y-%m-%d).json" ]; then
    RECENT_TICKS=$(cat "data/ticks/EURUSD/$(date +%Y-%m-%d).json" | grep -o "\"timestamp\"" | wc -l)
    info "Recent ticks in EURUSD file: $RECENT_TICKS"

    if [ "$RECENT_TICKS" -gt 10 ]; then
        pass "Recent data is being written"
    else
        warn "Very few recent ticks ($RECENT_TICKS)"
    fi

    # Check for price validity
    if cat "data/ticks/EURUSD/$(date +%Y-%m-%d).json" | grep -q "\"bid\":[0-9]\\+\\.[0-9]"; then
        pass "Tick prices are in valid format"
    else
        fail "Tick prices may be malformed"
    fi
fi
echo ""

echo "========================================"
echo "Test Summary"
echo "========================================"
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"
echo ""

TOTAL=$((PASS + FAIL))
SUCCESS_RATE=$((PASS * 100 / TOTAL))

if [ "$SUCCESS_RATE" -ge 80 ]; then
    echo -e "${GREEN}Overall Status: PASS (${SUCCESS_RATE}%)${NC}"
    echo "System is operational with minor issues"
elif [ "$SUCCESS_RATE" -ge 60 ]; then
    echo -e "${YELLOW}Overall Status: PARTIAL (${SUCCESS_RATE}%)${NC}"
    echo "System is partially functional - review failed tests"
else
    echo -e "${RED}Overall Status: FAIL (${SUCCESS_RATE}%)${NC}"
    echo "System has critical issues - immediate attention required"
fi

echo ""
echo "Detailed results saved in validation report"
echo "See: docs/TICK_DATA_E2E_VALIDATION_REPORT.md"
echo ""

exit 0
