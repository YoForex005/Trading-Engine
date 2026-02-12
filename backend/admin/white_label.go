package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// WhiteLabelConfig represents a brand configuration for white-label deployment
type WhiteLabelConfig struct {
	ID              int64      `json:"id"`
	BrandName       string     `json:"brandName"`
	PlatformName    string     `json:"platformName"`
	LogoURL         string     `json:"logoUrl"`
	FaviconURL      string     `json:"faviconUrl"`
	PrimaryColor    string     `json:"primaryColor"`
	SecondaryColor  string     `json:"secondaryColor"`
	AccentColor     string     `json:"accentColor"`
	FontFamily      string     `json:"fontFamily"`
	CustomCSS       string     `json:"customCss"`
	CustomDomain    string     `json:"customDomain"`
	EmailSender     string     `json:"emailSender"`
	LoginPageText   string     `json:"loginPageText"`
	FooterText      string     `json:"footerText"`
	SupportEmail    string     `json:"supportEmail"`
	SupportPhone    string     `json:"supportPhone"`
	SocialLinks     SocialLinks `json:"socialLinks"`
	Features        BrandFeatures `json:"features"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"deletedAt,omitempty"`
	IsActive        bool       `json:"isActive"`
}

// SocialLinks represents social media links for the brand
type SocialLinks struct {
	Facebook  string `json:"facebook,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	LinkedIn  string `json:"linkedin,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	YouTube   string `json:"youtube,omitempty"`
}

// BrandFeatures represents feature toggles for the brand
type BrandFeatures struct {
	ShowLiveChat       bool `json:"showLiveChat"`
	ShowNewsFeed       bool `json:"showNewsFeed"`
	ShowEconomicCalendar bool `json:"showEconomicCalendar"`
	ShowTradingSignals bool `json:"showTradingSignals"`
	AllowDemoAccounts  bool `json:"allowDemoAccounts"`
	AllowCopyTrading   bool `json:"allowCopyTrading"`
}

// WhiteLabelStore manages white-label brand configurations
type WhiteLabelStore struct {
	configs map[int64]*WhiteLabelConfig
	mu      sync.RWMutex
	nextID  int64
}

// WhiteLabelService provides business logic for white-label management
type WhiteLabelService struct {
	store *WhiteLabelStore
}

// NewWhiteLabelService creates a new white-label service with default configs
func NewWhiteLabelService() *WhiteLabelService {
	store := &WhiteLabelStore{
		configs: make(map[int64]*WhiteLabelConfig),
		nextID:  1,
	}

	// Initialize with 5 default brand configurations
	defaultConfigs := []*WhiteLabelConfig{
		{
			ID:           1,
			BrandName:    "RTX5 Default",
			PlatformName: "RTX5 Trading Platform",
			LogoURL:      "/assets/logos/rtx5-logo.svg",
			FaviconURL:   "/assets/logos/rtx5-favicon.ico",
			PrimaryColor: "#1a73e8",
			SecondaryColor: "#34a853",
			AccentColor:  "#fbbc04",
			FontFamily:   "Inter, system-ui, sans-serif",
			CustomCSS:    "",
			CustomDomain: "trade.rtx5.com",
			EmailSender:  "noreply@rtx5.com",
			LoginPageText: "Welcome to RTX5 - Your Premier Trading Platform",
			FooterText:   "© 2026 RTX5. All rights reserved. Risk Warning: Trading involves substantial risk.",
			SupportEmail: "support@rtx5.com",
			SupportPhone: "+1-800-RTX-5000",
			SocialLinks: SocialLinks{
				Facebook: "https://facebook.com/rtx5trading",
				Twitter:  "https://twitter.com/rtx5official",
				LinkedIn: "https://linkedin.com/company/rtx5",
			},
			Features: BrandFeatures{
				ShowLiveChat:         true,
				ShowNewsFeed:         true,
				ShowEconomicCalendar: true,
				ShowTradingSignals:   true,
				AllowDemoAccounts:    true,
				AllowCopyTrading:     true,
			},
			CreatedAt: time.Now().Add(-90 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
			IsActive:  true,
		},
		{
			ID:           2,
			BrandName:    "ProTrader Elite",
			PlatformName: "ProTrader Elite",
			LogoURL:      "/assets/logos/protrader-logo.svg",
			FaviconURL:   "/assets/logos/protrader-favicon.ico",
			PrimaryColor: "#2c3e50",
			SecondaryColor: "#e74c3c",
			AccentColor:  "#f39c12",
			FontFamily:   "Roboto, Arial, sans-serif",
			CustomCSS:    ".header { background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%); }",
			CustomDomain: "app.protraderelite.com",
			EmailSender:  "notifications@protraderelite.com",
			LoginPageText: "Elite Trading Solutions for Professional Traders",
			FooterText:   "ProTrader Elite - Licensed and Regulated. CFDs are complex instruments.",
			SupportEmail: "help@protraderelite.com",
			SupportPhone: "+44-20-7946-0958",
			SocialLinks: SocialLinks{
				Facebook: "https://facebook.com/protraderelite",
				Twitter:  "https://twitter.com/protraderelite",
				LinkedIn: "https://linkedin.com/company/protrader-elite",
			},
			Features: BrandFeatures{
				ShowLiveChat:         true,
				ShowNewsFeed:         true,
				ShowEconomicCalendar: true,
				ShowTradingSignals:   false,
				AllowDemoAccounts:    true,
				AllowCopyTrading:     false,
			},
			CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-5 * 24 * time.Hour),
			IsActive:  true,
		},
		{
			ID:           3,
			BrandName:    "ForexMaster",
			PlatformName: "ForexMaster Platform",
			LogoURL:      "/assets/logos/forexmaster-logo.svg",
			FaviconURL:   "/assets/logos/forexmaster-favicon.ico",
			PrimaryColor: "#16a085",
			SecondaryColor: "#27ae60",
			AccentColor:  "#f1c40f",
			FontFamily:   "Poppins, sans-serif",
			CustomCSS:    ".trading-panel { border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); }",
			CustomDomain: "platform.forexmaster.io",
			EmailSender:  "info@forexmaster.io",
			LoginPageText: "Master the Forex Markets - Trade with Confidence",
			FooterText:   "© 2026 ForexMaster. Trading forex carries a high level of risk.",
			SupportEmail: "support@forexmaster.io",
			SupportPhone: "+1-888-367-3962",
			SocialLinks: SocialLinks{
				Facebook:  "https://facebook.com/forexmasterio",
				Twitter:   "https://twitter.com/forexmaster",
				Instagram: "https://instagram.com/forexmaster",
				YouTube:   "https://youtube.com/c/forexmaster",
			},
			Features: BrandFeatures{
				ShowLiveChat:         true,
				ShowNewsFeed:         true,
				ShowEconomicCalendar: true,
				ShowTradingSignals:   true,
				AllowDemoAccounts:    true,
				AllowCopyTrading:     true,
			},
			CreatedAt: time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
			IsActive:  true,
		},
		{
			ID:           4,
			BrandName:    "GlobalTrade Partners",
			PlatformName: "GTP Trading Suite",
			LogoURL:      "/assets/logos/gtp-logo.svg",
			FaviconURL:   "/assets/logos/gtp-favicon.ico",
			PrimaryColor: "#8e44ad",
			SecondaryColor: "#3498db",
			AccentColor:  "#e67e22",
			FontFamily:   "Open Sans, Helvetica, sans-serif",
			CustomCSS:    ":root { --sidebar-width: 280px; } .chart-container { background: #1e1e2e; }",
			CustomDomain: "trading.globaltradepartners.com",
			EmailSender:  "no-reply@globaltradepartners.com",
			LoginPageText: "Global Trading Solutions - Your Gateway to Financial Markets",
			FooterText:   "GlobalTrade Partners Ltd. Authorized and Regulated by FCA.",
			SupportEmail: "clientservices@globaltradepartners.com",
			SupportPhone: "+44-20-3287-6543",
			SocialLinks: SocialLinks{
				LinkedIn: "https://linkedin.com/company/globaltradepartners",
				Twitter:  "https://twitter.com/gtptrading",
			},
			Features: BrandFeatures{
				ShowLiveChat:         false,
				ShowNewsFeed:         true,
				ShowEconomicCalendar: true,
				ShowTradingSignals:   false,
				AllowDemoAccounts:    false,
				AllowCopyTrading:     false,
			},
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * 24 * time.Hour),
			IsActive:  true,
		},
		{
			ID:           5,
			BrandName:    "TradeSmart Asia",
			PlatformName: "TradeSmart",
			LogoURL:      "/assets/logos/tradesmart-logo.svg",
			FaviconURL:   "/assets/logos/tradesmart-favicon.ico",
			PrimaryColor: "#d35400",
			SecondaryColor: "#c0392b",
			AccentColor:  "#f39c12",
			FontFamily:   "Noto Sans, sans-serif",
			CustomCSS:    ".app-container { font-size: 14px; } .button-primary { border-radius: 8px; }",
			CustomDomain: "app.tradesmart.asia",
			EmailSender:  "alerts@tradesmart.asia",
			LoginPageText: "Asia's Leading Trading Platform - Start Trading Today",
			FooterText:   "© 2026 TradeSmart Asia. Licensed by MAS Singapore.",
			SupportEmail: "support@tradesmart.asia",
			SupportPhone: "+65-6789-1234",
			SocialLinks: SocialLinks{
				Facebook:  "https://facebook.com/tradesmartasia",
				Instagram: "https://instagram.com/tradesmartasia",
				YouTube:   "https://youtube.com/c/tradesmartasia",
			},
			Features: BrandFeatures{
				ShowLiveChat:         true,
				ShowNewsFeed:         true,
				ShowEconomicCalendar: true,
				ShowTradingSignals:   true,
				AllowDemoAccounts:    true,
				AllowCopyTrading:     true,
			},
			CreatedAt: time.Now().Add(-20 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
			IsActive:  true,
		},
	}

	store.mu.Lock()
	for _, config := range defaultConfigs {
		store.configs[config.ID] = config
		if config.ID >= store.nextID {
			store.nextID = config.ID + 1
		}
	}
	store.mu.Unlock()

	return &WhiteLabelService{store: store}
}

// ListConfigs returns all active brand configurations
func (wls *WhiteLabelService) ListConfigs(includeDeleted bool) []*WhiteLabelConfig {
	wls.store.mu.RLock()
	defer wls.store.mu.RUnlock()

	configs := make([]*WhiteLabelConfig, 0, len(wls.store.configs))
	for _, config := range wls.store.configs {
		if !includeDeleted && config.DeletedAt != nil {
			continue
		}
		configs = append(configs, config)
	}

	return configs
}

// GetConfig returns a specific brand configuration by ID
func (wls *WhiteLabelService) GetConfig(id int64) (*WhiteLabelConfig, error) {
	wls.store.mu.RLock()
	defer wls.store.mu.RUnlock()

	config, exists := wls.store.configs[id]
	if !exists {
		return nil, fmt.Errorf("white-label config not found")
	}

	if config.DeletedAt != nil {
		return nil, fmt.Errorf("white-label config has been deleted")
	}

	return config, nil
}

// CreateConfig creates a new brand configuration
func (wls *WhiteLabelService) CreateConfig(config *WhiteLabelConfig) (*WhiteLabelConfig, error) {
	wls.store.mu.Lock()
	defer wls.store.mu.Unlock()

	// Validate required fields
	if config.BrandName == "" {
		return nil, fmt.Errorf("brand name is required")
	}
	if config.PlatformName == "" {
		return nil, fmt.Errorf("platform name is required")
	}
	if config.CustomDomain == "" {
		return nil, fmt.Errorf("custom domain is required")
	}

	// Check for duplicate domain
	for _, existing := range wls.store.configs {
		if existing.DeletedAt == nil && existing.CustomDomain == config.CustomDomain {
			return nil, fmt.Errorf("domain already in use by another brand")
		}
	}

	// Assign new ID and timestamps
	config.ID = wls.store.nextID
	wls.store.nextID++
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	config.DeletedAt = nil
	config.IsActive = true

	wls.store.configs[config.ID] = config

	return config, nil
}

// UpdateConfig updates an existing brand configuration
func (wls *WhiteLabelService) UpdateConfig(id int64, updates *WhiteLabelConfig) (*WhiteLabelConfig, error) {
	wls.store.mu.Lock()
	defer wls.store.mu.Unlock()

	config, exists := wls.store.configs[id]
	if !exists {
		return nil, fmt.Errorf("white-label config not found")
	}

	if config.DeletedAt != nil {
		return nil, fmt.Errorf("cannot update deleted config")
	}

	// Validate domain uniqueness if being changed
	if updates.CustomDomain != "" && updates.CustomDomain != config.CustomDomain {
		for _, existing := range wls.store.configs {
			if existing.ID != id && existing.DeletedAt == nil && existing.CustomDomain == updates.CustomDomain {
				return nil, fmt.Errorf("domain already in use by another brand")
			}
		}
		config.CustomDomain = updates.CustomDomain
	}

	// Update fields
	if updates.BrandName != "" {
		config.BrandName = updates.BrandName
	}
	if updates.PlatformName != "" {
		config.PlatformName = updates.PlatformName
	}
	if updates.LogoURL != "" {
		config.LogoURL = updates.LogoURL
	}
	if updates.FaviconURL != "" {
		config.FaviconURL = updates.FaviconURL
	}
	if updates.PrimaryColor != "" {
		config.PrimaryColor = updates.PrimaryColor
	}
	if updates.SecondaryColor != "" {
		config.SecondaryColor = updates.SecondaryColor
	}
	if updates.AccentColor != "" {
		config.AccentColor = updates.AccentColor
	}
	if updates.FontFamily != "" {
		config.FontFamily = updates.FontFamily
	}
	if updates.CustomCSS != "" {
		config.CustomCSS = updates.CustomCSS
	}
	if updates.EmailSender != "" {
		config.EmailSender = updates.EmailSender
	}
	if updates.LoginPageText != "" {
		config.LoginPageText = updates.LoginPageText
	}
	if updates.FooterText != "" {
		config.FooterText = updates.FooterText
	}
	if updates.SupportEmail != "" {
		config.SupportEmail = updates.SupportEmail
	}
	if updates.SupportPhone != "" {
		config.SupportPhone = updates.SupportPhone
	}

	// Update nested structs
	config.SocialLinks = updates.SocialLinks
	config.Features = updates.Features
	config.IsActive = updates.IsActive

	config.UpdatedAt = time.Now()

	return config, nil
}

// DeleteConfig soft-deletes a brand configuration
func (wls *WhiteLabelService) DeleteConfig(id int64) error {
	wls.store.mu.Lock()
	defer wls.store.mu.Unlock()

	config, exists := wls.store.configs[id]
	if !exists {
		return fmt.Errorf("white-label config not found")
	}

	if config.DeletedAt != nil {
		return fmt.Errorf("config already deleted")
	}

	now := time.Now()
	config.DeletedAt = &now
	config.IsActive = false

	return nil
}

// GetPreview returns a preview of the brand configuration with all branding applied
func (wls *WhiteLabelService) GetPreview(id int64) (map[string]interface{}, error) {
	config, err := wls.GetConfig(id)
	if err != nil {
		return nil, err
	}

	// Generate preview data structure
	preview := map[string]interface{}{
		"brandConfig": config,
		"cssVariables": map[string]string{
			"--primary-color":   config.PrimaryColor,
			"--secondary-color": config.SecondaryColor,
			"--accent-color":    config.AccentColor,
			"--font-family":     config.FontFamily,
		},
		"htmlMeta": map[string]string{
			"title":       config.PlatformName,
			"favicon":     config.FaviconURL,
			"description": fmt.Sprintf("%s - Professional Trading Platform", config.PlatformName),
		},
		"loginPage": map[string]string{
			"logo":    config.LogoURL,
			"heading": config.LoginPageText,
			"footer":  config.FooterText,
		},
		"emailTemplateVars": map[string]string{
			"fromEmail":    config.EmailSender,
			"brandName":    config.BrandName,
			"supportEmail": config.SupportEmail,
			"supportPhone": config.SupportPhone,
		},
		"customCSS": config.CustomCSS,
		"features":  config.Features,
		"social":    config.SocialLinks,
		"domain":    config.CustomDomain,
	}

	return preview, nil
}

// WhiteLabelHandler handles HTTP requests for white-label management
type WhiteLabelHandler struct {
	service     *WhiteLabelService
	authService *auth.Service
}

// NewWhiteLabelHandler creates a new white-label HTTP handler
func NewWhiteLabelHandler(service *WhiteLabelService, authService *auth.Service) *WhiteLabelHandler {
	return &WhiteLabelHandler{
		service:     service,
		authService: authService,
	}
}

// ListConfigs handles GET /admin/white-label
func (wlh *WhiteLabelHandler) ListConfigs(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	if _, err := wlh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	includeDeleted := r.URL.Query().Get("includeDeleted") == "true"
	configs := wlh.service.ListConfigs(includeDeleted)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    configs,
		"count":   len(configs),
	})
}

// GetConfig handles GET /admin/white-label/:id
func (wlh *WhiteLabelHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	if _, err := wlh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	config, err := wlh.service.GetConfig(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// CreateConfig handles POST /admin/white-label
func (wlh *WhiteLabelHandler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	if _, err := wlh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var config WhiteLabelConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := wlh.service.CreateConfig(&config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    created,
		"message": "Brand configuration created successfully",
	})
}

// UpdateConfig handles PUT /admin/white-label/:id
func (wlh *WhiteLabelHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	if _, err := wlh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var updates WhiteLabelConfig
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := wlh.service.UpdateConfig(id, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    updated,
		"message": "Brand configuration updated successfully",
	})
}

// DeleteConfig handles DELETE /admin/white-label/:id
func (wlh *WhiteLabelHandler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	if _, err := wlh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := wlh.service.DeleteConfig(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Brand configuration deleted successfully",
	})
}

// PreviewConfig handles POST /admin/white-label/:id/preview
func (wlh *WhiteLabelHandler) PreviewConfig(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Admin authentication
	if _, err := wlh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	idStr := pathParts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	preview, err := wlh.service.GetPreview(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    preview,
	})
}
