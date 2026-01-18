package risk

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Account represents a trading account with risk metrics
type Account struct {
	ID          int64   `json:"id"`
	UserID      string  `json:"userId"`
	ClientID    string  `json:"clientId"`
	Balance     float64 `json:"balance"`
	Equity      float64 `json:"equity"`
	Margin      float64 `json:"margin"`
	FreeMargin  float64 `json:"freeMargin"`
	MarginLevel float64 `json:"marginLevel"` // Equity/Margin as percentage
	Leverage    int     `json:"leverage"`
	Currency    string  `json:"currency"`
}

// Position represents a trading position
type Position struct {
	ID            int64     `json:"id"`
	AccountID     int64     `json:"accountId"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // BUY or SELL
	Volume        float64   `json:"volume"`
	OpenPrice     float64   `json:"openPrice"`
	CurrentPrice  float64   `json:"currentPrice"`
	UnrealizedPnL float64   `json:"unrealizedPnL"`
	OpenTime      time.Time `json:"openTime"`
	StopLoss      float64   `json:"stopLoss,omitempty"`
	TakeProfit    float64   `json:"takeProfit,omitempty"`
}

// OrderRecord tracks order history for analytics
type OrderRecord struct {
	AccountID  int64     `json:"accountId"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Volume     float64   `json:"volume"`
	Price      float64   `json:"price"`
	OrderTime  time.Time `json:"orderTime"`
}

// Engine handles risk calculations and checks
type Engine struct {
	accounts              map[int64]*Account
	positions             map[int64]*Position
	clientProfiles        map[string]*ClientRiskProfile
	instrumentParams      map[string]*InstrumentRiskParams
	alerts                []*RiskAlert
	dailyPnL              map[int64]float64
	priceCache            map[string]PriceQuote
	nextAccountID         int64
	nextPositionID        int64
	circuitBreakerManager *CircuitBreakerManager

	// Extended tracking for stub method implementation
	peakEquity            map[int64]float64  // accountID -> peak equity
	liquidationEvents     []LiquidationEvent // Historical liquidation events
	orderHistory          []OrderRecord      // Order history for analytics
	creditUsage           map[string]float64 // clientID -> used credit
	correlationMatrix     map[string]map[string]float64 // symbol1 -> symbol2 -> correlation

	mu                    sync.RWMutex
}

// PriceQuote holds bid/ask prices
type PriceQuote struct {
	Bid float64
	Ask float64
}

func NewEngine() *Engine {
	// Initialize engine without hardcoded accounts
	// Accounts should be created via CreateAccount() or loaded from database
	engine := &Engine{
		accounts:         make(map[int64]*Account),
		positions:        make(map[int64]*Position),
		clientProfiles:   make(map[string]*ClientRiskProfile),
		instrumentParams: make(map[string]*InstrumentRiskParams),
		alerts:           make([]*RiskAlert, 0),
		dailyPnL:         make(map[int64]float64),
		priceCache:       make(map[string]PriceQuote),
		nextAccountID:    1,
		nextPositionID:   1,

		// Extended tracking fields
		peakEquity:        make(map[int64]float64),
		liquidationEvents: make([]LiquidationEvent, 0),
		orderHistory:      make([]OrderRecord, 0),
		creditUsage:       make(map[string]float64),
		correlationMatrix: make(map[string]map[string]float64),
	}
	engine.circuitBreakerManager = NewCircuitBreakerManager(engine)
	return engine
}

// PreTradeCheck validates if an order can be placed
func (e *Engine) PreTradeCheck(accountID int64, symbol string, volume float64, price float64) error {
	e.mu.RLock()
	account, ok := e.accounts[accountID]
	e.mu.RUnlock()

	if !ok {
		return errors.New("account not found")
	}

	// Calculate required margin
	contractSize := 100000.0 // Standard lot
	requiredMargin := (volume * contractSize * price) / float64(account.Leverage)

	if requiredMargin > account.FreeMargin {
		return errors.New("insufficient margin")
	}

	return nil
}

// UpdateMargin updates the margin used for an account
func (e *Engine) UpdateMargin(accountID int64, marginDelta float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	account.Margin += marginDelta
	account.FreeMargin = account.Equity - account.Margin

	if account.Margin > 0 {
		account.MarginLevel = (account.Equity / account.Margin) * 100
	} else {
		account.MarginLevel = 0
	}

	return nil
}

// UpdateEquity updates the equity based on floating PnL
func (e *Engine) UpdateEquity(accountID int64, floatingPnL float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return errors.New("account not found")
	}

	account.Equity = account.Balance + floatingPnL
	account.FreeMargin = account.Equity - account.Margin

	if account.Margin > 0 {
		account.MarginLevel = (account.Equity / account.Margin) * 100
	}

	// Check for margin call / stop out
	if account.MarginLevel > 0 && account.MarginLevel < 50 {
		// Would trigger liquidation in production
		return errors.New("margin call triggered")
	}

	return nil
}

// GetAccount returns account info
func (e *Engine) GetAccount(accountID int64) (*Account, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	account, ok := e.accounts[accountID]
	if !ok {
		return nil, errors.New("account not found")
	}

	return account, nil
}

// CreateAccount creates a new trading account
func (e *Engine) CreateAccount(userID string, balance float64, leverage int) *Account {
	e.mu.Lock()
	defer e.mu.Unlock()

	account := &Account{
		ID:          e.nextAccountID,
		UserID:      userID,
		ClientID:    userID,
		Balance:     balance,
		Equity:      balance,
		Margin:      0,
		FreeMargin:  balance,
		MarginLevel: 0,
		Leverage:    leverage,
		Currency:    "USD",
	}

	e.accounts[account.ID] = account
	e.nextAccountID++
	return account
}

// SetClientRiskProfile sets or updates a client's risk profile
func (e *Engine) SetClientRiskProfile(profile *ClientRiskProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clientProfiles[profile.ClientID] = profile
}

// GetClientRiskProfile retrieves a client's risk profile
func (e *Engine) GetClientRiskProfile(clientID string) *ClientRiskProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.clientProfiles[clientID]
}

// SetInstrumentRiskParams sets risk parameters for an instrument
func (e *Engine) SetInstrumentRiskParams(params *InstrumentRiskParams) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.instrumentParams[params.Symbol] = params
}

// GetInstrumentRiskParams retrieves risk parameters for an instrument
func (e *Engine) GetInstrumentRiskParams(symbol string) *InstrumentRiskParams {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.instrumentParams[symbol]
}

// GetPosition retrieves a position by ID
func (e *Engine) GetPosition(positionID int64) *Position {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.positions[positionID]
}

// GetCurrentPrice retrieves the current bid/ask price for a symbol
// Returns (bid, ask)
func (e *Engine) GetCurrentPrice(symbol string) (float64, float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	quote, ok := e.priceCache[symbol]
	if !ok {
		return 0, 0
	}

	return quote.Bid, quote.Ask
}

// UpdatePrice updates the price cache for a symbol
func (e *Engine) UpdatePrice(symbol string, bid, ask float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.priceCache[symbol] = PriceQuote{Bid: bid, Ask: ask}
}

// ClosePosition closes a position with a close price and reason
func (e *Engine) ClosePosition(positionID int64, closePrice float64, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	position, ok := e.positions[positionID]
	if !ok {
		return errors.New("position not found")
	}

	// Calculate realized PnL
	priceDiff := closePrice - position.OpenPrice
	if position.Side == "SELL" {
		priceDiff = -priceDiff
	}
	realizedPnL := priceDiff * position.Volume * getContractSize(position.Symbol)

	// Update daily PnL
	e.dailyPnL[position.AccountID] += realizedPnL

	// Update account balance
	if account, ok := e.accounts[position.AccountID]; ok {
		account.Balance += realizedPnL
		account.Equity = account.Balance
	}

	// Remove position
	delete(e.positions, positionID)

	return nil
}

// getContractSize returns the contract size for a symbol
func getContractSize(symbol string) float64 {
	switch symbol {
	case "XAUUSD":
		return 100.0
	case "XAGUSD":
		return 5000.0
	case "BTCUSD", "ETHUSD", "XRPUSD":
		return 1.0
	default:
		if len(symbol) == 6 {
			return 100000.0 // Forex standard lot
		}
		return 1.0
	}
}

// StoreAlert stores a risk alert
func (e *Engine) StoreAlert(alert *RiskAlert) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alerts = append(e.alerts, alert)
	return nil
}

// GetRecentAlerts retrieves recent alerts
func (e *Engine) GetRecentAlerts(limit int) []RiskAlert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	start := len(e.alerts) - limit
	if start < 0 {
		start = 0
	}

	// Convert []*RiskAlert to []RiskAlert
	result := make([]RiskAlert, 0, len(e.alerts)-start)
	for i := start; i < len(e.alerts); i++ {
		result = append(result, *e.alerts[i])
	}

	return result
}

// GetAllPositions retrieves all positions for an account
func (e *Engine) GetAllPositions(accountID int64) []*Position {
	e.mu.RLock()
	defer e.mu.RUnlock()

	positions := make([]*Position, 0)
	for _, pos := range e.positions {
		if pos.AccountID == accountID {
			positions = append(positions, pos)
		}
	}

	return positions
}

// GetAllAccounts retrieves all accounts
func (e *Engine) GetAllAccounts() []*Account {
	e.mu.RLock()
	defer e.mu.RUnlock()

	accounts := make([]*Account, 0, len(e.accounts))
	for _, acc := range e.accounts {
		accounts = append(accounts, acc)
	}

	return accounts
}

// GetDailyPnL retrieves daily PnL for an account
func (e *Engine) GetDailyPnL(accountID int64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.dailyPnL[accountID]
}

// GetTodayMarginCallCount returns count of margin calls today (stub)
func (e *Engine) GetTodayMarginCallCount() int {
	// TODO: Implement margin call tracking
	return 0
}

// GetTodayLiquidationCount returns count of liquidations today (stub)
func (e *Engine) GetTodayLiquidationCount() int {
	// TODO: Implement liquidation tracking
	return 0
}

// GetAccountByID retrieves an account by ID
func (e *Engine) GetAccountByID(accountID int64) (*Account, error) {
	return e.GetAccount(accountID)
}

// GetVolatility returns volatility for a symbol (stub)
func (e *Engine) GetVolatility(symbol string) float64 {
	// TODO: Implement volatility calculation
	return 0.15 // Default 15%
}

// GetPositionCount returns the number of positions for an account
func (e *Engine) GetPositionCount(accountID int64) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	count := 0
	for _, pos := range e.positions {
		if pos.AccountID == accountID {
			count++
		}
	}
	return count
}

// GetPeakEquity returns the peak equity for an account
func (e *Engine) GetPeakEquity(accountID int64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check if we have a peak equity record
	if peak, exists := e.peakEquity[accountID]; exists {
		return peak
	}

	// If no peak recorded, return current equity
	account, _ := e.GetAccount(accountID)
	if account != nil {
		e.mu.RUnlock()
		e.mu.Lock()
		e.peakEquity[accountID] = account.Equity
		e.mu.Unlock()
		e.mu.RLock()
		return account.Equity
	}
	return 0
}

// UpdatePeakEquity should be called whenever account equity changes
func (e *Engine) UpdatePeakEquity(accountID int64, currentEquity float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	peak, exists := e.peakEquity[accountID]
	if !exists || currentEquity > peak {
		e.peakEquity[accountID] = currentEquity
	}
}

// StoreLiquidationEvent stores a liquidation event
func (e *Engine) StoreLiquidationEvent(event LiquidationEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Store the liquidation event
	e.liquidationEvents = append(e.liquidationEvents, event)

	// Also create a risk alert for immediate notification
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_LIQ_%d", time.Now().Unix()),
		AccountID: event.AccountID,
		AlertType: "LIQUIDATION",
		Severity:  "CRITICAL",
		Message:   fmt.Sprintf("Account liquidated: %d positions closed, PnL: $%.2f", len(event.PositionsClosed), event.TotalPnL),
		CreatedAt: time.Now(),
	}
	e.alerts = append(e.alerts, alert)
}

// IsMarketOpen checks if the market is currently open for a symbol
func (e *Engine) IsMarketOpen(symbol string) bool {
	now := time.Now().UTC()
	weekday := now.Weekday()
	hour := now.Hour()

	// Determine market type from symbol
	switch {
	case isCrypto(symbol):
		// Crypto markets are 24/7
		return true

	case isForex(symbol):
		// Forex markets: Sunday 22:00 UTC - Friday 22:00 UTC
		if weekday == time.Saturday {
			return false
		}
		if weekday == time.Sunday && hour < 22 {
			return false
		}
		if weekday == time.Friday && hour >= 22 {
			return false
		}
		return true

	case isStock(symbol):
		// Stock markets: Monday-Friday, 14:30-21:00 UTC (NYSE hours)
		if weekday == time.Saturday || weekday == time.Sunday {
			return false
		}
		if hour < 14 || hour >= 21 {
			return false
		}
		return true

	case isCommodity(symbol):
		// Commodities: 24/5 (closed weekends)
		if weekday == time.Saturday || weekday == time.Sunday {
			return false
		}
		return true

	default:
		// Unknown instrument type - assume 24/5
		if weekday == time.Saturday || weekday == time.Sunday {
			return false
		}
		return true
	}
}

// Helper functions to determine instrument type
func isCrypto(symbol string) bool {
	cryptos := []string{"BTC", "ETH", "BNB", "SOL", "XRP", "ADA", "DOGE", "LTC"}
	for _, crypto := range cryptos {
		if len(symbol) >= 3 && symbol[:3] == crypto {
			return true
		}
	}
	return false
}

func isForex(symbol string) bool {
	// Forex pairs are 6 characters (EURUSD, GBPUSD, etc.)
	if len(symbol) != 6 {
		return false
	}
	currencies := []string{"USD", "EUR", "GBP", "JPY", "CHF", "AUD", "NZD", "CAD"}
	hasFirst := false
	hasSecond := false
	for _, curr := range currencies {
		if symbol[:3] == curr {
			hasFirst = true
		}
		if symbol[3:] == curr {
			hasSecond = true
		}
	}
	return hasFirst && hasSecond
}

func isStock(symbol string) bool {
	// Stocks typically have ticker symbols like AAPL, MSFT, TSLA
	// This is a simple check - in production, use a lookup table
	return len(symbol) >= 2 && len(symbol) <= 5 && !isCrypto(symbol) && !isForex(symbol)
}

func isCommodity(symbol string) bool {
	commodities := []string{"GOLD", "XAU", "XAG", "OIL", "WTI", "BRENT", "NATGAS"}
	for _, comm := range commodities {
		if len(symbol) >= len(comm) && symbol[:len(comm)] == comm {
			return true
		}
	}
	return false
}

// GetLiquidationEvents returns liquidation events for an account
func (e *Engine) GetLiquidationEvents(accountID int64) []LiquidationEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	events := make([]LiquidationEvent, 0)
	for _, event := range e.liquidationEvents {
		if event.AccountID == accountID {
			events = append(events, event)
		}
	}
	return events
}

// GetCorrelation returns the correlation between two symbols
func (e *Engine) GetCorrelation(symbol1, symbol2 string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Same symbol has perfect correlation
	if symbol1 == symbol2 {
		return 1.0
	}

	// Check if we have a correlation matrix entry
	if corr, exists := e.correlationMatrix[symbol1]; exists {
		if val, ok := corr[symbol2]; ok {
			return val
		}
	}

	// Check reverse direction
	if corr, exists := e.correlationMatrix[symbol2]; exists {
		if val, ok := corr[symbol1]; ok {
			return val
		}
	}

	// If no correlation data, return 0 (independent)
	return 0.0
}

// SetCorrelation sets the correlation between two symbols
func (e *Engine) SetCorrelation(symbol1, symbol2 string, correlation float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Ensure correlation is between -1 and 1
	if correlation < -1 {
		correlation = -1
	} else if correlation > 1 {
		correlation = 1
	}

	// Initialize map if needed
	if e.correlationMatrix[symbol1] == nil {
		e.correlationMatrix[symbol1] = make(map[string]float64)
	}

	// Store correlation (bidirectional)
	e.correlationMatrix[symbol1][symbol2] = correlation
}

// GetSymbolExposure returns the total exposure for a symbol
func (e *Engine) GetSymbolExposure(symbol string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	totalExposure := 0.0
	for _, pos := range e.positions {
		if pos.Symbol == symbol {
			exposure := pos.Volume * pos.CurrentPrice
			if pos.Side == "BUY" {
				totalExposure += exposure
			} else {
				totalExposure -= exposure
			}
		}
	}

	return totalExposure
}

// GetAverageOrderSize returns average order size for account/symbol
func (e *Engine) GetAverageOrderSize(accountID int64, symbol string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var totalVolume float64
	var count int

	for _, order := range e.orderHistory {
		if order.AccountID == accountID && (symbol == "" || order.Symbol == symbol) {
			totalVolume += order.Volume
			count++
		}
	}

	if count == 0 {
		return 0.0
	}
	return totalVolume / float64(count)
}

// RecordOrder records an order in the history for analytics
func (e *Engine) RecordOrder(accountID int64, symbol, side string, volume, price float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	record := OrderRecord{
		AccountID: accountID,
		Symbol:    symbol,
		Side:      side,
		Volume:    volume,
		Price:     price,
		OrderTime: time.Now(),
	}
	e.orderHistory = append(e.orderHistory, record)

	// Keep only last 10000 orders to prevent unlimited growth
	if len(e.orderHistory) > 10000 {
		e.orderHistory = e.orderHistory[len(e.orderHistory)-10000:]
	}
}

// GetTotalExposure returns total exposure for an account
func (e *Engine) GetTotalExposure(accountID int64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	totalExposure := 0.0
	for _, pos := range e.positions {
		if pos.AccountID == accountID {
			exposure := pos.Volume * pos.CurrentPrice
			totalExposure += exposure
		}
	}

	return totalExposure
}

// GetCircuitBreaker returns circuit breaker for a symbol
func (e *Engine) GetCircuitBreaker(symbol string) *CircuitBreaker {
	if e.circuitBreakerManager == nil {
		return nil
	}
	return e.circuitBreakerManager.GetBreaker(symbol)
}

// GetUsedCredit returns used credit for a client
func (e *Engine) GetUsedCredit(clientID string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if credit, exists := e.creditUsage[clientID]; exists {
		return credit
	}
	return 0.0
}

// UpdateCreditUsage updates the credit usage for a client
func (e *Engine) UpdateCreditUsage(clientID string, creditDelta float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	current := e.creditUsage[clientID]
	e.creditUsage[clientID] = current + creditDelta

	// Prevent negative credit usage
	if e.creditUsage[clientID] < 0 {
		e.creditUsage[clientID] = 0
	}
}

// now returns current time (for testing purposes)
func (e *Engine) now() time.Time {
	return time.Now()
}
