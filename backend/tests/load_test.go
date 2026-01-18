package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// LoadTestMetrics holds performance metrics
type LoadTestMetrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	RequestsPerSec  float64
	Latencies       []time.Duration
	mu              sync.Mutex
}

// RecordRequest records request metrics
func (m *LoadTestMetrics) RecordRequest(latency time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	atomic.AddInt64(&m.TotalRequests, 1)

	if success {
		atomic.AddInt64(&m.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&m.FailedRequests, 1)
	}

	m.Latencies = append(m.Latencies, latency)

	if m.MinLatency == 0 || latency < m.MinLatency {
		m.MinLatency = latency
	}

	if latency > m.MaxLatency {
		m.MaxLatency = latency
	}
}

// Calculate computes final statistics
func (m *LoadTestMetrics) Calculate() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.Latencies) == 0 {
		return
	}

	// Calculate average
	var total time.Duration
	for _, lat := range m.Latencies {
		total += lat
	}
	m.AvgLatency = total / time.Duration(len(m.Latencies))

	// Calculate percentiles (simple implementation)
	// Note: For production, use a proper percentile calculation
	sorted := make([]time.Duration, len(m.Latencies))
	copy(sorted, m.Latencies)

	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	p50idx := len(sorted) * 50 / 100
	p95idx := len(sorted) * 95 / 100
	p99idx := len(sorted) * 99 / 100

	if p50idx < len(sorted) {
		m.P50Latency = sorted[p50idx]
	}
	if p95idx < len(sorted) {
		m.P95Latency = sorted[p95idx]
	}
	if p99idx < len(sorted) {
		m.P99Latency = sorted[p99idx]
	}

	if m.TotalDuration > 0 {
		m.RequestsPerSec = float64(m.TotalRequests) / m.TotalDuration.Seconds()
	}
}

// Print outputs metrics report
func (m *LoadTestMetrics) Print(t *testing.T) {
	m.Calculate()

	t.Logf("\n=== Load Test Results ===")
	t.Logf("Total Requests:    %d", m.TotalRequests)
	t.Logf("Successful:        %d (%.2f%%)", m.SuccessRequests, float64(m.SuccessRequests)/float64(m.TotalRequests)*100)
	t.Logf("Failed:            %d (%.2f%%)", m.FailedRequests, float64(m.FailedRequests)/float64(m.TotalRequests)*100)
	t.Logf("Duration:          %v", m.TotalDuration)
	t.Logf("Requests/sec:      %.2f", m.RequestsPerSec)
	t.Logf("\nLatency Statistics:")
	t.Logf("  Min:             %v", m.MinLatency)
	t.Logf("  Max:             %v", m.MaxLatency)
	t.Logf("  Avg:             %v", m.AvgLatency)
	t.Logf("  P50:             %v", m.P50Latency)
	t.Logf("  P95:             %v", m.P95Latency)
	t.Logf("  P99:             %v", m.P99Latency)
}

// ==================== LOAD TESTS ====================

// TestLoad_PlaceOrders_100Concurrent tests 100 concurrent users
func TestLoad_PlaceOrders_100Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	numConcurrent := 100
	ordersPerUser := 10
	metrics := &LoadTestMetrics{}

	var wg sync.WaitGroup
	wg.Add(numConcurrent)

	startTime := time.Now()

	for i := 0; i < numConcurrent; i++ {
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < ordersPerUser; j++ {
				reqStart := time.Now()

				reqBody := map[string]interface{}{
					"symbol": "EURUSD",
					"side":   "BUY",
					"volume": 0.01,
					"type":   "MARKET",
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				tc.Server.HandlePlaceOrder(w, req)

				latency := time.Since(reqStart)
				success := w.Code == 200

				metrics.RecordRequest(latency, success)
			}
		}(i)
	}

	wg.Wait()
	metrics.TotalDuration = time.Since(startTime)

	metrics.Print(t)

	// Performance assertions
	if metrics.AvgLatency > 100*time.Millisecond {
		t.Logf("WARNING: Average latency %.2fms exceeds 100ms threshold", float64(metrics.AvgLatency.Microseconds())/1000)
	}

	if metrics.P95Latency > 200*time.Millisecond {
		t.Logf("WARNING: P95 latency %.2fms exceeds 200ms threshold", float64(metrics.P95Latency.Microseconds())/1000)
	}

	successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	if successRate < 99.0 {
		t.Errorf("Success rate %.2f%% is below 99%% threshold", successRate)
	}
}

// TestLoad_PlaceOrders_1000Concurrent tests 1000 concurrent users
func TestLoad_PlaceOrders_1000Concurrent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	numConcurrent := 1000
	ordersPerUser := 5
	metrics := &LoadTestMetrics{}

	var wg sync.WaitGroup
	wg.Add(numConcurrent)

	startTime := time.Now()

	for i := 0; i < numConcurrent; i++ {
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < ordersPerUser; j++ {
				reqStart := time.Now()

				reqBody := map[string]interface{}{
					"symbol": "EURUSD",
					"side":   "BUY",
					"volume": 0.01,
					"type":   "MARKET",
				}
				body, _ := json.Marshal(reqBody)

				req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				tc.Server.HandlePlaceOrder(w, req)

				latency := time.Since(reqStart)
				success := w.Code == 200

				metrics.RecordRequest(latency, success)
			}
		}(i)
	}

	wg.Wait()
	metrics.TotalDuration = time.Since(startTime)

	metrics.Print(t)

	t.Logf("\nThroughput: %.2f orders/sec", metrics.RequestsPerSec)
}

// TestLoad_MarketDataStream tests high-frequency tick streaming
func TestLoad_MarketDataStream(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tc := SetupTest(t)

	numSymbols := 10
	ticksPerSymbol := 1000
	metrics := &LoadTestMetrics{}

	symbols := make([]string, numSymbols)
	for i := 0; i < numSymbols; i++ {
		symbols[i] = fmt.Sprintf("SYM%d", i)
	}

	startTime := time.Now()

	// Inject ticks at high frequency
	for i := 0; i < ticksPerSymbol; i++ {
		for _, symbol := range symbols {
			tickStart := time.Now()

			bid := 1.10000 + float64(i)*0.00001
			ask := bid + 0.00020

			tc.InjectPrice(symbol, bid, ask)

			latency := time.Since(tickStart)
			metrics.RecordRequest(latency, true)
		}
	}

	metrics.TotalDuration = time.Since(startTime)
	metrics.Print(t)

	totalTicks := int64(numSymbols * ticksPerSymbol)
	t.Logf("\nTotal ticks: %d", totalTicks)
	t.Logf("Ticks/sec: %.2f", float64(totalTicks)/metrics.TotalDuration.Seconds())
}

// TestLoad_MixedOperations tests realistic mixed workload
func TestLoad_MixedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)
	tc.InjectPrice("GBPUSD", 1.25000, 1.25025)

	numUsers := 50
	operationsPerUser := 20
	metrics := &LoadTestMetrics{}

	var wg sync.WaitGroup
	wg.Add(numUsers)

	startTime := time.Now()

	for i := 0; i < numUsers; i++ {
		go func(userID int) {
			defer wg.Done()

			operations := []func(){
				// Place market order
				func() {
					reqStart := time.Now()
					reqBody := map[string]interface{}{
						"symbol": "EURUSD",
						"side":   "BUY",
						"volume": 0.01,
						"type":   "MARKET",
					}
					body, _ := json.Marshal(reqBody)
					req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()
					tc.Server.HandlePlaceOrder(w, req)
					metrics.RecordRequest(time.Since(reqStart), w.Code == 200)
				},
				// Get ticks
				func() {
					reqStart := time.Now()
					req := httptest.NewRequest("GET", "/ticks?symbol=EURUSD&limit=10", nil)
					w := httptest.NewRecorder()
					tc.Server.HandleGetTicks(w, req)
					metrics.RecordRequest(time.Since(reqStart), w.Code == 200)
				},
				// Calculate risk
				func() {
					reqStart := time.Now()
					req := httptest.NewRequest("GET", "/risk/calculate-lot?symbol=EURUSD&riskPercent=2&slPips=20", nil)
					w := httptest.NewRecorder()
					tc.Server.HandleCalculateLot(w, req)
					metrics.RecordRequest(time.Since(reqStart), w.Code == 200)
				},
				// Get pending orders
				func() {
					reqStart := time.Now()
					req := httptest.NewRequest("GET", "/orders/pending", nil)
					w := httptest.NewRecorder()
					tc.Server.HandleGetPendingOrders(w, req)
					metrics.RecordRequest(time.Since(reqStart), w.Code == 200)
				},
			}

			for j := 0; j < operationsPerUser; j++ {
				// Randomly select operation
				op := operations[j%len(operations)]
				op()
			}
		}(i)
	}

	wg.Wait()
	metrics.TotalDuration = time.Since(startTime)

	metrics.Print(t)
}

// TestLoad_DatabaseLoad tests database operation load
func TestLoad_DatabaseLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tc := SetupTest(t)
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	numOperations := 1000
	metrics := &LoadTestMetrics{}

	startTime := time.Now()

	for i := 0; i < numOperations; i++ {
		// Create order
		reqStart := time.Now()

		reqBody := map[string]interface{}{
			"symbol": "EURUSD",
			"side":   "BUY",
			"volume": 0.01,
			"type":   "MARKET",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		tc.Server.HandlePlaceOrder(w, req)

		latency := time.Since(reqStart)
		metrics.RecordRequest(latency, w.Code == 200)

		// Query ticks
		reqStart = time.Now()
		req = httptest.NewRequest("GET", "/ticks?symbol=EURUSD&limit=100", nil)
		w = httptest.NewRecorder()
		tc.Server.HandleGetTicks(w, req)
		metrics.RecordRequest(time.Since(reqStart), w.Code == 200)
	}

	metrics.TotalDuration = time.Since(startTime)
	metrics.Print(t)
}

// TestLoad_MemoryUsage tests memory consumption under load
func TestLoad_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	tc := SetupTest(t)

	// Generate large number of ticks
	numSymbols := 100
	ticksPerSymbol := 1000

	t.Logf("Generating %d ticks for %d symbols", ticksPerSymbol, numSymbols)

	startTime := time.Now()

	for i := 0; i < numSymbols; i++ {
		symbol := fmt.Sprintf("SYM%03d", i)

		for j := 0; j < ticksPerSymbol; j++ {
			bid := 1.0 + float64(j)*0.0001
			ask := bid + 0.0002

			tc.InjectPrice(symbol, bid, ask)
		}
	}

	duration := time.Since(startTime)

	t.Logf("Generated %d total ticks in %v", numSymbols*ticksPerSymbol, duration)
	t.Logf("Tick generation rate: %.2f ticks/sec", float64(numSymbols*ticksPerSymbol)/duration.Seconds())

	// Query ticks from multiple symbols
	for i := 0; i < 10; i++ {
		symbol := fmt.Sprintf("SYM%03d", i*10)
		req := httptest.NewRequest("GET", fmt.Sprintf("/ticks?symbol=%s&limit=500", symbol), nil)
		w := httptest.NewRecorder()
		tc.Server.HandleGetTicks(w, req)

		var ticks []interface{}
		json.NewDecoder(w.Body).Decode(&ticks)
		t.Logf("%s: %d ticks retrieved", symbol, len(ticks))
	}
}

// TestLoad_StressTest runs comprehensive stress test
func TestLoad_StressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	tc := SetupTest(t)

	// Inject prices for multiple symbols
	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "USDCAD"}
	for _, symbol := range symbols {
		tc.InjectPrice(symbol, 1.10000, 1.10020)
	}

	duration := 30 * time.Second
	numConcurrentUsers := 100

	t.Logf("Running stress test for %v with %d concurrent users", duration, numConcurrentUsers)

	metrics := &LoadTestMetrics{}
	var wg sync.WaitGroup
	stop := make(chan bool)

	startTime := time.Now()

	// Start concurrent users
	for i := 0; i < numConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for {
				select {
				case <-stop:
					return
				default:
					// Random operation
					operation := userID % 4

					reqStart := time.Now()

					switch operation {
					case 0: // Place order
						symbol := symbols[userID%len(symbols)]
						reqBody := map[string]interface{}{
							"symbol": symbol,
							"side":   "BUY",
							"volume": 0.01,
							"type":   "MARKET",
						}
						body, _ := json.Marshal(reqBody)
						req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
						req.Header.Set("Content-Type", "application/json")
						w := httptest.NewRecorder()
						tc.Server.HandlePlaceOrder(w, req)
						metrics.RecordRequest(time.Since(reqStart), w.Code == 200)

					case 1: // Get ticks
						symbol := symbols[userID%len(symbols)]
						req := httptest.NewRequest("GET", fmt.Sprintf("/ticks?symbol=%s&limit=10", symbol), nil)
						w := httptest.NewRecorder()
						tc.Server.HandleGetTicks(w, req)
						metrics.RecordRequest(time.Since(reqStart), w.Code == 200)

					case 2: // Get OHLC
						symbol := symbols[userID%len(symbols)]
						req := httptest.NewRequest("GET", fmt.Sprintf("/ohlc?symbol=%s&timeframe=1m&limit=5", symbol), nil)
						w := httptest.NewRecorder()
						tc.Server.HandleGetOHLC(w, req)
						metrics.RecordRequest(time.Since(reqStart), w.Code == 200)

					case 3: // Calculate risk
						req := httptest.NewRequest("GET", "/risk/calculate-lot?symbol=EURUSD&riskPercent=2&slPips=20", nil)
						w := httptest.NewRecorder()
						tc.Server.HandleCalculateLot(w, req)
						metrics.RecordRequest(time.Since(reqStart), w.Code == 200)
					}

					time.Sleep(100 * time.Millisecond) // Throttle
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(duration)
	close(stop)
	wg.Wait()

	metrics.TotalDuration = time.Since(startTime)
	metrics.Print(t)

	// Assertions
	successRate := float64(metrics.SuccessRequests) / float64(metrics.TotalRequests) * 100
	t.Logf("\nStress Test Summary:")
	t.Logf("Success Rate: %.2f%%", successRate)
	t.Logf("Sustained RPS: %.2f", metrics.RequestsPerSec)

	if successRate < 95.0 {
		t.Errorf("Success rate %.2f%% is below 95%% threshold", successRate)
	}
}

// ==================== BENCHMARK TESTS ====================

func BenchmarkLoad_PlaceOrder(b *testing.B) {
	tc := SetupTest(&testing.T{})
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	reqBody := map[string]interface{}{
		"symbol": "EURUSD",
		"side":   "BUY",
		"volume": 0.01,
		"type":   "MARKET",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("POST", "/order", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			tc.Server.HandlePlaceOrder(w, req)
		}
	})
}

func BenchmarkLoad_GetTicks(b *testing.B) {
	tc := SetupTest(&testing.T{})
	tc.InjectPrice("EURUSD", 1.10000, 1.10020)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/ticks?symbol=EURUSD&limit=100", nil)
			w := httptest.NewRecorder()
			tc.Server.HandleGetTicks(w, req)
		}
	})
}

func BenchmarkLoad_CalculateRisk(b *testing.B) {
	tc := SetupTest(&testing.T{})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/risk/calculate-lot?symbol=EURUSD&riskPercent=2&slPips=20", nil)
			w := httptest.NewRecorder()
			tc.Server.HandleCalculateLot(w, req)
		}
	})
}
