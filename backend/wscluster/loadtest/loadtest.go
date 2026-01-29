package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// LoadTest configuration
type LoadTestConfig struct {
	TargetURL          string
	TotalConnections   int
	ConnectionsPerSec  int
	TestDuration       time.Duration
	MessagesPerConn    int
	MessageInterval    time.Duration
	PayloadSize        int
	Verbose            bool
}

// LoadTestMetrics tracks test results
type LoadTestMetrics struct {
	TotalConnections      int64
	ActiveConnections     int64
	FailedConnections     int64
	TotalMessagesSent     int64
	TotalMessagesReceived int64
	FailedMessages        int64

	ConnectionLatencies   []time.Duration
	MessageLatencies      []time.Duration

	StartTime             time.Time
	EndTime               time.Time

	mu sync.RWMutex
}

func main() {
	config := &LoadTestConfig{}

	flag.StringVar(&config.TargetURL, "url", "ws://localhost:8080/ws", "WebSocket endpoint URL")
	flag.IntVar(&config.TotalConnections, "connections", 1000, "Total number of connections")
	flag.IntVar(&config.ConnectionsPerSec, "rate", 100, "Connections per second")
	flag.DurationVar(&config.TestDuration, "duration", 60*time.Second, "Test duration")
	flag.IntVar(&config.MessagesPerConn, "messages", 100, "Messages per connection")
	flag.DurationVar(&config.MessageInterval, "interval", 1*time.Second, "Interval between messages")
	flag.IntVar(&config.PayloadSize, "payload", 100, "Message payload size in bytes")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose logging")
	flag.Parse()

	log.Printf("Starting load test with configuration:")
	log.Printf("  Target URL: %s", config.TargetURL)
	log.Printf("  Total connections: %d", config.TotalConnections)
	log.Printf("  Connection rate: %d/sec", config.ConnectionsPerSec)
	log.Printf("  Test duration: %s", config.TestDuration)
	log.Printf("  Messages per connection: %d", config.MessagesPerConn)
	log.Printf("  Message interval: %s", config.MessageInterval)
	log.Printf("  Payload size: %d bytes", config.PayloadSize)

	metrics := &LoadTestMetrics{
		StartTime: time.Now(),
		ConnectionLatencies: make([]time.Duration, 0, config.TotalConnections),
		MessageLatencies: make([]time.Duration, 0, config.TotalConnections*config.MessagesPerConn),
	}

	runLoadTest(config, metrics)

	metrics.EndTime = time.Now()

	printResults(config, metrics)
}

func runLoadTest(config *LoadTestConfig, metrics *LoadTestMetrics) {
	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	defer cancel()

	var wg sync.WaitGroup
	connectionSemaphore := make(chan struct{}, config.ConnectionsPerSec)

	// Connection ticker for rate limiting
	ticker := time.NewTicker(time.Second / time.Duration(config.ConnectionsPerSec))
	defer ticker.Stop()

	// Status reporter
	go reportStatus(ctx, metrics)

	for i := 0; i < config.TotalConnections; i++ {
		select {
		case <-ctx.Done():
			log.Println("Test duration reached, stopping...")
			wg.Wait()
			return
		case <-ticker.C:
			wg.Add(1)
			connectionSemaphore <- struct{}{}

			go func(connID int) {
				defer wg.Done()
				defer func() { <-connectionSemaphore }()

				runConnection(ctx, connID, config, metrics)
			}(i)
		}
	}

	wg.Wait()
}

func runConnection(ctx context.Context, connID int, config *LoadTestConfig, metrics *LoadTestMetrics) {
	// Connect
	connStart := time.Now()

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(config.TargetURL, nil)
	if err != nil {
		atomic.AddInt64(&metrics.FailedConnections, 1)
		if config.Verbose {
			log.Printf("Connection %d failed: %v", connID, err)
		}
		return
	}
	defer conn.Close()

	connLatency := time.Since(connStart)

	atomic.AddInt64(&metrics.TotalConnections, 1)
	atomic.AddInt64(&metrics.ActiveConnections, 1)
	defer atomic.AddInt64(&metrics.ActiveConnections, -1)

	metrics.mu.Lock()
	metrics.ConnectionLatencies = append(metrics.ConnectionLatencies, connLatency)
	metrics.mu.Unlock()

	if config.Verbose {
		log.Printf("Connection %d established (latency: %s)", connID, connLatency)
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(config.TestDuration))

	// Start message receiver
	receiverDone := make(chan struct{})
	go receiveMessages(conn, connID, config, metrics, receiverDone)

	// Send messages
	ticker := time.NewTicker(config.MessageInterval)
	defer ticker.Stop()

	payload := make([]byte, config.PayloadSize)
	for i := 0; i < len(payload); i++ {
		payload[i] = byte('A' + (i % 26))
	}

	for i := 0; i < config.MessagesPerConn; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msgStart := time.Now()

			message := fmt.Sprintf(`{"id":%d,"conn":%d,"msg":%d,"payload":"%s","timestamp":%d}`,
				connID*config.MessagesPerConn+i,
				connID,
				i,
				string(payload),
				time.Now().UnixNano(),
			)

			err := conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				atomic.AddInt64(&metrics.FailedMessages, 1)
				if config.Verbose {
					log.Printf("Connection %d failed to send message %d: %v", connID, i, err)
				}
				continue
			}

			msgLatency := time.Since(msgStart)
			atomic.AddInt64(&metrics.TotalMessagesSent, 1)

			metrics.mu.Lock()
			metrics.MessageLatencies = append(metrics.MessageLatencies, msgLatency)
			metrics.mu.Unlock()
		}
	}

	// Wait a bit for pending messages
	time.Sleep(1 * time.Second)

	close(receiverDone)
}

func receiveMessages(conn *websocket.Conn, connID int, config *LoadTestConfig, metrics *LoadTestMetrics, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}

			atomic.AddInt64(&metrics.TotalMessagesReceived, 1)
		}
	}
}

func reportStatus(ctx context.Context, metrics *LoadTestMetrics) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Printf("Status: %d active connections, %d messages sent, %d messages received",
				atomic.LoadInt64(&metrics.ActiveConnections),
				atomic.LoadInt64(&metrics.TotalMessagesSent),
				atomic.LoadInt64(&metrics.TotalMessagesReceived),
			)
		}
	}
}

func printResults(config *LoadTestConfig, metrics *LoadTestMetrics) {
	duration := metrics.EndTime.Sub(metrics.StartTime)

	fmt.Println("\n========================================")
	fmt.Println("Load Test Results")
	fmt.Println("========================================")

	fmt.Printf("\nTest Configuration:\n")
	fmt.Printf("  Target URL:           %s\n", config.TargetURL)
	fmt.Printf("  Total connections:    %d\n", config.TotalConnections)
	fmt.Printf("  Test duration:        %s\n", duration)

	fmt.Printf("\nConnection Statistics:\n")
	fmt.Printf("  Successful:           %d\n", metrics.TotalConnections)
	fmt.Printf("  Failed:               %d\n", metrics.FailedConnections)
	fmt.Printf("  Success rate:         %.2f%%\n",
		float64(metrics.TotalConnections)/float64(config.TotalConnections)*100)

	if len(metrics.ConnectionLatencies) > 0 {
		connLatencies := calculateLatencyStats(metrics.ConnectionLatencies)
		fmt.Printf("  Connection latency:\n")
		fmt.Printf("    Min:     %s\n", connLatencies.Min)
		fmt.Printf("    Max:     %s\n", connLatencies.Max)
		fmt.Printf("    Average: %s\n", connLatencies.Avg)
		fmt.Printf("    P50:     %s\n", connLatencies.P50)
		fmt.Printf("    P95:     %s\n", connLatencies.P95)
		fmt.Printf("    P99:     %s\n", connLatencies.P99)
	}

	fmt.Printf("\nMessage Statistics:\n")
	fmt.Printf("  Sent:                 %d\n", metrics.TotalMessagesSent)
	fmt.Printf("  Received:             %d\n", metrics.TotalMessagesReceived)
	fmt.Printf("  Failed:               %d\n", metrics.FailedMessages)
	fmt.Printf("  Messages/sec:         %.2f\n",
		float64(metrics.TotalMessagesSent)/duration.Seconds())

	if len(metrics.MessageLatencies) > 0 {
		msgLatencies := calculateLatencyStats(metrics.MessageLatencies)
		fmt.Printf("  Message latency:\n")
		fmt.Printf("    Min:     %s\n", msgLatencies.Min)
		fmt.Printf("    Max:     %s\n", msgLatencies.Max)
		fmt.Printf("    Average: %s\n", msgLatencies.Avg)
		fmt.Printf("    P50:     %s\n", msgLatencies.P50)
		fmt.Printf("    P95:     %s\n", msgLatencies.P95)
		fmt.Printf("    P99:     %s\n", msgLatencies.P99)
	}

	fmt.Printf("\nThroughput:\n")
	fmt.Printf("  Connections/sec:      %.2f\n",
		float64(metrics.TotalConnections)/duration.Seconds())
	fmt.Printf("  Messages/sec:         %.2f\n",
		float64(metrics.TotalMessagesSent)/duration.Seconds())

	bytesPerMessage := config.PayloadSize + 100 // Approximate overhead
	totalBytes := metrics.TotalMessagesSent * int64(bytesPerMessage)
	fmt.Printf("  Throughput:           %.2f MB/s\n",
		float64(totalBytes)/duration.Seconds()/1024/1024)

	fmt.Println("\n========================================")
}

type LatencyStats struct {
	Min time.Duration
	Max time.Duration
	Avg time.Duration
	P50 time.Duration
	P95 time.Duration
	P99 time.Duration
}

func calculateLatencyStats(latencies []time.Duration) *LatencyStats {
	if len(latencies) == 0 {
		return &LatencyStats{}
	}

	// Sort latencies
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Simple bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate stats
	var sum time.Duration
	for _, l := range sorted {
		sum += l
	}

	return &LatencyStats{
		Min: sorted[0],
		Max: sorted[len(sorted)-1],
		Avg: sum / time.Duration(len(sorted)),
		P50: sorted[len(sorted)*50/100],
		P95: sorted[len(sorted)*95/100],
		P99: sorted[len(sorted)*99/100],
	}
}
