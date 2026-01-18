# Trading Engine Architecture Documentation

**Project:** RTX Trading Engine
**Version:** 1.0
**Last Updated:** 2026-01-18

---

## Overview

This directory contains comprehensive architecture documentation for the production-ready multi-asset trading engine. The system supports Forex, CFDs, and Cryptocurrency trading with flexible execution models (A-Book STP, B-Book internal, C-Book hybrid).

---

## Documentation Structure

### 1. **SYSTEM_ARCHITECTURE.md**
Comprehensive system design document covering:
- High-level architecture overview
- Component architecture (8 microservices)
- Database design (PostgreSQL, TimescaleDB, Redis)
- Message queue architecture (RabbitMQ)
- Real-time quote system
- Order execution pipeline
- Admin control panel design
- Client trading interface
- Security & compliance
- Performance targets & scalability
- Deployment architecture
- API endpoint reference

**Read this first** for a complete understanding of the system.

### 2. **ADR-001-MICROSERVICES-ARCHITECTURE.md**
Architecture Decision Record documenting:
- Why microservices over monolith
- Trade-offs and rationale
- Considered alternatives (monolith, serverless, SOA)
- Consequences (positive and negative)
- Implementation guidelines
- Rollout plan (20-week timeline)
- Success criteria

**Read this** to understand the architectural decisions and reasoning.

### 3. **C4-DIAGRAMS.md**
Visual architecture using C4 model:
- **Level 1:** System Context (users, external systems)
- **Level 2:** Container Diagram (services, databases, communication)
- **Level 3:** Component Diagrams (OMS, B-Book Engine internals)
- **Data Flow Diagrams** (market data, order execution)
- **Deployment View** (Kubernetes, cloud infrastructure)

**Read this** for visual understanding of system structure.

---

## Quick Reference

### Architecture at a Glance

```
Client Applications (Web/Mobile/MT4)
            ↓
    API Gateway (NGINX)
            ↓
┌───────────────────────────────────────┐
│  Microservices Layer                  │
│  - Client API                         │
│  - Admin API                          │
│  - WebSocket Server                   │
│  - OMS (Order Management)             │
│  - Smart Order Router                 │
│  - B-Book Execution Engine            │
│  - FIX Gateway                        │
│  - LP Manager                         │
└───────────────────────────────────────┘
            ↓
┌───────────────────────────────────────┐
│  Data Layer                           │
│  - PostgreSQL (transactional)         │
│  - TimescaleDB (time-series)          │
│  - Redis Cluster (cache)              │
│  - RabbitMQ Cluster (messaging)       │
└───────────────────────────────────────┘
```

### Key Design Principles

1. **Microservices Architecture**
   - Independent scaling of components
   - Fault isolation
   - Technology flexibility
   - Team independence
   - Zero-downtime deployments

2. **Domain-Driven Design**
   - **Trading Context:** Orders, Positions, Executions
   - **Market Data Context:** Quotes, Ticks, OHLC
   - **Admin Context:** Accounts, Users, Configuration
   - **Risk Context:** Margin, Exposure, Limits

3. **Performance First**
   - Order execution: <50ms
   - Quote delivery: <100ms
   - API response: <200ms (p99)
   - 10,000 concurrent users

4. **Reliability**
   - 99.9% uptime SLA
   - Automatic failover
   - Circuit breakers
   - Graceful degradation

---

## Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Backend** | Go 1.21+ | High performance, concurrency |
| **API Framework** | Gin | Fast HTTP routing |
| **FIX Protocol** | quickfix-go | LP connectivity |
| **WebSocket** | gorilla/websocket | Real-time streaming |
| **Database (OLTP)** | PostgreSQL 16+ | ACID transactions |
| **Database (TSDB)** | TimescaleDB 2.14+ | Tick data, OHLC |
| **Cache** | Redis 7.2+ Cluster | Sub-ms quotes |
| **Message Queue** | RabbitMQ 3.12+ | Async processing |
| **Container Orchestration** | Kubernetes 1.28+ | Service deployment |
| **Load Balancer** | NGINX 1.25+ | Traffic distribution |
| **Monitoring** | Prometheus + Grafana | Metrics & dashboards |
| **Tracing** | Jaeger | Distributed tracing |

---

## Core Components

### 1. **FIX Gateway Service** (Go)
- Manages FIX 4.4 connections to LPs (YOFX1, Prime XM)
- Automatic reconnection with exponential backoff
- Market data subscriptions (35=V, 35=W)
- Order execution for A-Book routing

### 2. **LP Manager Service** (Go)
- Aggregates quotes from multiple providers
- Normalizes different protocols (FIX, REST, WebSocket)
- Applies spread markup
- Implements failover logic

### 3. **Order Management System - OMS** (Go)
- Order validation and lifecycle management
- Position tracking and P&L calculation
- Risk checks (margin, limits, exposure)
- State machine for order flow

### 4. **Smart Order Router** (Go)
- Routes orders to A/B/C-Book based on:
  - Client classification (VIP, profitable, beginner)
  - Position size (large orders → A-Book)
  - Symbol volatility (high vol → A-Book)
  - LP availability
  - Risk-based rules

### 5. **B-Book Execution Engine** (Go)
- Internal order matching
- Position management (hedging/netting modes)
- P&L calculation (unrealized/realized)
- Margin calculation and stop-out detection
- Ledger system (deposits, withdrawals, trades)

### 6. **WebSocket Server** (Go)
- Real-time quote streaming
- Account updates (balance, equity, margin)
- Position updates (unrealized P&L)
- Order execution notifications

### 7. **Client API** (Go + Gin)
- RESTful endpoints for trading operations
- Authentication & authorization (JWT)
- Rate limiting
- Market data API (ticks, OHLC, symbols)

### 8. **Admin API** (Go + Gin)
- Account management
- Financial operations (deposit/withdraw)
- Execution mode control (A/B/C-Book)
- Symbol management
- LP management
- System configuration

---

## Database Design

### PostgreSQL (Transactional Data)
- `accounts` - Client accounts
- `orders` - Order history
- `positions` - Open and closed positions
- `trades` - Execution records
- `ledger` - Financial transactions

### TimescaleDB (Time-Series Data)
- `ticks` - Tick data from LPs (hypertable)
- `ohlc` - OHLC candlesticks (hypertable)
- Continuous aggregates for automatic OHLC generation
- Compression policies (older than 7 days)

### Redis (Caching & Real-Time)
- `quote:{symbol}` - Latest quotes (60s TTL)
- `session:{token}` - User sessions
- `positions:{accountId}` - Real-time position state
- `orderbook:{symbol}:bids/asks` - Order book levels

---

## Message Queue (RabbitMQ)

### Queues
- `orders.validate` - Order validation pipeline
- `orders.route` - Routing decisions
- `executions.bbook` - B-Book execution
- `executions.abook` - A-Book execution
- `notifications.email` - Email notifications
- `notifications.push` - Push notifications

### Dead Letter Queue (DLQ)
- Failed messages after 3 retries
- Admin review and manual reprocessing

---

## Real-Time Quote System

### Flow
```
LP Feeds → LP Adapters → LP Manager → Quote Normalization
    ↓
Redis Cache (TTL: 60s) ← WebSocket Hub ← Subscribers
    ↓
TimescaleDB (Tick Storage)
    ↓
B-Book Engine (Price for execution)
```

### Performance
- Quote latency: <100ms (LP to client)
- Failover time: <5 seconds
- Rate limiting: 10,000 msg/sec per WebSocket server

---

## Order Execution Pipeline

### Flow
```
Client Order → API Gateway → Client API → OMS
    ↓
Order Validation (symbol, volume, price)
    ↓
Risk Check (margin, limits)
    ↓
Smart Order Router (A/B/C-Book decision)
    ↓
├─ A-Book → FIX Gateway → LP Execution
├─ B-Book → Internal Matching → Position Created
└─ C-Book → Partial to LP + Partial Internal
    ↓
Position Created → Trade Recorded → WebSocket Notification
```

### State Machine
```
PENDING → ACCEPTED → FILLING → FILLED
              ↓
          REJECTED
```

---

## Admin Control Panel

### Features

#### Account Management
- Create/suspend/activate accounts
- Modify leverage (1:1 to 1:500)
- Change margin mode (hedging/netting)
- Reset passwords

#### Financial Operations
- **Deposits:** Bank transfer, credit card, crypto
- **Withdrawals:** Approval workflow
- **Adjustments:** Manual balance corrections
- **Bonuses:** Promotional credits

#### Execution Model Control
- Global mode: A-Book, B-Book, or C-Book
- Client-specific overrides (VIP → A-Book)
- Symbol-based rules (BTCUSD → A-Book only)
- Routing logic configuration

#### Symbol Management
- Enable/disable trading symbols
- Configure spread markup
- Set commission per lot
- Adjust min/max volume
- Configure swap rates

#### Risk Parameters
- Margin call level (default: 100%)
- Stop-out level (default: 50%)
- Max leverage per account type
- Max position size limits
- Max daily volume limits

#### LP Management
- Connect/disconnect LP sessions
- Monitor LP health (latency, uptime)
- Configure failover rules
- View LP quote statistics

---

## Client Trading Interface

### Dashboard Components

1. **Account Summary**
   - Balance, Equity, Margin
   - Free Margin, Margin Level
   - Unrealized P&L

2. **Quick Trade Panel**
   - Symbol selection
   - Volume input with lot calculator
   - Real-time Bid/Ask display
   - SL/TP inputs
   - One-click BUY/SELL

3. **Chart**
   - TradingView integration
   - Multiple timeframes (M1 to D1)
   - Technical indicators (MA, RSI, MACD)

4. **Open Positions**
   - Real-time P&L updates
   - Modify SL/TP
   - Close position (full or partial)
   - Breakeven and trailing stop

5. **Pending Orders**
   - Limit, Stop, Stop-Limit orders
   - Modify or cancel

6. **Trade History**
   - Closed positions
   - Realized P&L
   - Export to CSV/PDF

### Advanced Features

- **Risk Calculator:** Calculate lot size from risk percentage
- **Margin Preview:** Check margin before placing order
- **Trailing Stop:** Fixed, step, or ATR-based
- **Partial Close:** Close 25%, 50%, or custom percentage
- **One-Click Trading:** Pre-configured order templates

---

## Security & Compliance

### Authentication
- JWT tokens (24h for clients, 8h for admins)
- Two-factor authentication (2FA) for admin
- Password hashing (bcrypt with salt)

### Authorization
- Role-based access control (RBAC)
  - SUPER_ADMIN, ADMIN, RISK_MANAGER, SUPPORT, CLIENT

### Encryption
- TLS 1.3 for all connections
- AES-256 for database encryption at rest
- Separate certificates for admin vs client APIs

### Audit Logging
- All critical operations logged
- 7-year retention (regulatory compliance)
- Tamper-proof logging

### Compliance
- KYC/AML verification workflow
- Transaction monitoring
- Automated alerts for suspicious activity
- Regulatory reporting (MiFID II, EMIR, Dodd-Frank)

---

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Order Execution Latency | <50ms | API request → position created |
| Quote Latency | <100ms | LP → client WebSocket |
| API Response Time (p99) | <200ms | 99th percentile |
| WebSocket Throughput | 10,000 msg/s | Per server instance |
| Database Write Throughput | 5,000 TPS | Transactions per second |
| Concurrent Users | 10,000 | Per cluster |
| System Uptime | 99.9% | Monthly SLA |

---

## Scalability Strategy

### Horizontal Scaling
- **Client API:** 3-10 pods (Kubernetes auto-scaling)
- **WebSocket Server:** 5-15 pods (based on connections)
- **OMS:** 5-10 pods (based on order volume)

### Vertical Scaling
- **PostgreSQL:** 16 vCPU, 64GB RAM, SSD
- **TimescaleDB:** 32 vCPU, 128GB RAM, NVMe

### Caching
- **L1 (In-Memory):** Symbol specs, account state
- **L2 (Redis):** Quotes, sessions, position state
- **L3 (Read Replicas):** Historical queries, reporting

---

## Deployment Architecture

### Cloud Infrastructure
- **Provider:** AWS / GCP / Azure
- **Container Orchestration:** Kubernetes (EKS/GKE/AKS)
- **Load Balancer:** NGINX with SSL termination
- **Networking:** VPC with public/private subnets

### High Availability
- Multi-AZ deployment (3 availability zones)
- Auto-scaling groups
- Health checks and auto-recovery
- Disaster recovery site (active-passive)

### Backup & Recovery
- **RTO (Recovery Time Objective):** 1 hour
- **RPO (Recovery Point Objective):** 15 minutes
- PostgreSQL: Full backup daily, incremental 6-hourly
- TimescaleDB: Snapshots daily
- Redis: RDB snapshots hourly

---

## Monitoring & Observability

### Metrics (Prometheus + Grafana)
- System health (CPU, memory, disk, network)
- Trading activity (volume, orders, positions)
- LP performance (latency, uptime, quote freshness)
- Error rates (API errors, DB errors, LP failures)

### Logging (ELK Stack)
- Centralized logging
- Correlation IDs for request tracing
- Log retention: 90 days

### Tracing (Jaeger)
- Distributed tracing across microservices
- Request flow visualization
- Bottleneck identification

### Alerting (PagerDuty)
- Critical: DB down, all LPs disconnected, high error rate
- Warning: Single LP down, high latency, low disk space
- Info: Daily backup complete, deployment complete

---

## Development Roadmap

### Phase 1: Core Services (Weeks 1-4)
- FIX Gateway Service
- LP Manager Service
- B-Book Execution Engine
- PostgreSQL + TimescaleDB + Redis
- **Milestone:** Execute B-Book orders with real-time quotes

### Phase 2: Client Trading (Weeks 5-8)
- Client API Service
- WebSocket Server
- OMS (Order Management System)
- **Milestone:** Clients can trade via web platform

### Phase 3: Smart Routing (Weeks 9-12)
- Smart Order Router
- A-Book execution via FIX Gateway
- **Milestone:** A-Book/B-Book routing operational

### Phase 4: Administration (Weeks 13-16)
- Admin API Service
- Financial operations (deposit/withdraw)
- Symbol management
- **Milestone:** Full broker management capabilities

### Phase 5: Production Hardening (Weeks 17-20)
- Load testing (10,000 concurrent users)
- Disaster recovery testing
- Security audit
- Performance optimization
- **Milestone:** Production-ready system

---

## Contributing

### Architecture Updates
When modifying the architecture:
1. Update relevant diagrams
2. Create ADR if it's a significant decision
3. Update this README
4. Get review from lead architect

### ADR Process
For significant architectural changes:
1. Create new ADR document (ADR-XXX-TITLE.md)
2. Follow template: Context → Decision → Rationale → Consequences
3. Get stakeholder approval
4. Link from this README

---

## References

### External Documentation
- [Microservices Patterns](https://microservices.io/)
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/)
- [FIX Protocol Specification](https://www.fixtrading.org/)
- [TimescaleDB Documentation](https://docs.timescale.com/)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/)

### Internal Documentation
- API Documentation: `/docs/api/`
- Database Schemas: `/docs/database/`
- Deployment Guide: `/docs/deployment/`
- Operations Runbook: `/docs/operations/`

---

## Contact

**System Architect:** [Name]
**Lead Engineer:** [Name]
**DevOps Lead:** [Name]

For architecture questions or proposals, create an issue in the repository or contact the architecture team.

---

**Last Updated:** 2026-01-18
**Document Version:** 1.0
**Next Review:** 2026-02-18
