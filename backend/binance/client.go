package binance

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Binance combined stream endpoint for book ticker (best bid/ask)
	BinanceWSURL = "wss://stream.binance.com:9443/stream?streams=btcusdt@bookTicker/ethusdt@bookTicker"
)

// Quote represents a price quote from Binance
type Quote struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Timestamp int64   `json:"timestamp"`
}

// BookTickerEvent represents Binance book ticker (best bid/ask)
type BookTickerEvent struct {
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

// StreamMessage wraps the stream data
type StreamMessage struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// Client handles Binance WebSocket connection
type Client struct {
	conn       *websocket.Conn
	quotesChan chan Quote
	stopChan   chan struct{}
	mu         sync.RWMutex
	connected  bool
	symbols    []string
}

// NewClient creates a new Binance client
func NewClient() *Client {
	return &Client{
		quotesChan: make(chan Quote, 100),
		stopChan:   make(chan struct{}),
		symbols:    []string{"BTCUSD", "ETHUSD"},
	}
}

// Connect establishes WebSocket connection to Binance
func (c *Client) Connect(symbols []string) error {
	c.symbols = symbols

	log.Printf("[Binance] Connecting to %s...", BinanceWSURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(BinanceWSURL, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.mu.Lock()
	c.connected = true
	c.mu.Unlock()

	log.Println("[Binance] WebSocket connected successfully")

	// Start reading messages
	go c.readMessages()

	// Start heartbeat (Binance expects pong within 10 minutes)
	go c.heartbeat()

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
			log.Printf("[Binance] Read error: %v", err)
			go c.reconnect()
			return
		}

		c.handleMessage(message)
	}
}

func (c *Client) handleMessage(message []byte) {
	// Parse combined stream message
	var streamMsg StreamMessage
	if err := json.Unmarshal(message, &streamMsg); err != nil {
		return
	}

	// Parse book ticker data
	var ticker BookTickerEvent
	if err := json.Unmarshal(streamMsg.Data, &ticker); err != nil {
		return
	}

	if ticker.Symbol == "" || ticker.BidPrice == "" || ticker.AskPrice == "" {
		return
	}

	// Parse bid/ask prices
	bid, err := strconv.ParseFloat(ticker.BidPrice, 64)
	if err != nil {
		return
	}
	ask, err := strconv.ParseFloat(ticker.AskPrice, 64)
	if err != nil {
		return
	}

	if bid <= 0 || ask <= 0 {
		return
	}

	// Convert symbol: BTCUSDT -> BTCUSD
	symbol := strings.Replace(ticker.Symbol, "USDT", "USD", 1)

	quote := Quote{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now().UnixMilli(),
	}

	log.Printf("[Binance] TICK: %s bid=%.2f ask=%.2f", symbol, bid, ask)

	select {
	case c.quotesChan <- quote:
	default:
	}
}

func (c *Client) heartbeat() {
	ticker := time.NewTicker(3 * time.Minute) // Binance requires ping every 10 min, we do 3
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
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("[Binance] Ping failed: %v", err)
				}
			}
		}
	}
}

func (c *Client) reconnect() {
	log.Println("[Binance] Attempting reconnect in 3 seconds...")
	time.Sleep(3 * time.Second)

	if err := c.Connect(c.symbols); err != nil {
		log.Printf("[Binance] Reconnect failed: %v", err)
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
