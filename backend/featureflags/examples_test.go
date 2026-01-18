package featureflags

import (
	"fmt"
	"testing"
	"time"
)

// Example: Basic boolean flag
func ExampleSDK_IsEnabled() {
	sdk := NewSDK()

	// Create a boolean flag
	flag := NewFlag("new-dashboard", "New Dashboard").
		Boolean(true).
		Description("Enable new dashboard UI").
		Environment("production").
		Build()

	sdk.CreateFlag(flag, "admin@example.com")

	// Check if enabled for a user
	ctx := &EvaluationContext{
		UserID:  "user123",
		Country: "US",
		Tier:    "premium",
	}

	enabled := sdk.IsEnabled("new-dashboard", "user123", ctx)
	fmt.Println("Dashboard enabled:", enabled)
	// Output: Dashboard enabled: true
}

// Example: Percentage rollout
func ExampleSDK_percentageRollout() {
	sdk := NewSDK()

	// Create a flag with 25% rollout
	flag := NewFlag("beta-features", "Beta Features").
		Percentage(25).
		Description("Release beta features to 25% of users").
		Build()

	sdk.CreateFlag(flag, "admin@example.com")

	// Check for multiple users
	enabled := 0
	for i := 0; i < 100; i++ {
		userID := fmt.Sprintf("user%d", i)
		if sdk.IsEnabled("beta-features", userID, nil) {
			enabled++
		}
	}

	fmt.Printf("Approximately %d%% of users enabled\n", enabled)
}

// Example: Targeted rollout
func ExampleSDK_targetedRollout() {
	sdk := NewSDK()

	// Create targeting rules
	segment := NewSegment().
		Country("US", "CA", "UK").
		Tier("premium", "vip").
		DeviceType("mobile").
		Build()

	flag := NewFlag("mobile-checkout", "Mobile Checkout").
		Boolean(true).
		Rules(segment...).
		Description("New mobile checkout for premium users in select countries").
		Build()

	sdk.CreateFlag(flag, "admin@example.com")

	// User in target segment
	ctx1 := &EvaluationContext{
		UserID:     "user1",
		Country:    "US",
		Tier:       "premium",
		DeviceType: "mobile",
	}
	fmt.Println("Premium mobile US user:", sdk.IsEnabled("mobile-checkout", "user1", ctx1))

	// User not in segment
	ctx2 := &EvaluationContext{
		UserID:     "user2",
		Country:    "FR",
		Tier:       "free",
		DeviceType: "desktop",
	}
	fmt.Println("Free desktop FR user:", sdk.IsEnabled("mobile-checkout", "user2", ctx2))
}

// Example: A/B test experiment
func ExampleSDK_abTest() {
	sdk := NewSDK()

	// Create an A/B test experiment
	experiment := NewExperiment("checkout-flow-test", "Checkout Flow A/B Test").
		Description("Test new vs old checkout flow").
		Control("control", "Old Checkout", 50).
		Variant("variant-a", "New Checkout", 50, map[string]interface{}{
			"flow":  "single-page",
			"color": "blue",
		}).
		Traffic(100).
		SampleSize(10000).
		Duration(7*24*time.Hour, 30*24*time.Hour).
		Confidence(0.95).
		Build()

	sdk.CreateExperiment(experiment, "product@example.com")
	sdk.StartExperiment("checkout-flow-test")

	// Get variant for a user
	variant := sdk.GetExperimentVariant("checkout-flow-test", "user123", nil)
	if variant != nil {
		fmt.Println("Assigned variant:", variant.Name)
		fmt.Println("Config:", variant.Config)

		// Track conversion
		sdk.TrackConversion("checkout-flow-test", "user123", "purchase", 99.99)
	}
}

// Example: Multivariate test
func ExampleSDK_multivariateTest() {
	sdk := NewSDK()

	// Create multivariate flag
	variants := []Variant{
		{ID: "red", Name: "Red Button", Weight: 33, Payload: map[string]interface{}{"color": "red"}},
		{ID: "green", Name: "Green Button", Weight: 33, Payload: map[string]interface{}{"color": "green"}},
		{ID: "blue", Name: "Blue Button", Weight: 34, Payload: map[string]interface{}{"color": "blue"}},
	}

	flag := NewFlag("button-color", "Button Color Test").
		Multivariate(variants...).
		Build()

	sdk.CreateFlag(flag, "designer@example.com")

	// Get variant for user
	variant := sdk.GetVariant("button-color", "user123", nil)
	if variant != nil {
		fmt.Println("Button color:", variant.Payload["color"])
	}
}

// Example: Gradual rollout
func ExampleSDK_gradualRollout() {
	sdk := NewSDK()

	// Create a gradual rollout
	rollout := &Rollout{
		ID:          "api-v2-rollout",
		Name:        "API v2 Rollout",
		Description: "Gradual rollout of API v2",
		FlagID:      "api-v2",
		Strategy:    StrategyExponential,
		AutoProgress: true,
		MinStageDuration: 2 * time.Hour,
		RollbackOnError: true,
		RollbackThresholds: []RollbackThreshold{
			{Metric: "error_rate", Threshold: 0.05, Duration: 5 * time.Minute},
			{Metric: "latency", Threshold: 1000, Duration: 5 * time.Minute},
		},
	}

	sdk.CreateRollout(rollout)
	sdk.StartRollout("api-v2-rollout")

	// Check if user is in rollout
	inRollout := sdk.IsInRollout("api-v2-rollout", "user123")
	fmt.Println("User in rollout:", inRollout)

	// Get current rollout percentage
	percentage := sdk.GetRolloutPercentage("api-v2-rollout")
	fmt.Println("Current rollout:", percentage, "%")
}

// Example: Kill switch
func ExampleSDK_killSwitch() {
	sdk := NewSDK()

	// Create a kill switch (instantly disable feature)
	flag := &Flag{
		ID:          "payment-processing",
		Name:        "Payment Processing",
		Type:        FlagTypeKillSwitch,
		Enabled:     false, // Disabled by default
		Description: "Emergency kill switch for payment processing",
	}

	sdk.CreateFlag(flag, "admin@example.com")

	// In case of emergency, toggle to disable
	sdk.ToggleFlag("payment-processing", false, "ops@example.com")
}

// Example: Feature dependencies
func ExampleSDK_dependencies() {
	sdk := NewSDK()

	// Create parent feature
	parentFlag := NewFlag("new-api", "New API").
		Boolean(true).
		Build()
	sdk.CreateFlag(parentFlag, "admin@example.com")

	// Create dependent feature
	dependentFlag := &Flag{
		ID:          "new-api-analytics",
		Name:        "New API Analytics",
		Type:        FlagTypeBoolean,
		Enabled:     true,
		DefaultValue: true,
		Dependencies: []FlagDependency{
			{FlagID: "new-api", Operator: "enabled"},
		},
	}
	sdk.CreateFlag(dependentFlag, "admin@example.com")

	// Analytics only enabled if new-api is enabled
	enabled := sdk.IsEnabled("new-api-analytics", "user123", nil)
	fmt.Println("Analytics enabled:", enabled)
}

// Example: Scheduled flag
func ExampleSDK_scheduledFlag() {
	sdk := NewSDK()

	// Schedule flag to enable at specific time
	enableTime := time.Now().Add(24 * time.Hour)
	disableTime := enableTime.Add(7 * 24 * time.Hour)

	flag := NewFlag("black-friday-sale", "Black Friday Sale").
		Boolean(true).
		EnableAt(enableTime).
		DisableAt(disableTime).
		Description("Automatically enable during Black Friday week").
		Build()

	sdk.CreateFlag(flag, "marketing@example.com")
}

// Example: Funnel analysis
func ExampleSDK_funnelAnalysis() {
	sdk := NewSDK()

	// Create funnel
	funnel := &Funnel{
		ID:   "checkout-funnel",
		Name: "Checkout Funnel",
		Steps: []FunnelStep{
			{ID: "view-cart", Name: "View Cart", EventType: "view", Order: 1},
			{ID: "add-shipping", Name: "Add Shipping", EventType: "shipping", Order: 2},
			{ID: "add-payment", Name: "Add Payment", EventType: "payment", Order: 3},
			{ID: "complete", Name: "Complete Purchase", EventType: "conversion", Order: 4},
		},
	}

	sdk.Analytics().CreateFunnel(funnel)

	// Analyze funnel for experiment
	// events := []ExperimentEvent{...}
	// analysis := sdk.Analytics().AnalyzeFunnel("checkout-funnel", "checkout-test", events)
}

// Example: Sample size calculation
func ExampleSampleSizeCalculator() {
	calc := &SampleSizeCalculator{
		BaselineRate:        0.10, // 10% conversion rate
		MinDetectableEffect: 0.10, // Want to detect 10% relative change (1% absolute)
		Confidence:          0.95, // 95% confidence
		Power:               0.80, // 80% power
		NumVariants:         2,    // A/B test
	}

	sampleSize := calc.Calculate()
	fmt.Printf("Required sample size: %d users per variant\n", sampleSize/2)
}

// Example: Callbacks for analytics
func ExampleSDK_callbacks() {
	sdk := NewSDK()

	// Register flag evaluation callback
	sdk.OnFlagEvaluation(func(eval *FlagEvaluation) {
		fmt.Printf("Flag %s evaluated for user %s: %v\n",
			eval.FlagID, eval.UserID, eval.Enabled)
		// Send to analytics platform
	})

	// Register experiment assignment callback
	sdk.OnExperimentAssignment(func(experimentID, variantID, userID string) {
		fmt.Printf("User %s assigned to variant %s in experiment %s\n",
			userID, variantID, experimentID)
		// Send to analytics platform
	})
}

// Benchmark flag evaluation performance
func BenchmarkFlagEvaluation(b *testing.B) {
	sdk := NewSDK()

	flag := NewFlag("test-flag", "Test Flag").
		Boolean(true).
		Build()

	sdk.CreateFlag(flag, "test")

	ctx := &EvaluationContext{
		UserID:  "user123",
		Country: "US",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sdk.IsEnabled("test-flag", "user123", ctx)
	}
}

// Benchmark experiment assignment
func BenchmarkExperimentAssignment(b *testing.B) {
	sdk := NewSDK()

	experiment := NewExperiment("test-exp", "Test Experiment").
		Control("control", "Control", 50).
		Variant("variant", "Variant", 50, nil).
		Build()

	sdk.CreateExperiment(experiment, "test")
	sdk.StartExperiment("test-exp")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sdk.GetExperimentVariant("test-exp", fmt.Sprintf("user%d", i), nil)
	}
}

// Test flag evaluation performance target (<1ms)
func TestFlagEvaluationPerformance(t *testing.T) {
	sdk := NewSDK()

	flag := NewFlag("perf-test", "Performance Test").
		Boolean(true).
		Build()

	sdk.CreateFlag(flag, "test")

	ctx := &EvaluationContext{
		UserID:  "user123",
		Country: "US",
	}

	// Warm up cache
	for i := 0; i < 100; i++ {
		sdk.IsEnabled("perf-test", "user123", ctx)
	}

	// Measure
	start := time.Now()
	iterations := 10000
	for i := 0; i < iterations; i++ {
		sdk.IsEnabled("perf-test", "user123", ctx)
	}
	elapsed := time.Since(start)

	avgLatency := elapsed / time.Duration(iterations)
	if avgLatency > time.Millisecond {
		t.Errorf("Average latency %v exceeds 1ms target", avgLatency)
	}

	t.Logf("Average latency: %v (target: <1ms)", avgLatency)
}
