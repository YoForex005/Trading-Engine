package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleLPComparison(t *testing.T) {
	// Skip if database not available
	handler, err := NewAnalyticsLPHandler()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer handler.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Default query (last 24h, latency metric)",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "With time range",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-19T00:00:00Z",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "With symbol filter",
			queryParams:    "?symbol=EURUSD",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Fill rate metric",
			queryParams:    "?metric=fill_rate",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Slippage metric",
			queryParams:    "?metric=slippage",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Invalid metric",
			queryParams:    "?metric=invalid",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/analytics/lp/comparison"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleLPComparison(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusOK {
				var response LPComparisonResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Check response structure
				if response.LPs == nil {
					t.Error("Expected lps array in response")
				}

				// If there are LPs, validate structure
				for _, lp := range response.LPs {
					if lp.Name == "" {
						t.Error("LP name should not be empty")
					}
					if lp.Rank < 1 {
						t.Error("LP rank should be >= 1")
					}
				}
			}
		})
	}
}

func TestHandleLPPerformance(t *testing.T) {
	handler, err := NewAnalyticsLPHandler()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer handler.Close()

	tests := []struct {
		name           string
		lpName         string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid LP name",
			lpName:         "OANDA",
			queryParams:    "",
			expectedStatus: http.StatusOK, // or 404 if no data
		},
		{
			name:           "With time range",
			lpName:         "OANDA",
			queryParams:    "?start_time=2026-01-01T00:00:00Z&end_time=2026-01-19T00:00:00Z",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "With symbol filter",
			lpName:         "OANDA",
			queryParams:    "?symbol=EURUSD",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty LP name",
			lpName:         "",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/analytics/lp/performance/"+tt.lpName+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleLPPerformance(w, req)

			// Allow both OK and NotFound as valid responses (depends on data)
			if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d or 404, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusOK {
				var response LPPerformanceResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Validate response structure
				if response.LPName == "" {
					t.Error("LP name should not be empty in response")
				}

				if response.Metrics.LatencyP50 < 0 {
					t.Error("Latency P50 should be >= 0")
				}

				if response.Timeline == nil {
					t.Error("Timeline should be present (can be empty array)")
				}
			}
		})
	}
}

func TestHandleLPRanking(t *testing.T) {
	handler, err := NewAnalyticsLPHandler()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer handler.Close()

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Default query",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Latency metric",
			queryParams:    "?metric=latency&limit=5",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Fill rate metric",
			queryParams:    "?metric=fill_rate&limit=10",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Volume metric",
			queryParams:    "?metric=volume&limit=3",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "Invalid metric",
			queryParams:    "?metric=invalid",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/analytics/lp/ranking"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleLPRanking(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusOK {
				var response LPRankingResponse
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Validate rankings
				if response.Rankings == nil {
					t.Error("Rankings should be present")
				}

				for i, ranking := range response.Rankings {
					if ranking.Rank != i+1 {
						t.Errorf("Expected rank %d, got %d", i+1, ranking.Rank)
					}

					if ranking.Percentile < 0 || ranking.Percentile > 100 {
						t.Errorf("Percentile should be between 0 and 100, got %f", ranking.Percentile)
					}
				}
			}
		})
	}
}

func TestBuildHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		metric   string
		expected string
	}{
		{"latency rank", "latency", "ROW_NUMBER() OVER (ORDER BY avg_latency_ms ASC NULLS LAST)"},
		{"fill_rate rank", "fill_rate", "ROW_NUMBER() OVER (ORDER BY fill_rate_pct DESC NULLS LAST)"},
		{"slippage rank", "slippage", "ROW_NUMBER() OVER (ORDER BY avg_slippage_bps ASC NULLS LAST)"},
		{"volume rank", "volume", "ROW_NUMBER() OVER (ORDER BY volume_24h DESC NULLS LAST)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildRankExpression(tt.metric)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSymbolFilter(t *testing.T) {
	tests := []struct {
		symbol   string
		expected string
	}{
		{"", ""},
		{"EURUSD", " AND o.symbol = 'EURUSD'"},
		{"BTCUSD", " AND o.symbol = 'BTCUSD'"},
	}

	for _, tt := range tests {
		t.Run("symbol_"+tt.symbol, func(t *testing.T) {
			result := buildSymbolFilter(tt.symbol)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	setCORSHeaders(w)

	headers := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("Expected CORS header %s to be set", header)
		}
	}
}

func TestOPTIONSRequest(t *testing.T) {
	handler, err := NewAnalyticsLPHandler()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer handler.Close()

	endpoints := []string{
		"/api/analytics/lp/comparison",
		"/api/analytics/lp/ranking",
	}

	for _, endpoint := range endpoints {
		t.Run("OPTIONS_"+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", endpoint, nil)
			w := httptest.NewRecorder()

			switch endpoint {
			case "/api/analytics/lp/comparison":
				handler.HandleLPComparison(w, req)
			case "/api/analytics/lp/ranking":
				handler.HandleLPRanking(w, req)
			}

			if w.Code != http.StatusOK {
				t.Errorf("OPTIONS request should return 200, got %d", w.Code)
			}
		})
	}
}

func TestInvalidMethods(t *testing.T) {
	handler, err := NewAnalyticsLPHandler()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer handler.Close()

	tests := []struct {
		method   string
		endpoint string
		handler  http.HandlerFunc
	}{
		{"POST", "/api/analytics/lp/comparison", handler.HandleLPComparison},
		{"PUT", "/api/analytics/lp/comparison", handler.HandleLPComparison},
		{"DELETE", "/api/analytics/lp/comparison", handler.HandleLPComparison},
		{"POST", "/api/analytics/lp/ranking", handler.HandleLPRanking},
	}

	for _, tt := range tests {
		t.Run(tt.method+"_"+tt.endpoint, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected 405 for %s request, got %d", tt.method, w.Code)
			}
		})
	}
}

func TestTimeRangeParsing(t *testing.T) {
	handler, err := NewAnalyticsLPHandler()
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
	}
	defer handler.Close()

	// Test valid RFC3339 time
	validTime := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	req := httptest.NewRequest("GET", "/api/analytics/lp/comparison?start_time="+validTime, nil)
	w := httptest.NewRecorder()

	handler.HandleLPComparison(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Valid time should be accepted, got status %d", w.Code)
	}

	// Test invalid time format (should use default)
	req = httptest.NewRequest("GET", "/api/analytics/lp/comparison?start_time=invalid", nil)
	w = httptest.NewRecorder()

	handler.HandleLPComparison(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Invalid time should default to 24h, got status %d", w.Code)
	}
}
