/**
 * Signal Store with Zustand
 * Manages trading signals from signal providers
 */

import { create } from 'zustand';
import { devtools } from 'zustand/middleware';

export type SignalDirection = 'BUY' | 'SELL';
export type SignalStatus = 'new' | 'active' | 'closed';
export type SignalProvider = 'RTX5 Algo' | 'FX Maestro' | 'CryptoEdge';
export type SymbolGroup = 'All' | 'Forex' | 'Crypto' | 'Metals';

export interface Signal {
  id: string;
  provider: SignalProvider;
  symbol: string;
  symbolGroup: 'Forex' | 'Crypto' | 'Metals';
  direction: SignalDirection;
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  confidence: number; // 0-100
  status: SignalStatus;
  timestamp: Date;
  analysis: string;
  indicators: string[];
  suggestedLotSize: number;
  timeframe: string;
  currentPrice?: number;
  closePrice?: number;
  closeTime?: Date;
  pnl?: number;
  duration?: number; // in minutes
}

export interface ProviderStats {
  provider: SignalProvider;
  winRate: number;
  totalSignals: number;
  avgPnLPerSignal: number;
  bestSignal: number;
  worstSignal: number;
  profitFactor: number;
}

interface SignalState {
  signals: Signal[];
  historicalSignals: Signal[];
  selectedProvider: SignalProvider | 'All';
  symbolGroupFilter: SymbolGroup;
  signalTypeFilter: 'All' | SignalDirection;
  timeframeFilter: string;
  autoRefresh: boolean;
  refreshInterval: number; // seconds

  setSelectedProvider: (provider: SignalProvider | 'All') => void;
  setSymbolGroupFilter: (group: SymbolGroup) => void;
  setSignalTypeFilter: (type: 'All' | SignalDirection) => void;
  setTimeframeFilter: (timeframe: string) => void;
  setAutoRefresh: (enabled: boolean) => void;
  setRefreshInterval: (seconds: number) => void;
  getFilteredSignals: () => Signal[];
  getProviderStats: (provider: SignalProvider) => ProviderStats;
  refreshSignals: () => void;
}

// Mock data generators
function generateMockSignals(): Signal[] {
  const providers: SignalProvider[] = ['RTX5 Algo', 'FX Maestro', 'CryptoEdge'];
  const forexSymbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'NZDUSD'];
  const cryptoSymbols = ['BTCUSD', 'ETHUSD', 'XRPUSD'];
  const metalSymbols = ['XAUUSD', 'XAGUSD'];
  const timeframes = ['M15', 'H1', 'H4', 'D1'];

  const signals: Signal[] = [];

  // Generate 15 active signals
  for (let i = 0; i < 15; i++) {
    const provider = providers[Math.floor(Math.random() * providers.length)];
    const symbolGroupRand = Math.random();
    let symbol: string;
    let symbolGroup: 'Forex' | 'Crypto' | 'Metals';

    if (symbolGroupRand < 0.6) {
      symbol = forexSymbols[Math.floor(Math.random() * forexSymbols.length)];
      symbolGroup = 'Forex';
    } else if (symbolGroupRand < 0.85) {
      symbol = cryptoSymbols[Math.floor(Math.random() * cryptoSymbols.length)];
      symbolGroup = 'Crypto';
    } else {
      symbol = metalSymbols[Math.floor(Math.random() * metalSymbols.length)];
      symbolGroup = 'Metals';
    }

    const direction: SignalDirection = Math.random() > 0.5 ? 'BUY' : 'SELL';
    const basePrice = symbolGroup === 'Crypto' ? 50000 + Math.random() * 50000 :
                     symbolGroup === 'Metals' ? 2000 + Math.random() * 500 :
                     1.0 + Math.random() * 0.5;

    const entryPrice = basePrice;
    const slDistance = basePrice * (0.005 + Math.random() * 0.015); // 0.5-2% SL
    const tpDistance = basePrice * (0.01 + Math.random() * 0.03); // 1-4% TP

    const stopLoss = direction === 'BUY' ? entryPrice - slDistance : entryPrice + slDistance;
    const takeProfit = direction === 'BUY' ? entryPrice + tpDistance : entryPrice - tpDistance;

    const confidence = 60 + Math.random() * 35; // 60-95%
    const status: SignalStatus = i < 3 ? 'new' : 'active';
    const timeframe = timeframes[Math.floor(Math.random() * timeframes.length)];

    const indicatorSets = [
      ['RSI oversold', 'MACD crossover'],
      ['EMA crossover', 'Bollinger Bands squeeze'],
      ['Support level bounce', 'Momentum divergence'],
      ['Fibonacci retracement', 'Volume spike'],
      ['Trend continuation', 'Stochastic overbought'],
    ];

    const indicators = indicatorSets[Math.floor(Math.random() * indicatorSets.length)];

    const analysisList = [
      'Strong uptrend continuation with bullish momentum indicators.',
      'Reversal signal detected at key support level with high confidence.',
      'Breakout above resistance confirmed by volume spike.',
      'Pullback to moving average support presents low-risk entry.',
      'Momentum divergence suggests potential trend reversal.',
      'Consolidation pattern breakout with strong directional bias.',
    ];

    const analysis = analysisList[Math.floor(Math.random() * analysisList.length)];

    const minutesAgo = Math.floor(Math.random() * 240); // Last 4 hours
    const timestamp = new Date(Date.now() - minutesAgo * 60 * 1000);

    signals.push({
      id: `sig-${Date.now()}-${i}`,
      provider,
      symbol,
      symbolGroup,
      direction,
      entryPrice,
      stopLoss,
      takeProfit,
      confidence: Math.round(confidence),
      status,
      timestamp,
      analysis,
      indicators,
      suggestedLotSize: 0.01 + Math.random() * 0.5,
      timeframe,
      currentPrice: entryPrice + (Math.random() - 0.5) * (basePrice * 0.005),
    });
  }

  return signals.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
}

function generateHistoricalSignals(): Signal[] {
  const providers: SignalProvider[] = ['RTX5 Algo', 'FX Maestro', 'CryptoEdge'];
  const forexSymbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD'];
  const cryptoSymbols = ['BTCUSD', 'ETHUSD'];
  const metalSymbols = ['XAUUSD'];
  const timeframes = ['M15', 'H1', 'H4'];

  const signals: Signal[] = [];

  // Generate 20 historical signals
  for (let i = 0; i < 20; i++) {
    const provider = providers[Math.floor(Math.random() * providers.length)];
    const symbolGroupRand = Math.random();
    let symbol: string;
    let symbolGroup: 'Forex' | 'Crypto' | 'Metals';

    if (symbolGroupRand < 0.7) {
      symbol = forexSymbols[Math.floor(Math.random() * forexSymbols.length)];
      symbolGroup = 'Forex';
    } else if (symbolGroupRand < 0.9) {
      symbol = cryptoSymbols[Math.floor(Math.random() * cryptoSymbols.length)];
      symbolGroup = 'Crypto';
    } else {
      symbol = metalSymbols[0];
      symbolGroup = 'Metals';
    }

    const direction: SignalDirection = Math.random() > 0.5 ? 'BUY' : 'SELL';
    const basePrice = symbolGroup === 'Crypto' ? 50000 + Math.random() * 50000 :
                     symbolGroup === 'Metals' ? 2000 + Math.random() * 500 :
                     1.0 + Math.random() * 0.5;

    const entryPrice = basePrice;
    const slDistance = basePrice * (0.005 + Math.random() * 0.015);
    const tpDistance = basePrice * (0.01 + Math.random() * 0.03);

    const stopLoss = direction === 'BUY' ? entryPrice - slDistance : entryPrice + slDistance;
    const takeProfit = direction === 'BUY' ? entryPrice + tpDistance : entryPrice - tpDistance;

    // 65% win rate
    const isWin = Math.random() < 0.65;
    const closePrice = isWin ? takeProfit : stopLoss;

    const pnl = direction === 'BUY'
      ? (closePrice - entryPrice) * 100
      : (entryPrice - closePrice) * 100;

    const daysAgo = 1 + Math.floor(Math.random() * 30); // Last 30 days
    const timestamp = new Date(Date.now() - daysAgo * 24 * 60 * 60 * 1000);
    const closeTime = new Date(timestamp.getTime() + (30 + Math.random() * 600) * 60 * 1000);
    const duration = Math.floor((closeTime.getTime() - timestamp.getTime()) / 60000);

    const confidence = 60 + Math.random() * 35;
    const timeframe = timeframes[Math.floor(Math.random() * timeframes.length)];

    signals.push({
      id: `hist-${Date.now()}-${i}`,
      provider,
      symbol,
      symbolGroup,
      direction,
      entryPrice,
      stopLoss,
      takeProfit,
      confidence: Math.round(confidence),
      status: 'closed',
      timestamp,
      analysis: '',
      indicators: [],
      suggestedLotSize: 0.01 + Math.random() * 0.5,
      timeframe,
      closePrice,
      closeTime,
      pnl,
      duration,
    });
  }

  return signals.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
}

export const useSignalStore = create<SignalState>()(
  devtools(
    (set, get) => ({
      signals: generateMockSignals(),
      historicalSignals: generateHistoricalSignals(),
      selectedProvider: 'All',
      symbolGroupFilter: 'All',
      signalTypeFilter: 'All',
      timeframeFilter: 'All',
      autoRefresh: true,
      refreshInterval: 30,

      setSelectedProvider: (provider) => set({ selectedProvider: provider }),
      setSymbolGroupFilter: (group) => set({ symbolGroupFilter: group }),
      setSignalTypeFilter: (type) => set({ signalTypeFilter: type }),
      setTimeframeFilter: (timeframe) => set({ timeframeFilter: timeframe }),
      setAutoRefresh: (enabled) => set({ autoRefresh: enabled }),
      setRefreshInterval: (seconds) => set({ refreshInterval: seconds }),

      getFilteredSignals: () => {
        const state = get();
        let filtered = state.signals;

        if (state.selectedProvider !== 'All') {
          filtered = filtered.filter((s) => s.provider === state.selectedProvider);
        }

        if (state.symbolGroupFilter !== 'All') {
          filtered = filtered.filter((s) => s.symbolGroup === state.symbolGroupFilter);
        }

        if (state.signalTypeFilter !== 'All') {
          filtered = filtered.filter((s) => s.direction === state.signalTypeFilter);
        }

        if (state.timeframeFilter !== 'All') {
          filtered = filtered.filter((s) => s.timeframe === state.timeframeFilter);
        }

        return filtered;
      },

      getProviderStats: (provider) => {
        const state = get();
        const providerSignals = state.historicalSignals.filter((s) => s.provider === provider);

        const winningSignals = providerSignals.filter((s) => s.pnl && s.pnl > 0);
        const losingSignals = providerSignals.filter((s) => s.pnl && s.pnl <= 0);

        const winRate = providerSignals.length > 0
          ? (winningSignals.length / providerSignals.length) * 100
          : 0;

        const totalPnL = providerSignals.reduce((sum, s) => sum + (s.pnl || 0), 0);
        const avgPnLPerSignal = providerSignals.length > 0 ? totalPnL / providerSignals.length : 0;

        const bestSignal = Math.max(...providerSignals.map((s) => s.pnl || 0), 0);
        const worstSignal = Math.min(...providerSignals.map((s) => s.pnl || 0), 0);

        const grossProfit = winningSignals.reduce((sum, s) => sum + (s.pnl || 0), 0);
        const grossLoss = Math.abs(losingSignals.reduce((sum, s) => sum + (s.pnl || 0), 0));
        const profitFactor = grossLoss > 0 ? grossProfit / grossLoss : grossProfit > 0 ? 99 : 0;

        return {
          provider,
          winRate: Math.round(winRate * 10) / 10,
          totalSignals: providerSignals.length,
          avgPnLPerSignal: Math.round(avgPnLPerSignal * 100) / 100,
          bestSignal: Math.round(bestSignal * 100) / 100,
          worstSignal: Math.round(worstSignal * 100) / 100,
          profitFactor: Math.round(profitFactor * 100) / 100,
        };
      },

      refreshSignals: () => {
        set({
          signals: generateMockSignals(),
          historicalSignals: generateHistoricalSignals(),
        });
      },
    }),
    { name: 'signal-store' }
  )
);
