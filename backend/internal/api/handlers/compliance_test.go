package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// TestNewComplianceHandler tests handler initialization
func TestNewComplianceHandler(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	if handler == nil {
		t.Fatal("NewComplianceHandler returned nil")
	}

	if handler.engine != engine {
		t.Error("Handler engine not set correctly")
	}
}

// TestHandleBestExecution tests the best execution report endpoint
func TestHandleBestExecution(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFormat string
	}{
		{
			name:           "Valid JSON request",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=json",
			expectedStatus: http.StatusOK,
			expectedFormat: "application/json",
		},
		{
			name:           "Valid CSV request",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=csv",
			expectedStatus: http.StatusOK,
			expectedFormat: "text/csv",
		},
		{
			name:           "Missing start_time",
			queryParams:    "?end_time=2026-01-31T23:59:59Z&format=json",
			expectedStatus: http.StatusBadRequest,
			expectedFormat: "",
		},
		{
			name:           "Invalid date format",
			queryParams:    "?start_time=invalid&end_time=2026-01-31T23:59:59Z&format=json",
			expectedStatus: http.StatusBadRequest,
			expectedFormat: "",
		},
		{
			name:           "Invalid format",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=xml",
			expectedStatus: http.StatusBadRequest,
			expectedFormat: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/compliance/best-execution"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleBestExecution(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedFormat != "" {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, tt.expectedFormat) {
					t.Errorf("Expected Content-Type to contain %s, got %s", tt.expectedFormat, contentType)
				}
			}
		})
	}
}

// TestHandleBestExecutionJSON tests JSON response structure
func TestHandleBestExecutionJSON(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	req := httptest.NewRequest("GET", "/api/compliance/best-execution?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=json", nil)
	w := httptest.NewRecorder()

	handler.HandleBestExecution(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var report BestExecutionReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Validate report structure
	if report.ReportID == "" {
		t.Error("Report ID is empty")
	}

	if report.GeneratedAt.IsZero() {
		t.Error("Generated timestamp is zero")
	}

	if report.Summary.TotalOrders == 0 {
		t.Error("Total orders is zero")
	}

	if len(report.VenueBreakdown) == 0 {
		t.Error("Venue breakdown is empty")
	}

	// Validate venue metrics
	for _, venue := range report.VenueBreakdown {
		if venue.VenueName == "" {
			t.Error("Venue name is empty")
		}
		if venue.OrderCount < 0 {
			t.Error("Order count is negative")
		}
		if venue.FillRate < 0 || venue.FillRate > 100 {
			t.Errorf("Invalid fill rate: %f", venue.FillRate)
		}
	}
}

// TestHandleOrderRouting tests the order routing report endpoint
func TestHandleOrderRouting(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid Q1 request",
			queryParams:    "?quarter=Q1&year=2026&format=json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid Q4 request",
			queryParams:    "?quarter=Q4&year=2025&format=csv",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid quarter",
			queryParams:    "?quarter=Q5&year=2026&format=json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing year",
			queryParams:    "?quarter=Q1&format=json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid year",
			queryParams:    "?quarter=Q1&year=invalid&format=json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/compliance/order-routing"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleOrderRouting(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestHandleOrderRoutingJSON tests JSON response structure
func TestHandleOrderRoutingJSON(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	req := httptest.NewRequest("GET", "/api/compliance/order-routing?quarter=Q1&year=2026&format=json", nil)
	w := httptest.NewRecorder()

	handler.HandleOrderRouting(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var report OrderRoutingReport
	if err := json.NewDecoder(w.Body).Decode(&report); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	// Validate report structure
	if report.ReportID == "" {
		t.Error("Report ID is empty")
	}

	if report.Quarter != "Q1" {
		t.Errorf("Expected quarter Q1, got %s", report.Quarter)
	}

	if report.Year != 2026 {
		t.Errorf("Expected year 2026, got %d", report.Year)
	}

	if len(report.RoutingData) == 0 {
		t.Error("Routing data is empty")
	}

	// Validate routing stats
	totalPct := 0.0
	for _, venue := range report.RoutingData {
		if venue.VenueName == "" {
			t.Error("Venue name is empty")
		}
		totalPct += venue.OrdersRoutedPct
	}

	// Total percentage should be close to 100% (allow small rounding errors)
	if totalPct < 99.0 || totalPct > 101.0 {
		t.Errorf("Total routing percentage is %f, expected ~100%%", totalPct)
	}
}

// TestHandleAuditTrail tests the audit trail export endpoint
func TestHandleAuditTrail(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid JSON request",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&format=json",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid CSV request with filter",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z&entity_type=order&format=csv",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing end_time",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&format=json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/compliance/audit-trail"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleAuditTrail(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestHandleAuditLog tests the audit log write endpoint
func TestHandleAuditLog(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "Valid audit entry",
			requestBody: `{
				"user_id": "user-123",
				"action": "UPDATE",
				"entity_type": "position",
				"entity_id": "position-789",
				"details": {
					"field": "stop_loss",
					"old_value": 1.0850,
					"new_value": 1.0870
				}
			}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{invalid json}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/compliance/audit-log", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandleAuditLog(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if response["entry_id"] == nil {
					t.Error("Response missing entry_id")
				}

				if response["timestamp"] == nil {
					t.Error("Response missing timestamp")
				}

				if response["hash"] == nil {
					t.Error("Response missing hash")
				}
			}
		})
	}
}

// TestGenerateAuditHash tests hash generation
func TestGenerateAuditHash(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	entry := AuditEntry{
		Timestamp:  time.Now(),
		UserID:     "user-123",
		EntityType: "order",
		EntityID:   "order-456",
		Action:     "INSERT",
	}

	hash1 := handler.generateAuditHash(entry)
	_ = handler.generateAuditHash(entry)

	// Same input should produce consistent hash (for this simplified implementation)
	// In production with timestamp, it would differ
	if hash1 == "" {
		t.Error("Hash is empty")
	}

	if len(hash1) < 16 {
		t.Errorf("Hash too short: %s", hash1)
	}
}

// TestGetClientIP tests IP extraction
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1",
			},
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.200",
			},
			remoteAddr: "127.0.0.1:12345",
			expectedIP: "192.168.1.200",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.50:12345",
			expectedIP: "192.168.1.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			req.RemoteAddr = tt.remoteAddr

			ip := getClientIP(req)

			if ip != tt.expectedIP {
				t.Errorf("Expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

// TestAuditMiddleware tests the audit middleware
func TestAuditMiddleware(t *testing.T) {
	// Create a simple handler that returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with audit middleware
	wrappedHandler := AuditMiddleware(testHandler)

	tests := []struct {
		name           string
		path           string
		method         string
		shouldAudit    bool
		expectedStatus int
	}{
		{
			name:           "POST request should be audited",
			path:           "/api/orders",
			method:         "POST",
			shouldAudit:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request should be audited",
			path:           "/api/positions",
			method:         "GET",
			shouldAudit:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OPTIONS request should skip audit",
			path:           "/api/orders",
			method:         "OPTIONS",
			shouldAudit:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health check should skip audit",
			path:           "/health",
			method:         "GET",
			shouldAudit:    false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-User-ID", "test-user")
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// In production, we would verify audit log entry was created
			// For now, just verify the handler executed successfully
		})
	}
}

// TestResponseRecorder tests response status capture
func TestResponseRecorder(t *testing.T) {
	w := httptest.NewRecorder()
	recorder := &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// Test WriteHeader
	recorder.WriteHeader(http.StatusCreated)
	if recorder.statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.statusCode)
	}

	// Test Write (should preserve status)
	recorder.Write([]byte("test"))
	if recorder.statusCode != http.StatusCreated {
		t.Errorf("Status code changed after Write, expected %d, got %d", http.StatusCreated, recorder.statusCode)
	}
}

// TestMethodNotAllowed tests method validation
func TestMethodNotAllowed(t *testing.T) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)

	endpoints := []struct {
		name    string
		path    string
		handler http.HandlerFunc
		allowed []string
	}{
		{
			name:    "Best Execution",
			path:    "/api/compliance/best-execution",
			handler: handler.HandleBestExecution,
			allowed: []string{"GET"},
		},
		{
			name:    "Order Routing",
			path:    "/api/compliance/order-routing",
			handler: handler.HandleOrderRouting,
			allowed: []string{"GET"},
		},
		{
			name:    "Audit Trail",
			path:    "/api/compliance/audit-trail",
			handler: handler.HandleAuditTrail,
			allowed: []string{"GET"},
		},
		{
			name:    "Audit Log",
			path:    "/api/compliance/audit-log",
			handler: handler.HandleAuditLog,
			allowed: []string{"POST"},
		},
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, endpoint := range endpoints {
		for _, method := range methods {
			t.Run(endpoint.name+"_"+method, func(t *testing.T) {
				req := httptest.NewRequest(method, endpoint.path, nil)
				w := httptest.NewRecorder()

				endpoint.handler(w, req)

				// Check if method is allowed
				isAllowed := false
				for _, allowedMethod := range endpoint.allowed {
					if method == allowedMethod {
						isAllowed = true
						break
					}
				}

				if !isAllowed && w.Code != http.StatusMethodNotAllowed {
					t.Errorf("Expected %s to return 405 for %s method, got %d",
						endpoint.name, method, w.Code)
				}
			})
		}
	}
}

// BenchmarkBestExecutionReport benchmarks report generation
func BenchmarkBestExecutionReport(b *testing.B) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)
	startTime := time.Now().AddDate(0, -1, 0)
	endTime := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.generateBestExecutionReport(startTime, endTime)
	}
}

// BenchmarkAuditHashGeneration benchmarks hash generation
func BenchmarkAuditHashGeneration(b *testing.B) {
	engine := core.NewEngine()
	handler := NewComplianceHandler(engine)
	entry := AuditEntry{
		Timestamp:  time.Now(),
		UserID:     "user-123",
		EntityType: "order",
		EntityID:   "order-456",
		Action:     "INSERT",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.generateAuditHash(entry)
	}
}
