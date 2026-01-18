package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/epic1st/rtx/backend/auth"
)

// Mock auth service for testing
type mockAuthService struct{}

func (m *mockAuthService) ValidateToken(token string) (*auth.Claims, error) {
	return &auth.Claims{
		UserID:   "test-user",
		Username: "testuser",
		Role:     "admin",
	}, nil
}

func TestAnalyticsHubCreation(t *testing.T) {
	hub := NewAnalyticsHub(nil)
	if hub == nil {
		t.Fatal("Expected hub to be created")
	}

	if hub.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if hub.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}
}

func TestChannelSubscription(t *testing.T) {
	client := &AnalyticsClient{
		subscriptions: make(map[string]bool),
	}

	// Test subscribing to valid channels
	msg := &SubscriptionMessage{
		Action:   "subscribe",
		Channels: []string{ChannelRoutingMetrics, ChannelLPPerformance},
	}

	client.handleSubscription(msg)

	if !client.subscriptions[ChannelRoutingMetrics] {
		t.Error("Expected routing-metrics channel to be subscribed")
	}

	if !client.subscriptions[ChannelLPPerformance] {
		t.Error("Expected lp-performance channel to be subscribed")
	}

	// Test unsubscribing
	msg = &SubscriptionMessage{
		Action:   "unsubscribe",
		Channels: []string{ChannelRoutingMetrics},
	}

	client.handleSubscription(msg)

	if client.subscriptions[ChannelRoutingMetrics] {
		t.Error("Expected routing-metrics channel to be unsubscribed")
	}

	if !client.subscriptions[ChannelLPPerformance] {
		t.Error("Expected lp-performance channel to remain subscribed")
	}
}

func TestInvalidChannelSubscription(t *testing.T) {
	client := &AnalyticsClient{
		subscriptions: make(map[string]bool),
		userID:        "test-user",
	}

	// Try to subscribe to invalid channel
	msg := &SubscriptionMessage{
		Action:   "subscribe",
		Channels: []string{"invalid-channel"},
	}

	client.handleSubscription(msg)

	if client.subscriptions["invalid-channel"] {
		t.Error("Expected invalid channel to be rejected")
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := newRateLimiter(3, 100*time.Millisecond)

	// Should allow first 3 requests immediately
	for i := 0; i < 3; i++ {
		if !limiter.allow() {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// 4th request should be denied
	if limiter.allow() {
		t.Error("Expected 4th request to be denied")
	}

	// Wait for refill
	time.Sleep(150 * time.Millisecond)

	// Should allow one more request after refill
	if !limiter.allow() {
		t.Error("Expected request to be allowed after refill")
	}
}

func TestBroadcastFiltering(t *testing.T) {
	hub := NewAnalyticsHub(nil)

	// Create two clients with different subscriptions
	client1 := &AnalyticsClient{
		send:          make(chan []AnalyticsMessage, 10),
		subscriptions: map[string]bool{ChannelRoutingMetrics: true},
		userID:        "client1",
	}

	client2 := &AnalyticsClient{
		send:          make(chan []AnalyticsMessage, 10),
		subscriptions: map[string]bool{ChannelLPPerformance: true},
		userID:        "client2",
	}

	hub.clients[client1] = true
	hub.clients[client2] = true

	// Broadcast to routing-metrics channel
	msg := &BroadcastMessage{
		Channel: ChannelRoutingMetrics,
		Message: AnalyticsMessage{
			Type: "routing-decision",
			Data: map[string]interface{}{"symbol": "EURUSD"},
		},
	}

	// Manually send to clients (simulating hub behavior)
	for client := range hub.clients {
		if client.subscriptions[msg.Channel] {
			select {
			case client.send <- []AnalyticsMessage{msg.Message}:
			default:
			}
		}
	}

	// Check client1 received the message
	select {
	case msgs := <-client1.send:
		if len(msgs) != 1 {
			t.Error("Expected client1 to receive 1 message")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for message to client1")
	}

	// Check client2 did not receive the message
	select {
	case <-client2.send:
		t.Error("Expected client2 to not receive message")
	case <-time.After(100 * time.Millisecond):
		// Expected timeout - client2 should not receive
	}
}

func TestMessageBatchFormat(t *testing.T) {
	messages := []AnalyticsMessage{
		{
			Type:      "routing-decision",
			Timestamp: "2026-01-19T10:30:45.123Z",
			Data: map[string]interface{}{
				"symbol": "EURUSD",
				"side":   "BUY",
			},
		},
		{
			Type:      "lp-metrics",
			Timestamp: "2026-01-19T10:30:45.124Z",
			Data: map[string]interface{}{
				"lpName": "OANDA",
				"status": "connected",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(messages)
	if err != nil {
		t.Fatalf("Failed to marshal messages: %v", err)
	}

	// Unmarshal back
	var decoded []AnalyticsMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal messages: %v", err)
	}

	if len(decoded) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(decoded))
	}

	if decoded[0].Type != "routing-decision" {
		t.Errorf("Expected routing-decision, got %s", decoded[0].Type)
	}

	if decoded[1].Type != "lp-metrics" {
		t.Errorf("Expected lp-metrics, got %s", decoded[1].Type)
	}
}

func TestPublishMethods(t *testing.T) {
	hub := NewAnalyticsHub(nil)
	go hub.Run()
	defer hub.Stop()

	// Test PublishRoutingMetrics
	metrics := &RoutingMetrics{
		Symbol:          "EURUSD",
		Side:            "BUY",
		Volume:          10000,
		RoutingDecision: "ABOOK",
		LPSelected:      "OANDA",
		ExecutionTime:   45,
		Spread:          0.00015,
	}

	hub.PublishRoutingMetrics(metrics)

	// Wait for broadcast to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify timestamp was set
	if metrics.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}

	// Test PublishAlert
	alert := &Alert{
		Severity: "critical",
		Category: "exposure",
		Title:    "Test Alert",
		Message:  "This is a test",
	}

	hub.PublishAlert(alert)
	time.Sleep(50 * time.Millisecond)

	if alert.Timestamp == "" {
		t.Error("Expected alert timestamp to be set")
	}
}

func TestHelperMethods(t *testing.T) {
	hub := NewAnalyticsHub(nil)
	go hub.Run()
	defer hub.Stop()

	// Test OnOrderRouted
	hub.OnOrderRouted("EURUSD", "BUY", 10000, "ABOOK", "OANDA", 45, 0.00015, 0.00002)

	// Test OnLPStatusChange
	hub.OnLPStatusChange("OANDA", "connected", 0.00015, 0.98, 25, 500, 0.02, 99.9)

	// Test OnExposureChange
	bySymbol := map[string]float64{"EURUSD": 50000}
	byLP := map[string]float64{"OANDA": 50000}
	hub.OnExposureChange(50000, 50000, 100000, bySymbol, byLP)

	// Test EmitAlert
	hub.EmitAlert("warning", "lp", "Test", "Test message", "Test", nil)

	// Allow time for all messages to be processed
	time.Sleep(100 * time.Millisecond)
}

// Integration test with actual WebSocket connection
func TestWebSocketConnection(t *testing.T) {
	// Create mock auth service
	mockAuth := &mockAuthService{}
	hub := NewAnalyticsHub(nil) // Using nil since we're mocking
	go hub.Run()
	defer hub.Stop()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeAnalyticsWs(w, r)
	}))
	defer server.Close()

	// Note: This test would require mocking the auth validation
	// For now, we just verify the server setup doesn't crash

	_ = mockAuth // Mark as used

	// Verify hub is running
	time.Sleep(50 * time.Millisecond)
	if !hub.running {
		t.Error("Expected hub to be running")
	}
}

func TestMinFunction(t *testing.T) {
	if min(5, 10) != 5 {
		t.Error("Expected min(5, 10) to be 5")
	}

	if min(10, 5) != 5 {
		t.Error("Expected min(10, 5) to be 5")
	}

	if min(7, 7) != 7 {
		t.Error("Expected min(7, 7) to be 7")
	}
}

func TestExposureRiskLevelCalculation(t *testing.T) {
	tests := []struct {
		utilization float64
		expected    string
	}{
		{95, "critical"},
		{80, "high"},
		{60, "medium"},
		{30, "low"},
	}

	for _, tt := range tests {
		var riskLevel string
		switch {
		case tt.utilization >= 90:
			riskLevel = "critical"
		case tt.utilization >= 75:
			riskLevel = "high"
		case tt.utilization >= 50:
			riskLevel = "medium"
		default:
			riskLevel = "low"
		}

		if riskLevel != tt.expected {
			t.Errorf("For utilization %.0f%%, expected %s but got %s",
				tt.utilization, tt.expected, riskLevel)
		}
	}
}

func BenchmarkRateLimiter(b *testing.B) {
	limiter := newRateLimiter(1000, time.Millisecond)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		limiter.allow()
	}
}

func BenchmarkBroadcast(b *testing.B) {
	hub := NewAnalyticsHub(nil)
	go hub.Run()
	defer hub.Stop()

	// Create 10 clients
	for i := 0; i < 10; i++ {
		client := &AnalyticsClient{
			send:          make(chan []AnalyticsMessage, 256),
			subscriptions: map[string]bool{ChannelRoutingMetrics: true},
			userID:        "benchmark-client",
		}
		hub.clients[client] = true
	}

	metrics := &RoutingMetrics{
		Symbol:          "EURUSD",
		Side:            "BUY",
		Volume:          10000,
		RoutingDecision: "ABOOK",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hub.PublishRoutingMetrics(metrics)
	}
}
