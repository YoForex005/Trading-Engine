package features

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Strategy Automation Framework
// - Strategy definition and execution
// - Backtesting framework
// - Paper trading mode
// - Strategy performance analytics
// - REST API for algo trading
// - WebSocket for low-latency execution

// StrategyType defines the type of strategy
type StrategyType string

const (
	StrategyTypeIndicatorBased StrategyType = "INDICATOR"
	StrategyTypePatternBased   StrategyType = "PATTERN"
	StrategyTypeMLBased        StrategyType = "ML"
	StrategyTypeArbitrage      StrategyType = "ARBITRAGE"
	StrategyTypeCustom         StrategyType = "CUSTOM"
)

// StrategyMode defines execution mode
type StrategyMode string

const (
	ModeLive      StrategyMode = "LIVE"
	ModePaper     StrategyMode = "PAPER"
	ModeBacktest  StrategyMode = "BACKTEST"
)

// Strategy represents a trading strategy
type Strategy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        StrategyType           `json:"type"`
	Mode        StrategyMode           `json:"mode"`
	Symbols     []string               `json:"symbols"`
	Timeframe   string                 `json:"timeframe"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`

	// Risk Management
	MaxPositions     int     `json:"maxPositions"`
	MaxRiskPerTrade  float64 `json:"maxRiskPerTrade"`  // %
	MaxDrawdown      float64 `json:"maxDrawdown"`      // %
	MaxDailyLoss     float64 `json:"maxDailyLoss"`     // $
	DefaultStopLoss  float64 `json:"defaultStopLoss"`  // pips
	DefaultTakeProfit float64 `json:"defaultTakeProfit"` // pips

	// Performance
	Stats *StrategyStats `json:"stats,omitempty"`
}

// StrategyStats tracks strategy performance
type StrategyStats struct {
	TotalTrades      int       `json:"totalTrades"`
	WinningTrades    int       `json:"winningTrades"`
	LosingTrades     int       `json:"losingTrades"`
	WinRate          float64   `json:"winRate"`
	TotalProfit      float64   `json:"totalProfit"`
	TotalLoss        float64   `json:"totalLoss"`
	NetProfit        float64   `json:"netProfit"`
	ProfitFactor     float64   `json:"profitFactor"`
	AverageWin       float64   `json:"averageWin"`
	AverageLoss      float64   `json:"averageLoss"`
	LargestWin       float64   `json:"largestWin"`
	LargestLoss      float64   `json:"largestLoss"`
	MaxDrawdown      float64   `json:"maxDrawdown"`
	SharpeRatio      float64   `json:"sharpeRatio"`
	SortinoRatio     float64   `json:"sortinoRatio"`
	CalmarRatio      float64   `json:"calmarRatio"`
	RecoveryFactor   float64   `json:"recoveryFactor"`
	ExpectancyPips   float64   `json:"expectancyPips"`
	MAE              float64   `json:"mae"` // Maximum Adverse Excursion
	MFE              float64   `json:"mfe"` // Maximum Favorable Excursion
	ConsecutiveWins  int       `json:"consecutiveWins"`
	ConsecutiveLosses int       `json:"consecutiveLosses"`
	LastUpdated      time.Time `json:"lastUpdated"`
}

// StrategySignal represents a trading signal
type StrategySignal struct {
	ID         string    `json:"id"`
	StrategyID string    `json:"strategyId"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"` // BUY or SELL
	Confidence float64   `json:"confidence"` // 0-100
	Price      float64   `json:"price"`
	StopLoss   float64   `json:"stopLoss"`
	TakeProfit float64   `json:"takeProfit"`
	Volume     float64   `json:"volume"`
	Reasoning  string    `json:"reasoning"`
	Timestamp  time.Time `json:"timestamp"`
	Executed   bool      `json:"executed"`
	OrderID    string    `json:"orderId,omitempty"`
}

// BacktestResult represents backtest results
type BacktestResult struct {
	StrategyID    string          `json:"strategyId"`
	StartDate     time.Time       `json:"startDate"`
	EndDate       time.Time       `json:"endDate"`
	InitialBalance float64         `json:"initialBalance"`
	FinalBalance   float64         `json:"finalBalance"`
	Stats          *StrategyStats  `json:"stats"`
	Trades         []BacktestTrade `json:"trades"`
	EquityCurve    []EquityPoint   `json:"equityCurve"`
	DrawdownCurve  []DrawdownPoint `json:"drawdownCurve"`
	CompletedAt    time.Time       `json:"completedAt"`
}

// BacktestTrade represents a trade in backtest
type BacktestTrade struct {
	EntryTime   time.Time `json:"entryTime"`
	ExitTime    time.Time `json:"exitTime"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"`
	Volume      float64   `json:"volume"`
	EntryPrice  float64   `json:"entryPrice"`
	ExitPrice   float64   `json:"exitPrice"`
	Profit      float64   `json:"profit"`
	ProfitPips  float64   `json:"profitPips"`
	MAE         float64   `json:"mae"`
	MFE         float64   `json:"mfe"`
	Commission  float64   `json:"commission"`
	Swap        float64   `json:"swap"`
	ExitReason  string    `json:"exitReason"` // SL, TP, Signal, Manual
}

// EquityPoint for equity curve
type EquityPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Balance   float64   `json:"balance"`
	Equity    float64   `json:"equity"`
}

// DrawdownPoint for drawdown curve
type DrawdownPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Drawdown       float64   `json:"drawdown"`
	DrawdownPercent float64   `json:"drawdownPercent"`
}

// StrategyService manages trading strategies
type StrategyService struct {
	mu         sync.RWMutex
	strategies map[string]*Strategy
	signals    map[string][]*StrategySignal // strategyID -> signals
	backtests  map[string]*BacktestResult

	// Callbacks
	priceCallback   func(symbol string) (bid, ask float64, ok bool)
	executeCallback func(signal *StrategySignal) (string, error)
	dataCallback    func(symbol, timeframe string, start, end time.Time) ([]OHLCBar, error)

	// Services
	indicatorService *IndicatorService
}

// NewStrategyService creates the strategy service
func NewStrategyService(indicatorService *IndicatorService) *StrategyService {
	svc := &StrategyService{
		strategies:       make(map[string]*Strategy),
		signals:          make(map[string][]*StrategySignal),
		backtests:        make(map[string]*BacktestResult),
		indicatorService: indicatorService,
	}

	go svc.processLoop()

	log.Println("[StrategyService] Initialized with backtesting and paper trading support")
	return svc
}

// SetCallbacks configures callbacks
func (s *StrategyService) SetCallbacks(
	priceCallback func(symbol string) (bid, ask float64, ok bool),
	executeCallback func(signal *StrategySignal) (string, error),
	dataCallback func(symbol, timeframe string, start, end time.Time) ([]OHLCBar, error),
) {
	s.priceCallback = priceCallback
	s.executeCallback = executeCallback
	s.dataCallback = dataCallback
}

// CreateStrategy creates a new strategy
func (s *StrategyService) CreateStrategy(strategy *Strategy) (*Strategy, error) {
	if strategy.Name == "" {
		return nil, errors.New("strategy name required")
	}

	strategy.ID = uuid.New().String()
	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()
	strategy.Stats = &StrategyStats{
		LastUpdated: time.Now(),
	}

	s.mu.Lock()
	s.strategies[strategy.ID] = strategy
	s.signals[strategy.ID] = make([]*StrategySignal, 0)
	s.mu.Unlock()

	log.Printf("[Strategy] Created: %s (%s mode)", strategy.Name, strategy.Mode)
	return strategy, nil
}

// UpdateStrategy updates an existing strategy
func (s *StrategyService) UpdateStrategy(strategyID string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	strategy, exists := s.strategies[strategyID]
	if !exists {
		return errors.New("strategy not found")
	}

	// Update fields from map
	if name, ok := updates["name"].(string); ok {
		strategy.Name = name
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		strategy.Enabled = enabled
	}
	if mode, ok := updates["mode"].(string); ok {
		strategy.Mode = StrategyMode(mode)
	}
	if config, ok := updates["config"].(map[string]interface{}); ok {
		strategy.Config = config
	}

	strategy.UpdatedAt = time.Now()

	log.Printf("[Strategy] Updated: %s", strategy.Name)
	return nil
}

// DeleteStrategy deletes a strategy
func (s *StrategyService) DeleteStrategy(strategyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.strategies[strategyID]; !exists {
		return errors.New("strategy not found")
	}

	delete(s.strategies, strategyID)
	delete(s.signals, strategyID)

	log.Printf("[Strategy] Deleted: %s", strategyID)
	return nil
}

// GetStrategy retrieves a strategy
func (s *StrategyService) GetStrategy(strategyID string) (*Strategy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategy, exists := s.strategies[strategyID]
	if !exists {
		return nil, errors.New("strategy not found")
	}

	return strategy, nil
}

// GetAllStrategies returns all strategies
func (s *StrategyService) GetAllStrategies() []*Strategy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	strategies := make([]*Strategy, 0, len(s.strategies))
	for _, strategy := range s.strategies {
		strategies = append(strategies, strategy)
	}

	return strategies
}

// GenerateSignal creates a strategy signal
func (s *StrategyService) GenerateSignal(signal *StrategySignal) error {
	signal.ID = uuid.New().String()
	signal.Timestamp = time.Now()
	signal.Executed = false

	s.mu.Lock()
	s.signals[signal.StrategyID] = append(s.signals[signal.StrategyID], signal)
	s.mu.Unlock()

	log.Printf("[Signal] Generated: %s %s %s @ %.5f (confidence: %.1f%%)",
		signal.StrategyID[:8], signal.Side, signal.Symbol, signal.Price, signal.Confidence)

	return nil
}

// GetSignals returns signals for a strategy
func (s *StrategyService) GetSignals(strategyID string, limit int) []*StrategySignal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	signals, exists := s.signals[strategyID]
	if !exists {
		return make([]*StrategySignal, 0)
	}

	if limit > 0 && len(signals) > limit {
		return signals[len(signals)-limit:]
	}

	return signals
}

// RunBacktest executes a backtest for a strategy
func (s *StrategyService) RunBacktest(
	strategyID string,
	startDate, endDate time.Time,
	initialBalance float64,
) (*BacktestResult, error) {

	strategy, err := s.GetStrategy(strategyID)
	if err != nil {
		return nil, err
	}

	log.Printf("[Backtest] Starting for %s from %v to %v", strategy.Name, startDate, endDate)

	result := &BacktestResult{
		StrategyID:     strategyID,
		StartDate:      startDate,
		EndDate:        endDate,
		InitialBalance: initialBalance,
		FinalBalance:   initialBalance,
		Stats:          &StrategyStats{},
		Trades:         make([]BacktestTrade, 0),
		EquityCurve:    make([]EquityPoint, 0),
		DrawdownCurve:  make([]DrawdownPoint, 0),
	}

	// Fetch historical data for all symbols
	for _, symbol := range strategy.Symbols {
		if s.dataCallback == nil {
			return nil, errors.New("data callback not set")
		}

		bars, err := s.dataCallback(symbol, strategy.Timeframe, startDate, endDate)
		if err != nil {
			log.Printf("[Backtest] Error fetching data for %s: %v", symbol, err)
			continue
		}

		// Run strategy logic on historical bars
		s.backtestSymbol(strategy, symbol, bars, result)
	}

	// Calculate final statistics
	s.calculateBacktestStats(result)

	result.CompletedAt = time.Now()

	s.mu.Lock()
	s.backtests[strategyID] = result
	s.mu.Unlock()

	log.Printf("[Backtest] Completed for %s: %.2f%% return, %d trades",
		strategy.Name,
		((result.FinalBalance-result.InitialBalance)/result.InitialBalance)*100,
		result.Stats.TotalTrades)

	return result, nil
}

func (s *StrategyService) backtestSymbol(
	strategy *Strategy,
	symbol string,
	bars []OHLCBar,
	result *BacktestResult,
) {
	// This is a simplified backtest - real implementation would
	// call strategy-specific logic based on strategy.Type

	// Example: Simple MA crossover strategy
	if strategy.Type == StrategyTypeIndicatorBased {
		s.backtestIndicatorStrategy(strategy, symbol, bars, result)
	}
}

func (s *StrategyService) backtestIndicatorStrategy(
	strategy *Strategy,
	symbol string,
	bars []OHLCBar,
	result *BacktestResult,
) {
	// Example: MA crossover
	fastPeriod := 10
	slowPeriod := 20

	if fp, ok := strategy.Config["fastPeriod"].(float64); ok {
		fastPeriod = int(fp)
	}
	if sp, ok := strategy.Config["slowPeriod"].(float64); ok {
		slowPeriod = int(sp)
	}

	// Calculate MAs
	fastMA := s.calculateSMA(bars, fastPeriod)
	slowMA := s.calculateSMA(bars, slowPeriod)

	var currentTrade *BacktestTrade

	for i := slowPeriod; i < len(bars); i++ {
		// Check for crossover
		if currentTrade == nil {
			// Look for entry
			if fastMA[i] > slowMA[i] && fastMA[i-1] <= slowMA[i-1] {
				// Bullish crossover - BUY
				currentTrade = &BacktestTrade{
					EntryTime:  time.Unix(bars[i].Timestamp, 0),
					Symbol:     symbol,
					Side:       "BUY",
					Volume:     1.0,
					EntryPrice: bars[i].Close,
				}
			} else if fastMA[i] < slowMA[i] && fastMA[i-1] >= slowMA[i-1] {
				// Bearish crossover - SELL
				currentTrade = &BacktestTrade{
					EntryTime:  time.Unix(bars[i].Timestamp, 0),
					Symbol:     symbol,
					Side:       "SELL",
					Volume:     1.0,
					EntryPrice: bars[i].Close,
				}
			}
		} else {
			// Check for exit
			shouldExit := false
			exitReason := ""

			if currentTrade.Side == "BUY" {
				// Exit on bearish crossover
				if fastMA[i] < slowMA[i] && fastMA[i-1] >= slowMA[i-1] {
					shouldExit = true
					exitReason = "Signal"
				}
			} else {
				// Exit on bullish crossover
				if fastMA[i] > slowMA[i] && fastMA[i-1] <= slowMA[i-1] {
					shouldExit = true
					exitReason = "Signal"
				}
			}

			if shouldExit {
				currentTrade.ExitTime = time.Unix(bars[i].Timestamp, 0)
				currentTrade.ExitPrice = bars[i].Close
				currentTrade.ExitReason = exitReason

				// Calculate profit
				if currentTrade.Side == "BUY" {
					currentTrade.Profit = (currentTrade.ExitPrice - currentTrade.EntryPrice) * currentTrade.Volume * 100000
				} else {
					currentTrade.Profit = (currentTrade.EntryPrice - currentTrade.ExitPrice) * currentTrade.Volume * 100000
				}

				result.Trades = append(result.Trades, *currentTrade)
				result.FinalBalance += currentTrade.Profit

				currentTrade = nil
			}
		}
	}
}

func (s *StrategyService) calculateBacktestStats(result *BacktestResult) {
	stats := result.Stats

	stats.TotalTrades = len(result.Trades)
	stats.TotalProfit = 0
	stats.TotalLoss = 0

	for _, trade := range result.Trades {
		if trade.Profit > 0 {
			stats.WinningTrades++
			stats.TotalProfit += trade.Profit
			if trade.Profit > stats.LargestWin {
				stats.LargestWin = trade.Profit
			}
		} else {
			stats.LosingTrades++
			stats.TotalLoss += trade.Profit
			if trade.Profit < stats.LargestLoss {
				stats.LargestLoss = trade.Profit
			}
		}
	}

	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinningTrades) / float64(stats.TotalTrades) * 100
	}

	stats.NetProfit = stats.TotalProfit + stats.TotalLoss

	if stats.TotalLoss != 0 {
		stats.ProfitFactor = stats.TotalProfit / math.Abs(stats.TotalLoss)
	}

	if stats.WinningTrades > 0 {
		stats.AverageWin = stats.TotalProfit / float64(stats.WinningTrades)
	}

	if stats.LosingTrades > 0 {
		stats.AverageLoss = stats.TotalLoss / float64(stats.LosingTrades)
	}

	// Calculate Sharpe Ratio (simplified)
	stats.SharpeRatio = s.calculateSharpeRatio(result.Trades)

	stats.LastUpdated = time.Now()
}

func (s *StrategyService) calculateSharpeRatio(trades []BacktestTrade) float64 {
	if len(trades) == 0 {
		return 0
	}

	// Calculate average return
	var totalReturn float64
	for _, trade := range trades {
		totalReturn += trade.Profit
	}
	avgReturn := totalReturn / float64(len(trades))

	// Calculate standard deviation
	var variance float64
	for _, trade := range trades {
		variance += math.Pow(trade.Profit-avgReturn, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(trades)))

	if stdDev == 0 {
		return 0
	}

	// Sharpe ratio (assuming risk-free rate of 0)
	return avgReturn / stdDev
}

func (s *StrategyService) calculateSMA(bars []OHLCBar, period int) []float64 {
	sma := make([]float64, len(bars))

	for i := period - 1; i < len(bars); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += bars[j].Close
		}
		sma[i] = sum / float64(period)
	}

	return sma
}

// processLoop monitors strategies and executes signals
func (s *StrategyService) processLoop() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		s.processStrategies()
	}
}

func (s *StrategyService) processStrategies() {
	s.mu.RLock()
	strategies := make([]*Strategy, 0)
	for _, strategy := range s.strategies {
		if strategy.Enabled && strategy.Mode != ModeBacktest {
			strategies = append(strategies, strategy)
		}
	}
	s.mu.RUnlock()

	// Process each enabled strategy
	for _, strategy := range strategies {
		// Check for signals and execute if in live mode
		signals := s.GetSignals(strategy.ID, 10)
		for _, signal := range signals {
			if !signal.Executed && strategy.Mode == ModeLive {
				if s.executeCallback != nil {
					orderID, err := s.executeCallback(signal)
					if err == nil {
						signal.Executed = true
						signal.OrderID = orderID
						log.Printf("[Strategy] Executed signal %s: %s", signal.ID[:8], orderID)
					}
				}
			}
		}
	}
}

// ExportStrategyJSON exports strategy configuration
func (s *StrategyService) ExportStrategyJSON(strategyID string) (string, error) {
	strategy, err := s.GetStrategy(strategyID)
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(strategy, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ImportStrategyJSON imports strategy from JSON
func (s *StrategyService) ImportStrategyJSON(jsonData string) (*Strategy, error) {
	var strategy Strategy
	if err := json.Unmarshal([]byte(jsonData), &strategy); err != nil {
		return nil, err
	}

	return s.CreateStrategy(&strategy)
}
