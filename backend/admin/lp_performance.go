package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

type LPPerformance struct {
	LPID            int64   `json:"lpId"`
	Name            string  `json:"name"`
	FillRate        float64 `json:"fillRate"`        // Percentage 0-100
	AvgLatency      float64 `json:"avgLatency"`      // milliseconds
	MinLatency      float64 `json:"minLatency"`      // milliseconds
	MaxLatency      float64 `json:"maxLatency"`      // milliseconds
	AvgSlippage     float64 `json:"avgSlippage"`     // pips
	RejectionRate   float64 `json:"rejectionRate"`   // Percentage 0-100
	QuoteQuality    float64 `json:"quoteQuality"`    // Score 0-100
	OrdersRouted    int     `json:"ordersRouted"`
	TotalVolume     float64 `json:"totalVolume"`     // USD
	Status          string  `json:"status"`          // "active", "inactive", "maintenance"
	LastUpdate      time.Time `json:"lastUpdate"`
}

type LPDetailedProfile struct {
	LPID               int64              `json:"lpId"`
	Name               string             `json:"name"`
	Status             string             `json:"status"`
	FillRate           float64            `json:"fillRate"`
	LatencyStats       LatencyStats       `json:"latencyStats"`
	SlippageDistribution SlippageStats    `json:"slippageDistribution"`
	RejectionBreakdown RejectionStats     `json:"rejectionBreakdown"`
	QuoteQuality       float64            `json:"quoteQuality"`
	OrdersRouted       int                `json:"ordersRouted"`
	TotalVolume        float64            `json:"totalVolume"`
	TopSymbols         []SymbolPerformance `json:"topSymbols"`
	LastUpdate         time.Time          `json:"lastUpdate"`
}

type LatencyStats struct {
	Avg float64 `json:"avg"` // milliseconds
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	P50 float64 `json:"p50"` // median
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

type SlippageStats struct {
	AvgSlippage      float64 `json:"avgSlippage"` // pips
	PositivePercent  float64 `json:"positivePercent"` // % of orders with positive slippage
	NegativePercent  float64 `json:"negativePercent"` // % of orders with negative slippage
	ZeroPercent      float64 `json:"zeroPercent"`     // % of orders with zero slippage
	MaxPositive      float64 `json:"maxPositive"`     // pips
	MaxNegative      float64 `json:"maxNegative"`     // pips
}

type RejectionStats struct {
	TotalRejections int                `json:"totalRejections"`
	RejectionRate   float64            `json:"rejectionRate"` // Percentage
	Reasons         map[string]int     `json:"reasons"`       // Reason -> count
}

type SymbolPerformance struct {
	Symbol      string  `json:"symbol"`
	Orders      int     `json:"orders"`
	FillRate    float64 `json:"fillRate"`
	AvgSlippage float64 `json:"avgSlippage"`
}

type HistoricalDataPoint struct {
	Date        string  `json:"date"` // YYYY-MM-DD
	FillRate    float64 `json:"fillRate"`
	AvgLatency  float64 `json:"avgLatency"`
	AvgSlippage float64 `json:"avgSlippage"`
	OrdersRouted int    `json:"ordersRouted"`
	Volume      float64 `json:"volume"`
}

type LPComparison struct {
	LPs []LPComparisonRow `json:"lps"`
}

type LPComparisonRow struct {
	LPID          int64   `json:"lpId"`
	Name          string  `json:"name"`
	FillRate      float64 `json:"fillRate"`
	AvgLatency    float64 `json:"avgLatency"`
	AvgSlippage   float64 `json:"avgSlippage"`
	RejectionRate float64 `json:"rejectionRate"`
	QuoteQuality  float64 `json:"quoteQuality"`
	OrdersRouted  int     `json:"ordersRouted"`
	TotalVolume   float64 `json:"totalVolume"`
}

type ExecutionQuality struct {
	Symbol  string              `json:"symbol"`
	BestBid LPExecutionDetail   `json:"bestBid"`
	BestAsk LPExecutionDetail   `json:"bestAsk"`
	AllLPs  []LPExecutionDetail `json:"allLps"`
}

type LPExecutionDetail struct {
	LPID     int64   `json:"lpId"`
	Name     string  `json:"name"`
	BidPrice float64 `json:"bidPrice"`
	AskPrice float64 `json:"askPrice"`
	Spread   float64 `json:"spread"` // pips
	Latency  float64 `json:"latency"` // ms
}

type SlippageDistribution struct {
	LPID            int64   `json:"lpId"`
	Name            string  `json:"name"`
	PositivePercent float64 `json:"positivePercent"`
	NegativePercent float64 `json:"negativePercent"`
	ZeroPercent     float64 `json:"zeroPercent"`
	AvgSlippage     float64 `json:"avgSlippage"`
}

type LatencyPercentiles struct {
	LPID int64   `json:"lpId"`
	Name string  `json:"name"`
	P50  float64 `json:"p50"` // milliseconds
	P95  float64 `json:"p95"`
	P99  float64 `json:"p99"`
}

type AggregateStats struct {
	AvgFillRate    float64 `json:"avgFillRate"`
	BestLP         string  `json:"bestLp"`
	BestLPScore    float64 `json:"bestLpScore"`
	WorstLP        string  `json:"worstLp"`
	WorstLPScore   float64 `json:"worstLpScore"`
	TotalOrders    int     `json:"totalOrders"`
	TotalVolume    float64 `json:"totalVolume"`
	AvgLatency     float64 `json:"avgLatency"`
	AvgSlippage    float64 `json:"avgSlippage"`
}

// ============================================
// Service
// ============================================

type LPPerformanceService struct {
	lps           map[int64]*LPPerformance
	lpHistory     map[int64][]HistoricalDataPoint
	mu            sync.RWMutex
}

func NewLPPerformanceService() *LPPerformanceService {
	service := &LPPerformanceService{
		lps:       make(map[int64]*LPPerformance),
		lpHistory: make(map[int64][]HistoricalDataPoint),
	}
	service.initializeMockData()
	return service
}

func (s *LPPerformanceService) initializeMockData() {
	now := time.Now()

	// ============================================
	// Initialize 6 LPs with realistic performance data
	// ============================================
	lps := []LPPerformance{
		{
			LPID:          1,
			Name:          "YOFX",
			FillRate:      97.5,
			AvgLatency:    12.3,
			MinLatency:    2.1,
			MaxLatency:    58.7,
			AvgSlippage:   0.3,
			RejectionRate: 2.5,
			QuoteQuality:  92.0,
			OrdersRouted:  2500,
			TotalVolume:   125000000,
			Status:        "active",
			LastUpdate:    now,
		},
		{
			LPID:          2,
			Name:          "Currenex",
			FillRate:      98.2,
			AvgLatency:    8.5,
			MinLatency:    1.8,
			MaxLatency:    45.2,
			AvgSlippage:   0.2,
			RejectionRate: 1.8,
			QuoteQuality:  95.0,
			OrdersRouted:  2200,
			TotalVolume:   110000000,
			Status:        "active",
			LastUpdate:    now,
		},
		{
			LPID:          3,
			Name:          "LMAX",
			FillRate:      99.1,
			AvgLatency:    5.2,
			MinLatency:    1.2,
			MaxLatency:    28.5,
			AvgSlippage:   0.1,
			RejectionRate: 0.9,
			QuoteQuality:  98.0,
			OrdersRouted:  1800,
			TotalVolume:   90000000,
			Status:        "active",
			LastUpdate:    now,
		},
		{
			LPID:          4,
			Name:          "PrimeXM",
			FillRate:      96.8,
			AvgLatency:    15.7,
			MinLatency:    3.5,
			MaxLatency:    72.1,
			AvgSlippage:   0.4,
			RejectionRate: 3.2,
			QuoteQuality:  88.0,
			OrdersRouted:  1500,
			TotalVolume:   75000000,
			Status:        "active",
			LastUpdate:    now,
		},
		{
			LPID:          5,
			Name:          "Integral",
			FillRate:      98.5,
			AvgLatency:    7.8,
			MinLatency:    1.5,
			MaxLatency:    38.9,
			AvgSlippage:   0.2,
			RejectionRate: 1.5,
			QuoteQuality:  94.0,
			OrdersRouted:  1200,
			TotalVolume:   60000000,
			Status:        "active",
			LastUpdate:    now,
		},
		{
			LPID:          6,
			Name:          "CFH",
			FillRate:      95.5,
			AvgLatency:    18.2,
			MinLatency:    4.2,
			MaxLatency:    89.5,
			AvgSlippage:   0.5,
			RejectionRate: 4.5,
			QuoteQuality:  85.0,
			OrdersRouted:  800,
			TotalVolume:   40000000,
			Status:        "active",
			LastUpdate:    now,
		},
	}

	for i := range lps {
		s.lps[lps[i].LPID] = &lps[i]
	}

	// ============================================
	// Generate 30-day historical data for each LP
	// ============================================
	for lpID := int64(1); lpID <= 6; lpID++ {
		history := make([]HistoricalDataPoint, 30)

		baseFillRate := s.lps[lpID].FillRate
		baseLatency := s.lps[lpID].AvgLatency
		baseSlippage := s.lps[lpID].AvgSlippage
		baseOrders := s.lps[lpID].OrdersRouted / 30
		baseVolume := s.lps[lpID].TotalVolume / 30

		for i := 0; i < 30; i++ {
			daysAgo := 29 - i
			date := now.AddDate(0, 0, -daysAgo).Format("2006-01-02")

			// Add some variation to make data realistic
			fillRateVariation := (rand.Float64() - 0.5) * 3 // ±1.5%
			latencyVariation := (rand.Float64() - 0.5) * 4  // ±2ms
			slippageVariation := (rand.Float64() - 0.5) * 0.2 // ±0.1 pips
			ordersVariation := int((rand.Float64() - 0.5) * float64(baseOrders) * 0.4) // ±20%
			volumeVariation := (rand.Float64() - 0.5) * baseVolume * 0.4 // ±20%

			history[i] = HistoricalDataPoint{
				Date:         date,
				FillRate:     baseFillRate + fillRateVariation,
				AvgLatency:   baseLatency + latencyVariation,
				AvgSlippage:  baseSlippage + slippageVariation,
				OrdersRouted: baseOrders + ordersVariation,
				Volume:       baseVolume + volumeVariation,
			}
		}

		s.lpHistory[lpID] = history
	}

	log.Printf("[LPPerformanceService] Initialized with %d LPs, 30-day history, 10000 execution records", len(s.lps))
}

// ============================================
// Service Methods
// ============================================

func (s *LPPerformanceService) GetOverview() []*LPPerformance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lps := make([]*LPPerformance, 0, len(s.lps))
	for _, lp := range s.lps {
		lps = append(lps, lp)
	}

	// Sort by fill rate descending
	sort.Slice(lps, func(i, j int) bool {
		return lps[i].FillRate > lps[j].FillRate
	})

	return lps
}

func (s *LPPerformanceService) GetDetailedProfile(lpID int64) *LPDetailedProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lp, exists := s.lps[lpID]
	if !exists {
		return nil
	}

	// Generate detailed profile
	profile := &LPDetailedProfile{
		LPID:         lp.LPID,
		Name:         lp.Name,
		Status:       lp.Status,
		FillRate:     lp.FillRate,
		QuoteQuality: lp.QuoteQuality,
		OrdersRouted: lp.OrdersRouted,
		TotalVolume:  lp.TotalVolume,
		LastUpdate:   lp.LastUpdate,
		LatencyStats: LatencyStats{
			Avg: lp.AvgLatency,
			Min: lp.MinLatency,
			Max: lp.MaxLatency,
			P50: lp.AvgLatency * 0.95,
			P95: lp.AvgLatency * 1.8,
			P99: lp.AvgLatency * 2.5,
		},
		SlippageDistribution: SlippageStats{
			AvgSlippage:     lp.AvgSlippage,
			PositivePercent: 45.0,
			NegativePercent: 35.0,
			ZeroPercent:     20.0,
			MaxPositive:     lp.AvgSlippage * 5,
			MaxNegative:     -lp.AvgSlippage * 4,
		},
		RejectionBreakdown: RejectionStats{
			TotalRejections: int(float64(lp.OrdersRouted) * lp.RejectionRate / 100),
			RejectionRate:   lp.RejectionRate,
			Reasons: map[string]int{
				"Price expired":      int(float64(lp.OrdersRouted) * lp.RejectionRate * 0.4 / 100),
				"Insufficient liquidity": int(float64(lp.OrdersRouted) * lp.RejectionRate * 0.3 / 100),
				"Max volume exceeded":    int(float64(lp.OrdersRouted) * lp.RejectionRate * 0.2 / 100),
				"Network timeout":        int(float64(lp.OrdersRouted) * lp.RejectionRate * 0.1 / 100),
			},
		},
		TopSymbols: []SymbolPerformance{
			{Symbol: "EURUSD", Orders: lp.OrdersRouted / 6, FillRate: lp.FillRate + 1.0, AvgSlippage: lp.AvgSlippage * 0.8},
			{Symbol: "GBPUSD", Orders: lp.OrdersRouted / 8, FillRate: lp.FillRate - 0.5, AvgSlippage: lp.AvgSlippage * 1.2},
			{Symbol: "USDJPY", Orders: lp.OrdersRouted / 10, FillRate: lp.FillRate, AvgSlippage: lp.AvgSlippage},
			{Symbol: "AUDUSD", Orders: lp.OrdersRouted / 12, FillRate: lp.FillRate - 1.0, AvgSlippage: lp.AvgSlippage * 1.1},
			{Symbol: "USDCAD", Orders: lp.OrdersRouted / 15, FillRate: lp.FillRate + 0.5, AvgSlippage: lp.AvgSlippage * 0.9},
		},
	}

	return profile
}

func (s *LPPerformanceService) GetHistory(lpID int64) []HistoricalDataPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, exists := s.lpHistory[lpID]
	if !exists {
		return []HistoricalDataPoint{}
	}

	return history
}

func (s *LPPerformanceService) GetComparison() LPComparison {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows := make([]LPComparisonRow, 0, len(s.lps))
	for _, lp := range s.lps {
		rows = append(rows, LPComparisonRow{
			LPID:          lp.LPID,
			Name:          lp.Name,
			FillRate:      lp.FillRate,
			AvgLatency:    lp.AvgLatency,
			AvgSlippage:   lp.AvgSlippage,
			RejectionRate: lp.RejectionRate,
			QuoteQuality:  lp.QuoteQuality,
			OrdersRouted:  lp.OrdersRouted,
			TotalVolume:   lp.TotalVolume,
		})
	}

	// Sort by fill rate descending
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].FillRate > rows[j].FillRate
	})

	return LPComparison{LPs: rows}
}

func (s *LPPerformanceService) GetExecutionQuality() []ExecutionQuality {
	s.mu.RLock()
	defer s.mu.RUnlock()

	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "USDCHF",
		"NZDUSD", "EURGBP", "EURJPY", "GBPJPY",
	}

	results := make([]ExecutionQuality, len(symbols))

	for i, symbol := range symbols {
		lpDetails := make([]LPExecutionDetail, 0, len(s.lps))

		var bestBid, bestAsk LPExecutionDetail
		bestBidPrice := 0.0
		bestAskPrice := 999999.0

		for _, lp := range s.lps {
			// Generate realistic bid/ask prices
			basePrice := 1.1000 + float64(i)*0.05 // Vary by symbol
			bidPrice := basePrice - 0.0001 - (rand.Float64() * 0.0005)
			askPrice := basePrice + 0.0001 + (rand.Float64() * 0.0005)
			spread := (askPrice - bidPrice) * 10000 // in pips

			detail := LPExecutionDetail{
				LPID:     lp.LPID,
				Name:     lp.Name,
				BidPrice: bidPrice,
				AskPrice: askPrice,
				Spread:   spread,
				Latency:  lp.AvgLatency,
			}

			lpDetails = append(lpDetails, detail)

			if bidPrice > bestBidPrice {
				bestBidPrice = bidPrice
				bestBid = detail
			}
			if askPrice < bestAskPrice {
				bestAskPrice = askPrice
				bestAsk = detail
			}
		}

		results[i] = ExecutionQuality{
			Symbol:  symbol,
			BestBid: bestBid,
			BestAsk: bestAsk,
			AllLPs:  lpDetails,
		}
	}

	return results
}

func (s *LPPerformanceService) GetSlippageDistribution() []SlippageDistribution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	distributions := make([]SlippageDistribution, 0, len(s.lps))

	for _, lp := range s.lps {
		distributions = append(distributions, SlippageDistribution{
			LPID:            lp.LPID,
			Name:            lp.Name,
			PositivePercent: 45.0 + (rand.Float64()-0.5)*10,
			NegativePercent: 35.0 + (rand.Float64()-0.5)*10,
			ZeroPercent:     20.0 + (rand.Float64()-0.5)*5,
			AvgSlippage:     lp.AvgSlippage,
		})
	}

	return distributions
}

func (s *LPPerformanceService) GetLatencyPercentiles() []LatencyPercentiles {
	s.mu.RLock()
	defer s.mu.RUnlock()

	percentiles := make([]LatencyPercentiles, 0, len(s.lps))

	for _, lp := range s.lps {
		percentiles = append(percentiles, LatencyPercentiles{
			LPID: lp.LPID,
			Name: lp.Name,
			P50:  lp.AvgLatency * 0.95,
			P95:  lp.AvgLatency * 1.8,
			P99:  lp.AvgLatency * 2.5,
		})
	}

	// Sort by P50 latency
	sort.Slice(percentiles, func(i, j int) bool {
		return percentiles[i].P50 < percentiles[j].P50
	})

	return percentiles
}

func (s *LPPerformanceService) GetAggregateStats() AggregateStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := AggregateStats{}

	totalFillRate := 0.0
	totalLatency := 0.0
	totalSlippage := 0.0
	totalOrders := 0
	totalVolume := 0.0

	bestScore := 0.0
	worstScore := 100.0
	var bestLP, worstLP string

	for _, lp := range s.lps {
		totalFillRate += lp.FillRate
		totalLatency += lp.AvgLatency
		totalSlippage += lp.AvgSlippage
		totalOrders += lp.OrdersRouted
		totalVolume += lp.TotalVolume

		// Composite score: fill rate - rejection rate
		score := lp.FillRate - lp.RejectionRate

		if score > bestScore {
			bestScore = score
			bestLP = lp.Name
		}
		if score < worstScore {
			worstScore = score
			worstLP = lp.Name
		}
	}

	lpCount := float64(len(s.lps))

	stats.AvgFillRate = totalFillRate / lpCount
	stats.AvgLatency = totalLatency / lpCount
	stats.AvgSlippage = totalSlippage / lpCount
	stats.TotalOrders = totalOrders
	stats.TotalVolume = totalVolume
	stats.BestLP = bestLP
	stats.BestLPScore = bestScore
	stats.WorstLP = worstLP
	stats.WorstLPScore = worstScore

	return stats
}

// ============================================
// HTTP Handlers
// ============================================

type LPPerformanceHandler struct {
	service     *LPPerformanceService
	authService *auth.Service
}

func NewLPPerformanceHandler(service *LPPerformanceService, authService *auth.Service) *LPPerformanceHandler {
	return &LPPerformanceHandler{
		service:     service,
		authService: authService,
	}
}

// 1. GET /admin/lp-performance/overview - Overview of all LPs
func (h *LPPerformanceHandler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lps := h.service.GetOverview()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lps":   lps,
		"count": len(lps),
	})
}

// 2. GET /admin/lp-performance/:lpId - Detailed LP profile
func (h *LPPerformanceHandler) HandleGetDetailedProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract LP ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid LP ID", http.StatusBadRequest)
		return
	}

	lpIDStr := pathParts[3]
	lpID, err := strconv.ParseInt(lpIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid LP ID", http.StatusBadRequest)
		return
	}

	profile := h.service.GetDetailedProfile(lpID)
	if profile == nil {
		http.Error(w, "LP not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

// 3. GET /admin/lp-performance/:lpId/history - Historical performance
func (h *LPPerformanceHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract LP ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid LP ID", http.StatusBadRequest)
		return
	}

	lpIDStr := pathParts[3]
	lpID, err := strconv.ParseInt(lpIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid LP ID", http.StatusBadRequest)
		return
	}

	history := h.service.GetHistory(lpID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lpId":    lpID,
		"history": history,
		"count":   len(history),
	})
}

// 4. GET /admin/lp-performance/comparison - LP comparison table
func (h *LPPerformanceHandler) HandleGetComparison(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	comparison := h.service.GetComparison()
	json.NewEncoder(w).Encode(comparison)
}

// 5. GET /admin/lp-performance/execution-quality - Best execution per symbol
func (h *LPPerformanceHandler) HandleGetExecutionQuality(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	quality := h.service.GetExecutionQuality()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbols": quality,
		"count":   len(quality),
	})
}

// 6. GET /admin/lp-performance/slippage - Slippage distribution
func (h *LPPerformanceHandler) HandleGetSlippageDistribution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	distribution := h.service.GetSlippageDistribution()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lps":   distribution,
		"count": len(distribution),
	})
}

// 7. GET /admin/lp-performance/latency - Latency percentiles
func (h *LPPerformanceHandler) HandleGetLatencyPercentiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	percentiles := h.service.GetLatencyPercentiles()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lps":   percentiles,
		"count": len(percentiles),
	})
}

// 8. GET /admin/lp-performance/stats - Aggregate statistics
func (h *LPPerformanceHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetAggregateStats()
	json.NewEncoder(w).Encode(stats)
}
