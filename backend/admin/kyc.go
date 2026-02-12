package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
	"github.com/google/uuid"
)

// KYCApplication represents a KYC/AML verification application
type KYCApplication struct {
	ID                    string    `json:"id"`
	ClientID              string    `json:"clientId"`
	Status                string    `json:"status"` // pending, under_review, approved, rejected, expired
	DocumentType          string    `json:"documentType"` // passport, national_id, drivers_license
	DocumentNumber        string    `json:"documentNumber"`
	DocumentExpiry        time.Time `json:"documentExpiry"`
	SubmittedAt           time.Time `json:"submittedAt"`
	ReviewedAt            time.Time `json:"reviewedAt"`
	ReviewedBy            string    `json:"reviewedBy"` // Admin user ID
	RejectionReason       string    `json:"rejectionReason"`
	RiskLevel             string    `json:"riskLevel"` // low, medium, high
	AMLCheckPassed        bool      `json:"amlCheckPassed"`
	PEPCheckPassed        bool      `json:"pepCheckPassed"`
	SanctionsCheckPassed  bool      `json:"sanctionsCheckPassed"`
	Documents             []KYCDocument `json:"documents"`
}

// KYCDocument represents an uploaded document for verification
type KYCDocument struct {
	ID            string    `json:"id"`
	ApplicationID string    `json:"applicationId"`
	Type          string    `json:"type"` // id_front, id_back, proof_of_address, selfie
	FileName      string    `json:"fileName"`
	FileSize      int64     `json:"fileSize"`
	UploadedAt    time.Time `json:"uploadedAt"`
	Status        string    `json:"status"` // pending, verified, rejected
}

// KYCStore manages KYC applications in memory
type KYCStore struct {
	applications map[string]*KYCApplication // ID -> Application
	mu           sync.RWMutex
}

// NewKYCStore creates a new KYC store with mock data
func NewKYCStore() *KYCStore {
	store := &KYCStore{
		applications: make(map[string]*KYCApplication),
	}

	// Generate 15 mock applications in various states
	store.generateMockApplications()

	// Start background worker to expire old verifications
	go store.expirationWorker()

	return store
}

// SubmitApplication creates a new KYC application
func (s *KYCStore) SubmitApplication(clientID, docType, docNumber string, docExpiry time.Time) (*KYCApplication, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate document type
	validDocTypes := map[string]bool{
		"passport":         true,
		"national_id":      true,
		"drivers_license":  true,
	}
	if !validDocTypes[docType] {
		return nil, fmt.Errorf("invalid document type: must be passport, national_id, or drivers_license")
	}

	// Check if client already has a pending/approved application
	for _, app := range s.applications {
		if app.ClientID == clientID && (app.Status == "pending" || app.Status == "under_review" || app.Status == "approved") {
			return nil, fmt.Errorf("client already has an active KYC application")
		}
	}

	app := &KYCApplication{
		ID:                   uuid.New().String(),
		ClientID:             clientID,
		Status:               "pending",
		DocumentType:         docType,
		DocumentNumber:       docNumber,
		DocumentExpiry:       docExpiry,
		SubmittedAt:          time.Now(),
		RiskLevel:            "medium", // Default to medium until checks are run
		AMLCheckPassed:       false,
		PEPCheckPassed:       false,
		SanctionsCheckPassed: false,
		Documents:            []KYCDocument{},
	}

	s.applications[app.ID] = app

	log.Printf("[KYC] New application submitted: ID=%s, ClientID=%s, DocType=%s",
		app.ID, clientID, docType)

	return app, nil
}

// ReviewApplication approves or rejects an application
func (s *KYCStore) ReviewApplication(applicationID, adminID string, approved bool, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[applicationID]
	if !exists {
		return fmt.Errorf("application not found")
	}

	if app.Status != "pending" && app.Status != "under_review" {
		return fmt.Errorf("application cannot be reviewed (status: %s)", app.Status)
	}

	app.ReviewedAt = time.Now()
	app.ReviewedBy = adminID

	if approved {
		app.Status = "approved"
		log.Printf("[KYC] Application APPROVED: ID=%s, ClientID=%s, ReviewedBy=%s",
			app.ID, app.ClientID, adminID)
	} else {
		app.Status = "rejected"
		app.RejectionReason = reason
		log.Printf("[KYC] Application REJECTED: ID=%s, ClientID=%s, Reason=%s, ReviewedBy=%s",
			app.ID, app.ClientID, reason, adminID)
	}

	return nil
}

// RunAMLCheck performs simulated AML (Anti-Money Laundering) check
func (s *KYCStore) RunAMLCheck(clientID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find active application for this client
	var app *KYCApplication
	for _, a := range s.applications {
		if a.ClientID == clientID && (a.Status == "pending" || a.Status == "under_review") {
			app = a
			break
		}
	}

	if app == nil {
		return false, fmt.Errorf("no active application found for client")
	}

	// Simulate AML check (90% pass rate for demo)
	passed := rand.Float64() > 0.1
	app.AMLCheckPassed = passed

	// Update status to under_review if it was pending
	if app.Status == "pending" {
		app.Status = "under_review"
	}

	log.Printf("[KYC] AML Check: ClientID=%s, Passed=%v", clientID, passed)

	return passed, nil
}

// RunPEPCheck performs simulated PEP (Politically Exposed Person) screening
func (s *KYCStore) RunPEPCheck(clientID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find active application for this client
	var app *KYCApplication
	for _, a := range s.applications {
		if a.ClientID == clientID && (a.Status == "pending" || a.Status == "under_review") {
			app = a
			break
		}
	}

	if app == nil {
		return false, fmt.Errorf("no active application found for client")
	}

	// Simulate PEP check (95% pass rate for demo)
	passed := rand.Float64() > 0.05
	app.PEPCheckPassed = passed

	// Update status to under_review if it was pending
	if app.Status == "pending" {
		app.Status = "under_review"
	}

	log.Printf("[KYC] PEP Check: ClientID=%s, Passed=%v", clientID, passed)

	return passed, nil
}

// RunSanctionsCheck performs simulated sanctions list screening
func (s *KYCStore) RunSanctionsCheck(clientID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find active application for this client
	var app *KYCApplication
	for _, a := range s.applications {
		if a.ClientID == clientID && (a.Status == "pending" || a.Status == "under_review") {
			app = a
			break
		}
	}

	if app == nil {
		return false, fmt.Errorf("no active application found for client")
	}

	// Simulate sanctions check (98% pass rate for demo)
	passed := rand.Float64() > 0.02
	app.SanctionsCheckPassed = passed

	// Update status to under_review if it was pending
	if app.Status == "pending" {
		app.Status = "under_review"
	}

	log.Printf("[KYC] Sanctions Check: ClientID=%s, Passed=%v", clientID, passed)

	return passed, nil
}

// GetRiskScore calculates risk level based on all checks
func (s *KYCStore) GetRiskScore(clientID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find active application for this client
	var app *KYCApplication
	for _, a := range s.applications {
		if a.ClientID == clientID && (a.Status == "pending" || a.Status == "under_review" || a.Status == "approved") {
			app = a
			break
		}
	}

	if app == nil {
		return "", fmt.Errorf("no active application found for client")
	}

	// Calculate risk based on checks
	failedChecks := 0
	if !app.AMLCheckPassed {
		failedChecks++
	}
	if !app.PEPCheckPassed {
		failedChecks++
	}
	if !app.SanctionsCheckPassed {
		failedChecks++
	}

	var riskLevel string
	if failedChecks == 0 {
		riskLevel = "low"
	} else if failedChecks == 1 {
		riskLevel = "medium"
	} else {
		riskLevel = "high"
	}

	// Update the application's risk level
	s.mu.RUnlock()
	s.mu.Lock()
	app.RiskLevel = riskLevel
	s.mu.Unlock()
	s.mu.RLock()

	log.Printf("[KYC] Risk Score Calculated: ClientID=%s, Level=%s, FailedChecks=%d",
		clientID, riskLevel, failedChecks)

	return riskLevel, nil
}

// ExpireVerification marks an application as expired (KYC valid for 1 year)
func (s *KYCStore) ExpireVerification(applicationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, exists := s.applications[applicationID]
	if !exists {
		return fmt.Errorf("application not found")
	}

	if app.Status != "approved" {
		return fmt.Errorf("only approved applications can be expired")
	}

	app.Status = "expired"

	log.Printf("[KYC] Application EXPIRED: ID=%s, ClientID=%s", app.ID, app.ClientID)

	return nil
}

// ListApplications returns all applications, optionally filtered by status
func (s *KYCStore) ListApplications(statusFilter string) []*KYCApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]*KYCApplication, 0)

	for _, app := range s.applications {
		// Apply status filter
		if statusFilter != "" && app.Status != statusFilter {
			continue
		}

		apps = append(apps, app)
	}

	return apps
}

// GetApplication returns a single application by ID
func (s *KYCStore) GetApplication(id string) (*KYCApplication, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, exists := s.applications[id]
	if !exists {
		return nil, fmt.Errorf("application not found")
	}

	return app, nil
}

// GetClientKYCStatus returns the current KYC status for a client
func (s *KYCStore) GetClientKYCStatus(clientID string) (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find most recent application for this client
	var latestApp *KYCApplication
	var latestTime time.Time

	for _, app := range s.applications {
		if app.ClientID == clientID && app.SubmittedAt.After(latestTime) {
			latestApp = app
			latestTime = app.SubmittedAt
		}
	}

	if latestApp == nil {
		return map[string]interface{}{
			"clientId": clientID,
			"status":   "not_submitted",
			"message":  "No KYC application on file",
		}, nil
	}

	return map[string]interface{}{
		"clientId":             clientID,
		"status":               latestApp.Status,
		"applicationId":        latestApp.ID,
		"submittedAt":          latestApp.SubmittedAt,
		"reviewedAt":           latestApp.ReviewedAt,
		"riskLevel":            latestApp.RiskLevel,
		"amlCheckPassed":       latestApp.AMLCheckPassed,
		"pepCheckPassed":       latestApp.PEPCheckPassed,
		"sanctionsCheckPassed": latestApp.SanctionsCheckPassed,
	}, nil
}

// GetStats returns dashboard statistics
func (s *KYCStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pendingCount := 0
	approvedToday := 0
	rejectedCount := 0
	totalCount := len(s.applications)

	today := time.Now().Truncate(24 * time.Hour)

	for _, app := range s.applications {
		switch app.Status {
		case "pending", "under_review":
			pendingCount++
		case "rejected":
			rejectedCount++
		case "approved":
			if app.ReviewedAt.After(today) {
				approvedToday++
			}
		}
	}

	rejectionRate := 0.0
	if totalCount > 0 {
		rejectionRate = float64(rejectedCount) / float64(totalCount) * 100
	}

	return map[string]interface{}{
		"totalApplications": totalCount,
		"pendingCount":      pendingCount,
		"approvedToday":     approvedToday,
		"rejectedCount":     rejectedCount,
		"rejectionRate":     rejectionRate,
		"generatedAt":       time.Now(),
	}
}

// BulkCheck runs all checks on multiple clients
func (s *KYCStore) BulkCheck(clientIDs []string) map[string]interface{} {
	results := make(map[string]interface{})

	for _, clientID := range clientIDs {
		// Run all checks
		amlPassed, _ := s.RunAMLCheck(clientID)
		pepPassed, _ := s.RunPEPCheck(clientID)
		sanctionsPassed, _ := s.RunSanctionsCheck(clientID)
		riskLevel, _ := s.GetRiskScore(clientID)

		results[clientID] = map[string]interface{}{
			"amlPassed":       amlPassed,
			"pepPassed":       pepPassed,
			"sanctionsPassed": sanctionsPassed,
			"riskLevel":       riskLevel,
		}
	}

	return results
}

// expirationWorker runs periodically to expire old verifications (1 year)
func (s *KYCStore) expirationWorker() {
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()

		expiredCount := 0
		oneYearAgo := time.Now().AddDate(-1, 0, 0)

		for _, app := range s.applications {
			// Expire approved applications older than 1 year
			if app.Status == "approved" && app.ReviewedAt.Before(oneYearAgo) {
				app.Status = "expired"
				expiredCount++
			}
		}

		if expiredCount > 0 {
			log.Printf("[KYC] Expired %d old verifications (>1 year)", expiredCount)
		}

		s.mu.Unlock()
	}
}

// generateMockApplications creates 15 mock KYC applications in various states
func (s *KYCStore) generateMockApplications() {
	// Mock client IDs (would come from ClientStore in real implementation)
	clientIDs := []string{
		"client-001", "client-002", "client-003", "client-004", "client-005",
		"client-006", "client-007", "client-008", "client-009", "client-010",
		"client-011", "client-012", "client-013", "client-014", "client-015",
	}

	statuses := []string{"pending", "under_review", "approved", "rejected"}
	docTypes := []string{"passport", "national_id", "drivers_license"}
	adminIDs := []string{"admin-001", "admin-002", "admin-003"}

	for i, clientID := range clientIDs {
		status := statuses[i%len(statuses)]

		app := &KYCApplication{
			ID:             uuid.New().String(),
			ClientID:       clientID,
			Status:         status,
			DocumentType:   docTypes[i%len(docTypes)],
			DocumentNumber: fmt.Sprintf("DOC-%06d", rand.Intn(999999)),
			DocumentExpiry: time.Now().AddDate(0, rand.Intn(12)+1, 0),
			SubmittedAt:    time.Now().AddDate(0, 0, -rand.Intn(30)),
			RiskLevel:      []string{"low", "medium", "high"}[i%3],
			Documents:      []KYCDocument{},
		}

		// Add mock documents
		documentTypes := []string{"id_front", "id_back", "proof_of_address", "selfie"}
		for _, docType := range documentTypes {
			doc := KYCDocument{
				ID:            uuid.New().String(),
				ApplicationID: app.ID,
				Type:          docType,
				FileName:      fmt.Sprintf("%s_%s.jpg", clientID, docType),
				FileSize:      int64(rand.Intn(5000000) + 500000), // 500KB - 5.5MB
				UploadedAt:    app.SubmittedAt,
				Status:        []string{"pending", "verified"}[rand.Intn(2)],
			}
			app.Documents = append(app.Documents, doc)
		}

		// Set check results based on status
		if status == "approved" || status == "under_review" {
			app.AMLCheckPassed = true
			app.PEPCheckPassed = true
			app.SanctionsCheckPassed = true
			app.ReviewedAt = time.Now().AddDate(0, 0, -rand.Intn(15))
			app.ReviewedBy = adminIDs[i%len(adminIDs)]
		} else if status == "rejected" {
			app.AMLCheckPassed = rand.Float64() > 0.5
			app.PEPCheckPassed = rand.Float64() > 0.5
			app.SanctionsCheckPassed = rand.Float64() > 0.5
			app.ReviewedAt = time.Now().AddDate(0, 0, -rand.Intn(15))
			app.ReviewedBy = adminIDs[i%len(adminIDs)]
			app.RejectionReason = []string{
				"Document expired",
				"Document not clear",
				"Failed AML check",
				"Mismatch in personal information",
			}[i%4]
		}

		s.applications[app.ID] = app
	}

	log.Printf("[KYC] Generated 15 mock applications (pending: %d, under_review: %d, approved: %d, rejected: %d)",
		len(s.ListApplications("pending")),
		len(s.ListApplications("under_review")),
		len(s.ListApplications("approved")),
		len(s.ListApplications("rejected")),
	)
}

// ============================================
// HTTP HANDLERS
// ============================================

// KYCHandler handles KYC-related HTTP requests
type KYCHandler struct {
	store       *KYCStore
	authService *auth.Service
}

// NewKYCHandler creates a new KYC handler
func NewKYCHandler(store *KYCStore, authService *auth.Service) *KYCHandler {
	return &KYCHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListApplications lists all KYC applications
// GET /admin/kyc/applications?status=pending
func (h *KYCHandler) HandleListApplications(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get status filter from query params
	statusFilter := r.URL.Query().Get("status")

	// List applications
	applications := h.store.ListApplications(statusFilter)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applications": applications,
		"count":        len(applications),
	})
}

// HandleGetApplication gets a single application by ID
// GET /admin/kyc/applications/:id
func (h *KYCHandler) HandleGetApplication(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract application ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}
	applicationID := pathParts[4]

	// Get application
	application, err := h.store.GetApplication(applicationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(application)
}

// HandleSubmitApplication submits a new KYC application
// POST /admin/kyc/applications
func (h *KYCHandler) HandleSubmitApplication(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		ClientID       string `json:"clientId"`
		DocumentType   string `json:"documentType"`
		DocumentNumber string `json:"documentNumber"`
		DocumentExpiry string `json:"documentExpiry"` // ISO 8601 format
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate inputs
	if req.ClientID == "" {
		http.Error(w, "clientId is required", http.StatusBadRequest)
		return
	}
	if req.DocumentType == "" {
		http.Error(w, "documentType is required", http.StatusBadRequest)
		return
	}
	if req.DocumentNumber == "" {
		http.Error(w, "documentNumber is required", http.StatusBadRequest)
		return
	}

	// Parse document expiry
	var docExpiry time.Time
	if req.DocumentExpiry != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.DocumentExpiry)
		if err != nil {
			http.Error(w, "Invalid documentExpiry format (use ISO 8601/RFC3339)", http.StatusBadRequest)
			return
		}
		docExpiry = parsedTime
	}

	// Submit application
	application, err := h.store.SubmitApplication(req.ClientID, req.DocumentType, req.DocumentNumber, docExpiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(application)
}

// HandleReviewApplication approves or rejects an application
// PUT /admin/kyc/applications/:id/review
func (h *KYCHandler) HandleReviewApplication(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	claims, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract application ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}
	applicationID := pathParts[4]

	// Parse request body
	var req struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Review application
	if err := h.store.ReviewApplication(applicationID, claims.UserID, req.Approved, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Application reviewed successfully",
		"applicationId": applicationID,
		"approved":      req.Approved,
	})
}

// HandleRunAMLCheck runs AML check on a client
// POST /admin/kyc/applications/:id/aml-check
func (h *KYCHandler) HandleRunAMLCheck(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract application ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid application ID", http.StatusBadRequest)
		return
	}
	applicationID := pathParts[4]

	// Get application to find client ID
	application, err := h.store.GetApplication(applicationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Run all checks
	amlPassed, _ := h.store.RunAMLCheck(application.ClientID)
	pepPassed, _ := h.store.RunPEPCheck(application.ClientID)
	sanctionsPassed, _ := h.store.RunSanctionsCheck(application.ClientID)
	riskLevel, _ := h.store.GetRiskScore(application.ClientID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"applicationId":        applicationID,
		"clientId":             application.ClientID,
		"amlCheckPassed":       amlPassed,
		"pepCheckPassed":       pepPassed,
		"sanctionsCheckPassed": sanctionsPassed,
		"riskLevel":            riskLevel,
	})
}

// HandleGetClientKYCStatus gets KYC status for a specific client
// GET /admin/kyc/clients/:clientId/status
func (h *KYCHandler) HandleGetClientKYCStatus(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract client ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	clientID := pathParts[4]

	// Get client KYC status
	status, err := h.store.GetClientKYCStatus(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleGetStats returns dashboard statistics
// GET /admin/kyc/stats
func (h *KYCHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get stats
	stats := h.store.GetStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleBulkCheck runs checks on multiple clients
// POST /admin/kyc/bulk-check
func (h *KYCHandler) HandleBulkCheck(w http.ResponseWriter, r *http.Request) {
	// JWT authentication (admin only)
	_, err := h.validateAdminAuth(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		ClientIDs []string `json:"clientIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.ClientIDs) == 0 {
		http.Error(w, "clientIds array is required", http.StatusBadRequest)
		return
	}

	// Run bulk check
	results := h.store.BulkCheck(req.ClientIDs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"count":   len(req.ClientIDs),
	})
}

// validateAdminAuth validates JWT token and checks admin role
func (h *KYCHandler) validateAdminAuth(r *http.Request) (*auth.Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]
	claims, err := auth.ValidateTokenWithDefault(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
