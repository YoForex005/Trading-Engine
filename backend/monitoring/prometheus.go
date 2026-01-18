package monitoring

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Order Execution Metrics
	orderLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trading_order_execution_latency_milliseconds",
			Help:    "Order execution latency in milliseconds (p50, p95, p99)",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		},
		[]string{"order_type", "symbol", "execution_mode"},
	)

	orderTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_orders_total",
			Help: "Total number of orders by type and status",
		},
		[]string{"order_type", "status", "execution_mode"},
	)

	orderErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_order_errors_total",
			Help: "Total number of order errors by type",
		},
		[]string{"order_type", "error_type"},
	)

	// WebSocket Metrics
	wsConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "trading_websocket_connections",
			Help: "Current number of active WebSocket connections",
		},
	)

	wsMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_websocket_messages_total",
			Help: "Total WebSocket messages sent by type",
		},
		[]string{"message_type"},
	)

	// Position Metrics
	activePositions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_active_positions",
			Help: "Number of active positions by symbol",
		},
		[]string{"symbol", "side"},
	)

	positionPnL = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_position_pnl_usd",
			Help: "Current P&L of positions in USD",
		},
		[]string{"account_id", "symbol"},
	)

	// Trade Volume Metrics
	tradeVolume = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_volume_lots_total",
			Help: "Total trading volume in lots",
		},
		[]string{"symbol", "side", "execution_mode"},
	)

	tradePnL = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_pnl_usd_total",
			Help: "Total realized P&L in USD",
		},
		[]string{"account_id", "symbol"},
	)

	// API Request Metrics
	apiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_api_requests_total",
			Help: "Total API requests by endpoint and status",
		},
		[]string{"endpoint", "method", "status"},
	)

	apiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trading_api_request_duration_milliseconds",
			Help:    "API request duration in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
		[]string{"endpoint", "method"},
	)

	// Database Metrics
	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trading_db_query_duration_milliseconds",
			Help:    "Database query duration in milliseconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 25, 50, 100, 250},
		},
		[]string{"operation", "table"},
	)

	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "trading_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	// LP Connectivity Metrics
	lpConnected = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_lp_connected",
			Help: "LP connection status (1=connected, 0=disconnected)",
		},
		[]string{"lp_name", "connection_type"},
	)

	lpLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trading_lp_latency_milliseconds",
			Help:    "LP response latency in milliseconds",
			Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
		},
		[]string{"lp_name", "operation"},
	)

	lpQuotesReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_lp_quotes_received_total",
			Help: "Total quotes received from LP",
		},
		[]string{"lp_name", "symbol"},
	)

	// Memory and Runtime Metrics
	memoryUsageBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "trading_memory_usage_bytes",
			Help: "Current memory usage in bytes",
		},
	)

	goroutineCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "trading_goroutines_count",
			Help: "Current number of goroutines",
		},
	)

	// Account Metrics
	accountBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_account_balance_usd",
			Help: "Current account balance in USD",
		},
		[]string{"account_id", "account_type"},
	)

	accountEquity = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_account_equity_usd",
			Help: "Current account equity in USD",
		},
		[]string{"account_id"},
	)

	accountMarginUsed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "trading_account_margin_used_usd",
			Help: "Current margin used in USD",
		},
		[]string{"account_id"},
	)

	// SLO Metrics
	sloOrderExecutionSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_slo_order_execution_success_total",
			Help: "Total successful order executions (for SLO calculation)",
		},
		[]string{"execution_mode"},
	)

	sloOrderExecutionLatencyTarget = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trading_slo_order_execution_within_target_total",
			Help: "Orders executed within latency target (100ms)",
		},
		[]string{"execution_mode"},
	)
)

// MetricsCollector handles metrics collection and exposure
type MetricsCollector struct {
	registry *prometheus.Registry
	mu       sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
}

// Handler returns the HTTP handler for /metrics endpoint
func (mc *MetricsCollector) Handler() http.Handler {
	return promhttp.Handler()
}

// RecordOrderExecution records order execution metrics
func RecordOrderExecution(orderType, symbol, executionMode string, latencyMs float64, success bool) {
	orderLatency.WithLabelValues(orderType, symbol, executionMode).Observe(latencyMs)

	status := "success"
	if !success {
		status = "failed"
	}
	orderTotal.WithLabelValues(orderType, status, executionMode).Inc()

	// SLO tracking
	if success {
		sloOrderExecutionSuccess.WithLabelValues(executionMode).Inc()
		if latencyMs <= 100 {
			sloOrderExecutionLatencyTarget.WithLabelValues(executionMode).Inc()
		}
	}
}

// RecordOrderError records order error
func RecordOrderError(orderType, errorType string) {
	orderErrors.WithLabelValues(orderType, errorType).Inc()
}

// SetWebSocketConnections sets the current WebSocket connection count
func SetWebSocketConnections(count int) {
	wsConnections.Set(float64(count))
}

// RecordWebSocketMessage records WebSocket message
func RecordWebSocketMessage(messageType string) {
	wsMessagesTotal.WithLabelValues(messageType).Inc()
}

// SetActivePositions sets active position count
func SetActivePositions(symbol, side string, count int) {
	activePositions.WithLabelValues(symbol, side).Set(float64(count))
}

// SetPositionPnL sets position P&L
func SetPositionPnL(accountID, symbol string, pnl float64) {
	positionPnL.WithLabelValues(accountID, symbol).Set(pnl)
}

// RecordTradeVolume records trading volume
func RecordTradeVolume(symbol, side, executionMode string, volumeLots float64) {
	tradeVolume.WithLabelValues(symbol, side, executionMode).Add(volumeLots)
}

// RecordTradePnL records realized P&L
func RecordTradePnL(accountID, symbol string, pnl float64) {
	tradePnL.WithLabelValues(accountID, symbol).Add(pnl)
}

// RecordAPIRequest records API request metrics
func RecordAPIRequest(endpoint, method, status string, durationMs float64) {
	apiRequestsTotal.WithLabelValues(endpoint, method, status).Inc()
	apiRequestDuration.WithLabelValues(endpoint, method).Observe(durationMs)
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, durationMs float64) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(durationMs)
}

// SetDBConnections sets active database connections
func SetDBConnections(count int) {
	dbConnectionsActive.Set(float64(count))
}

// SetLPConnected sets LP connection status
func SetLPConnected(lpName, connectionType string, connected bool) {
	value := 0.0
	if connected {
		value = 1.0
	}
	lpConnected.WithLabelValues(lpName, connectionType).Set(value)
}

// RecordLPLatency records LP latency
func RecordLPLatency(lpName, operation string, latencyMs float64) {
	lpLatency.WithLabelValues(lpName, operation).Observe(latencyMs)
}

// RecordLPQuote records LP quote reception
func RecordLPQuote(lpName, symbol string) {
	lpQuotesReceived.WithLabelValues(lpName, symbol).Inc()
}

// SetMemoryUsage sets memory usage
func SetMemoryUsage(bytes uint64) {
	memoryUsageBytes.Set(float64(bytes))
}

// SetGoroutineCount sets goroutine count
func SetGoroutineCount(count int) {
	goroutineCount.Set(float64(count))
}

// SetAccountBalance sets account balance
func SetAccountBalance(accountID, accountType string, balance float64) {
	accountBalance.WithLabelValues(accountID, accountType).Set(balance)
}

// SetAccountEquity sets account equity
func SetAccountEquity(accountID string, equity float64) {
	accountEquity.WithLabelValues(accountID).Set(equity)
}

// SetAccountMarginUsed sets account margin used
func SetAccountMarginUsed(accountID string, marginUsed float64) {
	accountMarginUsed.WithLabelValues(accountID).Set(marginUsed)
}

// APIRequestMiddleware wraps HTTP handlers to record metrics
func APIRequestMiddleware(endpoint string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		handler(wrapped, r)

		duration := float64(time.Since(start).Milliseconds())
		RecordAPIRequest(endpoint, r.Method, http.StatusText(wrapped.statusCode), duration)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
