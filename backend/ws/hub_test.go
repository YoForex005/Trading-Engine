package ws

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/logging"
	"github.com/gorilla/websocket"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests
	logging.Init(slog.LevelInfo)
	os.Exit(m.Run())
}

func TestWebSocketHub_SingleClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Wait for client to register
	time.Sleep(50 * time.Millisecond)

	// Broadcast message
	testMsg := []byte(`{"type":"tick","symbol":"EURUSD","bid":1.0850,"ask":1.0852}`)
	select {
	case hub.broadcast <- testMsg:
	default:
		t.Fatal("failed to send to broadcast channel")
	}

	// Verify client receives message
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, received, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if !bytes.Equal(received, testMsg) {
		t.Errorf("got %s, want %s", received, testMsg)
	}
}

func TestWebSocketHub_MultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect multiple clients
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	numClients := 5
	clients := make([]*websocket.Conn, numClients)

	for i := 0; i < numClients; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("client %d dial failed: %v", i, err)
		}
		defer conn.Close()
		clients[i] = conn
	}

	// Wait for all clients to register
	time.Sleep(100 * time.Millisecond)

	// Broadcast message
	testMsg := []byte(`{"type":"tick","symbol":"BTCUSD","bid":95000,"ask":95010}`)
	select {
	case hub.broadcast <- testMsg:
	default:
		t.Fatal("failed to send to broadcast channel")
	}

	// Verify all clients receive
	for i, client := range clients {
		client.SetReadDeadline(time.Now().Add(time.Second))
		_, received, err := client.ReadMessage()
		if err != nil {
			t.Fatalf("client %d read failed: %v", i, err)
		}
		if !bytes.Equal(received, testMsg) {
			t.Errorf("client %d: got %s, want %s", i, received, testMsg)
		}
	}
}

func TestWebSocketHub_ClientDisconnect(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect two clients
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client1 dial failed: %v", err)
	}
	defer client1.Close()

	client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("client2 dial failed: %v", err)
	}

	// Wait for registration
	time.Sleep(50 * time.Millisecond)

	// Disconnect client2
	client2.Close()
	time.Sleep(50 * time.Millisecond) // Wait for unregister

	// Broadcast message
	testMsg := []byte(`{"type":"tick","symbol":"EURUSD","bid":1.0850}`)
	select {
	case hub.broadcast <- testMsg:
	default:
		t.Fatal("failed to send to broadcast channel")
	}

	// client1 should receive
	client1.SetReadDeadline(time.Now().Add(time.Second))
	_, received, err := client1.ReadMessage()
	if err != nil {
		t.Fatalf("client1 read failed: %v", err)
	}
	if !bytes.Equal(received, testMsg) {
		t.Errorf("client1: got %s, want %s", received, testMsg)
	}

	// Hub should handle client2 disconnect gracefully (no panics)
	// Test passes if we reach here without panic
}

func TestWebSocketHub_RaceCondition(t *testing.T) {
	// Run with: go test -race
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect and disconnect clients concurrently while broadcasting
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	done := make(chan bool, 2)

	// Goroutine: connect/disconnect clients
	go func() {
		for i := 0; i < 10; i++ {
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err == nil {
				time.Sleep(10 * time.Millisecond)
				conn.Close()
			}
		}
		done <- true
	}()

	// Goroutine: broadcast messages
	go func() {
		for i := 0; i < 20; i++ {
			select {
			case hub.broadcast <- []byte(`{"type":"tick"}`):
			default:
			}
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// If no race conditions detected by -race flag, test passes
}

func TestWebSocketHub_BroadcastTick(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Wait for client to register
	time.Sleep(50 * time.Millisecond)

	// Broadcast a tick using the BroadcastTick method
	tick := &MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.0850,
		Ask:       1.0852,
		Spread:    0.0002,
		Timestamp: time.Now().Unix(),
		LP:        "oanda",
	}
	hub.BroadcastTick(tick)

	// Verify client receives tick
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, received, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	// Verify it's valid JSON containing the symbol
	if !bytes.Contains(received, []byte("EURUSD")) {
		t.Errorf("received message does not contain EURUSD: %s", received)
	}
	if !bytes.Contains(received, []byte("1.085")) {
		t.Errorf("received message does not contain bid price: %s", received)
	}
}

func TestWebSocketHub_GetLatestPrice(t *testing.T) {
	hub := NewHub()

	// Broadcast a tick
	tick := &MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.0850,
		Ask:       1.0852,
		Timestamp: time.Now().Unix(),
		LP:        "oanda",
	}

	// Manually set latest price (simulating BroadcastTick behavior)
	hub.mu.Lock()
	hub.latestPrices["EURUSD"] = tick
	hub.mu.Unlock()

	// Get latest price
	latest := hub.GetLatestPrice("EURUSD")
	if latest == nil {
		t.Fatal("expected latest price, got nil")
	}

	if latest.Symbol != "EURUSD" {
		t.Errorf("got symbol %s, want EURUSD", latest.Symbol)
	}
	if latest.Bid != 1.0850 {
		t.Errorf("got bid %.4f, want 1.0850", latest.Bid)
	}

	// Get non-existent symbol
	nonExistent := hub.GetLatestPrice("NONEXISTENT")
	if nonExistent != nil {
		t.Error("expected nil for non-existent symbol")
	}
}

func TestWebSocketHub_LPPriority(t *testing.T) {
	hub := NewHub()

	// Set LP priorities
	hub.SetLPPriority("oanda", 1)   // Higher priority
	hub.SetLPPriority("binance", 2) // Lower priority

	// Broadcast from higher priority LP
	tick1 := &MarketTick{
		Symbol: "BTCUSD",
		Bid:    95000,
		Ask:    95010,
		LP:     "oanda",
	}
	hub.mu.Lock()
	hub.latestPrices["BTCUSD"] = tick1
	hub.mu.Unlock()

	// Try to broadcast from lower priority LP
	tick2 := &MarketTick{
		Symbol: "BTCUSD",
		Bid:    94900, // Different price
		Ask:    94910,
		LP:     "binance",
	}

	// Simulate priority check logic from BroadcastTick
	hub.mu.Lock()
	existingTick := hub.latestPrices["BTCUSD"]
	existingPriority := hub.lpPriority[existingTick.LP]
	newPriority := hub.lpPriority[tick2.LP]

	shouldUpdate := true
	if newPriority > existingPriority {
		shouldUpdate = false
	}
	hub.mu.Unlock()

	if shouldUpdate {
		t.Error("should not update with lower priority LP")
	}

	// Verify original price is still there
	latest := hub.GetLatestPrice("BTCUSD")
	if latest.Bid != 95000 {
		t.Errorf("got bid %.0f, want 95000 (from higher priority LP)", latest.Bid)
	}
}

func TestWebSocketHub_NoClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Broadcast message with no clients connected
	testMsg := []byte(`{"type":"tick","symbol":"EURUSD","bid":1.0850}`)
	select {
	case hub.broadcast <- testMsg:
		// Should not block or panic
	case <-time.After(100 * time.Millisecond):
		t.Fatal("broadcast blocked with no clients")
	}

	// Test passes if no panic
}
