# External Integrations

**Analysis Date:** 2026-01-15

## APIs & External Services

**Binance (Cryptocurrency Market Data):**
- Purpose: Real-time crypto quotes (BTC, ETH, BNB, SOL, XRP, ADA, DOGE, MATIC, AVAX, DOT)
- Integration: WebSocket (wss://stream.binance.com:9443/stream)
- Files: `backend/binance/client.go`, `backend/lpmanager/adapters/binance.go`
- SDK/Client: Custom WebSocket client using gorilla/websocket
- Authentication: Public API (no auth required)
- Endpoints: Book ticker stream for best bid/ask

**OANDA (Forex Trading):**
- Purpose: FX instrument prices and order execution
- Integration: REST API + Streaming
- Files: `backend/oanda/client.go`, `backend/lpmanager/adapters/oanda.go`
- Endpoints:
  - Streaming: https://stream-fxpractice.oanda.com
  - REST: https://api-fxpractice.oanda.com
- Authentication: Bearer token (API key)
- **Security Issue**: API key hardcoded in `backend/cmd/server/main.go` line 23
- Credentials: `OANDA_API_KEY`, `OANDA_ACCOUNT_ID` (should use env vars)

**FlexyMarkets (Alternative Price Feed):**
- Purpose: Additional price feed service
- Integration: Socket.IO over WebSocket
- Files: `backend/flexymarkets/client.go`, `backend/lpmanager/adapters/flexy.go`
- URL: https://quotes.instantswap.app
- Authentication: None
- Protocol: Socket.IO v4 (EIO=4)

**Binance Public API (Historical Data):**
- Purpose: Historical OHLC candlestick data for charting
- Integration: REST API
- Files: `clients/desktop/src/services/ExternalDataService.ts`
- Endpoint: https://api.binance.com/api/v3/klines
- Authentication: Public API (no auth required)
- Usage: Fetch 500+ historical candles for chart initialization
- Symbol mapping: BTCUSD → BTCUSDT, ETHUSD → ETHUSDT

## Data Storage

**File-Based Storage:**
- Configuration: JSON files in `backend/data/lp_config.json`
- Tick history: `backend/data/ticks/{symbol}/{date}.json` (Git LFS)
- OHLC cache: `backend/data/ohlc/{symbol}/{timeframe}.json`
- Connection: File system I/O
- Client: Go standard library `os`, `io/ioutil`

**Browser Storage:**
- IndexedDB: Chart data caching (`clients/desktop/src/services/DataCache.ts`)
- localStorage: Indicator preferences (`clients/desktop/src/services/IndicatorStorage.ts`)
- Purpose: Reduce API calls, persist user preferences

## Authentication & Identity

**JWT Authentication:**
- Provider: Custom implementation
- Files: `backend/auth/token.go`, `backend/auth/service.go`
- Library: golang-jwt/jwt v5.3.0
- Token storage: Client-side (not specified - likely localStorage)
- **Security Issue**: Hardcoded JWT secret `"super_secret_dev_key_do_not_use_in_prod"` in `backend/auth/token.go` line 18

**Password Hashing:**
- Library: golang.org/x/crypto/bcrypt
- Files: `backend/auth/service.go`
- Method: bcrypt with default cost
- **Security Issue**: Plaintext password fallback for legacy accounts (lines 81-84)

## WebSocket Communication

**Internal WebSocket Gateway:**
- Purpose: Real-time price ticks and account updates
- Endpoint: ws://localhost:8080/ws
- Files: `backend/ws/hub.go`
- Protocol: Custom JSON-based tick streaming
- Features:
  - Auto-reconnect with 2-second retry
  - Throttled updates at 10 FPS
  - Latest price cache per symbol
- **Security Issue**: CORS disabled (`CheckOrigin: func(r *http.Request) bool { return true }`)

## Environment Configuration

**Development:**
- Required config: Hardcoded in source (major security gap)
- LP configuration: `backend/data/lp_config.json`
- Frontend URLs: Hardcoded `http://localhost:8080` in multiple files
- Secrets: Not managed via environment variables

**Production:**
- No production configuration found
- Secrets management: Not implemented (should use env vars)
- **Recommendation**: Implement .env file support, use environment variables for all secrets

## Configuration Files

**LP Manager Config:**
`backend/data/lp_config.json`:
```json
{
  "lps": [
    {
      "id": "binance",
      "enabled": true,
      "priority": 1,
      "name": "Binance",
      "type": "WebSocket"
    },
    {
      "id": "oanda",
      "enabled": false,
      "priority": 2,
      "name": "OANDA",
      "type": "REST/Streaming"
    },
    {
      "id": "flexy",
      "enabled": false,
      "priority": 3,
      "name": "FlexyMarkets",
      "type": "Socket.IO"
    }
  ]
}
```

## API Endpoints (Internal)

**Backend REST API:**
- Base URL: http://localhost:8080 (hardcoded in frontend)
- Authentication: `/login`
- Account: `/api/account/summary`
- Positions: `/api/positions`, `/api/positions/close`, `/api/positions/modify`
- Orders: `/api/orders/market`
- Symbols: `/api/symbols`
- Trades: `/api/trades`
- Ledger: `/api/ledger`
- Admin: `/admin/*` (deposits, withdrawals, symbol management)
- Config: `/api/config`

## External API Dependencies

**Dependencies:**
- Binance API availability
- OANDA API availability (when enabled)
- FlexyMarkets API availability (when enabled)

**Fallback:**
- LP Manager handles connection failures gracefully
- System continues with available LPs
- Frontend reconnects WebSocket automatically

## Rate Limits

**Not explicitly configured:**
- Binance: Public API rate limits apply
- OANDA: Practice account limits
- FlexyMarkets: Unknown

## Security Considerations

**CRITICAL Issues:**
1. Hardcoded OANDA API key in `backend/cmd/server/main.go`
2. Hardcoded JWT secret in `backend/auth/token.go`
3. CORS disabled for WebSocket connections

**Recommendations:**
1. Implement environment variable configuration
2. Create .env.example file
3. Enable CORS with whitelist
4. Rotate all hardcoded credentials
5. Use secrets management service

---

*Integration audit: 2026-01-15*
*Update when adding/removing external services*
