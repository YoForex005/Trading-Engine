package orders

import (
	"log"
	"math"
	"sync"
	"time"
)

// TrailingStopType defines the type of trailing stop
type TrailingStopType string

const (
	TrailingFixed TrailingStopType = "FIXED" // Fixed pip distance
	TrailingStep  TrailingStopType = "STEP"  // Stepped trailing
	TrailingATR   TrailingStopType = "ATR"   // ATR-based
)

// TrailingStop represents an active trailing stop
type TrailingStop struct {
	TradeID      string           `json:"tradeId"`
	Symbol       string           `json:"symbol"`
	Side         string           `json:"side"` // BUY or SELL
	Distance     float64          `json:"distance"` // In pips or ATR multiplier
	StepSize     float64          `json:"stepSize,omitempty"` // For stepped trailing
	Type         TrailingStopType `json:"type"`
	CurrentSL    float64          `json:"currentSL"`
	HighestPrice float64          `json:"highestPrice"` // For longs
	LowestPrice  float64          `json:"lowestPrice"`  // For shorts
	Active       bool             `json:"active"`
}

// TrailingStopService manages trailing stops
type TrailingStopService struct {
	mu            sync.RWMutex
	trailingStops map[string]*TrailingStop
	priceCallback func(symbol string) (bid, ask float64, ok bool)
	modifySLCallback func(tradeID string, newSL float64) error
	atrCallback   func(symbol string, period int) float64
}

// NewTrailingStopService creates a new trailing stop service
func NewTrailingStopService() *TrailingStopService {
	svc := &TrailingStopService{
		trailingStops: make(map[string]*TrailingStop),
	}

	go svc.processLoop()
	
	log.Println("[TrailingStopService] Initialized")
	return svc
}

// SetPriceCallback sets the price fetcher
func (s *TrailingStopService) SetPriceCallback(fn func(symbol string) (bid, ask float64, ok bool)) {
	s.priceCallback = fn
}

// SetModifySLCallback sets the SL modification function
func (s *TrailingStopService) SetModifySLCallback(fn func(tradeID string, newSL float64) error) {
	s.modifySLCallback = fn
}

// SetATRCallback sets the ATR calculator
func (s *TrailingStopService) SetATRCallback(fn func(symbol string, period int) float64) {
	s.atrCallback = fn
}

// SetTrailingStop adds or updates a trailing stop
func (s *TrailingStopService) SetTrailingStop(tradeID, symbol, side string, tsType TrailingStopType, distance float64, stepSize float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current price to initialize
	var initialPrice float64
	if s.priceCallback != nil {
		bid, ask, ok := s.priceCallback(symbol)
		if ok {
			if side == "BUY" {
				initialPrice = bid
			} else {
				initialPrice = ask
			}
		}
	}

	ts := &TrailingStop{
		TradeID:      tradeID,
		Symbol:       symbol,
		Side:         side,
		Type:         tsType,
		Distance:     distance,
		StepSize:     stepSize,
		HighestPrice: initialPrice,
		LowestPrice:  initialPrice,
		Active:       true,
	}

	// Calculate initial SL
	pipValue := getPipValue(symbol)
	if side == "BUY" {
		ts.CurrentSL = initialPrice - (distance * pipValue)
	} else {
		ts.CurrentSL = initialPrice + (distance * pipValue)
	}

	s.trailingStops[tradeID] = ts
	log.Printf("[TrailingStop] Set for %s: %s %.1f pips", tradeID, tsType, distance)
	return nil
}

// RemoveTrailingStop removes a trailing stop
func (s *TrailingStopService) RemoveTrailingStop(tradeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.trailingStops, tradeID)
}

// GetTrailingStop returns a trailing stop
func (s *TrailingStopService) GetTrailingStop(tradeID string) (*TrailingStop, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ts, ok := s.trailingStops[tradeID]
	return ts, ok
}

// GetAllTrailingStops returns all active trailing stops
func (s *TrailingStopService) GetAllTrailingStops() []*TrailingStop {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stops := make([]*TrailingStop, 0, len(s.trailingStops))
	for _, ts := range s.trailingStops {
		stops = append(stops, ts)
	}
	return stops
}

func (s *TrailingStopService) processLoop() {
	ticker := time.NewTicker(200 * time.Millisecond)
	for range ticker.C {
		s.processTrailingStops()
	}
}

func (s *TrailingStopService) processTrailingStops() {
	if s.priceCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ts := range s.trailingStops {
		if !ts.Active {
			continue
		}

		bid, ask, ok := s.priceCallback(ts.Symbol)
		if !ok {
			continue
		}

		pipValue := getPipValue(ts.Symbol)
		distance := ts.Distance

		// For ATR-based, recalculate distance
		if ts.Type == TrailingATR && s.atrCallback != nil {
			atr := s.atrCallback(ts.Symbol, 14)
			distance = atr * ts.Distance // Distance is ATR multiplier
		}

		var newSL float64
		shouldUpdate := false

		if ts.Side == "BUY" {
			// For longs: trail below price
			if bid > ts.HighestPrice {
				ts.HighestPrice = bid
				newSL = bid - (distance * pipValue)
				
				if ts.Type == TrailingStep {
					// Step trailing: only update in steps
					stepPips := ts.StepSize * pipValue
					newSL = math.Floor(newSL/stepPips) * stepPips
				}
				
				if newSL > ts.CurrentSL {
					shouldUpdate = true
				}
			}
		} else {
			// For shorts: trail above price
			if ask < ts.LowestPrice || ts.LowestPrice == 0 {
				ts.LowestPrice = ask
				newSL = ask + (distance * pipValue)
				
				if ts.Type == TrailingStep {
					stepPips := ts.StepSize * pipValue
					newSL = math.Ceil(newSL/stepPips) * stepPips
				}
				
				if newSL < ts.CurrentSL || ts.CurrentSL == 0 {
					shouldUpdate = true
				}
			}
		}

		if shouldUpdate && newSL > 0 {
			ts.CurrentSL = newSL
			log.Printf("[TrailingStop] Updated %s: new SL = %.5f", ts.TradeID, newSL)
			
			if s.modifySLCallback != nil {
				go s.modifySLCallback(ts.TradeID, newSL)
			}
		}
	}
}

// getPipValue returns the pip value for a symbol
func getPipValue(symbol string) float64 {
	// JPY pairs have 0.01 pip value
	if len(symbol) >= 6 && symbol[3:6] == "JPY" {
		return 0.01
	}
	// Gold
	if symbol == "XAUUSD" {
		return 0.1
	}
	// Indices
	if symbol == "US30USD" || symbol == "SPX500USD" || symbol == "NAS100USD" {
		return 1.0
	}
	// Default forex
	return 0.0001
}
