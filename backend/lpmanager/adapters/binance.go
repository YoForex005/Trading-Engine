package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/gorilla/websocket"
)

const (
	BinanceRestURL = "https://api.binance.com/api/v3"
	BinanceWSURL   = "wss://stream.binance.com:9443/stream"
)

// BinanceAdapter implements LPAdapter for Binance
type BinanceAdapter struct {
	id                string
	name              string
	conn              *websocket.Conn
	quotesChan        chan lpmanager.Quote
	stopChan          chan struct{}
	mu                sync.RWMutex
	connected         bool
	symbols           []lpmanager.SymbolInfo
	subscribedSymbols []string
	lastTick          time.Time
	errorMsg          string
}

// NewBinanceAdapter creates a new Binance adapter
func NewBinanceAdapter() *BinanceAdapter {
	return &BinanceAdapter{
		id:         "binance",
		name:       "Binance",
		quotesChan: make(chan lpmanager.Quote, 500),
		stopChan:   make(chan struct{}),
		symbols:    []lpmanager.SymbolInfo{},
	}
}

func (b *BinanceAdapter) ID() string   { return b.id }
func (b *BinanceAdapter) Name() string { return b.name }
func (b *BinanceAdapter) Type() string { return "WebSocket" }

func (b *BinanceAdapter) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

func (b *BinanceAdapter) GetStatus() lpmanager.LPStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return lpmanager.LPStatus{
		ID:           b.id,
		Name:         b.name,
		Type:         "WebSocket",
		Connected:    b.connected,
		Enabled:      true,
		SymbolCount:  len(b.symbols),
		LastTick:     b.lastTick,
		ErrorMessage: b.errorMsg,
	}
}

func (b *BinanceAdapter) GetQuotesChan() <-chan lpmanager.Quote {
	return b.quotesChan
}

// GetSymbols fetches all available trading pairs from Binance
func (b *BinanceAdapter) GetSymbols() ([]lpmanager.SymbolInfo, error) {
	log.Println("[Binance] Fetching all available symbols...")

	resp, err := http.Get(BinanceRestURL + "/exchangeInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var exchangeInfo struct {
		Symbols []struct {
			Symbol     string `json:"symbol"`
			BaseAsset  string `json:"baseAsset"`
			QuoteAsset string `json:"quoteAsset"`
			Status     string `json:"status"`
			Filters    []struct {
				FilterType string `json:"filterType"`
				MinQty     string `json:"minQty"`
				MaxQty     string `json:"maxQty"`
				StepSize   string `json:"stepSize"`
			} `json:"filters"`
		} `json:"symbols"`
	}

	if err := json.Unmarshal(body, &exchangeInfo); err != nil {
		return nil, fmt.Errorf("failed to parse exchange info: %w", err)
	}

	symbols := make([]lpmanager.SymbolInfo, 0)

	for _, s := range exchangeInfo.Symbols {
		// Only include active USDT pairs (we'll convert to USD format)
		if s.Status != "TRADING" {
			continue
		}
		if s.QuoteAsset != "USDT" {
			continue
		}

		// Convert BTCUSDT -> BTCUSD format
		displayName := strings.Replace(s.Symbol, "USDT", "USD", 1)

		// Get lot size from filters
		var minLot, maxLot, stepSize float64
		for _, f := range s.Filters {
			if f.FilterType == "LOT_SIZE" {
				minLot, _ = strconv.ParseFloat(f.MinQty, 64)
				maxLot, _ = strconv.ParseFloat(f.MaxQty, 64)
				stepSize, _ = strconv.ParseFloat(f.StepSize, 64)
				break
			}
		}

		// Determine type
		symbolType := "crypto"

		symbols = append(symbols, lpmanager.SymbolInfo{
			Symbol:        displayName,
			DisplayName:   displayName,
			BaseCurrency:  s.BaseAsset,
			QuoteCurrency: "USD",
			MinLotSize:    minLot,
			MaxLotSize:    maxLot,
			LotStep:       stepSize,
			Type:          symbolType,
		})
	}

	log.Printf("[Binance] Found %d tradeable symbols", len(symbols))
	b.symbols = symbols
	return symbols, nil
}

// Connect establishes WebSocket connection
func (b *BinanceAdapter) Connect() error {
	log.Println("[Binance] Connecting...")

	// First fetch symbols if not already done
	if len(b.symbols) == 0 {
		if _, err := b.GetSymbols(); err != nil {
			b.errorMsg = err.Error()
			return err
		}
	}

	b.mu.Lock()
	b.connected = true
	b.errorMsg = ""
	b.mu.Unlock()

	log.Println("[Binance] Ready to subscribe to symbols")
	return nil
}

// Subscribe starts streaming prices for given symbols
func (b *BinanceAdapter) Subscribe(symbols []string) error {
	if len(symbols) == 0 {
		return nil
	}

	// Close existing connection if any
	if b.conn != nil {
		b.conn.Close()
	}

	// Build stream URL
	streams := make([]string, 0, len(symbols))
	for _, sym := range symbols {
		// Convert BTCUSD -> btcusdt@bookTicker
		binanceSym := strings.ToLower(strings.Replace(sym, "USD", "USDT", 1))
		streams = append(streams, binanceSym+"@bookTicker")
	}

	wsURL := BinanceWSURL + "?streams=" + strings.Join(streams, "/")
	log.Printf("[Binance] Subscribing to %d symbols: %s", len(symbols), wsURL[:min(150, len(wsURL))])

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		b.errorMsg = err.Error()
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	b.conn = conn
	b.subscribedSymbols = symbols

	b.mu.Lock()
	b.connected = true
	b.mu.Unlock()

	log.Println("[Binance] WebSocket connected, reading messages...")

	// Start reading messages
	go b.readMessages()

	// Start heartbeat
	go b.heartbeat()

	return nil
}

func (b *BinanceAdapter) Unsubscribe(symbols []string) error {
	// For Binance, unsubscribe means closing the connection
	// A new Subscribe call with different symbols would reconnect
	return nil
}

func (b *BinanceAdapter) Disconnect() error {
	b.mu.Lock()
	b.connected = false
	b.mu.Unlock()

	close(b.stopChan)
	if b.conn != nil {
		b.conn.Close()
	}
	return nil
}

func (b *BinanceAdapter) readMessages() {
	defer func() {
		b.mu.Lock()
		b.connected = false
		b.mu.Unlock()
	}()

	for {
		select {
		case <-b.stopChan:
			return
		default:
		}

		_, message, err := b.conn.ReadMessage()
		if err != nil {
			log.Printf("[Binance] Read error: %v", err)
			go b.reconnect()
			return
		}

		b.handleMessage(message)
	}
}

func (b *BinanceAdapter) handleMessage(message []byte) {
	// Parse combined stream message
	var streamMsg struct {
		Stream string          `json:"stream"`
		Data   json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &streamMsg); err != nil {
		return
	}

	// Check if symbol contains BTC to log (noisy otherwise)
	if strings.Contains(string(streamMsg.Data), "BTCUSDT") {
		log.Printf("[DEBUG] BTC Payload: %s", string(streamMsg.Data))
	}

	// Parse book ticker
	// CRITICAL: Binance uses case-sensitive fields!
	// b = Best Bid Price
	// B = Best Bid Quantity
	// a = Best Ask Price
	// A = Best Ask Quantity
	var ticker struct {
		Symbol   string `json:"s"`
		BidPrice string `json:"b"`
		BidQty   string `json:"B"`
		AskPrice string `json:"a"`
		AskQty   string `json:"A"`
	}

	if err := json.Unmarshal(streamMsg.Data, &ticker); err != nil {
		return
	}

	bid, _ := strconv.ParseFloat(ticker.BidPrice, 64)
	ask, _ := strconv.ParseFloat(ticker.AskPrice, 64)
	// bidQty, _ := strconv.ParseFloat(ticker.BidQty, 64)
	// askQty, _ := strconv.ParseFloat(ticker.AskQty, 64)

	if bid <= 0 || ask <= 0 {
		return
	}

	// Convert BTCUSDT -> BTCUSD
	symbol := strings.Replace(ticker.Symbol, "USDT", "USD", 1)

	quote := lpmanager.Quote{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now().UnixMilli(),
		LP:        b.id,
	}

	b.mu.Lock()
	b.lastTick = time.Now()
	b.mu.Unlock()

	select {
	case b.quotesChan <- quote:
	default:
	}
}

func (b *BinanceAdapter) heartbeat() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopChan:
			return
		case <-ticker.C:
			if b.IsConnected() && b.conn != nil {
				b.conn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}
}

func (b *BinanceAdapter) reconnect() {
	log.Println("[Binance] Reconnecting in 3 seconds...")
	time.Sleep(3 * time.Second)

	if err := b.Subscribe(b.subscribedSymbols); err != nil {
		log.Printf("[Binance] Reconnect failed: %v", err)
		go b.reconnect()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
