# Performance Testing Quick Reference

## 🚀 Quick Commands

```bash
# Run all tests (excluding 24h soak test)
cd backend/scripts/perf && ./run-load-tests.sh

# Run individual test
k6 run tests/performance/load-test.js
k6 run tests/performance/stress-test.js
k6 run tests/performance/spike-test.js

# Run Go benchmarks
cd tests/performance && go test -bench=. -benchmem

# Analyze results
./scripts/perf/analyze-results.sh
```

## 📊 Performance Targets at a Glance

| Metric | Target | Critical Threshold |
|--------|--------|-------------------|
| API p95 | < 200ms | < 500ms |
| API p99 | < 500ms | < 1000ms |
| Order Execution p95 | < 50ms | < 100ms |
| WebSocket Latency p95 | < 10ms | < 20ms |
| Error Rate | < 1% | < 5% |
| Order Success Rate | > 99% | > 95% |
| Concurrent Users | > 5000 | > 3000 |
| Breaking Point | > 7000 | > 5000 |
| Spike Recovery | > 95% | > 90% |

## 🧪 Test Matrix

| Test | Duration | Users | Purpose |
|------|----------|-------|---------|
| Load | 25 min | 1000 | Realistic load behavior |
| Stress | 19 min | 10000 | Find breaking point |
| Spike | 20 min | 10000 | Sudden load spikes |
| Soak | 24 hours | 500 | Detect memory leaks |
| Benchmarks | 10 min/test | N/A | Critical path performance |

## 🎯 Test Scenarios

### Load Test User Flow
1. Login → 2. Account Info → 3. Market Data → 4. Place Order → 5. Order Status → 6. Positions → 7. WebSocket (30%)

### Stress Test Stages
0→500→1K→2K→3K→5K→7K→10K users (2min each)

### Spike Test Events
- **3min:** 100→2K users (1min hold)
- **8min:** 100→5K users (2min hold)
- **15min:** 100→10K users (30sec hold)

### Soak Test Checks (Every 100 iterations)
- Health check response time
- Memory usage trend
- Connection pool health
- Data integrity

## 🔍 Results Location

```
tests/performance/results/YYYYMMDD_HHMMSS/
├── load-test.json              # Raw k6 data
├── load-test-summary.json      # Aggregated metrics
├── load-test.log               # Console output
├── stress-test-results.json    # Breaking point analysis
├── spike-test-results.json     # Spike resilience
├── soak-test-results.json      # 24h stability
├── go-benchmarks.txt           # Go benchmark results
└── analysis-report.md          # Consolidated report
```

## ⚠️ Regression Indicators

| Indicator | Meaning | Action |
|-----------|---------|--------|
| 🔴 p95 > +10% | Performance degraded | **DO NOT DEPLOY** |
| 🔴 Error rate > +50% | Stability issue | **DO NOT DEPLOY** |
| 🟡 p95 ±10% | Performance stable | Monitor |
| 🟢 p95 < -10% | Performance improved | Update baseline |

## 🛠️ Troubleshooting

### High Latency
```bash
# Check slow endpoints
grep "http_req_duration" results/*/load-test-summary.json

# Profile server
go tool pprof http://localhost:8080/debug/pprof/profile
```

### High Error Rate
```bash
# Check error distribution
jq '.metrics.http_req_failed' results/*/load-test-summary.json

# View server logs
tail -f server.log | grep ERROR
```

### Memory Leak
```bash
# Run soak test
k6 run tests/performance/soak-test.js

# Check memory trend
jq '.metrics.memory_leak_indicator' results/*/soak-test-results.json
```

## 📈 Baseline Management

```bash
# Save current results as baseline
cp results/latest/load-test-summary.json tests/performance/baseline.json

# Compare against baseline
./scripts/perf/analyze-results.sh

# View baseline
cat tests/performance/baseline.json | jq
```

## 🔄 CI/CD Integration

```yaml
# Add to .github/workflows/performance.yml
- name: Performance Tests
  run: |
    export RUN_SOAK_TEST=false
    ./scripts/perf/run-load-tests.sh
    ./scripts/perf/analyze-results.sh || exit 1
```

## 📊 k6 Options Quick Reference

```javascript
// Custom VUs and duration
k6 run --vus 500 --duration 10m test.js

// With environment variables
k6 run --env BASE_URL=http://prod:8080 test.js

// Output to different format
k6 run --out json=results.json test.js
k6 run --out influxdb=http://localhost:8086 test.js

// Custom thresholds
k6 run --threshold http_req_duration=p(95)<200 test.js
```

## 🔧 Go Benchmark Options

```bash
# Basic benchmark
go test -bench=.

# With memory stats
go test -bench=. -benchmem

# Longer run time
go test -bench=. -benchtime=30s

# Specific benchmark
go test -bench=BenchmarkOrderExecution

# CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof

# Parallel execution
go test -bench=. -cpu=1,2,4,8
```

## 📞 Quick Help

```bash
# k6 help
k6 run --help
k6 --help

# View test script
cat tests/performance/load-test.js

# View analysis script
cat scripts/perf/analyze-results.sh

# Check server health
curl http://localhost:8080/health
```

## 💡 Pro Tips

1. **Always baseline first** - Establish performance baseline before changes
2. **Run locally before CI** - Catch regressions early
3. **Monitor trends** - Track performance over time, not just pass/fail
4. **Realistic data** - Use production-like data for accurate results
5. **Think time matters** - Include realistic user think time
6. **Warm up first** - Let system warm up before measuring
7. **Isolate tests** - Run tests in isolated environment for consistency
8. **Document changes** - Note what changed when performance shifts

## 🎯 Common Use Cases

### Before Deployment
```bash
./scripts/perf/run-load-tests.sh
./scripts/perf/analyze-results.sh
# If all green, safe to deploy
```

### After Optimization
```bash
# Run load test
k6 run tests/performance/load-test.js

# Compare results
./scripts/perf/analyze-results.sh

# If improved, update baseline
```

### Debugging Performance Issue
```bash
# Run stress test to find limits
k6 run tests/performance/stress-test.js

# Run Go benchmarks to identify slow paths
go test -bench=. -benchmem -cpuprofile=cpu.prof

# Profile results
go tool pprof cpu.prof
```

### Weekly Health Check
```bash
# Run load test
./scripts/perf/run-load-tests.sh

# Check trends
diff tests/performance/baseline.json results/latest/*-summary.json
```

---

**Remember:** Performance testing is most valuable when run consistently!
