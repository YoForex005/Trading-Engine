package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/epic1st/rtx/backend/abook"
	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/fix"
	"github.com/epic1st/rtx/backend/internal/api/handlers"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/oms"
	"github.com/epic1st/rtx/backend/orders"
	"github.com/epic1st/rtx/backend/risk"
	"github.com/epic1st/rtx/backend/router"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
)

type Server struct {
	authService *auth.Service
	bbookAPI    *handlers.APIHandler

	omsService      *oms.Service
	riskEngine      *risk.Engine
	smartRouter     *router.SmartRouter
	fixGateway      *fix.FIXGateway
	hub             *ws.Hub
	tickStore       *tickstore.TickStore
	orderService    *orders.OrderService
	positionManager *orders.PositionManager
	trailingService *orders.TrailingStopService
	riskCalculator  *risk.RiskCalculator

	// A-Book execution
	abookEngine     *abook.ExecutionEngine
	abookHandler    *handlers.ABookHandler
}

func NewServer(authService *auth.Service, bbookAPI *handlers.APIHandler, lpMgr *lpmanager.Manager) *Server {
	fixGateway := fix.NewFIXGateway()
	riskEngine := risk.NewEngine()

	// Initialize A-Book execution engine
	abookEngine := abook.NewExecutionEngine(fixGateway, lpMgr, riskEngine)
	abookHandler := handlers.NewABookHandler(abookEngine)

	return &Server{
		authService:     authService,
		bbookAPI:        bbookAPI,
		omsService:      oms.NewService(),
		riskEngine:      riskEngine,
		smartRouter:     router.NewSmartRouter(),
		fixGateway:      fixGateway,
		orderService:    orders.NewOrderService(),
		positionManager: orders.NewPositionManager(true), // Hedging mode
		trailingService: orders.NewTrailingStopService(),
		riskCalculator:  risk.NewRiskCalculator(),
		abookEngine:     abookEngine,
		abookHandler:    abookHandler,
	}
}

func (s *Server) SetHub(hub *ws.Hub) {
	s.hub = hub
}

func (s *Server) SetTickStore(ts *tickstore.TickStore) {
	s.tickStore = ts
}

// GetOrderService returns the order service for external wiring
func (s *Server) GetOrderService() *orders.OrderService {
	return s.orderService
}

// GetPositionManager returns the position manager
func (s *Server) GetPositionManager() *orders.PositionManager {
	return s.positionManager
}

// GetTrailingService returns the trailing stop service
func (s *Server) GetTrailingService() *orders.TrailingStopService {
	return s.trailingService
}

// GetRiskCalculator returns the risk calculator
func (s *Server) GetRiskCalculator() *risk.RiskCalculator {
	return s.riskCalculator
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, user, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp := struct {
		Token string     `json:"token"`
		User  *auth.User `json:"user"`
	}{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		AccountID string  `json:"accountId,omitempty"`
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"`
		Volume    float64 `json:"volume"`
		Type      string  `json:"type,omitempty"` // Default MARKET
		Price     float64 `json:"price,omitempty"`
		SL        float64 `json:"sl,omitempty"`
		TP        float64 `json:"tp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Default to MARKET order
	if req.Type == "" {
		req.Type = "MARKET"
	}

	// Default account ID
	if req.AccountID == "" {
		req.AccountID = "demo_001"
	}

	log.Printf("[A-Book] Executing %s %s %.2f lots %s via LP",
		req.Side, req.Symbol, req.Volume, req.Type)

	// A-Book Execution via FIX
	orderReq := &abook.OrderRequest{
		AccountID: req.AccountID,
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Volume:    req.Volume,
		Price:     req.Price,
		SL:        req.SL,
		TP:        req.TP,
	}

	order, err := s.abookEngine.PlaceOrder(orderReq)
	if err != nil {
		log.Printf("[A-Book] Order placement failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"order":   order,
		"message": "Order sent to LP",
	})
}

// HandlePlaceLimitOrder handles limit order placement
func (s *Server) HandlePlaceLimitOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Symbol string  `json:"symbol"`
		Side   string  `json:"side"`
		Volume float64 `json:"volume"`
		Price  float64 `json:"price"`
		SL     float64 `json:"sl,omitempty"`
		TP     float64 `json:"tp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	side := orders.OrderSideBuy
	if req.Side == "SELL" {
		side = orders.OrderSideSell
	}

	order, err := s.orderService.PlaceLimitOrder(req.Symbol, side, req.Volume, req.Price, req.SL, req.TP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// HandlePlaceStopOrder handles stop order placement
func (s *Server) HandlePlaceStopOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Symbol       string  `json:"symbol"`
		Side         string  `json:"side"`
		Volume       float64 `json:"volume"`
		TriggerPrice float64 `json:"triggerPrice"`
		SL           float64 `json:"sl,omitempty"`
		TP           float64 `json:"tp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	side := orders.OrderSideBuy
	if req.Side == "SELL" {
		side = orders.OrderSideSell
	}

	order, err := s.orderService.PlaceStopOrder(req.Symbol, side, req.Volume, req.TriggerPrice, req.SL, req.TP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// HandlePlaceStopLimitOrder handles stop-limit order placement
func (s *Server) HandlePlaceStopLimitOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Symbol       string  `json:"symbol"`
		Side         string  `json:"side"`
		Volume       float64 `json:"volume"`
		TriggerPrice float64 `json:"triggerPrice"`
		LimitPrice   float64 `json:"limitPrice"`
		SL           float64 `json:"sl,omitempty"`
		TP           float64 `json:"tp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	side := orders.OrderSideBuy
	if req.Side == "SELL" {
		side = orders.OrderSideSell
	}

	order, err := s.orderService.PlaceStopLimitOrder(req.Symbol, side, req.Volume, req.TriggerPrice, req.LimitPrice, req.SL, req.TP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

// HandleGetPendingOrders returns all pending orders
func (s *Server) HandleGetPendingOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	orders := s.orderService.GetPendingOrders()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// HandleCancelOrder cancels a pending order
func (s *Server) HandleCancelOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		OrderID string `json:"orderId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.orderService.CancelOrder(req.OrderID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "orderId": req.OrderID})
}

// HandlePartialClose handles partial position close
func (s *Server) HandlePartialClose(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TradeID string  `json:"tradeId"`
		Percent float64 `json:"percent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For OANDA, we need to close partial units
	// if s.hub != nil && s.hub.GetOandaClient() != nil {
	// 	trades, err := s.hub.GetOandaClient().GetOpenTrades()
	//     // ...
	// }

	http.Error(w, "Partial close disabled (Dynamic LP Manager migration)", http.StatusServiceUnavailable)
}

// HandleCloseAll closes all positions
func (s *Server) HandleCloseAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		Symbol string `json:"symbol,omitempty"` // Empty = all symbols
	}

	json.NewDecoder(r.Body).Decode(&req) // Optional body

	// Legacy OANDA logic removed
	http.Error(w, "Close All disabled (Dynamic LP Manager migration)", http.StatusServiceUnavailable)
}

// HandleModifySLTP modifies stop loss and take profit
func (s *Server) HandleModifySLTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TradeID string  `json:"tradeId"`
		SL      float64 `json:"sl"`
		TP      float64 `json:"tp"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement OANDA trade modification
	log.Printf("[ORDER] Modify SL/TP for %s: SL=%.5f TP=%.5f", req.TradeID, req.SL, req.TP)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tradeId": req.TradeID,
		"sl":      req.SL,
		"tp":      req.TP,
	})
}

// HandleBreakeven sets SL to entry price
func (s *Server) HandleBreakeven(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TradeID string `json:"tradeId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Legacy OANDA logic removed

	http.Error(w, "Trade not found", http.StatusNotFound)
}

// HandleSetTrailingStop sets a trailing stop
func (s *Server) HandleSetTrailingStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TradeID  string  `json:"tradeId"`
		Symbol   string  `json:"symbol"`
		Side     string  `json:"side"`
		Type     string  `json:"type"` // FIXED, STEP, ATR
		Distance float64 `json:"distance"`
		StepSize float64 `json:"stepSize,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tsType := orders.TrailingFixed
	switch req.Type {
	case "STEP":
		tsType = orders.TrailingStep
	case "ATR":
		tsType = orders.TrailingATR
	}

	s.trailingService.SetTrailingStop(req.TradeID, req.Symbol, req.Side, tsType, req.Distance, req.StepSize)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"tradeId":  req.TradeID,
		"type":     req.Type,
		"distance": req.Distance,
	})
}

// HandleCalculateLot calculates lot size from risk
func (s *Server) HandleCalculateLot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	symbol := r.URL.Query().Get("symbol")
	riskPercent, _ := strconv.ParseFloat(r.URL.Query().Get("riskPercent"), 64)
	slPips, _ := strconv.ParseFloat(r.URL.Query().Get("slPips"), 64)

	if symbol == "" || riskPercent <= 0 || slPips <= 0 {
		http.Error(w, "Missing required parameters: symbol, riskPercent, slPips", http.StatusBadRequest)
		return
	}

	result, err := s.riskCalculator.CalculateLotFromRisk(riskPercent, slPips, symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleMarginPreview previews margin requirements
func (s *Server) HandleMarginPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	symbol := r.URL.Query().Get("symbol")
	volume, _ := strconv.ParseFloat(r.URL.Query().Get("volume"), 64)
	side := r.URL.Query().Get("side")

	if symbol == "" || volume <= 0 {
		http.Error(w, "Missing required parameters: symbol, volume", http.StatusBadRequest)
		return
	}

	// Get current margin from OANDA
	var currentMargin, freeMargin float64
	leverage := 50

	if s.hub != nil {
		// OANDA logic removed
	}

	result, err := s.riskCalculator.PreviewMargin(symbol, volume, side, leverage, currentMargin, freeMargin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleGetAccountInfo returns detailed account information
func (s *Server) HandleGetAccountInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Legacy OANDA logic removed
	http.Error(w, "No LP connection", http.StatusServiceUnavailable)
}

func (s *Server) HandleGetAccount(w http.ResponseWriter, r *http.Request) {
	s.HandleGetAccountInfo(w, r)
}

func (s *Server) HandleGetPositions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Legacy OANDA logic removed
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]interface{}{})
}

func (s *Server) HandleClosePosition(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req struct {
		TradeID string `json:"tradeId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Legacy OANDA logic removed
	http.Error(w, "No LP connection", http.StatusServiceUnavailable)
}

func (s *Server) HandleGetTicks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Missing symbol parameter", http.StatusBadRequest)
		return
	}

	limit := 500
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if s.tickStore == nil {
		http.Error(w, "Tick store not initialized", http.StatusServiceUnavailable)
		return
	}

	ticks := s.tickStore.GetHistory(symbol, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticks)
}

func (s *Server) HandleGetOHLC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "Missing symbol parameter", http.StatusBadRequest)
		return
	}

	timeframe := int64(60)
	if tfStr := r.URL.Query().Get("timeframe"); tfStr != "" {
		switch tfStr {
		case "1m":
			timeframe = 60
		case "5m":
			timeframe = 300
		case "15m":
			timeframe = 900
		case "1h":
			timeframe = 3600
		case "4h":
			timeframe = 14400
		case "1d":
			timeframe = 86400
		default:
			if tf, err := strconv.ParseInt(tfStr, 10, 64); err == nil && tf > 0 {
				timeframe = tf
			}
		}
	}

	limit := 500
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if s.tickStore == nil {
		http.Error(w, "Tick store not initialized", http.StatusServiceUnavailable)
		return
	}

	ohlc := s.tickStore.GetOHLC(symbol, timeframe, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ohlc)
}

func (s *Server) HandleGetRoutes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	rules := s.smartRouter.GetRules()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// HandleLPStatus returns the status of LPs (Legacy - use /admin/lp-status)
func (s *Server) HandleLPStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Legacy endpoint compatibility
	// The new endpoint is handled by handlers.LPHandler
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "managed_by_lp_manager",
		"info":   "Use /admin/lp-status for detailed info",
	})
}

func (s *Server) ConnectToLP(sessionID string) error {
	return s.fixGateway.Connect(sessionID)
}

func (s *Server) DisconnectLP(sessionID string) error {
	return s.fixGateway.Disconnect(sessionID)
}

func (s *Server) GetFIXStatus() map[string]string {
	return s.fixGateway.GetStatus()
}
