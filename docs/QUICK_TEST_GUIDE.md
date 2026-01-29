# Quick Test Guide - Tick Data Flow

## 5-Minute System Check

### Prerequisites
- Backend server running on port 7999
- Terminal access

### Step 1: Check Server Health (30 seconds)

```bash
# Test server is running
curl http://localhost:7999/health

# Expected: "OK"
```

### Step 2: Verify FIX Connection (30 seconds)

```bash
# Check FIX gateway status
curl http://localhost:7999/admin/fix/status | jq

# Expected: YOFX1 and YOFX2 both "LOGGED_IN"
```

### Step 3: Check Tick Storage (1 minute)

```bash
# Count tick files
find backend/data/ticks -type f -name "*.json" | wc -l

# Expected: 50+ files

# Check storage size
du -sh backend/data/ticks

# Expected: 100+ MB

# List symbols with data
ls backend/data/ticks

# Expected: EURUSD, GBPUSD, AUDCAD, XAUUSD, etc.
```

### Step 4: Test REST API (1 minute)

```bash
# Get latest EURUSD ticks
curl "http://localhost:7999/ticks?symbol=EURUSD&limit=5" | jq

# Expected: Array of 5 ticks with bid/ask/spread

# Get OHLC bars
curl "http://localhost:7999/ohlc?symbol=EURUSD&timeframe=1m&limit=3" | jq

# Expected: Array of OHLC bars
```

### Step 5: Check Market Data Flow (1 minute)

```bash
# Check tick flow stats
curl http://localhost:7999/admin/fix/ticks | jq

# Expected:
# {
#   "totalTickCount": 1000+,
#   "symbolCount": 10+,
#   "latestTicks": { ... }
# }
```

### Step 6: Verify Symbol Coverage (1 minute)

```bash
# Get available symbols
curl http://localhost:7999/api/symbols/available | jq | head -20

# Expected: 60+ symbols

# Get subscribed symbols
curl http://localhost:7999/api/symbols/subscribed | jq

# Expected: 30+ symbols
```

---

## Quick Status Check

### ✅ All Tests Pass
**System Status: OPERATIONAL**
- Market data is flowing
- Ticks are being stored
- APIs are responsive
- Frontend can connect

### ⚠️ Some Tests Fail
**System Status: PARTIAL**
- Check FIX connection (YOFX1/YOFX2)
- Verify market hours (data only during trading hours)
- Check logs for errors

### ❌ Most Tests Fail
**System Status: CRITICAL**
- Server not running or wrong port
- FIX gateway not configured
- Database/storage issues

---

## Common Issues & Fixes

### Issue: No tick files found

**Possible Causes:**
1. System just started (wait 5 minutes)
2. FIX not connected (check `/admin/fix/status`)
3. Market closed (check trading hours)

**Fix:**
```bash
# Check FIX status
curl http://localhost:7999/admin/fix/status

# Manually subscribe to EURUSD
curl -X POST http://localhost:7999/api/symbols/subscribe \
  -H "Content-Type: application/json" \
  -d '{"symbol": "EURUSD"}'
```

### Issue: API returns empty array

**Possible Causes:**
1. Symbol not subscribed
2. No recent data
3. Ring buffer not yet populated

**Fix:**
```bash
# Check subscribed symbols
curl http://localhost:7999/api/symbols/subscribed

# Subscribe to symbol
curl -X POST http://localhost:7999/api/symbols/subscribe \
  -H "Content-Type: application/json" \
  -d '{"symbol": "EURUSD"}'

# Wait 30 seconds and retry
```

### Issue: FIX status shows "DISCONNECTED"

**Possible Causes:**
1. Wrong credentials
2. Network/firewall issues
3. LP server down

**Fix:**
```bash
# Check FIX config
cat backend/fix/config/sessions.json

# Reconnect manually
curl -X POST http://localhost:7999/admin/fix/connect \
  -H "Content-Type: application/json" \
  -d '{"sessionId": "YOFX2"}'
```

---

## Performance Benchmarks

### Expected Performance

| Metric | Expected | Acceptable | Poor |
|--------|----------|------------|------|
| Ticks/sec | 100-200 | 50-100 | <50 |
| API latency (/ticks) | <100ms | <500ms | >1s |
| API latency (/ohlc) | <50ms | <200ms | >500ms |
| Symbols with data | 30+ | 10+ | <10 |
| Storage size (per day) | 5-10MB | 2-5MB | <1MB |

### Test API Performance

```bash
# Measure API latency
time curl -s "http://localhost:7999/ticks?symbol=EURUSD&limit=100" > /dev/null

# Expected: <0.5 seconds

# Test under load (100 requests)
for i in {1..100}; do
  curl -s "http://localhost:7999/ticks?symbol=EURUSD&limit=10" > /dev/null &
done
wait

# Expected: All complete within 10 seconds
```

---

## WebSocket Quick Test

### Using wscat (Node.js)

```bash
# Install wscat
npm install -g wscat

# Connect to WebSocket
wscat -c ws://localhost:7999/ws

# Expected: Real-time tick messages every 0.5-2 seconds
# {"type":"tick","symbol":"EURUSD","bid":1.08234,"ask":1.08244,...}
```

### Using JavaScript (Browser Console)

```javascript
const ws = new WebSocket("ws://localhost:7999/ws");

ws.onopen = () => console.log("Connected");
ws.onmessage = (e) => console.log(JSON.parse(e.data));
ws.onerror = (e) => console.error("Error:", e);

// Expected: Tick messages logged to console
```

---

## Data Quality Checks

### Verify Tick Data Integrity

```bash
# Check recent EURUSD data
tail -100 backend/data/ticks/EURUSD/$(date +%Y-%m-%d).json | jq

# Verify:
# 1. Valid JSON format ✅
# 2. Bid < Ask ✅
# 3. Spread = Ask - Bid ✅
# 4. Timestamps are recent ✅
# 5. LP field is present ✅
```

### Check for Data Gaps

```bash
# Count ticks per hour for EURUSD today
cat backend/data/ticks/EURUSD/$(date +%Y-%m-%d).json | \
  jq -r '.[] | .timestamp' | \
  cut -d'T' -f2 | cut -d':' -f1 | \
  sort | uniq -c

# Expected: Relatively even distribution during market hours
```

---

## Troubleshooting Checklist

### ✅ Pre-Test Checklist

- [ ] Server is running (`curl http://localhost:7999/health`)
- [ ] Port 7999 is not blocked by firewall
- [ ] Enough disk space (>10GB free)
- [ ] FIX credentials are configured
- [ ] Market is open (for live data)

### ✅ Data Flow Checklist

- [ ] FIX sessions are LOGGED_IN
- [ ] Symbols are subscribed (30+)
- [ ] Tick files are being created
- [ ] Storage is growing over time
- [ ] API returns recent data
- [ ] WebSocket broadcasts ticks

### ✅ Performance Checklist

- [ ] API latency <500ms
- [ ] Memory usage <500MB
- [ ] CPU usage <50%
- [ ] Disk writes are async
- [ ] No error logs

---

## Quick Fixes

### Restart Everything

```bash
# Stop server
pkill -f "go run.*server"

# Clear logs (optional)
rm backend/*.log

# Restart server
cd backend/cmd/server
go run main.go

# Wait 1 minute for FIX connection
sleep 60

# Re-run tests
```

### Reset Tick Storage

```bash
# CAUTION: This deletes all tick data!

# Backup first
tar -czf tick_backup_$(date +%Y%m%d).tar.gz backend/data/ticks

# Clear tick data
rm -rf backend/data/ticks/*

# Restart server to rebuild
```

### Force Symbol Subscription

```bash
# Subscribe to all major symbols
for symbol in EURUSD GBPUSD USDJPY AUDUSD USDCAD USDCHF NZDUSD; do
  curl -X POST http://localhost:7999/api/symbols/subscribe \
    -H "Content-Type: application/json" \
    -d "{\"symbol\": \"$symbol\"}"
  sleep 1
done

# Verify subscriptions
curl http://localhost:7999/api/symbols/subscribed | jq
```

---

## Success Criteria

### System is Working ✅ if:

1. ✅ FIX sessions are LOGGED_IN
2. ✅ 50+ tick files exist
3. ✅ Storage is 100+ MB
4. ✅ API returns recent ticks (<1 hour old)
5. ✅ WebSocket broadcasts real-time data
6. ✅ 30+ symbols are subscribed

### System Needs Attention ⚠️ if:

1. ⚠️ Only 1-2 FIX sessions connected
2. ⚠️ Less than 10 tick files
3. ⚠️ Storage is <50MB
4. ⚠️ API returns old data (>1 day)
5. ⚠️ WebSocket connects but no data
6. ⚠️ Less than 10 symbols subscribed

### System is Broken ❌ if:

1. ❌ No FIX sessions connected
2. ❌ No tick files
3. ❌ Storage is empty
4. ❌ API returns errors or empty
5. ❌ WebSocket won't connect
6. ❌ No symbols subscribed

---

## Next Steps After Testing

### If All Tests Pass ✅

1. Review full validation report (`TICK_DATA_E2E_VALIDATION_REPORT.md`)
2. Implement recommended improvements (SQLite, compression, retention)
3. Set up monitoring and alerts
4. Schedule regular backups

### If Some Tests Fail ⚠️

1. Check specific failure logs
2. Verify FIX configuration
3. Test during market hours
4. Check network connectivity
5. Review error logs in `backend/*.log`

### If Most Tests Fail ❌

1. Check server logs for errors
2. Verify environment configuration
3. Test FIX credentials manually
4. Ensure ports are not blocked
5. Contact system administrator

---

## Additional Resources

- **Full Validation Report:** `docs/TICK_DATA_E2E_VALIDATION_REPORT.md`
- **Test Summary:** `docs/TICK_DATA_TEST_SUMMARY.md`
- **Automated Test Script:** `scripts/test_tick_data_flow.sh`

---

**Last Updated:** January 20, 2026
