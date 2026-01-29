package alerts

import "time"

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeThreshold AlertType = "threshold"
	AlertTypeAnomaly   AlertType = "anomaly"
	AlertTypePattern   AlertType = "pattern"
)

// AlertSeverity represents the urgency of the alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "LOW"
	AlertSeverityMedium   AlertSeverity = "MEDIUM"
	AlertSeverityHigh     AlertSeverity = "HIGH"
	AlertSeverityCritical AlertSeverity = "CRITICAL"
)

// AlertStatus represents the lifecycle state of an alert
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSnoozed      AlertStatus = "snoozed"
)

// Alert represents a triggered alert event
type Alert struct {
	ID          string        `json:"id"`
	RuleID      string        `json:"ruleId,omitempty"`
	AccountID   string        `json:"accountId"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Status      AlertStatus   `json:"status"`
	Title       string        `json:"title"`
	Message     string        `json:"message"`
	Metric      string        `json:"metric"`
	Value       float64       `json:"value"`
	Threshold   float64       `json:"threshold,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	AckedBy     string        `json:"ackedBy,omitempty"`
	AckedAt     *time.Time    `json:"ackedAt,omitempty"`
	ResolvedAt  *time.Time    `json:"resolvedAt,omitempty"`
	SnoozedUntil *time.Time   `json:"snoozedUntil,omitempty"`
	Fingerprint string        `json:"fingerprint"` // For deduplication
}

// AlertRule represents a configured alert rule
type AlertRule struct {
	ID          string        `json:"id"`
	AccountID   string        `json:"accountId"` // Empty for system-wide rules
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Type        AlertType     `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Enabled     bool          `json:"enabled"`

	// Threshold configuration
	Metric      string  `json:"metric"`
	Operator    string  `json:"operator"` // ">", "<", ">=", "<=", "=="
	Threshold   float64 `json:"threshold"`

	// Anomaly detection config
	ZScoreThreshold float64 `json:"zScoreThreshold,omitempty"` // Default: 3.0
	LookbackPeriod  int     `json:"lookbackPeriod,omitempty"`  // Samples for baseline (default: 100)

	// Pattern detection config
	Pattern         string `json:"pattern,omitempty"`         // e.g., "consecutive_failures"
	PatternCount    int    `json:"patternCount,omitempty"`    // e.g., 5 consecutive events
	PatternWindow   int    `json:"patternWindow,omitempty"`   // Time window in seconds

	// Notification channels
	Channels        []string `json:"channels"`        // ["dashboard", "email", "sms"]
	CooldownSeconds int      `json:"cooldownSeconds"` // Minimum time between alerts

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// MetricSnapshot represents a point-in-time account metric
type MetricSnapshot struct {
	AccountID      string    `json:"accountId"`
	Timestamp      time.Time `json:"timestamp"`
	Balance        float64   `json:"balance"`
	Equity         float64   `json:"equity"`
	Margin         float64   `json:"margin"`
	FreeMargin     float64   `json:"freeMargin"`
	MarginLevel    float64   `json:"marginLevel"`
	ExposurePercent float64  `json:"exposurePercent"`
	PositionCount  int       `json:"positionCount"`
	PnL            float64   `json:"pnl"`
}

// NotificationChannel represents a delivery mechanism
type NotificationChannel string

const (
	ChannelDashboard NotificationChannel = "dashboard"
	ChannelEmail     NotificationChannel = "email"
	ChannelSMS       NotificationChannel = "sms"
	ChannelWebhook   NotificationChannel = "webhook"
)

// Notification represents a pending notification to be sent
type Notification struct {
	ID        string              `json:"id"`
	AlertID   string              `json:"alertId"`
	AccountID string              `json:"accountID"`
	Channel   NotificationChannel `json:"channel"`
	To        string              `json:"to"` // Email address, phone number, or webhook URL
	Subject   string              `json:"subject,omitempty"`
	Body      string              `json:"body"`
	SentAt    *time.Time          `json:"sentAt,omitempty"`
	Error     string              `json:"error,omitempty"`
	Retries   int                 `json:"retries"`
	CreatedAt time.Time           `json:"createdAt"`
}

// AccountMetrics provides interface to retrieve account data
type AccountMetrics interface {
	GetSnapshot(accountID string) (*MetricSnapshot, error)
}
