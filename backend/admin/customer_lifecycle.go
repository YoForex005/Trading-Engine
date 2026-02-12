package admin

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// LifecycleStage represents the current stage of a client
type LifecycleStage string

const (
	StageLead      LifecycleStage = "Lead"
	StageNewClient LifecycleStage = "NewClient"
	StageActive    LifecycleStage = "Active"
	StageDormant   LifecycleStage = "Dormant"
	StageAtRisk    LifecycleStage = "AtRisk"
	StageChurned   LifecycleStage = "Churned"
)

// ClientLifecycle represents a client's lifecycle data
type ClientLifecycle struct {
	ClientID          string         `json:"client_id"`
	Name              string         `json:"name"`
	Email             string         `json:"email"`
	Stage             LifecycleStage `json:"stage"`
	RegisteredAt      time.Time      `json:"registered_at"`
	VerifiedAt        *time.Time     `json:"verified_at,omitempty"`
	FirstDepositAt    *time.Time     `json:"first_deposit_at,omitempty"`
	FirstTradeAt      *time.Time     `json:"first_trade_at,omitempty"`
	LastTradeAt       *time.Time     `json:"last_trade_at,omitempty"`
	LastLoginAt       *time.Time     `json:"last_login_at,omitempty"`
	TotalDeposits     float64        `json:"total_deposits"`
	TotalWithdrawals  float64        `json:"total_withdrawals"`
	TotalTrades       int            `json:"total_trades"`
	LifetimeVolume    float64        `json:"lifetime_volume"`
	LifetimeRevenue   float64        `json:"lifetime_revenue"`
	ChurnRisk         float64        `json:"churn_risk"`
	DaysSinceLastTrade int           `json:"days_since_last_trade"`
	DaysSinceLastLogin int           `json:"days_since_last_login"`
	AccountAge        int            `json:"account_age_days"`
}

// RetentionCohort represents retention data for a registration cohort
type RetentionCohort struct {
	Month         string  `json:"month"`
	Registered    int     `json:"registered"`
	Retained30d   int     `json:"retained_30d"`
	Retained60d   int     `json:"retained_60d"`
	Retained90d   int     `json:"retained_90d"`
	Retained180d  int     `json:"retained_180d"`
	Retained365d  int     `json:"retained_365d"`
	Retention30d  float64 `json:"retention_30d_pct"`
	Retention60d  float64 `json:"retention_60d_pct"`
	Retention90d  float64 `json:"retention_90d_pct"`
	Retention180d float64 `json:"retention_180d_pct"`
	Retention365d float64 `json:"retention_365d_pct"`
}

// FunnelMetrics represents conversion funnel data
type FunnelMetrics struct {
	Registered      int     `json:"registered"`
	Verified        int     `json:"verified"`
	Deposited       int     `json:"deposited"`
	Traded          int     `json:"traded"`
	Active          int     `json:"active"`
	VerificationRate float64 `json:"verification_rate"`
	DepositRate      float64 `json:"deposit_rate"`
	TradingRate      float64 `json:"trading_rate"`
	ActivationRate   float64 `json:"activation_rate"`
}

// StageDistribution represents client count per lifecycle stage
type StageDistribution struct {
	Stage              LifecycleStage `json:"stage"`
	Count              int            `json:"count"`
	Percentage         float64        `json:"percentage"`
	AvgRevenue         float64        `json:"avg_revenue"`
	AvgTrades          float64        `json:"avg_trades"`
	AvgTimeInStage     int            `json:"avg_time_in_stage_days"`
}

// LifecycleSegment represents aggregated metrics for a lifecycle stage
type LifecycleSegment struct {
	Stage           LifecycleStage `json:"stage"`
	ClientCount     int            `json:"client_count"`
	AvgRevenue      float64        `json:"avg_revenue"`
	AvgTrades       float64        `json:"avg_trades"`
	AvgDeposits     float64        `json:"avg_deposits"`
	AvgWithdrawals  float64        `json:"avg_withdrawals"`
	AvgVolume       float64        `json:"avg_volume"`
	TotalRevenue    float64        `json:"total_revenue"`
}

// MonthlyTrend represents trend data for a month
type MonthlyTrend struct {
	Month           string  `json:"month"`
	NewRegistrations int    `json:"new_registrations"`
	NewActivations   int    `json:"new_activations"`
	Churned          int    `json:"churned"`
	ChurnRate        float64 `json:"churn_rate"`
	ActiveClients    int    `json:"active_clients"`
	Revenue          float64 `json:"revenue"`
}

// LifecycleTimeline represents key events in a client's lifecycle
type LifecycleTimeline struct {
	Events []TimelineEvent `json:"events"`
}

// TimelineEvent represents a single lifecycle event
type TimelineEvent struct {
	Date        time.Time `json:"date"`
	Event       string    `json:"event"`
	Description string    `json:"description"`
	Value       float64   `json:"value,omitempty"`
}

// LifecycleStore manages customer lifecycle data
type LifecycleStore struct {
	mu      sync.RWMutex
	clients []ClientLifecycle
	cohorts []RetentionCohort
}

// NewLifecycleStore creates a new lifecycle store with mock data
func NewLifecycleStore() *LifecycleStore {
	store := &LifecycleStore{
		clients: generateMockLifecycleClients(),
		cohorts: generateMockCohorts(),
	}
	return store
}

// generateMockLifecycleClients creates 200 mock clients across all lifecycle stages
func generateMockLifecycleClients() []ClientLifecycle {
	clients := make([]ClientLifecycle, 0, 200)
	rand.Seed(time.Now().UnixNano())

	names := []string{
		"John Smith", "Sarah Johnson", "Michael Brown", "Emily Davis", "David Wilson",
		"Lisa Anderson", "Robert Taylor", "Jennifer Thomas", "William Jackson", "Mary White",
		"James Harris", "Patricia Martin", "Christopher Thompson", "Linda Garcia", "Daniel Martinez",
		"Barbara Robinson", "Matthew Clark", "Nancy Rodriguez", "Anthony Lewis", "Karen Lee",
		"Mark Walker", "Betty Hall", "Donald Allen", "Helen Young", "Paul Hernandez",
		"Sandra King", "Andrew Wright", "Donna Lopez", "Joshua Hill", "Carol Scott",
		"Kevin Baker", "Michelle Adams", "Brian Perez", "Deborah Roberts", "George Turner",
		"Stephanie Phillips", "Edward Campbell", "Sharon Mitchell", "Ronald Carter", "Cynthia Evans",
	}

	stageWeights := map[LifecycleStage]int{
		StageLead:      20,
		StageNewClient: 30,
		StageActive:    80,
		StageDormant:   30,
		StageAtRisk:    25,
		StageChurned:   15,
	}

	clientID := 1000

	for stage, count := range stageWeights {
		for i := 0; i < count; i++ {
			client := ClientLifecycle{
				ClientID: fmt.Sprintf("CL%d", clientID),
				Name:     names[rand.Intn(len(names))],
				Email:    fmt.Sprintf("client%d@example.com", clientID),
				Stage:    stage,
			}
			clientID++

			daysAgo := rand.Intn(365)
			client.RegisteredAt = time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour)
			client.AccountAge = daysAgo

			switch stage {
			case StageLead:
				client.ChurnRisk = 0.1 + rand.Float64()*0.2
				client.DaysSinceLastLogin = rand.Intn(7)

			case StageNewClient:
				verifiedAt := client.RegisteredAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.VerifiedAt = &verifiedAt

				if rand.Float64() < 0.7 {
					depositAt := verifiedAt.Add(time.Duration(rand.Intn(14)) * 24 * time.Hour)
					client.FirstDepositAt = &depositAt
					client.TotalDeposits = 500 + rand.Float64()*4500
				}

				client.ChurnRisk = 0.3 + rand.Float64()*0.2
				client.DaysSinceLastLogin = rand.Intn(14)

			case StageActive:
				verifiedAt := client.RegisteredAt.Add(time.Duration(rand.Intn(3)) * 24 * time.Hour)
				client.VerifiedAt = &verifiedAt

				depositAt := verifiedAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.FirstDepositAt = &depositAt
				client.TotalDeposits = 2000 + rand.Float64()*48000

				tradeAt := depositAt.Add(time.Duration(rand.Intn(3)) * 24 * time.Hour)
				client.FirstTradeAt = &tradeAt

				lastTrade := time.Now().Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.LastTradeAt = &lastTrade

				lastLogin := time.Now().Add(-time.Duration(rand.Intn(3)) * 24 * time.Hour)
				client.LastLoginAt = &lastLogin

				client.TotalTrades = 10 + rand.Intn(500)
				client.LifetimeVolume = 50000 + rand.Float64()*950000
				client.LifetimeRevenue = 100 + rand.Float64()*9900
				client.TotalWithdrawals = client.TotalDeposits * (0.3 + rand.Float64()*0.4)
				client.ChurnRisk = 0.05 + rand.Float64()*0.15
				client.DaysSinceLastTrade = rand.Intn(7)
				client.DaysSinceLastLogin = rand.Intn(3)

			case StageDormant:
				verifiedAt := client.RegisteredAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.VerifiedAt = &verifiedAt

				depositAt := verifiedAt.Add(time.Duration(rand.Intn(14)) * 24 * time.Hour)
				client.FirstDepositAt = &depositAt
				client.TotalDeposits = 1000 + rand.Float64()*9000

				tradeAt := depositAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.FirstTradeAt = &tradeAt

				lastTrade := time.Now().Add(-time.Duration(31+rand.Intn(60)) * 24 * time.Hour)
				client.LastTradeAt = &lastTrade

				lastLogin := time.Now().Add(-time.Duration(20+rand.Intn(40)) * 24 * time.Hour)
				client.LastLoginAt = &lastLogin

				client.TotalTrades = 5 + rand.Intn(50)
				client.LifetimeVolume = 10000 + rand.Float64()*90000
				client.LifetimeRevenue = 50 + rand.Float64()*950
				client.TotalWithdrawals = client.TotalDeposits * (0.4 + rand.Float64()*0.3)
				client.ChurnRisk = 0.5 + rand.Float64()*0.2
				client.DaysSinceLastTrade = 31 + rand.Intn(60)
				client.DaysSinceLastLogin = 20 + rand.Intn(40)

			case StageAtRisk:
				verifiedAt := client.RegisteredAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.VerifiedAt = &verifiedAt

				depositAt := verifiedAt.Add(time.Duration(rand.Intn(14)) * 24 * time.Hour)
				client.FirstDepositAt = &depositAt
				client.TotalDeposits = 1500 + rand.Float64()*13500

				tradeAt := depositAt.Add(time.Duration(rand.Intn(7)) * 24 * time.Hour)
				client.FirstTradeAt = &tradeAt

				lastTrade := time.Now().Add(-time.Duration(15+rand.Intn(30)) * 24 * time.Hour)
				client.LastTradeAt = &lastTrade

				lastLogin := time.Now().Add(-time.Duration(10+rand.Intn(20)) * 24 * time.Hour)
				client.LastLoginAt = &lastLogin

				client.TotalTrades = 10 + rand.Intn(100)
				client.LifetimeVolume = 25000 + rand.Float64()*175000
				client.LifetimeRevenue = 75 + rand.Float64()*1425
				client.TotalWithdrawals = client.TotalDeposits * (0.5 + rand.Float64()*0.3)
				client.ChurnRisk = 0.7 + rand.Float64()*0.25
				client.DaysSinceLastTrade = 15 + rand.Intn(30)
				client.DaysSinceLastLogin = 10 + rand.Intn(20)

			case StageChurned:
				verifiedAt := client.RegisteredAt.Add(time.Duration(rand.Intn(14)) * 24 * time.Hour)
				client.VerifiedAt = &verifiedAt

				if rand.Float64() < 0.8 {
					depositAt := verifiedAt.Add(time.Duration(rand.Intn(21)) * 24 * time.Hour)
					client.FirstDepositAt = &depositAt
					client.TotalDeposits = 500 + rand.Float64()*4500

					if rand.Float64() < 0.6 {
						tradeAt := depositAt.Add(time.Duration(rand.Intn(14)) * 24 * time.Hour)
						client.FirstTradeAt = &tradeAt

						lastTrade := time.Now().Add(-time.Duration(91+rand.Intn(180)) * 24 * time.Hour)
						client.LastTradeAt = &lastTrade

						client.TotalTrades = 1 + rand.Intn(20)
						client.LifetimeVolume = 5000 + rand.Float64()*45000
						client.LifetimeRevenue = 10 + rand.Float64()*190
						client.DaysSinceLastTrade = 91 + rand.Intn(180)
					}

					client.TotalWithdrawals = client.TotalDeposits * (0.6 + rand.Float64()*0.3)
				}

				lastLogin := time.Now().Add(-time.Duration(91+rand.Intn(180)) * 24 * time.Hour)
				client.LastLoginAt = &lastLogin

				client.ChurnRisk = 0.95 + rand.Float64()*0.05
				client.DaysSinceLastLogin = 91 + rand.Intn(180)
			}

			clients = append(clients, client)
		}
	}

	return clients
}

// generateMockCohorts creates 12 months of retention cohort data
func generateMockCohorts() []RetentionCohort {
	cohorts := make([]RetentionCohort, 0, 12)
	now := time.Now()

	for i := 11; i >= 0; i-- {
		monthDate := now.AddDate(0, -i, 0)
		month := monthDate.Format("2006-01")

		registered := 150 + rand.Intn(100)
		retained30d := int(float64(registered) * (0.65 + rand.Float64()*0.15))
		retained60d := int(float64(retained30d) * (0.75 + rand.Float64()*0.15))
		retained90d := int(float64(retained60d) * (0.80 + rand.Float64()*0.12))
		retained180d := int(float64(retained90d) * (0.70 + rand.Float64()*0.15))
		retained365d := int(float64(retained180d) * (0.65 + rand.Float64()*0.15))

		if i < 1 {
			retained365d = 0
		}
		if i < 6 {
			retained180d = int(float64(retained90d) * (0.75 + rand.Float64()*0.10))
		}

		cohort := RetentionCohort{
			Month:        month,
			Registered:   registered,
			Retained30d:  retained30d,
			Retained60d:  retained60d,
			Retained90d:  retained90d,
			Retained180d: retained180d,
			Retained365d: retained365d,
		}

		if registered > 0 {
			cohort.Retention30d = (float64(retained30d) / float64(registered)) * 100
			cohort.Retention60d = (float64(retained60d) / float64(registered)) * 100
			cohort.Retention90d = (float64(retained90d) / float64(registered)) * 100
			cohort.Retention180d = (float64(retained180d) / float64(registered)) * 100
			cohort.Retention365d = (float64(retained365d) / float64(registered)) * 100
		}

		cohorts = append(cohorts, cohort)
	}

	return cohorts
}

// LifecycleHandler handles lifecycle requests
type LifecycleHandler struct {
	store       *LifecycleStore
	authService *auth.Service
}

// NewLifecycleHandler creates a new lifecycle handler
func NewLifecycleHandler(store *LifecycleStore, authService *auth.Service) *LifecycleHandler {
	return &LifecycleHandler{
		store:       store,
		authService: authService,
	}
}

// HandleGetOverview handles GET /admin/lifecycle/overview
func (h *LifecycleHandler) HandleGetOverview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	stageMap := make(map[LifecycleStage]*StageDistribution)
	for _, stage := range []LifecycleStage{StageLead, StageNewClient, StageActive, StageDormant, StageAtRisk, StageChurned} {
		stageMap[stage] = &StageDistribution{
			Stage: stage,
		}
	}

	totalClients := len(h.store.clients)

	for _, client := range h.store.clients {
		dist := stageMap[client.Stage]
		dist.Count++
		dist.AvgRevenue += client.LifetimeRevenue
		dist.AvgTrades += float64(client.TotalTrades)
		dist.AvgTimeInStage += client.AccountAge
	}

	distributions := make([]StageDistribution, 0, len(stageMap))
	for _, dist := range stageMap {
		if dist.Count > 0 {
			dist.Percentage = (float64(dist.Count) / float64(totalClients)) * 100
			dist.AvgRevenue = dist.AvgRevenue / float64(dist.Count)
			dist.AvgTrades = dist.AvgTrades / float64(dist.Count)
			dist.AvgTimeInStage = dist.AvgTimeInStage / dist.Count
		}
		distributions = append(distributions, *dist)
	}

	sort.Slice(distributions, func(i, j int) bool {
		order := map[LifecycleStage]int{
			StageLead: 0, StageNewClient: 1, StageActive: 2,
			StageDormant: 3, StageAtRisk: 4, StageChurned: 5,
		}
		return order[distributions[i].Stage] < order[distributions[j].Stage]
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"total_clients": totalClients,
		"distributions": distributions,
	})
}

// HandleGetFunnel handles GET /admin/lifecycle/funnel
func (h *LifecycleHandler) HandleGetFunnel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	funnel := FunnelMetrics{
		Registered: len(h.store.clients),
	}

	for _, client := range h.store.clients {
		if client.VerifiedAt != nil {
			funnel.Verified++
		}
		if client.FirstDepositAt != nil {
			funnel.Deposited++
		}
		if client.FirstTradeAt != nil {
			funnel.Traded++
		}
		if client.Stage == StageActive {
			funnel.Active++
		}
	}

	if funnel.Registered > 0 {
		funnel.VerificationRate = (float64(funnel.Verified) / float64(funnel.Registered)) * 100
		funnel.DepositRate = (float64(funnel.Deposited) / float64(funnel.Registered)) * 100
		funnel.TradingRate = (float64(funnel.Traded) / float64(funnel.Registered)) * 100
		funnel.ActivationRate = (float64(funnel.Active) / float64(funnel.Registered)) * 100
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"funnel":  funnel,
	})
}

// HandleGetRetention handles GET /admin/lifecycle/retention
func (h *LifecycleHandler) HandleGetRetention(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	cohorts := h.store.cohorts
	h.store.mu.RUnlock()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cohorts": cohorts,
		"count":   len(cohorts),
	})
}

// HandleGetAtRisk handles GET /admin/lifecycle/at-risk
func (h *LifecycleHandler) HandleGetAtRisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	atRiskClients := make([]ClientLifecycle, 0)
	for _, client := range h.store.clients {
		if client.ChurnRisk > 0.7 {
			atRiskClients = append(atRiskClients, client)
		}
	}

	sort.Slice(atRiskClients, func(i, j int) bool {
		return atRiskClients[i].ChurnRisk > atRiskClients[j].ChurnRisk
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"clients": atRiskClients,
		"count":   len(atRiskClients),
	})
}

// HandleGetSegments handles GET /admin/lifecycle/segments
func (h *LifecycleHandler) HandleGetSegments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	segmentMap := make(map[LifecycleStage]*LifecycleSegment)
	for _, stage := range []LifecycleStage{StageLead, StageNewClient, StageActive, StageDormant, StageAtRisk, StageChurned} {
		segmentMap[stage] = &LifecycleSegment{
			Stage: stage,
		}
	}

	for _, client := range h.store.clients {
		seg := segmentMap[client.Stage]
		seg.ClientCount++
		seg.TotalRevenue += client.LifetimeRevenue
		seg.AvgRevenue += client.LifetimeRevenue
		seg.AvgTrades += float64(client.TotalTrades)
		seg.AvgDeposits += client.TotalDeposits
		seg.AvgWithdrawals += client.TotalWithdrawals
		seg.AvgVolume += client.LifetimeVolume
	}

	segments := make([]LifecycleSegment, 0, len(segmentMap))
	for _, seg := range segmentMap {
		if seg.ClientCount > 0 {
			seg.AvgRevenue = seg.AvgRevenue / float64(seg.ClientCount)
			seg.AvgTrades = seg.AvgTrades / float64(seg.ClientCount)
			seg.AvgDeposits = seg.AvgDeposits / float64(seg.ClientCount)
			seg.AvgWithdrawals = seg.AvgWithdrawals / float64(seg.ClientCount)
			seg.AvgVolume = seg.AvgVolume / float64(seg.ClientCount)
		}
		segments = append(segments, *seg)
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i].TotalRevenue > segments[j].TotalRevenue
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"segments": segments,
		"count":    len(segments),
	})
}

// HandleGetClientTimeline handles GET /admin/lifecycle/client/:id
func (h *LifecycleHandler) HandleGetClientTimeline(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Client ID required", http.StatusBadRequest)
		return
	}
	clientID := pathParts[4]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var client *ClientLifecycle
	for i := range h.store.clients {
		if h.store.clients[i].ClientID == clientID {
			client = &h.store.clients[i]
			break
		}
	}

	if client == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	timeline := LifecycleTimeline{
		Events: make([]TimelineEvent, 0),
	}

	timeline.Events = append(timeline.Events, TimelineEvent{
		Date:        client.RegisteredAt,
		Event:       "Registration",
		Description: "Client registered account",
	})

	if client.VerifiedAt != nil {
		timeline.Events = append(timeline.Events, TimelineEvent{
			Date:        *client.VerifiedAt,
			Event:       "Verification",
			Description: "Account verified",
		})
	}

	if client.FirstDepositAt != nil {
		timeline.Events = append(timeline.Events, TimelineEvent{
			Date:        *client.FirstDepositAt,
			Event:       "First Deposit",
			Description: "First deposit made",
			Value:       client.TotalDeposits,
		})
	}

	if client.FirstTradeAt != nil {
		timeline.Events = append(timeline.Events, TimelineEvent{
			Date:        *client.FirstTradeAt,
			Event:       "First Trade",
			Description: "First trade executed",
		})
	}

	if client.LastTradeAt != nil && client.DaysSinceLastTrade > 30 {
		timeline.Events = append(timeline.Events, TimelineEvent{
			Date:        *client.LastTradeAt,
			Event:       "Inactivity Start",
			Description: fmt.Sprintf("No trades for %d days", client.DaysSinceLastTrade),
		})
	}

	if client.ChurnRisk > 0.7 {
		timeline.Events = append(timeline.Events, TimelineEvent{
			Date:        time.Now(),
			Event:       "High Churn Risk",
			Description: fmt.Sprintf("Churn risk: %.1f%%", client.ChurnRisk*100),
			Value:       client.ChurnRisk,
		})
	}

	sort.Slice(timeline.Events, func(i, j int) bool {
		return timeline.Events[i].Date.Before(timeline.Events[j].Date)
	})

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"client":   client,
		"timeline": timeline,
	})
}

// HandleGetTrends handles GET /admin/lifecycle/trends
func (h *LifecycleHandler) HandleGetTrends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	trends := make([]MonthlyTrend, 0, 12)
	now := time.Now()

	for i := 11; i >= 0; i-- {
		monthDate := now.AddDate(0, -i, 0)
		month := monthDate.Format("2006-01")

		newRegs := 0
		newActivations := 0
		churned := 0
		activeClients := 0
		revenue := 0.0

		monthStart := time.Date(monthDate.Year(), monthDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		monthEnd := monthStart.AddDate(0, 1, 0)

		for _, client := range h.store.clients {
			if client.RegisteredAt.After(monthStart) && client.RegisteredAt.Before(monthEnd) {
				newRegs++
			}

			if client.FirstTradeAt != nil && client.FirstTradeAt.After(monthStart) && client.FirstTradeAt.Before(monthEnd) {
				newActivations++
			}

			if client.Stage == StageChurned {
				if client.LastLoginAt != nil {
					daysSinceChurn := int(time.Since(*client.LastLoginAt).Hours() / 24)
					if daysSinceChurn >= 90 && daysSinceChurn < 120 {
						churned++
					}
				}
			}

			if client.Stage == StageActive {
				activeClients++
			}

			if client.RegisteredAt.Before(monthEnd) {
				revenue += client.LifetimeRevenue / float64(math.Max(1, float64(client.AccountAge/30)))
			}
		}

		churnRate := 0.0
		if activeClients > 0 {
			churnRate = (float64(churned) / float64(activeClients+churned)) * 100
		}

		trends = append(trends, MonthlyTrend{
			Month:            month,
			NewRegistrations: newRegs,
			NewActivations:   newActivations,
			Churned:          churned,
			ChurnRate:        churnRate,
			ActiveClients:    activeClients,
			Revenue:          revenue,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"trends":  trends,
		"count":   len(trends),
	})
}
