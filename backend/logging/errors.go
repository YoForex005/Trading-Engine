package logging

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ErrorTracker tracks and aggregates errors for alerting
type ErrorTracker struct {
	mu               sync.RWMutex
	errors           map[string]*ErrorStats
	alertThresholds  map[string]int
	alertCallbacks   []AlertCallback
	cleanupInterval  time.Duration
	retentionPeriod  time.Duration
	stopChan         chan struct{}
}

// ErrorStats tracks statistics for a specific error
type ErrorStats struct {
	ErrorType      string
	Message        string
	Count          int64
	FirstSeen      time.Time
	LastSeen       time.Time
	Occurrences    []time.Time
	Contexts       []map[string]interface{}
	StackTraces    []string
	AffectedUsers  map[string]bool
	Severity       string
	Alerted        bool
}

// AlertCallback is called when an error threshold is exceeded
type AlertCallback func(stats *ErrorStats)

// NewErrorTracker creates a new error tracker
func NewErrorTracker() *ErrorTracker {
	et := &ErrorTracker{
		errors: make(map[string]*ErrorStats),
		alertThresholds: map[string]int{
			"critical": 1,   // Alert immediately for critical errors
			"high":     5,   // Alert after 5 occurrences
			"medium":   10,  // Alert after 10 occurrences
			"low":      50,  // Alert after 50 occurrences
		},
		cleanupInterval: 5 * time.Minute,
		retentionPeriod: 1 * time.Hour,
		stopChan:        make(chan struct{}),
	}

	// Start cleanup goroutine
	go et.cleanupLoop()

	return et
}

// Track records an error occurrence
func (et *ErrorTracker) Track(ctx context.Context, err error, severity string, extra map[string]interface{}) {
	if err == nil {
		return
	}

	errorKey := fmt.Sprintf("%s:%s", severity, err.Error())

	et.mu.Lock()
	defer et.mu.Unlock()

	stats, exists := et.errors[errorKey]
	if !exists {
		stats = &ErrorStats{
			ErrorType:     getErrorType(err),
			Message:       err.Error(),
			FirstSeen:     time.Now(),
			Contexts:      make([]map[string]interface{}, 0),
			StackTraces:   make([]string, 0),
			AffectedUsers: make(map[string]bool),
			Severity:      severity,
		}
		et.errors[errorKey] = stats
	}

	stats.Count++
	stats.LastSeen = time.Now()
	stats.Occurrences = append(stats.Occurrences, time.Now())

	// Store context information
	if extra != nil {
		stats.Contexts = append(stats.Contexts, extra)
	}

	// Track affected users
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		stats.AffectedUsers[userID] = true
	}

	// Store stack trace for new occurrences (limit to last 10)
	if len(stats.StackTraces) < 10 {
		stats.StackTraces = append(stats.StackTraces, getStackTrace())
	}

	// Check alert threshold
	threshold := et.alertThresholds[severity]
	if !stats.Alerted && stats.Count >= int64(threshold) {
		stats.Alerted = true
		et.triggerAlerts(stats)
	}
}

// RegisterAlertCallback adds a callback for error alerts
func (et *ErrorTracker) RegisterAlertCallback(callback AlertCallback) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.alertCallbacks = append(et.alertCallbacks, callback)
}

// GetStats returns current error statistics
func (et *ErrorTracker) GetStats() map[string]*ErrorStats {
	et.mu.RLock()
	defer et.mu.RUnlock()

	// Create a copy
	stats := make(map[string]*ErrorStats)
	for k, v := range et.errors {
		statsCopy := *v
		stats[k] = &statsCopy
	}

	return stats
}

// GetTopErrors returns the top N errors by count
func (et *ErrorTracker) GetTopErrors(n int) []*ErrorStats {
	et.mu.RLock()
	defer et.mu.RUnlock()

	var errors []*ErrorStats
	for _, stats := range et.errors {
		errors = append(errors, stats)
	}

	// Sort by count (simple bubble sort for small n)
	for i := 0; i < len(errors)-1; i++ {
		for j := i + 1; j < len(errors); j++ {
			if errors[j].Count > errors[i].Count {
				errors[i], errors[j] = errors[j], errors[i]
			}
		}
	}

	if n > len(errors) {
		n = len(errors)
	}

	return errors[:n]
}

// Clear resets all error statistics
func (et *ErrorTracker) Clear() {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.errors = make(map[string]*ErrorStats)
}

// Stop stops the error tracker cleanup loop
func (et *ErrorTracker) Stop() {
	close(et.stopChan)
}

// triggerAlerts calls all registered alert callbacks
func (et *ErrorTracker) triggerAlerts(stats *ErrorStats) {
	for _, callback := range et.alertCallbacks {
		go callback(stats) // Run in goroutine to avoid blocking
	}
}

// cleanupLoop periodically removes old error data
func (et *ErrorTracker) cleanupLoop() {
	ticker := time.NewTicker(et.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			et.cleanup()
		case <-et.stopChan:
			return
		}
	}
}

// cleanup removes errors older than retention period
func (et *ErrorTracker) cleanup() {
	et.mu.Lock()
	defer et.mu.Unlock()

	cutoff := time.Now().Add(-et.retentionPeriod)
	for key, stats := range et.errors {
		if stats.LastSeen.Before(cutoff) {
			delete(et.errors, key)
		}
	}
}

// getErrorType extracts error type from error
func getErrorType(err error) string {
	return fmt.Sprintf("%T", err)
}

// Global error tracker
var globalErrorTracker = NewErrorTracker()

// TrackError tracks an error in the global tracker
func TrackError(ctx context.Context, err error, severity string, extra map[string]interface{}) {
	globalErrorTracker.Track(ctx, err, severity, extra)
}

// GetErrorStats returns global error statistics
func GetErrorStats() map[string]*ErrorStats {
	return globalErrorTracker.GetStats()
}

// GetTopErrors returns top errors from global tracker
func GetTopErrors(n int) []*ErrorStats {
	return globalErrorTracker.GetTopErrors(n)
}

// RegisterErrorAlert registers a global error alert callback
func RegisterErrorAlert(callback AlertCallback) {
	globalErrorTracker.RegisterAlertCallback(callback)
}
