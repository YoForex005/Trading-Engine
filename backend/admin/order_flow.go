package admin

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================
// Order Flow Structs
// ============================================

type OrderFlowEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"` // buy, sell
	Volume    float64   `json:"volume"`
	Price     float64   `json:"price"`
	Aggressor bool      `json:"aggressor"` // true if market taker
	LP        string    `json:"lp"`        // liquidity provider
	IsLarge   bool      `json:"is_large"`  // >10 lots
	Toxicity  float64   `json:"toxicity"`  // flow toxicity score 0-1
}

type FlowMetrics struct {
	Symbol                string    `json:"symbol"`
	BuyVolume             float64   `json:"buy_volume"`
	SellVolume            float64   `json:"sell_volume"`
	NetFlow               float64   `json:"net_flow"`
	VWAP                  float64   `json:"vwap"`
	LargeOrderCount       int       `json:"large_order_count"`
	RetailVsInstitutional float64   `json:"retail_vs_institutional"` // ratio
	ToxicityScore         float64   `json:"toxicity_score"`
	ImbalanceRatio        float64   `json:"imbalance_ratio"` // buy/sell ratio
	TotalTrades           int       `json:"total_trades"`
	LastUpdated           time.Time `json:"last_updated"`
}

type VolumeHeatmapData struct {
	Symbol      string              `json:"symbol"`
	TimeBuckets []string            `json:"time_buckets"`
	PriceLevels []float64           `json:"price_levels"`
	VolumeGrid  [][]float64         `json:"volume_grid"` // [time][price]
	Generated   time.Time           `json:"generated"`
}

type LargeOrder struct {
	OrderFlowEntry
	Significance float64 `json:"significance"` // how much larger than average
}

type OrderImbalance struct {
	Symbol         string  `json:"symbol"`
	BuyVolume      float64 `json:"buy_volume"`
	SellVolume     float64 `json:"sell_volume"`
	ImbalanceRatio float64 `json:"imbalance_ratio"`
	NetFlow        float64 `json:"net_flow"`
	Direction      string  `json:"direction"` // bullish, bearish, neutral
}

type TopSymbolFlow struct {
	Symbol      string  `json:"symbol"`
	TotalVolume float64 `json:"total_volume"`
	TradeCount  int     `json:"trade_count"`
	BuyRatio    float64 `json:"buy_ratio"`
	AvgPrice    float64 `json:"avg_price"`
	Rank        int     `json:"rank"`
}

// ============================================
// Order Flow Service
// ============================================

type OrderFlowService struct {
	mu           sync.RWMutex
	entries      []OrderFlowEntry
	nextID       int64
	symbols      []string
	lps          []string
	threshold    float64 // large order threshold (lots)
}

func NewOrderFlowService() *OrderFlowService {
	service := &OrderFlowService{
		entries:   make([]OrderFlowEntry, 0, 500),
		nextID:    1,
		threshold: 10.0,
		symbols: []string{
			"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
			"NZDUSD", "USDCHF", "EURGBP", "EURJPY", "GBPJPY",
			"AUDJPY", "EURAUD", "XAUUSD", "XAGUSD", "BTCUSD",
			"ETHUSD", "US30", "NAS100", "SPX500", "UK100",
		},
		lps: []string{"YOFX", "Oanda", "LMAX", "Currenex", "EBS", "Hotspot"},
	}

	service.generateMockData()
	return service
}

func (s *OrderFlowService) generateMockData() {
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	// Base prices for symbols
	basePrices := map[string]float64{
		"EURUSD": 1.0850, "GBPUSD": 1.2650, "USDJPY": 148.50, "AUDUSD": 0.6450,
		"USDCAD": 1.3550, "NZDUSD": 0.5850, "USDCHF": 0.8850, "EURGBP": 0.8550,
		"EURJPY": 161.20, "GBPJPY": 187.80, "AUDJPY": 95.80, "EURAUD": 1.6820,
		"XAUUSD": 2650.00, "XAGUSD": 30.50, "BTCUSD": 95000.00, "ETHUSD": 3200.00,
		"US30": 42500.00, "NAS100": 18500.00, "SPX500": 5800.00, "UK100": 8200.00,
	}

	// Generate 500 order flow entries
	for i := 0; i < 500; i++ {
		symbol := s.symbols[rand.Intn(len(s.symbols))]
		basePrice := basePrices[symbol]

		// Random price variation
		priceVariation := (rand.Float64()*2 - 1) * 0.001 * basePrice
		price := basePrice + priceVariation

		// Random volume (0.1 to 50 lots)
		volume := 0.1 + rand.Float64()*49.9

		// 30% chance of large order
		isLarge := volume > s.threshold

		// Random side (slightly biased toward buy in trending market)
		side := "buy"
		if rand.Float64() > 0.52 {
			side = "sell"
		}

		// 70% are aggressors (market takers)
		aggressor := rand.Float64() < 0.70

		// Random LP
		lp := s.lps[rand.Intn(len(s.lps))]

		// Calculate toxicity score (higher for large aggressive orders)
		toxicity := 0.1 + rand.Float64()*0.3
		if isLarge && aggressor {
			toxicity += 0.4
		}
		toxicity = math.Min(toxicity, 1.0)

		// Time distribution - most recent entries
		minutesAgo := int(math.Pow(rand.Float64(), 2) * 120) // skew toward recent
		timestamp := now.Add(-time.Duration(minutesAgo) * time.Minute)

		entry := OrderFlowEntry{
			ID:        s.nextID,
			Timestamp: timestamp,
			Symbol:    symbol,
			Side:      side,
			Volume:    volume,
			Price:     price,
			Aggressor: aggressor,
			LP:        lp,
			IsLarge:   isLarge,
			Toxicity:  toxicity,
		}

		s.entries = append(s.entries, entry)
		s.nextID++
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(s.entries, func(i, j int) bool {
		return s.entries[i].Timestamp.After(s.entries[j].Timestamp)
	})
}

func (s *OrderFlowService) GetLiveFlow(symbol, side string, minSize float64, limit, offset int) ([]OrderFlowEntry, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]OrderFlowEntry, 0)
	for _, entry := range s.entries {
		if symbol != "" && entry.Symbol != symbol {
			continue
		}
		if side != "" && entry.Side != side {
			continue
		}
		if minSize > 0 && entry.Volume < minSize {
			continue
		}
		filtered = append(filtered, entry)
	}

	total := len(filtered)

	if offset >= total {
		return []OrderFlowEntry{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *OrderFlowService) GetMetrics(symbol string) *FlowMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var buyVol, sellVol, totalValue float64
	var largeCount, totalTrades int
	var toxicitySum float64

	for _, entry := range s.entries {
		if entry.Symbol != symbol {
			continue
		}

		totalTrades++
		totalValue += entry.Price * entry.Volume

		if entry.Side == "buy" {
			buyVol += entry.Volume
		} else {
			sellVol += entry.Volume
		}

		if entry.IsLarge {
			largeCount++
		}

		toxicitySum += entry.Toxicity
	}

	if totalTrades == 0 {
		return nil
	}

	totalVol := buyVol + sellVol
	vwap := totalValue / totalVol
	netFlow := buyVol - sellVol
	imbalanceRatio := buyVol / sellVol
	avgToxicity := toxicitySum / float64(totalTrades)

	// Retail vs institutional (based on order size distribution)
	retailRatio := 0.7 // 70% retail, 30% institutional (mock)

	return &FlowMetrics{
		Symbol:                symbol,
		BuyVolume:             buyVol,
		SellVolume:            sellVol,
		NetFlow:               netFlow,
		VWAP:                  vwap,
		LargeOrderCount:       largeCount,
		RetailVsInstitutional: retailRatio,
		ToxicityScore:         avgToxicity,
		ImbalanceRatio:        imbalanceRatio,
		TotalTrades:           totalTrades,
		LastUpdated:           time.Now(),
	}
}

func (s *OrderFlowService) GetHeatmap(symbol string) *VolumeHeatmapData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if symbol == "" {
		symbol = "EURUSD"
	}

	// Generate time buckets (last 2 hours, 5-minute buckets)
	now := time.Now()
	buckets := make([]string, 24)
	for i := 0; i < 24; i++ {
		t := now.Add(-time.Duration(23-i) * 5 * time.Minute)
		buckets[i] = t.Format("15:04")
	}

	// Get price range for symbol
	var minPrice, maxPrice float64 = math.MaxFloat64, 0
	for _, entry := range s.entries {
		if entry.Symbol == symbol {
			if entry.Price < minPrice {
				minPrice = entry.Price
			}
			if entry.Price > maxPrice {
				maxPrice = entry.Price
			}
		}
	}

	// Generate 20 price levels
	priceLevels := make([]float64, 20)
	priceStep := (maxPrice - minPrice) / 19
	for i := 0; i < 20; i++ {
		priceLevels[i] = minPrice + float64(i)*priceStep
	}

	// Generate volume grid
	volumeGrid := make([][]float64, 24)
	for i := range volumeGrid {
		volumeGrid[i] = make([]float64, 20)
		for j := range volumeGrid[i] {
			// Simulate volume distribution
			volumeGrid[i][j] = rand.Float64() * 100
		}
	}

	return &VolumeHeatmapData{
		Symbol:      symbol,
		TimeBuckets: buckets,
		PriceLevels: priceLevels,
		VolumeGrid:  volumeGrid,
		Generated:   now,
	}
}

func (s *OrderFlowService) GetLargeOrders(threshold float64, limit int) []LargeOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if threshold == 0 {
		threshold = s.threshold
	}

	// Calculate average volume for significance
	var totalVol float64
	var count int
	for _, entry := range s.entries {
		totalVol += entry.Volume
		count++
	}
	avgVol := totalVol / float64(count)

	largeOrders := make([]LargeOrder, 0)
	for _, entry := range s.entries {
		if entry.Volume >= threshold {
			significance := entry.Volume / avgVol
			largeOrders = append(largeOrders, LargeOrder{
				OrderFlowEntry: entry,
				Significance:   significance,
			})
		}
	}

	// Sort by volume descending
	sort.Slice(largeOrders, func(i, j int) bool {
		return largeOrders[i].Volume > largeOrders[j].Volume
	})

	if limit > 0 && limit < len(largeOrders) {
		largeOrders = largeOrders[:limit]
	}

	return largeOrders
}

func (s *OrderFlowService) GetImbalance() []OrderImbalance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate imbalance per symbol
	imbalanceMap := make(map[string]*OrderImbalance)

	for _, entry := range s.entries {
		if imbalanceMap[entry.Symbol] == nil {
			imbalanceMap[entry.Symbol] = &OrderImbalance{
				Symbol: entry.Symbol,
			}
		}

		if entry.Side == "buy" {
			imbalanceMap[entry.Symbol].BuyVolume += entry.Volume
		} else {
			imbalanceMap[entry.Symbol].SellVolume += entry.Volume
		}
	}

	// Calculate ratios and directions
	result := make([]OrderImbalance, 0, len(imbalanceMap))
	for _, imb := range imbalanceMap {
		imb.NetFlow = imb.BuyVolume - imb.SellVolume

		if imb.SellVolume > 0 {
			imb.ImbalanceRatio = imb.BuyVolume / imb.SellVolume
		} else {
			imb.ImbalanceRatio = 999.0
		}

		// Determine direction
		if imb.ImbalanceRatio > 1.2 {
			imb.Direction = "bullish"
		} else if imb.ImbalanceRatio < 0.8 {
			imb.Direction = "bearish"
		} else {
			imb.Direction = "neutral"
		}

		result = append(result, *imb)
	}

	// Sort by absolute imbalance
	sort.Slice(result, func(i, j int) bool {
		return math.Abs(result[i].NetFlow) > math.Abs(result[j].NetFlow)
	})

	return result
}

func (s *OrderFlowService) GetTopSymbols(limit int) []TopSymbolFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Aggregate by symbol
	symbolMap := make(map[string]*TopSymbolFlow)

	for _, entry := range s.entries {
		if symbolMap[entry.Symbol] == nil {
			symbolMap[entry.Symbol] = &TopSymbolFlow{
				Symbol: entry.Symbol,
			}
		}

		flow := symbolMap[entry.Symbol]
		flow.TotalVolume += entry.Volume
		flow.TradeCount++
		flow.AvgPrice += entry.Price

		if entry.Side == "buy" {
			// Count buys for ratio calculation
		}
	}

	// Calculate averages and ratios
	result := make([]TopSymbolFlow, 0, len(symbolMap))
	for _, flow := range symbolMap {
		flow.AvgPrice /= float64(flow.TradeCount)

		// Calculate buy ratio
		var buyCount int
		for _, entry := range s.entries {
			if entry.Symbol == flow.Symbol && entry.Side == "buy" {
				buyCount++
			}
		}
		flow.BuyRatio = float64(buyCount) / float64(flow.TradeCount)

		result = append(result, *flow)
	}

	// Sort by total volume descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalVolume > result[j].TotalVolume
	})

	// Assign ranks
	for i := range result {
		result[i].Rank = i + 1
	}

	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result
}

// ============================================
// Order Flow Handler
// ============================================

type OrderFlowHandler struct {
	service     *OrderFlowService
	authService *AuthService
}

func NewOrderFlowHandler(service *OrderFlowService, authService *AuthService) *OrderFlowHandler {
	return &OrderFlowHandler{
		service:     service,
		authService: authService,
	}
}

func (h *OrderFlowHandler) HandleGetLiveFlow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	side := r.URL.Query().Get("side")
	minSizeStr := r.URL.Query().Get("min_size")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	minSize := 0.0
	if minSizeStr != "" {
		if parsed, err := strconv.ParseFloat(minSizeStr, 64); err == nil {
			minSize = parsed
		}
	}

	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsed
		}
	}

	entries, total := h.service.GetLiveFlow(symbol, side, minSize, limit, offset)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *OrderFlowHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Extract symbol from URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/order-flow/metrics/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Symbol not specified", http.StatusBadRequest)
		return
	}
	symbol := parts[0]

	metrics := h.service.GetMetrics(symbol)
	if metrics == nil {
		http.Error(w, "No data for symbol", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(metrics)
}

func (h *OrderFlowHandler) HandleGetHeatmap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	heatmap := h.service.GetHeatmap(symbol)

	json.NewEncoder(w).Encode(heatmap)
}

func (h *OrderFlowHandler) HandleGetLargeOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	thresholdStr := r.URL.Query().Get("threshold")
	limitStr := r.URL.Query().Get("limit")

	threshold := 0.0
	if thresholdStr != "" {
		if parsed, err := strconv.ParseFloat(thresholdStr, 64); err == nil {
			threshold = parsed
		}
	}

	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	largeOrders := h.service.GetLargeOrders(threshold, limit)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"large_orders": largeOrders,
		"total":        len(largeOrders),
		"threshold":    threshold,
	})
}

func (h *OrderFlowHandler) HandleGetImbalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	imbalances := h.service.GetImbalance()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"imbalances": imbalances,
		"total":      len(imbalances),
	})
}

func (h *OrderFlowHandler) HandleGetTopSymbols(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	topSymbols := h.service.GetTopSymbols(limit)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"top_symbols": topSymbols,
		"total":       len(topSymbols),
	})
}
