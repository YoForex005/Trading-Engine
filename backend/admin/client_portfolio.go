package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ClientPortfolio represents a client's trading portfolio summary
type ClientPortfolio struct {
	ClientID             int64     `json:"client_id"`
	ClientName           string    `json:"client_name"`
	Group                string    `json:"group"`
	TotalEquity          float64   `json:"total_equity"`
	TotalBalance         float64   `json:"total_balance"`
	UnrealizedPnL        float64   `json:"unrealized_pnl"`
	RealizedPnL          float64   `json:"realized_pnl"`
	MarginUsed           float64   `json:"margin_used"`
	FreeMargin           float64   `json:"free_margin"`
	MarginLevel          float64   `json:"margin_level"`
	OpenPositions        int       `json:"open_positions"`
	TotalTrades          int       `json:"total_trades"`
	WinRate              float64   `json:"win_rate"`
	AvgHoldingTimeHours  float64   `json:"avg_holding_time_hours"`
	PreferredSymbols     []string  `json:"preferred_symbols"`
	TradingStyle         string    `json:"trading_style"` // scalper, day_trader, swing, position
	RiskScore            float64   `json:"risk_score"`
	LastTradeAt          time.Time `json:"last_trade_at"`
}

// TradingPattern represents detected trading behavior patterns
type TradingPattern struct {
	ID             int64     `json:"id"`
	ClientID       int64     `json:"client_id"`
	ClientName     string    `json:"client_name"`
	Pattern        string    `json:"pattern"` // overtrading, revenge_trading, martingale, consistent_winner, high_frequency, news_trader, trend_follower
	Confidence     float64   `json:"confidence"`
	DetectedAt     time.Time `json:"detected_at"`
	Evidence       string    `json:"evidence"`
	Recommendation string    `json:"recommendation"`
	Status         string    `json:"status"` // active, resolved, monitoring
}

// PortfolioSnapshot represents daily portfolio snapshots for equity curve
type PortfolioSnapshot struct {
	ClientID           int64     `json:"client_id"`
	Date               time.Time `json:"date"`
	Equity             float64   `json:"equity"`
	Balance            float64   `json:"balance"`
	OpenPnL            float64   `json:"open_pnl"`
	ClosedPnL          float64   `json:"closed_pnl"`
	DepositsToday      float64   `json:"deposits_today"`
	WithdrawalsToday   float64   `json:"withdrawals_today"`
}

// ClientComparison represents client metrics for comparison/leaderboard
type ClientComparison struct {
	ClientID     int64   `json:"client_id"`
	ClientName   string  `json:"client_name"`
	Metric       string  `json:"metric"`
	Value        float64 `json:"value"`
	Rank         int     `json:"rank"`
	Percentile   float64 `json:"percentile"`
}

// PortfolioStats represents overall portfolio analytics
type PortfolioStats struct {
	TotalClients       int                    `json:"total_clients"`
	TotalAUM           float64                `json:"total_aum"` // Assets Under Management
	AvgEquity          float64                `json:"avg_equity"`
	TotalOpenPositions int                    `json:"total_open_positions"`
	TotalTrades        int                    `json:"total_trades"`
	AvgWinRate         float64                `json:"avg_win_rate"`
	StyleDistribution  map[string]int         `json:"style_distribution"`
	PatternDistribution map[string]int        `json:"pattern_distribution"`
	TopPerformers      []ClientPortfolio      `json:"top_performers"`
	RiskDistribution   map[string]int         `json:"risk_distribution"`
}

// ClientPortfolioService manages client portfolio analysis
type ClientPortfolioService struct {
	mu            sync.RWMutex
	portfolios    map[int64]*ClientPortfolio
	patterns      map[int64]*TradingPattern
	snapshots     map[int64][]PortfolioSnapshot // clientId -> snapshots
	patternID     int64
}

// NewClientPortfolioService creates a new client portfolio service with mock data
func NewClientPortfolioService() *ClientPortfolioService {
	s := &ClientPortfolioService{
		portfolios: make(map[int64]*ClientPortfolio),
		patterns:   make(map[int64]*TradingPattern),
		snapshots:  make(map[int64][]PortfolioSnapshot),
		patternID:  1,
	}
	s.initializeMockData()
	return s
}

func (s *ClientPortfolioService) initializeMockData() {
	rand.Seed(time.Now().UnixNano())

	groups := []string{"retail", "vip", "professional", "institutional", "demo"}
	styles := []string{"scalper", "day_trader", "swing", "position"}
	patterns := []string{"overtrading", "revenge_trading", "martingale", "consistent_winner", "high_frequency", "news_trader", "trend_follower"}
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "NZDUSD", "XAUUSD", "BTCUSD", "ETHUSD", "SPX500"}

	// Create 200 client portfolios
	for i := int64(1); i <= 200; i++ {
		balance := 1000.0 + rand.Float64()*99000.0
		unrealizedPnL := (rand.Float64() - 0.5) * balance * 0.1
		realizedPnL := (rand.Float64() - 0.3) * balance * 0.2
		equity := balance + unrealizedPnL
		marginUsed := equity * (0.1 + rand.Float64()*0.4)
		freeMargin := equity - marginUsed
		marginLevel := 0.0
		if marginUsed > 0 {
			marginLevel = (equity / marginUsed) * 100.0
		}

		openPos := rand.Intn(10)
		totalTrades := 50 + rand.Intn(450)
		winRate := 0.35 + rand.Float64()*0.35
		avgHolding := 0.5 + rand.Float64()*48.0
		riskScore := 1.0 + rand.Float64()*9.0

		numSymbols := 1 + rand.Intn(4)
		preferredSymbols := make([]string, numSymbols)
		for j := 0; j < numSymbols; j++ {
			preferredSymbols[j] = symbols[rand.Intn(len(symbols))]
		}

		s.portfolios[i] = &ClientPortfolio{
			ClientID:            i,
			ClientName:          fmt.Sprintf("Client-%d", i),
			Group:               groups[rand.Intn(len(groups))],
			TotalEquity:         equity,
			TotalBalance:        balance,
			UnrealizedPnL:       unrealizedPnL,
			RealizedPnL:         realizedPnL,
			MarginUsed:          marginUsed,
			FreeMargin:          freeMargin,
			MarginLevel:         marginLevel,
			OpenPositions:       openPos,
			TotalTrades:         totalTrades,
			WinRate:             winRate,
			AvgHoldingTimeHours: avgHolding,
			PreferredSymbols:    preferredSymbols,
			TradingStyle:        styles[rand.Intn(len(styles))],
			RiskScore:           riskScore,
			LastTradeAt:         time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour),
		}
	}

	// Create 150 detected patterns across clients
	for i := 0; i < 150; i++ {
		clientID := int64(1 + rand.Intn(200))
		client := s.portfolios[clientID]
		pattern := patterns[rand.Intn(len(patterns))]
		confidence := 0.6 + rand.Float64()*0.39

		evidence := ""
		recommendation := ""
		status := "active"

		switch pattern {
		case "overtrading":
			evidence = fmt.Sprintf("Client executed %d trades in 24h, exceeding normal volume by 300%%", 50+rand.Intn(100))
			recommendation = "Consider implementing trade frequency limits or cooling-off periods"
		case "revenge_trading":
			evidence = fmt.Sprintf("Client placed %d trades within 2 hours after major loss, with increasing lot sizes", 5+rand.Intn(10))
			recommendation = "Monitor closely for emotional trading. Consider intervention if pattern persists"
		case "martingale":
			evidence = fmt.Sprintf("Detected doubling of position size after %d consecutive losses", 3+rand.Intn(3))
			recommendation = "High-risk strategy detected. Consider client education or risk limits"
		case "consistent_winner":
			evidence = fmt.Sprintf("Win rate %.1f%% over %d trades with stable equity curve", client.WinRate*100, client.TotalTrades)
			recommendation = "Positive pattern. Consider for VIP tier or copy trading program"
			status = "monitoring"
		case "high_frequency":
			evidence = fmt.Sprintf("Average holding time %.1f minutes with %d trades per day", client.AvgHoldingTimeHours*60, 20+rand.Intn(80))
			recommendation = "Ensure adequate infrastructure and execution quality for HFT client"
		case "news_trader":
			evidence = fmt.Sprintf("%.1f%% of trades executed within 5 minutes of major news events", 60.0+rand.Float64()*35.0)
			recommendation = "Monitor for requotes and slippage during volatile periods"
		case "trend_follower":
			evidence = fmt.Sprintf("Majority of trades align with D1 trend direction, avg hold time %.1fh", client.AvgHoldingTimeHours)
			recommendation = "Stable strategy. Good candidate for long-term retention"
			status = "monitoring"
		}

		if rand.Float64() < 0.2 {
			status = "resolved"
		}

		s.patterns[s.patternID] = &TradingPattern{
			ID:             s.patternID,
			ClientID:       clientID,
			ClientName:     client.ClientName,
			Pattern:        pattern,
			Confidence:     confidence,
			DetectedAt:     time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
			Evidence:       evidence,
			Recommendation: recommendation,
			Status:         status,
		}
		s.patternID++
	}

	// Create 30-day equity snapshots for top 50 clients
	for clientID := int64(1); clientID <= 50; clientID++ {
		client := s.portfolios[clientID]
		snapshots := make([]PortfolioSnapshot, 30)

		currentEquity := client.TotalEquity
		for day := 0; day < 30; day++ {
			date := time.Now().AddDate(0, 0, -30+day)
			dailyChange := (rand.Float64() - 0.48) * currentEquity * 0.05
			currentEquity += dailyChange

			openPnL := (rand.Float64() - 0.5) * currentEquity * 0.03
			closedPnL := dailyChange - openPnL

			deposits := 0.0
			withdrawals := 0.0
			if rand.Float64() < 0.1 {
				deposits = rand.Float64() * 5000.0
			}
			if rand.Float64() < 0.05 {
				withdrawals = rand.Float64() * 2000.0
			}

			snapshots[day] = PortfolioSnapshot{
				ClientID:         clientID,
				Date:             date,
				Equity:           currentEquity,
				Balance:          currentEquity - openPnL,
				OpenPnL:          openPnL,
				ClosedPnL:        closedPnL,
				DepositsToday:    deposits,
				WithdrawalsToday: withdrawals,
			}
		}

		s.snapshots[clientID] = snapshots
	}
}

// GetAllPortfolios returns all client portfolios with filtering
func (s *ClientPortfolioService) GetAllPortfolios(group, style string, minEquity float64) []*ClientPortfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ClientPortfolio, 0)
	for _, p := range s.portfolios {
		if group != "" && p.Group != group {
			continue
		}
		if style != "" && p.TradingStyle != style {
			continue
		}
		if minEquity > 0 && p.TotalEquity < minEquity {
			continue
		}
		result = append(result, p)
	}

	return result
}

// GetPortfolioByID returns a specific client portfolio
func (s *ClientPortfolioService) GetPortfolioByID(clientID int64) *ClientPortfolio {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.portfolios[clientID]
}

// GetPortfolioHistory returns historical snapshots for a client
func (s *ClientPortfolioService) GetPortfolioHistory(clientID int64) []PortfolioSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshots[clientID]
}

// GetClientPatterns returns patterns detected for a specific client
func (s *ClientPortfolioService) GetClientPatterns(clientID int64) []*TradingPattern {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*TradingPattern, 0)
	for _, p := range s.patterns {
		if p.ClientID == clientID {
			result = append(result, p)
		}
	}

	return result
}

// GetAllPatterns returns all detected patterns with filtering
func (s *ClientPortfolioService) GetAllPatterns(patternType, status string) []*TradingPattern {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*TradingPattern, 0)
	for _, p := range s.patterns {
		if patternType != "" && p.Pattern != patternType {
			continue
		}
		if status != "" && p.Status != status {
			continue
		}
		result = append(result, p)
	}

	return result
}

// GetLeaderboard returns top clients by specified metric
func (s *ClientPortfolioService) GetLeaderboard(metric string, limit int) []ClientComparison {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	type rankEntry struct {
		clientID int64
		value    float64
	}

	entries := make([]rankEntry, 0, len(s.portfolios))

	for id, p := range s.portfolios {
		value := 0.0
		switch metric {
		case "pnl":
			value = p.RealizedPnL
		case "equity":
			value = p.TotalEquity
		case "volume":
			value = float64(p.TotalTrades)
		case "win_rate":
			value = p.WinRate
		case "margin_level":
			value = p.MarginLevel
		default:
			value = p.TotalEquity
		}
		entries = append(entries, rankEntry{clientID: id, value: value})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].value > entries[j].value
	})

	result := make([]ClientComparison, 0)
	totalClients := len(entries)

	for i, entry := range entries {
		if i >= limit {
			break
		}

		client := s.portfolios[entry.clientID]
		percentile := (1.0 - float64(i)/float64(totalClients)) * 100.0

		result = append(result, ClientComparison{
			ClientID:   entry.clientID,
			ClientName: client.ClientName,
			Metric:     metric,
			Value:      entry.value,
			Rank:       i + 1,
			Percentile: percentile,
		})
	}

	return result
}

// CompareClients compares specific clients on a given metric
func (s *ClientPortfolioService) CompareClients(clientIDs []int64, metric string) []ClientComparison {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type rankEntry struct {
		clientID int64
		value    float64
	}

	entries := make([]rankEntry, 0)

	for _, id := range clientIDs {
		p, exists := s.portfolios[id]
		if !exists {
			continue
		}

		value := 0.0
		switch metric {
		case "pnl":
			value = p.RealizedPnL
		case "equity":
			value = p.TotalEquity
		case "volume":
			value = float64(p.TotalTrades)
		case "win_rate":
			value = p.WinRate
		case "margin_level":
			value = p.MarginLevel
		default:
			value = p.TotalEquity
		}
		entries = append(entries, rankEntry{clientID: id, value: value})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].value > entries[j].value
	})

	result := make([]ClientComparison, 0)
	totalClients := len(entries)

	for i, entry := range entries {
		client := s.portfolios[entry.clientID]
		percentile := (1.0 - float64(i)/float64(totalClients)) * 100.0

		result = append(result, ClientComparison{
			ClientID:   entry.clientID,
			ClientName: client.ClientName,
			Metric:     metric,
			Value:      entry.value,
			Rank:       i + 1,
			Percentile: percentile,
		})
	}

	return result
}

// GetStats returns overall portfolio statistics
func (s *ClientPortfolioService) GetStats() *PortfolioStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &PortfolioStats{
		TotalClients:        len(s.portfolios),
		StyleDistribution:   make(map[string]int),
		PatternDistribution: make(map[string]int),
		RiskDistribution:    make(map[string]int),
		TopPerformers:       make([]ClientPortfolio, 0),
	}

	totalEquity := 0.0
	totalOpenPos := 0
	totalTrades := 0
	totalWinRate := 0.0

	for _, p := range s.portfolios {
		totalEquity += p.TotalEquity
		totalOpenPos += p.OpenPositions
		totalTrades += p.TotalTrades
		totalWinRate += p.WinRate

		stats.StyleDistribution[p.TradingStyle]++

		if p.RiskScore < 3.0 {
			stats.RiskDistribution["low"]++
		} else if p.RiskScore < 6.0 {
			stats.RiskDistribution["medium"]++
		} else {
			stats.RiskDistribution["high"]++
		}
	}

	stats.TotalAUM = totalEquity
	stats.AvgEquity = totalEquity / float64(len(s.portfolios))
	stats.TotalOpenPositions = totalOpenPos
	stats.TotalTrades = totalTrades
	stats.AvgWinRate = totalWinRate / float64(len(s.portfolios))

	for _, p := range s.patterns {
		if p.Status == "active" {
			stats.PatternDistribution[p.Pattern]++
		}
	}

	// Get top 5 performers by PnL
	topPerformers := s.GetLeaderboard("pnl", 5)
	for _, tp := range topPerformers {
		if p, exists := s.portfolios[tp.ClientID]; exists {
			stats.TopPerformers = append(stats.TopPerformers, *p)
		}
	}

	return stats
}

// ClientPortfolioHandler handles HTTP requests for client portfolio API
type ClientPortfolioHandler struct {
	service     *ClientPortfolioService
	authService *AuthService
}

// NewClientPortfolioHandler creates a new client portfolio handler
func NewClientPortfolioHandler(service *ClientPortfolioService, authService *AuthService) *ClientPortfolioHandler {
	return &ClientPortfolioHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetAllPortfolios returns all client portfolios with optional filters
func (h *ClientPortfolioHandler) HandleGetAllPortfolios(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	group := r.URL.Query().Get("group")
	style := r.URL.Query().Get("style")
	minEquityStr := r.URL.Query().Get("min_equity")

	minEquity := 0.0
	if minEquityStr != "" {
		if val, err := strconv.ParseFloat(minEquityStr, 64); err == nil {
			minEquity = val
		}
	}

	portfolios := h.service.GetAllPortfolios(group, style, minEquity)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(portfolios)
}

// HandleGetPortfolioByID returns a specific client portfolio
func (h *ClientPortfolioHandler) HandleGetPortfolioByID(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/admin/portfolio/clients/"):]
	if slashIdx := strings.Index(idStr, "/"); slashIdx != -1 {
		idStr = idStr[:slashIdx]
	}

	clientID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	portfolio := h.service.GetPortfolioByID(clientID)
	if portfolio == nil {
		http.Error(w, "Portfolio not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(portfolio)
}

// HandleGetPortfolioHistory returns 30-day equity curve for a client
func (h *ClientPortfolioHandler) HandleGetPortfolioHistory(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	clientID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	snapshots := h.service.GetPortfolioHistory(clientID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(snapshots)
}

// HandleGetClientPatterns returns detected patterns for a specific client
func (h *ClientPortfolioHandler) HandleGetClientPatterns(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	clientID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	patterns := h.service.GetClientPatterns(clientID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(patterns)
}

// HandleGetLeaderboard returns top clients by metric
func (h *ClientPortfolioHandler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "equity"
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	leaderboard := h.service.GetLeaderboard(metric, limit)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(leaderboard)
}

// HandleGetAllPatterns returns all detected patterns with filters
func (h *ClientPortfolioHandler) HandleGetAllPatterns(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	patternType := r.URL.Query().Get("pattern")
	status := r.URL.Query().Get("status")

	patterns := h.service.GetAllPatterns(patternType, status)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(patterns)
}

// HandleCompareClients compares specific clients on a metric
func (h *ClientPortfolioHandler) HandleCompareClients(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	clientIDsStr := r.URL.Query().Get("client_ids")
	if clientIDsStr == "" {
		http.Error(w, "client_ids parameter required", http.StatusBadRequest)
		return
	}

	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "equity"
	}

	idStrs := strings.Split(clientIDsStr, ",")
	clientIDs := make([]int64, 0)

	for _, idStr := range idStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
		if err == nil {
			clientIDs = append(clientIDs, id)
		}
	}

	if len(clientIDs) == 0 {
		http.Error(w, "No valid client IDs provided", http.StatusBadRequest)
		return
	}

	comparison := h.service.CompareClients(clientIDs, metric)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(comparison)
}

// HandleGetStats returns overall portfolio analytics
func (h *ClientPortfolioHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()

	// Round AUM and avg equity to 2 decimal places
	stats.TotalAUM = math.Round(stats.TotalAUM*100) / 100
	stats.AvgEquity = math.Round(stats.AvgEquity*100) / 100
	stats.AvgWinRate = math.Round(stats.AvgWinRate*10000) / 10000

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
