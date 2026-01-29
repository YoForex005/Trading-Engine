# Execution Models: A-Book, B-Book, and C-Book

## Overview

The RTX Trading Engine supports multiple execution models that determine how client orders are processed. Understanding these models is crucial for risk management, pricing strategies, and regulatory compliance.

## Execution Model Comparison

| Feature | A-Book | B-Book | C-Book (Hybrid) |
|---------|--------|--------|-----------------|
| **Counterparty** | Liquidity Provider | Broker (Internal) | Mixed |
| **Market Risk** | Client bears | Broker bears | Shared |
| **Execution Speed** | LP-dependent (100-500ms) | Instant (<10ms) | Mixed |
| **Slippage** | Market conditions | Controlled | Mixed |
| **Spread** | LP spread + markup | Broker-defined | Dynamic |
| **Profit Model** | Commission + markup | Client losses | Optimized |
| **Regulatory** | STP/DMA compliant | Market making | Requires disclosure |

## A-Book (Agency Model / STP)

### Description
In A-Book execution, the broker acts as an intermediary, routing all client orders directly to external liquidity providers. The broker does not take the opposite side of trades.

### Flow Diagram
```
┌──────────┐         ┌──────────┐         ┌─────────────┐
│  Client  │────────▶│  Broker  │────────▶│ Liquidity   │
│  Order   │         │ (Router) │         │ Provider    │
└──────────┘         └──────────┘         └─────────────┘
                          │                      │
                          │                      │
                     Commission             Market Price
                       Earned                 Execution
```

### How It Works
1. Client submits order via API
2. Broker validates order (volume, margin, symbol)
3. Order is routed to configured liquidity provider(s)
4. LP executes at market price
5. Execution confirmation sent back to broker
6. Client position created with LP execution price
7. Broker earns commission or spread markup

### Revenue Model
- **Commission**: Fixed fee per trade or lot
- **Spread Markup**: Add pips to LP spread
- **Example**: LP spread 1.0 pips, broker adds 0.5 pips = 1.5 pips total

### Advantages
- **No Market Risk**: Broker doesn't bear directional risk
- **Regulatory Compliance**: Meets STP/DMA requirements
- **Scalability**: Can handle unlimited volume
- **Client Trust**: Transparent execution model

### Disadvantages
- **LP Dependency**: Execution quality depends on LP
- **Latency**: Network latency to LP (100-500ms)
- **LP Costs**: Must maintain LP relationships
- **Slippage**: Market slippage passed to clients
- **Limited Control**: Cannot control execution price

### Configuration in RTX
```json
{
  "executionMode": "ABOOK",
  "lpRouter": {
    "primary": "OANDA",
    "failover": ["YoFx", "Binance"],
    "routingLogic": "BEST_PRICE"
  },
  "markup": {
    "type": "SPREAD",
    "pips": 0.5
  }
}
```

### Use Cases
- Large professional clients
- High-volume traders
- Regulatory compliance requirements
- Low-risk broker model

## B-Book (Market Making / Internal Execution)

### Description
In B-Book execution, the broker acts as the counterparty to all client trades. Orders are executed internally using the broker's own balance sheet and price feed.

### Flow Diagram
```
┌──────────┐         ┌──────────────────┐
│  Client  │────────▶│  Broker Engine   │
│  Order   │         │  (Counterparty)  │
└──────────┘         └──────────────────┘
                              │
                              ▼
                     Internal Matching
                     Balance Updated
                     Position Created
```

### How It Works
1. Client submits order via API
2. Broker validates order (volume, margin, symbol)
3. Order executed immediately at current bid/ask
4. Position created in broker's internal ledger
5. Broker takes opposite position (assumes market risk)
6. Client P&L inversely affects broker P&L
7. Broker may hedge net exposure externally (optional)

### Revenue Model
- **Client Losses**: Broker profits from losing traders
- **Spread**: Broker defines spread and keeps it all
- **Hedging**: May hedge net risk to LPs
- **Example**: If client loses $100, broker gains $100 (before hedging)

### Advantages
- **Instant Execution**: No LP latency (<10ms)
- **Full Control**: Broker controls pricing and execution
- **Higher Profit Potential**: Keep spread + client losses
- **Flexibility**: Can offer custom spreads/conditions
- **No LP Costs**: No LP fees or commissions

### Disadvantages
- **Market Risk**: Broker bears full directional risk
- **Capital Requirements**: Requires large balance sheet
- **Concentration Risk**: Large winning clients = big losses
- **Regulatory Scrutiny**: Some jurisdictions require disclosure
- **Conflicts of Interest**: Broker profits when clients lose

### Risk Management Strategies
1. **Net Position Hedging**: Hedge aggregate exposure to LP
2. **Client Segmentation**: B-Book small/losing, A-Book winners
3. **Dynamic Spreads**: Widen spreads during high volatility
4. **Volume Limits**: Cap maximum position sizes
5. **Stop Loss Enforcement**: Ensure SL/TP execution

### Configuration in RTX
```json
{
  "executionMode": "BBOOK",
  "spread": {
    "EURUSD": 1.5,
    "GBPUSD": 2.0,
    "BTCUSD": 50.0
  },
  "hedging": {
    "enabled": true,
    "threshold": 100000,
    "provider": "OANDA"
  },
  "riskLimits": {
    "maxPositionSize": 100.0,
    "marginCallLevel": 50
  }
}
```

### Use Cases
- Small retail clients
- Demo accounts
- Traders with consistent losing record
- High-frequency small trades
- Bonus/promotional accounts

## C-Book (Hybrid Model)

### Description
C-Book combines A-Book and B-Book execution based on client profitability, risk profile, and trading behavior. This is the most sophisticated and profitable model when executed correctly.

### Flow Diagram
```
┌──────────┐         ┌─────────────────┐
│  Client  │────────▶│  Risk Engine    │
│  Order   │         │  (Classifier)   │
└──────────┘         └────────┬────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
              ┌─────▼──────┐     ┌─────▼──────┐
              │  B-Book    │     │  A-Book    │
              │ (Internal) │     │ (LP Route) │
              └────────────┘     └────────────┘
```

### How It Works
1. Client submits order
2. Risk engine analyzes client profile:
   - Trading history (win/loss ratio)
   - Account size and activity
   - Trade size and frequency
   - Volatility exposure
   - Risk score
3. Algorithm decides routing:
   - **Low Risk** → B-Book (internal execution)
   - **High Risk** → A-Book (LP passthrough)
4. Order executed via selected model
5. Client classification updated continuously

### Classification Criteria

**B-Book Candidates (Low Risk)**
- Win rate < 40%
- Average trade duration < 1 hour
- Small position sizes (< 1 lot)
- Demo accounts
- Bonus accounts
- Inactive traders (< 10 trades/month)

**A-Book Candidates (High Risk)**
- Win rate > 60%
- Large position sizes (> 10 lots)
- Professional/institutional traders
- Consistent profitability
- Algorithmic traders
- News/scalping strategies

### Dynamic Reclassification
```
Client Performance Tracking:
  ├─ Daily P&L monitoring
  ├─ Win/loss ratio calculation
  ├─ Volume analysis
  ├─ Strategy detection (scalping, swing, etc.)
  └─ Risk score update

If Risk Score Changes:
  ├─ Migrate B-Book → A-Book (winning too much)
  └─ Migrate A-Book → B-Book (consistent losses)
```

### Revenue Optimization
```
Total Revenue = B-Book Profits + A-Book Commissions

B-Book Profits:
  ├─ Spread income
  ├─ Client losses (unhedged portion)
  └─ Hedging gains (if net position profitable)

A-Book Commissions:
  ├─ Spread markup
  └─ Commission per trade
```

### Configuration Example
```json
{
  "executionMode": "CBOOK",
  "classification": {
    "algorithm": "MACHINE_LEARNING",
    "reclassificationInterval": "24h",
    "thresholds": {
      "winRateHigh": 0.60,
      "winRateLow": 0.40,
      "volumeThreshold": 10.0,
      "profitThreshold": 10000
    }
  },
  "routing": {
    "bbook": {
      "enabled": true,
      "hedgeThreshold": 100000
    },
    "abook": {
      "enabled": true,
      "providers": ["OANDA", "YoFx"]
    }
  }
}
```

### Advantages
- **Optimized Profitability**: Best of both models
- **Risk Mitigation**: Winning clients hedged externally
- **Flexibility**: Adapt to client behavior
- **Scalability**: Handle diverse client base

### Disadvantages
- **Complexity**: Requires sophisticated risk engine
- **Regulatory Risk**: Must disclose conflict of interest
- **Technology Costs**: Advanced algorithms and monitoring
- **Operational Risk**: Misclassification can be costly

## Regulatory Considerations

### Disclosure Requirements
- **EU/ESMA**: Must disclose execution policy
- **US/CFTC**: STP brokers must route to LPs
- **AU/ASIC**: Market making model disclosure required
- **UK/FCA**: Best execution obligation

### Best Execution Obligation
Regardless of model, brokers must:
1. Obtain best possible execution for clients
2. Monitor execution quality
3. Document execution policy
4. Provide execution reports
5. Review execution annually

## Hedging Strategies for B-Book

### 1. Net Position Hedging
```
Client Positions:
  ├─ Client A: +10 lots EURUSD (Buy)
  ├─ Client B: +5 lots EURUSD (Buy)
  └─ Client C: -8 lots EURUSD (Sell)

Broker Net Exposure: +7 lots EURUSD (Buy)

Hedging Action:
  └─ Open -7 lots EURUSD with LP (Sell)
  └─ Result: Zero net market exposure
```

### 2. Threshold-Based Hedging
```
if abs(netExposure) > threshold:
    hedge(netExposure)
else:
    assume_risk(netExposure)
```

### 3. Dynamic Hedging
```
Hedge Ratio = f(volatility, net_exposure, client_behavior)

High Volatility:  Hedge 100%
Medium Volatility: Hedge 70%
Low Volatility: Hedge 30%
```

## Implementation in RTX Trading Engine

### Current Support
- ✅ **B-Book**: Fully implemented (`internal/core/engine.go`)
- ✅ **A-Book**: Basic routing (`lpmanager/`, `fix/gateway.go`)
- ⏳ **C-Book**: Planned (requires ML risk classifier)

### Toggle Execution Mode
```bash
# Via API
curl -X POST http://localhost:7999/admin/execution-mode \
  -H "Content-Type: application/json" \
  -d '{"mode": "BBOOK"}'

# Response
{
  "success": true,
  "oldMode": "ABOOK",
  "newMode": "BBOOK",
  "message": "Execution mode updated"
}
```

### Code Example: B-Book Execution
```go
// B-Book market order execution
position, err := engine.ExecuteMarketOrder(
    accountID,  // int64
    "EURUSD",   // symbol
    "BUY",      // side
    0.1,        // volume (lots)
    1.0850,     // stop loss
    1.0950,     // take profit
)

if err != nil {
    return err
}

// Position created instantly, broker is counterparty
log.Printf("Position opened: #%d @ %.5f", position.ID, position.OpenPrice)
```

## Conclusion

Each execution model has distinct advantages and risks:
- **A-Book**: Low risk, regulatory compliant, but lower profit margins
- **B-Book**: High profit potential, instant execution, but significant market risk
- **C-Book**: Optimal profitability, but requires sophisticated technology

The choice depends on:
- Broker's risk appetite
- Regulatory environment
- Technology capabilities
- Target client base
- Capital availability

RTX Trading Engine provides the flexibility to implement any model, with configurable parameters for risk management and revenue optimization.
