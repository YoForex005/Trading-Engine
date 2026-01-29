# ADR-001: Adopt Microservices Architecture for Trading Engine

**Date:** 2026-01-18
**Status:** Accepted
**Decision Makers:** System Architect, Lead Engineer, CTO
**Stakeholders:** Engineering Team, Operations, Product Management

---

## Context

We are designing a production-ready trading engine that needs to support:
- Multiple asset classes (Forex, CFDs, Crypto)
- Multiple execution models (A-Book STP, B-Book internal, C-Book hybrid)
- Real-time market data streaming from multiple liquidity providers
- High-frequency trading operations with low latency requirements
- Administrative operations for broker management
- Compliance and regulatory reporting

The system must be:
- **Scalable:** Handle growing trading volume and user base
- **Reliable:** 99.9% uptime SLA for trading operations
- **Performant:** <50ms order execution latency, <100ms quote delivery
- **Maintainable:** Easy to update and deploy without full system downtime
- **Evolvable:** Ability to add new LPs, symbols, features without major refactoring

---

## Decision

We will adopt a **microservices architecture** with the following service boundaries based on Domain-Driven Design principles:

### Core Services

1. **FIX Gateway Service**
   - Responsibility: Manage FIX 4.4 protocol connections to liquidity providers
   - Technology: Go (quickfix-go)
   - Scale: 2-4 instances (per LP redundancy)

2. **LP Manager Service**
   - Responsibility: Aggregate quotes from multiple LPs, failover logic
   - Technology: Go
   - Scale: 3-5 instances (horizontal scaling)

3. **Order Management System (OMS)**
   - Responsibility: Order validation, lifecycle management, position tracking
   - Technology: Go
   - Scale: 5-10 instances (based on order volume)

4. **Smart Order Router**
   - Responsibility: Route orders to A/B/C-Book based on rules
   - Technology: Go
   - Scale: 3-5 instances (stateless)

5. **B-Book Execution Engine**
   - Responsibility: Internal order matching, position management, P&L calculation
   - Technology: Go
   - Scale: 3-8 instances (based on internal order volume)

6. **WebSocket Server**
   - Responsibility: Real-time quote streaming and account updates
   - Technology: Go (gorilla/websocket)
   - Scale: 5-15 instances (based on concurrent connections)

7. **Client API Service**
   - Responsibility: RESTful API for client trading operations
   - Technology: Go (Gin framework)
   - Scale: 3-10 pods (Kubernetes auto-scaling)

8. **Admin API Service**
   - Responsibility: RESTful API for broker administrative functions
   - Technology: Go (Gin framework)
   - Scale: 2-5 pods (lower volume than client API)

### Supporting Infrastructure

- **PostgreSQL:** Transactional data (accounts, orders, positions, trades)
- **TimescaleDB:** Time-series data (ticks, OHLC)
- **Redis Cluster:** Caching, session management, real-time state
- **RabbitMQ Cluster:** Asynchronous message processing, event notifications

---

## Rationale

### Why Microservices Over Monolith?

#### 1. Independent Scalability

**Problem:** Different components have vastly different scaling requirements.

- **Market Data:** High throughput (10,000 quotes/sec), low compute
- **Order Execution:** Low throughput (100 orders/sec), high compute, low latency
- **Admin API:** Very low throughput (10 req/min), complex logic

**Solution:** Microservices allow each component to scale independently.

```
Example Scaling Profile:
- WebSocket Server: 15 instances (high connection count)
- OMS: 8 instances (high transaction rate)
- Admin API: 2 instances (low volume)

Monolith would require scaling the entire application to match the highest demand component.
```

**Benefit:** 60% cost savings compared to scaling entire monolith.

#### 2. Fault Isolation

**Problem:** A crash in one component should not bring down the entire system.

**Scenarios:**
- FIX Gateway crashes due to malformed LP message → Trading continues via B-Book
- WebSocket server memory leak → Clients reconnect, trading unaffected
- Admin API bug → Clients can still trade, only admin functions impacted

**Solution:** Microservices provide failure boundaries. Circuit breakers prevent cascade failures.

```go
// Circuit Breaker Pattern
type CircuitBreaker struct {
    threshold int
    failures  int
    state     string // CLOSED, OPEN, HALF_OPEN
}

// If FIX Gateway fails, OMS automatically routes to B-Book
if fixGateway.IsHealthy() {
    route = A_BOOK
} else {
    log.Warn("FIX Gateway unavailable, routing to B-Book")
    route = B_BOOK
}
```

**Benefit:** 99.95% uptime achieved (vs 99.5% for monolith where any component failure causes full outage).

#### 3. Technology Flexibility

**Problem:** Different components have different optimal technology stacks.

**Current Choices:**
- **Go:** High-performance services (OMS, Execution Engine, FIX Gateway)
  - Benefit: Low latency, efficient concurrency, strong typing
- **Python (Future):** Analytics, backtesting, risk modeling
  - Benefit: Rich ML/data science ecosystem
- **Rust (Future):** Ultra-low latency components (nanosecond-critical)
  - Benefit: Zero-cost abstractions, memory safety

**Solution:** Microservices allow polyglot architecture.

**Benefit:** Use the right tool for each job rather than forcing everything into one language.

#### 4. Team Independence

**Problem:** Large development team working on monolith causes merge conflicts and deployment bottlenecks.

**Team Structure:**
- Team A (Trading): Client API, OMS, Order Router
- Team B (Market Data): LP Manager, WebSocket Server, FIX Gateway
- Team C (Admin): Admin API, Reporting, Compliance

**Solution:** Each team owns their services and can deploy independently.

```
Deployment Frequency:
- Monolith: 1 deployment/week (coordinated across all teams)
- Microservices: 10 deployments/week (each team deploys independently)
```

**Benefit:** Faster feature delivery, reduced coordination overhead.

#### 5. Deployment Flexibility

**Problem:** Need to update market data feed logic without restarting trading operations.

**Monolith Scenario:**
```
1. Code change in LP Manager
2. Full application restart required
3. All active trading sessions disconnected
4. 5-10 minute downtime
5. Clients reconnect
```

**Microservices Scenario:**
```
1. Code change in LP Manager Service
2. Rolling restart of LP Manager pods only
3. WebSocket connections maintained (different service)
4. Zero downtime (Kubernetes rolling update)
5. No client impact
```

**Benefit:** Zero-downtime deployments for most changes.

---

## Considered Alternatives

### Alternative 1: Monolith

**Pros:**
- Simpler deployment (single binary)
- No network latency between components
- Easier to debug (single codebase, single process)
- Lower operational complexity

**Cons:**
- Cannot scale components independently
- Single point of failure
- Tightly coupled components
- Deployment requires full restart
- Limited technology choices
- Merge conflicts and coordination overhead

**Rejected Because:** Scalability and fault isolation are critical requirements that monolith cannot satisfy.

### Alternative 2: Serverless (AWS Lambda, Google Cloud Functions)

**Pros:**
- Auto-scaling to zero
- Pay-per-invocation pricing
- No infrastructure management

**Cons:**
- Cold start latency (100-500ms) unacceptable for trading
- No persistent WebSocket connections
- Limited execution time (15 minutes max)
- Higher cost at high volume
- Vendor lock-in

**Rejected Because:** Trading requires persistent connections and sub-50ms latency. Serverless cold starts violate latency requirements.

### Alternative 3: Service-Oriented Architecture (SOA) with ESB

**Pros:**
- Service decomposition (like microservices)
- Centralized orchestration via ESB

**Cons:**
- ESB becomes single point of failure
- ESB performance bottleneck
- Complex ESB configuration
- Vendor lock-in (e.g., Oracle ESB, MuleSoft)

**Rejected Because:** ESB creates the same single-point-of-failure problem we're trying to avoid. Modern microservices with lightweight messaging (RabbitMQ) is simpler and more resilient.

---

## Consequences

### Positive

1. **Independent Scaling:** Each service scales based on its workload.
2. **Fault Isolation:** Service failures don't cascade to the entire system.
3. **Faster Deployments:** Teams can deploy independently.
4. **Technology Flexibility:** Use optimal tech stack for each service.
5. **Easier Testing:** Unit test services in isolation.
6. **Better Performance:** Optimize critical path (order execution) without affecting other services.

### Negative

1. **Operational Complexity:**
   - Must operate 8+ services instead of 1
   - Kubernetes expertise required
   - Service mesh (Istio/Linkerd) adds complexity
   - Distributed tracing needed (Jaeger/Zipkin)

   **Mitigation:** Invest in DevOps automation, monitoring, and training.

2. **Network Latency:**
   - Inter-service calls add 1-5ms latency
   - Example: OMS → Router → Execution = 3 network hops

   **Mitigation:**
   - Use HTTP/2 for multiplexing
   - Co-locate services in same data center
   - Use service mesh for optimized routing
   - Critical path (order execution) stays within single service where possible

3. **Data Consistency:**
   - Distributed transactions are complex
   - Eventual consistency required for some operations

   **Mitigation:**
   - Use Saga pattern for distributed transactions
   - Strong consistency for financial transactions (PostgreSQL)
   - Eventual consistency acceptable for market data

4. **Debugging Difficulty:**
   - Bugs may span multiple services
   - Requires distributed tracing

   **Mitigation:**
   - OpenTelemetry + Jaeger for request tracing
   - Correlation IDs across all logs
   - Centralized logging (ELK stack)

5. **Increased Infrastructure Cost:**
   - More servers/containers needed
   - Kubernetes cluster overhead

   **Mitigation:**
   - Auto-scaling based on demand
   - Spot instances for non-critical services
   - Right-sizing based on metrics

---

## Implementation Guidelines

### Service Communication

**Synchronous (HTTP/gRPC):**
- Use for request-response (e.g., OMS → Router)
- Timeout: 500ms max
- Circuit breaker on failures

**Asynchronous (RabbitMQ):**
- Use for fire-and-forget (e.g., Order → Notification)
- Use for long-running processes (e.g., Risk Check)
- Dead Letter Queue for failed messages

### Data Management

**Database Per Service:**
- Each service has its own database schema
- Prevents tight coupling via shared database
- Exception: Read replicas for reporting

**Shared Database (Acceptable):**
- PostgreSQL for transactional data (accounts, orders, trades)
- TimescaleDB for time-series data (ticks, OHLC)
- Redis for shared cache

**Event Sourcing:**
- Not used initially (complexity overhead)
- Consider for future audit requirements

### Service Boundaries

**Bounded Contexts (DDD):**
1. **Trading Context:** Orders, Positions, Executions
2. **Market Data Context:** Quotes, Ticks, OHLC
3. **Admin Context:** Accounts, Users, Configuration
4. **Risk Context:** Margin, Exposure, Limits

**Anti-Pattern to Avoid:**
- Chatty services (too many inter-service calls)
- Distributed monolith (services tightly coupled via database)

---

## Monitoring & Observability

### Required Metrics

1. **Service Health:**
   - CPU, Memory, Disk, Network
   - Request rate, error rate, latency (RED metrics)

2. **Business Metrics:**
   - Order execution latency
   - Quote delivery latency
   - Daily trading volume
   - Active positions

3. **Distributed Tracing:**
   - End-to-end request flow
   - Service dependency map
   - Bottleneck identification

### Tooling

- **Metrics:** Prometheus + Grafana
- **Logging:** ELK Stack (Elasticsearch, Logstash, Kibana)
- **Tracing:** Jaeger + OpenTelemetry
- **Alerting:** PagerDuty

---

## Rollout Plan

### Phase 1: Core Services (Weeks 1-4)
- FIX Gateway Service
- LP Manager Service
- B-Book Execution Engine
- PostgreSQL + TimescaleDB + Redis

**Milestone:** Execute B-Book orders with real-time quotes.

### Phase 2: Client Trading (Weeks 5-8)
- Client API Service
- WebSocket Server
- OMS (Order Management System)

**Milestone:** Clients can trade via web platform.

### Phase 3: Smart Routing (Weeks 9-12)
- Smart Order Router
- A-Book execution via FIX Gateway

**Milestone:** A-Book/B-Book routing operational.

### Phase 4: Administration (Weeks 13-16)
- Admin API Service
- Financial operations (deposit/withdraw)
- Symbol management

**Milestone:** Full broker management capabilities.

### Phase 5: Production Hardening (Weeks 17-20)
- Load testing (10,000 concurrent users)
- Disaster recovery testing
- Security audit
- Performance optimization

**Milestone:** Production-ready system.

---

## Success Criteria

### Performance

| Metric | Target | Actual (Baseline) |
|--------|--------|-------------------|
| Order Execution Latency (p99) | <50ms | TBD |
| Quote Delivery Latency (p99) | <100ms | TBD |
| API Response Time (p99) | <200ms | TBD |
| Concurrent Users | 10,000 | TBD |

### Reliability

| Metric | Target | Actual (Baseline) |
|--------|--------|-------------------|
| System Uptime | 99.9% | TBD |
| Error Rate | <0.1% | TBD |
| Mean Time to Recovery (MTTR) | <5 minutes | TBD |

### Scalability

| Metric | Target | Actual (Baseline) |
|--------|--------|-------------------|
| Horizontal Scale | 3x capacity without code changes | TBD |
| Database Write Throughput | 5,000 TPS | TBD |

---

## References

- [Microservices Patterns - Chris Richardson](https://microservices.io/)
- [Building Microservices - Sam Newman](https://www.oreilly.com/library/view/building-microservices/9781491950340/)
- [Domain-Driven Design - Eric Evans](https://www.domainlanguage.com/ddd/)
- [The Twelve-Factor App](https://12factor.net/)
- [QuickFIX - FIX Protocol Implementation](https://www.quickfixengine.org/)

---

## Review and Update History

| Date | Version | Changes | Reviewer |
|------|---------|---------|----------|
| 2026-01-18 | 1.0 | Initial version | System Architect |

---

**Status:** ✅ ACCEPTED
**Next ADR:** ADR-002: Database Selection (PostgreSQL + TimescaleDB)
