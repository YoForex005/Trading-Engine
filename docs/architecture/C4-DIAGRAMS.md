# C4 Model Architecture Diagrams

**System:** RTX Trading Engine
**Date:** 2026-01-18
**Version:** 1.0

This document contains C4 model diagrams (Context, Container, Component, Code) for the trading engine architecture.

---

## C4 Model Overview

The C4 model provides a hierarchical view of the system:
- **Level 1 - Context:** System interactions with users and external systems
- **Level 2 - Container:** High-level technology choices and communication
- **Level 3 - Component:** Internal structure of key containers
- **Level 4 - Code:** Class-level diagrams (not included for brevity)

---

## Level 1: System Context Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        SYSTEM CONTEXT                                │
└─────────────────────────────────────────────────────────────────────┘

Actors:
┌─────────────┐
│   Retail    │  Uses web/mobile platform to trade Forex, CFDs, Crypto
│   Trader    │  Views real-time quotes, places orders, manages positions
└──────┬──────┘
       │
       │ HTTPS/WebSocket
       ▼
┌─────────────────────────────────────────────────────────────────┐
│                                                                  │
│              RTX TRADING ENGINE PLATFORM                         │
│                                                                  │
│  - Real-time market data streaming                              │
│  - Order execution (A-Book STP / B-Book Internal / C-Book)     │
│  - Position and risk management                                 │
│  - Account management and reporting                             │
│  - Administrative functions for broker operations               │
│                                                                  │
└──────┬──────────────────────────────────────────────┬───────────┘
       │                                              │
       │ FIX 4.4 / REST / WebSocket                  │ HTTPS (Admin)
       ▼                                              ▼
┌─────────────────────────────┐           ┌──────────────────┐
│  Liquidity Providers (LPs)  │           │  Broker Admin    │
│  - YOFX1 (FIX Trading)      │           │  - Account mgmt  │
│  - OANDA (REST API)         │           │  - Deposits/     │
│  - Binance (WebSocket)      │           │    Withdrawals   │
│  - Prime XM (FIX)           │           │  - Symbol config │
│                             │           │  - LP management │
│  Provides:                  │           │  - Reports       │
│  - Real-time price quotes   │           └──────────────────┘
│  - Order execution (A-Book) │
│  - Market data              │
└─────────────────────────────┘

External Systems:
┌──────────────────┐
│  Payment Gateway │  Bank transfers, credit card, crypto deposits/withdrawals
└──────────────────┘

┌──────────────────┐
│  Email Service   │  Transactional emails (order confirmations, withdrawals)
└──────────────────┘

┌──────────────────┐
│  SMS Provider    │  2FA, withdrawal confirmations, margin call alerts
└──────────────────┘
```

### Key Interactions

1. **Retail Trader → Trading Platform**
   - Protocol: HTTPS (REST API), WebSocket (real-time quotes)
   - Actions: Login, view quotes, place orders, manage positions
   - Response Time: <200ms (API), <100ms (WebSocket)

2. **Trading Platform → Liquidity Providers**
   - Protocol: FIX 4.4 (YOFX1, Prime XM), REST (OANDA), WebSocket (Binance)
   - Actions: Subscribe to quotes, send orders (A-Book), request positions
   - Response Time: <50ms (quote latency), <100ms (order execution)

3. **Broker Admin → Trading Platform**
   - Protocol: HTTPS (REST API)
   - Actions: Create accounts, process deposits/withdrawals, configure system
   - Response Time: <500ms (complex operations)

---

## Level 2: Container Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          RTX TRADING ENGINE - CONTAINERS                         │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────┐
│                                CLIENT LAYER                                       │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐                 │
│  │  Web Platform   │  │  Mobile Apps    │  │  MT4/MT5 API    │                 │
│  │  (React)        │  │  (React Native) │  │  (FIX/Binary)   │                 │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘                 │
│           │                    │                    │                            │
│           └────────────────────┼────────────────────┘                            │
│                                │                                                 │
└────────────────────────────────┼─────────────────────────────────────────────────┘
                                 │
                                 │ HTTPS/WSS
                                 ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY (NGINX)                                  │
│  - SSL Termination  - Rate Limiting  - Load Balancing  - CORS                   │
└────────────────────────────────┬─────────────────────────────────────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────┐
│   Client API    │  │   Admin API     │  │  WebSocket Server   │
│   (Go/Gin)      │  │   (Go/Gin)      │  │  (Go/Gorilla)       │
│                 │  │                 │  │                     │
│ - Trading ops   │  │ - Account mgmt  │  │ - Real-time quotes  │
│ - Account info  │  │ - Deposits/     │  │ - Account updates   │
│ - Market data   │  │   Withdrawals   │  │ - Position updates  │
└────────┬────────┘  └────────┬────────┘  └────────┬────────────┘
         │                    │                     │
         └────────────────────┼─────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                     ORDER MANAGEMENT SYSTEM (OMS)                                 │
│  Container: Go Application                                                        │
│  - Order validation and lifecycle management                                     │
│  - Position tracking and P&L calculation                                         │
│  - Risk checks (margin, limits, exposure)                                        │
└────────────────────────────────┬─────────────────────────────────────────────────┘
                                 │
                                 ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                          SMART ORDER ROUTER                                       │
│  Container: Go Application                                                        │
│  - Client classification (VIP, profitable, beginner)                             │
│  - Position size analysis                                                        │
│  - Symbol volatility check                                                       │
│  - LP availability check                                                         │
│  - Risk-based routing decision (A/B/C-Book)                                      │
└──────────────┬───────────────────────────┬───────────────────────────────────────┘
               │                           │
       A-BOOK │                   B-BOOK │                   C-BOOK
               ▼                           ▼
┌──────────────────────────┐  ┌──────────────────────────┐
│   FIX Gateway Service    │  │  B-Book Execution Engine │
│   (Go/quickfix-go)       │  │  (Go)                    │
│                          │  │                          │
│ - FIX 4.4 sessions       │  │ - Internal order book    │
│ - Multiple LPs           │  │ - Position management    │
│ - Reconnection logic     │  │ - P&L calculation        │
│ - Market data (35=W)     │  │ - Margin calculation     │
└───────┬──────────────────┘  └──────────┬───────────────┘
        │                                │
        │ FIX 4.4                        │
        ▼                                │
┌──────────────────────────┐            │
│  Liquidity Providers     │            │
│  - YOFX1 (FIX)           │            │
│  - OANDA (REST)          │            │
│  - Binance (WebSocket)   │            │
└──────────────────────────┘            │
                                        │
        ┌───────────────────────────────┴───────────────────────────────┐
        │                                                                │
        ▼                                                                ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              LP MANAGER SERVICE                                  │
│  Container: Go Application                                                       │
│  - Quote aggregation from multiple LPs                                          │
│  - Failover logic (primary/secondary LP)                                        │
│  - Spread markup application                                                    │
│  - Quote normalization and validation                                           │
└────────────────────────────────┬────────────────────────────────────────────────┘
                                 │
                                 │ Publish Quotes
                                 ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                          DATA & MESSAGING LAYER                                   │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ PostgreSQL   │  │ TimescaleDB  │  │ Redis Cluster│  │  RabbitMQ    │       │
│  │              │  │              │  │              │  │   Cluster    │       │
│  │ - Accounts   │  │ - Ticks      │  │ - Quotes     │  │ - Orders     │       │
│  │ - Orders     │  │ - OHLC       │  │ - Sessions   │  │ - Executions │       │
│  │ - Positions  │  │ - Aggregates │  │ - Cache      │  │ - Notifs     │       │
│  │ - Trades     │  │              │  │              │  │              │       │
│  │ - Ledger     │  │              │  │              │  │              │       │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘       │
└──────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────┐
│                       OBSERVABILITY & MONITORING                                  │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │ Prometheus   │  │   Grafana    │  │    Jaeger    │  │  ELK Stack   │       │
│  │ (Metrics)    │  │ (Dashboards) │  │ (Tracing)    │  │ (Logging)    │       │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘       │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### Container Technologies

| Container | Technology | Reason |
|-----------|-----------|--------|
| Client API | Go + Gin | High performance, concurrency, fast JSON serialization |
| Admin API | Go + Gin | Consistency with client API, strong typing |
| WebSocket Server | Go + Gorilla | Efficient goroutines for concurrent connections |
| OMS | Go | Low latency, critical path performance |
| Smart Router | Go | Fast decision logic, minimal overhead |
| FIX Gateway | Go + quickfix-go | Industry-standard FIX library, high performance |
| B-Book Engine | Go | Critical execution logic, low latency |
| LP Manager | Go | Concurrent LP connections, quote aggregation |
| PostgreSQL | RDBMS | ACID compliance for financial transactions |
| TimescaleDB | Time-series DB | Optimized for tick data and OHLC |
| Redis | In-memory cache | Sub-millisecond read latency for quotes |
| RabbitMQ | Message queue | Reliable async processing, dead-letter queues |

---

## Level 3: Component Diagram - Order Management System (OMS)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                   ORDER MANAGEMENT SYSTEM (OMS) - COMPONENTS                     │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────┐
│                              API LAYER                                            │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────┐  ┌────────────────────┐  ┌────────────────────┐        │
│  │  Order Controller  │  │ Position Controller│  │ Account Controller │        │
│  │                    │  │                    │  │                    │        │
│  │ - PlaceOrder()     │  │ - GetPositions()   │  │ - GetSummary()     │        │
│  │ - ModifyOrder()    │  │ - ClosePosition()  │  │ - GetTrades()      │        │
│  │ - CancelOrder()    │  │ - ModifySLTP()     │  │ - GetLedger()      │        │
│  └─────────┬──────────┘  └─────────┬──────────┘  └─────────┬──────────┘        │
└────────────┼─────────────────────────┼─────────────────────────┼─────────────────┘
             │                         │                         │
             ▼                         ▼                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                           BUSINESS LOGIC LAYER                                    │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                      Order Service                                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐│  │
│  │  │Order         │  │Order         │  │Order         │  │Pending Order ││  │
│  │  │Validator     │  │State Machine │  │Executor      │  │Manager       ││  │
│  │  │              │  │              │  │              │  │              ││  │
│  │  │- Volume OK?  │  │PENDING→      │  │- Route to    │  │- Trigger     ││  │
│  │  │- Symbol OK?  │  │ACCEPTED→     │  │  A/B/C-Book  │  │  conditions  ││  │
│  │  │- Price valid?│  │FILLED        │  │- Confirm fill│  │- Limit orders││  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘│  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                    Position Service                                       │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐│  │
│  │  │Position      │  │Position      │  │SL/TP         │  │Trailing Stop ││  │
│  │  │Manager       │  │Tracker       │  │Monitor       │  │Handler       ││  │
│  │  │              │  │              │  │              │  │              ││  │
│  │  │- Create pos  │  │- Update P&L  │  │- Check SL    │  │- Adjust SL   ││  │
│  │  │- Close pos   │  │- Mark prices │  │- Check TP    │  │  on price    ││  │
│  │  │- Partial     │  │- Calc equity │  │- Auto-close  │  │  movement    ││  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘│  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                       Risk Service                                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐│  │
│  │  │Margin        │  │Exposure      │  │Margin Call   │  │Stop Out      ││  │
│  │  │Calculator    │  │Monitor       │  │Detector      │  │Handler       ││  │
│  │  │              │  │              │  │              │  │              ││  │
│  │  │- Required    │  │- Symbol      │  │- Level check │  │- Force close ││  │
│  │  │  margin      │  │  exposure    │  │  (100%)      │  │  at 50%      ││  │
│  │  │- Free margin │  │- Client      │  │- Send alert  │  │- Prioritize  ││  │
│  │  │- Level calc  │  │  exposure    │  │              │  │  by size     ││  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘│  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │                      Account Service                                      │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │  │
│  │  │Account       │  │Ledger        │  │P&L           │                  │  │
│  │  │Manager       │  │Manager       │  │Calculator    │                  │  │
│  │  │              │  │              │  │              │                  │  │
│  │  │- Get account │  │- Record      │  │- Unrealized  │                  │  │
│  │  │- Update      │  │  deposits    │  │- Realized    │                  │  │
│  │  │  balance     │  │- Record      │  │- Commission  │                  │  │
│  │  │- Suspend     │  │  trades      │  │- Swap        │                  │  │
│  │  └──────────────┘  └──────────────┘  └──────────────┘                  │  │
│  └──────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────┘
             │                         │                         │
             ▼                         ▼                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            DATA ACCESS LAYER                                      │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐             │
│  │ Order Repository │  │Position Repository│ │Account Repository│             │
│  │ (PostgreSQL)     │  │ (PostgreSQL)      │ │ (PostgreSQL)     │             │
│  │                  │  │                   │ │                  │             │
│  │ - SaveOrder()    │  │ - SavePosition()  │ │ - GetAccount()   │             │
│  │ - GetOrder()     │  │ - GetPositions()  │ │ - UpdateBalance()│             │
│  │ - UpdateStatus() │  │ - UpdatePnL()     │ │ - GetLedger()    │             │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘             │
│                                                                                  │
│  ┌──────────────────┐  ┌──────────────────┐                                   │
│  │ Quote Provider   │  │ Event Publisher  │                                   │
│  │ (Redis)          │  │ (RabbitMQ)       │                                   │
│  │                  │  │                   │                                   │
│  │ - GetQuote()     │  │ - PublishOrder() │                                   │
│  │ - Subscribe()    │  │ - PublishExec()  │                                   │
│  └──────────────────┘  └──────────────────┘                                   │
└──────────────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

#### Order Service

**Order Validator:**
```go
func (v *OrderValidator) Validate(order *Order) error {
    // Check symbol exists
    if !v.symbolRepo.Exists(order.Symbol) {
        return ErrInvalidSymbol
    }

    // Check volume within limits
    spec := v.symbolRepo.GetSpec(order.Symbol)
    if order.Volume < spec.MinVolume || order.Volume > spec.MaxVolume {
        return ErrInvalidVolume
    }

    // Check price (for limit/stop orders)
    if order.Type != "MARKET" && order.Price <= 0 {
        return ErrInvalidPrice
    }

    return nil
}
```

**Order State Machine:**
```
States: PENDING → ACCEPTED → FILLING → FILLED
                        ↓
                    REJECTED

Events:
- SubmitOrder() → PENDING
- ValidateOrder() → ACCEPTED or REJECTED
- RouteOrder() → FILLING
- FillOrder() → FILLED
- RejectOrder() → REJECTED
```

**Order Executor:**
```go
func (e *OrderExecutor) Execute(order *Order) (*Position, error) {
    // Route to appropriate venue
    venue := e.router.Route(order)

    switch venue {
    case A_BOOK:
        return e.executeABook(order)
    case B_BOOK:
        return e.executeBBook(order)
    case C_BOOK:
        return e.executeCBook(order)
    }
}
```

#### Position Service

**Position Tracker:**
```go
func (pt *PositionTracker) UpdatePrices(symbol string, bid, ask float64) {
    positions := pt.getOpenPositions(symbol)

    for _, pos := range positions {
        if pos.Side == "BUY" {
            pos.CurrentPrice = bid
        } else {
            pos.CurrentPrice = ask
        }

        pos.UnrealizedPnL = pt.calculatePnL(pos)

        // Check SL/TP triggers
        pt.checkStopLoss(pos)
        pt.checkTakeProfit(pos)
    }
}
```

#### Risk Service

**Margin Calculator:**
```go
func (mc *MarginCalculator) Calculate(account *Account) (*MarginSummary, error) {
    positions := mc.positionRepo.GetOpenPositions(account.ID)

    var totalMargin float64
    var unrealizedPnL float64

    for _, pos := range positions {
        spec := mc.symbolRepo.GetSpec(pos.Symbol)
        margin := (pos.Volume * spec.ContractSize * pos.OpenPrice) / account.Leverage
        totalMargin += margin
        unrealizedPnL += pos.UnrealizedPnL
    }

    equity := account.Balance + unrealizedPnL
    freeMargin := equity - totalMargin
    marginLevel := (equity / totalMargin) * 100 // Percentage

    return &MarginSummary{
        Equity:      equity,
        Margin:      totalMargin,
        FreeMargin:  freeMargin,
        MarginLevel: marginLevel,
    }
}
```

---

## Level 3: Component Diagram - B-Book Execution Engine

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    B-BOOK EXECUTION ENGINE - COMPONENTS                          │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────────┐
│                              EXECUTION LAYER                                      │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                        Order Execution Engine                               │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐          │ │
│  │  │Market Order│  │Limit Order │  │Stop Order  │  │Stop-Limit  │          │ │
│  │  │Handler     │  │Handler     │  │Handler     │  │Handler     │          │ │
│  │  │            │  │            │  │            │  │            │          │ │
│  │  │- Instant   │  │- Add to    │  │- Monitor   │  │- Convert to│          │ │
│  │  │  fill      │  │  order book│  │  trigger   │  │  limit on  │          │ │
│  │  │- Get quote │  │- Wait for  │  │- Execute   │  │  trigger   │          │ │
│  │  │- Create pos│  │  match     │  │  on hit    │  │            │          │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘          │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                          Position Manager                                   │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐          │ │
│  │  │Position    │  │Position    │  │Partial     │  │Netting vs  │          │ │
│  │  │Creator     │  │Closer      │  │Close       │  │Hedging     │          │ │
│  │  │            │  │            │  │Handler     │  │Mode        │          │ │
│  │  │- Validate  │  │- Get exit  │  │- Reduce    │  │- HEDGING:  │          │ │
│  │  │- Deduct    │  │  price     │  │  volume    │  │  Multiple  │          │ │
│  │  │  margin    │  │- Calc P&L  │  │- Calc      │  │  positions │          │ │
│  │  │- Record    │  │- Update    │  │  partial   │  │- NETTING:  │          │ │
│  │  │  position  │  │  balance   │  │  P&L       │  │  One pos   │          │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘          │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────────┘
             │                         │                         │
             ▼                         ▼                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                          CALCULATION LAYER                                        │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                         P&L Calculator                                      │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐          │ │
│  │  │Unrealized  │  │Realized    │  │Commission  │  │Swap        │          │ │
│  │  │P&L         │  │P&L         │  │Calculator  │  │Calculator  │          │ │
│  │  │            │  │            │  │            │  │            │          │ │
│  │  │- Mark to   │  │- Close     │  │- Per lot   │  │- Daily     │          │ │
│  │  │  market    │  │  price diff│  │  fee       │  │  rollover  │          │ │
│  │  │- Live      │  │- Final P&L │  │- Entry +   │  │- Long/Short│          │ │
│  │  │  updates   │  │- Net of    │  │  exit      │  │  rates     │          │ │
│  │  │            │  │  commission│  │            │  │            │          │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘          │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                        Margin Calculator                                    │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐          │ │
│  │  │Required    │  │Free Margin │  │Margin Level│  │Margin Call │          │ │
│  │  │Margin      │  │            │  │            │  │& Stop Out  │          │ │
│  │  │            │  │            │  │            │  │            │          │ │
│  │  │- (Volume × │  │- Equity -  │  │- (Equity / │  │- MC: 100%  │          │ │
│  │  │  Contract  │  │  Used      │  │  Margin) × │  │- SO: 50%   │          │ │
│  │  │  × Price)  │  │  Margin    │  │  100       │  │- Force     │          │ │
│  │  │  /Leverage │  │            │  │            │  │  close     │          │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘          │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────────┘
             │                         │                         │
             ▼                         ▼                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            ACCOUNTING LAYER                                       │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌────────────────────────────────────────────────────────────────────────────┐ │
│  │                           Ledger Manager                                    │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐          │ │
│  │  │Deposit     │  │Withdrawal  │  │Realized PnL│  │Commission  │          │ │
│  │  │Recorder    │  │Recorder    │  │Recorder    │  │Recorder    │          │ │
│  │  │            │  │            │  │            │  │            │          │ │
│  │  │- Credit    │  │- Debit     │  │- Credit or │  │- Debit     │          │ │
│  │  │  account   │  │  account   │  │  debit     │  │  account   │          │ │
│  │  │- Record    │  │- Validate  │  │- Link to   │  │- Per trade │          │ │
│  │  │  method    │  │  balance   │  │  trade     │  │            │          │ │
│  │  └────────────┘  └────────────┘  └────────────┘  └────────────┘          │ │
│  │                                                                             │ │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐                          │ │
│  │  │Swap        │  │Adjustment  │  │Bonus       │                          │ │
│  │  │Recorder    │  │Recorder    │  │Recorder    │                          │ │
│  │  │            │  │            │  │            │                          │ │
│  │  │- Daily     │  │- Manual    │  │- Promotion │                          │ │
│  │  │  rollover  │  │  adjust    │  │  credit    │                          │ │
│  │  │- Per pos   │  │- Admin auth│  │- Admin auth│                          │ │
│  │  └────────────┘  └────────────┘  └────────────┘                          │ │
│  └────────────────────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────────────┘
             │                         │                         │
             ▼                         ▼                         ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            PERSISTENCE LAYER                                      │
├──────────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐             │
│  │  Account Store   │  │  Position Store  │  │   Ledger Store   │             │
│  │  (PostgreSQL)    │  │  (PostgreSQL)    │  │  (PostgreSQL)    │             │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘             │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Diagrams

### Market Data Flow

```
LP Feeds (YOFX1, OANDA, Binance)
         │
         │ FIX 4.4 / REST / WebSocket
         ▼
┌──────────────────┐
│   LP Adapters    │  Normalize quotes from different protocols
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│   LP Manager     │  - Aggregate quotes
│                  │  - Apply spread markup
│                  │  - Failover logic
└────────┬─────────┘
         │
         ├──────────────────────┬──────────────────────┐
         │                      │                      │
         ▼                      ▼                      ▼
┌──────────────┐  ┌──────────────────┐  ┌──────────────────┐
│    Redis     │  │  B-Book Engine   │  │  WebSocket Hub   │
│  (Cache)     │  │  (Execution)     │  │  (Streaming)     │
│              │  │                  │  │                  │
│ TTL: 60s     │  │ - Price for exec │  │ - Broadcast to   │
│              │  │ - P&L calc       │  │   subscribers    │
└──────────────┘  └──────────────────┘  └────────┬─────────┘
                                                  │
                                                  │
                                                  ▼
                                        ┌──────────────────┐
                                        │     Clients      │
                                        │  (Web/Mobile)    │
                                        └──────────────────┘
```

### Order Execution Flow (B-Book)

```
┌────────────┐
│   Client   │  Places market order: BUY 1.0 EURUSD
└─────┬──────┘
      │
      │ POST /api/orders/market
      ▼
┌────────────────┐
│   Client API   │  Validates JWT, rate limit
└─────┬──────────┘
      │
      ▼
┌────────────────┐
│      OMS       │
│                │
│ 1. Validate:   │  - Symbol exists
│                │  - Volume within limits
│                │  - Account active
│                │
│ 2. Risk Check: │  - Calculate required margin
│                │  - Check free margin
│                │  - Verify position limits
│                │
│ 3. Route:      │  - Client classification
│                │  - Position size
│                │  - LP availability
│                │  → Route to B-Book
└─────┬──────────┘
      │
      ▼
┌────────────────┐
│  B-Book Engine │
│                │
│ 1. Get Quote:  │  Redis: quote:EURUSD
│                │  Bid: 1.08450, Ask: 1.08452
│                │
│ 2. Fill Order: │  Execute at Ask: 1.08452
│                │
│ 3. Create Pos: │  Position #12345
│                │  Side: BUY, Vol: 1.0
│                │  Open: 1.08452
│                │
│ 4. Deduct Fee: │  Commission: $7.00
│                │  Balance: $10,000 → $9,993
│                │
│ 5. Record:     │  PostgreSQL: positions, trades, ledger
└─────┬──────────┘
      │
      ├────────────────────────┬────────────────────────┐
      │                        │                        │
      ▼                        ▼                        ▼
┌────────────┐  ┌────────────────────┐  ┌────────────────────┐
│ PostgreSQL │  │   WebSocket Hub    │  │    RabbitMQ        │
│            │  │                    │  │                    │
│ - Position │  │ Broadcast:         │  │ Queue:             │
│ - Trade    │  │ {                  │  │ notifications.push │
│ - Ledger   │  │   type: "position",│  │                    │
│            │  │   action: "opened" │  │ Send push notif    │
└────────────┘  │ }                  │  │ to mobile app      │
                │                    │  └────────────────────┘
                │ → Client receives  │
                │   real-time update │
                └────────────────────┘
```

---

## Deployment View

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          PRODUCTION DEPLOYMENT                                   │
└─────────────────────────────────────────────────────────────────────────────────┘

Cloud Provider: AWS / GCP / Azure
Region: us-east-1 (Primary), us-west-2 (DR)

┌──────────────────────────────────────────────────────────────────────────────────┐
│                                VPC (10.0.0.0/16)                                  │
├──────────────────────────────────────────────────────────────────────────────────┤
│                                                                                   │
│  ┌──────────────────────────────────────────────────────────────────────────┐  │
│  │  Public Subnet (10.0.1.0/24)                                              │  │
│  │                                                                            │  │
│  │  ┌──────────────────┐  ┌──────────────────┐                              │  │
│  │  │ Load Balancer 1  │  │ Load Balancer 2  │                              │  │
│  │  │ (Client Traffic) │  │ (Admin Traffic)  │                              │  │
│  │  └────────┬─────────┘  └────────┬─────────┘                              │  │
│  └───────────┼──────────────────────┼────────────────────────────────────────┘  │
│              │                      │                                            │
│  ┌───────────┼──────────────────────┼────────────────────────────────────────┐  │
│  │  Private Subnet - App (10.0.10.0/24)                                      │  │
│  │           │                      │                                        │  │
│  │           ▼                      ▼                                        │  │
│  │  ┌────────────────────────────────────────────────────────────────────┐  │  │
│  │  │            Kubernetes Cluster (EKS/GKE/AKS)                         │  │  │
│  │  │                                                                      │  │  │
│  │  │  Namespace: trading                                                 │  │  │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐              │  │  │
│  │  │  │Client API│ │Client API│ │Client API│ │Client API│ (4 pods)     │  │  │
│  │  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘              │  │  │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐                           │  │  │
│  │  │  │   OMS    │ │   OMS    │ │   OMS    │           (3 pods)        │  │  │
│  │  │  └──────────┘ └──────────┘ └──────────┘                           │  │  │
│  │  │  ┌──────────┐ ┌──────────┐                                        │  │  │
│  │  │  │  Router  │ │  Router  │                       (2 pods)         │  │  │
│  │  │  └──────────┘ └──────────┘                                        │  │  │
│  │  │                                                                      │  │  │
│  │  │  Namespace: execution                                               │  │  │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐                           │  │  │
│  │  │  │  B-Book  │ │  B-Book  │ │  B-Book  │         (3 pods)          │  │  │
│  │  │  └──────────┘ └──────────┘ └──────────┘                           │  │  │
│  │  │  ┌──────────┐ ┌──────────┐                                        │  │  │
│  │  │  │   FIX    │ │   FIX    │                       (2 pods)         │  │  │
│  │  │  │  Gateway │ │  Gateway │                                        │  │  │
│  │  │  └──────────┘ └──────────┘                                        │  │  │
│  │  │                                                                      │  │  │
│  │  │  Namespace: market-data                                            │  │  │
│  │  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐│  │  │
│  │  │  │WebSocket │ │WebSocket │ │WebSocket │ │WebSocket │ │WebSocket ││  │  │
│  │  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘│  │  │
│  │  │  (5 pods)                                                          │  │  │
│  │  │  ┌──────────┐ ┌──────────┐                                        │  │  │
│  │  │  │LP Manager│ │LP Manager│                       (2 pods)         │  │  │
│  │  │  └──────────┘ └──────────┘                                        │  │  │
│  │  │                                                                      │  │  │
│  │  │  Namespace: admin                                                  │  │  │
│  │  │  ┌──────────┐ ┌──────────┐                                        │  │  │
│  │  │  │Admin API │ │Admin API │                       (2 pods)         │  │  │
│  │  │  └──────────┘ └──────────┘                                        │  │  │
│  │  └──────────────────────────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                               │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  Private Subnet - Data (10.0.20.0/24)                                 │  │
│  │                                                                         │  │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐          │  │
│  │  │  PostgreSQL    │  │  TimescaleDB   │  │ Redis Cluster  │          │  │
│  │  │  (Master)      │  │                │  │ (3 masters +   │          │  │
│  │  │                │  │                │  │  3 replicas)   │          │  │
│  │  │  + 2 Replicas  │  │  + Compression │  │                │          │  │
│  │  └────────────────┘  └────────────────┘  └────────────────┘          │  │
│  │                                                                         │  │
│  │  ┌────────────────────────────────────┐                               │  │
│  │  │       RabbitMQ Cluster             │                               │  │
│  │  │  ┌──────┐  ┌──────┐  ┌──────┐     │                               │  │
│  │  │  │Node 1│  │Node 2│  │Node 3│     │                               │  │
│  │  │  └──────┘  └──────┘  └──────┘     │                               │  │
│  │  └────────────────────────────────────┘                               │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────┘

Monitoring Stack (Separate VPC):
┌──────────────────────────────────────────────────────────────────────────────────┐
│  Prometheus + Grafana + Jaeger + ELK Stack                                       │
│  (Collects metrics, traces, logs from all services)                              │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

**Document Status:** Complete
**Next Steps:** Detailed component specifications, API contracts, database schemas
