import type { WSMessage, WSMessageType } from '@/types';

type MessageHandler<T = unknown> = (data: T) => void;
type ErrorHandler = (error: Error) => void;
type ConnectionHandler = () => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private handlers = new Map<WSMessageType, Set<MessageHandler>>();
  private errorHandlers = new Set<ErrorHandler>();
  private connectHandlers = new Set<ConnectionHandler>();
  private disconnectHandlers = new Set<ConnectionHandler>();
  private reconnectTimer: NodeJS.Timeout | null = null;

  constructor(url: string) {
    this.url = url;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('[WebSocket] Connected');
          this.reconnectAttempts = 0;
          this.connectHandlers.forEach(handler => handler());
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WSMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('[WebSocket] Failed to parse message:', error);
          }
        };

        this.ws.onerror = (event) => {
          console.error('[WebSocket] Error:', event);
          const error = new Error('WebSocket error');
          this.errorHandlers.forEach(handler => handler(error));
          reject(error);
        };

        this.ws.onclose = () => {
          console.log('[WebSocket] Disconnected');
          this.disconnectHandlers.forEach(handler => handler());
          this.scheduleReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('[WebSocket] Max reconnection attempts reached');
      return;
    }

    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
    console.log(`[WebSocket] Reconnecting in ${delay}ms...`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect().catch(error => {
        console.error('[WebSocket] Reconnection failed:', error);
      });
    }, delay);
  }

  private handleMessage(message: WSMessage): void {
    const handlers = this.handlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message.data);
        } catch (error) {
          console.error(`[WebSocket] Handler error for ${message.type}:`, error);
        }
      });
    }
  }

  on<T = unknown>(type: WSMessageType, handler: MessageHandler<T>): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set());
    }
    this.handlers.get(type)!.add(handler as MessageHandler);

    return () => {
      const handlers = this.handlers.get(type);
      if (handlers) {
        handlers.delete(handler as MessageHandler);
      }
    };
  }

  onError(handler: ErrorHandler): () => void {
    this.errorHandlers.add(handler);
    return () => this.errorHandlers.delete(handler);
  }

  onConnect(handler: ConnectionHandler): () => void {
    this.connectHandlers.add(handler);
    return () => this.connectHandlers.delete(handler);
  }

  onDisconnect(handler: ConnectionHandler): () => void {
    this.disconnectHandlers.add(handler);
    return () => this.disconnectHandlers.delete(handler);
  }

  send<T = unknown>(type: WSMessageType, data: T): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('[WebSocket] Cannot send message: not connected');
      return;
    }

    const message: WSMessage<T> = {
      type,
      timestamp: Date.now(),
      data,
    };

    this.ws.send(JSON.stringify(message));
  }

  get isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}

// Singleton instance
let wsInstance: WebSocketService | null = null;

export const getWebSocketService = (): WebSocketService => {
  if (!wsInstance) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    wsInstance = new WebSocketService(`${protocol}//${host}/ws`);
  }
  return wsInstance;
};
