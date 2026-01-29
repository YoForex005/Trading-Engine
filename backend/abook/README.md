# A-Book Execution System

Complete production-grade A-Book execution system for routing client orders to multiple Liquidity Providers via FIX 4.4 protocol.

## Architecture Overview

```
Client Order → API → ExecutionEngine → SmartOrderRouter → FIX Gateway → LP
                ↓                           ↓
            RiskManager              Quote Aggregation
                ↓                           ↓
         Pre-Trade Checks           Best Price Selection
```

## Components

### 1. ExecutionEngine (`engine.go`)

Core execution engine that manages order lifecycle from placement to fill.

**Key Features:**
- Order validation and pre-trade risk checks
- LP selection via Smart Order Router
- FIX protocol integration for order execution
- Execution report processing and position management
- Real-time execution quality metrics (TCA)
- Callbacks for fills, rejections, and updates

**Order Flow:**
1. Validate order parameters
2. Run pre-trade risk checks
3. Select best LP via SOR
4. Route to LP via FIX
5. Monitor execution reports
6. Create positions on fills
7. Update metrics and notify clients

### 2. SmartOrderRouter (`sor.go`)

Multi-LP quote aggregation and smart order routing with health-based failover.

**Key Features:**
- Real-time quote aggregation from all connected LPs
- Best bid/ask calculation across LPs
- LP health scoring (fill rate, slippage, latency)
- Automatic failover to healthy LPs
- Anti-gaming logic (no last-look abuse)

**Health Score Components:**
- Fill Rate: 40%
- Slippage: 30% (lower is better)
- Latency: 20% (lower is better)
- Reject Rate: 10% (lower is better)

### 3. RiskManager (`risk.go`)

Comprehensive pre-trade and post-trade risk management.

**Risk Checks:**
- Position size limits (per order, per symbol, total)
- Exposure limits (per symbol, total notional)
- Daily trade limits
- Daily loss limits with kill switch
- Margin level checks
- Trading hours enforcement
- Volatility circuit breakers

**Kill Switch:**
Automatically activated when:
- Daily loss limit exceeded
- Manual activation by admin

### 4. FIX Integration (`../fix/orders.go`)

FIX 4.4 protocol integration for LP communication.

**Supported Message Types:**
- NewOrderSingle (35=D) - Place order
- OrderCancelRequest (35=F) - Cancel order
- OrderCancelReplaceRequest (35=G) - Modify order
- ExecutionReport (35=8) - Fill/reject notifications
- OrderStatusRequest (35=H) - Query order status

**Message Flow:**
```
Application → NewOrderSingle → LP
LP → ExecutionReport(NEW) → Application
LP → ExecutionReport(FILL) → Application
```

## API Endpoints

### Place Order
```http
POST /abook/orders
Content-Type: application/json

{
  "accountId": "demo_001",
  "symbol": "EURUSD",
  "side": "BUY",
  "type": "MARKET",
  "volume": 1.0,
  "price": 0,
  "sl": 1.0850,
  "tp": 1.0950
}
```

Response:
```json
{
  "success": true,
  "order": {
    "id": "uuid",
    "clientOrderId": "uuid",
    "symbol": "EURUSD",
    "side": "BUY",
    "type": "MARKET",
    "volume": 1.0,
    "status": "SENT",
    "selectedLP": "YOFX1",
    "createdAt": "2026-01-18T10:00:00Z"
  }
}
```

### Cancel Order
```http
POST /abook/orders/cancel
Content-Type: application/json

{
  "orderId": "uuid"
}
```

### Close Position
```http
POST /abook/positions/close
Content-Type: application/json

{
  "positionId": "uuid"
}
```

### Get Positions
```http
GET /abook/positions?accountId=demo_001
```

### Get Execution Metrics
```http
GET /abook/metrics
```

Response:
```json
{
  "totalOrders": 1000,
  "filledOrders": 950,
  "rejectedOrders": 50,
  "partialFills": 10,
  "avgSlippage": 0.00005,
  "avgLatency": "125ms",
  "fillRateByLP": {
    "YOFX1": 0.95,
    "LMAX": 0.98
  },
  "slippageByLP": {
    "YOFX1": 0.00003,
    "LMAX": 0.00002
  }
}
```

## Liquidity Provider Integration

### Supported LPs

1. **YOFx (via T4B FIX server)**
   - Session ID: YOFX1
   - Host: 23.106.238.138:12336
   - Protocol: FIX 4.4
   - Supported: OANDA, Binance, other LPs

2. **LMAX Exchange**
   - Session ID: LMAX_PROD
   - Host: fix.lmax.com:443
   - Protocol: FIX 4.4
   - SSL: Yes

3. **Additional LPs** (configured via FIX Gateway)
   - Currenex
   - Integral
   - EBS

### Adding New LP

1. Register FIX session in `fix.NewFIXGateway()`:
```go
"NEW_LP": {
    ID:           "NEW_LP",
    Name:         "New LP Name",
    Host:         "lp.example.com",
    Port:         443,
    SenderCompID: "YOUR_ID",
    TargetCompID: "LP_ID",
    BeginString:  "FIX.4.4",
    SSL:          true,
}
```

2. Map LP ID in SOR:
```go
func (s *SmartOrderRouter) mapLPToSession(lpID string) string {
    mapping := map[string]string{
        "new_lp": "NEW_LP",
    }
    return mapping[lpID]
}
```

## Configuration

### Account Risk Limits
```go
limits := &abook.AccountLimits{
    MaxPositionSize:       100.0,
    MaxTotalExposure:      1000000.0,
    MaxPositionsPerSymbol: 10,
    MaxTotalPositions:     50,
    MaxDailyLoss:          10000.0,
    MaxDailyTrades:        100,
    EnableKillSwitch:      true,
}
riskManager.SetAccountLimits("accountID", limits)
```

### Symbol Limits
```go
symbolLimits := &abook.SymbolLimits{
    Symbol:           "EURUSD",
    MaxPositionSize:  50.0,
    MaxTotalExposure: 500000.0,
    CircuitBreaker: &abook.CircuitBreaker{
        Enabled:            true,
        PriceChangePercent: 2.0,
        TimeWindow:         5 * time.Minute,
        CooldownPeriod:     30 * time.Minute,
    },
}
riskManager.SetSymbolLimits("EURUSD", symbolLimits)
```

## Performance Targets

- **Order Routing Latency:** < 10ms
- **Fill Rate:** > 95%
- **Average Slippage:** < 0.5 pips
- **LP Connection Recovery:** < 5 seconds

## Error Handling

### Common Errors

1. **"No LP available"**
   - No healthy LP found for symbol
   - Check LP connection status
   - Check quote availability

2. **"Risk check failed: insufficient margin"**
   - Account does not have enough free margin
   - Reduce position size or close positions

3. **"Daily loss limit exceeded"**
   - Kill switch activated
   - Admin must deactivate manually

4. **"Trading hours: not allowed"**
   - Symbol trading hours restriction
   - Check symbol configuration

### Recovery Procedures

1. **LP Disconnection:**
   - Automatic reconnection with exponential backoff
   - Failover to alternative LP
   - Orders in flight reconciled on reconnect

2. **Duplicate Fills:**
   - Sequence number tracking prevents duplicates
   - Message store for gap fill requests

3. **Order Timeout:**
   - Automatic order status request after 30 seconds
   - Manual cancel if no response after 60 seconds

## Monitoring

### Key Metrics

1. **Execution Quality:**
   - Fill rate per LP
   - Average slippage per LP
   - Average latency per LP

2. **Risk Metrics:**
   - Total exposure per account
   - Daily P&L per account
   - Position count per symbol

3. **System Health:**
   - LP connection status
   - FIX session health
   - Quote staleness

### Alerts

Configure alerts for:
- Fill rate < 90%
- Slippage > 1 pip
- Latency > 500ms
- LP disconnection
- Kill switch activation
- Circuit breaker trigger

## Testing

### Unit Tests
```bash
go test ./backend/abook/...
```

### Integration Tests
```bash
go test ./backend/abook/... -tags=integration
```

### Load Tests
```bash
go test ./backend/abook/... -bench=. -benchmem
```

## Security

1. **FIX Authentication:**
   - Username/password credentials
   - Sequence number validation
   - Message signature (if supported by LP)

2. **Pre-Trade Validation:**
   - Symbol whitelist
   - Account-level limits
   - Real-time margin checks

3. **Data Encryption:**
   - SSL/TLS for LP connections
   - Secure credential storage

## Troubleshooting

### Debug Mode
```bash
export FIX_DEBUG=true
export ABOOK_DEBUG=true
./backend/server
```

### Logs
- Execution logs: `/var/log/abook/execution.log`
- FIX logs: `/var/log/fix/messages.log`
- Risk logs: `/var/log/abook/risk.log`

### Common Issues

**Q: Orders stuck in SENT status**
A: Check FIX session connectivity and LP acknowledgment

**Q: High slippage**
A: Review LP quote quality and market volatility

**Q: Frequent rejections**
A: Check risk limits and margin availability

## Production Checklist

- [ ] Configure all LP FIX sessions
- [ ] Set account risk limits
- [ ] Configure symbol trading hours
- [ ] Enable circuit breakers for volatile pairs
- [ ] Set up monitoring and alerts
- [ ] Configure backup LPs for failover
- [ ] Test kill switch functionality
- [ ] Verify execution quality metrics
- [ ] Set up daily reconciliation
- [ ] Document incident response procedures

## Support

For issues or questions:
- Email: support@rtx-trading.com
- Slack: #abook-execution
- Docs: https://docs.rtx-trading.com/abook
