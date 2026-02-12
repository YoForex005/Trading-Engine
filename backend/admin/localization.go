package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Localization Structs
// ============================================

type Language struct {
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	NativeName      string    `json:"native_name"`
	Enabled         bool      `json:"enabled"`
	IsDefault       bool      `json:"is_default"`
	Completion      float64   `json:"completion"`
	TotalKeys       int       `json:"total_keys"`
	TranslatedKeys  int       `json:"translated_keys"`
	LastUpdated     time.Time `json:"last_updated"`
	UpdatedBy       string    `json:"updated_by"`
}

type Translation struct {
	Key         string            `json:"key"`
	Translations map[string]string `json:"translations"` // language code -> translation
	Category    string            `json:"category"`
	LastUpdated time.Time         `json:"last_updated"`
}

type TranslationImport struct {
	LanguageCode string            `json:"language_code"`
	Translations map[string]string `json:"translations"`
	Overwrite    bool              `json:"overwrite"`
}

// ============================================
// Localization Service
// ============================================

type LocalizationService struct {
	mu           sync.RWMutex
	languages    map[string]*Language
	translations map[string]*Translation // key -> Translation
}

func NewLocalizationService() *LocalizationService {
	service := &LocalizationService{
		languages:    make(map[string]*Language),
		translations: make(map[string]*Translation),
	}

	// Initialize 10 languages
	now := time.Now()

	service.languages["en"] = &Language{
		Code:           "en",
		Name:           "English",
		NativeName:     "English",
		Enabled:        true,
		IsDefault:      true,
		Completion:     100.0,
		TotalKeys:      500,
		TranslatedKeys: 500,
		LastUpdated:    now,
		UpdatedBy:      "system",
	}

	service.languages["es"] = &Language{
		Code:           "es",
		Name:           "Spanish",
		NativeName:     "Español",
		Enabled:        true,
		IsDefault:      false,
		Completion:     95.2,
		TotalKeys:      500,
		TranslatedKeys: 476,
		LastUpdated:    now.Add(-2 * time.Hour),
		UpdatedBy:      "translator@rtx5.com",
	}

	service.languages["fr"] = &Language{
		Code:           "fr",
		Name:           "French",
		NativeName:     "Français",
		Enabled:        true,
		IsDefault:      false,
		Completion:     92.8,
		TotalKeys:      500,
		TranslatedKeys: 464,
		LastUpdated:    now.Add(-5 * time.Hour),
		UpdatedBy:      "translator@rtx5.com",
	}

	service.languages["de"] = &Language{
		Code:           "de",
		Name:           "German",
		NativeName:     "Deutsch",
		Enabled:        true,
		IsDefault:      false,
		Completion:     88.4,
		TotalKeys:      500,
		TranslatedKeys: 442,
		LastUpdated:    now.Add(-12 * time.Hour),
		UpdatedBy:      "translator@rtx5.com",
	}

	service.languages["ja"] = &Language{
		Code:           "ja",
		Name:           "Japanese",
		NativeName:     "日本語",
		Enabled:        true,
		IsDefault:      false,
		Completion:     85.6,
		TotalKeys:      500,
		TranslatedKeys: 428,
		LastUpdated:    now.Add(-24 * time.Hour),
		UpdatedBy:      "translator-jp@rtx5.com",
	}

	service.languages["zh"] = &Language{
		Code:           "zh",
		Name:           "Chinese",
		NativeName:     "中文",
		Enabled:        true,
		IsDefault:      false,
		Completion:     90.2,
		TotalKeys:      500,
		TranslatedKeys: 451,
		LastUpdated:    now.Add(-18 * time.Hour),
		UpdatedBy:      "translator-cn@rtx5.com",
	}

	service.languages["ar"] = &Language{
		Code:           "ar",
		Name:           "Arabic",
		NativeName:     "العربية",
		Enabled:        false,
		IsDefault:      false,
		Completion:     72.4,
		TotalKeys:      500,
		TranslatedKeys: 362,
		LastUpdated:    now.Add(-72 * time.Hour),
		UpdatedBy:      "translator-ar@rtx5.com",
	}

	service.languages["pt"] = &Language{
		Code:           "pt",
		Name:           "Portuguese",
		NativeName:     "Português",
		Enabled:        true,
		IsDefault:      false,
		Completion:     87.0,
		TotalKeys:      500,
		TranslatedKeys: 435,
		LastUpdated:    now.Add(-36 * time.Hour),
		UpdatedBy:      "translator@rtx5.com",
	}

	service.languages["ru"] = &Language{
		Code:           "ru",
		Name:           "Russian",
		NativeName:     "Русский",
		Enabled:        true,
		IsDefault:      false,
		Completion:     83.8,
		TotalKeys:      500,
		TranslatedKeys: 419,
		LastUpdated:    now.Add(-48 * time.Hour),
		UpdatedBy:      "translator-ru@rtx5.com",
	}

	service.languages["ko"] = &Language{
		Code:           "ko",
		Name:           "Korean",
		NativeName:     "한국어",
		Enabled:        false,
		IsDefault:      false,
		Completion:     68.2,
		TotalKeys:      500,
		TranslatedKeys: 341,
		LastUpdated:    now.Add(-96 * time.Hour),
		UpdatedBy:      "translator-ko@rtx5.com",
	}

	// Initialize 500+ translation keys
	service.initializeTranslations()

	return service
}

func (s *LocalizationService) initializeTranslations() {
	categories := []string{
		"common", "auth", "trading", "account", "navigation",
		"errors", "validation", "notifications", "reports", "settings",
	}

	// Common translations (50 keys)
	commonKeys := []string{
		"app.title", "app.welcome", "app.loading", "app.error", "app.success",
		"button.save", "button.cancel", "button.delete", "button.edit", "button.create",
		"button.confirm", "button.close", "button.back", "button.next", "button.submit",
		"button.search", "button.filter", "button.export", "button.import", "button.refresh",
		"label.name", "label.email", "label.phone", "label.address", "label.date",
		"label.time", "label.status", "label.type", "label.amount", "label.currency",
		"status.active", "status.inactive", "status.pending", "status.completed", "status.cancelled",
		"common.yes", "common.no", "common.all", "common.none", "common.other",
		"common.total", "common.subtotal", "common.discount", "common.tax", "common.fee",
		"common.from", "common.to", "common.select", "common.choose", "common.optional",
	}

	// Auth translations (40 keys)
	authKeys := []string{
		"auth.login", "auth.logout", "auth.register", "auth.forgot_password", "auth.reset_password",
		"auth.username", "auth.password", "auth.confirm_password", "auth.email", "auth.phone",
		"auth.remember_me", "auth.forgot", "auth.create_account", "auth.have_account", "auth.no_account",
		"auth.verification_code", "auth.verify", "auth.resend_code", "auth.expired", "auth.invalid",
		"auth.success.login", "auth.success.logout", "auth.success.register", "auth.success.reset",
		"auth.error.invalid_credentials", "auth.error.user_exists", "auth.error.user_not_found",
		"auth.error.weak_password", "auth.error.passwords_mismatch", "auth.error.invalid_email",
		"auth.2fa.enable", "auth.2fa.disable", "auth.2fa.code", "auth.2fa.backup_codes",
		"auth.session.expired", "auth.session.timeout", "auth.session.active", "auth.session.inactive",
		"auth.permissions.denied", "auth.permissions.required",
	}

	// Trading translations (80 keys)
	tradingKeys := []string{
		"trading.buy", "trading.sell", "trading.close", "trading.modify", "trading.cancel",
		"trading.market_order", "trading.limit_order", "trading.stop_order", "trading.pending_order",
		"trading.symbol", "trading.volume", "trading.lots", "trading.price", "trading.bid", "trading.ask",
		"trading.spread", "trading.leverage", "trading.margin", "trading.free_margin", "trading.equity",
		"trading.balance", "trading.profit", "trading.loss", "trading.pnl", "trading.commission",
		"trading.swap", "trading.stop_loss", "trading.take_profit", "trading.trailing_stop",
		"trading.position", "trading.positions", "trading.order", "trading.orders", "trading.history",
		"trading.open_positions", "trading.closed_positions", "trading.pending_orders",
		"trading.trade_size", "trading.pip_value", "trading.pip_cost", "trading.required_margin",
		"trading.market_execution", "trading.instant_execution", "trading.limit_price",
		"trading.expiry", "trading.good_till_cancelled", "trading.good_till_date",
		"trading.hedge", "trading.netting", "trading.account_type", "trading.demo", "trading.live",
		"trading.success.order_placed", "trading.success.order_modified", "trading.success.order_closed",
		"trading.error.insufficient_margin", "trading.error.invalid_volume", "trading.error.invalid_price",
		"trading.error.market_closed", "trading.error.off_quotes", "trading.error.requote",
		"trading.error.trade_disabled", "trading.error.max_orders", "trading.error.connection",
		"trading.chart", "trading.timeframe", "trading.indicators", "trading.drawing_tools",
		"trading.watchlist", "trading.favorites", "trading.recent", "trading.trending",
	}

	// Account translations (60 keys)
	accountKeys := []string{
		"account.profile", "account.settings", "account.security", "account.preferences",
		"account.personal_info", "account.contact_info", "account.verification", "account.documents",
		"account.deposit", "account.withdraw", "account.transfer", "account.transactions",
		"account.balance", "account.equity", "account.margin", "account.credit", "account.bonus",
		"account.account_number", "account.account_type", "account.account_currency",
		"account.leverage", "account.group", "account.server", "account.status",
		"account.created", "account.updated", "account.verified", "account.unverified",
		"account.active", "account.inactive", "account.suspended", "account.closed",
		"account.deposit.methods", "account.deposit.minimum", "account.deposit.maximum",
		"account.withdraw.methods", "account.withdraw.minimum", "account.withdraw.maximum",
		"account.withdraw.pending", "account.withdraw.approved", "account.withdraw.rejected",
		"account.kyc.required", "account.kyc.pending", "account.kyc.approved", "account.kyc.rejected",
		"account.documents.id", "account.documents.address", "account.documents.bank",
		"account.notifications.email", "account.notifications.sms", "account.notifications.push",
		"account.password.change", "account.password.current", "account.password.new",
		"account.2fa.enabled", "account.2fa.disabled", "account.2fa.setup",
	}

	// Navigation translations (40 keys)
	navKeys := []string{
		"nav.dashboard", "nav.trading", "nav.accounts", "nav.portfolio", "nav.reports",
		"nav.settings", "nav.help", "nav.support", "nav.profile", "nav.logout",
		"nav.market_watch", "nav.charts", "nav.orders", "nav.positions", "nav.history",
		"nav.news", "nav.calendar", "nav.analysis", "nav.tools", "nav.education",
		"nav.deposits", "nav.withdrawals", "nav.transfers", "nav.statements",
		"nav.notifications", "nav.messages", "nav.alerts", "nav.preferences",
		"nav.security", "nav.verification", "nav.documents", "nav.api",
		"nav.admin", "nav.users", "nav.traders", "nav.brokers", "nav.symbols",
		"nav.groups", "nav.commissions", "nav.margins", "nav.swaps",
	}

	// Errors translations (50 keys)
	errorKeys := []string{
		"error.unknown", "error.server", "error.network", "error.timeout", "error.unauthorized",
		"error.forbidden", "error.not_found", "error.bad_request", "error.conflict",
		"error.validation", "error.required_field", "error.invalid_format", "error.min_length",
		"error.max_length", "error.min_value", "error.max_value", "error.invalid_email",
		"error.invalid_phone", "error.invalid_date", "error.invalid_time", "error.invalid_url",
		"error.file_too_large", "error.file_type", "error.upload_failed", "error.download_failed",
		"error.connection_lost", "error.connection_failed", "error.reconnecting",
		"error.rate_limit", "error.quota_exceeded", "error.service_unavailable",
		"error.maintenance", "error.deprecated", "error.unsupported",
		"error.duplicate_entry", "error.not_allowed", "error.expired", "error.invalid_token",
		"error.session_expired", "error.access_denied", "error.insufficient_balance",
		"error.insufficient_permissions", "error.operation_failed", "error.try_again",
		"error.contact_support", "error.feature_disabled", "error.coming_soon",
	}

	// Validation translations (40 keys)
	validationKeys := []string{
		"validation.required", "validation.email", "validation.phone", "validation.url",
		"validation.min", "validation.max", "validation.minLength", "validation.maxLength",
		"validation.pattern", "validation.numeric", "validation.alpha", "validation.alphanumeric",
		"validation.match", "validation.unique", "validation.exists", "validation.between",
		"validation.positive", "validation.negative", "validation.integer", "validation.decimal",
		"validation.date", "validation.time", "validation.datetime", "validation.before",
		"validation.after", "validation.future", "validation.past", "validation.age",
		"validation.password.weak", "validation.password.medium", "validation.password.strong",
		"validation.file.size", "validation.file.type", "validation.image.dimensions",
		"validation.array.min", "validation.array.max", "validation.select", "validation.checkbox",
		"validation.radio", "validation.accept_terms",
	}

	// Notifications translations (40 keys)
	notifKeys := []string{
		"notif.new_message", "notif.new_order", "notif.order_filled", "notif.order_cancelled",
		"notif.position_closed", "notif.stop_loss_hit", "notif.take_profit_hit",
		"notif.margin_call", "notif.deposit_received", "notif.withdrawal_processed",
		"notif.verification_required", "notif.document_approved", "notif.document_rejected",
		"notif.price_alert", "notif.news_alert", "notif.system_alert", "notif.maintenance",
		"notif.new_feature", "notif.promotion", "notif.bonus_credited", "notif.competition",
		"notif.trade_executed", "notif.trade_modified", "notif.trade_closed",
		"notif.account_verified", "notif.password_changed", "notif.2fa_enabled",
		"notif.login_detected", "notif.suspicious_activity", "notif.session_expired",
		"notif.market_open", "notif.market_close", "notif.spread_widening",
		"notif.high_volatility", "notif.economic_event", "notif.dividend_payment",
		"notif.swap_triple", "notif.rollover", "notif.commission_charged",
	}

	// Reports translations (40 keys)
	reportKeys := []string{
		"report.daily", "report.weekly", "report.monthly", "report.yearly", "report.custom",
		"report.performance", "report.pnl", "report.trades", "report.positions", "report.orders",
		"report.balance", "report.equity", "report.margin", "report.commissions", "report.swaps",
		"report.deposits", "report.withdrawals", "report.transfers", "report.statements",
		"report.tax", "report.compliance", "report.risk", "report.exposure", "report.volatility",
		"report.summary", "report.detailed", "report.export", "report.download", "report.print",
		"report.date_range", "report.filter", "report.sort", "report.group", "report.total",
		"report.average", "report.maximum", "report.minimum", "report.count",
		"report.win_rate", "report.profit_factor", "report.sharpe_ratio",
	}

	// Settings translations (60 keys)
	settingsKeys := []string{
		"settings.general", "settings.account", "settings.trading", "settings.notifications",
		"settings.security", "settings.privacy", "settings.appearance", "settings.language",
		"settings.timezone", "settings.currency", "settings.date_format", "settings.time_format",
		"settings.theme", "settings.theme.light", "settings.theme.dark", "settings.theme.auto",
		"settings.font_size", "settings.chart_type", "settings.chart_colors", "settings.sound",
		"settings.alerts", "settings.email_notifications", "settings.sms_notifications",
		"settings.push_notifications", "settings.desktop_notifications",
		"settings.two_factor", "settings.sessions", "settings.devices", "settings.api_access",
		"settings.password", "settings.pin", "settings.biometric", "settings.backup",
		"settings.export_data", "settings.import_data", "settings.delete_account",
		"settings.trading.one_click", "settings.trading.confirm", "settings.trading.default_volume",
		"settings.trading.default_sl", "settings.trading.default_tp", "settings.trading.max_slippage",
		"settings.chart.default_timeframe", "settings.chart.default_indicators",
		"settings.chart.save_layout", "settings.chart.auto_scroll", "settings.chart.crosshair",
		"settings.notifications.trade", "settings.notifications.account", "settings.notifications.market",
		"settings.notifications.system", "settings.notifications.marketing",
		"settings.privacy.profile", "settings.privacy.activity", "settings.privacy.trading",
	}

	allKeys := [][]string{
		commonKeys, authKeys, tradingKeys, accountKeys, navKeys,
		errorKeys, validationKeys, notifKeys, reportKeys, settingsKeys,
	}

	idx := 0
	for catIdx, keys := range allKeys {
		category := categories[catIdx]
		for _, key := range keys {
			s.translations[key] = &Translation{
				Key:      key,
				Category: category,
				Translations: map[string]string{
					"en": s.generateEnglishTranslation(key),
				},
				LastUpdated: time.Now(),
			}
			idx++
		}
	}
}

func (s *LocalizationService) generateEnglishTranslation(key string) string {
	parts := strings.Split(key, ".")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		return strings.Title(strings.ReplaceAll(lastPart, "_", " "))
	}
	return key
}

func (s *LocalizationService) GetLanguages() []Language {
	s.mu.RLock()
	defer s.mu.RUnlock()

	langs := make([]Language, 0, len(s.languages))
	for _, lang := range s.languages {
		langs = append(langs, *lang)
	}
	return langs
}

func (s *LocalizationService) ToggleLanguage(code string, enabled bool, adminEmail string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	lang, exists := s.languages[code]
	if !exists {
		return false
	}

	lang.Enabled = enabled
	lang.LastUpdated = time.Now()
	lang.UpdatedBy = adminEmail
	return true
}

func (s *LocalizationService) GetTranslations(languageCode string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	for key, trans := range s.translations {
		if val, ok := trans.Translations[languageCode]; ok {
			result[key] = val
		}
	}
	return result
}

func (s *LocalizationService) UpdateTranslation(key, languageCode, value, adminEmail string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trans, exists := s.translations[key]
	if !exists {
		return fmt.Errorf("translation key not found: %s", key)
	}

	trans.Translations[languageCode] = value
	trans.LastUpdated = time.Now()

	// Update language completion
	if lang, ok := s.languages[languageCode]; ok {
		translated := 0
		for _, t := range s.translations {
			if _, ok := t.Translations[languageCode]; ok {
				translated++
			}
		}
		lang.TranslatedKeys = translated
		lang.Completion = float64(translated) / float64(len(s.translations)) * 100.0
		lang.LastUpdated = time.Now()
		lang.UpdatedBy = adminEmail
	}

	return nil
}

func (s *LocalizationService) ImportTranslations(imp TranslationImport, adminEmail string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.languages[imp.LanguageCode]; !exists {
		return 0, fmt.Errorf("language not found: %s", imp.LanguageCode)
	}

	imported := 0
	for key, value := range imp.Translations {
		trans, exists := s.translations[key]
		if !exists {
			continue
		}

		if imp.Overwrite || trans.Translations[imp.LanguageCode] == "" {
			trans.Translations[imp.LanguageCode] = value
			trans.LastUpdated = time.Now()
			imported++
		}
	}

	// Update language completion
	if lang, ok := s.languages[imp.LanguageCode]; ok {
		translated := 0
		for _, t := range s.translations {
			if _, ok := t.Translations[imp.LanguageCode]; ok {
				translated++
			}
		}
		lang.TranslatedKeys = translated
		lang.Completion = float64(translated) / float64(len(s.translations)) * 100.0
		lang.LastUpdated = time.Now()
		lang.UpdatedBy = adminEmail
	}

	return imported, nil
}

func (s *LocalizationService) ExportTranslations(languageCode string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.languages[languageCode]; !exists {
		return nil, fmt.Errorf("language not found: %s", languageCode)
	}

	result := make(map[string]string)
	for key, trans := range s.translations {
		if val, ok := trans.Translations[languageCode]; ok {
			result[key] = val
		}
	}

	return result, nil
}

// ============================================
// Localization Handler
// ============================================

type LocalizationHandler struct {
	service     *LocalizationService
	authService *AuthService
}

func NewLocalizationHandler(service *LocalizationService, authService *AuthService) *LocalizationHandler {
	return &LocalizationHandler{
		service:     service,
		authService: authService,
	}
}

func (h *LocalizationHandler) HandleListLanguages(w http.ResponseWriter, r *http.Request) {
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

	languages := h.service.GetLanguages()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"languages": languages,
		"total":     len(languages),
	})
}

func (h *LocalizationHandler) HandleToggleLanguage(w http.ResponseWriter, r *http.Request) {
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

	// Extract language code from URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/localization/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Language code not specified", http.StatusBadRequest)
		return
	}
	code := parts[0]

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !h.service.ToggleLanguage(code, req.Enabled, claims.Username) {
		http.Error(w, "Language not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Language status updated",
	})
}

func (h *LocalizationHandler) HandleGetTranslations(w http.ResponseWriter, r *http.Request) {
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

	// Extract language code from URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/localization/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Language code not specified", http.StatusBadRequest)
		return
	}
	code := parts[0]

	translations := h.service.GetTranslations(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"language":     code,
		"translations": translations,
		"total":        len(translations),
	})
}

func (h *LocalizationHandler) HandleUpdateTranslation(w http.ResponseWriter, r *http.Request) {
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

	var req struct {
		Key          string `json:"key"`
		LanguageCode string `json:"language_code"`
		Value        string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateTranslation(req.Key, req.LanguageCode, req.Value, claims.Username); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Translation updated",
	})
}

func (h *LocalizationHandler) HandleImportTranslations(w http.ResponseWriter, r *http.Request) {
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

	var imp TranslationImport
	if err := json.NewDecoder(r.Body).Decode(&imp); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	imported, err := h.service.ImportTranslations(imp, claims.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"imported": imported,
		"message":  fmt.Sprintf("Imported %d translations", imported),
	})
}

func (h *LocalizationHandler) HandleExportTranslations(w http.ResponseWriter, r *http.Request) {
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

	// Extract language code from URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/localization/"), "/")
	if len(parts) < 2 || parts[0] == "" {
		http.Error(w, "Language code not specified", http.StatusBadRequest)
		return
	}
	code := parts[0]

	translations, err := h.service.ExportTranslations(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"language":     code,
		"translations": translations,
		"total":        len(translations),
	})
}
