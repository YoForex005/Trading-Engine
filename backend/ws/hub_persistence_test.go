package ws

import (
	"sync/atomic"
	"testing"
	"time"
)

// MockTickStore tracks all stored ticks for testing
type MockTickStore struct {
	storedTicks int64
	ticks       []struct {
		symbol string
		bid    float64
		ask    float64
	}
}

func (m *MockTickStore) StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time) {
	atomic.AddInt64(&m.storedTicks, 1)
}

// TestTickPersistence_NoClients verifies ticks are stored even without WebSocket clients
func TestTickPersistence_NoClients(t *testing.T) {
	hub := NewHub()
	mockStore := &MockTickStore{}
	hub.SetTickStore(mockStore)

	// Send 10 ticks with NO clients connected
	for i := 0; i < 10; i++ {
		tick := &MarketTick{
			Type:      "tick",
			Symbol:    "EURUSD",
			Bid:       1.08000 + float64(i)*0.00001,
			Ask:       1.08010 + float64(i)*0.00001,
			Spread:    0.00010,
			Timestamp: time.Now().Unix(),
			LP:        "TEST",
		}
		hub.BroadcastTick(tick)
	}

	// Wait for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify ALL ticks were stored
	stored := atomic.LoadInt64(&mockStore.storedTicks)
	if stored != 10 {
		t.Errorf("Expected 10 ticks to be stored without clients, got %d", stored)
	}
}

// TestTickPersistence_DisabledSymbol verifies ticks are stored even for disabled symbols
func TestTickPersistence_DisabledSymbol(t *testing.T) {
	hub := NewHub()
	mockStore := &MockTickStore{}
	hub.SetTickStore(mockStore)

	// Disable the symbol
	hub.UpdateDisabledSymbols(map[string]bool{
		"EURUSD": true,
	})

	// Send ticks for disabled symbol
	for i := 0; i < 5; i++ {
		tick := &MarketTick{
			Type:      "tick",
			Symbol:    "EURUSD",
			Bid:       1.08000,
			Ask:       1.08010,
			Spread:    0.00010,
			Timestamp: time.Now().Unix(),
			LP:        "TEST",
		}
		hub.BroadcastTick(tick)
	}

	// Wait for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify ALL ticks were stored even though symbol is disabled
	stored := atomic.LoadInt64(&mockStore.storedTicks)
	if stored != 5 {
		t.Errorf("Expected 5 ticks to be stored for disabled symbol, got %d", stored)
	}
}

// TestTickPersistence_ThrottledTicks verifies throttled ticks are still stored
func TestTickPersistence_ThrottledTicks(t *testing.T) {
	hub := NewHub()
	mockStore := &MockTickStore{}
	hub.SetTickStore(mockStore)

	// Send ticks with minimal price changes (will be throttled)
	basePrice := 1.08000
	for i := 0; i < 10; i++ {
		tick := &MarketTick{
			Type:      "tick",
			Symbol:    "EURUSD",
			Bid:       basePrice + float64(i)*0.0000001, // Tiny change - will be throttled
			Ask:       basePrice + 0.00010,
			Spread:    0.00010,
			Timestamp: time.Now().Unix(),
			LP:        "TEST",
		}
		hub.BroadcastTick(tick)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify ALL ticks were stored even though most were throttled from broadcast
	stored := atomic.LoadInt64(&mockStore.storedTicks)
	if stored != 10 {
		t.Errorf("Expected 10 ticks to be stored (including throttled), got %d", stored)
	}

	// Verify many were throttled from broadcast
	throttled := atomic.LoadInt64(&hub.ticksThrottled)
	if throttled < 5 {
		t.Errorf("Expected at least 5 ticks to be throttled, got %d", throttled)
	}
}

// TestOptimizedHub_Persistence tests the optimized hub as well
func TestOptimizedHub_Persistence(t *testing.T) {
	hub := NewOptimizedHub()
	mockStore := &MockTickStore{}
	hub.SetTickStore(mockStore)

	// Send ticks with NO clients connected
	for i := 0; i < 10; i++ {
		tick := &MarketTick{
			Type:      "tick",
			Symbol:    "GBPUSD",
			Bid:       1.25000 + float64(i)*0.00001,
			Ask:       1.25010 + float64(i)*0.00001,
			Spread:    0.00010,
			Timestamp: time.Now().Unix(),
			LP:        "TEST",
		}
		hub.BroadcastTickOptimized(tick)
	}

	// Wait for async operations
	time.Sleep(100 * time.Millisecond)

	// Verify ALL ticks were stored
	stored := atomic.LoadInt64(&mockStore.storedTicks)
	if stored != 10 {
		t.Errorf("OptimizedHub: Expected 10 ticks to be stored without clients, got %d", stored)
	}
}
