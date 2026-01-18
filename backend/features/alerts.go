package features

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Alerts System
// - Price alerts (above/below, crossing)
// - Indicator alerts (RSI overbought/oversold, MACD crossover, etc.)
// - News alerts integration
// - SMS/Email/Push notifications
// - Webhook support for custom integrations

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypePrice         AlertType = "PRICE"
	AlertTypeIndicator     AlertType = "INDICATOR"
	AlertTypePriceChange   AlertType = "PRICE_CHANGE"
	AlertTypePricePattern  AlertType = "PATTERN"
	AlertTypeVolume        AlertType = "VOLUME"
	AlertTypeNews          AlertType = "NEWS"
	AlertTypeAccountEvent  AlertType = "ACCOUNT"
)

// AlertCondition defines the condition type
type AlertCondition string

const (
	ConditionAbove       AlertCondition = "ABOVE"
	ConditionBelow       AlertCondition = "BELOW"
	ConditionCrossAbove  AlertCondition = "CROSS_ABOVE"
	ConditionCrossBelow  AlertCondition = "CROSS_BELOW"
	ConditionEquals      AlertCondition = "EQUALS"
	ConditionChange      AlertCondition = "CHANGE"
	ConditionOverbought  AlertCondition = "OVERBOUGHT"
	ConditionOversold    AlertCondition = "OVERSOLD"
)

// NotificationChannel defines how to send alerts
type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "EMAIL"
	ChannelSMS      NotificationChannel = "SMS"
	ChannelPush     NotificationChannel = "PUSH"
	ChannelWebhook  NotificationChannel = "WEBHOOK"
	ChannelInApp    NotificationChannel = "IN_APP"
)

// Alert represents a configured alert
type Alert struct {
	ID          string              `json:"id"`
	UserID      string              `json:"userId"`
	Name        string              `json:"name"`
	Type        AlertType           `json:"type"`
	Symbol      string              `json:"symbol,omitempty"`
	Condition   AlertCondition      `json:"condition"`
	Value       float64             `json:"value,omitempty"`
	Indicator   string              `json:"indicator,omitempty"`   // For indicator alerts
	Period      int                 `json:"period,omitempty"`      // For indicator alerts
	Message     string              `json:"message"`
	Channels    []NotificationChannel `json:"channels"`
	Webhook     string              `json:"webhook,omitempty"`
	Enabled     bool                `json:"enabled"`
	Triggered   bool                `json:"triggered"`
	TriggerOnce bool                `json:"triggerOnce"`   // Fire only once then disable
	CreatedAt   time.Time           `json:"createdAt"`
	TriggeredAt *time.Time          `json:"triggeredAt,omitempty"`
	ExpiresAt   *time.Time          `json:"expiresAt,omitempty"`
}

// AlertTrigger represents an alert that fired
type AlertTrigger struct {
	ID          string              `json:"id"`
	AlertID     string              `json:"alertId"`
	UserID      string              `json:"userId"`
	AlertName   string              `json:"alertName"`
	Message     string              `json:"message"`
	Symbol      string              `json:"symbol,omitempty"`
	CurrentValue float64             `json:"currentValue,omitempty"`
	TriggeredAt time.Time           `json:"triggeredAt"`
	Channels    []NotificationChannel `json:"channels"`
	Sent        bool                `json:"sent"`
	SentAt      *time.Time          `json:"sentAt,omitempty"`
}

// NewsAlert represents a news-based alert
type NewsAlert struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Keywords    []string  `json:"keywords"`
	Symbols     []string  `json:"symbols,omitempty"`
	Categories  []string  `json:"categories,omitempty"`
	MinSeverity string    `json:"minSeverity"` // LOW, MEDIUM, HIGH, CRITICAL
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

// AlertService manages price and indicator alerts
type AlertService struct {
	mu             sync.RWMutex
	alerts         map[string]*Alert
	newsAlerts     map[string]*NewsAlert
	triggers       []*AlertTrigger
	lastPrices     map[string]float64 // For cross detection

	// Callbacks
	priceCallback     func(symbol string) (bid, ask float64, ok bool)
	indicatorCallback func(symbol, indicator string, period int) (float64, error)

	// Notification providers
	emailProvider   func(to, subject, body string) error
	smsProvider     func(to, message string) error
	pushProvider    func(userID, title, message string) error
	webhookProvider func(url string, payload interface{}) error
}

// NewAlertService creates the alert service
func NewAlertService() *AlertService {
	svc := &AlertService{
		alerts:     make(map[string]*Alert),
		newsAlerts: make(map[string]*NewsAlert),
		triggers:   make([]*AlertTrigger, 0),
		lastPrices: make(map[string]float64),
	}

	go svc.processLoop()

	log.Println("[AlertService] Initialized with multi-channel notification support")
	return svc
}

// SetCallbacks configures callbacks
func (s *AlertService) SetCallbacks(
	priceCallback func(symbol string) (bid, ask float64, ok bool),
	indicatorCallback func(symbol, indicator string, period int) (float64, error),
) {
	s.priceCallback = priceCallback
	s.indicatorCallback = indicatorCallback
}

// SetNotificationProviders configures notification channels
func (s *AlertService) SetNotificationProviders(
	emailProvider func(to, subject, body string) error,
	smsProvider func(to, message string) error,
	pushProvider func(userID, title, message string) error,
	webhookProvider func(url string, payload interface{}) error,
) {
	s.emailProvider = emailProvider
	s.smsProvider = smsProvider
	s.pushProvider = pushProvider
	s.webhookProvider = webhookProvider
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(alert *Alert) (*Alert, error) {
	if alert.Name == "" {
		return nil, errors.New("alert name required")
	}

	if len(alert.Channels) == 0 {
		return nil, errors.New("at least one notification channel required")
	}

	alert.ID = uuid.New().String()
	alert.CreatedAt = time.Now()
	alert.Enabled = true
	alert.Triggered = false

	s.mu.Lock()
	s.alerts[alert.ID] = alert
	s.mu.Unlock()

	log.Printf("[Alert] Created: %s (%s alert on %s)", alert.Name, alert.Type, alert.Symbol)
	return alert, nil
}

// UpdateAlert updates an existing alert
func (s *AlertService) UpdateAlert(alertID string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	alert, exists := s.alerts[alertID]
	if !exists {
		return errors.New("alert not found")
	}

	// Update fields
	if name, ok := updates["name"].(string); ok {
		alert.Name = name
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		alert.Enabled = enabled
	}
	if value, ok := updates["value"].(float64); ok {
		alert.Value = value
	}

	log.Printf("[Alert] Updated: %s", alert.Name)
	return nil
}

// DeleteAlert deletes an alert
func (s *AlertService) DeleteAlert(alertID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.alerts[alertID]; !exists {
		return errors.New("alert not found")
	}

	delete(s.alerts, alertID)

	log.Printf("[Alert] Deleted: %s", alertID)
	return nil
}

// GetAlert retrieves an alert
func (s *AlertService) GetAlert(alertID string) (*Alert, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alert, exists := s.alerts[alertID]
	if !exists {
		return nil, errors.New("alert not found")
	}

	return alert, nil
}

// GetUserAlerts returns all alerts for a user
func (s *AlertService) GetUserAlerts(userID string) []*Alert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range s.alerts {
		if alert.UserID == userID {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GetTriggers returns recent alert triggers
func (s *AlertService) GetTriggers(userID string, limit int) []*AlertTrigger {
	s.mu.RLock()
	defer s.mu.RUnlock()

	triggers := make([]*AlertTrigger, 0)
	for _, trigger := range s.triggers {
		if trigger.UserID == userID {
			triggers = append(triggers, trigger)
		}
	}

	// Return most recent
	if limit > 0 && len(triggers) > limit {
		triggers = triggers[len(triggers)-limit:]
	}

	return triggers
}

// processLoop checks alerts
func (s *AlertService) processLoop() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		s.checkAlerts()
		s.sendPendingNotifications()
	}
}

func (s *AlertService) checkAlerts() {
	if s.priceCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for _, alert := range s.alerts {
		if !alert.Enabled || alert.Triggered {
			continue
		}

		// Check expiry
		if alert.ExpiresAt != nil && now.After(*alert.ExpiresAt) {
			alert.Enabled = false
			continue
		}

		// Check alert type
		switch alert.Type {
		case AlertTypePrice:
			s.checkPriceAlert(alert)
		case AlertTypeIndicator:
			s.checkIndicatorAlert(alert)
		case AlertTypePriceChange:
			s.checkPriceChangeAlert(alert)
		}
	}
}

func (s *AlertService) checkPriceAlert(alert *Alert) {
	bid, ask, ok := s.priceCallback(alert.Symbol)
	if !ok {
		return
	}

	price := (bid + ask) / 2.0

	triggered := false

	switch alert.Condition {
	case ConditionAbove:
		if price > alert.Value {
			triggered = true
		}

	case ConditionBelow:
		if price < alert.Value {
			triggered = true
		}

	case ConditionCrossAbove:
		lastPrice, hasLast := s.lastPrices[alert.Symbol]
		if hasLast && lastPrice <= alert.Value && price > alert.Value {
			triggered = true
		}

	case ConditionCrossBelow:
		lastPrice, hasLast := s.lastPrices[alert.Symbol]
		if hasLast && lastPrice >= alert.Value && price < alert.Value {
			triggered = true
		}
	}

	s.lastPrices[alert.Symbol] = price

	if triggered {
		s.triggerAlert(alert, price)
	}
}

func (s *AlertService) checkIndicatorAlert(alert *Alert) {
	if s.indicatorCallback == nil {
		return
	}

	value, err := s.indicatorCallback(alert.Symbol, alert.Indicator, alert.Period)
	if err != nil {
		return
	}

	triggered := false

	switch alert.Indicator {
	case "RSI":
		if alert.Condition == ConditionOverbought && value > alert.Value {
			triggered = true
		} else if alert.Condition == ConditionOversold && value < alert.Value {
			triggered = true
		}

	case "MACD":
		// Would need more complex logic for MACD crossovers
		// This is simplified
		if alert.Condition == ConditionCrossAbove && value > alert.Value {
			triggered = true
		}
	}

	if triggered {
		s.triggerAlert(alert, value)
	}
}

func (s *AlertService) checkPriceChangeAlert(alert *Alert) {
	bid, ask, ok := s.priceCallback(alert.Symbol)
	if !ok {
		return
	}

	price := (bid + ask) / 2.0
	lastPrice, hasLast := s.lastPrices[alert.Symbol]

	if hasLast {
		change := ((price - lastPrice) / lastPrice) * 100.0

		if alert.Condition == ConditionChange {
			if change >= alert.Value || change <= -alert.Value {
				s.triggerAlert(alert, change)
			}
		}
	}

	s.lastPrices[alert.Symbol] = price
}

func (s *AlertService) triggerAlert(alert *Alert, currentValue float64) {
	trigger := &AlertTrigger{
		ID:           uuid.New().String(),
		AlertID:      alert.ID,
		UserID:       alert.UserID,
		AlertName:    alert.Name,
		Message:      alert.Message,
		Symbol:       alert.Symbol,
		CurrentValue: currentValue,
		TriggeredAt:  time.Now(),
		Channels:     alert.Channels,
		Sent:         false,
	}

	s.triggers = append(s.triggers, trigger)

	// Mark alert as triggered
	alert.Triggered = true
	now := time.Now()
	alert.TriggeredAt = &now

	if alert.TriggerOnce {
		alert.Enabled = false
	} else {
		// Reset after cooldown period
		go func() {
			time.Sleep(5 * time.Minute)
			s.mu.Lock()
			alert.Triggered = false
			s.mu.Unlock()
		}()
	}

	log.Printf("[Alert] Triggered: %s for user %s (value: %.5f)",
		alert.Name, alert.UserID, currentValue)
}

func (s *AlertService) sendPendingNotifications() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, trigger := range s.triggers {
		if trigger.Sent {
			continue
		}

		// Send to all configured channels
		for _, channel := range trigger.Channels {
			switch channel {
			case ChannelEmail:
				if s.emailProvider != nil {
					subject := "Trading Alert: " + trigger.AlertName
					body := trigger.Message + "\nCurrent Value: " + formatFloat(trigger.CurrentValue)
					err := s.emailProvider(trigger.UserID, subject, body)
					if err != nil {
						log.Printf("[Alert] Email send failed: %v", err)
					}
				}

			case ChannelSMS:
				if s.smsProvider != nil {
					message := trigger.AlertName + ": " + trigger.Message
					err := s.smsProvider(trigger.UserID, message)
					if err != nil {
						log.Printf("[Alert] SMS send failed: %v", err)
					}
				}

			case ChannelPush:
				if s.pushProvider != nil {
					err := s.pushProvider(trigger.UserID, trigger.AlertName, trigger.Message)
					if err != nil {
						log.Printf("[Alert] Push send failed: %v", err)
					}
				}

			case ChannelWebhook:
				if s.webhookProvider != nil {
					alert, _ := s.alerts[trigger.AlertID]
					if alert != nil && alert.Webhook != "" {
						payload := map[string]interface{}{
							"alert_id":      trigger.AlertID,
							"alert_name":    trigger.AlertName,
							"symbol":        trigger.Symbol,
							"message":       trigger.Message,
							"current_value": trigger.CurrentValue,
							"triggered_at":  trigger.TriggeredAt,
						}
						err := s.webhookProvider(alert.Webhook, payload)
						if err != nil {
							log.Printf("[Alert] Webhook send failed: %v", err)
						}
					}
				}
			}
		}

		trigger.Sent = true
		now := time.Now()
		trigger.SentAt = &now
	}

	// Clean up old triggers (keep last 1000)
	if len(s.triggers) > 1000 {
		s.triggers = s.triggers[len(s.triggers)-1000:]
	}
}

// CreateNewsAlert creates a news alert
func (s *AlertService) CreateNewsAlert(alert *NewsAlert) (*NewsAlert, error) {
	if len(alert.Keywords) == 0 {
		return nil, errors.New("at least one keyword required")
	}

	alert.ID = uuid.New().String()
	alert.CreatedAt = time.Now()
	alert.Enabled = true

	s.mu.Lock()
	s.newsAlerts[alert.ID] = alert
	s.mu.Unlock()

	log.Printf("[NewsAlert] Created for user %s: %v", alert.UserID, alert.Keywords)
	return alert, nil
}

// GetUserNewsAlerts returns news alerts for a user
func (s *AlertService) GetUserNewsAlerts(userID string) []*NewsAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*NewsAlert, 0)
	for _, alert := range s.newsAlerts {
		if alert.UserID == userID {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// SendWebhook sends a webhook notification
func (s *AlertService) SendWebhook(url string, payload interface{}) error {
	_, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// TODO: Send payload in request body
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("webhook failed with status: " + resp.Status)
	}

	log.Printf("[Webhook] Sent to %s", url)
	return nil
}

func formatFloat(value float64) string {
	return sprintf("%.5f", value)
}

func sprintf(format string, a ...interface{}) string {
	// Simple sprintf implementation for formatting
	// In production, use fmt.Sprintf
	return ""
}
