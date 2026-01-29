import React from 'react';
import { ChevronUp, ChevronDown } from 'lucide-react';

interface OneClickTradingProps {
    symbol: string;
    bid: number;
    ask: number;
    volume: number;
    onVolumeChange: (vol: number) => void;
    onBuy: () => void;
    onSell: () => void;
    isHidden?: boolean;
}

export const OneClickTrading: React.FC<OneClickTradingProps> = ({
    symbol,
    bid,
    ask,
    volume,
    onVolumeChange,
    onBuy,
    onSell,
    isHidden = false
}) => {
    if (isHidden) return null;

    return (
        <div className="absolute top-12 left-2 z-20 flex flex-col bg-zinc-900 border border-zinc-600 rounded-sm shadow-xl w-48 select-none">
            {/* Header / Draggable Area */}
            <div className="flex items-center justify-between px-2 py-0.5 bg-zinc-800 border-b border-zinc-700 cursor-move">
                <span className="text-[10px] font-bold text-zinc-400">{symbol}</span>
                <div className="flex flex-col items-center justify-center cursor-pointer hover:bg-zinc-700 rounded p-0.5">
                    <ChevronUp size={8} className="text-zinc-500" />
                    <ChevronDown size={8} className="text-zinc-500 -mt-1" />
                </div>
            </div>

            {/* Main Buttons */}
            <div className="flex border-b border-zinc-700 h-16">
                {/* SELL */}
                <button
                    onClick={onSell}
                    className="flex-1 flex flex-col items-center justify-center gap-0.5 bg-zinc-900 hover:bg-[#3a0d0d] transition-colors border-r border-zinc-700 group relative overflow-hidden"
                >
                    <span className="text-[9px] font-bold text-zinc-500 uppercase z-10">Sell</span>
                    <div className="flex items-baseline gap-[1px] z-10">
                        <span className="text-lg font-bold text-zinc-300 group-hover:text-red-400">
                            {bid.toFixed(2)}
                        </span>
                        <span className="text-xs font-medium text-zinc-500 group-hover:text-red-500 align-top -mt-1">
                            {Math.floor((bid * 100000) % 10)}
                        </span>
                    </div>

                    <div className="absolute inset-0 bg-gradient-to-b from-transparent to-red-900/10 opacity-0 group-hover:opacity-100 transition-opacity" />
                </button>

                {/* Center Column: Volume */}
                <div className="w-[1px] bg-zinc-700" />

                {/* BUY */}
                <button
                    onClick={onBuy}
                    className="flex-1 flex flex-col items-center justify-center gap-0.5 bg-zinc-900 hover:bg-[#0d2a3a] transition-colors group relative overflow-hidden"
                >
                    <span className="text-[9px] font-bold text-zinc-500 uppercase z-10">Buy</span>
                    <div className="flex items-baseline gap-[1px] z-10">
                        <span className="text-lg font-bold text-zinc-300 group-hover:text-blue-400">
                            {ask.toFixed(2)}
                        </span>
                        <span className="text-xs font-medium text-zinc-500 group-hover:text-blue-500 align-top -mt-1">
                            {Math.floor((ask * 100000) % 10)}
                        </span>
                    </div>
                    <div className="absolute inset-0 bg-gradient-to-b from-transparent to-blue-900/10 opacity-0 group-hover:opacity-100 transition-opacity" />
                </button>
            </div>

            {/* Footer / Volume */}
            <div className="flex items-center justify-center p-1 bg-zinc-800">
                <input
                    type="number"
                    step={0.01}
                    min={0.01}
                    value={volume}
                    onChange={(e) => onVolumeChange(parseFloat(e.target.value))}
                    className="w-16 h-5 bg-zinc-900 border border-zinc-600 rounded-sm text-center text-xs text-white font-mono focus:outline-none focus:border-zinc-500"
                />
            </div>
        </div>
    );
};
