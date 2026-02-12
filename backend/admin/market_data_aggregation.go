package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================
// Data Structures
// ============================================

// DataSource represents a market data source
type DataSource struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`          // "LP", "exchange", "aggregator"
	Status        string    `json:"status"`        // "active", "inactive", "error"
	URL           string    `json:"url"`
	Protocol      string    `json:"protocol"`      // "FIX", "REST", "WS"
	Symbols       []string  `json:"symbols"`
	LatencyMs     float64   `json:"latency_ms"`
	UptimePercent float64   `json:"uptime_percent"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Priority      int       `json:"priority"`      // 1=highest, 10=lowest
	Weight        float64   `json:"weight"`        // for weighted average (0.0-1.0)
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PriceAggregationRule represents how prices are aggregated for a symbol
type PriceAggregationRule struct {
	ID                int64            `json:"id"`
	Symbol            string           `json:"symbol"`
	Method            string           `json:"method"` // "best_bid_ask", "weighted_average", "primary_failover", "vwap"
	Sources           []DataSourceRef  `json:"sources"`
	FallbackOrder     []int            `json:"fallback_order"` // source IDs in priority order
	MaxSpreadPips     float64          `json:"max_spread_pips"`
	StaleThresholdMs  int              `json:"stale_threshold_ms"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// DataSourceRef represents a reference to a data source in aggregation
type DataSourceRef struct {
	SourceID int64   `json:"source_id"`
	Name     string  `json:"name"`
	Weight   float64 `json:"weight,omitempty"`
}

// AggregatedQuote represents an aggregated price quote
type AggregatedQuote struct {
	Symbol      string        `json:"symbol"`
	BestBid     float64       `json:"best_bid"`
	BestAsk     float64       `json:"best_ask"`
	Spread      float64       `json:"spread"`       // in pips
	SourceCount int           `json:"source_count"`
	Sources     []QuoteSource `json:"sources"`
	Timestamp   time.Time     `json:"timestamp"`
	IsStale     bool          `json:"is_stale"`
}

// QuoteSource represents a quote from a specific source
type QuoteSource struct {
	SourceID  int64   `json:"source_id"`
	Name      string  `json:"name"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
}

// DataSourceHealth represents health metrics for a data source
type DataSourceHealth struct {
	SourceID        int64     `json:"source_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	LatencyMs       float64   `json:"latency_ms"`
	MessagesPerSec  float64   `json:"messages_per_sec"`
	ErrorRate       float64   `json:"error_rate"`    // percentage
	LastError       string    `json:"last_error,omitempty"`
	Uptime24h       float64   `json:"uptime_24h"`    // percentage
	UptimeWeek      float64   `json:"uptime_week"`   // percentage
	LastHeartbeat   time.Time `json:"last_heartbeat"`
}

// CreateDataSourceRequest represents request to add a new data source
type CreateDataSourceRequest struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	URL      string   `json:"url"`
	Protocol string   `json:"protocol"`
	Symbols  []string `json:"symbols"`
	Priority int      `json:"priority"`
	Weight   float64  `json:"weight"`
}

// UpdateDataSourceRequest represents request to update a data source
type UpdateDataSourceRequest struct {
	Priority int     `json:"priority,omitempty"`
	Weight   float64 `json:"weight,omitempty"`
	Status   string  `json:"status,omitempty"`
}

// UpdateAggregationRuleRequest represents request to update aggregation rule
type UpdateAggregationRuleRequest struct {
	Method           string  `json:"method,omitempty"`
	Sources          []int64 `json:"sources,omitempty"` // source IDs
	MaxSpreadPips    float64 `json:"max_spread_pips,omitempty"`
	StaleThresholdMs int     `json:"stale_threshold_ms,omitempty"`
}

// MarketDataStats represents overall market data statistics
type MarketDataStats struct {
	TotalSources     int                       `json:"total_sources"`
	ActiveSources    int                       `json:"active_sources"`
	InactiveSources  int                       `json:"inactive_sources"`
	ErrorSources     int                       `json:"error_sources"`
	AvgLatencyMs     float64                   `json:"avg_latency_ms"`
	TotalSymbols     int                       `json:"total_symbols"`
	StaleQuoteCount  int                       `json:"stale_quote_count"`
	SourcesByType    map[string]int            `json:"sources_by_type"`
	SourcesByProtocol map[string]int           `json:"sources_by_protocol"`
	HealthBySource   map[string]DataSourceHealth `json:"health_by_source"`
}

// ============================================
// Service
// ============================================

// MarketDataAggregationService manages market data sources and aggregation
type MarketDataAggregationService struct {
	mu                sync.RWMutex
	sources           map[int64]*DataSource
	aggregationRules  map[string]*PriceAggregationRule // key: symbol
	quotes            map[string]*AggregatedQuote      // key: symbol
	nextSourceID      int64
	nextRuleID        int64
}

// NewMarketDataAggregationService creates a new market data aggregation service
func NewMarketDataAggregationService() *MarketDataAggregationService {
	s := &MarketDataAggregationService{
		sources:          make(map[int64]*DataSource),
		aggregationRules: make(map[string]*PriceAggregationRule),
		quotes:           make(map[string]*AggregatedQuote),
		nextSourceID:     7,
		nextRuleID:       31,
	}

	// Initialize 6 data sources
	s.initDataSources()

	// Initialize 30 aggregation rules for major symbols
	s.initAggregationRules()

	// Generate mock aggregated quotes
	s.generateAggregatedQuotes()

	log.Printf("[MarketDataAggregationService] Initialized with %d data sources, %d aggregation rules, %d quotes",
		len(s.sources), len(s.aggregationRules), len(s.quotes))

	return s
}

func (s *MarketDataAggregationService) initDataSources() {
	now := time.Now()

	sources := []*DataSource{
		{
			ID:            1,
			Name:          "YOFX Primary",
			Type:          "LP",
			Status:        "active",
			URL:           "fix://yofx.com:4011",
			Protocol:      "FIX",
			Symbols:       []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "USDCHF", "NZDUSD", "EURGBP", "EURJPY", "GBPJPY"},
			LatencyMs:     15.5,
			UptimePercent: 99.8,
			LastHeartbeat: now.Add(-2 * time.Second),
			Priority:      1,
			Weight:        0.4,
			CreatedAt:     now.Add(-180 * 24 * time.Hour),
			UpdatedAt:     now.Add(-5 * time.Hour),
		},
		{
			ID:            2,
			Name:          "YOFX Backup",
			Type:          "LP",
			Status:        "active",
			URL:           "fix://yofx-backup.com:4012",
			Protocol:      "FIX",
			Symbols:       []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "USDCHF", "NZDUSD", "EURGBP", "EURJPY", "GBPJPY"},
			LatencyMs:     18.2,
			UptimePercent: 99.5,
			LastHeartbeat: now.Add(-3 * time.Second),
			Priority:      2,
			Weight:        0.3,
			CreatedAt:     now.Add(-180 * 24 * time.Hour),
			UpdatedAt:     now.Add(-10 * time.Hour),
		},
		{
			ID:            3,
			Name:          "Currenex",
			Type:          "aggregator",
			Status:        "active",
			URL:           "wss://currenex.com/stream",
			Protocol:      "WS",
			Symbols:       []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "XAUUSD", "XAGUSD"},
			LatencyMs:     22.8,
			UptimePercent: 99.2,
			LastHeartbeat: now.Add(-5 * time.Second),
			Priority:      3,
			Weight:        0.15,
			CreatedAt:     now.Add(-150 * 24 * time.Hour),
			UpdatedAt:     now.Add(-15 * time.Hour),
		},
		{
			ID:            4,
			Name:          "LMAX",
			Type:          "exchange",
			Status:        "active",
			URL:           "wss://lmax.com/feed",
			Protocol:      "WS",
			Symbols:       []string{"EURUSD", "GBPUSD", "USDJPY", "BTCUSD", "ETHUSD"},
			LatencyMs:     12.3,
			UptimePercent: 99.9,
			LastHeartbeat: now.Add(-1 * time.Second),
			Priority:      1,
			Weight:        0.1,
			CreatedAt:     now.Add(-120 * 24 * time.Hour),
			UpdatedAt:     now.Add(-7 * time.Hour),
		},
		{
			ID:            5,
			Name:          "PrimeXM",
			Type:          "LP",
			Status:        "inactive",
			URL:           "https://api.primexm.com/quotes",
			Protocol:      "REST",
			Symbols:       []string{"EURUSD", "GBPUSD", "USDJPY"},
			LatencyMs:     45.0,
			UptimePercent: 98.5,
			LastHeartbeat: now.Add(-2 * time.Minute),
			Priority:      5,
			Weight:        0.05,
			CreatedAt:     now.Add(-90 * 24 * time.Hour),
			UpdatedAt:     now.Add(-3 * time.Hour),
		},
		{
			ID:            6,
			Name:          "Internal B-Book",
			Type:          "LP",
			Status:        "active",
			URL:           "internal://bbook",
			Protocol:      "WS",
			Symbols:       []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "XAUUSD", "BTCUSD", "ETHUSD"},
			LatencyMs:     2.5,
			UptimePercent: 99.95,
			LastHeartbeat: now,
			Priority:      10,
			Weight:        0.0, // B-Book is fallback only
			CreatedAt:     now.Add(-365 * 24 * time.Hour),
			UpdatedAt:     now.Add(-1 * time.Hour),
		},
	}

	for _, src := range sources {
		s.sources[src.ID] = src
	}
}

func (s *MarketDataAggregationService) initAggregationRules() {
	now := time.Now()

	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "USDCHF", "NZDUSD", "EURGBP", "EURJPY", "GBPJPY",
		"AUDJPY", "EURAUD", "EURCHF", "GBPAUD", "GBPCAD", "GBPCHF", "AUDCAD", "AUDCHF", "AUDNZD", "CADCHF",
		"CADJPY", "CHFJPY", "EURCAD", "NZDJPY", "XAUUSD", "XAGUSD", "BTCUSD", "ETHUSD", "SPX500", "US30",
	}

	methods := []string{"best_bid_ask", "weighted_average", "primary_failover", "vwap"}

	for i, symbol := range symbols {
		method := methods[i%len(methods)]

		// Determine which sources provide this symbol
		sourceRefs := make([]DataSourceRef, 0)
		fallbackOrder := make([]int, 0)

		for _, src := range s.sources {
			if containsString(src.Symbols, symbol) {
				sourceRefs = append(sourceRefs, DataSourceRef{
					SourceID: src.ID,
					Name:     src.Name,
					Weight:   src.Weight,
				})
				fallbackOrder = append(fallbackOrder, int(src.ID))
			}
		}

		rule := &PriceAggregationRule{
			ID:               int64(i + 1),
			Symbol:           symbol,
			Method:           method,
			Sources:          sourceRefs,
			FallbackOrder:    fallbackOrder,
			MaxSpreadPips:    3.0 + float64(i%5), // 3-7 pips
			StaleThresholdMs: 1000 + (i%3)*500,    // 1000-2000ms
			UpdatedAt:        now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour),
		}

		s.aggregationRules[symbol] = rule
	}
}

func (s *MarketDataAggregationService) generateAggregatedQuotes() {
	now := time.Now()

	basePrices := map[string]float64{
		"EURUSD": 1.08500, "GBPUSD": 1.26500, "USDJPY": 149.500, "AUDUSD": 0.63500,
		"USDCAD": 1.36500, "USDCHF": 0.88500, "NZDUSD": 0.58500, "EURGBP": 0.85500,
		"EURJPY": 162.000, "GBPJPY": 189.000, "AUDJPY": 95.000, "EURAUD": 1.70500,
		"EURCHF": 0.96000, "GBPAUD": 1.99000, "GBPCAD": 1.72500, "GBPCHF": 1.12000,
		"AUDCAD": 0.86500, "AUDCHF": 0.56500, "AUDNZD": 1.08500, "CADCHF": 0.65000,
		"CADJPY": 109.500, "CHFJPY": 169.000, "EURCAD": 1.48000, "NZDJPY": 87.500,
		"XAUUSD": 2050.00, "XAGUSD": 24.50, "BTCUSD": 43500.0, "ETHUSD": 2300.0,
		"SPX500": 4750.0, "US30": 37800.0,
	}

	for symbol, basePrice := range basePrices {
		rule, exists := s.aggregationRules[symbol]
		if !exists {
			continue
		}

		// Generate quotes from multiple sources
		sources := make([]QuoteSource, 0)
		for _, srcRef := range rule.Sources {
			src := s.sources[srcRef.SourceID]
			if src == nil || src.Status != "active" {
				continue
			}

			spread := 0.0001 + rand.Float64()*0.0003 // 1-4 pips for forex
			if strings.Contains(symbol, "JPY") {
				spread *= 100
			}
			if strings.HasPrefix(symbol, "XAU") || strings.HasPrefix(symbol, "XAG") {
				spread = 0.5 + rand.Float64()*1.5
			}
			if strings.Contains(symbol, "BTC") || strings.Contains(symbol, "ETH") {
				spread = 10 + rand.Float64()*20
			}

			bid := basePrice - spread/2 + (rand.Float64()-0.5)*spread
			ask := bid + spread

			sources = append(sources, QuoteSource{
				SourceID:  src.ID,
				Name:      src.Name,
				Bid:       bid,
				Ask:       ask,
				Timestamp: now.Add(time.Duration(-rand.Intn(100)) * time.Millisecond),
			})
		}

		if len(sources) == 0 {
			continue
		}

		// Aggregate based on method
		var bestBid, bestAsk float64
		switch rule.Method {
		case "best_bid_ask":
			bestBid = sources[0].Bid
			bestAsk = sources[0].Ask
			for _, src := range sources {
				if src.Bid > bestBid {
					bestBid = src.Bid
				}
				if src.Ask < bestAsk {
					bestAsk = src.Ask
				}
			}
		case "weighted_average":
			totalWeight := 0.0
			weightedBid := 0.0
			weightedAsk := 0.0
			for _, src := range sources {
				srcData := s.sources[src.SourceID]
				if srcData != nil {
					weight := srcData.Weight
					if weight > 0 {
						weightedBid += src.Bid * weight
						weightedAsk += src.Ask * weight
						totalWeight += weight
					}
				}
			}
			if totalWeight > 0 {
				bestBid = weightedBid / totalWeight
				bestAsk = weightedAsk / totalWeight
			}
		case "primary_failover":
			// Use first active source in fallback order
			bestBid = sources[0].Bid
			bestAsk = sources[0].Ask
		case "vwap":
			// Simplified VWAP (volume-weighted average price)
			bestBid = sources[0].Bid
			bestAsk = sources[0].Ask
			for _, src := range sources {
				bestBid += src.Bid
				bestAsk += src.Ask
			}
			bestBid /= float64(len(sources))
			bestAsk /= float64(len(sources))
		}

		spread := (bestAsk - bestBid) / 0.0001 // in pips for forex
		if strings.Contains(symbol, "JPY") {
			spread /= 100
		}

		isStale := rand.Float64() < 0.05 // 5% stale quotes

		s.quotes[symbol] = &AggregatedQuote{
			Symbol:      symbol,
			BestBid:     bestBid,
			BestAsk:     bestAsk,
			Spread:      spread,
			SourceCount: len(sources),
			Sources:     sources,
			Timestamp:   now,
			IsStale:     isStale,
		}
	}
}

func containsString(arr []string, val string) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

// GetAllSources returns all data sources
func (s *MarketDataAggregationService) GetAllSources() []*DataSource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sources := make([]*DataSource, 0, len(s.sources))
	for _, src := range s.sources {
		sources = append(sources, src)
	}
	return sources
}

// CreateSource creates a new data source
func (s *MarketDataAggregationService) CreateSource(req CreateDataSourceRequest) (*DataSource, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	source := &DataSource{
		ID:            s.nextSourceID,
		Name:          req.Name,
		Type:          req.Type,
		Status:        "inactive", // start as inactive
		URL:           req.URL,
		Protocol:      req.Protocol,
		Symbols:       req.Symbols,
		LatencyMs:     0,
		UptimePercent: 0,
		LastHeartbeat: now,
		Priority:      req.Priority,
		Weight:        req.Weight,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.sources[source.ID] = source
	s.nextSourceID++

	return source, nil
}

// GetSourceByID returns a data source by ID
func (s *MarketDataAggregationService) GetSourceByID(id int64) *DataSource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sources[id]
}

// GetSourceHealth returns health metrics for a data source
func (s *MarketDataAggregationService) GetSourceHealth(id int64) *DataSourceHealth {
	s.mu.RLock()
	defer s.mu.RUnlock()

	src := s.sources[id]
	if src == nil {
		return nil
	}

	// Generate mock health data
	health := &DataSourceHealth{
		SourceID:        src.ID,
		Name:            src.Name,
		Status:          src.Status,
		LatencyMs:       src.LatencyMs,
		MessagesPerSec:  100 + rand.Float64()*900,  // 100-1000 msg/s
		ErrorRate:       rand.Float64() * 0.5,      // 0-0.5%
		Uptime24h:       99.0 + rand.Float64()*0.9, // 99-99.9%
		UptimeWeek:      src.UptimePercent,
		LastHeartbeat:   src.LastHeartbeat,
	}

	if src.Status == "error" {
		health.LastError = "Connection timeout after 30s"
	}

	return health
}

// UpdateSource updates a data source
func (s *MarketDataAggregationService) UpdateSource(id int64, req UpdateDataSourceRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	src := s.sources[id]
	if src == nil {
		return nil
	}

	if req.Priority > 0 {
		src.Priority = req.Priority
	}
	if req.Weight >= 0 {
		src.Weight = req.Weight
	}
	if req.Status != "" {
		src.Status = req.Status
	}

	src.UpdatedAt = time.Now()
	return nil
}

// GetAggregationRules returns all aggregation rules
func (s *MarketDataAggregationService) GetAggregationRules() []*PriceAggregationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]*PriceAggregationRule, 0, len(s.aggregationRules))
	for _, rule := range s.aggregationRules {
		rules = append(rules, rule)
	}
	return rules
}

// UpdateAggregationRule updates an aggregation rule for a symbol
func (s *MarketDataAggregationService) UpdateAggregationRule(symbol string, req UpdateAggregationRuleRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule := s.aggregationRules[symbol]
	if rule == nil {
		return nil
	}

	if req.Method != "" {
		rule.Method = req.Method
	}
	if len(req.Sources) > 0 {
		// Rebuild source refs
		sourceRefs := make([]DataSourceRef, 0)
		for _, srcID := range req.Sources {
			if src := s.sources[srcID]; src != nil {
				sourceRefs = append(sourceRefs, DataSourceRef{
					SourceID: srcID,
					Name:     src.Name,
					Weight:   src.Weight,
				})
			}
		}
		rule.Sources = sourceRefs
	}
	if req.MaxSpreadPips > 0 {
		rule.MaxSpreadPips = req.MaxSpreadPips
	}
	if req.StaleThresholdMs > 0 {
		rule.StaleThresholdMs = req.StaleThresholdMs
	}

	rule.UpdatedAt = time.Now()
	return nil
}

// GetAggregatedQuotes returns all aggregated quotes
func (s *MarketDataAggregationService) GetAggregatedQuotes() []*AggregatedQuote {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quotes := make([]*AggregatedQuote, 0, len(s.quotes))
	for _, quote := range s.quotes {
		quotes = append(quotes, quote)
	}
	return quotes
}

// GetStats returns overall market data statistics
func (s *MarketDataAggregationService) GetStats() MarketDataStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	activeSources := 0
	inactiveSources := 0
	errorSources := 0
	totalLatency := 0.0
	sourcesByType := make(map[string]int)
	sourcesByProtocol := make(map[string]int)
	healthBySource := make(map[string]DataSourceHealth)

	for _, src := range s.sources {
		switch src.Status {
		case "active":
			activeSources++
			totalLatency += src.LatencyMs
		case "inactive":
			inactiveSources++
		case "error":
			errorSources++
		}

		sourcesByType[src.Type]++
		sourcesByProtocol[src.Protocol]++

		health := DataSourceHealth{
			SourceID:        src.ID,
			Name:            src.Name,
			Status:          src.Status,
			LatencyMs:       src.LatencyMs,
			MessagesPerSec:  100 + rand.Float64()*900,
			ErrorRate:       rand.Float64() * 0.5,
			Uptime24h:       99.0 + rand.Float64()*0.9,
			UptimeWeek:      src.UptimePercent,
			LastHeartbeat:   src.LastHeartbeat,
		}
		healthBySource[src.Name] = health
	}

	avgLatency := 0.0
	if activeSources > 0 {
		avgLatency = totalLatency / float64(activeSources)
	}

	staleQuoteCount := 0
	for _, quote := range s.quotes {
		if quote.IsStale {
			staleQuoteCount++
		}
	}

	return MarketDataStats{
		TotalSources:      len(s.sources),
		ActiveSources:     activeSources,
		InactiveSources:   inactiveSources,
		ErrorSources:      errorSources,
		AvgLatencyMs:      avgLatency,
		TotalSymbols:      len(s.quotes),
		StaleQuoteCount:   staleQuoteCount,
		SourcesByType:     sourcesByType,
		SourcesByProtocol: sourcesByProtocol,
		HealthBySource:    healthBySource,
	}
}

// ============================================
// HTTP Handlers
// ============================================

// MarketDataAggregationHandler handles market data aggregation HTTP requests
type MarketDataAggregationHandler struct {
	service     *MarketDataAggregationService
	authService interface {
		ValidateAdminToken(r *http.Request) (int64, error)
	}
}

// NewMarketDataAggregationHandler creates a new market data aggregation handler
func NewMarketDataAggregationHandler(service *MarketDataAggregationService, authService interface {
	ValidateAdminToken(r *http.Request) (int64, error)
}) *MarketDataAggregationHandler {
	return &MarketDataAggregationHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetSources handles GET /admin/market-data/sources
func (h *MarketDataAggregationHandler) HandleGetSources(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sources := h.service.GetAllSources()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(sources)
}

// HandleCreateSource handles POST /admin/market-data/sources
func (h *MarketDataAggregationHandler) HandleCreateSource(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateDataSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	source, err := h.service.CreateSource(req)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(source)
}

// HandleGetSourceByID handles GET /admin/market-data/sources/:id
func (h *MarketDataAggregationHandler) HandleGetSourceByID(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/market-data/sources/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	source := h.service.GetSourceByID(id)
	if source == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	health := h.service.GetSourceHealth(id)

	response := map[string]interface{}{
		"source": source,
		"health": health,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateSource handles PUT /admin/market-data/sources/:id
func (h *MarketDataAggregationHandler) HandleUpdateSource(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/market-data/sources/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid source ID", http.StatusBadRequest)
		return
	}

	var req UpdateDataSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSource(id, req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Data source updated successfully",
	})
}

// HandleGetAggregationRules handles GET /admin/market-data/aggregation-rules
func (h *MarketDataAggregationHandler) HandleGetAggregationRules(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rules := h.service.GetAggregationRules()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(rules)
}

// HandleUpdateAggregationRule handles PUT /admin/market-data/aggregation-rules/:symbol
func (h *MarketDataAggregationHandler) HandleUpdateAggregationRule(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract symbol from URL
	symbol := strings.TrimPrefix(r.URL.Path, "/admin/market-data/aggregation-rules/")

	var req UpdateAggregationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAggregationRule(symbol, req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Aggregation rule updated successfully",
	})
}

// HandleGetQuotes handles GET /admin/market-data/quotes
func (h *MarketDataAggregationHandler) HandleGetQuotes(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	quotes := h.service.GetAggregatedQuotes()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(quotes)
}

// HandleGetStats handles GET /admin/market-data/stats
func (h *MarketDataAggregationHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
