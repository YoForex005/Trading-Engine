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

// SwapEngine handles overnight financing fees (swap/rollover) for open positions
type SwapEngine struct {
	mu             sync.RWMutex
	engine         *core.Engine
	rolloverTime   string        // Time when swap is applied (HH:MM format, UTC)
	tripleSwapDay  time.Weekday  // Day when triple swap is charged (default: Wednesday)
	history        []SwapRecord
	stopScheduler  chan bool
	schedulerRunning bool
}

// SwapRecord records a single swap transaction
type SwapRecord struct {
	ID          string    `json:"id"`
	AccountID   int64     `json:"accountId"`
	PositionID  int64     `json:"positionId"`
	Symbol      string    `json:"symbol"`
	Direction   string    `json:"direction"` // BUY or SELL
	Volume      float64   `json:"volume"`
	SwapRate    float64   `json:"swapRate"`
	SwapAmount  float64   `json:"swapAmount"` // Actual amount charged/credited
	IsTriple    bool      `json:"isTriple"`   // True if triple swap was applied
	Timestamp   time.Time `json:"timestamp"`
}

// NewSwapEngine creates a new swap engine
func NewSwapEngine(engine *core.Engine) *SwapEngine {
	// Read rollover time from env (default: 21:00 UTC)
	rolloverTime := "21:00"
	if envTime := os.Getenv("ROLLOVER_TIME"); envTime != "" {
		rolloverTime = envTime
	}

	// Read triple swap day from env (default: Wednesday = 3)
	tripleSwapDay := time.Wednesday
	if envDay := os.Getenv("TRIPLE_SWAP_DAY"); envDay != "" {
		if dayInt, err := strconv.Atoi(envDay); err == nil && dayInt >= 0 && dayInt <= 6 {
			tripleSwapDay = time.Weekday(dayInt)
		}
	}

	se := &SwapEngine{
		engine:        engine,
		rolloverTime:  rolloverTime,
		tripleSwapDay: tripleSwapDay,
		history:       make([]SwapRecord, 0),
		stopScheduler: make(chan bool),
	}

	log.Printf("[SwapEngine] Initialized with rollover time: %s UTC, triple swap day: %s", rolloverTime, tripleSwapDay)
	return se
}

// StartSwapScheduler starts the background scheduler that processes daily swap
func (se *SwapEngine) StartSwapScheduler() {
	se.mu.Lock()
	if se.schedulerRunning {
		se.mu.Unlock()
		log.Println("[SwapEngine] Scheduler already running")
		return
	}
	se.schedulerRunning = true
	se.mu.Unlock()

	go func() {
		log.Println("[SwapEngine] Scheduler started")
		ticker := time.NewTicker(1 * time.Minute) // Check every minute
		defer ticker.Stop()

		lastProcessedDate := ""

		for {
			select {
			case <-ticker.C:
				now := time.Now().UTC()
				currentDate := now.Format("2006-01-02")
				currentTime := now.Format("15:04")

				// Check if rollover time has passed and we haven't processed today yet
				if currentTime >= se.rolloverTime && lastProcessedDate != currentDate {
					log.Printf("[SwapEngine] Rollover time reached (%s UTC) - processing daily swap", se.rolloverTime)
					se.ProcessDailySwap()
					lastProcessedDate = currentDate
				}

			case <-se.stopScheduler:
				log.Println("[SwapEngine] Scheduler stopped")
				se.mu.Lock()
				se.schedulerRunning = false
				se.mu.Unlock()
				return
			}
		}
	}()
}

// StopSwapScheduler stops the background scheduler
func (se *SwapEngine) StopSwapScheduler() {
	se.mu.Lock()
	if !se.schedulerRunning {
		se.mu.Unlock()
		return
	}
	se.mu.Unlock()

	close(se.stopScheduler)
	log.Println("[SwapEngine] Stop signal sent to scheduler")
}

// ProcessDailySwap processes swap for all open positions
func (se *SwapEngine) ProcessDailySwap() {
	se.mu.Lock()
	defer se.mu.Unlock()

	now := time.Now().UTC()
	isTripleSwapDay := now.Weekday() == se.tripleSwapDay

	if isTripleSwapDay {
		log.Printf("[SwapEngine] Processing TRIPLE SWAP (day: %s)", se.tripleSwapDay)
	} else {
		log.Printf("[SwapEngine] Processing daily swap")
	}

	// Get all open positions from engine
	positions := se.engine.GetAllPositions()
	processedCount := 0

	for _, position := range positions {
		// Skip closed positions (check if CloseTime is not zero or Status is not empty/closed)
		if !position.CloseTime.IsZero() || position.Status == "CLOSED" {
			continue
		}

		// Calculate swap for this position
		swapAmount, swapRate, err := se.calculateSwapForPosition(position, isTripleSwapDay)
		if err != nil {
			log.Printf("[SwapEngine] Error calculating swap for position %d: %v", position.ID, err)
			continue
		}

		// Skip if swap is zero
		if swapAmount == 0 {
			continue
		}

		// Apply swap to account balance
		account, ok := se.engine.GetAccount(position.AccountID)
		if !ok || account == nil {
			log.Printf("[SwapEngine] Account %d not found for position %d", position.AccountID, position.ID)
			continue
		}

		// Update account balance via ledger
		if _, err := se.engine.GetLedger().Adjust(position.AccountID, swapAmount, "Overnight Swap", "SYSTEM"); err != nil {
			log.Printf("[SwapEngine] Failed to update balance for account %d: %v", position.AccountID, err)
			continue
		}

		// Record swap transaction
		record := SwapRecord{
			ID:         fmt.Sprintf("swap_%d_%d", position.ID, now.Unix()),
			AccountID:  position.AccountID,
			PositionID: position.ID,
			Symbol:     position.Symbol,
			Direction:  position.Side,
			Volume:     position.Volume,
			SwapRate:   swapRate,
			SwapAmount: swapAmount,
			IsTriple:   isTripleSwapDay,
			Timestamp:  now,
		}
		se.history = append(se.history, record)

		processedCount++
		log.Printf("[SwapEngine] Position %d: %s %s %.2f lots, swap: %.2f (rate: %.2f)",
			position.ID, position.Side, position.Symbol, position.Volume, swapAmount, swapRate)
	}

	log.Printf("[SwapEngine] Daily swap complete: %d positions processed", processedCount)
}

// calculateSwapForPosition calculates swap for a single position
// Formula: swap = rate * volume * contractSize / 10
// If triple swap day: swap *= 3
func (se *SwapEngine) calculateSwapForPosition(position *core.Position, isTripleSwapDay bool) (swapAmount float64, swapRate float64, err error) {
	// Get symbol spec
	symbolSpec := se.engine.GetOrCreateSymbol(position.Symbol)
	if symbolSpec == nil {
		return 0, 0, fmt.Errorf("symbol spec not found for %s", position.Symbol)
	}

	// Determine swap rate based on position direction
	if position.Side == "BUY" {
		swapRate = symbolSpec.SwapLong
	} else if position.Side == "SELL" {
		swapRate = symbolSpec.SwapShort
	} else {
		return 0, 0, fmt.Errorf("invalid position side: %s", position.Side)
	}

	// Calculate swap amount
	// Formula: swap = rate * volume * contractSize / 10
	swapAmount = swapRate * position.Volume * symbolSpec.ContractSize / 10

	// Apply triple swap on specified day (e.g., Wednesday to cover weekend)
	if isTripleSwapDay {
		swapAmount *= 3
	}

	return swapAmount, swapRate, nil
}

// CalculateSwap calculates projected swap for a position without applying it
// Useful for showing traders what swap they'll be charged
func (se *SwapEngine) CalculateSwap(symbol string, direction string, volume float64) (swapAmount float64, swapRate float64, err error) {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Get symbol spec
	symbolSpec := se.engine.GetOrCreateSymbol(symbol)
	if symbolSpec == nil {
		return 0, 0, fmt.Errorf("symbol spec not found for %s", symbol)
	}

	// Determine swap rate based on direction
	if direction == "BUY" {
		swapRate = symbolSpec.SwapLong
	} else if direction == "SELL" {
		swapRate = symbolSpec.SwapShort
	} else {
		return 0, 0, fmt.Errorf("invalid direction: %s", direction)
	}

	// Calculate swap amount (single day, not triple)
	swapAmount = swapRate * volume * symbolSpec.ContractSize / 10

	return swapAmount, swapRate, nil
}

// GetSwapHistory returns swap history for an account
func (se *SwapEngine) GetSwapHistory(accountID int64, limit int) []SwapRecord {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Filter by account ID
	filtered := make([]SwapRecord, 0)
	for i := len(se.history) - 1; i >= 0; i-- {
		if se.history[i].AccountID == accountID {
			filtered = append(filtered, se.history[i])
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}

	return filtered
}

// GetAllSwapHistory returns all swap history (for admin)
func (se *SwapEngine) GetAllSwapHistory(limit int) []SwapRecord {
	se.mu.RLock()
	defer se.mu.RUnlock()

	// Return most recent records
	start := 0
	if limit > 0 && len(se.history) > limit {
		start = len(se.history) - limit
	}

	// Return copy to prevent external modification
	result := make([]SwapRecord, len(se.history)-start)
	copy(result, se.history[start:])

	return result
}

// HandleGetSwapRates handles GET /api/swap/rates - returns swap rates for all symbols
func (se *SwapEngine) HandleGetSwapRates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	se.mu.RLock()
	defer se.mu.RUnlock()

	// Get all symbols from engine
	symbols := se.engine.GetSymbols()

	// Build response
	type SwapRateResponse struct {
		Symbol       string  `json:"symbol"`
		SwapLong     float64 `json:"swapLong"`
		SwapShort    float64 `json:"swapShort"`
		ContractSize float64 `json:"contractSize"`
	}

	rates := make([]SwapRateResponse, 0, len(symbols))
	for _, symbolSpec := range symbols {
		rates = append(rates, SwapRateResponse{
			Symbol:       symbolSpec.Symbol,
			SwapLong:     symbolSpec.SwapLong,
			SwapShort:    symbolSpec.SwapShort,
			ContractSize: symbolSpec.ContractSize,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rates":         rates,
		"rolloverTime":  se.rolloverTime,
		"tripleSwapDay": se.tripleSwapDay.String(),
		"count":         len(rates),
	})
}

// HandleGetSwapHistory handles GET /api/swap/history - returns swap history
func (se *SwapEngine) HandleGetSwapHistory(w http.ResponseWriter, r *http.Request) {
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

	var history []SwapRecord

	if accountIDStr != "" {
		// Get history for specific account
		accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid accountId", http.StatusBadRequest)
			return
		}
		history = se.GetSwapHistory(accountID, limit)
	} else {
		// Get all history (admin only - TODO: add auth check)
		history = se.GetAllSwapHistory(limit)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
	})
}

// HandleCalculateSwap handles GET /api/swap/calculate - calculates projected swap
func (se *SwapEngine) HandleCalculateSwap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	symbol := r.URL.Query().Get("symbol")
	direction := r.URL.Query().Get("direction")
	volumeStr := r.URL.Query().Get("volume")

	if symbol == "" || direction == "" || volumeStr == "" {
		http.Error(w, "Missing required parameters: symbol, direction, volume", http.StatusBadRequest)
		return
	}

	volume, err := strconv.ParseFloat(volumeStr, 64)
	if err != nil || volume <= 0 {
		http.Error(w, "Invalid volume", http.StatusBadRequest)
		return
	}

	// Calculate swap
	swapAmount, swapRate, err := se.CalculateSwap(symbol, direction, volume)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol":           symbol,
		"direction":        direction,
		"volume":           volume,
		"swapRate":         swapRate,
		"swapAmountDaily":  swapAmount,
		"swapAmountTriple": swapAmount * 3,
	})
}
