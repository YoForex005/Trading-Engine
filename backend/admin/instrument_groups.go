package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// ============================================
// Types
// ============================================

type SpreadType string

const (
	SpreadTypeFixed    SpreadType = "fixed"
	SpreadTypeVariable SpreadType = "variable"
)

type CommissionModel string

const (
	CommissionModelNone       CommissionModel = "none"
	CommissionModelPerLot     CommissionModel = "per_lot"
	CommissionModelPercentage CommissionModel = "percentage"
)

type MarginMode string

const (
	MarginModeRetail       MarginMode = "retail"
	MarginModeProfessional MarginMode = "professional"
	MarginModeHedge        MarginMode = "hedge"
)

// ============================================
// Data Structures
// ============================================

type InstrumentGroup struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	Description           string          `json:"description"`
	DefaultLeverage       int             `json:"defaultLeverage"`
	SpreadType            SpreadType      `json:"spreadType"`
	CommissionModel       CommissionModel `json:"commissionModel"`
	CommissionRate        float64         `json:"commissionRate"`
	MarginMode            MarginMode      `json:"marginMode"`
	TradingHoursTemplate  string          `json:"tradingHoursTemplate"`
	SymbolCount           int             `json:"symbolCount"`
	IsActive              bool            `json:"isActive"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

type SymbolInGroup struct {
	Symbol          string          `json:"symbol"`
	GroupID         string          `json:"groupId"`
	GroupName       string          `json:"groupName"`
	Leverage        int             `json:"leverage"`
	SpreadType      SpreadType      `json:"spreadType"`
	CommissionModel CommissionModel `json:"commissionModel"`
	CommissionRate  float64         `json:"commissionRate"`
	MarginMode      MarginMode      `json:"marginMode"`
	TradingHours    string          `json:"tradingHours"`
	IsActive        bool            `json:"isActive"`
	AddedAt         time.Time       `json:"addedAt"`
}

type InstrumentStats struct {
	TotalGroups         int    `json:"totalGroups"`
	TotalSymbols        int    `json:"totalSymbols"`
	MostActiveGroupID   string `json:"mostActiveGroupId"`
	MostActiveGroupName string `json:"mostActiveGroupName"`
	MostActiveSymbols   int    `json:"mostActiveSymbols"`
}

// ============================================
// Store
// ============================================

type InstrumentGroupStore struct {
	mu      sync.RWMutex
	groups  map[string]*InstrumentGroup
	symbols map[string]*SymbolInGroup // key: symbol name
}

func NewInstrumentGroupStore() *InstrumentGroupStore {
	store := &InstrumentGroupStore{
		groups:  make(map[string]*InstrumentGroup),
		symbols: make(map[string]*SymbolInGroup),
	}

	// Generate mock data
	store.generateMockData()

	return store
}

func (s *InstrumentGroupStore) generateMockData() {
	now := time.Now()

	// 8 predefined instrument groups
	groups := []struct {
		name             string
		description      string
		leverage         int
		spreadType       SpreadType
		commissionModel  CommissionModel
		commissionRate   float64
		marginMode       MarginMode
		tradingHours     string
	}{
		{
			"Forex Majors",
			"Major currency pairs with high liquidity",
			500,
			SpreadTypeVariable,
			CommissionModelNone,
			0,
			MarginModeRetail,
			"24/5",
		},
		{
			"Forex Minors",
			"Cross currency pairs without USD",
			400,
			SpreadTypeVariable,
			CommissionModelNone,
			0,
			MarginModeRetail,
			"24/5",
		},
		{
			"Forex Exotics",
			"Emerging market currency pairs",
			100,
			SpreadTypeFixed,
			CommissionModelPerLot,
			5.0,
			MarginModeRetail,
			"24/5",
		},
		{
			"Metals",
			"Precious metals (Gold, Silver, Platinum)",
			200,
			SpreadTypeVariable,
			CommissionModelNone,
			0,
			MarginModeRetail,
			"23/5",
		},
		{
			"Energies",
			"Energy commodities (Oil, Gas)",
			100,
			SpreadTypeFixed,
			CommissionModelPerLot,
			3.0,
			MarginModeRetail,
			"23/5",
		},
		{
			"Indices",
			"Stock market indices",
			200,
			SpreadTypeVariable,
			CommissionModelPercentage,
			0.1,
			MarginModeProfessional,
			"Market Hours",
		},
		{
			"Crypto",
			"Cryptocurrency pairs",
			50,
			SpreadTypeVariable,
			CommissionModelPercentage,
			0.2,
			MarginModeRetail,
			"24/7",
		},
		{
			"Commodities",
			"Agricultural and industrial commodities",
			100,
			SpreadTypeFixed,
			CommissionModelPerLot,
			4.0,
			MarginModeRetail,
			"Market Hours",
		},
	}

	for i, g := range groups {
		group := &InstrumentGroup{
			ID:                   fmt.Sprintf("group-%d", i+1),
			Name:                 g.name,
			Description:          g.description,
			DefaultLeverage:      g.leverage,
			SpreadType:           g.spreadType,
			CommissionModel:      g.commissionModel,
			CommissionRate:       g.commissionRate,
			MarginMode:           g.marginMode,
			TradingHoursTemplate: g.tradingHours,
			SymbolCount:          0, // Will update below
			IsActive:             true,
			CreatedAt:            now.Add(-time.Duration(90-i*10) * 24 * time.Hour),
			UpdatedAt:            now.Add(-time.Duration(i) * 24 * time.Hour),
		}
		s.groups[group.ID] = group
	}

	// 150 symbols distributed across 8 groups
	symbolData := []struct {
		groupID string
		symbols []string
	}{
		{
			"group-1", // Forex Majors
			[]string{
				"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "AUDUSD", "USDCAD", "NZDUSD",
				"EURJPY", "GBPJPY", "EURGBP", "AUDJPY", "EURAUD", "EURCHF", "AUDNZD",
				"NZDJPY", "GBPAUD", "GBPNZD", "CHFJPY", "AUDCAD", "CADJPY",
			},
		},
		{
			"group-2", // Forex Minors
			[]string{
				"EURAUD", "EURNZD", "EURCAD", "GBPCAD", "AUDCAD", "NZDCAD",
				"AUDCHF", "NZDCHF", "GBPCHF", "CADCHF", "EURNOK", "EURSEK",
				"GBPSEK", "USDSEK", "USDNOK", "EURPLN", "USDPLN", "EURCZK",
				"USDCZK", "EURHUF",
			},
		},
		{
			"group-3", // Forex Exotics
			[]string{
				"USDTRY", "USDZAR", "USDMXN", "USDCNH", "USDINR", "USDSGD",
				"USDHKD", "USDTHB", "EURSGD", "EURTRY", "EURZAR", "GBPZAR",
				"AUDHKD", "AUDSGD", "NZDSGD", "GBPHKD", "EURHKD", "SGDJPY",
				"ZARJPY", "TRYJPY",
			},
		},
		{
			"group-4", // Metals
			[]string{
				"XAUUSD", "XAGUSD", "XPTUSD", "XPDUSD", "XAUEUR", "XAUGBP",
				"XAUJPY", "XAUAUD", "XAUCAD", "XAUCHF", "XAUHKD", "XAUSGD",
				"XAUNZD", "XAGEUR", "XAGGBP", "XAGJPY", "XAGAUD", "XAGCAD",
				"XAGCHF", "XAUXAG",
			},
		},
		{
			"group-5", // Energies
			[]string{
				"WTICOUSD", "BCOUSD", "NATGASUSD", "WTICOEUR", "BCOEUR",
				"WTICOGBP", "BCOGBP", "NATGASEUR", "NATGASGBP",
				"HEATINGOILUSD", "GASOLINEUSD", "RBOBGASUSD",
			},
		},
		{
			"group-6", // Indices
			[]string{
				"SPX500USD", "US30USD", "NAS100USD", "US2000USD", "DE30EUR",
				"UK100GBP", "FR40EUR", "ESPIXEUR", "JP225USD", "JP225YJPY",
				"AU200AUD", "HK33HKD", "CN50USD", "CHINAHHKD", "SG30SGD",
				"CH20CHF", "NL25EUR", "EU50EUR", "VIX", "DAX40EUR",
			},
		},
		{
			"group-7", // Crypto
			[]string{
				"BTCUSD", "ETHUSD", "XRPUSD", "LTCUSD", "BCHUSD",
				"BNBUSD", "ADAUSD", "SOLUSD", "DOTUSD", "LINKUSD",
				"MATICUSD", "AVAXUSD", "UNIUSD", "ATOMUSD", "XLMUSD",
				"ALGOUSD", "VETUSD", "FILUSD", "TRXUSD", "ETCUSD",
			},
		},
		{
			"group-8", // Commodities
			[]string{
				"CORNUSD", "WHEATUSD", "SOYBNUSD", "SUGARUSD", "COFFEEUSD",
				"COCOAUSD", "COTTONUSD", "LUMBERJUSD", "COPPERUSD", "XCUUSD",
				"ORANGEJUICEUSD", "RICEUSD", "OATSUSD", "PALLADIUMU", "RHODIUMUSD",
				"PLATINUMUSD", "NICKELUSD", "ZINEUSD", "ALUMINIUMUSD",
			},
		},
	}

	for _, data := range symbolData {
		group := s.groups[data.groupID]
		if group == nil {
			continue
		}

		for _, symbolName := range data.symbols {
			symbol := &SymbolInGroup{
				Symbol:          symbolName,
				GroupID:         data.groupID,
				GroupName:       group.Name,
				Leverage:        group.DefaultLeverage,
				SpreadType:      group.SpreadType,
				CommissionModel: group.CommissionModel,
				CommissionRate:  group.CommissionRate,
				MarginMode:      group.MarginMode,
				TradingHours:    group.TradingHoursTemplate,
				IsActive:        true,
				AddedAt:         group.CreatedAt,
			}
			s.symbols[symbolName] = symbol
			group.SymbolCount++
		}
	}

	log.Printf("[InstrumentGroupStore] Mock data generated: %d groups, %d symbols",
		len(s.groups), len(s.symbols))
}

// ============================================
// Handler
// ============================================

type InstrumentGroupHandler struct {
	store       *InstrumentGroupStore
	authService *auth.Service
}

func NewInstrumentGroupHandler(store *InstrumentGroupStore, authService *auth.Service) *InstrumentGroupHandler {
	return &InstrumentGroupHandler{
		store:       store,
		authService: authService,
	}
}

// ============================================
// 1. GET /admin/instruments/groups — List all groups
// ============================================

func (h *InstrumentGroupHandler) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	groups := []*InstrumentGroup{}
	for _, group := range h.store.groups {
		groups = append(groups, group)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": groups,
		"total":  len(groups),
	})
}

// ============================================
// 2. POST /admin/instruments/groups — Create group
// ============================================

func (h *InstrumentGroupHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name                 string          `json:"name"`
		Description          string          `json:"description"`
		DefaultLeverage      int             `json:"defaultLeverage"`
		SpreadType           SpreadType      `json:"spreadType"`
		CommissionModel      CommissionModel `json:"commissionModel"`
		CommissionRate       float64         `json:"commissionRate"`
		MarginMode           MarginMode      `json:"marginMode"`
		TradingHoursTemplate string          `json:"tradingHoursTemplate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	now := time.Now()
	groupID := fmt.Sprintf("group-%d", len(h.store.groups)+1)

	group := &InstrumentGroup{
		ID:                   groupID,
		Name:                 req.Name,
		Description:          req.Description,
		DefaultLeverage:      req.DefaultLeverage,
		SpreadType:           req.SpreadType,
		CommissionModel:      req.CommissionModel,
		CommissionRate:       req.CommissionRate,
		MarginMode:           req.MarginMode,
		TradingHoursTemplate: req.TradingHoursTemplate,
		SymbolCount:          0,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	h.store.groups[groupID] = group

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// ============================================
// 3. GET /admin/instruments/groups/:id — Group details with symbols
// ============================================

func (h *InstrumentGroupHandler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/instruments/groups/"), "/")
	if len(parts) == 0 || parts[0] == "" || parts[0] == "stats" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}
	groupID := parts[0]

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	group, exists := h.store.groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Get symbols in this group
	symbols := []*SymbolInGroup{}
	for _, symbol := range h.store.symbols {
		if symbol.GroupID == groupID {
			symbols = append(symbols, symbol)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"group":   group,
		"symbols": symbols,
		"total":   len(symbols),
	})
}

// ============================================
// 4. PUT /admin/instruments/groups/:id — Update group
// ============================================

func (h *InstrumentGroupHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/instruments/groups/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}
	groupID := parts[0]

	var req struct {
		Description          *string          `json:"description,omitempty"`
		DefaultLeverage      *int             `json:"defaultLeverage,omitempty"`
		SpreadType           *SpreadType      `json:"spreadType,omitempty"`
		CommissionModel      *CommissionModel `json:"commissionModel,omitempty"`
		CommissionRate       *float64         `json:"commissionRate,omitempty"`
		MarginMode           *MarginMode      `json:"marginMode,omitempty"`
		TradingHoursTemplate *string          `json:"tradingHoursTemplate,omitempty"`
		IsActive             *bool            `json:"isActive,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	group, exists := h.store.groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Update fields
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.DefaultLeverage != nil {
		group.DefaultLeverage = *req.DefaultLeverage
	}
	if req.SpreadType != nil {
		group.SpreadType = *req.SpreadType
	}
	if req.CommissionModel != nil {
		group.CommissionModel = *req.CommissionModel
	}
	if req.CommissionRate != nil {
		group.CommissionRate = *req.CommissionRate
	}
	if req.MarginMode != nil {
		group.MarginMode = *req.MarginMode
	}
	if req.TradingHoursTemplate != nil {
		group.TradingHoursTemplate = *req.TradingHoursTemplate
	}
	if req.IsActive != nil {
		group.IsActive = *req.IsActive
	}
	group.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// ============================================
// 5. DELETE /admin/instruments/groups/:id — Delete group
// ============================================

func (h *InstrumentGroupHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/instruments/groups/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}
	groupID := parts[0]

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	group, exists := h.store.groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check if group has symbols
	if group.SymbolCount > 0 {
		http.Error(w, "Cannot delete group with symbols", http.StatusBadRequest)
		return
	}

	delete(h.store.groups, groupID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Group deleted successfully",
	})
}

// ============================================
// 6. GET /admin/instruments/groups/:id/symbols — Symbols in group
// ============================================

func (h *InstrumentGroupHandler) HandleGetGroupSymbols(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract group ID from path
	parts := strings.Split(r.URL.Path, "/")
	var groupID string
	for i, part := range parts {
		if part == "groups" && i+1 < len(parts) {
			groupID = parts[i+1]
			break
		}
	}

	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	if _, exists := h.store.groups[groupID]; !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	symbols := []*SymbolInGroup{}
	for _, symbol := range h.store.symbols {
		if symbol.GroupID == groupID {
			symbols = append(symbols, symbol)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbols": symbols,
		"total":   len(symbols),
	})
}

// ============================================
// 7. POST /admin/instruments/groups/:id/symbols — Add symbol to group
// ============================================

func (h *InstrumentGroupHandler) HandleAddSymbol(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract group ID
	parts := strings.Split(r.URL.Path, "/")
	var groupID string
	for i, part := range parts {
		if part == "groups" && i+1 < len(parts) {
			groupID = parts[i+1]
			break
		}
	}

	if groupID == "" {
		http.Error(w, "Group ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Symbol string `json:"symbol"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	group, exists := h.store.groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Check if symbol already exists
	if _, exists := h.store.symbols[req.Symbol]; exists {
		http.Error(w, "Symbol already exists in another group", http.StatusBadRequest)
		return
	}

	now := time.Now()
	symbol := &SymbolInGroup{
		Symbol:          req.Symbol,
		GroupID:         groupID,
		GroupName:       group.Name,
		Leverage:        group.DefaultLeverage,
		SpreadType:      group.SpreadType,
		CommissionModel: group.CommissionModel,
		CommissionRate:  group.CommissionRate,
		MarginMode:      group.MarginMode,
		TradingHours:    group.TradingHoursTemplate,
		IsActive:        true,
		AddedAt:         now,
	}

	h.store.symbols[req.Symbol] = symbol
	group.SymbolCount++
	group.UpdatedAt = now

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(symbol)
}

// ============================================
// 8. DELETE /admin/instruments/groups/:id/symbols/:symbol — Remove symbol
// ============================================

func (h *InstrumentGroupHandler) HandleRemoveSymbol(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract group ID and symbol
	parts := strings.Split(r.URL.Path, "/")
	var groupID, symbolName string
	for i, part := range parts {
		if part == "groups" && i+1 < len(parts) {
			groupID = parts[i+1]
			if i+3 < len(parts) && parts[i+2] == "symbols" {
				symbolName = parts[i+3]
			}
			break
		}
	}

	if groupID == "" || symbolName == "" {
		http.Error(w, "Group ID and symbol required", http.StatusBadRequest)
		return
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	group, exists := h.store.groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	symbol, exists := h.store.symbols[symbolName]
	if !exists || symbol.GroupID != groupID {
		http.Error(w, "Symbol not found in this group", http.StatusNotFound)
		return
	}

	delete(h.store.symbols, symbolName)
	group.SymbolCount--
	group.UpdatedAt = time.Now()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Symbol removed successfully",
	})
}

// ============================================
// 9. GET /admin/instruments/stats — Instrument stats
// ============================================

func (h *InstrumentGroupHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.store.mu.RLock()
	defer h.store.mu.RUnlock()

	totalGroups := len(h.store.groups)
	totalSymbols := len(h.store.symbols)

	// Find most active group
	var mostActiveGroupID, mostActiveGroupName string
	mostActiveSymbols := 0

	for _, group := range h.store.groups {
		if group.SymbolCount > mostActiveSymbols {
			mostActiveSymbols = group.SymbolCount
			mostActiveGroupID = group.ID
			mostActiveGroupName = group.Name
		}
	}

	stats := InstrumentStats{
		TotalGroups:         totalGroups,
		TotalSymbols:        totalSymbols,
		MostActiveGroupID:   mostActiveGroupID,
		MostActiveGroupName: mostActiveGroupName,
		MostActiveSymbols:   mostActiveSymbols,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
