package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/epic1st/rtx/backend/datapipeline"
	"github.com/epic1st/rtx/backend/lpmanager"
	"github.com/epic1st/rtx/backend/ws"
)

// This example demonstrates how to integrate the new data pipeline
// with the existing LP Manager and WebSocket system

func main() {
	log.Println("=== Real-Time Data Pipeline Integration Example ===")

	// 1. Initialize the new data pipeline
	pipelineConfig := datapipeline.DefaultPipelineConfig()
	pipelineConfig.RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	pipelineConfig.WorkerCount = 4
	pipelineConfig.TickBufferSize = 10000
	pipelineConfig.EnableDeduplication = true

	pipeline, err := datapipeline.NewPipeline(pipelineConfig)
	if err != nil {
		log.Fatalf("Failed to create pipeline: %v", err)
	}

	// Start the pipeline
	if err := pipeline.Start(); err != nil {
		log.Fatalf("Failed to start pipeline: %v", err)
	}

	log.Println("[Pipeline] Started successfully")

	// 2. Create integration adapter
	adapter := datapipeline.NewIntegrationAdapter(pipeline)

	// Start monitoring
	adapter.StartMonitoring()

	// 3. Initialize existing LP Manager
	lpMgr := lpmanager.NewManager("data/lp_config.json")

	// Load config (simplified - actual implementation may vary)
	if err := lpMgr.LoadConfig(); err != nil {
		log.Printf("[LPManager] Failed to load config: %v", err)
	}

	// 4. Create a bridge between LP Manager and Pipeline
	go bridgeLPManagerToPipeline(lpMgr, adapter)

	// 5. Initialize WebSocket hub (existing)
	hub := ws.NewHub()
	go hub.Run()

	// 6. Bridge pipeline to WebSocket hub
	go bridgePipelineToWebSocket(pipeline, hub)

	// 7. Setup HTTP API with both old and new endpoints
	mux := http.NewServeMux()

	// Register new pipeline API
	apiHandler := datapipeline.NewAPIHandler(pipeline)
	apiHandler.RegisterRoutes(mux)

	// Legacy endpoints (example)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start HTTP server
	server := &http.Server{
		Addr:    ":7999",
		Handler: mux,
	}

	go func() {
		log.Println("[HTTP] Server listening on :7999")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 8. Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("\n[Main] Shutdown signal received, gracefully stopping...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := pipeline.Stop(); err != nil {
		log.Printf("Pipeline shutdown error: %v", err)
	}

	log.Println("[Main] Shutdown complete")
}

// bridgeLPManagerToPipeline forwards quotes from LP Manager to the pipeline
func bridgeLPManagerToPipeline(lpMgr *lpmanager.Manager, adapter *datapipeline.IntegrationAdapter) {
	log.Println("[Bridge] Starting LP Manager -> Pipeline bridge")

	// Get quotes channel from LP Manager
	quotesChan := lpMgr.GetQuotesChan()

	var quoteCount int64
	for quote := range quotesChan {
		quoteCount++

		// Process quote through adapter
		err := adapter.ProcessLPQuote(
			quote.LP,
			quote.Symbol,
			quote.Bid,
			quote.Ask,
			quote.Timestamp,
		)

		if err != nil && quoteCount%1000 == 0 {
			log.Printf("[Bridge] Warning: Failed to process quote: %v", err)
		}

		if quoteCount%10000 == 0 {
			log.Printf("[Bridge] Processed %d quotes from LP Manager", quoteCount)
		}
	}

	log.Println("[Bridge] LP Manager bridge closed")
}

// bridgePipelineToWebSocket forwards pipeline ticks to WebSocket hub
func bridgePipelineToWebSocket(pipeline *datapipeline.MarketDataPipeline, hub *ws.Hub) {
	log.Println("[Bridge] Starting Pipeline -> WebSocket bridge")

	// Subscribe to Redis pub/sub for all quotes
	_ = pipeline.GetDistributor() // TODO: Use distributor for proper Redis pub/sub

	// This is a simplified example - in production, use proper Redis subscription
	// For now, we'll poll for latest quotes and broadcast them

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	storage := pipeline.GetStorageManager()
	processedSymbols := make(map[string]time.Time)

	for range ticker.C {
		// This is a demo - in production, use Redis pub/sub properly
		// Get list of symbols from storage and broadcast latest ticks

		// For demonstration, we'll just broadcast if we have new data
		// In production, you'd subscribe to Redis pub/sub channels

		// Example: broadcast latest EURUSD tick
		symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "BTCUSD", "ETHUSD"}

		for _, symbol := range symbols {
			tick, err := storage.GetLatestTick(symbol)
			if err != nil {
				continue
			}

			// Check if we already broadcast this tick
			lastTime, exists := processedSymbols[symbol]
			if exists && !tick.Timestamp.After(lastTime) {
				continue // Already broadcast
			}

			// Convert to WebSocket MarketTick format
			wsTick := &ws.MarketTick{
				Type:      "tick",
				Symbol:    tick.Symbol,
				Bid:       tick.Bid,
				Ask:       tick.Ask,
				Spread:    tick.Spread,
				Timestamp: tick.Timestamp.Unix(),
				LP:        tick.Source,
			}

			// Broadcast to all WebSocket clients
			hub.BroadcastTick(wsTick)

			// Update last processed time
			processedSymbols[symbol] = tick.Timestamp
		}
	}
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
