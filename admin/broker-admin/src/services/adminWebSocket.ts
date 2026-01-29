/**
 * Admin WebSocket Service
 * Real-time updates for admin panel (MT5 Manager parity)
 * - Account updates (balance, equity, P/L)
 * - Order updates (new, modified, closed)
 * - Position updates (real-time P/L)
 * - System events (margin calls, suspensions)
 * PERFORMANCE FIX: Real-time admin data instead of polling
 */

export type AdminEventType =
  | 'account_update'
  | 'order_new'
  | 'order_modify'
  | 'order_close'
  | 'position_update'
  | 'margin_call'
  | 'account_suspended'
  | 'system_event';

export interface AdminEvent {
  type: AdminEventType;
  timestamp: number;
  data: any;
}

export interface Account {
  id: number;
  login: string;
  name: string;
  group: string;
  leverage: string;
  balance: string;
  credit: string;
  equity: string;
  margin: string;
  freeMargin: string;
  marginLevel: string;
  profit: string;
  floatingPL: string;
  swap: string;
  commission: string;
  status: 'ACTIVE' | 'SUSPENDED' | 'MARGIN_CALL';
  currency: string;
  country?: string;
  email?: string;
  comment?: string;
  [key: string]: any;
}

export interface Order {
  id: number;
  login: string;
  symbol: string;
  type: string;
  volume: number;
  price: number;
  sl: number;
  tp: number;
  profit: number;
  openTime: string;
  [key: string]: any;
}

type EventCallback = (event: AdminEvent) => void;
type ConnectionStateCallback = (state: 'connecting' | 'connected' | 'disconnected' | 'error') => void;

class AdminWebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectDelay = 1000;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private pingInterval: ReturnType<typeof setInterval> | null = null;

  private eventListeners: Map<AdminEventType | '*', Set<EventCallback>> = new Map();
  private stateListeners: Set<ConnectionStateCallback> = new Set();
  private connectionState: 'connecting' | 'connected' | 'disconnected' | 'error' = 'disconnected';

  constructor(url: string) {
    this.url = url;
  }

  /**
   * Connect to admin WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[AdminWS] Already connected');
      return;
    }

    this.updateState('connecting');

    // Get auth token from localStorage
    const token = localStorage.getItem('admin_token') || localStorage.getItem('rtx_token');
    const urlWithAuth = token ? `${this.url}?token=${encodeURIComponent(token)}` : this.url;

    console.log(`[AdminWS] Connecting to ${this.url}...`);

    try {
      this.ws = new WebSocket(urlWithAuth);

      this.ws.onopen = () => {
        console.log('[AdminWS] Connected successfully');
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;
        this.updateState('connected');
        this.startPingInterval();

        // Subscribe to all admin events
        this.send({
          type: 'subscribe',
          channels: ['accounts', 'orders', 'positions', 'system'],
        });
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.handleMessage(data);
        } catch (error) {
          console.error('[AdminWS] Failed to parse message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('[AdminWS] WebSocket error:', error);
        this.updateState('error');
      };

      this.ws.onclose = (event) => {
        console.log(`[AdminWS] Connection closed (code: ${event.code})`);
        this.updateState('disconnected');
        this.stopPingInterval();

        // Check for auth failure
        if (event.code === 1008 || event.code === 401) {
          console.error('[AdminWS] Authentication failed');
          // Don't auto-reconnect on auth failure
          return;
        }

        this.attemptReconnect();
      };
    } catch (error) {
      console.error('[AdminWS] Connection failed:', error);
      this.updateState('error');
      this.attemptReconnect();
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    console.log('[AdminWS] Disconnecting...');

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    this.stopPingInterval();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.reconnectAttempts = 0;
    this.updateState('disconnected');
  }

  /**
   * Subscribe to specific admin event types
   * @param eventType - Event type or '*' for all events
   * @param callback - Callback function
   * @returns Unsubscribe function
   */
  on(eventType: AdminEventType | '*', callback: EventCallback): () => void {
    if (!this.eventListeners.has(eventType)) {
      this.eventListeners.set(eventType, new Set());
    }

    this.eventListeners.get(eventType)!.add(callback);

    // Return unsubscribe function
    return () => {
      const callbacks = this.eventListeners.get(eventType);
      if (callbacks) {
        callbacks.delete(callback);
        if (callbacks.size === 0) {
          this.eventListeners.delete(eventType);
        }
      }
    };
  }

  /**
   * Listen to connection state changes
   */
  onStateChange(callback: ConnectionStateCallback): () => void {
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
  getState(): typeof this.connectionState {
    return this.connectionState;
  }

  /**
   * Send message to server
   */
  private send(message: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('[AdminWS] Cannot send message - not connected');
    }
  }

  /**
   * Handle incoming messages
   */
  private handleMessage(data: any): void {
    const event: AdminEvent = {
      type: data.type,
      timestamp: data.timestamp || Date.now(),
      data: data.data || data,
    };

    // Notify specific event listeners
    const specificListeners = this.eventListeners.get(event.type);
    if (specificListeners) {
      specificListeners.forEach((callback) => callback(event));
    }

    // Notify wildcard listeners
    const wildcardListeners = this.eventListeners.get('*');
    if (wildcardListeners) {
      wildcardListeners.forEach((callback) => callback(event));
    }
  }

  /**
   * Start ping interval to keep connection alive
   */
  private startPingInterval(): void {
    if (this.pingInterval) return;

    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.send({ type: 'ping' });
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
      console.error('[AdminWS] Max reconnect attempts reached');
      this.updateState('error');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(
      this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
      30000 // Max 30 seconds
    );

    console.log(
      `[AdminWS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`
    );

    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Update connection state and notify listeners
   */
  private updateState(state: typeof this.connectionState): void {
    if (this.connectionState === state) return;

    this.connectionState = state;
    this.stateListeners.forEach((listener) => listener(state));
  }
}

// Singleton instance
let adminWSInstance: AdminWebSocketService | null = null;

/**
 * Get admin WebSocket service instance
 */
export const getAdminWebSocket = (url?: string): AdminWebSocketService => {
  if (!adminWSInstance && url) {
    adminWSInstance = new AdminWebSocketService(url);
  }

  if (!adminWSInstance) {
    throw new Error('Admin WebSocket service not initialized. Provide URL on first call.');
  }

  return adminWSInstance;
};

/**
 * Terminate admin WebSocket service
 */
export const terminateAdminWebSocket = (): void => {
  if (adminWSInstance) {
    adminWSInstance.disconnect();
    adminWSInstance = null;
  }
};
