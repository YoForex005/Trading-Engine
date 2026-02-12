package admin

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"time"
)

// ============================================
// Performance Monitoring Structs
// ============================================

type PerformanceMetrics struct {
	Timestamp          time.Time `json:"timestamp"`
	CPUUsage           float64   `json:"cpu_usage"`           // percentage
	MemoryUsage        float64   `json:"memory_usage"`        // MB
	MemoryPercent      float64   `json:"memory_percent"`      // percentage
	GoroutineCount     int       `json:"goroutine_count"`     // active goroutines
	GCPauseMs          float64   `json:"gc_pause_ms"`         // GC pause time
	UptimeSeconds      int64     `json:"uptime_seconds"`      // uptime in seconds
	RequestsPerSecond  float64   `json:"requests_per_second"` // RPS
	AvgLatencyMs       float64   `json:"avg_latency_ms"`      // average latency
	ErrorRate          float64   `json:"error_rate"`          // percentage
	ActiveConnections  int       `json:"active_connections"`  // HTTP connections
	WSConnections      int       `json:"ws_connections"`      // WebSocket connections
}

type EndpointMetrics struct {
	Path          string    `json:"path"`
	Method        string    `json:"method"`
	AvgLatencyMs  float64   `json:"avg_latency_ms"`
	P95LatencyMs  float64   `json:"p95_latency_ms"`
	P99LatencyMs  float64   `json:"p99_latency_ms"`
	RequestCount  int64     `json:"request_count"`
	ErrorCount    int64     `json:"error_count"`
	ErrorRate     float64   `json:"error_rate"` // percentage
	LastCalled    time.Time `json:"last_called"`
	ThroughputRPS float64   `json:"throughput_rps"`
}

type ServiceHealth struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"` // healthy, degraded, unhealthy, down
	Uptime       int64     `json:"uptime"` // seconds
	LastCheck    time.Time `json:"last_check"`
	ResponseTime float64   `json:"response_time"` // ms
	ErrorRate    float64   `json:"error_rate"`    // percentage
	Message      string    `json:"message,omitempty"`
}

type ConnectionMetrics struct {
	Timestamp         time.Time `json:"timestamp"`
	ActiveConnections int       `json:"active_connections"`
	WSConnections     int       `json:"ws_connections"`
	TotalConnections  int       `json:"total_connections"`
}

type PerfErrorRatePoint struct {
	Timestamp time.Time `json:"timestamp"`
	ErrorRate float64   `json:"error_rate"`
	Errors    int       `json:"errors"`
	Requests  int       `json:"requests"`
}

// ============================================
// Performance Monitor Service
// ============================================

type PerformanceMonitorService struct {
	mu                sync.RWMutex
	systemHistory     []PerformanceMetrics
	endpoints         map[string]*EndpointMetrics // key: method:path
	services          map[string]*ServiceHealth
	connectionHistory []ConnectionMetrics
	errorRateHistory  []PerfErrorRatePoint
	startTime         time.Time
}

func NewPerformanceMonitorService() *PerformanceMonitorService {
	service := &PerformanceMonitorService{
		systemHistory:     make([]PerformanceMetrics, 0, 60),
		endpoints:         make(map[string]*EndpointMetrics),
		services:          make(map[string]*ServiceHealth),
		connectionHistory: make([]ConnectionMetrics, 0, 60),
		errorRateHistory:  make([]PerfErrorRatePoint, 0, 60),
		startTime:         time.Now().Add(-2 * time.Hour), // mock uptime
	}

	service.generateMockData()
	return service
}

func (s *PerformanceMonitorService) generateMockData() {
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	// Generate 60 data points of system metrics (1 per minute for last hour)
	for i := 59; i >= 0; i-- {
		timestamp := now.Add(-time.Duration(i) * time.Minute)

		// CPU usage: 20-80% with some spikes
		cpuUsage := 30.0 + rand.Float64()*40.0
		if rand.Float64() < 0.1 {
			cpuUsage += 20.0 // 10% chance of spike
		}

		// Memory usage: 500-2000 MB
		memUsage := 800.0 + rand.Float64()*800.0
		memPercent := (memUsage / 4096.0) * 100.0 // assume 4GB total

		// Goroutines: 50-300
		goroutines := 100 + rand.Intn(200)

		// GC pause: 0.1-5ms
		gcPause := 0.5 + rand.Float64()*4.0

		// RPS: 100-1000
		rps := 200.0 + rand.Float64()*700.0

		// Avg latency: 5-50ms
		avgLatency := 10.0 + rand.Float64()*35.0

		// Error rate: 0.1-3%
		errorRate := 0.2 + rand.Float64()*2.5

		// Connections: 50-500
		activeConns := 100 + rand.Intn(400)
		wsConns := 20 + rand.Intn(180)

		uptime := int64(timestamp.Sub(s.startTime).Seconds())

		s.systemHistory = append(s.systemHistory, PerformanceMetrics{
			Timestamp:          timestamp,
			CPUUsage:           cpuUsage,
			MemoryUsage:        memUsage,
			MemoryPercent:      memPercent,
			GoroutineCount:     goroutines,
			GCPauseMs:          gcPause,
			UptimeSeconds:      uptime,
			RequestsPerSecond:  rps,
			AvgLatencyMs:       avgLatency,
			ErrorRate:          errorRate,
			ActiveConnections:  activeConns,
			WSConnections:      wsConns,
		})

		// Connection history
		s.connectionHistory = append(s.connectionHistory, ConnectionMetrics{
			Timestamp:         timestamp,
			ActiveConnections: activeConns,
			WSConnections:     wsConns,
			TotalConnections:  activeConns + wsConns,
		})

		// Error rate history
		totalRequests := int(rps * 60) // requests per minute
		errors := int(float64(totalRequests) * errorRate / 100.0)
		s.errorRateHistory = append(s.errorRateHistory, PerfErrorRatePoint{
			Timestamp: timestamp,
			ErrorRate: errorRate,
			Errors:    errors,
			Requests:  totalRequests,
		})
	}

	// Generate 30 endpoint performance entries
	endpoints := []struct {
		path   string
		method string
	}{
		{"/api/market-data/quotes", "GET"},
		{"/api/market-data/ticks", "GET"},
		{"/api/market-data/candles", "GET"},
		{"/api/orders", "POST"},
		{"/api/orders", "GET"},
		{"/api/positions", "GET"},
		{"/api/positions/close", "POST"},
		{"/admin/users", "GET"},
		{"/admin/users", "POST"},
		{"/admin/symbols", "GET"},
		{"/admin/groups", "GET"},
		{"/admin/analytics", "GET"},
		{"/admin/competitions", "GET"},
		{"/admin/white-label", "GET"},
		{"/admin/promotions", "GET"},
		{"/admin/client-notes", "GET"},
		{"/admin/platform-config", "GET"},
		{"/admin/order-flow/live", "GET"},
		{"/admin/localization/languages", "GET"},
		{"/ws/market-data", "WS"},
		{"/ws/admin", "WS"},
		{"/api/auth/login", "POST"},
		{"/api/auth/logout", "POST"},
		{"/api/account/balance", "GET"},
		{"/api/account/history", "GET"},
		{"/admin/spreads/live", "GET"},
		{"/admin/risk/exposure", "GET"},
		{"/admin/lp/status", "GET"},
		{"/api/symbols/search", "GET"},
		{"/admin/performance/system", "GET"},
	}

	for _, ep := range endpoints {
		key := ep.method + ":" + ep.path

		// Request count: 1000-100000
		reqCount := int64(1000 + rand.Intn(99000))

		// Error count: 0.1-5% of requests
		errorRate := 0.1 + rand.Float64()*4.9
		errorCount := int64(float64(reqCount) * errorRate / 100.0)

		// Avg latency: 5-200ms
		avgLatency := 10.0 + rand.Float64()*180.0

		// P95: 1.5-2x avg
		p95Latency := avgLatency * (1.5 + rand.Float64()*0.5)

		// P99: 2-3x avg
		p99Latency := avgLatency * (2.0 + rand.Float64()*1.0)

		// Throughput: based on request count over 1 hour
		throughput := float64(reqCount) / 3600.0

		// Last called: within last 5 minutes
		lastCalled := now.Add(-time.Duration(rand.Intn(300)) * time.Second)

		s.endpoints[key] = &EndpointMetrics{
			Path:          ep.path,
			Method:        ep.method,
			AvgLatencyMs:  avgLatency,
			P95LatencyMs:  p95Latency,
			P99LatencyMs:  p99Latency,
			RequestCount:  reqCount,
			ErrorCount:    errorCount,
			ErrorRate:     errorRate,
			LastCalled:    lastCalled,
			ThroughputRPS: throughput,
		}
	}

	// Generate 8 service health checks
	services := []struct {
		name   string
		status string
	}{
		{"trading-engine", "healthy"},
		{"market-data", "healthy"},
		{"fix-gateway", "healthy"},
		{"websocket", "healthy"},
		{"admin-api", "healthy"},
		{"database", "degraded"},
		{"risk-engine", "healthy"},
		{"auth", "healthy"},
	}

	for _, svc := range services {
		// Response time: 1-50ms
		responseTime := 5.0 + rand.Float64()*40.0
		if svc.status == "degraded" {
			responseTime = 50.0 + rand.Float64()*100.0
		}

		// Error rate: 0.1-5%
		errorRate := 0.1 + rand.Float64()*4.9
		if svc.status == "degraded" {
			errorRate = 5.0 + rand.Float64()*10.0
		}

		// Uptime: 1-7 days
		uptime := int64(86400 + rand.Intn(518400))

		message := ""
		if svc.status == "degraded" {
			message = "High latency detected"
		}

		s.services[svc.name] = &ServiceHealth{
			Name:         svc.name,
			Status:       svc.status,
			Uptime:       uptime,
			LastCheck:    now.Add(-time.Duration(rand.Intn(60)) * time.Second),
			ResponseTime: responseTime,
			ErrorRate:    errorRate,
			Message:      message,
		}
	}
}

func (s *PerformanceMonitorService) GetCurrentMetrics() *PerformanceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.systemHistory) == 0 {
		return nil
	}

	// Return most recent metrics
	return &s.systemHistory[len(s.systemHistory)-1]
}

func (s *PerformanceMonitorService) GetMetricsHistory() []PerformanceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.systemHistory
}

func (s *PerformanceMonitorService) GetEndpointMetrics() []EndpointMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]EndpointMetrics, 0, len(s.endpoints))
	for _, ep := range s.endpoints {
		result = append(result, *ep)
	}

	// Sort by request count descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].RequestCount > result[j].RequestCount
	})

	return result
}

func (s *PerformanceMonitorService) GetServiceHealth() []ServiceHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]ServiceHealth, 0, len(s.services))
	for _, svc := range s.services {
		result = append(result, *svc)
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func (s *PerformanceMonitorService) GetSlowQueries(limit int) []EndpointMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]EndpointMetrics, 0, len(s.endpoints))
	for _, ep := range s.endpoints {
		result = append(result, *ep)
	}

	// Sort by p99 latency descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].P99LatencyMs > result[j].P99LatencyMs
	})

	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result
}

func (s *PerformanceMonitorService) GetErrorRateTrend() []PerfErrorRatePoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.errorRateHistory
}

func (s *PerformanceMonitorService) GetConnectionMetrics() []ConnectionMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.connectionHistory
}

// ============================================
// Performance Monitor Handler
// ============================================

type PerformanceMonitorHandler struct {
	service     *PerformanceMonitorService
	authService *AuthService
}

func NewPerformanceMonitorHandler(service *PerformanceMonitorService, authService *AuthService) *PerformanceMonitorHandler {
	return &PerformanceMonitorHandler{
		service:     service,
		authService: authService,
	}
}

func (h *PerformanceMonitorHandler) HandleGetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	metrics := h.service.GetCurrentMetrics()
	if metrics == nil {
		http.Error(w, "No metrics available", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

func (h *PerformanceMonitorHandler) HandleGetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	history := h.service.GetMetricsHistory()

	// Calculate summary stats
	var totalCPU, totalMem, totalLatency float64
	var maxCPU, maxMem, maxLatency float64
	var minCPU, minMem, minLatency float64 = math.MaxFloat64, math.MaxFloat64, math.MaxFloat64

	for _, m := range history {
		totalCPU += m.CPUUsage
		totalMem += m.MemoryPercent
		totalLatency += m.AvgLatencyMs

		if m.CPUUsage > maxCPU {
			maxCPU = m.CPUUsage
		}
		if m.CPUUsage < minCPU {
			minCPU = m.CPUUsage
		}

		if m.MemoryPercent > maxMem {
			maxMem = m.MemoryPercent
		}
		if m.MemoryPercent < minMem {
			minMem = m.MemoryPercent
		}

		if m.AvgLatencyMs > maxLatency {
			maxLatency = m.AvgLatencyMs
		}
		if m.AvgLatencyMs < minLatency {
			minLatency = m.AvgLatencyMs
		}
	}

	count := float64(len(history))
	avgCPU := totalCPU / count
	avgMem := totalMem / count
	avgLatency := totalLatency / count

	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"summary": map[string]interface{}{
			"data_points": len(history),
			"cpu": map[string]float64{
				"avg": avgCPU,
				"min": minCPU,
				"max": maxCPU,
			},
			"memory": map[string]float64{
				"avg": avgMem,
				"min": minMem,
				"max": maxMem,
			},
			"latency": map[string]float64{
				"avg": avgLatency,
				"min": minLatency,
				"max": maxLatency,
			},
		},
	})
}

func (h *PerformanceMonitorHandler) HandleGetEndpointMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	endpoints := h.service.GetEndpointMetrics()

	// Calculate totals
	var totalRequests, totalErrors int64
	var totalLatency float64
	for _, ep := range endpoints {
		totalRequests += ep.RequestCount
		totalErrors += ep.ErrorCount
		totalLatency += ep.AvgLatencyMs
	}

	avgLatency := totalLatency / float64(len(endpoints))
	errorRate := float64(totalErrors) / float64(totalRequests) * 100.0

	json.NewEncoder(w).Encode(map[string]interface{}{
		"endpoints": endpoints,
		"summary": map[string]interface{}{
			"total_endpoints":   len(endpoints),
			"total_requests":    totalRequests,
			"total_errors":      totalErrors,
			"overall_error_rate": errorRate,
			"avg_latency_ms":    avgLatency,
		},
	})
}

func (h *PerformanceMonitorHandler) HandleGetServiceHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	services := h.service.GetServiceHealth()

	// Count by status
	statusCounts := make(map[string]int)
	for _, svc := range services {
		statusCounts[svc.Status]++
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
		"summary": map[string]interface{}{
			"total_services": len(services),
			"status_counts":  statusCounts,
		},
	})
}

func (h *PerformanceMonitorHandler) HandleGetSlowQueries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	slowQueries := h.service.GetSlowQueries(20) // top 20

	json.NewEncoder(w).Encode(map[string]interface{}{
		"slow_queries": slowQueries,
		"total":        len(slowQueries),
	})
}

func (h *PerformanceMonitorHandler) HandleGetErrorRate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	errorRates := h.service.GetErrorRateTrend()

	// Calculate average
	var totalErrorRate float64
	var totalErrors, totalRequests int
	for _, point := range errorRates {
		totalErrorRate += point.ErrorRate
		totalErrors += point.Errors
		totalRequests += point.Requests
	}

	avgErrorRate := totalErrorRate / float64(len(errorRates))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error_rate_trend": errorRates,
		"summary": map[string]interface{}{
			"data_points":     len(errorRates),
			"avg_error_rate":  avgErrorRate,
			"total_errors":    totalErrors,
			"total_requests":  totalRequests,
		},
	})
}

func (h *PerformanceMonitorHandler) HandleGetConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	connections := h.service.GetConnectionMetrics()

	// Calculate averages
	var totalHTTP, totalWS int
	for _, conn := range connections {
		totalHTTP += conn.ActiveConnections
		totalWS += conn.WSConnections
	}

	avgHTTP := float64(totalHTTP) / float64(len(connections))
	avgWS := float64(totalWS) / float64(len(connections))

	// Get current (last entry)
	var currentHTTP, currentWS int
	if len(connections) > 0 {
		last := connections[len(connections)-1]
		currentHTTP = last.ActiveConnections
		currentWS = last.WSConnections
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"connection_history": connections,
		"summary": map[string]interface{}{
			"data_points":              len(connections),
			"current_http_connections": currentHTTP,
			"current_ws_connections":   currentWS,
			"avg_http_connections":     avgHTTP,
			"avg_ws_connections":       avgWS,
		},
	})
}
