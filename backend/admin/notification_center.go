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
)

// ============================================
// Notification Center Structs
// ============================================

type Notification struct {
	ID            int64     `json:"id"`
	Type          string    `json:"type"`     // trade_executed, margin_call, deposit_received, etc.
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	Severity      string    `json:"severity"`      // info, success, warning, error, critical
	RecipientType string    `json:"recipient_type"` // client, admin, all
	RecipientID   string    `json:"recipient_id"`
	Channel       string    `json:"channel"`        // in_app, email, sms, push, webhook
	Status        string    `json:"status"`         // pending, sent, delivered, read, failed
	CreatedAt     time.Time `json:"created_at"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	ActionURL     string    `json:"action_url,omitempty"`
}

type NotificationRule struct {
	ID           int64             `json:"id"`
	Name         string            `json:"name"`
	TriggerEvent string            `json:"trigger_event"` // trade_executed, margin_call, etc.
	Conditions   map[string]string `json:"conditions"`
	Channels     []string          `json:"channels"`
	Template     string            `json:"template"`
	IsActive     bool              `json:"is_active"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type NotificationStats struct {
	TotalSent     int                   `json:"total_sent"`
	TotalRead     int                   `json:"total_read"`
	ReadRate      float64               `json:"read_rate"`      // percentage
	ByChannel     map[string]int        `json:"by_channel"`
	BySeverity    map[string]int        `json:"by_severity"`
	ByType        map[string]int        `json:"by_type"`
	ByStatus      map[string]int        `json:"by_status"`
	Generated     time.Time             `json:"generated"`
}

type BroadcastRequest struct {
	Title       string   `json:"title"`
	Message     string   `json:"message"`
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Channels    []string `json:"channels"`
	TargetType  string   `json:"target_type"`  // all, segment
	SegmentID   string   `json:"segment_id,omitempty"`
	ActionURL   string   `json:"action_url,omitempty"`
}

type SendNotificationRequest struct {
	RecipientID   string   `json:"recipient_id"`
	RecipientType string   `json:"recipient_type"` // client, admin
	Title         string   `json:"title"`
	Message       string   `json:"message"`
	Type          string   `json:"type"`
	Severity      string   `json:"severity"`
	Channels      []string `json:"channels"`
	ActionURL     string   `json:"action_url,omitempty"`
}

// ============================================
// Notification Center Service
// ============================================

type NotificationCenterService struct {
	mu            sync.RWMutex
	notifications []Notification
	rules         map[int64]*NotificationRule
	nextNotifID   int64
	nextRuleID    int64
	clientIDs     []string
}

func NewNotificationCenterService() *NotificationCenterService {
	service := &NotificationCenterService{
		notifications: make([]Notification, 0, 500),
		rules:         make(map[int64]*NotificationRule),
		nextNotifID:   1,
		nextRuleID:    1,
		clientIDs:     make([]string, 50),
	}

	// Generate 50 client IDs
	for i := 0; i < 50; i++ {
		service.clientIDs[i] = fmt.Sprintf("client-%d", 10000+i)
	}

	service.generateMockData()
	return service
}

func (s *NotificationCenterService) generateMockData() {
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	notifTypes := []string{
		"trade_executed", "margin_call", "deposit_received", "withdrawal_processed",
		"price_alert", "system_maintenance", "account_update", "security_alert",
		"promotion", "custom",
	}

	channels := []string{"in_app", "email", "sms", "push", "webhook"}
	severities := []string{"info", "success", "warning", "error", "critical"}
	statuses := []string{"sent", "delivered", "read", "failed"}

	// Generate 500 notifications
	for i := 0; i < 500; i++ {
		notifType := notifTypes[rand.Intn(len(notifTypes))]
		channel := channels[rand.Intn(len(channels))]
		severity := severities[rand.Intn(len(severities))]
		status := statuses[rand.Intn(len(statuses))]

		// Weight severity distribution
		if rand.Float64() < 0.6 {
			severity = "info"
		} else if rand.Float64() < 0.8 {
			severity = "success"
		}

		// Most are delivered/read
		if rand.Float64() < 0.85 {
			status = "read"
		} else if rand.Float64() < 0.95 {
			status = "delivered"
		}

		clientID := s.clientIDs[rand.Intn(len(s.clientIDs))]

		// Time distribution - recent notifications
		minutesAgo := int(math.Pow(rand.Float64(), 2) * 1440) // last 24 hours, skewed to recent
		createdAt := now.Add(-time.Duration(minutesAgo) * time.Minute)

		title, message, actionURL := s.generateNotificationContent(notifType)

		notif := Notification{
			ID:            s.nextNotifID,
			Type:          notifType,
			Title:         title,
			Message:       message,
			Severity:      severity,
			RecipientType: "client",
			RecipientID:   clientID,
			Channel:       channel,
			Status:        status,
			CreatedAt:     createdAt,
			ActionURL:     actionURL,
		}

		// Set read time for read notifications
		if status == "read" {
			readMinutesAfter := rand.Intn(120) + 5 // read within 5-125 minutes
			readAt := createdAt.Add(time.Duration(readMinutesAfter) * time.Minute)
			notif.ReadAt = &readAt
		}

		s.notifications = append(s.notifications, notif)
		s.nextNotifID++
	}

	// Sort by creation time descending
	sort.Slice(s.notifications, func(i, j int) bool {
		return s.notifications[i].CreatedAt.After(s.notifications[j].CreatedAt)
	})

	// Generate 15 notification rules
	ruleConfigs := []struct {
		name    string
		event   string
		active  bool
		channels []string
	}{
		{"Trade Execution Alert", "trade_executed", true, []string{"in_app", "push"}},
		{"Margin Call Warning", "margin_call", true, []string{"in_app", "email", "sms", "push"}},
		{"Deposit Confirmation", "deposit_received", true, []string{"in_app", "email"}},
		{"Withdrawal Processing", "withdrawal_processed", true, []string{"in_app", "email", "push"}},
		{"Price Alert Notification", "price_alert", true, []string{"in_app", "push"}},
		{"System Maintenance Notice", "system_maintenance", true, []string{"in_app", "email"}},
		{"Account Update", "account_update", true, []string{"in_app", "email"}},
		{"Security Alert", "security_alert", true, []string{"in_app", "email", "sms", "push"}},
		{"Promotion Announcement", "promotion", true, []string{"in_app", "email"}},
		{"Large Trade Alert", "trade_executed", true, []string{"in_app", "email"}},
		{"Low Balance Warning", "account_update", false, []string{"in_app", "push"}},
		{"VIP Client Welcome", "account_update", true, []string{"in_app", "email", "push"}},
		{"Trading Competition", "promotion", true, []string{"in_app", "email"}},
		{"Risk Threshold Breach", "margin_call", true, []string{"in_app", "email", "sms"}},
		{"Custom Broadcast", "custom", true, []string{"in_app", "email", "push"}},
	}

	for i, cfg := range ruleConfigs {
		template := fmt.Sprintf("{{.title}} - {{.message}}")
		conditions := map[string]string{
			"min_amount": "100",
			"client_tier": "standard",
		}

		if cfg.event == "margin_call" {
			conditions["margin_level"] = "80"
		}

		rule := &NotificationRule{
			ID:           s.nextRuleID,
			Name:         cfg.name,
			TriggerEvent: cfg.event,
			Conditions:   conditions,
			Channels:     cfg.channels,
			Template:     template,
			IsActive:     cfg.active,
			CreatedAt:    now.Add(-time.Duration(30-i) * 24 * time.Hour),
			UpdatedAt:    now.Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour),
		}

		s.rules[s.nextRuleID] = rule
		s.nextRuleID++
	}
}

func (s *NotificationCenterService) generateNotificationContent(notifType string) (string, string, string) {
	switch notifType {
	case "trade_executed":
		return "Trade Executed", "Your EUR/USD order has been executed at 1.0850", "/trades"
	case "margin_call":
		return "Margin Call Warning", "Your margin level has fallen to 80%. Please add funds.", "/account/deposit"
	case "deposit_received":
		return "Deposit Confirmed", "Your deposit of $1,000 has been credited to your account.", "/account/balance"
	case "withdrawal_processed":
		return "Withdrawal Processed", "Your withdrawal request of $500 is being processed.", "/account/transactions"
	case "price_alert":
		return "Price Alert", "EUR/USD has reached your target price of 1.0900", "/charts"
	case "system_maintenance":
		return "Scheduled Maintenance", "Platform maintenance scheduled for tonight 2:00-4:00 AM UTC", ""
	case "account_update":
		return "Account Updated", "Your account profile has been successfully updated.", "/account/profile"
	case "security_alert":
		return "Security Alert", "New login detected from unknown device. Was this you?", "/account/security"
	case "promotion":
		return "Special Offer", "Get 20% bonus on your next deposit! Limited time offer.", "/promotions"
	default:
		return "Notification", "You have a new notification", ""
	}
}

func (s *NotificationCenterService) GetNotifications(notifType, severity, status, channel string, limit, offset int) ([]Notification, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]Notification, 0)
	for _, notif := range s.notifications {
		if notifType != "" && notif.Type != notifType {
			continue
		}
		if severity != "" && notif.Severity != severity {
			continue
		}
		if status != "" && notif.Status != status {
			continue
		}
		if channel != "" && notif.Channel != channel {
			continue
		}
		filtered = append(filtered, notif)
	}

	total := len(filtered)

	if offset >= total {
		return []Notification{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *NotificationCenterService) GetClientNotifications(clientID string) []Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Notification, 0)
	for _, notif := range s.notifications {
		if notif.RecipientID == clientID {
			result = append(result, notif)
		}
	}

	return result
}

func (s *NotificationCenterService) SendNotification(req SendNotificationRequest) (*Notification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Send to each channel
	for _, channel := range req.Channels {
		notif := Notification{
			ID:            s.nextNotifID,
			Type:          req.Type,
			Title:         req.Title,
			Message:       req.Message,
			Severity:      req.Severity,
			RecipientType: req.RecipientType,
			RecipientID:   req.RecipientID,
			Channel:       channel,
			Status:        "sent",
			CreatedAt:     time.Now(),
			ActionURL:     req.ActionURL,
		}

		s.notifications = append([]Notification{notif}, s.notifications...)
		s.nextNotifID++
	}

	// Return first notification as reference
	return &s.notifications[0], nil
}

func (s *NotificationCenterService) GetRules() []NotificationRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]NotificationRule, 0, len(s.rules))
	for _, rule := range s.rules {
		result = append(result, *rule)
	}

	// Sort by ID
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

func (s *NotificationCenterService) CreateRule(rule NotificationRule) (*NotificationRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule.ID = s.nextRuleID
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	s.rules[s.nextRuleID] = &rule
	s.nextRuleID++

	return &rule, nil
}

func (s *NotificationCenterService) UpdateRule(id int64, updates map[string]interface{}) (*NotificationRule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.rules[id]
	if !exists {
		return nil, fmt.Errorf("rule not found")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		rule.Name = name
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		rule.IsActive = isActive
	}
	if conditions, ok := updates["conditions"].(map[string]string); ok {
		rule.Conditions = conditions
	}
	if channels, ok := updates["channels"].([]string); ok {
		rule.Channels = channels
	}

	rule.UpdatedAt = time.Now()

	return rule, nil
}

func (s *NotificationCenterService) GetStats() *NotificationStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &NotificationStats{
		ByChannel:  make(map[string]int),
		BySeverity: make(map[string]int),
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		Generated:  time.Now(),
	}

	totalRead := 0
	for _, notif := range s.notifications {
		stats.TotalSent++

		if notif.ReadAt != nil {
			totalRead++
		}

		stats.ByChannel[notif.Channel]++
		stats.BySeverity[notif.Severity]++
		stats.ByType[notif.Type]++
		stats.ByStatus[notif.Status]++
	}

	stats.TotalRead = totalRead
	if stats.TotalSent > 0 {
		stats.ReadRate = float64(totalRead) / float64(stats.TotalSent) * 100.0
	}

	return stats
}

func (s *NotificationCenterService) BroadcastNotification(req BroadcastRequest) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	recipients := s.clientIDs
	if req.TargetType == "segment" && req.SegmentID != "" {
		// Filter by segment (simplified - just take first 20 clients)
		recipients = s.clientIDs[:20]
	}

	sent := 0
	for _, clientID := range recipients {
		for _, channel := range req.Channels {
			notif := Notification{
				ID:            s.nextNotifID,
				Type:          req.Type,
				Title:         req.Title,
				Message:       req.Message,
				Severity:      req.Severity,
				RecipientType: "client",
				RecipientID:   clientID,
				Channel:       channel,
				Status:        "sent",
				CreatedAt:     time.Now(),
				ActionURL:     req.ActionURL,
			}

			s.notifications = append([]Notification{notif}, s.notifications...)
			s.nextNotifID++
			sent++
		}
	}

	return sent, nil
}

// ============================================
// Notification Center Handler
// ============================================

type NotificationCenterHandler struct {
	service     *NotificationCenterService
	authService *AuthService
}

func NewNotificationCenterHandler(service *NotificationCenterService, authService *AuthService) *NotificationCenterHandler {
	return &NotificationCenterHandler{
		service:     service,
		authService: authService,
	}
}

func (h *NotificationCenterHandler) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse query parameters
	notifType := r.URL.Query().Get("type")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")
	channel := r.URL.Query().Get("channel")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsed
		}
	}

	notifications, total := h.service.GetNotifications(notifType, severity, status, channel, limit, offset)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

func (h *NotificationCenterHandler) HandleGetClientNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Extract client ID from URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/notifications/client/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Client ID not specified", http.StatusBadRequest)
		return
	}
	clientID := parts[0]

	notifications := h.service.GetClientNotifications(clientID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"notifications": notifications,
		"client_id":     clientID,
		"total":         len(notifications),
	})
}

func (h *NotificationCenterHandler) HandleSendNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req SendNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	notif, err := h.service.SendNotification(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"notification": notif,
		"channels_sent": len(req.Channels),
	})
}

func (h *NotificationCenterHandler) HandleGetRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	rules := h.service.GetRules()

	// Count active/inactive
	activeCount := 0
	for _, rule := range rules {
		if rule.IsActive {
			activeCount++
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules":         rules,
		"total":         len(rules),
		"active":        activeCount,
		"inactive":      len(rules) - activeCount,
	})
}

func (h *NotificationCenterHandler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var rule NotificationRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	createdRule, err := h.service.CreateRule(rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rule":    createdRule,
	})
}

func (h *NotificationCenterHandler) HandleUpdateRule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Extract rule ID from URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/notifications/rules/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Rule ID not specified", http.StatusBadRequest)
		return
	}

	ruleID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	updatedRule, err := h.service.UpdateRule(ruleID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"rule":    updatedRule,
	})
}

func (h *NotificationCenterHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	stats := h.service.GetStats()

	json.NewEncoder(w).Encode(stats)
}

func (h *NotificationCenterHandler) HandleBroadcast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req BroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	sent, err := h.service.BroadcastNotification(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"sent_count": sent,
		"target":     req.TargetType,
	})
}
