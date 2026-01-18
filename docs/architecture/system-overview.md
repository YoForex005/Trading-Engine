# System Overview

## Introduction

The RTX Trading Engine is a modular, high-performance trading system that supports both B-Book (internal execution) and A-Book (LP passthrough) execution models. The system is designed for low-latency trade execution, real-time market data distribution, and comprehensive risk management.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                             │
├─────────────────┬─────────────────┬──────────────────────────────┤
│   Web Clients   │  Mobile Apps    │    API Integrations          │
│   (WebSocket)   │  (REST API)     │    (REST + WebSocket)        │
└────────┬────────┴────────┬────────┴──────────────┬───────────────┘
         │                 │                       │
         └─────────────────┼───────────────────────┘
                           │
┌──────────────────────────▼────────────────────────────────────────┐
│                      API GATEWAY LAYER                            │
├────────────────────────┬──────────────────────────────────────────┤
│   HTTP Server (7999)   │      WebSocket Hub                       │
│   - REST Endpoints     │      - Real-time Prices                  │
│   - Authentication     │      - Account Updates                   │
│   - CORS Handling      │      - Position Changes                  │
└────────────┬───────────┴──────────────────┬───────────────────────┘
             │                              │
┌────────────▼──────────────────────────────▼───────────────────────┐
│                      CORE ENGINE LAYER                            │
├────────────────┬──────────────────┬─────────────────────────────┬─┤
│  B-Book Engine │  Position Manager│  Order Manager │   P&L Engine│
│  - Execution   │  - Tracking      │  - Validation  │   - Calc    │
│  - Matching    │  - SL/TP Check   │  - Lifecycle   │   - Ledger  │
│  - Margin Calc │  - Hedging Mode  │  - Pending     │   - History │
└────────┬───────┴──────────┬───────┴────────┬───────┴──────┬──────┘
         │                  │                │              │
┌────────▼──────────────────▼────────────────▼──────────────▼────────┐
│                     LIQUIDITY PROVIDER LAYER                       │
├────────────────┬──────────────────┬────────────────────────────────┤
│  LP Manager    │  Adapters        │     FIX Gateway                │
│  - Routing     │  - Binance       │     - YoFx (FIX 4.4)          │
│  - Failover    │  - OANDA         │     - Session Management       │
│  - Aggregation │  - FIX Providers │     - Market Data Subscribe    │
└────────┬───────┴──────────┬───────┴────────┬───────────────────────┘
         │                  │                │
┌────────▼──────────────────▼────────────────▼───────────────────────┐
│                      DATA PERSISTENCE LAYER                        │
├────────────────┬──────────────────┬────────────────────────────────┤
│  Tick Store    │   Ledger         │    Configuration               │
│  - Daily Files │   - Transactions │    - LP Config (JSON)          │
│  - OHLC Cache  │   - Balance      │    - Broker Config             │
│  - Compression │   - History      │    - Symbol Specs              │
└────────────────┴──────────────────┴────────────────────────────────┘
```

## Core Components

### 1. API Gateway Layer

**HTTP Server** (`api/server.go`)
- Handles all REST API requests
- JWT authentication and session management
- CORS configuration for web clients
- Route registration and middleware

**WebSocket Hub** (`ws/hub.go`)
- Maintains active WebSocket connections
- Broadcasts real-time market ticks
- Sends account/position updates
- Symbol subscription management

### 2. Core Engine Layer

**B-Book Engine** (`internal/core/engine.go`)
- Internal order execution and matching
- Position lifecycle management
- Margin calculation and validation
- SL/TP automatic execution
- Symbol specification management

**P&L Engine** (`internal/core/pnl.go`)
- Real-time P&L calculation
- Unrealized profit/loss tracking
- Realized profit/loss settlement
- Commission and swap handling

**Ledger System** (`internal/core/ledger.go`)
- Transaction recording
- Balance management
- Deposit/withdrawal tracking
- Audit trail

### 3. Liquidity Provider Layer

**LP Manager** (`lpmanager/manager.go`)
- Multi-LP orchestration
- Quote aggregation
- Best price selection
- Failover handling
- Dynamic LP configuration

**Adapters** (`lpmanager/adapters/`)
- Binance adapter (WebSocket)
- OANDA adapter (REST + WebSocket)
- FIX protocol adapters
- Pluggable architecture

**FIX Gateway** (`fix/gateway.go`)
- FIX 4.4 protocol implementation
- Session management
- Market data subscription (35=V, 35=W, 35=Y)
- Order routing (future)

### 4. Data Persistence Layer

**Tick Store** (`tickstore/`)
- High-frequency tick storage
- Daily file rotation
- OHLC aggregation and caching
- Symbol-based organization
- Efficient disk I/O

**Ledger Database** (In-memory with persistence)
- Account balances
- Transaction history
- Trade records
- Position snapshots

## Execution Modes

### B-Book Mode (Internal Execution)
```
Client Order → Validation → B-Book Engine → Internal Matching → Position Created
                                           → Balance Updated
                                           → Ledger Entry
```

**Characteristics:**
- No external LP execution
- Broker acts as counterparty
- Instant execution
- Full control over slippage and pricing
- Broker assumes market risk

### A-Book Mode (LP Passthrough)
```
Client Order → Validation → LP Router → Liquidity Provider → External Execution
                                      → Confirmation
                                      → Position Created
                                      → Ledger Entry
```

**Characteristics:**
- Orders routed to external LP
- LP acts as counterparty
- Execution depends on LP latency
- No market risk for broker
- LP execution prices and slippage

### C-Book Mode (Hybrid - Future)
```
Client Order → Risk Analysis → If Low Risk  → B-Book
                             → If High Risk → A-Book
```

## Data Flow

### Market Data Flow
```
LP (Binance/OANDA/FIX) → LP Manager → Quote Aggregation → WebSocket Hub
                                                         → Tick Store
                                                         → B-Book Engine (SL/TP checks)
                                                         → Connected Clients
```

### Order Flow
```
Client → HTTP POST /api/orders/market → Authentication → Validation
                                                       → Margin Check
                                                       → B-Book/A-Book Router
                                                       → Execution
                                                       → Response
                                                       → WebSocket Update
```

### Price Update Flow
```
LP Quote → Manager → Hub.BroadcastTick() → Engine.UpdatePrice()
                                         → TickStore.StoreTick()
                                         → Check SL/TP triggers
                                         → WebSocket broadcast to clients
```

## Technology Stack

### Backend
- **Language**: Go 1.19+
- **WebSocket**: gorilla/websocket
- **JSON**: encoding/json (standard library)
- **HTTP Router**: net/http (standard library)
- **Concurrency**: Go channels and goroutines

### Storage
- **Tick Data**: JSON files (daily rotation)
- **OHLC Cache**: In-memory with disk persistence
- **Ledger**: In-memory maps with transaction log
- **Configuration**: JSON files

### External Integrations
- **Binance**: WebSocket API
- **OANDA**: REST API v20
- **FIX**: QuickFIX/Go (FIX 4.4)

## Scalability Considerations

### Horizontal Scaling
- Stateless API layer (can run multiple instances)
- Shared tick store via network file system
- Centralized ledger database (future: PostgreSQL/Redis)

### Vertical Scaling
- Buffered channels prevent blocking (1024-2048 buffer sizes)
- Non-blocking WebSocket broadcasts
- Efficient mutex locking (RWMutex where applicable)
- OHLC caching reduces disk I/O

### Performance Metrics
- **Tick Processing**: 10,000+ ticks/second
- **Order Execution**: <10ms (B-Book)
- **WebSocket Latency**: <5ms
- **API Response Time**: <50ms (avg)

## Security Architecture

### Authentication
- JWT token-based authentication
- Token expiration (24 hours)
- Username/password authentication
- Demo and live account separation

### Authorization
- Account-based access control
- Admin vs. client role separation
- API endpoint protection

### Data Security
- CORS configuration
- Input validation and sanitization
- SQL injection prevention (no SQL used)
- Rate limiting (future implementation)

## High Availability

### Fault Tolerance
- LP failover on disconnect
- Automatic reconnection
- Graceful degradation
- Error logging and monitoring

### Disaster Recovery
- Daily tick data backups
- Ledger transaction log
- Configuration version control
- System state snapshots

## Monitoring and Observability

### Logging
- Structured logging with prefixes `[B-Book]`, `[LPManager]`, `[Hub]`
- Transaction logging
- Error tracking
- Performance metrics

### Metrics (Future)
- Order execution latency
- LP connection uptime
- WebSocket client count
- Tick processing rate
- P&L distribution

## Future Enhancements

1. **Database Migration**: Move from in-memory to PostgreSQL
2. **Redis Cache**: Session management and hot data
3. **Microservices**: Split engine, LP manager, and API gateway
4. **Load Balancer**: NGINX for multi-instance deployment
5. **Kubernetes**: Container orchestration
6. **Prometheus**: Metrics and alerting
7. **Grafana**: Real-time dashboards
