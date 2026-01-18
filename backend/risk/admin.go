package risk

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"
)

// AdminController provides admin controls for risk management
type AdminController struct {
	engine *Engine
}

// NewAdminController creates a new admin controller
func NewAdminController(engine *Engine) *AdminController {
	return &AdminController{
		engine: engine,
	}
}

// SetClientRiskProfile sets or updates a client's risk profile
func (ac *AdminController) SetClientRiskProfile(profile *ClientRiskProfile) error {
	if profile.ClientID == "" {
		return fmt.Errorf("client ID required")
	}

	// Validate profile
	if profile.StopOutLevel >= profile.MarginCallLevel {
		return fmt.Errorf("stop-out level must be below margin call level")
	}

	if profile.MaxLeverage < 1 || profile.MaxLeverage > 500 {
		return fmt.Errorf("invalid leverage: %d (must be 1-500)", profile.MaxLeverage)
	}

	ac.engine.SetClientRiskProfile(profile)

	log.Printf("[Admin] Set risk profile for client %s: tier=%s, leverage=%d, stopout=%.2f%%",
		profile.ClientID, profile.RiskTier, profile.MaxLeverage, profile.StopOutLevel)

	return nil
}

// SetInstrumentRiskParams sets risk parameters for an instrument
func (ac *AdminController) SetInstrumentRiskParams(params *InstrumentRiskParams) error {
	if params.Symbol == "" {
		return fmt.Errorf("symbol required")
	}

	ac.engine.SetInstrumentRiskParams(params)

	log.Printf("[Admin] Set instrument params for %s: leverage=%d, margin=%.2f%%, maxSize=%.2f",
		params.Symbol, params.MaxLeverage, params.MarginRequirement, params.MaxPositionSize)

	return nil
}

// ForceClosePosition forcibly closes a position (admin override)
func (ac *AdminController) ForceClosePosition(
	positionID int64,
	reason string,
	adminID string,
) error {

	position := ac.engine.GetPosition(positionID)
	if position == nil {
		return fmt.Errorf("position %d not found", positionID)
	}

	// Get close price
	bid, ask := ac.engine.GetCurrentPrice(position.Symbol)
	closePrice := bid
	if position.Side == "SELL" {
		closePrice = ask
	}

	err := ac.engine.ClosePosition(positionID, closePrice, fmt.Sprintf("ADMIN_CLOSE: %s", reason))
	if err != nil {
		return err
	}

	log.Printf("[Admin] Position %d force-closed by %s: %s", positionID, adminID, reason)

	// Create alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_ADMIN_CLOSE_%d", time.Now().Unix()),
		AccountID: position.AccountID,
		Symbol:    position.Symbol,
		AlertType: "ADMIN_FORCE_CLOSE",
		Severity:  RiskLevelHigh,
		Message:   fmt.Sprintf("Position %d force-closed by admin %s: %s", positionID, adminID, reason),
		Data: map[string]interface{}{
			"admin_id":    adminID,
			"position_id": positionID,
			"reason":      reason,
		},
		CreatedAt: time.Now(),
	}
	ac.engine.StoreAlert(alert)

	return nil
}

// ForceCloseAllPositions closes all positions for an account
func (ac *AdminController) ForceCloseAllPositions(
	accountID int64,
	reason string,
	adminID string,
) (int, error) {

	positions := ac.engine.GetAllPositions(accountID)
	closed := 0

	for _, pos := range positions {
		err := ac.ForceClosePosition(pos.ID, reason, adminID)
		if err != nil {
			log.Printf("[Admin] Failed to close position %d: %v", pos.ID, err)
			continue
		}
		closed++
	}

	log.Printf("[Admin] Force-closed %d/%d positions for account %d by %s",
		closed, len(positions), accountID, adminID)

	return closed, nil
}

// BlockClient prevents a client from trading
func (ac *AdminController) BlockClient(
	clientID string,
	reason string,
	adminID string,
) error {

	profile := ac.engine.GetClientRiskProfile(clientID)
	if profile == nil {
		return fmt.Errorf("client profile not found")
	}

	// Set all limits to zero to block trading
	profile.MaxPositions = 0
	profile.MaxExposurePercent = 0
	profile.DailyLossLimit = 0

	ac.engine.SetClientRiskProfile(profile)

	log.Printf("[Admin] Client %s blocked by %s: %s", clientID, adminID, reason)

	// Create alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_CLIENT_BLOCK_%s", clientID),
		AlertType: "CLIENT_BLOCKED",
		Severity:  RiskLevelCritical,
		Message:   fmt.Sprintf("Client %s blocked by admin %s: %s", clientID, adminID, reason),
		Data: map[string]interface{}{
			"admin_id":  adminID,
			"client_id": clientID,
			"reason":    reason,
		},
		CreatedAt: time.Now(),
	}
	ac.engine.StoreAlert(alert)

	return nil
}

// AdjustLeverage adjusts leverage for a client
func (ac *AdminController) AdjustLeverage(
	clientID string,
	newLeverage int,
	adminID string,
) error {

	if newLeverage < 1 || newLeverage > 500 {
		return fmt.Errorf("invalid leverage: %d", newLeverage)
	}

	profile := ac.engine.GetClientRiskProfile(clientID)
	if profile == nil {
		return fmt.Errorf("client profile not found")
	}

	oldLeverage := profile.MaxLeverage
	profile.MaxLeverage = newLeverage

	ac.engine.SetClientRiskProfile(profile)

	log.Printf("[Admin] Leverage adjusted for client %s: %d -> %d by %s",
		clientID, oldLeverage, newLeverage, adminID)

	return nil
}

// SetDailyLossLimit sets daily loss limit for a client
func (ac *AdminController) SetDailyLossLimit(
	clientID string,
	lossLimit float64,
	adminID string,
) error {

	if lossLimit < 0 {
		return fmt.Errorf("loss limit must be positive")
	}

	profile := ac.engine.GetClientRiskProfile(clientID)
	if profile == nil {
		return fmt.Errorf("client profile not found")
	}

	profile.DailyLossLimit = lossLimit

	ac.engine.SetClientRiskProfile(profile)

	log.Printf("[Admin] Daily loss limit set for client %s: %.2f by %s",
		clientID, lossLimit, adminID)

	return nil
}

// EnableInstrument enables trading for an instrument
func (ac *AdminController) EnableInstrument(symbol string, adminID string) error {
	params := ac.engine.GetInstrumentRiskParams(symbol)
	if params == nil {
		return fmt.Errorf("instrument %s not configured", symbol)
	}

	params.AllowNewPositions = true
	ac.engine.SetInstrumentRiskParams(params)

	log.Printf("[Admin] Instrument %s enabled by %s", symbol, adminID)
	return nil
}

// DisableInstrument disables trading for an instrument
func (ac *AdminController) DisableInstrument(
	symbol string,
	reason string,
	adminID string,
) error {

	params := ac.engine.GetInstrumentRiskParams(symbol)
	if params == nil {
		return fmt.Errorf("instrument %s not configured", symbol)
	}

	params.AllowNewPositions = false
	ac.engine.SetInstrumentRiskParams(params)

	log.Printf("[Admin] Instrument %s disabled by %s: %s", symbol, adminID, reason)

	// Create alert
	alert := &RiskAlert{
		ID:        fmt.Sprintf("ALERT_INSTR_DISABLE_%s", symbol),
		Symbol:    symbol,
		AlertType: "INSTRUMENT_DISABLED",
		Severity:  RiskLevelMedium,
		Message:   fmt.Sprintf("Instrument %s disabled by admin %s: %s", symbol, adminID, reason),
		Data: map[string]interface{}{
			"admin_id": adminID,
			"reason":   reason,
		},
		CreatedAt: time.Now(),
	}
	ac.engine.StoreAlert(alert)

	return nil
}

// GetRiskDashboard returns real-time risk dashboard data
func (ac *AdminController) GetRiskDashboard() map[string]interface{} {
	accounts := ac.engine.GetAllAccounts()

	totalEquity := 0.0
	totalMargin := 0.0
	totalExposure := 0.0
	activePositions := 0
	marginCallsToday := 0
	liquidationsToday := 0

	riskDistribution := make(map[RiskLevel]int)
	topRiskyAccounts := make([]RiskyAccountSummary, 0)

	for _, account := range accounts {
		totalEquity += account.Equity
		totalMargin += account.Margin

		positions := ac.engine.GetAllPositions(account.ID)
		activePositions += len(positions)

		// Calculate exposure
		for _, pos := range positions {
			contractSize := ac.getContractSize(pos.Symbol)
			notional := pos.Volume * contractSize * pos.CurrentPrice
			totalExposure += notional
		}

		// Check risk level
		riskLevel := ac.calculateAccountRiskLevel(account)
		riskDistribution[riskLevel]++

		// Identify risky accounts
		if riskLevel == RiskLevelHigh || riskLevel == RiskLevelCritical {
			summary := ac.getAccountRiskSummary(account)
			topRiskyAccounts = append(topRiskyAccounts, summary)
		}
	}

	// Get margin calls and liquidations
	marginCallsToday = ac.engine.GetTodayMarginCallCount()
	liquidationsToday = ac.engine.GetTodayLiquidationCount()

	// Get exposure by symbol
	exposureMonitor := NewExposureMonitor(ac.engine)
	aggregateExposure := exposureMonitor.GetAggregateExposure()

	// Get active circuit breakers
	cbManager := ac.engine.circuitBreakerManager
	activeBreakers := cbManager.GetActiveBreakers()

	// Get recent alerts
	alerts := ac.engine.GetRecentAlerts(20)

	avgMarginLevel := 0.0
	if totalMargin > 0 {
		avgMarginLevel = (totalEquity / totalMargin) * 100
	}

	return map[string]interface{}{
		"timestamp":           time.Now(),
		"accounts":            len(accounts),
		"total_equity":        totalEquity,
		"total_margin":        totalMargin,
		"total_exposure":      totalExposure,
		"active_positions":    activePositions,
		"margin_calls_today":  marginCallsToday,
		"liquidations_today":  liquidationsToday,
		"avg_margin_level":    avgMarginLevel,
		"exposure_by_symbol":  aggregateExposure,
		"risk_distribution":   riskDistribution,
		"top_risky_accounts":  topRiskyAccounts,
		"active_breakers":     activeBreakers,
		"recent_alerts":       alerts,
	}
}

// calculateAccountRiskLevel determines risk level for an account
func (ac *AdminController) calculateAccountRiskLevel(account *Account) RiskLevel {
	profile := ac.engine.GetClientRiskProfile(account.ClientID)
	if profile == nil {
		return RiskLevelNone
	}

	// Check margin level
	if account.MarginLevel > 0 {
		if account.MarginLevel < profile.StopOutLevel {
			return RiskLevelCritical
		}
		if account.MarginLevel < profile.MarginCallLevel {
			return RiskLevelHigh
		}
		if account.MarginLevel < profile.MarginCallLevel*1.5 {
			return RiskLevelMedium
		}
	}

	// Check daily loss
	dailyPnL := ac.engine.GetDailyPnL(account.ID)
	if dailyPnL < 0 {
		lossPercent := (dailyPnL / account.Balance) * 100
		if lossPercent < -10 {
			return RiskLevelHigh
		}
		if lossPercent < -5 {
			return RiskLevelMedium
		}
	}

	return RiskLevelLow
}

// getAccountRiskSummary creates a risk summary for an account
func (ac *AdminController) getAccountRiskSummary(account *Account) RiskyAccountSummary {
	dailyPnL := ac.engine.GetDailyPnL(account.ID)

	exposureMonitor := NewExposureMonitor(ac.engine)
	metrics, _ := exposureMonitor.CalculateExposure(account.ID)

	exposure := 0.0
	if metrics != nil {
		exposure = metrics.TotalExposure
	}

	// Calculate risk score (0-100)
	riskScore := 0.0

	// Margin level component (40 points)
	profile := ac.engine.GetClientRiskProfile(account.ClientID)
	if profile != nil && account.MarginLevel > 0 {
		marginRisk := math.Max(0, (profile.MarginCallLevel - account.MarginLevel) / profile.MarginCallLevel)
		riskScore += marginRisk * 40
	}

	// Daily loss component (30 points)
	if dailyPnL < 0 && account.Balance > 0 {
		lossPercent := math.Abs(dailyPnL / account.Balance)
		riskScore += math.Min(lossPercent * 300, 30) // Cap at 30
	}

	// Exposure component (30 points)
	if account.Equity > 0 {
		exposureRatio := exposure / account.Equity
		riskScore += math.Min(exposureRatio * 10, 30) // Cap at 30
	}

	// Collect risk flags
	flags := make([]string, 0)
	if account.MarginLevel < profile.StopOutLevel {
		flags = append(flags, "NEAR_STOPOUT")
	}
	if dailyPnL < 0 && math.Abs(dailyPnL) > profile.DailyLossLimit*0.8 {
		flags = append(flags, "APPROACHING_DAILY_LIMIT")
	}
	if metrics != nil && metrics.ConcentrationRisk > 0.7 {
		flags = append(flags, "HIGH_CONCENTRATION")
	}
	if metrics != nil && metrics.LiquidityScore < 0.5 {
		flags = append(flags, "LOW_LIQUIDITY")
	}

	return RiskyAccountSummary{
		AccountID:   account.ID,
		MarginLevel: account.MarginLevel,
		DailyPnL:    dailyPnL,
		Exposure:    exposure,
		RiskScore:   riskScore,
		Flags:       flags,
	}
}

// getContractSize returns contract size for an instrument
func (ac *AdminController) getContractSize(symbol string) float64 {
	switch symbol {
	case "XAUUSD":
		return 100.0
	case "XAGUSD":
		return 5000.0
	case "BTCUSD", "ETHUSD":
		return 1.0
	default:
		if len(symbol) == 6 {
			return 100000.0
		}
		return 1.0
	}
}

// GenerateRiskReport generates a comprehensive risk report
func (ac *AdminController) GenerateRiskReport(reportType string) (*RiskReport, error) {
	accounts := ac.engine.GetAllAccounts()

	// Calculate aggregate metrics
	totalEquity := 0.0
	totalMargin := 0.0
	totalExposure := 0.0
	activePositions := 0

	exposureBySymbol := make(map[string]float64)
	riskDistribution := make(map[RiskLevel]int)
	topRiskyAccounts := make([]RiskyAccountSummary, 0)

	for _, account := range accounts {
		totalEquity += account.Equity
		totalMargin += account.Margin

		positions := ac.engine.GetAllPositions(account.ID)
		activePositions += len(positions)

		// Aggregate exposure
		for _, pos := range positions {
			contractSize := ac.getContractSize(pos.Symbol)
			notional := pos.Volume * contractSize * pos.CurrentPrice
			totalExposure += notional

			if pos.Side == "BUY" {
				exposureBySymbol[pos.Symbol] += notional
			} else {
				exposureBySymbol[pos.Symbol] -= notional
			}
		}

		// Risk distribution
		riskLevel := ac.calculateAccountRiskLevel(account)
		riskDistribution[riskLevel]++

		// Top risky accounts
		if riskLevel == RiskLevelHigh || riskLevel == RiskLevelCritical {
			summary := ac.getAccountRiskSummary(account)
			topRiskyAccounts = append(topRiskyAccounts, summary)
		}
	}

	// Sort risky accounts by risk score
	sort.Slice(topRiskyAccounts, func(i, j int) bool {
		return topRiskyAccounts[i].RiskScore > topRiskyAccounts[j].RiskScore
	})

	// Limit to top 20
	if len(topRiskyAccounts) > 20 {
		topRiskyAccounts = topRiskyAccounts[:20]
	}

	avgMarginLevel := 0.0
	if totalMargin > 0 {
		avgMarginLevel = (totalEquity / totalMargin) * 100
	}

	// Get alerts
	alerts := ac.engine.GetRecentAlerts(100)

	report := &RiskReport{
		ReportID:          fmt.Sprintf("RISK_%s_%d", reportType, time.Now().Unix()),
		GeneratedAt:       time.Now(),
		ReportType:        reportType,
		AccountCount:      len(accounts),
		TotalEquity:       totalEquity,
		TotalMargin:       totalMargin,
		TotalExposure:     totalExposure,
		ActivePositions:   activePositions,
		MarginCallsToday:  ac.engine.GetTodayMarginCallCount(),
		LiquidationsToday: ac.engine.GetTodayLiquidationCount(),
		AverageMarginLevel: avgMarginLevel,
		ExposureBySymbol:  exposureBySymbol,
		RiskDistribution:  riskDistribution,
		TopRiskyAccounts:  topRiskyAccounts,
		Alerts:            alerts,
	}

	log.Printf("[Admin] Generated %s risk report: %d accounts, %.2f total equity",
		reportType, report.AccountCount, report.TotalEquity)

	return report, nil
}
