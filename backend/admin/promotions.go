package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// Promotion represents a bonus or promotional offer
type Promotion struct {
	ID                   int64            `json:"id"`
	Name                 string           `json:"name"`
	Type                 string           `json:"type"` // welcome, deposit, cashback, referral, volume, no-deposit
	Description          string           `json:"description"`
	Terms                string           `json:"terms"`
	Status               string           `json:"status"` // active, paused, ended, scheduled
	StartDate            time.Time        `json:"startDate"`
	EndDate              time.Time        `json:"endDate"`
	Budget               float64          `json:"budget"`
	BudgetUsed           float64          `json:"budgetUsed"`
	MinDeposit           float64          `json:"minDeposit,omitempty"`
	MaxBonus             float64          `json:"maxBonus,omitempty"`
	BonusPercent         float64          `json:"bonusPercent,omitempty"` // For percentage-based bonuses
	BonusAmount          float64          `json:"bonusAmount,omitempty"`  // For fixed-amount bonuses
	EligibleAccountTypes []string         `json:"eligibleAccountTypes"`
	CountriesIncluded    []string         `json:"countriesIncluded,omitempty"`
	CountriesExcluded    []string         `json:"countriesExcluded,omitempty"`
	ClaimCount           int              `json:"claimCount"`
	TotalClaimed         float64          `json:"totalClaimed"`
	VolumeTiers          []VolumeTier     `json:"volumeTiers,omitempty"` // For volume-based bonuses
	CashbackPerLot       float64          `json:"cashbackPerLot,omitempty"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

// VolumeTier represents a volume-based bonus tier
type VolumeTier struct {
	MinLots int     `json:"minLots"`
	Bonus   float64 `json:"bonus"`
}

// PromotionClaim represents a client claiming a promotion
type PromotionClaim struct {
	ID           int64     `json:"id"`
	PromotionID  int64     `json:"promotionId"`
	ClientID     int64     `json:"clientId"`
	ClientName   string    `json:"clientName"`
	ClientEmail  string    `json:"clientEmail"`
	BonusAwarded float64   `json:"bonusAwarded"`
	ClaimedAt    time.Time `json:"claimedAt"`
	Status       string    `json:"status"` // pending, approved, credited, rejected
	Notes        string    `json:"notes,omitempty"`
}

// PromotionStore manages promotions and claims
type PromotionStore struct {
	promotions map[int64]*Promotion
	claims     map[int64]*PromotionClaim
	claimsByPromo map[int64][]*PromotionClaim
	mu         sync.RWMutex
	nextPromotionID int64
	nextClaimID     int64
}

// PromotionService provides business logic for promotion management
type PromotionService struct {
	store *PromotionStore
}

// NewPromotionService creates a new promotion service with default promotions
func NewPromotionService() *PromotionService {
	store := &PromotionStore{
		promotions:    make(map[int64]*Promotion),
		claims:        make(map[int64]*PromotionClaim),
		claimsByPromo: make(map[int64][]*PromotionClaim),
		nextPromotionID: 1,
		nextClaimID:     1,
	}

	now := time.Now()

	// Initialize with 6 default promotions
	defaultPromotions := []*Promotion{
		{
			ID:          1,
			Name:        "Welcome Bonus",
			Type:        "welcome",
			Description: "Get 100% bonus on your first deposit up to $500",
			Terms:       "New clients only. Minimum deposit $100. Bonus credited within 24 hours. Trading volume requirement: 20x bonus amount before withdrawal.",
			Status:      "active",
			StartDate:   now.Add(-30 * 24 * time.Hour),
			EndDate:     now.Add(365 * 24 * time.Hour),
			Budget:      50000,
			BudgetUsed:  12350,
			MinDeposit:  100,
			MaxBonus:    500,
			BonusPercent: 100,
			EligibleAccountTypes: []string{"Standard", "ECN", "VIP"},
			CountriesExcluded:    []string{"US", "CA"},
			ClaimCount:           47,
			TotalClaimed:         12350,
			CreatedAt:            now.Add(-30 * 24 * time.Hour),
			UpdatedAt:            now.Add(-2 * 24 * time.Hour),
		},
		{
			ID:          2,
			Name:        "Deposit Bonus 50%",
			Type:        "deposit",
			Description: "50% deposit bonus up to $5000 on all deposits",
			Terms:       "Available for all clients. No minimum deposit. Bonus valid for 90 days. Volume requirement: 15x bonus amount.",
			Status:      "active",
			StartDate:   now.Add(-15 * 24 * time.Hour),
			EndDate:     now.Add(75 * 24 * time.Hour),
			Budget:      100000,
			BudgetUsed:  34200,
			MinDeposit:  0,
			MaxBonus:    5000,
			BonusPercent: 50,
			EligibleAccountTypes: []string{"Standard", "ECN", "VIP", "Islamic"},
			ClaimCount:           89,
			TotalClaimed:         34200,
			CreatedAt:            now.Add(-15 * 24 * time.Hour),
			UpdatedAt:            now.Add(-1 * 24 * time.Hour),
		},
		{
			ID:          3,
			Name:        "Cashback Program",
			Type:        "cashback",
			Description: "$2 cashback per standard lot traded",
			Terms:       "Automatic cashback credited weekly. Available on all Forex pairs. No minimum volume. Cashback paid directly to trading account.",
			Status:      "active",
			StartDate:   now.Add(-60 * 24 * time.Hour),
			EndDate:     now.Add(300 * 24 * time.Hour),
			Budget:      200000,
			BudgetUsed:  45600,
			CashbackPerLot: 2.0,
			EligibleAccountTypes: []string{"Standard", "ECN", "VIP"},
			ClaimCount:           312,
			TotalClaimed:         45600,
			CreatedAt:            now.Add(-60 * 24 * time.Hour),
			UpdatedAt:            now.Add(-3 * time.Hour),
		},
		{
			ID:          4,
			Name:        "Referral Bonus",
			Type:        "referral",
			Description: "$50 for each friend you refer who makes a deposit",
			Terms:       "Friend must be a new client and deposit minimum $200. Bonus credited after friend completes 5 trades. No limit on referrals.",
			Status:      "active",
			StartDate:   now.Add(-90 * 24 * time.Hour),
			EndDate:     now.Add(270 * 24 * time.Hour),
			Budget:      25000,
			BudgetUsed:  3850,
			MinDeposit:  200,
			BonusAmount: 50,
			EligibleAccountTypes: []string{"Standard", "ECN", "VIP", "Islamic"},
			ClaimCount:           77,
			TotalClaimed:         3850,
			CreatedAt:            now.Add(-90 * 24 * time.Hour),
			UpdatedAt:            now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:          5,
			Name:        "Volume Bonus",
			Type:        "volume",
			Description: "Trade more, earn more! Tiered bonuses based on monthly volume",
			Terms:       "Volume calculated monthly. Bonus paid at end of month. Tiers: 10-49 lots = $100, 50-99 lots = $500, 100+ lots = $1500.",
			Status:      "active",
			StartDate:   now.Add(-45 * 24 * time.Hour),
			EndDate:     now.Add(315 * 24 * time.Hour),
			Budget:      75000,
			BudgetUsed:  18900,
			EligibleAccountTypes: []string{"Standard", "ECN", "VIP"},
			VolumeTiers: []VolumeTier{
				{MinLots: 10, Bonus: 100},
				{MinLots: 50, Bonus: 500},
				{MinLots: 100, Bonus: 1500},
			},
			ClaimCount:   143,
			TotalClaimed: 18900,
			CreatedAt:    now.Add(-45 * 24 * time.Hour),
			UpdatedAt:    now.Add(-1 * 24 * time.Hour),
		},
		{
			ID:          6,
			Name:        "No-Deposit Bonus",
			Type:        "no-deposit",
			Description: "$25 free bonus - no deposit required!",
			Terms:       "New clients only. KYC verification required. Cannot be withdrawn. Use for trading only. Profits from bonus trading can be withdrawn after 5 lots traded.",
			Status:      "active",
			StartDate:   now.Add(-20 * 24 * time.Hour),
			EndDate:     now.Add(40 * 24 * time.Hour),
			Budget:      10000,
			BudgetUsed:  2575,
			BonusAmount: 25,
			EligibleAccountTypes: []string{"Demo", "Micro", "Standard"},
			CountriesExcluded:    []string{"US", "CA", "AU"},
			ClaimCount:           103,
			TotalClaimed:         2575,
			CreatedAt:            now.Add(-20 * 24 * time.Hour),
			UpdatedAt:            now.Add(-6 * time.Hour),
		},
	}

	store.mu.Lock()
	for _, promo := range defaultPromotions {
		store.promotions[promo.ID] = promo
		store.claimsByPromo[promo.ID] = []*PromotionClaim{}
		if promo.ID >= store.nextPromotionID {
			store.nextPromotionID = promo.ID + 1
		}
	}

	// Generate some mock claims
	claimClientData := []struct {
		name  string
		email string
	}{
		{"John Smith", "john.smith@email.com"},
		{"Sarah Johnson", "sarah.j@email.com"},
		{"Michael Chen", "m.chen@email.com"},
		{"Emma Wilson", "emma.wilson@email.com"},
		{"David Martinez", "d.martinez@email.com"},
		{"Lisa Anderson", "lisa.a@email.com"},
		{"James Taylor", "james.t@email.com"},
		{"Maria Garcia", "maria.g@email.com"},
		{"Robert Brown", "robert.b@email.com"},
		{"Jennifer Lee", "jennifer.lee@email.com"},
	}

	claimID := int64(1)
	for promoID := int64(1); promoID <= 6; promoID++ {
		numClaims := 3 + (promoID % 4) // 3-6 claims per promotion
		for i := 0; i < int(numClaims); i++ {
			clientData := claimClientData[i%len(claimClientData)]
			bonusAmount := 50.0 + float64(i*20)

			if promoID == 3 { // Cashback
				bonusAmount = 2.0 * float64(10+i*5)
			} else if promoID == 4 { // Referral
				bonusAmount = 50.0
			} else if promoID == 6 { // No-deposit
				bonusAmount = 25.0
			}

			claim := &PromotionClaim{
				ID:           claimID,
				PromotionID:  promoID,
				ClientID:     int64(1000 + i),
				ClientName:   clientData.name,
				ClientEmail:  clientData.email,
				BonusAwarded: bonusAmount,
				ClaimedAt:    now.Add(time.Duration(-i*24-int(promoID)*12) * time.Hour),
				Status:       "credited",
			}
			store.claims[claimID] = claim
			store.claimsByPromo[promoID] = append(store.claimsByPromo[promoID], claim)
			claimID++
		}
	}
	store.nextClaimID = claimID
	store.mu.Unlock()

	return &PromotionService{store: store}
}

// ListPromotions returns all promotions with optional status filter
func (ps *PromotionService) ListPromotions(statusFilter string) []*Promotion {
	ps.store.mu.RLock()
	defer ps.store.mu.RUnlock()

	promotions := make([]*Promotion, 0, len(ps.store.promotions))
	for _, promo := range ps.store.promotions {
		if statusFilter != "" && promo.Status != statusFilter {
			continue
		}
		promotions = append(promotions, promo)
	}

	return promotions
}

// GetPromotion returns a specific promotion by ID
func (ps *PromotionService) GetPromotion(id int64) (*Promotion, error) {
	ps.store.mu.RLock()
	defer ps.store.mu.RUnlock()

	promo, exists := ps.store.promotions[id]
	if !exists {
		return nil, fmt.Errorf("promotion not found")
	}

	return promo, nil
}

// CreatePromotion creates a new promotion
func (ps *PromotionService) CreatePromotion(promo *Promotion) (*Promotion, error) {
	ps.store.mu.Lock()
	defer ps.store.mu.Unlock()

	// Validate required fields
	if promo.Name == "" {
		return nil, fmt.Errorf("promotion name is required")
	}
	if promo.Type == "" {
		return nil, fmt.Errorf("promotion type is required")
	}
	if promo.Budget <= 0 {
		return nil, fmt.Errorf("budget must be greater than 0")
	}

	// Assign new ID and timestamps
	promo.ID = ps.store.nextPromotionID
	ps.store.nextPromotionID++
	promo.CreatedAt = time.Now()
	promo.UpdatedAt = time.Now()
	promo.BudgetUsed = 0
	promo.ClaimCount = 0
	promo.TotalClaimed = 0

	if promo.Status == "" {
		promo.Status = "scheduled"
	}

	ps.store.promotions[promo.ID] = promo
	ps.store.claimsByPromo[promo.ID] = []*PromotionClaim{}

	return promo, nil
}

// UpdatePromotion updates an existing promotion
func (ps *PromotionService) UpdatePromotion(id int64, updates *Promotion) (*Promotion, error) {
	ps.store.mu.Lock()
	defer ps.store.mu.Unlock()

	promo, exists := ps.store.promotions[id]
	if !exists {
		return nil, fmt.Errorf("promotion not found")
	}

	// Update fields
	if updates.Name != "" {
		promo.Name = updates.Name
	}
	if updates.Description != "" {
		promo.Description = updates.Description
	}
	if updates.Terms != "" {
		promo.Terms = updates.Terms
	}
	if updates.Budget > 0 {
		promo.Budget = updates.Budget
	}
	if updates.MinDeposit >= 0 {
		promo.MinDeposit = updates.MinDeposit
	}
	if updates.MaxBonus >= 0 {
		promo.MaxBonus = updates.MaxBonus
	}
	if updates.BonusPercent >= 0 {
		promo.BonusPercent = updates.BonusPercent
	}
	if updates.BonusAmount >= 0 {
		promo.BonusAmount = updates.BonusAmount
	}
	if updates.CashbackPerLot >= 0 {
		promo.CashbackPerLot = updates.CashbackPerLot
	}
	if len(updates.EligibleAccountTypes) > 0 {
		promo.EligibleAccountTypes = updates.EligibleAccountTypes
	}
	if len(updates.VolumeTiers) > 0 {
		promo.VolumeTiers = updates.VolumeTiers
	}
	if !updates.StartDate.IsZero() {
		promo.StartDate = updates.StartDate
	}
	if !updates.EndDate.IsZero() {
		promo.EndDate = updates.EndDate
	}

	promo.UpdatedAt = time.Now()

	return promo, nil
}

// UpdatePromotionStatus updates the status of a promotion
func (ps *PromotionService) UpdatePromotionStatus(id int64, status string) (*Promotion, error) {
	ps.store.mu.Lock()
	defer ps.store.mu.Unlock()

	promo, exists := ps.store.promotions[id]
	if !exists {
		return nil, fmt.Errorf("promotion not found")
	}

	validStatuses := map[string]bool{
		"active":    true,
		"paused":    true,
		"ended":     true,
		"scheduled": true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: must be active, paused, ended, or scheduled")
	}

	promo.Status = status
	promo.UpdatedAt = time.Now()

	return promo, nil
}

// GetPromotionClaims returns all claims for a specific promotion
func (ps *PromotionService) GetPromotionClaims(promotionID int64) ([]*PromotionClaim, error) {
	ps.store.mu.RLock()
	defer ps.store.mu.RUnlock()

	if _, exists := ps.store.promotions[promotionID]; !exists {
		return nil, fmt.Errorf("promotion not found")
	}

	claims := ps.store.claimsByPromo[promotionID]
	return claims, nil
}

// ClaimPromotion manually awards a bonus to a client
func (ps *PromotionService) ClaimPromotion(promotionID, clientID int64, clientName, clientEmail string, bonusAmount float64) (*PromotionClaim, error) {
	ps.store.mu.Lock()
	defer ps.store.mu.Unlock()

	promo, exists := ps.store.promotions[promotionID]
	if !exists {
		return nil, fmt.Errorf("promotion not found")
	}

	if promo.Status != "active" {
		return nil, fmt.Errorf("promotion is not active")
	}

	// Check budget availability
	if promo.BudgetUsed+bonusAmount > promo.Budget {
		return nil, fmt.Errorf("insufficient promotion budget")
	}

	// Create claim
	claim := &PromotionClaim{
		ID:           ps.store.nextClaimID,
		PromotionID:  promotionID,
		ClientID:     clientID,
		ClientName:   clientName,
		ClientEmail:  clientEmail,
		BonusAwarded: bonusAmount,
		ClaimedAt:    time.Now(),
		Status:       "credited",
	}

	ps.store.nextClaimID++
	ps.store.claims[claim.ID] = claim
	ps.store.claimsByPromo[promotionID] = append(ps.store.claimsByPromo[promotionID], claim)

	// Update promotion stats
	promo.ClaimCount++
	promo.TotalClaimed += bonusAmount
	promo.BudgetUsed += bonusAmount
	promo.UpdatedAt = time.Now()

	return claim, nil
}

// GetPromotionStats returns summary statistics
func (ps *PromotionService) GetPromotionStats() map[string]interface{} {
	ps.store.mu.RLock()
	defer ps.store.mu.RUnlock()

	activeCount := 0
	totalBudget := 0.0
	totalBudgetUsed := 0.0
	claimsToday := 0

	today := time.Now().Truncate(24 * time.Hour)

	for _, promo := range ps.store.promotions {
		if promo.Status == "active" {
			activeCount++
		}
		totalBudget += promo.Budget
		totalBudgetUsed += promo.BudgetUsed
	}

	for _, claim := range ps.store.claims {
		if claim.ClaimedAt.After(today) {
			claimsToday++
		}
	}

	return map[string]interface{}{
		"activePromotions": activeCount,
		"totalBudget":      totalBudget,
		"budgetUsed":       totalBudgetUsed,
		"budgetRemaining":  totalBudget - totalBudgetUsed,
		"budgetUsedPct":    (totalBudgetUsed / totalBudget) * 100,
		"claimsToday":      claimsToday,
		"totalClaims":      len(ps.store.claims),
	}
}

// PromotionHandler handles HTTP requests for promotion management
type PromotionHandler struct {
	service     *PromotionService
	authService *auth.Service
}

// NewPromotionHandler creates a new promotion HTTP handler
func NewPromotionHandler(service *PromotionService, authService *auth.Service) *PromotionHandler {
	return &PromotionHandler{
		service:     service,
		authService: authService,
	}
}

// ListPromotions handles GET /admin/promotions
func (ph *PromotionHandler) ListPromotions(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	statusFilter := r.URL.Query().Get("status")
	promotions := ph.service.ListPromotions(statusFilter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    promotions,
		"count":   len(promotions),
	})
}

// GetPromotion handles GET /admin/promotions/:id
func (ph *PromotionHandler) GetPromotion(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	promotion, err := ph.service.GetPromotion(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get claim history
	claims, _ := ph.service.GetPromotionClaims(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    promotion,
		"claims":  claims,
	})
}

// CreatePromotion handles POST /admin/promotions
func (ph *PromotionHandler) CreatePromotion(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var promo Promotion
	if err := json.NewDecoder(r.Body).Decode(&promo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := ph.service.CreatePromotion(&promo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    created,
		"message": "Promotion created successfully",
	})
}

// UpdatePromotion handles PUT /admin/promotions/:id
func (ph *PromotionHandler) UpdatePromotion(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates Promotion
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := ph.service.UpdatePromotion(id, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    updated,
		"message": "Promotion updated successfully",
	})
}

// UpdatePromotionStatus handles PUT /admin/promotions/:id/status
func (ph *PromotionHandler) UpdatePromotionStatus(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := ph.service.UpdatePromotionStatus(id, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    updated,
		"message": fmt.Sprintf("Promotion status updated to %s", req.Status),
	})
}

// GetPromotionClaims handles GET /admin/promotions/:id/claims
func (ph *PromotionHandler) GetPromotionClaims(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	claims, err := ph.service.GetPromotionClaims(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    claims,
		"count":   len(claims),
	})
}

// ClaimPromotion handles POST /admin/promotions/:id/claim
func (ph *PromotionHandler) ClaimPromotion(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req struct {
		ClientID    int64   `json:"clientId"`
		ClientName  string  `json:"clientName"`
		ClientEmail string  `json:"clientEmail"`
		BonusAmount float64 `json:"bonusAmount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claim, err := ph.service.ClaimPromotion(id, req.ClientID, req.ClientName, req.ClientEmail, req.BonusAmount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    claim,
		"message": "Bonus awarded successfully",
	})
}

// GetPromotionStats handles GET /admin/promotions/stats
func (ph *PromotionHandler) GetPromotionStats(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := ph.authService.ValidateToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := ph.service.GetPromotionStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}
