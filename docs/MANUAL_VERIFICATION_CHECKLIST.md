# Manual Verification Checklist for Market Data Pipeline

## Quick Start Verification (5 minutes)

Use this checklist to quickly verify the pipeline is working correctly.

### Pre-Flight Checks

- [ ] Backend running: `go run ./cmd/server/main.go`
- [ ] Redis running: `redis-server`
- [ ] FIX session configured: `backend/fix/config/sessions.json` exists
- [ ] LP credentials valid: YOFX proxy reachable

### Step 1: Verify FIX Connection

**Time: 1 minute**

```bash
# Terminal: Monitor FIX logs
tail -f backend/fixstore/YOFX2.msgs
```

**Expected Output:**
```
FIX.4.4 | 35=A | ... | 49=YOFX2 | 56=YOFX | ...  (Logon message)
FIX.4.4 | 35=0 | ... | (Heartbeat every 30 seconds)
```

**Checklist:**
- [ ] See "35=A" (logon) message within 10 seconds
- [ ] See "35=0" (heartbeat) messages regularly
- [ ] No error messages or disconnects
- [ ] Session ID shows: YOFX2

### Step 2: Verify Market Data Reception

**Time: 1 minute**

```bash
# Terminal: Check market data messages
tail -f backend/fixstore/YOFX2.msgs | grep "35=D"
```

**Expected Output:**
```
FIX.4.4 | 35=D | ... | 55=EURUSD | 1D7A=1.0850 | 1D7B=1.0851 | ...
FIX.4.4 | 35=D | ... | 55=GBPUSD | 1D7A=1.2750 | 1D7B=1.2752 | ...
...
```

**Checklist:**
- [ ] Market data messages (35=D) appearing
- [ ] Multiple symbols (EURUSD, GBPUSD, etc.)
- [ ] New messages every 100-1000ms
- [ ] Bid and Ask prices realistic (not 0 or extreme values)

### Step 3: Verify Pipeline Processing

**Time: 1 minute**

```bash
# Terminal: Check pipeline stats
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data | {ticks_received, avg_latency_ms, clients_connected}'
```

**Expected Output:**
```json
{
  "ticks_received": 1250,
  "avg_latency_ms": 3.5,
  "clients_connected": 0
}
```

**Checklist:**
- [ ] ticks_received > 0 (should grow)
- [ ] avg_latency_ms < 10
- [ ] Run command again after 5 seconds
- [ ] ticks_received increased significantly

### Step 4: Verify Redis Storage

**Time: 1 minute**

```bash
# Terminal: Check Redis data
redis-cli LLEN market_data:EURUSD
redis-cli LINDEX market_data:EURUSD 0 | jq .
```

**Expected Output:**
```
(integer) 847
{"symbol":"EURUSD","bid":1.0850,"ask":1.0851,"spread":0.0001,"timestamp":"2025-01-20T14:35:22Z",...}
```

**Checklist:**
- [ ] Tick count > 0
- [ ] Tick count grows when you run again
- [ ] Tick JSON is valid (jq can parse it)
- [ ] bid < ask (always)

### Step 5: Verify WebSocket Connection

**Time: 1 minute**

```bash
# Get auth token first
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

# Connect with wscat (install: npm install -g wscat)
wscat -c "ws://localhost:8080/ws?token=$TOKEN"
```

**Expected Output:**
```
Connected (press CTRL+C to quit)
> {"type":"market_tick","symbol":"EURUSD","bid":1.0850,"ask":1.0851,"timestamp":...}
> {"type":"market_tick","symbol":"GBPUSD","bid":1.2750,"ask":1.2752,"timestamp":...}
...
```

**Checklist:**
- [ ] Connection succeeds (no 401 error)
- [ ] Market ticks arriving immediately
- [ ] New ticks every second
- [ ] Multiple symbols in stream

---

## Component-Level Verification (15 minutes)

### A. FIX Gateway Verification

#### A1. FIX Connection Stability

**Test Duration:** 2 minutes

```bash
# Watch for disconnections
tail -f backend/fixstore/YOFX2.msgs | grep -i "disconnect\|logout\|error"
```

**Pass Criteria:**
- [ ] No disconnect messages
- [ ] No error messages
- [ ] Heartbeats every 30 seconds without gaps

#### A2. FIX Symbol Subscription

**Test Duration:** 1 minute

```bash
# Check that multiple symbols are being quoted
cat backend/fixstore/YOFX2.msgs | grep "35=D" | cut -d'|' -f7 | sort | uniq | head -20
```

**Expected Output:**
```
55=EURUSD
55=GBPUSD
55=USDJPY
55=AUDUSD
...
```

**Pass Criteria:**
- [ ] 10+ different symbols being quoted
- [ ] All expected symbols present (EURUSD, GBPUSD, USDJPY)
- [ ] New symbols appearing regularly

#### A3. FIX Quote Format

**Test Duration:** 1 minute

```bash
# Verify each quote has required fields
tail -20 backend/fixstore/YOFX2.msgs | grep "35=D" | head -3
```

**Check Each Quote:**
- [ ] Contains 35=D (market data snapshot)
- [ ] Contains 55=XXXXX (symbol)
- [ ] Contains 1D7A=X.XXXX (bid price)
- [ ] Contains 1D7B=X.XXXX (ask price)
- [ ] Contains 52=... (timestamp)

---

### B. Pipeline Processing Verification

#### B1. Tick Ingestion Rate

**Test Duration:** 2 minutes

```bash
# Get stats at T=0
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_received' > /tmp/stats_0.txt

# Wait 30 seconds
sleep 30

# Get stats at T=30
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_received' > /tmp/stats_30.txt

# Calculate rate
DIFF=$(($(cat /tmp/stats_30.txt) - $(cat /tmp/stats_0.txt)))
RATE=$((DIFF / 30))
echo "Ingestion rate: $RATE ticks/sec"
```

**Pass Criteria:**
- [ ] Rate > 100 ticks/sec (at least 100 per second)
- [ ] Rate consistent across multiple measurements
- [ ] No significant fluctuations

#### B2. Data Quality Validation

**Test Duration:** 1 minute

```bash
# Check for duplicates detected
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_duplicate'

# Check for out-of-order ticks
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_out_of_order'

# Check for invalid prices
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data.ticks_invalid'
```

**Pass Criteria:**
- [ ] No duplicate tick IDs (or very few < 1%)
- [ ] No out-of-order ticks (or very few < 1%)
- [ ] No invalid prices (should be 0)

#### B3. OHLC Generation

**Test Duration:** 2 minutes

```bash
# Get OHLC bars for a symbol
curl -s "http://localhost:8080/api/ohlc?symbol=EURUSD&timeframe=1m&limit=5"

# Count total 1m bars generated
redis-cli LLEN ohlc:EURUSD:1m
redis-cli LLEN ohlc:EURUSD:5m
redis-cli LLEN ohlc:EURUSD:1h
```

**Check Each Bar:**
- [ ] Has OHLC values (Open, High, Low, Close)
- [ ] Low ≤ Open, Close ≤ High (always)
- [ ] Timestamp on minute boundary (for 1m bars)
- [ ] Close ≠ Open (unless no price change)

**Pass Criteria:**
- [ ] 1m bars: > 50
- [ ] 5m bars: > 10
- [ ] 1h bars: > 1
- [ ] All OHLC relationships valid

---

### C. Redis Storage Verification

#### C1. Tick Storage

**Test Duration:** 1 minute

```bash
# Get tick history for symbol
redis-cli LRANGE market_data:EURUSD 0 4

# Verify timestamps are in descending order (newest first)
redis-cli LRANGE market_data:EURUSD 0 100 | jq -r '.timestamp' | head -5
```

**Pass Criteria:**
- [ ] Ticks stored in Redis
- [ ] Newest tick first (descending timestamp)
- [ ] Timestamps within last 5 minutes
- [ ] No gaps in tick history

#### C2. Redis Memory Usage

**Test Duration:** 1 minute

```bash
# Check Redis memory usage
redis-cli INFO memory | grep used_memory_human

# Check total keys
redis-cli DBSIZE

# Get typical tick size
redis-cli LINDEX market_data:EURUSD 0 | wc -c
```

**Pass Criteria:**
- [ ] Memory < 500MB (for 1 day of data)
- [ ] Keys > 100 (multiple symbols stored)
- [ ] Each tick ~100-200 bytes

---

### D. WebSocket Broadcasting Verification

#### D1. Single Client Connection

**Test Duration:** 2 minutes

**Terminal 1:**
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

wscat -c "ws://localhost:8080/ws?token=$TOKEN" | tee client1.log
```

**Terminal 2 (after 30 seconds):**
```bash
# Count messages
wc -l client1.log

# Check message rate
LINES=$(wc -l < client1.log)
RATE=$((LINES / 30))
echo "WebSocket rate: $RATE messages/sec"
```

**Pass Criteria:**
- [ ] Connection succeeds
- [ ] Continuous stream of messages
- [ ] Rate > 10 messages/sec
- [ ] All messages valid JSON

#### D2. Multi-Client Broadcasting

**Test Duration:** 2 minutes

**Terminal 1, 2, 3:**
```bash
# Connect 3 clients
for i in 1 2 3; do
  TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
    -H "Content-Type: application/json" \
    -d '{"username":"trader","password":"trader"}' | jq -r '.token')

  wscat -c "ws://localhost:8080/ws?token=$TOKEN" > client_$i.log &
done

# Wait 30 seconds
sleep 30

# Kill all connections (Ctrl+C on all terminals)

# Compare message counts
wc -l client_*.log
```

**Pass Criteria:**
- [ ] All 3 clients connected
- [ ] Similar message counts (within 10% of each other)
- [ ] All messages same (same symbol/price data)
- [ ] No disconnections during 30 seconds

#### D3. Throttling Effectiveness

**Test Duration:** 2 minutes

**Terminal 1:**
```bash
# Get hub stats before
curl -s http://localhost:8080/api/admin/hub-stats | jq '.throttle_percent'

# Wait 1 minute

# Get hub stats after
curl -s http://localhost:8080/api/admin/hub-stats | jq '{received, broadcast, throttle_percent}'
```

**Expected Output:**
```json
{
  "received": 15000,
  "broadcast": 12500,
  "throttle_percent": 16.7
}
```

**Pass Criteria:**
- [ ] Throttling active (> 10% reduction)
- [ ] Broadcast count < received count
- [ ] CPU usage reasonable despite high tick rate

---

## Integration Testing (20 minutes)

### Integration Test 1: FIX → Pipeline → Redis

**Objective:** Verify complete data flow from FIX to Redis

**Steps:**

1. Note current tick count:
```bash
EURUSD_BEFORE=$(redis-cli LLEN market_data:EURUSD)
echo "EURUSD ticks before: $EURUSD_BEFORE"
```

2. Wait 60 seconds:
```bash
sleep 60
```

3. Check tick count:
```bash
EURUSD_AFTER=$(redis-cli LLEN market_data:EURUSD)
DIFF=$((EURUSD_AFTER - EURUSD_BEFORE))
echo "New ticks: $DIFF (expected: > 100)"
```

4. Verify data integrity:
```bash
# Check latest tick
redis-cli LINDEX market_data:EURUSD 0 | jq '{symbol, bid, ask, spread, timestamp}'

# Compare with previous tick
redis-cli LINDEX market_data:EURUSD 1 | jq '{symbol, bid, ask, timestamp}'
```

**Pass Criteria:**
- [ ] DIFF > 100 (many new ticks in 60 seconds)
- [ ] Latest tick bid < ask
- [ ] Timestamps advancing (newest > oldest)
- [ ] Same symbol in all ticks

### Integration Test 2: Pipeline → WebSocket → Frontend

**Objective:** Verify data flows to frontend

**Steps:**

1. Connect WebSocket and save output:
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

wscat -c "ws://localhost:8080/ws?token=$TOKEN" | tee ws_flow.log &
WS_PID=$!
sleep 30
kill $WS_PID
```

2. Extract unique symbols:
```bash
cat ws_flow.log | jq -r '.symbol' | sort | uniq
```

3. Compare with Redis:
```bash
redis-cli KEYS market_data:* | sed 's/market_data://' | sort
```

**Pass Criteria:**
- [ ] Symbols in WebSocket match Redis
- [ ] Latest WebSocket price matches latest Redis price
- [ ] WebSocket symbols within 5 minutes of Redis timestamp

### Integration Test 3: End-to-End Load Test

**Objective:** Verify pipeline handles sustained load

**Duration:** 5 minutes

**Steps:**

1. Get baseline stats:
```bash
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data | {ticks_received, avg_latency_ms, dropped}'
```

2. Connect multiple WebSocket clients:
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"trader","password":"trader"}' | jq -r '.token')

for i in {1..5}; do
  wscat -c "ws://localhost:8080/ws?token=$TOKEN" > client_$i.log &
done
```

3. Wait 5 minutes and monitor:
```bash
# Watch stats every 30 seconds
for i in {1..10}; do
  echo "=== Check $i ==="
  curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data | {ticks_received, avg_latency_ms, dropped, clients_connected}'
  sleep 30
done
```

4. Kill WebSocket clients and check final stats:
```bash
pkill wscat
curl -s http://localhost:8080/api/admin/pipeline-stats | jq '.data'
```

**Pass Criteria:**
- [ ] Latency stays < 15ms under load
- [ ] Dropped ticks = 0
- [ ] Steady tick ingestion rate
- [ ] All clients remain connected
- [ ] Memory doesn't grow unbounded

---

## Frontend Verification (10 minutes)

### Frontend Test 1: Market Watch Display

**Steps:**

1. Open frontend: http://localhost:3000
2. Login with trader/trader
3. Navigate to Trading → Market Watch
4. Observe for 30 seconds:
   - [ ] Prices updating continuously
   - [ ] Bid/Ask showing correct values
   - [ ] Spread calculated correctly (< 5 pips for majors)
   - [ ] Timestamps updating
   - [ ] Price color changes (green up, red down)

### Frontend Test 2: Chart Display

**Steps:**

1. Navigate to Trading → Charts
2. Select 1m timeframe
3. Observe for 30 seconds:
   - [ ] Candles displaying
   - [ ] High ≥ Open, Close ≥ Low
   - [ ] Candle colors correct (green for up closes)
   - [ ] Time axis showing correct timestamps
   - [ ] Scroll/zoom working

4. Switch to 5m timeframe:
   - [ ] Candles update
   - [ ] 5m bar open = previous 1m bar close
   - [ ] Chart updates smoothly

### Frontend Test 3: Real-Time Updates

**Steps:**

1. Note EURUSD price
2. Wait 10 seconds
3. Verify price changed
4. Check timestamp updated
5. Check latency in browser console: F12 → Console
   - [ ] Look for [WS] messages
   - [ ] Timestamps show <1 second latency

---

## Troubleshooting Quick Reference

| Symptom | Likely Cause | Check |
|---------|--------------|-------|
| No FIX messages | LP offline | Ping LP, check proxy |
| High latency | CPU overload | Check `top`, reduce tick rate |
| WebSocket disconnects | Auth expired | Re-login, check token |
| Stale prices | No data flow | Check FIX logs |
| Memory growing | Leak in pipeline | Restart, monitor memory |
| Client not receiving data | Buffer overflow | Increase buffer sizes |

---

## Sign-Off Template

```
Pipeline Verification Report
Date: _______________
Tester: _______________

Basic Checks (5 min):
  [ ] FIX connected
  [ ] Market data flowing
  [ ] Pipeline processing
  [ ] Redis storing data
  [ ] WebSocket connected

Component Tests (15 min):
  [ ] FIX stability
  [ ] Pipeline latency
  [ ] OHLC generation
  [ ] Redis queries
  [ ] WebSocket broadcast

Integration Tests (20 min):
  [ ] FIX → Redis flow
  [ ] Pipeline → WebSocket flow
  [ ] Load test

Frontend Tests (10 min):
  [ ] Market Watch
  [ ] Charts
  [ ] Real-time updates

Overall Result: PASS / FAIL

Issues found: _______________________________________________________________

Recommendations: ____________________________________________________________

Signed: _________________________ Date: _______________
```
