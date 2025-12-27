import React, { useState } from 'react';
import { ArrowUpDown, MoreHorizontal, TrendingUp, TrendingDown, Info, BarChart2 } from 'lucide-react';
import { Sparkline } from './Sparkline';

interface Tick {
    symbol: string;
    bid: number;
    ask: number;
    spread?: number;
    timestamp: number;
    prevBid?: number;
    lp?: string;
}

interface MarketWatchRowProps {
    symbol: string;
    tick: Tick; // tick might be undefined initially
    history: number[];
    selected: boolean;
    onClick: () => void;
    onContextMenu: (e: React.MouseEvent) => void;
}

export function MarketWatchRow({ symbol, tick, history, selected, onClick, onContextMenu }: MarketWatchRowProps) {
    const direction = tick?.prevBid !== undefined
        ? tick.bid > tick.prevBid ? 'up' : tick.bid < tick.prevBid ? 'down' : 'none'
        : 'none';

    const formatPrice = (price: number, sym: string) => {
        if (!price) return '---';
        const isJPY = sym.includes('JPY');
        const digits = isJPY ? 3 : 5;
        return price.toFixed(digits);
    };

    const spread = tick?.spread ? tick.spread.toFixed(1) : '-.-';

    // Calculate change % (simplified based on history if available, else 0)
    // Ideally this comes from Day Open, but we'll use history[0] as proxy if available
    const changePct = history.length > 0 && tick?.bid
        ? ((tick.bid - history[0]) / history[0] * 100)
        : 0;

    const changeColor = changePct > 0 ? 'text-emerald-400' : changePct < 0 ? 'text-red-400' : 'text-zinc-500';

    return (
        <div
            onClick={onClick}
            onContextMenu={onContextMenu}
            className={`group flex items-center justify-between px-3 py-2 cursor-pointer transition-all text-xs border-b border-zinc-800/50 ${selected ? 'bg-emerald-500/5 border-l-2 border-l-emerald-500' : 'hover:bg-zinc-800/50 border-l-2 border-l-transparent'
                }`}
        >
            {/* Symbol & Spread */}
            <div className="flex flex-col gap-0.5 w-24">
                <div className="flex items-center gap-1.5 font-bold text-zinc-200">
                    {symbol}
                </div>
                <div className="flex items-center gap-2 text-[10px] text-zinc-500">
                    <span className="flex items-center gap-0.5 bg-zinc-800/50 px-1 rounded">
                        {spread}
                    </span>
                    <span className={changeColor}>{changePct > 0 ? '+' : ''}{changePct.toFixed(2)}%</span>
                </div>
            </div>

            {/* Sparkline */}
            <div className="hidden sm:block opacity-50">
                <Sparkline
                    data={history}
                    width={50}
                    height={20}
                    color={changePct >= 0 ? '#10b981' : '#ef4444'}
                />
            </div>

            {/* Prices */}
            <div className="text-right font-mono min-w-[80px]">
                <div className={`text-sm ${direction === 'up' ? 'text-emerald-400' : direction === 'down' ? 'text-red-400' : 'text-zinc-300'}`}>
                    {tick?.bid ? formatPrice(tick.bid, symbol) : '---'}
                </div>
                <div className="text-[10px] text-zinc-500">
                    {tick?.ask ? formatPrice(tick.ask, symbol) : '---'}
                </div>
            </div>
        </div>
    );
}
