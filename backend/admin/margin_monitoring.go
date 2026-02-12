package admin

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

type MarginLevel struct {
	ClientID      int64     `json:"clientId"`
	ClientName    string    `json:"clientName"`
	Equity        float64   `json:"equity"`        // USD
	UsedMargin    float64   `json:"usedMargin"`    // USD
	FreeMargin    float64   `json:"freeMargin"`    // USD
	MarginLevel   float64   `json:"marginLevel"`   // % (equity/used_margin * 100)
	Status        string    `json:"status"`        // "safe", "warning", "margin_call", "stop_out"
	OpenPositions int       `json:"openPositions"`
	LastUpdate    time.Time `json:"lastUpdate"`
}

type PositionMarginDetail struct {
	PositionID     int64   `json:"positionId"`
	Symbol         string  `json:"symbol"`
	Direction      string  `json:"direction"` // "long" or "short"
	Volume         float64 `json:"volume"`    // lots
	EntryPrice     float64 `json:"entryPrice"`
	CurrentPrice   float64 `json:"currentPrice"`
	MarginRequired float64 `json:"marginRequired"` // USD
	UnrealizedPnL  float64 `json:"unrealizedPnL"`  // USD
	Leverage       int     `json:"leverage"`       // e.g., 100, 500
}

type ClientMarginBreakdown struct {
	ClientID       int64                  `json:"clientId"`
	ClientName     string                 `json:"clientName"`
	Equity         float64                `json:"equity"`
	UsedMargin     float64                `json:"usedMargin"`
	FreeMargin     float64                `json:"freeMargin"`
	MarginLevel    float64                `json:"marginLevel"`
	Status         string                 `json:"status"`
	Positions      []PositionMarginDetail `json:"positions"`
	TotalPositions int                    `json:"totalPositions"`
}

type MarginCallRecord struct {
	ID              int64      `json:"id"`
	ClientID        int64      `json:"clientId"`
	ClientName      string     `json:"clientName"`
	TriggeredAt     time.Time  `json:"triggeredAt"`
	MarginLevelAtTrigger float64 `json:"marginLevelAtTrigger"` // %
	EquityAtTrigger float64    `json:"equityAtTrigger"`       // USD
	Status          string     `json:"status"`                // "active" or "resolved"
	ResolutionMethod string    `json:"resolutionMethod,omitempty"` // "deposit_received", "positions_closed", "manual_override"
	ResolvedAt      *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy      string     `json:"resolvedBy,omitempty"`
	Notes           string     `json:"notes,omitempty"`
}

type LiquidationEvent struct {
	ID           int64     `json:"id"`
	ClientID     int64     `json:"clientId"`
	ClientName   string    `json:"clientName"`
	PositionID   int64     `json:"positionId"`
	Symbol       string    `json:"symbol"`
	Direction    string    `json:"direction"`
	Volume       float64   `json:"volume"`
	EntryPrice   float64   `json:"entryPrice"`
	ClosePrice   float64   `json:"closePrice"`
	Loss         float64   `json:"loss"` // USD (negative)
	LiquidatedAt time.Time `json:"liquidatedAt"`
	Reason       string    `json:"reason"` // "stop_out", "margin_level_below_threshold"
}

type MarginThreshold struct {
	GroupID          int64     `json:"groupId"`
	GroupName        string    `json:"groupName"`
	WarningLevel     float64   `json:"warningLevel"`     // % (e.g., 120%)
	MarginCallLevel  float64   `json:"marginCallLevel"`  // % (e.g., 100%)
	StopOutLevel     float64   `json:"stopOutLevel"`     // % (e.g., 50%)
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ============================================
// Service
// ============================================

type MarginMonitoringService struct {
	marginLevels  map[int64]*MarginLevel
	marginCalls   map[int64]*MarginCallRecord
	liquidations  map[int64]*LiquidationEvent
	thresholds    map[int64]*MarginThreshold
	positions     map[int64][]PositionMarginDetail // clientID -> positions
	nextCallID    int64
	nextLiquidationID int64
	mu            sync.RWMutex
}

func NewMarginMonitoringService() *MarginMonitoringService {
	service := &MarginMonitoringService{
		marginLevels:  make(map[int64]*MarginLevel),
		marginCalls:   make(map[int64]*MarginCallRecord),
		liquidations:  make(map[int64]*LiquidationEvent),
		thresholds:    make(map[int64]*MarginThreshold),
		positions:     make(map[int64][]PositionMarginDetail),
		nextCallID:    1,
		nextLiquidationID: 1,
	}
	service.initializeMockData()
	return service
}

func (s *MarginMonitoringService) initializeMockData() {
	now := time.Now()

	// ============================================
	// Initialize 5 Trading Groups with Margin Thresholds
	// ============================================
	thresholds := []MarginThreshold{
		{
			GroupID:         1,
			GroupName:       "Retail Traders",
			WarningLevel:    120.0,
			MarginCallLevel: 100.0,
			StopOutLevel:    50.0,
			UpdatedAt:       now.AddDate(0, -3, 0),
		},
		{
			GroupID:         2,
			GroupName:       "Professional Traders",
			WarningLevel:    150.0,
			MarginCallLevel: 80.0,
			StopOutLevel:    30.0,
			UpdatedAt:       now.AddDate(0, -2, 0),
		},
		{
			GroupID:         3,
			GroupName:       "Institutional",
			WarningLevel:    200.0,
			MarginCallLevel: 70.0,
			StopOutLevel:    20.0,
			UpdatedAt:       now.AddDate(0, -1, 0),
		},
		{
			GroupID:         4,
			GroupName:       "VIP Clients",
			WarningLevel:    150.0,
			MarginCallLevel: 75.0,
			StopOutLevel:    25.0,
			UpdatedAt:       now.AddDate(0, 0, -15),
		},
		{
			GroupID:         5,
			GroupName:       "Demo Accounts",
			WarningLevel:    120.0,
			MarginCallLevel: 100.0,
			StopOutLevel:    50.0,
			UpdatedAt:       now.AddDate(0, -6, 0),
		},
	}

	for i := range thresholds {
		s.thresholds[thresholds[i].GroupID] = &thresholds[i]
	}

	// ============================================
	// Initialize 200 Clients with Margin Data
	// ============================================
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "XAUUSD", "BTCUSD", "ETHUSD", "US30", "NAS100", "SPX500"}
	directions := []string{"long", "short"}

	for clientID := int64(1); clientID <= 200; clientID++ {
		// Random equity and margin
		equity := 1000 + rand.Float64()*49000 // $1,000 to $50,000
		usedMargin := equity * (0.1 + rand.Float64()*0.7) // 10% to 80% of equity
		freeMargin := equity - usedMargin
		marginLevel := (equity / usedMargin) * 100

		// Determine status based on margin level
		var status string
		if marginLevel >= 200 {
			status = "safe"
		} else if marginLevel >= 120 {
			status = "warning"
		} else if marginLevel >= 50 {
			status = "margin_call"
		} else {
			status = "stop_out"
		}

		// Random number of open positions (1-5)
		numPositions := rand.Intn(5) + 1

		ml := &MarginLevel{
			ClientID:      clientID,
			ClientName:    "Client" + strconv.FormatInt(clientID, 10),
			Equity:        equity,
			UsedMargin:    usedMargin,
			FreeMargin:    freeMargin,
			MarginLevel:   marginLevel,
			Status:        status,
			OpenPositions: numPositions,
			LastUpdate:    now.Add(-time.Duration(rand.Intn(3600)) * time.Second),
		}

		s.marginLevels[clientID] = ml

		// Create positions for this client
		positions := make([]PositionMarginDetail, numPositions)
		totalMarginRequired := 0.0

		for i := 0; i < numPositions; i++ {
			symbol := symbols[rand.Intn(len(symbols))]
			direction := directions[rand.Intn(2)]
			volume := 0.01 + rand.Float64()*2.0 // 0.01 to 2.01 lots
			leverage := []int{50, 100, 200, 500}[rand.Intn(4)]

			entryPrice := 1.0 + rand.Float64()*100
			currentPrice := entryPrice * (1 + (rand.Float64()-0.5)*0.1) // ±5%

			// Calculate margin required
			marginRequired := (volume * 100000 * entryPrice) / float64(leverage)
			totalMarginRequired += marginRequired

			// Calculate P&L
			var pnl float64
			if direction == "long" {
				pnl = (currentPrice - entryPrice) * volume * 100000
			} else {
				pnl = (entryPrice - currentPrice) * volume * 100000
			}

			positions[i] = PositionMarginDetail{
				PositionID:     int64(i + 1),
				Symbol:         symbol,
				Direction:      direction,
				Volume:         volume,
				EntryPrice:     entryPrice,
				CurrentPrice:   currentPrice,
				MarginRequired: marginRequired,
				UnrealizedPnL:  pnl,
				Leverage:       leverage,
			}
		}

		// Adjust used margin to match total margin required
		if totalMarginRequired > 0 {
			ml.UsedMargin = totalMarginRequired
			ml.FreeMargin = ml.Equity - totalMarginRequired
			if totalMarginRequired > 0 {
				ml.MarginLevel = (ml.Equity / totalMarginRequired) * 100
			}
		}

		s.positions[clientID] = positions
	}

	// ============================================
	// Initialize 50 Margin Calls (30 resolved, 20 active)
	// ============================================
	for i := 0; i < 50; i++ {
		daysAgo := rand.Intn(60)
		triggeredAt := now.AddDate(0, 0, -daysAgo)
		clientID := int64(rand.Intn(200) + 1)

		status := "active"
		if i < 30 {
			status = "resolved"
		}

		call := &MarginCallRecord{
			ID:                   int64(i + 1),
			ClientID:             clientID,
			ClientName:           "Client" + strconv.FormatInt(clientID, 10),
			TriggeredAt:          triggeredAt,
			MarginLevelAtTrigger: 50 + rand.Float64()*50, // 50% to 100%
			EquityAtTrigger:      1000 + rand.Float64()*19000,
			Status:               status,
		}

		if status == "resolved" {
			methods := []string{"deposit_received", "positions_closed", "manual_override"}
			call.ResolutionMethod = methods[rand.Intn(3)]
			resolvedAt := triggeredAt.Add(time.Duration(rand.Intn(48)) * time.Hour)
			call.ResolvedAt = &resolvedAt
			call.ResolvedBy = "admin" + strconv.Itoa(rand.Intn(5)+1) + "@rtx5.com"

			switch call.ResolutionMethod {
			case "deposit_received":
				call.Notes = "Client deposited additional funds to restore margin level"
			case "positions_closed":
				call.Notes = "Client closed losing positions to reduce margin usage"
			case "manual_override":
				call.Notes = "Admin approved temporary margin override for valued client"
			}
		}

		s.marginCalls[call.ID] = call
	}

	s.nextCallID = 51

	// ============================================
	// Initialize 100 Liquidation Events
	// ============================================
	for i := 0; i < 100; i++ {
		daysAgo := rand.Intn(90)
		liquidatedAt := now.AddDate(0, 0, -daysAgo)
		clientID := int64(rand.Intn(200) + 1)
		symbol := symbols[rand.Intn(len(symbols))]
		direction := directions[rand.Intn(2)]

		volume := 0.01 + rand.Float64()*5.0
		entryPrice := 1.0 + rand.Float64()*100
		closePrice := entryPrice * (1 - 0.05 - rand.Float64()*0.15) // 5% to 20% loss

		var loss float64
		if direction == "long" {
			loss = (closePrice - entryPrice) * volume * 100000
		} else {
			loss = (entryPrice - closePrice) * volume * 100000
		}

		reasons := []string{"stop_out", "margin_level_below_threshold"}

		liquidation := &LiquidationEvent{
			ID:           int64(i + 1),
			ClientID:     clientID,
			ClientName:   "Client" + strconv.FormatInt(clientID, 10),
			PositionID:   int64(rand.Intn(1000) + 1),
			Symbol:       symbol,
			Direction:    direction,
			Volume:       volume,
			EntryPrice:   entryPrice,
			ClosePrice:   closePrice,
			Loss:         loss,
			LiquidatedAt: liquidatedAt,
			Reason:       reasons[rand.Intn(2)],
		}

		s.liquidations[liquidation.ID] = liquidation
	}

	s.nextLiquidationID = 101

	log.Printf("[MarginMonitoringService] Initialized with %d clients, %d margin calls, %d liquidations, %d threshold groups",
		len(s.marginLevels), len(s.marginCalls), len(s.liquidations), len(s.thresholds))
}

// ============================================
// Service Methods
// ============================================

func (s *MarginMonitoringService) GetAllMarginLevels() []MarginLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	levels := make([]MarginLevel, 0, len(s.marginLevels))
	for _, level := range s.marginLevels {
		levels = append(levels, *level)
	}

	// Sort by margin level ascending (most at risk first)
	sort.Slice(levels, func(i, j int) bool {
		return levels[i].MarginLevel < levels[j].MarginLevel
	})

	return levels
}

func (s *MarginMonitoringService) GetClientMarginBreakdown(clientID int64) *ClientMarginBreakdown {
	s.mu.RLock()
	defer s.mu.RUnlock()

	level, exists := s.marginLevels[clientID]
	if !exists {
		return nil
	}

	positions := s.positions[clientID]

	breakdown := &ClientMarginBreakdown{
		ClientID:       level.ClientID,
		ClientName:     level.ClientName,
		Equity:         level.Equity,
		UsedMargin:     level.UsedMargin,
		FreeMargin:     level.FreeMargin,
		MarginLevel:    level.MarginLevel,
		Status:         level.Status,
		Positions:      positions,
		TotalPositions: len(positions),
	}

	return breakdown
}

func (s *MarginMonitoringService) GetClientsAtRisk() []MarginLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	atRisk := make([]MarginLevel, 0)

	for _, level := range s.marginLevels {
		if level.MarginLevel < 150.0 {
			atRisk = append(atRisk, *level)
		}
	}

	// Sort by margin level ascending (most at risk first)
	sort.Slice(atRisk, func(i, j int) bool {
		return atRisk[i].MarginLevel < atRisk[j].MarginLevel
	})

	return atRisk
}

func (s *MarginMonitoringService) GetMarginCalls(filterStatus string) []MarginCallRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	calls := make([]MarginCallRecord, 0)

	for _, call := range s.marginCalls {
		if filterStatus != "" && call.Status != filterStatus {
			continue
		}
		calls = append(calls, *call)
	}

	// Sort by triggered time descending (most recent first)
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].TriggeredAt.After(calls[j].TriggeredAt)
	})

	return calls
}

func (s *MarginMonitoringService) ResolveMarginCall(callID int64, method, resolvedBy, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	call, exists := s.marginCalls[callID]
	if !exists {
		return nil // Call not found
	}

	resolvedAt := time.Now()
	call.Status = "resolved"
	call.ResolutionMethod = method
	call.ResolvedAt = &resolvedAt
	call.ResolvedBy = resolvedBy
	call.Notes = notes

	return nil
}

func (s *MarginMonitoringService) GetLiquidations() []LiquidationEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	liquidations := make([]LiquidationEvent, 0, len(s.liquidations))
	for _, liq := range s.liquidations {
		liquidations = append(liquidations, *liq)
	}

	// Sort by liquidated time descending (most recent first)
	sort.Slice(liquidations, func(i, j int) bool {
		return liquidations[i].LiquidatedAt.After(liquidations[j].LiquidatedAt)
	})

	return liquidations
}

func (s *MarginMonitoringService) GetThresholds() []MarginThreshold {
	s.mu.RLock()
	defer s.mu.RUnlock()

	thresholds := make([]MarginThreshold, 0, len(s.thresholds))
	for _, threshold := range s.thresholds {
		thresholds = append(thresholds, *threshold)
	}

	sort.Slice(thresholds, func(i, j int) bool {
		return thresholds[i].GroupID < thresholds[j].GroupID
	})

	return thresholds
}

func (s *MarginMonitoringService) UpdateThreshold(groupID int64, warning, marginCall, stopOut float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	threshold, exists := s.thresholds[groupID]
	if !exists {
		return nil // Group not found
	}

	threshold.WarningLevel = warning
	threshold.MarginCallLevel = marginCall
	threshold.StopOutLevel = stopOut
	threshold.UpdatedAt = time.Now()

	return nil
}

// ============================================
// HTTP Handlers
// ============================================

type MarginMonitoringHandler struct {
	service     *MarginMonitoringService
	authService *auth.Service
}

func NewMarginMonitoringHandler(service *MarginMonitoringService, authService *auth.Service) *MarginMonitoringHandler {
	return &MarginMonitoringHandler{
		service:     service,
		authService: authService,
	}
}

// 1. GET /admin/margin/levels - All clients with margin levels
func (h *MarginMonitoringHandler) HandleGetAllLevels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	levels := h.service.GetAllMarginLevels()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": levels,
		"count":   len(levels),
	})
}

// 2. GET /admin/margin/levels/:clientId - Client margin breakdown
func (h *MarginMonitoringHandler) HandleGetClientBreakdown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	clientIDStr := strings.TrimPrefix(r.URL.Path, "/admin/margin/levels/")
	clientID, err := strconv.ParseInt(clientIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	breakdown := h.service.GetClientMarginBreakdown(clientID)
	if breakdown == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(breakdown)
}

// 3. GET /admin/margin/at-risk - Clients approaching margin call
func (h *MarginMonitoringHandler) HandleGetAtRisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	clients := h.service.GetClientsAtRisk()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	})
}

// 4. GET /admin/margin/calls - Margin call records
func (h *MarginMonitoringHandler) HandleGetMarginCalls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filterStatus := r.URL.Query().Get("status")
	calls := h.service.GetMarginCalls(filterStatus)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"calls": calls,
		"count": len(calls),
	})
}

// 5. POST /admin/margin/calls/:id/resolve - Resolve margin call
func (h *MarginMonitoringHandler) HandleResolveMarginCall(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract call ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid call ID", http.StatusBadRequest)
		return
	}

	callIDStr := pathParts[4]
	callID, err := strconv.ParseInt(callIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid call ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Method     string `json:"method"`
		ResolvedBy string `json:"resolvedBy"`
		Notes      string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ResolveMarginCall(callID, req.Method, req.ResolvedBy, req.Notes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Margin call resolved successfully",
		"callId":  callID,
	})
}

// 6. GET /admin/margin/liquidations - Liquidation history
func (h *MarginMonitoringHandler) HandleGetLiquidations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	liquidations := h.service.GetLiquidations()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"liquidations": liquidations,
		"count":        len(liquidations),
	})
}

// 7. GET /admin/margin/thresholds - Get margin thresholds per group
func (h *MarginMonitoringHandler) HandleGetThresholds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	thresholds := h.service.GetThresholds()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": thresholds,
		"count":  len(thresholds),
	})
}

// 8. PUT /admin/margin/thresholds/:groupId - Update thresholds
func (h *MarginMonitoringHandler) HandleUpdateThreshold(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupIDStr := strings.TrimPrefix(r.URL.Path, "/admin/margin/thresholds/")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req struct {
		WarningLevel    float64 `json:"warningLevel"`
		MarginCallLevel float64 `json:"marginCallLevel"`
		StopOutLevel    float64 `json:"stopOutLevel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateThreshold(groupID, req.WarningLevel, req.MarginCallLevel, req.StopOutLevel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Margin thresholds updated successfully",
		"groupId": groupID,
	})
}
