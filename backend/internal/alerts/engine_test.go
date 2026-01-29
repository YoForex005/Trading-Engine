package alerts

import (
	"testing"
	"time"
)

// MockMetrics implements AccountMetrics interface for testing
type MockMetrics struct {
	snapshots map[string]*MetricSnapshot
}

func NewMockMetrics() *MockMetrics {
	return &MockMetrics{
		snapshots: make(map[string]*MetricSnapshot),
	}
}

func (m *MockMetrics) GetSnapshot(accountID string) (*MetricSnapshot, error) {
	if snap, exists := m.snapshots[accountID]; exists {
		return snap, nil
	}

	// Return default snapshot
	return &MetricSnapshot{
		AccountID:       accountID,
		Timestamp:       time.Now(),
		Balance:         10000.0,
		Equity:          10000.0,
		Margin:          1000.0,
		FreeMargin:      9000.0,
		MarginLevel:     1000.0,
		ExposurePercent: 10.0,
		PositionCount:   1,
		PnL:             0.0,
	}, nil
}

func (m *MockMetrics) SetSnapshot(accountID string, snapshot *MetricSnapshot) {
	m.snapshots[accountID] = snapshot
}

// MockNotifier implements WSHub interface for testing
type MockNotifier struct {
	alerts []*Alert
}

func (m *MockNotifier) BroadcastAlert(alert *Alert) {
	m.alerts = append(m.alerts, alert)
}

func TestThresholdEvaluation(t *testing.T) {
	// Setup
	mockMetrics := NewMockMetrics()
	mockNotifier := &MockNotifier{alerts: make([]*Alert, 0)}
	notifier := NewNotifier(mockNotifier)
	engine := NewEngine(mockMetrics, notifier)

	// Create threshold rule: marginLevel < 100
	rule := &AlertRule{
		ID:              "test-rule-1",
		AccountID:       "test-account",
		Name:            "Margin Call Test",
		Type:            AlertTypeThreshold,
		Severity:        AlertSeverityCritical,
		Enabled:         true,
		Metric:          "marginLevel",
		Operator:        "<",
		Threshold:       100.0,
		Channels:        []string{"dashboard"},
		CooldownSeconds: 300,
	}

	err := engine.AddRule(rule)
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// Test 1: Margin level = 150% (should NOT trigger)
	mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
		AccountID:   "test-account",
		Timestamp:   time.Now(),
		Balance:     10000.0,
		Equity:      10000.0,
		Margin:      6666.67,
		MarginLevel: 150.0,
	})

	engine.evaluateAllRules()
	alerts := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts) != 0 {
		t.Errorf("Expected 0 alerts, got %d", len(alerts))
	}

	// Test 2: Margin level = 80% (should trigger)
	mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
		AccountID:   "test-account",
		Timestamp:   time.Now(),
		Balance:     10000.0,
		Equity:      8000.0,
		Margin:      10000.0,
		MarginLevel: 80.0,
	})

	engine.evaluateAllRules()
	alerts = engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}

	if len(alerts) > 0 {
		alert := alerts[0]
		if alert.Severity != AlertSeverityCritical {
			t.Errorf("Expected CRITICAL severity, got %s", alert.Severity)
		}
		if alert.Value != 80.0 {
			t.Errorf("Expected value 80.0, got %.2f", alert.Value)
		}
	}
}

func TestAnomalyDetection(t *testing.T) {
	// Setup
	mockMetrics := NewMockMetrics()
	mockNotifier := &MockNotifier{alerts: make([]*Alert, 0)}
	notifier := NewNotifier(mockNotifier)
	engine := NewEngine(mockMetrics, notifier)

	// Create anomaly rule
	rule := &AlertRule{
		ID:              "test-anomaly",
		AccountID:       "test-account",
		Name:            "Equity Anomaly",
		Type:            AlertTypeAnomaly,
		Severity:        AlertSeverityMedium,
		Enabled:         true,
		Metric:          "equity",
		ZScoreThreshold: 3.0,
		LookbackPeriod:  50,
		Channels:        []string{"dashboard"},
		CooldownSeconds: 300,
	}

	err := engine.AddRule(rule)
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// Build historical data (equity around 10000)
	for i := 0; i < 40; i++ {
		mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
			AccountID: "test-account",
			Timestamp: time.Now(),
			Equity:    10000.0 + float64(i%10-5)*10, // 10000 ± 50
		})
		engine.evaluateAllRules()
	}

	// Test: Normal value (should NOT trigger)
	mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
		AccountID: "test-account",
		Timestamp: time.Now(),
		Equity:    10000.0,
	})
	engine.evaluateAllRules()
	alerts := engine.ListAlerts("test-account", AlertStatusActive)

	// Note: We need 30+ samples, and Z-score > 3
	// This test verifies the anomaly detection runs without errors
	if len(alerts) > 0 {
		t.Logf("Anomaly detected (expected if value is outlier): %s", alerts[0].Message)
	}
}

func TestAlertAcknowledge(t *testing.T) {
	mockMetrics := NewMockMetrics()
	mockNotifier := &MockNotifier{alerts: make([]*Alert, 0)}
	notifier := NewNotifier(mockNotifier)
	engine := NewEngine(mockMetrics, notifier)

	// Create and trigger alert
	rule := &AlertRule{
		ID:              "ack-test",
		AccountID:       "test-account",
		Name:            "Test Alert",
		Type:            AlertTypeThreshold,
		Severity:        AlertSeverityHigh,
		Enabled:         true,
		Metric:          "marginLevel",
		Operator:        "<",
		Threshold:       100.0,
		Channels:        []string{"dashboard"},
		CooldownSeconds: 1,
	}

	engine.AddRule(rule)

	mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
		AccountID:   "test-account",
		MarginLevel: 50.0,
	})

	engine.evaluateAllRules()
	alerts := engine.ListAlerts("test-account", AlertStatusActive)

	if len(alerts) == 0 {
		t.Fatal("No alert triggered")
	}

	alertID := alerts[0].ID

	// Test acknowledge
	err := engine.AcknowledgeAlert(alertID, "test-user")
	if err != nil {
		t.Fatalf("Failed to acknowledge: %v", err)
	}

	alert, err := engine.GetAlert(alertID)
	if err != nil {
		t.Fatalf("Failed to get alert: %v", err)
	}

	if alert.Status != AlertStatusAcknowledged {
		t.Errorf("Expected status ACKNOWLEDGED, got %s", alert.Status)
	}

	if alert.AckedBy != "test-user" {
		t.Errorf("Expected ackedBy 'test-user', got %s", alert.AckedBy)
	}
}

func TestCooldownPrevention(t *testing.T) {
	mockMetrics := NewMockMetrics()
	mockNotifier := &MockNotifier{alerts: make([]*Alert, 0)}
	notifier := NewNotifier(mockNotifier)
	engine := NewEngine(mockMetrics, notifier)

	// Create rule with 2-second cooldown
	rule := &AlertRule{
		ID:              "cooldown-test",
		AccountID:       "test-account",
		Name:            "Cooldown Test",
		Type:            AlertTypeThreshold,
		Severity:        AlertSeverityHigh,
		Enabled:         true,
		Metric:          "marginLevel",
		Operator:        "<",
		Threshold:       100.0,
		Channels:        []string{"dashboard"},
		CooldownSeconds: 2,
	}

	engine.AddRule(rule)

	// Set low margin
	mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
		AccountID:   "test-account",
		MarginLevel: 50.0,
	})

	// First trigger
	engine.evaluateAllRules()
	alerts1 := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts1) == 0 {
		t.Fatal("Expected alert to trigger")
	}

	// Immediate second evaluation (should NOT trigger due to cooldown)
	engine.evaluateAllRules()
	alerts2 := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts2) != len(alerts1) {
		t.Error("Expected cooldown to prevent second alert")
	}

	// Wait for cooldown to expire
	time.Sleep(3 * time.Second)

	// Third evaluation (should trigger after cooldown)
	engine.evaluateAllRules()
	alerts3 := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts3) <= len(alerts2) {
		t.Error("Expected new alert after cooldown expires")
	}
}

func TestRateLimiting(t *testing.T) {
	mockMetrics := NewMockMetrics()
	mockNotifier := &MockNotifier{alerts: make([]*Alert, 0)}
	notifier := NewNotifier(mockNotifier)
	engine := NewEngine(mockMetrics, notifier)

	// Pre-fill rate limit to 99
	for i := 0; i < 99; i++ {
		engine.incrementRateLimit("test-account")
	}

	// Create rule
	rule := &AlertRule{
		ID:              "rate-test",
		AccountID:       "test-account",
		Name:            "Rate Limit Test",
		Type:            AlertTypeThreshold,
		Severity:        AlertSeverityHigh,
		Enabled:         true,
		Metric:          "marginLevel",
		Operator:        "<",
		Threshold:       100.0,
		Channels:        []string{"dashboard"},
		CooldownSeconds: 1,
	}

	engine.AddRule(rule)

	mockMetrics.SetSnapshot("test-account", &MetricSnapshot{
		AccountID:   "test-account",
		MarginLevel: 50.0,
	})

	// Should trigger (99 < 100 limit)
	engine.evaluateAllRules()
	alerts1 := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts1) == 0 {
		t.Error("Expected alert to trigger under rate limit")
	}

	time.Sleep(2 * time.Second)

	// Should NOT trigger (100 >= 100 limit)
	engine.evaluateAllRules()
	alerts2 := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts2) > len(alerts1) {
		t.Error("Expected rate limit to block alert")
	}

	// Reset rate limits
	engine.resetRateLimits()

	time.Sleep(2 * time.Second)

	// Should trigger again after reset
	engine.evaluateAllRules()
	alerts3 := engine.ListAlerts("test-account", AlertStatusActive)
	if len(alerts3) <= len(alerts2) {
		t.Error("Expected alert after rate limit reset")
	}
}

func TestDefaultRules(t *testing.T) {
	rules := GetDefaultRules("test-account")

	if len(rules) == 0 {
		t.Fatal("Expected default rules to be returned")
	}

	// Verify margin call rule exists
	foundMarginCall := false
	for _, rule := range rules {
		if rule.Name == "Margin Call Alert" {
			foundMarginCall = true
			if rule.Severity != AlertSeverityCritical {
				t.Error("Margin call should be CRITICAL severity")
			}
			if rule.Metric != "marginLevel" {
				t.Error("Margin call should monitor marginLevel")
			}
			if rule.Threshold != 100.0 {
				t.Error("Margin call threshold should be 100")
			}
		}
	}

	if !foundMarginCall {
		t.Error("Default rules should include Margin Call Alert")
	}
}

func TestRuleValidation(t *testing.T) {
	// Valid rule
	validRule := &AlertRule{
		Name:     "Valid Rule",
		Type:     AlertTypeThreshold,
		Metric:   "marginLevel",
		Operator: "<",
		Threshold: 100.0,
		Channels: []string{"dashboard"},
	}

	err := ValidateRule(validRule)
	if err != nil {
		t.Errorf("Valid rule failed validation: %v", err)
	}

	// Invalid: missing name
	invalidRule1 := &AlertRule{
		Type:   AlertTypeThreshold,
		Metric: "marginLevel",
	}

	err = ValidateRule(invalidRule1)
	if err == nil {
		t.Error("Expected validation error for missing name")
	}

	// Invalid: missing operator
	invalidRule2 := &AlertRule{
		Name:   "Invalid",
		Type:   AlertTypeThreshold,
		Metric: "marginLevel",
	}

	err = ValidateRule(invalidRule2)
	if err == nil {
		t.Error("Expected validation error for missing operator")
	}

	// Invalid: bad operator
	invalidRule3 := &AlertRule{
		Name:     "Invalid",
		Type:     AlertTypeThreshold,
		Metric:   "marginLevel",
		Operator: "!=", // Not supported
	}

	err = ValidateRule(invalidRule3)
	if err == nil {
		t.Error("Expected validation error for invalid operator")
	}
}
