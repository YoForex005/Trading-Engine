# Export & Reporting API Specification
## RESTful API Endpoints and Webhooks

---

## Base URL
```
https://api.trading-engine.local/api/v1
```

---

## Authentication
All export endpoints require:
- Header: `X-Account-ID` (required)
- Header: `X-User-ID` (required)
- Header: `Authorization: Bearer {jwt-token}` (recommended)

---

## Export Endpoints

### 1. Export Trades

#### GET /export/trades

Export trade history in specified format.

**Query Parameters:**
```
format      string  required  csv|json|excel|parquet
start       date    required  2006-01-02
end         date    required  2006-01-02
columns     string  optional  comma-separated column names
timezone    string  optional  America/New_York (default: UTC)
currency    string  optional  USD,EUR,GBP (for conversion)
compressed  bool    optional  true|false (gzip compression)
```

**Example Request:**
```bash
curl -X GET "https://api.trading-engine.local/api/v1/export/trades?format=csv&start=2024-01-01&end=2024-01-31" \
  -H "X-Account-ID: acc_12345" \
  -H "X-User-ID: user_98765" \
  -H "Authorization: Bearer eyJhbGc..."
```

**Response (200 OK):**
```
Content-Type: text/csv
Content-Disposition: attachment; filename="trades_20240119_120000.csv"
X-Audit-ID: audit_xyz789
X-Total-Records: 156

ID,Symbol,Side,Volume,Open Price,...
trade_1,EURUSD,BUY,1.0,1.0950,...
trade_2,GBPUSD,SELL,2.0,1.2750,...
```

**Response (400 Bad Request):**
```json
{
  "error": "invalid_date_format",
  "message": "Date must be in YYYY-MM-DD format",
  "field": "start",
  "example": "2024-01-01"
}
```

**Response (429 Too Many Requests):**
```json
{
  "error": "rate_limit_exceeded",
  "retry_after": 60,
  "message": "Maximum 10 exports per minute"
}
```

**Available Columns:**
```
id, symbol, side, volume, open_price, close_price, profit,
commission, swap, open_time, close_time, holding_time,
mae, mfe, exit_reason, pips, risk_reward_ratio
```

---

### 2. Export Performance Report

#### GET /export/performance

Export performance metrics and analysis.

**Query Parameters:**
```
format      string  required  json|excel|pdf
start       date    required  2006-01-02
end         date    required  2006-01-02
include     string  optional  metrics,attribution,drawdown,all
timezone    string  optional  America/New_York
```

**Example Request:**
```bash
curl -X GET "https://api.trading-engine.local/api/v1/export/performance?format=excel&start=2024-01-01&end=2024-01-31&include=all" \
  -H "X-Account-ID: acc_12345" \
  -H "X-User-ID: user_98765"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "format": "excel",
  "filename": "performance_report_20240119_120000.xlsx",
  "size_bytes": 156234,
  "sheets": ["Summary", "Trades", "Attribution", "Statistics"],
  "generated_at": "2024-01-19T12:00:00Z",
  "audit_id": "audit_abc123"
}
```

**Report Contents (Summary Sheet):**
```
Key Metrics:
- Total Trades: 156
- Win Rate: 62.18%
- Net Profit: $5,234.50
- Sharpe Ratio: 1.45
- Sortino Ratio: 2.13
- Max Drawdown: -8.5%
- Profit Factor: 2.34
- Recovery Factor: 3.12
```

---

### 3. Export Tax Report

#### GET /export/tax-report

Export annual tax report for regulatory filing.

**Query Parameters:**
```
format      string  required  pdf|excel|json
year        int     required  2024
currency    string  optional  USD (default)
jurisdiction string optional  US|EU|UK|CANADA
```

**Example Request:**
```bash
curl -X GET "https://api.trading-engine.local/api/v1/export/tax-report?format=pdf&year=2024&jurisdiction=US" \
  -H "X-Account-ID: acc_12345"
```

**Response (200 OK):**
```json
{
  "status": "success",
  "year": 2024,
  "total_trades": 312,
  "net_profit_usd": 15450.75,
  "short_term_gains": 12300.50,
  "long_term_gains": 3150.25,
  "total_commission": 487.60,
  "total_swap": 234.10,
  "by_symbol": {
    "EURUSD": { "profit": 5123.45, "trades": 98 },
    "GBPUSD": { "profit": 4567.30, "trades": 87 }
  },
  "generated_at": "2024-01-19T12:00:00Z"
}
```

---

### 4. Create Scheduled Export

#### POST /export/schedules

Create a recurring automated export.

**Request Body:**
```json
{
  "name": "Monthly Performance Report",
  "report_type": "performance",
  "frequency": "monthly",
  "day_of_month": 1,
  "time_utc": "09:00:00",
  "format": "pdf",
  "recipients": ["trader@example.com", "advisor@example.com"],
  "include_charts": true,
  "include_trades": true,
  "timezone": "America/New_York",
  "enabled": true,
  "data_retention_days": 365
}
```

**Response (201 Created):**
```json
{
  "id": "schedule_xyz789",
  "status": "active",
  "next_run": "2024-02-01T14:00:00Z",
  "last_run": null,
  "run_count": 0,
  "created_at": "2024-01-19T12:00:00Z"
}
```

**Frequency Values:**
- `daily` (09:00 UTC)
- `weekly` (Monday, 09:00 UTC)
- `monthly` (1st of month, 09:00 UTC)
- `quarterly` (1st of quarter, 09:00 UTC)
- `annually` (Jan 1, 09:00 UTC)

---

### 5. List Scheduled Exports

#### GET /export/schedules

List all scheduled exports for account.

**Query Parameters:**
```
status    string  optional  active|paused|all
limit     int     optional  10-100 (default: 20)
offset    int     optional  0
```

**Response (200 OK):**
```json
{
  "total": 5,
  "limit": 20,
  "offset": 0,
  "schedules": [
    {
      "id": "schedule_xyz789",
      "name": "Monthly Performance Report",
      "frequency": "monthly",
      "next_run": "2024-02-01T14:00:00Z",
      "recipients": ["trader@example.com"],
      "enabled": true,
      "last_run": "2024-01-01T14:00:00Z",
      "run_count": 1
    }
  ]
}
```

---

### 6. Update Scheduled Export

#### PATCH /export/schedules/{scheduleId}

Update scheduled export configuration.

**Request Body:**
```json
{
  "name": "Updated Report Name",
  "recipients": ["newemail@example.com"],
  "enabled": false,
  "time_utc": "10:00:00"
}
```

**Response (200 OK):**
```json
{
  "id": "schedule_xyz789",
  "status": "updated",
  "updated_at": "2024-01-19T12:00:00Z"
}
```

---

### 7. Delete Scheduled Export

#### DELETE /export/schedules/{scheduleId}

**Response (204 No Content):**

---

### 8. Get Export History

#### GET /export/history

Retrieve past exports and audit trail.

**Query Parameters:**
```
limit       int     optional  50
offset      int     optional  0
format      string  optional  csv|json|excel|pdf
start_date  date    optional  2006-01-02
end_date    date    optional  2006-01-02
status      string  optional  success|failed|all
```

**Response (200 OK):**
```json
{
  "total": 42,
  "exports": [
    {
      "id": "export_abc123",
      "type": "trades",
      "format": "csv",
      "filename": "trades_20240119_120000.csv",
      "record_count": 156,
      "file_size_bytes": 45678,
      "status": "completed",
      "requested_at": "2024-01-19T12:00:00Z",
      "completed_at": "2024-01-19T12:00:15Z",
      "requested_by": "user_98765",
      "ip_address": "192.168.1.100"
    }
  ]
}
```

---

### 9. GDPR Data Export

#### POST /export/gdpr-request

Request complete personal data export (GDPR Article 15).

**Request Body:**
```json
{
  "format": "json",
  "include_audit_log": true,
  "encrypt_with_pgp": false
}
```

**Response (202 Accepted):**
```json
{
  "request_id": "gdpr_req_xyz789",
  "status": "processing",
  "estimated_completion": "2024-01-19T18:00:00Z",
  "download_expires_at": "2024-02-19T12:00:00Z",
  "message": "Request queued. You'll receive email when ready."
}
```

**Response (200 OK) - if ready:**
```json
{
  "request_id": "gdpr_req_xyz789",
  "status": "ready",
  "download_url": "https://api.trading-engine.local/api/v1/export/gdpr/xyz789/download?token=token123",
  "download_token": "token123",
  "expires_at": "2024-02-19T12:00:00Z",
  "contains": ["personal_data", "trades", "reports", "audit_log"]
}
```

---

### 10. Download GDPR Export

#### GET /export/gdpr/{requestId}/download

Download the prepared GDPR data export.

**Query Parameters:**
```
token  string  required  Download token from GDPR request
```

**Response (200 OK):**
```
Content-Type: application/zip
Content-Disposition: attachment; filename="gdpr_export_20240119.zip"
Content-Length: 2345678

[Binary ZIP file containing JSON files]
```

---

## Webhook Events

### Webhook Format

All webhook events POST to registered endpoint with:

**Headers:**
```
X-Webhook-Signature: sha256=<HMAC-SHA256 of body>
X-Timestamp: 2024-01-19T12:00:00Z
X-Event-ID: evt_xyz789
X-Retry-Count: 0
```

**Body:**
```json
{
  "event_type": "export:completed",
  "timestamp": "2024-01-19T12:00:00Z",
  "account_id": "acc_12345",
  "data": {
    // Event-specific data
  }
}
```

---

### Webhook Event Types

#### export:started
```json
{
  "event_type": "export:started",
  "export_id": "export_abc123",
  "type": "trades",
  "format": "csv",
  "requested_by": "user_98765"
}
```

#### export:completed
```json
{
  "event_type": "export:completed",
  "export_id": "export_abc123",
  "type": "trades",
  "format": "csv",
  "filename": "trades_20240119_120000.csv",
  "record_count": 156,
  "duration_seconds": 15,
  "file_size_bytes": 45678,
  "download_url": "https://api.trading-engine.local/api/v1/export/downloads/abc123"
}
```

#### export:failed
```json
{
  "event_type": "export:failed",
  "export_id": "export_abc123",
  "type": "trades",
  "error": "database_timeout",
  "message": "Export query timed out after 60 seconds",
  "retry_possible": true
}
```

#### report:scheduled
```json
{
  "event_type": "report:scheduled",
  "schedule_id": "schedule_xyz789",
  "name": "Monthly Performance Report",
  "next_run": "2024-02-01T14:00:00Z"
}
```

---

### Register Webhook Endpoint

#### POST /webhooks

**Request Body:**
```json
{
  "url": "https://yourapp.example.com/webhooks/trading-engine",
  "secret": "whsec_xyz789",
  "events": [
    "export:started",
    "export:completed",
    "export:failed",
    "report:scheduled"
  ],
  "active": true
}
```

**Response (201 Created):**
```json
{
  "id": "webhook_abc123",
  "url": "https://yourapp.example.com/webhooks/trading-engine",
  "events": ["export:started", "export:completed", "export:failed"],
  "active": true,
  "test_pending": true,
  "created_at": "2024-01-19T12:00:00Z"
}
```

---

### Test Webhook

#### POST /webhooks/{webhookId}/test

**Response (200 OK):**
```json
{
  "status": "success",
  "response_status": 200,
  "response_time_ms": 145
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid_request",
  "message": "Date must be in YYYY-MM-DD format",
  "field": "start",
  "details": {
    "provided": "01/01/2024",
    "expected_format": "2024-01-01"
  }
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized",
  "message": "Missing or invalid authentication token"
}
```

### 403 Forbidden
```json
{
  "error": "insufficient_permissions",
  "message": "User does not have access to this account"
}
```

### 404 Not Found
```json
{
  "error": "not_found",
  "message": "Export schedule not found",
  "id": "schedule_xyz789"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate_limit_exceeded",
  "message": "Maximum 10 concurrent exports per account",
  "retry_after": 30
}
```

### 500 Internal Server Error
```json
{
  "error": "internal_error",
  "message": "An unexpected error occurred",
  "error_id": "err_xyz789",
  "support_email": "support@trading-engine.local"
}
```

---

## Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/export/trades` | 10 requests | 60 seconds |
| `/export/performance` | 10 requests | 60 seconds |
| `/export/tax-report` | 5 requests | 60 seconds |
| `/export/schedules` | 100 requests | 3600 seconds |
| `/export/history` | 30 requests | 60 seconds |
| `/export/gdpr-request` | 2 requests | 86400 seconds |

**Rate Limit Headers:**
```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 8
X-RateLimit-Reset: 2024-01-19T12:01:00Z
```

---

## Data Retention & Compliance

### Export Retention Policy
- **Default:** 30 days
- **Configurable:** 1-365 days per export
- **Automatic Deletion:** Expired exports deleted automatically
- **Audit Log:** Deletion events logged indefinitely

### GDPR Compliance
- **Right to Access:** GET /export/gdpr-request
- **Right to Erasure:** DELETE /accounts/{accountId}/data
- **Data Portability:** All exports in standard formats
- **Consent Tracking:** Optional consent headers

---

## Code Examples

### Python - Export Trades
```python
import requests
from datetime import datetime, timedelta

headers = {
    'X-Account-ID': 'acc_12345',
    'X-User-ID': 'user_98765',
    'Authorization': 'Bearer token_xyz'
}

params = {
    'format': 'csv',
    'start': (datetime.now() - timedelta(days=30)).strftime('%Y-%m-%d'),
    'end': datetime.now().strftime('%Y-%m-%d')
}

response = requests.get(
    'https://api.trading-engine.local/api/v1/export/trades',
    params=params,
    headers=headers
)

if response.status_code == 200:
    with open('trades.csv', 'wb') as f:
        f.write(response.content)
    print(f"Audit ID: {response.headers['X-Audit-ID']}")
else:
    print(f"Error: {response.status_code}")
    print(response.json())
```

### JavaScript - Create Scheduled Export
```javascript
async function createScheduledExport() {
  const response = await fetch(
    'https://api.trading-engine.local/api/v1/export/schedules',
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Account-ID': 'acc_12345',
        'X-User-ID': 'user_98765',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        name: 'Weekly Report',
        report_type: 'performance',
        frequency: 'weekly',
        format: 'pdf',
        recipients: ['trader@example.com'],
        enabled: true
      })
    }
  );

  const result = await response.json();
  console.log('Schedule ID:', result.id);
  return result;
}
```

### cURL - Export with Webhook Notification
```bash
# Create export that triggers webhook
curl -X POST "https://api.trading-engine.local/api/v1/export/trades" \
  -H "X-Account-ID: acc_12345" \
  -H "X-User-ID: user_98765" \
  -H "Content-Type: application/json" \
  -d '{
    "format": "excel",
    "start": "2024-01-01",
    "end": "2024-01-31",
    "webhook_url": "https://yourapp.example.com/webhooks/export"
  }'
```

---

## Pagination

All list endpoints support pagination:

**Query Parameters:**
```
limit   int  optional  1-100 (default: 20)
offset  int  optional  0
```

**Response:**
```json
{
  "total": 245,
  "limit": 20,
  "offset": 0,
  "has_more": true,
  "next_offset": 20,
  "items": [...]
}
```

---

## Versioning

Current API Version: **v1**

Future versions will support:
- `/api/v2` - Enhanced features
- Backward compatibility maintained for 12 months
- Deprecation notices 6 months in advance

---

## SDK Clients

### Official SDKs (Coming Soon)
- Python: `pip install trading-engine-sdk`
- JavaScript/TypeScript: `npm install trading-engine-sdk`
- Go: `go get github.com/epic1st/trading-engine-sdk`

### Community SDKs
- Java: [tradeengine-java](https://github.com/community/tradeengine-java)
- Ruby: [tradeengine-ruby](https://github.com/community/tradeengine-ruby)

---

## Support

- **Documentation:** https://docs.trading-engine.local
- **API Status:** https://status.trading-engine.local
- **Email Support:** support@trading-engine.local
- **Slack:** #api-support in workspace
