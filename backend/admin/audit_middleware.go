//go:build rtx_legacy_admin
// +build rtx_legacy_admin

package admin

import (
	"net/http"
	"strings"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// AuditMiddleware creates middleware that logs admin actions
func AuditMiddleware(logger *AuditLogger, authService AuthTokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip audit logging endpoints themselves to avoid recursion
			if strings.HasPrefix(r.URL.Path, "/admin/audit/") {
				next.ServeHTTP(w, r)
				return
			}

			// Capture response status
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				written:        false,
			}

			// Extract admin info from JWT token
			adminID, adminEmail := extractAdminInfo(r, authService)

			// Get IP address
			ipAddress := getIPAddress(r)

			// Start time
			startTime := time.Now()

			// Call next handler
			next.ServeHTTP(wrapper, r)

			// Log the action
			action := mapHTTPToAction(r.Method, r.URL.Path)
			resource, resourceID := extractResourceInfo(r.URL.Path)

			entry := AuditLogEntry{
				AdminID:    adminID,
				AdminEmail: adminEmail,
				Action:     action,
				Resource:   resource,
				ResourceID: resourceID,
				Details: map[string]interface{}{
					"method":       r.Method,
					"path":         r.URL.Path,
					"query":        r.URL.RawQuery,
					"userAgent":    r.Header.Get("User-Agent"),
					"responseTime": time.Since(startTime).Milliseconds(),
				},
				IPAddress: ipAddress,
				Timestamp: time.Now().Unix(),
				Success:   wrapper.statusCode >= 200 && wrapper.statusCode < 300,
			}

			if !entry.Success {
				entry.ErrorMsg = http.StatusText(wrapper.statusCode)
			}

			logger.Log(entry)
		})
	}
}

// AuthTokenValidator interface for token validation in audit middleware
type AuthTokenValidator interface {
	ValidateToken(token string) (*AuditClaims, error)
}

// AuditClaims represents JWT claims for audit middleware
type AuditClaims struct {
	UserID string
	Email  string
}

// extractAdminInfo extracts admin ID and email from JWT token
func extractAdminInfo(r *http.Request, authService AuthTokenValidator) (string, string) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "anonymous", "anonymous"
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "anonymous", "anonymous"
	}

	token := parts[1]

	// Validate token
	if authService != nil {
		claims, err := authService.ValidateToken(token)
		if err == nil && claims != nil {
			return claims.UserID, claims.Email
		}
	}

	return "anonymous", "anonymous"
}

// mapHTTPToAction maps HTTP method + path to audit action
func mapHTTPToAction(method, path string) AuditAction {
	// Match specific patterns
	if strings.Contains(path, "/admin/users") {
		if method == "POST" {
			return ActionCreateUser
		}
		if method == "PUT" || method == "PATCH" {
			return ActionModifyUser
		}
		if strings.Contains(path, "/disable") {
			return ActionDisableUser
		}
		if strings.Contains(path, "/enable") {
			return ActionEnableUser
		}
	}

	if strings.Contains(path, "/admin/symbols") {
		if strings.Contains(path, "/toggle") {
			return ActionToggleSymbol
		}
		if strings.Contains(path, "/spread") {
			return ActionUpdateSpread
		}
		return ActionModifySymbol
	}

	if strings.Contains(path, "/admin/deposit") {
		return ActionDeposit
	}

	if strings.Contains(path, "/admin/withdraw") {
		return ActionWithdraw
	}

	if strings.Contains(path, "/admin/adjust") {
		return ActionAdjustBalance
	}

	if strings.Contains(path, "/admin/account") {
		if method == "POST" {
			return ActionCreateAccount
		}
		return ActionModifyAccount
	}

	if strings.Contains(path, "/admin/export") {
		return ActionExportData
	}

	if strings.Contains(path, "/admin/risk") {
		return ActionModifyRiskLimit
	}

	// Default: use path as action
	return AuditAction(method + "_" + strings.TrimPrefix(path, "/admin/"))
}

// extractResourceInfo extracts resource type and ID from path
func extractResourceInfo(path string) (string, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// Extract resource type from path
	// e.g., /admin/users/123 -> resource: "user", id: "123"
	// e.g., /admin/symbols/EURUSD -> resource: "symbol", id: "EURUSD"

	if len(parts) < 2 {
		return "", ""
	}

	resource := ""
	resourceID := ""

	// Find resource type
	for i, part := range parts {
		if part == "admin" && i+1 < len(parts) {
			resource = strings.TrimSuffix(parts[i+1], "s") // Remove plural 's'
			if i+2 < len(parts) && parts[i+2] != "" {
				// Check if next part looks like an ID (not an action)
				if !strings.Contains(parts[i+2], "toggle") &&
					!strings.Contains(parts[i+2], "disable") &&
					!strings.Contains(parts[i+2], "enable") &&
					!strings.Contains(parts[i+2], "spread") {
					resourceID = parts[i+2]
				}
			}
			break
		}
	}

	return resource, resourceID
}
