package adapters

import (
	"testing"
	"time"
)

func TestBinanceAdapter_Connect(t *testing.T) {
	t.Skip("Integration test - requires Binance WebSocket access")

	// This demonstrates the pattern for integration testing
	// Binance WebSocket is public, but test demonstrates the pattern

	adapter := NewBinanceAdapter()
	if adapter == nil {
		t.Fatal("failed to create Binance adapter")
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

func TestBinanceAdapter_Reconnect(t *testing.T) {
	t.Skip("Integration test - requires Binance WebSocket")

	adapter := NewBinanceAdapter()

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

func TestBinanceAdapter_QuoteStream(t *testing.T) {
	t.Skip("Integration test - requires Binance WebSocket")

	adapter := NewBinanceAdapter()
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer adapter.Disconnect()

	// Subscribe to quotes (Binance uses different symbol format)
	symbols := []string{"BTCUSDT"}
	err = adapter.Subscribe(symbols)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Wait for at least one quote with timeout
	quoteChan := adapter.GetQuotesChan()
	select {
	case quote := <-quoteChan:
		if quote.Symbol != "BTCUSD" { // Adapter normalizes to standard format
			t.Errorf("got quote for %s, want BTCUSD", quote.Symbol)
		}
		if quote.Bid == 0 || quote.Ask == 0 {
			t.Error("quote has zero bid or ask")
		}
		if quote.LP != "binance" {
			t.Errorf("got LP %s, want binance", quote.LP)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for quote")
	}
}

func TestBinanceAdapter_GetSymbols(t *testing.T) {
	t.Skip("Integration test - requires Binance API")

	adapter := NewBinanceAdapter()
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
		if sym.Type != "crypto" {
			t.Errorf("Binance symbol has type %s, want crypto", sym.Type)
		}
	}
}

func TestBinanceAdapter_MultiSymbolStream(t *testing.T) {
	t.Skip("Integration test - requires Binance WebSocket")

	adapter := NewBinanceAdapter()
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	defer adapter.Disconnect()

	// Subscribe to multiple symbols
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	err = adapter.Subscribe(symbols)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Collect quotes for different symbols
	quoteChan := adapter.GetQuotesChan()
	receivedSymbols := make(map[string]bool)
	timeout := time.After(15 * time.Second)

	for len(receivedSymbols) < 3 {
		select {
		case quote := <-quoteChan:
			receivedSymbols[quote.Symbol] = true
			t.Logf("Received quote for %s: bid=%.2f, ask=%.2f", quote.Symbol, quote.Bid, quote.Ask)
		case <-timeout:
			t.Fatalf("timeout waiting for all symbols, got: %v", receivedSymbols)
		}
	}

	// Verify we got all three symbols
	expectedSymbols := []string{"BTCUSD", "ETHUSD", "BNBUSD"}
	for _, expected := range expectedSymbols {
		if !receivedSymbols[expected] {
			t.Errorf("did not receive quote for %s", expected)
		}
	}
}

func TestBinanceAdapter_DisconnectWhileStreaming(t *testing.T) {
	t.Skip("Integration test - requires Binance WebSocket")

	adapter := NewBinanceAdapter()
	err := adapter.Connect()
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	// Subscribe to quotes
	err = adapter.Subscribe([]string{"BTCUSDT"})
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
