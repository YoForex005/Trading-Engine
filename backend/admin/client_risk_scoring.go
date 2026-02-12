package admin

import (
	"encoding/json"
	"fmt"
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
// Client Risk Scoring / Credit Assessment
// ============================================

type RiskFactor struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Weight      float64 `json:"weight"` // percentage (must sum to 100%)
	Description string  `json:"description"`
}

type ClientRiskScore struct {
	ClientID       int64              `json:"clientId"`
	ClientName     string             `json:"clientName"`
	OverallScore   float64            `json:"overallScore"` // 0-100
	Category       string             `json:"category"`     // Low, Medium, High, Critical
	FactorScores   map[string]float64 `json:"factorScores"` // individual factor scores
	ScoreHistory   []ScoreHistoryPoint `json:"scoreHistory,omitempty"` // last 30 days
	LastCalculated time.Time          `json:"lastCalculated"`
	LastChanged    time.Time          `json:"lastChanged"`
	PreviousScore  float64            `json:"previousScore"`
}

type ScoreHistoryPoint struct {
	Date  string  `json:"date"`
	Score float64 `json:"score"`
}

type RiskDistribution struct {
	Bucket     string `json:"bucket"` // "0-10", "10-20", etc.
	Count      int    `json:"count"`
	Percentage float64 `json:"percentage"`
}

type ClientRiskAlert struct {
	ClientID     int64   `json:"clientId"`
	ClientName   string  `json:"clientName"`
	OldScore     float64 `json:"oldScore"`
	NewScore     float64 `json:"newScore"`
	Change       float64 `json:"change"`
	OldCategory  string  `json:"oldCategory"`
	NewCategory  string  `json:"newCategory"`
	DetectedAt   time.Time `json:"detectedAt"`
}

type RiskStats struct {
	AverageScore   float64 `json:"averageScore"`
	MedianScore    float64 `json:"medianScore"`
	LowRiskCount   int     `json:"lowRiskCount"`
	MediumRiskCount int    `json:"mediumRiskCount"`
	HighRiskCount  int     `json:"highRiskCount"`
	CriticalCount  int     `json:"criticalCount"`
	TotalClients   int     `json:"totalClients"`
	ScoreTrend     string  `json:"scoreTrend"` // improving, stable, deteriorating
	LastUpdated    time.Time `json:"lastUpdated"`
}

// ============================================
// Service
// ============================================

type RiskScoringService struct {
	clients   map[int64]*ClientRiskScore
	factors   map[int64]*RiskFactor
	alerts    []*ClientRiskAlert
	nextFactorID int64
	mu        sync.RWMutex
}

func NewRiskScoringService() *RiskScoringService {
	s := &RiskScoringService{
		clients:   make(map[int64]*ClientRiskScore),
		factors:   make(map[int64]*RiskFactor),
		alerts:    []*ClientRiskAlert{},
		nextFactorID: 1,
	}
	s.initializeFactors()
	s.generateMockData()
	return s
}

func (s *RiskScoringService) initializeFactors() {
	factors := []RiskFactor{
		{ID: 1, Name: "Trading Volume", Weight: 20.0, Description: "Total trading volume and frequency"},
		{ID: 2, Name: "Drawdown History", Weight: 15.0, Description: "Historical maximum drawdown percentage"},
		{ID: 3, Name: "Leverage Usage", Weight: 15.0, Description: "Average leverage used in trading"},
		{ID: 4, Name: "Deposit/Withdrawal Ratio", Weight: 10.0, Description: "Ratio of deposits to withdrawals"},
		{ID: 5, Name: "Account Age", Weight: 10.0, Description: "Length of time account has been active"},
		{ID: 6, Name: "Margin Utilization", Weight: 10.0, Description: "Percentage of margin typically used"},
		{ID: 7, Name: "Loss Streak", Weight: 10.0, Description: "Current consecutive losing trades"},
		{ID: 8, Name: "Geographic Risk", Weight: 10.0, Description: "Risk based on client location"},
	}

	for _, f := range factors {
		factor := f
		s.factors[factor.ID] = &factor
		s.nextFactorID++
	}
}

func (s *RiskScoringService) generateMockData() {
	clientNames := make([]string, 200)
	baseNames := []string{
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
	}

	for i := 0; i < 200; i++ {
		if i < len(baseNames) {
			clientNames[i] = baseNames[i]
		} else {
			clientNames[i] = fmt.Sprintf("Client %d", i+1)
		}
	}

	now := time.Now()

	// Generate clients with risk scores
	for clientID := int64(1); clientID <= int64(len(clientNames)); clientID++ {
		clientName := clientNames[clientID-1]

		// Generate factor scores (0-100 for each factor)
		factorScores := make(map[string]float64)
		factorScores["Trading Volume"] = rand.Float64() * 100
		factorScores["Drawdown History"] = rand.Float64() * 100
		factorScores["Leverage Usage"] = rand.Float64() * 100
		factorScores["Deposit/Withdrawal Ratio"] = rand.Float64() * 100
		factorScores["Account Age"] = rand.Float64() * 100
		factorScores["Margin Utilization"] = rand.Float64() * 100
		factorScores["Loss Streak"] = rand.Float64() * 100
		factorScores["Geographic Risk"] = rand.Float64() * 100

		// Calculate weighted overall score
		overallScore := 0.0
		for _, factor := range s.factors {
			overallScore += (factorScores[factor.Name] * factor.Weight / 100.0)
		}

		category := s.getCategory(overallScore)

		// Generate 30-day score history
		history := make([]ScoreHistoryPoint, 30)
		currentScore := overallScore
		for i := 29; i >= 0; i-- {
			date := now.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
			// Add random variation (±5 points)
			variation := -5.0 + rand.Float64()*10.0
			currentScore = currentScore + variation
			if currentScore < 0 {
				currentScore = 0
			}
			if currentScore > 100 {
				currentScore = 100
			}
			history[29-i] = ScoreHistoryPoint{
				Date:  date,
				Score: currentScore,
			}
		}

		previousScore := history[0].Score

		client := &ClientRiskScore{
			ClientID:       clientID,
			ClientName:     clientName,
			OverallScore:   overallScore,
			Category:       category,
			FactorScores:   factorScores,
			ScoreHistory:   history,
			LastCalculated: now,
			LastChanged:    now.Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour),
			PreviousScore:  previousScore,
		}
		s.clients[clientID] = client

		// Generate alerts for clients with significant score changes
		scoreDiff := overallScore - previousScore
		if scoreDiff > 10 || scoreDiff < -10 {
			alert := &ClientRiskAlert{
				ClientID:    clientID,
				ClientName:  clientName,
				OldScore:    previousScore,
				NewScore:    overallScore,
				Change:      scoreDiff,
				OldCategory: s.getCategory(previousScore),
				NewCategory: category,
				DetectedAt:  now.Add(-time.Duration(rand.Intn(7)) * 24 * time.Hour),
			}
			s.alerts = append(s.alerts, alert)
		}
	}

	log.Printf("[RiskScoring] Generated %d clients with risk scores, 8 risk factors, %d alerts", len(s.clients), len(s.alerts))
}

func (s *RiskScoringService) getCategory(score float64) string {
	if score >= 76 {
		return "Critical"
	} else if score >= 51 {
		return "High"
	} else if score >= 26 {
		return "Medium"
	}
	return "Low"
}

func (s *RiskScoringService) ListClients(sortBy, sortOrder, category string) []*ClientRiskScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*ClientRiskScore, 0, len(s.clients))
	for _, c := range s.clients {
		if category != "" && c.Category != category {
			continue
		}
		// Create copy without score history for list view
		clientCopy := &ClientRiskScore{
			ClientID:       c.ClientID,
			ClientName:     c.ClientName,
			OverallScore:   c.OverallScore,
			Category:       c.Category,
			FactorScores:   c.FactorScores,
			LastCalculated: c.LastCalculated,
			LastChanged:    c.LastChanged,
			PreviousScore:  c.PreviousScore,
		}
		clients = append(clients, clientCopy)
	}

	switch sortBy {
	case "score":
		sort.Slice(clients, func(i, j int) bool {
			if sortOrder == "asc" {
				return clients[i].OverallScore < clients[j].OverallScore
			}
			return clients[i].OverallScore > clients[j].OverallScore
		})
	case "name":
		sort.Slice(clients, func(i, j int) bool {
			if sortOrder == "asc" {
				return clients[i].ClientName < clients[j].ClientName
			}
			return clients[i].ClientName > clients[j].ClientName
		})
	case "category":
		categoryOrder := map[string]int{"Critical": 4, "High": 3, "Medium": 2, "Low": 1}
		sort.Slice(clients, func(i, j int) bool {
			if sortOrder == "asc" {
				return categoryOrder[clients[i].Category] < categoryOrder[clients[j].Category]
			}
			return categoryOrder[clients[i].Category] > categoryOrder[clients[j].Category]
		})
	default:
		sort.Slice(clients, func(i, j int) bool {
			return clients[i].ClientID < clients[j].ClientID
		})
	}

	return clients
}

func (s *RiskScoringService) GetClient(clientID int64) *ClientRiskScore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clients[clientID]
}

func (s *RiskScoringService) RecalculateScore(clientID int64) (*ClientRiskScore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client not found")
	}

	// Recalculate with slight random variation
	newOverallScore := client.OverallScore + (-5.0 + rand.Float64()*10.0)
	if newOverallScore < 0 {
		newOverallScore = 0
	}
	if newOverallScore > 100 {
		newOverallScore = 100
	}

	client.PreviousScore = client.OverallScore
	client.OverallScore = newOverallScore
	client.Category = s.getCategory(newOverallScore)
	client.LastCalculated = time.Now()
	client.LastChanged = time.Now()

	// Add to score history
	newPoint := ScoreHistoryPoint{
		Date:  time.Now().Format("2006-01-02"),
		Score: newOverallScore,
	}
	client.ScoreHistory = append(client.ScoreHistory[1:], newPoint)

	return client, nil
}

func (s *RiskScoringService) ListFactors() []*RiskFactor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	factors := make([]*RiskFactor, 0, len(s.factors))
	for _, f := range s.factors {
		factors = append(factors, f)
	}

	sort.Slice(factors, func(i, j int) bool {
		return factors[i].ID < factors[j].ID
	})

	return factors
}

func (s *RiskScoringService) UpdateFactor(id int64, newWeight float64) (*RiskFactor, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	factor, exists := s.factors[id]
	if !exists {
		return nil, fmt.Errorf("factor not found")
	}

	// Validate total weight is still 100%
	totalWeight := 0.0
	for fid, f := range s.factors {
		if fid == id {
			totalWeight += newWeight
		} else {
			totalWeight += f.Weight
		}
	}

	if totalWeight != 100.0 {
		return nil, fmt.Errorf("total weight must equal 100%%, current total would be %.2f%%", totalWeight)
	}

	factor.Weight = newWeight

	return factor, nil
}

func (s *RiskScoringService) GetDistribution() []RiskDistribution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buckets := make(map[string]int)
	bucketRanges := []string{"0-10", "10-20", "20-30", "30-40", "40-50", "50-60", "60-70", "70-80", "80-90", "90-100"}

	for _, bucket := range bucketRanges {
		buckets[bucket] = 0
	}

	total := len(s.clients)

	for _, client := range s.clients {
		bucketIndex := int(client.OverallScore / 10)
		if bucketIndex >= 10 {
			bucketIndex = 9
		}
		buckets[bucketRanges[bucketIndex]]++
	}

	distribution := make([]RiskDistribution, 0, len(bucketRanges))
	for _, bucket := range bucketRanges {
		count := buckets[bucket]
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100.0
		}
		distribution = append(distribution, RiskDistribution{
			Bucket:     bucket,
			Count:      count,
			Percentage: percentage,
		})
	}

	return distribution
}

func (s *RiskScoringService) GetAlerts() []*ClientRiskAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	alerts := make([]*ClientRiskAlert, len(s.alerts))
	copy(alerts, s.alerts)

	// Sort by detected time desc
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].DetectedAt.After(alerts[j].DetectedAt)
	})

	return alerts
}

func (s *RiskScoringService) GetStats() RiskStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := RiskStats{
		LastUpdated: time.Now(),
	}

	scores := make([]float64, 0, len(s.clients))

	for _, client := range s.clients {
		stats.TotalClients++
		scores = append(scores, client.OverallScore)

		switch client.Category {
		case "Low":
			stats.LowRiskCount++
		case "Medium":
			stats.MediumRiskCount++
		case "High":
			stats.HighRiskCount++
		case "Critical":
			stats.CriticalCount++
		}
	}

	// Calculate average
	if len(scores) > 0 {
		sum := 0.0
		for _, score := range scores {
			sum += score
		}
		stats.AverageScore = sum / float64(len(scores))

		// Calculate median
		sort.Float64s(scores)
		mid := len(scores) / 2
		if len(scores)%2 == 0 {
			stats.MedianScore = (scores[mid-1] + scores[mid]) / 2
		} else {
			stats.MedianScore = scores[mid]
		}
	}

	// Calculate trend (compare recent scores to older scores)
	recentAvg := 0.0
	olderAvg := 0.0
	recentCount := 0
	olderCount := 0

	for _, client := range s.clients {
		if len(client.ScoreHistory) >= 15 {
			// Last 7 days
			for i := len(client.ScoreHistory) - 7; i < len(client.ScoreHistory); i++ {
				recentAvg += client.ScoreHistory[i].Score
				recentCount++
			}
			// Days 15-22 (older reference)
			for i := 7; i < 14; i++ {
				olderAvg += client.ScoreHistory[i].Score
				olderCount++
			}
		}
	}

	if recentCount > 0 && olderCount > 0 {
		recentAvg /= float64(recentCount)
		olderAvg /= float64(olderCount)

		diff := recentAvg - olderAvg
		if diff < -2 {
			stats.ScoreTrend = "improving" // Lower scores are better
		} else if diff > 2 {
			stats.ScoreTrend = "deteriorating"
		} else {
			stats.ScoreTrend = "stable"
		}
	} else {
		stats.ScoreTrend = "stable"
	}

	return stats
}

// ============================================
// HTTP Handlers
// ============================================

type RiskScoringHandler struct {
	service     *RiskScoringService
	authService *auth.Service
}

func NewRiskScoringHandler(service *RiskScoringService, authService *auth.Service) *RiskScoringHandler {
	return &RiskScoringHandler{
		service:     service,
		authService: authService,
	}
}

// GET /admin/risk-scoring/clients
func (h *RiskScoringHandler) HandleListClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query()
	sortBy := query.Get("sortBy")
	sortOrder := query.Get("sortOrder")
	if sortOrder == "" {
		sortOrder = "desc"
	}
	category := query.Get("category")

	clients := h.service.ListClients(sortBy, sortOrder, category)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"count":   len(clients),
	})
}

// GET /admin/risk-scoring/clients/:id
func (h *RiskScoringHandler) HandleGetClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/risk-scoring/clients/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	client := h.service.GetClient(clientID)
	if client == nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(client)
}

// POST /admin/risk-scoring/clients/:id/recalculate
func (h *RiskScoringHandler) HandleRecalculate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/risk-scoring/clients/"), "/")
	clientID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	client, err := h.service.RecalculateScore(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(client)
}

// GET /admin/risk-scoring/factors
func (h *RiskScoringHandler) HandleListFactors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	factors := h.service.ListFactors()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"factors": factors,
		"count":   len(factors),
	})
}

// PUT /admin/risk-scoring/factors/:id
func (h *RiskScoringHandler) HandleUpdateFactor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/risk-scoring/factors/"), "/")
	factorID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid factor ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Weight float64 `json:"weight"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	factor, err := h.service.UpdateFactor(factorID, req.Weight)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(factor)
}

// GET /admin/risk-scoring/distribution
func (h *RiskScoringHandler) HandleGetDistribution(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	distribution := h.service.GetDistribution()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"distribution": distribution,
		"buckets":      len(distribution),
	})
}

// GET /admin/risk-scoring/alerts
func (h *RiskScoringHandler) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	alerts := h.service.GetAlerts()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GET /admin/risk-scoring/stats
func (h *RiskScoringHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats := h.service.GetStats()
	json.NewEncoder(w).Encode(stats)
}
