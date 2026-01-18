package admin

import "time"

// AdminRole defines admin privilege levels
type AdminRole string

const (
	RoleSuperAdmin AdminRole = "SUPER_ADMIN" // Full access to everything
	RoleAdmin      AdminRole = "ADMIN"       // Manage users, funds, orders
	RoleSupport    AdminRole = "SUPPORT"     // View-only, limited modifications
)

// Admin represents an admin user
type Admin struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"` // Never expose in JSON
	Role            AdminRole `json:"role"`
	IPWhitelist     []string  `json:"ipWhitelist"` // Allowed IPs
	TwoFactorSecret string    `json:"-"`           // TOTP secret
	TwoFactorEnabled bool      `json:"twoFactorEnabled"`
	Status          string    `json:"status"` // ACTIVE, DISABLED, SUSPENDED
	LastLogin       time.Time `json:"lastLogin"`
	CreatedAt       time.Time `json:"createdAt"`
	CreatedBy       string    `json:"createdBy"`
}

// AdminSession tracks admin login sessions
type AdminSession struct {
	SessionID  string    `json:"sessionId"`
	AdminID    int64     `json:"adminId"`
	Username   string    `json:"username"`
	Role       AdminRole `json:"role"`
	IPAddress  string    `json:"ipAddress"`
	UserAgent  string    `json:"userAgent"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastActive time.Time `json:"lastActive"`
}

// UserGroup defines a trading group with custom settings
type UserGroup struct {
	ID              int64             `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	ExecutionMode   string            `json:"executionMode"` // BBOOK, ABOOK, HYBRID
	Markup          float64           `json:"markup"`        // Spread markup in pips
	Commission      float64           `json:"commission"`    // Commission per lot
	MaxLeverage     float64           `json:"maxLeverage"`
	EnabledSymbols  []string          `json:"enabledSymbols"`
	SymbolSettings  map[string]SymbolGroupSettings `json:"symbolSettings"`
	DefaultBalance  float64           `json:"defaultBalance"`
	MarginMode      string            `json:"marginMode"` // HEDGING, NETTING
	Status          string            `json:"status"`     // ACTIVE, DISABLED
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	CreatedBy       string            `json:"createdBy"`
}

// SymbolGroupSettings contains per-symbol group settings
type SymbolGroupSettings struct {
	Symbol         string  `json:"symbol"`
	Markup         float64 `json:"markup"`     // Override group markup
	Commission     float64 `json:"commission"` // Override group commission
	MaxVolume      float64 `json:"maxVolume"`
	MinVolume      float64 `json:"minVolume"`
	Disabled       bool    `json:"disabled"`
}

// AuditEntry represents an audit trail record
type AuditEntry struct {
	ID          int64       `json:"id"`
	AdminID     int64       `json:"adminId"`
	AdminName   string      `json:"adminName"`
	Action      string      `json:"action"` // USER_UPDATE, FUND_DEPOSIT, ORDER_MODIFY, etc.
	EntityType  string      `json:"entityType"` // USER, ORDER, POSITION, GROUP
	EntityID    int64       `json:"entityId"`
	Changes     interface{} `json:"changes"` // JSON object of changes
	Reason      string      `json:"reason"`
	IPAddress   string      `json:"ipAddress"`
	UserAgent   string      `json:"userAgent"`
	Status      string      `json:"status"` // SUCCESS, FAILED
	ErrorMsg    string      `json:"errorMsg,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// FundOperation represents a fund management operation
type FundOperation struct {
	ID          int64     `json:"id"`
	AccountID   int64     `json:"accountId"`
	Type        string    `json:"type"` // DEPOSIT, WITHDRAW, ADJUSTMENT, BONUS
	Amount      float64   `json:"amount"`
	Method      string    `json:"method"` // BANK, CRYPTO, CARD, MANUAL
	Reference   string    `json:"reference"`
	Description string    `json:"description"`
	Reason      string    `json:"reason"`
	AdminID     int64     `json:"adminId"`
	AdminName   string    `json:"adminName"`
	Status      string    `json:"status"` // PENDING, COMPLETED, REJECTED
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// OrderModification represents an admin order modification
type OrderModification struct {
	ID          int64       `json:"id"`
	OrderID     int64       `json:"orderId"`
	PositionID  int64       `json:"positionId,omitempty"`
	Action      string      `json:"action"` // MODIFY, REVERSE, DELETE, CLOSE
	Changes     interface{} `json:"changes"` // JSON of modifications
	Reason      string      `json:"reason"`
	AdminID     int64       `json:"adminId"`
	AdminName   string      `json:"adminName"`
	CreatedAt   time.Time   `json:"createdAt"`
}

// UserAccountInfo extends Account with group info
type UserAccountInfo struct {
	ID            int64     `json:"id"`
	AccountNumber string    `json:"accountNumber"`
	UserID        string    `json:"userId"`
	Username      string    `json:"username"`
	Email         string    `json:"email,omitempty"`
	Balance       float64   `json:"balance"`
	Equity        float64   `json:"equity"`
	Margin        float64   `json:"margin"`
	FreeMargin    float64   `json:"freeMargin"`
	Leverage      float64   `json:"leverage"`
	GroupID       int64     `json:"groupId,omitempty"`
	GroupName     string    `json:"groupName,omitempty"`
	Status        string    `json:"status"`
	IsDemo        bool      `json:"isDemo"`
	CreatedAt     time.Time `json:"createdAt"`
	LastLogin     *time.Time `json:"lastLogin,omitempty"`
	OpenPositions int       `json:"openPositions"`
	TotalVolume   float64   `json:"totalVolume"`
	TotalPnL      float64   `json:"totalPnL"`
}

// AdminStats provides admin dashboard statistics
type AdminStats struct {
	TotalUsers       int     `json:"totalUsers"`
	ActiveUsers      int     `json:"activeUsers"`
	DemoAccounts     int     `json:"demoAccounts"`
	LiveAccounts     int     `json:"liveAccounts"`
	TotalBalance     float64 `json:"totalBalance"`
	TotalEquity      float64 `json:"totalEquity"`
	TotalMargin      float64 `json:"totalMargin"`
	OpenPositions    int     `json:"openPositions"`
	TotalVolume      float64 `json:"totalVolume"`
	TodayDeposits    float64 `json:"todayDeposits"`
	TodayWithdrawals float64 `json:"todayWithdrawals"`
	TodayPnL         float64 `json:"todayPnL"`
}
