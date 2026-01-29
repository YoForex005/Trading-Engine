# Payment Gateway Integration

Comprehensive payment gateway system supporting deposits, withdrawals, multiple payment providers, and PCI DSS compliant security.

## Features

### Payment Methods
- **Credit/Debit Cards** - Stripe, Braintree
- **Bank Transfers** - ACH, SEPA, Wire (Wise/TransferWise)
- **E-wallets** - PayPal, Skrill, Neteller
- **Cryptocurrencies** - Bitcoin, Ethereum, USDT (Coinbase, BitPay, Circle)
- **Local Payment Methods** - Region-specific payment options

### Deposit Flow
- Instant card deposits with 3D Secure
- Bank transfer tracking (1-3 day processing)
- Crypto deposit monitoring with confirmations
- Automatic account crediting
- Deposit limits per method and verification level
- Multi-layer fraud detection
- PCI DSS Level 1 compliance

### Withdrawal Flow
- 2FA and email verification
- Same-method withdrawal rule (AML compliance)
- Multi-tier withdrawal limits (daily, weekly, monthly)
- Manual approval for large amounts (>$10,000 or high risk)
- Automatic small withdrawals
- Fast crypto withdrawals (<30 min)
- Bank transfer processing
- Dynamic fee calculation

### Security Features
- **PCI DSS Level 1 Compliance**
  - Card tokenization (no card storage)
  - AES-256 encryption at rest
  - TLS 1.3 encryption in transit

- **Fraud Detection**
  - Velocity checks (transactions per hour/day)
  - Geo-blocking (country-based restrictions)
  - IP reputation checking
  - Device fingerprinting
  - Amount anomaly detection
  - Risk scoring (0-100)

- **Anti-Money Laundering (AML)**
  - Same-method withdrawal rule
  - Transaction pattern analysis
  - Suspicious activity flagging

### Payment Reconciliation
- Automatic payment matching
- Provider status synchronization
- Settlement reporting
- Chargeback handling
- Dispute management
- Failed payment retry logic (max 3 attempts)
- Refund processing

## Architecture

```
payments/
├── gateway.go           # Core abstractions and interfaces
├── deposits.go          # Deposit processing service
├── withdrawals.go       # Withdrawal processing service
├── reconciliation.go    # Payment reconciliation service
├── providers.go         # Payment provider implementations
├── security.go          # PCI DSS security and fraud detection
├── repository.go        # Data persistence interface
└── gateway_test.go      # Comprehensive tests
```

## Provider Integrations

### Stripe
- Card payments with 3D Secure
- ACH bank transfers
- Instant deposits
- Webhook support

### Coinbase Commerce
- Bitcoin deposits/withdrawals
- Ethereum deposits/withdrawals
- USDT stablecoin support
- Confirmation monitoring

### PayPal
- E-wallet deposits/withdrawals
- OAuth redirect flow
- Instant processing
- Webhook notifications

### Wise (TransferWise)
- International bank transfers
- SEPA transfers
- Wire transfers
- Multi-currency support

### Circle
- USDC stablecoin
- Card payments
- Bank transfers
- Instant settlements

## Usage Examples

### Initialize Services

```go
import "github.com/epic1st/rtx/backend/payments"

// Create providers
stripe := payments.NewStripeProvider("sk_live_...", "whsec_...")
coinbase := payments.NewCoinbaseProvider("api_key", "secret")
paypal := payments.NewPayPalProvider("client_id", "secret", false)

providers := map[payments.PaymentProvider]payments.Provider{
    payments.ProviderStripe:   stripe,
    payments.ProviderCoinbase: coinbase,
    payments.ProviderPayPal:   paypal,
}

// Create security service
fraudRules := payments.DefaultFraudRules()
security := payments.NewSecurityService(
    "encryption-key-32-bytes-long",
    fraudRules,
    ipReputationChecker,
    repository,
)

// Create limits checker
limitsChecker := NewLimitsChecker(repository)

// Create deposit service
depositService := payments.NewDepositService(
    nil,
    providers,
    security,
    limitsChecker,
    repository,
)

// Create withdrawal service
withdrawalService := payments.NewWithdrawalService(
    nil,
    providers,
    security,
    limitsChecker,
    repository,
    withdrawalVerifier,
)
```

### Process a Card Deposit

```go
req := &payments.PaymentRequest{
    UserID:   "user123",
    Type:     payments.TypeDeposit,
    Method:   payments.MethodCard,
    Amount:   100.0,
    Currency: "USD",
    PaymentDetails: map[string]string{
        "card_token": "tok_visa",
    },
    IPAddress: "192.168.1.1",
    DeviceID:  "device123",
}

resp, err := depositService.ProcessDeposit(ctx, req)
if err != nil {
    log.Printf("Deposit failed: %v", err)
    return
}

if resp.RequiresAction {
    // Redirect user to 3D Secure
    fmt.Printf("Redirect to: %s\n", resp.RedirectURL)
} else {
    fmt.Printf("Deposit successful: %s\n", resp.TransactionID)
}
```

### Process a Crypto Deposit

```go
req := &payments.PaymentRequest{
    UserID:   "user123",
    Type:     payments.TypeDeposit,
    Method:   payments.MethodBitcoin,
    Amount:   0.01,
    Currency: "BTC",
    IPAddress: "192.168.1.1",
}

resp, err := depositService.ProcessCryptoDeposit(ctx, req)
if err != nil {
    log.Printf("Crypto deposit failed: %v", err)
    return
}

// Get deposit address from response
depositAddress := resp.ActionData["deposit_address"]
fmt.Printf("Send BTC to: %s\n", depositAddress)
fmt.Printf("Required confirmations: 3\n")

// Monitor confirmations
go func() {
    for {
        time.Sleep(1 * time.Minute)
        err := depositService.MonitorCryptoDeposit(ctx, resp.TransactionID)
        if err != nil {
            log.Printf("Monitor error: %v", err)
        }
    }
}()
```

### Process a Withdrawal

```go
req := &payments.PaymentRequest{
    UserID:   "user123",
    Type:     payments.TypeWithdrawal,
    Method:   payments.MethodBankTransfer,
    Amount:   500.0,
    Currency: "USD",
    PaymentDetails: map[string]string{
        "account_number": "12345678",
        "routing_number": "021000021",
    },
    IPAddress: "192.168.1.1",
}

resp, err := withdrawalService.ProcessWithdrawal(ctx, req)
if err != nil {
    log.Printf("Withdrawal failed: %v", err)
    return
}

if resp.Status == payments.StatusPending {
    fmt.Printf("Withdrawal pending approval: %s\n", resp.Message)
} else {
    fmt.Printf("Withdrawal processing: %s\n", resp.TransactionID)
    fmt.Printf("Estimated time: %s\n", resp.EstimatedTime)
}
```

### Manual Withdrawal Approval

```go
// Approve withdrawal
err := withdrawalService.ApproveWithdrawal(ctx, "WTH-123", "admin-user-id")
if err != nil {
    log.Printf("Approval failed: %v", err)
    return
}

// Or reject withdrawal
err = withdrawalService.RejectWithdrawal(ctx, "WTH-123", "admin-user-id", "Suspicious activity")
```

### Handle Provider Webhook

```go
func handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
    payload, _ := ioutil.ReadAll(r.Body)
    signature := r.Header.Get("Stripe-Signature")

    provider := providers[payments.ProviderStripe]

    // Verify signature
    err := provider.VerifyWebhookSignature(ctx, payload, []byte(signature))
    if err != nil {
        http.Error(w, "Invalid signature", 401)
        return
    }

    // Parse webhook
    event, err := provider.ParseWebhook(ctx, payload)
    if err != nil {
        http.Error(w, "Parse error", 400)
        return
    }

    // Update transaction status
    switch event.EventType {
    case "payment_intent.succeeded":
        // Update deposit to completed
    case "payout.paid":
        // Update withdrawal to completed
    }

    w.WriteHeader(200)
}
```

### Run Reconciliation

```go
// Reconcile all transactions for the day
from := time.Now().Add(-24 * time.Hour)
to := time.Now()

results, err := reconciliationService.ReconcileTransactions(ctx, from, to)
if err != nil {
    log.Printf("Reconciliation failed: %v", err)
    return
}

// Check for discrepancies
for _, result := range results {
    if !result.Matched {
        log.Printf("Discrepancy found: %s - %s\n",
            result.TransactionID, result.Discrepancy)
    }
}
```

### Generate Settlement Report

```go
from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
to := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

report, err := reconciliationService.GenerateSettlementReport(ctx, from, to)
if err != nil {
    log.Printf("Report generation failed: %v", err)
    return
}

fmt.Printf("Settlement Report: %s to %s\n", from.Format("2006-01-02"), to.Format("2006-01-02"))
fmt.Printf("Total Deposits: $%.2f\n", report.TotalDeposits)
fmt.Printf("Total Withdrawals: $%.2f\n", report.TotalWithdrawals)
fmt.Printf("Total Fees: $%.2f\n", report.TotalFees)
fmt.Printf("Net Settlement: $%.2f\n", report.NetSettlement)
```

## Payment Limits

Limits are based on user verification level:

### Level 0 (Unverified)
- Card: $10-$500 (max $1,000/day)
- PayPal: $10-$300 (max $500/day)

### Level 1 (Email Verified)
- Card: $10-$2,000 (max $5,000/day)
- Bank Transfer: $50-$5,000 (max $10,000/day)
- PayPal: $10-$1,000 (max $2,000/day)

### Level 2 (KYC Verified)
- Card: $10-$10,000 (max $50,000/day)
- Bank Transfer: $50-$50,000 (max $100,000/day)
- Wire: $100-$100,000 (max $500,000/day)
- Crypto: $100-$50,000 (max $100,000/day)

## Fraud Detection

Risk scoring (0-100):
- **0-30**: Low risk (auto-approve)
- **30-60**: Medium risk (auto-approve with monitoring)
- **60-80**: High risk (require manual review)
- **80-100**: Critical risk (block transaction)

Checks performed:
1. Velocity check (transactions per hour/day)
2. IP reputation (Tor, proxy, VPN detection)
3. Geo-location (country blocking)
4. Amount anomaly (unusual transaction size)
5. Device fingerprinting
6. Account age
7. Withdrawal patterns

## Testing

Run all payment tests:

```bash
cd /Users/epic1st/Documents/trading\ engine/backend/payments
go test -v
```

Run specific tests:

```bash
go test -v -run TestDepositService_ProcessDeposit
go test -v -run TestSecurityService_CheckDeposit
go test -v -run TestReconciliationService
```

## Database Migration

Apply payment tables migration:

```bash
mysql -u root -p trading_engine < ../migrations/006_add_payment_tables.sql
```

## Security Considerations

1. **Never store raw card numbers** - Always tokenize
2. **Encrypt sensitive data** - Use AES-256-GCM
3. **Verify webhook signatures** - Prevent replay attacks
4. **Log all transactions** - For audit and compliance
5. **Rate limit endpoints** - Prevent brute force
6. **Use HTTPS only** - TLS 1.3 minimum
7. **Implement IP whitelisting** - For admin operations
8. **Regular security audits** - PCI DSS quarterly scans

## Compliance

- **PCI DSS Level 1** - Quarterly compliance validation
- **AML/KYC** - Know Your Customer verification
- **GDPR** - User data protection and privacy
- **SOC 2 Type II** - Security and availability controls

## Monitoring

Key metrics to monitor:
- Transaction success rate
- Average processing time
- Fraud detection accuracy
- Chargeback rate
- Settlement reconciliation rate
- Provider uptime

## Support

For issues or questions:
- Technical: backend-team@example.com
- Compliance: compliance@example.com
- Security: security@example.com
