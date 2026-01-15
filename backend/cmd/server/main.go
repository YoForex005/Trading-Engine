package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/epic1st/rtx/backend/api"
	"github.com/epic1st/rtx/backend/auth"
	"github.com/epic1st/rtx/backend/internal/api/handlers"
	"github.com/epic1st/rtx/backend/internal/core"
	"github.com/epic1st/rtx/backend/internal/database"
	"github.com/epic1st/rtx/backend/internal/database/repository"
	"github.com/epic1st/rtx/backend/internal/health"
	"github.com/epic1st/rtx/backend/internal/logging"
	"github.com/epic1st/rtx/backend/internal/migration"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/lpmanager/adapters"
	"github.com/epic1st/rtx/backend/tickstore"
	"github.com/epic1st/rtx/backend/ws"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type BrokerConfig struct {
	BrokerName        string          `json:"brokerName"`
	PriceFeedLP       string          `json:"priceFeedLP"`   // Current price feed provider
	ExecutionMode     string          `json:"executionMode"` // BBOOK or ABOOK
	DefaultLeverage   int             `json:"defaultLeverage"`
	DefaultBalance    float64         `json:"defaultBalance"` // Demo account starting balance
	MarginMode        string          `json:"marginMode"`     // "HEDGING" or "NETTING"
	MaxTicksPerSymbol int             `json:"maxTicksPerSymbol"`
	DisabledSymbols   map[string]bool `json:"disabledSymbols"` // Map for fast lookup (true = disabled)
}

// Global broker configuration - modifiable from admin
var brokerConfig = BrokerConfig{
	BrokerName:        "RTX Trading",
	PriceFeedLP:       "OANDA", // Can be changed to other LPs
	ExecutionMode:     "BBOOK", // BBOOK (internal) or ABOOK (LP passthrough)
	DefaultLeverage:   100,
	DefaultBalance:    5000.0,
	MarginMode:        "HEDGING",
	MaxTicksPerSymbol: 50000,
}

// For backward compatibility
var executionMode = "BBOOK"

func main() {
	ctx := context.Background()

	// Initialize structured JSON logging
	logger := logging.NewLogger()
	slog.SetDefault(logger)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Warn("No .env file found, using environment variables")
	}

	// Load OANDA credentials from environment
	oandaAPIKey := os.Getenv("OANDA_API_KEY")
	oandaAccountID := os.Getenv("OANDA_ACCOUNT_ID")

	// Validate critical credentials
	if oandaAPIKey == "" {
		log.Println("[WARN] OANDA_API_KEY not set - OANDA adapter will fail to connect")
	}
	if oandaAccountID == "" {
		log.Println("[WARN] OANDA_ACCOUNT_ID not set - OANDA adapter will fail to connect")
	}

	log.Println("╔═══════════════════════════════════════════════════════════╗")
	log.Printf("║          %s - Backend v3.0                ║", brokerConfig.BrokerName)
	log.Printf("║        %s Mode + %s LP                 ║", brokerConfig.ExecutionMode, brokerConfig.PriceFeedLP)
	log.Println("╚═══════════════════════════════════════════════════════════╝")

	// 1. Initialize database connection pool
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable required")
	}

	if err := database.InitPool(ctx, dbURL); err != nil {
		log.Fatalf("Failed to initialize database pool: %v", err)
	}
	defer database.Close()

	pool := database.GetPool()
	log.Println("Database connection pool initialized")

	// 2. Create repositories
	accountRepo := repository.NewAccountRepository(pool)
	positionRepo := repository.NewPositionRepository(pool)
	orderRepo := repository.NewOrderRepository(pool)
	tradeRepo := repository.NewTradeRepository(pool)

	// 3. Run one-time data migration (safe to run on every startup - idempotent)
	jsonPath := "./data/bbook/engine_state.json"
	if err := migration.MigrateFromJSON(ctx, jsonPath, accountRepo, positionRepo, orderRepo, tradeRepo); err != nil {
		log.Fatalf("Failed to migrate data: %v", err)
	}

	// Initialize tick storage using config
	tickStore := tickstore.NewTickStore("default", brokerConfig.MaxTicksPerSymbol)

	// Initialize B-Book engine with repository dependencies
	bbookEngine := core.NewEngine(accountRepo, positionRepo, orderRepo, tradeRepo)

	// Load accounts from database into cache
	if err := bbookEngine.LoadAccounts(ctx); err != nil {
		log.Fatalf("Failed to load accounts from database: %v", err)
	}
	log.Printf("Engine initialized with accounts from database")

	// Initialize P/L engine
	pnlEngine := core.NewPnLEngine(bbookEngine)

	// Create B-Book API handlers
	apiHandler := handlers.NewAPIHandler(bbookEngine, pnlEngine)

	// Create Auth Service
	authService := auth.NewService(bbookEngine)

	// Initialize HTTP server with dependencies
	server := api.NewServer(authService, apiHandler)

	// Create demo account with configured balance
	// Password "password" is hashed internally by CreateAccount using bcrypt
	demoAccount := bbookEngine.CreateAccount("demo-user", "Demo User", "password", true)
	bbookEngine.GetLedger().SetBalance(demoAccount.ID, brokerConfig.DefaultBalance)
	demoAccount.Balance = brokerConfig.DefaultBalance
	log.Printf("[B-Book] Demo account created: %s with $%.2f", demoAccount.AccountNumber, brokerConfig.DefaultBalance)
	hub := ws.NewHub()

	// Set tick store on hub for storing incoming ticks
	hub.SetTickStore(tickStore)

	// Set B-Book engine on hub for dynamic symbol registration
	hub.SetBBookEngine(bbookEngine)
	apiHandler.SetHub(hub)

	// Set tick store on server for API access
	server.SetTickStore(tickStore)

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

	// Register Adapters
	lpMgr.RegisterAdapter(adapters.NewBinanceAdapter())
	if oandaAPIKey != "" && oandaAccountID != "" {
		lpMgr.RegisterAdapter(adapters.NewOANDAAdapter(oandaAPIKey, oandaAccountID))
		log.Println("[LP Manager] OANDA adapter registered")
	} else {
		log.Println("[LP Manager] OANDA adapter skipped (credentials not configured)")
	}
	lpMgr.RegisterAdapter(adapters.NewFlexyAdapter())

	// Load Config
	if err := lpMgr.LoadConfig(); err != nil {
		log.Printf("[LPManager] Failed to load config: %v", err)
	}

	// Load LP priorities from config and set them on Hub
	if config := lpMgr.GetConfig(); config != nil {
		for _, lp := range config.LPs {
			hub.SetLPPriority(lp.ID, lp.Priority)
		}
		log.Printf("[Main] Loaded LP priorities: Binance=%d, OANDA=%d, FlexyMarkets=%d",
			config.LPs[0].Priority, config.LPs[1].Priority, config.LPs[2].Priority)
	}

	// Pass hub to server
	server.SetHub(hub)

	// Start WebSocket hub
	go hub.Run()

	// Start LP Manager Aggregation
	lpMgr.StartQuoteAggregation()

	// Periodically sync symbols from enabled LPs to Engine
	go func() {
		// Initial sync after a short delay to allow connections
		time.Sleep(2 * time.Second)

		syncTicker := time.NewTicker(30 * time.Second)
		for {
			adapters := lpMgr.GetEnabledAdapters()
			for _, adapter := range adapters {
				if syms, err := adapter.GetSymbols(); err == nil {
					for _, s := range syms {
						// Register/Update in engine
						existing := bbookEngine.GetSymbol(s.Symbol)
						var spec *core.SymbolSpec

						if existing != nil {
							spec = existing
						} else {
							// Default values if missing
							contractSize := 100000.0
							if s.Type == "crypto" {
								contractSize = 1.0
							}
							spec = &core.SymbolSpec{
								Symbol:           s.Symbol,
								ContractSize:     contractSize,
								PipSize:          s.PipValue,
								MinVolume:        s.MinLotSize,
								MaxVolume:        s.MaxLotSize,
								VolumeStep:       s.LotStep,
								MarginPercent:    1.0,
								CommissionPerLot: 0.0,
								Disabled:         false,
								AvailableLPs:     []string{},
							}
						}

						// Update AvailableLPs
						found := false
						for _, lp := range spec.AvailableLPs {
							if lp == adapter.ID() {
								found = true
								break
							}
						}
						if !found {
							spec.AvailableLPs = append(spec.AvailableLPs, adapter.ID())
						}

						// Set Default SourceLP if empty
						if spec.SourceLP == "" {
							spec.SourceLP = adapter.ID()
						}

						bbookEngine.UpdateSymbol(spec)
					}
				}
			}
			<-syncTicker.C
		}
	}()

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
	// REGISTER API ROUTES
	// ============================================

	// Observability endpoints
	http.HandleFunc("/health/live", health.LivenessHandler)
	http.HandleFunc("/health/ready", health.ReadinessHandler(pool))
	http.Handle("/metrics", promhttp.Handler())

	// Health (legacy - keep for backward compatibility)
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
	http.HandleFunc("/api/positions/sl-tp", apiHandler.HandleSetPositionSLTP) // New: Set SL/TP orders

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
	http.HandleFunc("/admin/symbols/source", apiHandler.HandleAdminUpdateSymbolSource)

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
	/*
		// Initialize LP Manager (Moved to top)
		// lpMgr := lpmanager.NewManager("data/lp_config.json")
		// ... registration moved
	*/

	// Create LP Handler
	lpHandler := handlers.NewLPHandler(lpMgr)

	// Admin LP Management Endpoints
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

	slog.Info("server ready", "mode", "B-BOOK")
	slog.Info("http api", "url", "http://localhost:8080")
	slog.Info("websocket", "url", "ws://localhost:8080/ws")
	slog.Info("observability endpoints ready",
		"metrics", "http://localhost:8080/metrics",
		"liveness", "http://localhost:8080/health/live",
		"readiness", "http://localhost:8080/health/ready")
	slog.Info("demo account created",
		"account_number", demoAccount.AccountNumber,
		"balance", brokerConfig.DefaultBalance)

	log.Println("")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  SERVER READY - B-BOOK TRADING ENGINE")
	log.Println("═══════════════════════════════════════════════════════════")
	log.Println("  HTTP API:    http://localhost:8080")
	log.Println("  WebSocket:   ws://localhost:8080/ws")
	log.Println("")
	log.Println("  OBSERVABILITY:")
	log.Println("    GET  /metrics               - Prometheus Metrics")
	log.Println("    GET  /health/live           - Liveness Probe")
	log.Println("    GET  /health/ready          - Readiness Probe")
	log.Println("")
	log.Println("  B-BOOK API (RTX Internal Balance):")
	log.Println("    GET  /api/account/summary   - RTX Balance/Equity/Margin")
	log.Println("    GET  /api/positions         - RTX Open Positions")
	log.Println("    POST /api/orders/market     - Execute Market Order")
	log.Println("    POST /api/positions/close   - Close Position")
	log.Println("    GET  /api/trades            - Trade History")
	log.Println("    GET  /api/ledger            - Transaction History")
	log.Println("")
	log.Println("  ADMIN ENDPOINTS:")
	log.Println("    GET  /admin/accounts        - List All Accounts")
	log.Println("    POST /admin/deposit         - Add Funds (Bank/Crypto)")
	log.Println("    POST /admin/withdraw        - Withdraw Funds")
	log.Println("    POST /admin/adjust          - Manual Adjustment")
	log.Println("    POST /admin/bonus           - Add Bonus")
	log.Println("    GET  /admin/ledger          - View All Transactions")
	log.Println("")
	log.Printf("  Demo Account: %s | Balance: $%.2f", demoAccount.AccountNumber, brokerConfig.DefaultBalance)
	log.Println("═══════════════════════════════════════════════════════════")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
