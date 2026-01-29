package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// LogEntry represents a structured log entry compatible with ELK, Datadog, CloudWatch
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	AccountID   string                 `json:"account_id,omitempty"`
	TradeID     string                 `json:"trade_id,omitempty"`
	OrderID     string                 `json:"order_id,omitempty"`
	Symbol      string                 `json:"symbol,omitempty"`
	Component   string                 `json:"component,omitempty"`
	Function    string                 `json:"function,omitempty"`
	File        string                 `json:"file,omitempty"`
	Line        int                    `json:"line,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Duration    float64                `json:"duration_ms,omitempty"` // For performance logging
	Extra       map[string]interface{} `json:"extra,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	Hostname    string                 `json:"hostname,omitempty"`
	PID         int                    `json:"pid,omitempty"`
}

// Logger provides structured logging with multiple outputs
type Logger struct {
	mu          sync.RWMutex
	level       LogLevel
	outputs     []io.Writer
	hooks       []Hook
	environment string
	hostname    string
	pid         int
	sampling    *SamplingConfig
}

// SamplingConfig controls log sampling to reduce volume in production
type SamplingConfig struct {
	Enabled     bool
	Rate        float64 // 0.0 to 1.0 - percentage of logs to keep
	KeepErrors  bool    // Always keep ERROR and FATAL logs
	SampleCount int64
	mu          sync.Mutex
}

// Hook allows external integrations (Sentry, custom handlers)
type Hook interface {
	Fire(entry *LogEntry) error
	Levels() []LogLevel
}

// NewLogger creates a new structured logger
func NewLogger(level LogLevel, outputs ...io.Writer) *Logger {
	if len(outputs) == 0 {
		outputs = []io.Writer{os.Stdout}
	}

	hostname, _ := os.Hostname()

	return &Logger{
		level:       level,
		outputs:     outputs,
		environment: getEnvironment(),
		hostname:    hostname,
		pid:         os.Getpid(),
		sampling: &SamplingConfig{
			Enabled:    false,
			Rate:       1.0,
			KeepErrors: true,
		},
	}
}

// SetLevel changes the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// AddHook registers a log hook for external integrations
func (l *Logger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// EnableSampling enables log sampling for production
func (l *Logger) EnableSampling(rate float64, keepErrors bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sampling.Enabled = true
	l.sampling.Rate = rate
	l.sampling.KeepErrors = keepErrors
}

// DisableSampling disables log sampling
func (l *Logger) DisableSampling() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.sampling.Enabled = false
}

// WithContext creates a logger with context values
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: l,
		ctx:    ctx,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...Field) {
	l.log(DEBUG, message, nil, fields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...Field) {
	l.log(INFO, message, nil, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...Field) {
	l.log(WARN, message, nil, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields ...Field) {
	l.log(ERROR, message, err, fields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, fields ...Field) {
	l.log(FATAL, message, err, fields...)
	os.Exit(1)
}

// log is the internal logging implementation
func (l *Logger) log(level LogLevel, message string, err error, fields ...Field) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}

	// Check sampling
	if l.sampling.Enabled && !l.shouldSample(level) {
		l.mu.RUnlock()
		return
	}
	l.mu.RUnlock()

	// Build log entry
	entry := &LogEntry{
		Timestamp:   time.Now().UTC(),
		Level:       levelNames[level],
		Message:     message,
		Environment: l.environment,
		Hostname:    l.hostname,
		PID:         l.pid,
		Extra:       make(map[string]interface{}),
	}

	// Add caller information
	if pc, file, line, ok := runtime.Caller(2); ok {
		entry.File = trimPath(file)
		entry.Line = line
		if fn := runtime.FuncForPC(pc); fn != nil {
			entry.Function = trimFunctionName(fn.Name())
		}
	}

	// Add error information
	if err != nil {
		entry.Error = err.Error()
		if level >= ERROR {
			entry.StackTrace = getStackTrace()
		}
	}

	// Add custom fields
	for _, field := range fields {
		field.Apply(entry)
	}

	// Execute hooks
	l.mu.RLock()
	for _, hook := range l.hooks {
		if containsLevel(hook.Levels(), level) {
			_ = hook.Fire(entry) // Ignore hook errors to prevent log failures
		}
	}
	l.mu.RUnlock()

	// Write to outputs
	l.writeEntry(entry)
}

// shouldSample determines if a log should be sampled
func (l *Logger) shouldSample(level LogLevel) bool {
	if !l.sampling.Enabled {
		return true
	}

	// Always keep errors and fatals
	if l.sampling.KeepErrors && (level >= ERROR) {
		return true
	}

	l.sampling.mu.Lock()
	defer l.sampling.mu.Unlock()

	l.sampling.SampleCount++
	// Simple sampling: keep every Nth log based on rate
	threshold := int64(1.0 / l.sampling.Rate)
	return l.sampling.SampleCount%threshold == 0
}

// writeEntry writes the log entry to all outputs
func (l *Logger) writeEntry(entry *LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple output if JSON marshaling fails
		data = []byte(fmt.Sprintf(`{"level":"%s","message":"Failed to marshal log: %v"}`, entry.Level, err))
	}
	data = append(data, '\n')

	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, output := range l.outputs {
		_, _ = output.Write(data) // Ignore write errors to prevent cascading failures
	}
}

// ContextLogger wraps Logger with context
type ContextLogger struct {
	logger *Logger
	ctx    context.Context
}

// Debug logs with context
func (cl *ContextLogger) Debug(message string, fields ...Field) {
	fields = append(fields, FieldsFromContext(cl.ctx)...)
	cl.logger.Debug(message, fields...)
}

// Info logs with context
func (cl *ContextLogger) Info(message string, fields ...Field) {
	fields = append(fields, FieldsFromContext(cl.ctx)...)
	cl.logger.Info(message, fields...)
}

// Warn logs with context
func (cl *ContextLogger) Warn(message string, fields ...Field) {
	fields = append(fields, FieldsFromContext(cl.ctx)...)
	cl.logger.Warn(message, fields...)
}

// Error logs with context
func (cl *ContextLogger) Error(message string, err error, fields ...Field) {
	fields = append(fields, FieldsFromContext(cl.ctx)...)
	cl.logger.Error(message, err, fields...)
}

// Fatal logs with context
func (cl *ContextLogger) Fatal(message string, err error, fields ...Field) {
	fields = append(fields, FieldsFromContext(cl.ctx)...)
	cl.logger.Fatal(message, err, fields...)
}

// Helper functions

func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

func trimPath(path string) string {
	// Trim to relative path from project root
	if idx := strings.Index(path, "/backend/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func trimFunctionName(name string) string {
	// Extract just the function name from full path
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}

func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func containsLevel(levels []LogLevel, level LogLevel) bool {
	for _, l := range levels {
		if l == level {
			return true
		}
	}
	return false
}

// Global logger instance (can be replaced)
var defaultLogger = NewLogger(INFO)

// Package-level logging functions for convenience

func Debug(message string, fields ...Field) {
	defaultLogger.Debug(message, fields...)
}

func Info(message string, fields ...Field) {
	defaultLogger.Info(message, fields...)
}

func Warn(message string, fields ...Field) {
	defaultLogger.Warn(message, fields...)
}

func Error(message string, err error, fields ...Field) {
	defaultLogger.Error(message, err, fields...)
}

func Fatal(message string, err error, fields ...Field) {
	defaultLogger.Fatal(message, err, fields...)
}

func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

func AddHook(hook Hook) {
	defaultLogger.AddHook(hook)
}

func WithContext(ctx context.Context) *ContextLogger {
	return defaultLogger.WithContext(ctx)
}
