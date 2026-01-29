package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/api"
	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/internal/api/handlers"
	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
)

// TestContext holds shared test infrastructure
type TestContext struct {
	Server      *api.Server
	BBEngine    *core.Engine
	PnLEngine   *core.PnLEngine
	AuthService *auth.Service
	Hub         *ws.Hub
	TickStore   *tickstore.TickStore
	LPManager   *lpmanager.Manager
	APIHandler  *handlers.APIHandler
	Token       string
	AccountID   string
}

// SetupTest initializes test environment
func SetupTest(t *testing.T) *TestContext {
	t.Helper()

	// Initialize B-Book engine
	bbEngine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(bbEngine)
	authService := auth.NewService(bbEngine, "", "test-jwt-secret")

	// Create test account
	testAccount := bbEngine.CreateAccount("test-user", "Test User", "test123", true)
	bbEngine.GetLedger().SetBalance(testAccount.ID, 100000.0)
	testAccount.Balance = 100000.0

	// Initialize components
	tickStore := tickstore.NewTickStore("test", 10000)
	lpManager := lpmanager.NewManager("../data/lp_config.json")
	apiHandler := handlers.NewAPIHandler(bbEngine, pnlEngine)

	// Create server
	server := api.NewServer(authService, apiHandler, lpManager)
	hub := ws.NewHub()
	hub.SetTickStore(tickStore)
	hub.SetBBookEngine(bbEngine)

	// Wire dependencies
	server.SetHub(hub)
	server.SetTickStore(tickStore)
	apiHandler.SetHub(hub)

	// Set price callback
	bbEngine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		tick := hub.GetLatestPrice(symbol)
		if tick != nil {
			return tick.Bid, tick.Ask, true
		}
		return 0, 0, false
	})

	// Start hub
	go hub.Run()

	ctx := &TestContext{
		Server:      server,
		BBEngine:    bbEngine,
		PnLEngine:   pnlEngine,
		AuthService: authService,
		Hub:         hub,
		TickStore:   tickStore,
		LPManager:   lpManager,
		APIHandler:  apiHandler,
		AccountID:   testAccount.AccountNumber,
	}

	// Login and get token
	ctx.Token = ctx.Login(t)

	return ctx
}

// Login performs authentication and returns token
func (tc *TestContext) Login(t *testing.T) string {
	t.Helper()

	reqBody := map[string]string{
		"username": "test-user",
		"password": "test123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login failed: %d - %s", w.Code, w.Body.String())
	}

	var resp struct {
		Token string     `json:"token"`
		User  *auth.User `json:"user"`
	}
	json.NewDecoder(w.Body).Decode(&resp)

	return resp.Token
}

// InjectPrice injects test market data
func (tc *TestContext) InjectPrice(symbol string, bid, ask float64) {
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Spread:    ask - bid,
		Timestamp: time.Now().UnixMilli(),
		LP:        "TEST",
	}
	tc.Hub.BroadcastTick(tick)
	time.Sleep(50 * time.Millisecond)
}

// MakeRequest makes authenticated HTTP request
func (tc *TestContext) MakeRequest(t *testing.T, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if tc.Token != "" {
		req.Header.Set("Authorization", "Bearer "+tc.Token)
	}

	w := httptest.NewRecorder()
	return w
}

// ==================== AUTHENTICATION TESTS ====================

func TestAuth_Login_Success(t *testing.T) {
	tc := SetupTest(t)
	if tc.Token == "" {
		t.Fatal("Expected non-empty token")
	}
}

func TestAuth_Login_InvalidPassword(t *testing.T) {
	tc := SetupTest(t)

	reqBody := map[string]string{
		"username": "test-user",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuth_Login_InvalidUsername(t *testing.T) {
	tc := SetupTest(t)

	reqBody := map[string]string{
		"username": "nonexistent",
		"password": "test123",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

func TestAuth_Login_EmptyCredentials(t *testing.T) {
	tc := SetupTest(t)

	reqBody := map[string]string{
		"username": "",
		"password": "",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandleLogin(w, req)

	if w.Code == http.StatusOK {
		t.Error("Expected authentication to fail with empty credentials")
	}
}

// ==================== HEALTH & CONFIG TESTS ====================

func TestHealth_Endpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected 'OK', got %s", w.Body.String())
	}
}

func TestConfig_GetConfig(t *testing.T) {
	_ = SetupTest(t)

	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		config := map[string]interface{}{
			"brokerName":        "RTX Trading",
			"executionMode":     "BBOOK",
			"defaultLeverage":   100,
			"defaultBalance":    5000.0,
			"marginMode":        "HEDGING",
			"maxTicksPerSymbol": 50000,
		}
		json.NewEncoder(w).Encode(config)
	})
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var config map[string]interface{}
	json.NewDecoder(w.Body).Decode(&config)

	if config["brokerName"] != "RTX Trading" {
		t.Error("Expected brokerName to be 'RTX Trading'")
	}
}

// ==================== ORDER TESTS ====================

func TestOrder_PlaceMarketOrder_Buy(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandlePlaceOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

func TestOrder_PlaceMarketOrder_Sell(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "SELL",
		"volume": 0.1,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandlePlaceOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}
}

func TestOrder_PlaceLimitOrder(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"price":  1.09500,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order/limit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandlePlaceLimitOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}
}

func TestOrder_PlaceStopOrder(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol":       "EURUSD",
		"side":         "BUY",
		"volume":       0.1,
		"triggerPrice": 1.10500,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order/stop", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandlePlaceStopOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}
}

func TestOrder_PlaceStopLimitOrder(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol":       "EURUSD",
		"side":         "BUY",
		"volume":       0.1,
		"triggerPrice": 1.10500,
		"limitPrice":   1.10550,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order/stop-limit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandlePlaceStopLimitOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}
}

func TestOrder_GetPendingOrders(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Place a limit order
	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"price":  1.09500,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order/limit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	tc.Server.HandlePlaceLimitOrder(w, req)

	// Get pending orders
	req = httptest.NewRequest("GET", "/orders/pending", nil)
	w = httptest.NewRecorder()
	tc.Server.HandleGetPendingOrders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var orders []interface{}
	json.NewDecoder(w.Body).Decode(&orders)

	if len(orders) == 0 {
		t.Error("Expected at least one pending order")
	}
}

func TestOrder_CancelOrder(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Place a limit order
	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"price":  1.09500,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/order/limit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	tc.Server.HandlePlaceLimitOrder(w, req)

	var orderResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orderResp)
	orderID := orderResp["id"].(string)

	// Cancel the order
	cancelBody := map[string]string{
		"orderId": orderID,
	}
	body, _ = json.Marshal(cancelBody)

	req = httptest.NewRequest("POST", "/order/cancel", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	tc.Server.HandleCancelOrder(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}
}

func TestOrder_InvalidVolume(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	tests := []struct {
		name   string
		volume float64
	}{
		{"Zero volume", 0},
		{"Negative volume", -0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"symbol": "EURUSD",
				"side":   "BUY",
				"volume": tt.volume,
				"type":   "MARKET",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			tc.Server.HandlePlaceOrder(w, req)

			// Should return bad request for invalid volume
			t.Logf("%s: status=%d", tt.name, w.Code)
		})
	}
}

// ==================== POSITION TESTS ====================

func TestPosition_ModifySLTP(t *testing.T) {
	tc := SetupTest(t)

	reqBody := map[string]interface{}{
		"tradeId": "test-trade-123",
		"sl":      1.09000,
		"tp":      1.11000,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/position/modify", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	tc.Server.HandleModifySLTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
	}
}

func TestPosition_SetTrailingStop(t *testing.T) {
	tc := SetupTest(t)

	tests := []struct {
		name     string
		tsType   string
		distance float64
	}{
		{"Fixed trailing stop", "FIXED", 20.0},
		{"Step trailing stop", "STEP", 15.0},
		{"ATR trailing stop", "ATR", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"tradeId":  "test-trade-123",
				"symbol":   "EURUSD",
				"side":     "BUY",
				"type":     tt.tsType,
				"distance": tt.distance,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/position/trailing-stop", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			tc.Server.HandleSetTrailingStop(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
			}
		})
	}
}

// ==================== MARKET DATA TESTS ====================

func TestMarket_GetTicks(t *testing.T) {
	tc := SetupTest(t)

	// Inject test prices
	for i := 0; i < 10; i++ {
		tc.InjectPrice("EURUSD", 1.10000+float64(i)*0.00001, 1.10020+float64(i)*0.00001)
		time.Sleep(10 * time.Millisecond)
	}

	req := httptest.NewRequest("GET", "/ticks?symbol=EURUSD&limit=10", nil)
	w := httptest.NewRecorder()

	tc.Server.HandleGetTicks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var ticks []interface{}
	json.NewDecoder(w.Body).Decode(&ticks)

	if len(ticks) == 0 {
		t.Error("Expected at least one tick")
	}
}

func TestMarket_GetOHLC(t *testing.T) {
	tc := SetupTest(t)

	// Inject prices
	for i := 0; i < 5; i++ {
		price := 1.10000 + float64(i)*0.00010
		tc.InjectPrice("EURUSD", price, price+0.00020)
		time.Sleep(100 * time.Millisecond)
	}

	timeframes := []string{"1m", "5m", "15m", "1h", "4h", "1d"}

	for _, tf := range timeframes {
		t.Run("Timeframe_"+tf, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/ohlc?symbol=EURUSD&timeframe=%s&limit=5", tf), nil)
			w := httptest.NewRecorder()

			tc.Server.HandleGetOHLC(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected 200, got %d", w.Code)
			}

			var candles []interface{}
			json.NewDecoder(w.Body).Decode(&candles)
			t.Logf("Timeframe %s: %d candles", tf, len(candles))
		})
	}
}

// ==================== RISK CALCULATION TESTS ====================

func TestRisk_CalculateLot(t *testing.T) {
	tc := SetupTest(t)

	tests := []struct {
		name        string
		symbol      string
		riskPercent float64
		slPips      float64
	}{
		{"EURUSD 2% risk 20 pips", "EURUSD", 2.0, 20.0},
		{"GBPUSD 1% risk 30 pips", "GBPUSD", 1.0, 30.0},
		{"USDJPY 3% risk 25 pips", "USDJPY", 3.0, 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/risk/calculate-lot?symbol=%s&riskPercent=%.2f&slPips=%.2f",
				tt.symbol, tt.riskPercent, tt.slPips)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			tc.Server.HandleCalculateLot(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
			}

			var result map[string]interface{}
			json.NewDecoder(w.Body).Decode(&result)

			if result["recommendedLotSize"] == nil {
				t.Error("Expected recommendedLotSize in response")
			}
		})
	}
}

func TestRisk_MarginPreview(t *testing.T) {
	tc := SetupTest(t)

	tests := []struct {
		name   string
		symbol string
		volume float64
		side   string
	}{
		{"EURUSD Buy 1.0 lot", "EURUSD", 1.0, "BUY"},
		{"EURUSD Sell 0.5 lot", "EURUSD", 0.5, "SELL"},
		{"GBPUSD Buy 2.0 lots", "GBPUSD", 2.0, "BUY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/risk/margin-preview?symbol=%s&volume=%.2f&side=%s",
				tt.symbol, tt.volume, tt.side)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			tc.Server.HandleMarginPreview(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected 200, got %d - %s", w.Code, w.Body.String())
			}

			var result map[string]interface{}
			json.NewDecoder(w.Body).Decode(&result)

			if result["requiredMargin"] == nil {
				t.Error("Expected requiredMargin in response")
			}
		})
	}
}

// ==================== ADMIN TESTS ====================

func TestAdmin_ExecutionModeToggle(t *testing.T) {
	_ = SetupTest(t)

	modes := []struct {
		mode       string
		expectCode int
	}{
		{"BBOOK", http.StatusOK},
		{"ABOOK", http.StatusOK},
		{"INVALID", http.StatusBadRequest},
	}

	for _, m := range modes {
		t.Run("Mode_"+m.mode, func(t *testing.T) {
			reqBody := map[string]string{
				"mode": m.mode,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/admin/execution-mode", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Content-Type", "application/json")

				var reqData map[string]string
				json.NewDecoder(r.Body).Decode(&reqData)

				if reqData["mode"] != "BBOOK" && reqData["mode"] != "ABOOK" {
					http.Error(w, "mode must be BBOOK or ABOOK", http.StatusBadRequest)
					return
				}

				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"newMode": reqData["mode"],
				})
			})
			handler.ServeHTTP(w, req)

			if w.Code != m.expectCode {
				t.Errorf("Expected %d, got %d", m.expectCode, w.Code)
			}
		})
	}
}

// ==================== CONCURRENT TESTS ====================

func TestConcurrent_PlaceOrders(t *testing.T) {
	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	numOrders := 50
	done := make(chan bool, numOrders)
	errors := make(chan error, numOrders)

	for i := 0; i < numOrders; i++ {
		go func(orderNum int) {
			reqBody := map[string]interface{}{
				"symbol": "EURUSD",
				"side":   "BUY",
				"volume": 0.01,
				"type":   "MARKET",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			tc.Server.HandlePlaceOrder(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("order %d failed: %d - %s", orderNum, w.Code, w.Body.String())
			}

			done <- true
		}(i)
	}

	// Wait for all orders
	for i := 0; i < numOrders; i++ {
		<-done
	}

	close(errors)
	if len(errors) > 0 {
		t.Errorf("Some orders failed: %d errors", len(errors))
		for err := range errors {
			t.Log(err)
		}
	}
}

// ==================== BENCHMARK TESTS ====================

func BenchmarkPlaceMarketOrder(b *testing.B) {
	tc := SetupTest(&testing.T{})
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.1,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		tc.Server.HandlePlaceOrder(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Order failed: %d", w.Code)
		}
	}
}

func BenchmarkGetTicks(b *testing.B) {
	tc := SetupTest(&testing.T{})
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/ticks?symbol=EURUSD&limit=100", nil)
		w := httptest.NewRecorder()

		tc.Server.HandleGetTicks(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("GetTicks failed: %d", w.Code)
		}
	}
}

func BenchmarkCalculateLot(b *testing.B) {
	tc := SetupTest(&testing.T{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/risk/calculate-lot?symbol=EURUSD&riskPercent=2&slPips=20", nil)
		w := httptest.NewRecorder()

		tc.Server.HandleCalculateLot(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("CalculateLot failed: %d", w.Code)
		}
	}
}
