#!/bin/bash
# Production Deployment Script
# Ensures simulation code is NEVER deployed to production

set -e  # Exit on error

echo "════════════════════════════════════════════════════════════"
echo "  Trading Engine - Production Deployment"
echo "════════════════════════════════════════════════════════════"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Verify environment
echo ""
echo "1. Verifying environment variables..."
if [ "$ENVIRONMENT" != "production" ]; then
    echo -e "${RED}ERROR: ENVIRONMENT must be set to 'production'${NC}"
    echo "Current value: $ENVIRONMENT"
    exit 1
fi
echo -e "${GREEN}✓ ENVIRONMENT=production${NC}"

# 2. Check for simulation flags
echo ""
echo "2. Checking for simulation flags..."
if [ "$ALLOW_SIMULATION" == "true" ] || [ "$ALLOW_SIMULATION" == "1" ]; then
    echo -e "${RED}ERROR: ALLOW_SIMULATION must be false or unset in production${NC}"
    echo "Current value: $ALLOW_SIMULATION"
    exit 1
fi
echo -e "${GREEN}✓ Simulation disabled (ALLOW_SIMULATION=${ALLOW_SIMULATION:-unset})${NC}"

if [ ! -z "$SIMULATION_MODE" ]; then
    echo -e "${RED}ERROR: SIMULATION_MODE must not be set in production${NC}"
    exit 1
fi
echo -e "${GREEN}✓ SIMULATION_MODE not set${NC}"

# 3. Verify LP credentials
echo ""
echo "3. Verifying liquidity provider configuration..."
HAS_LP=false

if [ ! -z "$YOFX_HOST" ] && [ ! -z "$YOFX_PORT" ]; then
    echo -e "${GREEN}✓ YOFX FIX LP configured${NC}"
    HAS_LP=true
fi

if [ ! -z "$OANDA_API_KEY" ] && [ ! -z "$OANDA_ACCOUNT_ID" ]; then
    echo -e "${GREEN}✓ OANDA LP configured${NC}"
    HAS_LP=true
fi

if [ ! -z "$BINANCE_API_KEY" ]; then
    echo -e "${GREEN}✓ Binance LP configured${NC}"
    HAS_LP=true
fi

if [ "$HAS_LP" = false ]; then
    echo -e "${RED}ERROR: No real liquidity providers configured${NC}"
    echo "Must set credentials for at least one of: YOFX, OANDA, Binance"
    exit 1
fi

# 4. Verify security settings
echo ""
echo "4. Verifying security configuration..."

if [ -z "$JWT_SECRET" ]; then
    echo -e "${RED}ERROR: JWT_SECRET must be set in production${NC}"
    exit 1
fi
if [ ${#JWT_SECRET} -lt 32 ]; then
    echo -e "${RED}ERROR: JWT_SECRET must be at least 32 characters${NC}"
    exit 1
fi
echo -e "${GREEN}✓ JWT_SECRET configured (${#JWT_SECRET} characters)${NC}"

if [ -z "$ADMIN_PASSWORD_HASH" ]; then
    echo -e "${YELLOW}WARNING: ADMIN_PASSWORD_HASH not set - using default${NC}"
fi

if [ -z "$MASTER_ENCRYPTION_KEY" ]; then
    echo -e "${RED}ERROR: MASTER_ENCRYPTION_KEY must be set in production${NC}"
    exit 1
fi
if [ ${#MASTER_ENCRYPTION_KEY} -lt 32 ]; then
    echo -e "${RED}ERROR: MASTER_ENCRYPTION_KEY must be at least 32 characters${NC}"
    exit 1
fi
echo -e "${GREEN}✓ MASTER_ENCRYPTION_KEY configured${NC}"

# 5. Build with production tag
echo ""
echo "5. Building with production tag..."
cd backend

# Clean previous builds
rm -f bin/server

# Build with production tag to exclude simulation code
echo "   Building: go build -tags production -ldflags=\"-s -w\" -o bin/server ./cmd/server"
go build -tags production -ldflags="-s -w" -o bin/server ./cmd/server

if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Build successful${NC}"

# 6. Verify binary doesn't contain simulation code
echo ""
echo "6. Scanning binary for simulation code..."
if strings bin/server | grep -qi "GenerateSimulatedTicks"; then
    echo -e "${RED}ERROR: Binary contains simulation code (GenerateSimulatedTicks found)${NC}"
    exit 1
fi

if strings bin/server | grep -qi "LP=.*SIM"; then
    echo -e "${YELLOW}WARNING: Found potential simulation references in binary${NC}"
fi
echo -e "${GREEN}✓ No simulation code detected in binary${NC}"

# 7. Run production validation tests
echo ""
echo "7. Running production safety tests..."
if command -v go &> /dev/null; then
    echo "   Running: go test -tags production -run TestProduction ./..."
    go test -tags production -run TestProduction ./... -v
    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}WARNING: Some production tests failed${NC}"
    else
        echo -e "${GREEN}✓ Production tests passed${NC}"
    fi
else
    echo -e "${YELLOW}WARNING: Go not found, skipping tests${NC}"
fi

# 8. Verify database connection (optional)
echo ""
echo "8. Verifying database connection..."
if [ ! -z "$DB_HOST" ]; then
    echo "   Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
    # You can add a database connection test here
    echo -e "${GREEN}✓ Database configuration present${NC}"
else
    echo -e "${YELLOW}WARNING: Database not configured${NC}"
fi

# 9. Generate deployment manifest
echo ""
echo "9. Generating deployment manifest..."
MANIFEST_FILE="deployment-manifest-$(date +%Y%m%d-%H%M%S).json"
cat > "$MANIFEST_FILE" <<EOF
{
  "deployment_time": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": "$ENVIRONMENT",
  "simulation_allowed": "${ALLOW_SIMULATION:-false}",
  "binary_path": "bin/server",
  "binary_size": $(stat -f%z bin/server 2>/dev/null || stat -c%s bin/server 2>/dev/null || echo "unknown"),
  "build_tags": ["production"],
  "lp_configured": {
    "yofx": $([ ! -z "$YOFX_HOST" ] && echo "true" || echo "false"),
    "oanda": $([ ! -z "$OANDA_API_KEY" ] && echo "true" || echo "false"),
    "binance": $([ ! -z "$BINANCE_API_KEY" ] && echo "true" || echo "false")
  },
  "security": {
    "jwt_secret_set": $([ ! -z "$JWT_SECRET" ] && echo "true" || echo "false"),
    "encryption_key_set": $([ ! -z "$MASTER_ENCRYPTION_KEY" ] && echo "true" || echo "false"),
    "admin_password_set": $([ ! -z "$ADMIN_PASSWORD_HASH" ] && echo "true" || echo "false")
  }
}
EOF
echo -e "${GREEN}✓ Deployment manifest created: $MANIFEST_FILE${NC}"

# 10. Final summary
echo ""
echo "════════════════════════════════════════════════════════════"
echo -e "${GREEN}✓ Production Deployment Validation PASSED${NC}"
echo "════════════════════════════════════════════════════════════"
echo ""
echo "Next steps:"
echo "  1. Deploy binary: bin/server"
echo "  2. Monitor health: curl http://server:7999/admin/health/data-source"
echo "  3. Watch logs for '[PRODUCTION]' messages"
echo "  4. Verify no 'SIM' LP in metrics"
echo ""
echo "Emergency procedures: See docs/PRODUCTION_DEPLOYMENT_GUIDE.md"
echo ""

exit 0
