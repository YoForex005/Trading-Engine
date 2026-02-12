#!/bin/bash

# RBAC Testing Script for RTX5 Admin Backend
# Tests role-based access control implementation

set -e

API_URL="${API_URL:-http://localhost:7999}"
ADMIN_USER="${ADMIN_USER:-admin}"
ADMIN_PASS="${ADMIN_PASS:-Admin@123}"

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║           RTX5 RBAC Testing Script                         ║"
echo "║           API URL: $API_URL                      ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test function
test_endpoint() {
    local desc=$1
    local method=$2
    local endpoint=$3
    local token=$4
    local expected_code=$5
    local data=$6

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    echo -n "Testing: $desc ... "

    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            "$API_URL$endpoint" 2>/dev/null || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_URL$endpoint" 2>/dev/null || echo "000")
    fi

    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" == "$expected_code" ]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}✗ FAIL${NC} (Expected $expected_code, got $http_code)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        if [ "$http_code" != "000" ]; then
            echo "  Response: $(echo "$response" | head -n-1)"
        fi
    fi
}

# Login function
login_admin() {
    local username=$1
    local password=$2

    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
        "$API_URL/admin/auth/login")

    token=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

    if [ -z "$token" ]; then
        echo -e "${RED}ERROR: Failed to login as $username${NC}"
        echo "Response: $response"
        exit 1
    fi

    echo "$token"
}

echo "═══════════════════════════════════════════════════════════"
echo "Step 1: Login as SUPER_ADMIN"
echo "═══════════════════════════════════════════════════════════"

SUPER_ADMIN_TOKEN=$(login_admin "$ADMIN_USER" "$ADMIN_PASS")
echo -e "${GREEN}✓ Logged in as SUPER_ADMIN${NC}"
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "Step 2: Test Role Management Endpoints (SUPER_ADMIN)"
echo "═══════════════════════════════════════════════════════════"

test_endpoint "List all roles" "GET" "/admin/roles" "$SUPER_ADMIN_TOKEN" "200"
test_endpoint "Get user role" "GET" "/admin/users/role?userId=1" "$SUPER_ADMIN_TOKEN" "200"
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "Step 3: Test SUPER_ADMIN Access (Should Allow All)"
echo "═══════════════════════════════════════════════════════════"

test_endpoint "View users (SUPER_ADMIN)" "GET" "/admin/users" "$SUPER_ADMIN_TOKEN" "200"
test_endpoint "View orders (SUPER_ADMIN)" "GET" "/admin/orders" "$SUPER_ADMIN_TOKEN" "200"
test_endpoint "View positions (SUPER_ADMIN)" "GET" "/admin/positions" "$SUPER_ADMIN_TOKEN" "200"
test_endpoint "View groups (SUPER_ADMIN)" "GET" "/admin/groups" "$SUPER_ADMIN_TOKEN" "200"
test_endpoint "View audit log (SUPER_ADMIN)" "GET" "/admin/audit" "$SUPER_ADMIN_TOKEN" "200"
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "Step 4: Test Invalid Token"
echo "═══════════════════════════════════════════════════════════"

test_endpoint "Invalid token" "GET" "/admin/users" "invalid_token_12345" "401"
test_endpoint "Missing token" "GET" "/admin/users" "" "401"
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "Step 5: Test Permission Boundaries"
echo "═══════════════════════════════════════════════════════════"

# These tests verify permission checking is working
# Note: Without creating additional admin users, we can only test SUPER_ADMIN access
# To fully test, you would need to:
# 1. Create admin with each role (ADMIN, MANAGER, DEALER, VIEWER)
# 2. Login as each
# 3. Test their specific permissions

echo -e "${YELLOW}Note: Full permission testing requires creating admins with different roles${NC}"
echo -e "${YELLOW}Current tests verify SUPER_ADMIN has full access${NC}"
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "Step 6: Test Protected vs Unprotected Routes"
echo "═══════════════════════════════════════════════════════════"

# Login/Logout should work without RBAC
test_endpoint "Login (no auth required)" "POST" "/admin/auth/login" "" "200" "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}"

# All other routes should require auth
test_endpoint "Users without auth" "GET" "/admin/users" "" "401"
test_endpoint "Orders without auth" "GET" "/admin/orders" "" "401"
test_endpoint "Groups without auth" "GET" "/admin/groups" "" "401"
echo ""

echo "═══════════════════════════════════════════════════════════"
echo "Test Summary"
echo "═══════════════════════════════════════════════════════════"
echo "Total Tests:  $TOTAL_TESTS"
echo -e "Passed:       ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed:       ${RED}$FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
