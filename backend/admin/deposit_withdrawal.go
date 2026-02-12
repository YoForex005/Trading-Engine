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

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

// TransactionType represents deposit or withdrawal
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
)

// TransactionStatus represents the current state of a transaction
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusApproved   TransactionStatus = "approved"
	TransactionStatusRejected   TransactionStatus = "rejected"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
)

// PaymentMethod represents the method used for the transaction
type PaymentMethod string

const (
	PaymentMethodBankWire    PaymentMethod = "bank_wire"
	PaymentMethodCreditCard  PaymentMethod = "credit_card"
	PaymentMethodCrypto      PaymentMethod = "crypto"
	PaymentMethodSkrill      PaymentMethod = "skrill"
	PaymentMethodNeteller    PaymentMethod = "neteller"
	PaymentMethodPayPal      PaymentMethod = "paypal"
)

// Transaction represents a deposit or withdrawal transaction
type Transaction struct {
	ID            string            `json:"id"`
	ClientID      int64             `json:"clientId"`
	ClientName    string            `json:"clientName"`
	Type          TransactionType   `json:"type"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Method        PaymentMethod     `json:"method"`
	Status        TransactionStatus `json:"status"`
	RequestedAt   time.Time         `json:"requestedAt"`
	ProcessedAt   *time.Time        `json:"processedAt,omitempty"`
	ProcessedBy   string            `json:"processedBy,omitempty"`
	Notes         string            `json:"notes,omitempty"`
	BankRef       string            `json:"bankRef,omitempty"`
	WalletAddress string            `json:"walletAddress,omitempty"`
}

// TransactionStats represents statistics about transactions
type TransactionStats struct {
	PendingCount       int     `json:"pendingCount"`
	TodayDepositsTotal float64 `json:"todayDepositsTotal"`
	TodayWithdrawalsTotal float64 `json:"todayWithdrawalsTotal"`
	ProcessingCount    int     `json:"processingCount"`
	TodayDepositsCount int     `json:"todayDepositsCount"`
	TodayWithdrawalsCount int  `json:"todayWithdrawalsCount"`
}

// DailyTransactionSummary represents daily deposit/withdrawal totals
type DailyTransactionSummary struct {
	Date              string  `json:"date"`
	DepositsTotal     float64 `json:"depositsTotal"`
	WithdrawalsTotal  float64 `json:"withdrawalsTotal"`
	DepositsCount     int     `json:"depositsCount"`
	WithdrawalsCount  int     `json:"withdrawalsCount"`
	NetFlow           float64 `json:"netFlow"`
}

// ============================================
// In-Memory Store
// ============================================

// TransactionStore holds all transaction data in memory with thread-safe access
type TransactionStore struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
	nextID       int
}

// NewTransactionStore creates a new transaction store with mock data
func NewTransactionStore() *TransactionStore {
	store := &TransactionStore{
		transactions: make(map[string]*Transaction),
		nextID:       1,
	}
	store.generateMockData()
	return store
}

// generateMockData creates 150 mock transactions (80 deposits, 70 withdrawals)
func (s *TransactionStore) generateMockData() {
	methods := []PaymentMethod{
		PaymentMethodBankWire,
		PaymentMethodCreditCard,
		PaymentMethodCrypto,
		PaymentMethodSkrill,
		PaymentMethodNeteller,
		PaymentMethodPayPal,
	}

	statuses := []TransactionStatus{
		TransactionStatusPending,
		TransactionStatusApproved,
		TransactionStatusRejected,
		TransactionStatusProcessing,
		TransactionStatusCompleted,
		TransactionStatusFailed,
	}

	currencies := []string{"USD", "EUR", "GBP", "AUD", "CAD"}
	clientNames := []string{
		"John Smith", "Emma Wilson", "Michael Chen", "Sarah Johnson", "David Lee",
		"Maria Garcia", "James Brown", "Lisa Anderson", "Robert Taylor", "Jennifer Martinez",
		"William Davis", "Jessica Miller", "Richard White", "Karen Moore", "Thomas Jackson",
		"Nancy Harris", "Charles Martin", "Betty Thompson", "Christopher Garcia", "Dorothy Robinson",
	}

	pendingCount := 0
	now := time.Now()

	// Generate 80 deposits
	for i := 0; i < 80; i++ {
		txID := fmt.Sprintf("TXN%06d", s.nextID)
		s.nextID++

		method := methods[rand.Intn(len(methods))]
		currency := currencies[rand.Intn(len(currencies))]
		clientName := clientNames[rand.Intn(len(clientNames))]
		clientID := int64(1000 + rand.Intn(9000))

		var status TransactionStatus
		if pendingCount < 15 && rand.Float64() < 0.2 {
			status = TransactionStatusPending
			pendingCount++
		} else {
			status = statuses[rand.Intn(len(statuses))]
		}

		amount := 100 + rand.Float64()*49900 // $100 - $50,000
		requestedAt := now.Add(-time.Duration(rand.Intn(90*24)) * time.Hour)

		tx := &Transaction{
			ID:          txID,
			ClientID:    clientID,
			ClientName:  clientName,
			Type:        TransactionTypeDeposit,
			Amount:      math.Round(amount*100) / 100,
			Currency:    currency,
			Method:      method,
			Status:      status,
			RequestedAt: requestedAt,
		}

		// Add processed info for non-pending transactions
		if status != TransactionStatusPending {
			processedAt := requestedAt.Add(time.Duration(1+rand.Intn(48)) * time.Hour)
			tx.ProcessedAt = &processedAt
			tx.ProcessedBy = "admin@rtx5.com"
		}

		// Add reference numbers based on method
		if method == PaymentMethodBankWire {
			tx.BankRef = fmt.Sprintf("WIRE%d", 100000+rand.Intn(900000))
		} else if method == PaymentMethodCrypto {
			tx.WalletAddress = fmt.Sprintf("0x%x", rand.Uint64())
		}

		if status == TransactionStatusRejected {
			tx.Notes = "Insufficient verification"
		}

		s.transactions[txID] = tx
	}

	// Generate 70 withdrawals
	for i := 0; i < 70; i++ {
		txID := fmt.Sprintf("TXN%06d", s.nextID)
		s.nextID++

		method := methods[rand.Intn(len(methods))]
		currency := currencies[rand.Intn(len(currencies))]
		clientName := clientNames[rand.Intn(len(clientNames))]
		clientID := int64(1000 + rand.Intn(9000))

		var status TransactionStatus
		if pendingCount < 25 && rand.Float64() < 0.3 {
			status = TransactionStatusPending
			pendingCount++
		} else {
			status = statuses[rand.Intn(len(statuses))]
		}

		amount := 50 + rand.Float64()*19950 // $50 - $20,000
		requestedAt := now.Add(-time.Duration(rand.Intn(90*24)) * time.Hour)

		tx := &Transaction{
			ID:          txID,
			ClientID:    clientID,
			ClientName:  clientName,
			Type:        TransactionTypeWithdrawal,
			Amount:      math.Round(amount*100) / 100,
			Currency:    currency,
			Method:      method,
			Status:      status,
			RequestedAt: requestedAt,
		}

		// Add processed info for non-pending transactions
		if status != TransactionStatusPending {
			processedAt := requestedAt.Add(time.Duration(1+rand.Intn(48)) * time.Hour)
			tx.ProcessedAt = &processedAt
			tx.ProcessedBy = "admin@rtx5.com"
		}

		// Add reference numbers based on method
		if method == PaymentMethodBankWire {
			tx.BankRef = fmt.Sprintf("WIRE%d", 100000+rand.Intn(900000))
		} else if method == PaymentMethodCrypto {
			tx.WalletAddress = fmt.Sprintf("0x%x", rand.Uint64())
		}

		if status == TransactionStatusRejected {
			reasons := []string{
				"Insufficient verification",
				"Duplicate request",
				"Account restrictions",
				"Invalid bank details",
			}
			tx.Notes = reasons[rand.Intn(len(reasons))]
		}

		s.transactions[txID] = tx
	}
}

// ============================================
// Handler
// ============================================

// TransactionHandler handles deposit/withdrawal transaction endpoints
type TransactionHandler struct {
	store       *TransactionStore
	authService *auth.Service
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(authService *auth.Service) *TransactionHandler {
	return &TransactionHandler{
		store:       NewTransactionStore(),
		authService: authService,
	}
}

// ============================================
// HTTP Handlers
// ============================================

// HandleGetTransactions returns paginated list of transactions with filters
// GET /admin/transactions?page=1&limit=50&type=deposit&status=pending&method=bank_wire&clientId=1001&dateFrom=2026-01-01&dateTo=2026-02-11
func (h *TransactionHandler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	txType := query.Get("type")
	status := query.Get("status")
	method := query.Get("method")
	clientIDStr := query.Get("clientId")
	dateFrom := query.Get("dateFrom")
	dateTo := query.Get("dateTo")

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	// Filter transactions
	var filtered []*Transaction
	for _, tx := range h.store.transactions {
		// Apply filters
		if txType != "" && string(tx.Type) != txType {
			continue
		}
		if status != "" && string(tx.Status) != status {
			continue
		}
		if method != "" && string(tx.Method) != method {
			continue
		}
		if clientIDStr != "" {
			clientID, _ := strconv.ParseInt(clientIDStr, 10, 64)
			if tx.ClientID != clientID {
				continue
			}
		}
		if dateFrom != "" {
			fromDate, err := time.Parse("2006-01-02", dateFrom)
			if err == nil && tx.RequestedAt.Before(fromDate) {
				continue
			}
		}
		if dateTo != "" {
			toDate, err := time.Parse("2006-01-02", dateTo)
			if err == nil && tx.RequestedAt.After(toDate.Add(24*time.Hour)) {
				continue
			}
		}

		filtered = append(filtered, tx)
	}

	// Sort by requested date (newest first)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].RequestedAt.Before(filtered[j].RequestedAt) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Paginate
	total := len(filtered)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedTx := filtered[start:end]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"transactions": paginatedTx,
			"pagination": map[string]interface{}{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"totalPages": (total + limit - 1) / limit,
			},
		},
	})
}

// HandleGetTransactionByID returns full details of a single transaction
// GET /admin/transactions/:id
func (h *TransactionHandler) HandleGetTransactionByID(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract transaction ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}
	txID := parts[len(parts)-1]

	h.store.mu.RLock()
	tx, exists := h.store.transactions[txID]
	h.store.mu.RUnlock()

	if !exists {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    tx,
	})
}

// HandleApproveTransaction approves a pending transaction
// PUT /admin/transactions/:id/approve
func (h *TransactionHandler) HandleApproveTransaction(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract transaction ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}
	txID := parts[len(parts)-2]

	h.store.mu.Lock()
	tx, exists := h.store.transactions[txID]
	if !exists {
		h.store.mu.Unlock()
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if tx.Status != TransactionStatusPending {
		h.store.mu.Unlock()
		http.Error(w, "Transaction is not pending", http.StatusBadRequest)
		return
	}

	// Approve the transaction
	now := time.Now()
	tx.Status = TransactionStatusApproved
	tx.ProcessedAt = &now
	tx.ProcessedBy = "admin@rtx5.com"
	tx.Notes = "Approved by admin"
	h.store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Transaction approved successfully",
		"data":    tx,
	})
}

// HandleRejectTransaction rejects a pending transaction
// PUT /admin/transactions/:id/reject
func (h *TransactionHandler) HandleRejectTransaction(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract transaction ID from URL path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}
	txID := parts[len(parts)-2]

	// Parse request body for rejection reason
	var reqBody struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if reqBody.Reason == "" {
		reqBody.Reason = "Rejected by admin"
	}

	h.store.mu.Lock()
	tx, exists := h.store.transactions[txID]
	if !exists {
		h.store.mu.Unlock()
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	if tx.Status != TransactionStatusPending {
		h.store.mu.Unlock()
		http.Error(w, "Transaction is not pending", http.StatusBadRequest)
		return
	}

	// Reject the transaction
	now := time.Now()
	tx.Status = TransactionStatusRejected
	tx.ProcessedAt = &now
	tx.ProcessedBy = "admin@rtx5.com"
	tx.Notes = reqBody.Reason
	h.store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Transaction rejected successfully",
		"data":    tx,
	})
}

// HandleGetTransactionStats returns transaction statistics
// GET /admin/transactions/stats
func (h *TransactionHandler) HandleGetTransactionStats(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats := TransactionStats{}

	for _, tx := range h.store.transactions {
		// Count pending
		if tx.Status == TransactionStatusPending {
			stats.PendingCount++
		}

		// Count processing
		if tx.Status == TransactionStatusProcessing {
			stats.ProcessingCount++
		}

		// Today's deposits and withdrawals
		if tx.RequestedAt.After(todayStart) || tx.RequestedAt.Equal(todayStart) {
			if tx.Type == TransactionTypeDeposit && (tx.Status == TransactionStatusApproved || tx.Status == TransactionStatusCompleted) {
				stats.TodayDepositsTotal += tx.Amount
				stats.TodayDepositsCount++
			} else if tx.Type == TransactionTypeWithdrawal && (tx.Status == TransactionStatusApproved || tx.Status == TransactionStatusCompleted) {
				stats.TodayWithdrawalsTotal += tx.Amount
				stats.TodayWithdrawalsCount++
			}
		}
	}

	stats.TodayDepositsTotal = math.Round(stats.TodayDepositsTotal*100) / 100
	stats.TodayWithdrawalsTotal = math.Round(stats.TodayWithdrawalsTotal*100) / 100

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// HandleGetDailyTransactions returns daily deposit/withdrawal totals for last 30 days
// GET /admin/transactions/daily
func (h *TransactionHandler) HandleGetDailyTransactions(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	now := time.Now()
	dailyData := make(map[string]*DailyTransactionSummary)

	// Initialize last 30 days
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		dailyData[dateStr] = &DailyTransactionSummary{
			Date:              dateStr,
			DepositsTotal:     0,
			WithdrawalsTotal:  0,
			DepositsCount:     0,
			WithdrawalsCount:  0,
			NetFlow:           0,
		}
	}

	// Aggregate transactions by day
	for _, tx := range h.store.transactions {
		dateStr := tx.RequestedAt.Format("2006-01-02")
		summary, exists := dailyData[dateStr]
		if !exists {
			continue
		}

		if tx.Status == TransactionStatusApproved || tx.Status == TransactionStatusCompleted {
			if tx.Type == TransactionTypeDeposit {
				summary.DepositsTotal += tx.Amount
				summary.DepositsCount++
			} else if tx.Type == TransactionTypeWithdrawal {
				summary.WithdrawalsTotal += tx.Amount
				summary.WithdrawalsCount++
			}
		}
	}

	// Calculate net flow and round amounts
	var result []DailyTransactionSummary
	for _, summary := range dailyData {
		summary.DepositsTotal = math.Round(summary.DepositsTotal*100) / 100
		summary.WithdrawalsTotal = math.Round(summary.WithdrawalsTotal*100) / 100
		summary.NetFlow = math.Round((summary.DepositsTotal-summary.WithdrawalsTotal)*100) / 100
		result = append(result, *summary)
	}

	// Sort by date (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Date < result[j].Date {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// HandleCreateManualTransaction creates a manual deposit/withdrawal entry
// POST /admin/transactions/manual
func (h *TransactionHandler) HandleCreateManualTransaction(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		ClientID      int64           `json:"clientId"`
		ClientName    string          `json:"clientName"`
		Type          TransactionType `json:"type"`
		Amount        float64         `json:"amount"`
		Currency      string          `json:"currency"`
		Method        PaymentMethod   `json:"method"`
		Notes         string          `json:"notes"`
		BankRef       string          `json:"bankRef"`
		WalletAddress string          `json:"walletAddress"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if reqBody.ClientID == 0 || reqBody.ClientName == "" || reqBody.Amount <= 0 || reqBody.Currency == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	txID := fmt.Sprintf("TXN%06d", h.store.nextID)
	h.store.nextID++

	now := time.Now()
	tx := &Transaction{
		ID:            txID,
		ClientID:      reqBody.ClientID,
		ClientName:    reqBody.ClientName,
		Type:          reqBody.Type,
		Amount:        math.Round(reqBody.Amount*100) / 100,
		Currency:      reqBody.Currency,
		Method:        reqBody.Method,
		Status:        TransactionStatusCompleted,
		RequestedAt:   now,
		ProcessedAt:   &now,
		ProcessedBy:   "admin@rtx5.com",
		Notes:         reqBody.Notes,
		BankRef:       reqBody.BankRef,
		WalletAddress: reqBody.WalletAddress,
	}

	h.store.transactions[txID] = tx
	h.store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Manual transaction created successfully",
		"data":    tx,
	})
}

// ============================================
// Helper Functions
// ============================================

