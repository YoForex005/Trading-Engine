/**
 * Order Flow Dashboard
 * Visualizes real-time order flow and market microstructure data
 */

'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { API_CONFIG } from '../../config/api';

// Types
interface FlowTrade {
  id: number;
  timestamp: number;
  symbol: string;
  side: 'BUY' | 'SELL';
  volume: number;
  price: number;
  aggressor: boolean;
  lp: string;
}

interface FlowMetrics {
  totalVolume: number;
  buyVolume: number;
  sellVolume: number;
  buyRatio: number;
  vwap: number;
  toxicityScore: number;
  largeOrderCount: number;
}

interface SymbolImbalance {
  symbol: string;
  buyVolume: number;
  sellVolume: number;
  netVolume: number;
  direction: 'BUY' | 'SELL' | 'NEUTRAL';
}

interface LargeOrder extends FlowTrade {
  significance: number; // 1-10 score
}

interface VolumeCell {
  time: number;
  price: number;
  volume: number;
}

// Mock data generators (replace with real API calls)
function generateMockTrades(count: number): FlowTrade[] {
  const trades: FlowTrade[] = [];
  const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'XAUUSD'];
  const lps = ['YOFX', 'Binance', 'Oanda', 'IC Markets'];
  const now = Date.now();

  for (let i = 0; i < count; i++) {
    const secondsAgo = Math.floor(Math.random() * 300);
    trades.push({
      id: i + 1,
      timestamp: now - secondsAgo * 1000,
      symbol: symbols[Math.floor(Math.random() * symbols.length)],
      side: Math.random() > 0.5 ? 'BUY' : 'SELL',
      volume: parseFloat((Math.random() * 20 + 1).toFixed(2)),
      price: 1.0850 + (Math.random() - 0.5) * 0.01,
      aggressor: Math.random() > 0.3,
      lp: lps[Math.floor(Math.random() * lps.length)],
    });
  }

  return trades.sort((a, b) => b.timestamp - a.timestamp);
}

function calculateMetrics(trades: FlowTrade[]): FlowMetrics {
  const buyTrades = trades.filter(t => t.side === 'BUY');
  const sellTrades = trades.filter(t => t.side === 'SELL');

  const buyVolume = buyTrades.reduce((sum, t) => sum + t.volume, 0);
  const sellVolume = sellTrades.reduce((sum, t) => sum + t.volume, 0);
  const totalVolume = buyVolume + sellVolume;

  const vwap = trades.reduce((sum, t) => sum + t.price * t.volume, 0) / totalVolume;

  const largeOrderCount = trades.filter(t => t.volume >= 10).length;

  // Toxicity score based on aggressor ratio, large orders, imbalance
  const aggressorRatio = trades.filter(t => t.aggressor).length / trades.length;
  const imbalance = Math.abs(buyVolume - sellVolume) / totalVolume;
  const toxicityScore = Math.min(100, (aggressorRatio * 40 + imbalance * 40 + (largeOrderCount / trades.length) * 20));

  return {
    totalVolume,
    buyVolume,
    sellVolume,
    buyRatio: totalVolume > 0 ? (buyVolume / totalVolume) * 100 : 50,
    vwap,
    toxicityScore,
    largeOrderCount,
  };
}

function calculateSymbolImbalances(trades: FlowTrade[]): SymbolImbalance[] {
  const symbolMap = new Map<string, { buy: number; sell: number }>();

  trades.forEach(trade => {
    const current = symbolMap.get(trade.symbol) || { buy: 0, sell: 0 };
    if (trade.side === 'BUY') current.buy += trade.volume;
    else current.sell += trade.volume;
    symbolMap.set(trade.symbol, current);
  });

  return Array.from(symbolMap.entries()).map(([symbol, volumes]): SymbolImbalance => {
    const netVolume = volumes.buy - volumes.sell;
    return {
      symbol,
      buyVolume: volumes.buy,
      sellVolume: volumes.sell,
      netVolume,
      direction: Math.abs(netVolume) < 5 ? 'NEUTRAL' : netVolume > 0 ? 'BUY' : 'SELL',
    };
  }).sort((a, b) => Math.abs(b.netVolume) - Math.abs(a.netVolume));
}

function extractLargeOrders(trades: FlowTrade[], threshold: number): LargeOrder[] {
  return trades
    .filter(t => t.volume >= threshold)
    .map(t => ({
      ...t,
      significance: Math.min(10, Math.floor((t.volume / threshold) * 5)),
    }))
    .sort((a, b) => b.significance - a.significance);
}

// Sub-components
interface FlowTapeProps {
  trades: FlowTrade[];
}

const FlowTape: React.FC<FlowTapeProps> = ({ trades }) => {
  return (
    <div className="bg-zinc-900 rounded border border-zinc-700 overflow-hidden">
      <div className="bg-zinc-800 border-b border-zinc-700 px-3 py-2">
        <h3 className="text-sm font-medium text-zinc-100">Live Order Flow Tape</h3>
      </div>
      <div className="h-[300px] overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700">
        <table className="w-full text-xs">
          <thead className="sticky top-0 bg-zinc-800 border-b border-zinc-700">
            <tr>
              <th className="px-2 py-1 text-left text-zinc-400">Time</th>
              <th className="px-2 py-1 text-left text-zinc-400">Symbol</th>
              <th className="px-2 py-1 text-left text-zinc-400">Side</th>
              <th className="px-2 py-1 text-right text-zinc-400">Volume</th>
              <th className="px-2 py-1 text-right text-zinc-400">Price</th>
              <th className="px-2 py-1 text-center text-zinc-400">Agg</th>
              <th className="px-2 py-1 text-left text-zinc-400">LP</th>
            </tr>
          </thead>
          <tbody>
            {trades.slice(0, 100).map((trade) => (
              <tr key={trade.id} className="border-b border-zinc-800 hover:bg-zinc-800/50">
                <td className="px-2 py-1 text-zinc-400">
                  {new Date(trade.timestamp).toLocaleTimeString('en-US', { hour12: false })}
                </td>
                <td className="px-2 py-1 text-zinc-300 font-medium">{trade.symbol}</td>
                <td className={`px-2 py-1 font-medium ${
                  trade.side === 'BUY' ? 'text-emerald-400' : 'text-red-400'
                }`}>
                  {trade.side}
                </td>
                <td className="px-2 py-1 text-right text-zinc-300">{trade.volume.toFixed(2)}</td>
                <td className="px-2 py-1 text-right text-zinc-300">{trade.price.toFixed(5)}</td>
                <td className="px-2 py-1 text-center">
                  {trade.aggressor && <span className="text-orange-400">●</span>}
                </td>
                <td className="px-2 py-1 text-zinc-400 text-xs">{trade.lp}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

interface MetricsCardsProps {
  metrics: FlowMetrics;
}

const MetricsCards: React.FC<MetricsCardsProps> = ({ metrics }) => {
  const cards = [
    { label: 'Total Volume', value: metrics.totalVolume.toFixed(2), unit: 'lots', color: 'text-blue-400' },
    { label: 'Buy/Sell Ratio', value: `${metrics.buyRatio.toFixed(1)}%`, unit: 'buy', color: metrics.buyRatio > 50 ? 'text-emerald-400' : 'text-red-400' },
    { label: 'VWAP', value: metrics.vwap.toFixed(5), unit: '', color: 'text-purple-400' },
    { label: 'Toxicity Score', value: metrics.toxicityScore.toFixed(1), unit: '/100', color: metrics.toxicityScore > 60 ? 'text-red-400' : 'text-emerald-400' },
    { label: 'Large Orders', value: metrics.largeOrderCount.toString(), unit: 'count', color: 'text-yellow-400' },
  ];

  return (
    <div className="grid grid-cols-5 gap-3">
      {cards.map((card, idx) => (
        <div key={idx} className="bg-zinc-900 rounded border border-zinc-700 p-3">
          <div className="text-xs text-zinc-400 mb-1">{card.label}</div>
          <div className={`text-xl font-bold ${card.color}`}>
            {card.value}
            {card.unit && <span className="text-xs text-zinc-500 ml-1">{card.unit}</span>}
          </div>
        </div>
      ))}
    </div>
  );
};

interface VolumeHeatmapProps {
  trades: FlowTrade[];
}

const VolumeHeatmap: React.FC<VolumeHeatmapProps> = ({ trades }) => {
  const cells = useMemo(() => {
    // Create 10 price levels x 12 time slots
    const priceSteps = 10;
    const timeSteps = 12;
    const minPrice = Math.min(...trades.map(t => t.price));
    const maxPrice = Math.max(...trades.map(t => t.price));
    const priceRange = maxPrice - minPrice;

    const now = Date.now();
    const timeRange = 300000; // 5 minutes

    const grid: number[][] = Array(priceSteps).fill(0).map(() => Array(timeSteps).fill(0));

    trades.forEach(trade => {
      const timeSlot = Math.floor(((now - trade.timestamp) / timeRange) * timeSteps);
      const priceSlot = Math.floor(((trade.price - minPrice) / priceRange) * priceSteps);

      if (timeSlot >= 0 && timeSlot < timeSteps && priceSlot >= 0 && priceSlot < priceSteps) {
        grid[priceSteps - 1 - priceSlot][timeSteps - 1 - timeSlot] += trade.volume;
      }
    });

    const maxVolume = Math.max(...grid.flat());

    return { grid, maxVolume, minPrice, maxPrice, priceRange };
  }, [trades]);

  return (
    <div className="bg-zinc-900 rounded border border-zinc-700 p-3">
      <h3 className="text-sm font-medium text-zinc-100 mb-3">Volume Heatmap (Time × Price)</h3>
      <div className="grid grid-cols-12 gap-1">
        {cells.grid.map((row, rowIdx) =>
          row.map((volume, colIdx) => {
            const intensity = cells.maxVolume > 0 ? volume / cells.maxVolume : 0;
            const bgColor = volume === 0
              ? 'bg-zinc-800'
              : intensity > 0.7
              ? 'bg-red-500'
              : intensity > 0.4
              ? 'bg-orange-500'
              : 'bg-yellow-500';

            return (
              <div
                key={`${rowIdx}-${colIdx}`}
                className={`h-6 rounded ${bgColor} opacity-${Math.floor(intensity * 100)}`}
                title={`Volume: ${volume.toFixed(2)}`}
              />
            );
          })
        )}
      </div>
      <div className="flex justify-between text-xs text-zinc-500 mt-2">
        <span>5m ago</span>
        <span>Now</span>
      </div>
    </div>
  );
};

interface ImbalanceChartProps {
  imbalances: SymbolImbalance[];
}

const ImbalanceChart: React.FC<ImbalanceChartProps> = ({ imbalances }) => {
  const maxVolume = Math.max(...imbalances.map(i => Math.max(i.buyVolume, i.sellVolume)));

  return (
    <div className="bg-zinc-900 rounded border border-zinc-700 p-3">
      <h3 className="text-sm font-medium text-zinc-100 mb-3">Order Imbalance by Symbol</h3>
      <div className="space-y-2">
        {imbalances.map((imb, idx) => {
          const buyPercent = (imb.buyVolume / maxVolume) * 50;
          const sellPercent = (imb.sellVolume / maxVolume) * 50;

          return (
            <div key={idx} className="flex items-center gap-2">
              <div className="w-16 text-xs text-zinc-400 font-medium">{imb.symbol}</div>
              <div className="flex-1 flex items-center">
                <div
                  className="bg-red-500 h-4 rounded-l"
                  style={{ width: `${sellPercent}%` }}
                />
                <div className="px-2 text-xs text-zinc-400">
                  {imb.direction === 'BUY' ? '→' : imb.direction === 'SELL' ? '←' : '•'}
                </div>
                <div
                  className="bg-emerald-500 h-4 rounded-r"
                  style={{ width: `${buyPercent}%` }}
                />
              </div>
              <div className="w-12 text-xs text-right text-zinc-400">
                {Math.abs(imb.netVolume).toFixed(1)}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

interface LargeOrdersPanelProps {
  orders: LargeOrder[];
}

const LargeOrdersPanel: React.FC<LargeOrdersPanelProps> = ({ orders }) => {
  return (
    <div className="bg-zinc-900 rounded border border-zinc-700 overflow-hidden">
      <div className="bg-zinc-800 border-b border-zinc-700 px-3 py-2">
        <h3 className="text-sm font-medium text-zinc-100">Large Orders (≥10 lots)</h3>
      </div>
      <div className="h-[200px] overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700">
        <table className="w-full text-xs">
          <thead className="sticky top-0 bg-zinc-800 border-b border-zinc-700">
            <tr>
              <th className="px-2 py-1 text-left text-zinc-400">Time</th>
              <th className="px-2 py-1 text-left text-zinc-400">Symbol</th>
              <th className="px-2 py-1 text-left text-zinc-400">Side</th>
              <th className="px-2 py-1 text-right text-zinc-400">Volume</th>
              <th className="px-2 py-1 text-right text-zinc-400">Significance</th>
            </tr>
          </thead>
          <tbody>
            {orders.map((order) => (
              <tr key={order.id} className="border-b border-zinc-800 hover:bg-zinc-800/50">
                <td className="px-2 py-1 text-zinc-400">
                  {new Date(order.timestamp).toLocaleTimeString('en-US', { hour12: false })}
                </td>
                <td className="px-2 py-1 text-zinc-300 font-medium">{order.symbol}</td>
                <td className={`px-2 py-1 font-medium ${
                  order.side === 'BUY' ? 'text-emerald-400' : 'text-red-400'
                }`}>
                  {order.side}
                </td>
                <td className="px-2 py-1 text-right text-yellow-400 font-bold">{order.volume.toFixed(2)}</td>
                <td className="px-2 py-1 text-right">
                  <span className="text-orange-400">{order.significance}/10</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

interface TopSymbolsChartProps {
  imbalances: SymbolImbalance[];
}

const TopSymbolsChart: React.FC<TopSymbolsChartProps> = ({ imbalances }) => {
  const topSymbols = useMemo(() => {
    return imbalances
      .map(i => ({
        symbol: i.symbol,
        totalVolume: i.buyVolume + i.sellVolume,
      }))
      .sort((a, b) => b.totalVolume - a.totalVolume)
      .slice(0, 5);
  }, [imbalances]);

  const maxVolume = topSymbols[0]?.totalVolume || 1;

  return (
    <div className="bg-zinc-900 rounded border border-zinc-700 p-3">
      <h3 className="text-sm font-medium text-zinc-100 mb-3">Top Symbols by Flow Volume</h3>
      <div className="space-y-2">
        {topSymbols.map((item, idx) => (
          <div key={idx} className="flex items-center gap-2">
            <div className="w-12 text-xs text-zinc-400 font-medium">{item.symbol}</div>
            <div className="flex-1">
              <div
                className="bg-blue-500 h-6 rounded flex items-center justify-end pr-2"
                style={{ width: `${(item.totalVolume / maxVolume) * 100}%` }}
              >
                <span className="text-xs text-white font-medium">{item.totalVolume.toFixed(1)}</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

interface ToxicityGaugeProps {
  score: number;
}

const ToxicityGauge: React.FC<ToxicityGaugeProps> = ({ score }) => {
  const angle = (score / 100) * 180;
  const color = score > 70 ? '#ef4444' : score > 40 ? '#f59e0b' : '#10b981';

  return (
    <div className="bg-zinc-900 rounded border border-zinc-700 p-3">
      <h3 className="text-sm font-medium text-zinc-100 mb-2">Flow Toxicity</h3>
      <div className="flex flex-col items-center">
        <svg width="140" height="80" viewBox="0 0 140 80">
          {/* Background arc */}
          <path
            d="M 20 70 A 50 50 0 0 1 120 70"
            fill="none"
            stroke="#3f3f46"
            strokeWidth="12"
          />
          {/* Colored arc */}
          <path
            d="M 20 70 A 50 50 0 0 1 120 70"
            fill="none"
            stroke={color}
            strokeWidth="12"
            strokeDasharray="157"
            strokeDashoffset={157 - (157 * score) / 100}
          />
          {/* Needle */}
          <line
            x1="70"
            y1="70"
            x2={70 + Math.cos((180 - angle) * (Math.PI / 180)) * 45}
            y2={70 - Math.sin((180 - angle) * (Math.PI / 180)) * 45}
            stroke={color}
            strokeWidth="2"
          />
          <circle cx="70" cy="70" r="4" fill={color} />
        </svg>
        <div className="text-2xl font-bold mt-2" style={{ color }}>
          {score.toFixed(1)}
        </div>
        <div className="text-xs text-zinc-500">/ 100</div>
      </div>
    </div>
  );
};

// Main Component
export default function OrderFlowDashboard() {
  const [trades, setTrades] = useState<FlowTrade[]>([]);
  const [selectedSymbol, setSelectedSymbol] = useState<string>('ALL');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [largeOrderThreshold] = useState(10);

  // Initialize data
  useEffect(() => {
    setTrades(generateMockTrades(200));
  }, []);

  // Auto-refresh
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      setTrades(generateMockTrades(200));
    }, 5000);

    return () => clearInterval(interval);
  }, [autoRefresh]);

  // Filter trades by symbol
  const filteredTrades = useMemo(() => {
    return selectedSymbol === 'ALL' ? trades : trades.filter(t => t.symbol === selectedSymbol);
  }, [trades, selectedSymbol]);

  // Calculate metrics
  const metrics = useMemo(() => calculateMetrics(filteredTrades), [filteredTrades]);
  const imbalances = useMemo(() => calculateSymbolImbalances(trades), [trades]);
  const largeOrders = useMemo(() => extractLargeOrders(filteredTrades, largeOrderThreshold), [filteredTrades, largeOrderThreshold]);

  const symbols = ['ALL', ...Array.from(new Set(trades.map(t => t.symbol)))];

  return (
    <div className="p-6 bg-zinc-950 min-h-screen text-zinc-100">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-zinc-100">Order Flow Dashboard</h1>
        <p className="text-sm text-zinc-400 mt-1">Real-time order flow and market microstructure analysis</p>
      </div>

      {/* Controls */}
      <div className="mb-6 flex items-center gap-4">
        <div className="flex items-center gap-2">
          <label className="text-sm text-zinc-400">Symbol:</label>
          <select
            value={selectedSymbol}
            onChange={(e) => setSelectedSymbol(e.target.value)}
            className="bg-zinc-900 border border-zinc-700 rounded px-3 py-1 text-sm text-zinc-100"
          >
            {symbols.map(sym => (
              <option key={sym} value={sym}>{sym}</option>
            ))}
          </select>
        </div>
        <label className="flex items-center gap-2 text-sm text-zinc-400">
          <input
            type="checkbox"
            checked={autoRefresh}
            onChange={(e) => setAutoRefresh(e.target.checked)}
            className="w-4 h-4"
          />
          Auto-refresh (5s)
        </label>
      </div>

      {/* Metrics Cards */}
      <MetricsCards metrics={metrics} />

      {/* Main Grid */}
      <div className="grid grid-cols-3 gap-4 mt-4">
        {/* Left Column */}
        <div className="col-span-2 space-y-4">
          <FlowTape trades={filteredTrades} />
          <VolumeHeatmap trades={filteredTrades} />
        </div>

        {/* Right Column */}
        <div className="space-y-4">
          <ToxicityGauge score={metrics.toxicityScore} />
          <TopSymbolsChart imbalances={imbalances} />
        </div>
      </div>

      {/* Bottom Grid */}
      <div className="grid grid-cols-2 gap-4 mt-4">
        <ImbalanceChart imbalances={imbalances} />
        <LargeOrdersPanel orders={largeOrders} />
      </div>
    </div>
  );
}
