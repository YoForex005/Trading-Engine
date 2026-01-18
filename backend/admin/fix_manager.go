package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/epic1st/rtx/backend/fix"
)

// FIXManager provides admin controls for FIX API provisioning
type FIXManager struct {
	provisioning *fix.ProvisioningService
}

// NewFIXManager creates a new FIX manager
func NewFIXManager(provisioning *fix.ProvisioningService) *FIXManager {
	return &FIXManager{
		provisioning: provisioning,
	}
}

// ProvisionUserRequest represents an admin request to provision FIX access
type ProvisionUserRequest struct {
	UserID        string   `json:"user_id"`
	RateLimitTier string   `json:"rate_limit_tier"` // basic, standard, premium, unlimited
	MaxSessions   int      `json:"max_sessions"`
	ExpiresInDays int      `json:"expires_in_days,omitempty"`
	AllowedIPs    []string `json:"allowed_ips,omitempty"`
	// User context for rule evaluation
	AccountBalance float64  `json:"account_balance"`
	TradingVolume  float64  `json:"trading_volume"`
	AccountAgeDays int      `json:"account_age_days"`
	KYCLevel       int      `json:"kyc_level"`
	Groups         []string `json:"groups,omitempty"`
	BypassRules    bool     `json:"bypass_rules"` // Admin can bypass rules
}

// ProvisionUser provisions FIX API access for a user (admin operation)
func (fm *FIXManager) ProvisionUser(req *ProvisionUserRequest) (*fix.ProvisioningResponse, error) {
	// Build user context
	userContext := &fix.UserContext{
		UserID:         req.UserID,
		AccountBalance: req.AccountBalance,
		TradingVolume:  req.TradingVolume,
		AccountAge:     time.Duration(req.AccountAgeDays) * 24 * time.Hour,
		KYCLevel:       req.KYCLevel,
		Groups:         req.Groups,
		IsAdmin:        req.BypassRules,
		CustomFields:   make(map[string]interface{}),
	}

	// Build provisioning request
	var expiresIn *time.Duration
	if req.ExpiresInDays > 0 {
		duration := time.Duration(req.ExpiresInDays) * 24 * time.Hour
		expiresIn = &duration
	}

	provReq := &fix.ProvisioningRequest{
		UserID:        req.UserID,
		UserContext:   userContext,
		RateLimitTier: req.RateLimitTier,
		MaxSessions:   req.MaxSessions,
		ExpiresIn:     expiresIn,
		AllowedIPs:    req.AllowedIPs,
	}

	return fm.provisioning.ProvisionAccess(provReq), nil
}

// RevokeUserAccess revokes FIX access for a user
func (fm *FIXManager) RevokeUserAccess(userID, reason string) error {
	return fm.provisioning.RevokeUserAccess(userID, reason)
}

// SuspendUserAccess temporarily suspends FIX access
func (fm *FIXManager) SuspendUserAccess(userID, reason string) error {
	return fm.provisioning.GetCredentialStore().SuspendCredentials(userID, reason)
}

// ReactivateUserAccess reactivates suspended FIX access
func (fm *FIXManager) ReactivateUserAccess(userID string) error {
	return fm.provisioning.GetCredentialStore().ReactivateCredentials(userID)
}

// RegeneratePassword generates a new password for a user
func (fm *FIXManager) RegeneratePassword(userID string) (string, error) {
	return fm.provisioning.GetCredentialStore().RegeneratePassword(userID)
}

// UpdateRateLimitTier changes a user's rate limit tier
func (fm *FIXManager) UpdateRateLimitTier(userID, newTier string) error {
	return fm.provisioning.GetRateLimiter().UpdateUserTier(userID, newTier)
}

// GetUserCredentials retrieves credentials for a user
func (fm *FIXManager) GetUserCredentials(userID string) (*fix.FIXCredentials, error) {
	return fm.provisioning.GetCredentialStore().GetCredentials(userID)
}

// ListAllCredentials lists all FIX credentials
func (fm *FIXManager) ListAllCredentials() []*fix.FIXCredentials {
	return fm.provisioning.GetCredentialStore().ListAllCredentials()
}

// GetActiveSessions returns all active FIX sessions
func (fm *FIXManager) GetActiveSessions() map[string][]*fix.ActiveFIXSession {
	return fm.provisioning.GetAllSessions()
}

// GetUserSessions returns active sessions for a specific user
func (fm *FIXManager) GetUserSessions(userID string) []*fix.ActiveFIXSession {
	return fm.provisioning.GetUserSessions(userID)
}

// KillSession forcefully terminates a FIX session
func (fm *FIXManager) KillSession(sessionID string) error {
	return fm.provisioning.KillSession(sessionID)
}

// GetUserRateLimitState returns rate limit state for a user
func (fm *FIXManager) GetUserRateLimitState(userID string) (*fix.UserRateLimitState, error) {
	return fm.provisioning.GetRateLimiter().GetUserState(userID)
}

// GetAllRateLimitStates returns rate limit states for all users
func (fm *FIXManager) GetAllRateLimitStates() []*fix.UserRateLimitState {
	return fm.provisioning.GetRateLimiter().GetAllUserStates()
}

// ResetUserViolations resets rate limit violations for a user
func (fm *FIXManager) ResetUserViolations(userID string) error {
	return fm.provisioning.GetRateLimiter().ResetViolations(userID)
}

// AddAccessRule adds a new access rule
func (fm *FIXManager) AddAccessRule(rule *fix.AccessRule) error {
	return fm.provisioning.GetRulesEngine().AddRule(rule)
}

// RemoveAccessRule removes an access rule
func (fm *FIXManager) RemoveAccessRule(ruleID string) error {
	return fm.provisioning.GetRulesEngine().RemoveRule(ruleID)
}

// EnableAccessRule enables an access rule
func (fm *FIXManager) EnableAccessRule(ruleID string) error {
	return fm.provisioning.GetRulesEngine().EnableRule(ruleID)
}

// DisableAccessRule disables an access rule
func (fm *FIXManager) DisableAccessRule(ruleID string) error {
	return fm.provisioning.GetRulesEngine().DisableRule(ruleID)
}

// ListAccessRules lists all access rules
func (fm *FIXManager) ListAccessRules() []*fix.AccessRule {
	return fm.provisioning.GetRulesEngine().ListRules()
}

// SetUserRules sets custom rules for a specific user
func (fm *FIXManager) SetUserRules(userID string, ruleIDs []string) error {
	return fm.provisioning.GetRulesEngine().SetUserRules(userID, ruleIDs)
}

// GetUserRules gets custom rules for a specific user
func (fm *FIXManager) GetUserRules(userID string) []string {
	return fm.provisioning.GetRulesEngine().GetUserRules(userID)
}

// AddRateLimitTier adds a custom rate limit tier
func (fm *FIXManager) AddRateLimitTier(tier *fix.RateLimitTier) error {
	return fm.provisioning.GetRateLimiter().AddCustomTier(tier)
}

// ListRateLimitTiers lists all available rate limit tiers
func (fm *FIXManager) ListRateLimitTiers() []*fix.RateLimitTier {
	return fm.provisioning.GetRateLimiter().ListTiers()
}

// SystemStats provides overall system statistics
type SystemStats struct {
	TotalUsers       int                          `json:"total_users"`
	ActiveUsers      int                          `json:"active_users"`
	TotalSessions    int                          `json:"total_sessions"`
	CredentialStats  CredentialStats              `json:"credential_stats"`
	SessionsByUser   map[string]int               `json:"sessions_by_user"`
	RateLimitTiers   map[string]int               `json:"rate_limit_tiers"`
}

// CredentialStats provides credential statistics
type CredentialStats struct {
	Active    int `json:"active"`
	Revoked   int `json:"revoked"`
	Suspended int `json:"suspended"`
	Expired   int `json:"expired"`
}

// GetSystemStats returns overall system statistics
func (fm *FIXManager) GetSystemStats() *SystemStats {
	allCreds := fm.provisioning.GetCredentialStore().ListAllCredentials()
	allSessions := fm.provisioning.GetAllSessions()

	stats := &SystemStats{
		TotalUsers:     len(allCreds),
		SessionsByUser: make(map[string]int),
		RateLimitTiers: make(map[string]int),
		CredentialStats: CredentialStats{},
	}

	// Count credentials by status
	for _, cred := range allCreds {
		switch cred.Status {
		case fix.CredentialStatusActive:
			stats.CredentialStats.Active++
		case fix.CredentialStatusRevoked:
			stats.CredentialStats.Revoked++
		case fix.CredentialStatusSuspended:
			stats.CredentialStats.Suspended++
		case fix.CredentialStatusExpired:
			stats.CredentialStats.Expired++
		}

		// Count by tier
		stats.RateLimitTiers[cred.RateLimitTier]++
	}

	// Count sessions
	for userID, sessions := range allSessions {
		sessionCount := len(sessions)
		stats.TotalSessions += sessionCount
		stats.SessionsByUser[userID] = sessionCount
		if sessionCount > 0 {
			stats.ActiveUsers++
		}
	}

	return stats
}

// ===== HTTP Handlers =====

// RegisterHTTPHandlers registers HTTP handlers for the FIX manager
func (fm *FIXManager) RegisterHTTPHandlers(mux *http.ServeMux) {
	// User provisioning
	mux.HandleFunc("/admin/fix/provision", fm.handleProvisionUser)
	mux.HandleFunc("/admin/fix/revoke", fm.handleRevokeUser)
	mux.HandleFunc("/admin/fix/suspend", fm.handleSuspendUser)
	mux.HandleFunc("/admin/fix/reactivate", fm.handleReactivateUser)
	mux.HandleFunc("/admin/fix/regenerate-password", fm.handleRegeneratePassword)

	// Credentials
	mux.HandleFunc("/admin/fix/credentials", fm.handleListCredentials)
	mux.HandleFunc("/admin/fix/credentials/user", fm.handleGetUserCredentials)

	// Sessions
	mux.HandleFunc("/admin/fix/sessions", fm.handleListSessions)
	mux.HandleFunc("/admin/fix/sessions/user", fm.handleGetUserSessions)
	mux.HandleFunc("/admin/fix/sessions/kill", fm.handleKillSession)

	// Rate limiting
	mux.HandleFunc("/admin/fix/rate-limits", fm.handleListRateLimits)
	mux.HandleFunc("/admin/fix/rate-limits/update", fm.handleUpdateRateLimit)
	mux.HandleFunc("/admin/fix/rate-limits/reset-violations", fm.handleResetViolations)
	mux.HandleFunc("/admin/fix/rate-limits/tiers", fm.handleListTiers)

	// Rules
	mux.HandleFunc("/admin/fix/rules", fm.handleListRules)
	mux.HandleFunc("/admin/fix/rules/add", fm.handleAddRule)
	mux.HandleFunc("/admin/fix/rules/remove", fm.handleRemoveRule)
	mux.HandleFunc("/admin/fix/rules/enable", fm.handleEnableRule)
	mux.HandleFunc("/admin/fix/rules/disable", fm.handleDisableRule)

	// Stats
	mux.HandleFunc("/admin/fix/stats", fm.handleGetStats)
}

func (fm *FIXManager) handleProvisionUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ProvisionUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := fm.ProvisionUser(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (fm *FIXManager) handleRevokeUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	reason := r.URL.Query().Get("reason")

	if err := fm.RevokeUserAccess(userID, reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s access revoked", userID)
}

func (fm *FIXManager) handleSuspendUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	reason := r.URL.Query().Get("reason")

	if err := fm.SuspendUserAccess(userID, reason); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s access suspended", userID)
}

func (fm *FIXManager) handleReactivateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	if err := fm.ReactivateUserAccess(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s access reactivated", userID)
}

func (fm *FIXManager) handleRegeneratePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	newPassword, err := fm.RegeneratePassword(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"user_id":      userID,
		"new_password": newPassword,
	})
}

func (fm *FIXManager) handleListCredentials(w http.ResponseWriter, r *http.Request) {
	credentials := fm.ListAllCredentials()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(credentials)
}

func (fm *FIXManager) handleGetUserCredentials(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	creds, err := fm.GetUserCredentials(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(creds)
}

func (fm *FIXManager) handleListSessions(w http.ResponseWriter, r *http.Request) {
	sessions := fm.GetActiveSessions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (fm *FIXManager) handleGetUserSessions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	sessions := fm.GetUserSessions(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (fm *FIXManager) handleKillSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")

	if err := fm.KillSession(sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Session %s killed", sessionID)
}

func (fm *FIXManager) handleListRateLimits(w http.ResponseWriter, r *http.Request) {
	states := fm.GetAllRateLimitStates()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(states)
}

func (fm *FIXManager) handleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	newTier := r.URL.Query().Get("tier")

	if err := fm.UpdateRateLimitTier(userID, newTier); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User %s rate limit updated to %s", userID, newTier)
}

func (fm *FIXManager) handleResetViolations(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")

	if err := fm.ResetUserViolations(userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Violations reset for user %s", userID)
}

func (fm *FIXManager) handleListTiers(w http.ResponseWriter, r *http.Request) {
	tiers := fm.ListRateLimitTiers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tiers)
}

func (fm *FIXManager) handleListRules(w http.ResponseWriter, r *http.Request) {
	rules := fm.ListAccessRules()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func (fm *FIXManager) handleAddRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rule fix.AccessRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := fm.AddAccessRule(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (fm *FIXManager) handleRemoveRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("rule_id")

	if err := fm.RemoveAccessRule(ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Rule %s removed", ruleID)
}

func (fm *FIXManager) handleEnableRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("rule_id")

	if err := fm.EnableAccessRule(ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Rule %s enabled", ruleID)
}

func (fm *FIXManager) handleDisableRule(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("rule_id")

	if err := fm.DisableAccessRule(ruleID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Rule %s disabled", ruleID)
}

func (fm *FIXManager) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := fm.GetSystemStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
