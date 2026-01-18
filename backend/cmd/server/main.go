package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/admin"
	"github.com/epic1st/rtx/backend/api"
	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/cbook"
	"github.com/epic1st/rtx/backend/config"
	"github.com/epic1st/rtx/backend/fix"
	"github.com/epic1st/rtx/backend/internal/alerts"
	"github.com/epic1st/rtx/backend/internal/api/handlers"
	"github.com/epic1st/rtx/backend/internal/api/websocket"
	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/lpmanager/adapters"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
)

type BrokerConfig struct {
	BrokerName        string          `json:"brokerName"`
	BrokerDisplayName string          `json:"brokerDisplayName"` // For UI display
	PriceFeedLP       string          `json:"priceFeedLP"`
	PriceFeedName     string          `json:"priceFeedName"` // Display name for price feed
	ExecutionMode     string          `json:"executionMode"`
	DefaultLeverage   int             `json:"defaultLeverage"`
	DefaultBalance    float64         `json:"defaultBalance"`
	MarginMode        string          `json:"marginMode"`
	MaxTicksPerSymbol int             `json:"maxTicksPerSymbol"`
	DisabledSymbols   map[string]bool `json:"disabledSymbols"`
}

// Global broker configuration - loaded from config
var brokerConfig BrokerConfig

// For backward compatibility
var executionMode string

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize broker config from loaded configuration
	brokerConfig = BrokerConfig{
		BrokerName:        cfg.Broker.Name,
		BrokerDisplayName: cfg.Broker.DisplayName,
		PriceFeedLP:       cfg.Broker.PriceFeedLP,
		PriceFeedName:     cfg.Broker.PriceFeedName,
		ExecutionMode:     cfg.Broker.ExecutionMode,
		DefaultLeverage:   cfg.Broker.DefaultLeverage,
		DefaultBalance:    cfg.Broker.DefaultBalance,
		MarginMode:        cfg.Broker.MarginMode,
		MaxTicksPerSymbol: cfg.Broker.MaxTicksPerSymbol,
		DisabledSymbols:   make(map[string]bool),
	}
	executionMode = cfg.Broker.ExecutionMode

	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Printf("║          %s - Backend v3.0                ║", brokerConfig.BrokerName)
	log.Printf("║        %s Mode + %s LP                 ║", brokerConfig.ExecutionMode, brokerConfig.PriceFeedLP)
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	// Initialize tick storage using config
	tickStore := tickstore.NewTickStore("default", brokerConfig.MaxTicksPerSymbol)

	// Initialize B-Book engine
	bbookEngine := core.NewEngine()

	// Initialize P/L engine
	pnlEngine := core.NewPnLEngine(bbookEngine)

	// Initialize C-Book routing engine
	cbookEngine := cbook.NewCBookEngine()

	// Create B-Book API handlers
	apiHandler := handlers.NewAPIHandler(bbookEngine, pnlEngine)
	apiHandler.SetCBookEngine(cbookEngine)

	// Create Compliance Handler
	complianceHandler := handlers.NewComplianceHandler(bbookEngine)

	// Create Auth Service with admin credentials and JWT secret from config
	authService := auth.NewService(bbookEngine, cfg.Admin.Password, cfg.JWT.Secret)

	// Create demo account with configured balance (only if configured)
	if brokerConfig.DefaultBalance > 0 {
		demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
		bbookEngine.GetLedger().SetBalance(demoAccount.ID, brokerConfig.DefaultBalance)
		demoAccount.Balance = brokerConfig.DefaultBalance
		log.Printf("[B-Book] Demo account created: %s with $%.2f", demoAccount.AccountNumber, brokerConfig.DefaultBalance)
	}
	hub := ws.NewHub()

	// Set tick store on hub for storing incoming ticks
	hub.SetTickStore(tickStore)

	// Set B-Book engine on hub for dynamic symbol registration
	hub.SetBBookEngine(bbookEngine)

	// Set auth service on hub for WebSocket authentication
	hub.SetAuthService(authService)

	// Initialize Analytics WebSocket Hub
	analyticsHub := websocket.InitializeAnalyticsHub(authService)
	log.Println("[Analytics] Real-time analytics WebSocket hub initialized")

	apiHandler.SetHub(hub)

	// Wire B-Book engine to get prices from market data
	bbookEngine.SetPriceCallback(func(symbol string) (bid, ask float64, ok bool) {
		tick := hub.GetLatestPrice(symbol)
		if tick != nil {
			return tick.Bid, tick.Ask, true
		}
		return 0, 0, false
	})

	// Initialize LP Manager
	lpMgr := lpmanager.NewManager("data/lp_config.json")

	// Register Adapters with credentials from config
	if cfg.LP.BinanceAPIKey != "" {
		lpMgr.RegisterAdapter(adapters.NewBinanceAdapter())
		log.Println("[LP] Binance adapter registered")
	}
	if cfg.LP.OandaAPIKey != "" && cfg.LP.OandaAccountID != "" {
		lpMgr.RegisterAdapter(adapters.NewOANDAAdapter(cfg.LP.OandaAPIKey, cfg.LP.OandaAccountID))
		log.Println("[LP] OANDA adapter registered")
	} else {
		log.Println("[LP WARNING] OANDA credentials not configured - OANDA adapter disabled")
	}

	// Load Config
	if err := lpMgr.LoadConfig(); err != nil {
		log.Printf("[LPManager] Failed to load config: %v", err)
	}

	// Initialize HTTP server with dependencies (pass lpMgr for A-Book)
	server := api.NewServer(authService, apiHandler, lpMgr)

	// Set tick store on server for API access
	server.SetTickStore(tickStore)

	// Pass hub to server
	server.SetHub(hub)

	// Start WebSocket hub
	go hub.Run()

	// Start LP Manager Aggregation
	lpMgr.StartQuoteAggregation()

	// Connect all enabled LPs - REMOVED (Handled internally by StartQuoteAggregation)

	// Pipe quotes from LP Manager to Hub
	go func() {
		var quoteCount int64 = 0
		for quote := range lpMgr.GetQuotesChan() {
			quoteCount++
			if quoteCount%1000 == 1 {
				log.Printf("[Main] Piping quote #%d to Hub: %s @ %.5f", quoteCount, quote.Symbol, quote.Bid)
			}
			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    quote.Symbol,
				Bid:       quote.Bid,
				Ask:       quote.Ask,
				Spread:    quote.Ask - quote.Bid,
				Timestamp: quote.Timestamp,
				LP:        quote.LP,
			}
			hub.BroadcastTick(tick)
		}
		log.Println("[Main] Quote pipe closed!")
	}()

	// ============================================
	// Initialize Alert System
	// ============================================
	log.Println("[AlertSystem] Initializing intelligent alerting engine...")

	// Create WebSocket alert broadcaster
	wsAlertHub := alerts.NewWSAlertHub(hub)

	// Create notification dispatcher
	notifier := alerts.NewNotifier(wsAlertHub)

	// Create metrics adapter to connect alerts to B-Book engine
	metricsAdapter := alerts.NewBBookMetricsAdapter(bbookEngine, pnlEngine)

	// Create alert engine
	alertEngine := alerts.NewEngine(metricsAdapter, notifier)

	// Create alert API handlers
	alertsHandler := handlers.NewAlertsHandler(alertEngine)

	// Start alert engine (5-second evaluation loop)
	alertEngine.Start()

	// Start notification workers
	notifier.Start()

	log.Println("[AlertSystem] Alert engine started - evaluating every 5 seconds")

	// ============================================
	// Initialize Admin System
	// ============================================
	adminHandler := admin.NewAdminHandler(bbookEngine)
	log.Println("[Admin] Admin system initialized")

	// Initialize FIX Provisioning (optional)
	if cfg.FIX.ProvisioningEnabled {
		// Create audit logger
		auditLogger := &fix.SimpleAuditLogger{}

		// Create FIX provisioning service
		fixProvisioning, err := fix.NewProvisioningService(
			cfg.FIX.ProvisioningStorePath,
			cfg.FIX.MasterPassword,
			auditLogger,
		)
		if err != nil {
			log.Printf("[FIX] Failed to initialize provisioning service: %v", err)
		} else {
			// Create FIX manager with provisioning service
			_ = admin.NewFIXManager(fixProvisioning) // FIX manager created but routes not registered yet
			log.Println("[FIX] FIX provisioning system initialized")
		}
	}

	// ============================================
	// REGISTER API ROUTES
	// ============================================

	// Health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Swagger API Documentation
	http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger-ui.html")
	})
	http.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "swagger.yaml")
	})

	// ===== DYNAMIC BROKER CONFIGURATION API =====
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			// Return current config
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(brokerConfig)
			return
		}

		if r.Method == "POST" {
			// Update config from admin
			var newConfig BrokerConfig
			if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Apply non-empty values
			if newConfig.BrokerName != "" {
				brokerConfig.BrokerName = newConfig.BrokerName
			}
			if newConfig.PriceFeedLP != "" {
				brokerConfig.PriceFeedLP = newConfig.PriceFeedLP
			}
			if newConfig.ExecutionMode != "" {
				brokerConfig.ExecutionMode = newConfig.ExecutionMode
				executionMode = newConfig.ExecutionMode // Sync legacy variable
			}
			if newConfig.DefaultLeverage > 0 {
				brokerConfig.DefaultLeverage = newConfig.DefaultLeverage
			}
			if newConfig.DefaultBalance > 0 {
				brokerConfig.DefaultBalance = newConfig.DefaultBalance
			}
			if newConfig.MarginMode != "" {
				brokerConfig.MarginMode = newConfig.MarginMode
			}

			log.Printf("[ADMIN] Config updated: %+v", brokerConfig)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"config":  brokerConfig,
			})
			return
		}
	})

	// Auth
	http.HandleFunc("/login", server.HandleLogin)

	// ===== B-BOOK API (RTX Internal) =====
	// These use our internal balance/equity, NOT OANDA

	// Account
	// Routing Preview (non-executing routing decision API)
	http.HandleFunc("/api/routing/preview", apiHandler.HandleRoutingPreview)

	// ===== ROUTING ANALYTICS =====
	// Analytics endpoints for routing metrics
	http.HandleFunc("/api/analytics/routing/breakdown", apiHandler.HandleRoutingBreakdown)
	http.HandleFunc("/api/analytics/routing/timeline", apiHandler.HandleRoutingTimeline)
	http.HandleFunc("/api/analytics/routing/confidence", apiHandler.HandleRoutingConfidence)

	// ===== COMPLIANCE & REGULATORY REPORTING =====
	// MiFID II RTS 27/28 - Best Execution Reporting
	http.HandleFunc("/api/compliance/best-execution", complianceHandler.HandleBestExecution)

	// SEC Rule 606 - Order Routing Disclosure
	http.HandleFunc("/api/compliance/order-routing", complianceHandler.HandleOrderRouting)

	// Audit Trail Export (7-year retention)
	http.HandleFunc("/api/compliance/audit-trail", complianceHandler.HandleAuditTrail)

	// Internal Audit Logging (WORM pattern)
	http.HandleFunc("/api/compliance/audit-log", complianceHandler.HandleAuditLog)

	// ===== ROUTING RULES MANAGEMENT =====
	// Routing Rules CRUD endpoints
	http.HandleFunc("/api/routing/rules", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			apiHandler.HandleListRoutingRules(w, r)
			return
		}

		if r.Method == "POST" {
			apiHandler.HandleCreateRoutingRule(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/routing/rules/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "PUT" {
			apiHandler.HandleUpdateRoutingRule(w, r)
			return
		}

		if r.Method == "DELETE" {
			apiHandler.HandleDeleteRoutingRule(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/routing/rules/reorder", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "POST" {
			apiHandler.HandleReorderRoutingRules(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// ===== ANALYTICS API - RULE EFFECTIVENESS =====
	// Rule effectiveness metrics endpoints
	http.HandleFunc("/api/analytics/rules/effectiveness", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			apiHandler.HandleGetRuleEffectiveness(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/analytics/rules/calculate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "POST" {
			apiHandler.HandleCalculateMetrics(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	// Note: This must be registered AFTER /api/analytics/rules/effectiveness to avoid path conflicts
	http.HandleFunc("/api/analytics/rules/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			apiHandler.HandleGetRuleMetrics(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/account/summary", apiHandler.HandleGetAccountSummary)
	http.HandleFunc("/api/account/create", apiHandler.HandleCreateAccount)

	// Positions (B-Book)
	http.HandleFunc("/api/symbols", apiHandler.HandleGetSymbols)
	http.HandleFunc("/api/positions", apiHandler.HandleGetPositions)
	http.HandleFunc("/api/positions/close", apiHandler.HandleClosePosition)
	http.HandleFunc("/api/positions/close-bulk", apiHandler.HandleCloseBulk)

	// Orders (B-Book)
	http.HandleFunc("/api/orders", apiHandler.HandleGetOrders)
	http.HandleFunc("/api/orders/market", apiHandler.HandlePlaceMarketOrder)

	// Trades & Ledger
	http.HandleFunc("/api/trades", apiHandler.HandleGetTrades)
	http.HandleFunc("/api/ledger", apiHandler.HandleGetLedger)

	// Position Management
	http.HandleFunc("/api/positions/modify", apiHandler.HandleModifyPosition)

	// ===== ALERT ENDPOINTS =====
	// Alert management API
	http.HandleFunc("/api/alerts", alertsHandler.HandleListAlerts)
	http.HandleFunc("/api/alerts/acknowledge", alertsHandler.HandleAcknowledgeAlert)
	http.HandleFunc("/api/alerts/snooze", alertsHandler.HandleSnoozeAlert)
	http.HandleFunc("/api/alerts/resolve", alertsHandler.HandleResolveAlert)

	// Alert rules management
	http.HandleFunc("/api/alerts/rules", alertsHandler.HandleListRules)
	http.HandleFunc("/api/alerts/rules/create", alertsHandler.HandleCreateRule)

	// Individual rule operations (must be after /api/alerts/rules to avoid conflicts)
	http.HandleFunc("/api/alerts/rules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
			w.WriteHeader(http.StatusOK)
			return
		}

		switch r.Method {
		case "GET":
			alertsHandler.HandleGetRule(w, r)
		case "PUT":
			alertsHandler.HandleUpdateRule(w, r)
		case "DELETE":
			alertsHandler.HandleDeleteRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("[AlertSystem] Alert API routes registered")

	// Analytics - Exposure Heatmap
	http.HandleFunc("/api/analytics/exposure/heatmap", apiHandler.HandleExposureHeatmap)
	http.HandleFunc("/api/analytics/exposure/current", apiHandler.HandleCurrentExposure)
	http.HandleFunc("/api/analytics/exposure/history/", apiHandler.HandleExposureHistory)

	// ===== ADMIN ENDPOINTS =====
	// For deposit/withdraw/adjust (Super	// Admin Endpoints
	http.HandleFunc("/admin/accounts", apiHandler.HandleAdminGetAccounts)
	http.HandleFunc("/admin/deposit", apiHandler.HandleAdminDeposit)
	http.HandleFunc("/admin/withdraw", apiHandler.HandleAdminWithdraw)
	http.HandleFunc("/admin/adjust", apiHandler.HandleAdminAdjust)
	http.HandleFunc("/admin/bonus", apiHandler.HandleAdminBonus)
	http.HandleFunc("/admin/ledger", apiHandler.HandleAdminGetLedgerAll)
	http.HandleFunc("/admin/reset-password", apiHandler.HandleAdminResetPassword)
	http.HandleFunc("/admin/account/update", apiHandler.HandleAdminUpdateAccount)
	http.HandleFunc("/admin/symbols", apiHandler.HandleAdminGetSymbols)
	http.HandleFunc("/admin/symbols/toggle", apiHandler.HandleAdminToggleSymbol)
	http.HandleFunc("/api/admin/symbols/", apiHandler.HandleAdminUpdateSymbol)

	// Execution Mode Toggle (A-Book vs B-Book)
	http.HandleFunc("/admin/execution-mode", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mode": executionMode,
				"description": map[string]string{
					"BBOOK": "Internal execution - orders processed by RTX engine using internal balance",
					"ABOOK": "LP passthrough - orders routed to OANDA (requires active LP connection)",
				},
				"priceFeed": "OANDA", // Always OANDA for prices
			})
			return
		}

		if r.Method == "POST" {
			var req struct {
				Mode string `json:"mode"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.Mode != "BBOOK" && req.Mode != "ABOOK" {
				http.Error(w, "mode must be BBOOK or ABOOK", http.StatusBadRequest)
				return
			}

			oldMode := executionMode
			executionMode = req.Mode
			log.Printf("[ADMIN] Execution mode changed: %s → %s", oldMode, executionMode)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"oldMode": oldMode,
				"newMode": executionMode,
				"message": "Execution mode updated. Price feed remains connected to OANDA.",
			})
			return
		}
	})

	// ===== LEGACY ENDPOINTS (OANDA passthrough) =====
	// Keep for compatibility but prefer /api/ routes

	http.HandleFunc("/order", server.HandlePlaceOrder) // OANDA
	http.HandleFunc("/order/limit", server.HandlePlaceLimitOrder)
	http.HandleFunc("/order/stop", server.HandlePlaceStopOrder)
	http.HandleFunc("/order/stop-limit", server.HandlePlaceStopLimitOrder)
	http.HandleFunc("/orders/pending", server.HandleGetPendingOrders)
	http.HandleFunc("/order/cancel", server.HandleCancelOrder)

	// OANDA account (legacy)
	http.HandleFunc("/account", server.HandleGetAccount) // Shows OANDA balance
	http.HandleFunc("/account/info", server.HandleGetAccountInfo)
	http.HandleFunc("/positions", server.HandleGetPositions) // OANDA positions
	http.HandleFunc("/position/close", server.HandleClosePosition)
	http.HandleFunc("/position/partial-close", server.HandlePartialClose)
	http.HandleFunc("/position/close-all", server.HandleCloseAll)
	http.HandleFunc("/position/modify", server.HandleModifySLTP)
	http.HandleFunc("/position/breakeven", server.HandleBreakeven)
	http.HandleFunc("/position/trailing-stop", server.HandleSetTrailingStop)

	// Risk Calculator
	http.HandleFunc("/risk/calculate-lot", server.HandleCalculateLot)
	http.HandleFunc("/risk/margin-preview", server.HandleMarginPreview)

	// Market Data
	http.HandleFunc("/ticks", server.HandleGetTicks)
	http.HandleFunc("/ohlc", server.HandleGetOHLC)

	// Admin (legacy)
	http.HandleFunc("/admin/routes", server.HandleGetRoutes)

	// ===== NEW ADMIN SYSTEM =====
	// Register comprehensive admin routes
	adminHandler.RegisterRoutes(http.DefaultServeMux)
	log.Println("[Admin] Comprehensive admin system routes registered")
	/*
		// Initialize LP Manager (Moved to top)
		// lpMgr := lpmanager.NewManager("data/lp_config.json")
		// ... registration moved
	*/

	// Create LP Handler
	lpHandler := handlers.NewLPHandler(lpMgr)

	// ===== ADMIN LP MANAGEMENT ENDPOINTS (v1 - /admin/lps) =====
	http.HandleFunc("/admin/lps", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			lpHandler.HandleListLPs(w, r)
		} else if r.Method == "POST" {
			lpHandler.HandleAddLP(w, r)
		} else {
			// Options
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
	})

	http.HandleFunc("/admin/lps/", func(w http.ResponseWriter, r *http.Request) {
		// Handle subpaths like /admin/lps/{id}/toggle
		if os.Getenv("DEBUG") == "true" {
			log.Printf("LP Request: %s %s", r.Method, r.URL.Path)
		}

		if len(r.URL.Path) > len("/admin/lps/") {
			// suffix := r.URL.Path[len("/admin/lps/"):]

			if strings.HasSuffix(r.URL.Path, "/toggle") {
				lpHandler.HandleToggleLP(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/symbols") {
				lpHandler.HandleLPSymbols(w, r)
				return
			}
			// If it's just ID, it's Update or Delete
			if r.Method == "PUT" {
				lpHandler.HandleUpdateLP(w, r)
				return
			}
			if r.Method == "DELETE" {
				lpHandler.HandleDeleteLP(w, r)
				return
			}
		}
	})

	http.HandleFunc("/admin/lp-status", lpHandler.HandleLPStatus)

	// ===== ADMIN LP MANAGEMENT ENDPOINTS (v2 - /api/admin/lp) =====
	// GET /api/admin/liquidity-providers - List all LPs with status
	http.HandleFunc("/api/admin/liquidity-providers", lpHandler.HandleAdminLiquidityProviders)

	// POST /api/admin/lp/{name}/toggle - Enable/disable LP by name
	http.HandleFunc("/api/admin/lp/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/toggle") {
			lpHandler.HandleToggleLPByName(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/subscriptions") {
			if r.Method == "GET" {
				lpHandler.HandleGetLPSubscriptions(w, r)
				return
			}
			if r.Method == "PUT" {
				lpHandler.HandleUpdateLPSubscriptions(w, r)
				return
			}
		}
	})

	// ===== FIX SESSION MANAGEMENT =====
	// FIX Session Status
	http.HandleFunc("/admin/fix/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		status := make(map[string]interface{})
		status["sessions"] = server.GetFIXStatus()
		json.NewEncoder(w).Encode(status)
	})

	// Connect FIX Session
	http.HandleFunc("/admin/fix/connect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			SessionID string `json:"sessionId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := server.ConnectToLP(req.SessionID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"sessionId": req.SessionID,
			"message":   "Connection initiated",
		})
	})

	// Disconnect FIX Session
	http.HandleFunc("/admin/fix/disconnect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			SessionID string `json:"sessionId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := server.DisconnectLP(req.SessionID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"sessionId": req.SessionID,
			"message":   "Disconnected",
		})
	})

	// Auto-Connect YOFX1 FIX Session on startup
	go func() {
		time.Sleep(3 * time.Second) // Wait for other services to initialize
		log.Println("[FIX] Auto-connecting YOFX1 session...")
		if err := server.ConnectToLP("YOFX1"); err != nil {
			log.Printf("[FIX] Failed to auto-connect YOFX1: %v", err)
		}
	}()

	// Backend restart endpoint (graceful)
	http.HandleFunc("/admin/restart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Println("[Admin] Backend restart requested from Admin Panel")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Restart initiated. Server will restart in 2 seconds.",
		})

		// Graceful restart - exit and let process manager (systemd, pm2, etc.) restart
		go func() {
			time.Sleep(2 * time.Second)
			log.Println("[Admin] Performing graceful shutdown for restart...")
			os.Exit(0)
		}()
	})

	// WebSocket for real-time prices AND account updates
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	// WebSocket for B-Book account updates
	http.HandleFunc("/ws/account", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement account-specific WebSocket
		ws.ServeWs(hub, w, r)
	})

	// WebSocket for analytics (routing metrics, LP performance, exposure, alerts)
	websocket.RegisterAnalyticsRoutes(analyticsHub, nil)

	log.Println("")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  SERVER READY - B-BOOK TRADING ENGINE")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  HTTP API:    http://localhost:7999")
	log.Println("  WebSocket:   ws://localhost:7999/ws")
	log.Println("")
	log.Println("  B-BOOK API (RTX Internal Balance):")
	log.Println("    GET  /api/account/summary   - RTX Balance/Equity/Margin")
	log.Println("    GET  /api/positions         - RTX Open Positions")
	log.Println("    POST /api/orders/market     - Execute Market Order")
	log.Println("    POST /api/positions/close   - Close Position")
	log.Println("    GET  /api/trades            - Trade History")
	log.Println("    GET  /api/ledger            - Transaction History")
	log.Println("")
	log.Println("  ANALYTICS API:")
	log.Println("    GET  /api/analytics/exposure/heatmap        - Exposure Heatmap Data")
	log.Println("    GET  /api/analytics/exposure/current        - Current Exposure by Symbol")
	log.Println("    GET  /api/analytics/exposure/history/{sym}  - Symbol Exposure Timeline")
	log.Println("")
	log.Println("  ADMIN ENDPOINTS:")
	log.Println("    GET  /admin/accounts        - List All Accounts")
	log.Println("    POST /admin/deposit         - Add Funds (Bank/Crypto)")
	log.Println("    POST /admin/withdraw        - Withdraw Funds")
	log.Println("    POST /admin/adjust          - Manual Adjustment")
	log.Println("    POST /admin/bonus           - Add Bonus")
	log.Println("    GET  /admin/ledger          - View All Transactions")
	log.Println("")
	if brokerConfig.DefaultBalance > 0 {
		log.Printf("  Demo Account: Demo User | Balance: $%.2f", brokerConfig.DefaultBalance)
	}
	log.Println("═══════════════════════════════════════════════════════════")

	port := ":" + cfg.Port
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
