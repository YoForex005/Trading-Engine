package adapters

import (
	"testing"
	"time"
)

func TestOANDAAdapter_Connect(t *testing.T) {
	t.Skip("Integration test - requires OANDA API credentials")

	// This demonstrates the pattern for integration testing
	// In CI/CD, use test credentials or mock OANDA API

	adapter := NewOANDAAdapter("test-api-key", "test-account-id")
	if adapter == nil {
		t.Fatal("failed to create OANDA adapter")
	}

	err := adapter.Connect()
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer adapter.Disconnect()

	// Verify connection status
	if !adapter.IsConnected() {
		t.Error("adapter should report connected after Connect()")
	}
}

func TestOANDAAdapter_Reconnect(t *testing.T) {
	t.Skip("Integration test - requires OANDA API")

	adapter := NewOANDAAdapter("test-api-key", "test-account-id")

	// Initial connect
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("initial connect failed: %v", err)
	}

	// Simulate disconnect
	adapter.Disconnect()

	// Brief wait to ensure disconnection
	time.Sleep(100 * time.Millisecond)

	// Verify reconnect works
	err = adapter.Connect()
	if err != nil {
		t.Fatalf("reconnect failed: %v", err)
	}
	defer adapter.Disconnect()

	if !adapter.IsConnected() {
		t.Error("adapter should reconnect successfully")
	}
}

func TestOANDAAdapter_QuoteStream(t *testing.T) {
	t.Skip("Integration test - requires OANDA API")

	adapter := NewOANDAAdapter("test-api-key", "test-account-id")
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer adapter.Disconnect()

	// Subscribe to quotes
	symbols := []string{"EURUSD"}
	err = adapter.Subscribe(symbols)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Wait for at least one quote with timeout
	quoteChan := adapter.GetQuotesChan()
	select {
	case quote := <-quoteChan:
		if quote.Symbol != "EURUSD" {
			t.Errorf("got quote for %s, want EURUSD", quote.Symbol)
		}
		if quote.Bid == 0 || quote.Ask == 0 {
			t.Error("quote has zero bid or ask")
		}
		if quote.LP != "oanda" {
			t.Errorf("got LP %s, want oanda", quote.LP)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for quote")
	}
}

func TestOANDAAdapter_GetSymbols(t *testing.T) {
	t.Skip("Integration test - requires OANDA API")

	adapter := NewOANDAAdapter("test-api-key", "test-account-id")
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer adapter.Disconnect()

	// Fetch available symbols
	symbols, err := adapter.GetSymbols()
	if err != nil {
		t.Fatalf("GetSymbols failed: %v", err)
	}

	if len(symbols) == 0 {
		t.Error("expected at least some symbols")
	}

	// Verify symbol structure
	for _, sym := range symbols {
		if sym.Symbol == "" {
			t.Error("symbol has empty Symbol field")
		}
		if sym.Type == "" {
			t.Error("symbol has empty Type field")
		}
	}
}

func TestOANDAAdapter_DisconnectWhileStreaming(t *testing.T) {
	t.Skip("Integration test - requires OANDA API")

	adapter := NewOANDAAdapter("test-api-key", "test-account-id")
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	// Subscribe to quotes
	err = adapter.Subscribe([]string{"EURUSD"})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Start receiving quotes
	quoteChan := adapter.GetQuotesChan()
	go func() {
		for range quoteChan {
			// Consume quotes
		}
	}()

	// Wait briefly for streaming to start
	time.Sleep(500 * time.Millisecond)

	// Disconnect should be graceful (no panic)
	err = adapter.Disconnect()
	if err != nil {
		t.Errorf("disconnect failed: %v", err)
	}

	// Verify disconnected
	if adapter.IsConnected() {
		t.Error("adapter should report disconnected after Disconnect()")
	}
}
