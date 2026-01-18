package wscluster

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// EventType represents different cluster event types
type EventType string

const (
	EventTypeMessage       EventType = "message"
	EventTypeBroadcast     EventType = "broadcast"
	EventTypeNodeJoin      EventType = "node_join"
	EventTypeNodeLeave     EventType = "node_leave"
	EventTypeHeartbeat     EventType = "heartbeat"
	EventTypeFailover      EventType = "failover"
	EventTypeScaleUp       EventType = "scale_up"
	EventTypeScaleDown     EventType = "scale_down"
	EventTypeConnectionMigrate EventType = "connection_migrate"
)

// ClusterEvent represents an event in the cluster
type ClusterEvent struct {
	Type      EventType              `json:"type"`
	NodeID    string                 `json:"node_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      interface{}            `json:"data"`
	MessageID string                 `json:"message_id,omitempty"`
}

// BroadcastMessage represents a message to be broadcast to all connections
type BroadcastMessage struct {
	Type        string                 `json:"type"`
	Payload     interface{}            `json:"payload"`
	TargetUser  string                 `json:"target_user,omitempty"`
	TargetRoom  string                 `json:"target_room,omitempty"`
	ExcludeNode string                 `json:"exclude_node,omitempty"`
	Compressed  bool                   `json:"compressed"`
	Binary      bool                   `json:"binary"`
	Priority    int                    `json:"priority"` // 0=low, 1=normal, 2=high, 3=critical
}

// PubSubManager handles Redis Pub/Sub for cluster communication
type PubSubManager struct {
	client       *redis.Client
	nodeID       string

	// Pub/Sub channels
	pubsub       *redis.PubSub
	subscribers  map[string][]chan *ClusterEvent
	subMu        sync.RWMutex

	// Message queue for reliability
	messageQueue chan *ClusterEvent
	queueSize    int

	// Batching
	batchEnabled  bool
	batchInterval time.Duration
	batchSize     int
	pendingBatch  []*BroadcastMessage
	batchMu       sync.Mutex

	// Context
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup

	// Metrics
	publishedMessages   int64
	receivedMessages    int64
	droppedMessages     int64
	batchesSent         int64
	metricsMu           sync.RWMutex
}

// PubSubMetrics tracks pub/sub performance
type PubSubMetrics struct {
	PublishedMessages int64     `json:"published_messages"`
	ReceivedMessages  int64     `json:"received_messages"`
	DroppedMessages   int64     `json:"dropped_messages"`
	BatchesSent       int64     `json:"batches_sent"`
	QueueDepth        int       `json:"queue_depth"`
	Timestamp         time.Time `json:"timestamp"`
}

// NewPubSubManager creates a new PubSub manager
func NewPubSubManager(client *redis.Client, nodeID string) (*PubSubManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	pm := &PubSubManager{
		client:        client,
		nodeID:        nodeID,
		subscribers:   make(map[string][]chan *ClusterEvent),
		messageQueue:  make(chan *ClusterEvent, 10000),
		queueSize:     10000,
		batchEnabled:  true,
		batchInterval: 100 * time.Millisecond,
		batchSize:     100,
		pendingBatch:  make([]*BroadcastMessage, 0, 100),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Subscribe to cluster channel
	pm.pubsub = client.Subscribe(ctx, "ws:cluster:events")

	return pm, nil
}

// Start begins processing pub/sub messages
func (pm *PubSubManager) Start(ctx context.Context) error {
	pm.wg.Add(3)
	go pm.receiveLoop()
	go pm.processQueue()
	go pm.batchProcessor()

	return nil
}

// Stop gracefully shuts down the pub/sub manager
func (pm *PubSubManager) Stop() error {
	pm.cancel()

	// Close pubsub
	if pm.pubsub != nil {
		pm.pubsub.Close()
	}

	// Wait for goroutines
	pm.wg.Wait()

	return nil
}

// PublishEvent publishes an event to the cluster
func (pm *PubSubManager) PublishEvent(event *ClusterEvent) error {
	if event.NodeID == "" {
		event.NodeID = pm.nodeID
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := pm.client.Publish(pm.ctx, "ws:cluster:events", data).Err(); err != nil {
		pm.metricsMu.Lock()
		pm.droppedMessages++
		pm.metricsMu.Unlock()
		return fmt.Errorf("failed to publish event: %w", err)
	}

	pm.metricsMu.Lock()
	pm.publishedMessages++
	pm.metricsMu.Unlock()

	return nil
}

// Broadcast sends a message to all nodes in the cluster
func (pm *PubSubManager) Broadcast(msg *BroadcastMessage) error {
	// If batching is enabled and priority is not critical, queue for batch
	if pm.batchEnabled && msg.Priority < 3 {
		pm.batchMu.Lock()
		pm.pendingBatch = append(pm.pendingBatch, msg)
		shouldFlush := len(pm.pendingBatch) >= pm.batchSize
		pm.batchMu.Unlock()

		if shouldFlush {
			return pm.flushBatch()
		}
		return nil
	}

	// Send immediately for critical messages
	event := &ClusterEvent{
		Type:      EventTypeBroadcast,
		NodeID:    pm.nodeID,
		Timestamp: time.Now(),
		Data:      msg,
	}

	return pm.PublishEvent(event)
}

// BroadcastToUser sends a message to a specific user across all nodes
func (pm *PubSubManager) BroadcastToUser(userID string, msgType string, payload interface{}) error {
	msg := &BroadcastMessage{
		Type:       msgType,
		Payload:    payload,
		TargetUser: userID,
		Priority:   2, // High priority for user-targeted messages
	}

	return pm.Broadcast(msg)
}

// BroadcastToRoom sends a message to all users in a room across all nodes
func (pm *PubSubManager) BroadcastToRoom(roomID string, msgType string, payload interface{}) error {
	msg := &BroadcastMessage{
		Type:       msgType,
		Payload:    payload,
		TargetRoom: roomID,
		Priority:   1, // Normal priority for room messages
	}

	return pm.Broadcast(msg)
}

// Subscribe registers a channel to receive events of a specific type
func (pm *PubSubManager) Subscribe(eventType string) chan *ClusterEvent {
	pm.subMu.Lock()
	defer pm.subMu.Unlock()

	ch := make(chan *ClusterEvent, 100)
	pm.subscribers[eventType] = append(pm.subscribers[eventType], ch)

	return ch
}

// Unsubscribe removes a channel from receiving events
func (pm *PubSubManager) Unsubscribe(eventType string, ch chan *ClusterEvent) {
	pm.subMu.Lock()
	defer pm.subMu.Unlock()

	subs := pm.subscribers[eventType]
	for i, sub := range subs {
		if sub == ch {
			pm.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// receiveLoop listens for incoming Redis pub/sub messages
func (pm *PubSubManager) receiveLoop() {
	defer pm.wg.Done()

	ch := pm.pubsub.Channel()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}

			var event ClusterEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				// Log error
				continue
			}

			// Don't process our own messages
			if event.NodeID == pm.nodeID {
				continue
			}

			pm.metricsMu.Lock()
			pm.receivedMessages++
			pm.metricsMu.Unlock()

			// Queue for processing
			select {
			case pm.messageQueue <- &event:
			default:
				// Queue full, drop message
				pm.metricsMu.Lock()
				pm.droppedMessages++
				pm.metricsMu.Unlock()
			}
		}
	}
}

// processQueue processes queued events
func (pm *PubSubManager) processQueue() {
	defer pm.wg.Done()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case event := <-pm.messageQueue:
			pm.dispatchEvent(event)
		}
	}
}

// dispatchEvent sends event to all subscribers
func (pm *PubSubManager) dispatchEvent(event *ClusterEvent) {
	pm.subMu.RLock()
	defer pm.subMu.RUnlock()

	// Send to wildcard subscribers (subscribed to "cluster")
	for _, ch := range pm.subscribers["cluster"] {
		select {
		case ch <- event:
		default:
			// Subscriber not keeping up, skip
		}
	}

	// Send to type-specific subscribers
	eventType := string(event.Type)
	for _, ch := range pm.subscribers[eventType] {
		select {
		case ch <- event:
		default:
			// Subscriber not keeping up, skip
		}
	}
}

// batchProcessor periodically flushes pending batches
func (pm *PubSubManager) batchProcessor() {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			// Flush any pending messages before shutdown
			pm.flushBatch()
			return
		case <-ticker.C:
			pm.flushBatch()
		}
	}
}

// flushBatch sends all pending batched messages
func (pm *PubSubManager) flushBatch() error {
	pm.batchMu.Lock()
	if len(pm.pendingBatch) == 0 {
		pm.batchMu.Unlock()
		return nil
	}

	batch := pm.pendingBatch
	pm.pendingBatch = make([]*BroadcastMessage, 0, pm.batchSize)
	pm.batchMu.Unlock()

	// Create batch event
	event := &ClusterEvent{
		Type:      EventTypeBroadcast,
		NodeID:    pm.nodeID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"batch":    true,
			"messages": batch,
		},
	}

	if err := pm.PublishEvent(event); err != nil {
		return err
	}

	pm.metricsMu.Lock()
	pm.batchesSent++
	pm.metricsMu.Unlock()

	return nil
}

// GetMetrics returns current pub/sub metrics
func (pm *PubSubManager) GetMetrics() *PubSubMetrics {
	pm.metricsMu.RLock()
	defer pm.metricsMu.RUnlock()

	return &PubSubMetrics{
		PublishedMessages: pm.publishedMessages,
		ReceivedMessages:  pm.receivedMessages,
		DroppedMessages:   pm.droppedMessages,
		BatchesSent:       pm.batchesSent,
		QueueDepth:        len(pm.messageQueue),
		Timestamp:         time.Now(),
	}
}

// SetBatchConfig configures batching behavior
func (pm *PubSubManager) SetBatchConfig(enabled bool, interval time.Duration, size int) {
	pm.batchMu.Lock()
	defer pm.batchMu.Unlock()

	pm.batchEnabled = enabled
	if interval > 0 {
		pm.batchInterval = interval
	}
	if size > 0 {
		pm.batchSize = size
	}
}

// PublishMessage publishes a direct message to a channel
func (pm *PubSubManager) PublishMessage(channel string, data interface{}) error {
	msg, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := pm.client.Publish(pm.ctx, channel, msg).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// SubscribeChannel subscribes to a specific Redis channel
func (pm *PubSubManager) SubscribeChannel(channel string) chan string {
	pubsub := pm.client.Subscribe(pm.ctx, channel)
	ch := make(chan string, 100)

	pm.wg.Add(1)
	go func() {
		defer pm.wg.Done()
		defer pubsub.Close()
		defer close(ch)

		msgCh := pubsub.Channel()
		for {
			select {
			case <-pm.ctx.Done():
				return
			case msg := <-msgCh:
				if msg == nil {
					continue
				}

				select {
				case ch <- msg.Payload:
				default:
					// Channel full, drop message
				}
			}
		}
	}()

	return ch
}
