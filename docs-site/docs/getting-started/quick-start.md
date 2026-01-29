---
id: quick-start
title: Quick Start Guide
sidebar_label: Quick Start
sidebar_position: 2
description: Get up and running with Trading Platform in 5 minutes
keywords:
  - quick start
  - getting started
  - first trade
  - tutorial
---

# Quick Start Guide

Get started with Trading Platform in just 5 minutes. This guide will walk you through creating an account, making your first deposit, and executing your first trade.

## Prerequisites

Before you begin, make sure you have:

- A valid email address
- Government-issued ID (for verification)
- Initial deposit (minimum $100 USD)
- Basic understanding of trading concepts

## Step 1: Create Your Account

### Sign Up

1. Visit [yourtradingplatform.com/register](https://yourtradingplatform.com/register)
2. Fill in your details:
   - Email address
   - Strong password (min 12 characters)
   - Country of residence
   - Accept Terms of Service

3. Click "Create Account"

```mdx-code-block
import CreateAccountDemo from '@site/src/components/CreateAccountDemo';

<CreateAccountDemo />
```

### Verify Your Email

1. Check your inbox for verification email
2. Click the verification link
3. Your account is now active!

## Step 2: Complete KYC Verification

For your security and regulatory compliance:

1. Log in to your account
2. Navigate to **Settings > Verification**
3. Upload required documents:
   - Government-issued ID (passport, driver's license)
   - Proof of address (utility bill, bank statement)
4. Wait for approval (typically 1-2 hours)

:::tip Fast Track Verification
Complete your verification during business hours (9 AM - 5 PM EST) for faster approval.
:::

## Step 3: Make Your First Deposit

### Deposit Methods

We support multiple deposit methods:

| Method | Processing Time | Min Deposit | Fees |
|--------|----------------|-------------|------|
| Bank Transfer | 1-3 business days | $100 | Free |
| Credit/Debit Card | Instant | $100 | 2.9% |
| Crypto (USDT) | 10-30 minutes | $100 | Network fees |
| PayPal | Instant | $100 | 3.5% |

### Make a Deposit

1. Go to **Wallet > Deposit**
2. Select your preferred method
3. Enter deposit amount
4. Follow the payment instructions
5. Funds will appear in your account after processing

```bash
# Example: Check your account balance via API
curl -X GET https://api.yourtradingplatform.com/v1/account/balance \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json"
```

Response:
```json
{
  "balance": 1000.00,
  "currency": "USD",
  "available": 1000.00,
  "margin_used": 0.00,
  "margin_available": 1000.00
}
```

## Step 4: Explore the Platform

### Web Platform

1. Log in at [app.yourtradingplatform.com](https://app.yourtradingplatform.com)
2. Familiarize yourself with the interface:
   - **Watchlist**: Monitor your favorite instruments
   - **Charts**: View real-time price charts
   - **Order Panel**: Place and manage orders
   - **Positions**: View open positions
   - **History**: Review past trades

### Key Interface Elements

```
+----------------------------------------------------------+
|  EURUSD  |  BTCUSD  |  XAUUSD  |  SPX500  |  [Watchlist] |
+----------------------------------------------------------+
|                                                           |
|  [Price Chart - TradingView]                             |
|                                                           |
|  EUR/USD: 1.0850                                         |
|                                                           |
+---------------------------+-------------------------------+
|  Order Panel              |  Open Positions              |
|  Symbol: EURUSD          |  EURUSD: +0.5 lots           |
|  Type: Market            |  P&L: +$45.00                |
|  Volume: 0.1 lots        |                              |
|  [BUY] [SELL]           |                              |
+---------------------------+-------------------------------+
```

## Step 5: Place Your First Trade

### Understanding the Basics

Before trading, understand these key concepts:

- **Symbol**: The instrument you're trading (e.g., EUR/USD)
- **Volume**: The size of your trade (lots)
- **Leverage**: Multiply your buying power (e.g., 1:100)
- **Spread**: Difference between bid and ask price
- **Pip**: Smallest price movement (e.g., 0.0001 for EUR/USD)

### Execute a Market Order

Let's execute a simple buy order for EUR/USD:

1. **Select Symbol**: Choose EUR/USD from the watchlist
2. **Set Volume**: Enter 0.01 lots (micro lot)
3. **Optional**: Set Stop Loss and Take Profit
4. **Click BUY** or **SELL**

```javascript
// Example: Place a market order via API
const response = await fetch('https://api.yourtradingplatform.com/v1/orders', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    symbol: 'EURUSD',
    type: 'market',
    side: 'buy',
    volume: 0.01,
    stop_loss: 1.0800,
    take_profit: 1.0900
  })
});

const order = await response.json();
console.log('Order placed:', order);
```

Response:
```json
{
  "order_id": "ORD-123456789",
  "symbol": "EURUSD",
  "type": "market",
  "side": "buy",
  "volume": 0.01,
  "price": 1.0850,
  "stop_loss": 1.0800,
  "take_profit": 1.0900,
  "status": "filled",
  "timestamp": "2026-01-18T14:30:00Z"
}
```

:::success Congratulations!
You've successfully placed your first trade! 🎉
:::

## Step 6: Monitor Your Position

### Real-Time Monitoring

After placing a trade:

1. Go to **Positions** tab
2. View your open position:
   - Entry price
   - Current price
   - Profit/Loss (real-time)
   - Stop Loss / Take Profit levels

3. Monitor via WebSocket for real-time updates:

```javascript
const ws = new WebSocket('wss://api.yourtradingplatform.com/v1/stream');

ws.onopen = () => {
  ws.send(JSON.stringify({
    action: 'subscribe',
    channel: 'positions',
    api_key: 'YOUR_API_KEY'
  }));
};

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Position update:', update);
};
```

### Close Your Position

When you're ready to close:

1. Go to **Positions** tab
2. Find your EUR/USD position
3. Click **Close**
4. Confirm the close

Or via API:

```javascript
const response = await fetch('https://api.yourtradingplatform.com/v1/positions/close', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    position_id: 'POS-123456789'
  })
});
```

## Pro Tips for New Traders

### Start Small
- Begin with micro lots (0.01)
- Use demo account first
- Don't risk more than 1-2% per trade

### Use Risk Management
- **Always** set Stop Loss orders
- Use position sizing calculator
- Monitor your margin level

### Learn Continuously
- Read our [Trading Guide](/docs/trading-guide/overview)
- Watch educational videos
- Join our community Discord

### Use the API
- Automate your strategies
- Backtest before going live
- Start with our [API documentation](/docs/api/overview)

## What's Next?

Now that you've completed your first trade, explore:

1. **[Trading Guide](/docs/trading-guide/overview)** - Learn advanced trading concepts
2. **[API Documentation](/docs/api/overview)** - Build automated strategies
3. **[Risk Management](/docs/trading-guide/risk/overview)** - Protect your capital
4. **[Integrations](/docs/integrations/overview)** - Connect with MetaTrader, TradingView

## Need Help?

If you encounter any issues:

- Check our [FAQ](/docs/troubleshooting/faq)
- Visit [Support Center](https://support.yourtradingplatform.com)
- Email: support@yourtradingplatform.com
- Live Chat: Available 24/7

Happy trading! 📈
