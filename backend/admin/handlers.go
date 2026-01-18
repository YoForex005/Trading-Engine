package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/epic1st/rtx/backend/internal/core"
)

// AdminHandler provides HTTP handlers for admin operations
type AdminHandler struct {
	authService  *AuthService
	userMgmt     *UserManagementService
	fundMgmt     *FundManagementService
	orderMgmt    *OrderManagementService
	groupMgmt    *GroupManagementService
	auditLog     *AuditLog
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(engine *core.Engine) *AdminHandler {
	auditLog := NewAuditLog(10000)
	authService := NewAuthService()
	userMgmt := NewUserManagementService(engine, authService, auditLog)
	fundMgmt := NewFundManagementService(engine, auditLog)
	orderMgmt := NewOrderManagementService(engine, auditLog)
	groupMgmt := NewGroupManagementService(auditLog)

	return &AdminHandler{
		authService: authService,
		userMgmt:    userMgmt,
		fundMgmt:    fundMgmt,
		orderMgmt:   orderMgmt,
		groupMgmt:   groupMgmt,
		auditLog:    auditLog,
	}
}

// Helper functions

func cors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func (h *AdminHandler) authenticate(r *http.Request) (*Admin, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, http.ErrNoCookie
	}

	// Extract Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, http.ErrNoCookie
	}

	sessionID := parts[1]
	ipAddress := getIPAddress(r)

	admin, err := h.authService.ValidateSession(sessionID, ipAddress)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Authentication Endpoints

func (h *AdminHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	userAgent := r.UserAgent()

	session, err := h.authService.Login(req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		respondError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"session": session,
		"token":   session.SessionID,
	})
}

func (h *AdminHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	sessionID := parts[1]

	h.authService.Logout(sessionID)

	log.Printf("[Admin] %s logged out", admin.Username)

	respondJSON(w, map[string]bool{"success": true})
}

// User Management Endpoints

func (h *AdminHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "view_users") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	users, err := h.userMgmt.GetAllUsers()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, users)
}

func (h *AdminHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "view_users") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	accountID, _ := strconv.ParseInt(r.URL.Query().Get("accountId"), 10, 64)
	if accountID == 0 {
		respondError(w, "Missing accountId", http.StatusBadRequest)
		return
	}

	user, err := h.userMgmt.GetUser(accountID)
	if err != nil {
		respondError(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, user)
}

func (h *AdminHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_user") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID  int64    `json:"accountId"`
		Leverage   *float64 `json:"leverage,omitempty"`
		MarginMode *string  `json:"marginMode,omitempty"`
		GroupID    *int64   `json:"groupId,omitempty"`
		Email      *string  `json:"email,omitempty"`
		Reason     string   `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.userMgmt.UpdateUserAccount(req.AccountID, req.Leverage, req.MarginMode, req.GroupID, req.Email, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleEnableUser(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_user") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID int64  `json:"accountId"`
		Reason    string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.userMgmt.EnableUserAccount(req.AccountID, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleDisableUser(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_user") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID int64  `json:"accountId"`
		Reason    string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.userMgmt.DisableUserAccount(req.AccountID, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleResetUserPassword(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_user") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID   int64  `json:"accountId"`
		NewPassword string `json:"newPassword"`
		Reason      string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.userMgmt.ResetUserPassword(req.AccountID, req.NewPassword, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

// Fund Management Endpoints

func (h *AdminHandler) HandleDeposit(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "fund_deposit") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Method      string  `json:"method"`
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
		Reason      string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	operation, err := h.fundMgmt.Deposit(req.AccountID, req.Amount, req.Method, req.Reference, req.Description, req.Reason, admin, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"operation": operation,
	})
}

func (h *AdminHandler) HandleWithdraw(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "fund_withdraw") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Method      string  `json:"method"`
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
		Reason      string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	operation, err := h.fundMgmt.Withdraw(req.AccountID, req.Amount, req.Method, req.Reference, req.Description, req.Reason, admin, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"operation": operation,
	})
}

func (h *AdminHandler) HandleAdjust(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "fund_deposit") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		Reason      string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	operation, err := h.fundMgmt.Adjust(req.AccountID, req.Amount, req.Description, req.Reason, admin, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"operation": operation,
	})
}

func (h *AdminHandler) HandleBonus(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "fund_deposit") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		AccountID   int64   `json:"accountId"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
		Reason      string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	operation, err := h.fundMgmt.AddBonus(req.AccountID, req.Amount, req.Description, req.Reason, admin, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success":   true,
		"operation": operation,
	})
}

// Order Management Endpoints

func (h *AdminHandler) HandleGetAllOrders(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "view_orders") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	status := r.URL.Query().Get("status")
	orders, err := h.orderMgmt.GetAllOrders(status)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, orders)
}

func (h *AdminHandler) HandleGetAllPositions(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "view_orders") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	positions, err := h.orderMgmt.GetAllPositions()
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, positions)
}

func (h *AdminHandler) HandleModifyOrder(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_order") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		OrderID int64    `json:"orderId"`
		Price   *float64 `json:"price,omitempty"`
		Volume  *float64 `json:"volume,omitempty"`
		SL      *float64 `json:"sl,omitempty"`
		TP      *float64 `json:"tp,omitempty"`
		Reason  string   `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.orderMgmt.ModifyOrder(req.OrderID, req.Price, req.Volume, req.SL, req.TP, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleModifyPosition(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_order") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		PositionID int64    `json:"positionId"`
		SL         *float64 `json:"sl,omitempty"`
		TP         *float64 `json:"tp,omitempty"`
		Reason     string   `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.orderMgmt.ModifyPosition(req.PositionID, req.SL, req.TP, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleReversePosition(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_order") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		PositionID int64  `json:"positionId"`
		Reason     string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.orderMgmt.ReversePosition(req.PositionID, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "close_position") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		PositionID int64   `json:"positionId"`
		Volume     float64 `json:"volume"`
		Reason     string  `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.orderMgmt.ClosePosition(req.PositionID, req.Volume, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleDeleteOrder(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_order") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		OrderID int64  `json:"orderId"`
		Reason  string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.orderMgmt.DeleteOrder(req.OrderID, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

// Group Management Endpoints

func (h *AdminHandler) HandleGetGroups(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "view_groups") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	groups := h.groupMgmt.ListGroups()
	respondJSON(w, groups)
}

func (h *AdminHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_group") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		ExecutionMode  string   `json:"executionMode"`
		Markup         float64  `json:"markup"`
		Commission     float64  `json:"commission"`
		MaxLeverage    float64  `json:"maxLeverage"`
		EnabledSymbols []string `json:"enabledSymbols"`
		DefaultBalance float64  `json:"defaultBalance"`
		MarginMode     string   `json:"marginMode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	group, err := h.groupMgmt.CreateGroup(req.Name, req.Description, req.ExecutionMode, req.Markup, req.Commission, req.MaxLeverage, req.DefaultBalance, req.EnabledSymbols, req.MarginMode, admin, ipAddress)
	if err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"group":   group,
	})
}

func (h *AdminHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_group") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		GroupID        int64     `json:"groupId"`
		Name           *string   `json:"name,omitempty"`
		Description    *string   `json:"description,omitempty"`
		ExecutionMode  *string   `json:"executionMode,omitempty"`
		Markup         *float64  `json:"markup,omitempty"`
		Commission     *float64  `json:"commission,omitempty"`
		MaxLeverage    *float64  `json:"maxLeverage,omitempty"`
		EnabledSymbols []string  `json:"enabledSymbols,omitempty"`
		DefaultBalance *float64  `json:"defaultBalance,omitempty"`
		MarginMode     *string   `json:"marginMode,omitempty"`
		Reason         string    `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.groupMgmt.UpdateGroup(req.GroupID, req.Name, req.Description, req.ExecutionMode, req.Markup, req.Commission, req.MaxLeverage, req.DefaultBalance, req.EnabledSymbols, req.MarginMode, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func (h *AdminHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	admin, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !h.authService.CheckPermission(admin, "modify_group") {
		respondError(w, "Insufficient permissions", http.StatusForbidden)
		return
	}

	var req struct {
		GroupID int64  `json:"groupId"`
		Reason  string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ipAddress := getIPAddress(r)
	if err := h.groupMgmt.DeleteGroup(req.GroupID, admin, req.Reason, ipAddress); err != nil {
		respondError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

// Audit Trail Endpoints

func (h *AdminHandler) HandleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	cors(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	_, err := h.authenticate(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	action := r.URL.Query().Get("action")
	entityType := r.URL.Query().Get("entityType")

	var actionPtr, entityTypePtr *string
	if action != "" {
		actionPtr = &action
	}
	if entityType != "" {
		entityTypePtr = &entityType
	}

	entries := h.auditLog.GetEntries(nil, actionPtr, entityTypePtr, nil, nil, nil, limit)
	respondJSON(w, entries)
}

// RegisterRoutes registers all admin routes
func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	// Authentication
	mux.HandleFunc("/admin/auth/login", h.HandleLogin)
	mux.HandleFunc("/admin/auth/logout", h.HandleLogout)

	// User Management
	mux.HandleFunc("/admin/users", h.HandleGetUsers)
	mux.HandleFunc("/admin/user", h.HandleGetUser)
	mux.HandleFunc("/admin/user/update", h.HandleUpdateUser)
	mux.HandleFunc("/admin/user/enable", h.HandleEnableUser)
	mux.HandleFunc("/admin/user/disable", h.HandleDisableUser)
	mux.HandleFunc("/admin/user/reset-password", h.HandleResetUserPassword)

	// Fund Management
	mux.HandleFunc("/admin/fund/deposit", h.HandleDeposit)
	mux.HandleFunc("/admin/fund/withdraw", h.HandleWithdraw)
	mux.HandleFunc("/admin/fund/adjust", h.HandleAdjust)
	mux.HandleFunc("/admin/fund/bonus", h.HandleBonus)

	// Order Management
	mux.HandleFunc("/admin/orders", h.HandleGetAllOrders)
	mux.HandleFunc("/admin/positions", h.HandleGetAllPositions)
	mux.HandleFunc("/admin/order/modify", h.HandleModifyOrder)
	mux.HandleFunc("/admin/order/delete", h.HandleDeleteOrder)
	mux.HandleFunc("/admin/position/modify", h.HandleModifyPosition)
	mux.HandleFunc("/admin/position/reverse", h.HandleReversePosition)
	mux.HandleFunc("/admin/position/close", h.HandleClosePosition)

	// Group Management
	mux.HandleFunc("/admin/groups", h.HandleGetGroups)
	mux.HandleFunc("/admin/group/create", h.HandleCreateGroup)
	mux.HandleFunc("/admin/group/update", h.HandleUpdateGroup)
	mux.HandleFunc("/admin/group/delete", h.HandleDeleteGroup)

	// Audit Trail
	mux.HandleFunc("/admin/audit", h.HandleGetAuditLog)

	log.Println("[Admin] Admin routes registered")
}
