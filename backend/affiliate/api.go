package affiliate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// AffiliateAPI handles HTTP endpoints for affiliate system
type AffiliateAPI struct {
	programManager  *ProgramManager
	trackingManager *TrackingManager
	commissionMgr   *CommissionManager
	referralManager *ReferralManager
	analytics       *AnalyticsEngine
}

// NewAffiliateAPI creates a new affiliate API handler
func NewAffiliateAPI() *AffiliateAPI {
	pm := NewProgramManager()
	tm := NewTrackingManager()
	cm := NewCommissionManager(pm, tm)
	rm := NewReferralManager()
	ae := NewAnalyticsEngine(pm, tm, cm)

	// Initialize default program
	pm.CreateProgram(&AffiliateProgram{
		Name:                "Standard Affiliate Program",
		CommissionModel:     "HYBRID",
		CPAAmount:           100.00,
		RevSharePercent:     25.00,
		MinPayout:           100.00,
		PayoutSchedule:      "MONTHLY",
		CookieDuration:      30,
		SubAffiliateEnabled: true,
		SubAffiliatePercent: 10.00,
	})

	return &AffiliateAPI{
		programManager:  pm,
		trackingManager: tm,
		commissionMgr:   cm,
		referralManager: rm,
		analytics:       ae,
	}
}

// RegisterRoutes registers all affiliate API routes
func (api *AffiliateAPI) RegisterRoutes() {
	// Affiliate Management
	http.HandleFunc("/api/affiliate/register", api.handleCORS(api.HandleRegisterAffiliate))
	http.HandleFunc("/api/affiliate/login", api.handleCORS(api.HandleAffiliateLogin))
	http.HandleFunc("/api/affiliate/profile", api.handleCORS(api.HandleGetProfile))
	http.HandleFunc("/api/affiliate/update", api.handleCORS(api.HandleUpdateProfile))

	// Links
	http.HandleFunc("/api/affiliate/links", api.handleCORS(api.HandleGetLinks))
	http.HandleFunc("/api/affiliate/links/create", api.handleCORS(api.HandleCreateLink))

	// Dashboard & Analytics
	http.HandleFunc("/api/affiliate/dashboard", api.handleCORS(api.HandleGetDashboard))
	http.HandleFunc("/api/affiliate/stats", api.handleCORS(api.HandleGetStats))
	http.HandleFunc("/api/affiliate/funnel", api.handleCORS(api.HandleGetFunnel))
	http.HandleFunc("/api/affiliate/traffic", api.handleCORS(api.HandleGetTraffic))
	http.HandleFunc("/api/affiliate/performance", api.handleCORS(api.HandleGetPerformance))

	// Commissions & Payouts
	http.HandleFunc("/api/affiliate/commissions", api.handleCORS(api.HandleGetCommissions))
	http.HandleFunc("/api/affiliate/payouts", api.handleCORS(api.HandleGetPayouts))
	http.HandleFunc("/api/affiliate/payout/request", api.handleCORS(api.HandleRequestPayout))

	// Marketing Materials
	http.HandleFunc("/api/affiliate/materials", api.handleCORS(api.HandleGetMaterials))

	// Referral Program (User-to-User)
	http.HandleFunc("/api/referral/code", api.handleCORS(api.HandleGetReferralCode))
	http.HandleFunc("/api/referral/apply", api.handleCORS(api.HandleApplyReferralCode))
	http.HandleFunc("/api/referral/stats", api.handleCORS(api.HandleGetReferralStats))
	http.HandleFunc("/api/referral/leaderboard", api.handleCORS(api.HandleGetLeaderboard))

	// Public Tracking (No auth)
	http.HandleFunc("/track/click", api.handleCORS(api.HandleTrackClick))
	http.HandleFunc("/track/pixel.gif", api.handleCORS(api.HandleTrackingPixel))

	// Admin Endpoints
	http.HandleFunc("/admin/affiliate/list", api.handleCORS(api.HandleAdminListAffiliates))
	http.HandleFunc("/admin/affiliate/approve", api.handleCORS(api.HandleAdminApproveAffiliate))
	http.HandleFunc("/admin/affiliate/suspend", api.handleCORS(api.HandleAdminSuspendAffiliate))
	http.HandleFunc("/admin/affiliate/commissions/approve", api.handleCORS(api.HandleAdminApproveCommission))
	http.HandleFunc("/admin/affiliate/payouts/process", api.handleCORS(api.HandleAdminProcessPayout))
	http.HandleFunc("/admin/affiliate/fraud", api.handleCORS(api.HandleAdminGetFraudIncidents))
}

// handleCORS wraps handlers with CORS headers
func (api *AffiliateAPI) handleCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

// HandleRegisterAffiliate handles affiliate registration
func (api *AffiliateAPI) HandleRegisterAffiliate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CompanyName string `json:"companyName"`
		ContactName string `json:"contactName"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Country     string `json:"country"`
		Website     string `json:"website"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	affiliate := &Affiliate{
		CompanyName: req.CompanyName,
		ContactName: req.ContactName,
		Email:       req.Email,
		Phone:       req.Phone,
		Country:     req.Country,
		Website:     req.Website,
	}

	result, err := api.programManager.RegisterAffiliate(affiliate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"affiliate": result,
		"message":   "Application submitted. You will be notified once approved.",
	})
}

// HandleGetDashboard returns affiliate dashboard statistics
func (api *AffiliateAPI) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get affiliate ID from auth token (simplified - use JWT in production)
	affiliateID := api.getAffiliateIDFromRequest(r)
	if affiliateID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := api.analytics.GetDashboardStats(affiliateID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// HandleCreateLink creates a new tracking link
func (api *AffiliateAPI) HandleCreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	affiliateID := api.getAffiliateIDFromRequest(r)
	if affiliateID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		LandingPage string `json:"landingPage"`
		Campaign    string `json:"campaign"`
		Source      string `json:"source"`
		Medium      string `json:"medium"`
		Content     string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	link := &AffiliateLink{
		AffiliateID: affiliateID,
		LandingPage: req.LandingPage,
		Campaign:    req.Campaign,
		Source:      req.Source,
		Medium:      req.Medium,
		Content:     req.Content,
	}

	result, err := api.trackingManager.CreateLink(link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleTrackClick handles public click tracking
func (api *AffiliateAPI) HandleTrackClick(w http.ResponseWriter, r *http.Request) {
	linkCode := r.URL.Query().Get("ref")
	if linkCode == "" {
		http.Redirect(w, r, "https://yourbroker.com/signup", http.StatusTemporaryRedirect)
		return
	}

	// Extract tracking data
	ipAddress := api.getClientIP(r)
	userAgent := r.UserAgent()
	referrer := r.Referer()
	landingPage := r.URL.Query().Get("landing")

	// Track click
	click, err := api.trackingManager.TrackClick(linkCode, ipAddress, userAgent, referrer, landingPage)
	if err != nil {
		log.Printf("[Tracking] Error tracking click: %v", err)
	}

	// Set tracking cookie
	if click != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "aff_click",
			Value:    click.ClickID,
			Path:     "/",
			MaxAge:   30 * 24 * 60 * 60, // 30 days
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	// Redirect to landing page or default
	redirectURL := "https://yourbroker.com/signup?ref=" + linkCode
	if landingPage != "" {
		redirectURL = landingPage + "?ref=" + linkCode
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// HandleTrackingPixel serves a 1x1 tracking pixel
func (api *AffiliateAPI) HandleTrackingPixel(w http.ResponseWriter, r *http.Request) {
	// Track pixel view
	linkCode := r.URL.Query().Get("ref")
	if linkCode != "" {
		ipAddress := api.getClientIP(r)
		userAgent := r.UserAgent()
		referrer := r.Referer()
		api.trackingManager.TrackClick(linkCode, ipAddress, userAgent, referrer, "")
	}

	// Serve transparent 1x1 GIF
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	// 1x1 transparent GIF
	gif := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b}
	w.Write(gif)
}

// HandleGetReferralCode gets or creates a referral code for a user
func (api *AffiliateAPI) HandleGetReferralCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := api.getUserIDFromRequest(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has existing codes
	codes := api.referralManager.GetUserReferralCodes(userID)

	if len(codes) == 0 && r.Method == "POST" {
		// Create new code
		code, err := api.referralManager.CreateReferralCode(userID, 25.00, 25.00, 0, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		codes = []*ReferralCode{code}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"codes": codes,
	})
}

// HandleApplyReferralCode applies a referral code during signup
func (api *AffiliateAPI) HandleApplyReferralCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Code   string `json:"code"`
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reward, err := api.referralManager.ApplyReferralCode(req.Code, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Auto-credit rewards
	api.referralManager.CreditReferralReward(reward.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"reward":  reward,
		"message": fmt.Sprintf("Congratulations! You've received $%.2f bonus!", reward.RefereeReward),
	})
}

// HandleGetLeaderboard returns top referrers
func (api *AffiliateAPI) HandleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	leaderboard := api.referralManager.GetTopReferrers(limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard": leaderboard,
	})
}

// Helper functions

func (api *AffiliateAPI) getAffiliateIDFromRequest(r *http.Request) int64 {
	// In production, extract from JWT token
	// Simplified for demo
	idStr := r.Header.Get("X-Affiliate-ID")
	if idStr == "" {
		idStr = r.URL.Query().Get("affiliate_id")
	}
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return id
	}
	return 0
}

func (api *AffiliateAPI) getUserIDFromRequest(r *http.Request) string {
	// In production, extract from JWT token
	return r.Header.Get("X-User-ID")
}

func (api *AffiliateAPI) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Admin handlers

func (api *AffiliateAPI) HandleAdminListAffiliates(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	affiliates := api.programManager.ListAffiliates(status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(affiliates)
}

func (api *AffiliateAPI) HandleAdminApproveAffiliate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AffiliateID int64 `json:"affiliateId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := api.programManager.ApproveAffiliate(req.AffiliateID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Affiliate approved successfully",
	})
}

func (api *AffiliateAPI) HandleAdminSuspendAffiliate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AffiliateID int64  `json:"affiliateId"`
		Reason      string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := api.programManager.SuspendAffiliate(req.AffiliateID, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Affiliate suspended",
	})
}

// Stub handlers for remaining endpoints
func (api *AffiliateAPI) HandleAffiliateLogin(w http.ResponseWriter, r *http.Request)       {}
func (api *AffiliateAPI) HandleGetProfile(w http.ResponseWriter, r *http.Request)           {}
func (api *AffiliateAPI) HandleUpdateProfile(w http.ResponseWriter, r *http.Request)        {}
func (api *AffiliateAPI) HandleGetLinks(w http.ResponseWriter, r *http.Request)             {}
func (api *AffiliateAPI) HandleGetStats(w http.ResponseWriter, r *http.Request)             {}
func (api *AffiliateAPI) HandleGetFunnel(w http.ResponseWriter, r *http.Request)            {}
func (api *AffiliateAPI) HandleGetTraffic(w http.ResponseWriter, r *http.Request)           {}
func (api *AffiliateAPI) HandleGetPerformance(w http.ResponseWriter, r *http.Request)       {}
func (api *AffiliateAPI) HandleGetCommissions(w http.ResponseWriter, r *http.Request)       {}
func (api *AffiliateAPI) HandleGetPayouts(w http.ResponseWriter, r *http.Request)           {}
func (api *AffiliateAPI) HandleRequestPayout(w http.ResponseWriter, r *http.Request)        {}
func (api *AffiliateAPI) HandleGetMaterials(w http.ResponseWriter, r *http.Request)         {}
func (api *AffiliateAPI) HandleGetReferralStats(w http.ResponseWriter, r *http.Request)     {}
func (api *AffiliateAPI) HandleAdminApproveCommission(w http.ResponseWriter, r *http.Request) {}
func (api *AffiliateAPI) HandleAdminProcessPayout(w http.ResponseWriter, r *http.Request)   {}
func (api *AffiliateAPI) HandleAdminGetFraudIncidents(w http.ResponseWriter, r *http.Request) {}
