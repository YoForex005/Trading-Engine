package datapipeline

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// BenchmarkTickIngestion benchmarks raw tick ingestion throughput
func BenchmarkTickIngestion(b *testing.B) {
	config := DefaultPipelineConfig()
	config.RedisAddr = "localhost:6379"
	config.EnableDeduplication = false // Disable for max throughput test

	pipeline, err := NewPipeline(config)
	if err != nil {
		b.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		b.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		symbol := symbols[i%len(symbols)]
		tick := &RawTick{
			Source:    "BENCH",
			Symbol:    symbol,
			Bid:       1.08450 + rand.Float64()*0.001,
			Ask:       1.08452 + rand.Float64()*0.001,
			Timestamp: time.Now().Unix(),
		}

		pipeline.IngestTick(tick)
	}

	b.StopTimer()

	stats := pipeline.GetStats()
	b.Logf("Processed: %d, Dropped: %d, Invalid: %d",
		stats.TicksProcessed, stats.TicksDropped, stats.TicksInvalid)
}

// BenchmarkOHLCGeneration benchmarks OHLC bar generation
func BenchmarkOHLCGeneration(b *testing.B) {
	config := DefaultPipelineConfig()
	config.RedisAddr = "localhost:6379"

	pipeline, err := NewPipeline(config)
	if err != nil {
		b.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		b.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	// Generate normalized ticks
	symbol := "EURUSD"
	baseTime := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tick := &NormalizedTick{
			Symbol:    symbol,
			Bid:       1.08450 + rand.Float64()*0.001,
			Ask:       1.08452 + rand.Float64()*0.001,
			Spread:    0.00002,
			Timestamp: baseTime.Add(time.Duration(i) * time.Millisecond),
			Source:    "BENCH",
			TickID:    fmt.Sprintf("bench-%d", i),
		}

		pipeline.ohlcEngine.ProcessTick(tick)
	}

	b.StopTimer()

	stats := pipeline.GetStats()
	b.Logf("OHLC Bars Generated: %d", stats.OHLCBarsGenerated)
}

// BenchmarkNormalization benchmarks tick normalization
func BenchmarkNormalization(b *testing.B) {
	ingester := NewDataIngester(DefaultPipelineConfig(), &PipelineStats{})

	rawTick := &RawTick{
		Source:    "TEST",
		Symbol:    "EURUSD",
		Bid:       1.08450,
		Ask:       1.08452,
		Timestamp: time.Now().Unix(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ingester.normalizeTick(rawTick)
	}
}

// BenchmarkDeduplication benchmarks tick deduplication
func BenchmarkDeduplication(b *testing.B) {
	config := DefaultPipelineConfig()
	config.EnableDeduplication = true

	ingester := NewDataIngester(config, &PipelineStats{})
	ingester.ctx = context.Background()

	tick := &NormalizedTick{
		Symbol:    "EURUSD",
		Bid:       1.08450,
		Ask:       1.08452,
		Spread:    0.00002,
		Timestamp: time.Now(),
		Source:    "TEST",
		TickID:    "test-123",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Change tick ID each time for realistic test
		tick.TickID = fmt.Sprintf("test-%d", i)
		ingester.isDuplicate(tick)
	}
}

// BenchmarkRedisStorage benchmarks Redis storage operations
func BenchmarkRedisStorage(b *testing.B) {
	// Skip if Redis not available
	config := DefaultPipelineConfig()
	pipeline, err := NewPipeline(config)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}

	if err := pipeline.Start(); err != nil {
		b.Skipf("Failed to start pipeline: %v", err)
		return
	}
	defer pipeline.Stop()

	storage := pipeline.storage

	tick := &NormalizedTick{
		Symbol:     "EURUSD",
		Bid:        1.08450,
		Ask:        1.08452,
		Spread:     0.00002,
		Timestamp:  time.Now(),
		Source:     "TEST",
		TickID:     "bench-tick",
		ReceivedAt: time.Now(),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tick.Timestamp = time.Now().Add(time.Duration(i) * time.Millisecond)
		tick.TickID = fmt.Sprintf("bench-%d", i)
		storage.StoreTick(tick)
	}
}

// BenchmarkFullPipeline benchmarks end-to-end pipeline
func BenchmarkFullPipeline(b *testing.B) {
	config := DefaultPipelineConfig()
	config.WorkerCount = 8
	config.TickBufferSize = 50000

	pipeline, err := NewPipeline(config)
	if err != nil {
		b.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		b.Fatalf("Failed to start pipeline: %v", err)
	}
	defer pipeline.Stop()

	symbols := []string{
		"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD",
		"BTCUSD", "ETHUSD", "XRPUSD", "BNBUSD", "SOLUSD",
	}

	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()

	for i := 0; i < b.N; i++ {
		symbol := symbols[i%len(symbols)]
		tick := &RawTick{
			Source:    "BENCH",
			Symbol:    symbol,
			Bid:       1000.0 + rand.Float64()*100,
			Ask:       1000.2 + rand.Float64()*100,
			Timestamp: time.Now().Unix(),
		}

		pipeline.IngestTick(tick)
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	elapsed := time.Since(startTime)
	b.StopTimer()

	stats := pipeline.GetStats()

	throughput := float64(b.N) / elapsed.Seconds()

	b.Logf("=== Pipeline Benchmark Results ===")
	b.Logf("Total Ticks: %d", b.N)
	b.Logf("Duration: %v", elapsed)
	b.Logf("Throughput: %.0f ticks/sec", throughput)
	b.Logf("Received: %d", stats.TicksReceived)
	b.Logf("Processed: %d", stats.TicksProcessed)
	b.Logf("Dropped: %d", stats.TicksDropped)
	b.Logf("Invalid: %d", stats.TicksInvalid)
	b.Logf("Duplicates: %d", stats.TicksDuplicate)
	b.Logf("OHLC Bars: %d", stats.OHLCBarsGenerated)
	b.Logf("Avg Tick Latency: %.2fms", stats.AvgTickLatencyMs)
	b.Logf("Avg OHLC Latency: %.2fms", stats.AvgOHLCLatencyMs)
	b.Logf("Avg Distribution Latency: %.2fms", stats.AvgDistributionLatencyMs)
}

// TestTickNormalization tests tick normalization
func TestTickNormalization(t *testing.T) {
	ingester := NewDataIngester(DefaultPipelineConfig(), &PipelineStats{})

	tests := []struct {
		name      string
		rawTick   *RawTick
		expectErr bool
	}{
		{
			name: "Valid Unix timestamp",
			rawTick: &RawTick{
				Source:    "TEST",
				Symbol:    "EURUSD",
				Bid:       1.08450,
				Ask:       1.08452,
				Timestamp: time.Now().Unix(),
			},
			expectErr: false,
		},
		{
			name: "Valid ISO8601 timestamp",
			rawTick: &RawTick{
				Source:    "TEST",
				Symbol:    "EURUSD",
				Bid:       1.08450,
				Ask:       1.08452,
				Timestamp: time.Now().Format(time.RFC3339),
			},
			expectErr: false,
		},
		{
			name: "Invalid bid/ask (bid > ask)",
			rawTick: &RawTick{
				Source:    "TEST",
				Symbol:    "EURUSD",
				Bid:       1.08452,
				Ask:       1.08450,
				Timestamp: time.Now().Unix(),
			},
			expectErr: true,
		},
		{
			name: "Invalid prices (negative)",
			rawTick: &RawTick{
				Source:    "TEST",
				Symbol:    "EURUSD",
				Bid:       -1.08450,
				Ask:       1.08452,
				Timestamp: time.Now().Unix(),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := ingester.normalizeTick(tt.rawTick)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if normalized == nil {
					t.Errorf("Expected normalized tick but got nil")
				}
			}
		})
	}
}

// TestOHLCAlignment tests OHLC bar time alignment
func TestOHLCAlignment(t *testing.T) {
	engine := NewOHLCEngine(DefaultPipelineConfig(), &PipelineStats{})

	tests := []struct {
		name      string
		timestamp time.Time
		timeframe Timeframe
		expected  time.Time
	}{
		{
			name:      "M1 alignment",
			timestamp: time.Date(2024, 1, 1, 12, 34, 56, 0, time.UTC),
			timeframe: TF_M1,
			expected:  time.Date(2024, 1, 1, 12, 34, 0, 0, time.UTC),
		},
		{
			name:      "H1 alignment",
			timestamp: time.Date(2024, 1, 1, 12, 34, 56, 0, time.UTC),
			timeframe: TF_H1,
			expected:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name:      "D1 alignment",
			timestamp: time.Date(2024, 1, 1, 12, 34, 56, 0, time.UTC),
			timeframe: TF_D1,
			expected:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aligned := engine.alignTimestamp(tt.timestamp, tt.timeframe)
			if !aligned.Equal(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, aligned)
			}
		})
	}
}
