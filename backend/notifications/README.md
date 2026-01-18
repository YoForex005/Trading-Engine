# Notifications System

Comprehensive multi-channel notification system for the trading platform.

## Features

### Multiple Channels
- **Email** - SMTP with HTML templates (SendGrid/AWS SES compatible)
- **SMS** - Twilio integration with delivery tracking
- **Push** - Firebase Cloud Messaging for mobile apps
- **Webhook** - HMAC-signed webhooks for integrations
- **In-App** - Real-time notifications in web dashboard
- **Telegram** - (Future) Bot integration

### Key Capabilities

#### User Preferences
- Per-event type channel selection
- Quiet hours (suppress non-critical notifications)
- Priority-based filtering
- Unsubscribe/resubscribe management
- Localization support
- Import/export preferences

#### Reliability
- **Rate Limiting** - Per-channel limits (minute/hour/day)
- **Retry Logic** - Exponential backoff for failures
- **Delivery Tracking** - Full delivery history with status
- **Batching** - Combine multiple notifications into digests
- **Priority Handling** - Critical notifications bypass quiet hours

#### Templates
- HTML email templates with variables
- SMS formatting helpers
- Push notification payloads
- Webhook event builders
- Support for A/B testing

#### Compliance
- GDPR consent tracking
- CAN-SPAM unsubscribe links
- Opt-in/opt-out management
- Audit trail for all deliveries

## Architecture

```
NotificationManager (coordinator)
├── EmailNotifier → SMTPProvider / SendGrid / SES
├── SMSNotifier → TwilioProvider
├── PushNotifier → FCMProvider
├── WebhookNotifier → HTTP with HMAC
├── PreferencesManager → User preferences store
├── RateLimiter → Channel-specific limits
├── RetryQueue → Failed notification retry
└── BatchProcessor → Notification batching
```

## Usage

### Initialize the System

```go
import "github.com/epic1st/rtx/backend/notifications"

// Configure providers
smtpConfig := notifications.SMTPConfig{
    Host:     "smtp.sendgrid.net",
    Port:     587,
    Username: "apikey",
    Password: os.Getenv("SENDGRID_API_KEY"),
    From:     "noreply@tradingplatform.com",
    FromName: "Trading Platform",
    UseTLS:   true,
}

twilioConfig := notifications.TwilioConfig{
    AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
    AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
    FromNumber: "+1234567890",
}

fcmConfig := notifications.FCMConfig{
    ServerKey: os.Getenv("FCM_SERVER_KEY"),
    ProjectID: "trading-platform",
}

// Create notifiers
rateLimiter := notifications.NewRateLimiter(notifications.DefaultRateLimitConfigs())
emailNotifier := notifications.NewEmailNotifier(
    notifications.NewSMTPProvider(smtpConfig),
    "noreply@tradingplatform.com",
    "Trading Platform",
    rateLimiter,
)
smsNotifier := notifications.NewSMSNotifier(
    notifications.NewTwilioProvider(twilioConfig),
    rateLimiter,
)
pushNotifier := notifications.NewPushNotifier(
    notifications.NewFCMProvider(fcmConfig),
    rateLimiter,
)
webhookNotifier := notifications.NewWebhookNotifier(rateLimiter)

// Create preference and delivery stores
prefsStore := notifications.NewInMemoryPreferencesStore()
deliveryStore := notifications.NewInMemoryDeliveryStore()
prefsManager := notifications.NewPreferencesManager(prefsStore)

// Create notification manager
manager := notifications.NewNotificationManager(
    emailNotifier,
    smsNotifier,
    pushNotifier,
    webhookNotifier,
    prefsManager,
    deliveryStore,
    rateLimiter,
)
```

### Send Notifications

```go
// Order executed notification
notif := notifications.CreateOrderExecutedNotification(
    userID,
    orderID,
    "BTCUSD",
    "BUY",
    0.5,
    50000.00,
)

contacts := notifications.UserContacts{
    Email:        "user@example.com",
    Phone:        "+1234567890",
    DeviceTokens: []string{"fcm-device-token"},
}

err := manager.Send(ctx, notif, contacts)

// Margin call warning
notif = notifications.CreateMarginCallNotification(
    userID,
    accountID,
    45.5,  // Current margin level
    100.0, // Required margin level
)

err = manager.Send(ctx, notif, contacts)

// Position closed with P&L
notif = notifications.CreatePositionClosedNotification(
    userID,
    positionID,
    "ETHUSD",
    1250.50, // Profit
)

err = manager.Send(ctx, notif, contacts)
```

### Manage User Preferences

```go
// Get user preferences
prefs, err := prefsManager.GetPreferences(ctx, userID)

// Update channel preference for specific notification type
err = prefsManager.SetChannelPreference(
    ctx,
    userID,
    notifications.NotifOrderExecuted,
    []notifications.NotificationChannel{
        notifications.ChannelEmail,
        notifications.ChannelPush,
    },
    true, // enabled
)

// Set quiet hours
err = prefsManager.SetQuietHours(
    ctx,
    userID,
    true,      // enabled
    "22:00",   // start time
    "08:00",   // end time
    "America/New_York",
)

// Unsubscribe from all non-critical notifications
err = prefsManager.UnsubscribeAll(ctx, userID)

// Export preferences
jsonData, err := prefsManager.ExportPreferences(ctx, userID)
```

### Webhooks

```go
// Register webhook for user
webhookConfig := notifications.WebhookConfig{
    URL:           "https://api.algotrader.com/webhook",
    Secret:        "webhook-secret-key",
    ContentType:   "application/json",
    Timeout:       30 * time.Second,
    RetryAttempts: 3,
    Headers: map[string]string{
        "X-API-Key": "algo-trader-api-key",
    },
}

webhookNotifier.RegisterWebhook(userID, webhookConfig)

// Webhook payload will be automatically sent
// Verify incoming webhooks:
isValid := notifications.VerifySignature(payload, signature, secret)
```

## Notification Types

### Trading Events
- `order_executed` - Order filled
- `position_closed` - Position closed with P&L
- `margin_call_warning` - Margin level below threshold
- `stop_out` - Position auto-closed due to insufficient margin
- `price_movement` - Significant price change alert

### Account Events
- `balance_change` - Account balance changed
- `deposit_received` - Deposit confirmed
- `withdrawal_complete` - Withdrawal processed

### Security Events
- `login_new_device` - Login from unrecognized device
- `password_changed` - Password updated
- `security_alert` - Suspicious activity detected

### System Events
- `trading_hours_change` - Market hours updated
- `system_maintenance` - Scheduled maintenance
- `news_alert` - Important market news

## Rate Limits (Default)

| Channel | Per Minute | Per Hour | Per Day |
|---------|-----------|----------|---------|
| Email   | 5         | 20       | 100     |
| SMS     | 2         | 10       | 30      |
| Push    | 10        | 50       | 200     |
| Webhook | 10        | 100      | 1000    |
| In-App  | 20        | 100      | 500     |

## Email Templates

Pre-built templates included:
- Welcome email
- Order confirmation
- Position closed
- Margin call warning
- Daily/weekly summaries
- Monthly statements

Load custom templates:
```go
smtpProvider := notifications.NewSMTPProvider(smtpConfig)
err := smtpProvider.LoadTemplate("custom-template", htmlContent)
```

## Retry Configuration

Default retry behavior:
- Max attempts: 5
- Initial delay: 1 second
- Max delay: 5 minutes
- Backoff factor: 2.0 (exponential)

Customize:
```go
retryQueue := notifications.NewRetryQueue(manager, deliveryStore)
retryQueue.config = notifications.RetryConfig{
    MaxAttempts:   3,
    InitialDelay:  2 * time.Second,
    MaxDelay:      10 * time.Minute,
    BackoffFactor: 1.5,
}
```

## Monitoring

```go
// Get delivery history
history, err := manager.GetDeliveryHistory(ctx, userID, 50)

// Check notification status
records, err := manager.GetNotificationStatus(ctx, notificationID)

// Get rate limit usage
perMin, perHour, perDay := rateLimiter.GetUsage(userID, channel)

// Get retry queue size
queueSize := retryQueue.GetQueueSize()
```

## Production Considerations

1. **Use persistent stores** - Replace in-memory stores with Redis/PostgreSQL
2. **Configure email provider** - Use SendGrid or AWS SES instead of SMTP
3. **Set up monitoring** - Track delivery rates, failures, retry queue size
4. **Implement logging** - Log all notification events for audit
5. **Add metrics** - Instrument with Prometheus/StatsD
6. **Configure alerts** - Alert on high failure rates or queue buildup
7. **Test templates** - A/B test email templates for engagement
8. **Localization** - Add translation support for multiple languages
9. **Compliance** - Ensure GDPR/CAN-SPAM compliance
10. **Security** - Rotate webhook secrets regularly

## Future Enhancements

- Telegram bot integration
- WhatsApp Business API
- Slack integration
- Discord webhooks
- Custom template engine
- Advanced batching rules
- Machine learning for send-time optimization
- Real-time delivery status updates via WebSocket
- Multi-language templates
- Template preview/testing tool
