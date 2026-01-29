package monitoring

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// LogLevel represents logging severity
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry represents a structured log entry in JSON format
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Service    string                 `json:"service"`
	Message    string                 `json:"message"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Source     string                 `json:"source,omitempty"`
}

// Logger provides structured JSON logging
type Logger struct {
	serviceName string
	output      io.Writer
	minLevel    LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(serviceName string) *Logger {
	return &Logger{
		serviceName: serviceName,
		output:      os.Stdout,
		minLevel:    INFO,
	}
}

// SetOutput sets the logger output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}

// SetMinLevel sets the minimum log level
func (l *Logger) SetMinLevel(level LogLevel) {
	l.minLevel = level
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, err error) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Service:   l.serviceName,
		Message:   message,
		Fields:    fields,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Add source file and line for ERROR and FATAL
	if level == ERROR || level == FATAL {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			entry.Source = fmt.Sprintf("%s:%d", file, line)
		}
	}

	// Add stack trace for FATAL
	if level == FATAL {
		buf := make([]byte, 4096)
		n := runtime.Stack(buf, false)
		entry.StackTrace = string(buf[:n])
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple output if JSON marshaling fails
		fmt.Fprintf(l.output, "[%s] %s: %s (marshal error: %v)\n",
			entry.Timestamp, level, message, err)
		return
	}

	fmt.Fprintln(l.output, string(data))

	if level == FATAL {
		os.Exit(1)
	}
}

// shouldLog checks if message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levelPriority := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}

	return levelPriority[level] >= levelPriority[l.minLevel]
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log(DEBUG, message, fields, nil)
}

// Info logs an info message
func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(INFO, message, fields, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log(WARN, message, fields, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields map[string]interface{}) {
	l.log(ERROR, message, fields, err)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, fields map[string]interface{}) {
	l.log(FATAL, message, fields, err)
}

// WithTracing adds trace and span IDs to log entry
func (l *Logger) WithTracing(level LogLevel, message string, traceID, spanID string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	if fields == nil {
		fields = make(map[string]interface{})
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level,
		Service:   l.serviceName,
		Message:   message,
		Fields:    fields,
		TraceID:   traceID,
		SpanID:    spanID,
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.output, string(data))
}

// OrderLog logs order-related events
func (l *Logger) OrderLog(orderID, symbol, side, orderType, status string, volume, price float64, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["order_id"] = orderID
	fields["symbol"] = symbol
	fields["side"] = side
	fields["order_type"] = orderType
	fields["status"] = status
	fields["volume"] = volume
	fields["price"] = price
	fields["event_type"] = "order"

	l.Info(fmt.Sprintf("Order %s: %s %s %.2f %s @ %.5f", status, side, symbol, volume, orderType, price), fields)
}

// TradeLog logs trade execution events
func (l *Logger) TradeLog(tradeID, orderID, symbol, side string, volume, price, pnl float64, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["trade_id"] = tradeID
	fields["order_id"] = orderID
	fields["symbol"] = symbol
	fields["side"] = side
	fields["volume"] = volume
	fields["price"] = price
	fields["pnl"] = pnl
	fields["event_type"] = "trade"

	l.Info(fmt.Sprintf("Trade executed: %s %s %.2f @ %.5f | P&L: %.2f", side, symbol, volume, price, pnl), fields)
}

// PerformanceLog logs performance metrics
func (l *Logger) PerformanceLog(operation string, durationMs float64, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["operation"] = operation
	fields["duration_ms"] = durationMs
	fields["event_type"] = "performance"

	level := INFO
	if durationMs > 1000 {
		level = WARN
	}

	l.log(level, fmt.Sprintf("Performance: %s took %.2fms", operation, durationMs), fields, nil)
}

// SecurityLog logs security-related events
func (l *Logger) SecurityLog(event, userID, ipAddress, action string, success bool, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}

	fields["security_event"] = event
	fields["user_id"] = userID
	fields["ip_address"] = ipAddress
	fields["action"] = action
	fields["success"] = success
	fields["event_type"] = "security"

	level := INFO
	if !success {
		level = WARN
	}

	l.log(level, fmt.Sprintf("Security: %s - %s by %s from %s", event, action, userID, ipAddress), fields, nil)
}

// Global logger instance
var globalLogger = NewLogger("trading-engine")

// GetLogger returns the global logger
func GetLogger() *Logger {
	return globalLogger
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}
