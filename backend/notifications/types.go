package notifications

import (
	"context"
	"time"
)

// Severity levels for notifications
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Category types for notifications
type Category string

const (
	CategoryTrading  Category = "trading"
	CategoryAccount  Category = "account"
	CategorySecurity Category = "security"
	CategorySystem   Category = "system"
)

// NotificationChannel represents delivery channels
type NotificationChannel string

const (
	ChannelWebSocket   NotificationChannel = "websocket"
	ChannelEmail       NotificationChannel = "email"
	ChannelBrowserPush NotificationChannel = "browser_push"
	ChannelSMS         NotificationChannel = "sms"
	ChannelPush        NotificationChannel = "push"
	ChannelWebhook     NotificationChannel = "webhook"
	ChannelInApp       NotificationChannel = "in_app"
)

// DeliveryStatus represents the status of a notification delivery attempt
type DeliveryStatus string

const (
	StatusPending   DeliveryStatus = "pending"
	StatusSent      DeliveryStatus = "sent"
	StatusDelivered DeliveryStatus = "delivered"
	StatusFailed    DeliveryStatus = "failed"
	StatusRetrying  DeliveryStatus = "retrying"
)

// Priority represents the priority level of a notification
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// NotificationType is a type alias for notification type strings
type NotificationType = string

// Notification type constants
const (
	NotifMarginCallWarning NotificationType = "margin_call_warning"
	NotifStopOut           NotificationType = "stop_out"
	NotifSecurityAlert     NotificationType = "security_alert"
	NotifOrderExecuted     NotificationType = "order_executed"
	NotifPositionClosed    NotificationType = "position_closed"
	NotifLoginNewDevice    NotificationType = "login_new_device"
	NotifBalanceChange     NotificationType = "balance_change"
	NotifPriceMovement     NotificationType = "price_movement"
	NotifNewsAlert         NotificationType = "news_alert"
	NotifTradingHoursChange NotificationType = "trading_hours_change"
	NotifSystemMaintenance NotificationType = "system_maintenance"
)

// Notification represents a user notification
type Notification struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"userId"`
	Type        NotificationType       `json:"type"` // e.g., "order_filled", "margin_warning", "login_alert"
	Severity    Severity               `json:"severity"`
	Category    Category               `json:"category"`
	Priority    Priority               `json:"priority,omitempty"`
	Title       string                 `json:"title"`
	Subject     string                 `json:"subject,omitempty"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`        // Additional structured data
	ActionItems []ActionItem           `json:"actionItems,omitempty"` // Optional actions user can take
	Channels    []NotificationChannel  `json:"channels,omitempty"`    // Delivery channels
	Read        bool                   `json:"read"`
	CreatedAt   int64                  `json:"createdAt"` // Unix timestamp
}

// ActionItem represents an actionable item in a notification
type ActionItem struct {
	Label string `json:"label"` // e.g., "View Order", "Add Funds"
	URL   string `json:"url"`   // e.g., "/orders/123", "/account/deposit"
	Type  string `json:"type"`  // e.g., "primary", "secondary", "danger"
}

// DeliveryRecord tracks notification delivery across channels
type DeliveryRecord struct {
	ID             string              `json:"id"`
	NotificationID string              `json:"notificationId"`
	UserID         string              `json:"userId"`
	Channel        NotificationChannel `json:"channel"`
	Status         DeliveryStatus      `json:"status"`
	Attempts       int                 `json:"attempts"`
	Error          string              `json:"error,omitempty"`
	ProviderID     string              `json:"providerId,omitempty"`
	SentAt         int64               `json:"sentAt,omitempty"`          // Unix timestamp (legacy)
	DeliveredAt    *time.Time          `json:"deliveredAt,omitempty"`
	LastAttemptAt  *time.Time          `json:"lastAttemptAt,omitempty"`
	CreatedAt      time.Time           `json:"createdAt_record"`
	UpdatedAt      time.Time           `json:"updatedAt"`
}

// RateLimitConfig holds rate limit configuration for a notification channel
type RateLimitConfig struct {
	Channel      NotificationChannel
	MaxPerMinute int
	MaxPerHour   int
	MaxPerDay    int
}

// RetryConfig holds configuration for notification retry logic
type RetryConfig struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffFactor   float64
	RetryableErrors []string
}

// DeliveryStore interface for persisting delivery records
type DeliveryStore interface {
	Save(ctx context.Context, record *DeliveryRecord) error
}

// UserPreferences holds a user's notification preferences
type UserPreferences struct {
	UserID         string                                `json:"userId"`
	Preferences    map[NotificationType]ChannelPreference `json:"preferences"`
	Locale         string                                `json:"locale"`
	Timezone       string                                `json:"timezone"`
	UnsubscribeAll bool                                  `json:"unsubscribeAll"`
	QuietHours     *QuietHours                           `json:"quietHours,omitempty"`
	UpdatedAt      time.Time                             `json:"updatedAt"`
}

// ChannelPreference defines how a notification type should be delivered
type ChannelPreference struct {
	Enabled         bool                  `json:"enabled"`
	Channels        []NotificationChannel `json:"channels"`
	MinimumPriority Priority              `json:"minimumPriority"`
}

// QuietHours defines a period during which non-critical notifications are suppressed
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"startTime"` // HH:MM format
	EndTime   string `json:"endTime"`   // HH:MM format
	Timezone  string `json:"timezone"`
}

// UserContacts holds contact information for a user across channels
type UserContacts struct {
	UserID       string   `json:"userId"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone,omitempty"`
	DeviceTokens []string `json:"deviceTokens,omitempty"`
}

// NotificationTemplate for common notification types
type NotificationTemplate struct {
	Type     string
	Severity Severity
	Category Category
	Title    string
	Message  string // Can contain placeholders like {symbol}, {price}
}

// Common notification templates
var Templates = map[string]NotificationTemplate{
	"order_filled": {
		Type:     "order_filled",
		Severity: SeverityInfo,
		Category: CategoryTrading,
		Title:    "Order Filled",
		Message:  "Your {orderType} order for {volume} lots of {symbol} was filled at {price}",
	},
	"margin_warning": {
		Type:     "margin_warning",
		Severity: SeverityWarning,
		Category: CategoryAccount,
		Title:    "Margin Level Warning",
		Message:  "Your margin level is at {level}%. Please add funds to avoid liquidation.",
	},
	"margin_call": {
		Type:     "margin_call",
		Severity: SeverityCritical,
		Category: CategoryAccount,
		Title:    "Margin Call",
		Message:  "URGENT: Your margin level is critically low at {level}%. Add funds immediately to prevent position closure.",
	},
	"login_alert": {
		Type:     "login_alert",
		Severity: SeverityWarning,
		Category: CategorySecurity,
		Title:    "New Login Detected",
		Message:  "Login from {location} at {time}. If this wasn't you, secure your account immediately.",
	},
	"position_closed": {
		Type:     "position_closed",
		Severity: SeverityInfo,
		Category: CategoryTrading,
		Title:    "Position Closed",
		Message:  "Your {symbol} position was closed. P/L: {pnl}",
	},
	"stop_loss_triggered": {
		Type:     "stop_loss_triggered",
		Severity: SeverityWarning,
		Category: CategoryTrading,
		Title:    "Stop Loss Triggered",
		Message:  "Stop loss triggered for {symbol} at {price}. Position closed with P/L: {pnl}",
	},
	"take_profit_triggered": {
		Type:     "take_profit_triggered",
		Severity: SeverityInfo,
		Category: CategoryTrading,
		Title:    "Take Profit Triggered",
		Message:  "Take profit triggered for {symbol} at {price}. Position closed with P/L: {pnl}",
	},
}
