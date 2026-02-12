# RBAC Quick Reference

## Roles (Hierarchy)

```
SUPER_ADMIN (100) ► ADMIN (80) ► MANAGER (60) ► DEALER (40) ► SUPPORT/VIEWER (20/10)
```

## Permissions

| Permission | Description |
|------------|-------------|
| `ACCOUNTS_READ` | View user accounts and balances |
| `ACCOUNTS_WRITE` | Modify accounts, deposits, withdrawals |
| `TRADING_READ` | View orders and positions |
| `TRADING_EXECUTE` | Execute, modify, close trades |
| `SETTINGS_READ` | View system settings and groups |
| `SETTINGS_WRITE` | Modify system settings |
| `USERS_READ` | View user list |
| `USERS_MANAGE` | Create/delete admins, assign roles |
| `REPORTS_READ` | View audit logs and reports |
| `LP_MANAGE` | Manage liquidity providers |
| `RISK_MANAGE` | Manage risk parameters |

## Role Permissions Matrix

| Role | Permissions |
|------|------------|
| **SUPER_ADMIN** | ALL (11 permissions) |
| **ADMIN** | ALL except USERS_MANAGE (10 permissions) |
| **MANAGER** | ACCOUNTS_READ, TRADING_READ, REPORTS_READ, SETTINGS_READ |
| **DEALER** | ACCOUNTS_READ, TRADING_READ, TRADING_EXECUTE |
| **SUPPORT/VIEWER** | ACCOUNTS_READ, TRADING_READ, SETTINGS_READ, REPORTS_READ |

## API Endpoints

### Role Management (SUPER_ADMIN only)
```bash
# List all roles
GET /admin/roles
Authorization: Bearer {token}

# Get user's role
GET /admin/users/role?userId=123
Authorization: Bearer {token}

# Update user's role
PUT /admin/users/role/update
Authorization: Bearer {token}
Content-Type: application/json
{
  "userId": 123,
  "role": "DEALER",
  "reason": "Promoted to dealer position"
}
```

### Protected Endpoints by Permission

**ACCOUNTS_WRITE**:
- `POST /admin/fund/deposit`
- `POST /admin/fund/withdraw`
- `POST /admin/fund/adjust`
- `POST /admin/fund/bonus`
- `PUT /admin/user/update`
- `POST /admin/user/enable`
- `POST /admin/user/disable`
- `POST /admin/user/reset-password`

**TRADING_READ**:
- `GET /admin/orders`
- `GET /admin/positions`

**TRADING_EXECUTE**:
- `PUT /admin/order/modify`
- `DELETE /admin/order/delete`
- `PUT /admin/position/modify`
- `PUT /admin/position/reverse`
- `POST /admin/position/close`

**SETTINGS_WRITE**:
- `POST /admin/group/create`
- `PUT /admin/group/update`
- `DELETE /admin/group/delete`

**USERS_READ**:
- `GET /admin/users`
- `GET /admin/user`

**SETTINGS_READ**:
- `GET /admin/groups`

**REPORTS_READ**:
- `GET /admin/audit`

## HTTP Response Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `401` | Unauthorized (invalid/missing token) |
| `403` | Forbidden (insufficient permissions) |
| `400` | Bad request (invalid data) |

## Code Examples

### Check Permission (Go)
```go
import "github.com/epic1st/rtx/backend/admin"

// Check single permission
if admin.HasPermission(user.Role, admin.PermTradingExecute) {
    // Allow
}

// Check role level
if admin.HasRole(user.Role, admin.RoleManager) {
    // User is at least Manager
}
```

### Create Admin with Role (Go)
```go
adminHandler := admin.NewAdminHandler(engine)
authService := adminHandler.GetAuthService()
roleStore := adminHandler.GetRoleStore()

// Create dealer admin
dealer, err := authService.CreateAdmin(
    "dealer1",
    "dealer@example.com",
    "SecurePass123!",
    admin.RoleDealer,
    []string{"192.168.1.0/24"}, // IP whitelist
    "admin", // Created by
)

// Store role
roleStore.SetRole(dealer.ID, admin.RoleDealer)
```

### API Call (JavaScript)
```javascript
// Login
const { token } = await fetch('/admin/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password })
}).then(r => r.json());

// Protected request
const response = await fetch('/admin/orders', {
    headers: { 'Authorization': `Bearer ${token}` }
});

if (response.status === 403) {
    alert('Insufficient permissions');
}
```

## Troubleshooting

### "Unauthorized" Error
- Check token is present: `Authorization: Bearer {token}`
- Verify token hasn't expired (8-hour sessions)
- Try re-login

### "Insufficient permissions" Error
- Check user's role: `GET /admin/users/role?userId={id}`
- Verify required permission for endpoint
- Contact SUPER_ADMIN to upgrade role

### Logs
```bash
# Check RBAC logs
grep "\[RBAC\]" server.log

# Check admin auth logs
grep "\[AdminAuth\]" server.log
```

## Common Tasks

### Give User Admin Access
```bash
curl -X PUT http://localhost:7999/admin/users/role/update \
  -H "Authorization: Bearer {super_admin_token}" \
  -H "Content-Type: application/json" \
  -d '{"userId": 5, "role": "ADMIN", "reason": "Promoted to admin"}'
```

### List All Roles
```bash
curl http://localhost:7999/admin/roles \
  -H "Authorization: Bearer {super_admin_token}"
```

### Check Current User Role
```bash
curl "http://localhost:7999/admin/users/role?userId=5" \
  -H "Authorization: Bearer {token}"
```

## Default Credentials

- **Username**: `admin`
- **Password**: `Admin@123`
- **Role**: `SUPER_ADMIN`

**⚠️ CHANGE DEFAULT PASSWORD IN PRODUCTION**
