#!/bin/bash
# Monitor progress of Agents 1, 2, 3 for MT5 parity implementation

echo "==================================================================="
echo "MT5 Parity Implementation - Agent Progress Monitor"
echo "==================================================================="
echo ""
echo "Checking status of Agents 1, 2, 3..."
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Agent 1: Backend Throttling Config
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Agent 1: Backend Throttling Config (backend/ws/hub.go)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if grep -q "mt5Mode" backend/ws/hub.go 2>/dev/null; then
    echo -e "${GREEN}✅ COMPLETED${NC} - MT5 mode flag found in hub.go"

    # Check for environment variable support
    if grep -q "MT5_MODE" backend/ws/hub.go backend/cmd/server/main.go 2>/dev/null; then
        echo -e "${GREEN}✅ COMPLETED${NC} - Environment variable support found"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC} - MT5 flag exists but no env variable support"
    fi

    # Check for throttling bypass
    if grep -q "if.*mt5Mode" backend/ws/hub.go 2>/dev/null; then
        echo -e "${GREEN}✅ COMPLETED${NC} - Throttling bypass logic implemented"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC} - Throttling bypass not found"
    fi
else
    echo -e "${RED}❌ PENDING${NC} - No mt5Mode flag found"
fi
echo ""

# Agent 2: Flash Animations
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Agent 2: Flash Animations (MarketWatchPanel.tsx)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if grep -q "flashStates" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null; then
    echo -e "${GREEN}✅ COMPLETED${NC} - Flash state management found"

    # Check for CSS animations
    if grep -q "animate-flash" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null || \
       grep -q "flash-green\|flash-red" clients/desktop/src/app/globals.css clients/desktop/src/index.css 2>/dev/null; then
        echo -e "${GREEN}✅ COMPLETED${NC} - Flash CSS animations found"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC} - Flash state exists but CSS missing"
    fi

    # Check for flash clearing timeout
    if grep -q "setTimeout.*flashStates" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null; then
        echo -e "${GREEN}✅ COMPLETED${NC} - Flash clearing timeout implemented"
    else
        echo -e "${YELLOW}⚠️  PARTIAL${NC} - Flash clearing timeout not found"
    fi
else
    echo -e "${RED}❌ PENDING${NC} - No flash state management found"
fi
echo ""

# Agent 3: State Consolidation
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Agent 3: State Consolidation (App.tsx)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if local ticks state is removed
if ! grep -q "const \[ticks, setTicks\]" clients/desktop/src/App.tsx 2>/dev/null; then
    echo -e "${GREEN}✅ COMPLETED${NC} - Local ticks state removed from App.tsx"
else
    echo -e "${RED}❌ PENDING${NC} - Local ticks state still exists in App.tsx"
fi

# Check if MarketWatchPanel uses Zustand
if grep -q "useAppStore.*ticks" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null; then
    echo -e "${GREEN}✅ COMPLETED${NC} - MarketWatchPanel uses Zustand for ticks"
else
    echo -e "${RED}❌ PENDING${NC} - MarketWatchPanel still uses props for ticks"
fi

# Check if ticks prop removed from MarketWatchPanel call
if ! grep -q "ticks={ticks}" clients/desktop/src/App.tsx 2>/dev/null; then
    echo -e "${GREEN}✅ COMPLETED${NC} - ticks prop removed from MarketWatchPanel"
else
    echo -e "${RED}❌ PENDING${NC} - ticks prop still passed to MarketWatchPanel"
fi
echo ""

# Overall Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Overall Status Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

AGENT1_DONE=0
AGENT2_DONE=0
AGENT3_DONE=0

# Agent 1 check
if grep -q "mt5Mode" backend/ws/hub.go 2>/dev/null && \
   grep -q "if.*mt5Mode" backend/ws/hub.go 2>/dev/null; then
    AGENT1_DONE=1
fi

# Agent 2 check
if grep -q "flashStates" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null && \
   grep -q "setTimeout.*flashStates" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null; then
    AGENT2_DONE=1
fi

# Agent 3 check
if ! grep -q "const \[ticks, setTicks\]" clients/desktop/src/App.tsx 2>/dev/null && \
   grep -q "useAppStore.*ticks" clients/desktop/src/components/layout/MarketWatchPanel.tsx 2>/dev/null; then
    AGENT3_DONE=1
fi

TOTAL_DONE=$((AGENT1_DONE + AGENT2_DONE + AGENT3_DONE))

echo ""
if [ $AGENT1_DONE -eq 1 ]; then
    echo -e "Agent 1: ${GREEN}✅ COMPLETED${NC}"
else
    echo -e "Agent 1: ${RED}❌ PENDING${NC}"
fi

if [ $AGENT2_DONE -eq 1 ]; then
    echo -e "Agent 2: ${GREEN}✅ COMPLETED${NC}"
else
    echo -e "Agent 2: ${RED}❌ PENDING${NC}"
fi

if [ $AGENT3_DONE -eq 1 ]; then
    echo -e "Agent 3: ${GREEN}✅ COMPLETED${NC}"
else
    echo -e "Agent 3: ${RED}❌ PENDING${NC}"
fi

echo ""
echo "Progress: $TOTAL_DONE/3 agents completed"
echo ""

if [ $TOTAL_DONE -eq 3 ]; then
    echo -e "${GREEN}🎉 ALL AGENTS COMPLETED - Ready for integration testing!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Run integration tests: see docs/INTEGRATION_TEST_PLAN.md"
    echo "2. Build backend: cd backend && go build cmd/server/main.go"
    echo "3. Build frontend: cd clients/desktop && npm run build"
    echo "4. Run E2E tests with MT5 mode enabled"
else
    echo -e "${YELLOW}⏳ WAITING for $(expr 3 - $TOTAL_DONE) agent(s) to complete${NC}"
    echo ""
    echo "Agent 4 (Integration) is standing by..."
fi

echo "==================================================================="
