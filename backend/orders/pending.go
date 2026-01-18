package orders

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OrderType defines the type of order
type OrderType string

const (
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeStop      OrderType = "STOP"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

// OrderSide defines buy or sell
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderSubtype for specific pending order types
type OrderSubtype string

const (
	SubtypeBuyLimit  OrderSubtype = "BUY_LIMIT"
	SubtypeSellLimit OrderSubtype = "SELL_LIMIT"
	SubtypeBuyStop   OrderSubtype = "BUY_STOP"
	SubtypeSellStop  OrderSubtype = "SELL_STOP"
)

// OrderStatus represents the current state of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusTriggered OrderStatus = "TRIGGERED"
	StatusFilled    OrderStatus = "FILLED"
	StatusCancelled OrderStatus = "CANCELLED"
	StatusExpired   OrderStatus = "EXPIRED"
	StatusRejected  OrderStatus = "REJECTED"
)

// PendingOrder represents a pending order in the system
type PendingOrder struct {
	ID           string      `json:"id"`
	Symbol       string      `json:"symbol"`
	Side         OrderSide   `json:"side"`
	Type         OrderType   `json:"type"`
	Subtype      OrderSubtype `json:"subtype"`
	Volume       float64     `json:"volume"`
	EntryPrice   float64     `json:"entryPrice,omitempty"`   // For limit orders
	TriggerPrice float64     `json:"triggerPrice,omitempty"` // For stop orders
	LimitPrice   float64     `json:"limitPrice,omitempty"`   // For stop-limit orders
	SL           float64     `json:"sl,omitempty"`
	TP           float64     `json:"tp,omitempty"`
	OCOPairID    string      `json:"ocoPairId,omitempty"`
	Expiry       *time.Time  `json:"expiry,omitempty"`
	Status       OrderStatus `json:"status"`
	CreatedAt    time.Time   `json:"createdAt"`
	TriggeredAt  *time.Time  `json:"triggeredAt,omitempty"`
	MaxSlippage  float64     `json:"maxSlippage,omitempty"`
}

// TPLadder represents a take-profit level
type TPLadder struct {
	Level        int         `json:"level"` // 1, 2, 3
	Price        float64     `json:"price"`
	ClosePercent float64     `json:"closePercent"` // e.g. 50.0
	Status       OrderStatus `json:"status"`
}

// OrderService manages all order operations
type OrderService struct {
	mu            sync.RWMutex
	pendingOrders map[string]*PendingOrder
	tpLadders     map[string][]TPLadder // tradeId -> TP levels
	priceCallback func(symbol string) (bid, ask float64, ok bool)
	execCallback  func(order *PendingOrder) error
}

// NewOrderService creates a new order service
func NewOrderService() *OrderService {
	svc := &OrderService{
		pendingOrders: make(map[string]*PendingOrder),
		tpLadders:     make(map[string][]TPLadder),
	}
	
	// Start background processor for pending orders
	go svc.processLoop()
	
	log.Println("[OrderService] Initialized with pending order processor")
	return svc
}

// SetPriceCallback sets the function to get current prices
func (s *OrderService) SetPriceCallback(fn func(symbol string) (bid, ask float64, ok bool)) {
	s.priceCallback = fn
}

// SetExecutionCallback sets the function to execute triggered orders
func (s *OrderService) SetExecutionCallback(fn func(order *PendingOrder) error) {
	s.execCallback = fn
}

// PlaceLimitOrder creates a limit order
func (s *OrderService) PlaceLimitOrder(symbol string, side OrderSide, volume, price, sl, tp float64) (*PendingOrder, error) {
	if price <= 0 {
		return nil, errors.New("invalid limit price")
	}

	subtype := SubtypeBuyLimit
	if side == OrderSideSell {
		subtype = SubtypeSellLimit
	}

	order := &PendingOrder{
		ID:         uuid.New().String(),
		Symbol:     symbol,
		Side:       side,
		Type:       OrderTypeLimit,
		Subtype:    subtype,
		Volume:     volume,
		EntryPrice: price,
		SL:         sl,
		TP:         tp,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
	}

	s.mu.Lock()
	s.pendingOrders[order.ID] = order
	s.mu.Unlock()

	log.Printf("[OrderService] Limit order placed: %s %s %.2f lots @ %.5f", side, symbol, volume, price)
	return order, nil
}

// PlaceStopOrder creates a stop order
func (s *OrderService) PlaceStopOrder(symbol string, side OrderSide, volume, triggerPrice, sl, tp float64) (*PendingOrder, error) {
	if triggerPrice <= 0 {
		return nil, errors.New("invalid trigger price")
	}

	subtype := SubtypeBuyStop
	if side == OrderSideSell {
		subtype = SubtypeSellStop
	}

	order := &PendingOrder{
		ID:           uuid.New().String(),
		Symbol:       symbol,
		Side:         side,
		Type:         OrderTypeStop,
		Subtype:      subtype,
		Volume:       volume,
		TriggerPrice: triggerPrice,
		SL:           sl,
		TP:           tp,
		Status:       StatusPending,
		CreatedAt:    time.Now(),
	}

	s.mu.Lock()
	s.pendingOrders[order.ID] = order
	s.mu.Unlock()

	log.Printf("[OrderService] Stop order placed: %s %s %.2f lots @ trigger %.5f", side, symbol, volume, triggerPrice)
	return order, nil
}

// PlaceStopLimitOrder creates a stop-limit order
func (s *OrderService) PlaceStopLimitOrder(symbol string, side OrderSide, volume, triggerPrice, limitPrice, sl, tp float64) (*PendingOrder, error) {
	if triggerPrice <= 0 || limitPrice <= 0 {
		return nil, errors.New("invalid prices")
	}

	order := &PendingOrder{
		ID:           uuid.New().String(),
		Symbol:       symbol,
		Side:         side,
		Type:         OrderTypeStopLimit,
		Subtype:      SubtypeBuyStop, // Will be updated based on logic
		Volume:       volume,
		TriggerPrice: triggerPrice,
		LimitPrice:   limitPrice,
		SL:           sl,
		TP:           tp,
		Status:       StatusPending,
		CreatedAt:    time.Now(),
	}

	if side == OrderSideSell {
		order.Subtype = SubtypeSellStop
	}

	s.mu.Lock()
	s.pendingOrders[order.ID] = order
	s.mu.Unlock()

	log.Printf("[OrderService] Stop-Limit order placed: %s %s %.2f lots @ trigger %.5f limit %.5f",
		side, symbol, volume, triggerPrice, limitPrice)
	return order, nil
}

// PlaceOCO creates a One-Cancels-Other order pair
func (s *OrderService) PlaceOCO(order1, order2 *PendingOrder) error {
	if order1 == nil || order2 == nil {
		return errors.New("both orders required for OCO")
	}

	// Link them together
	order1.OCOPairID = order2.ID
	order2.OCOPairID = order1.ID

	s.mu.Lock()
	s.pendingOrders[order1.ID] = order1
	s.pendingOrders[order2.ID] = order2
	s.mu.Unlock()

	log.Printf("[OrderService] OCO pair created: %s <-> %s", order1.ID, order2.ID)
	return nil
}

// CancelOrder cancels a pending order
func (s *OrderService) CancelOrder(orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.pendingOrders[orderID]
	if !exists {
		return errors.New("order not found")
	}

	order.Status = StatusCancelled
	delete(s.pendingOrders, orderID)

	// Cancel OCO pair if exists
	if order.OCOPairID != "" {
		if pairOrder, ok := s.pendingOrders[order.OCOPairID]; ok {
			pairOrder.Status = StatusCancelled
			delete(s.pendingOrders, order.OCOPairID)
		}
	}

	log.Printf("[OrderService] Order cancelled: %s", orderID)
	return nil
}

// GetPendingOrders returns all pending orders
func (s *OrderService) GetPendingOrders() []*PendingOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*PendingOrder, 0, len(s.pendingOrders))
	for _, order := range s.pendingOrders {
		orders = append(orders, order)
	}
	return orders
}

// GetPendingOrdersBySymbol returns pending orders for a symbol
func (s *OrderService) GetPendingOrdersBySymbol(symbol string) []*PendingOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var orders []*PendingOrder
	for _, order := range s.pendingOrders {
		if order.Symbol == symbol {
			orders = append(orders, order)
		}
	}
	return orders
}

// SetTPLadder sets take-profit levels for a trade
func (s *OrderService) SetTPLadder(tradeID string, levels []TPLadder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tpLadders[tradeID] = levels
	log.Printf("[OrderService] TP Ladder set for trade %s: %d levels", tradeID, len(levels))
}

// GetTPLadder returns TP levels for a trade
func (s *OrderService) GetTPLadder(tradeID string) []TPLadder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tpLadders[tradeID]
}

// processLoop checks pending orders against current prices
func (s *OrderService) processLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
	for range ticker.C {
		s.checkPendingOrders()
		s.checkTPLadders()
	}
}

func (s *OrderService) checkPendingOrders() {
	if s.priceCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, order := range s.pendingOrders {
		if order.Status != StatusPending {
			continue
		}

		// Check expiry
		if order.Expiry != nil && time.Now().After(*order.Expiry) {
			order.Status = StatusExpired
			delete(s.pendingOrders, id)
			log.Printf("[OrderService] Order expired: %s", id)
			continue
		}

		bid, ask, ok := s.priceCallback(order.Symbol)
		if !ok {
			continue
		}

		triggered := false

		switch order.Type {
		case OrderTypeLimit:
			// Buy limit triggers when ask <= entry price
			// Sell limit triggers when bid >= entry price
			if order.Side == OrderSideBuy && ask <= order.EntryPrice {
				triggered = true
			} else if order.Side == OrderSideSell && bid >= order.EntryPrice {
				triggered = true
			}

		case OrderTypeStop:
			// Buy stop triggers when ask >= trigger price
			// Sell stop triggers when bid <= trigger price
			if order.Side == OrderSideBuy && ask >= order.TriggerPrice {
				triggered = true
			} else if order.Side == OrderSideSell && bid <= order.TriggerPrice {
				triggered = true
			}

		case OrderTypeStopLimit:
			// First check trigger, then it becomes a limit order
			if order.Side == OrderSideBuy && ask >= order.TriggerPrice {
				order.Type = OrderTypeLimit
				order.EntryPrice = order.LimitPrice
				log.Printf("[OrderService] Stop-Limit triggered, now Limit @ %.5f", order.LimitPrice)
			} else if order.Side == OrderSideSell && bid <= order.TriggerPrice {
				order.Type = OrderTypeLimit
				order.EntryPrice = order.LimitPrice
				log.Printf("[OrderService] Stop-Limit triggered, now Limit @ %.5f", order.LimitPrice)
			}
		}

		if triggered {
			now := time.Now()
			order.TriggeredAt = &now
			order.Status = StatusTriggered

			// Execute the order
			if s.execCallback != nil {
				go func(o *PendingOrder) {
					if err := s.execCallback(o); err != nil {
						log.Printf("[OrderService] Execution failed: %v", err)
						o.Status = StatusRejected
					} else {
						o.Status = StatusFilled
					}
				}(order)
			}

			// Cancel OCO pair
			if order.OCOPairID != "" {
				if pairOrder, exists := s.pendingOrders[order.OCOPairID]; exists {
					pairOrder.Status = StatusCancelled
					delete(s.pendingOrders, order.OCOPairID)
					log.Printf("[OrderService] OCO pair cancelled: %s", order.OCOPairID)
				}
			}

			delete(s.pendingOrders, id)
			log.Printf("[OrderService] Order triggered: %s %s %s @ %.5f",
				order.Side, order.Symbol, order.Type, bid)
		}
	}
}

func (s *OrderService) checkTPLadders() {
	// TP Ladder processing will be implemented with position management
}
