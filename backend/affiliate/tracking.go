package affiliate

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// TrackingManager handles click and conversion tracking
type TrackingManager struct {
	mu              sync.RWMutex
	links           map[int64]*AffiliateLink
	linksByCode     map[string]*AffiliateLink
	clicks          []*Click
	conversions     []*Conversion
	cookieStore     map[string]*ClickCookie // clickID -> cookie
	ipClickHistory  map[string][]time.Time // IP -> click timestamps (fraud detection)
	deviceFingerprints map[string][]time.Time // Device fingerprint -> timestamps
	fraudDetector   *FraudDetector
}

// ClickCookie represents a tracking cookie
type ClickCookie struct {
	ClickID     string
	AffiliateID int64
	ClickTime   time.Time
	ExpiresAt   time.Time
	IPAddress   string
	UserAgent   string
}

// NewTrackingManager creates a new tracking manager
func NewTrackingManager() *TrackingManager {
	tm := &TrackingManager{
		links:              make(map[int64]*AffiliateLink),
		linksByCode:        make(map[string]*AffiliateLink),
		clicks:             make([]*Click, 0),
		conversions:        make([]*Conversion, 0),
		cookieStore:        make(map[string]*ClickCookie),
		ipClickHistory:     make(map[string][]time.Time),
		deviceFingerprints: make(map[string][]time.Time),
		fraudDetector:      NewFraudDetector(),
	}

	// Start cleanup goroutine
	go tm.cleanupExpiredCookies()

	return tm
}

// CreateLink creates a new affiliate tracking link
func (tm *TrackingManager) CreateLink(link *AffiliateLink) (*AffiliateLink, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Generate link code if not provided
	if link.LinkCode == "" {
		link.LinkCode = tm.generateLinkCode()
	}

	// Check if code already exists
	if _, exists := tm.linksByCode[link.LinkCode]; exists {
		return nil, errors.New("link code already exists")
	}

	// Generate ID
	link.ID = int64(len(tm.links) + 1)
	link.CreatedAt = time.Now()
	link.IsActive = true

	// Build full URL
	baseURL := "https://yourbroker.com/signup"
	if link.LandingPage != "" {
		baseURL = link.LandingPage
	}
	link.FullURL = tm.buildTrackingURL(baseURL, link.LinkCode, link.Campaign, link.Source, link.Medium, link.Content)

	tm.links[link.ID] = link
	tm.linksByCode[link.LinkCode] = link

	log.Printf("[Tracking] Created link: %s (Affiliate=%d, Campaign=%s)", link.LinkCode, link.AffiliateID, link.Campaign)
	return link, nil
}

// TrackClick records a click on an affiliate link
func (tm *TrackingManager) TrackClick(linkCode, ipAddress, userAgent, referrer, landingPage string) (*Click, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Find link
	link, ok := tm.linksByCode[linkCode]
	if !ok {
		return nil, errors.New("invalid tracking link")
	}

	if !link.IsActive {
		return nil, errors.New("link is inactive")
	}

	// Generate click ID
	clickID := tm.generateClickID()

	// Parse device info
	device, browser, os := tm.parseUserAgent(userAgent)

	// Detect fraud
	isFraudulent, fraudReason := tm.fraudDetector.DetectFraud(ipAddress, userAgent, tm.ipClickHistory, tm.deviceFingerprints)

	// Determine if unique click
	isUnique := tm.isUniqueClick(link.AffiliateID, ipAddress, userAgent)

	click := &Click{
		ID:           int64(len(tm.clicks) + 1),
		AffiliateID:  link.AffiliateID,
		LinkID:       link.ID,
		ClickID:      clickID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Device:       device,
		Browser:      browser,
		OS:           os,
		Referrer:     referrer,
		LandingPage:  landingPage,
		IsUnique:     isUnique,
		IsFraudulent: isFraudulent,
		FraudReason:  fraudReason,
		CreatedAt:    time.Now(),
	}

	// Geo-location (simplified - in production use MaxMind GeoIP2)
	click.Country, click.City = tm.getGeoLocation(ipAddress)

	tm.clicks = append(tm.clicks, click)

	// Update link stats
	link.TotalClicks++
	if isUnique {
		link.UniqueClicks++
	}

	// Create tracking cookie (30 days default)
	if !isFraudulent {
		cookie := &ClickCookie{
			ClickID:     clickID,
			AffiliateID: link.AffiliateID,
			ClickTime:   time.Now(),
			ExpiresAt:   time.Now().Add(30 * 24 * time.Hour), // 30 days
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
		}
		tm.cookieStore[clickID] = cookie

		// Track IP/device history for fraud detection
		tm.ipClickHistory[ipAddress] = append(tm.ipClickHistory[ipAddress], time.Now())
		fingerprint := tm.generateDeviceFingerprint(ipAddress, userAgent)
		tm.deviceFingerprints[fingerprint] = append(tm.deviceFingerprints[fingerprint], time.Now())
	}

	log.Printf("[Tracking] Click recorded: LinkCode=%s, ClickID=%s, IP=%s, Unique=%v, Fraud=%v",
		linkCode, clickID, ipAddress, isUnique, isFraudulent)

	return click, nil
}

// TrackConversion records a conversion (signup/deposit/trade)
func (tm *TrackingManager) TrackConversion(clickID, userID string, accountID int64, conversionType string, value float64) (*Conversion, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Find cookie
	cookie, ok := tm.cookieStore[clickID]
	if !ok {
		return nil, errors.New("invalid or expired click ID")
	}

	// Check if cookie expired
	if time.Now().After(cookie.ExpiresAt) {
		return nil, errors.New("tracking cookie has expired")
	}

	// Create conversion
	conversion := &Conversion{
		ID:               int64(len(tm.conversions) + 1),
		AffiliateID:      cookie.AffiliateID,
		ClickID:          clickID,
		UserID:           userID,
		AccountID:        accountID,
		ConversionType:   conversionType,
		AttributionModel: "LAST_CLICK", // Default attribution
		Value:            value,
		Status:           "PENDING",
		CreatedAt:        time.Now(),
	}

	tm.conversions = append(tm.conversions, conversion)

	// Update click conversion time
	for _, click := range tm.clicks {
		if click.ClickID == clickID {
			now := time.Now()
			click.ConvertedAt = &now
			break
		}
	}

	// Update link conversion count
	for _, link := range tm.links {
		if link.AffiliateID == cookie.AffiliateID {
			link.Conversions++
			break
		}
	}

	log.Printf("[Tracking] Conversion recorded: Type=%s, Affiliate=%d, User=%s, Value=%.2f",
		conversionType, cookie.AffiliateID, userID, value)

	return conversion, nil
}

// ApproveConversion approves a pending conversion
func (tm *TrackingManager) ApproveConversion(conversionID int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, conv := range tm.conversions {
		if conv.ID == conversionID {
			conv.Status = "APPROVED"
			now := time.Now()
			conv.ApprovedAt = &now
			log.Printf("[Tracking] Approved conversion: ID=%d, Affiliate=%d", conversionID, conv.AffiliateID)
			return nil
		}
	}
	return errors.New("conversion not found")
}

// RejectConversion rejects a conversion
func (tm *TrackingManager) RejectConversion(conversionID int64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, conv := range tm.conversions {
		if conv.ID == conversionID {
			conv.Status = "REJECTED"
			log.Printf("[Tracking] Rejected conversion: ID=%d, Affiliate=%d", conversionID, conv.AffiliateID)
			return nil
		}
	}
	return errors.New("conversion not found")
}

// GetClickByCookie retrieves click information from cookie
func (tm *TrackingManager) GetClickByCookie(clickID string) (*ClickCookie, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	cookie, ok := tm.cookieStore[clickID]
	return cookie, ok
}

// GetAffiliateClicks returns all clicks for an affiliate
func (tm *TrackingManager) GetAffiliateClicks(affiliateID int64, startDate, endDate time.Time) []*Click {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var result []*Click
	for _, click := range tm.clicks {
		if click.AffiliateID == affiliateID &&
			click.CreatedAt.After(startDate) &&
			click.CreatedAt.Before(endDate) {
			result = append(result, click)
		}
	}
	return result
}

// GetAffiliateConversions returns all conversions for an affiliate
func (tm *TrackingManager) GetAffiliateConversions(affiliateID int64, startDate, endDate time.Time) []*Conversion {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var result []*Conversion
	for _, conv := range tm.conversions {
		if conv.AffiliateID == affiliateID &&
			conv.CreatedAt.After(startDate) &&
			conv.CreatedAt.Before(endDate) {
			result = append(result, conv)
		}
	}
	return result
}

// GetLinkStats returns statistics for a specific link
func (tm *TrackingManager) GetLinkStats(linkID int64) (*AffiliateLink, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	link, ok := tm.links[linkID]
	if !ok {
		return nil, errors.New("link not found")
	}
	return link, nil
}

// Helper functions

func (tm *TrackingManager) generateLinkCode() string {
	for {
		b := make([]byte, 6)
		rand.Read(b)
		code := base64.URLEncoding.EncodeToString(b)[:8]
		code = strings.ToLower(strings.ReplaceAll(code, "-", ""))
		code = strings.ReplaceAll(code, "_", "")
		if _, exists := tm.linksByCode[code]; !exists {
			return code
		}
	}
}

func (tm *TrackingManager) generateClickID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (tm *TrackingManager) buildTrackingURL(baseURL, linkCode, campaign, source, medium, content string) string {
	url := baseURL + "?ref=" + linkCode
	if campaign != "" {
		url += "&utm_campaign=" + campaign
	}
	if source != "" {
		url += "&utm_source=" + source
	}
	if medium != "" {
		url += "&utm_medium=" + medium
	}
	if content != "" {
		url += "&utm_content=" + content
	}
	return url
}

func (tm *TrackingManager) parseUserAgent(userAgent string) (device, browser, os string) {
	ua := strings.ToLower(userAgent)

	// Device detection
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		device = "MOBILE"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		device = "TABLET"
	} else {
		device = "DESKTOP"
	}

	// Browser detection
	if strings.Contains(ua, "chrome") {
		browser = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "Firefox"
	} else if strings.Contains(ua, "safari") {
		browser = "Safari"
	} else if strings.Contains(ua, "edge") {
		browser = "Edge"
	} else {
		browser = "Other"
	}

	// OS detection
	if strings.Contains(ua, "windows") {
		os = "Windows"
	} else if strings.Contains(ua, "mac") {
		os = "macOS"
	} else if strings.Contains(ua, "linux") {
		os = "Linux"
	} else if strings.Contains(ua, "android") {
		os = "Android"
	} else if strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") {
		os = "iOS"
	} else {
		os = "Other"
	}

	return device, browser, os
}

func (tm *TrackingManager) getGeoLocation(ipAddress string) (country, city string) {
	// Simplified geo-location
	// In production, use MaxMind GeoIP2 or similar service
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return "Unknown", "Unknown"
	}

	// Check if private IP
	if ip.IsPrivate() || ip.IsLoopback() {
		return "Local", "Local"
	}

	// Default for now
	return "Unknown", "Unknown"
}

func (tm *TrackingManager) isUniqueClick(affiliateID int64, ipAddress, userAgent string) bool {
	// Check last 24 hours for same IP + UserAgent + Affiliate
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, click := range tm.clicks {
		if click.AffiliateID == affiliateID &&
			click.IPAddress == ipAddress &&
			click.UserAgent == userAgent &&
			click.CreatedAt.After(cutoff) {
			return false
		}
	}
	return true
}

func (tm *TrackingManager) generateDeviceFingerprint(ipAddress, userAgent string) string {
	// Simple fingerprint - in production, use more sophisticated methods
	return ipAddress + "|" + userAgent
}

func (tm *TrackingManager) cleanupExpiredCookies() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		tm.mu.Lock()
		now := time.Now()
		for clickID, cookie := range tm.cookieStore {
			if now.After(cookie.ExpiresAt) {
				delete(tm.cookieStore, clickID)
			}
		}
		tm.mu.Unlock()
	}
}

// FraudDetector handles fraud detection
type FraudDetector struct {
	maxClicksPerIP       int
	maxClicksPerDevice   int
	timeWindow           time.Duration
	suspiciousIPPrefixes []string
}

func NewFraudDetector() *FraudDetector {
	return &FraudDetector{
		maxClicksPerIP:     100, // Max clicks per IP per hour
		maxClicksPerDevice: 50,  // Max clicks per device per hour
		timeWindow:         1 * time.Hour,
		suspiciousIPPrefixes: []string{
			"10.", "192.168.", "172.16.", // Private IPs
		},
	}
}

func (fd *FraudDetector) DetectFraud(ipAddress, userAgent string, ipHistory, deviceHistory map[string][]time.Time) (bool, string) {
	now := time.Now()
	cutoff := now.Add(-fd.timeWindow)

	// Check IP click rate
	if clicks, ok := ipHistory[ipAddress]; ok {
		recentClicks := 0
		for _, clickTime := range clicks {
			if clickTime.After(cutoff) {
				recentClicks++
			}
		}
		if recentClicks > fd.maxClicksPerIP {
			return true, "IP_VELOCITY_EXCEEDED"
		}
	}

	// Check for suspicious IPs
	for _, prefix := range fd.suspiciousIPPrefixes {
		if strings.HasPrefix(ipAddress, prefix) {
			return true, "SUSPICIOUS_IP"
		}
	}

	// Check for bot patterns in user agent
	ua := strings.ToLower(userAgent)
	botKeywords := []string{"bot", "crawler", "spider", "scraper"}
	for _, keyword := range botKeywords {
		if strings.Contains(ua, keyword) {
			return true, "BOT_DETECTED"
		}
	}

	// Check for missing user agent
	if userAgent == "" {
		return true, "MISSING_USER_AGENT"
	}

	return false, ""
}
