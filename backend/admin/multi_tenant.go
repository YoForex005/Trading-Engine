package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Tenant represents a white-label broker instance
type Tenant struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	CompanyName      string         `json:"company_name"`
	Status           string         `json:"status"`
	Domain           string         `json:"domain"`
	Branding         TenantBranding `json:"branding"`
	Config           TenantConfig   `json:"config"`
	Quotas           TenantQuotas   `json:"quotas"`
	CurrentUsage     TenantUsage    `json:"current_usage"`
	ContactEmail     string         `json:"contact_email"`
	ContactPhone     string         `json:"contact_phone"`
	AdminUser        string         `json:"admin_user"`
	CreatedAt        time.Time      `json:"created_at"`
	ActivatedAt      *time.Time     `json:"activated_at,omitempty"`
	TrialExpiresAt   *time.Time     `json:"trial_expires_at,omitempty"`
	LastActivityAt   time.Time      `json:"last_activity_at"`
	IsolationScore   float64        `json:"isolation_score"`
	ComplianceScore  float64        `json:"compliance_score"`
}

// TenantBranding represents custom branding configuration
type TenantBranding struct {
	LogoURL         string            `json:"logo_url"`
	FaviconURL      string            `json:"favicon_url"`
	PrimaryColor    string            `json:"primary_color"`
	SecondaryColor  string            `json:"secondary_color"`
	AccentColor     string            `json:"accent_color"`
	BackgroundColor string            `json:"background_color"`
	FontFamily      string            `json:"font_family"`
	CustomCSS       string            `json:"custom_css,omitempty"`
	EmailHeader     string            `json:"email_header"`
	EmailFooter     string            `json:"email_footer"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// TenantConfig represents tenant-specific configuration
type TenantConfig struct {
	CustomDomain       string   `json:"custom_domain"`
	SupportedLanguages []string `json:"supported_languages"`
	DefaultLanguage    string   `json:"default_language"`
	Timezone           string   `json:"timezone"`
	Currency           string   `json:"currency"`
	EmailFromName      string   `json:"email_from_name"`
	EmailFromAddress   string   `json:"email_from_address"`
	SMTPConfigured     bool     `json:"smtp_configured"`
	TwoFactorEnabled   bool     `json:"two_factor_enabled"`
	KYCRequired        bool     `json:"kyc_required"`
	MinDeposit         float64  `json:"min_deposit"`
	MaxLeverage        int      `json:"max_leverage"`
	AllowedCountries   []string `json:"allowed_countries,omitempty"`
	BlockedCountries   []string `json:"blocked_countries,omitempty"`
}

// TenantQuotas represents resource quotas for a tenant
type TenantQuotas struct {
	MaxAccounts            int `json:"max_accounts"`
	MaxSymbols             int `json:"max_symbols"`
	MaxConcurrentConns     int `json:"max_concurrent_connections"`
	MaxAPICallsPerMin      int `json:"max_api_calls_per_min"`
	MaxStorageGB           int `json:"max_storage_gb"`
	MaxAdminUsers          int `json:"max_admin_users"`
	MaxMonthlyVolumeLots   int `json:"max_monthly_volume_lots"`
}

// TenantUsage represents current resource usage
type TenantUsage struct {
	CurrentAccounts       int     `json:"current_accounts"`
	CurrentSymbols        int     `json:"current_symbols"`
	CurrentConns          int     `json:"current_connections"`
	APICallsToday         int     `json:"api_calls_today"`
	StorageUsedGB         float64 `json:"storage_used_gb"`
	AdminUsersCount       int     `json:"admin_users_count"`
	MonthlyVolumeLots     int     `json:"monthly_volume_lots"`
	AccountUtilization    float64 `json:"account_utilization_pct"`
	SymbolUtilization     float64 `json:"symbol_utilization_pct"`
	ConnectionUtilization float64 `json:"connection_utilization_pct"`
}

// TenantActivityLog represents an activity log entry
type TenantActivityLog struct {
	ID          int64     `json:"id"`
	TenantID    int64     `json:"tenant_id"`
	TenantName  string    `json:"tenant_name"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	PerformedBy string    `json:"performed_by"`
	IPAddress   string    `json:"ip_address"`
	Timestamp   time.Time `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MultiTenantService manages white-label broker instances
type MultiTenantService struct {
	mu            sync.RWMutex
	tenants       map[int64]*Tenant
	activityLogs  []*TenantActivityLog
	nextTenantID  int64
	nextLogID     int64
}

// NewMultiTenantService creates a new multi-tenant service
func NewMultiTenantService() *MultiTenantService {
	s := &MultiTenantService{
		tenants:      make(map[int64]*Tenant),
		activityLogs: make([]*TenantActivityLog, 0),
		nextTenantID: 1,
		nextLogID:    1,
	}
	s.initMockData()
	return s
}

func (s *MultiTenantService) initMockData() {
	// Generate 8 tenant instances
	tenants := []struct {
		name        string
		companyName string
		status      string
		domain      string
		primaryColor string
		languages   []string
		currency    string
	}{
		{"FXPrime", "FX Prime Limited", "active", "fxprime.rtx5.com", "#1E40AF", []string{"en", "es", "pt"}, "USD"},
		{"GlobalTraders", "Global Traders Inc", "active", "globaltraders.rtx5.com", "#059669", []string{"en", "zh", "ja"}, "USD"},
		{"EuroCapital", "Euro Capital Partners", "active", "eurocapital.rtx5.com", "#7C3AED", []string{"en", "de", "fr", "it"}, "EUR"},
		{"AsiaFX", "Asia FX Group", "trial", "asiafx.rtx5.com", "#DC2626", []string{"en", "zh", "ko", "ja"}, "USD"},
		{"SwissBroker", "Swiss Broker AG", "active", "swissbroker.rtx5.com", "#0891B2", []string{"en", "de", "fr"}, "CHF"},
		{"LatamTrade", "Latam Trade Solutions", "suspended", "latamtrade.rtx5.com", "#EA580C", []string{"es", "pt", "en"}, "USD"},
		{"MiddleEastFX", "Middle East FX", "pending_setup", "mefx.rtx5.com", "#16A34A", []string{"en", "ar"}, "USD"},
		{"AfricaCapital", "Africa Capital Markets", "decommissioned", "africacapital.rtx5.com", "#9333EA", []string{"en", "fr"}, "USD"},
	}

	statuses := []string{"active", "active", "active", "trial", "active", "suspended", "pending_setup", "decommissioned"}

	for i, tenant := range tenants {
		status := statuses[i]

		// Generate quotas based on status
		var quotas TenantQuotas
		switch status {
		case "trial":
			quotas = TenantQuotas{
				MaxAccounts:           100,
				MaxSymbols:            50,
				MaxConcurrentConns:    100,
				MaxAPICallsPerMin:     500,
				MaxStorageGB:          10,
				MaxAdminUsers:         2,
				MaxMonthlyVolumeLots:  10000,
			}
		case "active":
			quotas = TenantQuotas{
				MaxAccounts:           5000 + rand.Intn(10000),
				MaxSymbols:            200 + rand.Intn(300),
				MaxConcurrentConns:    1000 + rand.Intn(4000),
				MaxAPICallsPerMin:     5000 + rand.Intn(10000),
				MaxStorageGB:          100 + rand.Intn(400),
				MaxAdminUsers:         10 + rand.Intn(40),
				MaxMonthlyVolumeLots:  100000 + rand.Intn(400000),
			}
		default:
			quotas = TenantQuotas{
				MaxAccounts:           1000,
				MaxSymbols:            100,
				MaxConcurrentConns:    500,
				MaxAPICallsPerMin:     1000,
				MaxStorageGB:          50,
				MaxAdminUsers:         5,
				MaxMonthlyVolumeLots:  50000,
			}
		}

		// Generate current usage
		usage := TenantUsage{
			CurrentAccounts:   int(float64(quotas.MaxAccounts) * (0.3 + rand.Float64()*0.5)),
			CurrentSymbols:    int(float64(quotas.MaxSymbols) * (0.5 + rand.Float64()*0.4)),
			CurrentConns:      int(float64(quotas.MaxConcurrentConns) * (0.2 + rand.Float64()*0.3)),
			APICallsToday:     rand.Intn(quotas.MaxAPICallsPerMin * 1000),
			StorageUsedGB:     float64(quotas.MaxStorageGB) * (0.3 + rand.Float64()*0.4),
			AdminUsersCount:   rand.Intn(quotas.MaxAdminUsers) + 1,
			MonthlyVolumeLots: int(float64(quotas.MaxMonthlyVolumeLots) * (0.4 + rand.Float64()*0.5)),
		}
		usage.AccountUtilization = float64(usage.CurrentAccounts) / float64(quotas.MaxAccounts) * 100
		usage.SymbolUtilization = float64(usage.CurrentSymbols) / float64(quotas.MaxSymbols) * 100
		usage.ConnectionUtilization = float64(usage.CurrentConns) / float64(quotas.MaxConcurrentConns) * 100

		createdAt := time.Now().AddDate(0, 0, -rand.Intn(365))
		var activatedAt, trialExpiresAt *time.Time

		if status == "active" {
			activated := createdAt.Add(time.Duration(rand.Intn(72)) * time.Hour)
			activatedAt = &activated
		} else if status == "trial" {
			activated := createdAt.Add(time.Duration(rand.Intn(24)) * time.Hour)
			activatedAt = &activated
			trialExpiry := time.Now().AddDate(0, 0, rand.Intn(30))
			trialExpiresAt = &trialExpiry
		}

		t := &Tenant{
			ID:          s.nextTenantID,
			Name:        tenant.name,
			CompanyName: tenant.companyName,
			Status:      status,
			Domain:      tenant.domain,
			Branding: TenantBranding{
				LogoURL:         fmt.Sprintf("https://%s/assets/logo.png", tenant.domain),
				FaviconURL:      fmt.Sprintf("https://%s/assets/favicon.ico", tenant.domain),
				PrimaryColor:    tenant.primaryColor,
				SecondaryColor:  "#6B7280",
				AccentColor:     "#F59E0B",
				BackgroundColor: "#F9FAFB",
				FontFamily:      "Inter, sans-serif",
				EmailHeader:     fmt.Sprintf("https://%s/assets/email-header.png", tenant.domain),
				EmailFooter:     fmt.Sprintf("© 2026 %s. All rights reserved.", tenant.companyName),
			},
			Config: TenantConfig{
				CustomDomain:       tenant.domain,
				SupportedLanguages: tenant.languages,
				DefaultLanguage:    tenant.languages[0],
				Timezone:           "UTC",
				Currency:           tenant.currency,
				EmailFromName:      tenant.companyName,
				EmailFromAddress:   fmt.Sprintf("noreply@%s", tenant.domain),
				SMTPConfigured:     status == "active",
				TwoFactorEnabled:   status == "active",
				KYCRequired:        true,
				MinDeposit:         100.0 + rand.Float64()*400.0,
				MaxLeverage:        []int{100, 200, 500, 1000}[rand.Intn(4)],
			},
			Quotas:           quotas,
			CurrentUsage:     usage,
			ContactEmail:     fmt.Sprintf("admin@%s", tenant.domain),
			ContactPhone:     fmt.Sprintf("+1-555-%04d", rand.Intn(10000)),
			AdminUser:        fmt.Sprintf("admin_%s", strings.ToLower(tenant.name)),
			CreatedAt:        createdAt,
			ActivatedAt:      activatedAt,
			TrialExpiresAt:   trialExpiresAt,
			LastActivityAt:   time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour),
			IsolationScore:   95.0 + rand.Float64()*4.9,
			ComplianceScore:  90.0 + rand.Float64()*9.9,
		}

		s.tenants[s.nextTenantID] = t
		s.nextTenantID++
	}

	// Generate 15 tenant activity logs
	actions := []string{
		"tenant_created", "tenant_activated", "tenant_suspended", "tenant_reactivated",
		"config_updated", "branding_updated", "quotas_increased", "quotas_decreased",
		"admin_user_added", "admin_user_removed", "domain_changed", "status_changed",
		"trial_extended", "tenant_decommissioned", "isolation_verified",
	}

	descriptions := map[string]string{
		"tenant_created":         "New tenant instance created",
		"tenant_activated":       "Tenant activated and ready for use",
		"tenant_suspended":       "Tenant suspended due to policy violation",
		"tenant_reactivated":     "Tenant reactivated after suspension",
		"config_updated":         "Tenant configuration updated",
		"branding_updated":       "Custom branding applied",
		"quotas_increased":       "Resource quotas increased",
		"quotas_decreased":       "Resource quotas decreased",
		"admin_user_added":       "New admin user added",
		"admin_user_removed":     "Admin user removed",
		"domain_changed":         "Custom domain updated",
		"status_changed":         "Tenant status changed",
		"trial_extended":         "Trial period extended",
		"tenant_decommissioned":  "Tenant instance decommissioned",
		"isolation_verified":     "Tenant isolation verification completed",
	}

	for i := 0; i < 15; i++ {
		tenantID := int64(rand.Intn(8) + 1)
		tenant := s.tenants[tenantID]
		action := actions[rand.Intn(len(actions))]

		log := &TenantActivityLog{
			ID:          s.nextLogID,
			TenantID:    tenantID,
			TenantName:  tenant.Name,
			Action:      action,
			Description: descriptions[action],
			PerformedBy: fmt.Sprintf("admin_user_%d", rand.Intn(5)+1),
			IPAddress:   fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256)),
			Timestamp:   time.Now().Add(-time.Duration(rand.Intn(720)) * time.Hour),
		}

		s.activityLogs = append(s.activityLogs, log)
		s.nextLogID++
	}

	log.Println("[MultiTenant] Mock data initialized: 8 tenant instances, 15 activity logs, isolation metrics")
}

// GetTenants returns all tenant instances
func (s *MultiTenantService) GetTenants() []*Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenants := make([]*Tenant, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		tenants = append(tenants, tenant)
	}
	return tenants
}

// GetTenant returns a specific tenant by ID
func (s *MultiTenantService) GetTenant(id int64) (*Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenant, exists := s.tenants[id]
	if !exists {
		return nil, fmt.Errorf("tenant not found")
	}
	return tenant, nil
}

// CreateTenant creates a new tenant instance
func (s *MultiTenantService) CreateTenant(data map[string]interface{}) (*Tenant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := data["name"].(string)
	companyName := data["company_name"].(string)
	domain := data["domain"].(string)

	tenant := &Tenant{
		ID:          s.nextTenantID,
		Name:        name,
		CompanyName: companyName,
		Status:      "pending_setup",
		Domain:      domain,
		Branding: TenantBranding{
			LogoURL:         fmt.Sprintf("https://%s/assets/logo.png", domain),
			PrimaryColor:    "#1E40AF",
			SecondaryColor:  "#6B7280",
			AccentColor:     "#F59E0B",
			BackgroundColor: "#F9FAFB",
			FontFamily:      "Inter, sans-serif",
		},
		Config: TenantConfig{
			CustomDomain:       domain,
			SupportedLanguages: []string{"en"},
			DefaultLanguage:    "en",
			Timezone:           "UTC",
			Currency:           "USD",
			EmailFromName:      companyName,
			EmailFromAddress:   fmt.Sprintf("noreply@%s", domain),
			KYCRequired:        true,
			MinDeposit:         100.0,
			MaxLeverage:        100,
		},
		Quotas: TenantQuotas{
			MaxAccounts:           1000,
			MaxSymbols:            100,
			MaxConcurrentConns:    500,
			MaxAPICallsPerMin:     1000,
			MaxStorageGB:          50,
			MaxAdminUsers:         5,
			MaxMonthlyVolumeLots:  50000,
		},
		CurrentUsage:    TenantUsage{},
		ContactEmail:    fmt.Sprintf("admin@%s", domain),
		AdminUser:       fmt.Sprintf("admin_%s", strings.ToLower(name)),
		CreatedAt:       time.Now(),
		LastActivityAt:  time.Now(),
		IsolationScore:  100.0,
		ComplianceScore: 100.0,
	}

	s.tenants[s.nextTenantID] = tenant

	// Create activity log
	logEntry := &TenantActivityLog{
		ID:          s.nextLogID,
		TenantID:    s.nextTenantID,
		TenantName:  name,
		Action:      "tenant_created",
		Description: "New tenant instance created",
		PerformedBy: "system_admin",
		IPAddress:   "192.168.1.1",
		Timestamp:   time.Now(),
	}
	s.activityLogs = append(s.activityLogs, logEntry)
	s.nextLogID++

	s.nextTenantID++
	return tenant, nil
}

// UpdateTenant updates tenant configuration
func (s *MultiTenantService) UpdateTenant(id int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tenant, exists := s.tenants[id]
	if !exists {
		return fmt.Errorf("tenant not found")
	}

	// Update branding
	if branding, ok := updates["branding"].(map[string]interface{}); ok {
		if logo, ok := branding["logo_url"].(string); ok {
			tenant.Branding.LogoURL = logo
		}
		if primaryColor, ok := branding["primary_color"].(string); ok {
			tenant.Branding.PrimaryColor = primaryColor
		}
	}

	// Update config
	if config, ok := updates["config"].(map[string]interface{}); ok {
		if languages, ok := config["supported_languages"].([]interface{}); ok {
			langs := make([]string, len(languages))
			for i, lang := range languages {
				langs[i] = lang.(string)
			}
			tenant.Config.SupportedLanguages = langs
		}
	}

	tenant.LastActivityAt = time.Now()

	// Create activity log
	logEntry := &TenantActivityLog{
		ID:          s.nextLogID,
		TenantID:    id,
		TenantName:  tenant.Name,
		Action:      "config_updated",
		Description: "Tenant configuration updated",
		PerformedBy: "admin_user",
		Timestamp:   time.Now(),
	}
	s.activityLogs = append(s.activityLogs, logEntry)
	s.nextLogID++

	return nil
}

// UpdateTenantStatus updates tenant status
func (s *MultiTenantService) UpdateTenantStatus(id int64, newStatus string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tenant, exists := s.tenants[id]
	if !exists {
		return fmt.Errorf("tenant not found")
	}

	oldStatus := tenant.Status
	tenant.Status = newStatus
	tenant.LastActivityAt = time.Now()

	if newStatus == "active" && tenant.ActivatedAt == nil {
		now := time.Now()
		tenant.ActivatedAt = &now
	}

	// Create activity log
	logEntry := &TenantActivityLog{
		ID:          s.nextLogID,
		TenantID:    id,
		TenantName:  tenant.Name,
		Action:      "status_changed",
		Description: fmt.Sprintf("Status changed from %s to %s", oldStatus, newStatus),
		PerformedBy: "admin_user",
		Timestamp:   time.Now(),
	}
	s.activityLogs = append(s.activityLogs, logEntry)
	s.nextLogID++

	return nil
}

// GetTenantUsage returns tenant resource usage
func (s *MultiTenantService) GetTenantUsage(id int64) (*TenantUsage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenant, exists := s.tenants[id]
	if !exists {
		return nil, fmt.Errorf("tenant not found")
	}
	return &tenant.CurrentUsage, nil
}

// GetTenantLogs returns activity logs for a tenant
func (s *MultiTenantService) GetTenantLogs(id int64) []*TenantActivityLog {
	s.mu.RLock()
	defer s.mu.RUnlock()

	logs := make([]*TenantActivityLog, 0)
	for _, log := range s.activityLogs {
		if log.TenantID == id {
			logs = append(logs, log)
		}
	}
	return logs
}

// GetStats returns platform-wide tenancy statistics
func (s *MultiTenantService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statusCounts := make(map[string]int)
	totalAccounts := 0
	totalSymbols := 0
	totalConns := 0
	avgIsolation := 0.0
	avgCompliance := 0.0

	for _, tenant := range s.tenants {
		statusCounts[tenant.Status]++
		totalAccounts += tenant.CurrentUsage.CurrentAccounts
		totalSymbols += tenant.CurrentUsage.CurrentSymbols
		totalConns += tenant.CurrentUsage.CurrentConns
		avgIsolation += tenant.IsolationScore
		avgCompliance += tenant.ComplianceScore
	}

	tenantCount := len(s.tenants)
	if tenantCount > 0 {
		avgIsolation /= float64(tenantCount)
		avgCompliance /= float64(tenantCount)
	}

	return map[string]interface{}{
		"total_tenants":      tenantCount,
		"status_breakdown":   statusCounts,
		"total_accounts":     totalAccounts,
		"total_symbols":      totalSymbols,
		"total_connections":  totalConns,
		"avg_isolation_score": avgIsolation,
		"avg_compliance_score": avgCompliance,
		"total_activity_logs": len(s.activityLogs),
	}
}

// MultiTenantHandler handles HTTP requests for multi-tenancy
type MultiTenantHandler struct {
	service     *MultiTenantService
	authService *AuthService
}

// NewMultiTenantHandler creates a new multi-tenant handler
func NewMultiTenantHandler(service *MultiTenantService, authService *AuthService) *MultiTenantHandler {
	return &MultiTenantHandler{
		service:     service,
		authService: authService,
	}
}

// authenticateAdmin validates the admin session from the request
func (h *MultiTenantHandler) authenticateAdmin(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header")
	}

	sessionID := parts[1]
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return h.authService.ValidateSession(sessionID, ip)
}

// HandleListTenants handles GET /admin/tenants
func (h *MultiTenantHandler) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenants := h.service.GetTenants()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenants": tenants,
		"total":   len(tenants),
	})
}

// HandleGetTenant handles GET /admin/tenants/:id
func (h *MultiTenantHandler) HandleGetTenant(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
	// Remove any path suffix like /usage, /logs, /status
	idStr = strings.Split(idStr, "/")[0]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	tenant, err := h.service.GetTenant(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(tenant)
}

// HandleCreateTenant handles POST /admin/tenants
func (h *MultiTenantHandler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tenant, err := h.service.CreateTenant(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tenant)
}

// HandleUpdateTenant handles PUT /admin/tenants/:id
func (h *MultiTenantHandler) HandleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTenant(id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant updated successfully",
	})
}

// HandleUpdateTenantStatus handles PUT /admin/tenants/:id/status
func (h *MultiTenantHandler) HandleUpdateTenantStatus(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "status" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newStatus, ok := data["status"].(string)
	if !ok {
		http.Error(w, "Status field required", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTenantStatus(id, newStatus); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant status updated successfully",
	})
}

// HandleGetTenantUsage handles GET /admin/tenants/:id/usage
func (h *MultiTenantHandler) HandleGetTenantUsage(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "usage" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	usage, err := h.service.GetTenantUsage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(usage)
}

// HandleGetTenantLogs handles GET /admin/tenants/:id/logs
func (h *MultiTenantHandler) HandleGetTenantLogs(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "logs" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	logs := h.service.GetTenantLogs(id)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  logs,
		"total": len(logs),
	})
}

// HandleGetStats handles GET /admin/tenants/stats
func (h *MultiTenantHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authenticateAdmin(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
