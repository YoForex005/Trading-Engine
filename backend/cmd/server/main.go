package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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
	"github.com/epic1st/rtx/backend/internal/compression"
	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/internal/middleware"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/lpmanager/adapters"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
	"github.com/joho/godotenv"
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

// Global tick tracker for debugging market data flow
var (
	latestTicks    = make(map[string]*ws.MarketTick)
	tickMutex      sync.RWMutex
	totalTickCount int64
)

func main() {
	// ============================================
	// GC TUNING - Prevents memory crashes during high-frequency quote processing
	// ============================================
	// Load .env explicitly
	if err := godotenv.Load(); err != nil {
		log.Printf("[WARN] No .env file found: %v", err)
	} else {
		log.Printf("[INIT] Loaded .env file. YOFX_PROXY_HOST=%s", os.Getenv("YOFX_PROXY_HOST"))
	}

	// GC TUNING - Prevents memory crashes during high-frequency quote processing
	// GOGC=50: More frequent, shorter GC pauses (default 100)
	// GOMEMLIMIT=2GiB: Hard cap prevents OOM crashes
	if os.Getenv("GOGC") == "" {
		os.Setenv("GOGC", "50")
		log.Println("[GC] Set GOGC=50 for more frequent garbage collection")
	}
	if os.Getenv("GOMEMLIMIT") == "" {
		os.Setenv("GOMEMLIMIT", "2GiB")
		log.Println("[GC] Set GOMEMLIMIT=2GiB to prevent OOM crashes")
	}

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// ============================================
	// PRODUCTION VALIDATION - CRITICAL SAFEGUARD
	// ============================================
	// Prevents simulation code from running in production
	if cfg.Environment == "production" {
		log.Println("[PRODUCTION] Running production environment validation...")
		validator := config.NewProductionValidator(cfg)
		if err := validator.ValidateProduction(); err != nil {
			log.Fatalf("❌ PRODUCTION VALIDATION FAILED: %v\nServer will not start with invalid production configuration.", err)
		}
		log.Println("[PRODUCTION] ✓ Production validation passed - environment is production-safe")
	} else {
		log.Printf("[ENVIRONMENT] Running in %s mode (production validation skipped)", cfg.Environment)
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

	// Initialize OPTIMIZED tick storage with SQLite backend:
	// - Ring buffers (bounded memory, O(1) operations)
	// - Quote throttling (skip < 0.001% price changes)
	// - Async batch writer (non-blocking disk persistence)
	// - SQLite persistent storage with daily rotation
	tickStoreConfig := tickstore.ProductionConfig("BROKER-001")
	tickStore := tickstore.NewOptimizedTickStoreWithConfig(tickStoreConfig)

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

	// Initialize Compression Service
	var compressor *compression.Compressor
	if retentionCfg, err := compression.LoadRetentionConfig("backend/config/retention.yaml"); err != nil {
		log.Printf("[Compression] Failed to load retention config: %v (compression disabled)", err)
	} else {
		compressorCfg := retentionCfg.ToCompressorConfig()
		compressor = compression.NewCompressor(compressorCfg)
		if compressor.IsEnabled() {
			compressor.Start()
			log.Println("[Compression] Compression scheduler started - scans every 7 days")
		} else {
			log.Println("[Compression] Compression disabled by configuration")
		}
	}

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

			// ALSO Store in TickStore (which updates OHLC)
			// quote.Timestamp is likely int64 (ms or ns) or time.Time.
			// Assuming int64 ms based on previous code usage
			ts := time.Unix(0, quote.Timestamp*int64(time.Millisecond))
			tickStore.StoreTick(quote.Symbol, quote.Bid, quote.Ask, quote.Ask-quote.Bid, quote.LP, ts)
		}
		log.Println("[Main] Quote pipe closed!")
	}()


	// NOTE: FIX market data is piped via the dedicated goroutine below (line ~1594)
	// to avoid competing readers on the same channel with mismatched timestamp units.

	// AUTO-DISCOVERY: Request Security List when FIX logs in
	go func() {
		log.Println("[AutoDiscovery] Client started, waiting for networking...")
		time.Sleep(5 * time.Second)

		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			log.Println("[AutoDiscovery] FIX Gateway not available")
			return
		}

		discoveryDone := false
		for {
			if discoveryDone {
				time.Sleep(1 * time.Minute) // Check periodically for re-connects
			}

			status := fixGateway.GetStatus()
			targetSession := ""

			if status["YOFX2"] == "LOGGED_IN" {
				targetSession = "YOFX2"
			} else if status["YOFX1"] == "LOGGED_IN" {
				targetSession = "YOFX1"
			}

			if targetSession != "" {
				// Check if we need to request
				securities := fixGateway.GetSecurities()
				if len(securities) == 0 {
					log.Printf("[AutoDiscovery] Requesting Security List from %s...", targetSession)
					_, err := fixGateway.RequestSecurityList(targetSession)
					if err != nil {
						log.Printf("[AutoDiscovery] Failed to request security list: %v", err)
					} else {
						// Wait for response
						time.Sleep(5 * time.Second)
						securities = fixGateway.GetSecurities()
						log.Printf("[AutoDiscovery] Discovered %d securities", len(securities))
						if len(securities) > 0 {
							discoveryDone = true
						}
					}
				} else {
					discoveryDone = true
				}
			}

			time.Sleep(10 * time.Second)
		}
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
	// INITIALIZE RATE LIMITING
	// ============================================
	rateLimitConfig, err := config.LoadRateLimitingConfig()
	if err != nil {
		log.Printf("[RateLimit] Failed to load rate limiting config: %v (using defaults)", err)
		rateLimitConfig = config.RateLimitingConfig{
			Enabled:           true,
			RequestsPerSecond: 10,
			RequestsPerMinute: 500,
			BurstSize:         20,
			CleanupInterval:   "5m",
			ClientTimeout:     "10m",
		}
	}

	var rateLimiter *middleware.RateLimiter
	if rateLimitConfig.Enabled {
		// Create rate limiter with config
		rlConfig := middleware.RateLimitConfig{
			RequestsPerSecond: rateLimitConfig.RequestsPerSecond,
			RequestsPerMinute: rateLimitConfig.RequestsPerMinute,
			BurstSize:         rateLimitConfig.BurstSize,
			CleanupInterval:   config.ParseDuration(rateLimitConfig.CleanupInterval),
			ClientTimeout:     config.ParseDuration(rateLimitConfig.ClientTimeout),
		}

		rateLimiter = middleware.NewRateLimiter(rlConfig)
		log.Printf("[RateLimit] Rate limiting enabled: %.0f req/s, burst=%d", rlConfig.RequestsPerSecond, rlConfig.BurstSize)
	} else {
		log.Println("[RateLimit] Rate limiting disabled")
	}

	// ============================================
	// REGISTER API ROUTES
	// ============================================

	// Initialize History Handler
	historyHandler := api.NewHistoryHandler(tickStore)
	historyHandler.RegisterRoutes(http.DefaultServeMux)
	log.Println("[History] History API registered")

	// Initialize Drawings Handler (Requested Fix)
	drawingsHandler := api.NewDrawingsHandler()
	drawingsHandler.RegisterRoutes(http.DefaultServeMux)
	log.Println("[Drawings] Drawings API registered")

	// Health (no rate limit)
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
			// Dynamically detect active LP from FIX sessions
			activePriceFeedLP := brokerConfig.PriceFeedLP // Default from config
			activePriceFeedName := brokerConfig.PriceFeedName

			fixGateway := server.GetFIXGateway()
			if fixGateway != nil {
				status := fixGateway.GetStatus()
				// Check for YOFX2 (Market Data Feed) being logged in
				if status["YOFX2"] == "LOGGED_IN" {
					activePriceFeedLP = "YOFX"
					activePriceFeedName = "YOFX Market Data Feed"
				} else if status["YOFX1"] == "LOGGED_IN" {
					activePriceFeedLP = "YOFX"
					activePriceFeedName = "YOFX Trading Account"
				}
				// Add more LP detection logic as needed (dynamic, no hardcoding)
				for sessionID, sessionStatus := range status {
					if sessionStatus == "LOGGED_IN" && activePriceFeedLP == brokerConfig.PriceFeedLP {
						// Use the session name as the LP if still using default
						activePriceFeedLP = sessionID
						activePriceFeedName = sessionID + " (FIX)"
					}
				}
			}

			// Return config with dynamic LP info
			response := struct {
				BrokerName        string            `json:"brokerName"`
				BrokerDisplayName string            `json:"brokerDisplayName"`
				PriceFeedLP       string            `json:"priceFeedLP"`
				PriceFeedName     string            `json:"priceFeedName"`
				ExecutionMode     string            `json:"executionMode"`
				DefaultLeverage   int               `json:"defaultLeverage"`
				DefaultBalance    float64           `json:"defaultBalance"`
				MarginMode        string            `json:"marginMode"`
				MaxTicksPerSymbol int               `json:"maxTicksPerSymbol"`
				DisabledSymbols   map[string]bool   `json:"disabledSymbols"`
				FIXStatus         map[string]string `json:"fixStatus,omitempty"`
			}{
				BrokerName:        brokerConfig.BrokerName,
				BrokerDisplayName: brokerConfig.BrokerDisplayName,
				PriceFeedLP:       activePriceFeedLP,
				PriceFeedName:     activePriceFeedName,
				ExecutionMode:     brokerConfig.ExecutionMode,
				DefaultLeverage:   brokerConfig.DefaultLeverage,
				DefaultBalance:    brokerConfig.DefaultBalance,
				MarginMode:        brokerConfig.MarginMode,
				MaxTicksPerSymbol: brokerConfig.MaxTicksPerSymbol,
				DisabledSymbols:   brokerConfig.DisabledSymbols,
			}

			// Include FIX session status for debugging/monitoring
			if fixGateway != nil {
				response.FIXStatus = fixGateway.GetStatus()
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
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

	// Symbol Specification API - must be before /api/symbols/available to avoid conflicts
	http.HandleFunc("/api/symbols/", func(w http.ResponseWriter, r *http.Request) {
		// Handle /api/symbols/{symbol}/spec
		if strings.HasSuffix(r.URL.Path, "/spec") {
			server.HandleGetSymbolSpec(w, r)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	})

	// Symbol Management API (for Market Watch)
	http.HandleFunc("/api/symbols/available", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Dynamic Symbol List from FIX Gateway
		availableSymbols := make([]map[string]interface{}, 0)

		fixGateway := server.GetFIXGateway()
		if fixGateway != nil {
			securities := fixGateway.GetSecurities()
			if len(securities) > 0 {
				for _, sec := range securities {
					// LOGIC: Majors are Root Level (Empty Category), others are nested
					category := "TradingA.Other"

					// Define Majors that should be at root
					majors := map[string]bool{
						"EURUSD": true, "GBPUSD": true, "USDCHF": true,
						"USDJPY": true, "USDCAD": true, "AUDUSD": true,
					}

					if majors[sec.Symbol] {
						category = "" // Root Level
					} else if strings.Contains(sec.Symbol, "XAU") || strings.Contains(sec.Symbol, "XAG") {
						category = "TradingA.CFD-Metals"
					} else if strings.Contains(sec.Symbol, "BTC") || strings.Contains(sec.Symbol, "ETH") {
						category = "TradingA.Crypto"
					} else if strings.Contains(sec.Symbol, "US") && !strings.Contains(sec.Symbol, "USD") {
						category = "TradingA.Indices"
					} else {
						category = "TradingA.CFD-FX"
					}

					availableSymbols = append(availableSymbols, map[string]interface{}{
						"symbol":   sec.Symbol,
						"name":     sec.Symbol,
						"category": category,
						"digits":   sec.Digits,
					})
				}
				log.Printf("[API] Serving %d dynamic symbols from FIX", len(availableSymbols))
			}
		}

		// Fallback if FIX is not ready yet or returned no symbols
		if len(availableSymbols) == 0 {
			log.Println("[API] FIX symbols not ready, returning default structure for menu")
			availableSymbols = []map[string]interface{}{
				// Root Level Majors (Empty Category)
				{"symbol": "EURUSD", "name": "EURUSD", "category": "", "digits": 5},
				{"symbol": "GBPUSD", "name": "GBPUSD", "category": "", "digits": 5},
				{"symbol": "USDJPY", "name": "USDJPY", "category": "", "digits": 3},
				{"symbol": "USDCHF", "name": "USDCHF", "category": "", "digits": 5},
				{"symbol": "USDCAD", "name": "USDCAD", "category": "", "digits": 5},
				{"symbol": "AUDUSD", "name": "AUDUSD", "category": "", "digits": 5},

				// Nested Items (TradingA Folder)
				// 1. CFD-FX (Minors)
				{"symbol": "AUDNZD", "name": "AUDNZD", "category": "TradingA.CFD-FX", "digits": 5},
				{"symbol": "EURAUD", "name": "EURAUD", "category": "TradingA.CFD-FX", "digits": 5},
				{"symbol": "GBPJPY", "name": "GBPJPY", "category": "TradingA.CFD-FX", "digits": 3},

				// 2. CFD-Metals
				{"symbol": "XAUUSD", "name": "Gold vs USD", "category": "TradingA.CFD-Metals", "digits": 2},
				{"symbol": "XAGUSD", "name": "Silver vs USD", "category": "TradingA.CFD-Metals", "digits": 3},

				// 3. Crypto
				{"symbol": "BTCUSD", "name": "Bitcoin", "category": "TradingA.Crypto", "digits": 2},
				{"symbol": "ETHUSD", "name": "Ethereum", "category": "TradingA.Crypto", "digits": 2},
			}
		}

		json.NewEncoder(w).Encode(availableSymbols)
	})

	// Subscribe to a symbol (triggers FIX market data subscription)
	http.HandleFunc("/api/symbols/subscribe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Symbol string `json:"symbol"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Symbol == "" {
			http.Error(w, "Symbol is required", http.StatusBadRequest)
			return
		}

		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "FIX gateway not available",
			})
			return
		}

		// Check if already subscribed
		if fixGateway.IsSymbolSubscribed(req.Symbol) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"symbol":  req.Symbol,
				"message": "Already subscribed",
			})
			return
		}

		// Subscribe via YOFX2 (market data session)
		mdReqID, err := fixGateway.SubscribeMarketData("YOFX2", req.Symbol)
		if err != nil {
			log.Printf("[API] Symbol subscription failed for %s: %v", req.Symbol, err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"symbol":  req.Symbol,
				"error":   err.Error(),
			})
			return
		}

		log.Printf("[API] Subscribed to %s market data (MDReqID: %s)", req.Symbol, mdReqID)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"symbol":  req.Symbol,
			"mdReqId": mdReqID,
			"message": "Subscribed successfully",
		})
	})

	// Get list of currently subscribed symbols
	http.HandleFunc("/api/symbols/subscribed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			json.NewEncoder(w).Encode([]string{})
			return
		}

		subscribedSymbols := fixGateway.GetSubscribedSymbols()
		json.NewEncoder(w).Encode(subscribedSymbols)
	})

	// Unsubscribe from a symbol (removes from FIX market data subscriptions)
	http.HandleFunc("/api/symbols/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Symbol string `json:"symbol"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Symbol == "" {
			http.Error(w, "Symbol is required", http.StatusBadRequest)
			return
		}

		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "FIX gateway not available",
			})
			return
		}

		// Check if symbol is subscribed
		if !fixGateway.IsSymbolSubscribed(req.Symbol) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"symbol":  req.Symbol,
				"message": "Symbol not subscribed",
			})
			return
		}

		// Unsubscribe via YOFX2 (market data session)
		err := fixGateway.UnsubscribeMarketDataBySymbol("YOFX2", req.Symbol)
		if err != nil {
			log.Printf("[API] Symbol unsubscription failed for %s: %v", req.Symbol, err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"symbol":  req.Symbol,
				"error":   err.Error(),
			})
			return
		}

		log.Printf("[API] Unsubscribed from %s market data", req.Symbol)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"symbol":  req.Symbol,
			"message": "Unsubscribed successfully",
		})
	})

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

	// Diagnostics - Market Data Status
	http.HandleFunc("/api/diagnostics/market-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fixGateway := server.GetFIXGateway()
		diagnostics := map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		if fixGateway == nil {
			diagnostics["status"] = "unavailable"
			diagnostics["error"] = "FIX gateway not initialized"
			diagnostics["subscriptions"] = []string{}
			diagnostics["activeStreams"] = 0
			json.NewEncoder(w).Encode(diagnostics)
			return
		}

		// Get subscribed symbols
		subscribedSymbols := fixGateway.GetSubscribedSymbols()
		diagnostics["status"] = "connected"
		diagnostics["subscriptions"] = subscribedSymbols
		diagnostics["activeStreams"] = len(subscribedSymbols)

		// Get FIX session status
		fixStatus := fixGateway.GetStatus()
		diagnostics["fixSessions"] = fixStatus

		// Get tick stats from hub
		tickMutex.RLock()
		tickStats := make(map[string]interface{})
		for symbol, tick := range latestTicks {
			tickStats[symbol] = map[string]interface{}{
				"bid":       tick.Bid,
				"ask":       tick.Ask,
				"spread":    tick.Spread,
				"timestamp": tick.Timestamp,
			}
		}
		tickMutex.RUnlock()

		diagnostics["latestTicks"] = tickStats
		diagnostics["totalTicksReceived"] = totalTickCount

		// Calculate latency stats (if available)
		latencyStats := map[string]interface{}{
			"average": "N/A",
			"min":     "N/A",
			"max":     "N/A",
		}
		diagnostics["latency"] = latencyStats

		json.NewEncoder(w).Encode(diagnostics)
	})

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

	// Market Data with dynamic FIX subscription
	http.HandleFunc("/ticks", func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		if symbol != "" {
			// Check if symbol is subscribed, if not subscribe dynamically
			fixGateway := server.GetFIXGateway()
			if fixGateway != nil && !fixGateway.IsSymbolSubscribed(symbol) {
				// Subscribe to this symbol on YOFX2 (market data session)
				if _, err := fixGateway.SubscribeMarketData("YOFX2", symbol); err != nil {
					log.Printf("[FIX] Dynamic subscription for %s failed: %v", symbol, err)
				} else {
					log.Printf("[FIX] Dynamically subscribed to %s market data", symbol)
				}
			}
		}
		server.HandleGetTicks(w, r)
	})
	http.HandleFunc("/ohlc", server.HandleGetOHLC)

	// Admin (legacy)
	http.HandleFunc("/admin/routes", server.HandleGetRoutes)

	// ===== NEW ADMIN SYSTEM =====
	// Register comprehensive admin routes
	adminHandler.RegisterRoutes(http.DefaultServeMux)
	log.Println("[Admin] Comprehensive admin system routes registered")

	// Register FIX connection management endpoints
	fixConnHandler := admin.NewFIXConnectionHandler(server.GetFIXGateway(), server.GetConnectionManager())
	fixConnHandler.RegisterRoutes(http.DefaultServeMux)
	log.Println("[Admin] FIX connection management endpoints registered")
	/*
		// Initialize LP Manager (Moved to top)
		// lpMgr := lpmanager.NewManager("data/lp_config.json")
		// ... registration moved
	*/

	// Create LP Handler
	lpHandler := handlers.NewLPHandler(lpMgr)

	// ===== HISTORICAL DATA API =====
	// (Already initialized and registered above via RegisterRoutes)
	// historyHandler := api.NewHistoryHandler(tickStore)

	// Register history routes on a router (using http.DefaultServeMux for now)
	// In production, use gorilla/mux for better routing
	// CRITICAL FIX: Register query parameter endpoint for frontend
	// Frontend uses: GET /api/history/ticks?symbol=EURUSD&date=2026-01-20&limit=5000
	// http.HandleFunc("/api/history/ticks", historyHandler.HandleGetTicksQuery)

	// Path-based endpoint (legacy): GET /api/history/ticks/EURUSD
	/*
		http.HandleFunc("/api/history/ticks/", func(w http.ResponseWriter, r *http.Request) {
			// Extract symbol from path
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) >= 5 && parts[4] != "" {
				// Store symbol in mux.Vars equivalent
				r = r.WithContext(r.Context())
				historyHandler.HandleGetTicks(w, r)
			} else {
				http.Error(w, "Symbol required", http.StatusBadRequest)
			}
		})
	*/
	/*
		http.HandleFunc("/api/history/ticks/bulk", historyHandler.HandleBulkDownload)
		http.HandleFunc("/api/history/available", historyHandler.HandleGetAvailable)
		http.HandleFunc("/api/history/symbols", historyHandler.HandleGetSymbols)
		http.HandleFunc("/admin/history/backfill", historyHandler.HandleBackfill)
	*/
	log.Println("[HistoryAPI] Historical data API routes registered (via RegisterRoutes)")

	// ===== ADMIN HISTORY MANAGEMENT (Comprehensive Controls) =====
	adminHistoryHandler := api.NewAdminHistoryHandler(tickStore, authService)
	http.HandleFunc("/admin/history/stats", adminHistoryHandler.HandleGetStats)
	http.HandleFunc("/admin/history/import", adminHistoryHandler.HandleImportData)
	http.HandleFunc("/admin/history/cleanup", adminHistoryHandler.HandleCleanupOldData)
	http.HandleFunc("/admin/history/compress", adminHistoryHandler.HandleCompressData)
	http.HandleFunc("/admin/history/backup", adminHistoryHandler.HandleBackup)
	http.HandleFunc("/admin/history/monitoring", adminHistoryHandler.HandleGetMonitoring)
	log.Println("[AdminHistory] Admin data management routes registered")

	// ===== COMPRESSION MANAGEMENT ENDPOINTS =====
	// Get compression metrics
	http.HandleFunc("/admin/compression/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		if compressor == nil {
			http.Error(w, "Compression service not initialized", http.StatusServiceUnavailable)
			return
		}

		metrics := compressor.GetMetrics()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"filesCompressed":  metrics.FilesCompressed,
			"bytesOriginal":    metrics.BytesOriginal,
			"bytesCompressed":  metrics.BytesCompressed,
			"bytesSaved":       metrics.BytesOriginal - metrics.BytesCompressed,
			"compressionRatio": float64(metrics.BytesCompressed) / float64(metrics.BytesOriginal+1),
			"errorCount":       metrics.ErrorCount,
			"lastError":        metrics.LastError,
			"lastCompression":  metrics.LastCompression,
		})
	})

	// Trigger manual compression
	http.HandleFunc("/admin/compression/trigger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if compressor == nil {
			http.Error(w, "Compression service not initialized", http.StatusServiceUnavailable)
			return
		}

		if !compressor.IsEnabled() {
			http.Error(w, "Compression is disabled", http.StatusBadRequest)
			return
		}

		// Run compression in background
		go compressor.TriggerCompression()

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Compression scan triggered in background",
		})
	})

	// Compress specific file manually
	http.HandleFunc("/admin/compression/file", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if compressor == nil {
			http.Error(w, "Compression service not initialized", http.StatusServiceUnavailable)
			return
		}

		var req struct {
			FilePath string `json:"filePath"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.FilePath == "" {
			http.Error(w, "filePath is required", http.StatusBadRequest)
			return
		}

		if err := compressor.CompressFile(req.FilePath); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"filePath": req.FilePath,
			"message":  "File compressed successfully",
		})
	})

	log.Println("[Compression] Compression management endpoints registered")

	// ===== ADMIN LP MANAGEMENT ENDPOINTS (v1 - /admin/lps) =====
	http.HandleFunc("/admin/lps", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			lpHandler.HandleListLPs(w, r)
		case "POST":
			lpHandler.HandleAddLP(w, r)
		default:
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
	// Note: /admin/fix/status, /admin/fix/connect, /admin/fix/disconnect are registered via fixConnHandler.RegisterRoutes()

	// Manual FIX Subscription endpoint
	http.HandleFunc("/admin/fix/subscribe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req struct {
			SessionID string `json:"sessionId"`
			Symbol    string `json:"symbol"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			http.Error(w, "FIX gateway not available", http.StatusServiceUnavailable)
			return
		}

		mdReqID, err := fixGateway.SubscribeMarketData(req.SessionID, req.Symbol)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"sessionId": req.SessionID,
			"symbol":    req.Symbol,
			"mdReqId":   mdReqID,
		})
	})

	// Subscribe all forex symbols
	http.HandleFunc("/admin/fix/subscribe-all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			http.Error(w, "FIX gateway not available", http.StatusServiceUnavailable)
			return
		}

		// Forex symbols for testing (includes XAUUSD gold)
		forexSymbols := []string{
			"EURUSD", "GBPUSD", "USDJPY", "XAUUSD",
		}

		results := make([]map[string]interface{}, 0)
		for _, symbol := range forexSymbols {
			mdReqID, err := fixGateway.SubscribeMarketData("YOFX2", symbol)
			if err != nil {
				results = append(results, map[string]interface{}{
					"symbol": symbol,
					"error":  err.Error(),
				})
			} else {
				results = append(results, map[string]interface{}{
					"symbol":  symbol,
					"mdReqId": mdReqID,
					"success": true,
				})
			}
			time.Sleep(50 * time.Millisecond)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"subscriptions": results,
		})
	})

	// Start FIX Connection Manager for automatic reconnection and health monitoring
	connMgr := server.GetConnectionManager()
	if connMgr != nil {
		connMgr.Start()
		log.Println("[FIX] Connection manager started - monitoring connection health")
	}

	// Auto-Connect FIX Sessions on startup with automatic reconnection enabled
	go func() {
		time.Sleep(3 * time.Second) // Wait for other services to initialize

		// Enable auto-reconnect for YOFX sessions
		if connMgr != nil {
			connMgr.EnableAutoReconnect("YOFX1")
			connMgr.EnableAutoReconnect("YOFX2")
			log.Println("[FIX] Auto-reconnect enabled for YOFX1 and YOFX2 with exponential backoff (5s-5m)")
		}

		// Connect YOFX1 (Trading)
		log.Println("[FIX] Auto-connecting YOFX1 session (Trading)...")
		if err := server.ConnectToLP("YOFX1"); err != nil {
			log.Printf("[FIX] Failed to auto-connect YOFX1: %v (will auto-retry)", err)
		}

		// Connect YOFX2 (Market Data) after short delay
		time.Sleep(2 * time.Second)
		log.Println("[FIX] Auto-connecting YOFX2 session (Market Data)...")
		if err := server.ConnectToLP("YOFX2"); err != nil {
			log.Printf("[FIX] Failed to auto-connect YOFX2: %v (will auto-retry)", err)
		} else {
			// First request security list to discover available symbols
			time.Sleep(2 * time.Second)
			fixGateway := server.GetFIXGateway()
			if fixGateway != nil {
				// Request available securities from YOFX
				// log.Println("[FIX] Requesting security list from YOFX2...")
				// if _, err := fixGateway.RequestSecurityList("YOFX2"); err != nil {
				// 	log.Printf("[FIX] Failed to request security list: %v", err)
				// }

				// Wait for security list response before subscribing
				time.Sleep(2 * time.Second)

				// All major forex pairs and metals available on YOFX
				forexSymbols := []string{
					// Major pairs
					"EURUSD", "GBPUSD", "USDJPY", "USDCHF", "USDCAD",
					"AUDUSD", "NZDUSD",
					// Cross pairs
					"EURGBP", "EURJPY", "GBPJPY", "EURAUD", "EURCAD",
					"EURCHF", "AUDCAD", "AUDCHF", "AUDJPY", "AUDNZD",
					"CADCHF", "CADJPY", "CHFJPY", "GBPAUD", "GBPCAD",
					"GBPCHF", "GBPNZD", "NZDCAD", "NZDCHF", "NZDJPY",
					// Metals
					"XAUUSD", "XAGUSD",
				}
				log.Printf("[FIX] Auto-subscribing to %d forex symbols on YOFX2...", len(forexSymbols))
				for _, symbol := range forexSymbols {
					// Step 1: Request security definition (35=c) for FIX 4.4 compliance
					if _, err := fixGateway.RequestSecurityDefinition("YOFX2", symbol); err != nil {
						log.Printf("[FIX] SecurityDefinition request failed for %s: %v", symbol, err)
					} else {
						log.Printf("[FIX] SecurityDefinition requested for %s", symbol)
					}

					// Step 2: Subscribe to market data (35=V)
					if _, err := fixGateway.SubscribeMarketData("YOFX2", symbol); err != nil {
						log.Printf("[FIX] Failed to subscribe %s: %v", symbol, err)
					} else {
						log.Printf("[FIX] Subscribed to %s market data", symbol)
					}
					time.Sleep(100 * time.Millisecond) // Rate limit subscriptions
				}
			}
		}
	}()

	// Pipe FIX market data to WebSocket hub
	go func() {
		fixGateway := server.GetFIXGateway()
		if fixGateway == nil {
			log.Println("[FIX-WS] FIX gateway not available for market data pipe")
			return
		}

		var tickCount int64 = 0
		log.Println("[FIX-WS] Starting FIX market data → WebSocket hub pipe...")

		for md := range fixGateway.GetMarketData() {
			tickCount++
			if tickCount%100 == 1 {
				log.Printf("[FIX-WS] Piping FIX tick #%d: %s Bid=%.5f Ask=%.5f High24h=%.5f Low24h=%.5f",
					tickCount, md.Symbol, md.Bid, md.Ask, md.High24h, md.Low24h)
			}

			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    md.Symbol,
				Bid:       md.Bid,
				Ask:       md.Ask,
				Spread:    md.Ask - md.Bid,
				Timestamp: md.Timestamp.UnixMilli(),
				LP:        "YOFX",
				High24h:   md.High24h,
				Low24h:    md.Low24h,
			}

			// Store latest tick for debugging
			tickMutex.Lock()
			latestTicks[md.Symbol] = tick
			totalTickCount++
			tickMutex.Unlock()
			hub.BroadcastTick(tick)
		}
		log.Println("[FIX-WS] FIX market data pipe closed!")
	}()

	// ============================================
	// SIMULATION REMOVED - 2026-01-30
	// ============================================
	// Previously had hybrid simulation fallback (lines 1657-1775)
	// Reason: Using 100% real YOFX data only - simulation no longer needed
	// Backup saved to: backend/removed_simulation_backup.go.txt
	//
	// All market data now flows exclusively from:
	// 1. FIX Gateway (YOFX1 Trading + YOFX2 Market Data)
	// 2. LP Manager (Binance, OANDA aggregation)
	//
	// No simulated ticks are generated. If no real data is available,
	// the frontend will display "No data" instead of fake prices.

	// Debug endpoint to check market data flow
	http.HandleFunc("/admin/fix/ticks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		tickMutex.RLock()
		response := map[string]interface{}{
			"totalTickCount": totalTickCount,
			"symbolCount":    len(latestTicks),
			"latestTicks":    latestTicks,
		}
		tickMutex.RUnlock()

		json.NewEncoder(w).Encode(response)
	})

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
	if compressor != nil && compressor.IsEnabled() {
		cfg := compressor.GetConfig()
		log.Println("")
		log.Println("  DATA COMPRESSION:")
		log.Printf("    Enabled: Yes | Schedule: %s | Max Age: %d seconds", cfg.Schedule, cfg.MaxAgeSeconds)
		log.Printf("    Data Dir: %s | Max Concurrency: %d", cfg.DataDir, cfg.MaxConcurrency)
		log.Println("    Metrics: GET /admin/compression/metrics")
		log.Println("    Manual Trigger: POST /admin/compression/trigger")
		log.Println("    Compress File: POST /admin/compression/file")
	}
	log.Println("═══════════════════════════════════════════════════════════")

	// Gracefully shutdown services on exit
	if rateLimiter != nil {
		defer rateLimiter.Stop()
	}
	if compressor != nil && compressor.IsEnabled() {
		defer compressor.Stop()
	}

	port := ":" + cfg.Port
	log.Printf("Starting server on port %s", port)

	// Create HTTP server with rate limiting middleware
	var handler http.Handler
	if rateLimiter != nil {
		// Wrap all routes with rate limiting (except excluded paths)
		handler = rateLimiter.MiddlewareWithExclusions(rateLimitConfig.Exclusions)(http.DefaultServeMux)
	} else {
		handler = http.DefaultServeMux
	}

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}
