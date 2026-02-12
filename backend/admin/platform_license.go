package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// License represents a broker's platform license
type License struct {
	ID                 int64     `json:"id"`
	BrokerID           int64     `json:"broker_id"`
	BrokerName         string    `json:"broker_name"`
	Plan               string    `json:"plan"` // starter, professional, enterprise, ultimate
	Status             string    `json:"status"` // active, expired, suspended, trial
	MaxAccounts        int       `json:"max_accounts"`
	MaxSymbols         int       `json:"max_symbols"`
	MaxConcurrentUsers int       `json:"max_concurrent_users"`
	Features           []string  `json:"features"`
	StartDate          time.Time `json:"start_date"`
	ExpiryDate         time.Time `json:"expiry_date"`
	AutoRenew          bool      `json:"auto_renew"`
	MonthlyFee         float64   `json:"monthly_fee"`
	Currency           string    `json:"currency"`
}

// FeatureEntitlement represents a platform feature and which plans include it
type FeatureEntitlement struct {
	FeatureKey  string   `json:"feature_key"`
	FeatureName string   `json:"feature_name"`
	Description string   `json:"description"`
	Plans       []string `json:"plans"` // which plans include this feature
	Enabled     bool     `json:"enabled"`
	Category    string   `json:"category"` // trading, analytics, compliance, api, white_label
}

// UsageMetric represents current usage vs limits for a broker
type UsageMetric struct {
	BrokerID    int64     `json:"broker_id"`
	Metric      string    `json:"metric"` // accounts, symbols, users, api_calls, storage_gb
	Current     float64   `json:"current"`
	Limit       float64   `json:"limit"`
	Percentage  float64   `json:"percentage"`
	LastUpdated time.Time `json:"last_updated"`
}

// BillingRecord represents a billing transaction
type BillingRecord struct {
	ID         int64     `json:"id"`
	BrokerID   int64     `json:"broker_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Period     string    `json:"period"` // e.g. "2026-01", "2026-02"
	Status     string    `json:"status"` // paid, pending, overdue, cancelled
	InvoiceURL string    `json:"invoice_url"`
	PaidAt     time.Time `json:"paid_at,omitempty"`
	DueDate    time.Time `json:"due_date"`
	CreatedAt  time.Time `json:"created_at"`
}

// LicenseStats represents overall licensing statistics
type LicenseStats struct {
	TotalLicenses      int                `json:"total_licenses"`
	ActiveLicenses     int                `json:"active_licenses"`
	TotalRevenue       float64            `json:"total_revenue"`
	MonthlyRecurring   float64            `json:"monthly_recurring_revenue"`
	PlanDistribution   map[string]int     `json:"plan_distribution"`
	StatusDistribution map[string]int     `json:"status_distribution"`
	ExpiringSoon       []License          `json:"expiring_soon"` // within 30 days
	RevenueByPlan      map[string]float64 `json:"revenue_by_plan"`
}

// UpdateLicenseRequest represents request to update a license
type UpdateLicenseRequest struct {
	Plan               string   `json:"plan"`
	Status             string   `json:"status"`
	MaxAccounts        int      `json:"max_accounts"`
	MaxSymbols         int      `json:"max_symbols"`
	MaxConcurrentUsers int      `json:"max_concurrent_users"`
	Features           []string `json:"features"`
	AutoRenew          bool     `json:"auto_renew"`
	MonthlyFee         float64  `json:"monthly_fee"`
}

// PlatformLicenseService manages platform licensing
type PlatformLicenseService struct {
	mu               sync.RWMutex
	licenses         map[int64]*License
	features         []FeatureEntitlement
	usageMetrics     map[int64][]UsageMetric // brokerId -> metrics
	billingRecords   map[int64][]BillingRecord // brokerId -> records
	licenseIDSeq     int64
	billingIDSeq     int64
}

// NewPlatformLicenseService creates a new platform license service with mock data
func NewPlatformLicenseService() *PlatformLicenseService {
	s := &PlatformLicenseService{
		licenses:       make(map[int64]*License),
		usageMetrics:   make(map[int64][]UsageMetric),
		billingRecords: make(map[int64][]BillingRecord),
		licenseIDSeq:   1,
		billingIDSeq:   1,
	}
	s.initializeMockData()
	return s
}

func (s *PlatformLicenseService) initializeMockData() {
	rand.Seed(time.Now().UnixNano())

	// Initialize 25 feature entitlements
	s.features = []FeatureEntitlement{
		// Trading features
		{FeatureKey: "basic_trading", FeatureName: "Basic Trading", Description: "Spot forex trading", Plans: []string{"starter", "professional", "enterprise", "ultimate"}, Enabled: true, Category: "trading"},
		{FeatureKey: "cfd_trading", FeatureName: "CFD Trading", Description: "CFD on stocks, indices, commodities", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "trading"},
		{FeatureKey: "crypto_trading", FeatureName: "Crypto Trading", Description: "Cryptocurrency trading", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "trading"},
		{FeatureKey: "advanced_orders", FeatureName: "Advanced Order Types", Description: "OCO, trailing stop, etc", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "trading"},
		{FeatureKey: "copy_trading", FeatureName: "Copy Trading", Description: "Social trading platform", Plans: []string{"enterprise", "ultimate"}, Enabled: true, Category: "trading"},
		{FeatureKey: "algorithmic_trading", FeatureName: "Algorithmic Trading", Description: "API for algo strategies", Plans: []string{"ultimate"}, Enabled: true, Category: "trading"},

		// Analytics features
		{FeatureKey: "basic_analytics", FeatureName: "Basic Analytics", Description: "Standard charts and reports", Plans: []string{"starter", "professional", "enterprise", "ultimate"}, Enabled: true, Category: "analytics"},
		{FeatureKey: "advanced_analytics", FeatureName: "Advanced Analytics", Description: "Custom dashboards, indicators", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "analytics"},
		{FeatureKey: "risk_analytics", FeatureName: "Risk Analytics", Description: "Real-time risk monitoring", Plans: []string{"enterprise", "ultimate"}, Enabled: true, Category: "analytics"},
		{FeatureKey: "performance_attribution", FeatureName: "Performance Attribution", Description: "Detailed P&L analysis", Plans: []string{"enterprise", "ultimate"}, Enabled: true, Category: "analytics"},
		{FeatureKey: "predictive_analytics", FeatureName: "Predictive Analytics", Description: "AI-powered market insights", Plans: []string{"ultimate"}, Enabled: true, Category: "analytics"},

		// Compliance features
		{FeatureKey: "basic_compliance", FeatureName: "Basic Compliance", Description: "KYC/AML verification", Plans: []string{"starter", "professional", "enterprise", "ultimate"}, Enabled: true, Category: "compliance"},
		{FeatureKey: "regulatory_reporting", FeatureName: "Regulatory Reporting", Description: "Automated regulatory reports", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "compliance"},
		{FeatureKey: "trade_surveillance", FeatureName: "Trade Surveillance", Description: "Market abuse detection", Plans: []string{"enterprise", "ultimate"}, Enabled: true, Category: "compliance"},
		{FeatureKey: "audit_trail", FeatureName: "Audit Trail", Description: "Complete activity logging", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "compliance"},

		// API features
		{FeatureKey: "rest_api", FeatureName: "REST API", Description: "Basic REST API access", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "api"},
		{FeatureKey: "websocket_api", FeatureName: "WebSocket API", Description: "Real-time data streaming", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "api"},
		{FeatureKey: "fix_protocol", FeatureName: "FIX Protocol", Description: "FIX 4.4 connectivity", Plans: []string{"enterprise", "ultimate"}, Enabled: true, Category: "api"},
		{FeatureKey: "unlimited_api_calls", FeatureName: "Unlimited API Calls", Description: "No rate limits", Plans: []string{"ultimate"}, Enabled: true, Category: "api"},

		// White-label features
		{FeatureKey: "basic_branding", FeatureName: "Basic Branding", Description: "Logo and colors", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "white_label"},
		{FeatureKey: "custom_domain", FeatureName: "Custom Domain", Description: "Your own domain", Plans: []string{"enterprise", "ultimate"}, Enabled: true, Category: "white_label"},
		{FeatureKey: "white_label_mobile", FeatureName: "White-Label Mobile Apps", Description: "iOS/Android apps", Plans: []string{"ultimate"}, Enabled: true, Category: "white_label"},

		// Support features
		{FeatureKey: "email_support", FeatureName: "Email Support", Description: "24h response time", Plans: []string{"starter", "professional", "enterprise", "ultimate"}, Enabled: true, Category: "support"},
		{FeatureKey: "priority_support", FeatureName: "Priority Support", Description: "4h response time", Plans: []string{"professional", "enterprise", "ultimate"}, Enabled: true, Category: "support"},
		{FeatureKey: "dedicated_account_manager", FeatureName: "Dedicated Account Manager", Description: "Personal support contact", Plans: []string{"ultimate"}, Enabled: true, Category: "support"},
	}

	// Plan configurations
	planConfigs := map[string]struct {
		maxAccounts        int
		maxSymbols         int
		maxConcurrentUsers int
		monthlyFee         float64
		features           []string
	}{
		"starter": {
			maxAccounts:        100,
			maxSymbols:         50,
			maxConcurrentUsers: 50,
			monthlyFee:         499.00,
			features:           []string{"basic_trading", "basic_analytics", "basic_compliance", "email_support"},
		},
		"professional": {
			maxAccounts:        500,
			maxSymbols:         200,
			maxConcurrentUsers: 200,
			monthlyFee:         1499.00,
			features:           []string{"basic_trading", "cfd_trading", "crypto_trading", "advanced_orders", "basic_analytics", "advanced_analytics", "basic_compliance", "regulatory_reporting", "audit_trail", "rest_api", "websocket_api", "basic_branding", "email_support", "priority_support"},
		},
		"enterprise": {
			maxAccounts:        2000,
			maxSymbols:         500,
			maxConcurrentUsers: 1000,
			monthlyFee:         4999.00,
			features:           []string{"basic_trading", "cfd_trading", "crypto_trading", "advanced_orders", "copy_trading", "basic_analytics", "advanced_analytics", "risk_analytics", "performance_attribution", "basic_compliance", "regulatory_reporting", "trade_surveillance", "audit_trail", "rest_api", "websocket_api", "fix_protocol", "basic_branding", "custom_domain", "email_support", "priority_support"},
		},
		"ultimate": {
			maxAccounts:        10000,
			maxSymbols:         1000,
			maxConcurrentUsers: 5000,
			monthlyFee:         14999.00,
			features:           []string{"basic_trading", "cfd_trading", "crypto_trading", "advanced_orders", "copy_trading", "algorithmic_trading", "basic_analytics", "advanced_analytics", "risk_analytics", "performance_attribution", "predictive_analytics", "basic_compliance", "regulatory_reporting", "trade_surveillance", "audit_trail", "rest_api", "websocket_api", "fix_protocol", "unlimited_api_calls", "basic_branding", "custom_domain", "white_label_mobile", "email_support", "priority_support", "dedicated_account_manager"},
		},
	}

	plans := []string{"starter", "professional", "professional", "enterprise", "enterprise", "enterprise", "ultimate", "ultimate", "starter", "professional"}
	statuses := []string{"active", "active", "active", "active", "active", "active", "trial", "active", "expired", "suspended"}

	// Create 10 broker licenses
	for i := 0; i < 10; i++ {
		brokerID := int64(i + 1)
		plan := plans[i]
		status := statuses[i]
		config := planConfigs[plan]

		startDate := time.Now().AddDate(0, -rand.Intn(12), -rand.Intn(30))
		expiryDate := startDate.AddDate(1, 0, 0)

		if status == "trial" {
			startDate = time.Now().AddDate(0, 0, -rand.Intn(14))
			expiryDate = startDate.AddDate(0, 0, 30)
		}

		license := &License{
			ID:                 s.licenseIDSeq,
			BrokerID:           brokerID,
			BrokerName:         fmt.Sprintf("Broker-%d", brokerID),
			Plan:               plan,
			Status:             status,
			MaxAccounts:        config.maxAccounts,
			MaxSymbols:         config.maxSymbols,
			MaxConcurrentUsers: config.maxConcurrentUsers,
			Features:           config.features,
			StartDate:          startDate,
			ExpiryDate:         expiryDate,
			AutoRenew:          status == "active" && rand.Float64() < 0.8,
			MonthlyFee:         config.monthlyFee,
			Currency:           "USD",
		}

		s.licenses[s.licenseIDSeq] = license
		s.licenseIDSeq++

		// Create usage metrics for this broker
		metrics := []UsageMetric{
			{
				BrokerID:    brokerID,
				Metric:      "accounts",
				Current:     float64(int(float64(config.maxAccounts) * (0.3 + rand.Float64()*0.6))),
				Limit:       float64(config.maxAccounts),
				LastUpdated: time.Now(),
			},
			{
				BrokerID:    brokerID,
				Metric:      "symbols",
				Current:     float64(int(float64(config.maxSymbols) * (0.4 + rand.Float64()*0.5))),
				Limit:       float64(config.maxSymbols),
				LastUpdated: time.Now(),
			},
			{
				BrokerID:    brokerID,
				Metric:      "users",
				Current:     float64(int(float64(config.maxConcurrentUsers) * (0.2 + rand.Float64()*0.6))),
				Limit:       float64(config.maxConcurrentUsers),
				LastUpdated: time.Now(),
			},
			{
				BrokerID:    brokerID,
				Metric:      "api_calls",
				Current:     float64(100000 + rand.Intn(900000)),
				Limit:       1000000,
				LastUpdated: time.Now(),
			},
			{
				BrokerID:    brokerID,
				Metric:      "storage_gb",
				Current:     10.0 + rand.Float64()*90.0,
				Limit:       100.0,
				LastUpdated: time.Now(),
			},
		}

		for i := range metrics {
			if metrics[i].Limit > 0 {
				metrics[i].Percentage = (metrics[i].Current / metrics[i].Limit) * 100
			}
		}

		s.usageMetrics[brokerID] = metrics

		// Create billing records (12 months for active brokers, fewer for others)
		numRecords := 12
		if status == "trial" {
			numRecords = 1
		} else if status == "expired" || status == "suspended" {
			numRecords = 3 + rand.Intn(6)
		}

		records := make([]BillingRecord, numRecords)
		for j := 0; j < numRecords; j++ {
			month := time.Now().AddDate(0, -j, 0)
			period := month.Format("2006-01")
			dueDate := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)

			recordStatus := "paid"
			paidAt := dueDate.AddDate(0, 0, rand.Intn(15))

			if j == 0 && (status == "suspended" || status == "expired") {
				recordStatus = "overdue"
				paidAt = time.Time{}
			} else if j == 0 && rand.Float64() < 0.1 {
				recordStatus = "pending"
				paidAt = time.Time{}
			}

			records[j] = BillingRecord{
				ID:         s.billingIDSeq,
				BrokerID:   brokerID,
				Amount:     config.monthlyFee,
				Currency:   "USD",
				Period:     period,
				Status:     recordStatus,
				InvoiceURL: fmt.Sprintf("https://invoices.rtx5.com/%d/%s.pdf", brokerID, period),
				PaidAt:     paidAt,
				DueDate:    dueDate,
				CreatedAt:  dueDate.AddDate(0, 0, -5),
			}
			s.billingIDSeq++
		}

		s.billingRecords[brokerID] = records
	}
}

// GetAllLicenses returns all broker licenses
func (s *PlatformLicenseService) GetAllLicenses() []*License {
	s.mu.RLock()
	defer s.mu.RUnlock()

	licenses := make([]*License, 0, len(s.licenses))
	for _, lic := range s.licenses {
		licenses = append(licenses, lic)
	}

	return licenses
}

// GetLicenseByID returns a specific license
func (s *PlatformLicenseService) GetLicenseByID(licenseID int64) *License {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.licenses[licenseID]
}

// UpdateLicense updates license configuration
func (s *PlatformLicenseService) UpdateLicense(licenseID int64, req UpdateLicenseRequest) (*License, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	license, exists := s.licenses[licenseID]
	if !exists {
		return nil, fmt.Errorf("license not found")
	}

	if req.Plan != "" {
		validPlans := map[string]bool{
			"starter": true, "professional": true, "enterprise": true, "ultimate": true,
		}
		if !validPlans[req.Plan] {
			return nil, fmt.Errorf("invalid plan")
		}
		license.Plan = req.Plan
	}

	if req.Status != "" {
		validStatuses := map[string]bool{
			"active": true, "expired": true, "suspended": true, "trial": true,
		}
		if !validStatuses[req.Status] {
			return nil, fmt.Errorf("invalid status")
		}
		license.Status = req.Status
	}

	if req.MaxAccounts > 0 {
		license.MaxAccounts = req.MaxAccounts
	}
	if req.MaxSymbols > 0 {
		license.MaxSymbols = req.MaxSymbols
	}
	if req.MaxConcurrentUsers > 0 {
		license.MaxConcurrentUsers = req.MaxConcurrentUsers
	}
	if len(req.Features) > 0 {
		license.Features = req.Features
	}
	if req.MonthlyFee > 0 {
		license.MonthlyFee = req.MonthlyFee
	}

	license.AutoRenew = req.AutoRenew

	return license, nil
}

// GetAllFeatures returns all feature entitlements
func (s *PlatformLicenseService) GetAllFeatures() []FeatureEntitlement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.features
}

// GetUsageMetrics returns usage metrics for a broker
func (s *PlatformLicenseService) GetUsageMetrics(brokerID int64) []UsageMetric {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.usageMetrics[brokerID]
}

// GetBillingHistory returns billing records for a broker
func (s *PlatformLicenseService) GetBillingHistory(brokerID int64) []BillingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.billingRecords[brokerID]
}

// RenewLicense manually renews a license
func (s *PlatformLicenseService) RenewLicense(licenseID int64) (*License, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	license, exists := s.licenses[licenseID]
	if !exists {
		return nil, fmt.Errorf("license not found")
	}

	// Extend expiry date by 1 year
	license.ExpiryDate = license.ExpiryDate.AddDate(1, 0, 0)
	license.Status = "active"

	// Create a new billing record
	now := time.Now()
	period := now.Format("2006-01")

	newRecord := BillingRecord{
		ID:         s.billingIDSeq,
		BrokerID:   license.BrokerID,
		Amount:     license.MonthlyFee,
		Currency:   license.Currency,
		Period:     period,
		Status:     "paid",
		InvoiceURL: fmt.Sprintf("https://invoices.rtx5.com/%d/%s.pdf", license.BrokerID, period),
		PaidAt:     now,
		DueDate:    now,
		CreatedAt:  now,
	}
	s.billingIDSeq++

	s.billingRecords[license.BrokerID] = append(s.billingRecords[license.BrokerID], newRecord)

	return license, nil
}

// GetLicenseStats returns overall licensing statistics
func (s *PlatformLicenseService) GetLicenseStats() *LicenseStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &LicenseStats{
		PlanDistribution:   make(map[string]int),
		StatusDistribution: make(map[string]int),
		RevenueByPlan:      make(map[string]float64),
		ExpiringSoon:       make([]License, 0),
	}

	totalRevenue := 0.0
	monthlyRecurring := 0.0
	activeCount := 0

	thirtyDaysFromNow := time.Now().AddDate(0, 0, 30)

	for _, lic := range s.licenses {
		stats.TotalLicenses++
		stats.PlanDistribution[lic.Plan]++
		stats.StatusDistribution[lic.Status]++

		if lic.Status == "active" {
			activeCount++
			monthlyRecurring += lic.MonthlyFee
			stats.RevenueByPlan[lic.Plan] += lic.MonthlyFee
		}

		// Check if expiring soon
		if lic.Status == "active" && lic.ExpiryDate.Before(thirtyDaysFromNow) {
			stats.ExpiringSoon = append(stats.ExpiringSoon, *lic)
		}
	}

	stats.ActiveLicenses = activeCount
	stats.MonthlyRecurring = math.Round(monthlyRecurring*100) / 100

	// Calculate total revenue from billing records
	for _, records := range s.billingRecords {
		for _, record := range records {
			if record.Status == "paid" {
				totalRevenue += record.Amount
			}
		}
	}

	stats.TotalRevenue = math.Round(totalRevenue*100) / 100

	// Round revenue by plan
	for plan, revenue := range stats.RevenueByPlan {
		stats.RevenueByPlan[plan] = math.Round(revenue*100) / 100
	}

	return stats
}

// PlatformLicenseHandler handles HTTP requests for platform license API
type PlatformLicenseHandler struct {
	service     *PlatformLicenseService
	authService *AuthService
}

// NewPlatformLicenseHandler creates a new platform license handler
func NewPlatformLicenseHandler(service *PlatformLicenseService, authService *AuthService) *PlatformLicenseHandler {
	return &PlatformLicenseHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetAllLicenses returns all broker licenses
func (h *PlatformLicenseHandler) HandleGetAllLicenses(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	licenses := h.service.GetAllLicenses()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(licenses)
}

// HandleGetLicenseByID returns license detail with usage metrics
func (h *PlatformLicenseHandler) HandleGetLicenseByID(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/admin/licenses/"):]
	if slashIdx := strings.Index(idStr, "/"); slashIdx != -1 {
		idStr = idStr[:slashIdx]
	}

	licenseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid license ID", http.StatusBadRequest)
		return
	}

	license := h.service.GetLicenseByID(licenseID)
	if license == nil {
		http.Error(w, "License not found", http.StatusNotFound)
		return
	}

	usage := h.service.GetUsageMetrics(license.BrokerID)

	response := map[string]interface{}{
		"license": license,
		"usage":   usage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateLicense updates license configuration
func (h *PlatformLicenseHandler) HandleUpdateLicense(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Path[len("/admin/licenses/"):]
	if slashIdx := strings.Index(idStr, "/"); slashIdx != -1 {
		idStr = idStr[:slashIdx]
	}

	licenseID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid license ID", http.StatusBadRequest)
		return
	}

	var req UpdateLicenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	license, err := h.service.UpdateLicense(licenseID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(license)
}

// HandleGetAllFeatures returns all available features by plan
func (h *PlatformLicenseHandler) HandleGetAllFeatures(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	features := h.service.GetAllFeatures()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(features)
}

// HandleGetUsageMetrics returns current usage vs limits for broker
func (h *PlatformLicenseHandler) HandleGetUsageMetrics(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	licenseID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid license ID", http.StatusBadRequest)
		return
	}

	license := h.service.GetLicenseByID(licenseID)
	if license == nil {
		http.Error(w, "License not found", http.StatusNotFound)
		return
	}

	usage := h.service.GetUsageMetrics(license.BrokerID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(usage)
}

// HandleGetBillingHistory returns billing history for broker
func (h *PlatformLicenseHandler) HandleGetBillingHistory(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	licenseID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid license ID", http.StatusBadRequest)
		return
	}

	license := h.service.GetLicenseByID(licenseID)
	if license == nil {
		http.Error(w, "License not found", http.StatusNotFound)
		return
	}

	billing := h.service.GetBillingHistory(license.BrokerID)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(billing)
}

// HandleRenewLicense manually renews a license
func (h *PlatformLicenseHandler) HandleRenewLicense(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	licenseID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid license ID", http.StatusBadRequest)
		return
	}

	license, err := h.service.RenewLicense(licenseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("License renewed successfully until %s", license.ExpiryDate.Format("2006-01-02")),
		"license": license,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleGetLicenseStats returns overall licensing statistics
func (h *PlatformLicenseHandler) HandleGetLicenseStats(w http.ResponseWriter, r *http.Request) {
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetLicenseStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
