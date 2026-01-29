/**
 * Enhanced WebSocket Service with Advanced Features
 * - Connection status monitoring
 * - Automatic reconnection with exponential backoff
 * - Message queue for offline mode
 * - Heartbeat/ping-pong mechanism
 * - Multi-symbol subscription management
 * - Performance metrics tracking
 */

export type SubscriptionCallback = (data: unknown) => void;
export type ConnectionStateCallback = (state: ConnectionState) => void;
export type MetricsCallback = (metrics: ConnectionMetrics) => void;

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error' | 'offline';

export interface WebSocketMessage {
  type: string;
  symbol?: string;
  channel?: string;
  data?: unknown;
  timestamp?: number;
  [key: string]: unknown;
}

export interface ConnectionMetrics {
  messagesSent: number;
  messagesReceived: number;
  reconnectAttempts: number;
  lastLatency: number;
  averageLatency: number;
  uptime: number;
  lastConnectedAt: number | null;
  lastDisconnectedAt: number | null;
}

interface QueuedMessage {
  message: WebSocketMessage;
  timestamp: number;
  retries: number;
}

export class EnhancedWebSocketService {
  private ws: WebSocket | null = null;
  private url: string;

  // Connection management
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private baseReconnectDelay = 1000;
  private maxReconnectDelay = 30000;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

  // Heartbeat/ping-pong
  private pingInterval: ReturnType<typeof setInterval> | null = null;
  private pongTimeout: ReturnType<typeof setTimeout> | null = null;
  private pingIntervalMs = 30000; // 30 seconds
  private pongTimeoutMs = 5000; // 5 seconds
  private lastPongTime = 0;

  // Subscription management
  private subscribers: Map<string, Set<SubscriptionCallback>> = new Map();
  private stateListeners: Set<ConnectionStateCallback> = new Set();
  private metricsListeners: Set<MetricsCallback> = new Set();
  private connectionState: ConnectionState = 'disconnected';

  // Message buffering and throttling
  private tickBuffer: Map<string, WebSocketMessage> = new Map();
  private messageQueue: QueuedMessage[] = [];
  private maxQueueSize = 1000;
  private flushInterval: ReturnType<typeof setInterval> | null = null;
  private updateThrottleMs = 50; // 50ms = 20 updates/second per symbol

  // Metrics tracking
  private metrics: ConnectionMetrics = {
    messagesSent: 0,
    messagesReceived: 0,
    reconnectAttempts: 0,
    lastLatency: 0,
    averageLatency: 0,
    uptime: 0,
    lastConnectedAt: null,
    lastDisconnectedAt: null,
  };
  private latencyMeasurements: number[] = [];
  private maxLatencyMeasurements = 100;
  private connectStartTime = 0;

  constructor(url: string) {
    this.url = url;

    // Listen to online/offline events
    if (typeof window !== 'undefined') {
      window.addEventListener('online', this.handleOnline);
      window.addEventListener('offline', this.handleOffline);
    }
  }

  /**
   * Connect to WebSocket server
   */
  public connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[WS] Already connected');
      return;
    }

    this.updateState('connecting');
    this.connectStartTime = Date.now();

    // Get auth token and add to URL
    const token = localStorage.getItem('rtx_token');
    const urlWithAuth = this.addTokenToUrl(this.url, token);

    console.log(`[WS] Connecting to ${this.url}...`);

    try {
      this.ws = new WebSocket(urlWithAuth);

      this.ws.onopen = () => {
        console.log('[WS] Connected successfully');
        const latency = Date.now() - this.connectStartTime;
        this.recordLatency(latency);

        this.reconnectAttempts = 0;
        this.metrics.lastConnectedAt = Date.now();
        this.updateState('connected');
        this.startPingInterval();
        this.startFlushInterval();
        this.processMessageQueue();
        this.resubscribeChannels();
        this.notifyMetricsListeners();
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as WebSocketMessage;
          this.handleMessage(data);
          this.metrics.messagesReceived++;
        } catch (error) {
          console.error('[WS] Failed to parse message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('[WS] Error:', error);
        this.updateState('error');
      };

      this.ws.onclose = (event) => {
        console.log(`[WS] Connection closed (code: ${event.code}, reason: ${event.reason})`);
        this.metrics.lastDisconnectedAt = Date.now();
        this.updateState('disconnected');
        this.stopPingInterval();
        this.stopFlushInterval();
        this.notifyMetricsListeners();

        // Check if close was due to authentication failure
        if (event.code === 1008 || event.code === 401) {
          console.error('[WS] Authentication failed - clearing token');
          localStorage.removeItem('rtx_token');
          window.dispatchEvent(new CustomEvent('ws-auth-failed'));
          return;
        }

        this.attemptReconnect();
      };
    } catch (error) {
      console.error('[WS] Connection failed:', error);
      this.updateState('error');
      this.attemptReconnect();
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  public disconnect(): void {
    console.log('[WS] Disconnecting...');

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    this.stopPingInterval();
    this.stopFlushInterval();

    if (this.ws) {
      this.ws.close(1000, 'Client initiated disconnect');
      this.ws = null;
    }

    this.reconnectAttempts = 0;
    this.updateState('disconnected');
  }

  /**
   * Subscribe to a channel
   */
  public subscribe(channel: string, callback: SubscriptionCallback): () => void {
    if (!this.subscribers.has(channel)) {
      this.subscribers.set(channel, new Set());
    }

    this.subscribers.get(channel)!.add(callback);

    // Send subscription message if connected
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.sendMessage({
        type: 'subscribe',
        channel,
      });
    }

    // Return unsubscribe function
    return () => {
      const callbacks = this.subscribers.get(channel);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.subscribers.delete(channel);
          this.sendMessage({
            type: 'unsubscribe',
            channel,
          });
        }
      }
    };
  }

  /**
   * Listen to connection state changes
   */
  public onStateChange(callback: ConnectionStateCallback): () => void {
    this.stateListeners.add(callback);
    callback(this.connectionState);

    return () => {
      this.stateListeners.delete(callback);
    };
  }

  /**
   * Listen to metrics updates
   */
  public onMetricsChange(callback: MetricsCallback): () => void {
    this.metricsListeners.add(callback);
    callback(this.metrics);

    return () => {
      this.metricsListeners.delete(callback);
    };
  }

  /**
   * Get current connection state
   */
  public getState(): ConnectionState {
    return this.connectionState;
  }

  /**
   * Get current metrics
   */
  public getMetrics(): ConnectionMetrics {
    return { ...this.metrics };
  }

  /**
   * Send message to WebSocket server (with offline queueing)
   */
  public sendMessage(message: WebSocketMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      try {
        this.ws.send(JSON.stringify(message));
        this.metrics.messagesSent++;
      } catch (error) {
        console.error('[WS] Failed to send message:', error);
        this.queueMessage(message);
      }
    } else {
      this.queueMessage(message);
    }
  }

  /**
   * Queue message for later delivery
   */
  private queueMessage(message: WebSocketMessage): void {
    if (this.messageQueue.length >= this.maxQueueSize) {
      console.warn('[WS] Message queue full, dropping oldest message');
      this.messageQueue.shift();
    }

    this.messageQueue.push({
      message,
      timestamp: Date.now(),
      retries: 0,
    });
  }

  /**
   * Process queued messages when connection is restored
   */
  private processMessageQueue(): void {
    if (this.messageQueue.length === 0) return;

    console.log(`[WS] Processing ${this.messageQueue.length} queued messages`);

    const now = Date.now();
    const maxAge = 60000; // 1 minute

    this.messageQueue = this.messageQueue.filter((queued) => {
      // Drop messages older than maxAge
      if (now - queued.timestamp > maxAge) {
        console.warn('[WS] Dropping stale queued message');
        return false;
      }

      // Send message
      this.sendMessage(queued.message);
      return false; // Remove from queue
    });
  }

  /**
   * Handle incoming WebSocket messages
   */
  private handleMessage(data: WebSocketMessage): void {
    const { type } = data;

    // Handle pong responses
    if (type === 'pong') {
      this.handlePong();
      return;
    }

    // Buffer tick updates for throttling
    if (type === 'tick' && data.symbol) {
      this.tickBuffer.set(data.symbol, data);
      return;
    }

    // Broadcast to subscribers
    this.broadcastToSubscribers(data);
  }

  /**
   * Broadcast message to relevant subscribers
   */
  private broadcastToSubscribers(data: WebSocketMessage): void {
    const channel = data.channel || data.symbol || data.type;

    // Broadcast to channel-specific subscribers
    const callbacks = this.subscribers.get(channel);
    if (callbacks) {
      callbacks.forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error('[WS] Callback error:', error);
        }
      });
    }

    // Broadcast to wildcard subscribers
    const wildcardCallbacks = this.subscribers.get('*');
    if (wildcardCallbacks) {
      wildcardCallbacks.forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error('[WS] Wildcard callback error:', error);
        }
      });
    }
  }

  /**
   * Flush buffered tick updates
   */
  private flushTickBuffer(): void {
    if (this.tickBuffer.size === 0) return;

    this.tickBuffer.forEach((data) => {
      this.broadcastToSubscribers(data);
    });

    this.tickBuffer.clear();
  }

  /**
   * Start flush interval for throttled updates
   */
  private startFlushInterval(): void {
    if (this.flushInterval) return;

    this.flushInterval = setInterval(() => {
      this.flushTickBuffer();
    }, this.updateThrottleMs);
  }

  /**
   * Stop flush interval
   */
  private stopFlushInterval(): void {
    if (this.flushInterval) {
      clearInterval(this.flushInterval);
      this.flushInterval = null;
    }
  }

  /**
   * Start ping interval to keep connection alive
   */
  private startPingInterval(): void {
    if (this.pingInterval) return;

    this.pingInterval = setInterval(() => {
      this.sendPing();
    }, this.pingIntervalMs);
  }

  /**
   * Stop ping interval
   */
  private stopPingInterval(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }

    if (this.pongTimeout) {
      clearTimeout(this.pongTimeout);
      this.pongTimeout = null;
    }
  }

  /**
   * Send ping message
   */
  private sendPing(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const pingTime = Date.now();
      this.sendMessage({ type: 'ping', timestamp: pingTime });

      // Set timeout for pong response
      this.pongTimeout = setTimeout(() => {
        console.warn('[WS] Pong timeout - connection may be dead');
        this.ws?.close();
      }, this.pongTimeoutMs);
    }
  }

  /**
   * Handle pong response
   */
  private handlePong(): void {
    if (this.pongTimeout) {
      clearTimeout(this.pongTimeout);
      this.pongTimeout = null;
    }

    const now = Date.now();
    if (this.lastPongTime > 0) {
      const latency = now - this.lastPongTime;
      this.recordLatency(latency);
    }
    this.lastPongTime = now;
  }

  /**
   * Record latency measurement
   */
  private recordLatency(latency: number): void {
    this.latencyMeasurements.push(latency);
    if (this.latencyMeasurements.length > this.maxLatencyMeasurements) {
      this.latencyMeasurements.shift();
    }

    this.metrics.lastLatency = latency;
    this.metrics.averageLatency =
      this.latencyMeasurements.reduce((sum, l) => sum + l, 0) /
      this.latencyMeasurements.length;

    this.notifyMetricsListeners();
  }

  /**
   * Attempt to reconnect with exponential backoff
   */
  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WS] Max reconnect attempts reached');
      this.updateState('error');
      return;
    }

    this.reconnectAttempts++;
    this.metrics.reconnectAttempts++;

    const delay = Math.min(
      this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      this.maxReconnectDelay
    );

    console.log(
      `[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Resubscribe to all channels after reconnection
   */
  private resubscribeChannels(): void {
    console.log(`[WS] Resubscribing to ${this.subscribers.size} channels`);

    this.subscribers.forEach((_, channel) => {
      if (channel !== '*') {
        this.sendMessage({
          type: 'subscribe',
          channel,
        });
      }
    });
  }

  /**
   * Update connection state and notify listeners
   */
  private updateState(state: ConnectionState): void {
    if (this.connectionState === state) return;

    console.log(`[WS] State change: ${this.connectionState} -> ${state}`);
    this.connectionState = state;

    this.stateListeners.forEach((listener) => {
      try {
        listener(state);
      } catch (error) {
        console.error('[WS] State listener error:', error);
      }
    });
  }

  /**
   * Notify metrics listeners
   */
  private notifyMetricsListeners(): void {
    this.metricsListeners.forEach((listener) => {
      try {
        listener(this.metrics);
      } catch (error) {
        console.error('[WS] Metrics listener error:', error);
      }
    });
  }

  /**
   * Add auth token to WebSocket URL
   */
  private addTokenToUrl(url: string, token: string | null): string {
    if (!token) {
      console.warn('[WS] No auth token found');
      return url;
    }

    try {
      const urlObj = new URL(url, window.location.origin);
      urlObj.searchParams.set('token', token);
      return urlObj.toString();
    } catch (error) {
      console.error('[WS] Failed to add token to URL:', error);
      return url;
    }
  }

  /**
   * Handle online event
   */
  private handleOnline = (): void => {
    console.log('[WS] Network online - attempting reconnect');
    if (this.connectionState === 'offline' || this.connectionState === 'disconnected') {
      this.connect();
    }
  };

  /**
   * Handle offline event
   */
  private handleOffline = (): void => {
    console.log('[WS] Network offline');
    this.updateState('offline');
  };

  /**
   * Cleanup resources
   */
  public destroy(): void {
    this.disconnect();

    if (typeof window !== 'undefined') {
      window.removeEventListener('online', this.handleOnline);
      window.removeEventListener('offline', this.handleOffline);
    }

    this.subscribers.clear();
    this.stateListeners.clear();
    this.metricsListeners.clear();
    this.tickBuffer.clear();
    this.messageQueue = [];
  }
}

// Singleton instance
let wsInstance: EnhancedWebSocketService | null = null;

export const getEnhancedWebSocketService = (url?: string): EnhancedWebSocketService => {
  if (!wsInstance && url) {
    wsInstance = new EnhancedWebSocketService(url);
  }

  if (!wsInstance) {
    throw new Error('WebSocket service not initialized. Provide URL on first call.');
  }

  return wsInstance;
};

// Export for cleanup
export const destroyWebSocketService = (): void => {
  if (wsInstance) {
    wsInstance.destroy();
    wsInstance = null;
  }
};
