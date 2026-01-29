package affiliate

import (
	"time"
)

// AffiliateProgram represents the affiliate program configuration
type AffiliateProgram struct {
	ID                  int64     `json:"id" db:"id"`
	Name                string    `json:"name" db:"name"`
	CommissionModel     string    `json:"commissionModel" db:"commission_model"` // CPA, REVSHARE, HYBRID
	CPAAmount           float64   `json:"cpaAmount" db:"cpa_amount"`
	RevSharePercent     float64   `json:"revSharePercent" db:"revshare_percent"`
	MinPayout           float64   `json:"minPayout" db:"min_payout"`
	PayoutSchedule      string    `json:"payoutSchedule" db:"payout_schedule"` // MONTHLY, BIWEEKLY, ON_DEMAND
	CookieDuration      int       `json:"cookieDuration" db:"cookie_duration"` // Days
	SubAffiliateEnabled bool      `json:"subAffiliateEnabled" db:"sub_affiliate_enabled"`
	SubAffiliatePercent float64   `json:"subAffiliatePercent" db:"sub_affiliate_percent"`
	Status              string    `json:"status" db:"status"` // ACTIVE, PAUSED, CLOSED
	CreatedAt           time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time `json:"updatedAt" db:"updated_at"`
}

// Affiliate represents an affiliate partner
type Affiliate struct {
	ID                 int64     `json:"id" db:"id"`
	UserID             string    `json:"userId" db:"user_id"`
	AffiliateCode      string    `json:"affiliateCode" db:"affiliate_code"` // Unique code
	CompanyName        string    `json:"companyName" db:"company_name"`
	ContactName        string    `json:"contactName" db:"contact_name"`
	Email              string    `json:"email" db:"email"`
	Phone              string    `json:"phone" db:"phone"`
	Country            string    `json:"country" db:"country"`
	Website            string    `json:"website" db:"website"`
	Status             string    `json:"status" db:"status"` // PENDING, ACTIVE, SUSPENDED, BANNED
	Tier               int       `json:"tier" db:"tier"`     // Commission tier (1-5)
	ParentAffiliateID  *int64    `json:"parentAffiliateId,omitempty" db:"parent_affiliate_id"`
	CommissionModel    string    `json:"commissionModel" db:"commission_model"`
	CustomCPA          *float64  `json:"customCpa,omitempty" db:"custom_cpa"`
	CustomRevShare     *float64  `json:"customRevShare,omitempty" db:"custom_revshare"`
	PayoutMethod       string    `json:"payoutMethod" db:"payout_method"` // BANK, PAYPAL, CRYPTO, WIRE
	BankDetails        string    `json:"bankDetails,omitempty" db:"bank_details"`
	CryptoAddress      string    `json:"cryptoAddress,omitempty" db:"crypto_address"`
	TaxID              string    `json:"taxId,omitempty" db:"tax_id"`
	TotalEarnings      float64   `json:"totalEarnings" db:"total_earnings"`
	TotalPaid          float64   `json:"totalPaid" db:"total_paid"`
	PendingBalance     float64   `json:"pendingBalance" db:"pending_balance"`
	LifetimeClicks     int64     `json:"lifetimeClicks" db:"lifetime_clicks"`
	LifetimeSignups    int64     `json:"lifetimeSignups" db:"lifetime_signups"`
	LifetimeDeposits   int64     `json:"lifetimeDeposits" db:"lifetime_deposits"`
	ConversionRate     float64   `json:"conversionRate" db:"conversion_rate"`
	FraudScore         float64   `json:"fraudScore" db:"fraud_score"` // 0-100
	LastActivityAt     time.Time `json:"lastActivityAt" db:"last_activity_at"`
	CreatedAt          time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt          time.Time `json:"updatedAt" db:"updated_at"`
}

// AffiliateLink represents a tracking link
type AffiliateLink struct {
	ID            int64     `json:"id" db:"id"`
	AffiliateID   int64     `json:"affiliateId" db:"affiliate_id"`
	LinkCode      string    `json:"linkCode" db:"link_code"` // Short code
	FullURL       string    `json:"fullUrl" db:"full_url"`
	LandingPage   string    `json:"landingPage" db:"landing_page"`
	Campaign      string    `json:"campaign" db:"campaign"`
	Source        string    `json:"source" db:"source"` // facebook, google, email
	Medium        string    `json:"medium" db:"medium"` // cpc, banner, social
	Content       string    `json:"content" db:"content"`
	TotalClicks   int64     `json:"totalClicks" db:"total_clicks"`
	UniqueClicks  int64     `json:"uniqueClicks" db:"unique_clicks"`
	Conversions   int64     `json:"conversions" db:"conversions"`
	IsActive      bool      `json:"isActive" db:"is_active"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// Click represents a click tracking record
type Click struct {
	ID            int64     `json:"id" db:"id"`
	AffiliateID   int64     `json:"affiliateId" db:"affiliate_id"`
	LinkID        int64     `json:"linkId" db:"link_id"`
	ClickID       string    `json:"clickId" db:"click_id"` // UUID
	IPAddress     string    `json:"ipAddress" db:"ip_address"`
	UserAgent     string    `json:"userAgent" db:"user_agent"`
	Country       string    `json:"country" db:"country"`
	City          string    `json:"city" db:"city"`
	Device        string    `json:"device" db:"device"` // DESKTOP, MOBILE, TABLET
	Browser       string    `json:"browser" db:"browser"`
	OS            string    `json:"os" db:"os"`
	Referrer      string    `json:"referrer" db:"referrer"`
	LandingPage   string    `json:"landingPage" db:"landing_page"`
	IsUnique      bool      `json:"isUnique" db:"is_unique"`
	IsFraudulent  bool      `json:"isFraudulent" db:"is_fraudulent"`
	FraudReason   string    `json:"fraudReason,omitempty" db:"fraud_reason"`
	ConvertedAt   *time.Time `json:"convertedAt,omitempty" db:"converted_at"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// Conversion represents a successful referral conversion
type Conversion struct {
	ID              int64     `json:"id" db:"id"`
	AffiliateID     int64     `json:"affiliateId" db:"affiliate_id"`
	ClickID         string    `json:"clickId" db:"click_id"`
	UserID          string    `json:"userId" db:"user_id"`
	AccountID       int64     `json:"accountId" db:"account_id"`
	ConversionType  string    `json:"conversionType" db:"conversion_type"` // SIGNUP, DEPOSIT, FIRST_TRADE
	AttributionModel string   `json:"attributionModel" db:"attribution_model"` // FIRST_CLICK, LAST_CLICK, LINEAR
	Value           float64   `json:"value" db:"value"` // Deposit amount for DEPOSIT conversions
	Status          string    `json:"status" db:"status"` // PENDING, APPROVED, REJECTED
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	ApprovedAt      *time.Time `json:"approvedAt,omitempty" db:"approved_at"`
}

// Commission represents an earned commission
type Commission struct {
	ID              int64     `json:"id" db:"id"`
	AffiliateID     int64     `json:"affiliateId" db:"affiliate_id"`
	ConversionID    *int64    `json:"conversionId,omitempty" db:"conversion_id"`
	AccountID       int64     `json:"accountId" db:"account_id"`
	CommissionType  string    `json:"commissionType" db:"commission_type"` // CPA, REVSHARE
	Amount          float64   `json:"amount" db:"amount"`
	Currency        string    `json:"currency" db:"currency"`
	Description     string    `json:"description" db:"description"`
	Period          string    `json:"period,omitempty" db:"period"` // For RevShare: 2024-01
	TradingVolume   float64   `json:"tradingVolume,omitempty" db:"trading_volume"`
	TradingFees     float64   `json:"tradingFees,omitempty" db:"trading_fees"`
	Status          string    `json:"status" db:"status"` // PENDING, APPROVED, PAID, REVERSED
	PayoutID        *int64    `json:"payoutId,omitempty" db:"payout_id"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
	PaidAt          *time.Time `json:"paidAt,omitempty" db:"paid_at"`
}

// Payout represents a payment to an affiliate
type Payout struct {
	ID              int64     `json:"id" db:"id"`
	AffiliateID     int64     `json:"affiliateId" db:"affiliate_id"`
	Amount          float64   `json:"amount" db:"amount"`
	Currency        string    `json:"currency" db:"currency"`
	Method          string    `json:"method" db:"method"` // BANK, PAYPAL, CRYPTO, WIRE
	Status          string    `json:"status" db:"status"` // PENDING, PROCESSING, COMPLETED, FAILED
	TransactionRef  string    `json:"transactionRef,omitempty" db:"transaction_ref"`
	BankDetails     string    `json:"bankDetails,omitempty" db:"bank_details"`
	CryptoAddress   string    `json:"cryptoAddress,omitempty" db:"crypto_address"`
	Notes           string    `json:"notes,omitempty" db:"notes"`
	ProcessedBy     string    `json:"processedBy,omitempty" db:"processed_by"`
	ProcessedAt     *time.Time `json:"processedAt,omitempty" db:"processed_at"`
	CompletedAt     *time.Time `json:"completedAt,omitempty" db:"completed_at"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// ReferralCode represents a referral code for user-to-user referrals
type ReferralCode struct {
	ID              int64     `json:"id" db:"id"`
	UserID          string    `json:"userId" db:"user_id"`
	Code            string    `json:"code" db:"code"` // Unique code
	ReferrerBonus   float64   `json:"referrerBonus" db:"referrer_bonus"`
	RefereeBonus    float64   `json:"refereeBonus" db:"referee_bonus"`
	TotalUses       int64     `json:"totalUses" db:"total_uses"`
	MaxUses         int       `json:"maxUses" db:"max_uses"` // 0 = unlimited
	ExpiresAt       *time.Time `json:"expiresAt,omitempty" db:"expires_at"`
	IsActive        bool      `json:"isActive" db:"is_active"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// ReferralReward represents a reward from a referral
type ReferralReward struct {
	ID              int64     `json:"id" db:"id"`
	ReferralCodeID  int64     `json:"referralCodeId" db:"referral_code_id"`
	ReferrerUserID  string    `json:"referrerUserId" db:"referrer_user_id"`
	RefereeUserID   string    `json:"refereeUserId" db:"referee_user_id"`
	ReferrerReward  float64   `json:"referrerReward" db:"referrer_reward"`
	RefereeReward   float64   `json:"refereeReward" db:"referee_reward"`
	Status          string    `json:"status" db:"status"` // PENDING, CREDITED, EXPIRED
	CreditedAt      *time.Time `json:"creditedAt,omitempty" db:"credited_at"`
	CreatedAt       time.Time `json:"createdAt" db:"created_at"`
}

// MarketingMaterial represents downloadable marketing content
type MarketingMaterial struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // BANNER, EMAIL, VIDEO, LANDING_PAGE, SOCIAL
	Format      string    `json:"format" db:"format"` // JPG, PNG, HTML, MP4
	FileURL     string    `json:"fileUrl" db:"file_url"`
	PreviewURL  string    `json:"previewUrl" db:"preview_url"`
	Dimensions  string    `json:"dimensions,omitempty" db:"dimensions"` // 728x90, 300x250
	Language    string    `json:"language" db:"language"`
	Tags        string    `json:"tags,omitempty" db:"tags"` // Comma-separated
	Downloads   int64     `json:"downloads" db:"downloads"`
	IsActive    bool      `json:"isActive" db:"is_active"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// AffiliateStats represents real-time statistics
type AffiliateStats struct {
	AffiliateID       int64   `json:"affiliateId"`
	TodayClicks       int64   `json:"todayClicks"`
	TodaySignups      int64   `json:"todaySignups"`
	TodayDeposits     int64   `json:"todayDeposits"`
	TodayEarnings     float64 `json:"todayEarnings"`
	WeekClicks        int64   `json:"weekClicks"`
	WeekSignups       int64   `json:"weekSignups"`
	WeekDeposits      int64   `json:"weekDeposits"`
	WeekEarnings      float64 `json:"weekEarnings"`
	MonthClicks       int64   `json:"monthClicks"`
	MonthSignups      int64   `json:"monthSignups"`
	MonthDeposits     int64   `json:"monthDeposits"`
	MonthEarnings     float64 `json:"monthEarnings"`
	TotalClicks       int64   `json:"totalClicks"`
	TotalSignups      int64   `json:"totalSignups"`
	TotalDeposits     int64   `json:"totalDeposits"`
	TotalEarnings     float64 `json:"totalEarnings"`
	PendingBalance    float64 `json:"pendingBalance"`
	AvailableBalance  float64 `json:"availableBalance"`
	ConversionRate    float64 `json:"conversionRate"`
}

// FraudDetectionRule represents a fraud detection configuration
type FraudDetectionRule struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"` // IP_FRAUD, CLICK_FRAUD, DUPLICATE_ACCOUNT, VELOCITY
	Threshold   float64   `json:"threshold" db:"threshold"`
	Action      string    `json:"action" db:"action"` // FLAG, BLOCK, SUSPEND
	IsActive    bool      `json:"isActive" db:"is_active"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}
