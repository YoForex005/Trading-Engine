# Quick Start Guide

Get your WebSocket cluster running in 5 minutes.

## Prerequisites

- Go 1.24+
- Redis 7.0+
- 10k+ file descriptor limit

```bash
# Check Go version
go version

# Check Redis
redis-cli ping

# Increase file descriptors
ulimit -n 200000
```

## Installation

```bash
cd backend
go get github.com/redis/go-redis/v9
go get github.com/gorilla/websocket
go get github.com/google/uuid
```

## Step 1: Start Redis

```bash
# Local Redis
redis-server

# Or Docker
docker run -d -p 6379:6379 redis:7-alpine
```

## Step 2: Create WebSocket Server

Create `backend/cmd/wsserver/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/epic1st/rtx/backend/wscluster"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Configure Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis not available:", err)
	}

	// Configure cluster
	config := &wscluster.ClusterConfig{
		RedisClient:       redisClient,
		NodeAddress:       "127.0.0.1",
		NodePort:          8080,
		MaxConnections:    10000,
		HeartbeatInterval: 5 * time.Second,
		NodeTimeout:       15 * time.Second,
		ScaleUpThreshold:  0.8,
		ScaleDownThreshold: 0.3,
		MinNodes:          2,
		MaxNodes:          20,
		EnableCompression: true,
		EnableBatching:    true,
		BatchInterval:     100 * time.Millisecond,
		MaxBatchSize:      100,
	}

	// Create WebSocket server with clustering
	server, err := wscluster.NewWSServer(config)
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}
	defer server.Shutdown()

	// Configure load balancer
	lb := server.cluster.GetLoadBalancer()
	lb.SetAlgorithm(wscluster.AlgorithmAdaptive)
	lb.SetThresholds(0.85, 0.85, 100.0)

	// Configure failover
	failover := server.cluster.GetFailover()
	failover.SetStrategy(wscluster.FailoverGraceful)
	failover.SetGracePeriod(5 * time.Second)

	// Register custom message handlers
	server.RegisterHandler("order", handleOrder)
	server.RegisterHandler("subscribe_ticker", handleSubscribeTicker)

	// HTTP routes
	http.HandleFunc("/ws", server.HandleWebSocket)
	http.HandleFunc("/health", handleHealth(server))
	http.HandleFunc("/metrics", handleMetrics(server))

	// Start metrics reporter
	go reportMetrics(server)

	// Start server
	log.Println("WebSocket cluster node starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleOrder(conn *wscluster.WSConnection, data interface{}) {
	log.Printf("Order from user %s: %+v", conn.UserID, data)

	// Process order...

	// Send response
	// (Implementation in integration_example.go)
}

func handleSubscribeTicker(conn *wscluster.WSConnection, data interface{}) {
	payload, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	symbol, ok := payload["symbol"].(string)
	if !ok {
		return
	}

	conn.mu.Lock()
	conn.Subscriptions["ticker:"+symbol] = true
	conn.mu.Unlock()

	log.Printf("User %s subscribed to ticker: %s", conn.UserID, symbol)
}

func handleHealth(server *wscluster.WSServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := server.GetMetrics()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Simple health check
		if metrics["healthy_nodes"].(int) > 0 {
			w.Write([]byte(`{"status":"healthy"}`))
		} else {
			w.Write([]byte(`{"status":"degraded"}`))
		}
	}
}

func handleMetrics(server *wscluster.WSServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := server.cluster.GetMetrics()

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		// Prometheus format
		w.Write([]byte(fmt.Sprintf(`
ws_cluster_total_connections %d
ws_cluster_total_nodes %d
ws_cluster_healthy_nodes %d
ws_cluster_messages_per_sec %d
ws_cluster_bytes_per_sec %d
`,
			metrics.TotalConnections,
			metrics.TotalNodes,
			metrics.HealthyNodes,
			metrics.TotalMessagesPerSec,
			metrics.TotalBytesPerSec,
		)))
	}
}

func reportMetrics(server *wscluster.WSServer) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := server.GetMetrics()
		log.Printf("Metrics: %+v", metrics)

		clusterMetrics := server.cluster.GetMetrics()
		log.Printf("Cluster: %d connections across %d nodes",
			clusterMetrics.TotalConnections,
			clusterMetrics.TotalNodes,
		)
	}
}
```

## Step 3: Build and Run

```bash
# Build
cd backend/cmd/wsserver
go build -o wsserver

# Run first node
./wsserver

# In another terminal, run second node
PORT=8081 ./wsserver

# Run third node
PORT=8082 ./wsserver
```

## Step 4: Test Connection

```bash
# Install wscat
npm install -g wscat

# Connect to node
wscat -c ws://localhost:8080/ws?user_id=user123

# Send ping
{"type":"ping"}

# Subscribe to ticker
{"type":"subscribe_ticker","data":{"symbol":"BTCUSD"}}

# Send order
{"type":"order","data":{"symbol":"BTCUSD","side":"buy","quantity":1}}
```

## Step 5: Monitor Cluster

```bash
# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics

# Redis cluster data
redis-cli KEYS "ws:cluster:node:*"
redis-cli GET "ws:cluster:metrics"
```

## Step 6: Load Test

```bash
# Build load tester
cd backend/wscluster/loadtest
go build -o loadtest

# Test 1000 connections
./loadtest -url ws://localhost:8080/ws -connections 1000

# Test 10k connections
./loadtest -url ws://localhost:8080/ws -connections 10000 -rate 500
```

## Step 7: Test Failover

```bash
# Kill node 1
pkill -f "wsserver.*8080"

# Watch cluster automatically migrate connections
redis-cli KEYS "ws:cluster:node:*"

# Clients will reconnect to other nodes automatically
```

## Broadcasting Messages

From your application code:

```go
// Broadcast to all connections
server.BroadcastToCluster("price_update", map[string]interface{}{
	"symbol": "BTCUSD",
	"price":  45000,
	"time":   time.Now().Unix(),
})

// Send to specific user (across all nodes)
pubsub := server.cluster.GetPubSub()
pubsub.BroadcastToUser("user123", "notification", map[string]string{
	"message": "Your order has been filled",
})

// Send to room/channel
pubsub.BroadcastToRoom("BTCUSD", "tick", tickData)
```

## Production Deployment

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data

  ws-node-1:
    build: .
    ports:
      - "8080:8080"
    environment:
      - REDIS_URL=redis:6379
      - NODE_ID=node-1
      - NODE_PORT=8080
    depends_on:
      - redis

  ws-node-2:
    build: .
    ports:
      - "8081:8080"
    environment:
      - REDIS_URL=redis:6379
      - NODE_ID=node-2
      - NODE_PORT=8081
    depends_on:
      - redis

  ws-node-3:
    build: .
    ports:
      - "8082:8080"
    environment:
      - REDIS_URL=redis:6379
      - NODE_ID=node-3
      - NODE_PORT=8082
    depends_on:
      - redis

  haproxy:
    image: haproxy:2.8-alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
    depends_on:
      - ws-node-1
      - ws-node-2
      - ws-node-3

volumes:
  redis-data:
```

Create `haproxy.cfg`:

```
global
    maxconn 100000

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms

frontend websocket_frontend
    bind *:443 ssl crt /etc/ssl/certs/server.pem
    default_backend websocket_backend

backend websocket_backend
    balance source  # Sticky sessions via IP hash
    option httpchk GET /health
    server ws1 ws-node-1:8080 check
    server ws2 ws-node-2:8081 check
    server ws3 ws-node-3:8082 check
```

Start cluster:

```bash
docker-compose up -d
docker-compose ps
docker-compose logs -f
```

## Monitoring with Prometheus

Create `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'websocket-cluster'
    static_configs:
      - targets:
        - 'localhost:8080'
        - 'localhost:8081'
        - 'localhost:8082'
    scrape_interval: 10s
```

Start Prometheus:

```bash
docker run -d \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

## Troubleshooting

### "Too many open files"
```bash
ulimit -n 200000
# Make permanent in /etc/security/limits.conf
```

### Redis connection refused
```bash
redis-cli ping
# Check Redis is running
# Check firewall rules
```

### High CPU usage
```bash
# Reduce connection rate
# Enable batching
# Add more nodes
```

### Connections not balanced
```bash
# Check load balancer algorithm
# Verify sticky sessions
# Check node health
```

## Next Steps

1. ✅ Test with production traffic
2. ✅ Set up monitoring dashboards
3. ✅ Configure auto-scaling
4. ✅ Enable SSL/TLS
5. ✅ Implement rate limiting
6. ✅ Add authentication
7. ✅ Deploy to production

## Resources

- **Architecture**: See `ARCHITECTURE.md`
- **API Reference**: See `README.md`
- **Load Testing**: See `loadtest/README.md`
- **Examples**: See `example_test.go`

## Support

For issues and questions:
- Check documentation
- Review example code
- Test with load tester
- Monitor metrics

---

**You're now running a production-ready WebSocket cluster! 🚀**
