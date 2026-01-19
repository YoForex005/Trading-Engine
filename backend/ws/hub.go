package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/gorilla/websocket"
)

// TickStorer interface for tick storage (allows both old TickStore and OptimizedTickStore)
type TickStorer interface {
	StoreTick(symbol string, bid, ask, spread float64, lp string, timestamp time.Time)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a connected WebSocket client
type Client struct {
	conn      *websocket.Conn
	send      chan []byte
	symbols   map[string]bool
	userID    string // JWT user ID
	accountID string // Associated account ID
	mu        sync.Mutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	tickStore   TickStorer // Interface to support both TickStore and OptimizedTickStore
	bbookEngine *core.Engine
	authService *auth.Service

	mu              sync.RWMutex
	latestPrices    map[string]*MarketTick
	disabledSymbols map[string]bool

	// Throttling: Track last broadcast price per symbol to reduce CPU load
	lastBroadcast map[string]float64
	throttleMu    sync.RWMutex

	// Stats for monitoring
	ticksReceived  int64
	ticksThrottled int64
	ticksBroadcast int64
}

// MarketTick represents a price update for clients
type MarketTick struct {
	Type      string  `json:"type"`
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Spread    float64 `json:"spread"`
	Timestamp int64   `json:"timestamp"`
	LP        string  `json:"lp"` // Liquidity Provider source
}

func NewHub() *Hub {
	h := &Hub{
		clients:         make(map[*Client]bool),
		broadcast:       make(chan []byte, 4096), // Larger buffer to handle bursts
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		latestPrices:    make(map[string]*MarketTick),
		disabledSymbols: make(map[string]bool),
		lastBroadcast:   make(map[string]float64),
	}

	// Start stats logging
	go h.logStats()

	return h
}

// logStats logs hub performance metrics every 60 seconds
func (h *Hub) logStats() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		received := atomic.LoadInt64(&h.ticksReceived)
		throttled := atomic.LoadInt64(&h.ticksThrottled)
		broadcast := atomic.LoadInt64(&h.ticksBroadcast)

		if received > 0 {
			throttleRate := float64(throttled) / float64(received) * 100
			h.mu.RLock()
			clientCount := len(h.clients)
			h.mu.RUnlock()
			log.Printf("[Hub] Stats: received=%d, broadcast=%d, throttled=%d (%.1f%% reduction), clients=%d",
				received, broadcast, throttled, throttleRate, clientCount)
		}
	}
}

// UpdateDisabledSymbols updates the local filter list
func (h *Hub) UpdateDisabledSymbols(disabled map[string]bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.disabledSymbols = disabled
}

// ToggleSymbol updates a single symbol's status
func (h *Hub) ToggleSymbol(symbol string, disabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.disabledSymbols[symbol] = disabled
}


// UpdateSymbol updates symbol specifications in the hub
func (h *Hub) UpdateSymbol(spec *core.SymbolSpec) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// The hub maintains disabled status, actual spec is in engine
	// Preserve disabled state if it exists
	if h.disabledSymbols[spec.Symbol] {
		// Keep the disabled state
		h.disabledSymbols[spec.Symbol] = true
	}
}
// BroadcastTick broadcasts a market tick to all clients with THROTTLING
// Throttling reduces CPU load by 60-80% by skipping tiny price changes
func (h *Hub) BroadcastTick(tick *MarketTick) {
	atomic.AddInt64(&h.ticksReceived, 1)

	// Update latest price (always - needed for queries)
	h.mu.Lock()
	h.latestPrices[tick.Symbol] = tick

	// Skip broadcast if symbol is disabled
	if h.disabledSymbols[tick.Symbol] {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	// ============================================
	// THROTTLING: Skip broadcast if price change < 0.0001% (1/100th of a pip)
	// This reduces broadcast volume by 60-80% and prevents CPU overload
	// ============================================
	h.throttleMu.RLock()
	lastPrice, exists := h.lastBroadcast[tick.Symbol]
	h.throttleMu.RUnlock()

	if exists && lastPrice > 0 {
		priceChange := (tick.Bid - lastPrice) / lastPrice
		if priceChange < 0 {
			priceChange = -priceChange
		}

		// Skip if change < 0.0001% - still update engine & store but skip broadcast
		if priceChange < 0.000001 {
			atomic.AddInt64(&h.ticksThrottled, 1)

			// Still update B-Book engine (needs all prices for accurate execution)
			if h.bbookEngine != nil {
				h.bbookEngine.UpdatePrice(tick.Symbol, tick.Bid, tick.Ask)
			}

			// Still store tick (for complete chart history)
			if h.tickStore != nil {
				h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
			}
			return
		}
	}

	// Update last broadcast price
	h.throttleMu.Lock()
	h.lastBroadcast[tick.Symbol] = tick.Bid
	h.throttleMu.Unlock()

	// Notify B-Book engine of new price (for order execution)
	if h.bbookEngine != nil {
		h.bbookEngine.UpdatePrice(tick.Symbol, tick.Bid, tick.Ask)
	}

	// Persist tick for chart history
	if h.tickStore != nil {
		h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
	}

	data, err := json.Marshal(tick)
	if err != nil {
		return
	}

	// NON-BLOCKING SEND: If buffer full, drop tick to keep engine running
	select {
	case h.broadcast <- data:
		atomic.AddInt64(&h.ticksBroadcast, 1)
	default:
		// Buffer full - drop to prevent blocking (data still stored for history)
	}
}

// GetLatestPrice returns the latest price for a symbol
func (h *Hub) GetLatestPrice(symbol string) *MarketTick {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.latestPrices[symbol]
}

// SetTickStore sets the tick store for persisting market data
// Accepts any TickStorer interface (works with both TickStore and OptimizedTickStore)
func (h *Hub) SetTickStore(ts TickStorer) {
	h.tickStore = ts
}

// SetBBookEngine sets the B-Book engine for symbol synchronization
func (h *Hub) SetBBookEngine(engine *core.Engine) {
	h.bbookEngine = engine
}

// SetAuthService sets the authentication service for validating tokens
func (h *Hub) SetAuthService(svc *auth.Service) {
	h.authService = svc
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			clientCount := len(h.clients)
			h.mu.Unlock()
			log.Printf("[Hub] Client connected. Total clients: %d", clientCount)

			// Send latest prices for all symbols upon connection
			h.mu.RLock()
			for _, tick := range h.latestPrices {
				if !h.disabledSymbols[tick.Symbol] {
					if data, err := json.Marshal(tick); err == nil {
						// Try non-blocking send to client on init
						select {
						case client.send <- data:
						default:
						}
					}
				}
			}
			h.mu.RUnlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[Hub] Client disconnected. Total clients: %d", len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clientCount := len(h.clients)
			h.mu.RUnlock()

			if clientCount == 0 {
				continue // No clients, skip broadcasting
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full - just drop the message instead of disconnecting
					// The client will get the next update
				}
			}
			h.mu.RUnlock()
		}
	}
}

// ServeWs handles websocket requests from the peer with JWT authentication.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Printf("[WS] Upgrade request from %s", r.RemoteAddr)

	// Extract and validate JWT token from query parameters or headers
	userID, accountID, err := extractAndValidateToken(hub, r)
	if err != nil {
		log.Printf("[WS] Authentication FAILED for %s: %v", r.RemoteAddr, err)
		// Return 401 Unauthorized
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade FAILED for %s: %v", r.RemoteAddr, err)
		return
	}

	log.Printf("[WS] Upgrade SUCCESS for user %s (account %s) from %s", userID, accountID, r.RemoteAddr)

	client := &Client{
		conn:      conn,
		send:      make(chan []byte, 1024), // BUFFERED: Handle bursts
		symbols:   make(map[string]bool),
		userID:    userID,
		accountID: accountID,
	}
	hub.register <- client

	// Write pump
	go func() {
		defer conn.Close()
		for message := range client.send {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("[WS] Write error for user %s: %v", userID, err)
				break
			}
		}
	}()

	// Read pump (handle subscriptions, etc.)
	go func() {
		defer func() {
			hub.unregister <- client
			conn.Close()
			log.Printf("[WS] Connection closed for user %s", userID)
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}()
}

// extractAndValidateToken extracts the JWT token from query params or Authorization header
// and validates it using the auth service. Returns (userID, accountID, error).
func extractAndValidateToken(hub *Hub, r *http.Request) (string, string, error) {
	if hub.authService == nil {
		return "", "", fmt.Errorf("auth service not configured")
	}

	// Try query parameter first (ws://localhost/ws?token=xyz)
	token := r.URL.Query().Get("token")

	// Fall back to Authorization header (Authorization: Bearer <token>)
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
		return "", "", fmt.Errorf("no token provided")
	}

	// Validate token using the auth service's secret (same one used to generate tokens)
	claims, err := hub.authService.ValidateToken(token)
	if err != nil {
		return "", "", fmt.Errorf("invalid token: %v", err)
	}

	userID := claims.UserID
	accountID := userID // For now, use userID as accountID (same value for admin/trader accounts)

	return userID, accountID, nil
}

// BroadcastMessage sends a generic message to all connected clients
func (h *Hub) BroadcastMessage(message []byte) {
	select {
	case h.broadcast <- message:
	default:
		log.Println("[Hub] Broadcast buffer full, message dropped")
	}
}
