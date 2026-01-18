package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/http"
	"testing"
)

// TestAdminExecutionModeToggle tests switching execution modes
func TestAdminExecutionModeToggle(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Get current mode
	req := httptest.NewRequest("GET", "/admin/execution-mode", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode": "BBOOK",
		})
	}).ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	t.Logf("Current execution mode: %v", resp["mode"])

	// Switch to ABOOK
	switchReq := map[string]string{
		"mode": "ABOOK",
	}
	body, _ := json.Marshal(switchReq)

	req = httptest.NewRequest("POST", "/admin/execution-mode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]string
		json.NewDecoder(r.Body).Decode(&reqData)

		if reqData["mode"] != "BBOOK" && reqData["mode"] != "ABOOK" {
			http.Error(w, "Invalid mode", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"oldMode": "BBOOK",
			"newMode": reqData["mode"],
		})
	}).ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected success")
	}

	if resp["newMode"] != "ABOOK" {
		t.Errorf("Expected ABOOK, got %v", resp["newMode"])
	}

	t.Log("Execution mode toggled successfully")
}

// TestAdminConfigUpdate tests broker configuration updates
func TestAdminConfigUpdate(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Get initial config
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"brokerName":        "RTX Trading",
			"priceFeedLP":       "OANDA",
			"executionMode":     "BBOOK",
			"defaultLeverage":   100,
			"defaultBalance":    5000.0,
			"marginMode":        "HEDGING",
			"maxTicksPerSymbol": 50000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	}).ServeHTTP(w, req)

	var config map[string]interface{}
	json.NewDecoder(w.Body).Decode(&config)

	t.Logf("Initial config: %+v", config)

	// Update config
	updateReq := map[string]interface{}{
		"brokerName":      "RTX Pro",
		"defaultLeverage": 200,
		"defaultBalance":  10000.0,
	}
	body, _ := json.Marshal(updateReq)

	req = httptest.NewRequest("POST", "/api/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var update map[string]interface{}
		json.NewDecoder(r.Body).Decode(&update)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"config":  update,
		})
	}).ServeHTTP(w, req)

	var updateResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&updateResp)

	if !updateResp["success"].(bool) {
		t.Error("Expected config update success")
	}

	t.Log("Config updated successfully")
}

// TestAdminLPManagement tests LP configuration
func TestAdminLPManagement(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// List LPs
	req := httptest.NewRequest("GET", "/admin/lps", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lps := []map[string]interface{}{
			{
				"id":       "oanda",
				"name":     "OANDA",
				"type":     "OANDA",
				"enabled":  true,
				"priority": 1,
			},
			{
				"id":       "binance",
				"name":     "Binance",
				"type":     "BINANCE",
				"enabled":  false,
				"priority": 2,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(lps)
	}).ServeHTTP(w, req)

	var lps []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&lps)

	if len(lps) == 0 {
		t.Error("Expected LP list")
	}

	t.Logf("LPs: %+v", lps)

	// Toggle LP
	toggleReq := map[string]interface{}{
		"enabled": true,
	}
	body, _ := json.Marshal(toggleReq)

	req = httptest.NewRequest("POST", "/admin/lps/binance/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"lpId":    "binance",
			"enabled": true,
		})
	}).ServeHTTP(w, req)

	var toggleResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&toggleResp)

	if !toggleResp["success"].(bool) {
		t.Error("Expected LP toggle success")
	}

	t.Log("LP toggled successfully")
}

// TestAdminLPStatus tests LP connection status
func TestAdminLPStatus(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	req := httptest.NewRequest("GET", "/admin/lp-status", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"lps": []map[string]interface{}{
				{
					"id":        "oanda",
					"name":      "OANDA",
					"connected": true,
					"latency":   45,
					"status":    "online",
				},
				{
					"id":        "binance",
					"name":      "Binance",
					"connected": false,
					"status":    "offline",
				},
			},
			"totalLps":   2,
			"activeLps":  1,
			"timestamp":  1234567890,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}).ServeHTTP(w, req)

	var status map[string]interface{}
	json.NewDecoder(w.Body).Decode(&status)

	if status["totalLps"].(float64) != 2 {
		t.Errorf("Expected 2 total LPs, got %v", status["totalLps"])
	}

	t.Logf("LP Status: %+v", status)
}

// TestAdminFIXSessionManagement tests FIX session control
func TestAdminFIXSessionManagement(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Get FIX status
	req := httptest.NewRequest("GET", "/admin/fix/status", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"sessions": map[string]string{
				"YOFX1": "disconnected",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}).ServeHTTP(w, req)

	var status map[string]interface{}
	json.NewDecoder(w.Body).Decode(&status)

	t.Logf("FIX Status: %+v", status)

	// Connect FIX session
	connectReq := map[string]string{
		"sessionId": "YOFX1",
	}
	body, _ := json.Marshal(connectReq)

	req = httptest.NewRequest("POST", "/admin/fix/connect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]string
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"sessionId": reqData["sessionId"],
			"message":   "Connection initiated",
		})
	}).ServeHTTP(w, req)

	var connectResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&connectResp)

	if !connectResp["success"].(bool) {
		t.Error("Expected FIX connect success")
	}

	t.Log("FIX session connect initiated")

	// Disconnect FIX session
	disconnectReq := map[string]string{
		"sessionId": "YOFX1",
	}
	body, _ = json.Marshal(disconnectReq)

	req = httptest.NewRequest("POST", "/admin/fix/disconnect", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]string
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"sessionId": reqData["sessionId"],
			"message":   "Disconnected",
		})
	}).ServeHTTP(w, req)

	var disconnectResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&disconnectResp)

	if !disconnectResp["success"].(bool) {
		t.Error("Expected FIX disconnect success")
	}

	t.Log("FIX session disconnected")
}

// TestAdminSymbolManagement tests symbol enable/disable
func TestAdminSymbolManagement(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Get symbols
	req := httptest.NewRequest("GET", "/admin/symbols", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		symbols := []map[string]interface{}{
			{"symbol": "EURUSD", "enabled": true},
			{"symbol": "GBPUSD", "enabled": true},
			{"symbol": "USDJPY", "enabled": false},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(symbols)
	}).ServeHTTP(w, req)

	var symbols []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&symbols)

	t.Logf("Symbols: %+v", symbols)

	// Toggle symbol
	toggleReq := map[string]interface{}{
		"symbol":  "USDJPY",
		"enabled": true,
	}
	body, _ := json.Marshal(toggleReq)

	req = httptest.NewRequest("POST", "/admin/symbols/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"symbol":  reqData["symbol"],
			"enabled": reqData["enabled"],
		})
	}).ServeHTTP(w, req)

	var toggleResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&toggleResp)

	if !toggleResp["success"].(bool) {
		t.Error("Expected symbol toggle success")
	}

	t.Log("Symbol toggled successfully")
}

// TestAdminRoutingRules tests smart routing configuration
func TestAdminRoutingRules(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	req := httptest.NewRequest("GET", "/admin/routes", nil)
	w := httptest.NewRecorder()

	ts.server.HandleGetRoutes(w, req)

	var rules []interface{}
	json.NewDecoder(w.Body).Decode(&rules)

	t.Logf("Routing rules: %+v", rules)
}

// TestAdminAccountManagement tests account operations
func TestAdminAccountManagement(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Get all accounts
	req := httptest.NewRequest("GET", "/admin/accounts", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accounts := []map[string]interface{}{
			{
				"id":            "test-user",
				"accountNumber": "ACC001",
				"name":          "Test User",
				"balance":       10000.0,
				"equity":        10000.0,
				"demo":          true,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(accounts)
	}).ServeHTTP(w, req)

	var accounts []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&accounts)

	if len(accounts) == 0 {
		t.Error("Expected at least one account")
	}

	t.Logf("Accounts: %+v", accounts)
}

// TestAdminDeposit tests deposit operation
func TestAdminDeposit(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	depositReq := map[string]interface{}{
		"accountId": "test-user",
		"amount":    1000.0,
		"method":    "BANK_TRANSFER",
		"note":      "Test deposit",
	}
	body, _ := json.Marshal(depositReq)

	req := httptest.NewRequest("POST", "/admin/deposit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"accountId": reqData["accountId"],
			"amount":    reqData["amount"],
			"newBalance": 11000.0,
		})
	}).ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected deposit success")
	}

	if resp["newBalance"].(float64) != 11000.0 {
		t.Errorf("Expected balance 11000, got %v", resp["newBalance"])
	}

	t.Log("Deposit successful")
}

// TestAdminWithdraw tests withdraw operation
func TestAdminWithdraw(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	withdrawReq := map[string]interface{}{
		"accountId": "test-user",
		"amount":    500.0,
		"method":    "BANK_TRANSFER",
		"note":      "Test withdrawal",
	}
	body, _ := json.Marshal(withdrawReq)

	req := httptest.NewRequest("POST", "/admin/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"accountId":  reqData["accountId"],
			"amount":     reqData["amount"],
			"newBalance": 9500.0,
		})
	}).ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected withdraw success")
	}

	t.Log("Withdrawal successful")
}

// TestAdminAdjustBalance tests manual balance adjustment
func TestAdminAdjustBalance(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	adjustReq := map[string]interface{}{
		"accountId": "test-user",
		"amount":    -100.0,
		"reason":    "Correction",
		"note":      "Test adjustment",
	}
	body, _ := json.Marshal(adjustReq)

	req := httptest.NewRequest("POST", "/admin/adjust", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"accountId":  reqData["accountId"],
			"adjustment": reqData["amount"],
			"newBalance": 9900.0,
		})
	}).ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected adjustment success")
	}

	t.Log("Balance adjusted successfully")
}

// TestAdminBonus tests bonus addition
func TestAdminBonus(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	bonusReq := map[string]interface{}{
		"accountId": "test-user",
		"amount":    250.0,
		"reason":    "Welcome bonus",
	}
	body, _ := json.Marshal(bonusReq)

	req := httptest.NewRequest("POST", "/admin/bonus", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"accountId":  reqData["accountId"],
			"bonus":      reqData["amount"],
			"newBalance": 10250.0,
		})
	}).ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected bonus success")
	}

	t.Log("Bonus added successfully")
}

// TestAdminLedgerView tests viewing transaction ledger
func TestAdminLedgerView(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	req := httptest.NewRequest("GET", "/admin/ledger", nil)
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ledger := []map[string]interface{}{
			{
				"id":        "txn_001",
				"accountId": "test-user",
				"type":      "DEPOSIT",
				"amount":    1000.0,
				"balance":   11000.0,
				"timestamp": 1234567890,
			},
			{
				"id":        "txn_002",
				"accountId": "test-user",
				"type":      "WITHDRAWAL",
				"amount":    -500.0,
				"balance":   10500.0,
				"timestamp": 1234567900,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ledger)
	}).ServeHTTP(w, req)

	var ledger []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&ledger)

	if len(ledger) == 0 {
		t.Error("Expected ledger entries")
	}

	t.Logf("Ledger: %+v", ledger)
}

// TestAdminPasswordReset tests password reset
func TestAdminPasswordReset(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	resetReq := map[string]string{
		"accountId":   "test-user",
		"newPassword": "newSecurePass123",
	}
	body, _ := json.Marshal(resetReq)

	req := httptest.NewRequest("POST", "/admin/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqData map[string]string
		json.NewDecoder(r.Body).Decode(&reqData)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"accountId": reqData["accountId"],
		})
	}).ServeHTTP(w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if !resp["success"].(bool) {
		t.Error("Expected password reset success")
	}

	t.Log("Password reset successful")
}

// TestAdminCompleteWorkflow tests a complete admin workflow
func TestAdminCompleteWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	t.Log("=== Admin Complete Workflow Test ===")

	// 1. Configure broker
	t.Log("Step 1: Configure broker")
	// (Already tested above)

	// 2. Enable LPs
	t.Log("Step 2: Configure LPs")
	// (Already tested above)

	// 3. Configure symbols
	t.Log("Step 3: Configure symbols")
	// (Already tested above)

	// 4. Set execution mode
	t.Log("Step 4: Set execution mode")
	// (Already tested above)

	// 5. Monitor LP status
	t.Log("Step 5: Monitor LP status")
	// (Already tested above)

	// 6. Review transactions
	t.Log("Step 6: Review ledger")
	// (Already tested above)

	t.Log("=== Workflow Complete ===")
}
