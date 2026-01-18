package wscluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WSConnection represents a WebSocket connection with metadata
type WSConnection struct {
	ID           string
	Conn         *websocket.Conn
	UserID       string
	Send         chan []byte
	LastActivity time.Time
	Subscriptions map[string]bool
	mu           sync.RWMutex
}

// WSServer integrates cluster with WebSocket handling
type WSServer struct {
	cluster      *Cluster
	upgrader     websocket.Upgrader
	connections  map[string]*WSConnection
	connMu       sync.RWMutex
	affinity     *SessionAffinity

	// Message handlers
	handlers     map[string]func(*WSConnection, interface{})
	handlersMu   sync.RWMutex

	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewWSServer creates a new WebSocket server with cluster support
func NewWSServer(clusterConfig *ClusterConfig) (*WSServer, error) {
	cluster, err := NewCluster(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	server := &WSServer{
		cluster:     cluster,
		connections: make(map[string]*WSConnection),
		affinity:    NewSessionAffinity(clusterConfig.RedisClient, ctx),
		handlers:    make(map[string]func(*WSConnection, interface{})),
		ctx:         ctx,
		cancel:      cancel,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure CORS as needed
			},
		},
	}

	// Set up default handlers
	server.setupDefaultHandlers()

	// Start cluster
	if err := cluster.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cluster: %w", err)
	}

	// Subscribe to cluster broadcasts
	server.wg.Add(1)
	go server.handleClusterBroadcasts()

	return server, nil
}

// setupDefaultHandlers registers default message handlers
func (s *WSServer) setupDefaultHandlers() {
	// Subscribe to symbol/channel
	s.RegisterHandler("subscribe", func(conn *WSConnection, data interface{}) {
		payload, ok := data.(map[string]interface{})
		if !ok {
			return
		}

		channel, ok := payload["channel"].(string)
		if !ok {
			return
		}

		conn.mu.Lock()
		conn.Subscriptions[channel] = true
		conn.mu.Unlock()

		log.Printf("User %s subscribed to %s", conn.UserID, channel)
	})

	// Unsubscribe
	s.RegisterHandler("unsubscribe", func(conn *WSConnection, data interface{}) {
		payload, ok := data.(map[string]interface{})
		if !ok {
			return
		}

		channel, ok := payload["channel"].(string)
		if !ok {
			return
		}

		conn.mu.Lock()
		delete(conn.Subscriptions, channel)
		conn.mu.Unlock()

		log.Printf("User %s unsubscribed from %s", conn.UserID, channel)
	})

	// Ping/Pong
	s.RegisterHandler("ping", func(conn *WSConnection, data interface{}) {
		s.SendToConnection(conn.ID, map[string]interface{}{
			"type": "pong",
			"timestamp": time.Now().Unix(),
		})
	})
}

// RegisterHandler registers a message handler
func (s *WSServer) RegisterHandler(msgType string, handler func(*WSConnection, interface{})) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	s.handlers[msgType] = handler
}

// HandleWebSocket handles WebSocket upgrade and connection
func (s *WSServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from query or JWT
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = uuid.New().String()
	}

	// Check session affinity
	existingMapping, err := s.affinity.GetNodeForUser(userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// If user has existing session on different node, redirect
	if existingMapping != nil && existingMapping.NodeID != s.cluster.localNode.ID {
		// Return redirect response
		w.Header().Set("X-Redirect-To", fmt.Sprintf("ws://%s/ws", existingMapping.NodeAddress))
		http.Error(w, "Redirect to assigned node", http.StatusTemporaryRedirect)
		return
	}

	// Select node using load balancer if no existing session
	if existingMapping == nil {
		node, err := s.cluster.loadBalancer.SelectNode(r.RemoteAddr, userID)
		if err != nil {
			http.Error(w, "No available nodes", http.StatusServiceUnavailable)
			return
		}

		// If selected node is not local, redirect
		if node.ID != s.cluster.localNode.ID {
			w.Header().Set("X-Redirect-To", fmt.Sprintf("ws://%s:%d/ws", node.Address, node.Port))
			http.Error(w, "Redirect to selected node", http.StatusTemporaryRedirect)
			return
		}
	}

	// Upgrade connection
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Create connection object
	connID := uuid.New().String()
	wsConn := &WSConnection{
		ID:           connID,
		Conn:         conn,
		UserID:       userID,
		Send:         make(chan []byte, 256),
		LastActivity: time.Now(),
		Subscriptions: make(map[string]bool),
	}

	// Register connection
	s.connMu.Lock()
	s.connections[connID] = wsConn
	s.connMu.Unlock()

	// Assign session
	s.affinity.AssignSession(
		userID,
		connID,
		s.cluster.localNode.ID,
		fmt.Sprintf("%s:%d", s.cluster.localNode.Address, s.cluster.localNode.Port),
		map[string]interface{}{
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		},
	)

	// Update cluster stats
	s.updateClusterStats(1)

	log.Printf("New connection: %s (user: %s)", connID, userID)

	// Handle connection
	s.wg.Add(2)
	go s.readPump(wsConn)
	go s.writePump(wsConn)
}

// readPump reads messages from WebSocket
func (s *WSServer) readPump(conn *WSConnection) {
	defer s.wg.Done()
	defer s.closeConnection(conn)

	conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.Conn.SetPongHandler(func(string) error {
		conn.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.LastActivity = time.Now()
		return nil
	})

	for {
		_, message, err := conn.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		conn.LastActivity = time.Now()

		// Parse message
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		// Get message type
		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		// Handle message
		s.handlersMu.RLock()
		handler, exists := s.handlers[msgType]
		s.handlersMu.RUnlock()

		if exists {
			handler(conn, msg["data"])
		}
	}
}

// writePump writes messages to WebSocket
func (s *WSServer) writePump(conn *WSConnection) {
	defer s.wg.Done()

	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-conn.Send:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				conn.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			conn.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// closeConnection closes and cleans up a connection
func (s *WSServer) closeConnection(conn *WSConnection) {
	s.connMu.Lock()
	delete(s.connections, conn.ID)
	s.connMu.Unlock()

	conn.Conn.Close()
	close(conn.Send)

	// Remove session affinity
	s.affinity.RemoveSession(conn.UserID, conn.ID)

	// Update cluster stats
	s.updateClusterStats(-1)

	log.Printf("Connection closed: %s (user: %s)", conn.ID, conn.UserID)
}

// handleClusterBroadcasts handles broadcasts from other cluster nodes
func (s *WSServer) handleClusterBroadcasts() {
	defer s.wg.Done()

	eventChan := s.cluster.pubsub.Subscribe("broadcast")

	for {
		select {
		case <-s.ctx.Done():
			return
		case event := <-eventChan:
			if event.Type != EventTypeBroadcast {
				continue
			}

			msg, ok := event.Data.(*BroadcastMessage)
			if !ok {
				continue
			}

			// Convert payload to JSON
			data, err := json.Marshal(map[string]interface{}{
				"type":    msg.Type,
				"payload": msg.Payload,
			})
			if err != nil {
				continue
			}

			// Broadcast to local connections
			if msg.TargetUser != "" {
				// Send to specific user
				s.SendToUser(msg.TargetUser, data)
			} else if msg.TargetRoom != "" {
				// Send to room (all subscribed to channel)
				s.BroadcastToChannel(msg.TargetRoom, data)
			} else {
				// Send to all local connections
				s.BroadcastToAll(data)
			}
		}
	}
}

// BroadcastToAll sends message to all local connections
func (s *WSServer) BroadcastToAll(message []byte) {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	for _, conn := range s.connections {
		select {
		case conn.Send <- message:
		default:
			// Channel full, skip
		}
	}
}

// BroadcastToChannel sends message to all connections subscribed to a channel
func (s *WSServer) BroadcastToChannel(channel string, message []byte) {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	for _, conn := range s.connections {
		conn.mu.RLock()
		subscribed := conn.Subscriptions[channel]
		conn.mu.RUnlock()

		if subscribed {
			select {
			case conn.Send <- message:
			default:
				// Channel full, skip
			}
		}
	}
}

// SendToUser sends message to a specific user
func (s *WSServer) SendToUser(userID string, message []byte) {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	for _, conn := range s.connections {
		if conn.UserID == userID {
			select {
			case conn.Send <- message:
			default:
				// Channel full, skip
			}
		}
	}
}

// SendToConnection sends message to a specific connection
func (s *WSServer) SendToConnection(connID string, data interface{}) error {
	s.connMu.RLock()
	conn, exists := s.connections[connID]
	s.connMu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	select {
	case conn.Send <- message:
		return nil
	default:
		return fmt.Errorf("send channel full")
	}
}

// BroadcastToCluster broadcasts message to all nodes in cluster
func (s *WSServer) BroadcastToCluster(msgType string, payload interface{}) error {
	return s.cluster.pubsub.Broadcast(&BroadcastMessage{
		Type:     msgType,
		Payload:  payload,
		Priority: 1,
	})
}

// updateClusterStats updates cluster connection stats
func (s *WSServer) updateClusterStats(delta int64) {
	s.connMu.RLock()
	activeConns := int64(len(s.connections))
	s.connMu.RUnlock()

	s.cluster.UpdateLocalNode(activeConns, 0, 0)
}

// GetMetrics returns server metrics
func (s *WSServer) GetMetrics() map[string]interface{} {
	s.connMu.RLock()
	connCount := len(s.connections)
	s.connMu.RUnlock()

	clusterMetrics := s.cluster.GetMetrics()

	return map[string]interface{}{
		"local_connections":   connCount,
		"cluster_connections": clusterMetrics.TotalConnections,
		"cluster_nodes":       clusterMetrics.TotalNodes,
		"healthy_nodes":       clusterMetrics.HealthyNodes,
	}
}

// Shutdown gracefully shuts down the server
func (s *WSServer) Shutdown() error {
	s.cancel()

	// Close all connections
	s.connMu.Lock()
	for _, conn := range s.connections {
		conn.Conn.Close()
	}
	s.connMu.Unlock()

	// Stop cluster
	s.cluster.Stop()

	// Wait for goroutines
	s.wg.Wait()

	return nil
}
