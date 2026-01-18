/**
 * Global Application State Management using Zustand
 * Centralized state for trading application
 */

import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

export interface Tick {
  symbol: string;
  bid: number;
  ask: number;
  spread: number;
  timestamp: number;
  lp?: string;
  prevBid?: number;
  prevAsk?: number;
}

export interface Position {
  id: number;
  symbol: string;
  side: 'BUY' | 'SELL';
  volume: number;
  openPrice: number;
  currentPrice: number;
  openTime: string;
  sl: number;
  tp: number;
  swap: number;
  commission: number;
  unrealizedPnL: number;
}

export interface Order {
  id: string;
  symbol: string;
  type: 'LIMIT' | 'STOP' | 'STOP_LIMIT';
  side: 'BUY' | 'SELL';
  volume: number;
  price: number;
  triggerPrice?: number;
  sl?: number;
  tp?: number;
  status: 'PENDING' | 'TRIGGERED' | 'FILLED' | 'CANCELLED';
  createdAt: string;
}

export interface Account {
  balance: number;
  equity: number;
  margin: number;
  freeMargin: number;
  unrealizedPL: number;
  marginLevel: number;
  currency: string;
}

export interface Trade {
  id: number;
  symbol: string;
  type: string;
  side: string;
  volume: number;
  openPrice: number;
  closePrice?: number;
  openTime: string;
  closeTime?: string;
  profit: number;
  commission: number;
  swap: number;
}

interface AppState {
  // Authentication
  isAuthenticated: boolean;
  accountId: string | null;
  authToken: string | null;

  // Market Data
  ticks: Record<string, Tick>;
  selectedSymbol: string;

  // Trading
  positions: Position[];
  orders: Order[];
  trades: Trade[];

  // Account
  account: Account | null;

  // UI State
  isChartMaximized: boolean;
  chartType: 'candlestick' | 'line' | 'area';
  timeframe: '1m' | '5m' | '15m' | '1h' | '4h' | '1d';
  orderVolume: number;

  // WebSocket
  wsConnected: boolean;

  // Loading States
  isLoadingAccount: boolean;
  isLoadingPositions: boolean;
  isPlacingOrder: boolean;

  // Actions
  setAuthenticated: (isAuth: boolean, accountId?: string, authToken?: string) => void;
  setAuthToken: (token: string | null) => void;
  clearAuth: () => void;
  setTick: (symbol: string, tick: Tick) => void;
  setTicks: (ticks: Record<string, Tick>) => void;
  setSelectedSymbol: (symbol: string) => void;
  setPositions: (positions: Position[]) => void;
  setOrders: (orders: Order[]) => void;
  setTrades: (trades: Trade[]) => void;
  setAccount: (account: Account | null) => void;
  setChartMaximized: (isMaximized: boolean) => void;
  setChartType: (type: 'candlestick' | 'line' | 'area') => void;
  setTimeframe: (tf: '1m' | '5m' | '15m' | '1h' | '4h' | '1d') => void;
  setOrderVolume: (volume: number) => void;
  setWsConnected: (connected: boolean) => void;
  setLoadingStates: (states: Partial<{
    isLoadingAccount: boolean;
    isLoadingPositions: boolean;
    isPlacingOrder: boolean;
  }>) => void;
  reset: () => void;
}

const initialState = {
  isAuthenticated: false,
  accountId: null,
  authToken: null,
  ticks: {},
  selectedSymbol: '',
  positions: [],
  orders: [],
  trades: [],
  account: null,
  isChartMaximized: false,
  chartType: 'candlestick' as const,
  timeframe: '1m' as const,
  orderVolume: 0.01,
  wsConnected: false,
  isLoadingAccount: false,
  isLoadingPositions: false,
  isPlacingOrder: false,
};

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set) => ({
        ...initialState,

        setAuthenticated: (isAuth, accountId, authToken) =>
          set({ isAuthenticated: isAuth, accountId: accountId || null, authToken: authToken || null }),

        setAuthToken: (token) =>
          set({ authToken: token }),

        clearAuth: () =>
          set({ isAuthenticated: false, accountId: null, authToken: null }),

        setTick: (symbol, tick) =>
          set((state) => {
            const prevTick = state.ticks[symbol];
            return {
              ticks: {
                ...state.ticks,
                [symbol]: {
                  ...tick,
                  prevBid: prevTick?.bid,
                  prevAsk: prevTick?.ask,
                },
              },
            };
          }),

        setTicks: (ticks) => set({ ticks }),

        setSelectedSymbol: (symbol) => set({ selectedSymbol: symbol }),

        setPositions: (positions) => set({ positions }),

        setOrders: (orders) => set({ orders }),

        setTrades: (trades) => set({ trades }),

        setAccount: (account) => set({ account }),

        setChartMaximized: (isMaximized) => set({ isChartMaximized: isMaximized }),

        setChartType: (type) => set({ chartType: type }),

        setTimeframe: (tf) => set({ timeframe: tf }),

        setOrderVolume: (volume) => set({ orderVolume: volume }),

        setWsConnected: (connected) => set({ wsConnected: connected }),

        setLoadingStates: (states) => set(states),

        reset: () => set(initialState),
      }),
      {
        name: 'trading-app-storage',
        partialize: (state) => ({
          selectedSymbol: state.selectedSymbol,
          chartType: state.chartType,
          timeframe: state.timeframe,
          orderVolume: state.orderVolume,
        }),
      }
    )
  )
);
