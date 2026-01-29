# Error Codes and Handling Guide

## HTTP Status Codes

| Status Code | Name | Description |
|-------------|------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict (e.g., duplicate) |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Service temporarily unavailable |

## Error Response Format

All errors follow this structure:

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Example Error Response

```json
{
  "error": "Insufficient margin for order execution",
  "code": "MARGIN_INSUFFICIENT",
  "details": {
    "requiredMargin": 250.00,
    "availableMargin": 150.00,
    "symbol": "EURUSD",
    "volume": 0.5
  }
}
```

## Authentication Errors

### AUTH_REQUIRED

**HTTP Status:** 401

**Description:** No authentication token provided

**Example:**
```json
{
  "error": "Authentication required",
  "code": "AUTH_REQUIRED"
}
```

**Solution:** Include JWT token in Authorization header
```http
Authorization: Bearer <your-jwt-token>
```

### AUTH_INVALID

**HTTP Status:** 401

**Description:** Invalid or expired JWT token

**Example:**
```json
{
  "error": "Invalid authentication token",
  "code": "AUTH_INVALID"
}
```

**Solution:**
1. Check token format
2. Obtain new token via `/login`
3. Verify token not expired

### AUTH_EXPIRED

**HTTP Status:** 401

**Description:** JWT token has expired

**Example:**
```json
{
  "error": "Authentication token expired",
  "code": "AUTH_EXPIRED",
  "details": {
    "expiredAt": "2026-01-18T12:00:00Z"
  }
}
```

**Solution:** Re-authenticate via `/login` endpoint

### PERMISSION_DENIED

**HTTP Status:** 403

**Description:** User lacks required permissions

**Example:**
```json
{
  "error": "Admin access required",
  "code": "PERMISSION_DENIED",
  "details": {
    "requiredRole": "ADMIN",
    "userRole": "TRADER"
  }
}
```

**Solution:** Contact administrator for access

## Order Execution Errors

### INVALID_SYMBOL

**HTTP Status:** 400

**Description:** Symbol not found or disabled

**Example:**
```json
{
  "error": "Symbol not found or disabled",
  "code": "INVALID_SYMBOL",
  "details": {
    "symbol": "INVALID",
    "availableSymbols": ["EURUSD", "BTCUSD", "XAUUSD"]
  }
}
```

**Solution:** Use valid symbol from `/api/symbols`

### INVALID_VOLUME

**HTTP Status:** 400

**Description:** Invalid lot size

**Example:**
```json
{
  "error": "Invalid volume",
  "code": "INVALID_VOLUME",
  "details": {
    "volume": 0.001,
    "minVolume": 0.01,
    "maxVolume": 100.0,
    "volumeStep": 0.01
  }
}
```

**Solution:** Adjust volume to meet constraints

### MARGIN_INSUFFICIENT

**HTTP Status:** 400

**Description:** Not enough free margin

**Example:**
```json
{
  "error": "Insufficient margin for order execution",
  "code": "MARGIN_INSUFFICIENT",
  "details": {
    "requiredMargin": 500.00,
    "freeMargin": 300.00,
    "symbol": "EURUSD",
    "volume": 1.0
  }
}
```

**Solution:**
1. Reduce order size
2. Close existing positions
3. Deposit funds

### MARKET_CLOSED

**HTTP Status:** 400

**Description:** Market is closed for trading

**Example:**
```json
{
  "error": "Market closed for trading",
  "code": "MARKET_CLOSED",
  "details": {
    "symbol": "XAUUSD",
    "opensAt": "2026-01-19T22:00:00Z"
  }
}
```

**Solution:** Wait for market to open

### PRICE_STALE

**HTTP Status:** 400

**Description:** No recent price data available

**Example:**
```json
{
  "error": "No price data available",
  "code": "PRICE_STALE",
  "details": {
    "symbol": "EURUSD",
    "lastUpdate": "2026-01-18T10:00:00Z"
  }
}
```

**Solution:**
1. Wait for LP reconnection
2. Check `/admin/lp-status`

### INVALID_SL_TP

**HTTP Status:** 400

**Description:** Invalid stop loss or take profit

**Example:**
```json
{
  "error": "Invalid stop loss or take profit",
  "code": "INVALID_SL_TP",
  "details": {
    "side": "BUY",
    "currentPrice": 1.0950,
    "sl": 1.1000,
    "tp": 1.0900,
    "reason": "SL must be below entry for BUY orders"
  }
}
```

**Solution:** Adjust SL/TP levels:
- **BUY:** SL < entry, TP > entry
- **SELL:** SL > entry, TP < entry

### ORDER_NOT_FOUND

**HTTP Status:** 404

**Description:** Order ID not found

**Example:**
```json
{
  "error": "Order not found",
  "code": "ORDER_NOT_FOUND",
  "details": {
    "orderId": "ORD-12345"
  }
}
```

**Solution:** Verify order ID

### ORDER_NOT_PENDING

**HTTP Status:** 400

**Description:** Cannot cancel filled/closed order

**Example:**
```json
{
  "error": "Order cannot be cancelled",
  "code": "ORDER_NOT_PENDING",
  "details": {
    "orderId": "ORD-12345",
    "status": "FILLED"
  }
}
```

**Solution:** Only pending orders can be cancelled

## Position Errors

### POSITION_NOT_FOUND

**HTTP Status:** 404

**Description:** Position ID not found

**Example:**
```json
{
  "error": "Position not found",
  "code": "POSITION_NOT_FOUND",
  "details": {
    "positionId": 12345
  }
}
```

**Solution:** Verify position ID from `/api/positions`

### POSITION_ALREADY_CLOSED

**HTTP Status:** 400

**Description:** Attempting to close already closed position

**Example:**
```json
{
  "error": "Position already closed",
  "code": "POSITION_ALREADY_CLOSED",
  "details": {
    "positionId": 12345,
    "closedAt": "2026-01-18T10:30:00Z"
  }
}
```

**Solution:** Refresh position list

### PARTIAL_CLOSE_INVALID

**HTTP Status:** 400

**Description:** Invalid partial close volume

**Example:**
```json
{
  "error": "Invalid partial close volume",
  "code": "PARTIAL_CLOSE_INVALID",
  "details": {
    "positionVolume": 1.0,
    "closeVolume": 1.5,
    "reason": "Close volume exceeds position volume"
  }
}
```

**Solution:** Close volume must be ≤ position volume

## Account Errors

### ACCOUNT_NOT_FOUND

**HTTP Status:** 404

**Description:** Account ID not found

**Example:**
```json
{
  "error": "Account not found",
  "code": "ACCOUNT_NOT_FOUND",
  "details": {
    "accountId": 999
  }
}
```

**Solution:** Verify account ID

### ACCOUNT_DISABLED

**HTTP Status:** 403

**Description:** Account is disabled

**Example:**
```json
{
  "error": "Account disabled",
  "code": "ACCOUNT_DISABLED",
  "details": {
    "accountId": 123,
    "reason": "Account suspended by admin"
  }
}
```

**Solution:** Contact support

### INSUFFICIENT_BALANCE

**HTTP Status:** 400

**Description:** Not enough balance for withdrawal

**Example:**
```json
{
  "error": "Insufficient balance",
  "code": "INSUFFICIENT_BALANCE",
  "details": {
    "balance": 500.00,
    "withdrawalAmount": 1000.00
  }
}
```

**Solution:** Reduce withdrawal amount

### WITHDRAWAL_LIMIT_EXCEEDED

**HTTP Status:** 400

**Description:** Daily/monthly withdrawal limit exceeded

**Example:**
```json
{
  "error": "Withdrawal limit exceeded",
  "code": "WITHDRAWAL_LIMIT_EXCEEDED",
  "details": {
    "dailyLimit": 10000.00,
    "currentUsage": 9500.00,
    "requestedAmount": 1000.00
  }
}
```

**Solution:** Wait for limit reset or request increase

## LP and FIX Session Errors

### LP_NOT_CONNECTED

**HTTP Status:** 503

**Description:** Liquidity provider not connected

**Example:**
```json
{
  "error": "No LP connection available",
  "code": "LP_NOT_CONNECTED",
  "details": {
    "lp": "YOFX",
    "lastConnected": "2026-01-18T09:00:00Z"
  }
}
```

**Solution:**
1. Check `/admin/lp-status`
2. Reconnect via `/admin/fix/connect`

### FIX_SESSION_NOT_FOUND

**HTTP Status:** 404

**Description:** FIX session ID not found

**Example:**
```json
{
  "error": "FIX session not found",
  "code": "FIX_SESSION_NOT_FOUND",
  "details": {
    "sessionId": "INVALID"
  }
}
```

**Solution:** Use valid session ID (e.g., "YOFX1")

### LP_TIMEOUT

**HTTP Status:** 504

**Description:** LP response timeout

**Example:**
```json
{
  "error": "LP timeout",
  "code": "LP_TIMEOUT",
  "details": {
    "lp": "OANDA",
    "timeout": 5000,
    "symbol": "EURUSD"
  }
}
```

**Solution:** Retry or use different LP

## Validation Errors

### INVALID_REQUEST

**HTTP Status:** 400

**Description:** Malformed request body

**Example:**
```json
{
  "error": "Invalid request body",
  "code": "INVALID_REQUEST",
  "details": {
    "parseError": "Unexpected token at position 15"
  }
}
```

**Solution:** Check JSON syntax

### MISSING_PARAMETER

**HTTP Status:** 400

**Description:** Required parameter missing

**Example:**
```json
{
  "error": "Missing required parameter",
  "code": "MISSING_PARAMETER",
  "details": {
    "missingFields": ["symbol", "volume"]
  }
}
```

**Solution:** Provide all required fields

### INVALID_PARAMETER_TYPE

**HTTP Status:** 400

**Description:** Parameter has wrong type

**Example:**
```json
{
  "error": "Invalid parameter type",
  "code": "INVALID_PARAMETER_TYPE",
  "details": {
    "field": "volume",
    "expected": "number",
    "received": "string"
  }
}
```

**Solution:** Use correct data types

## Rate Limiting Errors

### RATE_LIMIT_EXCEEDED

**HTTP Status:** 429

**Description:** Too many requests

**Example:**
```json
{
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "details": {
    "limit": 100,
    "window": "1m",
    "retryAfter": 45
  }
}
```

**Solution:** Wait before retrying

**Headers:**
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1705838445
Retry-After: 45
```

## System Errors

### INTERNAL_ERROR

**HTTP Status:** 500

**Description:** Unexpected server error

**Example:**
```json
{
  "error": "Internal server error",
  "code": "INTERNAL_ERROR",
  "details": {
    "errorId": "err_123456789"
  }
}
```

**Solution:** Contact support with errorId

### SERVICE_UNAVAILABLE

**HTTP Status:** 503

**Description:** Service temporarily unavailable

**Example:**
```json
{
  "error": "Service unavailable",
  "code": "SERVICE_UNAVAILABLE",
  "details": {
    "reason": "Database maintenance",
    "estimatedRestore": "2026-01-18T12:00:00Z"
  }
}
```

**Solution:** Retry after maintenance window

### DATABASE_ERROR

**HTTP Status:** 500

**Description:** Database operation failed

**Example:**
```json
{
  "error": "Database error",
  "code": "DATABASE_ERROR",
  "details": {
    "operation": "INSERT"
  }
}
```

**Solution:** Retry or contact support

## Error Handling Best Practices

### Client-Side Error Handling

```javascript
async function placeOrder(orderData) {
  try {
    const response = await fetch('http://localhost:7999/api/orders/market', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(orderData)
    });

    if (!response.ok) {
      const error = await response.json();
      throw new APIError(error.code, error.error, error.details);
    }

    return await response.json();
  } catch (error) {
    handleError(error);
  }
}

function handleError(error) {
  switch (error.code) {
    case 'AUTH_EXPIRED':
      // Redirect to login
      window.location.href = '/login';
      break;

    case 'MARGIN_INSUFFICIENT':
      // Show margin warning
      showMarginWarning(error.details);
      break;

    case 'RATE_LIMIT_EXCEEDED':
      // Retry after delay
      setTimeout(() => retry(), error.details.retryAfter * 1000);
      break;

    default:
      // Generic error message
      showErrorToast(error.message);
  }
}
```

### Python Error Handling

```python
import requests
from typing import Dict, Any

class RTXAPIError(Exception):
    def __init__(self, code: str, message: str, details: Dict[str, Any] = None):
        self.code = code
        self.message = message
        self.details = details or {}
        super().__init__(self.message)

def place_order(token: str, order_data: dict) -> dict:
    try:
        response = requests.post(
            'http://localhost:7999/api/orders/market',
            headers={
                'Authorization': f'Bearer {token}',
                'Content-Type': 'application/json'
            },
            json=order_data,
            timeout=10
        )

        if not response.ok:
            error = response.json()
            raise RTXAPIError(
                error.get('code', 'UNKNOWN'),
                error.get('error', 'Unknown error'),
                error.get('details')
            )

        return response.json()

    except requests.exceptions.Timeout:
        raise RTXAPIError('TIMEOUT', 'Request timeout')
    except requests.exceptions.ConnectionError:
        raise RTXAPIError('CONNECTION_ERROR', 'Cannot connect to server')

# Usage
try:
    result = place_order(token, {
        'symbol': 'EURUSD',
        'side': 'BUY',
        'volume': 0.1
    })
except RTXAPIError as e:
    if e.code == 'MARGIN_INSUFFICIENT':
        print(f"Not enough margin: {e.details}")
    elif e.code == 'AUTH_EXPIRED':
        # Re-authenticate
        token = login()
    else:
        print(f"Error: {e.message}")
```

### Retry Logic with Exponential Backoff

```javascript
async function retryWithBackoff(fn, maxRetries = 3) {
  let delay = 1000; // Start with 1 second

  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (i === maxRetries - 1) throw error;

      if (error.code === 'RATE_LIMIT_EXCEEDED') {
        // Use Retry-After if available
        delay = (error.details.retryAfter || delay / 1000) * 1000;
      }

      await new Promise(resolve => setTimeout(resolve, delay));
      delay *= 2; // Exponential backoff
    }
  }
}

// Usage
const order = await retryWithBackoff(() => placeOrder(orderData));
```

## Summary Table

| Error Code | HTTP | Retry? | Description |
|------------|------|--------|-------------|
| AUTH_REQUIRED | 401 | ❌ | Add auth token |
| AUTH_INVALID | 401 | ❌ | Re-authenticate |
| AUTH_EXPIRED | 401 | ❌ | Re-authenticate |
| PERMISSION_DENIED | 403 | ❌ | Contact admin |
| INVALID_SYMBOL | 400 | ❌ | Use valid symbol |
| MARGIN_INSUFFICIENT | 400 | ❌ | Deposit or reduce size |
| MARKET_CLOSED | 400 | ✅ | Wait for open |
| PRICE_STALE | 400 | ✅ | Wait for LP |
| ORDER_NOT_FOUND | 404 | ❌ | Verify ID |
| LP_NOT_CONNECTED | 503 | ✅ | Wait for connection |
| RATE_LIMIT_EXCEEDED | 429 | ✅ | Exponential backoff |
| INTERNAL_ERROR | 500 | ✅ | Report to support |
| SERVICE_UNAVAILABLE | 503 | ✅ | Wait for service |

**Legend:**
- ✅ Retry recommended with exponential backoff
- ❌ Do not retry, fix error condition first
