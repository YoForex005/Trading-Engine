# Trading Engine Database Architecture

## Executive Summary

This document outlines a production-grade, scalable database architecture for the RTX Trading Engine. The design prioritizes ACID compliance, high-performance tick data ingestion, real-time quote delivery, and regulatory compliance.

## Architecture Overview

### Database Stack Selection

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  (Go Trading Engine)                         │
└─────────────────────────────────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
        ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  PostgreSQL  │  │ TimescaleDB  │  │    Redis     │
│  (Primary)   │  │ (Time-Series)│  │   (Cache)    │
├──────────────┤  ├──────────────┤  ├──────────────┤
│ • Users      │  │ • Tick Data  │  │ • Quotes     │
│ • Accounts   │  │ • OHLC       │  │ • Sessions   │
│ • Orders     │  │ • Market     │  │ • Rate Limit │
│ • Positions  │  │   Depth      │  │ • Positions  │
│ • Trades     │  │              │  │   Cache      │
│ • Ledger     │  │              │  │              │
│ • Audit      │  │              │  │              │
└──────────────┘  └──────────────┘  └──────────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                           ▼
                  ┌──────────────┐
                  │  NATS/RabbitMQ│
                  │ (Message Queue)│
                  ├──────────────┤
                  │ • Order Route │
                  │ • Execution   │
                  │ • Market Data │
                  │   Distribution│
                  └──────────────┘
```

### Why This Stack?

#### PostgreSQL (Primary Database)
- **ACID Compliance**: Critical for financial transactions
- **Mature Ecosystem**: Battle-tested in financial services
- **Advanced Features**: JSONB, Full-text search, Advanced indexing
- **Strong Consistency**: Required for account balances and positions
- **Audit Trail**: Comprehensive logging capabilities

#### TimescaleDB (Time-Series Data)
- **Built on PostgreSQL**: Leverages all PostgreSQL features
- **Hypertables**: Automatic partitioning by time
- **Compression**: 90%+ compression for historical data
- **Continuous Aggregates**: Real-time OHLC generation
- **Retention Policies**: Automatic data lifecycle management
- **Performance**: 10-100x faster for time-series queries

#### Redis (Cache & Real-Time)
- **Sub-millisecond Latency**: Critical for real-time quotes
- **Pub/Sub**: Real-time market data distribution
- **Rate Limiting**: Sliding window counters
- **Session Store**: Fast user session lookup
- **Position Cache**: Quick position validation

#### NATS/RabbitMQ (Message Queue)
- **Order Routing**: Reliable order delivery to LPs
- **Market Data**: Fan-out distribution to multiple consumers
- **Asynchronous Processing**: Decouples trade execution
- **Dead Letter Queue**: Failed order handling

## Database Schema Design

### Schema Organization

```
trading_engine_db/
├── users/              # Authentication & authorization
├── accounts/           # Trading accounts & balances
├── instruments/        # Symbol specifications
├── orders/             # Order lifecycle
├── positions/          # Position management
├── transactions/       # Financial ledger
├── risk/              # Risk management
└── audit/             # Compliance & audit trails
```

## Performance Targets

| Metric | Target | Strategy |
|--------|--------|----------|
| Order Latency | <50ms p99 | In-memory order matching |
| Tick Ingestion | 100k ticks/sec | TimescaleDB hypertables |
| Quote Delivery | <5ms | Redis cache |
| Balance Query | <10ms | Indexed lookups + cache |
| OHLC Generation | Real-time | Continuous aggregates |
| Backup Time | <15min | Incremental WAL backups |

## Data Flow

### Order Execution Flow

```
Client → API → Validation → PostgreSQL (Order Created)
                    ↓
              Message Queue
                    ↓
         ┌──────────┴──────────┐
         │                     │
    LP Router              B-Book Engine
         │                     │
         ↓                     ↓
    FIX Gateway          PostgreSQL
         │               (Position Created)
         ↓                     │
    Order Fill                 ↓
         │                Redis Cache
         │                 (Update)
         └──────────┬──────────┘
                    ↓
            Client WebSocket
```

### Market Data Flow

```
LP Feed → FIX Gateway → TimescaleDB (Ticks)
                 ↓
           Redis (Cache) → WebSocket → Clients
                 ↓
      Continuous Aggregate (OHLC)
```

## Scalability Considerations

### Horizontal Scaling
- **Read Replicas**: 2-3 replicas for analytics and reporting
- **Connection Pooling**: PgBouncer for connection management
- **Sharding**: By account_id for multi-region deployment

### Vertical Scaling
- **CPU**: 16+ cores for concurrent query processing
- **RAM**: 64GB+ for PostgreSQL shared_buffers
- **Storage**: NVMe SSDs for low-latency I/O
- **Network**: 10Gbps for market data ingestion

## Security Measures

### Data Protection
- **Encryption at Rest**: AES-256 for all databases
- **Encryption in Transit**: TLS 1.3 for all connections
- **Column-Level Encryption**: Sensitive PII (email, phone)
- **Row-Level Security**: Account isolation

### Access Control
- **Role-Based Access**: admin, trader, readonly, system
- **Database Users**: Separate credentials per service
- **Password Policy**: Bcrypt with cost factor 12
- **Audit Logging**: All admin actions logged

### Compliance
- **PII Handling**: GDPR-compliant data retention
- **Audit Trail**: 7-year retention for regulatory compliance
- **Immutable Logs**: Write-once audit tables
- **Data Masking**: PII redaction for non-production environments

## Disaster Recovery

### Backup Strategy
- **Full Backup**: Daily at 02:00 UTC
- **Incremental Backup**: WAL archiving every 5 minutes
- **Retention**: 30 days online, 90 days cold storage
- **Cross-Region**: Replicate to 2 geographic regions

### Recovery Objectives
- **RTO (Recovery Time Objective)**: <1 hour
- **RPO (Recovery Point Objective)**: <5 minutes
- **Failover**: Automatic with pgpool or Patroni
- **Testing**: Monthly DR drills

## Migration Strategy

### Phase 1: Schema Creation (Week 1)
- Create PostgreSQL schemas
- Create TimescaleDB hypertables
- Set up Redis configuration
- Configure NATS message queues

### Phase 2: Data Migration (Week 2)
- Migrate current in-memory data to PostgreSQL
- ETL historical tick data to TimescaleDB
- Validate data integrity
- Performance testing

### Phase 3: Dual-Write (Week 3)
- Write to both in-memory and database
- Compare results
- Monitor performance
- Fix discrepancies

### Phase 4: Read Cutover (Week 4)
- Gradually shift reads to database
- Monitor latency and errors
- Rollback plan ready
- Full cutover with feature flag

### Phase 5: Cleanup (Week 5)
- Remove in-memory storage
- Optimize indexes
- Tune database parameters
- Document lessons learned

## Cost Estimation

### Infrastructure Costs (Monthly)

| Component | Specs | Monthly Cost |
|-----------|-------|--------------|
| PostgreSQL Primary | 16 vCPU, 64GB RAM, 1TB SSD | $800 |
| PostgreSQL Replica 1 | 16 vCPU, 64GB RAM, 1TB SSD | $800 |
| PostgreSQL Replica 2 | 16 vCPU, 64GB RAM, 1TB SSD | $800 |
| TimescaleDB | 8 vCPU, 32GB RAM, 2TB SSD | $600 |
| Redis | 4 vCPU, 16GB RAM | $200 |
| NATS | 4 vCPU, 8GB RAM | $100 |
| Backup Storage | 5TB | $100 |
| **Total** | | **$3,400/month** |

## Monitoring & Alerting

### Key Metrics

#### PostgreSQL
- Query latency (p50, p95, p99)
- Connection pool usage
- Replication lag
- Disk I/O utilization
- Cache hit ratio

#### TimescaleDB
- Chunk compression ratio
- Hypertable size growth
- Continuous aggregate lag
- Retention policy execution

#### Redis
- Memory usage
- Eviction rate
- Key hit rate
- Pub/sub backlog

### Alerting Thresholds
- Replication lag > 10 seconds
- Connection pool > 80% utilized
- Disk space < 20% free
- Query latency p99 > 100ms
- Cache hit ratio < 90%

## Next Steps

1. Review and approve architecture
2. Set up development environment
3. Create DDL scripts (see 02-ddl-schema.sql)
4. Implement migration scripts (see 03-migration-plan.md)
5. Configure monitoring dashboards
6. Execute phased migration

---

**Document Version**: 1.0
**Last Updated**: 2026-01-18
**Author**: Database Architecture Team
**Status**: Draft - Pending Approval
