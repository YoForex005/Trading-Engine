# Quick Login Guide - Trading Engine

## TL;DR - Start Coding Now

**Working credentials (NO database required):**

### Admin Panel
- URL: `http://localhost:7999/admin/login`
- Username: `admin`
- Password: `Admin@123`

### Desktop Client
- URL: `http://localhost:5173` (after starting client)
- Username: `demo-user`
- Password: `password`

---

## Start Backend

```bash
cd C:\Users\Yofor\Desktop\Trading-Engine\backend
go run cmd/server/main.go
```

**Expected output:**
```
╔═══════════════════════════════════════════════════════════╗
║          RTX Trading - Backend v3.0                       ║
║        BBOOK Mode + OANDA LP                              ║
╚═══════════════════════════════════════════════════════════╝
[B-Book] Demo account created: ACC-1 with $10000.00
[INFO] Server running on :7999
```

---

## Start Desktop Client

```bash
cd C:\Users\Yofor\Desktop\Trading-Engine\clients\desktop
npm run dev
```

**Then open:** `http://localhost:5173`

---

## Test Login (cURL)

### Admin Login
```bash
curl -X POST http://localhost:7999/admin/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"admin\",\"password\":\"Admin@123\",\"ip_address\":\"127.0.0.1\",\"user_agent\":\"curl\"}"
```

### Demo User Login
```bash
curl -X POST http://localhost:7999/login \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"demo-user\",\"password\":\"password\"}"
```

---

## Database Issue?

**Symptom:** Login fails with "Connection refused"

**Quick Fix:** These accounts work WITHOUT database:
- ✅ Admin: `admin / Admin@123`
- ✅ Demo: `demo-user / password`

**For full system:** See `TEST_CREDENTIALS_REPORT.md`

---

## What Works Without Database

| Feature | Status |
|---------|--------|
| Admin Panel Login | ✅ Works |
| Desktop Client Login (demo-user) | ✅ Works |
| Place Demo Orders | ✅ Works |
| View Market Data | ✅ Works |
| Admin User Management | ❌ Needs DB |
| Persistent Accounts | ❌ Needs DB |
| Historical Data | ❌ Needs DB |

---

## Default Configuration

From `backend/.env`:

```env
PORT=7999
DEFAULT_ACCOUNT_BALANCE=10000.0
DEFAULT_ACCOUNT_LEVERAGE=100
EXECUTION_MODE=BBOOK
```

**Demo account gets $10,000 balance automatically.**

---

## Password Security Notes

**Development Mode:**
- Hardcoded credentials are INTENTIONAL for local dev
- Backend auto-creates demo account on startup
- No database setup required for basic testing

**Production Mode:**
- NEVER use these credentials
- Set `ADMIN_PASSWORD_HASH` in .env
- Use strong, unique passwords
- Enable 2FA

---

## Troubleshooting

### "Connection refused" error
- **Not a login issue** - PostgreSQL is stopped
- Use in-memory accounts (admin, demo-user)
- Or start PostgreSQL: `net start postgresql-x64-18` (requires admin)

### "Invalid credentials"
- Username is case-sensitive
- Check you're using correct endpoint:
  - `/admin/login` for admin
  - `/login` for users

### Port already in use
- Default port: 7999
- Change in .env: `PORT=8080`

---

## Advanced: Database Setup

If you need database users:

1. **Start PostgreSQL** (requires admin rights)
2. **Run migrations**: `go run cmd/migrate/main.go`
3. **Seed data**: `bash scripts/db/seed-test-data.sh`

See full guide: `docs/TEST_CREDENTIALS_REPORT.md`

---

## More Info

- Full credentials report: `docs/TEST_CREDENTIALS_REPORT.md`
- README: `README.md`
- Environment setup: `backend/.env.example`
