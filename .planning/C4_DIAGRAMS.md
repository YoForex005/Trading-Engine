# C4 Architecture Diagrams for Real-Time Analytics Dashboard

## Context Diagram (C1)

Shows the system in scope and external systems it interacts with.

```
┌─────────────────────────────────────────────────────────────┐
│                                                              │
│  ┌────────────────┐         ┌──────────────────────────┐   │
│  │  Market Data   │         │  Real-Time Analytics    │   │
│  │  Providers     │────────▶│    Dashboard System     │   │
│  │ (Exchanges,    │         │  (This System)          │   │
│  │  IB, EOD)      │         │                          │   │
│  └────────────────┘         │  - Live price updates   │   │
│                             │  - OHLC candles        │   │
│  ┌────────────────┐         │  - Portfolio analytics  │   │
│  │   User         │◀────────│  - Performance metrics  │   │
│  │  Applications  │         │                          │   │
│  │  (Web, Mobile, │         └──────────────────────────┘   │
│  │   Desktop)     │                                        │
│  └────────────────┘                                        │
│                                                             │
│  ┌────────────────┐         ┌──────────────────────────┐   │
│  │ Compliance &   │◀────────│  Regulatory Data        │   │
│  │ Risk Systems   │         │  - Trade logs           │   │
│  └────────────────┘         │  - Audit trails         │   │
│                             └──────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Container Diagram (C2)

Shows major technology choices and how they communicate.

```
┌──────────────────────────────────────────────────────────────────────┐
│                    Real-Time Analytics System                        │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              Data Ingestion Layer                          │   │
│  │  ┌──────────────┐    ┌──────────────┐                    │   │
│  │  │ Market Data  │    │ User Events  │                    │   │
│  │  │ Collectors   │    │ (Web/Mobile) │                    │   │
│  │  │ (Kafka       │    │              │                    │   │
│  │  │  Producer)   │    │              │                    │   │
│  │  └──────┬───────┘    └──────┬───────┘                    │   │
│  │         │                    │                            │   │
│  │         └────────┬───────────┘                            │   │
│  │                  │                                        │   │
│  │         ┌────────▼──────────┐                            │   │
│  │         │  Kafka Cluster    │                            │   │
│  │         │  (Event Stream)   │                            │   │
│  │         │  Topics:          │                            │   │
│  │         │  - trades         │                            │   │
│  │         │  - quotes         │                            │   │
│  │         │  - events         │                            │   │
│  │         └────────┬──────────┘                            │   │
│  │                  │                                        │   │
│  └──────────────────┼────────────────────────────────────────┘   │
│                     │                                              │
│  ┌──────────────────▼────────────────────────────────────────┐   │
│  │         Stream Processing Layer                          │   │
│  │  ┌──────────────────────────────────────────────┐        │   │
│  │  │ Kafka Streams / Apache Flink                 │        │   │
│  │  │ Jobs:                                        │        │   │
│  │  │ - 1-min OHLCV aggregation                    │        │   │
│  │  │ - 5-min candles                              │        │   │
│  │  │ - Real-time statistics                       │        │   │
│  │  │ - Volume profiles                            │        │   │
│  │  └─────────────┬──────────────────────────────┘        │   │
│  │               │                                         │   │
│  └───────────────┼─────────────────────────────────────────┘   │
│                  │                                              │
│  ┌───────────────▼─────────────────────────────────────────┐   │
│  │         Storage Layer                                   │   │
│  │  ┌──────────────────┐      ┌──────────────────────┐    │   │
│  │  │  ClickHouse      │      │  TimescaleDB        │    │   │
│  │  │  (Cluster)       │      │  (Postgres)         │    │   │
│  │  │                  │      │                      │    │   │
│  │  │  - Raw trades    │      │  - Complex queries  │    │   │
│  │  │  - Tick data     │      │  - Reports          │    │   │
│  │  │  - Aggregates    │      │  - Analytics        │    │   │
│  │  │  TTL: 7 days     │      │  TTL: 2 years       │    │   │
│  │  └────────┬─────────┘      └──────────┬───────────┘    │   │
│  │           │                           │                │   │
│  │  ┌────────▼───────────────────────────▼────────┐      │   │
│  │  │  Redis Cache                                 │      │   │
│  │  │  - Hot price data (5s TTL)                  │      │   │
│  │  │  - Aggregations (1m TTL)                    │      │   │
│  │  │  - User sessions (24h TTL)                  │      │   │
│  │  │  Size: 50-100GB                             │      │   │
│  │  └──────────────────────────────────────────────┘      │   │
│  │                                                         │   │
│  └─────────────────────────────────────────────────────────┘   │
│                        │                                         │
│  ┌─────────────────────▼──────────────────────────────────┐    │
│  │           API Layer (Backend Services)                 │    │
│  │  ┌────────────────────────────────────────────────┐   │    │
│  │  │  Node.js / TypeScript Server                   │   │    │
│  │  │  - WebSocket Server                            │   │    │
│  │  │  - REST API endpoints                          │   │    │
│  │  │  - Authentication & Authorization              │   │    │
│  │  │  - Rate limiting & throttling                  │   │    │
│  │  │  - Data aggregation service                    │   │    │
│  │  │                                                 │   │    │
│  │  │  Endpoints:                                    │   │    │
│  │  │  - POST /api/prices (WebSocket)                │   │    │
│  │  │  - GET /api/historical/:symbol                 │   │    │
│  │  │  - GET /api/aggregates/:symbol/:timeframe      │   │    │
│  │  │  - POST /api/orders                            │   │    │
│  │  │  - GET /api/portfolio                          │   │    │
│  │  └────────────────────────────────────────────────┘   │    │
│  │                                                         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                        │                                          │
│  ┌─────────────────────▼──────────────────────────────────┐     │
│  │         Client Applications                            │     │
│  │  ┌────────────────┐  ┌─────────────┐  ┌────────────┐  │     │
│  │  │  Web           │  │  Desktop    │  │  Mobile    │  │     │
│  │  │  Dashboard     │  │  Client     │  │  App       │  │     │
│  │  │  (React)       │  │ (Electron)  │  │(React Native)│ │     │
│  │  └────────────────┘  └─────────────┘  └────────────┘  │     │
│  │                                                         │     │
│  └─────────────────────────────────────────────────────────┘     │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐    │
│  │              Monitoring & Observability                 │    │
│  │  ┌──────────────┐  ┌──────────┐  ┌────────────┐        │    │
│  │  │ Prometheus   │  │ Grafana  │  │ Jaeger     │        │    │
│  │  │ (Metrics)    │  │(Dash)    │  │ (Tracing)  │        │    │
│  │  └──────────────┘  └──────────┘  └────────────┘        │    │
│  │  ┌──────────────┐  ┌────────────────────────────────┐   │    │
│  │  │ ELK Stack    │  │ PagerDuty / Slack (Alerting)   │   │    │
│  │  │ (Log Agg)    │  │                                │   │    │
│  │  └──────────────┘  └────────────────────────────────┘   │    │
│  │                                                          │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

---

## Component Diagram (C3) - API Layer

Detailed view of the API layer components.

```
┌──────────────────────────────────────────────────────────────────┐
│                     API Layer Components                         │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  Load Balancer / Gateway                   │ │
│  │  - Route requests to healthy API nodes                     │ │
│  │  - SSL/TLS termination                                     │ │
│  │  - Request logging                                         │ │
│  └────────────┬──────────────────┬──────────────────┬─────────┘ │
│               │                  │                  │            │
│  ┌────────────▼──────┐ ┌────────▼──────┐ ┌────────▼──────────┐ │
│  │  API Node 1       │ │  API Node 2    │ │  API Node 3       │ │
│  │                   │ │                │ │                   │ │
│  │ ┌───────────────┐ │ │ ┌────────────┐ │ │ ┌───────────────┐ │ │
│  │ │ WebSocket     │ │ │ │ WebSocket  │ │ │ │ WebSocket     │ │ │
│  │ │ Manager       │ │ │ │ Manager    │ │ │ │ Manager       │ │ │
│  │ │ - Price feeds │ │ │ │ - Broadcasts│ │ │ │ - Live orders │ │ │
│  │ │ - Connections │ │ │ │ - 1000 msg/ │ │ │ │ - Auth check  │ │ │
│  │ │ - Metrics     │ │ │ │   sec/node  │ │ │ │ - Compression │ │ │
│  │ └───────────────┘ │ │ └────────────┘ │ │ └───────────────┘ │ │
│  │                   │ │                │ │                   │ │
│  │ ┌───────────────┐ │ │ ┌────────────┐ │ │ ┌───────────────┐ │ │
│  │ │ REST          │ │ │ │ REST       │ │ │ │ REST          │ │ │
│  │ │ Controller    │ │ │ │ Controller │ │ │ │ Controller    │ │ │
│  │ │ - Auth        │ │ │ │ - Prices   │ │ │ │ - Historical  │ │ │
│  │ │ - Orders      │ │ │ │ - OHLC     │ │ │ │ - Analytics   │ │ │
│  │ │ - Portfolio   │ │ │ │ - Stats    │ │ │ │ - Reports     │ │ │
│  │ └───────────────┘ │ │ └────────────┘ │ │ └───────────────┘ │ │
│  │                   │ │                │ │                   │ │
│  │ ┌───────────────┐ │ │ ┌────────────┐ │ │ ┌───────────────┐ │ │
│  │ │ Data Service  │ │ │ │ Data       │ │ │ │ Data Service  │ │ │
│  │ │ - Query cache │ │ │ │ Service    │ │ │ │ - Compute avg │ │ │
│  │ │ - Compute agg │ │ │ │ - Validate │ │ │ │ - Format data │ │ │
│  │ │ - Format resp │ │ │ │ - Transform│ │ │ │ - Rate limit  │ │ │
│  │ └──────┬────────┘ │ │ └──────┬─────┘ │ │ └──────┬────────┘ │ │
│  │        │          │ │        │       │ │        │          │ │
│  └────────┼──────────┘ └────────┼───────┘ └────────┼──────────┘ │
│           │                     │                 │               │
│  ┌────────┴─────────────────────┴─────────────────┴─────────────┐ │
│  │              Shared Services                                │ │
│  │  ┌───────────────┐  ┌──────────────┐  ┌──────────────────┐ │ │
│  │  │ Auth Service  │  │ Cache Layer  │  │ Logger Service   │ │ │
│  │  │ - JWT verify  │  │ (Redis)      │  │ - Request logs   │ │ │
│  │  │ - Token reval │  │ - Hot keys   │  │ - Error tracking │ │ │
│  │  │ - Permissions │  │ - Session    │  │ - Metrics        │ │ │
│  │  └───────────────┘  └──────────────┘  └──────────────────┘ │ │
│  │                                                              │ │
│  │  ┌───────────────┐  ┌──────────────┐  ┌──────────────────┐ │ │
│  │  │ DB Connection │  │ Rate Limiter │  │ Circuit Breaker  │ │ │
│  │  │ Pool          │  │ (per-user)   │  │ (for DB/cache)   │ │ │
│  │  │ - ClickHouse  │  │ - Token      │  │ - Opens/closes   │ │ │
│  │  │ - TimescaleDB │  │   bucket     │  │ - Protects stack │ │ │
│  │  │ - Max: 500    │  │ - 1000 req/s │  │                  │ │ │
│  │  └───────────────┘  └──────────────┘  └──────────────────┘ │ │
│  │                                                              │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                   │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │        Connections to Storage Layer                          │ │
│  │        (Load balancer routes to appropriate DB)             │ │
│  │  ┌──────────────┐          ┌─────────────────────┐          │ │
│  │  │ ClickHouse   │          │ TimescaleDB + Redis │          │ │
│  │  │ Cluster      │          │ (Analytics)         │          │ │
│  │  └──────────────┘          └─────────────────────┘          │ │
│  │                                                              │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

---

## Component Diagram (C3) - Storage Layer

Detailed view of the storage and data aggregation components.

```
┌──────────────────────────────────────────────────────────────────┐
│                    Storage Layer Components                      │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │           Stream Processing / Aggregation                  │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │ Kafka Streams Topology                               │ │ │
│  │  │                                                      │ │ │
│  │  │ Input: trades-topic (100k events/sec)              │ │ │
│  │  │   ↓                                                 │ │ │
│  │  │ [Stream: Parse & Validate]                         │ │ │
│  │  │   ↓                                                 │ │ │
│  │  │ [Window: 1-minute tumbling]                         │ │ │
│  │  │ [Aggregate: First, Max, Min, Last, Sum]             │ │ │
│  │  │   ↓                                                 │ │ │
│  │  │ [Suppress: Emit only on state change]              │ │ │
│  │  │   ↓                                                 │ │ │
│  │  │ Output Topics:                                      │ │ │
│  │  │   - aggregates-1m (1-min OHLC bars)               │ │ │
│  │  │   - volume-profile-1m                              │ │ │
│  │  │   - tick-count-1m                                  │ │ │
│  │  │                                                      │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  │                                                            │ │
│  └────────────────────────────────────────────────────────────┘ │
│                     ↓              ↓              ↓             │
│  ┌──────────────────┴──────────────┴──────────────┴──────────┐ │
│  │                  Storage Cluster                          │ │
│  │                                                           │ │
│  │  ┌──────────────────────────────────────────────────┐   │ │
│  │  │  ClickHouse Cluster (3+ nodes)                  │   │ │
│  │  │                                                  │   │ │
│  │  │  ┌────────────────────────────────────────────┐ │   │ │
│  │  │  │ Shard 1 (Symbols A-G)                     │ │   │ │
│  │  │  │ - trades_raw (all ticks)                  │ │   │ │
│  │  │  │ - ohlc_1m (1-min bars)                    │ │   │ │
│  │  │  │ - volume_profile (depth)                  │ │   │ │
│  │  │  │ - Replica 1 & 2 (HA)                      │ │   │ │
│  │  │  │ - TTL: 7 days                             │ │   │ │
│  │  │  │ - Compression: enabled                    │ │   │ │
│  │  │  └────────────────────────────────────────────┘ │   │ │
│  │  │                                                  │   │ │
│  │  │  ┌────────────────────────────────────────────┐ │   │ │
│  │  │  │ Shard 2 (Symbols H-O)                     │ │   │ │
│  │  │  │ - Same structure as Shard 1              │ │   │ │
│  │  │  │ - Load balanced across 3 nodes           │ │   │ │
│  │  │  │ - Write throughput: 100k events/sec      │ │   │ │
│  │  │  └────────────────────────────────────────────┘ │   │ │
│  │  │                                                  │   │ │
│  │  │  ┌────────────────────────────────────────────┐ │   │ │
│  │  │  │ Shard 3 (Symbols P-Z)                     │ │   │ │
│  │  │  │ - Same structure as other shards         │ │   │ │
│  │  │  │ - Query parallelization                   │ │   │ │
│  │  │  │ - Aggregate queries < 100ms p99           │ │   │ │
│  │  │  └────────────────────────────────────────────┘ │   │ │
│  │  │                                                  │   │ │
│  │  │  Query Router:                                  │   │ │
│  │  │  - Hash(symbol) % 3 = target shard            │   │ │
│  │  │  - Parallel queries on all shards             │   │ │
│  │  │  - Result merge on coordinator              │   │ │
│  │  └──────────────────────────────────────────────────┘   │ │
│  │                                                           │ │
│  │  ┌──────────────────────────────────────────────────┐   │ │
│  │  │  TimescaleDB (PostgreSQL)                       │   │ │
│  │  │                                                  │   │ │
│  │  │  ┌────────────────────────────────────────────┐ │   │ │
│  │  │  │ Hypertable: trades (time-partitioned)     │ │   │ │
│  │  │  │ - Partition: 1 hour chunks               │ │   │ │
│  │  │  │ - Compression after 48 hours             │ │   │ │
│  │  │  │ - Retention: 2 years                      │ │   │ │
│  │  │  └────────────────────────────────────────────┘ │   │ │
│  │  │                                                  │   │ │
│  │  │  ┌────────────────────────────────────────────┐ │   │ │
│  │  │  │ Materialized Views:                       │ │   │ │
│  │  │  │ - ohlc_daily (refreshes every 1 min)     │ │   │ │
│  │  │  │ - volume_by_symbol (pre-aggregated)      │ │   │ │
│  │  │  │ - correlation_matrix (cross-assets)      │ │   │ │
│  │  │  └────────────────────────────────────────────┘ │   │ │
│  │  │                                                  │   │ │
│  │  │  Indexes:                                       │   │ │
│  │  │  - (symbol, time DESC) - primary lookup     │   │ │
│  │  │  - (price) - range queries                   │   │ │
│  │  │  - (time) - time-range scans               │   │ │ │
│  │  └──────────────────────────────────────────────────┘   │ │
│  │                                                           │ │
│  │  ┌──────────────────────────────────────────────────┐   │ │
│  │  │  Redis Cluster (Cache Layer)                    │   │ │
│  │  │                                                  │   │ │
│  │  │  Node 1 (Master):                               │   │ │
│  │  │  - Hot symbols (AAPL, GOOGL, MSFT)            │   │ │
│  │  │  - Size: 20GB                                  │   │ │
│  │  │  - Keys: price:{symbol}, ohlc:{symbol}:{TF}  │   │ │
│  │  │                                                  │   │ │
│  │  │  Node 2 & 3 (Replicas):                        │   │ │
│  │  │  - Real-time replication                      │   │ │
│  │  │  - Failover to new master < 1 sec            │   │ │
│  │  │  - Sentinel for auto-failover                │   │ │
│  │  │                                                  │   │ │
│  │  │  Cache Policies:                               │   │ │
│  │  │  - LRU eviction when full                     │   │ │
│  │  │  - Write-through for consistency              │   │ │
│  │  │  - Async flush to DB                          │   │ │
│  │  └──────────────────────────────────────────────────┘   │ │
│  │                                                           │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
└──────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Diagram - Real-Time Price Update

Shows how a single price update flows through the system.

```
1. Market Data Feed arrives
   ├─ Symbol: AAPL
   ├─ Price: 150.25
   ├─ Volume: 50000
   ├─ Bid: 150.24
   ├─ Ask: 150.26
   └─ Timestamp: 2024-01-19 10:30:45.123Z

                     ↓
2. Kafka Producer
   └─ Topic: trades
      Partition: hash(AAPL) % 10 = Partition 3
      Offset: 1,000,000

                     ↓
3. Stream Processing (Kafka Streams)
   ├─ Window: 1-minute tumbling
   ├─ Aggregation:
   │  ├─ open: first(price)
   │  ├─ high: max(price)
   │  ├─ low: min(price)
   │  └─ close: last(price)
   └─ Output topics:
      ├─ aggregates-1m (1-min OHLC)
      └─ volume-profile-1m

                     ↓
4. Storage (Parallel writes)
   ├─ ClickHouse
   │  ├─ Insert to trades_raw (Shard 1)
   │  ├─ Materialized view updates ohlc_1m
   │  └─ TTL check (delete if > 7 days)
   │
   ├─ TimescaleDB
   │  ├─ Insert to trades hypertable
   │  └─ Auto-compress if > 48 hours old
   │
   └─ Redis
      ├─ SET price:AAPL 150.25 EX 5
      ├─ INCR quote_counter
      └─ LPUSH price_history:AAPL 150.25

                     ↓
5. API Layer (Subscription Processing)
   ├─ Check subscriptions in Redis
   │  └─ Users watching AAPL: [user1, user2, user3]
   │
   ├─ WebSocket Manager
   │  ├─ Compute delta from last price (150.24 → 150.25)
   │  ├─ Compress message (MessagePack)
   │  └─ Send to 3 WebSocket connections
   │
   └─ Cache update
      └─ last_price:AAPL = 150.25

                     ↓
6. Client Application (Browser)
   ├─ Receive delta message
   │  {
   │    "type": "delta",
   │    "symbol": "AAPL",
   │    "price": 150.25,
   │    "bid": 150.24,
   │    "timestamp": 1234567890
   │  }
   │
   ├─ Buffer update (add to update buffer)
   │
   └─ Flush every 100ms
      ├─ Apply all buffered updates
      ├─ Update DOM
      ├─ Trigger re-render
      └─ Display to user

TOTAL LATENCY: ~100ms (Market → Display)
- Kafka: 10ms
- Stream processing: 20ms
- Storage writes: 30ms
- API + WebSocket: 20ms
- Browser rendering: 20ms
```

---

## Deployment Diagram

Shows how components are distributed across infrastructure.

```
┌──────────────────────────────────────────────────────────┐
│              Kubernetes Cluster (3 AZs)                  │
│                                                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────┐ │
│  │  AZ-1           │  │  AZ-2           │  │  AZ-3    │ │
│  │  us-east-1a     │  │  us-east-1b     │  │us-east-1c │
│  │                 │  │                 │  │           │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌───────┐ │ │
│  │ │ Kafka       │ │  │ │ Kafka       │ │  │ │Kafka  │ │ │
│  │ │ Node 1      │ │  │ │ Node 2      │ │  │ │Node 3 │ │ │
│  │ │ (Broker 1)  │ │  │ │ (Broker 2)  │ │  │ │(Broker3)│ │
│  │ │ Pod: 1      │ │  │ │ Pod: 1      │ │  │ │Pod: 1 │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └───────┘ │ │
│  │                 │  │                 │  │           │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌───────┐ │ │
│  │ │ ClickHouse  │ │  │ │ ClickHouse  │ │  │ │Click  │ │ │
│  │ │ Shard 1     │ │  │ │ Shard 2     │ │  │ │House  │ │ │
│  │ │ Replica 1   │ │  │ │ Replica 1   │ │  │ │Shard3 │ │ │
│  │ │ Pod: 2      │ │  │ │ Pod: 2      │ │  │ │Pod: 2 │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └───────┘ │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │           │ │
│  │ │ Flink Job   │ │  │ │ Flink Job   │ │  │           │ │
│  │ │ (Stream Proc)│ │  │ │ (Backup)    │ │  │           │ │
│  │ │ Pod: 1      │ │  │ │ Pod: 1      │ │  │           │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │           │ │
│  │                 │  │                 │  │           │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌───────┐ │ │
│  │ │ API Server  │ │  │ │ API Server  │ │  │ │API    │ │ │
│  │ │ (WS + REST) │ │  │ │ (WS + REST) │ │  │ │Server │ │ │
│  │ │ Pod: 3      │ │  │ │ Pod: 3      │ │  │ │Pod: 2 │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └───────┘ │ │
│  │                 │  │                 │  │           │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌───────┐ │ │
│  │ │ Redis       │ │  │ │ Redis       │ │  │ │Redis  │ │ │
│  │ │ Master      │ │  │ │ Slave       │ │  │ │Slave  │ │ │
│  │ │ Pod: 1      │ │  │ │ Pod: 1      │ │  │ │Pod: 1 │ │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │ └───────┘ │ │
│  │                 │  │                 │  │           │ │
│  │ ┌─────────────┐ │  │ ┌─────────────┐ │  │           │ │
│  │ │ TimescaleDB │ │  │ │ TimescaleDB │ │  │           │ │
│  │ │ Primary     │ │  │ │ Standby     │ │  │           │ │
│  │ │ Pod: 1      │ │  │ │ Pod: 1      │ │  │           │ │
│  │ └─────────────┘ │  │ └─────────────┘ │  │           │ │
│  │                 │  │                 │  │           │ │
│  └─────────────────┘  └─────────────────┘  └───────────┘ │
│          ↑                   ↑                    ↑        │
│          └───────────────────┴────────────────────┘        │
│                      Networking                           │
│                  (Service Mesh: Istio)                    │
│                                                           │
│  ┌─────────────────────────────────────────────────────┐ │
│  │          Ingress Controller (Nginx)                 │ │
│  │  - TLS termination                                  │ │
│  │  - Rate limiting                                    │ │
│  │  - Sticky sessions for WebSocket                   │ │
│  │  - Canary deployments                              │ │
│  └─────────────────────────────────────────────────────┘ │
│                                                           │
│  ┌─────────────────────────────────────────────────────┐ │
│  │          Monitoring Stack (same cluster)            │ │
│  │  - Prometheus                                       │ │
│  │  - Grafana                                          │ │
│  │  - Jaeger                                           │ │
│  │  - ELK Stack                                        │ │
│  └─────────────────────────────────────────────────────┘ │
│                                                           │
└──────────────────────────────────────────────────────────┘
         ↓
    ┌─────────────────┐
    │ External Load   │
    │ Balancer (ALB)  │
    │ - Auto-scales   │
    │ - Health checks │
    │ - SSL/TLS       │
    └─────────────────┘
         ↓
    ┌─────────────────────┐
    │ Client Applications │
    │ - Web browsers      │
    │ - Mobile apps       │
    │ - Desktop clients   │
    └─────────────────────┘
```

---

## Network Topology - Real-Time Data Flow

Shows network paths and latency characteristics.

```
┌──────────────────────────────────────────────────────────────┐
│             CDN (CloudFront)                                 │
│  - Static assets (JS/CSS)                                   │
│  - Latency: < 10ms                                          │
└────────────────┬─────────────────────────────────────────────┘
                 │
          ┌──────▼───────┐
          │ ALB (Edge)   │
          │ us-east-1    │
          └──────┬───────┘
          ┌──────▴───────┐
          │ Sticky Sess. │
          │ (for WS)     │
          └──────┬───────┘
    ┌─────────────┼─────────────┐
    │             │             │
┌───▼──┐    ┌──────▼──┐    ┌───▼──┐
│ API  │    │  API    │    │ API  │
│ Pod1 │────│  Pod2   │────│ Pod3 │
│10Gbps│    │10Gbps   │    │10Gbps│
└───┬──┘    └────┬────┘    └──┬───┘
    │            │             │
    └────────────┼─────────────┘
                 │
      ┌──────────▴──────────┐
      │ Service Mesh (Istio)│
      │ - mTLS encryption   │
      │ - Load balancing    │
      │ - Circuit breaker   │
      └──────────┬──────────┘
          ┌──────┴───────┐
          │              │
    ┌─────▼──────┐  ┌────▼──────┐
    │ ClickHouse │  │ TimescaleDB│
    │ Cluster    │  │ + Redis    │
    │ (Kafka)    │  │            │
    └────────────┘  └────────────┘

Latencies:
- Client → ALB: 2-5ms
- ALB → API Pod: 0.5-1ms
- API Pod → ClickHouse: 5-10ms (local reads)
- API Pod → Redis: 1-2ms (hot cache)
- API Pod → Client (WS): 1-2ms

Total WebSocket latency: 10-20ms
Total REST query latency: 20-40ms
```

---

## End-to-End System Interaction Sequence

Shows timing and order of operations.

```
Time  Market Feed     Kafka       Stream         Storage        API          WebSocket
      Producer       Topic        Processor      Layer          Layer        Layer
|      |              |             |              |              |            |
|      |              |             |              |              |            |
0ms   TICK            |             |              |              |            |
      AAPL=150.25     |             |              |              |            |
      Vol=50k         |             |              |              |            |
      |               |             |              |              |            |
5ms   |         PUSH  |             |              |              |            |
      |         ----->| trades-0001 |              |              |            |
      |               | (Partition) |              |              |            |
      |               |             |              |              |            |
10ms  |               |       PULL  |              |              |            |
      |               |      <------|              |              |            |
      |               |             | AGGREGATE   |              |            |
      |               |             | (window)    |              |            |
      |               |             |             |              |            |
20ms  |               |             |        WRITE|              |            |
      |               |             |        ---->| INSERT       |            |
      |               |             |             | trades_raw   |            |
      |               |             |             |              |            |
      |               |             |        EMIT|              |            |
      |               |             |        ---->| UPDATE CACHE |            |
      |               |             |             |              |            |
25ms  |               |             |             |         QUERY|            |
      |               |             |             |         <----| price:AAPL |
      |               |             |             |              |            |
30ms  |               |             |             |              | BROADCAST  |
      |               |             |             |              | ---------->|
      |               |             |             |              | Delta msg  |
      |               |             |             |              |            | RECEIVE
      |               |             |             |              |            | 2K users
      |               |             |             |              |            |
40ms  |               |             |             |              |            | UPDATE DOM
      |               |             |             |              |            | (batch)
      |               |             |             |              |            | - Update
      |               |             |             |              |            |   price
      |               |             |             |              |            | - Trigger
      |               |             |             |              |            |   animation
      |               |             |             |              |            | - Repaint
      |               |             |             |              |            |
50ms  |               |             |             |              |            | DISPLAY

SUMMARY:
- E2E Latency: ~40-50ms (excellent for real-time trading)
- Peak throughput: 100k events/sec sustained
- P99 latency: < 100ms under load
```

---

## Scalability Diagram

Shows how components scale with load.

```
Load: 100k events/sec

┌─────────────────────────────────────┐
│ Kafka Cluster Scaling               │
├─────────────────────────────────────┤
│ Nodes needed: 3                     │
│ Partitions: 100 (10x redundancy)   │
│ Throughput per node: 33k/sec        │
│ Replication factor: 3               │
│ Disk per node: 10TB (10 days)       │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ClickHouse Scaling                  │
├─────────────────────────────────────┤
│ Nodes needed: 3 (sharded)           │
│ Shards: 3                           │
│ Replicas per shard: 2               │
│ Throughput per shard: 33k/sec       │
│ Memory per node: 128GB              │
│ Disk per node: 50TB                 │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ API Servers Scaling                 │
├─────────────────────────────────────┤
│ Pods needed: 10                     │
│ WS connections per pod: 5000        │
│ Total WS connections: 50k           │
│ REST requests/sec: 10k              │
│ Requests per pod: 1k/sec            │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Redis Cache Scaling                 │
├─────────────────────────────────────┤
│ Master nodes: 1                     │
│ Slave nodes: 2                      │
│ Total capacity: 100GB               │
│ Throughput: 1M ops/sec              │
│ Hit ratio: 95%                      │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ Load per user (monitoring)          │
├─────────────────────────────────────┤
│ Avg price updates/sec: 50           │
│ Avg bandwidth/user: 100 Kbps        │
│ Storage per user/month: 50 MB       │
│ Concurrent WS connections: 50k      │
│ Total users supported: 2 million    │
└─────────────────────────────────────┘

SCALING STRATEGY:
- Kafka: Add partitions → distribute across brokers
- ClickHouse: Add shards → hash-based routing
- API: Add pods → auto-scale via Kubernetes HPA
- Redis: Manual rebalancing → consistent hashing
```

---

End of C4 Diagrams Document
