package models

import "time"

// Jurisdiction represents regulatory jurisdiction
type Jurisdiction string

const (
	JurisdictionEU        Jurisdiction = "EU"         // MiFID II
	JurisdictionUK        Jurisdiction = "UK"         // FCA
	JurisdictionUS        Jurisdiction = "US"         // CFTC/NFA
	JurisdictionAustralia Jurisdiction = "AUSTRALIA"  // ASIC
	JurisdictionCyprus    Jurisdiction = "CYPRUS"     // CySEC
	JurisdictionSA        Jurisdiction = "SOUTH_AFRICA" // FSCA
)

// RegulatoryFramework defines compliance requirements per jurisdiction
type RegulatoryFramework struct {
	Jurisdiction            Jurisdiction
	TransactionReporting    bool
	BestExecutionReporting  bool
	PositionReporting       bool
	NegativeBalanceProtection bool
	LeverageLimits          map[string]int // instrument -> max leverage
	MinimumMarginLevel      float64
	ClientMoneyProtection   bool
	AuditTrailRetention     int // years
	RiskWarningsRequired    bool
}

// ClientClassification represents client categorization
type ClientClassification string

const (
	ClientRetail              ClientClassification = "RETAIL"
	ClientProfessional        ClientClassification = "PROFESSIONAL"
	ClientEligibleCounterparty ClientClassification = "ELIGIBLE_COUNTERPARTY"
)

// TransactionReport represents regulatory transaction report
type TransactionReport struct {
	ID                  string
	ReportType          string // MiFID_II, EMIR, CAT
	Jurisdiction        Jurisdiction
	TransactionID       string
	OrderID             string
	ExecutionID         string
	ClientID            string
	Symbol              string
	ISIN                string
	Side                string
	Quantity            float64
	Price               float64
	ExecutionTimestamp  time.Time
	TradingVenue        string
	LiquidityProvider   string
	Currency            string
	ClientClassification ClientClassification
	InvestmentDecisionMaker string
	ExecutingTrader     string
	// MiFID II specific (27 fields)
	TransmissionOfOrder bool
	BuyerIdentification string
	SellerIdentification string
	ShortSellingIndicator bool
	WaiverIndicator     string
	CreatedAt           time.Time
	SubmittedAt         *time.Time
	Status              string // PENDING, SUBMITTED, FAILED
}

// BestExecutionReport represents RTS 27/28 report data
type BestExecutionReport struct {
	ID              string
	ReportType      string // RTS_27, RTS_28
	Period          string // Q1_2026, Q2_2026, etc
	InstrumentClass string
	Venue           string
	LPName          string
	// RTS 27 metrics
	PriceImprovement      float64
	FillRate              float64
	AverageExecutionTime  float64 // ms
	SlippageRate          float64
	RejectionRate         float64
	// RTS 28 top venues
	ExecutionVolume       float64
	NumberOfOrders        int
	PassiveOrders         int
	AggressiveOrders      int
	DirectedOrders        int
	CreatedAt             time.Time
	PublishedAt           *time.Time
}

// KYCRecord represents Know Your Customer data
type KYCRecord struct {
	ID                  string
	ClientID            string
	FullName            string
	DateOfBirth         time.Time
	Nationality         string
	ResidenceCountry    string
	Address             string
	DocumentType        string // PASSPORT, DRIVERS_LICENSE, ID_CARD
	DocumentNumber      string
	DocumentExpiry      time.Time
	DocumentVerified    bool
	VerificationProvider string // Onfido, Jumio, Trulioo
	ProofOfAddress      string
	AddressVerified     bool
	PEPStatus           string // NOT_PEP, PEP, CLOSE_ASSOCIATE
	SanctionsMatch      bool
	SanctionsLists      []string // OFAC, UN, EU
	RiskRating          string   // LOW, MEDIUM, HIGH
	OngoingMonitoring   bool
	LastScreening       time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// AMLAlert represents suspicious activity detection
type AMLAlert struct {
	ID                string
	ClientID          string
	AlertType         string // LARGE_DEPOSIT, RAPID_TURNOVER, UNUSUAL_PATTERN
	Description       string
	Severity          string // LOW, MEDIUM, HIGH, CRITICAL
	TransactionIDs    []string
	DetectedAt        time.Time
	Status            string // PENDING, INVESTIGATING, RESOLVED, ESCALATED
	AssignedTo        string
	SARFiled          bool // Suspicious Activity Report
	SARFiledAt        *time.Time
	Resolution        string
	ResolvedAt        *time.Time
}

// PositionReport represents regulatory position reporting
type PositionReport struct {
	ID              string
	ReportType      string // EMIR, CFTC_LARGE_TRADER
	ClientID        string
	Symbol          string
	ISIN            string
	PositionType    string // LONG, SHORT
	Quantity        float64
	NotionalValue   float64
	Currency        string
	Jurisdiction    Jurisdiction
	ReportingDate   time.Time
	SubmittedAt     *time.Time
	Status          string
}

// AuditTrailEntry represents immutable audit log
type AuditTrailEntry struct {
	ID            string
	EventType     string // ORDER_PLACED, ORDER_MODIFIED, TRADE_EXECUTED, etc
	UserID        string
	UserRole      string
	ClientID      string
	OrderID       string
	TradeID       string
	Symbol        string
	Action        string
	Before        string // JSON state before
	After         string // JSON state after
	IPAddress     string
	UserAgent     string
	Timestamp     time.Time
	Hash          string // SHA256 hash for tamper detection
	PreviousHash  string // Chain to previous entry
}

// LeverageLimit represents jurisdiction-specific leverage caps
type LeverageLimit struct {
	ID               string
	Jurisdiction     Jurisdiction
	InstrumentClass  string // MAJOR_PAIRS, MINOR_PAIRS, GOLD, INDICES, etc
	ClientClass      ClientClassification
	MaxLeverage      int
	WarningThreshold int
	EffectiveFrom    time.Time
}

// RiskWarning represents mandatory risk disclosure
type RiskWarning struct {
	ID              string
	Jurisdiction    Jurisdiction
	ProductType     string // CFD, FOREX, CRYPTO
	WarningType     string // GENERAL, LEVERAGE, PERCENTAGE_LOSS
	Language        string
	Content         string
	DisplayTrigger  string // ACCOUNT_OPENING, PRE_TRADE, STATEMENT
	MandatoryDisplay bool
	EffectiveFrom   time.Time
}

// ClientStatement represents periodic client reporting
type ClientStatement struct {
	ID                    string
	ClientID              string
	StatementType         string // DAILY, MONTHLY, ANNUAL
	Period                string
	OpeningBalance        float64
	Deposits              float64
	Withdrawals           float64
	RealizedPnL           float64
	UnrealizedPnL         float64
	ClosingBalance        float64
	Fees                  float64
	TotalTrades           int
	WinningTrades         int
	LosingTrades          int
	PercentageLoss        float64 // Required disclosure
	TaxInformation        string
	GeneratedAt           time.Time
	SentAt                *time.Time
}

// Complaint represents client complaint tracking
type Complaint struct {
	ID              string
	ClientID        string
	ComplaintType   string // EXECUTION, PRICING, WITHDRAWAL, PLATFORM
	Severity        string
	Description     string
	SubmittedAt     time.Time
	Status          string // NEW, INVESTIGATING, RESOLVED, ESCALATED
	AssignedTo      string
	EscalationLevel int
	ResolutionDue   time.Time
	ResolvedAt      *time.Time
	Resolution      string
	ClientSatisfied bool
	RegulatoryReport bool // Reported to regulator
	ReportedAt      *time.Time
}

// SegregatedAccount represents client fund segregation
type SegregatedAccount struct {
	ID                string
	ClientID          string
	TrusteeBank       string
	AccountNumber     string
	Currency          string
	Balance           float64
	CompanyFunds      float64 // Should always be 0
	LastReconciliation time.Time
	ReconciliationStatus string
	Discrepancies     []string
}

// GDPRConsent represents data protection consent
type GDPRConsent struct {
	ID              string
	ClientID        string
	ConsentType     string // MARKETING, PROFILING, DATA_SHARING
	ConsentGiven    bool
	ConsentDate     time.Time
	WithdrawnAt     *time.Time
	Purpose         string
	LegalBasis      string
	ExpiryDate      *time.Time
}

// DataPortabilityRequest represents GDPR data export
type DataPortabilityRequest struct {
	ID              string
	ClientID        string
	RequestDate     time.Time
	Status          string // PENDING, PROCESSING, COMPLETED
	DataPackage     string // JSON or XML export
	ExpiryDate      time.Time
	CompletedAt     *time.Time
}

// RightToErasureRequest represents GDPR deletion request
type RightToErasureRequest struct {
	ID              string
	ClientID        string
	RequestDate     time.Time
	Reason          string
	Status          string
	RejectionReason string // Legal obligation to retain
	ProcessedAt     *time.Time
	DataDeleted     bool
}
