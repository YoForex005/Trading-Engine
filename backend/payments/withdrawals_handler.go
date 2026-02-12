package payments

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// WithdrawalsHandler handles admin withdrawal approval endpoints
type WithdrawalsHandler struct {
	approvalService *WithdrawalApprovalService
	repository      Repository
}

// NewWithdrawalsHandler creates a new withdrawals handler
func NewWithdrawalsHandler(approvalService *WithdrawalApprovalService, repository Repository) *WithdrawalsHandler {
	return &WithdrawalsHandler{
		approvalService: approvalService,
		repository:      repository,
	}
}

// HandleGetPending handles GET /admin/withdrawals/pending
func (h *WithdrawalsHandler) HandleGetPending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pending := h.approvalService.GetPendingRequests()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending": pending,
		"count":   len(pending),
	})
}

// HandleGetHistory handles GET /admin/withdrawals/history
func (h *WithdrawalsHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var status ApprovalStatus
	if statusStr != "" {
		status = ApprovalStatus(strings.ToUpper(statusStr))
	}

	limit := 20 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	history := h.approvalService.GetHistory(status, limit, offset)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"count":   len(history),
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleGetDetail handles GET /admin/withdrawals/{id}
func (h *WithdrawalsHandler) HandleGetDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	txID := strings.TrimPrefix(r.URL.Path, "/admin/withdrawals/")
	if txID == "" || txID == "pending" || txID == "history" {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}

	request, err := h.approvalService.GetRequest(txID)
	if err != nil {
		if err == ErrTransactionNotFound {
			http.Error(w, "Withdrawal request not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

// HandleApprove handles POST /admin/withdrawals/{id}/approve
func (h *WithdrawalsHandler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/withdrawals/")
	txID := strings.TrimSuffix(path, "/approve")

	if txID == "" {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}

	// TODO: Extract admin ID from JWT token
	// For now, use placeholder
	adminID := "admin-user" // Should come from auth middleware

	// Approve the withdrawal
	if err := h.approvalService.Approve(txID, adminID); err != nil {
		log.Printf("[Withdrawals] Approval failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the transaction and process payout
	tx, err := h.repository.GetTransaction(context.Background(), txID)
	if err != nil {
		log.Printf("[Withdrawals] Failed to get transaction after approval: %v", err)
		http.Error(w, "Approval recorded but payout initiation failed", http.StatusInternalServerError)
		return
	}

	// Update transaction status to processing
	tx.Status = StatusPending
	h.repository.UpdateStatus(txID, StatusPending)

	// NOTE: Actual payout processing should be handled by the gateway's provider
	// For now, we just mark as approved. The gateway will handle the payout.

	log.Printf("[Withdrawals] Withdrawal %s approved by %s", txID, adminID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Withdrawal approved successfully",
		"txId":    txID,
	})
}

// HandleReject handles POST /admin/withdrawals/{id}/reject
func (h *WithdrawalsHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/withdrawals/")
	txID := strings.TrimSuffix(path, "/reject")

	if txID == "" {
		http.Error(w, "Transaction ID required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Reason == "" {
		http.Error(w, "Rejection reason is required", http.StatusBadRequest)
		return
	}

	// TODO: Extract admin ID from JWT token
	adminID := "admin-user" // Should come from auth middleware

	// Reject the withdrawal
	if err := h.approvalService.Reject(txID, adminID, req.Reason); err != nil {
		log.Printf("[Withdrawals] Rejection failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update transaction status to rejected
	tx, err := h.repository.GetTransaction(context.Background(), txID)
	if err == nil {
		tx.Status = StatusCancelled
		tx.FailureReason = req.Reason
		h.repository.UpdateStatus(txID, StatusCancelled)
	}

	log.Printf("[Withdrawals] Withdrawal %s rejected by %s: %s", txID, adminID, req.Reason)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Withdrawal rejected successfully",
		"txId":    txID,
		"reason":  req.Reason,
	})
}

// RegisterRoutes registers all withdrawal admin routes
func (h *WithdrawalsHandler) RegisterRoutes(mux *http.ServeMux, corsMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// GET /admin/withdrawals/pending
	mux.HandleFunc("/admin/withdrawals/pending", corsMiddleware(h.HandleGetPending))

	// GET /admin/withdrawals/history
	mux.HandleFunc("/admin/withdrawals/history", corsMiddleware(h.HandleGetHistory))

	// GET /admin/withdrawals/{id}
	// POST /admin/withdrawals/{id}/approve
	// POST /admin/withdrawals/{id}/reject
	mux.HandleFunc("/admin/withdrawals/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if strings.HasSuffix(path, "/approve") {
			h.HandleApprove(w, r)
		} else if strings.HasSuffix(path, "/reject") {
			h.HandleReject(w, r)
		} else {
			// Detail endpoint
			h.HandleGetDetail(w, r)
		}
	}))

	log.Println("[Withdrawals] Admin withdrawal approval routes registered")
}
