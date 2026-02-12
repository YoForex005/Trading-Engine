import React, { useEffect, useMemo, useState } from 'react';
import { useCurrentTick } from '../store/useMarketDataStore';
import { ArrowUp, ArrowDown, X } from 'lucide-react';

interface DepthLevel {
    price: number;
    volume: number;
    total: number;
    type: 'bid' | 'ask';
}

interface DepthOfMarketProps {
    symbol: string;
    onClose?: () => void;
    onPlaceOrder?: (side: 'BUY' | 'SELL', price: number, volume: number) => void;
}

export const DepthOfMarket: React.FC<DepthOfMarketProps> = ({ symbol, onClose, onPlaceOrder }) => {
    const currentTick = useCurrentTick(symbol);
    const [volume, setVolume] = useState(1.0);

    // Simulate depth data based on current tick with realistic distribution
    const depthLevels = useMemo(() => {
        if (!currentTick) return { asks: [], bids: [], totalBidVolume: 0, totalAskVolume: 0, maxVolume: 0 };

        const levels = 15;
        const spread = currentTick.ask - currentTick.bid;
        // Use pip-based steps for realistic order book
        const pipSize = symbol.includes('JPY') ? 0.01 : 0.0001;
        const step = Math.max(pipSize, spread / 4);

        const asks: DepthLevel[] = [];
        const bids: DepthLevel[] = [];
        let totalBidVolume = 0;
        let totalAskVolume = 0;
        let maxVolume = 0;

        // Generate Asks with realistic volume distribution (lower volume further from market)
        for (let i = 0; i < levels; i++) {
            const price = currentTick.ask + (i * step);
            // Volume decreases exponentially as price moves away from market
            const baseVolume = 50 - (i * 2);
            const randomFactor = 0.7 + (Math.random() * 0.6); // 0.7-1.3x
            const vol = Math.max(1, Math.floor(baseVolume * randomFactor));

            asks.push({
                price,
                volume: vol,
                total: 0,
                type: 'ask'
            });
            totalAskVolume += vol;
            maxVolume = Math.max(maxVolume, vol);
        }
        asks.reverse(); // Show highest price on top

        // Generate Bids with realistic volume distribution
        for (let i = 0; i < levels; i++) {
            const price = currentTick.bid - (i * step);
            const baseVolume = 50 - (i * 2);
            const randomFactor = 0.7 + (Math.random() * 0.6);
            const vol = Math.max(1, Math.floor(baseVolume * randomFactor));

            bids.push({
                price,
                volume: vol,
                total: 0,
                type: 'bid'
            });
            totalBidVolume += vol;
            maxVolume = Math.max(maxVolume, vol);
        }

        // Calculate cumulative totals
        let cumulativeAsk = 0;
        asks.forEach(level => {
            cumulativeAsk += level.volume;
            level.total = cumulativeAsk;
        });

        let cumulativeBid = 0;
        bids.forEach(level => {
            cumulativeBid += level.volume;
            level.total = cumulativeBid;
        });

        return { asks, bids, totalBidVolume, totalAskVolume, maxVolume };
    }, [currentTick, symbol]);

    if (!currentTick) {
        return (
            <div className="flex items-center justify-center h-full bg-[#1e1e1e] text-zinc-500 text-xs">
                Waiting for data...
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full bg-[#1e1e1e] border-l border-[#2d3436] w-80 select-none">
            {/* Header */}
            <div className="flex items-center justify-between px-3 py-2 bg-[#2d3436] border-b border-[#1e1e1e]">
                <div className="flex flex-col">
                    <span className="font-bold text-zinc-200 text-sm">{symbol}</span>
                    <span className="text-[10px] text-zinc-500">Depth of Market</span>
                </div>
                <button
                    onClick={onClose}
                    className="text-zinc-400 hover:text-white transition-colors p-1 hover:bg-zinc-700/50 rounded"
                >
                    <X size={16} />
                </button>
            </div>

            {/* Current Price Display */}
            <div className="px-3 py-2 bg-[#252526] border-b border-[#2d3436]">
                <div className="flex justify-between items-center text-xs mb-1">
                    <span className="text-zinc-400">Bid</span>
                    <span className="font-mono text-blue-400 font-semibold">{currentTick.bid.toFixed(5)}</span>
                </div>
                <div className="flex justify-between items-center text-xs mb-1">
                    <span className="text-zinc-400">Ask</span>
                    <span className="font-mono text-red-400 font-semibold">{currentTick.ask.toFixed(5)}</span>
                </div>
                <div className="flex justify-between items-center text-xs">
                    <span className="text-zinc-400">Spread</span>
                    <span className="font-mono text-zinc-300">{((currentTick.ask - currentTick.bid) * 10000).toFixed(1)} pips</span>
                </div>
            </div>

            {/* Trading Control */}
            <div className="p-3 border-b border-[#2d3436] bg-[#1e1e1e]">
                <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-zinc-400 font-medium">Volume</span>
                    <input
                        type="number"
                        step="0.01"
                        min="0.01"
                        value={volume}
                        onChange={(e) => setVolume(parseFloat(e.target.value) || 0.01)}
                        className="w-24 bg-[#252526] border border-[#3f3f46] rounded px-2 py-1 text-right text-sm text-white focus:outline-none focus:border-blue-500 font-mono"
                    />
                </div>
                <div className="flex gap-2">
                    <button
                        onClick={() => onPlaceOrder?.('SELL', currentTick.bid, volume)}
                        className="flex-1 bg-red-600 hover:bg-red-700 active:bg-red-800 text-white text-xs py-2 rounded font-medium transition-colors shadow-lg"
                    >
                        SELL
                    </button>
                    <button
                        onClick={() => onPlaceOrder?.('BUY', currentTick.ask, volume)}
                        className="flex-1 bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white text-xs py-2 rounded font-medium transition-colors shadow-lg"
                    >
                        BUY
                    </button>
                </div>
            </div>

            {/* Ladder with Depth Bars */}
            <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700 scrollbar-track-transparent">
                <table className="w-full text-xs border-collapse">
                    <thead className="bg-[#2d3436] text-zinc-400 sticky top-0 z-10 shadow-md">
                        <tr>
                            <th className="px-2 py-1.5 text-right font-medium w-24">Price</th>
                            <th className="px-2 py-1.5 text-right font-medium w-16">Volume</th>
                            <th className="px-2 py-1.5 text-right font-medium w-20">Total</th>
                        </tr>
                    </thead>
                    <tbody>
                        {/* Asks (Red) - Sell offers */}
                        {depthLevels.asks.map((level, i) => {
                            const isTopLevel = i === depthLevels.asks.length - 1;
                            const depthPercent = (level.volume / depthLevels.maxVolume) * 100;

                            return (
                                <tr
                                    key={`ask-${i}`}
                                    className="group hover:bg-red-900/30 cursor-pointer transition-colors relative"
                                    onClick={() => onPlaceOrder?.('BUY', level.price, volume)}
                                    title={`Click to buy at ${level.price.toFixed(5)}`}
                                >
                                    {/* Depth bar background */}
                                    <td className="relative px-2 py-1 text-right font-mono text-red-400 font-medium">
                                        <div
                                            className="absolute inset-y-0 right-0 bg-red-900/20 transition-all"
                                            style={{ width: `${depthPercent}%` }}
                                        />
                                        <span className={`relative z-10 ${isTopLevel ? 'text-red-300 font-bold' : ''}`}>
                                            {level.price.toFixed(5)}
                                        </span>
                                    </td>
                                    <td className="px-2 py-1 text-right font-mono text-zinc-300 relative">
                                        <span className="relative z-10">{level.volume.toFixed(1)}</span>
                                    </td>
                                    <td className="px-2 py-1 text-right font-mono text-zinc-500 text-[10px] relative">
                                        <span className="relative z-10">{level.total.toFixed(1)}</span>
                                    </td>
                                </tr>
                            );
                        })}

                        {/* Mid Spread Separator */}
                        <tr className="bg-zinc-900 border-y-2 border-yellow-600/30">
                            <td colSpan={3} className="text-center py-1.5">
                                <div className="flex items-center justify-center gap-2">
                                    <span className="text-[10px] text-zinc-500 font-medium">SPREAD</span>
                                    <span className="text-xs text-yellow-500 font-mono font-bold">
                                        {((currentTick.ask - currentTick.bid) * 10000).toFixed(1)}
                                    </span>
                                </div>
                            </td>
                        </tr>

                        {/* Bids (Blue) - Buy orders */}
                        {depthLevels.bids.map((level, i) => {
                            const isTopLevel = i === 0;
                            const depthPercent = (level.volume / depthLevels.maxVolume) * 100;

                            return (
                                <tr
                                    key={`bid-${i}`}
                                    className="group hover:bg-blue-900/30 cursor-pointer transition-colors relative"
                                    onClick={() => onPlaceOrder?.('SELL', level.price, volume)}
                                    title={`Click to sell at ${level.price.toFixed(5)}`}
                                >
                                    <td className="relative px-2 py-1 text-right font-mono text-blue-400 font-medium">
                                        <div
                                            className="absolute inset-y-0 right-0 bg-blue-900/20 transition-all"
                                            style={{ width: `${depthPercent}%` }}
                                        />
                                        <span className={`relative z-10 ${isTopLevel ? 'text-blue-300 font-bold' : ''}`}>
                                            {level.price.toFixed(5)}
                                        </span>
                                    </td>
                                    <td className="px-2 py-1 text-right font-mono text-zinc-300 relative">
                                        <span className="relative z-10">{level.volume.toFixed(1)}</span>
                                    </td>
                                    <td className="px-2 py-1 text-right font-mono text-zinc-500 text-[10px] relative">
                                        <span className="relative z-10">{level.total.toFixed(1)}</span>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>

            {/* Footer with total volumes */}
            <div className="px-3 py-2 bg-[#252526] border-t border-[#2d3436] grid grid-cols-2 gap-2 text-xs">
                <div className="bg-blue-900/20 rounded px-2 py-1.5">
                    <div className="text-zinc-500 text-[10px] mb-0.5">Total Bid</div>
                    <div className="text-blue-400 font-mono font-semibold">{depthLevels.totalBidVolume.toFixed(1)}</div>
                </div>
                <div className="bg-red-900/20 rounded px-2 py-1.5">
                    <div className="text-zinc-500 text-[10px] mb-0.5">Total Ask</div>
                    <div className="text-red-400 font-mono font-semibold">{depthLevels.totalAskVolume.toFixed(1)}</div>
                </div>
            </div>
        </div>
    );
};
