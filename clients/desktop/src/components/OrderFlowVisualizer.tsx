/**
 * Order Flow Visualizer
 * Real-time order flow and volume analysis panel
 */

import React, { useState, useMemo, useEffect, useRef, useCallback } from 'react';
import { useOrderFlowStore, type TimeframeType } from '../store/useOrderFlowStore';
import { buildApiUrl } from '../config/api';
import { useAppStore } from '../store/useAppStore';

// Types
interface Trade {
  id: number;
  timestamp: number;
  symbol: string;
  side: 'BUY' | 'SELL';
  volume: number;
  price: number;
  aggressor: boolean; // true if market order (taker)
}

interface PriceLevel {
  price: number;
  buyVolume: number;
  sellVolume: number;
}

interface DeltaBar {
  timestamp: number;
  buyVolume: number;
  sellVolume: number;
  delta: number; // buyVolume - sellVolume
}

// Mock data generators
function generateMockTrades(symbol: string, count: number): Trade[] {
  const trades: Trade[] = [];
  const now = Date.now();
  const basePrice = symbol === 'EURUSD' ? 1.0850 : symbol === 'GBPUSD' ? 1.2650 : 1.0950;

  for (let i = 0; i < count; i++) {
    const secondsAgo = Math.floor(Math.random() * 3600); // last hour
    const priceVariation = (Math.random() - 0.5) * 0.0020;
    const side = Math.random() > 0.5 ? 'BUY' : 'SELL';

    trades.push({
      id: count - i,
      timestamp: now - secondsAgo * 1000,
      symbol,
      side,
      volume: parseFloat((Math.random() * 10 + 0.1).toFixed(2)),
      price: parseFloat((basePrice + priceVariation).toFixed(5)),
      aggressor: Math.random() > 0.3, // 70% market orders
    });
  }

  return trades.sort((a, b) => b.timestamp - a.timestamp); // newest first
}

function generateVolumeLevels(symbol: string, count: number): PriceLevel[] {
  const levels: PriceLevel[] = [];
  const basePrice = symbol === 'EURUSD' ? 1.0850 : symbol === 'GBPUSD' ? 1.2650 : 1.0950;
  const tickSize = 0.0001;

  for (let i = 0; i < count; i++) {
    const offset = (i - count / 2) * tickSize;
    levels.push({
      price: parseFloat((basePrice + offset).toFixed(5)),
      buyVolume: parseFloat((Math.random() * 50 + 5).toFixed(2)),
      sellVolume: parseFloat((Math.random() * 50 + 5).toFixed(2)),
    });
  }

  return levels.sort((a, b) => b.price - a.price); // descending
}

function generateDeltaBars(timeframe: TimeframeType, count: number): DeltaBar[] {
  const bars: DeltaBar[] = [];
  const now = Date.now();
  const intervalMs = timeframe === '1m' ? 60000 : timeframe === '5m' ? 300000 : timeframe === '15m' ? 900000 : 3600000;

  for (let i = 0; i < count; i++) {
    const buyVolume = parseFloat((Math.random() * 100 + 20).toFixed(2));
    const sellVolume = parseFloat((Math.random() * 100 + 20).toFixed(2));

    bars.push({
      timestamp: now - (count - i) * intervalMs,
      buyVolume,
      sellVolume,
      delta: buyVolume - sellVolume,
    });
  }

  return bars;
}

// Sub-components
interface OrderFlowTapeProps {
  trades: Trade[];
  largeOrderThreshold: number;
  autoScroll: boolean;
}

const OrderFlowTape: React.FC<OrderFlowTapeProps> = ({ trades, largeOrderThreshold, autoScroll }) => {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [flashingIds, setFlashingIds] = useState<Set<number>>(new Set());

  useEffect(() => {
    if (autoScroll && scrollRef.current) {
      scrollRef.current.scrollTop = 0;
    }
  }, [trades, autoScroll]);

  // Flash large orders
  useEffect(() => {
    const largeOrders = trades.filter(t => t.volume >= largeOrderThreshold).map(t => t.id);
    if (largeOrders.length > 0) {
      setFlashingIds(new Set(largeOrders));
      const timer = setTimeout(() => setFlashingIds(new Set()), 1000);
      return () => clearTimeout(timer);
    }
  }, [trades, largeOrderThreshold]);

  return (
    <div ref={scrollRef} className="h-full overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700">
      <table className="w-full text-[10px]">
        <thead className="sticky top-0 bg-zinc-800 border-b border-zinc-700">
          <tr>
            <th className="px-2 py-1 text-left text-zinc-400">Time</th>
            <th className="px-2 py-1 text-left text-zinc-400">Side</th>
            <th className="px-2 py-1 text-right text-zinc-400">Volume</th>
            <th className="px-2 py-1 text-right text-zinc-400">Price</th>
            <th className="px-2 py-1 text-center text-zinc-400">Agg</th>
          </tr>
        </thead>
        <tbody>
          {trades.slice(0, 100).map((trade) => {
            const isLarge = trade.volume >= largeOrderThreshold;
            const isFlashing = flashingIds.has(trade.id);

            return (
              <tr
                key={trade.id}
                className={`border-b border-zinc-800 transition-colors ${
                  isFlashing ? 'animate-pulse bg-yellow-500/20' : 'hover:bg-zinc-800/50'
                }`}
              >
                <td className="px-2 py-1 text-zinc-400">
                  {new Date(trade.timestamp).toLocaleTimeString('en-US', { hour12: false })}
                </td>
                <td className={`px-2 py-1 font-medium ${
                  trade.side === 'BUY' ? 'text-emerald-400' : 'text-red-400'
                }`}>
                  {trade.side}
                </td>
                <td className={`px-2 py-1 text-right ${isLarge ? 'text-yellow-400 font-bold' : 'text-zinc-300'}`}>
                  {trade.volume.toFixed(2)}
                </td>
                <td className="px-2 py-1 text-right text-zinc-300">
                  {trade.price.toFixed(5)}
                </td>
                <td className="px-2 py-1 text-center">
                  {trade.aggressor && <span className="text-orange-400">●</span>}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};

interface VolumeProfileProps {
  levels: PriceLevel[];
}

const VolumeProfile: React.FC<VolumeProfileProps> = ({ levels }) => {
  const maxVolume = useMemo(() => {
    return Math.max(...levels.map(l => l.buyVolume + l.sellVolume));
  }, [levels]);

  return (
    <div className="h-full overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700">
      <svg width="100%" height={levels.length * 20} className="bg-zinc-900">
        {levels.map((level, idx) => {
          const totalVolume = level.buyVolume + level.sellVolume;
          const buyPercent = (level.buyVolume / maxVolume) * 50;
          const sellPercent = (level.sellVolume / maxVolume) * 50;
          const y = idx * 20;

          return (
            <g key={idx}>
              {/* Price label */}
              <text x="5" y={y + 14} className="text-[9px] fill-zinc-400">
                {level.price.toFixed(5)}
              </text>

              {/* Buy volume bar (right) */}
              <rect
                x="50%"
                y={y + 2}
                width={`${buyPercent}%`}
                height="16"
                className="fill-emerald-500/60"
              />

              {/* Sell volume bar (left) */}
              <rect
                x={`${50 - sellPercent}%`}
                y={y + 2}
                width={`${sellPercent}%`}
                height="16"
                className="fill-red-500/60"
              />

              {/* Center line */}
              <line
                x1="50%"
                y1={y}
                x2="50%"
                y2={y + 20}
                className="stroke-zinc-700"
                strokeWidth="1"
              />
            </g>
          );
        })}
      </svg>
    </div>
  );
};

interface DeltaChartProps {
  bars: DeltaBar[];
  showCumulative: boolean;
}

const DeltaChart: React.FC<DeltaChartProps> = ({ bars, showCumulative }) => {
  const svgWidth = 600;
  const svgHeight = 200;
  const barWidth = svgWidth / bars.length;

  const maxDelta = useMemo(() => {
    return Math.max(...bars.map(b => Math.abs(b.delta)));
  }, [bars]);

  const cumulativeDelta = useMemo(() => {
    let cumulative = 0;
    return bars.map(b => {
      cumulative += b.delta;
      return cumulative;
    });
  }, [bars]);

  const maxCumulative = useMemo(() => {
    return Math.max(...cumulativeDelta.map(Math.abs));
  }, [cumulativeDelta]);

  return (
    <div className="h-full bg-zinc-900 p-2">
      <svg width="100%" height={svgHeight} viewBox={`0 0 ${svgWidth} ${svgHeight}`}>
        {/* Center line */}
        <line
          x1="0"
          y1={svgHeight / 2}
          x2={svgWidth}
          y2={svgHeight / 2}
          className="stroke-zinc-700"
          strokeWidth="1"
        />

        {/* Delta bars */}
        {bars.map((bar, idx) => {
          const barHeight = (Math.abs(bar.delta) / maxDelta) * (svgHeight / 2 - 10);
          const x = idx * barWidth;
          const y = bar.delta >= 0 ? svgHeight / 2 - barHeight : svgHeight / 2;
          const color = bar.delta >= 0 ? '#10b981' : '#ef4444';

          return (
            <rect
              key={idx}
              x={x}
              y={y}
              width={barWidth - 1}
              height={barHeight}
              fill={color}
              opacity="0.7"
            />
          );
        })}

        {/* Cumulative delta line */}
        {showCumulative && (
          <polyline
            points={cumulativeDelta
              .map((val, idx) => {
                const x = idx * barWidth + barWidth / 2;
                const y = svgHeight / 2 - (val / maxCumulative) * (svgHeight / 2 - 20);
                return `${x},${y}`;
              })
              .join(' ')}
            fill="none"
            stroke="#3b82f6"
            strokeWidth="2"
          />
        )}
      </svg>
      <div className="flex justify-between text-[9px] text-zinc-500 mt-1">
        <span>Delta Chart</span>
        {showCumulative && <span className="text-blue-400">Cumulative Delta (blue line)</span>}
      </div>
    </div>
  );
};

interface ImbalanceGaugeProps {
  buyVolume: number;
  sellVolume: number;
}

const ImbalanceGauge: React.FC<ImbalanceGaugeProps> = ({ buyVolume, sellVolume }) => {
  const total = buyVolume + sellVolume;
  const buyPercent = total > 0 ? (buyVolume / total) * 100 : 50;
  const sellPercent = total > 0 ? (sellVolume / total) * 100 : 50;

  return (
    <div className="bg-zinc-800 rounded p-3 border border-zinc-700">
      <div className="text-[10px] text-zinc-400 mb-2">Buy/Sell Imbalance</div>
      <div className="flex items-center gap-2">
        <div className="text-[9px] text-emerald-400 w-12 text-right">{buyPercent.toFixed(1)}%</div>
        <div className="flex-1 h-6 bg-zinc-900 rounded overflow-hidden flex">
          <div
            className="bg-emerald-500 transition-all duration-300"
            style={{ width: `${buyPercent}%` }}
          />
          <div
            className="bg-red-500 transition-all duration-300"
            style={{ width: `${sellPercent}%` }}
          />
        </div>
        <div className="text-[9px] text-red-400 w-12">{sellPercent.toFixed(1)}%</div>
      </div>
      <div className="flex justify-between mt-2 text-[9px] text-zinc-500">
        <span>Buy: {buyVolume.toFixed(2)}</span>
        <span>Sell: {sellVolume.toFixed(2)}</span>
      </div>
    </div>
  );
};

// Main component
export const OrderFlowVisualizer: React.FC = () => {
  const {
    selectedSymbol,
    timeframe,
    largeOrderThreshold,
    showCumulativeDelta,
    autoScroll,
    setSelectedSymbol,
    setTimeframe,
    setLargeOrderThreshold,
    setShowCumulativeDelta,
    setAutoScroll,
  } = useOrderFlowStore();

  const [currentTime, setCurrentTime] = useState(Date.now());
  const [trades, setTrades] = useState<Trade[]>(() => generateMockTrades(selectedSymbol, 200));
  const [dataSource, setDataSource] = useState<'api' | 'mock'>('mock');

  // Fetch trades from real backend API: GET /admin/order-flow/live
  const fetchLiveFlow = useCallback(async () => {
    try {
      const authToken = useAppStore.getState().authToken;
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      if (authToken) {
        headers['Authorization'] = `Bearer ${authToken}`;
      }

      const response = await fetch(
        buildApiUrl(`/admin/order-flow/live?symbol=${selectedSymbol}&limit=200`),
        { headers }
      );

      if (response.ok) {
        const data = await response.json();
        if (data.entries && Array.isArray(data.entries) && data.entries.length > 0) {
          // Map backend OrderFlowEntry to component Trade type
          const apiTrades: Trade[] = data.entries.map((entry: any, idx: number) => ({
            id: entry.id || idx,
            timestamp: new Date(entry.timestamp).getTime(),
            symbol: entry.symbol || selectedSymbol,
            side: (entry.side || '').toUpperCase() === 'BUY' ? 'BUY' as const : 'SELL' as const,
            volume: entry.volume || 0,
            price: entry.price || 0,
            aggressor: entry.aggressor ?? false,
          }));
          setTrades(apiTrades.sort((a, b) => b.timestamp - a.timestamp));
          setDataSource('api');
          return;
        }
      }

      // API returned no data or error, fall back to mock
      if (dataSource !== 'api') {
        setTrades(generateMockTrades(selectedSymbol, 200));
        setDataSource('mock');
      }
    } catch (error) {
      console.warn('[OrderFlow] Failed to fetch live flow, using mock data:', error);
      if (dataSource !== 'api') {
        setTrades(generateMockTrades(selectedSymbol, 200));
        setDataSource('mock');
      }
    }
  }, [selectedSymbol, dataSource]);

  // Fetch on symbol change and set up polling
  useEffect(() => {
    fetchLiveFlow();
    const interval = setInterval(fetchLiveFlow, 5000); // update every 5 seconds
    return () => clearInterval(interval);
  }, [fetchLiveFlow]);

  // Regenerate dependent data
  const volumeLevels = useMemo(() => generateVolumeLevels(selectedSymbol, 20), [selectedSymbol]);
  const deltaBars = useMemo(() => generateDeltaBars(timeframe, 60), [timeframe, currentTime]);

  // Update time for delta bars refresh
  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentTime(Date.now());
    }, 30000); // refresh delta bars every 30 seconds
    return () => clearInterval(interval);
  }, []);

  // Calculate total buy/sell for imbalance gauge
  const { totalBuy, totalSell } = useMemo(() => {
    const recentTrades = trades.slice(0, 50); // last 50 trades
    return recentTrades.reduce(
      (acc, t) => {
        if (t.side === 'BUY') acc.totalBuy += t.volume;
        else acc.totalSell += t.volume;
        return acc;
      },
      { totalBuy: 0, totalSell: 0 }
    );
  }, [trades]);

  const symbols = ['EURUSD', 'GBPUSD', 'USDJPY', 'AUDUSD', 'USDCAD'];
  const timeframes: TimeframeType[] = ['1m', '5m', '15m', '1h'];

  return (
    <div className="h-full bg-zinc-900 text-zinc-100 flex flex-col">
      {/* Header Controls */}
      <div className="bg-zinc-800 border-b border-zinc-700 p-2 flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span className="text-[10px] text-zinc-400">Symbol:</span>
          <select
            value={selectedSymbol}
            onChange={(e) => setSelectedSymbol(e.target.value)}
            className="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-[11px] text-zinc-100"
          >
            {symbols.map((sym) => (
              <option key={sym} value={sym}>
                {sym}
              </option>
            ))}
          </select>
        </div>

        <div className="flex items-center gap-1">
          <span className="text-[10px] text-zinc-400 mr-1">Timeframe:</span>
          {timeframes.map((tf) => (
            <button
              key={tf}
              onClick={() => setTimeframe(tf)}
              className={`px-2 py-1 text-[10px] rounded transition-colors ${
                timeframe === tf
                  ? 'bg-blue-500 text-white'
                  : 'bg-zinc-700 text-zinc-300 hover:bg-zinc-600'
              }`}
            >
              {tf}
            </button>
          ))}
        </div>

        <div className="flex items-center gap-3">
          <label className="flex items-center gap-1 text-[10px] text-zinc-400">
            <input
              type="checkbox"
              checked={showCumulativeDelta}
              onChange={(e) => setShowCumulativeDelta(e.target.checked)}
              className="w-3 h-3"
            />
            Cumulative
          </label>
          <label className="flex items-center gap-1 text-[10px] text-zinc-400">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
              className="w-3 h-3"
            />
            Auto-scroll
          </label>
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-zinc-400">Large order:</span>
            <input
              type="number"
              value={largeOrderThreshold}
              onChange={(e) => setLargeOrderThreshold(parseFloat(e.target.value) || 5)}
              className="bg-zinc-900 border border-zinc-700 rounded px-2 py-1 text-[11px] text-zinc-100 w-16"
              step="0.5"
              min="1"
            />
            <span className="text-[10px] text-zinc-400">lots</span>
          </div>
        </div>
      </div>

      {/* Main Layout */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left: Order Flow Tape */}
        <div className="w-1/4 border-r border-zinc-700 flex flex-col">
          <div className="bg-zinc-800 border-b border-zinc-700 px-2 py-1 text-[10px] text-zinc-400 font-medium">
            Order Flow Tape
          </div>
          <OrderFlowTape
            trades={trades}
            largeOrderThreshold={largeOrderThreshold}
            autoScroll={autoScroll}
          />
        </div>

        {/* Center: Volume Profile */}
        <div className="w-1/4 border-r border-zinc-700 flex flex-col">
          <div className="bg-zinc-800 border-b border-zinc-700 px-2 py-1 text-[10px] text-zinc-400 font-medium">
            Volume Profile
          </div>
          <VolumeProfile levels={volumeLevels} />
        </div>

        {/* Right: Delta Chart + Imbalance */}
        <div className="flex-1 flex flex-col">
          <div className="bg-zinc-800 border-b border-zinc-700 px-2 py-1 text-[10px] text-zinc-400 font-medium">
            Delta Analysis
          </div>
          <div className="flex-1 overflow-hidden">
            <DeltaChart bars={deltaBars} showCumulative={showCumulativeDelta} />
          </div>
          <div className="p-2">
            <ImbalanceGauge buyVolume={totalBuy} sellVolume={totalSell} />
          </div>
        </div>
      </div>
    </div>
  );
};
