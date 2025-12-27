package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/bbook"
	"github.com/epic1st/rtx/backend/instantswap"
	"github.com/epic1st/rtx/backend/oanda"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents a connected WebSocket client
type Client struct {
	conn    *websocket.Conn
	send    chan []byte
	symbols map[string]bool
	mu      sync.Mutex
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients           map[*Client]bool
	broadcast         chan []byte
	register          chan *Client
	unregister        chan *Client
	oandaClient       *oanda.Client
	instantswapClient *instantswap.Client
	tickStore         *tickstore.TickStore
	bbookEngine       *bbook.Engine
	mu                sync.RWMutex
	latestPrices      map[string]*MarketTick
	activeLP          string // "OANDA" or "INSTANTSWAP"
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
	return &Hub{
		clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		latestPrices: make(map[string]*MarketTick),
	}
}

// GetLatestPrice returns the latest price for a symbol
func (h *Hub) GetLatestPrice(symbol string) *MarketTick {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.latestPrices[symbol]
}

// SetTickStore sets the tick store for persisting market data
func (h *Hub) SetTickStore(ts *tickstore.TickStore) {
	h.tickStore = ts
}

// SetBBookEngine sets the B-Book engine for symbol synchronization
func (h *Hub) SetBBookEngine(engine *bbook.Engine) {
	h.bbookEngine = engine
}

// GetActiveLP returns the currently active liquidity provider
func (h *Hub) GetActiveLP() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.activeLP == "" {
		return "NONE"
	}
	return h.activeLP
}

// SetActiveLP sets the active LP (called after successful connection)
func (h *Hub) SetActiveLP(lp string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.activeLP = lp
	log.Printf("[Hub] Active LP set to: %s", lp)
}

// LoadLatestPricesFromStore loads the last known price for each symbol from the tick store
func (h *Hub) LoadLatestPricesFromStore() {
	if h.tickStore == nil {
		return
	}

	symbols := h.tickStore.GetSymbols()
	count := 0

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, symbol := range symbols {
		history := h.tickStore.GetHistory(symbol, 1)
		if len(history) > 0 {
			lastTick := history[0]

			h.latestPrices[symbol] = &MarketTick{
				Type:      "tick",
				Symbol:    symbol,
				Bid:       lastTick.Bid,
				Ask:       lastTick.Ask,
				Spread:    lastTick.Spread,
				Timestamp: lastTick.Timestamp.UnixMilli(),
				LP:        lastTick.LP,
			}
			count++
		}
	}

	log.Printf("[WS] Loaded latest prices for %d symbols from TickStore", count)
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client connected. Total: %d", len(h.clients))

			// Send latest prices to new client
			h.sendLatestPrices(client)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WS] Client disconnected. Total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) sendLatestPrices(client *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, tick := range h.latestPrices {
		data, err := json.Marshal(tick)
		if err == nil {
			client.send <- data
		}
	}
}

// ConnectToOandaLP connects to OANDA as our Liquidity Provider
func (h *Hub) ConnectToOandaLP(apiKey string) error {
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║        CONNECTING TO LIQUIDITY PROVIDER: OANDA            ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	h.oandaClient = oanda.NewClient(apiKey)

	// Get accounts
	accounts, err := h.oandaClient.GetAccounts()
	if err != nil {
		return err
	}
	log.Printf("[LP:OANDA] Found accounts: %v", accounts)

	// Get account summary
	summary, err := h.oandaClient.GetAccountSummary()
	if err != nil {
		log.Printf("[LP:OANDA] Warning: Could not get account summary: %v", err)
	} else {
		log.Printf("[LP:OANDA] Account: %s | Balance: %s %s | NAV: %s",
			summary.ID, summary.Balance, summary.Currency, summary.NAV)
	}

	// Fetch ALL available instruments from OANDA
	log.Println("[LP:OANDA] Fetching all available instruments...")
	oandaInstruments, err := h.oandaClient.GetInstruments()
	if err != nil {
		log.Printf("[LP:OANDA] Error fetching instruments: %v. Falling back to default list.", err)
	}

	var instruments []string
	if len(oandaInstruments) > 0 {
		log.Printf("[LP:OANDA] Found %d instruments available for trading", len(oandaInstruments))

		for _, inst := range oandaInstruments {
			// Convert format: EUR_USD -> EURUSD
			symbol := strings.Replace(inst.Name, "_", "", -1)
			instruments = append(instruments, inst.Name) // Keep OANDA format for streaming subscription

			// Sync to B-Book Engine
			if h.bbookEngine != nil {
				// Parse pip location -> pip size
				pipSize := 1.0
				if inst.PipLocation < 0 {
					// e.g. -4 -> 0.0001
					denom := 1.0
					for i := 0; i < -inst.PipLocation; i++ {
						denom *= 10
					}
					pipSize = 1.0 / denom
				} else {
					for i := 0; i < inst.PipLocation; i++ {
						pipSize *= 10
					}
				}

				// Approximations for pip value if not available
				pipValue := 10.0
				if strings.Contains(symbol, "JPY") {
					pipValue = 9.0 // Approx
				}

				spec := &bbook.SymbolSpec{
					Symbol:           symbol,
					ContractSize:     100000, // Standard lot size assumption, OANDA uses units
					PipSize:          pipSize,
					PipValue:         pipValue, // Need dynamic pip value calc really, but generic for now
					MinVolume:        0.01,
					MaxVolume:        1000,
					VolumeStep:       0.01,
					MarginPercent:    1, // Default 1% (100:1)
					CommissionPerLot: 0,
				}

				// Adjust specific asset classes
				if inst.Type == "CFD" {
					spec.ContractSize = 1
					spec.MarginPercent = 1 // 1%
				} else if inst.Type == "METAL" {
					spec.ContractSize = 100 // 100oz Gold
				}

				h.bbookEngine.UpdateSymbol(spec)
			}
		}
	} else {
		// Fallback list
		instruments = []string{
			"EUR_USD", "GBP_USD", "USD_JPY", "USD_CHF", "AUD_USD", "USD_CAD", "NZD_USD",
			"EUR_GBP", "EUR_JPY", "GBP_JPY", "EUR_CHF", "EUR_AUD", "GBP_AUD",
			"AUD_JPY", "CHF_JPY", "CAD_JPY", "NZD_JPY",
			"XAU_USD", "XAG_USD", "BTC_USD", "ETH_USD",
		}
	}

	log.Printf("[LP:OANDA] Subscribing to %d instruments...", len(instruments))

	// Start streaming prices from OANDA LP
	if err := h.oandaClient.StreamPrices(instruments); err != nil {
		return err
	}

	// Process incoming LP prices and broadcast to clients
	go h.processLPFeed()

	return nil
}

// ConnectToInstantSwapLP connects to InstantSwap provider
func (h *Hub) ConnectToInstantSwapLP() error {
	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║        CONNECTING TO LIQUIDITY PROVIDER: INSTANTSWAP      ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	h.instantswapClient = instantswap.NewClient()

	if err := h.instantswapClient.Connect(); err != nil {
		return err
	}

	// Subscribe
	h.instantswapClient.Subscribe()

	log.Println("[LP:INSTANTSWAP] Connected and subscribing to flow")

	go h.processInstantSwapFeed()

	return nil
}

func (h *Hub) processInstantSwapFeed() {
	log.Println("[LP:INSTANTSWAP] Price feed processor started")

	for price := range h.instantswapClient.GetPricesChan() {
		// Spread calculation (approx for now, or use embedded spread if available)
		spread := (price.Ask - price.Bid)

		// In forex pips or just raw?
		// If Forex (e.g. 1.1234), spread is usually small (0.0001)
		// UI expects spread in pips likely, OR raw spread.
		// Existing OANDA logic: spread = (ask - bid) * 100000

		symbol := price.Symbol

		// Heuristic spread scaling
		displaySpread := spread
		if strings.Contains(symbol, "JPY") {
			displaySpread = spread * 100
		} else if len(symbol) == 6 && !strings.Contains(symbol, "XAU") && !strings.Contains(symbol, "BTC") {
			// Forex 5 digit
			displaySpread = spread * 100000
		} else {
			// Metals/Crypto - use raw or similar
			if strings.Contains(symbol, "XAU") {
				displaySpread = spread * 100
			}
		}

		tick := &MarketTick{
			Type:      "tick",
			Symbol:    symbol,
			Bid:       price.Bid,
			Ask:       price.Ask,
			Spread:    displaySpread,
			Timestamp: price.Timestamp,
			LP:        "INSTANTSWAP",
		}

		// Store latest price
		h.mu.Lock()
		h.latestPrices[symbol] = tick
		h.mu.Unlock()

		// Store tick in tick store
		if h.tickStore != nil {
			h.tickStore.StoreTick(symbol, price.Bid, price.Ask, displaySpread, "INSTANTSWAP", time.UnixMilli(price.Timestamp))
		}

		// Sync to B-Book Engine if symbol doesn't exist
		if h.bbookEngine != nil {
			// Basic symbol registration if unknown
			// In production, we should have a symbol list. here we just auto-add.
		}

		// Broadcast
		data, _ := json.Marshal(tick)
		h.broadcast <- data
	}
}

func (h *Hub) processLPFeed() {
	log.Println("[LP:OANDA] Price feed processor started")

	for price := range h.oandaClient.GetPricesChan() {
		if len(price.Bids) == 0 || len(price.Asks) == 0 {
			continue
		}

		bid, _ := strconv.ParseFloat(price.Bids[0].Price, 64)
		ask, _ := strconv.ParseFloat(price.Asks[0].Price, 64)

		// Convert OANDA format (EUR_USD) to standard format (EURUSD)
		symbol := strings.Replace(price.Instrument, "_", "", 1)

		spread := (ask - bid) * 100000 // Spread in pips for forex

		// Adjust spread calculation for non-forex
		if strings.HasPrefix(symbol, "XAU") || strings.HasPrefix(symbol, "XAG") {
			spread = (ask - bid) * 100 // Metals
		} else if strings.HasPrefix(symbol, "BTC") || strings.HasPrefix(symbol, "ETH") {
			spread = ask - bid // Crypto in absolute
		} else if strings.Contains(symbol, "JPY") {
			spread = (ask - bid) * 100 // JPY pairs
		}

		tick := &MarketTick{
			Type:      "tick",
			Symbol:    symbol,
			Bid:       bid,
			Ask:       ask,
			Spread:    spread,
			Timestamp: price.Time.UnixMilli(),
			LP:        "OANDA",
		}

		// Store latest price
		h.mu.Lock()
		h.latestPrices[symbol] = tick
		h.mu.Unlock()

		// Store tick in tick store for historical data
		if h.tickStore != nil {
			h.tickStore.StoreTick(symbol, bid, ask, spread, "OANDA", time.Now())
		}

		// Broadcast to all connected clients
		data, _ := json.Marshal(tick)
		h.broadcast <- data
	}

	log.Println("[LP:OANDA] Price feed disconnected")
}

// GetOandaClient returns the OANDA client for order execution
func (h *Hub) GetOandaClient() *oanda.Client {
	return h.oandaClient
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{
		conn:    conn,
		send:    make(chan []byte, 256),
		symbols: make(map[string]bool),
	}
	hub.register <- client

	// Write pump
	go func() {
		defer conn.Close()
		for message := range client.send {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				break
			}
		}
	}()

	// Read pump (handle subscriptions, etc.)
	go func() {
		defer func() {
			hub.unregister <- client
			conn.Close()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			log.Println("[WS] Received:", string(message))
		}
	}()
}
