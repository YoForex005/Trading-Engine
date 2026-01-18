# Analytics Handler Stubs Needed

The integration tests require the following handler methods to be implemented in the API handler.

## Required Handler Methods

### 1. HandleRoutingMetrics
```go
func (h *APIHandler) HandleRoutingMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse timeRange query parameter (1h, 24h, 7d)
	// Aggregate routing decisions from database/memory
	// Calculate:
	//   - abook_count
	//   - bbook_count
	//   - cbook_count
	//   - total_volume
	//   - avg_latency
	// Return JSON response
}
```

### 2. HandleLPPerformance
```go
func (h *APIHandler) HandleLPPerformance(w http.ResponseWriter, r *http.Request) {
	// Parse timeRange and optional lp query parameters
	// Aggregate LP metrics:
	//   - latency_avg
	//   - spread_avg
	//   - volume
	//   - uptime
	//   - quote_count
	// Return JSON response with LP array
}
```

### 3. HandleExposureHeatmap
```go
func (h *APIHandler) HandleExposureHeatmap(w http.ResponseWriter, r *http.Request) {
	// Parse groupBy query parameter (symbol, side, account)
	// Aggregate current open positions
	// Calculate:
	//   - buy_volume
	//   - sell_volume
	//   - net_exposure
	//   - position_count
	// Return JSON response grouped by requested dimension
}
```

### 4. HandleRuleEffectiveness
```go
func (h *APIHandler) HandleRuleEffectiveness(w http.ResponseWriter, r *http.Request) {
	// Analyze routing rule performance
	// Calculate for each rule:
	//   - win_rate
	//   - sharpe_ratio
	//   - total_trades
	//   - avg_profit
	//   - avg_loss
	// Return JSON response with rules array
}
```

### 5. HandleComplianceReport
```go
func (h *APIHandler) HandleComplianceReport(w http.ResponseWriter, r *http.Request) {
	// Parse start and end date parameters
	// Validate date range
	// Aggregate compliance metrics:
	//   - total_trades
	//   - total_volume
	//   - largest_trade
	//   - unusual_patterns
	//   - regulatory_flags
	// Return JSON response
}
```

## Response Formats

### Routing Metrics Response
```json
{
  "abook_count": 150,
  "bbook_count": 320,
  "cbook_count": 80,
  "total_volume": 15000.5,
  "avg_latency": 45.2,
  "time_range": "24h"
}
```

### LP Performance Response
```json
{
  "lps": [
    {
      "name": "oanda",
      "latency_avg": 50.5,
      "spread_avg": 0.0002,
      "volume": 10000,
      "uptime": 99.9,
      "quote_count": 50000
    },
    {
      "name": "binance",
      "latency_avg": 30.2,
      "spread_avg": 0.0001,
      "volume": 15000,
      "uptime": 99.8,
      "quote_count": 75000
    }
  ],
  "time_range": "24h"
}
```

### Exposure Heatmap Response
```json
{
  "exposure": [
    {
      "group": "EURUSD",
      "buy_volume": 10.5,
      "sell_volume": 5.2,
      "net_exposure": 5.3,
      "position_count": 15
    },
    {
      "group": "GBPUSD",
      "buy_volume": 8.0,
      "sell_volume": 12.5,
      "net_exposure": -4.5,
      "position_count": 20
    }
  ],
  "group_by": "symbol"
}
```

### Rule Effectiveness Response
```json
{
  "rules": [
    {
      "id": 1,
      "name": "High Volume to C-Book",
      "win_rate": 0.65,
      "sharpe_ratio": 1.8,
      "total_trades": 150,
      "avg_profit": 50.5,
      "avg_loss": -30.2
    }
  ],
  "overall_sharpe": 1.5,
  "overall_win_rate": 0.60
}
```

### Compliance Report Response
```json
{
  "start_date": "2024-01-01",
  "end_date": "2024-01-31",
  "total_trades": 5000,
  "total_volume": 150000.50,
  "largest_trade": 1000.0,
  "unusual_patterns": 3,
  "regulatory_flags": 0,
  "summary": "All metrics within normal ranges"
}
```

## Error Responses

All handlers should return appropriate HTTP status codes and JSON error responses:

```json
{
  "error": "Invalid time range parameter",
  "code": "INVALID_PARAMETER",
  "details": "timeRange must be one of: 1h, 24h, 7d"
}
```

Common status codes:
- 200: Success
- 400: Bad Request (invalid parameters)
- 401: Unauthorized
- 404: Not Found (invalid LP name, etc.)
- 429: Too Many Requests (rate limiting)
- 500: Internal Server Error

## Implementation Checklist

- [ ] Add handler methods to APIHandler struct
- [ ] Implement parameter parsing and validation
- [ ] Add data aggregation logic
- [ ] Implement error handling
- [ ] Add CORS headers
- [ ] Register routes in main.go
- [ ] Add authentication middleware
- [ ] Add rate limiting
- [ ] Write unit tests for each handler
- [ ] Run integration tests
- [ ] Update API documentation (swagger.yaml)
