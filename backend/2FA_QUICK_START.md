# 2FA Quick Start Guide

## Setup (5 minutes)

### 1. Add Environment Variables

Add to your `.env` file:

```bash
# Generate a 32-byte key:
# Linux/Mac: openssl rand -base64 32 | head -c 32
# PowerShell: [Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Minimum 0 -Maximum 256 }))

TOTP_ENCRYPTION_KEY=your_32_byte_key_here_exactly32
TOTP_ISSUER=RTX5 Trading Engine
```

### 2. Install Dependencies

```bash
cd backend
go get github.com/pquerna/otp
go mod tidy
```

### 3. Start Server

```bash
go run cmd/server/main.go
```

## Quick Test (2 minutes)

### Step 1: Login as Admin

```bash
curl -X POST http://localhost:7999/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'
```

Save the `token` from response.

### Step 2: Setup 2FA

```bash
TOKEN="your_token_here"

curl -X POST http://localhost:7999/admin/2fa/setup \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"adminId":1}'
```

Copy the `qrCodeUrl` from response.

### Step 3: Scan QR Code

1. Open Google Authenticator or Authy on your phone
2. Add new account → Scan QR code
3. Or paste the URL in a browser to see QR code

### Step 4: Verify Code

Get the 6-digit code from your authenticator app:

```bash
curl -X POST http://localhost:7999/admin/2fa/verify \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"adminId":1,"code":"123456"}'
```

Replace `123456` with your actual code.

### Step 5: Enable 2FA

```bash
curl -X POST http://localhost:7999/admin/2fa/enable \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"adminId":1,"code":"654321"}'
```

**Save the backup codes!** They're one-time use only.

### Step 6: Test Login with 2FA

```bash
# First step - password
curl -X POST http://localhost:7999/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}'
```

Response will include:
```json
{
  "success": false,
  "requires2FA": true,
  "partialToken": "eyJhbGc...",
  "message": "2FA verification required..."
}
```

Save the `partialToken`.

```bash
# Second step - TOTP code
PARTIAL_TOKEN="your_partial_token_here"

curl -X POST http://localhost:7999/admin/2fa/validate \
  -H "Content-Type: application/json" \
  -d '{"token":"'$PARTIAL_TOKEN'","code":"789012"}'
```

Replace `789012` with code from authenticator app.

Success! You now have the full session token.

## Troubleshooting

### "Invalid 2FA code"
- Check your phone's time is synchronized
- TOTP codes expire every 30 seconds
- Try the next code if current one fails

### "TOTP_ENCRYPTION_KEY environment variable is required"
- Add `TOTP_ENCRYPTION_KEY` to `.env` file
- Must be exactly 32 bytes (32 characters)

### "Failed to decrypt secret"
- Don't change `TOTP_ENCRYPTION_KEY` after enabling 2FA
- If changed, all existing secrets become invalid

### Server restart lost my 2FA config
- Current implementation uses in-memory storage
- Implement database persistence for production

## Disable 2FA

```bash
TOKEN="your_session_token"

curl -X POST http://localhost:7999/admin/2fa/disable \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"adminId":1,"code":"123456"}'
```

## Generate New Backup Codes

```bash
curl -X POST http://localhost:7999/admin/2fa/backup-codes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

Old backup codes become invalid.

## API Endpoints Summary

| Endpoint | Auth Required | Purpose |
|----------|---------------|---------|
| `POST /admin/2fa/setup` | Yes | Generate QR code |
| `POST /admin/2fa/verify` | Yes | Verify during setup |
| `POST /admin/2fa/enable` | Yes | Enable 2FA + get backup codes |
| `POST /admin/2fa/disable` | Yes | Disable 2FA |
| `POST /admin/2fa/validate` | No (partial token) | Validate TOTP during login |
| `POST /admin/2fa/backup-codes` | Yes | Generate new backup codes |

## Security Notes

- **Encryption Key:** Never commit `TOTP_ENCRYPTION_KEY` to git
- **Backup Codes:** Store securely, each can only be used once
- **QR Codes:** Don't share screenshots of QR codes
- **Partial Tokens:** Expire in 5 minutes
- **Full Sessions:** Expire in 8 hours (configurable)

## Next Steps

1. Implement database persistence
2. Add rate limiting (5 attempts per 15 min)
3. Create frontend UI for enrollment
4. Add recovery flow for lost devices
5. Monitor failed 2FA attempts

See `2FA_IMPLEMENTATION_SUMMARY.md` for complete documentation.
