package flexymarkets

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Socket.IO endpoint
	BaseURL = "https://quotes.instantswap.app"
)

// Quote represents a price quote from FlexyMarkets
type Quote struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Timestamp int64   `json:"timestamp"`
}

// Client handles FlexyMarkets Socket.IO connection
type Client struct {
	conn       *websocket.Conn
	quotesChan chan Quote
	stopChan   chan struct{}
	mu         sync.RWMutex
	connected  bool
	symbols    []string
	sid        string
}

// NewClient creates a new FlexyMarkets client
func NewClient() *Client {
	return &Client{
		quotesChan: make(chan Quote, 100),
		stopChan:   make(chan struct{}),
		symbols:    []string{},
	}
}

// handshake performs Socket.IO handshake via polling
func (c *Client) handshake() (string, error) {
	handshakeURL := fmt.Sprintf("%s/socket.io/?EIO=4&transport=polling", BaseURL)

	resp, err := http.Get(handshakeURL)
	if err != nil {
		return "", fmt.Errorf("handshake failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read handshake: %w", err)
	}

	// Socket.IO response starts with a packet type code (e.g., "0{...}")
	bodyStr := string(body)
	if len(bodyStr) < 2 || bodyStr[0] != '0' {
		return "", fmt.Errorf("unexpected handshake response: %s", bodyStr)
	}

	// Parse JSON after the packet type
	var handshakeData struct {
		Sid          string   `json:"sid"`
		Upgrades     []string `json:"upgrades"`
		PingInterval int      `json:"pingInterval"`
		PingTimeout  int      `json:"pingTimeout"`
	}

	if err := json.Unmarshal([]byte(bodyStr[1:]), &handshakeData); err != nil {
		return "", fmt.Errorf("failed to parse handshake: %w", err)
	}

	log.Printf("[FlexyMarkets] Handshake successful, sid=%s", handshakeData.Sid)
	return handshakeData.Sid, nil
}

// Connect establishes Socket.IO connection to FlexyMarkets
func (c *Client) Connect(symbols []string) error {
	c.symbols = symbols

	log.Printf("[FlexyMarkets] Connecting to %s...", BaseURL)

	// Step 1: HTTP handshake
	sid, err := c.handshake()
	if err != nil {
		return err
	}
	c.sid = sid

	// Step 2: Upgrade to WebSocket
	wsURL := fmt.Sprintf("wss://quotes.instantswap.app/socket.io/?EIO=4&transport=websocket&sid=%s", url.QueryEscape(sid))

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	c.conn = conn
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	log.Println("[FlexyMarkets] WebSocket connected, sending probe...")

	// Step 3: Send probe and wait for response
	if err := conn.WriteMessage(websocket.TextMessage, []byte("2probe")); err != nil {
		return fmt.Errorf("failed to send probe: %w", err)
	}

	// Wait for 3probe response
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read probe response: %w", err)
	}

	if string(msg) != "3probe" {
		log.Printf("[FlexyMarkets] Expected 3probe, got: %s", string(msg))
	} else {
		log.Println("[FlexyMarkets] Probe successful, sending upgrade...")
	}

	// Step 4: Send upgrade confirmation (packet type 5)
	if err := conn.WriteMessage(websocket.TextMessage, []byte("5")); err != nil {
		return fmt.Errorf("failed to send upgrade: %w", err)
	}

	// Step 5: Send Socket.IO connect to default namespace
	if err := conn.WriteMessage(websocket.TextMessage, []byte("40")); err != nil {
		log.Printf("[FlexyMarkets] Failed to send namespace connect: %v", err)
	}

	log.Println("[FlexyMarkets] Upgrade complete, now listening for quotes...")

	// Start reading messages
	go c.readMessages()

	// Start heartbeat immediately (ping every 20 seconds to be safe)
	go c.heartbeat()

	// Send initial ping right away
	go func() {
		time.Sleep(1 * time.Second)
		if c.conn != nil {
			c.conn.WriteMessage(websocket.TextMessage, []byte("2"))
		}
	}()

	return nil
}

func (c *Client) readMessages() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		if c.conn != nil {
			c.conn.Close()
		}
	}()

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("[FlexyMarkets] Read error: %v", err)
			go c.reconnect()
			return
		}

		msgStr := string(message)

		// Debug: Log all messages
		if len(msgStr) > 0 && msgStr[0] == '4' {
			log.Printf("[FlexyMarkets] Received: %s", msgStr[:min(100, len(msgStr))])
		}

		// Socket.IO packet types:
		// 0 - open, 2 - ping, 3 - pong, 4 - message
		// 40 - Socket.IO connect, 42 - event

		if msgStr == "3" {
			// Pong response
			continue
		}

		if strings.HasPrefix(msgStr, "40") {
			// Socket.IO connect confirmation
			log.Println("[FlexyMarkets] Socket.IO connected")
			continue
		}

		if strings.HasPrefix(msgStr, "42") {
			// Event message: 42["eventName", data]
			c.handleEvent(msgStr[2:])
		}

		// Also try parsing raw JSON for quote data
		if strings.HasPrefix(msgStr, "4") && len(msgStr) > 2 {
			c.handleEvent(msgStr[2:])
		}
	}
}

func (c *Client) handleEvent(eventJSON string) {
	// Try to parse as array: ["eventName", data]
	var event []json.RawMessage
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		// Try as direct object
		c.tryParseQuote(eventJSON)
		return
	}

	if len(event) < 1 {
		return
	}

	// Get event name
	var eventName string
	json.Unmarshal(event[0], &eventName)

	log.Printf("[FlexyMarkets] Event: %s", eventName)

	// Handle symbolChart event (OHLC data)
	if eventName == "symbolChart" && len(event) >= 2 {
		c.handleSymbolChart(string(event[1]))
		return
	}

	// Handle quotes event (real-time ticks)
	if eventName == "quotes" && len(event) >= 2 {
		c.handleQuotes(string(event[1]))
		return
	}

	if len(event) >= 2 {
		c.tryParseQuote(string(event[1]))
	}
}

// handleQuotes parses the quotes event with proper tick data
func (c *Client) handleQuotes(data string) {
	// Format: [{"Symbol":"BTCUSD","Bid":"90000","Ask":"90050",...}, ...]
	var quotes []struct {
		Symbol      string `json:"Symbol"`
		Bid         string `json:"Bid"`
		Ask         string `json:"Ask"`
		Digits      string `json:"Digits"`
		Datetime    string `json:"Datetime"`
		DatetimeMsc string `json:"DatetimeMsc"`
	}

	if err := json.Unmarshal([]byte(data), &quotes); err != nil {
		log.Printf("[FlexyMarkets] Failed to parse quotes: %v", err)
		return
	}

	for _, q := range quotes {
		// Filter for crypto symbols
		sym := strings.Replace(q.Symbol, "_", "", -1)
		if sym != "BTCUSD" && sym != "ETHUSD" {
			continue
		}

		bid := parseFloat(q.Bid)
		ask := parseFloat(q.Ask)

		if bid <= 0 || ask <= 0 {
			continue
		}

		quote := Quote{
			Symbol:    sym,
			Bid:       bid,
			Ask:       ask,
			Timestamp: time.Now().UnixMilli(),
		}

		log.Printf("[FlexyMarkets] TICK: %s bid=%.2f ask=%.2f", sym, bid, ask)

		select {
		case c.quotesChan <- quote:
		default:
		}
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// handleSymbolChart parses OHLC candle data and converts to quote
func (c *Client) handleSymbolChart(data string) {
	// Format: [[timestamp, open, high, low, close], ...]
	var candles [][]float64
	if err := json.Unmarshal([]byte(data), &candles); err != nil {
		log.Printf("[FlexyMarkets] Failed to parse symbolChart: %v", err)
		return
	}

	if len(candles) == 0 {
		return
	}

	// Use the last candle for the current price
	lastCandle := candles[len(candles)-1]
	if len(lastCandle) < 5 {
		return
	}

	// Extract OHLC: [timestamp, open, high, low, close]
	closePrice := lastCandle[4]

	// Determine symbol based on price range
	var symbol string
	if closePrice > 10000 {
		symbol = "BTCUSD"
	} else if closePrice > 1000 {
		symbol = "ETHUSD"
	} else {
		// Unknown, skip
		return
	}

	// Create bid/ask from close price (add small spread)
	spread := closePrice * 0.0005 // 0.05% spread
	bid := closePrice - spread/2
	ask := closePrice + spread/2

	quote := Quote{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: int64(lastCandle[0]) * 1000, // Convert to milliseconds
	}

	log.Printf("[FlexyMarkets] Quote from chart: %s bid=%.2f ask=%.2f", symbol, bid, ask)

	select {
	case c.quotesChan <- quote:
	default:
	}
}

func (c *Client) tryParseQuote(data string) {
	// Try multiple quote formats
	var quote Quote

	// Format 1: {"symbol": "BTCUSD", "bid": 90000, "ask": 90050}
	if err := json.Unmarshal([]byte(data), &quote); err == nil && quote.Symbol != "" {
		if quote.Timestamp == 0 {
			quote.Timestamp = time.Now().UnixMilli()
		}
		c.sendQuote(quote)
		return
	}

	// Format 2: {"s": "BTCUSD", "b": 90000, "a": 90050}
	var shortQuote struct {
		S string  `json:"s"`
		B float64 `json:"b"`
		A float64 `json:"a"`
	}
	if err := json.Unmarshal([]byte(data), &shortQuote); err == nil && shortQuote.S != "" {
		c.sendQuote(Quote{
			Symbol:    shortQuote.S,
			Bid:       shortQuote.B,
			Ask:       shortQuote.A,
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// Format 3: Array of quotes
	var quotes []Quote
	if err := json.Unmarshal([]byte(data), &quotes); err == nil {
		for _, q := range quotes {
			if q.Symbol != "" {
				if q.Timestamp == 0 {
					q.Timestamp = time.Now().UnixMilli()
				}
				c.sendQuote(q)
			}
		}
	}
}

func (c *Client) sendQuote(quote Quote) {
	// Filter for our symbols
	symbolClean := strings.Replace(quote.Symbol, "_", "", -1)
	for _, sym := range c.symbols {
		symClean := strings.Replace(sym, "_", "", -1)
		if symbolClean == symClean || quote.Symbol == sym {
			log.Printf("[FlexyMarkets] Quote: %s bid=%.2f ask=%.2f", quote.Symbol, quote.Bid, quote.Ask)
			select {
			case c.quotesChan <- quote:
			default:
			}
			return
		}
	}
}

func (c *Client) heartbeat() {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.mu.RLock()
			connected := c.connected
			c.mu.RUnlock()

			if connected && c.conn != nil {
				// Socket.IO ping (packet type 2)
				if err := c.conn.WriteMessage(websocket.TextMessage, []byte("2")); err != nil {
					log.Printf("[FlexyMarkets] Heartbeat failed: %v", err)
				}
			}
		}
	}
}

func (c *Client) reconnect() {
	log.Println("[FlexyMarkets] Attempting reconnect in 5 seconds...")
	time.Sleep(5 * time.Second)

	if err := c.Connect(c.symbols); err != nil {
		log.Printf("[FlexyMarkets] Reconnect failed: %v", err)
		go c.reconnect()
	}
}

// GetQuotesChan returns the channel for receiving quotes
func (c *Client) GetQuotesChan() <-chan Quote {
	return c.quotesChan
}

// IsConnected checks if WebSocket is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Stop closes the WebSocket connection
func (c *Client) Stop() {
	close(c.stopChan)
	if c.conn != nil {
		c.conn.Close()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
