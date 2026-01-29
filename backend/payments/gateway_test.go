package payments

import (
	"context"
	"testing"
	"time"
)

// Mock implementations for testing

type MockRepository struct {
	transactions map[string]*Transaction
	balances     map[string]float64
	depositMethods map[string][]PaymentMethod
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		transactions: make(map[string]*Transaction),
		balances:     make(map[string]float64),
		depositMethods: make(map[string][]PaymentMethod),
	}
}

func (r *MockRepository) SaveTransaction(ctx context.Context, tx *Transaction) error {
	r.transactions[tx.ID] = tx
	return nil
}

func (r *MockRepository) UpdateTransaction(ctx context.Context, tx *Transaction) error {
	r.transactions[tx.ID] = tx
	return nil
}

func (r *MockRepository) GetTransaction(ctx context.Context, txID string) (*Transaction, error) {
	if tx, ok := r.transactions[txID]; ok {
		return tx, nil
	}
	return nil, ErrTransactionNotFound
}

func (r *MockRepository) GetUserTransactions(ctx context.Context, userID string, limit, offset int) ([]*Transaction, error) {
	var result []*Transaction
	for _, tx := range r.transactions {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *MockRepository) GetTransactionsByTimeRange(ctx context.Context, from, to time.Time) ([]*Transaction, error) {
	var result []*Transaction
	for _, tx := range r.transactions {
		if tx.CreatedAt.After(from) && tx.CreatedAt.Before(to) {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *MockRepository) GetProviderTransactions(ctx context.Context, provider PaymentProvider, from, to time.Time) ([]*Transaction, error) {
	var result []*Transaction
	for _, tx := range r.transactions {
		if tx.Provider == provider && tx.CreatedAt.After(from) && tx.CreatedAt.Before(to) {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *MockRepository) GetCompletedTransactions(ctx context.Context, from, to time.Time) ([]*Transaction, error) {
	var result []*Transaction
	for _, tx := range r.transactions {
		if tx.Status == StatusCompleted && tx.CreatedAt.After(from) && tx.CreatedAt.Before(to) {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *MockRepository) CountUserTransactions(ctx context.Context, userID string, from, to time.Time) (int, error) {
	count := 0
	for _, tx := range r.transactions {
		if tx.UserID == userID && tx.CreatedAt.After(from) && tx.CreatedAt.Before(to) {
			count++
		}
	}
	return count, nil
}

func (r *MockRepository) GetUserBalance(ctx context.Context, userID string, currency string) (float64, error) {
	key := userID + ":" + currency
	if balance, ok := r.balances[key]; ok {
		return balance, nil
	}
	return 0, nil
}

func (r *MockRepository) CreditUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error {
	key := userID + ":" + currency
	r.balances[key] = r.balances[key] + amount
	return nil
}

func (r *MockRepository) DebitUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error {
	key := userID + ":" + currency
	if r.balances[key] < amount {
		return ErrInsufficientFunds
	}
	r.balances[key] = r.balances[key] - amount
	return nil
}

func (r *MockRepository) ReserveUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error {
	return r.DebitUserBalance(ctx, userID, amount, currency, txID)
}

func (r *MockRepository) UnreserveUserBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error {
	return r.CreditUserBalance(ctx, userID, amount, currency, txID)
}

func (r *MockRepository) DebitReservedBalance(ctx context.Context, userID string, amount float64, currency string, txID string) error {
	return nil
}

func (r *MockRepository) GetUserVerificationLevel(ctx context.Context, userID string) (int, error) {
	return 2, nil
}

func (r *MockRepository) GetUserCreatedAt(ctx context.Context, userID string) (time.Time, error) {
	return time.Now().Add(-30 * 24 * time.Hour), nil
}

func (r *MockRepository) GetUserLastIP(ctx context.Context, userID string) (string, error) {
	return "192.168.1.1", nil
}

func (r *MockRepository) GetUserAverageTransactionAmount(ctx context.Context, userID string) (float64, error) {
	return 100.0, nil
}

func (r *MockRepository) GetUserTotalDeposits(ctx context.Context, userID string) (float64, error) {
	return 1000.0, nil
}

func (r *MockRepository) GetUserTotalWithdrawals(ctx context.Context, userID string) (float64, error) {
	return 500.0, nil
}

func (r *MockRepository) GetUserDepositMethods(ctx context.Context, userID string) ([]PaymentMethod, error) {
	if methods, ok := r.depositMethods[userID]; ok {
		return methods, nil
	}
	return []PaymentMethod{MethodCard, MethodBankTransfer}, nil
}

func (r *MockRepository) HasPendingWithdrawal(ctx context.Context, userID string) (bool, error) {
	for _, tx := range r.transactions {
		if tx.UserID == userID && tx.Type == TypeWithdrawal && tx.Status == StatusPending {
			return true, nil
		}
	}
	return false, nil
}

func (r *MockRepository) HasDepositedWithMethod(ctx context.Context, userID string, method PaymentMethod) (bool, error) {
	methods, _ := r.GetUserDepositMethods(ctx, userID)
	for _, m := range methods {
		if m == method {
			return true, nil
		}
	}
	return false, nil
}

func (r *MockRepository) GetDeviceFailedTransactionCount(ctx context.Context, deviceID string) (int, error) {
	return 0, nil
}

func (r *MockRepository) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	return 1.0, nil
}

func (r *MockRepository) SaveExchangeRate(ctx context.Context, from, to string, rate float64) error {
	return nil
}

type MockFraudDetector struct{}

func (f *MockFraudDetector) CheckDeposit(ctx context.Context, req *PaymentRequest) (*FraudCheck, error) {
	return &FraudCheck{
		RiskScore: 10,
		RiskLevel: "low",
		Blocked:   false,
		Flags:     []string{},
		Checks:    make(map[string]string),
	}, nil
}

func (f *MockFraudDetector) CheckWithdrawal(ctx context.Context, req *PaymentRequest) (*FraudCheck, error) {
	return &FraudCheck{
		RiskScore: 10,
		RiskLevel: "low",
		Blocked:   false,
		Flags:     []string{},
		Checks:    make(map[string]string),
	}, nil
}

type MockLimitsChecker struct{}

func (l *MockLimitsChecker) CheckDepositLimits(ctx context.Context, userID string, method PaymentMethod, amount float64) error {
	if amount > 10000 {
		return ErrLimitExceeded
	}
	return nil
}

func (l *MockLimitsChecker) CheckWithdrawalLimits(ctx context.Context, userID string, method PaymentMethod, amount float64) error {
	if amount > 5000 {
		return ErrLimitExceeded
	}
	return nil
}

func (l *MockLimitsChecker) GetLimits(ctx context.Context, userID string, method PaymentMethod) (*PaymentLimits, error) {
	return &PaymentLimits{
		Method:          method,
		MinAmount:       10,
		MaxAmount:       10000,
		DailyLimit:      50000,
		WeeklyLimit:     200000,
		MonthlyLimit:    500000,
		RequiresVerification: false,
	}, nil
}

type MockWithdrawalVerifier struct{}

func (v *MockWithdrawalVerifier) VerifyWithdrawalRequest(ctx context.Context, req *PaymentRequest) error {
	return nil
}

// Tests

func TestDepositService_ProcessDeposit(t *testing.T) {
	repo := NewMockRepository()
	fraud := &MockFraudDetector{}
	limits := &MockLimitsChecker{}

	stripe := NewStripeProvider("sk_test_123", "whsec_123")
	providers := map[PaymentProvider]Provider{
		ProviderStripe: stripe,
	}

	depositService := NewDepositService(nil, providers, fraud, limits, repo)

	req := &PaymentRequest{
		UserID:    "user123",
		Type:      TypeDeposit,
		Method:    MethodCard,
		Amount:    100.0,
		Currency:  "USD",
		IPAddress: "192.168.1.1",
	}

	resp, err := depositService.ProcessDeposit(context.Background(), req)
	if err != nil {
		t.Fatalf("ProcessDeposit failed: %v", err)
	}

	if resp.TransactionID == "" {
		t.Error("Expected transaction ID")
	}

	if resp.Status != StatusProcessing {
		t.Errorf("Expected status %s, got %s", StatusProcessing, resp.Status)
	}
}

func TestWithdrawalService_ProcessWithdrawal(t *testing.T) {
	repo := NewMockRepository()
	fraud := &MockFraudDetector{}
	limits := &MockLimitsChecker{}
	verifier := &MockWithdrawalVerifier{}

	// Set user balance
	repo.balances["user123:USD"] = 1000.0

	stripe := NewStripeProvider("sk_test_123", "whsec_123")
	providers := map[PaymentProvider]Provider{
		ProviderStripe: stripe,
	}

	withdrawalService := NewWithdrawalService(nil, providers, fraud, limits, repo, verifier)

	req := &PaymentRequest{
		UserID:    "user123",
		Type:      TypeWithdrawal,
		Method:    MethodCard,
		Amount:    100.0,
		Currency:  "USD",
		IPAddress: "192.168.1.1",
	}

	resp, err := withdrawalService.ProcessWithdrawal(context.Background(), req)
	if err != nil {
		t.Fatalf("ProcessWithdrawal failed: %v", err)
	}

	if resp.TransactionID == "" {
		t.Error("Expected transaction ID")
	}

	// Check balance was reserved
	balance, _ := repo.GetUserBalance(context.Background(), "user123", "USD")
	if balance != 900.0 {
		t.Errorf("Expected balance 900.0, got %.2f", balance)
	}
}

func TestSecurityService_CheckDeposit(t *testing.T) {
	repo := NewMockRepository()
	fraudRules := DefaultFraudRules()
	security := NewSecurityService("test-encryption-key", fraudRules, nil, repo)

	req := &PaymentRequest{
		UserID:    "user123",
		Method:    MethodCard,
		Amount:    100.0,
		Currency:  "USD",
		IPAddress: "192.168.1.1",
		Country:   "US",
	}

	check, err := security.CheckDeposit(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckDeposit failed: %v", err)
	}

	if check.Blocked {
		t.Error("Expected transaction not to be blocked")
	}

	if check.RiskLevel == "critical" {
		t.Error("Expected low risk level")
	}
}

func TestSecurityService_BlockedCountry(t *testing.T) {
	repo := NewMockRepository()
	fraudRules := DefaultFraudRules()
	security := NewSecurityService("test-encryption-key", fraudRules, nil, repo)

	req := &PaymentRequest{
		UserID:    "user123",
		Method:    MethodCard,
		Amount:    100.0,
		Currency:  "USD",
		IPAddress: "192.168.1.1",
		Country:   "IR", // Iran - blocked country
	}

	check, err := security.CheckDeposit(context.Background(), req)
	if err != nil {
		t.Fatalf("CheckDeposit failed: %v", err)
	}

	if !check.Blocked {
		t.Error("Expected transaction to be blocked")
	}

	if check.Reason != "blocked country" {
		t.Errorf("Expected reason 'blocked country', got '%s'", check.Reason)
	}
}

func TestReconciliationService_ReconcileTransaction(t *testing.T) {
	repo := NewMockRepository()

	stripe := NewStripeProvider("sk_test_123", "whsec_123")
	providers := map[PaymentProvider]Provider{
		ProviderStripe: stripe,
	}

	reconciliation := NewReconciliationService(providers, repo)

	// Create a test transaction
	tx := &Transaction{
		ID:           "DEP-123",
		UserID:       "user123",
		Type:         TypeDeposit,
		Provider:     ProviderStripe,
		ProviderTxID: "pi_123",
		Status:       StatusCompleted,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	repo.SaveTransaction(context.Background(), tx)

	result, err := reconciliation.ReconcileTransaction(context.Background(), "DEP-123")
	if err != nil {
		t.Fatalf("ReconcileTransaction failed: %v", err)
	}

	if !result.Matched {
		t.Error("Expected transaction to match")
	}
}

func TestSecurityService_TokenizeCard(t *testing.T) {
	repo := NewMockRepository()
	fraudRules := DefaultFraudRules()
	security := NewSecurityService("test-encryption-key", fraudRules, nil, repo)

	// Valid card number (test Visa)
	token, err := security.TokenizeCard("4242424242424242", "123")
	if err != nil {
		t.Fatalf("TokenizeCard failed: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be generated")
	}

	// Invalid card number
	_, err = security.TokenizeCard("1234567890", "123")
	if err == nil {
		t.Error("Expected error for invalid card number")
	}
}

func TestNewTransactionID(t *testing.T) {
	depositID := NewTransactionID(TypeDeposit)
	if depositID[:4] != "DEP-" {
		t.Errorf("Expected deposit ID to start with 'DEP-', got %s", depositID)
	}

	withdrawalID := NewTransactionID(TypeWithdrawal)
	if withdrawalID[:4] != "WTH-" {
		t.Errorf("Expected withdrawal ID to start with 'WTH-', got %s", withdrawalID)
	}

	refundID := NewTransactionID(TypeRefund)
	if refundID[:4] != "REF-" {
		t.Errorf("Expected refund ID to start with 'REF-', got %s", refundID)
	}
}
