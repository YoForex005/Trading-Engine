package admin

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/core"
)

// OrderManagementService handles admin order operations
type OrderManagementService struct {
	mu            sync.RWMutex
	engine        *core.Engine
	auditLog      *AuditLog
	modifications map[int64]*OrderModification
	nextModID     int64
}

// NewOrderManagementService creates a new order management service
func NewOrderManagementService(engine *core.Engine, auditLog *AuditLog) *OrderManagementService {
	return &OrderManagementService{
		engine:        engine,
		auditLog:      auditLog,
		modifications: make(map[int64]*OrderModification),
		nextModID:     1,
	}
}

// GetAllOrders returns all client orders in real-time
func (s *OrderManagementService) GetAllOrders(status string) ([]*core.Order, error) {
	var allOrders []*core.Order

	// Get orders from all accounts (1-1000)
	for i := int64(1); i <= 1000; i++ {
		orders := s.engine.GetOrders(i, status)
		allOrders = append(allOrders, orders...)
	}

	return allOrders, nil
}

// GetAllPositions returns all open positions
func (s *OrderManagementService) GetAllPositions() ([]*core.Position, error) {
	positions := s.engine.GetAllPositions()
	return positions, nil
}

// ModifyOrder modifies an order's price, volume, SL, or TP
func (s *OrderManagementService) ModifyOrder(orderID int64, price *float64, volume *float64, sl *float64, tp *float64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the order (iterate through accounts)
	var order *core.Order
	var accountID int64

	for i := int64(1); i <= 1000; i++ {
		orders := s.engine.GetOrders(i, "")
		for _, o := range orders {
			if o.ID == orderID {
				order = o
				accountID = i
				break
			}
		}
		if order != nil {
			break
		}
	}

	if order == nil {
		return errors.New("order not found")
	}

	if order.Status != "PENDING" && order.Status != "OPEN" {
		return fmt.Errorf("cannot modify order with status %s", order.Status)
	}

	changes := make(map[string]interface{})
	oldValues := make(map[string]interface{})

	// Modify price
	if price != nil && *price != order.Price {
		oldValues["price"] = order.Price
		changes["price"] = *price
		order.Price = *price
	}

	// Modify volume
	if volume != nil && *volume != order.Volume {
		oldValues["volume"] = order.Volume
		changes["volume"] = *volume
		order.Volume = *volume
	}

	// Modify SL
	if sl != nil && *sl != order.SL {
		oldValues["sl"] = order.SL
		changes["sl"] = *sl
		order.SL = *sl
	}

	// Modify TP
	if tp != nil && *tp != order.TP {
		oldValues["tp"] = order.TP
		changes["tp"] = *tp
		order.TP = *tp
	}

	// Create modification record
	modification := &OrderModification{
		ID:      s.nextModID,
		OrderID: orderID,
		Action:  "MODIFY",
		Changes: map[string]interface{}{
			"old": oldValues,
			"new": changes,
		},
		Reason:    reason,
		AdminID:   admin.ID,
		AdminName: admin.Username,
		CreatedAt: time.Now(),
	}
	s.nextModID++
	s.modifications[modification.ID] = modification

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "ORDER_MODIFY", "ORDER", orderID, map[string]interface{}{
		"accountID": accountID,
		"old":       oldValues,
		"new":       changes,
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[OrderMgmt] Order #%d modified by %s: %v", orderID, admin.Username, changes)

	return nil
}

// ModifyPosition modifies a position's SL/TP or volume
func (s *OrderManagementService) ModifyPosition(positionID int64, sl *float64, tp *float64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the position
	var position *core.Position
	allPositions := s.engine.GetAllPositions()
	for _, p := range allPositions {
		if p.ID == positionID {
			position = p
			break
		}
	}

	if position == nil {
		return errors.New("position not found")
	}

	if position.Status != "OPEN" {
		return errors.New("position is not open")
	}

	changes := make(map[string]interface{})
	oldValues := make(map[string]interface{})

	// Modify SL/TP via engine
	newSL := position.SL
	newTP := position.TP

	if sl != nil {
		oldValues["sl"] = position.SL
		changes["sl"] = *sl
		newSL = *sl
	}

	if tp != nil {
		oldValues["tp"] = position.TP
		changes["tp"] = *tp
		newTP = *tp
	}

	if _, err := s.engine.ModifyPosition(positionID, newSL, newTP); err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "POSITION_MODIFY", "POSITION", positionID, map[string]interface{}{
			"old":    oldValues,
			"new":    changes,
			"reason": reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return fmt.Errorf("failed to modify position: %w", err)
	}

	// Create modification record
	modification := &OrderModification{
		ID:         s.nextModID,
		PositionID: positionID,
		Action:     "MODIFY",
		Changes: map[string]interface{}{
			"old": oldValues,
			"new": changes,
		},
		Reason:    reason,
		AdminID:   admin.ID,
		AdminName: admin.Username,
		CreatedAt: time.Now(),
	}
	s.nextModID++
	s.modifications[modification.ID] = modification

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "POSITION_MODIFY", "POSITION", positionID, map[string]interface{}{
		"accountID": position.AccountID,
		"old":       oldValues,
		"new":       changes,
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[OrderMgmt] Position #%d modified by %s: %v", positionID, admin.Username, changes)

	return nil
}

// ReversePosition reverses a position (BUY → SELL, SELL → BUY)
func (s *OrderManagementService) ReversePosition(positionID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the position
	var position *core.Position
	allPositions := s.engine.GetAllPositions()
	for _, p := range allPositions {
		if p.ID == positionID {
			position = p
			break
		}
	}

	if position == nil {
		return errors.New("position not found")
	}

	if position.Status != "OPEN" {
		return errors.New("position is not open")
	}

	// Close current position
	closeTrade, err := s.engine.ClosePosition(positionID, position.Volume)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "POSITION_REVERSE", "POSITION", positionID, map[string]interface{}{
			"reason": reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return fmt.Errorf("failed to close position: %w", err)
	}

	// Open opposite position
	reverseSide := "BUY"
	if position.Side == "BUY" {
		reverseSide = "SELL"
	}

	newPosition, err := s.engine.ExecuteMarketOrder(position.AccountID, position.Symbol, reverseSide, position.Volume, position.SL, position.TP)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "POSITION_REVERSE", "POSITION", positionID, map[string]interface{}{
			"reason": reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return fmt.Errorf("failed to open reverse position: %w", err)
	}

	// Create modification record
	modification := &OrderModification{
		ID:         s.nextModID,
		PositionID: positionID,
		Action:     "REVERSE",
		Changes: map[string]interface{}{
			"oldSide":       position.Side,
			"newSide":       reverseSide,
			"closedPnL":     closeTrade.RealizedPnL,
			"newPositionID": newPosition.ID,
		},
		Reason:    reason,
		AdminID:   admin.ID,
		AdminName: admin.Username,
		CreatedAt: time.Now(),
	}
	s.nextModID++
	s.modifications[modification.ID] = modification

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "POSITION_REVERSE", "POSITION", positionID, map[string]interface{}{
		"accountID":     position.AccountID,
		"oldSide":       position.Side,
		"newSide":       reverseSide,
		"closedPnL":     closeTrade.RealizedPnL,
		"newPositionID": newPosition.ID,
		"reason":        reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[OrderMgmt] Position #%d reversed by %s: %s → %s, PnL: %.2f",
		positionID, admin.Username, position.Side, reverseSide, closeTrade.RealizedPnL)

	return nil
}

// ClosePosition manually closes a position
func (s *OrderManagementService) ClosePosition(positionID int64, volume float64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get the position
	var position *core.Position
	allPositions := s.engine.GetAllPositions()
	for _, p := range allPositions {
		if p.ID == positionID {
			position = p
			break
		}
	}

	if position == nil {
		return errors.New("position not found")
	}

	if position.Status != "OPEN" {
		return errors.New("position is not open")
	}

	// Close position
	trade, err := s.engine.ClosePosition(positionID, volume)
	if err != nil {
		s.auditLog.Log(admin.ID, admin.Username, "POSITION_CLOSE", "POSITION", positionID, map[string]interface{}{
			"volume": volume,
			"reason": reason,
		}, reason, ipAddress, "", "FAILED", err.Error())
		return fmt.Errorf("failed to close position: %w", err)
	}

	// Create modification record
	modification := &OrderModification{
		ID:         s.nextModID,
		PositionID: positionID,
		Action:     "CLOSE",
		Changes: map[string]interface{}{
			"volume":      volume,
			"realizedPnL": trade.RealizedPnL,
			"closePrice":  trade.Price,
		},
		Reason:    reason,
		AdminID:   admin.ID,
		AdminName: admin.Username,
		CreatedAt: time.Now(),
	}
	s.nextModID++
	s.modifications[modification.ID] = modification

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "POSITION_CLOSE", "POSITION", positionID, map[string]interface{}{
		"accountID":   position.AccountID,
		"volume":      volume,
		"realizedPnL": trade.RealizedPnL,
		"closePrice":  trade.Price,
		"reason":      reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[OrderMgmt] Position #%d closed by %s: %.2f lots, PnL: %.2f",
		positionID, admin.Username, volume, trade.RealizedPnL)

	return nil
}

// DeleteOrder cancels/deletes a pending order
func (s *OrderManagementService) DeleteOrder(orderID int64, admin *Admin, reason string, ipAddress string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the order
	var order *core.Order
	var accountID int64

	for i := int64(1); i <= 1000; i++ {
		orders := s.engine.GetOrders(i, "")
		for _, o := range orders {
			if o.ID == orderID {
				order = o
				accountID = i
				break
			}
		}
		if order != nil {
			break
		}
	}

	if order == nil {
		return errors.New("order not found")
	}

	if order.Status != "PENDING" {
		return fmt.Errorf("cannot delete order with status %s", order.Status)
	}

	// Mark order as cancelled
	order.Status = "CANCELLED"
	order.RejectReason = fmt.Sprintf("Cancelled by admin %s: %s", admin.Username, reason)

	// Create modification record
	modification := &OrderModification{
		ID:      s.nextModID,
		OrderID: orderID,
		Action:  "DELETE",
		Changes: map[string]interface{}{
			"oldStatus": "PENDING",
			"newStatus": "CANCELLED",
			"reason":    reason,
		},
		Reason:    reason,
		AdminID:   admin.ID,
		AdminName: admin.Username,
		CreatedAt: time.Now(),
	}
	s.nextModID++
	s.modifications[modification.ID] = modification

	// Log audit
	s.auditLog.Log(admin.ID, admin.Username, "ORDER_DELETE", "ORDER", orderID, map[string]interface{}{
		"accountID": accountID,
		"reason":    reason,
	}, reason, ipAddress, "", "SUCCESS", "")

	log.Printf("[OrderMgmt] Order #%d deleted by %s: %s", orderID, admin.Username, reason)

	return nil
}

// GetModifications returns order modifications with optional filters
func (s *OrderManagementService) GetModifications(orderID *int64, positionID *int64, limit int) []*OrderModification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var mods []*OrderModification
	for _, mod := range s.modifications {
		if orderID != nil && mod.OrderID != *orderID {
			continue
		}
		if positionID != nil && mod.PositionID != *positionID {
			continue
		}
		mods = append(mods, mod)
	}

	// Sort by created_at descending
	for i := 0; i < len(mods)-1; i++ {
		for j := i + 1; j < len(mods); j++ {
			if mods[i].CreatedAt.Before(mods[j].CreatedAt) {
				mods[i], mods[j] = mods[j], mods[i]
			}
		}
	}

	if limit > 0 && limit < len(mods) {
		return mods[:limit]
	}

	return mods
}
