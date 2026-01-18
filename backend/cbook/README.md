# C-Book Hybrid Routing System

A sophisticated A-Book/B-Book routing engine with intelligent client profiling, machine learning predictions, dynamic risk management, and comprehensive compliance tracking.

## Overview

The C-Book system implements intelligent hybrid routing that maximizes broker profitability while managing risk. It automatically classifies clients based on trading behavior and routes orders optimally between A-Book (hedged with LP) and B-Book (internalized).

## Key Features

### 1. Client Profiling Engine

**Automatic Classification:**
- **RETAIL**: High probability losers → Mostly B-Book (90%)
- **SEMI_PRO**: Borderline profitable → Balanced (50/50)
- **PROFESSIONAL**: Consistent winners → Mostly A-Book (80%)
- **TOXIC**: Arbitrageurs, latency traders → A-Book or Reject

**Tracked Metrics:**
- Win rate and P&L history
- Sharpe ratio and max drawdown
- Average trade size and hold time
- Order-to-fill ratio (cancellation rate)
- Instrument concentration
- Time-of-day patterns
- Toxicity score (0-100)

**Toxicity Detection:**
- Win rate > 55% (suspicious)
- High Sharpe ratio (> 2.0)
- Short hold times (< 1 minute)
- High order cancellation rate (> 50%)
- Concentrated instrument trading
- Pattern correlation with known toxic clients

### 2. Dynamic Routing Rules

**Rule-Based Routing:**
- Manual rules with priority system
- Filter by account ID, symbol, volume, classification
- Support for account groups and user roles
- Configurable hedge percentages for partial routing

**Intelligent Auto-Routing:**
```
Retail (Win rate < 48%)         → 90% B-Book, 10% A-Book
Semi-Pro (Win rate 48-52%)      → 50% B-Book, 50% A-Book
Professional (Win rate > 52%)   → 20% B-Book, 80% A-Book
Toxic (Win rate > 55%, high SR) → 100% A-Book or Reject
```

**Volume-Based Overrides:**
- Large orders (≥ 10 lots) → Always A-Book
- Reduces broker exposure to large trades

**Exposure-Based Adjustment:**
- Monitors net exposure per symbol
- Auto-hedge when exposure exceeds threshold
- Dynamic hedge ratio adjustment

**Volatility-Based Routing:**
- High volatility (> 2%) → Increase A-Book percentage
- Reduces risk during market uncertainty

### 3. Risk-Based Exposure Management

**Symbol Exposure Tracking:**
- Real-time net exposure (long - short)
- Gross exposure (total positions)
- Per-symbol limits with auto-hedge triggers

**Configurable Limits:**
```go
ExposureLimit{
    Symbol:           "EURUSD",
    MaxNetExposure:   500,   // 500 lots max
    MaxGrossExposure: 1000,  // 1000 lots total
    AutoHedgeLevel:   300,   // Hedge when > 300 lots
}
```

**Automatic Hedging:**
- Routes to A-Book when exposure exceeds threshold
- Gradual increase in A-Book percentage as exposure approaches limit
- Prevents excessive B-Book risk concentration

### 4. Machine Learning Integration

**Profitability Prediction:**
- Online learning (updates in real-time)
- Predicts client win rate and risk score
- Feature-based scoring with learned weights

**Features Used:**
- Historical win rate
- Sharpe ratio
- Average hold time (log scale)
- Order-to-fill ratio
- Instrument concentration
- Total volume (log scale)
- Toxicity score
- Max drawdown
- Average trade size
- Time consistency

**Model Training:**
- Gradient descent with L2 regularization
- Mini-batch training every 100 samples
- Automatic feature normalization
- Confidence scoring based on sample size

**ML-Enhanced Routing:**
- Overrides rule-based routing when confidence > 70%
- More conservative (safer for broker) decisions
- Continuous improvement from outcomes

### 5. Partial Hedging

**Dynamic Split:**
- Route portion to A-Book, portion to B-Book
- Example: 70% A-Book, 30% B-Book
- Configurable per client, rule, or prediction

**Benefits:**
- Balances risk and profit opportunity
- Smooth transition between full A-Book and B-Book
- Reduces binary decision risk

### 6. Comprehensive Analytics

**Client Performance Tracking:**
- Total P&L breakdown (A-Book vs B-Book)
- Broker net P&L (B-Book profit - A-Book costs)
- Trade counts by routing type
- Win rate, average win/loss, profit factor
- Sharpe ratio, max drawdown, VaR 95%

**Symbol Performance:**
- P&L by instrument
- Volume distribution
- A-Book/B-Book percentages
- Client concentration

**Routing Effectiveness:**
- Classification accuracy (retail, pro, toxic)
- Routing optimality rate
- Cost-benefit analysis (hedging costs vs B-Book profits)
- ROI on hedging decisions

**Top Performers:**
- Ranked by broker profit contribution
- Identify most profitable clients (losers in B-Book)
- Identify risky clients (winners requiring hedging)

### 7. Compliance & Audit Trail

**Complete Audit Logging:**
- Every routing decision logged
- Captures client profile, ML prediction, decision rationale
- Records actual trade outcome
- Evaluates if decision was optimal

**Compliance Checks:**
- Best execution validation
- Excessive B-Book exposure warnings
- Toxic client in B-Book alerts
- Discrimination prevention

**Regulatory Reporting:**
- Best execution reports
- Audit trail export for regulators
- Routing decision justification
- Transparency in decision-making

**Alert System:**
- Real-time compliance alerts (INFO, WARNING, CRITICAL)
- Tracks unresolved issues
- Alert resolution workflow

### 8. Admin Controls

**Manual Overrides:**
- Force specific client to A-Book or B-Book
- Override ML recommendations
- Temporary routing rules

**Configuration:**
- Exposure limits per symbol
- Classification thresholds
- ML model parameters
- Compliance rules

**Real-Time Monitoring:**
- Live exposure dashboard
- Active client classifications
- Routing decision stream
- Alert notifications

## Architecture

```
CBookEngine (Main Orchestrator)
├── ClientProfileEngine (Client Classification)
│   ├── Client metrics tracking
│   ├── Toxicity score calculation
│   └── Automatic classification
│
├── RoutingEngine (Decision Making)
│   ├── Rule-based routing
│   ├── Exposure management
│   └── Decision history
│
├── MLPredictor (Machine Learning)
│   ├── Feature extraction
│   ├── Online learning
│   └── Profitability prediction
│
├── RoutingAnalytics (Performance Tracking)
│   ├── Client performance
│   ├── Symbol performance
│   └── Effectiveness metrics
│
└── ComplianceEngine (Audit & Compliance)
    ├── Audit trail logging
    ├── Compliance checking
    └── Regulatory reporting
```

## API Endpoints

### Client Profiling
```
GET  /api/cbook/profiles                      # Get all profiles
GET  /api/cbook/profiles?classification=RETAIL # Filter by classification
GET  /api/cbook/profile/?accountId=123        # Get specific profile
```

### Routing
```
POST /api/cbook/route                         # Make routing decision
GET  /api/cbook/routing/stats                 # Routing statistics
GET  /api/cbook/routing/rules                 # Get all rules
POST /api/cbook/routing/rules                 # Create rule
PUT  /api/cbook/routing/rules                 # Update rule
DELETE /api/cbook/routing/rules?id=rule_001   # Delete rule
```

### Exposure Management
```
GET  /api/cbook/exposure                      # All exposures
GET  /api/cbook/exposure?symbol=EURUSD        # Symbol exposure
POST /api/cbook/exposure/limits               # Set exposure limit
```

### Analytics
```
GET  /api/cbook/analytics/performance?accountId=123  # Client performance
GET  /api/cbook/analytics/performance?limit=10       # Top performers
GET  /api/cbook/analytics/pnl                        # P&L report
GET  /api/cbook/analytics/effectiveness?period=1W    # Effectiveness metrics
GET  /api/cbook/analytics/recommendations            # Optimization suggestions
```

### Machine Learning
```
GET  /api/cbook/ml/stats                      # Model statistics
GET  /api/cbook/ml/predict?accountId=123      # Get prediction
GET  /api/cbook/ml/export                     # Export model
POST /api/cbook/ml/import                     # Import model
```

### Compliance
```
GET  /api/cbook/compliance/audit?accountId=123&limit=100  # Audit logs
GET  /api/cbook/compliance/alerts?severity=CRITICAL       # Compliance alerts
GET  /api/cbook/compliance/best-execution?start=...       # Best execution report
GET  /api/cbook/compliance/export                         # Export audit trail
```

### Dashboard
```
GET  /api/cbook/dashboard                     # Comprehensive dashboard
GET  /api/cbook/admin/config                  # Get configuration
POST /api/cbook/admin/config                  # Update configuration
```

## Usage Example

### 1. Initialize Engine

```go
import "trading-engine/backend/cbook"

// Create engine
engine := cbook.NewCBookEngine()

// Configure
engine.EnableML(true)
engine.EnableAutoLearning(true)
engine.EnableStrictCompliance(true)
```

### 2. Make Routing Decision

```go
decision, err := engine.RouteOrder(
    accountID:        12345,
    userID:          "user_001",
    username:        "john_doe",
    symbol:          "EURUSD",
    side:            "BUY",
    volume:          2.5,
    currentVolatility: 0.015,
)

// Decision contains:
// - Action: A_BOOK, B_BOOK, PARTIAL_HEDGE, REJECT
// - ABookPercent: 0-100
// - BBookPercent: 0-100
// - TargetLP: "LMAX_PROD"
// - Reason: "Retail trader - 90% B-Book"
// - ToxicityScore: 25.3
```

### 3. Record Trade Outcome

```go
trade := cbook.TradeRecord{
    TradeID:    67890,
    Symbol:     "EURUSD",
    Volume:     2.5,
    OpenPrice:  1.1234,
    ClosePrice: 1.1256,
    PnL:        55.0,
    OpenTime:   time.Now().Add(-1 * time.Hour),
    CloseTime:  time.Now(),
    HoldTime:   3600,
    IsWinner:   true,
}

engine.RecordTrade(accountID, trade)

// This will:
// 1. Update client profile
// 2. Train ML model (if auto-learning enabled)
// 3. Recalculate metrics
```

### 4. Add Manual Rule

```go
rule := &cbook.RoutingRule{
    ID:              "vip_rule_001",
    Priority:        100,
    AccountIDs:      []int64{12345, 67890},
    Action:          cbook.ActionABook,
    TargetLP:        "LMAX_PROD",
    Enabled:         true,
    Description:     "VIP clients always A-Book",
}

engine.AddRoutingRule(rule)
```

### 5. Set Exposure Limits

```go
limit := &cbook.ExposureLimit{
    Symbol:           "EURUSD",
    MaxNetExposure:   500,
    MaxGrossExposure: 1000,
    AutoHedgeLevel:   300,
}

engine.SetExposureLimit("EURUSD", limit)
```

### 6. Get Dashboard Data

```go
dashboard := engine.GetDashboardData()

// Returns:
// - Total decisions
// - Client classifications breakdown
// - Routing statistics
// - ML model stats
// - Compliance stats
// - Top performers
// - Effectiveness metrics
// - Recommendations
```

## Performance Considerations

### Memory Usage
- Stores last 1000 trades per client
- Keeps 10,000 routing decisions
- Stores 10,000 training samples for ML
- 100,000 audit logs (circular buffer)

### Optimization Tips
1. **Tune ML training frequency** - Default: every 100 samples
2. **Adjust audit log retention** - Reduce if memory constrained
3. **Use rule priorities** - Higher priority rules evaluated first
4. **Monitor exposure updates** - Real-time tracking per symbol

## Security & Compliance

### Data Protection
- All audit logs timestamped and immutable
- Client data encrypted at rest (external implementation)
- API authentication required (external implementation)

### Regulatory Compliance
- MiFID II best execution reporting
- Complete audit trail for all decisions
- Transparency in routing rationale
- No discrimination beyond risk-based criteria

### Risk Management
- Multiple layers of exposure protection
- Automatic hedging triggers
- Alert system for anomalies
- Manual override capabilities

## Best Practices

### 1. Client Classification
- Let system classify clients automatically
- Require minimum 50 trades for stable classification
- Review toxic clients manually before rejection
- Monitor classification drift over time

### 2. Routing Rules
- Use rules for specific exceptions
- Keep rule count minimal (< 20)
- Prefer ML-based routing for general cases
- Document rule rationale

### 3. Exposure Management
- Set conservative limits initially
- Adjust based on historical data
- Monitor exposure during volatile markets
- Review limits weekly

### 4. ML Model
- Enable auto-learning in production
- Export model weekly for backup
- Monitor training accuracy
- Retrain from scratch quarterly

### 5. Compliance
- Review critical alerts daily
- Export audit trail monthly
- Generate best execution reports quarterly
- Keep audit logs for 7 years (regulatory requirement)

## Troubleshooting

### High B-Book Losses
1. Check if too many profitable clients in B-Book
2. Review ML prediction accuracy
3. Lower toxicity thresholds
4. Increase A-Book percentage for semi-pro clients

### Low ROI
1. Review hedging costs vs B-Book profits
2. Consider reducing A-Book percentage for retail
3. Check if classification is too conservative
4. Optimize exposure limits

### Compliance Alerts
1. Review alert details and root cause
2. Adjust routing rules if systematic issue
3. Check for manual overrides causing problems
4. Update thresholds if too sensitive

## Future Enhancements

### Planned Features
1. **Advanced ML Models**
   - Deep learning for pattern recognition
   - LSTM for time-series prediction
   - Ensemble models for higher accuracy

2. **Real-Time Risk Analytics**
   - Live VaR calculation
   - Stress testing scenarios
   - Correlation analysis

3. **Multi-LP Routing**
   - Route to multiple LPs based on pricing
   - Smart order routing (SOR)
   - LP performance tracking

4. **Advanced Netting**
   - Cross-client netting before hedging
   - Net exposure optimization
   - Reduced hedging costs

5. **API Rate Limiting**
   - Per-client rate limits
   - Suspicious pattern detection
   - DDoS protection

## License

Copyright © 2026 Trading Engine. All rights reserved.

## Support

For issues, questions, or feature requests, please contact:
- Technical Support: support@tradingengine.com
- Compliance Questions: compliance@tradingengine.com
- Sales: sales@tradingengine.com
