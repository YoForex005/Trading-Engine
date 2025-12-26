package oms

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Order represents a trading order
type Order struct {
	ID            string    `json:"id"`
	AccountID     string    `json:"accountId"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"` // BUY or SELL
	Type          string    `json:"type"` // MARKET, LIMIT, STOP
	Volume        float64   `json:"volume"`
	PriceRequest  float64   `json:"priceRequest"`
	PriceExecuted float64   `json:"priceExecuted"`
	SL            float64   `json:"sl"`
	TP            float64   `json:"tp"`
	Status        string    `json:"status"` // PENDING, FILLED, REJECTED, CANCELED
	RoutingType   string    `json:"routingType"` // A_BOOK, B_BOOK
	CreatedAt     time.Time `json:"createdAt"`
	FilledAt      *time.Time `json:"filledAt,omitempty"`
}

// Position represents an open position
type Position struct {
	ID           string    `json:"id"`
	OrderID      string    `json:"orderId"`
	AccountID    string    `json:"accountId"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"`
	Volume       float64   `json:"volume"`
	OpenPrice    float64   `json:"openPrice"`
	CurrentPrice float64   `json:"currentPrice"`
	SL           float64   `json:"sl"`
	TP           float64   `json:"tp"`
	Swap         float64   `json:"swap"`
	Profit       float64   `json:"profit"`
	OpenTime     time.Time `json:"openTime"`
}

// Service handles order management
type Service struct {
	orders    map[string]*Order
	positions map[string]*Position
	mu        sync.RWMutex
}

func NewService() *Service {
	return &Service{
		orders:    make(map[string]*Order),
		positions: make(map[string]*Position),
	}
}

// PlaceOrder creates a new order
func (s *Service) PlaceOrder(req PlaceOrderRequest) (*Order, error) {
	// Validation
	if req.Volume <= 0 || req.Volume > 100 {
		return nil, errors.New("invalid volume")
	}
	if req.Side != "BUY" && req.Side != "SELL" {
		return nil, errors.New("invalid side")
	}

	order := &Order{
		ID:           uuid.New().String(),
		AccountID:    req.AccountID,
		Symbol:       req.Symbol,
		Side:         req.Side,
		Type:         req.Type,
		Volume:       req.Volume,
		PriceRequest: req.Price,
		SL:           req.SL,
		TP:           req.TP,
		Status:       "PENDING",
		RoutingType:  s.determineRouting(req),
		CreatedAt:    time.Now(),
	}

	// For MARKET orders, execute immediately (B-Book simulation)
	if order.Type == "MARKET" {
		order.Status = "FILLED"
		order.PriceExecuted = req.Price
		now := time.Now()
		order.FilledAt = &now

		// Create position
		position := &Position{
			ID:           uuid.New().String(),
			OrderID:      order.ID,
			AccountID:    order.AccountID,
			Symbol:       order.Symbol,
			Side:         order.Side,
			Volume:       order.Volume,
			OpenPrice:    order.PriceExecuted,
			CurrentPrice: order.PriceExecuted,
			SL:           order.SL,
			TP:           order.TP,
			OpenTime:     now,
		}

		s.mu.Lock()
		s.orders[order.ID] = order
		s.positions[position.ID] = position
		s.mu.Unlock()

		return order, nil
	}

	s.mu.Lock()
	s.orders[order.ID] = order
	s.mu.Unlock()

	return order, nil
}

// determineRouting decides A-Book vs B-Book
func (s *Service) determineRouting(req PlaceOrderRequest) string {
	// Simple rule: Large orders go to A-Book
	if req.Volume >= 10.0 {
		return "A_BOOK"
	}
	// Default to B-Book
	return "B_BOOK"
}

// GetPositions returns all open positions for an account
func (s *Service) GetPositions(accountID string) []*Position {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Position
	for _, pos := range s.positions {
		if pos.AccountID == accountID {
			result = append(result, pos)
		}
	}
	return result
}

// ClosePosition closes an open position
func (s *Service) ClosePosition(positionID string, closePrice float64) (*Position, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos, ok := s.positions[positionID]
	if !ok {
		return nil, errors.New("position not found")
	}

	// Calculate profit
	if pos.Side == "BUY" {
		pos.Profit = (closePrice - pos.OpenPrice) * pos.Volume * 100000
	} else {
		pos.Profit = (pos.OpenPrice - closePrice) * pos.Volume * 100000
	}

	// Remove from active positions
	delete(s.positions, positionID)

	return pos, nil
}

// PlaceOrderRequest is the input for placing an order
type PlaceOrderRequest struct {
	AccountID string  `json:"accountId"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Volume    float64 `json:"volume"`
	Price     float64 `json:"price"`
	SL        float64 `json:"sl"`
	TP        float64 `json:"tp"`
}
