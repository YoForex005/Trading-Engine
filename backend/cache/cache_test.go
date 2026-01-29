package cache

import (
	"context"
	"testing"
	"time"
)

// TestMemoryCache tests in-memory cache functionality
func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache(1024*1024, 100) // 1MB, 100 items
	ctx := context.Background()

	// Test Set and Get
	t.Run("SetGet", func(t *testing.T) {
		err := cache.Set(ctx, "key1", "value1", 1*time.Minute)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		value, err := cache.Get(ctx, "key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if value != "value1" {
			t.Errorf("Expected 'value1', got %v", value)
		}
	})

	// Test TTL expiration
	t.Run("TTLExpiration", func(t *testing.T) {
		err := cache.Set(ctx, "key2", "value2", 100*time.Millisecond)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		// Should exist immediately
		value, err := cache.Get(ctx, "key2")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if value != "value2" {
			t.Errorf("Expected 'value2', got %v", value)
		}

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		_, err = cache.Get(ctx, "key2")
		if err != ErrExpired {
			t.Errorf("Expected ErrExpired, got %v", err)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		cache.Set(ctx, "key3", "value3", 1*time.Minute)

		err := cache.Delete(ctx, "key3")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err = cache.Get(ctx, "key3")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound after delete, got %v", err)
		}
	})

	// Test Exists
	t.Run("Exists", func(t *testing.T) {
		cache.Set(ctx, "key4", "value4", 1*time.Minute)

		exists, err := cache.Exists(ctx, "key4")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Key should exist")
		}

		cache.Delete(ctx, "key4")

		exists, err = cache.Exists(ctx, "key4")
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Key should not exist after delete")
		}
	})

	// Test LRU eviction
	t.Run("LRUEviction", func(t *testing.T) {
		smallCache := NewMemoryCache(0, 3) // Max 3 items

		smallCache.Set(ctx, "k1", "v1", 1*time.Minute)
		smallCache.Set(ctx, "k2", "v2", 1*time.Minute)
		smallCache.Set(ctx, "k3", "v3", 1*time.Minute)

		// Access k1 to make it most recent
		smallCache.Get(ctx, "k1")

		// Add k4, should evict k2 (least recently used)
		smallCache.Set(ctx, "k4", "v4", 1*time.Minute)

		_, err := smallCache.Get(ctx, "k2")
		if err != ErrNotFound {
			t.Error("k2 should have been evicted")
		}

		// k1, k3, k4 should still exist
		if _, err := smallCache.Get(ctx, "k1"); err != nil {
			t.Error("k1 should exist")
		}
		if _, err := smallCache.Get(ctx, "k3"); err != nil {
			t.Error("k3 should exist")
		}
		if _, err := smallCache.Get(ctx, "k4"); err != nil {
			t.Error("k4 should exist")
		}
	})

	// Test Stats
	t.Run("Stats", func(t *testing.T) {
		// Use a fresh cache to avoid interference from other tests
		freshCache := NewMemoryCache(1024*1024, 100)

		freshCache.Set(ctx, "s1", "v1", 1*time.Minute)
		freshCache.Get(ctx, "s1")      // Hit
		freshCache.Get(ctx, "missing") // Miss

		stats := freshCache.Stats()

		if stats.Hits != 1 {
			t.Errorf("Expected 1 hit, got %d", stats.Hits)
		}
		if stats.Misses != 1 {
			t.Errorf("Expected 1 miss, got %d", stats.Misses)
		}
		if stats.Sets != 1 {
			t.Errorf("Expected 1 set, got %d", stats.Sets)
		}
		if stats.HitRate != 0.5 {
			t.Errorf("Expected 50%% hit rate, got %.2f%%", stats.HitRate*100)
		}
	})
}

// TestMultiTierCache tests multi-tier cache functionality
func TestMultiTierCache(t *testing.T) {
	// Skip Redis tests if Redis is not available
	// In production, use environment variable to enable Redis tests

	ctx := context.Background()

	// Loader function simulates database
	dbData := map[string]interface{}{
		"db_key1": "db_value1",
		"db_key2": "db_value2",
	}

	loader := func(ctx context.Context, key string) (interface{}, error) {
		if value, ok := dbData[key]; ok {
			return value, nil
		}
		return nil, ErrNotFound
	}

	// Create multi-tier cache (L1 only for testing)
	cache, err := NewMultiTierCache(1024*1024, 100, nil, loader)
	if err != nil {
		t.Fatalf("Failed to create multi-tier cache: %v", err)
	}

	// Test L1 cache
	t.Run("L1Cache", func(t *testing.T) {
		err := cache.Set(ctx, "l1_key", "l1_value", 1*time.Minute)
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		value, err := cache.Get(ctx, "l1_key")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if value != "l1_value" {
			t.Errorf("Expected 'l1_value', got %v", value)
		}
	})

	// Test loader (L3)
	t.Run("LoaderL3", func(t *testing.T) {
		// Key not in cache, should load from loader
		value, err := cache.Get(ctx, "db_key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if value != "db_value1" {
			t.Errorf("Expected 'db_value1' from loader, got %v", value)
		}

		// Should now be in L1 cache
		value, err = cache.Get(ctx, "db_key1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if value != "db_value1" {
			t.Errorf("Expected 'db_value1' from L1, got %v", value)
		}
	})

	// Test GetMulti
	t.Run("GetMulti", func(t *testing.T) {
		cache.Set(ctx, "m1", "v1", 1*time.Minute)
		cache.Set(ctx, "m2", "v2", 1*time.Minute)

		results, err := cache.GetMulti(ctx, []string{"m1", "m2", "db_key2"})
		if err != nil {
			t.Fatalf("GetMulti failed: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		if results["m1"] != "v1" {
			t.Error("m1 should be v1")
		}
		if results["m2"] != "v2" {
			t.Error("m2 should be v2")
		}
		if results["db_key2"] != "db_value2" {
			t.Error("db_key2 should be db_value2 from loader")
		}
	})
}

// TestCacheKey tests cache key generation
func TestCacheKey(t *testing.T) {
	tests := []struct {
		namespace string
		key       string
		expected  string
	}{
		{"symbols", "BTCUSD", "symbols:BTCUSD"},
		{"", "key", "key"},
		{"accounts", "123", "accounts:123"},
	}

	for _, tt := range tests {
		result := CacheKey(tt.namespace, tt.key)
		if result != tt.expected {
			t.Errorf("CacheKey(%s, %s) = %s, want %s", tt.namespace, tt.key, result, tt.expected)
		}
	}
}

// BenchmarkMemoryCacheGet benchmarks cache get operations
func BenchmarkMemoryCacheGet(b *testing.B) {
	cache := NewMemoryCache(10*1024*1024, 10000)
	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := CacheKey(NS_Symbols, string(rune('A'+i%26)))
		cache.Set(ctx, key, map[string]interface{}{"symbol": key}, 1*time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := CacheKey(NS_Symbols, string(rune('A'+i%26)))
		cache.Get(ctx, key)
	}
}

// BenchmarkMemoryCacheSet benchmarks cache set operations
func BenchmarkMemoryCacheSet(b *testing.B) {
	cache := NewMemoryCache(10*1024*1024, 10000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := CacheKey(NS_Symbols, string(rune('A'+i%26)))
		cache.Set(ctx, key, map[string]interface{}{"symbol": key}, 1*time.Hour)
	}
}

// BenchmarkMultiTierCache benchmarks multi-tier cache
func BenchmarkMultiTierCache(b *testing.B) {
	loader := func(ctx context.Context, key string) (interface{}, error) {
		return map[string]interface{}{"key": key}, nil
	}

	cache, _ := NewMultiTierCache(10*1024*1024, 10000, nil, loader)
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < 1000; i++ {
		key := CacheKey(NS_Symbols, string(rune('A'+i%26)))
		cache.Set(ctx, key, map[string]interface{}{"symbol": key}, 1*time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := CacheKey(NS_Symbols, string(rune('A'+i%26)))
		cache.Get(ctx, key)
	}
}
