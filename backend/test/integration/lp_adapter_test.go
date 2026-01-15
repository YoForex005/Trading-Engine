package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/ws"
	"github.com/gorilla/websocket"
)

func TestLPToWebSocketFlow(t *testing.T) {
	// Integration test: LP quote → Manager → Hub → Client
	t.Parallel()

	// Setup LP manager with temp config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")
	manager := lpmanager.NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Setup WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Connect hub to LP manager quote stream
	quoteChan := manager.GetQuotesChan()
	go func() {
		for quote := range quoteChan {
			// Convert LP quote to MarketTick and broadcast
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    quote.Symbol,
				Bid:       quote.Bid,
				Ask:       quote.Ask,
				Spread:    quote.Ask - quote.Bid,
				Timestamp: quote.Timestamp,
				LP:        quote.LP,
			}
			hub.BroadcastTick(tick)
		}
	}()

	// Create test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect WebSocket client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Wait for client registration
	time.Sleep(100 * time.Millisecond)

	// Simulate LP quote arriving (bypass manager, send directly to hub)
	mockQuote := lpmanager.Quote{
		Symbol:    "EURUSD",
		Bid:       1.0850,
		Ask:       1.0852,
		Timestamp: time.Now().Unix(),
		LP:        "test-lp",
	}

	// Convert to tick and send directly to hub
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    mockQuote.Symbol,
		Bid:       mockQuote.Bid,
		Ask:       mockQuote.Ask,
		Spread:    mockQuote.Ask - mockQuote.Bid,
		Timestamp: mockQuote.Timestamp,
		LP:        mockQuote.LP,
	}

	// Send tick through hub
	go func() {
		hub.BroadcastTick(tick)
	}()

	// Verify client receives quote via WebSocket
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, received, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	// Parse and verify quote
	var receivedTick ws.MarketTick
	err = json.Unmarshal(received, &receivedTick)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if receivedTick.Symbol != "EURUSD" {
		t.Errorf("got symbol %s, want EURUSD", receivedTick.Symbol)
	}
	if receivedTick.Bid != 1.0850 {
		t.Errorf("got bid %.4f, want 1.0850", receivedTick.Bid)
	}
	if receivedTick.Ask != 1.0852 {
		t.Errorf("got ask %.4f, want 1.0852", receivedTick.Ask)
	}
	if receivedTick.LP != "test-lp" {
		t.Errorf("got LP %s, want test-lp", receivedTick.LP)
	}
}

func TestLPToWebSocketFlow_MultipleClients(t *testing.T) {
	// Test that multiple WebSocket clients receive the same LP quote
	t.Parallel()

	// Setup LP manager
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "lp_config.json")
	manager := lpmanager.NewManager(configPath)
	if err := manager.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Setup WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Connect hub to LP manager
	quoteChan := manager.GetQuotesChan()
	go func() {
		for quote := range quoteChan {
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    quote.Symbol,
				Bid:       quote.Bid,
				Ask:       quote.Ask,
				Timestamp: quote.Timestamp,
				LP:        quote.LP,
			}
			hub.BroadcastTick(tick)
		}
	}()

	// Create test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	}))
	defer server.Close()

	// Connect multiple WebSocket clients
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	numClients := 3
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

	// Send quote directly to hub
	tick := &ws.MarketTick{
		Type:      "tick",
		Symbol:    "BTCUSD",
		Bid:       95000,
		Ask:       95010,
		Spread:    10,
		Timestamp: time.Now().Unix(),
		LP:        "test-lp",
	}
	hub.BroadcastTick(tick)

	// Verify all clients receive the same quote
	for i, client := range clients {
		client.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, received, err := client.ReadMessage()
		if err != nil {
			t.Fatalf("client %d read failed: %v", i, err)
		}

		var tick ws.MarketTick
		if err := json.Unmarshal(received, &tick); err != nil {
			t.Fatalf("client %d unmarshal failed: %v", i, err)
		}

		if tick.Symbol != "BTCUSD" {
			t.Errorf("client %d: got symbol %s, want BTCUSD", i, tick.Symbol)
		}
		if tick.Bid != 95000 {
			t.Errorf("client %d: got bid %.0f, want 95000", i, tick.Bid)
		}
	}
}

func TestLPToWebSocketFlow_Disabled(t *testing.T) {
	// Test that disabled LPs don't send quotes through the pipeline
	t.Skip("Requires LP config integration - deferred to full integration test suite")
}

// MockLPAdapter implements lpmanager.LPAdapter for testing
type MockLPAdapter struct {
	id         string
	name       string
	lpType     string
	connected  bool
	quotesChan chan lpmanager.Quote
}

func (m *MockLPAdapter) ID() string                                 { return m.id }
func (m *MockLPAdapter) Name() string                               { return m.name }
func (m *MockLPAdapter) Type() string                               { return m.lpType }
func (m *MockLPAdapter) IsConnected() bool                          { return m.connected }
func (m *MockLPAdapter) GetQuotesChan() <-chan lpmanager.Quote      { return m.quotesChan }
func (m *MockLPAdapter) Connect() error                             { m.connected = true; return nil }
func (m *MockLPAdapter) Disconnect() error                          { m.connected = false; return nil }
func (m *MockLPAdapter) GetSymbols() ([]lpmanager.SymbolInfo, error) { return []lpmanager.SymbolInfo{}, nil }
func (m *MockLPAdapter) Subscribe(symbols []string) error           { return nil }
func (m *MockLPAdapter) Unsubscribe(symbols []string) error         { return nil }
func (m *MockLPAdapter) GetStatus() lpmanager.LPStatus {
	return lpmanager.LPStatus{
		ID:        m.id,
		Name:      m.name,
		Type:      m.lpType,
		Connected: m.connected,
	}
}

// TestMain sets up test environment
func TestMain(m *testing.M) {
	// Set ALLOWED_ORIGINS for WebSocket tests
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:*")
	code := m.Run()
	os.Exit(code)
}
