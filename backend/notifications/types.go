package notifications

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// Trading Events
	NotifOrderExecuted      NotificationType = "order_executed"
	NotifPositionClosed     NotificationType = "position_closed"
	NotifMarginCallWarning  NotificationType = "margin_call_warning"
	NotifStopOut            NotificationType = "stop_out"
	NotifPriceMovement      NotificationType = "price_movement"

	// Account Events
	NotifBalanceChange      NotificationType = "balance_change"
	NotifDepositReceived    NotificationType = "deposit_received"
	NotifWithdrawalComplete NotificationType = "withdrawal_complete"

	// Security Events
	NotifLoginNewDevice     NotificationType = "login_new_device"
	NotifPasswordChanged    NotificationType = "password_changed"
	NotifSecurityAlert      NotificationType = "security_alert"

	// System Events
	NotifTradingHoursChange NotificationType = "trading_hours_change"
	NotifSystemMaintenance  NotificationType = "system_maintenance"
	NotifNewsAlert          NotificationType = "news_alert"
)

// NotificationChannel represents delivery channel
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelSMS      NotificationChannel = "sms"
	ChannelPush     NotificationChannel = "push"
	ChannelInApp    NotificationChannel = "in_app"
	ChannelWebhook  NotificationChannel = "webhook"
	ChannelTelegram NotificationChannel = "telegram"
)

// Priority levels for notifications
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityNormal   Priority = "normal"
	PriorityLow      Priority = "low"
)

// DeliveryStatus tracks notification delivery
type DeliveryStatus string

const (
	StatusPending   DeliveryStatus = "pending"
	StatusSent      DeliveryStatus = "sent"
	StatusDelivered DeliveryStatus = "delivered"
	StatusFailed    DeliveryStatus = "failed"
	StatusRetrying  DeliveryStatus = "retrying"
	StatusCancelled DeliveryStatus = "cancelled"
)

// Notification represents a notification to be sent
type Notification struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Type        NotificationType       `json:"type"`
	Priority    Priority               `json:"priority"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Channels    []NotificationChannel  `json:"channels"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// DeliveryRecord tracks the delivery of a notification through a specific channel
type DeliveryRecord struct {
	ID             string              `json:"id"`
	NotificationID string              `json:"notification_id"`
	UserID         string              `json:"user_id"`
	Channel        NotificationChannel `json:"channel"`
	Status         DeliveryStatus      `json:"status"`
	Attempts       int                 `json:"attempts"`
	LastAttemptAt  *time.Time          `json:"last_attempt_at,omitempty"`
	DeliveredAt    *time.Time          `json:"delivered_at,omitempty"`
	Error          string              `json:"error,omitempty"`
	ProviderID     string              `json:"provider_id,omitempty"` // External provider message ID
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

// UserPreferences stores notification preferences per user
type UserPreferences struct {
	UserID       string                         `json:"user_id"`
	Preferences  map[NotificationType]ChannelPreference `json:"preferences"`
	QuietHours   *QuietHours                    `json:"quiet_hours,omitempty"`
	Locale       string                         `json:"locale"`
	Timezone     string                         `json:"timezone"`
	UnsubscribeAll bool                          `json:"unsubscribe_all"`
	UpdatedAt    time.Time                      `json:"updated_at"`
}

// ChannelPreference defines which channels are enabled for a notification type
type ChannelPreference struct {
	Enabled  bool                    `json:"enabled"`
	Channels []NotificationChannel   `json:"channels"`
	MinimumPriority Priority         `json:"minimum_priority"`
}

// QuietHours defines when to suppress non-critical notifications
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	Timezone  string `json:"timezone"`
}

// Template represents a notification template
type Template struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        NotificationType       `json:"type"`
	Channel     NotificationChannel    `json:"channel"`
	Locale      string                 `json:"locale"`
	Subject     string                 `json:"subject,omitempty"`
	Body        string                 `json:"body"`
	HTMLBody    string                 `json:"html_body,omitempty"`
	Variables   []string               `json:"variables"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsActive    bool                   `json:"is_active"`
	Version     int                    `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// BatchNotification groups multiple notifications for efficient delivery
type BatchNotification struct {
	ID            string          `json:"id"`
	UserID        string          `json:"user_id"`
	Channel       NotificationChannel `json:"channel"`
	Notifications []Notification  `json:"notifications"`
	CreatedAt     time.Time       `json:"created_at"`
	ScheduledAt   time.Time       `json:"scheduled_at"`
}

// RateLimitConfig defines rate limiting rules
type RateLimitConfig struct {
	Channel     NotificationChannel `json:"channel"`
	MaxPerMinute int               `json:"max_per_minute"`
	MaxPerHour   int               `json:"max_per_hour"`
	MaxPerDay    int               `json:"max_per_day"`
}

// RetryConfig defines retry behavior for failed notifications
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}
