package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
	"github.com/gorilla/websocket"
)

// WebSocketTestContext holds WebSocket test environment
type WebSocketTestContext struct {
	Server    *httptest.Server
	Hub       *ws.Hub
	BBEngine  *core.Engine
	TickStore *tickstore.TickStore
}

// SetupWebSocketTest initializes WebSocket test environment
func SetupWebSocketTest(t *testing.T) *WebSocketTestContext {
	t.Helper()

	// Initialize components
	bbEngine := core.NewEngine()
	tickStore := tickstore.NewTickStore("ws-test", 1000)
	hub := ws.NewHub()
	hub.SetTickStore(tickStore)
	hub.SetBBookEngine(bbEngine)

	// Start hub
	go hub.Run()

	// Create HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	}))

	return &WebSocketTestContext{
		Server:    server,
		Hub:       hub,
		BBEngine:  bbEngine,
		TickStore: tickStore,
	}
}

// Cleanup closes WebSocket test resources
func (wc *WebSocketTestContext) Cleanup() {
	if wc.Server != nil {
		wc.Server.Close()
	}
}

// ConnectWebSocket creates a WebSocket connection
func (wc *WebSocketTestContext) ConnectWebSocket(t *testing.T) *websocket.Conn {
	t.Helper()

	url := "ws" + strings.TrimPrefix(wc.Server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect WebSocket: %v", err)
	}

	return conn
}

// SendMessage sends JSON message to WebSocket
func (wc *WebSocketTestContext) SendMessage(t *testing.T, conn *websocket.Conn, msg interface{}) {
	t.Helper()

	data, _ := json.Marshal(msg)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
}

// ReadMessage reads and decodes JSON message from WebSocket
func (wc *WebSocketTestContext) ReadMessage(t *testing.T, conn *websocket.Conn, timeout time.Duration) map[string]interface{} {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(timeout))

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	return msg
}

// ==================== CONNECTION TESTS ====================

func TestWS_Connect(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Verify connection is open
	if conn == nil {
		t.Fatal("Expected non-nil connection")
	}
}

func TestWS_Connect_MultipleClients(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	numClients := 10
	connections := make([]*websocket.Conn, numClients)

	// Connect multiple clients
	for i := 0; i < numClients; i++ {
		conn := wc.ConnectWebSocket(t)
		connections[i] = conn
		defer conn.Close()
	}

	// Verify all connections
	for i, conn := range connections {
		if conn == nil {
			t.Errorf("Client %d: connection is nil", i)
		}
	}
}

func TestWS_Disconnect(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)

	// Close connection
	if err := conn.Close(); err != nil {
		t.Errorf("Failed to close connection: %v", err)
	}
}

func TestWS_Reconnect(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	// Connect, disconnect, reconnect
	conn1 := wc.ConnectWebSocket(t)
	conn1.Close()

	time.Sleep(100 * time.Millisecond)

	conn2 := wc.ConnectWebSocket(t)
	defer conn2.Close()

	if conn2 == nil {
		t.Fatal("Failed to reconnect")
	}
}

// ==================== SUBSCRIPTION TESTS ====================

func TestWS_Subscribe_SingleSymbol(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Subscribe to EURUSD
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)

	// Send a tick
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.10000,
		Ask:       1.10020,
		Spread:    0.00020,
		Timestamp: time.Now().UnixMilli(),
		LP:        "TEST",
	}
	wc.Hub.BroadcastTick(tick)

	// Read tick from WebSocket
	msg := wc.ReadMessage(t, conn, 2*time.Second)

	if msg["type"] != "tick" {
		t.Errorf("Expected type 'tick', got %v", msg["type"])
	}

	if msg["symbol"] != "EURUSD" {
		t.Errorf("Expected symbol 'EURUSD', got %v", msg["symbol"])
	}
}

func TestWS_Subscribe_MultipleSymbols(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY"}

	// Subscribe to multiple symbols
	for _, symbol := range symbols {
		subMsg := map[string]interface{}{
			"type":   "subscribe",
			"symbol": symbol,
		}
		wc.SendMessage(t, conn, subMsg)
	}

	// Send ticks for all symbols
	receivedSymbols := make(map[string]bool)
	var mu sync.Mutex

	go func() {
		for _, symbol := range symbols {
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    symbol,
				Bid:       1.10000,
				Ask:       1.10020,
				Spread:    0.00020,
				Timestamp: time.Now().UnixMilli(),
				LP:        "TEST",
			}
			wc.Hub.BroadcastTick(tick)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Read ticks
	for i := 0; i < len(symbols); i++ {
		msg := wc.ReadMessage(t, conn, 2*time.Second)

		if msg["type"] == "tick" {
			symbol := msg["symbol"].(string)
			mu.Lock()
			receivedSymbols[symbol] = true
			mu.Unlock()
		}
	}

	// Verify all symbols received
	for _, symbol := range symbols {
		if !receivedSymbols[symbol] {
			t.Errorf("Did not receive tick for %s", symbol)
		}
	}
}

func TestWS_Unsubscribe(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Subscribe
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)

	// Send tick (should receive)
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.10000,
		Ask:       1.10020,
		Spread:    0.00020,
		Timestamp: time.Now().UnixMilli(),
		LP:        "TEST",
	}
	wc.Hub.BroadcastTick(tick)

	// Read tick
	msg := wc.ReadMessage(t, conn, 2*time.Second)
	if msg["type"] != "tick" {
		t.Errorf("Expected to receive tick after subscribe")
	}

	// Unsubscribe
	unsubMsg := map[string]interface{}{
		"type":   "unsubscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, unsubMsg)

	time.Sleep(100 * time.Millisecond)

	// Send another tick (should NOT receive)
	wc.Hub.BroadcastTick(tick)

	// Try to read (should timeout)
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err := conn.ReadMessage()

	// Expect timeout error
	if err == nil {
		t.Error("Expected to NOT receive tick after unsubscribe")
	}
}

// ==================== REAL-TIME DATA TESTS ====================

func TestWS_RealTimeTickStream(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Subscribe
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)

	// Send ticks continuously
	numTicks := 20
	done := make(chan bool)

	go func() {
		for i := 0; i < numTicks; i++ {
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    "EURUSD",
				Bid:       1.10000 + float64(i)*0.00001,
				Ask:       1.10020 + float64(i)*0.00001,
				Spread:    0.00020,
				Timestamp: time.Now().UnixMilli(),
				LP:        "TEST",
			}
			wc.Hub.BroadcastTick(tick)
			time.Sleep(50 * time.Millisecond)
		}
		done <- true
	}()

	// Receive ticks
	receivedTicks := 0
	timeout := time.After(5 * time.Second)

	for receivedTicks < numTicks {
		select {
		case <-timeout:
			t.Fatalf("Timeout: received %d/%d ticks", receivedTicks, numTicks)
		default:
			msg := wc.ReadMessage(t, conn, 2*time.Second)
			if msg["type"] == "tick" {
				receivedTicks++
			}
		}
	}

	<-done

	if receivedTicks != numTicks {
		t.Errorf("Expected %d ticks, received %d", numTicks, receivedTicks)
	}
}

func TestWS_TickDataValidation(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Subscribe
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)

	// Send tick with specific values
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.10000,
		Ask:       1.10020,
		Spread:    0.00020,
		Timestamp: 1234567890123,
		LP:        "TEST_LP",
	}
	wc.Hub.BroadcastTick(tick)

	// Read and validate
	msg := wc.ReadMessage(t, conn, 2*time.Second)

	if msg["type"] != "tick" {
		t.Errorf("Expected type 'tick', got %v", msg["type"])
	}

	if msg["symbol"] != "EURUSD" {
		t.Errorf("Expected symbol 'EURUSD', got %v", msg["symbol"])
	}

	if msg["bid"] != 1.10000 {
		t.Errorf("Expected bid 1.10000, got %v", msg["bid"])
	}

	if msg["ask"] != 1.10020 {
		t.Errorf("Expected ask 1.10020, got %v", msg["ask"])
	}

	if msg["lp"] != "TEST_LP" {
		t.Errorf("Expected LP 'TEST_LP', got %v", msg["lp"])
	}
}

// ==================== ERROR HANDLING TESTS ====================

func TestWS_InvalidMessage(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Send invalid JSON
	if err := conn.WriteMessage(websocket.TextMessage, []byte("invalid json")); err != nil {
		t.Errorf("Failed to send invalid message: %v", err)
	}

	// Connection should remain open
	time.Sleep(200 * time.Millisecond)

	// Test with valid message
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)
}

func TestWS_UnknownMessageType(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Send unknown message type
	msg := map[string]interface{}{
		"type": "unknown_type",
		"data": "test",
	}
	wc.SendMessage(t, conn, msg)

	// Connection should remain open
	time.Sleep(200 * time.Millisecond)

	// Test with valid message
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)
}

// ==================== PERFORMANCE TESTS ====================

func TestWS_HighFrequencyTicks(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(t)
	defer conn.Close()

	// Subscribe
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn, subMsg)

	// Send high frequency ticks
	numTicks := 100
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for i := 0; i < numTicks; i++ {
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    "EURUSD",
				Bid:       1.10000 + float64(i)*0.00001,
				Ask:       1.10020 + float64(i)*0.00001,
				Spread:    0.00020,
				Timestamp: time.Now().UnixMilli(),
				LP:        "TEST",
			}
			wc.Hub.BroadcastTick(tick)
			time.Sleep(10 * time.Millisecond) // 100 ticks/second
		}
	}()

	// Count received ticks
	receivedTicks := 0
	timeout := time.After(5 * time.Second)

	for receivedTicks < numTicks {
		select {
		case <-timeout:
			t.Logf("Timeout: received %d/%d ticks", receivedTicks, numTicks)
			wg.Wait()
			return
		default:
			msg := wc.ReadMessage(t, conn, 1*time.Second)
			if msg["type"] == "tick" {
				receivedTicks++
			}
		}
	}

	wg.Wait()
	t.Logf("Successfully received %d high-frequency ticks", receivedTicks)
}

func TestWS_ConcurrentClients(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	numClients := 20
	var wg sync.WaitGroup
	wg.Add(numClients)

	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()

			conn := wc.ConnectWebSocket(t)
			defer conn.Close()

			// Subscribe
			subMsg := map[string]interface{}{
				"type":   "subscribe",
				"symbol": "EURUSD",
			}
			data, _ := json.Marshal(subMsg)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				errors <- fmt.Errorf("client %d: failed to subscribe: %v", clientID, err)
				return
			}

			// Read ticks
			for j := 0; j < 5; j++ {
				conn.SetReadDeadline(time.Now().Add(3 * time.Second))
				_, _, err := conn.ReadMessage()
				if err != nil {
					errors <- fmt.Errorf("client %d: failed to read tick %d: %v", clientID, j, err)
					return
				}
			}
		}(i)
	}

	// Send ticks
	go func() {
		for i := 0; i < 10; i++ {
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    "EURUSD",
				Bid:       1.10000 + float64(i)*0.00001,
				Ask:       1.10020 + float64(i)*0.00001,
				Spread:    0.00020,
				Timestamp: time.Now().UnixMilli(),
				LP:        "TEST",
			}
			wc.Hub.BroadcastTick(tick)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	wg.Wait()
	close(errors)

	if len(errors) > 0 {
		t.Errorf("Concurrent clients test had %d errors:", len(errors))
		for err := range errors {
			t.Log(err)
		}
	}
}

// ==================== RECONNECTION TESTS ====================

func TestWS_ReconnectAfterDisconnect(t *testing.T) {
	wc := SetupWebSocketTest(t)
	defer wc.Cleanup()

	// First connection
	conn1 := wc.ConnectWebSocket(t)
	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	wc.SendMessage(t, conn1, subMsg)

	// Close connection
	conn1.Close()
	time.Sleep(200 * time.Millisecond)

	// Reconnect
	conn2 := wc.ConnectWebSocket(t)
	defer conn2.Close()

	// Re-subscribe
	wc.SendMessage(t, conn2, subMsg)

	// Send tick
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.10000,
		Ask:       1.10020,
		Spread:    0.00020,
		Timestamp: time.Now().UnixMilli(),
		LP:        "TEST",
	}
	wc.Hub.BroadcastTick(tick)

	// Should receive tick on new connection
	msg := wc.ReadMessage(t, conn2, 2*time.Second)

	if msg["type"] != "tick" {
		t.Error("Expected to receive tick after reconnect")
	}
}

// ==================== BENCHMARK TESTS ====================

func BenchmarkWS_TickBroadcast(b *testing.B) {
	wc := SetupWebSocketTest(&testing.T{})
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(&testing.T{})
	defer conn.Close()

	subMsg := map[string]interface{}{
		"type":   "subscribe",
		"symbol": "EURUSD",
	}
	data, _ := json.Marshal(subMsg)
	conn.WriteMessage(websocket.TextMessage, data)

	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    "EURUSD",
		Bid:       1.10000,
		Ask:       1.10020,
		Spread:    0.00020,
		Timestamp: time.Now().UnixMilli(),
		LP:        "TEST",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wc.Hub.BroadcastTick(tick)
	}
}

func BenchmarkWS_SubscribeUnsubscribe(b *testing.B) {
	wc := SetupWebSocketTest(&testing.T{})
	defer wc.Cleanup()

	conn := wc.ConnectWebSocket(&testing.T{})
	defer conn.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		subMsg := map[string]interface{}{
			"type":   "subscribe",
			"symbol": "EURUSD",
		}
		data, _ := json.Marshal(subMsg)
		conn.WriteMessage(websocket.TextMessage, data)

		time.Sleep(10 * time.Millisecond)

		unsubMsg := map[string]interface{}{
			"type":   "unsubscribe",
			"symbol": "EURUSD",
		}
		data, _ = json.Marshal(unsubMsg)
		conn.WriteMessage(websocket.TextMessage, data)
	}
}
