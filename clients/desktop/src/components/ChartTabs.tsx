import React from 'react';
import { X } from 'lucide-react';
import { type Timeframe } from './TradingChart';

export interface ChartTab {
    id: string;
    symbol: string;
    timeframe: Timeframe;
}

interface ChartTabsProps {
    tabs: ChartTab[];
    activeTabId: string;
    onTabClick: (id: string) => void;
    onTabClose: (id: string) => void;
}

export const ChartTabs: React.FC<ChartTabsProps> = ({ tabs, activeTabId, onTabClick, onTabClose }) => {
    return (
        <div className="h-8 bg-[#18181b] flex items-center border-t border-[#2d3436] px-1 select-none overflow-x-auto scrollbar-thin scrollbar-thumb-zinc-700">
            {tabs.map((tab) => {
                const isActive = tab.id === activeTabId;
                return (
                    <div
                        key={tab.id}
                        onClick={() => onTabClick(tab.id)}
                        className={`
                            group flex items-center min-w-[120px] max-w-[160px] h-[26px] px-2 rounded-t mr-1 cursor-pointer text-xs
                            border-t border-r border-l border-b-0 transition-colors
                            ${isActive
                                ? 'bg-white text-black border-zinc-400 font-medium' // Active: Light (MT5 style often has active tab distinct) or stick to dark theme
                                : 'bg-[#1e1e1e] text-zinc-400 border-transparent hover:bg-[#2a2e39] hover:border-zinc-700'}
                        `}
                    >
                        <span className="flex-1 truncate">
                            {tab.symbol}, {tab.timeframe}
                        </span>

                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onTabClose(tab.id);
                            }}
                            className={`p-0.5 rounded-sm opacity-0 group-hover:opacity-100 transition-opacity ${isActive ? 'hover:bg-zinc-200 text-black' : 'hover:bg-zinc-600 hover:text-white'}`}
                        >
                            <X size={10} />
                        </button>
                    </div>
                );
            })}
        </div>
    );
};
