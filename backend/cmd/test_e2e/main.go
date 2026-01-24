package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type PipelineStats struct {
	Data struct {
		TicksReceived      int64   `json:"ticks_received"`
		TicksProcessed     int64   `json:"ticks_processed"`
		TicksDropped       int64   `json:"ticks_dropped"`
		AvgLatencyMs       float64 `json:"avg_latency_ms"`
		ClientsConnected   int     `json:"clients_connected"`
		OHLCBarsGenerated  int64   `json:"ohlc_bars_generated"`
		QuotesDistributed  int64   `json:"quotes_distributed"`
	} `json:"data"`
}

type MarketTick struct {
	Type      string  `json:"type"`
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Spread    float64 `json:"spread"`
	Timestamp int64   `json:"timestamp"`
	LP        string  `json:"lp"`
}

func main() {
	backendURL := flag.String("backend", "http://localhost:8080", "Backend URL")
	redisAddr := flag.String("redis", "localhost:6379", "Redis address")
	wsURL := flag.String("ws", "ws://localhost:8080", "WebSocket URL")
	token := flag.String("token", "", "Auth token (if empty, will try to get from login)")
	duration := flag.Duration("duration", 30*time.Second, "Test duration")
	flag.Parse()

	ctx := context.Background()

	fmt.Println("========================================")
	fmt.Println("Market Data Pipeline E2E Test")
	fmt.Println("========================================")
	fmt.Printf("Backend: %s\n", *backendURL)
	fmt.Printf("Redis: %s\n", *redisAddr)
	fmt.Printf("WebSocket: %s\n", *wsURL)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Println()

	// 1. Check backend health
	fmt.Println("1. Checking backend health...")
	statsOk, statsData := checkBackendHealth(*backendURL)
	if !statsOk {
		log.Fatal("Backend health check failed")
	}
	fmt.Printf("   ✓ Backend OK (latency: %.2fms, clients: %d)\n", statsData.Data.AvgLatencyMs, statsData.Data.ClientsConnected)

	// 2. Check Redis
	fmt.Println("\n2. Checking Redis...")
	rdb := redis.NewClient(&redis.Options{Addr: *redisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis connection failed: ", err)
	}

	eurusdCount, _ := rdb.LLen(ctx, "market_data:EURUSD").Result()
	fmt.Printf("   ✓ Redis OK (EURUSD ticks: %d)\n", eurusdCount)

	// 3. Get or create auth token
	fmt.Println("\n3. Authentication...")
	authToken := *token
	if authToken == "" {
		var err error
		authToken, err = getAuthToken(*backendURL)
		if err != nil {
			log.Fatalf("Failed to get auth token: %v", err)
		}
	}
	fmt.Println("   ✓ Token obtained")

	// 4. Connect WebSocket
	fmt.Println("\n4. Connecting WebSocket...")
	ws, err := connectWebSocket(*wsURL, authToken)
	if err != nil {
		log.Fatal("WebSocket connection failed: ", err)
	}
	defer ws.Close()
	fmt.Println("   ✓ WebSocket connected")

	// 5. Receive and validate data
	fmt.Println("\n5. Receiving market data...")
	fmt.Printf("   Collecting for %v...\n", duration)

	stats := &TestStats{
		TicksReceived: 0,
		TicksValid:    0,
		TicksInvalid:  0,
		Symbols:       make(map[string]int),
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}

	timeout := time.After(*duration)
	for {
		select {
		case <-timeout:
			goto testDone
		default:
			ws.SetReadDeadline(time.Now().Add(2 * time.Second))

			_, data, err := ws.ReadMessage()
			if err != nil {
				continue
			}

			var tick MarketTick
			if err := json.Unmarshal(data, &tick); err != nil {
				continue
			}

			if tick.Type == "market_tick" {
				stats.TicksReceived++
				stats.Symbols[tick.Symbol]++

				// Validate tick
				if validateTick(&tick) {
					stats.TicksValid++
				} else {
					stats.TicksInvalid++
					fmt.Printf("   Invalid tick: %v\n", tick)
				}

				if stats.TicksReceived%500 == 0 {
					fmt.Printf("   Received %d ticks, %d valid\n", stats.TicksReceived, stats.TicksValid)
				}
			}
		}
	}

testDone:
	stats.EndTime = time.Now()

	// 6. Verify data in Redis
	fmt.Println("\n6. Verifying Redis storage...")
	for symbol := range stats.Symbols {
		count, _ := rdb.LLen(ctx, fmt.Sprintf("market_data:%s", symbol)).Result()
		fmt.Printf("   %s: %d ticks stored\n", symbol, count)
	}

	// 7. Print results
	fmt.Println("\n========================================")
	fmt.Println("Test Results")
	fmt.Println("========================================")
	fmt.Printf("Duration: %v\n", stats.EndTime.Sub(stats.StartTime))
	fmt.Printf("Total ticks received: %d\n", stats.TicksReceived)
	fmt.Printf("Valid ticks: %d\n", stats.TicksValid)
	fmt.Printf("Invalid ticks: %d\n", stats.TicksInvalid)
	fmt.Printf("Unique symbols: %d\n", len(stats.Symbols))
	fmt.Printf("Avg throughput: %.0f ticks/sec\n", float64(stats.TicksReceived)/stats.EndTime.Sub(stats.StartTime).Seconds())
	fmt.Println("\nSymbols received:")
	for symbol, count := range stats.Symbols {
		fmt.Printf("  %s: %d\n", symbol, count)
	}

	// 8. Check final pipeline stats
	fmt.Println("\n========================================")
	fmt.Println("Pipeline Statistics")
	fmt.Println("========================================")
	_, finalStats := checkBackendHealth(*backendURL)
	fmt.Printf("Total ticks processed by pipeline: %d\n", finalStats.Data.TicksProcessed)
	fmt.Printf("Pipeline latency: %.2f ms\n", finalStats.Data.AvgLatencyMs)
	fmt.Printf("Dropped ticks: %d\n", finalStats.Data.TicksDropped)
	fmt.Printf("OHLC bars generated: %d\n", finalStats.Data.OHLCBarsGenerated)

	// 9. Determine overall result
	fmt.Println("\n========================================")
	if stats.TicksReceived > 100 && stats.TicksValid > 95 {
		fmt.Println("✓ TEST PASSED")
		return
	} else {
		fmt.Println("✗ TEST FAILED")
		if stats.TicksReceived < 100 {
			fmt.Println("  Reason: Insufficient data received")
		}
		if stats.TicksInvalid > 5 {
			fmt.Println("  Reason: Too many invalid ticks")
		}
	}
}

func checkBackendHealth(backendURL string) (bool, *PipelineStats) {
	resp, err := http.Get(fmt.Sprintf("%s/api/admin/pipeline-stats", backendURL))
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	var stats PipelineStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return false, nil
	}

	return true, &stats
}

func getAuthToken(backendURL string) (string, error) {
	loginReq := map[string]string{
		"username": "trader",
		"password": "trader",
	}

	data, _ := json.Marshal(loginReq)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(
		fmt.Sprintf("%s/api/login", backendURL),
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

func connectWebSocket(wsURL, token string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	fullURL := fmt.Sprintf("%s/ws?token=%s", wsURL, token)
	ws, _, err := dialer.Dial(fullURL, nil)
	return ws, err
}

func validateTick(tick *MarketTick) bool {
	// Check basic fields
	if tick.Symbol == "" {
		return false
	}

	if tick.Bid <= 0 || tick.Ask <= 0 {
		return false
	}

	// Bid should be < Ask
	if tick.Bid >= tick.Ask {
		return false
	}

	// Spread should be reasonable (< 50 pips)
	spread := (tick.Ask - tick.Bid) / tick.Bid * 10000 // Convert to pips
	if spread > 50 || spread < 0 {
		return false
	}

	// Timestamp should be recent
	if tick.Timestamp == 0 {
		return false
	}
	tickTime := time.UnixMilli(tick.Timestamp)
	if time.Since(tickTime) > 1*time.Minute {
		return false
	}

	return true
}

type TestStats struct {
	TicksReceived int
	TicksValid    int
	TicksInvalid  int
	Symbols       map[string]int
	StartTime     time.Time
	EndTime       time.Time
}
