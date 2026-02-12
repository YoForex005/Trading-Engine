package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SocialSignalProvider represents a social trading signal provider
type SocialSignalProvider struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Username        string    `json:"username"`
	DisplayName     string    `json:"display_name"`
	Status          string    `json:"status"` // verified, pending_review, suspended, rejected
	TotalReturn     float64   `json:"total_return_percent"`
	MonthlyReturn   float64   `json:"monthly_return_percent"`
	MaxDrawdown     float64   `json:"max_drawdown_percent"`
	FollowersCount  int       `json:"followers_count"`
	CopiersCount    int       `json:"copiers_count"`
	WinRate         float64   `json:"win_rate"`
	TotalTrades     int       `json:"total_trades"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	ActiveSince     time.Time `json:"active_since"`
	LastTradeAt     time.Time `json:"last_trade_at"`
	MinimumCopy     float64   `json:"minimum_copy_amount"`
	Description     string    `json:"description"`
	Strategy        string    `json:"strategy"`
	RiskLevel       string    `json:"risk_level"` // low, medium, high, extreme
}

// CopyRelationship represents a follower-provider copy trading relationship
type CopyRelationship struct {
	ID                int64     `json:"id"`
	FollowerID        int64     `json:"follower_id"`
	FollowerName      string    `json:"follower_name"`
	ProviderID        int64     `json:"provider_id"`
	ProviderName      string    `json:"provider_name"`
	Status            string    `json:"status"` // active, paused, stopped
	CopyMode          string    `json:"copy_mode"` // fixed_lot, proportional, equity_percentage
	AllocationAmount  float64   `json:"allocation_amount"`
	CopyRatio         float64   `json:"copy_ratio"` // multiplier for proportional mode
	EquityPercentage  float64   `json:"equity_percentage,omitempty"` // for equity_percentage mode
	StartedAt         time.Time `json:"started_at"`
	TotalCopiedTrades int       `json:"total_copied_trades"`
	TotalPnL          float64   `json:"total_pnl"`
	MaxDrawdown       float64   `json:"max_drawdown"`
}

// ProviderPerformance detailed performance metrics
type ProviderPerformance struct {
	ProviderID        int64     `json:"provider_id"`
	TotalReturn       float64   `json:"total_return_percent"`
	MonthlyReturn     float64   `json:"monthly_return_percent"`
	WeeklyReturn      float64   `json:"weekly_return_percent"`
	DailyReturn       float64   `json:"daily_return_percent"`
	MaxDrawdown       float64   `json:"max_drawdown_percent"`
	CurrentDrawdown   float64   `json:"current_drawdown_percent"`
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`
	AvgWin            float64   `json:"avg_win"`
	AvgLoss           float64   `json:"avg_loss"`
	ProfitFactor      float64   `json:"profit_factor"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	SortinoRatio      float64   `json:"sortino_ratio"`
	CalmarRatio       float64   `json:"calmar_ratio"`
	ConsistencyScore  float64   `json:"consistency_score"`
	RiskAdjustedReturn float64  `json:"risk_adjusted_return"`
	LastUpdated       time.Time `json:"last_updated"`
}

// CopySettings represents copy trading configuration
type CopySettings struct {
	CopyID            int64   `json:"copy_id"`
	CopyMode          string  `json:"copy_mode"`
	AllocationAmount  float64 `json:"allocation_amount"`
	CopyRatio         float64 `json:"copy_ratio"`
	EquityPercentage  float64 `json:"equity_percentage,omitempty"`
	MaxDailyLoss      float64 `json:"max_daily_loss"`
	MaxTotalLoss      float64 `json:"max_total_loss"`
	StopCopyOnDD      float64 `json:"stop_copy_on_drawdown_percent"`
	CopyStopLoss      bool    `json:"copy_stop_loss"`
	CopyTakeProfit    bool    `json:"copy_take_profit"`
	ReverseMode       bool    `json:"reverse_mode"`
}

// SocialTradingStats platform-wide social trading statistics
type SocialTradingStats struct {
	TotalProviders       int     `json:"total_providers"`
	VerifiedProviders    int     `json:"verified_providers"`
	PendingProviders     int     `json:"pending_providers"`
	SuspendedProviders   int     `json:"suspended_providers"`
	TotalCopiers         int     `json:"total_copiers"`
	ActiveCopyRels       int     `json:"active_copy_relationships"`
	TotalVolumeUSD       float64 `json:"total_volume_usd"`
	AvgProviderReturn    float64 `json:"avg_provider_return_percent"`
	AvgCopierPnL         float64 `json:"avg_copier_pnl"`
	TopProviderReturn    float64 `json:"top_provider_return_percent"`
	TopProviderID        int64   `json:"top_provider_id"`
	TopProviderName      string  `json:"top_provider_name"`
}

// SocialTradingService manages social/copy trading
type SocialTradingService struct {
	mu               sync.RWMutex
	providers        map[int64]*SocialSignalProvider
	copyRelationships map[int64]*CopyRelationship
	copySettings     map[int64]*CopySettings
	nextProviderID   int64
	nextCopyID       int64
}

// NewSocialTradingService creates a new service with mock data
func NewSocialTradingService() *SocialTradingService {
	s := &SocialTradingService{
		providers:        make(map[int64]*SocialSignalProvider),
		copyRelationships: make(map[int64]*CopyRelationship),
		copySettings:     make(map[int64]*CopySettings),
		nextProviderID:   1,
		nextCopyID:       1,
	}
	s.generateMockData()
	return s
}

func (s *SocialTradingService) generateMockData() {
	statuses := []string{"verified", "verified", "verified", "verified", "pending_review", "suspended"}
	strategies := []string{"Trend Following", "Scalping", "Swing Trading", "Breakout", "Grid Trading", "News Trading", "Mean Reversion"}
	riskLevels := []string{"low", "medium", "high", "extreme"}
	copyModes := []string{"fixed_lot", "proportional", "equity_percentage"}

	providerUsernames := []string{
		"ProTrader_Mike", "ForexKing_Alex", "ScalpMaster_Sam", "TrendGuru_Lisa", "SwingPro_Tom",
		"GridExpert_Jane", "NewsHunter_Paul", "PipCollector_Emma", "RiskManager_Bob", "Momentum_Chris",
		"Breakout_Sarah", "DayTrader_John", "Consistent_Mary", "AggressivePro_Dan", "ConservativeTrade_Anna",
	}

	// Generate 15 signal providers
	for i := 0; i < 15; i++ {
		providerID := s.nextProviderID
		s.nextProviderID++

		status := statuses[i%len(statuses)]
		strategy := strategies[i%len(strategies)]
		riskLevel := riskLevels[i%len(riskLevels)]

		// Performance varies by risk level
		var totalReturn, monthlyReturn, maxDD, winRate, sharpe float64
		totalTrades := 100 + rand.Intn(900)

		switch riskLevel {
		case "low":
			totalReturn = 5 + rand.Float64()*20    // 5-25%
			monthlyReturn = 1 + rand.Float64()*3   // 1-4%
			maxDD = 3 + rand.Float64()*7           // 3-10%
			winRate = 0.55 + rand.Float64()*0.15   // 55-70%
			sharpe = 1.2 + rand.Float64()*0.8      // 1.2-2.0
		case "medium":
			totalReturn = 20 + rand.Float64()*40   // 20-60%
			monthlyReturn = 3 + rand.Float64()*5   // 3-8%
			maxDD = 8 + rand.Float64()*12          // 8-20%
			winRate = 0.50 + rand.Float64()*0.15   // 50-65%
			sharpe = 0.8 + rand.Float64()*0.7      // 0.8-1.5
		case "high":
			totalReturn = 50 + rand.Float64()*80   // 50-130%
			monthlyReturn = 7 + rand.Float64()*10  // 7-17%
			maxDD = 15 + rand.Float64()*20         // 15-35%
			winRate = 0.45 + rand.Float64()*0.15   // 45-60%
			sharpe = 0.5 + rand.Float64()*0.6      // 0.5-1.1
		case "extreme":
			totalReturn = 100 + rand.Float64()*200 // 100-300%
			monthlyReturn = 15 + rand.Float64()*25 // 15-40%
			maxDD = 30 + rand.Float64()*40         // 30-70%
			winRate = 0.40 + rand.Float64()*0.15   // 40-55%
			sharpe = 0.3 + rand.Float64()*0.5      // 0.3-0.8
		}

		followersCount := rand.Intn(500) + 10
		copiersCount := int(float64(followersCount) * (0.3 + rand.Float64()*0.4)) // 30-70% of followers

		provider := &SocialSignalProvider{
			ID:             providerID,
			UserID:         1000 + providerID,
			Username:       providerUsernames[i],
			DisplayName:    providerUsernames[i],
			Status:         status,
			TotalReturn:    totalReturn,
			MonthlyReturn:  monthlyReturn,
			MaxDrawdown:    maxDD,
			FollowersCount: followersCount,
			CopiersCount:   copiersCount,
			WinRate:        winRate,
			TotalTrades:    totalTrades,
			SharpeRatio:    sharpe,
			ActiveSince:    time.Now().AddDate(0, -rand.Intn(36), 0), // up to 3 years
			LastTradeAt:    time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
			MinimumCopy:    []float64{100, 200, 500, 1000, 2000}[rand.Intn(5)],
			Description:    fmt.Sprintf("Experienced %s trader with %s risk profile", strategy, riskLevel),
			Strategy:       strategy,
			RiskLevel:      riskLevel,
		}
		s.providers[providerID] = provider
	}

	// Generate 30 copy relationships
	followerNames := []string{
		"Alice_Investor", "Bob_Copier", "Charlie_Follower", "Diana_Trader", "Edward_Copy",
		"Fiona_Mirror", "George_Auto", "Hannah_Smart", "Ian_Track", "Julia_Follow",
	}

	for i := 0; i < 30; i++ {
		copyID := s.nextCopyID
		s.nextCopyID++

		// Pick random provider (only verified ones for active copies)
		providerID := int64(rand.Intn(12) + 1) // first 12 are more likely verified
		provider, exists := s.providers[providerID]
		if !exists {
			continue
		}

		followerID := int64(i%10 + 1)
		followerName := followerNames[followerID-1]

		status := "active"
		if rand.Float64() < 0.15 {
			status = []string{"paused", "stopped"}[rand.Intn(2)]
		}

		copyMode := copyModes[i%len(copyModes)]
		allocationAmount := 1000 + rand.Float64()*9000 // $1k-$10k
		copyRatio := 0.1 + rand.Float64()*1.9          // 0.1x - 2.0x
		equityPercentage := 0.0
		if copyMode == "equity_percentage" {
			equityPercentage = 5 + rand.Float64()*20 // 5-25%
		}

		startedAt := time.Now().AddDate(0, 0, -rand.Intn(180))
		totalCopiedTrades := rand.Intn(200) + 10
		totalPnL := (rand.Float64()*4000 - 1000) // -$1k to +$3k
		maxDD := rand.Float64() * 800

		copyRel := &CopyRelationship{
			ID:                copyID,
			FollowerID:        followerID,
			FollowerName:      followerName,
			ProviderID:        providerID,
			ProviderName:      provider.Username,
			Status:            status,
			CopyMode:          copyMode,
			AllocationAmount:  allocationAmount,
			CopyRatio:         copyRatio,
			EquityPercentage:  equityPercentage,
			StartedAt:         startedAt,
			TotalCopiedTrades: totalCopiedTrades,
			TotalPnL:          totalPnL,
			MaxDrawdown:       maxDD,
		}
		s.copyRelationships[copyID] = copyRel

		// Generate copy settings
		settings := &CopySettings{
			CopyID:           copyID,
			CopyMode:         copyMode,
			AllocationAmount: allocationAmount,
			CopyRatio:        copyRatio,
			EquityPercentage: equityPercentage,
			MaxDailyLoss:     200 + rand.Float64()*800,
			MaxTotalLoss:     500 + rand.Float64()*2000,
			StopCopyOnDD:     10 + rand.Float64()*20,
			CopyStopLoss:     rand.Float64() > 0.3,
			CopyTakeProfit:   rand.Float64() > 0.3,
			ReverseMode:      rand.Float64() > 0.9, // 10% use reverse
		}
		s.copySettings[copyID] = settings
	}

	log.Printf("[SocialTrading] Generated mock data: %d providers, %d copy relationships",
		len(s.providers), len(s.copyRelationships))
}

// SocialTradingHandler handles HTTP requests
type SocialTradingHandler struct {
	service     *SocialTradingService
	authService *AuthService
}

// NewSocialTradingHandler creates a new handler
func NewSocialTradingHandler(service *SocialTradingService, authService *AuthService) *SocialTradingHandler {
	return &SocialTradingHandler{
		service:     service,
		authService: authService,
	}
}

// HandleListProviders handles GET /admin/social-trading/providers
func (h *SocialTradingHandler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	// Get filter parameters
	statusFilter := r.URL.Query().Get("status")
	rankBy := r.URL.Query().Get("rank_by") // return, risk_adjusted, consistency

	providers := make([]*SocialSignalProvider, 0)
	for _, provider := range h.service.providers {
		if statusFilter != "" && provider.Status != statusFilter {
			continue
		}
		providers = append(providers, provider)
	}

	// Sort by ranking criteria
	switch rankBy {
	case "risk_adjusted":
		sort.Slice(providers, func(i, j int) bool {
			// Risk-adjusted return = total return / max drawdown
			riski := providers[i].TotalReturn / math.Max(providers[i].MaxDrawdown, 1)
			riskj := providers[j].TotalReturn / math.Max(providers[j].MaxDrawdown, 1)
			return riski > riskj
		})
	case "consistency":
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].SharpeRatio > providers[j].SharpeRatio
		})
	default: // return
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].TotalReturn > providers[j].TotalReturn
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"data":     providers,
		"total":    len(providers),
		"rank_by":  rankBy,
	})
}

// HandleGetProvider handles GET /admin/social-trading/providers/:id
func (h *SocialTradingHandler) HandleGetProvider(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	providerID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	provider, exists := h.service.providers[providerID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Get followers (copiers)
	followers := make([]*CopyRelationship, 0)
	for _, copy := range h.service.copyRelationships {
		if copy.ProviderID == providerID {
			followers = append(followers, copy)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"data":      provider,
		"followers": followers,
	})
}

// HandleUpdateProvider handles PUT /admin/social-trading/providers/:id
func (h *SocialTradingHandler) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	providerID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Status      string  `json:"status"`
		MinimumCopy float64 `json:"minimum_copy_amount"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.service.mu.Lock()
	defer h.service.mu.Unlock()

	provider, exists := h.service.providers[providerID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Update provider
	if updateData.Status != "" {
		provider.Status = updateData.Status
	}
	if updateData.MinimumCopy > 0 {
		provider.MinimumCopy = updateData.MinimumCopy
	}
	if updateData.Description != "" {
		provider.Description = updateData.Description
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Provider updated successfully",
		"data":    provider,
	})
}

// HandleGetProviderPerformance handles GET /admin/social-trading/providers/:id/performance
func (h *SocialTradingHandler) HandleGetProviderPerformance(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	providerID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	provider, exists := h.service.providers[providerID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	// Calculate detailed performance metrics
	winningTrades := int(float64(provider.TotalTrades) * provider.WinRate)
	losingTrades := provider.TotalTrades - winningTrades

	avgWin := 0.0
	avgLoss := 0.0
	profitFactor := 0.0

	if winningTrades > 0 {
		avgWin = 150 + rand.Float64()*200
	}
	if losingTrades > 0 {
		avgLoss = 80 + rand.Float64()*120
	}
	if avgLoss > 0 {
		profitFactor = (avgWin * float64(winningTrades)) / (avgLoss * float64(losingTrades))
	}

	// Calculate additional ratios
	sortinoRatio := provider.SharpeRatio * 1.4 // Sortino typically higher
	calmarRatio := provider.TotalReturn / math.Max(provider.MaxDrawdown, 1)
	consistencyScore := (provider.WinRate * 0.4) + (math.Min(provider.SharpeRatio/3, 0.3) * 0.6)
	riskAdjustedReturn := provider.TotalReturn / math.Max(provider.MaxDrawdown, 1)

	performance := &ProviderPerformance{
		ProviderID:         providerID,
		TotalReturn:        provider.TotalReturn,
		MonthlyReturn:      provider.MonthlyReturn,
		WeeklyReturn:       provider.MonthlyReturn / 4.33,
		DailyReturn:        provider.MonthlyReturn / 30,
		MaxDrawdown:        provider.MaxDrawdown,
		CurrentDrawdown:    rand.Float64() * provider.MaxDrawdown * 0.5,
		TotalTrades:        provider.TotalTrades,
		WinningTrades:      winningTrades,
		LosingTrades:       losingTrades,
		WinRate:            provider.WinRate,
		AvgWin:             avgWin,
		AvgLoss:            avgLoss,
		ProfitFactor:       profitFactor,
		SharpeRatio:        provider.SharpeRatio,
		SortinoRatio:       sortinoRatio,
		CalmarRatio:        calmarRatio,
		ConsistencyScore:   consistencyScore,
		RiskAdjustedReturn: riskAdjustedReturn,
		LastUpdated:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    performance,
	})
}

// HandleListCopyRelationships handles GET /admin/social-trading/copies
func (h *SocialTradingHandler) HandleListCopyRelationships(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	// Get filter parameters
	statusFilter := r.URL.Query().Get("status")
	providerIDStr := r.URL.Query().Get("provider_id")
	followerIDStr := r.URL.Query().Get("follower_id")

	var providerIDFilter, followerIDFilter int64
	if providerIDStr != "" {
		providerIDFilter, _ = strconv.ParseInt(providerIDStr, 10, 64)
	}
	if followerIDStr != "" {
		followerIDFilter, _ = strconv.ParseInt(followerIDStr, 10, 64)
	}

	copies := make([]*CopyRelationship, 0)
	for _, copy := range h.service.copyRelationships {
		if statusFilter != "" && copy.Status != statusFilter {
			continue
		}
		if providerIDFilter > 0 && copy.ProviderID != providerIDFilter {
			continue
		}
		if followerIDFilter > 0 && copy.FollowerID != followerIDFilter {
			continue
		}
		copies = append(copies, copy)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    copies,
		"total":   len(copies),
	})
}

// HandleGetCopyRelationship handles GET /admin/social-trading/copies/:id
func (h *SocialTradingHandler) HandleGetCopyRelationship(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid copy ID", http.StatusBadRequest)
		return
	}

	copyID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid copy ID", http.StatusBadRequest)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	copy, exists := h.service.copyRelationships[copyID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Copy relationship not found", http.StatusNotFound)
		return
	}

	settings := h.service.copySettings[copyID]

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"data":     copy,
		"settings": settings,
	})
}

// HandleUpdateCopySettings handles PUT /admin/social-trading/copies/:id
func (h *SocialTradingHandler) HandleUpdateCopySettings(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid copy ID", http.StatusBadRequest)
		return
	}

	copyID, err := strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid copy ID", http.StatusBadRequest)
		return
	}

	var updateData struct {
		Status           string  `json:"status"`
		AllocationAmount float64 `json:"allocation_amount"`
		CopyRatio        float64 `json:"copy_ratio"`
		EquityPercentage float64 `json:"equity_percentage"`
		MaxDailyLoss     float64 `json:"max_daily_loss"`
		MaxTotalLoss     float64 `json:"max_total_loss"`
		StopCopyOnDD     float64 `json:"stop_copy_on_drawdown_percent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.service.mu.Lock()
	defer h.service.mu.Unlock()

	copy, exists := h.service.copyRelationships[copyID]
	if !exists {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Copy relationship not found", http.StatusNotFound)
		return
	}

	settings, settingsExist := h.service.copySettings[copyID]
	if !settingsExist {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Copy settings not found", http.StatusNotFound)
		return
	}

	// Update copy relationship
	if updateData.Status != "" {
		copy.Status = updateData.Status
	}
	if updateData.AllocationAmount > 0 {
		copy.AllocationAmount = updateData.AllocationAmount
		settings.AllocationAmount = updateData.AllocationAmount
	}
	if updateData.CopyRatio > 0 {
		copy.CopyRatio = updateData.CopyRatio
		settings.CopyRatio = updateData.CopyRatio
	}
	if updateData.EquityPercentage > 0 {
		copy.EquityPercentage = updateData.EquityPercentage
		settings.EquityPercentage = updateData.EquityPercentage
	}

	// Update settings
	if updateData.MaxDailyLoss > 0 {
		settings.MaxDailyLoss = updateData.MaxDailyLoss
	}
	if updateData.MaxTotalLoss > 0 {
		settings.MaxTotalLoss = updateData.MaxTotalLoss
	}
	if updateData.StopCopyOnDD > 0 {
		settings.StopCopyOnDD = updateData.StopCopyOnDD
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Copy settings updated successfully",
		"data":     copy,
		"settings": settings,
	})
}

// HandleGetStats handles GET /admin/social-trading/stats
func (h *SocialTradingHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.service.mu.RLock()
	defer h.service.mu.RUnlock()

	stats := &SocialTradingStats{
		TotalProviders: len(h.service.providers),
	}

	totalReturn := 0.0
	var topProvider *SocialSignalProvider
	uniqueCopiers := make(map[int64]bool)
	totalVolume := 0.0

	for _, provider := range h.service.providers {
		switch provider.Status {
		case "verified":
			stats.VerifiedProviders++
		case "pending_review":
			stats.PendingProviders++
		case "suspended":
			stats.SuspendedProviders++
		}

		totalReturn += provider.TotalReturn

		if topProvider == nil || provider.TotalReturn > topProvider.TotalReturn {
			topProvider = provider
		}
	}

	if len(h.service.providers) > 0 {
		stats.AvgProviderReturn = totalReturn / float64(len(h.service.providers))
	}

	if topProvider != nil {
		stats.TopProviderReturn = topProvider.TotalReturn
		stats.TopProviderID = topProvider.ID
		stats.TopProviderName = topProvider.Username
	}

	totalCopierPnL := 0.0
	for _, copy := range h.service.copyRelationships {
		if copy.Status == "active" {
			stats.ActiveCopyRels++
		}
		uniqueCopiers[copy.FollowerID] = true
		totalCopierPnL += copy.TotalPnL
		totalVolume += copy.AllocationAmount
	}

	stats.TotalCopiers = len(uniqueCopiers)
	stats.TotalVolumeUSD = totalVolume

	if stats.TotalCopiers > 0 {
		stats.AvgCopierPnL = totalCopierPnL / float64(stats.TotalCopiers)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}
