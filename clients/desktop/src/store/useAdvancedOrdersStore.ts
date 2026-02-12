import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export type OrderSide = 'BUY' | 'SELL';
export type TimeInForce = 'GTC' | 'GTD' | 'IOC' | 'FOK';
export type OrderType = 'MARKET' | 'LIMIT' | 'STOP';

// OCO Order State
export interface OCOOrder {
  symbol: string;
  volume: number;
  order1: {
    type: OrderType;
    side: OrderSide;
    price: number;
  };
  order2: {
    type: OrderType;
    side: OrderSide;
    price: number;
  };
}

// Trailing Stop Order State
export interface TrailingStopOrder {
  symbol: string;
  volume: number;
  side: OrderSide;
  trailingDistance: number; // in pips
  stopLoss?: number;
  takeProfit?: number;
}

// Time-based Order State
export interface TimeBasedOrder {
  symbol: string;
  volume: number;
  type: OrderType;
  side: OrderSide;
  price?: number;
  timeInForce: TimeInForce;
  expiryDate?: string; // ISO date string for GTD
}

// Bracket Order State
export interface BracketOrder {
  symbol: string;
  volume: number;
  side: OrderSide;
  entryPrice: number;
  takeProfitPrice: number;
  stopLossPrice: number;
}

// Scaled Order State
export interface ScaledOrder {
  symbol: string;
  totalVolume: number;
  side: OrderSide;
  numberOfOrders: number;
  startPrice: number;
  endPrice: number;
  priceIncrement: number;
}

export interface ScaledOrderPreview {
  orderNumber: number;
  price: number;
  volume: number;
}

interface AdvancedOrdersState {
  // OCO
  ocoOrder: OCOOrder;
  setOCOOrder: (order: Partial<OCOOrder>) => void;

  // Trailing Stop
  trailingStopOrder: TrailingStopOrder;
  setTrailingStopOrder: (order: Partial<TrailingStopOrder>) => void;

  // Time-based
  timeBasedOrder: TimeBasedOrder;
  setTimeBasedOrder: (order: Partial<TimeBasedOrder>) => void;

  // Bracket
  bracketOrder: BracketOrder;
  setBracketOrder: (order: Partial<BracketOrder>) => void;

  // Scaled
  scaledOrder: ScaledOrder;
  setScaledOrder: (order: Partial<ScaledOrder>) => void;
  calculateScaledOrderPreview: () => ScaledOrderPreview[];

  // Actions
  submitOCOOrder: () => Promise<void>;
  submitTrailingStopOrder: () => Promise<void>;
  submitTimeBasedOrder: () => Promise<void>;
  submitBracketOrder: () => Promise<void>;
  submitScaledOrder: () => Promise<void>;

  resetAllOrders: () => void;
}

const DEFAULT_OCO_ORDER: OCOOrder = {
  symbol: 'EURUSD',
  volume: 1.0,
  order1: { type: 'LIMIT', side: 'BUY', price: 1.1000 },
  order2: { type: 'STOP', side: 'SELL', price: 1.0950 },
};

const DEFAULT_TRAILING_STOP: TrailingStopOrder = {
  symbol: 'EURUSD',
  volume: 1.0,
  side: 'BUY',
  trailingDistance: 20,
};

const DEFAULT_TIME_BASED: TimeBasedOrder = {
  symbol: 'EURUSD',
  volume: 1.0,
  type: 'LIMIT',
  side: 'BUY',
  price: 1.1000,
  timeInForce: 'GTC',
};

const DEFAULT_BRACKET: BracketOrder = {
  symbol: 'EURUSD',
  volume: 1.0,
  side: 'BUY',
  entryPrice: 1.1000,
  takeProfitPrice: 1.1050,
  stopLossPrice: 1.0950,
};

const DEFAULT_SCALED: ScaledOrder = {
  symbol: 'EURUSD',
  totalVolume: 5.0,
  side: 'BUY',
  numberOfOrders: 5,
  startPrice: 1.1000,
  endPrice: 1.1020,
  priceIncrement: 0.0005,
};

export const useAdvancedOrdersStore = create<AdvancedOrdersState>()(
  devtools(
    (set, get) => ({
      // Initial states
      ocoOrder: DEFAULT_OCO_ORDER,
      trailingStopOrder: DEFAULT_TRAILING_STOP,
      timeBasedOrder: DEFAULT_TIME_BASED,
      bracketOrder: DEFAULT_BRACKET,
      scaledOrder: DEFAULT_SCALED,

      // Setters
      setOCOOrder: (order) =>
        set((state) => ({
          ocoOrder: { ...state.ocoOrder, ...order },
        })),

      setTrailingStopOrder: (order) =>
        set((state) => ({
          trailingStopOrder: { ...state.trailingStopOrder, ...order },
        })),

      setTimeBasedOrder: (order) =>
        set((state) => ({
          timeBasedOrder: { ...state.timeBasedOrder, ...order },
        })),

      setBracketOrder: (order) =>
        set((state) => ({
          bracketOrder: { ...state.bracketOrder, ...order },
        })),

      setScaledOrder: (order) =>
        set((state) => ({
          scaledOrder: { ...state.scaledOrder, ...order },
        })),

      // Calculate scaled order preview
      calculateScaledOrderPreview: () => {
        const { scaledOrder } = get();
        const preview: ScaledOrderPreview[] = [];
        const volumePerOrder = scaledOrder.totalVolume / scaledOrder.numberOfOrders;

        for (let i = 0; i < scaledOrder.numberOfOrders; i++) {
          preview.push({
            orderNumber: i + 1,
            price: scaledOrder.startPrice + i * scaledOrder.priceIncrement,
            volume: parseFloat(volumePerOrder.toFixed(2)),
          });
        }

        return preview;
      },

      // Submit actions (mock API calls)
      submitOCOOrder: async () => {
        const { ocoOrder } = get();
        console.log('[AdvancedOrders] Submitting OCO Order:', ocoOrder);
        // TODO: Replace with actual API call
        await new Promise((resolve) => setTimeout(resolve, 500));
        alert('OCO Order submitted successfully!');
      },

      submitTrailingStopOrder: async () => {
        const { trailingStopOrder } = get();
        console.log('[AdvancedOrders] Submitting Trailing Stop:', trailingStopOrder);
        await new Promise((resolve) => setTimeout(resolve, 500));
        alert('Trailing Stop Order submitted successfully!');
      },

      submitTimeBasedOrder: async () => {
        const { timeBasedOrder } = get();
        console.log('[AdvancedOrders] Submitting Time-based Order:', timeBasedOrder);
        await new Promise((resolve) => setTimeout(resolve, 500));
        alert('Time-based Order submitted successfully!');
      },

      submitBracketOrder: async () => {
        const { bracketOrder } = get();
        console.log('[AdvancedOrders] Submitting Bracket Order:', bracketOrder);
        await new Promise((resolve) => setTimeout(resolve, 500));
        alert('Bracket Order submitted successfully!');
      },

      submitScaledOrder: async () => {
        const { scaledOrder } = get();
        console.log('[AdvancedOrders] Submitting Scaled Order:', scaledOrder);
        await new Promise((resolve) => setTimeout(resolve, 500));
        alert('Scaled Order submitted successfully!');
      },

      resetAllOrders: () => {
        set({
          ocoOrder: DEFAULT_OCO_ORDER,
          trailingStopOrder: DEFAULT_TRAILING_STOP,
          timeBasedOrder: DEFAULT_TIME_BASED,
          bracketOrder: DEFAULT_BRACKET,
          scaledOrder: DEFAULT_SCALED,
        });
      },
    }),
    { name: 'AdvancedOrdersStore' }
  )
);
