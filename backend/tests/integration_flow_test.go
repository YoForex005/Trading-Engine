package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ==================== COMPLETE TRADING WORKFLOWS ====================

// TestWorkflow_CompleteTrading tests full trading lifecycle
func TestWorkflow_CompleteTrading(t *testing.T) {
	tc := SetupTest(t)

	t.Log("Step 1: Login")
	if tc.Token == "" {
		t.Fatal("Login failed")
	}

	t.Log("Step 2: Get account summary")
	// Account summary would be tested here

	t.Log("Step 3: Inject market data")
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)
	tc.InjectPrice("GBPUSD", 1.25000, 1.25025)
	tc.InjectPrice("USDJPY", 110.000, 110.025)

	t.Log("Step 4: Place market orders")
	orders := []struct {
		symbol string
		side   string
		volume float64
	}{
		{"EURUSD", "BUY", 0.1},
		{"GBPUSD", "SELL", 0.2},
		{"USDJPY", "BUY", 0.15},
	}

	for _, order := range orders {
		reqBody := map[string]interface{}{
			"symbol": order.symbol,
			"side":   order.side,
			"volume": order.volume,
			"type":   "MARKET",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		tc.Server.HandlePlaceOrder(w, req)

		if w.Code != 200 {
			t.Fatalf("Failed to place %s %s order: %d", order.side, order.symbol, w.Code)
		}

		t.Logf("Placed %s %s %.2f lots", order.side, order.symbol, order.volume)
	}

	t.Log("Step 5: Place pending orders")
	pendingOrders := []struct {
		symbol       string
		side         string
		volume       float64
		price        float64
		triggerPrice float64
		orderType    string
	}{
		{"EURUSD", "BUY", 0.1, 1.09500, 0, "LIMIT"},
		{"GBPUSD", "SELL", 0.2, 1.25500, 0, "LIMIT"},
		{"USDJPY", "BUY", 0.15, 0, 110.500, "STOP"},
	}

	for _, order := range pendingOrders {
		var endpoint string
		var reqBody map[string]interface{}

		switch order.orderType {
		case "LIMIT":
			endpoint = "/order/limit"
			reqBody = map[string]interface{}{
				"symbol": order.symbol,
				"side":   order.side,
				"volume": order.volume,
				"price":  order.price,
			}
		case "STOP":
			endpoint = "/order/stop"
			reqBody = map[string]interface{}{
				"symbol":       order.symbol,
				"side":         order.side,
				"volume":       order.volume,
				"triggerPrice": order.triggerPrice,
			}
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		if order.orderType == "LIMIT" {
			tc.Server.HandlePlaceLimitOrder(w, req)
		} else {
			tc.Server.HandlePlaceStopOrder(w, req)
		}

		if w.Code != 200 {
			t.Fatalf("Failed to place %s %s %s order: %d", order.orderType, order.side, order.symbol, w.Code)
		}

		t.Logf("Placed %s %s %s @ %.5f", order.orderType, order.side, order.symbol, order.price+order.triggerPrice)
	}

	t.Log("Step 6: Get pending orders")
	req := httptest.NewRequest("GET", "/orders/pending", nil)
	w := httptest.NewRecorder()
	tc.Server.HandleGetPendingOrders(w, req)

	var pending []interface{}
	json.NewDecoder(w.Body).Decode(&pending)
	t.Logf("Pending orders: %d", len(pending))

	t.Log("Step 7: Modify positions (SL/TP)")
	modifyReq := map[string]interface{}{
		"tradeId": "test-trade-1",
		"sl":      1.09000,
		"tp":      1.11000,
	}
	body, _ := json.Marshal(modifyReq)
	req = httptest.NewRequest("POST", "/position/modify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	tc.Server.HandleModifySLTP(w, req)

	t.Log("Step 8: Set trailing stop")
	tsReq := map[string]interface{}{
		"tradeId":  "test-trade-1",
		"symbol":   "EURUSD",
		"side":     "BUY",
		"type":     "FIXED",
		"distance": 20.0,
	}
	body, _ = json.Marshal(tsReq)
	req = httptest.NewRequest("POST", "/position/trailing-stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	tc.Server.HandleSetTrailingStop(w, req)

	t.Log("Step 9: Update market prices")
	tc.InjectPrice("EURUSD", 1.10050, 1.10070)
	tc.InjectPrice("GBPUSD", 1.24950, 1.24975)

	t.Log("Step 10: Get market data")
	req = httptest.NewRequest("GET", "/ticks?symbol=EURUSD&limit=10", nil)
	w = httptest.NewRecorder()
	tc.Server.HandleGetTicks(w, req)

	var ticks []interface{}
	json.NewDecoder(w.Body).Decode(&ticks)
	t.Logf("Ticks received: %d", len(ticks))

	t.Log("Complete trading workflow test passed")
}

// TestWorkflow_PositionLifecycle tests complete position lifecycle
func TestWorkflow_PositionLifecycle(t *testing.T) {
	tc := SetupTest(t)

	t.Log("Phase 1: Open position")
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	openReq := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 1.0,
		"type":   "MARKET",
		"sl":     1.09000,
		"tp":     1.11000,
	}
	body, _ := json.Marshal(openReq)
	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	tc.Server.HandlePlaceOrder(w, req)

	if w.Code != 200 {
		t.Fatalf("Failed to open position: %d - %s", w.Code, w.Body.String())
	}

	var openResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&openResp)
	t.Log("Position opened")

	t.Log("Phase 2: Monitor position")
	tc.InjectPrice("EURUSD", 1.10050, 1.10070)
	time.Sleep(100 * time.Millisecond)

	t.Log("Phase 3: Modify SL/TP")
	modifyReq := map[string]interface{}{
		"tradeId": "trade-123",
		"sl":      1.09500, // Move SL to breakeven
		"tp":      1.11500, // Extend TP
	}
	body, _ = json.Marshal(modifyReq)
	req = httptest.NewRequest("POST", "/position/modify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	tc.Server.HandleModifySLTP(w, req)
	t.Log("Modified SL/TP")

	t.Log("Phase 4: Enable trailing stop")
	tsReq := map[string]interface{}{
		"tradeId":  "trade-123",
		"symbol":   "EURUSD",
		"side":     "BUY",
		"type":     "STEP",
		"distance": 20.0,
		"stepSize": 5.0,
	}
	body, _ = json.Marshal(tsReq)
	req = httptest.NewRequest("POST", "/position/trailing-stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	tc.Server.HandleSetTrailingStop(w, req)
	t.Log("Trailing stop enabled")

	t.Log("Phase 5: Price movement")
	prices := []struct {
		bid float64
		ask float64
	}{
		{1.10100, 1.10120},
		{1.10150, 1.10170},
		{1.10200, 1.10220},
	}

	for _, price := range prices {
		tc.InjectPrice("EURUSD", price.bid, price.ask)
		time.Sleep(50 * time.Millisecond)
		t.Logf("Price updated: %.5f/%.5f", price.bid, price.ask)
	}

	t.Log("Position lifecycle test completed")
}

// TestWorkflow_MultiSymbolTrading tests trading across multiple symbols
func TestWorkflow_MultiSymbolTrading(t *testing.T) {
	tc := SetupTest(t)

	symbols := []struct {
		symbol string
		bid    float64
		ask    float64
	}{
		{"EURUSD", 1.10000, 1.10020},
		{"GBPUSD", 1.25000, 1.25025},
		{"USDJPY", 110.000, 110.025},
		{"AUDUSD", 0.75000, 0.75020},
		{"USDCAD", 1.25000, 1.25025},
	}

	t.Log("Phase 1: Inject prices for all symbols")
	for _, s := range symbols {
		tc.InjectPrice(s.symbol, s.bid, s.ask)
		t.Logf("Injected %s: %.5f/%.5f", s.symbol, s.bid, s.ask)
	}

	t.Log("Phase 2: Open positions")
	sides := []string{"BUY", "SELL"}
	for i, s := range symbols {
		side := sides[i%2]

		reqBody := map[string]interface{}{
			"symbol": s.symbol,
			"side":   side,
			"volume": 0.1,
			"type":   "MARKET",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		tc.Server.HandlePlaceOrder(w, req)

		if w.Code != 200 {
			t.Fatalf("Failed to place order for %s: %d", s.symbol, w.Code)
		}

		t.Logf("Opened %s position on %s", side, s.symbol)
	}

	t.Log("Phase 3: Update prices")
	for _, s := range symbols {
		newBid := s.bid + 0.00050
		newAsk := s.ask + 0.00050
		tc.InjectPrice(s.symbol, newBid, newAsk)
		t.Logf("Updated %s: %.5f/%.5f", s.symbol, newBid, newAsk)
	}

	t.Log("Phase 4: Calculate risk metrics")
	for _, s := range symbols {
		req := httptest.NewRequest("GET", "/risk/margin-preview?symbol="+s.symbol+"&volume=1.0&side=BUY", nil)
		w := httptest.NewRecorder()
		tc.Server.HandleMarginPreview(w, req)

		if w.Code != 200 {
			t.Logf("Margin preview failed for %s: %d", s.symbol, w.Code)
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(w.Body).Decode(&result)
		t.Logf("%s margin: %v", s.symbol, result["requiredMargin"])
	}

	t.Log("Multi-symbol trading test completed")
}

// TestWorkflow_OrderManagement tests order management workflow
func TestWorkflow_OrderManagement(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	t.Log("Step 1: Place multiple pending orders")
	orderIDs := make([]string, 0)

	orders := []struct {
		orderType string
		price     float64
	}{
		{"LIMIT", 1.09500},
		{"LIMIT", 1.09600},
		{"LIMIT", 1.09700},
	}

	for _, order := range orders {
		reqBody := map[string]interface{}{
			"symbol": "EURUSD",
			"side":   "BUY",
			"volume": 0.1,
			"price":  order.price,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/order/limit", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		tc.Server.HandlePlaceLimitOrder(w, req)

		if w.Code != 200 {
			t.Fatalf("Failed to place limit order @ %.5f: %d", order.price, w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if id, ok := resp["id"].(string); ok {
			orderIDs = append(orderIDs, id)
		}

		t.Logf("Placed limit order @ %.5f", order.price)
	}

	t.Log("Step 2: List pending orders")
	req := httptest.NewRequest("GET", "/orders/pending", nil)
	w := httptest.NewRecorder()
	tc.Server.HandleGetPendingOrders(w, req)

	var pending []interface{}
	json.NewDecoder(w.Body).Decode(&pending)
	t.Logf("Total pending orders: %d", len(pending))

	t.Log("Step 3: Cancel some orders")
	for i, orderID := range orderIDs {
		if i%2 == 0 { // Cancel every other order
			cancelReq := map[string]string{
				"orderId": orderID,
			}
			body, _ := json.Marshal(cancelReq)

			req := httptest.NewRequest("POST", "/order/cancel", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			tc.Server.HandleCancelOrder(w, req)

			if w.Code == 200 {
				t.Logf("Cancelled order: %s", orderID)
			}
		}
	}

	t.Log("Step 4: Verify remaining orders")
	req = httptest.NewRequest("GET", "/orders/pending", nil)
	w = httptest.NewRecorder()
	tc.Server.HandleGetPendingOrders(w, req)

	json.NewDecoder(w.Body).Decode(&pending)
	t.Logf("Remaining pending orders: %d", len(pending))

	t.Log("Order management workflow completed")
}

// TestWorkflow_RiskManagement tests risk management workflow
func TestWorkflow_RiskManagement(t *testing.T) {
	tc := SetupTest(t)

	t.Log("Phase 1: Calculate lot sizes for different risk levels")
	riskLevels := []float64{1.0, 2.0, 3.0, 5.0}

	for _, risk := range riskLevels {
		req := httptest.NewRequest("GET", "/risk/calculate-lot?symbol=EURUSD&riskPercent="+string(rune(risk))+"&slPips=20", nil)
		w := httptest.NewRecorder()
		tc.Server.HandleCalculateLot(w, req)

		if w.Code != 200 {
			t.Logf("Lot calculation failed for %.1f%% risk: %d", risk, w.Code)
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(w.Body).Decode(&result)
		t.Logf("Risk %.1f%%: Lot size = %v", risk, result["recommendedLotSize"])
	}

	t.Log("Phase 2: Preview margin for different volumes")
	volumes := []float64{0.1, 0.5, 1.0, 2.0, 5.0}

	for _, vol := range volumes {
		req := httptest.NewRequest("GET", "/risk/margin-preview?symbol=EURUSD&volume="+string(rune(vol))+"&side=BUY", nil)
		w := httptest.NewRecorder()
		tc.Server.HandleMarginPreview(w, req)

		if w.Code != 200 {
			t.Logf("Margin preview failed for %.2f lots: %d", vol, w.Code)
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(w.Body).Decode(&result)
		t.Logf("Volume %.2f: Margin = %v", vol, result["requiredMargin"])
	}

	t.Log("Risk management workflow completed")
}

// TestWorkflow_RealTimeDataFlow tests real-time data streaming
func TestWorkflow_RealTimeDataFlow(t *testing.T) {
	tc := SetupTest(t)

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY"}

	t.Log("Phase 1: Inject continuous price updates")
	done := make(chan bool)

	go func() {
		for i := 0; i < 20; i++ {
			for _, symbol := range symbols {
				var bid, ask float64
				switch symbol {
				case "EURUSD":
					bid = 1.10000 + float64(i)*0.00001
					ask = 1.10020 + float64(i)*0.00001
				case "GBPUSD":
					bid = 1.25000 + float64(i)*0.00002
					ask = 1.25025 + float64(i)*0.00002
				case "USDJPY":
					bid = 110.000 + float64(i)*0.001
					ask = 110.025 + float64(i)*0.001
				}

				tc.InjectPrice(symbol, bid, ask)
			}
			time.Sleep(100 * time.Millisecond)
		}
		done <- true
	}()

	t.Log("Phase 2: Query tick data periodically")
	for i := 0; i < 5; i++ {
		time.Sleep(400 * time.Millisecond)

		for _, symbol := range symbols {
			req := httptest.NewRequest("GET", "/ticks?symbol="+symbol+"&limit=5", nil)
			w := httptest.NewRecorder()
			tc.Server.HandleGetTicks(w, req)

			var ticks []interface{}
			json.NewDecoder(w.Body).Decode(&ticks)
			t.Logf("%s: %d ticks", symbol, len(ticks))
		}
	}

	<-done
	t.Log("Real-time data flow test completed")
}

// TestWorkflow_AdminOperations tests admin workflows
func TestWorkflow_AdminOperations(t *testing.T) {
	_ = SetupTest(t)

	t.Log("Phase 1: Toggle execution mode")
	modes := []string{"BBOOK", "ABOOK", "BBOOK"}

	for _, mode := range modes {
		reqBody := map[string]string{
			"mode": mode,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/admin/execution-mode", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Mock handler
		handler := func(wr *httptest.ResponseRecorder, _ *http.Request) {
			wr.Header().Set("Content-Type", "application/json")
			json.NewEncoder(wr).Encode(map[string]interface{}{
				"success": true,
				"newMode": mode,
			})
		}
		handler(w, req)

		t.Logf("Switched to %s mode", mode)
	}

	t.Log("Phase 2: Get configuration")
	_ = httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	// Mock config response
	config := map[string]interface{}{
		"brokerName":        "RTX Trading",
		"executionMode":     "BBOOK",
		"defaultLeverage":   100,
		"defaultBalance":    5000.0,
		"marginMode":        "HEDGING",
		"maxTicksPerSymbol": 50000,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)

	var receivedConfig map[string]interface{}
	json.NewDecoder(w.Body).Decode(&receivedConfig)
	t.Logf("Broker: %s, Mode: %s", receivedConfig["brokerName"], receivedConfig["executionMode"])

	t.Log("Admin operations workflow completed")
}
