/**
 * Professional Depth of Market (Order Book)
 * Level 2 market depth with cumulative volume visualization and real-time updates
 */

import { useState, useEffect, useMemo, useCallback } from 'react';
import { Layers, TrendingUp, TrendingDown, BarChart2 } from 'lucide-react';
import type { MarketDepth, MarketDepthEntry } from '../../types/trading';

type DepthOfMarketProps = {
  symbol: string;
  depth?: 10 | 20 | 50;
};

export const DepthOfMarket = ({ symbol, depth = 20 }: DepthOfMarketProps) => {
  const [marketDepth, setMarketDepth] = useState<MarketDepth | null>(null);
  const [selectedDepth, setSelectedDepth] = useState<10 | 20 | 50>(depth);
  const [grouping, setGrouping] = useState<number>(1); // Price grouping (pips)
  const [showCumulative, setShowCumulative] = useState(true);

  // Fetch market depth data
  useEffect(() => {
    if (!symbol) return;

    const fetchDepth = async () => {
      try {
        const response = await fetch(
          `http://localhost:8080/api/market-depth?symbol=${symbol}&depth=${selectedDepth}`
        );

        if (response.ok) {
          const data = await response.json();
          setMarketDepth(data);
        } else {
          // Generate mock data if API not available
          setMarketDepth(generateMockDepth(symbol, selectedDepth));
        }
      } catch (error) {
        console.error('Failed to fetch market depth:', error);
        setMarketDepth(generateMockDepth(symbol, selectedDepth));
      }
    };

    fetchDepth();
    const interval = setInterval(fetchDepth, 1000);

    return () => clearInterval(interval);
  }, [symbol, selectedDepth]);

  // Calculate max volume for visualization
  const maxVolume = useMemo(() => {
    if (!marketDepth) return 0;
    const allVolumes = [...marketDepth.bids, ...marketDepth.asks].map((e) => e.volume);
    return Math.max(...allVolumes, 0);
  }, [marketDepth]);

  // Group price levels if needed
  const groupedBids = useMemo(() => {
    if (!marketDepth || grouping === 1) return marketDepth?.bids || [];

    return groupPriceLevels(marketDepth.bids, grouping, 'bid');
  }, [marketDepth, grouping]);

  const groupedAsks = useMemo(() => {
    if (!marketDepth || grouping === 1) return marketDepth?.asks || [];

    return groupPriceLevels(marketDepth.asks, grouping, 'ask');
  }, [marketDepth, grouping]);

  if (!marketDepth) {
    return (
      <div className="flex items-center justify-center h-full bg-zinc-900 text-zinc-500">
        <Layers className="w-6 h-6 animate-pulse mr-2" />
        <span>Loading depth...</span>
      </div>
    );
  }

  const midPrice = (marketDepth.bids[0]?.price + marketDepth.asks[0]?.price) / 2;

  return (
    <div className="flex flex-col h-full bg-zinc-900">
      {/* Header */}
      <div className="flex items-center justify-between px-3 py-2 border-b border-zinc-800 bg-zinc-900/50">
        <div>
          <div className="flex items-center gap-2">
            <Layers className="w-4 h-4 text-emerald-400" />
            <h3 className="text-xs font-semibold text-zinc-300 uppercase tracking-wide">Depth of Market</h3>
          </div>
          <p className="text-[10px] text-zinc-500 mt-0.5">{symbol}</p>
        </div>

        <div className="flex items-center gap-2">
          <select
            value={selectedDepth}
            onChange={(e) => setSelectedDepth(Number(e.target.value) as 10 | 20 | 50)}
            className="px-2 py-1 text-[10px] bg-zinc-800 border border-zinc-700 rounded text-zinc-300 focus:outline-none focus:border-emerald-500"
          >
            <option value={10}>10 Levels</option>
            <option value={20}>20 Levels</option>
            <option value={50}>50 Levels</option>
          </select>
        </div>
      </div>

      {/* Spread Info */}
      <div className="px-3 py-2 border-b border-zinc-800 bg-zinc-900/30">
        <div className="grid grid-cols-3 gap-4 text-xs">
          <div>
            <div className="text-[10px] text-zinc-500 mb-0.5">Spread</div>
            <div className="font-mono text-zinc-300">{marketDepth.spread.toFixed(5)}</div>
          </div>
          <div>
            <div className="text-[10px] text-zinc-500 mb-0.5">Pips</div>
            <div className="font-mono text-zinc-300">{marketDepth.spreadPips.toFixed(1)}</div>
          </div>
          <div>
            <div className="text-[10px] text-zinc-500 mb-0.5">Mid Price</div>
            <div className="font-mono text-emerald-400">{midPrice.toFixed(5)}</div>
          </div>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-between px-3 py-1.5 border-b border-zinc-800 bg-zinc-900/20 text-[10px]">
        <button
          onClick={() => setShowCumulative(!showCumulative)}
          className={`px-2 py-1 rounded font-medium transition-colors ${
            showCumulative
              ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
              : 'bg-zinc-800 text-zinc-500 hover:text-zinc-300'
          }`}
        >
          Cumulative
        </button>

        <div className="flex items-center gap-1">
          <span className="text-zinc-500">Group:</span>
          <select
            value={grouping}
            onChange={(e) => setGrouping(Number(e.target.value))}
            className="px-1.5 py-0.5 bg-zinc-800 border border-zinc-700 rounded text-zinc-300 focus:outline-none"
          >
            <option value={1}>1 pip</option>
            <option value={5}>5 pips</option>
            <option value={10}>10 pips</option>
          </select>
        </div>
      </div>

      {/* Column Headers */}
      <div className="grid grid-cols-3 gap-2 px-3 py-1.5 border-b border-zinc-800 bg-zinc-900/30 text-[10px] font-medium text-zinc-500 uppercase tracking-wide">
        <div className="text-right">Price</div>
        <div className="text-right">Volume</div>
        <div className="text-right">{showCumulative ? 'Total' : 'Orders'}</div>
      </div>

      {/* Order Book */}
      <div className="flex-1 overflow-y-auto">
        {/* Asks (Sell Orders) - Reversed */}
        <div className="border-b-2 border-red-500/20">
          {[...groupedAsks].reverse().map((ask, index) => (
            <DepthRow
              key={`ask-${index}`}
              entry={ask}
              maxVolume={maxVolume}
              side="ask"
              showCumulative={showCumulative}
            />
          ))}
        </div>

        {/* Current Spread Indicator */}
        <div className="sticky top-0 z-10 bg-zinc-900 border-y border-zinc-700 shadow-lg">
          <div className="flex items-center justify-between px-3 py-2">
            <div className="flex items-center gap-2">
              <TrendingDown className="w-4 h-4 text-red-400" />
              <div>
                <div className="text-[10px] text-zinc-500">Best Bid</div>
                <div className="text-sm font-mono font-bold text-red-400">
                  {marketDepth.bids[0]?.price.toFixed(5)}
                </div>
              </div>
            </div>

            <div className="text-center">
              <div className="text-[10px] text-zinc-500">Spread</div>
              <div className="text-xs font-mono text-zinc-400">
                {marketDepth.spreadPips.toFixed(1)} pips
              </div>
            </div>

            <div className="flex items-center gap-2">
              <div className="text-right">
                <div className="text-[10px] text-zinc-500">Best Ask</div>
                <div className="text-sm font-mono font-bold text-emerald-400">
                  {marketDepth.asks[0]?.price.toFixed(5)}
                </div>
              </div>
              <TrendingUp className="w-4 h-4 text-emerald-400" />
            </div>
          </div>
        </div>

        {/* Bids (Buy Orders) */}
        <div className="border-t-2 border-emerald-500/20">
          {groupedBids.map((bid, index) => (
            <DepthRow
              key={`bid-${index}`}
              entry={bid}
              maxVolume={maxVolume}
              side="bid"
              showCumulative={showCumulative}
            />
          ))}
        </div>
      </div>

      {/* Footer Stats */}
      <div className="flex items-center justify-between px-3 py-2 border-t border-zinc-800 bg-zinc-900/50 text-[10px]">
        <div className="flex items-center gap-1">
          <div className="w-2 h-2 rounded-full bg-emerald-500"></div>
          <span className="text-zinc-500">Bid Total:</span>
          <span className="font-mono text-emerald-400">
            {formatVolume(groupedBids.reduce((sum, b) => sum + b.volume, 0))}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-2 h-2 rounded-full bg-red-500"></div>
          <span className="text-zinc-500">Ask Total:</span>
          <span className="font-mono text-red-400">
            {formatVolume(groupedAsks.reduce((sum, a) => sum + a.volume, 0))}
          </span>
        </div>
      </div>
    </div>
  );
};

// Depth Row Component
const DepthRow = ({
  entry,
  maxVolume,
  side,
  showCumulative,
}: {
  entry: MarketDepthEntry;
  maxVolume: number;
  side: 'bid' | 'ask';
  showCumulative: boolean;
}) => {
  const percentage = maxVolume > 0 ? (entry.volume / maxVolume) * 100 : 0;
  const bgColor = side === 'bid' ? 'bg-emerald-500/10' : 'bg-red-500/10';
  const textColor = side === 'bid' ? 'text-emerald-400' : 'text-red-400';
  const barColor = side === 'bid' ? 'bg-emerald-500/30' : 'bg-red-500/30';

  return (
    <div className={`relative group hover:bg-zinc-800/50 transition-colors`}>
      {/* Volume bar background */}
      <div
        className={`absolute ${side === 'bid' ? 'right-0' : 'left-0'} top-0 h-full ${barColor} transition-all duration-300`}
        style={{ width: `${percentage}%` }}
      />

      {/* Content */}
      <div className="relative grid grid-cols-3 gap-2 px-3 py-1.5 text-xs">
        <div className={`font-mono ${textColor} font-medium text-right`}>
          {entry.price.toFixed(5)}
        </div>
        <div className="text-right font-mono text-zinc-300">
          {formatVolume(entry.volume)}
        </div>
        <div className="text-right font-mono text-zinc-500 text-[10px]">
          {showCumulative ? formatVolume(entry.cumulative) : (entry.orders || '-')}
        </div>
      </div>
    </div>
  );
};

// Utility Functions
const formatVolume = (volume: number): string => {
  if (volume >= 1000000) {
    return `${(volume / 1000000).toFixed(2)}M`;
  }
  if (volume >= 1000) {
    return `${(volume / 1000).toFixed(2)}K`;
  }
  return volume.toFixed(2);
};

const groupPriceLevels = (
  levels: MarketDepthEntry[],
  grouping: number,
  side: 'bid' | 'ask'
): MarketDepthEntry[] => {
  const pipValue = 0.0001;
  const groupSize = grouping * pipValue;

  const grouped = new Map<number, MarketDepthEntry>();

  levels.forEach((level) => {
    const groupKey =
      side === 'bid'
        ? Math.floor(level.price / groupSize) * groupSize
        : Math.ceil(level.price / groupSize) * groupSize;

    const existing = grouped.get(groupKey);

    if (existing) {
      existing.volume += level.volume;
      existing.cumulative += level.volume;
      existing.orders = (existing.orders || 0) + (level.orders || 1);
    } else {
      grouped.set(groupKey, {
        price: groupKey,
        volume: level.volume,
        cumulative: level.cumulative,
        orders: level.orders || 1,
      });
    }
  });

  return Array.from(grouped.values()).sort((a, b) =>
    side === 'bid' ? b.price - a.price : a.price - b.price
  );
};

// Mock data generator
const generateMockDepth = (symbol: string, depth: number): MarketDepth => {
  const basePrice = 1.0850; // Example price
  const pipValue = symbol.includes('JPY') ? 0.01 : 0.0001;

  const bids: MarketDepthEntry[] = [];
  const asks: MarketDepthEntry[] = [];

  let cumulativeBid = 0;
  let cumulativeAsk = 0;

  for (let i = 0; i < depth; i++) {
    const bidPrice = basePrice - i * pipValue * (Math.random() * 2 + 1);
    const bidVolume = Math.random() * 100 + 20;
    cumulativeBid += bidVolume;

    bids.push({
      price: bidPrice,
      volume: bidVolume,
      cumulative: cumulativeBid,
      orders: Math.floor(Math.random() * 10) + 1,
    });

    const askPrice = basePrice + i * pipValue * (Math.random() * 2 + 1);
    const askVolume = Math.random() * 100 + 20;
    cumulativeAsk += askVolume;

    asks.push({
      price: askPrice,
      volume: askVolume,
      cumulative: cumulativeAsk,
      orders: Math.floor(Math.random() * 10) + 1,
    });
  }

  const spread = asks[0].price - bids[0].price;
  const spreadPips = spread / pipValue;

  return {
    symbol,
    bids,
    asks,
    spread,
    spreadPips,
    timestamp: Date.now(),
  };
};
