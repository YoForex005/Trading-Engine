# 🔐 Trading Engine - Environment Variables & Credentials

> **⚠️ CONFIDENTIAL DOCUMENT**  
> Store securely. Do not share or commit to version control.  
> Generated: 2026-01-19

---

## 📋 Table of Contents

1. [FIX API Credentials (T4B/YOFX)](#fix-api-credentials)
2. [Database Configuration](#database-configuration)
3. [Redis Cache](#redis-cache)
4. [JWT & Authentication](#jwt--authentication)
5. [Liquidity Provider APIs](#liquidity-provider-apis)
6. [Admin Configuration](#admin-configuration)
7. [Monitoring & Observability](#monitoring--observability)
8. [Complete .env Template](#complete-env-template)

---

## 🔗 FIX API Credentials

### T4B/YOFX FIX Connection

| Setting | Value |
|---------|-------|
| **Server IP** | `23.106.238.138` |
| **Port** | `12336` |
| **Protocol** | FIX 4.4 |
| **SSL** | `false` |

### YOFX1 - Trading Session

| Setting | Value |
|---------|-------|
| **Session ID** | `YOFX1` |
| **Purpose** | Order placement & trading operations |
| **SenderCompID** | `YOFX1` |
| **TargetCompID** | `YOFX` |
| **Username** | `YOFX1` |
| **Password** | `Brand#143` |
| **Trading Account** | `50153` |
| **Heartbeat** | `30 seconds` |

### YOFX2 - Market Data Session

| Setting | Value |
|---------|-------|
| **Session ID** | `YOFX2` |
| **Purpose** | Real-time market data feeds |
| **SenderCompID** | `YOFX2` |
| **TargetCompID** | `YOFX` |
| **Username** | `YOFX2` |
| **Password** | `Brand#143` |
| **Trading Account** | `50153` |
| **Heartbeat** | `30 seconds` |

---

## 🗄️ Database Configuration

### PostgreSQL

| Setting | Value |
|---------|-------|
| **Host** | `localhost` (dev) / `postgres` (docker) |
| **Port** | `5432` |
| **Database Name** | `trading_engine` |
| **Username** | `trading` |
| **Password** | `trading_pass` |
| **SSL Mode** | `disable` |

**Connection String:**
```
postgresql://trading:trading_pass@localhost:5432/trading_engine?sslmode=disable
```

**Docker Connection String:**
```
postgresql://trading:trading_pass@postgres:5432/trading_engine?sslmode=disable
```

---

## 🔴 Redis Cache

| Setting | Value |
|---------|-------|
| **Host** | `localhost` (dev) / `redis` (docker) |
| **Port** | `6379` |
| **Password** | `rtx_redis` (default) |
| **Database** | `0` |

**Connection String:**
```
redis://:rtx_redis@localhost:6379/0
```

---

## 🔑 JWT & Authentication

| Setting | Env Variable | Notes |
|---------|--------------|-------|
| **JWT Secret** | `JWT_SECRET` | Min 32 bytes. Generate with: `openssl rand -base64 32` |
| **JWT Expiry** | `JWT_EXPIRY` | Default: `24h` |
| **Master Encryption Key** | `MASTER_ENCRYPTION_KEY` | Required in production |
| **FIX Master Password** | `FIX_MASTER_PASSWORD` | For FIX credential encryption |

**Generate Secure JWT Secret:**
```bash
openssl rand -base64 32
```

---

## 💹 Liquidity Provider APIs

### OANDA

| Setting | Env Variable | Value |
|---------|--------------|-------|
| **API Endpoint** | - | `https://api-fxpractice.oanda.com` |
| **API Key** | `OANDA_API_KEY` | `977e1a77e25bac3a688011d6b0e845dd-8e3ab3a7682d9351af4c33be65e89b70` |
| **Account ID** | `OANDA_ACCOUNT_ID` | `101-004-37008470-002` |

> **Note**: These credentials were found in the archived `Trading-Engine/.env` file. Verify they are still valid.

### Binance

| Setting | Env Variable | Value |
|---------|--------------|-------|
| **API Key** | `BINANCE_API_KEY` | *Your Binance API Key* |
| **Secret Key** | `BINANCE_SECRET_KEY` | *Your Binance Secret Key* |

---

## 👤 Admin Configuration

| Setting | Env Variable | Default |
|---------|--------------|---------|
| **Admin Email** | `ADMIN_EMAIL` | `admin@example.com` |
| **Admin Password Hash** | `ADMIN_PASSWORD_HASH` | (bcrypt hash) |
| **IP Whitelist** | `ADMIN_IP_WHITELIST` | `127.0.0.1,::1` |

---

## 📊 Monitoring & Observability

### Grafana

| Setting | Env Variable | Default |
|---------|--------------|---------|
| **Admin Password** | `GRAFANA_PASSWORD` | `admin` |
| **Port** | - | `3000` |

### Prometheus

| Setting | Value |
|---------|-------|
| **Port** | `9091` |
| **Retention** | `30 days` |

### Jaeger (Tracing)

| Port | Purpose |
|------|---------|
| `5775/udp` | Agent (compact) |
| `6831/udp` | Agent (thrift) |
| `16686` | UI |
| `14268` | Collector |

---

## 📝 Complete .env Template

Copy this template and fill in your values:

```env
# ===========================================
# SERVER CONFIGURATION
# ===========================================
PORT=7999
ENVIRONMENT=development

# ===========================================
# DATABASE (PostgreSQL)
# ===========================================
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
DB_USER=trading
DB_PASSWORD=trading_pass
DB_SSL_MODE=disable

# Full connection string alternative:
# DATABASE_URL=postgresql://trading:trading_pass@localhost:5432/trading_engine?sslmode=disable

# ===========================================
# REDIS CACHE
# ===========================================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=rtx_redis
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10

# Full URL alternative:
# REDIS_URL=redis://:rtx_redis@localhost:6379/0

# ===========================================
# JWT AUTHENTICATION
# ===========================================
JWT_SECRET=your-secure-secret-here-min-32-bytes
JWT_EXPIRY=24h

# ===========================================
# ENCRYPTION
# ===========================================
MASTER_ENCRYPTION_KEY=your-master-key-here

# ===========================================
# ADMIN
# ===========================================
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD_HASH=your-bcrypt-hash
ADMIN_IP_WHITELIST=127.0.0.1,::1

# ===========================================
# BROKER SETTINGS
# ===========================================
BROKER_NAME=RTX Trading
BROKER_DISPLAY_NAME=YoForex
PRICE_FEED_LP=OANDA
PRICE_FEED_NAME=YoForex LP
EXECUTION_MODE=BBOOK
DEFAULT_LEVERAGE=100
DEFAULT_BALANCE=5000.0
MARGIN_MODE=HEDGING
MAX_TICKS_PER_SYMBOL=50000

# ===========================================
# DEFAULT ACCOUNT SETTINGS
# ===========================================
DEFAULT_ACCOUNT_BALANCE=10000.0
DEFAULT_ACCOUNT_LEVERAGE=100
DEFAULT_ACCOUNT_CURRENCY=USD

# ===========================================
# LIQUIDITY PROVIDERS
# ===========================================
# OANDA
OANDA_API_KEY=your-oanda-api-key
OANDA_ACCOUNT_ID=your-oanda-account-id

# Binance (optional)
BINANCE_API_KEY=your-binance-api-key
BINANCE_SECRET_KEY=your-binance-secret-key

# ===========================================
# FIX CONFIGURATION
# ===========================================
FIX_PROVISIONING_ENABLED=false
FIX_PROVISIONING_STORE_PATH=./data/fix_credentials
FIX_MASTER_PASSWORD=your-fix-master-password

# ===========================================
# CORS
# ===========================================
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# ===========================================
# COMPLIANCE
# ===========================================
COMPLIANCE_ENABLED=true
AUDIT_RETENTION_YEARS=7
COMPLIANCE_ARCHIVE_PATH=./data/compliance_reports
COMPLIANCE_AUTO_ARCHIVE=true
COMPLIANCE_TAMPER_PROOF=true
COMPLIANCE_ADMIN_ONLY=true
COMPLIANCE_MIFID_II=true
COMPLIANCE_SEC_RULE_606=true

# ===========================================
# MONITORING
# ===========================================
GRAFANA_PASSWORD=admin
LOG_LEVEL=info
```

---

## 🔧 Service URLs (Local Development)

| Service | URL |
|---------|-----|
| **Backend API** | http://localhost:7999 |
| **WebSocket** | ws://localhost:8080 |
| **Admin (Super)** | http://localhost:3000 |
| **Admin (Broker)** | http://localhost:3001 |
| **Desktop Client** | http://localhost:5173 |
| **Grafana** | http://localhost:3000 |
| **Prometheus** | http://localhost:9091 |
| **Jaeger UI** | http://localhost:16686 |

---

## 🚀 Quick Verification Commands

```bash
# Test PostgreSQL connection
psql postgresql://trading:trading_pass@localhost:5432/trading_engine -c "SELECT 1"

# Test Redis connection
redis-cli -h localhost -p 6379 -a rtx_redis PING

# Test FIX port accessibility
nc -zv 23.106.238.138 12336

# Test API health
curl http://localhost:7999/health
```

---

> **📌 Remember:**
> - Never commit real credentials to Git
> - Rotate passwords regularly
> - Use different credentials for production
> - Store this document in a secure location (encrypted drive, password manager)
