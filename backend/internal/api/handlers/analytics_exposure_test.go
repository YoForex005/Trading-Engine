package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

func TestHandleCurrentExposure(t *testing.T) {
	// Create engine and handler
	engine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(engine)
	handler := NewAPIHandler(engine, pnlEngine)

	// Create test account
	account := engine.CreateAccount("test-user", "Test User", "password", true)
	engine.GetLedger().SetBalance(account.ID, 100000)
	account.Balance = 100000

	// Set up price callback
	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Execute some test orders to create positions
	_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 1.0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to execute order: %v", err)
	}

	_, err = engine.ExecuteMarketOrder(account.ID, "EURUSD", "SELL", 0.5, 0, 0)
	if err != nil {
		t.Fatalf("Failed to execute order: %v", err)
	}

	_, err = engine.ExecuteMarketOrder(account.ID, "GBPUSD", "BUY", 2.0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to execute order: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/analytics/exposure/current", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleCurrentExposure(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if _, ok := response["symbols"]; !ok {
		t.Error("Response missing 'symbols' field")
	}

	if _, ok := response["timestamp"]; !ok {
		t.Error("Response missing 'timestamp' field")
	}

	symbols, ok := response["symbols"].([]interface{})
	if !ok || len(symbols) == 0 {
		t.Error("Expected non-empty symbols array")
	}

	// Verify we have exposure data for both symbols
	symbolsFound := make(map[string]bool)
	for _, sym := range symbols {
		symData, ok := sym.(map[string]interface{})
		if !ok {
			continue
		}
		symbol, ok := symData["symbol"].(string)
		if ok {
			symbolsFound[symbol] = true
		}
	}

	if !symbolsFound["EURUSD"] {
		t.Error("Expected EURUSD in exposure data")
	}
	if !symbolsFound["GBPUSD"] {
		t.Error("Expected GBPUSD in exposure data")
	}
}

func TestHandleExposureHeatmap(t *testing.T) {
	// Create engine and handler
	engine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(engine)
	handler := NewAPIHandler(engine, pnlEngine)

	// Create test account
	account := engine.CreateAccount("test-user", "Test User", "password", true)
	engine.GetLedger().SetBalance(account.ID, 100000)
	account.Balance = 100000

	// Set up price callback
	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Execute some test orders
	_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 1.0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to execute order: %v", err)
	}

	// Create request with time range
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	req := httptest.NewRequest("GET", "/api/analytics/exposure/heatmap?start_time="+
		string(rune(startTime.Unix()))+"&end_time="+string(rune(now.Unix()))+"&interval=15m", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleExposureHeatmap(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	requiredFields := []string{"timestamps", "symbols", "data", "max_exposure", "min_exposure", "interval"}
	for _, field := range requiredFields {
		if _, ok := response[field]; !ok {
			t.Errorf("Response missing '%s' field", field)
		}
	}
}

func TestHandleExposureHistory(t *testing.T) {
	// Create engine and handler
	engine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(engine)
	handler := NewAPIHandler(engine, pnlEngine)

	// Create test account
	account := engine.CreateAccount("test-user", "Test User", "password", true)
	engine.GetLedger().SetBalance(account.ID, 100000)
	account.Balance = 100000

	// Set up price callback
	engine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		return 1.1000, 1.1002, true
	})

	// Execute some test orders
	_, err := engine.ExecuteMarketOrder(account.ID, "EURUSD", "BUY", 1.0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to execute order: %v", err)
	}

	// Create request for EURUSD history
	req := httptest.NewRequest("GET", "/api/analytics/exposure/history/EURUSD?interval=1h", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.HandleExposureHistory(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response ExposureTimeline
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Symbol != "EURUSD" {
		t.Errorf("Expected symbol EURUSD, got %s", response.Symbol)
	}

	if len(response.Timeline) == 0 {
		t.Error("Expected non-empty timeline")
	}

	// Verify timeline entry structure
	if len(response.Timeline) > 0 {
		entry := response.Timeline[0]
		if entry.Timestamp == 0 {
			t.Error("Expected non-zero timestamp")
		}
	}
}

func TestCalculateSymbolExposures(t *testing.T) {
	// Create engine and handler
	engine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(engine)
	handler := NewAPIHandler(engine, pnlEngine)

	// Create mock positions
	positions := []*core.Position{
		{
			ID:           1,
			Symbol:       "EURUSD",
			Side:         "BUY",
			Volume:       1.0,
			OpenPrice:    1.1000,
			CurrentPrice: 1.1010,
		},
		{
			ID:           2,
			Symbol:       "EURUSD",
			Side:         "SELL",
			Volume:       0.5,
			OpenPrice:    1.1005,
			CurrentPrice: 1.1010,
		},
		{
			ID:           3,
			Symbol:       "GBPUSD",
			Side:         "BUY",
			Volume:       2.0,
			OpenPrice:    1.2500,
			CurrentPrice: 1.2510,
		},
	}

	// Calculate exposures
	exposures := handler.calculateSymbolExposures(positions)

	// Verify we have 2 symbols
	if len(exposures) != 2 {
		t.Fatalf("Expected 2 symbols, got %d", len(exposures))
	}

	// Find EURUSD exposure
	var eurusdExposure *SymbolExposure
	for i := range exposures {
		if exposures[i].Symbol == "EURUSD" {
			eurusdExposure = &exposures[i]
			break
		}
	}

	if eurusdExposure == nil {
		t.Fatal("EURUSD exposure not found")
	}

	// Verify EURUSD has long and short positions
	if eurusdExposure.Long == 0 {
		t.Error("Expected non-zero long exposure for EURUSD")
	}

	if eurusdExposure.Short == 0 {
		t.Error("Expected non-zero short exposure for EURUSD")
	}

	// Net exposure should be positive (1.0 lots long - 0.5 lots short)
	if eurusdExposure.NetExposure <= 0 {
		t.Error("Expected positive net exposure for EURUSD")
	}
}

func TestParseInterval(t *testing.T) {
	engine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(engine)
	handler := NewAPIHandler(engine, pnlEngine)

	tests := []struct {
		interval string
		expected time.Duration
	}{
		{"15m", 15 * time.Minute},
		{"1h", 1 * time.Hour},
		{"4h", 4 * time.Hour},
		{"1d", 24 * time.Hour},
		{"invalid", 1 * time.Hour}, // default
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			result := handler.parseInterval(tt.interval)
			if result != tt.expected {
				t.Errorf("parseInterval(%s) = %v, expected %v", tt.interval, result, tt.expected)
			}
		})
	}
}
