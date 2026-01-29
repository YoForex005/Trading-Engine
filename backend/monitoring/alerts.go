package monitoring

import (
	"fmt"
	"sync"
	"time"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	SeverityInfo     AlertSeverity = "info"
	SeverityWarning  AlertSeverity = "warning"
	SeverityCritical AlertSeverity = "critical"
)

// Alert represents a monitoring alert
type Alert struct {
	Name        string
	Severity    AlertSeverity
	Message     string
	Timestamp   time.Time
	Labels      map[string]string
	Annotations map[string]string
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	Name        string
	Description string
	Query       string
	Threshold   float64
	Duration    time.Duration
	Severity    AlertSeverity
	Enabled     bool
}

// AlertManager manages alerting rules and notifications
type AlertManager struct {
	rules         map[string]*AlertRule
	activeAlerts  map[string]*Alert
	alertHistory  []*Alert
	mu            sync.RWMutex
	logger        *Logger
	maxHistory    int
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		rules:        make(map[string]*AlertRule),
		activeAlerts: make(map[string]*Alert),
		alertHistory: make([]*Alert, 0),
		logger:       GetLogger(),
		maxHistory:   1000,
	}
}

// RegisterRule registers an alert rule
func (am *AlertManager) RegisterRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.rules[rule.Name] = rule
}

// FireAlert fires an alert
func (am *AlertManager) FireAlert(alert *Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Add to active alerts
	am.activeAlerts[alert.Name] = alert

	// Add to history
	am.alertHistory = append(am.alertHistory, alert)
	if len(am.alertHistory) > am.maxHistory {
		am.alertHistory = am.alertHistory[1:]
	}

	// Log alert
	fields := map[string]interface{}{
		"alert_name":     alert.Name,
		"severity":       alert.Severity,
		"labels":         alert.Labels,
		"annotations":    alert.Annotations,
		"event_type":     "alert",
	}

	logLevel := INFO
	switch alert.Severity {
	case SeverityWarning:
		logLevel = WARN
	case SeverityCritical:
		logLevel = ERROR
	}

	am.logger.log(logLevel, fmt.Sprintf("ALERT: %s - %s", alert.Name, alert.Message), fields, nil)
}

// ResolveAlert resolves an active alert
func (am *AlertManager) ResolveAlert(name string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert, exists := am.activeAlerts[name]; exists {
		delete(am.activeAlerts, name)

		am.logger.Info(fmt.Sprintf("Alert resolved: %s", name), map[string]interface{}{
			"alert_name":  name,
			"severity":    alert.Severity,
			"duration":    time.Since(alert.Timestamp).Seconds(),
			"event_type":  "alert_resolved",
		})
	}
}

// GetActiveAlerts returns all active alerts
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// Predefined Alert Rules for Trading Engine

// GetDefaultAlertRules returns default alert rules for trading engine
func GetDefaultAlertRules() []*AlertRule {
	return []*AlertRule{
		{
			Name:        "HighOrderLatency",
			Description: "Order execution latency exceeds 500ms",
			Query:       "trading_order_execution_latency_milliseconds{quantile=\"0.95\"} > 500",
			Threshold:   500,
			Duration:    2 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "CriticalOrderLatency",
			Description: "Order execution latency exceeds 2000ms",
			Query:       "trading_order_execution_latency_milliseconds{quantile=\"0.95\"} > 2000",
			Threshold:   2000,
			Duration:    1 * time.Minute,
			Severity:    SeverityCritical,
			Enabled:     true,
		},
		{
			Name:        "HighOrderErrorRate",
			Description: "Order error rate exceeds 5%",
			Query:       "rate(trading_order_errors_total[5m]) > 0.05",
			Threshold:   0.05,
			Duration:    5 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "LPDisconnected",
			Description: "LP connection lost",
			Query:       "trading_lp_connected == 0",
			Threshold:   0,
			Duration:    30 * time.Second,
			Severity:    SeverityCritical,
			Enabled:     true,
		},
		{
			Name:        "HighLPLatency",
			Description: "LP latency exceeds 1000ms",
			Query:       "trading_lp_latency_milliseconds{quantile=\"0.95\"} > 1000",
			Threshold:   1000,
			Duration:    2 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "HighMemoryUsage",
			Description: "Memory usage exceeds 80%",
			Query:       "trading_memory_usage_bytes / trading_memory_total_bytes > 0.8",
			Threshold:   0.8,
			Duration:    5 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "HighGoroutineCount",
			Description: "Goroutine count exceeds 10000",
			Query:       "trading_goroutines_count > 10000",
			Threshold:   10000,
			Duration:    5 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "HighAPIErrorRate",
			Description: "API error rate exceeds 5%",
			Query:       "rate(trading_api_requests_total{status=~\"5..\"}[5m]) > 0.05",
			Threshold:   0.05,
			Duration:    5 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "SlowDatabaseQueries",
			Description: "Database query latency exceeds 100ms",
			Query:       "trading_db_query_duration_milliseconds{quantile=\"0.95\"} > 100",
			Threshold:   100,
			Duration:    5 * time.Minute,
			Severity:    SeverityWarning,
			Enabled:     true,
		},
		{
			Name:        "LowWebSocketConnections",
			Description: "No active WebSocket connections",
			Query:       "trading_websocket_connections == 0",
			Threshold:   0,
			Duration:    2 * time.Minute,
			Severity:    SeverityInfo,
			Enabled:     false, // Optional monitoring
		},
		{
			Name:        "SLOViolation",
			Description: "Order execution SLO violated (< 95% within 100ms)",
			Query:       "trading_slo_order_execution_within_target_total / trading_slo_order_execution_success_total < 0.95",
			Threshold:   0.95,
			Duration:    10 * time.Minute,
			Severity:    SeverityCritical,
			Enabled:     true,
		},
	}
}

// Global alert manager
var globalAlertManager = NewAlertManager()

// GetAlertManager returns the global alert manager
func GetAlertManager() *AlertManager {
	return globalAlertManager
}

// SetGlobalAlertManager sets the global alert manager
func SetGlobalAlertManager(am *AlertManager) {
	globalAlertManager = am
}
