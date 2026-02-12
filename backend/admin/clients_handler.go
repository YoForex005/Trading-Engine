//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ClientInfo represents a trader/client account
type ClientInfo struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Email            string             `json:"email"`
	Balance          float64            `json:"balance"`
	Equity           float64            `json:"equity"`
	OpenPositions    int                `json:"openPositions"`
	KYCStatus        string             `json:"kycStatus"` // pending, verified, rejected
	RegistrationDate time.Time          `json:"registrationDate"`
	LastLogin        time.Time          `json:"lastLogin"`
	Status           string             `json:"status"` // active, suspended, closed
	TradingStats     TradingStats       `json:"tradingStats"`
	RecentTrades     []RecentTrade      `json:"recentTrades"`
	Deposits         []ClientTransaction      `json:"deposits"`
	Withdrawals      []ClientTransaction      `json:"withdrawals"`
}

// TradingStats holds trading statistics for a client
type TradingStats struct {
	TotalTrades    int     `json:"totalTrades"`
	WinRate        float64 `json:"winRate"`
	TotalProfit    float64 `json:"totalProfit"`
	TotalVolume    float64 `json:"totalVolume"`
	AvgTradeSize   float64 `json:"avgTradeSize"`
	LargestProfit  float64 `json:"largestProfit"`
	LargestLoss    float64 `json:"largestLoss"`
}

// RecentTrade represents a recent trading activity
type RecentTrade struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Type        string    `json:"type"` // buy, sell
	Volume      float64   `json:"volume"`
	OpenPrice   float64   `json:"openPrice"`
	ClosePrice  float64   `json:"closePrice"`
	Profit      float64   `json:"profit"`
	OpenTime    time.Time `json:"openTime"`
	CloseTime   time.Time `json:"closeTime"`
}

// ClientTransaction represents a deposit or withdrawal
type ClientTransaction struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // deposit, withdrawal
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // pending, completed, rejected
	Method    string    `json:"method"` // bank_transfer, credit_card, crypto
	CreatedAt time.Time `json:"createdAt"`
}

// ClientStore manages client data in memory
type ClientStore struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
}

// NewClientStore creates a new client store with mock data
func NewClientStore() *ClientStore {
	store := &ClientStore{
		clients: make(map[string]*ClientInfo),
	}

	// Generate 25 mock clients
	store.generateMockClients()

	return store
}

// generateMockClients creates 25 mock clients for testing
func (s *ClientStore) generateMockClients() {
	names := []string{
		"John Smith", "Emma Johnson", "Michael Brown", "Sophia Davis", "James Wilson",
		"Olivia Martinez", "William Anderson", "Ava Taylor", "Robert Thomas", "Isabella Garcia",
		"David Rodriguez", "Mia Hernandez", "Richard Moore", "Charlotte Martin", "Joseph Jackson",
		"Amelia Thompson", "Thomas White", "Harper Lopez", "Charles Lee", "Evelyn Walker",
		"Daniel Hall", "Abigail Allen", "Matthew Young", "Emily King", "Christopher Wright",
	}

	emails := []string{
		"john.smith@example.com", "emma.johnson@example.com", "michael.brown@example.com",
		"sophia.davis@example.com", "james.wilson@example.com", "olivia.martinez@example.com",
		"william.anderson@example.com", "ava.taylor@example.com", "robert.thomas@example.com",
		"isabella.garcia@example.com", "david.rodriguez@example.com", "mia.hernandez@example.com",
		"richard.moore@example.com", "charlotte.martin@example.com", "joseph.jackson@example.com",
		"amelia.thompson@example.com", "thomas.white@example.com", "harper.lopez@example.com",
		"charles.lee@example.com", "evelyn.walker@example.com", "daniel.hall@example.com",
		"abigail.allen@example.com", "matthew.young@example.com", "emily.king@example.com",
		"christopher.wright@example.com",
	}

	kycStatuses := []string{"verified", "verified", "verified", "pending", "verified"}
	statuses := []string{"active", "active", "active", "suspended", "active"}

	for i := 0; i < 25; i++ {
		balance := 5000.0 + float64(i*1000)
		equity := balance + float64((i%10-5)*100)
		openPositions := i % 5

		client := &ClientInfo{
			ID:               fmt.Sprintf("client-%d", i+1),
			Name:             names[i],
			Email:            emails[i],
			Balance:          balance,
			Equity:           equity,
			OpenPositions:    openPositions,
			KYCStatus:        kycStatuses[i%5],
			RegistrationDate: time.Now().AddDate(0, -i, 0),
			LastLogin:        time.Now().Add(-time.Duration(i) * time.Hour),
			Status:           statuses[i%5],
			TradingStats: TradingStats{
				TotalTrades:   50 + i*10,
				WinRate:       45.0 + float64(i%20),
				TotalProfit:   1000.0 + float64(i*500),
				TotalVolume:   100.0 + float64(i*50),
				AvgTradeSize:  1.5 + float64(i%10)*0.5,
				LargestProfit: 500.0 + float64(i*50),
				LargestLoss:   -300.0 - float64(i*20),
			},
			RecentTrades: s.generateMockTrades(3),
			Deposits:     s.generateMockClientTransactions("deposit", 2),
			Withdrawals:  s.generateMockClientTransactions("withdrawal", 1),
		}

		s.clients[client.ID] = client
	}

	log.Printf("[ClientStore] Generated 25 mock clients")
}

// generateMockTrades creates mock recent trades
func (s *ClientStore) generateMockTrades(count int) []RecentTrade {
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD"}
	types := []string{"buy", "sell"}

	trades := make([]RecentTrade, count)
	for i := 0; i < count; i++ {
		openPrice := 1.1000 + float64(i%10)*0.001
		closePrice := openPrice + (float64((i%2)*2-1) * 0.002)
		volume := 1.0 + float64(i%5)*0.5
		profit := (closePrice - openPrice) * volume * 100000

		trades[i] = RecentTrade{
			ID:         fmt.Sprintf("trade-%d", i+1),
			Symbol:     symbols[i%len(symbols)],
			Type:       types[i%len(types)],
			Volume:     volume,
			OpenPrice:  openPrice,
			ClosePrice: closePrice,
			Profit:     profit,
			OpenTime:   time.Now().Add(-time.Duration(i*2) * time.Hour),
			CloseTime:  time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	return trades
}

// generateMockClientTransactions creates mock deposit/withdrawal transactions
func (s *ClientStore) generateMockClientTransactions(txType string, count int) []ClientTransaction {
	methods := []string{"bank_transfer", "credit_card", "crypto"}
	statuses := []string{"completed", "pending"}

	transactions := make([]ClientTransaction, count)
	for i := 0; i < count; i++ {
		amount := 1000.0 + float64(i*500)
		if txType == "withdrawal" {
			amount = 500.0 + float64(i*200)
		}

		transactions[i] = ClientTransaction{
			ID:        fmt.Sprintf("%s-%d", txType, i+1),
			Type:      txType,
			Amount:    amount,
			Status:    statuses[i%len(statuses)],
			Method:    methods[i%len(methods)],
			CreatedAt: time.Now().Add(-time.Duration(i*24) * time.Hour),
		}
	}

	return transactions
}

// GetClients returns all clients with optional filtering
func (s *ClientStore) GetClients(statusFilter, kycFilter string) []*ClientInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*ClientInfo, 0)

	for _, client := range s.clients {
		// Apply filters
		if statusFilter != "" && client.Status != statusFilter {
			continue
		}

		if kycFilter != "" && client.KYCStatus != kycFilter {
			continue
		}

		clients = append(clients, client)
	}

	return clients
}

// GetClient returns a single client by ID
func (s *ClientStore) GetClient(id string) (*ClientInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[id]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	return client, nil
}

// SuspendClient suspends a client account
func (s *ClientStore) SuspendClient(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[id]
	if !exists {
		return fmt.Errorf("client not found")
	}

	client.Status = "suspended"

	log.Printf("[ClientStore] Suspended client %s (%s)", id, client.Name)

	return nil
}

// ActivateClient activates a suspended client account
func (s *ClientStore) ActivateClient(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[id]
	if !exists {
		return fmt.Errorf("client not found")
	}

	client.Status = "active"

	log.Printf("[ClientStore] Activated client %s (%s)", id, client.Name)

	return nil
}

// ClientsHandler handles client management API endpoints
type ClientsHandler struct {
	store       *ClientStore
	authService *auth.Service
}

// NewClientsHandler creates a new clients handler
func NewClientsHandler(store *ClientStore, authService *auth.Service) *ClientsHandler {
	return &ClientsHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListClients returns a list of clients with optional filters
// GET /admin/clients?status=active&kyc=verified
func (h *ClientsHandler) HandleListClients(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
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

	// Parse query parameters for filtering
	statusFilter := r.URL.Query().Get("status")
	kycFilter := r.URL.Query().Get("kyc")

	// Get clients with filters
	clients := h.store.GetClients(statusFilter, kycFilter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	})
}

// HandleGetClient returns detailed information about a specific client
// GET /admin/clients/:id
func (h *ClientsHandler) HandleGetClient(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
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

	// Extract client ID from URL path
	// Expected: /admin/clients/{id}
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := pathParts[3]

	// Get client
	client, err := h.store.GetClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

// HandleSuspendClient suspends a client account
// PUT /admin/clients/:id/suspend
func (h *ClientsHandler) HandleSuspendClient(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
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

	// Extract client ID from URL path
	// Expected: /admin/clients/{id}/suspend
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := pathParts[3]

	// Suspend client
	if err := h.store.SuspendClient(clientID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get updated client
	client, _ := h.store.GetClient(clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Client suspended successfully",
		"client":  client,
	})
}

// HandleActivateClient activates a suspended client account
// PUT /admin/clients/:id/activate
func (h *ClientsHandler) HandleActivateClient(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
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

	// Extract client ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := pathParts[3]

	// Activate client
	if err := h.store.ActivateClient(clientID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get updated client
	client, _ := h.store.GetClient(clientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Client activated successfully",
		"client":  client,
	})
}

// HandleSendNotification sends a notification to a client
// POST /admin/clients/:id/notification
func (h *ClientsHandler) HandleSendNotification(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
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

	// Extract client ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := pathParts[3]

	// Parse request body
	var req struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.Subject == "" || req.Message == "" {
		http.Error(w, "Subject and message are required", http.StatusBadRequest)
		return
	}

	// Verify client exists
	client, err := h.store.GetClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// In a real implementation, this would send an actual notification
	// For now, just log it
	log.Printf("[Notifications] Sending notification to client %s (%s): %s - %s",
		clientID, client.Name, req.Subject, req.Message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Notification sent successfully",
		"clientId": clientID,
		"subject":  req.Subject,
	})
}

// validateAdminAuth validates JWT token and checks admin role
func (h *ClientsHandler) validateAdminAuth(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := auth.ValidateTokenWithDefault(tokenString)
	if err != nil {
		return nil, err
	}

	// For now, we'll assume all authenticated users can manage clients
	// In production, you'd check claims.Role == "admin"

	return claims, nil
}
