# End-to-End Market Data Pipeline Test Plan

## Executive Summary

This document provides a comprehensive testing and verification guide for the market data pipeline spanning from FIX gateway ingestion through WebSocket broadcast to frontend display. The pipeline processes market data through three critical stages: **FIX → Pipeline → WebSocket → Frontend**.

**Pipeline Architecture:**
```
FIX Gateway (YOFX)
    ↓
Raw Tick Ingestion
    ↓
Normalization & Validation
    ↓
OHLC Generation
    ↓
Redis Distribution
    ↓
WebSocket Broadcast
    ↓
Frontend Display
```

---

## Part 1: Component Testing

### 1.1 FIX Gateway Testing

#### 1.1.1 FIX Connection Verification

**Test ID:** FIX-001
**Objective:** Verify FIX gateway can establish and maintain connection to LP

**Prerequisites:**
- FIX gateway is configured (backend/fix/config/sessions.json)
- YOFX LP credentials are valid
- Network connectivity to LP endpoint available

**Manual Verification Steps:**

1. Start the backend server:
```bash
cd D:\Tading engine\Trading-Engine\backend\cmd\server
go run main.go
```

2. Monitor FIX connection logs:
```bash
# Look for successful connection output
# Expected: [FIX] Session YOFX2 connected successfully
# Check log file: backend/fixstore/YOFX2.msgs
```

3. Verify session establishment:
```bash
# Check for FIX logon message (MsgType A)
cat backend/fixstore/YOFX2.msgs | grep -i "msgtype=A"
```

**Expected Behavior:**
- Connection established within 10 seconds (logonTimeout: 10)
- Heartbeat messages every 30 seconds (heartBeatInterval: 30)
- Session ID matches config: YOFX2
- No authentication errors

**Failure Diagnosis:**
```
If connection fails:
1. Check proxy tunnel: 81.29.145.69:49527
   - Verify credentials: fGUqTcsdMsBZlms / 3eo1qF91WA7Fyku
2. Verify LP endpoint: 23.106.238.138:12336
3. Check FIX protocol config: FIX.4.4
4. Review fixstore/YOFX2.msgs for error messages
```

---

#### 1.1.2 Market Data Subscription Test

**Test ID:** FIX-002
**Objective:** Verify FIX gateway can subscribe and receive market data

**Manual Steps:**

1. With FIX gateway running, monitor incoming quotes:
```bash
# Watch real-time FIX messages
tail -f backend/fixstore/YOFX2.msgs | grep -i "35=D"  # MarketDataSnapshot messages
```

2. Verify quote structure in raw messages:
```bash
# Expected fields in each quote: Symbol (55), Bid (1D7A), Ask (1D7B), Timestamp
cat backend/fixstore/YOFX2.msgs | tail -100 | head -20
```

3. Check for multiple symbols:
```bash
# Count unique symbols
cat backend/fixstore/YOFX2.msgs | grep -oP '55=\K[^\\x01]*' | sort | uniq
```

**Expected Behavior:**
- Quotes received for subscribed symbols (EURUSD, GBPUSD, USDJPY, etc.)
- Quote frequency: 1-10 per second per symbol
- Bid/Ask spread: 0-5 pips for major pairs
- Timestamps: Within 1 second of current time

**Success Criteria:**
```
✓ EURUSD: Bid 1.0850, Ask 1.0851 (1 pip spread)
✓ GBPUSD: Bid 1.2750, Ask 1.2752 (2 pip spread)
✓ Quote arrival: Within 1 second
✓ No duplicate ticks
```

---

### 1.2 Data Pipeline Testing

#### 1.2.1 Tick Ingestion and Normalization

**Test ID:** PIPE-001
**Objective:** Verify pipeline ingests and normalizes raw FIX ticks

**Test Script:** `backend/test_pipeline.go`

```bash
cd backend
go test -v ./datapipeline -run TestTickIngestion -timeout 30s
```

**Verification:**

1. Start backend with pipeline:
```bash
go run ./cmd/server/main.go &
sleep 2
```

2. Monitor pipeline stats via admin API:
```bash
# Get pipeline statistics
curl -X GET http://localhost:8080/api/admin/pipeline-stats

# Expected response:
{
  "ticks_received": 1250,
  "ticks_processed": 1248,
  "ticks_dropped": 0,
  "ticks_duplicate": 2,
  "avg_latency_ms": 2.3,
  "clients_connected": 1
}
```

3. Verify normalized tick format:
```bash
# Check Redis for normalized ticks
redis-cli LRANGE market_data:EURUSD 0 5

# Expected format:
# {
#   "symbol": "EURUSD",
#   "bid": 1.0850,
#   "ask": 1.0851,
#   "spread": 0.0001,
#   "timestamp": "2025-01-20T14:35:22.123Z",
#   "source": "YOFX",
#   "tick_id": "abc123def456"
# }
```

**Success Criteria:**
- Ticks received > 0
- Processing latency < 5ms average
- No drops (TicksDropped = 0)
- Duplicates detected and removed

---

#### 1.2.2 OHLC Bar Generation

**Test ID:** PIPE-002
**Objective:** Verify OHLC engine generates bars correctly

**Test Script:**

```bash
cd backend
go test -v ./datapipeline -run TestOHLCGeneration -timeout 30s
```

**Manual Verification:**

1. Subscribe to OHLC channel and monitor bar generation:
```bash
# Get 1-minute OHLC bars for EURUSD
curl -X GET "http://localhost:8080/api/ohlc?symbol=EURUSD&timeframe=1m&limit=10"

# Expected response:
[
  {
    "symbol": "EURUSD",
    "open": 1.0845,
    "high": 1.0855,
    "low": 1.0840,
    "close": 1.0852,
    "timestamp": "2025-01-20T14:35:00Z",
    "volume": 245
  },
  ...
]
```

2. Verify bar timing:
```bash
# Bars should close exactly on minute/hour boundaries
# 1m bars: close at :00 seconds
# 5m bars: close at :00 and :05, :10, :15... seconds
```

3. Validate OHLC values:
```bash
# In each bar: open <= close AND low <= high
# Spread in each bar is reasonable (not >10 pips for majors)
```

**Success Criteria:**
- Bar generation latency: < 10ms
- Bars complete on time boundaries
- OHLC relationships valid (Low ≤ Open,Close ≤ High)
- Volume accurate (tick count per period)

---

#### 1.2.3 Data Quality Validation

**Test ID:** PIPE-003
**Objective:** Verify pipeline detects and handles data quality issues

**Test Scenarios:**

1. **Duplicate Detection:**
```bash
# Inject same tick twice
curl -X POST http://localhost:8080/api/admin/inject-tick \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "bid": 1.0850,
    "ask": 1.0851,
    "timestamp": "2025-01-20T14:35:22Z",
    "source": "YOFX"
  }'

# Wait 100ms and inject again with same data
# Expected: Pipeline detects duplicate, TicksDuplicate++
```

2. **Out-of-Order Detection:**
```bash
# Inject tick with old timestamp
curl -X POST http://localhost:8080/api/admin/inject-tick \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "bid": 1.0850,
    "ask": 1.0851,
    "timestamp": "2025-01-20T14:34:00Z",  # 1 minute old
    "source": "YOFX"
  }'

# Expected: TicksOutOfOrder++, tick still processed with warning
```

3. **Price Sanity Check:**
```bash
# Inject tick with unrealistic price change
curl -X POST http://localhost:8080/api/admin/inject-tick \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "EURUSD",
    "bid": 0.5000,    # 50% drop from 1.0850
    "ask": 0.5001,
    "timestamp": "2025-01-20T14:35:23Z",
    "source": "YOFX"
  }'

# Expected: TicksInvalid++, tick rejected with warning
# Check config: PriceSanityThreshold = 0.10 (10% max change)
```

**Verification Commands:**
```bash
# Check quality metrics after tests
curl -X GET http://localhost:8080/api/admin/pipeline-stats | jq '.data_quality'

# Expected output shows all issues detected:
{
  "duplicates_detected": 1,
  "out_of_order_detected": 1,
  "invalid_prices_rejected": 1
}
```

---

### 1.3 Redis Distribution Testing

#### 1.3.1 Tick Storage Verification

**Test ID:** REDIS-001
**Objective:** Verify pipeline stores ticks in Redis correctly

**Prerequisites:**
- Redis running on localhost:6379
- Pipeline started and connected

**Manual Verification:**

```bash
# Connect to Redis
redis-cli

# List all symbol keys
KEYS market_data:*

# Get latest 5 ticks for EURUSD
LRANGE market_data:EURUSD 0 4

# Check tick count
LLEN market_data:EURUSD

# Verify TTL on old ticks (should auto-expire after 30 days per config)
TTL market_data:EURUSD
```

**Expected Output:**
```
# Keys output (multiple symbols)
1) "market_data:EURUSD"
2) "market_data:GBPUSD"
3) "market_data:USDJPY"
...

# LRANGE output (tick list)
1) "{\"symbol\":\"EURUSD\",\"bid\":1.0850,\"ask\":1.0851,...}"
2) "{\"symbol\":\"EURUSD\",\"bid\":1.0851,\"ask\":1.0852,...}"
...

# TTL (should be very large, ~2592000 seconds = 30 days)
(integer) 2592000

# LLEN (count should grow)
(integer) 1847
```

---

#### 1.3.2 OHLC Bar Storage

**Test ID:** REDIS-002
**Objective:** Verify OHLC bars stored with proper indexing

**Manual Verification:**

```bash
redis-cli

# List OHLC keys (stored by symbol and timeframe)
KEYS ohlc:*

# Get 1-minute bars for EURUSD
LRANGE ohlc:EURUSD:1m 0 2

# Check OHLC bar format
# Each bar stored as: timestamp|open|high|low|close|volume
```

**Expected Output:**
```
1) "ohlc:EURUSD:1m"
2) "ohlc:EURUSD:5m"
3) "ohlc:EURUSD:15m"
4) "ohlc:EURUSD:1h"
...

# 1m bars
1) "1705778100|1.0845|1.0855|1.0840|1.0852|245"
2) "1705778160|1.0852|1.0860|1.0850|1.0858|312"
```

---

### 1.4 WebSocket Hub Testing

#### 1.4.1 Client Connection Verification

**Test ID:** WS-001
**Objective:** Verify WebSocket hub accepts and manages connections

**Test Setup:**

```bash
# In one terminal, start backend
cd backend/cmd/server
go run main.go

# In another terminal, use wscat to connect
npm install -g wscat

# Connect with token
YOFX_TOKEN=$(curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

wscat -c "ws://localhost:8080/ws?token=$YOFX_TOKEN"
```

**Expected Output:**
```
Connected (press CTRL+C to quit)
> [received] {"type":"market_tick","symbol":"EURUSD","bid":1.0850,"ask":1.0851,"timestamp":1705778122000}
> [received] {"type":"market_tick","symbol":"EURUSD","bid":1.0851,"ask":1.0852,"timestamp":1705778122100}
...
```

**Verification Steps:**

1. Verify connection accepted:
```bash
# Check backend logs for:
# [Hub] Client connected. Total clients: 1
```

2. Verify initial price snapshot sent:
```bash
# First 5 messages should be latest prices for all symbols
# Then continuous stream of market_tick updates
```

3. Monitor hub stats:
```bash
# Backend logs every 60 seconds:
# [Hub] Stats: received=12450, broadcast=10230, throttled=2220 (17.8% reduction), clients=1
```

**Success Criteria:**
- Connection established in < 500ms
- Initial snapshot sent (all symbols)
- Continuous tick stream received
- Hub stats show non-zero broadcast count

---

#### 1.4.2 Throttling Verification

**Test ID:** WS-002
**Objective:** Verify WebSocket throttling reduces broadcast load

**Test Procedure:**

1. Connect with wscat and count messages:
```bash
# Monitor message count for 10 seconds
wscat -c "ws://localhost:8080/ws?token=$YOFX_TOKEN" | wc -l
# Typical: ~100 messages/sec with throttling
# Without throttling: ~1000+ messages/sec (10x more)
```

2. Check throttling ratio in backend logs:
```bash
# Look for:
# [Hub] Stats: received=12450, broadcast=10230, throttled=2220 (17.8% reduction)
# This means 17.8% of ticks were throttled (not broadcast to clients)
```

3. Verify price accuracy not affected:
```bash
# Throttled ticks still update B-Book engine
# Verify by checking order execution prices
```

**Expected Behavior:**
- Throttle rate: 10-30% of ticks (reduces CPU by 60-80%)
- Price accuracy: Still within 1 pip for execution
- Spread: Still accurate despite throttling

---

#### 1.4.3 Multiple Client Broadcasting

**Test ID:** WS-003
**Objective:** Verify WebSocket broadcasts to multiple clients correctly

**Test Setup:**

```bash
# Terminal 1: Connect client 1
wscat -c "ws://localhost:8080/ws?token=$TOKEN1" > client1.log

# Terminal 2: Connect client 2
wscat -c "ws://localhost:8080/ws?token=$TOKEN2" > client2.log

# Terminal 3: Connect client 3
wscat -c "ws://localhost:8080/ws?token=$TOKEN3" > client3.log
```

**Verification:**

```bash
# Wait 10 seconds, then check all clients received same data
wc -l client1.log client2.log client3.log
# Output should show ~similar line counts
# 600 client1.log
# 597 client2.log
# 595 client3.log

# Verify same tick data
tail -1 client1.log
tail -1 client2.log
tail -1 client3.log
# All three should have similar latest prices
```

**Backend Logs:**
```
[Hub] Client connected. Total clients: 1
[Hub] Client connected. Total clients: 2
[Hub] Client connected. Total clients: 3
[Hub] Stats: received=15000, broadcast=13200, throttled=1800 (12%), clients=3
```

**Success Criteria:**
- All 3 clients receive data
- Hub stats show correct client count
- Broadcast happens to all clients simultaneously

---

---

## Part 2: Integration Testing

### 2.1 FIX → Pipeline Integration

**Test ID:** INT-001
**Objective:** Verify end-to-end flow from FIX to pipeline processing

**Test Procedure:**

1. Start backend and monitor flow:
```bash
# Terminal 1: Start backend
cd backend/cmd/server
go run main.go 2>&1 | tee backend.log

# Terminal 2: Monitor FIX messages
tail -f backend/fixstore/YOFX2.msgs | grep "35=D" | tail -20

# Terminal 3: Monitor pipeline stats
while true; do
  curl -s http://localhost:8080/api/admin/pipeline-stats | jq '{
    received: .data.ticks_received,
    processed: .data.ticks_processed,
    latency_ms: .data.avg_latency_ms
  }'
  sleep 5
done
```

2. Verify timing:
```
Expected sequence (for single EURUSD quote):
- T+0ms: FIX receives quote from LP
- T+2ms: Quote parsed and ingested into pipeline
- T+4ms: Quote normalized and validated
- T+6ms: Quote stored in Redis
- T+7ms: Quote distributed to WebSocket
- T+8ms: Quote broadcast to connected clients
Total latency: 8ms
```

3. Verify no data loss:
```bash
# After running for 1 minute, check:
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '{
  received: .data.ticks_received,
  processed: .data.ticks_processed,
  dropped: .data.ticks_dropped,
  lost_percent: (.data.ticks_dropped / .data.ticks_received * 100)
}'

# Expected: dropped should be 0 (unless buffer overflow)
# If dropped > 0: increase buffer sizes in config
```

---

### 2.2 Pipeline → Redis → WebSocket Integration

**Test ID:** INT-002
**Objective:** Verify data flows correctly through Redis to WebSocket clients

**Test Procedure:**

1. Set up test harness:
```bash
# Create test script: test_data_flow.sh
#!/bin/bash

# Start backend
cd backend/cmd/server
go run main.go &
BACKEND_PID=$!
sleep 3

# Connect WebSocket client and save messages
redis-cli FLUSHDB  # Clear Redis for clean test
wscat -c "ws://localhost:8080/ws?token=$YOFX_TOKEN" > ws_output.log &
WS_PID=$!

# Wait for data to flow
sleep 10

# Kill processes
kill $BACKEND_PID $WS_PID

# Analyze results
echo "=== WebSocket Message Count ==="
wc -l ws_output.log

echo "=== Redis Tick Count ==="
redis-cli LLEN market_data:EURUSD

echo "=== Check for data integrity ==="
redis-cli LRANGE market_data:EURUSD 0 0 | jq .

exit 0
```

2. Run test:
```bash
chmod +x test_data_flow.sh
./test_data_flow.sh
```

3. Verify data integrity:
```bash
# Check that ticks in Redis match WebSocket messages
cat ws_output.log | jq '.symbol + ":" + (.bid|tostring)' | sort > ws_ticks.txt
redis-cli LRANGE market_data:EURUSD 0 -1 | jq '.symbol + ":" + (.bid|tostring)' | sort > redis_ticks.txt

# Compare
diff ws_ticks.txt redis_ticks.txt
# Should have minimal differences (only order difference acceptable)
```

**Success Criteria:**
- WebSocket messages > 0
- Redis tick count > 0
- No ticks only in Redis (data reaches WS)
- All ticks in Redis also broadcast to WS

---

### 2.3 Full Pipeline Load Test

**Test ID:** INT-003
**Objective:** Verify pipeline handles sustained high-frequency data

**Test Script:** `backend/cmd/load_test/main.go`

```bash
cd backend/cmd/load_test
go run main.go \
  -ticks-per-second 1000 \
  -symbols 15 \
  -clients 10 \
  -duration 60s
```

**Monitoring:**

```bash
# Terminal 1: Pipeline stats
watch -n 1 'curl -s http://localhost:8080/api/admin/pipeline-stats | jq'

# Terminal 2: Redis memory
redis-cli INFO memory | grep used_memory_human

# Terminal 3: System resources
htop
# Monitor CPU, memory for go processes
```

**Test Scenarios:**

1. **1000 ticks/sec, 1 client:**
```
Expected:
- Latency: 5-10ms average
- CPU: 20-30%
- Memory: 150MB
- Dropped: 0
```

2. **1000 ticks/sec, 10 clients:**
```
Expected:
- Latency: 10-15ms average
- CPU: 40-50%
- Memory: 250MB
- Dropped: 0
- WebSocket broadcast: 10x load
```

3. **5000 ticks/sec spike, 10 clients:**
```
Expected:
- Latency: 15-20ms average
- CPU: 70-80%
- Memory: 300MB
- Dropped: < 1% (acceptable during spikes)
```

**Success Criteria:**
- Sustained 1000 ticks/sec with <1% drop
- No client disconnections
- Memory stable (no leaks)
- Latency stays under 20ms

---

---

## Part 3: Frontend Integration Testing

### 3.1 WebSocket Connection in Browser

**Test ID:** FRONT-001
**Objective:** Verify frontend connects to backend WebSocket and receives data

**Test Setup:**

1. Start backend:
```bash
cd backend/cmd/server
go run main.go
```

2. Start frontend:
```bash
cd clients/desktop
npm start
# Opens http://localhost:3000
```

3. Open browser console (F12 → Console tab)

**Manual Verification:**

1. Login to application:
```
Username: trader
Password: trader
```

2. Monitor WebSocket in browser console:
```javascript
// In browser console, manually trigger logging
window.addEventListener('message', (e) => {
  if (e.data.type === 'market_tick') {
    console.log('Market Tick:', e.data);
  }
});
```

3. Check WebSocket status:
```javascript
// Check connection state
if (window.WebSocket) {
  console.log('WebSocket supported');
  // Look for [WS] logs in console
}
```

**Expected Output in Console:**
```
[WS] Connecting to ws://localhost:8080/ws...
[WS] Connected successfully
[WS] Message received: {"type":"market_tick","symbol":"EURUSD","bid":1.0850...}
[WS] Message received: {"type":"market_tick","symbol":"EURUSD","bid":1.0851...}
```

**Verification Checklist:**
- [ ] WebSocket connects without error
- [ ] Market Watch shows live prices
- [ ] Prices update in real-time (< 1 sec)
- [ ] Bid/Ask spread displayed correctly
- [ ] Multiple symbols update simultaneously

---

### 3.2 Market Watch Display Verification

**Test ID:** FRONT-002
**Objective:** Verify market data displays correctly in frontend UI

**Manual Steps:**

1. Open Market Watch component:
```
Click: Trading → Market Watch
```

2. Verify symbol columns:
```
Expected columns:
- Symbol (e.g., EURUSD)
- Bid (e.g., 1.0850)
- Ask (e.g., 1.0851)
- Spread (e.g., 1.0)
- Change (color coded: green for up, red for down)
- Last Updated (timestamp)
```

3. Verify real-time updates:
```
- Watch bid/ask for 30 seconds
- Both should change (move up or down)
- Timestamp should update with each change
- Spread should stay within normal range (0-5 pips for majors)
```

4. Test symbol filtering:
```
Actions:
- Right-click on symbol
- Select "Hide" (if available)
- Verify symbol disappears from list

- Right-click again
- Select "Show"
- Verify symbol reappears
```

5. Check color coding:
```
Expected behavior:
- Green when price moved up since last update
- Red when price moved down
- Gray when price unchanged
```

**Success Criteria:**
```
✓ All symbols display with correct format
✓ Prices update within 1 second
✓ Spread stays reasonable (< 5 pips majors, < 20 pips exotics)
✓ Bid always < Ask
✓ Color coding works correctly
✓ Timestamp updates with each tick
```

---

### 3.3 Chart/OHLC Display Verification

**Test ID:** FRONT-003
**Objective:** Verify charts display OHLC data correctly

**Manual Steps:**

1. Open chart view:
```
Click: Trading → Charts
```

2. Verify timeframe controls:
```
Expected timeframes:
- 1m (1 minute)
- 5m (5 minutes)
- 15m (15 minutes)
- 1h (1 hour)
- 4h (4 hours)
- 1d (1 day)
```

3. Select timeframe and verify bars:
```
Action: Click "1m" timeframe
Expected:
- Chart loads OHLC bars for EURUSD
- Each bar represents 1 minute
- Bars close at :00 seconds
- Open ≤ Close or Open ≥ Close (one or other)
- Low ≤ High always
```

4. Verify bar values:
```javascript
// In console, verify bar math
// For EURUSD 1m bar:
// Open: 1.0850
// High: 1.0855 (should be >= all other prices)
// Low: 1.0840 (should be <= all other prices)
// Close: 1.0852

// Check: Low ≤ {Open, Close} ≤ High
if (low <= open && open <= high && low <= close && close <= high) {
  console.log('OHLC valid');
} else {
  console.log('OHLC invalid!');
}
```

5. Test timeframe switching:
```
Action: Switch from 1m → 5m
Expected:
- Chart updates
- Each bar now represents 5 minutes
- Fewer bars displayed (1/5th as many)
- Bar values aggregate correctly (5 × 1m bars → 1 × 5m bar)
```

**Success Criteria:**
```
✓ All timeframes load without error
✓ OHLC relationships valid (Low ≤ Open,Close ≤ High)
✓ Bars align on time boundaries (:00, :05, :10... for 5m)
✓ Candle colors correct (green for up close, red for down close)
✓ Switching timeframes works smoothly
```

---

---

## Part 4: System Verification Commands

### 4.1 Quick Health Check Script

**File:** `scripts/pipeline_health_check.sh`

```bash
#!/bin/bash

echo "=== Market Data Pipeline Health Check ==="
echo "Time: $(date)"
echo

# Check backend service
echo "1. Backend Service Status:"
if pgrep -f "go run.*server" > /dev/null; then
  echo "   ✓ Backend running"
else
  echo "   ✗ Backend NOT running"
  exit 1
fi

# Check Redis
echo "2. Redis Status:"
if redis-cli ping > /dev/null 2>&1; then
  REDIS_MEMORY=$(redis-cli INFO memory | grep used_memory_human | cut -d: -f2)
  echo "   ✓ Redis online (Memory: $REDIS_MEMORY)"
else
  echo "   ✗ Redis NOT running"
fi

# Check FIX connection
echo "3. FIX Gateway Status:"
FIX_STATUS=$(curl -s http://localhost:8080/api/admin/fix-status 2>/dev/null | jq -r '.YOFX2' 2>/dev/null)
if [ "$FIX_STATUS" = "connected" ]; then
  echo "   ✓ FIX connected"
elif [ "$FIX_STATUS" = "connecting" ]; then
  echo "   ⏳ FIX connecting..."
else
  echo "   ✗ FIX disconnected"
fi

# Check pipeline stats
echo "4. Pipeline Statistics:"
STATS=$(curl -s http://localhost:8080/api/admin/pipeline-stats 2>/dev/null)
if [ -n "$STATS" ]; then
  TICKS=$(echo $STATS | jq -r '.data.ticks_received' 2>/dev/null)
  LATENCY=$(echo $STATS | jq -r '.data.avg_latency_ms' 2>/dev/null)
  DROPPED=$(echo $STATS | jq -r '.data.ticks_dropped' 2>/dev/null)
  echo "   Ticks received: $TICKS"
  echo "   Avg latency: ${LATENCY}ms"
  echo "   Dropped: $DROPPED"

  if (( $(echo "$LATENCY < 10" | bc -l) )); then
    echo "   ✓ Latency acceptable"
  else
    echo "   ⚠ Latency high (> 10ms)"
  fi
else
  echo "   ✗ Cannot reach pipeline stats"
fi

# Check Redis data volume
echo "5. Data Storage:"
EURUSD_COUNT=$(redis-cli LLEN market_data:EURUSD 2>/dev/null)
echo "   EURUSD ticks in Redis: $EURUSD_COUNT"

# Check WebSocket connections
echo "6. WebSocket Clients:"
WS_CLIENTS=$(curl -s http://localhost:8080/api/admin/ws-status 2>/dev/null | jq -r '.connected_clients' 2>/dev/null)
echo "   Connected clients: $WS_CLIENTS"

echo
echo "=== Health Check Complete ==="
```

**Run it:**
```bash
chmod +x scripts/pipeline_health_check.sh
./scripts/pipeline_health_check.sh
```

---

### 4.2 Diagnostic Queries

**Redis Diagnostics:**
```bash
# Check all market data keys
redis-cli KEYS 'market_data:*' | wc -l

# Get size of tick history
redis-cli LLEN market_data:EURUSD

# Check OHLC bars generated
redis-cli KEYS 'ohlc:*' | sort

# Monitor Redis memory growth
watch -n 5 'redis-cli INFO memory | grep used_memory_human'

# Find oldest tick
redis-cli LINDEX market_data:EURUSD -1 | jq '.timestamp'

# Find newest tick
redis-cli LINDEX market_data:EURUSD 0 | jq '.timestamp'
```

**Backend Diagnostics:**
```bash
# Check FIX message count
wc -l backend/fixstore/YOFX2.msgs

# Check for errors in FIX log
grep -i error backend/fixstore/YOFX2.msgs | tail -10

# Monitor backend memory
watch -n 2 'ps aux | grep "go run"'

# Check active goroutines
curl -s http://localhost:6060/debug/pprof/goroutine | grep -c goroutine
```

**Frontend Diagnostics:**
```bash
# Check WebSocket connection state in DevTools
# F12 → Network → WS → Click ws connection → Frames tab
# Should see incoming messages like:
# {"type":"market_tick","symbol":"EURUSD",...}

# Monitor frontend performance
# F12 → Performance → Record 30 seconds → Stop
# Check for:
# - JavaScript execution time
# - WebSocket message processing time
# - UI rendering time
```

---

---

## Part 5: Troubleshooting Guide

### 5.1 Common Issues and Solutions

#### Issue: No market data received

**Symptoms:**
- WebSocket connects but no ticks received
- Pipeline stats show 0 ticks_received
- Charts are empty

**Diagnosis Steps:**
```bash
# 1. Check FIX connection
curl -s http://localhost:8080/api/admin/fix-status | jq .

# 2. Check FIX log for errors
tail -50 backend/fixstore/YOFX2.msgs | grep -i "error\|fail\|disconnect"

# 3. Check pipeline status
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data | {received, processed, dropped}'

# 4. Verify Redis is working
redis-cli PING

# 5. Check symbol subscription
# Look in FIX logs for MarketDataRequest (MsgType: V)
grep "35=V" backend/fixstore/YOFX2.msgs | tail -5
```

**Solutions:**
1. **FIX not connected:** Check LP credentials and proxy settings in backend/fix/config/sessions.json
2. **No subscriptions:** Verify symbol list is configured, restart pipeline to resubscribe
3. **Redis down:** Start Redis: `redis-server`
4. **Pipeline crashed:** Check backend logs, restart with: `go run ./cmd/server/main.go`

---

#### Issue: High latency (> 20ms)

**Symptoms:**
- Backend logs show avg_latency_ms > 20
- Frontend updates slow (> 1 second between ticks)
- Charts lag behind current market

**Diagnosis:**
```bash
# 1. Check CPU/Memory usage
top -n 1 | grep go

# 2. Check Redis response time
redis-cli --latency-history

# 3. Check network latency to LP
ping 23.106.238.138

# 4. Check tick arrival rate
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_received'
# Run again after 5 seconds
# Should see significant increase (100+ new ticks per second)

# 5. Check for buffer overflow
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_dropped'
# If > 0, buffers are too small
```

**Solutions:**
1. **Buffer overflow:** Increase buffer sizes in backend/config.json:
```json
{
  "TickBufferSize": 50000,  // Increase from 10000
  "OHLCBufferSize": 5000,   // Increase from 1000
  "DistributionBufferSize": 10000  // Increase from 5000
}
```

2. **CPU overload:** Reduce tick broadcast frequency (increase throttle threshold)
3. **Network latency:** Consider co-locating backend near LP
4. **Memory pressure:** Monitor with `htop`, restart if > 1GB

---

#### Issue: WebSocket disconnects frequently

**Symptoms:**
- Connection drops every few minutes
- Console shows "connection closed (code: 1006)"
- Frontend shows "disconnected" status

**Diagnosis:**
```bash
# 1. Check WebSocket hub logs
grep "Connection closed" backend.log | tail -10

# 2. Check for authentication errors
grep "Authentication" backend.log | tail -5

# 3. Check token expiration
# Frontend stores token in localStorage
# Tokens expire after 24 hours
curl -X GET http://localhost:8080/api/auth/verify \
  -H "Authorization: Bearer $TOKEN" | jq .

# 4. Check proxy tunnel stability
curl -v http://81.29.145.69:49527
```

**Solutions:**
1. **Token expired:** Clear localStorage and re-login:
```javascript
localStorage.removeItem('rtx_token');
// Reload page
```

2. **Proxy unstable:** Check proxy credentials or switch to direct connection
3. **Browser issue:** Try different browser (Chrome, Firefox)
4. **Server overloaded:** Reduce number of connected clients

---

#### Issue: Data corruption (wrong bid/ask values)

**Symptoms:**
- Bid >= Ask (should be Bid < Ask)
- Spread > 50 pips (abnormal)
- Price jumps 10%+ in single tick
- Charts show impossible OHLC relationships (High < Low)

**Diagnosis:**
```bash
# 1. Check latest Redis ticks for corruption
redis-cli LRANGE market_data:EURUSD 0 10 | jq 'select(.bid >= .ask)'

# 2. Check pipeline quality metrics
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data_quality'

# 3. Check FIX message parsing
# Look for unusual bid/ask values in YOFX2.msgs
grep "1D7A\|1D7B" backend/fixstore/YOFX2.msgs | tail -5

# 4. Verify OHLC calculation
# Check for invalid bars in Redis
redis-cli LRANGE ohlc:EURUSD:1m 0 5 | jq 'select(.high < .low)'
```

**Solutions:**
1. **Enable price validation:** Ensure PriceSanityThreshold is configured:
```json
{
  "PriceSanityThreshold": 0.10  // Reject 10%+ changes
}
```

2. **Clear corrupted data:**
```bash
redis-cli DEL market_data:EURUSD  # Clear ticks for symbol
redis-cli DEL ohlc:EURUSD:*       # Clear all OHLC bars
# Restart pipeline to regenerate from FIX
```

3. **Check FIX source:** May be receiving bad data from LP
   - Verify with LP support
   - Check LP documentation for data format

---

### 5.2 Performance Tuning Guide

**For Low Latency (< 5ms):**
```go
// In backend/datapipeline/pipeline.go
PipelineConfig{
  TickBufferSize: 50000,
  OHLCBufferSize: 10000,
  DistributionBufferSize: 20000,
  WorkerCount: 8,  // Increase for more parallelism
  EnableDeduplication: false,  // Skip dedup if latency critical
  EnableOutOfOrderCheck: false,
}
```

**For High Throughput (1000+ ticks/sec):**
```go
PipelineConfig{
  TickBufferSize: 100000,
  OHLCBufferSize: 20000,
  DistributionBufferSize: 50000,
  WorkerCount: 16,
  EnableDeduplication: true,  // Keep enabled
  EnableOutOfOrderCheck: false,  // Skip for speed
}
```

**For Data Quality (max accuracy):**
```go
PipelineConfig{
  TickBufferSize: 10000,
  OHLCBufferSize: 5000,
  DistributionBufferSize: 5000,
  WorkerCount: 4,
  EnableDeduplication: true,
  EnableOutOfOrderCheck: true,
  MaxTickAgeSeconds: 30,
  PriceSanityThreshold: 0.05,  // Strict 5% threshold
}
```

---

---

## Part 6: Test Automation Scripts

### 6.1 Automated End-to-End Test

**File:** `backend/cmd/test_e2e/main.go`

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== End-to-End Pipeline Test ===")

	// 1. Check backend health
	fmt.Println("\n1. Checking backend health...")
	resp, err := http.Get("http://localhost:8080/api/admin/pipeline-stats")
	if err != nil {
		log.Fatal("Backend not running: ", err)
	}
	defer resp.Body.Close()

	var stats map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&stats)
	fmt.Printf("   Pipeline stats: %v\n", stats)

	// 2. Check Redis
	fmt.Println("\n2. Checking Redis...")
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	count, err := rdb.LLen(ctx, "market_data:EURUSD").Result()
	if err != nil {
		log.Fatal("Redis not running: ", err)
	}
	fmt.Printf("   EURUSD ticks in Redis: %d\n", count)

	// 3. Connect WebSocket
	fmt.Println("\n3. Connecting WebSocket...")
	token := "test_token"  // Get real token first
	ws, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("ws://localhost:8080/ws?token=%s", token), nil)
	if err != nil {
		log.Fatal("WebSocket failed: ", err)
	}
	defer ws.Close()

	// 4. Receive messages
	fmt.Println("\n4. Receiving market data...")
	tickCount := 0
	timeout := time.After(10 * time.Second)

	for {
		select {
		case <-timeout:
			goto done
		default:
			ws.SetReadDeadline(time.Now().Add(1 * time.Second))

			_, data, err := ws.ReadMessage()
			if err != nil {
				continue
			}

			var tick map[string]interface{}
			json.Unmarshal(data, &tick)

			if tick["type"] == "market_tick" {
				tickCount++
				if tickCount%100 == 0 {
					fmt.Printf("   Received %d ticks\n", tickCount)
				}
			}
		}
	}

done:
	fmt.Printf("\n=== Test Results ===\n")
	fmt.Printf("Total ticks received: %d\n", tickCount)

	if tickCount > 100 {
		fmt.Println("✓ TEST PASSED")
	} else {
		fmt.Println("✗ TEST FAILED - insufficient data")
	}
}
```

**Run:**
```bash
cd backend/cmd/test_e2e
go run main.go
```

---

### 6.2 Load Test Script

**File:** `test_load.sh`

```bash
#!/bin/bash

echo "=== Market Data Pipeline Load Test ==="
echo "Configuring: 1000 ticks/sec, 10 clients, 60 seconds"
echo

# Start backend with performance config
export PIPELINE_CONFIG="{
  \"TickBufferSize\": 100000,
  \"WorkerCount\": 16
}"

cd backend/cmd/server
go run main.go &
BACKEND_PID=$!
sleep 3

# Create 10 WebSocket clients
for i in {1..10}; do
  wscat -c "ws://localhost:8080/ws?token=test" > client_$i.log &
  CLIENT_PID=$!
  echo "Client $i started (PID: $CLIENT_PID)"
done

# Monitor for 60 seconds
for i in {1..12}; do
  sleep 5
  STATS=$(curl -s http://localhost:8080/api/admin/pipeline-stats)
  echo "Iteration $i:"
  echo "$STATS" | jq '{
    ticks_received: .data.ticks_received,
    latency_ms: .data.avg_latency_ms,
    dropped: .data.ticks_dropped,
    clients: .data.clients_connected
  }'
done

# Cleanup
kill $BACKEND_PID
killall wscat

# Analyze results
echo
echo "=== Final Analysis ==="
TOTAL_LINES=$(cat client_*.log | wc -l)
TOTAL_CLIENTS=10
AVG_PER_CLIENT=$((TOTAL_LINES / TOTAL_CLIENTS))

echo "Total messages across all clients: $TOTAL_LINES"
echo "Average per client: $AVG_PER_CLIENT"
echo "Rate: $((AVG_PER_CLIENT / 60)) messages/sec/client"

if (( $(echo "$AVG_PER_CLIENT > 60000" | bc -l) )); then
  echo "✓ PASS: High throughput maintained"
else
  echo "✗ FAIL: Throughput below target"
fi

rm client_*.log
```

**Run:**
```bash
chmod +x test_load.sh
./test_load.sh
```

---

---

## Part 7: Success Criteria Checklist

### 7.1 Component-Level Success Criteria

- [ ] **FIX Connection:**
  - [ ] Connects within 10 seconds
  - [ ] Maintains heartbeat (every 30 seconds)
  - [ ] Logs show "connected" status
  - [ ] No authentication errors

- [ ] **Data Ingestion:**
  - [ ] Receives 100+ ticks/sec
  - [ ] Latency < 5ms per tick
  - [ ] No buffer overflow (TicksDropped = 0)
  - [ ] Duplicate detection working

- [ ] **OHLC Generation:**
  - [ ] Bars close on time boundaries
  - [ ] OHLC relationships valid
  - [ ] Volume accurate
  - [ ] Generation latency < 10ms

- [ ] **Redis Storage:**
  - [ ] Ticks persisted with TTL
  - [ ] OHLC bars indexed by symbol/timeframe
  - [ ] Memory usage reasonable (< 500MB for 1 day data)
  - [ ] Recovery after restart works

- [ ] **WebSocket Distribution:**
  - [ ] Accepts multiple clients (10+)
  - [ ] Broadcasts to all clients simultaneously
  - [ ] Throttling reduces CPU by 60%+
  - [ ] No message loss

---

### 7.2 Integration-Level Success Criteria

- [ ] **FIX → Pipeline:**
  - [ ] 100% of FIX ticks processed
  - [ ] End-to-end latency < 20ms
  - [ ] No data corruption
  - [ ] Bidirectional synchronization

- [ ] **Pipeline → Redis → WebSocket:**
  - [ ] All ticks reach WebSocket clients
  - [ ] Redis acts as reliable buffer
  - [ ] Clients see same data
  - [ ] Historical data queryable

- [ ] **Sustained Load Performance:**
  - [ ] 1000 ticks/sec continuous
  - [ ] Latency stays < 15ms under load
  - [ ] Memory stable (no leaks)
  - [ ] CPU < 80%

---

### 7.3 Frontend-Level Success Criteria

- [ ] **Market Watch Display:**
  - [ ] All symbols shown
  - [ ] Bid/Ask updated in real-time
  - [ ] Spread calculated correctly
  - [ ] Color coding works

- [ ] **Chart Functionality:**
  - [ ] All timeframes load
  - [ ] OHLC candles display correctly
  - [ ] Timeframe switching smooth
  - [ ] Performance acceptable (no jank)

- [ ] **Data Accuracy:**
  - [ ] Frontend data matches Redis
  - [ ] OHLC math validated
  - [ ] No stale data displayed
  - [ ] Latency < 1 second visible

---

---

## Part 8: Documentation

### Test Reports Template

**File:** `test_report_template.md`

```markdown
# Pipeline Test Report

**Date:** [Date]
**Tester:** [Name]
**Duration:** [Start - End]

## Preconditions
- [ ] Backend running
- [ ] Redis running
- [ ] FIX connected
- [ ] Clients available

## Test Results

### Component Tests
- [ ] FIX-001: Connection - PASS/FAIL
- [ ] FIX-002: Market Data - PASS/FAIL
- [ ] PIPE-001: Ingestion - PASS/FAIL
- [ ] PIPE-002: OHLC - PASS/FAIL
- [ ] PIPE-003: Quality - PASS/FAIL
- [ ] REDIS-001: Storage - PASS/FAIL
- [ ] REDIS-002: OHLC - PASS/FAIL
- [ ] WS-001: Connection - PASS/FAIL
- [ ] WS-002: Throttling - PASS/FAIL
- [ ] WS-003: Multi-Client - PASS/FAIL

### Metrics

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Latency (ms) | XX | < 10 | ✓/✗ |
| Throughput | XX ticks/sec | > 500 | ✓/✗ |
| Dropped | X | = 0 | ✓/✗ |
| CPU (%) | XX | < 50 | ✓/✗ |
| Memory (MB) | XX | < 500 | ✓/✗ |

### Issues Found
1. [Issue description]
2. [Issue description]

### Recommendations
1. [Recommendation]
2. [Recommendation]

**Overall Result:** PASS / FAIL
```

---

## Appendix A: Quick Reference Commands

```bash
# Start all services
cd backend/cmd/server && go run main.go &
redis-server &

# Backend health
curl http://localhost:8080/api/admin/pipeline-stats | jq .

# WebSocket test
YOFX_TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

wscat -c "ws://localhost:8080/ws?token=$YOFX_TOKEN"

# Redis monitoring
redis-cli MONITOR

# Check tick flow
redis-cli LLEN market_data:EURUSD

# Clear data
redis-cli FLUSHDB

# View FIX logs
tail -f backend/fixstore/YOFX2.msgs
```

---

## Appendix B: Configuration Reference

**Default Buffer Sizes:**
```
TickBufferSize: 10,000 ticks
OHLCBufferSize: 1,000 bars
DistributionBufferSize: 5,000 ticks
WebSocket Client Buffer: 1,024 messages
```

**Default Thresholds:**
```
MaxTickAgeSeconds: 60 (reject ticks older than 1 minute)
PriceSanityThreshold: 0.10 (10% max change per tick)
StaleQuoteThreshold: 10 seconds (warning after no updates)
HealthCheckInterval: 30 seconds
```

**Default Retention:**
```
Hot Data (Redis): 1,000 latest ticks per symbol
Warm Data (TimescaleDB): 30 days
OHLC History: 2 years (if DB configured)
```

---

End of End-to-End Test Plan
