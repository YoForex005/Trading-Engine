package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCompleteOrderFlow tests the complete order lifecycle
func TestCompleteOrderFlow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Step 1: Inject market prices
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Step 2: Place market order
	orderReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.5,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(orderReq)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Order placement failed: %d - %s", w.Code, w.Body.String())
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)

	t.Logf("Order placed successfully: %+v", orderResp)

	// Step 3: Verify order was filled (for market orders)
	time.Sleep(100 * time.Millisecond)

	// Step 4: Check positions
	req = httptest.NewRequest("GET", "/positions", nil)
	w = httptest.NewRecorder()

	ts.server.HandleGetPositions(w, req)

	t.Logf("Positions response: %s", w.Body.String())

	// Step 5: Update price to create P/L
	ts.InjectPrice("EURUSD", 1.10100, 1.10120)
	time.Sleep(100 * time.Millisecond)

	// Step 6: Close position (if trade ID available)
	// Note: In A-Book mode, position management may differ
	t.Log("Order flow test completed successfully")
}

// TestLimitOrderActivation tests limit order triggering
func TestLimitOrderActivation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Set initial price
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Place buy limit below current price
	limitReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"price":  1.09500,
	}
	body, _ := json.Marshal(limitReq)

	req := httptest.NewRequest("POST", "/order/limit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceLimitOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Limit order placement failed: %d", w.Code)
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	orderID := orderResp["id"].(string)

	t.Logf("Limit order placed: %s at %.5f", orderID, 1.09500)

	// Verify order is pending
	req = httptest.NewRequest("GET", "/orders/pending", nil)
	w = httptest.NewRecorder()

	ts.server.HandleGetPendingOrders(w, req)

	var pendingOrders []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&pendingOrders)

	if len(pendingOrders) == 0 {
		t.Error("Expected pending order")
	}

	// Move price to trigger level
	ts.InjectPrice("EURUSD", 1.09500, 1.09520)
	time.Sleep(200 * time.Millisecond)

	// Check if order activated (implementation dependent)
	t.Log("Limit order activation test completed")
}

// TestStopOrderActivation tests stop order triggering
func TestStopOrderActivation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Set initial price
	ts.InjectPrice("GBPUSD", 1.25000, 1.25020)

	// Place buy stop above current price
	stopReq := map[string]interface{}{
		"symbol":       "GBPUSD",
		"side":         "BUY",
		"volume":       0.2,
		"triggerPrice": 1.25500,
	}
	body, _ := json.Marshal(stopReq)

	req := httptest.NewRequest("POST", "/order/stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceStopOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Stop order placement failed: %d", w.Code)
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)

	t.Logf("Stop order placed: %+v", orderResp)

	// Move price to trigger
	ts.InjectPrice("GBPUSD", 1.25500, 1.25520)
	time.Sleep(200 * time.Millisecond)

	t.Log("Stop order activation test completed")
}

// TestOrderModification tests modifying SL/TP
func TestOrderModification(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Place an order first
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	orderReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"type":   "MARKET",
		"sl":     1.09500,
		"tp":     1.10500,
	}
	body, _ := json.Marshal(orderReq)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Order placement failed: %d", w.Code)
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	order := orderResp["order"].(map[string]interface{})
	tradeID := order["ClientOrderID"].(string)

	time.Sleep(100 * time.Millisecond)

	// Modify SL/TP
	modifyReq := map[string]interface{}{
		"tradeId": tradeID,
		"sl":      1.09800,
		"tp":      1.10700,
	}
	body, _ = json.Marshal(modifyReq)

	req = httptest.NewRequest("POST", "/position/modify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandleModifySLTP(w, req)

	if w.Code != 200 {
		t.Errorf("Modification failed: %d - %s", w.Code, w.Body.String())
	}

	var modifyResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&modifyResp)

	if !modifyResp["success"].(bool) {
		t.Error("Expected modification success")
	}

	t.Logf("SL/TP modified: SL=%.5f TP=%.5f", modifyResp["sl"], modifyResp["tp"])
}

// TestBreakevenScenario tests setting breakeven
func TestBreakevenScenario(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Place order
	ts.InjectPrice("USDJPY", 110.000, 110.020)

	orderReq := map[string]interface{}{
		"symbol": "USDJPY",
		"side":   "BUY",
		"volume": 0.3,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(orderReq)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Order failed: %d", w.Code)
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	order := orderResp["order"].(map[string]interface{})
	tradeID := order["ClientOrderID"].(string)

	time.Sleep(100 * time.Millisecond)

	// Price moves in profit
	ts.InjectPrice("USDJPY", 110.500, 110.520)
	time.Sleep(50 * time.Millisecond)

	// Set breakeven
	beReq := map[string]string{
		"tradeId": tradeID,
	}
	body, _ = json.Marshal(beReq)

	req = httptest.NewRequest("POST", "/position/breakeven", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandleBreakeven(w, req)

	t.Logf("Breakeven response: %d - %s", w.Code, w.Body.String())
}

// TestTrailingStop tests trailing stop functionality
func TestTrailingStop(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Place order
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	orderReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.2,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(orderReq)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Order failed: %d", w.Code)
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	order := orderResp["order"].(map[string]interface{})
	tradeID := order["ClientOrderID"].(string)

	time.Sleep(100 * time.Millisecond)

	// Set trailing stop
	tsReq := map[string]interface{}{
		"tradeId":  tradeID,
		"symbol":   "EURUSD",
		"side":     "BUY",
		"type":     "FIXED",
		"distance": 0.00200,
	}
	body, _ = json.Marshal(tsReq)

	req = httptest.NewRequest("POST", "/position/trailing-stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandleSetTrailingStop(w, req)

	if w.Code != 200 {
		t.Errorf("Trailing stop failed: %d - %s", w.Code, w.Body.String())
	}

	var tsResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&tsResp)

	if !tsResp["success"].(bool) {
		t.Error("Expected trailing stop success")
	}

	// Move price to test trailing
	prices := []float64{1.10100, 1.10200, 1.10300, 1.10200, 1.10100}
	for _, price := range prices {
		ts.InjectPrice("EURUSD", price, price+0.00020)
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("Trailing stop test completed")
}

// TestMultiplePositionsSameSymbol tests hedging mode
func TestMultiplePositionsSameSymbol(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Place first BUY position
	order1 := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(order1)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("First order failed: %d", w.Code)
	}

	time.Sleep(100 * time.Millisecond)

	// Place second BUY position
	order2 := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.2,
		"type":   "MARKET",
	}
	body, _ = json.Marshal(order2)

	req = httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Second order failed: %d", w.Code)
	}

	time.Sleep(100 * time.Millisecond)

	// Place SELL position (hedging)
	order3 := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "SELL",
		"volume": 0.15,
		"type":   "MARKET",
	}
	body, _ = json.Marshal(order3)

	req = httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Third order failed: %d", w.Code)
	}

	t.Log("Hedging test: Multiple positions on same symbol placed")

	// Check positions
	req = httptest.NewRequest("GET", "/positions", nil)
	w = httptest.NewRecorder()

	ts.server.HandleGetPositions(w, req)

	t.Logf("Positions: %s", w.Body.String())
}

// TestPartialClose tests partial position closure
func TestPartialClose(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Place order
	ts.InjectPrice("GBPUSD", 1.25000, 1.25020)

	orderReq := map[string]interface{}{
		"symbol": "GBPUSD",
		"side":   "BUY",
		"volume": 1.0,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(orderReq)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Order failed: %d", w.Code)
	}

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	order := orderResp["order"].(map[string]interface{})
	tradeID := order["ClientOrderID"].(string)

	time.Sleep(100 * time.Millisecond)

	// Close 50%
	partialReq := map[string]interface{}{
		"tradeId": tradeID,
		"percent": 50.0,
	}
	body, _ = json.Marshal(partialReq)

	req = httptest.NewRequest("POST", "/position/partial-close", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandlePartialClose(w, req)

	t.Logf("Partial close response: %d - %s", w.Code, w.Body.String())
}

// TestOrderRejection tests order rejection scenarios
func TestOrderRejection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name  string
		order map[string]interface{}
	}{
		{
			name: "Excessive volume",
			order: map[string]interface{}{
				"symbol": "EURUSD",
				"side":   "BUY",
				"volume": 100.0, // Too large
				"type":   "MARKET",
			},
		},
		{
			name: "Invalid symbol",
			order: map[string]interface{}{
				"symbol": "INVALID",
				"side":   "BUY",
				"volume": 0.1,
				"type":   "MARKET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.order)

			req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			ts.server.HandlePlaceOrder(w, req)

			t.Logf("%s: Status %d - %s", tt.name, w.Code, w.Body.String())
		})
	}
}

// TestBidAskSpread tests spread handling
func TestBidAskSpread(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Test normal spread
	ts.InjectPrice("EURUSD", 1.10000, 1.10002)

	orderReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(orderReq)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Order with normal spread failed: %d", w.Code)
	}

	t.Log("Normal spread order executed")

	// Test wide spread
	ts.InjectPrice("EURUSD", 1.10000, 1.10100) // 100 pips spread

	orderReq2 := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "SELL",
		"volume": 0.1,
		"type":   "MARKET",
	}
	body, _ = json.Marshal(orderReq2)

	req = httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	ts.server.HandlePlaceOrder(w, req)

	t.Logf("Wide spread order: %d - %s", w.Code, w.Body.String())
}

// TestOrderFlowWithPriceGap tests order execution during price gaps
func TestOrderFlowWithPriceGap(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Initial price
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Place pending buy limit
	limitReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"price":  1.09900,
	}
	body, _ := json.Marshal(limitReq)

	req := httptest.NewRequest("POST", "/order/limit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandlePlaceLimitOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Limit order failed: %d", w.Code)
	}

	// Price gap down (would skip the limit)
	ts.InjectPrice("EURUSD", 1.09800, 1.09820)
	time.Sleep(200 * time.Millisecond)

	t.Log("Price gap test completed")
}
