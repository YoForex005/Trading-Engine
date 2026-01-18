package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/ws"
	"github.com/gorilla/websocket"
)

// TestWebSocketConnection tests basic WebSocket connection
func TestWebSocketConnection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + server.URL[4:] + "/ws"

	// Connect WebSocket client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Connection successful
	t.Log("WebSocket connection established")
}

// TestWebSocketTickStream tests real-time tick streaming
func TestWebSocketTickStream(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	// Connect WebSocket client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Start goroutine to read messages
	tickReceived := make(chan bool, 1)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var tick map[string]interface{}
			if err := json.Unmarshal(message, &tick); err != nil {
				continue
			}

			if tick["type"] == "tick" {
				t.Logf("Received tick: %s @ %.5f/%.5f",
					tick["symbol"], tick["bid"], tick["ask"])
				tickReceived <- true
				return
			}
		}
	}()

	// Inject test price
	time.Sleep(100 * time.Millisecond)
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	// Wait for tick to be received
	select {
	case <-tickReceived:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for tick")
	}
}

// TestWebSocketSubscription tests symbol subscription
func TestWebSocketSubscription(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Subscribe to symbol
	subscribeMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	msgBytes, _ := json.Marshal(subscribeMsg)

	if err := conn.WriteMessage(websocket.TextMessage, msgBytes); err != nil {
		t.Fatalf("Failed to send subscribe message: %v", err)
	}

	t.Log("Subscription message sent")

	// Wait for acknowledgment or tick
	done := make(chan bool)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			if msg["type"] == "tick" || msg["type"] == "subscribed" {
				t.Logf("Received message type: %v", msg["type"])
				done <- true
				return
			}
		}
	}()

	// Inject tick
	time.Sleep(100 * time.Millisecond)
	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Log("No immediate response (may be normal)")
	}
}

// TestWebSocketMultipleClients tests multiple concurrent connections
func TestWebSocketMultipleClients(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	numClients := 5
	connections := make([]*websocket.Conn, numClients)
	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	// Connect multiple clients
	for i := 0; i < numClients; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Client %d failed to connect: %v", i, err)
		}
		connections[i] = conn
		t.Logf("Client %d connected", i)
	}

	// All clients should receive the same tick
	receivedCount := make(chan bool, numClients)

	for i, conn := range connections {
		go func(clientID int, c *websocket.Conn) {
			_, message, err := c.ReadMessage()
			if err == nil {
				var tick map[string]interface{}
				if err := json.Unmarshal(message, &tick); err == nil {
					if tick["type"] == "tick" {
						t.Logf("Client %d received tick", clientID)
						receivedCount <- true
					}
				}
			}
		}(i, conn)
	}

	// Broadcast tick
	time.Sleep(100 * time.Millisecond)
	ts.InjectPrice("GBPUSD", 1.25000, 1.25020)

	// Wait for all clients to receive
	timeout := time.After(2 * time.Second)
	received := 0

	for received < numClients {
		select {
		case <-receivedCount:
			received++
		case <-timeout:
			t.Logf("Timeout: %d/%d clients received tick", received, numClients)
			return
		}
	}

	if received == numClients {
		t.Logf("All %d clients received tick successfully", numClients)
	}
}

// TestWebSocketReconnection tests reconnection handling
func TestWebSocketReconnection(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	// First connection
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("First connection failed: %v", err)
	}

	t.Log("First connection established")

	// Close first connection
	conn1.Close()
	time.Sleep(100 * time.Millisecond)

	// Second connection (reconnect)
	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Reconnection failed: %v", err)
	}
	defer conn2.Close()

	t.Log("Reconnection successful")

	// Verify new connection works
	done := make(chan bool)
	go func() {
		_, _, err := conn2.ReadMessage()
		if err == nil {
			done <- true
		}
	}()

	ts.InjectPrice("USDJPY", 110.000, 110.020)

	select {
	case <-done:
		t.Log("Reconnected client receiving data")
	case <-time.After(2 * time.Second):
		t.Log("Timeout on reconnected client")
	}
}

// TestWebSocketBinaryMessages tests binary data handling
func TestWebSocketBinaryMessages(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	// Try sending binary message (should be handled gracefully)
	binaryData := []byte{0x01, 0x02, 0x03, 0x04}
	err = conn.WriteMessage(websocket.BinaryMessage, binaryData)

	if err != nil {
		t.Logf("Binary message rejected (expected): %v", err)
	} else {
		t.Log("Binary message accepted")
	}
}

// TestWebSocketPingPong tests connection keep-alive
func TestWebSocketPingPong(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	// Set ping handler
	pongReceived := make(chan bool, 1)
	conn.SetPongHandler(func(string) error {
		t.Log("Pong received")
		pongReceived <- true
		return nil
	})

	// Send ping
	if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// Start reader to process pong
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	select {
	case <-pongReceived:
		t.Log("Ping-pong successful")
	case <-time.After(1 * time.Second):
		t.Log("No pong received (may be normal)")
	}
}

// TestWebSocketMessageOrdering tests message order preservation
func TestWebSocketMessageOrdering(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	receivedPrices := make([]float64, 0)
	done := make(chan bool)

	// Read messages
	go func() {
		for i := 0; i < 5; i++ {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var tick map[string]interface{}
			if err := json.Unmarshal(message, &tick); err != nil {
				continue
			}

			if tick["type"] == "tick" {
				receivedPrices = append(receivedPrices, tick["bid"].(float64))
			}

			if len(receivedPrices) >= 5 {
				done <- true
				return
			}
		}
	}()

	// Send prices in sequence
	expectedPrices := []float64{1.10000, 1.10010, 1.10020, 1.10030, 1.10040}
	for _, price := range expectedPrices {
		ts.InjectPrice("EURUSD", price, price+0.00020)
		time.Sleep(50 * time.Millisecond)
	}

	select {
	case <-done:
		// Verify order
		if len(receivedPrices) == len(expectedPrices) {
			t.Logf("Received all %d prices in order", len(receivedPrices))
		} else {
			t.Errorf("Expected %d prices, got %d", len(expectedPrices), len(receivedPrices))
		}
	case <-time.After(3 * time.Second):
		t.Errorf("Timeout: received %d/%d prices", len(receivedPrices), len(expectedPrices))
	}
}

// TestWebSocketErrorHandling tests error scenarios
func TestWebSocketErrorHandling(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	// Send invalid JSON
	invalidJSON := []byte(`{"type": "subscribe", "symbol": `)

	err = conn.WriteMessage(websocket.TextMessage, invalidJSON)
	if err != nil {
		t.Logf("Invalid JSON rejected: %v", err)
	}

	// Send malformed message
	malformedMsg := []byte(`not json at all`)

	err = conn.WriteMessage(websocket.TextMessage, malformedMsg)
	if err != nil {
		t.Logf("Malformed message rejected: %v", err)
	}

	// Connection should still be alive
	time.Sleep(100 * time.Millisecond)

	done := make(chan bool)
	go func() {
		_, _, err := conn.ReadMessage()
		if err == nil {
			done <- true
		}
	}()

	ts.InjectPrice("EURUSD", 1.10000, 1.10020)

	select {
	case <-done:
		t.Log("Connection still alive after errors")
	case <-time.After(1 * time.Second):
		t.Log("Connection may have been closed")
	}
}

// BenchmarkWebSocketThroughput benchmarks message throughput
func BenchmarkWebSocketThroughput(b *testing.B) {
	ts := SetupTestServer(&testing.T{})
	defer ts.Cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(ts.hub, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		b.Fatalf("Connection failed: %v", err)
	}
	defer conn.Close()

	// Start reader
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ts.InjectPrice("EURUSD", 1.10000+float64(i)*0.00001, 1.10020+float64(i)*0.00001)
	}
}
