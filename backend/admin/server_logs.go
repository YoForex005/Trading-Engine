package admin

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LogLevel string

const (
	LogLevelINFO  LogLevel = "INFO"
	LogLevelWARN  LogLevel = "WARN"
	LogLevelERROR LogLevel = "ERROR"
	LogLevelDEBUG LogLevel = "DEBUG"
	LogLevelFATAL LogLevel = "FATAL"
)

type LogEntry struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Level      LogLevel  `json:"level"`
	Source     string    `json:"source"`
	Message    string    `json:"message"`
	StackTrace string    `json:"stackTrace,omitempty"`
	RequestID  string    `json:"requestId,omitempty"`
	Duration   int       `json:"duration,omitempty"` // in milliseconds
}

type LogStats struct {
	ErrorsLastHour   int     `json:"errorsLastHour"`
	WarningsLastHour int     `json:"warningsLastHour"`
	AvgResponseTime  float64 `json:"avgResponseTime"`
	UptimePercent    float64 `json:"uptimePercent"`
}

type ErrorRatePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

type TopError struct {
	Message        string    `json:"message"`
	Count          int       `json:"count"`
	LastOccurrence time.Time `json:"lastOccurrence"`
	Module         string    `json:"module"`
}

type LogStore struct {
	mu      sync.RWMutex
	logs    []*LogEntry
	modules []string
}

var globalLogStore *LogStore

func init() {
	globalLogStore = NewLogStore()
	go globalLogStore.startLogGeneration()
}

func NewLogStore() *LogStore {
	store := &LogStore{
		logs: make([]*LogEntry, 0, 500),
		modules: []string{
			"auth",
			"trading-engine",
			"market-data",
			"websocket",
			"fix-gateway",
			"risk-engine",
			"admin-api",
			"database",
			"scheduler",
			"notifications",
		},
	}

	store.generateInitialLogs()
	return store
}

func (s *LogStore) generateInitialLogs() {
	now := time.Now()

	// Distribution: 70% INFO, 15% WARN, 10% ERROR, 4% DEBUG, 1% FATAL
	distributions := []struct {
		level LogLevel
		count int
	}{
		{LogLevelINFO, 350},
		{LogLevelWARN, 75},
		{LogLevelERROR, 50},
		{LogLevelDEBUG, 20},
		{LogLevelFATAL, 5},
	}

	for _, dist := range distributions {
		for i := 0; i < dist.count; i++ {
			// Spread logs over last 24 hours
			timestamp := now.Add(-time.Duration(rand.Intn(24*60)) * time.Minute)
			module := s.modules[rand.Intn(len(s.modules))]

			entry := &LogEntry{
				ID:        uuid.New().String(),
				Timestamp: timestamp,
				Level:     dist.level,
				Source:    module,
				Message:   s.generateMessage(module, dist.level),
				RequestID: fmt.Sprintf("req-%s", uuid.New().String()[:8]),
				Duration:  rand.Intn(500) + 10,
			}

			if dist.level == LogLevelERROR || dist.level == LogLevelFATAL {
				entry.StackTrace = s.generateStackTrace(module)
			}

			s.logs = append(s.logs, entry)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(s.logs, func(i, j int) bool {
		return s.logs[i].Timestamp.After(s.logs[j].Timestamp)
	})
}

func (s *LogStore) generateMessage(module string, level LogLevel) string {
	messages := map[string]map[LogLevel][]string{
		"auth": {
			LogLevelINFO:  {"Login successful for user admin@rtx5.com", "Session created for user trader123", "Token refreshed for session", "Logout completed", "Password changed successfully"},
			LogLevelWARN:  {"Failed login attempt from IP 192.168.1.45", "Multiple login attempts detected", "Session expired for inactive user", "Weak password detected"},
			LogLevelERROR: {"Authentication failed: invalid credentials", "JWT token validation failed", "Session store connection lost", "User account locked due to multiple failures"},
			LogLevelDEBUG: {"Validating user credentials", "Checking session expiry", "Refreshing JWT token"},
			LogLevelFATAL: {"Critical: Auth service crashed - unable to recover"},
		},
		"trading-engine": {
			LogLevelINFO:  {"Order executed: EURUSD Buy 1.0 lots @1.08245", "Position opened: GBPUSD Sell 0.5 lots", "Position closed: USDJPY #12345 P&L: $125.50", "Stop loss triggered for position #67890", "Take profit hit for order #54321"},
			LogLevelWARN:  {"Margin warning: Account balance low", "Slippage detected: 2 pips on EURUSD order", "Order queue backlog: 15 pending orders", "High volatility detected for XAUUSD"},
			LogLevelERROR: {"Order execution failed: insufficient margin", "Position close error: symbol not found", "Engine synchronization error", "Invalid order parameters received"},
			LogLevelDEBUG: {"Calculating margin requirement", "Validating order against risk limits", "Checking symbol trading hours"},
			LogLevelFATAL: {"Critical: Trading engine core failure - all trading halted"},
		},
		"market-data": {
			LogLevelINFO:  {"Tick received: EURUSD Bid=1.08245 Ask=1.08255", "Quote updated for GBPUSD", "Market session opened: London", "Candle generated for EURUSD M1", "Spread narrowed to 1.0 pips on USDJPY"},
			LogLevelWARN:  {"Delayed tick data received (2s lag)", "High spread detected: XAUUSD 5.0 pips", "No quotes received for AUDNZD (10s)", "Price spike detected: BTCUSD +2%"},
			LogLevelERROR: {"Market data feed disconnected", "Failed to parse incoming tick", "Symbol configuration missing for EURCZK", "Data validation failed: invalid bid/ask spread"},
			LogLevelDEBUG: {"Processing tick for EURUSD", "Updating candle cache", "Broadcasting price update to 45 clients"},
			LogLevelFATAL: {"Critical: All market data feeds offline"},
		},
		"websocket": {
			LogLevelINFO:  {"Client connected: session-abc123", "Client disconnected: session-def456", "Message sent to 23 subscribers", "Heartbeat received from client", "Subscription added: EURUSD quotes"},
			LogLevelWARN:  {"Client connection timeout: session-ghi789", "Slow client detected: message queue at 500", "Reconnection attempt #3 from client", "Invalid message format received"},
			LogLevelERROR: {"WebSocket upgrade failed for client", "Failed to broadcast: client connection closed", "Message serialization error", "Client authentication failed"},
			LogLevelDEBUG: {"Processing subscription request", "Sending tick update to client", "Client keepalive ping"},
			LogLevelFATAL: {"Critical: WebSocket server crashed"},
		},
		"fix-gateway": {
			LogLevelINFO:  {"FIX session established with LP-YOFX", "Market data request sent to LP", "FIX message received: ExecutionReport", "Logon successful to YOFX1", "Heartbeat sent to LP"},
			LogLevelWARN:  {"FIX sequence gap detected: expected 1234, received 1240", "Session not established: retrying connection", "Slow FIX response from LP (2s)", "Market data request rejected by LP"},
			LogLevelERROR: {"FIX connection lost to LP-YOFX", "Failed to parse FIX message", "Invalid FIX message format", "Authentication failed with LP"},
			LogLevelDEBUG: {"Sending FIX NewOrderSingle", "Processing FIX ExecutionReport", "Validating FIX session credentials"},
			LogLevelFATAL: {"Critical: FIX gateway crashed - all LP connections lost"},
		},
		"risk-engine": {
			LogLevelINFO:  {"Risk check passed for order #12345", "Margin call resolved for account #67890", "Exposure updated: EURUSD net -$5,420", "Daily loss limit reset", "Leverage check completed"},
			LogLevelWARN:  {"Margin call triggered for account #45678 (margin level: 85%)", "Risk limit approaching: 80% of max exposure", "High concentration in EURUSD detected", "Drawdown threshold reached: -5%"},
			LogLevelERROR: {"Risk check failed: exposure limit exceeded", "Forced position close failed for account #11111", "Invalid risk parameters in configuration", "Margin calculation error"},
			LogLevelDEBUG: {"Calculating account equity", "Checking exposure limits", "Updating risk metrics"},
			LogLevelFATAL: {"Critical: Risk engine failure - trading suspended"},
		},
		"admin-api": {
			LogLevelINFO:  {"Admin action: Created user trader456", "Config updated: spread markup changed", "Report generated: Daily P&L", "Withdrawal approved: $1,000 for account #99999", "Symbol added: XRPUSD"},
			LogLevelWARN:  {"Concurrent admin modification detected", "Large bulk operation initiated: 500 records", "Deprecated API endpoint called", "Rate limit warning: 80% of quota used"},
			LogLevelERROR: {"Admin action failed: insufficient permissions", "Database update failed for user modification", "Invalid API parameters received", "Export generation failed: timeout"},
			LogLevelDEBUG: {"Validating admin permissions", "Executing database query", "Generating export file"},
			LogLevelFATAL: {"Critical: Admin API crashed"},
		},
		"database": {
			LogLevelINFO:  {"Query executed successfully in 45ms", "Connection pool: 8/20 active", "Transaction committed", "Index rebuilt successfully", "Backup completed"},
			LogLevelWARN:  {"Slow query detected: 2.5s for positions table", "Connection pool near capacity: 18/20", "Table lock timeout", "High disk I/O detected"},
			LogLevelERROR: {"Connection pool exhausted", "Query timeout after 30s", "Foreign key constraint violation", "Deadlock detected and resolved", "Database connection failed"},
			LogLevelDEBUG: {"Opening database connection", "Preparing SQL statement", "Executing parameterized query"},
			LogLevelFATAL: {"Critical: Database connection lost - unable to reconnect"},
		},
		"scheduler": {
			LogLevelINFO:  {"Scheduled task executed: Daily report generation", "Cron job completed: Database cleanup", "Task queued: Send end-of-day emails", "Scheduled tick compression started", "Backup task completed successfully"},
			LogLevelWARN:  {"Task execution delayed: queue backlog", "Previous task still running: skipping iteration", "Task took longer than expected: 5 minutes", "Retry attempt #2 for failed task"},
			LogLevelERROR: {"Task execution failed: Daily report generation", "Scheduled job missed: Email send", "Task timeout after 10 minutes", "Failed to queue task: scheduler busy"},
			LogLevelDEBUG: {"Checking task schedule", "Executing scheduled task", "Updating task status"},
			LogLevelFATAL: {"Critical: Scheduler service crashed"},
		},
		"notifications": {
			LogLevelINFO:  {"Email sent successfully to user@example.com", "Push notification delivered to 45 devices", "SMS sent: Margin call alert", "Webhook triggered: Order executed", "In-app notification created"},
			LogLevelWARN:  {"Email delivery delayed: SMTP queue full", "Push notification failed: device token invalid", "Notification rate limit reached", "Webhook response slow: 3s"},
			LogLevelERROR: {"Failed to send email: SMTP connection refused", "Push notification service unavailable", "SMS delivery failed: invalid phone number", "Webhook timeout after 10s"},
			LogLevelDEBUG: {"Preparing email template", "Sending push to FCM", "Formatting SMS message"},
			LogLevelFATAL: {"Critical: Notification service crashed"},
		},
	}

	moduleMessages, ok := messages[module]
	if !ok {
		return fmt.Sprintf("[%s] Log entry", module)
	}

	levelMessages, ok := moduleMessages[level]
	if !ok || len(levelMessages) == 0 {
		return fmt.Sprintf("[%s] %s: Log entry", module, level)
	}

	return levelMessages[rand.Intn(len(levelMessages))]
}

func (s *LogStore) generateStackTrace(module string) string {
	traces := []string{
		"at OrderExecutor.execute(order_executor.go:145)\nat TradingEngine.processOrder(engine.go:234)\nat main.handleRequest(main.go:89)",
		"at Database.query(database.go:67)\nat Repository.findById(repository.go:123)\nat Handler.getAccount(handler.go:45)",
		"at WebSocket.broadcast(websocket.go:178)\nat Hub.sendMessage(hub.go:234)\nat main.processMessage(main.go:156)",
		"at FIXGateway.connect(fix_gateway.go:89)\nat Session.establish(session.go:234)\nat main.initializeFIX(main.go:67)",
		"at RiskEngine.checkLimits(risk_engine.go:123)\nat Validator.validate(validator.go:78)\nat Handler.createOrder(handler.go:234)",
	}
	return traces[rand.Intn(len(traces))]
}

func (s *LogStore) startLogGeneration() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.addNewLog()
	}
}

func (s *LogStore) addNewLog() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine level based on distribution
	r := rand.Float64()
	var level LogLevel
	if r < 0.70 {
		level = LogLevelINFO
	} else if r < 0.85 {
		level = LogLevelWARN
	} else if r < 0.95 {
		level = LogLevelERROR
	} else if r < 0.99 {
		level = LogLevelDEBUG
	} else {
		level = LogLevelFATAL
	}

	module := s.modules[rand.Intn(len(s.modules))]

	entry := &LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Level:     level,
		Source:    module,
		Message:   s.generateMessage(module, level),
		RequestID: fmt.Sprintf("req-%s", uuid.New().String()[:8]),
		Duration:  rand.Intn(500) + 10,
	}

	if level == LogLevelERROR || level == LogLevelFATAL {
		entry.StackTrace = s.generateStackTrace(module)
	}

	// Add to front and keep max 1000 logs
	s.logs = append([]*LogEntry{entry}, s.logs...)
	if len(s.logs) > 1000 {
		s.logs = s.logs[:1000]
	}
}

// HTTP Handlers

func HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	globalLogStore.mu.RLock()
	defer globalLogStore.mu.RUnlock()

	// Parse query parameters
	level := r.URL.Query().Get("level")
	source := r.URL.Query().Get("source")
	search := r.URL.Query().Get("search")
	timeRange := r.URL.Query().Get("timeRange") // e.g., "1h", "24h", "7d"

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Filter logs
	filtered := make([]*LogEntry, 0)
	cutoffTime := time.Now()
	if timeRange != "" {
		switch timeRange {
		case "1h":
			cutoffTime = time.Now().Add(-1 * time.Hour)
		case "24h":
			cutoffTime = time.Now().Add(-24 * time.Hour)
		case "7d":
			cutoffTime = time.Now().Add(-7 * 24 * time.Hour)
		}
	}

	for _, log := range globalLogStore.logs {
		if level != "" && string(log.Level) != level {
			continue
		}
		if source != "" && log.Source != source {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(search)) {
			continue
		}
		if timeRange != "" && log.Timestamp.Before(cutoffTime) {
			continue
		}
		filtered = append(filtered, log)
	}

	// Paginate
	total := len(filtered)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginated := filtered[start:end]

	response := map[string]interface{}{
		"logs":       paginated,
		"total":      total,
		"page":       page,
		"pageSize":   pageSize,
		"totalPages": (total + pageSize - 1) / pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(response)
}

func HandleGetLogStats(w http.ResponseWriter, r *http.Request) {
	globalLogStore.mu.RLock()
	defer globalLogStore.mu.RUnlock()

	oneHourAgo := time.Now().Add(-1 * time.Hour)

	var errorsLastHour, warningsLastHour int
	var totalDuration int
	var count int

	for _, log := range globalLogStore.logs {
		if log.Timestamp.After(oneHourAgo) {
			if log.Level == LogLevelERROR || log.Level == LogLevelFATAL {
				errorsLastHour++
			}
			if log.Level == LogLevelWARN {
				warningsLastHour++
			}
		}
		if log.Duration > 0 {
			totalDuration += log.Duration
			count++
		}
	}

	avgResponseTime := 0.0
	if count > 0 {
		avgResponseTime = float64(totalDuration) / float64(count)
	}

	// Calculate uptime (mock: based on error rate)
	uptimePercent := 99.9
	if errorsLastHour > 10 {
		uptimePercent = 99.5
	}
	if errorsLastHour > 50 {
		uptimePercent = 98.0
	}

	stats := LogStats{
		ErrorsLastHour:   errorsLastHour,
		WarningsLastHour: warningsLastHour,
		AvgResponseTime:  avgResponseTime,
		UptimePercent:    uptimePercent,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(stats)
}

func HandleGetErrorRate(w http.ResponseWriter, r *http.Request) {
	globalLogStore.mu.RLock()
	defer globalLogStore.mu.RUnlock()

	// Generate 60 data points (errors per minute over last 60 minutes)
	points := make([]ErrorRatePoint, 60)
	now := time.Now()

	for i := 0; i < 60; i++ {
		minuteStart := now.Add(-time.Duration(60-i) * time.Minute)
		minuteEnd := minuteStart.Add(1 * time.Minute)

		count := 0
		for _, log := range globalLogStore.logs {
			if (log.Level == LogLevelERROR || log.Level == LogLevelFATAL) &&
				log.Timestamp.After(minuteStart) && log.Timestamp.Before(minuteEnd) {
				count++
			}
		}

		points[i] = ErrorRatePoint{
			Timestamp: minuteStart,
			Count:     count,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(points)
}

func HandleGetTopErrors(w http.ResponseWriter, r *http.Request) {
	globalLogStore.mu.RLock()
	defer globalLogStore.mu.RUnlock()

	// Count errors by message
	errorCounts := make(map[string]*TopError)

	for _, log := range globalLogStore.logs {
		if log.Level == LogLevelERROR || log.Level == LogLevelFATAL {
			if existing, ok := errorCounts[log.Message]; ok {
				existing.Count++
				if log.Timestamp.After(existing.LastOccurrence) {
					existing.LastOccurrence = log.Timestamp
				}
			} else {
				errorCounts[log.Message] = &TopError{
					Message:        log.Message,
					Count:          1,
					LastOccurrence: log.Timestamp,
					Module:         log.Source,
				}
			}
		}
	}

	// Convert to slice and sort by count
	topErrors := make([]*TopError, 0, len(errorCounts))
	for _, err := range errorCounts {
		topErrors = append(topErrors, err)
	}

	sort.Slice(topErrors, func(i, j int) bool {
		return topErrors[i].Count > topErrors[j].Count
	})

	// Return top 5
	if len(topErrors) > 5 {
		topErrors = topErrors[:5]
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(topErrors)
}

func HandleExportLogs(w http.ResponseWriter, r *http.Request) {
	globalLogStore.mu.RLock()
	defer globalLogStore.mu.RUnlock()

	// Parse query parameters
	level := r.URL.Query().Get("level")
	source := r.URL.Query().Get("source")

	// Filter logs
	filtered := make([]*LogEntry, 0)
	for _, log := range globalLogStore.logs {
		if level != "" && string(log.Level) != level {
			continue
		}
		if source != "" && log.Source != source {
			continue
		}
		filtered = append(filtered, log)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=logs-export.json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	json.NewEncoder(w).Encode(filtered)
}
