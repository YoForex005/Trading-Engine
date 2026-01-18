# WebSocket Protocol Specification

## Overview

The RTX Trading Engine uses WebSocket protocol for real-time, bidirectional communication between the server and clients. This enables low-latency market data streaming and instant account updates.

## Connection

### Endpoint
```
ws://localhost:7999/ws
```

Production:
```
wss://api.rtxtrading.com/ws
```

### Connection Establishment

**JavaScript Example**
```javascript
const ws = new WebSocket('ws://localhost:7999/ws');

ws.onopen = function(event) {
    console.log('WebSocket connected');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

ws.onclose = function(event) {
    console.log('WebSocket disconnected:', event.code, event.reason);
};
```

**Python Example**
```python
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print(f"Received: {data}")

def on_open(ws):
    print("WebSocket connected")
    # Subscribe to symbols
    ws.send(json.dumps({
        "type": "subscribe",
        "symbols": ["EURUSD", "GBPUSD"]
    }))

ws = websocket.WebSocketApp(
    "ws://localhost:7999/ws",
    on_open=on_open,
    on_message=on_message
)

ws.run_forever()
```

**Go Example**
```go
import (
    "encoding/json"
    "github.com/gorilla/websocket"
)

conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:7999/ws", nil)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

for {
    _, message, err := conn.ReadMessage()
    if err != nil {
        log.Println("Read error:", err)
        return
    }

    var data map[string]interface{}
    json.Unmarshal(message, &data)
    fmt.Printf("Received: %+v\n", data)
}
```

## Message Types

### 1. Market Tick (Server → Client)

Real-time price updates for subscribed symbols.

**Message Structure**
```json
{
  "type": "tick",
  "symbol": "EURUSD",
  "bid": 1.09050,
  "ask": 1.09055,
  "spread": 0.00005,
  "timestamp": 1706400000000,
  "lp": "OANDA"
}
```

**Field Descriptions**
| Field | Type | Description |
|-------|------|-------------|
| type | string | Always "tick" |
| symbol | string | Trading symbol |
| bid | float | Current bid price |
| ask | float | Current ask price |
| spread | float | Difference between ask and bid |
| timestamp | int64 | Unix timestamp in milliseconds |
| lp | string | Liquidity provider source |

**Frequency**: Up to 1000 ticks per second during high volatility

**Example Flow**
```
Server broadcasts:
  EURUSD: Bid 1.09050, Ask 1.09055
  GBPUSD: Bid 1.26500, Ask 1.26505
  USDJPY: Bid 147.850, Ask 147.900
  ...
```

### 2. Position Update (Server → Client)

Sent when position opens, closes, or is modified.

**Position Opened**
```json
{
  "type": "position_opened",
  "position": {
    "id": 1234,
    "accountId": 1,
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.1,
    "openPrice": 1.09055,
    "currentPrice": 1.09055,
    "openTime": "2024-01-18T10:30:00Z",
    "sl": 1.08555,
    "tp": 1.09555,
    "unrealizedPnL": 0.0,
    "commission": 2.50,
    "status": "OPEN"
  }
}
```

**Position Closed**
```json
{
  "type": "position_closed",
  "position": {
    "id": 1234,
    "accountId": 1,
    "symbol": "EURUSD",
    "side": "BUY",
    "volume": 0.1,
    "openPrice": 1.09055,
    "closePrice": 1.09305,
    "openTime": "2024-01-18T10:30:00Z",
    "closeTime": "2024-01-18T11:45:00Z",
    "closeReason": "MANUAL",
    "realizedPnL": 25.00,
    "commission": 2.50,
    "netPnL": 22.50,
    "status": "CLOSED"
  }
}
```

**Position Modified**
```json
{
  "type": "position_modified",
  "position": {
    "id": 1234,
    "sl": 1.08750,
    "tp": 1.09750,
    "modifiedAt": "2024-01-18T11:00:00Z"
  }
}
```

**Close Reasons**
- **MANUAL**: Closed by user
- **STOP_LOSS**: SL triggered
- **TAKE_PROFIT**: TP triggered
- **MARGIN_CALL**: Insufficient margin
- **ADMIN**: Closed by administrator

### 3. Account Update (Server → Client)

Sent when account balance or equity changes.

```json
{
  "type": "account_update",
  "account": {
    "accountId": 1,
    "balance": 5000.00,
    "equity": 5025.50,
    "margin": 100.00,
    "freeMargin": 4925.50,
    "marginLevel": 5025.50,
    "unrealizedPnL": 25.50,
    "openPositions": 2
  }
}
```

**Triggers**
- Position opened
- Position closed
- Deposit/withdrawal
- Margin call
- Swap applied

### 4. Order Update (Server → Client)

Sent when order status changes.

**Order Filled**
```json
{
  "type": "order_filled",
  "order": {
    "id": 5678,
    "accountId": 1,
    "symbol": "GBPUSD",
    "type": "MARKET",
    "side": "SELL",
    "volume": 0.2,
    "status": "FILLED",
    "filledPrice": 1.26505,
    "filledAt": "2024-01-18T10:35:00Z",
    "positionId": 1235
  }
}
```

**Order Rejected**
```json
{
  "type": "order_rejected",
  "order": {
    "id": 5679,
    "accountId": 1,
    "symbol": "EURUSD",
    "type": "MARKET",
    "side": "BUY",
    "volume": 10.0,
    "status": "REJECTED",
    "rejectReason": "Insufficient margin"
  }
}
```

### 5. Heartbeat (Server ↔ Client)

Keep-alive mechanism to detect connection issues.

**Server → Client**
```json
{
  "type": "heartbeat",
  "timestamp": 1706400000000
}
```

**Client → Server (Optional)**
```json
{
  "type": "pong",
  "timestamp": 1706400000000
}
```

**Interval**: Every 30 seconds

### 6. Error Messages (Server → Client)

Sent when errors occur.

```json
{
  "type": "error",
  "code": 400,
  "message": "Invalid symbol",
  "details": "Symbol INVALID does not exist"
}
```

**Error Codes**
| Code | Description |
|------|-------------|
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

## Client → Server Messages

### Subscribe to Symbols
```json
{
  "type": "subscribe",
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"]
}
```

**Response**
```json
{
  "type": "subscribed",
  "symbols": ["EURUSD", "GBPUSD", "USDJPY"],
  "count": 3
}
```

### Unsubscribe from Symbols
```json
{
  "type": "unsubscribe",
  "symbols": ["USDJPY"]
}
```

**Response**
```json
{
  "type": "unsubscribed",
  "symbols": ["USDJPY"]
}
```

### Get Snapshot
```json
{
  "type": "snapshot",
  "symbols": ["EURUSD"]
}
```

**Response**
```json
{
  "type": "snapshot_response",
  "data": {
    "EURUSD": {
      "bid": 1.09050,
      "ask": 1.09055,
      "timestamp": 1706400000000
    }
  }
}
```

## Connection Lifecycle

### 1. Initial Connection
```
Client → Connect to ws://localhost:7999/ws
Server → Connection Accepted
Server → Send latest prices for all symbols
```

### 2. Authentication (Future)
```
Client → { "type": "auth", "token": "JWT_TOKEN" }
Server → { "type": "auth_success", "accountId": 1 }
```

### 3. Symbol Subscription
```
Client → { "type": "subscribe", "symbols": ["EURUSD", "GBPUSD"] }
Server → { "type": "subscribed", "symbols": ["EURUSD", "GBPUSD"] }
Server → Broadcast ticks for subscribed symbols
```

### 4. Receiving Updates
```
Server → Market ticks (continuous)
Server → Position updates (on changes)
Server → Account updates (on balance changes)
Server → Heartbeat (every 30s)
```

### 5. Disconnection
```
Client → Close connection
Server → Clean up resources
```

**Reconnection Logic**
```javascript
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;

function connect() {
    const ws = new WebSocket('ws://localhost:7999/ws');

    ws.onclose = function() {
        if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
            console.log(`Reconnecting in ${delay}ms...`);
            setTimeout(connect, delay);
        }
    };

    ws.onopen = function() {
        reconnectAttempts = 0;
        console.log('Connected');
    };
}
```

## Performance Characteristics

### Latency
- **Tick Latency**: < 5ms (server processing)
- **Network Latency**: Depends on client location
- **Total Latency**: < 50ms (typical)

### Throughput
- **Ticks per Second**: Up to 10,000
- **Concurrent Clients**: 1,000+ (per instance)
- **Message Size**: 150-300 bytes (average)

### Buffering
- **Server Buffer**: 2,048 messages per client
- **Client Buffer**: Recommended 1,024 messages
- **Overflow Behavior**: Drop oldest messages

## Best Practices

### 1. Handle Reconnection
Always implement exponential backoff for reconnection:
```javascript
const backoff = Math.min(1000 * Math.pow(2, attempts), 30000);
```

### 2. Buffer Management
Process messages quickly to avoid buffer overflow:
```javascript
ws.onmessage = function(event) {
    // Process immediately or queue for async processing
    requestAnimationFrame(() => processMessage(event.data));
};
```

### 3. Subscribe Selectively
Only subscribe to symbols you need:
```javascript
// Good
ws.send(JSON.stringify({
    type: "subscribe",
    symbols: ["EURUSD", "GBPUSD"]  // Only what you need
}));

// Bad
ws.send(JSON.stringify({
    type: "subscribe",
    symbols: getAllSymbols()  // 200+ symbols = high bandwidth
}));
```

### 4. Heartbeat Monitoring
Implement client-side heartbeat timeout:
```javascript
let lastHeartbeat = Date.now();

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.type === 'heartbeat') {
        lastHeartbeat = Date.now();
    }
};

setInterval(() => {
    if (Date.now() - lastHeartbeat > 60000) {
        console.error('Heartbeat timeout, reconnecting...');
        ws.close();
    }
}, 10000);
```

### 5. Error Handling
Handle all error scenarios:
```javascript
ws.onerror = function(error) {
    console.error('WebSocket error:', error);
    // Log to monitoring service
    logError(error);
};

ws.onclose = function(event) {
    console.log('Closed:', event.code, event.reason);
    if (event.code === 1006) {
        // Abnormal closure, reconnect
        reconnect();
    }
};
```

## Security Considerations

### 1. Authentication (Future)
JWT token sent after connection:
```json
{
  "type": "auth",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 2. Account Isolation
Each client only receives data for their account:
- Own positions
- Own account balance
- Market data (public)

### 3. Rate Limiting (Future)
- Max 100 messages per second per client
- Automatic throttling on excess

### 4. Message Validation
All client messages validated:
- JSON schema validation
- Symbol existence check
- Permission verification

## Testing

### WebSocket Testing Tools

**wscat (Command Line)**
```bash
npm install -g wscat
wscat -c ws://localhost:7999/ws
```

**Postman**
1. Create new WebSocket Request
2. Enter: ws://localhost:7999/ws
3. Click Connect
4. Send JSON messages

**Browser Console**
```javascript
const ws = new WebSocket('ws://localhost:7999/ws');
ws.onmessage = (e) => console.log(JSON.parse(e.data));
ws.send(JSON.stringify({type: "subscribe", symbols: ["EURUSD"]}));
```

### Load Testing

**Using websocat**
```bash
# Install websocat
brew install websocat

# Connect multiple clients
for i in {1..100}; do
  websocat ws://localhost:7999/ws &
done
```

**Using Python**
```python
import asyncio
import websockets

async def client(id):
    async with websockets.connect('ws://localhost:7999/ws') as ws:
        while True:
            msg = await ws.recv()
            print(f"Client {id}: {msg}")

async def main():
    await asyncio.gather(*[client(i) for i in range(100)])

asyncio.run(main())
```

## Troubleshooting

### Connection Refused
```
Error: WebSocket connection failed: Connection refused
```
**Solution**: Verify server is running on correct port

### Messages Not Received
```
Connected but no ticks received
```
**Solution**: Check LP connections, verify symbols are subscribed

### Buffer Overflow
```
Warning: Message buffer full, dropping messages
```
**Solution**: Process messages faster, reduce subscriptions

### Heartbeat Timeout
```
No heartbeat received for 60 seconds
```
**Solution**: Check network connectivity, reconnect

## See Also

- [API Documentation](../api/openapi.yaml)
- [Market Data Guide](../concepts/market-data.md)
- [System Architecture](system-overview.md)
