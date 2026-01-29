# FIX 4.4 Production Enhancement Implementation Plan

## Executive Summary

This document outlines the comprehensive enhancement plan for the FIX 4.4 implementation, transforming it from a functional prototype into a production-grade trading gateway.

**Current State**: Basic FIX connectivity with limited message support
**Target State**: Production-ready multi-LP FIX gateway with full message support, recovery, monitoring

## Phase 1: Core Message Handling Enhancement (Priority: CRITICAL)

### 1.1 Session-Level Message Handlers

**Location**: `internal/message/session/`

#### Files to Create:
- `handlers.go` - Core session message handlers
- `resend.go` - ResendRequest and gap fill logic
- `sequence.go` - Sequence number management
- `reject.go` - Reject message handling

#### Implementation Details:

**ResendRequest Handling (35=2)**
```go
type ResendHandler struct {
    msgStore    MessageStore
    maxResendGap int
}

func (h *ResendHandler) HandleResendRequest(session *Session, beginSeqNo, endSeqNo int) error {
    // 1. Validate sequence range
    // 2. Retrieve messages from store
    // 3. Resend messages with PossDupFlag=Y
    // 4. Send SequenceReset for gap-filled messages
}
```

**SequenceReset Handling (35=4)**
- GapFill mode (tag 123=Y)
- Reset mode (tag 123=N)
- Validation against expected sequence

**TestRequest/Heartbeat Enhancement**
- Automatic TestRequest on heartbeat timeout
- Configurable timeout thresholds
- Disconnect on multiple missed heartbeats

### 1.2 Gap Detection & Recovery

**Automatic Gap Detection**
```go
type SequenceMonitor struct {
    expectedSeqNum int
    gapTimeout     time.Duration
}

func (m *SequenceMonitor) CheckMessage(receivedSeqNum int) GapStatus {
    // Returns: NoGap, GapDetected, DuplicateDetected
}
```

**Recovery Strategy**:
1. Detect gap in incoming sequence
2. Wait 500ms for out-of-order delivery
3. Send ResendRequest for missing range
4. Queue subsequent messages
5. Replay queue after gap filled

### 1.3 Duplicate Detection

**Implementation**:
- Track last 1000 sequence numbers per session
- Check PossDupFlag (tag 43)
- Compare OrigSendingTime (tag 122) vs SendingTime (tag 52)
- Idempotent processing for duplicates

## Phase 2: Enhanced Market Data System (Priority: HIGH)

### 2.1 Full Market Data Message Support

**Location**: `internal/message/marketdata/`

#### Files to Create:
- `subscription_manager.go` - Subscription lifecycle
- `snapshot_handler.go` - Full market data snapshot (35=W)
- `incremental_handler.go` - Incremental updates (35=X)
- `book_builder.go` - Order book construction
- `quote_handler.go` - QuoteRequest/Quote (35=R/35=S)

#### Message Types to Implement:

**MarketDataSnapshot (35=W) - Enhanced**
```go
type MDSnapshot struct {
    MDReqID     string
    Symbol      string
    Entries     []MDEntry
    LastUpdate  time.Time
}

type MDEntry struct {
    Type        MDEntryType // Bid, Offer, Trade, OpeningPrice, ClosingPrice
    Price       decimal.Decimal
    Size        decimal.Decimal
    Position    int     // Position in book (0=best bid/ask)
    Time        time.Time
    QuoteID     string  // For quote-based pricing
}
```

**MarketDataIncremental (35=X)**
```go
type MDIncremental struct {
    MDReqID     string
    Updates     []MDUpdate
}

type MDUpdate struct {
    Action      MDUpdateAction // New, Change, Delete
    Entry       MDEntry
}
```

**Subscription Types**:
- Top-of-book (tag 264=1)
- Full depth (tag 264=0 or N)
- Trade ticks (tag 269=2)
- Opening/Closing prices (tag 269=4/5)

### 2.2 Order Book Management

```go
type OrderBook struct {
    Symbol      string
    Bids        *PriceLevel // Red-black tree for performance
    Asks        *PriceLevel
    LastUpdate  time.Time
    mutex       sync.RWMutex
}

func (b *OrderBook) ApplySnapshot(snapshot MDSnapshot)
func (b *OrderBook) ApplyIncremental(update MDIncremental)
func (b *OrderBook) GetTopOfBook() (bid, ask *MDEntry)
func (b *OrderBook) GetDepth(levels int) (bids, asks []MDEntry)
```

### 2.3 Quote Request/Response

**QuoteRequest (35=R)**
- Request for indicative or firm quotes
- Single or multiple instruments
- Spot or forward quotes

**Quote (35=S)**
- Bid/Offer prices
- Quote validity period
- Minimum and maximum quantities

## Phase 3: Advanced Order Management (Priority: HIGH)

### 3.1 Missing Order Messages

**Location**: `internal/message/order/`

#### Files to Create:
- `handlers.go` - Order message handlers
- `modify.go` - OrderCancelReplaceRequest logic
- `status.go` - Order status tracking
- `validator.go` - Pre-send validation

**OrderCancelReplaceRequest (35=G)**
```go
type OrderModifyRequest struct {
    OrigClOrdID string
    ClOrdID     string
    Symbol      string
    Side        Side
    OrderQty    decimal.Decimal
    Price       decimal.Decimal  // New price
    OrdType     OrderType
}

func (g *Gateway) ModifyOrder(req OrderModifyRequest) error {
    // 1. Validate original order exists
    // 2. Generate new ClOrdID
    // 3. Send 35=G message
    // 4. Track pending modify
}
```

**OrderCancelReject (35=9)**
```go
type OrderCancelReject struct {
    OrderID         string
    ClOrdID         string
    OrigClOrdID     string
    CxlRejReason    CancelRejectReason
    CxlRejResponseTo ResponseType
    Text            string
}

// Reasons: TooLateToCancel, UnknownOrder, BrokerOption, etc.
```

**OrderMassStatusRequest (35=AF) - Enhanced**
- Filter by symbol
- Filter by side
- Filter by order status
- Time range filtering

### 3.2 Order State Machine

```go
type OrderState int

const (
    OrderStatePendingNew OrderState = iota
    OrderStateNew
    OrderStatePartiallyFilled
    OrderStateFilled
    OrderStatePendingCancel
    OrderStateCanceled
    OrderStatePendingReplace
    OrderStateReplaced
    OrderStateRejected
    OrderStateExpired
)

type OrderTracker struct {
    orders      map[string]*Order // ClOrdID -> Order
    lpOrders    map[string]string // OrderID -> ClOrdID
    mutex       sync.RWMutex
}

func (t *OrderTracker) TrackOrder(order *Order)
func (t *OrderTracker) UpdateState(clOrdID string, newState OrderState)
func (t *OrderTracker) GetOrder(clOrdID string) (*Order, bool)
```

### 3.3 Pre-Send Validation

```go
type OrderValidator struct {
    symbolValidator SymbolValidator
    priceValidator  PriceValidator
    qtyValidator    QuantityValidator
}

func (v *OrderValidator) ValidateNewOrder(order NewOrderRequest) error {
    // 1. Symbol exists and is tradable
    // 2. Price within collar bands
    // 3. Quantity within min/max limits
    // 4. Sufficient margin/credit
    // 5. Trading hours check
    // 6. Duplicate ClOrdID check
}
```

## Phase 4: Robust Session Management (Priority: HIGH)

### 4.1 Enhanced Session Lifecycle

**Location**: `internal/session/`

#### Files to Create:
- `manager.go` - Session lifecycle coordinator
- `scheduler.go` - Trading hours scheduling
- `recovery.go` - Session recovery logic
- `heartbeat.go` - Enhanced heartbeat monitoring

**Session Scheduler**
```go
type SessionScheduler struct {
    tradingHours map[string]TradingSchedule
}

type TradingSchedule struct {
    TimeZone    *time.Location
    OpenTime    time.Time
    CloseTime   time.Time
    DaysOfWeek  []time.Weekday
    Holidays    []time.Time
}

func (s *SessionScheduler) IsMarketOpen(sessionID string) bool
func (s *SessionScheduler) NextOpenTime(sessionID string) time.Time
func (s *SessionScheduler) ShouldAutoConnect(sessionID string) bool
```

**Connection State Machine**
```go
type ConnectionState int

const (
    StateDisconnected ConnectionState = iota
    StateConnecting
    StateConnected
    StateLoggingIn
    StateLoggedIn
    StateLoggingOut
    StateReconnecting
)
```

### 4.2 Automatic Reconnection

```go
type ReconnectionPolicy struct {
    MaxAttempts     int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    BackoffFactor   float64
    ResetOnSuccess  bool
}

type Reconnector struct {
    policy          ReconnectionPolicy
    attempt         int
    lastAttempt     time.Time
}

func (r *Reconnector) ShouldReconnect() bool
func (r *Reconnector) NextAttemptDelay() time.Duration
func (r *Reconnector) RecordAttempt(success bool)
```

### 4.3 Session Recovery After Crash

**Crash Recovery Steps**:
1. Load persisted sequence numbers
2. Reconnect to LP
3. Logon with stored sequences (no reset flag)
4. LP may send ResendRequest for missed messages
5. Resend stored messages if requested
6. Resume normal operation

**Persistence Strategy**:
```go
type SessionStore interface {
    SaveSequences(sessionID string, outSeq, inSeq int) error
    LoadSequences(sessionID string) (outSeq, inSeq int, error)
    SaveMessage(sessionID string, seqNum int, msg []byte) error
    LoadMessages(sessionID string, beginSeq, endSeq int) ([][]byte, error)
    ClearMessages(sessionID string, beforeSeq int) error
}

// Implementations:
// - FileStore (current)
// - SQLiteStore (recommended)
// - PostgreSQLStore (production)
```

## Phase 5: Comprehensive Error Handling (Priority: HIGH)

### 5.1 Reject Message Handling

**Location**: `internal/message/reject/`

**Reject Types**:

```go
type RejectType int

const (
    SessionReject         // 35=3
    BusinessMessageReject // 35=j
    OrderCancelReject     // 35=9
    MarketDataReject      // 35=Y
)

type RejectHandler struct {
    logger      Logger
    alerter     Alerter
    metrics     MetricsRecorder
}

func (h *RejectHandler) HandleSessionReject(msg string)
func (h *RejectHandler) HandleBusinessReject(msg string)
func (h *RejectHandler) HandleOrderReject(msg string)
func (h *RejectHandler) HandleMarketDataReject(msg string)
```

**SessionReject (35=3)**
- Invalid tag number
- Required tag missing
- Tag not defined for message type
- Tag value invalid
- Incorrect data format
- Incorrect checksum

**BusinessMessageReject (35=j)**
- Unsupported message type
- Conditionally required field missing
- Not authorized
- Throttle limit exceeded

### 5.2 Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    state           CircuitState
    failureThreshold int
    successThreshold int
    timeout         time.Duration
    failures        int
    successes       int
    lastFailure     time.Time
}

func (cb *CircuitBreaker) Allow() bool
func (cb *CircuitBreaker) RecordSuccess()
func (cb *CircuitBreaker) RecordFailure()
```

**Use Cases**:
- Too many order rejects -> Stop sending orders
- Repeated connection failures -> Back off reconnection
- Market data subscription failures -> Delay resubscription

## Phase 6: Performance Optimization (Priority: MEDIUM)

### 6.1 Message Parsing Optimization

**Current**: String operations with `strings.Split()`
**Optimized**: Byte-level parsing with zero allocations

```go
type FastParser struct {
    buffer   []byte
    tags     map[int]*TagPosition
}

type TagPosition struct {
    Start int
    End   int
}

func (p *FastParser) Parse(msg []byte) error {
    // Single pass through message
    // No allocations
    // Direct byte slice references
}

func (p *FastParser) GetTag(tag int) []byte
func (p *FastParser) GetTagInt(tag int) (int, error)
func (p *FastParser) GetTagFloat(tag int) (float64, error)
```

**Performance Target**: 10x faster than string-based parsing

### 6.2 Connection Pooling

```go
type ConnectionPool struct {
    sessions    map[string]*SessionPool
    maxPerLP    int
}

type SessionPool struct {
    connections []*Connection
    available   chan *Connection
    inUse       map[*Connection]bool
}

func (p *ConnectionPool) Get(sessionID string) (*Connection, error)
func (p *ConnectionPool) Release(conn *Connection)
func (p *ConnectionPool) Health() PoolHealth
```

**Benefits**:
- Multiple concurrent order streams
- Dedicated market data connections
- Failover connections

### 6.3 Batching for High-Frequency

```go
type MessageBatcher struct {
    queue       chan Message
    batchSize   int
    batchTimeout time.Duration
}

func (b *MessageBatcher) Send(msg Message)
func (b *MessageBatcher) Flush() error
```

**Use Cases**:
- Batch multiple orders (within same symbol)
- Combine market data subscriptions
- Reduce network round-trips

## Phase 7: Testing & Validation (Priority: MEDIUM)

### 7.1 FIX Conformance Testing

**Location**: `cmd/tests/conformance/`

**Test Suites**:
- Session-level message tests
- Application message tests
- Sequence number scenarios
- Gap detection and recovery
- Reject handling
- Performance benchmarks

```go
type ConformanceTest struct {
    Name        string
    Description string
    Steps       []TestStep
    Validator   TestValidator
}

type TestStep struct {
    Action      StepAction // Send, Receive, Wait
    Message     string
    ExpectedResponse string
    Timeout     time.Duration
}
```

### 7.2 LP Simulator

```go
type LPSimulator struct {
    port        int
    sessions    map[string]*SimSession
    behavior    SimulatorBehavior
}

type SimulatorBehavior struct {
    RejectOrders        bool
    SendMarketData      bool
    SimulateLatency     time.Duration
    InduceGaps          bool
    DropHeartbeats      bool
}

func (s *LPSimulator) Start()
func (s *LPSimulator) HandleLogon(session *SimSession)
func (s *LPSimulator) HandleOrder(session *SimSession, msg string)
```

### 7.3 Scenario Testing

**Scenarios**:
1. Normal trading flow
2. Network disconnection during order
3. Sequence gap during market data
4. LP restart (sequence reset)
5. Rapid order submissions
6. Market data flood
7. Heartbeat timeout
8. Duplicate message handling

## Phase 8: Monitoring & Diagnostics (Priority: MEDIUM)

### 8.1 Metrics Collection

**Location**: `internal/metrics/`

```go
type FIXMetrics struct {
    MessagesReceived   *prometheus.CounterVec
    MessagesSent       *prometheus.CounterVec
    MessageLatency     *prometheus.HistogramVec
    OrderRejects       *prometheus.CounterVec
    SessionUptime      *prometheus.GaugeVec
    SequenceGaps       *prometheus.CounterVec
    HeartbeatMisses    *prometheus.CounterVec
}

func (m *FIXMetrics) RecordMessageReceived(msgType string)
func (m *FIXMetrics) RecordLatency(msgType string, duration time.Duration)
func (m *FIXMetrics) RecordReject(rejectType string, reason string)
```

**Key Metrics**:
- Message rate (in/out per second)
- Order-to-ack latency (p50, p95, p99)
- Market data tick-to-process latency
- Sequence gap count
- Reject rate by type
- Session availability

### 8.2 Structured Logging

```go
type FIXLogger struct {
    logger      *zap.Logger
    piiRedactor *PIIRedactor
}

func (l *FIXLogger) LogMessageSent(sessionID, msgType string, msg []byte)
func (l *FIXLogger) LogMessageReceived(sessionID, msgType string, msg []byte)
func (l *FIXLogger) LogError(sessionID string, err error, context map[string]interface{})
```

**Log Levels**:
- DEBUG: All FIX messages (with PII redaction)
- INFO: Session lifecycle events
- WARN: Rejects, gaps, heartbeat misses
- ERROR: Connection failures, unhandled messages

**PII Redaction**:
- Redact passwords (tag 554)
- Mask account numbers (tag 1)
- Sanitize free-text fields (tag 58)

### 8.3 Health Checks

```go
type HealthChecker struct {
    sessions    map[string]*Session
}

type HealthStatus struct {
    Overall     HealthState
    Sessions    map[string]SessionHealth
}

type SessionHealth struct {
    State           HealthState
    Connected       bool
    LoggedIn        bool
    LastHeartbeat   time.Time
    SequencesInSync bool
    LastError       error
}

func (h *HealthChecker) Check() HealthStatus
```

**Health Endpoint**: `GET /health`
```json
{
  "status": "healthy",
  "sessions": {
    "YOFX1": {
      "connected": true,
      "logged_in": true,
      "last_heartbeat": "2026-01-18T10:30:15Z",
      "sequences_in_sync": true,
      "uptime_seconds": 3600
    }
  }
}
```

## Phase 9: Configuration Management (Priority: LOW)

### 9.1 Dynamic Session Configuration

**Location**: `config/`

**Configuration Schema**:
```yaml
sessions:
  - id: YOFX1
    name: "YOFX Trading"
    enabled: true
    host: ${YOFX_HOST}
    port: 12336
    sender_comp_id: YOFX1
    target_comp_id: YOFX
    username: ${YOFX_USERNAME}
    password: ${YOFX_PASSWORD}
    trading_account: "50153"
    ssl: false
    heartbeat_interval: 30
    reconnection:
      max_attempts: 5
      initial_delay: 1s
      max_delay: 60s
      backoff_factor: 2.0
    trading_hours:
      timezone: "America/New_York"
      open_time: "09:30"
      close_time: "16:00"
      days_of_week: [1,2,3,4,5]
    limits:
      max_orders_per_second: 10
      max_subscriptions: 100
```

### 9.2 Environment-Based Overrides

```
FIX_STORE_DIR=./fixstore
FIX_LOG_LEVEL=INFO
FIX_ENABLE_METRICS=true
FIX_METRICS_PORT=9090
YOFX_HOST=23.106.238.138
YOFX_USERNAME=YOFX1
YOFX_PASSWORD=<secret>
```

### 9.3 Hot Reload

```go
type ConfigWatcher struct {
    configPath string
    reloadChan chan ConfigChange
}

func (w *ConfigWatcher) Watch()
func (w *ConfigWatcher) OnChange(handler func(ConfigChange))
```

**Reloadable Without Restart**:
- Trading hours
- Reconnection policy
- Rate limits
- Log level
- Enable/disable sessions

**Requires Restart**:
- Session credentials
- Host/Port changes
- SSL settings

## Phase 10: Documentation (Priority: MEDIUM)

### 10.1 FIX Message Dictionary

**Location**: `docs/fix_messages.md`

**Format**:
```markdown
## NewOrderSingle (35=D)

**Direction**: Client -> LP
**Purpose**: Submit new order to market

### Required Tags:
| Tag | Name | Type | Description | Example |
|-----|------|------|-------------|---------|
| 11 | ClOrdID | String | Unique order ID | YOFX1_1737201234567 |
| 55 | Symbol | String | Instrument symbol | EURUSD |
| 54 | Side | Char | 1=Buy, 2=Sell | 1 |
| 38 | OrderQty | Qty | Order quantity | 100000.0 |
| 40 | OrdType | Char | 1=Market, 2=Limit | 2 |

### Optional Tags:
| Tag | Name | Description |
|-----|------|-------------|
| 44 | Price | Limit price (required for OrdType=2) |
| 59 | TimeInForce | 0=Day, 1=GTC, 3=IOC, 4=FOK |

### Example:
\```
35=D|11=YOFX1_123|55=EURUSD|54=1|38=100000|40=2|44=1.08500|60=20260118-10:30:15.123
\```

### Possible Responses:
- ExecutionReport (35=8) with OrdStatus=0 (New)
- ExecutionReport (35=8) with OrdStatus=8 (Rejected)
```

### 10.2 LP Integration Guide

**Location**: `docs/lp_integration_guide.md`

**Sections**:
1. Prerequisites
2. Configuration
3. Connection testing
4. Order placement workflow
5. Market data subscription
6. Error handling
7. Troubleshooting
8. LP-specific notes

### 10.3 Troubleshooting Guide

**Location**: `docs/troubleshooting.md`

**Common Issues**:
1. **Connection timeout**
   - Firewall blocking port
   - Incorrect credentials
   - IP not whitelisted
   - Solution: Check network, verify config

2. **Sequence number mismatch**
   - Counterparty reset sequences
   - Client crashed without proper logout
   - Solution: Manual sequence reset

3. **Order rejects**
   - Invalid symbol
   - Outside trading hours
   - Insufficient margin
   - Solution: Check symbol list, trading schedule

4. **Market data not received**
   - Symbol not subscribed
   - Subscription rejected
   - Network congestion
   - Solution: Check subscription status, logs

### 10.4 Architecture Decision Records (ADRs)

**Location**: `docs/decisions/`

**ADRs to Create**:
1. ADR-001: Message Store Choice (File vs SQLite vs Postgres)
2. ADR-002: Parsing Strategy (String vs Byte-level)
3. ADR-003: Connection Pool Design
4. ADR-004: Metrics Collection Approach
5. ADR-005: Error Recovery Strategy

## Implementation Timeline

### Week 1: Critical Foundations
- [ ] Phase 1.1: ResendRequest handling
- [ ] Phase 1.2: Gap detection
- [ ] Phase 4.1: Session lifecycle
- [ ] Phase 5.1: Basic reject handling

### Week 2: Order Management
- [ ] Phase 3.1: OrderCancelReplace
- [ ] Phase 3.2: Order state machine
- [ ] Phase 3.3: Pre-send validation
- [ ] Phase 7.1: Basic conformance tests

### Week 3: Market Data
- [ ] Phase 2.1: Incremental updates
- [ ] Phase 2.2: Order book
- [ ] Phase 2.3: Quote request/response
- [ ] Phase 8.1: Metrics collection

### Week 4: Reliability & Testing
- [ ] Phase 4.2: Reconnection logic
- [ ] Phase 4.3: Crash recovery
- [ ] Phase 7.2: LP Simulator
- [ ] Phase 7.3: Scenario testing

### Week 5: Optimization & Monitoring
- [ ] Phase 6.1: Fast parser
- [ ] Phase 6.2: Connection pooling
- [ ] Phase 8.2: Structured logging
- [ ] Phase 8.3: Health checks

### Week 6: Documentation & Polish
- [ ] Phase 10.1: Message dictionary
- [ ] Phase 10.2: Integration guide
- [ ] Phase 10.3: Troubleshooting
- [ ] Phase 9.1: Configuration management

## Success Criteria

### Functional Requirements
- [ ] All FIX 4.4 session messages supported
- [ ] Complete order lifecycle (new, modify, cancel)
- [ ] Full market data (snapshot, incremental, quotes)
- [ ] Automatic gap recovery
- [ ] Session persistence across restarts

### Non-Functional Requirements
- [ ] Order-to-ack latency: p99 < 100ms
- [ ] Market data latency: p99 < 50ms
- [ ] Session availability: 99.9%
- [ ] No message loss during reconnection
- [ ] Handles 1000 orders/second per session

### Quality Requirements
- [ ] 90% unit test coverage
- [ ] All scenarios pass conformance tests
- [ ] Zero critical security issues
- [ ] PII properly redacted in logs
- [ ] Comprehensive documentation

## Risk Assessment

### High Risk
1. **Sequence number corruption**: Mitigation = SQLite store with transactions
2. **Message loss during crash**: Mitigation = WAL mode + fsync
3. **Performance degradation**: Mitigation = Benchmarking + profiling

### Medium Risk
1. **LP-specific quirks**: Mitigation = Extensive testing with LP simulator
2. **Market data flood**: Mitigation = Backpressure + circuit breaker
3. **Configuration errors**: Mitigation = Schema validation

### Low Risk
1. **Documentation drift**: Mitigation = Doc generation from code
2. **Metric accuracy**: Mitigation = Prometheus best practices

## Maintenance Plan

### Daily
- Monitor session health
- Check reject rates
- Review error logs

### Weekly
- Clean up old message stores (>7 days)
- Review performance metrics
- Check sequence number health

### Monthly
- Rotate logs
- Update LP configurations
- Performance benchmarking
- Security audit

### Quarterly
- Update FIX spec compliance
- Review and optimize hot paths
- Update documentation
- Disaster recovery drill

---

**Document Version**: 1.0
**Created**: 2026-01-18
**Author**: Backend API Developer Agent (Claude)
**Status**: Ready for Implementation
