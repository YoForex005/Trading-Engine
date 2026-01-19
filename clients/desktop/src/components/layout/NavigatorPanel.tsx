import React, { useState } from 'react';
import {
    Folder,
    TrendingUp,
    Code2,
    User,
    Globe,
    ChevronRight,
    ChevronDown,
    Server,
    Zap,
    ShoppingBag,
    Radio,
    Cloud,
    FileCode,
    Bot
} from 'lucide-react';

export const NavigatorPanel: React.FC = () => {
    return (
        <div className="flex flex-col h-full bg-[#1e1e1e] text-zinc-300 select-none border-r border-zinc-700">
            <div className="px-2 py-1 bg-[#2d3436] border-b border-zinc-700 text-xs font-bold text-zinc-400 uppercase tracking-wider">
                Navigator
            </div>
            <div className="flex-1 overflow-y-auto p-1 scrollbar-thin scrollbar-thumb-zinc-700">
                {/* Accounts */}
                <TreeItem label="Accounts" icon={<User size={14} />} defaultOpen>
                    <TreeItem label="RTX5-Server (Live)" icon={<Server size={14} className="text-emerald-500" />}>
                        <TreeItem label="900500: John Doe" icon={<User size={14} className="text-emerald-400" />}>
                            <div className="pl-6 py-0.5 text-[10px] text-zinc-500 font-mono">1:100 USD</div>
                        </TreeItem>
                    </TreeItem>
                    <TreeItem label="RTX5-Demo (Demo)" icon={<Server size={14} className="text-amber-500" />}>
                        <TreeItem label="109283: Demo User" icon={<User size={14} className="text-amber-400" />} />
                    </TreeItem>
                </TreeItem>

                {/* Indicators */}
                <TreeItem label="Indicators" icon={<TrendingUp size={14} />} defaultOpen>
                    <TreeItem label="Trend" icon={<Folder size={14} className="text-amber-300" />}>
                        <TreeItem label="Moving Average" icon={<TrendingUp size={14} />} />
                        <TreeItem label="Bollinger Bands" icon={<TrendingUp size={14} />} />
                    </TreeItem>
                    <TreeItem label="Oscillators" icon={<Folder size={14} className="text-amber-300" />}>
                        <TreeItem label="RSI" icon={<TrendingUp size={14} />} />
                        <TreeItem label="MACD" icon={<TrendingUp size={14} />} />
                        <TreeItem label="Stochastic" icon={<TrendingUp size={14} />} />
                    </TreeItem>
                    <TreeItem label="Volumes" icon={<Folder size={14} className="text-amber-300" />} />
                    <TreeItem label="Bill Williams" icon={<Folder size={14} className="text-amber-300" />} />
                </TreeItem>

                {/* RTX5 Advisors */}
                <TreeItem label="RTX5 Advisors" icon={<Bot size={14} className="text-blue-400" />}>
                    <TreeItem label="RTX5_Scalper_Pro" icon={<Bot size={14} className="text-blue-300" />} />
                    <TreeItem label="News_Event_Trader" icon={<Bot size={14} className="text-blue-300" />} />
                    <TreeItem label="Examples" icon={<Folder size={14} className="text-amber-300" />}>
                        <TreeItem label="MACD Sample" icon={<Bot size={14} className="text-zinc-400" />} />
                        <TreeItem label="Moving Average" icon={<Bot size={14} className="text-zinc-400" />} />
                    </TreeItem>
                </TreeItem>

                {/* RTX5 Scripts */}
                <TreeItem label="RTX5 Scripts" icon={<FileCode size={14} className="text-purple-400" />}>
                    <TreeItem label="CloseAll" icon={<FileCode size={14} className="text-purple-300" />} />
                    <TreeItem label="CalculateLot" icon={<FileCode size={14} className="text-purple-300" />} />
                </TreeItem>

                {/* RTX5 Services */}
                <TreeItem label="RTX5 Services" icon={<Zap size={14} className="text-yellow-400" />} defaultOpen>
                    <TreeItem label="RTX5 Marketplace" icon={<ShoppingBag size={14} className="text-blue-400" />} />
                    <TreeItem label="RTX5 Signals" icon={<Radio size={14} className="text-emerald-400" />} />
                    <TreeItem label="RTX5 Cloud Hosting" icon={<Cloud size={14} className="text-sky-400" />} />
                </TreeItem>
            </div>
        </div>
    );
};

function TreeItem({
    label,
    icon,
    children,
    defaultOpen = false
}: {
    label: string,
    icon: React.ReactNode,
    children?: React.ReactNode,
    defaultOpen?: boolean
}) {
    const [isOpen, setIsOpen] = useState(defaultOpen);
    const hasChildren = React.Children.count(children) > 0;

    return (
        <div className="pl-2">
            <div
                className="flex items-center gap-1.5 py-0.5 px-1 hover:bg-[#2d3436] cursor-pointer rounded-sm group select-none"
                onClick={() => hasChildren && setIsOpen(!isOpen)}
                onDoubleClick={() => hasChildren && setIsOpen(!isOpen)}
            >
                {hasChildren && (
                    <span className="text-zinc-500 group-hover:text-zinc-300">
                        {isOpen ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
                    </span>
                )}
                {!hasChildren && <span className="w-3" />}
                <span className="text-zinc-400 group-hover:text-zinc-200">{icon}</span>
                <span className="text-xs font-medium text-zinc-300 group-hover:text-white truncate">{label}</span>
            </div>
            {isOpen && children && (
                <div className="border-l border-zinc-700 ml-2.5">
                    {children}
                </div>
            )}
        </div>
    );
}
