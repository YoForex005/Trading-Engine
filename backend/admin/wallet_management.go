package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Multi-Currency Wallet / Account Balance Management
// ============================================

type CurrencyBalance struct {
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	USDValue float64 `json:"usdValue"`
}

type ClientWallet struct {
	ClientID    int64             `json:"clientId"`
	ClientName  string            `json:"clientName"`
	Balances    []CurrencyBalance `json:"balances"`
	TotalUSD    float64           `json:"totalUsd"`
	LastUpdated time.Time         `json:"lastUpdated"`
}

type BalanceAdjustment struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"clientId"`
	ClientName  string    `json:"clientName"`
	Currency    string    `json:"currency"`
	Amount      float64   `json:"amount"`
	Reason      string    `json:"reason"`
	AdminName   string    `json:"adminName"`
	Timestamp   time.Time `json:"timestamp"`
	OldBalance  float64   `json:"oldBalance"`
	NewBalance  float64   `json:"newBalance"`
}

type Currency struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Symbol       string  `json:"symbol"`
	ExchangeRate float64 `json:"exchangeRate"` // to USD
	IsBase       bool    `json:"isBase"`
	Enabled      bool    `json:"enabled"`
}

type WalletTransaction struct {
	ID            int64     `json:"id"`
	Type          string    `json:"type"` // transfer, adjustment, deposit, withdrawal
	FromClientID  int64     `json:"fromClientId,omitempty"`
	FromClient    string    `json:"fromClient,omitempty"`
	ToClientID    int64     `json:"toClientId,omitempty"`
	ToClient      string    `json:"toClient,omitempty"`
	Currency      string    `json:"currency"`
	Amount        float64   `json:"amount"`
	USDValue      float64   `json:"usdValue"`
	Reason        string    `json:"reason,omitempty"`
	AdminName     string    `json:"adminName,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

type TransferRequest struct {
	FromClientID int64   `json:"fromClientId"`
	ToClientID   int64   `json:"toClientId"`
	Currency     string  `json:"currency"`
	Amount       float64 `json:"amount"`
	Reason       string  `json:"reason"`
	AdminName    string  `json:"adminName"`
}

type AdjustmentRequest struct {
	Currency  string  `json:"currency"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	AdminName string  `json:"adminName"`
}

type WalletStats struct {
	TotalAUM           float64            `json:"totalAum"` // Assets Under Management in USD
	CurrencyDistribution map[string]float64 `json:"currencyDistribution"` // USD value per currency
	TotalClients       int                `json:"totalClients"`
	TotalBalances      int                `json:"totalBalances"`
	LastUpdated        time.Time          `json:"lastUpdated"`
}

type ReconciliationReport struct {
	Currency        string  `json:"currency"`
	ExpectedBalance float64 `json:"expectedBalance"`
	ActualBalance   float64 `json:"actualBalance"`
	Difference      float64 `json:"difference"`
	ClientCount     int     `json:"clientCount"`
	Status          string  `json:"status"` // ok, warning, critical
}

// ============================================
// Service
// ============================================

type WalletService struct {
	wallets       map[int64]*ClientWallet
	currencies    map[string]*Currency
	adjustments   map[int64]*BalanceAdjustment
	transactions  map[int64]*WalletTransaction
	nextAdjustID  int64
	nextTxID      int64
	mu            sync.RWMutex
}

func NewWalletService() *WalletService {
	s := &WalletService{
		wallets:      make(map[int64]*ClientWallet),
		currencies:   make(map[string]*Currency),
		adjustments:  make(map[int64]*BalanceAdjustment),
		transactions: make(map[int64]*WalletTransaction),
		nextAdjustID: 1,
		nextTxID:     1,
	}
	s.initializeCurrencies()
	s.generateMockData()
	return s
}

func (s *WalletService) initializeCurrencies() {
	currencies := []Currency{
		{Code: "USD", Name: "US Dollar", Symbol: "$", ExchangeRate: 1.0, IsBase: true, Enabled: true},
		{Code: "EUR", Name: "Euro", Symbol: "€", ExchangeRate: 1.08, IsBase: false, Enabled: true},
		{Code: "GBP", Name: "British Pound", Symbol: "£", ExchangeRate: 1.26, IsBase: false, Enabled: true},
		{Code: "JPY", Name: "Japanese Yen", Symbol: "¥", ExchangeRate: 0.0067, IsBase: false, Enabled: true},
		{Code: "AUD", Name: "Australian Dollar", Symbol: "A$", ExchangeRate: 0.63, IsBase: false, Enabled: true},
		{Code: "CAD", Name: "Canadian Dollar", Symbol: "C$", ExchangeRate: 0.70, IsBase: false, Enabled: true},
		{Code: "CHF", Name: "Swiss Franc", Symbol: "Fr", ExchangeRate: 1.12, IsBase: false, Enabled: true},
		{Code: "BTC", Name: "Bitcoin", Symbol: "₿", ExchangeRate: 45000.0, IsBase: false, Enabled: true},
	}

	for _, c := range currencies {
		currency := c
		s.currencies[c.Code] = &currency
	}
}

func (s *WalletService) generateMockData() {
	clientNames := []string{
		"John Smith", "Emma Johnson", "Michael Brown", "Sarah Davis",
		"James Wilson", "Emily Taylor", "David Anderson", "Jessica Martinez",
		"Robert Thomas", "Jennifer Jackson", "William White", "Mary Harris",
		"Christopher Martin", "Linda Thompson", "Daniel Garcia", "Patricia Robinson",
		"Matthew Clark", "Barbara Rodriguez", "Joseph Lewis", "Susan Lee",
		"Charles Walker", "Nancy Hall", "Thomas Allen", "Lisa Young",
		"Richard King", "Margaret Wright", "Mark Lopez", "Betty Hill",
		"Donald Scott", "Sandra Green", "Paul Adams", "Ashley Baker",
		"Steven Nelson", "Dorothy Carter", "Andrew Mitchell", "Kimberly Perez",
		"Joshua Roberts", "Elizabeth Turner", "Kenneth Phillips", "Donna Campbell",
		"Kevin Parker", "Carol Evans", "Brian Edwards", "Michelle Collins",
		"George Stewart", "Amanda Sanchez", "Edward Morris", "Melissa Rogers",
		"Ronald Reed", "Deborah Cook", "Timothy Morgan", "Stephanie Bell",
		"Jason Murphy", "Rebecca Bailey", "Jeffrey Rivera", "Laura Cooper",
	}

	activeCurrencies := []string{"USD", "EUR", "GBP", "JPY", "AUD", "CAD", "CHF", "BTC"}
	now := time.Now()

	// Generate client wallets (one per client name)
	for clientID := int64(1); clientID <= int64(len(clientNames)); clientID++ {
		clientName := clientNames[clientID-1]

		var balances []CurrencyBalance
		totalUSD := 0.0

		// Each client has 2-5 currency balances
		numCurrencies := 2 + rand.Intn(4)
		selectedCurrencies := make(map[string]bool)

		// Always include USD
		usdBalance := 1000.0 + rand.Float64()*99000.0 // $1k - $100k
		balances = append(balances, CurrencyBalance{
			Currency: "USD",
			Balance:  usdBalance,
			USDValue: usdBalance,
		})
		selectedCurrencies["USD"] = true
		totalUSD += usdBalance

		// Add random other currencies
		for len(selectedCurrencies) < numCurrencies {
			currency := activeCurrencies[rand.Intn(len(activeCurrencies))]
			if selectedCurrencies[currency] {
				continue
			}

			curr := s.currencies[currency]
			var balance float64
			if currency == "BTC" {
				balance = 0.01 + rand.Float64()*2.0 // 0.01 - 2 BTC
			} else if currency == "JPY" {
				balance = 100000.0 + rand.Float64()*9900000.0 // 100k - 10M JPY
			} else {
				balance = 1000.0 + rand.Float64()*49000.0 // 1k - 50k
			}

			usdValue := balance * curr.ExchangeRate

			balances = append(balances, CurrencyBalance{
				Currency: currency,
				Balance:  balance,
				USDValue: usdValue,
			})
			selectedCurrencies[currency] = true
			totalUSD += usdValue
		}

		wallet := &ClientWallet{
			ClientID:    clientID,
			ClientName:  clientName,
			Balances:    balances,
			TotalUSD:    totalUSD,
			LastUpdated: now,
		}
		s.wallets[clientID] = wallet
	}

	// Generate 100 balance adjustments
	adjustmentReasons := []string{
		"Manual balance correction",
		"Compensatio for system error",
		"Promotional bonus",
		"Refund processed",
		"Commission correction",
		"Trading competition prize",
		"VIP client bonus",
		"Fee waiver credit",
	}

	adminNames := []string{"admin_alice", "admin_bob", "admin_carol", "admin_dave"}

	numClients := len(clientNames)
	for i := 0; i < 100; i++ {
		clientID := int64(1 + rand.Intn(numClients))
		wallet := s.wallets[clientID]
		currency := wallet.Balances[rand.Intn(len(wallet.Balances))].Currency
		amount := -500.0 + rand.Float64()*1500.0 // -$500 to +$1000

		oldBalance := wallet.Balances[0].Balance
		newBalance := oldBalance + amount

		adjustment := &BalanceAdjustment{
			ID:         s.nextAdjustID,
			ClientID:   clientID,
			ClientName: wallet.ClientName,
			Currency:   currency,
			Amount:     amount,
			Reason:     adjustmentReasons[rand.Intn(len(adjustmentReasons))],
			AdminName:  adminNames[rand.Intn(len(adminNames))],
			Timestamp:  now.Add(-time.Duration(rand.Intn(90*24)) * time.Hour),
			OldBalance: oldBalance,
			NewBalance: newBalance,
		}
		s.adjustments[adjustment.ID] = adjustment
		s.nextAdjustID++
	}

	// Generate 200 wallet transactions
	txTypes := []string{"transfer", "adjustment", "deposit", "withdrawal"}

	for i := 0; i < 200; i++ {
		txType := txTypes[rand.Intn(len(txTypes))]
		currency := activeCurrencies[rand.Intn(len(activeCurrencies))]
		curr := s.currencies[currency]

		var tx *WalletTransaction

		switch txType {
		case "transfer":
			fromID := int64(1 + rand.Intn(numClients))
			toID := int64(1 + rand.Intn(numClients))
			if fromID == toID {
				toID = (toID % int64(numClients)) + 1
			}

			amount := 100.0 + rand.Float64()*4900.0
			usdValue := amount * curr.ExchangeRate

			tx = &WalletTransaction{
				ID:           s.nextTxID,
				Type:         "transfer",
				FromClientID: fromID,
				FromClient:   s.wallets[fromID].ClientName,
				ToClientID:   toID,
				ToClient:     s.wallets[toID].ClientName,
				Currency:     currency,
				Amount:       amount,
				USDValue:     usdValue,
				Reason:       "Internal transfer",
				AdminName:    adminNames[rand.Intn(len(adminNames))],
				Timestamp:    now.Add(-time.Duration(rand.Intn(60*24)) * time.Hour),
			}

		case "deposit":
			clientID := int64(1 + rand.Intn(numClients))
			amount := 500.0 + rand.Float64()*9500.0
			usdValue := amount * curr.ExchangeRate

			tx = &WalletTransaction{
				ID:         s.nextTxID,
				Type:       "deposit",
				ToClientID: clientID,
				ToClient:   s.wallets[clientID].ClientName,
				Currency:   currency,
				Amount:     amount,
				USDValue:   usdValue,
				Timestamp:  now.Add(-time.Duration(rand.Intn(60*24)) * time.Hour),
			}

		case "withdrawal":
			clientID := int64(1 + rand.Intn(numClients))
			amount := 500.0 + rand.Float64()*4500.0
			usdValue := amount * curr.ExchangeRate

			tx = &WalletTransaction{
				ID:           s.nextTxID,
				Type:         "withdrawal",
				FromClientID: clientID,
				FromClient:   s.wallets[clientID].ClientName,
				Currency:     currency,
				Amount:       amount,
				USDValue:     usdValue,
				Timestamp:    now.Add(-time.Duration(rand.Intn(60*24)) * time.Hour),
			}

		case "adjustment":
			clientID := int64(1 + rand.Intn(numClients))
			amount := -200.0 + rand.Float64()*1200.0
			usdValue := amount * curr.ExchangeRate

			tx = &WalletTransaction{
				ID:         s.nextTxID,
				Type:       "adjustment",
				ToClientID: clientID,
				ToClient:   s.wallets[clientID].ClientName,
				Currency:   currency,
				Amount:     amount,
				USDValue:   usdValue,
				Reason:     adjustmentReasons[rand.Intn(len(adjustmentReasons))],
				AdminName:  adminNames[rand.Intn(len(adminNames))],
				Timestamp:  now.Add(-time.Duration(rand.Intn(60*24)) * time.Hour),
			}
		}

		s.transactions[tx.ID] = tx
		s.nextTxID++
	}

	log.Printf("[WalletManagement] Generated %d client wallets, 8 currencies, 100 adjustments, 200 transactions", len(s.wallets))
}

func (s *WalletService) ListWallets() []*ClientWallet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallets := make([]*ClientWallet, 0, len(s.wallets))
	for _, w := range s.wallets {
		wallets = append(wallets, w)
	}

	sort.Slice(wallets, func(i, j int) bool {
		return wallets[i].TotalUSD > wallets[j].TotalUSD
	})

	return wallets
}

func (s *WalletService) GetWallet(clientID int64) *ClientWallet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wallets[clientID]
}

func (s *WalletService) AdjustBalance(clientID int64, req AdjustmentRequest) (*BalanceAdjustment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wallet, exists := s.wallets[clientID]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	curr, exists := s.currencies[req.Currency]
	if !exists {
		return nil, fmt.Errorf("currency not supported")
	}

	// Find currency balance
	var oldBalance float64
	found := false
	for i, bal := range wallet.Balances {
		if bal.Currency == req.Currency {
			oldBalance = bal.Balance
			wallet.Balances[i].Balance += req.Amount
			wallet.Balances[i].USDValue = wallet.Balances[i].Balance * curr.ExchangeRate
			found = true
			break
		}
	}

	if !found {
		// Add new currency balance
		oldBalance = 0
		wallet.Balances = append(wallet.Balances, CurrencyBalance{
			Currency: req.Currency,
			Balance:  req.Amount,
			USDValue: req.Amount * curr.ExchangeRate,
		})
	}

	newBalance := oldBalance + req.Amount

	// Recalculate total USD
	totalUSD := 0.0
	for _, bal := range wallet.Balances {
		totalUSD += bal.USDValue
	}
	wallet.TotalUSD = totalUSD
	wallet.LastUpdated = time.Now()

	// Create adjustment record
	adjustment := &BalanceAdjustment{
		ID:         s.nextAdjustID,
		ClientID:   clientID,
		ClientName: wallet.ClientName,
		Currency:   req.Currency,
		Amount:     req.Amount,
		Reason:     req.Reason,
		AdminName:  req.AdminName,
		Timestamp:  time.Now(),
		OldBalance: oldBalance,
		NewBalance: newBalance,
	}
	s.adjustments[adjustment.ID] = adjustment
	s.nextAdjustID++

	// Create transaction
	tx := &WalletTransaction{
		ID:         s.nextTxID,
		Type:       "adjustment",
		ToClientID: clientID,
		ToClient:   wallet.ClientName,
		Currency:   req.Currency,
		Amount:     req.Amount,
		USDValue:   req.Amount * curr.ExchangeRate,
		Reason:     req.Reason,
		AdminName:  req.AdminName,
		Timestamp:  time.Now(),
	}
	s.transactions[tx.ID] = tx
	s.nextTxID++

	return adjustment, nil
}

func (s *WalletService) ListCurrencies() []*Currency {
	s.mu.RLock()
	defer s.mu.RUnlock()

	currencies := make([]*Currency, 0, len(s.currencies))
	for _, c := range s.currencies {
		currencies = append(currencies, c)
	}

	return currencies
}

func (s *WalletService) ListTransactions(filters map[string]string, limit, offset int) ([]*WalletTransaction, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*WalletTransaction
	for _, tx := range s.transactions {
		match := true

		if txType, ok := filters["type"]; ok && txType != "" && tx.Type != txType {
			match = false
		}
		if currency, ok := filters["currency"]; ok && currency != "" && tx.Currency != currency {
			match = false
		}
		if clientID, ok := filters["clientId"]; ok && clientID != "" {
			id, _ := strconv.ParseInt(clientID, 10, 64)
			if tx.FromClientID != id && tx.ToClientID != id {
				match = false
			}
		}

		if match {
			filtered = append(filtered, tx)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	total := len(filtered)
	if offset >= total {
		return []*WalletTransaction{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *WalletService) Transfer(req TransferRequest) (*WalletTransaction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromWallet, exists := s.wallets[req.FromClientID]
	if !exists {
		return nil, fmt.Errorf("source wallet not found")
	}

	toWallet, exists := s.wallets[req.ToClientID]
	if !exists {
		return nil, fmt.Errorf("destination wallet not found")
	}

	curr, exists := s.currencies[req.Currency]
	if !exists {
		return nil, fmt.Errorf("currency not supported")
	}

	// Check if source has sufficient balance
	var fromBalance *CurrencyBalance
	for i := range fromWallet.Balances {
		if fromWallet.Balances[i].Currency == req.Currency {
			fromBalance = &fromWallet.Balances[i]
			break
		}
	}

	if fromBalance == nil || fromBalance.Balance < req.Amount {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Deduct from source
	fromBalance.Balance -= req.Amount
	fromBalance.USDValue = fromBalance.Balance * curr.ExchangeRate

	// Add to destination
	found := false
	for i := range toWallet.Balances {
		if toWallet.Balances[i].Currency == req.Currency {
			toWallet.Balances[i].Balance += req.Amount
			toWallet.Balances[i].USDValue = toWallet.Balances[i].Balance * curr.ExchangeRate
			found = true
			break
		}
	}

	if !found {
		toWallet.Balances = append(toWallet.Balances, CurrencyBalance{
			Currency: req.Currency,
			Balance:  req.Amount,
			USDValue: req.Amount * curr.ExchangeRate,
		})
	}

	// Recalculate totals
	fromWallet.TotalUSD = 0
	for _, bal := range fromWallet.Balances {
		fromWallet.TotalUSD += bal.USDValue
	}
	fromWallet.LastUpdated = time.Now()

	toWallet.TotalUSD = 0
	for _, bal := range toWallet.Balances {
		toWallet.TotalUSD += bal.USDValue
	}
	toWallet.LastUpdated = time.Now()

	// Create transaction
	tx := &WalletTransaction{
		ID:           s.nextTxID,
		Type:         "transfer",
		FromClientID: req.FromClientID,
		FromClient:   fromWallet.ClientName,
		ToClientID:   req.ToClientID,
		ToClient:     toWallet.ClientName,
		Currency:     req.Currency,
		Amount:       req.Amount,
		USDValue:     req.Amount * curr.ExchangeRate,
		Reason:       req.Reason,
		AdminName:    req.AdminName,
		Timestamp:    time.Now(),
	}
	s.transactions[tx.ID] = tx
	s.nextTxID++

	return tx, nil
}

func (s *WalletService) GetStats() WalletStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := WalletStats{
		CurrencyDistribution: make(map[string]float64),
		LastUpdated:          time.Now(),
	}

	for _, wallet := range s.wallets {
		stats.TotalClients++
		stats.TotalAUM += wallet.TotalUSD

		for _, bal := range wallet.Balances {
			stats.TotalBalances++
			stats.CurrencyDistribution[bal.Currency] += bal.USDValue
		}
	}

	return stats
}

func (s *WalletService) GetReconciliation() []ReconciliationReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	reports := make(map[string]*ReconciliationReport)

	for _, wallet := range s.wallets {
		for _, bal := range wallet.Balances {
			if _, exists := reports[bal.Currency]; !exists {
				reports[bal.Currency] = &ReconciliationReport{
					Currency: bal.Currency,
				}
			}
			reports[bal.Currency].ActualBalance += bal.Balance
			reports[bal.Currency].ClientCount++
		}
	}

	// For mock data, expected = actual (in real system, this would come from ledger)
	result := make([]ReconciliationReport, 0, len(reports))
	for _, report := range reports {
		report.ExpectedBalance = report.ActualBalance
		report.Difference = report.ActualBalance - report.ExpectedBalance

		if report.Difference == 0 {
			report.Status = "ok"
		} else if report.Difference < 100 {
			report.Status = "warning"
		} else {
			report.Status = "critical"
		}

		result = append(result, *report)
	}

	return result
}

// ============================================
// HTTP Handlers
// ============================================

type WalletHandler struct {
	service     *WalletService
	authService *auth.Service
}

func NewWalletHandler(service *WalletService, authService *auth.Service) *WalletHandler {
	return &WalletHandler{
		service:     service,
		authService: authService,
	}
}

// GET /admin/wallets
func (h *WalletHandler) HandleListWallets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	wallets := h.service.ListWallets()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallets": wallets,
		"count":   len(wallets),
	})
}

// GET /admin/wallets/:clientId
func (h *WalletHandler) HandleGetWallet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/wallets/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	wallet := h.service.GetWallet(clientID)
	if wallet == nil {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(wallet)
}

// POST /admin/wallets/:clientId/adjust
func (h *WalletHandler) HandleAdjustBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/wallets/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var req AdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	adjustment, err := h.service.AdjustBalance(clientID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(adjustment)
}

// GET /admin/wallets/currencies
func (h *WalletHandler) HandleGetCurrencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currencies := h.service.ListCurrencies()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"currencies": currencies,
		"count":      len(currencies),
	})
}

// GET /admin/wallets/transactions
func (h *WalletHandler) HandleListTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(query.Get("offset"))

	filters := map[string]string{
		"type":     query.Get("type"),
		"currency": query.Get("currency"),
		"clientId": query.Get("clientId"),
	}

	transactions, total := h.service.ListTransactions(filters, limit, offset)

	response := map[string]interface{}{
		"transactions": transactions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	}

	json.NewEncoder(w).Encode(response)
}

// POST /admin/wallets/transfer
func (h *WalletHandler) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.service.Transfer(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(tx)
}

// GET /admin/wallets/stats
func (h *WalletHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// GET /admin/wallets/reconciliation
func (h *WalletHandler) HandleGetReconciliation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	reports := h.service.GetReconciliation()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"count":   len(reports),
	})
}
