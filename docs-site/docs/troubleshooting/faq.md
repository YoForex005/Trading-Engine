---
id: faq
title: Frequently Asked Questions
sidebar_label: FAQ
sidebar_position: 10
description: Common questions and answers about Trading Platform
---

# Frequently Asked Questions

## Account & Registration

### How do I create an account?

1. Visit [yourtradingplatform.com/register](https://yourtradingplatform.com/register)
2. Fill in your details (email, password, country)
3. Verify your email address
4. Complete KYC verification

See our [Quick Start Guide](/docs/getting-started/quick-start) for detailed instructions.

### What documents do I need for verification?

Required documents:
- Government-issued ID (passport, driver's license, national ID)
- Proof of address (utility bill, bank statement) dated within last 3 months

### How long does verification take?

- Typical verification time: 1-2 hours during business hours
- After hours: Up to 24 hours
- Complex cases: 2-3 business days

### Can I have multiple accounts?

No. Each person is allowed one live account per email address. Multiple demo accounts are allowed.

## Deposits & Withdrawals

### What is the minimum deposit?

Minimum deposit varies by payment method:
- Bank transfer: $100 USD
- Credit/debit card: $100 USD
- Cryptocurrency: $100 USD equivalent
- PayPal: $100 USD

### How long do deposits take?

| Method | Processing Time |
|--------|----------------|
| Bank transfer | 1-3 business days |
| Credit/debit card | Instant |
| Cryptocurrency | 10-30 minutes (depends on network) |
| PayPal | Instant |

### Are there deposit fees?

- Bank transfer: Free
- Credit/debit card: 2.9%
- Cryptocurrency: Network fees only
- PayPal: 3.5%

### How do I withdraw funds?

1. Log in to your account
2. Navigate to **Wallet > Withdrawal**
3. Select withdrawal method
4. Enter amount
5. Confirm withdrawal

Withdrawals are processed to the same method used for deposit (anti-money laundering requirement).

### What are the withdrawal limits?

- Minimum withdrawal: $50 USD
- Maximum withdrawal: $50,000 USD per transaction
- Daily limit: $100,000 USD
- Higher limits available for VIP accounts

### How long do withdrawals take?

| Method | Processing Time |
|--------|----------------|
| Bank transfer | 2-5 business days |
| Credit/debit card | 3-5 business days |
| Cryptocurrency | 1-24 hours |
| PayPal | 1-2 business days |

## Trading

### What instruments can I trade?

- **Forex**: 70+ currency pairs
- **Cryptocurrencies**: 20+ crypto pairs
- **Commodities**: Gold, Silver, Oil, etc.
- **Indices**: S&P 500, Nasdaq, DAX, etc.
- **Stocks**: 1000+ global stocks

See [Symbol Specifications](/docs/reference/symbols/forex) for the complete list.

### What is the minimum trade size?

- Forex: 0.01 lots (micro lot)
- Crypto: 0.001 BTC (varies by coin)
- Commodities: 0.1 lots
- Indices: 0.1 lots

### What leverage do you offer?

Leverage varies by instrument and account type:
- Forex: Up to 1:500
- Crypto: Up to 1:20
- Commodities: Up to 1:100
- Indices: Up to 1:100
- Stocks: Up to 1:5

### What are your trading hours?

- **Forex**: 24/5 (Sunday 22:00 GMT - Friday 22:00 GMT)
- **Crypto**: 24/7/365
- **Commodities**: Varies by instrument
- **Indices**: Varies by market
- **Stocks**: Exchange hours + extended hours

See [Trading Hours](/docs/reference/hours/overview) for detailed schedules.

### Do you allow hedging?

Yes, we allow hedging (opening opposite positions on the same instrument).

### Do you allow scalping?

Yes, scalping is allowed. We support all trading strategies.

### What is your order execution policy?

- Market execution (no dealing desk)
- No requotes
- Average execution time: <50ms
- Price improvement when possible

## Spreads & Fees

### What are your spreads?

Spreads vary by account type and instrument:

**Standard Account:**
- EUR/USD: From 1.5 pips
- GBP/USD: From 2.0 pips
- BTC/USD: From 10 pips

**Pro Account:**
- EUR/USD: From 0.2 pips
- GBP/USD: From 0.4 pips
- BTC/USD: From 2 pips

See [Fees & Commissions](/docs/reference/fees/spreads) for complete pricing.

### Do you charge commissions?

- **Standard accounts**: No commissions, wider spreads
- **Pro accounts**: $3.50 per lot per side, tighter spreads
- **VIP accounts**: Custom pricing

### What are swap rates?

Swap (rollover) rates are charged for positions held overnight:
- Long positions: May receive or pay swap
- Short positions: May receive or pay swap
- Rates depend on interest rate differential

See [Swap Rates](/docs/reference/fees/swap-rates) for current rates.

### Are there any hidden fees?

No. All fees are clearly disclosed:
- Spreads
- Commissions (if applicable)
- Swap rates
- Deposit/withdrawal fees

## Platform & Technical

### Which platforms do you support?

- Web platform (browser-based)
- Desktop app (Windows, macOS, Linux)
- Mobile apps (iOS, Android)
- MetaTrader 4/5
- TradingView integration
- API access (REST, WebSocket, FIX 4.4)

### Do you offer a demo account?

Yes, free demo account with:
- $10,000 virtual funds
- All features of live account
- Unlimited duration
- Reset balance anytime

### Can I use Expert Advisors (EAs)?

Yes, EAs are fully supported on:
- MetaTrader 4/5
- Our API (build custom bots)

### What are your API rate limits?

- REST API: 100 requests/second
- WebSocket: 10 concurrent connections
- FIX 4.4: 10,000 messages/second

See [API Documentation](/docs/api/overview) for details.

### Why am I getting "insufficient margin" error?

This occurs when you don't have enough margin to open a position:

**Solutions:**
1. Reduce position size
2. Deposit more funds
3. Close other positions to free up margin
4. Use lower leverage

Calculate required margin using our [Margin Calculator](/docs/trading-guide/margin/margin-requirements).

### My order wasn't executed. Why?

Common reasons:
1. **Limit orders**: Price didn't reach your limit price
2. **Market closed**: Trading hours ended
3. **Insufficient margin**: Not enough balance
4. **Invalid price**: Price too far from market
5. **Position limits**: Maximum positions reached

## Security

### Is my money safe?

Yes. We implement multiple security measures:
- Segregated client accounts
- Bank-grade encryption (AES-256)
- Two-factor authentication (2FA)
- Regular security audits
- Regulatory compliance

### Do you offer 2FA?

Yes, we strongly recommend enabling two-factor authentication:
1. Go to Settings > Security
2. Click "Enable 2FA"
3. Scan QR code with authenticator app
4. Enter verification code

Supported apps: Google Authenticator, Authy, Microsoft Authenticator

### What happens if I forget my password?

1. Click "Forgot Password" on login page
2. Enter your email address
3. Check email for reset link
4. Create new password
5. Log in with new password

## API & Integration

### How do I get API keys?

1. Log in to your account
2. Navigate to Settings > API Keys
3. Click "Create New API Key"
4. Set permissions (read-only or read-write)
5. Save your API key securely

### Can I use the API on a demo account?

Yes, API access is available for both demo and live accounts.

### Which programming languages are supported?

We provide official SDKs for:
- Python
- JavaScript/TypeScript
- Go
- Java
- C#

You can use any language that supports HTTP/WebSocket.

### Does the API support paper trading?

Yes, use our sandbox environment:
- URL: `https://api-sandbox.yourtradingplatform.com`
- Full feature parity with production
- Virtual funds
- No real money involved

## Regulations & Compliance

### Are you regulated?

Yes, we are regulated by [list your regulatory authorities]:
- [Authority 1]
- [Authority 2]

License numbers: [License numbers]

### Where are client funds held?

Client funds are held in segregated accounts at tier-1 banks:
- [Bank 1]
- [Bank 2]

### What is your negative balance protection?

You cannot lose more than your deposit. If your account goes negative, we absorb the loss.

## Still Have Questions?

### Contact Support

- **Email**: support@yourtradingplatform.com
- **Live Chat**: Available 24/7 on our website
- **Phone**: +1 (555) 123-4567
- **Discord**: [Join our community](https://discord.gg/yourtradingplatform)

### Response Times

- Live chat: < 2 minutes
- Email: < 4 hours during business hours
- Phone: Immediate

### Additional Resources

- [Getting Started Guide](/docs/getting-started/introduction)
- [Trading Guide](/docs/trading-guide/overview)
- [API Documentation](/docs/api/overview)
- [Video Tutorials](https://yourtradingplatform.com/tutorials)
