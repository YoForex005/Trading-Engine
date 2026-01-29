# RTX Trading Engine API Documentation

Complete API documentation for the RTX Trading Engine v3.0

## Documentation Index

### 📘 Main Documentation

- **[API.md](./API.md)** - Complete API reference with all endpoints, examples, and usage guide
- **[openapi.yaml](./openapi.yaml)** - OpenAPI 3.0 specification (machine-readable)

### 🔐 Authentication

- **[AUTHENTICATION.md](./AUTHENTICATION.md)** - JWT authentication flow, token management, and security best practices

### 🌐 Real-time Data

- **[WEBHOOKS.md](./WEBHOOKS.md)** - WebSocket API for real-time market data streaming

### ⚠️ Error Handling

- **[ERRORS.md](./ERRORS.md)** - Complete error codes reference, error handling patterns, and troubleshooting

### ⏱️ Rate Limiting

- **[RATE_LIMITS.md](./RATE_LIMITS.md)** - API rate limits, quotas, and best practices

### 📮 Postman Collection

- **[postman-collection.json](./postman-collection.json)** - Ready-to-import Postman collection for testing

## Quick Links

| Resource | Description |
|----------|-------------|
| **Base URL** | `http://localhost:7999` (dev) |
| **WebSocket** | `ws://localhost:7999/ws` |
| **API Docs** | `http://localhost:7999/docs` |
| **OpenAPI Spec** | `http://localhost:7999/swagger.yaml` |

## Getting Started

### 1. Start the Server

```bash
cd backend
go run cmd/server/main.go
```

Server will start on `http://localhost:7999`

### 2. Test the Connection

```bash
curl http://localhost:7999/health
# Expected: OK
```

### 3. Login

```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "0",
    "username": "admin",
    "role": "ADMIN"
  }
}
```

### 4. Make Authenticated Request

```bash
TOKEN="your-jwt-token-here"

curl -X GET http://localhost:7999/api/account/summary \
  -H "Authorization: Bearer $TOKEN"
```

## Documentation Structure

```
docs/api/
├── README.md                    # This file
├── API.md                       # Complete API reference
├── openapi.yaml                 # OpenAPI 3.0 specification
├── AUTHENTICATION.md            # JWT authentication guide
├── WEBHOOKS.md                  # WebSocket API documentation
├── ERRORS.md                    # Error codes and handling
├── RATE_LIMITS.md              # Rate limiting information
└── postman-collection.json      # Postman collection
```

## API Categories

### Authentication
- Login and JWT token management
- Admin and trader account authentication

### B-Book Orders (Internal Execution)
- Market orders executed using internal balance
- Position management with internal P/L tracking

### A-Book Orders (LP Passthrough)
- Orders routed to OANDA, Binance, or YoFX
- FIX 4.4 protocol support for institutional LPs
- Market, limit, stop, and stop-limit orders

### Position Management
- Open/close positions
- Modify stop loss and take profit
- Partial position close
- Trailing stops (fixed, step, ATR)

### Account Information
- Balance, equity, margin
- Unrealized and realized P/L
- Margin level calculations

### Market Data
- Real-time tick data
- OHLC (candlestick) data
- Multiple timeframes (1m, 5m, 15m, 1h, 4h, 1d)

### Risk Management
- Position sizing calculators
- Margin preview
- Risk percentage calculations

### Admin Endpoints
- Account management
- LP configuration
- FIX session management
- Broker configuration

## Using the Postman Collection

### Import Collection

1. Open Postman
2. Click **Import**
3. Select `postman-collection.json`
4. Collection will be imported with all endpoints

### Configure Variables

1. Go to collection **Variables** tab
2. Set `base_url` (default: `http://localhost:7999`)
3. Login using "Authentication > Login (Admin)"
4. Token will be automatically saved to `jwt_token` variable

### Make Requests

All requests are pre-configured with:
- Correct HTTP methods
- Required headers
- Example request bodies
- Bearer token authentication

## Interactive API Documentation

### Swagger UI (Planned)

```bash
# Serve OpenAPI spec with Swagger UI
npx swagger-ui-serve openapi.yaml
```

Visit: `http://localhost:8080`

### Redoc (Alternative)

```bash
# Serve with Redoc
npx redoc-cli serve openapi.yaml
```

## Code Examples

### JavaScript/TypeScript

```typescript
// Login
const response = await fetch('http://localhost:7999/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ username: 'admin', password: 'password' })
});

const { token } = await response.json();

// Place order
const order = await fetch('http://localhost:7999/api/orders/market', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    symbol: 'EURUSD',
    side: 'BUY',
    volume: 0.1
  })
});
```

### Python

```python
import requests

# Login
response = requests.post('http://localhost:7999/login', json={
    'username': 'admin',
    'password': 'password'
})
token = response.json()['token']

# Place order
response = requests.post(
    'http://localhost:7999/api/orders/market',
    headers={'Authorization': f'Bearer {token}'},
    json={'symbol': 'EURUSD', 'side': 'BUY', 'volume': 0.1}
)
```

### Go

```go
// Login
body := bytes.NewBuffer([]byte(`{"username":"admin","password":"password"}`))
resp, _ := http.Post("http://localhost:7999/login", "application/json", body)

var loginResp struct {
    Token string `json:"token"`
}
json.NewDecoder(resp.Body).Decode(&loginResp)

// Place order
req, _ := http.NewRequest("POST", "http://localhost:7999/api/orders/market",
    bytes.NewBuffer([]byte(`{"symbol":"EURUSD","side":"BUY","volume":0.1}`)))
req.Header.Set("Authorization", "Bearer "+loginResp.Token)
req.Header.Set("Content-Type", "application/json")
```

### curl

```bash
# Complete workflow
TOKEN=$(curl -s -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r '.token')

# Get account summary
curl -X GET http://localhost:7999/api/account/summary \
  -H "Authorization: Bearer $TOKEN"

# Place order
curl -X POST http://localhost:7999/api/orders/market \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"symbol":"EURUSD","side":"BUY","volume":0.1}'

# Get positions
curl -X GET http://localhost:7999/api/positions \
  -H "Authorization: Bearer $TOKEN"
```

## WebSocket Example

```javascript
const ws = new WebSocket('ws://localhost:7999/ws');

ws.onopen = () => {
  console.log('Connected to RTX Trading Engine');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === 'tick') {
    console.log(`${data.symbol}: Bid=${data.bid} Ask=${data.ask} LP=${data.lp}`);
    updateChart(data);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected, reconnecting...');
  setTimeout(connect, 3000);
};
```

## Key Concepts

### B-Book vs A-Book

**B-Book (Internal Execution):**
- Orders executed against internal balance
- No LP connection required
- Instant execution
- Full control over pricing and execution
- Use endpoints: `/api/orders/*`, `/api/positions/*`

**A-Book (LP Passthrough):**
- Orders routed to external liquidity providers
- Requires LP connection (OANDA, Binance, YoFX)
- Real market execution
- Price depends on LP
- Use endpoints: `/order`, `/order/limit`, `/order/stop`

### Execution Mode Toggle

Admin can switch between B-Book and A-Book:

```bash
curl -X POST http://localhost:7999/admin/execution-mode \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"mode":"ABOOK"}'
```

**Modes:**
- `BBOOK`: Internal execution (default)
- `ABOOK`: LP passthrough

### Margin Modes

**Hedging Mode (Default):**
- Multiple positions per symbol allowed
- Each position independent
- Can have both BUY and SELL on same symbol

**Netting Mode:**
- One net position per symbol
- Opposite orders reduce position size
- Simpler P/L calculation

## Testing

### Unit Tests

```bash
# Test API endpoints
go test ./api/...

# Test handlers
go test ./internal/api/handlers/...
```

### Integration Tests

```bash
# Start server
go run cmd/server/main.go

# Run integration tests
go test ./tests/integration/...
```

### Manual Testing

Use Postman collection or curl examples above.

## Troubleshooting

### Cannot Connect

**Problem:** Connection refused

**Solution:**
1. Verify server is running: `curl http://localhost:7999/health`
2. Check port 7999 is not in use: `lsof -i :7999`
3. Check firewall settings

### 401 Unauthorized

**Problem:** All requests return 401

**Solution:**
1. Verify token is included: `Authorization: Bearer <token>`
2. Check token not expired (24h lifetime)
3. Re-authenticate via `/login`

### No Market Data

**Problem:** WebSocket connected but no ticks

**Solution:**
1. Check LP status: `GET /admin/lp-status`
2. Verify symbols enabled: `GET /api/config`
3. Check LP connections in server logs

### Orders Not Executing

**Problem:** Order placement fails

**Solution:**
1. Check margin available: `GET /api/account/summary`
2. Verify symbol exists: `GET /api/symbols`
3. Check execution mode matches endpoint (B-Book vs A-Book)

## Support

- **GitHub Issues:** https://github.com/epic1st/rtx/issues
- **Email:** support@rtxtrading.com
- **Documentation:** http://localhost:7999/docs

## Version History

- **v3.0.0** (2026-01-18)
  - Complete API documentation
  - OpenAPI 3.0 specification
  - Postman collection
  - WebSocket documentation
  - Error reference
  - Rate limiting guide

## License

Proprietary - RTX Trading Engine

---

**Last Updated:** 2026-01-18
**API Version:** 3.0.0
