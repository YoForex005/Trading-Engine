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

    // Simulate depth data based on current tick
    const depthLevels = useMemo(() => {
        if (!currentTick) return { asks: [], bids: [] };

        const levels = 10;
        const spread = currentTick.ask - currentTick.bid;
        const step = spread > 0 ? spread / 2 : 0.0001; // Estimate step size

        const asks: DepthLevel[] = [];
        const bids: DepthLevel[] = [];

        // Generate Asks (Price ascending)
        for (let i = 0; i < levels; i++) {
            const price = currentTick.ask + (i * step);
            asks.push({
                price,
                volume: Math.floor(Math.random() * 50) + 1, // Simulated volume
                total: 0,
                type: 'ask'
            });
        }
        // Reverse asks to show highest on top? Standard DOM shows Ask ladder ascending up from mid?
        // Standard DOM:
        // High Price (Ask)
        // ...
        // Best Ask
        // Best Bid
        // ...
        // Low Price (Bid)
        asks.reverse();

        // Generate Bids (Price descending)
        for (let i = 0; i < levels; i++) {
            const price = currentTick.bid - (i * step);
            bids.push({
                price,
                volume: Math.floor(Math.random() * 50) + 1,
                total: 0,
                type: 'bid'
            });
        }

        return { asks, bids };
    }, [currentTick]);

    if (!currentTick) {
        return (
            <div className="flex items-center justify-center h-full bg-[#1e1e1e] text-zinc-500 text-xs">
                Waiting for data...
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full bg-[#1e1e1e] border-l border-[#2d3436] w-64 select-none">
            {/* Header */}
            <div className="flex items-center justify-between px-2 py-1 bg-[#2d3436] border-b border-[#1e1e1e]">
                <span className="font-bold text-zinc-200 text-sm">DOM: {symbol}</span>
                <button onClick={onClose} className="text-zinc-400 hover:text-white">
                    <X size={14} />
                </button>
            </div>

            {/* Trading Control */}
            <div className="p-2 border-b border-[#2d3436] bg-[#252526]">
                <div className="flex items-center justify-between mb-2">
                    <span className="text-xs text-zinc-400">Volume</span>
                    <input
                        type="number"
                        step="0.01"
                        min="0.01"
                        value={volume}
                        onChange={(e) => setVolume(parseFloat(e.target.value))}
                        className="w-20 bg-[#1e1e1e] border border-[#3f3f46] rounded px-1 text-right text-sm text-white focus:outline-none focus:border-blue-500"
                    />
                </div>
                <div className="flex gap-1">
                    <button
                        onClick={() => onPlaceOrder?.('SELL', currentTick.bid, volume)}
                        className="flex-1 bg-red-600 hover:bg-red-700 text-white text-xs py-1 rounded flex items-center justify-center gap-1 transition-colors"
                    >
                        Sell Mkt
                    </button>
                    <button
                        onClick={() => onPlaceOrder?.('BUY', currentTick.ask, volume)}
                        className="flex-1 bg-blue-600 hover:bg-blue-700 text-white text-xs py-1 rounded flex items-center justify-center gap-1 transition-colors"
                    >
                        Buy Mkt
                    </button>
                </div>
            </div>

            {/* Ladder */}
            <div className="flex-1 overflow-y-auto scrollbar-thin scrollbar-thumb-zinc-700">
                <table className="w-full text-xs border-collapse">
                    <thead className="bg-[#2d3436] text-zinc-400 sticky top-0 z-10">
                        <tr>
                            <th className="px-1 py-1 text-right font-normal">Price</th>
                            <th className="px-1 py-1 text-right font-normal">Vol</th>
                        </tr>
                    </thead>
                    <tbody>
                        {/* Asks (Red) */}
                        {depthLevels.asks.map((level, i) => (
                            <tr
                                key={`ask-${i}`}
                                className="group hover:bg-[#3f3f46] cursor-pointer"
                                onClick={() => onPlaceOrder?.('SELL', level.price, volume)} // Click on Ask to Sell Limit? Or Buy Stop? Usually Click Ask -> Buy (Lift Offer). 
                            // Standard: Click Ask = Buy Limit at that price? Or Market if Best.
                            // Simplification: Click -> Fill limit price input
                            >
                                <td className="px-1 py-0.5 text-right font-mono text-red-400 bg-red-900/10 group-hover:bg-red-900/30">
                                    {level.price.toFixed(5)}
                                </td>
                                <td className="px-1 py-0.5 text-right font-mono text-zinc-300 bg-red-900/10 group-hover:bg-red-900/30">
                                    {level.volume}
                                </td>
                            </tr>
                        ))}

                        {/* Mid Spread Gap */}
                        <tr className="bg-[#1e1e1e] h-1 border-t border-b border-zinc-800">
                            <td colSpan={2} className="text-center text-[10px] text-zinc-600">
                                {(currentTick.ask - currentTick.bid).toFixed(5)}
                            </td>
                        </tr>

                        {/* Bids (Blue/Green) */}
                        {depthLevels.bids.map((level, i) => (
                            <tr
                                key={`bid-${i}`}
                                className="group hover:bg-[#3f3f46] cursor-pointer"
                                onClick={() => onPlaceOrder?.('BUY', level.price, volume)}
                            >
                                <td className="px-1 py-0.5 text-right font-mono text-blue-400 bg-blue-900/10 group-hover:bg-blue-900/30">
                                    {level.price.toFixed(5)}
                                </td>
                                <td className="px-1 py-0.5 text-right font-mono text-zinc-300 bg-blue-900/10 group-hover:bg-blue-900/30">
                                    {level.volume}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};
