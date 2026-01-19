package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

// Global tick tracker for debugging market data flow
var (
	latestTicks    = make(map[string]*ws.MarketTick)
	tickMutex      sync.RWMutex
	totalTickCount int64
)

// HistoricalTick represents a tick from OANDA historical data
type HistoricalTick struct {
	BrokerID  string  `json:"broker_id"`
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Spread    float64 `json:"spread"`
	Timestamp string  `json:"timestamp"`
	LP        string  `json:"lp"`
}

// HistoricalDataCache stores loaded historical tick data
type HistoricalDataCache struct {
	Ticks      []HistoricalTick
	LastIndex  int
	Symbol     string
	AvgSpread  float64
	PipSize    float64
	mu         sync.RWMutex
}

// Global cache for historical data
var historicalCache = make(map[string]*HistoricalDataCache)
var historicalCacheMutex sync.RWMutex

// loadHistoricalTickData loads OANDA historical tick data from files
func loadHistoricalTickData(symbol string, dataDir string) (*HistoricalDataCache, error) {
	// Check if already cached
	historicalCacheMutex.RLock()
	if cache, exists := historicalCache[symbol]; exists {
		historicalCacheMutex.RUnlock()
		return cache, nil
	}
	historicalCacheMutex.RUnlock()

	// Build path to historical data
	tickPath := filepath.Join(dataDir, "data", "ticks", symbol)

	// Find the most recent data file
	files, err := ioutil.ReadDir(tickPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tick directory for %s: %v", symbol, err)
	}

	var latestFile string
	for i := len(files) - 1; i >= 0; i-- {
		if strings.HasSuffix(files[i].Name(), ".json") {
			latestFile = filepath.Join(tickPath, files[i].Name())
			break
		}
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no historical tick data found for %s", symbol)
	}

	// Read and parse the file
	data, err := ioutil.ReadFile(latestFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read tick file: %v", err)
	}

	var ticks []HistoricalTick
	if err := json.Unmarshal(data, &ticks); err != nil {
		return nil, fmt.Errorf("failed to parse tick data: %v", err)
	}

	if len(ticks) == 0 {
		return nil, fmt.Errorf("no ticks found in file")
	}

	// Calculate average spread from OANDA data
	var totalSpread float64
	oandaCount := 0
	for _, tick := range ticks {
		if tick.LP == "OANDA" {
			totalSpread += tick.Ask - tick.Bid
			oandaCount++
		}
	}

	avgSpread := 0.00015 // default 1.5 pips for 4-decimal pairs
	if oandaCount > 0 {
		avgSpread = totalSpread / float64(oandaCount)
	}

	// Determine pip size based on symbol
	pipSize := 0.0001 // default for most pairs
	if strings.Contains(symbol, "JPY") || strings.Contains(symbol, "HKD") {
		pipSize = 0.01 // 2-decimal pairs
	}

	cache := &HistoricalDataCache{
		Ticks:     ticks,
		LastIndex: 0,
		Symbol:    symbol,
		AvgSpread: avgSpread,
		PipSize:   pipSize,
	}

	// Store in cache
	historicalCacheMutex.Lock()
	historicalCache[symbol] = cache
	historicalCacheMutex.Unlock()

	log.Printf("[HISTORICAL] Loaded %d ticks for %s from %s (avg spread: %.5f, pip: %.5f)",
		len(ticks), symbol, filepath.Base(latestFile), avgSpread, pipSize)

	return cache, nil
}

// getNextHistoricalTick gets the next tick from historical data with small variations
func (cache *HistoricalDataCache) getNextHistoricalTick() *ws.MarketTick {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Get base tick from historical data (cycle through if needed)
	baseTick := cache.Ticks[cache.LastIndex]
	cache.LastIndex = (cache.LastIndex + 1) % len(cache.Ticks)

	// Use historical price as base, but add small random variation to simulate live market
	// This gives realistic price levels while still showing movement
	variation := (rand.Float64()*2 - 1) * cache.PipSize * 2 // -2 to +2 pips

	bid := baseTick.Bid + variation
	ask := bid + cache.AvgSpread

	return &ws.MarketTick{
		Type:      "tick",
		Symbol:    cache.Symbol,
		Bid:       bid,
		Ask:       ask,
		Spread:    ask - bid,
		Timestamp: time.Now().Unix(),
		LP:        "OANDA-HISTORICAL", // Clearly marked as historical data
	}
}

func main() {
	// ============================================
	// GC TUNING - Prevents memory crashes during high-frequency quote processing
	// ============================================
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

	// Initialize OPTIMIZED tick storage with:
	// - Ring buffers (bounded memory, O(1) operations)
	// - Quote throttling (skip < 0.001% price changes)
	// - Async batch writer (non-blocking disk persistence)
	tickStore := tickstore.NewOptimizedTickStore("default", brokerConfig.MaxTicksPerSymbol)

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

	// Symbol Management API (for Market Watch)
	http.HandleFunc("/api/symbols/available", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Comprehensive list of all tradeable symbols from YOFX and other LPs
		availableSymbols := []map[string]interface{}{
			// Major Forex Pairs
			{"symbol": "EURUSD", "name": "Euro/US Dollar", "category": "forex.major", "digits": 5},
			{"symbol": "GBPUSD", "name": "British Pound/US Dollar", "category": "forex.major", "digits": 5},
			{"symbol": "USDJPY", "name": "US Dollar/Japanese Yen", "category": "forex.major", "digits": 3},
			{"symbol": "USDCHF", "name": "US Dollar/Swiss Franc", "category": "forex.major", "digits": 5},
			{"symbol": "USDCAD", "name": "US Dollar/Canadian Dollar", "category": "forex.major", "digits": 5},
			{"symbol": "AUDUSD", "name": "Australian Dollar/US Dollar", "category": "forex.major", "digits": 5},
			{"symbol": "NZDUSD", "name": "New Zealand Dollar/US Dollar", "category": "forex.major", "digits": 5},
			// Cross Pairs
			{"symbol": "EURGBP", "name": "Euro/British Pound", "category": "forex.cross", "digits": 5},
			{"symbol": "EURJPY", "name": "Euro/Japanese Yen", "category": "forex.cross", "digits": 3},
			{"symbol": "GBPJPY", "name": "British Pound/Japanese Yen", "category": "forex.cross", "digits": 3},
			{"symbol": "EURAUD", "name": "Euro/Australian Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "EURCAD", "name": "Euro/Canadian Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "EURCHF", "name": "Euro/Swiss Franc", "category": "forex.cross", "digits": 5},
			{"symbol": "AUDCAD", "name": "Australian Dollar/Canadian Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "AUDCHF", "name": "Australian Dollar/Swiss Franc", "category": "forex.cross", "digits": 5},
			{"symbol": "AUDJPY", "name": "Australian Dollar/Japanese Yen", "category": "forex.cross", "digits": 3},
			{"symbol": "AUDNZD", "name": "Australian Dollar/New Zealand Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "CADCHF", "name": "Canadian Dollar/Swiss Franc", "category": "forex.cross", "digits": 5},
			{"symbol": "CADJPY", "name": "Canadian Dollar/Japanese Yen", "category": "forex.cross", "digits": 3},
			{"symbol": "CHFJPY", "name": "Swiss Franc/Japanese Yen", "category": "forex.cross", "digits": 3},
			{"symbol": "GBPAUD", "name": "British Pound/Australian Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "GBPCAD", "name": "British Pound/Canadian Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "GBPCHF", "name": "British Pound/Swiss Franc", "category": "forex.cross", "digits": 5},
			{"symbol": "GBPNZD", "name": "British Pound/New Zealand Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "NZDCAD", "name": "New Zealand Dollar/Canadian Dollar", "category": "forex.cross", "digits": 5},
			{"symbol": "NZDCHF", "name": "New Zealand Dollar/Swiss Franc", "category": "forex.cross", "digits": 5},
			{"symbol": "NZDJPY", "name": "New Zealand Dollar/Japanese Yen", "category": "forex.cross", "digits": 3},
			// Exotic Pairs
			{"symbol": "EURNOK", "name": "Euro/Norwegian Krone", "category": "forex.exotic", "digits": 5},
			{"symbol": "EURSEK", "name": "Euro/Swedish Krona", "category": "forex.exotic", "digits": 5},
			{"symbol": "EURTRY", "name": "Euro/Turkish Lira", "category": "forex.exotic", "digits": 5},
			{"symbol": "EURZAR", "name": "Euro/South African Rand", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDNOK", "name": "US Dollar/Norwegian Krone", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDSEK", "name": "US Dollar/Swedish Krona", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDTRY", "name": "US Dollar/Turkish Lira", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDZAR", "name": "US Dollar/South African Rand", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDMXN", "name": "US Dollar/Mexican Peso", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDSGD", "name": "US Dollar/Singapore Dollar", "category": "forex.exotic", "digits": 5},
			{"symbol": "USDHKD", "name": "US Dollar/Hong Kong Dollar", "category": "forex.exotic", "digits": 5},
			// Metals
			{"symbol": "XAUUSD", "name": "Gold/US Dollar", "category": "metals", "digits": 2},
			{"symbol": "XAGUSD", "name": "Silver/US Dollar", "category": "metals", "digits": 3},
			{"symbol": "XPTUSD", "name": "Platinum/US Dollar", "category": "metals", "digits": 2},
			{"symbol": "XPDUSD", "name": "Palladium/US Dollar", "category": "metals", "digits": 2},
			// Indices
			{"symbol": "US30USD", "name": "Dow Jones 30", "category": "indices", "digits": 1},
			{"symbol": "SPX500USD", "name": "S&P 500", "category": "indices", "digits": 1},
			{"symbol": "NAS100USD", "name": "NASDAQ 100", "category": "indices", "digits": 1},
			{"symbol": "UK100GBP", "name": "UK 100", "category": "indices", "digits": 1},
			{"symbol": "DE30EUR", "name": "Germany 30", "category": "indices", "digits": 1},
			{"symbol": "JP225USD", "name": "Japan 225", "category": "indices", "digits": 0},
			// Crypto (if supported)
			{"symbol": "BTCUSD", "name": "Bitcoin/US Dollar", "category": "crypto", "digits": 2},
			{"symbol": "ETHUSD", "name": "Ethereum/US Dollar", "category": "crypto", "digits": 2},
			// Commodities
			{"symbol": "WTICOUSD", "name": "WTI Crude Oil", "category": "commodities", "digits": 3},
			{"symbol": "BCOUSD", "name": "Brent Crude Oil", "category": "commodities", "digits": 3},
			{"symbol": "NATGASUSD", "name": "Natural Gas", "category": "commodities", "digits": 3},
		}

		// Check which symbols are currently subscribed via FIX
		fixGateway := server.GetFIXGateway()
		if fixGateway != nil {
			subscribedSymbols := fixGateway.GetSubscribedSymbols()
			subscribedMap := make(map[string]bool)
			for _, s := range subscribedSymbols {
				subscribedMap[s] = true
			}
			// Add subscribed status to each symbol
			for _, sym := range availableSymbols {
				symbol := sym["symbol"].(string)
				sym["subscribed"] = subscribedMap[symbol]
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
			"success":  true,
			"symbol":   req.Symbol,
			"mdReqId":  mdReqID,
			"message":  "Subscribed successfully",
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

	// Auto-Connect FIX Sessions on startup
	go func() {
		time.Sleep(3 * time.Second) // Wait for other services to initialize

		// Connect YOFX1 (Trading)
		log.Println("[FIX] Auto-connecting YOFX1 session (Trading)...")
		if err := server.ConnectToLP("YOFX1"); err != nil {
			log.Printf("[FIX] Failed to auto-connect YOFX1: %v", err)
		}

		// Connect YOFX2 (Market Data) after short delay
		time.Sleep(2 * time.Second)
		log.Println("[FIX] Auto-connecting YOFX2 session (Market Data)...")
		if err := server.ConnectToLP("YOFX2"); err != nil {
			log.Printf("[FIX] Failed to auto-connect YOFX2: %v", err)
		} else {
			// First request security list to discover available symbols
			time.Sleep(2 * time.Second)
			fixGateway := server.GetFIXGateway()
			if fixGateway != nil {
				// Request available securities from YOFX
				log.Println("[FIX] Requesting security list from YOFX2...")
				if _, err := fixGateway.RequestSecurityList("YOFX2"); err != nil {
					log.Printf("[FIX] Failed to request security list: %v", err)
				}

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
				log.Printf("[FIX-WS] Piping FIX tick #%d: %s Bid=%.5f Ask=%.5f",
					tickCount, md.Symbol, md.Bid, md.Ask)
			}

			tick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    md.Symbol,
				Bid:       md.Bid,
				Ask:       md.Ask,
				Spread:    md.Ask - md.Bid,
				Timestamp: md.Timestamp.Unix(),
				LP:        "YOFX", // FIX LP source
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

	// Simulated market data fallback - uses OANDA historical data when LP unavailable
	go func() {
		// Wait 30 seconds to see if real market data arrives
		time.Sleep(30 * time.Second)

		tickMutex.RLock()
		hasRealData := totalTickCount > 0
		tickMutex.RUnlock()

		if hasRealData {
			log.Println("[SIM-MD] Real market data detected, simulation not needed")
			return
		}

		log.Println("[SIM-MD] No real market data after 30s - starting OANDA historical data simulation")
		log.Println("[SIM-MD] Using real OANDA tick data with small variations for realistic prices")

		// Get current working directory (backend/cmd/server when running server.exe)
		workDir, err := os.Getwd()
		if err != nil {
			log.Printf("[SIM-MD] Failed to get working directory: %v", err)
			workDir = "."
		}

		// Determine data directory - check if we're in backend/cmd/server or backend
		dataDir := workDir
		if strings.HasSuffix(workDir, "cmd\\server") || strings.HasSuffix(workDir, "cmd/server") {
			// We're in backend/cmd/server, go up to backend
			dataDir = filepath.Join(workDir, "..", "..")
		} else if !strings.HasSuffix(workDir, "backend") {
			// We might be in project root, add backend
			dataDir = filepath.Join(workDir, "backend")
		}

		log.Printf("[SIM-MD] Loading historical data from: %s", dataDir)

		// Symbols to simulate with historical data
		symbols := []string{
			"EURUSD", "GBPUSD", "USDJPY", "AUDUSD",
			"USDCAD", "USDCHF", "NZDUSD", "EURGBP",
			"EURJPY", "GBPJPY", "AUDJPY", "AUDCAD",
			"AUDCHF", "AUDNZD", "AUDSGD", "AUDHKD",
		}

		// Load historical data for all symbols
		historicalDataLoaded := make(map[string]*HistoricalDataCache)
		for _, symbol := range symbols {
			cache, err := loadHistoricalTickData(symbol, dataDir)
			if err != nil {
				log.Printf("[SIM-MD] Failed to load historical data for %s: %v (will skip)", symbol, err)
				continue
			}
			historicalDataLoaded[symbol] = cache
		}

		if len(historicalDataLoaded) == 0 {
			log.Println("[SIM-MD] ERROR: No historical data loaded, cannot simulate market data")
			return
		}

		log.Printf("[SIM-MD] Successfully loaded historical data for %d symbols", len(historicalDataLoaded))

		// Generate simulated ticks every 500ms using historical data
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			tickMutex.RLock()
			realDataArrived := totalTickCount > 0
			tickMutex.RUnlock()

			if realDataArrived {
				log.Println("[SIM-MD] Real market data now available - stopping historical simulation")
				return
			}

			// Generate tick for each symbol using historical data
			for symbol, cache := range historicalDataLoaded {
				tick := cache.getNextHistoricalTick()

				tickMutex.Lock()
				latestTicks[symbol] = tick
				tickMutex.Unlock()

				hub.BroadcastTick(tick)
			}
		}
	}()

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
