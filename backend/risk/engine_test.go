package risk

import "testing"

// TestEngineAdminIntegration verifies that Engine has all methods required by AdminController
func TestEngineAdminIntegration(t *testing.T) {
	engine := NewEngine()
	admin := NewAdminController(engine)

	// Test SetClientRiskProfile
	profile := &ClientRiskProfile{
		ClientID:        "test_client",
		MaxLeverage:     100,
		StopOutLevel:    30,
		MarginCallLevel: 80,
	}
	err := admin.SetClientRiskProfile(profile)
	if err != nil {
		t.Errorf("SetClientRiskProfile failed: %v", err)
	}

	// Test SetInstrumentRiskParams
	params := &InstrumentRiskParams{
		Symbol:      "EURUSD",
		MaxLeverage: 30,
	}
	err = admin.SetInstrumentRiskParams(params)
	if err != nil {
		t.Errorf("SetInstrumentRiskParams failed: %v", err)
	}

	// Verify methods exist and can be called
	_ = engine.GetPosition(1)
	bid, ask := engine.GetCurrentPrice("EURUSD")
	_, _ = bid, ask
	_ = engine.GetAllPositions(1)
	_ = engine.GetClientRiskProfile("test_client")
	_ = engine.GetDailyPnL(1)
	_ = engine.GetAllAccounts()
	_ = engine.GetTodayMarginCallCount()
	_ = engine.GetTodayLiquidationCount()

	alert := &RiskAlert{
		ID:        "test_alert",
		AlertType: "TEST",
		Severity:  RiskLevelLow,
		Message:   "Test alert",
	}
	_ = engine.StoreAlert(alert)

	t.Log("All required methods are implemented")
}
