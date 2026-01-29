package features

import (
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Advanced Order Types
// - Trailing Stop (already exists)
// - OCO (One-Cancels-Other) (already exists)
// - Bracket Orders (Entry + SL + TP together)
// - Iceberg Orders (Hidden Volume)
// - TWAP (Time-Weighted Average Price)
// - VWAP (Volume-Weighted Average Price)
// - Fill-or-Kill (FOK)
// - Immediate-or-Cancel (IOC)
// - Good-till-Date (GTD)
// - Good-till-Cancelled (GTC)

// TimeInForce defines order time validity
type TimeInForce string

const (
	TIF_GTC TimeInForce = "GTC" // Good Till Cancelled
	TIF_GTD TimeInForce = "GTD" // Good Till Date
	TIF_FOK TimeInForce = "FOK" // Fill or Kill
	TIF_IOC TimeInForce = "IOC" // Immediate or Cancel
	TIF_DAY TimeInForce = "DAY" // Day order
)

// ExecutionAlgorithm for algo orders
type ExecutionAlgorithm string

const (
	ALGO_TWAP    ExecutionAlgorithm = "TWAP"    // Time-Weighted Average Price
	ALGO_VWAP    ExecutionAlgorithm = "VWAP"    // Volume-Weighted Average Price
	ALGO_ICEBERG ExecutionAlgorithm = "ICEBERG" // Hidden volume
	ALGO_POV     ExecutionAlgorithm = "POV"     // Percent of Volume
)

// BracketOrder represents entry with automatic SL and TP
type BracketOrder struct {
	ID           string      `json:"id"`
	Symbol       string      `json:"symbol"`
	Side         string      `json:"side"` // BUY or SELL
	Volume       float64     `json:"volume"`
	EntryPrice   float64     `json:"entryPrice"`
	EntryType    string      `json:"entryType"` // MARKET, LIMIT, STOP
	StopLoss     float64     `json:"stopLoss"`
	TakeProfit   float64     `json:"takeProfit"`
	Status       string      `json:"status"` // PENDING, ACTIVE, FILLED, CANCELLED
	EntryOrderID string      `json:"entryOrderId,omitempty"`
	SLOrderID    string      `json:"slOrderId,omitempty"`
	TPOrderID    string      `json:"tpOrderId,omitempty"`
	CreatedAt    time.Time   `json:"createdAt"`
	TimeInForce  TimeInForce `json:"timeInForce"`
	ExpiryTime   *time.Time  `json:"expiryTime,omitempty"`
}

// IcebergOrder hides total volume, shows only visible portion
type IcebergOrder struct {
	ID            string      `json:"id"`
	Symbol        string      `json:"symbol"`
	Side          string      `json:"side"`
	TotalVolume   float64     `json:"totalVolume"`
	VisibleVolume float64     `json:"visibleVolume"`
	FilledVolume  float64     `json:"filledVolume"`
	Price         float64     `json:"price"`
	Status        string      `json:"status"`
	CreatedAt     time.Time   `json:"createdAt"`
	TimeInForce   TimeInForce `json:"timeInForce"`
}

// TWAPOrder executes volume evenly over time
type TWAPOrder struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"`
	TotalVolume  float64   `json:"totalVolume"`
	FilledVolume float64   `json:"filledVolume"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	Interval     int       `json:"interval"` // Seconds between slices
	MinPrice     float64   `json:"minPrice,omitempty"`
	MaxPrice     float64   `json:"maxPrice,omitempty"`
	Status       string    `json:"status"`
	SliceCount   int       `json:"sliceCount"`
	Slices       []OrderSlice `json:"slices"`
}

// VWAPOrder executes targeting VWAP price
type VWAPOrder struct {
	ID              string    `json:"id"`
	Symbol          string    `json:"symbol"`
	Side            string    `json:"side"`
	TotalVolume     float64   `json:"totalVolume"`
	FilledVolume    float64   `json:"filledVolume"`
	StartTime       time.Time `json:"startTime"`
	EndTime         time.Time `json:"endTime"`
	TargetVWAP      float64   `json:"targetVwap"`
	MaxDeviation    float64   `json:"maxDeviation"` // % from VWAP
	Status          string    `json:"status"`
	VolumeProfile   []VolumeSlice `json:"volumeProfile"`
}

// OrderSlice represents a portion of algo order
type OrderSlice struct {
	SliceNumber  int       `json:"sliceNumber"`
	Volume       float64   `json:"volume"`
	FilledVolume float64   `json:"filledVolume"`
	Price        float64   `json:"price,omitempty"`
	ExecutedAt   *time.Time `json:"executedAt,omitempty"`
	Status       string    `json:"status"`
}

// VolumeSlice for VWAP volume distribution
type VolumeSlice struct {
	TimeSlot     time.Time `json:"timeSlot"`
	Volume       float64   `json:"volume"`
	TargetVolume float64   `json:"targetVolume"`
	VWAP         float64   `json:"vwap"`
}

// AdvancedOrderService manages advanced order types
type AdvancedOrderService struct {
	mu             sync.RWMutex
	bracketOrders  map[string]*BracketOrder
	icebergOrders  map[string]*IcebergOrder
	twapOrders     map[string]*TWAPOrder
	vwapOrders     map[string]*VWAPOrder

	// Callbacks
	priceCallback    func(symbol string) (bid, ask float64, ok bool)
	executeCallback  func(symbol, side string, volume, price float64) (string, error)
	cancelCallback   func(orderID string) error
	volumeCallback   func(symbol string) float64 // For VWAP
}

// NewAdvancedOrderService creates the service
func NewAdvancedOrderService() *AdvancedOrderService {
	svc := &AdvancedOrderService{
		bracketOrders: make(map[string]*BracketOrder),
		icebergOrders: make(map[string]*IcebergOrder),
		twapOrders:    make(map[string]*TWAPOrder),
		vwapOrders:    make(map[string]*VWAPOrder),
	}

	go svc.processLoop()

	log.Println("[AdvancedOrderService] Initialized with TWAP, VWAP, Iceberg, Bracket orders")
	return svc
}

// SetCallbacks configures the service callbacks
func (s *AdvancedOrderService) SetCallbacks(
	priceCallback func(symbol string) (bid, ask float64, ok bool),
	executeCallback func(symbol, side string, volume, price float64) (string, error),
	cancelCallback func(orderID string) error,
	volumeCallback func(symbol string) float64,
) {
	s.priceCallback = priceCallback
	s.executeCallback = executeCallback
	s.cancelCallback = cancelCallback
	s.volumeCallback = volumeCallback
}

// PlaceBracketOrder creates a bracket order (entry + SL + TP)
func (s *AdvancedOrderService) PlaceBracketOrder(
	symbol, side string,
	volume, entryPrice, stopLoss, takeProfit float64,
	entryType string,
	tif TimeInForce,
	expiryTime *time.Time,
) (*BracketOrder, error) {

	if volume <= 0 {
		return nil, errors.New("invalid volume")
	}

	// Validate SL/TP placement
	if side == "BUY" {
		if stopLoss >= entryPrice || takeProfit <= entryPrice {
			return nil, errors.New("invalid SL/TP for BUY order")
		}
	} else {
		if stopLoss <= entryPrice || takeProfit >= entryPrice {
			return nil, errors.New("invalid SL/TP for SELL order")
		}
	}

	bracket := &BracketOrder{
		ID:          uuid.New().String(),
		Symbol:      symbol,
		Side:        side,
		Volume:      volume,
		EntryPrice:  entryPrice,
		EntryType:   entryType,
		StopLoss:    stopLoss,
		TakeProfit:  takeProfit,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
		TimeInForce: tif,
		ExpiryTime:  expiryTime,
	}

	s.mu.Lock()
	s.bracketOrders[bracket.ID] = bracket
	s.mu.Unlock()

	log.Printf("[BracketOrder] Created %s %s %.2f @ %.5f SL:%.5f TP:%.5f",
		side, symbol, volume, entryPrice, stopLoss, takeProfit)

	return bracket, nil
}

// PlaceIcebergOrder creates an iceberg order with hidden volume
func (s *AdvancedOrderService) PlaceIcebergOrder(
	symbol, side string,
	totalVolume, visibleVolume, price float64,
	tif TimeInForce,
) (*IcebergOrder, error) {

	if totalVolume <= visibleVolume {
		return nil, errors.New("total volume must be greater than visible volume")
	}

	iceberg := &IcebergOrder{
		ID:            uuid.New().String(),
		Symbol:        symbol,
		Side:          side,
		TotalVolume:   totalVolume,
		VisibleVolume: visibleVolume,
		FilledVolume:  0,
		Price:         price,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		TimeInForce:   tif,
	}

	s.mu.Lock()
	s.icebergOrders[iceberg.ID] = iceberg
	s.mu.Unlock()

	log.Printf("[IcebergOrder] Created %s %s total:%.2f visible:%.2f @ %.5f",
		side, symbol, totalVolume, visibleVolume, price)

	return iceberg, nil
}

// PlaceTWAPOrder creates a TWAP order
func (s *AdvancedOrderService) PlaceTWAPOrder(
	symbol, side string,
	totalVolume float64,
	startTime, endTime time.Time,
	interval int,
	minPrice, maxPrice float64,
) (*TWAPOrder, error) {

	if endTime.Before(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	duration := endTime.Sub(startTime)
	sliceCount := int(duration.Seconds()) / interval

	if sliceCount < 1 {
		sliceCount = 1
	}

	volumePerSlice := totalVolume / float64(sliceCount)

	twap := &TWAPOrder{
		ID:          uuid.New().String(),
		Symbol:      symbol,
		Side:        side,
		TotalVolume: totalVolume,
		StartTime:   startTime,
		EndTime:     endTime,
		Interval:    interval,
		MinPrice:    minPrice,
		MaxPrice:    maxPrice,
		Status:      "PENDING",
		SliceCount:  sliceCount,
		Slices:      make([]OrderSlice, sliceCount),
	}

	// Initialize slices
	for i := 0; i < sliceCount; i++ {
		twap.Slices[i] = OrderSlice{
			SliceNumber: i + 1,
			Volume:      volumePerSlice,
			Status:      "PENDING",
		}
	}

	s.mu.Lock()
	s.twapOrders[twap.ID] = twap
	s.mu.Unlock()

	log.Printf("[TWAPOrder] Created %s %s %.2f over %v in %d slices",
		side, symbol, totalVolume, duration, sliceCount)

	return twap, nil
}

// PlaceVWAPOrder creates a VWAP order
func (s *AdvancedOrderService) PlaceVWAPOrder(
	symbol, side string,
	totalVolume float64,
	startTime, endTime time.Time,
	maxDeviation float64,
) (*VWAPOrder, error) {

	if endTime.Before(startTime) {
		return nil, errors.New("end time must be after start time")
	}

	vwap := &VWAPOrder{
		ID:           uuid.New().String(),
		Symbol:       symbol,
		Side:         side,
		TotalVolume:  totalVolume,
		StartTime:    startTime,
		EndTime:      endTime,
		MaxDeviation: maxDeviation,
		Status:       "PENDING",
	}

	s.mu.Lock()
	s.vwapOrders[vwap.ID] = vwap
	s.mu.Unlock()

	log.Printf("[VWAPOrder] Created %s %s %.2f targeting VWAP ±%.2f%%",
		side, symbol, totalVolume, maxDeviation)

	return vwap, nil
}

// CancelBracketOrder cancels a bracket order
func (s *AdvancedOrderService) CancelBracketOrder(orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bracket, exists := s.bracketOrders[orderID]
	if !exists {
		return errors.New("bracket order not found")
	}

	bracket.Status = "CANCELLED"

	// Cancel child orders if they exist
	if s.cancelCallback != nil {
		if bracket.EntryOrderID != "" {
			s.cancelCallback(bracket.EntryOrderID)
		}
		if bracket.SLOrderID != "" {
			s.cancelCallback(bracket.SLOrderID)
		}
		if bracket.TPOrderID != "" {
			s.cancelCallback(bracket.TPOrderID)
		}
	}

	delete(s.bracketOrders, orderID)

	log.Printf("[BracketOrder] Cancelled %s", orderID)
	return nil
}

// GetBracketOrders returns all bracket orders
func (s *AdvancedOrderService) GetBracketOrders() []*BracketOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*BracketOrder, 0, len(s.bracketOrders))
	for _, order := range s.bracketOrders {
		orders = append(orders, order)
	}
	return orders
}

// GetTWAPOrders returns all TWAP orders
func (s *AdvancedOrderService) GetTWAPOrders() []*TWAPOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*TWAPOrder, 0, len(s.twapOrders))
	for _, order := range s.twapOrders {
		orders = append(orders, order)
	}
	return orders
}

// processLoop handles order execution
func (s *AdvancedOrderService) processLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	for range ticker.C {
		s.processBracketOrders()
		s.processIcebergOrders()
		s.processTWAPOrders()
		s.processVWAPOrders()
	}
}

func (s *AdvancedOrderService) processBracketOrders() {
	if s.priceCallback == nil || s.executeCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, bracket := range s.bracketOrders {
		// Check expiry for GTD orders
		if bracket.TimeInForce == TIF_GTD && bracket.ExpiryTime != nil {
			if time.Now().After(*bracket.ExpiryTime) {
				bracket.Status = "EXPIRED"
				delete(s.bracketOrders, id)
				log.Printf("[BracketOrder] Expired %s", id)
				continue
			}
		}

		if bracket.Status == "PENDING" {
			// Check if entry should trigger
			bid, ask, ok := s.priceCallback(bracket.Symbol)
			if !ok {
				continue
			}

			shouldTrigger := false

			if bracket.EntryType == "MARKET" {
				shouldTrigger = true
			} else if bracket.EntryType == "LIMIT" {
				if bracket.Side == "BUY" && ask <= bracket.EntryPrice {
					shouldTrigger = true
				} else if bracket.Side == "SELL" && bid >= bracket.EntryPrice {
					shouldTrigger = true
				}
			} else if bracket.EntryType == "STOP" {
				if bracket.Side == "BUY" && ask >= bracket.EntryPrice {
					shouldTrigger = true
				} else if bracket.Side == "SELL" && bid <= bracket.EntryPrice {
					shouldTrigger = true
				}
			}

			if shouldTrigger {
				// Execute entry
				price := ask
				if bracket.Side == "SELL" {
					price = bid
				}

				orderID, err := s.executeCallback(bracket.Symbol, bracket.Side, bracket.Volume, price)
				if err == nil {
					bracket.EntryOrderID = orderID
					bracket.Status = "ACTIVE"
					log.Printf("[BracketOrder] Entry executed %s @ %.5f", id, price)

					// Place SL and TP orders
					// (Would need integration with pending order system)
				}
			}
		}
	}
}

func (s *AdvancedOrderService) processIcebergOrders() {
	if s.priceCallback == nil || s.executeCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for id, iceberg := range s.icebergOrders {
		if iceberg.Status != "PENDING" && iceberg.Status != "PARTIAL" {
			continue
		}

		remaining := iceberg.TotalVolume - iceberg.FilledVolume
		if remaining <= 0 {
			iceberg.Status = "FILLED"
			delete(s.icebergOrders, id)
			continue
		}

		// Execute visible portion
		volumeToExecute := math.Min(iceberg.VisibleVolume, remaining)

		bid, ask, ok := s.priceCallback(iceberg.Symbol)
		if !ok {
			continue
		}

		price := ask
		if iceberg.Side == "SELL" {
			price = bid
		}

		// Check if price matches limit
		if (iceberg.Side == "BUY" && ask <= iceberg.Price) ||
		   (iceberg.Side == "SELL" && bid >= iceberg.Price) {

			_, err := s.executeCallback(iceberg.Symbol, iceberg.Side, volumeToExecute, price)
			if err == nil {
				iceberg.FilledVolume += volumeToExecute
				iceberg.Status = "PARTIAL"
				log.Printf("[IcebergOrder] Partial fill %s: %.2f/%.2f",
					id, iceberg.FilledVolume, iceberg.TotalVolume)
			}
		}
	}
}

func (s *AdvancedOrderService) processTWAPOrders() {
	if s.executeCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for id, twap := range s.twapOrders {
		if twap.Status != "PENDING" && twap.Status != "ACTIVE" {
			continue
		}

		// Check if order should start
		if now.Before(twap.StartTime) {
			continue
		}

		// Check if order should end
		if now.After(twap.EndTime) {
			twap.Status = "COMPLETED"
			delete(s.twapOrders, id)
			continue
		}

		twap.Status = "ACTIVE"

		// Calculate which slice should execute now
		elapsed := now.Sub(twap.StartTime)
		currentSlice := int(elapsed.Seconds()) / twap.Interval

		if currentSlice >= len(twap.Slices) {
			continue
		}

		slice := &twap.Slices[currentSlice]
		if slice.Status == "PENDING" {
			bid, ask, ok := s.priceCallback(twap.Symbol)
			if !ok {
				continue
			}

			price := ask
			if twap.Side == "SELL" {
				price = bid
			}

			// Check price limits
			if twap.MinPrice > 0 && price < twap.MinPrice {
				continue
			}
			if twap.MaxPrice > 0 && price > twap.MaxPrice {
				continue
			}

			_, err := s.executeCallback(twap.Symbol, twap.Side, slice.Volume, price)
			if err == nil {
				slice.FilledVolume = slice.Volume
				slice.Price = price
				now := time.Now()
				slice.ExecutedAt = &now
				slice.Status = "FILLED"
				twap.FilledVolume += slice.Volume

				log.Printf("[TWAPOrder] Slice %d/%d executed %s @ %.5f",
					currentSlice+1, twap.SliceCount, id, price)
			}
		}
	}
}

func (s *AdvancedOrderService) processVWAPOrders() {
	if s.executeCallback == nil || s.volumeCallback == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for id, vwap := range s.vwapOrders {
		if vwap.Status != "PENDING" && vwap.Status != "ACTIVE" {
			continue
		}

		if now.Before(vwap.StartTime) {
			continue
		}

		if now.After(vwap.EndTime) {
			vwap.Status = "COMPLETED"
			delete(s.vwapOrders, id)
			continue
		}

		vwap.Status = "ACTIVE"

		// Calculate current VWAP (simplified - would need historical data)
		bid, ask, ok := s.priceCallback(vwap.Symbol)
		if !ok {
			continue
		}

		currentPrice := (bid + ask) / 2
		_ = s.volumeCallback(vwap.Symbol) // TODO: Use volume in VWAP calculation

		// Simple VWAP calculation (in production, use proper VWAP from tick data)
		if vwap.TargetVWAP == 0 {
			vwap.TargetVWAP = currentPrice
		}

		// Update VWAP estimate
		vwap.TargetVWAP = (vwap.TargetVWAP + currentPrice) / 2

		// Check if current price is within deviation
		deviation := math.Abs(currentPrice-vwap.TargetVWAP) / vwap.TargetVWAP * 100

		if deviation <= vwap.MaxDeviation {
			// Execute a portion of the order
			remaining := vwap.TotalVolume - vwap.FilledVolume
			if remaining > 0 {
				// Execute small portion based on time remaining
				duration := vwap.EndTime.Sub(vwap.StartTime)
				elapsed := now.Sub(vwap.StartTime)
				targetFilled := vwap.TotalVolume * float64(elapsed) / float64(duration)
				volumeToExecute := math.Min(targetFilled-vwap.FilledVolume, remaining)

				if volumeToExecute > 0 {
					price := ask
					if vwap.Side == "SELL" {
						price = bid
					}

					_, err := s.executeCallback(vwap.Symbol, vwap.Side, volumeToExecute, price)
					if err == nil {
						vwap.FilledVolume += volumeToExecute
						log.Printf("[VWAPOrder] Partial fill %s: %.2f/%.2f @ %.5f (VWAP: %.5f)",
							id, vwap.FilledVolume, vwap.TotalVolume, price, vwap.TargetVWAP)
					}
				}
			}
		}
	}
}
