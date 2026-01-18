package notifications

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookConfig holds webhook configuration
type WebhookConfig struct {
	URL           string
	Secret        string
	ContentType   string
	Timeout       time.Duration
	RetryAttempts int
	Headers       map[string]string
}

// WebhookProvider implements webhook notifications
type WebhookProvider struct {
	config     WebhookConfig
	httpClient *http.Client
}

// NewWebhookProvider creates a new webhook provider
func NewWebhookProvider(config WebhookConfig) *WebhookProvider {
	if config.ContentType == "" {
		config.ContentType = "application/json"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	return &WebhookProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// WebhookPayload represents the webhook payload structure
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp int64                  `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

// Send sends a webhook notification
func (p *WebhookProvider) Send(ctx context.Context, event string, data map[string]interface{}, userID string) (string, error) {
	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now().Unix(),
		UserID:    userID,
		Data:      data,
	}

	return p.sendWithRetry(ctx, payload)
}

// sendWithRetry sends the webhook with retry logic
func (p *WebhookProvider) sendWithRetry(ctx context.Context, payload WebhookPayload) (string, error) {
	var lastErr error

	for attempt := 0; attempt < p.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}

		messageID, err := p.sendRequest(ctx, payload)
		if err == nil {
			return messageID, nil
		}

		lastErr = err
	}

	return "", fmt.Errorf("webhook failed after %d attempts: %w", p.config.RetryAttempts, lastErr)
}

// sendRequest sends a single webhook request
func (p *WebhookProvider) sendRequest(ctx context.Context, payload WebhookPayload) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", p.config.ContentType)
	req.Header.Set("User-Agent", "TradingPlatform-Webhook/1.0")

	// Add custom headers
	for key, value := range p.config.Headers {
		req.Header.Set(key, value)
	}

	// Add HMAC signature if secret is configured
	if p.config.Secret != "" {
		signature := p.generateSignature(payloadBytes)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Add timestamp
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", payload.Timestamp))

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	// Generate message ID from response or timestamp
	messageID := resp.Header.Get("X-Message-ID")
	if messageID == "" {
		messageID = fmt.Sprintf("%d-%s", payload.Timestamp, payload.Event)
	}

	return messageID, nil
}

// generateSignature creates HMAC-SHA256 signature for webhook payload
func (p *WebhookProvider) generateSignature(payload []byte) string {
	h := hmac.New(sha256.New, []byte(p.config.Secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies webhook signature from incoming webhooks
func VerifySignature(payload []byte, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// WebhookNotifier handles webhook notifications
type WebhookNotifier struct {
	providers   map[string]*WebhookProvider
	rateLimiter *RateLimiter
}

// NewWebhookNotifier creates a new webhook notifier
func NewWebhookNotifier(rateLimiter *RateLimiter) *WebhookNotifier {
	return &WebhookNotifier{
		providers:   make(map[string]*WebhookProvider),
		rateLimiter: rateLimiter,
	}
}

// RegisterWebhook registers a webhook endpoint for a user
func (n *WebhookNotifier) RegisterWebhook(userID string, config WebhookConfig) {
	n.providers[userID] = NewWebhookProvider(config)
}

// UnregisterWebhook removes a webhook endpoint
func (n *WebhookNotifier) UnregisterWebhook(userID string) {
	delete(n.providers, userID)
}

// Send sends a webhook notification
func (n *WebhookNotifier) Send(ctx context.Context, notif *Notification) (*DeliveryRecord, error) {
	provider, exists := n.providers[notif.UserID]
	if !exists {
		return nil, fmt.Errorf("no webhook configured for user %s", notif.UserID)
	}

	// Check rate limit
	if !n.rateLimiter.Allow(notif.UserID, ChannelWebhook) {
		return nil, fmt.Errorf("rate limit exceeded for user %s", notif.UserID)
	}

	record := &DeliveryRecord{
		ID:             generateID(),
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Channel:        ChannelWebhook,
		Status:         StatusPending,
		Attempts:       1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	now := time.Now()
	record.LastAttemptAt = &now

	messageID, err := provider.Send(ctx, string(notif.Type), notif.Data, notif.UserID)
	if err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		return record, err
	}

	record.Status = StatusDelivered
	record.ProviderID = messageID
	record.DeliveredAt = &now

	return record, nil
}

// Webhook event builders for different notification types

// BuildOrderExecutedWebhook builds webhook payload for order executed
func BuildOrderExecutedWebhook(orderID, symbol, side string, quantity, price float64, executedAt time.Time) map[string]interface{} {
	return map[string]interface{}{
		"order_id":    orderID,
		"symbol":      symbol,
		"side":        side,
		"quantity":    quantity,
		"price":       price,
		"executed_at": executedAt.Format(time.RFC3339),
	}
}

// BuildPositionClosedWebhook builds webhook payload for position closed
func BuildPositionClosedWebhook(positionID, symbol string, pnl, closedPrice float64, closedAt time.Time) map[string]interface{} {
	return map[string]interface{}{
		"position_id":  positionID,
		"symbol":       symbol,
		"pnl":          pnl,
		"closed_price": closedPrice,
		"closed_at":    closedAt.Format(time.RFC3339),
	}
}

// BuildMarginCallWebhook builds webhook payload for margin call
func BuildMarginCallWebhook(accountID string, marginLevel, equity, usedMargin, freeMargin float64) map[string]interface{} {
	return map[string]interface{}{
		"account_id":   accountID,
		"margin_level": marginLevel,
		"equity":       equity,
		"used_margin":  usedMargin,
		"free_margin":  freeMargin,
		"severity":     "critical",
	}
}

// BuildPriceAlertWebhook builds webhook payload for price alert
func BuildPriceAlertWebhook(symbol string, price, changePercent float64, alertID string, triggeredAt time.Time) map[string]interface{} {
	return map[string]interface{}{
		"alert_id":       alertID,
		"symbol":         symbol,
		"current_price":  price,
		"change_percent": changePercent,
		"triggered_at":   triggeredAt.Format(time.RFC3339),
	}
}

// BuildBalanceChangeWebhook builds webhook payload for balance change
func BuildBalanceChangeWebhook(accountID, changeType string, amount, previousBalance, newBalance float64) map[string]interface{} {
	return map[string]interface{}{
		"account_id":       accountID,
		"change_type":      changeType,
		"amount":           amount,
		"previous_balance": previousBalance,
		"new_balance":      newBalance,
	}
}

// BuildSecurityAlertWebhook builds webhook payload for security alert
func BuildSecurityAlertWebhook(alertType, ipAddress, device, location string, timestamp time.Time) map[string]interface{} {
	return map[string]interface{}{
		"alert_type": alertType,
		"ip_address": ipAddress,
		"device":     device,
		"location":   location,
		"timestamp":  timestamp.Format(time.RFC3339),
		"severity":   "high",
	}
}
