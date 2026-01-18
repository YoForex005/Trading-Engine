# Real-Time Market Data Pipeline

High-performance, scalable data pipeline for market data ingestion, processing, and distribution.

## Features

### 1. Market Data Ingestion
- ✅ Multi-source ingestion (FIX, WebSocket, REST)
- ✅ Tick normalization across providers
- ✅ Timestamp format handling (Unix, ISO8601)
- ✅ Tick deduplication
- ✅ Out-of-order tick handling
- ✅ Data validation and sanity checks

### 2. OHLC Aggregation Engine
- ✅ Real-time OHLC from ticks
- ✅ Multiple timeframes (M1, M5, M15, H1, H4, D1)
- ✅ Aligned bars (exact minute/hour boundaries)
- ✅ Volume aggregation
- ✅ Session boundary handling
- ✅ Backfill support

### 3. Quote Distribution
- ✅ Redis pub/sub for horizontal scaling
- ✅ WebSocket broadcasting
- ✅ Per-instrument subscriptions
- ✅ Rate limiting per client (100 msgs/sec)
- ✅ Quote throttling (price change threshold)
- ✅ Snapshot + incremental updates

### 4. Data Storage Strategy
- ✅ Hot: Redis (last 1000 ticks per symbol)
- ✅ Warm: Redis sorted sets (30 days)
- ⏳ Cold: TimescaleDB/S3 (future)
- ✅ OHLC: Redis sorted sets
- ✅ Fast retrieval for charts

### 5. WebSocket Architecture
- ✅ Horizontal scaling ready
- ✅ Redis pub/sub message distribution
- ✅ Connection state management
- ✅ Heartbeat/ping-pong
- ✅ Graceful degradation
- ⏳ Binary protocol (MessagePack/Protobuf) - future

### 6. Market Data API
- ✅ `GET /api/quotes/latest/:symbol` - Latest quote
- ✅ `GET /api/quotes/stream` - SSE stream
- ✅ `GET /api/quotes/history/:symbol` - Recent quotes
- ✅ `GET /api/ohlc/:symbol` - Historical OHLC
- ✅ `GET /api/ohlc/latest/:symbol` - Active bar
- ✅ `GET /api/ticks/:symbol` - Recent ticks (compat)
- ✅ WebSocket protocol support

### 7. Data Quality & Monitoring
- ✅ Stale quote detection
- ✅ Abnormal price spike detection
- ✅ Feed health monitoring
- ✅ Alert system
- ✅ Tick-to-quote latency tracking
- ✅ Distribution latency monitoring

### 8. Performance Optimization
- ✅ In-memory quote cache
- ✅ Batch processing
- ✅ Worker pool architecture
- ✅ Non-blocking channels
- ✅ Efficient serialization
- ✅ Connection pooling ready

### 9. Failover & High Availability
- ✅ Multiple data source support
- ✅ Redis Sentinel ready
- ✅ Auto-reconnection
- ✅ Graceful degradation
- ⏳ Circuit breakers - future

### 10. Admin Features
- ✅ Feed status dashboard
- ✅ Pipeline statistics
- ✅ Health checks
- ✅ Alert monitoring
- ✅ Storage cleanup
- ✅ Performance metrics

## Architecture

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   LP Data   │────▶│   Ingester   │────▶│  OHLC       │
│  Sources    │     │  (Normalize) │     │  Engine     │
└─────────────┘     └──────────────┘     └─────────────┘
                            │                    │
                            ▼                    ▼
                    ┌──────────────┐     ┌─────────────┐
                    │ Distributor  │     │  Storage    │
                    │ (Redis Pub)  │     │  (Redis)    │
                    └──────────────┘     └─────────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │  WebSocket   │
                    │   Clients    │
                    └──────────────┘
```

## Usage

### Initialize Pipeline

```go
package main

import (
    "github.com/epic1st/rtx/backend/datapipeline"
    "log"
)

func main() {
    // Create configuration
    config := datapipeline.DefaultPipelineConfig()
    config.RedisAddr = "localhost:6379"
    config.WorkerCount = 4

    // Create pipeline
    pipeline, err := datapipeline.NewPipeline(config)
    if err != nil {
        log.Fatal(err)
    }

    // Start pipeline
    if err := pipeline.Start(); err != nil {
        log.Fatal(err)
    }

    log.Println("Pipeline started successfully")
}
```

### Ingest Data

```go
// From LP Manager
rawTick := &datapipeline.RawTick{
    Source:    "OANDA",
    Symbol:    "EURUSD",
    Bid:       1.08450,
    Ask:       1.08452,
    Timestamp: time.Now().Unix(),
}

pipeline.IngestTick(rawTick)
```

### Integration with Existing Code

```go
// Create adapter
adapter := datapipeline.NewIntegrationAdapter(pipeline)

// Process existing MarketTick format
adapter.ProcessMarketTick(&datapipeline.MarketTickCompat{
    Symbol:    "BTCUSD",
    Bid:       45000.50,
    Ask:       45001.00,
    LP:        "Binance",
    Timestamp: time.Now().Unix(),
})

// Start monitoring
adapter.StartMonitoring()
```

### API Integration

```go
// Create API handler
apiHandler := datapipeline.NewAPIHandler(pipeline)

// Register routes
mux := http.NewServeMux()
apiHandler.RegisterRoutes(mux)

// Start server
http.ListenAndServe(":8080", mux)
```

## API Endpoints

### Quotes

```bash
# Get latest quote
curl http://localhost:7999/api/quotes/latest/EURUSD

# Get quote history
curl http://localhost:7999/api/quotes/history/EURUSD?limit=100

# Stream quotes (SSE)
curl http://localhost:7999/api/quotes/stream?symbol=EURUSD&symbol=GBPUSD
```

### OHLC

```bash
# Get OHLC bars
curl http://localhost:7999/api/ohlc/EURUSD?timeframe=1m&limit=100

# Get active bar
curl http://localhost:7999/api/ohlc/latest/EURUSD?timeframe=1m
```

### Pipeline Management

```bash
# Get statistics
curl http://localhost:7999/api/pipeline/stats

# Health check
curl http://localhost:7999/api/pipeline/health

# Feed health
curl http://localhost:7999/api/pipeline/feed-health

# Get alerts
curl http://localhost:7999/api/pipeline/alerts?limit=50
```

### Admin

```bash
# Trigger cleanup
curl -X POST http://localhost:7999/admin/pipeline/cleanup
```

## Performance Targets

- **Tick Ingestion**: < 1ms latency
- **OHLC Aggregation**: < 5ms latency
- **Distribution**: < 10ms latency
- **Total Tick-to-Client**: < 100ms
- **Throughput**: 100,000+ ticks/second
- **Concurrent Clients**: 10,000+

## Configuration

```go
type PipelineConfig struct {
    // Redis
    RedisAddr              string
    RedisPassword          string
    RedisDB                int

    // Performance
    TickBufferSize         int    // 10000
    OHLCBufferSize         int    // 1000
    DistributionBufferSize int    // 5000
    WorkerCount            int    // 4

    // Data Quality
    EnableDeduplication    bool   // true
    EnableOutOfOrderCheck  bool   // true
    MaxTickAgeSeconds      int    // 60
    PriceSanityThreshold   float64 // 0.10 (10%)

    // Storage
    HotDataRetention       int    // 1000 ticks
    WarmDataRetentionDays  int    // 30 days

    // Monitoring
    EnableHealthChecks     bool          // true
    HealthCheckInterval    time.Duration // 30s
    StaleQuoteThreshold    time.Duration // 10s
}
```

## Monitoring

### Pipeline Statistics

```json
{
  "ticks_received": 1000000,
  "ticks_processed": 999500,
  "ticks_dropped": 500,
  "ticks_duplicate": 100,
  "ticks_out_of_order": 50,
  "ticks_invalid": 20,
  "ohlc_bars_generated": 5000,
  "quotes_distributed": 999500,
  "clients_connected": 150,
  "avg_tick_latency_ms": 0.8,
  "avg_ohlc_latency_ms": 3.2,
  "avg_distribution_latency_ms": 5.1,
  "stale_feeds_detected": 2,
  "abnormal_spikes_detected": 5
}
```

### Feed Health

```json
{
  "EURUSD": {
    "symbol": "EURUSD",
    "last_tick_time": "2026-01-18T12:00:00Z",
    "tick_count": 10000,
    "is_stale": false,
    "stale_seconds": 0.5,
    "status": "healthy"
  }
}
```

## Redis Data Structure

### Ticks (Sorted Sets)
```
Key: ticks:{SYMBOL}
Score: Unix timestamp
Member: JSON-serialized NormalizedTick
```

### Latest Tick (String)
```
Key: tick:latest:{SYMBOL}
Value: JSON-serialized NormalizedTick
TTL: 1 hour
```

### OHLC (Sorted Sets)
```
Key: ohlc:{SYMBOL}:{TIMEFRAME}
Score: Bar open time (Unix timestamp)
Member: JSON-serialized OHLCBar
```

### Pub/Sub Channels
```
quotes                    - All quotes
quotes:{SYMBOL}           - Symbol-specific quotes
ohlc                      - All OHLC bars
ohlc:{SYMBOL}:{TIMEFRAME} - Specific OHLC updates
```

## Future Enhancements

1. **TimescaleDB Integration** - Long-term tick storage
2. **Binary Protocol** - MessagePack/Protobuf for WebSocket
3. **Circuit Breakers** - Auto-failover on source failures
4. **Compression** - Tick data compression
5. **Multi-Region** - Geographical distribution
6. **Machine Learning** - Anomaly detection
7. **Replay Mode** - Historical data replay
8. **Backpressure** - Intelligent flow control

## License

Proprietary - RTX Trading Platform
