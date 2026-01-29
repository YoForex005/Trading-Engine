package risk

import "time"

// MarginCalculationType defines the method for margin calculation
type MarginCalculationType string

const (
	MarginRetail    MarginCalculationType = "RETAIL"     // Fixed percentage per instrument
	MarginPortfolio MarginCalculationType = "PORTFOLIO"  // Cross-margining
	MarginSPAN      MarginCalculationType = "SPAN"       // For futures
)

// RiskLimitType defines different types of risk limits
type RiskLimitType string

const (
	LimitDailyLoss      RiskLimitType = "DAILY_LOSS"
	LimitMaxDrawdown    RiskLimitType = "MAX_DRAWDOWN"
	LimitPositionSize   RiskLimitType = "POSITION_SIZE"
	LimitExposure       RiskLimitType = "EXPOSURE"
	LimitLeverage       RiskLimitType = "LEVERAGE"
	LimitConcentration  RiskLimitType = "CONCENTRATION"
)

// RiskLevel defines the severity of a risk event
type RiskLevel string

const (
	RiskLevelNone     RiskLevel = "NONE"
	RiskLevelLow      RiskLevel = "LOW"
	RiskLevelMedium   RiskLevel = "MEDIUM"
	RiskLevelHigh     RiskLevel = "HIGH"
	RiskLevelCritical RiskLevel = "CRITICAL"
)

// CircuitBreakerStatus defines the status of circuit breakers
type CircuitBreakerStatus string

const (
	CircuitNormal   CircuitBreakerStatus = "NORMAL"
	CircuitWarning  CircuitBreakerStatus = "WARNING"
	CircuitTripped  CircuitBreakerStatus = "TRIPPED"
	CircuitDisabled CircuitBreakerStatus = "DISABLED"
)

// LiquidationPriority defines the priority for closing positions
type LiquidationPriority string

const (
	LiquidationLargestLoss    LiquidationPriority = "LARGEST_LOSS"
	LiquidationHighestMargin  LiquidationPriority = "HIGHEST_MARGIN"
	LiquidationOldestPosition LiquidationPriority = "OLDEST_POSITION"
	LiquidationLowestProfit   LiquidationPriority = "LOWEST_PROFIT"
)

// InstrumentRiskParams contains risk parameters for a specific instrument
type InstrumentRiskParams struct {
	Symbol             string  `json:"symbol"`
	MaxLeverage        int     `json:"maxLeverage"`        // Maximum leverage allowed
	MarginRequirement  float64 `json:"marginRequirement"`  // Margin % required (e.g., 1% = 100:1)
	MaxPositionSize    float64 `json:"maxPositionSize"`    // Maximum position size in lots
	MaxExposure        float64 `json:"maxExposure"`        // Maximum exposure in base currency
	VolatilityLimit    float64 `json:"volatilityLimit"`    // Circuit breaker threshold
	AllowNewPositions  bool    `json:"allowNewPositions"`  // Can open new positions
	RequireStopLoss    bool    `json:"requireStopLoss"`    // SL required for new trades
	MaxSlippage        float64 `json:"maxSlippage"`        // Max allowed slippage in pips
	TradingSessionOnly bool    `json:"tradingSessionOnly"` // Only during market hours
}

// ClientRiskProfile contains risk limits for a specific client
type ClientRiskProfile struct {
	ClientID           string                `json:"clientId"`
	RiskTier           string                `json:"riskTier"` // RETAIL/PROFESSIONAL/INSTITUTIONAL
	MaxLeverage        int                   `json:"maxLeverage"`
	DailyLossLimit     float64               `json:"dailyLossLimit"`     // Max daily loss in account currency
	MaxDrawdownPercent float64               `json:"maxDrawdownPercent"` // Max drawdown from peak equity
	MaxPositions       int                   `json:"maxPositions"`       // Max concurrent positions
	MaxExposurePercent float64               `json:"maxExposurePercent"` // Max total exposure as % of equity
	MarginCallLevel    float64               `json:"marginCallLevel"`    // Margin level % for margin call
	StopOutLevel       float64               `json:"stopOutLevel"`       // Margin level % for auto-liquidation
	AllowHedging       bool                  `json:"allowHedging"`
	AllowScalping      bool                  `json:"allowScalping"`
	RequireStopLoss    bool                  `json:"requireStopLoss"`
	CreditLimit        float64               `json:"creditLimit"`        // Maximum credit exposure
	InstrumentLimits   map[string]float64    `json:"instrumentLimits"`   // Per-instrument position limits
	MarginMethod       MarginCalculationType `json:"marginMethod"`
	ApiRateLimit       int                   `json:"apiRateLimit"`       // Max API calls per second
	MaxOrderSize       float64               `json:"maxOrderSize"`       // Max single order size
	FatFingerThreshold float64               `json:"fatFingerThreshold"` // Reject orders > X% of typical size
}

// ExposureMetrics contains real-time exposure calculations
type ExposureMetrics struct {
	AccountID          int64              `json:"accountId"`
	TotalExposure      float64            `json:"totalExposure"`      // Total notional exposure
	NetExposure        float64            `json:"netExposure"`        // Net exposure (longs - shorts)
	GrossExposure      float64            `json:"grossExposure"`      // Gross exposure (longs + shorts)
	ExposureBySymbol   map[string]float64 `json:"exposureBySymbol"`   // Per-instrument exposure
	ExposureByAsset    map[string]float64 `json:"exposureByAsset"`    // Per-asset-class exposure
	Delta              float64            `json:"delta"`              // Total delta
	Gamma              float64            `json:"gamma"`              // Total gamma
	Vega               float64            `json:"vega"`               // Total vega
	ConcentrationRisk  float64            `json:"concentrationRisk"`  // 0-1 score
	CorrelationRisk    float64            `json:"correlationRisk"`    // 0-1 score
	ValueAtRisk95      float64            `json:"valueAtRisk95"`      // 95% VaR
	ValueAtRisk99      float64            `json:"valueAtRisk99"`      // 99% VaR
	LiquidityScore     float64            `json:"liquidityScore"`     // 0-1, higher = more liquid
	Timestamp          time.Time          `json:"timestamp"`
}

// MarginCall represents a margin call event
type MarginCall struct {
	ID              string    `json:"id"`
	AccountID       int64     `json:"accountId"`
	MarginLevel     float64   `json:"marginLevel"`
	RequiredDeposit float64   `json:"requiredDeposit"` // Amount needed to restore margin
	Severity        RiskLevel `json:"severity"`
	TriggeredAt     time.Time `json:"triggeredAt"`
	ResolvedAt      *time.Time `json:"resolvedAt,omitempty"`
	Status          string    `json:"status"` // ACTIVE/RESOLVED/LIQUIDATED
	Actions         []string  `json:"actions"` // Actions taken
}

// LiquidationEvent represents an auto-liquidation
type LiquidationEvent struct {
	ID                string              `json:"id"`
	AccountID         int64               `json:"accountId"`
	TriggerReason     string              `json:"triggerReason"`
	MarginLevelBefore float64             `json:"marginLevelBefore"`
	PositionsClosed   []LiquidatedPosition `json:"positionsClosed"`
	TotalPnL          float64             `json:"totalPnL"`
	Slippage          float64             `json:"slippage"`
	ExecutedAt        time.Time           `json:"executedAt"`
}

// LiquidatedPosition represents a position closed during liquidation
type LiquidatedPosition struct {
	PositionID    int64   `json:"positionId"`
	Symbol        string  `json:"symbol"`
	Volume        float64 `json:"volume"`
	OpenPrice     float64 `json:"openPrice"`
	ClosePrice    float64 `json:"closePrice"`
	PnL           float64 `json:"pnl"`
	Slippage      float64 `json:"slippage"`
}

// CircuitBreaker represents a trading halt condition
type CircuitBreaker struct {
	ID          string               `json:"id"`
	Type        string               `json:"type"` // VOLATILITY/LOSS_LIMIT/NEWS_EVENT/FAT_FINGER/SYSTEM
	Symbol      string               `json:"symbol,omitempty"`
	Status      CircuitBreakerStatus `json:"status"`
	Threshold   float64              `json:"threshold"`
	CurrentValue float64             `json:"currentValue"`
	TriggeredAt *time.Time           `json:"triggeredAt,omitempty"`
	ResetAt     *time.Time           `json:"resetAt,omitempty"`
	Message     string               `json:"message"`
	AutoReset   bool                 `json:"autoReset"`
	ResetAfter  time.Duration        `json:"resetAfter"` // Auto-reset duration
}

// RiskAlert represents a risk management alert
type RiskAlert struct {
	ID          string    `json:"id"`
	AccountID   int64     `json:"accountId,omitempty"`
	Symbol      string    `json:"symbol,omitempty"`
	AlertType   string    `json:"alertType"`
	Severity    RiskLevel `json:"severity"`
	Message     string    `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	Acknowledged bool     `json:"acknowledged"`
}

// PreTradeCheckResult contains the result of pre-trade validation
type PreTradeCheckResult struct {
	Allowed       bool                   `json:"allowed"`
	Reason        string                 `json:"reason,omitempty"`
	RequiredMargin float64               `json:"requiredMargin"`
	FreeMargin    float64                `json:"freeMargin"`
	MarginLevel   float64                `json:"marginLevel"`
	Warnings      []string               `json:"warnings,omitempty"`
	Checks        map[string]bool        `json:"checks"` // Individual check results
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// StressTestScenario defines a stress test scenario
type StressTestScenario struct {
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	PriceShocks     map[string]float64 `json:"priceShocks"`     // symbol -> % change
	VolatilityShock float64            `json:"volatilityShock"` // % change in volatility
	CorrelationShift float64           `json:"correlationShift"` // Change in correlation
	LiquidityShock  float64            `json:"liquidityShock"`  // Reduction in liquidity
}

// StressTestResult contains the results of a stress test
type StressTestResult struct {
	Scenario        string    `json:"scenario"`
	AccountID       int64     `json:"accountId"`
	EquityBefore    float64   `json:"equityBefore"`
	EquityAfter     float64   `json:"equityAfter"`
	PnLImpact       float64   `json:"pnlImpact"`
	MarginLevelAfter float64  `json:"marginLevelAfter"`
	WouldLiquidate  bool      `json:"wouldLiquidate"`
	MaxDrawdown     float64   `json:"maxDrawdown"`
	VaRExceeded     bool      `json:"varExceeded"`
	TestedAt        time.Time `json:"testedAt"`
}

// RiskReport contains aggregated risk metrics
type RiskReport struct {
	ReportID          string                 `json:"reportId"`
	GeneratedAt       time.Time              `json:"generatedAt"`
	ReportType        string                 `json:"reportType"` // DAILY/WEEKLY/MONTHLY/ADHOC
	AccountCount      int                    `json:"accountCount"`
	TotalEquity       float64                `json:"totalEquity"`
	TotalMargin       float64                `json:"totalMargin"`
	TotalExposure     float64                `json:"totalExposure"`
	ActivePositions   int                    `json:"activePositions"`
	MarginCallsToday  int                    `json:"marginCallsToday"`
	LiquidationsToday int                    `json:"liquidationsToday"`
	AverageMarginLevel float64               `json:"averageMarginLevel"`
	ExposureBySymbol  map[string]float64     `json:"exposureBySymbol"`
	RiskDistribution  map[RiskLevel]int      `json:"riskDistribution"`
	VaR95             float64                `json:"var95"`
	VaR99             float64                `json:"var99"`
	StressTestResults []StressTestResult     `json:"stressTestResults,omitempty"`
	TopRiskyAccounts  []RiskyAccountSummary  `json:"topRiskyAccounts"`
	Alerts            []RiskAlert            `json:"alerts"`
}

// RiskyAccountSummary contains summary of high-risk accounts
type RiskyAccountSummary struct {
	AccountID      int64     `json:"accountId"`
	MarginLevel    float64   `json:"marginLevel"`
	DailyPnL       float64   `json:"dailyPnL"`
	Exposure       float64   `json:"exposure"`
	RiskScore      float64   `json:"riskScore"` // 0-100
	Flags          []string  `json:"flags"`
}

// CorrelationMatrix stores correlation between instruments
type CorrelationMatrix struct {
	Symbols      []string    `json:"symbols"`
	Correlations [][]float64 `json:"correlations"` // NxN matrix
	UpdatedAt    time.Time   `json:"updatedAt"`
}

// OperationalLimits contains system-wide operational risk limits
type OperationalLimits struct {
	MaxOrdersPerSecond      int     `json:"maxOrdersPerSecond"`
	MaxOpenOrders           int     `json:"maxOpenOrders"`
	MaxPositionsPerAccount  int     `json:"maxPositionsPerAccount"`
	MaxSystemExposure       float64 `json:"maxSystemExposure"`
	MaintenanceMode         bool    `json:"maintenanceMode"`
	EmergencyStopEnabled    bool    `json:"emergencyStopEnabled"`
	AllowNewAccounts        bool    `json:"allowNewAccounts"`
	MaxConcurrentConnections int    `json:"maxConcurrentConnections"`
}

// CreditRiskAssessment contains credit risk evaluation
type CreditRiskAssessment struct {
	ClientID          string    `json:"clientId"`
	CreditScore       float64   `json:"creditScore"` // 0-1000
	CreditLimit       float64   `json:"creditLimit"`
	UsedCredit        float64   `json:"usedCredit"`
	AvailableCredit   float64   `json:"availableCredit"`
	PaymentHistory    []string  `json:"paymentHistory"` // Recent payment statuses
	DefaultProbability float64  `json:"defaultProbability"` // 0-1
	RiskRating        string    `json:"riskRating"` // AAA/AA/A/BBB/BB/B/CCC/CC/C/D
	LastReviewed      time.Time `json:"lastReviewed"`
	NextReviewDue     time.Time `json:"nextReviewDue"`
}

// RegulatoryLimit contains regulatory compliance limits
type RegulatoryLimit struct {
	Regulation      string  `json:"regulation"` // ESMA/NFA/FCA/ASIC/etc
	LimitType       string  `json:"limitType"`
	MaxLeverage     int     `json:"maxLeverage,omitempty"`
	MarginCloseout  float64 `json:"marginCloseout,omitempty"` // Margin % for closeout
	NegativeBalance bool    `json:"negativeBalance"` // Negative balance protection
	RequireWarnings bool    `json:"requireWarnings"` // Risk warnings required
	Jurisdiction    string  `json:"jurisdiction"`
}
