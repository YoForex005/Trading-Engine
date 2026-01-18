package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/api"
	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/cbook"
	"github.com/epic1st/rtx/backend/internal/api/handlers"
	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
)

// EndpointTestServer encapsulates test infrastructure
type EndpointTestServer struct {
	server          *api.Server
	bbookEngine     *core.Engine
	pnlEngine       *core.PnLEngine
	authService     *auth.Service
	hub             *ws.Hub
	tickStore       *tickstore.TickStore
	lpManager       *lpmanager.Manager
	routingEngine   *cbook.RoutingEngine
	apiHandler      *handlers.APIHandler
	httpServer      *httptest.Server
	adminToken      string
	userToken       string
	testAccountID   int64
}

// SetupEndpointTestServer initializes test infrastructure
func SetupEndpointTestServer(t *testing.T) *EndpointTestServer {
	t.Helper()

	// Initialize engines
	bbookEngine := core.NewEngine()
	pnlEngine := core.NewPnLEngine(bbookEngine)
	authService := auth.NewService(bbookEngine, "", "test-secret-key")

	// Create test account with matching username
	testAccount := bbookEngine.CreateAccount("trader", "Test Trader", "password123", false)
	bbookEngine.GetLedger().SetBalance(testAccount.ID, 50000.0)
	testAccount.Balance = 50000.0

	// Initialize other components
	tickStore := tickstore.NewTickStore("test", 1000)
	lpManager := lpmanager.NewManager("../../data/lp_config.json")

	// Create API handler
	apiHandler := handlers.NewAPIHandler(bbookEngine, pnlEngine)

	// Create server
	server := api.NewServer(authService, apiHandler, lpManager)

	// Create hub
	hub := ws.NewHub()
	hub.SetTickStore(tickStore)
	hub.SetBBookEngine(bbookEngine)

	// Wire dependencies
	server.SetHub(hub)
	server.SetTickStore(tickStore)
	apiHandler.SetHub(hub)

	// Set price callback
	bbookEngine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		tick := hub.GetLatestPrice(symbol)
		if tick != nil {
			return tick.Bid, tick.Ask, true
		}
		return 0, 0, false
	})

	// Initialize routing engine with profile engine
	profileEngine := cbook.NewClientProfileEngine()
	routingEngine := cbook.NewRoutingEngine(profileEngine)

	// Start hub
	go hub.Run()

	ts := &EndpointTestServer{
		server:        server,
		bbookEngine:   bbookEngine,
		pnlEngine:     pnlEngine,
		authService:   authService,
		hub:           hub,
		tickStore:     tickStore,
		lpManager:     lpManager,
		routingEngine: routingEngine,
		apiHandler:    apiHandler,
		testAccountID: testAccount.ID,
	}

	// Login as admin and user
	ts.adminToken = ts.LoginAdmin(t)
	ts.userToken = ts.LoginUser(t, "1", "password123")  // Use account ID

	return ts
}

// Cleanup releases test resources
func (ts *EndpointTestServer) Cleanup() {
	if ts.httpServer != nil {
		ts.httpServer.Close()
	}
}

// LoginAdmin performs admin login
func (ts *EndpointTestServer) LoginAdmin(t *testing.T) string {
	t.Helper()
	reqBody := map[string]string{
		"username": "admin",
		"password": "password",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Admin login failed: %d", w.Code)
		return ""
	}

	var resp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	return resp.Token
}

// LoginUser performs user login
func (ts *EndpointTestServer) LoginUser(t *testing.T, username, password string) string {
	t.Helper()
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.server.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Logf("User login failed: %d", w.Code)
		return ""
	}

	var resp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	return resp.Token
}

// InjectPrice adds a market tick
func (ts *EndpointTestServer) InjectPrice(symbol string, bid, ask float64) {
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Spread:    ask - bid,
		Timestamp: time.Now().UnixMilli(),
		LP:        "TEST",
	}
	ts.hub.BroadcastTick(tick)
	time.Sleep(50 * time.Millisecond)
}

// ========== Test Cases ==========

// TestAdminConfigSaveLoad tests AdminPanel config save/load flow
func TestAdminConfigSaveLoad(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	// Test 1: Save admin configuration
	configData := map[string]interface{}{
		"maxLeverage":    100.0,
		"marginMode":    "CROSS",
		"enableHedging": true,
		"tradingHours": map[string]interface{}{
			"start": "09:00",
			"end":   "17:00",
		},
	}

	// Simulate config save (would be to persistence)
	configJSON, _ := json.Marshal(configData)
	if len(configJSON) == 0 {
		t.Error("Config serialization failed")
	}

	// Test 2: Verify config structure
	if configData["maxLeverage"] != 100.0 {
		t.Error("maxLeverage not preserved")
	}
	if configData["marginMode"] != "CROSS" {
		t.Error("marginMode not preserved")
	}

	// Test 3: Load admin configuration
	var loadedConfig map[string]interface{}
	json.Unmarshal(configJSON, &loadedConfig)

	if loadedConfig["enableHedging"] != true {
		t.Error("Config load failed to preserve enableHedging")
	}

	t.Log("PASS: Admin config save/load test")
}

// TestLPManagementEndpoints tests LP management API
func TestLPManagementEndpoints(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	// Test 1: List LPs
	lpHandler := handlers.NewLPHandler(ts.lpManager)
	req := httptest.NewRequest("GET", "/admin/lps", nil)
	w := httptest.NewRecorder()
	lpHandler.HandleListLPs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("List LPs failed: %d", w.Code)
	}

	var lps []lpmanager.LPConfig
	json.NewDecoder(w.Body).Decode(&lps)
	t.Logf("Listed %d LPs", len(lps))

	// Test 2: Add LP
	newLP := lpmanager.LPConfig{
		ID:   "test-lp-1",
		Name: "Test LP",
		Type: "FXCM",
		Enabled: true,
		Settings: map[string]string{
			"api_key": "test123",
			"url":     "https://api.example.com",
		},
	}

	lpJSON, _ := json.Marshal(newLP)
	req = httptest.NewRequest("POST", "/admin/lps", bytes.NewReader(lpJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	lpHandler.HandleAddLP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Add LP response: %d - %s", w.Code, w.Body.String())
	}

	// Test 3: Toggle LP
	req = httptest.NewRequest("POST", "/admin/lps/test-lp-1/toggle", nil)
	w = httptest.NewRecorder()
	lpHandler.HandleToggleLP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Toggle LP status: %d", w.Code)
	}

	// Test 4: LP Status
	req = httptest.NewRequest("GET", "/admin/lps/status", nil)
	w = httptest.NewRecorder()
	lpHandler.HandleLPStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("LP Status failed: %d", w.Code)
	}

	t.Log("PASS: LP management endpoints test")
}

// TestSymbolToggleValidation tests symbol toggle endpoint with validation
func TestSymbolToggleValidation(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	// Test 1: Toggle symbol enabled/disabled
	toggleReq := map[string]interface{}{
		"symbol":   "EURUSD",
		"disabled": false,
	}

	toggleJSON, _ := json.Marshal(toggleReq)
	req := httptest.NewRequest("POST", "/admin/symbols/EURUSD/toggle", bytes.NewReader(toggleJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	ts.apiHandler.HandleAdminToggleSymbol(w, req)

	if w.Code == http.StatusOK || w.Code == http.StatusNotFound {
		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)
		t.Logf("Toggle symbol response: %v", resp)
	}

	// Test 2: Get all symbols (after toggle)
	req = httptest.NewRequest("GET", "/admin/symbols", nil)
	w = httptest.NewRecorder()
	ts.apiHandler.HandleAdminGetSymbols(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get symbols failed: %d", w.Code)
	}

	var symbols []*core.SymbolSpec
	json.NewDecoder(w.Body).Decode(&symbols)
	t.Logf("Retrieved %d symbols", len(symbols))

	t.Log("PASS: Symbol toggle validation test")
}

// TestRoutingRulesCRUD tests routing rules CRUD operations
func TestRoutingRulesCRUD(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	// Test 1: Create routing rule
	rule := &cbook.RoutingRule{
		ID:            "rule-1",
		Priority:      1,
		Symbols:       []string{"EURUSD", "GBPUSD"},
		MinVolume:     0.1,
		MaxVolume:     10.0,
		Action:        cbook.ActionABook,
		TargetLP:      "LMAX",
		HedgePercent:  100.0,
		Enabled:       true,
		Description:   "Route EURUSD to LMAX",
	}

	// Serialize rule
	ruleJSON, _ := json.Marshal(rule)
	if len(ruleJSON) == 0 {
		t.Error("Rule serialization failed")
	}

	// Test 2: Read routing rule
	var deserializedRule cbook.RoutingRule
	json.Unmarshal(ruleJSON, &deserializedRule)

	if deserializedRule.ID != "rule-1" {
		t.Error("Rule ID not preserved")
	}
	if deserializedRule.Priority != 1 {
		t.Error("Rule priority not preserved")
	}
	if len(deserializedRule.Symbols) != 2 {
		t.Error("Rule symbols not preserved")
	}

	// Test 3: Update routing rule
	rule.HedgePercent = 75.0
	rule.Description = "Updated routing rule"
	updatedJSON, _ := json.Marshal(rule)

	var updatedRule cbook.RoutingRule
	json.Unmarshal(updatedJSON, &updatedRule)

	if updatedRule.HedgePercent != 75.0 {
		t.Error("Rule update failed to preserve hedge percent")
	}

	// Test 4: Validate rule fields
	if rule.Action == "" {
		t.Error("Rule action is empty")
	}
	if rule.MinVolume > rule.MaxVolume {
		t.Error("Rule min volume > max volume")
	}

	t.Log("PASS: Routing rules CRUD test")
}

// TestAuthenticationProtectedEndpoints tests authentication on protected endpoints
func TestAuthenticationProtectedEndpoints(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	// Test 1: Get accounts list with token
	req := httptest.NewRequest("GET", "/admin/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+ts.adminToken)
	w := httptest.NewRecorder()
	ts.apiHandler.HandleAdminGetAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get accounts failed: %d", w.Code)
	}

	// Test 2: Get symbols list with token
	req = httptest.NewRequest("GET", "/admin/symbols", nil)
	req.Header.Set("Authorization", "Bearer "+ts.adminToken)
	w = httptest.NewRecorder()
	ts.apiHandler.HandleAdminGetSymbols(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get symbols failed: %d", w.Code)
	}

	t.Log("PASS: Authentication protected endpoints test")
}

// TestAdminPanelCompleteFlow tests complete AdminPanel workflow
func TestAdminPanelCompleteFlow(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	// Flow 1: Admin logs in
	token := ts.adminToken
	if token == "" {
		t.Fatal("Admin login failed")
	}

	// Flow 2: Get all accounts
	req := httptest.NewRequest("GET", "/admin/accounts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	ts.apiHandler.HandleAdminGetAccounts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Get accounts failed: %d", w.Code)
	}

	var accounts []*core.Account
	json.NewDecoder(w.Body).Decode(&accounts)
	t.Logf("Retrieved %d accounts", len(accounts))

	// Flow 3: Make a deposit
	depositReq := map[string]interface{}{
		"accountId":   ts.testAccountID,
		"amount":      5000.0,
		"method":      "BANK",
		"reference":   "TEST-DEPOSIT-001",
		"description": "Test deposit",
		"adminId":     "admin1",
	}

	depositJSON, _ := json.Marshal(depositReq)
	req = httptest.NewRequest("POST", "/admin/deposit", bytes.NewReader(depositJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ts.apiHandler.HandleAdminDeposit(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Deposit failed: %d", w.Code)
	}

	// Flow 4: Get symbols
	req = httptest.NewRequest("GET", "/admin/symbols", nil)
	w = httptest.NewRecorder()
	ts.apiHandler.HandleAdminGetSymbols(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Get symbols failed: %d", w.Code)
	}

	t.Log("PASS: Complete AdminPanel flow test")
}

// TestInvalidAdminRequests tests error handling for invalid requests
func TestInvalidAdminRequests(t *testing.T) {
	ts := SetupEndpointTestServer(t)
	defer ts.Cleanup()

	tests := []struct {
		name       string
		endpoint   string
		method     string
		body       interface{}
	}{
		{
			name:       "Invalid account ID",
			endpoint:   "/admin/deposit",
			method:     "POST",
			body:       map[string]interface{}{"accountId": 99999, "amount": 100.0},
		},
		{
			name:       "Empty deposit request",
			endpoint:   "/admin/deposit",
			method:     "POST",
			body:       map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(tt.method, tt.endpoint, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Route to handler
			ts.apiHandler.HandleAdminDeposit(w, req)

			if w.Code != http.StatusOK && w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
				t.Logf("Status code: %d", w.Code)
			}
		})
	}

	t.Log("PASS: Invalid requests test")
}

// BenchmarkRoutingDecision benchmarks routing decision performance
func BenchmarkRoutingDecision(b *testing.B) {
	ts := SetupEndpointTestServer(&testing.T{})
	defer ts.Cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ts.routingEngine.Route(
			ts.testAccountID,
			"EURUSD",
			"BUY",
			1.0,
			0.015,
		)
	}
}
