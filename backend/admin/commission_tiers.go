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
// Multi-Tier Commission / Rebate Engine
// ============================================

type CommissionTier struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	Level           int     `json:"level"`
	VolumeThreshold float64 `json:"volumeThreshold"` // lots
	CommissionRate  float64 `json:"commissionRate"`  // $ per lot
	RebateRate      float64 `json:"rebateRate"`      // $ per lot
	SpreadMarkup    float64 `json:"spreadMarkup"`    // pips
	IsDefault       bool    `json:"isDefault"`
	ClientCount     int     `json:"clientCount"`
}

type ClientCommission struct {
	ClientID        int64     `json:"clientId"`
	ClientName      string    `json:"clientName"`
	CurrentTier     string    `json:"currentTier"`
	TierID          int64     `json:"tierId"`
	MonthlyVolume   float64   `json:"monthlyVolume"`   // lots
	TotalCommission float64   `json:"totalCommission"` // $
	TotalRebate     float64   `json:"totalRebate"`     // $
	NetCost         float64   `json:"netCost"`         // commission - rebate
	LastCalculated  time.Time `json:"lastCalculated"`
}

type CommissionTransaction struct {
	ID         int64     `json:"id"`
	ClientID   int64     `json:"clientId"`
	ClientName string    `json:"clientName"`
	TradeID    int64     `json:"tradeId"`
	Symbol     string    `json:"symbol"`
	Volume     float64   `json:"volume"`     // lots
	Commission float64   `json:"commission"` // $
	Rebate     float64   `json:"rebate"`     // $
	NetCost    float64   `json:"netCost"`    // commission - rebate
	Tier       string    `json:"tier"`
	Timestamp  time.Time `json:"timestamp"`
}

type TierRule struct {
	ID       int64  `json:"id"`
	TierID   int64  `json:"tierId"`
	Metric   string `json:"metric"`   // volume, trade_count, deposit_amount
	Operator string `json:"operator"` // >=, <=, ==
	Value    string `json:"value"`
	Action   string `json:"action"` // auto_upgrade, auto_downgrade
}

type CommissionStats struct {
	TotalCommissions   float64            `json:"totalCommissions"`
	TotalRebates       float64            `json:"totalRebates"`
	NetCommissions     float64            `json:"netCommissions"`
	AvgPerClient       float64            `json:"avgPerClient"`
	TierDistribution   map[string]int     `json:"tierDistribution"`
	TopTierClients     int                `json:"topTierClients"`
	TotalVolume        float64            `json:"totalVolume"`
	LastUpdated        time.Time          `json:"lastUpdated"`
}

type TierOverrideRequest struct {
	TierID int64  `json:"tierId"`
	Reason string `json:"reason"`
}

// ============================================
// Service
// ============================================

type CommissionTierService struct {
	tiers        map[int64]*CommissionTier
	clients      map[int64]*ClientCommission
	transactions map[int64]*CommissionTransaction
	rules        map[int64]*TierRule
	nextTierID   int64
	nextTxID     int64
	nextRuleID   int64
	mu           sync.RWMutex
}

func NewCommissionTierService() *CommissionTierService {
	s := &CommissionTierService{
		tiers:        make(map[int64]*CommissionTier),
		clients:      make(map[int64]*ClientCommission),
		transactions: make(map[int64]*CommissionTransaction),
		rules:        make(map[int64]*TierRule),
		nextTierID:   1,
		nextTxID:     1,
		nextRuleID:   1,
	}
	s.generateMockData()
	return s
}

func (s *CommissionTierService) generateMockData() {
	// Create 5 commission tiers
	tiers := []struct {
		name            string
		level           int
		volumeThreshold float64
		commissionRate  float64
		rebateRate      float64
		spreadMarkup    float64
		isDefault       bool
	}{
		{"Bronze", 1, 0, 7.0, 0.0, 0.5, true},
		{"Silver", 2, 100, 6.0, 0.5, 0.4, false},
		{"Gold", 3, 500, 5.0, 1.0, 0.3, false},
		{"Platinum", 4, 2000, 4.0, 1.5, 0.2, false},
		{"VIP", 5, 10000, 3.0, 2.0, 0.1, false},
	}

	for _, t := range tiers {
		tier := &CommissionTier{
			ID:              s.nextTierID,
			Name:            t.name,
			Level:           t.level,
			VolumeThreshold: t.volumeThreshold,
			CommissionRate:  t.commissionRate,
			RebateRate:      t.rebateRate,
			SpreadMarkup:    t.spreadMarkup,
			IsDefault:       t.isDefault,
		}
		s.tiers[tier.ID] = tier
		s.nextTierID++
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
		"Tyler Simmons", "Marie Foster", "Aaron Gonzales", "Diane Bryant",
		"Jose Alexander", "Joyce Russell", "Adam Griffin", "Evelyn Hayes",
		"Nathan Diaz", "Ruby Myers", "Zachary Ford", "Phyllis Hamilton",
		"Douglas Graham", "Gloria Sullivan", "Peter Wallace", "Virginia Woods",
		"Henry Barnes", "Kathryn Kennedy", "Carl Lane", "Teresa Cole",
		"Arthur Owens", "Christina Reynolds", "Russell Fisher", "Doris Ellis",
		"Keith Gibson", "Mildred Harrison", "Jeremy Hawkins", "Sara Nichols",
		"Lawrence Chapman", "Rose Mason", "Sean Guerrero", "Theresa Meyer",
		"Christian Reid", "Beverly Dunn", "Albert Medina", "Denise Webb",
		"Austin Morales", "Jacqueline Howell", "Benjamin Romero", "Cheryl Welch",
		"Samuel Graham", "Martha Meyer", "Gabriel Dixon", "Katherine Harvey",
		"Logan Stone", "Judith Curtis", "Jack Jennings", "Ann Oliver",
		"Dylan Lynch", "Julie May", "Noah Robertson", "Victoria Parsons",
		"Elijah Harper", "Amber Knight", "Lucas Cunningham", "Brittany Hunt",
	}

	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
		"NZDUSD", "EURGBP", "EURJPY", "GBPJPY", "AUDJPY",
	}

	// Generate 150 clients with tier assignments
	tierCounts := make(map[int64]int)
	now := time.Now()

	for clientID := int64(1); clientID <= int64(len(clientNames)); clientID++ {
		clientName := clientNames[clientID-1]

		// Distribute clients across tiers: 40% Bronze, 30% Silver, 20% Gold, 8% Platinum, 2% VIP
		var tierID int64
		var tierName string
		roll := rand.Float64()

		if roll < 0.40 {
			tierID = 1
			tierName = "Bronze"
		} else if roll < 0.70 {
			tierID = 2
			tierName = "Silver"
		} else if roll < 0.90 {
			tierID = 3
			tierName = "Gold"
		} else if roll < 0.98 {
			tierID = 4
			tierName = "Platinum"
		} else {
			tierID = 5
			tierName = "VIP"
		}

		tier := s.tiers[tierID]
		tierCounts[tierID]++

		// Generate monthly volume based on tier
		var monthlyVolume float64
		switch tierName {
		case "Bronze":
			monthlyVolume = float64(rand.Intn(100)) // 0-100 lots
		case "Silver":
			monthlyVolume = 100 + float64(rand.Intn(400)) // 100-500 lots
		case "Gold":
			monthlyVolume = 500 + float64(rand.Intn(1500)) // 500-2000 lots
		case "Platinum":
			monthlyVolume = 2000 + float64(rand.Intn(8000)) // 2000-10000 lots
		case "VIP":
			monthlyVolume = 10000 + float64(rand.Intn(20000)) // 10000+ lots
		}

		totalCommission := monthlyVolume * tier.CommissionRate
		totalRebate := monthlyVolume * tier.RebateRate
		netCost := totalCommission - totalRebate

		client := &ClientCommission{
			ClientID:        clientID,
			ClientName:      clientName,
			CurrentTier:     tierName,
			TierID:          tierID,
			MonthlyVolume:   monthlyVolume,
			TotalCommission: totalCommission,
			TotalRebate:     totalRebate,
			NetCost:         netCost,
			LastCalculated:  now,
		}
		s.clients[clientID] = client
	}

	// Update tier client counts
	for tierID, count := range tierCounts {
		s.tiers[tierID].ClientCount = count
	}

	// Generate 500 commission transactions
	numClients := len(clientNames)
	for i := 0; i < 500; i++ {
		clientID := int64(1 + rand.Intn(numClients))
		client := s.clients[clientID]
		tier := s.tiers[client.TierID]

		symbol := symbols[rand.Intn(len(symbols))]
		volume := 0.1 + rand.Float64()*9.9 // 0.1 - 10 lots

		commission := volume * tier.CommissionRate
		rebate := volume * tier.RebateRate
		netCost := commission - rebate

		timestamp := now.Add(-time.Duration(rand.Intn(30*24)) * time.Hour) // Last 30 days

		tx := &CommissionTransaction{
			ID:         s.nextTxID,
			ClientID:   clientID,
			ClientName: client.ClientName,
			TradeID:    int64(10000 + rand.Intn(90000)),
			Symbol:     symbol,
			Volume:     volume,
			Commission: commission,
			Rebate:     rebate,
			NetCost:    netCost,
			Tier:       client.CurrentTier,
			Timestamp:  timestamp,
		}
		s.transactions[tx.ID] = tx
		s.nextTxID++
	}

	log.Printf("[CommissionTiers] Generated 5 tiers, 150 clients (40%% Bronze, 30%% Silver, 20%% Gold, 8%% Platinum, 2%% VIP), 500 transactions")
}

func (s *CommissionTierService) ListTiers() []*CommissionTier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tiers := make([]*CommissionTier, 0, len(s.tiers))
	for _, tier := range s.tiers {
		tiers = append(tiers, tier)
	}

	// Sort by level
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].Level < tiers[j].Level
	})

	return tiers
}

func (s *CommissionTierService) CreateTier(tier *CommissionTier) *CommissionTier {
	s.mu.Lock()
	defer s.mu.Unlock()

	tier.ID = s.nextTierID
	s.tiers[tier.ID] = tier
	s.nextTierID++

	return tier
}

func (s *CommissionTierService) UpdateTier(id int64, updates map[string]interface{}) (*CommissionTier, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tier, exists := s.tiers[id]
	if !exists {
		return nil, fmt.Errorf("tier not found")
	}

	if val, ok := updates["volumeThreshold"].(float64); ok {
		tier.VolumeThreshold = val
	}
	if val, ok := updates["commissionRate"].(float64); ok {
		tier.CommissionRate = val
	}
	if val, ok := updates["rebateRate"].(float64); ok {
		tier.RebateRate = val
	}
	if val, ok := updates["spreadMarkup"].(float64); ok {
		tier.SpreadMarkup = val
	}

	return tier, nil
}

func (s *CommissionTierService) ListClients(sortBy string, sortOrder string, limit, offset int) ([]*ClientCommission, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*ClientCommission, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}

	// Sort
	switch sortBy {
	case "volume":
		sort.Slice(clients, func(i, j int) bool {
			if sortOrder == "desc" {
				return clients[i].MonthlyVolume > clients[j].MonthlyVolume
			}
			return clients[i].MonthlyVolume < clients[j].MonthlyVolume
		})
	case "commission":
		sort.Slice(clients, func(i, j int) bool {
			if sortOrder == "desc" {
				return clients[i].TotalCommission > clients[j].TotalCommission
			}
			return clients[i].TotalCommission < clients[j].TotalCommission
		})
	case "netCost":
		sort.Slice(clients, func(i, j int) bool {
			if sortOrder == "desc" {
				return clients[i].NetCost > clients[j].NetCost
			}
			return clients[i].NetCost < clients[j].NetCost
		})
	default:
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].ClientID < clients[j].ClientID
		})
	}

	total := len(clients)
	if offset >= total {
		return []*ClientCommission{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return clients[offset:end], total
}

func (s *CommissionTierService) OverrideClientTier(clientID, tierID int64) (*ClientCommission, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	tier, exists := s.tiers[tierID]
	if !exists {
		return nil, fmt.Errorf("tier not found")
	}

	// Update old tier count
	oldTier := s.tiers[client.TierID]
	oldTier.ClientCount--

	// Update client tier
	client.TierID = tierID
	client.CurrentTier = tier.Name

	// Recalculate commission/rebate based on new tier
	client.TotalCommission = client.MonthlyVolume * tier.CommissionRate
	client.TotalRebate = client.MonthlyVolume * tier.RebateRate
	client.NetCost = client.TotalCommission - client.TotalRebate
	client.LastCalculated = time.Now()

	// Update new tier count
	tier.ClientCount++

	return client, nil
}

func (s *CommissionTierService) ListTransactions(filters map[string]string, limit, offset int) ([]*CommissionTransaction, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*CommissionTransaction
	for _, tx := range s.transactions {
		match := true

		if clientID, ok := filters["clientId"]; ok && clientID != "" {
			id, _ := strconv.ParseInt(clientID, 10, 64)
			if tx.ClientID != id {
				match = false
			}
		}
		if tier, ok := filters["tier"]; ok && tier != "" && tx.Tier != tier {
			match = false
		}
		if symbol, ok := filters["symbol"]; ok && symbol != "" && tx.Symbol != symbol {
			match = false
		}

		if match {
			filtered = append(filtered, tx)
		}
	}

	// Sort by timestamp desc
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	total := len(filtered)
	if offset >= total {
		return []*CommissionTransaction{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *CommissionTierService) GetStats() CommissionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := CommissionStats{
		TierDistribution: make(map[string]int),
		LastUpdated:      time.Now(),
	}

	for _, client := range s.clients {
		stats.TotalCommissions += client.TotalCommission
		stats.TotalRebates += client.TotalRebate
		stats.TotalVolume += client.MonthlyVolume
		stats.TierDistribution[client.CurrentTier]++

		if client.CurrentTier == "VIP" || client.CurrentTier == "Platinum" {
			stats.TopTierClients++
		}
	}

	stats.NetCommissions = stats.TotalCommissions - stats.TotalRebates

	if len(s.clients) > 0 {
		stats.AvgPerClient = stats.NetCommissions / float64(len(s.clients))
	}

	return stats
}

func (s *CommissionTierService) CalculateCommission(clientID int64) (*ClientCommission, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	return client, nil
}

// ============================================
// HTTP Handlers
// ============================================

type CommissionTierHandler struct {
	service     *CommissionTierService
	authService *auth.Service
}

func NewCommissionTierHandler(service *CommissionTierService, authService *auth.Service) *CommissionTierHandler {
	return &CommissionTierHandler{
		service:     service,
		authService: authService,
	}
}

// GET /admin/commissions/tiers
func (h *CommissionTierHandler) HandleListTiers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tiers := h.service.ListTiers()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
		"count": len(tiers),
	})
}

// POST /admin/commissions/tiers
func (h *CommissionTierHandler) HandleCreateTier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var tier CommissionTier
	if err := json.NewDecoder(r.Body).Decode(&tier); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created := h.service.CreateTier(&tier)
	json.NewEncoder(w).Encode(created)
}

// PUT /admin/commissions/tiers/:id
func (h *CommissionTierHandler) HandleUpdateTier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/commissions/tiers/"), "/")
	tierID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid tier ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tier, err := h.service.UpdateTier(tierID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(tier)
}

// GET /admin/commissions/clients
func (h *CommissionTierHandler) HandleListClients(w http.ResponseWriter, r *http.Request) {
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

	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(query.Get("offset"))

	clients, total := h.service.ListClients(sortBy, sortOrder, limit, offset)

	response := map[string]interface{}{
		"clients": clients,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}

	json.NewEncoder(w).Encode(response)
}

// PUT /admin/commissions/clients/:id/tier
func (h *CommissionTierHandler) HandleOverrideClientTier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/commissions/clients/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var req TierOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	client, err := h.service.OverrideClientTier(clientID, req.TierID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(client)
}

// GET /admin/commissions/transactions
func (h *CommissionTierHandler) HandleListTransactions(w http.ResponseWriter, r *http.Request) {
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
		"clientId": query.Get("clientId"),
		"tier":     query.Get("tier"),
		"symbol":   query.Get("symbol"),
	}

	transactions, total := h.service.ListTransactions(filters, limit, offset)

	response := map[string]interface{}{
		"transactions": transactions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	}

	json.NewEncoder(w).Encode(response)
}

// GET /admin/commissions/stats
func (h *CommissionTierHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// GET /admin/commissions/calculate/:clientId
func (h *CommissionTierHandler) HandleCalculateCommission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/commissions/calculate/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	client, err := h.service.CalculateCommission(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(client)
}
