package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NotificationManager coordinates all notification channels
type NotificationManager struct {
	emailNotifier     *EmailNotifier
	smsNotifier       *SMSNotifier
	pushNotifier      *PushNotifier
	webhookNotifier   *WebhookNotifier
	prefsManager      *PreferencesManager
	deliveryStore     DeliveryStore
	retryQueue        *RetryQueue
	batchProcessor    *BatchProcessor
	rateLimiter       *RateLimiter
}

// DeliveryStore interface for storing delivery records
type DeliveryStore interface {
	Save(ctx context.Context, record *DeliveryRecord) error
	Get(ctx context.Context, id string) (*DeliveryRecord, error)
	GetByNotification(ctx context.Context, notificationID string) ([]*DeliveryRecord, error)
	GetByUser(ctx context.Context, userID string, limit int) ([]*DeliveryRecord, error)
	UpdateStatus(ctx context.Context, id string, status DeliveryStatus) error
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(
	emailNotifier *EmailNotifier,
	smsNotifier *SMSNotifier,
	pushNotifier *PushNotifier,
	webhookNotifier *WebhookNotifier,
	prefsManager *PreferencesManager,
	deliveryStore DeliveryStore,
	rateLimiter *RateLimiter,
) *NotificationManager {
	nm := &NotificationManager{
		emailNotifier:   emailNotifier,
		smsNotifier:     smsNotifier,
		pushNotifier:    pushNotifier,
		webhookNotifier: webhookNotifier,
		prefsManager:    prefsManager,
		deliveryStore:   deliveryStore,
		rateLimiter:     rateLimiter,
	}

	nm.retryQueue = NewRetryQueue(nm, deliveryStore)
	nm.batchProcessor = NewBatchProcessor(nm)

	return nm
}

// Send sends a notification through all appropriate channels
func (m *NotificationManager) Send(ctx context.Context, notif *Notification, userContacts UserContacts) error {
	// Get user's preferred channels
	channels, err := m.prefsManager.GetChannelsForNotification(ctx, notif.UserID, notif)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	// Send through each channel concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(channels))

	for _, channel := range channels {
		wg.Add(1)
		go func(ch NotificationChannel) {
			defer wg.Done()

			record, err := m.sendToChannel(ctx, notif, ch, userContacts)
			if err != nil {
				errChan <- fmt.Errorf("failed to send via %s: %w", ch, err)
				return
			}

			// Save delivery record
			if err := m.deliveryStore.Save(ctx, record); err != nil {
				errChan <- fmt.Errorf("failed to save delivery record: %w", err)
				return
			}

			// If failed, add to retry queue
			if record.Status == StatusFailed {
				m.retryQueue.Add(record)
			}
		}(channel)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification partially failed: %v", errors)
	}

	return nil
}

// sendToChannel sends a notification to a specific channel
func (m *NotificationManager) sendToChannel(ctx context.Context, notif *Notification, channel NotificationChannel, contacts UserContacts) (*DeliveryRecord, error) {
	switch channel {
	case ChannelEmail:
		if contacts.Email == "" {
			return nil, fmt.Errorf("no email address for user")
		}
		return m.emailNotifier.Send(ctx, notif, contacts.Email)

	case ChannelSMS:
		if contacts.Phone == "" {
			return nil, fmt.Errorf("no phone number for user")
		}
		return m.smsNotifier.Send(ctx, notif, contacts.Phone)

	case ChannelPush:
		if len(contacts.DeviceTokens) == 0 {
			return nil, fmt.Errorf("no device tokens for user")
		}
		return m.pushNotifier.Send(ctx, notif, contacts.DeviceTokens)

	case ChannelWebhook:
		return m.webhookNotifier.Send(ctx, notif)

	case ChannelInApp:
		// In-app notifications are handled separately (stored in database for UI to fetch)
		return m.createInAppRecord(notif), nil

	default:
		return nil, fmt.Errorf("unsupported channel: %s", channel)
	}
}

// createInAppRecord creates a delivery record for in-app notifications
func (m *NotificationManager) createInAppRecord(notif *Notification) *DeliveryRecord {
	now := time.Now()
	return &DeliveryRecord{
		ID:             generateID(),
		NotificationID: notif.ID,
		UserID:         notif.UserID,
		Channel:        ChannelInApp,
		Status:         StatusDelivered,
		Attempts:       1,
		CreatedAt:      now,
		UpdatedAt:      now,
		DeliveredAt:    &now,
	}
}

// SendBatch sends multiple notifications efficiently
func (m *NotificationManager) SendBatch(ctx context.Context, notifications []Notification, contacts map[string]UserContacts) error {
	return m.batchProcessor.ProcessBatch(ctx, notifications, contacts)
}

// GetDeliveryHistory returns delivery history for a user
func (m *NotificationManager) GetDeliveryHistory(ctx context.Context, userID string, limit int) ([]*DeliveryRecord, error) {
	return m.deliveryStore.GetByUser(ctx, userID, limit)
}

// GetNotificationStatus returns the delivery status of a notification
func (m *NotificationManager) GetNotificationStatus(ctx context.Context, notificationID string) ([]*DeliveryRecord, error) {
	return m.deliveryStore.GetByNotification(ctx, notificationID)
}

// RetryFailed retries failed notifications
func (m *NotificationManager) RetryFailed(ctx context.Context) error {
	return m.retryQueue.ProcessRetries(ctx)
}

// UserContacts holds user contact information
type UserContacts struct {
	Email        string
	Phone        string
	DeviceTokens []string
}

// Helper functions for creating notifications

// CreateOrderExecutedNotification creates an order executed notification
func CreateOrderExecutedNotification(userID, orderID, symbol, side string, quantity, price float64) *Notification {
	return &Notification{
		ID:       generateID(),
		UserID:   userID,
		Type:     NotifOrderExecuted,
		Priority: PriorityHigh,
		Subject:  "Order Executed",
		Message:  FormatOrderExecutedSMS(orderID, symbol, side, quantity, price),
		Data: map[string]interface{}{
			"order_id": orderID,
			"symbol":   symbol,
			"side":     side,
			"quantity": quantity,
			"price":    price,
		},
		Channels:  []NotificationChannel{ChannelEmail, ChannelSMS, ChannelPush, ChannelInApp},
		CreatedAt: time.Now(),
	}
}

// CreateMarginCallNotification creates a margin call notification
func CreateMarginCallNotification(userID, accountID string, marginLevel, required float64) *Notification {
	return &Notification{
		ID:       generateID(),
		UserID:   userID,
		Type:     NotifMarginCallWarning,
		Priority: PriorityCritical,
		Subject:  "⚠️ Margin Call Warning",
		Message:  FormatMarginCallSMS(marginLevel, required),
		Data: map[string]interface{}{
			"account_id":    accountID,
			"margin_level":  marginLevel,
			"required_level": required,
		},
		Channels:  []NotificationChannel{ChannelEmail, ChannelSMS, ChannelPush, ChannelInApp, ChannelWebhook},
		CreatedAt: time.Now(),
	}
}

// CreatePositionClosedNotification creates a position closed notification
func CreatePositionClosedNotification(userID, positionID, symbol string, pnl float64) *Notification {
	subject := "Position Closed"
	if pnl >= 0 {
		subject = "Position Closed - Profit"
	} else {
		subject = "Position Closed - Loss"
	}

	return &Notification{
		ID:       generateID(),
		UserID:   userID,
		Type:     NotifPositionClosed,
		Priority: PriorityHigh,
		Subject:  subject,
		Message:  fmt.Sprintf("Position %s (%s) closed with P&L: $%.2f", positionID[:8], symbol, pnl),
		Data: map[string]interface{}{
			"position_id": positionID,
			"symbol":      symbol,
			"pnl":         pnl,
		},
		Channels:  []NotificationChannel{ChannelEmail, ChannelPush, ChannelInApp, ChannelWebhook},
		CreatedAt: time.Now(),
	}
}

// CreateLoginAlertNotification creates a new device login notification
func CreateLoginAlertNotification(userID, device, location, ipAddress string) *Notification {
	return &Notification{
		ID:       generateID(),
		UserID:   userID,
		Type:     NotifLoginNewDevice,
		Priority: PriorityHigh,
		Subject:  "New Login Detected",
		Message:  FormatLoginNewDeviceSMS(device, location),
		Data: map[string]interface{}{
			"device":     device,
			"location":   location,
			"ip_address": ipAddress,
		},
		Channels:  []NotificationChannel{ChannelEmail, ChannelSMS, ChannelPush},
		CreatedAt: time.Now(),
	}
}

// CreatePriceAlertNotification creates a price alert notification
func CreatePriceAlertNotification(userID, symbol, alertID string, price, changePercent float64) *Notification {
	return &Notification{
		ID:       generateID(),
		UserID:   userID,
		Type:     NotifPriceMovement,
		Priority: PriorityNormal,
		Subject:  fmt.Sprintf("Price Alert: %s", symbol),
		Message:  FormatPriceAlertSMS(symbol, price, changePercent),
		Data: map[string]interface{}{
			"symbol":         symbol,
			"price":          price,
			"change_percent": changePercent,
			"alert_id":       alertID,
		},
		Channels:  []NotificationChannel{ChannelPush, ChannelInApp},
		CreatedAt: time.Now(),
	}
}
