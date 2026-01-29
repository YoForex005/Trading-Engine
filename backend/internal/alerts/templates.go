package alerts

// GetDefaultRules returns pre-configured alert rule templates
func GetDefaultRules(accountID string) []*AlertRule {
	return []*AlertRule{
		// Critical: Margin call (margin level < 100%)
		{
			Name:        "Margin Call Alert",
			Description: "Triggers when margin level drops below 100%",
			Type:        AlertTypeThreshold,
			Severity:    AlertSeverityCritical,
			Enabled:     true,
			Metric:      "marginLevel",
			Operator:    "<",
			Threshold:   100.0,
			Channels:    []string{"dashboard", "email"},
			CooldownSeconds: 300, // 5 minutes
			AccountID:   accountID,
		},

		// High: Low margin warning (margin level < 150%)
		{
			Name:        "Low Margin Warning",
			Description: "Warning when margin level drops below 150%",
			Type:        AlertTypeThreshold,
			Severity:    AlertSeverityHigh,
			Enabled:     true,
			Metric:      "marginLevel",
			Operator:    "<",
			Threshold:   150.0,
			Channels:    []string{"dashboard"},
			CooldownSeconds: 600, // 10 minutes
			AccountID:   accountID,
		},

		// High: High exposure (> 80% of equity)
		{
			Name:        "High Exposure Alert",
			Description: "Triggers when margin usage exceeds 80% of equity",
			Type:        AlertTypeThreshold,
			Severity:    AlertSeverityHigh,
			Enabled:     true,
			Metric:      "exposurePercent",
			Operator:    ">",
			Threshold:   80.0,
			Channels:    []string{"dashboard"},
			CooldownSeconds: 600,
			AccountID:   accountID,
		},

		// Medium: Large unrealized loss (> 20% of balance)
		{
			Name:        "Large Unrealized Loss",
			Description: "Alert when unrealized loss exceeds 20% of account balance",
			Type:        AlertTypeThreshold,
			Severity:    AlertSeverityMedium,
			Enabled:     false, // Disabled by default, user can enable
			Metric:      "pnl",
			Operator:    "<",
			Threshold:   -1000.0, // Should be calculated as % of balance
			Channels:    []string{"dashboard"},
			CooldownSeconds: 900, // 15 minutes
			AccountID:   accountID,
		},

		// Low: Equity anomaly detection
		{
			Name:        "Equity Anomaly Detection",
			Description: "Detects unusual equity changes using Z-score analysis",
			Type:        AlertTypeAnomaly,
			Severity:    AlertSeverityMedium,
			Enabled:     false, // Disabled by default, needs historical data
			Metric:      "equity",
			ZScoreThreshold: 3.0,
			LookbackPeriod:  100,
			Channels:    []string{"dashboard"},
			CooldownSeconds: 1800, // 30 minutes
			AccountID:   accountID,
		},

		// Medium: Free margin running low
		{
			Name:        "Low Free Margin",
			Description: "Alert when free margin drops below $500",
			Type:        AlertTypeThreshold,
			Severity:    AlertSeverityMedium,
			Enabled:     true,
			Metric:      "freeMargin",
			Operator:    "<",
			Threshold:   500.0,
			Channels:    []string{"dashboard"},
			CooldownSeconds: 900,
			AccountID:   accountID,
		},

		// Low: Too many open positions
		{
			Name:        "Position Count Warning",
			Description: "Alert when number of open positions exceeds 10",
			Type:        AlertTypeThreshold,
			Severity:    AlertSeverityLow,
			Enabled:     false,
			Metric:      "positionCount",
			Operator:    ">",
			Threshold:   10.0,
			Channels:    []string{"dashboard"},
			CooldownSeconds: 1800,
			AccountID:   accountID,
		},
	}
}

// GetMetricDescription returns human-readable descriptions for metrics
func GetMetricDescription(metric string) string {
	descriptions := map[string]string{
		"balance":         "Account cash balance",
		"equity":          "Balance + unrealized P/L",
		"margin":          "Total margin used by open positions",
		"freeMargin":      "Available margin for new positions",
		"marginLevel":     "Equity / Margin × 100 (%)",
		"exposurePercent": "Margin used / Equity × 100 (%)",
		"pnl":             "Total unrealized profit/loss",
		"positionCount":   "Number of open positions",
	}

	if desc, exists := descriptions[metric]; exists {
		return desc
	}

	return "Unknown metric"
}

// GetSeverityColor returns UI color code for severity level
func GetSeverityColor(severity AlertSeverity) string {
	colors := map[AlertSeverity]string{
		AlertSeverityLow:      "#4CAF50", // Green
		AlertSeverityMedium:   "#FF9800", // Orange
		AlertSeverityHigh:     "#F44336", // Red
		AlertSeverityCritical: "#9C27B0", // Purple
	}

	if color, exists := colors[severity]; exists {
		return color
	}

	return "#757575" // Gray default
}

// ValidateRule checks if an alert rule is properly configured
func ValidateRule(rule *AlertRule) error {
	if rule.Name == "" {
		return ErrInvalidRule("name is required")
	}

	if rule.Type == "" {
		return ErrInvalidRule("type is required")
	}

	if rule.Metric == "" {
		return ErrInvalidRule("metric is required")
	}

	// Type-specific validation
	switch rule.Type {
	case AlertTypeThreshold:
		if rule.Operator == "" {
			return ErrInvalidRule("operator is required for threshold alerts")
		}
		validOperators := map[string]bool{
			">": true, "<": true, ">=": true, "<=": true, "==": true,
		}
		if !validOperators[rule.Operator] {
			return ErrInvalidRule("invalid operator: must be >, <, >=, <=, or ==")
		}

	case AlertTypeAnomaly:
		if rule.ZScoreThreshold <= 0 {
			rule.ZScoreThreshold = 3.0 // Default
		}
		if rule.LookbackPeriod <= 0 {
			rule.LookbackPeriod = 100 // Default
		}

	case AlertTypePattern:
		if rule.Pattern == "" {
			return ErrInvalidRule("pattern is required for pattern alerts")
		}
		if rule.PatternCount <= 0 {
			return ErrInvalidRule("patternCount must be > 0")
		}
	}

	// Validate channels
	if len(rule.Channels) == 0 {
		rule.Channels = []string{"dashboard"} // Default
	}

	validChannels := map[string]bool{
		"dashboard": true,
		"email":     true,
		"sms":       true,
		"webhook":   true,
	}

	for _, channel := range rule.Channels {
		if !validChannels[channel] {
			return ErrInvalidRule("invalid channel: " + channel)
		}
	}

	return nil
}

// ErrInvalidRule creates a validation error
func ErrInvalidRule(message string) error {
	return &ValidationError{Message: message}
}

// ValidationError represents a rule validation error
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return "invalid alert rule: " + e.Message
}
