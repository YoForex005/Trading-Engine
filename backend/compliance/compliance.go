package compliance

import (
	"github.com/epic1st/rtx/backend/compliance/handlers"
	"github.com/epic1st/rtx/backend/compliance/models"
	"github.com/epic1st/rtx/backend/compliance/repository"
	"github.com/epic1st/rtx/backend/compliance/services"
)

// ComplianceSystem coordinates all compliance and regulatory services
type ComplianceSystem struct {
	Repository         *repository.ComplianceRepository
	TransactionService *services.TransactionReportingService
	KYCAMLService      *services.KYCAMLService
	AuditService       *services.AuditTrailService
	BestExecService    *services.BestExecutionService
	LeverageService    *services.LeverageLimitsService
	Handler            *handlers.ComplianceHandler
}

// NewComplianceSystem initializes the complete compliance system
func NewComplianceSystem() *ComplianceSystem {
	// Initialize repository
	repo := repository.NewComplianceRepository()

	// Initialize services
	transactionService := services.NewTransactionReportingService(repo)
	kycService := services.NewKYCAMLService(repo)
	auditService := services.NewAuditTrailService(repo)
	bestExecService := services.NewBestExecutionService(repo)
	leverageService := services.NewLeverageLimitsService(repo)

	// Initialize HTTP handler
	handler := handlers.NewComplianceHandler(
		transactionService,
		kycService,
		auditService,
		bestExecService,
		leverageService,
	)

	return &ComplianceSystem{
		Repository:         repo,
		TransactionService: transactionService,
		KYCAMLService:      kycService,
		AuditService:       auditService,
		BestExecService:    bestExecService,
		LeverageService:    leverageService,
		Handler:            handler,
	}
}

// OnOrderPlaced hooks into order placement for compliance tracking
func (cs *ComplianceSystem) OnOrderPlaced(userID, clientID, orderID, symbol string, orderData interface{}, ip, ua string) {
	// Log audit trail
	cs.AuditService.LogOrderPlaced(userID, clientID, orderID, symbol, orderData, ip, ua)
}

// OnTradeExecuted hooks into trade execution for reporting and tracking
func (cs *ComplianceSystem) OnTradeExecuted(
	userID, clientID, orderID, tradeID, symbol string,
	tradeData interface{},
	executedPrice, quantity float64,
	lpName string,
	jurisdiction string,
	clientClass string,
	ip, ua string,
) {
	// Log audit trail
	cs.AuditService.LogTradeExecuted(userID, clientID, orderID, tradeID, symbol, tradeData, ip, ua)

	// Create transaction report if required
	if requiresReporting(jurisdiction) {
		cs.TransactionService.CreateTransactionReport(
			getJurisdiction(jurisdiction),
			orderID,
			tradeID,
			clientID,
			symbol,
			getSide(tradeData),
			quantity,
			executedPrice,
			lpName,
			getClientClass(clientClass),
		)
	}

	// Track execution metrics for best execution reporting
	cs.BestExecService.TrackExecution(
		orderID,
		symbol,
		"OTC",
		lpName,
		getRequestedPrice(tradeData),
		executedPrice,
		quantity,
		getLatency(tradeData),
		getFillType(tradeData),
	)
}

// OnPositionClosed hooks into position closure
func (cs *ComplianceSystem) OnPositionClosed(userID, clientID, tradeID, symbol string, positionData interface{}, ip, ua string) {
	cs.AuditService.LogPositionClosed(userID, clientID, tradeID, symbol, positionData, ip, ua)
}

// OnWithdrawal hooks into withdrawal requests
func (cs *ComplianceSystem) OnWithdrawal(userID, clientID string, amount float64, ip, ua string) {
	cs.AuditService.LogWithdrawal(userID, clientID, amount, ip, ua)

	// Monitor for AML - large withdrawals
	if amount >= 10000 {
		cs.KYCAMLService.CreateAMLAlert(
			clientID,
			"LARGE_WITHDRAWAL",
			"Large withdrawal detected",
			"MEDIUM",
		)
	}
}

// OnDeposit hooks into deposits
func (cs *ComplianceSystem) OnDeposit(userID, clientID string, amount float64, ip, ua string) {
	cs.AuditService.LogDeposit(userID, clientID, amount, ip, ua)

	// Monitor for AML - large deposits
	cs.KYCAMLService.MonitorTransactionPatterns(clientID, []string{}, amount)
}

// OnAccountModification hooks into account changes
func (cs *ComplianceSystem) OnAccountModification(userID, clientID, action string, before, after interface{}, ip, ua string) {
	cs.AuditService.LogAccountModification(userID, clientID, action, before, after, ip, ua)
}

// ValidateLeverageCompliance validates leverage against regulatory limits
func (cs *ComplianceSystem) ValidateLeverageCompliance(
	jurisdiction, clientClass, symbol, instrumentClass string,
	requestedLeverage int,
) (bool, int, string, error) {

	isValid, maxLeverage, err := cs.LeverageService.ValidateLeverage(
		getJurisdiction(jurisdiction),
		getClientClass(clientClass),
		symbol,
		instrumentClass,
		requestedLeverage,
	)

	warning := cs.LeverageService.DisplayLeverageWarning(
		getClientClass(clientClass),
		requestedLeverage,
		"en",
	)

	return isValid, maxLeverage, warning, err
}

// Helper functions for type conversions

func requiresReporting(jurisdiction string) bool {
	return jurisdiction == "EU" || jurisdiction == "UK" || jurisdiction == "US"
}

func getJurisdiction(j string) models.Jurisdiction {
	switch j {
	case "EU":
		return models.JurisdictionEU
	case "UK":
		return models.JurisdictionUK
	case "US":
		return models.JurisdictionUS
	default:
		return models.JurisdictionEU
	}
}

func getClientClass(c string) models.ClientClassification {
	switch c {
	case "RETAIL":
		return models.ClientRetail
	case "PROFESSIONAL":
		return models.ClientProfessional
	default:
		return models.ClientRetail
	}
}

func getSide(tradeData interface{}) string {
	// Extract side from trade data
	if data, ok := tradeData.(map[string]interface{}); ok {
		if side, ok := data["side"].(string); ok {
			return side
		}
	}
	return "BUY"
}

func getRequestedPrice(tradeData interface{}) float64 {
	if data, ok := tradeData.(map[string]interface{}); ok {
		if price, ok := data["requestedPrice"].(float64); ok {
			return price
		}
	}
	return 0
}

func getLatency(tradeData interface{}) int64 {
	if data, ok := tradeData.(map[string]interface{}); ok {
		if latency, ok := data["latency"].(int64); ok {
			return latency
		}
	}
	return 0
}

func getFillType(tradeData interface{}) string {
	if data, ok := tradeData.(map[string]interface{}); ok {
		if fillType, ok := data["fillType"].(string); ok {
			return fillType
		}
	}
	return "AGGRESSIVE"
}
