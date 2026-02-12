package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Client Document / KYC Management
// ============================================

// Document types
const (
	DocTypePassport            = "passport"
	DocTypeNationalID          = "national_id"
	DocTypeDriversLicense      = "drivers_license"
	DocTypeUtilityBill         = "utility_bill"
	DocTypeBankStatement       = "bank_statement"
	DocTypeSelfie              = "selfie"
	DocTypeProofOfAddress      = "proof_of_address"
	DocTypeTaxCertificate      = "tax_certificate"
	DocTypeCompanyRegistration = "company_registration"
	DocTypePowerOfAttorney     = "power_of_attorney"
)

// Document statuses
const (
	DocStatusPending              = "pending"
	DocStatusApproved             = "approved"
	DocStatusRejected             = "rejected"
	DocStatusExpired              = "expired"
	DocStatusResubmissionRequired = "resubmission_required"
)

type Document struct {
	ID              int64      `json:"id"`
	ClientID        int64      `json:"clientId"`
	ClientName      string     `json:"clientName"`
	Type            string     `json:"type"`
	FileName        string     `json:"fileName"`
	FileSize        int64      `json:"fileSize"` // bytes
	MimeType        string     `json:"mimeType"`
	Status          string     `json:"status"`
	UploadedAt      time.Time  `json:"uploadedAt"`
	ReviewedAt      *time.Time `json:"reviewedAt,omitempty"`
	ReviewedBy      string     `json:"reviewedBy,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
	Notes           string     `json:"notes,omitempty"`
	RejectionReason string     `json:"rejectionReason,omitempty"`
}

type DocumentStats struct {
	TotalDocuments     int                `json:"totalDocuments"`
	ByStatus           map[string]int     `json:"byStatus"`
	ByType             map[string]int     `json:"byType"`
	PendingCount       int                `json:"pendingCount"`
	AvgReviewTimeHours float64            `json:"avgReviewTimeHours"`
	ExpiringWithin30   int                `json:"expiringWithin30"`
	LastUpdated        time.Time          `json:"lastUpdated"`
}

type ReviewRequest struct {
	Status          string `json:"status"` // approved, rejected
	Notes           string `json:"notes"`
	RejectionReason string `json:"rejectionReason,omitempty"`
	ReviewedBy      string `json:"reviewedBy"`
}

type ResubmissionRequest struct {
	Reason string `json:"reason"`
	Notes  string `json:"notes"`
}

// ============================================
// Service
// ============================================

type ClientDocumentService struct {
	documents map[int64]*Document
	nextID    int64
	mu        sync.RWMutex
}

func NewClientDocumentService() *ClientDocumentService {
	s := &ClientDocumentService{
		documents: make(map[int64]*Document),
		nextID:    1,
	}
	s.generateMockData()
	return s
}

func (s *ClientDocumentService) generateMockData() {
	docTypes := []string{
		DocTypePassport, DocTypeNationalID, DocTypeDriversLicense,
		DocTypeUtilityBill, DocTypeBankStatement, DocTypeSelfie,
		DocTypeProofOfAddress, DocTypeTaxCertificate,
		DocTypeCompanyRegistration, DocTypePowerOfAttorney,
	}

	clientNames := []string{
		"John Smith", "Emma Johnson", "Michael Brown", "Sarah Davis",
		"James Wilson", "Emily Taylor", "David Anderson", "Jessica Martinez",
		"Robert Thomas", "Jennifer Jackson", "William White", "Mary Harris",
		"Christopher Martin", "Linda Thompson", "Daniel Garcia", "Patricia Robinson",
		"Matthew Clark", "Barbara Rodriguez", "Joseph Lewis", "Susan Lee",
		"Charles Walker", "Nancy Hall", "Thomas Allen", "Lisa Young",
		"Richard King", "Margaret Wright", "Mark Lopez", "Betty Hill",
		"Donald Scott", "Sandra Green", "Paul Adams", "Ashley Baker",
		"Steven Nelson", "Dorothy Carter", "Andrew Mitchell", "Kimberly Perez",
		"Joshua Roberts", "Elizabeth Turner", "Kenneth Phillips", "Donna Campbell",
		"Kevin Parker", "Carol Evans", "Brian Edwards", "Michelle Collins",
		"George Stewart", "Amanda Sanchez", "Edward Morris", "Melissa Rogers",
		"Ronald Reed", "Deborah Cook", "Timothy Morgan", "Stephanie Bell",
		"Jason Murphy", "Rebecca Bailey", "Jeffrey Rivera", "Laura Cooper",
		"Ryan Richardson", "Helen Cox", "Jacob Howard", "Sharon Ward",
		"Gary Torres", "Cynthia Peterson", "Nicholas Gray", "Kathleen Ramirez",
		"Eric James", "Angela Watson", "Jonathan Brooks", "Shirley Kelly",
		"Stephen Sanders", "Anna Price", "Larry Bennett", "Brenda Wood",
		"Justin Ross", "Pamela Henderson", "Scott Coleman", "Nicole Jenkins",
	}

	mimeTypes := map[string]string{
		"pdf":  "application/pdf",
		"jpg":  "image/jpeg",
		"png":  "image/png",
		"jpeg": "image/jpeg",
	}

	reviewers := []string{"admin_alice", "admin_bob", "admin_carol", "admin_dave", "admin_eve"}

	// Generate 300 documents across 80 clients
	now := time.Now()
	docCount := 0
	targetCounts := map[string]int{
		DocStatusApproved:             180, // 60%
		DocStatusPending:              60,  // 20%
		DocStatusRejected:             30,  // 10%
		DocStatusExpired:              15,  // 5%
		DocStatusResubmissionRequired: 15,  // 5%
	}

	statusList := []string{}
	for status, count := range targetCounts {
		for i := 0; i < count; i++ {
			statusList = append(statusList, status)
		}
	}
	rand.Shuffle(len(statusList), func(i, j int) { statusList[i], statusList[j] = statusList[j], statusList[i] })

	for clientID := int64(1); clientID <= int64(len(clientNames)); clientID++ {
		clientName := clientNames[clientID-1]
		numDocs := 2 + rand.Intn(6) // 2-7 documents per client

		for i := 0; i < numDocs && docCount < 300; i++ {
			docType := docTypes[rand.Intn(len(docTypes))]
			status := statusList[docCount]

			ext := []string{"pdf", "jpg", "png", "jpeg"}[rand.Intn(4)]
			fileName := fmt.Sprintf("%s_%s.%s", strings.ReplaceAll(strings.ToLower(clientName), " ", "_"), docType, ext)
			fileSize := int64(50000 + rand.Intn(5000000)) // 50KB - 5MB

			uploadedAt := now.Add(-time.Duration(rand.Intn(180)) * 24 * time.Hour) // Last 6 months

			doc := &Document{
				ID:         s.nextID,
				ClientID:   clientID,
				ClientName: clientName,
				Type:       docType,
				FileName:   fileName,
				FileSize:   fileSize,
				MimeType:   mimeTypes[ext],
				Status:     status,
				UploadedAt: uploadedAt,
			}

			// Set expiration for document types that expire
			if docType == DocTypePassport || docType == DocTypeNationalID || docType == DocTypeDriversLicense {
				expiresAt := uploadedAt.Add(time.Duration(365+rand.Intn(1825)) * 24 * time.Hour) // 1-5 years
				doc.ExpiresAt = &expiresAt

				// If status is expired, set expiration in the past
				if status == DocStatusExpired {
					pastExpiry := now.Add(-time.Duration(1+rand.Intn(180)) * 24 * time.Hour)
					doc.ExpiresAt = &pastExpiry
				}
			}

			// Set review details for non-pending documents
			if status != DocStatusPending {
				reviewedAt := uploadedAt.Add(time.Duration(1+rand.Intn(72)) * time.Hour) // 1-72 hours after upload
				doc.ReviewedAt = &reviewedAt
				doc.ReviewedBy = reviewers[rand.Intn(len(reviewers))]

				switch status {
				case DocStatusApproved:
					notes := []string{
						"Document verified successfully",
						"All details match client profile",
						"Clear and legible document",
						"Approved after verification",
					}
					doc.Notes = notes[rand.Intn(len(notes))]

				case DocStatusRejected:
					reasons := []string{
						"Document is blurry or unreadable",
						"Document has expired",
						"Name does not match profile",
						"Document appears to be altered",
						"Incomplete information",
						"Wrong document type submitted",
					}
					doc.RejectionReason = reasons[rand.Intn(len(reasons))]
					doc.Notes = "Please resubmit a valid document"

				case DocStatusResubmissionRequired:
					doc.RejectionReason = "Additional verification needed"
					doc.Notes = "Please submit a more recent copy of this document"
				}
			}

			s.documents[doc.ID] = doc
			s.nextID++
			docCount++
		}
	}

	log.Printf("[ClientDocuments] Generated %d documents across 80 clients (60%% approved, 20%% pending, 10%% rejected, 5%% expired, 5%% resubmission required)", docCount)
}

func (s *ClientDocumentService) ListDocuments(filters map[string]string, limit, offset int) ([]*Document, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*Document
	for _, doc := range s.documents {
		match := true

		if status, ok := filters["status"]; ok && status != "" && doc.Status != status {
			match = false
		}
		if docType, ok := filters["type"]; ok && docType != "" && doc.Type != docType {
			match = false
		}
		if clientID, ok := filters["clientId"]; ok && clientID != "" {
			id, _ := strconv.ParseInt(clientID, 10, 64)
			if doc.ClientID != id {
				match = false
			}
		}

		if match {
			filtered = append(filtered, doc)
		}
	}

	total := len(filtered)
	if offset >= total {
		return []*Document{}, total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return filtered[offset:end], total
}

func (s *ClientDocumentService) GetClientDocuments(clientID int64) []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var docs []*Document
	for _, doc := range s.documents {
		if doc.ClientID == clientID {
			docs = append(docs, doc)
		}
	}
	return docs
}

func (s *ClientDocumentService) GetDocument(id int64) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.documents[id]
}

func (s *ClientDocumentService) ReviewDocument(id int64, req ReviewRequest) (*Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, exists := s.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	if req.Status != DocStatusApproved && req.Status != DocStatusRejected {
		return nil, fmt.Errorf("invalid status: must be 'approved' or 'rejected'")
	}

	now := time.Now()
	doc.Status = req.Status
	doc.ReviewedAt = &now
	doc.ReviewedBy = req.ReviewedBy
	doc.Notes = req.Notes
	doc.RejectionReason = req.RejectionReason

	return doc, nil
}

func (s *ClientDocumentService) GetPendingDocuments() []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pending []*Document
	for _, doc := range s.documents {
		if doc.Status == DocStatusPending {
			pending = append(pending, doc)
		}
	}
	return pending
}

func (s *ClientDocumentService) GetExpiringDocuments(days int) []*Document {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	var expiring []*Document

	for _, doc := range s.documents {
		if doc.ExpiresAt != nil && doc.ExpiresAt.Before(cutoff) && doc.ExpiresAt.After(time.Now()) {
			expiring = append(expiring, doc)
		}
	}
	return expiring
}

func (s *ClientDocumentService) GetStats() DocumentStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := DocumentStats{
		TotalDocuments: len(s.documents),
		ByStatus:       make(map[string]int),
		ByType:         make(map[string]int),
		LastUpdated:    time.Now(),
	}

	var totalReviewTime time.Duration
	var reviewedCount int

	expiring30 := time.Now().Add(30 * 24 * time.Hour)

	for _, doc := range s.documents {
		stats.ByStatus[doc.Status]++
		stats.ByType[doc.Type]++

		if doc.Status == DocStatusPending {
			stats.PendingCount++
		}

		if doc.ReviewedAt != nil {
			reviewTime := doc.ReviewedAt.Sub(doc.UploadedAt)
			totalReviewTime += reviewTime
			reviewedCount++
		}

		if doc.ExpiresAt != nil && doc.ExpiresAt.Before(expiring30) && doc.ExpiresAt.After(time.Now()) {
			stats.ExpiringWithin30++
		}
	}

	if reviewedCount > 0 {
		stats.AvgReviewTimeHours = totalReviewTime.Hours() / float64(reviewedCount)
	}

	return stats
}

func (s *ClientDocumentService) RequestResubmission(id int64, req ResubmissionRequest) (*Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	doc, exists := s.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	doc.Status = DocStatusResubmissionRequired
	doc.RejectionReason = req.Reason
	doc.Notes = req.Notes

	return doc, nil
}

// ============================================
// HTTP Handlers
// ============================================

type ClientDocumentHandler struct {
	service     *ClientDocumentService
	authService *auth.Service
}

func NewClientDocumentHandler(service *ClientDocumentService, authService *auth.Service) *ClientDocumentHandler {
	return &ClientDocumentHandler{
		service:     service,
		authService: authService,
	}
}

// GET /admin/documents
func (h *ClientDocumentHandler) HandleListDocuments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(query.Get("offset"))

	filters := map[string]string{
		"status":   query.Get("status"),
		"type":     query.Get("type"),
		"clientId": query.Get("clientId"),
	}

	documents, total := h.service.ListDocuments(filters, limit, offset)

	response := map[string]interface{}{
		"documents": documents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	}

	json.NewEncoder(w).Encode(response)
}

// GET /admin/documents/client/:id
func (h *ClientDocumentHandler) HandleGetClientDocuments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/documents/client/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	documents := h.service.GetClientDocuments(clientID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clientId":  clientID,
		"documents": documents,
		"count":     len(documents),
	})
}

// GET /admin/documents/:id
func (h *ClientDocumentHandler) HandleGetDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/documents/"), "/")
	docID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	doc := h.service.GetDocument(docID)
	if doc == nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(doc)
}

// PUT /admin/documents/:id/review
func (h *ClientDocumentHandler) HandleReviewDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/documents/"), "/")
	docID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	doc, err := h.service.ReviewDocument(docID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(doc)
}

// GET /admin/documents/pending
func (h *ClientDocumentHandler) HandleGetPendingDocuments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	documents := h.service.GetPendingDocuments()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
	})
}

// GET /admin/documents/expiring
func (h *ClientDocumentHandler) HandleGetExpiringDocuments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	documents := h.service.GetExpiringDocuments(days)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
		"days":      days,
	})
}

// GET /admin/documents/stats
func (h *ClientDocumentHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}

// POST /admin/documents/:id/request-resubmission
func (h *ClientDocumentHandler) HandleRequestResubmission(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/documents/"), "/")
	docID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	var req ResubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	doc, err := h.service.RequestResubmission(docID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(doc)
}
