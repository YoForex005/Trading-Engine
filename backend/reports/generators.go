package reports

import (
	"fmt"
	"sort"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// ReportGenerator generates reports from engine data
type ReportGenerator struct {
	engine *core.Engine
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(engine *core.Engine) *ReportGenerator {
	return &ReportGenerator{
		engine: engine,
	}
}

// DailyTradingSummary generates a daily trading summary report
func (g *ReportGenerator) DailyTradingSummary(date time.Time) *Report {
	// Get all accounts and trades
	allAccounts := g.engine.GetAllAccounts()

	// Aggregate trades for the day
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.Add(24 * time.Hour)

	var allTrades []core.Trade
	for _, acc := range allAccounts {
		trades := g.engine.GetTrades(acc.ID)
		allTrades = append(allTrades, trades...)
	}

	// Filter trades for this day
	var dayTrades []core.Trade
	var totalVolume float64
	var totalPnL float64
	symbolVolumes := make(map[string]float64)
	var largestTrade *core.Trade
	largestTradeVolume := 0.0

	for _, trade := range allTrades {
		if trade.ExecutedAt.After(dayStart) && trade.ExecutedAt.Before(dayEnd) {
			dayTrades = append(dayTrades, trade)
			totalVolume += trade.Volume
			totalPnL += trade.RealizedPnL
			symbolVolumes[trade.Symbol] += trade.Volume

			if trade.Volume > largestTradeVolume {
				largestTradeVolume = trade.Volume
				largestTrade = &trade
			}
		}
	}

	// Get top 5 symbols by volume
	type symbolVolume struct {
		Symbol string
		Volume float64
	}
	var symbolList []symbolVolume
	for symbol, volume := range symbolVolumes {
		symbolList = append(symbolList, symbolVolume{Symbol: symbol, Volume: volume})
	}
	sort.Slice(symbolList, func(i, j int) bool {
		return symbolList[i].Volume > symbolList[j].Volume
	})
	if len(symbolList) > 5 {
		symbolList = symbolList[:5]
	}

	// Count accounts created today
	newAccountsCount := 0
	for _, acc := range allAccounts {
		accCreated := time.Unix(acc.CreatedAt, 0)
		if accCreated.After(dayStart) && accCreated.Before(dayEnd) {
			newAccountsCount++
		}
	}

	// Count active accounts (accounts with trades today)
	activeAccountsMap := make(map[int64]bool)
	for _, trade := range dayTrades {
		activeAccountsMap[trade.AccountID] = true
	}

	// Build report data
	data := map[string]interface{}{
		"date":              date.Format("2006-01-02"),
		"totalTrades":       len(dayTrades),
		"totalVolume":       totalVolume,
		"totalPnL":          totalPnL,
		"topSymbols":        symbolList,
		"activeAccounts":    len(activeAccountsMap),
		"newAccounts":       newAccountsCount,
		"totalAccounts":     len(allAccounts),
	}

	if largestTrade != nil {
		data["largestTrade"] = map[string]interface{}{
			"symbol": largestTrade.Symbol,
			"volume": largestTrade.Volume,
			"side":   largestTrade.Side,
			"pnl":    largestTrade.RealizedPnL,
		}
	}

	return &Report{
		Type:        "daily_trading",
		Title:       fmt.Sprintf("Daily Trading Summary - %s", date.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Period:      date.Format("2006-01-02"),
		Data:        data,
		Format:      "json",
	}
}

// DailyRiskReport generates a daily risk report
func (g *ReportGenerator) DailyRiskReport(date time.Time) *Report {
	allAccounts := g.engine.GetAllAccounts()
	allPositions := g.engine.GetAllPositions()

	// Calculate exposure by symbol
	exposureBySymbol := make(map[string]float64)
	for _, pos := range allPositions {
		if pos.Status == "OPEN" {
			// Calculate notional value (simplified)
			notional := pos.Volume * pos.CurrentPrice * 100000 // Assuming forex contract size
			exposureBySymbol[pos.Symbol] += notional
		}
	}

	// Calculate margin metrics
	var totalMarginUsed float64
	var totalEquity float64
	var accountsNearMarginCall []map[string]interface{}

	for _, acc := range allAccounts {
		summary, err := g.engine.GetAccountSummary(acc.ID)
		if err != nil {
			continue
		}

		totalMarginUsed += summary.Margin
		totalEquity += summary.Equity

		// Check for accounts near margin call (margin level < 150%)
		if summary.MarginLevel > 0 && summary.MarginLevel < 150 {
			accountsNearMarginCall = append(accountsNearMarginCall, map[string]interface{}{
				"accountId":     acc.ID,
				"accountNumber": acc.AccountNumber,
				"marginLevel":   summary.MarginLevel,
				"equity":        summary.Equity,
				"margin":        summary.Margin,
			})
		}
	}

	// Calculate average margin utilization
	avgMarginUtilization := 0.0
	if totalEquity > 0 {
		avgMarginUtilization = (totalMarginUsed / totalEquity) * 100
	}

	// Build report data
	data := map[string]interface{}{
		"date":                   date.Format("2006-01-02"),
		"exposureBySymbol":       exposureBySymbol,
		"avgMarginUtilization":   avgMarginUtilization,
		"totalMarginUsed":        totalMarginUsed,
		"totalEquity":            totalEquity,
		"accountsNearMarginCall": accountsNearMarginCall,
		"openPositions":          len(allPositions),
		"circuitBreakerTriggers": 0, // Placeholder - would track actual triggers
	}

	return &Report{
		Type:        "daily_risk",
		Title:       fmt.Sprintf("Daily Risk Report - %s", date.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Period:      date.Format("2006-01-02"),
		Data:        data,
		Format:      "json",
	}
}

// WeeklyRevenueReport generates a weekly revenue report
func (g *ReportGenerator) WeeklyRevenueReport(startDate, endDate time.Time) *Report {
	allAccounts := g.engine.GetAllAccounts()
	ledger := g.engine.GetLedger()

	var commissionRevenue float64
	var swapRevenue float64
	revenueBySymbol := make(map[string]float64)

	// Get all ledger entries for the week
	for _, acc := range allAccounts {
		entries := ledger.GetEntries(acc.ID)
		for _, entry := range entries {
			if entry.CreatedAt.After(startDate) && entry.CreatedAt.Before(endDate) {
				switch entry.Type {
				case "COMMISSION":
					// Commissions are broker revenue (negative for client = positive for broker)
					commissionRevenue += -entry.Amount
				case "SWAP":
					// Swaps can be revenue or cost
					swapRevenue += -entry.Amount
				}
			}
		}
	}

	// Get all trades for the week to calculate spread revenue
	var allTrades []core.Trade
	for _, acc := range allAccounts {
		trades := g.engine.GetTrades(acc.ID)
		allTrades = append(allTrades, trades...)
	}

	spreadRevenue := 0.0
	for _, trade := range allTrades {
		if trade.ExecutedAt.After(startDate) && trade.ExecutedAt.Before(endDate) {
			// Estimate spread revenue: ~$10 per lot average
			estimatedSpread := trade.Volume * 10.0
			spreadRevenue += estimatedSpread
			revenueBySymbol[trade.Symbol] += estimatedSpread
		}
	}

	totalRevenue := spreadRevenue + commissionRevenue + swapRevenue

	// Calculate week-over-week change (would need historical data)
	// For now, use placeholder
	weekOverWeekChange := 0.0

	// Build report data
	data := map[string]interface{}{
		"startDate":          startDate.Format("2006-01-02"),
		"endDate":            endDate.Format("2006-01-02"),
		"spreadRevenue":      spreadRevenue,
		"commissionRevenue":  commissionRevenue,
		"swapRevenue":        swapRevenue,
		"totalRevenue":       totalRevenue,
		"revenueBySymbol":    revenueBySymbol,
		"weekOverWeekChange": weekOverWeekChange,
	}

	// Format period as ISO week
	year, week := startDate.ISOWeek()
	period := fmt.Sprintf("%d-W%02d", year, week)

	return &Report{
		Type:        "weekly_revenue",
		Title:       fmt.Sprintf("Weekly Revenue Report - Week %d, %d", week, year),
		GeneratedAt: time.Now(),
		Period:      period,
		Data:        data,
		Format:      "json",
	}
}
