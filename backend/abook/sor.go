package abook

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/lpmanager"
)

// SmartOrderRouter aggregates quotes from multiple LPs and selects best execution venue
type SmartOrderRouter struct {
	lpManager      *lpmanager.Manager
	quoteCache     map[string]*AggregatedQuote
	quoteCacheMu   sync.RWMutex
	lpHealthScore  map[string]*LPHealth
	healthMu       sync.RWMutex
}

// AggregatedQuote represents best bid/ask from multiple LPs
type AggregatedQuote struct {
	Symbol         string
	BestBid        float64
	BestAsk        float64
	BestBidLP      string
	BestAskLP      string
	BestBidSize    float64
	BestAskSize    float64
	LPQuotes       map[string]*LPQuote
	LastUpdate     time.Time
}

// LPQuote represents a quote from a specific LP
type LPQuote struct {
	LP         string
	Bid        float64
	Ask        float64
	BidSize    float64
	AskSize    float64
	Spread     float64
	Timestamp  time.Time
}

// LPSelection represents the selected LP for execution
type LPSelection struct {
	LPID         string
	SessionID    string
	Price        float64
	Reason       string
	HealthScore  float64
}

// LPHealth tracks LP health metrics
type LPHealth struct {
	LPID            string
	HealthScore     float64  // 0-1, higher is better
	FillRate        float64
	AvgSlippage     float64
	AvgLatency      time.Duration
	RejectRate      float64
	LastRejectTime  time.Time
	ConnectionState string
	LastUpdate      time.Time
}

// NewSmartOrderRouter creates a new smart order router
func NewSmartOrderRouter(lpManager *lpmanager.Manager) *SmartOrderRouter {
	sor := &SmartOrderRouter{
		lpManager:     lpManager,
		quoteCache:    make(map[string]*AggregatedQuote),
		lpHealthScore: make(map[string]*LPHealth),
	}

	// Start quote aggregation
	go sor.aggregateQuotes()

	// Start health monitoring
	go sor.monitorLPHealth()

	return sor
}

// SelectLP selects the best LP for order execution
func (s *SmartOrderRouter) SelectLP(symbol, side string, volume float64) (*LPSelection, error) {
	// 1. Get aggregated quotes
	s.quoteCacheMu.RLock()
	quote, exists := s.quoteCache[symbol]
	s.quoteCacheMu.RUnlock()

	if !exists || time.Since(quote.LastUpdate) > 5*time.Second {
		return nil, fmt.Errorf("no recent quote for %s", symbol)
	}

	// 2. Determine target LP based on side
	var targetLP string
	var targetPrice float64

	if side == "BUY" {
		// For BUY, we want the best ASK (lowest)
		targetLP = quote.BestAskLP
		targetPrice = quote.BestAsk
	} else {
		// For SELL, we want the best BID (highest)
		targetLP = quote.BestBidLP
		targetPrice = quote.BestBid
	}

	if targetLP == "" {
		return nil, errors.New("no LP available for execution")
	}

	// 3. Check LP health
	s.healthMu.RLock()
	health, exists := s.lpHealthScore[targetLP]
	s.healthMu.RUnlock()

	if !exists {
		health = &LPHealth{
			LPID:        targetLP,
			HealthScore: 0.8, // Default
		}
	}

	// 4. If health score is too low, try alternative LP
	if health.HealthScore < 0.5 {
		log.Printf("[SOR] Primary LP %s has low health (%.2f), finding alternative",
			targetLP, health.HealthScore)

		altLP, altPrice := s.findAlternativeLP(symbol, side, targetLP, quote)
		if altLP != "" {
			targetLP = altLP
			targetPrice = altPrice
		}
	}

	// 5. Map LP ID to FIX session ID
	sessionID := s.mapLPToSession(targetLP)
	if sessionID == "" {
		return nil, fmt.Errorf("no FIX session for LP %s", targetLP)
	}

	selection := &LPSelection{
		LPID:        targetLP,
		SessionID:   sessionID,
		Price:       targetPrice,
		Reason:      fmt.Sprintf("Best %s price", side),
		HealthScore: health.HealthScore,
	}

	log.Printf("[SOR] Selected %s for %s %s @ %.5f (health: %.2f)",
		targetLP, side, symbol, targetPrice, health.HealthScore)

	return selection, nil
}

// GetAggregatedQuote returns the aggregated quote for a symbol
func (s *SmartOrderRouter) GetAggregatedQuote(symbol string) (*AggregatedQuote, error) {
	s.quoteCacheMu.RLock()
	defer s.quoteCacheMu.RUnlock()

	quote, exists := s.quoteCache[symbol]
	if !exists {
		return nil, fmt.Errorf("no quote for %s", symbol)
	}

	if time.Since(quote.LastUpdate) > 10*time.Second {
		return nil, fmt.Errorf("stale quote for %s", symbol)
	}

	return quote, nil
}

// GetLPHealth returns health metrics for an LP
func (s *SmartOrderRouter) GetLPHealth(lpID string) (*LPHealth, error) {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()

	health, exists := s.lpHealthScore[lpID]
	if !exists {
		return nil, fmt.Errorf("no health data for %s", lpID)
	}

	return health, nil
}

// UpdateLPHealth updates health metrics for an LP
func (s *SmartOrderRouter) UpdateLPHealth(lpID string, fillRate, avgSlippage float64, avgLatency time.Duration) {
	s.healthMu.Lock()
	defer s.healthMu.Unlock()

	health, exists := s.lpHealthScore[lpID]
	if !exists {
		health = &LPHealth{
			LPID: lpID,
		}
		s.lpHealthScore[lpID] = health
	}

	health.FillRate = fillRate
	health.AvgSlippage = avgSlippage
	health.AvgLatency = avgLatency

	// Calculate health score (0-1)
	// Components:
	// - Fill rate: 40%
	// - Slippage: 30% (lower is better)
	// - Latency: 20% (lower is better)
	// - Reject rate: 10% (lower is better)

	fillScore := fillRate // Already 0-1
	slippageScore := 1.0 - min(abs(avgSlippage)*100, 1.0) // Convert pips to score
	latencyScore := 1.0 - min(float64(avgLatency.Milliseconds())/1000.0, 1.0)
	rejectScore := 1.0 - health.RejectRate

	health.HealthScore = (fillScore * 0.4) + (slippageScore * 0.3) + (latencyScore * 0.2) + (rejectScore * 0.1)
	health.LastUpdate = time.Now()

	log.Printf("[SOR] Updated health for %s: score=%.2f fill=%.2f slippage=%.5f latency=%v",
		lpID, health.HealthScore, fillRate, avgSlippage, avgLatency)
}

// RecordReject records a rejection from an LP
func (s *SmartOrderRouter) RecordReject(lpID string) {
	s.healthMu.Lock()
	defer s.healthMu.Unlock()

	health, exists := s.lpHealthScore[lpID]
	if !exists {
		health = &LPHealth{
			LPID: lpID,
		}
		s.lpHealthScore[lpID] = health
	}

	health.RejectRate = min(health.RejectRate+0.05, 1.0)
	health.LastRejectTime = time.Now()

	// Recalculate health score
	fillScore := health.FillRate
	slippageScore := 1.0 - min(abs(health.AvgSlippage)*100, 1.0)
	latencyScore := 1.0 - min(float64(health.AvgLatency.Milliseconds())/1000.0, 1.0)
	rejectScore := 1.0 - health.RejectRate

	health.HealthScore = (fillScore * 0.4) + (slippageScore * 0.3) + (latencyScore * 0.2) + (rejectScore * 0.1)
}

// aggregateQuotes continuously aggregates quotes from all LPs
func (s *SmartOrderRouter) aggregateQuotes() {
	log.Println("[SOR] Quote aggregation started")

	quotesChan := s.lpManager.GetQuotesChan()

	for quote := range quotesChan {
		s.quoteCacheMu.Lock()

		aggQuote, exists := s.quoteCache[quote.Symbol]
		if !exists {
			aggQuote = &AggregatedQuote{
				Symbol:   quote.Symbol,
				LPQuotes: make(map[string]*LPQuote),
			}
			s.quoteCache[quote.Symbol] = aggQuote
		}

		// Add LP quote
		lpQuote := &LPQuote{
			LP:        quote.LP,
			Bid:       quote.Bid,
			Ask:       quote.Ask,
			BidSize:   1000000, // TODO: Get from LP
			AskSize:   1000000,
			Spread:    quote.Ask - quote.Bid,
			Timestamp: time.Unix(0, quote.Timestamp*int64(time.Millisecond)),
		}

		aggQuote.LPQuotes[quote.LP] = lpQuote

		// Recalculate best bid/ask
		s.recalculateBestPrices(aggQuote)

		aggQuote.LastUpdate = time.Now()

		s.quoteCacheMu.Unlock()
	}
}

// recalculateBestPrices finds the best bid and ask across all LPs
func (s *SmartOrderRouter) recalculateBestPrices(aggQuote *AggregatedQuote) {
	var bestBid, bestAsk float64
	var bestBidLP, bestAskLP string
	var bestBidSize, bestAskSize float64

	for lpID, lpQuote := range aggQuote.LPQuotes {
		// Skip stale quotes (older than 5 seconds)
		if time.Since(lpQuote.Timestamp) > 5*time.Second {
			continue
		}

		// Best BID (highest)
		if lpQuote.Bid > bestBid {
			bestBid = lpQuote.Bid
			bestBidLP = lpID
			bestBidSize = lpQuote.BidSize
		}

		// Best ASK (lowest)
		if bestAsk == 0 || lpQuote.Ask < bestAsk {
			bestAsk = lpQuote.Ask
			bestAskLP = lpID
			bestAskSize = lpQuote.AskSize
		}
	}

	aggQuote.BestBid = bestBid
	aggQuote.BestAsk = bestAsk
	aggQuote.BestBidLP = bestBidLP
	aggQuote.BestAskLP = bestAskLP
	aggQuote.BestBidSize = bestBidSize
	aggQuote.BestAskSize = bestAskSize
}

// findAlternativeLP finds an alternative LP if primary is unhealthy
func (s *SmartOrderRouter) findAlternativeLP(symbol, side, excludeLP string, quote *AggregatedQuote) (string, float64) {
	type lpCandidate struct {
		lpID   string
		price  float64
		health float64
	}

	candidates := make([]*lpCandidate, 0)

	for lpID, lpQuote := range quote.LPQuotes {
		if lpID == excludeLP {
			continue
		}

		// Skip stale quotes
		if time.Since(lpQuote.Timestamp) > 5*time.Second {
			continue
		}

		var price float64
		if side == "BUY" {
			price = lpQuote.Ask
		} else {
			price = lpQuote.Bid
		}

		// Get health score
		s.healthMu.RLock()
		health, exists := s.lpHealthScore[lpID]
		s.healthMu.RUnlock()

		healthScore := 0.5 // Default
		if exists {
			healthScore = health.HealthScore
		}

		candidates = append(candidates, &lpCandidate{
			lpID:   lpID,
			price:  price,
			health: healthScore,
		})
	}

	if len(candidates) == 0 {
		return "", 0
	}

	// Sort by health score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].health > candidates[j].health
	})

	// Return best healthy LP
	return candidates[0].lpID, candidates[0].price
}

// monitorLPHealth monitors LP connection health
func (s *SmartOrderRouter) monitorLPHealth() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := s.lpManager.GetStatus()

		s.healthMu.Lock()
		for lpID, lpStatus := range status {
			health, exists := s.lpHealthScore[lpID]
			if !exists {
				health = &LPHealth{
					LPID:        lpID,
					HealthScore: 0.5,
				}
				s.lpHealthScore[lpID] = health
			}

			if lpStatus.Connected {
				health.ConnectionState = "CONNECTED"
			} else {
				health.ConnectionState = "DISCONNECTED"
				health.HealthScore = 0 // No health if disconnected
			}
		}
		s.healthMu.Unlock()
	}
}

// mapLPToSession maps LP ID to FIX session ID
func (s *SmartOrderRouter) mapLPToSession(lpID string) string {
	// Map LP Manager IDs to FIX session IDs
	mapping := map[string]string{
		"oanda":         "YOFX1", // OANDA via YOFx
		"binance":       "YOFX1", // Binance also via YOFx (if supported)
		"lmax":          "LMAX_PROD",
		"currenex":      "CURRENEX",
		"integral":      "INTEGRAL",
		"ebs":           "EBS",
		"flexymarkets":  "YOFX1",
	}

	sessionID, exists := mapping[lpID]
	if !exists {
		// Default to first available session
		return "YOFX1"
	}

	return sessionID
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// abs is defined in risk.go, removed duplicate
