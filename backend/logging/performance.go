package logging

import (
	"context"
	"sync"
	"time"
)

// PerformanceMetrics tracks performance metrics for logging
type PerformanceMetrics struct {
	mu              sync.RWMutex
	slowQueries     []*SlowQuery
	slowEndpoints   []*SlowEndpoint
	slowQueryThreshold     time.Duration
	slowEndpointThreshold  time.Duration
}

// SlowQuery represents a slow database query
type SlowQuery struct {
	Query       string
	Duration    time.Duration
	Timestamp   time.Time
	Context     map[string]interface{}
	StackTrace  string
}

// SlowEndpoint represents a slow HTTP endpoint
type SlowEndpoint struct {
	Method      string
	Path        string
	Duration    time.Duration
	Timestamp   time.Time
	StatusCode  int
	RequestID   string
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		slowQueries:            make([]*SlowQuery, 0),
		slowEndpoints:          make([]*SlowEndpoint, 0),
		slowQueryThreshold:     100 * time.Millisecond,
		slowEndpointThreshold:  1000 * time.Millisecond,
	}
}

// LogSlowQuery logs a slow database query
func (pm *PerformanceMetrics) LogSlowQuery(ctx context.Context, query string, duration time.Duration, logger *Logger) {
	if duration < pm.slowQueryThreshold {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	sq := &SlowQuery{
		Query:      query,
		Duration:   duration,
		Timestamp:  time.Now(),
		Context:    make(map[string]interface{}),
		StackTrace: getStackTrace(),
	}

	pm.slowQueries = append(pm.slowQueries, sq)

	// Keep only last 100 slow queries
	if len(pm.slowQueries) > 100 {
		pm.slowQueries = pm.slowQueries[1:]
	}

	// Log the slow query
	logger.Warn("Slow Database Query",
		String("query", truncateString(query, 200)),
		Float64("duration_ms", float64(duration.Milliseconds())),
		String("threshold_ms", pm.slowQueryThreshold.String()),
	)
}

// LogSlowEndpoint logs a slow HTTP endpoint
func (pm *PerformanceMetrics) LogSlowEndpoint(method, path string, duration time.Duration, statusCode int, requestID string, logger *Logger) {
	if duration < pm.slowEndpointThreshold {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	se := &SlowEndpoint{
		Method:     method,
		Path:       path,
		Duration:   duration,
		Timestamp:  time.Now(),
		StatusCode: statusCode,
		RequestID:  requestID,
	}

	pm.slowEndpoints = append(pm.slowEndpoints, se)

	// Keep only last 100 slow endpoints
	if len(pm.slowEndpoints) > 100 {
		pm.slowEndpoints = pm.slowEndpoints[1:]
	}

	// Log the slow endpoint
	logger.Warn("Slow HTTP Endpoint",
		String("method", method),
		String("path", path),
		Float64("duration_ms", float64(duration.Milliseconds())),
		Int("status_code", statusCode),
		RequestID(requestID),
		String("threshold_ms", pm.slowEndpointThreshold.String()),
	)
}

// GetSlowQueries returns recent slow queries
func (pm *PerformanceMetrics) GetSlowQueries() []*SlowQuery {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	queries := make([]*SlowQuery, len(pm.slowQueries))
	copy(queries, pm.slowQueries)
	return queries
}

// GetSlowEndpoints returns recent slow endpoints
func (pm *PerformanceMetrics) GetSlowEndpoints() []*SlowEndpoint {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	endpoints := make([]*SlowEndpoint, len(pm.slowEndpoints))
	copy(endpoints, pm.slowEndpoints)
	return endpoints
}

// SetSlowQueryThreshold sets the threshold for slow query detection
func (pm *PerformanceMetrics) SetSlowQueryThreshold(threshold time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.slowQueryThreshold = threshold
}

// SetSlowEndpointThreshold sets the threshold for slow endpoint detection
func (pm *PerformanceMetrics) SetSlowEndpointThreshold(threshold time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.slowEndpointThreshold = threshold
}

// truncateString truncates a string to maxLen
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Global performance metrics instance
var globalPerfMetrics = NewPerformanceMetrics()

// LogSlowQuery logs a slow query using the global metrics tracker
func LogSlowQuery(ctx context.Context, query string, duration time.Duration) {
	globalPerfMetrics.LogSlowQuery(ctx, query, duration, defaultLogger)
}

// LogSlowEndpoint logs a slow endpoint using the global metrics tracker
func LogSlowEndpoint(method, path string, duration time.Duration, statusCode int, requestID string) {
	globalPerfMetrics.LogSlowEndpoint(method, path, duration, statusCode, requestID, defaultLogger)
}

// GetSlowQueries returns global slow queries
func GetSlowQueries() []*SlowQuery {
	return globalPerfMetrics.GetSlowQueries()
}

// GetSlowEndpoints returns global slow endpoints
func GetSlowEndpoints() []*SlowEndpoint {
	return globalPerfMetrics.GetSlowEndpoints()
}
