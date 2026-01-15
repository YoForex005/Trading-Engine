# Local Development Environment Guide

This guide helps you set up and run the Trading Engine platform on your local machine for development.

## Prerequisites

Before starting, ensure you have the following installed:

- **Docker** and **Docker Compose** (version 2.0+)
- **Go** 1.21 or higher (for backend development)
- **Bun** (for frontend development)
- **Git** (for version control)
- **PostgreSQL client tools** (optional, for database access)

### Installing Prerequisites

**macOS:**
```bash
# Install Docker Desktop
brew install --cask docker

# Install Go
brew install go

# Install Bun
curl -fsSL https://bun.sh/install | bash

# Install PostgreSQL client (optional)
brew install postgresql@16
```

**Linux (Ubuntu/Debian):**
```bash
# Install Docker
sudo apt-get update
sudo apt-get install docker.io docker-compose-plugin

# Install Go
sudo apt-get install golang-1.21

# Install Bun
curl -fsSL https://bun.sh/install | bash

# Install PostgreSQL client (optional)
sudo apt-get install postgresql-client
```

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/YOUR_ORG/trading-engine.git
cd trading-engine
```

### 2. Configure Environment

Create a `.env` file in the project root:

```bash
# Database configuration
DATABASE_URL=postgresql://postgres:password@localhost:5432/trading_engine

# Redis configuration
REDIS_URL=localhost:6379

# JWT configuration
JWT_SECRET=your-development-secret-change-in-production

# OANDA API configuration
OANDA_API_KEY=your-oanda-api-key
OANDA_ACCOUNT_ID=your-oanda-account-id

# Logging
LOG_LEVEL=debug
```

### 3. Start All Services

```bash
docker-compose up -d
```

This starts:
- PostgreSQL (port 5432)
- Redis (port 6379)
- Backend API (port 8080)
- Frontend UI (port 5173)
- Prometheus (port 9090)
- Grafana (port 3001)

### 4. Check Status

```bash
# Check all services are running
docker-compose ps

# View logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend
docker-compose logs -f frontend
```

### 5. Access Services

Once all services are running:

- **Frontend UI**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## Development Workflow

### Backend Development (Go)

For backend development with hot reload:

**Option 1: Run backend locally (recommended for development)**
```bash
# Start only database and Redis
docker-compose up -d db redis

# Run backend locally
cd backend
go mod download
go run cmd/server/main.go
```

**Option 2: Use Docker Compose**
```bash
# Backend runs in Docker with source code mounted
docker-compose up -d backend

# View logs
docker-compose logs -f backend

# Restart after changes
docker-compose restart backend
```

**Run tests:**
```bash
cd backend
go test ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/core/...
```

### Frontend Development (React)

For frontend development with hot reload:

**Option 1: Run frontend locally (recommended for development)**
```bash
cd clients/desktop
bun install
bun run dev
```

The frontend will be available at http://localhost:5173 with hot reload enabled.

**Option 2: Use Docker Compose**
```bash
# Frontend runs in Docker with hot reload
docker-compose up -d frontend

# View logs
docker-compose logs -f frontend
```

**Run tests:**
```bash
cd clients/desktop

# Type checking
bun run typecheck

# Run tests
bun run test

# Lint
bun run lint
```

## Database Management

### Running Migrations

Migrations are automatically run when the backend starts. To run manually:

```bash
# Using Docker Compose
docker-compose exec backend ./server migrate up

# Or run locally
cd backend
go run cmd/server/main.go migrate up
```

### Accessing PostgreSQL

**Using Docker:**
```bash
docker-compose exec db psql -U postgres -d trading_engine
```

**Using local client:**
```bash
psql -h localhost -U postgres -d trading_engine
```

**Common queries:**
```sql
-- List all tables
\dt

-- View accounts
SELECT * FROM accounts;

-- View positions
SELECT * FROM positions WHERE status = 'open';

-- View recent trades
SELECT * FROM trades ORDER BY executed_at DESC LIMIT 10;

-- View audit log
SELECT * FROM audit_log ORDER BY changed_at DESC LIMIT 20;
```

### Resetting Database

To start with a fresh database:

```bash
# Stop and remove volumes
docker-compose down -v

# Start again (migrations run automatically)
docker-compose up -d
```

**Warning:** This will delete all data!

### Manual Backup/Restore

**Backup:**
```bash
docker-compose exec backend bash /app/scripts/backup-db.sh
```

**Restore:**
```bash
# Restore from backup file
gunzip < /backups/postgres/trading_engine_20260116_120000.sql.gz | \
  docker-compose exec -T db psql -U postgres -d trading_engine
```

## Redis Management

### Accessing Redis

```bash
docker-compose exec redis redis-cli
```

**Common commands:**
```bash
# Test connection
PING

# View all keys
KEYS *

# Get cached value
GET tick:BTCUSD

# View OHLC cache
GET ohlc:BTCUSD:H1

# Clear all cache
FLUSHALL

# Monitor commands in real-time
MONITOR
```

### Cache Inspection

The platform caches market data in Redis:

- **Ticks**: Key pattern `tick:SYMBOL`, TTL: 60s
- **OHLC**: Key pattern `ohlc:SYMBOL:TIMEFRAME`, TTL varies (1hr-7d)

## Monitoring and Metrics

### Prometheus

Access Prometheus at http://localhost:9090

**Useful queries:**
```promql
# Order processing rate
rate(trading_orders_processed_total[5m])

# Active positions by symbol
trading_positions_active{symbol="BTCUSD"}

# Order latency P95
histogram_quantile(0.95, trading_order_duration_seconds_bucket)

# Backend health
up{job="trading-engine-backend"}
```

### Grafana

Access Grafana at http://localhost:3001 (admin/admin)

**Setup data source:**
1. Configuration → Data Sources → Add data source
2. Select Prometheus
3. URL: http://prometheus:9090
4. Save & Test

**Import dashboards:**
Dashboards can be provisioned in `monitoring/grafana/dashboards/`

### Log Viewing

**All services:**
```bash
docker-compose logs -f
```

**Structured JSON logs (backend):**
```bash
# View as JSON
docker-compose logs backend | jq .

# Filter by level
docker-compose logs backend | jq 'select(.level=="ERROR")'

# Search for specific errors
docker-compose logs backend | jq 'select(.msg | contains("database"))'

# Follow live logs
docker-compose logs -f backend | jq .
```

## Troubleshooting

### Port Conflicts

**Error:** "Bind for 0.0.0.0:8080 failed: port is already allocated"

**Solution:**
```bash
# Check what's using the port
lsof -i :8080

# Kill the process or modify docker-compose.yml
# Change port mapping: "8081:8080" instead of "8080:8080"
```

### Database Connection Errors

**Error:** "database connection refused" or "role does not exist"

**Solution:**
```bash
# Wait for database to be ready (takes 5-10 seconds after start)
docker-compose logs db

# Check database is healthy
docker-compose exec db pg_isready

# Verify connection string in .env
cat .env | grep DATABASE_URL

# Restart backend
docker-compose restart backend
```

### Frontend Proxy Errors

**Error:** "Failed to fetch" or "CORS error" when calling API

**Solution:**
```bash
# Ensure backend is running
curl http://localhost:8080/health

# Check backend logs for errors
docker-compose logs backend

# Verify VITE_API_URL environment variable
# In docker-compose.yml, frontend service should have:
# VITE_API_URL=http://localhost:8080
```

### Redis Connection Issues

**Error:** "redis: connection refused"

**Solution:**
```bash
# Check Redis is running
docker-compose ps redis

# Test Redis connectivity
docker-compose exec redis redis-cli PING

# Verify Redis URL in .env
cat .env | grep REDIS_URL

# Restart Redis
docker-compose restart redis
```

### Backend Build Errors

**Error:** "go: module not found" or build failures

**Solution:**
```bash
# Clean Go cache
go clean -modcache

# Download dependencies
cd backend
go mod download
go mod tidy

# Rebuild
docker-compose build backend
docker-compose up -d backend
```

### Frontend Build Errors

**Error:** "Cannot find module" or TypeScript errors

**Solution:**
```bash
# Clean and reinstall dependencies
cd clients/desktop
rm -rf node_modules bun.lockb
bun install

# Check for type errors
bun run typecheck

# Rebuild
docker-compose build frontend
docker-compose up -d frontend
```

### Container Crashes on Startup

**Solution:**
```bash
# View crash logs
docker-compose logs backend

# Check for missing environment variables
docker-compose config

# Verify all required services are running
docker-compose ps

# Restart with fresh state
docker-compose down
docker-compose up -d
```

### Database Migration Failures

**Error:** Migration errors or schema mismatches

**Solution:**
```bash
# Check migration status
docker-compose exec backend ./server migrate status

# Force re-run migrations
docker-compose down -v
docker-compose up -d

# Or manually apply migrations
docker-compose exec db psql -U postgres -d trading_engine < backend/db/migrations/000001_init.up.sql
```

## Performance Tips

### Faster Rebuilds

**Use Docker BuildKit:**
```bash
export DOCKER_BUILDKIT=1
docker-compose build
```

**Cache dependencies:**
Dependencies are cached in Docker layers. Only reinstall when go.mod or package.json changes.

### Resource Allocation

**Docker Desktop settings:**
- Increase memory allocation to 4-8GB
- Increase CPUs to 4+
- Enable VirtioFS file sharing (macOS) for faster volume mounts

### Development Mode Optimization

**Backend:**
```bash
# Run locally instead of Docker for faster iteration
docker-compose up -d db redis  # Only dependencies
cd backend
go run cmd/server/main.go  # Direct execution
```

**Frontend:**
```bash
# Run locally for instant hot reload
cd clients/desktop
bun run dev  # Vite dev server
```

## Environment Variables Reference

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://postgres:password@localhost:5432/trading_engine` |
| `REDIS_URL` | Redis connection string | `localhost:6379` |
| `JWT_SECRET` | JWT signing secret | `change-me-in-production` |
| `OANDA_API_KEY` | OANDA API key | `your-api-key` |
| `OANDA_ACCOUNT_ID` | OANDA account ID | `your-account-id` |

### Optional Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `LOG_LEVEL` | Logging level | `info` | `debug`, `warn`, `error` |
| `PORT` | Backend HTTP port | `8080` | `8080` |
| `VITE_API_URL` | Frontend API URL | `http://localhost:8080` | `http://localhost:8080` |

## Next Steps

Once your development environment is running:

1. **Explore the API**: Visit http://localhost:8080/health
2. **Open the UI**: Navigate to http://localhost:5173
3. **Check metrics**: View Prometheus at http://localhost:9090
4. **Set up monitoring**: Configure Grafana dashboards at http://localhost:3001
5. **Read the code**: Start with `backend/cmd/server/main.go` and `clients/desktop/src/App.tsx`

## Related Documentation

- [Docker Deployment](./DOCKER.md)
- [CI/CD Pipeline](./CI_CD.md)
- [Monitoring Guide](./MONITORING.md)
- [Operations Runbook](./OPERATIONS.md)
