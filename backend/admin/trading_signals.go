package admin

import (
	"encoding/json"
	"fmt"
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
// Trading Signal / Signal Provider Management
// ============================================

type SignalProvider struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	WinRate       float64   `json:"winRate"`       // percentage
	TotalSignals  int       `json:"totalSignals"`
	Subscribers   int       `json:"subscribers"`
	MonthlyFee    float64   `json:"monthlyFee"`    // $
	Rating        float64   `json:"rating"`        // 0-5
	Status        string    `json:"status"`        // active, suspended, pending
	Since         time.Time `json:"since"`
	AvgPips       float64   `json:"avgPips"`
	TotalPips     float64   `json:"totalPips"`
	ClosedSignals int       `json:"closedSignals"`
}

type TradingSignal struct {
	ID          int64      `json:"id"`
	ProviderID  int64      `json:"providerId"`
	ProviderName string    `json:"providerName"`
	Symbol      string     `json:"symbol"`
	Direction   string     `json:"direction"`   // buy, sell
	EntryPrice  float64    `json:"entryPrice"`
	StopLoss    float64    `json:"stopLoss"`
	TakeProfit  float64    `json:"takeProfit"`
	Status      string     `json:"status"`      // active, hit_tp, hit_sl, closed
	CreatedAt   time.Time  `json:"createdAt"`
	ClosedAt    *time.Time `json:"closedAt,omitempty"`
	Result      string     `json:"result,omitempty"` // win, loss, breakeven
	Pips        float64    `json:"pips,omitempty"`
}

type SignalSubscription struct {
	ID         int64     `json:"id"`
	ClientID   int64     `json:"clientId"`
	ClientName string    `json:"clientName"`
	ProviderID int64     `json:"providerId"`
	ProviderName string  `json:"providerName"`
	StartDate  time.Time `json:"startDate"`
	Status     string    `json:"status"`    // active, cancelled, expired
	AutoTrade  bool      `json:"autoTrade"`
	MonthlyFee float64   `json:"monthlyFee"`
}

type SignalStats struct {
	TotalProviders    int     `json:"totalProviders"`
	ActiveProviders   int     `json:"activeProviders"`
	TotalSignals      int     `json:"totalSignals"`
	ActiveSignals     int     `json:"activeSignals"`
	ClosedSignals     int     `json:"closedSignals"`
	AvgWinRate        float64 `json:"avgWinRate"`
	TotalSubscribers  int     `json:"totalSubscribers"`
	MonthlyRevenue    float64 `json:"monthlyRevenue"`
	LastUpdated       time.Time `json:"lastUpdated"`
}

type ProviderUpdate struct {
	Status     string  `json:"status,omitempty"`
	MonthlyFee float64 `json:"monthlyFee,omitempty"`
}

// ============================================
// Service
// ============================================

type TradingSignalService struct {
	providers     map[int64]*SignalProvider
	signals       map[int64]*TradingSignal
	subscriptions map[int64]*SignalSubscription
	nextProviderID int64
	nextSignalID   int64
	nextSubID      int64
	mu             sync.RWMutex
}

func NewTradingSignalService() *TradingSignalService {
	s := &TradingSignalService{
		providers:     make(map[int64]*SignalProvider),
		signals:       make(map[int64]*TradingSignal),
		subscriptions: make(map[int64]*SignalSubscription),
		nextProviderID: 1,
		nextSignalID:   1,
		nextSubID:      1,
	}
	s.generateMockData()
	return s
}

func (s *TradingSignalService) generateMockData() {
	providerNames := []string{
		"ForexMaster Pro", "PipHunter Elite", "TrendFollower X",
		"ScalpKing Premium", "SwingTrade Genius", "DayTrade Signals",
		"FX Profits Daily", "Global Forex Signals", "Market Movers Pro",
		"Trading Academy VIP", "Signal Factory", "Pro Trader Hub",
		"Elite FX Signals", "Forex Winners", "Smart Trade Alerts",
	}

	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
		"NZDUSD", "EURGBP", "EURJPY", "GBPJPY", "AUDJPY",
	}

	clientNames := []string{
		"John Smith", "Emma Johnson", "Michael Brown", "Sarah Davis",
		"James Wilson", "Emily Taylor", "David Anderson", "Jessica Martinez",
		"Robert Thomas", "Jennifer Jackson", "William White", "Mary Harris",
		"Christopher Martin", "Linda Thompson", "Daniel Garcia", "Patricia Robinson",
		"Matthew Clark", "Barbara Rodriguez", "Joseph Lewis", "Susan Lee",
		"Charles Walker", "Nancy Hall", "Thomas Allen", "Lisa Young",
		"Richard King", "Margaret Wright", "Mark Lopez", "Betty Hill",
		"Donald Scott", "Sandra Green", "Paul Adams", "Ashley Baker",
		"Steven Nelson", "Dorothy Carter", "Andrew Mitchell", "Kimberly Perez",
		"Joshua Roberts", "Elizabeth Turner", "Kenneth Phillips", "Donna Campbell",
		"Kevin Parker", "Carol Evans", "Brian Edwards", "Michelle Collins",
		"George Stewart", "Amanda Sanchez", "Edward Morris", "Melissa Rogers",
		"Ronald Reed", "Deborah Cook", "Timothy Morgan", "Stephanie Bell",
		"Jason Murphy", "Rebecca Bailey", "Jeffrey Rivera", "Laura Cooper",
		"Ryan Richardson", "Helen Cox", "Jacob Howard", "Sharon Ward",
		"Gary Torres", "Cynthia Peterson", "Nicholas Gray", "Kathleen Ramirez",
		"Eric James", "Angela Watson", "Jonathan Brooks", "Shirley Kelly",
		"Stephen Sanders", "Anna Price", "Larry Bennett", "Brenda Wood",
		"Justin Ross", "Pamela Henderson", "Scott Coleman", "Nicole Jenkins",
		"Frank Perry", "Christine Powell", "Dennis Long", "Janet Patterson",
		"Jerry Hughes", "Carolyn Flores", "Alexander Washington", "Frances Butler",
	}

	now := time.Now()

	// Generate 15 signal providers
	for i := 0; i < 15; i++ {
		winRate := 45.0 + rand.Float64()*45.0 // 45-90%
		totalSignals := 50 + rand.Intn(450)  // 50-500 signals
		subscribers := rand.Intn(100)        // 0-100 subscribers
		monthlyFee := 50.0 + rand.Float64()*450.0 // $50-$500
		rating := 2.5 + rand.Float64()*2.5  // 2.5-5.0

		var status string
		r := rand.Float64()
		if r < 0.8 {
			status = "active"
		} else if r < 0.95 {
			status = "pending"
		} else {
			status = "suspended"
		}

		since := now.Add(-time.Duration(30+rand.Intn(700)) * 24 * time.Hour) // 1-24 months ago

		closedSignals := int(float64(totalSignals) * 0.6) // 60% closed
		totalPips := float64(closedSignals) * (10.0 + rand.Float64()*40.0) // avg 10-50 pips
		avgPips := 0.0
		if closedSignals > 0 {
			avgPips = totalPips / float64(closedSignals)
		}

		provider := &SignalProvider{
			ID:            s.nextProviderID,
			Name:          providerNames[i],
			WinRate:       winRate,
			TotalSignals:  totalSignals,
			Subscribers:   subscribers,
			MonthlyFee:    monthlyFee,
			Rating:        rating,
			Status:        status,
			Since:         since,
			AvgPips:       avgPips,
			TotalPips:     totalPips,
			ClosedSignals: closedSignals,
		}
		s.providers[provider.ID] = provider
		s.nextProviderID++
	}

	// Generate 500 signals (300 closed, 200 active)
	for i := 0; i < 500; i++ {
		providerID := int64(1 + rand.Intn(15))
		provider := s.providers[providerID]

		symbol := symbols[rand.Intn(len(symbols))]
		direction := "buy"
		if rand.Float64() < 0.5 {
			direction = "sell"
		}

		entryPrice := 1.0000 + rand.Float64()*0.5000

		var stopLoss, takeProfit float64
		if direction == "buy" {
			stopLoss = entryPrice - (0.0010 + rand.Float64()*0.0040)  // 10-50 pips SL
			takeProfit = entryPrice + (0.0020 + rand.Float64()*0.0080) // 20-100 pips TP
		} else {
			stopLoss = entryPrice + (0.0010 + rand.Float64()*0.0040)
			takeProfit = entryPrice - (0.0020 + rand.Float64()*0.0080)
		}

		createdAt := now.Add(-time.Duration(rand.Intn(90*24)) * time.Hour) // Last 90 days

		var status, result string
		var closedAt *time.Time
		var pips float64

		if i < 300 { // Closed signals
			closeTime := createdAt.Add(time.Duration(1+rand.Intn(72)) * time.Hour)
			closedAt = &closeTime

			r := rand.Float64()
			if r < provider.WinRate/100.0 { // Win based on provider's win rate
				status = "hit_tp"
				result = "win"
				if direction == "buy" {
					pips = (takeProfit - entryPrice) * 10000
				} else {
					pips = (entryPrice - takeProfit) * 10000
				}
			} else if r < 0.95 { // Loss
				status = "hit_sl"
				result = "loss"
				if direction == "buy" {
					pips = (stopLoss - entryPrice) * 10000
				} else {
					pips = (entryPrice - stopLoss) * 10000
				}
			} else { // Breakeven
				status = "closed"
				result = "breakeven"
				pips = 0
			}
		} else { // Active signals
			status = "active"
		}

		signal := &TradingSignal{
			ID:           s.nextSignalID,
			ProviderID:   providerID,
			ProviderName: provider.Name,
			Symbol:       symbol,
			Direction:    direction,
			EntryPrice:   entryPrice,
			StopLoss:     stopLoss,
			TakeProfit:   takeProfit,
			Status:       status,
			CreatedAt:    createdAt,
			ClosedAt:     closedAt,
			Result:       result,
			Pips:         pips,
		}
		s.signals[signal.ID] = signal
		s.nextSignalID++
	}

	// Generate 200 subscriptions across 80 clients
	for i := 0; i < 200; i++ {
		clientID := int64(1 + rand.Intn(len(clientNames)))
		clientName := clientNames[clientID-1]
		providerID := int64(1 + rand.Intn(15))
		provider := s.providers[providerID]

		startDate := now.Add(-time.Duration(rand.Intn(180)) * 24 * time.Hour) // Last 6 months

		var status string
		r := rand.Float64()
		if r < 0.75 {
			status = "active"
		} else if r < 0.90 {
			status = "cancelled"
		} else {
			status = "expired"
		}

		autoTrade := rand.Float64() < 0.4 // 40% have auto-trade enabled

		sub := &SignalSubscription{
			ID:           s.nextSubID,
			ClientID:     clientID,
			ClientName:   clientName,
			ProviderID:   providerID,
			ProviderName: provider.Name,
			StartDate:    startDate,
			Status:       status,
			AutoTrade:    autoTrade,
			MonthlyFee:   provider.MonthlyFee,
		}
		s.subscriptions[sub.ID] = sub
		s.nextSubID++
	}

	log.Printf("[TradingSignals] Generated 15 signal providers, 500 signals (300 closed, 200 active), 200 subscriptions across %d clients", len(clientNames))
}

func (s *TradingSignalService) ListProviders(sortBy, sortOrder string) []*SignalProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*SignalProvider, 0, len(s.providers))
	for _, p := range s.providers {
		providers = append(providers, p)
	}

	switch sortBy {
	case "winRate":
		sort.Slice(providers, func(i, j int) bool {
			if sortOrder == "asc" {
				return providers[i].WinRate < providers[j].WinRate
			}
			return providers[i].WinRate > providers[j].WinRate
		})
	case "subscribers":
		sort.Slice(providers, func(i, j int) bool {
			if sortOrder == "asc" {
				return providers[i].Subscribers < providers[j].Subscribers
			}
			return providers[i].Subscribers > providers[j].Subscribers
		})
	case "rating":
		sort.Slice(providers, func(i, j int) bool {
			if sortOrder == "asc" {
				return providers[i].Rating < providers[j].Rating
			}
			return providers[i].Rating > providers[j].Rating
		})
	default:
		sort.Slice(providers, func(i, j int) bool {
			return providers[i].ID < providers[j].ID
		})
	}

	return providers
}

func (s *TradingSignalService) GetProvider(id int64) (*SignalProvider, []*TradingSignal) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider := s.providers[id]
	if provider == nil {
		return nil, nil
	}

	// Get recent signals from this provider
	var signals []*TradingSignal
	for _, sig := range s.signals {
		if sig.ProviderID == id {
			signals = append(signals, sig)
		}
	}

	// Sort by created date desc
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].CreatedAt.After(signals[j].CreatedAt)
	})

	// Return top 20 most recent
	if len(signals) > 20 {
		signals = signals[:20]
	}

	return provider, signals
}

func (s *TradingSignalService) UpdateProvider(id int64, update ProviderUpdate) (*SignalProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	provider, exists := s.providers[id]
	if !exists {
		return nil, fmt.Errorf("provider not found")
	}

	if update.Status != "" {
		provider.Status = update.Status
	}
	if update.MonthlyFee > 0 {
		provider.MonthlyFee = update.MonthlyFee
	}

	return provider, nil
}

func (s *TradingSignalService) GetActiveSignals() []*TradingSignal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var active []*TradingSignal
	for _, sig := range s.signals {
		if sig.Status == "active" {
			active = append(active, sig)
		}
	}

	// Sort by created date desc
	sort.Slice(active, func(i, j int) bool {
		return active[i].CreatedAt.After(active[j].CreatedAt)
	})

	return active
}

func (s *TradingSignalService) GetSignalHistory(filters map[string]string, limit, offset int) ([]*TradingSignal, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*TradingSignal
	for _, sig := range s.signals {
		if sig.Status == "active" {
			continue // Only closed signals
		}

		match := true
		if providerID, ok := filters["providerId"]; ok && providerID != "" {
			id, _ := strconv.ParseInt(providerID, 10, 64)
			if sig.ProviderID != id {
				match = false
			}
		}
		if symbol, ok := filters["symbol"]; ok && symbol != "" && sig.Symbol != symbol {
			match = false
		}
		if result, ok := filters["result"]; ok && result != "" && sig.Result != result {
			match = false
		}

		if match {
			filtered = append(filtered, sig)
		}
	}

	// Sort by closed date desc
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].ClosedAt != nil && filtered[j].ClosedAt != nil {
			return filtered[i].ClosedAt.After(*filtered[j].ClosedAt)
		}
		return false
	})

	total := len(filtered)
	if offset >= total {
		return []*TradingSignal{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *TradingSignalService) GetSubscriptions(filters map[string]string) []*SignalSubscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*SignalSubscription
	for _, sub := range s.subscriptions {
		match := true

		if providerID, ok := filters["providerId"]; ok && providerID != "" {
			id, _ := strconv.ParseInt(providerID, 10, 64)
			if sub.ProviderID != id {
				match = false
			}
		}
		if clientID, ok := filters["clientId"]; ok && clientID != "" {
			id, _ := strconv.ParseInt(clientID, 10, 64)
			if sub.ClientID != id {
				match = false
			}
		}
		if status, ok := filters["status"]; ok && status != "" && sub.Status != status {
			match = false
		}

		if match {
			filtered = append(filtered, sub)
		}
	}

	return filtered
}

func (s *TradingSignalService) GetStats() SignalStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := SignalStats{
		LastUpdated: time.Now(),
	}

	var totalWinRate float64
	var activeProviderCount int

	for _, p := range s.providers {
		stats.TotalProviders++
		if p.Status == "active" {
			stats.ActiveProviders++
			activeProviderCount++
			totalWinRate += p.WinRate
		}
	}

	if activeProviderCount > 0 {
		stats.AvgWinRate = totalWinRate / float64(activeProviderCount)
	}

	for _, sig := range s.signals {
		stats.TotalSignals++
		if sig.Status == "active" {
			stats.ActiveSignals++
		} else {
			stats.ClosedSignals++
		}
	}

	for _, sub := range s.subscriptions {
		if sub.Status == "active" {
			stats.TotalSubscribers++
			stats.MonthlyRevenue += sub.MonthlyFee
		}
	}

	return stats
}

func (s *TradingSignalService) GetLeaderboard() []*SignalProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]*SignalProvider, 0, len(s.providers))
	for _, p := range s.providers {
		if p.Status == "active" {
			providers = append(providers, p)
		}
	}

	// Sort by composite score: winRate * 0.4 + rating * 20 + (subscribers/10)
	sort.Slice(providers, func(i, j int) bool {
		scoreI := providers[i].WinRate*0.4 + providers[i].Rating*20.0 + float64(providers[i].Subscribers)/10.0
		scoreJ := providers[j].WinRate*0.4 + providers[j].Rating*20.0 + float64(providers[j].Subscribers)/10.0
		return scoreI > scoreJ
	})

	// Return top 10
	if len(providers) > 10 {
		providers = providers[:10]
	}

	return providers
}

// ============================================
// HTTP Handlers
// ============================================

type TradingSignalHandler struct {
	service     *TradingSignalService
	authService *auth.Service
}

func NewTradingSignalHandler(service *TradingSignalService, authService *auth.Service) *TradingSignalHandler {
	return &TradingSignalHandler{
		service:     service,
		authService: authService,
	}
}

// GET /admin/signals/providers
func (h *TradingSignalHandler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	sortBy := query.Get("sortBy")
	sortOrder := query.Get("sortOrder")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	providers := h.service.ListProviders(sortBy, sortOrder)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"providers": providers,
		"count":     len(providers),
	})
}

// GET /admin/signals/providers/:id
func (h *TradingSignalHandler) HandleGetProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/signals/providers/"), "/")
	providerID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	provider, signals := h.service.GetProvider(providerID)
	if provider == nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider":      provider,
		"recentSignals": signals,
	})
}

// PUT /admin/signals/providers/:id
func (h *TradingSignalHandler) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/signals/providers/"), "/")
	providerID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid provider ID", http.StatusBadRequest)
		return
	}

	var update ProviderUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	provider, err := h.service.UpdateProvider(providerID, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(provider)
}

// GET /admin/signals/active
func (h *TradingSignalHandler) HandleGetActiveSignals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	signals := h.service.GetActiveSignals()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"signals": signals,
		"count":   len(signals),
	})
}

// GET /admin/signals/history
func (h *TradingSignalHandler) HandleGetSignalHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(query.Get("offset"))

	filters := map[string]string{
		"providerId": query.Get("providerId"),
		"symbol":     query.Get("symbol"),
		"result":     query.Get("result"),
	}

	signals, total := h.service.GetSignalHistory(filters, limit, offset)

	response := map[string]interface{}{
		"signals": signals,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}

	json.NewEncoder(w).Encode(response)
}

// GET /admin/signals/subscriptions
func (h *TradingSignalHandler) HandleGetSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	filters := map[string]string{
		"providerId": query.Get("providerId"),
		"clientId":   query.Get("clientId"),
		"status":     query.Get("status"),
	}

	subscriptions := h.service.GetSubscriptions(filters)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}

// GET /admin/signals/stats
func (h *TradingSignalHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// GET /admin/signals/leaderboard
func (h *TradingSignalHandler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	providers := h.service.GetLeaderboard()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard": providers,
		"count":       len(providers),
	})
}
