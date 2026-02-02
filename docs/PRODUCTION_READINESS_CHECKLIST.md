# Production Readiness Checklist

## Version: 1.0.0 | Date: 2026-01-30

---

## 1. Code Quality & Cleanup

### 1.1 Simulation Code Removal
- [ ] **Search for "SIM" in codebase**
  ```bash
  grep -r "\"SIM\"" backend/ clients/ --include="*.go" --include="*.ts"
  ```
  Expected: Zero results

- [ ] **Verify no simulation LP assignments**
  ```bash
  grep -r "LP.*=.*\"SIM" backend/ --include="*.go"
  ```
  Expected: Zero results

- [ ] **Check main.go for simulation goroutines**
  ```bash
  grep -A 5 "Simulated\|Hybrid" backend/cmd/server/main.go
  ```
  Expected: Commented out or removed

**Pass Criteria**: ✅ Zero active simulation code

---

## 2. YOFX Connection Validation

### 2.1 FIX Sessions
- [ ] **Check YOFX1 status**
  ```bash
  curl -s http://localhost:7999/admin/fix/status | jq -r '.sessions.YOFX1'
  ```
  Expected: `"LOGGED_IN"`

- [ ] **Check YOFX2 status**
  ```bash
  curl -s http://localhost:7999/admin/fix/status | jq -r '.sessions.YOFX2'
  ```
  Expected: `"LOGGED_IN"`

- [ ] **Verify 5-minute stability**
  ```bash
  for i in {1..60}; do
    STATUS=$(curl -s http://localhost:7999/admin/fix/status | jq -r '.sessions.YOFX2')
    [ "$STATUS" != "LOGGED_IN" ] && echo "FAIL at minute $((i/12))" && exit 1
    sleep 5
  done
  echo "PASS: Stable for 5 minutes"
  ```

**Pass Criteria**: ✅ Both sessions stable 5+ minutes

---

## 3. LP Tag Verification

### 3.1 Backend LP Tags
- [ ] **Check all ticks have YOFX LP**
  ```bash
  curl -s http://localhost:7999/api/diagnostics/market-data | \
    jq '.latestTicks | to_entries[] | {symbol: .key, LP: .value.LP}'
  ```
  Expected: All LP values are "YOFX"

- [ ] **Search for SIM tags (should fail)**
  ```bash
  curl -s http://localhost:7999/api/diagnostics/market-data | \
    jq -r '.latestTicks[].LP' | grep -i sim
  ```
  Expected: Exit code 1 (no matches)

**Pass Criteria**: ✅ 100% of ticks have LP="YOFX"

### 3.2 Frontend LP Display
- [ ] Open client at http://localhost:5173
- [ ] Navigate to Market Watch
- [ ] Verify LP column shows "YOFX" for all symbols
- [ ] Verify LP tags are GREEN color (#10b981)
- [ ] Test tooltip on hover shows "YOFX"

**Pass Criteria**: ✅ All symbols display "YOFX" in green

---

## 4. Database Validation

### 4.1 No Simulation Data
- [ ] **Search for SIM entries**
  ```bash
  DB_PATH="data/ticks/ticks_BROKER-001_$(date +%Y%m%d).db"
  SIM_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks WHERE lp LIKE '%SIM%';")
  echo "SIM entries: $SIM_COUNT"
  [ $SIM_COUNT -eq 0 ] && echo "PASS" || echo "FAIL: Found $SIM_COUNT simulation entries"
  ```

- [ ] **Verify YOFX data exists**
  ```bash
  YOFX_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM ticks WHERE lp = 'YOFX';")
  echo "YOFX entries: $YOFX_COUNT"
  [ $YOFX_COUNT -gt 0 ] && echo "PASS" || echo "FAIL: No YOFX data"
  ```

- [ ] **Check LP distribution**
  ```bash
  sqlite3 "$DB_PATH" "SELECT lp, COUNT(*) FROM ticks GROUP BY lp;"
  ```
  Expected: Only "YOFX" rows

**Pass Criteria**: ✅ Zero SIM entries, all YOFX

---

## 5. Price & Market Data Accuracy

### 5.1 Price Validation
- [ ] **Get EURUSD price**
  ```bash
  curl -s http://localhost:7999/api/diagnostics/market-data | \
    jq '.latestTicks.EURUSD | {bid, ask}'
  ```

- [ ] **Compare with Investing.com**
  - Tolerance: Within ±0.0005 (5 pips)

- [ ] **Verify spread is reasonable**
  - EURUSD: 0.00010 - 0.00030 (1-3 pips)
  - XAUUSD: 0.10 - 0.50

- [ ] **Check bid < ask** (no anomalies)

**Pass Criteria**: ✅ Prices accurate within ±0.1%

---

## 6. Performance & Latency

### 6.1 Latency Targets
- [ ] **API response time <100ms**
  ```bash
  time curl -s http://localhost:7999/api/diagnostics/market-data > /dev/null
  ```

- [ ] **Tick rate >50/sec**
  ```bash
  TICKS_START=$(curl -s http://localhost:7999/admin/fix/ticks | jq '.totalTickCount')
  sleep 60
  TICKS_END=$(curl -s http://localhost:7999/admin/fix/ticks | jq '.totalTickCount')
  RATE=$(( (TICKS_END - TICKS_START) / 60 ))
  echo "Tick rate: $RATE/sec"
  [ $RATE -ge 50 ] && echo "PASS" || echo "FAIL"
  ```

**Pass Criteria**: ✅ Latency <100ms, tick rate >50/sec

### 6.2 Resource Usage
- [ ] CPU <50%
- [ ] Memory <500MB
- [ ] Database <100MB/day growth

**Pass Criteria**: ✅ Efficient resource usage

---

## 7. Security

### 7.1 No Hardcoded Secrets
- [ ] **Search for hardcoded passwords**
  ```bash
  grep -r "password.*=.*\"" backend/ --include="*.go" | grep -v ".env"
  ```
  Expected: Zero results

- [ ] **Verify .env in .gitignore**
  ```bash
  grep "\.env" .gitignore
  ```

- [ ] **Check no .env committed to git**
  ```bash
  git ls-files | grep ".env"
  ```
  Expected: Empty

**Pass Criteria**: ✅ No hardcoded secrets

---

## 8. Monitoring & Logging

### 8.1 Health Checks
- [ ] **Health endpoint returns OK**
  ```bash
  curl -s http://localhost:7999/health
  ```
  Expected: `OK`

- [ ] **Error logging functional**
  - Test invalid symbol subscription
  - Check backend.log for error entry

**Pass Criteria**: ✅ Monitoring active, errors logged

---

## 9. Final Sign-Off

### Checklist Summary
- [ ] No simulation code in production
- [ ] FIX sessions stable (5+ minutes)
- [ ] All LP tags are "YOFX"
- [ ] Database has zero SIM entries
- [ ] Latency <100ms (p95)
- [ ] CPU <50%, Memory <500MB
- [ ] No hardcoded secrets
- [ ] Health check active

### Go/No-Go Decision
**Status**: ☐ GO  ☐ NO-GO

### Sign-Off
| Role | Name | Date |
|------|------|------|
| Lead Developer | _____ | _____ |
| QA Engineer | _____ | _____ |
| DevOps | _____ | _____ |

---

## Post-Deployment (Within 1 Hour)
- [ ] Health check 200 OK
- [ ] FIX logged in
- [ ] Ticks flowing
- [ ] No errors in logs

**Next Review**: Every 30 days or before major release
