package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/epic1st/rtx/backend/logging"
)

func main() {
	// Example 1: Basic logging
	basicLogging()

	// Example 2: Logging with context
	contextLogging()

	// Example 3: HTTP middleware
	httpMiddleware()

	// Example 4: Audit logging
	auditLogging()

	// Example 5: Performance monitoring
	performanceMonitoring()

	// Example 6: Error tracking
	errorTracking()

	// Example 7: Production setup
	productionSetup()
}

// Example 1: Basic Logging
func basicLogging() {
	// Simple info log
	logging.Info("Application started", logging.String("version", "1.0.0"))

	// Log with multiple fields
	logging.Info("Order received",
		logging.OrderID("ORD-123"),
		logging.Symbol("EURUSD"),
		logging.Float64("volume", 1.5),
		logging.String("side", "BUY"),
	)

	// Warning log
	logging.Warn("High volume detected",
		logging.Symbol("EURUSD"),
		logging.Float64("volume", 100.0),
	)

	// Error log
	err := errors.New("connection timeout")
	logging.Error("Failed to connect to LP", err,
		logging.Component("lp-manager"),
		logging.String("lp", "OANDA"),
	)

	// Debug log (only shows if level is DEBUG)
	logging.Debug("Processing tick",
		logging.Symbol("EURUSD"),
		logging.Float64("bid", 1.1234),
		logging.Float64("ask", 1.1236),
	)
}

// Example 2: Logging with Context
func contextLogging() {
	// Create context with request ID
	ctx := context.Background()
	ctx = logging.ContextWithRequestID(ctx, "req-123-456")
	ctx = logging.ContextWithUserID(ctx, "user-789")
	ctx = logging.ContextWithAccountID(ctx, "acc-001")

	// All logs will include request ID, user ID, and account ID
	logging.WithContext(ctx).Info("Processing order",
		logging.OrderID("ORD-456"),
		logging.Symbol("GBPUSD"),
	)

	logging.WithContext(ctx).Error("Insufficient margin", errors.New("margin too low"),
		logging.Float64("required_margin", 1000.0),
		logging.Float64("available_margin", 500.0),
	)
}

// Example 3: HTTP Middleware
func httpMiddleware() {
	// Create logger
	logger := logging.NewLogger(logging.INFO)

	// Add Sentry hook if DSN is available
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		sentryHook, err := logging.NewSentryHook(dsn, "production")
		if err == nil {
			logger.AddHook(sentryHook)
		}
	}

	// Note: This would be used in actual HTTP server setup
	// loggingMiddleware := logging.HTTPLoggingMiddleware(logger)
	// panicMiddleware := logging.PanicRecoveryMiddleware(logger)
	//
	// http.Handle("/api/orders",
	//     panicMiddleware(loggingMiddleware(http.HandlerFunc(handleOrders))),
	// )
}

// Example 4: Audit Logging
func auditLogging() {
	// Initialize audit logger
	auditLogger, err := logging.NewAuditLogger("./logs/audit")
	if err != nil {
		logging.Error("Failed to initialize audit logger", err)
		return
	}
	defer auditLogger.Close()

	ctx := context.Background()
	ctx = logging.ContextWithUserID(ctx, "user-123")
	ctx = logging.ContextWithAccountID(ctx, "acc-456")

	// Log order placement
	auditLogger.LogOrderPlacement(
		ctx,
		"ORD-789",
		"EURUSD",
		"BUY",
		1.5,
		1.1234,
		"MARKET",
		"acc-456",
	)

	// Log authentication
	auditLogger.LogAuthentication(ctx, "user-123", "192.168.1.100", "password")

	// Log position close
	auditLogger.LogPositionClose(ctx, "POS-123", "acc-456", 150.50, 1.1250)

	// Log admin action
	auditLogger.LogAdminAction(
		ctx,
		"admin-001",
		"update_leverage",
		"account",
		"acc-456",
		map[string]interface{}{"leverage": 50},
		map[string]interface{}{"leverage": 100},
	)

	// Log LP routing decision
	auditLogger.LogLPRouting(
		ctx,
		"ORD-789",
		"OANDA",
		"best_execution",
		"acc-456",
		map[string]interface{}{
			"spread":   0.00002,
			"latency":  50,
			"priority": "high",
		},
	)
}

// Example 5: Performance Monitoring
func performanceMonitoring() {
	ctx := context.Background()

	// Monitor slow database query
	start := time.Now()
	// Simulate query execution
	time.Sleep(150 * time.Millisecond)
	duration := time.Since(start)

	// This will log if duration > 100ms
	logging.LogSlowQuery(ctx, "SELECT * FROM positions WHERE account_id = ?", duration)

	// Monitor slow endpoint
	logging.LogSlowEndpoint(
		"POST",
		"/api/orders/market",
		1200*time.Millisecond, // 1.2 seconds
		200,
		"req-123",
	)

	// Get performance stats
	slowQueries := logging.GetSlowQueries()
	logging.Info("Slow query count", logging.Int("count", len(slowQueries)))

	slowEndpoints := logging.GetSlowEndpoints()
	logging.Info("Slow endpoint count", logging.Int("count", len(slowEndpoints)))
}

// Example 6: Error Tracking
func errorTracking() {
	ctx := logging.ContextWithUserID(context.Background(), "user-123")

	// Track an error
	err := errors.New("database connection failed")
	logging.TrackError(ctx, err, "critical", map[string]interface{}{
		"host":     "localhost",
		"port":     5432,
		"database": "trading_engine",
	})

	// Register alert callback
	logging.RegisterErrorAlert(func(stats *logging.ErrorStats) {
		logging.Warn("Error threshold exceeded",
			logging.String("error", stats.Message),
			logging.Int64("count", stats.Count),
			logging.String("severity", stats.Severity),
		)
		// Send notification (email, Slack, etc.)
	})

	// Get error statistics
	stats := logging.GetErrorStats()
	for key, errorStat := range stats {
		logging.Info("Error statistics",
			logging.String("key", key),
			logging.Int64("count", errorStat.Count),
			logging.String("severity", errorStat.Severity),
		)
	}

	// Get top errors
	topErrors := logging.GetTopErrors(5)
	for i, errorStat := range topErrors {
		logging.Info("Top error",
			logging.Int("rank", i+1),
			logging.String("error", errorStat.Message),
			logging.Int64("count", errorStat.Count),
		)
	}
}

// Example 7: Production Setup
func productionSetup() {
	// Create rotating file writer
	rotatingWriter, err := logging.NewRotatingFileWriter(logging.RotationConfig{
		Filename:           "./logs/production.log",
		MaxSizeMB:          100,
		MaxAge:             7 * 24 * time.Hour,
		MaxBackups:         30,
		CompressionEnabled: true,
	})
	if err != nil {
		panic(err)
	}
	defer rotatingWriter.Close()

	// Create multi-writer (file + stdout)
	multiWriter := logging.NewMultiWriter(rotatingWriter, os.Stdout)

	// Create logger
	logger := logging.NewLogger(logging.INFO, multiWriter)

	// Enable sampling to reduce log volume (keep 10% of INFO, all errors)
	logger.EnableSampling(0.1, true)

	// Add Sentry hook for error tracking
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		sentryHook, err := logging.NewSentryHook(dsn, "production")
		if err == nil {
			logger.AddHook(sentryHook)
			logging.Info("Sentry integration enabled")
		}
	}

	// Set as default logger
	logging.SetLevel(logging.INFO)

	// Now all package-level logging functions use this configuration
	logging.Info("Production logging initialized",
		logging.String("environment", "production"),
		logging.String("log_file", "./logs/production.log"),
		logging.Bool("sampling_enabled", true),
		logging.Float64("sampling_rate", 0.1),
	)
}
