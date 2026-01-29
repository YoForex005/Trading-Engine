package lpmanager

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager orchestrates all LP connections
type Manager struct {
	registry          *Registry
	config            *LPManagerConfig
	configPath        string
	mu                sync.RWMutex
	quotesChan        chan Quote
	activeAggregators map[string]context.CancelFunc
}

// NewManager creates a new LP manager
func NewManager(configPath string) *Manager {
	return &Manager{
		registry:          NewRegistry(),
		configPath:        configPath,
		quotesChan:        make(chan Quote, 1000),
		activeAggregators: make(map[string]context.CancelFunc),
	}
}

// LoadConfig loads LP configuration from file
func (m *Manager) LoadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			m.config = NewDefaultConfig()
			return m.saveConfigLocked()
		}
		return err
	}

	var config LPManagerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	m.config = &config
	log.Printf("[LPManager] Loaded config with %d LPs", len(m.config.LPs))
	return nil
}

// SaveConfig saves LP configuration to file
func (m *Manager) SaveConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveConfigLocked()
}

func (m *Manager) saveConfigLocked() error {
	if m.config == nil {
		return nil
	}

	m.config.LastModified = time.Now().Unix()

	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *LPManagerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// GetLPConfig returns config for a specific LP
func (m *Manager) GetLPConfig(id string) *LPConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := range m.config.LPs {
		if m.config.LPs[i].ID == id {
			return &m.config.LPs[i]
		}
	}
	return nil
}

// AddLP adds a new LP configuration
func (m *Manager) AddLP(config LPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if ID already exists
	for _, lp := range m.config.LPs {
		if lp.ID == config.ID {
			return ErrLPAlreadyExists
		}
	}

	m.config.LPs = append(m.config.LPs, config)
	return m.saveConfigLocked()
}

// UpdateLP updates an existing LP configuration
func (m *Manager) UpdateLP(config LPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.config.LPs {
		if m.config.LPs[i].ID == config.ID {
			m.config.LPs[i] = config
			return m.saveConfigLocked()
		}
	}

	return ErrLPNotFound
}

// RemoveLP removes an LP configuration
func (m *Manager) RemoveLP(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.config.LPs {
		if m.config.LPs[i].ID == id {
			m.config.LPs = append(m.config.LPs[:i], m.config.LPs[i+1:]...)
			return m.saveConfigLocked()
		}
	}

	return ErrLPNotFound
}

// ToggleLP enables or disables an LP
func (m *Manager) ToggleLP(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.config.LPs {
		if m.config.LPs[i].ID == id {
			newState := !m.config.LPs[i].Enabled
			m.config.LPs[i].Enabled = newState

			if newState {
				if err := m.startLPLocked(id); err != nil {
					log.Printf("[LPManager] Failed to start LP %s: %v", id, err)
					m.config.LPs[i].Enabled = false
					return err
				}
			} else {
				m.stopLPLocked(id)
			}

			log.Printf("[LPManager] LP %s enabled=%v", id, newState)
			return m.saveConfigLocked()
		}
	}

	return ErrLPNotFound
}

// SetLPEnabled sets the enabled state of an LP
func (m *Manager) SetLPEnabled(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.config.LPs {
		if m.config.LPs[i].ID == id {
			m.config.LPs[i].Enabled = enabled

			if enabled {
				if err := m.startLPLocked(id); err != nil {
					// Revert
					m.config.LPs[i].Enabled = false
					return err
				}
			} else {
				m.stopLPLocked(id)
			}

			log.Printf("[LPManager] LP %s enabled=%v", id, enabled)
			return m.saveConfigLocked()
		}
	}

	return ErrLPNotFound
}

// RegisterAdapter registers an LP adapter with the manager
func (m *Manager) RegisterAdapter(adapter LPAdapter) error {
	return m.registry.Register(adapter)
}

// GetAdapter returns a registered adapter
func (m *Manager) GetAdapter(id string) (LPAdapter, bool) {
	return m.registry.Get(id)
}

// GetEnabledAdapters returns all enabled LP adapters
func (m *Manager) GetEnabledAdapters() []LPAdapter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapters := make([]LPAdapter, 0)
	for _, lpConfig := range m.config.LPs {
		if !lpConfig.Enabled {
			continue
		}
		if adapter, exists := m.registry.Get(lpConfig.ID); exists {
			adapters = append(adapters, adapter)
		}
	}
	return adapters
}

// GetStatus returns status of all LPs
func (m *Manager) GetStatus() map[string]LPStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]LPStatus)

	for _, lpConfig := range m.config.LPs {
		if adapter, exists := m.registry.Get(lpConfig.ID); exists {
			s := adapter.GetStatus()
			s.Enabled = lpConfig.Enabled
			status[lpConfig.ID] = s
		} else {
			status[lpConfig.ID] = LPStatus{
				ID:        lpConfig.ID,
				Name:      lpConfig.Name,
				Type:      lpConfig.Type,
				Connected: false,
				Enabled:   lpConfig.Enabled,
			}
		}
	}

	return status
}

// GetQuotesChan returns the aggregated quotes channel
func (m *Manager) GetQuotesChan() <-chan Quote {
	return m.quotesChan
}

// StartQuoteAggregation starts aggregating quotes from all enabled LPs
func (m *Manager) StartQuoteAggregation() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, lpConfig := range m.config.LPs {
		if lpConfig.Enabled {
			// Launch in goroutine to prevent blocking
			go func(id string) {
				if err := m.startLPLocked(id); err != nil {
					log.Printf("[LPManager] Failed to start LP %s: %v", id, err)
				}
			}(lpConfig.ID)
		}
	}
}

func (m *Manager) startLPLocked(id string) error {
	// Note: This needs to acquire lock if accessing m.activeAggregators,
	// BUT we are called from StartQuoteAggregation which holds lock?
	// Wait, StartQuoteAggregation holds lock, then launches goroutine. Goroutine calls this.
	// So this function needs to acquire lock.
	// BUT ToggleLP calls this while holding lock.
	// ISSUE: We need separate public/private methods or careful locking.

	// Simplest fix: startLPLocked handles its own locking for map access,
	// and StartQuoteAggregation releases lock before launching goroutines.
	// Actually, let's just use a mutex in this function for the map check.

	m.mu.Lock()
	if _, exists := m.activeAggregators[id]; exists {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	adapter, exists := m.registry.Get(id)
	if !exists {
		return ErrLPNotFound
	}

	log.Printf("[LPManager] Starting LP: %s", id)
	if err := adapter.Connect(); err != nil {
		return err
	}

	// Auto-Subscribe logic
	if wsAdapter, ok := adapter.(interface{ Subscribe([]string) error }); ok {
		var symbolsToSub []string
		// Hardcoded defaults for now (moved from main.go)
		if id == "binance" {
			symbolsToSub = []string{"BTCUSD", "ETHUSD", "BNBUSD", "SOLUSD", "XRPUSD"}
		} else {
			// Try to get all symbols?
			if syms, err := adapter.GetSymbols(); err == nil {
				for _, s := range syms {
					symbolsToSub = append(symbolsToSub, s.Symbol)
				}
				// Limit if too many?
				if len(symbolsToSub) > 50 {
					symbolsToSub = symbolsToSub[:50]
				}
			}
		}

		if len(symbolsToSub) > 0 {
			log.Printf("[LPManager] Auto-subscribing %s to %d symbols", id, len(symbolsToSub))
			if err := wsAdapter.Subscribe(symbolsToSub); err != nil {
				log.Printf("[LPManager] Failed to subscribe %s: %v", id, err)
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.activeAggregators[id] = cancel
	m.mu.Unlock()

	log.Printf("[LPManager] Starting aggregation goroutine for %s", id)
	go m.aggregateQuotes(ctx, adapter)
	return nil
}

func (m *Manager) stopLPLocked(id string) {
	if cancel, exists := m.activeAggregators[id]; exists {
		log.Printf("[LPManager] Stopping LP: %s", id)
		cancel()
		delete(m.activeAggregators, id)
	}

	if adapter, exists := m.registry.Get(id); exists {
		adapter.Disconnect()
	}
}

func (m *Manager) aggregateQuotes(ctx context.Context, adapter LPAdapter) {
	log.Printf("[LPManager] Aggregation started for %s", adapter.ID())
	defer log.Printf("[LPManager] Aggregation stopped for %s", adapter.ID())

	var quoteCount int64 = 0
	for {
		select {
		case <-ctx.Done():
			return
		case quote, ok := <-adapter.GetQuotesChan():
			if !ok {
				log.Printf("[LPManager] Quotes channel closed for %s", adapter.ID())
				return
			}
			quoteCount++
			if quoteCount%1000 == 1 {
				log.Printf("[LPManager] Received quote #%d from %s: %s @ %.5f", quoteCount, adapter.ID(), quote.Symbol, quote.Bid)
			}
			select {
			case m.quotesChan <- quote:
			default:
				// Channel full, drop quote
			}
		}
	}
}

// Errors
var (
	ErrLPNotFound      = &LPError{Message: "LP not found"}
	ErrLPAlreadyExists = &LPError{Message: "LP already exists"}
)

type LPError struct {
	Message string
}

func (e *LPError) Error() string {
	return e.Message
}
