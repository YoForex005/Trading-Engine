package core

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/epic1st/rtx/backend/internal/database/repository"
)

// OrderMonitor continuously monitors pending orders with trigger prices
// and executes them when market conditions are met
type OrderMonitor struct {
	mu              sync.RWMutex
	orderRepo       *repository.OrderRepository
	engine          *Engine
	stopChan        chan struct{}
	tickInterval    time.Duration
	priceCallback   func(symbol string) (bid, ask float64, ok bool)
	running         bool
}

// NewOrderMonitor creates a new order monitoring service
func NewOrderMonitor(
	orderRepo *repository.OrderRepository,
	engine *Engine,
	priceCallback func(symbol string) (bid, ask float64, ok bool),
) *OrderMonitor {
	return &OrderMonitor{
		orderRepo:     orderRepo,
		engine:        engine,
		priceCallback: priceCallback,
		stopChan:      make(chan struct{}),
		tickInterval:  100 * time.Millisecond, // Check every 100ms
	}
}

// Start begins monitoring pending orders
func (om *OrderMonitor) Start() {
	om.mu.Lock()
	if om.running {
		om.mu.Unlock()
		return
	}
	om.running = true
	om.mu.Unlock()

	log.Println("[OrderMonitor] Starting order monitor service (tick interval: 100ms)")

	go om.monitorLoop()
}

// Stop gracefully stops the monitor
func (om *OrderMonitor) Stop() {
	om.mu.Lock()
	defer om.mu.Unlock()

	if !om.running {
		return
	}

	log.Println("[OrderMonitor] Stopping order monitor service")
	close(om.stopChan)
	om.running = false
}

// monitorLoop runs the continuous monitoring cycle
func (om *OrderMonitor) monitorLoop() {
	ticker := time.NewTicker(om.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-om.stopChan:
			log.Println("[OrderMonitor] Monitor loop stopped")
			return
		case <-ticker.C:
			om.checkPriceTriggers()
		}
	}
}

// checkPriceTriggers checks all pending orders and executes those whose trigger conditions are met
func (om *OrderMonitor) checkPriceTriggers() {
	ctx := context.Background()

	// Fetch all pending orders with trigger prices
	orders, err := om.orderRepo.ListPendingWithTriggers(ctx)
	if err != nil {
		log.Printf("[OrderMonitor] Error fetching pending orders: %v", err)
		return
	}

	if len(orders) == 0 {
		return // No pending orders to monitor
	}

	// Check for expired orders FIRST
	om.checkOrderExpiry(ctx, orders)

	// Update trailing stops BEFORE checking triggers
	om.updateTrailingStops(ctx, orders)

	// Check each order for trigger conditions
	for _, order := range orders {
		// Skip if already expired (status changed in checkOrderExpiry)
		if order.Status != "PENDING" {
			continue
		}
		om.checkOrder(ctx, order)
	}
}

// checkOrder evaluates a single order for trigger conditions
func (om *OrderMonitor) checkOrder(ctx context.Context, order *repository.Order) {
	// Get current market price
	bid, ask, ok := om.priceCallback(order.Symbol)
	if !ok {
		return // No price available
	}

	triggered := false

	// Determine trigger logic based on order type
	switch order.Type {
	case "STOP_LOSS":
		// SL for BUY position: triggers when bid <= trigger_price (market fell)
		// SL for SELL position: triggers when ask >= trigger_price (market rose)
		if order.Side == "BUY" && bid <= order.TriggerPrice {
			triggered = true
		} else if order.Side == "SELL" && ask >= order.TriggerPrice {
			triggered = true
		}

	case "TAKE_PROFIT":
		// TP for BUY position: triggers when bid >= trigger_price (market rose)
		// TP for SELL position: triggers when ask <= trigger_price (market fell)
		if order.Side == "BUY" && bid >= order.TriggerPrice {
			triggered = true
		} else if order.Side == "SELL" && ask <= order.TriggerPrice {
			triggered = true
		}

	case "STOP":
		// Buy Stop: triggers when ask >= trigger_price (breakout upward)
		// Sell Stop: triggers when bid <= trigger_price (breakout downward)
		if order.Side == "BUY" && ask >= order.TriggerPrice {
			triggered = true
		} else if order.Side == "SELL" && bid <= order.TriggerPrice {
			triggered = true
		}

	case "LIMIT":
		// Buy Limit: triggers when ask <= trigger_price (price falls to limit)
		// Sell Limit: triggers when bid >= trigger_price (price rises to limit)
		if order.Side == "BUY" && ask <= order.TriggerPrice {
			triggered = true
		} else if order.Side == "SELL" && bid >= order.TriggerPrice {
			triggered = true
		}
	}

	if triggered {
		log.Printf("[OrderMonitor] Trigger hit for Order #%d (%s %s) @ Bid:%.5f Ask:%.5f Trigger:%.5f",
			order.ID, order.Type, order.Symbol, bid, ask, order.TriggerPrice)

		// Execute the triggered order
		if err := om.engine.ExecuteTriggeredOrder(ctx, order.ID); err != nil {
			log.Printf("[OrderMonitor] Failed to execute triggered order #%d: %v", order.ID, err)
		}
	}
}

// checkOrderExpiry checks for expired orders and cancels them
func (om *OrderMonitor) checkOrderExpiry(ctx context.Context, orders []*repository.Order) {
	now := time.Now()

	for _, order := range orders {
		// Only process orders with expiry_time set
		if order.ExpiryTime == nil {
			continue
		}

		// Check if order has expired
		if now.After(*order.ExpiryTime) {
			// Cancel expired order
			order.Status = "CANCELLED"
			order.RejectReason = "Expired"

			if err := om.orderRepo.UpdateStatus(ctx, order.ID, "CANCELLED", nil); err != nil {
				log.Printf("[OrderMonitor] Failed to cancel expired order #%d: %v", order.ID, err)
				continue
			}

			log.Printf("[OrderMonitor] Order #%d expired and cancelled (expiry: %s)",
				order.ID, order.ExpiryTime.Format(time.RFC3339))

			// TODO: Broadcast order update via WebSocket
		}
	}
}

// updateTrailingStops adjusts trigger prices for trailing stop orders as market moves favorably
func (om *OrderMonitor) updateTrailingStops(ctx context.Context, orders []*repository.Order) {
	om.mu.Lock()
	defer om.mu.Unlock()

	for _, order := range orders {
		// Only process TRAILING_STOP orders with a trailing_delta set
		if order.Type != "TRAILING_STOP" || order.TrailingDelta == nil || *order.TrailingDelta <= 0 {
			continue
		}

		// Get current market price
		bid, ask, ok := om.priceCallback(order.Symbol)
		if !ok {
			continue // No price available for this symbol
		}

		trailingDelta := *order.TrailingDelta
		oldTrigger := order.TriggerPrice
		var newTrigger float64
		shouldUpdate := false

		// Calculate new trigger based on current price and trailing delta
		if order.Side == "BUY" {
			// For BUY (long position): trailing stop moves up with rising price
			// Trigger = current_ask - trailing_delta
			newTrigger = ask - trailingDelta

			// Only update if new trigger is HIGHER than old (more favorable)
			if newTrigger > oldTrigger {
				shouldUpdate = true
			}
		} else if order.Side == "SELL" {
			// For SELL (short position): trailing stop moves down with falling price
			// Trigger = current_bid + trailing_delta
			newTrigger = bid + trailingDelta

			// Only update if new trigger is LOWER than old (more favorable)
			if newTrigger < oldTrigger {
				shouldUpdate = true
			}
		}

		if shouldUpdate {
			// Update trigger price in database
			order.TriggerPrice = newTrigger
			if err := om.orderRepo.UpdateTriggerPrice(ctx, order.ID, newTrigger); err != nil {
				log.Printf("[OrderMonitor] Failed to update trailing stop #%d trigger: %v", order.ID, err)
				continue
			}

			log.Printf("[OrderMonitor] Trailing stop #%d adjusted: %.5f → %.5f (delta: %.5f, %s @ Bid:%.5f Ask:%.5f)",
				order.ID, oldTrigger, newTrigger, trailingDelta, order.Symbol, bid, ask)
		}
	}
}
