#!/bin/bash

# Market Data Pipeline Health Check Script
# Purpose: Verify all components of the pipeline are working correctly
# Usage: ./pipeline_health_check.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Market Data Pipeline Health Check${NC}"
echo -e "${BLUE}========================================${NC}"
echo "Time: $(date)"
echo

# Check backend service
echo -e "${BLUE}1. Backend Service Status:${NC}"
if pgrep -f "go run.*server" > /dev/null; then
  echo -e "${GREEN}   ✓ Backend running${NC}"
  BACKEND_OK=1
else
  echo -e "${RED}   ✗ Backend NOT running${NC}"
  echo -e "${YELLOW}   Start with: cd backend/cmd/server && go run main.go${NC}"
  BACKEND_OK=0
fi

# Check Redis
echo -e "${BLUE}2. Redis Status:${NC}"
if redis-cli ping > /dev/null 2>&1; then
  REDIS_MEMORY=$(redis-cli INFO memory 2>/dev/null | grep used_memory_human | cut -d: -f2)
  REDIS_KEYS=$(redis-cli DBSIZE 2>/dev/null | grep keys | awk '{print $1}')
  echo -e "${GREEN}   ✓ Redis online${NC}"
  echo "   Memory: $REDIS_MEMORY"
  echo "   Keys: $REDIS_KEYS"
  REDIS_OK=1
else
  echo -e "${RED}   ✗ Redis NOT running${NC}"
  echo -e "${YELLOW}   Start with: redis-server${NC}"
  REDIS_OK=0
fi

# Check FIX connection
if [ $BACKEND_OK -eq 1 ]; then
  echo -e "${BLUE}3. FIX Gateway Status:${NC}"
  FIX_STATUS=$(curl -s http://localhost:8080/api/admin/fix-status 2>/dev/null | jq -r '.YOFX2' 2>/dev/null || echo "unknown")

  if [ "$FIX_STATUS" = "connected" ]; then
    echo -e "${GREEN}   ✓ FIX connected (YOFX2)${NC}"
    FIX_OK=1
  elif [ "$FIX_STATUS" = "connecting" ]; then
    echo -e "${YELLOW}   ⏳ FIX connecting... (YOFX2)${NC}"
    FIX_OK=1
  else
    echo -e "${RED}   ✗ FIX disconnected or unknown (Status: $FIX_STATUS)${NC}"
    FIX_OK=0
  fi

  # Check pipeline stats
  echo -e "${BLUE}4. Pipeline Statistics:${NC}"
  STATS=$(curl -s http://localhost:8080/api/admin/pipeline-stats 2>/dev/null)

  if [ -n "$STATS" ]; then
    TICKS=$(echo "$STATS" | jq -r '.data.ticks_received // 0' 2>/dev/null)
    PROCESSED=$(echo "$STATS" | jq -r '.data.ticks_processed // 0' 2>/dev/null)
    LATENCY=$(echo "$STATS" | jq -r '.data.avg_latency_ms // 0' 2>/dev/null)
    DROPPED=$(echo "$STATS" | jq -r '.data.ticks_dropped // 0' 2>/dev/null)
    CLIENTS=$(echo "$STATS" | jq -r '.data.clients_connected // 0' 2>/dev/null)

    echo "   Ticks received: $TICKS"
    echo "   Ticks processed: $PROCESSED"
    echo "   Avg latency: ${LATENCY}ms"
    echo "   Dropped: $DROPPED"
    echo "   Connected clients: $CLIENTS"

    # Latency check
    if (( $(echo "$LATENCY < 10" | bc -l 2>/dev/null || echo "0") )); then
      echo -e "${GREEN}   ✓ Latency acceptable (< 10ms)${NC}"
      LATENCY_OK=1
    elif (( $(echo "$LATENCY < 20" | bc -l 2>/dev/null || echo "0") )); then
      echo -e "${YELLOW}   ⚠ Latency acceptable but elevated (${LATENCY}ms)${NC}"
      LATENCY_OK=1
    else
      echo -e "${RED}   ✗ Latency HIGH (${LATENCY}ms > 20ms)${NC}"
      LATENCY_OK=0
    fi

    # Dropped check
    if [ "$DROPPED" -eq 0 ]; then
      echo -e "${GREEN}   ✓ No dropped ticks${NC}"
      DROPPED_OK=1
    else
      echo -e "${YELLOW}   ⚠ Dropped ticks detected: $DROPPED${NC}"
      DROPPED_OK=0
    fi
  else
    echo -e "${RED}   ✗ Cannot reach pipeline stats${NC}"
    LATENCY_OK=0
    DROPPED_OK=0
  fi

  # Check Redis data volume
  echo -e "${BLUE}5. Market Data Storage:${NC}"
  EURUSD_COUNT=$(redis-cli LLEN market_data:EURUSD 2>/dev/null || echo "0")
  GBPUSD_COUNT=$(redis-cli LLEN market_data:GBPUSD 2>/dev/null || echo "0")

  if [ "$EURUSD_COUNT" -gt 0 ] || [ "$GBPUSD_COUNT" -gt 0 ]; then
    echo -e "${GREEN}   ✓ Market data in Redis${NC}"
    echo "   EURUSD ticks: $EURUSD_COUNT"
    echo "   GBPUSD ticks: $GBPUSD_COUNT"
    STORAGE_OK=1
  else
    echo -e "${YELLOW}   ⚠ No ticks in Redis (pipeline may just have started)${NC}"
    STORAGE_OK=1  # Not a critical failure
  fi

  # Check WebSocket connections
  echo -e "${BLUE}6. WebSocket Status:${NC}"
  if [ "$CLIENTS" -gt 0 ]; then
    echo -e "${GREEN}   ✓ WebSocket clients connected: $CLIENTS${NC}"
    WS_OK=1
  else
    echo -e "${YELLOW}   ⚠ No WebSocket clients connected${NC}"
    WS_OK=1  # Not critical if no clients
  fi

else
  echo -e "${YELLOW}3. FIX Gateway Status: SKIPPED (backend not running)${NC}"
  echo -e "${YELLOW}4. Pipeline Statistics: SKIPPED (backend not running)${NC}"
  echo -e "${YELLOW}5. Market Data Storage: SKIPPED (backend not running)${NC}"
  echo -e "${YELLOW}6. WebSocket Status: SKIPPED (backend not running)${NC}"
  FIX_OK=0
  LATENCY_OK=0
  DROPPED_OK=0
  STORAGE_OK=0
  WS_OK=0
fi

# FIX message monitoring (if available)
if [ -f "backend/fixstore/YOFX2.msgs" ]; then
  echo -e "${BLUE}7. FIX Message Log:${NC}"
  FIX_LINES=$(wc -l < backend/fixstore/YOFX2.msgs 2>/dev/null || echo "0")
  echo "   Total FIX messages: $FIX_LINES"

  RECENT_QUOTES=$(grep -c "35=D" backend/fixstore/YOFX2.msgs 2>/dev/null || echo "0")
  if [ "$RECENT_QUOTES" -gt 0 ]; then
    echo -e "${GREEN}   ✓ Market data quotes received: $RECENT_QUOTES${NC}"
    QUOTES_OK=1
  else
    echo -e "${YELLOW}   ⚠ No market data quotes in log${NC}"
    QUOTES_OK=0
  fi
fi

# Summary
echo
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Health Check Summary${NC}"
echo -e "${BLUE}========================================${NC}"

OVERALL_OK=1

if [ $BACKEND_OK -eq 1 ]; then
  echo -e "${GREEN}✓ Backend Service${NC}"
else
  echo -e "${RED}✗ Backend Service${NC}"
  OVERALL_OK=0
fi

if [ $REDIS_OK -eq 1 ]; then
  echo -e "${GREEN}✓ Redis${NC}"
else
  echo -e "${RED}✗ Redis${NC}"
  OVERALL_OK=0
fi

if [ $FIX_OK -eq 1 ]; then
  echo -e "${GREEN}✓ FIX Gateway${NC}"
else
  echo -e "${YELLOW}⚠ FIX Gateway (may still be connecting)${NC}"
fi

if [ $LATENCY_OK -eq 1 ]; then
  echo -e "${GREEN}✓ Pipeline Latency${NC}"
else
  echo -e "${RED}✗ Pipeline Latency${NC}"
  OVERALL_OK=0
fi

if [ $DROPPED_OK -eq 1 ]; then
  echo -e "${GREEN}✓ No Dropped Ticks${NC}"
else
  echo -e "${YELLOW}⚠ Dropped Ticks Detected${NC}"
fi

if [ $STORAGE_OK -eq 1 ]; then
  echo -e "${GREEN}✓ Data Storage${NC}"
else
  echo -e "${YELLOW}⚠ Data Storage${NC}"
fi

echo
if [ $OVERALL_OK -eq 1 ]; then
  echo -e "${GREEN}OVERALL STATUS: HEALTHY${NC}"
  exit 0
else
  echo -e "${RED}OVERALL STATUS: ISSUES DETECTED${NC}"
  exit 1
fi
