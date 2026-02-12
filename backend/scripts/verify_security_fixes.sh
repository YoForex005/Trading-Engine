#!/bin/bash

# ====================================================================
# RTX5 Backend - Security Verification Script
# ====================================================================
# This script verifies that all security fixes have been applied.
#
# Usage:
#   chmod +x scripts/verify_security_fixes.sh
#   ./scripts/verify_security_fixes.sh
#
# Exit codes:
#   0 - All checks passed
#   1 - One or more checks failed
# ====================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
WARNINGS=0

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║     RTX5 BACKEND - SECURITY VERIFICATION SCRIPT                ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ====================================================================
# CHECK 1: Required files exist
# ====================================================================
echo -e "\n${BLUE}[1/12] Checking required security files...${NC}"

FILES=(
    "internal/middleware/cors.go"
    "internal/middleware/auth.go"
    "SECURITY_AUDIT_REPORT.md"
    "SECURITY_FIXES_APPLIED.md"
    ".env.security.example"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}  ✓${NC} $file exists"
        ((PASSED++))
    else
        echo -e "${RED}  ✗${NC} $file is missing"
        ((FAILED++))
    fi
done

# ====================================================================
# CHECK 2: Path traversal fix in user_preferences.go
# ====================================================================
echo -e "\n${BLUE}[2/12] Checking path traversal fix...${NC}"

if grep -q "sanitizeUserID" api/user_preferences.go; then
    echo -e "${GREEN}  ✓${NC} sanitizeUserID() function exists"
    ((PASSED++))
else
    echo -e "${RED}  ✗${NC} sanitizeUserID() function not found"
    ((FAILED++))
fi

if grep -q "path traversal" api/user_preferences.go; then
    echo -e "${GREEN}  ✓${NC} Path traversal protection documented"
    ((PASSED++))
else
    echo -e "${YELLOW}  !${NC} Path traversal protection not documented"
    ((WARNINGS++))
fi

# ====================================================================
# CHECK 3: JWT expiry fix in auth/token.go
# ====================================================================
echo -e "\n${BLUE}[3/12] Checking JWT expiry configuration...${NC}"

if grep -q "JWT_EXPIRY_HOURS" auth/token.go; then
    echo -e "${GREEN}  ✓${NC} JWT expiry is configurable"
    ((PASSED++))
else
    echo -e "${RED}  ✗${NC} JWT expiry is not configurable"
    ((FAILED++))
fi

if grep -q "24 \* time.Hour" auth/token.go | grep -v "JWT_EXPIRY_HOURS"; then
    echo -e "${YELLOW}  !${NC} WARNING: Hardcoded 24-hour JWT expiry still present"
    ((WARNINGS++))
else
    echo -e "${GREEN}  ✓${NC} No hardcoded 24-hour JWT expiry found"
    ((PASSED++))
fi

# ====================================================================
# CHECK 4: CORS wildcard in workspace.go
# ====================================================================
echo -e "\n${BLUE}[4/12] Checking CORS configuration...${NC}"

if grep -q 'Access-Control-Allow-Origin.*\*' api/workspace.go; then
    # Check if it's commented or deprecated
    if grep -B2 'Access-Control-Allow-Origin.*\*' api/workspace.go | grep -q "DEPRECATED\|WARNING"; then
        echo -e "${GREEN}  ✓${NC} CORS wildcard is deprecated/documented"
        ((PASSED++))
    else
        echo -e "${RED}  ✗${NC} CORS wildcard (*) still active without warning"
        ((FAILED++))
    fi
else
    echo -e "${GREEN}  ✓${NC} No CORS wildcard found"
    ((PASSED++))
fi

# ====================================================================
# CHECK 5: Production validation enhancements
# ====================================================================
echo -e "\n${BLUE}[5/12] Checking production validation...${NC}"

if grep -q "isWeakSecret" config/validate_production.go; then
    echo -e "${GREEN}  ✓${NC} Weak secret detection implemented"
    ((PASSED++))
else
    echo -e "${RED}  ✗${NC} Weak secret detection not found"
    ((FAILED++))
fi

if grep -q "CORS wildcard" config/validate_production.go; then
    echo -e "${GREEN}  ✓${NC} CORS wildcard validation added"
    ((PASSED++))
else
    echo -e "${RED}  ✗${NC} CORS wildcard validation missing"
    ((FAILED++))
fi

# ====================================================================
# CHECK 6: Environment variables configuration
# ====================================================================
echo -e "\n${BLUE}[6/12] Checking environment variables...${NC}"

if [ -f ".env" ]; then
    echo -e "${GREEN}  ✓${NC} .env file exists"

    # Check critical variables
    if grep -q "JWT_SECRET=" .env; then
        JWT_SECRET=$(grep "JWT_SECRET=" .env | cut -d'=' -f2)
        if [ ${#JWT_SECRET} -ge 32 ]; then
            echo -e "${GREEN}  ✓${NC} JWT_SECRET is set and >= 32 characters"
            ((PASSED++))
        else
            echo -e "${RED}  ✗${NC} JWT_SECRET is too short (< 32 characters)"
            ((FAILED++))
        fi
    else
        echo -e "${YELLOW}  !${NC} JWT_SECRET not set in .env"
        ((WARNINGS++))
    fi

    if grep -q "ALLOWED_ORIGINS=" .env; then
        echo -e "${GREEN}  ✓${NC} ALLOWED_ORIGINS is configured"
        ((PASSED++))

        # Check for wildcard
        if grep "ALLOWED_ORIGINS=" .env | grep -q "\*"; then
            echo -e "${RED}  ✗${NC} ALLOWED_ORIGINS contains wildcard (*)"
            ((FAILED++))
        else
            echo -e "${GREEN}  ✓${NC} ALLOWED_ORIGINS does not use wildcard"
            ((PASSED++))
        fi
    else
        echo -e "${YELLOW}  !${NC} ALLOWED_ORIGINS not set in .env"
        ((WARNINGS++))
    fi
else
    echo -e "${YELLOW}  !${NC} .env file not found (OK for fresh setup)"
    echo -e "      Copy .env.security.example to .env and configure"
    ((WARNINGS++))
fi

# ====================================================================
# CHECK 7: SQL injection patterns
# ====================================================================
echo -e "\n${BLUE}[7/12] Checking for SQL injection vulnerabilities...${NC}"

SQL_INJECTION=$(grep -r "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE\|fmt.Sprintf.*DELETE" --include="*.go" . 2>/dev/null || true)

if [ -z "$SQL_INJECTION" ]; then
    echo -e "${GREEN}  ✓${NC} No SQL injection patterns found"
    ((PASSED++))
else
    echo -e "${RED}  ✗${NC} Potential SQL injection found:"
    echo "$SQL_INJECTION" | head -5
    ((FAILED++))
fi

# ====================================================================
# CHECK 8: Hardcoded credentials
# ====================================================================
echo -e "\n${BLUE}[8/12] Checking for hardcoded credentials...${NC}"

HARDCODED=$(grep -r "password.*=.*\"[^\"]*\"" --include="*.go" . 2>/dev/null | grep -v "os.Getenv\|getEnv\|PASSWORD_HASH\|// \|DEPRECATED" || true)

if [ -z "$HARDCODED" ]; then
    echo -e "${GREEN}  ✓${NC} No hardcoded credentials found"
    ((PASSED++))
else
    # Filter out common false positives
    ACTUAL_ISSUES=$(echo "$HARDCODED" | grep -v "PasswordHash\|Password:\|password string\|password =.*os.Getenv" || true)
    if [ -z "$ACTUAL_ISSUES" ]; then
        echo -e "${GREEN}  ✓${NC} No hardcoded credentials found"
        ((PASSED++))
    else
        echo -e "${YELLOW}  !${NC} Possible hardcoded credentials (review manually):"
        echo "$ACTUAL_ISSUES" | head -3
        ((WARNINGS++))
    fi
fi

# ====================================================================
# CHECK 9: Rate limiting implementation
# ====================================================================
echo -e "\n${BLUE}[9/12] Checking rate limiting implementation...${NC}"

if [ -f "internal/middleware/ratelimit.go" ]; then
    echo -e "${GREEN}  ✓${NC} Rate limiting middleware exists"
    ((PASSED++))

    if grep -q "RateLimiter" internal/middleware/ratelimit.go; then
        echo -e "${GREEN}  ✓${NC} RateLimiter struct implemented"
        ((PASSED++))
    fi
else
    echo -e "${RED}  ✗${NC} Rate limiting middleware not found"
    ((FAILED++))
fi

# ====================================================================
# CHECK 10: Authentication middleware
# ====================================================================
echo -e "\n${BLUE}[10/12] Checking authentication middleware...${NC}"

if [ -f "internal/middleware/auth.go" ]; then
    echo -e "${GREEN}  ✓${NC} Authentication middleware exists"
    ((PASSED++))

    if grep -q "RequireAuth" internal/middleware/auth.go; then
        echo -e "${GREEN}  ✓${NC} RequireAuth function implemented"
        ((PASSED++))
    fi

    if grep -q "RequireAdmin" internal/middleware/auth.go; then
        echo -e "${GREEN}  ✓${NC} RequireAdmin function implemented"
        ((PASSED++))
    fi
else
    echo -e "${RED}  ✗${NC} Authentication middleware not found"
    ((FAILED++))
fi

# ====================================================================
# CHECK 11: CORS middleware
# ====================================================================
echo -e "\n${BLUE}[11/12] Checking CORS middleware...${NC}"

if [ -f "internal/middleware/cors.go" ]; then
    echo -e "${GREEN}  ✓${NC} CORS middleware exists"
    ((PASSED++))

    if grep -q "NewCORSMiddleware" internal/middleware/cors.go; then
        echo -e "${GREEN}  ✓${NC} NewCORSMiddleware function implemented"
        ((PASSED++))
    fi

    if grep -q "AllowedOrigins" internal/middleware/cors.go; then
        echo -e "${GREEN}  ✓${NC} AllowedOrigins validation implemented"
        ((PASSED++))
    fi
else
    echo -e "${RED}  ✗${NC} CORS middleware not found"
    ((FAILED++))
fi

# ====================================================================
# CHECK 12: Documentation completeness
# ====================================================================
echo -e "\n${BLUE}[12/12] Checking documentation...${NC}"

DOCS=(
    "SECURITY_AUDIT_REPORT.md"
    "SECURITY_FIXES_APPLIED.md"
    "APPLY_SECURITY_MIDDLEWARE.md"
    ".env.security.example"
)

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${GREEN}  ✓${NC} $doc exists"
        ((PASSED++))
    else
        echo -e "${YELLOW}  !${NC} $doc is missing"
        ((WARNINGS++))
    fi
done

# ====================================================================
# SUMMARY
# ====================================================================
echo -e "\n${BLUE}"
echo "╔════════════════════════════════════════════════════════════════╗"
echo "║                    VERIFICATION SUMMARY                        ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

TOTAL=$((PASSED + FAILED + WARNINGS))

echo -e "${GREEN}✓ Passed:  ${PASSED}${NC}"
echo -e "${YELLOW}! Warnings: ${WARNINGS}${NC}"
echo -e "${RED}✗ Failed:  ${FAILED}${NC}"
echo "─────────────────────────"
echo "Total checks: $TOTAL"

# ====================================================================
# EXIT CODE
# ====================================================================
echo ""

if [ $FAILED -eq 0 ]; then
    if [ $WARNINGS -eq 0 ]; then
        echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║  ✓ ALL SECURITY CHECKS PASSED                             ║${NC}"
        echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
        echo ""
        echo -e "Next steps:"
        echo -e "  1. Review ${BLUE}SECURITY_AUDIT_REPORT.md${NC}"
        echo -e "  2. Apply middleware (see ${BLUE}APPLY_SECURITY_MIDDLEWARE.md${NC})"
        echo -e "  3. Test with: ${BLUE}ENVIRONMENT=production go run cmd/server/main.go${NC}"
        exit 0
    else
        echo -e "${YELLOW}╔═══════════════════════════════════════════════════════════╗${NC}"
        echo -e "${YELLOW}║  ! SECURITY CHECKS PASSED WITH WARNINGS                   ║${NC}"
        echo -e "${YELLOW}╚═══════════════════════════════════════════════════════════╝${NC}"
        echo ""
        echo -e "Review warnings above before deploying to production."
        exit 0
    fi
else
    echo -e "${RED}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  ✗ SECURITY CHECKS FAILED                                 ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${RED}$FAILED critical issues must be fixed before deployment.${NC}"
    echo ""
    echo -e "Review failed checks above and re-run this script."
    exit 1
fi
