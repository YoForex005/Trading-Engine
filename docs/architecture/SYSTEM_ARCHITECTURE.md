# Trading Engine System Architecture

## Executive Summary

This document outlines the comprehensive production-ready architecture for a multi-asset trading engine supporting Forex, CFDs, and Cryptocurrency trading with A/B/C-Book execution models.

**Version:** 1.0
**Date:** 2026-01-18
**Status:** Design Phase

---

## 1. Architecture Overview

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CLIENT APPLICATIONS                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │ Web Platform │  │ Mobile Apps  │  │ MT4/MT5 API  │                  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                  │
└─────────┼──────────────────┼──────────────────┼──────────────────────────┘
          │                  │                  │
          │    WebSocket     │      REST API    │    FIX/Binary
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          API GATEWAY (NGINX)                             │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  Rate Limiting │ Auth │ SSL/TLS │ Load Balancing │ Compression │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────┬───────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                        MICROSERVICES LAYER                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │   Client     │  │   Admin      │  │  WebSocket   │                 │
│  │     API      │  │     API      │  │    Server    │                 │
│  │   (REST)     │  │   (REST)     │  │   (Quotes)   │                 │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                 │
│         │                  │                  │                          │
│         └──────────────────┼──────────────────┘                          │
│                            │                                             │
│  ┌─────────────────────────┴──────────────────────────────┐            │
│  │             ORDER MANAGEMENT SYSTEM (OMS)               │            │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐       │            │
│  │  │   Order    │  │  Position  │  │   Risk     │       │            │
│  │  │ Validator  │  │  Manager   │  │   Engine   │       │            │
│  │  └────────────┘  └────────────┘  └────────────┘       │            │
│  └─────────────────────────┬──────────────────────────────┘            │
│                            │                                             │
│  ┌─────────────────────────┴──────────────────────────────┐            │
│  │           SMART ORDER ROUTER                            │            │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐       │            │
│  │  │  A-Book    │  │  B-Book    │  │  C-Book    │       │            │
│  │  │  (STP/LP)  │  │ (Internal) │  │ (Hybrid)   │       │            │
│  │  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘       │            │
│  └────────┼───────────────┼───────────────┼──────────────┘            │
│           │               │               │                             │
│  ┌────────┴──────┐  ┌────┴────────┐  ┌───┴──────────┐                │
│  │ FIX Gateway   │  │  B-Book     │  │  Aggregation │                │
│  │  (LP Conn)    │  │  Engine     │  │    Logic     │                │
│  └───────┬───────┘  └─────────────┘  └──────────────┘                │
│          │                                                              │
└──────────┼──────────────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    LIQUIDITY PROVIDERS (LPs)                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │  YOFX1   │  │  OANDA   │  │ Binance  │  │  Prime   │               │
│  │  (FIX)   │  │  (REST)  │  │   (WS)   │  │   XM     │               │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘               │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                          DATA LAYER                                      │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                 │
│  │ PostgreSQL   │  │ TimescaleDB  │  │    Redis     │                 │
│  │  (Accounts,  │  │   (Ticks,    │  │  (Quotes,    │                 │
│  │   Orders,    │  │    OHLC)     │  │   Sessions)  │                 │
│  │   Trades)    │  │              │  │              │                 │
│  └──────────────┘  └──────────────┘  └──────────────┘                 │
│                                                                          │
│  ┌──────────────────────────────────────────────────────┐              │
│  │            RabbitMQ (Message Queue)                   │              │
│  │  - Async Order Processing                            │              │
│  │  - Event Notifications                               │              │
│  │  - Risk Check Pipeline                               │              │
│  └──────────────────────────────────────────────────────┘              │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Architecture Decision: Microservices vs Monolith

**Decision:** Adopt **Microservices Architecture** with bounded contexts

**Rationale:**

| Factor | Microservices Benefit | Trade-off |
|--------|----------------------|-----------|
| **Scalability** | Independent scaling of market data vs execution | Operational complexity |
| **Fault Isolation** | FIX gateway crash doesn't affect B-Book execution | Network latency between services |
| **Technology Flexibility** | Go for performance-critical (execution), Python for analytics | Polyglot stack maintenance |
| **Team Independence** | Separate teams for trading, admin, market data | Inter-team coordination overhead |
| **Deployment** | Deploy quote engine without restarting execution | Distributed tracing complexity |

**Selected Pattern:** Domain-Driven Design (DDD) with bounded contexts:
- **Trading Context:** Orders, Positions, Executions
- **Market Data Context:** Quotes, OHLC, Tick Storage
- **Admin Context:** Accounts, Users, Configuration
- **Risk Context:** Margin, Exposure, Limits

---

## 2. Component Architecture

### 2.1 FIX Gateway Service

**Responsibility:** Manage FIX 4.4 connections to liquidity providers

**Technology:** Go (quickfix-go)

**Key Features:**
- Multiple concurrent FIX sessions (YOFX1, YOFX2, Prime XM)
- Automatic reconnection with exponential backoff
- Session state persistence (sequence numbers)
- Message routing to Order Router
- Market data subscription (35=V, 35=W)

**Configuration:**
```json
{
  "sessions": [
    {
      "id": "YOFX1",
      "type": "trading",
      "senderCompID": "YOFX1",
      "targetCompID": "YOFX",
      "host": "23.106.238.138",
      "port": 12336,
      "heartbeatInterval": 30,
      "reconnectInterval": 5,
      "maxReconnects": 10
    }
  ]
}
```

**Endpoints:**
- `POST /fix/connect` - Initiate FIX session
- `POST /fix/disconnect` - Graceful disconnect
- `GET /fix/status` - Session health check
- `GET /fix/sessions` - List all sessions

**State Machine:**
```
[Disconnected] --connect--> [Connecting] --logon--> [Connected]
                                |                        |
                                +------- timeout --------+
                                |                        |
                                v                        v
                            [Failed] <--logout----- [Disconnected]
                                |
                          retry backoff
                                |
                                v
                          [Connecting]
```

### 2.2 Order Router Service

**Responsibility:** Route orders to appropriate execution venue (A/B/C-Book)

**Technology:** Go

**Routing Logic:**
```go
type RoutingDecision struct {
    Venue      string // "A-BOOK", "B-BOOK", "C-BOOK"
    Reason     string
    Confidence float64
}

func (r *Router) Route(order Order, client Client) RoutingDecision {
    // 1. Check client classification
    if client.Classification == "VIP" || client.ProfitableTrader {
        return A_BOOK // Route to LP
    }

    // 2. Position size check
    if order.Volume > r.config.BBookMaxVolume {
        return A_BOOK // Too large for internal book
    }

    // 3. Symbol volatility
    if r.getVolatility(order.Symbol) > r.config.HighVolThreshold {
        return A_BOOK // High risk, route to LP
    }

    // 4. LP availability check
    if !r.lpManager.IsHealthy(order.Symbol) {
        return B_BOOK // Fallback to internal
    }

    // 5. Risk-based hybrid (C-Book)
    exposure := r.riskEngine.GetExposure(order.Symbol)
    if exposure > r.config.MaxExposure {
        return C_BOOK // Partial hedge
    }

    return B_BOOK // Default to internal
}
```

**Configuration:**
```json
{
  "routingRules": {
    "bBookMaxVolume": 10.0,
    "highVolThreshold": 0.05,
    "maxExposure": 1000000,
    "vipClients": ["CLIENT001", "CLIENT002"]
  },
  "execution": {
    "slippageControl": true,
    "maxSlippagePips": 2,
    "requotes": false
  }
}
```

### 2.3 B-Book Execution Engine

**Responsibility:** Internal order matching and position management

**Technology:** Go

**Architecture:**
```
┌─────────────────────────────────────────────────────────┐
│              B-Book Execution Engine                     │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Order      │  │   Position   │  │   Account    │ │
│  │   Book       │  │   Manager    │  │   Manager    │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │
│         │                  │                  │         │
│         └──────────────────┼──────────────────┘         │
│                            │                            │
│  ┌─────────────────────────┴──────────────────────┐   │
│  │          Margin Calculator                      │   │
│  │  - Required Margin                              │   │
│  │  - Free Margin                                  │   │
│  │  - Margin Level                                 │   │
│  │  - Stop Out Detection                           │   │
│  └─────────────────────────────────────────────────┘   │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │          P&L Calculator                           │  │
│  │  - Unrealized P&L (Mark-to-Market)               │  │
│  │  - Realized P&L (Closed Positions)               │  │
│  │  - Swap/Rollover Calculation                     │  │
│  │  - Commission Application                        │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │          Risk Management                          │  │
│  │  - Margin Call (100% level)                      │  │
│  │  - Stop Out (50% level)                          │  │
│  │  - Max Position Size                             │  │
│  │  - Max Total Exposure                            │  │
│  └──────────────────────────────────────────────────┘  │
│                                                          │
│  ┌──────────────────────────────────────────────────┐  │
│  │          Ledger System                            │  │
│  │  - Deposit                                        │  │
│  │  - Withdrawal                                     │  │
│  │  - Commission                                     │  │
│  │  - Swap                                           │  │
│  │  - Realized P&L                                   │  │
│  │  - Bonus/Adjustment                               │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**Data Flow:**
```
Order Received
    ↓
Validate Order (Symbol, Volume, Price)
    ↓
Calculate Required Margin
    ↓
Check Free Margin >= Required Margin
    ↓
Get Market Price (Bid/Ask from Quote Feed)
    ↓
Execute Order (Fill at Market Price)
    ↓
Create Position Record
    ↓
Deduct Commission from Balance
    ↓
Record Trade in Ledger
    ↓
Broadcast Position Update via WebSocket
    ↓
Store in Database (PostgreSQL)
```

**Position Lifecycle:**
```
[NEW ORDER] → [VALIDATION] → [MARGIN CHECK] → [EXECUTION]
                                                    ↓
                                              [OPEN POSITION]
                                                    ↓
                                          [PRICE UPDATES]
                                          (Unrealized P&L)
                                                    ↓
                                    [CLOSE TRIGGER] (Manual/SL/TP/Stop Out)
                                                    ↓
                                              [CLOSE POSITION]
                                                    ↓
                                        [CALCULATE REALIZED P&L]
                                                    ↓
                                          [UPDATE BALANCE]
                                                    ↓
                                           [RECORD TRADE]
```

### 2.4 Admin API Service

**Responsibility:** Administrative functions for broker operators

**Technology:** Go (Gin framework)

**Feature Modules:**

#### 2.4.1 Account Management
```go
// POST /admin/accounts
type CreateAccountRequest struct {
    Username      string  `json:"username" binding:"required"`
    Password      string  `json:"password" binding:"required"`
    Email         string  `json:"email" binding:"required,email"`
    AccountType   string  `json:"accountType"` // DEMO, LIVE
    InitialDeposit float64 `json:"initialDeposit"`
    Leverage      int     `json:"leverage" binding:"min=1,max=500"`
    MarginMode    string  `json:"marginMode"` // HEDGING, NETTING
    Currency      string  `json:"currency" binding:"required"`
}

// PUT /admin/accounts/{id}/suspend
// PUT /admin/accounts/{id}/activate
// PUT /admin/accounts/{id}/leverage
// DELETE /admin/accounts/{id}
```

#### 2.4.2 Financial Operations
```go
// POST /admin/deposit
type DepositRequest struct {
    AccountID int64   `json:"accountId"`
    Amount    float64 `json:"amount" binding:"min=0.01"`
    Method    string  `json:"method"` // BANK_TRANSFER, CRYPTO, CARD
    Reference string  `json:"reference"`
    Note      string  `json:"note"`
}

// POST /admin/withdraw
type WithdrawRequest struct {
    AccountID int64   `json:"accountId"`
    Amount    float64 `json:"amount" binding:"min=0.01"`
    Method    string  `json:"method"`
    BankInfo  string  `json:"bankInfo,omitempty"`
    CryptoAddr string `json:"cryptoAddr,omitempty"`
}

// POST /admin/adjust
type AdjustmentRequest struct {
    AccountID int64   `json:"accountId"`
    Amount    float64 `json:"amount"` // Can be negative
    Reason    string  `json:"reason" binding:"required"`
    Type      string  `json:"type"` // MANUAL, BONUS, CORRECTION
}
```

#### 2.4.3 Execution Model Control
```go
// POST /admin/execution-mode
type ExecutionModeRequest struct {
    Mode        string            `json:"mode"` // A-BOOK, B-BOOK, C-BOOK
    ClientID    *int64            `json:"clientId,omitempty"` // null = global
    SymbolRules map[string]string `json:"symbolRules,omitempty"` // Per-symbol
}

// GET /admin/execution-mode
type ExecutionModeResponse struct {
    GlobalMode    string                    `json:"globalMode"`
    ClientOverrides map[int64]string       `json:"clientOverrides"`
    SymbolRules   map[string]ExecutionRule `json:"symbolRules"`
}

type ExecutionRule struct {
    Mode            string  `json:"mode"`
    MaxVolume       float64 `json:"maxVolume"`
    PriorityLP      string  `json:"priorityLP"`
    FallbackMode    string  `json:"fallbackMode"`
}
```

#### 2.4.4 Symbol Management
```go
// GET /admin/symbols
// POST /admin/symbols/{symbol}/enable
// POST /admin/symbols/{symbol}/disable
// PUT /admin/symbols/{symbol}/settings
type SymbolSettings struct {
    Symbol          string  `json:"symbol"`
    Enabled         bool    `json:"enabled"`
    SpreadMarkup    float64 `json:"spreadMarkup"` // pips
    CommissionPerLot float64 `json:"commissionPerLot"`
    MinVolume       float64 `json:"minVolume"`
    MaxVolume       float64 `json:"maxVolume"`
    MarginPercent   float64 `json:"marginPercent"`
    SwapLong        float64 `json:"swapLong"`
    SwapShort       float64 `json:"swapShort"`
}
```

#### 2.4.5 Risk Parameters
```go
// PUT /admin/risk-settings
type RiskSettings struct {
    MarginCallLevel   float64 `json:"marginCallLevel"` // e.g., 100%
    StopOutLevel      float64 `json:"stopOutLevel"`    // e.g., 50%
    MaxLeverage       int     `json:"maxLeverage"`
    MaxPositionSize   float64 `json:"maxPositionSize"`
    MaxDailyVolume    float64 `json:"maxDailyVolume"`
    MaxOpenPositions  int     `json:"maxOpenPositions"`
}
```

#### 2.4.6 Liquidity Provider Management
```go
// GET /admin/lps
type LPStatus struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Type        string    `json:"type"` // FIX, REST, WEBSOCKET
    Connected   bool      `json:"connected"`
    Enabled     bool      `json:"enabled"`
    Latency     int       `json:"latency"` // ms
    LastQuote   time.Time `json:"lastQuote"`
    Symbols     []string  `json:"symbols"`
}

// POST /admin/lps/{id}/enable
// POST /admin/lps/{id}/disable
// POST /admin/lps/{id}/reconnect
```

### 2.5 Client Trading API

**Responsibility:** Client-facing trading operations

**Technology:** Go (Gin framework)

**Endpoints:**

#### Market Data
- `GET /api/symbols` - Available trading symbols
- `GET /api/ticks?symbol={symbol}&limit=500` - Tick history
- `GET /api/ohlc?symbol={symbol}&timeframe=1h&limit=500` - Candlestick data
- `WS /ws` - Real-time quote stream

#### Account Information
- `POST /api/login` - Authenticate user
- `GET /api/account/summary` - Balance, equity, margin
- `GET /api/account/positions` - Open positions
- `GET /api/account/orders` - Pending orders
- `GET /api/account/trades` - Trade history
- `GET /api/account/ledger` - Transaction history

#### Order Management
- `POST /api/orders/market` - Market order
- `POST /api/orders/limit` - Limit order
- `POST /api/orders/stop` - Stop order
- `POST /api/orders/stop-limit` - Stop-limit order
- `PUT /api/orders/{id}/modify` - Modify pending order
- `DELETE /api/orders/{id}` - Cancel pending order

#### Position Management
- `POST /api/positions/{id}/close` - Close position
- `POST /api/positions/{id}/close-partial` - Partial close
- `PUT /api/positions/{id}/modify` - Modify SL/TP
- `POST /api/positions/close-all` - Close all positions

#### Risk Tools
- `GET /api/risk/calculate-lot?risk=2&slPips=20&symbol=EURUSD` - Lot size calculator
- `GET /api/risk/margin-preview?symbol=EURUSD&volume=1.0&side=BUY` - Margin preview

### 2.6 WebSocket Server

**Responsibility:** Real-time market data and account updates

**Technology:** Go (gorilla/websocket)

**Channels:**

#### Quote Stream
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.08450,
  "ask": 1.08452,
  "spread": 0.00002,
  "timestamp": 1705598400000,
  "lp": "OANDA"
}
```

#### Account Updates
```json
{
  "type": "account_update",
  "balance": 10000.00,
  "equity": 10500.00,
  "margin": 500.00,
  "freeMargin": 10000.00,
  "marginLevel": 2100.00,
  "unrealizedPnL": 500.00
}
```

#### Position Updates
```json
{
  "type": "position_update",
  "action": "opened",
  "position": {
    "id": 12345,
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 1.0,
    "openPrice": 1.08450,
    "currentPrice": 1.08500,
    "unrealizedPnL": 50.00,
    "sl": 1.08300,
    "tp": 1.08700
  }
}
```

#### Order Execution
```json
{
  "type": "order_filled",
  "orderId": "ORD-12345",
  "positionId": 12345,
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 1.0,
  "filledPrice": 1.08450,
  "timestamp": 1705598400000
}
```

**Subscription Model:**
```go
type Subscription struct {
    Type    string   // "quotes", "account", "positions"
    Symbols []string // For quote subscriptions
}

// Client sends:
{
  "action": "subscribe",
  "type": "quotes",
  "symbols": ["EURUSD", "GBPUSD", "BTCUSD"]
}

// Client sends:
{
  "action": "subscribe",
  "type": "account"
}
```

---

## 3. Database Design

### 3.1 PostgreSQL (Transactional Data)

#### Schema: `accounts`
```sql
CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    account_type VARCHAR(20) NOT NULL, -- DEMO, LIVE
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    equity DECIMAL(20, 8) NOT NULL DEFAULT 0,
    margin DECIMAL(20, 8) NOT NULL DEFAULT 0,
    free_margin DECIMAL(20, 8) NOT NULL DEFAULT 0,
    margin_level DECIMAL(10, 2) NOT NULL DEFAULT 0,
    leverage INT NOT NULL DEFAULT 100,
    margin_mode VARCHAR(20) NOT NULL DEFAULT 'HEDGING', -- HEDGING, NETTING
    currency VARCHAR(10) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, SUSPENDED, CLOSED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    INDEX idx_user_id (user_id),
    INDEX idx_account_number (account_number),
    INDEX idx_status (status)
);
```

#### Schema: `positions`
```sql
CREATE TABLE positions (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    volume DECIMAL(20, 8) NOT NULL,
    open_price DECIMAL(20, 8) NOT NULL,
    current_price DECIMAL(20, 8) NOT NULL,
    open_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    close_time TIMESTAMPTZ,
    close_price DECIMAL(20, 8),
    sl DECIMAL(20, 8),
    tp DECIMAL(20, 8),
    swap DECIMAL(20, 8) NOT NULL DEFAULT 0,
    commission DECIMAL(20, 8) NOT NULL DEFAULT 0,
    unrealized_pnl DECIMAL(20, 8) NOT NULL DEFAULT 0,
    realized_pnl DECIMAL(20, 8),
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN', -- OPEN, CLOSED
    close_reason VARCHAR(50), -- MANUAL, SL_HIT, TP_HIT, STOP_OUT, MARGIN_CALL
    execution_venue VARCHAR(20) NOT NULL, -- BBOOK, ABOOK, CBOOK
    lp_order_id VARCHAR(100), -- LP reference if A-Book
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    INDEX idx_account_positions (account_id, status),
    INDEX idx_symbol_positions (symbol, status),
    INDEX idx_open_time (open_time),
    INDEX idx_close_time (close_time)
);
```

#### Schema: `orders`
```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    symbol VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL, -- MARKET, LIMIT, STOP, STOP_LIMIT
    side VARCHAR(10) NOT NULL, -- BUY, SELL
    volume DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8), -- For limit/stop orders
    trigger_price DECIMAL(20, 8), -- For stop orders
    sl DECIMAL(20, 8),
    tp DECIMAL(20, 8),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, FILLED, CANCELLED, REJECTED
    filled_price DECIMAL(20, 8),
    filled_at TIMESTAMPTZ,
    position_id BIGINT REFERENCES positions(id),
    reject_reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    INDEX idx_account_orders (account_id, status),
    INDEX idx_symbol_orders (symbol, status),
    INDEX idx_created_at (created_at)
);
```

#### Schema: `trades`
```sql
CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    position_id BIGINT REFERENCES positions(id),
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    volume DECIMAL(20, 8) NOT NULL,
    price DECIMAL(20, 8) NOT NULL,
    realized_pnl DECIMAL(20, 8),
    commission DECIMAL(20, 8) NOT NULL DEFAULT 0,
    swap DECIMAL(20, 8) NOT NULL DEFAULT 0,
    execution_venue VARCHAR(20) NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    INDEX idx_account_trades (account_id, executed_at),
    INDEX idx_position_trades (position_id),
    INDEX idx_executed_at (executed_at)
);
```

#### Schema: `ledger`
```sql
CREATE TABLE ledger (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES accounts(id),
    type VARCHAR(50) NOT NULL, -- DEPOSIT, WITHDRAWAL, COMMISSION, SWAP, REALIZED_PNL, BONUS, ADJUSTMENT
    amount DECIMAL(20, 8) NOT NULL, -- Can be negative
    balance_before DECIMAL(20, 8) NOT NULL,
    balance_after DECIMAL(20, 8) NOT NULL,
    reference_id BIGINT, -- trade_id, order_id, etc.
    reference_type VARCHAR(50), -- TRADE, ORDER, DEPOSIT, etc.
    method VARCHAR(50), -- BANK_TRANSFER, CRYPTO, CARD, etc.
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    INDEX idx_account_ledger (account_id, created_at),
    INDEX idx_type (type),
    INDEX idx_reference (reference_type, reference_id)
);
```

### 3.2 TimescaleDB (Time-Series Data)

#### Hypertable: `ticks`
```sql
CREATE TABLE ticks (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    bid DECIMAL(20, 8) NOT NULL,
    ask DECIMAL(20, 8) NOT NULL,
    spread DECIMAL(20, 8) NOT NULL,
    lp VARCHAR(50) NOT NULL, -- Liquidity provider source
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('ticks', 'time', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX idx_ticks_symbol_time ON ticks (symbol, time DESC);
CREATE INDEX idx_ticks_lp ON ticks (lp, time DESC);
```

#### Hypertable: `ohlc`
```sql
CREATE TABLE ohlc (
    time TIMESTAMPTZ NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    timeframe VARCHAR(10) NOT NULL, -- 1m, 5m, 15m, 1h, 4h, 1d
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    volume BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (time, symbol, timeframe)
);

SELECT create_hypertable('ohlc', 'time', chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_ohlc_symbol_tf_time ON ohlc (symbol, timeframe, time DESC);
```

**Continuous Aggregates:**
```sql
-- Automatically aggregate ticks to 1-minute OHLC
CREATE MATERIALIZED VIEW ohlc_1m
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS time,
    symbol,
    '1m' as timeframe,
    FIRST(bid, time) AS open,
    MAX(bid) AS high,
    MIN(bid) AS low,
    LAST(bid, time) AS close,
    COUNT(*) AS volume
FROM ticks
GROUP BY time_bucket('1 minute', time), symbol;

SELECT add_continuous_aggregate_policy('ohlc_1m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');
```

### 3.3 Redis (Caching & Real-Time State)

**Key Structure:**

```
# Latest quotes (TTL: 60s)
quote:{symbol} -> {bid: 1.08450, ask: 1.08452, timestamp: 1705598400}

# Account session
session:{token} -> {accountId: 12345, userId: "user123", expires: 1705598400}

# Real-time positions (updated on every price tick)
positions:{accountId} -> [{id: 123, symbol: "EURUSD", unrealizedPnL: 50.00}, ...]

# Order book (for B-Book)
orderbook:{symbol}:bids -> [{price: 1.08450, volume: 10.5}, ...]
orderbook:{symbol}:asks -> [{price: 1.08452, volume: 8.2}, ...]

# Rate limiting
ratelimit:{clientId}:{endpoint} -> {count: 45, window: 1705598400}

# WebSocket subscriptions
ws:subscriptions:{symbol} -> [connectionId1, connectionId2, ...]
```

**Data Structures:**
- **Strings:** Quotes, sessions
- **Hashes:** Account state, position state
- **Lists:** Order book levels
- **Sorted Sets:** Real-time leaderboards, P&L rankings
- **Pub/Sub:** Real-time event broadcasting

---

## 4. Message Queue Architecture (RabbitMQ)

### 4.1 Exchange and Queue Design

```
                        RabbitMQ
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
    [Orders]          [Executions]        [Notifications]
    Exchange          Exchange            Exchange
    (topic)           (fanout)            (topic)
        │                   │                   │
   ┌────┴────┐         ┌────┴────┐         ┌───┴───┐
   │         │         │         │         │       │
[Validate] [Route]  [BBook]  [ABook]  [Email] [Push]
  Queue    Queue     Queue    Queue     Queue  Queue
```

### 4.2 Queue Definitions

#### Order Processing Pipeline
```go
// orders.validate queue
type OrderValidationMessage struct {
    OrderID   int64   `json:"orderId"`
    AccountID int64   `json:"accountId"`
    Symbol    string  `json:"symbol"`
    Volume    float64 `json:"volume"`
    Price     float64 `json:"price,omitempty"`
}

// orders.route queue
type OrderRoutingMessage struct {
    OrderID int64  `json:"orderId"`
    Venue   string `json:"venue"` // A-BOOK, B-BOOK, C-BOOK
    Reason  string `json:"reason"`
}

// executions.bbook queue
type BBookExecutionMessage struct {
    OrderID   int64   `json:"orderId"`
    AccountID int64   `json:"accountId"`
    Symbol    string  `json:"symbol"`
    Side      string  `json:"side"`
    Volume    float64 `json:"volume"`
    Price     float64 `json:"price,omitempty"`
}

// executions.abook queue
type ABookExecutionMessage struct {
    OrderID   int64  `json:"orderId"`
    AccountID int64  `json:"accountId"`
    LPID      string `json:"lpId"`
    FIXMessage string `json:"fixMessage"`
}
```

#### Notification Pipeline
```go
// notifications.email queue
type EmailNotification struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
    Type    string `json:"type"` // DEPOSIT_CONFIRMED, WITHDRAWAL_APPROVED, etc.
}

// notifications.push queue
type PushNotification struct {
    AccountID int64  `json:"accountId"`
    Title     string `json:"title"`
    Message   string `json:"message"`
    Type      string `json:"type"`
}
```

### 4.3 Dead Letter Queue (DLQ)

```go
// All failed messages route to DLQ for manual review
type DLQMessage struct {
    OriginalQueue string      `json:"originalQueue"`
    Message       interface{} `json:"message"`
    Error         string      `json:"error"`
    RetryCount    int         `json:"retryCount"`
    Timestamp     int64       `json:"timestamp"`
}
```

**Retry Policy:**
- Retry 3 times with exponential backoff (5s, 15s, 45s)
- After 3 failures, route to DLQ
- Alert admin via monitoring system

---

## 5. Real-Time Quote System

### 5.1 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Liquidity Providers                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                  │
│  │  YOFX2   │  │  OANDA   │  │ Binance  │                  │
│  │  (FIX)   │  │  (REST)  │  │   (WS)   │                  │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘                  │
└────────┼─────────────┼─────────────┼────────────────────────┘
         │             │             │
         └──────┬──────┴──────┬──────┘
                │             │
         ┌──────▼─────────────▼──────┐
         │   LP Manager Service       │
         │  - Adapter Registry        │
         │  - Quote Aggregation       │
         │  - Failover Logic          │
         └──────────┬─────────────────┘
                    │
         ┌──────────▼─────────────────┐
         │  Quote Normalization       │
         │  - Symbol Mapping          │
         │  - Spread Application      │
         │  - Price Validation        │
         └──────────┬─────────────────┘
                    │
         ┌──────────▼─────────────────┐
         │      Redis Cache            │
         │  TTL: 60s per symbol       │
         │  Key: quote:{symbol}       │
         └──────────┬─────────────────┘
                    │
         ┌──────────▼─────────────────┐
         │   WebSocket Hub             │
         │  - Client Subscriptions    │
         │  - Rate Limiting           │
         │  - Broadcast Engine        │
         └──────────┬─────────────────┘
                    │
         ┌──────────▼─────────────────┐
         │   Connected Clients         │
         │  (Web, Mobile, MT4)        │
         └────────────────────────────┘
```

### 5.2 Quote Aggregation Logic

```go
type QuoteAggregator struct {
    lps          map[string]LPAdapter
    quoteChan    chan Quote
    subscribers  map[string][]chan Quote
    latestQuotes sync.Map // symbol -> Quote
}

func (qa *QuoteAggregator) AggregateQuotes() {
    for quote := range qa.quoteChan {
        // 1. Validate quote
        if !qa.validateQuote(quote) {
            continue
        }

        // 2. Apply spread markup
        quote = qa.applySpreadMarkup(quote)

        // 3. Store in Redis (with TTL)
        qa.storeQuote(quote)

        // 4. Update latest quote in memory
        qa.latestQuotes.Store(quote.Symbol, quote)

        // 5. Broadcast to WebSocket subscribers
        qa.broadcastQuote(quote)

        // 6. Store tick in TimescaleDB (async)
        qa.storeTickAsync(quote)
    }
}

func (qa *QuoteAggregator) applySpreadMarkup(quote Quote) Quote {
    settings := qa.getSymbolSettings(quote.Symbol)
    if settings.SpreadMarkup > 0 {
        halfMarkup := settings.SpreadMarkup / 2
        quote.Bid -= halfMarkup
        quote.Ask += halfMarkup
    }
    return quote
}
```

### 5.3 Failover & Redundancy

**Primary-Secondary LP Model:**
```go
type LPFailover struct {
    primary   LPAdapter
    secondary LPAdapter
    threshold time.Duration // 5 seconds
}

func (f *LPFailover) GetQuote(symbol string) (Quote, error) {
    // Try primary
    quote, err := f.primary.GetQuote(symbol)
    if err == nil && time.Since(quote.Timestamp) < f.threshold {
        return quote, nil
    }

    // Fallback to secondary
    log.Printf("[Failover] Primary LP stale/failed for %s, using secondary", symbol)
    return f.secondary.GetQuote(symbol)
}
```

**Health Monitoring:**
```go
type LPHealthMonitor struct {
    checkInterval time.Duration
    alertThreshold int // Failed checks before alert
}

func (m *LPHealthMonitor) Monitor(lp LPAdapter) {
    ticker := time.NewTicker(m.checkInterval)
    failedChecks := 0

    for range ticker.C {
        if !lp.IsHealthy() {
            failedChecks++
            if failedChecks >= m.alertThreshold {
                m.sendAlert("LP %s unhealthy: %d failed checks", lp.ID(), failedChecks)
            }
        } else {
            failedChecks = 0
        }
    }
}
```

### 5.4 Rate Limiting

**Per-Client WebSocket Rate Limiting:**
```go
type RateLimiter struct {
    maxUpdatesPerSecond int
    clients             map[string]*TokenBucket
}

type TokenBucket struct {
    tokens     int
    capacity   int
    refillRate int // tokens per second
    lastRefill time.Time
    mu         sync.Mutex
}

func (rl *RateLimiter) AllowUpdate(clientID string) bool {
    bucket := rl.getOrCreateBucket(clientID)
    return bucket.TakeToken()
}

func (tb *TokenBucket) TakeToken() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    // Refill tokens
    now := time.Now()
    elapsed := now.Sub(tb.lastRefill).Seconds()
    refill := int(elapsed * float64(tb.refillRate))

    if refill > 0 {
        tb.tokens = min(tb.capacity, tb.tokens+refill)
        tb.lastRefill = now
    }

    // Take token if available
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    return false
}
```

---

## 6. Order Execution Pipeline

### 6.1 Execution Flow

```
Client Order
    ↓
API Gateway
    ↓
Authentication & Authorization
    ↓
Rate Limiting Check
    ↓
Order Validation
  - Symbol exists
  - Volume within limits
  - Valid price (for limit/stop)
  - Account active
    ↓
Pre-Trade Risk Check
  - Sufficient margin
  - Position limits
  - Daily volume limits
    ↓
Smart Order Router
  - Client classification
  - Position size
  - Symbol risk
  - LP availability
    ↓
     ┌───────────────┬───────────────┬───────────────┐
     │               │               │               │
  A-Book         B-Book         C-Book
(LP Routing)   (Internal)      (Hybrid)
     │               │               │
     ▼               ▼               ▼
FIX Gateway    B-Book Engine   Split Logic
     │               │               │
     ▼               ▼               ▼
LP Execution   Internal Fill   Partial to LP
     │               │               │
     ▼               ▼               ▼
Fill Confirmation
     │               │               │
     └───────────────┴───────────────┘
                     │
                     ▼
           Position Created
                     │
                     ▼
           Trade Recorded in DB
                     │
                     ▼
         WebSocket Notification
                     │
                     ▼
           Client Updated
```

### 6.2 Order State Machine

```
[PENDING]
    │
    ├─ Validation Failed ─> [REJECTED]
    │
    ├─ Risk Check Failed ─> [REJECTED]
    │
    ├─ Validation Passed
    │       │
    │       ▼
    │   [ACCEPTED]
    │       │
    │       ├─ Market Order ─> [FILLING]
    │       │                      │
    │       │                      ├─ LP Filled ─> [FILLED]
    │       │                      │
    │       │                      ├─ LP Rejected ─> [REJECTED]
    │       │                      │
    │       │                      └─ Timeout ─> [EXPIRED]
    │       │
    │       └─ Limit/Stop Order ─> [WORKING]
    │                                  │
    │                                  ├─ Triggered ─> [FILLING] ─> [FILLED]
    │                                  │
    │                                  ├─ Cancelled ─> [CANCELLED]
    │                                  │
    │                                  └─ Expired ─> [EXPIRED]
    │
    └─ [CANCELLED] (Manual cancellation)
```

### 6.3 Position Tracking

```go
type PositionManager struct {
    positions     map[int64]*Position // positionID -> Position
    accountIndex  map[int64][]int64   // accountID -> []positionID
    symbolIndex   map[string][]int64  // symbol -> []positionID
    mu            sync.RWMutex
    priceCallback func(string) (bid, ask float64, ok bool)
}

func (pm *PositionManager) UpdatePrices() {
    pm.mu.Lock()
    defer pm.mu.Unlock()

    for _, pos := range pm.positions {
        if pos.Status != "OPEN" {
            continue
        }

        bid, ask, ok := pm.priceCallback(pos.Symbol)
        if !ok {
            continue
        }

        // Update current price
        if pos.Side == "BUY" {
            pos.CurrentPrice = bid
        } else {
            pos.CurrentPrice = ask
        }

        // Calculate unrealized P&L
        pos.UnrealizedPnL = pm.calculatePnL(pos)

        // Check SL/TP triggers
        pm.checkStopLoss(pos)
        pm.checkTakeProfit(pos)
    }
}

func (pm *PositionManager) calculatePnL(pos *Position) float64 {
    spec := pm.getSymbolSpec(pos.Symbol)

    var priceDiff float64
    if pos.Side == "BUY" {
        priceDiff = pos.CurrentPrice - pos.OpenPrice
    } else {
        priceDiff = pos.OpenPrice - pos.CurrentPrice
    }

    pips := priceDiff / spec.PipSize
    return pips * spec.PipValue * pos.Volume
}
```

### 6.4 P&L Calculation

**Unrealized P&L (Mark-to-Market):**
```
For BUY positions:
  P&L = (Current Bid - Open Price) / Pip Size × Pip Value × Volume

For SELL positions:
  P&L = (Open Price - Current Ask) / Pip Size × Pip Value × Volume
```

**Realized P&L (Closed Positions):**
```
For BUY positions:
  P&L = (Close Bid - Open Ask) / Pip Size × Pip Value × Volume - Commission - Swap

For SELL positions:
  P&L = (Open Bid - Close Ask) / Pip Size × Pip Value × Volume - Commission - Swap
```

**Example (EURUSD):**
```
Symbol: EURUSD
Contract Size: 100,000
Pip Size: 0.0001
Pip Value: $10 (for 1.0 lot)

BUY 1.0 lot at 1.08450
Current Bid: 1.08500

Unrealized P&L:
  = (1.08500 - 1.08450) / 0.0001 × 10 × 1.0
  = 0.00050 / 0.0001 × 10
  = 5 pips × $10
  = $50.00
```

---

## 7. Client Trading Interface Design

### 7.1 Dashboard Components

```
┌─────────────────────────────────────────────────────────────┐
│  RTX Trading Platform                    [Settings] [Logout] │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Account Summary                                      │  │
│  │  Balance: $10,000.00    Equity: $10,500.00           │  │
│  │  Margin: $500.00        Free Margin: $10,000.00      │  │
│  │  Margin Level: 2100%    Unrealized P&L: +$500.00    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Quick Trade Panel                                    │  │
│  │  Symbol: [EURUSD ▼]      Volume: [1.0] lots          │  │
│  │  Bid: 1.08450  Ask: 1.08452   Spread: 0.2 pips      │  │
│  │  SL: [1.08300]           TP: [1.08700]              │  │
│  │  [ SELL ]                            [ BUY ]         │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Chart (TradingView)                                  │  │
│  │  ┌────────────────────────────────────────────────┐  │  │
│  │  │                   EURUSD H1                      │  │  │
│  │  │  [Candlestick chart with indicators]            │  │  │
│  │  │  MA(20), MA(50), RSI(14), MACD                  │  │  │
│  │  └────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Open Positions                                       │  │
│  │  ┌────┬────────┬─────┬───────┬───────┬─────┬────┐  │  │
│  │  │ ID │ Symbol │Side │ Volume│ Price │ P&L │Act │  │  │
│  │  ├────┼────────┼─────┼───────┼───────┼─────┼────┤  │  │
│  │  │123 │EURUSD  │ BUY │  1.0  │1.08450│+50.0│[×] │  │  │
│  │  │124 │GBPUSD  │SELL │  0.5  │1.26300│-12.5│[×] │  │  │
│  │  └────┴────────┴─────┴───────┴───────┴─────┴────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Pending Orders                                       │  │
│  │  ┌────┬────────┬──────┬───────┬───────┬─────────┐  │  │
│  │  │ ID │ Symbol │ Type │Volume │ Price │ Actions  │  │  │
│  │  ├────┼────────┼──────┼───────┼───────┼─────────┤  │  │
│  │  │501 │EURUSD  │LIMIT │  1.0  │1.08200│[Edit][×]│  │  │
│  │  └────┴────────┴──────┴───────┴───────┴─────────┘  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Trade History                                        │  │
│  │  ┌────┬────────┬─────┬───────┬─────┬──────┬──────┐ │  │
│  │  │ ID │ Symbol │Side │ Volume│Price│ P&L  │ Time │ │  │
│  │  ├────┼────────┼─────┼───────┼─────┼──────┼──────┤ │  │
│  │  │122 │EURUSD  │ BUY │  1.0  │1.084│+125.0│12:30 │ │  │
│  │  │121 │GBPUSD  │SELL │  2.0  │1.263│ -80.0│11:15 │ │  │
│  │  └────┴────────┴─────┴───────┴─────┴──────┴──────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 7.2 Advanced Order Entry

```
┌───────────────────────────────────────────────────────┐
│  Advanced Order Entry                        [Close]  │
├───────────────────────────────────────────────────────┤
│                                                        │
│  Symbol: [EURUSD ▼]                                   │
│                                                        │
│  Order Type:                                          │
│  ( ) Market  (•) Limit  ( ) Stop  ( ) Stop-Limit     │
│                                                        │
│  Side:                                                │
│  (•) BUY     ( ) SELL                                 │
│                                                        │
│  Volume: [1.0] lots                                   │
│  (Risk Calculator: 2% risk = 0.8 lots for 20 pip SL) │
│                                                        │
│  Entry Price: [1.08200]                               │
│  (Current Ask: 1.08452)                               │
│                                                        │
│  Stop Loss:                                           │
│  [×] Enable  Price: [1.08100]  (-100 pips)           │
│                                                        │
│  Take Profit:                                         │
│  [×] Enable  Price: [1.08500]  (+300 pips)           │
│               Risk/Reward: 1:3                        │
│                                                        │
│  Trailing Stop:                                       │
│  [ ] Enable  Type: [Fixed ▼]  Distance: [20] pips    │
│                                                        │
│  Expiration:                                          │
│  [ ] Good Till Cancelled (GTC)                        │
│  [×] Good Till Date: [2026-01-25 23:59]              │
│                                                        │
│  ┌──────────────────────────────────────────────────┐│
│  │  Margin Preview                                   ││
│  │  Required Margin: $100.00                        ││
│  │  Available Margin: $10,000.00                    ││
│  │  Margin Level After: 1050%                       ││
│  │  Max Volume Available: 100.0 lots                ││
│  └──────────────────────────────────────────────────┘│
│                                                        │
│  [  Cancel  ]                     [  Place Order  ]   │
└───────────────────────────────────────────────────────┘
```

### 7.3 Position Management Modal

```
┌───────────────────────────────────────────────────────┐
│  Manage Position #12345                      [Close]  │
├───────────────────────────────────────────────────────┤
│                                                        │
│  Symbol: EURUSD                                       │
│  Side: BUY                                            │
│  Volume: 1.0 lots                                     │
│  Open Price: 1.08450                                  │
│  Current Price: 1.08500                               │
│  Unrealized P&L: +$50.00                              │
│  Open Time: 2026-01-18 14:30:00                       │
│                                                        │
│  ┌──────────────────────────────────────────────────┐│
│  │  Modify Stop Loss & Take Profit                  ││
│  │                                                   ││
│  │  Stop Loss:                                       ││
│  │  Current: 1.08300  New: [1.08350]  (-15 pips)   ││
│  │  [ ] Move to Breakeven                           ││
│  │                                                   ││
│  │  Take Profit:                                     ││
│  │  Current: 1.08700  New: [1.08750]  (+30 pips)   ││
│  │                                                   ││
│  │  [  Update SL/TP  ]                              ││
│  └──────────────────────────────────────────────────┘│
│                                                        │
│  ┌──────────────────────────────────────────────────┐│
│  │  Trailing Stop                                    ││
│  │  [ ] Enable Trailing Stop                        ││
│  │  Type: [Fixed ▼]  Distance: [20] pips            ││
│  │  Step: [10] pips                                  ││
│  │  [  Set Trailing Stop  ]                         ││
│  └──────────────────────────────────────────────────┘│
│                                                        │
│  ┌──────────────────────────────────────────────────┐│
│  │  Close Position                                   ││
│  │                                                   ││
│  │  Close Volume:                                    ││
│  │  (•) Full (1.0 lots)                             ││
│  │  ( ) Partial: [0.5] lots (50%)                   ││
│  │                                                   ││
│  │  [  Close 25%  ] [  Close 50%  ] [  Close 100% ]││
│  └──────────────────────────────────────────────────┘│
│                                                        │
│  [  Close Modal  ]                                    │
└───────────────────────────────────────────────────────┘
```

### 7.4 Account Statement

```
┌─────────────────────────────────────────────────────────────┐
│  Account Statement - RTX-000001                              │
│  Period: 2026-01-01 to 2026-01-31                    [Export]│
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Opening Balance: $10,000.00                                 │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Date       │ Type       │ Amount    │ Balance       │  │
│  ├─────────────┼────────────┼───────────┼───────────────┤  │
│  │ 2026-01-01  │ Deposit    │ +10,000.00│  10,000.00   │  │
│  │ 2026-01-02  │ Realized PnL│   +125.00│  10,125.00   │  │
│  │ 2026-01-02  │ Commission │    -10.00│  10,115.00   │  │
│  │ 2026-01-03  │ Realized PnL│    -80.00│  10,035.00   │  │
│  │ 2026-01-03  │ Swap       │     -2.50│  10,032.50   │  │
│  │ 2026-01-05  │ Realized PnL│   +200.00│  10,232.50   │  │
│  │ 2026-01-08  │ Withdrawal │  -1,000.00│   9,232.50   │  │
│  │ 2026-01-10  │ Bonus      │   +500.00│   9,732.50   │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  Closing Balance: $9,732.50                                  │
│                                                               │
│  Summary:                                                    │
│  Total Deposits:      $10,000.00                            │
│  Total Withdrawals:   $1,000.00                             │
│  Total Realized P&L:  $245.00                               │
│  Total Commission:    $10.00                                │
│  Total Swap:          $2.50                                 │
│  Total Bonus:         $500.00                               │
│  Net Change:          -$267.50 (-2.68%)                     │
│                                                               │
│  [  Download PDF  ]  [  Download CSV  ]                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 8. Admin Control Panel Design

### 8.1 Dashboard Overview

```
┌─────────────────────────────────────────────────────────────┐
│  RTX Admin Panel                      Admin User [Logout]   │
├─────────────────────────────────────────────────────────────┤
│  [Dashboard] [Accounts] [Trading] [LPs] [System] [Reports]  │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  System Health                                        │  │
│  │  Status: ● OPERATIONAL                               │  │
│  │  Uptime: 99.98%    Last Restart: 2026-01-01 00:00   │  │
│  │                                                        │  │
│  │  Services:                                            │  │
│  │  ● API Gateway      ● OMS           ● FIX Gateway    │  │
│  │  ● WebSocket        ● B-Book Engine ● LP Manager     │  │
│  │  ● PostgreSQL       ● TimescaleDB   ● Redis          │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Today's Trading Summary                              │  │
│  │  Total Volume: $15,432,500                           │  │
│  │  Total Trades: 1,247                                 │  │
│  │  Active Clients: 89                                  │  │
│  │  Open Positions: 234                                 │  │
│  │  Pending Orders: 56                                  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Execution Breakdown                                  │  │
│  │  A-Book (LP):  45%  ($6,944,625)                     │  │
│  │  B-Book (Int): 55%  ($8,487,875)                     │  │
│  │                                                        │  │
│  │  [View Detailed Routing Report]                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Liquidity Provider Status                            │  │
│  │  YOFX1:    ● Connected  Latency: 12ms   Quotes: 1.2K│  │
│  │  OANDA:    ● Connected  Latency: 45ms   Quotes: 850 │  │
│  │  Binance:  ● Connected  Latency: 8ms    Quotes: 2.1K│  │
│  │  Prime XM: ○ Disconnected                            │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Recent Alerts                                        │  │
│  │  🔴 High exposure on EURUSD: $850K (80% limit)       │  │
│  │  🟡 Client #12345 margin level: 105% (near call)    │  │
│  │  🟢 Daily backup completed successfully              │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 8.2 Account Management

```
┌─────────────────────────────────────────────────────────────┐
│  Account Management                                          │
├─────────────────────────────────────────────────────────────┤
│  Search: [___________]  Type: [All ▼]  Status: [All ▼]     │
│  [+ Create New Account]                                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ ID  │Account    │Username│Balance  │Status │Actions │  │
│  ├─────┼───────────┼────────┼─────────┼───────┼────────┤  │
│  │12345│RTX-000001 │john_doe│$10,250.0│ACTIVE │[Edit]  │  │
│  │     │Demo       │        │         │       │[Suspend│  │
│  │     │Leverage: 100x     Margin: HEDGING     │Deposit]│  │
│  │     │Created: 2026-01-01              │[Withdraw]│  │
│  ├─────┼───────────┼────────┼─────────┼───────┼────────┤  │
│  │12346│RTX-000002 │jane_sm │$50,000.0│ACTIVE │[Edit]  │  │
│  │     │Live       │        │         │       │[Suspend│  │
│  │     │Leverage: 50x      Margin: NETTING      │Deposit]│  │
│  │     │Created: 2026-01-05              │[Withdraw]│  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  Showing 1-25 of 1,247 accounts      [<] [1] [2] [3] [>]   │
└─────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────┐
│  Create New Account                          [Close]  │
├───────────────────────────────────────────────────────┤
│  Username: [____________]                             │
│  Email:    [____________]                             │
│  Password: [____________] [Generate Random]           │
│                                                        │
│  Account Type:                                        │
│  (•) Demo     ( ) Live                                │
│                                                        │
│  Initial Deposit: [$5,000.00]                         │
│  (Demo default: $5,000, Live: Awaiting deposit)      │
│                                                        │
│  Leverage: [100 ▼]  (Options: 1:1 to 1:500)          │
│  Margin Mode: [HEDGING ▼]  (HEDGING or NETTING)      │
│  Currency: [USD ▼]                                    │
│                                                        │
│  [  Cancel  ]                   [  Create Account  ]  │
└───────────────────────────────────────────────────────┘
```

### 8.3 Execution Model Configuration

```
┌─────────────────────────────────────────────────────────────┐
│  Execution Model Configuration                               │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Global Execution Mode                                │  │
│  │  (•) A-Book (STP to LP)                              │  │
│  │  ( ) B-Book (Internal Book)                          │  │
│  │  ( ) C-Book (Hybrid)                                 │  │
│  │                                                        │  │
│  │  Apply to: ( ) All Clients  (•) New Clients Only     │  │
│  │  [  Update Global Mode  ]                            │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Client-Specific Overrides                            │  │
│  │  [+ Add Client Override]                             │  │
│  │                                                        │  │
│  │  Client ID │ Username   │ Mode   │ Reason   │Actions│  │
│  │  ──────────┼────────────┼────────┼──────────┼──────│  │
│  │  12345     │ john_doe   │ A-BOOK │ VIP      │[Edit]│  │
│  │  12389     │ profitable │ A-BOOK │ Winning  │[Edit]│  │
│  │  12456     │ beginner   │ B-BOOK │ Small Vol│[Edit]│  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Symbol-Based Routing Rules                           │  │
│  │  [+ Add Symbol Rule]                                 │  │
│  │                                                        │  │
│  │  Symbol │ Default Mode│Max B-Book Vol│Priority LP │  │
│  │  ───────┼─────────────┼──────────────┼────────────│  │
│  │  EURUSD │ C-BOOK      │ 10.0 lots    │ YOFX1      │  │
│  │  BTCUSD │ A-BOOK      │ N/A          │ Binance    │  │
│  │  GBPJPY │ B-BOOK      │ 5.0 lots     │ OANDA      │  │
│  │  XAUUSD │ A-BOOK      │ N/A          │ YOFX1      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Routing Logic Configuration                          │  │
│  │                                                        │  │
│  │  B-Book Max Volume: [10.0] lots per order            │  │
│  │  High Volatility Threshold: [5]% (route to A-Book)   │  │
│  │  Max Symbol Exposure: [$1,000,000]                   │  │
│  │                                                        │  │
│  │  Profitable Trader Auto-Route:                       │  │
│  │  [×] Enable  Threshold: [+$10,000] or [>65%] win rate│  │
│  │                                                        │  │
│  │  [  Save Routing Configuration  ]                    │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 8.4 Symbol Management

```
┌─────────────────────────────────────────────────────────────┐
│  Symbol Management                                           │
├─────────────────────────────────────────────────────────────┤
│  Search: [___________]  Category: [All ▼]  [+ Add Symbol]  │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Symbol  │Enabled│Spread │Commission│Min  │Max  │Actions │  │
│          │       │Markup │Per Lot   │Vol  │Vol  │        │  │
│  ────────┼───────┼───────┼──────────┼─────┼─────┼────────│  │
│  EURUSD  │ ✓     │ +0.2  │ $7.00    │ 0.01│100.0│[Edit]  │  │
│  GBPUSD  │ ✓     │ +0.3  │ $7.00    │ 0.01│100.0│[Edit]  │  │
│  USDJPY  │ ✓     │ +0.2  │ $7.00    │ 0.01│100.0│[Edit]  │  │
│  BTCUSD  │ ✓     │ +$5   │ $10.00   │ 0.01│ 10.0│[Edit]  │  │
│  ETHUSD  │ ✓     │ +$2   │ $10.00   │ 0.01│ 50.0│[Edit]  │  │
│  XAUUSD  │ ✗     │ +$0.1 │ $15.00   │ 0.01│ 20.0│[Edit]  │  │
│                                                               │
└─────────────────────────────────────────────────────────────┘

┌───────────────────────────────────────────────────────┐
│  Edit Symbol: EURUSD                         [Close]  │
├───────────────────────────────────────────────────────┤
│  Enabled: [×] (Visible to clients)                    │
│                                                        │
│  ─── Pricing ───                                      │
│  Spread Markup: [+0.2] pips                           │
│  Commission Per Lot: [$7.00]                          │
│                                                        │
│  ─── Trading Limits ───                               │
│  Min Volume: [0.01] lots                              │
│  Max Volume: [100.0] lots                             │
│  Volume Step: [0.01] lots                             │
│                                                        │
│  ─── Margin ───                                       │
│  Contract Size: [100000]                              │
│  Margin Percentage: [1]%                              │
│  Pip Size: [0.0001]                                   │
│  Pip Value: [$10.00] (per 1.0 lot)                   │
│                                                        │
│  ─── Swap (Rollover) ───                              │
│  Swap Long: [-2.50] USD per lot per day              │
│  Swap Short: [+0.80] USD per lot per day             │
│  Swap 3-Day: [Wednesday ▼]                           │
│                                                        │
│  ─── Trading Hours ───                                │
│  [ ] 24/5 (Sunday 22:00 - Friday 22:00 GMT)          │
│  [×] Custom Schedule                                  │
│     Sunday:    [22:00] - [23:59]                     │
│     Monday:    [00:00] - [23:59]                     │
│     ...                                               │
│     Friday:    [00:00] - [22:00]                     │
│                                                        │
│  [  Cancel  ]                      [  Save Changes  ] │
└───────────────────────────────────────────────────────┘
```

### 8.5 Financial Operations

```
┌─────────────────────────────────────────────────────────────┐
│  Financial Operations                                        │
├─────────────────────────────────────────────────────────────┤
│  [Deposits] [Withdrawals] [Adjustments] [Bonuses]           │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Process Deposit                                      │  │
│  │                                                        │  │
│  │  Account: [Search by ID or Username ▼]               │  │
│  │  Selected: RTX-000001 (john_doe)                     │  │
│  │  Current Balance: $10,250.00                         │  │
│  │                                                        │  │
│  │  Amount: [$__________]                               │  │
│  │                                                        │  │
│  │  Method:                                              │  │
│  │  (•) Bank Transfer  ( ) Credit Card  ( ) Crypto      │  │
│  │                                                        │  │
│  │  Reference Number: [____________]                     │  │
│  │  (Transaction ID from payment processor)             │  │
│  │                                                        │  │
│  │  Note: [____________]                                 │  │
│  │  (Optional internal note)                            │  │
│  │                                                        │  │
│  │  [  Process Deposit  ]                               │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Pending Withdrawals (Approval Required)              │  │
│  │                                                        │  │
│  │  ID   │Account   │Amount   │Method │Requested │Action│  │
│  │  ────┼──────────┼─────────┼───────┼──────────┼──────│  │
│  │  5001│RTX-000002│$5,000.00│Bank   │2026-01-18│[App] │  │
│  │      │jane_sm   │         │       │10:30     │[Rej] │  │
│  │  5002│RTX-000045│$1,200.00│Crypto │2026-01-18│[App] │  │
│  │      │trader_x  │         │       │11:45     │[Rej] │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Manual Adjustment                                    │  │
│  │                                                        │  │
│  │  Account: [RTX-000001 ▼]                             │  │
│  │  Current Balance: $10,250.00                         │  │
│  │                                                        │  │
│  │  Adjustment Amount: [$__________]                    │  │
│  │  (Use + for credit, - for debit)                    │  │
│  │                                                        │  │
│  │  Type: [Manual Adjustment ▼]                         │  │
│  │  Options: Manual Adjustment, Error Correction,       │  │
│  │           Promotional Credit                          │  │
│  │                                                        │  │
│  │  Reason: [_____________________________]             │  │
│  │  (Required - will be shown in ledger)               │  │
│  │                                                        │  │
│  │  [  Apply Adjustment  ]                              │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 8.6 LP Management

```
┌─────────────────────────────────────────────────────────────┐
│  Liquidity Provider Management                               │
├─────────────────────────────────────────────────────────────┤
│  [+ Add New LP]                                              │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  YOFX1 (FIX 4.4 Trading)                             │  │
│  │  Status: ● CONNECTED  Latency: 12ms  Uptime: 99.95% │  │
│  │                                                        │  │
│  │  Session Details:                                     │  │
│  │  SenderCompID: YOFX1       TargetCompID: YOFX        │  │
│  │  Host: 23.106.238.138:12336                          │  │
│  │  Heartbeat: 30s            Reconnect: 5s             │  │
│  │                                                        │  │
│  │  Quotes Received: 1,247 (last 1 min)                 │  │
│  │  Symbols: 150 active                                  │  │
│  │                                                        │  │
│  │  [Disconnect] [Reconnect] [View Logs] [Edit Config] │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  OANDA (REST API)                                     │  │
│  │  Status: ● CONNECTED  Latency: 45ms  Uptime: 99.80% │  │
│  │                                                        │  │
│  │  API Details:                                         │  │
│  │  Environment: Practice                                │  │
│  │  Account ID: 101-004-37008470-002                    │  │
│  │  API Endpoint: api-fxpractice.oanda.com              │  │
│  │                                                        │  │
│  │  Quotes Received: 850 (last 1 min)                   │  │
│  │  Symbols: 120 active                                  │  │
│  │                                                        │  │
│  │  [Disconnect] [Reconnect] [View Logs] [Edit Config] │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Binance (WebSocket)                                  │  │
│  │  Status: ● CONNECTED  Latency: 8ms   Uptime: 99.99% │  │
│  │                                                        │  │
│  │  WebSocket Details:                                   │  │
│  │  URL: wss://stream.binance.com:9443/ws               │  │
│  │  Streams: 5 active                                    │  │
│  │                                                        │  │
│  │  Quotes Received: 2,134 (last 1 min)                 │  │
│  │  Symbols: 5 (BTCUSD, ETHUSD, BNBUSD, SOLUSD, XRPUSD)│  │
│  │                                                        │  │
│  │  [Disconnect] [Reconnect] [View Logs] [Edit Config] │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Prime XM (FIX 4.4 Trading)                          │  │
│  │  Status: ○ DISCONNECTED                              │  │
│  │                                                        │  │
│  │  Error: Connection timeout after 3 retries           │  │
│  │  Last Connected: 2026-01-17 14:30:00                 │  │
│  │                                                        │  │
│  │  [Connect] [Edit Config] [Remove LP]                │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 9. Security & Compliance

### 9.1 Authentication & Authorization

**JWT-based Authentication:**
```go
type JWTClaims struct {
    UserID    int64  `json:"userId"`
    AccountID int64  `json:"accountId"`
    Role      string `json:"role"` // ADMIN, CLIENT
    Expires   int64  `json:"exp"`
}

// Token expiry: 24 hours for clients, 8 hours for admins
```

**Role-Based Access Control (RBAC):**
```
Roles:
- SUPER_ADMIN: Full system access, can modify all settings
- ADMIN: Account management, financial operations, view reports
- RISK_MANAGER: View positions, modify risk parameters
- SUPPORT: View accounts, view ledgers (read-only)
- CLIENT: Trading operations only
```

**API Endpoints Protection:**
```go
// Middleware stack
request -> CORS -> RateLimiter -> AuthValidator -> RoleChecker -> Handler

// Example
router.POST("/admin/deposit",
    corsMiddleware(),
    rateLimiter(10, time.Minute),
    authMiddleware(),
    roleChecker("ADMIN", "SUPER_ADMIN"),
    adminHandler.HandleDeposit,
)
```

### 9.2 Data Encryption

**In Transit:**
- TLS 1.3 for all HTTP/WebSocket connections
- Separate certificates for admin and client APIs
- Forward secrecy enabled

**At Rest:**
- PostgreSQL: Transparent Data Encryption (TDE)
- Backup files: AES-256 encryption
- Sensitive fields (passwords): bcrypt hashing with salt

**API Keys:**
- Stored in environment variables (never in code)
- Rotated every 90 days
- Scoped to specific LP accounts

### 9.3 Audit Logging

**All critical operations logged:**
```go
type AuditLog struct {
    ID           int64     `json:"id"`
    Timestamp    time.Time `json:"timestamp"`
    UserID       int64     `json:"userId"`
    Action       string    `json:"action"` // LOGIN, DEPOSIT, ORDER_PLACED, etc.
    Resource     string    `json:"resource"` // ACCOUNT, ORDER, POSITION
    ResourceID   int64     `json:"resourceId"`
    Changes      string    `json:"changes"` // JSON of before/after
    IPAddress    string    `json:"ipAddress"`
    UserAgent    string    `json:"userAgent"`
}

// Retention: 7 years (regulatory compliance)
```

### 9.4 Compliance Features

**KYC/AML:**
- Client identity verification
- Document upload and verification workflow
- Transaction monitoring for suspicious activity
- Automated alerts for large withdrawals

**Regulatory Reporting:**
- Daily trade reports
- Monthly account statements
- Regulatory filings (MiFID II, EMIR, Dodd-Frank)

---

## 10. Performance & Scalability

### 10.1 Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Order Execution Latency | <50ms | Time from API request to position created |
| Quote Latency | <100ms | Time from LP to client WebSocket |
| API Response Time (p99) | <200ms | 99th percentile |
| WebSocket Message Throughput | 10,000 msg/s | Per server instance |
| Database Write Throughput | 5,000 TPS | Transactions per second |
| Concurrent Users | 10,000 | Per cluster |

### 10.2 Horizontal Scaling

**Stateless Services (Scale Out):**
- Client API (Kubernetes, 3-10 pods)
- Admin API (Kubernetes, 2-5 pods)
- WebSocket Server (Kubernetes, 5-15 pods)

**Stateful Services (Vertical + Read Replicas):**
- PostgreSQL (Master-Slave replication, 2 read replicas)
- Redis (Redis Cluster, 3 master + 3 replica nodes)
- TimescaleDB (Partitioning by time, compression policies)

**Message Queue:**
- RabbitMQ Cluster (3 nodes, mirrored queues)

### 10.3 Caching Strategy

```
L1 Cache (In-Memory, per service):
  - Symbol specs (1 hour TTL)
  - Account state (5 minutes TTL)

L2 Cache (Redis Cluster):
  - Latest quotes (60 seconds TTL)
  - Position summaries (30 seconds TTL)
  - Session tokens (24 hours TTL)

Database Read Replicas:
  - Historical data queries
  - Reporting and analytics
```

### 10.4 Database Optimization

**PostgreSQL:**
- Connection pooling (PgBouncer)
- Partitioning (trades table by month)
- Indexes on frequently queried columns
- Vacuum and analyze scheduled

**TimescaleDB:**
- Hypertables for ticks and OHLC
- Continuous aggregates for OHLC generation
- Compression policies (older than 7 days)
- Retention policies (delete older than 2 years)

---

## 11. Deployment Architecture

### 11.1 Infrastructure

```
┌─────────────────────────────────────────────────────────────┐
│                    PRODUCTION CLUSTER                        │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Load Balancer (HAProxy/NGINX)                       │  │
│  │  - SSL Termination                                   │  │
│  │  - Health Checks                                     │  │
│  │  - Request Routing                                   │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Kubernetes Cluster (3 Master, 6 Worker Nodes)       │  │
│  │                                                        │  │
│  │  Namespaces:                                          │  │
│  │  - trading (Client API, OMS, Order Router)           │  │
│  │  - admin (Admin API)                                  │  │
│  │  - market-data (WebSocket, LP Manager)               │  │
│  │  - execution (B-Book Engine, FIX Gateway)            │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Database Cluster                                     │  │
│  │  - PostgreSQL Master (16 vCPU, 64GB RAM, SSD)       │  │
│  │  - PostgreSQL Replica 1                              │  │
│  │  - PostgreSQL Replica 2                              │  │
│  │  - TimescaleDB (32 vCPU, 128GB RAM, NVMe)           │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Redis Cluster                                        │  │
│  │  - 3 Master Nodes                                     │  │
│  │  - 3 Replica Nodes                                    │  │
│  │  - Sentinel for failover                             │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  RabbitMQ Cluster                                     │  │
│  │  - 3 Nodes (Mirrored Queues)                         │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 11.2 Disaster Recovery

**Backup Strategy:**
- PostgreSQL: Full backup daily, incremental every 6 hours
- TimescaleDB: Continuous aggregates snapshots daily
- Redis: RDB snapshots every hour
- Configuration files: Version controlled in Git

**RTO/RPO Targets:**
- Recovery Time Objective (RTO): 1 hour
- Recovery Point Objective (RPO): 15 minutes

**Failover Plan:**
1. Automated database failover (PostgreSQL HA)
2. Kubernetes pod auto-restart
3. Multi-region DR site (active-passive)

---

## 12. Monitoring & Observability

### 12.1 Metrics Collection

**Prometheus + Grafana Stack:**
```yaml
metrics:
  - order_execution_latency (histogram)
  - websocket_connections (gauge)
  - lp_quote_latency (histogram)
  - database_query_duration (histogram)
  - http_requests_total (counter)
  - active_positions (gauge)
  - daily_volume (counter)
  - error_rate (counter)
```

**Dashboards:**
- System Health (CPU, Memory, Disk, Network)
- Trading Activity (Volume, Orders, Positions)
- LP Performance (Latency, Uptime, Quote Freshness)
- Error Rates (API errors, DB errors, LP failures)

### 12.2 Alerting

**PagerDuty Alerts:**
- Critical: DB down, All LPs disconnected, High error rate
- Warning: Single LP down, High latency, Low disk space
- Info: Daily backup complete, Deployment complete

### 12.3 Distributed Tracing

**OpenTelemetry + Jaeger:**
- Trace order lifecycle (API → OMS → Router → Execution → DB)
- Identify bottlenecks
- Debug cross-service issues

---

## Appendix A: Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Backend Language | Go | 1.21+ |
| API Framework | Gin | 1.9+ |
| FIX Engine | quickfix-go | Latest |
| Database (Transactional) | PostgreSQL | 16+ |
| Database (Time-Series) | TimescaleDB | 2.14+ |
| Cache | Redis | 7.2+ |
| Message Queue | RabbitMQ | 3.12+ |
| Container Orchestration | Kubernetes | 1.28+ |
| Monitoring | Prometheus + Grafana | Latest |
| Tracing | Jaeger | Latest |
| Load Balancer | NGINX | 1.25+ |

---

## Appendix B: API Endpoint Summary

### Client API
```
Authentication:
  POST   /api/login
  POST   /api/logout

Account:
  GET    /api/account/summary
  GET    /api/account/positions
  GET    /api/account/orders
  GET    /api/account/trades
  GET    /api/account/ledger

Market Data:
  GET    /api/symbols
  GET    /api/ticks?symbol=EURUSD&limit=500
  GET    /api/ohlc?symbol=EURUSD&timeframe=1h&limit=500
  WS     /ws

Orders:
  POST   /api/orders/market
  POST   /api/orders/limit
  POST   /api/orders/stop
  POST   /api/orders/stop-limit
  PUT    /api/orders/{id}/modify
  DELETE /api/orders/{id}

Positions:
  POST   /api/positions/{id}/close
  POST   /api/positions/{id}/close-partial
  PUT    /api/positions/{id}/modify
  POST   /api/positions/close-all

Risk Tools:
  GET    /api/risk/calculate-lot
  GET    /api/risk/margin-preview
```

### Admin API
```
Accounts:
  GET    /admin/accounts
  POST   /admin/accounts
  PUT    /admin/accounts/{id}
  DELETE /admin/accounts/{id}
  POST   /admin/accounts/{id}/suspend
  POST   /admin/accounts/{id}/activate

Financial:
  POST   /admin/deposit
  POST   /admin/withdraw
  POST   /admin/adjust
  POST   /admin/bonus

Execution:
  GET    /admin/execution-mode
  POST   /admin/execution-mode
  GET    /admin/client-overrides
  POST   /admin/client-overrides

Symbols:
  GET    /admin/symbols
  POST   /admin/symbols
  PUT    /admin/symbols/{symbol}
  POST   /admin/symbols/{symbol}/enable
  POST   /admin/symbols/{symbol}/disable

LPs:
  GET    /admin/lps
  POST   /admin/lps
  PUT    /admin/lps/{id}
  DELETE /admin/lps/{id}
  POST   /admin/lps/{id}/connect
  POST   /admin/lps/{id}/disconnect

FIX:
  GET    /admin/fix/status
  POST   /admin/fix/connect
  POST   /admin/fix/disconnect

System:
  GET    /admin/system/health
  GET    /admin/system/metrics
  POST   /admin/system/restart
```

---

**Document Status:** Design Complete
**Next Steps:** Implementation Phase, Database Schema Creation, Service Development
