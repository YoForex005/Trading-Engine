# Market Data Pipeline E2E Test Plan - Executive Summary

## Overview

This comprehensive test plan enables end-to-end verification of the market data pipeline spanning from FIX gateway through WebSocket broadcasting to frontend display.

### Pipeline Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    MARKET DATA PIPELINE                         │
└─────────────────────────────────────────────────────────────────┘

  FIX Gateway (YOFX LP)
         ↓
  ┌─────────────┐
  │ FIX Session │  ← Test: Connection, Authentication, Quote Reception
  └─────────────┘
         ↓
  ┌──────────────────────────┐
  │ Raw Tick Ingestion       │  ← Test: Buffer capacity, Drop rate
  │ & Normalization          │
  └──────────────────────────┘
         ↓
  ┌──────────────────────────┐
  │ Validation & Dedup       │  ← Test: Data quality, Duplicate detection
  └──────────────────────────┘
         ↓
  ┌──────────────────────────┐
  │ OHLC Generation          │  ← Test: Bar generation, Time alignment
  └──────────────────────────┘
         ↓
  ┌──────────────────────────┐
  │ Redis Storage            │  ← Test: Persistence, Query speed
  │ (Hot data layer)         │
  └──────────────────────────┘
         ↓
  ┌──────────────────────────┐
  │ Quote Distributor        │  ← Test: Throughput, Throttling
  │ (Redis Pub/Sub)          │
  └──────────────────────────┘
         ↓
  ┌──────────────────────────┐
  │ WebSocket Hub            │  ← Test: Client connections, Broadcasting
  └──────────────────────────┘
         ↓
  ┌──────────────────────────┐
  │ Frontend Clients         │  ← Test: Display accuracy, Latency
  │ (React/TypeScript)       │
  └──────────────────────────┘
```

---

## Test Categories & Scope

### Category 1: Component Testing (Unit + Integration)
Tests individual pipeline components in isolation.

**Estimated Duration:** 30 minutes

| Component | Test ID | Key Tests | Pass Criteria |
|-----------|---------|-----------|---------------|
| **FIX Gateway** | FIX-001 to FIX-002 | Connection, quote reception, protocol validation | <10sec connection, 100+ ticks/sec, valid FIX format |
| **Data Ingester** | PIPE-001 | Tick normalization, validation, deduplication | 0 drops under normal load, <5ms latency |
| **OHLC Engine** | PIPE-002 | Bar generation, time boundaries, OHLC math | Valid OHLC relationships, bars close on boundaries |
| **Quality Validation** | PIPE-003 | Duplicate detection, outlier detection, range checks | >95% duplicate detection, price sanity checks |
| **Redis Storage** | REDIS-001 to REDIS-002 | Persistence, TTL, query performance | All ticks stored, queryable in <100ms |
| **WebSocket Hub** | WS-001 to WS-003 | Client management, broadcasting, throttling | <1% message loss, 60%+ CPU reduction via throttling |

### Category 2: Integration Testing
Tests data flow across multiple components.

**Estimated Duration:** 20 minutes

| Scenario | Test ID | Coverage | Success Metrics |
|----------|---------|----------|-----------------|
| **FIX → Pipeline → Redis** | INT-001 | Full ingestion pipeline | 100% of FIX ticks reach Redis in <20ms |
| **Pipeline → Redis → WebSocket** | INT-002 | Full distribution pipeline | All Redis ticks broadcast to connected clients |
| **Sustained Load** | INT-003 | Performance under stress | 1000 ticks/sec, <1% drop, stable memory |

### Category 3: Frontend Testing
Tests display and user-facing functionality.

**Estimated Duration:** 15 minutes

| Feature | Test ID | Verification | Criteria |
|---------|---------|--------------|----------|
| **Market Watch** | FRONT-001 | Live price display, updates, accuracy | All symbols visible, updates <1sec, correct bid/ask |
| **Charts** | FRONT-002 | OHLC candles, timeframe switching | Valid candles, correct time boundaries, smooth rendering |
| **Real-Time Updates** | FRONT-003 | Latency, responsiveness | <1 second visible latency, no stale data |

### Category 4: System-Level Testing
Tests overall system performance and reliability.

**Estimated Duration:** 25 minutes

| Aspect | Test Duration | Load | Pass Criteria |
|--------|----------------|------|---------------|
| **Throughput** | 10 min | 1000 ticks/sec continuous | 0% drop, <15ms latency |
| **Stability** | 10 min | 10 concurrent WebSocket clients | All clients remain connected, no memory leaks |
| **Failover** | 5 min | Simulate LP disconnect, reconnect | Graceful disconnect, automatic reconnect |

---

## Test Execution Roadmap

### Phase 1: Pre-Test Setup (5 minutes)
```bash
# Start services
cd backend/cmd/server && go run main.go &
redis-server &

# Verify connectivity
curl http://localhost:8080/api/admin/pipeline-stats
redis-cli PING
```

### Phase 2: Quick Health Check (5 minutes)
```bash
# Run the health check script
./scripts/pipeline_health_check.sh
# or on Windows:
PowerShell -ExecutionPolicy Bypass -File scripts/verify_pipeline.ps1
```

**Expected Output:**
```
✓ Backend Service
✓ Redis
✓ FIX Gateway
✓ Pipeline Latency
✓ No Dropped Ticks
✓ Data Storage
OVERALL STATUS: HEALTHY
```

### Phase 3: Component Testing (30 minutes)

#### A. FIX Gateway (5 min)
```bash
# Manual test: Verify quote flow
tail -f backend/fixstore/YOFX2.msgs | grep "35=D"

# Expected: 10-20 quotes per second
```

#### B. Pipeline Ingestion (10 min)
```bash
# Automated test
cd backend/cmd/test_e2e
go run main.go -backend http://localhost:8080 -duration 30s
```

**Expected Output:**
```
Total ticks received: 850
Valid ticks: 845
Invalid ticks: 5
Unique symbols: 15
Avg throughput: 283 ticks/sec
✓ TEST PASSED
```

#### C. Redis Storage (5 min)
```bash
# Manual verification
redis-cli KEYS 'market_data:*' | wc -l   # Should be > 10
redis-cli LLEN market_data:EURUSD        # Should be > 500
```

#### D. WebSocket Distribution (10 min)
```bash
# Connect client and monitor
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

wscat -c "ws://localhost:8080/ws?token=$TOKEN"

# Expected: Continuous stream of market ticks
```

### Phase 4: Integration Testing (20 minutes)

#### A. Data Flow Verification (10 min)
```bash
# Verify FIX → Pipeline → Redis
BEFORE=$(redis-cli LLEN market_data:EURUSD)
sleep 30
AFTER=$(redis-cli LLEN market_data:EURUSD)
echo "New ticks in Redis: $((AFTER - BEFORE))"  # Should be > 100
```

#### B. Load Testing (10 min)
```bash
# Run E2E test with sustained load
cd backend/cmd/load_test
go run main.go -ticks-per-sec 1000 -clients 10 -duration 60s
```

### Phase 5: Frontend Testing (15 minutes)

1. Open http://localhost:3000
2. Login with trader/trader
3. Navigate to Market Watch → Verify live prices
4. Navigate to Charts → Verify OHLC candles
5. Observe for 5 minutes → Check data freshness

---

## Key Test Commands

### Quick Health Check
```bash
# Linux/Mac
./scripts/pipeline_health_check.sh

# Windows PowerShell
PowerShell -ExecutionPolicy Bypass -File scripts/verify_pipeline.ps1
```

### Manual Verification
```bash
# FIX status
curl http://localhost:8080/api/admin/fix-status | jq .YOFX2

# Pipeline stats
curl http://localhost:8080/api/admin/pipeline-stats | jq '.data'

# WebSocket stats
curl http://localhost:8080/api/admin/hub-stats | jq .

# Redis data
redis-cli LRANGE market_data:EURUSD 0 2 | jq .
redis-cli LLEN market_data:EURUSD
```

### Automated Testing
```bash
# E2E test
cd backend/cmd/test_e2e && go run main.go -duration 30s

# Load test
cd backend && go test -v ./datapipeline -run TestLoad -timeout 60s
```

---

## Success Criteria Checklist

### Minimum Requirements (PASS if all green)
- [ ] **FIX Connection:** Connected within 10 seconds, heartbeat every 30 seconds
- [ ] **Data Flow:** 100+ ticks per second flowing through pipeline
- [ ] **Latency:** Average latency < 10ms per tick
- [ ] **Storage:** All ticks stored in Redis with correct format
- [ ] **Broadcasting:** All data reaches WebSocket clients without loss
- [ ] **Frontend:** Prices display correctly with <1 second latency

### Performance Targets
- [ ] **Throughput:** 1000+ ticks/sec sustained
- [ ] **Latency:** <5ms average (p95 <10ms)
- [ ] **CPU:** <50% under normal load
- [ ] **Memory:** Stable, no growth over time
- [ ] **Client Connections:** Support 10+ concurrent WebSocket clients

### Quality Metrics
- [ ] **Data Quality:** >99% valid ticks (OHLC math, bid<ask)
- [ ] **Uptime:** 99.9% (no unexpected disconnects)
- [ ] **Duplicate Detection:** 100% of exact duplicates caught
- [ ] **Message Delivery:** 100% to connected clients (except throttled)

---

## Troubleshooting Quick Reference

| Issue | Diagnosis | Solution |
|-------|-----------|----------|
| No market data | Check FIX status | Verify LP credentials, proxy connection |
| High latency (>20ms) | Check CPU, Redis | Increase buffer sizes, optimize queries |
| WebSocket disconnects | Check auth | Re-login, verify token expiration |
| Dropped ticks | Check buffer size | Increase `TickBufferSize` in config |
| Memory growing | Check for leaks | Restart backend, monitor with `htop` |

---

## Test Report Template

**File:** `test_report_YYYY-MM-DD.md`

```markdown
# Pipeline Test Report
Date: [Date]
Tester: [Name]

## Environment
- Backend: [Version]
- Redis: [Version]
- LP: [YOFX/Other]
- Test Duration: [X minutes]

## Results

| Component | Status | Latency | Throughput | Notes |
|-----------|--------|---------|------------|-------|
| FIX | PASS/FAIL | XXms | XXX/sec | |
| Pipeline | PASS/FAIL | XXms | XXX/sec | |
| WebSocket | PASS/FAIL | XXms | XXX/sec | |
| Frontend | PASS/FAIL | XXms | - | |

## Metrics
- Avg Latency: XXms
- Peak Latency: XXms
- Memory Usage: XXMb
- CPU Usage: XX%
- Dropped Ticks: X

## Issues Found
1. [Issue description]
2. [Issue description]

## Sign-Off
- Overall Result: PASS / FAIL
- Approved By: _______________
- Date: _______________
```

---

## Test Artifacts

All test artifacts are organized in the codebase:

```
D:\Trading-Engine\
├── docs/
│   ├── E2E_TEST_PLAN.md                    # Detailed test plan (this file)
│   ├── MANUAL_VERIFICATION_CHECKLIST.md    # Step-by-step checklist
│   └── TEST_PLAN_SUMMARY.md               # Executive summary (this file)
├── scripts/
│   ├── pipeline_health_check.sh            # Linux/Mac health check
│   └── verify_pipeline.ps1                 # Windows PowerShell check
├── backend/cmd/
│   ├── test_e2e/main.go                   # Automated E2E test
│   ├── load_test/main.go                  # Load testing tool
│   └── server/main.go                     # Backend server
└── test_reports/                          # Generated test reports
    └── report_YYYY-MM-DD.md
```

---

## Key Files for Reference

| File | Purpose |
|------|---------|
| `backend/datapipeline/pipeline.go` | Main pipeline orchestrator |
| `backend/fix/gateway.go` | FIX protocol implementation |
| `backend/ws/hub.go` | WebSocket distribution hub |
| `backend/api/server.go` | REST API endpoints |
| `clients/desktop/src/services/websocket.ts` | Frontend WebSocket client |

---

## Getting Started

### 1. Read This First
- E2E_TEST_PLAN.md (comprehensive reference)
- MANUAL_VERIFICATION_CHECKLIST.md (quick checks)

### 2. Run Quick Health Check
```bash
./scripts/pipeline_health_check.sh    # Linux/Mac
PowerShell scripts/verify_pipeline.ps1 # Windows
```

### 3. Run Manual Verification
Follow MANUAL_VERIFICATION_CHECKLIST.md step-by-step

### 4. Run Automated Tests
```bash
cd backend/cmd/test_e2e
go run main.go
```

### 5. Run Load Test (optional)
```bash
cd backend/cmd/load_test
go run main.go -ticks-per-sec 1000 -duration 60s
```

### 6. Test Frontend
Open http://localhost:3000 and manually verify display

---

## Support & Questions

If tests fail:
1. Check the Troubleshooting section in E2E_TEST_PLAN.md
2. Review backend logs: `tail -f backend.log`
3. Check FIX logs: `tail -f backend/fixstore/YOFX2.msgs`
4. Verify Redis: `redis-cli PING`
5. Check network connectivity to LP

---

**Document Version:** 1.0
**Last Updated:** 2025-01-20
**Status:** Ready for Testing
