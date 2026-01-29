package monitoring

import (
	"runtime"
	"strconv"
	"time"
)

// RuntimeMetricsCollector collects runtime metrics periodically
type RuntimeMetricsCollector struct {
	interval time.Duration
	stopChan chan struct{}
}

// NewRuntimeMetricsCollector creates a new runtime metrics collector
func NewRuntimeMetricsCollector(interval time.Duration) *RuntimeMetricsCollector {
	return &RuntimeMetricsCollector{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start starts collecting runtime metrics
func (rmc *RuntimeMetricsCollector) Start() {
	ticker := time.NewTicker(rmc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rmc.collectMetrics()
		case <-rmc.stopChan:
			return
		}
	}
}

// Stop stops the runtime metrics collector
func (rmc *RuntimeMetricsCollector) Stop() {
	close(rmc.stopChan)
}

// collectMetrics collects and records runtime metrics
func (rmc *RuntimeMetricsCollector) collectMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	SetMemoryUsage(m.Alloc)

	// Goroutine count
	SetGoroutineCount(runtime.NumGoroutine())

	// Log if memory usage is high
	usedMB := float64(m.Alloc) / 1024 / 1024
	totalMB := float64(m.Sys) / 1024 / 1024
	usagePercent := (usedMB / totalMB) * 100

	if usagePercent > 80 {
		logger := GetLogger()
		logger.Warn("High memory usage detected", map[string]interface{}{
			"used_mb":       usedMB,
			"total_mb":      totalMB,
			"usage_percent": usagePercent,
			"goroutines":    runtime.NumGoroutine(),
		})

		// Fire alert
		alertManager := GetAlertManager()
		alertManager.FireAlert(&Alert{
			Name:      "HighMemoryUsage",
			Severity:  SeverityWarning,
			Message:   "Memory usage exceeded 80%",
			Timestamp: time.Now(),
			Labels: map[string]string{
				"component": "runtime",
				"severity":  "warning",
			},
			Annotations: map[string]string{
				"used_mb":       formatFloat(usedMB),
				"total_mb":      formatFloat(totalMB),
				"usage_percent": formatFloat(usagePercent),
			},
		})
	}

	// Check goroutine count
	goroutineCount := runtime.NumGoroutine()
	if goroutineCount > 10000 {
		logger := GetLogger()
		logger.Warn("High goroutine count detected", map[string]interface{}{
			"count": goroutineCount,
		})

		alertManager := GetAlertManager()
		alertManager.FireAlert(&Alert{
			Name:      "HighGoroutineCount",
			Severity:  SeverityWarning,
			Message:   "Goroutine count exceeded 10000",
			Timestamp: time.Now(),
			Labels: map[string]string{
				"component": "runtime",
				"severity":  "warning",
			},
			Annotations: map[string]string{
				"count": formatInt(goroutineCount),
			},
		})
	}
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatInt(i int) string {
	return strconv.Itoa(i)
}
