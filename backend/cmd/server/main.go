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

	// Create Admin Auth Service for admin panel handlers
	adminAuthService := admin.NewAuthService()

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
	adminAuthService = adminHandler.GetAuthService()
	log.Println("[Admin] Admin system initialized")

	// ============================================
	// Initialize White Label Service
	// ============================================
	whiteLabelService := admin.NewWhiteLabelService()
	whiteLabelHandler := admin.NewWhiteLabelHandler(whiteLabelService, authService)
	log.Println("[WhiteLabel] White-label branding system initialized with 5 default brand configs")

	// ============================================
	// Initialize Promotion Service
	// ============================================
	promotionService := admin.NewPromotionService()
	promotionHandler := admin.NewPromotionHandler(promotionService, authService)
	log.Println("[Promotions] Promotion/bonus system initialized with 6 default promotions")

	// ============================================
	// Initialize Client Segmentation System
	// ============================================
	segmentStore := admin.NewSegmentStore()
	segmentHandler := admin.NewSegmentHandler(segmentStore, authService)
	log.Println("[Segmentation] Client segmentation system initialized (7 predefined segments, 5 custom tags)")

	// ============================================
	// Initialize Client Notes Service
	// ============================================
	clientNoteService := admin.NewClientNoteService()
	clientNoteHandler := admin.NewClientNoteHandler(clientNoteService, authService)
	log.Println("[ClientNotes] Client notes/CRM system initialized with 50 clients and 220+ notes")

	// ============================================
	// Initialize Platform Configuration Service
	// ============================================
	platformConfigService := admin.NewPlatformConfigService()
	platformConfigHandler := admin.NewPlatformConfigHandler(platformConfigService, adminAuthService)
	log.Println("[PlatformConfig] Platform configuration system initialized with 6 sections and change history tracking")

	// ============================================
	// Initialize Trading Competition Service
	// ============================================
	competitionStore := admin.NewCompetitionStore()
	competitionHandler := admin.NewCompetitionHandler(competitionStore, authService)
	log.Println("[Competitions] Trading competition system initialized with 8 competitions and 100+ participants")

	// ============================================
	// Initialize Localization Service
	// ============================================
	localizationService := admin.NewLocalizationService()
	localizationHandler := admin.NewLocalizationHandler(localizationService, adminAuthService)
	log.Println("[Localization] Multi-language system initialized with 10 languages and 500+ translation keys")

	// ============================================
	// Initialize Order Flow Analytics Service
	// ============================================
	orderFlowService := admin.NewOrderFlowService()
	orderFlowHandler := admin.NewOrderFlowHandler(orderFlowService, adminAuthService)
	log.Println("[OrderFlow] Order flow analytics initialized with 500 entries across 20 symbols and market microstructure data")

	// ============================================
	// Initialize Performance Monitor Service
	// ============================================
	performanceMonitorService := admin.NewPerformanceMonitorService()
	performanceMonitorHandler := admin.NewPerformanceMonitorHandler(performanceMonitorService, adminAuthService)
	log.Println("[PerformanceMonitor] System metrics initialized with 60 data points, 30 endpoints, 8 service health checks")

	// ============================================
	// Initialize Notification Center Service
	// ============================================
	notificationCenterService := admin.NewNotificationCenterService()
	notificationCenterHandler := admin.NewNotificationCenterHandler(notificationCenterService, adminAuthService)
	log.Println("[NotificationCenter] Notification system initialized with 500 notifications across 50 clients and 15 rules")

	// ============================================
	// Initialize Client Document / KYC Service
	// ============================================
	clientDocumentService := admin.NewClientDocumentService()
	clientDocumentHandler := admin.NewClientDocumentHandler(clientDocumentService, authService)
	log.Println("[ClientDocuments] Document management initialized with 300 documents across 80 clients")

	// ============================================
	// Initialize Commission Tier Service
	// ============================================
	commissionTierService := admin.NewCommissionTierService()
	commissionTierHandler := admin.NewCommissionTierHandler(commissionTierService, authService)
	log.Println("[CommissionTiers] Multi-tier commission system initialized with 5 tiers, 150 clients, 500 transactions")

	// ============================================
	// Initialize Trading Signal Service
	// ============================================
	tradingSignalService := admin.NewTradingSignalService()
	tradingSignalHandler := admin.NewTradingSignalHandler(tradingSignalService, authService)
	log.Println("[TradingSignals] Signal provider system initialized with 15 providers, 500 signals, 200 subscriptions")

	// ============================================
	// Initialize Wallet Management Service
	// ============================================
	walletService := admin.NewWalletService()
	walletHandler := admin.NewWalletHandler(walletService, authService)
	log.Println("[WalletManagement] Multi-currency wallet system initialized with 50 wallets, 8 currencies, 200 transactions")

	// ============================================
	// Initialize Account Statement Service
	// ============================================
	statementService := admin.NewStatementService()
	statementHandler := admin.NewStatementHandler(statementService, authService)
	log.Println("[AccountStatement] Statement generation system initialized with 2000 trades, 500 transactions, 50 history entries")

	// ============================================
	// Initialize Risk Scoring Service
	// ============================================
	riskScoringService := admin.NewRiskScoringService()
	riskScoringHandler := admin.NewRiskScoringHandler(riskScoringService, authService)
	log.Println("[RiskScoring] Client risk scoring system initialized with 200 clients, 8 weighted factors, 30-day history")

	// ============================================
	// Initialize Trading Session Service
	// ============================================
	tradingSessionService := admin.NewTradingSessionService()
	tradingSessionHandler := admin.NewTradingSessionHandler(tradingSessionService, authService)
	log.Println("[TradingSession] Trading session system initialized with 5 sessions, 6 templates, 150 symbols, 30 holidays")

	// ============================================
	// Initialize Compliance Report Service
	// ============================================
	complianceReportService := admin.NewComplianceReportService()
	complianceReportHandler := admin.NewComplianceReportHandler(complianceReportService, authService)
	log.Println("[ComplianceReport] Regulatory compliance system initialized with 50 reports, 8 templates, 20 checks, 30 breaches")

	// ============================================
	// Initialize LP Performance Service
	// ============================================
	lpPerformanceService := admin.NewLPPerformanceService()
	lpPerformanceHandler := admin.NewLPPerformanceHandler(lpPerformanceService, authService)
	log.Println("[LPPerformance] LP performance analytics initialized with 6 LPs, 30-day history, 10000 execution records")

	// ============================================
	// Initialize Position Aggregation Service
	// ============================================
	positionAggregationService := admin.NewPositionAggregationService()
	positionAggregationHandler := admin.NewPositionAggregationHandler(positionAggregationService, authService)
	log.Println("[PositionAggregation] Position aggregation system initialized with 500 positions, 100 clients, 30 symbols, 5 groups")

	// ============================================
	// Initialize Margin Monitoring Service
	// ============================================
	marginMonitoringService := admin.NewMarginMonitoringService()
	marginMonitoringHandler := admin.NewMarginMonitoringHandler(marginMonitoringService, authService)
	log.Println("[MarginMonitoring] Margin monitoring system initialized with 200 clients, 50 margin calls, 100 liquidations, 5 threshold groups")

	// ============================================
	// Initialize Execution Policy Service
	// ============================================
	executionPolicyService := admin.NewExecutionPolicyService()
	executionPolicyHandler := admin.NewExecutionPolicyHandler(executionPolicyService, authService)
	log.Println("[ExecutionPolicy] Execution policy system initialized with 3 policies, 8 slippage configs, 5000 execution records")

	// ============================================
	// Initialize Fee Schedule Service
	// ============================================
	feeScheduleService := admin.NewFeeScheduleService()
	feeScheduleHandler := admin.NewFeeScheduleHandler(feeScheduleService, authService)
	log.Println("[FeeSchedule] Fee schedule system initialized with 6 schedules, 200 clients, 12-month revenue data, 30 waivers")

	// ============================================
	// Initialize Client Messaging Service
	// ============================================
	clientMessagingService := admin.NewClientMessagingService()
	clientMessagingHandler := admin.NewClientMessagingHandler(clientMessagingService, authService)
	log.Println("[ClientMessaging] Client messaging system initialized with 500 messages, 50 threads, 10 templates, 20 announcements")

	// ============================================
	// Initialize Market Data Aggregation Service
	// ============================================
	marketDataAggregationService := admin.NewMarketDataAggregationService()
	marketDataAggregationHandler := admin.NewMarketDataAggregationHandler(marketDataAggregationService, authService)
	log.Println("[MarketDataAggregation] Market data aggregation system initialized with 6 sources, 30 aggregation rules, 30 quotes")

	// ============================================
	// Initialize Trade Surveillance Service
	// ============================================
	tradeSurveillanceService := admin.NewTradeSurveillanceService()
	tradeSurveillanceHandler := admin.NewTradeSurveillanceHandler(tradeSurveillanceService, authService)
	log.Println("[TradeSurveillance] Trade surveillance system initialized with 8 rules, 100 alerts, 30-day trend data")

	// ============================================
	// Initialize Client Portfolio Analysis Service
	// ============================================
	clientPortfolioService := admin.NewClientPortfolioService()
	clientPortfolioHandler := admin.NewClientPortfolioHandler(clientPortfolioService, adminAuthService)
	log.Println("[ClientPortfolio] Client portfolio analysis system initialized with 200 portfolios, 150 patterns, 30-day snapshots")

	// ============================================
	// Initialize Server Cluster Management Service
	// ============================================
	serverClusterService := admin.NewServerClusterService()
	serverClusterHandler := admin.NewServerClusterHandler(serverClusterService, adminAuthService)
	log.Println("[ServerCluster] Server cluster management initialized with 5 nodes across 3 regions, 24h metrics, load balancer config")

	// ============================================
	// Initialize Platform License Management Service
	// ============================================
	platformLicenseService := admin.NewPlatformLicenseService()
	platformLicenseHandler := admin.NewPlatformLicenseHandler(platformLicenseService, adminAuthService)
	log.Println("[PlatformLicense] Platform licensing system initialized with 10 licenses, 25 features, usage metrics, billing records")

	// ============================================
	// Initialize Automated Trading / EA Management Service
	// ============================================
	automatedTradingService := admin.NewAutomatedTradingService()
	automatedTradingHandler := admin.NewAutomatedTradingHandler(automatedTradingService, adminAuthService)
	log.Println("[AutomatedTrading] Expert Advisor management initialized with 20 EAs, 8 strategies, performance tracking, resource monitoring")

	// ============================================
	// Initialize Social Trading / Copy Trading Service
	// ============================================
	socialTradingService := admin.NewSocialTradingService()
	socialTradingHandler := admin.NewSocialTradingHandler(socialTradingService, adminAuthService)
	log.Println("[SocialTrading] Copy trading system initialized with 15 signal providers, 30 copy relationships, performance rankings")

	// ============================================
	// Initialize API Gateway / Rate Limit Service
	// ============================================
	apiGatewayService := admin.NewAPIGatewayService()
	apiGatewayHandler := admin.NewAPIGatewayHandler(apiGatewayService, adminAuthService)
	log.Println("[APIGateway] API gateway initialized with 20 clients, 8 rate limit tiers, 1000 usage records, 10 IP rules, 5 geo rules")

	// ============================================
	// Initialize Market News / Economic Calendar Service
	// ============================================
	marketNewsService := admin.NewMarketNewsService()
	marketNewsHandler := admin.NewMarketNewsHandler(marketNewsService, adminAuthService)
	log.Println("[MarketNews] Market news system initialized with 50 articles from 5 sources, 30 economic events, sentiment analysis, trading pause rules")

	// ============================================
	// Initialize Multi-Tenant / White Label Service
	// ============================================
	multiTenantService := admin.NewMultiTenantService()
	multiTenantHandler := admin.NewMultiTenantHandler(multiTenantService, adminAuthService)
	log.Println("[MultiTenant] Multi-tenant system initialized with 8 tenant instances, custom branding, resource quotas, isolation metrics")

	// ============================================
	// Initialize MAM/PAMM System
	// ============================================
	mamStore := admin.NewMAMStore()
	mamHandler := admin.NewMAMHandler(mamStore, authService)
	log.Println("[MAM] MAM/PAMM system initialized (8 money managers, 40+ investors, 12 months performance)")

	// ============================================
	// Initialize Broker P&L Report System
	// ============================================
	brokerPnLHandler := admin.NewBrokerPnLHandler(authService)
	log.Println("[BrokerPnL] Broker P&L reporting system initialized (revenue breakdown, symbol groups, top performers)")

	// ============================================
	// Initialize Deposit/Withdrawal Processing System
	// ============================================
	transactionHandler := admin.NewTransactionHandler(authService)
	log.Println("[Transactions] Deposit/Withdrawal processing system initialized (150 mock transactions, 7 endpoints)")

	// ============================================
	// Initialize Audit Trail System
	// ============================================
	auditStore := admin.NewAuditStore()
	auditHandler := admin.NewAuditTrailHandler(auditStore, authService)
	log.Println("[AuditTrail] Audit trail system initialized (100 audit entries, 5 admin users)")

	// ============================================
	// Initialize Risk Dashboard System
	// ============================================
	riskStore := admin.NewRiskDashboardStore()
	riskHandler := admin.NewRiskDashboardHandler(riskStore, authService)
	log.Println("[RiskDashboard] Risk dashboard system initialized (500 positions, 30 symbols, VaR calculations, concentration analysis)")

	// ============================================
	// Initialize Customer Lifecycle Analytics System
	// ============================================
	lifecycleStore := admin.NewLifecycleStore()
	lifecycleHandler := admin.NewLifecycleHandler(lifecycleStore, authService)
	log.Println("[Lifecycle] Customer lifecycle analytics initialized (200 clients, 6 stages, 12-month cohorts, churn prediction)")

	// ============================================
	// Initialize Paper Trading / Demo Account System
	// ============================================
	paperTradingStore := admin.NewPaperTradingStore()
	paperTradingHandler := admin.NewPaperTradingHandler(paperTradingStore, authService)
	log.Println("[PaperTrading] Paper trading system initialized (50 demo accounts, 200 positions, conversion tracking)")

	// ============================================
	// Initialize Scheduled Reports System
	// ============================================
	scheduledReportsStore := admin.NewScheduledReportsStore()
	scheduledReportsHandler := admin.NewScheduledReportsHandler(scheduledReportsStore, authService)
	log.Println("[ScheduledReports] Scheduled reports system initialized (20 reports, 100 executions, 5 templates)")

	// ============================================
	// Initialize Fraud Detection System
	// ============================================
	fraudDetectionStore := admin.NewFraudDetectionStore()
	fraudDetectionHandler := admin.NewFraudDetectionHandler(fraudDetectionStore, authService)
	log.Println("[FraudDetection] Fraud detection system initialized (1000 logins, 50 alerts, 20 rules, 30 countries)")

	// ============================================
	// Initialize Backup Manager System
	// ============================================
	backupStore := admin.NewBackupStore()
	backupHandler := admin.NewBackupHandler(backupStore, authService)
	log.Println("[BackupManager] Backup manager system initialized (30 backups, 5 schedules, 8 components, 10 restore points)")

	// ============================================
	// Initialize Referral Program System
	// ============================================
	referralStore := admin.NewReferralStore()
	referralHandler := admin.NewReferralHandler(referralStore, authService)
	log.Println("[ReferralProgram] Referral program system initialized (50 links, 200 referrals, 4 tiers, 12-month trends)")

	// ============================================
	// Initialize Instrument Groups System
	// ============================================
	instrumentGroupStore := admin.NewInstrumentGroupStore()
	instrumentGroupHandler := admin.NewInstrumentGroupHandler(instrumentGroupStore, authService)
	log.Println("[InstrumentGroups] Instrument groups system initialized (8 groups, 150 symbols)")

	// ============================================
	// Initialize Price Alerts System
	// ============================================
	priceAlertStore := admin.NewPriceAlertStore()
	priceAlertHandler := admin.NewPriceAlertHandler(priceAlertStore, authService)
	_ = priceAlertHandler // TODO: register price alert routes
	log.Println("[PriceAlerts] Price alerts system initialized (500 alerts, 100 clients, 30 symbols, 200 triggered history)")

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

	// Initialize User Data Folder Handler
	userDataHandler := api.NewUserDataHandler()
	http.HandleFunc("/api/user/data-folder", userDataHandler.HandleGetDataFolderPath)
	http.HandleFunc("/api/user/data-folder/list", userDataHandler.HandleListFiles)
	http.HandleFunc("/api/user/data-folder/download", userDataHandler.HandleDownloadFile)
	http.HandleFunc("/api/user/data-folder/upload", userDataHandler.HandleUploadFile)
	http.HandleFunc("/api/user/data-folder/delete", userDataHandler.HandleDeleteFile)
	http.HandleFunc("/api/user/data-folder/mkdir", userDataHandler.HandleCreateDirectory)
	log.Println("[UserData] User data folder API registered")

	// Initialize Workspace API
	http.HandleFunc("/api/workspaces", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			server.HandleSaveWorkspace(w, r)
		} else if r.Method == http.MethodGet {
			server.HandleListWorkspaces(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/workspaces/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			server.HandleLoadWorkspace(w, r)
		} else if r.Method == http.MethodDelete {
			server.HandleDeleteWorkspace(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[Workspaces] Workspace API registered")

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

	// Print Preferences API
	printPrefsStore := api.NewPrintPreferencesStore()
	http.HandleFunc("/api/accounts/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Route to print preferences handlers
		if strings.Contains(r.URL.Path, "/print-preferences") {
			switch r.Method {
			case "GET":
				printPrefsStore.HandleGetPrintPreferences(w, r)
			case "POST":
				printPrefsStore.HandleSavePrintPreferences(w, r)
			case "DELETE":
				printPrefsStore.HandleDeletePrintPreferences(w, r)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	})

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

	// ===== BROKER P&L REPORTING API =====
	http.HandleFunc("/admin/broker-pnl/summary", brokerPnLHandler.HandleGetSummary)
	http.HandleFunc("/admin/broker-pnl/by-symbol-group", brokerPnLHandler.HandleGetBySymbolGroup)
	http.HandleFunc("/admin/broker-pnl/top-symbols", brokerPnLHandler.HandleGetTopSymbols)
	http.HandleFunc("/admin/broker-pnl/top-clients", brokerPnLHandler.HandleGetTopClients)
	http.HandleFunc("/admin/broker-pnl/monthly-revenue", brokerPnLHandler.HandleGetMonthlyRevenue)
	http.HandleFunc("/admin/broker-pnl/daily-trend", brokerPnLHandler.HandleGetDailyTrend)
	http.HandleFunc("/admin/broker-pnl/export", brokerPnLHandler.HandleExport)

	// ===== DEPOSIT/WITHDRAWAL PROCESSING API =====
	http.HandleFunc("/admin/transactions", transactionHandler.HandleGetTransactions)
	http.HandleFunc("/admin/transactions/stats", transactionHandler.HandleGetTransactionStats)
	http.HandleFunc("/admin/transactions/daily", transactionHandler.HandleGetDailyTransactions)
	http.HandleFunc("/admin/transactions/manual", transactionHandler.HandleCreateManualTransaction)
	// Note: /admin/transactions/:id handlers are registered separately below
	http.HandleFunc("/admin/transactions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/approve") {
			transactionHandler.HandleApproveTransaction(w, r)
		} else if strings.HasSuffix(path, "/reject") {
			transactionHandler.HandleRejectTransaction(w, r)
		} else {
			transactionHandler.HandleGetTransactionByID(w, r)
		}
	})

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

	// Register Email Template Management endpoints
	emailTemplateService := admin.NewEmailTemplateService()
	http.HandleFunc("/admin/email-templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			emailTemplateService.HandleListTemplates(w, r)
		} else if r.Method == http.MethodPost {
			emailTemplateService.HandleCreateTemplate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/email-templates/categories", emailTemplateService.HandleGetCategories)
	http.HandleFunc("/admin/email-templates/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, DELETE, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		// Extract template ID from path
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/email-templates/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Template ID required", http.StatusBadRequest)
			return
		}
		// Check for sub-routes
		if len(parts) > 1 {
			action := parts[1]
			if action == "preview" && r.Method == http.MethodPost {
				emailTemplateService.HandlePreviewTemplate(w, r)
			} else if action == "test" && r.Method == http.MethodPost {
				emailTemplateService.HandleSendTestEmail(w, r)
			} else {
				http.Error(w, "Unknown action", http.StatusNotFound)
			}
			return
		}

		// Base template operations
		if r.Method == http.MethodGet {
			emailTemplateService.HandleGetTemplate(w, r)
		} else if r.Method == http.MethodPut {
			emailTemplateService.HandleUpdateTemplate(w, r)
		} else if r.Method == http.MethodDelete {
			emailTemplateService.HandleArchiveTemplate(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[Admin] Email template management routes registered: 8 endpoints (list, create, get, update, archive, preview, test, categories)")
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

	// ============================================
	// White Label API Routes (6 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/white-label", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			whiteLabelHandler.ListConfigs(w, r)
		} else if r.Method == http.MethodPost {
			whiteLabelHandler.CreateConfig(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/white-label/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/preview") {
			whiteLabelHandler.PreviewConfig(w, r)
		} else if r.Method == http.MethodGet {
			whiteLabelHandler.GetConfig(w, r)
		} else if r.Method == http.MethodPut {
			whiteLabelHandler.UpdateConfig(w, r)
		} else if r.Method == http.MethodDelete {
			whiteLabelHandler.DeleteConfig(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[WhiteLabelAPI] White-label branding API registered: 6 endpoints (list, get, create, update, delete, preview)")

	// ============================================
	// Promotion / Bonus API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/promotions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			promotionHandler.ListPromotions(w, r)
		} else if r.Method == http.MethodPost {
			promotionHandler.CreatePromotion(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/promotions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/stats") {
			promotionHandler.GetPromotionStats(w, r)
		} else if strings.Contains(r.URL.Path, "/status") {
			promotionHandler.UpdatePromotionStatus(w, r)
		} else if strings.Contains(r.URL.Path, "/claims") && r.Method == http.MethodGet {
			promotionHandler.GetPromotionClaims(w, r)
		} else if strings.Contains(r.URL.Path, "/claim") && r.Method == http.MethodPost {
			promotionHandler.ClaimPromotion(w, r)
		} else if r.Method == http.MethodGet {
			promotionHandler.GetPromotion(w, r)
		} else if r.Method == http.MethodPut {
			promotionHandler.UpdatePromotion(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[PromotionsAPI] Promotion/bonus API registered: 8 endpoints (list, get, create, update, status, claims, claim, stats)")

	// ============================================
	// Client Segmentation API Routes (9 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/segments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			segmentHandler.HandleListSegments(w, r)
		} else if r.Method == http.MethodPost {
			segmentHandler.HandleCreateSegment(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/segments/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			segmentHandler.HandleGetSegmentDetail(w, r)
		} else if r.Method == http.MethodPut {
			segmentHandler.HandleUpdateSegment(w, r)
		} else if r.Method == http.MethodDelete {
			segmentHandler.HandleDeleteSegment(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			segmentHandler.HandleListTags(w, r)
		} else if r.Method == http.MethodPost {
			segmentHandler.HandleCreateTag(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/clients/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tags") {
			// Segmentation tag operations
			if r.Method == http.MethodPost && !strings.Contains(r.URL.Path, "/tags/") {
				segmentHandler.HandleApplyTag(w, r)
			} else if r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/tags/") {
				segmentHandler.HandleRemoveTag(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.Contains(r.URL.Path, "/notes") {
			// Client notes/CRM operations
			if r.Method == http.MethodGet {
				clientNoteHandler.GetClientNotes(w, r)
			} else if r.Method == http.MethodPost {
				clientNoteHandler.CreateNote(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
	log.Println("[SegmentationAPI] Client segmentation API registered: 9 endpoints (list/create/update/delete segments, list/create tags, apply/remove tags)")

	// ============================================
	// Client Notes/CRM API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/clients/notes/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientNoteHandler.GetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/clients/notes/follow-ups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientNoteHandler.GetFollowUps(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/clients/notes/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/complete-followup") {
			if r.Method == http.MethodPost {
				clientNoteHandler.CompleteFollowUp(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodPut {
			clientNoteHandler.UpdateNote(w, r)
		} else if r.Method == http.MethodDelete {
			clientNoteHandler.DeleteNote(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/clients/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientNoteHandler.SearchClients(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	// NOTE: /admin/clients/ handler (for notes) is merged into segmentation handler above
	log.Println("[ClientNotesAPI] Client notes/CRM API registered: 8 endpoints (list/create/update/delete notes, stats, follow-ups, complete, search)")

	// ============================================
	// Platform Configuration API Routes (5 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/platform-config/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			platformConfigHandler.HandleGetHistory(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/platform-config/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/platform-config/"), "/")
		if len(parts) > 1 && parts[1] == "reset" {
			if r.Method == http.MethodPost {
				platformConfigHandler.HandleResetSection(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			platformConfigHandler.HandleGetSection(w, r)
		} else if r.Method == http.MethodPut {
			platformConfigHandler.HandleUpdateSection(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/platform-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			platformConfigHandler.HandleGetAllConfig(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[PlatformConfigAPI] Platform configuration API registered: 5 endpoints (get all, get section, update section, reset section, history)")

	// ============================================
	// Trading Competition API Routes (7 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/competitions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			competitionHandler.HandleList(w, r)
		} else if r.Method == http.MethodPost {
			competitionHandler.HandleCreate(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/competitions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/leaderboard") {
			if r.Method == http.MethodGet {
				competitionHandler.HandleGetLeaderboard(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.Contains(r.URL.Path, "/disqualify/") {
			if r.Method == http.MethodPost {
				competitionHandler.HandleDisqualify(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			competitionHandler.HandleGetDetail(w, r)
		} else if r.Method == http.MethodPut {
			competitionHandler.HandleUpdate(w, r)
		} else if r.Method == http.MethodDelete {
			competitionHandler.HandleDelete(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[CompetitionsAPI] Trading competition API registered: 7 endpoints (list, create, get, update, delete, leaderboard, disqualify)")

	// ============================================
	// Localization API Routes (6 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/localization/languages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			localizationHandler.HandleListLanguages(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/localization/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			localizationHandler.HandleImportTranslations(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/localization/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			localizationHandler.HandleUpdateTranslation(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/localization/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/admin/localization/"), "/")
		if len(parts) >= 2 && parts[1] == "export" {
			if r.Method == http.MethodGet {
				localizationHandler.HandleExportTranslations(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(parts) >= 2 && parts[1] == "toggle" {
			if r.Method == http.MethodPost {
				localizationHandler.HandleToggleLanguage(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if len(parts) >= 2 && parts[1] == "translations" {
			if r.Method == http.MethodGet {
				localizationHandler.HandleGetTranslations(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
	log.Println("[LocalizationAPI] Multi-language API registered: 6 endpoints (list languages, toggle, get translations, update, import, export)")

	// ============================================
	// Order Flow Analytics API Routes (6 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/order-flow/live", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderFlowHandler.HandleGetLiveFlow(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/order-flow/heatmap", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderFlowHandler.HandleGetHeatmap(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/order-flow/large-orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderFlowHandler.HandleGetLargeOrders(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/order-flow/imbalance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderFlowHandler.HandleGetImbalance(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/order-flow/top-symbols", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderFlowHandler.HandleGetTopSymbols(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/order-flow/metrics/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderFlowHandler.HandleGetMetrics(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[OrderFlowAPI] Order flow analytics API registered: 6 endpoints (live stream, metrics, heatmap, large orders, imbalance, top symbols)")

	// ============================================
	// Performance Monitor API Routes (7 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/performance/system", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetSystemMetrics(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/performance/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetMetricsHistory(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/performance/endpoints", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetEndpointMetrics(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/performance/services", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetServiceHealth(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/performance/slow-queries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetSlowQueries(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/performance/error-rate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetErrorRate(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/performance/connections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			performanceMonitorHandler.HandleGetConnections(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[PerformanceMonitorAPI] Performance monitoring API registered: 7 endpoints (system metrics, history, endpoints, services, slow queries, error rate, connections)")

	// ============================================
	// Notification Center API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/notifications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notificationCenterHandler.HandleGetNotifications(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/notifications/client/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notificationCenterHandler.HandleGetClientNotifications(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/notifications/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			notificationCenterHandler.HandleSendNotification(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/notifications/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notificationCenterHandler.HandleGetRules(w, r)
		} else if r.Method == http.MethodPost {
			notificationCenterHandler.HandleCreateRule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/notifications/rules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			notificationCenterHandler.HandleUpdateRule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/notifications/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			notificationCenterHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/notifications/broadcast", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			notificationCenterHandler.HandleBroadcast(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[NotificationCenterAPI] Notification center API registered: 8 endpoints (list, client notifications, send, rules management, stats, broadcast)")

	// ============================================
	// Client Document / KYC Management API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/documents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientDocumentHandler.HandleListDocuments(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/documents/client/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientDocumentHandler.HandleGetClientDocuments(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/documents/pending", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientDocumentHandler.HandleGetPendingDocuments(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/documents/expiring", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientDocumentHandler.HandleGetExpiringDocuments(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/documents/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientDocumentHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/documents/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/review") {
			if r.Method == http.MethodPut {
				clientDocumentHandler.HandleReviewDocument(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.Contains(r.URL.Path, "/request-resubmission") {
			if r.Method == http.MethodPost {
				clientDocumentHandler.HandleRequestResubmission(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			// Single document details GET /admin/documents/:id
			if r.Method == http.MethodGet {
				clientDocumentHandler.HandleGetDocument(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	log.Println("[ClientDocumentsAPI] Client document management API registered: 8 endpoints (list, client docs, single doc, review, pending queue, expiring, stats, request resubmission)")

	// ============================================
	// Commission Tier / Rebate Engine API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/commissions/tiers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			commissionTierHandler.HandleListTiers(w, r)
		} else if r.Method == http.MethodPost {
			commissionTierHandler.HandleCreateTier(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/commissions/tiers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			commissionTierHandler.HandleUpdateTier(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/commissions/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			commissionTierHandler.HandleListClients(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/commissions/clients/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/tier") {
			if r.Method == http.MethodPut {
				commissionTierHandler.HandleOverrideClientTier(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
	http.HandleFunc("/admin/commissions/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			commissionTierHandler.HandleListTransactions(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/commissions/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			commissionTierHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/commissions/calculate/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			commissionTierHandler.HandleCalculateCommission(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[CommissionTiersAPI] Multi-tier commission engine API registered: 8 endpoints (list tiers, create tier, update tier, list clients, override tier, transactions, stats, calculate)")

	// ============================================
	// Trading Signal / Signal Provider API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/signals/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleListProviders(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/signals/providers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleGetProvider(w, r)
		} else if r.Method == http.MethodPut {
			tradingSignalHandler.HandleUpdateProvider(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/signals/active", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleGetActiveSignals(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/signals/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleGetSignalHistory(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/signals/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleGetSubscriptions(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/signals/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/signals/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSignalHandler.HandleGetLeaderboard(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[TradingSignalsAPI] Trading signal provider API registered: 8 endpoints (list providers, provider details, update provider, active signals, history, subscriptions, stats, leaderboard)")

	// ============================================
	// Multi-Currency Wallet Management API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/wallets/currencies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			walletHandler.HandleGetCurrencies(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/wallets/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			walletHandler.HandleListTransactions(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/wallets/transfer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			walletHandler.HandleTransfer(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/wallets/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			walletHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/wallets/reconciliation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			walletHandler.HandleGetReconciliation(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/wallets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			walletHandler.HandleListWallets(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/wallets/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/adjust") {
			if r.Method == http.MethodPost {
				walletHandler.HandleAdjustBalance(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			// Single wallet details GET /admin/wallets/:clientId
			if r.Method == http.MethodGet {
				walletHandler.HandleGetWallet(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	log.Println("[WalletManagementAPI] Multi-currency wallet API registered: 8 endpoints (list wallets, wallet details, adjust balance, currencies, transactions, transfer, stats, reconciliation)")

	// ============================================
	// Account Statement / Trade Export API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/statements/generate/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			statementHandler.HandleGenerateStatement(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/statements/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			statementHandler.HandleListTemplates(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/statements/history", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			statementHandler.HandleListHistory(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/statements/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			statementHandler.HandleExport(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/statements/scheduled", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			statementHandler.HandleListScheduled(w, r)
		} else if r.Method == http.MethodPost {
			statementHandler.HandleCreateSchedule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/statements/scheduled/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			statementHandler.HandleUpdateSchedule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/statements/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			statementHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[AccountStatementAPI] Account statement API registered: 8 endpoints (generate, templates, history, export, scheduled list/create/update, stats)")

	// ============================================
	// Client Risk Scoring / Credit Assessment API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/risk-scoring/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskScoringHandler.HandleListClients(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/clients/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskScoringHandler.HandleGetClient(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/clients/recalculate/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			riskScoringHandler.HandleRecalculate(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/factors", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskScoringHandler.HandleListFactors(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/factors/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			riskScoringHandler.HandleUpdateFactor(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/distribution", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskScoringHandler.HandleGetDistribution(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskScoringHandler.HandleGetAlerts(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/risk-scoring/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskScoringHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[RiskScoringAPI] Client risk scoring API registered: 8 endpoints (list clients, client profile, recalculate, factors, update factor, distribution, alerts, stats)")

	// ============================================
	// Trading Session / Market Hours Configuration API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/sessions/market-hours", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSessionHandler.HandleListSessions(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/sessions/market-hours/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSessionHandler.HandleGetSymbolSession(w, r)
		} else if r.Method == http.MethodPut {
			tradingSessionHandler.HandleUpdateSymbolSession(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/sessions/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSessionHandler.HandleListTemplates(w, r)
		} else if r.Method == http.MethodPost {
			tradingSessionHandler.HandleCreateTemplate(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/sessions/holidays", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSessionHandler.HandleListHolidays(w, r)
		} else if r.Method == http.MethodPost {
			tradingSessionHandler.HandleCreateHoliday(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/sessions/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradingSessionHandler.HandleGetMarketStatus(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[TradingSessionAPI] Trading session API registered: 8 endpoints (list sessions, symbol session get/update, templates list/create, holidays list/create, market status)")

	// ============================================
	// Regulatory Compliance Report Generation API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/compliance/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			complianceReportHandler.HandleListReports(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/reports/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			complianceReportHandler.HandleGenerateReport(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/reports/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			complianceReportHandler.HandleListTemplates(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/reports/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			complianceReportHandler.HandleGetReport(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/checks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			complianceReportHandler.HandleListChecks(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/checks/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			complianceReportHandler.HandleRunChecks(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/breaches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			complianceReportHandler.HandleListBreaches(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/compliance/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			complianceReportHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[ComplianceReportAPI] Regulatory compliance API registered: 8 endpoints (list reports, generate, get report, templates, checks list/run, breaches, stats)")

	// ============================================
	// LP Performance Analytics / Execution Quality API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/lp-performance/overview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lpPerformanceHandler.HandleGetOverview(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/lp-performance/comparison", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lpPerformanceHandler.HandleGetComparison(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/lp-performance/execution-quality", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lpPerformanceHandler.HandleGetExecutionQuality(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/lp-performance/slippage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lpPerformanceHandler.HandleGetSlippageDistribution(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/lp-performance/latency", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lpPerformanceHandler.HandleGetLatencyPercentiles(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/lp-performance/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			lpPerformanceHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/lp-performance/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Check if it's a history endpoint
			if strings.HasSuffix(r.URL.Path, "/history") {
				lpPerformanceHandler.HandleGetHistory(w, r)
			} else {
				lpPerformanceHandler.HandleGetDetailedProfile(w, r)
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[LPPerformanceAPI] LP performance analytics API registered: 8 endpoints (overview, detailed profile, history, comparison, execution quality, slippage, latency, stats)")

	// ============================================
	// Position Aggregation / Netting Mode Management API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/positions/aggregated", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetAggregated(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/aggregated/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetSymbolBreakdown(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/netting-mode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetNettingModes(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/netting-mode/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			positionAggregationHandler.HandleUpdateNettingMode(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/exposure", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetExposure(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/largest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetLargest(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/concentration", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetConcentration(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/positions/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			positionAggregationHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[PositionAggregationAPI] Position aggregation API registered: 8 endpoints (aggregated, symbol breakdown, netting mode get/update, exposure, largest, concentration, stats)")

	// ============================================
	// Margin Monitoring API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/margin/levels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marginMonitoringHandler.HandleGetAllLevels(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/client/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marginMonitoringHandler.HandleGetClientBreakdown(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/at-risk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marginMonitoringHandler.HandleGetAtRisk(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/calls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marginMonitoringHandler.HandleGetMarginCalls(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/calls/resolve", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			marginMonitoringHandler.HandleResolveMarginCall(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/liquidations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marginMonitoringHandler.HandleGetLiquidations(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/thresholds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marginMonitoringHandler.HandleGetThresholds(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/margin/thresholds/update", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			marginMonitoringHandler.HandleUpdateThreshold(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[MarginMonitoringAPI] Margin monitoring API registered: 8 endpoints (levels, client breakdown, at-risk, calls, resolve, liquidations, thresholds get/update)")

	// ============================================
	// Execution Policy API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/execution/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			executionPolicyHandler.HandleGetPolicies(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/execution/policies/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			executionPolicyHandler.HandleGetPolicyByID(w, r)
		} else if r.Method == http.MethodPut {
			executionPolicyHandler.HandleUpdatePolicy(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/execution/slippage-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			executionPolicyHandler.HandleGetSlippageConfigs(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/execution/slippage-config/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			executionPolicyHandler.HandleUpdateSlippageConfig(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/execution/requote-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			executionPolicyHandler.HandleGetRequoteConfig(w, r)
		} else if r.Method == http.MethodPut {
			executionPolicyHandler.HandleUpdateRequoteConfig(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/execution/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			executionPolicyHandler.HandleGetExecutionStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[ExecutionPolicyAPI] Execution policy API registered: 8 endpoints (policies get/update, slippage config get/update, requote config get/update, stats)")

	// ============================================
	// Fee Schedule API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/fees/schedules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			feeScheduleHandler.HandleGetSchedules(w, r)
		} else if r.Method == http.MethodPost {
			feeScheduleHandler.HandleCreateSchedule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/fees/schedules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			feeScheduleHandler.HandleGetScheduleByID(w, r)
		} else if r.Method == http.MethodPut {
			feeScheduleHandler.HandleUpdateSchedule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/fees/client/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			feeScheduleHandler.HandleGetClientFees(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/fees/revenue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			feeScheduleHandler.HandleGetRevenue(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/fees/waivers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			feeScheduleHandler.HandleGetWaivers(w, r)
		} else if r.Method == http.MethodPost {
			feeScheduleHandler.HandleCreateWaiver(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[FeeScheduleAPI] Fee schedule API registered: 8 endpoints (schedules CRUD, client fees, revenue, waivers)")

	// ============================================
	// Client Messaging API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/messaging/inbox", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientMessagingHandler.HandleGetInbox(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/messaging/threads/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientMessagingHandler.HandleGetThread(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/messaging/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			clientMessagingHandler.HandleSendMessage(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/messaging/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientMessagingHandler.HandleGetTemplates(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/messaging/bulk", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			clientMessagingHandler.HandleSendBulkMessage(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/messaging/announcements", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientMessagingHandler.HandleGetMessagingAnnouncements(w, r)
		} else if r.Method == http.MethodPost {
			clientMessagingHandler.HandleCreateMessagingAnnouncement(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/messaging/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientMessagingHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[ClientMessagingAPI] Client messaging API registered: 8 endpoints (inbox, threads, send, templates, bulk, announcements, stats)")

	// ============================================
	// Market Data Aggregation API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/market-data/sources", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketDataAggregationHandler.HandleGetSources(w, r)
		} else if r.Method == http.MethodPost {
			marketDataAggregationHandler.HandleCreateSource(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-data/sources/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketDataAggregationHandler.HandleGetSourceByID(w, r)
		} else if r.Method == http.MethodPut {
			marketDataAggregationHandler.HandleUpdateSource(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-data/aggregation-rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketDataAggregationHandler.HandleGetAggregationRules(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-data/aggregation-rules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			marketDataAggregationHandler.HandleUpdateAggregationRule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-data/quotes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketDataAggregationHandler.HandleGetQuotes(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-data/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketDataAggregationHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[MarketDataAggregationAPI] Market data aggregation API registered: 8 endpoints (sources CRUD, aggregation rules, quotes, stats)")

	// ============================================
	// Trade Surveillance API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/surveillance/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradeSurveillanceHandler.HandleGetAlerts(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/surveillance/alerts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradeSurveillanceHandler.HandleGetAlertByID(w, r)
		} else if r.Method == http.MethodPut {
			tradeSurveillanceHandler.HandleUpdateAlert(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/surveillance/rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradeSurveillanceHandler.HandleGetRules(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/surveillance/rules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			tradeSurveillanceHandler.HandleUpdateRule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/surveillance/clients/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradeSurveillanceHandler.HandleGetClientProfile(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/surveillance/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			tradeSurveillanceHandler.HandleTriggerScan(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/surveillance/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tradeSurveillanceHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[TradeSurveillanceAPI] Trade surveillance API registered: 8 endpoints (alerts, rules, client profiles, manual scan, stats)")

	// ============================================
	// Client Portfolio Analysis API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/portfolio/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientPortfolioHandler.HandleGetAllPortfolios(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/portfolio/clients/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/history") {
			if r.Method == http.MethodGet {
				clientPortfolioHandler.HandleGetPortfolioHistory(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/patterns") {
			if r.Method == http.MethodGet {
				clientPortfolioHandler.HandleGetClientPatterns(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			clientPortfolioHandler.HandleGetPortfolioByID(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/portfolio/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientPortfolioHandler.HandleGetLeaderboard(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/portfolio/patterns", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientPortfolioHandler.HandleGetAllPatterns(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/portfolio/comparison", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientPortfolioHandler.HandleCompareClients(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/portfolio/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			clientPortfolioHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[ClientPortfolioAPI] Client portfolio analysis API registered: 8 endpoints (portfolios, history, patterns, leaderboard, comparison, stats)")

	// ============================================
	// Server Cluster Management API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/cluster/nodes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			serverClusterHandler.HandleGetAllNodes(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/cluster/nodes/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/restart") {
			if r.Method == http.MethodPost {
				serverClusterHandler.HandleRestartNode(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			serverClusterHandler.HandleGetNodeByID(w, r)
		} else if r.Method == http.MethodPut {
			serverClusterHandler.HandleUpdateNode(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/cluster/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			serverClusterHandler.HandleGetClusterHealth(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/cluster/load-balancer", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			serverClusterHandler.HandleGetLoadBalancerConfig(w, r)
		} else if r.Method == http.MethodPut {
			serverClusterHandler.HandleUpdateLoadBalancerConfig(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/cluster/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			serverClusterHandler.HandleGetClusterStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[ServerClusterAPI] Server cluster management API registered: 8 endpoints (nodes, health, load-balancer, stats)")

	// ============================================
	// Platform License Management API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/licenses", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			platformLicenseHandler.HandleGetAllLicenses(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/licenses/features", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			platformLicenseHandler.HandleGetAllFeatures(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/licenses/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			platformLicenseHandler.HandleGetLicenseStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/licenses/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/usage") {
			if r.Method == http.MethodGet {
				platformLicenseHandler.HandleGetUsageMetrics(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/billing") {
			if r.Method == http.MethodGet {
				platformLicenseHandler.HandleGetBillingHistory(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/renew") {
			if r.Method == http.MethodPost {
				platformLicenseHandler.HandleRenewLicense(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			platformLicenseHandler.HandleGetLicenseByID(w, r)
		} else if r.Method == http.MethodPut {
			platformLicenseHandler.HandleUpdateLicense(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[PlatformLicenseAPI] Platform license management API registered: 8 endpoints (licenses, features, usage, billing, renew, stats)")

	// ============================================
	// Automated Trading / EA Management API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/automated-trading/eas", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			automatedTradingHandler.HandleListEAs(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/automated-trading/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			automatedTradingHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/automated-trading/eas/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/start") {
			if r.Method == http.MethodPost {
				automatedTradingHandler.HandleStartEA(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/stop") {
			if r.Method == http.MethodPost {
				automatedTradingHandler.HandleStopEA(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/logs") {
			if r.Method == http.MethodGet {
				automatedTradingHandler.HandleGetEALogs(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/performance") {
			if r.Method == http.MethodGet {
				automatedTradingHandler.HandleGetEAPerformance(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			automatedTradingHandler.HandleGetEA(w, r)
		} else if r.Method == http.MethodPut {
			automatedTradingHandler.HandleUpdateEA(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[AutomatedTradingAPI] EA management API registered: 8 endpoints (eas, start, stop, logs, performance, stats)")

	// ============================================
	// Social Trading / Copy Trading API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/social-trading/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			socialTradingHandler.HandleListProviders(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/social-trading/copies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			socialTradingHandler.HandleListCopyRelationships(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/social-trading/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			socialTradingHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/social-trading/providers/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/performance") {
			if r.Method == http.MethodGet {
				socialTradingHandler.HandleGetProviderPerformance(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			socialTradingHandler.HandleGetProvider(w, r)
		} else if r.Method == http.MethodPut {
			socialTradingHandler.HandleUpdateProvider(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/social-trading/copies/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			socialTradingHandler.HandleGetCopyRelationship(w, r)
		} else if r.Method == http.MethodPut {
			socialTradingHandler.HandleUpdateCopySettings(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[SocialTradingAPI] Social/copy trading API registered: 8 endpoints (providers, copies, performance, stats)")

	// ============================================
	// API Gateway / Rate Limit API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/api-gateway/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			apiGatewayHandler.HandleListClients(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/api-gateway/rate-limits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			apiGatewayHandler.HandleListRateLimits(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/api-gateway/usage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			apiGatewayHandler.HandleGetUsage(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/api-gateway/ip-rules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			apiGatewayHandler.HandleListIPRules(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/api-gateway/clients/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			apiGatewayHandler.HandleGetClient(w, r)
		} else if r.Method == http.MethodPut {
			apiGatewayHandler.HandleUpdateClient(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/api-gateway/rate-limits/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			apiGatewayHandler.HandleUpdateRateLimit(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/api-gateway/ip-rules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			apiGatewayHandler.HandleUpdateIPRule(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[APIGatewayAPI] API gateway management API registered: 8 endpoints (clients, rate-limits, usage, ip-rules)")

	// ============================================
	// Market News / Economic Calendar API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/market-news/articles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketNewsHandler.HandleListArticles(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-news/calendar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketNewsHandler.HandleListCalendar(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-news/sources", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketNewsHandler.HandleListSources(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-news/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketNewsHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-news/articles/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketNewsHandler.HandleGetArticle(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-news/calendar/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			marketNewsHandler.HandleGetCalendarEvent(w, r)
		} else if r.Method == http.MethodPut {
			marketNewsHandler.HandleUpdateCalendarEvent(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/market-news/sources/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			marketNewsHandler.HandleUpdateSource(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[MarketNewsAPI] Market news and economic calendar API registered: 8 endpoints (articles, calendar, sources, stats)")

	// ============================================
	// Multi-Tenant / White Label API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/tenants/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			multiTenantHandler.HandleGetStats(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/tenants", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			multiTenantHandler.HandleListTenants(w, r)
		} else if r.Method == http.MethodPost {
			multiTenantHandler.HandleCreateTenant(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/tenants/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/status") {
			if r.Method == http.MethodPut {
				multiTenantHandler.HandleUpdateTenantStatus(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/usage") {
			if r.Method == http.MethodGet {
				multiTenantHandler.HandleGetTenantUsage(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(path, "/logs") {
			if r.Method == http.MethodGet {
				multiTenantHandler.HandleGetTenantLogs(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.Method == http.MethodGet {
			multiTenantHandler.HandleGetTenant(w, r)
		} else if r.Method == http.MethodPut {
			multiTenantHandler.HandleUpdateTenant(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[MultiTenantAPI] Multi-tenant white label management API registered: 8 endpoints (tenants, status, usage, logs, stats)")

	// ============================================
	// Spread Monitor API Routes (9 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/spreads/live", admin.HandleGetLiveSpreads)
	http.HandleFunc("/admin/spreads/stats", admin.HandleGetSpreadStats)
	http.HandleFunc("/admin/spreads/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.HandleGetSpreadAlerts(w, r)
		} else if r.Method == http.MethodPost {
			admin.HandleCreateSpreadAlert(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/spreads/alerts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			admin.HandleUpdateSpreadAlert(w, r)
		} else if r.Method == http.MethodDelete {
			admin.HandleDeleteSpreadAlert(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/spreads/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/history") {
			admin.HandleGetSpreadHistory(w, r)
		} else if strings.Contains(r.URL.Path, "/heatmap") {
			admin.HandleGetSpreadHeatmap(w, r)
		} else if strings.Contains(r.URL.Path, "/lp-comparison") {
			admin.HandleGetLPComparison(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
	log.Println("[SpreadMonitorAPI] Spread monitor API registered: 9 endpoints (live, history, stats, heatmap, lp-comparison, alerts CRUD)")

	// ============================================
	// Server Logs API Routes (5 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/logs", admin.HandleGetLogs)
	http.HandleFunc("/admin/logs/stats", admin.HandleGetLogStats)
	http.HandleFunc("/admin/logs/error-rate", admin.HandleGetErrorRate)
	http.HandleFunc("/admin/logs/top-errors", admin.HandleGetTopErrors)
	http.HandleFunc("/admin/logs/export", admin.HandleExportLogs)
	log.Println("[ServerLogsAPI] Server logs API registered: 5 endpoints (list with filters, stats, error-rate, top-errors, export)")

	// ============================================
	// MAM/PAMM API Routes (9 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/mam", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			mamHandler.HandleListGroups(w, r)
		} else if r.Method == http.MethodPost {
			mamHandler.HandleCreateGroup(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/mam/stats", mamHandler.HandleGetStats)
	http.HandleFunc("/admin/mam/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/investors") {
			// Check if it's adding/removing investor or listing
			pathParts := strings.Split(r.URL.Path, "/")
			if r.Method == http.MethodGet && len(pathParts) == 5 {
				// GET /admin/mam/:id/investors
				mamHandler.HandleGetGroupInvestors(w, r)
			} else if r.Method == http.MethodPost && len(pathParts) == 5 {
				// POST /admin/mam/:id/investors
				mamHandler.HandleAddInvestor(w, r)
			} else if r.Method == http.MethodDelete && len(pathParts) == 6 {
				// DELETE /admin/mam/:id/investors/:investorId
				mamHandler.HandleRemoveInvestor(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			// Group CRUD operations
			if r.Method == http.MethodGet {
				mamHandler.HandleGetGroup(w, r)
			} else if r.Method == http.MethodPut {
				mamHandler.HandleUpdateGroup(w, r)
			} else if r.Method == http.MethodDelete {
				mamHandler.HandleDeleteGroup(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	log.Println("[MAMAPI] MAM/PAMM API registered: 9 endpoints (list/create/update/delete groups, list/add/remove investors, stats)")

	// ============================================
	// Audit Trail API Routes (5 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/audit-trail", auditHandler.HandleListEntries)
	http.HandleFunc("/admin/audit-trail/stats", auditHandler.HandleGetStats)
	http.HandleFunc("/admin/audit-trail/admins", auditHandler.HandleGetAdmins)
	http.HandleFunc("/admin/audit-trail/export", auditHandler.HandleExport)
	http.HandleFunc("/admin/audit-trail/", auditHandler.HandleGetEntry)
	log.Println("[AuditTrailAPI] Audit trail API registered: 5 endpoints (list with filters/pagination, stats, admins list, CSV export, detail)")

	// ============================================
	// Risk Dashboard API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/risk/overview", riskHandler.HandleGetOverview)
	http.HandleFunc("/admin/risk/exposure/by-client", riskHandler.HandleGetExposureByClient)
	http.HandleFunc("/admin/risk/exposure", riskHandler.HandleGetExposure)
	http.HandleFunc("/admin/risk/concentration", riskHandler.HandleGetConcentration)
	http.HandleFunc("/admin/risk/var", riskHandler.HandleGetVaR)
	http.HandleFunc("/admin/risk/alerts", riskHandler.HandleGetAlerts)
	http.HandleFunc("/admin/risk/limits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			riskHandler.HandleGetLimits(w, r)
		} else if r.Method == http.MethodPut {
			riskHandler.HandleUpdateLimits(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[RiskDashboardAPI] Risk dashboard API registered: 8 endpoints (overview, exposure by symbol/client, concentration, VaR, limits CRUD, alerts)")

	// ============================================
	// Customer Lifecycle Analytics API Routes (7 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/lifecycle/overview", lifecycleHandler.HandleGetOverview)
	http.HandleFunc("/admin/lifecycle/funnel", lifecycleHandler.HandleGetFunnel)
	http.HandleFunc("/admin/lifecycle/retention", lifecycleHandler.HandleGetRetention)
	http.HandleFunc("/admin/lifecycle/at-risk", lifecycleHandler.HandleGetAtRisk)
	http.HandleFunc("/admin/lifecycle/segments", lifecycleHandler.HandleGetSegments)
	http.HandleFunc("/admin/lifecycle/trends", lifecycleHandler.HandleGetTrends)
	http.HandleFunc("/admin/lifecycle/client/", lifecycleHandler.HandleGetClientTimeline)
	log.Println("[LifecycleAPI] Customer lifecycle analytics API registered: 7 endpoints (overview, funnel, retention cohorts, at-risk clients, segments, trends, client timeline)")

	// ============================================
	// Paper Trading / Demo Account API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/demo-accounts/stats", paperTradingHandler.HandleGetStats)
	http.HandleFunc("/admin/demo-accounts/conversions", paperTradingHandler.HandleGetConversions)
	http.HandleFunc("/admin/demo-accounts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			paperTradingHandler.HandleListAccounts(w, r)
		} else if r.Method == http.MethodPost {
			paperTradingHandler.HandleCreateAccount(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/demo-accounts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/positions") {
			paperTradingHandler.HandleGetPositions(w, r)
		} else {
			if r.Method == http.MethodGet {
				paperTradingHandler.HandleGetAccount(w, r)
			} else if r.Method == http.MethodPut {
				paperTradingHandler.HandleUpdateAccount(w, r)
			} else if r.Method == http.MethodDelete {
				paperTradingHandler.HandleDeleteAccount(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	log.Println("[PaperTradingAPI] Paper trading API registered: 8 endpoints (list/create/update/delete accounts, positions, stats, conversions)")

	// ============================================
	// Scheduled Reports / Auto-Export API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/reports/templates", scheduledReportsHandler.HandleGetTemplates)
	http.HandleFunc("/admin/reports/executions", scheduledReportsHandler.HandleGetExecutions)
	http.HandleFunc("/admin/reports/scheduled", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			scheduledReportsHandler.HandleListReports(w, r)
		} else if r.Method == http.MethodPost {
			scheduledReportsHandler.HandleCreateReport(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/reports/scheduled/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/run") {
			scheduledReportsHandler.HandleRunReport(w, r)
		} else {
			if r.Method == http.MethodGet {
				scheduledReportsHandler.HandleGetReport(w, r)
			} else if r.Method == http.MethodPut {
				scheduledReportsHandler.HandleUpdateReport(w, r)
			} else if r.Method == http.MethodDelete {
				scheduledReportsHandler.HandleDeleteReport(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	log.Println("[ScheduledReportsAPI] Scheduled reports API registered: 8 endpoints (list/create/update/delete reports, trigger run, executions, templates)")

	// ============================================
	// Fraud Detection / IP Geolocation API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/fraud/logins", fraudDetectionHandler.HandleGetLogins)
	http.HandleFunc("/admin/fraud/logins/", fraudDetectionHandler.HandleGetClientLogins)
	http.HandleFunc("/admin/fraud/alerts", fraudDetectionHandler.HandleGetAlerts)
	http.HandleFunc("/admin/fraud/alerts/resolve/", fraudDetectionHandler.HandleResolveAlert)
	http.HandleFunc("/admin/fraud/rules", fraudDetectionHandler.HandleGetRules)
	http.HandleFunc("/admin/fraud/rules/", fraudDetectionHandler.HandleUpdateRule)
	http.HandleFunc("/admin/fraud/geo", fraudDetectionHandler.HandleGetGeoData)
	http.HandleFunc("/admin/fraud/stats", fraudDetectionHandler.HandleGetStats)
	log.Println("[FraudDetectionAPI] Fraud detection API registered: 8 endpoints (logins, client logins, alerts, resolve alert, rules CRUD, geo data, stats)")

	// ============================================
	// Backup Manager / Recovery API Routes (9 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/backups", backupHandler.HandleListBackups)
	http.HandleFunc("/admin/backups/create", backupHandler.HandleCreateBackup)
	http.HandleFunc("/admin/backups/schedules", backupHandler.HandleListSchedules)
	http.HandleFunc("/admin/backups/components", backupHandler.HandleGetComponents)
	http.HandleFunc("/admin/backups/stats", backupHandler.HandleGetStats)
	http.HandleFunc("/admin/backups/schedules/", backupHandler.HandleUpdateSchedule)
	http.HandleFunc("/admin/backups/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/restore") {
			backupHandler.HandleRestoreBackup(w, r)
		} else if r.Method == http.MethodGet {
			backupHandler.HandleGetBackup(w, r)
		} else if r.Method == http.MethodDelete {
			backupHandler.HandleDeleteBackup(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	log.Println("[BackupManagerAPI] Backup manager API registered: 9 endpoints (list, create, details, delete, schedules CRUD, components, restore, stats)")

	// ============================================
	// Referral Program API Routes (8 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/referrals/links", referralHandler.HandleListLinks)
	http.HandleFunc("/admin/referrals/list", referralHandler.HandleListReferrals)
	http.HandleFunc("/admin/referrals/tiers", referralHandler.HandleListTiers)
	http.HandleFunc("/admin/referrals/stats", referralHandler.HandleGetStats)
	http.HandleFunc("/admin/referrals/top-referrers", referralHandler.HandleGetTopReferrers)
	http.HandleFunc("/admin/referrals/links/", referralHandler.HandleGetLink)
	http.HandleFunc("/admin/referrals/tiers/", referralHandler.HandleUpdateTier)
	http.HandleFunc("/admin/referrals/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/approve") {
			referralHandler.HandleApproveReferral(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
	log.Println("[ReferralProgramAPI] Referral program API registered: 8 endpoints (links, link details, referrals list, approve reward, tiers CRUD, stats, top referrers)")

	// ============================================
	// Instrument Groups / Symbol Category API Routes (9 endpoints - admin auth)
	// ============================================
	http.HandleFunc("/admin/instruments/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			instrumentGroupHandler.HandleListGroups(w, r)
		} else if r.Method == http.MethodPost {
			instrumentGroupHandler.HandleCreateGroup(w, r)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/admin/instruments/stats", instrumentGroupHandler.HandleGetStats)
	http.HandleFunc("/admin/instruments/groups/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/symbols") {
			if strings.Count(r.URL.Path, "/") == 5 { // /admin/instruments/groups/:id/symbols
				if r.Method == http.MethodGet {
					instrumentGroupHandler.HandleGetGroupSymbols(w, r)
				} else if r.Method == http.MethodPost {
					instrumentGroupHandler.HandleAddSymbol(w, r)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if strings.Count(r.URL.Path, "/") == 6 { // /admin/instruments/groups/:id/symbols/:symbol
				instrumentGroupHandler.HandleRemoveSymbol(w, r)
			}
		} else {
			if r.Method == http.MethodGet {
				instrumentGroupHandler.HandleGetGroup(w, r)
			} else if r.Method == http.MethodPut {
				instrumentGroupHandler.HandleUpdateGroup(w, r)
			} else if r.Method == http.MethodDelete {
				instrumentGroupHandler.HandleDeleteGroup(w, r)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
	log.Println("[InstrumentGroupsAPI] Instrument groups API registered: 9 endpoints (list, create, details, update, delete groups, list/add/remove symbols, stats)")

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

	// Create HTTP server with middleware chain
	var handler http.Handler
	if rateLimiter != nil {
		// Wrap all routes with rate limiting (except excluded paths)
		handler = rateLimiter.MiddlewareWithExclusions(rateLimitConfig.Exclusions)(http.DefaultServeMux)
	} else {
		handler = http.DefaultServeMux
	}

	// Apply global CORS middleware to ALL routes
	corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
	})
	handler = corsMiddleware(handler)
	log.Printf("[CORS] Global CORS middleware enabled for origins: %v", cfg.CORS.AllowedOrigins)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}

