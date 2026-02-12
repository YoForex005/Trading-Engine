package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Data Structures
// ============================================

// MarginAccount represents an account's margin status
type MarginAccount struct {
	AccountID     int64     `json:"accountId"`
	ClientName    string    `json:"clientName"`
	Equity        float64   `json:"equity"`
	MarginUsed    float64   `json:"marginUsed"`
	FreeMargin    float64   `json:"freeMargin"`
	MarginLevel   float64   `json:"marginLevel"` // Percentage (equity/marginUsed * 100)
	Status        string    `json:"status"`      // safe, warning, margin_call, stop_out
	OpenPositions int       `json:"openPositions"`
	LastUpdated   time.Time `json:"lastUpdated"`
}

// MarginCallEvent represents a historical margin call event
type MarginCallEvent struct {
	ID          int64     `json:"id"`
	AccountID   int64     `json:"accountId"`
	ClientName  string    `json:"clientName"`
	EventType   string    `json:"eventType"` // warning, margin_call, stop_out, liquidation
	TriggerLevel float64  `json:"triggerLevel"`
	ActionTaken string    `json:"actionTaken"`
	PnLImpact   float64   `json:"pnlImpact"`
	Timestamp   time.Time `json:"timestamp"`
}

// MarginSettings represents global margin call settings
type MarginSettings struct {
	MarginCallLevel float64 `json:"marginCallLevel"` // Default 100%
	StopOutLevel    float64 `json:"stopOutLevel"`    // Default 50%
	AutoNotify      bool    `json:"autoNotify"`
	AutoClose       bool    `json:"autoClose"`
	NotifyEmail     bool    `json:"notifyEmail"`
	NotifySMS       bool    `json:"notifySMS"`
	NotifyPush      bool    `json:"notifyPush"`
}

// MarginCallStats represents aggregate statistics
type MarginCallStats struct {
	AtRiskCount       int     `json:"atRiskCount"`       // Accounts in warning/margin_call/stop_out
	ActiveMarginCalls int     `json:"activeMarginCalls"` // Accounts in margin_call status
	LiquidationsToday int     `json:"liquidationsToday"`
	AvgMarginLevel    float64 `json:"avgMarginLevel"`
}

// ============================================
// In-Memory Store
// ============================================

type MarginCallStore struct {
	mu       sync.RWMutex
	accounts map[int64]*MarginAccount
	events   []*MarginCallEvent
	settings *MarginSettings
	nextEventID int64
}

func NewMarginCallStore() *MarginCallStore {
	store := &MarginCallStore{
		accounts: make(map[int64]*MarginAccount),
		events:   make([]*MarginCallEvent, 0),
		settings: &MarginSettings{
			MarginCallLevel: 100.0,
			StopOutLevel:    50.0,
			AutoNotify:      true,
			AutoClose:       false,
			NotifyEmail:     true,
			NotifySMS:       false,
			NotifyPush:      true,
		},
		nextEventID: 1,
	}

	now := time.Now()

	// ============================================
	// MOCK DATA - 30 ACCOUNTS
	// ============================================

	// Danger Zone (5 accounts) - Margin Level 20-50%
	store.accounts[100001] = &MarginAccount{
		AccountID:     100001,
		ClientName:    "John Martinez",
		Equity:        4800.00,
		MarginUsed:    22000.00,
		FreeMargin:    -17200.00,
		MarginLevel:   21.82,
		Status:        "stop_out",
		OpenPositions: 8,
		LastUpdated:   now.Add(-5 * time.Minute),
	}

	store.accounts[100002] = &MarginAccount{
		AccountID:     100002,
		ClientName:    "Sarah Thompson",
		Equity:        12500.00,
		MarginUsed:    42000.00,
		FreeMargin:    -29500.00,
		MarginLevel:   29.76,
		Status:        "stop_out",
		OpenPositions: 12,
		LastUpdated:   now.Add(-2 * time.Minute),
	}

	store.accounts[100003] = &MarginAccount{
		AccountID:     100003,
		ClientName:    "Michael Chen",
		Equity:        18200.00,
		MarginUsed:    45000.00,
		FreeMargin:    -26800.00,
		MarginLevel:   40.44,
		Status:        "margin_call",
		OpenPositions: 15,
		LastUpdated:   now.Add(-10 * time.Minute),
	}

	store.accounts[100004] = &MarginAccount{
		AccountID:     100004,
		ClientName:    "Emily Rodriguez",
		Equity:        22000.00,
		MarginUsed:    48000.00,
		FreeMargin:    -26000.00,
		MarginLevel:   45.83,
		Status:        "margin_call",
		OpenPositions: 10,
		LastUpdated:   now.Add(-15 * time.Minute),
	}

	store.accounts[100005] = &MarginAccount{
		AccountID:     100005,
		ClientName:    "David Kim",
		Equity:        28000.00,
		MarginUsed:    58000.00,
		FreeMargin:    -30000.00,
		MarginLevel:   48.28,
		Status:        "margin_call",
		OpenPositions: 18,
		LastUpdated:   now.Add(-8 * time.Minute),
	}

	// Warning Zone (8 accounts) - Margin Level 50-100%
	store.accounts[100006] = &MarginAccount{
		AccountID:     100006,
		ClientName:    "Jennifer Lee",
		Equity:        35000.00,
		MarginUsed:    65000.00,
		FreeMargin:    -30000.00,
		MarginLevel:   53.85,
		Status:        "warning",
		OpenPositions: 12,
		LastUpdated:   now.Add(-20 * time.Minute),
	}

	store.accounts[100007] = &MarginAccount{
		AccountID:     100007,
		ClientName:    "Robert Taylor",
		Equity:        48500.00,
		MarginUsed:    75000.00,
		FreeMargin:    -26500.00,
		MarginLevel:   64.67,
		Status:        "warning",
		OpenPositions: 9,
		LastUpdated:   now.Add(-12 * time.Minute),
	}

	store.accounts[100008] = &MarginAccount{
		AccountID:     100008,
		ClientName:    "Lisa Anderson",
		Equity:        58000.00,
		MarginUsed:    82000.00,
		FreeMargin:    -24000.00,
		MarginLevel:   70.73,
		Status:        "warning",
		OpenPositions: 14,
		LastUpdated:   now.Add(-5 * time.Minute),
	}

	store.accounts[100009] = &MarginAccount{
		AccountID:     100009,
		ClientName:    "James Wilson",
		Equity:        72000.00,
		MarginUsed:    95000.00,
		FreeMargin:    -23000.00,
		MarginLevel:   75.79,
		Status:        "warning",
		OpenPositions: 11,
		LastUpdated:   now.Add(-30 * time.Minute),
	}

	store.accounts[100010] = &MarginAccount{
		AccountID:     100010,
		ClientName:    "Amanda Garcia",
		Equity:        85000.00,
		MarginUsed:    105000.00,
		FreeMargin:    -20000.00,
		MarginLevel:   80.95,
		Status:        "warning",
		OpenPositions: 16,
		LastUpdated:   now.Add(-18 * time.Minute),
	}

	store.accounts[100011] = &MarginAccount{
		AccountID:     100011,
		ClientName:    "Christopher Brown",
		Equity:        92000.00,
		MarginUsed:    108000.00,
		FreeMargin:    -16000.00,
		MarginLevel:   85.19,
		Status:        "warning",
		OpenPositions: 13,
		LastUpdated:   now.Add(-25 * time.Minute),
	}

	store.accounts[100012] = &MarginAccount{
		AccountID:     100012,
		ClientName:    "Jessica Martinez",
		Equity:        105000.00,
		MarginUsed:    118000.00,
		FreeMargin:    -13000.00,
		MarginLevel:   88.98,
		Status:        "warning",
		OpenPositions: 10,
		LastUpdated:   now.Add(-10 * time.Minute),
	}

	store.accounts[100013] = &MarginAccount{
		AccountID:     100013,
		ClientName:    "Daniel White",
		Equity:        118000.00,
		MarginUsed:    125000.00,
		FreeMargin:    -7000.00,
		MarginLevel:   94.40,
		Status:        "warning",
		OpenPositions: 8,
		LastUpdated:   now.Add(-15 * time.Minute),
	}

	// Safe Zone (17 accounts) - Margin Level 100%+
	store.accounts[100014] = &MarginAccount{
		AccountID:     100014,
		ClientName:    "Michelle Johnson",
		Equity:        55000.00,
		MarginUsed:    50000.00,
		FreeMargin:    5000.00,
		MarginLevel:   110.00,
		Status:        "safe",
		OpenPositions: 5,
		LastUpdated:   now.Add(-5 * time.Minute),
	}

	store.accounts[100015] = &MarginAccount{
		AccountID:     100015,
		ClientName:    "Kevin Davis",
		Equity:        78000.00,
		MarginUsed:    65000.00,
		FreeMargin:    13000.00,
		MarginLevel:   120.00,
		Status:        "safe",
		OpenPositions: 7,
		LastUpdated:   now.Add(-8 * time.Minute),
	}

	store.accounts[100016] = &MarginAccount{
		AccountID:     100016,
		ClientName:    "Nicole Harris",
		Equity:        98000.00,
		MarginUsed:    72000.00,
		FreeMargin:    26000.00,
		MarginLevel:   136.11,
		Status:        "safe",
		OpenPositions: 6,
		LastUpdated:   now.Add(-12 * time.Minute),
	}

	store.accounts[100017] = &MarginAccount{
		AccountID:     100017,
		ClientName:    "Brandon Clark",
		Equity:        125000.00,
		MarginUsed:    85000.00,
		FreeMargin:    40000.00,
		MarginLevel:   147.06,
		Status:        "safe",
		OpenPositions: 9,
		LastUpdated:   now.Add(-20 * time.Minute),
	}

	store.accounts[100018] = &MarginAccount{
		AccountID:     100018,
		ClientName:    "Stephanie Lewis",
		Equity:        145000.00,
		MarginUsed:    90000.00,
		FreeMargin:    55000.00,
		MarginLevel:   161.11,
		Status:        "safe",
		OpenPositions: 8,
		LastUpdated:   now.Add(-15 * time.Minute),
	}

	store.accounts[100019] = &MarginAccount{
		AccountID:     100019,
		ClientName:    "Jason Walker",
		Equity:        168000.00,
		MarginUsed:    95000.00,
		FreeMargin:    73000.00,
		MarginLevel:   176.84,
		Status:        "safe",
		OpenPositions: 10,
		LastUpdated:   now.Add(-10 * time.Minute),
	}

	store.accounts[100020] = &MarginAccount{
		AccountID:     100020,
		ClientName:    "Rachel Hall",
		Equity:        192000.00,
		MarginUsed:    100000.00,
		FreeMargin:    92000.00,
		MarginLevel:   192.00,
		Status:        "safe",
		OpenPositions: 11,
		LastUpdated:   now.Add(-5 * time.Minute),
	}

	store.accounts[100021] = &MarginAccount{
		AccountID:     100021,
		ClientName:    "Eric Allen",
		Equity:        215000.00,
		MarginUsed:    102000.00,
		FreeMargin:    113000.00,
		MarginLevel:   210.78,
		Status:        "safe",
		OpenPositions: 9,
		LastUpdated:   now.Add(-30 * time.Minute),
	}

	store.accounts[100022] = &MarginAccount{
		AccountID:     100022,
		ClientName:    "Amy Young",
		Equity:        245000.00,
		MarginUsed:    105000.00,
		FreeMargin:    140000.00,
		MarginLevel:   233.33,
		Status:        "safe",
		OpenPositions: 7,
		LastUpdated:   now.Add(-25 * time.Minute),
	}

	store.accounts[100023] = &MarginAccount{
		AccountID:     100023,
		ClientName:    "Ryan King",
		Equity:        278000.00,
		MarginUsed:    110000.00,
		FreeMargin:    168000.00,
		MarginLevel:   252.73,
		Status:        "safe",
		OpenPositions: 12,
		LastUpdated:   now.Add(-18 * time.Minute),
	}

	store.accounts[100024] = &MarginAccount{
		AccountID:     100024,
		ClientName:    "Laura Wright",
		Equity:        312000.00,
		MarginUsed:    115000.00,
		FreeMargin:    197000.00,
		MarginLevel:   271.30,
		Status:        "safe",
		OpenPositions: 8,
		LastUpdated:   now.Add(-12 * time.Minute),
	}

	store.accounts[100025] = &MarginAccount{
		AccountID:     100025,
		ClientName:    "Timothy Scott",
		Equity:        348000.00,
		MarginUsed:    118000.00,
		FreeMargin:    230000.00,
		MarginLevel:   294.92,
		Status:        "safe",
		OpenPositions: 10,
		LastUpdated:   now.Add(-8 * time.Minute),
	}

	store.accounts[100026] = &MarginAccount{
		AccountID:     100026,
		ClientName:    "Rebecca Green",
		Equity:        385000.00,
		MarginUsed:    120000.00,
		FreeMargin:    265000.00,
		MarginLevel:   320.83,
		Status:        "safe",
		OpenPositions: 9,
		LastUpdated:   now.Add(-15 * time.Minute),
	}

	store.accounts[100027] = &MarginAccount{
		AccountID:     100027,
		ClientName:    "Gregory Adams",
		Equity:        425000.00,
		MarginUsed:    122000.00,
		FreeMargin:    303000.00,
		MarginLevel:   348.36,
		Status:        "safe",
		OpenPositions: 11,
		LastUpdated:   now.Add(-20 * time.Minute),
	}

	store.accounts[100028] = &MarginAccount{
		AccountID:     100028,
		ClientName:    "Samantha Baker",
		Equity:        468000.00,
		MarginUsed:    125000.00,
		FreeMargin:    343000.00,
		MarginLevel:   374.40,
		Status:        "safe",
		OpenPositions: 6,
		LastUpdated:   now.Add(-10 * time.Minute),
	}

	store.accounts[100029] = &MarginAccount{
		AccountID:     100029,
		ClientName:    "Benjamin Nelson",
		Equity:        512000.00,
		MarginUsed:    128000.00,
		FreeMargin:    384000.00,
		MarginLevel:   400.00,
		Status:        "safe",
		OpenPositions: 7,
		LastUpdated:   now.Add(-5 * time.Minute),
	}

	store.accounts[100030] = &MarginAccount{
		AccountID:     100030,
		ClientName:    "Victoria Carter",
		Equity:        558000.00,
		MarginUsed:    130000.00,
		FreeMargin:    428000.00,
		MarginLevel:   429.23,
		Status:        "safe",
		OpenPositions: 8,
		LastUpdated:   now.Add(-15 * time.Minute),
	}

	// ============================================
	// MOCK DATA - 15 HISTORICAL EVENTS
	// ============================================

	store.events = append(store.events, &MarginCallEvent{
		ID:           1,
		AccountID:    100001,
		ClientName:   "John Martinez",
		EventType:    "stop_out",
		TriggerLevel: 21.82,
		ActionTaken:  "3 positions auto-closed by system",
		PnLImpact:    -2400.00,
		Timestamp:    now.Add(-5 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           2,
		AccountID:    100002,
		ClientName:   "Sarah Thompson",
		EventType:    "stop_out",
		TriggerLevel: 29.76,
		ActionTaken:  "4 positions auto-closed by system",
		PnLImpact:    -3200.00,
		Timestamp:    now.Add(-2 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           3,
		AccountID:    100003,
		ClientName:   "Michael Chen",
		EventType:    "margin_call",
		TriggerLevel: 40.44,
		ActionTaken:  "Margin call notification sent",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-10 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           4,
		AccountID:    100004,
		ClientName:   "Emily Rodriguez",
		EventType:    "margin_call",
		TriggerLevel: 45.83,
		ActionTaken:  "Margin call notification sent",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-15 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           5,
		AccountID:    100005,
		ClientName:   "David Kim",
		EventType:    "margin_call",
		TriggerLevel: 48.28,
		ActionTaken:  "Margin call notification sent",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-8 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           6,
		AccountID:    100006,
		ClientName:   "Jennifer Lee",
		EventType:    "warning",
		TriggerLevel: 53.85,
		ActionTaken:  "Warning notification sent",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-20 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           7,
		AccountID:    100007,
		ClientName:   "Robert Taylor",
		EventType:    "warning",
		TriggerLevel: 64.67,
		ActionTaken:  "Warning notification sent",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-12 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           8,
		AccountID:    100008,
		ClientName:   "Lisa Anderson",
		EventType:    "warning",
		TriggerLevel: 70.73,
		ActionTaken:  "Warning notification sent",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-5 * time.Minute),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           9,
		AccountID:    100012,
		ClientName:   "Jessica Martinez",
		EventType:    "liquidation",
		TriggerLevel: 18.50,
		ActionTaken:  "All positions liquidated",
		PnLImpact:    -8500.00,
		Timestamp:    now.Add(-2 * time.Hour),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           10,
		AccountID:    100015,
		ClientName:   "Kevin Davis",
		EventType:    "liquidation",
		TriggerLevel: 22.30,
		ActionTaken:  "All positions liquidated",
		PnLImpact:    -6200.00,
		Timestamp:    now.Add(-4 * time.Hour),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           11,
		AccountID:    100018,
		ClientName:   "Stephanie Lewis",
		EventType:    "stop_out",
		TriggerLevel: 42.10,
		ActionTaken:  "2 positions auto-closed by system",
		PnLImpact:    -1800.00,
		Timestamp:    now.Add(-6 * time.Hour),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           12,
		AccountID:    100021,
		ClientName:   "Eric Allen",
		EventType:    "margin_call",
		TriggerLevel: 88.90,
		ActionTaken:  "Margin call notification sent, client added funds",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-8 * time.Hour),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           13,
		AccountID:    100024,
		ClientName:   "Laura Wright",
		EventType:    "warning",
		TriggerLevel: 92.40,
		ActionTaken:  "Warning notification sent, client closed positions",
		PnLImpact:    0.00,
		Timestamp:    now.Add(-12 * time.Hour),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           14,
		AccountID:    100027,
		ClientName:   "Gregory Adams",
		EventType:    "stop_out",
		TriggerLevel: 38.75,
		ActionTaken:  "5 positions auto-closed by system",
		PnLImpact:    -4100.00,
		Timestamp:    now.Add(-18 * time.Hour),
	})

	store.events = append(store.events, &MarginCallEvent{
		ID:           15,
		AccountID:    100030,
		ClientName:   "Victoria Carter",
		EventType:    "liquidation",
		TriggerLevel: 15.20,
		ActionTaken:  "All positions liquidated",
		PnLImpact:    -12300.00,
		Timestamp:    now.Add(-24 * time.Hour),
	})

	store.nextEventID = 16

	log.Printf("[MarginCalls] Initialized margin call system (30 accounts: 5 danger, 8 warning, 17 safe) with 15 historical events")

	return store
}

// ============================================
// Handler
// ============================================

type MarginCallHandler struct {
	store       *MarginCallStore
	authService *auth.Service
}

func NewMarginCallHandler(store *MarginCallStore, authService *auth.Service) *MarginCallHandler {
	return &MarginCallHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// Handler Methods
// ============================================

// HandleListAccounts returns all margin accounts with optional status filter
// GET /admin/margin-calls/accounts?status=warning
func (h *MarginCallHandler) HandleListAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	statusFilter := r.URL.Query().Get("status")

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	var accounts []*MarginAccount
	for _, acc := range h.store.accounts {
		if statusFilter == "" || acc.Status == statusFilter {
			accounts = append(accounts, acc)
		}
	}

	// Sort by margin level ascending (most at risk first)
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].MarginLevel < accounts[j].MarginLevel
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accounts,
		"total":    len(accounts),
	})
}

// HandleGetAccount returns single account detail
// GET /admin/margin-calls/accounts/:id
func (h *MarginCallHandler) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/margin-calls/accounts/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}

	var accountID int64
	if _, err := fmt.Sscanf(path, "%d", &accountID); err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	h.store.mu.RLock()
	account, exists := h.store.accounts[accountID]
	h.store.mu.RUnlock()

	if !exists {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

// HandleNotify sends margin call notification to account
// POST /admin/margin-calls/notify/:id
func (h *MarginCallHandler) HandleNotify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/margin-calls/notify/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}

	var accountID int64
	if _, err := fmt.Sscanf(path, "%d", &accountID); err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	account, exists := h.store.accounts[accountID]
	if !exists {
		h.store.mu.Unlock()
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	// Create event
	event := &MarginCallEvent{
		ID:           h.store.nextEventID,
		AccountID:    accountID,
		ClientName:   account.ClientName,
		EventType:    "margin_call",
		TriggerLevel: account.MarginLevel,
		ActionTaken:  "Manual margin call notification sent by admin",
		PnLImpact:    0.00,
		Timestamp:    time.Now(),
	}
	h.store.events = append(h.store.events, event)
	h.store.nextEventID++
	h.store.mu.Unlock()

	log.Printf("[MarginCalls] Manual margin call notification sent to account %d (%s) - Margin Level: %.2f%%",
		accountID, account.ClientName, account.MarginLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"message":     "Margin call notification sent successfully",
		"accountId":   accountID,
		"clientName":  account.ClientName,
		"marginLevel": account.MarginLevel,
		"eventId":     event.ID,
	})
}

// HandleCloseAll closes all positions for an account
// POST /admin/margin-calls/close-all/:id
func (h *MarginCallHandler) HandleCloseAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/admin/margin-calls/close-all/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Account ID required", http.StatusBadRequest)
		return
	}

	var accountID int64
	if _, err := fmt.Sscanf(path, "%d", &accountID); err != nil {
		http.Error(w, "Invalid account ID", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	account, exists := h.store.accounts[accountID]
	if !exists {
		h.store.mu.Unlock()
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}

	closedPositions := account.OpenPositions
	oldMarginLevel := account.MarginLevel

	// Simulate closing all positions - reset to safe state
	account.MarginUsed = 0
	account.FreeMargin = account.Equity
	account.MarginLevel = 999.99 // No margin used
	account.Status = "safe"
	account.OpenPositions = 0
	account.LastUpdated = time.Now()

	// Create event
	event := &MarginCallEvent{
		ID:           h.store.nextEventID,
		AccountID:    accountID,
		ClientName:   account.ClientName,
		EventType:    "liquidation",
		TriggerLevel: oldMarginLevel,
		ActionTaken:  fmt.Sprintf("All %d positions manually closed by admin", closedPositions),
		PnLImpact:    -account.Equity * 0.15, // Simulate 15% loss
		Timestamp:    time.Now(),
	}
	h.store.events = append(h.store.events, event)
	h.store.nextEventID++
	h.store.mu.Unlock()

	log.Printf("[MarginCalls] All positions closed for account %d (%s) - %d positions liquidated",
		accountID, account.ClientName, closedPositions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":          true,
		"message":          "All positions closed successfully",
		"accountId":        accountID,
		"clientName":       account.ClientName,
		"positionsClosed":  closedPositions,
		"newMarginLevel":   account.MarginLevel,
		"newStatus":        account.Status,
		"estimatedPnL":     event.PnLImpact,
		"eventId":          event.ID,
	})
}

// HandleGetHistory returns historical margin call events
// GET /admin/margin-calls/history
func (h *MarginCallHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	events := make([]*MarginCallEvent, len(h.store.events))
	copy(events, h.store.events)
	h.store.mu.RUnlock()

	// Sort by timestamp descending (newest first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// HandleGetSettings returns current margin call settings
// GET /admin/margin-calls/settings
func (h *MarginCallHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	settings := h.store.settings
	h.store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// HandleUpdateSettings updates margin call settings
// PUT /admin/margin-calls/settings
func (h *MarginCallHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var updates MarginSettings
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	if updates.MarginCallLevel > 0 {
		h.store.settings.MarginCallLevel = updates.MarginCallLevel
	}
	if updates.StopOutLevel > 0 {
		h.store.settings.StopOutLevel = updates.StopOutLevel
	}
	h.store.settings.AutoNotify = updates.AutoNotify
	h.store.settings.AutoClose = updates.AutoClose
	h.store.settings.NotifyEmail = updates.NotifyEmail
	h.store.settings.NotifySMS = updates.NotifySMS
	h.store.settings.NotifyPush = updates.NotifyPush
	h.store.mu.Unlock()

	log.Printf("[MarginCalls] Settings updated - Margin Call Level: %.2f%%, Stop Out Level: %.2f%%",
		h.store.settings.MarginCallLevel, h.store.settings.StopOutLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.store.settings)
}

// HandleGetStats returns aggregate margin call statistics
// GET /admin/margin-calls/stats
func (h *MarginCallHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Admin auth
	if _, err := h.authService.ValidateAdminToken(r); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	stats := MarginCallStats{
		AtRiskCount:       0,
		ActiveMarginCalls: 0,
		LiquidationsToday: 0,
		AvgMarginLevel:    0,
	}

	totalMarginLevel := 0.0
	for _, acc := range h.store.accounts {
		totalMarginLevel += acc.MarginLevel

		if acc.Status == "warning" || acc.Status == "margin_call" || acc.Status == "stop_out" {
			stats.AtRiskCount++
		}
		if acc.Status == "margin_call" {
			stats.ActiveMarginCalls++
		}
	}

	if len(h.store.accounts) > 0 {
		stats.AvgMarginLevel = totalMarginLevel / float64(len(h.store.accounts))
	}

	// Count liquidations today
	today := time.Now().Truncate(24 * time.Hour)
	for _, event := range h.store.events {
		if event.EventType == "liquidation" && event.Timestamp.After(today) {
			stats.LiquidationsToday++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
