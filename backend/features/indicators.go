package features

import (
	"errors"
	"log"
	"math"
	"sync"
)

// Technical Indicators Service
// Implements all major indicators for server-side calculation

// OHLCBar represents a price bar
type OHLCBar struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume,omitempty"`
}

// IndicatorValue represents a calculated indicator value
type IndicatorValue struct {
	Timestamp int64              `json:"timestamp"`
	Values    map[string]float64 `json:"values"` // e.g., {"value": 50.5, "signal": 45.2}
}

// IndicatorService calculates technical indicators
type IndicatorService struct {
	mu          sync.RWMutex
	priceCache  map[string][]OHLCBar // symbol -> recent bars
	maxBars     int                  // Maximum bars to keep in cache
	dataCallback func(symbol, timeframe string, count int) ([]OHLCBar, error)
}

// NewIndicatorService creates the indicator service
func NewIndicatorService(maxBars int) *IndicatorService {
	if maxBars < 200 {
		maxBars = 200 // Minimum for most indicators
	}

	svc := &IndicatorService{
		priceCache: make(map[string][]OHLCBar),
		maxBars:    maxBars,
	}

	log.Printf("[IndicatorService] Initialized with cache for %d bars", maxBars)
	return svc
}

// SetDataCallback sets the function to fetch historical data
func (s *IndicatorService) SetDataCallback(fn func(symbol, timeframe string, count int) ([]OHLCBar, error)) {
	s.dataCallback = fn
}

// UpdateBars adds new bars to cache
func (s *IndicatorService) UpdateBars(symbol string, bars []OHLCBar) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.priceCache[symbol] = bars
	if len(s.priceCache[symbol]) > s.maxBars {
		s.priceCache[symbol] = s.priceCache[symbol][len(s.priceCache[symbol])-s.maxBars:]
	}
}

// ===== Moving Averages =====

// SMA calculates Simple Moving Average
func (s *IndicatorService) SMA(symbol string, period int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period+50)
	if err != nil {
		return nil, err
	}

	if len(bars) < period {
		return nil, errors.New("insufficient data for SMA")
	}

	result := make([]IndicatorValue, 0)

	for i := period - 1; i < len(bars); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += bars[j].Close
		}
		avg := sum / float64(period)

		result = append(result, IndicatorValue{
			Timestamp: bars[i].Timestamp,
			Values:    map[string]float64{"sma": avg},
		})
	}

	return result, nil
}

// EMA calculates Exponential Moving Average
func (s *IndicatorService) EMA(symbol string, period int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period*3)
	if err != nil {
		return nil, err
	}

	if len(bars) < period {
		return nil, errors.New("insufficient data for EMA")
	}

	multiplier := 2.0 / float64(period+1)
	result := make([]IndicatorValue, 0)

	// Start with SMA as first EMA value
	smaSum := 0.0
	for i := 0; i < period; i++ {
		smaSum += bars[i].Close
	}
	ema := smaSum / float64(period)

	result = append(result, IndicatorValue{
		Timestamp: bars[period-1].Timestamp,
		Values:    map[string]float64{"ema": ema},
	})

	// Calculate EMA for remaining bars
	for i := period; i < len(bars); i++ {
		ema = (bars[i].Close-ema)*multiplier + ema
		result = append(result, IndicatorValue{
			Timestamp: bars[i].Timestamp,
			Values:    map[string]float64{"ema": ema},
		})
	}

	return result, nil
}

// WMA calculates Weighted Moving Average
func (s *IndicatorService) WMA(symbol string, period int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period+50)
	if err != nil {
		return nil, err
	}

	if len(bars) < period {
		return nil, errors.New("insufficient data for WMA")
	}

	result := make([]IndicatorValue, 0)
	weightSum := float64(period * (period + 1) / 2)

	for i := period - 1; i < len(bars); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			weight := float64(j + 1)
			sum += bars[i-period+1+j].Close * weight
		}
		wma := sum / weightSum

		result = append(result, IndicatorValue{
			Timestamp: bars[i].Timestamp,
			Values:    map[string]float64{"wma": wma},
		})
	}

	return result, nil
}

// ===== Oscillators =====

// RSI calculates Relative Strength Index
func (s *IndicatorService) RSI(symbol string, period int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period*3)
	if err != nil {
		return nil, err
	}

	if len(bars) < period+1 {
		return nil, errors.New("insufficient data for RSI")
	}

	result := make([]IndicatorValue, 0)
	gains := make([]float64, 0)
	losses := make([]float64, 0)

	// Calculate price changes
	for i := 1; i < len(bars); i++ {
		change := bars[i].Close - bars[i-1].Close
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, math.Abs(change))
		}
	}

	// Calculate initial averages
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI
	for i := period; i < len(gains); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)

		rs := avgGain / avgLoss
		rsi := 100.0 - (100.0 / (1.0 + rs))

		result = append(result, IndicatorValue{
			Timestamp: bars[i+1].Timestamp,
			Values:    map[string]float64{"rsi": rsi},
		})
	}

	return result, nil
}

// MACD calculates Moving Average Convergence Divergence
func (s *IndicatorService) MACD(symbol string, fastPeriod, slowPeriod, signalPeriod int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, slowPeriod*3)
	if err != nil {
		return nil, err
	}

	if len(bars) < slowPeriod {
		return nil, errors.New("insufficient data for MACD")
	}

	// Calculate fast and slow EMA
	fastEMA := s.calculateEMA(bars, fastPeriod)
	slowEMA := s.calculateEMA(bars, slowPeriod)

	// Calculate MACD line
	macdLine := make([]float64, 0)
	startIdx := slowPeriod - 1

	for i := startIdx; i < len(fastEMA); i++ {
		macdLine = append(macdLine, fastEMA[i]-slowEMA[i-startIdx])
	}

	// Calculate signal line (EMA of MACD)
	signalLine := s.calculateEMAFromValues(macdLine, signalPeriod)

	// Calculate histogram
	result := make([]IndicatorValue, 0)
	for i := signalPeriod - 1; i < len(macdLine); i++ {
		histogram := macdLine[i] - signalLine[i-signalPeriod+1]

		result = append(result, IndicatorValue{
			Timestamp: bars[startIdx+i].Timestamp,
			Values: map[string]float64{
				"macd":      macdLine[i],
				"signal":    signalLine[i-signalPeriod+1],
				"histogram": histogram,
			},
		})
	}

	return result, nil
}

// Stochastic calculates Stochastic Oscillator
func (s *IndicatorService) Stochastic(symbol string, kPeriod, dPeriod int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, kPeriod*2+dPeriod)
	if err != nil {
		return nil, err
	}

	if len(bars) < kPeriod {
		return nil, errors.New("insufficient data for Stochastic")
	}

	result := make([]IndicatorValue, 0)
	kValues := make([]float64, 0)

	// Calculate %K
	for i := kPeriod - 1; i < len(bars); i++ {
		high := bars[i].High
		low := bars[i].Low

		for j := i - kPeriod + 1; j < i; j++ {
			if bars[j].High > high {
				high = bars[j].High
			}
			if bars[j].Low < low {
				low = bars[j].Low
			}
		}

		k := 0.0
		if high != low {
			k = ((bars[i].Close - low) / (high - low)) * 100.0
		}

		kValues = append(kValues, k)
	}

	// Calculate %D (SMA of %K)
	for i := dPeriod - 1; i < len(kValues); i++ {
		sum := 0.0
		for j := i - dPeriod + 1; j <= i; j++ {
			sum += kValues[j]
		}
		d := sum / float64(dPeriod)

		result = append(result, IndicatorValue{
			Timestamp: bars[kPeriod-1+i].Timestamp,
			Values: map[string]float64{
				"k": kValues[i],
				"d": d,
			},
		})
	}

	return result, nil
}

// ===== Volatility Indicators =====

// BollingerBands calculates Bollinger Bands
func (s *IndicatorService) BollingerBands(symbol string, period int, stdDev float64) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period+50)
	if err != nil {
		return nil, err
	}

	if len(bars) < period {
		return nil, errors.New("insufficient data for Bollinger Bands")
	}

	result := make([]IndicatorValue, 0)

	for i := period - 1; i < len(bars); i++ {
		// Calculate SMA
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += bars[j].Close
		}
		sma := sum / float64(period)

		// Calculate standard deviation
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			variance += math.Pow(bars[j].Close-sma, 2)
		}
		stdDevValue := math.Sqrt(variance / float64(period))

		upper := sma + (stdDev * stdDevValue)
		lower := sma - (stdDev * stdDevValue)

		result = append(result, IndicatorValue{
			Timestamp: bars[i].Timestamp,
			Values: map[string]float64{
				"upper":  upper,
				"middle": sma,
				"lower":  lower,
			},
		})
	}

	return result, nil
}

// ATR calculates Average True Range
func (s *IndicatorService) ATR(symbol string, period int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period*2)
	if err != nil {
		return nil, err
	}

	if len(bars) < period+1 {
		return nil, errors.New("insufficient data for ATR")
	}

	result := make([]IndicatorValue, 0)
	trueRanges := make([]float64, 0)

	// Calculate True Range for each bar
	for i := 1; i < len(bars); i++ {
		tr1 := bars[i].High - bars[i].Low
		tr2 := math.Abs(bars[i].High - bars[i-1].Close)
		tr3 := math.Abs(bars[i].Low - bars[i-1].Close)

		tr := math.Max(tr1, math.Max(tr2, tr3))
		trueRanges = append(trueRanges, tr)
	}

	// Calculate initial ATR (SMA of TR)
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += trueRanges[i]
	}
	atr := sum / float64(period)

	result = append(result, IndicatorValue{
		Timestamp: bars[period].Timestamp,
		Values:    map[string]float64{"atr": atr},
	})

	// Calculate smoothed ATR
	for i := period; i < len(trueRanges); i++ {
		atr = ((atr * float64(period-1)) + trueRanges[i]) / float64(period)
		result = append(result, IndicatorValue{
			Timestamp: bars[i+1].Timestamp,
			Values:    map[string]float64{"atr": atr},
		})
	}

	return result, nil
}

// ADX calculates Average Directional Index
func (s *IndicatorService) ADX(symbol string, period int) ([]IndicatorValue, error) {
	bars, err := s.getBars(symbol, period*3)
	if err != nil {
		return nil, err
	}

	if len(bars) < period*2 {
		return nil, errors.New("insufficient data for ADX")
	}

	result := make([]IndicatorValue, 0)

	// Calculate +DM, -DM, and TR
	plusDM := make([]float64, 0)
	minusDM := make([]float64, 0)
	tr := make([]float64, 0)

	for i := 1; i < len(bars); i++ {
		highDiff := bars[i].High - bars[i-1].High
		lowDiff := bars[i-1].Low - bars[i].Low

		pDM := 0.0
		mDM := 0.0

		if highDiff > lowDiff && highDiff > 0 {
			pDM = highDiff
		}
		if lowDiff > highDiff && lowDiff > 0 {
			mDM = lowDiff
		}

		plusDM = append(plusDM, pDM)
		minusDM = append(minusDM, mDM)

		tr1 := bars[i].High - bars[i].Low
		tr2 := math.Abs(bars[i].High - bars[i-1].Close)
		tr3 := math.Abs(bars[i].Low - bars[i-1].Close)
		tr = append(tr, math.Max(tr1, math.Max(tr2, tr3)))
	}

	// Smooth +DM, -DM, TR
	smoothPlusDM := s.smoothValues(plusDM, period)
	smoothMinusDM := s.smoothValues(minusDM, period)
	smoothTR := s.smoothValues(tr, period)

	// Calculate +DI and -DI
	plusDI := make([]float64, len(smoothPlusDM))
	minusDI := make([]float64, len(smoothMinusDM))

	for i := 0; i < len(smoothPlusDM); i++ {
		if smoothTR[i] != 0 {
			plusDI[i] = 100 * smoothPlusDM[i] / smoothTR[i]
			minusDI[i] = 100 * smoothMinusDM[i] / smoothTR[i]
		}
	}

	// Calculate DX
	dx := make([]float64, len(plusDI))
	for i := 0; i < len(plusDI); i++ {
		diSum := plusDI[i] + minusDI[i]
		if diSum != 0 {
			dx[i] = 100 * math.Abs(plusDI[i]-minusDI[i]) / diSum
		}
	}

	// Calculate ADX (smoothed DX)
	adxValues := s.smoothValues(dx, period)

	for i := 0; i < len(adxValues); i++ {
		if i+period*2 < len(bars) {
			result = append(result, IndicatorValue{
				Timestamp: bars[i+period*2].Timestamp,
				Values: map[string]float64{
					"adx":     adxValues[i],
					"plusDI":  plusDI[i],
					"minusDI": minusDI[i],
				},
			})
		}
	}

	return result, nil
}

// ===== Support/Resistance =====

// PivotPoints calculates pivot points
func (s *IndicatorService) PivotPoints(symbol string) (map[string]float64, error) {
	bars, err := s.getBars(symbol, 2)
	if err != nil {
		return nil, err
	}

	if len(bars) < 1 {
		return nil, errors.New("insufficient data for Pivot Points")
	}

	lastBar := bars[len(bars)-1]
	pivot := (lastBar.High + lastBar.Low + lastBar.Close) / 3.0

	r1 := 2*pivot - lastBar.Low
	s1 := 2*pivot - lastBar.High
	r2 := pivot + (lastBar.High - lastBar.Low)
	s2 := pivot - (lastBar.High - lastBar.Low)
	r3 := lastBar.High + 2*(pivot-lastBar.Low)
	s3 := lastBar.Low - 2*(lastBar.High-pivot)

	return map[string]float64{
		"pivot": pivot,
		"r1":    r1,
		"r2":    r2,
		"r3":    r3,
		"s1":    s1,
		"s2":    s2,
		"s3":    s3,
	}, nil
}

// FibonacciRetracement calculates Fibonacci levels
func (s *IndicatorService) FibonacciRetracement(high, low float64) map[string]float64 {
	diff := high - low

	return map[string]float64{
		"0.0":   high,
		"23.6":  high - diff*0.236,
		"38.2":  high - diff*0.382,
		"50.0":  high - diff*0.5,
		"61.8":  high - diff*0.618,
		"78.6":  high - diff*0.786,
		"100.0": low,
	}
}

// ===== Helper Functions =====

func (s *IndicatorService) getBars(symbol string, count int) ([]OHLCBar, error) {
	s.mu.RLock()
	cached, exists := s.priceCache[symbol]
	s.mu.RUnlock()

	if exists && len(cached) >= count {
		return cached[len(cached)-count:], nil
	}

	// Fetch from callback if available
	if s.dataCallback != nil {
		bars, err := s.dataCallback(symbol, "M1", count)
		if err == nil {
			s.UpdateBars(symbol, bars)
			return bars, nil
		}
	}

	if exists {
		return cached, nil
	}

	return nil, errors.New("no data available for symbol")
}

func (s *IndicatorService) calculateEMA(bars []OHLCBar, period int) []float64 {
	multiplier := 2.0 / float64(period+1)
	ema := make([]float64, len(bars))

	// Start with SMA
	sum := 0.0
	for i := 0; i < period && i < len(bars); i++ {
		sum += bars[i].Close
	}
	ema[period-1] = sum / float64(period)

	// Calculate EMA
	for i := period; i < len(bars); i++ {
		ema[i] = (bars[i].Close-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

func (s *IndicatorService) calculateEMAFromValues(values []float64, period int) []float64 {
	multiplier := 2.0 / float64(period+1)
	ema := make([]float64, len(values))

	// Start with SMA
	sum := 0.0
	for i := 0; i < period && i < len(values); i++ {
		sum += values[i]
	}
	ema[period-1] = sum / float64(period)

	// Calculate EMA
	for i := period; i < len(values); i++ {
		ema[i] = (values[i]-ema[i-1])*multiplier + ema[i-1]
	}

	return ema
}

func (s *IndicatorService) smoothValues(values []float64, period int) []float64 {
	if len(values) < period {
		return values
	}

	result := make([]float64, len(values)-period+1)

	// Initial sum
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[0] = sum / float64(period)

	// Wilder's smoothing
	for i := period; i < len(values); i++ {
		result[i-period+1] = (result[i-period]*float64(period-1) + values[i]) / float64(period)
	}

	return result
}
