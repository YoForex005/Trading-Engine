# TOTP Two-Factor Authentication (2FA) Implementation

## Overview

This implementation adds TOTP-based Two-Factor Authentication (2FA) to the RTX5 Trading Engine backend admin system. It supports Google Authenticator, Authy, and other TOTP-compatible authenticator apps.

## Architecture

### Files Created

1. **`backend/auth/totp.go`** - TOTP generation and validation service
   - `NewTOTPService()` - Creates TOTP service with AES-256-GCM encryption
   - `GenerateTOTPSecret()` - Generates encrypted TOTP secret and QR code URL
   - `ValidateTOTPCode()` - Validates 6-digit TOTP codes
   - `encryptSecret()` / `decryptSecret()` - AES-256-GCM encryption for secrets

2. **`backend/auth/totp_store.go`** - In-memory TOTP configuration store
   - `TOTPConfig` - Stores secret, enabled status, backup codes
   - `EnableTOTP()` / `DisableTOTP()` - Enable/disable 2FA
   - `GenerateBackupCodes()` - Generates 10 one-time backup codes
   - `ValidateBackupCode()` - Validates and consumes backup codes

3. **`backend/auth/errors.go`** - 2FA-specific error types
   - `ErrTOTPNotEnabled`, `ErrInvalidTOTPCode`, `Err2FARequired`, etc.

4. **`backend/admin/totp_handlers.go`** - HTTP handlers for 2FA endpoints
   - `HandleTOTPSetup` - Initiates 2FA enrollment
   - `HandleTOTPVerify` - Verifies code during setup
   - `HandleTOTPEnable` - Enables 2FA after verification
   - `HandleTOTPDisable` - Disables 2FA
   - `HandleTOTPValidate` - Validates TOTP during login (second step)
   - `HandleBackupCodes` - Generates new backup codes

### Files Modified

1. **`backend/admin/auth.go`**
   - Updated `Login()` to check for 2FA and return partial session if enabled
   - Returns `"2FA_REQUIRED"` error when 2FA is enabled

2. **`backend/admin/handlers.go`**
   - Added JWT and time imports
   - Updated `HandleLogin()` to generate partial JWT token when 2FA is required
   - Updated `RegisterRoutes()` to register 2FA endpoints

3. **`backend/go.mod`**
   - Added `github.com/pquerna/otp v1.4.0` dependency

4. **`backend/.env.example`**
   - Added `TOTP_ENCRYPTION_KEY` (32 bytes for AES-256)
   - Added `TOTP_ISSUER` (display name in authenticator apps)

## API Endpoints

### 2FA Enrollment Flow

#### 1. Setup - POST `/admin/2fa/setup`
**Requires:** Admin authentication (Bearer token)

Initiates 2FA enrollment by generating a TOTP secret and QR code URL.

**Request:**
```json
{
  "adminId": 1
}
```

**Response:**
```json
{
  "secret": "BASE32ENCODEDSECRET",
  "qrCodeUrl": "otpauth://totp/RTX5%20Trading%20Engine:admin?secret=BASE32ENCODEDSECRET&issuer=RTX5%20Trading%20Engine",
  "message": "Scan QR code with Google Authenticator or Authy, then verify with a code"
}
```

#### 2. Verify - POST `/admin/2fa/verify`
**Requires:** Admin authentication (Bearer token)

Verifies that the admin can generate valid codes with their authenticator app.

**Request:**
```json
{
  "adminId": 1,
  "code": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Code verified. Call /admin/2fa/enable to complete setup"
}
```

#### 3. Enable - POST `/admin/2fa/enable`
**Requires:** Admin authentication (Bearer token)

Enables 2FA after successful verification and generates backup codes.

**Request:**
```json
{
  "adminId": 1,
  "code": "123456"
}
```

**Response:**
```json
{
  "backupCodes": [
    "AB12CD34EF56",
    "GH78IJ90KL12",
    ...
  ],
  "message": "2FA enabled successfully. Save these backup codes in a secure location. Each code can only be used once."
}
```

### Login Flow with 2FA

#### 1. Login - POST `/admin/auth/login`
**No authentication required**

**Request:**
```json
{
  "username": "admin",
  "password": "Admin@123"
}
```

**Response (2FA not enabled):**
```json
{
  "success": true,
  "session": {
    "sessionId": "abc123...",
    "adminId": 1,
    "username": "admin",
    "role": "SUPER_ADMIN",
    ...
  },
  "token": "abc123..."
}
```

**Response (2FA enabled):**
```json
{
  "success": false,
  "requires2FA": true,
  "partialToken": "eyJhbGciOiJIUzI1NiIs...",
  "message": "2FA verification required. POST to /admin/2fa/validate with code"
}
```

#### 2. Validate 2FA - POST `/admin/2fa/validate`
**No authentication required** (uses partial token)

**Request:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "code": "123456"
}
```

**Response:**
```json
{
  "sessionId": "def456...",
  "adminId": 1,
  "username": "admin",
  "role": "SUPER_ADMIN",
  ...
}
```

### 2FA Management

#### Disable 2FA - POST `/admin/2fa/disable`
**Requires:** Admin authentication (Bearer token)

**Request:**
```json
{
  "adminId": 1,
  "code": "123456"
}
```

**Response:**
```json
{
  "success": true,
  "message": "2FA disabled successfully"
}
```

#### Generate New Backup Codes - POST `/admin/2fa/backup-codes`
**Requires:** Admin authentication (Bearer token)

**Request:** Empty body

**Response:**
```json
{
  "backupCodes": [
    "NEW1CODE2HERE",
    "NEW3CODE4HERE",
    ...
  ],
  "message": "New backup codes generated. Previous codes are now invalid. Save these in a secure location."
}
```

## Security Features

1. **AES-256-GCM Encryption**
   - TOTP secrets are encrypted at rest using AES-256-GCM
   - Encryption key must be exactly 32 bytes (256 bits)
   - Each encryption uses a unique nonce

2. **Partial JWT Tokens**
   - Login returns a short-lived (5 min) JWT token when 2FA is required
   - Token contains `"2fa_required": true` claim
   - Different signing secret than full session tokens

3. **Backup Codes**
   - 10 one-time backup codes generated during enrollment
   - Each code is 12 characters (base64 encoded)
   - Codes are consumed after first use
   - Can be regenerated at any time

4. **Time Skew Tolerance**
   - Validates TOTP codes with 1 period grace (±30 seconds)
   - Prevents issues due to clock drift

5. **Backward Compatible**
   - Admins without 2FA enabled continue to login normally
   - No breaking changes to existing login flow

## Environment Variables

Add these to your `.env` file:

```bash
# TOTP_ENCRYPTION_KEY must be exactly 32 bytes for AES-256-GCM
# Generate with: openssl rand -base64 32 | head -c 32
TOTP_ENCRYPTION_KEY=your_32_byte_totp_encryption_key

# Display name in authenticator apps (optional)
TOTP_ISSUER=RTX5 Trading Engine
```

## Database Schema

The existing `Admin` struct in `backend/admin/types.go` already includes:

```go
type Admin struct {
    // ... other fields ...
    TwoFactorSecret  string `json:"-"`           // Encrypted TOTP secret
    TwoFactorEnabled bool   `json:"twoFactorEnabled"` // 2FA status
}
```

**Note:** Current implementation uses in-memory storage. For production, you should:
1. Store `TwoFactorSecret` and `TwoFactorEnabled` in PostgreSQL
2. Hash and store backup codes in the database
3. Implement proper state persistence

## Testing

### Manual Testing Flow

1. **Enable 2FA:**
   ```bash
   # Login as admin
   curl -X POST http://localhost:7999/admin/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"Admin@123"}'

   # Save the token
   TOKEN="<session_token>"

   # Setup 2FA
   curl -X POST http://localhost:7999/admin/2fa/setup \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"adminId":1}'

   # Scan QR code with Google Authenticator

   # Verify with generated code
   curl -X POST http://localhost:7999/admin/2fa/verify \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"adminId":1,"code":"123456"}'

   # Enable 2FA
   curl -X POST http://localhost:7999/admin/2fa/enable \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{"adminId":1,"code":"654321"}'
   ```

2. **Login with 2FA:**
   ```bash
   # First step - password
   curl -X POST http://localhost:7999/admin/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"Admin@123"}'

   # Response will include partialToken
   PARTIAL_TOKEN="<partial_token>"

   # Second step - TOTP code
   curl -X POST http://localhost:7999/admin/2fa/validate \
     -H "Content-Type: application/json" \
     -d '{"token":"'$PARTIAL_TOKEN'","code":"123456"}'
   ```

3. **Use Backup Code:**
   ```bash
   # If TOTP device is unavailable, use a backup code
   curl -X POST http://localhost:7999/admin/2fa/validate \
     -H "Content-Type: application/json" \
     -d '{"token":"'$PARTIAL_TOKEN'","code":"AB12CD34EF56"}'
   ```

## Frontend Integration

### React Example

```typescript
import { useState } from 'react';

function LoginForm() {
  const [step, setStep] = useState<'password' | '2fa'>('password');
  const [partialToken, setPartialToken] = useState<string>('');

  const handlePasswordSubmit = async (username: string, password: string) => {
    const response = await fetch('http://localhost:7999/admin/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });

    const data = await response.json();

    if (data.requires2FA) {
      setPartialToken(data.partialToken);
      setStep('2fa');
    } else {
      // Login successful, store token
      localStorage.setItem('token', data.token);
    }
  };

  const handle2FASubmit = async (code: string) => {
    const response = await fetch('http://localhost:7999/admin/2fa/validate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token: partialToken, code })
    });

    const data = await response.json();
    localStorage.setItem('token', data.sessionId);
  };

  // Render forms based on step...
}
```

## Production Considerations

1. **State Persistence**
   - Migrate from in-memory store to PostgreSQL
   - Add database migrations for `two_factor_secret`, `two_factor_enabled`, `backup_codes`

2. **Audit Logging**
   - All 2FA events are logged via `AuditLog`
   - Track: setup, enable, disable, failed attempts, backup code usage

3. **Rate Limiting**
   - Implement rate limiting on `/admin/2fa/validate`
   - Max 5 attempts per 15 minutes to prevent brute force

4. **Session Management**
   - Invalidate partial tokens after successful 2FA validation
   - Clean up expired partial tokens

5. **Recovery Flow**
   - Admin recovery process if 2FA device is lost
   - Require super admin approval for 2FA reset

6. **Monitoring**
   - Alert on multiple failed 2FA attempts
   - Track 2FA adoption rate
   - Monitor backup code usage patterns

## Known Limitations

1. **In-Memory Storage**
   - TOTP configs and backup codes are stored in memory
   - Will be lost on server restart
   - Must implement database persistence for production

2. **No Account Recovery**
   - No built-in recovery flow if both TOTP device and backup codes are lost
   - Requires manual intervention by super admin

3. **No IP-Based Risk Assessment**
   - Could enhance with IP-based risk scoring
   - Skip 2FA for trusted IPs (configurable)

4. **No SMS/Email Fallback**
   - Only supports TOTP-based 2FA
   - Could add SMS/Email as fallback methods

## Next Steps

1. Implement PostgreSQL persistence for TOTP configs
2. Add rate limiting to prevent brute force attacks
3. Create frontend UI components for 2FA enrollment
4. Add admin recovery flow for lost 2FA devices
5. Implement IP-based risk assessment
6. Add comprehensive unit tests
7. Add integration tests for full 2FA flow
8. Document API in OpenAPI/Swagger

## Support

For issues or questions:
- Check logs in `server.log` for 2FA events (prefix: `[2FA]`)
- Verify `TOTP_ENCRYPTION_KEY` is exactly 32 bytes
- Ensure authenticator app time is synchronized
- Check audit log for failed attempts: `GET /admin/audit?action=2FA`
