import { EventEmitter } from 'events';
import { Ticker, Position, Order, Notification } from '@/types';

const WS_URL = __DEV__
  ? 'ws://localhost:8080/ws'
  : 'wss://api.tradingengine.com/ws';

export type WebSocketMessage =
  | { type: 'ticker'; data: Ticker }
  | { type: 'position_update'; data: Position }
  | { type: 'order_update'; data: Order }
  | { type: 'notification'; data: Notification }
  | { type: 'pong'; data: { timestamp: number } };

class WebSocketService extends EventEmitter {
  private ws: WebSocket | null = null;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private pingInterval: NodeJS.Timeout | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private subscribedSymbols = new Set<string>();
  private isConnecting = false;

  connect(token: string): void {
    if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) {
      return;
    }

    this.isConnecting = true;
    this.ws = new WebSocket(`${WS_URL}?token=${token}`);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.isConnecting = false;
      this.reconnectAttempts = 0;
      this.emit('connected');

      // Start ping/pong heartbeat
      this.startHeartbeat();

      // Re-subscribe to symbols
      this.subscribedSymbols.forEach(symbol => {
        this.subscribe(symbol);
      });
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.emit('error', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.isConnecting = false;
      this.stopHeartbeat();
      this.emit('disconnected');
      this.handleReconnect(token);
    };
  }

  disconnect(): void {
    this.subscribedSymbols.clear();
    this.stopHeartbeat();

    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  subscribe(symbol: string): void {
    this.subscribedSymbols.add(symbol);
    this.send({
      type: 'subscribe',
      symbols: [symbol],
    });
  }

  unsubscribe(symbol: string): void {
    this.subscribedSymbols.delete(symbol);
    this.send({
      type: 'unsubscribe',
      symbols: [symbol],
    });
  }

  private send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  private handleMessage(message: WebSocketMessage): void {
    switch (message.type) {
      case 'ticker':
        this.emit('ticker', message.data);
        break;
      case 'position_update':
        this.emit('position_update', message.data);
        break;
      case 'order_update':
        this.emit('order_update', message.data);
        break;
      case 'notification':
        this.emit('notification', message.data);
        break;
      case 'pong':
        // Heartbeat received
        break;
      default:
        console.warn('Unknown message type:', message);
    }
  }

  private handleReconnect(token: string): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached');
      this.emit('max_reconnect_reached');
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    this.reconnectAttempts++;

    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

    this.reconnectTimeout = setTimeout(() => {
      this.connect(token);
    }, delay);
  }

  private startHeartbeat(): void {
    this.pingInterval = setInterval(() => {
      this.send({ type: 'ping', timestamp: Date.now() });
    }, 30000); // 30 seconds
  }

  private stopHeartbeat(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const websocketService = new WebSocketService();
