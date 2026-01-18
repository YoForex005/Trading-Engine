package risk

import (
	"log"
	"math"
	"time"
)

// ExposureMonitor monitors and calculates exposure metrics
type ExposureMonitor struct {
	engine *Engine
}

// NewExposureMonitor creates a new exposure monitor
func NewExposureMonitor(engine *Engine) *ExposureMonitor {
	return &ExposureMonitor{
		engine: engine,
	}
}

// CalculateExposure calculates comprehensive exposure metrics for an account
func (em *ExposureMonitor) CalculateExposure(accountID int64) (*ExposureMetrics, error) {
	positions := em.engine.GetAllPositions(accountID)

	exposureBySymbol := make(map[string]float64)
	exposureByAsset := make(map[string]float64)

	totalExposure := 0.0
	netExposure := 0.0
	grossExposure := 0.0

	// Calculate per-position exposure
	for _, pos := range positions {
		contractSize := em.getContractSize(pos.Symbol)
		notional := pos.Volume * contractSize * pos.CurrentPrice

		// Net exposure (signed)
		signedNotional := notional
		if pos.Side == "SELL" {
			signedNotional = -notional
		}

		exposureBySymbol[pos.Symbol] += signedNotional
		netExposure += signedNotional
		grossExposure += notional
		totalExposure += notional

		// Categorize by asset class
		assetClass := em.getAssetClass(pos.Symbol)
		exposureByAsset[assetClass] += notional
	}

	// Calculate Greeks (simplified)
	delta := em.calculateDelta(positions)
	gamma := em.calculateGamma(positions)
	vega := em.calculateVega(positions)

	// Calculate concentration risk
	concentrationRisk := em.calculateConcentrationRisk(exposureBySymbol, totalExposure)

	// Calculate correlation risk
	correlationRisk := em.calculateCorrelationRisk(positions)

	// Calculate Value at Risk
	var95, var99 := em.calculateVaR(accountID, positions)

	// Calculate liquidity score
	liquidityScore := em.calculateLiquidityScore(positions)

	metrics := &ExposureMetrics{
		AccountID:         accountID,
		TotalExposure:     totalExposure,
		NetExposure:       netExposure,
		GrossExposure:     grossExposure,
		ExposureBySymbol:  exposureBySymbol,
		ExposureByAsset:   exposureByAsset,
		Delta:             delta,
		Gamma:             gamma,
		Vega:              vega,
		ConcentrationRisk: concentrationRisk,
		CorrelationRisk:   correlationRisk,
		ValueAtRisk95:     var95,
		ValueAtRisk99:     var99,
		LiquidityScore:    liquidityScore,
		Timestamp:         time.Now(),
	}

	return metrics, nil
}

// calculateDelta calculates total delta exposure
func (em *ExposureMonitor) calculateDelta(positions []*Position) float64 {
	totalDelta := 0.0

	for _, pos := range positions {
		// For spot/forex positions, delta = notional
		// For options, delta would be option delta * notional
		contractSize := em.getContractSize(pos.Symbol)
		notional := pos.Volume * contractSize * pos.CurrentPrice

		if pos.Side == "BUY" {
			totalDelta += notional
		} else {
			totalDelta -= notional
		}
	}

	return totalDelta
}

// calculateGamma calculates total gamma exposure
func (em *ExposureMonitor) calculateGamma(positions []*Position) float64 {
	// Simplified gamma calculation
	// For spot positions, gamma ≈ 0
	// For options, gamma = option gamma * position size
	totalGamma := 0.0

	for _, pos := range positions {
		// Estimate gamma based on position size and volatility
		volatility := em.engine.GetVolatility(pos.Symbol)
		contractSize := em.getContractSize(pos.Symbol)
		notional := pos.Volume * contractSize * pos.CurrentPrice

		// Simplified: gamma ≈ notional * (volatility / price)
		if pos.CurrentPrice > 0 {
			gamma := notional * (volatility / pos.CurrentPrice) * 0.01
			totalGamma += gamma
		}
	}

	return totalGamma
}

// calculateVega calculates total vega exposure
func (em *ExposureMonitor) calculateVega(positions []*Position) float64 {
	// Simplified vega calculation
	// Vega measures sensitivity to volatility changes
	totalVega := 0.0

	for _, pos := range positions {
		contractSize := em.getContractSize(pos.Symbol)
		notional := pos.Volume * contractSize * pos.CurrentPrice

		// Simplified: vega ≈ notional * 0.01 (1% sensitivity)
		vega := notional * 0.01
		totalVega += vega
	}

	return totalVega
}

// calculateConcentrationRisk measures how concentrated positions are
func (em *ExposureMonitor) calculateConcentrationRisk(
	exposureBySymbol map[string]float64,
	totalExposure float64,
) float64 {
	if totalExposure == 0 {
		return 0
	}

	// Calculate Herfindahl-Hirschman Index (HHI)
	// HHI = sum of squared market shares
	// 0 = perfect diversification, 1 = fully concentrated
	hhi := 0.0

	for _, exposure := range exposureBySymbol {
		share := math.Abs(exposure) / totalExposure
		hhi += share * share
	}

	// Normalize to 0-1 scale
	// HHI ranges from 1/n (perfectly diversified) to 1 (concentrated)
	numInstruments := float64(len(exposureBySymbol))
	if numInstruments == 0 {
		return 0
	}

	minHHI := 1.0 / numInstruments
	normalizedRisk := (hhi - minHHI) / (1.0 - minHHI)

	return math.Max(0, math.Min(1, normalizedRisk))
}

// calculateCorrelationRisk measures correlation risk across positions
func (em *ExposureMonitor) calculateCorrelationRisk(positions []*Position) float64 {
	if len(positions) < 2 {
		return 0
	}

	// Calculate average absolute correlation between positions
	totalCorrelation := 0.0
	pairCount := 0

	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			// Get correlation between instruments
			corr := em.getCorrelation(positions[i].Symbol, positions[j].Symbol)

			// Weight by position sizes
			weight := math.Sqrt(positions[i].Volume * positions[j].Volume)
			totalCorrelation += math.Abs(corr) * weight
			pairCount++
		}
	}

	if pairCount == 0 {
		return 0
	}

	avgCorrelation := totalCorrelation / float64(pairCount)

	// Normalize to 0-1 scale (high correlation = high risk)
	return math.Max(0, math.Min(1, avgCorrelation))
}

// calculateVaR calculates Value at Risk at 95% and 99% confidence levels
func (em *ExposureMonitor) calculateVaR(
	accountID int64,
	positions []*Position,
) (float64, float64) {
	if len(positions) == 0 {
		return 0, 0
	}

	// Parametric VaR calculation
	// VaR = Portfolio Value × σ × Z-score

	account, err := em.engine.GetAccountByID(accountID)
	if err != nil {
		return 0, 0
	}

	portfolioValue := account.Equity

	// Calculate portfolio standard deviation
	portfolioVariance := 0.0

	for _, pos := range positions {
		contractSize := em.getContractSize(pos.Symbol)
		notional := pos.Volume * contractSize * pos.CurrentPrice
		weight := notional / portfolioValue

		volatility := em.engine.GetVolatility(pos.Symbol)
		if volatility == 0 {
			volatility = 0.15 // Default 15%
		}

		// Daily volatility
		dailyVol := volatility / math.Sqrt(252)

		// Variance contribution
		portfolioVariance += (weight * dailyVol) * (weight * dailyVol)
	}

	// Add correlation effects (simplified)
	for i := 0; i < len(positions); i++ {
		for j := i + 1; j < len(positions); j++ {
			contractSize1 := em.getContractSize(positions[i].Symbol)
			notional1 := positions[i].Volume * contractSize1 * positions[i].CurrentPrice
			weight1 := notional1 / portfolioValue

			contractSize2 := em.getContractSize(positions[j].Symbol)
			notional2 := positions[j].Volume * contractSize2 * positions[j].CurrentPrice
			weight2 := notional2 / portfolioValue

			vol1 := em.engine.GetVolatility(positions[i].Symbol) / math.Sqrt(252)
			vol2 := em.engine.GetVolatility(positions[j].Symbol) / math.Sqrt(252)

			if vol1 == 0 {
				vol1 = 0.15 / math.Sqrt(252)
			}
			if vol2 == 0 {
				vol2 = 0.15 / math.Sqrt(252)
			}

			corr := em.getCorrelation(positions[i].Symbol, positions[j].Symbol)

			portfolioVariance += 2 * weight1 * weight2 * vol1 * vol2 * corr
		}
	}

	portfolioStdDev := math.Sqrt(portfolioVariance)

	// Z-scores for confidence levels
	z95 := 1.645 // 95% confidence
	z99 := 2.326 // 99% confidence

	var95 := portfolioValue * portfolioStdDev * z95
	var99 := portfolioValue * portfolioStdDev * z99

	log.Printf("[VaR] Account #%d: Portfolio σ=%.4f, VaR95=%.2f, VaR99=%.2f",
		accountID, portfolioStdDev, var95, var99)

	return var95, var99
}

// calculateLiquidityScore calculates a liquidity score (0-1, higher = more liquid)
func (em *ExposureMonitor) calculateLiquidityScore(positions []*Position) float64 {
	if len(positions) == 0 {
		return 1.0
	}

	totalLiquidity := 0.0
	totalWeight := 0.0

	for _, pos := range positions {
		// Liquidity score based on instrument type
		liquidityScore := em.getInstrumentLiquidity(pos.Symbol)

		// Weight by position size
		contractSize := em.getContractSize(pos.Symbol)
		notional := pos.Volume * contractSize * pos.CurrentPrice

		totalLiquidity += liquidityScore * notional
		totalWeight += notional
	}

	if totalWeight == 0 {
		return 1.0
	}

	avgLiquidity := totalLiquidity / totalWeight
	return avgLiquidity
}

// getInstrumentLiquidity returns liquidity score for an instrument
func (em *ExposureMonitor) getInstrumentLiquidity(symbol string) float64 {
	// Major FX pairs: very liquid (1.0)
	majorPairs := map[string]bool{
		"EURUSD": true, "GBPUSD": true, "USDJPY": true,
		"USDCHF": true, "AUDUSD": true, "USDCAD": true,
	}
	if majorPairs[symbol] {
		return 1.0
	}

	// Minor FX pairs: liquid (0.8)
	if len(symbol) == 6 {
		return 0.8
	}

	// Gold: very liquid (0.95)
	if symbol == "XAUUSD" {
		return 0.95
	}

	// Major indices: liquid (0.9)
	indices := map[string]bool{
		"SPX500USD": true, "US30USD": true, "NAS100USD": true,
	}
	if indices[symbol] {
		return 0.9
	}

	// Crypto: moderate (0.6-0.7)
	cryptos := map[string]float64{
		"BTCUSD": 0.7,
		"ETHUSD": 0.65,
		"XRPUSD": 0.6,
	}
	if score, ok := cryptos[symbol]; ok {
		return score
	}

	// Default: moderate liquidity
	return 0.7
}

// getAssetClass categorizes instrument by asset class
func (em *ExposureMonitor) getAssetClass(symbol string) string {
	// FX
	if len(symbol) == 6 {
		return "FX"
	}

	// Metals
	metals := map[string]bool{
		"XAUUSD": true, "XAGUSD": true, "XPTUSD": true, "XPDUSD": true,
	}
	if metals[symbol] {
		return "METALS"
	}

	// Crypto
	cryptos := map[string]bool{
		"BTCUSD": true, "ETHUSD": true, "XRPUSD": true,
	}
	if cryptos[symbol] {
		return "CRYPTO"
	}

	// Indices
	indices := map[string]bool{
		"SPX500USD": true, "US30USD": true, "NAS100USD": true,
		"UK100GBP": true, "DE30EUR": true,
	}
	if indices[symbol] {
		return "INDICES"
	}

	// Commodities
	commodities := map[string]bool{
		"WTICOUSD": true, "BCOUSD": true, "NATGASUSD": true,
	}
	if commodities[symbol] {
		return "COMMODITIES"
	}

	return "OTHER"
}

// getCorrelation gets correlation between two instruments
func (em *ExposureMonitor) getCorrelation(symbol1, symbol2 string) float64 {
	if symbol1 == symbol2 {
		return 1.0
	}

	// Simplified correlation matrix
	// In production, this would use historical data

	// Same asset class = moderate correlation
	if em.getAssetClass(symbol1) == em.getAssetClass(symbol2) {
		// FX pairs sharing a currency = higher correlation
		if len(symbol1) == 6 && len(symbol2) == 6 {
			base1 := symbol1[:3]
			quote1 := symbol1[3:]
			base2 := symbol2[:3]
			quote2 := symbol2[3:]

			if base1 == base2 || quote1 == quote2 {
				return 0.7 // High correlation
			}
			if base1 == quote2 || quote1 == base2 {
				return -0.7 // High negative correlation
			}
			return 0.3 // Moderate correlation
		}

		return 0.5 // Same asset class
	}

	// Different asset classes = low correlation
	return 0.1
}

// getContractSize returns contract size
func (em *ExposureMonitor) getContractSize(symbol string) float64 {
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

// GetAggregateExposure calculates total broker exposure across all accounts
func (em *ExposureMonitor) GetAggregateExposure() map[string]float64 {
	aggregateExposure := make(map[string]float64)
	accounts := em.engine.GetAllAccounts()

	for _, account := range accounts {
		positions := em.engine.GetAllPositions(account.ID)

		for _, pos := range positions {
			contractSize := em.getContractSize(pos.Symbol)
			notional := pos.Volume * contractSize * pos.CurrentPrice

			if pos.Side == "BUY" {
				aggregateExposure[pos.Symbol] += notional
			} else {
				aggregateExposure[pos.Symbol] -= notional
			}
		}
	}

	return aggregateExposure
}
