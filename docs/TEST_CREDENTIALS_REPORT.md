# Test Credentials & Database Setup Report

**Generated:** 2026-02-13
**Status:** Database Not Initialized
**PostgreSQL Service:** STOPPED

---

## Executive Summary

The Trading Engine has test credentials configured in the code, but the **PostgreSQL database is not running**. Users cannot log in until the database service is started and migrations are run.

---

## 1. Available Test Credentials

### 1.1 Admin Panel Credentials (In-Memory)

**Location:** `backend/admin/auth.go` (Line 59)

| Username | Password | Role | Status |
|----------|----------|------|--------|
| `admin` | `Admin@123` | Super Admin | Active (In-Memory) |

**Details:**
- This admin user is created in-memory when the backend starts
- Does NOT require database connection
- Used for admin panel authentication at `/admin` endpoints
- Password is hardcoded in `NewAuthService()` function

**Code Reference:**
```go
// backend/admin/auth.go:59
superAdmin, err := svc.CreateAdmin("admin", "admin@rtx.local", "Admin@123", RoleSuperAdmin, nil, "SYSTEM")
```

### 1.2 Demo Trading Account (In-Memory)

**Location:** `backend/cmd/server/main.go` (Line 153)

| Username | Password | Balance | Account Type |
|----------|----------|---------|--------------|
| `demo-user` | `password` | $10,000 (configurable) | Demo Account |

**Details:**
- Created when backend starts if `DEFAULT_ACCOUNT_BALANCE` is configured
- Stored in B-Book engine (in-memory)
- Does NOT require database
- Used for desktop client login

**Code Reference:**
```go
// backend/cmd/server/main.go:153
demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
```

### 1.3 Database Test Users (NOT CREATED YET)

**Location:** `backend/scripts/db/seed-test-data.sh`

These users will be created when you run the seed script:

| Email | Username | Password | Role | KYC Status |
|-------|----------|----------|------|------------|
| `admin@tradingengine.com` | `admin` | (needs hash) | Super Admin, Risk Manager | Approved |
| `trader1@example.com` | `trader1` | (needs hash) | Trader | Approved |
| `trader2@example.com` | `trader2` | (needs hash) | Trader | Approved |
| `demo@example.com` | `demo` | (needs hash) | User | Pending |

**Note:** The seed script has placeholder hashes `$2a$10$YourHashedPasswordHere` which need to be replaced with actual bcrypt hashes.

---

## 2. Database Status

### 2.1 PostgreSQL Service

```
Service Name: postgresql-x64-18
State: STOPPED
Path: C:\Program Files\PostgreSQL\18\bin\psql.exe
```

**Issue:** PostgreSQL is installed but not running.

### 2.2 Database Connection Configuration

**File:** `backend/.env`

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
DB_USER=trading
DB_PASSWORD=trading_pass
DB_SSL_MODE=disable
```

**Connection Test Result:**
```
❌ Connection refused (0x0000274D/10061)
Is the server running on that host and accepting TCP/IP connections?
```

### 2.3 Migration Status

**Available Migrations:**
- `001_init_schema.sql` - Creates users, accounts, orders, positions tables
- `002_add_indexes.sql`
- `003_add_audit_tables.sql`
- `004_add_lp_tables.sql`
- `005_add_risk_tables.sql`
- `006_add_payment_tables.sql`
- `007_add_config_and_activity_tables.sql`
- `007_add_routing_rules.sql`
- `008_add_compliance_reporting.sql`
- `009_tick_history_timescaledb.sql`

**Status:** None have been run yet (database not accessible)

---

## 3. Setup Instructions

### Step 1: Start PostgreSQL Service

**Option A: Using Windows Services (Requires Admin)**
```cmd
REM Run Command Prompt as Administrator
net start postgresql-x64-18
```

**Option B: Using Services GUI**
1. Press `Win + R`, type `services.msc`, press Enter
2. Find "postgresql-x64-18"
3. Right-click → Start

### Step 2: Verify PostgreSQL is Running

```cmd
psql -h localhost -p 5432 -U postgres -c "SELECT version();"
```

Expected output: PostgreSQL version information

### Step 3: Create Database and User

```cmd
psql -h localhost -p 5432 -U postgres
```

```sql
-- Create database
CREATE DATABASE trading_engine;

-- Create user
CREATE USER trading WITH PASSWORD 'trading_pass';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE trading_engine TO trading;

-- Connect to database
\c trading_engine

-- Grant schema privileges
GRANT ALL ON SCHEMA public TO trading;
```

### Step 4: Run Migrations

**Option A: Using Migration Tool**
```cmd
cd C:\Users\Yofor\Desktop\Trading-Engine\backend
go run cmd/migrate/main.go
```

**Option B: Manual Migration**
```cmd
cd C:\Users\Yofor\Desktop\Trading-Engine\backend
psql -h localhost -p 5432 -U trading -d trading_engine -f migrations/001_init_schema.sql
psql -h localhost -p 5432 -U trading -d trading_engine -f migrations/002_add_indexes.sql
REM ... run remaining migrations in order
```

### Step 5: Seed Test Data (Optional)

**Generate bcrypt hashes for passwords:**
```bash
# Install bcrypt tool if needed
go install github.com/bitnami/bcrypt-cli@latest

# Generate hash for "password123"
bcrypt-cli "password123"
```

**Update and run seed script:**
```bash
cd C:\Users\Yofor\Desktop\Trading-Engine\backend\scripts\db
# Edit seed-test-data.sh to replace $2a$10$YourHashedPasswordHere with actual hashes
bash seed-test-data.sh
```

---

## 4. Login Testing

### 4.1 Admin Panel Login (Works WITHOUT Database)

**URL:** `http://localhost:7999/admin/login`

```json
POST /admin/login
{
  "username": "admin",
  "password": "Admin@123",
  "ip_address": "127.0.0.1",
  "user_agent": "Mozilla/5.0"
}
```

**Expected Response:**
```json
{
  "session_id": "<token>",
  "admin": {
    "username": "admin",
    "role": "SUPER_ADMIN"
  }
}
```

### 4.2 Desktop Client Login (Works WITHOUT Database)

**URL:** `http://localhost:7999/login`

```json
POST /login
{
  "username": "demo-user",
  "password": "password"
}
```

**Expected Response:**
```json
{
  "token": "<jwt_token>",
  "user": {
    "id": "1",
    "username": "demo-user",
    "role": "TRADER"
  }
}
```

### 4.3 Database User Login (Requires Database)

After running migrations and seed script:

```json
POST /login
{
  "username": "trader1@example.com",
  "password": "<password_you_set>"
}
```

---

## 5. Common Issues & Solutions

### Issue 1: "Connection refused" Error

**Symptom:**
```
psql: error: connection to server at "localhost" (::1), port 5432 failed
```

**Solution:**
- PostgreSQL service is not running
- Start service: `net start postgresql-x64-18` (requires admin)

### Issue 2: "Access is denied" when starting service

**Symptom:**
```
System error 5 has occurred.
Access is denied.
```

**Solution:**
- Open Command Prompt as Administrator
- Or use Services GUI (services.msc)

### Issue 3: Admin login works but trader login fails

**Symptom:**
- Admin panel login succeeds
- Desktop client login with demo-user works
- Database users don't work

**Cause:** Database is not running or migrations not applied

**Solution:**
1. Start PostgreSQL service
2. Run migrations
3. Seed test data

### Issue 4: Invalid password hash in seed script

**Symptom:**
```
INSERT INTO users ... VALUES (..., '$2a$10$YourHashedPasswordHere', ...)
```

**Solution:**
Generate real bcrypt hashes:
```bash
# Using Go
go run -c 'import "golang.org/x/crypto/bcrypt"; hash, _ := bcrypt.GenerateFromPassword([]byte("yourpassword"), bcrypt.DefaultCost); fmt.Println(string(hash))'

# Using online tool
# Visit: https://bcrypt-generator.com/
# Enter password, get hash
```

---

## 6. Security Recommendations

### Development Environment

✅ **Current Status:**
- Admin password is hardcoded (acceptable for dev)
- Demo account uses simple password (acceptable for dev)
- JWT secret can be set via `.env` (good)

### Production Environment

⚠️ **Required Changes:**

1. **Remove hardcoded credentials**
   - Set admin password via environment variable
   - Use strong, unique passwords

2. **Use environment-specific configuration**
   ```env
   # Production .env
   ADMIN_PASSWORD_HASH=$2a$12$<strong_hash>
   JWT_SECRET=<64_character_random_string>
   ```

3. **Database security**
   - Use strong database passwords
   - Enable SSL mode (`DB_SSL_MODE=require`)
   - Restrict database user permissions

4. **Password policy**
   - Minimum 12 characters
   - Require complexity (uppercase, lowercase, numbers, symbols)
   - Implement password expiration
   - Enable 2FA for admin accounts

---

## 7. Next Steps

### Immediate (To Get System Running)

1. ☐ Start PostgreSQL service (requires admin privileges)
2. ☐ Create database and user
3. ☐ Run migrations
4. ☐ Test in-memory logins (admin, demo-user)

### Short-term (For Full Functionality)

5. ☐ Generate bcrypt hashes for test users
6. ☐ Update seed script with real hashes
7. ☐ Run seed script to create database users
8. ☐ Test database user logins

### Long-term (For Production)

9. ☐ Externalize all credentials to environment variables
10. ☐ Implement password rotation policy
11. ☐ Enable 2FA for admin accounts
12. ☐ Set up database backups
13. ☐ Configure SSL/TLS for database connections

---

## 8. Credential Summary Table

| System | Username | Password | Works Without DB | Location |
|--------|----------|----------|------------------|----------|
| Admin Panel | `admin` | `Admin@123` | ✅ Yes | `backend/admin/auth.go:59` |
| Desktop Client | `demo-user` | `password` | ✅ Yes | `backend/cmd/server/main.go:153` |
| Desktop Client | `admin` | See .env | ❌ No | README.md says "admin/password" |
| Desktop Client | `trader` | See .env | ❌ No | README.md says "trader/password" |
| Database | `admin@tradingengine.com` | Not set | ❌ No | Seed script |
| Database | `trader1@example.com` | Not set | ❌ No | Seed script |
| Database | `trader2@example.com` | Not set | ❌ No | Seed script |
| Database | `demo@example.com` | Not set | ❌ No | Seed script |

---

## 9. Quick Start Guide

### For Developers (No Database Setup)

```bash
# 1. Start backend
cd C:\Users\Yofor\Desktop\Trading-Engine\backend
go run cmd/server/main.go

# 2. Test admin login
curl -X POST http://localhost:7999/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123","ip_address":"127.0.0.1"}'

# 3. Test demo login
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo-user","password":"password"}'
```

### For Full System (With Database)

1. **Start PostgreSQL** (requires admin)
   ```cmd
   REM Run as Administrator
   net start postgresql-x64-18
   ```

2. **Initialize Database**
   ```cmd
   cd C:\Users\Yofor\Desktop\Trading-Engine\backend
   go run cmd/migrate/main.go
   ```

3. **Seed Test Data**
   ```bash
   cd scripts/db
   bash seed-test-data.sh
   ```

4. **Start Backend**
   ```cmd
   cd C:\Users\Yofor\Desktop\Trading-Engine\backend
   go run cmd/server/main.go
   ```

---

## Appendix: File Locations

### Configuration Files
- Environment: `backend/.env`
- Database migrations: `backend/migrations/*.sql`
- Seed script: `backend/scripts/db/seed-test-data.sh`

### Authentication Code
- Admin auth: `backend/admin/auth.go`
- User auth: `backend/auth/service.go`
- Middleware: `backend/internal/middleware/auth.go`

### Main Entry Point
- Server: `backend/cmd/server/main.go`
- Migration tool: `backend/cmd/migrate/main.go`

---

**Report End**
