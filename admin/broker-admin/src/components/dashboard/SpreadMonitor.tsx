'use client';

import React, { useState, useMemo } from 'react';
import { X, Search, TrendingUp, TrendingDown, AlertTriangle, Bell, BellOff, Settings, BarChart3 } from 'lucide-react';

type SymbolGroup = 'all' | 'forex_majors' | 'forex_minors' | 'forex_exotics' | 'metals' | 'crypto' | 'indices';
type SpreadStatus = 'normal' | 'wide' | 'alert';

interface SpreadData {
  symbol: string;
  group: SymbolGroup;
  bid: number;
  ask: number;
  currentSpread: number; // in pips
  avgSpread: number;
  minSpread: number;
  maxSpread: number;
  markup: number; // broker markup in pips
  status: SpreadStatus;
  alertThreshold?: number;
  alertActive: boolean;
}

interface SpreadHistoryPoint {
  timestamp: string;
  spread: number;
}

interface LPSpreadData {
  lpName: string;
  spread: number;
  color: string;
}

export default function SpreadMonitor() {
  const [selectedGroup, setSelectedGroup] = useState<SymbolGroup>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedSymbol, setSelectedSymbol] = useState<string>('EURUSD');
  const [showAlertModal, setShowAlertModal] = useState(false);
  const [showComparisonModal, setShowComparisonModal] = useState(false);
  const [selectedForComparison, setSelectedForComparison] = useState<string[]>(['EURUSD']);
  const [alertSymbol, setAlertSymbol] = useState<string>('');
  const [alertThreshold, setAlertThreshold] = useState<number>(0);

  // Mock spread data for 30+ symbols
  const spreadData = useMemo<SpreadData[]>(() => [
    // Forex Majors
    { symbol: 'EURUSD', group: 'forex_majors', bid: 1.08245, ask: 1.08255, currentSpread: 1.0, avgSpread: 1.2, minSpread: 0.8, maxSpread: 2.5, markup: 0.3, status: 'normal', alertThreshold: 2.0, alertActive: true },
    { symbol: 'GBPUSD', group: 'forex_majors', bid: 1.26432, ask: 1.26447, currentSpread: 1.5, avgSpread: 1.5, minSpread: 1.0, maxSpread: 3.0, markup: 0.4, status: 'normal', alertActive: false },
    { symbol: 'USDJPY', group: 'forex_majors', bid: 148.235, ask: 148.250, currentSpread: 1.5, avgSpread: 1.3, minSpread: 1.0, maxSpread: 2.8, markup: 0.3, status: 'wide', alertActive: false },
    { symbol: 'USDCHF', group: 'forex_majors', bid: 0.87634, ask: 0.87646, currentSpread: 1.2, avgSpread: 1.4, minSpread: 0.9, maxSpread: 2.6, markup: 0.3, status: 'normal', alertActive: false },
    { symbol: 'AUDUSD', group: 'forex_majors', bid: 0.65234, ask: 0.65249, currentSpread: 1.5, avgSpread: 1.6, minSpread: 1.1, maxSpread: 3.2, markup: 0.4, status: 'normal', alertActive: false },
    { symbol: 'USDCAD', group: 'forex_majors', bid: 1.38456, ask: 1.38471, currentSpread: 1.5, avgSpread: 1.5, minSpread: 1.0, maxSpread: 2.9, markup: 0.4, status: 'normal', alertActive: false },
    { symbol: 'NZDUSD', group: 'forex_majors', bid: 0.59123, ask: 0.59141, currentSpread: 1.8, avgSpread: 1.9, minSpread: 1.3, maxSpread: 3.5, markup: 0.5, status: 'normal', alertActive: false },

    // Forex Minors
    { symbol: 'EURGBP', group: 'forex_minors', bid: 0.85632, ask: 0.85650, currentSpread: 1.8, avgSpread: 2.0, minSpread: 1.4, maxSpread: 3.8, markup: 0.5, status: 'normal', alertActive: false },
    { symbol: 'EURJPY', group: 'forex_minors', bid: 160.456, ask: 160.478, currentSpread: 2.2, avgSpread: 2.1, minSpread: 1.6, maxSpread: 4.2, markup: 0.6, status: 'wide', alertActive: false },
    { symbol: 'GBPJPY', group: 'forex_minors', bid: 187.234, ask: 187.262, currentSpread: 2.8, avgSpread: 2.5, minSpread: 1.9, maxSpread: 5.0, markup: 0.7, status: 'alert', alertThreshold: 2.6, alertActive: true },
    { symbol: 'EURCHF', group: 'forex_minors', bid: 0.94567, ask: 0.94587, currentSpread: 2.0, avgSpread: 2.2, minSpread: 1.5, maxSpread: 4.0, markup: 0.6, status: 'normal', alertActive: false },
    { symbol: 'EURAUD', group: 'forex_minors', bid: 1.65934, ask: 1.65962, currentSpread: 2.8, avgSpread: 2.9, minSpread: 2.0, maxSpread: 5.5, markup: 0.8, status: 'normal', alertActive: false },
    { symbol: 'GBPAUD', group: 'forex_minors', bid: 1.93745, ask: 1.93780, currentSpread: 3.5, avgSpread: 3.4, minSpread: 2.5, maxSpread: 6.5, markup: 1.0, status: 'wide', alertActive: false },
    { symbol: 'AUDCAD', group: 'forex_minors', bid: 0.90234, ask: 0.90262, currentSpread: 2.8, avgSpread: 2.7, minSpread: 2.0, maxSpread: 5.2, markup: 0.8, status: 'wide', alertActive: false },
    { symbol: 'AUDJPY', group: 'forex_minors', bid: 96.734, ask: 96.760, currentSpread: 2.6, avgSpread: 2.5, minSpread: 1.9, maxSpread: 4.8, markup: 0.7, status: 'wide', alertActive: false },

    // Forex Exotics
    { symbol: 'USDTRY', group: 'forex_exotics', bid: 33.4567, ask: 33.5234, currentSpread: 66.7, avgSpread: 65.0, minSpread: 45.0, maxSpread: 120.0, markup: 15.0, status: 'wide', alertActive: false },
    { symbol: 'USDZAR', group: 'forex_exotics', bid: 18.2345, ask: 18.2845, currentSpread: 50.0, avgSpread: 48.0, minSpread: 35.0, maxSpread: 85.0, markup: 12.0, status: 'wide', alertActive: false },
    { symbol: 'USDMXN', group: 'forex_exotics', bid: 19.8734, ask: 19.9234, currentSpread: 50.0, avgSpread: 52.0, minSpread: 38.0, maxSpread: 95.0, markup: 12.0, status: 'normal', alertActive: false },
    { symbol: 'USDSEK', group: 'forex_exotics', bid: 10.6234, ask: 10.6534, currentSpread: 30.0, avgSpread: 28.0, minSpread: 20.0, maxSpread: 55.0, markup: 8.0, status: 'wide', alertActive: false },
    { symbol: 'USDNOK', group: 'forex_exotics', bid: 10.8456, ask: 10.8756, currentSpread: 30.0, avgSpread: 29.0, minSpread: 21.0, maxSpread: 58.0, markup: 8.0, status: 'wide', alertActive: false },

    // Metals
    { symbol: 'XAUUSD', group: 'metals', bid: 2034.56, ask: 2034.86, currentSpread: 30.0, avgSpread: 28.0, minSpread: 20.0, maxSpread: 50.0, markup: 8.0, status: 'wide', alertThreshold: 32.0, alertActive: true },
    { symbol: 'XAGUSD', group: 'metals', bid: 23.456, ask: 23.486, currentSpread: 30.0, avgSpread: 32.0, minSpread: 22.0, maxSpread: 60.0, markup: 8.0, status: 'normal', alertActive: false },
    { symbol: 'XPTUSD', group: 'metals', bid: 912.34, ask: 913.34, currentSpread: 100.0, avgSpread: 95.0, minSpread: 70.0, maxSpread: 150.0, markup: 25.0, status: 'wide', alertActive: false },
    { symbol: 'XPDUSD', group: 'metals', bid: 956.78, ask: 957.78, currentSpread: 100.0, avgSpread: 98.0, minSpread: 75.0, maxSpread: 145.0, markup: 25.0, status: 'wide', alertActive: false },

    // Crypto
    { symbol: 'BTCUSD', group: 'crypto', bid: 43256.78, ask: 43276.45, currentSpread: 19.67, avgSpread: 18.5, minSpread: 12.0, maxSpread: 35.0, markup: 5.0, status: 'wide', alertActive: false },
    { symbol: 'ETHUSD', group: 'crypto', bid: 2234.56, ask: 2236.23, currentSpread: 16.7, avgSpread: 15.2, minSpread: 10.0, maxSpread: 28.0, markup: 4.0, status: 'wide', alertActive: false },
    { symbol: 'XRPUSD', group: 'crypto', bid: 0.5234, ask: 0.5256, currentSpread: 22.0, avgSpread: 20.0, minSpread: 15.0, maxSpread: 40.0, markup: 6.0, status: 'wide', alertActive: false },
    { symbol: 'SOLUSD', group: 'crypto', bid: 98.456, ask: 98.678, currentSpread: 22.2, avgSpread: 21.0, minSpread: 16.0, maxSpread: 38.0, markup: 6.0, status: 'wide', alertActive: false },

    // Indices
    { symbol: 'SPX500', group: 'indices', bid: 4823.45, ask: 4824.25, currentSpread: 80.0, avgSpread: 75.0, minSpread: 50.0, maxSpread: 120.0, markup: 20.0, status: 'wide', alertActive: false },
    { symbol: 'NAS100', group: 'indices', bid: 16745.67, ask: 16747.23, currentSpread: 156.0, avgSpread: 150.0, minSpread: 100.0, maxSpread: 250.0, markup: 40.0, status: 'wide', alertActive: false },
    { symbol: 'US30', group: 'indices', bid: 37456.78, ask: 37459.34, currentSpread: 256.0, avgSpread: 245.0, minSpread: 180.0, maxSpread: 380.0, markup: 60.0, status: 'wide', alertActive: false },
    { symbol: 'GER40', group: 'indices', bid: 16934.56, ask: 16936.23, currentSpread: 167.0, avgSpread: 160.0, minSpread: 110.0, maxSpread: 270.0, markup: 45.0, status: 'wide', alertActive: false },
    { symbol: 'UK100', group: 'indices', bid: 7645.23, ask: 7646.45, currentSpread: 122.0, avgSpread: 118.0, minSpread: 80.0, maxSpread: 200.0, markup: 35.0, status: 'wide', alertActive: false },
    { symbol: 'JP225', group: 'indices', bid: 36234.56, ask: 36237.89, currentSpread: 333.0, avgSpread: 320.0, minSpread: 220.0, maxSpread: 500.0, markup: 80.0, status: 'wide', alertActive: false },
  ], []);

  // Generate 24h spread history (144 points, 10-minute intervals)
  const generateSpreadHistory = (symbol: string): SpreadHistoryPoint[] => {
    const data = spreadData.find(s => s.symbol === symbol);
    if (!data) return [];

    const points: SpreadHistoryPoint[] = [];
    const now = Date.now();
    const baseSpread = data.avgSpread;

    for (let i = 143; i >= 0; i--) {
      const timestamp = new Date(now - i * 10 * 60 * 1000).toISOString();
      // Simulate realistic spread variation with wider spreads during low liquidity hours
      const hour = new Date(now - i * 10 * 60 * 1000).getHours();
      const isLowLiquidity = hour >= 22 || hour <= 6; // 10 PM - 6 AM
      const volatility = isLowLiquidity ? 0.4 : 0.2;
      const spread = Math.max(data.minSpread, baseSpread + (Math.random() - 0.5) * baseSpread * volatility);
      points.push({ timestamp, spread });
    }

    return points;
  };

  // Generate hourly heatmap data (24 hours)
  const generateHeatmapData = (symbol: string): number[] => {
    const data = spreadData.find(s => s.symbol === symbol);
    if (!data) return Array(24).fill(0);

    const baseSpread = data.avgSpread;
    return Array(24).fill(0).map((_, hour) => {
      const isLowLiquidity = hour >= 22 || hour <= 6;
      const multiplier = isLowLiquidity ? 1.3 : 1.0;
      return baseSpread * multiplier * (0.9 + Math.random() * 0.2);
    });
  };

  // Mock LP spread comparison data
  const getLPSpreads = (symbol: string): LPSpreadData[] => {
    const data = spreadData.find(s => s.symbol === symbol);
    if (!data) return [];

    const rawSpread = data.currentSpread - data.markup;
    return [
      { lpName: 'YOFX', spread: rawSpread, color: '#3B82F6' },
      { lpName: 'LP-PRIME', spread: rawSpread + 0.1, color: '#10B981' },
      { lpName: 'LMAX', spread: rawSpread + 0.2, color: '#F59E0B' },
      { lpName: 'B2Broker', spread: rawSpread - 0.1, color: '#8B5CF6' },
    ];
  };

  const filteredData = useMemo(() => {
    let data = spreadData;

    if (selectedGroup !== 'all') {
      data = data.filter(d => d.group === selectedGroup);
    }

    if (searchQuery) {
      data = data.filter(d => d.symbol.toLowerCase().includes(searchQuery.toLowerCase()));
    }

    return data;
  }, [spreadData, selectedGroup, searchQuery]);

  const stats = useMemo(() => {
    const avgSpread = filteredData.reduce((sum, d) => sum + d.currentSpread, 0) / filteredData.length;
    const widest = filteredData.reduce((max, d) => d.currentSpread > max.currentSpread ? d : max, filteredData[0]);
    const narrowest = filteredData.reduce((min, d) => d.currentSpread < min.currentSpread ? d : min, filteredData[0]);
    const alertsActive = filteredData.filter(d => d.alertActive).length;

    return { avgSpread, widest, narrowest, alertsActive };
  }, [filteredData]);

  const getStatusColor = (status: SpreadStatus) => {
    switch (status) {
      case 'normal': return 'text-[#10B981]';
      case 'wide': return 'text-[#F59E0B]';
      case 'alert': return 'text-[#EF4444]';
    }
  };

  const getStatusBg = (status: SpreadStatus) => {
    switch (status) {
      case 'normal': return 'bg-[#10B981]/10 border-[#10B981]/20';
      case 'wide': return 'bg-[#F59E0B]/10 border-[#F59E0B]/20';
      case 'alert': return 'bg-[#EF4444]/10 border-[#EF4444]/20';
    }
  };

  const handleSetAlert = () => {
    if (!alertSymbol || alertThreshold <= 0) return;
    // In real implementation, this would call API to set alert
    console.log(`Setting alert for ${alertSymbol} at ${alertThreshold} pips`);
    setShowAlertModal(false);
    setAlertSymbol('');
    setAlertThreshold(0);
  };

  const toggleComparison = (symbol: string) => {
    setSelectedForComparison(prev =>
      prev.includes(symbol)
        ? prev.filter(s => s !== symbol)
        : [...prev, symbol]
    );
  };

  const spreadHistory = useMemo(() => generateSpreadHistory(selectedSymbol), [selectedSymbol]);
  const heatmapData = useMemo(() => generateHeatmapData(selectedSymbol), [selectedSymbol]);
  const lpSpreads = useMemo(() => getLPSpreads(selectedSymbol), [selectedSymbol]);

  return (
    <div className="h-full flex flex-col bg-[#18181b] text-[#E5E5E5] overflow-hidden">
      {/* Header */}
      <div className="flex-shrink-0 px-6 py-4 border-b border-[#3f3f46]">
        <div className="flex items-center justify-between">
          <h1 className="text-xl font-semibold">Spread Monitor</h1>
          <button
            onClick={() => setShowAlertModal(true)}
            className="px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-sm rounded flex items-center gap-2 transition-colors"
          >
            <Bell size={14} />
            Configure Alerts
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="flex-shrink-0 px-6 py-4 grid grid-cols-4 gap-4">
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Avg Spread (All Symbols)</div>
          <div className="text-2xl font-bold text-[#3B82F6]">{stats.avgSpread.toFixed(2)} pips</div>
        </div>
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Widest Spread Now</div>
          <div className="text-2xl font-bold text-[#EF4444]">{stats.widest?.currentSpread.toFixed(2)} pips</div>
          <div className="text-xs text-[#666] mt-1">{stats.widest?.symbol}</div>
        </div>
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Narrowest Spread Now</div>
          <div className="text-2xl font-bold text-[#10B981]">{stats.narrowest?.currentSpread.toFixed(2)} pips</div>
          <div className="text-xs text-[#666] mt-1">{stats.narrowest?.symbol}</div>
        </div>
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
          <div className="text-xs text-[#888] mb-1">Spread Alerts Active</div>
          <div className="text-2xl font-bold text-[#F59E0B]">{stats.alertsActive}</div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex-shrink-0 px-6 py-3 border-b border-[#3f3f46] flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#666]" />
          <input
            type="text"
            placeholder="Search symbols..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-[#27272a] border border-[#3f3f46] rounded pl-9 pr-3 py-1.5 text-xs focus:outline-none focus:border-[#3B82F6]"
          />
        </div>
        <div className="flex gap-2">
          {(['all', 'forex_majors', 'forex_minors', 'forex_exotics', 'metals', 'crypto', 'indices'] as SymbolGroup[]).map(group => (
            <button
              key={group}
              onClick={() => setSelectedGroup(group)}
              className={`px-3 py-1.5 text-xs rounded transition-colors ${
                selectedGroup === group
                  ? 'bg-[#3B82F6] text-white'
                  : 'bg-[#27272a] text-[#888] hover:bg-[#3f3f46] border border-[#3f3f46]'
              }`}
            >
              {group === 'all' ? 'All' : group.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}
            </button>
          ))}
        </div>
        <button
          onClick={() => setShowComparisonModal(true)}
          className="px-3 py-1.5 text-xs bg-[#27272a] border border-[#3f3f46] rounded hover:bg-[#3f3f46] transition-colors flex items-center gap-2"
        >
          <BarChart3 size={14} />
          Compare Symbols
        </button>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto px-6 py-4">
        {/* Spread Table */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded overflow-hidden mb-6">
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-[#1e1e22] border-b border-[#3f3f46]">
                <tr>
                  <th className="px-3 py-2 text-left font-semibold text-[#888]">Symbol</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Bid</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Ask</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Current (pips)</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Avg (pips)</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Min (pips)</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Max (pips)</th>
                  <th className="px-3 py-2 text-right font-semibold text-[#888]">Markup (pips)</th>
                  <th className="px-3 py-2 text-center font-semibold text-[#888]">Status</th>
                  <th className="px-3 py-2 text-center font-semibold text-[#888]">Alert</th>
                </tr>
              </thead>
              <tbody>
                {filteredData.map((item, idx) => (
                  <tr
                    key={item.symbol}
                    onClick={() => setSelectedSymbol(item.symbol)}
                    className={`border-b border-[#3f3f46] hover:bg-[#2a2a2e] cursor-pointer transition-colors ${
                      selectedSymbol === item.symbol ? 'bg-[#2a2a2e]' : ''
                    }`}
                  >
                    <td className="px-3 py-2 font-semibold text-[#E5E5E5]">{item.symbol}</td>
                    <td className="px-3 py-2 text-right text-[#888]">{item.bid.toFixed(item.group === 'crypto' ? 2 : 5)}</td>
                    <td className="px-3 py-2 text-right text-[#888]">{item.ask.toFixed(item.group === 'crypto' ? 2 : 5)}</td>
                    <td className={`px-3 py-2 text-right font-semibold ${
                      item.currentSpread < item.avgSpread ? 'text-[#10B981]' :
                      item.currentSpread > item.avgSpread * 1.2 ? 'text-[#EF4444]' : 'text-[#F59E0B]'
                    }`}>
                      {item.currentSpread.toFixed(1)}
                    </td>
                    <td className="px-3 py-2 text-right text-[#888]">{item.avgSpread.toFixed(1)}</td>
                    <td className="px-3 py-2 text-right text-[#888]">{item.minSpread.toFixed(1)}</td>
                    <td className="px-3 py-2 text-right text-[#888]">{item.maxSpread.toFixed(1)}</td>
                    <td className="px-3 py-2 text-right text-[#888]">{item.markup.toFixed(1)}</td>
                    <td className="px-3 py-2 text-center">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs border ${getStatusBg(item.status)} ${getStatusColor(item.status)}`}>
                        {item.status.charAt(0).toUpperCase() + item.status.slice(1)}
                      </span>
                    </td>
                    <td className="px-3 py-2 text-center">
                      {item.alertActive ? (
                        <Bell size={14} className="inline text-[#F59E0B]" />
                      ) : (
                        <BellOff size={14} className="inline text-[#666]" />
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Spread History Chart */}
        <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4 mb-6">
          <h3 className="text-sm font-semibold mb-3">Spread History (24h) - {selectedSymbol}</h3>
          <div className="w-full h-64">
            <svg viewBox="0 0 800 200" className="w-full h-full">
              {/* Grid lines */}
              {[0, 1, 2, 3, 4].map(i => (
                <line key={i} x1="40" y1={i * 50} x2="780" y2={i * 50} stroke="#3f3f46" strokeWidth="0.5" />
              ))}
              {/* Y-axis labels */}
              {spreadHistory.length > 0 && (() => {
                const maxSpread = Math.max(...spreadHistory.map(p => p.spread));
                const minSpread = Math.min(...spreadHistory.map(p => p.spread));
                const range = maxSpread - minSpread;
                return [0, 1, 2, 3, 4].map(i => {
                  const value = maxSpread - (i * range / 4);
                  return (
                    <text key={i} x="35" y={i * 50 + 5} fill="#888" fontSize="10" textAnchor="end">
                      {value.toFixed(1)}
                    </text>
                  );
                });
              })()}
              {/* Line chart */}
              {spreadHistory.length > 0 && (() => {
                const maxSpread = Math.max(...spreadHistory.map(p => p.spread));
                const minSpread = Math.min(...spreadHistory.map(p => p.spread));
                const range = maxSpread - minSpread || 1;
                const points = spreadHistory.map((point, idx) => {
                  const x = 40 + (idx / (spreadHistory.length - 1)) * 740;
                  const y = 200 - ((point.spread - minSpread) / range) * 200;
                  return `${x},${y}`;
                }).join(' ');
                return (
                  <>
                    <polyline
                      points={points}
                      fill="none"
                      stroke="#3B82F6"
                      strokeWidth="2"
                    />
                    {/* Average line */}
                    {(() => {
                      const data = spreadData.find(s => s.symbol === selectedSymbol);
                      if (!data) return null;
                      const avgY = 200 - ((data.avgSpread - minSpread) / range) * 200;
                      return (
                        <>
                          <line x1="40" y1={avgY} x2="780" y2={avgY} stroke="#F59E0B" strokeWidth="1" strokeDasharray="4,4" />
                          <text x="785" y={avgY + 4} fill="#F59E0B" fontSize="10">Avg</text>
                        </>
                      );
                    })()}
                  </>
                );
              })()}
            </svg>
          </div>
        </div>

        {/* Two-column layout for heatmap and LP comparison */}
        <div className="grid grid-cols-2 gap-6 mb-6">
          {/* Time-of-Day Heatmap */}
          <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
            <h3 className="text-sm font-semibold mb-3">Spread by Hour of Day - {selectedSymbol}</h3>
            <div className="w-full">
              <svg viewBox="0 0 600 120" className="w-full">
                {heatmapData.map((spread, hour) => {
                  const data = spreadData.find(s => s.symbol === selectedSymbol);
                  if (!data) return null;
                  const intensity = (spread - data.minSpread) / (data.maxSpread - data.minSpread);
                  const color = intensity > 0.7 ? '#EF4444' : intensity > 0.4 ? '#F59E0B' : '#10B981';
                  const x = hour * 25;
                  return (
                    <g key={hour}>
                      <rect
                        x={x}
                        y="20"
                        width="23"
                        height="60"
                        fill={color}
                        opacity={0.3 + intensity * 0.7}
                      />
                      <text x={x + 11.5} y="95" fill="#888" fontSize="9" textAnchor="middle">
                        {hour.toString().padStart(2, '0')}
                      </text>
                      <text x={x + 11.5} y="50" fill="#E5E5E5" fontSize="8" textAnchor="middle">
                        {spread.toFixed(1)}
                      </text>
                    </g>
                  );
                })}
              </svg>
            </div>
          </div>

          {/* LP Spread Comparison */}
          <div className="bg-[#27272a] border border-[#3f3f46] rounded p-4">
            <h3 className="text-sm font-semibold mb-3">LP Spread Comparison - {selectedSymbol}</h3>
            <div className="space-y-3">
              {lpSpreads.map(lp => (
                <div key={lp.lpName} className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded" style={{ backgroundColor: lp.color }} />
                    <span className="text-xs text-[#E5E5E5]">{lp.lpName}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xs font-semibold" style={{ color: lp.color }}>
                      {lp.spread.toFixed(2)} pips
                    </span>
                    <div className="w-32 h-2 bg-[#1e1e22] rounded-full overflow-hidden">
                      <div
                        className="h-full rounded-full"
                        style={{
                          width: `${(lp.spread / Math.max(...lpSpreads.map(l => l.spread))) * 100}%`,
                          backgroundColor: lp.color
                        }}
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Alert Configuration Modal */}
      {showAlertModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg w-full max-w-md">
            <div className="flex items-center justify-between px-4 py-3 border-b border-[#3f3f46]">
              <h3 className="text-sm font-semibold">Configure Spread Alert</h3>
              <button onClick={() => setShowAlertModal(false)} className="text-[#888] hover:text-[#E5E5E5]">
                <X size={16} />
              </button>
            </div>
            <div className="p-4 space-y-4">
              <div>
                <label className="block text-xs text-[#888] mb-1">Symbol</label>
                <select
                  value={alertSymbol}
                  onChange={(e) => setAlertSymbol(e.target.value)}
                  className="w-full bg-[#1e1e22] border border-[#3f3f46] rounded px-3 py-2 text-xs focus:outline-none focus:border-[#3B82F6]"
                >
                  <option value="">Select symbol...</option>
                  {spreadData.map(item => (
                    <option key={item.symbol} value={item.symbol}>{item.symbol}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs text-[#888] mb-1">Alert Threshold (pips)</label>
                <input
                  type="number"
                  value={alertThreshold}
                  onChange={(e) => setAlertThreshold(parseFloat(e.target.value))}
                  step="0.1"
                  min="0"
                  className="w-full bg-[#1e1e22] border border-[#3f3f46] rounded px-3 py-2 text-xs focus:outline-none focus:border-[#3B82F6]"
                  placeholder="e.g., 2.0"
                />
              </div>
              <div className="flex gap-2 pt-2">
                <button
                  onClick={handleSetAlert}
                  className="flex-1 px-4 py-2 bg-[#3B82F6] hover:bg-[#2563EB] text-white text-xs rounded transition-colors"
                >
                  Set Alert
                </button>
                <button
                  onClick={() => setShowAlertModal(false)}
                  className="flex-1 px-4 py-2 bg-[#3f3f46] hover:bg-[#52525b] text-[#E5E5E5] text-xs rounded transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Symbol Comparison Modal */}
      {showComparisonModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#27272a] border border-[#3f3f46] rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between px-4 py-3 border-b border-[#3f3f46]">
              <h3 className="text-sm font-semibold">Compare Symbol Spreads</h3>
              <button onClick={() => setShowComparisonModal(false)} className="text-[#888] hover:text-[#E5E5E5]">
                <X size={16} />
              </button>
            </div>
            <div className="p-4 flex-1 overflow-auto">
              {/* Symbol selection */}
              <div className="mb-4">
                <label className="block text-xs text-[#888] mb-2">Select symbols to compare (max 5)</label>
                <div className="grid grid-cols-4 gap-2">
                  {spreadData.slice(0, 20).map(item => (
                    <button
                      key={item.symbol}
                      onClick={() => toggleComparison(item.symbol)}
                      disabled={!selectedForComparison.includes(item.symbol) && selectedForComparison.length >= 5}
                      className={`px-3 py-1.5 text-xs rounded transition-colors ${
                        selectedForComparison.includes(item.symbol)
                          ? 'bg-[#3B82F6] text-white'
                          : 'bg-[#1e1e22] text-[#888] hover:bg-[#3f3f46] border border-[#3f3f46]'
                      } disabled:opacity-50 disabled:cursor-not-allowed`}
                    >
                      {item.symbol}
                    </button>
                  ))}
                </div>
              </div>

              {/* Comparison chart */}
              <div className="bg-[#1e1e22] border border-[#3f3f46] rounded p-4">
                <h4 className="text-xs font-semibold mb-3">24h Spread Comparison</h4>
                <div className="w-full h-80">
                  <svg viewBox="0 0 800 300" className="w-full h-full">
                    {/* Grid */}
                    {[0, 1, 2, 3, 4, 5].map(i => (
                      <line key={i} x1="40" y1={i * 60} x2="780" y2={i * 60} stroke="#3f3f46" strokeWidth="0.5" />
                    ))}
                    {/* Lines for each selected symbol */}
                    {selectedForComparison.map((symbol, idx) => {
                      const history = generateSpreadHistory(symbol);
                      if (history.length === 0) return null;

                      const maxSpread = Math.max(...selectedForComparison.flatMap(s => generateSpreadHistory(s).map(p => p.spread)));
                      const minSpread = Math.min(...selectedForComparison.flatMap(s => generateSpreadHistory(s).map(p => p.spread)));
                      const range = maxSpread - minSpread || 1;

                      const colors = ['#3B82F6', '#10B981', '#F59E0B', '#8B5CF6', '#EF4444'];
                      const color = colors[idx % colors.length];

                      const points = history.map((point, i) => {
                        const x = 40 + (i / (history.length - 1)) * 740;
                        const y = 300 - ((point.spread - minSpread) / range) * 300;
                        return `${x},${y}`;
                      }).join(' ');

                      return (
                        <g key={symbol}>
                          <polyline
                            points={points}
                            fill="none"
                            stroke={color}
                            strokeWidth="2"
                          />
                          <text x="785" y={20 + idx * 15} fill={color} fontSize="11">
                            {symbol}
                          </text>
                        </g>
                      );
                    })}
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
