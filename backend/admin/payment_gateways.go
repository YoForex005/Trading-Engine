package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Payment Gateway Configuration API
// ============================================
// Manages payment gateway integrations and configurations
// for deposits and withdrawals.

// PaymentGateway represents a payment gateway configuration
type PaymentGateway struct {
	ID                int64                  `json:"id"`
	Name              string                 `json:"name"`              // Stripe, PayPal, Skrill, Neteller, Bank Wire
	Type              string                 `json:"type"`              // card, ewallet, bank_transfer, crypto
	Status            string                 `json:"status"`            // active, disabled, testing
	Config            PaymentGatewayConfig   `json:"config"`            // Gateway-specific configuration
	Limits            PaymentLimits          `json:"limits"`            // Transaction limits
	Fees              PaymentFees            `json:"fees"`              // Fee structure
	SupportedCurrencies []string             `json:"supportedCurrencies"` // e.g., ["USD", "EUR", "GBP"]
	DepositEnabled    bool                   `json:"depositEnabled"`
	WithdrawalEnabled bool                   `json:"withdrawalEnabled"`
	LastTestedAt      *time.Time             `json:"lastTestedAt,omitempty"`
	TestResult        string                 `json:"testResult,omitempty"` // success, failed
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         *time.Time             `json:"updatedAt,omitempty"`
}

// PaymentGatewayConfig holds gateway-specific configuration
type PaymentGatewayConfig struct {
	APIKey           string `json:"apiKey,omitempty"`           // Masked in responses
	SecretKey        string `json:"secretKey,omitempty"`        // Masked in responses
	PublishableKey   string `json:"publishableKey,omitempty"`
	MerchantID       string `json:"merchantId,omitempty"`
	WebhookURL       string `json:"webhookUrl,omitempty"`
	WebhookSecret    string `json:"webhookSecret,omitempty"`    // Masked in responses
	Environment      string `json:"environment"`                // sandbox, production
	AccountEmail     string `json:"accountEmail,omitempty"`
	BankAccountNumber string `json:"bankAccountNumber,omitempty"` // For bank wire
	BankIBAN         string `json:"bankIban,omitempty"`
	BankSWIFT        string `json:"bankSwift,omitempty"`
	BankName         string `json:"bankName,omitempty"`
	BankBranch       string `json:"bankBranch,omitempty"`
}

// PaymentLimits defines transaction limits
type PaymentLimits struct {
	MinDeposit      float64 `json:"minDeposit"`      // Minimum deposit amount
	MaxDeposit      float64 `json:"maxDeposit"`      // Maximum single deposit
	MinWithdrawal   float64 `json:"minWithdrawal"`   // Minimum withdrawal amount
	MaxWithdrawal   float64 `json:"maxWithdrawal"`   // Maximum single withdrawal
	DailyLimit      float64 `json:"dailyLimit"`      // Daily transaction limit
	MonthlyLimit    float64 `json:"monthlyLimit"`    // Monthly transaction limit
}

// PaymentFees defines fee structure
type PaymentFees struct {
	DepositFeePercent    float64 `json:"depositFeePercent"`    // % fee on deposits
	DepositFeeFixed      float64 `json:"depositFeeFixed"`      // Fixed fee per deposit
	WithdrawalFeePercent float64 `json:"withdrawalFeePercent"` // % fee on withdrawals
	WithdrawalFeeFixed   float64 `json:"withdrawalFeeFixed"`   // Fixed fee per withdrawal
	Currency             string  `json:"currency"`             // Fee currency (e.g., USD)
}

// PaymentTransaction represents a payment transaction (mock data for demo)
type PaymentTransaction struct {
	ID          string    `json:"id"`
	GatewayID   int64     `json:"gatewayId"`
	Type        string    `json:"type"`        // deposit, withdrawal
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`      // pending, completed, failed
	AccountID   int64     `json:"accountId"`
	ExternalRef string    `json:"externalRef"` // Gateway transaction reference
	CreatedAt   time.Time `json:"createdAt"`
}

// PaymentGatewayService manages payment gateway configurations
type PaymentGatewayService struct {
	mu           sync.RWMutex
	gateways     map[int64]*PaymentGateway
	transactions map[int64][]PaymentTransaction // Gateway ID -> transactions
	nextID       int64
}

// NewPaymentGatewayService creates a new payment gateway service with 5 default gateways
func NewPaymentGatewayService() *PaymentGatewayService {
	service := &PaymentGatewayService{
		gateways:     make(map[int64]*PaymentGateway),
		transactions: make(map[int64][]PaymentTransaction),
		nextID:       1,
	}

	// Create 5 default payment gateways
	defaultGateways := []struct {
		name                string
		gatewayType         string
		config              PaymentGatewayConfig
		limits              PaymentLimits
		fees                PaymentFees
		supportedCurrencies []string
		depositEnabled      bool
		withdrawalEnabled   bool
	}{
		{
			name:        "Stripe",
			gatewayType: "card",
			config: PaymentGatewayConfig{
				APIKey:         "sk_test_*********************",
				PublishableKey: "pk_test_*********************",
				WebhookURL:     "https://api.rtx5.com/webhooks/stripe",
				WebhookSecret:  "whsec_*********************",
				Environment:    "sandbox",
			},
			limits: PaymentLimits{
				MinDeposit:    10.00,
				MaxDeposit:    50000.00,
				MinWithdrawal: 20.00,
				MaxWithdrawal: 10000.00,
				DailyLimit:    100000.00,
				MonthlyLimit:  1000000.00,
			},
			fees: PaymentFees{
				DepositFeePercent:    2.9,
				DepositFeeFixed:      0.30,
				WithdrawalFeePercent: 0.0,
				WithdrawalFeeFixed:   0.0,
				Currency:             "USD",
			},
			supportedCurrencies: []string{"USD", "EUR", "GBP", "AUD", "CAD"},
			depositEnabled:      true,
			withdrawalEnabled:   false,
		},
		{
			name:        "PayPal",
			gatewayType: "ewallet",
			config: PaymentGatewayConfig{
				APIKey:       "sb-*********************",
				SecretKey:    "*********************",
				MerchantID:   "merchant_*********************",
				WebhookURL:   "https://api.rtx5.com/webhooks/paypal",
				Environment:  "sandbox",
				AccountEmail: "business@rtx5.com",
			},
			limits: PaymentLimits{
				MinDeposit:    5.00,
				MaxDeposit:    10000.00,
				MinWithdrawal: 10.00,
				MaxWithdrawal: 5000.00,
				DailyLimit:    50000.00,
				MonthlyLimit:  500000.00,
			},
			fees: PaymentFees{
				DepositFeePercent:    3.4,
				DepositFeeFixed:      0.30,
				WithdrawalFeePercent: 2.0,
				WithdrawalFeeFixed:   0.0,
				Currency:             "USD",
			},
			supportedCurrencies: []string{"USD", "EUR", "GBP", "AUD", "CAD", "JPY"},
			depositEnabled:      true,
			withdrawalEnabled:   true,
		},
		{
			name:        "Skrill",
			gatewayType: "ewallet",
			config: PaymentGatewayConfig{
				APIKey:       "api_*********************",
				SecretKey:    "secret_*********************",
				MerchantID:   "merchant_*********************",
				WebhookURL:   "https://api.rtx5.com/webhooks/skrill",
				Environment:  "production",
				AccountEmail: "payments@rtx5.com",
			},
			limits: PaymentLimits{
				MinDeposit:    10.00,
				MaxDeposit:    20000.00,
				MinWithdrawal: 20.00,
				MaxWithdrawal: 10000.00,
				DailyLimit:    75000.00,
				MonthlyLimit:  750000.00,
			},
			fees: PaymentFees{
				DepositFeePercent:    1.9,
				DepositFeeFixed:      0.0,
				WithdrawalFeePercent: 1.5,
				WithdrawalFeeFixed:   5.50,
				Currency:             "EUR",
			},
			supportedCurrencies: []string{"USD", "EUR", "GBP", "AUD"},
			depositEnabled:      true,
			withdrawalEnabled:   true,
		},
		{
			name:        "Neteller",
			gatewayType: "ewallet",
			config: PaymentGatewayConfig{
				APIKey:       "neteller_*********************",
				SecretKey:    "secret_*********************",
				MerchantID:   "merchant_*********************",
				WebhookURL:   "https://api.rtx5.com/webhooks/neteller",
				Environment:  "production",
				AccountEmail: "neteller@rtx5.com",
			},
			limits: PaymentLimits{
				MinDeposit:    10.00,
				MaxDeposit:    25000.00,
				MinWithdrawal: 20.00,
				MaxWithdrawal: 15000.00,
				DailyLimit:    100000.00,
				MonthlyLimit:  1000000.00,
			},
			fees: PaymentFees{
				DepositFeePercent:    0.0,
				DepositFeeFixed:      0.0,
				WithdrawalFeePercent: 1.5,
				WithdrawalFeeFixed:   0.0,
				Currency:             "USD",
			},
			supportedCurrencies: []string{"USD", "EUR", "GBP"},
			depositEnabled:      true,
			withdrawalEnabled:   true,
		},
		{
			name:        "Bank Wire",
			gatewayType: "bank_transfer",
			config: PaymentGatewayConfig{
				Environment:       "production",
				BankAccountNumber: "123456789",
				BankIBAN:          "GB29NWBK60161331926819",
				BankSWIFT:         "NWBKGB2L",
				BankName:          "RTX5 Trading Bank",
				BankBranch:        "London HQ",
			},
			limits: PaymentLimits{
				MinDeposit:    100.00,
				MaxDeposit:    1000000.00,
				MinWithdrawal: 100.00,
				MaxWithdrawal: 500000.00,
				DailyLimit:    5000000.00,
				MonthlyLimit:  50000000.00,
			},
			fees: PaymentFees{
				DepositFeePercent:    0.0,
				DepositFeeFixed:      0.0,
				WithdrawalFeePercent: 0.0,
				WithdrawalFeeFixed:   25.00,
				Currency:             "USD",
			},
			supportedCurrencies: []string{"USD", "EUR", "GBP"},
			depositEnabled:      true,
			withdrawalEnabled:   true,
		},
	}

	for _, dg := range defaultGateways {
		gateway := &PaymentGateway{
			ID:                  service.nextID,
			Name:                dg.name,
			Type:                dg.gatewayType,
			Status:              "active",
			Config:              dg.config,
			Limits:              dg.limits,
			Fees:                dg.fees,
			SupportedCurrencies: dg.supportedCurrencies,
			DepositEnabled:      dg.depositEnabled,
			WithdrawalEnabled:   dg.withdrawalEnabled,
			CreatedAt:           time.Now(),
		}

		service.gateways[service.nextID] = gateway

		// Generate mock transactions for demo (3-5 per gateway)
		service.generateMockTransactions(service.nextID, dg.name)

		service.nextID++
	}

	log.Printf("[PaymentGateways] Initialized with %d payment gateways (Stripe, PayPal, Skrill, Neteller, Bank Wire)", len(defaultGateways))
	return service
}

// generateMockTransactions creates mock transaction history for demo
func (pgs *PaymentGatewayService) generateMockTransactions(gatewayID int64, gatewayName string) {
	transactions := []PaymentTransaction{
		{
			ID:          fmt.Sprintf("txn_%d_001", gatewayID),
			GatewayID:   gatewayID,
			Type:        "deposit",
			Amount:      500.00,
			Currency:    "USD",
			Status:      "completed",
			AccountID:   1001,
			ExternalRef: fmt.Sprintf("%s_ext_12345", strings.ToLower(gatewayName)),
			CreatedAt:   time.Now().Add(-48 * time.Hour),
		},
		{
			ID:          fmt.Sprintf("txn_%d_002", gatewayID),
			GatewayID:   gatewayID,
			Type:        "deposit",
			Amount:      1200.00,
			Currency:    "USD",
			Status:      "completed",
			AccountID:   1002,
			ExternalRef: fmt.Sprintf("%s_ext_12346", strings.ToLower(gatewayName)),
			CreatedAt:   time.Now().Add(-24 * time.Hour),
		},
		{
			ID:          fmt.Sprintf("txn_%d_003", gatewayID),
			GatewayID:   gatewayID,
			Type:        "withdrawal",
			Amount:      300.00,
			Currency:    "USD",
			Status:      "pending",
			AccountID:   1001,
			ExternalRef: fmt.Sprintf("%s_ext_12347", strings.ToLower(gatewayName)),
			CreatedAt:   time.Now().Add(-12 * time.Hour),
		},
		{
			ID:          fmt.Sprintf("txn_%d_004", gatewayID),
			GatewayID:   gatewayID,
			Type:        "deposit",
			Amount:      2500.00,
			Currency:    "EUR",
			Status:      "completed",
			AccountID:   1003,
			ExternalRef: fmt.Sprintf("%s_ext_12348", strings.ToLower(gatewayName)),
			CreatedAt:   time.Now().Add(-6 * time.Hour),
		},
	}

	pgs.transactions[gatewayID] = transactions
}

// ListGateways returns all payment gateways
func (pgs *PaymentGatewayService) ListGateways() []*PaymentGateway {
	pgs.mu.RLock()
	defer pgs.mu.RUnlock()

	gateways := make([]*PaymentGateway, 0)
	for _, gateway := range pgs.gateways {
		// Mask sensitive data in list view
		maskedGateway := pgs.maskSensitiveData(gateway)
		gateways = append(gateways, maskedGateway)
	}
	return gateways
}

// GetGateway retrieves a specific payment gateway by ID
func (pgs *PaymentGatewayService) GetGateway(id int64) (*PaymentGateway, error) {
	pgs.mu.RLock()
	defer pgs.mu.RUnlock()

	gateway, exists := pgs.gateways[id]
	if !exists {
		return nil, fmt.Errorf("payment gateway %d not found", id)
	}

	// Mask sensitive data
	maskedGateway := pgs.maskSensitiveData(gateway)
	return maskedGateway, nil
}

// UpdateGateway updates a payment gateway configuration
func (pgs *PaymentGatewayService) UpdateGateway(id int64, updates *PaymentGateway) (*PaymentGateway, error) {
	pgs.mu.Lock()
	defer pgs.mu.Unlock()

	gateway, exists := pgs.gateways[id]
	if !exists {
		return nil, fmt.Errorf("payment gateway %d not found", id)
	}

	// Update fields
	if updates.Status != "" {
		gateway.Status = updates.Status
	}
	if updates.Config.APIKey != "" && updates.Config.APIKey != "***" {
		gateway.Config.APIKey = updates.Config.APIKey
	}
	if updates.Config.SecretKey != "" && updates.Config.SecretKey != "***" {
		gateway.Config.SecretKey = updates.Config.SecretKey
	}
	if updates.Config.WebhookURL != "" {
		gateway.Config.WebhookURL = updates.Config.WebhookURL
	}
	if updates.Config.Environment != "" {
		gateway.Config.Environment = updates.Config.Environment
	}

	// Update limits
	if updates.Limits.MinDeposit > 0 {
		gateway.Limits.MinDeposit = updates.Limits.MinDeposit
	}
	if updates.Limits.MaxDeposit > 0 {
		gateway.Limits.MaxDeposit = updates.Limits.MaxDeposit
	}
	if updates.Limits.MinWithdrawal > 0 {
		gateway.Limits.MinWithdrawal = updates.Limits.MinWithdrawal
	}
	if updates.Limits.MaxWithdrawal > 0 {
		gateway.Limits.MaxWithdrawal = updates.Limits.MaxWithdrawal
	}

	// Update fees
	if updates.Fees.DepositFeePercent >= 0 {
		gateway.Fees.DepositFeePercent = updates.Fees.DepositFeePercent
	}
	if updates.Fees.WithdrawalFeePercent >= 0 {
		gateway.Fees.WithdrawalFeePercent = updates.Fees.WithdrawalFeePercent
	}

	// Update enable flags
	gateway.DepositEnabled = updates.DepositEnabled
	gateway.WithdrawalEnabled = updates.WithdrawalEnabled

	now := time.Now()
	gateway.UpdatedAt = &now

	log.Printf("[PaymentGateways] Updated gateway: ID=%d, Name=%s", id, gateway.Name)
	return pgs.maskSensitiveData(gateway), nil
}

// TestConnection simulates a connection test to the payment gateway
func (pgs *PaymentGatewayService) TestConnection(id int64) (string, error) {
	pgs.mu.Lock()
	defer pgs.mu.Unlock()

	gateway, exists := pgs.gateways[id]
	if !exists {
		return "", fmt.Errorf("payment gateway %d not found", id)
	}

	// Simulate connection test (in production, this would make real API calls)
	now := time.Now()
	gateway.LastTestedAt = &now

	// Mock test result based on gateway status
	if gateway.Status == "active" {
		gateway.TestResult = "success"
		log.Printf("[PaymentGateways] Connection test PASSED for gateway %d (%s)", id, gateway.Name)
		return "Connection successful - gateway is operational", nil
	}

	gateway.TestResult = "failed"
	log.Printf("[PaymentGateways] Connection test FAILED for gateway %d (%s)", id, gateway.Name)
	return "Connection failed - gateway may be misconfigured or offline", fmt.Errorf("gateway is not active")
}

// GetTransactions returns recent transactions for a gateway
func (pgs *PaymentGatewayService) GetTransactions(id int64, limit int) ([]PaymentTransaction, error) {
	pgs.mu.RLock()
	defer pgs.mu.RUnlock()

	_, exists := pgs.gateways[id]
	if !exists {
		return nil, fmt.Errorf("payment gateway %d not found", id)
	}

	transactions, exists := pgs.transactions[id]
	if !exists || len(transactions) == 0 {
		return []PaymentTransaction{}, nil
	}

	// Return limited number of transactions
	if limit > 0 && limit < len(transactions) {
		return transactions[:limit], nil
	}
	return transactions, nil
}

// maskSensitiveData masks sensitive configuration data
func (pgs *PaymentGatewayService) maskSensitiveData(gateway *PaymentGateway) *PaymentGateway {
	masked := *gateway
	if gateway.Config.APIKey != "" {
		masked.Config.APIKey = "*********************"
	}
	if gateway.Config.SecretKey != "" {
		masked.Config.SecretKey = "*********************"
	}
	if gateway.Config.WebhookSecret != "" {
		masked.Config.WebhookSecret = "*********************"
	}
	return &masked
}

// ============================================
// HTTP Handlers
// ============================================

// PaymentGatewayHandler handles HTTP requests for payment gateways
type PaymentGatewayHandler struct {
	service     *PaymentGatewayService
	authService *auth.Service
}

// NewPaymentGatewayHandler creates a new payment gateway handler
func NewPaymentGatewayHandler(service *PaymentGatewayService, authService *auth.Service) *PaymentGatewayHandler {
	return &PaymentGatewayHandler{
		service:     service,
		authService: authService,
	}
}

// ListGateways handles GET /admin/payment-gateways
func (pgh *PaymentGatewayHandler) ListGateways(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	gateways := pgh.service.ListGateways()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gateways": gateways,
		"count":    len(gateways),
	})
}

// GetGateway handles GET /admin/payment-gateways/:id
func (pgh *PaymentGatewayHandler) GetGateway(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid gateway ID", http.StatusBadRequest)
		return
	}

	gateway, err := pgh.service.GetGateway(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gateway)
}

// UpdateGateway handles PUT /admin/payment-gateways/:id
func (pgh *PaymentGatewayHandler) UpdateGateway(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid gateway ID", http.StatusBadRequest)
		return
	}

	var updates PaymentGateway
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := pgh.service.UpdateGateway(id, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// TestConnection handles POST /admin/payment-gateways/:id/test
func (pgh *PaymentGatewayHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid gateway ID", http.StatusBadRequest)
		return
	}

	result, err := pgh.service.TestConnection(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // Return 200 even for failed tests
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": result,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": result,
	})
}

// GetTransactions handles GET /admin/payment-gateways/:id/transactions
func (pgh *PaymentGatewayHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Admin authentication
	token := r.Header.Get("Authorization")
	if token == "" || !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid gateway ID", http.StatusBadRequest)
		return
	}

	// Optional limit parameter
	limit := 10 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	transactions, err := pgh.service.GetTransactions(id, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gatewayId":    id,
		"transactions": transactions,
		"count":        len(transactions),
	})
}
