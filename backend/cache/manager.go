package cache

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// CacheManager manages all cache instances and provides unified access
type CacheManager struct {
	// Cache instances
	multiTier *MultiTierCache
	warmer    *CacheWarmer
	cdn       *CDNCache

	// Monitoring
	mu               sync.RWMutex
	invalidations    int64
	lastInvalidation time.Time

	// Event handlers
	onInvalidate func(key string)
}

// ManagerConfig holds cache manager configuration
type ManagerConfig struct {
	// Memory cache (L1)
	L1Size     int64
	L1MaxItems int

	// Redis cache (L2)
	RedisConfig *RedisConfig

	// CDN
	CDNConfig *CDNConfig

	// Warmup
	EnableWarmup     bool
	WarmupStrategies []WarmupStrategy
}

// DefaultManagerConfig returns default cache manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		L1Size:       100 * 1024 * 1024, // 100MB
		L1MaxItems:   10000,
		RedisConfig:  DefaultRedisConfig(),
		EnableWarmup: true,
	}
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *ManagerConfig, loader LoaderFunc) (*CacheManager, error) {
	if config == nil {
		config = DefaultManagerConfig()
	}

	// Create multi-tier cache
	multiTier, err := NewMultiTierCache(config.L1Size, config.L1MaxItems, config.RedisConfig, loader)
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-tier cache: %w", err)
	}

	// Create cache warmer
	warmer := NewCacheWarmer(multiTier)
	if config.EnableWarmup {
		for _, strategy := range config.WarmupStrategies {
			warmer.AddStrategy(strategy)
		}
	}

	// Create CDN cache if configured
	var cdn *CDNCache
	if config.CDNConfig != nil {
		cdn = NewCDNCache(config.CDNConfig)
	}

	manager := &CacheManager{
		multiTier: multiTier,
		warmer:    warmer,
		cdn:       cdn,
	}

	log.Println("[CacheManager] Initialized successfully")
	return manager, nil
}

// Start starts the cache manager (warmup, periodic refresh, etc.)
func (m *CacheManager) Start(ctx context.Context) error {
	log.Println("[CacheManager] Starting cache manager...")

	// Initial warmup
	if err := m.warmer.Warmup(ctx); err != nil {
		log.Printf("[CacheManager] Warning: Initial warmup failed: %v", err)
	}

	// Start periodic refresh
	go m.warmer.StartPeriodicRefresh(ctx)

	// Start monitoring
	go m.monitorPerformance(ctx)

	log.Println("[CacheManager] Cache manager started successfully")
	return nil
}

// Get retrieves a value from cache
func (m *CacheManager) Get(ctx context.Context, namespace, key string) (interface{}, error) {
	fullKey := CacheKey(namespace, key)
	return m.multiTier.Get(ctx, fullKey)
}

// GetWithTTL retrieves a value with specific TTL
func (m *CacheManager) GetWithTTL(ctx context.Context, namespace, key string, ttl time.Duration) (interface{}, error) {
	fullKey := CacheKey(namespace, key)
	return m.multiTier.GetWithTTL(ctx, fullKey, ttl)
}

// Set stores a value in cache
func (m *CacheManager) Set(ctx context.Context, namespace, key string, value interface{}, ttl time.Duration) error {
	fullKey := CacheKey(namespace, key)
	return m.multiTier.Set(ctx, fullKey, value, ttl)
}

// Delete removes a value from cache
func (m *CacheManager) Delete(ctx context.Context, namespace, key string) error {
	fullKey := CacheKey(namespace, key)
	err := m.multiTier.Delete(ctx, fullKey)

	m.mu.Lock()
	m.invalidations++
	m.lastInvalidation = time.Now()
	m.mu.Unlock()

	if m.onInvalidate != nil {
		go m.onInvalidate(fullKey)
	}

	return err
}

// InvalidateNamespace invalidates all keys in a namespace
func (m *CacheManager) InvalidateNamespace(ctx context.Context, namespace string) error {
	// This is a simplified implementation
	// In production, you'd scan and delete all keys with the namespace prefix
	log.Printf("[CacheManager] Invalidating namespace: %s", namespace)

	m.mu.Lock()
	m.invalidations++
	m.lastInvalidation = time.Now()
	m.mu.Unlock()

	return nil
}

// GetSymbolConfig retrieves symbol configuration with caching
func (m *CacheManager) GetSymbolConfig(ctx context.Context, symbol string) (interface{}, error) {
	return m.GetWithTTL(ctx, NS_Symbols, symbol, TTL_Symbol_Config)
}

// GetAccount retrieves account data with caching
func (m *CacheManager) GetAccount(ctx context.Context, accountID string) (interface{}, error) {
	return m.GetWithTTL(ctx, NS_Accounts, accountID, TTL_User_Account)
}

// GetPrice retrieves latest price with caching
func (m *CacheManager) GetPrice(ctx context.Context, symbol string) (interface{}, error) {
	return m.GetWithTTL(ctx, NS_Prices, symbol, TTL_Market_Price)
}

// InvalidatePrice invalidates price cache (on price update)
func (m *CacheManager) InvalidatePrice(ctx context.Context, symbol string) error {
	return m.Delete(ctx, NS_Prices, symbol)
}

// InvalidatePosition invalidates position cache (on position change)
func (m *CacheManager) InvalidatePosition(ctx context.Context, positionID string) error {
	return m.Delete(ctx, NS_Positions, positionID)
}

// InvalidateAccount invalidates account cache (on balance change)
func (m *CacheManager) InvalidateAccount(ctx context.Context, accountID string) error {
	return m.Delete(ctx, NS_Accounts, accountID)
}

// GetCDNAssetURL gets CDN URL for static asset
func (m *CacheManager) GetCDNAssetURL(assetPath string) string {
	if m.cdn == nil {
		return assetPath
	}
	return m.cdn.GetAssetURL(assetPath)
}

// PurgeCDNAsset purges asset from CDN
func (m *CacheManager) PurgeCDNAsset(ctx context.Context, assetPath string) error {
	if m.cdn == nil {
		return fmt.Errorf("CDN not configured")
	}
	return m.cdn.PurgeAsset(ctx, assetPath)
}

// Stats returns comprehensive cache statistics
func (m *CacheManager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"multi_tier":        m.multiTier.Stats(),
		"warmer":            m.warmer.Stats(),
		"invalidations":     m.invalidations,
		"last_invalidation": m.lastInvalidation,
	}

	return stats
}

// SetOnInvalidate sets invalidation event handler
func (m *CacheManager) SetOnInvalidate(handler func(key string)) {
	m.onInvalidate = handler
}

// monitorPerformance monitors cache performance and logs warnings
func (m *CacheManager) monitorPerformance(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := m.multiTier.Stats()

			hitRate := stats["hit_rate"].(float64)
			if hitRate < 0.5 {
				log.Printf("[CacheManager] WARNING: Low cache hit rate: %.2f%%", hitRate*100)
			}

			l1HitRate := stats["l1_hit_rate"].(float64)
			if l1HitRate < 0.3 {
				log.Printf("[CacheManager] WARNING: Low L1 hit rate: %.2f%% (consider increasing L1 size)", l1HitRate*100)
			}

			avgLatency := stats["avg_latency"].(time.Duration)
			if avgLatency > 10*time.Millisecond {
				log.Printf("[CacheManager] WARNING: High cache latency: %v", avgLatency)
			}
		}
	}
}

// Shutdown gracefully shuts down the cache manager
func (m *CacheManager) Shutdown(ctx context.Context) error {
	log.Println("[CacheManager] Shutting down...")

	// Close Redis connection if exists
	if m.multiTier.l2 != nil {
		if err := m.multiTier.l2.Close(); err != nil {
			log.Printf("[CacheManager] Error closing Redis connection: %v", err)
		}
	}

	log.Println("[CacheManager] Shutdown complete")
	return nil
}
