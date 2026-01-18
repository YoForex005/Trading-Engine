package monitoring

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthCheck represents overall system health
type HealthCheck struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     float64                    `json:"uptime_seconds"`
	Version    string                     `json:"version"`
	Components map[string]ComponentHealth `json:"components"`
}

// ReadinessCheck represents readiness probe status
type ReadinessCheck struct {
	Ready      bool                       `json:"ready"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	startTime  time.Time
	version    string
	checkers   map[string]HealthCheckFunc
	mu         sync.RWMutex
	lastCheck  time.Time
	lastResult *HealthCheck
}

// HealthCheckFunc is a function that performs a health check
type HealthCheckFunc func() ComponentHealth

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		version:   version,
		checkers:  make(map[string]HealthCheckFunc),
	}
}

// RegisterCheck registers a health check for a component
func (hc *HealthChecker) RegisterCheck(name string, checkFunc HealthCheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checkers[name] = checkFunc
}

// Check performs all health checks
func (hc *HealthChecker) Check() *HealthCheck {
	hc.mu.RLock()
	checkers := make(map[string]HealthCheckFunc)
	for k, v := range hc.checkers {
		checkers[k] = v
	}
	hc.mu.RUnlock()

	components := make(map[string]ComponentHealth)
	overallStatus := StatusHealthy

	// Run all health checks
	for name, checkFunc := range checkers {
		componentHealth := checkFunc()
		components[name] = componentHealth

		// Determine overall status
		if componentHealth.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if componentHealth.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	result := &HealthCheck{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     time.Since(hc.startTime).Seconds(),
		Version:    hc.version,
		Components: components,
	}

	hc.mu.Lock()
	hc.lastCheck = time.Now()
	hc.lastResult = result
	hc.mu.Unlock()

	return result
}

// CheckReadiness checks if the system is ready to serve traffic
func (hc *HealthChecker) CheckReadiness() *ReadinessCheck {
	hc.mu.RLock()
	checkers := make(map[string]HealthCheckFunc)
	for k, v := range hc.checkers {
		checkers[k] = v
	}
	hc.mu.RUnlock()

	components := make(map[string]ComponentHealth)
	ready := true

	// Critical components for readiness
	criticalComponents := []string{"database", "lp_connectivity", "websocket"}

	for name, checkFunc := range checkers {
		componentHealth := checkFunc()
		components[name] = componentHealth

		// Check if critical component is unhealthy
		for _, critical := range criticalComponents {
			if name == critical && componentHealth.Status == StatusUnhealthy {
				ready = false
			}
		}
	}

	return &ReadinessCheck{
		Ready:      ready,
		Timestamp:  time.Now(),
		Components: components,
	}
}

// HTTPHealthHandler returns an HTTP handler for /health endpoint
func (hc *HealthChecker) HTTPHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := hc.Check()

		statusCode := http.StatusOK
		if result.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if result.Status == StatusDegraded {
			statusCode = http.StatusOK // Still serving, but degraded
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(result)
	}
}

// HTTPReadinessHandler returns an HTTP handler for /ready endpoint
func (hc *HealthChecker) HTTPReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := hc.CheckReadiness()

		statusCode := http.StatusOK
		if !result.Ready {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(result)
	}
}

// Common health check functions

// MemoryHealthCheck checks memory usage
func MemoryHealthCheck(thresholdPercent float64) HealthCheckFunc {
	return func() ComponentHealth {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		usedMB := float64(m.Alloc) / 1024 / 1024
		totalMB := float64(m.Sys) / 1024 / 1024
		usagePercent := (usedMB / totalMB) * 100

		status := StatusHealthy
		message := "Memory usage normal"

		if usagePercent > thresholdPercent {
			status = StatusDegraded
			message = "High memory usage"
		}
		if usagePercent > 90 {
			status = StatusUnhealthy
			message = "Critical memory usage"
		}

		return ComponentHealth{
			Status:      status,
			Message:     message,
			LastChecked: time.Now(),
			Metadata: map[string]interface{}{
				"used_mb":        usedMB,
				"total_mb":       totalMB,
				"usage_percent":  usagePercent,
				"goroutines":     runtime.NumGoroutine(),
			},
		}
	}
}

// GoroutineHealthCheck checks goroutine count
func GoroutineHealthCheck(maxGoroutines int) HealthCheckFunc {
	return func() ComponentHealth {
		count := runtime.NumGoroutine()

		status := StatusHealthy
		message := "Goroutine count normal"

		if count > maxGoroutines {
			status = StatusDegraded
			message = "High goroutine count"
		}
		if count > maxGoroutines*2 {
			status = StatusUnhealthy
			message = "Critical goroutine count"
		}

		return ComponentHealth{
			Status:      status,
			Message:     message,
			LastChecked: time.Now(),
			Metadata: map[string]interface{}{
				"count":      count,
				"threshold":  maxGoroutines,
			},
		}
	}
}

// UptimeHealthCheck checks system uptime
func UptimeHealthCheck(startTime time.Time, minUptime time.Duration) HealthCheckFunc {
	return func() ComponentHealth {
		uptime := time.Since(startTime)

		status := StatusHealthy
		if uptime < minUptime {
			status = StatusDegraded
		}

		return ComponentHealth{
			Status:      status,
			Message:     "System running",
			LastChecked: time.Now(),
			Metadata: map[string]interface{}{
				"uptime_seconds": uptime.Seconds(),
				"started_at":     startTime.Format(time.RFC3339),
			},
		}
	}
}

// Global health checker
var globalHealthChecker = NewHealthChecker("v3.0.0")

// GetHealthChecker returns the global health checker
func GetHealthChecker() *HealthChecker {
	return globalHealthChecker
}

// SetGlobalHealthChecker sets the global health checker
func SetGlobalHealthChecker(hc *HealthChecker) {
	globalHealthChecker = hc
}
