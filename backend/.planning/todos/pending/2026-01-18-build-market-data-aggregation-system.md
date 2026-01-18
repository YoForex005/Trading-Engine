---
created: 2026-01-18T00:00:00
title: Build production-scale market data aggregation system
area: marketdata
files:
  - backend/datapipeline/pipeline.go
  - backend/tickstore/service.go
  - backend/lpmanager/lp.go
  - backend/marketdata/ (new directory)
---

## Problem

The trading engine needs a comprehensive market data aggregation and distribution system capable of handling production-scale loads similar to professional exchanges. Current implementation has basic tick storage and OHLC generation, but lacks:

1. **Multi-source aggregation** - Need to aggregate data from multiple LPs (LMAX, FlexyMarkets), crypto exchanges (Binance, Coinbase), FIX 4.4 feeds, REST APIs, and CSV imports
2. **Best bid/ask selection** - Real-time aggregation to find best prices across all sources
3. **Performance requirements** - Must process 100,000 ticks/second with <1ms source-to-WebSocket latency and <10ms OHLC calculation
4. **Data quality** - Outlier detection, spike filtering, gap detection/filling, source priority ranking, automatic failover
5. **Storage tiers** - Hot (7 days Redis), Warm (90 days TimescaleDB), Cold (>90 days S3)
6. **Market depth** - Level 2 order book aggregation
7. **24/7 operation** - Must maintain <0.01% data loss with automatic failover and recovery

Existing datapipeline has basic structure but needs enhancement for production-scale requirements.

## Solution

Create `backend/marketdata/` directory with 5 core modules:

1. **aggregator.go** - Multi-source price aggregation
   - Ingest from multiple LPs simultaneously
   - Best bid/ask selection algorithm
   - Source priority ranking and automatic failover
   - Support for FIX 4.4, WebSocket, REST, and CSV sources

2. **normalizer.go** - Data normalization and quality
   - Normalize tick formats from different sources
   - Outlier detection using statistical methods
   - Spike filtering (detect abnormal price movements)
   - Gap detection and intelligent filling
   - Data validation and sanitization

3. **cache.go** - High-performance tick cache
   - Redis-based hot storage (last 7 days)
   - <1ms read latency requirement
   - Circular buffer for tick streams
   - Support 1000+ symbols concurrently

4. **distributor.go** - WebSocket distribution
   - Real-time tick broadcasting to connected clients
   - Subscribe/unsubscribe per symbol
   - Rate limiting and backpressure handling
   - <1ms latency from source to client

5. **recorder.go** - Time-series storage
   - TimescaleDB integration for tick storage
   - Automatic partitioning by time (daily/weekly)
   - S3 archival for cold storage (>90 days)
   - OHLC generation for all timeframes (M1, M5, M15, M30, H1, H4, D1, W1, MN)
   - VWAP calculation
   - Market statistics (high, low, open, close, volume)

**Integration points:**
- Leverage existing `datapipeline/` components where possible
- Enhance `lpmanager/` to support multiple concurrent LPs
- Use existing `tickstore/` for compatibility
- Add FIX 4.4 market data support via `fix/` package

**Performance targets:**
- 100,000 ticks/second throughput
- <1ms source-to-WebSocket latency
- <10ms OHLC calculation
- Support 1000+ symbols
- <0.01% data loss rate
- 24/7 uptime with automatic recovery

**Testing requirements:**
- Load tests simulating 100k ticks/sec
- Failover and recovery tests
- Data quality validation tests
- Latency benchmarks
- Memory and CPU profiling

## Next Steps

1. Design detailed architecture and data flow diagrams
2. Create `backend/marketdata/` directory structure
3. Implement core modules (aggregator, normalizer, cache, distributor, recorder)
4. Integrate with existing systems (lpmanager, datapipeline, tickstore)
5. Add comprehensive tests and benchmarks
6. Performance tuning to meet latency and throughput targets
7. Add monitoring, alerting, and health checks
8. Document configuration and operational procedures
