package abook

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/fix"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/risk"
	"github.com/google/uuid"
)

// ExecutionEngine handles A-Book order execution to LPs
type ExecutionEngine struct {
	fixGateway    *fix.FIXGateway
	lpManager     *lpmanager.Manager
	riskEngine    *risk.Engine
	sor           *SmartOrderRouter
	execReports   chan *ExecutionReport
	orders        map[string]*Order
	positions     map[string]*Position
	mu            sync.RWMutex

	// Metrics
	metrics       *ExecutionMetrics

	// Callbacks
	onFill        func(order *Order, fill *Fill)
	onReject      func(order *Order, reason string)
	onUpdate      func(order *Order)
}

// Order represents an A-Book order
type Order struct {
	ID            string
	ClientOrderID string
	AccountID     string
	Symbol        string
	Side          string  // BUY or SELL
	Type          string  // MARKET, LIMIT
	Volume        float64
	Price         float64 // Limit price (0 for market)
	SL            float64
	TP            float64
	Status        string  // PENDING, SENT, PARTIAL, FILLED, REJECTED, CANCELED
	SelectedLP    string
	LPOrderID     string
	FilledQty     float64
	AvgFillPrice  float64
	Slippage      float64
	CreatedAt     time.Time
	SentAt        *time.Time
	FilledAt      *time.Time
	Fills         []*Fill
	RejectReason  string
}

// Fill represents a partial or full execution
type Fill struct {
	ID            string
	OrderID       string
	ExecID        string
	Symbol        string
	Side          string
	Quantity      float64
	Price         float64
	LP            string
	Timestamp     time.Time
	Commission    float64
}

// Position represents an open A-Book position
type Position struct {
	ID           string
	OrderID      string
	AccountID    string
	Symbol       string
	Side         string
	Volume       float64
	OpenPrice    float64
	CurrentPrice float64
	SL           float64
	TP           float64
	Swap         float64
	Commission   float64
	UnrealizedPL float64
	LP           string
	LPPositionID string
	OpenTime     time.Time
}

// ExecutionReport from LP
type ExecutionReport struct {
	OrderID       string
	ClientOrderID string
	ExecType      string // NEW, PARTIAL_FILL, FILL, REJECTED, CANCELED
	Symbol        string
	Side          string
	OrderQty      float64
	LastQty       float64
	LastPx        float64
	CumQty        float64
	AvgPx         float64
	OrdStatus     string
	LP            string
	LPOrderID     string
	Text          string
	Timestamp     time.Time
}

// ExecutionMetrics tracks execution quality
type ExecutionMetrics struct {
	TotalOrders       int64
	FilledOrders      int64
	RejectedOrders    int64
	PartialFills      int64
	AvgSlippage       float64
	AvgLatency        time.Duration
	FillRateByLP      map[string]float64
	SlippageByLP      map[string]float64
	AvgLatencyByLP    map[string]time.Duration
	mu                sync.RWMutex
}

// NewExecutionEngine creates a new A-Book execution engine
func NewExecutionEngine(
	fixGateway *fix.FIXGateway,
	lpManager *lpmanager.Manager,
	riskEngine *risk.Engine,
) *ExecutionEngine {
	engine := &ExecutionEngine{
		fixGateway:  fixGateway,
		lpManager:   lpManager,
		riskEngine:  riskEngine,
		sor:         NewSmartOrderRouter(lpManager),
		execReports: make(chan *ExecutionReport, 1000),
		orders:      make(map[string]*Order),
		positions:   make(map[string]*Position),
		metrics: &ExecutionMetrics{
			FillRateByLP:   make(map[string]float64),
			SlippageByLP:   make(map[string]float64),
			AvgLatencyByLP: make(map[string]time.Duration),
		},
	}

	// Start execution report processor
	go engine.processExecutionReports()

	return engine
}

// PlaceOrder executes an order to the LP
func (e *ExecutionEngine) PlaceOrder(req *OrderRequest) (*Order, error) {
	// 1. Pre-trade validation
	if err := e.validateOrder(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Pre-trade risk check
	accountID, err := strconv.ParseInt(req.AccountID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID: %w", err)
	}
	if err := e.riskEngine.PreTradeCheck(accountID, req.Symbol, req.Volume, req.Price); err != nil {
		return nil, fmt.Errorf("risk check failed: %w", err)
	}

	// 3. Create order
	order := &Order{
		ID:            uuid.New().String(),
		ClientOrderID: req.ClientOrderID,
		AccountID:     req.AccountID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Volume:        req.Volume,
		Price:         req.Price,
		SL:            req.SL,
		TP:            req.TP,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		Fills:         make([]*Fill, 0),
	}

	if order.ClientOrderID == "" {
		order.ClientOrderID = order.ID
	}

	// 4. Smart Order Routing - select best LP
	lpSelection, err := e.sor.SelectLP(req.Symbol, req.Side, req.Volume)
	if err != nil {
		order.Status = "REJECTED"
		order.RejectReason = fmt.Sprintf("No LP available: %v", err)
		e.mu.Lock()
		e.orders[order.ID] = order
		e.mu.Unlock()
		return order, err
	}

	order.SelectedLP = lpSelection.LPID

	// 5. Route to LP via FIX
	if err := e.routeToLP(order, lpSelection); err != nil {
		order.Status = "REJECTED"
		order.RejectReason = fmt.Sprintf("Routing failed: %v", err)
		e.mu.Lock()
		e.orders[order.ID] = order
		e.mu.Unlock()
		return order, err
	}

	order.Status = "SENT"
	now := time.Now()
	order.SentAt = &now

	e.mu.Lock()
	e.orders[order.ID] = order
	e.mu.Unlock()

	log.Printf("[A-Book] Order %s sent to %s for %s %.2f @ %.5f",
		order.ClientOrderID, order.SelectedLP, order.Symbol, order.Volume, order.Price)

	// Update metrics
	e.metrics.mu.Lock()
	e.metrics.TotalOrders++
	e.metrics.mu.Unlock()

	return order, nil
}

// CancelOrder cancels an order at the LP
func (e *ExecutionEngine) CancelOrder(orderID string) error {
	e.mu.RLock()
	order, exists := e.orders[orderID]
	e.mu.RUnlock()

	if !exists {
		return errors.New("order not found")
	}

	if order.Status == "FILLED" || order.Status == "CANCELED" || order.Status == "REJECTED" {
		return fmt.Errorf("cannot cancel order in status: %s", order.Status)
	}

	// Send cancel request via FIX
	_, cancelErr := e.fixGateway.CancelOrder(order.SelectedLP, order.ClientOrderID, order.Symbol, order.Side)
	if cancelErr != nil {
		return fmt.Errorf("cancel request failed: %w", cancelErr)
	}

	log.Printf("[A-Book] Cancel request sent for order %s", orderID)
	return nil
}

// ClosePosition closes a position by sending opposite order to LP
func (e *ExecutionEngine) ClosePosition(positionID string) error {
	e.mu.RLock()
	position, exists := e.positions[positionID]
	e.mu.RUnlock()

	if !exists {
		return errors.New("position not found")
	}

	// Create opposite order
	oppositeSide := "SELL"
	if position.Side == "SELL" {
		oppositeSide = "BUY"
	}

	req := &OrderRequest{
		AccountID: position.AccountID,
		Symbol:    position.Symbol,
		Side:      oppositeSide,
		Type:      "MARKET",
		Volume:    position.Volume,
		Price:     0, // Market order
	}

	closeOrder, err := e.PlaceOrder(req)
	if err != nil {
		return fmt.Errorf("failed to place closing order: %w", err)
	}

	log.Printf("[A-Book] Position %s closing with order %s", positionID, closeOrder.ID)
	return nil
}

// GetOrder retrieves an order by ID
func (e *ExecutionEngine) GetOrder(orderID string) (*Order, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	order, exists := e.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}

	return order, nil
}

// GetPositions returns all open positions
func (e *ExecutionEngine) GetPositions(accountID string) []*Position {
	e.mu.RLock()
	defer e.mu.RUnlock()

	positions := make([]*Position, 0)
	for _, pos := range e.positions {
		if accountID == "" || pos.AccountID == accountID {
			positions = append(positions, pos)
		}
	}

	return positions
}

// GetMetrics returns execution quality metrics
func (e *ExecutionEngine) GetMetrics() *ExecutionMetrics {
	e.metrics.mu.RLock()
	defer e.metrics.mu.RUnlock()

	// Create a copy
	metrics := &ExecutionMetrics{
		TotalOrders:    e.metrics.TotalOrders,
		FilledOrders:   e.metrics.FilledOrders,
		RejectedOrders: e.metrics.RejectedOrders,
		PartialFills:   e.metrics.PartialFills,
		AvgSlippage:    e.metrics.AvgSlippage,
		AvgLatency:     e.metrics.AvgLatency,
		FillRateByLP:   make(map[string]float64),
		SlippageByLP:   make(map[string]float64),
		AvgLatencyByLP: make(map[string]time.Duration),
	}

	for k, v := range e.metrics.FillRateByLP {
		metrics.FillRateByLP[k] = v
	}
	for k, v := range e.metrics.SlippageByLP {
		metrics.SlippageByLP[k] = v
	}
	for k, v := range e.metrics.AvgLatencyByLP {
		metrics.AvgLatencyByLP[k] = v
	}

	return metrics
}

// SetOnFillCallback sets the callback for fill events
func (e *ExecutionEngine) SetOnFillCallback(callback func(*Order, *Fill)) {
	e.onFill = callback
}

// SetOnRejectCallback sets the callback for reject events
func (e *ExecutionEngine) SetOnRejectCallback(callback func(*Order, string)) {
	e.onReject = callback
}

// SetOnUpdateCallback sets the callback for order update events
func (e *ExecutionEngine) SetOnUpdateCallback(callback func(*Order)) {
	e.onUpdate = callback
}

// processExecutionReports processes incoming execution reports from LPs
func (e *ExecutionEngine) processExecutionReports() {
	log.Println("[A-Book] Execution report processor started")

	// Subscribe to FIX gateway execution reports
	fixExecReports := e.fixGateway.GetExecutionReports()

	for {
		select {
		case fixReport := <-fixExecReports:
			e.handleFIXExecutionReport(&fixReport)
		case report := <-e.execReports:
			e.handleExecutionReport(report)
		}
	}
}

// handleFIXExecutionReport converts FIX report to internal format
func (e *ExecutionEngine) handleFIXExecutionReport(fixReport *fix.ExecutionReport) {
	report := &ExecutionReport{
		ClientOrderID: fixReport.OrderID,
		ExecType:      fixReport.ExecType,
		Symbol:        fixReport.Symbol,
		Side:          fixReport.Side,
		OrderQty:      fixReport.Volume,
		LastQty:       fixReport.Volume,
		LastPx:        fixReport.Price,
		CumQty:        fixReport.Volume,
		AvgPx:         fixReport.Price,
		LP:            "FIX",
		LPOrderID:     fixReport.LPOrderID,
		Text:          fixReport.Text,
		Timestamp:     fixReport.Timestamp,
	}

	// Map FIX exec type to our status
	switch fixReport.ExecType {
	case "NEW":
		report.OrdStatus = "NEW"
	case "FILLED":
		report.OrdStatus = "FILLED"
	case "REJECTED":
		report.OrdStatus = "REJECTED"
	case "CANCELED":
		report.OrdStatus = "CANCELED"
	}

	e.handleExecutionReport(report)
}

// handleExecutionReport processes an execution report
func (e *ExecutionEngine) handleExecutionReport(report *ExecutionReport) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Find order by client order ID
	var order *Order
	for _, o := range e.orders {
		if o.ClientOrderID == report.ClientOrderID {
			order = o
			break
		}
	}

	if order == nil {
		log.Printf("[A-Book] Received execution report for unknown order: %s", report.ClientOrderID)
		return
	}

	log.Printf("[A-Book] Execution report: %s %s %s - Qty: %.2f @ %.5f",
		order.ClientOrderID, report.ExecType, report.OrdStatus, report.LastQty, report.LastPx)

	switch report.ExecType {
	case "NEW":
		order.Status = "SENT"
		order.LPOrderID = report.LPOrderID

	case "PARTIAL_FILL", "FILL":
		// Create fill record
		fill := &Fill{
			ID:         uuid.New().String(),
			OrderID:    order.ID,
			ExecID:     report.LPOrderID,
			Symbol:     order.Symbol,
			Side:       order.Side,
			Quantity:   report.LastQty,
			Price:      report.LastPx,
			LP:         order.SelectedLP,
			Timestamp:  report.Timestamp,
			Commission: 0, // TODO: Calculate commission
		}

		order.Fills = append(order.Fills, fill)
		order.FilledQty = report.CumQty
		order.AvgFillPrice = report.AvgPx

		// Calculate slippage
		if order.Price > 0 {
			if order.Side == "BUY" {
				order.Slippage = report.AvgPx - order.Price
			} else {
				order.Slippage = order.Price - report.AvgPx
			}
		}

		if report.ExecType == "FILL" {
			order.Status = "FILLED"
			now := time.Now()
			order.FilledAt = &now

			// Create position
			e.createPosition(order, fill)

			// Update metrics
			e.metrics.mu.Lock()
			e.metrics.FilledOrders++
			e.updateMetrics(order)
			e.metrics.mu.Unlock()

			log.Printf("[A-Book] Order %s FILLED: %.2f @ %.5f (slippage: %.5f)",
				order.ClientOrderID, order.FilledQty, order.AvgFillPrice, order.Slippage)
		} else {
			order.Status = "PARTIAL"
			e.metrics.mu.Lock()
			e.metrics.PartialFills++
			e.metrics.mu.Unlock()
		}

		// Callback
		if e.onFill != nil {
			e.onFill(order, fill)
		}

	case "REJECTED":
		order.Status = "REJECTED"
		order.RejectReason = report.Text

		e.metrics.mu.Lock()
		e.metrics.RejectedOrders++
		e.metrics.mu.Unlock()

		log.Printf("[A-Book] Order %s REJECTED: %s", order.ClientOrderID, report.Text)

		// Callback
		if e.onReject != nil {
			e.onReject(order, report.Text)
		}

	case "CANCELED":
		order.Status = "CANCELED"
		log.Printf("[A-Book] Order %s CANCELED", order.ClientOrderID)
	}

	// Update callback
	if e.onUpdate != nil {
		e.onUpdate(order)
	}
}

// createPosition creates a new position from a filled order
func (e *ExecutionEngine) createPosition(order *Order, fill *Fill) {
	position := &Position{
		ID:           uuid.New().String(),
		OrderID:      order.ID,
		AccountID:    order.AccountID,
		Symbol:       order.Symbol,
		Side:         order.Side,
		Volume:       order.FilledQty,
		OpenPrice:    order.AvgFillPrice,
		CurrentPrice: order.AvgFillPrice,
		SL:           order.SL,
		TP:           order.TP,
		LP:           order.SelectedLP,
		LPPositionID: order.LPOrderID,
		OpenTime:     time.Now(),
	}

	e.positions[position.ID] = position
	log.Printf("[A-Book] Position created: %s %s %.2f @ %.5f",
		position.Symbol, position.Side, position.Volume, position.OpenPrice)
}

// updateMetrics updates execution quality metrics
func (e *ExecutionEngine) updateMetrics(order *Order) {
	// Update average slippage
	totalSlippage := e.metrics.AvgSlippage * float64(e.metrics.FilledOrders-1)
	e.metrics.AvgSlippage = (totalSlippage + order.Slippage) / float64(e.metrics.FilledOrders)

	// Update latency
	if order.SentAt != nil && order.FilledAt != nil {
		latency := order.FilledAt.Sub(*order.SentAt)
		totalLatency := e.metrics.AvgLatency * time.Duration(e.metrics.FilledOrders-1)
		e.metrics.AvgLatency = (totalLatency + latency) / time.Duration(e.metrics.FilledOrders)
	}

	// Update LP-specific metrics
	lp := order.SelectedLP
	if lp != "" {
		// Fill rate
		lpFilled := e.metrics.FillRateByLP[lp]
		e.metrics.FillRateByLP[lp] = lpFilled + 1

		// Slippage
		lpSlippage := e.metrics.SlippageByLP[lp]
		e.metrics.SlippageByLP[lp] = (lpSlippage + order.Slippage) / e.metrics.FillRateByLP[lp]

		// Latency
		if order.SentAt != nil && order.FilledAt != nil {
			latency := order.FilledAt.Sub(*order.SentAt)
			lpLatency := e.metrics.AvgLatencyByLP[lp]
			e.metrics.AvgLatencyByLP[lp] = (lpLatency + latency) / 2
		}
	}
}

// routeToLP sends the order to the selected LP via FIX
func (e *ExecutionEngine) routeToLP(order *Order, lpSelection *LPSelection) error {
	// TODO: FIX gateway SendOrder doesn't currently support order type parameter
	// Market and Limit orders are handled based on price (0 = market)

	// Convert side
	fixSide := "1" // Buy
	if order.Side == "SELL" {
		fixSide = "2" // Sell
	}

	// Send via FIX gateway (returns clOrdID)
	clOrdID, err := e.fixGateway.SendOrder(
		lpSelection.SessionID,
		order.Symbol,
		fixSide,
		order.Volume,
		order.Price,
	)

	if err != nil {
		return fmt.Errorf("FIX order send failed: %w", err)
	}

	// Store the LP order ID returned by FIX gateway
	order.LPOrderID = clOrdID

	return nil
}

// validateOrder validates order parameters
func (e *ExecutionEngine) validateOrder(req *OrderRequest) error {
	if req.Symbol == "" {
		return errors.New("symbol is required")
	}

	if req.Side != "BUY" && req.Side != "SELL" {
		return errors.New("side must be BUY or SELL")
	}

	if req.Type != "MARKET" && req.Type != "LIMIT" {
		return errors.New("type must be MARKET or LIMIT")
	}

	if req.Volume <= 0 {
		return errors.New("volume must be positive")
	}

	if req.Type == "LIMIT" && req.Price <= 0 {
		return errors.New("price is required for limit orders")
	}

	return nil
}

// OrderRequest represents a request to place an order
type OrderRequest struct {
	ClientOrderID string
	AccountID     string
	Symbol        string
	Side          string
	Type          string
	Volume        float64
	Price         float64
	SL            float64
	TP            float64
}
