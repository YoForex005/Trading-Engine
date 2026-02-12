package admin

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

// MAMGroup represents a Multi-Account Manager group
type MAMGroup struct {
	ID                 string    `json:"id"`
	ManagerAccountID   string    `json:"managerAccountId"`
	ManagerName        string    `json:"managerName"`
	StrategyName       string    `json:"strategyName"`
	AllocationMethod   string    `json:"allocationMethod"` // lot, percent, equity
	ManagementFeePct   float64   `json:"managementFeePct"`
	PerformanceFeePct  float64   `json:"performanceFeePct"`
	HighWaterMark      bool      `json:"highWaterMark"`
	MinInvestment      float64   `json:"minInvestment"`
	MaxDrawdownLimit   float64   `json:"maxDrawdownLimit"`
	MaxLotSize         float64   `json:"maxLotSize"`
	AllowedSymbols     []string  `json:"allowedSymbols"`
	TotalAUM           float64   `json:"totalAUM"`
	Status             string    `json:"status"` // active, paused, closed
	CreatedAt          time.Time `json:"createdAt"`
	InvestorCount      int       `json:"investorCount"`
	MonthlyReturn      float64   `json:"monthlyReturn"`
	YTDReturn          float64   `json:"ytdReturn"`
}

// Investor represents an investor in a MAM group
type Investor struct {
	ID            string    `json:"id"`
	GroupID       string    `json:"groupId"`
	AccountID     string    `json:"accountId"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	AllocatedPct  float64   `json:"allocatedPct"`
	CurrentEquity float64   `json:"currentEquity"`
	InitialEquity float64   `json:"initialEquity"`
	PnL           float64   `json:"pnl"`
	PnLPct        float64   `json:"pnlPct"`
	JoinDate      time.Time `json:"joinDate"`
	Status        string    `json:"status"` // active, pending, withdrawn
}

// MonthlyPerformance represents monthly performance metrics
type MonthlyPerformance struct {
	Month      string  `json:"month"`      // YYYY-MM
	Return     float64 `json:"return"`     // Monthly return %
	AUM        float64 `json:"aum"`        // Assets under management
	Trades     int     `json:"trades"`     // Number of trades
	WinRate    float64 `json:"winRate"`    // Win rate %
	Drawdown   float64 `json:"drawdown"`   // Max drawdown %
	Sharpe     float64 `json:"sharpe"`     // Sharpe ratio
}

// MAMStats represents overall MAM statistics
type MAMStats struct {
	TotalAUM         float64 `json:"totalAUM"`
	ActiveManagers   int     `json:"activeManagers"`
	TotalInvestors   int     `json:"totalInvestors"`
	AvgPerformance   float64 `json:"avgPerformance"`
	TotalGroups      int     `json:"totalGroups"`
	PausedGroups     int     `json:"pausedGroups"`
	ClosedGroups     int     `json:"closedGroups"`
}

// MAMStore manages MAM groups and investors
type MAMStore struct {
	mu              sync.RWMutex
	groups          map[string]*MAMGroup
	investors       map[string]*Investor
	groupInvestors  map[string][]string // groupID -> []investorID
	performance     map[string][]MonthlyPerformance // groupID -> []performance
}

// NewMAMStore creates a new MAM store with mock data
func NewMAMStore() *MAMStore {
	store := &MAMStore{
		groups:         make(map[string]*MAMGroup),
		investors:      make(map[string]*Investor),
		groupInvestors: make(map[string][]string),
		performance:    make(map[string][]MonthlyPerformance),
	}

	store.initMockData()
	return store
}

func (s *MAMStore) initMockData() {
	// Initialize 8 mock MAM groups
	mockGroups := []*MAMGroup{
		{
			ID:                "mam-001",
			ManagerAccountID:  "mgr-10001",
			ManagerName:       "Sarah Thompson",
			StrategyName:      "Conservative Growth",
			AllocationMethod:  "equity",
			ManagementFeePct:  2.0,
			PerformanceFeePct: 20.0,
			HighWaterMark:     true,
			MinInvestment:     10000,
			MaxDrawdownLimit:  15.0,
			MaxLotSize:        10.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD"},
			TotalAUM:          485000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-2, -3, 0),
			MonthlyReturn:     2.3,
			YTDReturn:         12.8,
		},
		{
			ID:                "mam-002",
			ManagerAccountID:  "mgr-10002",
			ManagerName:       "Michael Chen",
			StrategyName:      "Aggressive Scalping",
			AllocationMethod:  "lot",
			ManagementFeePct:  1.5,
			PerformanceFeePct: 25.0,
			HighWaterMark:     true,
			MinInvestment:     5000,
			MaxDrawdownLimit:  25.0,
			MaxLotSize:        20.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD"},
			TotalAUM:          720000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-1, -8, 0),
			MonthlyReturn:     4.5,
			YTDReturn:         28.3,
		},
		{
			ID:                "mam-003",
			ManagerAccountID:  "mgr-10003",
			ManagerName:       "Emily Rodriguez",
			StrategyName:      "Trend Following",
			AllocationMethod:  "percent",
			ManagementFeePct:  2.5,
			PerformanceFeePct: 15.0,
			HighWaterMark:     false,
			MinInvestment:     25000,
			MaxDrawdownLimit:  20.0,
			MaxLotSize:        15.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "XAUUSD", "BTCUSD"},
			TotalAUM:          950000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-3, 0, 0),
			MonthlyReturn:     3.2,
			YTDReturn:         18.5,
		},
		{
			ID:                "mam-004",
			ManagerAccountID:  "mgr-10004",
			ManagerName:       "David Park",
			StrategyName:      "Range Trading",
			AllocationMethod:  "equity",
			ManagementFeePct:  2.0,
			PerformanceFeePct: 20.0,
			HighWaterMark:     true,
			MinInvestment:     15000,
			MaxDrawdownLimit:  18.0,
			MaxLotSize:        12.0,
			AllowedSymbols:    []string{"EURUSD", "USDJPY", "GBPJPY", "EURJPY"},
			TotalAUM:          340000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-1, -2, 0),
			MonthlyReturn:     1.8,
			YTDReturn:         9.2,
		},
		{
			ID:                "mam-005",
			ManagerAccountID:  "mgr-10005",
			ManagerName:       "Lisa Anderson",
			StrategyName:      "Momentum Trading",
			AllocationMethod:  "percent",
			ManagementFeePct:  3.0,
			PerformanceFeePct: 30.0,
			HighWaterMark:     true,
			MinInvestment:     50000,
			MaxDrawdownLimit:  30.0,
			MaxLotSize:        25.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD", "BTCUSD", "ETHUSD"},
			TotalAUM:          1250000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-4, -6, 0),
			MonthlyReturn:     5.7,
			YTDReturn:         38.9,
		},
		{
			ID:                "mam-006",
			ManagerAccountID:  "mgr-10006",
			ManagerName:       "Robert Kim",
			StrategyName:      "Mean Reversion",
			AllocationMethod:  "lot",
			ManagementFeePct:  1.8,
			PerformanceFeePct: 18.0,
			HighWaterMark:     false,
			MinInvestment:     20000,
			MaxDrawdownLimit:  22.0,
			MaxLotSize:        18.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "AUDUSD", "NZDUSD"},
			TotalAUM:          580000,
			Status:            "paused",
			CreatedAt:         time.Now().AddDate(-2, -9, 0),
			MonthlyReturn:     0.0,
			YTDReturn:         8.4,
		},
		{
			ID:                "mam-007",
			ManagerAccountID:  "mgr-10007",
			ManagerName:       "Jennifer Martinez",
			StrategyName:      "Breakout Strategy",
			AllocationMethod:  "equity",
			ManagementFeePct:  2.2,
			PerformanceFeePct: 22.0,
			HighWaterMark:     true,
			MinInvestment:     30000,
			MaxDrawdownLimit:  20.0,
			MaxLotSize:        20.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "USDJPY", "XAUUSD", "XAGUSD"},
			TotalAUM:          780000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-1, -5, 0),
			MonthlyReturn:     3.8,
			YTDReturn:         22.1,
		},
		{
			ID:                "mam-008",
			ManagerAccountID:  "mgr-10008",
			ManagerName:       "Thomas Wilson",
			StrategyName:      "News Trading",
			AllocationMethod:  "percent",
			ManagementFeePct:  2.5,
			PerformanceFeePct: 25.0,
			HighWaterMark:     true,
			MinInvestment:     40000,
			MaxDrawdownLimit:  28.0,
			MaxLotSize:        30.0,
			AllowedSymbols:    []string{"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "USDCAD", "AUDUSD"},
			TotalAUM:          620000,
			Status:            "active",
			CreatedAt:         time.Now().AddDate(-2, -1, 0),
			MonthlyReturn:     4.1,
			YTDReturn:         24.6,
		},
	}

	for _, group := range mockGroups {
		s.groups[group.ID] = group
	}

	// Initialize 40+ mock investors across the 8 groups
	s.initMockInvestors()

	// Initialize 12 months of performance history for each group
	s.initPerformanceHistory()
}

func (s *MAMStore) initMockInvestors() {
	investorNames := []string{
		"John Smith", "Mary Johnson", "James Brown", "Patricia Davis", "Robert Miller",
		"Linda Wilson", "Michael Moore", "Barbara Taylor", "William Anderson", "Elizabeth Thomas",
		"David Jackson", "Jennifer White", "Richard Harris", "Susan Martin", "Joseph Thompson",
		"Jessica Garcia", "Charles Martinez", "Sarah Robinson", "Christopher Clark", "Karen Rodriguez",
		"Daniel Lewis", "Nancy Lee", "Matthew Walker", "Betty Hall", "Anthony Allen",
		"Margaret Young", "Mark Hernandez", "Lisa King", "Donald Wright", "Sandra Lopez",
		"Paul Hill", "Ashley Scott", "Steven Green", "Kimberly Adams", "Andrew Baker",
		"Donna Nelson", "Joshua Carter", "Emily Mitchell", "Kenneth Perez", "Michelle Roberts",
		"Kevin Turner", "Carol Phillips", "Brian Campbell", "Amanda Parker", "George Evans",
	}

	groupIDs := []string{"mam-001", "mam-002", "mam-003", "mam-004", "mam-005", "mam-006", "mam-007", "mam-008"}

	investorID := 1
	for _, groupID := range groupIDs {
		// Each group gets 5-7 investors
		numInvestors := 5 + (investorID % 3)

		for i := 0; i < numInvestors && investorID <= len(investorNames); i++ {
			name := investorNames[investorID-1]
			email := strings.ToLower(strings.ReplaceAll(name, " ", ".")) + "@investor.com"

			initialEquity := 10000.0 + float64(investorID)*5000
			pnl := -2000.0 + float64(investorID)*800
			currentEquity := initialEquity + pnl
			pnlPct := (pnl / initialEquity) * 100

			investor := &Investor{
				ID:            generateInvestorID(investorID),
				GroupID:       groupID,
				AccountID:     generateAccountID(investorID),
				Name:          name,
				Email:         email,
				AllocatedPct:  float64(10 + (investorID % 15)),
				CurrentEquity: currentEquity,
				InitialEquity: initialEquity,
				PnL:           pnl,
				PnLPct:        pnlPct,
				JoinDate:      time.Now().AddDate(0, -1*(investorID%12), -(investorID%28)),
				Status:        "active",
			}

			s.investors[investor.ID] = investor
			s.groupInvestors[groupID] = append(s.groupInvestors[groupID], investor.ID)
			investorID++
		}

		// Update group investor count
		if group, ok := s.groups[groupID]; ok {
			group.InvestorCount = len(s.groupInvestors[groupID])
		}
	}
}

func generateInvestorID(id int) string {
	return formatID("inv", id, 5)
}

func generateAccountID(id int) string {
	return formatID("acc", 50000+id, 6)
}

func formatID(prefix string, num int, digits int) string {
	format := "%s-%0" + string(rune(digits+48)) + "d"
	return strings.Replace(format, "%s", prefix, 1)[:len(prefix)+1] + strings.Repeat("0", digits-len(string(rune(num+48)))) + string(rune(num+48))
}

func (s *MAMStore) initPerformanceHistory() {
	// Generate 12 months of performance data for each group
	now := time.Now()

	baseReturns := map[string][]float64{
		"mam-001": {1.8, 2.2, 1.5, 2.8, 1.9, 2.4, 2.1, 1.7, 2.5, 2.3, 2.0, 2.3},
		"mam-002": {3.5, 4.2, 5.1, 3.8, 4.5, 6.2, 3.9, 4.8, 5.5, 4.1, 3.7, 4.5},
		"mam-003": {2.1, 2.8, 3.5, 2.4, 2.9, 3.2, 2.7, 3.1, 2.6, 3.4, 3.0, 3.2},
		"mam-004": {1.2, 1.5, 0.8, 1.9, 1.4, 1.6, 1.3, 1.1, 1.7, 2.0, 1.5, 1.8},
		"mam-005": {4.8, 5.5, 6.2, 4.9, 5.8, 7.1, 5.2, 6.5, 5.9, 6.8, 5.4, 5.7},
		"mam-006": {1.5, 1.8, 1.2, 2.1, 1.6, 0.9, 1.4, 1.0, 0.0, 0.0, 0.0, 0.0},
		"mam-007": {3.2, 3.8, 4.1, 3.5, 3.9, 4.5, 3.7, 4.2, 3.6, 4.0, 3.4, 3.8},
		"mam-008": {3.8, 4.3, 4.7, 4.0, 4.5, 5.2, 4.1, 4.8, 4.4, 4.9, 3.9, 4.1},
	}

	for groupID, returns := range baseReturns {
		performance := make([]MonthlyPerformance, 12)
		group := s.groups[groupID]
		baseAUM := group.TotalAUM * 0.8 // Start with 80% of current AUM

		for i := 0; i < 12; i++ {
			month := now.AddDate(0, -11+i, 0)
			aum := baseAUM * (1.0 + float64(i)*0.025)

			performance[i] = MonthlyPerformance{
				Month:    month.Format("2006-01"),
				Return:   returns[i],
				AUM:      aum,
				Trades:   50 + i*10,
				WinRate:  55.0 + float64(i%5),
				Drawdown: 5.0 + float64(i%8),
				Sharpe:   1.2 + float64(i%10)*0.1,
			}
		}

		s.performance[groupID] = performance
	}
}

// ListGroups returns all MAM groups
func (s *MAMStore) ListGroups() []*MAMGroup {
	s.mu.RLock()
	defer s.mu.RUnlock()

	groups := make([]*MAMGroup, 0, len(s.groups))
	for _, group := range s.groups {
		groups = append(groups, group)
	}
	return groups
}

// GetGroup returns a specific MAM group with investor list
func (s *MAMStore) GetGroup(id string) (*MAMGroup, []*Investor, []MonthlyPerformance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	group, ok := s.groups[id]
	if !ok {
		return nil, nil, nil, false
	}

	// Get investors for this group
	investorIDs := s.groupInvestors[id]
	investors := make([]*Investor, 0, len(investorIDs))
	for _, invID := range investorIDs {
		if investor, ok := s.investors[invID]; ok {
			investors = append(investors, investor)
		}
	}

	// Get performance history
	performance := s.performance[id]

	return group, investors, performance, true
}

// CreateGroup creates a new MAM group
func (s *MAMStore) CreateGroup(group *MAMGroup) *MAMGroup {
	s.mu.Lock()
	defer s.mu.Unlock()

	group.CreatedAt = time.Now()
	group.Status = "active"
	group.TotalAUM = 0
	group.InvestorCount = 0

	s.groups[group.ID] = group
	s.groupInvestors[group.ID] = []string{}
	s.performance[group.ID] = []MonthlyPerformance{}

	return group
}

// UpdateGroup updates an existing MAM group
func (s *MAMStore) UpdateGroup(id string, updates *MAMGroup) (*MAMGroup, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, ok := s.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}

	// Update fields
	group.StrategyName = updates.StrategyName
	group.AllocationMethod = updates.AllocationMethod
	group.ManagementFeePct = updates.ManagementFeePct
	group.PerformanceFeePct = updates.PerformanceFeePct
	group.HighWaterMark = updates.HighWaterMark
	group.MinInvestment = updates.MinInvestment
	group.MaxDrawdownLimit = updates.MaxDrawdownLimit
	group.MaxLotSize = updates.MaxLotSize
	group.AllowedSymbols = updates.AllowedSymbols

	return group, nil
}

// DeleteGroup deactivates a MAM group
func (s *MAMStore) DeleteGroup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, ok := s.groups[id]
	if !ok {
		return ErrGroupNotFound
	}

	group.Status = "closed"
	return nil
}

// GetGroupInvestors returns all investors in a group
func (s *MAMStore) GetGroupInvestors(groupID string) ([]*Investor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.groups[groupID]; !ok {
		return nil, ErrGroupNotFound
	}

	investorIDs := s.groupInvestors[groupID]
	investors := make([]*Investor, 0, len(investorIDs))
	for _, invID := range investorIDs {
		if investor, ok := s.investors[invID]; ok {
			investors = append(investors, investor)
		}
	}

	return investors, nil
}

// AddInvestor adds an investor to a MAM group
func (s *MAMStore) AddInvestor(investor *Investor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, ok := s.groups[investor.GroupID]
	if !ok {
		return ErrGroupNotFound
	}

	investor.JoinDate = time.Now()
	investor.Status = "active"

	s.investors[investor.ID] = investor
	s.groupInvestors[investor.GroupID] = append(s.groupInvestors[investor.GroupID], investor.ID)

	// Update group AUM and investor count
	group.TotalAUM += investor.CurrentEquity
	group.InvestorCount = len(s.groupInvestors[investor.GroupID])

	return nil
}

// RemoveInvestor removes an investor from a MAM group
func (s *MAMStore) RemoveInvestor(groupID, investorID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	group, ok := s.groups[groupID]
	if !ok {
		return ErrGroupNotFound
	}

	investor, ok := s.investors[investorID]
	if !ok {
		return ErrInvestorNotFound
	}

	// Update investor status
	investor.Status = "withdrawn"

	// Update group AUM
	group.TotalAUM -= investor.CurrentEquity

	// Remove from group investors list
	newInvestorList := []string{}
	for _, invID := range s.groupInvestors[groupID] {
		if invID != investorID {
			newInvestorList = append(newInvestorList, invID)
		}
	}
	s.groupInvestors[groupID] = newInvestorList
	group.InvestorCount = len(newInvestorList)

	return nil
}

// GetStats returns overall MAM statistics
func (s *MAMStore) GetStats() MAMStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalAUM float64
	var activeManagers int
	var totalInvestors int
	var totalPerformance float64
	var pausedGroups int
	var closedGroups int

	for _, group := range s.groups {
		totalAUM += group.TotalAUM
		if group.Status == "active" {
			activeManagers++
			totalPerformance += group.MonthlyReturn
		} else if group.Status == "paused" {
			pausedGroups++
		} else if group.Status == "closed" {
			closedGroups++
		}
	}

	for _, investor := range s.investors {
		if investor.Status == "active" {
			totalInvestors++
		}
	}

	avgPerformance := 0.0
	if activeManagers > 0 {
		avgPerformance = totalPerformance / float64(activeManagers)
	}

	return MAMStats{
		TotalAUM:       totalAUM,
		ActiveManagers: activeManagers,
		TotalInvestors: totalInvestors,
		AvgPerformance: avgPerformance,
		TotalGroups:    len(s.groups),
		PausedGroups:   pausedGroups,
		ClosedGroups:   closedGroups,
	}
}

// Custom errors
var (
	ErrGroupNotFound    = &MAMError{"MAM group not found"}
	ErrInvestorNotFound = &MAMError{"Investor not found"}
)

type MAMError struct {
	Message string
}

func (e *MAMError) Error() string {
	return e.Message
}

// ============================================
// HTTP Handlers
// ============================================

type MAMHandler struct {
	store       *MAMStore
	authService *auth.Service
}

func NewMAMHandler(store *MAMStore, authService *auth.Service) *MAMHandler {
	return &MAMHandler{
		store:       store,
		authService: authService,
	}
}

// HandleListGroups returns all MAM groups with summary stats
func (h *MAMHandler) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	groups := h.store.ListGroups()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": groups,
		"total":  len(groups),
	})
}

// HandleGetGroup returns group details with investor list and performance
func (h *MAMHandler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	groupID := pathParts[3]

	group, investors, performance, ok := h.store.GetGroup(groupID)
	if !ok {
		http.Error(w, "MAM group not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"group":       group,
		"investors":   investors,
		"performance": performance,
	})
}

// HandleCreateGroup creates a new MAM group
func (h *MAMHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var group MAMGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created := h.store.CreateGroup(&group)
	log.Printf("[MAM] Created MAM group: %s (%s)", created.StrategyName, created.ID)

	json.NewEncoder(w).Encode(created)
}

// HandleUpdateGroup updates an existing MAM group
func (h *MAMHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	groupID := pathParts[3]

	var updates MAMGroup
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.store.UpdateGroup(groupID, &updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[MAM] Updated MAM group: %s (%s)", updated.StrategyName, updated.ID)
	json.NewEncoder(w).Encode(updated)
}

// HandleDeleteGroup deactivates a MAM group
func (h *MAMHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}
	groupID := pathParts[3]

	err := h.store.DeleteGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[MAM] Deactivated MAM group: %s", groupID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "MAM group deactivated successfully",
	})
}

// HandleGetGroupInvestors returns all investors in a group
func (h *MAMHandler) HandleGetGroupInvestors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	groupID := ""
	for i, part := range pathParts {
		if part == "mam" && i+1 < len(pathParts) {
			groupID = pathParts[i+1]
			break
		}
	}

	if groupID == "" {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	investors, err := h.store.GetGroupInvestors(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"investors": investors,
		"total":     len(investors),
	})
}

// HandleAddInvestor adds an investor to a MAM group
func (h *MAMHandler) HandleAddInvestor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract group ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	groupID := ""
	for i, part := range pathParts {
		if part == "mam" && i+1 < len(pathParts) {
			groupID = pathParts[i+1]
			break
		}
	}

	if groupID == "" {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var investor Investor
	if err := json.NewDecoder(r.Body).Decode(&investor); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	investor.GroupID = groupID

	err := h.store.AddInvestor(&investor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[MAM] Added investor %s to group %s", investor.Name, groupID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"investor": investor,
	})
}

// HandleRemoveInvestor removes an investor from a MAM group
func (h *MAMHandler) HandleRemoveInvestor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract group ID and investor ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	groupID := ""
	investorID := ""

	for i, part := range pathParts {
		if part == "mam" && i+1 < len(pathParts) {
			groupID = pathParts[i+1]
		}
		if part == "investors" && i+1 < len(pathParts) {
			investorID = pathParts[i+1]
		}
	}

	if groupID == "" || investorID == "" {
		http.Error(w, "Invalid group ID or investor ID", http.StatusBadRequest)
		return
	}

	err := h.store.RemoveInvestor(groupID, investorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("[MAM] Removed investor %s from group %s", investorID, groupID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Investor removed successfully",
	})
}

// HandleGetStats returns overall MAM statistics
func (h *MAMHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	stats := h.store.GetStats()
	json.NewEncoder(w).Encode(stats)
}
