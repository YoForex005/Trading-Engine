# Regression Test Plan

## Version: 1.0.0 | Date: 2026-01-30
## Purpose: Ensure simulation removal does not break features

---

## 1. Test Priority

| Priority | Count | Pass Required |
|----------|-------|---------------|
| P0 (Critical) | 15 | 100% |
| P1 (Major) | 5 | 90% |
| P2 (Minor) | 3 | 80% |

---

## 2. P0 - Critical Tests

### REG-P0-001: Real-Time Tick Reception
**Objective**: Market data flows from YOFX to UI

**Steps**:
1. Start backend
2. Open Market Watch
3. Observe ticks for EURUSD

**Expected**:
- Ticks update every 1-2 seconds
- LP column shows "YOFX"
- Bid/Ask values realistic

**Risk**: HIGH

---

### REG-P0-002: Market Order Execution
**Objective**: Orders execute at real prices

**Steps**:
1. Place BUY 0.01 EURUSD order
2. Verify execution

**Expected**:
- Order fills immediately
- Execution price matches real market
- Position created

**Risk**: HIGH

---

### REG-P0-003: Database Persistence
**Objective**: Ticks stored in database

```bash
sqlite3 data/ticks/ticks_BROKER-001_$(date +%Y%m%d).db \
  "SELECT COUNT(*) FROM ticks WHERE lp='YOFX';"
```

**Expected**: >100 ticks with lp=YOFX

**Risk**: HIGH

---

### REG-P0-004: No Simulation Data
**Objective**: Zero SIM entries in database

```bash
SIM_COUNT=$(sqlite3 data/ticks/*.db "SELECT COUNT(*) FROM ticks WHERE lp LIKE '%SIM%';")
[ "$SIM_COUNT" -eq 0 ] && echo "PASS" || echo "FAIL"
```

**Expected**: 0 SIM entries

**Risk**: CRITICAL

---

## 3. P1 - Major Tests

### REG-P1-001: Chart Updates
**Objective**: Charts render with real data

**Expected**: TradingView charts show real YOFX prices

---

### REG-P1-002: WebSocket Account Updates
**Objective**: Balance updates via WebSocket

**Expected**: Account updates received within 1 second

---

## 4. Smoke Test (10 minutes)

```bash
#!/bin/bash
# smoke_test.sh

# Health check
curl http://localhost:7999/health || exit 1

# FIX connected
STATUS=$(curl -s http://localhost:7999/admin/fix/status | jq -r '.sessions.YOFX2')
[ "$STATUS" = "LOGGED_IN" ] || exit 1

# No SIM data
DB="data/ticks/ticks_BROKER-001_$(date +%Y%m%d).db"
SIM=$(sqlite3 "$DB" "SELECT COUNT(*) FROM ticks WHERE lp LIKE '%SIM%';")
[ "$SIM" -eq 0 ] || exit 1

echo "SMOKE TESTS PASSED"
```

---

## 5. Pass/Fail Criteria

### Production Go/No-Go:
- P0: 100% pass (15/15) REQUIRED
- P1: 90% pass (4/5) REQUIRED
- P2: 80% pass (2/3) ACCEPTABLE

### Blockers:
- Any P0 failure blocks deployment
- 2+ P1 failures require investigation

---

## 6. Test Matrix

| Area | P0 | P1 | P2 | Status |
|------|----|----|----| ------|
| Market Data | 3 | 0 | 0 | ☐ |
| Orders | 3 | 0 | 0 | ☐ |
| Charts | 2 | 1 | 0 | ☐ |
| WebSocket | 2 | 1 | 0 | ☐ |
| Admin | 2 | 1 | 0 | ☐ |
| Historical | 1 | 0 | 0 | ☐ |
| Alerts | 1 | 1 | 0 | ☐ |
| UI | 1 | 1 | 3 | ☐ |

---

## 7. Sign-Off

### Test Results
- P0 Pass: ___/15 (100% required)
- P1 Pass: ___/5 (90% required)
- P2 Pass: ___/3 (80% required)

### Approval
- [ ] QA Lead: ________
- [ ] Engineering: ________
- [ ] Product: ________

**Status**: ☐ APPROVED  ☐ BLOCKED
