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

	"github.com/epic1st/rtx/backend/auth"
)

// DemoAccount represents a paper trading demo account
type DemoAccount struct {
	ID             string    `json:"id"`
	OwnerName      string    `json:"owner_name"`
	Email          string    `json:"email"`
	Balance        float64   `json:"balance"`
	InitialBalance float64   `json:"initial_balance"`
	Leverage       int       `json:"leverage"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
	TotalTrades    int       `json:"total_trades"`
	TotalPnL       float64   `json:"total_pnl"`
	WinRate        float64   `json:"win_rate"`
	ConvertedAt    *time.Time `json:"converted_at,omitempty"`
	ConvertedToID  string    `json:"converted_to_id,omitempty"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// DemoPosition represents an open position in a demo account
type DemoPosition struct {
	ID           string    `json:"id"`
	AccountID    string    `json:"account_id"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"`
	Volume       float64   `json:"volume"`
	OpenPrice    float64   `json:"open_price"`
	CurrentPrice float64   `json:"current_price"`
	StopLoss     float64   `json:"stop_loss,omitempty"`
	TakeProfit   float64   `json:"take_profit,omitempty"`
	PnL          float64   `json:"pnl"`
	PnLPct       float64   `json:"pnl_pct"`
	OpenTime     time.Time `json:"open_time"`
	Commission   float64   `json:"commission"`
	Swap         float64   `json:"swap"`
}

// DemoStats represents overall demo account statistics
type DemoStats struct {
	TotalAccounts       int     `json:"total_accounts"`
	ActiveAccounts      int     `json:"active_accounts"`
	ExpiredAccounts     int     `json:"expired_accounts"`
	ConvertedAccounts   int     `json:"converted_accounts"`
	AvgWinRate          float64 `json:"avg_win_rate"`
	AvgPnL              float64 `json:"avg_pnl"`
	MostTradedSymbol    string  `json:"most_traded_symbol"`
	AvgTradesPerAccount float64 `json:"avg_trades_per_account"`
	ConversionRate      float64 `json:"conversion_rate"`
	TotalPositions      int     `json:"total_positions"`
	TotalVolume         float64 `json:"total_volume"`
}

// ConversionRecord represents a demo account that converted to live
type ConversionRecord struct {
	DemoAccountID   string    `json:"demo_account_id"`
	OwnerName       string    `json:"owner_name"`
	Email           string    `json:"email"`
	DemoBalance     float64   `json:"demo_balance"`
	DemoPnL         float64   `json:"demo_pnl"`
	DemoTrades      int       `json:"demo_trades"`
	DemoWinRate     float64   `json:"demo_win_rate"`
	ConvertedAt     time.Time `json:"converted_at"`
	LiveAccountID   string    `json:"live_account_id"`
	InitialDeposit  float64   `json:"initial_deposit"`
	DaysUntilConv   int       `json:"days_until_conversion"`
}

// PaperTradingStore manages demo accounts and positions
type PaperTradingStore struct {
	mu        sync.RWMutex
	accounts  []DemoAccount
	positions []DemoPosition
}

// NewPaperTradingStore creates a new paper trading store with mock data
func NewPaperTradingStore() *PaperTradingStore {
	store := &PaperTradingStore{}
	store.accounts = generateMockDemoAccounts()
	store.positions = generateMockDemoPositions(store.accounts)
	store.updateAccountMetrics()
	return store
}

// generateMockDemoAccounts creates 50 demo accounts with varied activity
func generateMockDemoAccounts() []DemoAccount {
	accounts := make([]DemoAccount, 0, 50)
	rand.Seed(time.Now().UnixNano())

	names := []string{
		"John Smith", "Sarah Johnson", "Michael Brown", "Emily Davis", "David Wilson",
		"Lisa Anderson", "Robert Taylor", "Jennifer Thomas", "William Jackson", "Mary White",
		"James Harris", "Patricia Martin", "Christopher Thompson", "Linda Garcia", "Daniel Martinez",
		"Barbara Robinson", "Matthew Clark", "Nancy Rodriguez", "Anthony Lewis", "Karen Lee",
		"Mark Walker", "Betty Hall", "Donald Allen", "Helen Young", "Paul Hernandez",
		"Sandra King", "Andrew Wright", "Donna Lopez", "Joshua Hill", "Carol Scott",
		"Kevin Baker", "Michelle Adams", "Brian Perez", "Deborah Roberts", "George Turner",
		"Stephanie Phillips", "Edward Campbell", "Sharon Mitchell", "Ronald Carter", "Cynthia Evans",
		"Thomas Moore", "Jessica Taylor", "Joseph Anderson", "Ashley Thomas", "Charles Jackson",
		"Amanda White", "Ryan Harris", "Melissa Martin", "Jacob Thompson", "Samantha Garcia",
	}

	leverages := []int{50, 100, 200, 500}
	initialBalances := []float64{10000, 25000, 50000, 100000}

	for i := 0; i < 50; i++ {
		accountID := fmt.Sprintf("DEMO%d", 10000+i)
		name := names[i]
		email := fmt.Sprintf("%s@example.com", strings.ToLower(strings.ReplaceAll(name, " ", ".")))

		daysAgo := rand.Intn(180) + 1 // Ensure daysAgo is at least 1
		createdAt := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour)
		expiresAt := createdAt.Add(30 * 24 * time.Hour)

		initialBalance := initialBalances[rand.Intn(len(initialBalances))]
		leverage := leverages[rand.Intn(len(leverages))]

		isActive := true
		if time.Now().After(expiresAt) {
			isActive = false
		}

		totalTrades := 0
		totalPnL := 0.0
		winRate := 0.0
		balance := initialBalance

		if daysAgo > 3 {
			totalTrades = rand.Intn(100) + 5
			pnlRange := initialBalance * 0.5
			totalPnL = (rand.Float64()*2 - 1) * pnlRange
			balance = initialBalance + totalPnL
			winRate = 30 + rand.Float64()*50
		}

		// Calculate hours since creation, ensuring it's positive
		hoursSinceCreation := daysAgo * 24
		if hoursSinceCreation < 1 {
			hoursSinceCreation = 1
		}

		account := DemoAccount{
			ID:             accountID,
			OwnerName:      name,
			Email:          email,
			Balance:        balance,
			InitialBalance: initialBalance,
			Leverage:       leverage,
			Currency:       "USD",
			CreatedAt:      createdAt,
			ExpiresAt:      expiresAt,
			IsActive:       isActive,
			TotalTrades:    totalTrades,
			TotalPnL:       totalPnL,
			WinRate:        winRate,
			LastActivityAt: createdAt.Add(time.Duration(rand.Intn(hoursSinceCreation)) * time.Hour),
		}

		if i < 12 && daysAgo > 7 && totalPnL > 0 && totalTrades > 20 {
			daysRange := daysAgo - 7
			if daysRange < 1 {
				daysRange = 1
			}
			conversionDays := 7 + rand.Intn(daysRange)
			convertedAt := createdAt.Add(time.Duration(conversionDays) * 24 * time.Hour)
			account.ConvertedAt = &convertedAt
			account.ConvertedToID = fmt.Sprintf("LIVE%d", 50000+i)
		}

		accounts = append(accounts, account)
	}

	return accounts
}

// generateMockDemoPositions creates 200 demo positions across accounts
func generateMockDemoPositions(accounts []DemoAccount) []DemoPosition {
	positions := make([]DemoPosition, 0, 200)
	rand.Seed(time.Now().UnixNano())

	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD",
		"EURGBP", "EURJPY", "GBPJPY", "XAUUSD", "XAGUSD", "BTCUSD", "ETHUSD",
	}

	prices := map[string]float64{
		"EURUSD": 1.0850, "GBPUSD": 1.2650, "USDJPY": 148.50, "USDCHF": 0.8850,
		"AUDUSD": 0.6550, "USDCAD": 1.3650, "NZDUSD": 0.5950, "EURGBP": 0.8580,
		"EURJPY": 161.10, "GBPJPY": 187.85, "XAUUSD": 2650.50, "XAGUSD": 30.45,
		"BTCUSD": 95500.0, "ETHUSD": 3450.0,
	}

	positionID := 1
	activeAccounts := make([]DemoAccount, 0)
	for _, acc := range accounts {
		if acc.IsActive && acc.TotalTrades > 0 {
			activeAccounts = append(activeAccounts, acc)
		}
	}

	if len(activeAccounts) == 0 {
		return positions
	}

	positionsPerAccount := 200 / len(activeAccounts)
	if positionsPerAccount == 0 {
		positionsPerAccount = 1
	}

	for _, account := range activeAccounts {
		numPositions := rand.Intn(positionsPerAccount) + 1
		if len(positions) >= 200 {
			break
		}

		for i := 0; i < numPositions && len(positions) < 200; i++ {
			symbol := symbols[rand.Intn(len(symbols))]
			currentPrice := prices[symbol]
			priceVariation := currentPrice * (rand.Float64()*0.04 - 0.02)
			openPrice := currentPrice + priceVariation

			side := "buy"
			if rand.Float64() < 0.5 {
				side = "sell"
			}

			volume := []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0}[rand.Intn(7)]

			var pnl float64
			if side == "buy" {
				pnl = (currentPrice - openPrice) * volume * 100000
			} else {
				pnl = (openPrice - currentPrice) * volume * 100000
			}

			pnlPct := (pnl / (openPrice * volume * 100000)) * 100

			daysAgo := rand.Intn(int(time.Since(account.CreatedAt).Hours() / 24))
			openTime := time.Now().Add(-time.Duration(daysAgo*24+rand.Intn(24)) * time.Hour)

			var sl, tp float64
			if side == "buy" {
				if rand.Float64() < 0.7 {
					sl = openPrice * (1 - 0.01*(1+rand.Float64()*2))
					tp = openPrice * (1 + 0.02*(1+rand.Float64()*3))
				}
			} else {
				if rand.Float64() < 0.7 {
					sl = openPrice * (1 + 0.01*(1+rand.Float64()*2))
					tp = openPrice * (1 - 0.02*(1+rand.Float64()*3))
				}
			}

			commission := volume * 7.0
			swap := float64(daysAgo) * volume * 0.5 * (rand.Float64()*2 - 1)

			position := DemoPosition{
				ID:           fmt.Sprintf("POS%d", 100000+positionID),
				AccountID:    account.ID,
				Symbol:       symbol,
				Side:         side,
				Volume:       volume,
				OpenPrice:    openPrice,
				CurrentPrice: currentPrice,
				StopLoss:     sl,
				TakeProfit:   tp,
				PnL:          pnl - commission + swap,
				PnLPct:       pnlPct,
				OpenTime:     openTime,
				Commission:   commission,
				Swap:         swap,
			}

			positions = append(positions, position)
			positionID++
		}
	}

	return positions
}

// updateAccountMetrics recalculates account metrics from positions
func (s *PaperTradingStore) updateAccountMetrics() {
	accountPnL := make(map[string]float64)
	accountTrades := make(map[string]int)

	for _, pos := range s.positions {
		accountPnL[pos.AccountID] += pos.PnL
		accountTrades[pos.AccountID]++
	}

	for i := range s.accounts {
		if pnl, exists := accountPnL[s.accounts[i].ID]; exists {
			s.accounts[i].Balance = s.accounts[i].InitialBalance + pnl
		}
	}
}

// PaperTradingHandler handles paper trading requests
type PaperTradingHandler struct {
	store       *PaperTradingStore
	authService *auth.Service
}

// NewPaperTradingHandler creates a new paper trading handler
func NewPaperTradingHandler(store *PaperTradingStore, authService *auth.Service) *PaperTradingHandler {
	return &PaperTradingHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListAccounts handles GET /admin/demo-accounts
func (h *PaperTradingHandler) HandleListAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	status := r.URL.Query().Get("status")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 50
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	filtered := make([]DemoAccount, 0)
	for _, acc := range h.store.accounts {
		if status == "active" && !acc.IsActive {
			continue
		}
		if status == "expired" && acc.IsActive {
			continue
		}
		if status == "converted" && acc.ConvertedAt == nil {
			continue
		}
		filtered = append(filtered, acc)
	}

	totalCount := len(filtered)
	start := (page - 1) * limit
	end := start + limit

	if start >= totalCount {
		filtered = []DemoAccount{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		filtered = filtered[start:end]
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    filtered,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       totalCount,
			"total_pages": int(math.Ceil(float64(totalCount) / float64(limit))),
		},
	})
}

// HandleCreateAccount handles POST /admin/demo-accounts
func (h *PaperTradingHandler) HandleCreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		OwnerName      string  `json:"owner_name"`
		Email          string  `json:"email"`
		InitialBalance float64 `json:"initial_balance"`
		Leverage       int     `json:"leverage"`
		ExpiryDays     int     `json:"expiry_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.OwnerName == "" || req.Email == "" {
		http.Error(w, "Owner name and email are required", http.StatusBadRequest)
		return
	}

	if req.InitialBalance <= 0 {
		req.InitialBalance = 10000
	}
	if req.Leverage <= 0 {
		req.Leverage = 100
	}
	if req.ExpiryDays <= 0 {
		req.ExpiryDays = 30
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	accountID := fmt.Sprintf("DEMO%d", 10000+len(h.store.accounts))
	now := time.Now()

	account := DemoAccount{
		ID:             accountID,
		OwnerName:      req.OwnerName,
		Email:          req.Email,
		Balance:        req.InitialBalance,
		InitialBalance: req.InitialBalance,
		Leverage:       req.Leverage,
		Currency:       "USD",
		CreatedAt:      now,
		ExpiresAt:      now.Add(time.Duration(req.ExpiryDays) * 24 * time.Hour),
		IsActive:       true,
		TotalTrades:    0,
		TotalPnL:       0,
		WinRate:        0,
		LastActivityAt: now,
	}

	h.store.accounts = append(h.store.accounts, account)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Demo account created successfully",
		"data":    account,
	})
}

// HandleGetAccount handles GET /admin/demo-accounts/:id
func (h *PaperTradingHandler) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}
	accountID := pathParts[3]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var account *DemoAccount
	for i := range h.store.accounts {
		if h.store.accounts[i].ID == accountID {
			account = &h.store.accounts[i]
			break
		}
	}

	if account == nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	positions := make([]DemoPosition, 0)
	for _, pos := range h.store.positions {
		if pos.AccountID == accountID {
			positions = append(positions, pos)
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"account":   account,
		"positions": positions,
		"position_count": len(positions),
	})
}

// HandleUpdateAccount handles PUT /admin/demo-accounts/:id
func (h *PaperTradingHandler) HandleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}
	accountID := pathParts[3]

	var req struct {
		ExtendDays     int     `json:"extend_days"`
		ResetBalance   bool    `json:"reset_balance"`
		NewLeverage    int     `json:"new_leverage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	var account *DemoAccount
	for i := range h.store.accounts {
		if h.store.accounts[i].ID == accountID {
			account = &h.store.accounts[i]
			break
		}
	}

	if account == nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	if req.ExtendDays > 0 {
		account.ExpiresAt = account.ExpiresAt.Add(time.Duration(req.ExtendDays) * 24 * time.Hour)
		account.IsActive = true
	}

	if req.ResetBalance {
		account.Balance = account.InitialBalance
		account.TotalPnL = 0
		account.TotalTrades = 0
		account.WinRate = 0
	}

	if req.NewLeverage > 0 {
		account.Leverage = req.NewLeverage
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Account updated successfully",
		"data":    account,
	})
}

// HandleDeleteAccount handles DELETE /admin/demo-accounts/:id
func (h *PaperTradingHandler) HandleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}
	accountID := pathParts[3]

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	found := false
	for i := range h.store.accounts {
		if h.store.accounts[i].ID == accountID {
			h.store.accounts[i].IsActive = false
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Account deactivated successfully",
	})
}

// HandleGetPositions handles GET /admin/demo-accounts/:id/positions
func (h *PaperTradingHandler) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}
	accountID := pathParts[3]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	positions := make([]DemoPosition, 0)
	totalPnL := 0.0
	totalVolume := 0.0

	for _, pos := range h.store.positions {
		if pos.AccountID == accountID {
			positions = append(positions, pos)
			totalPnL += pos.PnL
			totalVolume += pos.Volume
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"positions":    positions,
		"count":        len(positions),
		"total_pnl":    totalPnL,
		"total_volume": totalVolume,
	})
}

// HandleGetStats handles GET /admin/demo-accounts/stats
func (h *PaperTradingHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	stats := DemoStats{
		TotalAccounts:  len(h.store.accounts),
		TotalPositions: len(h.store.positions),
	}

	totalWinRate := 0.0
	totalPnL := 0.0
	totalTrades := 0
	totalVolume := 0.0
	symbolCounts := make(map[string]int)

	for _, acc := range h.store.accounts {
		if acc.IsActive {
			stats.ActiveAccounts++
		} else if time.Now().After(acc.ExpiresAt) {
			stats.ExpiredAccounts++
		}

		if acc.ConvertedAt != nil {
			stats.ConvertedAccounts++
		}

		totalWinRate += acc.WinRate
		totalPnL += acc.TotalPnL
		totalTrades += acc.TotalTrades
	}

	for _, pos := range h.store.positions {
		totalVolume += pos.Volume
		symbolCounts[pos.Symbol]++
	}

	if stats.TotalAccounts > 0 {
		stats.AvgWinRate = totalWinRate / float64(stats.TotalAccounts)
		stats.AvgPnL = totalPnL / float64(stats.TotalAccounts)
		stats.AvgTradesPerAccount = float64(totalTrades) / float64(stats.TotalAccounts)
		stats.ConversionRate = (float64(stats.ConvertedAccounts) / float64(stats.TotalAccounts)) * 100
	}

	stats.TotalVolume = totalVolume

	maxCount := 0
	for symbol, count := range symbolCounts {
		if count > maxCount {
			maxCount = count
			stats.MostTradedSymbol = symbol
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// HandleGetConversions handles GET /admin/demo-accounts/conversions
func (h *PaperTradingHandler) HandleGetConversions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	conversions := make([]ConversionRecord, 0)

	for _, acc := range h.store.accounts {
		if acc.ConvertedAt != nil {
			daysUntilConv := int(acc.ConvertedAt.Sub(acc.CreatedAt).Hours() / 24)
			initialDeposit := 1000 + rand.Float64()*9000

			conversion := ConversionRecord{
				DemoAccountID:  acc.ID,
				OwnerName:      acc.OwnerName,
				Email:          acc.Email,
				DemoBalance:    acc.Balance,
				DemoPnL:        acc.TotalPnL,
				DemoTrades:     acc.TotalTrades,
				DemoWinRate:    acc.WinRate,
				ConvertedAt:    *acc.ConvertedAt,
				LiveAccountID:  acc.ConvertedToID,
				InitialDeposit: initialDeposit,
				DaysUntilConv:  daysUntilConv,
			}
			conversions = append(conversions, conversion)
		}
	}

	sort.Slice(conversions, func(i, j int) bool {
		return conversions[i].ConvertedAt.After(conversions[j].ConvertedAt)
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"conversions": conversions,
		"count":       len(conversions),
	})
}
