package bbook

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	decutil "github.com/epic1st/rtx/backend/internal/decimal"
	"github.com/epic1st/rtx/backend/internal/database/repository"
	"github.com/epic1st/rtx/backend/internal/logging"
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
	logging.Default.Info("password updated",
		"account_number", account.AccountNumber,
		"account_id", account.ID,
	)
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

	logging.Default.Info("account updated",
		"account_number", account.AccountNumber,
		"account_id", account.ID,
		"leverage", account.Leverage,
		"margin_mode", account.MarginMode,
	)
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
	mu                     sync.RWMutex
	accounts               map[int64]*Account
	positions              map[int64]*Position
	orders                 map[int64]*Order
	trades                 []Trade
	symbols                map[string]*SymbolSpec
	nextPositionID         int64
	nextOrderID            int64
	nextTradeID            int64
	priceCallback          func(symbol string) (bid, ask float64, ok bool)
	ledger                 *Ledger
	marginStateRepo        *repository.MarginStateRepository
	riskLimitRepo          *repository.RiskLimitRepository
	symbolMarginConfigRepo *repository.SymbolMarginConfigRepository
	dailyStatsRepo         *repository.DailyStatsRepository
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
}

// NewEngine creates a new B-Book engine
func NewEngine() *Engine {
	return NewEngineWithRepos(nil, nil, nil, nil)
}

// NewEngineWithRepos creates a new B-Book engine with optional repository dependencies
func NewEngineWithRepos(
	marginStateRepo *repository.MarginStateRepository,
	riskLimitRepo *repository.RiskLimitRepository,
	symbolMarginConfigRepo *repository.SymbolMarginConfigRepository,
	dailyStatsRepo *repository.DailyStatsRepository,
) *Engine {
	e := &Engine{
		accounts:               make(map[int64]*Account),
		positions:              make(map[int64]*Position),
		orders:                 make(map[int64]*Order),
		trades:                 make([]Trade, 0),
		symbols:                make(map[string]*SymbolSpec),
		nextPositionID:         1,
		nextOrderID:            1,
		nextTradeID:            1,
		ledger:                 NewLedger(),
		marginStateRepo:        marginStateRepo,
		riskLimitRepo:          riskLimitRepo,
		symbolMarginConfigRepo: symbolMarginConfigRepo,
		dailyStatsRepo:         dailyStatsRepo,
	}

	// Initialize default symbols
	e.initDefaultSymbols()

	if logging.Default != nil {
		logging.Default.Info("[B-Book Engine] Initialized")
	}
	return e
}

func (e *Engine) initDefaultSymbols() {
	e.symbols["EURUSD"] = &SymbolSpec{Symbol: "EURUSD", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["GBPUSD"] = &SymbolSpec{Symbol: "GBPUSD", ContractSize: 100000, PipSize: 0.0001, PipValue: 10, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["USDJPY"] = &SymbolSpec{Symbol: "USDJPY", ContractSize: 100000, PipSize: 0.01, PipValue: 9.09, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 1}
	e.symbols["ETHUSD"] = &SymbolSpec{Symbol: "ETHUSD", ContractSize: 1, PipSize: 0.1, PipValue: 0.1, MinVolume: 0.01, MaxVolume: 100, VolumeStep: 0.01, MarginPercent: 5}

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

	account := &Account{
		ID:            id,
		AccountNumber: fmt.Sprintf("RTX-%06d", id),
		UserID:        userID,
		Username:      username,
		Password:      password,
		Currency:      "USD",
		Balance:       0,
		Leverage:      100,
		MarginMode:    "HEDGING",
		Status:        "ACTIVE",
		IsDemo:        isDemo,
	}

	e.accounts[id] = account
	logging.Default.Info("account created",
		"account_number", account.AccountNumber,
		"account_id", account.ID,
		"user_id", userID,
		"username", username,
		"is_demo", isDemo,
	)
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

	// Check daily loss and drawdown limits FIRST (if repositories available)
	if e.dailyStatsRepo != nil && e.riskLimitRepo != nil {
		ctx := context.Background()
		currentBalance := decutil.NewFromFloat64(account.Balance)

		// Check daily loss limit
		if err := CheckDailyLossLimit(
			ctx,
			accountID,
			currentBalance,
			e.dailyStatsRepo,
			e.riskLimitRepo,
		); err != nil {
			return nil, fmt.Errorf("daily loss limit check failed: %w", err)
		}

		// Check drawdown limit
		if err := CheckDrawdownLimit(
			ctx,
			accountID,
			currentBalance,
			e.dailyStatsRepo,
			e.riskLimitRepo,
		); err != nil {
			return nil, fmt.Errorf("drawdown limit check failed: %w", err)
		}
	}

	// Pre-trade risk validation (if repositories available)
	if e.marginStateRepo != nil && e.riskLimitRepo != nil && e.symbolMarginConfigRepo != nil {
		ctx := context.Background()

		// Get current margin state for validation
		marginState, err := e.marginStateRepo.GetByAccountID(ctx, accountID)
		var currentEquity, currentUsedMargin float64
		if err != nil {
			// Initialize default margin state if not exists
			currentEquity = account.Balance
			currentUsedMargin = 0.0
		} else {
			// Parse string values to float64 for validation
			fmt.Sscanf(marginState.Equity, "%f", &currentEquity)
			fmt.Sscanf(marginState.UsedMargin, "%f", &currentUsedMargin)
		}


		// Validate margin requirement
		if err := ValidateMarginRequirement(
			ctx,
			accountID,
			symbol,
			decutil.NewFromFloat64(volume),
			side,
			decutil.NewFromFloat64(fillPrice),
			decutil.NewFromFloat64(currentEquity),
			decutil.NewFromFloat64(currentUsedMargin),
			e.symbolMarginConfigRepo,
			e.riskLimitRepo,
		); err != nil {
			return nil, fmt.Errorf("margin validation failed: %w", err)
		}

		// Validate position limits
		openPositionCount := 0
		for _, pos := range e.positions {
			if pos.AccountID == accountID && pos.Status == "OPEN" {
				openPositionCount++
			}
		}

		if err := ValidatePositionLimits(
			ctx,
			accountID,
			symbol,
			decutil.NewFromFloat64(volume),
			openPositionCount,
			e.riskLimitRepo,
		); err != nil {
			return nil, fmt.Errorf("position limit validation failed: %w", err)
		}

		// Collect account positions for exposure validation
		var accountPositions []*Position
		for _, pos := range e.positions {
			if pos.AccountID == accountID {
				accountPositions = append(accountPositions, pos)
			}
		}

		// Validate symbol exposure (concentration risk)
		if err := ValidateSymbolExposure(
			ctx,
			accountID,
			symbol,
			decutil.NewFromFloat64(volume),
			decutil.NewFromFloat64(fillPrice),
			decutil.NewFromFloat64(currentEquity),
			accountPositions,
			e.symbolMarginConfigRepo,
			e.riskLimitRepo,
		); err != nil {
			return nil, fmt.Errorf("symbol exposure validation failed: %w", err)
		}

		// Validate total account exposure
		if err := ValidateTotalExposure(
			ctx,
			accountID,
			decutil.NewFromFloat64(currentEquity),
			accountPositions,
			e.riskLimitRepo,
		); err != nil {
			return nil, fmt.Errorf("total exposure validation failed: %w", err)
		}
	} else {
		// Fallback to old margin check if repositories not available
		requiredMargin := e.calculateMargin(symbol, volume, fillPrice, account.Leverage)
		summary, _ := e.getAccountSummaryUnlocked(accountID)
		if summary.FreeMargin < requiredMargin {
			return nil, fmt.Errorf("insufficient margin: required %.2f, available %.2f", requiredMargin, summary.FreeMargin)
		}
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

	logging.Default.Info("position opened",
		"position_id", positionID,
		"account_id", accountID,
		"symbol", symbol,
		"side", side,
		"volume", volume,
		"fill_price", fillPrice,
		"commission", commission,
	)

	// Recalculate margin after order execution
	if e.marginStateRepo != nil {
		ctx := context.Background()
		if err := e.UpdateMarginState(ctx, accountID); err != nil {
			logging.Default.Error("failed to update margin state",
				"account_id", accountID,
				"error", err,
			)
		}
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
	if closeVolume >= position.Volume {
		position.Status = "CLOSED"
		position.ClosePrice = closePrice
		position.CloseTime = now
	} else {
		position.Volume -= closeVolume
	}

	logging.Default.Info("position closed",
		"position_id", positionID,
		"account_id", account.ID,
		"symbol", position.Symbol,
		"volume", closeVolume,
		"close_price", closePrice,
		"realized_pnl", realizedPnL,
	)

	// Recalculate margin after position close
	if e.marginStateRepo != nil {
		ctx := context.Background()
		if err := e.UpdateMarginState(ctx, account.ID); err != nil {
			logging.Default.Error("failed to update margin state",
				"account_id", account.ID,
				"error", err,
			)
		}
	}

	// Update daily statistics
	if e.dailyStatsRepo != nil && e.riskLimitRepo != nil {
		ctx := context.Background()
		currentBalance := decutil.NewFromFloat64(account.Balance)
		realizedPLDecimal := decutil.NewFromFloat64(realizedPnL)
		tradeWon := realizedPnL > 0

		if err := UpdateDailyStats(
			ctx,
			account.ID,
			currentBalance,
			realizedPLDecimal,
			decutil.Zero(), // unrealized P&L = 0 after position closed
			true,           // trade closed
			tradeWon,
			e.dailyStatsRepo,
			e.riskLimitRepo,
		); err != nil {
			logging.Default.Error("failed to update daily stats",
				"account_id", account.ID,
				"error", err,
			)
		}
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

	logging.Default.Info("position modified",
		"position_id", positionID,
		"account_id", position.AccountID,
		"symbol", position.Symbol,
		"sl", sl,
		"tp", tp,
	)

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
