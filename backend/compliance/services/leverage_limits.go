package services

import (
	"fmt"

	"github.com/epic1st/rtx/backend/compliance/models"
	"github.com/epic1st/rtx/backend/compliance/repository"
)

// LeverageLimitsService enforces regulatory leverage restrictions
type LeverageLimitsService struct {
	repo *repository.ComplianceRepository
}

func NewLeverageLimitsService(repo *repository.ComplianceRepository) *LeverageLimitsService {
	return &LeverageLimitsService{
		repo: repo,
	}
}

// ValidateLeverage checks if requested leverage is within regulatory limits
func (s *LeverageLimitsService) ValidateLeverage(
	jurisdiction models.Jurisdiction,
	clientClass models.ClientClassification,
	symbol, instrumentClass string,
	requestedLeverage int,
) (bool, int, error) {

	// Get applicable leverage limit
	limit, err := s.repo.GetLeverageLimit(jurisdiction, instrumentClass, clientClass)
	if err != nil {
		return false, 0, err
	}

	if limit == nil {
		// No specific limit found, use default
		limit = s.getDefaultLimit(jurisdiction, clientClass)
	}

	isValid := requestedLeverage <= limit.MaxLeverage

	return isValid, limit.MaxLeverage, nil
}

// EnforceLeverage enforces leverage limit and returns allowed leverage
func (s *LeverageLimitsService) EnforceLeverage(
	jurisdiction models.Jurisdiction,
	clientClass models.ClientClassification,
	symbol, instrumentClass string,
	requestedLeverage int,
) (int, error) {

	valid, maxLeverage, err := s.ValidateLeverage(jurisdiction, clientClass, symbol, instrumentClass, requestedLeverage)
	if err != nil {
		return 0, err
	}

	if !valid {
		// Auto-cap to max allowed
		return maxLeverage, fmt.Errorf("leverage capped to %d:1 per %s regulations", maxLeverage, jurisdiction)
	}

	return requestedLeverage, nil
}

// GetESMALimits returns ESMA leverage limits (EU standard)
func (s *LeverageLimitsService) GetESMALimits(clientClass models.ClientClassification) map[string]int {
	if clientClass != models.ClientRetail {
		// Professional clients have no leverage restrictions
		return map[string]int{
			"MAJOR_PAIRS":      500,
			"MINOR_PAIRS":      500,
			"GOLD":             500,
			"INDICES":          500,
			"COMMODITIES":      500,
			"CRYPTOCURRENCIES": 500,
		}
	}

	// ESMA retail client limits
	return map[string]int{
		"MAJOR_PAIRS":      30, // EUR/USD, GBP/USD, USD/JPY, etc.
		"MINOR_PAIRS":      20, // Other currency pairs
		"GOLD":             20,
		"INDICES":          20,
		"COMMODITIES":      10,
		"CRYPTOCURRENCIES": 2,
	}
}

// CalculateRequiredMargin calculates required margin with leverage limits
func (s *LeverageLimitsService) CalculateRequiredMargin(
	jurisdiction models.Jurisdiction,
	clientClass models.ClientClassification,
	symbol, instrumentClass string,
	notionalValue float64,
	requestedLeverage int,
) (float64, int, error) {

	// Enforce leverage limit
	allowedLeverage, err := s.EnforceLeverage(jurisdiction, clientClass, symbol, instrumentClass, requestedLeverage)
	if err != nil {
		// Return calculated margin with warning
		requiredMargin := notionalValue / float64(allowedLeverage)
		return requiredMargin, allowedLeverage, err
	}

	requiredMargin := notionalValue / float64(allowedLeverage)
	return requiredMargin, allowedLeverage, nil
}

// CheckNegativeBalanceProtection ensures negative balance protection is enabled
func (s *LeverageLimitsService) CheckNegativeBalanceProtection(clientID string, equity float64) (bool, error) {
	// Negative balance protection required in EU, UK, Australia
	// Stop-out should occur before balance goes negative

	if equity <= 0 {
		return false, fmt.Errorf("negative balance detected for client %s", clientID)
	}

	// Check if approaching negative (margin level < 50%)
	// This would be integrated with margin calculation service
	return true, nil
}

// GetMarginCallLevel returns margin call and stop-out levels
func (s *LeverageLimitsService) GetMarginCallLevel(jurisdiction models.Jurisdiction) (float64, float64) {
	// Margin Call Level and Stop-Out Level
	// Standard industry practice

	switch jurisdiction {
	case models.JurisdictionEU, models.JurisdictionUK:
		return 100.0, 50.0 // Margin call at 100%, stop-out at 50%
	case models.JurisdictionAustralia:
		return 100.0, 50.0
	default:
		return 100.0, 30.0 // More lenient for other jurisdictions
	}
}

// DisplayLeverageWarning generates mandatory leverage warning
func (s *LeverageLimitsService) DisplayLeverageWarning(
	clientClass models.ClientClassification,
	leverage int,
	language string,
) string {

	if clientClass != models.ClientRetail {
		return "" // No warning required for professional clients
	}

	warnings := map[string]string{
		"en": fmt.Sprintf("Trading with leverage of %d:1 can result in significant losses. "+
			"You can lose more than your initial investment. "+
			"Ensure you understand the risks involved.", leverage),
		"de": fmt.Sprintf("Der Handel mit einem Hebel von %d:1 kann zu erheblichen Verlusten führen. "+
			"Sie können mehr als Ihre ursprüngliche Investition verlieren.", leverage),
		"fr": fmt.Sprintf("Le trading avec un effet de levier de %d:1 peut entraîner des pertes importantes. "+
			"Vous pouvez perdre plus que votre investissement initial.", leverage),
	}

	if warning, ok := warnings[language]; ok {
		return warning
	}

	return warnings["en"] // Default to English
}

// getDefaultLimit returns default leverage limit when no specific limit is configured
func (s *LeverageLimitsService) getDefaultLimit(
	jurisdiction models.Jurisdiction,
	clientClass models.ClientClassification,
) *models.LeverageLimit {

	// ESMA limits for EU/UK retail clients
	if (jurisdiction == models.JurisdictionEU || jurisdiction == models.JurisdictionUK) &&
		clientClass == models.ClientRetail {
		return &models.LeverageLimit{
			MaxLeverage: 30, // Most conservative (major pairs)
		}
	}

	// Default for professional or other jurisdictions
	return &models.LeverageLimit{
		MaxLeverage: 100,
	}
}
