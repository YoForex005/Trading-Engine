package monitoring

import (
	"net/http"
	"time"
)

// Example integration code showing how to use the monitoring package

// InitializeMonitoring sets up all monitoring components
func InitializeMonitoring(version string) {
	// 1. Setup structured logger
	logger := NewLogger("trading-engine")
	logger.SetMinLevel(INFO)
	SetGlobalLogger(logger)

	// 2. Setup health checker
	healthChecker := NewHealthChecker(version)
	SetGlobalHealthChecker(healthChecker)

	// Register health checks
	healthChecker.RegisterCheck("memory", MemoryHealthCheck(80.0))
	healthChecker.RegisterCheck("goroutines", GoroutineHealthCheck(10000))
	healthChecker.RegisterCheck("uptime", UptimeHealthCheck(time.Now(), 30*time.Second))

	// 3. Setup tracer
	tracer := NewTracer("trading-engine")
	SetGlobalTracer(tracer)

	// 4. Setup alert manager
	alertManager := NewAlertManager()
	SetGlobalAlertManager(alertManager)

	// Register default alert rules
	for _, rule := range GetDefaultAlertRules() {
		alertManager.RegisterRule(rule)
	}

	// 5. Start runtime metrics collector
	runtimeCollector := NewRuntimeMetricsCollector(30 * time.Second)
	go runtimeCollector.Start()

	logger.Info("Monitoring initialized", map[string]interface{}{
		"version": version,
		"components": []string{
			"logger",
			"health_checker",
			"tracer",
			"alert_manager",
			"runtime_collector",
		},
	})
}

// RegisterMonitoringEndpoints registers monitoring HTTP endpoints
func RegisterMonitoringEndpoints(mux *http.ServeMux) {
	metricsCollector := NewMetricsCollector()
	healthChecker := GetHealthChecker()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", metricsCollector.Handler())

	// Health check endpoints
	mux.HandleFunc("/health", healthChecker.HTTPHealthHandler())
	mux.HandleFunc("/ready", healthChecker.HTTPReadinessHandler())

	GetLogger().Info("Monitoring endpoints registered", map[string]interface{}{
		"endpoints": []string{"/metrics", "/health", "/ready"},
	})
}

// ExampleOrderExecution shows how to monitor order execution
func ExampleOrderExecution() {
	// Start tracing
	span := TraceOrderExecution("ORD-12345", "EURUSD", "MARKET")
	defer span.Finish()

	span.SetTag("account_id", "demo_001")
	span.SetTag("volume", 1.0)

	// Record start time
	startTime := time.Now()

	// Simulate order execution
	time.Sleep(50 * time.Millisecond)

	// Calculate latency
	latencyMs := float64(time.Since(startTime).Milliseconds())

	// Record metrics
	RecordOrderExecution("MARKET", "EURUSD", "ABOOK", latencyMs, true)
	RecordTradeVolume("EURUSD", "BUY", "ABOOK", 1.0)

	// Log event
	logger := GetLogger()
	logger.OrderLog("ORD-12345", "EURUSD", "BUY", "MARKET", "FILLED", 1.0, 1.1050, map[string]interface{}{
		"execution_time_ms": latencyMs,
		"lp_name":           "OANDA",
	})

	span.LogFields(map[string]interface{}{
		"execution_price": 1.1050,
		"latency_ms":      latencyMs,
	})
}

// ExampleAPIRequest shows how to monitor API requests
func ExampleAPIRequest(w http.ResponseWriter, r *http.Request) {
	// Start tracing
	span := TraceAPIRequest(r.Method, r.URL.Path)
	defer span.Finish()

	startTime := time.Now()

	// Process request
	// ... your handler logic here ...

	// Calculate duration
	durationMs := float64(time.Since(startTime).Milliseconds())

	// Record metrics
	RecordAPIRequest(r.URL.Path, r.Method, "200", durationMs)

	// Log request
	logger := GetLogger()
	logger.Info("API request processed", map[string]interface{}{
		"method":       r.Method,
		"path":         r.URL.Path,
		"duration_ms":  durationMs,
		"status":       200,
		"trace_id":     span.TraceID,
		"span_id":      span.SpanID,
	})
}

// ExampleLPMonitoring shows how to monitor LP connectivity
func ExampleLPMonitoring() {
	lpName := "OANDA"

	// Track connection status
	SetLPConnected(lpName, "FIX", true)

	// Start tracing LP communication
	span := TraceLPCommunication(lpName, "market_data_request")
	defer span.Finish()

	startTime := time.Now()

	// Simulate LP request
	time.Sleep(25 * time.Millisecond)

	latencyMs := float64(time.Since(startTime).Milliseconds())

	// Record metrics
	RecordLPLatency(lpName, "market_data", latencyMs)
	RecordLPQuote(lpName, "EURUSD")

	// Log event
	logger := GetLogger()
	logger.Info("LP quote received", map[string]interface{}{
		"lp_name":    lpName,
		"symbol":     "EURUSD",
		"latency_ms": latencyMs,
		"bid":        1.1050,
		"ask":        1.1052,
	})

	// Check if latency is high
	if latencyMs > 100 {
		logger.Warn("High LP latency detected", map[string]interface{}{
			"lp_name":    lpName,
			"latency_ms": latencyMs,
			"threshold":  100,
		})

		// Fire alert
		alertManager := GetAlertManager()
		alertManager.FireAlert(&Alert{
			Name:      "HighLPLatency",
			Severity:  SeverityWarning,
			Message:   "LP latency exceeded threshold",
			Timestamp: time.Now(),
			Labels: map[string]string{
				"lp_name": lpName,
			},
			Annotations: map[string]string{
				"latency_ms": "150",
				"threshold":  "100",
			},
		})
	}
}

// ExampleHealthCheck shows how to register custom health checks
func ExampleHealthCheck() {
	healthChecker := GetHealthChecker()

	// Register database health check
	healthChecker.RegisterCheck("database", func() ComponentHealth {
		// Check database connection
		connected := true // Replace with actual check

		if !connected {
			return ComponentHealth{
				Status:      StatusUnhealthy,
				Message:     "Database connection failed",
				LastChecked: time.Now(),
			}
		}

		return ComponentHealth{
			Status:      StatusHealthy,
			Message:     "Database connected",
			LastChecked: time.Now(),
			Metadata: map[string]interface{}{
				"connection_pool_size": 10,
				"active_connections":   5,
			},
		}
	})

	// Register LP connectivity check
	healthChecker.RegisterCheck("lp_connectivity", func() ComponentHealth {
		lpConnected := true // Replace with actual LP status check

		status := StatusHealthy
		message := "All LPs connected"

		if !lpConnected {
			status = StatusUnhealthy
			message = "LP connection lost"
		}

		return ComponentHealth{
			Status:      status,
			Message:     message,
			LastChecked: time.Now(),
			Metadata: map[string]interface{}{
				"lp_count":      2,
				"connected_lps": 2,
			},
		}
	})

	// Register WebSocket health check
	healthChecker.RegisterCheck("websocket", func() ComponentHealth {
		activeConnections := 10 // Replace with actual WebSocket connection count

		status := StatusHealthy
		message := "WebSocket running"

		if activeConnections == 0 {
			status = StatusDegraded
			message = "No active WebSocket connections"
		}

		return ComponentHealth{
			Status:      status,
			Message:     message,
			LastChecked: time.Now(),
			Metadata: map[string]interface{}{
				"active_connections": activeConnections,
			},
		}
	})
}

// ExamplePerformanceMonitoring shows how to monitor performance
func ExamplePerformanceMonitoring() {
	logger := GetLogger()

	// Monitor database query
	span := TraceDBQuery("SELECT", "positions")
	defer span.Finish()

	startTime := time.Now()

	// Simulate query
	time.Sleep(15 * time.Millisecond)

	durationMs := float64(time.Since(startTime).Milliseconds())

	// Record metrics
	RecordDBQuery("SELECT", "positions", durationMs)

	// Log performance
	logger.PerformanceLog("db_query_positions", durationMs, map[string]interface{}{
		"operation": "SELECT",
		"table":     "positions",
		"rows":      25,
	})

	// Alert if slow
	if durationMs > 100 {
		logger.Warn("Slow database query detected", map[string]interface{}{
			"operation":   "SELECT",
			"table":       "positions",
			"duration_ms": durationMs,
			"threshold":   100,
		})
	}
}

// ExampleAccountMetrics shows how to track account metrics
func ExampleAccountMetrics() {
	accountID := "demo_001"

	// Update account metrics
	SetAccountBalance(accountID, "demo", 5000.0)
	SetAccountEquity(accountID, 5150.50)
	SetAccountMarginUsed(accountID, 500.0)

	// Log account update
	logger := GetLogger()
	logger.Info("Account metrics updated", map[string]interface{}{
		"account_id":   accountID,
		"balance":      5000.0,
		"equity":       5150.50,
		"margin_used":  500.0,
		"margin_level": (5150.50 / 500.0) * 100, // 1030%
	})
}

// ExampleSecurityLogging shows how to log security events
func ExampleSecurityLogging() {
	logger := GetLogger()

	// Log successful login
	logger.SecurityLog("login", "user_123", "192.168.1.100", "LOGIN", true, map[string]interface{}{
		"user_agent": "Mozilla/5.0",
		"session_id": "sess_abc123",
	})

	// Log failed login attempt
	logger.SecurityLog("login_failed", "user_123", "192.168.1.100", "LOGIN", false, map[string]interface{}{
		"reason":     "invalid_password",
		"user_agent": "Mozilla/5.0",
	})

	// Log suspicious activity
	logger.SecurityLog("suspicious_activity", "user_123", "192.168.1.100", "MULTIPLE_FAILED_LOGINS", false, map[string]interface{}{
		"failed_attempts": 5,
		"time_window":     "5m",
	})
}

// WrapHandlerWithMonitoring wraps an HTTP handler with monitoring
func WrapHandlerWithMonitoring(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start tracing
		span := TraceAPIRequest(r.Method, endpoint)
		defer span.Finish()

		// Add trace ID to response headers
		w.Header().Set("X-Trace-ID", span.TraceID)

		startTime := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call original handler
		handler(wrapped, r)

		// Calculate duration
		durationMs := float64(time.Since(startTime).Milliseconds())

		// Record metrics
		RecordAPIRequest(endpoint, r.Method, http.StatusText(wrapped.statusCode), durationMs)

		// Log request
		logger := GetLogger()
		logger.Info("API request", map[string]interface{}{
			"method":      r.Method,
			"endpoint":    endpoint,
			"status":      wrapped.statusCode,
			"duration_ms": durationMs,
			"trace_id":    span.TraceID,
			"ip":          r.RemoteAddr,
		})

		// Alert on slow requests
		if durationMs > 1000 {
			logger.Warn("Slow API request", map[string]interface{}{
				"endpoint":    endpoint,
				"duration_ms": durationMs,
				"threshold":   1000,
			})
		}
	}
}
