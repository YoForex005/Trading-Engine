package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/gorilla/websocket"
)

const (
	// Message batching configuration
	maxBatchSize     = 50               // Batch 50 updates
	batchInterval    = 16 * time.Millisecond // 60 FPS target (1000ms / 60 = 16.67ms)

	// Buffer sizes
	writeBufferSize  = 4096
	readBufferSize   = 4096
	sendChannelSize  = 256

	// Backpressure configuration
	maxQueuedBatches = 10               // Drop old messages if client falls behind

	// Rate limiting
	maxMessagesPerSecond = 1000
	rateLimitWindow      = time.Second

	// Client timeouts
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 // Max message size for incoming messages
)

// Channel types for subscription management
const (
	ChannelRoutingMetrics  = "routing-metrics"
	ChannelLPPerformance   = "lp-performance"
	ChannelExposureUpdates = "exposure-updates"
	ChannelAlerts          = "alerts"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Configure allowed origins in production
	},
}

// AnalyticsMessage represents a message sent to clients
type AnalyticsMessage struct {
	Type      string      `json:"type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// SubscriptionMessage represents a client subscription request
type SubscriptionMessage struct {
	Action   string   `json:"action"` // "subscribe" or "unsubscribe"
	Channels []string `json:"channels"`
}

// AnalyticsClient represents a connected WebSocket client with analytics subscriptions
type AnalyticsClient struct {
	hub         *AnalyticsHub
	conn        *websocket.Conn
	send        chan []AnalyticsMessage // Batched messages
	userID      string
	accountID   string
	role        string

	// Subscription management
	mu            sync.RWMutex
	subscriptions map[string]bool

	// Rate limiting
	rateLimiter   *rateLimiter

	// Lifecycle
	closeChan     chan struct{}
	closeOnce     sync.Once
}

// rateLimiter implements a simple token bucket rate limiter
type rateLimiter struct {
	mu            sync.Mutex
	tokens        int
	maxTokens     int
	lastRefill    time.Time
	refillRate    time.Duration
}

func newRateLimiter(maxTokens int, refillRate time.Duration) *rateLimiter {
	return &rateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		lastRefill: time.Now(),
		refillRate: refillRate,
	}
}

func (rl *rateLimiter) allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed / rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens = min(rl.maxTokens, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// AnalyticsHub manages all analytics WebSocket connections
type AnalyticsHub struct {
	clients     map[*AnalyticsClient]bool
	register    chan *AnalyticsClient
	unregister  chan *AnalyticsClient
	broadcast   chan *BroadcastMessage
	authService *auth.Service

	mu          sync.RWMutex
	running     bool
	stopChan    chan struct{}
}

// BroadcastMessage represents a message to broadcast to clients
type BroadcastMessage struct {
	Channel string
	Message AnalyticsMessage
}

// NewAnalyticsHub creates a new analytics WebSocket hub
func NewAnalyticsHub(authService *auth.Service) *AnalyticsHub {
	return &AnalyticsHub{
		clients:     make(map[*AnalyticsClient]bool),
		register:    make(chan *AnalyticsClient),
		unregister:  make(chan *AnalyticsClient),
		broadcast:   make(chan *BroadcastMessage, 1024),
		authService: authService,
		stopChan:    make(chan struct{}),
	}
}

// Run starts the hub's main event loop
func (h *AnalyticsHub) Run() {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.mu.Unlock()

	log.Println("[AnalyticsHub] Started")

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			clientCount := len(h.clients)
			h.mu.Unlock()
			log.Printf("[AnalyticsHub] Client connected (user: %s, account: %s). Total: %d",
				client.userID, client.accountID, clientCount)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[AnalyticsHub] Client disconnected (user: %s). Total: %d",
					client.userID, len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this channel
				client.mu.RLock()
				subscribed := client.subscriptions[message.Channel]
				client.mu.RUnlock()

				if !subscribed {
					continue
				}

				// Non-blocking send with backpressure handling
				select {
				case client.send <- []AnalyticsMessage{message.Message}:
					// Successfully queued
				default:
					// Client's send buffer is full - log and skip
					// This prevents slow clients from blocking the hub
					log.Printf("[AnalyticsHub] Client %s buffer full, dropping message", client.userID)
				}
			}
			h.mu.RUnlock()

		case <-h.stopChan:
			log.Println("[AnalyticsHub] Shutting down")
			return
		}
	}
}

// Stop gracefully shuts down the hub
func (h *AnalyticsHub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	close(h.stopChan)
	h.running = false

	// Close all client connections
	for client := range h.clients {
		client.close()
	}
}

// Broadcast sends a message to all clients subscribed to a channel
func (h *AnalyticsHub) Broadcast(channel string, messageType string, data interface{}) {
	msg := AnalyticsMessage{
		Type:      messageType,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Data:      data,
	}

	select {
	case h.broadcast <- &BroadcastMessage{
		Channel: channel,
		Message: msg,
	}:
	default:
		// Broadcast buffer full - this should rarely happen
		log.Printf("[AnalyticsHub] Warning: broadcast buffer full for channel %s", channel)
	}
}

// ServeAnalyticsWs handles WebSocket upgrade for analytics connections
func (h *AnalyticsHub) ServeAnalyticsWs(w http.ResponseWriter, r *http.Request) {
	// Extract and validate JWT token
	userID, accountID, role, err := h.extractAndValidateToken(r)
	if err != nil {
		log.Printf("[AnalyticsWS] Authentication failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[AnalyticsWS] Upgrade failed: %v", err)
		return
	}

	log.Printf("[AnalyticsWS] Connection upgraded for user %s (role: %s)", userID, role)

	// Create client
	client := &AnalyticsClient{
		hub:           h,
		conn:          conn,
		send:          make(chan []AnalyticsMessage, sendChannelSize),
		userID:        userID,
		accountID:     accountID,
		role:          role,
		subscriptions: make(map[string]bool),
		rateLimiter:   newRateLimiter(maxMessagesPerSecond, rateLimitWindow/time.Duration(maxMessagesPerSecond)),
		closeChan:     make(chan struct{}),
	}

	// Register client
	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// extractAndValidateToken extracts and validates the JWT token from the request
func (h *AnalyticsHub) extractAndValidateToken(r *http.Request) (userID, accountID, role string, err error) {
	if h.authService == nil {
		return "", "", "", fmt.Errorf("auth service not configured")
	}

	// Try query parameter first (ws://host/ws/analytics?token=xyz)
	token := r.URL.Query().Get("token")

	// Fall back to Authorization header
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token = parts[1]
			}
		}
	}

	if token == "" {
		return "", "", "", fmt.Errorf("no token provided")
	}

	// Validate token
	claims, err := auth.ValidateTokenWithDefault(token)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid token: %v", err)
	}

	return claims.UserID, claims.UserID, claims.Role, nil
}

// readPump handles incoming messages from the client
func (c *AnalyticsClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[AnalyticsWS] Read error for user %s: %v", c.userID, err)
			}
			break
		}

		// Parse subscription message
		var subMsg SubscriptionMessage
		if err := json.Unmarshal(message, &subMsg); err != nil {
			log.Printf("[AnalyticsWS] Invalid subscription message from user %s: %v", c.userID, err)
			continue
		}

		// Handle subscription/unsubscription
		c.handleSubscription(&subMsg)
	}
}

// writePump handles outgoing messages to the client with batching
func (c *AnalyticsClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	batchTicker := time.NewTicker(batchInterval)
	defer func() {
		ticker.Stop()
		batchTicker.Stop()
		c.conn.Close()
	}()

	var batch []AnalyticsMessage

	for {
		select {
		case messages, ok := <-c.send:
			if !ok {
				// Channel closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Add messages to batch
			batch = append(batch, messages...)

			// Send batch if it reaches max size
			if len(batch) >= maxBatchSize {
				if err := c.sendBatch(batch); err != nil {
					log.Printf("[AnalyticsWS] Write error for user %s: %v", c.userID, err)
					return
				}
				batch = nil
			}

		case <-batchTicker.C:
			// Send batch on interval (60 FPS target)
			if len(batch) > 0 {
				if err := c.sendBatch(batch); err != nil {
					log.Printf("[AnalyticsWS] Write error for user %s: %v", c.userID, err)
					return
				}
				batch = nil
			}

		case <-ticker.C:
			// Send ping
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.closeChan:
			return
		}
	}
}

// sendBatch sends a batch of messages to the client
func (c *AnalyticsClient) sendBatch(batch []AnalyticsMessage) error {
	if !c.rateLimiter.allow() {
		log.Printf("[AnalyticsWS] Rate limit exceeded for user %s", c.userID)
		return nil // Skip this batch but don't disconnect
	}

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))

	// Send as JSON array
	data, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// handleSubscription processes subscription/unsubscription requests
func (c *AnalyticsClient) handleSubscription(msg *SubscriptionMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch msg.Action {
	case "subscribe":
		for _, channel := range msg.Channels {
			if c.isValidChannel(channel) {
				c.subscriptions[channel] = true
				log.Printf("[AnalyticsWS] User %s subscribed to channel: %s", c.userID, channel)
			} else {
				log.Printf("[AnalyticsWS] User %s attempted to subscribe to invalid channel: %s", c.userID, channel)
			}
		}

	case "unsubscribe":
		for _, channel := range msg.Channels {
			delete(c.subscriptions, channel)
			log.Printf("[AnalyticsWS] User %s unsubscribed from channel: %s", c.userID, channel)
		}

	default:
		log.Printf("[AnalyticsWS] Unknown subscription action from user %s: %s", c.userID, msg.Action)
	}
}

// isValidChannel checks if a channel name is valid
func (c *AnalyticsClient) isValidChannel(channel string) bool {
	validChannels := map[string]bool{
		ChannelRoutingMetrics:  true,
		ChannelLPPerformance:   true,
		ChannelExposureUpdates: true,
		ChannelAlerts:          true,
	}
	return validChannels[channel]
}

// close gracefully closes the client connection
func (c *AnalyticsClient) close() {
	c.closeOnce.Do(func() {
		close(c.closeChan)
	})
}

// Helper function for min (Go 1.21+)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
