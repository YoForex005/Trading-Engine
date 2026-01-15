package lpmanager

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLPManager_LoadConfig(t *testing.T) {
	t.Parallel()

	// Create temp directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)

	// First load should create default config
	err := manager.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify default config was created
	config := manager.GetConfig()
	if config == nil {
		t.Fatal("config is nil after LoadConfig")
	}

	if len(config.LPs) == 0 {
		t.Error("expected default LPs to be created")
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestLPManager_GetLPConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test O(1) lookup
	oandaConfig := manager.GetLPConfig("oanda")
	if oandaConfig == nil {
		// Debug: print what we have
		config := manager.GetConfig()
		t.Logf("Available LPs: %d", len(config.LPs))
		for _, lp := range config.LPs {
			t.Logf("  - %s", lp.ID)
		}
		t.Fatal("expected to find OANDA config")
	}

	if oandaConfig.ID != "oanda" {
		t.Errorf("got ID %s, want oanda", oandaConfig.ID)
	}

	// Test non-existent LP
	nonExistent := manager.GetLPConfig("nonexistent")
	if nonExistent != nil {
		t.Error("expected nil for non-existent LP")
	}
}

func TestLPManager_AddRemoveLP(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add new LP
	newLP := LPConfig{
		ID:       "test-lp",
		Name:     "Test LP",
		Type:     "TEST",
		Enabled:  false,
		Priority: 10,
		Settings: map[string]string{},
	}

	err := manager.AddLP(newLP)
	if err != nil {
		t.Fatalf("AddLP failed: %v", err)
	}

	// Verify LP was added
	config := manager.GetLPConfig("test-lp")
	if config == nil {
		t.Fatal("LP was not added")
	}
	if config.Name != "Test LP" {
		t.Errorf("got name %s, want Test LP", config.Name)
	}

	// Try to add duplicate - should fail
	err = manager.AddLP(newLP)
	if err != ErrLPAlreadyExists {
		t.Errorf("expected ErrLPAlreadyExists, got %v", err)
	}

	// Remove LP
	err = manager.RemoveLP("test-lp")
	if err != nil {
		t.Fatalf("RemoveLP failed: %v", err)
	}

	// Verify LP was removed
	config = manager.GetLPConfig("test-lp")
	if config != nil {
		t.Error("LP was not removed")
	}

	// Try to remove non-existent LP
	err = manager.RemoveLP("nonexistent")
	if err != ErrLPNotFound {
		t.Errorf("expected ErrLPNotFound, got %v", err)
	}
}

func TestLPManager_UpdateLP(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Update OANDA config
	updatedConfig := LPConfig{
		ID:       "oanda",
		Name:     "OANDA Updated",
		Type:     "OANDA",
		Enabled:  false,
		Priority: 5,
		Settings: map[string]string{"test": "value"},
	}

	err := manager.UpdateLP(updatedConfig)
	if err != nil {
		t.Fatalf("UpdateLP failed: %v", err)
	}

	// Verify update
	config := manager.GetLPConfig("oanda")
	if config.Name != "OANDA Updated" {
		t.Errorf("got name %s, want OANDA Updated", config.Name)
	}
	if config.Priority != 5 {
		t.Errorf("got priority %d, want 5", config.Priority)
	}

	// Try to update non-existent LP
	nonExistent := LPConfig{ID: "nonexistent"}
	err = manager.UpdateLP(nonExistent)
	if err != ErrLPNotFound {
		t.Errorf("expected ErrLPNotFound, got %v", err)
	}
}

func TestLPManager_ToggleLP(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Register a mock adapter
	mockAdapter := &MockAdapter{
		id:         "oanda",
		connected:  false,
		quotesChan: make(chan Quote, 10),
	}
	manager.RegisterAdapter(mockAdapter)

	// Get initial state
	initialConfig := manager.GetLPConfig("oanda")
	initialState := initialConfig.Enabled

	// Toggle LP
	err := manager.ToggleLP("oanda")
	if err != nil {
		t.Fatalf("ToggleLP failed: %v", err)
	}

	// Verify state changed
	newConfig := manager.GetLPConfig("oanda")
	if newConfig.Enabled == initialState {
		t.Error("LP enabled state did not change")
	}

	// Try to toggle non-existent LP
	err = manager.ToggleLP("nonexistent")
	if err != ErrLPNotFound {
		t.Errorf("expected ErrLPNotFound, got %v", err)
	}
}

func TestLPManager_GetEnabledAdapters(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Register mock adapters
	oandaAdapter := &MockAdapter{
		id:         "oanda",
		connected:  true,
		quotesChan: make(chan Quote, 10),
	}
	binanceAdapter := &MockAdapter{
		id:         "binance",
		connected:  true,
		quotesChan: make(chan Quote, 10),
	}

	manager.RegisterAdapter(oandaAdapter)
	manager.RegisterAdapter(binanceAdapter)

	// Get enabled adapters
	enabled := manager.GetEnabledAdapters()

	// Default config has both OANDA and Binance enabled
	if len(enabled) != 2 {
		t.Errorf("got %d enabled adapters, want 2", len(enabled))
	}

	// Disable one LP
	manager.SetLPEnabled("oanda", false)

	// Get enabled adapters again
	enabled = manager.GetEnabledAdapters()
	if len(enabled) != 1 {
		t.Errorf("got %d enabled adapters after disabling one, want 1", len(enabled))
	}
}

func TestLPManager_GetStatus(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Register mock adapter
	mockAdapter := &MockAdapter{
		id:         "oanda",
		name:       "OANDA",
		lpType:     "OANDA",
		connected:  true,
		quotesChan: make(chan Quote, 10),
	}
	manager.RegisterAdapter(mockAdapter)

	// Get status
	status := manager.GetStatus()

	if len(status) == 0 {
		t.Fatal("expected status for configured LPs")
	}

	oandaStatus, exists := status["oanda"]
	if !exists {
		t.Fatal("expected OANDA status")
	}

	if oandaStatus.Name != "OANDA" {
		t.Errorf("got name %s, want OANDA", oandaStatus.Name)
	}

	if !oandaStatus.Connected {
		t.Error("expected adapter to be connected")
	}
}

func TestLPManager_QuoteAggregation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Create mock adapters
	oandaAdapter := &MockAdapter{
		id:         "oanda",
		name:       "OANDA",
		lpType:     "OANDA",
		connected:  true,
		quotesChan: make(chan Quote, 10),
	}

	binanceAdapter := &MockAdapter{
		id:         "binance",
		name:       "Binance",
		lpType:     "BINANCE",
		connected:  true,
		quotesChan: make(chan Quote, 10),
	}

	manager.RegisterAdapter(oandaAdapter)
	manager.RegisterAdapter(binanceAdapter)

	// Start OANDA adapter
	err := manager.startLP("oanda")
	if err != nil {
		t.Fatalf("failed to start OANDA: %v", err)
	}

	// Send a test quote through OANDA adapter
	testQuote := Quote{
		Symbol:    "EURUSD",
		Bid:       1.0850,
		Ask:       1.0852,
		Timestamp: time.Now().Unix(),
		LP:        "oanda",
	}

	go func() {
		oandaAdapter.quotesChan <- testQuote
	}()

	// Receive quote from manager's aggregated channel
	select {
	case receivedQuote := <-manager.GetQuotesChan():
		if receivedQuote.Symbol != "EURUSD" {
			t.Errorf("got symbol %s, want EURUSD", receivedQuote.Symbol)
		}
		if receivedQuote.Bid != 1.0850 {
			t.Errorf("got bid %.4f, want 1.0850", receivedQuote.Bid)
		}
		if receivedQuote.LP != "oanda" {
			t.Errorf("got LP %s, want oanda", receivedQuote.LP)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for aggregated quote")
	}

	// Stop LP
	manager.stopLP("oanda")
}

// MockAdapter implements LPAdapter for testing
type MockAdapter struct {
	id         string
	name       string
	lpType     string
	connected  bool
	quotesChan chan Quote
}

func (m *MockAdapter) ID() string                        { return m.id }
func (m *MockAdapter) Name() string                      { return m.name }
func (m *MockAdapter) Type() string                      { return m.lpType }
func (m *MockAdapter) IsConnected() bool                 { return m.connected }
func (m *MockAdapter) GetQuotesChan() <-chan Quote       { return m.quotesChan }
func (m *MockAdapter) Connect() error                    { m.connected = true; return nil }
func (m *MockAdapter) Disconnect() error                 { m.connected = false; return nil }
func (m *MockAdapter) GetSymbols() ([]SymbolInfo, error) { return []SymbolInfo{}, nil }
func (m *MockAdapter) Subscribe(symbols []string) error  { return nil }
func (m *MockAdapter) Unsubscribe(symbols []string) error { return nil }
func (m *MockAdapter) GetStatus() LPStatus {
	return LPStatus{
		ID:        m.id,
		Name:      m.name,
		Type:      m.lpType,
		Connected: m.connected,
	}
}
