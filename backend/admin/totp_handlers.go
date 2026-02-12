package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/golang-jwt/jwt/v5"
)

// TOTPSetupRequest is the request for 2FA setup
type TOTPSetupRequest struct {
	AdminID int64 `json:"adminId"`
}

// TOTPSetupResponse contains QR code URL and secret for enrollment
type TOTPSetupResponse struct {
	Secret    string `json:"secret"`    // Base32 encoded secret (for manual entry)
	QRCodeURL string `json:"qrCodeUrl"` // URL for QR code generation
	Message   string `json:"message"`
}

// TOTPVerifyRequest is the request to verify TOTP code
type TOTPVerifyRequest struct {
	AdminID int64  `json:"adminId"`
	Code    string `json:"code"`
}

// TOTPEnableRequest is the request to enable 2FA after verification
type TOTPEnableRequest struct {
	AdminID int64  `json:"adminId"`
	Code    string `json:"code"` // Final verification code
}

// TOTPDisableRequest is the request to disable 2FA
type TOTPDisableRequest struct {
	AdminID int64  `json:"adminId"`
	Code    string `json:"code"` // TOTP code for confirmation
}

// TOTPValidateRequest is for validating TOTP during login
type TOTPValidateRequest struct {
	Token string `json:"token"` // Partial JWT token from first login step
	Code  string `json:"code"`  // 6-digit TOTP code
}

// BackupCodesResponse contains generated backup codes
type BackupCodesResponse struct {
	BackupCodes []string `json:"backupCodes"`
	Message     string   `json:"message"`
}

// HandleTOTPSetup generates a new TOTP secret for admin enrollment
func (h *AdminHandler) HandleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if 2FA is already enabled
	if admin.TwoFactorEnabled {
		respondError(w, "2FA is already enabled for this admin", http.StatusBadRequest)
		return
	}

	// Generate TOTP secret
	totpService := auth.NewTOTPService()
	secret, qrCodeURL, err := totpService.GenerateTOTPSecret(admin.Username)
	if err != nil {
		log.Printf("[2FA] Failed to generate TOTP secret for %s: %v", admin.Username, err)
		respondError(w, "Failed to generate 2FA secret", http.StatusInternalServerError)
		return
	}

	// Store the secret temporarily (will be enabled after verification)
	h.authService.mu.Lock()
	admin.TwoFactorSecret = secret
	h.authService.mu.Unlock()

	log.Printf("[2FA] Setup initiated for admin: %s", admin.Username)

	respondJSON(w, TOTPSetupResponse{
		Secret:    secret,
		QRCodeURL: qrCodeURL,
		Message:   "Scan QR code with Google Authenticator or Authy, then verify with a code",
	})
}

// HandleTOTPVerify verifies TOTP code during setup
func (h *AdminHandler) HandleTOTPVerify(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req TOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if secret exists
	if admin.TwoFactorSecret == "" {
		respondError(w, "2FA setup not initiated. Call /admin/2fa/setup first", http.StatusBadRequest)
		return
	}

	// Validate the code
	totpService := auth.NewTOTPService()
	valid, err := totpService.ValidateTOTPCode(admin.TwoFactorSecret, req.Code)
	if err != nil {
		log.Printf("[2FA] Verification error for %s: %v", admin.Username, err)
		respondError(w, "Verification failed", http.StatusInternalServerError)
		return
	}

	if !valid {
		respondError(w, "Invalid 2FA code", http.StatusUnauthorized)
		return
	}

	log.Printf("[2FA] Code verified successfully for admin: %s", admin.Username)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Code verified. Call /admin/2fa/enable to complete setup",
	})
}

// HandleTOTPEnable enables 2FA after successful verification
func (h *AdminHandler) HandleTOTPEnable(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req TOTPEnableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if 2FA is already enabled
	if admin.TwoFactorEnabled {
		respondError(w, "2FA is already enabled", http.StatusBadRequest)
		return
	}

	// Verify the code one more time before enabling
	totpService := auth.NewTOTPService()
	valid, err := totpService.ValidateTOTPCode(admin.TwoFactorSecret, req.Code)
	if err != nil || !valid {
		respondError(w, "Invalid 2FA code", http.StatusUnauthorized)
		return
	}

	// Enable 2FA
	h.authService.mu.Lock()
	admin.TwoFactorEnabled = true
	h.authService.mu.Unlock()

	// Generate backup codes
	backupCodes, err := auth.GenerateBackupCodes(10)
	if err != nil {
		log.Printf("[2FA] Failed to generate backup codes for %s: %v", admin.Username, err)
		respondError(w, "Failed to generate backup codes", http.StatusInternalServerError)
		return
	}

	// Store backup codes (in production, these should be hashed)
	totpStore := auth.NewTOTPStore()
	totpStore.EnableTOTP(strconv.FormatInt(admin.ID, 10), admin.TwoFactorSecret)
	totpStore.SetBackupCodes(strconv.FormatInt(admin.ID, 10), backupCodes)

	// Audit log
	h.auditLog.Log(admin.ID, admin.Username, "2FA_ENABLED", "ADMIN", admin.ID, nil, "2FA enabled", getIPAddress(r), r.UserAgent(), "SUCCESS", "")

	log.Printf("[2FA] 2FA enabled for admin: %s", admin.Username)

	respondJSON(w, BackupCodesResponse{
		BackupCodes: backupCodes,
		Message:     "2FA enabled successfully. Save these backup codes in a secure location. Each code can only be used once.",
	})
}

// HandleTOTPDisable disables 2FA for an admin
func (h *AdminHandler) HandleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req TOTPDisableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check if 2FA is enabled
	if !admin.TwoFactorEnabled {
		respondError(w, "2FA is not enabled", http.StatusBadRequest)
		return
	}

	// Verify the code before disabling
	totpService := auth.NewTOTPService()
	valid, err := totpService.ValidateTOTPCode(admin.TwoFactorSecret, req.Code)
	if err != nil || !valid {
		respondError(w, "Invalid 2FA code", http.StatusUnauthorized)
		return
	}

	// Disable 2FA
	h.authService.mu.Lock()
	admin.TwoFactorEnabled = false
	admin.TwoFactorSecret = ""
	h.authService.mu.Unlock()

	// Remove from TOTP store
	totpStore := auth.NewTOTPStore()
	totpStore.DisableTOTP(strconv.FormatInt(admin.ID, 10))

	// Audit log
	h.auditLog.Log(admin.ID, admin.Username, "2FA_DISABLED", "ADMIN", admin.ID, nil, "2FA disabled", getIPAddress(r), r.UserAgent(), "SUCCESS", "")

	log.Printf("[2FA] 2FA disabled for admin: %s", admin.Username)

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "2FA disabled successfully",
	})
}

// HandleTOTPValidate validates TOTP code during login (second step)
func (h *AdminHandler) HandleTOTPValidate(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req TOTPValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse the partial JWT token
	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("2fa_partial_token_secret"), nil // Use a different secret for partial tokens
	})

	if err != nil || !token.Valid {
		respondError(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Extract admin info from token
	adminIDFloat, ok := (*claims)["admin_id"].(float64)
	if !ok {
		respondError(w, "Invalid token format", http.StatusUnauthorized)
		return
	}
	adminID := int64(adminIDFloat)

	// Check if 2FA is required
	require2FA, ok := (*claims)["2fa_required"].(bool)
	if !ok || !require2FA {
		respondError(w, "2FA not required for this token", http.StatusBadRequest)
		return
	}

	// Get admin
	h.authService.mu.RLock()
	admin, exists := h.authService.admins[adminID]
	h.authService.mu.RUnlock()

	if !exists || !admin.TwoFactorEnabled {
		respondError(w, "Admin not found or 2FA not enabled", http.StatusUnauthorized)
		return
	}

	// Validate TOTP code
	totpService := auth.NewTOTPService()
	valid, err := totpService.ValidateTOTPCode(admin.TwoFactorSecret, req.Code)
	if err != nil || !valid {
		// Check backup codes
		totpStore := auth.NewTOTPStore()
		backupValid, _ := totpStore.ValidateBackupCode(strconv.FormatInt(adminID, 10), req.Code)
		if !backupValid {
			log.Printf("[2FA] Invalid 2FA code for admin %s from %s", admin.Username, getIPAddress(r))
			respondError(w, "Invalid 2FA code", http.StatusUnauthorized)
			return
		}
		log.Printf("[2FA] Backup code used for admin: %s", admin.Username)
	}

	// Generate full session token
	sessionID, err := generateSessionToken()
	if err != nil {
		respondError(w, "Failed to generate session", http.StatusInternalServerError)
		return
	}

	// Create session
	now := time.Now()
	session := &AdminSession{
		SessionID:  sessionID,
		AdminID:    admin.ID,
		Username:   admin.Username,
		Role:       admin.Role,
		IPAddress:  getIPAddress(r),
		UserAgent:  r.UserAgent(),
		CreatedAt:  now,
		ExpiresAt:  now.Add(8 * time.Hour),
		LastActive: now,
	}

	h.authService.mu.Lock()
	h.authService.sessions[sessionID] = session
	admin.LastLogin = now
	h.authService.mu.Unlock()

	log.Printf("[2FA] Admin logged in successfully after 2FA: %s from %s", admin.Username, getIPAddress(r))

	respondJSON(w, session)
}

// HandleBackupCodes generates new backup codes
func (h *AdminHandler) HandleBackupCodes(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate admin
	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if 2FA is enabled
	if !admin.TwoFactorEnabled {
		respondError(w, "2FA is not enabled", http.StatusBadRequest)
		return
	}

	// Generate new backup codes
	backupCodes, err := auth.GenerateBackupCodes(10)
	if err != nil {
		log.Printf("[2FA] Failed to generate backup codes for %s: %v", admin.Username, err)
		respondError(w, "Failed to generate backup codes", http.StatusInternalServerError)
		return
	}

	// Store backup codes
	totpStore := auth.NewTOTPStore()
	totpStore.SetBackupCodes(strconv.FormatInt(admin.ID, 10), backupCodes)

	// Audit log
	h.auditLog.Log(admin.ID, admin.Username, "2FA_BACKUP_CODES_GENERATED", "ADMIN", admin.ID, nil, "New backup codes generated", getIPAddress(r), r.UserAgent(), "SUCCESS", "")

	log.Printf("[2FA] New backup codes generated for admin: %s", admin.Username)

	respondJSON(w, BackupCodesResponse{
		BackupCodes: backupCodes,
		Message:     "New backup codes generated. Previous codes are now invalid. Save these in a secure location.",
	})
}
