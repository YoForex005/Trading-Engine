package wscluster_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/wscluster"
	"github.com/redis/go-redis/v9"
)

func TestClusterSetup(t *testing.T) {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Configure cluster
	config := &wscluster.ClusterConfig{
		RedisClient:       redisClient,
		NodeAddress:       "127.0.0.1",
		NodePort:          8080,
		MaxConnections:    1000,
		HeartbeatInterval: 1 * time.Second,
		NodeTimeout:       3 * time.Second,
		MinNodes:          1,
		MaxNodes:          5,
	}

	// Create cluster
	cluster, err := wscluster.NewCluster(config)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	// Start cluster
	if err := cluster.Start(); err != nil {
		t.Fatalf("Failed to start cluster: %v", err)
	}
	defer cluster.Stop()

	// Wait for heartbeat
	time.Sleep(2 * time.Second)

	// Get metrics
	metrics := cluster.GetMetrics()
	if metrics.TotalNodes < 1 {
		t.Errorf("Expected at least 1 node, got %d", metrics.TotalNodes)
	}

	t.Logf("Cluster metrics: %+v", metrics)
}

func TestLoadBalancing(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	config := &wscluster.ClusterConfig{
		RedisClient:    redisClient,
		NodeAddress:    "127.0.0.1",
		NodePort:       8080,
		MaxConnections: 1000,
	}

	cluster, err := wscluster.NewCluster(config)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	if err := cluster.Start(); err != nil {
		t.Fatalf("Failed to start cluster: %v", err)
	}
	defer cluster.Stop()

	lb := cluster.GetLoadBalancer()

	// Test different algorithms
	algorithms := []wscluster.LoadBalancingAlgorithm{
		wscluster.AlgorithmRoundRobin,
		wscluster.AlgorithmLeastConnections,
		wscluster.AlgorithmRandom,
		wscluster.AlgorithmAdaptive,
	}

	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			lb.SetAlgorithm(algo)

			node, err := lb.SelectNode("192.168.1.1", "user1")
			if err != nil {
				t.Logf("No nodes available for %s: %v", algo, err)
				return
			}

			t.Logf("Algorithm %s selected node: %s", algo, node.ID)
		})
	}
}

func TestPubSub(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	config := &wscluster.ClusterConfig{
		RedisClient: redisClient,
		NodeAddress: "127.0.0.1",
		NodePort:    8080,
	}

	cluster, err := wscluster.NewCluster(config)
	if err != nil {
		t.Fatalf("Failed to create cluster: %v", err)
	}

	if err := cluster.Start(); err != nil {
		t.Fatalf("Failed to start cluster: %v", err)
	}
	defer cluster.Stop()

	pubsub := cluster.GetPubSub()

	// Subscribe to broadcasts
	eventChan := pubsub.Subscribe("broadcast")
	defer pubsub.Unsubscribe("broadcast", eventChan)

	// Broadcast message
	msg := &wscluster.BroadcastMessage{
		Type:     "test",
		Payload:  "Hello, cluster!",
		Priority: 1,
	}

	if err := pubsub.Broadcast(msg); err != nil {
		t.Fatalf("Failed to broadcast: %v", err)
	}

	// Wait for message
	select {
	case event := <-eventChan:
		t.Logf("Received event: %+v", event)
	case <-time.After(2 * time.Second):
		t.Log("No event received (expected for single node)")
	}

	// Get metrics
	metrics := pubsub.GetMetrics()
	t.Logf("PubSub metrics: %+v", metrics)
}

func TestSessionAffinity(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	affinity := wscluster.NewSessionAffinity(redisClient, ctx)

	// Assign session
	err := affinity.AssignSession(
		"user123",
		"conn456",
		"node-1",
		"192.168.1.100:8080",
		map[string]interface{}{"client_ip": "1.2.3.4"},
	)
	if err != nil {
		t.Fatalf("Failed to assign session: %v", err)
	}

	// Retrieve by user
	mapping, err := affinity.GetNodeForUser("user123")
	if err != nil {
		t.Fatalf("Failed to get node: %v", err)
	}

	if mapping == nil {
		t.Fatal("Mapping not found")
	}

	if mapping.NodeID != "node-1" {
		t.Errorf("Expected node-1, got %s", mapping.NodeID)
	}

	// Retrieve by connection
	mapping2, err := affinity.GetNodeForConnection("conn456")
	if err != nil {
		t.Fatalf("Failed to get node by connection: %v", err)
	}

	if mapping2.NodeID != "node-1" {
		t.Errorf("Expected node-1, got %s", mapping2.NodeID)
	}

	// Remove session
	if err := affinity.RemoveSession("user123", "conn456"); err != nil {
		t.Fatalf("Failed to remove session: %v", err)
	}

	// Verify removed
	mapping3, err := affinity.GetNodeForUser("user123")
	if err != nil {
		t.Fatalf("Failed to check removed session: %v", err)
	}

	if mapping3 != nil {
		t.Error("Session should be removed")
	}

	t.Log("Session affinity test passed")
}

func ExampleCluster() {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	// Configure cluster
	config := &wscluster.ClusterConfig{
		RedisClient:       redisClient,
		NodeAddress:       "192.168.1.100",
		NodePort:          8080,
		MaxConnections:    10000,
		Region:            "us-east-1",
		HeartbeatInterval: 5 * time.Second,
		ScaleUpThreshold:  0.8,
		EnableBatching:    true,
	}

	// Create and start cluster
	cluster, err := wscluster.NewCluster(config)
	if err != nil {
		log.Fatal(err)
	}

	if err := cluster.Start(); err != nil {
		log.Fatal(err)
	}
	defer cluster.Stop()

	// Configure load balancer
	lb := cluster.GetLoadBalancer()
	lb.SetAlgorithm(wscluster.AlgorithmAdaptive)

	// Select node for new connection
	node, err := lb.SelectNode("192.168.1.1", "user123")
	if err != nil {
		log.Printf("Failed to select node: %v", err)
		return
	}

	fmt.Printf("Selected node: %s at %s:%d\n", node.ID, node.Address, node.Port)

	// Broadcast message
	pubsub := cluster.GetPubSub()
	pubsub.Broadcast(&wscluster.BroadcastMessage{
		Type:     "price_update",
		Payload:  map[string]interface{}{"symbol": "BTCUSD", "price": 45000},
		Priority: 2,
	})
}

func ExampleLoadBalancer_SelectNode() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	config := &wscluster.ClusterConfig{
		RedisClient: redisClient,
		NodeAddress: "127.0.0.1",
		NodePort:    8080,
	}

	cluster, _ := wscluster.NewCluster(config)
	cluster.Start()
	defer cluster.Stop()

	lb := cluster.GetLoadBalancer()

	// Use adaptive algorithm
	lb.SetAlgorithm(wscluster.AlgorithmAdaptive)
	lb.SetThresholds(0.85, 0.85, 100.0)

	// Select node
	node, err := lb.SelectNode("192.168.1.100", "user123")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Node: %s (load: %d/%d)\n", node.ID, node.Connections, node.MaxConnections)
}

func ExamplePubSubManager_Broadcast() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	config := &wscluster.ClusterConfig{
		RedisClient: redisClient,
	}

	cluster, _ := wscluster.NewCluster(config)
	cluster.Start()
	defer cluster.Stop()

	pubsub := cluster.GetPubSub()

	// Broadcast to all nodes
	pubsub.Broadcast(&wscluster.BroadcastMessage{
		Type:     "market_data",
		Payload:  map[string]interface{}{"symbol": "ETHUSD", "price": 3000},
		Priority: 1,
	})

	// Broadcast to specific user
	pubsub.BroadcastToUser("user123", "notification", map[string]string{
		"message": "Order filled",
	})

	// Broadcast to room
	pubsub.BroadcastToRoom("BTCUSD", "tick", map[string]float64{
		"bid": 44999.5,
		"ask": 45000.5,
	})
}
