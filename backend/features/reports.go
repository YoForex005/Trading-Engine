package features

import (
	"errors"
	"log"
	"math"
	"sort"
	"sync"
	"time"
)

// Advanced Reporting System
// - Tax reports
// - Performance attribution
// - Drawdown analysis
// - Sharpe ratio, Sortino ratio
// - MAE/MFE analysis (Maximum Adverse/Favorable Excursion)
// - R-multiple analysis
// - Win/Loss streak analysis
// - Time-based performance
// - Symbol-based performance

// Trade represents a completed trade for reporting
type Trade struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"accountId"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"` // BUY or SELL
	Volume      float64   `json:"volume"`
	OpenPrice   float64   `json:"openPrice"`
	ClosePrice  float64   `json:"closePrice"`
	OpenTime    time.Time `json:"openTime"`
	CloseTime   time.Time `json:"closeTime"`
	Profit      float64   `json:"profit"`
	ProfitPips  float64   `json:"profitPips"`
	Commission  float64   `json:"commission"`
	Swap        float64   `json:"swap"`
	MAE         float64   `json:"mae"` // Maximum Adverse Excursion
	MFE         float64   `json:"mfe"` // Maximum Favorable Excursion
	RMultiple   float64   `json:"rMultiple"`
	HoldingTime int64     `json:"holdingTime"` // Seconds
	ExitReason  string    `json:"exitReason"`  // TP, SL, Manual
}

// TaxReport represents tax reporting data
type TaxReport struct {
	AccountID     string    `json:"accountId"`
	Year          int       `json:"year"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	TotalTrades   int       `json:"totalTrades"`
	TotalProfit   float64   `json:"totalProfit"`
	TotalLoss     float64   `json:"totalLoss"`
	NetProfit     float64   `json:"netProfit"`
	TotalCommission float64 `json:"totalCommission"`
	TotalSwap     float64   `json:"totalSwap"`
	TaxableProfitUSD float64 `json:"taxableProfitUsd"`
	ShortTermGains float64  `json:"shortTermGains"` // Held < 1 year
	LongTermGains  float64  `json:"longTermGains"`  // Held >= 1 year
	TradesBySymbol map[string]*TaxSymbolData `json:"tradesBySymbol"`
	GeneratedAt   time.Time `json:"generatedAt"`
}

// TaxSymbolData represents tax data per symbol
type TaxSymbolData struct {
	Symbol      string  `json:"symbol"`
	Trades      int     `json:"trades"`
	NetProfit   float64 `json:"netProfit"`
	Commission  float64 `json:"commission"`
}

// PerformanceReport represents overall performance analysis
type PerformanceReport struct {
	AccountID       string    `json:"accountId"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`

	// Basic Metrics
	TotalTrades     int     `json:"totalTrades"`
	WinningTrades   int     `json:"winningTrades"`
	LosingTrades    int     `json:"losingTrades"`
	WinRate         float64 `json:"winRate"`
	TotalProfit     float64 `json:"totalProfit"`
	TotalLoss       float64 `json:"totalLoss"`
	NetProfit       float64 `json:"netProfit"`

	// Averages
	AverageWin      float64 `json:"averageWin"`
	AverageLoss     float64 `json:"averageLoss"`
	AverageTrade    float64 `json:"averageTrade"`
	LargestWin      float64 `json:"largestWin"`
	LargestLoss     float64 `json:"largestLoss"`

	// Risk Metrics
	ProfitFactor    float64 `json:"profitFactor"`
	SharpeRatio     float64 `json:"sharpeRatio"`
	SortinoRatio    float64 `json:"sortinoRatio"`
	CalmarRatio     float64 `json:"calmarRatio"`
	MaxDrawdown     float64 `json:"maxDrawdown"`
	MaxDrawdownPct  float64 `json:"maxDrawdownPct"`
	RecoveryFactor  float64 `json:"recoveryFactor"`

	// MAE/MFE
	AverageMAE      float64 `json:"averageMAE"`
	AverageMFE      float64 `json:"averageMFE"`
	MAEMFERatio     float64 `json:"maeMfeRatio"`

	// R-Multiple
	AverageRMultiple float64 `json:"averageRMultiple"`
	MedianRMultiple  float64 `json:"medianRMultiple"`

	// Streaks
	LongestWinStreak  int   `json:"longestWinStreak"`
	LongestLossStreak int   `json:"longestLossStreak"`
	CurrentStreak     int   `json:"currentStreak"`

	// Time Analysis
	AverageHoldingTime int64 `json:"averageHoldingTime"` // Seconds

	// Attribution
	ProfitBySymbol     map[string]float64 `json:"profitBySymbol"`
	TradesBySymbol     map[string]int     `json:"tradesBySymbol"`
	ProfitByHour       map[int]float64    `json:"profitByHour"`
	ProfitByDayOfWeek  map[string]float64 `json:"profitByDayOfWeek"`

	GeneratedAt     time.Time `json:"generatedAt"`
}

// DrawdownAnalysis represents drawdown analysis
type DrawdownAnalysis struct {
	AccountID       string    `json:"accountId"`
	StartDate       time.Time `json:"startDate"`
	EndDate         time.Time `json:"endDate"`

	// Current Drawdown
	CurrentDrawdown    float64 `json:"currentDrawdown"`
	CurrentDrawdownPct float64 `json:"currentDrawdownPct"`

	// Maximum Drawdown
	MaxDrawdown        float64   `json:"maxDrawdown"`
	MaxDrawdownPct     float64   `json:"maxDrawdownPct"`
	MaxDrawdownStart   time.Time `json:"maxDrawdownStart"`
	MaxDrawdownEnd     time.Time `json:"maxDrawdownEnd"`
	MaxDrawdownDuration int64    `json:"maxDrawdownDuration"` // Days

	// Recovery
	RecoveryTime       int64     `json:"recoveryTime"` // Days
	RecoveryFactor     float64   `json:"recoveryFactor"`

	// Drawdown Periods
	DrawdownPeriods    []DrawdownPeriod `json:"drawdownPeriods"`

	// Statistics
	AverageDrawdown    float64 `json:"averageDrawdown"`
	AverageRecoveryTime int64  `json:"averageRecoveryTime"` // Days

	GeneratedAt        time.Time `json:"generatedAt"`
}

// DrawdownPeriod represents a drawdown period
type DrawdownPeriod struct {
	StartDate      time.Time `json:"startDate"`
	EndDate        time.Time `json:"endDate"`
	RecoveryDate   *time.Time `json:"recoveryDate,omitempty"`
	Drawdown       float64   `json:"drawdown"`
	DrawdownPct    float64   `json:"drawdownPct"`
	Duration       int64     `json:"duration"`       // Days in drawdown
	RecoveryTime   int64     `json:"recoveryTime"`   // Days to recover
}

// ReportService generates advanced trading reports
type ReportService struct {
	mu          sync.RWMutex
	trades      map[string][]Trade // accountID -> trades

	// Callbacks
	getTradesCallback func(accountID string, startDate, endDate time.Time) ([]Trade, error)
}

// NewReportService creates the report service
func NewReportService() *ReportService {
	svc := &ReportService{
		trades: make(map[string][]Trade),
	}

	log.Println("[ReportService] Initialized with tax, performance, and drawdown reporting")
	return svc
}

// SetCallbacks configures callbacks
func (s *ReportService) SetCallbacks(
	getTradesCallback func(accountID string, startDate, endDate time.Time) ([]Trade, error),
) {
	s.getTradesCallback = getTradesCallback
}

// GenerateTaxReport generates a tax report
func (s *ReportService) GenerateTaxReport(accountID string, year int) (*TaxReport, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)

	trades, err := s.getTrades(accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	report := &TaxReport{
		AccountID:      accountID,
		Year:           year,
		StartDate:      startDate,
		EndDate:        endDate,
		TotalTrades:    len(trades),
		TradesBySymbol: make(map[string]*TaxSymbolData),
		GeneratedAt:    time.Now(),
	}

	for _, trade := range trades {
		// Calculate holding period
		holdingDays := int(trade.CloseTime.Sub(trade.OpenTime).Hours() / 24)

		netProfit := trade.Profit - trade.Commission - trade.Swap

		report.TotalCommission += trade.Commission
		report.TotalSwap += trade.Swap

		if netProfit > 0 {
			report.TotalProfit += netProfit

			// Classify as short-term or long-term
			if holdingDays >= 365 {
				report.LongTermGains += netProfit
			} else {
				report.ShortTermGains += netProfit
			}
		} else {
			report.TotalLoss += netProfit
		}

		// By symbol
		if _, exists := report.TradesBySymbol[trade.Symbol]; !exists {
			report.TradesBySymbol[trade.Symbol] = &TaxSymbolData{
				Symbol: trade.Symbol,
			}
		}
		symbolData := report.TradesBySymbol[trade.Symbol]
		symbolData.Trades++
		symbolData.NetProfit += netProfit
		symbolData.Commission += trade.Commission
	}

	report.NetProfit = report.TotalProfit + report.TotalLoss
	report.TaxableProfitUSD = report.NetProfit

	log.Printf("[TaxReport] Generated for %s year %d: %.2f net profit, %d trades",
		accountID, year, report.NetProfit, report.TotalTrades)

	return report, nil
}

// GeneratePerformanceReport generates a performance report
func (s *ReportService) GeneratePerformanceReport(
	accountID string,
	startDate, endDate time.Time,
) (*PerformanceReport, error) {

	trades, err := s.getTrades(accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(trades) == 0 {
		return nil, errors.New("no trades found for period")
	}

	report := &PerformanceReport{
		AccountID:        accountID,
		StartDate:        startDate,
		EndDate:          endDate,
		TotalTrades:      len(trades),
		ProfitBySymbol:   make(map[string]float64),
		TradesBySymbol:   make(map[string]int),
		ProfitByHour:     make(map[int]float64),
		ProfitByDayOfWeek: make(map[string]float64),
		GeneratedAt:      time.Now(),
	}

	// Calculate basic metrics
	returns := make([]float64, 0)
	downside := make([]float64, 0)
	rMultiples := make([]float64, 0)
	maes := make([]float64, 0)
	mfes := make([]float64, 0)
	holdingTimes := make([]int64, 0)

	currentStreak := 0
	winStreak := 0
	lossStreak := 0

	for _, trade := range trades {
		netProfit := trade.Profit - trade.Commission - trade.Swap

		if netProfit > 0 {
			report.WinningTrades++
			report.TotalProfit += netProfit

			if netProfit > report.LargestWin {
				report.LargestWin = netProfit
			}

			currentStreak++
			if currentStreak > winStreak {
				winStreak = currentStreak
			}
		} else {
			report.LosingTrades++
			report.TotalLoss += netProfit

			if netProfit < report.LargestLoss {
				report.LargestLoss = netProfit
			}

			currentStreak--
			if -currentStreak > lossStreak {
				lossStreak = -currentStreak
			}

			downside = append(downside, netProfit)
		}

		returns = append(returns, netProfit)

		// MAE/MFE
		maes = append(maes, trade.MAE)
		mfes = append(mfes, trade.MFE)

		// R-Multiple
		if trade.RMultiple != 0 {
			rMultiples = append(rMultiples, trade.RMultiple)
		}

		// Holding time
		holdingTimes = append(holdingTimes, trade.HoldingTime)

		// Attribution
		report.ProfitBySymbol[trade.Symbol] += netProfit
		report.TradesBySymbol[trade.Symbol]++

		hour := trade.CloseTime.Hour()
		report.ProfitByHour[hour] += netProfit

		dayOfWeek := trade.CloseTime.Weekday().String()
		report.ProfitByDayOfWeek[dayOfWeek] += netProfit
	}

	report.NetProfit = report.TotalProfit + report.TotalLoss

	// Calculate averages
	if report.WinningTrades > 0 {
		report.AverageWin = report.TotalProfit / float64(report.WinningTrades)
	}

	if report.LosingTrades > 0 {
		report.AverageLoss = report.TotalLoss / float64(report.LosingTrades)
	}

	if report.TotalTrades > 0 {
		report.WinRate = float64(report.WinningTrades) / float64(report.TotalTrades) * 100
		report.AverageTrade = report.NetProfit / float64(report.TotalTrades)
	}

	// Profit factor
	if report.TotalLoss != 0 {
		report.ProfitFactor = report.TotalProfit / math.Abs(report.TotalLoss)
	}

	// Sharpe Ratio
	report.SharpeRatio = calculateSharpeRatio(returns)

	// Sortino Ratio
	report.SortinoRatio = calculateSortinoRatio(returns, downside)

	// MAE/MFE
	if len(maes) > 0 {
		sum := 0.0
		for _, mae := range maes {
			sum += mae
		}
		report.AverageMAE = sum / float64(len(maes))
	}

	if len(mfes) > 0 {
		sum := 0.0
		for _, mfe := range mfes {
			sum += mfe
		}
		report.AverageMFE = sum / float64(len(mfes))
	}

	if report.AverageMAE != 0 {
		report.MAEMFERatio = report.AverageMFE / report.AverageMAE
	}

	// R-Multiple
	if len(rMultiples) > 0 {
		sum := 0.0
		for _, r := range rMultiples {
			sum += r
		}
		report.AverageRMultiple = sum / float64(len(rMultiples))

		sort.Float64s(rMultiples)
		mid := len(rMultiples) / 2
		if len(rMultiples)%2 == 0 {
			report.MedianRMultiple = (rMultiples[mid-1] + rMultiples[mid]) / 2
		} else {
			report.MedianRMultiple = rMultiples[mid]
		}
	}

	// Streaks
	report.LongestWinStreak = winStreak
	report.LongestLossStreak = lossStreak
	report.CurrentStreak = currentStreak

	// Holding time
	if len(holdingTimes) > 0 {
		sum := int64(0)
		for _, ht := range holdingTimes {
			sum += ht
		}
		report.AverageHoldingTime = sum / int64(len(holdingTimes))
	}

	log.Printf("[PerformanceReport] Generated for %s: %.2f%% win rate, %.2f Sharpe ratio",
		accountID, report.WinRate, report.SharpeRatio)

	return report, nil
}

// GenerateDrawdownAnalysis generates drawdown analysis
func (s *ReportService) GenerateDrawdownAnalysis(
	accountID string,
	startDate, endDate time.Time,
	initialBalance float64,
) (*DrawdownAnalysis, error) {

	trades, err := s.getTrades(accountID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(trades) == 0 {
		return nil, errors.New("no trades found for period")
	}

	// Sort trades by close time
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].CloseTime.Before(trades[j].CloseTime)
	})

	analysis := &DrawdownAnalysis{
		AccountID:       accountID,
		StartDate:       startDate,
		EndDate:         endDate,
		DrawdownPeriods: make([]DrawdownPeriod, 0),
		GeneratedAt:     time.Now(),
	}

	// Calculate equity curve and drawdowns
	balance := initialBalance
	peak := initialBalance
	var drawdownStart *time.Time
	var currentDD DrawdownPeriod
	totalDD := 0.0
	ddCount := 0
	totalRecovery := int64(0)

	for _, trade := range trades {
		netProfit := trade.Profit - trade.Commission - trade.Swap
		balance += netProfit

		if balance > peak {
			// New peak - end of drawdown period
			if drawdownStart != nil {
				currentDD.RecoveryDate = &trade.CloseTime
				days := int64(trade.CloseTime.Sub(*drawdownStart).Hours() / 24)
				currentDD.RecoveryTime = days
				totalRecovery += days

				analysis.DrawdownPeriods = append(analysis.DrawdownPeriods, currentDD)
				drawdownStart = nil
				ddCount++
			}

			peak = balance
		} else if balance < peak {
			// In drawdown
			dd := peak - balance
			ddPct := (dd / peak) * 100

			if drawdownStart == nil {
				// Start of new drawdown
				drawdownStart = &trade.CloseTime
				currentDD = DrawdownPeriod{
					StartDate:   trade.CloseTime,
					Drawdown:    dd,
					DrawdownPct: ddPct,
				}
			} else {
				// Update current drawdown
				currentDD.EndDate = trade.CloseTime
				currentDD.Drawdown = dd
				currentDD.DrawdownPct = ddPct
				currentDD.Duration = int64(trade.CloseTime.Sub(*drawdownStart).Hours() / 24)
			}

			totalDD += dd

			// Track maximum drawdown
			if dd > analysis.MaxDrawdown {
				analysis.MaxDrawdown = dd
				analysis.MaxDrawdownPct = ddPct
				analysis.MaxDrawdownStart = *drawdownStart
				analysis.MaxDrawdownEnd = trade.CloseTime
				analysis.MaxDrawdownDuration = currentDD.Duration
			}
		}
	}

	// Current drawdown
	if drawdownStart != nil {
		analysis.CurrentDrawdown = currentDD.Drawdown
		analysis.CurrentDrawdownPct = currentDD.DrawdownPct
	}

	// Calculate averages
	if len(analysis.DrawdownPeriods) > 0 {
		analysis.AverageDrawdown = totalDD / float64(len(analysis.DrawdownPeriods))
	}

	if ddCount > 0 {
		analysis.AverageRecoveryTime = totalRecovery / int64(ddCount)
	}

	// Recovery factor
	if analysis.MaxDrawdown != 0 {
		netProfit := balance - initialBalance
		analysis.RecoveryFactor = netProfit / analysis.MaxDrawdown
	}

	log.Printf("[DrawdownAnalysis] Generated for %s: %.2f max drawdown (%.2f%%)",
		accountID, analysis.MaxDrawdown, analysis.MaxDrawdownPct)

	return analysis, nil
}

// Helper functions

func (s *ReportService) getTrades(accountID string, startDate, endDate time.Time) ([]Trade, error) {
	if s.getTradesCallback != nil {
		return s.getTradesCallback(accountID, startDate, endDate)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	trades, exists := s.trades[accountID]
	if !exists {
		return make([]Trade, 0), nil
	}

	// Filter by date
	filtered := make([]Trade, 0)
	for _, trade := range trades {
		if trade.CloseTime.After(startDate) && trade.CloseTime.Before(endDate) {
			filtered = append(filtered, trade)
		}
	}

	return filtered, nil
}

func calculateSharpeRatio(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate average return
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	avgReturn := sum / float64(len(returns))

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-avgReturn, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	if stdDev == 0 {
		return 0
	}

	// Sharpe ratio (assuming risk-free rate of 0)
	return avgReturn / stdDev
}

func calculateSortinoRatio(returns, downside []float64) float64 {
	if len(returns) == 0 || len(downside) == 0 {
		return 0
	}

	// Calculate average return
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	avgReturn := sum / float64(len(returns))

	// Calculate downside deviation
	downsideVariance := 0.0
	for _, d := range downside {
		downsideVariance += math.Pow(d, 2)
	}
	downsideDev := math.Sqrt(downsideVariance / float64(len(downside)))

	if downsideDev == 0 {
		return 0
	}

	// Sortino ratio (assuming target return of 0)
	return avgReturn / downsideDev
}
