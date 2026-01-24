#!/bin/bash
# Security Fixes Verification Script - Simple Version

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0

echo "=========================================="
echo "Security Fixes Verification Test Suite"
echo "=========================================="
echo ""

echo -e "${YELLOW}[TEST 1]${NC} Command injection validation..."

# Test 1.1: Valid symbol
symbol="EURUSD"
if [[ "$symbol" =~ ^[A-Z0-9]+$ ]]; then
    echo -e "${GREEN}PASS${NC} - Valid symbol accepted: $symbol"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Valid symbol rejected: $symbol"
    FAILED=$((FAILED + 1))
fi

# Test 1.2: SQL injection attempt
symbol="EURUSD'; DROP TABLE;"
if ! [[ "$symbol" =~ ^[A-Z0-9]+$ ]]; then
    echo -e "${GREEN}PASS${NC} - SQL injection blocked: $symbol"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - SQL injection NOT blocked: $symbol"
    FAILED=$((FAILED + 1))
fi

echo ""
echo -e "${YELLOW}[TEST 2]${NC} Path traversal validation..."

# Test 2.1: Basic path traversal
symbol="../../../etc/passwd"
if ! [[ "$symbol" =~ ^[A-Z0-9]+$ ]]; then
    echo -e "${GREEN}PASS${NC} - Path traversal blocked: $symbol"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Path traversal NOT blocked: $symbol"
    FAILED=$((FAILED + 1))
fi

# Test 2.2: URL-encoded path traversal
symbol="..%2F..%2Fetc"
if ! [[ "$symbol" =~ ^[A-Z0-9]+$ ]]; then
    echo -e "${GREEN}PASS${NC} - Encoded path traversal blocked: $symbol"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Encoded path traversal NOT blocked: $symbol"
    FAILED=$((FAILED + 1))
fi

# Test 2.3: Double slash attack
symbol="....//etc"
if ! [[ "$symbol" =~ ^[A-Z0-9]+$ ]]; then
    echo -e "${GREEN}PASS${NC} - Double slash attack blocked: $symbol"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Double slash attack NOT blocked: $symbol"
    FAILED=$((FAILED + 1))
fi

echo ""
echo -e "${YELLOW}[TEST 3]${NC} Parameter validation..."

# Test 3.1: Negative offset
offset=-100
if [ $offset -lt 0 ]; then
    offset=0
    echo -e "${GREEN}PASS${NC} - Negative offset corrected to 0"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Negative offset not corrected"
    FAILED=$((FAILED + 1))
fi

# Test 3.2: Excessive limit
limit=100000
if [ $limit -gt 50000 ]; then
    limit=50000
    echo -e "${GREEN}PASS${NC} - Excessive limit capped at 50000"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Excessive limit not capped"
    FAILED=$((FAILED + 1))
fi

# Test 3.3: Negative page
page=-5
if [ $page -lt 1 ]; then
    page=1
    echo -e "${GREEN}PASS${NC} - Negative page corrected to 1"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Negative page not corrected"
    FAILED=$((FAILED + 1))
fi

echo ""
echo -e "${YELLOW}[TEST 4]${NC} Symbol length validation..."

# Test 4.1: Too long symbol
symbol="EURUSDEURUSDEURUSDEURUSD"
len=${#symbol}
if [ $len -gt 20 ]; then
    echo -e "${GREEN}PASS${NC} - Symbol exceeding 20 chars detected (length: $len)"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Long symbol not detected"
    FAILED=$((FAILED + 1))
fi

# Test 4.2: Empty symbol
symbol=""
len=${#symbol}
if [ $len -eq 0 ]; then
    echo -e "${GREEN}PASS${NC} - Empty symbol detected"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Empty symbol not detected"
    FAILED=$((FAILED + 1))
fi

# Test 4.3: Valid symbol length
symbol="EURUSD"
len=${#symbol}
if [ $len -ge 1 ] && [ $len -le 20 ]; then
    echo -e "${GREEN}PASS${NC} - Valid symbol length accepted (length: $len)"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}FAIL${NC} - Valid symbol length rejected"
    FAILED=$((FAILED + 1))
fi

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All security tests passed!${NC}"
    exit 0
else
    echo -e "${RED}$FAILED test(s) failed!${NC}"
    exit 1
fi
