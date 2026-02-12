package admin

import (
	"encoding/json"
	"errors"
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

	"github.com/gorilla/mux"
)

// StrategyProvider represents a copy trading strategy provider
type StrategyProvider struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	AccountID      string    `json:"accountId"`
	ReturnPct      float64   `json:"returnPct"`      // Total return percentage
	WinRate        float64   `json:"winRate"`        // Win rate (0-100)
	MaxDrawdown    float64   `json:"maxDrawdown"`    // Maximum drawdown percentage
	TotalTrades    int       `json:"totalTrades"`    // Total number of trades
	Followers      int       `json:"followers"`      // Number of followers
	ProfitFactor   float64   `json:"profitFactor"`   // Profit factor
	SharpeRatio    float64   `json:"sharpeRatio"`    // Sharpe ratio
	MonthlyReturns []float64 `json:"monthlyReturns"` // Last 12 months returns
	RiskLevel      string    `json:"riskLevel"`      // low, medium, high
	Status         string    `json:"status"`         // active, suspended, pending
	Strategy       string    `json:"strategy"`       // Strategy description
	CreatedAt      time.Time `json:"createdAt"`
	// Admin-only fields
	AccountBalance  float64   `json:"accountBalance,omitempty"`  // Only for admin
	LastTradeAt     time.Time `json:"lastTradeAt,omitempty"`     // Only for admin
	CommissionEarned float64  `json:"commissionEarned,omitempty"` // Only for admin
}

// CopyFollow represents a copy trading relationship
type CopyFollow struct {
	ID                string    `json:"id"`
	FollowerAccountID string    `json:"followerAccountId"`
	ProviderID        string    `json:"providerId"`
	LotMultiplier     float64   `json:"lotMultiplier"`  // Position size multiplier
	MaxPositions      int       `json:"maxPositions"`   // Max concurrent positions
	StopLossPct       float64   `json:"stopLossPct"`    // Stop loss percentage
	Status            string    `json:"status"`         // active, paused, closed
	StartedAt         time.Time `json:"startedAt"`
	TotalPnL          float64   `json:"totalPnL"`       // Total P&L since following
	CopiedTrades      int       `json:"copiedTrades"`   // Number of copied trades
}

// CopyTradingService manages copy trading functionality
type CopyTradingService struct {
	mu               sync.RWMutex
	providers        map[string]*StrategyProvider
	follows          map[string]*CopyFollow
	followerFollows  map[string][]string // followerAccountID -> []followID
	nextProviderID   int
	nextFollowID     int
}

// NewCopyTradingService creates a new copy trading service with mock data
func NewCopyTradingService() *CopyTradingService {
	svc := &CopyTradingService{
		providers:       make(map[string]*StrategyProvider),
		follows:         make(map[string]*CopyFollow),
		followerFollows: make(map[string][]string),
		nextProviderID:  21,
		nextFollowID:    1,
	}

	svc.createMockProviders()
	return svc
}

// createMockProviders creates 20 mock strategy providers with varied profiles
func (s *CopyTradingService) createMockProviders() {
	providers := []struct {
		name        string
		description string
		riskLevel   string
		strategy    string
		returnPct   float64
		winRate     float64
		maxDrawdown float64
		totalTrades int
		followers   int
		status      string
	}{
		// High-risk/high-return strategies
		{"Alex TrendMaster", "Aggressive trend following with high leverage", "high", "Trend Following + Momentum", 185.5, 58.2, 42.8, 342, 487, "active"},
		{"Maria Scalper Pro", "High-frequency scalping on major pairs", "high", "Scalping + Range Trading", 142.3, 62.1, 38.5, 1523, 312, "active"},
		{"Chen Volatility King", "Volatility breakout trading", "high", "Breakout + Volatility", 198.7, 54.3, 48.2, 287, 256, "active"},
		{"Rajesh Momentum", "Momentum trading with tight stops", "high", "Momentum + Price Action", 167.2, 59.8, 35.6, 445, 198, "active"},

		// Medium-risk strategies
		{"Sarah Balanced", "Balanced approach with risk management", "medium", "Multi-Strategy Balance", 87.4, 64.5, 22.3, 567, 834, "active"},
		{"Mike Swing Trader", "Swing trading major currency pairs", "medium", "Swing Trading", 95.2, 61.7, 24.8, 423, 672, "active"},
		{"Elena Price Action", "Pure price action trading", "medium", "Price Action", 78.9, 63.2, 19.5, 512, 543, "active"},
		{"David MultiPair", "Multi-pair correlation strategy", "medium", "Correlation + Diversification", 102.5, 60.4, 26.7, 389, 456, "active"},
		{"Lisa Smart Money", "Smart money concepts and order flow", "medium", "Smart Money + Order Flow", 92.8, 65.1, 21.2, 478, 721, "active"},
		{"Tom Risk Manager", "Conservative risk with steady gains", "medium", "Risk Management Focus", 68.3, 67.8, 15.4, 634, 892, "active"},

		// Low-risk/conservative strategies
		{"Robert Conservative", "Low-risk hedging strategy", "low", "Hedging + Income", 45.7, 72.3, 11.2, 789, 1234, "active"},
		{"Anna SafeTrader", "Capital preservation with steady growth", "low", "Capital Preservation", 38.2, 74.5, 8.9, 891, 1567, "active"},
		{"James Steady", "Long-term position trading", "low", "Position Trading", 52.4, 69.7, 12.8, 567, 998, "active"},
		{"Sophie Income", "Income-focused carry trades", "low", "Carry Trade + Income", 41.8, 71.2, 9.6, 723, 1432, "active"},

		// Mixed performance (some pending/suspended)
		{"Kevin Crypto", "Cryptocurrency momentum trading", "high", "Crypto Momentum", 112.5, 56.8, 31.4, 298, 234, "active"},
		{"Nina Grid Trader", "Grid trading system", "medium", "Grid Trading", 55.6, 68.4, 18.2, 445, 387, "active"},
		{"Oscar Algo", "Algorithmic trading system", "medium",
		"Algorithmic + ML", 83.7, 62.9, 23.5, 756, 512, "active"},
		{"Paula Fibonacci", "Fibonacci retracement strategy", "medium", "Fibonacci + Support/Resistance", 71.2, 64.3, 20.8, 489, 434, "active"},
		{"Quinn Reversal", "Mean reversion trading", "medium", "Mean Reversion", 64.8, 66.7, 17.9, 534, 378, "pending"},
		{"Ryan NewsTrader", "News and fundamental analysis", "high", "News Trading + Fundamentals", 98.3, 57.5, 28.6, 312, 167, "suspended"},
	}

	rand.Seed(time.Now().UnixNano())
	baseTime := time.Now().AddDate(0, -6, 0) // 6 months ago

	for i, p := range providers {
		id := fmt.Sprintf("provider-%d", i+1)

		// Generate realistic monthly returns
		monthlyReturns := make([]float64, 12)
		avgReturn := p.returnPct / 12.0
		volatility := p.maxDrawdown / 3.0

		for month := 0; month < 12; month++ {
			// Add randomness to monthly returns
			monthReturn := avgReturn + (rand.Float64()-0.5)*volatility*2
			// Ensure some negative months for realism
			if rand.Float64() < 0.25 { // 25% chance of negative month
				monthReturn = -math.Abs(monthReturn) * 0.6
			}
			monthlyReturns[month] = math.Round(monthReturn*100) / 100
		}

		// Calculate profit factor based on win rate
		profitFactor := (p.winRate / (100 - p.winRate)) * (1 + rand.Float64()*0.5)
		profitFactor = math.Round(profitFactor*100) / 100

		// Calculate Sharpe ratio based on risk level
		sharpeRatio := 0.0
		switch p.riskLevel {
		case "low":
			sharpeRatio = 1.5 + rand.Float64()*1.0
		case "medium":
			sharpeRatio = 1.0 + rand.Float64()*1.0
		case "high":
			sharpeRatio = 0.5 + rand.Float64()*1.0
		}
		sharpeRatio = math.Round(sharpeRatio*100) / 100

		provider := &StrategyProvider{
			ID:              id,
			Name:            p.name,
			Description:     p.description,
			AccountID:       fmt.Sprintf("ACC-%d", 10000+i),
			ReturnPct:       p.returnPct,
			WinRate:         p.winRate,
			MaxDrawdown:     p.maxDrawdown,
			TotalTrades:     p.totalTrades,
			Followers:       p.followers,
			ProfitFactor:    profitFactor,
			SharpeRatio:     sharpeRatio,
			MonthlyReturns:  monthlyReturns,
			RiskLevel:       p.riskLevel,
			Status:          p.status,
			Strategy:        p.strategy,
			CreatedAt:       baseTime.AddDate(0, 0, i*10),
			AccountBalance:  10000 + (p.returnPct/100)*10000 + rand.Float64()*5000,
			LastTradeAt:     time.Now().Add(-time.Duration(rand.Intn(72)) * time.Hour),
			CommissionEarned: float64(p.followers) * rand.Float64() * 500,
		}

		s.providers[id] = provider
	}

	log.Printf("[CopyTrading] Created 20 mock strategy providers")
}

// GetProviders returns all active providers (public endpoint)
func (s *CopyTradingService) GetProviders(sortBy string) []*StrategyProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*StrategyProvider, 0)
	for _, p := range s.providers {
		if p.Status == "active" {
			// Clone without admin fields
			pub := &StrategyProvider{
				ID:             p.ID,
				Name:           p.Name,
				Description:    p.Description,
				AccountID:      p.AccountID,
				ReturnPct:      p.ReturnPct,
				WinRate:        p.WinRate,
				MaxDrawdown:    p.MaxDrawdown,
				TotalTrades:    p.TotalTrades,
				Followers:      p.Followers,
				ProfitFactor:   p.ProfitFactor,
				SharpeRatio:    p.SharpeRatio,
				MonthlyReturns: p.MonthlyReturns,
				RiskLevel:      p.RiskLevel,
				Status:         p.Status,
				Strategy:       p.Strategy,
				CreatedAt:      p.CreatedAt,
			}
			providers = append(providers, pub)
		}
	}

	// Sort providers
	switch sortBy {
	case "return":
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].ReturnPct > providers[j].ReturnPct
		})
	case "winRate":
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].WinRate > providers[j].WinRate
		})
	case "sharpe":
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].SharpeRatio > providers[j].SharpeRatio
		})
	case "followers":
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].Followers > providers[j].Followers
		})
	default:
		// Default sort by return
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].ReturnPct > providers[j].ReturnPct
		})
	}

	return providers
}

// GetProviderByID returns a provider by ID (public endpoint)
func (s *CopyTradingService) GetProviderByID(providerID string) (*StrategyProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[providerID]
	if !exists {
		return nil, errors.New("provider not found")
	}

	// Return clone without admin fields
	return &StrategyProvider{
		ID:             provider.ID,
		Name:           provider.Name,
		Description:    provider.Description,
		AccountID:      provider.AccountID,
		ReturnPct:      provider.ReturnPct,
		WinRate:        provider.WinRate,
		MaxDrawdown:    provider.MaxDrawdown,
		TotalTrades:    provider.TotalTrades,
		Followers:      provider.Followers,
		ProfitFactor:   provider.ProfitFactor,
		SharpeRatio:    provider.SharpeRatio,
		MonthlyReturns: provider.MonthlyReturns,
		RiskLevel:      provider.RiskLevel,
		Status:         provider.Status,
		Strategy:       provider.Strategy,
		CreatedAt:      provider.CreatedAt,
	}, nil
}

// FollowProvider creates a new copy trading relationship
func (s *CopyTradingService) FollowProvider(followerAccountID, providerID string, lotMultiplier float64, maxPositions int, stopLossPct float64) (*CopyFollow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if provider exists and is active
	provider, exists := s.providers[providerID]
	if !exists {
		return nil, errors.New("provider not found")
	}
	if provider.Status != "active" {
		return nil, errors.New("provider is not active")
	}

	// Check if already following
	if follows, ok := s.followerFollows[followerAccountID]; ok {
		for _, followID := range follows {
			if follow, ok := s.follows[followID]; ok && follow.ProviderID == providerID && follow.Status == "active" {
				return nil, errors.New("already following this provider")
			}
		}
	}

	// Create follow
	followID := fmt.Sprintf("follow-%d", s.nextFollowID)
	s.nextFollowID++

	follow := &CopyFollow{
		ID:                followID,
		FollowerAccountID: followerAccountID,
		ProviderID:        providerID,
		LotMultiplier:     lotMultiplier,
		MaxPositions:      maxPositions,
		StopLossPct:       stopLossPct,
		Status:            "active",
		StartedAt:         time.Now(),
		TotalPnL:          0.0,
		CopiedTrades:      0,
	}

	s.follows[followID] = follow

	// Add to follower's follows
	if _, ok := s.followerFollows[followerAccountID]; !ok {
		s.followerFollows[followerAccountID] = make([]string, 0)
	}
	s.followerFollows[followerAccountID] = append(s.followerFollows[followerAccountID], followID)

	// Increment provider's follower count
	provider.Followers++

	log.Printf("[CopyTrading] Account %s started following provider %s", followerAccountID, providerID)

	return follow, nil
}

// UnfollowProvider stops a copy trading relationship
func (s *CopyTradingService) UnfollowProvider(followID string, followerAccountID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	follow, exists := s.follows[followID]
	if !exists {
		return errors.New("follow relationship not found")
	}

	// Verify ownership
	if follow.FollowerAccountID != followerAccountID {
		return errors.New("unauthorized to unfollow")
	}

	// Update status
	follow.Status = "closed"

	// Decrement provider's follower count
	if provider, ok := s.providers[follow.ProviderID]; ok {
		if provider.Followers > 0 {
			provider.Followers--
		}
	}

	log.Printf("[CopyTrading] Account %s unfollowed provider %s", followerAccountID, follow.ProviderID)

	return nil
}

// GetFollowing returns all active follows for a follower
func (s *CopyTradingService) GetFollowing(followerAccountID string) []*CopyFollow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*CopyFollow, 0)

	if followIDs, ok := s.followerFollows[followerAccountID]; ok {
		for _, followID := range followIDs {
			if follow, ok := s.follows[followID]; ok && follow.Status == "active" {
				result = append(result, follow)
			}
		}
	}

	return result
}

// GetProvidersAdmin returns all providers with admin fields
func (s *CopyTradingService) GetProvidersAdmin() []*StrategyProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*StrategyProvider, 0, len(s.providers))
	for _, p := range s.providers {
		providers = append(providers, p)
	}

	// Sort by followers (descending)
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Followers > providers[j].Followers
	})

	return providers
}

// UpdateProviderStatus updates a provider's status (admin only)
func (s *CopyTradingService) UpdateProviderStatus(providerID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, exists := s.providers[providerID]
	if !exists {
		return errors.New("provider not found")
	}

	// Validate status
	if status != "active" && status != "suspended" && status != "pending" {
		return errors.New("invalid status: must be active, suspended, or pending")
	}

	provider.Status = status

	log.Printf("[CopyTrading] Provider %s status updated to %s", providerID, status)

	return nil
}

// HTTP Handlers

// HandleGetProviders handles GET /api/copy-trading/providers (public)
func (s *CopyTradingService) HandleGetProviders(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	sortBy := r.URL.Query().Get("sortBy")
	providers := s.GetProviders(sortBy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"providers": providers,
		"count":     len(providers),
	})
}

// HandleGetProviderDetail handles GET /api/copy-trading/providers/:id (public)
func (s *CopyTradingService) HandleGetProviderDetail(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	providerID := vars["id"]

	provider, err := s.GetProviderByID(providerID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"provider": provider,
	})
}

// HandleFollowProvider handles POST /api/copy-trading/follow (client auth)
func (s *CopyTradingService) HandleFollowProvider(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	// Parse request body
	var req struct {
		ProviderID    string  `json:"providerId"`
		LotMultiplier float64 `json:"lotMultiplier"`
		MaxPositions  int     `json:"maxPositions"`
		StopLossPct   float64 `json:"stopLossPct"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// TODO: Get followerAccountID from authenticated user context
	// For now, use a mock value
	followerAccountID := "ACC-" + strconv.Itoa(20000+rand.Intn(1000))

	follow, err := s.FollowProvider(followerAccountID, req.ProviderID, req.LotMultiplier, req.MaxPositions, req.StopLossPct)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"follow":  follow,
	})
}

// HandleUnfollowProvider handles DELETE /api/copy-trading/follow/:id (client auth)
func (s *CopyTradingService) HandleUnfollowProvider(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	followID := vars["id"]

	// TODO: Get followerAccountID from authenticated user context
	// For now, extract from the follow record
	s.mu.RLock()
	follow, exists := s.follows[followID]
	if !exists {
		s.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Follow relationship not found",
		})
		return
	}
	followerAccountID := follow.FollowerAccountID
	s.mu.RUnlock()

	err := s.UnfollowProvider(followID, followerAccountID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Successfully unfollowed provider",
	})
}

// HandleGetFollowing handles GET /api/copy-trading/following (client auth)
func (s *CopyTradingService) HandleGetFollowing(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	// TODO: Get followerAccountID from authenticated user context
	// For now, use a mock value
	followerAccountID := "ACC-" + strconv.Itoa(20000+rand.Intn(1000))

	follows := s.GetFollowing(followerAccountID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"follows": follows,
		"count":   len(follows),
	})
}

// HandleGetProvidersAdmin handles GET /admin/copy-trading/providers (admin auth)
func (s *CopyTradingService) HandleGetProvidersAdmin(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	providers := s.GetProvidersAdmin()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"providers": providers,
		"count":     len(providers),
	})
}

// HandleUpdateProviderStatus handles PUT /admin/copy-trading/providers/:id/status (admin auth)
func (s *CopyTradingService) HandleUpdateProviderStatus(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	providerID := vars["id"]

	// Parse request body
	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	err := s.UpdateProviderStatus(providerID, strings.ToLower(req.Status))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Provider status updated to %s", req.Status),
	})
}
