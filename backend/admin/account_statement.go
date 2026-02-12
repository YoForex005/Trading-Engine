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
// Account Statement / Trade Export
// ============================================

type TradeEntry struct {
	ID         int64     `json:"id"`
	ClientID   int64     `json:"clientId"`
	Symbol     string    `json:"symbol"`
	Type       string    `json:"type"`       // buy, sell
	Volume     float64   `json:"volume"`
	EntryPrice float64   `json:"entryPrice"`
	ExitPrice  float64   `json:"exitPrice"`
	PnL        float64   `json:"pnl"`
	Commission float64   `json:"commission"`
	Swap       float64   `json:"swap"`
	OpenTime   time.Time `json:"openTime"`
	CloseTime  time.Time `json:"closeTime"`
}

type TransactionEntry struct {
	ID        int64     `json:"id"`
	ClientID  int64     `json:"clientId"`
	Type      string    `json:"type"` // deposit, withdrawal, fee, commission, rebate
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Reference string    `json:"reference"`
}

type DailyBalance struct {
	Date           string  `json:"date"`
	OpeningBalance float64 `json:"openingBalance"`
	Deposits       float64 `json:"deposits"`
	Withdrawals    float64 `json:"withdrawals"`
	TradingPnL     float64 `json:"tradingPnl"`
	Commissions    float64 `json:"commissions"`
	Swaps          float64 `json:"swaps"`
	ClosingBalance float64 `json:"closingBalance"`
}

type GeneratedStatement struct {
	ClientID       int64     `json:"clientId"`
	ClientName     string    `json:"clientName"`
	FromDate       string    `json:"fromDate"`
	ToDate         string    `json:"toDate"`
	Trades         []TradeEntry      `json:"trades"`
	Transactions   []TransactionEntry `json:"transactions"`
	DailyBalances  []DailyBalance    `json:"dailyBalances"`
	TotalPnL       float64   `json:"totalPnl"`
	TotalDeposits  float64   `json:"totalDeposits"`
	TotalWithdrawals float64 `json:"totalWithdrawals"`
	TotalCommissions float64 `json:"totalCommissions"`
	FinalBalance   float64   `json:"finalBalance"`
	GeneratedAt    time.Time `json:"generatedAt"`
}

type StatementTemplate struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"` // daily, weekly, monthly, custom
	Description string `json:"description"`
	DefaultDays int    `json:"defaultDays"`
}

type StatementHistory struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"clientId"`
	ClientName  string    `json:"clientName"`
	Format      string    `json:"format"` // csv, pdf, xlsx
	FromDate    string    `json:"fromDate"`
	ToDate      string    `json:"toDate"`
	GeneratedAt time.Time `json:"generatedAt"`
	DownloadURL string    `json:"downloadUrl"`
	FileSize    int64     `json:"fileSize"` // bytes
}

type ExportRequest struct {
	ClientID int64  `json:"clientId"`
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
	Format   string `json:"format"` // csv, pdf, xlsx
}

type ScheduledStatement struct {
	ID        int64     `json:"id"`
	ClientID  int64     `json:"clientId"`
	ClientName string   `json:"clientName"`
	Frequency string    `json:"frequency"` // daily, weekly, monthly
	Format    string    `json:"format"`
	EmailTo   string    `json:"emailTo"`
	IsActive  bool      `json:"isActive"`
	NextRun   time.Time `json:"nextRun"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ScheduleRequest struct {
	ClientID  int64  `json:"clientId"`
	Frequency string `json:"frequency"`
	Format    string `json:"format"`
	EmailTo   string `json:"emailTo"`
}

type StatementStats struct {
	TotalGenerated     int                `json:"totalGenerated"`
	PopularFormats     map[string]int     `json:"popularFormats"`
	AvgGenerationTime  float64            `json:"avgGenerationTime"` // ms
	TopClients         []ClientStatCount  `json:"topClients"`
	LastUpdated        time.Time          `json:"lastUpdated"`
}

type ClientStatCount struct {
	ClientID   int64  `json:"clientId"`
	ClientName string `json:"clientName"`
	Count      int    `json:"count"`
}

// ============================================
// Service
// ============================================

type StatementService struct {
	trades            map[int64]*TradeEntry
	transactions      map[int64]*TransactionEntry
	templates         map[int64]*StatementTemplate
	history           map[int64]*StatementHistory
	scheduled         map[int64]*ScheduledStatement
	clientNames       map[int64]string
	nextTradeID       int64
	nextTxID          int64
	nextHistoryID     int64
	nextScheduleID    int64
	mu                sync.RWMutex
}

func NewStatementService() *StatementService {
	s := &StatementService{
		trades:         make(map[int64]*TradeEntry),
		transactions:   make(map[int64]*TransactionEntry),
		templates:      make(map[int64]*StatementTemplate),
		history:        make(map[int64]*StatementHistory),
		scheduled:      make(map[int64]*ScheduledStatement),
		clientNames:    make(map[int64]string),
		nextTradeID:    1,
		nextTxID:       1,
		nextHistoryID:  1,
		nextScheduleID: 1,
	}
	s.initializeTemplates()
	s.generateMockData()
	return s
}

func (s *StatementService) initializeTemplates() {
	templates := []StatementTemplate{
		{ID: 1, Name: "Daily Summary", Type: "daily", Description: "Daily trading summary with P&L breakdown", DefaultDays: 1},
		{ID: 2, Name: "Weekly P&L", Type: "weekly", Description: "Weekly profit and loss statement", DefaultDays: 7},
		{ID: 3, Name: "Monthly Full", Type: "monthly", Description: "Comprehensive monthly statement with all transactions", DefaultDays: 30},
		{ID: 4, Name: "Custom Range", Type: "custom", Description: "Custom date range statement", DefaultDays: 0},
	}

	for _, t := range templates {
		template := t
		s.templates[template.ID] = &template
	}
}

func (s *StatementService) generateMockData() {
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
		"Ryan Richardson", "Helen Cox", "Jacob Howard", "Sharon Ward",
		"Gary Torres", "Cynthia Peterson", "Nicholas Gray", "Kathleen Ramirez",
		"Eric James", "Angela Watson", "Jonathan Brooks", "Shirley Kelly",
		"Stephen Sanders", "Anna Price", "Larry Bennett", "Brenda Wood",
		"Justin Ross", "Pamela Henderson", "Scott Coleman", "Nicole Jenkins",
		"Frank Perry", "Christine Powell", "Dennis Long", "Janet Patterson",
		"Jerry Hughes", "Carolyn Flores", "Alexander Washington", "Frances Butler",
		"Tyler Simmons", "Marie Foster", "Aaron Gonzales", "Diane Bryant",
		"Jose Alexander", "Joyce Russell", "Adam Griffin", "Evelyn Hayes",
		"Nathan Diaz", "Ruby Myers", "Zachary Ford", "Phyllis Hamilton",
		"Douglas Graham", "Gloria Sullivan", "Peter Wallace", "Virginia Woods",
	}

	for i := int64(1); i <= 100; i++ {
		s.clientNames[i] = clientNames[i-1]
	}

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "NZDUSD", "EURGBP", "EURJPY", "GBPJPY", "AUDJPY"}
	now := time.Now()

	// Generate 2000 trade entries over last 90 days
	for i := 0; i < 2000; i++ {
		clientID := int64(1 + rand.Intn(100))
		symbol := symbols[rand.Intn(len(symbols))]
		tradeType := "buy"
		if rand.Float64() < 0.5 {
			tradeType = "sell"
		}

		volume := 0.1 + rand.Float64()*9.9 // 0.1 - 10 lots
		entryPrice := 1.0000 + rand.Float64()*0.5000

		// Generate realistic P&L (70% winning trades)
		var pnl float64
		if rand.Float64() < 0.7 {
			pnl = 50.0 + rand.Float64()*450.0 // $50-$500 profit
		} else {
			pnl = -500.0 + rand.Float64()*450.0 // -$500 to -$50 loss
		}

		exitPrice := entryPrice + (pnl / (volume * 100000))
		commission := volume * 7.0 // $7 per lot
		swap := -2.0 + rand.Float64()*4.0 // -$2 to +$2

		openTime := now.Add(-time.Duration(rand.Intn(90*24)) * time.Hour)
		closeTime := openTime.Add(time.Duration(1+rand.Intn(72)) * time.Hour)

		trade := &TradeEntry{
			ID:         s.nextTradeID,
			ClientID:   clientID,
			Symbol:     symbol,
			Type:       tradeType,
			Volume:     volume,
			EntryPrice: entryPrice,
			ExitPrice:  exitPrice,
			PnL:        pnl,
			Commission: commission,
			Swap:       swap,
			OpenTime:   openTime,
			CloseTime:  closeTime,
		}
		s.trades[trade.ID] = trade
		s.nextTradeID++
	}

	// Generate 500 transactions
	txTypes := []string{"deposit", "withdrawal", "fee", "commission", "rebate"}
	for i := 0; i < 500; i++ {
		clientID := int64(1 + rand.Intn(100))
		txType := txTypes[rand.Intn(len(txTypes))]

		var amount float64
		switch txType {
		case "deposit":
			amount = 500.0 + rand.Float64()*9500.0 // $500-$10k
		case "withdrawal":
			amount = -100.0 - rand.Float64()*4900.0 // -$100 to -$5k
		case "fee", "commission":
			amount = -5.0 - rand.Float64()*95.0 // -$5 to -$100
		case "rebate":
			amount = 5.0 + rand.Float64()*45.0 // $5-$50
		}

		status := "completed"
		if txType == "withdrawal" && rand.Float64() < 0.1 {
			status = "pending"
		}

		tx := &TransactionEntry{
			ID:        s.nextTxID,
			ClientID:  clientID,
			Type:      txType,
			Amount:    amount,
			Currency:  "USD",
			Status:    status,
			Timestamp: now.Add(-time.Duration(rand.Intn(90*24)) * time.Hour),
			Reference: fmt.Sprintf("REF-%d", 100000+rand.Intn(900000)),
		}
		s.transactions[tx.ID] = tx
		s.nextTxID++
	}

	// Generate 50 statement history entries
	formats := []string{"csv", "pdf", "xlsx"}
	for i := 0; i < 50; i++ {
		clientID := int64(1 + rand.Intn(100))
		format := formats[rand.Intn(len(formats))]

		generatedAt := now.Add(-time.Duration(rand.Intn(60*24)) * time.Hour)
		daysBack := 1 + rand.Intn(30)
		fromDate := generatedAt.Add(-time.Duration(daysBack) * 24 * time.Hour).Format("2006-01-02")
		toDate := generatedAt.Format("2006-01-02")

		fileSize := int64(10000 + rand.Intn(990000)) // 10KB - 1MB

		history := &StatementHistory{
			ID:          s.nextHistoryID,
			ClientID:    clientID,
			ClientName:  s.clientNames[clientID],
			Format:      format,
			FromDate:    fromDate,
			ToDate:      toDate,
			GeneratedAt: generatedAt,
			DownloadURL: fmt.Sprintf("/downloads/statement-%d.%s", s.nextHistoryID, format),
			FileSize:    fileSize,
		}
		s.history[history.ID] = history
		s.nextHistoryID++
	}

	// Generate 20 scheduled statements
	frequencies := []string{"daily", "weekly", "monthly"}
	for i := 0; i < 20; i++ {
		clientID := int64(1 + rand.Intn(100))
		frequency := frequencies[rand.Intn(len(frequencies))]
		format := formats[rand.Intn(len(formats))]
		isActive := rand.Float64() < 0.8 // 80% active

		var nextRun time.Time
		switch frequency {
		case "daily":
			nextRun = now.Add(24 * time.Hour)
		case "weekly":
			nextRun = now.Add(7 * 24 * time.Hour)
		case "monthly":
			nextRun = now.Add(30 * 24 * time.Hour)
		}

		schedule := &ScheduledStatement{
			ID:         s.nextScheduleID,
			ClientID:   clientID,
			ClientName: s.clientNames[clientID],
			Frequency:  frequency,
			Format:     format,
			EmailTo:    fmt.Sprintf("%s@example.com", strings.ToLower(strings.ReplaceAll(s.clientNames[clientID], " ", "."))),
			IsActive:   isActive,
			NextRun:    nextRun,
			CreatedAt:  now.Add(-time.Duration(rand.Intn(180*24)) * time.Hour),
			UpdatedAt:  now,
		}
		s.scheduled[schedule.ID] = schedule
		s.nextScheduleID++
	}

	log.Printf("[AccountStatement] Generated 2000 trades, 500 transactions, 50 statement history entries, 20 scheduled statements across 100 clients")
}

func (s *StatementService) GenerateStatement(clientID int64, fromDate, toDate string, filters map[string]bool) (*GeneratedStatement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clientName, exists := s.clientNames[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	from, _ := time.Parse("2006-01-02", fromDate)
	to, _ := time.Parse("2006-01-02", toDate)

	stmt := &GeneratedStatement{
		ClientID:     clientID,
		ClientName:   clientName,
		FromDate:     fromDate,
		ToDate:       toDate,
		GeneratedAt:  time.Now(),
	}

	// Collect trades
	if filters["includeTrades"] {
		for _, trade := range s.trades {
			if trade.ClientID == clientID &&
			   trade.CloseTime.After(from) && trade.CloseTime.Before(to.Add(24*time.Hour)) {
				stmt.Trades = append(stmt.Trades, *trade)
				stmt.TotalPnL += trade.PnL
				stmt.TotalCommissions += trade.Commission
			}
		}
	}

	// Collect transactions
	if filters["includeDeposits"] || filters["includeFees"] {
		for _, tx := range s.transactions {
			if tx.ClientID == clientID &&
			   tx.Timestamp.After(from) && tx.Timestamp.Before(to.Add(24*time.Hour)) {
				include := false
				if filters["includeDeposits"] && (tx.Type == "deposit" || tx.Type == "withdrawal") {
					include = true
				}
				if filters["includeFees"] && (tx.Type == "fee" || tx.Type == "commission" || tx.Type == "rebate") {
					include = true
				}

				if include {
					stmt.Transactions = append(stmt.Transactions, *tx)
					if tx.Type == "deposit" {
						stmt.TotalDeposits += tx.Amount
					} else if tx.Type == "withdrawal" {
						stmt.TotalWithdrawals += -tx.Amount
					}
				}
			}
		}
	}

	// Calculate daily balances
	stmt.DailyBalances = s.calculateDailyBalances(clientID, from, to)
	if len(stmt.DailyBalances) > 0 {
		stmt.FinalBalance = stmt.DailyBalances[len(stmt.DailyBalances)-1].ClosingBalance
	}

	return stmt, nil
}

func (s *StatementService) calculateDailyBalances(clientID int64, from, to time.Time) []DailyBalance {
	balances := []DailyBalance{}
	runningBalance := 10000.0 // Starting balance

	for d := from; d.Before(to.Add(24*time.Hour)); d = d.Add(24 * time.Hour) {
		dateStr := d.Format("2006-01-02")
		daily := DailyBalance{
			Date:           dateStr,
			OpeningBalance: runningBalance,
		}

		// Sum trades for this day
		for _, trade := range s.trades {
			if trade.ClientID == clientID && trade.CloseTime.Format("2006-01-02") == dateStr {
				daily.TradingPnL += trade.PnL
				daily.Commissions += trade.Commission
				daily.Swaps += trade.Swap
			}
		}

		// Sum transactions for this day
		for _, tx := range s.transactions {
			if tx.ClientID == clientID && tx.Timestamp.Format("2006-01-02") == dateStr {
				if tx.Type == "deposit" {
					daily.Deposits += tx.Amount
				} else if tx.Type == "withdrawal" {
					daily.Withdrawals += -tx.Amount
				}
			}
		}

		daily.ClosingBalance = daily.OpeningBalance + daily.Deposits - daily.Withdrawals + daily.TradingPnL - daily.Commissions + daily.Swaps
		runningBalance = daily.ClosingBalance

		balances = append(balances, daily)
	}

	return balances
}

func (s *StatementService) ListTemplates() []*StatementTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*StatementTemplate, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, t)
	}

	return templates
}

func (s *StatementService) ListHistory(filters map[string]string, limit, offset int) ([]*StatementHistory, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*StatementHistory
	for _, h := range s.history {
		match := true

		if clientID, ok := filters["clientId"]; ok && clientID != "" {
			id, _ := strconv.ParseInt(clientID, 10, 64)
			if h.ClientID != id {
				match = false
			}
		}
		if format, ok := filters["format"]; ok && format != "" && h.Format != format {
			match = false
		}

		if match {
			filtered = append(filtered, h)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].GeneratedAt.After(filtered[j].GeneratedAt)
	})

	total := len(filtered)
	if offset >= total {
		return []*StatementHistory{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *StatementService) Export(req ExportRequest) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientName, exists := s.clientNames[req.ClientID]
	if !exists {
		return "", fmt.Errorf("client not found")
	}

	// Create history entry
	history := &StatementHistory{
		ID:          s.nextHistoryID,
		ClientID:    req.ClientID,
		ClientName:  clientName,
		Format:      req.Format,
		FromDate:    req.FromDate,
		ToDate:      req.ToDate,
		GeneratedAt: time.Now(),
		DownloadURL: fmt.Sprintf("/downloads/statement-%d-%s.%s", req.ClientID, req.FromDate, req.Format),
		FileSize:    int64(50000 + rand.Intn(950000)),
	}
	s.history[history.ID] = history
	s.nextHistoryID++

	return history.DownloadURL, nil
}

func (s *StatementService) ListScheduled(filters map[string]string) []*ScheduledStatement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*ScheduledStatement
	for _, sch := range s.scheduled {
		match := true

		if clientID, ok := filters["clientId"]; ok && clientID != "" {
			id, _ := strconv.ParseInt(clientID, 10, 64)
			if sch.ClientID != id {
				match = false
			}
		}
		if frequency, ok := filters["frequency"]; ok && frequency != "" && sch.Frequency != frequency {
			match = false
		}
		if active, ok := filters["active"]; ok && active == "true" && !sch.IsActive {
			match = false
		}

		if match {
			filtered = append(filtered, sch)
		}
	}

	return filtered
}

func (s *StatementService) CreateSchedule(req ScheduleRequest) (*ScheduledStatement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	clientName, exists := s.clientNames[req.ClientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	now := time.Now()
	var nextRun time.Time
	switch req.Frequency {
	case "daily":
		nextRun = now.Add(24 * time.Hour)
	case "weekly":
		nextRun = now.Add(7 * 24 * time.Hour)
	case "monthly":
		nextRun = now.Add(30 * 24 * time.Hour)
	default:
		return nil, fmt.Errorf("invalid frequency")
	}

	schedule := &ScheduledStatement{
		ID:         s.nextScheduleID,
		ClientID:   req.ClientID,
		ClientName: clientName,
		Frequency:  req.Frequency,
		Format:     req.Format,
		EmailTo:    req.EmailTo,
		IsActive:   true,
		NextRun:    nextRun,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	s.scheduled[schedule.ID] = schedule
	s.nextScheduleID++

	return schedule, nil
}

func (s *StatementService) UpdateSchedule(id int64, updates map[string]interface{}) (*ScheduledStatement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.scheduled[id]
	if !exists {
		return nil, fmt.Errorf("schedule not found")
	}

	if val, ok := updates["frequency"].(string); ok {
		schedule.Frequency = val
	}
	if val, ok := updates["format"].(string); ok {
		schedule.Format = val
	}
	if val, ok := updates["emailTo"].(string); ok {
		schedule.EmailTo = val
	}
	if val, ok := updates["isActive"].(bool); ok {
		schedule.IsActive = val
	}

	schedule.UpdatedAt = time.Now()

	return schedule, nil
}

func (s *StatementService) GetStats() StatementStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := StatementStats{
		TotalGenerated: len(s.history),
		PopularFormats: make(map[string]int),
		AvgGenerationTime: 1200.0, // Mock: 1200ms avg
		LastUpdated:    time.Now(),
	}

	for _, h := range s.history {
		stats.PopularFormats[h.Format]++
	}

	// Count statements per client
	clientCounts := make(map[int64]int)
	for _, h := range s.history {
		clientCounts[h.ClientID]++
	}

	// Sort and get top 10
	type kv struct {
		ClientID int64
		Count    int
	}
	var sorted []kv
	for k, v := range clientCounts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	if len(sorted) > 10 {
		sorted = sorted[:10]
	}

	for _, item := range sorted {
		stats.TopClients = append(stats.TopClients, ClientStatCount{
			ClientID:   item.ClientID,
			ClientName: s.clientNames[item.ClientID],
			Count:      item.Count,
		})
	}

	return stats
}

// ============================================
// HTTP Handlers
// ============================================

type StatementHandler struct {
	service     *StatementService
	authService *auth.Service
}

func NewStatementHandler(service *StatementService, authService *auth.Service) *StatementHandler {
	return &StatementHandler{
		service:     service,
		authService: authService,
	}
}

// GET /admin/statements/generate/:clientId
func (h *StatementHandler) HandleGenerateStatement(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/statements/generate/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	fromDate := query.Get("from")
	toDate := query.Get("to")
	if fromDate == "" || toDate == "" {
		http.Error(w, "from and to dates are required", http.StatusBadRequest)
		return
	}

	filters := map[string]bool{
		"includeTrades":   query.Get("includeTrades") != "false",
		"includeDeposits": query.Get("includeDeposits") != "false",
		"includeFees":     query.Get("includeFees") != "false",
	}

	statement, err := h.service.GenerateStatement(clientID, fromDate, toDate, filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(statement)
}

// GET /admin/statements/templates
func (h *StatementHandler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates := h.service.ListTemplates()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// GET /admin/statements/history
func (h *StatementHandler) HandleListHistory(w http.ResponseWriter, r *http.Request) {
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
		"clientId": query.Get("clientId"),
		"format":   query.Get("format"),
	}

	history, total := h.service.ListHistory(filters, limit, offset)

	response := map[string]interface{}{
		"history": history,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	}

	json.NewEncoder(w).Encode(response)
}

// POST /admin/statements/export
func (h *StatementHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	downloadURL, err := h.service.Export(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"downloadUrl": downloadURL,
		"format":      req.Format,
		"generated":   true,
	})
}

// GET /admin/statements/scheduled
func (h *StatementHandler) HandleListScheduled(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	filters := map[string]string{
		"clientId":  query.Get("clientId"),
		"frequency": query.Get("frequency"),
		"active":    query.Get("active"),
	}

	schedules := h.service.ListScheduled(filters)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schedules": schedules,
		"count":     len(schedules),
	})
}

// POST /admin/statements/scheduled
func (h *StatementHandler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.service.CreateSchedule(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(schedule)
}

// PUT /admin/statements/scheduled/:id
func (h *StatementHandler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/statements/scheduled/"), "/")
	scheduleID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.service.UpdateSchedule(scheduleID, updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(schedule)
}

// GET /admin/statements/stats
func (h *StatementHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}
