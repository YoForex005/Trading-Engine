package core

import (
	"log"
	"sync"
	"time"
)

// PnLEngine handles real-time P/L calculations
type PnLEngine struct {
	mu          sync.RWMutex
	engine      *Engine
	updateChan  chan AccountUpdate
	stopChan    chan struct{}
	subscribers map[int64][]chan AccountUpdate // accountID -> channels
}

// AccountUpdate contains real-time account data
type AccountUpdate struct {
	AccountID     int64            `json:"accountId"`
	Balance       float64          `json:"balance"`
	Equity        float64          `json:"equity"`
	Margin        float64          `json:"margin"`
	FreeMargin    float64          `json:"freeMargin"`
	MarginLevel   float64          `json:"marginLevel"`
	UnrealizedPnL float64          `json:"unrealizedPnL"`
	Positions     []PositionUpdate `json:"positions,omitempty"`
	Timestamp     time.Time        `json:"timestamp"`
}

// PositionUpdate contains real-time position data
type PositionUpdate struct {
	ID            int64   `json:"id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Volume        float64 `json:"volume"`
	OpenPrice     float64 `json:"openPrice"`
	CurrentPrice  float64 `json:"currentPrice"`
	UnrealizedPnL float64 `json:"unrealizedPnL"`
}

// NewPnLEngine creates a P/L calculation engine
func NewPnLEngine(engine *Engine) *PnLEngine {
	pnl := &PnLEngine{
		engine:      engine,
		updateChan:  make(chan AccountUpdate, 100),
		stopChan:    make(chan struct{}),
		subscribers: make(map[int64][]chan AccountUpdate),
	}

	go pnl.run()

	log.Println("[PnL Engine] Started")
	return pnl
}

// Subscribe adds a subscriber for account updates
func (p *PnLEngine) Subscribe(accountID int64, ch chan AccountUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[accountID] = append(p.subscribers[accountID], ch)
}

// Unsubscribe removes a subscriber
func (p *PnLEngine) Unsubscribe(accountID int64, ch chan AccountUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()

	subs := p.subscribers[accountID]
	for i, sub := range subs {
		if sub == ch {
			p.subscribers[accountID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}

// GetUpdateChannel returns the broadcast channel
func (p *PnLEngine) GetUpdateChannel() <-chan AccountUpdate {
	return p.updateChan
}

// run is the main loop
func (p *PnLEngine) run() {
	ticker := time.NewTicker(200 * time.Millisecond) // Update 5x per second
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-ticker.C:
			p.calculate()
		}
	}
}

// calculate updates all positions and broadcasts
func (p *PnLEngine) calculate() {
	// Update position prices from market data
	p.engine.UpdatePositionPrices()

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Get all accounts with open positions
	accountIDs := make(map[int64]bool)
	for _, pos := range p.engine.GetAllPositions() {
		accountIDs[pos.AccountID] = true
	}

	// Also include accounts with subscribers
	for accID := range p.subscribers {
		accountIDs[accID] = true
	}

	// Calculate and broadcast for each account
	for accountID := range accountIDs {
		summary, err := p.engine.GetAccountSummary(accountID)
		if err != nil {
			continue
		}

		// Build position updates
		positions := p.engine.GetPositions(accountID)
		posUpdates := make([]PositionUpdate, len(positions))
		for i, pos := range positions {
			posUpdates[i] = PositionUpdate{
				ID:            pos.ID,
				Symbol:        pos.Symbol,
				Side:          pos.Side,
				Volume:        pos.Volume,
				OpenPrice:     pos.OpenPrice,
				CurrentPrice:  pos.CurrentPrice,
				UnrealizedPnL: pos.UnrealizedPnL,
			}
		}

		update := AccountUpdate{
			AccountID:     accountID,
			Balance:       summary.Balance,
			Equity:        summary.Equity,
			Margin:        summary.Margin,
			FreeMargin:    summary.FreeMargin,
			MarginLevel:   summary.MarginLevel,
			UnrealizedPnL: summary.UnrealizedPnL,
			Positions:     posUpdates,
			Timestamp:     time.Now(),
		}

		// Send to general channel (non-blocking)
		select {
		case p.updateChan <- update:
		default:
		}

		// Send to specific subscribers
		for _, ch := range p.subscribers[accountID] {
			select {
			case ch <- update:
			default:
			}
		}
	}
}

// Stop stops the P/L engine
func (p *PnLEngine) Stop() {
	close(p.stopChan)
}

// ForceUpdate triggers an immediate update
func (p *PnLEngine) ForceUpdate() {
	p.calculate()
}
