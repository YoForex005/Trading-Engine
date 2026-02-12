package risk

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

// TradingEngine interface abstracts the underlying trading engine
type TradingEngine interface {
	GetAllAccounts() []*AccountInfo
	GetAccount(accountID int64) (*AccountInfo, error)
	GetPositions(accountID int64) []*PositionInfo
	ClosePosition(positionID int64, closePrice float64) error
	GetCurrentPrice(symbol string) (bid, ask float64)
	StoreAlert(alert *RiskAlert) error
	StoreLiquidationEvent(event LiquidationEvent)
}

// AccountInfo represents trading account info
type AccountInfo struct {
	ID          int64
	UserID      string
	Balance     float64
	Equity      float64
	Margin      float64
	FreeMargin  float64
	MarginLevel float64
}

// PositionInfo represents position info
type PositionInfo struct {
	ID            int64
	AccountID     int64
	Symbol        string
	Side          string
	Volume        float64
	OpenPrice     float64
	CurrentPrice  float64
	UnrealizedPnL float64
}

// MarginCallEngine monitors margin levels and triggers automated actions
type MarginCallEngine struct {
	engine                TradingEngine
	notificationManager   NotificationSender
	marginWarningLevel    float64 // % - configurable via env
	marginCallLevel       float64 // % - configurable via env
	stopOutLevel          float64 // % - configurable via env
	activeMarginCalls     map[int64]*MarginCall
	activeWarnings        map[int64]time.Time
	restrictedAccounts    map[int64]bool // Accounts with margin call - cannot open new positions
	monitoringActive      bool
	mu                    sync.RWMutex
	stopChan              chan struct{}
}

// NotificationSender interface for sending notifications
type NotificationSender interface {
	Send(userID, notifType string, data map[string]interface{}) error
}

// NewMarginCallEngine creates a new margin call automation engine
func NewMarginCallEngine(engine TradingEngine, notificationManager NotificationSender) *MarginCallEngine {
	// Load thresholds from environment variables with defaults
	warningLevel := getEnvFloat("MARGIN_WARNING_LEVEL", 120.0)
	callLevel := getEnvFloat("MARGIN_CALL_LEVEL", 100.0)
	stopOutLevel := getEnvFloat("STOP_OUT_LEVEL", 50.0)

	log.Printf("[MarginCallEngine] Initialized with thresholds: Warning=%.0f%%, Call=%.0f%%, StopOut=%.0f%%",
		warningLevel, callLevel, stopOutLevel)

	return &MarginCallEngine{
		engine:              engine,
		notificationManager: notificationManager,
		marginWarningLevel:  warningLevel,
		marginCallLevel:     callLevel,
		stopOutLevel:        stopOutLevel,
		activeMarginCalls:   make(map[int64]*MarginCall),
		activeWarnings:      make(map[int64]time.Time),
		restrictedAccounts:  make(map[int64]bool),
		monitoringActive:    false,
		stopChan:            make(chan struct{}),
	}
}

// Start begins monitoring margin levels in a background goroutine
func (mce *MarginCallEngine) Start() {
	mce.mu.Lock()
	if mce.monitoringActive {
		mce.mu.Unlock()
		log.Println("[MarginCallEngine] Already running")
		return
	}
	mce.monitoringActive = true
	mce.mu.Unlock()

	log.Println("[MarginCallEngine] Starting margin monitoring...")

	go mce.monitorLoop()
}

// Stop halts the monitoring loop
func (mce *MarginCallEngine) Stop() {
	mce.mu.Lock()
	if !mce.monitoringActive {
		mce.mu.Unlock()
		return
	}
	mce.monitoringActive = false
	mce.mu.Unlock()

	close(mce.stopChan)
	log.Println("[MarginCallEngine] Stopped margin monitoring")
}

// monitorLoop continuously monitors all accounts
func (mce *MarginCallEngine) monitorLoop() {
	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mce.checkAllAccounts()
		case <-mce.stopChan:
			return
		}
	}
}

// checkAllAccounts checks margin levels for all accounts
func (mce *MarginCallEngine) checkAllAccounts() {
	accounts := mce.engine.GetAllAccounts()

	for _, account := range accounts {
		mce.checkAccount(account)
	}
}

// checkAccount checks a single account's margin level and takes appropriate action
func (mce *MarginCallEngine) checkAccount(account *AccountInfo) {
	// Skip accounts with no margin used
	if account.Margin <= 0 {
		// Clear any active warnings or margin calls
		mce.clearMarginStatus(account.ID)
		return
	}

	marginLevel := account.MarginLevel

	// CRITICAL: Stop Out Level - Auto-liquidate positions
	if marginLevel > 0 && marginLevel <= mce.stopOutLevel {
		mce.handleStopOut(account)
		return
	}

	// URGENT: Margin Call Level - Restrict new positions
	if marginLevel > 0 && marginLevel <= mce.marginCallLevel {
		mce.handleMarginCall(account)
		return
	}

	// WARNING: Margin Warning Level - Send notification
	if marginLevel > 0 && marginLevel <= mce.marginWarningLevel {
		mce.handleMarginWarning(account)
		return
	}

	// Margin level is healthy - clear any active warnings/calls
	if marginLevel > mce.marginWarningLevel {
		mce.clearMarginStatus(account.ID)
	}
}

// handleMarginWarning sends a warning notification
func (mce *MarginCallEngine) handleMarginWarning(account *AccountInfo) {
	mce.mu.RLock()
	lastWarning, exists := mce.activeWarnings[account.ID]
	mce.mu.RUnlock()

	// Throttle warnings - send at most once per minute
	if exists && time.Since(lastWarning) < 1*time.Minute {
		return
	}

	mce.mu.Lock()
	mce.activeWarnings[account.ID] = time.Now()
	mce.mu.Unlock()

	log.Printf("[MarginWarning] Account #%d: Margin level %.2f%% <= %.2f%% (Warning threshold)",
		account.ID, account.MarginLevel, mce.marginWarningLevel)

	// Send notification
	if mce.notificationManager != nil {
		mce.notificationManager.Send(
			account.UserID,
			"margin_warning",
			map[string]interface{}{
				"accountId":    account.ID,
				"marginLevel":  account.MarginLevel,
				"equity":       account.Equity,
				"margin":       account.Margin,
				"freeMargin":   account.FreeMargin,
				"threshold":    mce.marginWarningLevel,
				"severity":     "warning",
				"message":      fmt.Sprintf("Margin level at %.2f%%. Please monitor your positions.", account.MarginLevel),
			},
		)
	}
}

// handleMarginCall triggers margin call actions
func (mce *MarginCallEngine) handleMarginCall(account *AccountInfo) {
	mce.mu.RLock()
	_, alreadyCalled := mce.activeMarginCalls[account.ID]
	mce.mu.RUnlock()

	if alreadyCalled {
		// Already in margin call - just update the status
		return
	}

	// Calculate required deposit to restore margin
	targetEquity := account.Margin * (mce.marginCallLevel / 100.0)
	requiredDeposit := targetEquity - account.Equity

	// Create margin call record
	marginCall := &MarginCall{
		ID:              fmt.Sprintf("MC_%d_%d", account.ID, time.Now().Unix()),
		AccountID:       account.ID,
		MarginLevel:     account.MarginLevel,
		RequiredDeposit: requiredDeposit,
		Severity:        RiskLevelHigh,
		TriggeredAt:     time.Now(),
		Status:          "ACTIVE",
		Actions:         []string{"New positions restricted", "Urgent notification sent"},
	}

	mce.mu.Lock()
	mce.activeMarginCalls[account.ID] = marginCall
	mce.restrictedAccounts[account.ID] = true
	mce.mu.Unlock()

	log.Printf("[MarginCall] Account #%d: MARGIN CALL TRIGGERED - Level %.2f%% <= %.2f%%, deposit needed: $%.2f",
		account.ID, account.MarginLevel, mce.marginCallLevel, requiredDeposit)

	// Send URGENT notification
	if mce.notificationManager != nil {
		mce.notificationManager.Send(
			account.UserID,
			"margin_call",
			map[string]interface{}{
				"accountId":       account.ID,
				"marginLevel":     account.MarginLevel,
				"equity":          account.Equity,
				"margin":          account.Margin,
				"freeMargin":      account.FreeMargin,
				"requiredDeposit": requiredDeposit,
				"threshold":       mce.marginCallLevel,
				"severity":        "error",
				"message":         fmt.Sprintf("MARGIN CALL: Account margin level at %.2f%%. New positions disabled. Deposit $%.2f or close positions.", account.MarginLevel, requiredDeposit),
				"actions":         []string{"Deposit funds", "Close positions", "Reduce leverage"},
			},
		)
	}

	// Store alert in risk engine
	if err := mce.engine.StoreAlert(&RiskAlert{
		ID:        marginCall.ID,
		AccountID: account.ID,
		AlertType: "MARGIN_CALL",
		Severity:  RiskLevelHigh,
		Message:   fmt.Sprintf("Margin call triggered at %.2f%%", account.MarginLevel),
		Data: map[string]interface{}{
			"marginLevel":     account.MarginLevel,
			"requiredDeposit": requiredDeposit,
		},
		CreatedAt: time.Now(),
	}); err != nil {
		log.Printf("[MarginCallEngine] Failed to store alert: %v", err)
	}
}

// handleStopOut auto-liquidates positions
func (mce *MarginCallEngine) handleStopOut(account *AccountInfo) {
	log.Printf("[StopOut] Account #%d: STOP OUT TRIGGERED - Level %.2f%% <= %.2f%% - Auto-liquidating positions",
		account.ID, account.MarginLevel, mce.stopOutLevel)

	// Get all positions for this account
	positions := mce.engine.GetPositions(account.ID)

	if len(positions) == 0 {
		log.Printf("[StopOut] Account #%d: No positions to liquidate", account.ID)
		return
	}

	// Sort positions by unrealized PnL (largest loss first)
	sortedPositions := make([]*PositionInfo, len(positions))
	copy(sortedPositions, positions)
	sort.Slice(sortedPositions, func(i, j int) bool {
		return sortedPositions[i].UnrealizedPnL < sortedPositions[j].UnrealizedPnL
	})

	liquidatedPositions := make([]LiquidatedPosition, 0)
	totalPnL := 0.0

	// Close positions until margin level is restored or all positions are closed
	for _, pos := range sortedPositions {
		// Get current price for liquidation
		bid, ask := mce.engine.GetCurrentPrice(pos.Symbol)
		closePrice := bid
		if pos.Side == "BUY" {
			closePrice = bid // Close BUY at bid
		} else {
			closePrice = ask // Close SELL at ask
		}

		// Calculate PnL for this position
		priceDiff := closePrice - pos.OpenPrice
		if pos.Side == "SELL" {
			priceDiff = -priceDiff
		}
		contractSize := getContractSize(pos.Symbol)
		realizedPnL := priceDiff * pos.Volume * contractSize

		// Close the position (close full volume)
		if err := mce.engine.ClosePosition(pos.ID, closePrice); err != nil {
			log.Printf("[StopOut] Failed to close position %d: %v", pos.ID, err)
			continue
		}

		liquidatedPositions = append(liquidatedPositions, LiquidatedPosition{
			PositionID: pos.ID,
			Symbol:     pos.Symbol,
			Volume:     pos.Volume,
			OpenPrice:  pos.OpenPrice,
			ClosePrice: closePrice,
			PnL:        realizedPnL,
			Slippage:   0, // TODO: Calculate slippage
		})

		totalPnL += realizedPnL

		log.Printf("[StopOut] Closed position %d: %s %.2f lots, PnL: $%.2f",
			pos.ID, pos.Symbol, pos.Volume, realizedPnL)

		// Check if margin level is now acceptable
		// Recalculate margin level after closing this position
		updatedAccount, _ := mce.engine.GetAccount(account.ID)
		if updatedAccount.MarginLevel > mce.marginCallLevel {
			log.Printf("[StopOut] Margin level restored to %.2f%% - stopping liquidation", updatedAccount.MarginLevel)
			break
		}
	}

	// Create liquidation event
	liquidationEvent := LiquidationEvent{
		ID:                fmt.Sprintf("LIQ_%d_%d", account.ID, time.Now().Unix()),
		AccountID:         account.ID,
		TriggerReason:     "STOP_OUT",
		MarginLevelBefore: account.MarginLevel,
		PositionsClosed:   liquidatedPositions,
		TotalPnL:          totalPnL,
		Slippage:          0, // TODO: Aggregate slippage
		ExecutedAt:        time.Now(),
	}

	// Store liquidation event
	mce.engine.StoreLiquidationEvent(liquidationEvent)

	// Send CRITICAL notification
	if mce.notificationManager != nil {
		mce.notificationManager.Send(
			account.UserID,
			"stop_out_triggered",
			map[string]interface{}{
				"accountId":       account.ID,
				"marginLevel":     account.MarginLevel,
				"positionsClosed": len(liquidatedPositions),
				"totalPnL":        totalPnL,
				"threshold":       mce.stopOutLevel,
				"severity":        "critical",
				"message":         fmt.Sprintf("STOP OUT: %d positions auto-closed. Total PnL: $%.2f", len(liquidatedPositions), totalPnL),
			},
		)
	}

	log.Printf("[StopOut] Account #%d: Liquidation complete - %d positions closed, total PnL: $%.2f",
		account.ID, len(liquidatedPositions), totalPnL)
}

// clearMarginStatus clears warnings and margin calls when margin level is restored
func (mce *MarginCallEngine) clearMarginStatus(accountID int64) {
	mce.mu.Lock()
	defer mce.mu.Unlock()

	// Check if there was an active margin call
	if marginCall, exists := mce.activeMarginCalls[accountID]; exists {
		// Mark as resolved
		now := time.Now()
		marginCall.ResolvedAt = &now
		marginCall.Status = "RESOLVED"

		log.Printf("[MarginCallEngine] Account #%d: Margin call RESOLVED - margin level restored",
			accountID)

		// Remove restrictions
		delete(mce.restrictedAccounts, accountID)
		delete(mce.activeMarginCalls, accountID)

		// Send resolution notification
		if mce.notificationManager != nil {
			account, _ := mce.engine.GetAccount(accountID)
			if account != nil {
				mce.notificationManager.Send(
					account.UserID,
					"margin_call_resolved",
					map[string]interface{}{
						"accountId":   accountID,
						"marginLevel": account.MarginLevel,
						"severity":    "info",
						"message":     fmt.Sprintf("Margin call resolved. Current margin level: %.2f%%", account.MarginLevel),
					},
				)
			}
		}
	}

	// Clear warnings
	delete(mce.activeWarnings, accountID)
}

// IsAccountRestricted checks if an account is under margin call restrictions
func (mce *MarginCallEngine) IsAccountRestricted(accountID int64) bool {
	mce.mu.RLock()
	defer mce.mu.RUnlock()
	return mce.restrictedAccounts[accountID]
}

// GetActiveMarginCalls returns all active margin calls
func (mce *MarginCallEngine) GetActiveMarginCalls() []*MarginCall {
	mce.mu.RLock()
	defer mce.mu.RUnlock()

	calls := make([]*MarginCall, 0, len(mce.activeMarginCalls))
	for _, call := range mce.activeMarginCalls {
		calls = append(calls, call)
	}
	return calls
}

// GetMarginCallStatus returns the margin call status for an account
func (mce *MarginCallEngine) GetMarginCallStatus(accountID int64) (*MarginCall, bool) {
	mce.mu.RLock()
	defer mce.mu.RUnlock()

	call, exists := mce.activeMarginCalls[accountID]
	return call, exists
}

// GetThresholds returns the configured margin thresholds
func (mce *MarginCallEngine) GetThresholds() (warning, call, stopOut float64) {
	return mce.marginWarningLevel, mce.marginCallLevel, mce.stopOutLevel
}

// UpdateThresholds updates the margin thresholds (for admin configuration)
func (mce *MarginCallEngine) UpdateThresholds(warning, call, stopOut float64) error {
	if warning <= call || call <= stopOut || stopOut <= 0 {
		return fmt.Errorf("invalid thresholds: warning(%.0f) > call(%.0f) > stopOut(%.0f) > 0", warning, call, stopOut)
	}

	mce.mu.Lock()
	mce.marginWarningLevel = warning
	mce.marginCallLevel = call
	mce.stopOutLevel = stopOut
	mce.mu.Unlock()

	log.Printf("[MarginCallEngine] Thresholds updated: Warning=%.0f%%, Call=%.0f%%, StopOut=%.0f%%",
		warning, call, stopOut)

	return nil
}

// getEnvFloat retrieves a float64 value from environment variable with a default
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getContractSize is defined in engine.go
