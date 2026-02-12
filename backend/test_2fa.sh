#!/bin/bash
# Test script for 2FA implementation
# Usage: ./test_2fa.sh

set -e

BASE_URL="http://localhost:7999"
USERNAME="admin"
PASSWORD="Admin@123"
ADMIN_ID=1

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "RTX5 2FA Implementation Test Suite"
echo "========================================="
echo ""

# Step 1: Login
echo -e "${YELLOW}[1/6] Testing initial login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
  echo -e "${GREEN}✓ Login successful${NC}"
  echo "  Token: ${TOKEN:0:20}..."
else
  echo -e "${RED}✗ Login failed${NC}"
  echo "  Response: $LOGIN_RESPONSE"
  exit 1
fi

# Step 2: Setup 2FA
echo ""
echo -e "${YELLOW}[2/6] Setting up 2FA...${NC}"
SETUP_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/2fa/setup" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"adminId\":$ADMIN_ID}")

SECRET=$(echo $SETUP_RESPONSE | jq -r '.secret')
QR_URL=$(echo $SETUP_RESPONSE | jq -r '.qrCodeUrl')

if [ "$SECRET" != "null" ] && [ "$SECRET" != "" ]; then
  echo -e "${GREEN}✓ 2FA setup initiated${NC}"
  echo "  Secret: ${SECRET:0:20}..."
  echo "  QR URL: ${QR_URL:0:50}..."
  echo ""
  echo -e "${YELLOW}⚠️  ACTION REQUIRED:${NC}"
  echo "  1. Scan this QR code with Google Authenticator:"
  echo "     $QR_URL"
  echo "  2. Or manually enter secret: $SECRET"
  echo ""
else
  echo -e "${RED}✗ 2FA setup failed${NC}"
  echo "  Response: $SETUP_RESPONSE"
  exit 1
fi

# Step 3: Get TOTP code from user
echo -e "${YELLOW}[3/6] Waiting for TOTP code...${NC}"
read -p "Enter 6-digit code from authenticator app: " TOTP_CODE

if [ ${#TOTP_CODE} -ne 6 ]; then
  echo -e "${RED}✗ Invalid code length (must be 6 digits)${NC}"
  exit 1
fi

# Step 4: Verify code
echo ""
echo -e "${YELLOW}[4/6] Verifying TOTP code...${NC}"
VERIFY_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/2fa/verify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"adminId\":$ADMIN_ID,\"code\":\"$TOTP_CODE\"}")

SUCCESS=$(echo $VERIFY_RESPONSE | jq -r '.success')

if [ "$SUCCESS" = "true" ]; then
  echo -e "${GREEN}✓ TOTP code verified${NC}"
else
  echo -e "${RED}✗ TOTP verification failed${NC}"
  echo "  Response: $VERIFY_RESPONSE"
  exit 1
fi

# Step 5: Enable 2FA
echo ""
echo -e "${YELLOW}[5/6] Enabling 2FA...${NC}"
read -p "Enter another 6-digit code to enable: " ENABLE_CODE

ENABLE_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/2fa/enable" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"adminId\":$ADMIN_ID,\"code\":\"$ENABLE_CODE\"}")

BACKUP_CODES=$(echo $ENABLE_RESPONSE | jq -r '.backupCodes')

if [ "$BACKUP_CODES" != "null" ] && [ "$BACKUP_CODES" != "" ]; then
  echo -e "${GREEN}✓ 2FA enabled successfully${NC}"
  echo ""
  echo -e "${YELLOW}⚠️  IMPORTANT: Save these backup codes:${NC}"
  echo "$BACKUP_CODES" | jq -r '.[]' | while read code; do
    echo "  • $code"
  done
  echo ""
else
  echo -e "${RED}✗ Failed to enable 2FA${NC}"
  echo "  Response: $ENABLE_RESPONSE"
  exit 1
fi

# Step 6: Test login with 2FA
echo ""
echo -e "${YELLOW}[6/6] Testing login with 2FA...${NC}"

# First step - password
LOGIN_2FA=$(curl -s -X POST "$BASE_URL/admin/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

REQUIRES_2FA=$(echo $LOGIN_2FA | jq -r '.requires2FA')
PARTIAL_TOKEN=$(echo $LOGIN_2FA | jq -r '.partialToken')

if [ "$REQUIRES_2FA" = "true" ]; then
  echo -e "${GREEN}✓ Login correctly requires 2FA${NC}"
  echo "  Partial token: ${PARTIAL_TOKEN:0:20}..."
else
  echo -e "${RED}✗ Login did not require 2FA${NC}"
  echo "  Response: $LOGIN_2FA"
  exit 1
fi

# Second step - TOTP
read -p "Enter current TOTP code for final validation: " FINAL_CODE

VALIDATE_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/2fa/validate" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$PARTIAL_TOKEN\",\"code\":\"$FINAL_CODE\"}")

FINAL_SESSION=$(echo $VALIDATE_RESPONSE | jq -r '.sessionId')

if [ "$FINAL_SESSION" != "null" ] && [ "$FINAL_SESSION" != "" ]; then
  echo -e "${GREEN}✓ 2FA validation successful${NC}"
  echo "  Full session ID: ${FINAL_SESSION:0:20}..."
else
  echo -e "${RED}✗ 2FA validation failed${NC}"
  echo "  Response: $VALIDATE_RESPONSE"
  exit 1
fi

# Success summary
echo ""
echo "========================================="
echo -e "${GREEN}✓ All tests passed!${NC}"
echo "========================================="
echo ""
echo "2FA is now active for admin account."
echo ""
echo "Next steps:"
echo "  • Save backup codes in a secure location"
echo "  • Test backup code login if needed"
echo "  • Configure database persistence for production"
echo ""
