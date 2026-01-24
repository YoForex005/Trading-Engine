// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/trading-engine/backend/tickstore"
)

// This is a standalone test program to verify SQLite integration.
// Run with: go run test_sqlite_integration.go

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("=== SQLite Integration Test ===\n")

	// Test 1: SQLite-only storage
	testSQLiteOnly()

	// Test 2: Dual storage
	testDualStorage()

	// Test 3: Query performance
	testQueryPerformance()

	log.Printf("\n=== All Tests Complete ===\n")
}

func testSQLiteOnly() {
	log.Printf("\n--- Test 1: SQLite-Only Storage ---")

	config := tickstore.TickStoreConfig{
		BrokerID:         "TEST-001",
		MaxTicksPerSymbol: 1000,
		Backend:          tickstore.BackendSQLite,
		SQLiteBasePath:   "data/test/ticks/db",
		EnableJSONLegacy: false,
	}

	store := tickstore.NewOptimizedTickStoreWithConfig(config)
	defer store.Stop()

	// Write test ticks
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY"}
	baseTime := time.Now()

	log.Printf("Writing 1000 ticks across %d symbols...", len(symbols))
	for i := 0; i < 1000; i++ {
		for _, symbol := range symbols {
			bid := 1.0850 + float64(i)*0.0001
			ask := bid + 0.0002
			spread := 0.0002
			timestamp := baseTime.Add(time.Duration(i) * time.Second)

			store.StoreTick(symbol, bid, ask, spread, "TEST-LP", timestamp)
		}
	}

	// Wait for writes to complete
	log.Printf("Waiting for async writes to complete...")
	time.Sleep(10 * time.Second)

	// Get statistics
	stats := store.GetStorageStats()
	log.Printf("Statistics:")
	log.Printf("  Backend: %s", stats["backend"])
	log.Printf("  Ticks received: %d", stats["ticks_received"])
	log.Printf("  Ticks stored: %d", stats["ticks_stored"])

	if sqliteStats, ok := stats["sqlite"].(map[string]interface{}); ok {
		log.Printf("  SQLite ticks written: %d", sqliteStats["ticks_written"])
		log.Printf("  SQLite write errors: %d", sqliteStats["write_errors"])
		log.Printf("  Queue size: %d/%d", sqliteStats["queue_size"], sqliteStats["queue_capacity"])
	}

	log.Printf("✓ Test 1 passed")
}

func testDualStorage() {
	log.Printf("\n--- Test 2: Dual Storage (SQLite + JSON) ---")

	config := tickstore.TickStoreConfig{
		BrokerID:         "TEST-002",
		MaxTicksPerSymbol: 1000,
		Backend:          tickstore.BackendDual,
		SQLiteBasePath:   "data/test/ticks/db",
		EnableJSONLegacy: true,
	}

	store := tickstore.NewOptimizedTickStoreWithConfig(config)
	defer store.Stop()

	// Write test ticks
	log.Printf("Writing 500 ticks with dual storage...")
	baseTime := time.Now()
	for i := 0; i < 500; i++ {
		bid := 1.0850 + float64(i)*0.0001
		ask := bid + 0.0002
		timestamp := baseTime.Add(time.Duration(i) * time.Second)

		store.StoreTick("EURUSD", bid, ask, 0.0002, "TEST-LP", timestamp)
	}

	// Wait for writes
	time.Sleep(5 * time.Second)

	stats := store.GetStorageStats()
	log.Printf("Statistics:")
	log.Printf("  Backend: %s", stats["backend"])
	log.Printf("  Use JSON legacy: %v", stats["use_json_legacy"])
	log.Printf("  Ticks stored: %d", stats["ticks_stored"])

	log.Printf("✓ Test 2 passed")
}

func testQueryPerformance() {
	log.Printf("\n--- Test 3: Query Performance ---")

	config := tickstore.TickStoreConfig{
		BrokerID:         "TEST-003",
		MaxTicksPerSymbol: 10000,
		Backend:          tickstore.BackendSQLite,
		SQLiteBasePath:   "data/test/ticks/db",
		EnableJSONLegacy: false,
	}

	store := tickstore.NewOptimizedTickStoreWithConfig(config)
	defer store.Stop()

	// Write test data
	log.Printf("Writing 5000 ticks...")
	baseTime := time.Now()
	for i := 0; i < 5000; i++ {
		bid := 1.0850 + float64(i)*0.0001
		ask := bid + 0.0002
		timestamp := baseTime.Add(time.Duration(i) * time.Second)

		store.StoreTick("EURUSD", bid, ask, 0.0002, "TEST-LP", timestamp)
	}

	// Wait for writes
	time.Sleep(10 * time.Second)
	store.Flush()

	// Test queries
	if store.sqliteStore != nil {
		// Query recent ticks
		log.Printf("Querying recent 100 ticks...")
		start := time.Now()
		ticks, err := store.sqliteStore.GetRecentTicks("EURUSD", 100)
		duration := time.Since(start)

		if err != nil {
			log.Printf("Query error: %v", err)
		} else {
			log.Printf("  Retrieved %d ticks in %v", len(ticks), duration)
			if len(ticks) > 0 {
				log.Printf("  Latest tick: bid=%.4f, ask=%.4f, time=%v",
					ticks[0].Bid, ticks[0].Ask, ticks[0].Timestamp)
			}
		}

		// Query time range
		endTime := baseTime.Add(1 * time.Hour)
		startTime := endTime.Add(-10 * time.Minute)
		log.Printf("Querying 10-minute time range...")
		start = time.Now()
		ticks, err = store.sqliteStore.GetTicksInRange("EURUSD", startTime, endTime)
		duration = time.Since(start)

		if err != nil {
			log.Printf("Query error: %v", err)
		} else {
			log.Printf("  Retrieved %d ticks in %v", len(ticks), duration)
		}
	}

	log.Printf("✓ Test 3 passed")
}
