# RBAC Implementation Summary

## Overview
Role-Based Access Control (RBAC) has been implemented for the RTX5 admin backend to provide fine-grained permission management across different administrative roles.

## Files Created

### 1. `backend/admin/rbac.go`
Core RBAC logic including:
- **Permission constants**: 11 permission types (ACCOUNTS_READ, ACCOUNTS_WRITE, TRADING_READ, TRADING_EXECUTE, SETTINGS_READ, SETTINGS_WRITE, USERS_READ, USERS_MANAGE, REPORTS_READ, LP_MANAGE, RISK_MANAGE)
- **Role hierarchy**: 6 roles with numerical levels (SUPER_ADMIN: 100, ADMIN: 80, MANAGER: 60, DEALER: 40, SUPPORT: 20, VIEWER: 10)
- **Permission mapping**: Maps each role to its allowed permissions
- **Helper functions**:
  - `HasPermission(role, permission)`: Check if role has specific permission
  - `HasRole(userRole, requiredRole)`: Check if user's role meets minimum requirement
  - `GetPermissions(role)`: Get all permissions for a role
  - `GetAllRoles()`: Get all roles with hierarchy levels

### 2. `backend/admin/rbac_middleware.go`
HTTP middleware for enforcing RBAC:
- **RBACMiddleware struct**: Wraps AuthService for authentication
- **Middleware functions**:
  - `RequirePermission(permission)`: Enforce single permission
  - `RequireRole(role)`: Enforce minimum role level
  - `RequireAnyPermission(permissions...)`: Enforce at least one permission
- **Features**:
  - CORS handling
  - Bearer token authentication
  - HTTP 403 Forbidden on insufficient permissions
  - Logging of permission denials

### 3. `backend/admin/roles_handler.go`
HTTP handlers for role management:
- **RoleStore**: In-memory store for user role assignments (map[userID]AdminRole)
- **HTTP Endpoints**:
  - `GET /admin/roles`: List all roles and their permissions (SUPER_ADMIN only)
  - `GET /admin/users/role?userId=X`: Get user's role
  - `PUT /admin/users/role/update`: Update user's role (SUPER_ADMIN only)
- **Thread-safe**: Uses sync.RWMutex for concurrent access

## Role Hierarchy

```
SUPER_ADMIN (100) ── All permissions
    ├─ ADMIN (80) ── All except USERS_MANAGE
    ├─ MANAGER (60) ── Read-only + Reports
    ├─ DEALER (40) ── Trading execution + Read
    ├─ SUPPORT (20) ── Read-only access
    └─ VIEWER (10) ── Read-only access
```

## Permission Matrix

| Role | Accounts | Trading | Settings | Users | Reports | LP | Risk |
|------|----------|---------|----------|-------|---------|----|----|
| **SUPER_ADMIN** | R/W | R/E | R/W | R/M | R | M | M |
| **ADMIN** | R/W | R/E | R/W | R | R | M | M |
| **MANAGER** | R | R | R | - | R | - | - |
| **DEALER** | R | R/E | - | - | - | - | - |
| **SUPPORT** | R | R | R | - | R | - | - |
| **VIEWER** | R | R | R | - | R | - | - |

Legend: R=Read, W=Write, E=Execute, M=Manage

## Integration with Existing Code

### Modified Files

#### `backend/admin/handlers.go`
**Changes**:
1. Added `rbacMiddleware` and `rolesHandler` fields to `AdminHandler`
2. Initialized RBAC components in `NewAdminHandler()`
3. Modified `RegisterRoutes()` to wrap handlers with permission checks:
   - User management routes → `PermUsersRead`, `PermAccountsWrite`
   - Fund management routes → `PermAccountsWrite`
   - Order management routes → `PermTradingRead`, `PermTradingExecute`
   - Group management routes → `PermSettingsRead`, `PermSettingsWrite`
   - Audit trail → `PermReportsRead`
4. Added helper methods:
   - `GetRoleStore()`: Access role store externally
   - `GetAuthService()`: Access auth service externally

#### `backend/admin/auth.go`
**Changes**:
1. Added role logging in `Login()` method (line 137)
2. Sessions now include role from admin record

## Backward Compatibility

**Preserved**:
- Existing `AdminRole` constants in `types.go` (RoleSuperAdmin, RoleAdmin, RoleSupport)
- Existing session-based authentication mechanism
- All existing handler signatures unchanged
- Default admin account created with RoleSuperAdmin

**Defaults**:
- Admins without explicit role assignment default to `ADMIN` role
- Existing sessions continue to work with role from admin record

## Route Protection Summary

### Unprotected Routes (Public)
- `POST /admin/auth/login`
- `POST /admin/auth/logout`

### Protected Routes

#### SUPER_ADMIN Only
- `GET /admin/roles` - List all roles
- `PUT /admin/users/role/update` - Change user roles

#### USERS_READ Permission
- `GET /admin/users` - List users
- `GET /admin/user` - Get user details
- `GET /admin/users/role` - Get user role

#### ACCOUNTS_WRITE Permission
- `PUT /admin/user/update` - Update user account
- `POST /admin/user/enable` - Enable user
- `POST /admin/user/disable` - Disable user
- `POST /admin/user/reset-password` - Reset password
- `POST /admin/fund/deposit` - Deposit funds
- `POST /admin/fund/withdraw` - Withdraw funds
- `POST /admin/fund/adjust` - Adjust balance
- `POST /admin/fund/bonus` - Add bonus

#### TRADING_READ Permission
- `GET /admin/orders` - View all orders
- `GET /admin/positions` - View all positions

#### TRADING_EXECUTE Permission
- `PUT /admin/order/modify` - Modify order
- `DELETE /admin/order/delete` - Delete order
- `PUT /admin/position/modify` - Modify position
- `PUT /admin/position/reverse` - Reverse position
- `POST /admin/position/close` - Close position

#### SETTINGS_READ Permission
- `GET /admin/groups` - List groups

#### SETTINGS_WRITE Permission
- `POST /admin/group/create` - Create group
- `PUT /admin/group/update` - Update group
- `DELETE /admin/group/delete` - Delete group

#### REPORTS_READ Permission
- `GET /admin/audit` - View audit log

## Usage Examples

### Creating Admin with Specific Role
```go
adminHandler := admin.NewAdminHandler(engine)
authService := adminHandler.GetAuthService()
roleStore := adminHandler.GetRoleStore()

// Create admin
newAdmin, err := authService.CreateAdmin("dealer1", "dealer@example.com", "password", admin.RoleDealer, nil, "SYSTEM")
if err != nil {
    log.Fatal(err)
}

// Assign role in role store (for JWT integration)
roleStore.SetRole(newAdmin.ID, admin.RoleDealer)
```

### Checking Permissions Programmatically
```go
import "github.com/epic1st/rtx/backend/admin"

// Check if role has permission
if admin.HasPermission(admin.RoleDealer, admin.PermTradingExecute) {
    // Allow trading operations
}

// Check role hierarchy
if admin.HasRole(currentUser.Role, admin.RoleManager) {
    // User is at least a Manager
}
```

### API Usage (Client Side)
```javascript
// Login
const response = await fetch('/admin/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'dealer1', password: 'password' })
});
const { token, session } = await response.json();

// Use token for authenticated requests
const orders = await fetch('/admin/orders', {
    headers: { 'Authorization': `Bearer ${token}` }
});

// Get user's role
const roleInfo = await fetch(`/admin/users/role?userId=${userId}`, {
    headers: { 'Authorization': `Bearer ${token}` }
});
```

## Security Considerations

### Implemented
1. **Role validation**: All role assignments validated against hierarchy
2. **Permission checks**: Every protected route requires explicit permission
3. **Audit logging**: Permission denials logged with username and role
4. **Session validation**: Token validated before permission check
5. **IP verification**: Existing IP whitelist mechanism preserved

### Best Practices
1. **Principle of least privilege**: Each role has minimal necessary permissions
2. **Defense in depth**: Middleware + handler-level checks
3. **Fail secure**: Unknown roles/permissions denied by default
4. **Audit trail**: All permission denials logged

## Testing Checklist

- [ ] Login as SUPER_ADMIN → should access all routes
- [ ] Login as ADMIN → should be denied `/admin/users/role/update`
- [ ] Login as MANAGER → should view but not modify
- [ ] Login as DEALER → should execute trades but not modify settings
- [ ] Login as VIEWER → should only read, all writes denied
- [ ] Create admin with each role → verify role stored correctly
- [ ] Update user role (SUPER_ADMIN only) → verify permission check
- [ ] Invalid token → should get 401 Unauthorized
- [ ] Valid token, insufficient permissions → should get 403 Forbidden

## Future Enhancements

### Recommended
1. **JWT integration**: Add role claim to JWT tokens (currently session-based)
2. **Persistence**: Store role assignments in PostgreSQL
3. **Dynamic permissions**: Allow runtime permission updates
4. **Permission inheritance**: Group-based permission inheritance
5. **Time-based roles**: Temporary elevated privileges
6. **Multi-factor auth**: Integrate with existing 2FA system

### Optional
1. **Custom permissions**: Allow admins to define custom permissions
2. **Permission templates**: Pre-configured role templates
3. **Activity monitoring**: Real-time permission usage dashboard
4. **Compliance reports**: Permission audit reports for compliance

## Configuration

No configuration files needed. All roles and permissions defined in code (`rbac.go`).

To modify permissions:
1. Edit `rolePermissions` map in `backend/admin/rbac.go`
2. Restart server

To add new roles:
1. Add constant to `backend/admin/rbac.go`
2. Add to `roleHierarchy` with numerical level
3. Add to `rolePermissions` with permission list
4. Restart server

## Deployment Notes

### Development
- Default SUPER_ADMIN created: username=admin, password=Admin@123
- Role store is in-memory (lost on restart)
- No persistence required

### Production
1. Change default admin password immediately
2. Create separate admin accounts for each role
3. Consider implementing PostgreSQL persistence for role store
4. Enable audit logging to file/database
5. Configure IP whitelisting per admin

## Support

For issues or questions:
- Check logs: `[RBAC]` prefix for permission-related logs
- Verify role assignment in `/admin/users/role` endpoint
- Review audit trail in `/admin/audit` for permission denials
