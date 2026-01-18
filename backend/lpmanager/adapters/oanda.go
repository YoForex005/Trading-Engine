package adapters

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/oanda"
)

// OANDAAdapter wraps the existing OANDA client to implement LPAdapter
type OANDAAdapter struct {
	id         string
	name       string
	client     *oanda.Client
	quotesChan chan lpmanager.Quote
	stopChan   chan struct{}
	mu         sync.RWMutex
	connected  bool
	symbols    []lpmanager.SymbolInfo
	lastTick   time.Time
	errorMsg   string
	apiKey     string
	accountID  string
}

// NewOANDAAdapter creates a new OANDA adapter
func NewOANDAAdapter(apiKey, accountID string) *OANDAAdapter {
	return &OANDAAdapter{
		id:         "oanda",
		name:       "OANDA",
		quotesChan: make(chan lpmanager.Quote, 500),
		stopChan:   make(chan struct{}),
		symbols:    []lpmanager.SymbolInfo{},
		apiKey:     apiKey,
		accountID:  accountID,
	}
}

func (o *OANDAAdapter) ID() string   { return o.id }
func (o *OANDAAdapter) Name() string { return o.name }
func (o *OANDAAdapter) Type() string { return "REST/Streaming" }

func (o *OANDAAdapter) IsConnected() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.connected
}

func (o *OANDAAdapter) GetStatus() lpmanager.LPStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return lpmanager.LPStatus{
		ID:           o.id,
		Name:         o.name,
		Type:         "REST/Streaming",
		Connected:    o.connected,
		Enabled:      true,
		SymbolCount:  len(o.symbols),
		LastTick:     o.lastTick,
		ErrorMessage: o.errorMsg,
	}
}

func (o *OANDAAdapter) GetQuotesChan() <-chan lpmanager.Quote {
	return o.quotesChan
}

// GetSymbols fetches all available instruments from OANDA
func (o *OANDAAdapter) GetSymbols() ([]lpmanager.SymbolInfo, error) {
	log.Println("[OANDA] Fetching all available instruments...")

	if o.client == nil {
		o.client = oanda.NewClient(o.apiKey)
	}

	instruments, err := o.client.GetInstruments()
	if err != nil {
		o.errorMsg = err.Error()
		return nil, err
	}

	symbols := make([]lpmanager.SymbolInfo, 0, len(instruments))
	for _, inst := range instruments {
		// Determine type based on symbol
		symbolType := "forex"
		if len(inst.Name) >= 3 {
			prefix := inst.Name[:3]
			if prefix == "BTC" || prefix == "ETH" || prefix == "XRP" || prefix == "LTC" {
				symbolType = "crypto"
			} else if prefix == "XAU" || prefix == "XAG" || prefix == "XPT" || prefix == "XPD" {
				symbolType = "commodity"
			} else if prefix == "SPX" || prefix == "NAS" || prefix == "US3" || prefix == "UK1" {
				symbolType = "index"
			}
		}

		// Parse min trade size
		var minLotSize float64
		fmt.Sscanf(inst.MinimumTradeSize, "%f", &minLotSize)

		symbols = append(symbols, lpmanager.SymbolInfo{
			Symbol:        inst.Name,
			DisplayName:   inst.DisplayName,
			BaseCurrency:  "",
			QuoteCurrency: "",
			MinLotSize:    minLotSize,
			Type:          symbolType,
		})

	}

	log.Printf("[OANDA] Found %d instruments", len(symbols))
	o.symbols = symbols
	return symbols, nil
}

// Connect establishes connection to OANDA
func (o *OANDAAdapter) Connect() error {
	log.Println("[OANDA] Connecting...")

	if o.client == nil {
		o.client = oanda.NewClient(o.apiKey)
	}

	// Fetch symbols
	if len(o.symbols) == 0 {
		if _, err := o.GetSymbols(); err != nil {
			return err
		}
	}

	o.mu.Lock()
	o.connected = true
	o.errorMsg = ""
	o.mu.Unlock()

	log.Println("[OANDA] Ready")
	return nil
}

// Subscribe starts streaming prices for given symbols
func (o *OANDAAdapter) Subscribe(symbols []string) error {
	if o.client == nil {
		return nil
	}

	// Start streaming
	if err := o.client.StreamPrices(symbols); err != nil {
		o.errorMsg = err.Error()
		return err
	}

	// Start processing prices
	go o.processPrices()

	return nil
}

func (o *OANDAAdapter) processPrices() {
	log.Println("[OANDA] Price feed processor started")

	for price := range o.client.GetPricesChan() {
		if len(price.Bids) == 0 || len(price.Asks) == 0 {
			continue
		}

		bid := 0.0
		ask := 0.0
		if len(price.Bids) > 0 {
			// Parse bid
			for _, b := range price.Bids {
				if b.Price != "" {
					var p float64
					fmt.Sscanf(b.Price, "%f", &p)
					bid = p
					break
				}
			}
		}
		if len(price.Asks) > 0 {
			// Parse ask
			for _, a := range price.Asks {
				if a.Price != "" {
					var p float64
					fmt.Sscanf(a.Price, "%f", &p)
					ask = p
					break
				}
			}
		}

		// Convert EUR_USD -> EURUSD
		symbol := price.Instrument
		symbol = strings.Replace(symbol, "_", "", -1)

		quote := lpmanager.Quote{
			Symbol:    symbol,
			Bid:       bid,
			Ask:       ask,
			Timestamp: price.Time.UnixMilli(),
			LP:        o.id,
		}

		o.mu.Lock()
		o.lastTick = time.Now()
		o.mu.Unlock()

		select {
		case o.quotesChan <- quote:
		default:
		}
	}
}

func (o *OANDAAdapter) Unsubscribe(symbols []string) error {
	return nil
}

func (o *OANDAAdapter) Disconnect() error {
	o.mu.Lock()
	o.connected = false
	o.mu.Unlock()

	close(o.stopChan)
	return nil
}
