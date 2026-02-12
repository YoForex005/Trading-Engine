//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Affiliate represents an Introducing Broker or Affiliate Partner
type Affiliate struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Status          string    `json:"status"` // ACTIVE, SUSPENDED, PENDING
	Tier            string    `json:"tier"`   // Bronze, Silver, Gold, Platinum
	ReferralCode    string    `json:"referralCode"`
	ReferredClients []string  `json:"referredClients"` // Account numbers
	TotalVolume     float64   `json:"totalVolume"`     // Total trading volume in USD
	CommissionEarned float64  `json:"commissionEarned"`
	PayoutStatus    string    `json:"payoutStatus"` // PENDING, PAID
	CreatedAt       time.Time `json:"createdAt"`
}

// AffiliateCommissionTier defines commission rates by volume
type AffiliateCommissionTier struct {
	Name          string  `json:"name"`
	MinVolume     float64 `json:"minVolume"`     // Minimum volume in USD
	MaxVolume     float64 `json:"maxVolume"`     // Maximum volume in USD (0 = unlimited)
	PipCommission float64 `json:"pipCommission"` // Commission per pip per lot
}

// Payout represents a commission payout request
type Payout struct {
	ID          int64     `json:"id"`
	AffiliateID int64     `json:"affiliateId"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"` // PENDING, PROCESSING, COMPLETED, REJECTED
	RequestedAt time.Time `json:"requestedAt"`
	ProcessedAt time.Time `json:"processedAt,omitempty"`
}

// CommissionRecord represents a single commission entry
type CommissionRecord struct {
	ID          int64     `json:"id"`
	AffiliateID int64     `json:"affiliateId"`
	ClientID    string    `json:"clientId"` // Account number
	Symbol      string    `json:"symbol"`
	Volume      float64   `json:"volume"`
	Pips        float64   `json:"pips"`
	Commission  float64   `json:"commission"`
	Timestamp   time.Time `json:"timestamp"`
}

// AffiliateService manages affiliates and commissions
type AffiliateService struct {
	mu                sync.RWMutex
	affiliates        map[int64]*Affiliate
	payouts           map[int64]*Payout
	commissionRecords map[int64][]*CommissionRecord // Affiliate ID -> records
	nextAffiliateID   int64
	nextPayoutID      int64
	nextCommissionID  int64
	tiers             []AffiliateCommissionTier
}

// NewAffiliateService creates a new affiliate service with mock data
func NewAffiliateService() *AffiliateService {
	svc := &AffiliateService{
		affiliates:        make(map[int64]*Affiliate),
		payouts:           make(map[int64]*Payout),
		commissionRecords: make(map[int64][]*CommissionRecord),
		nextAffiliateID:   1,
		nextPayoutID:      1,
		nextCommissionID:  1,
		tiers: []AffiliateCommissionTier{
			{Name: "Bronze", MinVolume: 0, MaxVolume: 100000, PipCommission: 0.5},
			{Name: "Silver", MinVolume: 100000, MaxVolume: 500000, PipCommission: 0.8},
			{Name: "Gold", MinVolume: 500000, MaxVolume: 1000000, PipCommission: 1.0},
			{Name: "Platinum", MinVolume: 1000000, MaxVolume: 0, PipCommission: 1.2}, // 0 = unlimited
		},
	}

	// Initialize with 15+ mock affiliates
	svc.initializeMockData()

	return svc
}

// initializeMockData creates realistic mock affiliates
func (as *AffiliateService) initializeMockData() {
	mockAffiliates := []struct {
		name            string
		email           string
		status          string
		totalVolume     float64
		commissionEarned float64
		payoutStatus    string
		referredCount   int
	}{
		{"Global FX Partners", "partners@globalfx.com", "ACTIVE", 2500000, 15600.50, "PAID", 25},
		{"TradePro Network", "admin@tradepro.net", "ACTIVE", 1800000, 12240.80, "PENDING", 18},
		{"Forex Elite IB", "contact@forexelite.io", "ACTIVE", 950000, 7820.00, "PAID", 12},
		{"Market Makers Guild", "info@mmguild.com", "ACTIVE", 650000, 4550.40, "PENDING", 8},
		{"Crypto Trading Hub", "support@cryptohub.com", "ACTIVE", 420000, 2856.00, "PAID", 15},
		{"Asian Markets Group", "asia@markets.group", "ACTIVE", 280000, 1848.00, "PENDING", 6},
		{"European Traders Co", "eu@traders.co", "ACTIVE", 150000, 1050.00, "PAID", 5},
		{"Latin America FX", "latam@fx.com", "ACTIVE", 95000, 475.00, "PENDING", 4},
		{"Middle East Trading", "me@trading.ae", "SUSPENDED", 180000, 1260.00, "PAID", 7},
		{"Pacific Rim Brokers", "pacific@brokers.au", "ACTIVE", 520000, 3640.00, "PENDING", 9},
		{"Algorithmic Traders", "algo@traders.tech", "ACTIVE", 1200000, 9360.00, "PAID", 14},
		{"Social Trading Network", "social@network.com", "ACTIVE", 340000, 2312.00, "PENDING", 11},
		{"Institutional Partners", "inst@partners.com", "ACTIVE", 3200000, 22080.00, "PAID", 32},
		{"Retail FX Affiliates", "retail@affiliates.net", "ACTIVE", 78000, 390.00, "PENDING", 3},
		{"High Frequency Trading", "hft@trading.io", "ACTIVE", 890000, 7120.00, "PAID", 10},
		{"Emerging Markets IB", "emerging@markets.com", "PENDING", 45000, 225.00, "PENDING", 2},
		{"Wealth Management Group", "wealth@mgmt.com", "ACTIVE", 1500000, 11400.00, "PAID", 20},
		{"Day Traders Alliance", "day@traders.org", "ACTIVE", 620000, 4340.00, "PENDING", 13},
	}

	now := time.Now()
	for i, mock := range mockAffiliates {
		affiliate := &Affiliate{
			ID:               int64(i + 1),
			Name:             mock.name,
			Email:            mock.email,
			Status:           mock.status,
			ReferralCode:     fmt.Sprintf("IB-%04d", i+1),
			TotalVolume:      mock.totalVolume,
			CommissionEarned: mock.commissionEarned,
			PayoutStatus:     mock.payoutStatus,
			CreatedAt:        now.AddDate(0, -6, -i*3), // Stagger creation dates
			ReferredClients:  []string{},
		}

		// Generate referred client IDs
		for j := 0; j < mock.referredCount; j++ {
			affiliate.ReferredClients = append(affiliate.ReferredClients, fmt.Sprintf("RTX-%06d", (i*100)+j+1))
		}

		// Calculate tier based on volume
		affiliate.Tier = as.calculateTier(affiliate.TotalVolume)

		as.affiliates[affiliate.ID] = affiliate

		// Generate some commission records for active affiliates
		if mock.status == "ACTIVE" && mock.referredCount > 0 {
			as.generateMockCommissions(affiliate.ID, mock.referredCount)
		}

		// Generate payout records
		if mock.payoutStatus == "PAID" {
			payout := &Payout{
				ID:          as.nextPayoutID,
				AffiliateID: affiliate.ID,
				Amount:      mock.commissionEarned,
				Status:      "COMPLETED",
				RequestedAt: now.AddDate(0, 0, -30),
				ProcessedAt: now.AddDate(0, 0, -25),
			}
			as.payouts[payout.ID] = payout
			as.nextPayoutID++
		}
	}

	as.nextAffiliateID = int64(len(mockAffiliates) + 1)
	log.Printf("[AffiliateService] Initialized with %d mock affiliates", len(mockAffiliates))
}

// generateMockCommissions creates sample commission records
func (as *AffiliateService) generateMockCommissions(affiliateID int64, count int) {
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "BTCUSD", "XAUUSD"}
	now := time.Now()

	for i := 0; i < count*2; i++ { // 2 commissions per client
		commission := &CommissionRecord{
			ID:          as.nextCommissionID,
			AffiliateID: affiliateID,
			ClientID:    fmt.Sprintf("RTX-%06d", (int(affiliateID)*100)+i+1),
			Symbol:      symbols[i%len(symbols)],
			Volume:      float64(1 + (i % 5)),     // 1-5 lots
			Pips:        float64(10 + (i % 50)),   // 10-60 pips
			Commission:  float64(5 + (i % 100)),   // $5-$105
			Timestamp:   now.AddDate(0, 0, -(i%90)), // Last 90 days
		}
		as.commissionRecords[affiliateID] = append(as.commissionRecords[affiliateID], commission)
		as.nextCommissionID++
	}
}

// calculateTier determines the tier based on total volume
func (as *AffiliateService) calculateTier(volume float64) string {
	for i := len(as.tiers) - 1; i >= 0; i-- {
		tier := as.tiers[i]
		if volume >= tier.MinVolume && (tier.MaxVolume == 0 || volume <= tier.MaxVolume) {
			return tier.Name
		}
	}
	return "Bronze" // Default
}

// ListAffiliates returns all affiliates
func (as *AffiliateService) ListAffiliates() []*Affiliate {
	as.mu.RLock()
	defer as.mu.RUnlock()

	affiliates := make([]*Affiliate, 0, len(as.affiliates))
	for _, aff := range as.affiliates {
		affiliates = append(affiliates, aff)
	}
	return affiliates
}

// GetAffiliate returns a single affiliate with detailed info
func (as *AffiliateService) GetAffiliate(id int64) (*Affiliate, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	aff, ok := as.affiliates[id]
	if !ok {
		return nil, fmt.Errorf("affiliate %d not found", id)
	}
	return aff, nil
}

// CreateAffiliate creates a new affiliate
func (as *AffiliateService) CreateAffiliate(name, email string) (*Affiliate, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if name == "" || email == "" {
		return nil, fmt.Errorf("name and email are required")
	}

	affiliate := &Affiliate{
		ID:               as.nextAffiliateID,
		Name:             name,
		Email:            email,
		Status:           "PENDING",
		Tier:             "Bronze",
		ReferralCode:     fmt.Sprintf("IB-%04d", as.nextAffiliateID),
		ReferredClients:  []string{},
		TotalVolume:      0,
		CommissionEarned: 0,
		PayoutStatus:     "PENDING",
		CreatedAt:        time.Now(),
	}

	as.affiliates[as.nextAffiliateID] = affiliate
	as.nextAffiliateID++

	log.Printf("[AffiliateService] Created affiliate #%d: %s (%s)", affiliate.ID, name, email)
	return affiliate, nil
}

// UpdateAffiliate updates an affiliate's details
func (as *AffiliateService) UpdateAffiliate(id int64, status *string, tier *string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	aff, ok := as.affiliates[id]
	if !ok {
		return fmt.Errorf("affiliate %d not found", id)
	}

	if status != nil {
		validStatuses := map[string]bool{"ACTIVE": true, "SUSPENDED": true, "PENDING": true}
		if !validStatuses[*status] {
			return fmt.Errorf("invalid status: %s", *status)
		}
		aff.Status = *status
		log.Printf("[AffiliateService] Updated affiliate #%d status to %s", id, *status)
	}

	if tier != nil {
		validTiers := map[string]bool{"Bronze": true, "Silver": true, "Gold": true, "Platinum": true}
		if !validTiers[*tier] {
			return fmt.Errorf("invalid tier: %s", *tier)
		}
		aff.Tier = *tier
		log.Printf("[AffiliateService] Updated affiliate #%d tier to %s", id, *tier)
	}

	return nil
}

// GetAffiliatePayouts returns all payouts for an affiliate
func (as *AffiliateService) GetAffiliatePayouts(affiliateID int64) []*Payout {
	as.mu.RLock()
	defer as.mu.RUnlock()

	payouts := make([]*Payout, 0)
	for _, payout := range as.payouts {
		if payout.AffiliateID == affiliateID {
			payouts = append(payouts, payout)
		}
	}
	return payouts
}

// ProcessPayout creates a payout request
func (as *AffiliateService) ProcessPayout(affiliateID int64, amount float64) (*Payout, error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	aff, ok := as.affiliates[affiliateID]
	if !ok {
		return nil, fmt.Errorf("affiliate %d not found", affiliateID)
	}

	if aff.Status != "ACTIVE" {
		return nil, fmt.Errorf("affiliate must be ACTIVE to process payouts")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("payout amount must be positive")
	}

	if amount > aff.CommissionEarned {
		return nil, fmt.Errorf("payout amount exceeds commission earned")
	}

	payout := &Payout{
		ID:          as.nextPayoutID,
		AffiliateID: affiliateID,
		Amount:      amount,
		Status:      "PROCESSING",
		RequestedAt: time.Now(),
	}

	as.payouts[payout.ID] = payout
	as.nextPayoutID++

	// Update affiliate payout status
	aff.PayoutStatus = "PROCESSING"

	log.Printf("[AffiliateService] Created payout #%d for affiliate #%d: $%.2f", payout.ID, affiliateID, amount)
	return payout, nil
}

// GetCommissionRecords returns commission history for an affiliate
func (as *AffiliateService) GetCommissionRecords(affiliateID int64, limit int) []*CommissionRecord {
	as.mu.RLock()
	defer as.mu.RUnlock()

	records := as.commissionRecords[affiliateID]
	if limit > 0 && limit < len(records) {
		return records[:limit]
	}
	return records
}

// GetAffiliateCommissionTiers returns all available commission tiers
func (as *AffiliateService) GetAffiliateCommissionTiers() []AffiliateCommissionTier {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.tiers
}

// AffiliateHandler handles HTTP requests for affiliate management
type AffiliateHandler struct {
	svc     *AffiliateService
	authSvc *AuthService
}

// NewAffiliateHandler creates a new affiliate handler
func NewAffiliateHandler(svc *AffiliateService, authSvc *AuthService) *AffiliateHandler {
	return &AffiliateHandler{
		svc:     svc,
		authSvc: authSvc,
	}
}

// ListAffiliates handles GET /admin/affiliates
func (ah *AffiliateHandler) ListAffiliates(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	affiliates := ah.svc.ListAffiliates()
	log.Printf("[AffiliateHandler] Admin %s listed %d affiliates", admin.Username, len(affiliates))

	respondJSON(w, map[string]interface{}{
		"success":    true,
		"affiliates": affiliates,
		"count":      len(affiliates),
	})
}

// GetAffiliate handles GET /admin/affiliates/:id
func (ah *AffiliateHandler) GetAffiliate(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract affiliate ID
	affiliateID, err := ah.extractAffiliateID(r)
	if err != nil {
		respondError(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	affiliate, err := ah.svc.GetAffiliate(affiliateID)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get payouts
	payouts := ah.svc.GetAffiliatePayouts(affiliateID)

	log.Printf("[AffiliateHandler] Admin %s retrieved affiliate #%d", admin.Username, affiliateID)

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"affiliate": affiliate,
		"payouts":   payouts,
	})
}

// CreateAffiliate handles POST /admin/affiliates
func (ah *AffiliateHandler) CreateAffiliate(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	affiliate, err := ah.svc.CreateAffiliate(req.Name, req.Email)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[AffiliateHandler] Admin %s created affiliate #%d: %s", admin.Username, affiliate.ID, req.Name)

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"affiliate": affiliate,
		"message":   "Affiliate created successfully",
	})
}

// UpdateAffiliate handles PUT /admin/affiliates/:id
func (ah *AffiliateHandler) UpdateAffiliate(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract affiliate ID
	affiliateID, err := ah.extractAffiliateID(r)
	if err != nil {
		respondError(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status *string `json:"status,omitempty"`
		Tier   *string `json:"tier,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ah.svc.UpdateAffiliate(affiliateID, req.Status, req.Tier); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch updated affiliate
	affiliate, _ := ah.svc.GetAffiliate(affiliateID)

	log.Printf("[AffiliateHandler] Admin %s updated affiliate #%d", admin.Username, affiliateID)

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"affiliate": affiliate,
		"message":   "Affiliate updated successfully",
	})
}

// ProcessPayout handles POST /admin/affiliates/:id/payout
func (ah *AffiliateHandler) ProcessPayout(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract affiliate ID
	affiliateID, err := ah.extractAffiliateID(r)
	if err != nil {
		respondError(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	payout, err := ah.svc.ProcessPayout(affiliateID, req.Amount)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[AffiliateHandler] Admin %s processed payout for affiliate #%d: $%.2f",
		admin.Username, affiliateID, req.Amount)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"payout":  payout,
		"message": "Payout request created",
	})
}

// GetCommissionHistory handles GET /admin/affiliates/:id/commissions
func (ah *AffiliateHandler) GetCommissionHistory(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract affiliate ID
	affiliateID, err := ah.extractAffiliateID(r)
	if err != nil {
		respondError(w, "Invalid affiliate ID", http.StatusBadRequest)
		return
	}

	// Verify affiliate exists
	_, err = ah.svc.GetAffiliate(affiliateID)
	if err != nil {
		respondError(w, "Affiliate not found", http.StatusNotFound)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	commissions := ah.svc.GetCommissionRecords(affiliateID, limit)

	log.Printf("[AffiliateHandler] Admin %s retrieved %d commissions for affiliate #%d",
		admin.Username, len(commissions), affiliateID)

	respondJSON(w, map[string]interface{}{
		"success":     true,
		"affiliateId": affiliateID,
		"commissions": commissions,
		"count":       len(commissions),
	})
}

// GetAffiliateCommissionTiers handles GET /admin/affiliates/tiers
func (ah *AffiliateHandler) GetAffiliateCommissionTiers(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := ah.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tiers := ah.svc.GetAffiliateCommissionTiers()

	log.Printf("[AffiliateHandler] Admin %s retrieved commission tiers", admin.Username)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"tiers":   tiers,
	})
}

// Helper methods

func (ah *AffiliateHandler) authenticate(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, http.ErrNoCookie
	}

	sessionID := parts[1]
	ipAddress := getIPAddress(r)

	admin, err := ah.authSvc.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (ah *AffiliateHandler) extractAffiliateID(r *http.Request) (int64, error) {
	// Extract affiliate ID from URL path
	// Expected paths:
	// - /admin/affiliates/:id
	// - /admin/affiliates/:id/payout
	// - /admin/affiliates/:id/commissions
	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Find "affiliates" and get the next part
	for i, part := range parts {
		if part == "affiliates" && i+1 < len(parts) {
			idStr := parts[i+1]
			// Skip if it's a subresource
			if idStr == "payout" || idStr == "commissions" || idStr == "tiers" {
				continue
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				return 0, err
			}
			return id, nil
		}
	}

	return 0, http.ErrNoCookie
}
