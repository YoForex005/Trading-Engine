# Rule Effectiveness Analytics API

## Overview

The Rule Effectiveness Analytics API provides comprehensive financial performance metrics for routing rules in the trading engine. It calculates industry-standard metrics including Sharpe Ratio, Profit Factor, and Maximum Drawdown.

## API Endpoints

### 1. GET /api/analytics/rules/effectiveness

Returns effectiveness metrics for all routing rules.

**Query Parameters:**
- `start_time` (optional): Unix timestamp - filter trades from this time
- `end_time` (optional): Unix timestamp - filter trades until this time
- `min_trades` (optional): Minimum number of trades required (default: 1)

**Response:**
```json
{
  "rules": [
    {
      "rule_id": "rule_12345",
      "sharpe_ratio": 1.85,
      "profit_factor": 2.3,
      "max_drawdown": 0.12,
      "total_pnl": 15420.50,
      "trade_count": 234,
      "win_rate": 62.5,
      "rank": 1
    }
  ]
}
```

**Example:**
```bash
curl "http://localhost:8080/api/analytics/rules/effectiveness?start_time=1704067200&min_trades=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### 2. GET /api/analytics/rules/{rule_id}/metrics

Returns detailed metrics for a specific routing rule, including timeline data.

**Query Parameters:**
- `start_time` (optional): Unix timestamp - filter trades from this time
- `end_time` (optional): Unix timestamp - filter trades until this time

**Response:**
```json
{
  "rule_id": "rule_12345",
  "sharpe_ratio": 1.85,
  "profit_factor": 2.3,
  "max_drawdown": 0.12,
  "total_pnl": 15420.50,
  "trade_count": 234,
  "win_rate": 62.5,
  "avg_trade_duration": 3600,
  "consecutive_wins": 8,
  "consecutive_losses": 3,
  "timeline": [
    {
      "timestamp": "2024-01-01T10:00:00Z",
      "pnl": 150.25,
      "equity": 150.25
    },
    {
      "timestamp": "2024-01-01T11:30:00Z",
      "pnl": -45.50,
      "equity": 104.75
    }
  ]
}
```

**Example:**
```bash
curl "http://localhost:8080/api/analytics/rules/rule_12345/metrics?start_time=1704067200" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

### 3. POST /api/analytics/rules/calculate

Calculates metrics for a custom set of trades (useful for backtesting).

**Request Body:**
```json
{
  "trades": [
    {
      "pnl": 150.25,
      "timestamp": "2024-01-01T10:00:00Z"
    },
    {
      "pnl": -45.50,
      "timestamp": "2024-01-01T11:30:00Z"
    }
  ],
  "risk_free_rate": 0.02
}
```

**Response:**
```json
{
  "sharpe_ratio": 1.45,
  "profit_factor": 1.8,
  "max_drawdown": 0.15,
  "win_rate": 55.0,
  "consecutive_wins": 5,
  "consecutive_losses": 2
}
```

**Example:**
```bash
curl -X POST "http://localhost:8080/api/analytics/rules/calculate" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "trades": [
      {"pnl": 100, "timestamp": "2024-01-01T10:00:00Z"},
      {"pnl": -50, "timestamp": "2024-01-01T11:00:00Z"},
      {"pnl": 75, "timestamp": "2024-01-01T12:00:00Z"}
    ],
    "risk_free_rate": 0.02
  }'
```

---

## Financial Metrics Explained

### Sharpe Ratio
Measures risk-adjusted returns.

**Formula:**
```
Sharpe Ratio = (Average Return - Risk Free Rate) / Standard Deviation of Returns
```

**Interpretation:**
- > 3.0: Excellent
- 2.0 - 3.0: Very Good
- 1.0 - 2.0: Good
- 0.0 - 1.0: Acceptable
- < 0.0: Poor (losing more than risk-free rate)

**Implementation:** Annualized using sqrt(252) for daily trading frequency.

---

### Profit Factor
Ratio of gross profits to gross losses.

**Formula:**
```
Profit Factor = Gross Profit / Gross Loss
```

**Interpretation:**
- > 2.0: Excellent
- 1.5 - 2.0: Good
- 1.0 - 1.5: Acceptable
- < 1.0: Losing strategy

**Edge Cases:**
- Returns `Infinity` if there are no losing trades
- Returns `0.0` if there are no winning trades

---

### Maximum Drawdown
The largest peak-to-trough decline in equity.

**Formula:**
```
Max Drawdown = (Peak Equity - Trough Equity) / Peak Equity
```

**Interpretation:**
- Expressed as a percentage (e.g., 0.15 = 15% drawdown)
- Lower is better
- > 50%: High risk
- 20-50%: Moderate risk
- < 20%: Low risk

**Implementation:** Calculates running equity curve and tracks maximum decline from any peak.

---

### Win Rate
Percentage of winning trades.

**Formula:**
```
Win Rate = (Winning Trades / Total Trades) × 100
```

**Interpretation:**
- > 60%: Excellent
- 50-60%: Good
- 40-50%: Acceptable (if profit factor is high)
- < 40%: Poor (unless profit factor is very high)

---

## Implementation Details

### File Locations
- **Handler:** `/backend/internal/api/handlers/analytics_rules.go`
- **Routes:** `/backend/cmd/server/main.go` (lines 375-429)
- **Tests:** `/backend/internal/api/handlers/analytics_rules_test.go`

### Key Functions

#### `calculateSharpeRatio(trades []core.Trade, riskFreeRate float64) float64`
- Computes annualized Sharpe Ratio
- Assumes 252 trading days per year
- Uses standard deviation of returns

#### `calculateProfitFactor(trades []core.Trade) float64`
- Sums gross profits and gross losses
- Handles edge cases (all wins, all losses)

#### `calculateMaxDrawdown(trades []core.Trade) float64`
- Sorts trades chronologically
- Builds equity curve
- Tracks peak and maximum decline

#### `calculateWinRate(trades []core.Trade) float64`
- Counts trades with PnL > 0
- Returns percentage

#### `calculateConsecutiveWinsLosses(trades []core.Trade) (int, int)`
- Tracks maximum consecutive winning/losing streaks
- Useful for understanding consistency

#### `calculateTimeline(trades []core.Trade) []TimelineDataPoint`
- Generates equity curve timeline
- Includes PnL and cumulative equity for each trade

---

## Authentication

All endpoints require admin authentication via Bearer token:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

Requests without valid authentication will receive `401 Unauthorized`.

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request body"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized"
}
```

### 404 Not Found
```json
{
  "error": "No trades found for this rule in the specified time range"
}
```

### 500 Internal Server Error
```json
{
  "error": "Routing engine not available"
}
```

---

## Usage Examples

### Scenario 1: Compare All Rules
```bash
# Get effectiveness metrics for all rules with at least 20 trades
curl "http://localhost:8080/api/analytics/rules/effectiveness?min_trades=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Scenario 2: Analyze Single Rule Performance
```bash
# Get detailed metrics and equity curve for a specific rule
curl "http://localhost:8080/api/analytics/rules/rule_12345/metrics" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Scenario 3: Backtest Custom Strategy
```bash
# Calculate metrics for backtest results
curl -X POST "http://localhost:8080/api/analytics/rules/calculate" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d @backtest_trades.json
```

---

## Performance Considerations

1. **Large Datasets:** For rules with thousands of trades, consider using time filters to limit the dataset.

2. **Real-time Calculations:** All metrics are calculated dynamically - no caching is performed.

3. **Database Queries:** The current implementation queries all trades and filters in memory. For production, consider:
   - Storing `rule_id` with each trade in the database
   - Implementing database-level filtering and aggregation
   - Adding indexes on `rule_id` and `executed_at` columns

---

## Future Enhancements

- [ ] Add Sortino Ratio (downside deviation)
- [ ] Add Calmar Ratio (return / max drawdown)
- [ ] Add monthly/yearly breakdowns
- [ ] Cache frequently accessed metrics
- [ ] Add real-time WebSocket updates
- [ ] Export to CSV/Excel
- [ ] Compare multiple rules side-by-side
- [ ] Add statistical significance tests
- [ ] Monte Carlo simulation support

---

## Testing

Run the test suite:
```bash
cd /backend
go test ./internal/api/handlers -run TestCalculate -v
```

Test coverage includes:
- Sharpe Ratio calculation edge cases
- Profit Factor with various win/loss scenarios
- Maximum Drawdown with complex equity curves
- Win Rate accuracy
- Consecutive wins/losses tracking
- Timeline generation

---

## Production Readiness Checklist

- [x] Production-ready financial calculations
- [x] Comprehensive error handling
- [x] Input validation
- [x] Admin authentication
- [x] CORS headers
- [x] Test suite
- [ ] Database integration (currently in-memory)
- [ ] Rate limiting
- [ ] Request logging
- [ ] Metric caching
- [ ] Performance profiling

---

## References

- Sharpe, William F. (1994). "The Sharpe Ratio". Journal of Portfolio Management.
- Aronson, David (2006). "Evidence-Based Technical Analysis". Wiley.
- Pardo, Robert (2008). "The Evaluation and Optimization of Trading Strategies". Wiley.
