package risk

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// EnhancedEngine is the comprehensive risk management engine
// It embeds Engine to maintain compatibility with existing components
type EnhancedEngine struct {
	*Engine // Embed Engine for compatibility
	mu sync.RWMutex

	// Core accounts and positions (override Engine's fields)
	enhancedAccounts  map[int64]*Account
	enhancedPositions map[int64]*Position
	nextID    int64

	// Risk configurations
	clientProfiles      map[string]*ClientRiskProfile
	instrumentParams    map[string]*InstrumentRiskParams
	operationalLimits   *OperationalLimits
	regulatoryLimits    map[string]*RegulatoryLimit
	correlationMatrix   *CorrelationMatrix

	// Risk components
	preTradeValidator       *PreTradeValidator
	marginCalculator        *MarginCalculator
	liquidationEngine       *LiquidationEngine
	circuitBreakerManager   *CircuitBreakerManager
	exposureMonitor         *ExposureMonitor
	adminController         *AdminController

	// Event tracking
	marginCalls       map[int64]*MarginCall
	liquidationEvents []LiquidationEvent
	alerts            []RiskAlert

	// Market data
	priceData       map[string]PriceData
	volatilityCache map[string]float64

	// Performance metrics
	dailyPnL      map[int64]float64
	peakEquity    map[int64]float64
	orderHistory  map[int64][]OrderInfo
	avgOrderSize  map[string]float64 // accountID_symbol -> avg size
}

// PriceData stores bid/ask prices
type PriceData struct {
	Bid       float64
	Ask       float64
	Timestamp time.Time
}

// OrderInfo tracks order history for fat finger detection
type OrderInfo struct {
	Symbol    string
	Volume    float64
	Timestamp time.Time
}

// NewEnhancedEngine creates a new comprehensive risk engine
func NewEnhancedEngine() *EnhancedEngine {
	// Create base engine first
	baseEngine := NewEngine()

	engine := &EnhancedEngine{
		Engine:              baseEngine, // Embed base engine
		enhancedAccounts:    make(map[int64]*Account),
		enhancedPositions:   make(map[int64]*Position),
		nextID:              1,
		clientProfiles:      make(map[string]*ClientRiskProfile),
		instrumentParams:    make(map[string]*InstrumentRiskParams),
		regulatoryLimits:    make(map[string]*RegulatoryLimit),
		marginCalls:         make(map[int64]*MarginCall),
		liquidationEvents:   make([]LiquidationEvent, 0),
		alerts:              make([]RiskAlert, 0),
		priceData:           make(map[string]PriceData),
		volatilityCache:     make(map[string]float64),
		dailyPnL:            make(map[int64]float64),
		peakEquity:          make(map[int64]float64),
		orderHistory:        make(map[int64][]OrderInfo),
		avgOrderSize:        make(map[string]float64),
		operationalLimits: &OperationalLimits{
			MaxOrdersPerSecond:      100,
			MaxOpenOrders:           500,
			MaxPositionsPerAccount:  50,
			MaxSystemExposure:       100000000, // $100M
			MaintenanceMode:         false,
			EmergencyStopEnabled:    false,
			AllowNewAccounts:        true,
			MaxConcurrentConnections: 1000,
		},
	}

	// Initialize components with base engine
	engine.preTradeValidator = NewPreTradeValidator(baseEngine)
	engine.marginCalculator = NewMarginCalculator(baseEngine)
	engine.liquidationEngine = NewLiquidationEngine(baseEngine)
	engine.circuitBreakerManager = NewCircuitBreakerManager(baseEngine)
	engine.exposureMonitor = NewExposureMonitor(baseEngine)
	engine.adminController = NewAdminController(baseEngine)

	// Set default risk profiles
	engine.setDefaultProfiles()

	return engine
}

// Start initializes and starts all risk components
func (e *EnhancedEngine) Start() {
	e.liquidationEngine.Start()
	e.circuitBreakerManager.Start()
	log.Println("[RiskEngine] Enhanced risk engine started")
}

// Stop gracefully stops all risk components
func (e *EnhancedEngine) Stop() {
	e.liquidationEngine.Stop()
	e.circuitBreakerManager.Stop()
	log.Println("[RiskEngine] Enhanced risk engine stopped")
}

// setDefaultProfiles sets default risk profiles and instrument parameters
func (e *EnhancedEngine) setDefaultProfiles() {
	// Default retail client profile
	retailProfile := &ClientRiskProfile{
		ClientID:           "default_retail",
		RiskTier:           "RETAIL",
		MaxLeverage:        30, // ESMA limit for major pairs
		DailyLossLimit:     1000,
		MaxDrawdownPercent: 50,
		MaxPositions:       20,
		MaxExposurePercent: 300,
		MarginCallLevel:    80,
		StopOutLevel:       50,
		AllowHedging:       true,
		AllowScalping:      true,
		RequireStopLoss:    false,
		CreditLimit:        10000,
		InstrumentLimits:   make(map[string]float64),
		MarginMethod:       MarginRetail,
		ApiRateLimit:       10,
		MaxOrderSize:       10,
		FatFingerThreshold: 5,
	}
	e.clientProfiles["default_retail"] = retailProfile

	// Default professional client profile
	professionalProfile := &ClientRiskProfile{
		ClientID:           "default_professional",
		RiskTier:           "PROFESSIONAL",
		MaxLeverage:        100,
		DailyLossLimit:     10000,
		MaxDrawdownPercent: 70,
		MaxPositions:       50,
		MaxExposurePercent: 500,
		MarginCallLevel:    60,
		StopOutLevel:       30,
		AllowHedging:       true,
		AllowScalping:      true,
		RequireStopLoss:    false,
		CreditLimit:        100000,
		InstrumentLimits:   make(map[string]float64),
		MarginMethod:       MarginPortfolio,
		ApiRateLimit:       50,
		MaxOrderSize:       100,
		FatFingerThreshold: 10,
	}
	e.clientProfiles["default_professional"] = professionalProfile

	// Set default instrument parameters for major pairs
	majorPairs := []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD"}
	for _, symbol := range majorPairs {
		params := &InstrumentRiskParams{
			Symbol:             symbol,
			MaxLeverage:        30,  // ESMA retail limit
			MarginRequirement:  3.33, // 30:1 = 3.33%
			MaxPositionSize:    100,
			MaxExposure:        10000000,
			VolatilityLimit:    50, // 50% annualized
			AllowNewPositions:  true,
			RequireStopLoss:    false,
			MaxSlippage:        5,
			TradingSessionOnly: false,
		}
		e.instrumentParams[symbol] = params
	}

	// Gold
	e.instrumentParams["XAUUSD"] = &InstrumentRiskParams{
		Symbol:             "XAUUSD",
		MaxLeverage:        20,
		MarginRequirement:  5,
		MaxPositionSize:    50,
		MaxExposure:        5000000,
		VolatilityLimit:    40,
		AllowNewPositions:  true,
		RequireStopLoss:    false,
		MaxSlippage:        10,
		TradingSessionOnly: false,
	}

	// Crypto
	cryptos := []string{"BTCUSD", "ETHUSD", "XRPUSD"}
	for _, symbol := range cryptos {
		params := &InstrumentRiskParams{
			Symbol:             symbol,
			MaxLeverage:        2,  // ESMA limit
			MarginRequirement:  50, // 2:1 = 50%
			MaxPositionSize:    10,
			MaxExposure:        1000000,
			VolatilityLimit:    100,
			AllowNewPositions:  true,
			RequireStopLoss:    false,
			MaxSlippage:        20,
			TradingSessionOnly: false,
		}
		e.instrumentParams[symbol] = params
	}
}

// ValidateNewOrder performs comprehensive pre-trade validation
func (e *EnhancedEngine) ValidateNewOrder(
	accountID int64,
	symbol string,
	side string,
	volume float64,
	price float64,
	orderType string,
) (*PreTradeCheckResult, error) {
	return e.preTradeValidator.ValidateOrder(accountID, symbol, side, volume, price, orderType)
}

// RecordOrder records an order for history tracking
func (e *EnhancedEngine) RecordOrder(accountID int64, symbol string, volume float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	order := OrderInfo{
		Symbol:    symbol,
		Volume:    volume,
		Timestamp: time.Now(),
	}

	e.orderHistory[accountID] = append(e.orderHistory[accountID], order)

	// Keep only last 100 orders
	if len(e.orderHistory[accountID]) > 100 {
		e.orderHistory[accountID] = e.orderHistory[accountID][1:]
	}

	// Update average order size
	key := fmt.Sprintf("%d_%s", accountID, symbol)
	total := 0.0
	count := 0
	for _, o := range e.orderHistory[accountID] {
		if o.Symbol == symbol {
			total += o.Volume
			count++
		}
	}
	if count > 0 {
		e.avgOrderSize[key] = total / float64(count)
	}
}

// GetAverageOrderSize returns average order size for fat finger detection
func (e *EnhancedEngine) GetAverageOrderSize(accountID int64, symbol string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := fmt.Sprintf("%d_%s", accountID, symbol)
	return e.avgOrderSize[key]
}

// UpdatePrice updates market price data
func (e *EnhancedEngine) UpdatePrice(symbol string, bid, ask float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.priceData[symbol] = PriceData{
		Bid:       bid,
		Ask:       ask,
		Timestamp: time.Now(),
	}

	// Update circuit breaker manager
	midPrice := (bid + ask) / 2
	e.circuitBreakerManager.UpdatePrice(symbol, midPrice)

	// Update positions' current prices
	for _, pos := range e.positions {
		if pos.Symbol == symbol {
			if pos.Side == "BUY" {
				pos.CurrentPrice = bid
			} else {
				pos.CurrentPrice = ask
			}
			// Recalculate unrealized P/L
			pos.UnrealizedPnL = e.calculatePositionPnL(pos)
		}
	}
}

// GetCurrentPrice gets current bid/ask for a symbol
func (e *EnhancedEngine) GetCurrentPrice(symbol string) (bid float64, ask float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if data, ok := e.priceData[symbol]; ok {
		return data.Bid, data.Ask
	}
	return 0, 0
}

// calculatePositionPnL calculates unrealized P/L for a position
func (e *EnhancedEngine) calculatePositionPnL(pos *Position) float64 {
	priceDiff := pos.CurrentPrice - pos.OpenPrice
	if pos.Side == "SELL" {
		priceDiff = -priceDiff
	}

	contractSize := e.getContractSize(pos.Symbol)
	return priceDiff * pos.Volume * contractSize
}

// getContractSize returns contract size for an instrument
func (e *EnhancedEngine) getContractSize(symbol string) float64 {
	switch symbol {
	case "XAUUSD":
		return 100.0
	case "XAGUSD":
		return 5000.0
	case "BTCUSD", "ETHUSD", "XRPUSD":
		return 1.0
	default:
		if len(symbol) == 6 {
			return 100000.0
		}
		return 1.0
	}
}

// Helper methods for engine components
func (e *EnhancedEngine) GetAccountByID(accountID int64) (*Account, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if account, ok := e.accounts[accountID]; ok {
		return account, nil
	}
	return nil, fmt.Errorf("account %d not found", accountID)
}

func (e *EnhancedEngine) GetAllAccounts() []*Account {
	e.mu.RLock()
	defer e.mu.RUnlock()

	accounts := make([]*Account, 0, len(e.accounts))
	for _, acc := range e.accounts {
		accounts = append(accounts, acc)
	}
	return accounts
}

func (e *EnhancedEngine) GetPosition(positionID int64) *Position {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.positions[positionID]
}

func (e *EnhancedEngine) GetAllPositions(accountID int64) []Position {
	e.mu.RLock()
	defer e.mu.RUnlock()

	positions := make([]Position, 0)
	for _, pos := range e.positions {
		if pos.AccountID == accountID {
			positions = append(positions, *pos)
		}
	}
	return positions
}

func (e *EnhancedEngine) GetPositionCount(accountID int64) int {
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

func (e *EnhancedEngine) GetClientRiskProfile(clientID string) *ClientRiskProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Try exact match first
	if profile, ok := e.clientProfiles[clientID]; ok {
		return profile
	}

	// Fall back to default profiles
	return e.clientProfiles["default_retail"]
}

func (e *EnhancedEngine) SetClientRiskProfile(profile *ClientRiskProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clientProfiles[profile.ClientID] = profile
}

func (e *EnhancedEngine) GetInstrumentRiskParams(symbol string) *InstrumentRiskParams {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.instrumentParams[symbol]
}

func (e *EnhancedEngine) SetInstrumentRiskParams(params *InstrumentRiskParams) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.instrumentParams[params.Symbol] = params
}

func (e *EnhancedEngine) GetDailyPnL(accountID int64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.dailyPnL[accountID]
}

func (e *EnhancedEngine) GetPeakEquity(accountID int64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	peak := e.peakEquity[accountID]
	if peak == 0 {
		// Initialize with current equity
		if account, ok := e.accounts[accountID]; ok {
			peak = account.Equity
			e.peakEquity[accountID] = peak
		}
	}
	return peak
}

func (e *EnhancedEngine) GetSymbolExposure(accountID int64, symbol string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	exposure := 0.0
	for _, pos := range e.positions {
		if pos.AccountID == accountID && pos.Symbol == symbol {
			contractSize := e.getContractSize(symbol)
			notional := pos.Volume * contractSize * pos.CurrentPrice

			if pos.Side == "BUY" {
				exposure += notional
			} else {
				exposure -= notional
			}
		}
	}
	return exposure
}

func (e *EnhancedEngine) GetTotalExposure(accountID int64) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	exposure := 0.0
	for _, pos := range e.positions {
		if pos.AccountID == accountID {
			contractSize := e.getContractSize(pos.Symbol)
			notional := pos.Volume * contractSize * pos.CurrentPrice
			exposure += notional
		}
	}
	return exposure
}

func (e *EnhancedEngine) GetCircuitBreaker(symbol string) *CircuitBreaker {
	breakers := e.circuitBreakerManager.GetAllBreakers()
	for _, breaker := range breakers {
		if breaker.Symbol == symbol {
			return &breaker
		}
	}
	return nil
}

func (e *EnhancedEngine) IsMarketOpen(symbol string) bool {
	// Simplified - assume 24/5 for forex, 24/7 for crypto
	cryptos := map[string]bool{"BTCUSD": true, "ETHUSD": true, "XRPUSD": true}
	if cryptos[symbol] {
		return true
	}

	now := time.Now()
	weekday := now.Weekday()
	return weekday != time.Saturday && weekday != time.Sunday
}

func (e *EnhancedEngine) GetUsedCredit(clientID string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	usedCredit := 0.0
	for _, account := range e.accounts {
		if account.ClientID == clientID {
			usedCredit += e.GetTotalExposure(account.ID)
		}
	}
	return usedCredit
}

func (e *EnhancedEngine) GetVolatility(symbol string) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if vol, ok := e.volatilityCache[symbol]; ok {
		return vol
	}
	return 0.15 // Default 15%
}

func (e *EnhancedEngine) GetCorrelation(symbol1, symbol2 string) float64 {
	// Simplified correlation
	if symbol1 == symbol2 {
		return 1.0
	}

	// Check if correlation matrix exists
	if e.correlationMatrix != nil {
		// Find indices
		idx1, idx2 := -1, -1
		for i, sym := range e.correlationMatrix.Symbols {
			if sym == symbol1 {
				idx1 = i
			}
			if sym == symbol2 {
				idx2 = i
			}
		}

		if idx1 >= 0 && idx2 >= 0 {
			return e.correlationMatrix.Correlations[idx1][idx2]
		}
	}

	return 0.3 // Default moderate correlation
}

func (e *EnhancedEngine) ClosePosition(positionID int64, closePrice float64, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	pos, ok := e.positions[positionID]
	if !ok {
		return fmt.Errorf("position %d not found", positionID)
	}

	// Calculate final P/L
	pnl := e.calculatePositionPnL(pos)

	// Update daily P/L
	e.dailyPnL[pos.AccountID] += pnl

	// Update account balance
	if account, ok := e.accounts[pos.AccountID]; ok {
		account.Balance += pnl
		account.Equity = account.Balance

		// Update peak equity
		if account.Equity > e.peakEquity[pos.AccountID] {
			e.peakEquity[pos.AccountID] = account.Equity
		}
	}

	// Remove position
	delete(e.positions, positionID)

	log.Printf("[RiskEngine] Closed position %d: %s %.2f lots, PnL: %.2f (%s)",
		positionID, pos.Symbol, pos.Volume, pnl, reason)

	return nil
}

func (e *EnhancedEngine) StoreAlert(alert *RiskAlert) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alerts = append(e.alerts, *alert)
}

func (e *EnhancedEngine) GetRecentAlerts(limit int) []RiskAlert {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if limit > len(e.alerts) {
		limit = len(e.alerts)
	}

	return e.alerts[len(e.alerts)-limit:]
}

func (e *EnhancedEngine) StoreLiquidationEvent(event *LiquidationEvent) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.liquidationEvents = append(e.liquidationEvents, *event)
}

func (e *EnhancedEngine) GetLiquidationEvents(accountID int64, limit int) []LiquidationEvent {
	e.mu.RLock()
	defer e.mu.RUnlock()

	events := make([]LiquidationEvent, 0)
	for i := len(e.liquidationEvents) - 1; i >= 0 && len(events) < limit; i-- {
		if e.liquidationEvents[i].AccountID == accountID {
			events = append(events, e.liquidationEvents[i])
		}
	}
	return events
}

func (e *EnhancedEngine) GetTodayMarginCallCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	count := 0
	for _, mc := range e.marginCalls {
		if mc.TriggeredAt.Format("2006-01-02") == today {
			count++
		}
	}
	return count
}

func (e *EnhancedEngine) GetTodayLiquidationCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	count := 0
	for _, liq := range e.liquidationEvents {
		if liq.ExecutedAt.Format("2006-01-02") == today {
			count++
		}
	}
	return count
}

func (e *EnhancedEngine) now() time.Time {
	return time.Now()
}
