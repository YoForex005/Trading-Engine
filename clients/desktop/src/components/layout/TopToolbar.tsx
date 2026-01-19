import React from 'react';
import {
    CandlestickChart,
    LineChart,
    MousePointer2,
    Crosshair,
    Minus,
    Type,
    Bell,
    Settings,
    BarChart3,
    Layers,
    ZoomIn,
    ZoomOut,
    Play,
    ChevronDown,
    LayoutTemplate
} from 'lucide-react';
import type { ChartType, Timeframe } from '../TradingChart';

interface TopToolbarProps {
    chartType: ChartType;
    timeframe: Timeframe;
    onChartTypeChange: (type: ChartType) => void;
    onTimeframeChange: (tf: Timeframe) => void;
}

import { MenuBar } from './MenuBar';

export const TopToolbar: React.FC<TopToolbarProps> = ({
    chartType,
    timeframe,
    onChartTypeChange,
    onTimeframeChange
}) => {
    return (
        <div className="flex flex-col z-50 relative">
            {/* 1. Header Bar (Topmost) */}
            <HeaderBar
                symbol="XAUUSD"
                timeframe={timeframe}
                account="900500"
                server="HexyMarkets-Server"
            />

            {/* 2. Menu Bar (File, View, etc.) */}
            <MenuBar />

            {/* 3. Main Toolbar (Icons Row) */}
            <MainToolbar
                chartType={chartType}
                timeframe={timeframe}
                onChartTypeChange={onChartTypeChange}
                onTimeframeChange={onTimeframeChange}
            />
        </div>
    );
};

// --- Sub-Components ---

const HeaderBar = ({ symbol, timeframe, account, server }: { symbol: string, timeframe: string, account: string, server: string }) => (
    <div className="h-8 bg-[#1e1e1e] flex items-center justify-between px-3 text-[11px] font-medium border-b border-[#2a2e39] select-none text-zinc-400">
        {/* Left: Platform Logo & Account */}
        <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-zinc-300">
                <div className="w-4 h-4 rounded bg-gradient-to-br from-blue-500 to-teal-400 flex items-center justify-center text-[9px] text-white font-bold tracking-tight">
                    H
                </div>
                <span className="font-bold tracking-wide text-zinc-200">HEXY</span>
            </div>
            <div className="flex items-center gap-1.5 pl-3 border-l border-zinc-700/50">
                <span className="text-zinc-500">ACC:</span>
                <span className="text-zinc-300 font-mono">{account}</span>
                <span className="text-zinc-600 px-1">•</span>
                <span className="text-zinc-500">{server}</span>
            </div>
        </div>

        {/* Center: Active Symbol Info */}
        <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center gap-2">
            <div className="bg-[#252525] px-3 py-0.5 rounded-full border border-zinc-800 flex items-center gap-2 shadow-sm">
                <span className="text-emerald-400 font-bold tracking-wider">{symbol}</span>
                <span className="w-[1px] h-3 bg-zinc-700"></span>
                <span className="text-blue-400 font-mono font-bold">{timeframe.toUpperCase()}</span>
            </div>
        </div>

        {/* Right: System Status */}
        <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 px-2 py-0.5 bg-[#252525] rounded border border-zinc-800">
                <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></div>
                <span className="text-emerald-500 font-bold text-[10px]">CONNECTED</span>
                <span className="text-zinc-600 text-[10px] ml-1">12ms</span>
            </div>
            <div className="flex items-center gap-3 text-zinc-500 border-l border-zinc-700/50 pl-3">
                <Bell size={13} className="hover:text-zinc-300 cursor-pointer transition-colors" />
                <Settings size={13} className="hover:text-zinc-300 cursor-pointer transition-colors" />
            </div>
        </div>
    </div>
);

const MainToolbar = ({
    chartType,
    timeframe,
    onChartTypeChange,
    onTimeframeChange
}: TopToolbarProps) => {
    return (
        <div className="h-10 bg-gradient-to-b from-[#252525] to-[#1e1e1e] border-b border-black flex items-center px-2 gap-3 shadow-md">
            {/* Section 1: Trading Actions */}
            <ToolbarGroup>
                <ToolButton icon={<Play size={15} className="fill-emerald-500 text-emerald-500" />} label="Algo Trading" activeColor="emerald" />
                <ToolButton icon={<Layers size={15} />} label="New Order" />
            </ToolbarGroup>

            <Divider />

            {/* Section 2: Chart Controls */}
            <ToolbarGroup>
                <ToolButton active={chartType === 'bar'} onClick={() => onChartTypeChange('bar')} icon={<BarChart3 size={16} />} title="Bar Chart" />
                <ToolButton active={chartType === 'candlestick'} onClick={() => onChartTypeChange('candlestick')} icon={<CandlestickChart size={16} />} title="Candlesticks" />
                <ToolButton active={chartType === 'area'} onClick={() => onChartTypeChange('area')} icon={<LineChart size={16} />} title="Line Chart" />
            </ToolbarGroup>

            <Divider />

            {/* Section 3: Tools */}
            <ToolbarGroup>
                <ToolButton icon={<Crosshair size={16} />} active />
                <ToolButton icon={<ZoomIn size={16} />} />
                <ToolButton icon={<ZoomOut size={16} />} />
                <ToolButton icon={<LayoutTemplate size={16} />} />
            </ToolbarGroup>

            <Divider />

            {/* Section 4: Drawing Tools */}
            <ToolbarGroup>
                <ToolButton icon={<MousePointer2 size={15} />} />
                <ToolButton icon={<Minus size={15} className="rotate-45" />} />
                <ToolButton icon={<Minus size={15} />} />
                <ToolButton icon={<Type size={15} />} />
                <div className="text-zinc-600 text-[10px] cursor-pointer hover:text-zinc-400 transition-colors">
                    <ChevronDown size={12} />
                </div>
            </ToolbarGroup>

            <div className="flex-1"></div>

            {/* Section 5: Timeframes */}
            <div className="flex items-center gap-0.5 bg-[#181818] p-0.5 rounded-md border border-zinc-800">
                {(['M1', 'M5', 'M15', 'M30', 'H1', 'H4', 'D1', 'W1', 'MN'] as string[]).map((tf) => (
                    <button
                        key={tf}
                        onClick={() => onTimeframeChange(tf.toLowerCase() as Timeframe)}
                        className={`px-2.5 py-1 text-[10px] font-bold rounded-[3px] transition-all ${timeframe.toUpperCase() === tf
                            ? 'bg-blue-600 text-white shadow-sm'
                            : 'text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800'
                            }`}
                    >
                        {tf}
                    </button>
                ))}
            </div>
        </div>
    );
};

// --- Utility Components ---

const ToolbarGroup = ({ children }: { children: React.ReactNode }) => (
    <div className="flex items-center gap-1">
        {children}
    </div>
);

const Divider = () => (
    <div className="w-[1px] h-5 bg-zinc-700/50 mx-1"></div>
);

interface ToolButtonProps {
    icon: React.ReactNode;
    label?: string;
    title?: string;
    active?: boolean;
    activeColor?: 'blue' | 'emerald';
    onClick?: () => void;
}

const ToolButton = ({ icon, label, title, active, activeColor = 'blue', onClick }: ToolButtonProps) => (
    <button
        onClick={onClick}
        title={title || label}
        className={`flex items-center gap-1.5 px-1.5 py-1.5 rounded transition-all
            ${active
                ? activeColor === 'blue'
                    ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20 shadow-inner'
                    : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-inner'
                : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#333] border border-transparent'
            }
        `}
    >
        {icon}
        {label && <span className="text-[11px] font-medium hidden lg:inline-block">{label}</span>}
    </button>
);
