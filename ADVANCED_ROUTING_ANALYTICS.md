# Advanced Analytics Metrics for Broker Routing Systems

## Executive Summary

This document provides comprehensive advanced analytics metrics for multi-asset broker routing systems. It extends beyond basic metrics (fill rate, spread) to cover sophisticated performance measurement, predictive analytics, and risk assessment frameworks.

The trading engine architecture includes:
- A-Book (Agency Model) with Smart Order Router (SOR)
- B-Book (Principal Model) with position management
- Multiple LP (Liquidity Provider) integrations
- Real-time quote aggregation and LP health monitoring

---

## 1. ADVANCED PERFORMANCE METRICS

### 1.1 Sharpe Ratio by Routing Strategy

**Purpose**: Measure risk-adjusted returns for different routing decisions

#### Formula
```
Sharpe Ratio = (Rp - Rf) / σp

Where:
  Rp = Average return of routing strategy
  Rf = Risk-free rate (e.g., 2% annual)
  σp = Standard deviation of returns
```

#### Calculation Method
```
1. Group trades by routing decision (Primary LP, Alternative LP, etc.)
2. Calculate realized P/L per trade: RealizedPnL = Volume * (Exit Price - Entry Price)
3. Calculate strategy returns: Rp = Σ(RealizedPnL) / NumTrades
4. Calculate standard deviation of trade returns
5. Apply risk-free rate adjustment
```

#### Implementation Example
```go
type RoutingStrategyMetrics struct {
    RoutingID        string
    TotalTrades      int64
    TotalPnL         float64
    AveragePnL       float64
    StdDeviation     float64
    SharpeRatio      float64
    WorstTrade       float64
    BestTrade        float64
    ConsecutiveLosses int
}

func CalculateSharpeRatio(trades []Trade, riskFreeRate float64) float64 {
    if len(trades) == 0 {
        return 0
    }

    // Calculate returns
    var returns []float64
    var totalReturn float64

    for _, trade := range trades {
        returnValue := trade.RealizedPnL / trade.EntryValue
        returns = append(returns, returnValue)
        totalReturn += returnValue
    }

    meanReturn := totalReturn / float64(len(returns))

    // Calculate standard deviation
    var variance float64
    for _, r := range returns {
        diff := r - meanReturn
        variance += diff * diff
    }
    stdDev := math.Sqrt(variance / float64(len(returns)-1))

    if stdDev == 0 {
        return 0
    }

    return (meanReturn - riskFreeRate) / stdDev
}
```

#### Visualization Recommendations
- **Line Chart**: Sharpe Ratio over time for each routing strategy
- **Scatter Plot**: Sharpe Ratio vs. Average Slippage by LP
- **Heatmap**: Sharpe Ratio matrix (Routing Strategy × Time Period)

---

### 1.2 Profit Factor by Routing Path

**Purpose**: Measure profitability efficiency relative to losses

#### Formula
```
Profit Factor = Gross Profit / Gross Loss

Where:
  Gross Profit = Sum of all winning trades
  Gross Loss = Sum of all losing trades (absolute value)

Interpretation:
  PF > 2.0 = Excellent
  PF > 1.5 = Good
  PF > 1.0 = Profitable
  PF < 1.0 = Unprofitable
```

#### Calculation Method
```
1. Identify routing paths (e.g., "BUY → LP_A → YOFX1", "SELL → LP_B → LMAX")
2. Group trades by routing path
3. Sum all winning trades per path
4. Sum all losing trades per path (as positive value)
5. Calculate ratio
```

#### Implementation Example
```go
type RoutingPath struct {
    PathID          string // e.g., "EURUSD_BUY_OANDA_YOFX1"
    Side            string // BUY/SELL
    Symbol          string
    SelectedLP      string
    SessionID       string
    TotalTrades     int64
    WinningTrades   int64
    LosingTrades    int64
    GrossProfit     float64
    GrossLoss       float64
    ProfitFactor    float64
    WinRate         float64
    LossRate        float64
}

func CalculateProfitFactor(trades []Trade) (float64, float64, float64) {
    var grossProfit, grossLoss float64
    var winningCount, losingCount int

    for _, trade := range trades {
        if trade.RealizedPnL > 0 {
            grossProfit += trade.RealizedPnL
            winningCount++
        } else if trade.RealizedPnL < 0 {
            grossLoss += math.Abs(trade.RealizedPnL)
            losingCount++
        }
    }

    if grossLoss == 0 {
        return math.MaxFloat64, float64(winningCount), float64(losingCount)
    }

    profitFactor := grossProfit / grossLoss
    winRate := float64(winningCount) / float64(len(trades))

    return profitFactor, winRate, float64(losingCount) / float64(len(trades))
}
```

#### Key Sub-Metrics
- **Winning Trade Count**: Number of profitable trades
- **Average Win Size**: Mean P/L of profitable trades
- **Average Loss Size**: Mean loss of unprofitable trades
- **Win/Loss Ratio**: Ratio of average win to average loss
- **Consecutive Wins/Losses**: Streak analysis for strategy stability

#### Visualization Recommendations
- **Bar Chart**: Profit Factor by routing path
- **Waterfall Chart**: Cumulative P/L breakdown by path
- **Gauge Chart**: Profit Factor with color zones (Red: <1.0, Yellow: 1.0-1.5, Green: >1.5)

---

### 1.3 Maximum Drawdown (MDd) by Liquidity Provider

**Purpose**: Measure the largest peak-to-trough decline in cumulative profits

#### Formula
```
Drawdown = (Trough Value - Peak Value) / Peak Value × 100%

Maximum Drawdown = Min Drawdown over period
Duration = Number of bars from peak to recovery
```

#### Calculation Method
```
1. Calculate cumulative P/L for each LP
2. Track running peak value
3. For each point: Calculate drawdown = (Current Value - Peak) / Peak
4. Record maximum negative drawdown
5. Measure duration to recovery
```

#### Implementation Example
```go
type LPDrawdownMetrics struct {
    LPID              string
    StartDate         time.Time
    EndDate           time.Time
    MaxDrawdown       float64  // As percentage
    MaxDrawdownValue  float64  // In USD
    DrawdownDuration  time.Duration
    RecoveryDate      time.Time
    ConsecutiveDownDays int
}

func CalculateMaxDrawdown(lpTrades []Trade) (*LPDrawdownMetrics, error) {
    if len(lpTrades) == 0 {
        return nil, errors.New("no trades for LP")
    }

    // Sort by execution time
    sort.Slice(lpTrades, func(i, j int) bool {
        return lpTrades[i].ExecutedAt.Before(lpTrades[j].ExecutedAt)
    })

    var maxDrawdown float64
    var peakValue float64 = 0
    var troughValue float64 = 0
    var maxDrawdownDate time.Time
    var recoveryDate time.Time

    cumPnL := 0.0

    for i, trade := range lpTrades {
        cumPnL += trade.RealizedPnL

        // Update peak
        if cumPnL > peakValue {
            peakValue = cumPnL
        }

        // Calculate drawdown
        if peakValue > 0 {
            drawdown := (cumPnL - peakValue) / peakValue
            if drawdown < maxDrawdown {
                maxDrawdown = drawdown
                maxDrawdownDate = trade.ExecutedAt
                troughValue = cumPnL
            }
        }

        // Check for recovery
        if cumPnL >= peakValue && !maxDrawdownDate.IsZero() {
            recoveryDate = trade.ExecutedAt
        }
    }

    duration := recoveryDate.Sub(maxDrawdownDate)

    return &LPDrawdownMetrics{
        LPID:             "", // Set externally
        MaxDrawdown:      maxDrawdown * 100,
        MaxDrawdownValue: troughValue - peakValue,
        DrawdownDuration: duration,
        RecoveryDate:     recoveryDate,
    }, nil
}
```

#### Enhanced Drawdown Types
- **Underwater Plot**: Visual representation of drawdown over time
- **Calmar Ratio**: Return / Max Drawdown (measures return per unit of downside risk)
- **Recovery Factor**: Net Profit / Max Drawdown
- **Painindex**: Sum of all drawdowns (Total Drawdown)

```go
// Calmar Ratio = Annual Return / Max Drawdown
calmarRatio := (annualReturn / math.Abs(maxDrawdown)) * 100

// Recovery Factor = Net Profit / Max Drawdown
recoveryFactor := totalNetProfit / math.Abs(maxDrawdownValue)

// Painindex = Sum of all drawdowns
var painIndex float64
for _, dd := range allDrawdowns {
    painIndex += dd
}
```

#### Visualization Recommendations
- **Equity Curve with Drawdown Bands**: Show cumulative P/L with shaded drawdown areas
- **Drawdown Duration Histogram**: Distribution of drawdown recovery times
- **Heatmap**: Max Drawdown by LP and Time Period

---

### 1.4 Win Rate by Client Toxicity Score

**Purpose**: Measure execution quality degradation with order toxicity

#### Client Toxicity Score
```
Toxicity Score = (Adverse Selection + Latency + Informed Order Ratio) / 3

Where:
  Adverse Selection = % of trades where next tick moved against the fill
  Latency = Order-to-fill time in milliseconds
  Informed Order Ratio = % of orders that move price significantly
```

#### Calculation Method
```
1. For each client:
   a. Calculate adverse selection: Count trades where next quote moved unfavorably
   b. Measure order latency: Time from submission to execution
   c. Identify informed orders: Orders that precede price movements

2. Normalize each component to 0-1 scale
3. Apply weighting: 40% Adverse Selection, 35% Latency, 25% Informed Ratio
4. Segment clients into toxicity buckets
5. Calculate win rate by bucket
```

#### Implementation Example
```go
type ClientToxicityProfile struct {
    ClientID              string
    ToxicityScore         float64 // 0-1, higher = more toxic
    AdverseSelectionRate  float64
    AvgOrderLatency       time.Duration
    InformedOrderRatio    float64
    TotalOrders           int64
    AverageWinRate        float64
    EstimatedCost         float64
}

func CalculateToxicityScore(client *ClientProfile, trades []Trade) *ClientToxicityProfile {
    // 1. Calculate adverse selection
    var adverseCount int
    for _, trade := range trades {
        nextTick := getNextMarketTick(trade.Symbol, trade.ExecutedAt)
        if nextTick != nil {
            if trade.Side == "BUY" && nextTick.Bid < trade.FilledPrice {
                adverseCount++
            } else if trade.Side == "SELL" && nextTick.Ask > trade.FilledPrice {
                adverseCount++
            }
        }
    }
    adverseRate := float64(adverseCount) / float64(len(trades))

    // 2. Calculate latency
    var totalLatency time.Duration
    for _, trade := range trades {
        latency := trade.ExecutedAt.Sub(trade.CreatedAt)
        totalLatency += latency
    }
    avgLatency := totalLatency / time.Duration(len(trades))
    latencyScore := math.Min(float64(avgLatency.Milliseconds())/1000.0, 1.0)

    // 3. Identify informed orders
    informedCount := 0
    for _, trade := range trades {
        // Check if order moved price significantly (more than 2x normal spread)
        priceMovement := calculatePriceMovement(trade)
        if priceMovement > 2.0 {
            informedCount++
        }
    }
    informedRatio := float64(informedCount) / float64(len(trades))

    // Calculate toxicity score
    toxicity := (adverseRate*0.4) + (latencyScore*0.35) + (informedRatio*0.25)

    // Calculate win rate by toxicity bucket
    var winCount int
    for _, trade := range trades {
        if trade.RealizedPnL > 0 {
            winCount++
        }
    }
    winRate := float64(winCount) / float64(len(trades))

    return &ClientToxicityProfile{
        ClientID:              client.ID,
        ToxicityScore:         toxicity,
        AdverseSelectionRate:  adverseRate,
        AvgOrderLatency:       avgLatency,
        InformedOrderRatio:    informedRatio,
        TotalOrders:           int64(len(trades)),
        AverageWinRate:        winRate,
        EstimatedCost:         toxicity * float64(len(trades)) * client.AvgOrderSize,
    }
}

// Toxicity Segmentation
func SegmentClientsByToxicity(profiles []ClientToxicityProfile) map[string][]ClientToxicityProfile {
    buckets := make(map[string][]ClientToxicityProfile)

    for _, profile := range profiles {
        var bucket string
        if profile.ToxicityScore < 0.25 {
            bucket = "LOW_TOXICITY"
        } else if profile.ToxicityScore < 0.50 {
            bucket = "MEDIUM_TOXICITY"
        } else if profile.ToxicityScore < 0.75 {
            bucket = "HIGH_TOXICITY"
        } else {
            bucket = "EXTREME_TOXICITY"
        }
        buckets[bucket] = append(buckets[bucket], profile)
    }

    return buckets
}
```

#### Toxicity Components Deep Dive

**Adverse Selection Analysis**
```go
type AdverseSelectionMetrics struct {
    DirectAS        float64 // % trades adversely affected immediately
    MicrostructureAS float64 // Adverse selection due to order timing
    InformationAS   float64 // Due to informed trading
    TemporalAS      float64 // Due to time of day patterns
}

func AnalyzeAdverseSelection(trade Trade, nextTick Quote) *AdverseSelectionMetrics {
    // Direct AS: Immediate adverse movement
    directAS := calculateDirectAverseness(trade, nextTick)

    // Other components require historical context
    // ...

    return &AdverseSelectionMetrics{
        DirectAS: directAS,
    }
}
```

#### Visualization Recommendations
- **Scatter Plot**: Toxicity Score vs. Win Rate (with client size as bubble size)
- **Box Plot**: Win Rate distribution by toxicity bucket
- **Time Series**: Toxicity score evolution for top clients
- **Toxicity Dashboard**: Client segmentation with metrics

---

### 1.5 Average Fill Time by Routing Decision

**Purpose**: Measure execution speed for different routing strategies

#### Key Metrics
```
Fill Time = Execution Timestamp - Order Submission Time

Components:
  Quote Reception Time: Time to receive quote from LP
  Decision Time: Time to select routing path
  Transmission Time: Time to send order to LP
  LP Processing Time: Time for LP to execute
  Confirmation Time: Time to receive fill confirmation
```

#### Calculation Method
```go
type FillTimeAnalysis struct {
    RoutingDecision     string
    TotalOrders         int64
    AverageFillTime     time.Duration
    MedianFillTime      time.Duration
    P95FillTime         time.Duration
    P99FillTime         time.Duration
    MinFillTime         time.Duration
    MaxFillTime         time.Duration
    StdDeviation        time.Duration
    QuoteReceptionTime  time.Duration
    DecisionLatency     time.Duration
    TransmissionTime    time.Duration
    LPProcessingTime    time.Duration
}

func CalculateFillTimeMetrics(orders []Order) *FillTimeAnalysis {
    if len(orders) == 0 {
        return nil
    }

    var fillTimes []time.Duration
    var totalTime time.Duration

    // Collect all fill times
    for _, order := range orders {
        if order.FilledAt != nil {
            fillTime := order.FilledAt.Sub(order.CreatedAt)
            fillTimes = append(fillTimes, fillTime)
            totalTime += fillTime
        }
    }

    // Calculate statistics
    sort.Slice(fillTimes, func(i, j int) bool {
        return fillTimes[i] < fillTimes[j]
    })

    avgFillTime := totalTime / time.Duration(len(fillTimes))
    medianFillTime := fillTimes[len(fillTimes)/2]
    p95FillTime := fillTimes[int(float64(len(fillTimes))*0.95)]
    p99FillTime := fillTimes[int(float64(len(fillTimes))*0.99)]

    // Calculate standard deviation
    var variance time.Duration
    for _, ft := range fillTimes {
        diff := ft - avgFillTime
        variance += diff * diff
    }
    stdDev := time.Duration(math.Sqrt(float64(variance) / float64(len(fillTimes))))

    return &FillTimeAnalysis{
        TotalOrders:     int64(len(fillTimes)),
        AverageFillTime: avgFillTime,
        MedianFillTime:  medianFillTime,
        P95FillTime:     p95FillTime,
        P99FillTime:     p99FillTime,
        MinFillTime:     fillTimes[0],
        MaxFillTime:     fillTimes[len(fillTimes)-1],
        StdDeviation:    stdDev,
    }
}

// Decompose fill time components
func AnalyzeFillTimeComponents(order Order, quoteReception, decision, transmission, lpProcessing time.Duration) {
    order.QuoteReceptionTime = quoteReception
    order.DecisionLatency = decision
    order.TransmissionTime = transmission
    order.LPProcessingTime = lpProcessing

    totalFillTime := quoteReception + decision + transmission + lpProcessing

    // Identify bottlenecks
    components := map[string]time.Duration{
        "Quote Reception": quoteReception,
        "Routing Decision": decision,
        "Transmission": transmission,
        "LP Processing": lpProcessing,
    }

    // Find slowest component
    for name, duration := range components {
        if duration == max(quoteReception, decision, transmission, lpProcessing) {
            log.Printf("Bottleneck: %s (%.2f ms)", name, float64(duration.Microseconds())/1000.0)
        }
    }
}
```

#### Fill Time by Order Type
```go
type FillTimeByOrderType struct {
    OrderType       string // MARKET, LIMIT, STOP, STOP_LIMIT
    AverageFillTime time.Duration
    Volatility      float64 // StdDev
    SuccessRate     float64 // % of orders filled
}
```

#### Visualization Recommendations
- **Box Plot**: Fill time distribution by routing decision
- **Histogram**: Fill time with percentile overlays (P50, P95, P99)
- **Time Series**: Average fill time over time for each routing path
- **Waterfall**: Decompose fill time into components

---

### 1.6 Slippage Distribution Analysis

**Purpose**: Comprehensive analysis of execution price vs. reference price

#### Slippage Definition
```
Slippage = Executed Price - Reference Price

Where:
  Reference Price = Best available price at time of order
  For BUY orders: Reference = Lowest Ask
  For SELL orders: Reference = Highest Bid

Slippage Types:
  Positive Slippage: Favorable execution (received better price)
  Negative Slippage: Unfavorable execution (received worse price)
  Bid-Ask Slippage: Slippage due to spread at execution
  Adverse Slippage: Price moved against order between submission and fill
```

#### Advanced Slippage Metrics
```go
type SlippageAnalysis struct {
    Symbol              string
    OrderSide           string
    TotalOrders         int64
    AvgSlippage         float64 // In pips
    MedianSlippage      float64
    P95Slippage         float64
    WorstSlippage       float64
    BestSlippage        float64
    PositiveSlippage    float64 // Orders with favorable slippage
    NegativeSlippage    float64 // Orders with adverse slippage
    SlippageVolatility  float64 // StdDev
    TotalSlippageCost   float64 // In USD
    CostAsPercentage    float64 // % of order value
}

func AnalyzeSlippage(orders []Order, referenceQuotes map[int64]Quote) *SlippageAnalysis {
    var slippages []float64
    var positiveCount, negativeCount int
    var totalCost float64

    for _, order := range orders {
        if order.FilledAt == nil {
            continue
        }

        refQuote := referenceQuotes[order.ID]
        var refPrice float64

        if order.Side == "BUY" {
            refPrice = refQuote.Ask
        } else {
            refPrice = refQuote.Bid
        }

        slippage := order.FilledPrice - refPrice
        slippages = append(slippages, slippage)

        if slippage > 0 {
            positiveCount++
        } else {
            negativeCount++
        }

        // Calculate cost impact
        orderValue := order.Volume * order.FilledPrice
        slippageCost := math.Abs(slippage) * order.Volume
        totalCost += slippageCost
    }

    // Sort for percentile calculations
    sort.Float64s(slippages)

    avgSlippage := calculateMean(slippages)
    medianSlippage := slippages[len(slippages)/2]
    p95Slippage := slippages[int(float64(len(slippages))*0.95)]
    stdDev := calculateStdDev(slippages)

    return &SlippageAnalysis{
        TotalOrders:        int64(len(slippages)),
        AvgSlippage:        avgSlippage,
        MedianSlippage:     medianSlippage,
        P95Slippage:        p95Slippage,
        PositiveSlippage:   float64(positiveCount) / float64(len(slippages)),
        NegativeSlippage:   float64(negativeCount) / float64(len(slippages)),
        SlippageVolatility: stdDev,
        TotalSlippageCost:  totalCost,
    }
}

// Slippage by market conditions
type SlippageByCondition struct {
    Condition           string // LOW_VOLATILITY, HIGH_VOLATILITY, OFF_HOURS, LIQUID_HOURS
    AvgSlippage         float64
    OrderCount          int64
    WorstSlippage       float64
}

func AnalyzeSlippageByCondition(orders []Order) map[string]*SlippageByCondition {
    conditionMap := make(map[string]*SlippageByCondition)

    for _, order := range orders {
        condition := determineMarketCondition(order.ExecutionTime)

        if _, exists := conditionMap[condition]; !exists {
            conditionMap[condition] = &SlippageByCondition{Condition: condition}
        }

        slippage := order.FilledPrice - order.ReferencePrice
        conditionMap[condition].AvgSlippage += slippage
        conditionMap[condition].OrderCount++
    }

    // Normalize
    for _, metrics := range conditionMap {
        metrics.AvgSlippage /= float64(metrics.OrderCount)
    }

    return conditionMap
}
```

#### Slippage Decomposition
```go
type SlippageComponent struct {
    SpreadComponent     float64 // Slippage due to bid-ask spread at execution
    AdverseComponent    float64 // Slippage due to price movement before execution
    ProcessingComponent float64 // Slippage due to order processing delays
}

func DecomposeSlippage(order Order, quoteAtSubmission, quoteAtExecution Quote) *SlippageComponent {
    totalSlippage := order.FilledPrice - quoteAtSubmission.Mid()

    // Component 1: Spread at execution
    spreadComponent := (quoteAtExecution.Ask - quoteAtExecution.Bid) / 2

    // Component 2: Adverse price movement
    adverseComponent := (quoteAtExecution.Mid() - quoteAtSubmission.Mid())

    // Component 3: Processing delay
    processingComponent := totalSlippage - spreadComponent - adverseComponent

    return &SlippageComponent{
        SpreadComponent:     spreadComponent,
        AdverseComponent:    adverseComponent,
        ProcessingComponent: processingComponent,
    }
}
```

#### Visualization Recommendations
- **Histogram**: Slippage distribution with bell curve overlay
- **Box Plot**: Slippage by symbol/LP/time period
- **Scatter Plot**: Order size vs. slippage (correlation analysis)
- **Heat Map**: Slippage matrix (Symbol × LP)

---

### 1.7 Volume-Weighted Average Latency (VWAL)

**Purpose**: Measure latency impact weighted by order size

#### Formula
```
VWAL = Σ(Order Latency × Order Volume) / Total Volume

Where:
  Order Latency = Time from submission to fill
  Order Volume = Size of the order in lots/units
  Total Volume = Sum of all order volumes
```

#### Implementation
```go
type VolumeWeightedLatency struct {
    VWAL              time.Duration
    SimpleAvgLatency  time.Duration // For comparison
    SmallOrderLatency time.Duration // <10 lots
    MediumOrderLatency time.Duration // 10-100 lots
    LargeOrderLatency time.Duration // >100 lots
    LatencyImpactCost float64 // Cost due to latency-induced slippage
}

func CalculateVWAL(orders []Order) *VolumeWeightedLatency {
    var totalLatencyWeight time.Duration
    var totalVolume float64
    var simpleLatencySum time.Duration

    // Segment by size
    var smallLatencies, mediumLatencies, largeLatencies []time.Duration

    for _, order := range orders {
        if order.FilledAt == nil {
            continue
        }

        latency := order.FilledAt.Sub(order.CreatedAt)
        totalLatencyWeight += latency * time.Duration(int64(order.Volume))
        totalVolume += order.Volume
        simpleLatencySum += latency

        // Segment
        if order.Volume < 10 {
            smallLatencies = append(smallLatencies, latency)
        } else if order.Volume < 100 {
            mediumLatencies = append(mediumLatencies, latency)
        } else {
            largeLatencies = append(largeLatencies, latency)
        }
    }

    vwal := totalLatencyWeight / time.Duration(int64(totalVolume))
    simpleAvg := simpleLatencySum / time.Duration(len(orders))

    return &VolumeWeightedLatency{
        VWAL:              vwal,
        SimpleAvgLatency:  simpleAvg,
        SmallOrderLatency: calculateMeanDuration(smallLatencies),
        MediumOrderLatency: calculateMeanDuration(mediumLatencies),
        LargeOrderLatency:  calculateMeanDuration(largeLatencies),
    }
}

// Cost impact analysis
func AnalyzeLatencyImpactCost(orders []Order) float64 {
    var totalCost float64

    // Baseline latency
    baselineLatency := time.Duration(50) * time.Millisecond

    for _, order := range orders {
        if order.FilledAt == nil {
            continue
        }

        actualLatency := order.FilledAt.Sub(order.CreatedAt)
        excessLatency := actualLatency - baselineLatency

        if excessLatency > 0 {
            // Estimate slippage from excess latency
            // Assumes 0.1 pips per 100ms of excess latency
            latencyCost := float64(excessLatency.Milliseconds()) / 100.0 * 0.1
            orderValue := order.Volume * order.FilledPrice
            totalCost += latencyCost * orderValue
        }
    }

    return totalCost
}
```

#### Visualization Recommendations
- **Time Series**: VWAL over time with trend line
- **Stacked Bar**: Latency contribution by order size
- **Comparison Chart**: VWAL vs. Simple Average Latency
- **Correlation Scatter**: Order size vs. latency

---

## 2. PREDICTIVE ANALYTICS

### 2.1 Forecasting Exposure Trends

**Purpose**: Predict future LP exposure patterns to optimize liquidity allocation

#### Approach: ARIMA Time Series Forecasting

```go
type ExposureForecast struct {
    Horizon         time.Duration
    ForecastedDates []time.Time
    ForecastedValues []float64
    ConfidenceIntervals []ConfidenceInterval
    TrendDirection  string // UP, DOWN, STABLE
    SeasonalPattern string // HOURLY, DAILY, WEEKLY
}

type ConfidenceInterval struct {
    Date      time.Time
    Lower95   float64
    Upper95   float64
    Forecast  float64
}

// ARIMA Parameters
type ARIMAParams struct {
    P int // AR order
    D int // Differencing order
    Q int // MA order
}

func ForecastExposure(historicalExposure []ExposurePoint, horizon int) *ExposureForecast {
    // 1. Test for stationarity (ADF test)
    // 2. Determine optimal ARIMA parameters using auto.arima equivalent
    params := AutoARIMA(historicalExposure)

    // 3. Fit ARIMA model
    model := FitARIMA(historicalExposure, params)

    // 4. Forecast and calculate confidence intervals
    forecast := model.Forecast(horizon)

    return forecast
}

// Alternative: Prophet for handling seasonality better
func ForecastExposureWithProphet(historicalData map[time.Time]float64) *ExposureForecast {
    // Install: github.com/oneofone/go-prophet
    // Handles seasonality, trend changes, and holidays

    // Build time series
    ts := make([]float64, 0)
    dates := make([]time.Time, 0)

    for date := range historicalData {
        dates = append(dates, date)
        ts = append(ts, historicalData[date])
    }

    // Prophet fitting and forecasting
    // ...

    return &ExposureForecast{}
}
```

#### Exposure Trend Analysis
```go
type ExposureTrendMetrics struct {
    CurrentExposure    float64
    TrendSlope         float64 // Change per day
    TrendStrength      float64 // R-squared value
    PredictedIn7Days   float64
    PredictedIn30Days  float64
    SeasonalityFactor  float64 // Magnitude of seasonal variation
    VolatilityForecast float64 // Expected exposure volatility
}

func AnalyzeExposureTrend(dailyExposure []DailyExposure) *ExposureTrendMetrics {
    // Linear regression for trend
    slope, intercept, rSquared := LinearRegression(dailyExposure)

    // Seasonality analysis (Fourier decomposition)
    trend, seasonal, residual := SeasonalDecomposition(dailyExposure)

    // Forecast using trend + seasonal
    forecast7 := slope*7 + intercept + seasonal[7]
    forecast30 := slope*30 + intercept + seasonal[30]

    return &ExposureTrendMetrics{
        TrendSlope:         slope,
        TrendStrength:      rSquared,
        PredictedIn7Days:   forecast7,
        PredictedIn30Days:  forecast30,
        SeasonalityFactor:  CalculateSeasonalityMagnitude(seasonal),
        VolatilityForecast: CalculateVolatilityForecast(residual),
    }
}
```

---

### 2.2 Anomaly Detection in Routing Patterns

**Purpose**: Identify unusual routing behaviors that may indicate problems

#### Isolation Forest Approach
```go
type RoutingAnomaly struct {
    RoutingPath         string
    AnomalyScore        float64 // 0-1, higher = more anomalous
    Severity            string  // LOW, MEDIUM, HIGH, CRITICAL
    DeviationType       string  // UNEXPECTED_LP_SELECTION, HIGH_SLIPPAGE, UNUSUAL_LATENCY
    ExpectedValue       float64
    ActualValue         float64
    PercentageDeviation float64
    Timestamp           time.Time
}

func DetectRoutingAnomalies(trades []Trade, historicalStats map[string]*RouteStatistics) []RoutingAnomaly {
    anomalies := []RoutingAnomaly{}

    for _, trade := range trades {
        routePath := trade.RoutingPath
        stats := historicalStats[routePath]

        if stats == nil {
            continue
        }

        // Check for anomalies
        if trade.Slippage > stats.AvgSlippage+3*stats.SlippageStdDev {
            // High slippage anomaly
            anomalies = append(anomalies, RoutingAnomaly{
                RoutingPath:         routePath,
                AnomalyScore:        CalculateAnomalyScore(trade.Slippage, stats),
                Severity:            "HIGH",
                DeviationType:       "HIGH_SLIPPAGE",
                ExpectedValue:       stats.AvgSlippage,
                ActualValue:         trade.Slippage,
                PercentageDeviation: (trade.Slippage - stats.AvgSlippage) / stats.AvgSlippage,
                Timestamp:           trade.ExecutedAt,
            })
        }

        if trade.FillTime > stats.AvgFillTime+3*stats.FillTimeStdDev {
            // High latency anomaly
            anomalies = append(anomalies, RoutingAnomaly{
                RoutingPath:         routePath,
                AnomalyScore:        CalculateAnomalyScore(float64(trade.FillTime.Milliseconds()), stats),
                Severity:            "MEDIUM",
                DeviationType:       "UNUSUAL_LATENCY",
                ExpectedValue:       float64(stats.AvgFillTime.Milliseconds()),
                ActualValue:         float64(trade.FillTime.Milliseconds()),
                PercentageDeviation: float64(trade.FillTime-stats.AvgFillTime) / float64(stats.AvgFillTime),
                Timestamp:           trade.ExecutedAt,
            })
        }
    }

    return anomalies
}

// Z-score based anomaly detection
func CalculateAnomalyScore(value, mean, stdDev float64) float64 {
    zScore := (value - mean) / stdDev
    // Convert to 0-1 scale using sigmoid
    return 1.0 / (1.0 + math.Exp(-zScore))
}

// Multivariate anomaly detection (Mahalanobis distance)
type MultivariatRouteAnomaly struct {
    RoutingPath       string
    MahalanobisDistance float64 // D >= 3 is anomalous
    AnomalyProbability float64
}

func DetectMultivariateAnomalies(trade Trade, historicalDistribution *RouteDistribution) float64 {
    // Use Mahalanobis distance in multi-dimensional feature space
    // Features: slippage, latency, fill rate, adverse selection

    features := []float64{
        trade.Slippage,
        float64(trade.FillTime.Milliseconds()),
        BoolToFloat(trade.IsFilled),
        trade.AdverseSelection,
    }

    distance := MahalanobisDistance(features, historicalDistribution.Mean, historicalDistribution.CovarianceMatrix)
    return distance
}
```

#### Detection Methods by Type

**Statistical Process Control (SPC)**
```go
func DetectShewhart(trades []Trade, k int) []int {
    // Upper Control Limit = mean + 3*stdDev
    // Lower Control Limit = mean - 3*stdDev

    mean := CalculateMean(trades)
    stdDev := CalculateStdDev(trades)

    UCL := mean + 3*stdDev
    LCL := mean - 3*stdDev

    anomalyIndices := []int{}
    for i, trade := range trades {
        if trade.Metric > UCL || trade.Metric < LCL {
            anomalyIndices = append(anomalyIndices, i)
        }
    }

    return anomalyIndices
}

// CUSUM (Cumulative Sum Control Chart)
func DetectCUSUM(trades []Trade, targetValue, k, h float64) []int {
    cusum := 0.0
    anomalies := []int{}

    for i, trade := range trades {
        cusum += (trade.Metric - targetValue)

        if cusum > h || cusum < -h {
            anomalies = append(anomalies, i)
            cusum = 0 // Reset
        }
    }

    return anomalies
}
```

---

### 2.3 Client Behavior Prediction

**Purpose**: Predict client trading patterns to optimize routing decisions

```go
type ClientBehaviorProfile struct {
    ClientID              string
    PreferredTimeWindow   string // LONDON, NY, ASIA, ALL_DAY
    TypicalOrderSize      float64
    PreferredSymbols      []string
    HedgingVsSpeculation  float64 // 0 = pure speculation, 1 = pure hedging
    AggregationRatio      float64 // % of orders that are order fills vs new orders
    ToxicityTrend         string  // INCREASING, DECREASING, STABLE
}

func PredictClientBehavior(client *Client, trainingData []Trade) *ClientBehaviorProfile {
    // 1. Identify preferred trading windows
    timeWindows := IdentifyTradingWindows(trainingData)

    // 2. Calculate average order size
    avgOrderSize := CalculateAverageOrderSize(trainingData)

    // 3. Extract preferred symbols
    symbolFrequency := CountSymbolFrequency(trainingData)
    topSymbols := GetTopN(symbolFrequency, 5)

    // 4. Determine hedging vs speculation (using position closure analysis)
    hedgingRatio := CalculateHedgingRatio(trainingData)

    // 5. Calculate order aggregation ratio
    aggregationRatio := CalculateAggregationRatio(trainingData)

    // 6. Analyze toxicity trend (logistic regression)
    toxicityTrend := PredictToxicityTrend(trainingData)

    return &ClientBehaviorProfile{
        ClientID:              client.ID,
        PreferredTimeWindow:   timeWindows,
        TypicalOrderSize:      avgOrderSize,
        PreferredSymbols:      topSymbols,
        HedgingVsSpeculation:  hedgingRatio,
        AggregationRatio:      aggregationRatio,
        ToxicityTrend:         toxicityTrend,
    }
}

// Predict next order parameters
type NextOrderPrediction struct {
    PredictedSymbol     string
    PredictedSize       float64
    PredictedSide       string // BUY or SELL
    PredictedTime       time.Time
    PredictionConfidence float64
}

func PredictNextOrder(clientProfile *ClientBehaviorProfile, recentContext []Trade) *NextOrderPrediction {
    // Use Markov chain for symbol transitions
    nextSymbol := PredictSymbolTransition(clientProfile, recentContext)

    // Use regression for order size
    nextSize := PredictOrderSize(clientProfile, recentContext)

    // Use classification for side (BUY vs SELL)
    nextSide := PredictOrderSide(clientProfile, recentContext)

    // Use time series for order timing
    nextTime := PredictOrderTiming(clientProfile, recentContext)

    return &NextOrderPrediction{
        PredictedSymbol:      nextSymbol,
        PredictedSize:        nextSize,
        PredictedSide:        nextSide,
        PredictedTime:        nextTime,
        PredictionConfidence: CalculateConfidence(clientProfile, recentContext),
    }
}
```

---

### 2.4 LP Performance Forecasting

**Purpose**: Predict LP quality degradation and optimal capacity allocation

```go
type LPPerformanceForecast struct {
    LPID              string
    ForecastWindow    time.Duration
    PredictedHealth   float64 // 0-1
    PredictedFillRate float64
    PredictedSlippage float64
    RiskOfDisconnection float64
    RecommendedCapacity float64 // % of total orders to route
}

// Machine Learning: Random Forest for LP performance
func ForecastLPPerformance(lpID string, historicalMetrics []LPMetrics) *LPPerformanceForecast {
    // Features for prediction
    features := ExtractFeatures(historicalMetrics)

    // Use Random Forest trained on historical patterns
    predictions := randomForestModel.Predict(features)

    // Extract components
    predictedHealth := predictions[0]
    predictedFillRate := predictions[1]
    predictedSlippage := predictions[2]

    // Risk of disconnection (binary classification)
    disconnectionRisk := LogisticRegression(features)

    // Optimize capacity allocation
    optimalCapacity := AllocateCapacityByPerformance(lpID, predictedHealth)

    return &LPPerformanceForecast{
        LPID:                  lpID,
        PredictedHealth:       predictedHealth,
        PredictedFillRate:     predictedFillRate,
        PredictedSlippage:     predictedSlippage,
        RiskOfDisconnection:   disconnectionRisk,
        RecommendedCapacity:   optimalCapacity,
    }
}

// Real-time degradation warning system
func MonitorLPDegradation(lpMetrics LPMetrics, baseline LPMetrics) *DegradationAlert {
    alert := &DegradationAlert{
        LPID:            lpMetrics.LPID,
        TriggeredAt:     time.Now(),
    }

    // Check each metric
    if lpMetrics.HealthScore < baseline.HealthScore * 0.8 {
        alert.Severity = "HIGH"
        alert.Message = fmt.Sprintf("Health score degraded by %.1f%%",
            (1 - lpMetrics.HealthScore/baseline.HealthScore) * 100)
    }

    if lpMetrics.AvgSlippage > baseline.AvgSlippage * 1.5 {
        alert.Severity = "MEDIUM"
        alert.Message = fmt.Sprintf("Slippage increased by %.1f%%",
            (lpMetrics.AvgSlippage/baseline.AvgSlippage - 1) * 100)
    }

    if lpMetrics.RejectRate > baseline.RejectRate + 0.05 {
        alert.Severity = "CRITICAL"
        alert.Message = "Rejection rate spike detected"
    }

    return alert
}
```

---

## 3. COMPARATIVE ANALYTICS

### 3.1 A/B Testing Framework for Routing Rules

**Purpose**: Statistically validate routing rule changes before full deployment

```go
type ABTest struct {
    TestID           string
    Name             string
    StartTime        time.Time
    EndTime          time.Time
    ControlRule      RoutingRule
    VariantRule      RoutingRule
    SampleSize       int64
    ControlMetrics   *RouteMetrics
    VariantMetrics   *RouteMetrics
    PValue           float64
    EffectSize       float64
    Winner           string // "CONTROL", "VARIANT", "INCONCLUSIVE"
    Recommendation   string
}

func RunABTest(controlTrades, variantTrades []Trade) *ABTest {
    control := CalculateMetrics(controlTrades)
    variant := CalculateMetrics(variantTrades)

    // Perform t-test
    pValue := TwoSampleTTest(control.Slippages, variant.Slippages)

    // Calculate effect size (Cohen's d)
    effectSize := CohensDEffect(control.AvgSlippage, variant.AvgSlippage,
        control.SlippageStdDev, variant.SlippageStdDev)

    // Determine winner
    winner := DetermineWinner(pValue, effectSize, control, variant)

    return &ABTest{
        ControlMetrics:   control,
        VariantMetrics:   variant,
        PValue:           pValue,
        EffectSize:       effectSize,
        Winner:           winner,
        Recommendation:   GenerateRecommendation(winner, pValue),
    }
}

// Statistical tests
func TwoSampleTTest(sample1, sample2 []float64) float64 {
    // Welch's t-test (doesn't assume equal variances)
    mean1 := CalculateMean(sample1)
    mean2 := CalculateMean(sample2)
    var1 := CalculateVariance(sample1)
    var2 := CalculateVariance(sample2)

    t := (mean1 - mean2) / math.Sqrt(var1/float64(len(sample1)) + var2/float64(len(sample2)))

    // Calculate p-value from t-distribution
    df := math.Pow((var1/float64(len(sample1)) + var2/float64(len(sample2))), 2) /
        (math.Pow(var1/float64(len(sample1)), 2)/(float64(len(sample1))-1) +
            math.Pow(var2/float64(len(sample2)), 2)/(float64(len(sample2))-1))

    pValue := TDistributionPValue(t, df)
    return pValue
}

// Effect size
func CohensDEffect(mean1, mean2, std1, std2 float64) float64 {
    pooledStd := math.Sqrt((std1*std1 + std2*std2) / 2)
    return (mean1 - mean2) / pooledStd
}
```

#### Multi-Metric Composite Score
```go
type CompositeScore struct {
    Slippage         float64 // Weight: 40%
    FillRate         float64 // Weight: 30%
    Latency          float64 // Weight: 20%
    ClientSatisfaction float64 // Weight: 10%
    CompositeScore   float64
}

func CalculateCompositeScore(metrics *RouteMetrics, weights map[string]float64) float64 {
    // Normalize each metric to 0-1
    normalizedSlippage := 1.0 - math.Min(metrics.AvgSlippage/0.5, 1.0) // 0.5 pips = threshold
    normalizedFillRate := metrics.FillRate
    normalizedLatency := 1.0 - math.Min(float64(metrics.AvgFillTime.Milliseconds())/500.0, 1.0)

    composite := (normalizedSlippage * weights["slippage"]) +
        (normalizedFillRate * weights["fillRate"]) +
        (normalizedLatency * weights["latency"])

    return composite
}
```

---

### 3.2 Before/After Rule Change Analysis

**Purpose**: Measure the impact of routing rule modifications

```go
type RuleChangeAnalysis struct {
    RuleID           string
    RuleName         string
    ChangeDescription string
    BeforePeriod     DateRange
    AfterPeriod      DateRange
    BeforeMetrics    *DetailedMetrics
    AfterMetrics     *DetailedMetrics
    Improvement      map[string]float64 // % change per metric
    StatisticalSignificance float64
    ROI              float64 // Return on implementation cost
}

type DetailedMetrics struct {
    AvgSlippage         float64
    SlippageStdDev      float64
    FillRate            float64
    AvgFillTime         time.Duration
    SuccessRate         float64
    ClientSatisfaction  float64
    TotalP&L            float64
    VolumeProcessed     float64
}

func AnalyzeRuleChange(beforeTrades, afterTrades []Trade, implementationCost float64) *RuleChangeAnalysis {
    before := CalculateDetailedMetrics(beforeTrades)
    after := CalculateDetailedMetrics(afterTrades)

    // Calculate improvements
    improvements := map[string]float64{
        "slippage": ((before.AvgSlippage - after.AvgSlippage) / before.AvgSlippage) * 100,
        "fillRate": ((after.FillRate - before.FillRate) / before.FillRate) * 100,
        "latency": ((before.AvgFillTime.Milliseconds() - after.AvgFillTime.Milliseconds()) /
            before.AvgFillTime.Milliseconds()) * 100,
        "pnl": ((after.TotalP&L - before.TotalP&L) / before.TotalP&L) * 100,
    }

    // Calculate ROI
    improvementValue := (after.TotalP&L - before.TotalP&L)
    roi := (improvementValue - implementationCost) / implementationCost

    // Statistical significance
    pValue := TwoSampleTTest(extractSlippages(beforeTrades), extractSlippages(afterTrades))

    return &RuleChangeAnalysis{
        BeforeMetrics:              before,
        AfterMetrics:               after,
        Improvement:                improvements,
        StatisticalSignificance:    pValue,
        ROI:                        roi,
    }
}
```

---

### 3.3 LP Cost Comparison

**Purpose**: Analyze cost-benefit tradeoffs across LPs

```go
type LPCostAnalysis struct {
    LPID                  string
    DirectCosts           float64 // Commission, spreads
    IndirectCosts         float64 // Slippage, latency impact
    TotalCostPerTrade     float64
    CostAsPercentageOfVolume float64
    CostPerMillionVolume  float64 // Normalized metric
    Quality               float64 // Fill rate, slippage, latency
    CostEfficiencyRatio   float64 // Quality / Total Cost
}

func AnalyzeLPCosts(lpTrades []Trade, lpCommissionStructure *CommissionStructure) *LPCostAnalysis {
    var directCosts, slippageCosts, latencyCosts float64
    var totalVolume float64

    for _, trade := range lpTrades {
        // Direct costs
        commission := lpCommissionStructure.GetCommission(trade.Symbol, trade.Volume)
        spread := trade.ExecutedPrice - trade.ReferencePrice
        directCosts += commission + spread

        // Indirect costs
        slippageCosts += math.Abs(trade.Slippage) * trade.Volume
        latencyCosts += CalculateLatencyCost(trade.FillTime)

        totalVolume += trade.Volume * trade.ExecutedPrice
    }

    totalCost := directCosts + slippageCosts + latencyCosts
    costPerTrade := totalCost / float64(len(lpTrades))
    costPercentage := (totalCost / totalVolume) * 100

    // Quality metrics
    quality := CalculateLPQuality(lpTrades)

    // Cost efficiency ratio
    costEfficiency := quality / totalCost

    return &LPCostAnalysis{
        DirectCosts:                  directCosts,
        IndirectCosts:                slippageCosts + latencyCosts,
        TotalCostPerTrade:            costPerTrade,
        CostAsPercentageOfVolume:     costPercentage,
        CostPerMillionVolume:         (totalCost / (totalVolume / 1000000)),
        Quality:                      quality,
        CostEfficiencyRatio:          costEfficiency,
    }
}

// Advanced: Cost attribution model
type CostAttribution struct {
    SpreadCost         float64 // % of total cost
    CommissionCost     float64
    SlippageCost       float64
    LatencyCost        float64
    AdverseSelectionCost float64
    ProcessingCost     float64
}

func AttributeCosts(trade Trade, refQuote Quote) *CostAttribution {
    totalCost := math.Abs(trade.Slippage) * trade.Volume

    // Spread component
    spreadComponent := (refQuote.Ask - refQuote.Bid) / 2 * trade.Volume

    // Commission
    commissionComponent := trade.Commission

    // Adverse selection
    adverseComponent := calculateAdverseSelection(trade, refQuote) * trade.Volume

    // Processing/latency
    processingComponent := CalculateLatencyCost(trade.FillTime)

    return &CostAttribution{
        SpreadCost:           (spreadComponent / totalCost) * 100,
        CommissionCost:       (commissionComponent / totalCost) * 100,
        SlippageCost:         ((math.Abs(trade.Slippage) - spreadComponent) / totalCost) * 100,
        LatencyCost:          (processingComponent / totalCost) * 100,
        AdverseSelectionCost: (adverseComponent / totalCost) * 100,
    }
}
```

---

### 3.4 Routing Path Profitability Analysis

**Purpose**: Calculate detailed profitability metrics for each routing decision

```go
type RoutingPathProfitability struct {
    PathID              string
    ExecutionSide       string
    TargetLP            string
    SessionID           string
    TotalTrades         int64
    TotalVolume         float64
    TotalRealizedPnL    float64
    AveragePnL          float64
    AveragePnLPerLot    float64
    PnLPerMillionVolume float64
    WinRate             float64
    ProfitFactor        float64
    Sortino             float64 // Downside risk-adjusted return
    Calmar              float64 // Return / Max Drawdown
    RecommendedAllocation float64 // % of orders to route this way
}

func AnalyzeRoutingPathProfitability(pathTrades []Trade) *RoutingPathProfitability {
    if len(pathTrades) == 0 {
        return nil
    }

    var totalPnL, totalVolume float64
    var winCount int
    var downturns []float64

    for _, trade := range pathTrades {
        totalPnL += trade.RealizedPnL
        totalVolume += trade.Volume

        if trade.RealizedPnL > 0 {
            winCount++
        } else if trade.RealizedPnL < 0 {
            downturns = append(downturns, trade.RealizedPnL)
        }
    }

    // Calculate metrics
    avgPnL := totalPnL / float64(len(pathTrades))
    avgPnLPerLot := totalPnL / totalVolume
    winRate := float64(winCount) / float64(len(pathTrades))

    // Profit Factor
    var grossProfit, grossLoss float64
    for _, trade := range pathTrades {
        if trade.RealizedPnL > 0 {
            grossProfit += trade.RealizedPnL
        } else {
            grossLoss += math.Abs(trade.RealizedPnL)
        }
    }
    profitFactor := grossProfit / grossLoss

    // Sortino Ratio
    downturnsVariance := CalculateVariance(downturns)
    downsideDeviation := math.Sqrt(downturnsVariance)
    sortino := avgPnL / downsideDeviation * math.Sqrt(252) // Annualized

    // Calmar Ratio
    maxDrawdown, _ := CalculateMaxDrawdown(pathTrades)
    calmar := (totalPnL / float64(len(pathTrades))) / math.Abs(maxDrawdown)

    // Optimal allocation (Kelly Criterion)
    kellyFraction := (winRate * (avgPnL / (avgPnL + math.Abs(avgPnL)))) - ((1 - winRate) / 1)
    optimalAllocation := math.Max(0, math.Min(kellyFraction, 0.25)) // Cap at 25%

    return &RoutingPathProfitability{
        TotalTrades:         int64(len(pathTrades)),
        TotalVolume:         totalVolume,
        TotalRealizedPnL:    totalPnL,
        AveragePnL:          avgPnL,
        AveragePnLPerLot:    avgPnLPerLot,
        PnLPerMillionVolume: (totalPnL / (totalVolume / 1000000)),
        WinRate:             winRate,
        ProfitFactor:        profitFactor,
        Sortino:             sortino,
        Calmar:              calmar,
        RecommendedAllocation: optimalAllocation,
    }
}

// Kelly Criterion for optimal position sizing
func KellyCriterion(winRate, winSize, lossSize float64) float64 {
    // f* = (p * b - q) / b
    // where p = win rate, q = loss rate, b = win/loss ratio
    q := 1 - winRate
    b := winSize / lossSize
    f := ((winRate * b) - q) / b

    // Fractional Kelly (use 25% of Kelly for safety)
    return math.Max(0, f * 0.25)
}
```

---

## 4. RISK METRICS

### 4.1 Value at Risk (VaR) by Routing Path

**Purpose**: Measure maximum loss at specified confidence level

#### Methods

**Historical VaR**
```go
type VaRMetrics struct {
    Confidence          float64 // e.g., 0.95 for 95%
    Horizon             time.Duration // 1 day, 10 days, etc.
    VaR                 float64 // In USD
    ConditionalVaR      float64 // Average loss beyond VaR
    VaRAsPercentage     float64 // % of portfolio value
    ExpectedShortfall   float64 // CVaR
}

func CalculateHistoricalVaR(pathTrades []Trade, confidence float64) *VaRMetrics {
    // Sort trades by P/L
    var pnls []float64
    for _, trade := range pathTrades {
        pnls = append(pnls, trade.RealizedPnL)
    }
    sort.Float64s(pnls)

    // Find VaR at confidence level
    index := int(float64(len(pnls)) * (1 - confidence))
    varValue := pnls[index]

    // Calculate Conditional VaR (average of losses beyond VaR)
    var conditionalLosses []float64
    for _, pnl := range pnls[:index+1] {
        if pnl < varValue {
            conditionalLosses = append(conditionalLosses, pnl)
        }
    }
    cvar := CalculateMean(conditionalLosses)

    return &VaRMetrics{
        Confidence:        confidence,
        VaR:              varValue,
        ConditionalVaR:   cvar,
        ExpectedShortfall: cvar,
    }
}
```

**Parametric VaR (Variance-Covariance)**
```go
func CalculateParametricVaR(mean, stdDev, confidence, portfolioValue float64) float64 {
    // VaR = Portfolio Value * (Mean + Z * StdDev)
    // Z for 95% = 1.645, for 99% = 2.33
    zScore := NormalInverse(1 - confidence)
    var := portfolioValue * (mean + (zScore * stdDev))
    return var
}
```

**Cornish-Fisher VaR (adjusts for skewness and kurtosis)**
```go
func CalculateCornishFisherVaR(returns []float64, confidence float64) float64 {
    mean := CalculateMean(returns)
    stdDev := CalculateStdDev(returns)
    skewness := CalculateSkewness(returns)
    kurtosis := CalculateKurtosis(returns)

    zScore := NormalInverse(1 - confidence)

    // CF-adjusted z-score
    w := zScore +
        (zScore*zScore - 1) / 6 * skewness +
        (zScore*zScore*zScore - 3*zScore) / 24 * kurtosis -
        (2*zScore*zScore*zScore - 5*zScore) / 36 * skewness*skewness

    return mean + w*stdDev
}
```

---

### 4.2 Concentration Risk by LP

**Purpose**: Measure risk of over-reliance on single LP

```go
type ConcentrationRisk struct {
    LPConcentration     map[string]float64 // % exposure per LP
    HerfindahlIndex     float64 // 0.33 = perfect concentration (3 LPs equal), 1.0 = single LP
    GiniCoefficient     float64 // 0-1, measures inequality
    TopLPExposure       float64 // % exposure to top LP
    Top3LPExposure      float64 // % exposure to top 3 LPs
    DiversificationRatio float64 // Number of effective LPs
    CorrelationRisk     float64 // Correlation between LP quality metrics
}

func AnalyzeConcentrationRisk(trades []Trade) *ConcentrationRisk {
    // Calculate LP exposure percentages
    lpVolumes := make(map[string]float64)
    totalVolume := 0.0

    for _, trade := range trades {
        lpVolumes[trade.SelectedLP] += trade.Volume
        totalVolume += trade.Volume
    }

    lpExposure := make(map[string]float64)
    for lp, volume := range lpVolumes {
        lpExposure[lp] = volume / totalVolume
    }

    // Herfindahl Index
    hhi := 0.0
    for _, exposure := range lpExposure {
        hhi += exposure * exposure
    }

    // Gini Coefficient
    gini := CalculateGiniCoefficient(lpExposure)

    // Top LP exposure
    topLP := 0.0
    top3LP := 0.0

    exposures := make([]float64, 0, len(lpExposure))
    for _, exp := range lpExposure {
        exposures = append(exposures, exp)
    }
    sort.Float64s(exposures)

    for i := 1; i <= 3 && i <= len(exposures); i++ {
        if i == 1 {
            topLP = exposures[len(exposures)-1]
        }
        top3LP += exposures[len(exposures)-i]
    }

    // Diversification Ratio
    divRatio := float64(1) / hhi

    return &ConcentrationRisk{
        LPConcentration:      lpExposure,
        HerfindahlIndex:      hhi,
        GiniCoefficient:      gini,
        TopLPExposure:        topLP,
        Top3LPExposure:       top3LP,
        DiversificationRatio: divRatio,
    }
}

// Herfindahl Index interpretation
// HHI < 0.15: Low concentration
// 0.15 < HHI < 0.25: Moderate concentration
// HHI > 0.25: High concentration
```

---

### 4.3 Counterparty Exposure Limits

**Purpose**: Monitor and enforce counterparty risk limits

```go
type CounterpartyExposure struct {
    CounterpartyID          string
    CurrentExposure         float64 // USD
    ExposureLimit           float64
    UtilizationPercentage   float64
    OpenPositions           int64
    PotentialLoss           float64 // If counterparty fails
    NetPresent Value        float64
    CreditQualityScore      float64 // 0-1
    EarlyWarningThreshold   float64 // Alert when 75% utilized
}

func MonitorCounterpartyExposure(trades []Trade, limits map[string]float64) []CounterpartyExposure {
    exposures := make(map[string]*CounterpartyExposure)

    for _, trade := range trades {
        if trade.Status == "OPEN" {
            counterparty := trade.CounterpartyID
            if _, exists := exposures[counterparty]; !exists {
                exposures[counterparty] = &CounterpartyExposure{
                    CounterpartyID:  counterparty,
                    ExposureLimit:   limits[counterparty],
                }
            }

            exposure := exposures[counterparty]
            exposure.CurrentExposure += trade.NotionalValue()
            exposure.OpenPositions++

            // Calculate potential loss
            if trade.UnrealizedPnL < 0 {
                exposure.PotentialLoss += math.Abs(trade.UnrealizedPnL)
            }
        }
    }

    // Calculate utilization
    result := make([]CounterpartyExposure, 0)
    for _, exp := range exposures {
        exp.UtilizationPercentage = (exp.CurrentExposure / exp.ExposureLimit) * 100

        // Trigger alerts
        if exp.UtilizationPercentage > 75 {
            log.Printf("ALERT: Counterparty %s exposure at %.1f%%",
                exp.CounterpartyID, exp.UtilizationPercentage)
        }

        result = append(result, *exp)
    }

    return result
}
```

---

### 4.4 Margin Utilization by Book

**Purpose**: Track margin usage across accounts and ensure buffer

```go
type MarginMetrics struct {
    TotalEquity         float64
    UsedMargin          float64
    FreeMargin          float64
    MarginLevel         float64 // Used / Total
    MarginBuffer        float64 // Free / Used
    EquityCurve         []float64
    MaxMarginUsage      float64
    TimeToMarginCall    time.Duration
    RequiredBufferLevel float64 // e.g., 30% free margin minimum
}

func CalculateMarginMetrics(accounts []Account) *MarginMetrics {
    var totalEquity, usedMargin float64

    for _, account := range accounts {
        totalEquity += account.Equity
        usedMargin += account.Margin
    }

    freeMargin := totalEquity - usedMargin
    marginLevel := usedMargin / totalEquity
    marginBuffer := freeMargin / usedMargin

    // Predict time to margin call (if losing money at current rate)
    timeToMarginCall := time.Duration(math.MaxInt64)
    if marginBuffer > 0 {
        daysToCall := (freeMargin / totalEquity) / 0.01 // Assuming 1% daily loss rate
        timeToMarginCall = time.Duration(daysToCall) * 24 * time.Hour
    }

    return &MarginMetrics{
        TotalEquity:        totalEquity,
        UsedMargin:         usedMargin,
        FreeMargin:         freeMargin,
        MarginLevel:        marginLevel,
        MarginBuffer:       marginBuffer,
        TimeToMarginCall:   timeToMarginCall,
        RequiredBufferLevel: 0.30, // 30% minimum
    }
}

// Margin warning levels
type MarginAlert struct {
    Level      string // "INFO", "WARNING", "CRITICAL"
    Threshold  float64 // Margin utilization %
    Message    string
}

var marginAlerts = []MarginAlert{
    {Level: "INFO", Threshold: 0.70, Message: "Approaching margin limit"},
    {Level: "WARNING", Threshold: 0.85, Message: "High margin utilization"},
    {Level: "CRITICAL", Threshold: 0.95, Message: "Imminent margin call risk"},
}
```

---

## 5. VISUALIZATION DASHBOARDS

### 5.1 Real-Time Routing Analytics Dashboard

**Key Components:**
1. **KPI Cards**: Current Sharpe Ratio, Profit Factor, Max Drawdown
2. **Routing Heatmap**: Performance matrix (LP × Symbol × Time)
3. **Execution Quality Trend**: Fill time and slippage over time
4. **Client Toxicity Distribution**: Histogram of toxicity scores
5. **LP Health Scorecard**: Color-coded health metrics
6. **P&L Attribution**: Pie chart breakdown by routing path
7. **Risk Gauge**: Margin utilization, concentration risk

### 5.2 Advanced Metrics Visualization

**Time Series Charts:**
- Sharpe Ratio evolution
- Profit Factor trends
- Equity curve (with drawdown shading)
- VWAL over time

**Comparative Charts:**
- LP cost comparison (bubble chart)
- Routing path profitability matrix
- Before/after rule change analysis

**Distribution Charts:**
- Slippage distribution (histogram)
- Fill time percentiles (box plot)
- Client toxicity segmentation (violin plot)

**Risk Dashboards:**
- VaR heat map (routing path × confidence level)
- Concentration risk gauge
- Counterparty exposure tracker
- Margin buffer visualization

---

## 6. IMPLEMENTATION ROADMAP

### Phase 1: Foundation Metrics (Weeks 1-4)
- Advanced Sharpe Ratio calculation
- Profit Factor by routing path
- Basic drawdown analysis
- Fill time metrics decomposition

### Phase 2: Predictive Analytics (Weeks 5-8)
- Exposure trend forecasting
- Anomaly detection system
- Client behavior profiling
- LP performance prediction

### Phase 3: Comparative Framework (Weeks 9-12)
- A/B testing infrastructure
- Rule change analysis tools
- LP cost comparison system
- Routing path profitability analysis

### Phase 4: Risk Management (Weeks 13-16)
- VaR calculations
- Concentration risk monitoring
- Counterparty exposure tracking
- Margin utilization alerts

### Phase 5: Visualization & Dashboards (Weeks 17-20)
- Real-time analytics dashboards
- Custom reporting tools
- Alert system integration
- Historical backtesting interface

---

## 7. DATA REQUIREMENTS & COLLECTION

### Data Points to Log for Each Trade
```
- Timestamp (submission, execution, confirmation)
- Client ID
- Symbol
- Side (BUY/SELL)
- Volume
- Reference price (best available)
- Executed price (actual fill)
- Selected LP
- Routing decision (which rule/logic)
- Order type
- Time-in-force
- SL/TP levels
- Status (filled, rejected, partial)
- Rejection reason (if applicable)
```

### Data Points for LP Quotes
```
- Timestamp
- LP ID
- Symbol
- Bid/Ask prices
- Bid/Ask sizes
- Spread
- Connection status
```

### Data Points for Accounts
```
- Account ID
- Equity curve (daily)
- Used/Free margin
- Open positions count
- Daily trade count
- Daily P&L
- Login times (for client behavior)
```

---

## 8. BEST PRACTICES & GUARDRAILS

1. **Minimum Sample Size**: Require 100+ trades per metric for statistical validity
2. **Backtest Period**: At least 3-6 months for meaningful pattern analysis
3. **Confidence Levels**: Use 95% or 99% for VaR; 0.05 p-value for tests
4. **Effect Size Thresholds**: Require Cohen's d > 0.2 for meaningful improvements
5. **Rebalance Frequency**: Review routing allocations monthly
6. **Risk Limits**: Auto-pause routing paths if VaR exceeds threshold
7. **Alert Thresholds**: Set actionable alerts (not too sensitive, not too loose)

---

## Conclusion

This advanced analytics framework provides comprehensive metrics, predictive models, and risk management tools for sophisticated broker routing systems. Implement these metrics progressively, validating each component against real trading data and refining based on business impact.

