package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================
// Data Structures
// ============================================

// ExecutionPolicy represents an order execution policy
type ExecutionPolicy struct {
	ID                int64             `json:"id"`
	Name              string            `json:"name"`     // "market", "instant", "exchange"
	DisplayName       string            `json:"display_name"`
	Description       string            `json:"description"`
	AllowRequotes     bool              `json:"allow_requotes"`
	MaxDeviation      float64           `json:"max_deviation"` // pips
	TimeoutMs         int               `json:"timeout_ms"`
	PartialFillAllowed bool             `json:"partial_fill_allowed"`
	Rules             map[string]interface{} `json:"rules"`
	IsActive          bool              `json:"is_active"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// SlippageConfig represents slippage tolerance settings per symbol group
type SlippageConfig struct {
	GroupID           int64   `json:"group_id"`
	GroupName         string  `json:"group_name"`
	MaxPositiveSlippage float64 `json:"max_positive_slippage"` // pips
	MaxNegativeSlippage float64 `json:"max_negative_slippage"` // pips
	WarningThreshold    float64 `json:"warning_threshold"`     // pips
	AutoRejectEnabled   bool    `json:"auto_reject_enabled"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// RequoteConfig represents requote settings
type RequoteConfig struct {
	Enabled           bool    `json:"enabled"`
	MaxDeviation      float64 `json:"max_deviation"`  // pips
	TimeoutMs         int     `json:"timeout_ms"`
	MaxAttempts       int     `json:"max_attempts"`
	ApplyToMarketOrders bool  `json:"apply_to_market_orders"`
	ApplyToLimitOrders  bool  `json:"apply_to_limit_orders"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// PolicyExecutionStats represents execution performance statistics
type PolicyExecutionStats struct {
	TotalExecutions     int                       `json:"total_executions"`
	AvgFillTimeMs       float64                   `json:"avg_fill_time_ms"`
	RequoteRate         float64                   `json:"requote_rate"`        // percentage
	FillRateByPolicy    map[string]float64        `json:"fill_rate_by_policy"` // policy name -> fill rate %
	SlippageDistribution PolicySlippageDistribution     `json:"slippage_distribution"`
	ExecutionsByHour    map[int]int               `json:"executions_by_hour"`
	TopSymbols          []SymbolExecutionStat     `json:"top_symbols"`
	PeriodStart         time.Time                 `json:"period_start"`
	PeriodEnd           time.Time                 `json:"period_end"`
}

// PolicySlippageDistribution represents slippage statistics
type PolicySlippageDistribution struct {
	AvgSlippage       float64            `json:"avg_slippage"`        // pips
	MinSlippage       float64            `json:"min_slippage"`        // pips
	MaxSlippage       float64            `json:"max_slippage"`        // pips
	PositiveSlippage  int                `json:"positive_slippage_count"`
	NegativeSlippage  int                `json:"negative_slippage_count"`
	ZeroSlippage      int                `json:"zero_slippage_count"`
	SlippageRanges    map[string]int     `json:"slippage_ranges"`     // "0-1", "1-2", etc.
}

// SymbolExecutionStat represents execution stats per symbol
type SymbolExecutionStat struct {
	Symbol          string  `json:"symbol"`
	ExecutionCount  int     `json:"execution_count"`
	AvgFillTimeMs   float64 `json:"avg_fill_time_ms"`
	AvgSlippage     float64 `json:"avg_slippage"`
	FillRate        float64 `json:"fill_rate"`
}

// PolicyExecutionRecord represents a single execution for statistics
type PolicyExecutionRecord struct {
	ID            int64     `json:"id"`
	OrderID       int64     `json:"order_id"`
	Symbol        string    `json:"symbol"`
	PolicyName    string    `json:"policy_name"`
	FillTimeMs    int       `json:"fill_time_ms"`
	Slippage      float64   `json:"slippage"` // pips
	RequoteCount  int       `json:"requote_count"`
	Filled        bool      `json:"filled"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// ============================================
// Service
// ============================================

// ExecutionPolicyService manages execution policies and configurations
type ExecutionPolicyService struct {
	mu                sync.RWMutex
	policies          map[int64]*ExecutionPolicy
	slippageConfigs   map[int64]*SlippageConfig
	requoteConfig     *RequoteConfig
	executionRecords  []PolicyExecutionRecord
}

// NewExecutionPolicyService creates a new execution policy service with mock data
func NewExecutionPolicyService() *ExecutionPolicyService {
	s := &ExecutionPolicyService{
		policies:        make(map[int64]*ExecutionPolicy),
		slippageConfigs: make(map[int64]*SlippageConfig),
		executionRecords: make([]PolicyExecutionRecord, 0),
	}

	// Initialize 3 execution policies
	s.initPolicies()

	// Initialize 8 symbol groups with slippage configs
	s.initSlippageConfigs()

	// Initialize requote config
	s.initRequoteConfig()

	// Generate 5000 execution records for stats
	s.generateExecutionRecords(5000)

	log.Printf("[ExecutionPolicyService] Initialized with %d policies, %d slippage configs, %d execution records",
		len(s.policies), len(s.slippageConfigs), len(s.executionRecords))

	return s
}

func (s *ExecutionPolicyService) initPolicies() {
	now := time.Now()

	policies := []*ExecutionPolicy{
		{
			ID:          1,
			Name:        "market",
			DisplayName: "Market Execution",
			Description: "Orders executed at best available market price with possible slippage",
			AllowRequotes: false,
			MaxDeviation: 5.0,
			TimeoutMs:    3000,
			PartialFillAllowed: true,
			Rules: map[string]interface{}{
				"execution_mode": "fill_or_kill",
				"price_improvement_allowed": true,
				"max_slippage_pips": 5.0,
			},
			IsActive:  true,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:          2,
			Name:        "instant",
			DisplayName: "Instant Execution",
			Description: "Orders executed at requested price or requoted if price moved",
			AllowRequotes: true,
			MaxDeviation: 2.0,
			TimeoutMs:    1500,
			PartialFillAllowed: false,
			Rules: map[string]interface{}{
				"execution_mode": "all_or_nothing",
				"requote_on_deviation": true,
				"max_requote_attempts": 3,
			},
			IsActive:  true,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-10 * 24 * time.Hour),
		},
		{
			ID:          3,
			Name:        "exchange",
			DisplayName: "Exchange Execution",
			Description: "Orders routed to external liquidity provider or exchange",
			AllowRequotes: false,
			MaxDeviation: 3.0,
			TimeoutMs:    5000,
			PartialFillAllowed: true,
			Rules: map[string]interface{}{
				"execution_mode": "best_execution",
				"route_to_lp": true,
				"lp_timeout_ms": 4000,
				"fallback_to_internal": false,
			},
			IsActive:  true,
			CreatedAt: now.Add(-180 * 24 * time.Hour),
			UpdatedAt: now.Add(-15 * 24 * time.Hour),
		},
	}

	for _, p := range policies {
		s.policies[p.ID] = p
	}
}

func (s *ExecutionPolicyService) initSlippageConfigs() {
	now := time.Now()

	groups := []struct {
		ID   int64
		Name string
		MaxPos float64
		MaxNeg float64
		Warn float64
		AutoReject bool
	}{
		{1, "Forex Majors", 3.0, 3.0, 2.0, false},
		{2, "Forex Minors", 5.0, 5.0, 3.0, false},
		{3, "Forex Exotics", 10.0, 10.0, 7.0, true},
		{4, "Crypto", 50.0, 50.0, 30.0, false},
		{5, "Indices", 5.0, 5.0, 3.0, false},
		{6, "Commodities", 8.0, 8.0, 5.0, false},
		{7, "Metals", 10.0, 10.0, 6.0, false},
		{8, "Energy", 15.0, 15.0, 10.0, true},
	}

	for _, g := range groups {
		s.slippageConfigs[g.ID] = &SlippageConfig{
			GroupID:             g.ID,
			GroupName:           g.Name,
			MaxPositiveSlippage: g.MaxPos,
			MaxNegativeSlippage: g.MaxNeg,
			WarningThreshold:    g.Warn,
			AutoRejectEnabled:   g.AutoReject,
			UpdatedAt:           now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour),
		}
	}
}

func (s *ExecutionPolicyService) initRequoteConfig() {
	s.requoteConfig = &RequoteConfig{
		Enabled:             true,
		MaxDeviation:        2.0,
		TimeoutMs:           1500,
		MaxAttempts:         3,
		ApplyToMarketOrders: false,
		ApplyToLimitOrders:  true,
		UpdatedAt:           time.Now().Add(-7 * 24 * time.Hour),
	}
}

func (s *ExecutionPolicyService) generateExecutionRecords(count int) {
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD", "XAUUSD", "BTCUSD", "ETHUSD", "SPX500", "US30"}
	policies := []string{"market", "instant", "exchange"}

	now := time.Now()
	startDate := now.Add(-30 * 24 * time.Hour)

	for i := 0; i < count; i++ {
		// Random execution time within last 30 days
		randomOffset := time.Duration(rand.Int63n(int64(30 * 24 * time.Hour)))
		executedAt := startDate.Add(randomOffset)

		// Random slippage (-10 to +10 pips, with bias toward small values)
		slippage := (rand.Float64() - 0.5) * 20 * rand.Float64()

		// Random fill time (50ms to 5000ms, with bias toward fast fills)
		fillTime := 50 + int(rand.Float64()*rand.Float64()*4950)

		// Requote count (0-3, with bias toward 0)
		requoteCount := 0
		if rand.Float64() < 0.15 { // 15% chance of requote
			requoteCount = 1 + rand.Intn(3)
		}

		// Fill success rate: 95%
		filled := rand.Float64() < 0.95

		record := PolicyExecutionRecord{
			ID:           int64(i + 1),
			OrderID:      int64(100000 + i),
			Symbol:       symbols[rand.Intn(len(symbols))],
			PolicyName:   policies[rand.Intn(len(policies))],
			FillTimeMs:   fillTime,
			Slippage:     slippage,
			RequoteCount: requoteCount,
			Filled:       filled,
			ExecutedAt:   executedAt,
		}

		s.executionRecords = append(s.executionRecords, record)
	}
}

// GetAllPolicies returns all execution policies
func (s *ExecutionPolicyService) GetAllPolicies() []*ExecutionPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*ExecutionPolicy, 0, len(s.policies))
	for _, p := range s.policies {
		policies = append(policies, p)
	}
	return policies
}

// GetPolicyByID returns a specific execution policy
func (s *ExecutionPolicyService) GetPolicyByID(id int64) *ExecutionPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.policies[id]
}

// UpdatePolicy updates an execution policy
func (s *ExecutionPolicyService) UpdatePolicy(id int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	policy, exists := s.policies[id]
	if !exists {
		return nil
	}

	// Apply updates
	if displayName, ok := updates["display_name"].(string); ok {
		policy.DisplayName = displayName
	}
	if description, ok := updates["description"].(string); ok {
		policy.Description = description
	}
	if allowRequotes, ok := updates["allow_requotes"].(bool); ok {
		policy.AllowRequotes = allowRequotes
	}
	if maxDeviation, ok := updates["max_deviation"].(float64); ok {
		policy.MaxDeviation = maxDeviation
	}
	if timeoutMs, ok := updates["timeout_ms"].(float64); ok {
		policy.TimeoutMs = int(timeoutMs)
	}
	if partialFill, ok := updates["partial_fill_allowed"].(bool); ok {
		policy.PartialFillAllowed = partialFill
	}
	if rules, ok := updates["rules"].(map[string]interface{}); ok {
		policy.Rules = rules
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		policy.IsActive = isActive
	}

	policy.UpdatedAt = time.Now()
	return nil
}

// GetAllSlippageConfigs returns all slippage configurations
func (s *ExecutionPolicyService) GetAllSlippageConfigs() []*SlippageConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := make([]*SlippageConfig, 0, len(s.slippageConfigs))
	for _, c := range s.slippageConfigs {
		configs = append(configs, c)
	}
	return configs
}

// UpdateSlippageConfig updates slippage configuration for a group
func (s *ExecutionPolicyService) UpdateSlippageConfig(groupID int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config, exists := s.slippageConfigs[groupID]
	if !exists {
		return nil
	}

	// Apply updates
	if maxPos, ok := updates["max_positive_slippage"].(float64); ok {
		config.MaxPositiveSlippage = maxPos
	}
	if maxNeg, ok := updates["max_negative_slippage"].(float64); ok {
		config.MaxNegativeSlippage = maxNeg
	}
	if warn, ok := updates["warning_threshold"].(float64); ok {
		config.WarningThreshold = warn
	}
	if autoReject, ok := updates["auto_reject_enabled"].(bool); ok {
		config.AutoRejectEnabled = autoReject
	}

	config.UpdatedAt = time.Now()
	return nil
}

// GetRequoteConfig returns current requote configuration
func (s *ExecutionPolicyService) GetRequoteConfig() *RequoteConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.requoteConfig
}

// UpdateRequoteConfig updates requote configuration
func (s *ExecutionPolicyService) UpdateRequoteConfig(updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Apply updates
	if enabled, ok := updates["enabled"].(bool); ok {
		s.requoteConfig.Enabled = enabled
	}
	if maxDev, ok := updates["max_deviation"].(float64); ok {
		s.requoteConfig.MaxDeviation = maxDev
	}
	if timeoutMs, ok := updates["timeout_ms"].(float64); ok {
		s.requoteConfig.TimeoutMs = int(timeoutMs)
	}
	if maxAttempts, ok := updates["max_attempts"].(float64); ok {
		s.requoteConfig.MaxAttempts = int(maxAttempts)
	}
	if applyMarket, ok := updates["apply_to_market_orders"].(bool); ok {
		s.requoteConfig.ApplyToMarketOrders = applyMarket
	}
	if applyLimit, ok := updates["apply_to_limit_orders"].(bool); ok {
		s.requoteConfig.ApplyToLimitOrders = applyLimit
	}

	s.requoteConfig.UpdatedAt = time.Now()
	return nil
}

// GetExecutionStats calculates execution statistics
func (s *ExecutionPolicyService) GetExecutionStats() PolicyExecutionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.executionRecords) == 0 {
		return PolicyExecutionStats{}
	}

	// Calculate stats
	var totalFillTime int64
	requoteCount := 0
	fillsByPolicy := make(map[string]int)
	totalByPolicy := make(map[string]int)
	slippages := make([]float64, 0)
	positiveSlippage := 0
	negativeSlippage := 0
	zeroSlippage := 0
	slippageRanges := make(map[string]int)
	execsByHour := make(map[int]int)
	symbolStats := make(map[string]*SymbolExecutionStat)

	var minSlippage, maxSlippage float64 = 1000, -1000
	var periodStart, periodEnd time.Time

	for i, rec := range s.executionRecords {
		if i == 0 {
			periodStart = rec.ExecutedAt
			periodEnd = rec.ExecutedAt
		} else {
			if rec.ExecutedAt.Before(periodStart) {
				periodStart = rec.ExecutedAt
			}
			if rec.ExecutedAt.After(periodEnd) {
				periodEnd = rec.ExecutedAt
			}
		}

		totalFillTime += int64(rec.FillTimeMs)

		if rec.RequoteCount > 0 {
			requoteCount++
		}

		totalByPolicy[rec.PolicyName]++
		if rec.Filled {
			fillsByPolicy[rec.PolicyName]++
		}

		slippages = append(slippages, rec.Slippage)
		if rec.Slippage > 0 {
			positiveSlippage++
		} else if rec.Slippage < 0 {
			negativeSlippage++
		} else {
			zeroSlippage++
		}

		if rec.Slippage < minSlippage {
			minSlippage = rec.Slippage
		}
		if rec.Slippage > maxSlippage {
			maxSlippage = rec.Slippage
		}

		// Slippage ranges
		absSlip := rec.Slippage
		if absSlip < 0 {
			absSlip = -absSlip
		}
		rangeKey := "0-1"
		if absSlip >= 10 {
			rangeKey = "10+"
		} else if absSlip >= 5 {
			rangeKey = "5-10"
		} else if absSlip >= 3 {
			rangeKey = "3-5"
		} else if absSlip >= 1 {
			rangeKey = "1-3"
		}
		slippageRanges[rangeKey]++

		// Executions by hour
		hour := rec.ExecutedAt.Hour()
		execsByHour[hour]++

		// Symbol stats
		if _, exists := symbolStats[rec.Symbol]; !exists {
			symbolStats[rec.Symbol] = &SymbolExecutionStat{
				Symbol: rec.Symbol,
			}
		}
		stat := symbolStats[rec.Symbol]
		stat.ExecutionCount++
		stat.AvgFillTimeMs += float64(rec.FillTimeMs)
		stat.AvgSlippage += rec.Slippage
		if rec.Filled {
			stat.FillRate++
		}
	}

	// Calculate averages
	avgFillTime := float64(totalFillTime) / float64(len(s.executionRecords))
	requoteRate := float64(requoteCount) / float64(len(s.executionRecords)) * 100

	fillRateByPolicy := make(map[string]float64)
	for policy, total := range totalByPolicy {
		if total > 0 {
			fillRateByPolicy[policy] = float64(fillsByPolicy[policy]) / float64(total) * 100
		}
	}

	avgSlippage := 0.0
	for _, s := range slippages {
		avgSlippage += s
	}
	avgSlippage /= float64(len(slippages))

	// Top symbols
	topSymbols := make([]SymbolExecutionStat, 0)
	for _, stat := range symbolStats {
		if stat.ExecutionCount > 0 {
			stat.AvgFillTimeMs /= float64(stat.ExecutionCount)
			stat.AvgSlippage /= float64(stat.ExecutionCount)
			stat.FillRate = stat.FillRate / float64(stat.ExecutionCount) * 100
		}
		topSymbols = append(topSymbols, *stat)
	}

	return PolicyExecutionStats{
		TotalExecutions:  len(s.executionRecords),
		AvgFillTimeMs:    avgFillTime,
		RequoteRate:      requoteRate,
		FillRateByPolicy: fillRateByPolicy,
		SlippageDistribution: PolicySlippageDistribution{
			AvgSlippage:      avgSlippage,
			MinSlippage:      minSlippage,
			MaxSlippage:      maxSlippage,
			PositiveSlippage: positiveSlippage,
			NegativeSlippage: negativeSlippage,
			ZeroSlippage:     zeroSlippage,
			SlippageRanges:   slippageRanges,
		},
		ExecutionsByHour: execsByHour,
		TopSymbols:       topSymbols,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
	}
}

// ============================================
// HTTP Handlers
// ============================================

// ExecutionPolicyHandler handles execution policy HTTP requests
type ExecutionPolicyHandler struct {
	service     *ExecutionPolicyService
	authService interface {
		ValidateAdminToken(r *http.Request) (int64, error)
	}
}

// NewExecutionPolicyHandler creates a new execution policy handler
func NewExecutionPolicyHandler(service *ExecutionPolicyService, authService interface {
	ValidateAdminToken(r *http.Request) (int64, error)
}) *ExecutionPolicyHandler {
	return &ExecutionPolicyHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetPolicies handles GET /admin/execution/policies
func (h *ExecutionPolicyHandler) HandleGetPolicies(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	policies := h.service.GetAllPolicies()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(policies)
}

// HandleGetPolicyByID handles GET /admin/execution/policies/:id
func (h *ExecutionPolicyHandler) HandleGetPolicyByID(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/execution/policies/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid policy ID", http.StatusBadRequest)
		return
	}

	policy := h.service.GetPolicyByID(id)
	if policy == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(policy)
}

// HandleUpdatePolicy handles PUT /admin/execution/policies/:id
func (h *ExecutionPolicyHandler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/execution/policies/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid policy ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdatePolicy(id, updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Policy updated successfully",
	})
}

// HandleGetSlippageConfigs handles GET /admin/execution/slippage-config
func (h *ExecutionPolicyHandler) HandleGetSlippageConfigs(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	configs := h.service.GetAllSlippageConfigs()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(configs)
}

// HandleUpdateSlippageConfig handles PUT /admin/execution/slippage-config/:groupId
func (h *ExecutionPolicyHandler) HandleUpdateSlippageConfig(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract group ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/execution/slippage-config/")
	groupID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSlippageConfig(groupID, updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Slippage config updated successfully",
	})
}

// HandleGetRequoteConfig handles GET /admin/execution/requote-config
func (h *ExecutionPolicyHandler) HandleGetRequoteConfig(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	config := h.service.GetRequoteConfig()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(config)
}

// HandleUpdateRequoteConfig handles PUT /admin/execution/requote-config
func (h *ExecutionPolicyHandler) HandleUpdateRequoteConfig(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateRequoteConfig(updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Requote config updated successfully",
	})
}

// HandleGetExecutionStats handles GET /admin/execution/stats
func (h *ExecutionPolicyHandler) HandleGetExecutionStats(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetExecutionStats()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(stats)
}
