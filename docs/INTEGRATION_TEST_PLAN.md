# Integration Test Plan - Trading Engine Real Data Flow

## Document Control
- **Version**: 1.0.0
- **Date**: 2026-01-30
- **Owner**: QA Team

## Overview
Comprehensive integration testing for YOFX FIX Gateway → Hub → WebSocket → UI data flow.

## Test Cases

### INT-001: Complete Data Pipeline
**Objective**: Verify end-to-end data flow without loss

**Steps**:
1. Verify FIX sessions logged in
2. Subscribe to EURUSD
3. Monitor WebSocket messages
4. Check database persistence
5. Verify UI updates

**Expected**: Data flows seamlessly YOFX → FIX → Hub → DB + WS → UI

**Acceptance**:
- FIX sessions stable 5+ minutes
- Ticks received <1 second latency
- LP tag always "YOFX" (never "SIM")
- UI updates real-time (<500ms)

### INT-002: Reconnection Scenarios
**Objective**: Test auto-recovery from disconnections

**Expected**: System recovers within 30 seconds, restores all subscriptions

### INT-003: Performance Under Load
**Objective**: Handle 100+ ticks/second

**Expected**: 100-500 ticks/sec, <50% CPU, <500MB RAM, <100ms latency

## Quick Test Commands

```bash
# Check FIX status
curl http://localhost:7999/admin/fix/status | jq '.sessions'

# Verify tick flow
curl http://localhost:7999/admin/fix/ticks | jq '{ticks: .totalTickCount}'

# Check for SIM entries (should be 0)
sqlite3 data/ticks/ticks_BROKER-001_$(date +%Y%m%d).db \
  "SELECT COUNT(*) FROM ticks WHERE lp LIKE '%SIM%';"
```

## Success Criteria
- ✅ E2E latency <100ms (p95)
- ✅ Tick throughput >100/sec
- ✅ 50+ concurrent WebSocket clients
- ✅ Zero simulation data
- ✅ Auto-recovery <30 seconds
