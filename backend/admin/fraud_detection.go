package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// AlertType represents the type of fraud alert
type AlertType string

const (
	AlertMultipleCountries   AlertType = "multiple_countries"
	AlertVelocityLogin       AlertType = "velocity_login"
	AlertImpossibleTravel    AlertType = "impossible_travel"
	AlertNewDevice           AlertType = "new_device"
	AlertUnusualHours        AlertType = "unusual_hours"
	AlertLargeWithdrawal     AlertType = "large_withdrawal"
	AlertMultipleFailedLogins AlertType = "multiple_failed_logins"
	AlertVPNDetected         AlertType = "vpn_detected"
	AlertTorDetected         AlertType = "tor_detected"
	AlertBlacklistedCountry  AlertType = "blacklisted_country"
)

// Severity represents the severity level of an alert
type Severity string

const (
	FraudSeverityLow      Severity = "low"
	FraudSeverityMedium   Severity = "medium"
	FraudSeverityHigh     Severity = "high"
	FraudSeverityCritical Severity = "critical"
)

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusIgnored  AlertStatus = "ignored"
)

// LoginStatus represents the status of a login attempt
type LoginStatus string

const (
	LoginSuccess LoginStatus = "success"
	LoginFailed  LoginStatus = "failed"
	LoginBlocked LoginStatus = "blocked"
)

// LoginAttempt represents a login attempt with geolocation data
type LoginAttempt struct {
	ID         string      `json:"id"`
	ClientID   string      `json:"client_id"`
	ClientName string      `json:"client_name"`
	IP         string      `json:"ip"`
	Country    string      `json:"country"`
	City       string      `json:"city"`
	Device     string      `json:"device"`
	Browser    string      `json:"browser"`
	Status     LoginStatus `json:"status"`
	Timestamp  time.Time   `json:"timestamp"`
	RiskScore  float64     `json:"risk_score"`
	Flags      []string    `json:"flags,omitempty"`
}

// FraudAlert represents a fraud detection alert
type FraudAlert struct {
	ID          string      `json:"id"`
	ClientID    string      `json:"client_id"`
	ClientName  string      `json:"client_name"`
	AlertType   AlertType   `json:"alert_type"`
	Severity    Severity    `json:"severity"`
	Description string      `json:"description"`
	DetectedAt  time.Time   `json:"detected_at"`
	ResolvedAt  *time.Time  `json:"resolved_at,omitempty"`
	ResolvedBy  string      `json:"resolved_by,omitempty"`
	Status      AlertStatus `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// FraudRule represents a fraud detection rule
type FraudRule struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         AlertType `json:"type"`
	Condition    string    `json:"condition"`
	Action       string    `json:"action"`
	IsActive     bool      `json:"is_active"`
	TriggerCount int       `json:"trigger_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GeoReport represents geographic login distribution
type GeoReport struct {
	Country       string  `json:"country"`
	LoginCount    int     `json:"login_count"`
	UniqueClients int     `json:"unique_clients"`
	RiskLevel     string  `json:"risk_level"`
	FlaggedCount  int     `json:"flagged_count"`
	SuccessRate   float64 `json:"success_rate"`
}

// FraudStats represents overall fraud statistics
type FraudStats struct {
	TotalAlerts       int                `json:"total_alerts"`
	ActiveAlerts      int                `json:"active_alerts"`
	ResolvedAlerts    int                `json:"resolved_alerts"`
	ResolutionRate    float64            `json:"resolution_rate"`
	AlertsByType      map[string]int     `json:"alerts_by_type"`
	TopRiskyCountries []string           `json:"top_risky_countries"`
	FlaggedClients    int                `json:"flagged_clients"`
	AvgRiskScore      float64            `json:"avg_risk_score"`
	BlockedLogins     int                `json:"blocked_logins"`
}

// FraudDetectionStore manages fraud detection data
type FraudDetectionStore struct {
	mu       sync.RWMutex
	logins   []LoginAttempt
	alerts   []FraudAlert
	rules    []FraudRule
	geoData  []GeoReport
}

// NewFraudDetectionStore creates a new fraud detection store with mock data
func NewFraudDetectionStore() *FraudDetectionStore {
	store := &FraudDetectionStore{
		rules: generateFraudRules(),
	}
	store.logins = generateLoginAttempts()
	store.alerts = generateFraudAlerts(store.logins)
	store.geoData = calculateGeoReports(store.logins)
	return store
}

// generateFraudRules creates 20 fraud detection rules
func generateFraudRules() []FraudRule {
	rules := []FraudRule{
		{
			ID:           "RULE001",
			Name:         "Multiple Countries in 1 Hour",
			Type:         AlertMultipleCountries,
			Condition:    "Logins from >2 countries within 60 minutes",
			Action:       "flag_account",
			IsActive:     true,
			TriggerCount: 15,
			CreatedAt:    time.Now().Add(-90 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-10 * 24 * time.Hour),
		},
		{
			ID:           "RULE002",
			Name:         "Velocity Attack Detection",
			Type:         AlertVelocityLogin,
			Condition:    ">10 login attempts within 5 minutes",
			Action:       "block_temporarily",
			IsActive:     true,
			TriggerCount: 8,
			CreatedAt:    time.Now().Add(-85 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-15 * 24 * time.Hour),
		},
		{
			ID:           "RULE003",
			Name:         "Impossible Travel",
			Type:         AlertImpossibleTravel,
			Condition:    "Physical distance >1000km in <1 hour",
			Action:       "require_2fa",
			IsActive:     true,
			TriggerCount: 12,
			CreatedAt:    time.Now().Add(-80 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-5 * 24 * time.Hour),
		},
		{
			ID:           "RULE004",
			Name:         "New Device Login",
			Type:         AlertNewDevice,
			Condition:    "Login from unrecognized device",
			Action:       "send_notification",
			IsActive:     true,
			TriggerCount: 42,
			CreatedAt:    time.Now().Add(-75 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-20 * 24 * time.Hour),
		},
		{
			ID:           "RULE005",
			Name:         "Unusual Hours Activity",
			Type:         AlertUnusualHours,
			Condition:    "Login between 2 AM - 5 AM local time",
			Action:       "flag_for_review",
			IsActive:     false,
			TriggerCount: 6,
			CreatedAt:    time.Now().Add(-70 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			ID:           "RULE006",
			Name:         "Large Withdrawal After Login",
			Type:         AlertLargeWithdrawal,
			Condition:    "Withdrawal >$50k within 30 min of login",
			Action:       "require_approval",
			IsActive:     true,
			TriggerCount: 5,
			CreatedAt:    time.Now().Add(-65 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-8 * 24 * time.Hour),
		},
		{
			ID:           "RULE007",
			Name:         "Multiple Failed Login Attempts",
			Type:         AlertMultipleFailedLogins,
			Condition:    ">5 failed attempts within 10 minutes",
			Action:       "lock_account",
			IsActive:     true,
			TriggerCount: 18,
			CreatedAt:    time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-12 * 24 * time.Hour),
		},
		{
			ID:           "RULE008",
			Name:         "VPN Detection",
			Type:         AlertVPNDetected,
			Condition:    "IP address identified as VPN",
			Action:       "flag_for_review",
			IsActive:     true,
			TriggerCount: 25,
			CreatedAt:    time.Now().Add(-55 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-3 * 24 * time.Hour),
		},
		{
			ID:           "RULE009",
			Name:         "Tor Network Detection",
			Type:         AlertTorDetected,
			Condition:    "IP address from Tor exit node",
			Action:       "block_immediately",
			IsActive:     true,
			TriggerCount: 9,
			CreatedAt:    time.Now().Add(-50 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-18 * 24 * time.Hour),
		},
		{
			ID:           "RULE010",
			Name:         "Blacklisted Country",
			Type:         AlertBlacklistedCountry,
			Condition:    "Login from high-risk country (Syria, N.Korea)",
			Action:       "block_immediately",
			IsActive:     true,
			TriggerCount: 3,
			CreatedAt:    time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-25 * 24 * time.Hour),
		},
	}

	for i := 10; i < 20; i++ {
		rule := FraudRule{
			ID:           fmt.Sprintf("RULE%03d", i+1),
			Name:         fmt.Sprintf("Custom Rule %d", i+1),
			Type:         AlertMultipleCountries,
			Condition:    "Custom condition",
			Action:       "flag_for_review",
			IsActive:     rand.Float64() > 0.3,
			TriggerCount: rand.Intn(30),
			CreatedAt:    time.Now().Add(-time.Duration(40-i) * 24 * time.Hour),
			UpdatedAt:    time.Now().Add(-time.Duration(rand.Intn(10)) * 24 * time.Hour),
		}
		rules = append(rules, rule)
	}

	return rules
}

// generateLoginAttempts creates 1000 login attempts across 100 clients
func generateLoginAttempts() []LoginAttempt {
	logins := make([]LoginAttempt, 0, 1000)
	rand.Seed(time.Now().UnixNano())

	countries := []string{
		"United States", "United Kingdom", "Germany", "France", "Japan", "Australia",
		"Canada", "Singapore", "Hong Kong", "Switzerland", "Netherlands", "Sweden",
		"Spain", "Italy", "Brazil", "Mexico", "India", "South Korea", "Thailand",
		"Malaysia", "Indonesia", "Philippines", "Vietnam", "South Africa", "UAE",
		"Russia", "China", "Turkey", "Poland", "Czech Republic",
	}

	cities := map[string][]string{
		"United States": {"New York", "Los Angeles", "Chicago", "Miami", "San Francisco"},
		"United Kingdom": {"London", "Manchester", "Birmingham", "Leeds", "Glasgow"},
		"Germany": {"Berlin", "Munich", "Hamburg", "Frankfurt", "Cologne"},
		"France": {"Paris", "Lyon", "Marseille", "Toulouse", "Nice"},
		"Japan": {"Tokyo", "Osaka", "Kyoto", "Yokohama", "Nagoya"},
	}

	devices := []string{"Desktop", "Mobile", "Tablet", "Unknown"}
	browsers := []string{"Chrome", "Firefox", "Safari", "Edge", "Opera"}
	statuses := []LoginStatus{LoginSuccess, LoginSuccess, LoginSuccess, LoginFailed, LoginBlocked}

	for i := 0; i < 1000; i++ {
		clientNum := i % 100
		clientID := fmt.Sprintf("CL%03d", clientNum)
		clientName := fmt.Sprintf("Client %d", clientNum)

		country := countries[rand.Intn(len(countries))]
		cityList := cities[country]
		city := country
		if cityList != nil {
			city = cityList[rand.Intn(len(cityList))]
		}

		ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
		device := devices[rand.Intn(len(devices))]
		browser := browsers[rand.Intn(len(browsers))]
		status := statuses[rand.Intn(len(statuses))]

		hoursAgo := rand.Intn(720)
		timestamp := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

		riskScore := rand.Float64() * 100
		flags := make([]string, 0)

		if riskScore > 80 {
			flags = append(flags, "high_risk")
		}
		if country == "Russia" || country == "China" {
			flags = append(flags, "high_risk_country")
			riskScore += 10
		}
		if status == LoginFailed {
			flags = append(flags, "failed_attempt")
			riskScore += 15
		}
		if rand.Float64() < 0.1 {
			flags = append(flags, "vpn_detected")
			riskScore += 20
		}

		if riskScore > 100 {
			riskScore = 100
		}

		login := LoginAttempt{
			ID:         fmt.Sprintf("LOGIN%d", 100000+i),
			ClientID:   clientID,
			ClientName: clientName,
			IP:         ip,
			Country:    country,
			City:       city,
			Device:     device,
			Browser:    browser,
			Status:     status,
			Timestamp:  timestamp,
			RiskScore:  riskScore,
			Flags:      flags,
		}

		logins = append(logins, login)
	}

	sort.Slice(logins, func(i, j int) bool {
		return logins[i].Timestamp.After(logins[j].Timestamp)
	})

	return logins
}

// generateFraudAlerts creates 50 fraud alerts based on login patterns
func generateFraudAlerts(logins []LoginAttempt) []FraudAlert {
	alerts := make([]FraudAlert, 0, 50)
	rand.Seed(time.Now().UnixNano())

	alertTypes := []AlertType{
		AlertMultipleCountries, AlertVelocityLogin, AlertImpossibleTravel,
		AlertNewDevice, AlertUnusualHours, AlertMultipleFailedLogins,
		AlertVPNDetected, AlertTorDetected,
	}

	severities := []Severity{FraudSeverityLow, FraudSeverityMedium, FraudSeverityHigh, FraudSeverityCritical}

	clientsWithAlerts := make(map[string]bool)
	alertCount := 0

	for _, login := range logins {
		if alertCount >= 50 {
			break
		}

		if login.RiskScore > 70 && rand.Float64() < 0.3 {
			if clientsWithAlerts[login.ClientID] && rand.Float64() < 0.7 {
				continue
			}

			alertType := alertTypes[rand.Intn(len(alertTypes))]
			severity := severities[rand.Intn(len(severities))]

			if login.RiskScore > 90 {
				severity = FraudSeverityCritical
			} else if login.RiskScore > 80 {
				severity = FraudSeverityHigh
			}

			description := generateAlertDescription(alertType, login)

			status := AlertStatusActive
			var resolvedAt *time.Time
			var resolvedBy string

			if rand.Float64() < 0.6 {
				status = AlertStatusResolved
				resolved := login.Timestamp.Add(time.Duration(1+rand.Intn(48)) * time.Hour)
				resolvedAt = &resolved
				resolvedBy = []string{"admin@rtx5.com", "security@rtx5.com", "manager@rtx5.com"}[rand.Intn(3)]
			}

			alert := FraudAlert{
				ID:          fmt.Sprintf("ALERT%d", 10000+alertCount),
				ClientID:    login.ClientID,
				ClientName:  login.ClientName,
				AlertType:   alertType,
				Severity:    severity,
				Description: description,
				DetectedAt:  login.Timestamp,
				ResolvedAt:  resolvedAt,
				ResolvedBy:  resolvedBy,
				Status:      status,
				Metadata: map[string]interface{}{
					"ip":         login.IP,
					"country":    login.Country,
					"risk_score": login.RiskScore,
				},
			}

			alerts = append(alerts, alert)
			clientsWithAlerts[login.ClientID] = true
			alertCount++
		}
	}

	return alerts
}

// generateAlertDescription creates a descriptive message for an alert
func generateAlertDescription(alertType AlertType, login LoginAttempt) string {
	switch alertType {
	case AlertMultipleCountries:
		return fmt.Sprintf("Login detected from %s after recent login from different country", login.Country)
	case AlertVelocityLogin:
		return fmt.Sprintf("Unusual number of login attempts from IP %s", login.IP)
	case AlertImpossibleTravel:
		return fmt.Sprintf("Impossible travel detected: Login from %s, %s", login.City, login.Country)
	case AlertNewDevice:
		return fmt.Sprintf("Login from new device: %s (%s)", login.Device, login.Browser)
	case AlertUnusualHours:
		return "Login attempt during unusual hours"
	case AlertMultipleFailedLogins:
		return fmt.Sprintf("Multiple failed login attempts from IP %s", login.IP)
	case AlertVPNDetected:
		return fmt.Sprintf("VPN detected from IP %s (%s)", login.IP, login.Country)
	case AlertTorDetected:
		return fmt.Sprintf("Tor network detected from IP %s", login.IP)
	case AlertBlacklistedCountry:
		return fmt.Sprintf("Login attempt from blacklisted country: %s", login.Country)
	default:
		return "Suspicious activity detected"
	}
}

// calculateGeoReports generates geographic distribution reports
func calculateGeoReports(logins []LoginAttempt) []GeoReport {
	countryData := make(map[string]*GeoReport)

	for _, login := range logins {
		if countryData[login.Country] == nil {
			countryData[login.Country] = &GeoReport{
				Country: login.Country,
			}
		}

		report := countryData[login.Country]
		report.LoginCount++

		if login.Status == LoginSuccess {
			report.SuccessRate += 1
		}

		if login.RiskScore > 70 {
			report.FlaggedCount++
		}
	}

	clientsPerCountry := make(map[string]map[string]bool)
	for _, login := range logins {
		if clientsPerCountry[login.Country] == nil {
			clientsPerCountry[login.Country] = make(map[string]bool)
		}
		clientsPerCountry[login.Country][login.ClientID] = true
	}

	reports := make([]GeoReport, 0)
	for country, report := range countryData {
		report.UniqueClients = len(clientsPerCountry[country])
		report.SuccessRate = (report.SuccessRate / float64(report.LoginCount)) * 100

		flaggedPct := float64(report.FlaggedCount) / float64(report.LoginCount) * 100
		if flaggedPct > 30 {
			report.RiskLevel = "high"
		} else if flaggedPct > 15 {
			report.RiskLevel = "medium"
		} else {
			report.RiskLevel = "low"
		}

		reports = append(reports, *report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].LoginCount > reports[j].LoginCount
	})

	return reports
}

// FraudDetectionHandler handles fraud detection requests
type FraudDetectionHandler struct {
	store       *FraudDetectionStore
	authService *auth.Service
}

// NewFraudDetectionHandler creates a new fraud detection handler
func NewFraudDetectionHandler(store *FraudDetectionStore, authService *auth.Service) *FraudDetectionHandler {
	return &FraudDetectionHandler{
		store:       store,
		authService: authService,
	}
}

// extractBearerToken extracts the JWT token from the Authorization header
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return authHeader
}

// HandleGetLogins handles GET /admin/fraud/logins
func (h *FraudDetectionHandler) HandleGetLogins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	country := r.URL.Query().Get("country")
	status := r.URL.Query().Get("status")
	riskFilter := r.URL.Query().Get("risk")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 100
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	filtered := make([]LoginAttempt, 0)
	for _, login := range h.store.logins {
		if country != "" && login.Country != country {
			continue
		}
		if status != "" && string(login.Status) != status {
			continue
		}
		if riskFilter == "high" && login.RiskScore < 70 {
			continue
		}
		if riskFilter == "medium" && (login.RiskScore < 40 || login.RiskScore >= 70) {
			continue
		}
		if riskFilter == "low" && login.RiskScore >= 40 {
			continue
		}
		filtered = append(filtered, login)
	}

	totalCount := len(filtered)
	start := (page - 1) * limit
	end := start + limit

	if start >= totalCount {
		filtered = []LoginAttempt{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		filtered = filtered[start:end]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    filtered,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       totalCount,
			"total_pages": int(math.Ceil(float64(totalCount) / float64(limit))),
		},
	})
}

// HandleGetClientLogins handles GET /admin/fraud/logins/client/:id
func (h *FraudDetectionHandler) HandleGetClientLogins(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}
	clientID := pathParts[5]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	clientLogins := make([]LoginAttempt, 0)
	for _, login := range h.store.logins {
		if login.ClientID == clientID {
			clientLogins = append(clientLogins, login)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"client_id": clientID,
		"logins":    clientLogins,
		"count":     len(clientLogins),
	})
}

// HandleGetAlerts handles GET /admin/fraud/alerts
func (h *FraudDetectionHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	alertType := r.URL.Query().Get("type")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	filtered := make([]FraudAlert, 0)
	for _, alert := range h.store.alerts {
		if alertType != "" && string(alert.AlertType) != alertType {
			continue
		}
		if severity != "" && string(alert.Severity) != severity {
			continue
		}
		if status != "" && string(alert.Status) != status {
			continue
		}
		filtered = append(filtered, alert)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"alerts":  filtered,
		"count":   len(filtered),
	})
}

// HandleResolveAlert handles PUT /admin/fraud/alerts/:id/resolve
func (h *FraudDetectionHandler) HandleResolveAlert(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, err := h.authService.ValidateToken(extractBearerToken(r))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Alert ID required", http.StatusBadRequest)
		return
	}
	alertID := pathParts[4]

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	var alert *FraudAlert
	for i := range h.store.alerts {
		if h.store.alerts[i].ID == alertID {
			alert = &h.store.alerts[i]
			break
		}
	}

	if alert == nil {
		http.Error(w, "Alert not found", http.StatusNotFound)
		return
	}

	now := time.Now()
	alert.ResolvedAt = &now
	alert.ResolvedBy = claims.Username
	alert.Status = AlertStatusResolved

	if req.Status == "ignored" {
		alert.Status = AlertStatusIgnored
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Alert resolved successfully",
		"alert":   alert,
	})
}

// HandleGetRules handles GET /admin/fraud/rules
func (h *FraudDetectionHandler) HandleGetRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	rules := h.store.rules
	h.store.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rules":   rules,
		"count":   len(rules),
	})
}

// HandleUpdateRule handles PUT /admin/fraud/rules/:id
func (h *FraudDetectionHandler) HandleUpdateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Rule ID required", http.StatusBadRequest)
		return
	}
	ruleID := pathParts[4]

	var req struct {
		IsActive  *bool  `json:"is_active"`
		Condition string `json:"condition"`
		Action    string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	var rule *FraudRule
	for i := range h.store.rules {
		if h.store.rules[i].ID == ruleID {
			rule = &h.store.rules[i]
			break
		}
	}

	if rule == nil {
		http.Error(w, "Rule not found", http.StatusNotFound)
		return
	}

	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Action != "" {
		rule.Action = req.Action
	}

	rule.UpdatedAt = time.Now()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Rule updated successfully",
		"rule":    rule,
	})
}

// HandleGetGeoData handles GET /admin/fraud/geo
func (h *FraudDetectionHandler) HandleGetGeoData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	geoData := h.store.geoData
	h.store.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"countries": geoData,
		"count":     len(geoData),
	})
}

// HandleGetStats handles GET /admin/fraud/stats
func (h *FraudDetectionHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateToken(extractBearerToken(r)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	stats := FraudStats{
		TotalAlerts:    len(h.store.alerts),
		AlertsByType:   make(map[string]int),
		TopRiskyCountries: make([]string, 0),
	}

	for _, alert := range h.store.alerts {
		if alert.Status == AlertStatusActive {
			stats.ActiveAlerts++
		} else if alert.Status == AlertStatusResolved {
			stats.ResolvedAlerts++
		}
		stats.AlertsByType[string(alert.AlertType)]++
	}

	if stats.TotalAlerts > 0 {
		stats.ResolutionRate = (float64(stats.ResolvedAlerts) / float64(stats.TotalAlerts)) * 100
	}

	flaggedClients := make(map[string]bool)
	totalRiskScore := 0.0
	blockedCount := 0

	for _, login := range h.store.logins {
		if login.RiskScore > 70 {
			flaggedClients[login.ClientID] = true
		}
		totalRiskScore += login.RiskScore
		if login.Status == LoginBlocked {
			blockedCount++
		}
	}

	stats.FlaggedClients = len(flaggedClients)
	stats.AvgRiskScore = totalRiskScore / float64(len(h.store.logins))
	stats.BlockedLogins = blockedCount

	riskyCountries := make([]struct {
		Country string
		Risk    float64
	}, 0)

	for _, geo := range h.store.geoData {
		if geo.RiskLevel == "high" || geo.FlaggedCount > 10 {
			riskPct := float64(geo.FlaggedCount) / float64(geo.LoginCount) * 100
			riskyCountries = append(riskyCountries, struct {
				Country string
				Risk    float64
			}{geo.Country, riskPct})
		}
	}

	sort.Slice(riskyCountries, func(i, j int) bool {
		return riskyCountries[i].Risk > riskyCountries[j].Risk
	})

	for i := 0; i < len(riskyCountries) && i < 5; i++ {
		stats.TopRiskyCountries = append(stats.TopRiskyCountries, riskyCountries[i].Country)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}
