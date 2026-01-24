# Market Data Pipeline Testing - START HERE

## Welcome!

This guide will help you test the market data pipeline. All documentation, scripts, and automated tests are ready to use.

---

## 🚀 Quick Start (Choose Your Path)

### Path 1: I have 5 minutes (Quick Health Check)
```bash
./scripts/pipeline_health_check.sh    # Linux/Mac
# or
PowerShell -ExecutionPolicy Bypass -File scripts/verify_pipeline.ps1  # Windows
```
**Result:** "OVERALL STATUS: HEALTHY" = everything works

---

### Path 2: I have 30 minutes (Automated Testing)
```bash
# Start backend and Redis (if not running)
cd backend/cmd/server && go run main.go &
redis-server &

# Run automated E2E test
cd backend/cmd/test_e2e
go run main.go -duration 30s
```
**Result:** "✓ TEST PASSED" = pipeline verified

---

### Path 3: I have 60+ minutes (Comprehensive Testing)
1. Read: `docs/TEST_PLAN_SUMMARY.md` (executive overview)
2. Follow: `docs/MANUAL_VERIFICATION_CHECKLIST.md` (step-by-step)
3. Run: Health check + E2E test
4. Test: Frontend manually
5. Document: Use test report template

**Result:** Comprehensive test report with sign-off

---

## 📋 Documentation Index

| Document | Purpose | Read Time | Best For |
|----------|---------|-----------|----------|
| **QUICK_TEST_REFERENCE.txt** | One-page reference card | 5 min | Quick lookup, bookmark this |
| **TEST_PLAN_SUMMARY.md** | Executive overview | 10 min | Understanding the big picture |
| **MANUAL_VERIFICATION_CHECKLIST.md** | Step-by-step testing | 50 min | Hands-on testing |
| **E2E_TEST_PLAN.md** | Complete reference | 2 hours | In-depth understanding |
| **TEST_PLAN_DELIVERABLES.md** | What was delivered | 10 min | Understanding scope |

### Navigation Tips
- **Confused?** Read TEST_PLAN_SUMMARY.md first
- **Need quick reference?** Open QUICK_TEST_REFERENCE.txt
- **Want detailed steps?** Use MANUAL_VERIFICATION_CHECKLIST.md
- **Need deep dive?** See E2E_TEST_PLAN.md
- **Something failing?** Jump to E2E_TEST_PLAN.md Part 5 (Troubleshooting)

---

## 🎯 Quick Test Commands

### Health Check (everything OK?)
```bash
./scripts/pipeline_health_check.sh
```

### Component Tests (each part working?)
```bash
# FIX Gateway
tail -f backend/fixstore/YOFX2.msgs | grep "35=D" | head -5

# Pipeline
curl http://localhost:8080/api/admin/pipeline-stats | jq '.data | {ticks_received, avg_latency_ms}'

# Redis
redis-cli LLEN market_data:EURUSD

# WebSocket
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')
wscat -c "ws://localhost:8080/ws?token=$TOKEN"
```

### Automated E2E Test
```bash
cd backend/cmd/test_e2e && go run main.go -duration 30s
```

### Frontend Test
Open http://localhost:3000 and login with trader/trader

---

## ✅ Success Checklist

Testing is successful if ALL of these are true:

- [ ] **FIX Connected:** See heartbeat messages every 30 seconds
- [ ] **Data Flowing:** 100+ ticks per second through pipeline
- [ ] **Latency Good:** < 10ms average latency
- [ ] **No Drops:** 0 dropped ticks under normal load
- [ ] **WebSocket OK:** Clients connect and receive data
- [ ] **Redis Storing:** Ticks persisted and queryable
- [ ] **Frontend Live:** Market Watch shows live prices
- [ ] **Charts Display:** OHLC candles render correctly

---

## 🔧 Setup (Do This First)

### 1. Start Backend
```bash
cd backend/cmd/server
go run main.go
```

### 2. Start Redis
```bash
redis-server
```

### 3. Verify Connection
```bash
curl http://localhost:8080/api/admin/pipeline-stats
```

---

## 🐛 Common Issues

| Problem | Solution |
|---------|----------|
| No market data | Check FIX connection: grep "35=A" backend/fixstore/YOFX2.msgs |
| High latency | Check CPU: `top` command |
| WebSocket fails | Re-login: Clear localStorage in browser |
| Test won't run | Make sure backend and Redis are running |

**For more issues:** See E2E_TEST_PLAN.md Part 5 (Troubleshooting)

---

## 📊 Test Results Summary

After testing, you should have:

✓ **Metrics:**
- Latency: < 10ms average
- Throughput: 100-1000 ticks/sec
- Drop rate: 0% (normal) to <1% (under load)

✓ **Data Quality:**
- Bid < Ask (always)
- OHLC math correct (Low ≤ Open,Close ≤ High)
- Timestamps recent (< 1 minute old)

✓ **System Health:**
- Memory stable (no growth over time)
- CPU reasonable (< 50% normal load)
- All components connected and responding

---

## 📁 Files You'll Need

```
docs/
├── QUICK_TEST_REFERENCE.txt          ← Bookmark this
├── TEST_PLAN_SUMMARY.md              ← Start here
├── MANUAL_VERIFICATION_CHECKLIST.md  ← Follow this
└── E2E_TEST_PLAN.md                  ← Reference this

scripts/
├── pipeline_health_check.sh           ← Run this first
└── verify_pipeline.ps1                ← Windows version

backend/cmd/test_e2e/
└── main.go                            ← Automated test
```

---

## 🎓 Learning Path

**Beginner:**
1. Read: QUICK_TEST_REFERENCE.txt (5 min)
2. Run: Health check script (5 min)
3. Read: TEST_PLAN_SUMMARY.md (10 min)

**Intermediate:**
1. Follow: MANUAL_VERIFICATION_CHECKLIST.md (50 min)
2. Run: Automated E2E test (10 min)
3. Test: Frontend manually (10 min)

**Advanced:**
1. Read: E2E_TEST_PLAN.md completely (2 hours)
2. Customize: Tests for your environment
3. Document: Create your own test reports

---

## 🚦 Test Status

| Component | Status | How to Test |
|-----------|--------|-------------|
| FIX Gateway | Ready | FIX-001, FIX-002 |
| Pipeline | Ready | PIPE-001, PIPE-002, PIPE-003 |
| Redis | Ready | REDIS-001, REDIS-002 |
| WebSocket | Ready | WS-001, WS-002, WS-003 |
| Integration | Ready | INT-001, INT-002, INT-003 |
| Frontend | Ready | FRONT-001, FRONT-002, FRONT-003 |

**Total:** 48 test scenarios ready to execute

---

## ⏱️ Time Estimates

| Activity | Time |
|----------|------|
| Health check | 5 min |
| Quick component test | 15 min |
| Automated E2E test | 10 min |
| Manual testing (full) | 50 min |
| Frontend testing | 15 min |
| Documentation | 5 min |
| **Total** | **70 min** |

---

## 📞 Need Help?

**Question:** What do I do first?
**Answer:** Run `./scripts/pipeline_health_check.sh`

**Question:** Tests are failing, what now?
**Answer:**
1. Check E2E_TEST_PLAN.md Part 5 (Troubleshooting)
2. Run diagnostic commands (see QUICK_TEST_REFERENCE.txt)
3. Check backend logs: `tail -f backend/fixstore/YOFX2.msgs`

**Question:** How do I test just one component?
**Answer:** See Test Coverage Matrix in TEST_PLAN_SUMMARY.md

**Question:** Tests pass but I want more details?
**Answer:** Read E2E_TEST_PLAN.md for comprehensive reference

---

## 🎯 Next Steps

1. **Read This File** (5 min) - You are here
2. **Run Health Check** (5 min)
   ```bash
   ./scripts/pipeline_health_check.sh
   ```
3. **Read Overview** (10 min)
   ```bash
   cat docs/TEST_PLAN_SUMMARY.md
   ```
4. **Choose Path** (5 min)
   - Quick? Run E2E test
   - Thorough? Follow MANUAL_VERIFICATION_CHECKLIST.md
5. **Execute Tests** (30-60 min)
   ```bash
   cd backend/cmd/test_e2e && go run main.go
   ```
6. **Document Results** (5 min)
   Use test report template in docs/

---

## 📝 Quick Reference

| What I Want To Do | Command | Doc Section |
|------------------|---------|-------------|
| Quick health check | `./scripts/pipeline_health_check.sh` | N/A |
| Automated E2E test | `cd backend/cmd/test_e2e && go run main.go` | N/A |
| Check FIX status | `curl http://localhost:8080/api/admin/fix-status` | QUICK_TEST_REFERENCE.txt |
| Check pipeline stats | `curl http://localhost:8080/api/admin/pipeline-stats` | QUICK_TEST_REFERENCE.txt |
| Test WebSocket | `wscat -c "ws://localhost:8080/ws?token=$TOKEN"` | MANUAL_VERIFICATION_CHECKLIST.md |
| View Redis data | `redis-cli LRANGE market_data:EURUSD 0 5` | QUICK_TEST_REFERENCE.txt |
| Troubleshoot issue | See E2E_TEST_PLAN.md Part 5 | E2E_TEST_PLAN.md |

---

## ✨ You're Ready!

Everything is prepared:
- Comprehensive documentation (7,400+ lines)
- Automated health check scripts
- E2E test program
- Manual verification checklists
- Troubleshooting guides

**Start with:** Run `./scripts/pipeline_health_check.sh`

**Questions?** Check the documentation above

**Let's go!**

---

**Last Updated:** 2025-01-20
**Version:** 1.0
**Status:** Ready for Testing
