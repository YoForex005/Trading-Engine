package admin

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// CopyMode defines how to copy trades
type CopyMode string

const (
	CopyModeProportional CopyMode = "proportional" // follower lot = master lot * multiplier
	CopyModeFixed        CopyMode = "fixed"        // follower lot = fixedLotSize
	CopyModeInverse      CopyMode = "inverse"      // opposite direction with multiplier
)

// CopyRelation represents a master-follower relationship
type CopyRelation struct {
	ID                int64      `json:"id"`
	MasterAccountID   int64      `json:"masterAccountId"`
	FollowerAccountID int64      `json:"followerAccountId"`
	CopyMode          CopyMode   `json:"copyMode"`
	FixedLotSize      float64    `json:"fixedLotSize,omitempty"`      // For fixed mode
	LotMultiplier     float64    `json:"lotMultiplier,omitempty"`     // For proportional mode
	MaxLots           float64    `json:"maxLots"`                     // Maximum lots per trade
	InverseCopy       bool       `json:"inverseCopy"`                 // Copy in opposite direction
	SkipSymbols       []string   `json:"skipSymbols,omitempty"`       // Symbols to not copy
	Status            string     `json:"status"`                      // active, paused, stopped
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	TotalTradesCopied int        `json:"totalTradesCopied"`
	LastCopyAt        *time.Time `json:"lastCopyAt,omitempty"`
}

// CopyLog tracks individual copy operations
type CopyLog struct {
	ID              int64      `json:"id"`
	RelationID      int64      `json:"relationId"`
	MasterTradeID   int64      `json:"masterTradeId"`
	FollowerTradeID int64      `json:"followerTradeId,omitempty"`
	Action          string     `json:"action"` // open, close, modify
	Status          string     `json:"status"` // success, failed
	ErrorMsg        string     `json:"errorMsg,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
	Details         string     `json:"details,omitempty"`
}

// TradeEvent represents a trade event from the master account
type TradeEvent struct {
	Action     string  // "open", "close", "modify_sl_tp"
	PositionID int64
	Symbol     string
	Side       string  // BUY/SELL
	Volume     float64
	Price      float64
	SL         float64
	TP         float64
}

// TradeCopierService manages trade copying relationships
type TradeCopierService struct {
	mu                sync.RWMutex
	engine            *core.Engine
	relations         map[int64]*CopyRelation
	logs              map[int64]*CopyLog
	positionMap       map[int64]int64 // masterPositionID -> followerPositionID
	nextRelationID    int64
	nextLogID         int64
	maxCopyRelations  int
	copyDelayMs       int
}

// NewTradeCopierService creates a new trade copier service
func NewTradeCopierService(engine *core.Engine) *TradeCopierService {
	// Get configuration from env
	maxRelations := 100
	if envMax := os.Getenv("MAX_COPY_RELATIONS"); envMax != "" {
		if parsed, err := strconv.Atoi(envMax); err == nil && parsed > 0 {
			maxRelations = parsed
		}
	}

	delayMs := 100 // Default 100ms delay between copy operations
	if envDelay := os.Getenv("COPY_DELAY_MS"); envDelay != "" {
		if parsed, err := strconv.Atoi(envDelay); err == nil && parsed >= 0 {
			delayMs = parsed
		}
	}

	svc := &TradeCopierService{
		engine:           engine,
		relations:        make(map[int64]*CopyRelation),
		logs:             make(map[int64]*CopyLog),
		positionMap:      make(map[int64]int64),
		nextRelationID:   1,
		nextLogID:        1,
		maxCopyRelations: maxRelations,
		copyDelayMs:      delayMs,
	}

	log.Printf("[TradeCopier] Service initialized (max relations: %d, copy delay: %dms)",
		maxRelations, delayMs)
	return svc
}

// CreateRelation creates a new copy relation
func (tcs *TradeCopierService) CreateRelation(masterAccountID, followerAccountID int64,
	copyMode CopyMode, fixedLotSize, lotMultiplier, maxLots float64,
	inverseCopy bool, skipSymbols []string) (*CopyRelation, error) {

	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	// Check max relations limit
	if len(tcs.relations) >= tcs.maxCopyRelations {
		return nil, fmt.Errorf("maximum copy relations limit reached (%d)", tcs.maxCopyRelations)
	}

	// Validate accounts exist
	if _, ok := tcs.engine.GetAccount(masterAccountID); !ok {
		return nil, errors.New("master account not found")
	}
	if _, ok := tcs.engine.GetAccount(followerAccountID); !ok {
		return nil, errors.New("follower account not found")
	}

	// Prevent self-copy
	if masterAccountID == followerAccountID {
		return nil, errors.New("cannot copy to the same account")
	}

	// Check for duplicate relation
	for _, rel := range tcs.relations {
		if rel.MasterAccountID == masterAccountID && rel.FollowerAccountID == followerAccountID {
			return nil, errors.New("copy relation already exists")
		}
	}

	// Validate copy mode parameters
	if copyMode == CopyModeFixed && fixedLotSize <= 0 {
		return nil, errors.New("fixedLotSize must be greater than 0 for fixed mode")
	}
	if copyMode == CopyModeProportional && lotMultiplier <= 0 {
		return nil, errors.New("lotMultiplier must be greater than 0 for proportional mode")
	}

	// Default max lots
	if maxLots <= 0 {
		maxLots = 100.0
	}

	relation := &CopyRelation{
		ID:                tcs.nextRelationID,
		MasterAccountID:   masterAccountID,
		FollowerAccountID: followerAccountID,
		CopyMode:          copyMode,
		FixedLotSize:      fixedLotSize,
		LotMultiplier:     lotMultiplier,
		MaxLots:           maxLots,
		InverseCopy:       inverseCopy,
		SkipSymbols:       skipSymbols,
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		TotalTradesCopied: 0,
	}

	tcs.nextRelationID++
	tcs.relations[relation.ID] = relation

	log.Printf("[TradeCopier] Relation created: ID=%d, Master=%d, Follower=%d, Mode=%s",
		relation.ID, masterAccountID, followerAccountID, copyMode)

	return relation, nil
}

// UpdateRelation updates an existing relation
func (tcs *TradeCopierService) UpdateRelation(id int64, copyMode *CopyMode, fixedLotSize,
	lotMultiplier, maxLots *float64, inverseCopy *bool, skipSymbols []string) error {

	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	relation, exists := tcs.relations[id]
	if !exists {
		return errors.New("copy relation not found")
	}

	if copyMode != nil {
		relation.CopyMode = *copyMode
	}
	if fixedLotSize != nil && *fixedLotSize > 0 {
		relation.FixedLotSize = *fixedLotSize
	}
	if lotMultiplier != nil && *lotMultiplier > 0 {
		relation.LotMultiplier = *lotMultiplier
	}
	if maxLots != nil && *maxLots > 0 {
		relation.MaxLots = *maxLots
	}
	if inverseCopy != nil {
		relation.InverseCopy = *inverseCopy
	}
	if skipSymbols != nil {
		relation.SkipSymbols = skipSymbols
	}

	relation.UpdatedAt = time.Now()

	log.Printf("[TradeCopier] Relation updated: ID=%d", id)
	return nil
}

// DeleteRelation deletes a copy relation
func (tcs *TradeCopierService) DeleteRelation(id int64) error {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	if _, exists := tcs.relations[id]; !exists {
		return errors.New("copy relation not found")
	}

	delete(tcs.relations, id)

	log.Printf("[TradeCopier] Relation deleted: ID=%d", id)
	return nil
}

// PauseRelation pauses a copy relation
func (tcs *TradeCopierService) PauseRelation(id int64) error {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	relation, exists := tcs.relations[id]
	if !exists {
		return errors.New("copy relation not found")
	}

	relation.Status = "paused"
	relation.UpdatedAt = time.Now()

	log.Printf("[TradeCopier] Relation paused: ID=%d", id)
	return nil
}

// ResumeRelation resumes a paused copy relation
func (tcs *TradeCopierService) ResumeRelation(id int64) error {
	tcs.mu.Lock()
	defer tcs.mu.Unlock()

	relation, exists := tcs.relations[id]
	if !exists {
		return errors.New("copy relation not found")
	}

	relation.Status = "active"
	relation.UpdatedAt = time.Now()

	log.Printf("[TradeCopier] Relation resumed: ID=%d", id)
	return nil
}

// GetRelation retrieves a copy relation
func (tcs *TradeCopierService) GetRelation(id int64) (*CopyRelation, error) {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	relation, exists := tcs.relations[id]
	if !exists {
		return nil, errors.New("copy relation not found")
	}

	return relation, nil
}

// ListRelations returns all copy relations
func (tcs *TradeCopierService) ListRelations() []*CopyRelation {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	relations := make([]*CopyRelation, 0, len(tcs.relations))
	for _, relation := range tcs.relations {
		relations = append(relations, relation)
	}

	return relations
}

// GetRelationLogs returns copy logs for a relation
func (tcs *TradeCopierService) GetRelationLogs(relationID int64, limit int) []*CopyLog {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	logs := make([]*CopyLog, 0)
	for _, copyLog := range tcs.logs {
		if copyLog.RelationID == relationID {
			logs = append(logs, copyLog)
			if limit > 0 && len(logs) >= limit {
				break
			}
		}
	}

	return logs
}

// GetStats returns trade copier statistics
func (tcs *TradeCopierService) GetStats() map[string]interface{} {
	tcs.mu.RLock()
	defer tcs.mu.RUnlock()

	activeRelations := 0
	pausedRelations := 0
	today := time.Now().Truncate(24 * time.Hour)
	tradesToday := 0

	for _, relation := range tcs.relations {
		if relation.Status == "active" {
			activeRelations++
		} else if relation.Status == "paused" {
			pausedRelations++
		}
	}

	for _, copyLog := range tcs.logs {
		if copyLog.Timestamp.After(today) && copyLog.Action == "open" && copyLog.Status == "success" {
			tradesToday++
		}
	}

	return map[string]interface{}{
		"totalRelations":  len(tcs.relations),
		"activeRelations": activeRelations,
		"pausedRelations": pausedRelations,
		"tradesToday":     tradesToday,
		"maxRelations":    tcs.maxCopyRelations,
	}
}

// OnTradeEvent handles trade events from master accounts
func (tcs *TradeCopierService) OnTradeEvent(masterAccountID int64, event TradeEvent) {
	tcs.mu.RLock()
	// Find all active relations for this master
	activeRelations := make([]*CopyRelation, 0)
	for _, relation := range tcs.relations {
		if relation.MasterAccountID == masterAccountID && relation.Status == "active" {
			activeRelations = append(activeRelations, relation)
		}
	}
	tcs.mu.RUnlock()

	if len(activeRelations) == 0 {
		return
	}

	// Process each relation
	for _, relation := range activeRelations {
		// Apply copy delay if configured
		if tcs.copyDelayMs > 0 {
			time.Sleep(time.Duration(tcs.copyDelayMs) * time.Millisecond)
		}

		// Check if symbol is skipped
		if tcs.isSymbolSkipped(event.Symbol, relation.SkipSymbols) {
			log.Printf("[TradeCopier] Symbol %s skipped for relation %d", event.Symbol, relation.ID)
			continue
		}

		switch event.Action {
		case "open":
			tcs.copyOpenPosition(relation, event)
		case "close":
			tcs.copyClosePosition(relation, event)
		case "modify_sl_tp":
			tcs.copyModifySLTP(relation, event)
		}
	}
}

// copyOpenPosition copies position opening to follower
func (tcs *TradeCopierService) copyOpenPosition(relation *CopyRelation, event TradeEvent) {
	// Calculate follower volume
	followerVolume := tcs.calculateFollowerVolume(relation, event.Volume)

	// Cap at max lots
	if followerVolume > relation.MaxLots {
		followerVolume = relation.MaxLots
	}

	// Determine side (inverse if enabled)
	followerSide := event.Side
	if relation.InverseCopy {
		if event.Side == "BUY" {
			followerSide = "SELL"
		} else {
			followerSide = "BUY"
		}
	}

	// Execute market order on follower account
	position, err := tcs.engine.ExecuteMarketOrder(
		relation.FollowerAccountID,
		event.Symbol,
		followerSide,
		followerVolume,
		event.SL,
		event.TP,
	)

	// Log the copy operation
	copyLog := &CopyLog{
		ID:            tcs.nextLogID,
		RelationID:    relation.ID,
		MasterTradeID: event.PositionID,
		Action:        "open",
		Timestamp:     time.Now(),
	}
	tcs.nextLogID++

	if err != nil {
		copyLog.Status = "failed"
		copyLog.ErrorMsg = err.Error()
		log.Printf("[TradeCopier] Failed to copy open: Relation=%d, Error=%v", relation.ID, err)
	} else {
		copyLog.Status = "success"
		copyLog.FollowerTradeID = position.ID
		copyLog.Details = fmt.Sprintf("Opened %s %s %.2f lots at %.5f",
			followerSide, event.Symbol, followerVolume, position.OpenPrice)

		// Map master position to follower position
		tcs.mu.Lock()
		tcs.positionMap[event.PositionID] = position.ID
		relation.TotalTradesCopied++
		now := time.Now()
		relation.LastCopyAt = &now
		tcs.mu.Unlock()

		log.Printf("[TradeCopier] Position copied: Relation=%d, Master=%d, Follower=%d",
			relation.ID, event.PositionID, position.ID)
	}

	tcs.mu.Lock()
	tcs.logs[copyLog.ID] = copyLog
	tcs.mu.Unlock()
}

// copyClosePosition copies position closing to follower
func (tcs *TradeCopierService) copyClosePosition(relation *CopyRelation, event TradeEvent) {
	// Find follower position
	tcs.mu.RLock()
	followerPositionID, exists := tcs.positionMap[event.PositionID]
	tcs.mu.RUnlock()

	copyLog := &CopyLog{
		ID:            tcs.nextLogID,
		RelationID:    relation.ID,
		MasterTradeID: event.PositionID,
		Action:        "close",
		Timestamp:     time.Now(),
	}
	tcs.nextLogID++

	if !exists {
		copyLog.Status = "failed"
		copyLog.ErrorMsg = "Follower position not found in map"
		tcs.mu.Lock()
		tcs.logs[copyLog.ID] = copyLog
		tcs.mu.Unlock()
		return
	}

	copyLog.FollowerTradeID = followerPositionID

	// Close follower position
	_, err := tcs.engine.ClosePosition(followerPositionID, event.Volume)

	if err != nil {
		copyLog.Status = "failed"
		copyLog.ErrorMsg = err.Error()
		log.Printf("[TradeCopier] Failed to copy close: Relation=%d, Error=%v", relation.ID, err)
	} else {
		copyLog.Status = "success"
		copyLog.Details = fmt.Sprintf("Closed position %d (%.2f lots)", followerPositionID, event.Volume)

		// Remove from position map
		tcs.mu.Lock()
		delete(tcs.positionMap, event.PositionID)
		now := time.Now()
		relation.LastCopyAt = &now
		tcs.mu.Unlock()

		log.Printf("[TradeCopier] Position close copied: Relation=%d, Follower=%d",
			relation.ID, followerPositionID)
	}

	tcs.mu.Lock()
	tcs.logs[copyLog.ID] = copyLog
	tcs.mu.Unlock()
}

// copyModifySLTP copies SL/TP modification to follower
func (tcs *TradeCopierService) copyModifySLTP(relation *CopyRelation, event TradeEvent) {
	// Find follower position
	tcs.mu.RLock()
	followerPositionID, exists := tcs.positionMap[event.PositionID]
	tcs.mu.RUnlock()

	copyLog := &CopyLog{
		ID:            tcs.nextLogID,
		RelationID:    relation.ID,
		MasterTradeID: event.PositionID,
		Action:        "modify",
		Timestamp:     time.Now(),
	}
	tcs.nextLogID++

	if !exists {
		copyLog.Status = "failed"
		copyLog.ErrorMsg = "Follower position not found in map"
		tcs.mu.Lock()
		tcs.logs[copyLog.ID] = copyLog
		tcs.mu.Unlock()
		return
	}

	copyLog.FollowerTradeID = followerPositionID

	// Modify follower position SL/TP
	_, err := tcs.engine.ModifyPosition(followerPositionID, event.SL, event.TP)

	if err != nil {
		copyLog.Status = "failed"
		copyLog.ErrorMsg = err.Error()
		log.Printf("[TradeCopier] Failed to copy modify: Relation=%d, Error=%v", relation.ID, err)
	} else {
		copyLog.Status = "success"
		copyLog.Details = fmt.Sprintf("Modified SL=%.5f, TP=%.5f", event.SL, event.TP)

		tcs.mu.Lock()
		now := time.Now()
		relation.LastCopyAt = &now
		tcs.mu.Unlock()

		log.Printf("[TradeCopier] Position modify copied: Relation=%d, Follower=%d",
			relation.ID, followerPositionID)
	}

	tcs.mu.Lock()
	tcs.logs[copyLog.ID] = copyLog
	tcs.mu.Unlock()
}

// Helper methods

func (tcs *TradeCopierService) calculateFollowerVolume(relation *CopyRelation, masterVolume float64) float64 {
	switch relation.CopyMode {
	case CopyModeFixed:
		return relation.FixedLotSize
	case CopyModeProportional, CopyModeInverse:
		return masterVolume * relation.LotMultiplier
	default:
		return masterVolume
	}
}

func (tcs *TradeCopierService) isSymbolSkipped(symbol string, skipSymbols []string) bool {
	for _, skip := range skipSymbols {
		if skip == symbol {
			return true
		}
	}
	return false
}
