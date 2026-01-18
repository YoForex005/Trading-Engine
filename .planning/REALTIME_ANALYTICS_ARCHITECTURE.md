# Real-Time Analytics Dashboard Architecture

## Executive Summary

This document provides comprehensive guidance for building scalable real-time analytics dashboards. It covers backend architecture, real-time data flows, performance optimization, high availability, and observability.

---

## 1. Backend Architecture

### 1.1 Time-Series Database Selection

#### InfluxDB
**Best for:** High-cardinality metrics, IoT/monitoring data, stock tickers

**Advantages:**
- Purpose-built for time-series with automatic retention policies
- Excellent compression (100:1 ratio typical)
- Fast downsampling and aggregation
- Built-in continuous aggregates
- Rich ecosystem (Grafana integration, Telegraf collectors)

**Disadvantages:**
- Can be expensive at scale (licenses for clustering)
- Complex query language (Flux)
- Retention policies can be tricky to manage

**Ideal for:** Trading metrics, tick data, minute-level OHLC data

**Example use case:**
```
Real-time trading data → InfluxDB → Dashboard updates every 100ms
Retention: 7 days raw, 1 year downsampled to 1-hour buckets
```

---

#### TimescaleDB
**Best for:** Complex queries, large analytical workloads, existing Postgres deployments

**Advantages:**
- Full SQL support (familiar to most teams)
- Automatic time-based partitioning
- Excellent for complex analytical queries
- Compression built-in (better than vanilla Postgres)
- Works with existing Postgres tools/ecosystem

**Disadvantages:**
- Slightly slower than InfluxDB for simple time-series reads
- Requires more disk space than InfluxDB
- Hypertables need careful planning

**Ideal for:** Portfolio analytics, complex report generation, historical backtesting

**Example use case:**
```
SELECT
  time_bucket('1 minute', timestamp) as minute,
  symbol,
  first(price) as open,
  max(price) as high,
  min(price) as low,
  last(price) as close
FROM trades
WHERE timestamp > now() - INTERVAL '1 day'
GROUP BY minute, symbol
ORDER BY minute DESC;
```

---

#### ClickHouse
**Best for:** Real-time analytics at massive scale, columnar analytics

**Advantages:**
- Extreme query performance (10-100x faster than alternatives)
- Excellent compression
- Distributed architecture from day 1
- Real-time inserts and queries simultaneously
- Cost-effective at scale (can handle billions of rows)

**Disadvantages:**
- Steeper learning curve
- Limited transaction support (insert-only paradigm)
- No secondary indexes
- Requires careful schema design

**Ideal for:** Real-time dashboards with 100k+ events/sec, backtesting analysis

**Example use case:**
```
SELECT
  symbol,
  countIf(price > 100) as high_trades,
  countIf(volume > 50000) as high_volume,
  sum(volume) as total_volume,
  avg(price) as avg_price
FROM trades
WHERE timestamp >= subtractMinutes(now(), 5)
GROUP BY symbol;
```

---

### 1.2 Recommended Technology Stack

```
┌─────────────────────────────────────────────────────┐
│                  Real-Time Data Sources              │
│         (Exchanges, Market Data Feeds, APIs)         │
└─────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────┐
│              Message Queue (Kafka/Pulsar)            │
│  - Event streaming with partition key strategy       │
│  - Topic: trades, quotes, events                     │
│  - Retention: 24 hours                               │
└─────────────────────────────────────────────────────┘
        ↓                    ↓                    ↓
   ┌────────┐         ┌────────┐         ┌────────┐
   │ Stream │         │ Stream │         │ Stream │
   │Process │         │Process │         │Process │
   │   1    │         │   2    │         │   3    │
   └────────┘         └────────┘         └────────┘
        ↓                    ↓                    ↓
    ┌──────────────────────────────────────────────┐
    │      Time-Series Database Cluster            │
    │  ┌──────────────────────────────────────┐   │
    │  │ ClickHouse (Primary) - Raw tick data  │   │
    │  │ Sharding by symbol                    │   │
    │  └──────────────────────────────────────┘   │
    │  ┌──────────────────────────────────────┐   │
    │  │ TimescaleDB (Analytics) - Queries    │   │
    │  │ Complex aggregations                  │   │
    │  └──────────────────────────────────────┘   │
    └──────────────────────────────────────────────┘
        ↓                              ↓
   ┌─────────────────┐         ┌────────────────┐
   │ Redis Cache     │         │ Materialized   │
   │ - Hot data      │         │ Views (Agg)    │
   │ - Counters      │         │ - OHLC bars    │
   │ - User state    │         │ - Statistics   │
   └─────────────────┘         └────────────────┘
        ↓                              ↓
    ┌──────────────────────────────────────────────┐
    │            REST/WebSocket API Layer           │
    │  - Real-time endpoint (WebSocket)            │
    │  - Historical endpoint (REST)                 │
    │  - Aggregation endpoint (REST)               │
    └──────────────────────────────────────────────┘
        ↓
┌──────────────────────────────────────────────────────┐
│          Client Applications                         │
│  - Web Dashboard (React)                            │
│  - Mobile App (React Native)                        │
│  - Desktop Client (Electron)                        │
└──────────────────────────────────────────────────────┘
```

---

### 1.3 Data Aggregation Pipelines

#### Stream Processing with Kafka Streams / Flink

```
Raw Tick Data Stream
    ↓
┌─────────────────────────────────────────┐
│ Windowed Aggregations                   │
│ - 1-minute OHLCV bars                   │
│ - 5-minute volume profiles               │
│ - Rolling 30-second averages             │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Materialized Views Storage               │
│ - Kafka topic: aggregates-1m             │
│ - ClickHouse materialized view           │
│ - Redis cache layer                      │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│ Real-Time Dashboards                    │
│ Subscribe to aggregates topic            │
│ Push updates via WebSocket               │
└─────────────────────────────────────────┘
```

**Implementation Example (Kafka Streams):**

```java
KStream<String, Trade> trades = builder.stream("trades-topic");

KStream<Windowed<String>, OHLCV> ohlcv = trades
    .map((k, v) -> new KeyValue<>(v.getSymbol(), v))
    .groupByKey()
    .windowedBy(TimeWindows.of(Duration.ofMinutes(1)))
    .aggregate(
        OHLCV::new,
        (key, trade, ohlcv) -> ohlcv.update(trade),
        Materialized
            .as(Stores.inMemoryWindowStore(
                "ohlcv-store",
                Duration.ofMinutes(5),
                Duration.ofMinutes(1),
                true))
    )
    .toStream()
    .map((windowed, ohlcv) ->
        new KeyValue<>(windowed.key(), ohlcv));

ohlcv.to("aggregates-1m");
```

---

### 1.4 Caching Strategy

#### Multi-Layer Cache

**Layer 1: Database Query Cache (Redis)**
```
┌─────────────────────────────────┐
│     Redis Cluster               │
│  ┌────────────────────────────┐ │
│  │ Hot Keys Cache             │ │
│  │ - Last price by symbol     │ │
│  │ - TTL: 5 seconds          │ │
│  │ - Size: < 1GB             │ │
│  └────────────────────────────┘ │
│  ┌────────────────────────────┐ │
│  │ Aggregation Cache          │ │
│  │ - Minute OHLC bars         │ │
│  │ - TTL: 1 minute           │ │
│  │ - Size: 10-50GB           │ │
│  └────────────────────────────┘ │
│  ┌────────────────────────────┐ │
│  │ User Session Data          │ │
│  │ - Portfolio positions      │ │
│  │ - Watchlists              │ │
│  │ - TTL: 24 hours           │ │
│  └────────────────────────────┘ │
└─────────────────────────────────┘
```

**Cache Patterns:**

1. **Write-Through** (for critical data)
   - Write to DB and cache simultaneously
   - Used for: User orders, account balances

2. **Write-Behind** (for high-volume data)
   - Write to cache, flush to DB asynchronously
   - Used for: Trade ticks, quotes

3. **Cache-Aside** (for read-heavy data)
   - Check cache first, fall back to DB
   - Used for: Historical data, reference data

**TTL Strategy:**
```
Real-time last price: 5 seconds
OHLC bars: 1 minute (auto-refresh when bar closes)
5-minute candles: 5 minutes
Daily closes: 24 hours
Reference data: 7 days
User data: 24 hours
```

---

## 2. Real-Time Data Flow Architecture

### 2.1 Protocol Selection: WebSocket vs SSE

#### WebSocket
**Best for:** Bidirectional real-time communication

**Advantages:**
- Full-duplex communication (client→server + server→client)
- Lower latency than polling
- Lower bandwidth than polling
- Native browser support

**Disadvantages:**
- Stateful connections (harder to load balance)
- Requires server-side connection management
- No built-in reconnection logic

**Ideal for:** Trading orders, live quotes, interactive commands

---

#### Server-Sent Events (SSE)
**Best for:** Server→client real-time updates (read-heavy)

**Advantages:**
- Automatic reconnection
- Simple server-side implementation
- Works over HTTP/1.1
- Easier load balancing
- Built-in event IDs for deduplication

**Disadvantages:**
- One-way communication (server only)
- Text-based (less efficient for binary data)
- No compression standard

**Ideal for:** Real-time price updates, market news, status feeds

---

#### Hybrid Approach (Recommended for Trading)

```
WebSocket → Real-time prices, live orders, two-way commands
SSE → Market news, system alerts, broadcast notifications
REST → Historical data, reports, one-off queries
```

**Architecture:**
```
┌────────────────────────────────┐
│        Client (Browser)         │
│ ┌─────────────────────────────┐ │
│ │ WebSocket Manager           │ │
│ │ - Prices, orders            │ │
│ │ - Reconnection logic        │ │
│ └─────────────────────────────┘ │
│ ┌─────────────────────────────┐ │
│ │ SSE Listener                │ │
│ │ - News, alerts              │ │
│ │ - Auto-reconnect            │ │
│ └─────────────────────────────┘ │
│ ┌─────────────────────────────┐ │
│ │ REST API Client             │ │
│ │ - Historical queries        │ │
│ │ - Background fetches        │ │
│ └─────────────────────────────┘ │
└────────────────────────────────┘
        ↓            ↓            ↓
    WebSocket     SSE         REST API
        ↓            ↓            ↓
┌────────────────────────────────────────────┐
│         API Gateway / Load Balancer        │
│  - Sticky sessions for WebSocket           │
│  - Round-robin for SSE                     │
│  - Standard LB for REST                    │
└────────────────────────────────────────────┘
```

---

### 2.2 Message Compression

#### Delta Encoding Strategy

**Problem:** Sending full price objects every 100ms = lots of bandwidth

**Solution:** Send only changes
```json
// Full update (initial subscription)
{
  "type": "snapshot",
  "symbol": "AAPL",
  "price": 150.25,
  "volume": 1000000,
  "bid": 150.24,
  "ask": 150.26,
  "timestamp": 1234567890
}

// Delta updates (every tick)
{
  "type": "delta",
  "symbol": "AAPL",
  "price": 150.26,  // only changed fields
  "bid": 150.25,
  "timestamp": 1234567891
}
```

**Compression Ratio:** ~10:1 vs full snapshots

---

#### Binary Encoding (MessagePack/Protobuf)

**JSON Example:**
```
{"symbol":"AAPL","price":150.25,"volume":1000000} = 48 bytes
```

**MessagePack:**
```
Same data = 16 bytes (3.3x compression)
```

**Protobuf:**
```
Same data = 12 bytes (4x compression)
```

**Implementation:**
```typescript
// Protobuf definition
message PriceUpdate {
  string symbol = 1;
  double price = 2;
  int64 volume = 3;
  int64 timestamp = 4;

  enum UpdateType {
    SNAPSHOT = 0;
    DELTA = 1;
  }
  UpdateType type = 5;
}

// Compressed message size: ~20 bytes vs 100+ bytes JSON
```

---

### 2.3 Client-Side Data Buffering

**Problem:** WebSocket updates coming at 1000+ msg/sec might overwhelm browser

**Solution:** Batch updates on client

```typescript
class RealTimeDataBuffer {
  private buffer: Map<string, PriceUpdate> = new Map();
  private flushInterval = 100; // ms

  onPriceUpdate(update: PriceUpdate) {
    // Store update
    this.buffer.set(update.symbol, update);

    // Don't update DOM yet
    // Wait for batch flush
  }

  flush() {
    // Apply all buffered updates to DOM at once
    // Efficient: 1 paint operation instead of 1000
    const updates = Array.from(this.buffer.values());
    updatePricesInBatch(updates);
    this.buffer.clear();
  }
}

// Flush every 100ms (10 FPS visual updates)
setInterval(() => buffer.flush(), 100);
```

**Benefits:**
- Reduces DOM reflows from 1000/sec to 10/sec
- Better browser performance
- Smoother animations

---

## 3. Performance Optimization

### 3.1 Database Indexing for Time-Series

#### ClickHouse Index Strategy

```sql
CREATE TABLE trades (
    symbol String,
    price Float64,
    volume Int64,
    timestamp DateTime,
    bid_price Float64,
    ask_price Float64
) ENGINE = MergeTree()
ORDER BY (symbol, timestamp)  -- Primary key
INDEX idx_price (price) TYPE minmax GRANULARITY 8192
INDEX idx_volume (volume) TYPE set(1000) GRANULARITY 8192
TTL timestamp + INTERVAL 7 DAY;
```

**Index Considerations:**
- **ORDER BY (symbol, timestamp)**: Primary index for most queries
- **Price/Volume indexes**: Optional, for range queries
- **TTL**: Auto-delete old data (7 days)

---

#### TimescaleDB Index Strategy

```sql
CREATE TABLE trades (
    time TIMESTAMPTZ NOT NULL,
    symbol TEXT NOT NULL,
    price FLOAT8 NOT NULL,
    volume INT NOT NULL
);

-- Create hypertable (auto-partitions by time)
SELECT create_hypertable('trades', 'time');

-- Create indexes
CREATE INDEX ON trades (symbol, time DESC)
    WHERE time > now() - INTERVAL '7 days';

CREATE INDEX ON trades (symbol, price)
    WHERE time > now() - INTERVAL '7 days';

-- Compression
ALTER TABLE trades SET (
    timescaledb.compress,
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('trades', INTERVAL '1 hour');
```

---

### 3.2 Query Optimization

#### Problem: Complex real-time aggregations are slow

#### Solution: Materialized Views

**ClickHouse Materialized View:**
```sql
CREATE TABLE ohlc_1m (
    symbol String,
    time DateTime,
    open Float64,
    high Float64,
    low Float64,
    close Float64,
    volume Int64
) ENGINE = MergeTree()
ORDER BY (symbol, time);

CREATE MATERIALIZED VIEW ohlc_1m_mv
TO ohlc_1m AS
SELECT
    symbol,
    toStartOfMinute(timestamp) as time,
    first(price) as open,
    max(price) as high,
    min(price) as low,
    last(price) as close,
    sum(volume) as volume
FROM trades
GROUP BY symbol, time;

-- Now this query is instant (pre-computed):
SELECT * FROM ohlc_1m WHERE symbol = 'AAPL'
ORDER BY time DESC LIMIT 100;
```

---

### 3.3 Pre-Computed Metrics vs On-Demand

#### Pre-Computed Strategy (Recommended)

**When to use:**
- Frequently accessed aggregations (OHLC bars)
- Expensive computations (volume-weighted averages)
- Real-time dashboard updates

**How it works:**
```
Tick stream → Kafka → Stream processor → Materialized view
                                        ↓
                                   Sub-second queries
```

**Storage:** ~10-20GB per day (vs 100GB+ raw ticks)

---

#### On-Demand Strategy

**When to use:**
- Ad-hoc historical queries
- Infrequently accessed metrics
- Custom user-defined aggregations

**Query pattern:**
```sql
-- User wants to see 15-minute bars for past 3 months
SELECT
    time_bucket('15 minutes', timestamp) as time,
    symbol,
    FIRST(price) as open,
    MAX(price) as high,
    MIN(price) as low,
    LAST(price) as close
FROM trades
WHERE timestamp > now() - INTERVAL '3 months'
    AND symbol = 'AAPL'
GROUP BY time, symbol
ORDER BY time DESC;
```

---

### 3.4 Horizontal Scaling

#### ClickHouse Distributed Architecture

```
┌────────────────────────────────────┐
│    Load Balancer (Round-robin)      │
└────────────────────────────────────┘
    ↓            ↓            ↓
┌──────────┐ ┌──────────┐ ┌──────────┐
│ Node 1   │ │ Node 2   │ │ Node 3   │
│ Shard: A │ │ Shard: B │ │ Shard: C │
│ Replica1 │ │ Replica1 │ │ Replica1 │
└──────────┘ └──────────┘ └──────────┘
    ↓            ↓            ↓
┌──────────┐ ┌──────────┐ ┌──────────┐
│ Node 1   │ │ Node 2   │ │ Node 3   │
│ Shard: A │ │ Shard: B │ │ Shard: C │
│ Replica2 │ │ Replica2 │ │ Replica2 │
└──────────┘ └──────────┘ └──────────┘
```

**Sharding Strategy:**
```
Partition key: hash(symbol) % 3

AAPL → Node 1 (Shard A)
GOOGL → Node 2 (Shard B)
MSFT → Node 3 (Shard C)
NVDA → Node 1 (Shard A)
...
```

**Benefits:**
- Linear scaling: 3 nodes = 3x throughput
- Fault tolerance: 1 replica per shard
- Query parallelization across shards

---

## 4. High Availability

### 4.1 Load Balancing WebSocket Connections

#### Problem: WebSockets are stateful (can't just round-robin)

#### Solution: Sticky Sessions + Health Checks

```
┌────────────────────────────────────┐
│    Load Balancer (HAProxy)          │
│  - Sticky sessions by IP            │
│  - Health check every 5s             │
│  - Auto-remove unhealthy nodes      │
└────────────────────────────────────┘
    ↓            ↓            ↓
┌──────────┐ ┌──────────┐ ┌──────────┐
│ WS Node1 │ │ WS Node2 │ │ WS Node3 │
│ 1000     │ │ 1000     │ │ 1000     │
│ conns    │ │ conns    │ │ conns    │
└──────────┘ └──────────┘ └──────────┘
```

**HAProxy Configuration:**
```
listen websocket
    bind 0.0.0.0:8080
    mode tcp
    balance source  # Sticky sessions

    server node1 10.0.1.1:8080 check inter 5s
    server node2 10.0.1.2:8080 check inter 5s
    server node3 10.0.1.3:8080 check inter 5s
```

---

### 4.2 Failover Strategies

#### Graceful Degradation

```
Market data feed down?
    ↓
Use cached prices (max age: 1 min)
Show "stale" indicator
    ↓
User still sees working dashboard
    ↓
Auto-reconnect when feed available
```

**Implementation:**
```typescript
class RealtimeDataService {
  async getPrice(symbol: string): Promise<Price> {
    try {
      // Try WebSocket first (live)
      return await this.websocket.getPrice(symbol);
    } catch (error) {
      // Fall back to cached price
      const cached = await this.cache.get(`price:${symbol}`);
      if (cached) {
        return { ...cached, isStale: true };
      }
      // Last resort: DB query
      return await this.db.getLastPrice(symbol);
    }
  }
}
```

---

#### Data Replication

**Multi-Region Setup:**
```
┌─────────────────┐        ┌─────────────────┐
│   US-East       │        │   EU-West       │
│ Primary DB      │ ←sync→ │ Replica DB      │
│ ClickHouse 1    │        │ ClickHouse 2    │
└─────────────────┘        └─────────────────┘

Replication lag: < 1 second
Failover time: < 30 seconds (automatic)
```

---

### 4.3 Circuit Breaker Pattern

```typescript
class CircuitBreaker {
  private state: 'CLOSED' | 'OPEN' | 'HALF_OPEN' = 'CLOSED';
  private failureCount = 0;
  private lastFailureTime = 0;
  private failureThreshold = 5;
  private resetTimeout = 30000; // 30 seconds

  async call<T>(fn: () => Promise<T>): Promise<T> {
    if (this.state === 'OPEN') {
      if (Date.now() - this.lastFailureTime > this.resetTimeout) {
        this.state = 'HALF_OPEN';
      } else {
        throw new Error('Circuit breaker OPEN');
      }
    }

    try {
      const result = await fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  private onSuccess() {
    this.failureCount = 0;
    this.state = 'CLOSED';
  }

  private onFailure() {
    this.failureCount++;
    this.lastFailureTime = Date.now();

    if (this.failureCount >= this.failureThreshold) {
      this.state = 'OPEN'; // Stop calling service
    }
  }
}

// Usage
const dbCircuit = new CircuitBreaker();

async function getPrice(symbol: string) {
  return dbCircuit.call(() =>
    database.query(`SELECT price FROM trades WHERE symbol = ?`, symbol)
  );
}
```

---

### 4.4 Rate Limiting

**Token Bucket Algorithm:**
```typescript
class RateLimiter {
  private tokens: number;
  private lastRefillTime: number = Date.now();
  private readonly capacity: number = 1000;
  private readonly refillRate: number = 100; // tokens per second

  allowRequest(): boolean {
    this.refillTokens();

    if (this.tokens > 0) {
      this.tokens--;
      return true;
    }
    return false;
  }

  private refillTokens() {
    const now = Date.now();
    const timePassed = (now - this.lastRefillTime) / 1000;
    this.tokens = Math.min(
      this.capacity,
      this.tokens + timePassed * this.refillRate
    );
    this.lastRefillTime = now;
  }
}

// Per-user rate limit: 1000 requests/second
const limiter = new RateLimiter();

app.use((req, res, next) => {
  if (!limiter.allowRequest()) {
    return res.status(429).json({ error: 'Rate limit exceeded' });
  }
  next();
});
```

---

## 5. Monitoring and Observability

### 5.1 Key Metrics to Track

#### Data Pipeline Metrics
```
Stream Lag = (Current timestamp - Message timestamp)

For real-time trading dashboard:
- Acceptable: < 100ms
- Warning: > 500ms
- Critical: > 5 seconds

Alert when lag exceeds 500ms:
Average lag over 1-minute window > 500ms
```

---

#### Database Metrics
```
Query Latency:
  - p50: < 50ms (median)
  - p95: < 200ms (95th percentile)
  - p99: < 1000ms (99th percentile)

Write Throughput:
  - ClickHouse: 100k+ events/sec per node
  - TimescaleDB: 10k+ inserts/sec per node

Storage Growth:
  - Raw trades: ~100MB per 100k events
  - OHLC aggregates: ~1MB per 100k events
```

---

#### API Metrics
```
WebSocket Connection Metrics:
  - Active connections per node
  - Message throughput (msg/sec)
  - Average message size
  - Connection churn rate

REST Endpoint Metrics:
  - Requests per second
  - Error rate (%)
  - Response time percentiles
```

---

### 5.2 Distributed Tracing

**OpenTelemetry Integration:**

```typescript
import { trace, metrics, context } from '@opentelemetry/api';
import { NodeSDK } from '@opentelemetry/sdk-node';

const sdk = new NodeSDK({
  traceExporter: new JaegerExporter(),
  metricReader: new PrometheusMetricReader(),
});
sdk.start();

const tracer = trace.getTracer('trading-dashboard');

// Trace a price update
app.ws('/prices', (ws) => {
  const span = tracer.startSpan('websocket:price-subscription');

  span.addEvent('connection_established', {
    remote_addr: ws.remoteAddress,
    timestamp: Date.now(),
  });

  ws.on('message', (msg) => {
    const querySpan = tracer.startSpan('db:query', {
      parent: span,
    });

    const result = await database.query(msg);

    querySpan.end();
    ws.send(result);
  });
});
```

**Trace Flow Example:**
```
websocket:price-subscription (100ms total)
├─ auth:verify-token (5ms)
├─ db:load-user-watchlist (20ms)
├─ kafka:subscribe-symbols (10ms)
└─ websocket:wait-first-message (65ms)
  └─ db:query-price (30ms)
```

---

### 5.3 Log Aggregation

**ELK Stack (Elasticsearch, Logstash, Kibana):**

```json
{
  "timestamp": "2024-01-19T10:30:45.123Z",
  "level": "INFO",
  "service": "trading-dashboard",
  "component": "price-service",
  "event": "price_update",
  "symbol": "AAPL",
  "price": 150.25,
  "latency_ms": 45,
  "trace_id": "abc123def456",
  "user_id": "user:12345"
}
```

**Kibana Dashboard:**
```
- Price update latency: 45ms (p95)
- WebSocket message throughput: 2,500 msg/sec
- Database connection pool utilization: 85%
- Error rate: 0.01%
```

---

### 5.4 Alert Configuration

**Critical Alerts:**
```yaml
alerts:
  - name: "High Stream Lag"
    condition: stream_lag_p99 > 5000  # 5 seconds
    severity: CRITICAL
    action: page_oncall

  - name: "WebSocket Connection Error Rate"
    condition: error_rate > 0.5%
    severity: HIGH
    action: send_slack_alert

  - name: "Database Query Timeout"
    condition: query_p99_latency > 10000  # 10 seconds
    severity: HIGH
    action: send_slack_alert

  - name: "Cache Hit Rate Drop"
    condition: cache_hit_rate < 80%
    severity: MEDIUM
    action: send_email_alert
```

---

## Recommended Architecture Stack

### For Trading Dashboard (Recommended)

```
┌─────────────────────────────────────────────────┐
│ Data Sources                                    │
│ - Market feeds (IB, EOD, exchanges)             │
│ - Trading venue APIs                            │
└─────────────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────────────┐
│ Message Queue: Apache Kafka                     │
│ - Partitions by symbol (100+ partitions)        │
│ - Retention: 24 hours                           │
│ - Throughput: 1M+ events/sec                    │
└─────────────────────────────────────────────────┘
         ↓              ↓              ↓
    ┌────────┐    ┌────────┐    ┌────────┐
    │ Flink  │    │ Spark  │    │ Kafka  │
    │ Stream │    │Streams │    │Streams │
    │Process │    │Process │    │Process │
    └────────┘    └────────┘    └────────┘
         ↓              ↓              ↓
┌───────────────────────────────────────────────┐
│ ClickHouse (Primary Analytics DB)             │
│ - Sharded cluster (3 nodes minimum)            │
│ - Replication factor: 2                        │
│ - Query performance: < 100ms p99               │
│ - Storage: 10-50GB per day                     │
└───────────────────────────────────────────────┘
         ↓              ↓
    ┌─────────┐   ┌──────────┐
    │ Redis   │   │TimescaleDB│
    │ Cache   │   │(Analytics)│
    │5-50GB   │   │Complex Q's│
    └─────────┘   └──────────┘
         ↓              ↓
┌───────────────────────────────────────────────┐
│ API Layer (Node.js / Go)                      │
│ - WebSocket: /prices, /trades, /orders        │
│ - REST: /historical, /aggregates              │
│ - 3+ nodes for HA                             │
└───────────────────────────────────────────────┘
         ↓              ↓              ↓
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│Web Dashboard │ │Desktop Client│ │Mobile App    │
│(React)       │ │(Electron)    │ │(React Native)│
└──────────────┘ └──────────────┘ └──────────────┘

Monitoring:
├─ Prometheus (metrics collection)
├─ Grafana (dashboards)
├─ Jaeger (distributed tracing)
├─ ELK (log aggregation)
└─ PagerDuty (alerting)
```

---

## Summary Table

| Aspect | Choice | Reason |
|--------|--------|--------|
| **Time-Series DB** | ClickHouse (primary) + TimescaleDB (analytics) | ClickHouse for speed (100k+ events/sec), TimescaleDB for complex queries |
| **Message Queue** | Kafka | Partitioning by symbol, horizontal scaling, durability |
| **Cache** | Redis Cluster | Sub-millisecond latency, distributed cache |
| **Real-Time Protocol** | WebSocket (prices) + SSE (news) | WebSocket for bidirectional, SSE for broadcast |
| **Stream Processing** | Kafka Streams / Flink | Close to data, minimal latency |
| **API Framework** | Node.js (TypeScript) + Go | Node for WebSockets, Go for backend services |
| **Monitoring** | Prometheus + Grafana | Industry standard, works well with Kubernetes |
| **Tracing** | Jaeger | Open source, distributed tracing |
| **Scaling** | Horizontal (Kubernetes) | Add nodes for more capacity |

---

## Implementation Roadmap

### Phase 1 (Weeks 1-4): Foundation
- Set up Kafka cluster (3 nodes)
- Deploy ClickHouse (2 nodes, replication)
- Implement basic REST API
- Simple WebSocket price feeds

### Phase 2 (Weeks 5-8): Real-Time
- Kafka Streams aggregations (1m, 5m, 1h bars)
- Redis cache layer
- Enhanced WebSocket with delta updates
- Performance testing (target: p99 < 100ms)

### Phase 3 (Weeks 9-12): Analytics
- TimescaleDB setup for complex queries
- Materialized views in both DBs
- Advanced aggregations
- Historical query optimization

### Phase 4 (Weeks 13-16): Operations
- Kubernetes deployment
- Monitoring/alerting stack
- Distributed tracing
- Automated failover

---

**Document Version:** 1.0
**Last Updated:** 2026-01-19
**Status:** Complete
