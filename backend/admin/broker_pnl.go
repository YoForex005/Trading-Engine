package admin

import (
	"encoding/csv"
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

// RevenueSource represents revenue broken down by source
type RevenueSource struct {
	Spreads     float64 `json:"spreads"`
	Commissions float64 `json:"commissions"`
	Swaps       float64 `json:"swaps"`
	Markup      float64 `json:"markup"`
	Fees        float64 `json:"fees"`
	Total       float64 `json:"total"`
}

// PnLSummary represents overall broker P&L summary
type PnLSummary struct {
	TotalRevenueMTD float64        `json:"totalRevenueMtd"`
	TotalRevenueYTD float64        `json:"totalRevenueYtd"`
	NetPnLMTD       float64        `json:"netPnlMtd"`
	RevenueMTD      *RevenueSource `json:"revenueMtd"`
	RevenueYTD      *RevenueSource `json:"revenueYtd"`
	BBookPnL        float64        `json:"bBookPnl"`
	ABookPnL        float64        `json:"aBookPnl"`
	Timestamp       time.Time      `json:"timestamp"`
}

// MonthlyRevenue represents revenue breakdown by month
type MonthlyRevenue struct {
	Month   string         `json:"month"`
	Year    int            `json:"year"`
	Sources *RevenueSource `json:"sources"`
}

// SymbolGroupPnL represents P&L for a symbol group
type SymbolGroupPnL struct {
	GroupName string  `json:"groupName"`
	Volume    float64 `json:"volume"`
	Revenue   float64 `json:"revenue"`
	Cost      float64 `json:"cost"`
	NetPnL    float64 `json:"netPnl"`
	PnLPct    float64 `json:"pnlPct"`
	Trades    int     `json:"trades"`
}

// ProfitableSymbol represents a profitable trading symbol
type ProfitableSymbol struct {
	Symbol  string  `json:"symbol"`
	Volume  float64 `json:"volume"`
	NetPnL  float64 `json:"netPnl"`
	Trades  int     `json:"trades"`
	Revenue float64 `json:"revenue"`
}

// ProfitableClient represents a profitable client
type ProfitableClient struct {
	ClientID         int64   `json:"clientId"`
	ClientName       string  `json:"clientName"`
	Volume           float64 `json:"volume"`
	PnLContribution  float64 `json:"pnlContribution"`
	WinRate          float64 `json:"winRate"`
	Trades           int     `json:"trades"`
	AccountBalance   float64 `json:"accountBalance"`
}

// DailyPnL represents daily P&L trend data
type DailyPnL struct {
	Date   string  `json:"date"`
	PnL    float64 `json:"pnl"`
	Volume float64 `json:"volume"`
	Trades int     `json:"trades"`
}

// BBookABookSplit represents B-Book vs A-Book comparison
type BBookABookSplit struct {
	BBook struct {
		Revenue    float64 `json:"revenue"`
		Trades     int     `json:"trades"`
		PnL        float64 `json:"pnl"`
		Percentage float64 `json:"percentage"`
	} `json:"bBook"`
	ABook struct {
		Revenue    float64 `json:"revenue"`
		Trades     int     `json:"trades"`
		PnL        float64 `json:"pnl"`
		Percentage float64 `json:"percentage"`
	} `json:"aBook"`
	TotalRevenue float64 `json:"totalRevenue"`
	TotalTrades  int     `json:"totalTrades"`
}

// ============================================
// In-Memory Store
// ============================================

type BrokerPnLStore struct {
	mu                 sync.RWMutex
	summary            *PnLSummary
	monthlyRevenue     []MonthlyRevenue
	symbolGroups       []SymbolGroupPnL
	profitableSymbols  []ProfitableSymbol
	profitableClients  []ProfitableClient
	dailyPnL           []DailyPnL
	bBookABookSplit    *BBookABookSplit
	lastUpdate         time.Time
}

func NewBrokerPnLStore() *BrokerPnLStore {
	store := &BrokerPnLStore{
		lastUpdate: time.Now(),
	}

	// Initialize with mock data
	store.initializeMockData()

	return store
}

func (s *BrokerPnLStore) initializeMockData() {
	now := time.Now()

	// Initialize summary
	s.summary = &PnLSummary{
		TotalRevenueMTD: 1847392.45,
		TotalRevenueYTD: 14582940.28,
		NetPnLMTD:       523847.92,
		RevenueMTD: &RevenueSource{
			Spreads:     892450.30,
			Commissions: 425890.50,
			Swaps:       198450.75,
			Markup:      245600.90,
			Fees:        85000.00,
			Total:       1847392.45,
		},
		RevenueYTD: &RevenueSource{
			Spreads:     7245890.20,
			Commissions: 3245678.90,
			Swaps:       1589234.45,
			Markup:      1892450.73,
			Fees:        609686.00,
			Total:       14582940.28,
		},
		BBookPnL:  745892.45,
		ABookPnL:  -221044.53,
		Timestamp: now,
	}

	// Initialize 6-month monthly revenue
	months := []string{"Aug", "Sep", "Oct", "Nov", "Dec", "Jan"}
	currentYear := now.Year()
	s.monthlyRevenue = make([]MonthlyRevenue, 0, 6)

	for i, month := range months {
		year := currentYear
		if i < 5 { // Aug-Dec are previous year
			year = currentYear - 1
		}

		baseSpreads := 800000.0 + rand.Float64()*200000.0
		baseCommissions := 350000.0 + rand.Float64()*100000.0
		baseSwaps := 150000.0 + rand.Float64()*80000.0
		baseMarkup := 200000.0 + rand.Float64()*100000.0
		baseFees := 70000.0 + rand.Float64()*30000.0

		total := baseSpreads + baseCommissions + baseSwaps + baseMarkup + baseFees

		s.monthlyRevenue = append(s.monthlyRevenue, MonthlyRevenue{
			Month: month,
			Year:  year,
			Sources: &RevenueSource{
				Spreads:     baseSpreads,
				Commissions: baseCommissions,
				Swaps:       baseSwaps,
				Markup:      baseMarkup,
				Fees:        baseFees,
				Total:       total,
			},
		})
	}

	// Initialize symbol groups
	s.symbolGroups = []SymbolGroupPnL{
		{
			GroupName: "Forex Majors",
			Volume:    12456789.50,
			Revenue:   892450.75,
			Cost:      456234.20,
			NetPnL:    436216.55,
			PnLPct:    48.89,
			Trades:    4523,
		},
		{
			GroupName: "Forex Minors",
			Volume:    5234567.80,
			Revenue:   423890.40,
			Cost:      198450.30,
			NetPnL:    225440.10,
			PnLPct:    53.18,
			Trades:    2145,
		},
		{
			GroupName: "Forex Exotics",
			Volume:    1892345.60,
			Revenue:   245678.90,
			Cost:      145890.50,
			NetPnL:    99788.40,
			PnLPct:    40.63,
			Trades:    892,
		},
		{
			GroupName: "Metals",
			Volume:    8945678.30,
			Revenue:   678945.60,
			Cost:      298450.80,
			NetPnL:    380494.80,
			PnLPct:    56.03,
			Trades:    1823,
		},
		{
			GroupName: "Crypto",
			Volume:    15234567.90,
			Revenue:   1245890.70,
			Cost:      892450.60,
			NetPnL:    353440.10,
			PnLPct:    28.37,
			Trades:    6734,
		},
		{
			GroupName: "Indices",
			Volume:    9876543.20,
			Revenue:   892345.80,
			Cost:      423890.70,
			NetPnL:    468455.10,
			PnLPct:    52.50,
			Trades:    2987,
		},
	}

	// Initialize top 10 profitable symbols
	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
		"XAUUSD", "BTCUSD", "US30", "EURGBP", "GBPJPY",
	}
	s.profitableSymbols = make([]ProfitableSymbol, 0, 10)

	for i, symbol := range symbols {
		volume := 5000000.0 - float64(i)*400000.0 + rand.Float64()*100000.0
		revenue := 450000.0 - float64(i)*35000.0 + rand.Float64()*10000.0
		netPnl := revenue * (0.4 + rand.Float64()*0.2)
		trades := 2000 - i*150 + rand.Intn(100)

		s.profitableSymbols = append(s.profitableSymbols, ProfitableSymbol{
			Symbol:  symbol,
			Volume:  volume,
			NetPnL:  netPnl,
			Trades:  trades,
			Revenue: revenue,
		})
	}

	// Initialize top 10 profitable clients
	s.profitableClients = make([]ProfitableClient, 0, 10)

	for i := 0; i < 10; i++ {
		clientID := int64(1001 + i)
		volume := 3000000.0 - float64(i)*250000.0 + rand.Float64()*80000.0
		pnlContribution := 180000.0 - float64(i)*15000.0 + rand.Float64()*5000.0
		winRate := 68.5 - float64(i)*3.2 + rand.Float64()*2.0
		trades := 1500 - i*120 + rand.Intn(80)
		balance := 50000.0 + rand.Float64()*200000.0

		s.profitableClients = append(s.profitableClients, ProfitableClient{
			ClientID:         clientID,
			ClientName:       fmt.Sprintf("Client-%d", clientID),
			Volume:           volume,
			PnLContribution:  pnlContribution,
			WinRate:          winRate,
			Trades:           trades,
			AccountBalance:   balance,
		})
	}

	// Initialize 30-day daily P&L
	s.dailyPnL = make([]DailyPnL, 0, 30)

	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		basePnL := 15000.0 + rand.Float64()*10000.0
		if rand.Float64() > 0.75 {
			basePnL = -basePnL * 0.3 // 25% chance of negative day
		}
		volume := 500000.0 + rand.Float64()*300000.0
		trades := 150 + rand.Intn(100)

		s.dailyPnL = append(s.dailyPnL, DailyPnL{
			Date:   date.Format("2006-01-02"),
			PnL:    basePnL,
			Volume: volume,
			Trades: trades,
		})
	}

	// Initialize B-Book vs A-Book split
	totalRevenue := s.summary.RevenueMTD.Total
	bBookRevenue := totalRevenue * 0.572
	aBookRevenue := totalRevenue * 0.428
	totalTrades := 18945
	bBookTrades := int(float64(totalTrades) * 0.572)
	aBookTrades := totalTrades - bBookTrades

	s.bBookABookSplit = &BBookABookSplit{
		TotalRevenue: totalRevenue,
		TotalTrades:  totalTrades,
	}
	s.bBookABookSplit.BBook.Revenue = bBookRevenue
	s.bBookABookSplit.BBook.Trades = bBookTrades
	s.bBookABookSplit.BBook.PnL = s.summary.BBookPnL
	s.bBookABookSplit.BBook.Percentage = 57.2
	s.bBookABookSplit.ABook.Revenue = aBookRevenue
	s.bBookABookSplit.ABook.Trades = aBookTrades
	s.bBookABookSplit.ABook.PnL = s.summary.ABookPnL
	s.bBookABookSplit.ABook.Percentage = 42.8
}

// ============================================
// Handler
// ============================================

type BrokerPnLHandler struct {
	store       *BrokerPnLStore
	authService *auth.Service
}

func NewBrokerPnLHandler(authService *auth.Service) *BrokerPnLHandler {
	return &BrokerPnLHandler{
		store:       NewBrokerPnLStore(),
		authService: authService,
	}
}

// ============================================
// HTTP Handlers
// ============================================

// HandleGetSummary returns overall P&L summary
func (h *BrokerPnLHandler) HandleGetSummary(w http.ResponseWriter, r *http.Request) {
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
	summary := h.store.summary
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    summary,
	})
}

// HandleGetBySymbolGroup returns P&L breakdown by symbol group
func (h *BrokerPnLHandler) HandleGetBySymbolGroup(w http.ResponseWriter, r *http.Request) {
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
	groups := h.store.symbolGroups
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    groups,
	})
}

// HandleGetTopSymbols returns top 10 most profitable symbols
func (h *BrokerPnLHandler) HandleGetTopSymbols(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	h.store.mu.RLock()
	symbols := h.store.profitableSymbols
	if len(symbols) > limit {
		symbols = symbols[:limit]
	}
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    symbols,
	})
}

// HandleGetTopClients returns top 10 most profitable clients
func (h *BrokerPnLHandler) HandleGetTopClients(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	h.store.mu.RLock()
	clients := h.store.profitableClients
	if len(clients) > limit {
		clients = clients[:limit]
	}
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    clients,
	})
}

// HandleGetMonthlyRevenue returns 6-month revenue breakdown
func (h *BrokerPnLHandler) HandleGetMonthlyRevenue(w http.ResponseWriter, r *http.Request) {
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
	revenue := h.store.monthlyRevenue
	split := h.store.bBookABookSplit
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"data":           revenue,
		"bBookABookSplit": split,
	})
}

// HandleGetDailyTrend returns 30-day daily P&L trend
func (h *BrokerPnLHandler) HandleGetDailyTrend(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Optional days parameter
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	h.store.mu.RLock()
	dailyData := h.store.dailyPnL
	if len(dailyData) > days {
		dailyData = dailyData[len(dailyData)-days:]
	}
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    dailyData,
	})
}

// HandleExport exports P&L data to CSV
func (h *BrokerPnLHandler) HandleExport(w http.ResponseWriter, r *http.Request) {
	cors(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	format := r.URL.Query().Get("format")
	if format != "csv" {
		http.Error(w, "Only CSV format is supported", http.StatusBadRequest)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=broker_pnl_report.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write summary section
	writer.Write([]string{"BROKER P&L SUMMARY"})
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Revenue MTD", fmt.Sprintf("%.2f", h.store.summary.TotalRevenueMTD)})
	writer.Write([]string{"Total Revenue YTD", fmt.Sprintf("%.2f", h.store.summary.TotalRevenueYTD)})
	writer.Write([]string{"Net P&L MTD", fmt.Sprintf("%.2f", h.store.summary.NetPnLMTD)})
	writer.Write([]string{"B-Book P&L", fmt.Sprintf("%.2f", h.store.summary.BBookPnL)})
	writer.Write([]string{"A-Book P&L", fmt.Sprintf("%.2f", h.store.summary.ABookPnL)})
	writer.Write([]string{})

	// Write revenue breakdown
	writer.Write([]string{"REVENUE BREAKDOWN MTD"})
	writer.Write([]string{"Source", "Amount"})
	writer.Write([]string{"Spreads", fmt.Sprintf("%.2f", h.store.summary.RevenueMTD.Spreads)})
	writer.Write([]string{"Commissions", fmt.Sprintf("%.2f", h.store.summary.RevenueMTD.Commissions)})
	writer.Write([]string{"Swaps", fmt.Sprintf("%.2f", h.store.summary.RevenueMTD.Swaps)})
	writer.Write([]string{"Markup", fmt.Sprintf("%.2f", h.store.summary.RevenueMTD.Markup)})
	writer.Write([]string{"Fees", fmt.Sprintf("%.2f", h.store.summary.RevenueMTD.Fees)})
	writer.Write([]string{})

	// Write symbol groups
	writer.Write([]string{"P&L BY SYMBOL GROUP"})
	writer.Write([]string{"Group", "Volume", "Revenue", "Cost", "Net P&L", "P&L %", "Trades"})
	for _, group := range h.store.symbolGroups {
		writer.Write([]string{
			group.GroupName,
			fmt.Sprintf("%.2f", group.Volume),
			fmt.Sprintf("%.2f", group.Revenue),
			fmt.Sprintf("%.2f", group.Cost),
			fmt.Sprintf("%.2f", group.NetPnL),
			fmt.Sprintf("%.2f", group.PnLPct),
			fmt.Sprintf("%d", group.Trades),
		})
	}
	writer.Write([]string{})

	// Write top symbols
	writer.Write([]string{"TOP 10 PROFITABLE SYMBOLS"})
	writer.Write([]string{"Symbol", "Volume", "Net P&L", "Trades", "Revenue"})
	for _, symbol := range h.store.profitableSymbols {
		writer.Write([]string{
			symbol.Symbol,
			fmt.Sprintf("%.2f", symbol.Volume),
			fmt.Sprintf("%.2f", symbol.NetPnL),
			fmt.Sprintf("%d", symbol.Trades),
			fmt.Sprintf("%.2f", symbol.Revenue),
		})
	}
	writer.Write([]string{})

	// Write top clients
	writer.Write([]string{"TOP 10 PROFITABLE CLIENTS"})
	writer.Write([]string{"Client ID", "Client Name", "Volume", "P&L Contribution", "Win Rate", "Trades", "Balance"})
	for _, client := range h.store.profitableClients {
		writer.Write([]string{
			fmt.Sprintf("%d", client.ClientID),
			client.ClientName,
			fmt.Sprintf("%.2f", client.Volume),
			fmt.Sprintf("%.2f", client.PnLContribution),
			fmt.Sprintf("%.2f%%", client.WinRate),
			fmt.Sprintf("%d", client.Trades),
			fmt.Sprintf("%.2f", client.AccountBalance),
		})
	}
	writer.Write([]string{})

	// Write daily P&L trend
	writer.Write([]string{"DAILY P&L TREND (30 DAYS)"})
	writer.Write([]string{"Date", "P&L", "Volume", "Trades"})
	for _, daily := range h.store.dailyPnL {
		writer.Write([]string{
			daily.Date,
			fmt.Sprintf("%.2f", daily.PnL),
			fmt.Sprintf("%.2f", daily.Volume),
			fmt.Sprintf("%d", daily.Trades),
		})
	}
}

// Helper function to filter data by time range
func filterByTimeRange(timeRange string, customStart, customEnd time.Time) (time.Time, time.Time) {
	now := time.Now()
	var start, end time.Time

	switch strings.ToLower(timeRange) {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = now
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = now.AddDate(0, 0, -(weekday - 1))
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end = now
	case "mtd":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = now
	case "qtd":
		quarter := int(math.Ceil(float64(now.Month()) / 3))
		quarterStartMonth := time.Month((quarter-1)*3 + 1)
		start = time.Date(now.Year(), quarterStartMonth, 1, 0, 0, 0, 0, now.Location())
		end = now
	case "ytd":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end = now
	case "custom":
		start = customStart
		end = customEnd
	default:
		// Default to MTD
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = now
	}

	return start, end
}
