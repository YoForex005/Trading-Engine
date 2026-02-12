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
// Data Structures
// ============================================

// ActivityEntry represents a client activity log entry
type ActivityEntry struct {
	ID         int64     `json:"id"`
	ClientID   int64     `json:"clientId"`
	ClientName string    `json:"clientName"`
	EventType  string    `json:"eventType"` // login, logout, trade_opened, trade_closed, deposit, withdrawal, password_change, settings_change, kyc_submitted, api_key_created
	IPAddress  string    `json:"ipAddress"`
	Device     string    `json:"device"`
	Browser    string    `json:"browser"`
	Country    string    `json:"country"`
	Details    string    `json:"details"`
	Timestamp  time.Time `json:"timestamp"`
}

// SuspiciousAlert represents a suspicious activity alert
type SuspiciousAlert struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"clientId"`
	ClientName  string    `json:"clientName"`
	AlertType   string    `json:"alertType"` // failed_logins, new_country, concurrent_sessions
	Description string    `json:"description"`
	IPAddress   string    `json:"ipAddress"`
	Severity    string    `json:"severity"` // low, medium, high
	Dismissed   bool      `json:"dismissed"`
	Timestamp   time.Time `json:"timestamp"`
}

// ActivityStats represents activity statistics
type ActivityStats struct {
	ActiveSessionsNow int `json:"activeSessionsNow"`
	LoginsToday       int `json:"loginsToday"`
	FailedLogins      int `json:"failedLogins"`
	UniqueIPs         int `json:"uniqueIps"`
}

// GeoStats represents geographical distribution
type GeoStats struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

// ============================================
// In-Memory Store
// ============================================

type ActivityLogStore struct {
	mu      sync.RWMutex
	entries []*ActivityEntry
	alerts  []*SuspiciousAlert
	nextID  int64
}

func NewActivityLogStore() *ActivityLogStore {
	store := &ActivityLogStore{
		entries: make([]*ActivityEntry, 0),
		alerts:  make([]*SuspiciousAlert, 0),
		nextID:  1,
	}

	now := time.Now()

	// Client names pool (30 unique clients)
	clientNames := []string{
		"John Smith", "Sarah Johnson", "Michael Williams", "Emily Brown", "David Jones",
		"Jessica Garcia", "Christopher Miller", "Amanda Davis", "Matthew Rodriguez", "Ashley Martinez",
		"Daniel Hernandez", "Jennifer Lopez", "James Wilson", "Elizabeth Anderson", "Robert Taylor",
		"Linda Thomas", "William Moore", "Barbara Jackson", "Richard White", "Susan Harris",
		"Joseph Martin", "Karen Thompson", "Charles Garcia", "Nancy Martinez", "Thomas Robinson",
		"Lisa Clark", "Christopher Lewis", "Betty Lee", "Mark Walker", "Sandra Hall",
	}

	// IP addresses pool (50 unique IPs from various countries)
	ipPool := []struct {
		IP      string
		Country string
	}{
		// USA (15 IPs)
		{"192.168.1.100", "United States"}, {"192.168.1.101", "United States"}, {"192.168.1.102", "United States"},
		{"74.125.224.72", "United States"}, {"74.125.224.73", "United States"}, {"74.125.224.74", "United States"},
		{"172.217.14.206", "United States"}, {"172.217.14.207", "United States"}, {"172.217.14.208", "United States"},
		{"104.244.42.1", "United States"}, {"104.244.42.2", "United States"}, {"104.244.42.3", "United States"},
		{"199.59.148.10", "United States"}, {"199.59.148.11", "United States"}, {"199.59.148.12", "United States"},

		// UK (8 IPs)
		{"82.102.23.45", "United Kingdom"}, {"82.102.23.46", "United Kingdom"}, {"82.102.23.47", "United Kingdom"},
		{"151.101.1.69", "United Kingdom"}, {"151.101.1.70", "United Kingdom"}, {"151.101.1.71", "United Kingdom"},
		{"185.45.5.1", "United Kingdom"}, {"185.45.5.2", "United Kingdom"},

		// Germany (5 IPs)
		{"46.101.135.15", "Germany"}, {"46.101.135.16", "Germany"}, {"46.101.135.17", "Germany"},
		{"195.201.20.1", "Germany"}, {"195.201.20.2", "Germany"},

		// Japan (5 IPs)
		{"210.152.135.1", "Japan"}, {"210.152.135.2", "Japan"}, {"210.152.135.3", "Japan"},
		{"133.242.128.1", "Japan"}, {"133.242.128.2", "Japan"},

		// Singapore (4 IPs)
		{"13.250.177.1", "Singapore"}, {"13.250.177.2", "Singapore"},
		{"54.169.1.1", "Singapore"}, {"54.169.1.2", "Singapore"},

		// Australia (4 IPs)
		{"1.144.0.1", "Australia"}, {"1.144.0.2", "Australia"},
		{"203.10.76.1", "Australia"}, {"203.10.76.2", "Australia"},

		// Canada (3 IPs)
		{"142.44.215.1", "Canada"}, {"142.44.215.2", "Canada"}, {"142.44.215.3", "Canada"},

		// France (3 IPs)
		{"51.15.228.1", "France"}, {"51.15.228.2", "France"}, {"51.15.228.3", "France"},

		// Netherlands (2 IPs)
		{"95.211.230.1", "Netherlands"}, {"95.211.230.2", "Netherlands"},

		// Switzerland (1 IP)
		{"194.150.245.1", "Switzerland"},
	}

	devices := []string{"Windows Desktop", "MacBook Pro", "iPhone 15", "Android Phone", "iPad", "Linux Workstation"}
	browsers := []string{"Chrome 121", "Firefox 122", "Safari 17", "Edge 121", "Opera 106"}
	eventTypes := []string{
		"login", "logout", "trade_opened", "trade_closed", "deposit",
		"withdrawal", "password_change", "settings_change", "kyc_submitted", "api_key_created",
	}

	// ============================================
	// MOCK DATA - 100 ACTIVITY ENTRIES
	// ============================================

	entryID := int64(1)

	// Generate 100 random activity entries over the last 24 hours
	for i := 0; i < 100; i++ {
		clientIdx := rand.Intn(len(clientNames))
		ipIdx := rand.Intn(len(ipPool))
		eventType := eventTypes[rand.Intn(len(eventTypes))]

		// Random timestamp within last 24 hours
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		timestamp := now.Add(-time.Duration(hoursAgo)*time.Hour - time.Duration(minutesAgo)*time.Minute)

		// Generate appropriate details based on event type
		var details string
		switch eventType {
		case "login":
			if rand.Intn(10) < 9 { // 90% success
				details = "Successful login"
			} else {
				details = "Failed login - Invalid credentials"
			}
		case "logout":
			details = "User logged out"
		case "trade_opened":
			symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD", "BTCUSD"}
			details = fmt.Sprintf("Opened %s trade, %.2f lots", symbols[rand.Intn(len(symbols))], rand.Float64()*10)
		case "trade_closed":
			pnl := (rand.Float64() - 0.4) * 5000 // -2000 to +3000
			details = fmt.Sprintf("Closed trade, P&L: $%.2f", pnl)
		case "deposit":
			amount := float64(rand.Intn(20000)+500) + rand.Float64()*100
			details = fmt.Sprintf("Deposit processed: $%.2f via Credit Card", amount)
		case "withdrawal":
			amount := float64(rand.Intn(10000)+100) + rand.Float64()*100
			details = fmt.Sprintf("Withdrawal requested: $%.2f to Bank Account", amount)
		case "password_change":
			details = "Password changed successfully"
		case "settings_change":
			settings := []string{"Leverage updated to 1:100", "Timezone set to UTC+8", "Email notifications enabled", "2FA enabled"}
			details = settings[rand.Intn(len(settings))]
		case "kyc_submitted":
			details = "KYC documents submitted for verification"
		case "api_key_created":
			details = "New API key generated"
		}

		entry := &ActivityEntry{
			ID:         entryID,
			ClientID:   int64(200000 + clientIdx),
			ClientName: clientNames[clientIdx],
			EventType:  eventType,
			IPAddress:  ipPool[ipIdx].IP,
			Device:     devices[rand.Intn(len(devices))],
			Browser:    browsers[rand.Intn(len(browsers))],
			Country:    ipPool[ipIdx].Country,
			Details:    details,
			Timestamp:  timestamp,
		}

		store.entries = append(store.entries, entry)
		entryID++
	}

	// Sort entries by timestamp descending (most recent first)
	sort.Slice(store.entries, func(i, j int) bool {
		return store.entries[i].Timestamp.After(store.entries[j].Timestamp)
	})

	store.nextID = entryID

	// ============================================
	// MOCK DATA - 5 SUSPICIOUS ALERTS
	// ============================================

	store.alerts = append(store.alerts, &SuspiciousAlert{
		ID:          1,
		ClientID:    200005,
		ClientName:  clientNames[5],
		AlertType:   "failed_logins",
		Description: "5 failed login attempts within 10 minutes from different IPs",
		IPAddress:   "192.168.1.100",
		Severity:    "high",
		Dismissed:   false,
		Timestamp:   now.Add(-2 * time.Hour),
	})

	store.alerts = append(store.alerts, &SuspiciousAlert{
		ID:          2,
		ClientID:    200012,
		ClientName:  clientNames[12],
		AlertType:   "new_country",
		Description: "Login from new country (Japan) - previously only logged in from USA",
		IPAddress:   "210.152.135.1",
		Severity:    "medium",
		Dismissed:   false,
		Timestamp:   now.Add(-5 * time.Hour),
	})

	store.alerts = append(store.alerts, &SuspiciousAlert{
		ID:          3,
		ClientID:    200018,
		ClientName:  clientNames[18],
		AlertType:   "concurrent_sessions",
		Description: "2 active sessions from different countries (USA and UK) at the same time",
		IPAddress:   "82.102.23.45",
		Severity:    "high",
		Dismissed:   false,
		Timestamp:   now.Add(-8 * time.Hour),
	})

	store.alerts = append(store.alerts, &SuspiciousAlert{
		ID:          4,
		ClientID:    200025,
		ClientName:  clientNames[25],
		AlertType:   "failed_logins",
		Description: "3 failed login attempts with common brute-force patterns",
		IPAddress:   "74.125.224.72",
		Severity:    "medium",
		Dismissed:   true,
		Timestamp:   now.Add(-12 * time.Hour),
	})

	store.alerts = append(store.alerts, &SuspiciousAlert{
		ID:          5,
		ClientID:    200008,
		ClientName:  clientNames[8],
		AlertType:   "new_country",
		Description: "Unusual login location (Singapore) - user typically logs in from Europe",
		IPAddress:   "13.250.177.1",
		Severity:    "low",
		Dismissed:   false,
		Timestamp:   now.Add(-18 * time.Hour),
	})

	log.Printf("[ActivityLog] Activity log system initialized (100 entries across 10 event types, 30 clients, 50 IPs, 5 alerts)")

	return store
}

// ============================================
// Handler
// ============================================

type ActivityLogHandler struct {
	store       *ActivityLogStore
	authService *auth.Service
}

func NewActivityLogHandler(store *ActivityLogStore, authService *auth.Service) *ActivityLogHandler {
	return &ActivityLogHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// Handler Methods
// ============================================

// HandleList returns paginated activity log with filters
// GET /admin/activity-log?type=login&clientId=200001&ip=192.168.1.100&page=1&pageSize=20
func (h *ActivityLogHandler) HandleList(w http.ResponseWriter, r *http.Request) {
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

	// Parse filters
	eventTypeFilter := r.URL.Query().Get("type")
	clientIDFilter := r.URL.Query().Get("clientId")
	ipFilter := r.URL.Query().Get("ip")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	// Default pagination
	page := 1
	pageSize := 20

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	// Filter entries
	var filteredEntries []*ActivityEntry
	for _, entry := range h.store.entries {
		// Apply filters
		if eventTypeFilter != "" && entry.EventType != eventTypeFilter {
			continue
		}
		if clientIDFilter != "" && fmt.Sprintf("%d", entry.ClientID) != clientIDFilter {
			continue
		}
		if ipFilter != "" && entry.IPAddress != ipFilter {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	total := len(filteredEntries)
	totalPages := (total + pageSize - 1) / pageSize

	// Calculate pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		start = 0
		end = 0
	}
	if end > total {
		end = total
	}

	var paginatedEntries []*ActivityEntry
	if start < total {
		paginatedEntries = filteredEntries[start:end]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries":    paginatedEntries,
		"page":       page,
		"pageSize":   pageSize,
		"total":      total,
		"totalPages": totalPages,
	})
}

// HandleGetStats returns activity statistics
// GET /admin/activity-log/stats
func (h *ActivityLogHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
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

	stats := ActivityStats{}

	// Track unique sessions (clientID + IP combinations)
	activeSessions := make(map[string]bool)
	uniqueIPs := make(map[string]bool)
	loginsToday := 0
	failedLogins := 0

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last30Minutes := now.Add(-30 * time.Minute)

	for _, entry := range h.store.entries {
		// Active sessions (login in last 30 minutes, no logout)
		if entry.EventType == "login" && entry.Timestamp.After(last30Minutes) && entry.Details == "Successful login" {
			sessionKey := fmt.Sprintf("%d-%s", entry.ClientID, entry.IPAddress)
			activeSessions[sessionKey] = true
		}

		// Logins today
		if entry.EventType == "login" && entry.Timestamp.After(todayStart) && entry.Details == "Successful login" {
			loginsToday++
		}

		// Failed logins
		if entry.EventType == "login" && strings.Contains(entry.Details, "Failed") {
			failedLogins++
		}

		// Unique IPs
		uniqueIPs[entry.IPAddress] = true
	}

	stats.ActiveSessionsNow = len(activeSessions)
	stats.LoginsToday = loginsToday
	stats.FailedLogins = failedLogins
	stats.UniqueIPs = len(uniqueIPs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleGetGeo returns top 10 countries by activity
// GET /admin/activity-log/geo
func (h *ActivityLogHandler) HandleGetGeo(w http.ResponseWriter, r *http.Request) {
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

	// Count activities by country
	countryMap := make(map[string]int)
	for _, entry := range h.store.entries {
		countryMap[entry.Country]++
	}

	// Convert to slice and sort
	var geoStats []GeoStats
	for country, count := range countryMap {
		geoStats = append(geoStats, GeoStats{
			Country: country,
			Count:   count,
		})
	}

	sort.Slice(geoStats, func(i, j int) bool {
		return geoStats[i].Count > geoStats[j].Count
	})

	// Return top 10
	if len(geoStats) > 10 {
		geoStats = geoStats[:10]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"countries": geoStats,
		"total":     len(geoStats),
	})
}

// HandleGetAlerts returns suspicious activity alerts
// GET /admin/activity-log/alerts
func (h *ActivityLogHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
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
	alerts := make([]*SuspiciousAlert, len(h.store.alerts))
	copy(alerts, h.store.alerts)
	h.store.mu.RUnlock()

	// Sort by timestamp descending (most recent first)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Timestamp.After(alerts[j].Timestamp)
	})

	// Count by status
	activeCount := 0
	dismissedCount := 0
	for _, alert := range alerts {
		if alert.Dismissed {
			dismissedCount++
		} else {
			activeCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts":         alerts,
		"total":          len(alerts),
		"active":         activeCount,
		"dismissed":      dismissedCount,
	})
}

// HandleDismissAlert dismisses a suspicious activity alert
// PUT /admin/activity-log/alerts/:id/dismiss
func (h *ActivityLogHandler) HandleDismissAlert(w http.ResponseWriter, r *http.Request) {
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
	path := strings.TrimPrefix(r.URL.Path, "/admin/activity-log/alerts/")
	path = strings.TrimSuffix(path, "/dismiss")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}

	var alertID int64
	if _, err := fmt.Sscanf(path, "%d", &alertID); err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	found := false
	for _, alert := range h.store.alerts {
		if alert.ID == alertID {
			alert.Dismissed = true
			found = true
			log.Printf("[ActivityLog] Alert #%d dismissed for client %s (%s)", alertID, alert.ClientName, alert.AlertType)
			break
		}
	}

	if !found {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Alert dismissed successfully",
		"alertId": alertID,
	})
}
