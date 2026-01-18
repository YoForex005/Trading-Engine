# Export API Reference

## Backend Endpoints Required

The frontend export functionality expects these backend API endpoints:

## 1. PDF Export (Server-Side Generation)

### Endpoint
```
GET /api/analytics/export/pdf
```

### Query Parameters
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| accountId | string | Yes | Account ID for filtering data |
| start | string | Yes | Start date (ISO 8601 format: YYYY-MM-DD) |
| end | string | Yes | End date (ISO 8601 format: YYYY-MM-DD) |

### Headers
```
Authorization: Bearer {jwt_token}
Content-Type: application/json
```

### Response
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="report_20260119_123456.pdf"
```

Binary PDF file with:
- Account summary
- Trade history table
- Performance metrics
- Charts (equity curve, P&L distribution)
- Symbol attribution breakdown

### Example Request
```bash
curl -X GET "http://localhost:8080/api/analytics/export/pdf?accountId=1&start=2024-01-01&end=2024-12-31" \
  -H "Authorization: Bearer eyJhbGc..." \
  -o report.pdf
```

### Go Implementation Reference
See: `/docs/EXPORT_IMPLEMENTATION_GUIDE.md` lines 619-669

## 2. Data Fetch Endpoints (Already Implemented)

### Get Trades
```
GET /api/trades?accountId={id}
```
Returns: Array of Trade objects

### Get Positions
```
GET /api/positions?accountId={id}
```
Returns: Array of Position objects

### Get Performance Report
```
GET /api/performance?accountId={id}&start={date}&end={date}
```
Returns: Performance metrics object

## Frontend Export Flow

### CSV Export (Client-Side)
1. Frontend fetches data from `/api/trades` or `/api/positions`
2. Filters by date range
3. Applies column selection
4. Generates CSV using papaparse
5. Downloads in browser

### PDF Export (Server-Side)
1. Frontend calls `/api/analytics/export/pdf` with date range
2. Backend generates PDF with Go PDF library
3. Returns binary PDF
4. Frontend triggers download

### JSON Export (Client-Side)
1. Frontend fetches data from appropriate endpoint
2. Filters and formats data
3. Converts to JSON
4. Downloads in browser

## Backend Libraries Needed

### PDF Generation
```bash
go get github.com/go-pdf/fpdf
```

### Excel (Optional Future Enhancement)
```bash
go get github.com/xuri/excelize/v2
```

### CSV (Built-in)
```go
import "encoding/csv"
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid date range",
  "message": "Start date must be before end date"
}
```

### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token"
}
```

### 404 Not Found
```json
{
  "error": "Account not found",
  "message": "No account with ID: 123"
}
```

### 500 Internal Server Error
```json
{
  "error": "Export failed",
  "message": "Failed to generate PDF"
}
```

## Security Requirements

1. **Authentication**: All endpoints require valid JWT token
2. **Authorization**: Users can only export their own account data
3. **Rate Limiting**: Limit exports to 10 per minute per user
4. **File Size Limits**: Max 100MB for PDF exports
5. **Audit Logging**: Log all export requests with timestamp, user, and data range

## Performance Guidelines

### PDF Generation
- Generate PDFs asynchronously for large datasets (>10,000 trades)
- Use streaming for large files to avoid memory issues
- Cache frequently requested reports for 5 minutes
- Implement pagination for very large datasets

### Response Times
- Small dataset (<100 records): <1 second
- Medium dataset (100-1,000 records): <3 seconds
- Large dataset (>1,000 records): <10 seconds
- Very large dataset (>10,000 records): Background job with notification

## Data Format Examples

### Trade Object
```json
{
  "id": "TRD-123",
  "symbol": "EURUSD",
  "side": "BUY",
  "volume": 1.0,
  "openPrice": 1.0950,
  "closePrice": 1.0975,
  "profit": 25.00,
  "commission": 2.50,
  "swap": 0.00,
  "openTime": "2024-01-15T10:30:00Z",
  "closeTime": "2024-01-15T14:20:00Z"
}
```

### Performance Report Object
```json
{
  "startDate": "2024-01-01T00:00:00Z",
  "endDate": "2024-12-31T23:59:59Z",
  "totalTrades": 150,
  "winningTrades": 90,
  "losingTrades": 60,
  "winRate": 60.0,
  "netProfit": 5420.50,
  "profitFactor": 1.85,
  "sharpeRatio": 1.45,
  "maxDrawdown": 1250.00,
  "maxDrawdownPct": 12.5,
  "profitBySymbol": {
    "EURUSD": 2500.00,
    "GBPUSD": 1800.00,
    "USDJPY": 1120.50
  }
}
```

## Testing

### Manual Testing
```bash
# Test PDF export
curl -X GET "http://localhost:8080/api/analytics/export/pdf?accountId=1&start=2024-01-01&end=2024-01-31" \
  -H "Authorization: Bearer $TOKEN" \
  -o test_report.pdf

# Verify PDF
file test_report.pdf
# Should output: test_report.pdf: PDF document, version 1.4
```

### Integration Testing
1. Create test account with sample data
2. Make export requests with various date ranges
3. Verify PDF contains correct data
4. Test error cases (invalid dates, missing account, etc.)
5. Test large datasets (>1,000 trades)

## Implementation Priority

1. **High Priority**: PDF export endpoint (most visible feature)
2. **Medium Priority**: Performance report endpoint enhancements
3. **Low Priority**: Excel export (future enhancement)
4. **Low Priority**: Scheduled exports (future enhancement)

## Next Steps

1. Implement Go backend PDF export handler
2. Add audit logging for exports
3. Test with real trading data
4. Add export history tracking
5. Implement scheduled exports (if needed)
