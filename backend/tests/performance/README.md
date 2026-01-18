# Performance Testing Suite

Comprehensive load testing and performance benchmarking for the Trading Engine.

## 📋 Overview

This suite includes 4 types of k6 tests, Go benchmarks, and automated analysis tools:

### Test Types

1. **Load Test** (`load-test.js`) - 1000 concurrent users, realistic behavior
2. **Stress Test** (`stress-test.js`) - Find system breaking point (up to 10K users)
3. **Spike Test** (`spike-test.js`) - Handle sudden traffic spikes
4. **Soak Test** (`soak-test.js`) - 24-hour endurance test
5. **Go Benchmarks** (`benchmark.go`) - Critical path performance

## 🚀 Quick Start

### Prerequisites

```bash
# Install k6 (macOS)
brew install k6

# Install k6 (Linux)
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 \
  --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Install jq for report analysis
brew install jq  # macOS
sudo apt-get install jq  # Linux
```

### Run All Tests

```bash
# Start the trading engine server first
cd backend
./server

# In another terminal, run all tests
cd backend/scripts/perf
./run-load-tests.sh
```

### Run Individual Tests

```bash
# Load test (25 minutes)
k6 run tests/performance/load-test.js

# Stress test (19 minutes)
k6 run tests/performance/stress-test.js

# Spike test (20 minutes)
k6 run tests/performance/spike-test.js

# Soak test (24 hours - use with caution!)
k6 run tests/performance/soak-test.js

# Go benchmarks
cd tests/performance
go test -bench=. -benchmem -benchtime=10s
```

## 📊 Performance Targets

### API Performance
- **p95 Response Time:** < 200ms
- **p99 Response Time:** < 500ms
- **Error Rate:** < 1%
- **Throughput:** > 10,000 orders/second

### Trading-Specific
- **Order Execution:** < 50ms (p95)
- **WebSocket Latency:** < 10ms (p95)
- **Order Success Rate:** > 99%

### System Capacity
- **Concurrent Users:** > 5,000
- **Breaking Point:** > 7,000 users with graceful degradation
- **Spike Recovery:** > 95% success rate
- **24-Hour Stability:** No memory leaks, < 50% performance degradation

## 📈 Test Scenarios

### 1. Load Test (25 minutes)

**Simulates:** 1000 concurrent users with realistic behavior

**User Flow:**
1. Login/Authentication
2. Get account info
3. Get market data
4. Place order (BTCUSD/ETHUSD/XRPUSD/BNBUSD/SOLUSD)
5. Check order status
6. Get positions
7. WebSocket connection (30% of users)

**Metrics Tracked:**
- HTTP request duration (p50, p95, p99)
- Order placement success rate
- Order execution time
- WebSocket message latency
- Error rates

**Stages:**
- 0-2min: Ramp to 200 users
- 2-7min: Ramp to 500 users
- 7-12min: Ramp to 1000 users
- 12-22min: Hold at 1000 users
- 22-25min: Ramp down

### 2. Stress Test (19 minutes)

**Simulates:** Progressive load increase to find breaking point

**Goal:** Determine system capacity and failure mode

**Stages:**
- 0-2min: 500 users (warm up)
- 2-4min: 1,000 users
- 4-6min: 2,000 users
- 6-8min: 3,000 users
- 8-10min: 5,000 users
- 10-12min: 7,000 users
- 12-14min: 10,000 users (maximum stress)
- 14-19min: Recovery phase

**Analysis:**
- Breaking point identification
- Graceful vs catastrophic degradation
- Recovery time
- System behavior under extreme load

### 3. Spike Test (20 minutes)

**Simulates:** Sudden traffic spikes (flash crashes, news events)

**Spike Events:**
1. **Moderate Spike** (@ 3min): 100 → 2,000 users for 1 minute
2. **Severe Spike** (@ 8min): 100 → 5,000 users for 2 minutes
3. **Extreme Spike** (@ 15min): 100 → 10,000 users for 30 seconds

**Analysis:**
- Spike handling capability
- Recovery success rate
- Error rates during spikes
- System resilience

### 4. Soak Test (24 hours)

**Simulates:** Sustained production load over extended period

**Goal:** Detect memory leaks, connection leaks, performance degradation

**Load:** 500 concurrent users (moderate sustained load)

**Checks Every Hour:**
- Memory usage trends
- Connection pool health
- Performance degradation
- Data integrity
- Error rates

**Red Flags:**
- Memory growth > 50%
- Performance degradation > 2x baseline
- Connection leaks > 10
- Any data corruption events

### 5. Go Benchmarks

**Critical Paths:**
- Order execution (`BenchmarkOrderExecution`) - Target: < 50ms
- WebSocket broadcast (`BenchmarkWebSocketBroadcast`) - Target: < 10ms
- Order matching (`BenchmarkOrderMatching`)
- Risk calculation (`BenchmarkRiskCalculation`)
- P&L calculation (`BenchmarkPnLCalculation`)
- Database queries (`BenchmarkDatabaseQuery`)
- JSON marshaling/unmarshaling
- Concurrent order processing
- Memory allocation patterns

**CPU Scalability:**
Tests run with 1, 2, 4, and 8 CPUs to measure parallel scaling.

## 🔍 Results Analysis

### Automated Analysis

```bash
# Analyze latest results
./scripts/perf/analyze-results.sh

# Analyze specific results
./scripts/perf/analyze-results.sh tests/performance/results/20260118_143000
```

### Baseline Comparison

The analyzer compares results against `tests/performance/baseline.json`:

- **Green (✓):** Performance improved or stable (< ±10%)
- **Yellow (≈):** Performance stable within ±10%
- **Red (✗):** Performance regression > 10%

**Regression Detection:**
- Fails the build if p95/p99 degrades > 10%
- Fails if error rate increases > 50%
- Recommended action: Do not deploy

### Update Baseline

After confirming good performance:

```bash
# The analyzer will prompt after successful run
# Or manually update baseline.json
```

## 📁 Results Structure

```
tests/performance/results/
└── 20260118_143000/              # Timestamp-based directory
    ├── load-test.json            # Raw k6 results
    ├── load-test-summary.json    # Summary metrics
    ├── load-test.log             # Console output
    ├── stress-test-results.json  # Custom analysis
    ├── spike-test-results.json
    ├── soak-test-results.json
    ├── go-benchmarks.txt         # Go benchmark output
    └── analysis-report.md        # Consolidated report
```

## 🎯 Performance Optimization Workflow

1. **Establish Baseline**
   ```bash
   ./scripts/perf/run-load-tests.sh
   # Save as baseline when system performs well
   ```

2. **Make Changes**
   - Code optimization
   - Configuration tuning
   - Infrastructure changes

3. **Run Tests**
   ```bash
   ./scripts/perf/run-load-tests.sh
   ```

4. **Analyze Results**
   - Automated analysis shows comparison
   - Check for regressions
   - Verify improvements

5. **Act on Results**
   - ✅ No regression: Deploy
   - ⚠️ Regression: Fix before deploy
   - ✅ Improvement: Update baseline

## 🛠️ Configuration

### Environment Variables

```bash
# Server endpoints
export BASE_URL="http://localhost:8080"
export WS_URL="ws://localhost:8080/ws"

# Test selection
export RUN_LOAD_TEST=true
export RUN_STRESS_TEST=true
export RUN_SPIKE_TEST=true
export RUN_SOAK_TEST=false  # 24h test, default disabled
export RUN_GO_BENCHMARKS=true
```

### Custom Test Duration

```bash
# Run shorter load test (5 minutes)
k6 run --duration 5m tests/performance/load-test.js

# Run with specific VUs
k6 run --vus 500 --duration 10m tests/performance/load-test.js
```

## 📊 Metrics Collected

### HTTP Metrics
- `http_req_duration` - Request duration (p50, p95, p99)
- `http_req_failed` - Failed request rate
- `http_reqs` - Total requests and RPS
- `http_req_blocked` - Time blocked before request
- `http_req_connecting` - Connection time
- `http_req_sending` - Sending time
- `http_req_receiving` - Receiving time
- `http_req_waiting` - Waiting time (TTFB)

### Custom Trading Metrics
- `order_placement_success` - Order success rate
- `order_execution_time` - Order execution duration
- `websocket_message_latency` - WebSocket message latency
- `websocket_connections` - Total WS connections
- `api_errors` - Total API errors

### Stability Metrics (Soak Test)
- `memory_leak_indicator` - Memory growth trend
- `performance_degradation` - Performance over time
- `connection_leaks` - Connection pool leaks
- `data_corruption_events` - Data integrity issues
- `system_stability` - Overall stability rate

## 🚨 Troubleshooting

### High Error Rates

1. Check server logs for errors
2. Verify database connection pool size
3. Check for resource exhaustion
4. Review error distribution by endpoint

### Slow Response Times

1. Profile slow endpoints
2. Check database query performance
3. Review network latency
4. Analyze connection pool usage

### WebSocket Issues

1. Check WebSocket connection limits
2. Review message queue sizes
3. Verify broadcast performance
4. Check for connection leaks

### Memory Leaks

1. Run soak test to confirm
2. Profile memory usage
3. Check for unclosed connections
4. Review goroutine leaks

## 📚 Best Practices

1. **Run Before Deployment**
   - Always run load tests before production deploy
   - Compare against baseline
   - Check for regressions

2. **Regular Soak Tests**
   - Run 24h test before major releases
   - Monitor for stability issues
   - Establish production readiness

3. **Baseline Management**
   - Update baseline after verified improvements
   - Keep historical baselines for comparison
   - Document baseline conditions

4. **Realistic Scenarios**
   - Use production-like data
   - Simulate real user behavior
   - Include think time

5. **Continuous Monitoring**
   - Track trends over time
   - Set up alerts for regressions
   - Integrate with CI/CD

## 🔗 Integration with CI/CD

### GitHub Actions Example

```yaml
name: Performance Tests

on:
  pull_request:
    branches: [main]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Start server
        run: |
          cd backend
          ./server &
          sleep 10

      - name: Install k6
        run: |
          sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
            --keyserver hkp://keyserver.ubuntu.com:80 \
            --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
          echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
            sudo tee /etc/apt/sources.list.d/k6.list
          sudo apt-get update
          sudo apt-get install k6 jq

      - name: Run performance tests
        run: |
          cd backend/scripts/perf
          export RUN_SOAK_TEST=false  # Skip 24h test in CI
          ./run-load-tests.sh

      - name: Upload results
        uses: actions/upload-artifact@v3
        with:
          name: performance-results
          path: backend/tests/performance/results/

      - name: Check for regressions
        run: |
          cd backend/scripts/perf
          ./analyze-results.sh || exit 1
```

## 📞 Support

For questions or issues:
1. Check test logs in `tests/performance/results/`
2. Review analysis report
3. Check server logs
4. File an issue with test results attached

---

**Note:** Performance testing should be done in an environment that closely matches production to get accurate results.
