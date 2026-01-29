package lpmanager

import (
	"fmt"
	"sync"
)

// Registry maintains all registered LP adapters
type Registry struct {
	adapters map[string]LPAdapter
	mu       sync.RWMutex
}

// NewRegistry creates a new LP registry
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]LPAdapter),
	}
}

// Register adds an LP adapter to the registry
func (r *Registry) Register(adapter LPAdapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := adapter.ID()
	if _, exists := r.adapters[id]; exists {
		return fmt.Errorf("LP adapter with ID '%s' already registered", id)
	}

	r.adapters[id] = adapter
	return nil
}

// Unregister removes an LP adapter from the registry
func (r *Registry) Unregister(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[id]; !exists {
		return fmt.Errorf("LP adapter with ID '%s' not found", id)
	}

	delete(r.adapters, id)
	return nil
}

// Get retrieves an LP adapter by ID
func (r *Registry) Get(id string) (LPAdapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[id]
	return adapter, exists
}

// List returns all registered LP adapters
func (r *Registry) List() []LPAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapters := make([]LPAdapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}
	return adapters
}

// ListIDs returns all registered LP adapter IDs
func (r *Registry) ListIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	return ids
}

// Count returns the number of registered adapters
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.adapters)
}
