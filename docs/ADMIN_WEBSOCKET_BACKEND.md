# Admin WebSocket Backend Implementation Guide

## Overview

This document specifies the backend WebSocket endpoint required for the Admin Panel real-time updates. The frontend expects a WebSocket server at `/admin-ws` that provides MT5 Manager parity for account and order monitoring.

---

## Endpoint Specification

### WebSocket URL
```
ws://localhost:7999/admin-ws
```

### Authentication
The WebSocket connection expects authentication via query parameter:
```
ws://localhost:7999/admin-ws?token={JWT_TOKEN}
```

**Token Sources**:
- `admin_token` from localStorage (primary)
- `rtx_token` from localStorage (fallback)

**Authentication Failure**:
- Close code: `1008` or `401`
- Client will NOT auto-reconnect on auth failure
- Client displays "Mock Data" indicator

---

## Message Protocol

All messages are JSON-formatted with the following structure:

### Client → Server Messages

#### 1. Subscribe to Channels
```json
{
  "type": "subscribe",
  "channels": ["accounts", "orders", "positions", "system"]
}
```

**Sent automatically on connection.**

#### 2. Ping (Keep-Alive)
```json
{
  "type": "ping"
}
```

**Sent every 30 seconds to keep connection alive.**

---

### Server → Client Messages

#### 1. Account Update Event
```json
{
  "type": "account_update",
  "timestamp": 1705747200000,
  "data": {
    "id": 12345,
    "login": "1000001",
    "name": "John Trader",
    "group": "RETAIL",
    "leverage": "1:100",
    "balance": "10000.00",
    "credit": "0.00",
    "equity": "10523.45",
    "margin": "1200.00",
    "freeMargin": "9323.45",
    "marginLevel": "876.95",
    "profit": "523.45",
    "floatingPL": "523.45",
    "swap": "-2.50",
    "commission": "0.00",
    "status": "ACTIVE",
    "currency": "USD",
    "country": "US",
    "email": "john@example.com",
    "comment": "VIP Client"
  }
}
```

**Account Status Values**:
- `ACTIVE` - Normal trading
- `SUSPENDED` - Trading disabled
- `MARGIN_CALL` - Margin level below threshold

**Trigger Conditions**:
- Balance change (deposit, withdrawal, P/L realization)
- Equity change (floating P/L update)
- Margin level change
- Status change (suspension, margin call)

**Update Frequency**: Real-time (as events occur)

---

#### 2. Order New Event
```json
{
  "type": "order_new",
  "timestamp": 1705747200000,
  "data": {
    "id": 67890,
    "login": "1000001",
    "symbol": "EURUSD",
    "type": "BUY",
    "volume": 1.0,
    "price": 1.08450,
    "sl": 1.08200,
    "tp": 1.08800,
    "profit": 0.00,
    "openTime": "2026-01-20T10:30:00Z"
  }
}
```

**Order Types**: `BUY`, `SELL`, `BUY_LIMIT`, `SELL_LIMIT`, `BUY_STOP`, `SELL_STOP`

---

#### 3. Order Modify Event
```json
{
  "type": "order_modify",
  "timestamp": 1705747200000,
  "data": {
    "id": 67890,
    "sl": 1.08300,
    "tp": 1.08900,
    "profit": 25.50
  }
}
```

**Modified Fields**: Any subset of order fields (SL, TP, profit, etc.)

---

#### 4. Order Close Event
```json
{
  "type": "order_close",
  "timestamp": 1705747200000,
  "data": {
    "id": 67890
  }
}
```

---

#### 5. Position Update Event
```json
{
  "type": "position_update",
  "timestamp": 1705747200000,
  "data": {
    "login": "1000001",
    "symbol": "EURUSD",
    "volume": 1.5,
    "avgPrice": 1.08400,
    "currentPrice": 1.08550,
    "profit": 225.00,
    "swap": -5.50,
    "commission": 0.00
  }
}
```

---

#### 6. Margin Call Event
```json
{
  "type": "margin_call",
  "timestamp": 1705747200000,
  "data": {
    "login": "1000001",
    "marginLevel": 85.30,
    "equity": 9523.45,
    "margin": 11200.00,
    "action": "WARNING"
  }
}
```

**Action Types**:
- `WARNING` - Margin level below 100%
- `STOP_OUT` - Margin level below 50% (close positions)

---

#### 7. Account Suspended Event
```json
{
  "type": "account_suspended",
  "timestamp": 1705747200000,
  "data": {
    "login": "1000001",
    "reason": "Manual suspension by admin",
    "suspendedBy": "admin_user"
  }
}
```

---

#### 8. System Event
```json
{
  "type": "system_event",
  "timestamp": 1705747200000,
  "data": {
    "level": "INFO",
    "message": "Market hours updated",
    "affectedAccounts": []
  }
}
```

**System Event Levels**: `INFO`, `WARNING`, `ERROR`, `CRITICAL`

---

## Connection Lifecycle

### 1. Client Connection Flow
```
1. Client creates WebSocket: new WebSocket(ws://host/admin-ws?token=...)
2. Server validates token
3. Server sends confirmation (implicit - connection stays open)
4. Client sends subscribe message
5. Server starts sending real-time events
6. Client sends ping every 30 seconds
7. Server responds to pings (optional pong)
```

### 2. Reconnection Logic
```
Initial Connection:
  ├── Success → Connected
  └── Failure → Retry with exponential backoff

Disconnection (code != 1008, 401):
  ├── Attempt 1: 1 second delay
  ├── Attempt 2: 2 second delay
  ├── Attempt 3: 4 second delay
  ├── ... (exponential)
  └── Attempt 10: 30 seconds (max delay)

After 10 failures: Display error, stop reconnecting
```

### 3. Fallback Polling
If WebSocket disconnected, client automatically falls back to HTTP polling:
```
GET /api/admin/accounts  (every 2 seconds)
GET /api/admin/orders    (every 2 seconds)
```

**Backend should implement both WebSocket and HTTP endpoints for resilience.**

---

## Backend Implementation (Go Example)

### WebSocket Handler Structure
```go
// admin/broker-admin/backend/ws/admin_hub.go

type AdminHub struct {
    clients    map[*AdminClient]bool
    broadcast  chan AdminEvent
    register   chan *AdminClient
    unregister chan *AdminClient
}

type AdminClient struct {
    hub      *AdminHub
    conn     *websocket.Conn
    send     chan []byte
    userID   string
    channels []string
}

type AdminEvent struct {
    Type      string      `json:"type"`
    Timestamp int64       `json:"timestamp"`
    Data      interface{} `json:"data"`
}

func (h *AdminHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true

        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }

        case event := <-h.broadcast:
            // Broadcast to all connected admin clients
            message, _ := json.Marshal(event)
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}

func (c *AdminClient) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    for {
        var msg map[string]interface{}
        err := c.conn.ReadJSON(&msg)
        if err != nil {
            break
        }

        // Handle subscribe message
        if msg["type"] == "subscribe" {
            channels := msg["channels"].([]interface{})
            for _, ch := range channels {
                c.channels = append(c.channels, ch.(string))
            }
        }
    }
}

func (c *AdminClient) writePump() {
    defer c.conn.Close()

    for message := range c.send {
        c.conn.WriteMessage(websocket.TextMessage, message)
    }
}
```

### Integration Points

#### 1. Account Balance Updates
```go
// When account balance changes (deposit, withdrawal, trade close)
hub.broadcast <- AdminEvent{
    Type:      "account_update",
    Timestamp: time.Now().UnixMilli(),
    Data: map[string]interface{}{
        "id":      account.ID,
        "login":   account.Login,
        "balance": account.Balance.String(),
        "equity":  account.Equity.String(),
        // ... other fields
    },
}
```

#### 2. New Order Placement
```go
// When trader places new order
hub.broadcast <- AdminEvent{
    Type:      "order_new",
    Timestamp: time.Now().UnixMilli(),
    Data: map[string]interface{}{
        "id":       order.ID,
        "login":    order.Login,
        "symbol":   order.Symbol,
        "type":     order.Type,
        "volume":   order.Volume,
        "price":    order.Price,
        "sl":       order.SL,
        "tp":       order.TP,
        "openTime": order.OpenTime.Format(time.RFC3339),
    },
}
```

#### 3. Position P/L Updates
```go
// On every tick that affects open positions
for _, position := range positions {
    hub.broadcast <- AdminEvent{
        Type:      "position_update",
        Timestamp: time.Now().UnixMilli(),
        Data: map[string]interface{}{
            "login":        position.Login,
            "symbol":       position.Symbol,
            "profit":       position.Profit,
            "currentPrice": currentTick.Bid,
        },
    }
}
```

---

## HTTP Fallback Endpoints

### GET /api/admin/accounts
```json
[
  {
    "id": 12345,
    "login": "1000001",
    "name": "John Trader",
    "balance": "10000.00",
    "equity": "10523.45",
    "profit": "523.45",
    "status": "ACTIVE"
  }
]
```

### GET /api/admin/orders
```json
[
  {
    "id": 67890,
    "login": "1000001",
    "symbol": "EURUSD",
    "type": "BUY",
    "volume": 1.0,
    "price": 1.08450,
    "profit": 25.50
  }
]
```

---

## Testing Guide

### 1. WebSocket Connection Test
```bash
# Using wscat
npm install -g wscat
wscat -c "ws://localhost:7999/admin-ws?token=test_token"

# Expected response: Connection established
# Send subscribe message:
{"type":"subscribe","channels":["accounts","orders"]}
```

### 2. Event Simulation
```bash
# Trigger account update (via backend API or direct DB)
curl -X POST http://localhost:7999/api/admin/test/account-update \
  -H "Content-Type: application/json" \
  -d '{"login":"1000001","balance":"15000.00"}'

# Expected: WebSocket clients receive account_update event
```

### 3. Frontend Integration Test
1. Start backend server with `/admin-ws` endpoint
2. Open admin panel: `http://localhost:3001`
3. Check connection indicator:
   - Should show "Connecting..." initially
   - Then "Live" with green indicator when connected
   - Shows real-time updates in accounts table

### 4. Fallback Test
1. Stop WebSocket server (keep HTTP endpoints running)
2. Admin panel should automatically switch to "Mock Data" indicator
3. Polling should activate (check Network tab - requests every 2s)
4. Restart WebSocket server
5. Panel should reconnect and show "Live" again

---

## Performance Requirements

| Metric | Target | MT5 Manager Parity |
|--------|--------|-------------------|
| WebSocket connection time | <100ms | ✓ |
| Event delivery latency | <50ms | ✓ |
| Concurrent admin connections | 50+ | ✓ |
| Events per second | 1000+ | ✓ |
| Memory per connection | <5MB | ✓ |
| Reconnection time | <2s | ✓ |

---

## Security Considerations

### 1. Authentication
- **JWT token validation** on every WebSocket connection
- **Token expiration** - disconnect client on expiry
- **Rate limiting** - max 1 connection per admin user

### 2. Authorization
- Only admin users can connect to `/admin-ws`
- Filter events by admin permissions (if implemented)
- Audit log all admin connections

### 3. Data Privacy
- Don't send sensitive data (passwords, API keys)
- Mask partial credit card numbers if displayed
- Log access to account data

---

## Monitoring & Diagnostics

### Backend Metrics to Track
```
- admin_ws_connections_active (gauge)
- admin_ws_events_sent_total (counter)
- admin_ws_connection_duration_seconds (histogram)
- admin_ws_reconnections_total (counter)
- admin_ws_auth_failures_total (counter)
```

### Logging
```
[AdminWS] Client connected: user=admin_user, ip=192.168.1.100
[AdminWS] Client subscribed to channels: [accounts, orders]
[AdminWS] Broadcast event: type=account_update, affected=1
[AdminWS] Client disconnected: user=admin_user, duration=3600s
```

---

## Next Steps for Backend Implementation

1. **Create AdminHub struct** in `backend/ws/admin_hub.go`
2. **Implement WebSocket handler** in `backend/api/admin_ws.go`
3. **Wire event triggers** in existing account/order handlers
4. **Add HTTP fallback endpoints** in `backend/api/admin.go`
5. **Test with frontend** using connection indicator
6. **Add monitoring** and logging

---

**Implementation Date**: January 20, 2026
**Phase**: 2B - Admin Panel WebSocket Integration
**Status**: Backend specification complete, ready for implementation
**Frontend**: ✅ Complete (adminWebSocket.ts, useAdminData.ts, AccountsView.tsx)
**Backend**: ⏳ Awaiting implementation based on this specification
