package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/database/repository"
	"golang.org/x/crypto/bcrypt"
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
	MarginMode    string      `json:"marginMode"` // HEDGING or NETTING
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

// Engine is the B-Book execution engine
type Engine struct {
	mu             sync.RWMutex

	// Repository layer (database persistence)
	accountRepo  *repository.AccountRepository
	positionRepo *repository.PositionRepository
	orderRepo    *repository.OrderRepository
	tradeRepo    *repository.TradeRepository

	// In-memory caches for hot data (performance optimization)
	accounts       map[int64]*Account
	positions      map[int64]*Position
	orders         map[int64]*Order
	trades         []Trade
	symbols        map[string]*SymbolSpec

	nextPositionID int64
	nextOrderID    int64
	nextTradeID    int64
	priceCallback  func(symbol string) (bid, ask float64, ok bool)
	ledger         *Ledger
}

// SymbolSpec contains symbol specifications
type SymbolSpec struct {
	Symbol           string   `json:"symbol"`
	ContractSize     float64  `json:"contractSize"`
	PipSize          float64  `json:"pipSize"`
	PipValue         float64  `json:"pipValue"`
	MinVolume        float64  `json:"minVolume"`
	MaxVolume        float64  `json:"maxVolume"`
	VolumeStep       float64  `json:"volumeStep"`
	MarginPercent    float64  `json:"marginPercent"`
	CommissionPerLot float64  `json:"commissionPerLot"`
	Disabled         bool     `json:"disabled"` // True if trading/feed is disabled
	SourceLP         string   `json:"sourceLP"` // Preferred LP for this symbol
	AvailableLPs     []string `json:"availableLPs"`
}

// NewEngine creates a new B-Book engine with repository dependencies
func NewEngine(
	accountRepo *repository.AccountRepository,
	positionRepo *repository.PositionRepository,
	orderRepo *repository.OrderRepository,
	tradeRepo *repository.TradeRepository,
) *Engine {
	e := &Engine{
		// Inject repositories
		accountRepo:  accountRepo,
		positionRepo: positionRepo,
		orderRepo:    orderRepo,
		tradeRepo:    tradeRepo,

		// Initialize in-memory caches
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

	// Initialize default symbols
	e.initDefaultSymbols()

	log.Println("[B-Book Engine] Initialized with database repositories")
	return e
}

// LoadAccounts loads all accounts from database into cache on startup
func (e *Engine) LoadAccounts(ctx context.Context) error {
	if e.accountRepo == nil {
		log.Println("[Engine] No account repository, skipping database load")
		return nil
	}

	accounts, err := e.accountRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to load accounts from database: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Convert repository.Account to core.Account and cache
	for _, repoAcc := range accounts {
		acc := &Account{
			ID:            repoAcc.ID,
			AccountNumber: repoAcc.AccountNumber,
			UserID:        repoAcc.UserID,
			Username:      repoAcc.Username,
			Password:      repoAcc.Password,
			Balance:       repoAcc.Balance,
			Equity:        repoAcc.Equity,
			Margin:        repoAcc.Margin,
			FreeMargin:    repoAcc.FreeMargin,
			MarginLevel:   repoAcc.MarginLevel,
			Leverage:      repoAcc.Leverage,
			MarginMode:    repoAcc.MarginMode,
			Currency:      repoAcc.Currency,
			Status:        repoAcc.Status,
			IsDemo:        repoAcc.IsDemo,
			CreatedAt:     repoAcc.CreatedAt.Unix(),
		}
		e.accounts[acc.ID] = acc

		// Update next ID counters
		if acc.ID >= e.nextPositionID {
			e.nextPositionID = acc.ID + 1
		}
	}

	log.Printf("[Engine] Loaded %d accounts from database", len(accounts))
	return nil
}

func (e *Engine) initDefaultSymbols() {
	e.symbols["EURUSD"] = &SymbolSpec{Symbol: "EURUSD", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["GBPUSD"] = &SymbolSpec{Symbol: "GBPUSD", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["USDJPY"] = &SymbolSpec{Symbol: "USDJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}

	// Crypto - Binance
	e.symbols["BTCUSD"] = &SymbolSpec{Symbol: "BTCUSD", ContractSize: 1, PipSize: 0.01, PipValue: 0.01, MinVolume: 0.001, MaxVolume: 10, VolumeStep: 0.001, MarginPercent: 5}
	e.symbols["ETHUSD"] = &SymbolSpec{Symbol: "ETHUSD", ContractSize: 1, PipSize: 0.01, PipValue: 0.01, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 5}
	e.symbols["BNBUSD"] = &SymbolSpec{Symbol: "BNBUSD", ContractSize: 1, PipSize: 0.01, PipValue: 0.01, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 5}
	e.symbols["SOLUSD"] = &SymbolSpec{Symbol: "SOLUSD", ContractSize: 1, PipSize: 0.01, PipValue: 0.01, MinVolume: 0.1, MaxVolume: 1000, VolumeStep: 0.1, MarginPercent: 5}
	e.symbols["XRPUSD"] = &SymbolSpec{Symbol: "XRPUSD", ContractSize: 1, PipSize: 0.0001, PipValue: 0.0001, MinVolume: 10, MaxVolume: 100000, VolumeStep: 10, MarginPercent: 5}

	// Major Pairs
	e.symbols["AUDUSD"] = &SymbolSpec{Symbol: "AUDUSD", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["USDCAD"] = &SymbolSpec{Symbol: "USDCAD", ContractSize: 100000, PipSize: 0.0001, PipValue: 7.5, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["USDCHF"] = &SymbolSpec{Symbol: "USDCHF", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["NZDUSD"] = &SymbolSpec{Symbol: "NZDUSD", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}

	// Cross Pairs
	e.symbols["EURGBP"] = &SymbolSpec{Symbol: "EURGBP", ContractSize: 100000, PipSize: 0.0001, PipValue: 13, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["EURJPY"] = &SymbolSpec{Symbol: "EURJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["GBPJPY"] = &SymbolSpec{Symbol: "GBPJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["AUDJPY"] = &SymbolSpec{Symbol: "AUDJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["CHFJPY"] = &SymbolSpec{Symbol: "CHFJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["CADJPY"] = &SymbolSpec{Symbol: "CADJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
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

// GetSymbol returns a specific symbol spec
func (e *Engine) GetSymbol(symbol string) *SymbolSpec {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if s, exists := e.symbols[symbol]; exists {
		// Return copy to avoid race conditions if caller modifies
		copy := *s
		return &copy
	}
	return nil
}

// SetSymbolSource updates the preferred LP for a symbol
func (e *Engine) SetSymbolSource(symbol, sourceLP string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	spec, ok := e.symbols[symbol]
	if !ok {
		return errors.New("symbol not found")
	}

	spec.SourceLP = sourceLP
	e.symbols[symbol] = spec
	return nil
}

// UpdatePrice updates the current price for a symbol and triggers position checks
func (e *Engine) UpdatePrice(symbol string, bid, ask float64, lpID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	spec, exists := e.symbols[symbol]
	if !exists {
		return
	}

	// Filter by SourceLP if set
	if spec.SourceLP != "" && lpID != "" && spec.SourceLP != lpID {
		return
	}

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

	// Hash password with bcrypt before storing
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password for user %s: %v", username, err)
		return nil
	}

	account := &Account{
		ID:            id,
		AccountNumber: fmt.Sprintf("RTX-%06d", id),
		UserID:        userID,
		Username:      username,
		Password:      string(passwordHash), // Store hash, not plaintext
		Currency:      "USD",
		Balance:       0,
		Leverage:      100,
		MarginMode:    "HEDGING",
		Status:        "ACTIVE",
		IsDemo:        isDemo,
	}

	// Persist to database if repository is available
	if e.accountRepo != nil {
		repoAcc := &repository.Account{
			AccountNumber: account.AccountNumber,
			UserID:        account.UserID,
			Username:      account.Username,
			Password:      account.Password,
			Balance:       account.Balance,
			Equity:        0,
			Margin:        0,
			FreeMargin:    0,
			MarginLevel:   0,
			Leverage:      account.Leverage,
			MarginMode:    account.MarginMode,
			Currency:      account.Currency,
			Status:        account.Status,
			IsDemo:        account.IsDemo,
		}

		ctx := context.Background()
		if err := e.accountRepo.Create(ctx, repoAcc); err != nil {
			log.Printf("[ERROR] Failed to persist account to database: %v", err)
			return nil
		}

		// Update ID from database
		account.ID = repoAcc.ID
		account.CreatedAt = repoAcc.CreatedAt.Unix()
	}

	e.accounts[account.ID] = account
	log.Printf("[Account] Created account %s with bcrypt-hashed password (User: %s, Username: %s)", account.AccountNumber, userID, username)
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

	// Create position
	positionID := e.nextPositionID
	e.nextPositionID++

	position := &Position{
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

	// Create trade record
	tradeID := e.nextTradeID
	e.nextTradeID++

	trade := Trade{
		ID:         tradeID,
		OrderID:    orderID,
		PositionID: positionID,
		AccountID:  accountID,
		Symbol:     symbol,
		Side:       side,
		Volume:     volume,
		Price:      fillPrice,
		Commission: commission,
		ExecutedAt: now,
	}
	e.trades = append(e.trades, trade)

	// Deduct commission from balance
	if commission > 0 {
		account.Balance -= commission
		e.ledger.RecordCommission(accountID, -commission, tradeID)
	}

	log.Printf("[B-Book] EXECUTED: %s %s %.2f lots @ %.5f (Position #%d)", side, symbol, volume, fillPrice, positionID)

	return position, nil
}

// CreatePendingOrderWithOCO creates a pending order with optional OCO linking
func (e *Engine) CreatePendingOrderWithOCO(accountID int64, orderType, symbol string, volume, triggerPrice, sl, tp float64, ocoLinkID *int64) (*Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx := context.Background()

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

	// Validate order type
	validTypes := map[string]bool{
		"BUY_LIMIT":  true,
		"SELL_LIMIT": true,
		"BUY_STOP":   true,
		"SELL_STOP":  true,
	}
	if !validTypes[orderType] {
		return nil, fmt.Errorf("invalid order type: %s (must be BUY_LIMIT, SELL_LIMIT, BUY_STOP, or SELL_STOP)", orderType)
	}

	// Get current price for validation
	if e.priceCallback == nil {
		return nil, errors.New("price feed not available")
	}

	bid, ask, ok := e.priceCallback(symbol)
	if !ok {
		return nil, fmt.Errorf("no price available for %s", symbol)
	}

	// Validate trigger price makes sense for order type
	switch orderType {
	case "BUY_LIMIT":
		if triggerPrice >= ask {
			return nil, fmt.Errorf("BUY_LIMIT trigger price (%.5f) must be below current ask (%.5f)", triggerPrice, ask)
		}
	case "SELL_LIMIT":
		if triggerPrice <= bid {
			return nil, fmt.Errorf("SELL_LIMIT trigger price (%.5f) must be above current bid (%.5f)", triggerPrice, bid)
		}
	case "BUY_STOP":
		if triggerPrice <= ask {
			return nil, fmt.Errorf("BUY_STOP trigger price (%.5f) must be above current ask (%.5f)", triggerPrice, ask)
		}
	case "SELL_STOP":
		if triggerPrice >= bid {
			return nil, fmt.Errorf("SELL_STOP trigger price (%.5f) must be below current bid (%.5f)", triggerPrice, bid)
		}
	}

	// Validate SL/TP if provided
	side := "BUY"
	if orderType == "SELL_LIMIT" || orderType == "SELL_STOP" {
		side = "SELL"
	}

	if sl > 0 {
		if side == "BUY" && sl >= triggerPrice {
			return nil, fmt.Errorf("stop-loss (%.5f) must be below trigger price (%.5f) for BUY orders", sl, triggerPrice)
		}
		if side == "SELL" && sl <= triggerPrice {
			return nil, fmt.Errorf("stop-loss (%.5f) must be above trigger price (%.5f) for SELL orders", sl, triggerPrice)
		}
	}

	if tp > 0 {
		if side == "BUY" && tp <= triggerPrice {
			return nil, fmt.Errorf("take-profit (%.5f) must be above trigger price (%.5f) for BUY orders", tp, triggerPrice)
		}
		if side == "SELL" && tp >= triggerPrice {
			return nil, fmt.Errorf("take-profit (%.5f) must be below trigger price (%.5f) for SELL orders", tp, triggerPrice)
		}
	}

	// Validate OCO link if provided
	if ocoLinkID != nil && *ocoLinkID > 0 {
		if e.orderRepo == nil {
			return nil, errors.New("order repository required for OCO linking")
		}

		// Verify linked order exists and is PENDING
		linkedOrder, err := e.orderRepo.GetByID(ctx, *ocoLinkID)
		if err != nil {
			return nil, fmt.Errorf("OCO linked order #%d not found: %w", *ocoLinkID, err)
		}

		if linkedOrder.Status != "PENDING" {
			return nil, fmt.Errorf("OCO linked order #%d is not PENDING (status: %s)", *ocoLinkID, linkedOrder.Status)
		}
	}

	// Create order (in-memory)
	orderID := e.nextOrderID
	e.nextOrderID++

	now := time.Now()
	order := &Order{
		ID:           orderID,
		AccountID:    accountID,
		Symbol:       symbol,
		Type:         orderType,
		Side:         side,
		Volume:       volume,
		TriggerPrice: triggerPrice,
		SL:           sl,
		TP:           tp,
		Status:       "PENDING",
		CreatedAt:    now,
	}
	e.orders[orderID] = order

	// Persist to database
	if e.orderRepo != nil {
		repoOrder := &repository.Order{
			AccountID:    accountID,
			Symbol:       symbol,
			Type:         orderType,
			Side:         side,
			Volume:       volume,
			TriggerPrice: triggerPrice,
			SL:           sl,
			TP:           tp,
			OCOLinkID:    ocoLinkID,
			Status:       "PENDING",
			CreatedAt:    now,
		}
		if err := e.orderRepo.Create(ctx, repoOrder); err != nil {
			log.Printf("[ERROR] Failed to persist pending order: %v", err)
			return nil, fmt.Errorf("failed to persist order: %w", err)
		}

		// Update in-memory order with database-generated ID
		order.ID = repoOrder.ID
		e.orders[repoOrder.ID] = order
		delete(e.orders, orderID)

		// If OCO link specified, update the linked order's OCOLinkID to point back
		if ocoLinkID != nil && *ocoLinkID > 0 {
			// Create bidirectional OCO link
			newOrderID := repoOrder.ID
			if err := e.orderRepo.UpdateOCOLink(ctx, *ocoLinkID, newOrderID); err != nil {
				log.Printf("[ERROR] Failed to update OCO bidirectional link: %v", err)
				// Don't fail the order creation, just log the error
			} else {
				log.Printf("[B-Book] OCO: Linked orders #%d ↔ #%d", newOrderID, *ocoLinkID)
			}
		}
	}

	logMsg := fmt.Sprintf("[B-Book] PENDING ORDER CREATED: %s %s %.2f lots @ %.5f (Order #%d)", orderType, symbol, volume, triggerPrice, order.ID)
	if ocoLinkID != nil && *ocoLinkID > 0 {
		logMsg += fmt.Sprintf(" OCO→#%d", *ocoLinkID)
	}
	log.Println(logMsg)

	return order, nil
}

// CreatePendingOrder creates a pending order without OCO linking (backward compatibility)
func (e *Engine) CreatePendingOrder(accountID int64, orderType, symbol string, volume, triggerPrice, sl, tp float64) (*Order, error) {
	return e.CreatePendingOrderWithOCO(accountID, orderType, symbol, volume, triggerPrice, sl, tp, nil)
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
	if closeVolume >= position.Volume {
		position.Status = "CLOSED"
		position.ClosePrice = closePrice
		position.CloseTime = now
	} else {
		position.Volume -= closeVolume
	}

	log.Printf("[B-Book] CLOSED: %s Position #%d %.2f lots @ %.5f | P/L: %.2f", position.Symbol, positionID, closeVolume, closePrice, realizedPnL)

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

// ExecuteTriggeredOrder executes an order that has been triggered by price conditions
// This is called by OrderMonitor when a pending order's trigger price is hit
func (e *Engine) ExecuteTriggeredOrder(ctx context.Context, orderID int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Load order from repository
	if e.orderRepo == nil {
		return errors.New("order repository not available")
	}

	order, err := e.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to load order: %w", err)
	}

	// Verify order is still pending
	if order.Status != "PENDING" {
		return fmt.Errorf("order #%d is not pending (status: %s)", orderID, order.Status)
	}

	// Get current market price
	if e.priceCallback == nil {
		return errors.New("price feed not available")
	}

	bid, ask, ok := e.priceCallback(order.Symbol)
	if !ok {
		return fmt.Errorf("no price available for %s", order.Symbol)
	}

	// Determine execution price
	var fillPrice float64
	if order.Side == "BUY" {
		fillPrice = ask
	} else {
		fillPrice = bid
	}

	// If this is a SL/TP order linked to a position, close the position
	if order.ParentPositionID != nil && *order.ParentPositionID > 0 {
		return e.executePositionClose(ctx, order, fillPrice)
	}

	// Otherwise, open a new position (for pending STOP/LIMIT orders)
	return e.executePositionOpen(ctx, order, fillPrice)
}

// CreateTrailingStop creates a trailing stop order for an existing position
func (e *Engine) CreateTrailingStop(ctx context.Context, positionID int64, trailingDelta float64) (*Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get the position
	position, ok := e.positions[positionID]
	if !ok {
		return nil, fmt.Errorf("position #%d not found", positionID)
	}

	if position.Status != "OPEN" {
		return nil, errors.New("position is not open")
	}

	// Get current market price
	if e.priceCallback == nil {
		return nil, errors.New("price feed not available")
	}

	bid, ask, ok := e.priceCallback(position.Symbol)
	if !ok {
		return nil, fmt.Errorf("no price available for %s", position.Symbol)
	}

	// Calculate initial trigger price
	var initialTrigger float64
	if position.Side == "BUY" {
		// For long positions: trigger = current_bid - delta
		initialTrigger = bid - trailingDelta
	} else {
		// For short positions: trigger = current_ask + delta
		initialTrigger = ask + trailingDelta
	}

	// Create order in database
	if e.orderRepo == nil {
		return nil, errors.New("order repository not available")
	}

	now := time.Now()
	parentPosID := position.ID
	repoOrder := &repository.Order{
		AccountID:        position.AccountID,
		Symbol:           position.Symbol,
		Type:             "TRAILING_STOP",
		Side:             position.Side, // Same side as position for closing
		Volume:           position.Volume,
		TriggerPrice:     initialTrigger,
		TrailingDelta:    &trailingDelta,
		ParentPositionID: &parentPosID,
		Status:           "PENDING",
		CreatedAt:        now,
	}

	if err := e.orderRepo.Create(ctx, repoOrder); err != nil {
		return nil, fmt.Errorf("failed to create trailing stop order: %w", err)
	}

	// Create in-memory order
	orderID := repoOrder.ID
	order := &Order{
		ID:           orderID,
		AccountID:    position.AccountID,
		Symbol:       position.Symbol,
		Type:         "TRAILING_STOP",
		Side:         position.Side,
		Volume:       position.Volume,
		TriggerPrice: initialTrigger,
		Status:       "PENDING",
		CreatedAt:    now,
	}

	e.orders[orderID] = order

	log.Printf("[B-Book] Created trailing stop #%d for position #%d: trigger=%.5f delta=%.5f",
		orderID, positionID, initialTrigger, trailingDelta)

	return order, nil
}

// executePositionClose handles SL/TP execution by closing the parent position
func (e *Engine) executePositionClose(ctx context.Context, order *repository.Order, closePrice float64) error {
	parentPosID := *order.ParentPositionID

	// Get position from in-memory cache
	position, ok := e.positions[parentPosID]
	if !ok {
		return fmt.Errorf("parent position #%d not found", parentPosID)
	}

	if position.Status != "OPEN" {
		return fmt.Errorf("parent position #%d is not open", parentPosID)
	}

	// Calculate realized P/L
	spec, _ := e.symbols[position.Symbol]
	realizedPnL := e.calculatePnL(position, closePrice, position.Volume, spec)

	// Get account
	account := e.accounts[position.AccountID]
	if account == nil {
		return fmt.Errorf("account #%d not found", position.AccountID)
	}

	// Update account balance
	account.Balance += realizedPnL

	// Record in ledger
	tradeID := e.nextTradeID
	e.nextTradeID++
	e.ledger.RecordRealizedPnL(account.ID, realizedPnL, tradeID)

	// Create closing trade record
	now := time.Now()
	closeSide := "CLOSE_BUY"
	if position.Side == "SELL" {
		closeSide = "CLOSE_SELL"
	}

	trade := Trade{
		ID:          tradeID,
		OrderID:     order.ID,
		PositionID:  parentPosID,
		AccountID:   account.ID,
		Symbol:      position.Symbol,
		Side:        closeSide,
		Volume:      position.Volume,
		Price:       closePrice,
		RealizedPnL: realizedPnL,
		ExecutedAt:  now,
	}
	e.trades = append(e.trades, trade)

	// Persist trade to database
	if e.tradeRepo != nil {
		orderIDPtr := &order.ID
		repoTrade := &repository.Trade{
			OrderID:     orderIDPtr,
			PositionID:  parentPosID,
			AccountID:   account.ID,
			Symbol:      position.Symbol,
			Side:        closeSide,
			Volume:      position.Volume,
			Price:       closePrice,
			RealizedPnL: realizedPnL,
			Commission:  0,
			ExecutedAt:  now,
		}
		if err := e.tradeRepo.Create(ctx, repoTrade); err != nil {
			log.Printf("[ERROR] Failed to persist trade to database: %v", err)
		}
	}

	// Close position
	position.Status = "CLOSED"
	position.ClosePrice = closePrice
	position.CloseTime = now
	position.CloseReason = order.Type // STOP_LOSS or TAKE_PROFIT

	// Persist position update (use Close method from repository)
	if e.positionRepo != nil {
		if err := e.positionRepo.Close(ctx, position.ID, closePrice, order.Type); err != nil {
			log.Printf("[ERROR] Failed to persist position close: %v", err)
		}
	}

	// Mark order as filled
	filledAt := now
	fillPrice := closePrice
	order.Status = "FILLED"
	order.FilledPrice = fillPrice
	order.FilledAt = &filledAt

	if e.orderRepo != nil {
		if err := e.orderRepo.UpdateStatus(ctx, order.ID, "FILLED", &fillPrice); err != nil {
			log.Printf("[ERROR] Failed to update order status: %v", err)
		}
	}

	// Update in-memory order cache
	if memOrder, ok := e.orders[order.ID]; ok {
		memOrder.Status = "FILLED"
		memOrder.FilledPrice = fillPrice
		memOrder.FilledAt = &filledAt
	}

	// Persist account balance update (use UpdateBalance method from repository)
	if e.accountRepo != nil {
		// Recalculate account summary for persistence
		accountSummary, _ := e.getAccountSummaryUnlocked(account.ID)
		if err := e.accountRepo.UpdateBalance(ctx, account.ID, account.Balance,
			accountSummary.Equity, accountSummary.Margin, accountSummary.FreeMargin,
			accountSummary.MarginLevel); err != nil {
			log.Printf("[ERROR] Failed to update account balance: %v", err)
		}
	}

	log.Printf("[B-Book] %s TRIGGERED: Closed Position #%d @ %.5f | P/L: %.2f",
		order.Type, parentPosID, closePrice, realizedPnL)

	// Cancel OCO linked order if exists
	if err := e.CancelOCOLinkedOrder(ctx, order.ID); err != nil {
		log.Printf("[Engine] Failed to cancel OCO linked order for #%d: %v", order.ID, err)
	}

	return nil
}

// executePositionOpen handles pending order execution by opening a new position
func (e *Engine) executePositionOpen(ctx context.Context, order *repository.Order, fillPrice float64) error {
	// Get account
	account, ok := e.accounts[order.AccountID]
	if !ok {
		return fmt.Errorf("account #%d not found", order.AccountID)
	}

	// Get symbol specs
	spec, ok := e.symbols[order.Symbol]
	if !ok {
		return fmt.Errorf("symbol %s not found", order.Symbol)
	}

	// Calculate required margin
	requiredMargin := e.calculateMargin(order.Symbol, order.Volume, fillPrice, account.Leverage)

	// Check free margin
	summary, _ := e.getAccountSummaryUnlocked(order.AccountID)
	if summary.FreeMargin < requiredMargin {
		// Reject order due to insufficient margin
		order.Status = "REJECTED"
		order.RejectReason = fmt.Sprintf("Insufficient margin: required %.2f, available %.2f",
			requiredMargin, summary.FreeMargin)

		if e.orderRepo != nil {
			e.orderRepo.UpdateStatus(ctx, order.ID, "REJECTED", nil)
		}

		log.Printf("[B-Book] Order #%d REJECTED: %s", order.ID, order.RejectReason)
		return errors.New(order.RejectReason)
	}

	// Calculate commission
	commission := spec.CommissionPerLot * order.Volume

	// Create position
	positionID := e.nextPositionID
	e.nextPositionID++

	now := time.Now()
	position := &Position{
		ID:           positionID,
		AccountID:    order.AccountID,
		Symbol:       order.Symbol,
		Side:         order.Side,
		Volume:       order.Volume,
		OpenPrice:    fillPrice,
		CurrentPrice: fillPrice,
		OpenTime:     now,
		SL:           order.SL,
		TP:           order.TP,
		Commission:   commission,
		Status:       "OPEN",
	}
	e.positions[positionID] = position

	// Update order
	order.Status = "FILLED"
	order.FilledPrice = fillPrice
	filledAt := now
	order.FilledAt = &filledAt
	order.PositionID = positionID

	// Persist to database
	if e.orderRepo != nil {
		e.orderRepo.UpdateStatus(ctx, order.ID, "FILLED", &fillPrice)
	}

	if e.positionRepo != nil {
		repoPos := &repository.Position{
			AccountID:     position.AccountID,
			Symbol:        position.Symbol,
			Side:          position.Side,
			Volume:        position.Volume,
			OpenPrice:     position.OpenPrice,
			CurrentPrice:  position.CurrentPrice,
			OpenTime:      position.OpenTime,
			SL:            position.SL,
			TP:            position.TP,
			Swap:          position.Swap,
			Commission:    position.Commission,
			UnrealizedPnL: 0,
			Status:        "OPEN",
		}
		if err := e.positionRepo.Create(ctx, repoPos); err != nil {
			log.Printf("[ERROR] Failed to persist position: %v", err)
		} else {
			position.ID = repoPos.ID
		}
	}

	// Deduct commission from balance
	if commission > 0 {
		account.Balance -= commission
		tradeID := e.nextTradeID
		e.nextTradeID++
		e.ledger.RecordCommission(order.AccountID, -commission, tradeID)
	}

	log.Printf("[B-Book] %s TRIGGERED: Opened Position #%d (%s %s %.2f lots @ %.5f)",
		order.Type, positionID, order.Side, order.Symbol, order.Volume, fillPrice)

	// Cancel OCO linked order if exists
	if err := e.CancelOCOLinkedOrder(ctx, order.ID); err != nil {
		log.Printf("[Engine] Failed to cancel OCO linked order for #%d: %v", order.ID, err)
	}

	return nil
}

// ModifyOrder modifies a pending order's parameters
func (e *Engine) ModifyOrder(orderID int64, triggerPrice, sl, tp, volume *float64, expiryTime *time.Time) (*Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx := context.Background()

	// Load order from repository
	if e.orderRepo == nil {
		return nil, errors.New("order repository not available")
	}

	order, err := e.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Validate order is PENDING
	if order.Status != "PENDING" {
		return nil, fmt.Errorf("cannot modify order in status: %s", order.Status)
	}

	// Update fields if provided
	updated := false

	if triggerPrice != nil && *triggerPrice > 0 {
		// Validate trigger price makes sense for order type
		if e.priceCallback != nil {
			bid, ask, ok := e.priceCallback(order.Symbol)
			if ok {
				switch order.Type {
				case "BUY_LIMIT":
					if *triggerPrice >= ask {
						return nil, fmt.Errorf("BUY_LIMIT trigger (%.5f) must be below ask (%.5f)", *triggerPrice, ask)
					}
				case "SELL_LIMIT":
					if *triggerPrice <= bid {
						return nil, fmt.Errorf("SELL_LIMIT trigger (%.5f) must be above bid (%.5f)", *triggerPrice, bid)
					}
				case "BUY_STOP":
					if *triggerPrice <= ask {
						return nil, fmt.Errorf("BUY_STOP trigger (%.5f) must be above ask (%.5f)", *triggerPrice, ask)
					}
				case "SELL_STOP":
					if *triggerPrice >= bid {
						return nil, fmt.Errorf("SELL_STOP trigger (%.5f) must be below bid (%.5f)", *triggerPrice, bid)
					}
				}
			}
		}
		order.TriggerPrice = *triggerPrice
		updated = true
	}

	if sl != nil {
		order.SL = *sl
		updated = true
	}

	if tp != nil {
		order.TP = *tp
		updated = true
	}

	if volume != nil && *volume > 0 {
		// Get symbol specs to validate volume
		if spec, ok := e.symbols[order.Symbol]; ok {
			if *volume < spec.MinVolume || *volume > spec.MaxVolume {
				return nil, fmt.Errorf("volume must be between %.2f and %.2f", spec.MinVolume, spec.MaxVolume)
			}
		}
		order.Volume = *volume
		updated = true
	}

	if expiryTime != nil {
		order.ExpiryTime = expiryTime
		updated = true
	}

	if !updated {
		return nil, errors.New("no modifications provided")
	}

	// Save changes to database
	if err := e.orderRepo.UpdateModifiable(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	// Update in-memory cache
	if memOrder, ok := e.orders[orderID]; ok {
		if triggerPrice != nil {
			memOrder.TriggerPrice = *triggerPrice
		}
		if sl != nil {
			memOrder.SL = *sl
		}
		if tp != nil {
			memOrder.TP = *tp
		}
		if volume != nil {
			memOrder.Volume = *volume
		}
	}

	log.Printf("[Engine] Modified order #%d", orderID)

	// Convert repository.Order to core.Order for return
	coreOrder := &Order{
		ID:           order.ID,
		AccountID:    order.AccountID,
		Symbol:       order.Symbol,
		Type:         order.Type,
		Side:         order.Side,
		Volume:       order.Volume,
		TriggerPrice: order.TriggerPrice,
		SL:           order.SL,
		TP:           order.TP,
		Status:       order.Status,
		CreatedAt:    order.CreatedAt,
	}

	return coreOrder, nil
}

// CancelOrder cancels a pending order
func (e *Engine) CancelOrder(orderID int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	ctx := context.Background()

	// Load order from repository
	if e.orderRepo == nil {
		return errors.New("order repository not available")
	}

	order, err := e.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}

	// Validate order is PENDING
	if order.Status != "PENDING" {
		return fmt.Errorf("cannot cancel order in status: %s", order.Status)
	}

	// Update status to CANCELLED
	order.Status = "CANCELLED"
	order.RejectReason = "User cancelled"

	if err := e.orderRepo.UpdateStatus(ctx, order.ID, "CANCELLED", nil); err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Update in-memory cache
	if memOrder, ok := e.orders[orderID]; ok {
		memOrder.Status = "CANCELLED"
		memOrder.RejectReason = "User cancelled"
	}

	log.Printf("[Engine] Cancelled order #%d (user requested)", orderID)

	return nil
}

// CancelOCOLinkedOrder cancels the linked order when one OCO order fills
func (e *Engine) CancelOCOLinkedOrder(ctx context.Context, orderID int64) error {
	if e.orderRepo == nil {
		return errors.New("order repository not available")
	}

	// Load the order that just filled
	order, err := e.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to load order: %w", err)
	}

	// Check if this order has an OCO link
	if order.OCOLinkID == nil || *order.OCOLinkID == 0 {
		// No OCO link, nothing to cancel
		return nil
	}

	linkedOrderID := *order.OCOLinkID

	// Load the linked order
	linkedOrder, err := e.orderRepo.GetByID(ctx, linkedOrderID)
	if err != nil {
		log.Printf("[Engine] Failed to load OCO linked order #%d: %v", linkedOrderID, err)
		return fmt.Errorf("failed to load linked order: %w", err)
	}

	// Only cancel if the linked order is still PENDING
	if linkedOrder.Status != "PENDING" {
		log.Printf("[Engine] OCO linked order #%d already %s, skipping cancellation", linkedOrderID, linkedOrder.Status)
		return nil
	}

	// Update linked order status to CANCELLED
	linkedOrder.Status = "CANCELLED"
	linkedOrder.RejectReason = "OCO order filled"

	if err := e.orderRepo.UpdateStatus(ctx, linkedOrder.ID, "CANCELLED", nil); err != nil {
		log.Printf("[Engine] Failed to cancel OCO linked order #%d: %v", linkedOrderID, err)
		return fmt.Errorf("failed to cancel linked order: %w", err)
	}

	// Update in-memory cache if exists
	if memOrder, ok := e.orders[linkedOrder.ID]; ok {
		memOrder.Status = "CANCELLED"
		memOrder.RejectReason = "OCO order filled"
	}

	log.Printf("[Engine] OCO: Cancelled order #%d (linked to filled order #%d)", linkedOrderID, orderID)

	// TODO: Broadcast WebSocket update for cancelled order

	return nil
}
