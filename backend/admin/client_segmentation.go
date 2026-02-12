package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

// Segment represents a client segment with rules
type Segment struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"` // predefined, custom
	Rules       []SegmentRule  `json:"rules"`
	ClientCount int            `json:"clientCount"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Deleted     bool           `json:"deleted,omitempty"`
}

// SegmentRule represents a single rule in a segment
type SegmentRule struct {
	Field    string      `json:"field"`    // balance, trades_per_month, days_since_registration, days_since_last_trade, monthly_volume, avg_hold_time_minutes
	Operator string      `json:"operator"` // gt, lt, gte, lte, eq, between
	Value    interface{} `json:"value"`    // number or array for between
}

// SegmentDetail includes segment with list of matching clients
type SegmentDetail struct {
	Segment
	Clients []SegmentClient `json:"clients"`
}

// SegmentClient represents a client in a segment
type SegmentClient struct {
	ClientID   int64     `json:"clientId"`
	ClientName string    `json:"clientName"`
	Balance    float64   `json:"balance"`
	Email      string    `json:"email"`
	JoinedAt   time.Time `json:"joinedAt"`
}

// Tag represents a custom client tag
type Tag struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	Description string    `json:"description"`
	ClientCount int       `json:"clientCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ClientTag represents a tag applied to a client
type ClientTag struct {
	ClientID int64 `json:"clientId"`
	TagID    int64 `json:"tagId"`
}

// ============================================
// In-Memory Store
// ============================================

type SegmentStore struct {
	mu             sync.RWMutex
	segments       map[string]*Segment
	tags           map[int64]*Tag
	clientTags     []ClientTag
	nextSegmentID  int
	nextTagID      int64
	mockClients    []SegmentClient // Mock client data for segment evaluation
}

func NewSegmentStore() *SegmentStore {
	store := &SegmentStore{
		segments:      make(map[string]*Segment),
		tags:          make(map[int64]*Tag),
		clientTags:    make([]ClientTag, 0),
		nextSegmentID: 8,
		nextTagID:     1,
		mockClients:   generateMockClients(),
	}

	now := time.Now()

	// ============================================
	// PREDEFINED SEGMENTS (7 default segments)
	// ============================================

	// VIP Segment: Balance > $50,000
	store.segments["seg-001"] = &Segment{
		ID:          "seg-001",
		Name:        "VIP Clients",
		Description: "High-value clients with account balance over $50,000",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "balance", Operator: "gt", Value: 50000},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Active Traders: >50 trades per month
	store.segments["seg-002"] = &Segment{
		ID:          "seg-002",
		Name:        "Active Traders",
		Description: "Clients with more than 50 trades per month",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "trades_per_month", Operator: "gt", Value: 50},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// New Clients: <30 days since registration
	store.segments["seg-003"] = &Segment{
		ID:          "seg-003",
		Name:        "New Clients",
		Description: "Clients who registered less than 30 days ago",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "days_since_registration", Operator: "lt", Value: 30},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Dormant: >90 days since last trade
	store.segments["seg-004"] = &Segment{
		ID:          "seg-004",
		Name:        "Dormant Clients",
		Description: "Clients inactive for more than 90 days",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "days_since_last_trade", Operator: "gt", Value: 90},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// High Volume: >$1M monthly volume
	store.segments["seg-005"] = &Segment{
		ID:          "seg-005",
		Name:        "High Volume Traders",
		Description: "Clients with monthly trading volume over $1,000,000",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "monthly_volume", Operator: "gt", Value: 1000000},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Scalpers: Avg hold time <5 minutes
	store.segments["seg-006"] = &Segment{
		ID:          "seg-006",
		Name:        "Scalpers",
		Description: "Clients with average trade hold time under 5 minutes",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "avg_hold_time_minutes", Operator: "lt", Value: 5},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Swing Traders: Avg hold time >1 day (1440 minutes)
	store.segments["seg-007"] = &Segment{
		ID:          "seg-007",
		Name:        "Swing Traders",
		Description: "Clients with average trade hold time over 1 day",
		Type:        "predefined",
		Rules: []SegmentRule{
			{Field: "avg_hold_time_minutes", Operator: "gt", Value: 1440},
		},
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Calculate client counts for each segment
	store.recalculateSegmentCounts()

	// ============================================
	// MOCK CUSTOM TAGS (5 tags)
	// ============================================

	store.tags[1] = &Tag{
		ID:          1,
		Name:        "Requires Follow-up",
		Color:       "#FF5733",
		Description: "Clients needing immediate attention",
		ClientCount: 0,
		CreatedAt:   now,
	}

	store.tags[2] = &Tag{
		ID:          2,
		Name:        "Educational Content",
		Color:       "#3498DB",
		Description: "Clients interested in trading education",
		ClientCount: 0,
		CreatedAt:   now,
	}

	store.tags[3] = &Tag{
		ID:          3,
		Name:        "Potential Churn",
		Color:       "#E74C3C",
		Description: "Clients at risk of leaving",
		ClientCount: 0,
		CreatedAt:   now,
	}

	store.tags[4] = &Tag{
		ID:          4,
		Name:        "Institutional",
		Color:       "#9B59B6",
		Description: "Institutional or corporate accounts",
		ClientCount: 0,
		CreatedAt:   now,
	}

	store.tags[5] = &Tag{
		ID:          5,
		Name:        "Referred",
		Color:       "#2ECC71",
		Description: "Clients who came through referral program",
		ClientCount: 0,
		CreatedAt:   now,
	}

	store.nextTagID = 6

	// Apply some mock tags to clients
	store.clientTags = append(store.clientTags,
		ClientTag{ClientID: 300001, TagID: 1},
		ClientTag{ClientID: 300001, TagID: 2},
		ClientTag{ClientID: 300005, TagID: 3},
		ClientTag{ClientID: 300008, TagID: 4},
		ClientTag{ClientID: 300012, TagID: 5},
		ClientTag{ClientID: 300015, TagID: 2},
		ClientTag{ClientID: 300018, TagID: 1},
	)

	// Recalculate tag counts
	store.recalculateTagCounts()

	log.Printf("[Segmentation] Client segmentation system initialized (7 predefined segments, 5 custom tags)")

	return store
}

// generateMockClients creates mock client data for segment evaluation
func generateMockClients() []SegmentClient {
	now := time.Now()
	clients := []SegmentClient{
		// VIP clients (balance > $50k)
		{ClientID: 300001, ClientName: "Alexander Morgan", Balance: 125000.00, Email: "alexander.morgan@email.com", JoinedAt: now.AddDate(0, -8, 0)},
		{ClientID: 300002, ClientName: "Victoria Chen", Balance: 87500.00, Email: "victoria.chen@email.com", JoinedAt: now.AddDate(0, -12, 0)},
		{ClientID: 300003, ClientName: "Sebastian Rodriguez", Balance: 156000.00, Email: "sebastian.r@email.com", JoinedAt: now.AddDate(0, -18, 0)},
		{ClientID: 300004, ClientName: "Isabella Thompson", Balance: 92000.00, Email: "isabella.t@email.com", JoinedAt: now.AddDate(-1, -2, 0)},
		{ClientID: 300005, ClientName: "Maximilian Schmidt", Balance: 78000.00, Email: "max.schmidt@email.com", JoinedAt: now.AddDate(0, -6, 0)},

		// Active traders (balance $20k-$50k)
		{ClientID: 300006, ClientName: "Olivia Anderson", Balance: 42000.00, Email: "olivia.a@email.com", JoinedAt: now.AddDate(0, -4, 0)},
		{ClientID: 300007, ClientName: "Benjamin Lee", Balance: 38500.00, Email: "benjamin.lee@email.com", JoinedAt: now.AddDate(0, -7, 0)},
		{ClientID: 300008, ClientName: "Charlotte Martinez", Balance: 45000.00, Email: "charlotte.m@email.com", JoinedAt: now.AddDate(0, -10, 0)},
		{ClientID: 300009, ClientName: "William Taylor", Balance: 33000.00, Email: "william.t@email.com", JoinedAt: now.AddDate(0, -5, 0)},
		{ClientID: 300010, ClientName: "Sophia Wilson", Balance: 41500.00, Email: "sophia.w@email.com", JoinedAt: now.AddDate(0, -9, 0)},

		// New clients (<30 days)
		{ClientID: 300011, ClientName: "James Brown", Balance: 15000.00, Email: "james.brown@email.com", JoinedAt: now.AddDate(0, 0, -15)},
		{ClientID: 300012, ClientName: "Emma Davis", Balance: 22000.00, Email: "emma.davis@email.com", JoinedAt: now.AddDate(0, 0, -8)},
		{ClientID: 300013, ClientName: "Oliver Garcia", Balance: 18500.00, Email: "oliver.garcia@email.com", JoinedAt: now.AddDate(0, 0, -25)},
		{ClientID: 300014, ClientName: "Amelia Miller", Balance: 12000.00, Email: "amelia.m@email.com", JoinedAt: now.AddDate(0, 0, -5)},
		{ClientID: 300015, ClientName: "Lucas Johnson", Balance: 20000.00, Email: "lucas.j@email.com", JoinedAt: now.AddDate(0, 0, -12)},

		// Dormant clients
		{ClientID: 300016, ClientName: "Mia White", Balance: 8500.00, Email: "mia.white@email.com", JoinedAt: now.AddDate(-1, -6, 0)},
		{ClientID: 300017, ClientName: "Ethan Harris", Balance: 6200.00, Email: "ethan.h@email.com", JoinedAt: now.AddDate(-2, 0, 0)},
		{ClientID: 300018, ClientName: "Ava Clark", Balance: 4800.00, Email: "ava.clark@email.com", JoinedAt: now.AddDate(-1, -8, 0)},

		// Mid-tier clients
		{ClientID: 300019, ClientName: "Noah Lewis", Balance: 28000.00, Email: "noah.lewis@email.com", JoinedAt: now.AddDate(0, -14, 0)},
		{ClientID: 300020, ClientName: "Emily Walker", Balance: 31500.00, Email: "emily.w@email.com", JoinedAt: now.AddDate(0, -11, 0)},
		{ClientID: 300021, ClientName: "Liam Hall", Balance: 27000.00, Email: "liam.hall@email.com", JoinedAt: now.AddDate(0, -16, 0)},
		{ClientID: 300022, ClientName: "Harper Allen", Balance: 24500.00, Email: "harper.a@email.com", JoinedAt: now.AddDate(0, -13, 0)},
		{ClientID: 300023, ClientName: "Mason Young", Balance: 29000.00, Email: "mason.y@email.com", JoinedAt: now.AddDate(0, -9, 0)},

		// Smaller accounts
		{ClientID: 300024, ClientName: "Ella King", Balance: 9500.00, Email: "ella.king@email.com", JoinedAt: now.AddDate(0, -3, 0)},
		{ClientID: 300025, ClientName: "Logan Wright", Balance: 11000.00, Email: "logan.w@email.com", JoinedAt: now.AddDate(0, -7, 0)},
		{ClientID: 300026, ClientName: "Avery Scott", Balance: 13500.00, Email: "avery.s@email.com", JoinedAt: now.AddDate(0, -5, 0)},
		{ClientID: 300027, ClientName: "Jackson Green", Balance: 10500.00, Email: "jackson.g@email.com", JoinedAt: now.AddDate(0, -8, 0)},
		{ClientID: 300028, ClientName: "Aria Adams", Balance: 16000.00, Email: "aria.a@email.com", JoinedAt: now.AddDate(0, -10, 0)},
		{ClientID: 300029, ClientName: "Carter Baker", Balance: 14500.00, Email: "carter.b@email.com", JoinedAt: now.AddDate(0, -6, 0)},
		{ClientID: 300030, ClientName: "Scarlett Nelson", Balance: 19000.00, Email: "scarlett.n@email.com", JoinedAt: now.AddDate(0, -12, 0)},
	}

	return clients
}

// recalculateSegmentCounts updates client counts for all segments
func (s *SegmentStore) recalculateSegmentCounts() {
	for segID, segment := range s.segments {
		count := 0
		for _, client := range s.mockClients {
			if s.clientMatchesSegment(client, segment) {
				count++
			}
		}
		s.segments[segID].ClientCount = count
	}
}

// recalculateTagCounts updates client counts for all tags
func (s *SegmentStore) recalculateTagCounts() {
	tagCounts := make(map[int64]int)
	for _, ct := range s.clientTags {
		tagCounts[ct.TagID]++
	}
	for tagID, count := range tagCounts {
		if tag, exists := s.tags[tagID]; exists {
			tag.ClientCount = count
		}
	}
}

// clientMatchesSegment checks if a client matches segment rules
func (s *SegmentStore) clientMatchesSegment(client SegmentClient, segment *Segment) bool {
	now := time.Now()

	for _, rule := range segment.Rules {
		var fieldValue float64

		switch rule.Field {
		case "balance":
			fieldValue = client.Balance
		case "trades_per_month":
			// Mock: Derive from balance (higher balance = more trades)
			fieldValue = client.Balance / 1000
		case "days_since_registration":
			fieldValue = float64(now.Sub(client.JoinedAt).Hours() / 24)
		case "days_since_last_trade":
			// Mock: Random based on client ID
			fieldValue = float64((client.ClientID % 200) + 1)
		case "monthly_volume":
			// Mock: Higher balance = higher volume
			fieldValue = client.Balance * 15
		case "avg_hold_time_minutes":
			// Mock: Derive from client ID pattern
			if client.ClientID%3 == 0 {
				fieldValue = 3 // Scalper
			} else if client.ClientID%5 == 0 {
				fieldValue = 2000 // Swing trader
			} else {
				fieldValue = 120 // Day trader
			}
		default:
			return false
		}

		// Evaluate operator
		switch rule.Operator {
		case "gt":
			if val, ok := rule.Value.(float64); ok {
				if !(fieldValue > val) {
					return false
				}
			}
		case "lt":
			if val, ok := rule.Value.(float64); ok {
				if !(fieldValue < val) {
					return false
				}
			}
		case "gte":
			if val, ok := rule.Value.(float64); ok {
				if !(fieldValue >= val) {
					return false
				}
			}
		case "lte":
			if val, ok := rule.Value.(float64); ok {
				if !(fieldValue <= val) {
					return false
				}
			}
		case "eq":
			if val, ok := rule.Value.(float64); ok {
				if !(fieldValue == val) {
					return false
				}
			}
		}
	}

	return true
}

// ============================================
// Handler
// ============================================

type SegmentHandler struct {
	store       *SegmentStore
	authService *auth.Service
}

func NewSegmentHandler(store *SegmentStore, authService *auth.Service) *SegmentHandler {
	return &SegmentHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// Segment Handler Methods
// ============================================

// HandleListSegments returns all segments with client counts
// GET /admin/segments
func (h *SegmentHandler) HandleListSegments(w http.ResponseWriter, r *http.Request) {
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

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var segments []*Segment
	for _, seg := range h.store.segments {
		if !seg.Deleted {
			segments = append(segments, seg)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"segments": segments,
		"total":    len(segments),
	})
}

// HandleGetSegmentDetail returns segment details with matching clients
// GET /admin/segments/:id
func (h *SegmentHandler) HandleGetSegmentDetail(w http.ResponseWriter, r *http.Request) {
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
	segmentID := strings.TrimPrefix(r.URL.Path, "/admin/segments/")
	if segmentID == "" || segmentID == r.URL.Path {
		http.Error(w, "Segment ID required", http.StatusBadRequest)
		return
	}

	h.store.mu.RLock()
	segment, exists := h.store.segments[segmentID]
	if !exists || segment.Deleted {
		h.store.mu.RUnlock()
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}

	// Find matching clients
	var matchingClients []SegmentClient
	for _, client := range h.store.mockClients {
		if h.store.clientMatchesSegment(client, segment) {
			matchingClients = append(matchingClients, client)
		}
	}

	h.store.mu.RUnlock()

	detail := SegmentDetail{
		Segment: *segment,
		Clients: matchingClients,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

// HandleCreateSegment creates a custom segment
// POST /admin/segments
func (h *SegmentHandler) HandleCreateSegment(w http.ResponseWriter, r *http.Request) {
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

	var newSegment Segment
	if err := json.NewDecoder(r.Body).Decode(&newSegment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	newSegment.ID = fmt.Sprintf("seg-%03d", h.store.nextSegmentID)
	h.store.nextSegmentID++
	newSegment.Type = "custom"
	newSegment.CreatedAt = time.Now()
	newSegment.UpdatedAt = time.Now()
	newSegment.Deleted = false

	// Calculate client count
	count := 0
	for _, client := range h.store.mockClients {
		if h.store.clientMatchesSegment(client, &newSegment) {
			count++
		}
	}
	newSegment.ClientCount = count

	h.store.segments[newSegment.ID] = &newSegment
	h.store.mu.Unlock()

	log.Printf("[Segmentation] Created custom segment: %s (%s)", newSegment.Name, newSegment.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newSegment)
}

// HandleUpdateSegment updates a custom segment
// PUT /admin/segments/:id
func (h *SegmentHandler) HandleUpdateSegment(w http.ResponseWriter, r *http.Request) {
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

	segmentID := strings.TrimPrefix(r.URL.Path, "/admin/segments/")
	if segmentID == "" || segmentID == r.URL.Path {
		http.Error(w, "Segment ID required", http.StatusBadRequest)
		return
	}

	var updates Segment
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	segment, exists := h.store.segments[segmentID]
	if !exists || segment.Deleted {
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}

	if segment.Type == "predefined" {
		http.Error(w, "Cannot update predefined segments", http.StatusForbidden)
		return
	}

	// Update fields
	if updates.Name != "" {
		segment.Name = updates.Name
	}
	if updates.Description != "" {
		segment.Description = updates.Description
	}
	if len(updates.Rules) > 0 {
		segment.Rules = updates.Rules
	}
	segment.UpdatedAt = time.Now()

	// Recalculate client count
	count := 0
	for _, client := range h.store.mockClients {
		if h.store.clientMatchesSegment(client, segment) {
			count++
		}
	}
	segment.ClientCount = count

	log.Printf("[Segmentation] Updated custom segment: %s (%s)", segment.Name, segment.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(segment)
}

// HandleDeleteSegment soft deletes a custom segment
// DELETE /admin/segments/:id
func (h *SegmentHandler) HandleDeleteSegment(w http.ResponseWriter, r *http.Request) {
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

	segmentID := strings.TrimPrefix(r.URL.Path, "/admin/segments/")
	if segmentID == "" || segmentID == r.URL.Path {
		http.Error(w, "Segment ID required", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	segment, exists := h.store.segments[segmentID]
	if !exists || segment.Deleted {
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}

	if segment.Type == "predefined" {
		http.Error(w, "Cannot delete predefined segments", http.StatusForbidden)
		return
	}

	segment.Deleted = true
	log.Printf("[Segmentation] Deleted custom segment: %s (%s)", segment.Name, segment.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Segment deleted successfully",
		"id":      segmentID,
	})
}

// ============================================
// Tag Handler Methods
// ============================================

// HandleListTags returns all custom tags
// GET /admin/tags
func (h *SegmentHandler) HandleListTags(w http.ResponseWriter, r *http.Request) {
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

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var tags []*Tag
	for _, tag := range h.store.tags {
		tags = append(tags, tag)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tags":  tags,
		"total": len(tags),
	})
}

// HandleCreateTag creates a new custom tag
// POST /admin/tags
func (h *SegmentHandler) HandleCreateTag(w http.ResponseWriter, r *http.Request) {
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

	var newTag Tag
	if err := json.NewDecoder(r.Body).Decode(&newTag); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	newTag.ID = h.store.nextTagID
	h.store.nextTagID++
	newTag.ClientCount = 0
	newTag.CreatedAt = time.Now()

	h.store.tags[newTag.ID] = &newTag
	h.store.mu.Unlock()

	log.Printf("[Segmentation] Created custom tag: %s (ID: %d)", newTag.Name, newTag.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTag)
}

// HandleApplyTag applies a tag to a client
// POST /admin/clients/:id/tags
func (h *SegmentHandler) HandleApplyTag(w http.ResponseWriter, r *http.Request) {
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

	// Extract client ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/clients/")
	path = strings.TrimSuffix(path, "/tags")
	if path == "" {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}

	var clientID int64
	if _, err := fmt.Sscanf(path, "%d", &clientID); err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var req struct {
		TagID int64 `json:"tagId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	// Check if tag exists
	if _, exists := h.store.tags[req.TagID]; !exists {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	// Check if already applied
	for _, ct := range h.store.clientTags {
		if ct.ClientID == clientID && ct.TagID == req.TagID {
			http.Error(w, "Tag already applied to client", http.StatusConflict)
			return
		}
	}

	// Apply tag
	h.store.clientTags = append(h.store.clientTags, ClientTag{
		ClientID: clientID,
		TagID:    req.TagID,
	})

	h.store.recalculateTagCounts()

	log.Printf("[Segmentation] Applied tag %d to client %d", req.TagID, clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Tag applied successfully",
		"clientId": clientID,
		"tagId":    req.TagID,
	})
}

// HandleRemoveTag removes a tag from a client
// DELETE /admin/clients/:id/tags/:tagId
func (h *SegmentHandler) HandleRemoveTag(w http.ResponseWriter, r *http.Request) {
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

	// Extract client ID and tag ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/clients/")
	parts := strings.Split(path, "/tags/")
	if len(parts) != 2 {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	var clientID, tagID int64
	if _, err := fmt.Sscanf(parts[0], "%d", &clientID); err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &tagID); err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	// Find and remove the tag
	found := false
	for i, ct := range h.store.clientTags {
		if ct.ClientID == clientID && ct.TagID == tagID {
			h.store.clientTags = append(h.store.clientTags[:i], h.store.clientTags[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Tag not found on client", http.StatusNotFound)
		return
	}

	h.store.recalculateTagCounts()

	log.Printf("[Segmentation] Removed tag %d from client %d", tagID, clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Tag removed successfully",
		"clientId": clientID,
		"tagId":    tagID,
	})
}
