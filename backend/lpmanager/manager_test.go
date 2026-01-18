package lpmanager

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// MockLPAdapter is a mock implementation for testing
type MockLPAdapter struct {
	id         string
	name       string
	lpType     string
	connected  bool
	quotesChan chan Quote
	symbols    []SymbolInfo
	mu         sync.RWMutex
}

func NewMockLPAdapter(id, name, lpType string) *MockLPAdapter {
	return &MockLPAdapter{
		id:         id,
		name:       name,
		lpType:     lpType,
		connected:  false,
		quotesChan: make(chan Quote, 100),
		symbols: []SymbolInfo{
			{Symbol: "EURUSD", DisplayName: "EUR/USD", Type: "forex"},
			{Symbol: "GBPUSD", DisplayName: "GBP/USD", Type: "forex"},
		},
	}
}

func (m *MockLPAdapter) ID() string   { return m.id }
func (m *MockLPAdapter) Name() string { return m.name }
func (m *MockLPAdapter) Type() string { return m.lpType }

func (m *MockLPAdapter) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = true
	return nil
}

func (m *MockLPAdapter) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	close(m.quotesChan)
	return nil
}

func (m *MockLPAdapter) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

func (m *MockLPAdapter) GetSymbols() ([]SymbolInfo, error) {
	return m.symbols, nil
}

func (m *MockLPAdapter) Subscribe(symbols []string) error {
	return nil
}

func (m *MockLPAdapter) Unsubscribe(symbols []string) error {
	return nil
}

func (m *MockLPAdapter) GetQuotesChan() <-chan Quote {
	return m.quotesChan
}

func (m *MockLPAdapter) GetStatus() LPStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return LPStatus{
		ID:          m.id,
		Name:        m.name,
		Type:        m.lpType,
		Connected:   m.connected,
		SymbolCount: len(m.symbols),
	}
}

// TestNewManager tests manager initialization
func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.registry == nil {
		t.Error("registry not initialized")
	}

	if manager.configPath != configPath {
		t.Errorf("configPath = %s, want %s", manager.configPath, configPath)
	}

	if manager.quotesChan == nil {
		t.Error("quotesChan not initialized")
	}

	if manager.activeAggregators == nil {
		t.Error("activeAggregators not initialized")
	}
}

// TestLoadConfig tests configuration loading
func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)

	// First load should create default config
	err := manager.LoadConfig()

	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	config := manager.GetConfig()

	if config == nil {
		t.Fatal("Config should not be nil after load")
	}

	if len(config.LPs) == 0 {
		t.Error("Default config should have LPs")
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should be created")
	}
}

// TestSaveConfig tests configuration saving
func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")

	manager := NewManager(configPath)
	manager.LoadConfig()

	// Modify config
	config := manager.GetConfig()
	config.PrimaryLP = "modified_lp"

	err := manager.SaveConfig()

	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Create new manager and load
	manager2 := NewManager(configPath)
	err = manager2.LoadConfig()

	if err != nil {
		t.Fatalf("Second LoadConfig() error = %v", err)
	}

	config2 := manager2.GetConfig()

	if config2.PrimaryLP != "modified_lp" {
		t.Errorf("PrimaryLP = %s, want modified_lp", config2.PrimaryLP)
	}
}

// TestAddLP tests adding LP configuration
func TestAddLP(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	newLP := LPConfig{
		ID:       "test_lp",
		Name:     "Test LP",
		Type:     "TEST",
		Enabled:  true,
		Priority: 10,
		Settings: map[string]string{"key": "value"},
	}

	err := manager.AddLP(newLP)

	if err != nil {
		t.Fatalf("AddLP() error = %v", err)
	}

	config := manager.GetConfig()
	found := false

	for _, lp := range config.LPs {
		if lp.ID == "test_lp" {
			found = true
			if lp.Name != "Test LP" {
				t.Errorf("Name = %s, want Test LP", lp.Name)
			}
			break
		}
	}

	if !found {
		t.Error("Added LP not found in config")
	}
}

// TestAddDuplicateLP tests adding duplicate LP
func TestAddDuplicateLP(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	lp := LPConfig{
		ID:   "oanda",
		Name: "OANDA",
		Type: "OANDA",
	}

	err := manager.AddLP(lp)

	if err == nil {
		t.Error("Expected error for duplicate LP, got nil")
	}

	if err != ErrLPAlreadyExists {
		t.Errorf("Error = %v, want ErrLPAlreadyExists", err)
	}
}

// TestUpdateLP tests updating LP configuration
func TestUpdateLP(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	updated := LPConfig{
		ID:       "oanda",
		Name:     "OANDA Updated",
		Type:     "OANDA",
		Enabled:  false,
		Priority: 99,
	}

	err := manager.UpdateLP(updated)

	if err != nil {
		t.Fatalf("UpdateLP() error = %v", err)
	}

	lp := manager.GetLPConfig("oanda")

	if lp == nil {
		t.Fatal("Updated LP not found")
	}

	if lp.Name != "OANDA Updated" {
		t.Errorf("Name = %s, want OANDA Updated", lp.Name)
	}

	if lp.Priority != 99 {
		t.Errorf("Priority = %d, want 99", lp.Priority)
	}
}

// TestUpdateNonExistentLP tests updating non-existent LP
func TestUpdateNonExistentLP(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	updated := LPConfig{
		ID:   "nonexistent",
		Name: "Non-existent",
		Type: "TEST",
	}

	err := manager.UpdateLP(updated)

	if err == nil {
		t.Error("Expected error for non-existent LP, got nil")
	}

	if err != ErrLPNotFound {
		t.Errorf("Error = %v, want ErrLPNotFound", err)
	}
}

// TestRemoveLP tests removing LP configuration
func TestRemoveLP(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	err := manager.RemoveLP("binance")

	if err != nil {
		t.Fatalf("RemoveLP() error = %v", err)
	}

	lp := manager.GetLPConfig("binance")

	if lp != nil {
		t.Error("LP should be removed")
	}

	// Try to remove again
	err = manager.RemoveLP("binance")

	if err == nil {
		t.Error("Expected error for removing non-existent LP, got nil")
	}
}

// TestGetLPConfig tests retrieving LP configuration
func TestGetLPConfig(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	lp := manager.GetLPConfig("oanda")

	if lp == nil {
		t.Fatal("GetLPConfig() returned nil for existing LP")
	}

	if lp.ID != "oanda" {
		t.Errorf("ID = %s, want oanda", lp.ID)
	}

	// Test non-existent LP
	nonExistent := manager.GetLPConfig("nonexistent")

	if nonExistent != nil {
		t.Error("Non-existent LP should return nil")
	}
}

// TestRegisterAdapter tests adapter registration
func TestRegisterAdapter(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))

	adapter := NewMockLPAdapter("test_lp", "Test LP", "TEST")

	err := manager.RegisterAdapter(adapter)

	if err != nil {
		t.Fatalf("RegisterAdapter() error = %v", err)
	}

	retrieved, ok := manager.GetAdapter("test_lp")

	if !ok {
		t.Fatal("GetAdapter() should return true for registered adapter")
	}

	if retrieved.ID() != "test_lp" {
		t.Errorf("Adapter ID = %s, want test_lp", retrieved.ID())
	}
}

// TestGetEnabledAdapters tests retrieving enabled adapters
func TestGetEnabledAdapters(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	// Register adapters
	adapter1 := NewMockLPAdapter("oanda", "OANDA", "OANDA")
	adapter2 := NewMockLPAdapter("binance", "Binance", "BINANCE")

	manager.RegisterAdapter(adapter1)
	manager.RegisterAdapter(adapter2)

	// Both should be enabled by default
	enabled := manager.GetEnabledAdapters()

	if len(enabled) != 2 {
		t.Errorf("Enabled adapters count = %d, want 2", len(enabled))
	}

	// Disable one
	manager.SetLPEnabled("binance", false)

	enabled = manager.GetEnabledAdapters()

	if len(enabled) != 1 {
		t.Errorf("Enabled adapters count after disable = %d, want 1", len(enabled))
	}

	if enabled[0].ID() != "oanda" {
		t.Errorf("Enabled adapter = %s, want oanda", enabled[0].ID())
	}
}

// TestGetStatus tests status retrieval
func TestGetStatus(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	adapter := NewMockLPAdapter("oanda", "OANDA", "OANDA")
	manager.RegisterAdapter(adapter)

	status := manager.GetStatus()

	if len(status) == 0 {
		t.Error("Status map should not be empty")
	}

	oandaStatus, ok := status["oanda"]

	if !ok {
		t.Fatal("OANDA status not found")
	}

	if oandaStatus.ID != "oanda" {
		t.Errorf("Status ID = %s, want oanda", oandaStatus.ID)
	}

	if oandaStatus.Name != "OANDA" {
		t.Errorf("Status Name = %s, want OANDA", oandaStatus.Name)
	}
}

// TestToggleLP tests enabling/disabling LP
func TestToggleLP(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	adapter := NewMockLPAdapter("oanda", "OANDA", "OANDA")
	manager.RegisterAdapter(adapter)

	// Initially enabled, toggle to disabled
	err := manager.ToggleLP("oanda")

	if err != nil {
		t.Fatalf("ToggleLP() error = %v", err)
	}

	lp := manager.GetLPConfig("oanda")

	if lp.Enabled {
		t.Error("LP should be disabled after toggle")
	}

	// Toggle back to enabled
	err = manager.ToggleLP("oanda")

	if err != nil {
		t.Fatalf("Second ToggleLP() error = %v", err)
	}

	lp = manager.GetLPConfig("oanda")

	if !lp.Enabled {
		t.Error("LP should be enabled after second toggle")
	}
}

// TestSetLPEnabled tests setting LP enabled state
func TestSetLPEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	adapter := NewMockLPAdapter("oanda", "OANDA", "OANDA")
	manager.RegisterAdapter(adapter)

	// Disable
	err := manager.SetLPEnabled("oanda", false)

	if err != nil {
		t.Fatalf("SetLPEnabled(false) error = %v", err)
	}

	lp := manager.GetLPConfig("oanda")

	if lp.Enabled {
		t.Error("LP should be disabled")
	}

	// Enable
	err = manager.SetLPEnabled("oanda", true)

	if err != nil {
		t.Fatalf("SetLPEnabled(true) error = %v", err)
	}

	lp = manager.GetLPConfig("oanda")

	if !lp.Enabled {
		t.Error("LP should be enabled")
	}
}

// TestGetQuotesChan tests quotes channel retrieval
func TestGetQuotesChan(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))

	quotesChan := manager.GetQuotesChan()

	if quotesChan == nil {
		t.Fatal("GetQuotesChan() returned nil")
	}

	// Verify it's a receive-only channel
	select {
	case <-quotesChan:
		// OK, channel is readable
	default:
		// OK, no quotes yet
	}
}

// TestConcurrentConfigOperations tests thread-safety
func TestConcurrentConfigOperations(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
	manager.LoadConfig()

	var wg sync.WaitGroup

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.GetConfig()
			manager.GetLPConfig("oanda")
			manager.GetStatus()
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			lp := LPConfig{
				ID:       "test_" + string(rune('a'+idx%26)),
				Name:     "Test",
				Type:     "TEST",
				Enabled:  true,
				Priority: idx,
			}

			manager.AddLP(lp)
		}(i)
	}

	wg.Wait()

	config := manager.GetConfig()

	if len(config.LPs) < 2 {
		t.Error("Should have at least 2 LPs after concurrent operations")
	}
}

// TestConcurrentAdapterOperations tests concurrent adapter access
func TestConcurrentAdapterOperations(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))

	// Register adapters concurrently
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			id := "adapter_" + string(rune('a'+idx%26))
			adapter := NewMockLPAdapter(id, "Adapter", "TEST")
			manager.RegisterAdapter(adapter)
		}(i)
	}

	wg.Wait()

	// Verify adapters were registered
	for i := 0; i < 20; i++ {
		id := "adapter_" + string(rune('a'+i%26))
		_, ok := manager.GetAdapter(id)
		if !ok {
			t.Errorf("Adapter %s not found", id)
		}
	}
}

// TestLPError tests LP error types
func TestLPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"LP not found", ErrLPNotFound, "LP not found"},
		{"LP already exists", ErrLPAlreadyExists, "LP already exists"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error message = %s, want %s", tt.err.Error(), tt.expected)
			}
		})
	}
}

// TestDefaultConfig tests default configuration
func TestDefaultConfig(t *testing.T) {
	config := NewDefaultConfig()

	if config == nil {
		t.Fatal("NewDefaultConfig() returned nil")
	}

	if len(config.LPs) == 0 {
		t.Error("Default config should have LPs")
	}

	if config.PrimaryLP == "" {
		t.Error("Default config should have primary LP set")
	}

	// Verify default LPs
	hasOanda := false
	hasBinance := false

	for _, lp := range config.LPs {
		if lp.ID == "oanda" {
			hasOanda = true
		}
		if lp.ID == "binance" {
			hasBinance = true
		}
	}

	if !hasOanda {
		t.Error("Default config should include OANDA")
	}

	if !hasBinance {
		t.Error("Default config should include Binance")
	}
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	t.Run("Load config from non-existent directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "subdir", "that", "does", "not", "exist", "config.json")

		manager := NewManager(configPath)
		err := manager.LoadConfig()

		// Should create default config
		if err != nil {
			t.Errorf("LoadConfig() should create default config, got error: %v", err)
		}

		// Directory should be created when saving
		err = manager.SaveConfig()

		if err != nil {
			t.Errorf("SaveConfig() should create directory, got error: %v", err)
		}
	})

	t.Run("Toggle non-existent LP", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
		manager.LoadConfig()

		err := manager.ToggleLP("nonexistent")

		if err == nil {
			t.Error("Expected error for non-existent LP, got nil")
		}
	})

	t.Run("SetLPEnabled for non-existent LP", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))
		manager.LoadConfig()

		err := manager.SetLPEnabled("nonexistent", true)

		if err == nil {
			t.Error("Expected error for non-existent LP, got nil")
		}
	})

	t.Run("GetAdapter for non-existent adapter", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))

		_, ok := manager.GetAdapter("nonexistent")

		if ok {
			t.Error("GetAdapter() should return false for non-existent adapter")
		}
	})

	t.Run("Save config without loading first", func(t *testing.T) {
		tmpDir := t.TempDir()
		manager := NewManager(filepath.Join(tmpDir, "lp_config.json"))

		// Should handle nil config gracefully
		err := manager.SaveConfig()

		if err != nil {
			t.Errorf("SaveConfig() with nil config error = %v", err)
		}
	})
}

// TestQuoteStructure tests Quote struct
func TestQuoteStructure(t *testing.T) {
	quote := Quote{
		Symbol:    "EURUSD",
		Bid:       1.1000,
		Ask:       1.1002,
		Timestamp: time.Now().Unix(),
		LP:        "oanda",
	}

	if quote.Symbol != "EURUSD" {
		t.Errorf("Symbol = %s, want EURUSD", quote.Symbol)
	}

	if quote.Bid >= quote.Ask {
		t.Error("Bid should be less than Ask")
	}

	if quote.LP == "" {
		t.Error("LP should not be empty")
	}
}

// TestSymbolInfoStructure tests SymbolInfo struct
func TestSymbolInfoStructure(t *testing.T) {
	symbol := SymbolInfo{
		Symbol:        "EURUSD",
		DisplayName:   "EUR/USD",
		BaseCurrency:  "EUR",
		QuoteCurrency: "USD",
		MinLotSize:    0.01,
		MaxLotSize:    100.0,
		LotStep:       0.01,
		PipValue:      10.0,
		Type:          "forex",
	}

	if symbol.Symbol != "EURUSD" {
		t.Errorf("Symbol = %s, want EURUSD", symbol.Symbol)
	}

	if symbol.Type != "forex" {
		t.Errorf("Type = %s, want forex", symbol.Type)
	}

	if symbol.MinLotSize >= symbol.MaxLotSize {
		t.Error("MinLotSize should be less than MaxLotSize")
	}
}
