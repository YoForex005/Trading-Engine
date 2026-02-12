package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/google/uuid"
)

// IPEntry represents a whitelist or blacklist entry
type IPEntry struct {
	ID        string    `json:"id"`
	IP        string    `json:"ip"`         // IP address or CIDR range (e.g., "192.168.1.0/24")
	Type      string    `json:"type"`       // "whitelist" or "blacklist"
	Reason    string    `json:"reason"`     // Why this IP was added
	CreatedBy string    `json:"createdBy"`  // Admin user ID who created this
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`  // Zero value = never expires
}

// IPFilterStore manages IP whitelist and blacklist entries
type IPFilterStore struct {
	entries map[string]*IPEntry // ID -> IPEntry
	mu      sync.RWMutex
}

// NewIPFilterStore creates a new IP filter store
func NewIPFilterStore() *IPFilterStore {
	store := &IPFilterStore{
		entries: make(map[string]*IPEntry),
	}

	// Start cleanup goroutine for expired entries
	go store.cleanupExpiredEntries()

	return store
}

// AddEntry adds a new IP filter entry
func (s *IPFilterStore) AddEntry(ip, ipType, reason, createdBy string, expiresAt time.Time) (*IPEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate IP or CIDR
	if err := s.validateIPOrCIDR(ip); err != nil {
		return nil, fmt.Errorf("invalid IP or CIDR: %w", err)
	}

	// Validate type
	if ipType != "whitelist" && ipType != "blacklist" {
		return nil, fmt.Errorf("type must be 'whitelist' or 'blacklist'")
	}

	entry := &IPEntry{
		ID:        uuid.New().String(),
		IP:        ip,
		Type:      ipType,
		Reason:    reason,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	s.entries[entry.ID] = entry

	log.Printf("[IPFilter] Added %s entry for %s (reason: %s, created by: %s)",
		ipType, ip, reason, createdBy)

	return entry, nil
}

// RemoveEntry removes an IP filter entry by ID
func (s *IPFilterStore) RemoveEntry(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		return fmt.Errorf("entry not found")
	}

	delete(s.entries, id)

	log.Printf("[IPFilter] Removed %s entry for %s (ID: %s)", entry.Type, entry.IP, id)

	return nil
}

// ListEntries returns all IP filter entries, optionally filtered by type
func (s *IPFilterStore) ListEntries(typeFilter string) []*IPEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]*IPEntry, 0)

	for _, entry := range s.entries {
		// Skip expired entries
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			continue
		}

		// Apply type filter
		if typeFilter != "" && entry.Type != typeFilter {
			continue
		}

		entries = append(entries, entry)
	}

	return entries
}

// CheckIP checks if an IP address is allowed based on whitelist/blacklist rules
// Returns (allowed bool, reason string)
func (s *IPFilterStore) CheckIP(ip string) (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Parse IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false, "Invalid IP address"
	}

	// Check blacklist first (takes priority)
	for _, entry := range s.entries {
		if entry.Type != "blacklist" {
			continue
		}

		// Skip expired entries
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			continue
		}

		if s.ipMatchesEntry(parsedIP, entry.IP) {
			return false, fmt.Sprintf("IP is blacklisted: %s", entry.Reason)
		}
	}

	// If there are whitelist entries, IP must be in whitelist
	hasWhitelist := false
	for _, entry := range s.entries {
		if entry.Type != "whitelist" {
			continue
		}

		// Skip expired entries
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			continue
		}

		hasWhitelist = true

		if s.ipMatchesEntry(parsedIP, entry.IP) {
			return true, "IP is whitelisted"
		}
	}

	// If no whitelist exists, allow by default
	if !hasWhitelist {
		return true, "No whitelist configured, allowing by default"
	}

	// IP not in whitelist
	return false, "IP not in whitelist"
}

// ipMatchesEntry checks if an IP matches an entry (supports CIDR ranges)
func (s *IPFilterStore) ipMatchesEntry(ip net.IP, entryIP string) bool {
	// Try parsing as CIDR first
	_, ipNet, err := net.ParseCIDR(entryIP)
	if err == nil {
		return ipNet.Contains(ip)
	}

	// Otherwise, exact IP match
	entryParsedIP := net.ParseIP(entryIP)
	if entryParsedIP == nil {
		return false
	}

	return ip.Equal(entryParsedIP)
}

// validateIPOrCIDR validates that a string is either a valid IP or CIDR range
func (s *IPFilterStore) validateIPOrCIDR(ip string) error {
	// Try parsing as IP
	if parsedIP := net.ParseIP(ip); parsedIP != nil {
		return nil
	}

	// Try parsing as CIDR
	if _, _, err := net.ParseCIDR(ip); err == nil {
		return nil
	}

	return fmt.Errorf("not a valid IP address or CIDR range")
}

// cleanupExpiredEntries runs periodically to remove expired entries
func (s *IPFilterStore) cleanupExpiredEntries() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		now := time.Now()
		expiredCount := 0

		for id, entry := range s.entries {
			if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
				delete(s.entries, id)
				expiredCount++
			}
		}

		if expiredCount > 0 {
			log.Printf("[IPFilter] Cleaned up %d expired entries", expiredCount)
		}

		s.mu.Unlock()
	}
}

// IPFilterHandler handles IP filter management endpoints
type IPFilterHandler struct {
	store       *IPFilterStore
	authService *auth.Service
}

// NewIPFilterHandler creates a new IP filter handler
func NewIPFilterHandler(store *IPFilterStore, authService *auth.Service) *IPFilterHandler {
	return &IPFilterHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListEntries lists all IP filter entries
// GET /admin/ip-filter?type=whitelist|blacklist
func (h *IPFilterHandler) HandleListEntries(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Get type filter from query params
	typeFilter := r.URL.Query().Get("type")

	// List entries
	entries := h.store.ListEntries(typeFilter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

// HandleAddEntry adds a new IP filter entry
// POST /admin/ip-filter
func (h *IPFilterHandler) HandleAddEntry(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	claims, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Parse request body
	var req struct {
		IP        string `json:"ip"`
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		ExpiresAt string `json:"expiresAt"` // ISO 8601 format, optional
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.IP == "" {
		http.Error(w, "IP is required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "Type is required (whitelist or blacklist)", http.StatusBadRequest)
		return
	}

	// Parse expiration time
	var expiresAt time.Time
	if req.ExpiresAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			http.Error(w, "Invalid expiresAt format (use ISO 8601/RFC3339)", http.StatusBadRequest)
			return
		}
		expiresAt = parsedTime
	}

	// Add entry
	entry, err := h.store.AddEntry(req.IP, req.Type, req.Reason, claims.UserID, expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// HandleRemoveEntry removes an IP filter entry
// DELETE /admin/ip-filter/:id
func (h *IPFilterHandler) HandleRemoveEntry(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Extract entry ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}
	entryID := pathParts[3]

	// Remove entry
	if err := h.store.RemoveEntry(entryID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "IP filter entry removed successfully",
	})
}

// HandleCheckIP checks if a specific IP is allowed
// GET /admin/ip-filter/check/:ip
func (h *IPFilterHandler) HandleCheckIP(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Extract IP from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid IP address", http.StatusBadRequest)
		return
	}
	ip := pathParts[4]

	// Check IP
	allowed, reason := h.store.CheckIP(ip)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ip":      ip,
		"allowed": allowed,
		"reason":  reason,
	})
}

// HandleBulkImport imports multiple IP filter entries from JSON
// POST /admin/ip-filter/import
func (h *IPFilterHandler) HandleBulkImport(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	claims, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

	// Parse request body
	var req struct {
		Entries []struct {
			IP        string `json:"ip"`
			Type      string `json:"type"`
			Reason    string `json:"reason"`
			ExpiresAt string `json:"expiresAt"`
		} `json:"entries"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Import entries
	imported := 0
	failed := 0
	errors := make([]string, 0)

	for _, entryReq := range req.Entries {
		// Parse expiration time
		var expiresAt time.Time
		if entryReq.ExpiresAt != "" {
			parsedTime, err := time.Parse(time.RFC3339, entryReq.ExpiresAt)
			if err != nil {
				failed++
				errors = append(errors, fmt.Sprintf("Invalid expiresAt for %s: %v", entryReq.IP, err))
				continue
			}
			expiresAt = parsedTime
		}

		// Add entry
		_, err := h.store.AddEntry(entryReq.IP, entryReq.Type, entryReq.Reason, claims.UserID, expiresAt)
		if err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("Failed to add %s: %v", entryReq.IP, err))
			continue
		}

		imported++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"imported": imported,
		"failed":   failed,
		"errors":   errors,
	})
}

// validateAdminAuth validates JWT token and checks admin role
func (h *IPFilterHandler) validateAdminAuth(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := auth.ValidateTokenWithDefault(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
