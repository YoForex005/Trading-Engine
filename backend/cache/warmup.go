package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// WarmupStrategy defines how to warm up the cache
type WarmupStrategy interface {
	// Warmup loads data into cache
	Warmup(ctx context.Context, cache *MultiTierCache) error

	// ShouldRefresh determines if cache should be refreshed
	ShouldRefresh() bool
}

// CacheWarmer manages cache warming on startup and periodic refresh
type CacheWarmer struct {
	cache      *MultiTierCache
	strategies []WarmupStrategy

	mu         sync.RWMutex
	lastWarmup time.Time
	warmupTime time.Duration

	// Configuration
	refreshInterval time.Duration
	enabled         bool
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(cache *MultiTierCache) *CacheWarmer {
	return &CacheWarmer{
		cache:           cache,
		strategies:      make([]WarmupStrategy, 0),
		refreshInterval: 1 * time.Hour,
		enabled:         true,
	}
}

// AddStrategy adds a warmup strategy
func (w *CacheWarmer) AddStrategy(strategy WarmupStrategy) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.strategies = append(w.strategies, strategy)
}

// Warmup executes all warmup strategies
func (w *CacheWarmer) Warmup(ctx context.Context) error {
	if !w.enabled {
		return nil
	}

	start := time.Now()
	log.Println("[CacheWarmer] Starting cache warmup...")

	w.mu.RLock()
	strategies := w.strategies
	w.mu.RUnlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(strategies))

	for _, strategy := range strategies {
		wg.Add(1)
		go func(s WarmupStrategy) {
			defer wg.Done()
			if err := s.Warmup(ctx, w.cache); err != nil {
				errors <- err
			}
		}(strategy)
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	duration := time.Since(start)
	w.mu.Lock()
	w.lastWarmup = time.Now()
	w.warmupTime = duration
	w.mu.Unlock()

	if len(errs) > 0 {
		log.Printf("[CacheWarmer] Warmup completed with errors in %v: %v", duration, errs)
		return fmt.Errorf("warmup completed with %d errors", len(errs))
	}

	log.Printf("[CacheWarmer] Cache warmup completed successfully in %v", duration)
	return nil
}

// StartPeriodicRefresh starts periodic cache refresh
func (w *CacheWarmer) StartPeriodicRefresh(ctx context.Context) {
	if !w.enabled {
		return
	}

	ticker := time.NewTicker(w.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.mu.RLock()
			strategies := w.strategies
			w.mu.RUnlock()

			for _, strategy := range strategies {
				if strategy.ShouldRefresh() {
					if err := strategy.Warmup(ctx, w.cache); err != nil {
						log.Printf("[CacheWarmer] Refresh error: %v", err)
					}
				}
			}
		}
	}
}

// SetRefreshInterval sets the refresh interval
func (w *CacheWarmer) SetRefreshInterval(interval time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.refreshInterval = interval
}

// SetEnabled enables/disables cache warming
func (w *CacheWarmer) SetEnabled(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.enabled = enabled
}

// Stats returns warmup statistics
func (w *CacheWarmer) Stats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return map[string]interface{}{
		"last_warmup":      w.lastWarmup,
		"warmup_duration":  w.warmupTime,
		"refresh_interval": w.refreshInterval,
		"strategies_count": len(w.strategies),
		"enabled":          w.enabled,
	}
}

// Common warmup strategies

// SymbolConfigWarmup warms up symbol configuration cache
type SymbolConfigWarmup struct {
	loader func(ctx context.Context) (map[string]interface{}, error)
}

func NewSymbolConfigWarmup(loader func(ctx context.Context) (map[string]interface{}, error)) *SymbolConfigWarmup {
	return &SymbolConfigWarmup{loader: loader}
}

func (s *SymbolConfigWarmup) Warmup(ctx context.Context, cache *MultiTierCache) error {
	symbols, err := s.loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load symbols: %w", err)
	}

	items := make(map[string]interface{})
	for symbol, config := range symbols {
		key := CacheKey(NS_Symbols, symbol)
		items[key] = config
	}

	if err := cache.SetMulti(ctx, items, TTL_Symbol_Config); err != nil {
		return fmt.Errorf("failed to cache symbols: %w", err)
	}

	log.Printf("[CacheWarmer] Loaded %d symbol configurations", len(symbols))
	return nil
}

func (s *SymbolConfigWarmup) ShouldRefresh() bool {
	return true // Always refresh on interval
}

// OHLCHistoricalWarmup warms up historical OHLC data
type OHLCHistoricalWarmup struct {
	loader  func(ctx context.Context, symbols []string) (map[string]interface{}, error)
	symbols []string
}

func NewOHLCHistoricalWarmup(
	loader func(ctx context.Context, symbols []string) (map[string]interface{}, error),
	symbols []string,
) *OHLCHistoricalWarmup {
	return &OHLCHistoricalWarmup{
		loader:  loader,
		symbols: symbols,
	}
}

func (o *OHLCHistoricalWarmup) Warmup(ctx context.Context, cache *MultiTierCache) error {
	data, err := o.loader(ctx, o.symbols)
	if err != nil {
		return fmt.Errorf("failed to load OHLC data: %w", err)
	}

	items := make(map[string]interface{})
	for key, value := range data {
		cacheKey := CacheKey(NS_OHLC, key)
		items[cacheKey] = value
	}

	if err := cache.SetMulti(ctx, items, TTL_OHLC_Historical); err != nil {
		return fmt.Errorf("failed to cache OHLC data: %w", err)
	}

	log.Printf("[CacheWarmer] Loaded %d OHLC datasets", len(data))
	return nil
}

func (o *OHLCHistoricalWarmup) ShouldRefresh() bool {
	return false // Historical data rarely changes
}

// AccountWarmup warms up frequently accessed accounts
type AccountWarmup struct {
	loader func(ctx context.Context) (map[string]interface{}, error)
}

func NewAccountWarmup(loader func(ctx context.Context) (map[string]interface{}, error)) *AccountWarmup {
	return &AccountWarmup{loader: loader}
}

func (a *AccountWarmup) Warmup(ctx context.Context, cache *MultiTierCache) error {
	accounts, err := a.loader(ctx)
	if err != nil {
		return fmt.Errorf("failed to load accounts: %w", err)
	}

	items := make(map[string]interface{})
	for accountID, account := range accounts {
		key := CacheKey(NS_Accounts, accountID)
		items[key] = account
	}

	if err := cache.SetMulti(ctx, items, TTL_User_Account); err != nil {
		return fmt.Errorf("failed to cache accounts: %w", err)
	}

	log.Printf("[CacheWarmer] Loaded %d accounts", len(accounts))
	return nil
}

func (a *AccountWarmup) ShouldRefresh() bool {
	return true // Accounts change frequently
}
