package features

import (
	"log"
	"time"
)

// IntegrationExample shows how to wire all advanced features together
// with the main trading engine
//
// Add this to your main.go:
//
// 1. Import: import "github.com/epic1st/rtx/backend/features"
// 2. Call: features.InitializeAdvancedFeatures(bbookEngine, hub, lpMgr)
//

// InitializeAdvancedFeatures sets up all advanced trading features
// and integrates them with the main engine
func InitializeAdvancedFeatures(
	bbookEngine interface{}, // *core.Engine
	hub interface{},         // *ws.Hub
	lpMgr interface{},       // *lpmanager.Manager
) *FeatureHandlers {

	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Println("║       Initializing Advanced Trading Features             ║")
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	// ===== 1. Initialize Indicator Service =====
	indicatorService := NewIndicatorService(200)

	// Set data callback to fetch historical bars
	indicatorService.SetDataCallback(func(symbol, timeframe string, count int) ([]OHLCBar, error) {
		// This would integrate with your OHLC data storage
		// For now, return empty - implement based on your tickstore
		return make([]OHLCBar, 0), nil
	})

	log.Println("[✓] Indicator Service initialized (SMA, EMA, RSI, MACD, BB, ATR, ADX, Stochastic, Pivots)")

	// ===== 2. Initialize Advanced Order Service =====
	advancedOrderService := NewAdvancedOrderService()

	// Set callbacks
	advancedOrderService.SetCallbacks(
		// Price callback
		func(symbol string) (bid, ask float64, ok bool) {
			// Integrate with hub to get latest prices
			// tick := hub.GetLatestPrice(symbol)
			// if tick != nil {
			//     return tick.Bid, tick.Ask, true
			// }
			return 0, 0, false
		},
		// Execute callback
		func(symbol, side string, volume, price float64) (string, error) {
			// Integrate with B-Book engine to execute order
			// order, err := bbookEngine.PlaceOrder(...)
			// return order.ID, err
			return "", nil
		},
		// Cancel callback
		func(orderID string) error {
			// Integrate with B-Book engine to cancel order
			// return bbookEngine.CancelOrder(orderID)
			return nil
		},
		// Volume callback for VWAP
		func(symbol string) float64 {
			// Return recent volume for symbol
			return 0
		},
	)

	log.Println("[✓] Advanced Order Service initialized (Bracket, TWAP, VWAP, Iceberg)")

	// ===== 3. Initialize Strategy Service =====
	strategyService := NewStrategyService(indicatorService)

	// Set callbacks
	strategyService.SetCallbacks(
		// Price callback
		func(symbol string) (bid, ask float64, ok bool) {
			// Same as above
			return 0, 0, false
		},
		// Execute callback
		func(signal *StrategySignal) (string, error) {
			// Execute strategy signal
			log.Printf("[Strategy] Executing signal: %s %s @ %.5f",
				signal.Side, signal.Symbol, signal.Price)
			// return bbookEngine.PlaceOrder(...)
			return "", nil
		},
		// Data callback for backtesting
		func(symbol, timeframe string, start, end time.Time) ([]OHLCBar, error) {
			// Fetch historical OHLC data for backtesting
			// return tickstore.GetOHLCBars(symbol, timeframe, start, end)
			return make([]OHLCBar, 0), nil
		},
	)

	log.Println("[✓] Strategy Service initialized (Backtesting, Paper Trading, Live Trading)")

	// ===== 4. Initialize Alert Service =====
	alertService := NewAlertService()

	// Set callbacks
	alertService.SetCallbacks(
		// Price callback
		func(symbol string) (bid, ask float64, ok bool) {
			// Same as above
			return 0, 0, false
		},
		// Indicator callback
		func(symbol, indicator string, period int) (float64, error) {
			// Calculate indicator value
			switch indicator {
			case "RSI":
				values, err := indicatorService.RSI(symbol, period)
				if err != nil || len(values) == 0 {
					return 0, err
				}
				return values[len(values)-1].Values["rsi"], nil
			case "MACD":
				values, err := indicatorService.MACD(symbol, 12, 26, 9)
				if err != nil || len(values) == 0 {
					return 0, err
				}
				return values[len(values)-1].Values["macd"], nil
			}
			return 0, nil
		},
	)

	// Set notification providers
	alertService.SetNotificationProviders(
		// Email provider
		func(to, subject, body string) error {
			log.Printf("[Email] To: %s, Subject: %s", to, subject)
			// Integrate with email service (SendGrid, AWS SES, etc.)
			return nil
		},
		// SMS provider
		func(to, message string) error {
			log.Printf("[SMS] To: %s, Message: %s", to, message)
			// Integrate with SMS service (Twilio, etc.)
			return nil
		},
		// Push notification provider
		func(userID, title, message string) error {
			log.Printf("[Push] User: %s, Title: %s", userID, title)
			// Integrate with push service (Firebase, OneSignal, etc.)
			return nil
		},
		// Webhook provider
		func(url string, payload interface{}) error {
			log.Printf("[Webhook] URL: %s", url)
			// HTTP POST to webhook URL
			return alertService.SendWebhook(url, payload)
		},
	)

	log.Println("[✓] Alert Service initialized (Price, Indicator, News alerts with Email/SMS/Push/Webhook)")

	// ===== 5. Initialize Report Service =====
	reportService := NewReportService()

	// Set callbacks
	reportService.SetCallbacks(
		// Get trades callback
		func(accountID string, startDate, endDate time.Time) ([]Trade, error) {
			// Fetch completed trades from B-Book engine
			// trades := bbookEngine.GetClosedTrades(accountID, startDate, endDate)
			// Convert to features.Trade format
			return make([]Trade, 0), nil
		},
	)

	log.Println("[✓] Report Service initialized (Tax Reports, Performance Analytics, Drawdown Analysis)")

	// ===== 6. Create API Handlers =====
	featureHandlers := NewFeatureHandlers(
		advancedOrderService,
		indicatorService,
		strategyService,
		alertService,
		reportService,
	)

	// Register all routes
	featureHandlers.RegisterRoutes()

	log.Println("[✓] API Handlers registered")
	log.Println("")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  ADVANCED FEATURES READY")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  Order Types:  Bracket, TWAP, VWAP, Iceberg, FOK, IOC, GTC")
	log.Println("  Indicators:   RSI, MACD, BB, ATR, ADX, Stochastic, Pivots")
	log.Println("  Strategies:   Backtesting, Paper Trading, Live Automation")
	log.Println("  Alerts:       Price, Indicator, News (Email/SMS/Push)")
	log.Println("  Reports:      Tax, Performance, Drawdown Analysis")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("")

	return featureHandlers
}

// Example: Create a demo strategy
func CreateDemoStrategy(strategyService *StrategyService) {
	strategy := &Strategy{
		Name:        "MA Crossover Demo",
		Description: "Simple moving average crossover strategy",
		Type:        StrategyTypeIndicatorBased,
		Mode:        ModePaper, // Paper trading mode
		Symbols:     []string{"EURUSD", "GBPUSD", "USDJPY"},
		Timeframe:   "M15",
		Enabled:     false, // Start disabled
		Config: map[string]interface{}{
			"fastPeriod": 10,
			"slowPeriod": 20,
		},
		MaxPositions:      3,
		MaxRiskPerTrade:   2.0,  // 2% per trade
		MaxDrawdown:       10.0, // 10% max drawdown
		MaxDailyLoss:      500.0,
		DefaultStopLoss:   50.0, // 50 pips
		DefaultTakeProfit: 100.0, // 100 pips
	}

	created, err := strategyService.CreateStrategy(strategy)
	if err != nil {
		log.Printf("[Error] Failed to create demo strategy: %v", err)
		return
	}

	log.Printf("[✓] Demo strategy created: %s (ID: %s)", created.Name, created.ID)
}

// Example: Create demo alerts
func CreateDemoAlerts(alertService *AlertService) {
	// Price alert
	priceAlert := &Alert{
		UserID:      "demo-user",
		Name:        "EURUSD Above 1.10",
		Type:        AlertTypePrice,
		Symbol:      "EURUSD",
		Condition:   ConditionAbove,
		Value:       1.10000,
		Message:     "EURUSD has crossed above 1.10000",
		Channels:    []NotificationChannel{ChannelInApp},
		TriggerOnce: true,
	}

	alertService.CreateAlert(priceAlert)

	// RSI alert
	rsiAlert := &Alert{
		UserID:    "demo-user",
		Name:      "EURUSD RSI Overbought",
		Type:      AlertTypeIndicator,
		Symbol:    "EURUSD",
		Indicator: "RSI",
		Period:    14,
		Condition: ConditionOverbought,
		Value:     70.0,
		Message:   "EURUSD RSI(14) is above 70 - overbought",
		Channels:  []NotificationChannel{ChannelInApp},
	}

	alertService.CreateAlert(rsiAlert)

	log.Println("[✓] Demo alerts created")
}

// Example: Run a backtest
func RunDemoBacktest(strategyService *StrategyService, strategyID string) {
	startDate := time.Now().AddDate(0, -6, 0) // 6 months ago
	endDate := time.Now()
	initialBalance := 10000.0

	log.Printf("[Backtest] Starting demo backtest...")
	log.Printf("[Backtest] Period: %v to %v", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	log.Printf("[Backtest] Initial Balance: $%.2f", initialBalance)

	result, err := strategyService.RunBacktest(strategyID, startDate, endDate, initialBalance)
	if err != nil {
		log.Printf("[Error] Backtest failed: %v", err)
		return
	}

	log.Println("")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  BACKTEST RESULTS")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Printf("  Total Trades:     %d", result.Stats.TotalTrades)
	log.Printf("  Winning Trades:   %d", result.Stats.WinningTrades)
	log.Printf("  Losing Trades:    %d", result.Stats.LosingTrades)
	log.Printf("  Win Rate:         %.2f%%", result.Stats.WinRate)
	log.Printf("  Net Profit:       $%.2f", result.Stats.NetProfit)
	log.Printf("  Profit Factor:    %.2f", result.Stats.ProfitFactor)
	log.Printf("  Sharpe Ratio:     %.2f", result.Stats.SharpeRatio)
	log.Printf("  Max Drawdown:     $%.2f", result.Stats.MaxDrawdown)
	log.Printf("  Final Balance:    $%.2f", result.FinalBalance)
	log.Printf("  Return:           %.2f%%",
		((result.FinalBalance-result.InitialBalance)/result.InitialBalance)*100)
	log.Println("═══════════════════════════════════════════════════════════")
}

// Example: Generate demo reports
func GenerateDemoReports(reportService *ReportService, accountID string) {
	// Tax report
	taxReport, err := reportService.GenerateTaxReport(accountID, 2024)
	if err == nil {
		log.Println("")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Println("  TAX REPORT 2024")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Printf("  Total Trades:      %d", taxReport.TotalTrades)
		log.Printf("  Net Profit:        $%.2f", taxReport.NetProfit)
		log.Printf("  Short-term Gains:  $%.2f", taxReport.ShortTermGains)
		log.Printf("  Long-term Gains:   $%.2f", taxReport.LongTermGains)
		log.Printf("  Taxable Profit:    $%.2f", taxReport.TaxableProfitUSD)
		log.Println("═══════════════════════════════════════════════════════════")
	}

	// Performance report
	startDate := time.Now().AddDate(0, -3, 0) // Last 3 months
	endDate := time.Now()

	perfReport, err := reportService.GeneratePerformanceReport(accountID, startDate, endDate)
	if err == nil {
		log.Println("")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Println("  PERFORMANCE REPORT (Last 3 Months)")
		log.Println("═══════════════════════════════════════════════════════════")
		log.Printf("  Total Trades:      %d", perfReport.TotalTrades)
		log.Printf("  Win Rate:          %.2f%%", perfReport.WinRate)
		log.Printf("  Net Profit:        $%.2f", perfReport.NetProfit)
		log.Printf("  Profit Factor:     %.2f", perfReport.ProfitFactor)
		log.Printf("  Sharpe Ratio:      %.2f", perfReport.SharpeRatio)
		log.Printf("  Sortino Ratio:     %.2f", perfReport.SortinoRatio)
		log.Printf("  Average MAE:       $%.2f", perfReport.AverageMAE)
		log.Printf("  Average MFE:       $%.2f", perfReport.AverageMFE)
		log.Printf("  Avg R-Multiple:    %.2f", perfReport.AverageRMultiple)
		log.Println("═══════════════════════════════════════════════════════════")
	}
}
