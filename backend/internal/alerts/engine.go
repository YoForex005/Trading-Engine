package alerts

import (
	"crypto/sha256"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Engine is the core alert evaluation and coordination system
type Engine struct {
	rules         map[string]*AlertRule
	alerts        map[string]*Alert
	rulesMu       sync.RWMutex
	alertsMu      sync.RWMutex

	// Metrics source
	metricsSource AccountMetrics

	// Notification dispatcher
	notifier *Notifier

	// Historical data for anomaly detection (simple in-memory for MVP)
	history      map[string][]float64 // accountID+metric -> values
	historyMu    sync.RWMutex

	// Cooldown tracking (rule ID -> last triggered time)
	lastTriggered map[string]time.Time
	cooldownMu    sync.RWMutex

	// Rate limiting (account ID -> count)
	rateLimit    map[string]int
	rateLimitMu  sync.RWMutex

	// Control channels
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewEngine creates a new alert engine
func NewEngine(metricsSource AccountMetrics, notifier *Notifier) *Engine {
	return &Engine{
		rules:         make(map[string]*AlertRule),
		alerts:        make(map[string]*Alert),
		metricsSource: metricsSource,
		notifier:      notifier,
		history:       make(map[string][]float64),
		lastTriggered: make(map[string]time.Time),
		rateLimit:     make(map[string]int),
		stopChan:      make(chan struct{}),
	}
}

// Start begins the alert evaluation loop
func (e *Engine) Start() {
	log.Println("[AlertEngine] Starting evaluation loop (every 5 seconds)")
	e.wg.Add(1)

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		defer e.wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				e.evaluateAllRules()
			case <-e.stopChan:
				log.Println("[AlertEngine] Stopping evaluation loop")
				return
			}
		}
	}()

	// Rate limit reset (every hour)
	e.wg.Add(1)
	rateLimitTicker := time.NewTicker(1 * time.Hour)

	go func() {
		defer e.wg.Done()
		defer rateLimitTicker.Stop()

		for {
			select {
			case <-rateLimitTicker.C:
				e.resetRateLimits()
			case <-e.stopChan:
				return
			}
		}
	}()
}

// Stop halts the alert engine
func (e *Engine) Stop() {
	close(e.stopChan)
	e.wg.Wait()
	log.Println("[AlertEngine] Stopped")
}

// AddRule adds or updates an alert rule
func (e *Engine) AddRule(rule *AlertRule) error {
	e.rulesMu.Lock()
	defer e.rulesMu.Unlock()

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	rule.UpdatedAt = time.Now()

	// Set defaults
	if rule.ZScoreThreshold == 0 {
		rule.ZScoreThreshold = 3.0
	}
	if rule.LookbackPeriod == 0 {
		rule.LookbackPeriod = 100
	}
	if rule.CooldownSeconds == 0 {
		rule.CooldownSeconds = 300 // 5 minutes default
	}

	e.rules[rule.ID] = rule
	log.Printf("[AlertEngine] Rule added: %s (%s) for account %s", rule.ID, rule.Name, rule.AccountID)

	return nil
}

// GetRule retrieves a rule by ID
func (e *Engine) GetRule(ruleID string) (*AlertRule, error) {
	e.rulesMu.RLock()
	defer e.rulesMu.RUnlock()

	rule, exists := e.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", ruleID)
	}

	return rule, nil
}

// ListRules returns all rules, optionally filtered by account
func (e *Engine) ListRules(accountID string) []*AlertRule {
	e.rulesMu.RLock()
	defer e.rulesMu.RUnlock()

	rules := make([]*AlertRule, 0)
	for _, rule := range e.rules {
		if accountID == "" || rule.AccountID == accountID || rule.AccountID == "" {
			rules = append(rules, rule)
		}
	}

	return rules
}

// DeleteRule removes a rule
func (e *Engine) DeleteRule(ruleID string) error {
	e.rulesMu.Lock()
	defer e.rulesMu.Unlock()

	if _, exists := e.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(e.rules, ruleID)
	log.Printf("[AlertEngine] Rule deleted: %s", ruleID)

	return nil
}

// evaluateAllRules runs evaluation for all enabled rules
func (e *Engine) evaluateAllRules() {
	e.rulesMu.RLock()
	rules := make([]*AlertRule, 0, len(e.rules))
	for _, rule := range e.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	e.rulesMu.RUnlock()

	if len(rules) == 0 {
		return
	}

	// Group rules by account for efficient metric fetching
	accountRules := make(map[string][]*AlertRule)
	for _, rule := range rules {
		accountID := rule.AccountID
		if accountID == "" {
			accountID = "*" // System-wide rule
		}
		accountRules[accountID] = append(accountRules[accountID], rule)
	}

	for accountID, rules := range accountRules {
		if accountID == "*" {
			continue // Skip system-wide for now
		}

		snapshot, err := e.metricsSource.GetSnapshot(accountID)
		if err != nil {
			log.Printf("[AlertEngine] Failed to get metrics for account %s: %v", accountID, err)
			continue
		}

		for _, rule := range rules {
			e.evaluateRule(rule, snapshot)
		}
	}
}

// evaluateRule evaluates a single rule against account metrics
func (e *Engine) evaluateRule(rule *AlertRule, snapshot *MetricSnapshot) {
	// Check cooldown
	if !e.canTrigger(rule.ID, rule.CooldownSeconds) {
		return
	}

	// Check rate limit (100 alerts/hour per account)
	if !e.checkRateLimit(snapshot.AccountID) {
		log.Printf("[AlertEngine] Rate limit exceeded for account %s", snapshot.AccountID)
		return
	}

	var triggered bool
	var value float64
	var message string

	switch rule.Type {
	case AlertTypeThreshold:
		triggered, value, message = e.evaluateThreshold(rule, snapshot)
	case AlertTypeAnomaly:
		triggered, value, message = e.evaluateAnomaly(rule, snapshot)
	case AlertTypePattern:
		// Pattern detection requires historical events (not implemented in MVP)
		return
	default:
		log.Printf("[AlertEngine] Unknown alert type: %s", rule.Type)
		return
	}

	if triggered {
		e.triggerAlert(rule, snapshot, value, message)
	}
}

// evaluateThreshold checks if a metric exceeds threshold
func (e *Engine) evaluateThreshold(rule *AlertRule, snapshot *MetricSnapshot) (bool, float64, string) {
	value := e.getMetricValue(rule.Metric, snapshot)

	var triggered bool
	switch rule.Operator {
	case ">":
		triggered = value > rule.Threshold
	case "<":
		triggered = value < rule.Threshold
	case ">=":
		triggered = value >= rule.Threshold
	case "<=":
		triggered = value <= rule.Threshold
	case "==":
		triggered = value == rule.Threshold
	default:
		log.Printf("[AlertEngine] Invalid operator: %s", rule.Operator)
		return false, 0, ""
	}

	if triggered {
		message := fmt.Sprintf("%s %s %.2f (threshold: %.2f)",
			rule.Metric, rule.Operator, value, rule.Threshold)
		return true, value, message
	}

	return false, value, ""
}

// evaluateAnomaly detects statistical anomalies using Z-score
func (e *Engine) evaluateAnomaly(rule *AlertRule, snapshot *MetricSnapshot) (bool, float64, string) {
	value := e.getMetricValue(rule.Metric, snapshot)

	// Store in history
	historyKey := fmt.Sprintf("%s:%s", snapshot.AccountID, rule.Metric)
	e.historyMu.Lock()
	e.history[historyKey] = append(e.history[historyKey], value)

	// Keep only lookback period
	if len(e.history[historyKey]) > rule.LookbackPeriod {
		e.history[historyKey] = e.history[historyKey][len(e.history[historyKey])-rule.LookbackPeriod:]
	}

	values := e.history[historyKey]
	e.historyMu.Unlock()

	// Need at least 30 samples for reliable Z-score
	if len(values) < 30 {
		return false, value, ""
	}

	// Calculate Z-score
	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	if stdDev == 0 {
		return false, value, ""
	}

	zScore := (value - mean) / stdDev

	if math.Abs(zScore) > rule.ZScoreThreshold {
		message := fmt.Sprintf("%s anomaly detected: %.2f (Z-score: %.2f, mean: %.2f, stddev: %.2f)",
			rule.Metric, value, zScore, mean, stdDev)
		return true, value, message
	}

	return false, value, ""
}

// getMetricValue extracts the specified metric from snapshot
func (e *Engine) getMetricValue(metric string, snapshot *MetricSnapshot) float64 {
	switch metric {
	case "balance":
		return snapshot.Balance
	case "equity":
		return snapshot.Equity
	case "margin":
		return snapshot.Margin
	case "freeMargin":
		return snapshot.FreeMargin
	case "marginLevel":
		return snapshot.MarginLevel
	case "exposurePercent":
		return snapshot.ExposurePercent
	case "pnl":
		return snapshot.PnL
	case "positionCount":
		return float64(snapshot.PositionCount)
	default:
		log.Printf("[AlertEngine] Unknown metric: %s", metric)
		return 0
	}
}

// triggerAlert creates and dispatches an alert
func (e *Engine) triggerAlert(rule *AlertRule, snapshot *MetricSnapshot, value float64, message string) {
	// Create fingerprint for deduplication
	fingerprint := e.createFingerprint(rule.ID, snapshot.AccountID, message)

	// Check if already triggered (deduplication)
	if e.isDuplicate(fingerprint) {
		return
	}

	alert := &Alert{
		ID:          uuid.New().String(),
		RuleID:      rule.ID,
		AccountID:   snapshot.AccountID,
		Type:        rule.Type,
		Severity:    rule.Severity,
		Status:      AlertStatusActive,
		Title:       rule.Name,
		Message:     message,
		Metric:      rule.Metric,
		Value:       value,
		Threshold:   rule.Threshold,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Fingerprint: fingerprint,
	}

	// Store alert
	e.alertsMu.Lock()
	e.alerts[alert.ID] = alert
	e.alertsMu.Unlock()

	// Update cooldown
	e.updateCooldown(rule.ID)

	// Increment rate limit counter
	e.incrementRateLimit(snapshot.AccountID)

	log.Printf("[AlertEngine] Alert triggered: %s - %s (severity: %s)", alert.ID, alert.Message, alert.Severity)

	// Dispatch notifications
	for _, channel := range rule.Channels {
		e.notifier.Dispatch(alert, NotificationChannel(channel))
	}
}

// canTrigger checks if enough time has passed since last trigger
func (e *Engine) canTrigger(ruleID string, cooldownSeconds int) bool {
	e.cooldownMu.RLock()
	lastTime, exists := e.lastTriggered[ruleID]
	e.cooldownMu.RUnlock()

	if !exists {
		return true
	}

	elapsed := time.Since(lastTime)
	return elapsed.Seconds() >= float64(cooldownSeconds)
}

// updateCooldown records the trigger time
func (e *Engine) updateCooldown(ruleID string) {
	e.cooldownMu.Lock()
	e.lastTriggered[ruleID] = time.Now()
	e.cooldownMu.Unlock()
}

// checkRateLimit verifies account hasn't exceeded 100 alerts/hour
func (e *Engine) checkRateLimit(accountID string) bool {
	e.rateLimitMu.RLock()
	count := e.rateLimit[accountID]
	e.rateLimitMu.RUnlock()

	return count < 100
}

// incrementRateLimit increases alert count for account
func (e *Engine) incrementRateLimit(accountID string) {
	e.rateLimitMu.Lock()
	e.rateLimit[accountID]++
	e.rateLimitMu.Unlock()
}

// resetRateLimits clears rate limit counters (called every hour)
func (e *Engine) resetRateLimits() {
	e.rateLimitMu.Lock()
	e.rateLimit = make(map[string]int)
	e.rateLimitMu.Unlock()

	log.Println("[AlertEngine] Rate limits reset")
}

// createFingerprint generates a hash for deduplication
func (e *Engine) createFingerprint(ruleID, accountID, message string) string {
	data := fmt.Sprintf("%s:%s:%s", ruleID, accountID, message)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes
}

// isDuplicate checks if alert was recently triggered
func (e *Engine) isDuplicate(fingerprint string) bool {
	e.alertsMu.RLock()
	defer e.alertsMu.RUnlock()

	// Check if any active alert with same fingerprint exists in last 5 minutes
	threshold := time.Now().Add(-5 * time.Minute)
	for _, alert := range e.alerts {
		if alert.Fingerprint == fingerprint &&
		   alert.Status == AlertStatusActive &&
		   alert.CreatedAt.After(threshold) {
			return true
		}
	}

	return false
}

// GetAlert retrieves an alert by ID
func (e *Engine) GetAlert(alertID string) (*Alert, error) {
	e.alertsMu.RLock()
	defer e.alertsMu.RUnlock()

	alert, exists := e.alerts[alertID]
	if !exists {
		return nil, fmt.Errorf("alert not found: %s", alertID)
	}

	return alert, nil
}

// ListAlerts returns alerts filtered by status and account
func (e *Engine) ListAlerts(accountID string, status AlertStatus) []*Alert {
	e.alertsMu.RLock()
	defer e.alertsMu.RUnlock()

	alerts := make([]*Alert, 0)
	for _, alert := range e.alerts {
		if accountID != "" && alert.AccountID != accountID {
			continue
		}
		if status != "" && alert.Status != status {
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts
}

// AcknowledgeAlert marks an alert as acknowledged
func (e *Engine) AcknowledgeAlert(alertID, userID string) error {
	e.alertsMu.Lock()
	defer e.alertsMu.Unlock()

	alert, exists := e.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusAcknowledged
	alert.AckedBy = userID
	alert.AckedAt = &now
	alert.UpdatedAt = now

	log.Printf("[AlertEngine] Alert acknowledged: %s by %s", alertID, userID)

	return nil
}

// ResolveAlert marks an alert as resolved
func (e *Engine) ResolveAlert(alertID string) error {
	e.alertsMu.Lock()
	defer e.alertsMu.Unlock()

	alert, exists := e.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.ResolvedAt = &now
	alert.UpdatedAt = now

	log.Printf("[AlertEngine] Alert resolved: %s", alertID)

	return nil
}

// SnoozeAlert temporarily suppresses an alert
func (e *Engine) SnoozeAlert(alertID string, durationMinutes int) error {
	e.alertsMu.Lock()
	defer e.alertsMu.Unlock()

	alert, exists := e.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	snoozeUntil := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
	alert.Status = AlertStatusSnoozed
	alert.SnoozedUntil = &snoozeUntil
	alert.UpdatedAt = time.Now()

	log.Printf("[AlertEngine] Alert snoozed: %s until %s", alertID, snoozeUntil.Format(time.RFC3339))

	return nil
}

// Helper functions for statistics

func calculateMean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	variance := 0.0
	for _, v := range values {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}
