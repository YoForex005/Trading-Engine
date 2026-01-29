package datapipeline

import (
	"log"
	"time"
)

// IntegrationAdapter adapts the existing LP manager to the new pipeline
type IntegrationAdapter struct {
	pipeline *MarketDataPipeline
}

// NewIntegrationAdapter creates a new integration adapter
func NewIntegrationAdapter(pipeline *MarketDataPipeline) *IntegrationAdapter {
	return &IntegrationAdapter{
		pipeline: pipeline,
	}
}

// ProcessLPQuote converts an LP quote to a raw tick and ingests it
func (a *IntegrationAdapter) ProcessLPQuote(lpName, symbol string, bid, ask float64, timestamp interface{}) error {
	rawTick := &RawTick{
		Source:    lpName,
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Timestamp: timestamp,
	}

	return a.pipeline.IngestTick(rawTick)
}

// ProcessMarketTick adapts existing MarketTick format (from ws.MarketTick)
func (a *IntegrationAdapter) ProcessMarketTick(marketTick *MarketTickCompat) error {
	rawTick := &RawTick{
		Source:    marketTick.LP,
		Symbol:    marketTick.Symbol,
		Bid:       marketTick.Bid,
		Ask:       marketTick.Ask,
		Timestamp: marketTick.Timestamp,
	}

	return a.pipeline.IngestTick(rawTick)
}

// MarketTickCompat represents the existing MarketTick structure for compatibility
type MarketTickCompat struct {
	Type      string
	Symbol    string
	Bid       float64
	Ask       float64
	Spread    float64
	Timestamp int64
	LP        string
}

// StartMonitoring starts pipeline monitoring updates
func (a *IntegrationAdapter) StartMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := a.pipeline.GetStats()

			log.Printf("[Pipeline Stats] Received: %d | Processed: %d | Dropped: %d | OHLC Bars: %d | Clients: %d",
				stats.TicksReceived,
				stats.TicksProcessed,
				stats.TicksDropped,
				stats.OHLCBarsGenerated,
				stats.ClientsConnected)

			log.Printf("[Pipeline Latency] Tick: %.2fms | OHLC: %.2fms | Distribution: %.2fms",
				stats.AvgTickLatencyMs,
				stats.AvgOHLCLatencyMs,
				stats.AvgDistributionLatencyMs)

			// Check for issues
			if stats.TicksDropped > 100 {
				log.Printf("[Pipeline WARN] High drop rate: %d ticks dropped", stats.TicksDropped)
			}

			if stats.StaleFeedsDetected > 0 {
				log.Printf("[Pipeline WARN] %d stale feeds detected", stats.StaleFeedsDetected)
			}
		}
	}()
}

// GetPipelineStats returns current pipeline statistics (for admin dashboard)
func (a *IntegrationAdapter) GetPipelineStats() PipelineStats {
	return a.pipeline.GetStats()
}

// GetFeedHealth returns feed health information
func (a *IntegrationAdapter) GetFeedHealth() map[string]FeedHealth {
	return a.pipeline.monitor.GetFeedHealth()
}

// SubscribeClient subscribes a WebSocket client to a symbol
func (a *IntegrationAdapter) SubscribeClient(clientID, symbol string) error {
	return a.pipeline.distributor.Subscribe(clientID, symbol)
}

// UnsubscribeClient unsubscribes a WebSocket client from a symbol
func (a *IntegrationAdapter) UnsubscribeClient(clientID, symbol string) error {
	return a.pipeline.distributor.Unsubscribe(clientID, symbol)
}

// DisconnectClient removes all subscriptions for a client
func (a *IntegrationAdapter) DisconnectClient(clientID string) {
	a.pipeline.distributor.DisconnectClient(clientID)
}
