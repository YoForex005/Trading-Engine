//go:build rtx_legacy_admin
// +build rtx_legacy_admin

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
	"github.com/google/uuid"
)

// AnnouncementType represents the type of announcement
type AnnouncementType string

const (
	AnnouncementInfo        AnnouncementType = "info"
	AnnouncementWarning     AnnouncementType = "warning"
	AnnouncementCritical    AnnouncementType = "critical"
	AnnouncementMaintenance AnnouncementType = "maintenance"
	AnnouncementPromo       AnnouncementType = "promo"
)

// TargetAudience defines who should see the announcement
type TargetAudience string

const (
	AudienceAll     TargetAudience = "all"
	AudienceGroup   TargetAudience = "group"
	AudienceAccount TargetAudience = "account"
)

// DisplayPosition defines where the announcement appears
type DisplayPosition string

const (
	PositionBanner       DisplayPosition = "banner"
	PositionPopup        DisplayPosition = "popup"
	PositionNotification DisplayPosition = "notification"
	PositionTicker       DisplayPosition = "ticker"
)

// AnnouncementStatus represents announcement lifecycle state
type AnnouncementStatus string

const (
	StatusActive    AnnouncementStatus = "active"
	StatusScheduled AnnouncementStatus = "scheduled"
	StatusExpired   AnnouncementStatus = "expired"
	StatusDraft     AnnouncementStatus = "draft"
)

// Announcement represents a platform announcement or banner
type Announcement struct {
	ID              string             `json:"id"`
	Title           string             `json:"title"`
	Message         string             `json:"message"`
	Type            AnnouncementType   `json:"type"`
	TargetAudience  TargetAudience     `json:"target_audience"`
	DisplayPosition DisplayPosition    `json:"display_position"`
	Priority        int                `json:"priority"` // 1-10, higher = more important
	Dismissable     bool               `json:"dismissable"`
	StartDate       time.Time          `json:"start_date"`
	EndDate         time.Time          `json:"end_date"`
	Status          AnnouncementStatus `json:"status"`
	ViewCount       int                `json:"view_count"`
	DismissCount    int                `json:"dismiss_count"`
	CreatedBy       string             `json:"created_by"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at,omitempty"`
}

// AnnouncementStore manages announcements in memory
type AnnouncementStore struct {
	mu            sync.RWMutex
	announcements map[string]*Announcement
	stopChan      chan struct{}
}

func NewAnnouncementStore() *AnnouncementStore {
	store := &AnnouncementStore{
		announcements: make(map[string]*Announcement),
		stopChan:      make(chan struct{}),
	}

	// Initialize with mock announcements
	store.initializeMockData()

	// Start background worker for auto-expiration
	go store.autoExpirationWorker()

	log.Println("[Announcements] Announcement system initialized with 8 mock announcements and auto-expiration")

	return store
}

func (s *AnnouncementStore) initializeMockData() {
	now := time.Now()

	mockAnnouncements := []*Announcement{
		{
			ID:              uuid.New().String(),
			Title:           "Welcome to RTX5 Trading Platform",
			Message:         "Thank you for choosing RTX5. Explore our advanced trading features and tools.",
			Type:            AnnouncementInfo,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionBanner,
			Priority:        5,
			Dismissable:     true,
			StartDate:       now.Add(-24 * time.Hour),
			EndDate:         now.Add(30 * 24 * time.Hour),
			Status:          StatusActive,
			ViewCount:       1234,
			DismissCount:    89,
			CreatedBy:       "admin",
			CreatedAt:       now.Add(-24 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "Scheduled Maintenance",
			Message:         "Our platform will undergo scheduled maintenance on Sunday, 3:00 AM - 5:00 AM UTC. Trading will be temporarily unavailable.",
			Type:            AnnouncementMaintenance,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionPopup,
			Priority:        9,
			Dismissable:     true,
			StartDate:       now.Add(-2 * time.Hour),
			EndDate:         now.Add(5 * 24 * time.Hour),
			Status:          StatusActive,
			ViewCount:       3456,
			DismissCount:    234,
			CreatedBy:       "system",
			CreatedAt:       now.Add(-2 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "New EUR/USD Spreads",
			Message:         "We've reduced EUR/USD spreads to 0.8 pips. Trade now with lower costs!",
			Type:            AnnouncementPromo,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionTicker,
			Priority:        6,
			Dismissable:     true,
			StartDate:       now.Add(-12 * time.Hour),
			EndDate:         now.Add(7 * 24 * time.Hour),
			Status:          StatusActive,
			ViewCount:       2345,
			DismissCount:    145,
			CreatedBy:       "marketing",
			CreatedAt:       now.Add(-12 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "Risk Disclaimer Update",
			Message:         "Important: Trading forex and CFDs carries a high level of risk. Please review our updated risk disclosure.",
			Type:            AnnouncementWarning,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionNotification,
			Priority:        8,
			Dismissable:     false,
			StartDate:       now.Add(-3 * 24 * time.Hour),
			EndDate:         now.Add(30 * 24 * time.Hour),
			Status:          StatusActive,
			ViewCount:       5678,
			DismissCount:    0, // Not dismissable
			CreatedBy:       "compliance",
			CreatedAt:       now.Add(-3 * 24 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "Critical: Password Security",
			Message:         "We detected unusual login attempts. Please change your password immediately and enable 2FA.",
			Type:            AnnouncementCritical,
			TargetAudience:  AudienceGroup,
			DisplayPosition: PositionPopup,
			Priority:        10,
			Dismissable:     false,
			StartDate:       now.Add(-6 * time.Hour),
			EndDate:         now.Add(24 * time.Hour),
			Status:          StatusActive,
			ViewCount:       234,
			DismissCount:    0,
			CreatedBy:       "security",
			CreatedAt:       now.Add(-6 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "Black Friday Special",
			Message:         "Get 50% bonus on deposits over $1000 this weekend only! Terms apply.",
			Type:            AnnouncementPromo,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionBanner,
			Priority:        7,
			Dismissable:     true,
			StartDate:       now.Add(2 * 24 * time.Hour),
			EndDate:         now.Add(5 * 24 * time.Hour),
			Status:          StatusScheduled,
			ViewCount:       0,
			DismissCount:    0,
			CreatedBy:       "marketing",
			CreatedAt:       now.Add(-1 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "Platform Update v2.5",
			Message:         "We've released version 2.5 with new charting tools, improved order execution, and mobile app enhancements.",
			Type:            AnnouncementInfo,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionNotification,
			Priority:        6,
			Dismissable:     true,
			StartDate:       now.Add(-7 * 24 * time.Hour),
			EndDate:         now.Add(-1 * 24 * time.Hour),
			Status:          StatusExpired,
			ViewCount:       8901,
			DismissCount:    7234,
			CreatedBy:       "product",
			CreatedAt:       now.Add(-7 * 24 * time.Hour),
		},
		{
			ID:              uuid.New().String(),
			Title:           "Draft: Q1 Trading Competition",
			Message:         "Join our Q1 trading competition and win up to $50,000 in prizes. Details coming soon.",
			Type:            AnnouncementPromo,
			TargetAudience:  AudienceAll,
			DisplayPosition: PositionBanner,
			Priority:        5,
			Dismissable:     true,
			StartDate:       now.Add(7 * 24 * time.Hour),
			EndDate:         now.Add(90 * 24 * time.Hour),
			Status:          StatusDraft,
			ViewCount:       0,
			DismissCount:    0,
			CreatedBy:       "marketing",
			CreatedAt:       now,
		},
	}

	for _, announcement := range mockAnnouncements {
		s.announcements[announcement.ID] = announcement
	}
}

// autoExpirationWorker runs every minute to check and expire announcements
func (s *AnnouncementStore) autoExpirationWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkExpiredAnnouncements()
		case <-s.stopChan:
			return
		}
	}
}

func (s *AnnouncementStore) checkExpiredAnnouncements() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for _, announcement := range s.announcements {
		// Expire active or scheduled announcements past their end date
		if (announcement.Status == StatusActive || announcement.Status == StatusScheduled) &&
			announcement.EndDate.Before(now) {
			announcement.Status = StatusExpired
			announcement.UpdatedAt = now
			expiredCount++
		}

		// Activate scheduled announcements that have reached their start date
		if announcement.Status == StatusScheduled && announcement.StartDate.Before(now) && announcement.EndDate.After(now) {
			announcement.Status = StatusActive
			announcement.UpdatedAt = now
		}
	}

	if expiredCount > 0 {
		log.Printf("[Announcements] Auto-expired %d announcements", expiredCount)
	}
}

// GetAll returns all announcements, optionally filtered by status
func (s *AnnouncementStore) GetAll(statusFilter string) []*Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	announcements := make([]*Announcement, 0)
	for _, announcement := range s.announcements {
		if statusFilter == "" || string(announcement.Status) == statusFilter {
			announcementCopy := *announcement
			announcements = append(announcements, &announcementCopy)
		}
	}

	return announcements
}

// GetActive returns only currently active announcements
func (s *AnnouncementStore) GetActive() []*Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	announcements := make([]*Announcement, 0)

	for _, announcement := range s.announcements {
		if announcement.Status == StatusActive &&
			announcement.StartDate.Before(now) &&
			announcement.EndDate.After(now) {
			announcementCopy := *announcement
			announcements = append(announcements, &announcementCopy)
		}
	}

	return announcements
}

// GetByID returns an announcement by ID
func (s *AnnouncementStore) GetByID(id string) (*Announcement, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	announcement, exists := s.announcements[id]
	if !exists {
		return nil, false
	}

	announcementCopy := *announcement
	return &announcementCopy, true
}

// Create creates a new announcement
func (s *AnnouncementStore) Create(announcement *Announcement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not provided
	if announcement.ID == "" {
		announcement.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	announcement.CreatedAt = now
	announcement.UpdatedAt = now

	// Initialize counters
	announcement.ViewCount = 0
	announcement.DismissCount = 0

	// Determine initial status based on dates
	if announcement.Status == "" {
		if announcement.StartDate.After(now) {
			announcement.Status = StatusScheduled
		} else if announcement.EndDate.Before(now) {
			announcement.Status = StatusExpired
		} else {
			announcement.Status = StatusActive
		}
	}

	s.announcements[announcement.ID] = announcement
	log.Printf("[Announcements] Created announcement: %s (Type: %s, Status: %s)", announcement.Title, announcement.Type, announcement.Status)

	return nil
}

// Update updates an existing announcement
func (s *AnnouncementStore) Update(id string, updates *Announcement) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	announcement, exists := s.announcements[id]
	if !exists {
		return fmt.Errorf("announcement not found")
	}

	// Update fields
	if updates.Title != "" {
		announcement.Title = updates.Title
	}
	if updates.Message != "" {
		announcement.Message = updates.Message
	}
	if updates.Type != "" {
		announcement.Type = updates.Type
	}
	if updates.TargetAudience != "" {
		announcement.TargetAudience = updates.TargetAudience
	}
	if updates.DisplayPosition != "" {
		announcement.DisplayPosition = updates.DisplayPosition
	}
	if updates.Priority > 0 {
		announcement.Priority = updates.Priority
	}
	announcement.Dismissable = updates.Dismissable

	if !updates.StartDate.IsZero() {
		announcement.StartDate = updates.StartDate
	}
	if !updates.EndDate.IsZero() {
		announcement.EndDate = updates.EndDate
	}
	if updates.Status != "" {
		announcement.Status = updates.Status
	}

	announcement.UpdatedAt = time.Now()

	log.Printf("[Announcements] Updated announcement: %s", id)
	return nil
}

// Delete deletes an announcement
func (s *AnnouncementStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.announcements[id]; !exists {
		return fmt.Errorf("announcement not found")
	}

	delete(s.announcements, id)
	log.Printf("[Announcements] Deleted announcement: %s", id)
	return nil
}

// RecordDismissal records that a user dismissed an announcement
func (s *AnnouncementStore) RecordDismissal(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	announcement, exists := s.announcements[id]
	if !exists {
		return fmt.Errorf("announcement not found")
	}

	if !announcement.Dismissable {
		return fmt.Errorf("announcement is not dismissable")
	}

	announcement.DismissCount++
	return nil
}

// IncrementViewCount increments the view count for an announcement
func (s *AnnouncementStore) IncrementViewCount(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if announcement, exists := s.announcements[id]; exists {
		announcement.ViewCount++
	}
}

// Stop stops background workers
func (s *AnnouncementStore) Stop() {
	close(s.stopChan)
}

// AnnouncementHandler handles HTTP requests for announcements
type AnnouncementHandler struct {
	store       *AnnouncementStore
	authService *auth.AuthService
}

func NewAnnouncementHandler(store *AnnouncementStore, authService *auth.AuthService) *AnnouncementHandler {
	return &AnnouncementHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListAll returns all announcements with optional status filter
// GET /admin/announcements?status=active
func (h *AnnouncementHandler) HandleListAll(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get status filter from query params
	statusFilter := r.URL.Query().Get("status")

	announcements := h.store.GetAll(statusFilter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

// HandleListActive returns only active announcements
// GET /admin/announcements/active
func (h *AnnouncementHandler) HandleListActive(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	announcements := h.store.GetActive()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}

// HandleCreate creates a new announcement
// POST /admin/announcements
func (h *AnnouncementHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var announcement Announcement
	if err := json.NewDecoder(r.Body).Decode(&announcement); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if announcement.Title == "" || announcement.Message == "" {
		http.Error(w, "Title and Message are required", http.StatusBadRequest)
		return
	}

	if err := h.store.Create(&announcement); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(announcement)
}

// HandleUpdate updates an existing announcement
// PUT /admin/announcements/:id
func (h *AnnouncementHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]

	var updates Announcement
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Update(id, &updates); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get updated announcement
	announcement, _ := h.store.GetByID(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcement)
}

// HandleDelete deletes an announcement
// DELETE /admin/announcements/:id
func (h *AnnouncementHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]

	if err := h.store.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDismiss records a dismissal for an announcement
// POST /admin/announcements/:id/dismiss
func (h *AnnouncementHandler) HandleDismiss(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Require admin authentication
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path: /admin/announcements/:id/dismiss
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	id := pathParts[3]

	if err := h.store.RecordDismissal(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "dismissed"})
}

// HandlePublicActive returns active announcements for clients (no auth required)
// GET /api/announcements
func (h *AnnouncementHandler) HandlePublicActive(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	announcements := h.store.GetActive()

	// Increment view counts for returned announcements
	for _, announcement := range announcements {
		h.store.IncrementViewCount(announcement.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(announcements)
}
