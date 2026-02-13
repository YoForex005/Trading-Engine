# Login & Terminal Troubleshooting Guide

**RTX Trading Engine - Comprehensive Login Documentation**

---

## Table of Contents

1. [Complete Login Flow](#complete-login-flow)
2. [Step-by-Step Testing Instructions](#step-by-step-testing-instructions)
3. [Common Issues and Solutions](#common-issues-and-solutions)
4. [Environment Setup Requirements](#environment-setup-requirements)
5. [Database Setup Instructions](#database-setup-instructions)
6. [Test Credentials Documentation](#test-credentials-documentation)
7. [API Endpoint Documentation](#api-endpoint-documentation)
8. [Frontend Integration Guide](#frontend-integration-guide)
9. [Backend Architecture](#backend-architecture)
10. [Debugging Checklist](#debugging-checklist)

---

## Complete Login Flow

### Overview

The RTX Trading Engine implements a **dual authentication system**:

1. **Client Login** (`/login`) - For trader accounts accessing the desktop terminal
2. **Admin Login** (`/admin/auth/login`) - For administrative panel access

### Client Login Flow (Detailed)

```
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 1: User Input                                                  │
│ ┌──────────────────────────────────────────────────────────┐        │
│ │ Account ID: 1                                            │        │
│ │ Password: password                                       │        │
│ │ Server: localhost:7999                                   │        │
│ └──────────────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 2: Frontend Validation (Login.tsx)                            │
│ • Validate non-empty username & password                           │
│ • Create AbortController for 10-second timeout                     │
│ • Construct request body                                           │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 3: HTTP Request                                               │
│ POST http://localhost:7999/login                                   │
│ Content-Type: application/json                                     │
│                                                                     │
│ Body:                                                               │
│ {                                                                   │
│   "username": "1",                                                  │
│   "password": "password"                                            │
│ }                                                                   │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 4: Backend Receipt (api/server.go:112)                        │
│ • HandleLogin receives request                                      │
│ • Parse JSON body                                                   │
│ • Call authService.Login(username, password)                        │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 5: Authentication Logic (auth/service.go:56)                  │
│                                                                     │
│ IF username == "admin":                                             │
│   → Verify against admin bcrypt hash                               │
│   → Return admin user & token                                      │
│ ELSE:                                                               │
│   → Try to parse username as numeric account ID                    │
│   → Lookup account in B-Book engine                                │
│   → If not found, try as string username                           │
│   → Verify password with bcrypt.CompareHashAndPassword             │
│   → Auto-upgrade plaintext passwords to bcrypt                     │
│   → Generate JWT token with user claims                            │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 6: Token Generation (auth/jwt.go)                             │
│ • Create JWT with claims:                                          │
│   - UserID: "1"                                                     │
│   - Username: "demo-user"                                           │
│   - Role: "TRADER"                                                  │
│   - ExpiresAt: now + 24 hours                                       │
│ • Sign with JWT_SECRET from .env                                   │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 7: Backend Response                                           │
│ HTTP 200 OK                                                         │
│ Content-Type: application/json                                     │
│                                                                     │
│ {                                                                   │
│   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",              │
│   "user": {                                                         │
│     "id": "1",                                                      │
│     "username": "demo-user",                                        │
│     "role": "TRADER"                                                │
│   }                                                                 │
│ }                                                                   │
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 8: Frontend Token Storage (Login.tsx:84-87)                   │
│ • Store in Zustand store: setAuthToken(data.token)                 │
│ • Store in localStorage: localStorage.setItem('rtx_token', token)  │
│ • Store user data: localStorage.setItem('rtx_user', user)          │
│ • Set authenticated state: setAuthenticated(true, accountId, token)│
└─────────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 9: Navigate to Terminal                                       │
│ • Call onLogin(accountId)                                           │
│ • App.tsx switches from Login to Terminal view                     │
│ • WebSocket connection established with token                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Admin Login Flow (Simplified)

```
POST /admin/auth/login
  → admin/handlers.go HandleLogin (line 127)
  → admin.AuthService.Login (admin/auth.go line 115)
  → Verify "admin" / "Admin@123" against bcrypt hash
  → Create AdminSession (8-hour expiry)
  → Return { sessionId, admin }
```

---

## Step-by-Step Testing Instructions

### Prerequisites

- Node.js v16+ installed
- Go 1.20+ installed (or use pre-built server.exe)
- Ports 7999, 5173 available
- Terminal/Command Prompt with admin privileges

### Test Procedure 1: Full System Start

#### 1. Start All Services

```bash
# From project root
START.bat
```

This will:
- Kill any existing processes on ports 7999, 5173, 3000, 3001
- Build backend from source (or use pre-built server.exe)
- Install npm dependencies if missing
- Start backend on port 7999
- Start desktop client on port 5173
- Start broker admin on port 3000
- Start super admin on port 3001

#### 2. Wait for Services to Start

```
[OK] Backend API:     http://localhost:7999
[OK] Desktop Client:  http://localhost:5173
[OK] Broker Admin:    http://localhost:3000
[OK] Super Admin:     http://localhost:3001
```

**Wait ~10-15 seconds** for all services to initialize.

#### 3. Verify Backend Health

Open browser or use curl:

```bash
curl http://localhost:7999/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2026-02-13T10:30:00Z"
}
```

#### 4. Open Desktop Client

Navigate to: http://localhost:5173

You should see the RTX Terminal login screen.

#### 5. Enter Login Credentials

- **Account ID:** `1`
- **Password:** `password`
- **Server:** `localhost:7999` (should be pre-filled)

#### 6. Click "Connect to Gate"

**Expected Behavior:**
- Loading state: "Connecting..."
- JWT token received within 1-2 seconds
- Redirect to trading terminal
- WebSocket connection established
- Market data starts streaming

**If Successful:**
- You'll see the main trading interface
- Account balance visible
- Symbol list populated
- Price quotes updating

### Test Procedure 2: Manual CLI Test

#### Test Backend Login Endpoint

```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"1","password":"password"}'
```

**Expected Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInVzZXJuYW1lIjoiZGVtby11c2VyIiwicm9sZSI6IlRSQURFUiIsImV4cCI6MTczOTQ1NDYwMH0.abc123...",
  "user": {
    "id": "1",
    "username": "demo-user",
    "role": "TRADER"
  }
}
```

#### Test with Invalid Credentials

```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"1","password":"wrongpassword"}'
```

**Expected Response:**
```
HTTP 401 Unauthorized
Unauthorized
```

#### Test Admin Login

```bash
curl -X POST http://localhost:7999/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'
```

**Expected Response:**
```json
{
  "sessionId": "base64-encoded-session-token",
  "admin": {
    "id": 1,
    "username": "admin",
    "email": "admin@rtx.local",
    "role": "SUPER_ADMIN"
  }
}
```

### Test Procedure 3: Frontend Only (If Backend Already Running)

#### 1. Navigate to desktop client folder

```bash
cd clients/desktop
```

#### 2. Install dependencies (if needed)

```bash
npm install
```

#### 3. Start Vite dev server

```bash
npm run dev
```

or

```bash
npx vite --host
```

#### 4. Open browser

Navigate to: http://localhost:5173

#### 5. Test login with default credentials

---

## Common Issues and Solutions

### Issue 1: "Request timeout - server not responding"

**Symptom:** Login button shows "Connecting..." then error message appears.

**Causes:**
- Backend not running
- Backend crashed during startup
- Port 7999 blocked by firewall
- Wrong server URL

**Solutions:**

1. **Check if backend is running:**
   ```bash
   curl http://localhost:7999/health
   ```

2. **Check backend window** - Look for errors in the console

3. **Verify port 7999 is listening:**
   ```bash
   netstat -an | findstr 7999
   ```

4. **Restart backend:**
   - Close backend window
   - Run: `cd backend && server.exe`

5. **Check firewall settings** - Allow port 7999

### Issue 2: "Invalid credentials"

**Symptom:** Error message "Invalid credentials" despite correct password.

**Causes:**
- Wrong account ID
- Account doesn't exist
- Password mismatch
- Account disabled

**Solutions:**

1. **Verify default account was created** - Check backend logs for:
   ```
   [B-Book] Demo account created: demo_001 with $10000.00
   ```

2. **Check account ID** - Should be `1` for the demo account

3. **Try admin login** instead:
   - Username: `admin`
   - Password: `Admin@123`

4. **Check backend/cmd/server/main.go** line 153 - Ensure demo account creation is enabled

### Issue 3: CORS Error in Browser Console

**Symptom:**
```
Access to fetch at 'http://localhost:7999/login' from origin 'http://localhost:5173'
has been blocked by CORS policy
```

**Causes:**
- CORS middleware not configured
- Frontend not in ALLOWED_ORIGINS list
- Backend not returning proper CORS headers

**Solutions:**

1. **Check backend/.env** - Verify ALLOWED_ORIGINS includes frontend URL:
   ```env
   ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:3001
   ```

2. **Check server.go HandleLogin** (line 113) - Ensure CORS headers are set:
   ```go
   w.Header().Set("Access-Control-Allow-Origin", "*")
   w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
   ```

3. **Restart backend** after changing .env

### Issue 4: "No authentication token received"

**Symptom:** Login succeeds (HTTP 200) but no token in response.

**Causes:**
- Backend response parsing error
- Invalid JSON response
- JWT generation failed

**Solutions:**

1. **Check backend logs** for JWT generation errors

2. **Verify JWT_SECRET** in backend/.env:
   ```env
   JWT_SECRET=your-secure-secret-here-min-32-bytes
   ```

3. **Test login with curl** to see raw response:
   ```bash
   curl -v -X POST http://localhost:7999/login \
     -H "Content-Type: application/json" \
     -d '{"username":"1","password":"password"}'
   ```

4. **Check auth/jwt.go** - Ensure GenerateJWTWithSecret is working

### Issue 5: WebSocket Connection Fails After Login

**Symptom:** Login successful but terminal shows "Disconnected" or no market data.

**Causes:**
- WebSocket not using auth token
- WebSocket URL incorrect
- Backend WebSocket hub not started

**Solutions:**

1. **Check WebSocket URL** in frontend:
   ```typescript
   const WS_URL = 'ws://localhost:7999/ws'
   ```

2. **Verify token is passed** in WebSocket connection:
   ```typescript
   ws = new WebSocket(`${WS_URL}?token=${authToken}`)
   ```

3. **Check backend logs** for WebSocket hub startup:
   ```
   [WebSocket] Hub started
   ```

4. **Verify backend/cmd/server/main.go line 229** - Hub.Run() is called:
   ```go
   go hub.Run()
   ```

### Issue 6: Backend Crashes on Startup

**Symptom:** Backend window closes immediately or shows panic/error.

**Common Errors:**

#### Error: "panic: runtime error: index out of range"

**Solution:** Update backend to commit `e44034f` which fixes clientNames panics in mock data.

#### Error: "panic: http: multiple registrations for /admin/clients/"

**Solution:** Commit `e44034f` fixes duplicate route registration. Update backend.

#### Error: "Failed to load .env file"

**Solution:** Ensure backend/.env exists. Copy from backend/.env.example if needed.

#### Error: "Database connection failed"

**Solution:**
- Check PostgreSQL is running (if using DB)
- Verify DB credentials in .env
- Or use in-memory mode (default)

### Issue 7: Frontend TypeScript Errors

**Symptom:** "Cannot find module" or type import errors.

**Solutions:**

1. **Update tsconfig.app.json** to enable verbatimModuleSyntax fix (commit e44034f)

2. **Fix type-only imports:**
   ```typescript
   // Before (ERROR)
   import { SetupTag } from './types'

   // After (FIXED)
   import type { SetupTag } from './types'
   ```

3. **Run type check:**
   ```bash
   npm run typecheck
   ```

### Issue 8: Port Already in Use

**Symptom:**
```
Error: listen EADDRINUSE: address already in use :::7999
```

**Solutions:**

1. **Kill existing process:**
   ```bash
   # Windows
   netstat -ano | findstr :7999
   taskkill /PID <PID> /F

   # Linux/Mac
   lsof -ti:7999 | xargs kill -9
   ```

2. **Use START.bat** which automatically cleans up ports

3. **Change port** in backend/.env if needed:
   ```env
   PORT=8000
   ```

---

## Environment Setup Requirements

### Required Software

| Software | Version | Required | Notes |
|----------|---------|----------|-------|
| Node.js | 16+ | ✅ Yes | For frontend and admin panels |
| npm | 9+ | ✅ Yes | Comes with Node.js |
| Go | 1.20+ | ⚠️ Optional | Can use pre-built server.exe |
| Git | Any | ⚠️ Optional | For version control |
| PostgreSQL | 13+ | ❌ No | Optional, uses in-memory by default |
| Redis | 6+ | ❌ No | Optional caching layer |

### System Requirements

- **OS:** Windows 10/11, Linux, macOS
- **RAM:** 4GB minimum, 8GB recommended
- **Disk:** 2GB free space
- **Ports:** 7999, 5173, 3000, 3001 available

### Environment Variables

#### Backend (.env)

Located at: `backend/.env`

```env
# Server
PORT=7999
ENVIRONMENT=development

# Database (Optional - uses in-memory if not configured)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
DB_USER=trading
DB_PASSWORD=trading_pass
DB_SSL_MODE=disable

# Authentication
JWT_SECRET=your-secure-secret-here-min-32-bytes
JWT_EXPIRY=24h

# Admin
ADMIN_EMAIL=admin@example.com
# Leave ADMIN_PASSWORD_HASH empty to use default "Admin@123"

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000,http://localhost:3001,http://localhost:3002

# Broker Settings
BROKER_NAME=RTX Trading
DEFAULT_BALANCE=10000.0
DEFAULT_LEVERAGE=100
```

#### Frontend (.env) - Optional

Located at: `clients/desktop/.env`

```env
VITE_API_URL=http://localhost:7999
VITE_WS_URL=ws://localhost:7999/ws
VITE_MARKET_WS_URL=ws://localhost:7999/market-data
```

---

## Database Setup Instructions

### Option 1: In-Memory Mode (Default - No Setup Required)

The backend uses in-memory B-Book engine by default:

- ✅ No database installation needed
- ✅ Demo account auto-created on startup
- ✅ Perfect for development/testing
- ⚠️ Data lost on restart

**Demo Account Details:**
- **Account ID:** 1
- **Username:** demo-user
- **Password:** password (auto-upgraded to bcrypt on first login)
- **Balance:** $10,000 (configurable in .env)

### Option 2: PostgreSQL (Production)

#### 1. Install PostgreSQL

```bash
# Windows: Download from postgresql.org
# Linux: apt install postgresql postgresql-contrib
# Mac: brew install postgresql
```

#### 2. Create Database

```sql
CREATE DATABASE trading_engine;
CREATE USER trading WITH PASSWORD 'trading_pass';
GRANT ALL PRIVILEGES ON DATABASE trading_engine TO trading;
```

#### 3. Run Migrations

```bash
cd backend
go run cmd/migrate/main.go
```

#### 4. Update .env

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=trading_engine
DB_USER=trading
DB_PASSWORD=trading_pass
DB_SSL_MODE=disable
```

#### 5. Restart Backend

The system will now persist accounts, orders, and positions to PostgreSQL.

---

## Test Credentials Documentation

### Default Credentials

| Type | Username/Account | Password | Role | Notes |
|------|------------------|----------|------|-------|
| **Client (Desktop)** | `1` | `password` | TRADER | Demo account, $10k balance |
| **Admin Panel** | `admin` | `Admin@123` | SUPER_ADMIN | Full admin access |

### Password Details

#### Client Password
- **Default:** Plaintext `"password"`
- **Auto-upgraded:** On first login to bcrypt hash
- **Stored in:** B-Book engine in-memory or PostgreSQL users table
- **Change:** Use auth.Service.UpdatePassword() or admin panel

#### Admin Password
- **Set in:** `backend/admin/auth.go` line 59
- **Default:** `Admin@123` (for development)
- **Production:** Set ADMIN_PASSWORD_HASH in .env with bcrypt hash
- **Generate hash:**
  ```bash
  # Using bcrypt CLI tool
  bcrypt hash "YourSecurePassword123!"
  ```

### Creating Additional Accounts

#### Via Code (Demo Accounts)

Edit `backend/cmd/server/main.go` around line 152:

```go
// Create additional demo account
demoAccount2 := bbookEngine.CreateAccount("demo-user-2", "Demo User 2", "password", true)
bbookEngine.GetLedger().SetBalance(demoAccount2.ID, 5000.0)
log.Printf("[B-Book] Demo account created: %s with $%.2f", demoAccount2.AccountNumber, 5000.0)
```

#### Via Admin Panel (Future)

- Navigate to http://localhost:3000
- Login as admin
- Go to "Users" → "Create Account"
- Fill in details and set password

---

## API Endpoint Documentation

### Authentication Endpoints

#### POST /login (Client Login)

**Request:**
```http
POST /login HTTP/1.1
Host: localhost:7999
Content-Type: application/json

{
  "username": "1",
  "password": "password"
}
```

**Response (Success):**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "1",
    "username": "demo-user",
    "role": "TRADER"
  }
}
```

**Response (Error):**
```http
HTTP/1.1 401 Unauthorized

Unauthorized
```

**Implementation:** `backend/api/server.go:112` → `backend/auth/service.go:56`

#### POST /admin/auth/login (Admin Login)

**Request:**
```http
POST /admin/auth/login HTTP/1.1
Host: localhost:7999
Content-Type: application/json

{
  "username": "admin",
  "password": "Admin@123"
}
```

**Response (Success):**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "sessionId": "k8j2h3g4f5d6s7a8",
  "admin": {
    "id": 1,
    "username": "admin",
    "email": "admin@rtx.local",
    "role": "SUPER_ADMIN",
    "status": "ACTIVE"
  }
}
```

**Implementation:** `backend/admin/handlers.go:127` → `backend/admin/auth.go:115`

### Protected Endpoints (Require Authentication)

All subsequent API calls must include the JWT token:

```http
GET /api/accounts HTTP/1.1
Host: localhost:7999
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Health Check

#### GET /health

**Request:**
```http
GET /health HTTP/1.1
Host: localhost:7999
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "healthy",
  "timestamp": "2026-02-13T10:30:00Z"
}
```

---

## Frontend Integration Guide

### Using the Login Component

#### Import

```typescript
import { Login } from './components/Login';
```

#### Usage

```typescript
function App() {
  const [accountId, setAccountId] = useState<string | null>(null);

  const handleLogin = (id: string) => {
    setAccountId(id);
    // Navigate to terminal or dashboard
  };

  if (!accountId) {
    return <Login onLogin={handleLogin} />;
  }

  return <TradingTerminal accountId={accountId} />;
}
```

### Accessing Auth State (Zustand)

```typescript
import { useAppStore } from './store/useAppStore';

function MyComponent() {
  const { authenticated, authToken } = useAppStore();

  if (!authenticated) {
    return <div>Please login</div>;
  }

  // Use authToken for API calls
  const fetchData = async () => {
    const response = await fetch('http://localhost:7999/api/accounts', {
      headers: {
        'Authorization': `Bearer ${authToken}`
      }
    });
    return response.json();
  };
}
```

### Making Authenticated API Calls

```typescript
import { API_ENDPOINTS } from './config/api';

async function placeOrder(orderData) {
  const token = localStorage.getItem('rtx_token');

  const response = await fetch(API_ENDPOINTS.order, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(orderData)
  });

  if (response.status === 401) {
    // Token expired - redirect to login
    window.location.href = '/login';
    return;
  }

  return response.json();
}
```

### WebSocket Authentication

```typescript
import { WS_URL } from './config/api';

function connectWebSocket() {
  const token = localStorage.getItem('rtx_token');

  const ws = new WebSocket(`${WS_URL}?token=${token}`);

  ws.onopen = () => {
    console.log('WebSocket connected');
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
    // Handle authentication error
  };

  return ws;
}
```

---

## Backend Architecture

### Authentication Services

#### 1. auth.Service (Client Authentication)

**File:** `backend/auth/service.go`

**Purpose:** Authenticate trader accounts accessing desktop terminal

**Key Methods:**
- `Login(username, password)` - Validate credentials, return JWT
- `GenerateToken(user)` - Create JWT with claims
- `ValidateToken(tokenString)` - Verify JWT signature
- `ValidateAdminToken(r *http.Request)` - Extract & validate from header

**Flow:**
```
Login Request
  → Parse username (numeric ID or string)
  → Lookup account in B-Book engine
  → Verify password (bcrypt or plaintext with auto-upgrade)
  → Generate JWT token (24h expiry)
  → Return token + user
```

#### 2. admin.AuthService (Admin Authentication)

**File:** `backend/admin/auth.go`

**Purpose:** Authenticate admin users for broker admin panel

**Key Methods:**
- `Login(username, password, ipAddress, userAgent)` - Create admin session
- `ValidateSession(sessionID, ipAddress)` - Verify session token
- `Logout(sessionID)` - Terminate session
- `CreateAdmin(...)` - Create new admin account

**Flow:**
```
Admin Login Request
  → Verify against adminsByUsername map
  → Check bcrypt password hash
  → Verify account status (ACTIVE)
  → Check IP whitelist (if configured)
  → Handle 2FA if enabled (TOTP)
  → Generate session token (8h expiry)
  → Return sessionID + admin
```

### Default Accounts

#### Demo Trading Account

**Created at:** `backend/cmd/server/main.go:152-156`

```go
if brokerConfig.DefaultBalance > 0 {
    demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
    bbookEngine.GetLedger().SetBalance(demoAccount.ID, brokerConfig.DefaultBalance)
    demoAccount.Balance = brokerConfig.DefaultBalance
    log.Printf("[B-Book] Demo account created: %s with $%.2f",
        demoAccount.AccountNumber, brokerConfig.DefaultBalance)
}
```

**Properties:**
- ID: 1 (auto-incremented)
- Username: "demo-user"
- Display Name: "Demo User"
- Password: "password" (plaintext, auto-upgraded to bcrypt on first login)
- Balance: $10,000 (from DEFAULT_BALANCE in .env)
- IsDemo: true

#### Super Admin Account

**Created at:** `backend/admin/auth.go:58-64`

```go
superAdmin, err := svc.CreateAdmin("admin", "admin@rtx.local", "Admin@123",
    RoleSuperAdmin, nil, "SYSTEM")
if err != nil {
    log.Printf("[AdminAuth] Failed to create default super admin: %v", err)
} else {
    log.Printf("[AdminAuth] Default super admin created: %s", superAdmin.Username)
}
```

**Properties:**
- ID: 1
- Username: "admin"
- Email: "admin@rtx.local"
- Password: "Admin@123" (bcrypt hashed)
- Role: SUPER_ADMIN
- Status: ACTIVE
- Created By: SYSTEM

### JWT Token Structure

**Generated by:** `backend/auth/jwt.go`

**Claims:**
```go
type Claims struct {
    UserID   string `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

**Example Token:**
```
Header: {"alg":"HS256","typ":"JWT"}
Payload: {
  "user_id": "1",
  "username": "demo-user",
  "role": "TRADER",
  "exp": 1739454600,  // 24 hours from now
  "iat": 1739368200
}
Signature: HMACSHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), secret)
```

**Verification:**
- Secret: JWT_SECRET from .env
- Algorithm: HS256
- Expiry: Checked on every validation

---

## Debugging Checklist

### Pre-Flight Checks

- [ ] Node.js installed (`node --version`)
- [ ] npm installed (`npm --version`)
- [ ] Go installed or server.exe exists (`go version` or `ls backend/server.exe`)
- [ ] Ports 7999, 5173 available (`netstat -an | findstr "7999 5173"`)
- [ ] backend/.env file exists
- [ ] ALLOWED_ORIGINS includes http://localhost:5173
- [ ] JWT_SECRET is set (at least 32 characters)

### Backend Health Checks

- [ ] Backend starts without errors
- [ ] Console shows: `[B-Book] Demo account created: demo_001 with $10000.00`
- [ ] Console shows: `[AdminAuth] Default super admin created: admin`
- [ ] Console shows: `[WebSocket] Hub started`
- [ ] Health endpoint responds: `curl http://localhost:7999/health`
- [ ] No panic/crash errors in backend window

### Frontend Health Checks

- [ ] Frontend dev server starts (`npm run dev` in clients/desktop)
- [ ] No TypeScript compilation errors
- [ ] Browser console clean (no CORS errors)
- [ ] Login page renders correctly
- [ ] Default server value is "localhost:7999"

### Login Test Checks

- [ ] Enter account ID "1" and password "password"
- [ ] Click "Connect to Gate"
- [ ] No timeout error (request completes in < 10 seconds)
- [ ] No "Invalid credentials" error
- [ ] Token received and stored in localStorage
- [ ] Authenticated state set in Zustand
- [ ] Redirect to terminal occurs
- [ ] WebSocket connection established

### Post-Login Checks

- [ ] Account balance displays ($10,000)
- [ ] Symbol list populated
- [ ] Market data streaming (prices updating)
- [ ] Can place test orders
- [ ] WebSocket stays connected
- [ ] No console errors

### Troubleshooting Commands

```bash
# Check backend logs
# (Look in backend window)

# Check if backend is running
curl http://localhost:7999/health

# Test login endpoint
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d '{"username":"1","password":"password"}'

# Check ports
netstat -an | findstr "7999 5173"

# Kill processes on port
# Windows:
netstat -ano | findstr :7999
taskkill /PID <PID> /F

# Restart backend
cd backend
server.exe

# Restart frontend
cd clients/desktop
npm run dev

# Clear browser cache
# Chrome: Ctrl+Shift+Delete → Clear all
# Check localStorage in DevTools: Application → Local Storage

# Check WebSocket
# Browser DevTools → Network → WS tab
```

---

## Conclusion

This guide covers the complete login system architecture, testing procedures, common issues, and debugging steps for the RTX Trading Engine.

### Quick Reference

**Default Login:**
- Account ID: `1`
- Password: `password`
- Server: `localhost:7999`

**Admin Login:**
- Username: `admin`
- Password: `Admin@123`

**Endpoints:**
- Client Login: `POST http://localhost:7999/login`
- Admin Login: `POST http://localhost:7999/admin/auth/login`
- Health Check: `GET http://localhost:7999/health`

**Start System:**
```bash
START.bat
```

**Manual Start:**
```bash
# Terminal 1: Backend
cd backend && server.exe

# Terminal 2: Frontend
cd clients/desktop && npm run dev
```

---

**Last Updated:** February 13, 2026
**Document Version:** 1.0
**Maintained By:** RTX Development Team
