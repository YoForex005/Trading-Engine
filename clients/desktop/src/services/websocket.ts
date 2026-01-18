/**
 * WebSocket Service with Auto-Reconnection and State Management
 * Handles real-time market data and account updates
 */

export type SubscriptionCallback = (data: any) => void;
export type ConnectionStateCallback = (state: ConnectionState) => void;

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'error';

interface WebSocketMessage {
  type: string;
  symbol?: string;
  bid?: number;
  ask?: number;
  spread?: number;
  timestamp?: number;
  lp?: string;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private pingInterval: ReturnType<typeof setInterval> | null = null;

  private subscribers: Map<string, Set<SubscriptionCallback>> = new Map();
  private stateListeners: Set<ConnectionStateCallback> = new Set();
  private connectionState: ConnectionState = 'disconnected';

  private tickBuffer: Map<string, WebSocketMessage> = new Map();
  private flushInterval: ReturnType<typeof setInterval> | null = null;
  private updateThrottle = 100; // ms

  constructor(url: string) {
    this.url = url;
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

    // Get auth token and add to URL
    const token = localStorage.getItem('rtx_token');
    const urlWithAuth = this.addTokenToUrl(this.url, token);

    console.log(`[WS] Connecting to ${this.url}...`);

    try {
      this.ws = new WebSocket(urlWithAuth);

      this.ws.onopen = () => {
        console.log('[WS] Connected successfully');
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;
        this.updateState('connected');
        this.startPingInterval();
        this.startFlushInterval();
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as WebSocketMessage;
          this.handleMessage(data);
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
        this.updateState('disconnected');
        this.stopPingInterval();
        this.stopFlushInterval();

        // Check if close was due to authentication failure (1008 = policy violation)
        if (event.code === 1008 || event.code === 401) {
          console.error('[WS] Authentication failed - clearing token and attempting re-authentication');
          localStorage.removeItem('rtx_token');
          // Dispatch auth failure event (let app handle login redirect)
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
      this.ws.close();
      this.ws = null;
    }

    this.reconnectAttempts = 0;
    this.updateState('disconnected');
  }

  /**
   * Subscribe to market data for a specific symbol
   */
  public subscribe(channel: string, callback: SubscriptionCallback): () => void {
    if (!this.subscribers.has(channel)) {
      this.subscribers.set(channel, new Set());
    }

    this.subscribers.get(channel)!.add(callback);

    // Send subscription message if needed
    this.sendMessage({
      type: 'subscribe',
      channel
    });

    // Return unsubscribe function
    return () => {
      const callbacks = this.subscribers.get(channel);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.subscribers.delete(channel);
          this.sendMessage({
            type: 'unsubscribe',
            channel
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

    // Immediately call with current state
    callback(this.connectionState);

    // Return unsubscribe function
    return () => {
      this.stateListeners.delete(callback);
    };
  }

  /**
   * Get current connection state
   */
  public getState(): ConnectionState {
    return this.connectionState;
  }

  /**
   * Send message to WebSocket server
   */
  private sendMessage(message: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  /**
   * Handle incoming WebSocket messages
   */
  private handleMessage(data: WebSocketMessage): void {
    const { type, symbol } = data;

    // Buffer tick updates for throttling
    if (type === 'tick' && symbol) {
      this.tickBuffer.set(symbol, data);
      return;
    }

    // Broadcast to all subscribers of this channel
    const channel = symbol || type;
    const callbacks = this.subscribers.get(channel);
    if (callbacks) {
      callbacks.forEach(callback => callback(data));
    }

    // Broadcast to wildcard subscribers
    const wildcardCallbacks = this.subscribers.get('*');
    if (wildcardCallbacks) {
      wildcardCallbacks.forEach(callback => callback(data));
    }
  }

  /**
   * Flush buffered tick updates
   */
  private flushTickBuffer(): void {
    if (this.tickBuffer.size === 0) return;

    this.tickBuffer.forEach((data, symbol) => {
      const callbacks = this.subscribers.get(symbol);
      if (callbacks) {
        callbacks.forEach(callback => callback(data));
      }

      // Broadcast to wildcard subscribers
      const wildcardCallbacks = this.subscribers.get('*');
      if (wildcardCallbacks) {
        wildcardCallbacks.forEach(callback => callback(data));
      }
    });

    this.tickBuffer.clear();
  }

  /**
   * Start flush interval for throttled tick updates
   */
  private startFlushInterval(): void {
    if (this.flushInterval) return;

    this.flushInterval = setInterval(() => {
      this.flushTickBuffer();
    }, this.updateThrottle);
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
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.sendMessage({ type: 'ping' });
      }
    }, 30000); // 30 seconds
  }

  /**
   * Stop ping interval
   */
  private stopPingInterval(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
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
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      30000 // Max 30 seconds
    );

    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Update connection state and notify listeners
   */
  private updateState(state: ConnectionState): void {
    if (this.connectionState === state) return;

    this.connectionState = state;
    this.stateListeners.forEach(listener => listener(state));
  }

  /**
   * Add auth token to WebSocket URL as query parameter
   */
  private addTokenToUrl(url: string, token: string | null): string {
    if (!token) {
      console.warn('[WS] No auth token found in localStorage');
      return url;
    }

    const urlObj = new URL(url, window.location.origin);
    urlObj.searchParams.set('token', token);
    return urlObj.toString();
  }
}

// Singleton instance
let wsInstance: WebSocketService | null = null;

export const getWebSocketService = (url?: string): WebSocketService => {
  if (!wsInstance && url) {
    wsInstance = new WebSocketService(url);
  }

  if (!wsInstance) {
    throw new Error('WebSocket service not initialized. Provide URL on first call.');
  }

  return wsInstance;
};
