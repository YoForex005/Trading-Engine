# Analytics WebSocket Server

Production-ready WebSocket server for real-time analytics updates with authentication, batching, and backpressure handling.

## Features

- ✅ **JWT Authentication** - Validates tokens on connection
- ✅ **Dynamic Subscriptions** - Subscribe/unsubscribe to specific channels
- ✅ **Message Batching** - Batches up to 50 updates every 16ms (60 FPS)
- ✅ **Backpressure Handling** - Drops old messages if client falls behind
- ✅ **Rate Limiting** - 1000 messages/second per client
- ✅ **Graceful Disconnection** - Proper cleanup on client disconnect
- ✅ **Ping/Pong** - Automatic connection health checks
- ✅ **Non-blocking Broadcasting** - Fast clients don't wait for slow ones

## Channels

| Channel | Description |
|---------|-------------|
| `routing-metrics` | Real-time routing decisions (A-Book vs B-Book) |
| `lp-performance` | LP performance metrics and status |
| `exposure-updates` | Net exposure changes and risk levels |
| `alerts` | Real-time system alerts |

## Quick Start

### 1. Initialize in main.go

```go
import (
    "github.com/epic1st/rtx/backend/internal/api/websocket"
)

func main() {
    // ... existing setup ...

    // Initialize analytics WebSocket hub
    analyticsHub := websocket.InitializeAnalyticsHub(authService)
    websocket.RegisterAnalyticsRoutes(analyticsHub, nil) // Uses http.DefaultServeMux

    // ... rest of server setup ...
}
```

### 2. Client Connection (JavaScript)

```javascript
// Get JWT token from login
const token = "your-jwt-token";

// Connect to WebSocket
const ws = new WebSocket(`ws://localhost:7999/ws/analytics?token=${token}`);

ws.onopen = () => {
    console.log('Connected to analytics WebSocket');

    // Subscribe to channels
    ws.send(JSON.stringify({
        action: 'subscribe',
        channels: ['routing-metrics', 'lp-performance', 'exposure-updates', 'alerts']
    }));
};

ws.onmessage = (event) => {
    const messages = JSON.parse(event.data); // Array of messages (batched)

    messages.forEach(msg => {
        console.log(`[${msg.type}] ${msg.timestamp}:`, msg.data);

        switch(msg.type) {
            case 'routing-decision':
                handleRoutingMetrics(msg.data);
                break;
            case 'lp-metrics':
                handleLPPerformance(msg.data);
                break;
            case 'exposure-change':
                handleExposureUpdate(msg.data);
                break;
            case 'alert':
                handleAlert(msg.data);
                break;
        }
    });
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('WebSocket connection closed');
};
```

### 3. Subscribe/Unsubscribe

```javascript
// Subscribe to additional channels
ws.send(JSON.stringify({
    action: 'subscribe',
    channels: ['alerts']
}));

// Unsubscribe from channels
ws.send(JSON.stringify({
    action: 'unsubscribe',
    channels: ['routing-metrics']
}));
```

## Publishing Updates

### From Routing Engine

```go
import "github.com/epic1st/rtx/backend/internal/api/websocket"

// After routing decision
websocket.NotifyRoutingDecision(
    "EURUSD",           // symbol
    "BUY",              // side
    10000.0,            // volume
    "ABOOK",            // decision
    "OANDA",            // LP selected
    45,                 // execution time (ms)
    0.00015,            // spread
    0.00002,            // slippage
)
```

### From LP Manager

```go
// When LP status changes
websocket.NotifyLPStatusChange(
    "OANDA",            // LP name
    "connected",        // status
    0.00015,            // avg spread
    0.98,               // execution quality (0-1)
    25,                 // latency (ms)
    500,                // quotes per second
    0.02,               // reject rate
    99.9,               // uptime percentage
)
```

### From Exposure Monitor

```go
// When exposure changes
bySymbol := map[string]float64{
    "EURUSD": 50000.0,
    "GBPUSD": 30000.0,
}
byLP := map[string]float64{
    "OANDA": 80000.0,
}

websocket.NotifyExposureChange(
    80000.0,            // total exposure
    80000.0,            // net exposure
    100000.0,           // exposure limit
    bySymbol,           // exposure by symbol
    byLP,               // exposure by LP
)
```

### Emit Alerts

```go
// Critical alert
websocket.NotifyAlert(
    "critical",                          // severity: info, warning, error, critical
    "exposure",                          // category: exposure, lp, routing, system
    "Exposure Limit Exceeded",           // title
    "Net exposure is at 95% of limit",   // message
    "ExposureMonitor",                   // source
    []string{                            // action items
        "Review open positions",
        "Consider hedging large positions",
    },
)

// LP disconnect warning
websocket.NotifyAlert(
    "warning",
    "lp",
    "LP Connection Lost",
    "Connection to OANDA LP was lost. Attempting reconnection.",
    "LPManager",
    []string{"Check LP credentials", "Verify network connectivity"},
)
```

## Message Format

All messages follow this structure:

```json
{
    "type": "routing-decision",
    "timestamp": "2026-01-19T10:30:45.123456Z",
    "data": {
        // Type-specific data
    }
}
```

### Routing Metrics

```json
{
    "type": "routing-decision",
    "timestamp": "2026-01-19T10:30:45.123456Z",
    "data": {
        "symbol": "EURUSD",
        "side": "BUY",
        "volume": 10000.0,
        "routingDecision": "ABOOK",
        "lpSelected": "OANDA",
        "executionTimeMs": 45,
        "spread": 0.00015,
        "slippage": 0.00002
    }
}
```

### LP Performance

```json
{
    "type": "lp-metrics",
    "timestamp": "2026-01-19T10:30:45.123456Z",
    "data": {
        "lpName": "OANDA",
        "status": "connected",
        "avgSpread": 0.00015,
        "executionQuality": 0.98,
        "latencyMs": 25,
        "quotesPerSecond": 500,
        "rejectRate": 0.02,
        "uptime": 99.9
    }
}
```

### Exposure Updates

```json
{
    "type": "exposure-change",
    "timestamp": "2026-01-19T10:30:45.123456Z",
    "data": {
        "totalExposure": 80000.0,
        "netExposure": 80000.0,
        "exposureLimit": 100000.0,
        "utilizationPct": 80.0,
        "riskLevel": "medium",
        "bySymbol": {
            "EURUSD": 50000.0,
            "GBPUSD": 30000.0
        },
        "byLP": {
            "OANDA": 80000.0
        }
    }
}
```

### Alerts

```json
{
    "type": "alert",
    "timestamp": "2026-01-19T10:30:45.123456Z",
    "data": {
        "id": "20260119-103045-abc123",
        "severity": "critical",
        "category": "exposure",
        "title": "Exposure Limit Exceeded",
        "message": "Net exposure is at 95% of limit",
        "source": "ExposureMonitor",
        "actionItems": [
            "Review open positions",
            "Consider hedging large positions"
        ]
    }
}
```

## Configuration

### Constants (in analytics.go)

```go
const (
    maxBatchSize     = 50                      // Messages per batch
    batchInterval    = 16 * time.Millisecond   // 60 FPS
    maxMessagesPerSecond = 1000                // Rate limit per client
    maxQueuedBatches = 10                      // Backpressure threshold
    writeWait        = 10 * time.Second        // Write timeout
    pongWait         = 60 * time.Second        // Read timeout
    pingPeriod       = 54 * time.Second        // Ping interval
)
```

### Tuning for High Throughput

For high-frequency updates (>1000 updates/sec):

1. Increase `maxBatchSize` to 100-200
2. Reduce `batchInterval` to 8ms (120 FPS)
3. Increase channel buffer sizes in client creation

### Tuning for Low Latency

For real-time alerts (<50ms latency):

1. Reduce `maxBatchSize` to 10-20
2. Reduce `batchInterval` to 5ms
3. Increase `maxMessagesPerSecond` rate limit

## Performance Characteristics

- **Batching**: Up to 50 messages sent every 16ms (3000 msgs/sec max per client)
- **Backpressure**: Drops old messages if client falls behind
- **Rate Limiting**: 1000 messages/second per client
- **Broadcasting**: Non-blocking - fast clients don't wait for slow ones
- **Memory**: ~1KB per client + message buffers

## Integration Checklist

- [ ] Initialize analytics hub in main.go
- [ ] Register WebSocket routes
- [ ] Add routing decision notifications
- [ ] Add LP status change notifications
- [ ] Add exposure update notifications
- [ ] Add alert notifications
- [ ] Test client connection with JWT
- [ ] Test subscription/unsubscription
- [ ] Test message batching (verify 60 FPS)
- [ ] Test backpressure (simulate slow client)
- [ ] Test rate limiting
- [ ] Test graceful disconnect

## Monitoring

The hub logs:
- Client connections/disconnections
- Subscription changes
- Buffer full warnings (backpressure)
- Rate limit violations
- Authentication failures

Example logs:
```
[AnalyticsHub] Started
[AnalyticsHub] Client connected (user: admin, account: admin). Total: 1
[AnalyticsWS] User admin subscribed to channel: routing-metrics
[AnalyticsHub] Client admin buffer full, dropping message
[AnalyticsWS] Rate limit exceeded for user admin
[AnalyticsHub] Client disconnected (user: admin). Total: 0
```

## Security Considerations

1. **JWT Validation**: All connections require valid JWT token
2. **Origin Checking**: Configure `CheckOrigin` for production
3. **Rate Limiting**: Prevents abuse (1000 msgs/sec per client)
4. **Message Size Limits**: Incoming messages limited to 512 bytes
5. **Timeout Handling**: Automatic disconnect for inactive clients

## Testing

See `analytics_test.go` for unit tests covering:
- Connection authentication
- Subscription management
- Message batching
- Backpressure handling
- Rate limiting
- Graceful disconnection

## Architecture

```
                    ┌─────────────────┐
                    │  AnalyticsHub   │
                    │  (Broadcaster)  │
                    └────────┬────────┘
                             │
                ┌────────────┼────────────┐
                │            │            │
         ┌──────▼─────┐ ┌───▼──────┐ ┌──▼───────┐
         │  Client 1  │ │ Client 2 │ │ Client N │
         │            │ │          │ │          │
         │ Batching   │ │ Batching │ │ Batching │
         │ 60 FPS     │ │ 60 FPS   │ │ 60 FPS   │
         └────────────┘ └──────────┘ └──────────┘
              ▲              ▲            ▲
              │              │            │
         Subscribe      Subscribe    Subscribe
         routing-metrics  lp-perf    exposure
```

## License

Part of RTX Trading Engine - Internal Use
