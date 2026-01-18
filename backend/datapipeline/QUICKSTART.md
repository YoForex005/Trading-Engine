# Quick Start Guide - 5 Minutes to Running Pipeline

Get the data pipeline running in 5 minutes for immediate testing.

## Prerequisites Check (1 minute)

```bash
# Check Go version (need 1.21+)
go version

# Check if Redis is installed
which redis-server

# If Redis not installed:
# macOS: brew install redis
# Ubuntu: sudo apt install redis-server
```

## Step 1: Start Redis (1 minute)

```bash
# macOS
brew services start redis

# Linux
sudo systemctl start redis

# Or run in foreground for testing
redis-server

# Verify Redis is running
redis-cli ping
# Should return: PONG
```

## Step 2: Install Dependencies (1 minute)

```bash
cd /Users/epic1st/Documents/trading\ engine/backend
go get github.com/redis/go-redis/v9
go mod tidy
```

## Step 3: Run the Example (2 minutes)

```bash
# Run the standalone example
cd /Users/epic1st/Documents/trading\ engine/backend/examples
go run pipeline_integration_example.go
```

You should see:
```
=== Real-Time Data Pipeline Integration Example ===
[Pipeline] Initialized with 4 workers, tick buffer: 10000
[Pipeline] Starting market data pipeline...
[Ingester] Started with 4 workers
[OHLCEngine] Started with timeframes: M1, M5, M15, H1, H4, D1
[Distributor] Started quote distribution workers
[Storage] Storage manager started
[Monitor] Data quality monitoring started
[Pipeline] Market data pipeline started successfully
[HTTP] Server listening on :7999
```

## Step 4: Test the Pipeline (1 minute)

Open a new terminal and test the endpoints:

### Test Health
```bash
curl http://localhost:7999/api/pipeline/health
```

### Test Stats
```bash
curl http://localhost:7999/api/pipeline/stats | jq
```

### Simulate Tick Ingestion
```bash
# The example automatically bridges LP Manager quotes
# Just watch the logs for tick processing
```

### Test Latest Quote (after some ticks)
```bash
curl http://localhost:7999/api/quotes/latest/BTCUSD
```

### Test OHLC
```bash
curl "http://localhost:7999/api/ohlc/BTCUSD?timeframe=1m&limit=5"
```

## What's Happening?

The example creates:
1. ✅ Data pipeline with Redis storage
2. ✅ LP Manager for quote generation
3. ✅ WebSocket hub for distribution
4. ✅ Bridge between LP Manager and pipeline
5. ✅ HTTP API server with all endpoints
6. ✅ Monitoring and health checks

## Monitoring Output

You should see logs like:
```
[Bridge] Processed 10000 quotes from LP Manager
[Pipeline Stats] Received: 10000 | Processed: 9995 | Dropped: 5
[Pipeline Latency] Tick: 0.8ms | OHLC: 3.2ms | Distribution: 5.1ms
[OHLCEngine] Closing bar: BTCUSD M1 [12:34:00 - 12:34:59] O:45000 H:45100 L:44900 C:45050
```

## Test WebSocket Connection

```javascript
// In browser console or Node.js
const ws = new WebSocket('ws://localhost:7999/ws');
ws.onmessage = (event) => {
    console.log('Tick:', JSON.parse(event.data));
};
```

## Performance Check

```bash
# Check pipeline statistics
curl -s http://localhost:7999/api/pipeline/stats | jq '{
  throughput: .ticks_received,
  latency_ms: .avg_tick_latency_ms,
  drop_rate: ((.ticks_dropped / .ticks_received) * 100)
}'
```

Expected output:
```json
{
  "throughput": 10000,
  "latency_ms": 0.8,
  "drop_rate": 0.05
}
```

## Stop the Pipeline

Press `Ctrl+C` and you'll see graceful shutdown:
```
[Main] Shutdown signal received, gracefully stopping...
[Pipeline] Stopping market data pipeline...
[Pipeline] Pipeline stopped
[Main] Shutdown complete
```

## Next Steps

### Option 1: Run Benchmarks
```bash
cd /Users/epic1st/Documents/trading\ engine/backend/datapipeline
go test -v -bench=BenchmarkFullPipeline -benchtime=10s
```

### Option 2: Integrate with Your Server
Follow the [INTEGRATION_CHECKLIST.md](INTEGRATION_CHECKLIST.md)

### Option 3: Explore API
- Read [README.md](README.md) for full API documentation
- Check [api.go](api.go) for all available endpoints
- Review [SUMMARY.md](SUMMARY.md) for architecture details

## Troubleshooting

### Redis Connection Error
```
Failed to connect to Redis: connection refused
```
**Solution**: Start Redis with `brew services start redis` or `redis-server`

### Port Already in Use
```
bind: address already in use
```
**Solution**: Change port in example code or stop conflicting service

### No Quotes Received
```
[Pipeline Stats] Received: 0
```
**Solution**:
- Check if LP Manager is configured correctly
- Verify LP credentials in `data/lp_config.json`
- Enable debug logging

### High Latency
```
Avg Tick Latency: 50ms
```
**Solution**:
- Increase worker count in config
- Check Redis performance: `redis-cli --latency`
- Reduce other system load

## Configuration Tuning

### For High Throughput
```go
config.WorkerCount = 8
config.TickBufferSize = 50000
config.EnableDeduplication = false
```

### For Low Latency
```go
config.WorkerCount = 16
config.TickBufferSize = 5000
config.EnableDeduplication = true
```

### For Low Memory
```go
config.HotDataRetention = 500
config.WarmDataRetentionDays = 7
```

## Monitoring Dashboard

Create a simple monitoring script:

```bash
#!/bin/bash
# monitor.sh

watch -n 5 'curl -s http://localhost:7999/api/pipeline/stats | jq "{
  ticks: .ticks_received,
  processed: .ticks_processed,
  dropped: .ticks_dropped,
  latency_ms: .avg_tick_latency_ms,
  clients: .clients_connected
}"'
```

Make it executable and run:
```bash
chmod +x monitor.sh
./monitor.sh
```

## Success Criteria

Your pipeline is working correctly if you see:

- [x] Health endpoint returns `"overall_status":"healthy"`
- [x] Stats show increasing tick counts
- [x] Average latency < 10ms
- [x] Drop rate < 0.1%
- [x] OHLC bars are being generated
- [x] Quotes endpoint returns recent data
- [x] WebSocket broadcasts are working

## Getting Help

If you encounter issues not covered here:

1. Check [INSTALLATION.md](INSTALLATION.md) for detailed setup
2. Review [README.md](README.md) for architecture details
3. Read [INTEGRATION_CHECKLIST.md](INTEGRATION_CHECKLIST.md) for production setup
4. Check Redis logs: `redis-cli MONITOR`
5. Enable debug logging in the pipeline

## Performance Expectations

With the example configuration, you should see:

| Metric | Value |
|--------|-------|
| Throughput | 10,000+ ticks/sec |
| Latency | < 5ms average |
| Memory | ~200MB |
| CPU | ~20% (4 workers) |
| Drop Rate | < 0.1% |

For production loads (100k+ ticks/sec), see [README.md](README.md) for tuning.

---

**That's it! You now have a running real-time data pipeline.**

Next: Follow [INTEGRATION_CHECKLIST.md](INTEGRATION_CHECKLIST.md) to integrate with your main trading engine.
