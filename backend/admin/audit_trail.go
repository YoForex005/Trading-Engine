package admin

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

// AuditTrailEntry represents a single audit trail log entry
type AuditTrailEntry struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	AdminID      string                 `json:"adminId"`
	AdminName    string                 `json:"adminName"`
	ActionType   string                 `json:"actionType"` // CREATE, UPDATE, DELETE, LOGIN, EXPORT, VIEW, CONFIG
	Module       string                 `json:"module"`     // Accounts, Symbols, Orders, Settings, Risk, LP, Reports, System, Security, Users
	TargetEntity string                 `json:"targetEntity"`
	Description  string                 `json:"description"`
	IPAddress    string                 `json:"ipAddress"`
	Status       string                 `json:"status"` // success, failed, blocked
	BeforeState  map[string]interface{} `json:"beforeState,omitempty"`
	AfterState   map[string]interface{} `json:"afterState,omitempty"`
	Reason       string                 `json:"reason,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AuditTrailStats represents audit trail statistics
type AuditTrailStats struct {
	TotalActionsToday int    `json:"totalActionsToday"`
	UniqueAdmins      int    `json:"uniquAdmins"`
	MostActiveModule  string `json:"mostActiveModule"`
	FailedActions     int    `json:"failedActions"`
	SuccessRate       float64 `json:"successRate"`
}

// AdminActivity represents admin user activity summary
type AdminActivity struct {
	AdminID    string `json:"adminId"`
	AdminName  string `json:"adminName"`
	ActionCount int   `json:"actionCount"`
	LastAction time.Time `json:"lastAction"`
}

// AuditStore manages audit trail entries
type AuditStore struct {
	mu      sync.RWMutex
	entries []*AuditTrailEntry
	nextID  int
}

// NewAuditStore creates a new audit store with mock data
func NewAuditStore() *AuditStore {
	store := &AuditStore{
		entries: make([]*AuditTrailEntry, 0, 100),
		nextID:  1,
	}

	store.initMockData()
	return store
}

func (s *AuditStore) initMockData() {
	adminUsers := []struct {
		ID   string
		Name string
	}{
		{"admin-001", "John Admin"},
		{"admin-002", "Sarah Manager"},
		{"admin-003", "Mike Supervisor"},
		{"admin-004", "Emily Director"},
		{"admin-005", "David Chief"},
	}

	actionTypes := []string{"CREATE", "UPDATE", "DELETE", "LOGIN", "EXPORT", "VIEW", "CONFIG"}
	modules := []string{"Accounts", "Symbols", "Orders", "Settings", "Risk", "LP", "Reports", "System", "Security", "Users"}
	statuses := []string{"success", "success", "success", "success", "failed", "blocked"}

	ipAddresses := []string{
		"192.168.1.100", "10.0.0.50", "172.16.0.25", "203.0.113.45",
		"198.51.100.78", "192.0.2.150", "10.1.1.200", "172.20.10.5",
	}

	// Generate 100 mock audit entries
	now := time.Now()

	for i := 0; i < 100; i++ {
		admin := adminUsers[i%len(adminUsers)]
		actionType := actionTypes[i%len(actionTypes)]
		module := modules[i%len(modules)]
		status := statuses[i%len(statuses)]
		ipAddr := ipAddresses[i%len(ipAddresses)]

		timestamp := now.Add(-time.Duration(i) * time.Hour)

		entry := &AuditTrailEntry{
			ID:        fmt.Sprintf("audit-%05d", s.nextID),
			Timestamp: timestamp,
			AdminID:   admin.ID,
			AdminName: admin.Name,
			ActionType: actionType,
			Module:    module,
			IPAddress: ipAddr,
			Status:    status,
		}

		// Generate realistic entries based on action type and module
		s.generateEntryDetails(entry, i)

		s.entries = append(s.entries, entry)
		s.nextID++
	}
}

func (s *AuditStore) generateEntryDetails(entry *AuditTrailEntry, seed int) {
	switch entry.ActionType {
	case "CREATE":
		switch entry.Module {
		case "Accounts":
			entry.TargetEntity = fmt.Sprintf("Account-%d", 50000+seed)
			entry.Description = fmt.Sprintf("Created new trading account %s", entry.TargetEntity)
			entry.AfterState = map[string]interface{}{
				"accountNumber": entry.TargetEntity,
				"balance":       10000.0,
				"leverage":      100,
				"status":        "active",
			}
		case "Symbols":
			symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD", "BTCUSD"}
			symbol := symbols[seed%len(symbols)]
			entry.TargetEntity = symbol
			entry.Description = fmt.Sprintf("Added symbol %s to trading platform", symbol)
			entry.AfterState = map[string]interface{}{
				"symbol":      symbol,
				"spread":      1.5,
				"enabled":     true,
				"contractSize": 100000,
			}
		case "Users":
			entry.TargetEntity = fmt.Sprintf("user-%d", 1000+seed)
			entry.Description = fmt.Sprintf("Created new admin user %s", entry.TargetEntity)
			entry.AfterState = map[string]interface{}{
				"username":    entry.TargetEntity,
				"role":        "admin",
				"permissions": []string{"read", "write", "delete"},
			}
		}

	case "UPDATE":
		switch entry.Module {
		case "Settings":
			entry.TargetEntity = "SystemSettings"
			entry.Description = "Updated system configuration"
			entry.BeforeState = map[string]interface{}{
				"maxLeverage":     500,
				"maintenanceMode": false,
			}
			entry.AfterState = map[string]interface{}{
				"maxLeverage":     1000,
				"maintenanceMode": false,
			}
		case "Risk":
			entry.TargetEntity = "RiskParameters"
			entry.Description = "Modified risk management settings"
			entry.BeforeState = map[string]interface{}{
				"maxDrawdown":   20.0,
				"stopOutLevel":  30.0,
			}
			entry.AfterState = map[string]interface{}{
				"maxDrawdown":   25.0,
				"stopOutLevel":  20.0,
			}
		case "LP":
			lps := []string{"YOFX", "Binance", "OANDA"}
			lp := lps[seed%len(lps)]
			entry.TargetEntity = lp
			entry.Description = fmt.Sprintf("Updated LP configuration for %s", lp)
			entry.BeforeState = map[string]interface{}{
				"enabled": true,
				"weight":  50,
			}
			entry.AfterState = map[string]interface{}{
				"enabled": true,
				"weight":  75,
			}
		}

	case "DELETE":
		switch entry.Module {
		case "Accounts":
			entry.TargetEntity = fmt.Sprintf("Account-%d", 40000+seed)
			entry.Description = fmt.Sprintf("Deleted trading account %s", entry.TargetEntity)
			entry.Reason = "Dormant account cleanup"
			entry.BeforeState = map[string]interface{}{
				"accountNumber": entry.TargetEntity,
				"balance":       0.0,
				"status":        "dormant",
			}
		case "Orders":
			entry.TargetEntity = fmt.Sprintf("Order-%d", 100000+seed)
			entry.Description = fmt.Sprintf("Cancelled pending order %s", entry.TargetEntity)
			entry.Reason = "Client request"
			entry.BeforeState = map[string]interface{}{
				"orderId": entry.TargetEntity,
				"type":    "LIMIT",
				"status":  "PENDING",
			}
		}

	case "LOGIN":
		entry.TargetEntity = "AdminPanel"
		entry.Description = fmt.Sprintf("Admin login from %s", entry.IPAddress)
		entry.Metadata = map[string]interface{}{
			"userAgent":  "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			"loginTime":  entry.Timestamp.Format(time.RFC3339),
			"sessionId":  fmt.Sprintf("session-%d", seed),
		}
		if entry.Status == "failed" {
			entry.Reason = "Invalid credentials"
		}

	case "EXPORT":
		switch entry.Module {
		case "Reports":
			reportTypes := []string{"Trading History", "P&L Statement", "Compliance Report"}
			reportType := reportTypes[seed%len(reportTypes)]
			entry.TargetEntity = reportType
			entry.Description = fmt.Sprintf("Exported %s", reportType)
			entry.Metadata = map[string]interface{}{
				"format":     "CSV",
				"dateRange":  "Last 30 days",
				"recordCount": 500 + seed,
			}
		}

	case "VIEW":
		entry.TargetEntity = entry.Module
		entry.Description = fmt.Sprintf("Viewed %s dashboard", entry.Module)

	case "CONFIG":
		switch entry.Module {
		case "System":
			entry.TargetEntity = "SystemConfiguration"
			entry.Description = "Modified system-wide settings"
			entry.BeforeState = map[string]interface{}{
				"timezone":         "UTC",
				"tradingHoursMode": "24/7",
			}
			entry.AfterState = map[string]interface{}{
				"timezone":         "EST",
				"tradingHoursMode": "Market Hours Only",
			}
		case "Security":
			entry.TargetEntity = "SecurityPolicy"
			entry.Description = "Updated security policies"
			entry.BeforeState = map[string]interface{}{
				"passwordExpiry":  90,
				"sessionTimeout":  30,
			}
			entry.AfterState = map[string]interface{}{
				"passwordExpiry":  60,
				"sessionTimeout":  15,
			}
		}
	}

	// Add failure reasons for failed/blocked entries
	if entry.Status == "failed" {
		entry.Reason = "Operation failed due to validation error"
	} else if entry.Status == "blocked" {
		entry.Reason = "Insufficient permissions"
	}
}

// ListEntries returns paginated and filtered audit entries
func (s *AuditStore) ListEntries(filters map[string]string, page, limit int) ([]*AuditTrailEntry, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter entries
	filtered := make([]*AuditTrailEntry, 0)
	for _, entry := range s.entries {
		if s.matchesFilters(entry, filters) {
			filtered = append(filtered, entry)
		}
	}

	total := len(filtered)

	// Pagination
	start := (page - 1) * limit
	if start >= total {
		return []*AuditTrailEntry{}, total
	}

	end := start + limit
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

func (s *AuditStore) matchesFilters(entry *AuditTrailEntry, filters map[string]string) bool {
	// Date range filter
	if dateFrom := filters["dateFrom"]; dateFrom != "" {
		if from, err := time.Parse("2006-01-02", dateFrom); err == nil {
			if entry.Timestamp.Before(from) {
				return false
			}
		}
	}
	if dateTo := filters["dateTo"]; dateTo != "" {
		if to, err := time.Parse("2006-01-02", dateTo); err == nil {
			if entry.Timestamp.After(to.Add(24 * time.Hour)) {
				return false
			}
		}
	}

	// Admin user filter
	if adminUser := filters["adminUser"]; adminUser != "" {
		if !strings.Contains(strings.ToLower(entry.AdminName), strings.ToLower(adminUser)) &&
			!strings.Contains(strings.ToLower(entry.AdminID), strings.ToLower(adminUser)) {
			return false
		}
	}

	// Action type filter
	if actionType := filters["actionType"]; actionType != "" {
		if entry.ActionType != actionType {
			return false
		}
	}

	// Module filter
	if module := filters["module"]; module != "" {
		if entry.Module != module {
			return false
		}
	}

	// Search text filter
	if search := filters["search"]; search != "" {
		searchLower := strings.ToLower(search)
		if !strings.Contains(strings.ToLower(entry.Description), searchLower) &&
			!strings.Contains(strings.ToLower(entry.TargetEntity), searchLower) &&
			!strings.Contains(strings.ToLower(entry.AdminName), searchLower) {
			return false
		}
	}

	return true
}

// GetEntry returns a specific audit entry by ID
func (s *AuditStore) GetEntry(id string) (*AuditTrailEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, entry := range s.entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return nil, false
}

// GetStats returns audit trail statistics
func (s *AuditStore) GetStats() AuditTrailStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var totalToday int
	var failedActions int
	var successActions int
	adminSet := make(map[string]bool)
	moduleCount := make(map[string]int)

	for _, entry := range s.entries {
		if entry.Timestamp.After(todayStart) {
			totalToday++
			adminSet[entry.AdminID] = true
		}

		if entry.Status == "failed" || entry.Status == "blocked" {
			failedActions++
		} else {
			successActions++
		}

		moduleCount[entry.Module]++
	}

	// Find most active module
	mostActiveModule := ""
	maxCount := 0
	for module, count := range moduleCount {
		if count > maxCount {
			maxCount = count
			mostActiveModule = module
		}
	}

	successRate := 0.0
	totalActions := successActions + failedActions
	if totalActions > 0 {
		successRate = (float64(successActions) / float64(totalActions)) * 100
	}

	return AuditTrailStats{
		TotalActionsToday: totalToday,
		UniqueAdmins:      len(adminSet),
		MostActiveModule:  mostActiveModule,
		FailedActions:     failedActions,
		SuccessRate:       successRate,
	}
}

// GetAdmins returns list of admin users with activity
func (s *AuditStore) GetAdmins() []AdminActivity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	adminMap := make(map[string]*AdminActivity)

	for _, entry := range s.entries {
		if activity, ok := adminMap[entry.AdminID]; ok {
			activity.ActionCount++
			if entry.Timestamp.After(activity.LastAction) {
				activity.LastAction = entry.Timestamp
			}
		} else {
			adminMap[entry.AdminID] = &AdminActivity{
				AdminID:     entry.AdminID,
				AdminName:   entry.AdminName,
				ActionCount: 1,
				LastAction:  entry.Timestamp,
			}
		}
	}

	admins := make([]AdminActivity, 0, len(adminMap))
	for _, activity := range adminMap {
		admins = append(admins, *activity)
	}

	return admins
}

// ExportEntries returns filtered entries for export
func (s *AuditStore) ExportEntries(filters map[string]string) []*AuditTrailEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]*AuditTrailEntry, 0)
	for _, entry := range s.entries {
		if s.matchesFilters(entry, filters) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// ============================================
// HTTP Handlers
// ============================================

type AuditTrailHandler struct {
	store       *AuditStore
	authService *auth.Service
}

func NewAuditTrailHandler(store *AuditStore, authService *auth.Service) *AuditTrailHandler {
	return &AuditTrailHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListEntries returns paginated and filtered audit entries
func (h *AuditTrailHandler) HandleListEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	filters := map[string]string{
		"dateFrom":   query.Get("dateFrom"),
		"dateTo":     query.Get("dateTo"),
		"adminUser":  query.Get("adminUser"),
		"actionType": query.Get("actionType"),
		"module":     query.Get("module"),
		"search":     query.Get("search"),
	}

	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	entries, total := h.store.ListEntries(filters, page, limit)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries":    entries,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + limit - 1) / limit,
	})
}

// HandleGetEntry returns full details of a specific audit entry
func (h *AuditTrailHandler) HandleGetEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract entry ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}
	entryID := pathParts[len(pathParts)-1]

	entry, ok := h.store.GetEntry(entryID)
	if !ok {
		http.Error(w, "Audit entry not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(entry)
}

// HandleGetStats returns audit trail statistics
func (h *AuditTrailHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := h.store.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// HandleGetAdmins returns list of admin users with activity
func (h *AuditTrailHandler) HandleGetAdmins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admins := h.store.GetAdmins()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"admins": admins,
		"total":  len(admins),
	})
}

// HandleExport exports filtered audit entries as CSV
func (h *AuditTrailHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse filters
	query := r.URL.Query()
	filters := map[string]string{
		"dateFrom":   query.Get("dateFrom"),
		"dateTo":     query.Get("dateTo"),
		"adminUser":  query.Get("adminUser"),
		"actionType": query.Get("actionType"),
		"module":     query.Get("module"),
		"search":     query.Get("search"),
	}

	format := query.Get("format")
	if format == "" {
		format = "csv"
	}

	entries := h.store.ExportEntries(filters)

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit-trail-%s.csv", time.Now().Format("2006-01-02")))

		writer := csv.NewWriter(w)
		defer writer.Flush()

		// Write CSV header
		header := []string{"ID", "Timestamp", "Admin", "Action", "Module", "Target", "Status", "IP Address", "Description"}
		writer.Write(header)

		// Write data rows
		for _, entry := range entries {
			row := []string{
				entry.ID,
				entry.Timestamp.Format("2006-01-02 15:04:05"),
				entry.AdminName,
				entry.ActionType,
				entry.Module,
				entry.TargetEntity,
				entry.Status,
				entry.IPAddress,
				entry.Description,
			}
			writer.Write(row)
		}

		log.Printf("[AuditTrail] Exported %d entries to CSV", len(entries))
	} else {
		http.Error(w, "Unsupported export format", http.StatusBadRequest)
	}
}
