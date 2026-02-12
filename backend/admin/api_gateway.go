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

// APIClient represents an API client/key with usage quotas
type APIClient struct {
	ID              int64     `json:"id"`
	ClientName      string    `json:"client_name"`
	APIKey          string    `json:"api_key"`
	Tier            string    `json:"tier"`
	Status          string    `json:"status"`
	RequestsToday   int       `json:"requests_today"`
	RequestsMonth   int       `json:"requests_month"`
	DailyQuota      int       `json:"daily_quota"`
	MonthlyQuota    int       `json:"monthly_quota"`
	RateLimitPerMin int       `json:"rate_limit_per_min"`
	ErrorRate       float64   `json:"error_rate"`
	AvgResponseTime float64   `json:"avg_response_time_ms"`
	LastUsed        time.Time `json:"last_used"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	IPWhitelist     []string  `json:"ip_whitelist,omitempty"`
	AllowedEndpoints []string `json:"allowed_endpoints,omitempty"`
}

// RateLimitTier represents a rate limit tier configuration
type RateLimitTier struct {
	TierName        string `json:"tier_name"`
	RequestsPerMin  int    `json:"requests_per_min"`
	DailyQuota      int    `json:"daily_quota"`
	MonthlyQuota    int    `json:"monthly_quota"`
	BurstLimit      int    `json:"burst_limit"`
	ConcurrentConns int    `json:"concurrent_connections"`
	Price           float64 `json:"price_per_month"`
}

// APIUsageRecord represents a single API call record
type APIUsageRecord struct {
	ID             int64     `json:"id"`
	ClientID       int64     `json:"client_id"`
	ClientName     string    `json:"client_name"`
	Endpoint       string    `json:"endpoint"`
	Method         string    `json:"method"`
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs float64   `json:"response_time_ms"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	Timestamp      time.Time `json:"timestamp"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

// IPThrottleRule represents an IP-based throttling rule
type IPThrottleRule struct {
	IPAddress   string    `json:"ip_address"`
	Action      string    `json:"action"`
	RateLimit   int       `json:"rate_limit"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	HitCount    int       `json:"hit_count"`
	LastHit     time.Time `json:"last_hit"`
}

// GeographicRule represents a region-based access rule
type GeographicRule struct {
	ID          int64     `json:"id"`
	Region      string    `json:"region"`
	CountryCode string    `json:"country_code"`
	Action      string    `json:"action"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// APIGatewayService manages API gateway, rate limits, and usage analytics
type APIGatewayService struct {
	mu              sync.RWMutex
	clients         map[int64]*APIClient
	rateLimitTiers  map[string]*RateLimitTier
	usageRecords    []*APIUsageRecord
	ipRules         map[string]*IPThrottleRule
	geoRules        []*GeographicRule
	nextClientID    int64
	nextRecordID    int64
	nextGeoRuleID   int64
}

// NewAPIGatewayService creates a new API gateway service
func NewAPIGatewayService() *APIGatewayService {
	s := &APIGatewayService{
		clients:        make(map[int64]*APIClient),
		rateLimitTiers: make(map[string]*RateLimitTier),
		usageRecords:   make([]*APIUsageRecord, 0),
		ipRules:        make(map[string]*IPThrottleRule),
		geoRules:       make([]*GeographicRule, 0),
		nextClientID:   1,
		nextRecordID:   1,
		nextGeoRuleID:  1,
	}
	s.initMockData()
	return s
}

func (s *APIGatewayService) initMockData() {
	// Initialize 8 rate limit tiers
	s.rateLimitTiers["free"] = &RateLimitTier{
		TierName:        "free",
		RequestsPerMin:  100,
		DailyQuota:      10000,
		MonthlyQuota:    250000,
		BurstLimit:      150,
		ConcurrentConns: 5,
		Price:           0,
	}
	s.rateLimitTiers["basic"] = &RateLimitTier{
		TierName:        "basic",
		RequestsPerMin:  500,
		DailyQuota:      50000,
		MonthlyQuota:    1500000,
		BurstLimit:      750,
		ConcurrentConns: 20,
		Price:           49.99,
	}
	s.rateLimitTiers["pro"] = &RateLimitTier{
		TierName:        "pro",
		RequestsPerMin:  2000,
		DailyQuota:      250000,
		MonthlyQuota:    7500000,
		BurstLimit:      3000,
		ConcurrentConns: 100,
		Price:           199.99,
	}
	s.rateLimitTiers["enterprise"] = &RateLimitTier{
		TierName:        "enterprise",
		RequestsPerMin:  10000,
		DailyQuota:      1000000,
		MonthlyQuota:    30000000,
		BurstLimit:      15000,
		ConcurrentConns: 500,
		Price:           999.99,
	}
	s.rateLimitTiers["internal"] = &RateLimitTier{
		TierName:        "internal",
		RequestsPerMin:  0, // unlimited
		DailyQuota:      0, // unlimited
		MonthlyQuota:    0, // unlimited
		BurstLimit:      0, // unlimited
		ConcurrentConns: 0, // unlimited
		Price:           0,
	}
	s.rateLimitTiers["partner"] = &RateLimitTier{
		TierName:        "partner",
		RequestsPerMin:  5000,
		DailyQuota:      500000,
		MonthlyQuota:    15000000,
		BurstLimit:      7500,
		ConcurrentConns: 250,
		Price:           499.99,
	}
	s.rateLimitTiers["trial"] = &RateLimitTier{
		TierName:        "trial",
		RequestsPerMin:  200,
		DailyQuota:      20000,
		MonthlyQuota:    100000,
		BurstLimit:      300,
		ConcurrentConns: 10,
		Price:           0,
	}
	s.rateLimitTiers["premium"] = &RateLimitTier{
		TierName:        "premium",
		RequestsPerMin:  3000,
		DailyQuota:      400000,
		MonthlyQuota:    12000000,
		BurstLimit:      4500,
		ConcurrentConns: 150,
		Price:           299.99,
	}

	// Generate 20 API clients
	clientNames := []string{
		"Mobile App iOS", "Mobile App Android", "Web Trading Platform",
		"Third-Party Analytics", "Partner Integration A", "Partner Integration B",
		"Internal Admin Tool", "Desktop Client Windows", "Desktop Client Mac",
		"Algorithmic Trading Bot", "Market Data Provider", "Signal Provider Platform",
		"Copy Trading Service", "Risk Management Tool", "Reporting Dashboard",
		"External CRM Integration", "Payment Gateway", "Compliance Tool",
		"Backup Service", "Testing Environment",
	}
	tiers := []string{"free", "basic", "pro", "enterprise", "internal", "partner", "trial", "premium"}
	statuses := []string{"active", "active", "active", "active", "suspended", "expired"}

	for i := 0; i < 20; i++ {
		tier := tiers[i%len(tiers)]
		tierConfig := s.rateLimitTiers[tier]
		status := statuses[i%len(statuses)]

		// Calculate usage based on tier and status
		requestsToday := 0
		requestsMonth := 0
		if status == "active" {
			if tierConfig.DailyQuota > 0 {
				requestsToday = rand.Intn(tierConfig.DailyQuota / 2)
				requestsMonth = requestsToday * (rand.Intn(15) + 15) // 15-30 days worth
			} else {
				// unlimited tier
				requestsToday = rand.Intn(100000) + 50000
				requestsMonth = requestsToday * 25
			}
		}

		errorRate := 0.5 + rand.Float64()*4.5 // 0.5% to 5%
		avgResponseTime := 50.0 + rand.Float64()*200.0 // 50-250ms

		var expiresAt *time.Time
		if tier == "trial" {
			expiry := time.Now().AddDate(0, 0, rand.Intn(30)+1)
			expiresAt = &expiry
		}

		client := &APIClient{
			ID:              s.nextClientID,
			ClientName:      clientNames[i],
			APIKey:          fmt.Sprintf("rtx5_api_%d_%s", s.nextClientID, generateRandomKey(24)),
			Tier:            tier,
			Status:          status,
			RequestsToday:   requestsToday,
			RequestsMonth:   requestsMonth,
			DailyQuota:      tierConfig.DailyQuota,
			MonthlyQuota:    tierConfig.MonthlyQuota,
			RateLimitPerMin: tierConfig.RequestsPerMin,
			ErrorRate:       errorRate,
			AvgResponseTime: avgResponseTime,
			LastUsed:        time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second),
			CreatedAt:       time.Now().AddDate(0, 0, -rand.Intn(365)),
			ExpiresAt:       expiresAt,
		}

		// Add IP whitelist for some clients
		if i%3 == 0 {
			client.IPWhitelist = []string{
				fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256)),
				fmt.Sprintf("10.0.%d.%d", rand.Intn(256), rand.Intn(256)),
			}
		}

		// Add allowed endpoints for restricted clients
		if i%4 == 0 {
			client.AllowedEndpoints = []string{
				"/api/market-data/*",
				"/api/account/balance",
				"/api/orders/history",
			}
		}

		s.clients[s.nextClientID] = client
		s.nextClientID++
	}

	// Generate 50 usage records per client (1000 total)
	endpoints := []string{
		"/api/market-data/quotes", "/api/market-data/history", "/api/orders/create",
		"/api/orders/cancel", "/api/account/balance", "/api/account/positions",
		"/api/trades/history", "/api/symbols/list", "/api/user/profile",
		"/api/risk/check", "/api/analytics/performance", "/api/reports/generate",
	}
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	statusCodes := []int{200, 200, 200, 200, 200, 201, 400, 401, 403, 404, 429, 500, 503}

	for clientID := int64(1); clientID <= 20; clientID++ {
		client := s.clients[clientID]
		if client == nil {
			continue
		}

		for j := 0; j < 50; j++ {
			endpoint := endpoints[rand.Intn(len(endpoints))]
			method := methods[rand.Intn(len(methods))]
			statusCode := statusCodes[rand.Intn(len(statusCodes))]
			responseTime := 10.0 + rand.Float64()*490.0 // 10-500ms

			var errorMsg string
			if statusCode >= 400 {
				switch statusCode {
				case 400:
					errorMsg = "Bad Request: Invalid parameters"
				case 401:
					errorMsg = "Unauthorized: Invalid API key"
				case 403:
					errorMsg = "Forbidden: Insufficient permissions"
				case 404:
					errorMsg = "Not Found: Resource does not exist"
				case 429:
					errorMsg = "Too Many Requests: Rate limit exceeded"
				case 500:
					errorMsg = "Internal Server Error"
				case 503:
					errorMsg = "Service Unavailable: Maintenance mode"
				}
			}

			record := &APIUsageRecord{
				ID:             s.nextRecordID,
				ClientID:       clientID,
				ClientName:     client.ClientName,
				Endpoint:       endpoint,
				Method:         method,
				StatusCode:     statusCode,
				ResponseTimeMs: responseTime,
				IPAddress:      fmt.Sprintf("203.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256)),
				UserAgent:      "RTX5-Client/1.0",
				Timestamp:      time.Now().Add(-time.Duration(rand.Intn(86400*7)) * time.Second),
				ErrorMessage:   errorMsg,
			}

			s.usageRecords = append(s.usageRecords, record)
			s.nextRecordID++
		}
	}

	// Generate 10 IP throttle rules
	ipActions := []string{"allow", "allow", "allow", "block", "throttle", "throttle"}
	ipReasons := []string{
		"Whitelisted corporate IP", "Known partner", "Internal network",
		"Suspicious activity detected", "Rate limit violations", "Repeated authentication failures",
		"Spam patterns detected", "Geographic restriction", "Manual block",
	}

	for i := 0; i < 10; i++ {
		action := ipActions[i%len(ipActions)]
		rateLimit := 0
		if action == "throttle" {
			rateLimit = []int{10, 50, 100, 200}[rand.Intn(4)]
		}

		rule := &IPThrottleRule{
			IPAddress:  fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256)),
			Action:     action,
			RateLimit:  rateLimit,
			Reason:     ipReasons[rand.Intn(len(ipReasons))],
			CreatedAt:  time.Now().AddDate(0, 0, -rand.Intn(180)),
			UpdatedAt:  time.Now().Add(-time.Duration(rand.Intn(48)) * time.Hour),
			HitCount:   rand.Intn(10000),
			LastHit:    time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second),
		}

		s.ipRules[rule.IPAddress] = rule
	}

	// Generate 5 geographic rules
	geoRegions := []string{"North America", "Europe", "Asia Pacific", "Middle East", "Africa"}
	geoCodes := []string{"US,CA,MX", "GB,DE,FR,IT,ES", "JP,CN,AU,SG,IN", "AE,SA,QA", "ZA,EG,NG"}
	geoActions := []string{"allow", "allow", "allow", "throttle", "block"}

	for i := 0; i < 5; i++ {
		rule := &GeographicRule{
			ID:          s.nextGeoRuleID,
			Region:      geoRegions[i],
			CountryCode: geoCodes[i],
			Action:      geoActions[i],
			Priority:    i + 1,
			CreatedAt:   time.Now().AddDate(0, 0, -rand.Intn(90)),
			UpdatedAt:   time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
		}

		s.geoRules = append(s.geoRules, rule)
		s.nextGeoRuleID++
	}

	log.Println("[APIGateway] Mock data initialized: 20 clients, 8 tiers, 1000 usage records, 10 IP rules, 5 geo rules")
}

func generateRandomKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// GetClients returns all API clients with usage stats
func (s *APIGatewayService) GetClients() []*APIClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*APIClient, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	return clients
}

// GetClient returns a specific API client with recent usage
func (s *APIGatewayService) GetClient(id int64) (*APIClient, []*APIUsageRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.clients[id]
	if !exists {
		return nil, nil, fmt.Errorf("client not found")
	}

	// Get recent usage records for this client
	recentRecords := make([]*APIUsageRecord, 0)
	for _, record := range s.usageRecords {
		if record.ClientID == id {
			recentRecords = append(recentRecords, record)
			if len(recentRecords) >= 50 {
				break
			}
		}
	}

	return client, recentRecords, nil
}

// UpdateClient updates client rate limits and quotas
func (s *APIGatewayService) UpdateClient(id int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[id]
	if !exists {
		return fmt.Errorf("client not found")
	}

	if tier, ok := updates["tier"].(string); ok {
		if tierConfig, exists := s.rateLimitTiers[tier]; exists {
			client.Tier = tier
			client.DailyQuota = tierConfig.DailyQuota
			client.MonthlyQuota = tierConfig.MonthlyQuota
			client.RateLimitPerMin = tierConfig.RequestsPerMin
		}
	}
	if status, ok := updates["status"].(string); ok {
		client.Status = status
	}
	if ipWhitelist, ok := updates["ip_whitelist"].([]interface{}); ok {
		client.IPWhitelist = make([]string, len(ipWhitelist))
		for i, ip := range ipWhitelist {
			client.IPWhitelist[i] = ip.(string)
		}
	}

	return nil
}

// GetRateLimitTiers returns all rate limit tier configurations
func (s *APIGatewayService) GetRateLimitTiers() []*RateLimitTier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tiers := make([]*RateLimitTier, 0, len(s.rateLimitTiers))
	for _, tier := range s.rateLimitTiers {
		tiers = append(tiers, tier)
	}
	return tiers
}

// UpdateRateLimitTier updates a rate limit tier configuration
func (s *APIGatewayService) UpdateRateLimitTier(tierName string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tier, exists := s.rateLimitTiers[tierName]
	if !exists {
		return fmt.Errorf("tier not found")
	}

	if rpm, ok := updates["requests_per_min"].(float64); ok {
		tier.RequestsPerMin = int(rpm)
	}
	if daily, ok := updates["daily_quota"].(float64); ok {
		tier.DailyQuota = int(daily)
	}
	if monthly, ok := updates["monthly_quota"].(float64); ok {
		tier.MonthlyQuota = int(monthly)
	}
	if burst, ok := updates["burst_limit"].(float64); ok {
		tier.BurstLimit = int(burst)
	}

	return nil
}

// GetUsageAnalytics returns global API usage analytics
func (s *APIGatewayService) GetUsageAnalytics() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalRequests := 0
	totalErrors := 0
	endpointCounts := make(map[string]int)
	statusCodeCounts := make(map[int]int)
	var totalResponseTime float64

	for _, record := range s.usageRecords {
		totalRequests++
		if record.StatusCode >= 400 {
			totalErrors++
		}
		endpointCounts[record.Endpoint]++
		statusCodeCounts[record.StatusCode]++
		totalResponseTime += record.ResponseTimeMs
	}

	avgResponseTime := 0.0
	if totalRequests > 0 {
		avgResponseTime = totalResponseTime / float64(totalRequests)
	}

	// Top 10 endpoints by volume
	type endpointStat struct {
		Endpoint string `json:"endpoint"`
		Count    int    `json:"count"`
	}
	topEndpoints := make([]endpointStat, 0)
	for endpoint, count := range endpointCounts {
		topEndpoints = append(topEndpoints, endpointStat{Endpoint: endpoint, Count: count})
	}

	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"total_requests":       totalRequests,
		"total_errors":         totalErrors,
		"error_rate":           errorRate,
		"avg_response_time_ms": avgResponseTime,
		"top_endpoints":        topEndpoints,
		"status_code_breakdown": statusCodeCounts,
		"active_clients":       len(s.clients),
	}
}

// GetIPRules returns all IP throttle rules
func (s *APIGatewayService) GetIPRules() []*IPThrottleRule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rules := make([]*IPThrottleRule, 0, len(s.ipRules))
	for _, rule := range s.ipRules {
		rules = append(rules, rule)
	}
	return rules
}

// UpdateIPRule updates an IP throttle rule
func (s *APIGatewayService) UpdateIPRule(ip string, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rule, exists := s.ipRules[ip]
	if !exists {
		// Create new rule
		rule = &IPThrottleRule{
			IPAddress: ip,
			CreatedAt: time.Now(),
		}
		s.ipRules[ip] = rule
	}

	if action, ok := updates["action"].(string); ok {
		rule.Action = action
	}
	if rateLimit, ok := updates["rate_limit"].(float64); ok {
		rule.RateLimit = int(rateLimit)
	}
	if reason, ok := updates["reason"].(string); ok {
		rule.Reason = reason
	}
	rule.UpdatedAt = time.Now()

	return nil
}

// APIGatewayHandler handles HTTP requests for API gateway management
type APIGatewayHandler struct {
	service     *APIGatewayService
	authService *AuthService
}

// NewAPIGatewayHandler creates a new API gateway handler
func NewAPIGatewayHandler(service *APIGatewayService, authService *AuthService) *APIGatewayHandler {
	return &APIGatewayHandler{
		service:     service,
		authService: authService,
	}
}

// HandleListClients handles GET /admin/api-gateway/clients
func (h *APIGatewayHandler) HandleListClients(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	clients := h.service.GetClients()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"total":   len(clients),
	})
}

// HandleGetClient handles GET /admin/api-gateway/clients/:id
func (h *APIGatewayHandler) HandleGetClient(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/api-gateway/clients/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	client, recentUsage, err := h.service.GetClient(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"client":       client,
		"recent_usage": recentUsage,
	})
}

// HandleUpdateClient handles PUT /admin/api-gateway/clients/:id
func (h *APIGatewayHandler) HandleUpdateClient(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/admin/api-gateway/clients/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateClient(id, updates); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Client updated successfully",
	})
}

// HandleListRateLimits handles GET /admin/api-gateway/rate-limits
func (h *APIGatewayHandler) HandleListRateLimits(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tiers := h.service.GetRateLimitTiers()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers": tiers,
		"total": len(tiers),
	})
}

// HandleUpdateRateLimit handles PUT /admin/api-gateway/rate-limits/:tier
func (h *APIGatewayHandler) HandleUpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tierName := strings.TrimPrefix(r.URL.Path, "/admin/api-gateway/rate-limits/")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateRateLimitTier(tierName, updates); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Rate limit tier updated successfully",
	})
}

// HandleGetUsage handles GET /admin/api-gateway/usage
func (h *APIGatewayHandler) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	analytics := h.service.GetUsageAnalytics()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(analytics)
}

// HandleListIPRules handles GET /admin/api-gateway/ip-rules
func (h *APIGatewayHandler) HandleListIPRules(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rules := h.service.GetIPRules()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ip_rules": rules,
		"total":    len(rules),
	})
}

// HandleUpdateIPRule handles PUT /admin/api-gateway/ip-rules/:ip
func (h *APIGatewayHandler) HandleUpdateIPRule(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := auth.ValidateTokenWithDefault(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ip := strings.TrimPrefix(r.URL.Path, "/admin/api-gateway/ip-rules/")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateIPRule(ip, updates); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "IP rule updated successfully",
	})
}
