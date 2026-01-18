# Trading Engine Configuration Guide

## Quick Start

### 1. Copy Environment Template
```bash
cd backend
cp .env.example .env
```

### 2. Generate Secure Credentials

#### Admin Password Hash
```bash
# Using htpasswd (recommended)
echo -n "YourSecurePassword" | htpasswd -niB -C 10 admin | cut -d ":" -f 2

# Using bcrypt CLI (if installed)
bcrypt-cli hash YourSecurePassword

# Using Go
go run -o /tmp/hashgen <<EOF
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("YourSecurePassword"), 10)
    fmt.Println(string(hash))
}
EOF
/tmp/hashgen
```

#### JWT Secret
```bash
# Generate random 32+ character secret
openssl rand -base64 32
```

#### Master Encryption Key
```bash
# Generate base64 encoded 32-byte key
openssl rand -base64 32
```

### 3. Configure Your Environment

Edit `.env` and set your values:

```bash
# Security
ADMIN_PASSWORD_HASH=$2a$10$... (paste your generated hash)
JWT_SECRET=... (paste your generated secret)
MASTER_ENCRYPTION_KEY=... (paste your generated key)

# LP Credentials (if using)
OANDA_API_KEY=your_oanda_key
OANDA_ACCOUNT_ID=your_oanda_account_id

# Database
DB_PASSWORD=your_secure_db_password

# Customize broker settings as needed
BROKER_NAME=Your Broker Name
DEFAULT_ACCOUNT_BALANCE=10000.0
```

### 4. Start the Server
```bash
go run cmd/server/main.go
```

---

## Configuration Reference

### Required for Production

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | Secret key for JWT tokens | `openssl rand -base64 32` |
| `MASTER_ENCRYPTION_KEY` | Encryption key for sensitive data | `openssl rand -base64 32` |
| `ADMIN_PASSWORD_HASH` | Bcrypt hash of admin password | `$2a$10$...` |

### Highly Recommended

| Variable | Description | Default |
|----------|-------------|---------|
| `ADMIN_EMAIL` | Admin email address | `admin@example.com` |
| `DB_PASSWORD` | Database password | Empty |
| `ENVIRONMENT` | Environment name | `development` |
| `PORT` | Server port | `7999` |

### LP Configuration

| Variable | Description | Required For |
|----------|-------------|--------------|
| `OANDA_API_KEY` | OANDA API key | OANDA LP integration |
| `OANDA_ACCOUNT_ID` | OANDA account ID | OANDA LP integration |
| `BINANCE_API_KEY` | Binance API key | Binance LP integration |
| `BINANCE_SECRET_KEY` | Binance secret key | Binance LP integration |

### Broker Settings

| Variable | Description | Default |
|----------|-------------|---------|
| `BROKER_NAME` | Your broker name | `RTX Trading` |
| `EXECUTION_MODE` | `BBOOK` or `ABOOK` | `BBOOK` |
| `MARGIN_MODE` | `HEDGING` or `NETTING` | `HEDGING` |
| `DEFAULT_ACCOUNT_BALANCE` | Default account balance | `10000.0` |
| `DEFAULT_ACCOUNT_LEVERAGE` | Default leverage | `100` |
| `DEFAULT_BALANCE` | Demo account balance | `5000.0` |

---

## Environment-Specific Configurations

### Development (.env.development)
```bash
ENVIRONMENT=development
PORT=7999
DB_NAME=trading_engine_dev

# Relaxed security for local dev
JWT_EXPIRY=720h
LOG_LEVEL=debug
```

### Staging (.env.staging)
```bash
ENVIRONMENT=staging
PORT=8000
DB_NAME=trading_engine_staging

# Tighter security
JWT_EXPIRY=24h
LOG_LEVEL=info
```

### Production (.env.production)
```bash
ENVIRONMENT=production
PORT=80

# All required fields MUST be set
JWT_SECRET=...
MASTER_ENCRYPTION_KEY=...
ADMIN_PASSWORD_HASH=...
DB_PASSWORD=...

# Strict settings
JWT_EXPIRY=8h
LOG_LEVEL=warn
DB_SSL_MODE=require
```

---

## Loading Priority

Configuration is loaded in this order (later overrides earlier):

1. **Default values** (in `config/config.go`)
2. **`.env` file** (if present)
3. **Environment variables** (highest priority)

Example:
```bash
# .env file has PORT=7999
# But environment variable overrides it
PORT=8080 go run cmd/server/main.go  # Will use 8080
```

---

## Security Best Practices

### ✅ DO
- Use strong, unique passwords for admin account
- Generate random JWT secrets (32+ characters)
- Rotate secrets regularly (every 90 days)
- Use different secrets per environment
- Keep `.env` file in `.gitignore`
- Use infrastructure secrets management in production (AWS Secrets Manager, Vault)
- Enable SSL/TLS for database connections in production
- Restrict admin IP whitelist to known IPs

### ❌ DON'T
- Commit `.env` file to version control
- Share secrets via email or chat
- Use default/weak passwords
- Reuse secrets across environments
- Hard-code credentials in source code
- Store secrets in plaintext
- Use `ENVIRONMENT=development` in production

---

## Troubleshooting

### Error: "Failed to load configuration"
**Cause**: Invalid environment variable format
**Fix**: Check your `.env` file syntax (no spaces around `=`)

### Error: "JWT_SECRET is required in production"
**Cause**: Running in production mode without JWT_SECRET
**Fix**: Set `JWT_SECRET` environment variable

### Warning: "No ADMIN_PASSWORD_HASH provided"
**Cause**: Admin password not configured
**Fix**: Generate bcrypt hash and set `ADMIN_PASSWORD_HASH`

### Error: "OANDA credentials not configured"
**Cause**: OANDA API key or account ID missing
**Fix**: Set `OANDA_API_KEY` and `OANDA_ACCOUNT_ID` (or leave empty if not using OANDA)

### Server starts but can't login as admin
**Cause**: Wrong password or invalid hash
**Fix**: Regenerate password hash and update `.env`

---

## Migration from Hardcoded Values

If you're migrating from the old hardcoded configuration:

### Old (risk/engine.go)
```go
accounts: map[int64]*Account{
    1: {
        UserID:   "user_001",
        Balance:  10000.00,
        Leverage: 100,
    },
}
```

### New (via config)
```bash
# In .env file
DEFAULT_ACCOUNT_BALANCE=10000.0
DEFAULT_ACCOUNT_LEVERAGE=100
DEFAULT_ACCOUNT_CURRENCY=USD
```

```go
// In code - create account dynamically
cfg, _ := config.Load()
account := engine.CreateAccount(
    userID,
    username,
    password,
    isDemo,
)
account.Balance = cfg.DefaultAccount.Balance
account.Leverage = cfg.DefaultAccount.Leverage
```

---

## Validation

The configuration system automatically validates:

✅ Required fields in production mode
✅ Type conversions (int, float, bool)
✅ Array parsing (comma-separated values)

Run validation manually:
```go
cfg, err := config.Load()
if err != nil {
    log.Fatalf("Configuration error: %v", err)
}

if err := cfg.Validate(); err != nil {
    log.Fatalf("Validation failed: %v", err)
}
```

---

## Advanced Configuration

### Custom Config Location
```bash
# Load from custom path
ENV_FILE=/path/to/custom.env go run cmd/server/main.go
```

### Override Single Value
```bash
# Keep .env but override one value
PORT=8080 go run cmd/server/main.go
```

### Multiple Environments
```bash
# Development
ln -s .env.development .env
go run cmd/server/main.go

# Production
ln -s .env.production .env
go run cmd/server/main.go
```

---

## Docker Configuration

### Dockerfile
```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

# Don't copy .env - use environment variables
CMD ["./server"]
```

### docker-compose.yml
```yaml
version: '3.8'
services:
  backend:
    build: .
    ports:
      - "7999:7999"
    environment:
      - PORT=7999
      - ENVIRONMENT=production
      - JWT_SECRET=${JWT_SECRET}
      - ADMIN_PASSWORD_HASH=${ADMIN_PASSWORD_HASH}
      - DB_HOST=postgres
      - DB_PASSWORD=${DB_PASSWORD}
    env_file:
      - .env.production  # Load additional vars from file
```

### Running
```bash
# Set secrets in host environment
export JWT_SECRET=$(openssl rand -base64 32)
export ADMIN_PASSWORD_HASH='$2a$10$...'
export DB_PASSWORD='secure_password'

docker-compose up
```

---

## Monitoring Configuration

### Log Configuration Status
```go
log.Printf("Environment: %s", cfg.Environment)
log.Printf("Port: %s", cfg.Port)
log.Printf("Database: %s@%s:%s/%s",
    cfg.Database.User,
    cfg.Database.Host,
    cfg.Database.Port,
    cfg.Database.Name)
log.Printf("LP Configured: OANDA=%v, Binance=%v",
    cfg.LP.OandaAPIKey != "",
    cfg.LP.BinanceAPIKey != "")
```

### Health Check Endpoint
```go
http.HandleFunc("/health/config", func(w http.ResponseWriter, r *http.Request) {
    status := map[string]bool{
        "jwt_configured":        cfg.JWT.Secret != "",
        "admin_configured":      cfg.Admin.Password != "",
        "encryption_configured": cfg.Encryption.MasterKey != "",
        "db_configured":         cfg.Database.Password != "",
        "oanda_configured":      cfg.LP.OandaAPIKey != "",
    }
    json.NewEncoder(w).Encode(status)
})
```

---

## Support

For issues or questions:
1. Check this guide first
2. Review error messages in logs
3. Verify `.env` file syntax
4. Test with minimal configuration
5. Check `docs/MOCK_DATA_REMOVAL.md` for migration details
