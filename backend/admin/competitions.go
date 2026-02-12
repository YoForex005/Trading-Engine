package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

// Competition represents a trading competition/contest
type Competition struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	Status               string                 `json:"status"` // active, upcoming, ended
	StartDate            time.Time              `json:"startDate"`
	EndDate              time.Time              `json:"endDate"`
	EntryFee             float64                `json:"entryFee"`
	PrizePool            float64                `json:"prizePool"`
	MaxParticipants      int                    `json:"maxParticipants"`
	CurrentParticipants  int                    `json:"currentParticipants"`
	EligibleAccountTypes []string               `json:"eligibleAccountTypes"`
	Rules                string                 `json:"rules"`
	PrizeDistribution    map[string]interface{} `json:"prizeDistribution"`
	CreatedAt            time.Time              `json:"createdAt"`
	CreatedBy            string                 `json:"createdBy,omitempty"`
}

// Participant represents a competition participant
type Participant struct {
	ID            string    `json:"id"`
	CompetitionID string    `json:"competitionId"`
	AccountID     string    `json:"accountId"`
	TraderName    string    `json:"traderName"`
	ProfitPct     float64   `json:"profitPct"`
	EquityGrowth  float64   `json:"equityGrowth"`
	TradesCount   int       `json:"tradesCount"`
	WinRate       float64   `json:"winRate"`
	Rank          int       `json:"rank,omitempty"`
	JoinedAt      time.Time `json:"joinedAt"`
	Status        string    `json:"status"` // active, disqualified
}

// CompetitionDetail includes competition with participants
type CompetitionDetail struct {
	Competition
	Participants []Participant `json:"participants"`
}

// ============================================
// In-Memory Store
// ============================================

type CompetitionStore struct {
	mu           sync.RWMutex
	competitions map[string]*Competition
	participants map[string][]Participant // competitionID -> participants
}

func NewCompetitionStore() *CompetitionStore {
	store := &CompetitionStore{
		competitions: make(map[string]*Competition),
		participants: make(map[string][]Participant),
	}

	now := time.Now()

	// ============================================
	// ACTIVE COMPETITIONS (3)
	// ============================================

	// Active Competition 1: Monthly Forex Challenge
	comp1 := &Competition{
		ID:                  "comp-001",
		Name:                "Monthly Forex Challenge - February 2026",
		Description:         "Compete with traders worldwide in this monthly forex trading contest. Highest % profit wins!",
		Status:              "active",
		StartDate:           now.AddDate(0, 0, -10),
		EndDate:             now.AddDate(0, 0, 18),
		EntryFee:            100.0,
		PrizePool:           12000.0,
		MaxParticipants:     150,
		CurrentParticipants: 87,
		EligibleAccountTypes: []string{"Standard", "ECN", "Pro"},
		Rules: "Minimum 10 trades required. No hedging. No EA allowed. Forex pairs only.",
		PrizeDistribution: map[string]interface{}{
			"1st": 5000.0,
			"2nd": 3000.0,
			"3rd": 2000.0,
			"4th-10th": 285.71, // 2000 / 7
		},
		CreatedAt: now.AddDate(0, 0, -15),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp1.ID] = comp1
	store.participants[comp1.ID] = generateParticipants(comp1.ID, 87, now.AddDate(0, 0, -10), true)

	// Active Competition 2: Crypto Trading Blitz
	comp2 := &Competition{
		ID:                  "comp-002",
		Name:                "Crypto Trading Blitz - Weekend Warriors",
		Description:         "48-hour intense crypto trading competition. Fast-paced action for thrill seekers!",
		Status:              "active",
		StartDate:           now.AddDate(0, 0, -1),
		EndDate:             now.AddDate(0, 0, 1),
		EntryFee:            50.0,
		PrizePool:           3000.0,
		MaxParticipants:     80,
		CurrentParticipants: 62,
		EligibleAccountTypes: []string{"Standard", "Crypto"},
		Rules: "Crypto pairs only (BTC, ETH, XRP, BNB, SOL). Minimum 20 trades. No stop loss hunting.",
		PrizeDistribution: map[string]interface{}{
			"1st": 1500.0,
			"2nd": 900.0,
			"3rd": 600.0,
		},
		CreatedAt: now.AddDate(0, 0, -5),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp2.ID] = comp2
	store.participants[comp2.ID] = generateParticipants(comp2.ID, 62, now.AddDate(0, 0, -1), true)

	// Active Competition 3: Indices Mastery Contest
	comp3 := &Competition{
		ID:                  "comp-003",
		Name:                "Indices Mastery Contest - Q1 2026",
		Description:         "Quarterly indices trading championship. Prove your mastery of global markets!",
		Status:              "active",
		StartDate:           now.AddDate(0, -1, 5),
		EndDate:             now.AddDate(0, 2, -5),
		EntryFee:            200.0,
		PrizePool:           25000.0,
		MaxParticipants:     200,
		CurrentParticipants: 143,
		EligibleAccountTypes: []string{"ECN", "Pro", "Institutional"},
		Rules: "Indices only (SPX500, NAS100, DE30, UK100, JP225). Minimum 30 trades. Max 5% risk per trade.",
		PrizeDistribution: map[string]interface{}{
			"1st":     10000.0,
			"2nd":     6000.0,
			"3rd":     4000.0,
			"4th":     2000.0,
			"5th":     1500.0,
			"6th-10th": 300.0,
		},
		CreatedAt: now.AddDate(0, -1, 0),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp3.ID] = comp3
	store.participants[comp3.ID] = generateParticipants(comp3.ID, 143, now.AddDate(0, -1, 5), true)

	// ============================================
	// UPCOMING COMPETITIONS (2)
	// ============================================

	// Upcoming Competition 1: Scalping Championship
	comp4 := &Competition{
		ID:                  "comp-004",
		Name:                "Scalping Championship - March 2026",
		Description:         "Speed and precision matter! Short-term trading contest for scalping experts.",
		Status:              "upcoming",
		StartDate:           now.AddDate(0, 0, 15),
		EndDate:             now.AddDate(0, 0, 45),
		EntryFee:            150.0,
		PrizePool:           18000.0,
		MaxParticipants:     120,
		CurrentParticipants: 34,
		EligibleAccountTypes: []string{"ECN", "Raw Spread"},
		Rules: "Average trade duration must be under 5 minutes. Minimum 100 trades. All instruments allowed.",
		PrizeDistribution: map[string]interface{}{
			"1st":     7000.0,
			"2nd":     4500.0,
			"3rd":     3000.0,
			"4th-10th": 500.0,
		},
		CreatedAt: now.AddDate(0, 0, -3),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp4.ID] = comp4
	store.participants[comp4.ID] = generateParticipants(comp4.ID, 34, now.AddDate(0, 0, -2), false)

	// Upcoming Competition 2: Metals Trading Cup
	comp5 := &Competition{
		ID:                  "comp-005",
		Name:                "Metals Trading Cup - Gold & Silver Challenge",
		Description:         "Focus on precious metals trading. XAU and XAG pairs only!",
		Status:              "upcoming",
		StartDate:           now.AddDate(0, 0, 20),
		EndDate:             now.AddDate(0, 1, 20),
		EntryFee:            75.0,
		PrizePool:           6000.0,
		MaxParticipants:     100,
		CurrentParticipants: 18,
		EligibleAccountTypes: []string{"Standard", "ECN", "Pro"},
		Rules: "Gold and Silver pairs only (XAUUSD, XAGUSD, XAUEUR, etc.). Minimum 15 trades.",
		PrizeDistribution: map[string]interface{}{
			"1st": 3000.0,
			"2nd": 1800.0,
			"3rd": 1200.0,
		},
		CreatedAt: now.AddDate(0, 0, -1),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp5.ID] = comp5
	store.participants[comp5.ID] = generateParticipants(comp5.ID, 18, now.AddDate(0, 0, -1), false)

	// ============================================
	// ENDED COMPETITIONS (3)
	// ============================================

	// Ended Competition 1: January Forex Blast
	comp6 := &Competition{
		ID:                  "comp-006",
		Name:                "January Forex Blast 2026",
		Description:         "Monthly forex competition - January edition. Congratulations to all participants!",
		Status:              "ended",
		StartDate:           now.AddDate(0, -1, -10),
		EndDate:             now.AddDate(0, 0, -3),
		EntryFee:            100.0,
		PrizePool:           15000.0,
		MaxParticipants:     150,
		CurrentParticipants: 120,
		EligibleAccountTypes: []string{"Standard", "ECN", "Pro"},
		Rules: "Minimum 10 trades required. No hedging. Forex pairs only.",
		PrizeDistribution: map[string]interface{}{
			"1st": 6000.0,
			"2nd": 3600.0,
			"3rd": 2400.0,
			"4th-10th": 428.57,
		},
		CreatedAt: now.AddDate(0, -2, 0),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp6.ID] = comp6
	store.participants[comp6.ID] = generateParticipants(comp6.ID, 120, now.AddDate(0, -1, -10), true)

	// Ended Competition 2: New Year Trading Championship
	comp7 := &Competition{
		ID:                  "comp-007",
		Name:                "New Year Trading Championship 2026",
		Description:         "Start the year strong! All instruments allowed in this mega contest.",
		Status:              "ended",
		StartDate:           time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		EndDate:             time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
		EntryFee:            250.0,
		PrizePool:           30000.0,
		MaxParticipants:     200,
		CurrentParticipants: 156,
		EligibleAccountTypes: []string{"Standard", "ECN", "Pro", "Institutional"},
		Rules: "All instruments allowed. Minimum 25 trades. Max 10% risk per trade.",
		PrizeDistribution: map[string]interface{}{
			"1st":     12000.0,
			"2nd":     7000.0,
			"3rd":     5000.0,
			"4th":     2500.0,
			"5th":     1500.0,
			"6th-10th": 400.0,
		},
		CreatedAt: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp7.ID] = comp7
	store.participants[comp7.ID] = generateParticipants(comp7.ID, 156, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), true)

	// Ended Competition 3: Swing Trading Marathon
	comp8 := &Competition{
		ID:                  "comp-008",
		Name:                "Swing Trading Marathon - December 2025",
		Description:         "Long-term strategy contest. Hold positions for 24+ hours to qualify.",
		Status:              "ended",
		StartDate:           time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		EndDate:             time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
		EntryFee:            120.0,
		PrizePool:           10000.0,
		MaxParticipants:     100,
		CurrentParticipants: 73,
		EligibleAccountTypes: []string{"Standard", "ECN"},
		Rules: "Minimum hold time 24 hours per trade. Minimum 8 trades. Forex and metals only.",
		PrizeDistribution: map[string]interface{}{
			"1st": 4500.0,
			"2nd": 2700.0,
			"3rd": 1800.0,
			"4th-5th": 500.0,
		},
		CreatedAt: time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC),
		CreatedBy: "admin@rtx5.com",
	}
	store.competitions[comp8.ID] = comp8
	store.participants[comp8.ID] = generateParticipants(comp8.ID, 73, time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC), true)

	log.Printf("[Competitions] Initialized %d competitions (3 active, 2 upcoming, 3 ended) with 593 total participants", len(store.competitions))

	return store
}

// generateParticipants creates realistic mock participants for a competition
func generateParticipants(competitionID string, count int, startDate time.Time, hasMetrics bool) []Participant {
	participants := make([]Participant, count)

	traderNames := []string{
		"ForexKing", "CryptoQueen", "IndexMaster", "ScalpPro", "SwingTrader",
		"DayTradingBoss", "PipHunter", "BullishBear", "BearishBull", "TrendFollower",
		"BreakoutSpecialist", "MomentumTrader", "ValueInvestor", "TechAnalyst", "FundamentalGuru",
		"RiskMaster", "ProfitSeeker", "MarketWizard", "ChartWhisperer", "OrderFlowPro",
		"VolumeAnalyst", "PriceActionKing", "PatternSpotter", "SignalProvider", "AlgoTrader",
		"QuantTrader", "StatArb", "GridMaster", "MartingalePro", "HedgeFundWannabe",
		"RetailWarrior", "InstitutionalSlayer", "LiquidityProvider", "MarketMaker", "Scalper2000",
		"DayTrader99", "SwingKing", "PositionTrader", "NewsTrader", "EconomicCalendarPro",
		"FibonacciMaster", "ElliottWaveExpert", "IchimokuTrader", "BollingerBandPro", "MACDKing",
		"RSISpecialist", "StochasticMaster", "ATRPro", "SupportResistanceGuru", "TrendlineArtist",
	}

	for i := 0; i < count; i++ {
		participant := Participant{
			ID:            fmt.Sprintf("part-%s-%03d", competitionID, i+1),
			CompetitionID: competitionID,
			AccountID:     fmt.Sprintf("ACC%06d", 100000+i),
			TraderName:    fmt.Sprintf("%s%d", traderNames[i%len(traderNames)], (i/len(traderNames))+1),
			JoinedAt:      startDate.Add(time.Duration(i*3) * time.Hour),
			Status:        "active",
		}

		if hasMetrics {
			// Generate realistic trading metrics
			// Top performers (first 10%)
			if i < count/10 {
				participant.ProfitPct = 50.0 + float64(10-i)*5.0      // 100% to 55%
				participant.EquityGrowth = 50.0 + float64(10-i)*5.0
				participant.TradesCount = 80 + i*2
				participant.WinRate = 70.0 + float64(10-i)
			} else if i < count/3 { // Good performers (next 23%)
				participant.ProfitPct = 20.0 + float64(count/3-i)*0.5
				participant.EquityGrowth = 20.0 + float64(count/3-i)*0.5
				participant.TradesCount = 50 + i
				participant.WinRate = 60.0 + float64(count/3-i)*0.2
			} else if i < count*2/3 { // Average performers (next 33%)
				participant.ProfitPct = 5.0 + float64(count*2/3-i)*0.2
				participant.EquityGrowth = 5.0 + float64(count*2/3-i)*0.2
				participant.TradesCount = 30 + i/2
				participant.WinRate = 50.0 + float64(count*2/3-i)*0.1
			} else { // Below average / losing (last 33%)
				participant.ProfitPct = -30.0 + float64(count-i)*0.5
				participant.EquityGrowth = -30.0 + float64(count-i)*0.5
				participant.TradesCount = 15 + i/3
				participant.WinRate = 35.0 + float64(count-i)*0.2
			}

			// Randomly disqualify 2-3% of participants
			if i%40 == 0 && i > 0 {
				participant.Status = "disqualified"
			}
		} else {
			// Upcoming competitions - no metrics yet
			participant.ProfitPct = 0.0
			participant.EquityGrowth = 0.0
			participant.TradesCount = 0
			participant.WinRate = 0.0
		}

		participants[i] = participant
	}

	return participants
}

// ============================================
// Handler
// ============================================

type CompetitionHandler struct {
	store       *CompetitionStore
	authService *auth.Service
}

func NewCompetitionHandler(store *CompetitionStore, authService *auth.Service) *CompetitionHandler {
	return &CompetitionHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// Handler Methods
// ============================================

// HandleList returns all competitions with optional filtering by status
// GET /admin/competitions?status=active
func (h *CompetitionHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	statusFilter := r.URL.Query().Get("status")

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var competitions []*Competition
	for _, comp := range h.store.competitions {
		if statusFilter == "" || comp.Status == statusFilter {
			// Add current participant count from participants map
			comp.CurrentParticipants = len(h.store.participants[comp.ID])
			competitions = append(competitions, comp)
		}
	}

	// Sort by start date descending (most recent first)
	sort.Slice(competitions, func(i, j int) bool {
		return competitions[i].StartDate.After(competitions[j].StartDate)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"competitions": competitions,
		"total":        len(competitions),
	})
}

// HandleCreate creates a new competition
// POST /admin/competitions
func (h *CompetitionHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var comp Competition
	if err := json.NewDecoder(r.Body).Decode(&comp); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate ID
	h.store.mu.Lock()
	comp.ID = fmt.Sprintf("comp-%03d", len(h.store.competitions)+1)
	comp.CreatedAt = time.Now()
	comp.CurrentParticipants = 0

	// Determine status based on dates
	now := time.Now()
	if now.Before(comp.StartDate) {
		comp.Status = "upcoming"
	} else if now.After(comp.EndDate) {
		comp.Status = "ended"
	} else {
		comp.Status = "active"
	}

	h.store.competitions[comp.ID] = &comp
	h.store.participants[comp.ID] = []Participant{}
	h.store.mu.Unlock()

	log.Printf("[Competitions] Created new competition: %s (%s)", comp.Name, comp.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comp)
}

// HandleGetDetail returns competition detail with participants
// GET /admin/competitions/:id
func (h *CompetitionHandler) HandleGetDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/admin/competitions/")
	if id == "" || id == r.URL.Path {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}

	h.store.mu.RLock()
	comp, exists := h.store.competitions[id]
	if !exists {
		h.store.mu.RUnlock()
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}

	participants := h.store.participants[id]

	// Calculate ranks
	rankedParticipants := make([]Participant, len(participants))
	copy(rankedParticipants, participants)

	// Sort by profit percentage descending
	sort.Slice(rankedParticipants, func(i, j int) bool {
		return rankedParticipants[i].ProfitPct > rankedParticipants[j].ProfitPct
	})

	// Assign ranks
	for i := range rankedParticipants {
		rankedParticipants[i].Rank = i + 1
	}

	h.store.mu.RUnlock()

	detail := CompetitionDetail{
		Competition:  *comp,
		Participants: rankedParticipants,
	}
	detail.CurrentParticipants = len(rankedParticipants)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// HandleUpdate updates an existing competition
// PUT /admin/competitions/:id
func (h *CompetitionHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/admin/competitions/")
	if id == "" || id == r.URL.Path {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}

	var updates Competition
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	comp, exists := h.store.competitions[id]
	if !exists {
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}

	// Update fields
	if updates.Name != "" {
		comp.Name = updates.Name
	}
	if updates.Description != "" {
		comp.Description = updates.Description
	}
	if !updates.StartDate.IsZero() {
		comp.StartDate = updates.StartDate
	}
	if !updates.EndDate.IsZero() {
		comp.EndDate = updates.EndDate
	}
	if updates.EntryFee > 0 {
		comp.EntryFee = updates.EntryFee
	}
	if updates.PrizePool > 0 {
		comp.PrizePool = updates.PrizePool
	}
	if updates.MaxParticipants > 0 {
		comp.MaxParticipants = updates.MaxParticipants
	}
	if len(updates.EligibleAccountTypes) > 0 {
		comp.EligibleAccountTypes = updates.EligibleAccountTypes
	}
	if updates.Rules != "" {
		comp.Rules = updates.Rules
	}
	if updates.PrizeDistribution != nil {
		comp.PrizeDistribution = updates.PrizeDistribution
	}
	if updates.Status != "" {
		comp.Status = updates.Status
	}

	log.Printf("[Competitions] Updated competition: %s (%s)", comp.Name, comp.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comp)
}

// HandleDelete deletes a competition
// DELETE /admin/competitions/:id
func (h *CompetitionHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/admin/competitions/")
	if id == "" || id == r.URL.Path {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	comp, exists := h.store.competitions[id]
	if !exists {
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}

	delete(h.store.competitions, id)
	delete(h.store.participants, id)

	log.Printf("[Competitions] Deleted competition: %s (%s)", comp.Name, comp.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Competition deleted successfully",
		"id":      id,
	})
}

// HandleGetLeaderboard returns ranked participants for a competition
// GET /admin/competitions/:id/leaderboard
func (h *CompetitionHandler) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path - handle /admin/competitions/:id/leaderboard
	path := strings.TrimPrefix(r.URL.Path, "/admin/competitions/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	id := parts[0]

	h.store.mu.RLock()
	comp, exists := h.store.competitions[id]
	if !exists {
		h.store.mu.RUnlock()
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}

	participants := h.store.participants[id]

	// Sort by profit percentage descending
	rankedParticipants := make([]Participant, len(participants))
	copy(rankedParticipants, participants)
	sort.Slice(rankedParticipants, func(i, j int) bool {
		// Active participants first, then by profit
		if rankedParticipants[i].Status != rankedParticipants[j].Status {
			return rankedParticipants[i].Status == "active"
		}
		return rankedParticipants[i].ProfitPct > rankedParticipants[j].ProfitPct
	})

	// Assign ranks (only to active participants)
	rank := 1
	for i := range rankedParticipants {
		if rankedParticipants[i].Status == "active" {
			rankedParticipants[i].Rank = rank
			rank++
		} else {
			rankedParticipants[i].Rank = 0 // Disqualified participants have no rank
		}
	}

	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"competitionId":   id,
		"competitionName": comp.Name,
		"leaderboard":     rankedParticipants,
		"totalActive":     rank - 1,
		"totalDisqualified": len(participants) - (rank - 1),
	})
}

// HandleDisqualify disqualifies a participant from a competition
// POST /admin/competitions/:id/disqualify
func (h *CompetitionHandler) HandleDisqualify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/competitions/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Competition ID required", http.StatusBadRequest)
		return
	}
	competitionID := parts[0]

	var req struct {
		ParticipantID string `json:"participantId"`
		Reason        string `json:"reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ParticipantID == "" {
		http.Error(w, "Participant ID required", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	_, exists := h.store.competitions[competitionID]
	if !exists {
		http.Error(w, "Competition not found", http.StatusNotFound)
		return
	}

	participants := h.store.participants[competitionID]
	found := false
	for i := range participants {
		if participants[i].ID == req.ParticipantID {
			participants[i].Status = "disqualified"
			found = true
			log.Printf("[Competitions] Disqualified participant %s from competition %s. Reason: %s",
				participants[i].TraderName, competitionID, req.Reason)
			break
		}
	}

	if !found {
		http.Error(w, "Participant not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Participant disqualified successfully",
		"participantId": req.ParticipantID,
		"reason":        req.Reason,
	})
}
