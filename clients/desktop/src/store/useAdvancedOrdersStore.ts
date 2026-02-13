import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { ordersApi } from '../services/api';
import { useAppStore } from './useAppStore';
import { useNotificationStore } from './useNotificationStore';

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

  // Loading / error state
  isSubmitting: boolean;
  submitError: string | null;

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
      isSubmitting: false,
      submitError: null,

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

      // Submit actions (real API calls)
      submitOCOOrder: async () => {
        const { ocoOrder } = get();
        const notify = useNotificationStore.getState().addNotification;
        set({ isSubmitting: true, submitError: null });

        try {
          // OCO = place two linked orders. Place each order individually via
          // the appropriate limit/stop endpoint, then the backend links them
          // via OCOPairID when both are pending.
          const order1Promise = ocoOrder.order1.type === 'LIMIT'
            ? ordersApi.placeLimitOrder({
                symbol: ocoOrder.symbol,
                side: ocoOrder.order1.side,
                volume: ocoOrder.volume,
                price: ocoOrder.order1.price,
              })
            : ordersApi.placeStopOrder({
                symbol: ocoOrder.symbol,
                side: ocoOrder.order1.side,
                volume: ocoOrder.volume,
                triggerPrice: ocoOrder.order1.price,
              });

          const order2Promise = ocoOrder.order2.type === 'LIMIT'
            ? ordersApi.placeLimitOrder({
                symbol: ocoOrder.symbol,
                side: ocoOrder.order2.side,
                volume: ocoOrder.volume,
                price: ocoOrder.order2.price,
              })
            : ordersApi.placeStopOrder({
                symbol: ocoOrder.symbol,
                side: ocoOrder.order2.side,
                volume: ocoOrder.volume,
                triggerPrice: ocoOrder.order2.price,
              });

          await Promise.all([order1Promise, order2Promise]);

          notify({
            type: 'trading',
            severity: 'info',
            title: 'OCO Order Placed',
            message: `OCO order pair submitted for ${ocoOrder.symbol}: ${ocoOrder.order1.type} ${ocoOrder.order1.side} @ ${ocoOrder.order1.price} / ${ocoOrder.order2.type} ${ocoOrder.order2.side} @ ${ocoOrder.order2.price}`,
          });
        } catch (error: any) {
          const errorMsg = error?.message || 'Failed to submit OCO order';
          set({ submitError: errorMsg });
          notify({
            type: 'trading',
            severity: 'error',
            title: 'OCO Order Failed',
            message: errorMsg,
          });
          throw error;
        } finally {
          set({ isSubmitting: false });
        }
      },

      submitTrailingStopOrder: async () => {
        const { trailingStopOrder } = get();
        const notify = useNotificationStore.getState().addNotification;
        set({ isSubmitting: true, submitError: null });

        try {
          // First place a market order to open the position, then set trailing stop
          const accountId = Number(useAppStore.getState().accountId) || 1;

          const position = await ordersApi.placeMarketOrder({
            accountId,
            symbol: trailingStopOrder.symbol,
            side: trailingStopOrder.side,
            volume: trailingStopOrder.volume,
            sl: trailingStopOrder.stopLoss,
            tp: trailingStopOrder.takeProfit,
          });

          // Now set the trailing stop on the newly opened position
          await ordersApi.setTrailingStop({
            tradeId: String((position as any)?.id || (position as any)?.positionId || ''),
            symbol: trailingStopOrder.symbol,
            side: trailingStopOrder.side,
            type: 'FIXED',
            distance: trailingStopOrder.trailingDistance,
          });

          notify({
            type: 'trading',
            severity: 'info',
            title: 'Trailing Stop Order Placed',
            message: `${trailingStopOrder.side} ${trailingStopOrder.volume} lots ${trailingStopOrder.symbol} with ${trailingStopOrder.trailingDistance} pip trailing stop`,
          });
        } catch (error: any) {
          const errorMsg = error?.message || 'Failed to submit trailing stop order';
          set({ submitError: errorMsg });
          notify({
            type: 'trading',
            severity: 'error',
            title: 'Trailing Stop Order Failed',
            message: errorMsg,
          });
          throw error;
        } finally {
          set({ isSubmitting: false });
        }
      },

      submitTimeBasedOrder: async () => {
        const { timeBasedOrder } = get();
        const notify = useNotificationStore.getState().addNotification;
        set({ isSubmitting: true, submitError: null });

        try {
          if (timeBasedOrder.type === 'MARKET') {
            // Market orders execute immediately -- place via market endpoint
            const accountId = Number(useAppStore.getState().accountId) || 1;
            await ordersApi.placeMarketOrder({
              accountId,
              symbol: timeBasedOrder.symbol,
              side: timeBasedOrder.side,
              volume: timeBasedOrder.volume,
            });
          } else if (timeBasedOrder.type === 'LIMIT') {
            await ordersApi.placeLimitOrder({
              symbol: timeBasedOrder.symbol,
              side: timeBasedOrder.side,
              volume: timeBasedOrder.volume,
              price: timeBasedOrder.price || 0,
            });
          } else if (timeBasedOrder.type === 'STOP') {
            await ordersApi.placeStopOrder({
              symbol: timeBasedOrder.symbol,
              side: timeBasedOrder.side,
              volume: timeBasedOrder.volume,
              triggerPrice: timeBasedOrder.price || 0,
            });
          }

          notify({
            type: 'trading',
            severity: 'info',
            title: 'Time-Based Order Placed',
            message: `${timeBasedOrder.type} ${timeBasedOrder.side} ${timeBasedOrder.volume} lots ${timeBasedOrder.symbol} (${timeBasedOrder.timeInForce}${timeBasedOrder.timeInForce === 'GTD' && timeBasedOrder.expiryDate ? ` until ${timeBasedOrder.expiryDate}` : ''})`,
          });
        } catch (error: any) {
          const errorMsg = error?.message || 'Failed to submit time-based order';
          set({ submitError: errorMsg });
          notify({
            type: 'trading',
            severity: 'error',
            title: 'Time-Based Order Failed',
            message: errorMsg,
          });
          throw error;
        } finally {
          set({ isSubmitting: false });
        }
      },

      submitBracketOrder: async () => {
        const { bracketOrder } = get();
        const notify = useNotificationStore.getState().addNotification;
        set({ isSubmitting: true, submitError: null });

        try {
          await ordersApi.placeBracketOrder({
            symbol: bracketOrder.symbol,
            side: bracketOrder.side,
            volume: bracketOrder.volume,
            entryPrice: bracketOrder.entryPrice,
            stopLoss: bracketOrder.stopLossPrice,
            takeProfit: bracketOrder.takeProfitPrice,
            entryType: 'LIMIT',
            timeInForce: 'GTC',
          });

          notify({
            type: 'trading',
            severity: 'info',
            title: 'Bracket Order Placed',
            message: `${bracketOrder.side} ${bracketOrder.volume} lots ${bracketOrder.symbol} @ ${bracketOrder.entryPrice} | SL: ${bracketOrder.stopLossPrice} | TP: ${bracketOrder.takeProfitPrice}`,
          });
        } catch (error: any) {
          const errorMsg = error?.message || 'Failed to submit bracket order';
          set({ submitError: errorMsg });
          notify({
            type: 'trading',
            severity: 'error',
            title: 'Bracket Order Failed',
            message: errorMsg,
          });
          throw error;
        } finally {
          set({ isSubmitting: false });
        }
      },

      submitScaledOrder: async () => {
        const { scaledOrder, calculateScaledOrderPreview } = get();
        const notify = useNotificationStore.getState().addNotification;
        set({ isSubmitting: true, submitError: null });

        try {
          // Place each scaled order as an individual limit order
          const preview = calculateScaledOrderPreview();
          const orderPromises = preview.map((slice) =>
            ordersApi.placeLimitOrder({
              symbol: scaledOrder.symbol,
              side: scaledOrder.side,
              volume: slice.volume,
              price: slice.price,
            })
          );

          const results = await Promise.allSettled(orderPromises);
          const succeeded = results.filter((r) => r.status === 'fulfilled').length;
          const failed = results.filter((r) => r.status === 'rejected').length;

          if (failed > 0 && succeeded === 0) {
            const firstError = results.find((r) => r.status === 'rejected') as PromiseRejectedResult;
            throw new Error(firstError.reason?.message || 'All scaled orders failed');
          }

          notify({
            type: 'trading',
            severity: failed > 0 ? 'warning' : 'info',
            title: 'Scaled Order Placed',
            message: failed > 0
              ? `${succeeded}/${preview.length} limit orders placed for ${scaledOrder.symbol} (${failed} failed)`
              : `${succeeded} limit orders placed for ${scaledOrder.side} ${scaledOrder.totalVolume} lots ${scaledOrder.symbol} from ${scaledOrder.startPrice} to ${scaledOrder.endPrice}`,
          });
        } catch (error: any) {
          const errorMsg = error?.message || 'Failed to submit scaled order';
          set({ submitError: errorMsg });
          notify({
            type: 'trading',
            severity: 'error',
            title: 'Scaled Order Failed',
            message: errorMsg,
          });
          throw error;
        } finally {
          set({ isSubmitting: false });
        }
      },

      resetAllOrders: () => {
        set({
          ocoOrder: DEFAULT_OCO_ORDER,
          trailingStopOrder: DEFAULT_TRAILING_STOP,
          timeBasedOrder: DEFAULT_TIME_BASED,
          bracketOrder: DEFAULT_BRACKET,
          scaledOrder: DEFAULT_SCALED,
          isSubmitting: false,
          submitError: null,
        });
      },
    }),
    { name: 'AdvancedOrdersStore' }
  )
);
