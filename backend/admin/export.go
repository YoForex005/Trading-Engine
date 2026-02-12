package admin

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

type AccountExport struct {
	AccountID     int64   `json:"accountId"`
	AccountNumber string  `json:"accountNumber"`
	Username      string  `json:"username"`
	Balance       float64 `json:"balance"`
	Equity        float64 `json:"equity"`
	Margin        float64 `json:"margin"`
	FreeMargin    float64 `json:"freeMargin"`
	MarginLevel   float64 `json:"marginLevel"`
	OpenPositions int     `json:"openPositions"`
	Leverage      float64 `json:"leverage"`
	MarginMode    string  `json:"marginMode"`
	Status        string  `json:"status"`
	IsDemo        bool    `json:"isDemo"`
	CreatedAt     string  `json:"createdAt"`
}

// ExportHandler handles data export operations
type ExportHandler struct {
	engine      *core.Engine
	authService *AuthService
}

// NewExportHandler creates a new export handler
func NewExportHandler(engine *core.Engine, authService *AuthService) *ExportHandler {
	return &ExportHandler{
		engine:      engine,
		authService: authService,
	}
}

// Helper to parse date filters
func parseDateRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// Default: last 30 days
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	if fromStr != "" {
		parsedFrom, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid 'from' date format, use YYYY-MM-DD")
		}
		from = parsedFrom
	}

	if toStr != "" {
		parsedTo, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid 'to' date format, use YYYY-MM-DD")
		}
		to = parsedTo.Add(24*time.Hour - time.Second) // End of day
	}

	return from, to, nil
}

// Helper to extract session token from Authorization header
func extractSessionToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}
	if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return authHeader[7:], nil
}

// HandleExportTrades exports closed trades as CSV or JSON
func (h *ExportHandler) HandleExportTrades(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	sessionToken, err := extractSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	admin, err := h.authService.ValidateSession(sessionToken, getIPAddress(r))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse date range
	from, to, err := parseDateRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse optional filters
	symbolFilter := r.URL.Query().Get("symbol")
	accountIDFilter := r.URL.Query().Get("accountId")
	var accountID int64
	if accountIDFilter != "" {
		accountID, err = strconv.ParseInt(accountIDFilter, 10, 64)
		if err != nil {
			http.Error(w, "Invalid accountId", http.StatusBadRequest)
			return
		}
	}

	// Get all trades
	allTrades := h.engine.GetTrades(0) // Pass 0 to get all trades

	// Filter trades
	var filteredTrades []core.Trade
	for _, trade := range allTrades {
		// Date range filter
		if trade.ExecutedAt.Before(from) || trade.ExecutedAt.After(to) {
			continue
		}

		// Symbol filter
		if symbolFilter != "" && trade.Symbol != symbolFilter {
			continue
		}

		// Account filter
		if accountID != 0 && trade.AccountID != accountID {
			continue
		}

		// Only include closing trades (with RealizedPnL)
		if trade.RealizedPnL != 0 || trade.Side == "CLOSE_BUY" || trade.Side == "CLOSE_SELL" {
			filteredTrades = append(filteredTrades, trade)
		}
	}

	// Determine format
	format := r.URL.Query().Get("format")
	if format == "csv" {
		h.exportTradesCSV(w, filteredTrades, from, to)
	} else {
		h.exportTradesJSON(w, filteredTrades)
	}

	log.Printf("[AdminExport] Admin %s exported %d trades (%s to %s)", admin.Username, len(filteredTrades), from.Format("2006-01-02"), to.Format("2006-01-02"))
}

func (h *ExportHandler) exportTradesCSV(w http.ResponseWriter, trades []core.Trade, from, to time.Time) {
	filename := fmt.Sprintf("trades_%s_%s.csv", from.Format("2006-01-02"), to.Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Allow-Origin", "*")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"TradeID", "AccountID", "Symbol", "Type", "Volume", "Price",
		"RealizedPnL", "Commission", "ExecutedAt",
	})

	// Write data
	for _, trade := range trades {
		writer.Write([]string{
			strconv.FormatInt(trade.ID, 10),
			strconv.FormatInt(trade.AccountID, 10),
			trade.Symbol,
			trade.Side,
			fmt.Sprintf("%.2f", trade.Volume),
			fmt.Sprintf("%.5f", trade.Price),
			fmt.Sprintf("%.2f", trade.RealizedPnL),
			fmt.Sprintf("%.2f", trade.Commission),
			trade.ExecutedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func (h *ExportHandler) exportTradesJSON(w http.ResponseWriter, trades []core.Trade) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(trades)
}

// HandleExportPositions exports all open positions as CSV or JSON
func (h *ExportHandler) HandleExportPositions(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	sessionToken, err := extractSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	admin, err := h.authService.ValidateSession(sessionToken, getIPAddress(r))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all open positions
	positions := h.engine.GetAllPositions()

	// Determine format
	format := r.URL.Query().Get("format")
	if format == "csv" {
		h.exportPositionsCSV(w, positions)
	} else {
		h.exportPositionsJSON(w, positions)
	}

	log.Printf("[AdminExport] Admin %s exported %d positions", admin.Username, len(positions))
}

func (h *ExportHandler) exportPositionsCSV(w http.ResponseWriter, positions []*core.Position) {
	filename := fmt.Sprintf("positions_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Allow-Origin", "*")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"PositionID", "AccountID", "Symbol", "Type", "Volume", "OpenPrice",
		"CurrentPrice", "UnrealizedPnL", "StopLoss", "TakeProfit", "OpenTime",
	})

	// Write data
	for _, pos := range positions {
		writer.Write([]string{
			strconv.FormatInt(pos.ID, 10),
			strconv.FormatInt(pos.AccountID, 10),
			pos.Symbol,
			pos.Side,
			fmt.Sprintf("%.2f", pos.Volume),
			fmt.Sprintf("%.5f", pos.OpenPrice),
			fmt.Sprintf("%.5f", pos.CurrentPrice),
			fmt.Sprintf("%.2f", pos.UnrealizedPnL),
			fmt.Sprintf("%.5f", pos.SL),
			fmt.Sprintf("%.5f", pos.TP),
			pos.OpenTime.Format("2006-01-02 15:04:05"),
		})
	}
}

func (h *ExportHandler) exportPositionsJSON(w http.ResponseWriter, positions []*core.Position) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(positions)
}

// HandleExportAccounts exports all accounts as CSV or JSON
func (h *ExportHandler) HandleExportAccounts(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	sessionToken, err := extractSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	admin, err := h.authService.ValidateSession(sessionToken, getIPAddress(r))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all accounts with summaries
	var accounts []AccountExport

	// Get all accounts from engine
	allAccounts := h.engine.GetAllAccounts()
	for _, acc := range allAccounts {
		summary, err := h.engine.GetAccountSummary(acc.ID)
		if err != nil {
			continue
		}

		accounts = append(accounts, AccountExport{
			AccountID:     acc.ID,
			AccountNumber: acc.AccountNumber,
			Username:      acc.Username,
			Balance:       summary.Balance,
			Equity:        summary.Equity,
			Margin:        summary.Margin,
			FreeMargin:    summary.FreeMargin,
			MarginLevel:   summary.MarginLevel,
			OpenPositions: summary.OpenPositions,
			Leverage:      acc.Leverage,
			MarginMode:    acc.MarginMode,
			Status:        acc.Status,
			IsDemo:        acc.IsDemo,
			CreatedAt:     time.Unix(acc.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}

	// Determine format
	format := r.URL.Query().Get("format")
	if format == "csv" {
		h.exportAccountsCSV(w, accounts)
	} else {
		h.exportAccountsJSON(w, accounts)
	}

	log.Printf("[AdminExport] Admin %s exported %d accounts", admin.Username, len(accounts))
}

func (h *ExportHandler) exportAccountsCSV(w http.ResponseWriter, accounts []AccountExport) {
	filename := fmt.Sprintf("accounts_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Allow-Origin", "*")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"AccountID", "AccountNumber", "Username", "Balance", "Equity", "Margin",
		"FreeMargin", "MarginLevel", "OpenPositions", "Leverage", "MarginMode",
		"Status", "IsDemo", "CreatedAt",
	})

	// Write data
	for _, acc := range accounts {
		writer.Write([]string{
			strconv.FormatInt(acc.AccountID, 10),
			acc.AccountNumber,
			acc.Username,
			fmt.Sprintf("%.2f", acc.Balance),
			fmt.Sprintf("%.2f", acc.Equity),
			fmt.Sprintf("%.2f", acc.Margin),
			fmt.Sprintf("%.2f", acc.FreeMargin),
			fmt.Sprintf("%.2f", acc.MarginLevel),
			strconv.Itoa(acc.OpenPositions),
			fmt.Sprintf("%.0f", acc.Leverage),
			acc.MarginMode,
			acc.Status,
			fmt.Sprintf("%t", acc.IsDemo),
			acc.CreatedAt,
		})
	}
}

func (h *ExportHandler) exportAccountsJSON(w http.ResponseWriter, accounts []AccountExport) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(accounts)
}

// HandleExportLedger exports ledger entries as CSV or JSON
func (h *ExportHandler) HandleExportLedger(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Authenticate admin
	sessionToken, err := extractSessionToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	admin, err := h.authService.ValidateSession(sessionToken, getIPAddress(r))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse date range
	from, to, err := parseDateRange(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse optional account filter
	accountIDFilter := r.URL.Query().Get("accountId")
	var accountID int64
	if accountIDFilter != "" {
		accountID, err = strconv.ParseInt(accountIDFilter, 10, 64)
		if err != nil {
			http.Error(w, "Invalid accountId", http.StatusBadRequest)
			return
		}
	}

	// Get ledger entries
	ledger := h.engine.GetLedger()
	var allEntries []core.LedgerEntry

	if accountID != 0 {
		allEntries = ledger.GetHistory(accountID, 0)
	} else {
		// Get all entries from all accounts
		allAccounts := h.engine.GetAllAccounts()
		for _, acc := range allAccounts {
			entries := ledger.GetHistory(acc.ID, 0)
			allEntries = append(allEntries, entries...)
		}
	}

	// Filter by date range
	var filteredEntries []core.LedgerEntry
	for _, entry := range allEntries {
		if !entry.CreatedAt.Before(from) && !entry.CreatedAt.After(to) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	// Determine format
	format := r.URL.Query().Get("format")
	if format == "csv" {
		h.exportLedgerCSV(w, filteredEntries, from, to)
	} else {
		h.exportLedgerJSON(w, filteredEntries)
	}

	log.Printf("[AdminExport] Admin %s exported %d ledger entries (%s to %s)", admin.Username, len(filteredEntries), from.Format("2006-01-02"), to.Format("2006-01-02"))
}

func (h *ExportHandler) exportLedgerCSV(w http.ResponseWriter, entries []core.LedgerEntry, from, to time.Time) {
	filename := fmt.Sprintf("ledger_%s_%s.csv", from.Format("2006-01-02"), to.Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Access-Control-Allow-Origin", "*")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"ID", "AccountID", "Type", "Amount", "BalanceAfter", "Currency",
		"Description", "PaymentMethod", "PaymentRef", "Status", "Timestamp",
	})

	// Write data
	for _, entry := range entries {
		writer.Write([]string{
			strconv.FormatInt(entry.ID, 10),
			strconv.FormatInt(entry.AccountID, 10),
			entry.Type,
			fmt.Sprintf("%.2f", entry.Amount),
			fmt.Sprintf("%.2f", entry.BalanceAfter),
			entry.Currency,
			entry.Description,
			entry.PaymentMethod,
			entry.PaymentRef,
			entry.Status,
			entry.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func (h *ExportHandler) exportLedgerJSON(w http.ResponseWriter, entries []core.LedgerEntry) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(entries)
}
