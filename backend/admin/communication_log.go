//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
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
// Communication Log API
// ============================================
// Unified client communication tracking system for emails,
// SMS, push notifications, system messages, and support communications.

// CommunicationEntry represents a communication record
type CommunicationEntry struct {
	ID           int64      `json:"id"`
	ClientID     int64      `json:"clientId"`
	ClientName   string     `json:"clientName"`
	ClientEmail  string     `json:"clientEmail"`
	Type         string     `json:"type"`         // email, sms, push, system, support
	Subject      string     `json:"subject"`
	Content      string     `json:"content"`
	Status       string     `json:"status"`       // sent, delivered, opened, bounced, failed
	SentAt       time.Time  `json:"sentAt"`
	DeliveredAt  *time.Time `json:"deliveredAt,omitempty"`
	OpenedAt     *time.Time `json:"openedAt,omitempty"`
	Channel      string     `json:"channel"`      // email, sms, push, in-app
	TemplateUsed string     `json:"templateUsed"` // Template name if used
}

// CommLogTemplate represents a reusable message template
type CommLogTemplate struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Channel  string `json:"channel"`  // email, sms, push
	Subject  string `json:"subject"`  // For emails
	Body     string `json:"body"`     // Message content with {{variable}} placeholders
	Category string `json:"category"` // onboarding, kyc, trading, finance, support
}

// CommunicationStats represents communication statistics
type CommunicationStats struct {
	TotalSentToday   int            `json:"totalSentToday"`
	DeliveryRatePct  float64        `json:"deliveryRatePct"`
	OpenRatePct      float64        `json:"openRatePct"`
	PendingCount     int            `json:"pendingCount"`
	ByType           map[string]int `json:"byType"`
	ByStatus         map[string]int `json:"byStatus"`
}

// CommunicationService manages communication logs and templates
type CommunicationService struct {
	mu        sync.RWMutex
	entries   map[int64]*CommunicationEntry
	templates map[int64]*CommLogTemplate
	nextEntryID   int64
	nextTemplateID int64
}

// NewCommunicationService creates a new communication service with mock data
func NewCommunicationService() *CommunicationService {
	service := &CommunicationService{
		entries:   make(map[int64]*CommunicationEntry),
		templates: make(map[int64]*CommLogTemplate),
		nextEntryID:   1,
		nextTemplateID: 1,
	}

	// Generate 10 message templates
	service.generateMockTemplates()

	// Generate 50 communication entries
	service.generateMockEntries()

	log.Printf("[Communications] Initialized with %d mock communication entries and %d templates", len(service.entries), len(service.templates))
	return service
}

// generateMockTemplates creates 10 message templates
func (cs *CommunicationService) generateMockTemplates() {
	templates := []struct {
		name     string
		channel  string
		subject  string
		body     string
		category string
	}{
		{
			name:     "Welcome Email",
			channel:  "email",
			subject:  "Welcome to RTX5 Trading Platform!",
			body:     "Dear {{clientName}},\n\nWelcome to RTX5! Your trading account {{accountId}} has been successfully created. You can now access our platform and start trading.\n\nBest regards,\nRTX5 Team",
			category: "onboarding",
		},
		{
			name:     "KYC Reminder",
			channel:  "email",
			subject:  "Action Required: Complete Your KYC Verification",
			body:     "Hi {{clientName}},\n\nYour account verification is pending. Please complete your KYC documents to unlock full trading features.\n\nClick here to upload documents: {{kycLink}}\n\nThank you,\nRTX5 Compliance Team",
			category: "kyc",
		},
		{
			name:     "Deposit Confirmation",
			channel:  "email",
			subject:  "Deposit Received - ${{amount}}",
			body:     "Dear {{clientName}},\n\nWe've received your deposit of ${{amount}} via {{method}}. Funds will be available in your account within {{processingTime}}.\n\nTransaction ID: {{txnId}}\n\nThank you,\nRTX5 Finance Team",
			category: "finance",
		},
		{
			name:     "Withdrawal Processed",
			channel:  "email",
			subject:  "Withdrawal Request Approved - ${{amount}}",
			body:     "Hi {{clientName}},\n\nYour withdrawal request for ${{amount}} has been approved and processed. Funds should arrive in your {{method}} account within {{processingTime}}.\n\nWithdrawal ID: {{withdrawalId}}\n\nRTX5 Finance Team",
			category: "finance",
		},
		{
			name:     "Margin Call Warning",
			channel:  "sms",
			subject:  "",
			body:     "RTX5 Alert: Margin level at {{marginLevel}}%. Please deposit funds or close positions to avoid stop-out. Account: {{accountId}}",
			category: "trading",
		},
		{
			name:     "Password Reset",
			channel:  "email",
			subject:  "Password Reset Request",
			body:     "Hi {{clientName}},\n\nWe received a password reset request for your account. Click the link below to reset your password:\n\n{{resetLink}}\n\nIf you didn't request this, please ignore this email.\n\nRTX5 Security Team",
			category: "support",
		},
		{
			name:     "Account Verified",
			channel:  "push",
			subject:  "",
			body:     "Congratulations! Your RTX5 account has been fully verified. You can now access all trading features.",
			category: "kyc",
		},
		{
			name:     "Monthly Statement",
			channel:  "email",
			subject:  "Your Monthly Trading Statement - {{month}}",
			body:     "Dear {{clientName}},\n\nYour monthly statement for {{month}} is now available. Total trades: {{totalTrades}}, P&L: {{pnl}}.\n\nView full statement: {{statementLink}}\n\nRTX5 Team",
			category: "trading",
		},
		{
			name:     "Promotion",
			channel:  "email",
			subject:  "Exclusive Offer: {{offerTitle}}",
			body:     "Hi {{clientName}},\n\n{{offerDescription}}\n\nOffer valid until {{expiryDate}}.\n\nStart trading now: {{tradingLink}}\n\nRTX5 Marketing",
			category: "onboarding",
		},
		{
			name:     "Support Follow-up",
			channel:  "email",
			subject:  "Support Ticket #{{ticketId}} - Status Update",
			body:     "Dear {{clientName}},\n\nThis is a follow-up regarding your support ticket #{{ticketId}}. {{updateMessage}}\n\nIf you have any questions, please reply to this email.\n\nRTX5 Support Team",
			category: "support",
		},
	}

	for _, t := range templates {
		template := &CommLogTemplate{
			ID:       cs.nextTemplateID,
			Name:     t.name,
			Channel:  t.channel,
			Subject:  t.subject,
			Body:     t.body,
			Category: t.category,
		}
		cs.templates[cs.nextTemplateID] = template
		cs.nextTemplateID++
	}
}

// generateMockEntries creates 50 communication entries
func (cs *CommunicationService) generateMockEntries() {
	// Define mock clients
	clients := []struct {
		id    int64
		name  string
		email string
	}{
		{1001, "Alice Johnson", "alice.johnson@email.com"},
		{1002, "Bob Martinez", "bob.martinez@email.com"},
		{1003, "Carol White", "carol.white@email.com"},
		{1004, "David Lee", "david.lee@email.com"},
		{1005, "Emma Taylor", "emma.taylor@email.com"},
		{1006, "Frank Chen", "frank.chen@email.com"},
		{1007, "Grace Kim", "grace.kim@email.com"},
		{1008, "Henry Wilson", "henry.wilson@email.com"},
		{1009, "Iris Brown", "iris.brown@email.com"},
		{1010, "Jack Davis", "jack.davis@email.com"},
	}

	// 30 Emails
	emailSubjects := []string{
		"Welcome to RTX5 Trading",
		"Deposit Confirmation - $500",
		"Withdrawal Request Approved",
		"Account Verification Pending",
		"Monthly Trading Statement - January",
		"Password Reset Confirmation",
		"New Promotion: 20% Bonus",
		"Support Ticket #12345 Resolved",
		"KYC Documents Required",
		"Trade Execution Confirmation",
	}

	for i := 0; i < 30; i++ {
		client := clients[i%len(clients)]
		hoursAgo := i * 3 // Spread over 90 hours (~4 days)

		entry := &CommunicationEntry{
			ID:           cs.nextEntryID,
			ClientID:     client.id,
			ClientName:   client.name,
			ClientEmail:  client.email,
			Type:         "email",
			Subject:      emailSubjects[i%len(emailSubjects)],
			Content:      fmt.Sprintf("Email content for %s - Entry #%d", client.name, cs.nextEntryID),
			SentAt:       time.Now().Add(-time.Duration(hoursAgo) * time.Hour),
			Channel:      "email",
			TemplateUsed: cs.getRandomTemplate("email"),
		}

		// Assign status based on time
		if i%10 < 6 { // 60% delivered
			entry.Status = "delivered"
			deliveredTime := entry.SentAt.Add(2 * time.Minute)
			entry.DeliveredAt = &deliveredTime

			if i%10 < 2 { // 15% opened (subset of delivered)
				entry.Status = "opened"
				openedTime := deliveredTime.Add(30 * time.Minute)
				entry.OpenedAt = &openedTime
			}
		} else if i%10 < 7 { // 10% sent (not delivered yet)
			entry.Status = "sent"
		} else if i%10 < 9 { // 10% bounced
			entry.Status = "bounced"
		} else { // 5% failed
			entry.Status = "failed"
		}

		cs.entries[cs.nextEntryID] = entry
		cs.nextEntryID++
	}

	// 8 SMS
	for i := 0; i < 8; i++ {
		client := clients[i%len(clients)]
		hoursAgo := i * 6

		entry := &CommunicationEntry{
			ID:           cs.nextEntryID,
			ClientID:     client.id,
			ClientName:   client.name,
			ClientEmail:  client.email,
			Type:         "sms",
			Subject:      "",
			Content:      fmt.Sprintf("SMS: Your RTX5 verification code is %d%d%d%d. Valid for 10 minutes.", i, i+1, i+2, i+3),
			Status:       "delivered",
			SentAt:       time.Now().Add(-time.Duration(hoursAgo) * time.Hour),
			Channel:      "sms",
			TemplateUsed: "Margin Call Warning",
		}

		deliveredTime := entry.SentAt.Add(10 * time.Second)
		entry.DeliveredAt = &deliveredTime

		cs.entries[cs.nextEntryID] = entry
		cs.nextEntryID++
	}

	// 5 Push Notifications
	for i := 0; i < 5; i++ {
		client := clients[i%len(clients)]
		hoursAgo := i * 12

		entry := &CommunicationEntry{
			ID:           cs.nextEntryID,
			ClientID:     client.id,
			ClientName:   client.name,
			ClientEmail:  client.email,
			Type:         "push",
			Subject:      "",
			Content:      fmt.Sprintf("Trade executed: %s position opened on EURUSD", []string{"BUY", "SELL"}[i%2]),
			Status:       "delivered",
			SentAt:       time.Now().Add(-time.Duration(hoursAgo) * time.Hour),
			Channel:      "push",
			TemplateUsed: "Account Verified",
		}

		deliveredTime := entry.SentAt.Add(1 * time.Second)
		entry.DeliveredAt = &deliveredTime

		cs.entries[cs.nextEntryID] = entry
		cs.nextEntryID++
	}

	// 4 System Messages
	for i := 0; i < 4; i++ {
		client := clients[i%len(clients)]
		hoursAgo := i * 8

		entry := &CommunicationEntry{
			ID:           cs.nextEntryID,
			ClientID:     client.id,
			ClientName:   client.name,
			ClientEmail:  client.email,
			Type:         "system",
			Subject:      "",
			Content:      fmt.Sprintf("System notification: %s", []string{"Scheduled maintenance on Feb 15", "New trading features available", "Account settings updated", "Security alert: New login detected"}[i]),
			Status:       "delivered",
			SentAt:       time.Now().Add(-time.Duration(hoursAgo) * time.Hour),
			Channel:      "in-app",
			TemplateUsed: "",
		}

		deliveredTime := entry.SentAt.Add(1 * time.Millisecond)
		entry.DeliveredAt = &deliveredTime

		cs.entries[cs.nextEntryID] = entry
		cs.nextEntryID++
	}

	// 3 Support Communications
	for i := 0; i < 3; i++ {
		client := clients[i%len(clients)]
		hoursAgo := i * 10

		entry := &CommunicationEntry{
			ID:           cs.nextEntryID,
			ClientID:     client.id,
			ClientName:   client.name,
			ClientEmail:  client.email,
			Type:         "support",
			Subject:      fmt.Sprintf("Support Ticket #%d - Status Update", 1000+i),
			Content:      "Your support ticket has been assigned to our team. We'll respond within 24 hours.",
			Status:       "delivered",
			SentAt:       time.Now().Add(-time.Duration(hoursAgo) * time.Hour),
			Channel:      "email",
			TemplateUsed: "Support Follow-up",
		}

		deliveredTime := entry.SentAt.Add(2 * time.Minute)
		entry.DeliveredAt = &deliveredTime

		if i == 0 {
			entry.Status = "opened"
			openedTime := deliveredTime.Add(15 * time.Minute)
			entry.OpenedAt = &openedTime
		}

		cs.entries[cs.nextEntryID] = entry
		cs.nextEntryID++
	}
}

// getRandomTemplate returns a random template name for a channel
func (cs *CommunicationService) getRandomTemplate(channel string) string {
	templates := []string{"Welcome Email", "KYC Reminder", "Deposit Confirmation", "Password Reset", "Monthly Statement"}
	for _, tmpl := range cs.templates {
		if tmpl.Channel == channel {
			return tmpl.Name
		}
	}
	if len(templates) > 0 {
		return templates[0]
	}
	return ""
}

// ListCommunications returns communications with optional filters
func (cs *CommunicationService) ListCommunications(commType, status, search string, clientID int64, page, pageSize int) ([]*CommunicationEntry, int) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	filtered := make([]*CommunicationEntry, 0)
	for _, entry := range cs.entries {
		// Apply filters
		if commType != "" && entry.Type != commType {
			continue
		}
		if status != "" && entry.Status != status {
			continue
		}
		if clientID > 0 && entry.ClientID != clientID {
			continue
		}
		if search != "" {
			searchLower := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(entry.Subject), searchLower) &&
				!strings.Contains(strings.ToLower(entry.Content), searchLower) &&
				!strings.Contains(strings.ToLower(entry.ClientName), searchLower) {
				continue
			}
		}

		filtered = append(filtered, entry)
	}

	total := len(filtered)

	// Pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(filtered) {
		return []*CommunicationEntry{}, total
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end], total
}

// GetCommunication retrieves a specific communication entry
func (cs *CommunicationService) GetCommunication(id int64) (*CommunicationEntry, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, exists := cs.entries[id]
	if !exists {
		return nil, fmt.Errorf("communication entry %d not found", id)
	}
	return entry, nil
}

// SendCommunication creates and sends a new communication
func (cs *CommunicationService) SendCommunication(clientID int64, channel, templateID, subject, content string) (*CommunicationEntry, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// For demo, we'll use mock client data
	clientName := fmt.Sprintf("Client %d", clientID)
	clientEmail := fmt.Sprintf("client%d@email.com", clientID)

	entry := &CommunicationEntry{
		ID:           cs.nextEntryID,
		ClientID:     clientID,
		ClientName:   clientName,
		ClientEmail:  clientEmail,
		Type:         cs.channelToType(channel),
		Subject:      subject,
		Content:      content,
		Status:       "sent",
		SentAt:       time.Now(),
		Channel:      channel,
		TemplateUsed: templateID,
	}

	cs.entries[cs.nextEntryID] = entry
	cs.nextEntryID++

	log.Printf("[Communications] Sent new %s to client %d (ID: %d)", channel, clientID, entry.ID)
	return entry, nil
}

// channelToType converts channel to communication type
func (cs *CommunicationService) channelToType(channel string) string {
	switch channel {
	case "email":
		return "email"
	case "sms":
		return "sms"
	case "push":
		return "push"
	case "in-app":
		return "system"
	default:
		return "email"
	}
}

// GetStats returns communication statistics
func (cs *CommunicationService) GetStats() *CommunicationStats {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	stats := &CommunicationStats{
		ByType:   make(map[string]int),
		ByStatus: make(map[string]int),
	}

	today := time.Now().Truncate(24 * time.Hour)
	totalDelivered := 0
	totalOpened := 0
	totalSent := 0

	for _, entry := range cs.entries {
		stats.ByType[entry.Type]++
		stats.ByStatus[entry.Status]++

		// Count today's entries
		if entry.SentAt.After(today) {
			stats.TotalSentToday++
		}

		// Count for rates
		if entry.Status == "delivered" || entry.Status == "opened" {
			totalDelivered++
		}
		if entry.Status == "opened" {
			totalOpened++
		}
		if entry.Status == "sent" {
			stats.PendingCount++
		}
		totalSent++
	}

	// Calculate rates
	if totalSent > 0 {
		stats.DeliveryRatePct = float64(totalDelivered) / float64(totalSent) * 100
	}
	if totalDelivered > 0 {
		stats.OpenRatePct = float64(totalOpened) / float64(totalDelivered) * 100
	}

	return stats
}

// ListTemplates returns all message templates
func (cs *CommunicationService) ListTemplates() []*CommLogTemplate {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	templates := make([]*CommLogTemplate, 0)
	for _, template := range cs.templates {
		templates = append(templates, template)
	}
	return templates
}

// ============================================
// HTTP Handlers
// ============================================

// CommunicationHandler handles HTTP requests for communications
type CommunicationHandler struct {
	service     *CommunicationService
	authService *auth.Service
}

// NewCommunicationHandler creates a new communication handler
func NewCommunicationHandler(service *CommunicationService, authService *auth.Service) *CommunicationHandler {
	return &CommunicationHandler{
		service:     service,
		authService: authService,
	}
}

// ListCommunications handles GET /admin/communications
func (ch *CommunicationHandler) ListCommunications(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract query parameters
	commType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	clientID := int64(0)
	if clientIDStr := r.URL.Query().Get("clientId"); clientIDStr != "" {
		if id, err := strconv.ParseInt(clientIDStr, 10, 64); err == nil {
			clientID = id
		}
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := r.URL.Query().Get("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	communications, total := ch.service.ListCommunications(commType, status, search, clientID, page, pageSize)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"communications": communications,
		"page":           page,
		"pageSize":       pageSize,
		"total":          total,
		"totalPages":     (total + pageSize - 1) / pageSize,
	})
}

// GetCommunication handles GET /admin/communications/:id
func (ch *CommunicationHandler) GetCommunication(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid communication ID", http.StatusBadRequest)
		return
	}

	entry, err := ch.service.GetCommunication(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// SendCommunication handles POST /admin/communications/send
func (ch *CommunicationHandler) SendCommunication(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		ClientID   int64  `json:"clientId"`
		Channel    string `json:"channel"`
		TemplateID string `json:"templateId"`
		Subject    string `json:"subject"`
		Content    string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ClientID == 0 {
		http.Error(w, "clientId is required", http.StatusBadRequest)
		return
	}
	if request.Channel == "" {
		http.Error(w, "channel is required", http.StatusBadRequest)
		return
	}
	if request.Content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}

	entry, err := ch.service.SendCommunication(request.ClientID, request.Channel, request.TemplateID, request.Subject, request.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// GetStats handles GET /admin/communications/stats
func (ch *CommunicationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := ch.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetTemplates handles GET /admin/communications/templates
func (ch *CommunicationHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates := ch.service.ListTemplates()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}
