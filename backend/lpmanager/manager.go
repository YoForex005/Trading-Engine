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
	lpConfigMap       map[string]*LPConfig // O(1) lookup map for LP configs
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
		lpConfigMap:       make(map[string]*LPConfig),
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
			// Populate the O(1) lookup map for default config
			m.lpConfigMap = make(map[string]*LPConfig, len(m.config.LPs))
			for i := range m.config.LPs {
				m.lpConfigMap[m.config.LPs[i].ID] = &m.config.LPs[i]
			}
			return m.saveConfigLocked()
		}
		return err
	}

	var config LPManagerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	m.config = &config

	// Populate the O(1) lookup map
	m.lpConfigMap = make(map[string]*LPConfig, len(m.config.LPs))
	for i := range m.config.LPs {
		m.lpConfigMap[m.config.LPs[i].ID] = &m.config.LPs[i]
	}

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

	// O(1) map lookup instead of O(n) iteration
	return m.lpConfigMap[id]
}

// AddLP adds a new LP configuration
func (m *Manager) AddLP(config LPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// O(1) check if ID already exists using map
	if _, exists := m.lpConfigMap[config.ID]; exists {
		return ErrLPAlreadyExists
	}

	m.config.LPs = append(m.config.LPs, config)
	// Update map with pointer to the newly added config
	m.lpConfigMap[config.ID] = &m.config.LPs[len(m.config.LPs)-1]
	return m.saveConfigLocked()
}

// UpdateLP updates an existing LP configuration
func (m *Manager) UpdateLP(config LPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// O(1) lookup to find the LP config
	lpConfig, exists := m.lpConfigMap[config.ID]
	if !exists {
		return ErrLPNotFound
	}

	// Update the config in place
	*lpConfig = config
	return m.saveConfigLocked()
}

// RemoveLP removes an LP configuration
func (m *Manager) RemoveLP(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// O(1) check if LP exists
	if _, exists := m.lpConfigMap[id]; !exists {
		return ErrLPNotFound
	}

	// Remove from map
	delete(m.lpConfigMap, id)

	// Remove from slice
	for i := range m.config.LPs {
		if m.config.LPs[i].ID == id {
			m.config.LPs = append(m.config.LPs[:i], m.config.LPs[i+1:]...)
			break
		}
	}

	return m.saveConfigLocked()
}

// ToggleLP enables or disables an LP
func (m *Manager) ToggleLP(id string) error {
	m.mu.Lock()
	// Defer unlock is NOT used here because we need to unlock before calling startLP
	// defer m.mu.Unlock()

	// O(1) lookup using map
	targetLP, exists := m.lpConfigMap[id]
	if !exists {
		m.mu.Unlock()
		return ErrLPNotFound
	}

	newState := !targetLP.Enabled
	targetLP.Enabled = newState

	// Save config while locked
	if err := m.saveConfigLocked(); err != nil {
		m.mu.Unlock()
		return err
	}

	m.mu.Unlock() // Release lock before I/O

	log.Printf("[LPManager] LP %s enabled=%v", id, newState)

	if newState {
		if err := m.startLP(id); err != nil {
			log.Printf("[LPManager] Failed to start LP %s: %v", id, err)
			// Revert state on failure
			m.mu.Lock()
			// O(1) lookup to revert the state
			if lp, exists := m.lpConfigMap[id]; exists {
				lp.Enabled = false
			}
			m.saveConfigLocked()
			m.mu.Unlock()
			return err
		}
	} else {
		m.stopLP(id)
	}

	return nil
}

// SetLPEnabled sets the enabled state of an LP
func (m *Manager) SetLPEnabled(id string, enabled bool) error {
	m.mu.Lock()

	// O(1) lookup using map
	targetLP, exists := m.lpConfigMap[id]
	if !exists {
		m.mu.Unlock()
		return ErrLPNotFound
	}

	targetLP.Enabled = enabled
	if err := m.saveConfigLocked(); err != nil {
		m.mu.Unlock()
		return err
	}
	m.mu.Unlock()

	log.Printf("[LPManager] LP %s enabled=%v", id, enabled)

	if enabled {
		if err := m.startLP(id); err != nil {
			// Revert using O(1) lookup
			m.mu.Lock()
			if lp, exists := m.lpConfigMap[id]; exists {
				lp.Enabled = false
			}
			m.saveConfigLocked()
			m.mu.Unlock()
			return err
		}
	} else {
		m.stopLP(id)
	}

	return nil
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
				if err := m.startLP(id); err != nil {
					log.Printf("[LPManager] Failed to start LP %s: %v", id, err)
				}
			}(lpConfig.ID)
		}
	}
}

func (m *Manager) startLP(id string) error {
	// Note: This function acquires locks internally only when needed.
	// It relies on m.activeAggregators check.

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

		// First, check if symbols are configured in LP config using O(1) lookup
		m.mu.Lock()
		var configSymbols []string
		if lp, exists := m.lpConfigMap[id]; exists {
			configSymbols = lp.Symbols
		}
		m.mu.Unlock()

		if len(configSymbols) > 0 {
			// Use symbols from config
			symbolsToSub = configSymbols
			log.Printf("[LPManager] Using %d configured symbols for %s", len(symbolsToSub), id)
		} else {
			// Fallback: fetch available symbols from adapter
			if syms, err := adapter.GetSymbols(); err == nil {
				for _, s := range syms {
					symbolsToSub = append(symbolsToSub, s.Symbol)
				}
				// Limit if too many
				if len(symbolsToSub) > 50 {
					symbolsToSub = symbolsToSub[:50]
				}
				log.Printf("[LPManager] Auto-discovered %d symbols for %s", len(symbolsToSub), id)
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

func (m *Manager) stopLP(id string) {
	m.mu.Lock()
	cancel, exists := m.activeAggregators[id]
	if exists {
		delete(m.activeAggregators, id)
	}
	m.mu.Unlock()

	if exists {
		log.Printf("[LPManager] Stopping LP: %s", id)
		cancel()
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
