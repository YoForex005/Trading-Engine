package risk

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// CommissionRecord records a single commission charge
type CommissionRecord struct {
	ID             string    `json:"id"`
	TradeID        int64     `json:"tradeId"`
	AccountID      int64     `json:"accountId"`
	Symbol         string    `json:"symbol"`
	Volume         float64   `json:"volume"`        // Lots
	Direction      string    `json:"direction"`     // BUY/SELL
	CommissionType string    `json:"commissionType"` // "per_lot", "percentage", "tiered"
	Rate           float64   `json:"rate"`          // The rate applied
	Amount         float64   `json:"amount"`        // Actual $ charged
	GroupID        string    `json:"groupId,omitempty"` // Trading group if applicable
	ChargedAt      time.Time `json:"chargedAt"`
}

// MonthlyVolume tracks trading volume per account per month
type MonthlyVolume struct {
	AccountID int64
	Month     string  // Format: "2006-01"
	Volume    float64 // Total lots traded this month
}

// CommissionEngine handles commission calculation and tracking
type CommissionEngine struct {
	mu              sync.RWMutex
	engine          *core.Engine
	groupStore      interface{} // TODO: Replace with actual GroupStore type when available
	defaultRate     float64     // Default commission per lot from COMMISSION_DEFAULT env
	history         []CommissionRecord
	monthlyVolumes  map[string]*MonthlyVolume // Key: "accountID_month"
}

// NewCommissionEngine creates a new commission engine
func NewCommissionEngine(engine *core.Engine) *CommissionEngine {
	// Read default commission from env (default: $3.50 per lot per side)
	defaultRate := 3.50
	if envRate := os.Getenv("COMMISSION_DEFAULT"); envRate != "" {
		if rate, err := strconv.ParseFloat(envRate, 64); err == nil && rate > 0 {
			defaultRate = rate
		}
	}

	ce := &CommissionEngine{
		engine:         engine,
		defaultRate:    defaultRate,
		history:        make([]CommissionRecord, 0),
		monthlyVolumes: make(map[string]*MonthlyVolume),
	}

	log.Printf("[CommissionEngine] Initialized with default rate: $%.2f per lot per side", defaultRate)
	return ce
}

// SetGroupStore sets the trading group store reference
func (ce *CommissionEngine) SetGroupStore(store interface{}) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.groupStore = store
	log.Println("[CommissionEngine] GroupStore reference set")
}

// getMonthlyVolume gets or creates monthly volume tracking for an account
func (ce *CommissionEngine) getMonthlyVolume(accountID int64) *MonthlyVolume {
	now := time.Now()
	month := now.Format("2006-01")
	key := fmt.Sprintf("%d_%s", accountID, month)

	if vol, exists := ce.monthlyVolumes[key]; exists {
		return vol
	}

	vol := &MonthlyVolume{
		AccountID: accountID,
		Month:     month,
		Volume:    0,
	}
	ce.monthlyVolumes[key] = vol
	return vol
}

// updateMonthlyVolume adds volume to monthly tracking
func (ce *CommissionEngine) updateMonthlyVolume(accountID int64, volume float64) {
	vol := ce.getMonthlyVolume(accountID)
	vol.Volume += volume
}

// getTieredDiscount calculates discount percentage based on monthly volume
// Tiers:
// - 0-10 lots/month: 0% discount (standard rate)
// - 10-50 lots/month: 10% discount
// - 50-200 lots/month: 20% discount
// - 200+ lots/month: 30% discount
func (ce *CommissionEngine) getTieredDiscount(accountID int64) float64 {
	vol := ce.getMonthlyVolume(accountID)
	monthlyVol := vol.Volume

	if monthlyVol >= 200 {
		return 0.30 // 30% discount
	} else if monthlyVol >= 50 {
		return 0.20 // 20% discount
	} else if monthlyVol >= 10 {
		return 0.10 // 10% discount
	}
	return 0.0 // No discount
}

// CalculateCommission calculates commission for a trade
// Returns: commission amount and CommissionRecord (not yet stored)
func (ce *CommissionEngine) CalculateCommission(accountID int64, symbol string, volume float64, price float64, direction string) (float64, CommissionRecord, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	if volume <= 0 {
		return 0, CommissionRecord{}, fmt.Errorf("invalid volume: %.2f", volume)
	}

	// TODO: Look up client's trading group when GroupStore is available
	// For now, use default "per_lot" model with default rate
	commissionType := "per_lot"
	rate := ce.defaultRate
	var amount float64

	// Get symbol spec for contract size (needed for percentage model)
	symbolSpec := ce.engine.GetOrCreateSymbol(symbol)
	if symbolSpec == nil {
		return 0, CommissionRecord{}, fmt.Errorf("symbol spec not found for %s", symbol)
	}

	// Calculate based on commission type
	switch commissionType {
	case "per_lot":
		// Per-Lot Fixed: amount = rate * volume * 2 (both sides)
		amount = rate * volume * 2.0

	case "percentage":
		// Percentage: amount = (volume * contractSize * price) * rate / 100
		notionalValue := volume * symbolSpec.ContractSize * price
		amount = notionalValue * (rate / 100.0)

	case "tiered":
		// Tiered: Apply volume-based discount to per-lot rate
		discount := ce.getTieredDiscount(accountID)
		baseAmount := rate * volume * 2.0
		amount = baseAmount * (1.0 - discount)

	default:
		return 0, CommissionRecord{}, fmt.Errorf("unknown commission type: %s", commissionType)
	}

	// Build commission record (not yet stored)
	record := CommissionRecord{
		ID:             "", // Will be set when charged
		TradeID:        0,  // Will be set when charged
		AccountID:      accountID,
		Symbol:         symbol,
		Volume:         volume,
		Direction:      direction,
		CommissionType: commissionType,
		Rate:           rate,
		Amount:         amount,
		GroupID:        "", // TODO: Set from trading group
		ChargedAt:      time.Time{}, // Will be set when charged
	}

	return amount, record, nil
}

// ChargeCommission charges commission and stores the record
func (ce *CommissionEngine) ChargeCommission(record CommissionRecord) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	// Set timestamp and ID
	record.ChargedAt = time.Now()
	record.ID = fmt.Sprintf("comm_%d_%d", record.TradeID, record.ChargedAt.UnixNano())

	// Deduct commission from account balance via ledger
	if _, err := ce.engine.GetLedger().Adjust(record.AccountID, -record.Amount, "Trading Commission", "SYSTEM"); err != nil {
		return fmt.Errorf("failed to deduct commission: %w", err)
	}

	// Update monthly volume tracking
	ce.updateMonthlyVolume(record.AccountID, record.Volume)

	// Store record in history
	ce.history = append(ce.history, record)

	log.Printf("[COMMISSION] Charged $%.2f on %s %.2f lots (account %d)",
		record.Amount, record.Symbol, record.Volume, record.AccountID)

	return nil
}

// GetCommissionHistory returns commission history for an account
func (ce *CommissionEngine) GetCommissionHistory(accountID int64, limit int) []CommissionRecord {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	// Filter by account ID
	filtered := make([]CommissionRecord, 0)
	for i := len(ce.history) - 1; i >= 0; i-- {
		if ce.history[i].AccountID == accountID {
			filtered = append(filtered, ce.history[i])
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}

	return filtered
}

// GetAllCommissionHistory returns all commission history (for admin)
func (ce *CommissionEngine) GetAllCommissionHistory(limit int) []CommissionRecord {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	// Return most recent records
	start := 0
	if limit > 0 && len(ce.history) > limit {
		start = len(ce.history) - limit
	}

	// Return copy
	result := make([]CommissionRecord, len(ce.history)-start)
	copy(result, ce.history[start:])

	return result
}

// GetCommissionSummary calculates commission summary for an account
func (ce *CommissionEngine) GetCommissionSummary(accountID int64) map[string]interface{} {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	totalCommissions := 0.0
	thisMonthCommissions := 0.0
	tradeCount := 0
	thisMonthTradeCount := 0

	now := time.Now()
	currentMonth := now.Format("2006-01")

	for _, record := range ce.history {
		if record.AccountID == accountID {
			totalCommissions += record.Amount
			tradeCount++

			// Check if this month
			recordMonth := record.ChargedAt.Format("2006-01")
			if recordMonth == currentMonth {
				thisMonthCommissions += record.Amount
				thisMonthTradeCount++
			}
		}
	}

	avgPerTrade := 0.0
	if tradeCount > 0 {
		avgPerTrade = totalCommissions / float64(tradeCount)
	}

	avgThisMonth := 0.0
	if thisMonthTradeCount > 0 {
		avgThisMonth = thisMonthCommissions / float64(thisMonthTradeCount)
	}

	// Get monthly volume
	monthlyVol := ce.getMonthlyVolume(accountID)

	return map[string]interface{}{
		"totalCommissions":     totalCommissions,
		"thisMonthCommissions": thisMonthCommissions,
		"avgPerTrade":          avgPerTrade,
		"avgThisMonth":         avgThisMonth,
		"tradeCount":           tradeCount,
		"thisMonthTradeCount":  thisMonthTradeCount,
		"monthlyVolume":        monthlyVol.Volume,
		"currentDiscount":      ce.getTieredDiscount(accountID) * 100, // As percentage
	}
}

// HTTP Handlers

// HandleGetCommissionRates handles GET /api/commissions/rates
func (ce *CommissionEngine) HandleGetCommissionRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ce.mu.RLock()
	defer ce.mu.RUnlock()

	// Build commission schedule
	type RateInfo struct {
		Type        string  `json:"type"`
		Rate        float64 `json:"rate"`
		Description string  `json:"description"`
	}

	type TierInfo struct {
		MinVolume   float64 `json:"minVolume"`
		MaxVolume   string  `json:"maxVolume"` // "unlimited" for top tier
		Discount    string  `json:"discount"`
		Description string  `json:"description"`
	}

	rates := []RateInfo{
		{
			Type:        "per_lot",
			Rate:        ce.defaultRate,
			Description: fmt.Sprintf("$%.2f per lot per side (both open and close)", ce.defaultRate),
		},
	}

	tiers := []TierInfo{
		{MinVolume: 0, MaxVolume: "10", Discount: "0%", Description: "Standard rate"},
		{MinVolume: 10, MaxVolume: "50", Discount: "10%", Description: "10% discount on commissions"},
		{MinVolume: 50, MaxVolume: "200", Discount: "20%", Description: "20% discount on commissions"},
		{MinVolume: 200, MaxVolume: "unlimited", Discount: "30%", Description: "30% discount on commissions"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"defaultRate":   ce.defaultRate,
		"rates":         rates,
		"tieredSchedule": tiers,
		"description":   "Commission schedule per trading group",
	})
}

// HandleGetCommissionHistory handles GET /api/commissions/history
func (ce *CommissionEngine) HandleGetCommissionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	accountIDStr := r.URL.Query().Get("accountId")
	limitStr := r.URL.Query().Get("limit")

	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var history []CommissionRecord

	if accountIDStr != "" {
		// Get history for specific account
		accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid accountId", http.StatusBadRequest)
			return
		}
		history = ce.GetCommissionHistory(accountID, limit)
	} else {
		// Get all history (admin only - TODO: add auth check)
		history = ce.GetAllCommissionHistory(limit)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
		"limit":   limit,
	})
}

// HandleGetCommissionSummary handles GET /api/commissions/summary
func (ce *CommissionEngine) HandleGetCommissionSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameter
	accountIDStr := r.URL.Query().Get("accountId")
	if accountIDStr == "" {
		http.Error(w, "Missing required parameter: accountId", http.StatusBadRequest)
		return
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid accountId", http.StatusBadRequest)
		return
	}

	summary := ce.GetCommissionSummary(accountID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
