package admin

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Platform Configuration Structs
// ============================================

type GeneralConfig struct {
	PlatformName    string `json:"platform_name"`
	CompanyName     string `json:"company_name"`
	Timezone        string `json:"timezone"`
	DefaultLanguage string `json:"default_language"`
	DateFormat      string `json:"date_format"`
	NumberFormat    string `json:"number_format"`
}

type TradingConfig struct {
	DefaultLeverage    int      `json:"default_leverage"`
	MaxLeverage        int      `json:"max_leverage"`
	MarginCallLevel    float64  `json:"margin_call_level"`
	StopOutLevel       float64  `json:"stop_out_level"`
	MaxOpenPositions   int      `json:"max_open_positions"`
	MaxPendingOrders   int      `json:"max_pending_orders"`
	AllowedTradeModes  []string `json:"allowed_trade_modes"`
}

type SecurityConfig struct {
	MinPasswordLength   int    `json:"min_password_length"`
	RequireUppercase    bool   `json:"require_uppercase"`
	RequireNumbers      bool   `json:"require_numbers"`
	RequireSpecial      bool   `json:"require_special"`
	MaxLoginAttempts    int    `json:"max_login_attempts"`
	LockoutDurationMin  int    `json:"lockout_duration_min"`
	SessionTimeoutMin   int    `json:"session_timeout_min"`
	Enforce2FA          bool   `json:"enforce_2fa"`
	IPRestrictionMode   string `json:"ip_restriction_mode"`
}

type NotificationsConfig struct {
	EmailEnabled    bool   `json:"email_enabled"`
	SMSEnabled      bool   `json:"sms_enabled"`
	PushEnabled     bool   `json:"push_enabled"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUser        string `json:"smtp_user"`
	SMTPFromAddress string `json:"smtp_from_address"`
}

type MaintenanceConfig struct {
	MaintenanceMode    bool   `json:"maintenance_mode"`
	MaintenanceMessage string `json:"maintenance_message"`
	ScheduledDowntime  string `json:"scheduled_downtime"`
	DataRetentionDays  int    `json:"data_retention_days"`
}

type APIConfig struct {
	RateLimitPerMinute int    `json:"rate_limit_per_minute"`
	APIVersion         string `json:"api_version"`
	WSMaxConnections   int    `json:"ws_max_connections"`
	MaxPayloadSizeKB   int    `json:"max_payload_size_kb"`
}

type PlatformConfig struct {
	General       GeneralConfig       `json:"general"`
	Trading       TradingConfig       `json:"trading"`
	Security      SecurityConfig      `json:"security"`
	Notifications NotificationsConfig `json:"notifications"`
	Maintenance   MaintenanceConfig   `json:"maintenance"`
	API           APIConfig           `json:"api"`
	LastUpdated   time.Time           `json:"last_updated"`
}

type ConfigHistory struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	AdminName string    `json:"admin_name"`
	Section   string    `json:"section"`
	Field     string    `json:"field"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
}

// ============================================
// Platform Config Service
// ============================================

type PlatformConfigService struct {
	mu              sync.RWMutex
	config          PlatformConfig
	history         []ConfigHistory
	nextHistoryID   int64
	defaultConfig   PlatformConfig
}

func NewPlatformConfigService() *PlatformConfigService {
	defaultConfig := PlatformConfig{
		General: GeneralConfig{
			PlatformName:    "RTX5 Trading Platform",
			CompanyName:     "RTX5 Global Trading",
			Timezone:        "UTC",
			DefaultLanguage: "en",
			DateFormat:      "YYYY-MM-DD",
			NumberFormat:    "1,234.56",
		},
		Trading: TradingConfig{
			DefaultLeverage:   100,
			MaxLeverage:       500,
			MarginCallLevel:   80.0,
			StopOutLevel:      50.0,
			MaxOpenPositions:  200,
			MaxPendingOrders:  100,
			AllowedTradeModes: []string{"instant", "market", "exchange"},
		},
		Security: SecurityConfig{
			MinPasswordLength:  8,
			RequireUppercase:   true,
			RequireNumbers:     true,
			RequireSpecial:     true,
			MaxLoginAttempts:   5,
			LockoutDurationMin: 30,
			SessionTimeoutMin:  1440,
			Enforce2FA:         false,
			IPRestrictionMode:  "none",
		},
		Notifications: NotificationsConfig{
			EmailEnabled:    true,
			SMSEnabled:      false,
			PushEnabled:     true,
			SMTPHost:        "smtp.example.com",
			SMTPPort:        587,
			SMTPUser:        "noreply@rtx5.com",
			SMTPFromAddress: "RTX5 Platform <noreply@rtx5.com>",
		},
		Maintenance: MaintenanceConfig{
			MaintenanceMode:    false,
			MaintenanceMessage: "Platform is currently under maintenance. Please check back soon.",
			ScheduledDowntime:  "",
			DataRetentionDays:  365,
		},
		API: APIConfig{
			RateLimitPerMinute: 100,
			APIVersion:         "v1",
			WSMaxConnections:   10000,
			MaxPayloadSizeKB:   1024,
		},
		LastUpdated: time.Now(),
	}

	// Create initial history entry
	history := []ConfigHistory{
		{
			ID:        1,
			Timestamp: time.Now().Add(-48 * time.Hour),
			AdminName: "System",
			Section:   "general",
			Field:     "platform_name",
			OldValue:  "Trading Platform",
			NewValue:  "RTX5 Trading Platform",
		},
		{
			ID:        2,
			Timestamp: time.Now().Add(-36 * time.Hour),
			AdminName: "admin@rtx5.com",
			Section:   "trading",
			Field:     "max_leverage",
			OldValue:  "400",
			NewValue:  "500",
		},
		{
			ID:        3,
			Timestamp: time.Now().Add(-24 * time.Hour),
			AdminName: "admin@rtx5.com",
			Section:   "security",
			Field:     "session_timeout_min",
			OldValue:  "720",
			NewValue:  "1440",
		},
		{
			ID:        4,
			Timestamp: time.Now().Add(-12 * time.Hour),
			AdminName: "security@rtx5.com",
			Section:   "security",
			Field:     "enforce_2fa",
			OldValue:  "true",
			NewValue:  "false",
		},
		{
			ID:        5,
			Timestamp: time.Now().Add(-6 * time.Hour),
			AdminName: "admin@rtx5.com",
			Section:   "notifications",
			Field:     "sms_enabled",
			OldValue:  "true",
			NewValue:  "false",
		},
		{
			ID:        6,
			Timestamp: time.Now().Add(-3 * time.Hour),
			AdminName: "devops@rtx5.com",
			Section:   "api",
			Field:     "rate_limit_per_minute",
			OldValue:  "60",
			NewValue:  "100",
		},
		{
			ID:        7,
			Timestamp: time.Now().Add(-1 * time.Hour),
			AdminName: "admin@rtx5.com",
			Section:   "maintenance",
			Field:     "data_retention_days",
			OldValue:  "180",
			NewValue:  "365",
		},
	}

	return &PlatformConfigService{
		config:        defaultConfig,
		history:       history,
		nextHistoryID: 8,
		defaultConfig: defaultConfig,
	}
}

func (s *PlatformConfigService) GetAllConfig() PlatformConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *PlatformConfigService) GetSection(section string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	switch strings.ToLower(section) {
	case "general":
		return s.config.General, true
	case "trading":
		return s.config.Trading, true
	case "security":
		return s.config.Security, true
	case "notifications":
		return s.config.Notifications, true
	case "maintenance":
		return s.config.Maintenance, true
	case "api":
		return s.config.API, true
	default:
		return nil, false
	}
}

func (s *PlatformConfigService) UpdateSection(section string, data []byte, adminName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch strings.ToLower(section) {
	case "general":
		var newConfig GeneralConfig
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return err
		}
		s.recordChanges("general", s.config.General, newConfig, adminName)
		s.config.General = newConfig
	case "trading":
		var newConfig TradingConfig
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return err
		}
		s.recordChanges("trading", s.config.Trading, newConfig, adminName)
		s.config.Trading = newConfig
	case "security":
		var newConfig SecurityConfig
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return err
		}
		s.recordChanges("security", s.config.Security, newConfig, adminName)
		s.config.Security = newConfig
	case "notifications":
		var newConfig NotificationsConfig
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return err
		}
		s.recordChanges("notifications", s.config.Notifications, newConfig, adminName)
		s.config.Notifications = newConfig
	case "maintenance":
		var newConfig MaintenanceConfig
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return err
		}
		s.recordChanges("maintenance", s.config.Maintenance, newConfig, adminName)
		s.config.Maintenance = newConfig
	case "api":
		var newConfig APIConfig
		if err := json.Unmarshal(data, &newConfig); err != nil {
			return err
		}
		s.recordChanges("api", s.config.API, newConfig, adminName)
		s.config.API = newConfig
	default:
		return nil
	}

	s.config.LastUpdated = time.Now()
	return nil
}

func (s *PlatformConfigService) ResetSection(section string, adminName string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch strings.ToLower(section) {
	case "general":
		s.recordChanges("general", s.config.General, s.defaultConfig.General, adminName)
		s.config.General = s.defaultConfig.General
	case "trading":
		s.recordChanges("trading", s.config.Trading, s.defaultConfig.Trading, adminName)
		s.config.Trading = s.defaultConfig.Trading
	case "security":
		s.recordChanges("security", s.config.Security, s.defaultConfig.Security, adminName)
		s.config.Security = s.defaultConfig.Security
	case "notifications":
		s.recordChanges("notifications", s.config.Notifications, s.defaultConfig.Notifications, adminName)
		s.config.Notifications = s.defaultConfig.Notifications
	case "maintenance":
		s.recordChanges("maintenance", s.config.Maintenance, s.defaultConfig.Maintenance, adminName)
		s.config.Maintenance = s.defaultConfig.Maintenance
	case "api":
		s.recordChanges("api", s.config.API, s.defaultConfig.API, adminName)
		s.config.API = s.defaultConfig.API
	default:
		return false
	}

	s.config.LastUpdated = time.Now()
	return true
}

func (s *PlatformConfigService) GetHistory(limit int) []ConfigHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.history) {
		limit = len(s.history)
	}

	// Return most recent entries first
	result := make([]ConfigHistory, 0, limit)
	start := len(s.history) - limit
	for i := len(s.history) - 1; i >= start; i-- {
		result = append(result, s.history[i])
	}

	return result
}

func (s *PlatformConfigService) recordChanges(section string, oldConfig, newConfig interface{}, adminName string) {
	oldJSON, _ := json.Marshal(oldConfig)
	newJSON, _ := json.Marshal(newConfig)

	var oldMap, newMap map[string]interface{}
	json.Unmarshal(oldJSON, &oldMap)
	json.Unmarshal(newJSON, &newMap)

	for field, newValue := range newMap {
		oldValue := oldMap[field]
		if oldValue != newValue {
			s.history = append(s.history, ConfigHistory{
				ID:        s.nextHistoryID,
				Timestamp: time.Now(),
				AdminName: adminName,
				Section:   section,
				Field:     field,
				OldValue:  toString(oldValue),
				NewValue:  toString(newValue),
			})
			s.nextHistoryID++
		}
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// ============================================
// Platform Config Handler
// ============================================

type PlatformConfigHandler struct {
	service     *PlatformConfigService
	authService *AuthService
}

func NewPlatformConfigHandler(service *PlatformConfigService, authService *AuthService) *PlatformConfigHandler {
	return &PlatformConfigHandler{
		service:     service,
		authService: authService,
	}
}

func (h *PlatformConfigHandler) HandleGetAllConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	config := h.service.GetAllConfig()
	json.NewEncoder(w).Encode(config)
}

func (h *PlatformConfigHandler) HandleGetSection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Extract section from URL path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/platform-config/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Section not specified", http.StatusBadRequest)
		return
	}
	section := parts[0]

	sectionData, found := h.service.GetSection(section)
	if !found {
		http.Error(w, "Section not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(sectionData)
}

func (h *PlatformConfigHandler) HandleUpdateSection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateTokenWithDefault(token)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Extract section from URL path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/platform-config/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Section not specified", http.StatusBadRequest)
		return
	}
	section := parts[0]

	adminName := claims.Username

	// Read request body
	var data json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSection(section, data, adminName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Section updated successfully",
	})
}

func (h *PlatformConfigHandler) HandleResetSection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateTokenWithDefault(token)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Extract section from URL path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/platform-config/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	section := pathParts[0]

	adminName := claims.Username

	if !h.service.ResetSection(section, adminName) {
		http.Error(w, "Section not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Section reset to defaults",
	})
}

func (h *PlatformConfigHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Default to last 50 entries
	limit := 50
	history := h.service.GetHistory(limit)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"total":   len(history),
	})
}
