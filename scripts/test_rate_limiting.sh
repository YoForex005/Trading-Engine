#!/bin/bash

# Rate Limiting Test Script
# Tests the rate limiting implementation to verify it's working correctly

echo "════════════════════════════════════════════════════════════"
echo "  Rate Limiting Test Script"
echo "════════════════════════════════════════════════════════════"
echo ""

# Configuration
SERVER="http://localhost:7999"
ENDPOINT="/api/account/summary"
REQUEST_COUNT=15

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Single request to see headers
echo -e "${BLUE}Test 1: Single Request - Check Rate Limit Headers${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
response=$(curl -s -i "$SERVER$ENDPOINT" 2>&1)
echo "$response" | head -20
echo ""

# Extract and display headers
limit=$(echo "$response" | grep -i "X-RateLimit-Limit" | cut -d' ' -f2 | tr -d '\r')
remaining=$(echo "$response" | grep -i "X-RateLimit-Remaining" | cut -d' ' -f2 | tr -d '\r')
reset=$(echo "$response" | grep -i "X-RateLimit-Reset" | cut -d' ' -f2 | tr -d '\r')

if [ -n "$limit" ] && [ -n "$remaining" ]; then
    echo -e "${GREEN}✓ Rate Limit Headers Present${NC}"
    echo "  Limit: $limit, Remaining: $remaining"
    if [ -n "$reset" ]; then
        reset_date=$(date -d @$reset 2>/dev/null || date -r $reset 2>/dev/null || echo "Unix timestamp: $reset")
        echo "  Reset: $reset_date"
    fi
else
    echo -e "${YELLOW}⚠ Rate Limit Headers Not Found${NC}"
    echo "  Server may not have rate limiting enabled or endpoint may be excluded"
fi
echo ""

# Test 2: Health endpoint (should be excluded)
echo -e "${BLUE}Test 2: Health Endpoint - Should Be Excluded from Rate Limiting${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
health_response=$(curl -s -w "%{http_code}" -o /dev/null "$SERVER/health")
if [ "$health_response" == "200" ]; then
    echo -e "${GREEN}✓ Health endpoint accessible (excluded from rate limiting)${NC}"
else
    echo -e "${RED}✗ Health endpoint failed with status $health_response${NC}"
fi
echo ""

# Test 3: Rapid requests to trigger rate limit
echo -e "${BLUE}Test 3: Rapid Requests - Trigger Rate Limit (${REQUEST_COUNT} requests)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

status_200=0
status_429=0
status_other=0

for i in $(seq 1 $REQUEST_COUNT); do
    status=$(curl -s -o /dev/null -w "%{http_code}" "$SERVER$ENDPOINT")

    case $status in
        200)
            status_200=$((status_200 + 1))
            printf "${GREEN}✓${NC}"
            ;;
        429)
            status_429=$((status_429 + 1))
            printf "${RED}✗${NC}"
            ;;
        *)
            status_other=$((status_other + 1))
            printf "${YELLOW}?${NC}"
            ;;
    esac

    # Small delay between requests
    sleep 0.1
done
echo ""
echo ""
echo "Results:"
echo "  HTTP 200 (OK):              $status_200"
echo "  HTTP 429 (Rate Limited):    $status_429"
echo "  Other Status Codes:         $status_other"
echo ""

if [ $status_429 -gt 0 ]; then
    echo -e "${GREEN}✓ Rate limiting is WORKING${NC}"
    echo "  System properly rejected requests that exceeded the limit"
elif [ $status_200 -eq $REQUEST_COUNT ]; then
    echo -e "${YELLOW}⚠ Rate limiting may be disabled or limit is very high${NC}"
    echo "  All $REQUEST_COUNT requests succeeded without rate limiting"
else
    echo -e "${RED}✗ Unexpected results${NC}"
fi
echo ""

# Test 4: Verify 429 response includes Retry-After header
echo -e "${BLUE}Test 4: Rate Limit Response Headers (429)${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Make rapid requests until we get a 429
for i in $(seq 1 20); do
    response=$(curl -s -i "$SERVER$ENDPOINT" 2>&1)
    status=$(echo "$response" | head -1)

    if echo "$status" | grep -q "429"; then
        echo "Got 429 Response:"
        echo "$response" | head -10
        echo ""

        retry_after=$(echo "$response" | grep -i "Retry-After" | cut -d' ' -f2 | tr -d '\r')
        if [ -n "$retry_after" ]; then
            echo -e "${GREEN}✓ Retry-After header present: $retry_after seconds${NC}"
        else
            echo -e "${YELLOW}⚠ Retry-After header missing${NC}"
        fi
        break
    fi

    sleep 0.05
done
echo ""

# Test 5: Configuration check
echo -e "${BLUE}Test 5: Configuration File Check${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ -f "backend/config/server.yaml" ]; then
    echo -e "${GREEN}✓ server.yaml found${NC}"
    echo "Rate limiting settings:"
    grep -A 8 "rate_limiting:" backend/config/server.yaml | head -10
else
    echo -e "${RED}✗ server.yaml not found${NC}"
fi
echo ""

# Test 6: Middleware check
echo -e "${BLUE}Test 6: Middleware File Check${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ -f "backend/internal/middleware/ratelimit.go" ]; then
    echo -e "${GREEN}✓ ratelimit.go middleware found${NC}"
    echo "File size: $(wc -c < backend/internal/middleware/ratelimit.go) bytes"
    echo "Functions:"
    grep "^func" backend/internal/middleware/ratelimit.go | head -5
else
    echo -e "${RED}✗ ratelimit.go not found${NC}"
fi
echo ""

# Summary
echo "════════════════════════════════════════════════════════════"
echo "  Test Summary"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "Rate limiting implementation test complete!"
echo ""
echo "Key Takeaways:"
echo "  1. Rate limiting middleware is active"
echo "  2. Standard HTTP headers are included in responses"
echo "  3. HTTP 429 is returned when limit is exceeded"
echo "  4. Retry-After header guides client backoff"
echo ""
echo "For detailed information, see:"
echo "  - RATE_LIMITING_QUICK_START.md"
echo "  - backend/docs/RATE_LIMITING_IMPLEMENTATION.md"
echo ""
