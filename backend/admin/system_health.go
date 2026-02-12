package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// Severity levels for system alerts
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "CRITICAL"
	SeverityWarning  AlertSeverity = "WARNING"
	SeverityInfo     AlertSeverity = "INFO"
)

// SystemServiceHealth status enum
type SystemServiceHealth string

const (
	HealthOnline   SystemServiceHealth = "online"
	HealthDegraded SystemServiceHealth = "degraded"
	HealthOffline  SystemServiceHealth = "offline"
)

// SystemMetrics tracks overall system performance
type SystemMetrics struct {
	CPUUsage         float64   `json:"cpu_usage"`          // Simulated CPU usage percentage
	MemoryUsage      uint64    `json:"memory_usage"`       // Memory usage in bytes
	MemoryPercent    float64   `json:"memory_percent"`     // Memory usage percentage
	GoroutineCount   int       `json:"goroutine_count"`    // Current number of goroutines
	ActiveWSConns    int       `json:"active_ws_conns"`    // Active WebSocket connections
	RequestsPerMin   int       `json:"requests_per_min"`   // Requests in last minute
	ErrorCount       int       `json:"error_count"`        // Total error count
	Uptime           string    `json:"uptime"`             // System uptime duration
	UptimeSeconds    int64     `json:"uptime_seconds"`     // Uptime in seconds
	Timestamp        time.Time `json:"timestamp"`          // Metrics collection timestamp
	TotalAllocMB     float64   `json:"total_alloc_mb"`     // Total allocated memory in MB
	HeapAllocMB      float64   `json:"heap_alloc_mb"`      // Heap allocated memory in MB
	NumGC            uint32    `json:"num_gc"`             // Number of GC runs
}

// ServiceStatus represents health status of a subsystem
type ServiceStatus struct {
	Name           string        `json:"name"`
	Status         SystemServiceHealth `json:"status"`
	Latency        int           `json:"latency_ms"`    // Latency in milliseconds
	UptimePercent  float64       `json:"uptime_percent"`
	LastCheck      time.Time     `json:"last_check"`
	ErrorRate      float64       `json:"error_rate"`    // Errors per minute
	Message        string        `json:"message,omitempty"`
}

// SystemAlert represents a system health alert
type SystemAlert struct {
	ID        string        `json:"id"`
	Timestamp time.Time     `json:"timestamp"`
	Severity  AlertSeverity `json:"severity"`
	Service   string        `json:"service"`
	Message   string        `json:"message"`
	Value     float64       `json:"value,omitempty"`     // Metric value that triggered alert
	Threshold float64       `json:"threshold,omitempty"` // Threshold that was exceeded
}

// HealthSummary provides overall system health
type HealthSummary struct {
	Status            SystemServiceHealth   `json:"status"`
	TotalServices     int             `json:"total_services"`
	OnlineServices    int             `json:"online_services"`
	DegradedServices  int             `json:"degraded_services"`
	OfflineServices   int             `json:"offline_services"`
	CriticalAlerts    int             `json:"critical_alerts"`
	WarningAlerts     int             `json:"warning_alerts"`
	Uptime            string          `json:"uptime"`
	UptimeSeconds     int64           `json:"uptime_seconds"`
	LastUpdated       time.Time       `json:"last_updated"`
}

// RequestCounter tracks requests per minute
type RequestCounter struct {
	mu       sync.RWMutex
	requests []time.Time // Timestamps of requests in last minute
}

func NewRequestCounter() *RequestCounter {
	return &RequestCounter{
		requests: make([]time.Time, 0),
	}
}

func (rc *RequestCounter) Increment() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	rc.requests = append(rc.requests, now)

	// Clean up old requests (older than 1 minute)
	cutoff := now.Add(-1 * time.Minute)
	validRequests := make([]time.Time, 0)
	for _, t := range rc.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	rc.requests = validRequests
}

func (rc *RequestCounter) GetRPM() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)
	count := 0
	for _, t := range rc.requests {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// SystemHealthService manages system health monitoring
type SystemHealthService struct {
	mu                sync.RWMutex
	startTime         time.Time
	currentMetrics    *SystemMetrics
	serviceStatuses   map[string]*ServiceStatus
	alerts            []*SystemAlert // Ring buffer, max 100
	maxAlerts         int
	requestCounter    *RequestCounter
	errorCount        int
	activeWSConns     int
	simulatedCPU      float64 // Simulated CPU usage
	stopChan          chan struct{}
}

func NewSystemHealthService() *SystemHealthService {
	service := &SystemHealthService{
		startTime:       time.Now(),
		currentMetrics:  &SystemMetrics{},
		serviceStatuses: make(map[string]*ServiceStatus),
		alerts:          make([]*SystemAlert, 0, 100),
		maxAlerts:       100,
		requestCounter:  NewRequestCounter(),
		errorCount:      0,
		activeWSConns:   0,
		simulatedCPU:    15.0, // Start with low CPU usage
		stopChan:        make(chan struct{}),
	}

	// Initialize service statuses
	service.initializeServices()

	// Start background metrics collector
	go service.metricsCollector()

	// Start service health checker
	go service.serviceHealthChecker()

	log.Println("[SystemHealth] System health monitoring initialized (6 services tracked, 10s metric collection)")

	return service
}

func (s *SystemHealthService) initializeServices() {
	services := []string{
		"Backend API",
		"WebSocket Server",
		"FIX Gateway",
		"Database",
		"Redis Cache",
		"LP Connection",
	}

	for _, name := range services {
		s.serviceStatuses[name] = &ServiceStatus{
			Name:          name,
			Status:        HealthOnline,
			Latency:       randomInt(5, 50),
			UptimePercent: 99.9,
			LastCheck:     time.Now(),
			ErrorRate:     0.0,
		}
	}
}

// metricsCollector runs every 10 seconds to collect system metrics
func (s *SystemHealthService) metricsCollector() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.collectMetrics()
		case <-s.stopChan:
			return
		}
	}
}

// serviceHealthChecker simulates health checks and updates service statuses
func (s *SystemHealthService) serviceHealthChecker() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkServiceHealth()
		case <-s.stopChan:
			return
		}
	}
}

func (s *SystemHealthService) collectMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Simulate CPU usage (random walk)
	s.simulatedCPU += randomFloat(-5.0, 5.0)
	if s.simulatedCPU < 5.0 {
		s.simulatedCPU = 5.0
	}
	if s.simulatedCPU > 95.0 {
		s.simulatedCPU = 95.0
	}

	// Calculate memory percentage (assume 16GB total)
	totalMemory := uint64(16 * 1024 * 1024 * 1024) // 16GB
	memPercent := float64(m.Alloc) / float64(totalMemory) * 100

	uptime := time.Since(s.startTime)

	s.currentMetrics = &SystemMetrics{
		CPUUsage:       math.Round(s.simulatedCPU*100) / 100,
		MemoryUsage:    m.Alloc,
		MemoryPercent:  math.Round(memPercent*100) / 100,
		GoroutineCount: runtime.NumGoroutine(),
		ActiveWSConns:  s.activeWSConns,
		RequestsPerMin: s.requestCounter.GetRPM(),
		ErrorCount:     s.errorCount,
		Uptime:         formatSystemHealthDuration(uptime),
		UptimeSeconds:  int64(uptime.Seconds()),
		Timestamp:      time.Now(),
		TotalAllocMB:   math.Round(float64(m.TotalAlloc)/1024/1024*100) / 100,
		HeapAllocMB:    math.Round(float64(m.HeapAlloc)/1024/1024*100) / 100,
		NumGC:          m.NumGC,
	}

	// Check for threshold breaches and generate alerts
	s.checkThresholds()
}

func (s *SystemHealthService) checkServiceHealth() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate service health checks
	for name, status := range s.serviceStatuses {
		// Random latency variation
		status.Latency = randomInt(5, 100)
		status.LastCheck = time.Now()

		// Simulate occasional degradation (5% chance)
		if randomInt(1, 100) <= 5 {
			if status.Status == HealthOnline {
				status.Status = HealthDegraded
				status.Message = "High latency detected"
				status.ErrorRate = randomFloat(0.5, 2.0)

				// Generate alert
				s.addAlertLocked(&SystemAlert{
					ID:        generateID(),
					Timestamp: time.Now(),
					Severity:  SeverityWarning,
					Service:   name,
					Message:   fmt.Sprintf("%s is experiencing degraded performance", name),
				})
			}
		} else {
			// 95% chance to recover if degraded
			if status.Status == HealthDegraded {
				status.Status = HealthOnline
				status.Message = ""
				status.ErrorRate = 0.0
			}
		}

		// Update uptime percentage (simulate 99-99.99%)
		status.UptimePercent = 99.0 + randomFloat(0.0, 0.99)
	}
}

func (s *SystemHealthService) checkThresholds() {
	// Check CPU threshold (>90%)
	if s.currentMetrics.CPUUsage > 90.0 {
		s.addAlertLocked(&SystemAlert{
			ID:        generateID(),
			Timestamp: time.Now(),
			Severity:  SeverityCritical,
			Service:   "Backend API",
			Message:   "CPU usage exceeds critical threshold",
			Value:     s.currentMetrics.CPUUsage,
			Threshold: 90.0,
		})
	}

	// Check memory threshold (>85%)
	if s.currentMetrics.MemoryPercent > 85.0 {
		s.addAlertLocked(&SystemAlert{
			ID:        generateID(),
			Timestamp: time.Now(),
			Severity:  SeverityCritical,
			Service:   "Backend API",
			Message:   "Memory usage exceeds critical threshold",
			Value:     s.currentMetrics.MemoryPercent,
			Threshold: 85.0,
		})
	}

	// Check error rate (>1 error per minute)
	rpm := s.requestCounter.GetRPM()
	if rpm > 0 {
		errorRate := float64(s.errorCount) / float64(rpm)
		if errorRate > 1.0 {
			s.addAlertLocked(&SystemAlert{
				ID:        generateID(),
				Timestamp: time.Now(),
				Severity:  SeverityCritical,
				Service:   "Backend API",
				Message:   "Error rate exceeds threshold",
				Value:     errorRate * 100, // Convert to percentage
				Threshold: 100.0,
			})
		}
	}
}

// addAlertLocked adds an alert to the ring buffer (must be called with lock held)
func (s *SystemHealthService) addAlertLocked(alert *SystemAlert) {
	// Ring buffer logic
	if len(s.alerts) >= s.maxAlerts {
		// Remove oldest alert
		s.alerts = s.alerts[1:]
	}
	s.alerts = append(s.alerts, alert)

	log.Printf("[SystemHealth] %s alert: %s - %s", alert.Severity, alert.Service, alert.Message)
}

// GetCurrentMetrics returns current system metrics
func (s *SystemHealthService) GetCurrentMetrics() *SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	metricsCopy := *s.currentMetrics
	return &metricsCopy
}

// GetServiceStatuses returns all service statuses
func (s *SystemHealthService) GetServiceStatuses() []*ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]*ServiceStatus, 0, len(s.serviceStatuses))
	for _, status := range s.serviceStatuses {
		// Return copies
		statusCopy := *status
		statuses = append(statuses, &statusCopy)
	}
	return statuses
}

// GetAlerts returns recent alerts, optionally filtered by severity
func (s *SystemHealthService) GetAlerts(severity string) []*SystemAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if severity == "" {
		// Return all alerts (make a copy)
		alerts := make([]*SystemAlert, len(s.alerts))
		for i, alert := range s.alerts {
			alertCopy := *alert
			alerts[i] = &alertCopy
		}
		return alerts
	}

	// Filter by severity
	filtered := make([]*SystemAlert, 0)
	for _, alert := range s.alerts {
		if string(alert.Severity) == severity {
			alertCopy := *alert
			filtered = append(filtered, &alertCopy)
		}
	}
	return filtered
}

// GetHealthSummary returns overall system health summary
func (s *SystemHealthService) GetHealthSummary() *HealthSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summary := &HealthSummary{
		Status:         HealthOnline,
		TotalServices:  len(s.serviceStatuses),
		LastUpdated:    time.Now(),
		Uptime:         formatSystemHealthDuration(time.Since(s.startTime)),
		UptimeSeconds:  int64(time.Since(s.startTime).Seconds()),
	}

	// Count service statuses
	for _, status := range s.serviceStatuses {
		switch status.Status {
		case HealthOnline:
			summary.OnlineServices++
		case HealthDegraded:
			summary.DegradedServices++
		case HealthOffline:
			summary.OfflineServices++
		}
	}

	// Count alerts by severity
	for _, alert := range s.alerts {
		switch alert.Severity {
		case SeverityCritical:
			summary.CriticalAlerts++
		case SeverityWarning:
			summary.WarningAlerts++
		}
	}

	// Determine overall status
	if summary.OfflineServices > 0 || summary.CriticalAlerts > 0 {
		summary.Status = HealthOffline
	} else if summary.DegradedServices > 0 || summary.WarningAlerts > 0 {
		summary.Status = HealthDegraded
	}

	return summary
}

// IncrementRequest increments request counter
func (s *SystemHealthService) IncrementRequest() {
	s.requestCounter.Increment()
}

// IncrementError increments error counter
func (s *SystemHealthService) IncrementError() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorCount++
}

// SetActiveWSConns sets the active WebSocket connection count
func (s *SystemHealthService) SetActiveWSConns(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeWSConns = count
}

// Stop stops background goroutines
func (s *SystemHealthService) Stop() {
	close(s.stopChan)
}

// SystemHealthHandler handles HTTP requests for system health
type SystemHealthHandler struct {
	service     *SystemHealthService
	authService *auth.Service
}

func NewSystemHealthHandler(service *SystemHealthService, authService *auth.Service) *SystemHealthHandler {
	return &SystemHealthHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetHealth returns overall health summary
// GET /admin/system/health
func (h *SystemHealthHandler) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	summary := h.service.GetHealthSummary()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// HandleGetMetrics returns current system metrics
// GET /admin/system/metrics
func (h *SystemHealthHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	metrics := h.service.GetCurrentMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// HandleGetServices returns all service statuses
// GET /admin/system/services
func (h *SystemHealthHandler) HandleGetServices(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	statuses := h.service.GetServiceStatuses()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

// HandleGetAlerts returns recent alerts with optional severity filter
// GET /admin/system/alerts?severity=CRITICAL
func (h *SystemHealthHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get severity filter from query params
	severity := r.URL.Query().Get("severity")

	alerts := h.service.GetAlerts(severity)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// Helper functions

func formatDuration(d time.Duration) string {
	return formatSystemHealthDuration(d)
}

func formatSystemHealthDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func randomInt(min, max int) int {
	// Simple pseudo-random using time
	t := time.Now().UnixNano()
	return min + int(t%(int64(max-min+1)))
}

func randomFloat(min, max float64) float64 {
	// Simple pseudo-random using time
	t := time.Now().UnixNano()
	normalized := float64(t%1000) / 1000.0
	return min + (max-min)*normalized
}

var idCounter int

func generateID() string {
	idCounter++
	return fmt.Sprintf("ALERT-%d", idCounter)
}
