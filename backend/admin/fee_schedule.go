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

// FeeSchedule represents a complete fee schedule
type FeeSchedule struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"` // "Standard", "Premium", "VIP", "Institutional", "Islamic", "Demo"
	Description string            `json:"description"`
	FeeRules    map[string]FeeRule `json:"fee_rules"` // key: fee_type
	IsActive    bool              `json:"is_active"`
	ClientCount int               `json:"client_count"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// FeeRule represents a single fee rule within a schedule
type FeeRule struct {
	FeeType     string  `json:"fee_type"` // "commission_per_lot", "spread_markup_pips", "swap_rate_annual", "deposit_fee_pct", "withdrawal_fee_flat", "inactivity_fee_monthly"
	Value       float64 `json:"value"`
	Currency    string  `json:"currency,omitempty"` // for flat fees
	Description string  `json:"description"`
	IsEnabled   bool    `json:"is_enabled"`
}

// ClientFeeMapping represents fee schedule assigned to a client
type ClientFeeMapping struct {
	ClientID     int64     `json:"client_id"`
	ClientName   string    `json:"client_name"`
	ScheduleID   int64     `json:"schedule_id"`
	ScheduleName string    `json:"schedule_name"`
	AssignedAt   time.Time `json:"assigned_at"`
}

// ClientFeeDetails represents fees applied to a specific client
type ClientFeeDetails struct {
	ClientID        int64             `json:"client_id"`
	ClientName      string            `json:"client_name"`
	ScheduleID      int64             `json:"schedule_id"`
	ScheduleName    string            `json:"schedule_name"`
	AppliedFees     map[string]FeeRule `json:"applied_fees"`
	ActiveWaivers   []FeeWaiver       `json:"active_waivers"`
	EstimatedMonthly float64          `json:"estimated_monthly_fees"` // USD
}

// FeeRevenue represents fee revenue analytics
type FeeRevenue struct {
	TotalCollected    float64                `json:"total_collected"` // USD
	ByFeeType         map[string]float64     `json:"by_fee_type"`
	MonthlyBreakdown  []FeeMonthlyRevenue     `json:"monthly_breakdown"` // last 12 months
	TopClients        []ClientRevenueContrib `json:"top_clients"`
	PeriodStart       time.Time              `json:"period_start"`
	PeriodEnd         time.Time              `json:"period_end"`
}

// FeeMonthlyRevenue represents revenue for a specific month
type FeeMonthlyRevenue struct {
	Month          string             `json:"month"` // "2026-01"
	TotalRevenue   float64            `json:"total_revenue"`
	ByFeeType      map[string]float64 `json:"by_fee_type"`
	ClientCount    int                `json:"client_count"`
}

// ClientRevenueContrib represents client's contribution to fee revenue
type ClientRevenueContrib struct {
	ClientID    int64              `json:"client_id"`
	ClientName  string             `json:"client_name"`
	TotalPaid   float64            `json:"total_paid"`
	ByFeeType   map[string]float64 `json:"by_fee_type"`
}

// FeeWaiver represents a fee waiver for a client
type FeeWaiver struct {
	ID         int64     `json:"id"`
	ClientID   int64     `json:"client_id"`
	ClientName string    `json:"client_name"`
	FeeType    string    `json:"fee_type"`
	Reason     string    `json:"reason"`
	StartDate  time.Time `json:"start_date"`
	ExpiryDate time.Time `json:"expiry_date"`
	WaivedBy   string    `json:"waived_by"` // admin username
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateFeeWaiverRequest represents request to create a fee waiver
type CreateFeeWaiverRequest struct {
	ClientID     int64  `json:"client_id"`
	FeeType      string `json:"fee_type"`
	Reason       string `json:"reason"`
	DurationDays int    `json:"duration_days"`
}

// CreateFeeScheduleRequest represents request to create a new fee schedule
type CreateFeeScheduleRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	FeeRules    map[string]FeeRule `json:"fee_rules"`
}

// ============================================
// Service
// ============================================

// FeeScheduleService manages fee schedules and billing
type FeeScheduleService struct {
	mu               sync.RWMutex
	schedules        map[int64]*FeeSchedule
	clientMappings   map[int64]*ClientFeeMapping // key: clientID
	waivers          map[int64]*FeeWaiver
	monthlyRevenue   []FeeMonthlyRevenue
	nextScheduleID   int64
	nextWaiverID     int64
}

// NewFeeScheduleService creates a new fee schedule service with mock data
func NewFeeScheduleService() *FeeScheduleService {
	s := &FeeScheduleService{
		schedules:      make(map[int64]*FeeSchedule),
		clientMappings: make(map[int64]*ClientFeeMapping),
		waivers:        make(map[int64]*FeeWaiver),
		nextScheduleID: 7,
		nextWaiverID:   31,
	}

	// Initialize 6 fee schedules
	s.initSchedules()

	// Map 200 clients to schedules
	s.initClientMappings()

	// Generate 12-month revenue data
	s.initRevenueData()

	// Create 30 active fee waivers
	s.initWaivers()

	log.Printf("[FeeScheduleService] Initialized with %d fee schedules, %d clients, 12-month revenue data, %d waivers",
		len(s.schedules), len(s.clientMappings), len(s.waivers))

	return s
}

func (s *FeeScheduleService) initSchedules() {
	now := time.Now()

	schedules := []*FeeSchedule{
		{
			ID:          1,
			Name:        "Standard",
			Description: "Standard retail account fee structure",
			FeeRules: map[string]FeeRule{
				"commission_per_lot":      {FeeType: "commission_per_lot", Value: 7.0, Currency: "USD", Description: "Commission per standard lot", IsEnabled: true},
				"spread_markup_pips":      {FeeType: "spread_markup_pips", Value: 0.5, Description: "Spread markup in pips", IsEnabled: true},
				"swap_rate_annual":        {FeeType: "swap_rate_annual", Value: 5.0, Description: "Annual swap rate percentage", IsEnabled: true},
				"deposit_fee_pct":         {FeeType: "deposit_fee_pct", Value: 0.0, Description: "Deposit fee percentage", IsEnabled: false},
				"withdrawal_fee_flat":     {FeeType: "withdrawal_fee_flat", Value: 25.0, Currency: "USD", Description: "Flat withdrawal fee", IsEnabled: true},
				"inactivity_fee_monthly":  {FeeType: "inactivity_fee_monthly", Value: 10.0, Currency: "USD", Description: "Monthly inactivity fee after 6 months", IsEnabled: true},
			},
			IsActive:    true,
			ClientCount: 80,
			CreatedAt:   now.Add(-365 * 24 * time.Hour),
			UpdatedAt:   now.Add(-30 * 24 * time.Hour),
		},
		{
			ID:          2,
			Name:        "Premium",
			Description: "Premium account with reduced fees",
			FeeRules: map[string]FeeRule{
				"commission_per_lot":      {FeeType: "commission_per_lot", Value: 5.0, Currency: "USD", Description: "Commission per standard lot", IsEnabled: true},
				"spread_markup_pips":      {FeeType: "spread_markup_pips", Value: 0.3, Description: "Spread markup in pips", IsEnabled: true},
				"swap_rate_annual":        {FeeType: "swap_rate_annual", Value: 4.5, Description: "Annual swap rate percentage", IsEnabled: true},
				"deposit_fee_pct":         {FeeType: "deposit_fee_pct", Value: 0.0, Description: "Deposit fee percentage", IsEnabled: false},
				"withdrawal_fee_flat":     {FeeType: "withdrawal_fee_flat", Value: 15.0, Currency: "USD", Description: "Flat withdrawal fee", IsEnabled: true},
				"inactivity_fee_monthly":  {FeeType: "inactivity_fee_monthly", Value: 0.0, Currency: "USD", Description: "No inactivity fee", IsEnabled: false},
			},
			IsActive:    true,
			ClientCount: 60,
			CreatedAt:   now.Add(-365 * 24 * time.Hour),
			UpdatedAt:   now.Add(-20 * 24 * time.Hour),
		},
		{
			ID:          3,
			Name:        "VIP",
			Description: "VIP account with minimal fees",
			FeeRules: map[string]FeeRule{
				"commission_per_lot":      {FeeType: "commission_per_lot", Value: 3.0, Currency: "USD", Description: "Commission per standard lot", IsEnabled: true},
				"spread_markup_pips":      {FeeType: "spread_markup_pips", Value: 0.1, Description: "Spread markup in pips", IsEnabled: true},
				"swap_rate_annual":        {FeeType: "swap_rate_annual", Value: 4.0, Description: "Annual swap rate percentage", IsEnabled: true},
				"deposit_fee_pct":         {FeeType: "deposit_fee_pct", Value: 0.0, Description: "Deposit fee percentage", IsEnabled: false},
				"withdrawal_fee_flat":     {FeeType: "withdrawal_fee_flat", Value: 0.0, Currency: "USD", Description: "Free withdrawals", IsEnabled: false},
				"inactivity_fee_monthly":  {FeeType: "inactivity_fee_monthly", Value: 0.0, Currency: "USD", Description: "No inactivity fee", IsEnabled: false},
			},
			IsActive:    true,
			ClientCount: 30,
			CreatedAt:   now.Add(-365 * 24 * time.Hour),
			UpdatedAt:   now.Add(-15 * 24 * time.Hour),
		},
		{
			ID:          4,
			Name:        "Institutional",
			Description: "Institutional account with volume-based pricing",
			FeeRules: map[string]FeeRule{
				"commission_per_lot":      {FeeType: "commission_per_lot", Value: 1.5, Currency: "USD", Description: "Commission per standard lot", IsEnabled: true},
				"spread_markup_pips":      {FeeType: "spread_markup_pips", Value: 0.0, Description: "Raw spread, no markup", IsEnabled: false},
				"swap_rate_annual":        {FeeType: "swap_rate_annual", Value: 3.5, Description: "Annual swap rate percentage", IsEnabled: true},
				"deposit_fee_pct":         {FeeType: "deposit_fee_pct", Value: 0.0, Description: "Deposit fee percentage", IsEnabled: false},
				"withdrawal_fee_flat":     {FeeType: "withdrawal_fee_flat", Value: 0.0, Currency: "USD", Description: "Free withdrawals", IsEnabled: false},
				"inactivity_fee_monthly":  {FeeType: "inactivity_fee_monthly", Value: 0.0, Currency: "USD", Description: "No inactivity fee", IsEnabled: false},
			},
			IsActive:    true,
			ClientCount: 15,
			CreatedAt:   now.Add(-365 * 24 * time.Hour),
			UpdatedAt:   now.Add(-10 * 24 * time.Hour),
		},
		{
			ID:          5,
			Name:        "Islamic",
			Description: "Swap-free Islamic account",
			FeeRules: map[string]FeeRule{
				"commission_per_lot":      {FeeType: "commission_per_lot", Value: 8.0, Currency: "USD", Description: "Commission per standard lot", IsEnabled: true},
				"spread_markup_pips":      {FeeType: "spread_markup_pips", Value: 0.7, Description: "Spread markup in pips", IsEnabled: true},
				"swap_rate_annual":        {FeeType: "swap_rate_annual", Value: 0.0, Description: "Swap-free account", IsEnabled: false},
				"deposit_fee_pct":         {FeeType: "deposit_fee_pct", Value: 0.0, Description: "Deposit fee percentage", IsEnabled: false},
				"withdrawal_fee_flat":     {FeeType: "withdrawal_fee_flat", Value: 30.0, Currency: "USD", Description: "Flat withdrawal fee", IsEnabled: true},
				"inactivity_fee_monthly":  {FeeType: "inactivity_fee_monthly", Value: 15.0, Currency: "USD", Description: "Monthly inactivity fee after 6 months", IsEnabled: true},
			},
			IsActive:    true,
			ClientCount: 10,
			CreatedAt:   now.Add(-365 * 24 * time.Hour),
			UpdatedAt:   now.Add(-25 * 24 * time.Hour),
		},
		{
			ID:          6,
			Name:        "Demo",
			Description: "Demo account - no fees",
			FeeRules: map[string]FeeRule{
				"commission_per_lot":      {FeeType: "commission_per_lot", Value: 0.0, Currency: "USD", Description: "No commission", IsEnabled: false},
				"spread_markup_pips":      {FeeType: "spread_markup_pips", Value: 0.0, Description: "Real spreads", IsEnabled: false},
				"swap_rate_annual":        {FeeType: "swap_rate_annual", Value: 0.0, Description: "No swap", IsEnabled: false},
				"deposit_fee_pct":         {FeeType: "deposit_fee_pct", Value: 0.0, Description: "No deposit fee", IsEnabled: false},
				"withdrawal_fee_flat":     {FeeType: "withdrawal_fee_flat", Value: 0.0, Currency: "USD", Description: "No withdrawal fee", IsEnabled: false},
				"inactivity_fee_monthly":  {FeeType: "inactivity_fee_monthly", Value: 0.0, Currency: "USD", Description: "No inactivity fee", IsEnabled: false},
			},
			IsActive:    true,
			ClientCount: 5,
			CreatedAt:   now.Add(-365 * 24 * time.Hour),
			UpdatedAt:   now.Add(-5 * 24 * time.Hour),
		},
	}

	for _, sched := range schedules {
		s.schedules[sched.ID] = sched
	}
}

func (s *FeeScheduleService) initClientMappings() {
	now := time.Now()

	// Distribution: Standard=80, Premium=60, VIP=30, Institutional=15, Islamic=10, Demo=5
	scheduleDistribution := []struct {
		scheduleID   int64
		scheduleName string
		count        int
	}{
		{1, "Standard", 80},
		{2, "Premium", 60},
		{3, "VIP", 30},
		{4, "Institutional", 15},
		{5, "Islamic", 10},
		{6, "Demo", 5},
	}

	clientID := int64(10001)
	for _, dist := range scheduleDistribution {
		for i := 0; i < dist.count; i++ {
			s.clientMappings[clientID] = &ClientFeeMapping{
				ClientID:     clientID,
				ClientName:   "Client-" + strconv.FormatInt(clientID, 10),
				ScheduleID:   dist.scheduleID,
				ScheduleName: dist.scheduleName,
				AssignedAt:   now.Add(time.Duration(-rand.Intn(180)) * 24 * time.Hour),
			}
			clientID++
		}
	}
}

func (s *FeeScheduleService) initRevenueData() {
	now := time.Now()
	s.monthlyRevenue = make([]FeeMonthlyRevenue, 0)

	feeTypes := []string{"commission_per_lot", "spread_markup_pips", "swap_rate_annual", "withdrawal_fee_flat", "inactivity_fee_monthly"}

	// Generate 12 months of revenue data
	for i := 11; i >= 0; i-- {
		month := now.AddDate(0, -i, 0)
		monthStr := month.Format("2006-01")

		byType := make(map[string]float64)
		totalRevenue := 0.0

		// Generate revenue for each fee type
		for _, feeType := range feeTypes {
			var revenue float64
			switch feeType {
			case "commission_per_lot":
				revenue = 50000 + rand.Float64()*30000 // $50k-$80k
			case "spread_markup_pips":
				revenue = 30000 + rand.Float64()*20000 // $30k-$50k
			case "swap_rate_annual":
				revenue = 15000 + rand.Float64()*10000 // $15k-$25k
			case "withdrawal_fee_flat":
				revenue = 5000 + rand.Float64()*3000 // $5k-$8k
			case "inactivity_fee_monthly":
				revenue = 2000 + rand.Float64()*1500 // $2k-$3.5k
			}
			byType[feeType] = revenue
			totalRevenue += revenue
		}

		s.monthlyRevenue = append(s.monthlyRevenue, FeeMonthlyRevenue{
			Month:        monthStr,
			TotalRevenue: totalRevenue,
			ByFeeType:    byType,
			ClientCount:  180 + rand.Intn(20), // 180-200 active clients
		})
	}
}

func (s *FeeScheduleService) initWaivers() {
	now := time.Now()
	feeTypes := []string{"commission_per_lot", "spread_markup_pips", "withdrawal_fee_flat", "inactivity_fee_monthly"}
	reasons := []string{
		"High-volume trader promotion",
		"New account welcome bonus",
		"Compensation for platform issue",
		"VIP program benefit",
		"Marketing campaign",
		"Retention incentive",
		"Special agreement",
	}
	admins := []string{"admin1", "admin2", "supervisor"}

	for i := 1; i <= 30; i++ {
		clientID := int64(10001 + rand.Intn(200))
		startDate := now.Add(time.Duration(-rand.Intn(60)) * 24 * time.Hour)
		durationDays := 30 + rand.Intn(150) // 30-180 days
		expiryDate := startDate.AddDate(0, 0, durationDays)

		// 80% active, 20% expired
		isActive := rand.Float64() < 0.8
		if !isActive {
			startDate = now.Add(time.Duration(-rand.Intn(90)-90) * 24 * time.Hour)
			expiryDate = now.Add(time.Duration(-rand.Intn(30)) * 24 * time.Hour)
		}

		clientName := "Client-" + strconv.FormatInt(clientID, 10)

		s.waivers[int64(i)] = &FeeWaiver{
			ID:         int64(i),
			ClientID:   clientID,
			ClientName: clientName,
			FeeType:    feeTypes[rand.Intn(len(feeTypes))],
			Reason:     reasons[rand.Intn(len(reasons))],
			StartDate:  startDate,
			ExpiryDate: expiryDate,
			WaivedBy:   admins[rand.Intn(len(admins))],
			IsActive:   isActive,
			CreatedAt:  startDate,
		}
	}
}

// GetAllSchedules returns all fee schedules
func (s *FeeScheduleService) GetAllSchedules() []*FeeSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedules := make([]*FeeSchedule, 0, len(s.schedules))
	for _, sched := range s.schedules {
		schedules = append(schedules, sched)
	}
	return schedules
}

// GetScheduleByID returns a specific fee schedule
func (s *FeeScheduleService) GetScheduleByID(id int64) *FeeSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.schedules[id]
}

// UpdateSchedule updates a fee schedule
func (s *FeeScheduleService) UpdateSchedule(id int64, updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	schedule, exists := s.schedules[id]
	if !exists {
		return nil
	}

	// Apply updates
	if description, ok := updates["description"].(string); ok {
		schedule.Description = description
	}
	if feeRulesData, ok := updates["fee_rules"].(map[string]interface{}); ok {
		// Update fee rules
		for feeType, ruleData := range feeRulesData {
			if ruleMap, ok := ruleData.(map[string]interface{}); ok {
				if rule, exists := schedule.FeeRules[feeType]; exists {
					if value, ok := ruleMap["value"].(float64); ok {
						rule.Value = value
					}
					if isEnabled, ok := ruleMap["is_enabled"].(bool); ok {
						rule.IsEnabled = isEnabled
					}
					schedule.FeeRules[feeType] = rule
				}
			}
		}
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		schedule.IsActive = isActive
	}

	schedule.UpdatedAt = time.Now()
	return nil
}

// CreateSchedule creates a new fee schedule
func (s *FeeScheduleService) CreateSchedule(req CreateFeeScheduleRequest) (*FeeSchedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	schedule := &FeeSchedule{
		ID:          s.nextScheduleID,
		Name:        req.Name,
		Description: req.Description,
		FeeRules:    req.FeeRules,
		IsActive:    true,
		ClientCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.schedules[schedule.ID] = schedule
	s.nextScheduleID++

	return schedule, nil
}

// GetClientFees returns fees applied to a specific client
func (s *FeeScheduleService) GetClientFees(clientID int64) *ClientFeeDetails {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mapping, exists := s.clientMappings[clientID]
	if !exists {
		return nil
	}

	schedule := s.schedules[mapping.ScheduleID]
	if schedule == nil {
		return nil
	}

	// Get active waivers for this client
	activeWaivers := make([]FeeWaiver, 0)
	now := time.Now()
	for _, waiver := range s.waivers {
		if waiver.ClientID == clientID && waiver.IsActive && waiver.ExpiryDate.After(now) {
			activeWaivers = append(activeWaivers, *waiver)
		}
	}

	// Estimate monthly fees (simplified calculation)
	estimatedMonthly := 0.0
	for _, rule := range schedule.FeeRules {
		if rule.IsEnabled {
			// Rough estimation based on fee type
			switch rule.FeeType {
			case "commission_per_lot":
				estimatedMonthly += rule.Value * 10 // assume 10 lots/month
			case "withdrawal_fee_flat":
				estimatedMonthly += rule.Value * 2 // assume 2 withdrawals/month
			case "inactivity_fee_monthly":
				// only if inactive (skip for estimate)
			}
		}
	}

	return &ClientFeeDetails{
		ClientID:         clientID,
		ClientName:       mapping.ClientName,
		ScheduleID:       mapping.ScheduleID,
		ScheduleName:     mapping.ScheduleName,
		AppliedFees:      schedule.FeeRules,
		ActiveWaivers:    activeWaivers,
		EstimatedMonthly: estimatedMonthly,
	}
}

// GetRevenue returns fee revenue analytics
func (s *FeeScheduleService) GetRevenue() FeeRevenue {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Calculate totals
	totalCollected := 0.0
	byFeeType := make(map[string]float64)

	for _, month := range s.monthlyRevenue {
		totalCollected += month.TotalRevenue
		for feeType, amount := range month.ByFeeType {
			byFeeType[feeType] += amount
		}
	}

	// Generate top clients (mock data)
	topClients := make([]ClientRevenueContrib, 0)
	for i := 0; i < 10; i++ {
		clientID := int64(10001 + i)
		mapping := s.clientMappings[clientID]
		if mapping == nil {
			continue
		}

		byType := make(map[string]float64)
		total := 0.0

		// Generate realistic amounts based on client's schedule
		schedule := s.schedules[mapping.ScheduleID]
		if schedule != nil {
			if rule, ok := schedule.FeeRules["commission_per_lot"]; ok && rule.IsEnabled {
				amount := rule.Value * float64(100+rand.Intn(900)) // 100-1000 lots
				byType["commission_per_lot"] = amount
				total += amount
			}
			if rule, ok := schedule.FeeRules["swap_rate_annual"]; ok && rule.IsEnabled {
				amount := 500 + rand.Float64()*2000 // $500-$2500
				byType["swap_rate_annual"] = amount
				total += amount
			}
		}

		topClients = append(topClients, ClientRevenueContrib{
			ClientID:   clientID,
			ClientName: mapping.ClientName,
			TotalPaid:  total,
			ByFeeType:  byType,
		})
	}

	periodStart := time.Now().AddDate(0, -11, 0)
	periodEnd := time.Now()

	return FeeRevenue{
		TotalCollected:   totalCollected,
		ByFeeType:        byFeeType,
		MonthlyBreakdown: s.monthlyRevenue,
		TopClients:       topClients,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
	}
}

// GetWaivers returns all fee waivers
func (s *FeeScheduleService) GetWaivers(activeOnly bool) []FeeWaiver {
	s.mu.RLock()
	defer s.mu.RUnlock()

	waivers := make([]FeeWaiver, 0)
	now := time.Now()

	for _, waiver := range s.waivers {
		if activeOnly {
			if waiver.IsActive && waiver.ExpiryDate.After(now) {
				waivers = append(waivers, *waiver)
			}
		} else {
			waivers = append(waivers, *waiver)
		}
	}

	return waivers
}

// CreateWaiver creates a new fee waiver
func (s *FeeScheduleService) CreateWaiver(req CreateFeeWaiverRequest, waivedBy string) (*FeeWaiver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiryDate := now.AddDate(0, 0, req.DurationDays)

	// Get client name
	clientName := "Unknown"
	if mapping, exists := s.clientMappings[req.ClientID]; exists {
		clientName = mapping.ClientName
	}

	waiver := &FeeWaiver{
		ID:         s.nextWaiverID,
		ClientID:   req.ClientID,
		ClientName: clientName,
		FeeType:    req.FeeType,
		Reason:     req.Reason,
		StartDate:  now,
		ExpiryDate: expiryDate,
		WaivedBy:   waivedBy,
		IsActive:   true,
		CreatedAt:  now,
	}

	s.waivers[waiver.ID] = waiver
	s.nextWaiverID++

	return waiver, nil
}

// ============================================
// HTTP Handlers
// ============================================

// FeeScheduleHandler handles fee schedule HTTP requests
type FeeScheduleHandler struct {
	service     *FeeScheduleService
	authService interface {
		ValidateAdminToken(r *http.Request) (int64, error)
	}
}

// NewFeeScheduleHandler creates a new fee schedule handler
func NewFeeScheduleHandler(service *FeeScheduleService, authService interface {
	ValidateAdminToken(r *http.Request) (int64, error)
}) *FeeScheduleHandler {
	return &FeeScheduleHandler{
		service:     service,
		authService: authService,
	}
}

// HandleGetSchedules handles GET /admin/fees/schedules
func (h *FeeScheduleHandler) HandleGetSchedules(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	schedules := h.service.GetAllSchedules()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(schedules)
}

// HandleGetScheduleByID handles GET /admin/fees/schedules/:id
func (h *FeeScheduleHandler) HandleGetScheduleByID(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/fees/schedules/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	schedule := h.service.GetScheduleByID(id)
	if schedule == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Schedule not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(schedule)
}

// HandleUpdateSchedule handles PUT /admin/fees/schedules/:id
func (h *FeeScheduleHandler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/fees/schedules/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid schedule ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateSchedule(id, updates); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Fee schedule updated successfully",
	})
}

// HandleCreateSchedule handles POST /admin/fees/schedules
func (h *FeeScheduleHandler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateFeeScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule, err := h.service.CreateSchedule(req)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// HandleGetClientFees handles GET /admin/fees/client/:clientId
func (h *FeeScheduleHandler) HandleGetClientFees(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract client ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/admin/fees/client/")
	clientID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	clientFees := h.service.GetClientFees(clientID)
	if clientFees == nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(clientFees)
}

// HandleGetRevenue handles GET /admin/fees/revenue
func (h *FeeScheduleHandler) HandleGetRevenue(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	revenue := h.service.GetRevenue()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(revenue)
}

// HandleGetWaivers handles GET /admin/fees/waivers
func (h *FeeScheduleHandler) HandleGetWaivers(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Query parameter: active_only (default: true)
	activeOnly := true
	if r.URL.Query().Get("active_only") == "false" {
		activeOnly = false
	}

	waivers := h.service.GetWaivers(activeOnly)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(waivers)
}

// HandleCreateWaiver handles POST /admin/fees/waivers
func (h *FeeScheduleHandler) HandleCreateWaiver(w http.ResponseWriter, r *http.Request) {
	// Validate admin authentication
	_, err := h.authService.ValidateAdminToken(r)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateFeeWaiverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, use a placeholder admin name (in real app, get from token)
	waivedBy := "admin"

	waiver, err := h.service.CreateWaiver(req, waivedBy)
	if err != nil {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(waiver)
}
