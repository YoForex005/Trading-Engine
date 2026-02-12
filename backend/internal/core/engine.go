package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Account represents a trading account
type Account struct {
	ID            int64       `json:"id"`
	AccountNumber string      `json:"accountNumber"`
	UserID        string      `json:"userId"`
	Username      string      `json:"username"` // New: Admin-assigned username
	Password      string      `json:"password"` // New: Admin-assigned password
	Balance       float64     `json:"balance"`
	Equity        float64     `json:"equity"`
	Margin        float64     `json:"margin"`
	FreeMargin    float64     `json:"freeMargin"`
	MarginLevel   float64     `json:"marginLevel"`
	Leverage      float64     `json:"leverage"`
	MarginMode    string      `json:"marginMode"`    // HEDGING or NETTING (deprecated, use PositionMode)
	PositionMode  string      `json:"positionMode"`  // "hedging" or "netting" - how positions are managed
	Currency      string      `json:"currency"`
	Status        string      `json:"status"` // ACTIVE, DISABLED
	IsDemo        bool        `json:"isDemo"`
	CreatedAt     int64       `json:"createdAt"`
	Positions     []*Position `json:"-"` // Internal use only
	Orders        []*Order    `json:"-"`
}

// UpdatePassword updates an account's password
func (e *Engine) UpdatePassword(accountID int64, newPassword string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	account.Password = newPassword
	log.Printf("[B-Book] Password updated for account %s", account.AccountNumber)
	return nil
}

// UpdateAccount updates account configuration
func (e *Engine) UpdateAccount(accountID int64, leverage float64, marginMode string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	if leverage > 0 {
		account.Leverage = leverage
	}
	if marginMode != "" {
		account.MarginMode = marginMode
	}

	log.Printf("[B-Book] Account %s updated: Leverage=%.0f, Mode=%s", account.AccountNumber, account.Leverage, account.MarginMode)
	return nil
}

// UpdatePositionMode updates an account's position mode
func (e *Engine) UpdatePositionMode(accountID int64, positionMode string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	// Validate position mode
	if err := ValidatePositionMode(positionMode); err != nil {
		return err
	}

	// Update position mode
	account.PositionMode = positionMode
	log.Printf("[B-Book] Account %s position mode updated to %s", account.AccountNumber, positionMode)
	return nil
}

// Position represents an open position
type Position struct {
	ID            int64     `json:"id"`
	AccountID     int64     `json:"accountId"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // BUY/SELL
	Volume        float64   `json:"volume"`
	OpenPrice     float64   `json:"openPrice"`
	CurrentPrice  float64   `json:"currentPrice"`
	OpenTime      time.Time `json:"openTime"`
	SL            float64   `json:"sl,omitempty"`
	TP            float64   `json:"tp,omitempty"`
	Swap          float64   `json:"swap"`
	Commission    float64   `json:"commission"`
	UnrealizedPnL float64   `json:"unrealizedPnL"`
	Status        string    `json:"status"`
	ClosePrice    float64   `json:"closePrice,omitempty"`
	CloseTime     time.Time `json:"closeTime,omitempty"`
	CloseReason   string    `json:"closeReason,omitempty"`
}

// Order represents a trading order
type Order struct {
	ID           int64      `json:"id"`
	AccountID    int64      `json:"accountId"`
	Symbol       string     `json:"symbol"`
	Type         string     `json:"type"` // MARKET/LIMIT/STOP/STOP_LIMIT
	Side         string     `json:"side"` // BUY/SELL
	Volume       float64    `json:"volume"`
	Price        float64    `json:"price,omitempty"`
	TriggerPrice float64    `json:"triggerPrice,omitempty"`
	SL           float64    `json:"sl,omitempty"`
	TP           float64    `json:"tp,omitempty"`
	Status       string     `json:"status"`
	FilledPrice  float64    `json:"filledPrice,omitempty"`
	FilledAt     *time.Time `json:"filledAt,omitempty"`
	PositionID   int64      `json:"positionId,omitempty"`
	RejectReason string     `json:"rejectReason,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// Trade represents an execution record
type Trade struct {
	ID          int64     `json:"id"`
	OrderID     int64     `json:"orderId"`
	PositionID  int64     `json:"positionId"`
	AccountID   int64     `json:"accountId"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"`
	Volume      float64   `json:"volume"`
	Price       float64   `json:"price"`
	RealizedPnL float64   `json:"realizedPnL"`
	Commission  float64   `json:"commission"`
	ExecutedAt  time.Time `json:"executedAt"`
}

// AccountSummary contains computed account data
type AccountSummary struct {
	AccountID     int64   `json:"accountId"`
	AccountNumber string  `json:"accountNumber"`
	Currency      string  `json:"currency"`
	Balance       float64 `json:"balance"`
	Equity        float64 `json:"equity"`
	Margin        float64 `json:"margin"`
	FreeMargin    float64 `json:"freeMargin"`
	MarginLevel   float64 `json:"marginLevel"` // Percentage
	UnrealizedPnL float64 `json:"unrealizedPnL"`
	Leverage      float64 `json:"leverage"`
	MarginMode    string  `json:"marginMode"`
	OpenPositions int     `json:"openPositions"`
}

// NotificationSender interface for sending notifications from engine
type NotificationSender interface {
	Send(userID, notifType string, data map[string]interface{}) error
}

// Engine is the B-Book execution engine
type Engine struct {
	mu                  sync.RWMutex
	accounts            map[int64]*Account
	positions           map[int64]*Position
	orders              map[int64]*Order
	trades              []Trade
	symbols             map[string]*SymbolSpec
	nextPositionID      int64
	nextOrderID         int64
	nextTradeID         int64
	priceCallback       func(symbol string) (bid, ask float64, ok bool)
	ledger              *Ledger
	db                  *sql.DB // PostgreSQL connection for persistence
	notificationManager NotificationSender // Notification manager for order events
	nettingEngine       *NettingEngine     // Netting mode position manager
}

// SymbolSpec contains symbol specifications
type SymbolSpec struct {
	Symbol           string  `json:"symbol"`
	ContractSize     float64 `json:"contractSize"`
	PipSize          float64 `json:"pipSize"`
	PipValue         float64 `json:"pipValue"`
	MinVolume        float64 `json:"minVolume"`
	MaxVolume        float64 `json:"maxVolume"`
	VolumeStep       float64 `json:"volumeStep"`
	MarginPercent    float64 `json:"marginPercent"`
	CommissionPerLot float64 `json:"commissionPerLot"`
	SpreadMarkup     float64 `json:"spreadMarkup"`  // Admin-configured spread markup in points
	MinSpread        float64 `json:"minSpread"`     // Minimum allowed spread in points
	MaxSpread        float64 `json:"maxSpread"`     // Maximum allowed spread in points
	SwapLong         float64 `json:"swapLong"`      // Swap rate for long positions (pips per lot per day)
	SwapShort        float64 `json:"swapShort"`     // Swap rate for short positions (pips per lot per day)
	Disabled         bool    `json:"disabled"`      // True if trading/feed is disabled
}

// NewEngine creates a new B-Book engine
func NewEngine() *Engine {
	e := &Engine{
		accounts:       make(map[int64]*Account),
		positions:      make(map[int64]*Position),
		orders:         make(map[int64]*Order),
		trades:         make([]Trade, 0),
		symbols:        make(map[string]*SymbolSpec),
		nextPositionID: 1,
		nextOrderID:    1,
		nextTradeID:    1,
		ledger:         NewLedger(),
	}

	// Initialize netting engine
	e.nettingEngine = NewNettingEngine(e)

	// Wire ledger back to engine for persistence
	e.ledger.SetEngine(e)

	// Load symbols dynamically from tick data directory
	tickDataDir := "./data/ticks"
	if _, err := os.Stat(tickDataDir); err == nil {
		if err := e.LoadSymbolsFromDirectory(tickDataDir); err != nil {
			log.Printf("[B-Book] Warning: Could not load symbols from tick data: %v", err)
			e.registerEssentialSymbols()
		}
	} else {
		e.registerEssentialSymbols()
	}

	log.Println("[B-Book Engine] Initialized")
	return e
}

// registerEssentialSymbols registers core trading symbols dynamically
func (e *Engine) registerEssentialSymbols() {
	essentials := []string{
		"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "NZDUSD", "USDCAD",
		"EURGBP", "EURJPY", "GBPJPY", "AUDJPY", "BTCUSD", "ETHUSD", "XAUUSD",
	}
	for _, symbol := range essentials {
		spec := GenerateSymbolSpec(symbol)
		e.symbols[symbol] = spec
	}
	log.Printf("[B-Book] Registered %d essential symbols", len(essentials))
}

// UpdateSymbol adds or updates a symbol specification
func (e *Engine) UpdateSymbol(spec *SymbolSpec) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.symbols[spec.Symbol] = spec
}

// GetSymbols returns all registered symbols
func (e *Engine) GetSymbols() []*SymbolSpec {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var symbols []*SymbolSpec
	for _, s := range e.symbols {
		symbols = append(symbols, s)
	}
	return symbols
}

// UpdatePrice updates the current price for a symbol and triggers position checks
func (e *Engine) UpdatePrice(symbol string, bid, ask float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 1. Update Symbol Price in Memory (if any specific field uses it)
	// Currently, positions track CurrentPrice individually, so we iterate them.

	for _, pos := range e.positions {
		if pos.Status != "OPEN" || pos.Symbol != symbol {
			continue
		}

		// Update in-memory current price for the position
		var currentPrice float64
		if pos.Side == "BUY" {
			currentPrice = bid // Close at Bid
			pos.CurrentPrice = bid
		} else {
			currentPrice = ask // Close at Ask
			pos.CurrentPrice = ask
		}

		// Recalculate PnL for display
		spec, ok := e.symbols[symbol]
		if ok {
			pos.UnrealizedPnL = e.calculatePnL(pos, currentPrice, pos.Volume, spec)
		}

		// Check Stop Loss
		if pos.SL > 0 {
			if (pos.Side == "BUY" && currentPrice <= pos.SL) || (pos.Side == "SELL" && currentPrice >= pos.SL) {
				log.Printf("[B-Book] SL Triggered for Position #%d (%s) @ %.5f", pos.ID, pos.Symbol, currentPrice)
				// We must unlock to call ClosePosition because it also locks.
				// However, ClosePosition takes a lock. We are currently holding a lock.
				// WE CANNOT CALL ClosePosition DIRECTLY. deadlock.
				// Strategy: Collect triggered positions and close them AFTER unlocking.
				go func(pid int64, vol float64) {
					if _, err := e.ClosePosition(pid, vol); err != nil {
						log.Printf("[B-Book] Failed to execute SL close for #%d: %v", pid, err)
					}
				}(pos.ID, pos.Volume)
				continue
			}
		}

		// Check Take Profit
		if pos.TP > 0 {
			if (pos.Side == "BUY" && currentPrice >= pos.TP) || (pos.Side == "SELL" && currentPrice <= pos.TP) {
				log.Printf("[B-Book] TP Triggered for Position #%d (%s) @ %.5f", pos.ID, pos.Symbol, currentPrice)
				go func(pid int64, vol float64) {
					if _, err := e.ClosePosition(pid, vol); err != nil {
						log.Printf("[B-Book] Failed to execute TP close for #%d: %v", pid, err)
					}
				}(pos.ID, pos.Volume)
			}
		}
	}
}

// SetPriceCallback sets the function to get current market prices
func (e *Engine) SetPriceCallback(fn func(symbol string) (bid, ask float64, ok bool)) {
	e.priceCallback = fn
}

// GetLedger returns the ledger
func (e *Engine) GetLedger() *Ledger {
	return e.ledger
}

// CreateAccount creates a new RTX account
func (e *Engine) CreateAccount(userID, username, password string, isDemo bool) *Account {
	e.mu.Lock()
	defer e.mu.Unlock()

	id := int64(len(e.accounts) + 1)

	// If no username provided, default to Account Number or UserID
	if username == "" {
		username = fmt.Sprintf("RTX-%06d", id)
	}

	// Get default position mode from environment variable
	defaultPositionMode := os.Getenv("DEFAULT_POSITION_MODE")
	if defaultPositionMode == "" {
		defaultPositionMode = string(PositionModeHedging) // Default to hedging if not set
	}

	// Validate position mode, fallback to hedging if invalid
	if err := ValidatePositionMode(defaultPositionMode); err != nil {
		log.Printf("[B-Book] WARNING: Invalid DEFAULT_POSITION_MODE '%s', using hedging", defaultPositionMode)
		defaultPositionMode = string(PositionModeHedging)
	}

	account := &Account{
		ID:            id,
		AccountNumber: fmt.Sprintf("RTX-%06d", id),
		UserID:        userID,
		Username:      username,
		Password:      password,
		Currency:      "USD",
		Balance:       0,
		Leverage:      100,
		MarginMode:    "HEDGING", // Deprecated, use PositionMode
		PositionMode:  defaultPositionMode,
		Status:        "ACTIVE",
		IsDemo:        isDemo,
		CreatedAt:     time.Now().Unix(),
	}

	e.accounts[id] = account

	// Persist to database
	if err := e.persistAccount(account); err != nil {
		log.Printf("[B-Book] ERROR: Failed to persist account to database: %v", err)
	}

	log.Printf("[B-Book] Created account %s (User: %s, Username: %s)", account.AccountNumber, userID, username)
	return account
}

// GetAccount returns an account by ID
func (e *Engine) GetAccount(accountID int64) (*Account, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	acc, ok := e.accounts[accountID]
	return acc, ok
}

// GetAccountByUser returns accounts for a user
func (e *Engine) GetAccountByUser(userID string) []*Account {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var accounts []*Account
	for _, acc := range e.accounts {
		if acc.UserID == userID {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

// GetAllAccounts returns all accounts
func (e *Engine) GetAllAccounts() []*Account {
	e.mu.RLock()
	defer e.mu.RUnlock()

	accounts := make([]*Account, 0, len(e.accounts))
	for _, acc := range e.accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}

// GetAccountSummary computes full account state
func (e *Engine) GetAccountSummary(accountID int64) (*AccountSummary, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return nil, errors.New("account not found")
	}

	// Calculate unrealized P/L and margin
	var unrealizedPnL float64
	var usedMargin float64
	openPositions := 0

	for _, pos := range e.positions {
		if pos.AccountID == accountID && pos.Status == "OPEN" {
			openPositions++
			unrealizedPnL += pos.UnrealizedPnL

			// Calculate margin for position
			spec, ok := e.symbols[pos.Symbol]
			if ok {
				usedMargin += e.calculatePositionMargin(pos, spec, account.Leverage)
			}
		}
	}

	equity := account.Balance + unrealizedPnL
	freeMargin := equity - usedMargin
	marginLevel := 0.0
	if usedMargin > 0 {
		marginLevel = (equity / usedMargin) * 100
	}

	return &AccountSummary{
		AccountID:     account.ID,
		AccountNumber: account.AccountNumber,
		Currency:      account.Currency,
		Balance:       account.Balance,
		Equity:        equity,
		Margin:        usedMargin,
		FreeMargin:    freeMargin,
		MarginLevel:   marginLevel,
		UnrealizedPnL: unrealizedPnL,
		Leverage:      account.Leverage,
		MarginMode:    account.MarginMode,
		OpenPositions: openPositions,
	}, nil
}

// ExecuteMarketOrder executes a market order
func (e *Engine) ExecuteMarketOrder(accountID int64, symbol, side string, volume, sl, tp float64) (*Position, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get account
	account, ok := e.accounts[accountID]
	if !ok {
		return nil, errors.New("account not found")
	}

	if account.Status != "ACTIVE" {
		return nil, errors.New("account is not active")
	}

	// Get symbol specs
	spec, ok := e.symbols[symbol]
	if !ok {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}

	// Validate volume
	if volume < spec.MinVolume || volume > spec.MaxVolume {
		return nil, fmt.Errorf("volume must be between %.2f and %.2f", spec.MinVolume, spec.MaxVolume)
	}

	// Get current price
	if e.priceCallback == nil {
		return nil, errors.New("price feed not available")
	}

	bid, ask, ok := e.priceCallback(symbol)
	if !ok {
		return nil, fmt.Errorf("no price available for %s", symbol)
	}

	// Determine fill price
	var fillPrice float64
	if side == "BUY" {
		fillPrice = ask
	} else if side == "SELL" {
		fillPrice = bid
	} else {
		return nil, errors.New("invalid side: must be BUY or SELL")
	}

	// Calculate required margin
	requiredMargin := e.calculateMargin(symbol, volume, fillPrice, account.Leverage)

	// Check free margin
	summary, _ := e.getAccountSummaryUnlocked(accountID)
	if summary.FreeMargin < requiredMargin {
		return nil, fmt.Errorf("insufficient margin: required %.2f, available %.2f", requiredMargin, summary.FreeMargin)
	}

	// Calculate commission
	commission := spec.CommissionPerLot * volume

	// Create order
	orderID := e.nextOrderID
	e.nextOrderID++

	now := time.Now()
	order := &Order{
		ID:          orderID,
		AccountID:   accountID,
		Symbol:      symbol,
		Type:        "MARKET",
		Side:        side,
		Volume:      volume,
		SL:          sl,
		TP:          tp,
		Status:      "FILLED",
		FilledPrice: fillPrice,
		FilledAt:    &now,
		CreatedAt:   now,
	}
	e.orders[orderID] = order

	// Position handling: branch based on account's position mode
	var position *Position
	var realizedPnL float64

	if account.PositionMode == string(PositionModeNetting) {
		// NETTING MODE: Use netting engine to adjust or create position
		var err error
		position, realizedPnL, err = e.nettingEngine.AdjustPositionUnlocked(accountID, symbol, side, volume, fillPrice, sl, tp, commission)
		if err != nil {
			return nil, fmt.Errorf("netting position adjustment failed: %w", err)
		}

		// If position was fully closed (nil), set order status accordingly
		if position == nil {
			order.Status = "CLOSED"
			log.Printf("[B-Book] NETTING MODE: Position fully closed via opposite order. Realized P&L: %.2f", realizedPnL)
		} else {
			order.PositionID = position.ID
		}
	} else {
		// HEDGING MODE (default): Create new position every time
		positionID := e.nextPositionID
		e.nextPositionID++

		position = &Position{
			ID:           positionID,
			AccountID:    accountID,
			Symbol:       symbol,
			Side:         side,
			Volume:       volume,
			OpenPrice:    fillPrice,
			CurrentPrice: fillPrice,
			OpenTime:     now,
			SL:           sl,
			TP:           tp,
			Commission:   commission,
			Status:       "OPEN",
		}
		e.positions[positionID] = position
		order.PositionID = positionID
	}

	// Derive positionID for logging/notifications (position may be nil in netting close)
	var positionID int64
	if position != nil {
		positionID = position.ID
	}

	// Create trade record
	tradeID := e.nextTradeID
	e.nextTradeID++

	trade := Trade{
		ID:          tradeID,
		OrderID:     orderID,
		PositionID:  0, // Will be set below if position exists
		AccountID:   accountID,
		Symbol:      symbol,
		Side:        side,
		Volume:      volume,
		Price:       fillPrice,
		RealizedPnL: realizedPnL,
		Commission:  commission,
		ExecutedAt:  now,
	}

	// Set position ID if position exists (may be nil in netting mode if fully closed)
	if position != nil {
		trade.PositionID = position.ID
	}

	e.trades = append(e.trades, trade)

	// Deduct commission from balance
	if commission > 0 {
		account.Balance -= commission
		e.ledger.RecordCommission(accountID, -commission, tradeID)
	}

	// Apply realized P&L from netting (if any)
	if realizedPnL != 0 {
		account.Balance += realizedPnL
		log.Printf("[B-Book] Applied realized P&L to balance: %.2f (new balance: %.2f)", realizedPnL, account.Balance)
	}

	// Persist to database in transaction
	if e.db != nil {
		tx, err := e.db.Begin()
		if err == nil {
			// Persist order, position, and trade atomically
			if err := e.persistOrderTx(tx, order); err != nil {
				tx.Rollback()
				log.Printf("[B-Book] ERROR: Failed to persist order: %v", err)
			} else if err := e.persistPositionTx(tx, position); err != nil {
				tx.Rollback()
				log.Printf("[B-Book] ERROR: Failed to persist position: %v", err)
			} else if err := e.persistTradeTx(tx, &trade); err != nil {
				tx.Rollback()
				log.Printf("[B-Book] ERROR: Failed to persist trade: %v", err)
			} else if err := tx.Commit(); err != nil {
				log.Printf("[B-Book] ERROR: Failed to commit transaction: %v", err)
			} else {
				log.Printf("[Persistence] ✓ Position #%d, Order #%d, Trade #%d atomically persisted", positionID, orderID, tradeID)
			}
		}
	}

	// Log execution
	if position != nil {
		log.Printf("[B-Book] EXECUTED: %s %s %.2f lots @ %.5f (Position #%d, Mode=%s)",
			side, symbol, volume, fillPrice, position.ID, account.PositionMode)
	} else {
		log.Printf("[B-Book] EXECUTED: %s %s %.2f lots @ %.5f (Position fully closed, Realized P&L: %.2f)",
			side, symbol, volume, fillPrice, realizedPnL)
	}

	// Send order filled notification
	if e.notificationManager != nil {
		go func() {
			notifData := map[string]interface{}{
				"orderId":    fmt.Sprintf("%d", orderID),
				"positionId": fmt.Sprintf("%d", positionID),
				"symbol":     symbol,
				"orderType":  side,
				"volume":     volume,
				"price":      fmt.Sprintf("%.5f", fillPrice),
			}
			if err := e.notificationManager.Send(account.UserID, "order_filled", notifData); err != nil {
				log.Printf("[Notifications] Failed to send order_filled notification: %v", err)
			}
		}()
	}

	return position, nil
}

// ClosePosition closes a position
func (e *Engine) ClosePosition(positionID int64, closeVolume float64) (*Trade, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	position, ok := e.positions[positionID]
	if !ok {
		return nil, errors.New("position not found")
	}

	if position.Status != "OPEN" {
		return nil, errors.New("position is not open")
	}

	// Get current price
	if e.priceCallback == nil {
		return nil, errors.New("price feed not available")
	}

	bid, ask, ok := e.priceCallback(position.Symbol)
	if !ok {
		return nil, errors.New("no price available")
	}

	// Determine close price (opposite of entry)
	var closePrice float64
	var closeSide string
	if position.Side == "BUY" {
		closePrice = bid
		closeSide = "CLOSE_BUY"
	} else {
		closePrice = ask
		closeSide = "CLOSE_SELL"
	}

	// Determine close volume
	if closeVolume <= 0 || closeVolume >= position.Volume {
		closeVolume = position.Volume
	}

	// Calculate realized P/L
	spec, _ := e.symbols[position.Symbol]
	realizedPnL := e.calculatePnL(position, closePrice, closeVolume, spec)

	// Get account
	account := e.accounts[position.AccountID]

	// Update account balance
	account.Balance += realizedPnL

	// Record in ledger
	tradeID := e.nextTradeID
	e.nextTradeID++

	e.ledger.RecordRealizedPnL(account.ID, realizedPnL, tradeID)

	now := time.Now()

	// Create closing trade
	trade := Trade{
		ID:          tradeID,
		PositionID:  positionID,
		AccountID:   account.ID,
		Symbol:      position.Symbol,
		Side:        closeSide,
		Volume:      closeVolume,
		Price:       closePrice,
		RealizedPnL: realizedPnL,
		ExecutedAt:  now,
	}
	e.trades = append(e.trades, trade)

	// Update position
	closeReason := "MANUAL"
	if closeVolume >= position.Volume {
		position.Status = "CLOSED"
		position.ClosePrice = closePrice
		position.CloseTime = now
		position.CloseReason = closeReason
	} else {
		position.Volume -= closeVolume
	}

	// Persist to database
	if e.db != nil {
		tx, err := e.db.Begin()
		if err == nil {
			// Persist closing trade
			if err := e.persistTradeTx(tx, &trade); err != nil {
				tx.Rollback()
				log.Printf("[B-Book] ERROR: Failed to persist closing trade: %v", err)
			} else if position.Status == "CLOSED" {
				// Position fully closed - update status
				if err := e.closePositionTx(tx, positionID, closePrice, now, closeReason); err != nil {
					tx.Rollback()
					log.Printf("[B-Book] ERROR: Failed to close position: %v", err)
				} else if err := tx.Commit(); err != nil {
					log.Printf("[B-Book] ERROR: Failed to commit transaction: %v", err)
				} else {
					log.Printf("[Persistence] ✓ Position #%d closed, trade #%d atomically persisted", positionID, tradeID)
				}
			} else {
				// Partial close - update volume
				if err := e.updatePositionTx(tx, position); err != nil {
					tx.Rollback()
					log.Printf("[B-Book] ERROR: Failed to update position: %v", err)
				} else if err := tx.Commit(); err != nil {
					log.Printf("[B-Book] ERROR: Failed to commit transaction: %v", err)
				} else {
					log.Printf("[Persistence] ✓ Partial close: Position #%d volume updated, trade #%d persisted", positionID, tradeID)
				}
			}
		}
	}

	log.Printf("[B-Book] CLOSED: %s Position #%d %.2f lots @ %.5f | P/L: %.2f", position.Symbol, positionID, closeVolume, closePrice, realizedPnL)

	// Send position closed notification
	if e.notificationManager != nil {
		go func() {
			notifData := map[string]interface{}{
				"positionId": fmt.Sprintf("%d", positionID),
				"symbol":     position.Symbol,
				"price":      fmt.Sprintf("%.5f", closePrice),
				"pnl":        fmt.Sprintf("%.2f", realizedPnL),
			}
			if err := e.notificationManager.Send(account.UserID, "position_closed", notifData); err != nil {
				log.Printf("[Notifications] Failed to send position_closed notification: %v", err)
			}
		}()
	}

	return &trade, nil
}

// ModifyPosition updates SL/TP for an open position
func (e *Engine) ModifyPosition(positionID int64, sl, tp float64) (*Position, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	position, ok := e.positions[positionID]
	if !ok {
		return nil, errors.New("position not found")
	}

	if position.Status != "OPEN" {
		return nil, errors.New("position is not open")
	}

	position.SL = sl
	position.TP = tp

	// Persist to database
	if err := e.updatePosition(position); err != nil {
		log.Printf("[B-Book] ERROR: Failed to persist position modification: %v", err)
	}

	log.Printf("[B-Book] MODIFIED: Position #%d SL: %.5f TP: %.5f", positionID, sl, tp)

	return position, nil
}

// GetPositions returns open positions for an account
func (e *Engine) GetPositions(accountID int64) []*Position {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var positions []*Position
	for _, pos := range e.positions {
		if pos.AccountID == accountID && pos.Status == "OPEN" {
			positions = append(positions, pos)
		}
	}
	return positions
}

// GetAllPositions returns all open positions
func (e *Engine) GetAllPositions() []*Position {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var positions []*Position
	for _, pos := range e.positions {
		if pos.Status == "OPEN" {
			positions = append(positions, pos)
		}
	}
	return positions
}

// GetOrders returns orders for an account
func (e *Engine) GetOrders(accountID int64, status string) []*Order {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var orders []*Order
	for _, order := range e.orders {
		if order.AccountID == accountID {
			if status == "" || order.Status == status {
				orders = append(orders, order)
			}
		}
	}
	return orders
}

// GetTrades returns trades for an account
func (e *Engine) GetTrades(accountID int64) []Trade {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var trades []Trade
	for _, trade := range e.trades {
		if trade.AccountID == accountID {
			trades = append(trades, trade)
		}
	}
	return trades
}

// UpdatePositionPrices updates current prices and P/L for all positions
func (e *Engine) UpdatePositionPrices() {
	if e.priceCallback == nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, pos := range e.positions {
		if pos.Status != "OPEN" {
			continue
		}

		bid, ask, ok := e.priceCallback(pos.Symbol)
		if !ok {
			continue
		}

		// Update current price
		if pos.Side == "BUY" {
			pos.CurrentPrice = bid
		} else {
			pos.CurrentPrice = ask
		}

		// Calculate unrealized P/L
		spec, ok := e.symbols[pos.Symbol]
		if ok {
			pos.UnrealizedPnL = e.calculatePnL(pos, pos.CurrentPrice, pos.Volume, spec)
		}
	}
}

// calculateMargin calculates required margin for a trade
func (e *Engine) calculateMargin(symbol string, volume, price float64, leverage float64) float64 {
	spec, ok := e.symbols[symbol]
	if !ok {
		return volume * price * 1000 / leverage // Fallback
	}

	// Margin = (Volume * ContractSize * Price) / Leverage
	notional := volume * spec.ContractSize * price
	return notional / leverage
}

// calculatePositionMargin calculates margin for an existing position
func (e *Engine) calculatePositionMargin(pos *Position, spec *SymbolSpec, leverage float64) float64 {
	notional := pos.Volume * spec.ContractSize * pos.OpenPrice
	return notional / leverage
}

// calculatePnL calculates P/L for a position
func (e *Engine) calculatePnL(pos *Position, currentPrice, volume float64, spec *SymbolSpec) float64 {
	if spec == nil {
		return 0
	}

	var priceDiff float64
	if pos.Side == "BUY" {
		priceDiff = currentPrice - pos.OpenPrice
	} else {
		priceDiff = pos.OpenPrice - currentPrice
	}

	// P/L = (PriceDiff / PipSize) * PipValue * Volume
	pips := priceDiff / spec.PipSize
	return pips * spec.PipValue * volume
}

// getAccountSummaryUnlocked is the unlocked version (caller must hold lock)
func (e *Engine) getAccountSummaryUnlocked(accountID int64) (*AccountSummary, error) {
	account, ok := e.accounts[accountID]
	if !ok {
		return nil, errors.New("account not found")
	}

	var unrealizedPnL float64
	var usedMargin float64

	for _, pos := range e.positions {
		if pos.AccountID == accountID && pos.Status == "OPEN" {
			unrealizedPnL += pos.UnrealizedPnL
			spec, ok := e.symbols[pos.Symbol]
			if ok {
				usedMargin += e.calculatePositionMargin(pos, spec, account.Leverage)
			}
		}
	}

	equity := account.Balance + unrealizedPnL
	freeMargin := equity - usedMargin

	return &AccountSummary{
		Balance:    account.Balance,
		Equity:     equity,
		Margin:     usedMargin,
		FreeMargin: freeMargin,
	}, nil
}

// ToggleSymbol enables or disables a symbol
func (e *Engine) ToggleSymbol(symbol string, disabled bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	spec, ok := e.symbols[symbol]
	if !ok {
		return fmt.Errorf("symbol %s not found", symbol)
	}

	spec.Disabled = disabled
	log.Printf("[B-Book] Symbol %s disabled=%v", symbol, disabled)
	return nil
}
