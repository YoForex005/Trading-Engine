package bbook

import (
	"context"
	"fmt"

	"github.com/govalues/decimal"
	decutil "github.com/epic1st/rtx/backend/internal/decimal"
	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// ValidateMarginRequirement checks if opening this order would breach margin requirements.
// Returns error if insufficient margin, nil if order can proceed.
func ValidateMarginRequirement(
	ctx context.Context,
	accountID int64,
	symbol string,
	volume decimal.Decimal,
	side string, // "BUY" or "SELL"
	currentPrice decimal.Decimal,
	currentEquity decimal.Decimal,
	currentUsedMargin decimal.Decimal,
	symbolMarginConfigRepo *repository.SymbolMarginConfigRepository,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	// 1. Get symbol margin configuration
	symbolConfig, err := symbolMarginConfigRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get symbol config: %w", err)
	}

	// 2. Calculate required margin for new position
	requiredMargin := CalculatePositionMargin(
		volume,
		decutil.MustParse(symbolConfig.ContractSize),
		currentPrice,
		decutil.MustParse(symbolConfig.MaxLeverage),
	)

	// 3. Calculate projected used margin (current + new position)
	projectedUsedMargin, err := currentUsedMargin.Add(requiredMargin)
	if err != nil {
		return fmt.Errorf("failed to calculate projected used margin: %w", err)
	}

	// 4. Calculate projected margin level
	projectedMarginLevel := CalculateMarginLevel(currentEquity, projectedUsedMargin)

	// 5. Get account risk limits
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		// Use default margin call level if no limits configured
		limits = &repository.RiskLimit{
			MarginCallLevel: "100.00",
		}
	}

	// 6. Check if projected margin level would trigger margin call
	marginCallLevel := decutil.MustParse(limits.MarginCallLevel)
	if projectedMarginLevel.Cmp(marginCallLevel) <= 0 {
		return fmt.Errorf(
			"insufficient margin: projected margin level %s%% would breach threshold %s%%",
			decutil.ToStringFixed(projectedMarginLevel, 2),
			decutil.ToStringFixed(marginCallLevel, 2),
		)
	}

	// 7. Check if free margin is sufficient (equity - used_margin > required_margin)
	freeMargin, err := currentEquity.Sub(currentUsedMargin)
	if err != nil {
		return fmt.Errorf("failed to calculate free margin: %w", err)
	}

	if freeMargin.Cmp(requiredMargin) < 0 {
		return fmt.Errorf(
			"insufficient free margin: required %s, available %s",
			decutil.ToStringFixed(requiredMargin, 2),
			decutil.ToStringFixed(freeMargin, 2),
		)
	}

	return nil // Validation passed
}

// ValidatePositionLimits checks if new order would breach position count or size limits.
// Returns error if limits exceeded, nil if order can proceed.
func ValidatePositionLimits(
	ctx context.Context,
	accountID int64,
	symbol string,
	volume decimal.Decimal,
	currentOpenPositions int,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	// Get account limits
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		// Use defaults if not configured
		limits = &repository.RiskLimit{
			MaxOpenPositions: 50,
		}
	}

	// Check max open positions
	if currentOpenPositions >= limits.MaxOpenPositions {
		return fmt.Errorf(
			"max open positions exceeded: %d/%d",
			currentOpenPositions,
			limits.MaxOpenPositions,
		)
	}

	// Check max position size (if configured)
	if limits.MaxPositionSizeLots != nil {
		maxSize := decutil.MustParse(*limits.MaxPositionSizeLots)
		if volume.Cmp(maxSize) > 0 {
			return fmt.Errorf(
				"position size %s lots exceeds limit %s lots",
				decutil.ToStringFixed(volume, 2),
				decutil.ToStringFixed(maxSize, 2),
			)
		}
	}

	return nil
}

// ValidateLeverage checks if requested leverage exceeds symbol or account limits.
// Returns error if leverage too high, nil if acceptable.
func ValidateLeverage(
	ctx context.Context,
	accountID int64,
	symbol string,
	requestedLeverage decimal.Decimal,
	symbolMarginConfigRepo *repository.SymbolMarginConfigRepository,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	// Get symbol max leverage (ESMA regulatory limit)
	symbolConfig, err := symbolMarginConfigRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get symbol config: %w", err)
	}

	symbolMaxLeverage := decutil.MustParse(symbolConfig.MaxLeverage)
	if requestedLeverage.Cmp(symbolMaxLeverage) > 0 {
		return fmt.Errorf(
			"leverage %s exceeds symbol limit %s (ESMA regulatory requirement)",
			decutil.ToStringFixed(requestedLeverage, 2),
			decutil.ToStringFixed(symbolMaxLeverage, 2),
		)
	}

	// Get account max leverage
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err == nil && limits.MaxLeverage != "" {
		accountMaxLeverage := decutil.MustParse(limits.MaxLeverage)
		if requestedLeverage.Cmp(accountMaxLeverage) > 0 {
			return fmt.Errorf(
				"leverage %s exceeds account limit %s",
				decutil.ToStringFixed(requestedLeverage, 2),
				decutil.ToStringFixed(accountMaxLeverage, 2),
			)
		}
	}

	return nil
}

// ValidateSymbolExposure checks if opening this order would create excessive exposure to a single symbol.
// Prevents concentration risk (e.g., account has 80% of equity in EURUSD).
func ValidateSymbolExposure(
	ctx context.Context,
	accountID int64,
	symbol string,
	newVolume decimal.Decimal,
	currentPrice decimal.Decimal,
	currentEquity decimal.Decimal,
	existingPositions []*Position,
	symbolMarginConfigRepo *repository.SymbolMarginConfigRepository,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	// Get symbol config
	symbolConfig, err := symbolMarginConfigRepo.GetBySymbol(ctx, symbol)
	if err != nil {
		return fmt.Errorf("failed to get symbol config: %w", err)
	}

	// Calculate total volume for this symbol (existing + new)
	totalVolume := newVolume
	for _, pos := range existingPositions {
		if pos.Symbol == symbol && pos.Status == "OPEN" {
			totalVolume, err = totalVolume.Add(decutil.NewFromFloat64(pos.Volume))
			if err != nil {
				return fmt.Errorf("failed to calculate total volume: %w", err)
			}
		}
	}

	// Calculate total exposure value (volume * contract_size * price)
	contractSize := decutil.MustParse(symbolConfig.ContractSize)
	totalExposure, err := totalVolume.Mul(contractSize)
	if err != nil {
		return fmt.Errorf("failed to calculate exposure: %w", err)
	}
	totalExposure, err = totalExposure.Mul(currentPrice)
	if err != nil {
		return fmt.Errorf("failed to calculate exposure: %w", err)
	}

	// Calculate exposure as percentage of equity
	exposurePercent, err := totalExposure.Quo(currentEquity)
	if err != nil {
		return fmt.Errorf("failed to calculate exposure percentage: %w", err)
	}
	exposurePercent, err = exposurePercent.Mul(decutil.NewFromInt64(100))
	if err != nil {
		return fmt.Errorf("failed to calculate exposure percentage: %w", err)
	}

	// Get risk limits
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		// Use default: max 40% exposure per symbol
		maxExposurePercent := decutil.MustParse("40.00")
		if exposurePercent.Cmp(maxExposurePercent) > 0 {
			return fmt.Errorf(
				"symbol exposure %s%% exceeds default limit %s%%",
				decutil.ToStringFixed(exposurePercent, 2),
				decutil.ToStringFixed(maxExposurePercent, 2),
			)
		}
		return nil
	}

	// Check account-specific exposure limit (if configured)
	if limits.MaxSymbolExposurePct != nil {
		maxExposure := decutil.MustParse(*limits.MaxSymbolExposurePct)
		if exposurePercent.Cmp(maxExposure) > 0 {
			return fmt.Errorf(
				"symbol exposure %s%% exceeds account limit %s%%",
				decutil.ToStringFixed(exposurePercent, 2),
				decutil.ToStringFixed(maxExposure, 2),
			)
		}
	}

	return nil
}

// ValidateTotalExposure checks if account's total market exposure is within limits.
// Calculates sum of all open position notional values vs equity.
func ValidateTotalExposure(
	ctx context.Context,
	accountID int64,
	currentEquity decimal.Decimal,
	existingPositions []*Position,
	riskLimitRepo *repository.RiskLimitRepository,
) error {
	// Calculate total notional exposure across all positions
	totalExposure := decutil.Zero()
	for _, pos := range existingPositions {
		if pos.Status != "OPEN" {
			continue
		}

		// Notional value = volume * contract_size * current_price
		volume := decutil.NewFromFloat64(pos.Volume)
		price := decutil.NewFromFloat64(pos.CurrentPrice)
		contractSize := decutil.MustParse("100000") // TODO: get from symbol config

		notional, err := volume.Mul(contractSize)
		if err != nil {
			return fmt.Errorf("failed to calculate notional: %w", err)
		}
		notional, err = notional.Mul(price)
		if err != nil {
			return fmt.Errorf("failed to calculate notional: %w", err)
		}
		totalExposure, err = totalExposure.Add(notional)
		if err != nil {
			return fmt.Errorf("failed to calculate total exposure: %w", err)
		}
	}

	// Calculate exposure as percentage of equity
	exposurePercent, err := totalExposure.Quo(currentEquity)
	if err != nil {
		return fmt.Errorf("failed to calculate exposure percentage: %w", err)
	}
	exposurePercent, err = exposurePercent.Mul(decutil.NewFromInt64(100))
	if err != nil {
		return fmt.Errorf("failed to calculate exposure percentage: %w", err)
	}

	// Get limits
	limits, err := riskLimitRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		// Use default: max 300% total exposure (3x equity with margin)
		maxExposure := decutil.MustParse("300.00")
		if exposurePercent.Cmp(maxExposure) > 0 {
			return fmt.Errorf(
				"total exposure %s%% exceeds default limit %s%%",
				decutil.ToStringFixed(exposurePercent, 2),
				decutil.ToStringFixed(maxExposure, 2),
			)
		}
		return nil
	}

	// Check account-specific total exposure limit
	if limits.MaxTotalExposurePct != nil {
		maxExposure := decutil.MustParse(*limits.MaxTotalExposurePct)
		if exposurePercent.Cmp(maxExposure) > 0 {
			return fmt.Errorf(
				"total exposure %s%% exceeds account limit %s%%",
				decutil.ToStringFixed(exposurePercent, 2),
				decutil.ToStringFixed(maxExposure, 2),
			)
		}
	}

	return nil
}
