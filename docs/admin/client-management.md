# Client Account Management Guide

## Overview

This guide covers account creation, configuration, balance management, and client lifecycle management in the RTX Trading Engine.

## Account Types

### Demo Accounts
- **Purpose**: Practice trading without real money
- **Balance**: Virtual funds (default $5,000)
- **Reset**: Can be reset to initial balance
- **Execution**: Same as live accounts
- **Risk**: No financial risk

### Live Accounts
- **Purpose**: Real money trading
- **Balance**: Actual deposited funds
- **KYC**: Requires identity verification
- **Execution**: Real market execution
- **Risk**: Real financial risk

## Creating Client Accounts

### Via API

#### Create Demo Account
```bash
curl -X POST http://localhost:7999/api/account/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "userId": "user123",
    "username": "trader001",
    "password": "SecurePass123!",
    "isDemo": true
  }'
```

Response:
```json
{
  "id": 2,
  "accountNumber": "RTX-000002",
  "userId": "user123",
  "username": "trader001",
  "balance": 5000.00,
  "equity": 5000.00,
  "leverage": 100,
  "marginMode": "HEDGING",
  "currency": "USD",
  "status": "ACTIVE",
  "isDemo": true,
  "createdAt": 1706400000
}
```

#### Create Live Account
```bash
curl -X POST http://localhost:7999/api/account/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "userId": "user456",
    "username": "professional_trader",
    "password": "VerySecurePass456!",
    "isDemo": false
  }'
```

### Account Configuration

#### Default Settings
```json
{
  "defaultLeverage": 100,
  "defaultBalance": 5000.00,
  "marginMode": "HEDGING",
  "currency": "USD",
  "status": "ACTIVE"
}
```

#### Configure at Creation
```bash
curl -X POST http://localhost:7999/api/account/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "userId": "vip_client",
    "username": "vip001",
    "password": "SecurePass!",
    "isDemo": false,
    "leverage": 500,
    "balance": 50000.00,
    "marginMode": "NETTING"
  }'
```

## Listing Accounts

### Get All Accounts
```bash
curl http://localhost:7999/admin/accounts \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Response:
```json
[
  {
    "id": 1,
    "accountNumber": "RTX-000001",
    "username": "demo-user",
    "balance": 5000.00,
    "equity": 5025.50,
    "margin": 100.00,
    "freeMargin": 4925.50,
    "marginLevel": 5025.50,
    "openPositions": 2,
    "status": "ACTIVE",
    "isDemo": true
  },
  {
    "id": 2,
    "accountNumber": "RTX-000002",
    "username": "trader001",
    "balance": 10000.00,
    "equity": 10150.00,
    "margin": 250.00,
    "freeMargin": 9900.00,
    "marginLevel": 4060.00,
    "openPositions": 3,
    "status": "ACTIVE",
    "isDemo": false
  }
]
```

### Filter by Status
```bash
# Get only active accounts
curl 'http://localhost:7999/admin/accounts?status=ACTIVE' \
  -H "Authorization: Bearer ADMIN_TOKEN"

# Get disabled accounts
curl 'http://localhost:7999/admin/accounts?status=DISABLED' \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

## Account Configuration Updates

### Update Leverage
```bash
curl -X POST http://localhost:7999/admin/account/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "leverage": 200
  }'
```

### Update Margin Mode
```bash
curl -X POST http://localhost:7999/admin/account/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "marginMode": "NETTING"
  }'
```

**Margin Modes:**
- **HEDGING**: Allow multiple positions on same symbol (opposite directions)
- **NETTING**: Net positions on same symbol into single position

### Update Both
```bash
curl -X POST http://localhost:7999/admin/account/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "leverage": 300,
    "marginMode": "HEDGING"
  }'
```

## Balance Management

### Deposits

#### Bank Transfer
```bash
curl -X POST http://localhost:7999/admin/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "amount": 5000.00,
    "method": "BANK_TRANSFER",
    "reference": "WIRE-20240118-001"
  }'
```

#### Credit Card
```bash
curl -X POST http://localhost:7999/admin/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "amount": 1000.00,
    "method": "CREDIT_CARD",
    "reference": "CC-20240118-002"
  }'
```

#### Cryptocurrency
```bash
curl -X POST http://localhost:7999/admin/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "amount": 2500.00,
    "method": "CRYPTO",
    "reference": "BTC-0x1234567890abcdef"
  }'
```

Response:
```json
{
  "success": true,
  "accountId": 2,
  "previousBalance": 10000.00,
  "newBalance": 12500.00,
  "amount": 2500.00,
  "transactionId": 1234,
  "timestamp": 1706400000
}
```

### Withdrawals

```bash
curl -X POST http://localhost:7999/admin/withdraw \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "amount": 1000.00,
    "method": "BANK_TRANSFER",
    "reference": "WITHDRAWAL-20240118-001"
  }'
```

**Validation Rules:**
- Amount must be > 0
- Amount must be ≤ free margin (not locked in positions)
- Account must be ACTIVE
- Minimum withdrawal: $50 (configurable)

### Manual Adjustments

For corrections, bonuses, or administrative changes:

```bash
curl -X POST http://localhost:7999/admin/adjust \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "amount": -100.00,
    "reason": "Commission error correction",
    "type": "CORRECTION"
  }'
```

**Adjustment Types:**
- **CORRECTION**: Fix errors
- **ADMIN**: Administrative adjustment
- **COMPENSATION**: Client compensation
- **FEE**: Fee charge

### Bonus Credits

```bash
curl -X POST http://localhost:7999/admin/bonus \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "amount": 500.00,
    "reason": "Welcome bonus",
    "withdrawable": false
  }'
```

**Bonus Parameters:**
- **withdrawable**: true = can withdraw, false = trading only
- **expiryDate**: Optional expiration timestamp
- **conditions**: Trading volume requirements (future)

## Password Management

### Reset Password (Admin)
```bash
curl -X POST http://localhost:7999/admin/reset-password \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "newPassword": "NewSecurePass123!"
  }'
```

Response:
```json
{
  "success": true,
  "accountId": 2,
  "accountNumber": "RTX-000002",
  "message": "Password updated successfully"
}
```

### Security Requirements
- Minimum 8 characters
- Must contain uppercase, lowercase, number
- Must contain special character (recommended)
- Cannot reuse last 3 passwords (future)

## Account Status Management

### Disable Account
```bash
curl -X POST http://localhost:7999/admin/account/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "status": "DISABLED"
  }'
```

**When disabled:**
- Cannot place new orders
- Existing positions remain open
- Can still close positions
- Cannot deposit/withdraw
- Login disabled (future)

### Enable Account
```bash
curl -X POST http://localhost:7999/admin/account/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "status": "ACTIVE"
  }'
```

## Transaction History

### View Ledger
```bash
curl 'http://localhost:7999/admin/ledger?accountId=2&limit=50' \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Response:
```json
[
  {
    "id": 1001,
    "accountId": 2,
    "type": "DEPOSIT",
    "amount": 5000.00,
    "balance": 15000.00,
    "reference": "BANK-001",
    "comment": "Bank transfer deposit",
    "timestamp": 1706400000
  },
  {
    "id": 1002,
    "accountId": 2,
    "type": "COMMISSION",
    "amount": -5.00,
    "balance": 14995.00,
    "reference": "Trade-1234",
    "comment": "0.1 lot EURUSD",
    "timestamp": 1706401000
  },
  {
    "id": 1003,
    "accountId": 2,
    "type": "REALIZED_PNL",
    "amount": 150.00,
    "balance": 15145.00,
    "reference": "Position-5678",
    "comment": "GBPUSD position closed",
    "timestamp": 1706402000
  }
]
```

### Ledger Entry Types
- **DEPOSIT**: Funds added
- **WITHDRAWAL**: Funds removed
- **REALIZED_PNL**: Profit/loss from closed positions
- **COMMISSION**: Trading commissions
- **SWAP**: Overnight interest
- **BONUS**: Promotional credits
- **ADJUSTMENT**: Manual adjustments
- **CORRECTION**: Error corrections

## Account Monitoring

### Get Account Summary
```bash
curl 'http://localhost:7999/api/account/summary?accountId=2' \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Response:
```json
{
  "accountId": 2,
  "accountNumber": "RTX-000002",
  "currency": "USD",
  "balance": 15145.00,
  "equity": 15220.50,
  "margin": 500.00,
  "freeMargin": 14720.50,
  "marginLevel": 3044.10,
  "unrealizedPnL": 75.50,
  "leverage": 100,
  "marginMode": "HEDGING",
  "openPositions": 3
}
```

### Monitor Margin Levels

**Margin Level Calculation:**
```
Margin Level = (Equity / Used Margin) × 100
```

**Risk Levels:**
- **> 100%**: Healthy
- **50-100%**: Warning
- **< 50%**: Margin Call
- **< 20%**: Stop Out (auto-close positions)

### Set Alerts (Future)
```bash
curl -X POST http://localhost:7999/admin/alerts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "type": "MARGIN_LEVEL",
    "threshold": 100,
    "action": "EMAIL",
    "recipient": "admin@broker.com"
  }'
```

## Client Segmentation

### Risk-Based Grouping

**Low Risk (B-Book Candidates)**
- Win rate < 40%
- Average trade < 1 lot
- Demo accounts
- New traders (< 30 days)

**High Risk (A-Book Candidates)**
- Win rate > 60%
- Average trade > 10 lots
- Consistent profitability
- Professional traders

### Tag Accounts (Future)
```bash
curl -X POST http://localhost:7999/admin/account/tags \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{
    "accountId": 2,
    "tags": ["VIP", "SCALPER", "ABOOK"]
  }'
```

## Compliance and Reporting

### KYC Status
```json
{
  "accountId": 2,
  "kycStatus": "VERIFIED",
  "documents": [
    {
      "type": "ID_CARD",
      "status": "APPROVED",
      "uploadedAt": 1706300000
    },
    {
      "type": "PROOF_OF_ADDRESS",
      "status": "APPROVED",
      "uploadedAt": 1706300500
    }
  ]
}
```

### Account Activity Report
```bash
curl 'http://localhost:7999/admin/reports/activity?accountId=2&from=2024-01-01&to=2024-01-31' \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

Response:
```json
{
  "accountId": 2,
  "period": {
    "from": "2024-01-01",
    "to": "2024-01-31"
  },
  "statistics": {
    "totalTrades": 150,
    "winningTrades": 85,
    "losingTrades": 65,
    "winRate": 56.67,
    "totalVolume": 45.5,
    "totalPnL": 2500.50,
    "totalCommission": 227.50,
    "netPnL": 2273.00,
    "largestWin": 450.00,
    "largestLoss": -320.00,
    "averageTrade": 15.15
  }
}
```

## Best Practices

### Account Creation
1. Verify client identity (KYC) before creating live account
2. Start with demo account for new traders
3. Use strong password requirements
4. Document account creation reason
5. Set appropriate leverage based on experience

### Balance Management
1. Verify deposit source (AML compliance)
2. Process withdrawals within 24 hours
3. Maintain audit trail for all transactions
4. Use 4-eyes principle for large adjustments
5. Document all manual adjustments

### Risk Management
1. Monitor margin levels daily
2. Set up automated margin call notifications
3. Review high-risk accounts weekly
4. Enforce position size limits
5. Disable accounts on suspicious activity

### Client Communication
1. Notify clients of deposit/withdrawal confirmations
2. Send margin call warnings before stop-out
3. Provide monthly account statements
4. Respond to inquiries within 24 hours
5. Document all client interactions

## Troubleshooting

### Account Not Found
```json
{
  "error": "Account not found",
  "accountId": 999
}
```
**Solution**: Verify account ID, check if account was deleted

### Insufficient Balance
```json
{
  "error": "Insufficient balance for withdrawal",
  "requested": 5000.00,
  "available": 3500.00
}
```
**Solution**: Check free margin, close positions if needed

### Disabled Account
```json
{
  "error": "Account is disabled",
  "accountId": 2
}
```
**Solution**: Re-enable account via admin/account/update

## See Also

- [Risk Parameters Configuration](risk-parameters.md)
- [Ledger and Reporting](reporting.md)
- [API Reference](../api/endpoints.md)
