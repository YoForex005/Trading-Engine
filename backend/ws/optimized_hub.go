package ws

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// OptimizedHub is a high-performance WebSocket hub with quote throttling
// and bounded memory to prevent crashes from high-frequency market data
type OptimizedHub struct {
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client

	mu              sync.RWMutex
	latestPrices    map[string]*MarketTick
	disabledSymbols map[string]bool

	// Throttling: Track last broadcast price per symbol
	lastBroadcast   map[string]float64
	throttleMu      sync.RWMutex

	// Stats
	ticksReceived   int64
	ticksThrottled  int64
	ticksBroadcast  int64

	// Engine references
	bbookEngine     *core.Engine
	tickStore       TickStorer
	authService     AuthService
}

// AuthService interface for authentication
type AuthService interface {
	ValidateToken(token string) (*TokenClaims, error)
}

// TokenClaims represents JWT claims (compatible with auth.Claims)
type TokenClaims struct {
	UserID    string
	AccountID string
}

// NewOptimizedHub creates a high-performance hub
func NewOptimizedHub() *OptimizedHub {
	return &OptimizedHub{
		clients:         make(map[*Client]bool),
		broadcast:       make(chan []byte, 4096), // Larger buffer
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		latestPrices:    make(map[string]*MarketTick),
		disabledSymbols: make(map[string]bool),
		lastBroadcast:   make(map[string]float64),
	}
}

// BroadcastTickOptimized broadcasts with throttling to prevent CPU overload
// Minimum price change threshold: 0.0001% (1/100th of a pip)
func (h *OptimizedHub) BroadcastTickOptimized(tick *MarketTick) {
	atomic.AddInt64(&h.ticksReceived, 1)

	// ============================================
	// CRITICAL FIX: ALWAYS PERSIST TICKS FIRST
	// Storage happens BEFORE any filtering/throttling
	// This ensures ALL market data is captured regardless of:
	// - WebSocket client connections
	// - Symbol enabled/disabled status
	// - Price change throttling
	// ============================================
	if h.tickStore != nil {
		h.tickStore.StoreTick(tick.Symbol, tick.Bid, tick.Ask, tick.Spread, tick.LP, time.Now())
	}

	// Update B-Book engine (needs all prices for execution)
	if h.bbookEngine != nil {
		h.bbookEngine.UpdatePrice(tick.Symbol, tick.Bid, tick.Ask)
	}

	// Update latest price (always)
	h.mu.Lock()
	h.latestPrices[tick.Symbol] = tick

	// Skip broadcast if disabled (but tick is already stored above)
	if h.disabledSymbols[tick.Symbol] {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	// Throttling: Skip broadcast if price hasn't changed enough
	// NOTE: Tick is already stored above, throttling only affects broadcast
	h.throttleMu.RLock()
	lastPrice, exists := h.lastBroadcast[tick.Symbol]
	h.throttleMu.RUnlock()

	if exists && lastPrice > 0 {
		priceChange := (tick.Bid - lastPrice) / lastPrice
		if priceChange < 0 {
			priceChange = -priceChange
		}

		// Skip broadcast if change < 0.0001% (tick already stored above)
		if priceChange < 0.000001 {
			atomic.AddInt64(&h.ticksThrottled, 1)
			return
		}
	}

	// Update last broadcast price
	h.throttleMu.Lock()
	h.lastBroadcast[tick.Symbol] = tick.Bid
	h.throttleMu.Unlock()

	// Marshal and broadcast
	data, err := json.Marshal(tick)
	if err != nil {
		return
	}

	// Non-blocking send
	select {
	case h.broadcast <- data:
		atomic.AddInt64(&h.ticksBroadcast, 1)
	default:
		// Buffer full - drop to prevent blocking
	}
}

// Run starts the optimized hub
func (h *OptimizedHub) RunOptimized() {
	// Stats logging
	go h.logStats()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			clientCount := len(h.clients)
			h.mu.Unlock()
			log.Printf("[OptimizedHub] Client connected. Total: %d", clientCount)

			// Send latest prices to new client
			h.sendLatestPrices(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[OptimizedHub] Client disconnected. Total: %d", len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			clientCount := len(h.clients)
			h.mu.RUnlock()

			if clientCount == 0 {
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full - skip (don't disconnect)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *OptimizedHub) sendLatestPrices(client *Client) {
	h.mu.RLock()
	for symbol, tick := range h.latestPrices {
		if !h.disabledSymbols[symbol] {
			if data, err := json.Marshal(tick); err == nil {
				select {
				case client.send <- data:
				default:
				}
			}
		}
	}
	h.mu.RUnlock()
}

func (h *OptimizedHub) logStats() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		received := atomic.LoadInt64(&h.ticksReceived)
		throttled := atomic.LoadInt64(&h.ticksThrottled)
		broadcast := atomic.LoadInt64(&h.ticksBroadcast)

		if received > 0 {
			throttleRate := float64(throttled) / float64(received) * 100
			log.Printf("[OptimizedHub] Stats: received=%d, broadcast=%d, throttled=%d (%.1f%% reduction)",
				received, broadcast, throttled, throttleRate)
		}
	}
}

// SetBBookEngine sets the B-Book engine reference
func (h *OptimizedHub) SetBBookEngine(engine *core.Engine) {
	h.bbookEngine = engine
}

// SetTickStore sets the tick store reference
func (h *OptimizedHub) SetTickStore(store TickStorer) {
	h.tickStore = store
}

// GetLatestPrice returns the latest price for a symbol
func (h *OptimizedHub) GetLatestPriceOptimized(symbol string) *MarketTick {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.latestPrices[symbol]
}

// UpdateDisabledSymbols updates disabled symbols
func (h *OptimizedHub) UpdateDisabledSymbolsOptimized(disabled map[string]bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.disabledSymbols = disabled
}
