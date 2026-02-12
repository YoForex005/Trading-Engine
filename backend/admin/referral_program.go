package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Types
// ============================================

type ReferralStatus string

const (
	ReferralStatusPending    ReferralStatus = "pending"
	ReferralStatusRegistered ReferralStatus = "registered"
	ReferralStatusDeposited  ReferralStatus = "deposited"
	ReferralStatusTraded     ReferralStatus = "traded"
	ReferralStatusQualified  ReferralStatus = "qualified"
	ReferralStatusExpired    ReferralStatus = "expired"
)

type RewardStatus string

const (
	RewardStatusPending   RewardStatus = "pending"
	RewardStatusApproved  RewardStatus = "approved"
	RewardStatusPaid      RewardStatus = "paid"
	RewardStatusCancelled RewardStatus = "cancelled"
)

// ============================================
// Data Structures
// ============================================

type ReferralLink struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"clientId"`
	ClientName    string    `json:"clientName"`
	Code          string    `json:"code"`
	URL           string    `json:"url"`
	Clicks        int       `json:"clicks"`
	Registrations int       `json:"registrations"`
	Conversions   int       `json:"conversions"`
	TotalReward   float64   `json:"totalReward"`
	CreatedAt     time.Time `json:"createdAt"`
	IsActive      bool      `json:"isActive"`
}

type Referral struct {
	ID            string         `json:"id"`
	ReferrerID    string         `json:"referrerId"`
	ReferrerName  string         `json:"referrerName"`
	ReferredID    string         `json:"referredId"`
	ReferredName  string         `json:"referredName"`
	ReferredEmail string         `json:"referredEmail"`
	Status        ReferralStatus `json:"status"`
	RegisteredAt  time.Time      `json:"registeredAt"`
	DepositedAt   *time.Time     `json:"depositedAt,omitempty"`
	TradedAt      *time.Time     `json:"tradedAt,omitempty"`
	Reward        float64        `json:"reward"`
	RewardStatus  RewardStatus   `json:"rewardStatus"`
	LinkCode      string         `json:"linkCode"`
}

type ReferralTier struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	MinReferrals       int     `json:"minReferrals"`
	RewardPerReferral  float64 `json:"rewardPerReferral"`
	BonusMultiplier    float64 `json:"bonusMultiplier"`
	Description        string  `json:"description"`
	IsActive           bool    `json:"isActive"`
}

type ReferralStats struct {
	TotalReferrers       int     `json:"totalReferrers"`
	TotalReferred        int     `json:"totalReferred"`
	TotalConverted       int     `json:"totalConverted"`
	ConversionRate       float64 `json:"conversionRate"`
	TotalRewardsPaid     float64 `json:"totalRewardsPaid"`
	AvgRewardPerReferrer float64 `json:"avgRewardPerReferrer"`
	TopReferrerID        string  `json:"topReferrerId"`
	TopReferrerName      string  `json:"topReferrerName"`
	TopReferrerCount     int     `json:"topReferrerCount"`
}

type ReferralMonthlyTrend struct {
	Month          string `json:"month"`
	Referrals      int    `json:"referrals"`
	Conversions    int    `json:"conversions"`
	RewardsPaid    float64 `json:"rewardsPaid"`
	ConversionRate float64 `json:"conversionRate"`
}

type TopReferrer struct {
	ClientID      string  `json:"clientId"`
	ClientName    string  `json:"clientName"`
	Conversions   int     `json:"conversions"`
	Registrations int     `json:"registrations"`
	TotalReward   float64 `json:"totalReward"`
	LinkCode      string  `json:"linkCode"`
}

// ============================================
// Store
// ============================================

type ReferralStore struct {
	mu            sync.RWMutex
	links         map[string]*ReferralLink
	referrals     map[string]*Referral
	tiers         map[string]*ReferralTier
	monthlyTrends []ReferralMonthlyTrend
}

func NewReferralStore() *ReferralStore {
	store := &ReferralStore{
		links:     make(map[string]*ReferralLink),
		referrals: make(map[string]*Referral),
		tiers:     make(map[string]*ReferralTier),
	}

	// Generate mock data
	store.generateMockData()

	return store
}

func (s *ReferralStore) generateMockData() {
	now := time.Now()

	// 4 referral tiers
	tiers := []struct {
		name            string
		minReferrals    int
		rewardPerRef    float64
		bonusMultiplier float64
		description     string
	}{
		{"Bronze", 1, 50.0, 1.0, "1-5 successful referrals"},
		{"Silver", 6, 75.0, 1.2, "6-15 successful referrals"},
		{"Gold", 16, 100.0, 1.5, "16-50 successful referrals"},
		{"Platinum", 51, 150.0, 2.0, "51+ successful referrals"},
	}

	for i, tier := range tiers {
		s.tiers[fmt.Sprintf("tier-%d", i+1)] = &ReferralTier{
			ID:                 fmt.Sprintf("tier-%d", i+1),
			Name:               tier.name,
			MinReferrals:       tier.minReferrals,
			RewardPerReferral:  tier.rewardPerRef,
			BonusMultiplier:    tier.bonusMultiplier,
			Description:        tier.description,
			IsActive:           true,
		}
	}

	// 50 referral links across 50 clients
	for i := 1; i <= 50; i++ {
		code := fmt.Sprintf("REF%04d", i)
		createdAt := now.Add(-time.Duration(rand.Intn(180)) * 24 * time.Hour)

		clicks := rand.Intn(500)
		registrations := clicks / (2 + rand.Intn(8)) // 10-50% click to registration
		conversions := registrations / (2 + rand.Intn(3)) // 33-50% registration to conversion

		link := &ReferralLink{
			ID:            fmt.Sprintf("link-%d", i),
			ClientID:      fmt.Sprintf("client-%d", i),
			ClientName:    fmt.Sprintf("Client %d", i),
			Code:          code,
			URL:           fmt.Sprintf("https://rtx5.com/register?ref=%s", code),
			Clicks:        clicks,
			Registrations: registrations,
			Conversions:   conversions,
			TotalReward:   0, // Will calculate below
			CreatedAt:     createdAt,
			IsActive:      rand.Float64() > 0.1, // 90% active
		}

		s.links[link.ID] = link
	}

	// 200 referrals with varied statuses
	referralID := 1
	statuses := []ReferralStatus{
		ReferralStatusQualified, ReferralStatusQualified, ReferralStatusQualified,
		ReferralStatusTraded, ReferralStatusTraded,
		ReferralStatusDeposited, ReferralStatusDeposited,
		ReferralStatusRegistered, ReferralStatusRegistered,
		ReferralStatusPending, ReferralStatusExpired,
	}
	rewardStatuses := []RewardStatus{
		RewardStatusPaid, RewardStatusPaid,
		RewardStatusApproved,
		RewardStatusPending,
		RewardStatusCancelled,
	}

	for _, link := range s.links {
		// Each link gets 3-5 referrals on average
		refCount := 2 + rand.Intn(6)
		if referralID+refCount > 200 {
			refCount = 200 - referralID
		}
		if refCount <= 0 {
			continue
		}

		for j := 0; j < refCount; j++ {
			status := statuses[rand.Intn(len(statuses))]
			registeredAt := link.CreatedAt.Add(time.Duration(rand.Intn(120)) * 24 * time.Hour)

			if registeredAt.After(now) {
				registeredAt = now.Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour)
			}

			var depositedAt, tradedAt *time.Time
			reward := 0.0
			rewardStatus := RewardStatusPending

			// Progression: registered -> deposited -> traded -> qualified
			if status == ReferralStatusDeposited || status == ReferralStatusTraded || status == ReferralStatusQualified {
				deposited := registeredAt.Add(time.Duration(1+rand.Intn(14)) * 24 * time.Hour)
				depositedAt = &deposited

				if status == ReferralStatusTraded || status == ReferralStatusQualified {
					traded := deposited.Add(time.Duration(1+rand.Intn(7)) * 24 * time.Hour)
					tradedAt = &traded

					if status == ReferralStatusQualified {
						// Calculate reward based on tier
						tier := s.getTierForReferrals(link.Conversions)
						reward = tier.RewardPerReferral
						rewardStatus = rewardStatuses[rand.Intn(len(rewardStatuses))]

						// Add to link total if reward is paid
						if rewardStatus == RewardStatusPaid {
							link.TotalReward += reward
						}
					}
				}
			}

			referral := &Referral{
				ID:            fmt.Sprintf("referral-%d", referralID),
				ReferrerID:    link.ClientID,
				ReferrerName:  link.ClientName,
				ReferredID:    fmt.Sprintf("referred-%d", referralID),
				ReferredName:  fmt.Sprintf("Referred Client %d", referralID),
				ReferredEmail: fmt.Sprintf("referred%d@example.com", referralID),
				Status:        status,
				RegisteredAt:  registeredAt,
				DepositedAt:   depositedAt,
				TradedAt:      tradedAt,
				Reward:        reward,
				RewardStatus:  rewardStatus,
				LinkCode:      link.Code,
			}

			s.referrals[referral.ID] = referral
			referralID++

			if referralID > 200 {
				break
			}
		}

		if referralID > 200 {
			break
		}
	}

	// Generate monthly trends (last 12 months)
	s.monthlyTrends = []ReferralMonthlyTrend{}
	for i := 11; i >= 0; i-- {
		month := now.AddDate(0, -i, 0)
		monthStr := month.Format("2006-01")

		var monthReferrals, monthConversions int
		var monthRewards float64

		for _, ref := range s.referrals {
			if ref.RegisteredAt.Format("2006-01") == monthStr {
				monthReferrals++
				if ref.Status == ReferralStatusQualified {
					monthConversions++
					if ref.RewardStatus == RewardStatusPaid {
						monthRewards += ref.Reward
					}
				}
			}
		}

		conversionRate := 0.0
		if monthReferrals > 0 {
			conversionRate = float64(monthConversions) / float64(monthReferrals) * 100
		}

		s.monthlyTrends = append(s.monthlyTrends, ReferralMonthlyTrend{
			Month:          monthStr,
			Referrals:      monthReferrals,
			Conversions:    monthConversions,
			RewardsPaid:    monthRewards,
			ConversionRate: conversionRate,
		})
	}

	log.Printf("[ReferralStore] Mock data generated: %d links, %d referrals, %d tiers, %d monthly trends",
		len(s.links), len(s.referrals), len(s.tiers), len(s.monthlyTrends))
}

func (s *ReferralStore) getTierForReferrals(count int) *ReferralTier {
	var selectedTier *ReferralTier

	for _, tier := range s.tiers {
		if count >= tier.MinReferrals {
			if selectedTier == nil || tier.MinReferrals > selectedTier.MinReferrals {
				selectedTier = tier
			}
		}
	}

	if selectedTier == nil {
		// Default to first tier
		for _, tier := range s.tiers {
			if selectedTier == nil || tier.MinReferrals < selectedTier.MinReferrals {
				selectedTier = tier
			}
		}
	}

	return selectedTier
}

// ============================================
// Handler
// ============================================

type ReferralHandler struct {
	store       *ReferralStore
	authService *auth.Service
}

func NewReferralHandler(store *ReferralStore, authService *auth.Service) *ReferralHandler {
	return &ReferralHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// 1. GET /admin/referrals/links — All referral links
// ============================================

func (h *ReferralHandler) HandleListLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	links := []*ReferralLink{}
	for _, link := range h.store.links {
		links = append(links, link)
	}

	// Sort by query param
	sortBy := r.URL.Query().Get("sortBy")
	switch sortBy {
	case "clicks":
		sort.Slice(links, func(i, j int) bool {
			return links[i].Clicks > links[j].Clicks
		})
	case "conversions":
		sort.Slice(links, func(i, j int) bool {
			return links[i].Conversions > links[j].Conversions
		})
	case "reward":
		sort.Slice(links, func(i, j int) bool {
			return links[i].TotalReward > links[j].TotalReward
		})
	default:
		// Sort by created date (newest first)
		sort.Slice(links, func(i, j int) bool {
			return links[i].CreatedAt.After(links[j].CreatedAt)
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"links": links,
		"total": len(links),
	})
}

// ============================================
// 2. GET /admin/referrals/links/:id — Link details with referrals
// ============================================

func (h *ReferralHandler) HandleGetLink(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/referrals/links/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Link ID required", http.StatusBadRequest)
		return
	}
	linkID := parts[0]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	link, exists := h.store.links[linkID]
	if !exists {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	// Get all referrals for this link
	referrals := []*Referral{}
	for _, ref := range h.store.referrals {
		if ref.ReferrerID == link.ClientID {
			referrals = append(referrals, ref)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"link":      link,
		"referrals": referrals,
		"total":     len(referrals),
	})
}

// ============================================
// 3. GET /admin/referrals/list — All referrals
// ============================================

func (h *ReferralHandler) HandleListReferrals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	statusFilter := r.URL.Query().Get("status")

	referrals := []*Referral{}
	for _, ref := range h.store.referrals {
		if statusFilter != "" && string(ref.Status) != statusFilter {
			continue
		}
		referrals = append(referrals, ref)
	}

	// Sort by registered date (newest first)
	sort.Slice(referrals, func(i, j int) bool {
		return referrals[i].RegisteredAt.After(referrals[j].RegisteredAt)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"referrals": referrals,
		"total":     len(referrals),
	})
}

// ============================================
// 4. PUT /admin/referrals/:id/approve — Approve referral reward
// ============================================

func (h *ReferralHandler) HandleApproveReferral(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	var referralID string
	for i, part := range pathParts {
		if part == "referrals" && i+1 < len(pathParts) {
			referralID = pathParts[i+1]
			break
		}
	}

	if referralID == "" || referralID == "links" || referralID == "list" || referralID == "tiers" || referralID == "stats" || referralID == "top-referrers" {
		http.Error(w, "Invalid referral ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	referral, exists := h.store.referrals[referralID]
	if !exists {
		http.Error(w, "Referral not found", http.StatusNotFound)
		return
	}

	if referral.Status != ReferralStatusQualified {
		http.Error(w, "Referral not qualified for reward", http.StatusBadRequest)
		return
	}

	if referral.RewardStatus == RewardStatusPaid {
		http.Error(w, "Reward already paid", http.StatusBadRequest)
		return
	}

	// Approve and mark as paid
	referral.RewardStatus = RewardStatusPaid

	// Update link total reward
	for _, link := range h.store.links {
		if link.ClientID == referral.ReferrerID {
			link.TotalReward += referral.Reward
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(referral)
}

// ============================================
// 5. GET /admin/referrals/tiers — Referral reward tiers
// ============================================

func (h *ReferralHandler) HandleListTiers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	tiers := []*ReferralTier{}
	for _, tier := range h.store.tiers {
		tiers = append(tiers, tier)
	}

	// Sort by min referrals
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].MinReferrals < tiers[j].MinReferrals
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
		"total": len(tiers),
	})
}

// ============================================
// 6. PUT /admin/referrals/tiers/:id — Update tier rewards
// ============================================

func (h *ReferralHandler) HandleUpdateTier(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/referrals/tiers/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Tier ID required", http.StatusBadRequest)
		return
	}
	tierID := parts[0]

	var req struct {
		RewardPerReferral *float64 `json:"rewardPerReferral,omitempty"`
		BonusMultiplier   *float64 `json:"bonusMultiplier,omitempty"`
		IsActive          *bool    `json:"isActive,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	tier, exists := h.store.tiers[tierID]
	if !exists {
		http.Error(w, "Tier not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.RewardPerReferral != nil {
		tier.RewardPerReferral = *req.RewardPerReferral
	}
	if req.BonusMultiplier != nil {
		tier.BonusMultiplier = *req.BonusMultiplier
	}
	if req.IsActive != nil {
		tier.IsActive = *req.IsActive
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tier)
}

// ============================================
// 7. GET /admin/referrals/stats — Program stats
// ============================================

func (h *ReferralHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	totalReferrers := 0
	totalReferred := len(h.store.referrals)
	totalConverted := 0
	totalRewardsPaid := 0.0

	referrerMap := make(map[string]int)

	for _, ref := range h.store.referrals {
		referrerMap[ref.ReferrerID]++

		if ref.Status == ReferralStatusQualified {
			totalConverted++
		}

		if ref.RewardStatus == RewardStatusPaid {
			totalRewardsPaid += ref.Reward
		}
	}

	totalReferrers = len(referrerMap)

	conversionRate := 0.0
	if totalReferred > 0 {
		conversionRate = float64(totalConverted) / float64(totalReferred) * 100
	}

	avgRewardPerReferrer := 0.0
	if totalReferrers > 0 {
		avgRewardPerReferrer = totalRewardsPaid / float64(totalReferrers)
	}

	// Find top referrer
	topReferrerID := ""
	topReferrerName := ""
	topReferrerCount := 0

	for _, link := range h.store.links {
		if link.Conversions > topReferrerCount {
			topReferrerCount = link.Conversions
			topReferrerID = link.ClientID
			topReferrerName = link.ClientName
		}
	}

	stats := ReferralStats{
		TotalReferrers:       totalReferrers,
		TotalReferred:        totalReferred,
		TotalConverted:       totalConverted,
		ConversionRate:       conversionRate,
		TotalRewardsPaid:     totalRewardsPaid,
		AvgRewardPerReferrer: avgRewardPerReferrer,
		TopReferrerID:        topReferrerID,
		TopReferrerName:      topReferrerName,
		TopReferrerCount:     topReferrerCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stats":         stats,
		"monthlyTrends": h.store.monthlyTrends,
	})
}

// ============================================
// 8. GET /admin/referrals/top-referrers — Top 20 referrers
// ============================================

func (h *ReferralHandler) HandleGetTopReferrers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	topReferrers := []TopReferrer{}
	for _, link := range h.store.links {
		topReferrers = append(topReferrers, TopReferrer{
			ClientID:      link.ClientID,
			ClientName:    link.ClientName,
			Conversions:   link.Conversions,
			Registrations: link.Registrations,
			TotalReward:   link.TotalReward,
			LinkCode:      link.Code,
		})
	}

	// Sort by conversions (descending)
	sort.Slice(topReferrers, func(i, j int) bool {
		return topReferrers[i].Conversions > topReferrers[j].Conversions
	})

	// Take top 20
	if len(topReferrers) > 20 {
		topReferrers = topReferrers[:20]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"topReferrers": topReferrers,
		"total":        len(topReferrers),
	})
}
