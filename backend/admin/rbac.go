package admin

// Permission represents a specific permission type
type Permission string

// Permission constants
const (
	// Account permissions
	PermAccountsRead  Permission = "ACCOUNTS_READ"
	PermAccountsWrite Permission = "ACCOUNTS_WRITE"

	// Trading permissions
	PermTradingRead    Permission = "TRADING_READ"
	PermTradingExecute Permission = "TRADING_EXECUTE"

	// Settings permissions
	PermSettingsRead  Permission = "SETTINGS_READ"
	PermSettingsWrite Permission = "SETTINGS_WRITE"

	// User management permissions
	PermUsersRead   Permission = "USERS_READ"
	PermUsersManage Permission = "USERS_MANAGE"

	// Reports permissions
	PermReportsRead Permission = "REPORTS_READ"

	// LP management permissions
	PermLPManage Permission = "LP_MANAGE"

	// Risk management permissions
	PermRiskManage Permission = "RISK_MANAGE"
)

// Role constants (using existing AdminRole type from types.go)
const (
	RBACRoleManager AdminRole = "MANAGER" // Manager role
	RBACRoleDealer  AdminRole = "DEALER"  // Dealer role
	RBACRoleViewer  AdminRole = "VIEWER"  // Viewer role
)

// rolePermissions maps roles to their permissions
var rolePermissions = map[AdminRole][]Permission{
	RoleSuperAdmin: {
		// Super admin has ALL permissions
		PermAccountsRead,
		PermAccountsWrite,
		PermTradingRead,
		PermTradingExecute,
		PermSettingsRead,
		PermSettingsWrite,
		PermUsersRead,
		PermUsersManage,
		PermReportsRead,
		PermLPManage,
		PermRiskManage,
	},
	RoleAdmin: {
		// Admin has all except USERS_MANAGE
		PermAccountsRead,
		PermAccountsWrite,
		PermTradingRead,
		PermTradingExecute,
		PermSettingsRead,
		PermSettingsWrite,
		PermUsersRead,
		PermReportsRead,
		PermLPManage,
		PermRiskManage,
	},
	RBACRoleManager: {
		// Manager has read access and reports
		PermAccountsRead,
		PermTradingRead,
		PermReportsRead,
		PermSettingsRead,
	},
	RBACRoleDealer: {
		// Dealer can read and execute trading
		PermAccountsRead,
		PermTradingRead,
		PermTradingExecute,
	},
	RBACRoleViewer: {
		// Viewer has read-only access
		PermAccountsRead,
		PermTradingRead,
		PermSettingsRead,
		PermReportsRead,
	},
	RoleSupport: {
		// Support has read-only access (existing role)
		PermAccountsRead,
		PermTradingRead,
		PermSettingsRead,
		PermReportsRead,
	},
}

// roleHierarchy defines the hierarchy of roles (higher value = more privilege)
var roleHierarchy = map[AdminRole]int{
	RoleSuperAdmin: 100,
	RoleAdmin:      80,
	RBACRoleManager:    60,
	RBACRoleDealer:     40,
	RoleSupport:    20,
	RBACRoleViewer:     10,
}

// HasPermission checks if a role has a specific permission
func HasPermission(role AdminRole, permission Permission) bool {
	permissions, exists := rolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// HasRole checks if userRole meets or exceeds requiredRole in the hierarchy
func HasRole(userRole AdminRole, requiredRole AdminRole) bool {
	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// GetPermissions returns all permissions for a given role
func GetPermissions(role AdminRole) []Permission {
	permissions, exists := rolePermissions[role]
	if !exists {
		return []Permission{}
	}

	// Return a copy to prevent modification
	result := make([]Permission, len(permissions))
	copy(result, permissions)
	return result
}

// GetAllRoles returns all available roles with their hierarchy levels
func GetAllRoles() map[AdminRole]int {
	result := make(map[AdminRole]int, len(roleHierarchy))
	for role, level := range roleHierarchy {
		result[role] = level
	}
	return result
}
