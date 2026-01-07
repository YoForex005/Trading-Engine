package adapters

import (
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/flexymarkets"
	"github.com/epic1st/rtx/backend/lpmanager"
)

type FlexyAdapter struct {
	id         string
	name       string
	client     *flexymarkets.Client
	quotesChan chan lpmanager.Quote
	mu         sync.RWMutex
	connected  bool
	symbols    []lpmanager.SymbolInfo
	lastTick   time.Time
}

func NewFlexyAdapter() *FlexyAdapter {
	return &FlexyAdapter{
		id:         "flexy",
		name:       "FlexyMarkets",
		client:     flexymarkets.NewClient(),
		quotesChan: make(chan lpmanager.Quote, 500),
		symbols: []lpmanager.SymbolInfo{
			{Symbol: "BTCUSD", DisplayName: "BTC/USD", MinLotSize: 0.01, Type: "crypto", PipValue: 1},
			{Symbol: "ETHUSD", DisplayName: "ETH/USD", MinLotSize: 0.01, Type: "crypto", PipValue: 1},
		},
	}
}

func (a *FlexyAdapter) ID() string   { return a.id }
func (a *FlexyAdapter) Name() string { return a.name }
func (a *FlexyAdapter) Type() string { return "socket.io" }

func (a *FlexyAdapter) Connect() error {
	log.Println("[FlexyAdapter] Connecting...")

	// Collect default symbols
	var syms []string
	for _, s := range a.symbols {
		syms = append(syms, s.Symbol)
	}

	// Connect with default symbols
	if err := a.client.Connect(syms); err != nil {
		return err
	}

	a.mu.Lock()
	a.connected = true
	a.mu.Unlock()

	// Start processing quotes in background
	go a.processQuotes()

	return nil
}

func (a *FlexyAdapter) Disconnect() error {
	a.client.Stop()
	a.mu.Lock()
	a.connected = false
	a.mu.Unlock()
	return nil
}

func (a *FlexyAdapter) IsConnected() bool {
	return a.client.IsConnected()
}

func (a *FlexyAdapter) GetSymbols() ([]lpmanager.SymbolInfo, error) {
	return a.symbols, nil
}

func (a *FlexyAdapter) Subscribe(symbols []string) error {
	return a.client.Subscribe(symbols)
}

func (a *FlexyAdapter) Unsubscribe(symbols []string) error {
	return nil
}

func (a *FlexyAdapter) GetQuotesChan() <-chan lpmanager.Quote {
	return a.quotesChan
}

func (a *FlexyAdapter) GetStatus() lpmanager.LPStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return lpmanager.LPStatus{
		ID:          a.id,
		Name:        a.name,
		Type:        "socket.io",
		Connected:   a.client.IsConnected(),
		Enabled:     true,
		SymbolCount: len(a.symbols),
		LastTick:    a.lastTick,
	}
}

func (a *FlexyAdapter) processQuotes() {
	for q := range a.client.GetQuotesChan() {
		quote := lpmanager.Quote{
			Symbol:    q.Symbol,
			Bid:       q.Bid,
			Ask:       q.Ask,
			Timestamp: q.Timestamp,
			LP:        a.id,
		}

		a.mu.Lock()
		a.lastTick = time.Now()
		a.mu.Unlock()

		select {
		case a.quotesChan <- quote:
		default:
		}
	}
}
