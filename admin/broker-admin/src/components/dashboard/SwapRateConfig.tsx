'use client';

import React, { useState, useMemo } from 'react';
import {
  TrendingUp,
  TrendingDown,
  Edit2,
  Calendar,
  DollarSign,
  CheckSquare,
  Square,
  RefreshCw,
  Calculator,
  History,
  ArrowUpDown
} from 'lucide-react';

// Types
type SwapType = 'points' | 'percentage' | 'currency';
type TripleSwapDay = 'Monday' | 'Tuesday' | 'Wednesday' | 'Thursday' | 'Friday';

interface SwapSymbol {
  symbol: string;
  longSwap: number;
  shortSwap: number;
  swapType: SwapType;
  tripleSwapDay: TripleSwapDay;
  lastUpdated: number;
  swapFree: boolean;
}

interface SwapHistory {
  id: string;
  timestamp: number;
  symbol: string;
  fieldChanged: 'longSwap' | 'shortSwap' | 'swapType' | 'tripleSwapDay';
  oldValue: string;
  newValue: string;
  adminUser: string;
}

// Mock Data - 30+ symbols with realistic swap rates
const MOCK_SYMBOLS: SwapSymbol[] = [
  { symbol: 'EURUSD', longSwap: -0.52, shortSwap: 0.18, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'GBPUSD', longSwap: -0.64, shortSwap: 0.22, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'USDJPY', longSwap: 0.15, shortSwap: -0.48, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 3, swapFree: false },
  { symbol: 'AUDUSD', longSwap: -0.42, shortSwap: 0.14, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'USDCAD', longSwap: -0.38, shortSwap: 0.12, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 4, swapFree: false },
  { symbol: 'USDCHF', longSwap: 0.08, shortSwap: -0.32, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'NZDUSD', longSwap: -0.46, shortSwap: 0.16, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 5, swapFree: false },
  { symbol: 'EURGBP', longSwap: -0.28, shortSwap: 0.08, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'EURJPY', longSwap: -0.36, shortSwap: 0.12, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 3, swapFree: false },
  { symbol: 'GBPJPY', longSwap: -0.58, shortSwap: 0.20, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'AUDCAD', longSwap: -0.34, shortSwap: 0.10, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 4, swapFree: false },
  { symbol: 'AUDCHF', longSwap: -0.30, shortSwap: 0.09, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 6, swapFree: false },
  { symbol: 'AUDJPY', longSwap: -0.44, shortSwap: 0.15, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'CHFJPY', longSwap: -0.40, shortSwap: 0.13, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 3, swapFree: false },
  { symbol: 'CADCHF', longSwap: -0.26, shortSwap: 0.08, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'EURCAD', longSwap: -0.48, shortSwap: 0.16, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'EURCHF', longSwap: -0.24, shortSwap: 0.07, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 5, swapFree: false },
  { symbol: 'EURAUD', longSwap: -0.56, shortSwap: 0.19, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'GBPCHF', longSwap: -0.50, shortSwap: 0.17, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 4, swapFree: false },
  { symbol: 'GBPCAD', longSwap: -0.62, shortSwap: 0.21, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'GBPAUD', longSwap: -0.68, shortSwap: 0.23, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 3, swapFree: false },
  { symbol: 'NZDCAD', longSwap: -0.38, shortSwap: 0.12, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'NZDCHF', longSwap: -0.32, shortSwap: 0.10, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 5, swapFree: false },
  { symbol: 'NZDJPY', longSwap: -0.42, shortSwap: 0.14, swapType: 'points', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'XAUUSD', longSwap: -2.85, shortSwap: 0.95, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'XAGUSD', longSwap: -0.15, shortSwap: 0.05, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 3, swapFree: false },
  { symbol: 'BTCUSD', longSwap: -12.50, shortSwap: 4.20, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'ETHUSD', longSwap: -3.80, shortSwap: 1.25, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
  { symbol: 'WTICOUSD', longSwap: -0.45, shortSwap: 0.15, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 4, swapFree: false },
  { symbol: 'US30USD', longSwap: -1.20, shortSwap: 0.40, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 1, swapFree: false },
  { symbol: 'SPX500USD', longSwap: -0.95, shortSwap: 0.32, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 3, swapFree: false },
  { symbol: 'NAS100USD', longSwap: -1.35, shortSwap: 0.45, swapType: 'currency', tripleSwapDay: 'Wednesday', lastUpdated: Date.now() - 86400000 * 2, swapFree: false },
];

// Generate mock swap history (30 days)
const generateSwapHistory = (): SwapHistory[] => {
  const history: SwapHistory[] = [];
  const admins = ['admin@rtx5.com', 'manager@rtx5.com', 'supervisor@rtx5.com'];
  const symbols = MOCK_SYMBOLS.slice(0, 15).map(s => s.symbol);
  const fields: Array<'longSwap' | 'shortSwap' | 'swapType' | 'tripleSwapDay'> = ['longSwap', 'shortSwap', 'swapType', 'tripleSwapDay'];

  for (let i = 0; i < 50; i++) {
    const daysAgo = Math.floor(Math.random() * 30);
    const symbol = symbols[Math.floor(Math.random() * symbols.length)];
    const field = fields[Math.floor(Math.random() * fields.length)];

    let oldValue = '';
    let newValue = '';

    if (field === 'longSwap' || field === 'shortSwap') {
      const base = Math.random() * 0.5;
      oldValue = (base).toFixed(2);
      newValue = (base + (Math.random() - 0.5) * 0.2).toFixed(2);
    } else if (field === 'swapType') {
      const types: SwapType[] = ['points', 'percentage', 'currency'];
      oldValue = types[Math.floor(Math.random() * types.length)];
      newValue = types[Math.floor(Math.random() * types.length)];
    } else {
      const days: TripleSwapDay[] = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'];
      oldValue = days[Math.floor(Math.random() * days.length)];
      newValue = days[Math.floor(Math.random() * days.length)];
    }

    history.push({
      id: `history-${i}`,
      timestamp: Date.now() - daysAgo * 86400000,
      symbol,
      fieldChanged: field,
      oldValue,
      newValue,
      adminUser: admins[Math.floor(Math.random() * admins.length)]
    });
  }

  return history.sort((a, b) => b.timestamp - a.timestamp);
};

const MOCK_HISTORY = generateSwapHistory();

export default function SwapRateConfig() {
  const [symbols, setSymbols] = useState<SwapSymbol[]>(MOCK_SYMBOLS);
  const [selectedSymbol, setSelectedSymbol] = useState<SwapSymbol | null>(null);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedSymbols, setSelectedSymbols] = useState<Set<string>>(new Set());
  const [showBulkModal, setShowBulkModal] = useState(false);
  const [bulkAdjustment, setBulkAdjustment] = useState<number>(0);
  const [bulkField, setBulkField] = useState<'longSwap' | 'shortSwap'>('longSwap');
  const [previewSymbol, setPreviewSymbol] = useState<string>('EURUSD');
  const [previewLots, setPreviewLots] = useState<number>(1.0);
  const [sortField, setSortField] = useState<keyof SwapSymbol>('symbol');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');

  // Summary stats
  const stats = useMemo(() => {
    const withSwaps = symbols.filter(s => !s.swapFree).length;
    const avgLong = symbols.filter(s => !s.swapFree).reduce((sum, s) => sum + s.longSwap, 0) / withSwaps;
    const avgShort = symbols.filter(s => !s.swapFree).reduce((sum, s) => sum + s.shortSwap, 0) / withSwaps;

    // Mock 24h revenue calculation (assume random positions)
    const revenue = symbols.reduce((sum, s) => {
      if (s.swapFree) return sum;
      const longRev = Math.abs(s.longSwap) * Math.random() * 100;
      const shortRev = Math.abs(s.shortSwap) * Math.random() * 100;
      return sum + longRev + shortRev;
    }, 0);

    return {
      withSwaps,
      avgLong,
      avgShort,
      revenue
    };
  }, [symbols]);

  // Sorting
  const sortedSymbols = useMemo(() => {
    const sorted = [...symbols].sort((a, b) => {
      const aVal = a[sortField];
      const bVal = b[sortField];

      if (typeof aVal === 'number' && typeof bVal === 'number') {
        return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
      }

      const aStr = String(aVal);
      const bStr = String(bVal);
      return sortDirection === 'asc'
        ? aStr.localeCompare(bStr)
        : bStr.localeCompare(aStr);
    });

    return sorted;
  }, [symbols, sortField, sortDirection]);

  // Handle sort
  const handleSort = (field: keyof SwapSymbol) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('asc');
    }
  };

  // Toggle selection
  const toggleSelection = (symbol: string) => {
    const newSet = new Set(selectedSymbols);
    if (newSet.has(symbol)) {
      newSet.delete(symbol);
    } else {
      newSet.add(symbol);
    }
    setSelectedSymbols(newSet);
  };

  // Select all / none
  const toggleSelectAll = () => {
    if (selectedSymbols.size === symbols.length) {
      setSelectedSymbols(new Set());
    } else {
      setSelectedSymbols(new Set(symbols.map(s => s.symbol)));
    }
  };

  // Edit symbol
  const handleEditSymbol = (symbol: SwapSymbol) => {
    setSelectedSymbol({ ...symbol });
    setShowEditModal(true);
  };

  // Save edited symbol
  const handleSaveEdit = () => {
    if (!selectedSymbol) return;

    setSymbols(prev => prev.map(s =>
      s.symbol === selectedSymbol.symbol
        ? { ...selectedSymbol, lastUpdated: Date.now() }
        : s
    ));

    setShowEditModal(false);
    setSelectedSymbol(null);
  };

  // Bulk update
  const handleBulkUpdate = () => {
    setSymbols(prev => prev.map(s => {
      if (!selectedSymbols.has(s.symbol)) return s;

      return {
        ...s,
        [bulkField]: s[bulkField] + bulkAdjustment,
        lastUpdated: Date.now()
      };
    }));

    setShowBulkModal(false);
    setSelectedSymbols(new Set());
    setBulkAdjustment(0);
  };

  // Calculate swap preview
  const calculateSwapPreview = (side: 'long' | 'short') => {
    const symbol = symbols.find(s => s.symbol === previewSymbol);
    if (!symbol) return 0;

    const swapRate = side === 'long' ? symbol.longSwap : symbol.shortSwap;

    if (symbol.swapType === 'currency') {
      return swapRate * previewLots;
    } else if (symbol.swapType === 'points') {
      // For FX pairs, assume 1 lot = 100,000 units
      const pipValue = 10; // $10 per pip for standard lot
      return swapRate * previewLots * pipValue;
    } else {
      // Percentage
      const notionalValue = 100000 * previewLots;
      return (swapRate / 100) * notionalValue;
    }
  };

  // Format timestamp
  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  };

  const formatDateTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return date.toLocaleString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <div className="flex flex-col h-full bg-[#121316] text-zinc-300 p-6 overflow-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-zinc-100">Swap Rate Configuration</h1>
          <p className="text-sm text-zinc-500 mt-1">Configure swap/rollover rates per symbol</p>
        </div>
        <button
          onClick={() => setShowBulkModal(true)}
          disabled={selectedSymbols.size === 0}
          className={`px-4 py-2 rounded text-sm font-medium transition-colors ${
            selectedSymbols.size === 0
              ? 'bg-zinc-800 text-zinc-600 cursor-not-allowed'
              : 'bg-blue-600 hover:bg-blue-700 text-white'
          }`}
        >
          <RefreshCw className="inline mr-2" size={14} />
          Bulk Update ({selectedSymbols.size})
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-zinc-500 uppercase">Symbols with Swaps</span>
            <CheckSquare size={16} className="text-blue-400" />
          </div>
          <div className="text-2xl font-bold text-zinc-100">{stats.withSwaps}</div>
          <div className="text-xs text-zinc-500 mt-1">of {symbols.length} total</div>
        </div>

        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-zinc-500 uppercase">Avg Long Swap</span>
            <TrendingDown size={16} className="text-rose-400" />
          </div>
          <div className="text-2xl font-bold text-zinc-100">{stats.avgLong.toFixed(2)}</div>
          <div className="text-xs text-zinc-500 mt-1">points</div>
        </div>

        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-zinc-500 uppercase">Avg Short Swap</span>
            <TrendingUp size={16} className="text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-zinc-100">{stats.avgShort.toFixed(2)}</div>
          <div className="text-xs text-zinc-500 mt-1">points</div>
        </div>

        <div className="bg-[#1E2026] border border-zinc-700 rounded p-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs text-zinc-500 uppercase">Total Swap Revenue (24h)</span>
            <DollarSign size={16} className="text-emerald-400" />
          </div>
          <div className="text-2xl font-bold text-zinc-100">${stats.revenue.toFixed(0)}</div>
          <div className="text-xs text-zinc-500 mt-1">estimated</div>
        </div>
      </div>

      {/* Main Content - 3 Column Layout */}
      <div className="grid grid-cols-3 gap-6 flex-1">
        {/* Left Column: Symbol Table */}
        <div className="col-span-2 bg-[#1E2026] border border-zinc-700 rounded flex flex-col">
          <div className="px-4 py-3 border-b border-zinc-700 flex items-center justify-between">
            <h2 className="text-sm font-bold text-zinc-100 uppercase">Symbol Swap Rates</h2>
            <button
              onClick={toggleSelectAll}
              className="text-xs text-blue-400 hover:text-blue-300 transition-colors"
            >
              {selectedSymbols.size === symbols.length ? 'Deselect All' : 'Select All'}
            </button>
          </div>

          <div className="flex-1 overflow-auto">
            <table className="w-full text-xs">
              <thead className="sticky top-0 bg-[#1E2026] border-b border-zinc-700">
                <tr className="text-zinc-400">
                  <th className="text-left p-2 w-8">
                    <button onClick={toggleSelectAll}>
                      {selectedSymbols.size === symbols.length ? (
                        <CheckSquare size={14} className="text-blue-400" />
                      ) : (
                        <Square size={14} />
                      )}
                    </button>
                  </th>
                  <th className="text-left p-2 cursor-pointer hover:text-zinc-200" onClick={() => handleSort('symbol')}>
                    <div className="flex items-center">
                      Symbol
                      <ArrowUpDown size={12} className="ml-1" />
                    </div>
                  </th>
                  <th className="text-right p-2 cursor-pointer hover:text-zinc-200" onClick={() => handleSort('longSwap')}>
                    <div className="flex items-center justify-end">
                      Long Swap
                      <ArrowUpDown size={12} className="ml-1" />
                    </div>
                  </th>
                  <th className="text-right p-2 cursor-pointer hover:text-zinc-200" onClick={() => handleSort('shortSwap')}>
                    <div className="flex items-center justify-end">
                      Short Swap
                      <ArrowUpDown size={12} className="ml-1" />
                    </div>
                  </th>
                  <th className="text-left p-2">Type</th>
                  <th className="text-left p-2">Triple Day</th>
                  <th className="text-left p-2">Last Updated</th>
                  <th className="text-center p-2 w-16">Action</th>
                </tr>
              </thead>
              <tbody>
                {sortedSymbols.map((symbol) => (
                  <tr
                    key={symbol.symbol}
                    className="border-b border-zinc-800 hover:bg-zinc-900 transition-colors"
                  >
                    <td className="p-2">
                      <button onClick={() => toggleSelection(symbol.symbol)}>
                        {selectedSymbols.has(symbol.symbol) ? (
                          <CheckSquare size={14} className="text-blue-400" />
                        ) : (
                          <Square size={14} className="text-zinc-600" />
                        )}
                      </button>
                    </td>
                    <td className="p-2 font-medium text-zinc-100">
                      {symbol.symbol}
                      {symbol.swapFree && (
                        <span className="ml-2 text-[10px] bg-emerald-900/30 text-emerald-400 px-1.5 py-0.5 rounded">
                          SWAP-FREE
                        </span>
                      )}
                    </td>
                    <td className={`p-2 text-right font-mono ${symbol.longSwap < 0 ? 'text-rose-400' : 'text-emerald-400'}`}>
                      {symbol.longSwap.toFixed(2)}
                    </td>
                    <td className={`p-2 text-right font-mono ${symbol.shortSwap < 0 ? 'text-rose-400' : 'text-emerald-400'}`}>
                      {symbol.shortSwap.toFixed(2)}
                    </td>
                    <td className="p-2 text-zinc-400 capitalize">{symbol.swapType}</td>
                    <td className="p-2 text-zinc-400">{symbol.tripleSwapDay}</td>
                    <td className="p-2 text-zinc-500 text-[10px]">{formatDate(symbol.lastUpdated)}</td>
                    <td className="p-2 text-center">
                      <button
                        onClick={() => handleEditSymbol(symbol)}
                        className="text-blue-400 hover:text-blue-300 transition-colors"
                      >
                        <Edit2 size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Right Column: Calculator & History */}
        <div className="flex flex-col gap-6">
          {/* Swap Calculator */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded">
            <div className="px-4 py-3 border-b border-zinc-700">
              <h2 className="text-sm font-bold text-zinc-100 uppercase flex items-center">
                <Calculator size={16} className="mr-2 text-blue-400" />
                Swap Calculator
              </h2>
            </div>

            <div className="p-4 space-y-4">
              <div>
                <label className="block text-xs text-zinc-500 mb-1">Symbol</label>
                <select
                  value={previewSymbol}
                  onChange={(e) => setPreviewSymbol(e.target.value)}
                  className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                >
                  {symbols.map((s) => (
                    <option key={s.symbol} value={s.symbol}>{s.symbol}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs text-zinc-500 mb-1">Position Size (Lots)</label>
                <input
                  type="number"
                  value={previewLots}
                  onChange={(e) => setPreviewLots(parseFloat(e.target.value) || 0)}
                  step="0.01"
                  min="0.01"
                  className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-100"
                />
              </div>

              <div className="pt-2 border-t border-zinc-700 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-xs text-zinc-500">Long Swap (daily):</span>
                  <span className={`text-sm font-mono font-bold ${calculateSwapPreview('long') < 0 ? 'text-rose-400' : 'text-emerald-400'}`}>
                    ${calculateSwapPreview('long').toFixed(2)}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-zinc-500">Short Swap (daily):</span>
                  <span className={`text-sm font-mono font-bold ${calculateSwapPreview('short') < 0 ? 'text-rose-400' : 'text-emerald-400'}`}>
                    ${calculateSwapPreview('short').toFixed(2)}
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Swap History */}
          <div className="bg-[#1E2026] border border-zinc-700 rounded flex flex-col flex-1">
            <div className="px-4 py-3 border-b border-zinc-700">
              <h2 className="text-sm font-bold text-zinc-100 uppercase flex items-center">
                <History size={16} className="mr-2 text-amber-400" />
                Recent Changes (30 days)
              </h2>
            </div>

            <div className="flex-1 overflow-auto">
              <table className="w-full text-xs">
                <thead className="sticky top-0 bg-[#1E2026] border-b border-zinc-700">
                  <tr className="text-zinc-400">
                    <th className="text-left p-2">Date</th>
                    <th className="text-left p-2">Symbol</th>
                    <th className="text-left p-2">Field</th>
                    <th className="text-left p-2">Change</th>
                  </tr>
                </thead>
                <tbody>
                  {MOCK_HISTORY.slice(0, 20).map((entry) => (
                    <tr key={entry.id} className="border-b border-zinc-800 hover:bg-zinc-900">
                      <td className="p-2 text-zinc-500 text-[10px]">
                        {formatDateTime(entry.timestamp)}
                      </td>
                      <td className="p-2 font-medium text-zinc-100">{entry.symbol}</td>
                      <td className="p-2 text-zinc-400 capitalize">
                        {entry.fieldChanged.replace(/([A-Z])/g, ' $1').trim()}
                      </td>
                      <td className="p-2">
                        <div className="flex items-center text-[10px]">
                          <span className="text-zinc-500">{entry.oldValue}</span>
                          <span className="mx-1 text-zinc-600">→</span>
                          <span className="text-blue-400">{entry.newValue}</span>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      {/* Edit Modal */}
      {showEditModal && selectedSymbol && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-full max-w-2xl">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between">
              <h2 className="text-lg font-bold text-zinc-100">
                Edit Swap Rates - {selectedSymbol.symbol}
              </h2>
              <button
                onClick={() => setShowEditModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                ✕
              </button>
            </div>

            <div className="p-6 space-y-6">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Long Swap Rate</label>
                  <input
                    type="number"
                    value={selectedSymbol.longSwap}
                    onChange={(e) => setSelectedSymbol({ ...selectedSymbol, longSwap: parseFloat(e.target.value) || 0 })}
                    step="0.01"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Short Swap Rate</label>
                  <input
                    type="number"
                    value={selectedSymbol.shortSwap}
                    onChange={(e) => setSelectedSymbol({ ...selectedSymbol, shortSwap: parseFloat(e.target.value) || 0 })}
                    step="0.01"
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Swap Type</label>
                  <select
                    value={selectedSymbol.swapType}
                    onChange={(e) => setSelectedSymbol({ ...selectedSymbol, swapType: e.target.value as SwapType })}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  >
                    <option value="points">Points</option>
                    <option value="percentage">Percentage</option>
                    <option value="currency">Currency</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm text-zinc-400 mb-2">Triple Swap Day</label>
                  <select
                    value={selectedSymbol.tripleSwapDay}
                    onChange={(e) => setSelectedSymbol({ ...selectedSymbol, tripleSwapDay: e.target.value as TripleSwapDay })}
                    className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                  >
                    <option value="Monday">Monday</option>
                    <option value="Tuesday">Tuesday</option>
                    <option value="Wednesday">Wednesday</option>
                    <option value="Thursday">Thursday</option>
                    <option value="Friday">Friday</option>
                  </select>
                </div>
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  id="swapFree"
                  checked={selectedSymbol.swapFree}
                  onChange={(e) => setSelectedSymbol({ ...selectedSymbol, swapFree: e.target.checked })}
                  className="mr-2"
                />
                <label htmlFor="swapFree" className="text-sm text-zinc-300">
                  Swap-Free (Islamic Accounts Override)
                </label>
              </div>
            </div>

            <div className="px-6 py-4 border-t border-zinc-700 flex justify-end gap-3">
              <button
                onClick={() => setShowEditModal(false)}
                className="px-4 py-2 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSaveEdit}
                className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm transition-colors"
              >
                Save Changes
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Bulk Update Modal */}
      {showBulkModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#1E2026] border border-zinc-700 rounded w-full max-w-md">
            <div className="px-6 py-4 border-b border-zinc-700 flex items-center justify-between">
              <h2 className="text-lg font-bold text-zinc-100">
                Bulk Swap Adjustment
              </h2>
              <button
                onClick={() => setShowBulkModal(false)}
                className="text-zinc-500 hover:text-zinc-300"
              >
                ✕
              </button>
            </div>

            <div className="p-6 space-y-4">
              <div className="bg-zinc-900 border border-zinc-700 rounded p-3">
                <div className="text-xs text-zinc-500 mb-1">Selected Symbols:</div>
                <div className="text-sm text-zinc-100 font-medium">{selectedSymbols.size} symbols</div>
              </div>

              <div>
                <label className="block text-sm text-zinc-400 mb-2">Field to Adjust</label>
                <select
                  value={bulkField}
                  onChange={(e) => setBulkField(e.target.value as 'longSwap' | 'shortSwap')}
                  className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                >
                  <option value="longSwap">Long Swap</option>
                  <option value="shortSwap">Short Swap</option>
                </select>
              </div>

              <div>
                <label className="block text-sm text-zinc-400 mb-2">Adjustment (add/subtract)</label>
                <input
                  type="number"
                  value={bulkAdjustment}
                  onChange={(e) => setBulkAdjustment(parseFloat(e.target.value) || 0)}
                  step="0.01"
                  placeholder="e.g., +0.10 or -0.05"
                  className="w-full bg-zinc-900 border border-zinc-700 rounded px-3 py-2 text-zinc-100"
                />
                <div className="text-xs text-zinc-500 mt-1">
                  Use positive values to increase, negative to decrease
                </div>
              </div>
            </div>

            <div className="px-6 py-4 border-t border-zinc-700 flex justify-end gap-3">
              <button
                onClick={() => setShowBulkModal(false)}
                className="px-4 py-2 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleBulkUpdate}
                className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm transition-colors"
              >
                Apply to {selectedSymbols.size} Symbols
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
