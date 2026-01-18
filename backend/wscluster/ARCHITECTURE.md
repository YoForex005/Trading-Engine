# WebSocket Cluster Architecture

## System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    Client Layer (100k+ clients)                   │
└──────────────────────────┬───────────────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────────────┐
│                  Load Balancer (HAProxy/Nginx)                    │
│  - Layer 7 Load Balancing                                         │
│  - SSL/TLS Termination                                            │
│  - Health Checks                                                  │
│  - Sticky Sessions (IP Hash)                                      │
└──────────────────────────┬───────────────────────────────────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
┌───────▼────────┐ ┌───────▼────────┐ ┌───────▼────────┐
│  WS Node 1     │ │  WS Node 2     │ │  WS Node N     │
│                │ │                │ │                │
│ Components:    │ │ Components:    │ │ Components:    │
│ • Cluster      │ │ • Cluster      │ │ • Cluster      │
│ • PubSub       │ │ • PubSub       │ │ • PubSub       │
│ • LoadBalancer │ │ • LoadBalancer │ │ • LoadBalancer │
│ • Failover     │ │ • Failover     │ │ • Failover     │
│ • Affinity     │ │ • Affinity     │ │ • Affinity     │
│ • Metrics      │ │ • Metrics      │ │ • Metrics      │
│                │ │                │ │                │
│ 10k conns      │ │ 10k conns      │ │ 10k conns      │
└───────┬────────┘ └───────┬────────┘ └───────┬────────┘
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                ┌──────────▼───────────┐
                │   Redis Cluster      │
                │                      │
                │ • Pub/Sub Channels   │
                │ • Session Affinity   │
                │ • Node Registry      │
                │ • Metrics Store      │
                │ • Message Queue      │
                └──────────┬───────────┘
                           │
                ┌──────────▼───────────┐
                │   PostgreSQL         │
                │                      │
                │ • User Data          │
                │ • Session History    │
                │ • Analytics          │
                └──────────────────────┘
```

## Component Details

### 1. Cluster Manager (`cluster.go`)

**Responsibilities:**
- Node discovery and registration
- Health monitoring (heartbeat)
- Cluster-wide metrics aggregation
- Auto-scaling triggers
- Node lifecycle management

**Key Features:**
- Automatic node discovery via Redis
- Configurable heartbeat intervals
- Dead node detection and removal
- Metrics publishing (connections, throughput, latency)
- Auto-scaling thresholds

**Data Structures:**
```go
type Cluster struct {
    localNode     *Node
    nodes         map[string]*Node
    pubsub        *PubSubManager
    loadBalancer  *LoadBalancer
    failover      *FailoverManager
    metrics       *ClusterMetrics
}

type Node struct {
    ID              string
    Address         string
    Port            int
    Status          NodeStatus
    Connections     int64
    MaxConnections  int64
    LastHeartbeat   time.Time
    CPUUsage        float64
    MemoryUsage     float64
}
```

### 2. PubSub Manager (`pubsub.go`)

**Responsibilities:**
- Cross-node message broadcasting
- Event distribution
- Message batching and compression
- Reliable message delivery

**Key Features:**
- Redis Pub/Sub integration
- Message priority queues
- Automatic batching (configurable)
- Event subscription system
- Message acknowledgment

**Message Flow:**
```
Node 1                     Redis                     Node 2
   │                         │                         │
   │──Publish("broadcast")──▶│                         │
   │                         │──Subscribe──────────────▶│
   │                         │                         │
   │                         │◀────Receive─────────────│
   │                         │                         │
   │                         │──Deliver Message───────▶│
```

### 3. Session Affinity (`sticky_session.go`)

**Responsibilities:**
- User-to-node mapping
- Session persistence
- Connection migration support
- Strategy-based routing

**Strategies:**
- `AffinityByUserID` - Same user → same node
- `AffinityByConnection` - Connection-level affinity
- `AffinityByHash` - Consistent hashing
- `AffinityByRegion` - Geographic routing
- `AffinityNone` - Pure load balancing

**Session Lifecycle:**
```
1. New Connection → Check existing mapping
2. If exists → Route to assigned node
3. If new → Load balancer selects node
4. Store mapping in Redis
5. On disconnect → Remove mapping (or keep with TTL)
6. On node failure → Migrate to new node
```

### 4. Load Balancer (`loadbalancer.go`)

**Responsibilities:**
- Node selection for new connections
- Distribution algorithm execution
- Load imbalance detection
- Rebalancing recommendations

**Algorithms:**

| Algorithm | Use Case | Complexity |
|-----------|----------|-----------|
| Round Robin | Equal nodes | O(1) |
| Least Connections | Varying capacity | O(n) |
| Weighted | Different hardware | O(n) |
| IP Hash | Session affinity | O(1) |
| Random | Simple distribution | O(1) |
| Adaptive | Production (recommended) | O(n) |

**Adaptive Algorithm Scoring:**
```
Score = (Available Capacity × 0.4) +
        ((1 - CPU Usage) × 0.2) +
        ((1 - Memory Usage) × 0.2) +
        ((1 - Error Rate) × 0.1) +
        (Normalized Throughput × 0.1)
```

### 5. Failover Manager (`failover.go`)

**Responsibilities:**
- Dead node detection
- Connection migration
- Load rebalancing
- Graceful degradation

**Failover Strategies:**

1. **Immediate Failover**
   - Migrate all connections immediately
   - Fastest recovery (< 1s)
   - Risk: In-flight message loss

2. **Graceful Failover**
   - Wait for grace period (5s default)
   - Allow in-flight operations to complete
   - Balanced approach

3. **Delayed Failover**
   - Async migration
   - Minimal impact on active nodes
   - Slower recovery (10-30s)

**Failover Flow:**
```
1. Node failure detected (missed heartbeats)
2. Mark node as offline
3. Select target node(s) for migration
4. Get sessions from failed node (Redis)
5. Reassign sessions to healthy nodes
6. Publish failover event
7. Clients reconnect to new nodes
8. Remove dead node from cluster
```

### 6. Metrics Collector (`metrics.go`)

**Responsibilities:**
- Real-time metrics collection
- Latency tracking (P50, P95, P99)
- Historical data storage
- Alert rule evaluation
- Prometheus export

**Tracked Metrics:**

**Per-Node:**
- Active connections
- Messages sent/received
- Bytes sent/received
- Latency percentiles
- CPU/Memory usage
- Error rates
- Reconnection rates

**Cluster-Wide:**
- Total connections
- Total nodes
- Healthy nodes
- Aggregate throughput
- Average latency
- Overall error rate

## Data Flow

### 1. Connection Establishment

```
Client                     Load Balancer              Node              Redis
  │                             │                       │                 │
  │──WebSocket Upgrade────────▶│                       │                 │
  │                             │──Select Node────────▶│                 │
  │                             │                       │                 │
  │                             │◀─────Node Selected───│                 │
  │                             │                       │                 │
  │◀────────Redirect───────────│                       │                 │
  │                             │                       │                 │
  │──────────────WebSocket Connect────────────────────▶│                 │
  │                             │                       │                 │
  │                             │                       │──Store Session─▶│
  │                             │                       │                 │
  │◀────────────Connection Established────────────────│                 │
```

### 2. Message Broadcasting

```
Node 1                     Redis                     Node 2              Client
  │                         │                         │                    │
  │──Publish("BTCUSD")────▶│                         │                    │
  │   tick update          │                         │                    │
  │                         │──Broadcast─────────────▶│                    │
  │                         │                         │                    │
  │                         │                         │──Filter by sub────▶│
  │                         │                         │                    │
  │                         │                         │──Send message─────▶│
```

### 3. Failover Scenario

```
Client         Node 1 (Failed)      Node 2         Redis         Load Balancer
  │                  │                  │             │                │
  │──Message────────▶│                  │             │                │
  │                  │ (CRASH)          │             │                │
  │                  │                  │             │                │
  │◀──Connection Lost│                  │             │                │
  │                  │                  │             │                │
  │                  │                  │◀─Detect─────│                │
  │                  │                  │  Failure    │                │
  │                  │                  │             │                │
  │                  │                  │──Migrate────▶│                │
  │                  │                  │  Sessions   │                │
  │                  │                  │             │                │
  │──Reconnect──────────────────────────────────────────────────────▶│
  │                  │                  │             │                │
  │◀────────────────Route to Node 2────────────────────────────────────│
  │                  │                  │             │                │
  │──────────────────Connect───────────▶│             │                │
  │                  │                  │             │                │
  │◀─────────────Connection Restored───│             │                │
```

## Scalability Considerations

### Horizontal Scaling

**Adding Nodes:**
1. Start new node with unique ID
2. Node registers in Redis
3. Cluster auto-discovers new node
4. Load balancer includes in rotation
5. New connections distributed to new node

**Removing Nodes:**
1. Mark node for shutdown
2. Stop accepting new connections
3. Drain existing connections (graceful timeout)
4. Deregister from Redis
5. Other nodes take over load

### Vertical Scaling

**Per-Node Capacity:**
- 10,000 connections per node (default)
- Configurable via `MaxConnections`
- Limited by: file descriptors, memory, CPU

**Optimization:**
- Increase file descriptors: `ulimit -n 200000`
- Tune TCP buffers
- Use connection pooling
- Enable message batching
- Use binary protocols (MessagePack)

### Auto-Scaling

**Scale-Up Triggers:**
- Cluster utilization > 80% (default)
- Individual node > 90%
- High latency (P99 > threshold)
- Error rate spike

**Scale-Down Triggers:**
- Cluster utilization < 30% (default)
- Sustained low load (5+ minutes)
- Excess capacity (2x minimum nodes)

## Performance Targets

| Metric | Target | Maximum |
|--------|--------|---------|
| Connections per node | 10,000 | 50,000 |
| Cluster total | 100,000+ | 1,000,000+ |
| Message latency (P99) | < 10ms | < 50ms |
| Failover time | < 2s | < 5s |
| Messages per second | 100,000+ | 1,000,000+ |
| Memory per connection | 4KB | 16KB |
| CPU per 1k connections | 1% | 5% |

## Deployment Patterns

### 1. Single Region

```
Load Balancer
    ├── Node 1 (us-east-1a)
    ├── Node 2 (us-east-1a)
    ├── Node 3 (us-east-1b)
    └── Node 4 (us-east-1b)

Redis Cluster (3 masters, 3 replicas)
PostgreSQL (Primary + Read Replicas)
```

### 2. Multi-Region

```
Global Load Balancer (GeoDNS)
    │
    ├── US-East Region
    │   ├── Regional LB
    │   ├── Nodes 1-4
    │   └── Redis + PostgreSQL
    │
    ├── EU-West Region
    │   ├── Regional LB
    │   ├── Nodes 5-8
    │   └── Redis + PostgreSQL
    │
    └── APAC Region
        ├── Regional LB
        ├── Nodes 9-12
        └── Redis + PostgreSQL
```

## Monitoring & Observability

### Health Checks
- HTTP `/health` endpoint
- WebSocket ping/pong
- Redis connection check
- Database connection check

### Metrics Export
- Prometheus format
- Custom metrics endpoint
- Real-time dashboards
- Historical trends

### Logging
- Structured JSON logs
- Connection events
- Error tracking
- Performance logs

### Alerting Rules
- Node down
- High CPU/memory
- High error rate
- High latency
- Low disk space
- Redis unavailable

## Security Considerations

### Authentication
- JWT tokens in WebSocket URL
- API key authentication
- OAuth2 integration

### Authorization
- Per-connection permissions
- Channel-level access control
- Rate limiting per user

### Network Security
- TLS/SSL encryption
- DDoS protection
- IP whitelisting
- Firewall rules

### Data Security
- Message encryption
- PII data handling
- Audit logs
- Compliance (GDPR, etc.)

## Future Enhancements

1. **Service Mesh Integration** (Istio, Linkerd)
2. **Kubernetes Operator** for auto-scaling
3. **GraphQL Subscriptions** support
4. **WebRTC** for peer-to-peer
5. **Binary Protocol** (Protocol Buffers)
6. **Edge Computing** (CDN integration)
7. **AI-based Load Prediction**
8. **Blockchain** for audit trails
