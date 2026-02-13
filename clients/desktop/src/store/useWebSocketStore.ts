/**
 * WebSocket Connection State Store
 * Centralized Zustand store for WebSocket connection status and metrics
 */

import { create } from 'zustand';

export type ConnectionState = 'connecting' | 'connected' | 'disconnected' | 'reconnecting' | 'error' | 'offline';

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

interface WebSocketStoreState {
  // Connection state
  connectionState: ConnectionState;
  isStale: boolean;
  lastMessageTime: number | null;

  // Metrics
  metrics: ConnectionMetrics;

  // Reconnect tracking
  reconnectCount: number;
  lastReconnectAt: number | null;

  // Actions
  setConnectionState: (state: ConnectionState) => void;
  updateMetrics: (metrics: Partial<ConnectionMetrics>) => void;
  markStale: (stale: boolean) => void;
  updateLastMessageTime: (time: number) => void;
  resetReconnectAttempts: () => void;
  incrementReconnectAttempts: () => void;
  recordReconnect: () => void;
}

export const useWebSocketStore = create<WebSocketStoreState>((set) => ({
  // Initial state
  connectionState: 'disconnected',
  isStale: false,
  lastMessageTime: null,
  reconnectCount: 0,
  lastReconnectAt: null,

  metrics: {
    messagesSent: 0,
    messagesReceived: 0,
    reconnectAttempts: 0,
    lastLatency: 0,
    averageLatency: 0,
    uptime: 0,
    lastConnectedAt: null,
    lastDisconnectedAt: null,
  },

  // Actions
  setConnectionState: (connectionState) => set({ connectionState }),

  updateMetrics: (metricsUpdate) =>
    set((state) => ({
      metrics: { ...state.metrics, ...metricsUpdate },
    })),

  markStale: (isStale) => set({ isStale }),

  updateLastMessageTime: (lastMessageTime) => set({ lastMessageTime, isStale: false }),

  resetReconnectAttempts: () =>
    set((state) => ({
      metrics: { ...state.metrics, reconnectAttempts: 0 },
    })),

  incrementReconnectAttempts: () =>
    set((state) => ({
      metrics: {
        ...state.metrics,
        reconnectAttempts: state.metrics.reconnectAttempts + 1,
      },
    })),

  recordReconnect: () =>
    set((state) => ({
      reconnectCount: state.reconnectCount + 1,
      lastReconnectAt: Date.now(),
    })),
}));
