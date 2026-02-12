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

// ClientNote represents a CRM note for a client
type ClientNote struct {
	ID                 int64      `json:"id"`
	ClientID           int64      `json:"clientId"`
	Category           string     `json:"category"` // General, Sales Call, Support Issue, Risk Alert, KYC Follow-up, VIP Service, Compliance
	Priority           string     `json:"priority"` // low, normal, high, urgent
	Text               string     `json:"text"`
	Author             string     `json:"author"`
	FollowUpDate       *time.Time `json:"followUpDate,omitempty"`
	FollowUpCompleted  bool       `json:"followUpCompleted"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// Client represents basic client information for notes
type Client struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	AccountType  string    `json:"accountType"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registeredAt"`
	LastNoteAt   *time.Time `json:"lastNoteAt,omitempty"`
}

// NoteStore manages client notes and client data
type NoteStore struct {
	notes       map[int64]*ClientNote
	notesByClient map[int64][]*ClientNote
	clients     map[int64]*Client
	mu          sync.RWMutex
	nextNoteID  int64
}

// ClientNoteService provides business logic for client note management
type ClientNoteService struct {
	store *NoteStore
}

// NewClientNoteService creates a new client note service with mock data
func NewClientNoteService() *ClientNoteService {
	store := &NoteStore{
		notes:         make(map[int64]*ClientNote),
		notesByClient: make(map[int64][]*ClientNote),
		clients:       make(map[int64]*Client),
		nextNoteID:    1,
	}

	now := time.Now()

	// Create 50 mock clients
	clientNames := []struct {
		name  string
		email string
	}{
		{"John Smith", "john.smith@email.com"},
		{"Sarah Johnson", "sarah.j@email.com"},
		{"Michael Chen", "m.chen@email.com"},
		{"Emma Wilson", "emma.wilson@email.com"},
		{"David Martinez", "d.martinez@email.com"},
		{"Lisa Anderson", "lisa.a@email.com"},
		{"James Taylor", "james.t@email.com"},
		{"Maria Garcia", "maria.g@email.com"},
		{"Robert Brown", "robert.b@email.com"},
		{"Jennifer Lee", "jennifer.lee@email.com"},
		{"William Davis", "w.davis@email.com"},
		{"Elizabeth Moore", "e.moore@email.com"},
		{"Christopher White", "c.white@email.com"},
		{"Jessica Harris", "j.harris@email.com"},
		{"Daniel Thompson", "d.thompson@email.com"},
		{"Amanda Clark", "a.clark@email.com"},
		{"Matthew Lewis", "m.lewis@email.com"},
		{"Ashley Walker", "a.walker@email.com"},
		{"Andrew Hall", "a.hall@email.com"},
		{"Melissa Young", "m.young@email.com"},
		{"Joshua King", "j.king@email.com"},
		{"Stephanie Wright", "s.wright@email.com"},
		{"Ryan Lopez", "r.lopez@email.com"},
		{"Nicole Hill", "n.hill@email.com"},
		{"Brandon Scott", "b.scott@email.com"},
		{"Rachel Green", "r.green@email.com"},
		{"Tyler Adams", "t.adams@email.com"},
		{"Lauren Baker", "l.baker@email.com"},
		{"Kevin Nelson", "k.nelson@email.com"},
		{"Samantha Carter", "s.carter@email.com"},
		{"Justin Mitchell", "j.mitchell@email.com"},
		{"Rebecca Perez", "r.perez@email.com"},
		{"Jason Roberts", "j.roberts@email.com"},
		{"Michelle Turner", "m.turner@email.com"},
		{"Eric Phillips", "e.phillips@email.com"},
		{"Kimberly Campbell", "k.campbell@email.com"},
		{"Brian Parker", "b.parker@email.com"},
		{"Heather Evans", "h.evans@email.com"},
		{"Jeffrey Edwards", "j.edwards@email.com"},
		{"Laura Collins", "l.collins@email.com"},
		{"Mark Stewart", "m.stewart@email.com"},
		{"Karen Sanchez", "k.sanchez@email.com"},
		{"Steven Morris", "s.morris@email.com"},
		{"Lisa Rodriguez", "l.rodriguez@email.com"},
		{"Paul Reed", "p.reed@email.com"},
		{"Nancy Cook", "n.cook@email.com"},
		{"Richard Bailey", "r.bailey@email.com"},
		{"Donna Rivera", "d.rivera@email.com"},
		{"Joseph Cooper", "j.cooper@email.com"},
		{"Carol Richardson", "c.richardson@email.com"},
	}

	accountTypes := []string{"Standard", "ECN", "VIP", "Islamic", "Micro"}
	statuses := []string{"active", "pending_kyc", "suspended", "vip"}

	store.mu.Lock()
	for i, clientData := range clientNames {
		clientID := int64(1000 + i)
		client := &Client{
			ID:           clientID,
			Name:         clientData.name,
			Email:        clientData.email,
			AccountType:  accountTypes[i%len(accountTypes)],
			Status:       statuses[i%len(statuses)],
			RegisteredAt: now.Add(time.Duration(-i*3-30) * 24 * time.Hour),
		}
		store.clients[clientID] = client
		store.notesByClient[clientID] = []*ClientNote{}
	}

	// Generate 220+ notes across all categories
	categories := []string{"General", "Sales Call", "Support Issue", "Risk Alert", "KYC Follow-up", "VIP Service", "Compliance"}
	priorities := []string{"low", "normal", "high", "urgent"}
	authors := []string{"Admin", "Sales Team", "Support", "Risk Manager", "Compliance Officer", "VIP Manager"}

	noteTemplates := map[string][]string{
		"General": {
			"Client requested information about trading conditions",
			"Follow-up call scheduled",
			"Client expressed interest in higher leverage",
			"Discussed account upgrade options",
			"General inquiry about platform features",
		},
		"Sales Call": {
			"Initial contact - client interested in forex trading",
			"Follow-up on demo account performance",
			"Discussed deposit bonus opportunities",
			"Client ready to fund live account",
			"Upsell to VIP account successful",
		},
		"Support Issue": {
			"Client reported login issues - resolved",
			"Question about withdrawal processing time",
			"Platform technical issue - escalated to dev team",
			"Trade execution query - explained to client",
			"Account verification documents received",
		},
		"Risk Alert": {
			"High leverage usage detected - monitoring required",
			"Unusual trading pattern observed",
			"Account approaching margin call level",
			"Large withdrawal request - verification needed",
			"Multiple failed login attempts detected",
		},
		"KYC Follow-up": {
			"Proof of address required - email sent",
			"ID verification pending review",
			"Enhanced due diligence required",
			"Client submitted all KYC documents",
			"Waiting for source of funds documentation",
		},
		"VIP Service": {
			"Assigned dedicated account manager",
			"Custom trading conditions approved",
			"VIP event invitation sent",
			"Monthly performance review scheduled",
			"Personalized trading strategy consultation",
		},
		"Compliance": {
			"AML screening completed - no issues",
			"PEP check required",
			"Transaction monitoring alert - false positive",
			"Regulatory reporting submitted",
			"Compliance training completion verified",
		},
	}

	noteID := int64(1)
	notesPerClient := []int{7, 5, 3, 8, 4, 6, 2, 5, 9, 3, 4, 7, 5, 6, 3, 8, 2, 4, 5, 7, 3, 6, 4, 5, 8, 3, 7, 4, 6, 5, 9, 3, 5, 7, 4, 6, 8, 3, 5, 4, 7, 6, 3, 8, 5, 4, 6, 7, 3, 5}

	for clientIdx := 0; clientIdx < 50; clientIdx++ {
		clientID := int64(1000 + clientIdx)
		numNotes := notesPerClient[clientIdx]

		for noteIdx := 0; noteIdx < numNotes; noteIdx++ {
			category := categories[(clientIdx*7+noteIdx)%len(categories)]
			priority := priorities[(clientIdx+noteIdx)%len(priorities)]

			templates := noteTemplates[category]
			text := templates[noteIdx%len(templates)]

			createdAt := now.Add(time.Duration(-clientIdx*2-noteIdx*12) * time.Hour)

			var followUpDate *time.Time
			var followUpCompleted bool

			// 30% of notes have follow-ups
			if noteIdx%10 < 3 {
				followUpDays := 1 + (noteIdx % 7)
				fd := createdAt.Add(time.Duration(followUpDays) * 24 * time.Hour)
				followUpDate = &fd

				// 60% of follow-ups are completed
				if noteIdx%5 < 3 {
					followUpCompleted = true
				}
			}

			note := &ClientNote{
				ID:                noteID,
				ClientID:          clientID,
				Category:          category,
				Priority:          priority,
				Text:              text,
				Author:            authors[(clientIdx+noteIdx)%len(authors)],
				FollowUpDate:      followUpDate,
				FollowUpCompleted: followUpCompleted,
				CreatedAt:         createdAt,
				UpdatedAt:         createdAt,
			}

			store.notes[noteID] = note
			store.notesByClient[clientID] = append(store.notesByClient[clientID], note)
			noteID++

			// Update client's lastNoteAt
			if store.clients[clientID].LastNoteAt == nil || createdAt.After(*store.clients[clientID].LastNoteAt) {
				store.clients[clientID].LastNoteAt = &createdAt
			}
		}
	}

	store.nextNoteID = noteID
	store.mu.Unlock()

	return &ClientNoteService{store: store}
}

// GetClientNotes returns all notes for a specific client
func (cns *ClientNoteService) GetClientNotes(clientID int64, categoryFilter string) ([]*ClientNote, error) {
	cns.store.mu.RLock()
	defer cns.store.mu.RUnlock()

	if _, exists := cns.store.clients[clientID]; !exists {
		return nil, fmt.Errorf("client not found")
	}

	notes := cns.store.notesByClient[clientID]

	if categoryFilter != "" {
		filtered := make([]*ClientNote, 0)
		for _, note := range notes {
			if note.Category == categoryFilter {
				filtered = append(filtered, note)
			}
		}
		return filtered, nil
	}

	return notes, nil
}

// CreateNote creates a new note for a client
func (cns *ClientNoteService) CreateNote(note *ClientNote) (*ClientNote, error) {
	cns.store.mu.Lock()
	defer cns.store.mu.Unlock()

	if _, exists := cns.store.clients[note.ClientID]; !exists {
		return nil, fmt.Errorf("client not found")
	}

	if note.Text == "" {
		return nil, fmt.Errorf("note text is required")
	}
	if note.Category == "" {
		note.Category = "General"
	}
	if note.Priority == "" {
		note.Priority = "normal"
	}

	note.ID = cns.store.nextNoteID
	cns.store.nextNoteID++
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	note.FollowUpCompleted = false

	cns.store.notes[note.ID] = note
	cns.store.notesByClient[note.ClientID] = append(cns.store.notesByClient[note.ClientID], note)

	// Update client's lastNoteAt
	now := time.Now()
	cns.store.clients[note.ClientID].LastNoteAt = &now

	return note, nil
}

// UpdateNote updates an existing note
func (cns *ClientNoteService) UpdateNote(clientID, noteID int64, updates *ClientNote) (*ClientNote, error) {
	cns.store.mu.Lock()
	defer cns.store.mu.Unlock()

	note, exists := cns.store.notes[noteID]
	if !exists {
		return nil, fmt.Errorf("note not found")
	}

	if note.ClientID != clientID {
		return nil, fmt.Errorf("note does not belong to this client")
	}

	if updates.Text != "" {
		note.Text = updates.Text
	}
	if updates.Category != "" {
		note.Category = updates.Category
	}
	if updates.Priority != "" {
		note.Priority = updates.Priority
	}
	if updates.Author != "" {
		note.Author = updates.Author
	}
	if updates.FollowUpDate != nil {
		note.FollowUpDate = updates.FollowUpDate
	}

	note.UpdatedAt = time.Now()

	return note, nil
}

// DeleteNote deletes a note
func (cns *ClientNoteService) DeleteNote(clientID, noteID int64) error {
	cns.store.mu.Lock()
	defer cns.store.mu.Unlock()

	note, exists := cns.store.notes[noteID]
	if !exists {
		return fmt.Errorf("note not found")
	}

	if note.ClientID != clientID {
		return fmt.Errorf("note does not belong to this client")
	}

	delete(cns.store.notes, noteID)

	// Remove from client's notes list
	clientNotes := cns.store.notesByClient[clientID]
	for i, n := range clientNotes {
		if n.ID == noteID {
			cns.store.notesByClient[clientID] = append(clientNotes[:i], clientNotes[i+1:]...)
			break
		}
	}

	return nil
}

// GetStats returns statistics about notes
func (cns *ClientNoteService) GetStats() map[string]interface{} {
	cns.store.mu.RLock()
	defer cns.store.mu.RUnlock()

	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

	notesToday := 0
	pendingFollowUps := 0
	overdueFollowUps := 0

	for _, note := range cns.store.notes {
		if note.CreatedAt.After(today) {
			notesToday++
		}

		if note.FollowUpDate != nil && !note.FollowUpCompleted {
			pendingFollowUps++
			if note.FollowUpDate.Before(now) {
				overdueFollowUps++
			}
		}
	}

	clientsWithoutNotes := 0
	for _, client := range cns.store.clients {
		if client.LastNoteAt == nil || client.LastNoteAt.Before(thirtyDaysAgo) {
			clientsWithoutNotes++
		}
	}

	return map[string]interface{}{
		"totalNotes":            len(cns.store.notes),
		"notesToday":            notesToday,
		"pendingFollowUps":      pendingFollowUps,
		"overdueFollowUps":      overdueFollowUps,
		"clientsWithoutNotes30": clientsWithoutNotes,
		"totalClients":          len(cns.store.clients),
	}
}

// GetFollowUps returns all upcoming follow-ups
func (cns *ClientNoteService) GetFollowUps() []map[string]interface{} {
	cns.store.mu.RLock()
	defer cns.store.mu.RUnlock()

	now := time.Now()
	followUps := make([]map[string]interface{}, 0)

	for _, note := range cns.store.notes {
		if note.FollowUpDate != nil && !note.FollowUpCompleted {
			client := cns.store.clients[note.ClientID]

			status := "upcoming"
			if note.FollowUpDate.Before(now) {
				status = "overdue"
			} else if note.FollowUpDate.Before(now.Add(24 * time.Hour)) {
				status = "today"
			}

			followUps = append(followUps, map[string]interface{}{
				"noteId":       note.ID,
				"clientId":     note.ClientID,
				"clientName":   client.Name,
				"clientEmail":  client.Email,
				"category":     note.Category,
				"priority":     note.Priority,
				"text":         note.Text,
				"followUpDate": note.FollowUpDate,
				"status":       status,
				"createdAt":    note.CreatedAt,
			})
		}
	}

	return followUps
}

// CompleteFollowUp marks a follow-up as completed
func (cns *ClientNoteService) CompleteFollowUp(noteID int64) (*ClientNote, error) {
	cns.store.mu.Lock()
	defer cns.store.mu.Unlock()

	note, exists := cns.store.notes[noteID]
	if !exists {
		return nil, fmt.Errorf("note not found")
	}

	if note.FollowUpDate == nil {
		return nil, fmt.Errorf("note has no follow-up")
	}

	note.FollowUpCompleted = true
	note.UpdatedAt = time.Now()

	return note, nil
}

// SearchClients searches for clients by name, email, or ID
func (cns *ClientNoteService) SearchClients(query string) []*Client {
	cns.store.mu.RLock()
	defer cns.store.mu.RUnlock()

	query = strings.ToLower(query)
	results := make([]*Client, 0)

	for _, client := range cns.store.clients {
		if strings.Contains(strings.ToLower(client.Name), query) ||
			strings.Contains(strings.ToLower(client.Email), query) ||
			strings.Contains(fmt.Sprintf("%d", client.ID), query) {
			results = append(results, client)
		}
	}

	return results
}

// GetClient returns a specific client
func (cns *ClientNoteService) GetClient(clientID int64) (*Client, error) {
	cns.store.mu.RLock()
	defer cns.store.mu.RUnlock()

	client, exists := cns.store.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	return client, nil
}

// ClientNoteHandler handles HTTP requests for client note management
type ClientNoteHandler struct {
	service     *ClientNoteService
	authService *auth.Service
}

// NewClientNoteHandler creates a new client note HTTP handler
func NewClientNoteHandler(service *ClientNoteService, authService *auth.Service) *ClientNoteHandler {
	return &ClientNoteHandler{
		service:     service,
		authService: authService,
	}
}

// GetClientNotes handles GET /admin/clients/:clientId/notes
func (cnh *ClientNoteHandler) GetClientNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	clientIDStr := pathParts[3]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	categoryFilter := r.URL.Query().Get("category")
	notes, err := cnh.service.GetClientNotes(clientID, categoryFilter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	client, _ := cnh.service.GetClient(clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"client":  client,
		"data":    notes,
		"count":   len(notes),
	})
}

// CreateNote handles POST /admin/clients/:clientId/notes
func (cnh *ClientNoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	clientIDStr := pathParts[3]
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var note ClientNote
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	note.ClientID = clientID
	created, err := cnh.service.CreateNote(&note)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    created,
		"message": "Note created successfully",
	})
}

// UpdateNote handles PUT /admin/clients/:clientId/notes/:noteId
func (cnh *ClientNoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	clientID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	noteID, err := strconv.ParseInt(pathParts[5], 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	var updates ClientNote
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := cnh.service.UpdateNote(clientID, noteID, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    updated,
		"message": "Note updated successfully",
	})
}

// DeleteNote handles DELETE /admin/clients/:clientId/notes/:noteId
func (cnh *ClientNoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	clientID, err := strconv.ParseInt(pathParts[3], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	noteID, err := strconv.ParseInt(pathParts[5], 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	if err := cnh.service.DeleteNote(clientID, noteID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Note deleted successfully",
	})
}

// GetStats handles GET /admin/clients/notes/stats
func (cnh *ClientNoteHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := cnh.service.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetFollowUps handles GET /admin/clients/notes/follow-ups
func (cnh *ClientNoteHandler) GetFollowUps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	followUps := cnh.service.GetFollowUps()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    followUps,
		"count":   len(followUps),
	})
}

// CompleteFollowUp handles POST /admin/clients/notes/:noteId/complete-followup
func (cnh *ClientNoteHandler) CompleteFollowUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	noteID, err := strconv.ParseInt(pathParts[4], 10, 64)
	if err != nil {
		http.Error(w, "Invalid note ID", http.StatusBadRequest)
		return
	}

	updated, err := cnh.service.CompleteFollowUp(noteID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    updated,
		"message": "Follow-up marked as completed",
	})
}

// SearchClients handles GET /admin/clients/search?q=term
func (cnh *ClientNoteHandler) SearchClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if _, err := cnh.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query required", http.StatusBadRequest)
		return
	}

	results := cnh.service.SearchClients(query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    results,
		"count":   len(results),
		"query":   query,
	})
}
