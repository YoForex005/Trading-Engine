import React, { useMemo } from 'react';

interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;
}

interface DepthOfMarketProps {
    symbol: string;
    tick: Tick | undefined;
}

// Generate synthetic levels around the current price
const generateLevels = (price: number, type: 'bid' | 'ask', count: number = 5) => {
    if (!price) return [];
    const levels = [];
    const spread = price * 0.0001; // Approx 1 pip

    for (let i = 0; i < count; i++) {
        const levelPrice = type === 'bid'
            ? price - (spread * i)
            : price + (spread * i);

        // Simulate volume bell curve (higher volume near BBO)
        const volume = Math.floor(Math.random() * 50) + 10 + (20 - i * 2);

        levels.push({
            price: levelPrice,
            volume: volume,
            total: volume // Cumulative would be calculated if needed
        });
    }
    return type === 'ask' ? levels.reverse() : levels; // Asks ascending (highest on top for standard DOM usually, but ladder view varies)
};

export function DepthOfMarket({ symbol, tick }: DepthOfMarketProps) {
    const bidLevels = useMemo(() => generateLevels(tick?.bid || 0, 'bid', 8), [tick?.bid, symbol]);
    const askLevels = useMemo(() => generateLevels(tick?.ask || 0, 'ask', 8), [tick?.ask, symbol]);

    const maxVol = 100; // For bar scaling

    return (
        <div className="flex flex-col h-full bg-[#09090b] border border-zinc-800 rounded-lg overflow-hidden text-xs font-mono">
            <div className="px-3 py-2 border-b border-zinc-800 bg-zinc-900/50 font-bold text-zinc-300">
                DOM (Level 2)
            </div>

            <div className="flex-1 overflow-y-auto">
                {/* Header */}
                <div className="grid grid-cols-3 gap-1 px-2 py-1 text-[10px] text-zinc-500 border-b border-zinc-800/50">
                    <div className="text-left">Vol</div>
                    <div className="text-center">Price</div>
                    <div className="text-right">Vol</div>
                </div>

                {/* Asks (Sell Limit Orders) - Red */}
                <div className="flex flex-col-reverse"> {/* Reverse to show highest price at top? No, standard ladder shows highest ask at top */}
                    {askLevels.map((level, i) => (
                        <div key={`ask-${i}`} className="grid grid-cols-3 gap-1 px-2 py-0.5 hover:bg-zinc-800/30 transition-colors group cursor-pointer">
                            <div></div>
                            <div className="text-center text-red-400 bg-red-500/5 group-hover:bg-red-500/10 rounded font-mono transition-colors">{level.price.toFixed(5)}</div>
                            <div className="relative text-right text-zinc-400">
                                <div
                                    className="absolute top-0 right-0 bottom-0 bg-red-500/20 rounded-l-sm transition-all duration-300 ease-out"
                                    style={{ width: `${(level.volume / maxVol) * 100}%` }}
                                />
                                <span className="relative z-10 mr-1 text-[9px] font-medium">{level.volume}</span>
                            </div>
                        </div>
                    ))}
                </div>

                {/* Spread / Current Price Indicator */}
                <div className="my-1 py-1 text-center bg-zinc-900 border-y border-zinc-800">
                    <span className="text-zinc-500 text-[10px] uppercase tracking-wider font-semibold">Spread: {(tick?.spread || 0).toFixed(1)}</span>
                </div>

                {/* Bids (Buy Limit Orders) - Green */}
                <div>
                    {bidLevels.map((level, i) => (
                        <div key={`bid-${i}`} className="grid grid-cols-3 gap-1 px-2 py-0.5 hover:bg-zinc-800/30 transition-colors group cursor-pointer">
                            <div className="relative text-left text-zinc-400">
                                <div
                                    className="absolute top-0 left-0 bottom-0 bg-emerald-500/20 rounded-r-sm transition-all duration-300 ease-out"
                                    style={{ width: `${(level.volume / maxVol) * 100}%` }}
                                />
                                <span className="relative z-10 ml-1 text-[9px] font-medium">{level.volume}</span>
                            </div>
                            <div className="text-center text-emerald-400 bg-emerald-500/5 group-hover:bg-emerald-500/10 rounded font-mono transition-colors">{level.price.toFixed(5)}</div>
                            <div></div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}
