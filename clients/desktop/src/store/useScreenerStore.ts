/**
 * Symbol Screener Store with Zustand
 * Manages screener filters, results, and presets
 */

import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

export type SymbolGroup = 'All' | 'Forex' | 'Metals' | 'Crypto' | 'Indices' | 'Commodities';
export type Signal = 'BUY' | 'SELL' | 'NEUTRAL';
export type SortField = 'symbol' | 'changePercent' | 'spread' | 'volume';
export type SortDirection = 'asc' | 'desc';
export type MAFilter = 'none' | 'above20' | 'below20' | 'above50' | 'below50' | 'above200' | 'below200';

export interface SymbolData {
  symbol: string;
  group: Exclude<SymbolGroup, 'All'>;
  lastPrice: number;
  changeAbs: number;
  changePercent: number;
  spread: number;
  volume: number;
  rsi14: number;
  ma20: number;
  ma50: number;
  ma200: number;
  signal: Signal;
  sparklineData: number[]; // Last 50 bars for mini chart
}

export interface FilterCriteria {
  symbolGroup: SymbolGroup;
  priceChangeMin: number;
  priceChangeMax: number;
  spreadMin: number;
  spreadMax: number;
  volumeMin: number;
  volumeMax: number;
  rsiMin: number;
  rsiMax: number;
  maFilter: MAFilter;
  sortBy: SortField;
  sortDirection: SortDirection;
}

export interface ScannerPreset {
  name: string;
  filters: FilterCriteria;
}

interface ScreenerState {
  allSymbols: SymbolData[];
  filteredSymbols: SymbolData[];
  filters: FilterCriteria;
  presets: ScannerPreset[];
  isFilterPanelCollapsed: boolean;

  setFilters: (filters: Partial<FilterCriteria>) => void;
  resetFilters: () => void;
  applyFilters: () => void;
  toggleFilterPanel: () => void;
  savePreset: (name: string) => void;
  loadPreset: (preset: ScannerPreset) => void;
  deletePreset: (name: string) => void;
  applyQuickFilter: (type: 'oversold' | 'overbought' | 'highVolume' | 'tightSpread' | 'trendingUp' | 'trendingDown') => void;
}

// Default filter criteria
const defaultFilters: FilterCriteria = {
  symbolGroup: 'All',
  priceChangeMin: -100,
  priceChangeMax: 100,
  spreadMin: 0,
  spreadMax: 100,
  volumeMin: 0,
  volumeMax: 1000000000,
  rsiMin: 0,
  rsiMax: 100,
  maFilter: 'none',
  sortBy: 'changePercent',
  sortDirection: 'desc',
};

// Generate mock symbol data
function generateMockSymbols(): SymbolData[] {
  const forexPairs = [
    'EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD', 'NZDUSD', 'USDCHF',
    'EURGBP', 'EURJPY', 'GBPJPY', 'AUDJPY', 'EURAUD', 'EURCHF', 'GBPAUD',
    'GBPCAD', 'EURCAD', 'AUDCAD', 'AUDNZD', 'NZDJPY', 'CADJPY'
  ];
  const metals = ['XAUUSD', 'XAGUSD', 'XPTUSD', 'XPDUSD', 'XCUUSD'];
  const cryptos = ['BTCUSD', 'ETHUSD', 'XRPUSD', 'LTCUSD', 'ADAUSD', 'SOLUSD', 'DOTUSD'];
  const indices = ['SPX500', 'US30', 'NAS100', 'UK100', 'DE30', 'JP225', 'AU200'];
  const commodities = ['WTICOUSD', 'BCOUSD', 'NATGASUSD', 'CORNUSD', 'WHEATUSD', 'SOYBNUSD'];

  const symbols: SymbolData[] = [];

  const addSymbols = (list: string[], group: Exclude<SymbolGroup, 'All'>, basePrice: number, priceRange: number) => {
    list.forEach((symbol) => {
      const lastPrice = basePrice + (Math.random() - 0.5) * priceRange;
      const changePercent = -5 + Math.random() * 10; // -5% to +5%
      const changeAbs = lastPrice * (changePercent / 100);
      const spread = group === 'Forex' ? 0.5 + Math.random() * 2 :
                     group === 'Crypto' ? 5 + Math.random() * 20 :
                     1 + Math.random() * 5;
      const volume = Math.floor(100000 + Math.random() * 900000);
      const rsi14 = 20 + Math.random() * 60; // 20-80
      const ma20 = lastPrice * (0.97 + Math.random() * 0.06);
      const ma50 = lastPrice * (0.95 + Math.random() * 0.10);
      const ma200 = lastPrice * (0.90 + Math.random() * 0.20);

      // Determine signal based on RSI and MA
      let signal: Signal = 'NEUTRAL';
      if (rsi14 < 35 && lastPrice > ma20) signal = 'BUY';
      else if (rsi14 > 65 && lastPrice < ma20) signal = 'SELL';
      else if (lastPrice > ma20 && lastPrice > ma50 && changePercent > 0.5) signal = 'BUY';
      else if (lastPrice < ma20 && lastPrice < ma50 && changePercent < -0.5) signal = 'SELL';

      // Generate sparkline data (last 50 bars)
      const sparklineData: number[] = [];
      let currentPrice = lastPrice * 0.98;
      for (let i = 0; i < 50; i++) {
        currentPrice += (Math.random() - 0.5) * (lastPrice * 0.005);
        sparklineData.push(currentPrice);
      }

      symbols.push({
        symbol,
        group,
        lastPrice,
        changeAbs,
        changePercent,
        spread,
        volume,
        rsi14,
        ma20,
        ma50,
        ma200,
        signal,
        sparklineData,
      });
    });
  };

  addSymbols(forexPairs, 'Forex', 1.2, 0.5);
  addSymbols(metals, 'Metals', 2000, 500);
  addSymbols(cryptos, 'Crypto', 50000, 40000);
  addSymbols(indices, 'Indices', 15000, 10000);
  addSymbols(commodities, 'Commodities', 75, 50);

  return symbols;
}

// Filter symbols based on criteria
function filterSymbols(symbols: SymbolData[], filters: FilterCriteria): SymbolData[] {
  let filtered = symbols.filter((s) => {
    // Symbol group filter
    if (filters.symbolGroup !== 'All' && s.group !== filters.symbolGroup) return false;

    // Price change filter
    if (s.changePercent < filters.priceChangeMin || s.changePercent > filters.priceChangeMax) return false;

    // Spread filter
    if (s.spread < filters.spreadMin || s.spread > filters.spreadMax) return false;

    // Volume filter
    if (s.volume < filters.volumeMin || s.volume > filters.volumeMax) return false;

    // RSI filter
    if (s.rsi14 < filters.rsiMin || s.rsi14 > filters.rsiMax) return false;

    // MA filter
    switch (filters.maFilter) {
      case 'above20':
        if (s.lastPrice <= s.ma20) return false;
        break;
      case 'below20':
        if (s.lastPrice >= s.ma20) return false;
        break;
      case 'above50':
        if (s.lastPrice <= s.ma50) return false;
        break;
      case 'below50':
        if (s.lastPrice >= s.ma50) return false;
        break;
      case 'above200':
        if (s.lastPrice <= s.ma200) return false;
        break;
      case 'below200':
        if (s.lastPrice >= s.ma200) return false;
        break;
    }

    return true;
  });

  // Sort results
  filtered.sort((a, b) => {
    let aVal: number, bVal: number;
    switch (filters.sortBy) {
      case 'symbol':
        return filters.sortDirection === 'asc'
          ? a.symbol.localeCompare(b.symbol)
          : b.symbol.localeCompare(a.symbol);
      case 'changePercent':
        aVal = a.changePercent;
        bVal = b.changePercent;
        break;
      case 'spread':
        aVal = a.spread;
        bVal = b.spread;
        break;
      case 'volume':
        aVal = a.volume;
        bVal = b.volume;
        break;
      default:
        return 0;
    }
    return filters.sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
  });

  return filtered;
}

export const useScreenerStore = create<ScreenerState>()(
  devtools(
    persist(
      (set, get) => {
        const allSymbols = generateMockSymbols();
        const filteredSymbols = filterSymbols(allSymbols, defaultFilters);

        return {
          allSymbols,
          filteredSymbols,
          filters: defaultFilters,
          presets: [],
          isFilterPanelCollapsed: false,

          setFilters: (newFilters) => {
            const updatedFilters = { ...get().filters, ...newFilters };
            set({ filters: updatedFilters });
          },

          resetFilters: () => {
            set({
              filters: defaultFilters,
              filteredSymbols: filterSymbols(get().allSymbols, defaultFilters)
            });
          },

          applyFilters: () => {
            const { allSymbols, filters } = get();
            const filtered = filterSymbols(allSymbols, filters);
            set({ filteredSymbols: filtered });
          },

          toggleFilterPanel: () => {
            set({ isFilterPanelCollapsed: !get().isFilterPanelCollapsed });
          },

          savePreset: (name) => {
            const { filters, presets } = get();
            const existingIndex = presets.findIndex((p) => p.name === name);

            if (existingIndex >= 0) {
              const updated = [...presets];
              updated[existingIndex] = { name, filters: { ...filters } };
              set({ presets: updated });
            } else {
              set({ presets: [...presets, { name, filters: { ...filters } }] });
            }
          },

          loadPreset: (preset) => {
            set({ filters: { ...preset.filters } });
            get().applyFilters();
          },

          deletePreset: (name) => {
            set({ presets: get().presets.filter((p) => p.name !== name) });
          },

          applyQuickFilter: (type) => {
            const baseFilters = { ...defaultFilters };

            switch (type) {
              case 'oversold':
                baseFilters.rsiMin = 0;
                baseFilters.rsiMax = 30;
                break;
              case 'overbought':
                baseFilters.rsiMin = 70;
                baseFilters.rsiMax = 100;
                break;
              case 'highVolume':
                baseFilters.volumeMin = 500000;
                break;
              case 'tightSpread':
                baseFilters.spreadMax = 2;
                break;
              case 'trendingUp':
                baseFilters.priceChangeMin = 1;
                baseFilters.maFilter = 'above20';
                break;
              case 'trendingDown':
                baseFilters.priceChangeMax = -1;
                baseFilters.maFilter = 'below20';
                break;
            }

            set({ filters: baseFilters });
            get().applyFilters();
          },
        };
      },
      {
        name: 'screener-storage',
        partialize: (state) => ({ presets: state.presets }),
      }
    ),
    { name: 'screener-store' }
  )
);
