package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SMSProvider interface for different SMS service providers
type SMSProvider interface {
	Send(ctx context.Context, to, message string) (string, error)
	GetDeliveryStatus(ctx context.Context, messageID string) (DeliveryStatus, error)
}

// TwilioConfig holds Twilio API configuration
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	BaseURL    string
}

// TwilioProvider implements SMS sending via Twilio
type TwilioProvider struct {
	config     TwilioConfig
	httpClient *http.Client
}

// NewTwilioProvider creates a new Twilio SMS provider
func NewTwilioProvider(config TwilioConfig) *TwilioProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.twilio.com/2010-04-01"
	}

	return &TwilioProvider{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// twilioResponse represents Twilio API response
type twilioResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Send sends an SMS via Twilio
func (p *TwilioProvider) Send(ctx context.Context, to, message string) (string, error) {
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages.json", p.config.BaseURL, p.config.AccountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", p.config.FromNumber)
	data.Set("Body", message)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(p.config.AccountSID, p.config.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var twilioResp twilioResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("twilio API error: %s (code: %d)", twilioResp.ErrorMessage, twilioResp.ErrorCode)
	}

	return twilioResp.SID, nil
}

// GetDeliveryStatus checks the delivery status of a message
func (p *TwilioProvider) GetDeliveryStatus(ctx context.Context, messageID string) (DeliveryStatus, error) {
	apiURL := fmt.Sprintf("%s/Accounts/%s/Messages/%s.json", p.config.BaseURL, p.config.AccountSID, messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return StatusFailed, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(p.config.AccountSID, p.config.AuthToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return StatusFailed, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return StatusFailed, fmt.Errorf("failed to read response: %w", err)
	}

	var twilioResp twilioResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return StatusFailed, fmt.Errorf("failed to parse response: %w", err)
	}

	return mapTwilioStatus(twilioResp.Status), nil
}

// mapTwilioStatus maps Twilio status to our DeliveryStatus
func mapTwilioStatus(twilioStatus string) DeliveryStatus {
	switch twilioStatus {
	case "queued", "accepted":
		return StatusPending
	case "sending", "sent":
		return StatusSent
	case "delivered":
		return StatusDelivered
	case "failed", "undelivered":
		return StatusFailed
	default:
		return StatusPending
	}
}

// SMSNotifier handles SMS notifications
type SMSNotifier struct {
	provider    SMSProvider
	rateLimiter *RateLimiter
	maxLength   int
}

// NewSMSNotifier creates a new SMS notifier
func NewSMSNotifier(provider SMSProvider, rateLimiter *RateLimiter) *SMSNotifier {
	return &SMSNotifier{
		provider:    provider,
		rateLimiter: rateLimiter,
		maxLength:   160, // Standard SMS length
	}
}

// Send sends an SMS notification
func (n *SMSNotifier) Send(ctx context.Context, notif *Notification, recipient string) (*DeliveryRecord, error) {
	// Check rate limit
	if !n.rateLimiter.Allow(notif.UserID, ChannelSMS) {
		return nil, fmt.Errorf("rate limit exceeded for user %s", notif.UserID)
	}

	record := &DeliveryRecord{
		ID:             generateID(),
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Channel:        ChannelSMS,
		Status:         StatusPending,
		Attempts:       1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Truncate message if too long
	message := notif.Message
	if len(message) > n.maxLength {
		message = message[:n.maxLength-3] + "..."
	}

	now := time.Now()
	record.LastAttemptAt = &now

	messageID, err := n.provider.Send(ctx, recipient, message)
	if err != nil {
		record.Status = StatusFailed
		record.Error = err.Error()
		return record, err
	}

	record.Status = StatusSent
	record.ProviderID = messageID

	return record, nil
}

// CheckStatus checks the delivery status of a sent SMS
func (n *SMSNotifier) CheckStatus(ctx context.Context, record *DeliveryRecord) error {
	if record.ProviderID == "" {
		return fmt.Errorf("no provider ID available")
	}

	status, err := n.provider.GetDeliveryStatus(ctx, record.ProviderID)
	if err != nil {
		return err
	}

	record.Status = status
	record.UpdatedAt = time.Now()

	if status == StatusDelivered {
		now := time.Now()
		record.DeliveredAt = &now
	}

	return nil
}

// SMS message templates for different notification types

// FormatOrderExecutedSMS formats an order executed notification for SMS
func FormatOrderExecutedSMS(orderID, symbol, side string, quantity, price float64) string {
	return fmt.Sprintf("Order %s executed: %s %s %.4f @ %.4f",
		orderID[:8], side, symbol, quantity, price)
}

// FormatMarginCallSMS formats a margin call warning for SMS
func FormatMarginCallSMS(marginLevel, required float64) string {
	return fmt.Sprintf("URGENT: Margin call! Your margin level is %.2f%% (required: %.2f%%). Please add funds or close positions.",
		marginLevel, required)
}

// FormatStopOutSMS formats a stop-out notification for SMS
func FormatStopOutSMS(positionID string, marginLevel float64) string {
	return fmt.Sprintf("STOP-OUT: Position %s closed. Margin level: %.2f%%. Contact support if you have questions.",
		positionID[:8], marginLevel)
}

// FormatLoginNewDeviceSMS formats a new device login alert for SMS
func FormatLoginNewDeviceSMS(device, location string) string {
	return fmt.Sprintf("New login detected from %s in %s. If this wasn't you, secure your account immediately.",
		device, location)
}

// FormatPriceAlertSMS formats a price movement alert for SMS
func FormatPriceAlertSMS(symbol string, price, changePercent float64) string {
	sign := "+"
	if changePercent < 0 {
		sign = ""
	}
	return fmt.Sprintf("Price Alert: %s is now %.4f (%s%.2f%%)",
		symbol, price, sign, changePercent)
}

// FormatBalanceChangeSMS formats a balance change notification for SMS
func FormatBalanceChangeSMS(changeType string, amount float64, newBalance float64) string {
	return fmt.Sprintf("%s: $%.2f. New balance: $%.2f",
		changeType, amount, newBalance)
}
