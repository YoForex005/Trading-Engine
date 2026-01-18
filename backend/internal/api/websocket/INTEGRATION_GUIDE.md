# Analytics WebSocket Integration Guide

This guide shows how to integrate the analytics WebSocket with existing backend systems.

## Overview

The analytics WebSocket broadcasts real-time updates for:
1. **Routing Metrics** - A-Book vs B-Book routing decisions
2. **LP Performance** - LP connection status and quality metrics
3. **Exposure Updates** - Real-time exposure monitoring
4. **Alerts** - Critical system alerts

## Integration Points

### 1. Routing Engine Integration

In `backend/cbook/engine.go` or wherever routing decisions are made:

```go
import "github.com/epic1st/rtx/backend/internal/api/websocket"

func (e *CBookEngine) RouteOrder(order *Order) (string, string, error) {
    startTime := time.Now()

    // Make routing decision
    decision, lp := e.makeRoutingDecision(order)

    // Execute order
    err := e.executeOrder(order, decision, lp)

    // Notify analytics WebSocket
    execTime := time.Since(startTime).Milliseconds()
    websocket.NotifyRoutingDecision(
        order.Symbol,
        order.Side,
        order.Volume,
        decision,  // "ABOOK" or "BBOOK"
        lp,        // LP name if ABOOK
        execTime,
        order.Spread,
        order.Slippage,
    )

    return decision, lp, err
}
```

### 2. LP Manager Integration

In `backend/lpmanager/manager.go`:

```go
import "github.com/epic1st/rtx/backend/internal/api/websocket"

// Call this when LP connection status changes
func (m *Manager) onLPStatusChange(lpName string, connected bool) {
    status := "disconnected"
    if connected {
        status = "connected"
    }

    // Get LP metrics
    metrics := m.getLPMetrics(lpName)

    websocket.NotifyLPStatusChange(
        lpName,
        status,
        metrics.AvgSpread,
        metrics.ExecutionQuality,
        metrics.LatencyMs,
        metrics.QuotesPerSecond,
        metrics.RejectRate,
        metrics.Uptime,
    )
}

// Call this periodically (e.g., every 5 seconds)
func (m *Manager) publishLPMetrics() {
    for lpName, lp := range m.liquidityProviders {
        metrics := m.calculateMetrics(lp)

        websocket.NotifyLPStatusChange(
            lpName,
            lp.Status,
            metrics.AvgSpread,
            metrics.ExecutionQuality,
            metrics.LatencyMs,
            metrics.QuotesPerSecond,
            metrics.RejectRate,
            metrics.Uptime,
        )
    }
}
```

### 3. Exposure Monitor Integration

Create a new exposure monitor or integrate into existing risk management:

```go
// backend/internal/core/exposure_monitor.go

package core

import (
    "time"
    "github.com/epic1st/rtx/backend/internal/api/websocket"
)

type ExposureMonitor struct {
    engine        *Engine
    exposureLimit float64
    checkInterval time.Duration
    stopChan      chan struct{}
}

func NewExposureMonitor(engine *Engine, limit float64) *ExposureMonitor {
    return &ExposureMonitor{
        engine:        engine,
        exposureLimit: limit,
        checkInterval: 5 * time.Second,
        stopChan:      make(chan struct{}),
    }
}

func (m *ExposureMonitor) Start() {
    go m.monitorLoop()
}

func (m *ExposureMonitor) Stop() {
    close(m.stopChan)
}

func (m *ExposureMonitor) monitorLoop() {
    ticker := time.NewTicker(m.checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            m.checkExposure()
        case <-m.stopChan:
            return
        }
    }
}

func (m *ExposureMonitor) checkExposure() {
    // Calculate current exposure
    totalExposure := 0.0
    bySymbol := make(map[string]float64)
    byLP := make(map[string]float64)

    positions := m.engine.GetAllPositions()
    for _, pos := range positions {
        exposure := pos.Volume * pos.CurrentPrice
        totalExposure += exposure

        bySymbol[pos.Symbol] += exposure
        if pos.LP != "" {
            byLP[pos.LP] += exposure
        }
    }

    // Publish to analytics WebSocket
    websocket.NotifyExposureChange(
        totalExposure,
        totalExposure, // Net exposure (can be calculated differently)
        m.exposureLimit,
        bySymbol,
        byLP,
    )

    // Check if limit exceeded
    utilizationPct := (totalExposure / m.exposureLimit) * 100
    if utilizationPct >= 90 {
        websocket.NotifyAlert(
            "critical",
            "exposure",
            "Exposure Limit Approaching",
            fmt.Sprintf("Current exposure is at %.1f%% of limit", utilizationPct),
            "ExposureMonitor",
            []string{
                "Review open positions",
                "Consider closing high-risk positions",
                "Increase exposure limit if justified",
            },
        )
    }
}
```

Then in `main.go`:

```go
// Initialize exposure monitor
exposureMonitor := core.NewExposureMonitor(bbookEngine, 100000.0)
exposureMonitor.Start()
log.Println("[ExposureMonitor] Started with $100,000 limit")
```

### 4. Alert System Integration

The existing alert system in `backend/internal/alerts` can broadcast to the analytics WebSocket:

```go
// In backend/internal/alerts/engine.go

import "github.com/epic1st/rtx/backend/internal/api/websocket"

func (e *Engine) triggerAlert(alert *Alert) {
    // Existing alert handling...

    // Also broadcast to analytics WebSocket
    actionItems := []string{}
    for _, action := range alert.Actions {
        actionItems = append(actionItems, action.Description)
    }

    websocket.NotifyAlert(
        alert.Severity,
        alert.Category,
        alert.Title,
        alert.Message,
        "AlertEngine",
        actionItems,
    )
}
```

### 5. Example: Complete Integration in main.go

```go
// After initializing all systems
func main() {
    // ... existing initialization ...

    // Initialize Analytics WebSocket Hub
    analyticsHub := websocket.InitializeAnalyticsHub(authService)
    websocket.RegisterAnalyticsRoutes(analyticsHub, nil)
    log.Println("[Analytics] WebSocket hub initialized at /ws/analytics")

    // Initialize Exposure Monitor
    exposureMonitor := core.NewExposureMonitor(bbookEngine, 100000.0)
    exposureMonitor.Start()

    // Start LP metrics publisher
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            lpMgr.PublishMetrics() // Calls websocket.NotifyLPStatusChange internally
        }
    }()

    // ... rest of server setup ...
}
```

## Testing the Integration

### 1. Start the Backend

```bash
cd backend
go run cmd/server/main.go
```

### 2. Connect with WebSocket Client

Use the TypeScript client (see `example-client.ts`):

```typescript
import { AnalyticsWebSocket } from './websocket-client'

const client = new AnalyticsWebSocket(
    {
        url: 'ws://localhost:7999/ws/analytics',
        token: 'your-jwt-token',
        channels: ['routing-metrics', 'lp-performance', 'exposure-updates', 'alerts'],
        debug: true,
    },
    {
        onRoutingMetrics: (data) => {
            console.log(`Routing: ${data.symbol} → ${data.routingDecision}`)
        },
        onLPPerformance: (data) => {
            console.log(`LP ${data.lpName}: ${data.status} (${data.latencyMs}ms)`)
        },
        onExposureUpdate: (data) => {
            console.log(`Exposure: ${data.utilizationPct.toFixed(1)}% (${data.riskLevel})`)
        },
        onAlert: (data) => {
            console.log(`Alert: ${data.severity} - ${data.title}`)
        },
    }
)

client.connect()
```

### 3. Test Each Channel

#### Test Routing Metrics
Place an order through the API:
```bash
curl -X POST http://localhost:7999/api/orders/market \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"symbol":"EURUSD","side":"BUY","volume":10000}'
```

Expected WebSocket message:
```json
{
    "type": "routing-decision",
    "timestamp": "2026-01-19T10:30:45.123Z",
    "data": {
        "symbol": "EURUSD",
        "side": "BUY",
        "volume": 10000,
        "routingDecision": "ABOOK",
        "lpSelected": "OANDA",
        "executionTimeMs": 45
    }
}
```

#### Test LP Performance
The LP manager should publish metrics every 5 seconds automatically.

#### Test Exposure Updates
The exposure monitor publishes every 5 seconds with current positions.

#### Test Alerts
Trigger an alert condition (e.g., exposure > 90%):
```bash
# Place multiple large orders to increase exposure
for i in {1..5}; do
  curl -X POST http://localhost:7999/api/orders/market \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"symbol":"EURUSD","side":"BUY","volume":50000}'
done
```

## Performance Tuning

### High-Frequency Trading (>1000 updates/sec)

Increase batch size and buffer:

```go
// In analytics.go
const (
    maxBatchSize    = 100               // Increase from 50
    batchInterval   = 8 * time.Millisecond  // Decrease from 16ms (120 FPS)
    sendChannelSize = 512               // Increase from 256
)
```

### Low-Latency Alerts (<50ms)

Decrease batch size for faster delivery:

```go
const (
    maxBatchSize  = 10                 // Smaller batches
    batchInterval = 5 * time.Millisecond  // Faster delivery
)
```

### Memory-Constrained Environments

Reduce buffer sizes:

```go
const (
    sendChannelSize = 64    // Reduce from 256
    maxQueuedBatches = 5    // Reduce from 10
)
```

## Monitoring

### Check Connected Clients

Add a status endpoint:

```go
http.HandleFunc("/api/analytics/status", func(w http.ResponseWriter, r *http.Request) {
    hub := websocket.GetAnalyticsHub()
    if hub == nil {
        http.Error(w, "Analytics hub not initialized", http.StatusServiceUnavailable)
        return
    }

    // TODO: Add method to get client count and subscriptions
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "running": hub.IsRunning(),
        "clients": hub.GetClientCount(),
    })
})
```

### Logging

The WebSocket logs:
- Client connections/disconnections with user ID
- Subscription changes per client
- Buffer full warnings (backpressure)
- Rate limit violations
- Broadcast channel buffer warnings

Example logs:
```
[AnalyticsHub] Started
[AnalyticsHub] Client connected (user: admin, account: admin). Total: 1
[AnalyticsWS] User admin subscribed to channel: routing-metrics
[AnalyticsWS] User admin subscribed to channel: alerts
[AnalyticsHub] Client admin buffer full, dropping message
[AnalyticsHub] Client disconnected (user: admin). Total: 0
```

## Security Checklist

- [x] JWT token validation on connection
- [x] Rate limiting (1000 msgs/sec per client)
- [x] Message size limits (512 bytes for incoming)
- [x] Timeout handling (60s pong timeout)
- [ ] Configure `CheckOrigin` for production (currently allows all)
- [ ] Add role-based channel access (admin-only channels)
- [ ] Add IP whitelisting if needed
- [ ] Enable TLS for production (wss://)

## Troubleshooting

### "Unauthorized" on Connection
- Verify JWT token is valid
- Check token is passed in `?token=` query parameter or `Authorization: Bearer` header
- Verify auth service is configured in hub

### No Messages Received
- Verify client subscribed to channels (send subscription message)
- Check backend is actually publishing to those channels
- Verify WebSocket connection is open (`ws.readyState === 1`)

### Connection Drops Frequently
- Check network stability
- Verify ping/pong interval (60s timeout)
- Enable auto-reconnect in client

### High Latency
- Check batch interval (16ms = 60 FPS)
- Reduce batch size for lower latency
- Check network conditions
- Verify server not overloaded

## Next Steps

1. Implement exposure monitor in production
2. Add LP metrics publishing
3. Connect routing engine
4. Add monitoring dashboard
5. Configure CORS/origin checking for production
6. Enable TLS (wss://)
7. Add role-based access control
8. Implement message persistence for critical alerts
