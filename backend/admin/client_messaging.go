package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================
// Data Structures
// ============================================

// Message represents a single message in the system
type Message struct {
	ID             int64     `json:"id"`
	ThreadID       int64     `json:"thread_id"`
	ClientID       int64     `json:"client_id"`
	ClientName     string    `json:"client_name"`
	SenderType     string    `json:"sender_type"`   // "client" or "admin"
	SenderName     string    `json:"sender_name"`
	Subject        string    `json:"subject"`
	Body           string    `json:"body"`
	Priority       string    `json:"priority"`      // "low", "normal", "high", "urgent"
	Status         string    `json:"status"`        // "unread", "read", "archived"
	AttachmentURL  string    `json:"attachment_url,omitempty"`
	SentAt         time.Time `json:"sent_at"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
}

// MessageThread represents a conversation thread with a client
type MessageThread struct {
	ThreadID        int64     `json:"thread_id"`
	ClientID        int64     `json:"client_id"`
	ClientName      string    `json:"client_name"`
	Subject         string    `json:"subject"`
	MessageCount    int       `json:"message_count"`
	UnreadCount     int       `json:"unread_count"`
	LastMessage     string    `json:"last_message"`
	LastMessageAt   time.Time `json:"last_message_at"`
	LastMessageFrom string    `json:"last_message_from"` // "client" or "admin"
	Messages        []Message `json:"messages"`
}

// MessageTemplate represents a reusable message template
type MessageTemplate struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`        // "welcome", "margin_warning", "kyc_reminder", "account_update", "custom"
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	Variables   []string  `json:"variables"`   // e.g., ["{{client_name}}", "{{account_balance}}"]
	IsActive    bool      `json:"is_active"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MessagingAnnouncement represents a platform-wide announcement for client messaging
type MessagingAnnouncement struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Priority     string    `json:"priority"`      // "info", "warning", "critical"
	TargetGroups []string  `json:"target_groups"` // group IDs or "all"
	CreatedBy    string    `json:"created_by"`
	IsActive     bool      `json:"is_active"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	ViewCount    int       `json:"view_count"`
}

// SendMessageRequest represents request to send a message
type SendMessageRequest struct {
	ClientID      int64  `json:"client_id"`
	Subject       string `json:"subject"`
	Body          string `json:"body"`
	Priority      string `json:"priority"`
	AttachmentURL string `json:"attachment_url,omitempty"`
}

// BulkMessageRequest represents request to send bulk messages
type BulkMessageRequest struct {
	GroupID   int64   `json:"group_id,omitempty"`
	ClientIDs []int64 `json:"client_ids,omitempty"`
	Subject   string  `json:"subject"`
	Body      string  `json:"body"`
	Priority  string  `json:"priority"`
}

// CreateMessagingAnnouncementRequest represents request to create announcement
type CreateMessagingAnnouncementRequest struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Priority     string   `json:"priority"`
	TargetGroups []string `json:"target_groups"`
	ExpiresAt    string   `json:"expires_at"` // ISO 8601 format
}

// MessagingStats represents messaging statistics
type MessagingStats struct {
	TotalMessages        int                `json:"total_messages"`
	TotalThreads         int                `json:"total_threads"`
	UnreadCount          int                `json:"unread_count"`
	AvgResponseTimeHours float64            `json:"avg_response_time_hours"`
	MessagesByPriority   map[string]int     `json:"messages_by_priority"`
	MessagesByStatus     map[string]int     `json:"messages_by_status"`
	TopClients           []ClientMessageStat `json:"top_clients"`
	Last30Days           []DailyMessageStat  `json:"last_30_days"`
}

// ClientMessageStat represents message statistics per client
type ClientMessageStat struct {
	ClientID     int64  `json:"client_id"`
	ClientName   string `json:"client_name"`
	MessageCount int    `json:"message_count"`
	UnreadCount  int    `json:"unread_count"`
}

// DailyMessageStat represents daily message statistics
type DailyMessageStat struct {
	Date         string `json:"date"`
	MessageCount int    `json:"message_count"`
	UnreadCount  int    `json:"unread_count"`
}

// ============================================
// Service
// ============================================

// ClientMessagingService manages client messaging and communications
type ClientMessagingService struct {
	mu              sync.RWMutex
	messages        map[int64]*Message
	threads         map[int64]*MessageThread    // key: threadID
	clientThreads   map[int64]int64             // key: clientID, value: threadID
	templates       map[int64]*MessageTemplate
	announcements   map[int64]*MessagingAnnouncement
	nextMessageID   int64
	nextThreadID    int64
	nextTemplateID  int64
	nextMessagingAnnouncementID int64
}

// NewClientMessagingService creates a new client messaging service with mock data
func NewClientMessagingService() *ClientMessagingService {
	s := &ClientMessagingService{
		messages:        make(map[int64]*Message),
		threads:         make(map[int64]*MessageThread),
		clientThreads:   make(map[int64]int64),
		templates:       make(map[int64]*MessageTemplate),
		announcements:   make(map[int64]*MessagingAnnouncement),
		nextMessageID:   501,
		nextThreadID:    51,
		nextTemplateID:  11,
		nextMessagingAnnouncementID: 21,
	}

	// Initialize 10 message templates
	s.initTemplates()

	// Initialize 20 announcements
	s.initMessagingAnnouncements()

	// Generate 500 messages across 100 clients (creates ~50 threads)
	s.generateMessages(500, 100)

	log.Printf("[ClientMessagingService] Initialized with %d messages, %d threads, %d templates, %d announcements",
		len(s.messages), len(s.threads), len(s.templates), len(s.announcements))

	return s
}

func (s *ClientMessagingService) initTemplates() {
	now := time.Now()

	templates := []*MessageTemplate{
		{
			ID:       1,
			Name:     "Welcome Message",
			Type:     "welcome",
			Subject:  "Welcome to RTX5 Trading Platform",
			Body:     "Dear {{client_name}},\n\nWelcome to RTX5! Your account has been successfully created. Your account number is {{account_id}}.\n\nBest regards,\nRTX5 Support Team",
			Variables: []string{"{{client_name}}", "{{account_id}}"},
			IsActive:  true,
			UsageCount: 150,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-30 * 24 * time.Hour),
		},
		{
			ID:       2,
			Name:     "Margin Call Warning",
			Type:     "margin_warning",
			Subject:  "URGENT: Margin Call Alert",
			Body:     "Dear {{client_name}},\n\nYour account margin level has dropped to {{margin_level}}%. Please deposit funds or close positions to avoid liquidation.\n\nCurrent equity: {{equity}}\nUsed margin: {{used_margin}}\n\nRTX5 Risk Team",
			Variables: []string{"{{client_name}}", "{{margin_level}}", "{{equity}}", "{{used_margin}}"},
			IsActive:  true,
			UsageCount: 85,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-15 * 24 * time.Hour),
		},
		{
			ID:       3,
			Name:     "KYC Reminder",
			Type:     "kyc_reminder",
			Subject:  "Action Required: Complete Your KYC Verification",
			Body:     "Dear {{client_name}},\n\nYour KYC verification is pending. Please upload the required documents within 7 days to continue trading.\n\nRequired documents:\n- Government ID\n- Proof of address\n\nRTX5 Compliance Team",
			Variables: []string{"{{client_name}}"},
			IsActive:  true,
			UsageCount: 65,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-20 * 24 * time.Hour),
		},
		{
			ID:       4,
			Name:     "Account Update Notification",
			Type:     "account_update",
			Subject:  "Your Account Has Been Updated",
			Body:     "Dear {{client_name}},\n\nYour account settings have been updated successfully.\n\nChanges:\n{{update_details}}\n\nIf you did not make these changes, please contact support immediately.\n\nRTX5 Support Team",
			Variables: []string{"{{client_name}}", "{{update_details}}"},
			IsActive:  true,
			UsageCount: 120,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-10 * 24 * time.Hour),
		},
		{
			ID:       5,
			Name:     "Withdrawal Approved",
			Type:     "account_update",
			Subject:  "Your Withdrawal Request Has Been Approved",
			Body:     "Dear {{client_name}},\n\nYour withdrawal request of {{amount}} has been approved and processed.\n\nTransaction ID: {{transaction_id}}\nExpected arrival: {{expected_date}}\n\nRTX5 Finance Team",
			Variables: []string{"{{client_name}}", "{{amount}}", "{{transaction_id}}", "{{expected_date}}"},
			IsActive:  true,
			UsageCount: 95,
			CreatedAt: now.Add(-150 * 24 * time.Hour),
			UpdatedAt: now.Add(-25 * 24 * time.Hour),
		},
		{
			ID:       6,
			Name:     "Password Reset",
			Type:     "account_update",
			Subject:  "Password Reset Confirmation",
			Body:     "Dear {{client_name}},\n\nYour password has been successfully reset. If you did not request this change, please contact support immediately.\n\nRTX5 Security Team",
			Variables: []string{"{{client_name}}"},
			IsActive:  true,
			UsageCount: 45,
			CreatedAt: now.Add(-120 * 24 * time.Hour),
			UpdatedAt: now.Add(-40 * 24 * time.Hour),
		},
		{
			ID:       7,
			Name:     "Trading Competition Invitation",
			Type:     "custom",
			Subject:  "You're Invited: Monthly Trading Competition",
			Body:     "Dear {{client_name}},\n\nJoin our monthly trading competition! Compete for prizes up to $10,000.\n\nStart date: {{start_date}}\nPrize pool: {{prize_pool}}\n\nRegister now!\n\nRTX5 Marketing Team",
			Variables: []string{"{{client_name}}", "{{start_date}}", "{{prize_pool}}"},
			IsActive:  true,
			UsageCount: 200,
			CreatedAt: now.Add(-90 * 24 * time.Hour),
			UpdatedAt: now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:       8,
			Name:     "VIP Upgrade Offer",
			Type:     "custom",
			Subject:  "Exclusive: Upgrade to VIP Account",
			Body:     "Dear {{client_name}},\n\nBased on your trading volume, you qualify for a VIP account upgrade with exclusive benefits:\n\n- Lower spreads\n- Dedicated account manager\n- Priority support\n\nContact us to upgrade!\n\nRTX5 VIP Services",
			Variables: []string{"{{client_name}}"},
			IsActive:  true,
			UsageCount: 30,
			CreatedAt: now.Add(-60 * 24 * time.Hour),
			UpdatedAt: now.Add(-12 * 24 * time.Hour),
		},
		{
			ID:       9,
			Name:     "Inactivity Warning",
			Type:     "custom",
			Subject:  "Your Account Has Been Inactive",
			Body:     "Dear {{client_name}},\n\nYour account has been inactive for {{days_inactive}} days. Inactivity fees may apply after 6 months.\n\nLog in to keep your account active.\n\nRTX5 Support Team",
			Variables: []string{"{{client_name}}", "{{days_inactive}}"},
			IsActive:  true,
			UsageCount: 55,
			CreatedAt: now.Add(-45 * 24 * time.Hour),
			UpdatedAt: now.Add(-8 * 24 * time.Hour),
		},
		{
			ID:       10,
			Name:     "Maintenance Notification",
			Type:     "custom",
			Subject:  "Scheduled Platform Maintenance",
			Body:     "Dear {{client_name}},\n\nScheduled maintenance on {{maintenance_date}} from {{start_time}} to {{end_time}}.\n\nServices affected: {{affected_services}}\n\nPlease close positions or set stop-loss orders before maintenance begins.\n\nRTX5 Technical Team",
			Variables: []string{"{{client_name}}", "{{maintenance_date}}", "{{start_time}}", "{{end_time}}", "{{affected_services}}"},
			IsActive:  true,
			UsageCount: 40,
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-3 * 24 * time.Hour),
		},
	}

	for _, tmpl := range templates {
		s.templates[tmpl.ID] = tmpl
	}
}

func (s *ClientMessagingService) initMessagingAnnouncements() {
	now := time.Now()

	priorities := []string{"info", "warning", "critical"}
	titles := []string{
		"New Trading Instruments Available",
		"System Maintenance Scheduled",
		"Updated Margin Requirements",
		"Holiday Trading Hours",
		"Platform Update v2.5 Released",
		"New Mobile App Features",
		"Spread Changes on Major Pairs",
		"Leverage Restrictions Update",
		"Trading Competition Results",
		"Withdrawal Processing Delays",
		"KYC Requirements Updated",
		"New Payment Methods Added",
		"Server Migration Notice",
		"Trading Signals Now Available",
		"Promotional Bonus Extended",
		"Risk Warning: High Volatility",
		"Economic Calendar Updated",
		"API Rate Limits Changed",
		"Account Security Tips",
		"Year-End Trading Schedule",
	}

	for i := 0; i < 20; i++ {
		priority := priorities[i%len(priorities)]
		isActive := rand.Float64() < 0.7 // 70% active

		createdAt := now.Add(time.Duration(-rand.Intn(60)) * 24 * time.Hour)
		expiresAt := createdAt.AddDate(0, 0, 30+rand.Intn(60)) // 30-90 days

		if !isActive {
			expiresAt = now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour)
		}

		targetGroups := []string{"all"}
		if rand.Float64() < 0.3 {
			targetGroups = []string{"VIP", "Premium"}
		}

		s.announcements[int64(i+1)] = &MessagingAnnouncement{
			ID:           int64(i + 1),
			Title:        titles[i],
			Body:         "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Important update regarding " + titles[i] + ". Please review carefully.",
			Priority:     priority,
			TargetGroups: targetGroups,
			CreatedBy:    "admin",
			IsActive:     isActive,
			ExpiresAt:    expiresAt,
			CreatedAt:    createdAt,
			ViewCount:    rand.Intn(500),
		}
	}
}

func (s *ClientMessagingService) generateMessages(count int, numClients int) {
	now := time.Now()
	priorities := []string{"low", "normal", "high", "urgent"}
	statuses := []string{"read", "read", "read", "unread"} // 75% read, 25% unread
	subjects := []string{
		"Question about withdrawal",
		"Unable to login",
		"Spread issue on EURUSD",
		"Leverage change request",
		"Account verification status",
		"Commission charges inquiry",
		"Platform freezing issue",
		"Deposit not credited",
		"Margin call explanation",
		"Trading hours question",
	}

	messageID := int64(1)
	threadID := int64(1)

	// Create messages for random clients
	clientsWithMessages := make(map[int64]bool)
	for i := 0; i < count; i++ {
		clientID := int64(10001 + rand.Intn(numClients))
		clientName := "Client-" + strconv.FormatInt(clientID, 10)

		// Get or create thread for this client
		var thread *MessageThread
		if existingThreadID, exists := s.clientThreads[clientID]; exists {
			thread = s.threads[existingThreadID]
		} else {
			thread = &MessageThread{
				ThreadID:   threadID,
				ClientID:   clientID,
				ClientName: clientName,
				Subject:    subjects[rand.Intn(len(subjects))],
				Messages:   make([]Message, 0),
			}
			s.threads[threadID] = thread
			s.clientThreads[clientID] = threadID
			threadID++
		}

		// Determine sender (70% from client, 30% admin reply)
		senderType := "client"
		senderName := clientName
		if rand.Float64() < 0.3 {
			senderType = "admin"
			senderName = "Support Agent"
		}

		// Create message
		sentAt := now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour)
		status := statuses[rand.Intn(len(statuses))]

		var readAt *time.Time
		if status == "read" {
			readTime := sentAt.Add(time.Duration(rand.Intn(24)) * time.Hour)
			readAt = &readTime
		}

		message := &Message{
			ID:         messageID,
			ThreadID:   thread.ThreadID,
			ClientID:   clientID,
			ClientName: clientName,
			SenderType: senderType,
			SenderName: senderName,
			Subject:    thread.Subject,
			Body:       "Message body " + strconv.FormatInt(messageID, 10) + ". Lorem ipsum dolor sit amet.",
			Priority:   priorities[rand.Intn(len(priorities))],
			Status:     status,
			SentAt:     sentAt,
			ReadAt:     readAt,
		}

		s.messages[messageID] = message
		thread.Messages = append(thread.Messages, *message)

		// Update thread metadata
		thread.MessageCount = len(thread.Messages)
		if status == "unread" && senderType == "client" {
			thread.UnreadCount++
		}
		thread.LastMessage = message.Body
		thread.LastMessageAt = sentAt
		thread.LastMessageFrom = senderType

		clientsWithMessages[clientID] = true
		messageID++
	}
}

// GetInbox returns admin inbox with all messages
func (s *ClientMessagingService) GetInbox(limit int) []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages := make([]Message, 0)
	for _, msg := range s.messages {
		messages = append(messages, *msg)
	}

	// Sort by sent date (most recent first) - simple implementation
	// In production, would use proper sorting

	if limit > 0 && len(messages) > limit {
		messages = messages[:limit]
	}

	return messages
}

// GetThreadByClient returns conversation thread with specific client
func (s *ClientMessagingService) GetThreadByClient(clientID int64) *MessageThread {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if threadID, exists := s.clientThreads[clientID]; exists {
		return s.threads[threadID]
	}
	return nil
}

// SendMessage sends a message to a client
func (s *ClientMessagingService) SendMessage(req SendMessageRequest, senderName string) (*Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	clientName := "Client-" + strconv.FormatInt(req.ClientID, 10)

	// Get or create thread
	var thread *MessageThread
	if threadID, exists := s.clientThreads[req.ClientID]; exists {
		thread = s.threads[threadID]
	} else {
		thread = &MessageThread{
			ThreadID:   s.nextThreadID,
			ClientID:   req.ClientID,
			ClientName: clientName,
			Subject:    req.Subject,
			Messages:   make([]Message, 0),
		}
		s.threads[s.nextThreadID] = thread
		s.clientThreads[req.ClientID] = s.nextThreadID
		s.nextThreadID++
	}

	// Create message
	message := &Message{
		ID:            s.nextMessageID,
		ThreadID:      thread.ThreadID,
		ClientID:      req.ClientID,
		ClientName:    clientName,
		SenderType:    "admin",
		SenderName:    senderName,
		Subject:       req.Subject,
		Body:          req.Body,
		Priority:      req.Priority,
		Status:        "unread",
		AttachmentURL: req.AttachmentURL,
		SentAt:        now,
	}

	s.messages[s.nextMessageID] = message
	thread.Messages = append(thread.Messages, *message)
	thread.MessageCount = len(thread.Messages)
	thread.LastMessage = message.Body
	thread.LastMessageAt = now
	thread.LastMessageFrom = "admin"

	s.nextMessageID++

	return message, nil
}

// GetTemplates returns all message templates
func (s *ClientMessagingService) GetTemplates() []*MessageTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*MessageTemplate, 0, len(s.templates))
	for _, tmpl := range s.templates {
		templates = append(templates, tmpl)
	}
	return templates
}

// SendBulkMessage sends message to multiple clients
func (s *ClientMessagingService) SendBulkMessage(req BulkMessageRequest, senderName string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine target client IDs
	var targetIDs []int64
	if len(req.ClientIDs) > 0 {
		targetIDs = req.ClientIDs
	} else if req.GroupID > 0 {
		// Mock: select clients based on group (simplified)
		for i := 0; i < 50; i++ {
			targetIDs = append(targetIDs, int64(10001+i))
		}
	}

	sentCount := 0
	now := time.Now()

	for _, clientID := range targetIDs {
		clientName := "Client-" + strconv.FormatInt(clientID, 10)

		// Get or create thread
		var thread *MessageThread
		if threadID, exists := s.clientThreads[clientID]; exists {
			thread = s.threads[threadID]
		} else {
			thread = &MessageThread{
				ThreadID:   s.nextThreadID,
				ClientID:   clientID,
				ClientName: clientName,
				Subject:    req.Subject,
				Messages:   make([]Message, 0),
			}
			s.threads[s.nextThreadID] = thread
			s.clientThreads[clientID] = s.nextThreadID
			s.nextThreadID++
		}

		message := &Message{
			ID:         s.nextMessageID,
			ThreadID:   thread.ThreadID,
			ClientID:   clientID,
			ClientName: clientName,
			SenderType: "admin",
			SenderName: senderName,
			Subject:    req.Subject,
			Body:       req.Body,
			Priority:   req.Priority,
			Status:     "unread",
			SentAt:     now,
		}

		s.messages[s.nextMessageID] = message
		thread.Messages = append(thread.Messages, *message)
		thread.MessageCount = len(thread.Messages)
		thread.LastMessage = message.Body
		thread.LastMessageAt = now
		thread.LastMessageFrom = "admin"

		s.nextMessageID++
		sentCount++
	}

	return sentCount, nil
}

// GetMessagingAnnouncements returns all announcements
func (s *ClientMessagingService) GetMessagingAnnouncements(activeOnly bool) []MessagingAnnouncement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	announcements := make([]MessagingAnnouncement, 0)
	now := time.Now()

	for _, ann := range s.announcements {
		if activeOnly {
			if ann.IsActive && ann.ExpiresAt.After(now) {
				announcements = append(announcements, *ann)
			}
		} else {
			announcements = append(announcements, *ann)
		}
	}

	return announcements
}

// CreateMessagingAnnouncement creates a new announcement
func (s *ClientMessagingService) CreateMessagingAnnouncement(req CreateMessagingAnnouncementRequest, createdBy string) (*MessagingAnnouncement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiresAt, _ := time.Parse(time.RFC3339, req.ExpiresAt)

	announcement := &MessagingAnnouncement{
		ID:           s.nextMessagingAnnouncementID,
		Title:        req.Title,
		Body:         req.Body,
		Priority:     req.Priority,
		TargetGroups: req.TargetGroups,
		CreatedBy:    createdBy,
		IsActive:     true,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		ViewCount:    0,
	}

	s.announcements[announcement.ID] = announcement
	s.nextMessagingAnnouncementID++

	return announcement, nil
}

// GetStats returns messaging statistics
func (s *ClientMessagingService) GetStats() MessagingStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unreadCount := 0
	messagesByPriority := make(map[string]int)
	messagesByStatus := make(map[string]int)
	clientMessageCount := make(map[int64]int)
	clientUnreadCount := make(map[int64]int)

	var totalResponseTime float64
	responseCount := 0

	for _, msg := range s.messages {
		if msg.Status == "unread" && msg.SenderType == "client" {
			unreadCount++
		}
		messagesByPriority[msg.Priority]++
		messagesByStatus[msg.Status]++

		clientMessageCount[msg.ClientID]++
		if msg.Status == "unread" && msg.SenderType == "client" {
			clientUnreadCount[msg.ClientID]++
		}

		// Calculate response time (simplified)
		if msg.SenderType == "admin" && msg.ReadAt != nil {
			responseTime := msg.ReadAt.Sub(msg.SentAt).Hours()
			totalResponseTime += responseTime
			responseCount++
		}
	}

	avgResponseTime := 0.0
	if responseCount > 0 {
		avgResponseTime = totalResponseTime / float64(responseCount)
	}

	// Top clients
	topClients := make([]ClientMessageStat, 0)
	for clientID, count := range clientMessageCount {
		if len(topClients) < 10 {
			clientName := "Client-" + strconv.FormatInt(clientID, 10)
			topClients = append(topClients, ClientMessageStat{
				ClientID:     clientID,
				ClientName:   clientName,
				MessageCount: count,
				UnreadCount:  clientUnreadCount[clientID],
			})
		}
	}

	// Last 30 days (mock data)
	last30Days := make([]DailyMessageStat, 0)
	now := time.Now()
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		last30Days = append(last30Days, DailyMessageStat{
			Date:         date.Format("2006-01-02"),
			MessageCount: 10 + rand.Intn(20),
			UnreadCount:  rand.Intn(5),
		})
	}

	return MessagingStats{
		TotalMessages:        len(s.messages),
		TotalThreads:         len(s.threads),
		UnreadCount:          unreadCount,
		AvgResponseTimeHours: avgResponseTime,
		MessagesByPriority:   messagesByPriority,
		MessagesByStatus:     messagesByStatus,
		TopClients:           topClients,
		Last30Days:           last30Days,
	}
}

// ============================================
// HTTP Handlers
// ============================================

// ClientMessagingHandler handles client messaging HTTP requests
type ClientMessagingHandler struct {
	service     *ClientMessagingService
	authService interface {
		ValidateAdminToken(r *http.Request) (int64, error)
	}
}

// NewClientMessagingHandler creates a new client messaging handler
func NewClientMessagingHandler(service *ClientMessagingService, authService interface {
	ValidateAdminToken(r *http.Request) (int64, error)
}) *ClientMessagingHandler {
	return &ClientMessagingHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetInbox handles GET /admin/messaging/inbox
func (h *ClientMessagingHandler) HandleGetInbox(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query parameter: limit (default: 100)
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	messages := h.service.GetInbox(limit)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(messages)
}

// HandleGetThread handles GET /admin/messaging/threads/:clientId
func (h *ClientMessagingHandler) HandleGetThread(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract client ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/messaging/threads/")
	clientID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	thread := h.service.GetThreadByClient(clientID)
	if thread == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(thread)
}

// HandleSendMessage handles POST /admin/messaging/send
func (h *ClientMessagingHandler) HandleSendMessage(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, use placeholder admin name (in real app, get from token)
	senderName := "Support Agent"

	message, err := h.service.SendMessage(req, senderName)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

// HandleGetTemplates handles GET /admin/messaging/templates
func (h *ClientMessagingHandler) HandleGetTemplates(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates := h.service.GetTemplates()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(templates)
}

// HandleSendBulkMessage handles POST /admin/messaging/bulk
func (h *ClientMessagingHandler) HandleSendBulkMessage(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req BulkMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	senderName := "Support Agent"

	sentCount, err := h.service.SendBulkMessage(req, senderName)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"sent_count": sentCount,
		"message":    "Bulk message sent successfully",
	})
}

// HandleGetMessagingAnnouncements handles GET /admin/messaging/announcements
func (h *ClientMessagingHandler) HandleGetMessagingAnnouncements(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query parameter: active_only (default: true)
	activeOnly := true
	if r.URL.Query().Get("active_only") == "false" {
		activeOnly = false
	}

	announcements := h.service.GetMessagingAnnouncements(activeOnly)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(announcements)
}

// HandleCreateMessagingAnnouncement handles POST /admin/messaging/announcements
func (h *ClientMessagingHandler) HandleCreateMessagingAnnouncement(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateMessagingAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdBy := "admin"

	announcement, err := h.service.CreateMessagingAnnouncement(req, createdBy)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(announcement)
}

// HandleGetStats handles GET /admin/messaging/stats
func (h *ClientMessagingHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
