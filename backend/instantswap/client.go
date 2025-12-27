package instantswap

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Socket.IO Protocol constants (Engine.IO v4)
const (
	PacketOpen    = '0'
	PacketClose   = '1'
	PacketPing    = '2'
	PacketPong    = '3'
	PacketMessage = '4'
)

// Socket.IO Message types
const (
	MsgConnect    = '0'
	MsgDisconnect = '1'
	MsgEvent      = '2'
	MsgAck        = '3'
	MsgError      = '4'
)

// Price represents a tick from the feed
type Price struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Timestamp int64   `json:"timestamp"`
}

type Client struct {
	conn       *websocket.Conn
	url        string
	pricesChan chan Price
	stopChan   chan struct{}
	mu         sync.Mutex
	connected  bool
}

func NewClient() *Client {
	return &Client{
		url:        "wss://quotes.instantswap.app/socket.io/?EIO=4&transport=websocket",
		pricesChan: make(chan Price, 1000),
		stopChan:   make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return err
	}

	log.Printf("[InstantSwap] Connecting to %s...", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.conn = conn

	// Start connection handler loop
	go c.readLoop()

	// Wait for connection to be established (handled in readLoop)
	// For standard Socket.IO, we don't strictly need to send a Connect packet for the default namespace '/'
	// providing the server accepts implicit connection, but sending "40" is safer.
	// However, simple testing often shows implicit works. Let's send "40" to be sure.
	c.sendMessage("40")

	return nil
}

func (c *Client) Subscribe() {
	// Emit 'join' events as discovered
	log.Println("[InstantSwap] Subscribing to rates...")

	// Socket.IO emit format: 42["event", data]
	// Emit "join": 42["join", "rates"]
	c.emit("join", "rates")
	c.emit("join", "prices")

	// Also emit "get_rates" periodically
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.stopChan:
				return
			case <-ticker.C:
				c.emit("get_rates", nil)
			}
		}
	}()
}

func (c *Client) emit(event string, data interface{}) {
	payload := []interface{}{event}
	if data != nil {
		payload = append(payload, data)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[InstantSwap] Error marshaling emit payload: %v", err)
		return
	}

	msg := fmt.Sprintf("42%s", string(jsonPayload))
	c.sendMessage(msg)
}

func (c *Client) sendMessage(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return
	}

	if err := c.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		log.Printf("[InstantSwap] Write error: %v", err)
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) readLoop() {
	defer func() {
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
			log.Printf("[InstantSwap] Read error: %v", err)

			// Simple reconnect logic
			time.Sleep(2 * time.Second)
			if c.reconnect() != nil {
				time.Sleep(5 * time.Second) // Backoff
			}
			continue
		}

		c.handleMessage(message)
	}
}

func (c *Client) reconnect() error {
	log.Println("[InstantSwap] Reconnecting...")
	u, _ := url.Parse(c.url)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}
	c.conn = conn
	c.sendMessage("40")
	c.Subscribe() // Resubscribe
	return nil
}

func (c *Client) handleMessage(msg []byte) {
	if len(msg) == 0 {
		return
	}

	packetType := msg[0]
	// content := string(msg[1:])

	switch packetType {
	case PacketOpen: // '0'
		// Handshake open
		c.mu.Lock()
		c.connected = true
		c.mu.Unlock()
		log.Println("[InstantSwap] Connected (Open packet received)")

		// Send initial ping if needed, usually client waits for ping from server or just pings periodically
		// For Engine.IO v4, client usually sends Ping(2), server sends Pong(3) in some configs,
		// OR server Pings(2) and client Pongs(3).
		// Probe showed: 2 -> 3. Server sends 2, Client sends 3.

	case PacketPing: // '2'
		// Server ping, must pong back
		c.sendMessage("3") // Pong

	case PacketMessage: // '4'
		// c.handleSocketIOMessage(content)
		// For performance, direct string handling
		strMsg := string(msg)
		if len(strMsg) < 2 {
			return
		}

		msgType := strMsg[1]
		if msgType == MsgEvent { // '2' -> "42"
			jsonPart := strMsg[2:]
			var payload []interface{}
			if err := json.Unmarshal([]byte(jsonPart), &payload); err == nil && len(payload) > 0 {
				event, ok := payload[0].(string)
				if ok && len(payload) > 1 {
					// Log the event name once
					// log.Printf("[InstantSwap] Received event: %s", event)
					// We received data
					data := payload[1]
					c.processData(event, data)
				} else if ok {
					// Event without data
					// log.Printf("[InstantSwap] Received event (no data): %s", event)
				}
			}
		}
	}
}

func (c *Client) processData(event string, data interface{}) {
	// log.Printf("[DEBUG] Processing event: %s", event)

	// Only process relevant events
	if event != "quotes" && event != "rates" {
		return
	}

	updates, ok := data.([]interface{})
	if !ok {
		// Maybe it's a single object
		// check if map
		return
	}

	for _, item := range updates {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Map fields based on assumptions or typical json keys
		// Need to be flexible.
		// "Symbol": "XAUUSD"
		// "Bid": "2032.50"
		// "Ask": "2033.10"

		sym, _ := itemMap["Symbol"].(string)
		if sym == "" {
			continue
		}

		// Normalize Symbol: EURUSD, BTCUSD, etc.

		var bid, ask float64

		// Parse Bid
		if v, ok := itemMap["Bid"].(string); ok {
			fmt.Sscanf(v, "%f", &bid)
		} else if v, ok := itemMap["Bid"].(float64); ok {
			bid = v
		}

		// Parse Ask
		if v, ok := itemMap["Ask"].(string); ok {
			fmt.Sscanf(v, "%f", &ask)
		} else if v, ok := itemMap["Ask"].(float64); ok {
			ask = v
		}

		if bid > 0 && ask > 0 {
			c.pricesChan <- Price{
				Symbol:    sym,
				Bid:       bid,
				Ask:       ask,
				Timestamp: time.Now().UnixMilli(),
			}
		}
	}
}

func (c *Client) GetPricesChan() <-chan Price {
	return c.pricesChan
}
