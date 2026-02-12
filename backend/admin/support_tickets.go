package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TicketStatus represents the state of a support ticket
type TicketStatus string

const (
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusPending  TicketStatus = "pending"
	TicketStatusResolved TicketStatus = "resolved"
	TicketStatusClosed   TicketStatus = "closed"
)

// TicketPriority represents the urgency level of a ticket
type TicketPriority string

const (
	PriorityLow    TicketPriority = "low"
	PriorityMedium TicketPriority = "medium"
	PriorityHigh   TicketPriority = "high"
	PriorityUrgent TicketPriority = "urgent"
)

// TicketCategory represents the type of support issue
type TicketCategory string

const (
	CategoryGeneral     TicketCategory = "general"
	CategoryTechnical   TicketCategory = "technical"
	CategoryAccount     TicketCategory = "account"
	CategoryDeposit     TicketCategory = "deposit"
	CategoryWithdrawal  TicketCategory = "withdrawal"
	CategoryTrading     TicketCategory = "trading"
	CategoryCompliance  TicketCategory = "compliance"
	CategoryBilling     TicketCategory = "billing"
)

// TicketMessage represents a message in a support ticket thread
type TicketMessage struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	SenderID  string    `json:"senderId"`
	SenderName string   `json:"senderName"`
	SenderType string   `json:"senderType"` // "client" or "support"
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// SupportTicket represents a customer support ticket
type SupportTicket struct {
	ID          string           `json:"id"`
	ClientID    string           `json:"clientId"`
	ClientName  string           `json:"clientName"`
	ClientEmail string           `json:"clientEmail"`
	Subject     string           `json:"subject"`
	Status      TicketStatus     `json:"status"`
	Priority    TicketPriority   `json:"priority"`
	Category    TicketCategory   `json:"category"`
	Messages    []TicketMessage  `json:"messages"`
	AssignedTo  string           `json:"assignedTo,omitempty"`
	AssignedToName string        `json:"assignedToName,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	ResolvedAt  *time.Time       `json:"resolvedAt,omitempty"`
	SLADeadline time.Time        `json:"slaDeadline"`
}

// TicketStats provides aggregate statistics for support tickets
type TicketStats struct {
	Total         int            `json:"total"`
	ByStatus      map[string]int `json:"byStatus"`
	ByPriority    map[string]int `json:"byPriority"`
	AvgResponseTime string       `json:"avgResponseTime"`
	SLACompliance   float64      `json:"slaCompliance"`
	UnassignedCount int          `json:"unassignedCount"`
}

// TicketService manages support ticket operations
type TicketService struct {
	mu           sync.RWMutex
	tickets      map[string]*SupportTicket
	nextTicketID int
}

// NewTicketService creates a new ticket service with mock data
func NewTicketService() *TicketService {
	svc := &TicketService{
		tickets:      make(map[string]*SupportTicket),
		nextTicketID: 1,
	}

	// Generate 25+ mock tickets
	svc.generateMockTickets()

	return svc
}

// generateMockTickets creates realistic test data
func (s *TicketService) generateMockTickets() {
	clientNames := []string{
		"John Doe", "Jane Smith", "Robert Johnson", "Maria Garcia", "David Lee",
		"Sarah Williams", "Michael Brown", "Emily Davis", "James Wilson", "Lisa Martinez",
		"Christopher Anderson", "Jennifer Taylor", "Daniel Thomas", "Patricia Moore", "Matthew Jackson",
		"Nancy White", "Charles Harris", "Betty Thompson", "Joseph Robinson", "Sandra Clark",
		"Kenneth Lewis", "Margaret Walker", "Steven Hall", "Ashley Allen", "Brian Young",
	}

	subjects := []string{
		"Unable to login to trading account",
		"Deposit not showing in my account",
		"Cannot place orders on EURUSD",
		"Withdrawal request pending for 3 days",
		"Need to update my KYC documents",
		"Account verification status inquiry",
		"Trading platform freezes frequently",
		"Missing transactions in account history",
		"Request to increase leverage limit",
		"Need help with stop loss orders",
		"Commission charges seem incorrect",
		"Cannot access mobile trading app",
		"Swap rates calculation question",
		"Account suspended without notification",
		"Request to close duplicate account",
		"Need statement for tax purposes",
		"API keys not working properly",
		"Chart data loading very slowly",
		"Margin call notification not received",
		"Unable to withdraw to new bank account",
		"Question about affiliate program",
		"Need to change registered email",
		"Trading signals not working",
		"Request refund for failed deposit",
		"Account hacked - urgent security issue",
	}

	categories := []TicketCategory{
		CategoryTechnical, CategoryDeposit, CategoryTrading, CategoryWithdrawal, CategoryCompliance,
		CategoryAccount, CategoryTechnical, CategoryAccount, CategoryTrading, CategoryTrading,
		CategoryBilling, CategoryTechnical, CategoryTrading, CategoryAccount, CategoryAccount,
		CategoryGeneral, CategoryTechnical, CategoryTechnical, CategoryTrading, CategoryWithdrawal,
		CategoryGeneral, CategoryAccount, CategoryGeneral, CategoryDeposit, CategoryCompliance,
	}

	priorities := []TicketPriority{
		PriorityMedium, PriorityHigh, PriorityMedium, PriorityHigh, PriorityLow,
		PriorityMedium, PriorityHigh, PriorityMedium, PriorityLow, PriorityMedium,
		PriorityLow, PriorityHigh, PriorityLow, PriorityUrgent, PriorityMedium,
		PriorityLow, PriorityMedium, PriorityHigh, PriorityMedium, PriorityHigh,
		PriorityLow, PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent,
	}

	statuses := []TicketStatus{
		TicketStatusOpen, TicketStatusPending, TicketStatusOpen, TicketStatusPending, TicketStatusResolved,
		TicketStatusOpen, TicketStatusPending, TicketStatusResolved, TicketStatusClosed, TicketStatusOpen,
		TicketStatusResolved, TicketStatusPending, TicketStatusClosed, TicketStatusOpen, TicketStatusPending,
		TicketStatusClosed, TicketStatusOpen, TicketStatusPending, TicketStatusResolved, TicketStatusPending,
		TicketStatusClosed, TicketStatusResolved, TicketStatusOpen, TicketStatusPending, TicketStatusOpen,
	}

	agents := []string{"", "Sarah Chen", "", "Mike Johnson", "Emily Rodriguez", "", "John Parker", "Lisa Anderson", "", "David Kim"}

	now := time.Now()

	for i := 0; i < 25; i++ {
		ticketID := fmt.Sprintf("TKT-%05d", s.nextTicketID)
		s.nextTicketID++

		createdAt := now.Add(time.Duration(-(24 - i)) * time.Hour)
		updatedAt := createdAt.Add(time.Duration(i*15) * time.Minute)

		// Calculate SLA deadline based on priority
		var slaHours int
		switch priorities[i] {
		case PriorityUrgent:
			slaHours = 4
		case PriorityHigh:
			slaHours = 12
		case PriorityMedium:
			slaHours = 24
		case PriorityLow:
			slaHours = 48
		}
		slaDeadline := createdAt.Add(time.Duration(slaHours) * time.Hour)

		clientID := fmt.Sprintf("C-%06d", 100000+i)
		clientEmail := strings.ToLower(strings.ReplaceAll(clientNames[i], " ", ".")) + "@example.com"

		// Create initial message from client
		messages := []TicketMessage{
			{
				ID:         fmt.Sprintf("MSG-%d-001", i+1),
				TicketID:   ticketID,
				SenderID:   clientID,
				SenderName: clientNames[i],
				SenderType: "client",
				Message:    fmt.Sprintf("Hello, I need help with: %s. This issue started earlier today and I need urgent assistance.", subjects[i]),
				CreatedAt:  createdAt,
			},
		}

		// Add support responses for some tickets
		if i%3 == 0 && len(agents) > i%len(agents) && agents[i%len(agents)] != "" {
			messages = append(messages, TicketMessage{
				ID:         fmt.Sprintf("MSG-%d-002", i+1),
				TicketID:   ticketID,
				SenderID:   fmt.Sprintf("ADMIN-%d", i%len(agents)),
				SenderName: agents[i%len(agents)],
				SenderType: "support",
				Message:    "Thank you for contacting our support team. I'm looking into this issue for you now. Could you provide more details about when this started?",
				CreatedAt:  createdAt.Add(30 * time.Minute),
			})

			// Client follow-up
			messages = append(messages, TicketMessage{
				ID:         fmt.Sprintf("MSG-%d-003", i+1),
				TicketID:   ticketID,
				SenderID:   clientID,
				SenderName: clientNames[i],
				SenderType: "client",
				Message:    "Thanks for the quick response! This issue started around 2 hours ago. I've tried clearing my cache but it didn't help.",
				CreatedAt:  createdAt.Add(45 * time.Minute),
			})
		}

		// Set resolved time for closed tickets
		var resolvedAt *time.Time
		if statuses[i] == TicketStatusResolved || statuses[i] == TicketStatusClosed {
			resolved := updatedAt.Add(2 * time.Hour)
			resolvedAt = &resolved
		}

		assignedTo := ""
		assignedToName := ""
		if len(agents) > i%len(agents) && agents[i%len(agents)] != "" {
			assignedTo = fmt.Sprintf("ADMIN-%d", i%len(agents))
			assignedToName = agents[i%len(agents)]
		}

		ticket := &SupportTicket{
			ID:             ticketID,
			ClientID:       clientID,
			ClientName:     clientNames[i],
			ClientEmail:    clientEmail,
			Subject:        subjects[i],
			Status:         statuses[i],
			Priority:       priorities[i],
			Category:       categories[i],
			Messages:       messages,
			AssignedTo:     assignedTo,
			AssignedToName: assignedToName,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			ResolvedAt:     resolvedAt,
			SLADeadline:    slaDeadline,
		}

		s.tickets[ticketID] = ticket
	}

	log.Printf("[TicketService] Generated %d mock support tickets", len(s.tickets))
}

// ListTickets returns tickets with optional filters
func (s *TicketService) ListTickets(status, priority, category, assignedTo string) []*SupportTicket {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*SupportTicket

	for _, ticket := range s.tickets {
		// Apply filters
		if status != "" && string(ticket.Status) != status {
			continue
		}
		if priority != "" && string(ticket.Priority) != priority {
			continue
		}
		if category != "" && string(ticket.Category) != category {
			continue
		}
		if assignedTo != "" {
			if assignedTo == "unassigned" && ticket.AssignedTo != "" {
				continue
			} else if assignedTo != "unassigned" && ticket.AssignedTo != assignedTo {
				continue
			}
		}

		result = append(result, ticket)
	}

	return result
}

// GetTicket returns a ticket by ID
func (s *TicketService) GetTicket(ticketID string) (*SupportTicket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("ticket not found: %s", ticketID)
	}

	return ticket, nil
}

// AddReply adds a message to a ticket
func (s *TicketService) AddReply(ticketID, senderID, senderName, senderType, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found: %s", ticketID)
	}

	messageID := fmt.Sprintf("MSG-%s-%03d", ticketID, len(ticket.Messages)+1)

	newMessage := TicketMessage{
		ID:         messageID,
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderName: senderName,
		SenderType: senderType,
		Message:    message,
		CreatedAt:  time.Now(),
	}

	ticket.Messages = append(ticket.Messages, newMessage)
	ticket.UpdatedAt = time.Now()

	// If support replied, update status to pending if it was open
	if senderType == "support" && ticket.Status == TicketStatusOpen {
		ticket.Status = TicketStatusPending
	}

	log.Printf("[TicketService] Reply added to ticket %s by %s (%s)", ticketID, senderName, senderType)

	return nil
}

// UpdateStatus changes the status of a ticket
func (s *TicketService) UpdateStatus(ticketID string, newStatus TicketStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found: %s", ticketID)
	}

	oldStatus := ticket.Status
	ticket.Status = newStatus
	ticket.UpdatedAt = time.Now()

	// Set resolved time if status is resolved or closed
	if (newStatus == TicketStatusResolved || newStatus == TicketStatusClosed) && ticket.ResolvedAt == nil {
		now := time.Now()
		ticket.ResolvedAt = &now
	}

	log.Printf("[TicketService] Ticket %s status changed: %s -> %s", ticketID, oldStatus, newStatus)

	return nil
}

// AssignTicket assigns a ticket to an agent
func (s *TicketService) AssignTicket(ticketID, agentID, agentName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ticket, exists := s.tickets[ticketID]
	if !exists {
		return fmt.Errorf("ticket not found: %s", ticketID)
	}

	ticket.AssignedTo = agentID
	ticket.AssignedToName = agentName
	ticket.UpdatedAt = time.Now()

	log.Printf("[TicketService] Ticket %s assigned to %s (%s)", ticketID, agentName, agentID)

	return nil
}

// GetStats returns aggregate statistics
func (s *TicketService) GetStats() TicketStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := TicketStats{
		Total:      len(s.tickets),
		ByStatus:   make(map[string]int),
		ByPriority: make(map[string]int),
	}

	var totalResponseTime time.Duration
	var responseCount int
	slaMetCount := 0

	for _, ticket := range s.tickets {
		// Count by status
		stats.ByStatus[string(ticket.Status)]++

		// Count by priority
		stats.ByPriority[string(ticket.Priority)]++

		// Count unassigned
		if ticket.AssignedTo == "" {
			stats.UnassignedCount++
		}

		// Calculate response time for tickets with at least 2 messages
		if len(ticket.Messages) >= 2 {
			responseTime := ticket.Messages[1].CreatedAt.Sub(ticket.Messages[0].CreatedAt)
			totalResponseTime += responseTime
			responseCount++
		}

		// Check SLA compliance
		if ticket.ResolvedAt != nil && ticket.ResolvedAt.Before(ticket.SLADeadline) {
			slaMetCount++
		}
	}

	// Calculate average response time
	if responseCount > 0 {
		avgSeconds := int(totalResponseTime.Seconds()) / responseCount
		stats.AvgResponseTime = fmt.Sprintf("%d min", avgSeconds/60)
	} else {
		stats.AvgResponseTime = "N/A"
	}

	// Calculate SLA compliance percentage
	if len(s.tickets) > 0 {
		stats.SLACompliance = float64(slaMetCount) / float64(len(s.tickets)) * 100
	}

	return stats
}

// TicketHandler provides HTTP handlers for ticket endpoints
type TicketHandler struct {
	service *TicketService
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler() *TicketHandler {
	return &TicketHandler{
		service: NewTicketService(),
	}
}

// HandleListTickets handles GET /admin/tickets
func (h *TicketHandler) HandleListTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters for filters
	status := r.URL.Query().Get("status")
	priority := r.URL.Query().Get("priority")
	category := r.URL.Query().Get("category")
	assignedTo := r.URL.Query().Get("assignedTo")

	tickets := h.service.ListTickets(status, priority, category, assignedTo)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// HandleGetTicket handles GET /admin/tickets/:id
func (h *TicketHandler) HandleGetTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ticket ID from URL path
	ticketID := strings.TrimPrefix(r.URL.Path, "/admin/tickets/")

	ticket, err := h.service.GetTicket(ticketID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

// HandleAddReply handles POST /admin/tickets/:id/reply
func (h *TicketHandler) HandleAddReply(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ticket ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/tickets/")
	ticketID := strings.TrimSuffix(path, "/reply")

	var req struct {
		SenderID   string `json:"senderId"`
		SenderName string `json:"senderName"`
		SenderType string `json:"senderType"` // "client" or "support"
		Message    string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	err := h.service.AddReply(ticketID, req.SenderID, req.SenderName, req.SenderType, req.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleUpdateStatus handles PUT /admin/tickets/:id/status
func (h *TicketHandler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ticket ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/tickets/")
	ticketID := strings.TrimSuffix(path, "/status")

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newStatus := TicketStatus(req.Status)
	if newStatus != TicketStatusOpen && newStatus != TicketStatusPending &&
	   newStatus != TicketStatusResolved && newStatus != TicketStatusClosed {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	err := h.service.UpdateStatus(ticketID, newStatus)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleAssignTicket handles PUT /admin/tickets/:id/assign
func (h *TicketHandler) HandleAssignTicket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ticket ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/tickets/")
	ticketID := strings.TrimSuffix(path, "/assign")

	var req struct {
		AgentID   string `json:"agentId"`
		AgentName string `json:"agentName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.AssignTicket(ticketID, req.AgentID, req.AgentName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// HandleGetStats handles GET /admin/tickets/stats
func (h *TicketHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
