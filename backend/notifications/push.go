package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// PushProvider interface for push notification services
type PushProvider interface {
	Send(ctx context.Context, tokens []string, title, body string, data map[string]interface{}) (string, error)
	SendToTopic(ctx context.Context, topic, title, body string, data map[string]interface{}) (string, error)
}

// FCMConfig holds Firebase Cloud Messaging configuration
type FCMConfig struct {
	ServerKey string
	ProjectID string
	BaseURL   string
}

// FCMProvider implements push notifications via Firebase Cloud Messaging
type FCMProvider struct {
	config     FCMConfig
	httpClient *http.Client
}

// NewFCMProvider creates a new FCM push provider
func NewFCMProvider(config FCMConfig) *FCMProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://fcm.googleapis.com/fcm/send"
	}

	return &FCMProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// fcmMessage represents an FCM message payload
type fcmMessage struct {
	To               string                 `json:"to,omitempty"`
	RegistrationIDs  []string               `json:"registration_ids,omitempty"`
	Priority         string                 `json:"priority"`
	Notification     fcmNotification        `json:"notification"`
	Data             map[string]interface{} `json:"data,omitempty"`
	TimeToLive       int                    `json:"time_to_live,omitempty"`
	CollapseKey      string                 `json:"collapse_key,omitempty"`
}

type fcmNotification struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	Icon        string `json:"icon,omitempty"`
	Sound       string `json:"sound,omitempty"`
	ClickAction string `json:"click_action,omitempty"`
	Badge       int    `json:"badge,omitempty"`
}

// fcmResponse represents FCM API response
type fcmResponse struct {
	MulticastID  int64 `json:"multicast_id"`
	Success      int   `json:"success"`
	Failure      int   `json:"failure"`
	CanonicalIDs int   `json:"canonical_ids"`
	Results      []struct {
		MessageID      string `json:"message_id,omitempty"`
		RegistrationID string `json:"registration_id,omitempty"`
		Error          string `json:"error,omitempty"`
	} `json:"results"`
}

// Send sends a push notification to specific device tokens
func (p *FCMProvider) Send(ctx context.Context, tokens []string, title, body string, data map[string]interface{}) (string, error) {
	message := fcmMessage{
		RegistrationIDs: tokens,
		Priority:        "high",
		Notification: fcmNotification{
			Title: title,
			Body:  body,
			Sound: "default",
		},
		Data:        data,
		TimeToLive:  86400, // 24 hours
	}

	return p.sendMessage(ctx, message)
}

// SendToTopic sends a push notification to a topic
func (p *FCMProvider) SendToTopic(ctx context.Context, topic, title, body string, data map[string]interface{}) (string, error) {
	message := fcmMessage{
		To:       fmt.Sprintf("/topics/%s", topic),
		Priority: "high",
		Notification: fcmNotification{
			Title: title,
			Body:  body,
			Sound: "default",
		},
		Data:        data,
		TimeToLive:  86400,
	}

	return p.sendMessage(ctx, message)
}

// sendMessage sends the FCM message
func (p *FCMProvider) sendMessage(ctx context.Context, message fcmMessage) (string, error) {
	payload, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.BaseURL, strings.NewReader(string(payload)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("key=%s", p.config.ServerKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("FCM API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var fcmResp fcmResponse
	if err := json.Unmarshal(body, &fcmResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if fcmResp.Failure > 0 {
		// Get first error
		for _, result := range fcmResp.Results {
			if result.Error != "" {
				return "", fmt.Errorf("FCM error: %s", result.Error)
			}
		}
	}

	return fmt.Sprintf("%d", fcmResp.MulticastID), nil
}

// PushNotifier handles push notifications
type PushNotifier struct {
	provider    PushProvider
	rateLimiter *RateLimiter
}

// NewPushNotifier creates a new push notifier
func NewPushNotifier(provider PushProvider, rateLimiter *RateLimiter) *PushNotifier {
	return &PushNotifier{
		provider:    provider,
		rateLimiter: rateLimiter,
	}
}

// Send sends a push notification
func (n *PushNotifier) Send(ctx context.Context, notif *Notification, deviceTokens []string) (*DeliveryRecord, error) {
	// Check rate limit
	if !n.rateLimiter.Allow(notif.UserID, ChannelPush) {
		return nil, fmt.Errorf("rate limit exceeded for user %s", notif.UserID)
	}

	record := &DeliveryRecord{
		ID:             generateID(),
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Channel:        ChannelPush,
		Status:         StatusPending,
		Attempts:       1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	now := time.Now()
	record.LastAttemptAt = &now

	messageID, err := n.provider.Send(ctx, deviceTokens, notif.Subject, notif.Message, notif.Data)
	if err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		return record, err
	}

	record.Status = StatusSent
	record.ProviderID = messageID
	record.DeliveredAt = &now

	return record, nil
}

// SendToTopic sends a push notification to a topic (e.g., all users subscribed to price alerts)
func (n *PushNotifier) SendToTopic(ctx context.Context, notif *Notification, topic string) (*DeliveryRecord, error) {
	record := &DeliveryRecord{
		ID:             generateID(),
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Channel:        ChannelPush,
		Status:         StatusPending,
		Attempts:       1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	now := time.Now()
	record.LastAttemptAt = &now

	messageID, err := n.provider.SendToTopic(ctx, topic, notif.Subject, notif.Message, notif.Data)
	if err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		return record, err
	}

	record.Status = StatusSent
	record.ProviderID = messageID
	record.DeliveredAt = &now

	return record, nil
}

// Push notification payload builders

// BuildOrderExecutedPush builds push notification for order executed
func BuildOrderExecutedPush(orderID, symbol, side string, quantity, price float64) map[string]interface{} {
	return map[string]interface{}{
		"type":     "order_executed",
		"order_id": orderID,
		"symbol":   symbol,
		"side":     side,
		"quantity": quantity,
		"price":    price,
		"action":   "open_order_details",
	}
}

// BuildMarginCallPush builds push notification for margin call
func BuildMarginCallPush(marginLevel, required float64) map[string]interface{} {
	return map[string]interface{}{
		"type":           "margin_call",
		"margin_level":   marginLevel,
		"required_level": required,
		"action":         "open_account",
		"priority":       "critical",
	}
}

// BuildPriceAlertPush builds push notification for price alert
func BuildPriceAlertPush(symbol string, price, changePercent float64, alertID string) map[string]interface{} {
	return map[string]interface{}{
		"type":           "price_alert",
		"symbol":         symbol,
		"price":          price,
		"change_percent": changePercent,
		"alert_id":       alertID,
		"action":         "open_chart",
	}
}

// BuildSecurityAlertPush builds push notification for security alert
func BuildSecurityAlertPush(alertType, device, location string) map[string]interface{} {
	return map[string]interface{}{
		"type":       "security_alert",
		"alert_type": alertType,
		"device":     device,
		"location":   location,
		"action":     "open_security_settings",
		"priority":   "high",
	}
}
