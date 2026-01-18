# Intelligent Alerting System

Real-time alerting engine for the trading platform with threshold monitoring, anomaly detection, and multi-channel notifications.

## Features

### Alert Types

1. **Threshold Alerts** - Trigger when metrics exceed configured thresholds
   - Margin level < 100% (margin call)
   - Exposure > 80% of equity
   - Free margin < $500
   - Custom metric thresholds

2. **Anomaly Detection** - Statistical outlier detection using Z-score analysis
   - Detects unusual account behavior
   - Requires 30+ historical samples
   - Configurable Z-score threshold (default: 3.0)

3. **Pattern Alerts** - Detect repeated events (coming soon)
   - Consecutive failures
   - Time-based patterns
   - Correlation detection

### Notification Channels

- **Dashboard** - Real-time WebSocket notifications to connected clients
- **Email** - SMTP email delivery (configure via environment variables)
- **SMS** - Twilio SMS notifications (optional)
- **Webhooks** - Custom HTTP POST integrations (optional)

### Alert Management

- **Acknowledge** - Mark alerts as reviewed
- **Snooze** - Temporarily suppress for 5-60 minutes
- **Resolve** - Mark condition as resolved
- **Cooldown** - Prevent alert spam (default: 5 minutes between same alert)
- **Rate Limiting** - Max 100 alerts/hour per account

## API Endpoints

### List Alerts
```http
GET /api/alerts?accountId=demo-user&status=active
```

**Response:**
```json
{
  "alerts": [
    {
      "id": "alert-123",
      "ruleId": "rule-456",
      "accountId": "demo-user",
      "type": "threshold",
      "severity": "HIGH",
      "status": "active",
      "title": "Low Margin Warning",
      "message": "marginLevel < 150.0 (threshold: 150.0)",
      "metric": "marginLevel",
      "value": 145.5,
      "threshold": 150.0,
      "createdAt": "2026-01-19T12:00:00Z"
    }
  ],
  "count": 1
}
```

### Acknowledge Alert
```http
POST /api/alerts/acknowledge
Content-Type: application/json

{
  "alertId": "alert-123",
  "userId": "demo-user"
}
```

### Snooze Alert
```http
POST /api/alerts/snooze
Content-Type: application/json

{
  "alertId": "alert-123",
  "durationMinutes": 30
}
```

### Create Alert Rule
```http
POST /api/alerts/rules
Content-Type: application/json

{
  "accountId": "demo-user",
  "name": "High Exposure Alert",
  "type": "threshold",
  "severity": "HIGH",
  "enabled": true,
  "metric": "exposurePercent",
  "operator": ">",
  "threshold": 80.0,
  "channels": ["dashboard", "email"],
  "cooldownSeconds": 600
}
```

### List Rules
```http
GET /api/alerts/rules?accountId=demo-user
```

### Update Rule
```http
PUT /api/alerts/rules/{ruleId}
Content-Type: application/json

{
  "name": "Updated Rule Name",
  "enabled": false,
  "threshold": 90.0
}
```

### Delete Rule
```http
DELETE /api/alerts/rules/{ruleId}
```

## Configuration

### Environment Variables

```bash
# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_ADDRESS=alerts@yourcompany.com

# SMS (Twilio - optional)
TWILIO_ACCOUNT_SID=ACxxxx
TWILIO_AUTH_TOKEN=xxxx
TWILIO_FROM_NUMBER=+1234567890
```

### Default Alert Rules

The system includes 7 pre-configured alert templates:

1. **Margin Call Alert** (CRITICAL) - Margin level < 100%
2. **Low Margin Warning** (HIGH) - Margin level < 150%
3. **High Exposure Alert** (HIGH) - Exposure > 80%
4. **Large Unrealized Loss** (MEDIUM) - Loss > 20% of balance
5. **Equity Anomaly Detection** (MEDIUM) - Statistical outliers
6. **Low Free Margin** (MEDIUM) - Free margin < $500
7. **Position Count Warning** (LOW) - More than 10 positions

Access default rules via:
```go
import "github.com/epic1st/rtx/backend/internal/alerts"

defaultRules := alerts.GetDefaultRules("demo-user")
for _, rule := range defaultRules {
    alertEngine.AddRule(rule)
}
```

## Metrics Available

- `balance` - Account cash balance
- `equity` - Balance + unrealized P/L
- `margin` - Total margin used
- `freeMargin` - Available margin
- `marginLevel` - Equity / Margin × 100 (%)
- `exposurePercent` - Margin / Equity × 100 (%)
- `pnl` - Unrealized profit/loss
- `positionCount` - Number of open positions

## Architecture

### Components

1. **Engine** (`engine.go`) - Core evaluation loop (runs every 5 seconds)
2. **Notifier** (`notifier.go`) - Multi-channel notification dispatcher
3. **MetricsAdapter** (`metrics_adapter.go`) - B-Book engine integration
4. **WSAlertHub** (`ws_integration.go`) - WebSocket broadcasting
5. **AlertsHandler** (`handlers/alerts.go`) - HTTP API handlers

### Data Flow

```
Market Data → Account Metrics → Alert Engine (5s loop)
                                      ↓
                              Rule Evaluation
                                      ↓
                        Threshold/Anomaly/Pattern Check
                                      ↓
                            Trigger Alert Event
                                      ↓
                    Deduplication → Rate Limiting
                                      ↓
                              Notifier Queue
                                      ↓
            Dashboard WebSocket | Email SMTP | SMS Twilio
```

### Evaluation Loop

The alert engine runs a background goroutine that evaluates all enabled rules every 5 seconds:

1. Fetch account metrics snapshot
2. Evaluate each enabled rule
3. Check cooldown and rate limits
4. Create alert if triggered
5. Queue notifications for delivery
6. Broadcast via WebSocket

### Deduplication

Alerts are deduplicated using a SHA256 fingerprint based on:
- Rule ID
- Account ID
- Alert message

Duplicate alerts within 5 minutes are suppressed.

## Performance

- **Evaluation Frequency**: 5 seconds
- **Notification Latency**: <500ms (dashboard), <2s (email/SMS)
- **Rate Limit**: 100 alerts/hour per account
- **Queue Capacity**: 1000 pending notifications
- **Worker Threads**: 3 parallel notification workers
- **Memory per Alert**: ~1KB

## Security

- **Authentication**: JWT token validation for WebSocket connections
- **Authorization**: Account-based rule filtering
- **Rate Limiting**: Token bucket algorithm per account
- **Audit Trail**: All alert events logged with timestamps

## Testing

```bash
# Run alert engine tests
cd backend/internal/alerts
go test -v

# Test specific functionality
go test -run TestThresholdEvaluation
go test -run TestAnomalyDetection
```

## Example Usage

### Programmatic Alert Creation

```go
import "github.com/epic1st/rtx/backend/internal/alerts"

// Create margin call alert
rule := &alerts.AlertRule{
    AccountID: "demo-user",
    Name: "Margin Call",
    Type: alerts.AlertTypeThreshold,
    Severity: alerts.AlertSeverityCritical,
    Enabled: true,
    Metric: "marginLevel",
    Operator: "<",
    Threshold: 100.0,
    Channels: []string{"dashboard", "email", "sms"},
    CooldownSeconds: 300,
}

err := alertEngine.AddRule(rule)
```

### WebSocket Alert Subscription (Frontend)

```typescript
const ws = new WebSocket('ws://localhost:7999/ws?token=your-jwt-token');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === 'alert') {
    const alert = data.alert;
    showNotification({
      title: alert.title,
      message: alert.message,
      severity: alert.severity,
      timestamp: alert.createdAt
    });
  }
};
```

## Troubleshooting

### Alerts Not Triggering

1. Check if rule is enabled: `GET /api/alerts/rules`
2. Verify metrics are being calculated: Check logs for `[MetricsAdapter]`
3. Check cooldown hasn't expired: `cooldownSeconds` setting
4. Verify rate limit not exceeded: 100/hour max

### Notifications Not Delivered

1. **Dashboard**: Check WebSocket connection and authentication
2. **Email**: Verify SMTP credentials and `SMTP_HOST` configuration
3. **SMS**: Check Twilio credentials and account balance
4. Check notification queue: Engine logs show `[Notifier]` status

### High CPU Usage

- Reduce evaluation frequency (modify ticker in `engine.go`)
- Disable unused anomaly detection rules (requires historical data)
- Limit number of active rules per account

## Future Enhancements

- [ ] Pattern detection implementation (consecutive failures)
- [ ] Composite alerts (multiple conditions)
- [ ] Predictive alerts (linear regression forecasting)
- [ ] Alert correlation analysis
- [ ] Machine learning for dynamic thresholds
- [ ] Slack/Teams integration
- [ ] PagerDuty escalation
- [ ] Business hours awareness
- [ ] Alert templates marketplace
- [ ] Database persistence (currently in-memory)

## License

Proprietary - YoForex Trading Engine
